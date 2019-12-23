#!/usr/bin/env bash

# Copyright 2014 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -o errexit
set -o nounset
set -o pipefail

KUBE_ROOT=$(git rev-parse --show-toplevel)

kube::util::array_contains() {
  local search="$1"
  local element
  shift
  for element; do
    if [[ "${element}" == "${search}" ]]; then
      return 0
    fi
  done
  return 1
}
function kube::util::read-array {
  local i=0
  unset -v "$1"
  while IFS= read -r "$1[i++]"; do :; done
  eval "[[ \${$1[--i]} ]]" || unset "$1[i]" # ensures last element isn't empty
}

FOCUS="${1:-}"
FOCUS="${FOCUS%/}" # Remove the ending "/"

# See https://staticcheck.io/docs/checks
CHECKS=(
  "all"
  "-S1*"   # Omit code simplifications for now.
  "-ST1*"  # Mostly stylistic, redundant w/ golint
)
export IFS=','; checks="${CHECKS[*]}"; unset IFS

# Packages to ignore due to bugs in staticcheck
# NOTE: To ignore issues detected a package, add it to the .staticcheck_failures blacklist
IGNORE=(
)
export IFS='|'; ignore_pattern="^(${IGNORE[*]})\$"; unset IFS

# Ensure that we find the binaries we build before anything else.
export GOBIN="${KUBE_ROOT}/_output/bin"
PATH="${GOBIN}:${PATH}"

# Install staticcheck from vendor
echo 'installing staticcheck from vendor'

go install k8s.io/kops/vendor/honnef.co/go/tools/cmd/staticcheck

cd "${KUBE_ROOT}"

# Check that the file is in alphabetical order
failure_file="${KUBE_ROOT}/hack/.staticcheck_failures"
if ! diff -u "${failure_file}" <(LC_ALL=C sort "${failure_file}"); then
  {
    echo
    echo "${failure_file} is not in alphabetical order. Please sort it:"
    echo
    echo "  LC_ALL=C sort -o ${failure_file} ${failure_file}"
    echo
  } >&2
  false
fi

all_packages=()
while IFS='' read -r line; do
  # Prepend './' to get staticcheck to treat these as paths, not packages.
  all_packages+=("./$line")
done < <( find -H . -type f -name \*.go | sed 's|/[^/]*$||' | sed 's|^./||' | LC_ALL=C sort -u |
            grep "^${FOCUS:-.}" |
            grep -vE "(third_party|generated|clientset_generated|vendor|/_)" |
            grep -vE "$ignore_pattern" )

failing_packages=()
if [[ -z $FOCUS ]]; then # Ignore failing_packages in FOCUS mode
  while IFS='' read -r line; do failing_packages+=("$line"); done < <(cat "$failure_file")
fi
errors=()
not_failing=()

while read -r error; do
  # Ignore compile errors caused by lack of files due to build tags.
  # TODO: Add verification for these directories.
  ignore_no_files="^-: build constraints exclude all Go files in .* \(compile\)"
  if [[ $error =~ $ignore_no_files ]]; then
    continue
  fi

  file="${error%%:*}"
  pkg="$(dirname "$file")"
  kube::util::array_contains "$pkg" "${failing_packages[@]}" && in_failing=$? || in_failing=$?
  if [[ "${in_failing}" -ne "0" ]]; then
    errors+=( "${error}" )
  elif [[ "${in_failing}" -eq "0" ]]; then
    really_failing+=( "$pkg" )
  fi
done < <(staticcheck -checks "${checks}" "${all_packages[@]}" 2>/dev/null || true)
echo staticcheck -checks "${checks}" "${all_packages[@]}"
export IFS=$'\n'  # Expand ${really_failing[*]} to separate lines
kube::util::read-array really_failing < <(sort -u <<<"${really_failing[*]}")
unset IFS
for pkg in "${failing_packages[@]}"; do
  if ! kube::util::array_contains "$pkg" "${really_failing[@]}"; then
    not_failing+=( "$pkg" )
  fi
done

# Check that all failing_packages actually still exist
gone=()
for p in "${failing_packages[@]}"; do
  if ! kube::util::array_contains "./$p" "${all_packages[@]}"; then
    gone+=( "$p" )
  fi
done

# Check to be sure all the packages that should pass check are.
if [ ${#errors[@]} -eq 0 ]; then
  echo 'Congratulations!  All Go source files have passed staticcheck.'
else
  {
    echo "Errors from staticcheck:"
    for err in "${errors[@]}"; do
      echo "$err"
    done
    echo
    echo 'Please review the above warnings. You can test via:'
    echo '  hack/verify-staticcheck.sh <failing package>'
    echo 'If the above warnings do not make sense, you can exempt the line or file. See:'
    echo '  https://staticcheck.io/docs/#ignoring-problems'
    echo
  } >&2
  exit 1
fi

if [[ ${#not_failing[@]} -gt 0 ]]; then
  {
    echo "Some packages in hack/.staticcheck_failures are passing staticcheck. Please remove them."
    echo
    for p in "${not_failing[@]}"; do
      echo "  $p"
    done
    echo
  } >&2
  exit 1
fi

if [[ ${#gone[@]} -gt 0 ]]; then
  {
    echo "Some packages in hack/.staticcheck_failures do not exist anymore. Please remove them."
    echo
    for p in "${gone[@]}"; do
      echo "  $p"
    done
    echo
  } >&2
  exit 1
fi