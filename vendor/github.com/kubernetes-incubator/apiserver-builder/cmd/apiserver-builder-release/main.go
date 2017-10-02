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

package main

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var targets []string
var output string
var dovendor bool
var test bool
var version string
var commit string
var userLocalVendor bool
var useBazel bool

var cachevendordir string

var DefaultTargets = []string{"linux:amd64", "darwin:amd64", "windows:amd64"}

func main() {
	buildCmd.Flags().StringSliceVar(&targets, "targets",
		DefaultTargets, "GOOS:GOARCH pair.  maybe specified multiple times.")
	buildCmd.Flags().StringVar(&cachevendordir, "vendordir", "",
		"if specified, use this directory for setting up vendor instead of creating a tmp directory.")
	buildCmd.Flags().StringVar(&output, "output", "apiserver-builder",
		"value name of the tar file to build")
	buildCmd.Flags().StringVar(&version, "version", "", "version name")
	buildCmd.Flags().BoolVar(&useBazel, "bazel", false, "use bazel to compile (faster, but no X-compile)")

	buildCmd.Flags().BoolVar(&dovendor, "vendor", true, "if true, fetch packages to vendor")
	buildCmd.Flags().BoolVar(&test, "test", true, "if true, run tests")
	cmd.AddCommand(buildCmd)

	vendorCmd.Flags().StringVar(&commit, "commit", "", "apiserver-builder commit")
	vendorCmd.Flags().StringVar(&version, "version", "", "version name")
	vendorCmd.Flags().StringVar(&cachevendordir, "vendordir", "",
		"if specified, use this directory for setting up vendor instead of creating a tmp directory.")
	vendorCmd.Flags().BoolVar(&userLocalVendor, "use-local-vendor", true, "if true, run use the local vendored code")
	cmd.AddCommand(vendorCmd)

	installCmd.Flags().StringVar(&version, "version", "", "version name")
	cmd.AddCommand(installCmd)

	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

var cmd = &cobra.Command{
	Use:   "apiserver-builder-release",
	Short: "apiserver-builder-release builds a .tar.gz release package",
	Long:  `apiserver-builder-release builds a .tar.gz release package`,
	Run:   RunMain,
}

func RunMain(cmd *cobra.Command, args []string) {
	cmd.Help()
}

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "build the binaries",
	Long:  `build the binaries`,
	Run:   RunBuild,
}

func TmpDir() string {
	dir, err := ioutil.TempDir(os.TempDir(), "apiserver-builder-release")
	if err != nil {
		log.Fatalf("failed to create temp directory %s %v", dir, err)
	}

	dir, err = filepath.EvalSymlinks(dir)
	if err != nil {
		log.Fatal(err)
	}

	err = os.Mkdir(filepath.Join(dir, "src"), 0700)
	if err != nil {
		log.Fatalf("failed to create directory %s %v", filepath.Join(dir, "src"), err)
	}

	err = os.Mkdir(filepath.Join(dir, "bin"), 0700)
	if err != nil {
		log.Fatalf("failed to create directory %s %v", filepath.Join(dir, "bin"), err)
	}
	return dir
}

func RunBuild(cmd *cobra.Command, args []string) {
	if len(version) == 0 {
		log.Fatal("must specify the --version flag")
	}
	if len(targets) == 0 {
		log.Fatal("must provide at least one --targets flag when building tools")
	}

	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	dir = filepath.Join(dir, "release", version)
	vendor := filepath.Join(dir, "src")

	if _, err := os.Stat(vendor); os.IsNotExist(err) {
		log.Fatalf("must first run `apiserver-builder-release vendor`.  could not find %s", vendor)
	}

	if !useBazel {
		// Build binaries for the targeted platforms in then tar
		for _, target := range targets {
			// Build binaries for this os:arch
			parts := strings.Split(target, ":")
			if len(parts) != 2 {
				log.Fatalf("--targets flags must be GOOS:GOARCH pairs [%s]", target)
			}
			goos := parts[0]
			goarch := parts[1]
			// Cleanup old binaries
			os.RemoveAll(filepath.Join(dir, "bin"))
			err := os.Mkdir(filepath.Join(dir, "bin"), 0700)
			if err != nil {
				log.Fatalf("failed to create directory %s %v", filepath.Join(dir, "bin"), err)
			}

			BuildVendorTar(dir)

			for _, pkg := range VendoredBuildPackages {
				Build(filepath.Join("cmd", "vendor", pkg, "main.go"),
					filepath.Join(dir, "bin", filepath.Base(pkg)),
					goos, goarch,
				)
			}
			for _, pkg := range OwnedBuildPackages {
				Build(filepath.Join(pkg, "main.go"),
					filepath.Join(dir, "bin", filepath.Base(pkg)),
					goos, goarch,
				)
			}
			PackageTar(goos, goarch, dir, vendor)
		}
	} else {
		os.MkdirAll(filepath.Join(dir, "bin"), 0700)
		BuildVendorTar(dir)
		BazelBuildCopy(dir, []string{
			"//cmd/apiregister-gen",
			"//cmd/apiserver-boot",
			"//cmd/vendor/github.com/kubernetes-incubator/reference-docs/gen-apidocs",
			"//cmd/vendor/k8s.io/kubernetes/cmd/libs/go2idl/client-gen",
			"//cmd/vendor/k8s.io/kubernetes/cmd/libs/go2idl/conversion-gen",
			"//cmd/vendor/k8s.io/kubernetes/cmd/libs/go2idl/deepcopy-gen",
			"//cmd/vendor/k8s.io/kubernetes/cmd/libs/go2idl/defaulter-gen",
			"//cmd/vendor/k8s.io/kubernetes/cmd/libs/go2idl/informer-gen",
			"//cmd/vendor/k8s.io/kubernetes/cmd/libs/go2idl/lister-gen",
			"//cmd/vendor/k8s.io/kubernetes/cmd/libs/go2idl/openapi-gen",
		}...)
		PackageTar("", "", dir, vendor)
	}
}

