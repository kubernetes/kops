/*
Copyright 2019 The Kubernetes Authors.

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

package vfs

import (
	"fmt"
	"io/ioutil"
	"os"

	"k8s.io/klog"
	"k8s.io/kops/util/pkg/hashing"
)

// VFSScan scans a source Path for changed files
type VFSScan struct {
	Base   Path
	hashes map[string]*hashing.Hash
}

func NewVFSScan(base Path) *VFSScan {
	return &VFSScan{Base: base}
}

type ChangeType string

const (
	ChangeTypeAdded    ChangeType = "ADDED"
	ChangeTypeRemoved  ChangeType = "REMOVED"
	ChangeTypeModified ChangeType = "MODIFIED"
)

type Change struct {
	ChangeType ChangeType
	Path       Path
	Hash       *hashing.Hash
}

// Scans for changes files.  On the first call will return all files as ChangeTypeAdded.
// On subsequent calls will return any changed files (using their hashes)
func (v *VFSScan) Scan() ([]Change, error) {
	allFiles, err := v.Base.ReadTree()
	if err != nil {
		return nil, fmt.Errorf("error reading dir %q: %v", v.Base, err)
	}

	files := make(map[string]Path)
	hashes := make(map[string]*hashing.Hash)
	for _, f := range allFiles {
		key := f.Path()
		files[key] = f
		hasHash, ok := f.(HasHash)
		if !ok {
			return nil, fmt.Errorf("source must support hashing: %T", f)
		}
		hash, err := hasHash.PreferredHash()
		if err != nil {
			return nil, fmt.Errorf("error hashing %q: %v", key, err)
		}

		hashes[key] = hash
	}

	if v.hashes == nil {
		v.hashes = hashes
		var changes []Change
		for k, f := range files {
			hash := hashes[k]
			changes = append(changes, Change{ChangeType: ChangeTypeAdded, Path: f, Hash: hash})
		}
		return changes, nil
	}

	var changes []Change
	for k, f := range files {
		oldHash := v.hashes[k]
		newHash := hashes[k]

		if oldHash == nil {
			changes = append(changes, Change{ChangeType: ChangeTypeAdded, Path: f, Hash: newHash})
		} else if !oldHash.Equal(newHash) {
			changes = append(changes, Change{ChangeType: ChangeTypeModified, Path: f, Hash: newHash})
		}
	}

	for k := range v.hashes {
		newHash := hashes[k]
		f := files[k]
		if newHash == nil {
			changes = append(changes, Change{ChangeType: ChangeTypeRemoved, Path: f, Hash: newHash})
		}
	}

	v.hashes = hashes
	return changes, nil
}

func SyncDir(src *VFSScan, destBase Path) error {
	changes, err := src.Scan()
	if err != nil {
		return fmt.Errorf("error scanning source dir %q: %v", src, err)
	}

	for _, change := range changes {
		f := change.Path

		relativePath, err := RelativePath(f, src.Base)
		if err != nil {
			return err
		}

		destFile := destBase.Join(relativePath)

		switch change.ChangeType {
		case ChangeTypeRemoved:
			err := destFile.Remove()
			if err != nil {
				if !os.IsNotExist(err) {
					return fmt.Errorf("error removing file %q: %v", destFile, err)
				}
			}
			continue

		case ChangeTypeModified, ChangeTypeAdded:
			hashMatch, err := hashesMatch(f, destFile)
			if err != nil {
				return err
			}
			if hashMatch {
				klog.V(2).Infof("File hashes match: %s and %s", f, destFile)
				continue
			}

			if err := CopyFile(f, destFile, nil); err != nil {
				return err
			}

		default:
			return fmt.Errorf("unknown change type: %q", change.ChangeType)
		}

	}

	return nil
}

// CopyFile copies the file at src to dest.  It uses a TempFile, rather than buffering in memory.
func CopyFile(src, dest Path, acl ACL) error {
	tempFile, err := ioutil.TempFile("", "")
	if err != nil {
		return fmt.Errorf("error creating temp file: %v", err)
	}

	defer func() {
		if err := os.Remove(tempFile.Name()); err != nil {
			klog.Warningf("error removing temp file %q: %v", tempFile.Name(), err)
		}
	}()
	defer tempFile.Close()

	if _, err := src.WriteTo(tempFile); err != nil {
		return fmt.Errorf("error reading source file %q: %v", src, err)
	}

	if _, err := tempFile.Seek(0, 0); err != nil {
		return fmt.Errorf("error seeking in temp file during copy: %v", err)
	}

	klog.V(2).Infof("Copying data from %s to %s", src, dest)
	err = dest.WriteFile(tempFile, acl)
	if err != nil {
		return fmt.Errorf("error writing dest file %q: %v", dest, err)
	}

	return nil
}

func hashesMatch(src, dest Path) (bool, error) {
	sh, ok := src.(HasHash)
	if !ok {
		return false, nil
	}

	dh, ok := dest.(HasHash)
	if !ok {
		return false, nil
	}

	{
		srcHash, err := sh.PreferredHash()
		if err != nil {
			klog.Warningf("error getting hash of source file %s: %v", src, err)
		} else if srcHash != nil {
			destHash, err := dh.Hash(srcHash.Algorithm)
			if err != nil {
				klog.Warningf("error comparing hash of dest file %s: %v", dest, err)
			} else if destHash != nil {
				return destHash.Equal(srcHash), nil
			}
		}
	}

	{
		destHash, err := dh.PreferredHash()
		if err != nil {
			klog.Warningf("error getting hash of dest file %s: %v", src, err)
		} else if destHash != nil {
			srcHash, err := dh.Hash(destHash.Algorithm)
			if err != nil {
				klog.Warningf("error comparing hash of src file %s: %v", dest, err)
			} else if srcHash != nil {
				return srcHash.Equal(destHash), nil
			}
		}
	}

	klog.Infof("No compatible hash: %s and %s", src, dest)
	return false, nil
}

// CopyTree copies all files in src to dest.  It copies the whole recursive subtree of files.
func CopyTree(src Path, dest Path, aclOracle ACLOracle) error {
	srcFiles, err := src.ReadTree()
	if err != nil {
		return fmt.Errorf("error reading source directory %q: %v", src, err)
	}

	destFiles, err := dest.ReadTree()
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("error reading source directory %q: %v", src, err)
		}
	}

	destFileMap := make(map[string]Path)
	for _, destFile := range destFiles {
		relativePath, err := RelativePath(dest, destFile)
		if err != nil {
			return err
		}

		destFileMap[relativePath] = destFile
	}

	for _, srcFile := range srcFiles {
		relativePath, err := RelativePath(src, srcFile)
		if err != nil {
			return err
		}

		destFile := destFileMap[relativePath]
		if destFile != nil {
			match, err := hashesMatch(srcFile, destFile)
			if err != nil {
				return err
			}
			if match {
				continue
			}
		}

		destFile = dest.Join(relativePath)

		acl, err := aclOracle(destFile)
		if err != nil {
			return err
		}
		if err := CopyFile(srcFile, destFile, acl); err != nil {
			return err
		}
	}

	return nil
}
