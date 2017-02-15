# Troubleshooting 

## Installation

##### Make sure `$GOPATH` is set, and your [workspace](https://golang.org/doc/code.html#Workspaces) is configured.

##### kops will not compile with symlinks in `$GOPATH`. See issue go issue [17451](https://github.com/golang/go/issues/17451) for more information

##### kops uses the relatively new Go vendoring, so building requires Go 1.6 or later. 

##### Kops will only compile if the source is checked out in `$GOPATH/src/k8s.io/kops`. If you try to use `$GOPATH/src/github.com/kubernetes/kops` you will run into issues with package imports not working as expected.