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

package init_repo

import (
	"archive/tar"
	"compress/gzip"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var vendorInstallCmd = &cobra.Command{
	Use:   "dep",
	Short: "Install Gopkg.toml, Gopkg.lock and vendor/.",
	Long:  `Install Gopkg.toml, Gopkg.lock and vendor/.`,
	Example: `# Bootstrap vendor/ from the src packaged with apiserver-boot
apiserver-boot init dep
`,
	Run: RunVendorInstall,
}

var builderCommit string
var Update bool

func AddVendorInstallCmd(cmd *cobra.Command) {
	vendorInstallCmd.Flags().StringVar(&builderCommit, "commit", "", "if specified with fetch, use this commit for the apiserver-builder deps")
	vendorInstallCmd.Flags().BoolVar(&Update, "update", false, "if true, don't touch Gopkg.toml or Gopkg.lock, and replace versions of packages managed by apiserver-boot.")
	cmd.AddCommand(vendorInstallCmd)
}

// DeleteOld delete all the versions for all packages it is going to untar
func DeleteOld() {
	// Move up two directories from the location of the `apiserver-boot`
	// executable to find the `vendor` directory we package with our
	// releases.
	e, err := os.Executable()
	if err != nil {
		log.Fatal("unable to get directory of apiserver-builder tools")
	}

	e = filepath.Dir(filepath.Dir(e))

	// read the file
	f := filepath.Join(e, "bin", "vendor.tar.gz")
	fr, err := os.Open(f)
	if err != nil {
		log.Fatalf("failed to read vendor tar file %s %v", f, err)
	}
	defer fr.Close()

	// setup gzip of tar
	gr, err := gzip.NewReader(fr)
	if err != nil {
		log.Fatalf("failed to read vendor tar file %s %v", f, err)
	}
	defer gr.Close()

	// setup tar reader
	tr := tar.NewReader(gr)

	for file, err := tr.Next(); err == nil; file, err = tr.Next() {
		p := filepath.Join(".", file.Name)
		// Delete existing directory first if upgrading
		if filepath.Dir(p) != "." {
			dir := filepath.Base(filepath.Dir(p))
			parent := filepath.Base(filepath.Dir(filepath.Dir(p)))
			gparent := filepath.Base(filepath.Dir(filepath.Dir(filepath.Dir(p))))

			// Delete the directory if it is a repo or package in a repo
			if dir != "vendor" && parent != "vendor" && !(gparent == "vendor" && parent == "github.com") {
				os.RemoveAll(filepath.Dir(p))
			}
		}
	}
}

func RunVendorInstall(cmd *cobra.Command, args []string) {
	// Delete old versions of the packages we manage before installing the new ones
	if Update {
		DeleteOld()
	}

	// Move up two directories from the location of the `apiserver-boot`
	// executable to find the `vendor` directory we package with our
	// releases.
	e, err := os.Executable()
	if err != nil {
		log.Fatal("unable to get directory of apiserver-builder tools")
	}

	e = filepath.Dir(filepath.Dir(e))

	// read the file
	f := filepath.Join(e, "bin", "vendor.tar.gz")
	fr, err := os.Open(f)
	if err != nil {
		log.Fatalf("failed to read vendor tar file %s %v", f, err)
	}
	defer fr.Close()

	// setup gzip of tar
	gr, err := gzip.NewReader(fr)
	if err != nil {
		log.Fatalf("failed to read vendor tar file %s %v", f, err)
	}
	defer gr.Close()

	// setup tar reader
	tr := tar.NewReader(gr)

	for file, err := tr.Next(); err == nil; file, err = tr.Next() {
		p := filepath.Join(".", file.Name)

		if Update && filepath.Dir(p) == "." {
			continue
		}

		err := os.MkdirAll(filepath.Dir(p), 0700)
		if err != nil {
			log.Fatalf("Could not create directory %s: %v", filepath.Dir(p), err)
		}
		b, err := ioutil.ReadAll(tr)
		if err != nil {
			log.Fatalf("Could not read file %s: %v", file.Name, err)
		}
		err = ioutil.WriteFile(p, b, os.FileMode(file.Mode))
		if err != nil {
			log.Fatalf("Could not write file %s: %v", p, err)
		}
	}
}
