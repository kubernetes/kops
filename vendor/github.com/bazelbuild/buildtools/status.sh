#!/bin/bash

set -e

buildifier_tags=$(git describe --tags)
IFS='-' read -a parse_tags <<< "$buildifier_tags"
echo "buildifierVersion ${parse_tags[0]}"

buildifier_rev=$(git rev-parse HEAD)
echo "buildScmRevision ${buildifier_rev}"
