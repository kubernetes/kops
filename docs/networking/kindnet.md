# Kindnet

***Kindnet support is experimental, and may be removed at any time***

## Introduction

* [kindnet](http://kindnet.es)

Kindnet focuses on providing essential networking functionality without unnecessary complexity.

## Installing

To install [kindnet](https://github.com/aojea/kindnet) - use `--networking kindnet`.

```sh
export ZONES=mylistofzone
kops create cluster \
  --zones $ZONES \
  --networking kindnet \
  --yes \
  --name myclustername.mydns.io
```

## Getting help

For problems with kindnet please post an issue to Github:

- [Kindnet Issues](https://github.com/aojea/kindnet/issues)

You can learn more about the different configurations options in https://kindnet.es/
