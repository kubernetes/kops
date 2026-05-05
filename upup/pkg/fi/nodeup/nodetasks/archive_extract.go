/*
Copyright 2026 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package nodetasks

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func extractArchive(archivePath, targetDir string, stripComponents int, pattern string) error {
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return fmt.Errorf("error creating directories %q: %v", targetDir, err)
	}

	targetRoot, err := filepath.Abs(targetDir)
	if err != nil {
		return fmt.Errorf("error resolving target directory %q: %v", targetDir, err)
	}
	targetRoot, err = filepath.EvalSymlinks(targetRoot)
	if err != nil {
		return fmt.Errorf("error resolving target directory %q: %v", targetDir, err)
	}

	tarReader, closeArchive, err := openArchive(archivePath)
	if err != nil {
		return err
	}
	defer closeArchive()

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return fmt.Errorf("error reading archive %q: %v", archivePath, err)
		}
		if pattern != "" && !archivePathMatches(pattern, header.Name) {
			continue
		}
		if err := extractArchiveEntry(tarReader, header, targetRoot, stripComponents); err != nil {
			return fmt.Errorf("error extracting %q from %q: %v", header.Name, archivePath, err)
		}
	}
}

func openArchive(archivePath string) (*tar.Reader, func() error, error) {
	file, err := os.Open(archivePath)
	if err != nil {
		return nil, nil, fmt.Errorf("error opening archive %q: %v", archivePath, err)
	}

	reader, gzipCloser, err := maybeGzipReader(file)
	if err != nil {
		file.Close()
		return nil, nil, fmt.Errorf("error reading archive %q: %v", archivePath, err)
	}

	closeArchive := func() error {
		if gzipCloser != nil {
			if err := gzipCloser.Close(); err != nil {
				_ = file.Close()
				return err
			}
		}
		return file.Close()
	}
	return tar.NewReader(reader), closeArchive, nil
}

func archivePathMatches(pattern, name string) bool {
	if pattern == name {
		return true
	}
	matches, err := path.Match(pattern, name)
	return err == nil && matches
}

func extractArchiveEntry(reader io.Reader, header *tar.Header, targetRoot string, stripComponents int) error {
	targetPath, ok, err := archiveTargetPath(targetRoot, header.Name, stripComponents)
	if err != nil || !ok {
		return err
	}

	switch header.Typeflag {
	case tar.TypeDir:
		return extractArchiveDirectory(targetRoot, targetPath, header)
	case tar.TypeReg, tar.TypeRegA:
		return extractArchiveRegularFile(reader, targetRoot, targetPath, header)
	case tar.TypeSymlink:
		return extractArchiveSymlink(targetRoot, targetPath, header)
	case tar.TypeLink:
		return extractArchiveHardlink(targetRoot, targetPath, header, stripComponents)
	case tar.TypeXGlobalHeader, tar.TypeXHeader:
		return nil
	default:
		return fmt.Errorf("unsupported archive entry type %q", header.Typeflag)
	}
}

func archiveTargetPath(targetRoot, name string, stripComponents int) (string, bool, error) {
	relativePath, ok, err := archiveRelativePath(name, stripComponents)
	if err != nil || !ok {
		return "", ok, err
	}
	targetPath := filepath.Join(targetRoot, filepath.FromSlash(relativePath))
	if err := ensurePathInside(targetRoot, targetPath); err != nil {
		return "", false, err
	}
	return targetPath, true, nil
}

func archiveRelativePath(name string, stripComponents int) (string, bool, error) {
	if stripComponents < 0 {
		return "", false, fmt.Errorf("strip components cannot be negative")
	}
	if name == "" {
		return "", false, nil
	}
	if path.IsAbs(name) {
		return "", false, fmt.Errorf("archive path %q is absolute", name)
	}

	var components []string
	for _, component := range strings.Split(name, "/") {
		switch component {
		case "", ".":
			continue
		case "..":
			return "", false, fmt.Errorf("archive path %q escapes target directory", name)
		default:
			components = append(components, component)
		}
	}

	if stripComponents > len(components) {
		return "", false, nil
	}
	components = components[stripComponents:]
	if len(components) == 0 {
		return "", false, nil
	}
	return path.Join(components...), true, nil
}

func extractArchiveDirectory(targetRoot, targetPath string, header *tar.Header) error {
	if err := ensureParentReady(targetRoot, targetPath); err != nil {
		return err
	}
	if err := ensureExistingSymlinkInside(targetRoot, targetPath); err != nil {
		return err
	}
	mode := archiveFileMode(header, 0o755)
	if err := os.MkdirAll(targetPath, mode); err != nil {
		return fmt.Errorf("error creating directory %q: %v", targetPath, err)
	}
	if err := ensureExistingSymlinkInside(targetRoot, targetPath); err != nil {
		return err
	}
	if err := os.Chmod(targetPath, mode); err != nil {
		return fmt.Errorf("error setting mode on directory %q: %v", targetPath, err)
	}
	return setArchiveModTime(targetPath, header)
}

func extractArchiveRegularFile(reader io.Reader, targetRoot, targetPath string, header *tar.Header) error {
	if err := ensureParentReady(targetRoot, targetPath); err != nil {
		return err
	}
	if err := removeExistingNonDirectory(targetPath); err != nil {
		return err
	}

	mode := archiveFileMode(header, 0o644)
	file, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return fmt.Errorf("error creating file %q: %v", targetPath, err)
	}
	if _, err := io.Copy(file, reader); err != nil {
		_ = file.Close()
		return fmt.Errorf("error writing file %q: %v", targetPath, err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("error closing file %q: %v", targetPath, err)
	}
	if err := os.Chmod(targetPath, mode); err != nil {
		return fmt.Errorf("error setting mode on file %q: %v", targetPath, err)
	}
	return setArchiveModTime(targetPath, header)
}

func extractArchiveSymlink(targetRoot, targetPath string, header *tar.Header) error {
	if header.Linkname == "" {
		return fmt.Errorf("symlink %q has empty target", header.Name)
	}
	if err := ensureParentReady(targetRoot, targetPath); err != nil {
		return err
	}
	if err := removeExistingNonDirectory(targetPath); err != nil {
		return err
	}
	if err := os.Symlink(header.Linkname, targetPath); err != nil {
		return fmt.Errorf("error creating symlink %q -> %q: %v", targetPath, header.Linkname, err)
	}
	return nil
}

func extractArchiveHardlink(targetRoot, targetPath string, header *tar.Header, stripComponents int) error {
	linkPath, ok, err := archiveTargetPath(targetRoot, header.Linkname, stripComponents)
	if err != nil || !ok {
		return err
	}
	if err := ensureExistingSymlinkInside(targetRoot, linkPath); err != nil {
		return err
	}
	if err := ensureParentReady(targetRoot, targetPath); err != nil {
		return err
	}
	if err := removeExistingNonDirectory(targetPath); err != nil {
		return err
	}
	if err := os.Link(linkPath, targetPath); err != nil {
		return fmt.Errorf("error creating hardlink %q -> %q: %v", targetPath, linkPath, err)
	}
	return setArchiveModTime(targetPath, header)
}

func archiveFileMode(header *tar.Header, defaultMode os.FileMode) os.FileMode {
	mode := os.FileMode(header.Mode) & os.ModePerm
	if mode == 0 {
		return defaultMode
	}
	return mode
}

func setArchiveModTime(targetPath string, header *tar.Header) error {
	if header.ModTime.IsZero() {
		return nil
	}
	if err := os.Chtimes(targetPath, header.ModTime, header.ModTime); err != nil {
		return fmt.Errorf("error setting timestamps on %q: %v", targetPath, err)
	}
	return nil
}

func ensureParentReady(targetRoot, targetPath string) error {
	if err := ensureParentPathSafe(targetRoot, targetPath); err != nil {
		return err
	}
	parent := filepath.Dir(targetPath)
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return fmt.Errorf("error creating directories %q: %v", parent, err)
	}
	return ensureParentPathSafe(targetRoot, targetPath)
}

func ensureParentPathSafe(targetRoot, targetPath string) error {
	parent := filepath.Dir(targetPath)
	if err := ensurePathInside(targetRoot, parent); err != nil {
		return err
	}

	relative, err := filepath.Rel(targetRoot, parent)
	if err != nil {
		return fmt.Errorf("error checking path %q: %v", parent, err)
	}
	if relative == "." {
		return nil
	}

	current := targetRoot
	for _, component := range strings.Split(relative, string(os.PathSeparator)) {
		current = filepath.Join(current, component)
		info, err := os.Lstat(current)
		if os.IsNotExist(err) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("error checking path %q: %v", current, err)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			if err := ensureExistingSymlinkInside(targetRoot, current); err != nil {
				return err
			}
			continue
		}
		if !info.IsDir() {
			return fmt.Errorf("archive parent path %q is not a directory", current)
		}
	}
	return nil
}

func ensureExistingSymlinkInside(targetRoot, targetPath string) error {
	info, err := os.Lstat(targetPath)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("error checking path %q: %v", targetPath, err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		return nil
	}
	resolved, err := filepath.EvalSymlinks(targetPath)
	if err != nil {
		return fmt.Errorf("error resolving symlink %q: %v", targetPath, err)
	}
	if err := ensurePathInside(targetRoot, resolved); err != nil {
		return fmt.Errorf("archive path %q resolves outside target directory: %v", targetPath, err)
	}
	return nil
}

func removeExistingNonDirectory(targetPath string) error {
	info, err := os.Lstat(targetPath)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("error checking path %q: %v", targetPath, err)
	}
	if info.IsDir() {
		return fmt.Errorf("cannot replace directory %q with archive entry", targetPath)
	}
	if err := os.Remove(targetPath); err != nil {
		return fmt.Errorf("error removing existing path %q: %v", targetPath, err)
	}
	return nil
}

func ensurePathInside(targetRoot, targetPath string) error {
	relative, err := filepath.Rel(targetRoot, targetPath)
	if err != nil {
		return fmt.Errorf("error checking path %q: %v", targetPath, err)
	}
	if relative == ".." || strings.HasPrefix(relative, ".."+string(os.PathSeparator)) {
		return fmt.Errorf("path %q is outside target directory %q", targetPath, targetRoot)
	}
	return nil
}
