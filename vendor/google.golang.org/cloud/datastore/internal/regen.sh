#!/bin/bash -e
#
# This script rebuilds the generated code for the protocol buffers.
# To run this you will need protoc and goprotobuf installed;
# see https://github.com/golang/protobuf for instructions.
# You also need Go and Git installed.

PKG=google.golang.org/cloud/datastore
UPSTREAM=https://github.com/googleapis/googleapis
UPSTREAM_SUBDIR=google
PB_UPSTREAM=https://github.com/google/protobuf
PB_UPSTREAM_SUBDIR=src

function die() {
  echo 1>&2 $*
  exit 1
}

# Sanity check that the right tools are accessible.
for tool in go git protoc protoc-gen-go; do
  q=$(which $tool) || die "didn't find $tool"
  echo 1>&2 "$tool: $q"
done

tmpdir=$(mktemp -d -t regen-cds.XXXXXX)
trap 'rm -rf $tmpdir' EXIT
tmpproto=$(mktemp -d -t regen-cds.XXXXXX)
trap 'rm -rf $tmpproto' EXIT

echo -n 1>&2 "finding package dir... "
pkgdir=$(go list -f '{{.Dir}}' $PKG)
echo 1>&2 $pkgdir
base=$(echo $pkgdir | sed "s,/$PKG\$,,")
echo 1>&2 "base: $base"
cd $base

echo 1>&2 "fetching latest datastore protos... "
git clone -q $UPSTREAM $tmpdir

echo 1>&2 "fetching latest core protos..."
git clone -q $PB_UPSTREAM $tmpproto

# Pass 1: build mapping from upstream filename to our filename.
declare -A filename_map
for f in $(cd $PKG && find internal -name '*.proto'); do
  echo -n 1>&2 "looking for latest version of $f... "
  up=$(cd $tmpdir/$UPSTREAM_SUBDIR && find * -name $(basename $f))
  echo 1>&2 $up
  if [ $(echo $up | wc -w) != "1" ]; then
    die "not exactly one match"
  fi
  filename_map[$up]=$f
done

# Pass 2: build sed script for fixing imports.
import_fixes=$tmpdir/fix_imports.sed
for up in "${!filename_map[@]}"; do
  f=${filename_map[$up]}
  echo >>$import_fixes "s,\"$UPSTREAM_SUBDIR/$up\",\"$PKG/$f\","
done
cat $import_fixes | sed 's,^,### ,' 1>&2

# Pass 3: copy files, making necessary adjustments.
for up in "${!filename_map[@]}"; do
  f=${filename_map[$up]}
  cat $tmpdir/$UPSTREAM_SUBDIR/$up |
    # Adjust proto imports.
    sed -f $import_fixes |
    # Drop unused imports.
    sed '/import "google\/api/d' |
    # Drop java options.
    sed '/option java/d' |
    # Drop the HTTP annotations.
    sed '/option.*google\.api\.http.*{.*};$/d' |
    cat > $PKG/$f
done

# Mappings of well-known proto types.
declare -A known_types
known_types[google/protobuf/struct.proto]=github.com/golang/protobuf/ptypes/struct
known_types[google/protobuf/timestamp.proto]=github.com/golang/protobuf/ptypes/timestamp
known_types[google/protobuf/wrappers.proto]=github.com/golang/protobuf/ptypes/wrappers
types_map=""
for f in "${!known_types[@]}"; do
  pkg=${known_types[$f]}
  types_map="$types_map,M$f=$pkg"
done

# Run protoc once per package.
for dir in $(find $PKG/internal -name '*.proto' | xargs dirname | sort | uniq); do
  echo 1>&2 "* $dir"
  protoc -I "$tmpproto/$PB_UPSTREAM_SUBDIR" -I . --go_out=plugins=grpc$types_map:. $dir/*.proto
done
echo 1>&2 "All OK"
