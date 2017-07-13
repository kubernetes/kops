#!/usr/bin/env bats

load helpers

function setup() {
  teardown_busybox
  setup_busybox
}

function teardown() {
  teardown_busybox
}

@test "runc start" {
  runc create --console /dev/pts/ptmx test_busybox1
  [ "$status" -eq 0 ]

  testcontainer test_busybox1 created

  runc create --console /dev/pts/ptmx test_busybox2
  [ "$status" -eq 0 ]

  testcontainer test_busybox2 created


  # start conatiner test_busybox1 and test_busybox2
  runc start test_busybox1 test_busybox2
  [ "$status" -eq 0 ]

  testcontainer test_busybox1 running
  testcontainer test_busybox2 running

  # delete test_busybox1 and test_busybox2
  runc delete --force test_busybox1 test_busybox2

  runc state test_busybox1
  [ "$status" -ne 0 ]

  runc state test_busybox2
  [ "$status" -ne 0 ]
}
