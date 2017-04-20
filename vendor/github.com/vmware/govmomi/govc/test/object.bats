#!/usr/bin/env bats

load test_helper

@test "object.destroy" {
    run govc object.destroy "/enoent"
    assert_failure

    run govc object.destroy
    assert_failure

    vm=$(new_id)
    run govc vm.create "$vm"
    assert_success

    # fails when powered on
    run govc object.destroy "vm/$vm"
    assert_failure

    run govc vm.power -off "$vm"
    assert_success

    run govc object.destroy "vm/$vm"
    assert_success
}

@test "object.rename" {
  run govc object.rename "/enoent" "nope"
  assert_failure

  vm=$(new_id)
  run govc vm.create -on=false "$vm"
  assert_success

  run govc object.rename "vm/$vm" "${vm}-renamed"
  assert_success

  run govc object.rename "vm/$vm" "${vm}-renamed"
  assert_failure

  run govc object.destroy "vm/${vm}-renamed"
  assert_success
}

@test "object.mv" {
  vcsim_env

  folder=$(new_id)

  run govc folder.create "vm/$folder"
  assert_success

  for _ in $(seq 1 3) ; do
    vm=$(new_id)
    run govc vm.create -folder "$folder" "$vm"
    assert_success
  done

  result=$(govc ls "vm/$folder" | wc -l)
  [ "$result" -eq "3" ]

  run govc folder.create "vm/${folder}-2"
  assert_success

  run govc object.mv "vm/$folder/*" "vm/${folder}-2"
  assert_success

  result=$(govc ls "vm/${folder}-2" | wc -l)
  [ "$result" -eq "3" ]

  result=$(govc ls "vm/$folder" | wc -l)
  [ "$result" -eq "0" ]
}

@test "object.collect" {
  run govc object.collect
  assert_success

  run govc object.collect -json
  assert_success

  run govc object.collect -
  assert_success

  run govc object.collect -json -
  assert_success

  run govc object.collect - content
  assert_success

  run govc object.collect -json - content
  assert_success

  root=$(govc object.collect - content | grep content.rootFolder | awk '{print $3}')

  dc=$(govc object.collect "$root" childEntity | awk '{print $3}' | cut -d, -f1)

  hostFolder=$(govc object.collect "$dc" hostFolder | awk '{print $3}')

  cr=$(govc object.collect "$hostFolder" childEntity | awk '{print $3}' | cut -d, -f1)

  host=$(govc object.collect "$cr" host | awk '{print $3}' | cut -d, -f1)

  run govc object.collect "$host"
  assert_success

  run govc object.collect "$host" hardware
  assert_success

  run govc object.collect "$host" hardware.systemInfo
  assert_success

  uuid=$(govc object.collect "$host" hardware.systemInfo.uuid | awk '{print $3}')
  uuid_s=$(govc object.collect -s "$host" hardware.systemInfo.uuid)
  assert_equal "$uuid" "$uuid_s"

  run govc object.collect "$(govc ls host | head -n1)"
  assert_success

  # test against slice of interface
  perfman=$(govc object.collect -s - content.perfManager)
  result=$(govc object.collect -s "$perfman" description.counterType)
  assert_equal "..." "$result"

  # test against an interface field
  run govc object.collect '/ha-datacenter/network/VM Network' summary
  assert_success
}
