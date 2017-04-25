#!/bin/bash

. "$(dirname "$0")/config.sh"

start_suite "Optional encryption via trusted subnets"

# Determine subnet for hosts given either an IP or name. We need to resolve
# these entries on the remote hosts to make sure we get the private
# IP addresses in the Circle/GCE context
HOST1_IP=$($SSH $HOST1 "getent hosts $HOST1" | grep $HOST1 | cut -d ' ' -f 1)
HOST2_IP=$($SSH $HOST2 "getent hosts $HOST2" | grep $HOST2 | cut -d ' ' -f 1)
HOST1_CIDR=$($SSH $HOST1 "ip addr show" | grep -oP $HOST1_IP/[0-9]+)
HOST2_CIDR=$($SSH $HOST2 "ip addr show" | grep -oP $HOST2_IP/[0-9]+)

# Check asymmetric trust - connections should be encrypted
weave_on $HOST1 launch --password wfvAwt7sj --trusted-subnets $HOST2_CIDR
weave_on $HOST2 launch --password wfvAwt7sj $HOST1
assert_raises "weave_on $HOST1 status connections | grep encrypted"
assert_raises "weave_on $HOST2 status connections | grep encrypted"

weave_on $HOST1 stop
weave_on $HOST2 stop

# Check symmetric trust - overlay in plaintext
weave_on $HOST1 launch --password wfvAwt7sj --trusted-subnets $HOST2_CIDR
weave_on $HOST2 launch --password wfvAwt7sj --trusted-subnets $HOST1_CIDR $HOST1
assert_raises "weave_on $HOST1 status connections | grep unencrypted"
assert_raises "weave_on $HOST2 status connections | grep unencrypted"

end_suite
