/*
Copyright 2017 The Kubernetes Authors.

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

package assettasks

import (
	"bytes"
	"fmt"
	"net/url"
	"os"
	"strings"

	"k8s.io/klog"
	"k8s.io/kops/pkg/acls"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/util/pkg/hashing"
	"k8s.io/kops/util/pkg/vfs"
)

// CopyFile copies an from a source file repository, to a target repository,
// typically used for highly secure clusters.
//go:generate fitask -type=CopyFile
type CopyFile struct {
	Name       *string
	SourceFile *string
	TargetFile *string
	SHA        *string
	Lifecycle  *fi.Lifecycle
}

var _ fi.CompareWithID = &CopyFile{}

func (e *CopyFile) CompareWithID() *string {
	// or should this be the SHA?
	return e.Name
}

// fileExtensionForSHA returns the expected extension for the given hash
// If the hash length is not recognized, it returns an error.
func fileExtensionForSHA(sha string) (string, error) {
	switch len(sha) {
	case 40:
		return ".sha1", nil
	case 64:
		return ".sha256", nil
	default:
		return "", fmt.Errorf("unhandled sha length for %q", sha)
	}
}

// Find attempts to find a file.
func (e *CopyFile) Find(c *fi.Context) (*CopyFile, error) {
	expectedSHA := strings.TrimSpace(fi.StringValue(e.SHA))

	shaExtension, err := fileExtensionForSHA(expectedSHA)
	if err != nil {
		return nil, err
	}

	targetSHAFile := fi.StringValue(e.TargetFile) + shaExtension

	targetSHABytes, err := vfs.Context.ReadFile(targetSHAFile)
	if err != nil {
		if os.IsNotExist(err) {
			klog.V(4).Infof("unable to download: %q, assuming target file is not present, and if not present may not be an error: %v",
				targetSHAFile, err)
			return nil, nil
		} else {
			klog.V(4).Infof("unable to download: %q, %v", targetSHAFile, err)
			// TODO should we throw err here?
			return nil, nil
		}
	}
	targetSHA := string(targetSHABytes)

	if strings.TrimSpace(targetSHA) == expectedSHA {
		actual := &CopyFile{
			Name:       e.Name,
			TargetFile: e.TargetFile,
			SHA:        e.SHA,
			SourceFile: e.SourceFile,
			Lifecycle:  e.Lifecycle,
		}
		klog.V(8).Infof("found matching target sha1 for file: %q", fi.StringValue(e.TargetFile))
		return actual, nil
	}

	klog.V(8).Infof("did not find same file, found mismatching target sha1 for file: %q", fi.StringValue(e.TargetFile))
	return nil, nil

}

// Run is the default run method.
func (e *CopyFile) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (s *CopyFile) CheckChanges(a, e, changes *CopyFile) error {
	if fi.StringValue(e.Name) == "" {
		return fi.RequiredField("Name")
	}
	if fi.StringValue(e.SourceFile) == "" {
		return fi.RequiredField("SourceFile")
	}
	if fi.StringValue(e.TargetFile) == "" {
		return fi.RequiredField("TargetFile")
	}
	return nil
}

func (_ *CopyFile) Render(c *fi.Context, a, e, changes *CopyFile) error {

	source := fi.StringValue(e.SourceFile)
	target := fi.StringValue(e.TargetFile)
	sourceSha := fi.StringValue(e.SHA)

	klog.V(2).Infof("copying bits from %q to %q", source, target)

	if err := transferFile(c, source, target, sourceSha); err != nil {
		return fmt.Errorf("unable to transfer %q to %q: %v", source, target, err)
	}

	return nil
}

// transferFile downloads a file from the source location, validates the file matches the SHA,
// and uploads the file to the target location.
func transferFile(c *fi.Context, source string, target string, sha string) error {

	// TODO drop file to disk, as vfs reads file into memory.  We load kubelet into memory for instance.
	// TODO in s3 can we do a copy file ... would need to test

	data, err := vfs.Context.ReadFile(source)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file not found %q: %v", source, err)
		}

		return fmt.Errorf("error downloading file %q: %v", source, err)
	}

	objectStore, err := buildVFSPath(target)
	if err != nil {
		return err
	}

	uploadVFS, err := vfs.Context.BuildVfsPath(objectStore)
	if err != nil {
		return fmt.Errorf("error building path %q: %v", objectStore, err)
	}

	shaExtension, err := fileExtensionForSHA(sha)
	if err != nil {
		return err
	}

	shaTarget := objectStore + shaExtension
	shaVFS, err := vfs.Context.BuildVfsPath(shaTarget)
	if err != nil {
		return fmt.Errorf("error building path %q: %v", shaTarget, err)
	}

	shaHash, err := hashing.FromString(strings.TrimSpace(sha))
	if err != nil {
		return fmt.Errorf("unable to parse sha: %q, %v", sha, err)
	}

	in := bytes.NewReader(data)
	dataHash, err := shaHash.Algorithm.Hash(in)
	if err != nil {
		return fmt.Errorf("unable to hash file %q downloaded: %v", source, err)
	}

	if !shaHash.Equal(dataHash) {
		return fmt.Errorf("the sha value in %q does not match %q calculated value %q", shaTarget, source, dataHash.String())
	}

	klog.Infof("uploading %q to %q", source, objectStore)
	if err := writeFile(c, uploadVFS, data); err != nil {
		return err
	}

	b := []byte(shaHash.Hex())
	if err := writeFile(c, shaVFS, b); err != nil {
		return err
	}

	return nil
}

func writeFile(c *fi.Context, p vfs.Path, data []byte) error {

	acl, err := acls.GetACL(p, c.Cluster)
	if err != nil {
		return err
	}

	if err = p.WriteFile(bytes.NewReader(data), acl); err != nil {
		return fmt.Errorf("error writing path %v: %v", p, err)
	}

	return nil
}

// buildVFSPath task a recognizable https url and transforms that URL into the equivalent url with the object
// store prefix.
func buildVFSPath(target string) (string, error) {
	if !strings.Contains(target, "://") || strings.HasPrefix(target, "memfs://") {
		return target, nil
	}

	var vfsPath string

	// Matches all S3 regional naming conventions:
	// https://docs.aws.amazon.com/general/latest/gr/rande.html#s3_region
	// and converts to a s3://<bucket>/<path> vfsPath
	s3VfsPath, err := vfs.VFSPath(target)
	if err == nil {
		vfsPath = s3VfsPath
	} else {
		// These matches only cover a subset of the URLs that you can use, but I am uncertain how to cover more of the possible
		// options.
		// This code parses the HOST and determines gs URLs.
		// For instance you can have the bucket name in the gs url hostname.
		u, err := url.Parse(target)
		if err != nil {
			return "", fmt.Errorf("Unable to parse Google Cloud Storage URL: %q", target)
		}
		if u.Host == "storage.googleapis.com" {
			vfsPath = "gs:/" + u.Path
		}
	}

	if vfsPath == "" {
		klog.Errorf("Unable to determine VFS path from supplied URL: %s", target)
		klog.Errorf("S3, Google Cloud Storage, and File Paths are supported.")
		klog.Errorf("For S3, please make sure that the supplied file repository URL adhere to S3 naming conventions, https://docs.aws.amazon.com/general/latest/gr/rande.html#s3_region.")
		klog.Errorf("For GCS, please make sure that the supplied file repository URL adheres to https://storage.googleapis.com/")
		if err != nil { // print the S3 error for more details
			return "", fmt.Errorf("Error Details: %v", err)
		}
		return "", fmt.Errorf("unable to determine vfs type for %q", target)
	}

	return vfsPath, nil
}
