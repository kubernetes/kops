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
	"strings"

	"github.com/golang/glog"
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
	// @justinsb not sure I like this but the problem is knowing if the sha is a file or a string
	// so I did a location or a sha.  We have both SHAs in remote file locations, and SHAs in strings
	SourceShaLocation *string
	SourceSha         *string
	TargetSha         *string
	Lifecycle         *fi.Lifecycle
}

var _ fi.CompareWithID = &CopyFile{}

func (e *CopyFile) CompareWithID() *string {
	return e.Name
}

func (e *CopyFile) Find(c *fi.Context) (*CopyFile, error) {
	// TODO do a head call on the file??
	return nil, nil

}

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
	sourceSha := fi.StringValue(e.SourceSha)
	sourceSHALocation := fi.StringValue(e.SourceShaLocation)

	glog.Infof("copying bits from %q to %q", source, target)

	if err := transferFile(c, source, target, sourceSha, sourceSHALocation); err != nil {
		return fmt.Errorf("unable to transfer %q to %q: %v", source, target, err)
	}

	return nil
}

func transferFile(c *fi.Context, source string, target string, sourceSHA string, sourceSHALocation string) error {
	data, err := vfs.Context.ReadFile(source)
	if err != nil {
		return fmt.Errorf("Error unable to read path %q: %v", source, err)
	}

	uploadVFS, err := buildVFSPath(target)
	if err != nil {
		return err
	}

	// Test the file matches the sourceSha
	{
		var sha string
		if sourceSHA != "" {
			sha = sourceSHA
		} else if sourceSHALocation != "" {
			shaBytes, err := vfs.Context.ReadFile(sourceSHALocation)
			if err != nil {
				return fmt.Errorf("Error unable to read path %q: %v", source, err)
			}
			sha = string(shaBytes)
		}

		if sha != "" {

			trimmedSHA := strings.TrimSpace(sha)

			in := bytes.NewReader(data)
			dataHash, err := hashing.HashAlgorithmSHA1.Hash(in)
			if err != nil {
				return fmt.Errorf("unable to hash sha from data: %v", err)
			}

			shaHash, err := hashing.FromString(trimmedSHA)
			if err != nil {
				return fmt.Errorf("unable to hash sha: %q, %v", sha, err)
			}

			if !shaHash.Equal(dataHash) {
				return fmt.Errorf("SHAs are not matching for %q", dataHash.String())
			}

			shaVFS, err := buildVFSPath(target + ".sha1")
			if err != nil {
				return err
			}

			b := bytes.NewBufferString(sha)
			if err := writeFile(c, shaVFS, b.Bytes()); err != nil {
				return fmt.Errorf("Error uploading file %q: %v", shaVFS, err)
			}
		}
	}

	if err := writeFile(c, uploadVFS, data); err != nil {
		return fmt.Errorf("Error uploading file %q: %v", uploadVFS, err)
	}

	return nil
}

func writeFile(c *fi.Context, vfsPath string, data []byte) error {
	glog.V(2).Infof("uploading to %q", vfsPath)
	p, err := vfs.Context.BuildVfsPath(vfsPath)
	if err != nil {
		return fmt.Errorf("error building path %q: %v", vfsPath, err)
	}

	acl, err := acls.GetACL(p, c.Cluster)
	if err != nil {
		return err
	}

	if err = p.WriteFile(data, acl); err != nil {
		return fmt.Errorf("error writing path %q: %v", vfsPath, err)
	}

	glog.V(2).Infof("upload complete: %q", vfsPath)

	return nil
}

func buildVFSPath(target string) (string, error) {
	if !strings.Contains(target, "://") || strings.HasPrefix(target, "memfs://") {
		return target, nil
	}

	u, err := url.Parse(target)
	if err != nil {
		return "", fmt.Errorf("unable to parse: %q", target)
	}

	var vfsPath string

	// remove the filename from the end of the path
	//pathSlice := strings.Split(u.Path, "/")
	//pathSlice = pathSlice[:len(pathSlice)-1]
	//path := strings.Join(pathSlice, "/")

	// TODO I am not a huge fan of this, but it would work
	// TODO @justinsb should we require an URL?
	if u.Host == "s3.amazonaws.com" {
		vfsPath = "s3:/" + u.Path
	} else if u.Host == "storage.googleapis.com" {
		vfsPath = "gs:/" + u.Path
	}

	if vfsPath == "" {
		glog.Errorf("unable to determine vfs path s3, google storage, and file paths are supported")
		glog.Errorf("for s3 use s3.amazonaws.com and for google storage use storage.googleapis.com hostnames.")
		return "", fmt.Errorf("unable to determine vfs for %q:", target)
	}

	return vfsPath, nil
}
