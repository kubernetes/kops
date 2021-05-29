/*
Copyright 2020 The Kubernetes Authors.

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

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
	"sigs.k8s.io/yaml"

	"k8s.io/kops/util/pkg/tables"

	"github.com/spf13/cobra"
	"k8s.io/kops/cmd/kops/util"
	"k8s.io/kops/upup/pkg/fi/cloudup"
)

type Image struct {
	Image  string `json:"image"`
	Mirror string `json:"mirror"`
}

type File struct {
	File   string `json:"file"`
	Mirror string `json:"mirror"`
	SHA    string `json:"sha"`
}

type AssetResult struct {
	// Images are the image assets we use (output).
	Images []*Image `json:"images,omitempty"`
	// FileAssets are the file assets we use (output).
	Files []*File `json:"files,omitempty"`
}

func NewCmdGetAssets(f *util.Factory, out io.Writer, options *GetOptions) *cobra.Command {
	getAssetsShort := i18n.T(`Display assets for cluster.`)

	getAssetsLong := templates.LongDesc(i18n.T(`
	Display assets for cluster.`))

	getAssetsExample := templates.Examples(i18n.T(`
	# Display all assets.
	kops get assets
	`))

	cmd := &cobra.Command{
		Use:     "assets",
		Short:   getAssetsShort,
		Long:    getAssetsLong,
		Example: getAssetsExample,
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.TODO()

			if err := rootCommand.ProcessArgs(args); err != nil {
				exitWithError(err)
			}

			err := RunGetAssets(ctx, f, out, options)
			if err != nil {
				exitWithError(err)
			}
		},
	}

	return cmd
}

func RunGetAssets(ctx context.Context, f *util.Factory, out io.Writer, options *GetOptions) error {

	clusterName := rootCommand.ClusterName()
	options.clusterName = clusterName
	if clusterName == "" {
		return fmt.Errorf("--name is required")
	}

	updateClusterResults, err := RunUpdateCluster(ctx, f, clusterName, out, &UpdateClusterOptions{
		Target: cloudup.TargetDryRun,
		Phase:  string(cloudup.PhaseStageAssets),
		Quiet:  true,
	})
	if err != nil {
		return err
	}

	result := AssetResult{
		Images: make([]*Image, 0, len(updateClusterResults.ImageAssets)),
		Files:  make([]*File, 0, len(updateClusterResults.FileAssets)),
	}

	seen := map[string]bool{}
	for _, containerAsset := range updateClusterResults.ImageAssets {
		image := Image{
			Image:  containerAsset.CanonicalLocation,
			Mirror: containerAsset.DockerImage,
		}
		if !seen[image.Image] {
			result.Images = append(result.Images, &image)
			seen[image.Image] = true
		}
	}
	seen = map[string]bool{}
	for _, fileAsset := range updateClusterResults.FileAssets {
		file := File{
			File:   fileAsset.CanonicalURL.String(),
			Mirror: fileAsset.DownloadURL.String(),
			SHA:    fileAsset.SHAValue,
		}
		if !seen[file.File] {
			result.Files = append(result.Files, &file)
			seen[file.File] = true
		}
	}

	switch options.output {
	case OutputTable:
		if err = containerOutputTable(result.Images, out); err != nil {
			return err
		}
		return fileOutputTable(result.Files, out)
	case OutputYaml:
		y, err := yaml.Marshal(result)
		if err != nil {
			return fmt.Errorf("unable to marshal YAML: %v", err)
		}
		if _, err := out.Write(y); err != nil {
			return fmt.Errorf("error writing to output: %v", err)
		}
	case OutputJSON:
		j, err := json.Marshal(result)
		if err != nil {
			return fmt.Errorf("unable to marshal JSON: %v", err)
		}
		if _, err := out.Write(j); err != nil {
			return fmt.Errorf("error writing to output: %v", err)
		}
	default:
		return fmt.Errorf("unsupported output format: %q", options.output)
	}

	return nil
}

func containerOutputTable(images []*Image, out io.Writer) error {
	fmt.Println("")
	t := &tables.Table{}
	t.AddColumn("IMAGE", func(i *Image) string {
		return i.Image
	})
	t.AddColumn("MIRROR", func(i *Image) string {
		return i.Mirror
	})

	columns := []string{"IMAGE", "MIRROR"}
	return t.Render(images, out, columns...)
}

func fileOutputTable(files []*File, out io.Writer) error {
	fmt.Println("")
	t := &tables.Table{}
	t.AddColumn("FILE", func(f *File) string {
		return f.File
	})
	t.AddColumn("MIRROR", func(f *File) string {
		return f.Mirror
	})
	t.AddColumn("SHA", func(f *File) string {
		return f.SHA
	})

	columns := []string{"FILE", "MIRROR", "SHA"}
	return t.Render(files, out, columns...)
}
