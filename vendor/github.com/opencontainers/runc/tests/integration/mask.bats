#!/usr/bin/env bats

load helpers

function setup() {
  teardown_busybox
  setup_busybox
}

function teardown() {
  teardown_busybox
}

@test "MaskPaths(file)" {
  # run busybox detached
  runc run -d --console /dev/pts/ptmx test_busybox
  [ "$status" -eq 0 ]

  wait_for_container 15 1 test_busybox

  runc exec test_busybox cat /proc/kcore
  [ "$status" -eq 0 ]
  [[ "${output}" == "" ]]

  runc exec test_busybox rm -f /proc/kcore
  [ "$status" -eq 1 ]
  [[ "${output}" == *"Permission denied"* ]]

  runc exec test_busybox umount /proc/kcore
  [ "$status" -eq 1 ]
  [[ "${output}" == *"Operation not permitted"* ]]
}

@test "MaskPaths(directory)" {
  # run busybox detached
  runc run -d --console /dev/pts/ptmx test_busybox
  [ "$status" -eq 0 ]

  wait_for_container 15 1 test_busybox

  runc exec test_busybox ls /sys/firmware
  [ "$status" -eq 0 ]
  [[ "${output}" == "" ]]

  runc exec test_busybox touch /sys/firmware/foo
  [ "$status" -eq 1 ]
  [[ "${output}" == *"Read-only file system"* ]]

  runc exec test_busybox rm -rf /sys/firmware
  [ "$status" -eq 1 ]
  [[ "${output}" == *"Read-only file system"* ]]

  runc exec test_busybox umount /sys/firmware
  [ "$status" -eq 1 ]
  [[ "${output}" == *"Operation not permitted"* ]]
}