func BazelBuildCopy(dest string, targets ...string) {
	args := append([]string{"build"}, targets...)
	c := exec.Command("bazel", args...)
	RunCmd(c, "")

	// Copy the binaries
	for _, t := range targets {
		name := filepath.Base(t)
		c := exec.Command("cp", filepath.Join("bazel-bin", t, name), filepath.Join(dest, "bin", name))
		RunCmd(c, "")
	}
}

func RunCmd(cmd *exec.Cmd, gopath string) {
	setgopath := len(gopath) > 0
	gopath, err := filepath.Abs(gopath)
	if err != nil {
		log.Fatal(err)
	}
	gopath, err = filepath.EvalSymlinks(gopath)
	if err != nil {
		log.Fatal(err)
	}
	if setgopath {
		cmd.Env = append(cmd.Env, fmt.Sprintf("GOPATH=%s", gopath))
	}
	for _, v := range os.Environ() {
		if strings.HasPrefix(v, "GOPATH=") && setgopath {
			continue
		}
		cmd.Env = append(cmd.Env, v)
	}

	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if len(cmd.Dir) == 0 && len(gopath) > 0 {
		cmd.Dir = gopath
	}
	fmt.Printf("%s\n", strings.Join(cmd.Args, " "))
	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}

func Build(input, output, goos, goarch string) {
	cmd := exec.Command("go", "build", "-o", output, input)

	// CGO_ENABLED=0 for statically compile binaries
	cmd.Env = []string{"CGO_ENABLED=0"}
	if len(goos) > 0 {
		cmd.Env = append(cmd.Env, fmt.Sprintf("GOOS=%s", goos))
	}
	if len(goarch) > 0 {
		cmd.Env = append(cmd.Env, fmt.Sprintf("GOARCH=%s", goarch))
	}
	for _, v := range os.Environ() {
		if strings.HasPrefix(v, "CGO_ENABLED=") {
			continue
		}
		if strings.HasPrefix(v, "GOOS=") && len(goos) > 0 {
			continue
		}
		if strings.HasPrefix(v, "GOARCH=") && len(goarch) > 0 {
			continue
		}
		cmd.Env = append(cmd.Env, v)
	}
	RunCmd(cmd, "")
}

var VendoredBuildPackages = []string{
	"github.com/kubernetes-incubator/reference-docs/gen-apidocs",
	"k8s.io/kubernetes/cmd/libs/go2idl/client-gen",
	"k8s.io/kubernetes/cmd/libs/go2idl/conversion-gen",
	"k8s.io/kubernetes/cmd/libs/go2idl/deepcopy-gen",
	"k8s.io/kubernetes/cmd/libs/go2idl/defaulter-gen",
	"k8s.io/kubernetes/cmd/libs/go2idl/informer-gen",
	"k8s.io/kubernetes/cmd/libs/go2idl/lister-gen",
	"k8s.io/kubernetes/cmd/libs/go2idl/openapi-gen",
}

var OwnedBuildPackages = []string{
	"cmd/apiregister-gen",
	"cmd/apiserver-boot",
}

