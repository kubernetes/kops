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

package cloudinit

import (
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/utils"
)

type CloudInitTarget struct {
	Config *CloudConfig
	out    io.Writer
	Tags   sets.String
}

type AddBehaviour int

const (
	Always AddBehaviour = iota
	Once
)

func NewCloudInitTarget(out io.Writer, tags sets.String) *CloudInitTarget {
	t := &CloudInitTarget{
		Config: &CloudConfig{},
		out:    out,
		Tags:   tags,
	}
	return t
}

var _ fi.Target = &CloudInitTarget{}

type CloudConfig struct {
	PackageUpdate bool `json:"package_update"`

	Packages     []string           `json:"packages,omitempty"`
	RunCommmands [][]string         `json:"runcmd,omitempty"`
	WriteFiles   []*CloudConfigFile `json:"write_files,omitempty"`
}

type CloudConfigFile struct {
	Encoding    string `json:"encoding,omitempty"`
	Owner       string `json:"owner,omitempty"`
	Path        string `json:"path,omitempty"`
	Permissions string `json:"permissions,omitempty"`
	Content     string `json:"content,omitempty"`
}

func (t *CloudInitTarget) HasTag(tag string) bool {
	_, found := t.Tags[tag]
	return found
}

func (t *CloudInitTarget) ProcessDeletions() bool {
	// We don't expect any, but it would be our job to process them
	return true
}

func (t *CloudInitTarget) AddMkdirpCommand(p string, dirMode os.FileMode) {
	t.AddCommand(Once, "mkdir", "-p", "-m", fi.FileModeToString(dirMode), p)

}
func (t *CloudInitTarget) AddDownloadCommand(addBehaviour AddBehaviour, url string, dest string) {
	// TODO: Create helper to download reliably and validate hash?
	// ... but then why not just use cloudup :-)
	t.AddCommand(addBehaviour, "curl", "-f", "--ipv4", "-Lo", dest, "--connect-timeout", "20", "--retry", "6", "--retry-delay", "10", url)
}

func (t *CloudInitTarget) fetch(p *fi.Source, destPath string) {
	// We could probably move this to fi.Source - it is likely to be the same for every provider
	if p.URL != "" {
		if p.Parent != nil {
			klog.Fatalf("unexpected parent with SourceURL in FetchInstructions: %v", p)
		}
		t.AddDownloadCommand(Once, p.URL, destPath)
	} else if p.ExtractFromArchive != "" {
		if p.Parent == nil {
			klog.Fatalf("unexpected ExtractFromArchive without parent in FetchInstructions: %v", p)
		}

		// TODO: Remove duplicate commands?
		archivePath := "/tmp/" + utils.SanitizeString(p.Parent.Key())
		t.fetch(p.Parent, archivePath)

		extractDir := "/tmp/extracted_" + utils.SanitizeString(p.Parent.Key())
		t.AddMkdirpCommand(extractDir, 0755)
		t.AddCommand(Once, "tar", "zxf", archivePath, "-C", extractDir)

		// Always because this shouldn't happen and we want an indication that it happened
		t.AddCommand(Always, "cp", path.Join(extractDir, p.ExtractFromArchive), destPath)
	} else {
		klog.Fatalf("unknown FetchInstructions: %v", p)
	}
}

func (t *CloudInitTarget) WriteFile(destPath string, contents fi.Resource, fileMode os.FileMode, dirMode os.FileMode) error {
	var p *fi.Source

	if hs, ok := contents.(fi.HasSource); ok {
		p = hs.GetSource()
	}

	if p != nil {
		t.AddMkdirpCommand(path.Dir(destPath), dirMode)
		t.fetch(p, destPath)
	} else {
		// TODO: No way to specify parent dir permissions?
		f := &CloudConfigFile{
			Encoding:    "b64",
			Owner:       "root:root",
			Permissions: fi.FileModeToString(fileMode),
			Path:        destPath,
		}

		d, err := fi.ResourceAsBytes(contents)
		if err != nil {
			return err
		}

		// Not a strict limit, just a sanity check
		if len(d) > 256*1024 {
			return fmt.Errorf("resource is very large (failed sanity-check): %v", contents)
		}

		f.Content = base64.StdEncoding.EncodeToString(d)

		t.Config.WriteFiles = append(t.Config.WriteFiles, f)
	}
	return nil
}

func (t *CloudInitTarget) Chown(path string, user, group string) {
	t.AddCommand(Always, "chown", user+":"+group, path)
}

func (t *CloudInitTarget) AddCommand(addBehaviour AddBehaviour, args ...string) {
	switch addBehaviour {
	case Always:
		break

	case Once:
		for _, c := range t.Config.RunCommmands {
			if utils.StringSlicesEqual(args, c) {
				klog.V(2).Infof("skipping pre-existing command because AddBehaviour=Once: %q", args)
				return
			}
		}

	default:
		klog.Fatalf("unknown AddBehaviour: %v", addBehaviour)
	}

	t.Config.RunCommmands = append(t.Config.RunCommmands, args)
}

func (t *CloudInitTarget) Finish(taskMap map[string]fi.Task) error {
	d, err := utils.YamlMarshal(t.Config)
	if err != nil {
		return fmt.Errorf("error serializing config to yaml: %v", err)
	}

	conf := "#cloud-config\n" + string(d)

	_, err = t.out.Write([]byte(conf))
	if err != nil {
		return fmt.Errorf("error writing cloud-init data to output: %v", err)
	}
	return nil
}
