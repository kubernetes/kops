This directory contains integration tests for weave.

## Requirements

You need three VMs with `docker` (`>=1.10.0`) installed and listening on TCP
port `2375` (see below). You also need to be able to SSH to these VMs without 
having to input anything.

The `Vagrantfile` in this directory constructs three such VMs.

If you are [building Weave Net using Vagrant](https://www.weave.works/docs/net/latest/building/),
it is recommended to run the tests from the build VM and not the host.


## Running tests

**TL;DR**: You can run the steps 2. to 7. below with one command: 

    make PROVIDER=vagrant integration-tests

**Detailed steps**:

  1. Start the build virtual machine (see above article for more details):

        vagrant up

  2. Start the three testing VMs:

        cd test
        vagrant up

  3. SSH into the build VM and go to Weave Net's sources:

        cd ..
        vagrant ssh
        # you are now on the build VM:
        cd ~/weave 

  4. Compile all code and dependencies:

        make
        make testrunner
        cd test

  5. Upload the weave images from where the `Makefile` puts them (`weave.tar.gz`) to 
     the three docker hosts, `docker load` these, and copies the `weave` script over:

        ./setup.sh

  6. Run individual tests, e.g.:

        ./200_dns_test.sh

     or run all tests (everything named `*_test.sh`):

        ./run_all.sh

  7. Stop all VMs:

        exit
        # you are now on your host machine
        vagrant destroy -f
        cd test
        vagrant destroy -f


## Using other VMs

By default the tests assume the Vagrant VMs are used.

To use other VMs, set the environment variable <var>HOSTS</var> to the
space-separated list of IP addresses of the docker hosts, and set the
environment variable <var>SSH</var> to a command that will log into
either (which may just be `ssh`).

## Making docker available over TCP

To make docker listen to a TCP socket, you will usually need to either
run it manually with an option like `-H tcp://0.0.0.0:2375`; or, for
apt-get installed docker (Ubuntu and Debian), add the line

```
DOCKER_OPTS="--host unix:///var/run/docker.sock --host tcp://0.0.0.0:2375"
```

to the file `/etc/default/docker`, then restart docker.

## Updating the GCE test image

When a new version of Docker is released, you willneed to update the GCE test image.
To do this, change the Docker version in `run-integration-tests.sh` and push the change.
Next build in CircleCI will detect that there is no template for this version of Docker and will first create the template before running tests.
Subsequent builds will then simply re-use the template.
