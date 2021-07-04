/*
Copyright 2021 The Kubernetes Authors.

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

	"k8s.io/kops/pkg/assets"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
	"sigs.k8s.io/yaml"

	"k8s.io/kops/util/pkg/tables"

	"github.com/spf13/cobra"
	"k8s.io/kops/cmd/kops/util"
	"k8s.io/kops/upup/pkg/fi/cloudup"
)

type GetAssetsOptions struct {
	*GetOptions
	Copy bool
}

type Image struct {
	Canonical string `json:"canonical"`
	Download  string `json:"download"`
}

type File struct {
	Canonical string `json:"canonical"`
	Download  string `json:"download"`
	SHA       string `json:"sha"`
}

type AssetResult struct {
	// Images are the image assets we use (output).
	Images []*Image `json:"images,omitempty"`
	// FileAssets are the file assets we use (output).
	Files []*File `json:"files,omitempty"`
}

func NewCmdGetAssets(f *util.Factory, out io.Writer, getOptions *GetOptions) *cobra.Command {
	options := GetAssetsOptions{
		GetOptions: getOptions,
	}

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

			err := RunGetAssets(ctx, f, out, &options)
			if err != nil {
				exitWithError(err)
			}
		},
	}

	cmd.Flags().BoolVar(&options.Copy, "copy", options.Copy, "copy assets to local repository")

	return cmd
}

func RunGetAssets(ctx context.Context, f *util.Factory, out io.Writer, options *GetAssetsOptions) error {

	clusterName := rootCommand.ClusterName(true)
	options.clusterName = clusterName
	if clusterName == "" {
		return fmt.Errorf("--name is required")
	}

	updateClusterResults, err := RunUpdateCluster(ctx, f, out, &UpdateClusterOptions{
		Target:      cloudup.TargetDryRun,
		GetAssets:   true,
		ClusterName: clusterName,
	})
	if err != nil {
		return err
	}

	result := AssetResult{
		Images: make([]*Image, 0, len(updateClusterResults.ImageAssets)),
		Files:  make([]*File, 0, len(updateClusterResults.FileAssets)),
	}

	seen := map[string]bool{}
	for _, imageAsset := range updateClusterResults.ImageAssets {
		image := Image{
			Canonical: imageAsset.CanonicalLocation,
			Download:  imageAsset.DownloadLocation,
		}
		if !seen[image.Canonical] {
			result.Images = append(result.Images, &image)
			seen[image.Canonical] = true
		}
	}

	seen = map[string]bool{}
	for _, fileAsset := range updateClusterResults.FileAssets {
		file := File{
			Canonical: fileAsset.CanonicalURL.String(),
			Download:  fileAsset.DownloadURL.String(),
			SHA:       fileAsset.SHAValue,
		}
		if !seen[file.Canonical] {
			result.Files = append(result.Files, &file)
			seen[file.Canonical] = true
		}
	}

	if options.Copy {
		err := assets.Copy(updateClusterResults.ImageAssets, updateClusterResults.FileAssets, updateClusterResults.Cluster)
		if err != nil {
			return err
		}
	}

	switch options.output {
	case OutputTable:
		if err = imageOutputTable(result.Images, out); err != nil {
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

func imageOutputTable(images []*Image, out io.Writer) error {
	fmt.Println("")
	t := &tables.Table{}
	t.AddColumn("CANONICAL", func(i *Image) string {
		return i.Canonical
	})
	t.AddColumn("DOWNLOAD", func(i *Image) string {
		return i.Download
	})

	columns := []string{"CANONICAL", "DOWNLOAD"}
	return t.Render(images, out, columns...)
}

func fileOutputTable(files []*File, out io.Writer) error {
	fmt.Println("")
	t := &tables.Table{}
	t.AddColumn("CANONICAL", func(f *File) string {
		return f.Canonical
	})
	t.AddColumn("DOWNLOAD", func(f *File) string {
		return f.Download
	})
	t.AddColumn("SHA", func(f *File) string {
		return f.SHA
	})

	columns := []string{"CANONICAL", "DOWNLOAD", "SHA"}
	return t.Render(files, out, columns...)
}
