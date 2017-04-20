---
title: Allocating IPs in a Specific Range
menu_order: 15
---

The default configurations for both Weave Net and Docker use [Private
Networks](https://en.wikipedia.org/wiki/Private_network), whose
addresses are never found on the public Internet, and subsequently reduces the
chance of IP overlap. However, it could be that you or your hosting provider
are using some of these private addresses in the same range, which will
cause a clash.

If after `weave launch`, the following error message
appears:

    Network 10.32.0.0/12 overlaps with existing route 10.0.0.0/8 on host.
    ERROR: Default --ipalloc-range 10.32.0.0/12 overlaps with existing route on host.
    You must pick another range and set it on all hosts.

As the message indicates, the default range that Weave Net would like to use is
`10.32.0.0/12` - a 12-bit prefix, where all addresses start with the bit
pattern 000010100010, or in decimal everything from 10.32.0.0 through
10.47.255.255.

However, your host is using a route for `10.0.0.0/8`,
which overlaps, since the first 8 bits are the same. In this case, if you used the default network
for an address like `10.32.5.6` the kernel would never be sure if this meant the
Weave Net network of `10.32.0.0/12` or the hosting network of
`10.0.0.0/8`.

If you are sure the addresses you want are not in use, then
explicitly setting the range with `--ipalloc-range` in the
command-line arguments to `weave launch` on all hosts forces Weave
Net to use that range, even though it overlaps. Otherwise, you can
pick a different range, preferably another subset of the [Private
Networks](https://en.wikipedia.org/wiki/Private_network).  For example
172.30.0.0/16.


**See Also**

 * [IP Addresses, Routes and Networks](/site/how-it-works/ip-addresses.md)