func BuildVendorTar(dir string) {
	// create the new file
	f := filepath.Join(dir, "bin", "glide.tar.gz")
	fw, err := os.Create(f)
	if err != nil {
		log.Fatalf("failed to create vendor tar file %s %v", f, err)
	}
	defer fw.Close()

	// setup gzip of tar
	gw := gzip.NewWriter(fw)
	defer gw.Close()

	// setup tar writer
	tw := tar.NewWriter(gw)
	defer tw.Close()

	srcdir := filepath.Join(dir)
	filepath.Walk(srcdir, TarFile{
		tw,
		0555,
		filepath.Join(srcdir, "src"),
		"",
	}.Do)
}

func PackageTar(goos, goarch, tooldir, vendordir string) {
	// create the new file
	fw, err := os.Create(fmt.Sprintf("%s-%s-%s-%s.tar.gz", output, version, goos, goarch))
	if err != nil {
		log.Fatalf("failed to create output file %s %v", output, err)
	}
	defer fw.Close()

	// setup gzip of tar
	gw := gzip.NewWriter(fw)
	defer gw.Close()

	// setup tar writer
	tw := tar.NewWriter(gw)
	defer tw.Close()

	// Add all of the bin files
	// Add all of the bin files
	filepath.Walk(filepath.Join(tooldir, "bin"), TarFile{
		tw,
		0555,
		tooldir,
		"",
	}.Do)
}

type TarFile struct {
	Writer *tar.Writer
	Mode   int64
	Root   string
	Parent string
}

func (t TarFile) Do(path string, info os.FileInfo, err error) error {
	if info.IsDir() {
		return nil
	}

	eval, err := filepath.EvalSymlinks(path)
	if err != nil {
		log.Fatal(err)
	}
	if eval != path {
		name := strings.Replace(path, t.Root, "", -1)
		if len(t.Parent) != 0 {
			name = filepath.Join(t.Parent, name)
		}
		linkName := strings.Replace(eval, t.Root, "", -1)
		if len(t.Parent) != 0 {
			linkName = filepath.Join(t.Parent, linkName)
		}
		hdr := &tar.Header{
			Name:     name,
			Mode:     t.Mode,
			Linkname: linkName,
		}
		if err := t.Writer.WriteHeader(hdr); err != nil {
			log.Fatalf("failed to write output for %s %v", path, err)
		}
		return nil
	}

	return t.Write(path)
}

func (t TarFile) Write(path string) error {
	// Get the relative name of the file
	name := strings.Replace(path, t.Root, "", -1)
	if len(t.Parent) != 0 {
		name = filepath.Join(t.Parent, name)
	}
	body, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("failed to read file %s %v", path, err)
	}
	if len(body) == 0 {
		return nil
	}

	hdr := &tar.Header{
		Name: name,
		Mode: t.Mode,
		Size: int64(len(body)),
	}
	if err := t.Writer.WriteHeader(hdr); err != nil {
		log.Fatalf("failed to write output for %s %v", path, err)
	}
	if _, err := t.Writer.Write(body); err != nil {
		log.Fatalf("failed to write output for %s %v", path, err)
	}
	return nil
}

var vendorCmd = &cobra.Command{
	Use:   "vendor",
	Short: "create vendored libraries for release",
	Long:  `create vendored libraries for release`,
	Run:   RunVendor,
}

func RunVendor(cmd *cobra.Command, args []string) {
	if len(version) == 0 {
		log.Fatal("must specify the --version flag")
	}

	// Create the release directory
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	dir = filepath.Join(dir, "release", version)
	os.MkdirAll(dir, 0700)

	if userLocalVendor {
		BuildLocalVendor(dir)
	} else {
		//Build binaries for the current platform so that we can use them
		for _, pkg := range VendoredBuildPackages {
			Build(filepath.Join("vendor", pkg, "main.go"),
				filepath.Join(dir, "bin", filepath.Base(pkg)),
				"", "",
			)
		}
		for _, pkg := range OwnedBuildPackages {
			Build(filepath.Join(pkg, "main.go"),
				filepath.Join(dir, "bin", filepath.Base(pkg)),
				"", "",
			)
		}
		BuildVendor(dir)
	}
}

