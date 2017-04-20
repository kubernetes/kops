---
title: Building Weave Net
menu_order: 120
---
You only need to build Weave Net if you want to work on the Weave Net codebase
(or you just enjoy building software).

Apart from the `weave` shell script, weave is delivered as a set of
container images.  There is no distribution-specific packaging, so in
principle it shouldn't matter which Linux distribution you build
under.  But naturally, Docker is a prerequisite (version 1.6.0 or
later).

The only way to build is by using the the build container; the
Makefile is setup to make this transparent.  This method is
documented below, and can be run directly on your machine or using
a Vagrant VM.

## <a name="ubuntu"></a>Building directly on your machine

The weave git repository should be cloned into
`$GOPATH/src/github.com/weaveworks/weave`, in accordance with [the go
workspace conventions](https://golang.org/doc/code.html#Workspaces):

```
$ WEAVE=github.com/weaveworks/weave
$ git clone https://$WEAVE $GOPATH/src/$WEAVE
$ cd $GOPATH/src/$WEAVE
```

Next install Docker if you haven't already, by following the instructions
[on the Docker site](https://docs.docker.com/installation/ubuntulinux/).

Then to actually build, simply do:

```
$ make
```

On a fresh repository, the Makefile will do the following:
- assemble the build container
- download specific versions of all the dependencies
- build the weave components in the build container
- package them into three Docker images (`weaveworks/weave`,
`weaveworks/weaveexec`, and `weaveworks/plugin`)
- Exported these images as `weave.tar.gz`

The first two steps may take a while - don't worry, they are
are cached and should not need to be redone very often.

## <a name="vagrant"></a>Building using Vagrant

If you aren't running Linux, or otherwise don't want to run the Docker
daemon outside a VM, you can use
[Vagrant](https://www.vagrantup.com/downloads.html) to run a
development environment. You'll probably need to install
[VirtualBox](https://www.virtualbox.org/wiki/Downloads) too, for
Vagrant to run VMs in.

First, check out the code:

```
$ git clone https://github.com/weaveworks/weave
$ cd weave
```

The `Vagrantfile` in the top directory constructs a VM that has

 * docker installed
 * go tools installed
 * weave dependencies installed
 * $GOPATH set to ~
 * the local working directory mapped as a synced folder into the
   right place in $GOPATH

Once you are in the working directory you can issue

```
$ vagrant up
```

and wait for a while (don't worry, the long download and package
installation is done just once). The working directory is sync'ed with
`~/src/github.com/weaveworks/weave` on the VM, so you can edit files and
use git and so on in the regular filesystem.

To build and run the code, you need to use the VM. To log in and build
the weave image, do

```
$ vagrant ssh
vm$ cd src/github.com/weaveworks/weave
vm$ make
```

The Docker daemon is also running in this VM, so you can then do

```
vm$ sudo ./weave launch
vm$ sudo docker ps
```

and so on.

If you are looking to just do a build and not run anything on this VM,
you can do so with

```
$ vagrant ssh -c 'make -C src/github.com/weaveworks/weave'
```

you should then find a `weave.tar.gz` container snapshot tarball in the
top-level directory. You can use that snapshot with `docker load`
against a different host, e.g.

```
$ export DOCKER_HOST=tcp://<HOST:PORT>
$ docker load < weave.tar.gz
```

You can provide extra Vagrant configuration by putting a file
`Vagrant.local` in the same place as `Vagrantfile`; for instance, to
forward additional ports.
