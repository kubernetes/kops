#!/usr/bin/env bash

# Copyright 2026 The Kubernetes Authors.
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

REPO_ROOT=$(git rev-parse --show-toplevel)
RUN_TEST="${REPO_ROOT}/tests/e2e/scenarios/scalability/run-test.sh"
TEST_ROOT=$(mktemp -d)
trap 'rm -rf "${TEST_ROOT}"' EXIT

fail() {
  echo "FAIL: $*" >&2
  exit 1
}

run_case() {
  local scenario="$1"
  local expected_exit_code="$2"
  local case_root="${TEST_ROOT}/${scenario}-${expected_exit_code}"
  local artifacts="${case_root}/artifacts"
  local args_file="${case_root}/kubetest2-args"
  local output_file="${case_root}/output"
  local variant="${scenario}-variant"
  local expected_tester

  mkdir -p \
    "${artifacts}" \
    "${case_root}/bin" \
    "${case_root}/home" \
    "${case_root}/gopath/src/k8s.io/perf-tests/clusterloader2/testing/load"

  cat >"${case_root}/bin/kubetest2" <<'EOF'
#!/usr/bin/env bash
printf '%s\n' "$@" >"${KUBETEST2_ARGS_FILE}"
printf '{"kubetest-version":"stub","tester-version":"stub"}\n' >"${ARTIFACTS}/metadata.json"
exit "${KUBETEST2_EXIT_CODE}"
EOF
  chmod +x "${case_root}/bin/kubetest2"

  local actual_exit_code=0
  PATH="${case_root}/bin:${PATH}" \
    HOME="${case_root}/home" \
    GOPATH="${case_root}/gopath" \
    ARTIFACTS="${artifacts}" \
    JOB_TYPE=periodic \
    REPO_OWNER=kubernetes \
    REPO_NAME=kops \
    CLOUD_PROVIDER=aws \
    CNI_PLUGIN=calico \
    SCALE_SCENARIO="${scenario}" \
    EXPERIMENT_VARIANT="${variant}" \
    KUBETEST2_ARGS_FILE="${args_file}" \
    KUBETEST2_EXIT_CODE="${expected_exit_code}" \
    "${RUN_TEST}" >"${output_file}" 2>&1 || actual_exit_code=$?

  if [[ "${actual_exit_code}" -ne "${expected_exit_code}" ]]; then
    fail "${scenario}: exit code ${actual_exit_code}, want ${expected_exit_code}; output: ${output_file}"
  fi
  if grep -q -- '--metadata' "${args_file}"; then
    fail "${scenario}: kubetest2 received the unsupported --metadata flag"
  fi

  if [[ "${scenario}" == "correctness" ]]; then
    expected_tester="--test=kops"
  else
    expected_tester="--test=clusterloader2"
  fi
  grep -qx -- "${expected_tester}" "${args_file}" || fail "${scenario}: missing ${expected_tester}"

  jq -e --arg variant "${variant}" \
    '.variant == $variant and .["kubetest-version"] == "stub" and .["tester-version"] == "stub"' \
    "${artifacts}/metadata.json" >/dev/null || fail "${scenario}: metadata was not merged"

  echo "PASS: ${scenario} preserves kubetest2 exit ${expected_exit_code} and merges variant metadata"
}

run_case correctness 0
run_case correctness 42
run_case performance 0
run_case performance 42
