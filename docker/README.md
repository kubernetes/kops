# Running Kops in Docker

The Dockerfile here is offered primarily as a way to build continuous
integration versions of `kops` until we figure out how we want to
release/package it.

To use it, e.g. (assumes your `$HOME` is correct and that `$KOPS_STATE_STORE` is correct):
```shell
$ docker build -t kops .
$ KOPS="docker run -v $HOME/.aws:/root/.aws:ro -v $HOME/.ssh:/root/.ssh:ro -v $HOME/.kube:/root/.kube -it kops kops --state=$KOPS_STATE_STORE"
```

This creates a shell variable that runs the `kops` container with `~/.aws` mounted in (for AWS credentials), `~/.ssh` mounted in (for SSH keys, for AWS specifically), and `~/.kube` mounted in (so `kubectl` can add newly created clusters).

After this, you can just use `$KOPS` where you would generally use `kops`, e.g. `$KOPS get cluster`.
