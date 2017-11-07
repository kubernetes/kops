# Installing the apiserver build tools

Requires the following to already be installed:
- kubectl
- go
- docker
- openssl
- base64
- glide (optional)
- tar
- cp

Installation instructions:

- Download the latest [release](https://github.com/kubernetes-incubator/apiserver-builder/releases)
- Extract to `/usr/local/apiserver-builder/`
  - Create this directory if it does not already exist
- Add `/usr/local/apiserver-builder/bin` to your path
  `export PATH=$PATH:/usr/local/apiserver-builder/bin`
- Test things are working by running `apiserver-boot -h`
