# Weave Net Benchmarks

## 2017-02-16: Fast Datapath Encryption

Results: https://www.weave.works/weave-net-performance-fast-datapath/

* Two `c3.8xlarge` instances running Ubuntu 16.04 LTS (`ami-d8f4deab`) in
`eu-west-1` region of AWS.
* 10 Gigabit [Enhanced Networking](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/enhanced-networking.html).
* `iperf 3.0.11`.

### fastdp

* Encrypted:

```
host1$ WEAVE_MTU=8916 weave launch --password=foobar
host1$ docker $(weave config) run --rm -ti --name server networkstatic/iperf3 -s

host2$ WEAVE_MTU=8916 weave launch --password=foobar host1
host2$ docker $(weave config) run --rm -ti networkstatic/iperf3 -c server
```

* Non-encrypted:

```
host1$ WEAVE_MTU=8950 weave launch
host1$ docker $(weave config) run --rm -ti --name server networkstatic/iperf3 -s

host2$ WEAVE_MTU=8950 weave launch --password=foobar host1
host2$ docker $(weave config) run --rm -ti networkstatic/iperf3 -c server
```

### sleeve

* Encrypted:

```
host1$ WEAVE_NO_FASTDP=1 weave launch --password=foobar
host1$ docker $(weave config) run --rm -ti --name server networkstatic/iperf3 -s

host2$ WEAVE_NO_FASTDP=1 weave launch --password=foobar host1
host2$ docker $(weave config) run --rm -ti networkstatic/iperf3 -c server
```

* Non-encrypted:

```
host1$ WEAVE_NO_FASTDP=1 weave launch
host1$ docker $(weave config) run --rm -ti --name server networkstatic/iperf3 -s

host2$ WEAVE_NO_FASTDP=1 weave launch --password=foobar host1
host2$ docker $(weave config) run --rm -ti networkstatic/iperf3 -c server
```

### host

* Non-encrypted:

```
host1$ iperf -s
host2$ iperf -c host1
```

* Encrypted:

```
host1$ export KEY1="0x466454f2b1a770f8f872f9afbc35ebeac57e00fc11ac86ed1f82716f010b20f0cf532274"
host1$ export KEY2="0x3531151241a770f8f872f9afbc35ebeac57e00fc11ac86ed1f82716f010b20f0cf532274"
host1# ip xfrm state add src ${IP1} dst ${IP2} proto esp spi 0x4c856ffc replay-window 256 flag esn reqid 0 mode transport aead 'rfc4106(gcm(aes))' ${KEY1} 128
host1# ip xfrm state add src ${IP2} dst ${IP1} proto esp spi 0x9b0830bc replay-window 256 flag esn reqid 0 mode transport aead 'rfc4106(gcm(aes))' ${KEY2} 128
host1# ip xfrm policy add src ${IP1}/32 dst ${IP2}/32 dir out tmpl src ${IP1} dst ${IP2} proto esp spi 0x4c856ffc reqid 0 mode transport
host1$ iperf -s

host2$ export KEY1="0x466454f2b1a770f8f872f9afbc35ebeac57e00fc11ac86ed1f82716f010b20f0cf532274"
host2$ export KEY2="0x3531151241a770f8f872f9afbc35ebeac57e00fc11ac86ed1f82716f010b20f0cf532274"
host2# ip xfrm state add src ${IP2} dst ${IP1} proto esp spi 0x9b0830bc reqid 0 replay-window 256 flag esn mode transport aead 'rfc4106(gcm(aes))' ${KEY2} 128
host2# ip xfrm state add src ${IP1} dst ${IP2} proto esp spi 0x4c856ffc reqid 0 replay-window 256 flag esn mode transport aead 'rfc4106(gcm(aes))' ${KEY1} 128
host2# ip xfrm policy add src ${IP2}/32 dst ${IP1}/32 dir out tmpl src ${IP1} dst ${IP2} proto esp spi 0x9b0830bc reqid 0 mode transport
host2$ iperf -c host1
```
