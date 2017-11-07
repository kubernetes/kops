# Kubernetes Reference Docs

Tools to build reference documentation for Kubernetes APIs and CLIs.

# Api Docs

## Generate new api docs

1. Update the Makefile for your environment
  - Set `K8SIOROOT` to the kubernetes/kubernetes.github.io root directory
    - If you have not already, clone the kubernetes/kubernetes.github.io repo
  - Set `K8SROOT` to the kubernetes/kubernetes root directory
    - If you have not already, clone the kubernetes/kubernetes repo
  - Set `MINOR_VERSION` to the kubernetes/kubernetes minor version (e.g. 8)

2. Copy the `swagger.json` file from Kubernetes.
  - Go to K8SROOT directory and checkout the `release-<Minor>` branch
  - Run `make updateapispec` to copy the `swagger.json` for the kubernetes release

3. Run `make api` to build the doc html and javascript (will fail)

4. Update the file `gen_open_api/config.yaml` to fix listed errors
  - New APIs will be listed as "Orphaned" types, and should be added to the ToC
  - New versions for existing APIs will be listed as "Orphaned" types, and should replace
    existing versions in the ToC
  - **Note:** These should only appear for API types, if field types are listed as
    Orphaned, then there is an issue with the swagger.json compatibility
    
5. Run `make api` again after fixing the errors

6. Create a new directory under kubernetes/kubernetes.github.io and
   update kubernetes/kubernetes.github.io
  - From kubernetes/kubernetes.io `mkdir docs/api-reference/v1.<Minor>`
  - Update `docs/reference/index.md` to include the new version
  - Grep for references to the old version, and update them

6. Run `map copyapi` to copy the files to the kubernetes/kubernetes.github.io directory

7. Update the left pannel in kubernetes/kubernetes.io
  - Edit `_data/reference.yml` and add a section for the new kubectl docs

# Cli

## Generate new kubectl docs

1. Update the Makefile for your environment
  - Set `K8SIOROOT` to the kubernetes/kubernetes.io root directory
  - Set `K8SROOT` to the kubernetes/kubernetes root directory
  - Set `MINOR_VERSION` to the kubernetes/kubernetes minor version (e.g. 8)

2. Create a new directory for the kubernetes version and copy over the contents from
   the last release
  - `mkdir gen-kubectldocs/generators/v1_<Minor>`
  - `cp -r gen-kubectldocs/generators/v1_<Minor-1>/* gen-kubectldocs/generators/v1_<Minor>`

2. Update the Kubernetes vendored src code to be for the current release
  - Edit `glide.yaml` to change the `version` to *release-1.<Minor>*
  - Run `glide update -v` to update the vendored code
  - Commit the modified files

3. Build the cli docs
  - Run `make cli`
  - Files will be copied under `gen-kubectldocs/generators/build/`
  - Open up `index.html` in a browser and make sure it looks good

4. Fix any errors about things missing from the `gen-kubectldocs/generatorstoc.yaml`
  - Commands appearing in the ToC that have been removed from kubectl should be removed from the toc
  - Commands appearing in kubectl but missing from the ToC should be added to a section in the `toc.yaml`

5. Build the docs again after the errors have been fixed

6. Copy the cli docs to kubernetes/kubernetes.github.io
  - From kubernetes/kubernetes.github.io `mkdir docs/user-guide/kubectl/v1.8`
  - Run `make copycli`
  - Files will be copied to the appropriate directory under `K8SIOROOT`
  - You may need to create a new directory for new kubernetes versions

7. Update the left pannel in kubernetes/kubernetes.io
  - Edit `_data/reference.yml` and add a section for the new kubectl docs

# Updating brodocs version

*May need to change the image repo to one you have write access to.*

1. Update Dockerfile so it will re-clone the repo

2. Run `make brodocs`