func BuildLocalVendor(tooldir string) {
	os.MkdirAll(filepath.Join(tooldir, "src"), 0700)
	c := exec.Command("cp", "-R", "-H",
		filepath.Join("vendor"),
		filepath.Join(tooldir, "src", "vendor"))
	RunCmd(c, "")
	os.MkdirAll(filepath.Join(tooldir, "src", "vendor", "github.com", "kubernetes-incubator", "apiserver-builder"), 0700)
	c = exec.Command("cp", "-R", "-H",
		filepath.Join("pkg"),
		filepath.Join(tooldir, "src", "vendor", "github.com", "kubernetes-incubator", "apiserver-builder", "pkg"))
	RunCmd(c, "")

	c = exec.Command("bash", "-c",
		fmt.Sprintf("find %s -name BUILD.bazel| xargs sed -i='' s'|//pkg|//vendor/github.com/kubernetes-incubator/apiserver-builder/pkg|g'",
			filepath.Join(tooldir, "src", "vendor", "github.com", "kubernetes-incubator", "apiserver-builder", "pkg"),
		))
	RunCmd(c, "")

	c = exec.Command("cp", "-R", "-H",
		filepath.Join("glide.yaml"),
		filepath.Join(tooldir, "src", "glide.yaml"))
	RunCmd(c, "")
	c = exec.Command("cp", "-R", "-H",
		filepath.Join("glide.lock"),
		filepath.Join(tooldir, "src", "glide.lock"))
	RunCmd(c, "")

}

func BuildVendor(tooldir string) string {
	vendordir := cachevendordir
	if len(vendordir) == 0 {
		vendordir = TmpDir()
		fmt.Printf("to rerun with cached glide use `--vendordir %s`\n", vendordir)
	}

	if len(vendordir) == 0 && len(commit) == 0 {
		log.Fatal("must specify the --commit flag")
	}

	vendordir, err := filepath.EvalSymlinks(vendordir)
	if err != nil {
		log.Fatal(err)
	}

	pkgDir := filepath.Join(vendordir, "src", "github.com", "kubernetes-incubator", "test")
	bootBin := filepath.Join(tooldir, "bin", "apiserver-boot")
	err = os.MkdirAll(pkgDir, 0700)
	if err != nil {
		log.Fatalf("failed to create directory %s %v", pkgDir, err)
	}

	ioutil.WriteFile(filepath.Join(pkgDir, "boilerplate.go.txt"), []byte(""), 0555)

	os.RemoveAll(filepath.Join(pkgDir, "pkg"))
	os.RemoveAll(filepath.Join(pkgDir, "docs"))
	os.RemoveAll(filepath.Join(pkgDir, "main.go"))

	cmd := exec.Command(bootBin, "init", "repo", "--domain", "k8s.io", "--install-deps=false")
	cmd.Dir = pkgDir
	RunCmd(cmd, vendordir)

	cmd = exec.Command(bootBin, "create", "group", "version", "resource", "--group", "misk", "--version", "v1beta1", "--kind", "Student")
	cmd.Dir = pkgDir
	RunCmd(cmd, vendordir)

	cmd = exec.Command(bootBin, "init", "glide", "--fetch", "--commit", commit)
	cmd.Dir = pkgDir
	RunCmd(cmd, vendordir)

	// Copy the vendored libraries.  This will make it easier to debug if there is a test failure.
	os.MkdirAll(filepath.Join(tooldir, "src"), 0700)
	c := exec.Command("cp", "-R", "-H",
		filepath.Join(pkgDir, "vendor"),
		filepath.Join(tooldir, "src", "vendor"))
	RunCmd(c, "")
	c = exec.Command("cp", "-R", "-H",
		filepath.Join(pkgDir, "glide.yaml"),
		filepath.Join(tooldir, "src", "glide.yaml"))
	RunCmd(c, "")
	c = exec.Command("cp", "-R", "-H",
		filepath.Join(pkgDir, "glide.lock"),
		filepath.Join(tooldir, "src", "glide.lock"))
	RunCmd(c, "")

	if test {
		cmd = exec.Command(bootBin, "build", "executables")
		cmd.Dir = pkgDir
		RunCmd(cmd, vendordir)

		cmd = exec.Command("go", "test", "./"+filepath.Join("pkg", "..."))
		cmd.Dir = pkgDir
		RunCmd(cmd, vendordir)
	}

	return pkgDir
}

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "install release locally",
	Long:  `install release locally`,
	Run:   RunInstall,
}

func RunInstall(cmd *cobra.Command, args []string) {
	if len(version) == 0 {
		log.Fatal("must specify the --version flag")
	}

	// Untar to to /usr/local/apiserver-build/
	os.Mkdir(filepath.Join("/", "usr", "local", "apiserver-builder"), 0700)
	c := exec.Command("tar", "-xzvf", fmt.Sprintf("%s-%s-%s-%s.tar.gz", output, version, "", ""),
		"-C", filepath.Join("/", "usr", "local", "apiserver-builder"),
	)
	RunCmd(c, "")

}
