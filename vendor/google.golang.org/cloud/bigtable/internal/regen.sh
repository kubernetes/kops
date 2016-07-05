#!/bin/bash -e
#
# This script rebuilds the generated code for the protocol buffers.
# To run this you will need protoc and goprotobuf installed;
# see https://github.com/golang/protobuf for instructions.
# You also need Go and Git installed.

PKG=google.golang.org/cloud/bigtable
UPSTREAM=https://github.com/GoogleCloudPlatform/cloud-bigtable-client
UPSTREAM_SUBDIR=bigtable-protos/src/main/proto
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

tmpdir=$(mktemp -d -t regen-cbt.XXXXXX)
trap 'rm -rf $tmpdir' EXIT
tmpproto=$(mktemp -d -t regen-cds.XXXXXX)
trap 'rm -rf $tmpproto' EXIT

echo -n 1>&2 "finding package dir... "
pkgdir=$(go list -f '{{.Dir}}' $PKG)
echo 1>&2 $pkgdir
base=$(echo $pkgdir | sed "s,/$PKG\$,,")
echo 1>&2 "base: $base"
cd $base

echo 1>&2 "fetching latest protos... "
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
  echo >>$import_fixes "s,\"$up\",\"$PKG/$f\","
done
cat $import_fixes | sed 's,^,### ,' 1>&2
# Pass 3: copy files, making necessary adjustments.
for up in "${!filename_map[@]}"; do
  f=${filename_map[$up]}
  cat $tmpdir/$UPSTREAM_SUBDIR/$up |
    # Adjust proto imports.
    sed -f $import_fixes |
    # Drop long-running cluster and instance RPC methods. They return a google.longrunning.Operation.
    sed '/^  rpc UndeleteCluster(/,/^  }$/d' |
    sed '/^  rpc CreateInstance(/,/^  }$/d' |
    sed '/^  rpc CreateCluster(/,/^  }$/d' |
    sed '/^  rpc UpdateCluster(/,/^  }$/d' |
    # Drop annotations and long-running operations. They aren't supported (yet).
    sed '/"google\/longrunning\/operations.proto"/d' |
    sed '/google.longrunning.Operation/d' |
    sed '/"google\/api\/annotations.proto"/d' |
    sed '/option.*google\.api\.http.*{.*};$/d' |
    cat > $PKG/$f
done

# Mappings of well-known proto types.
declare -A known_types
known_types[google/protobuf/any.proto]=github.com/golang/protobuf/ptypes/any
known_types[google/protobuf/duration.proto]=github.com/golang/protobuf/ptypes/duration
known_types[google/protobuf/timestamp.proto]=github.com/golang/protobuf/ptypes/timestamp
known_types[google/protobuf/empty.proto]=github.com/golang/protobuf/ptypes/empty
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
