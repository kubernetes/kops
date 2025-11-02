package hcloud

//go:generate go run github.com/vburenin/ifacemaker -f action.go -f action_watch.go -f action_waiter.go -s ActionClient -i IActionClient -p hcloud -o zz_action_client_iface.go
//go:generate go run github.com/vburenin/ifacemaker -f action.go -s ResourceActionClient -i IResourceActionClient -p hcloud -o zz_resource_action_client_iface.go
//go:generate go run github.com/vburenin/ifacemaker -f datacenter.go -s DatacenterClient -i IDatacenterClient -p hcloud -o zz_datacenter_client_iface.go
//go:generate go run github.com/vburenin/ifacemaker -f floating_ip.go -s FloatingIPClient -i IFloatingIPClient -p hcloud -o zz_floating_ip_client_iface.go
//go:generate go run github.com/vburenin/ifacemaker -f image.go -s ImageClient -i IImageClient -p hcloud -o zz_image_client_iface.go
//go:generate go run github.com/vburenin/ifacemaker -f iso.go -s ISOClient -i IISOClient -p hcloud -o zz_iso_client_iface.go
//go:generate go run github.com/vburenin/ifacemaker -f location.go -s LocationClient -i ILocationClient -p hcloud -o zz_location_client_iface.go
//go:generate go run github.com/vburenin/ifacemaker -f network.go -s NetworkClient -i INetworkClient -p hcloud -o zz_network_client_iface.go
//go:generate go run github.com/vburenin/ifacemaker -f pricing.go -s PricingClient -i IPricingClient -p hcloud -o zz_pricing_client_iface.go
//go:generate go run github.com/vburenin/ifacemaker -f server.go -s ServerClient -i IServerClient -p hcloud -o zz_server_client_iface.go
//go:generate go run github.com/vburenin/ifacemaker -f server_type.go -s ServerTypeClient -i IServerTypeClient -p hcloud -o zz_server_type_client_iface.go
//go:generate go run github.com/vburenin/ifacemaker -f ssh_key.go -s SSHKeyClient -i ISSHKeyClient -p hcloud -o zz_ssh_key_client_iface.go
//go:generate go run github.com/vburenin/ifacemaker -f volume.go -s VolumeClient -i IVolumeClient -p hcloud -o zz_volume_client_iface.go
//go:generate go run github.com/vburenin/ifacemaker -f load_balancer.go -s LoadBalancerClient -i ILoadBalancerClient -p hcloud -o zz_load_balancer_client_iface.go
//go:generate go run github.com/vburenin/ifacemaker -f load_balancer_type.go -s LoadBalancerTypeClient -i ILoadBalancerTypeClient -p hcloud -o zz_load_balancer_type_client_iface.go
//go:generate go run github.com/vburenin/ifacemaker -f certificate.go -s CertificateClient -i ICertificateClient -p hcloud -o zz_certificate_client_iface.go
//go:generate go run github.com/vburenin/ifacemaker -f firewall.go -s FirewallClient -i IFirewallClient -p hcloud -o zz_firewall_client_iface.go
//go:generate go run github.com/vburenin/ifacemaker -f placement_group.go -s PlacementGroupClient -i IPlacementGroupClient -p hcloud -o zz_placement_group_client_iface.go
//go:generate go run github.com/vburenin/ifacemaker -f rdns.go -s RDNSClient -i IRDNSClient -p hcloud -o zz_rdns_client_iface.go
//go:generate go run github.com/vburenin/ifacemaker -f primary_ip.go -s PrimaryIPClient -i IPrimaryIPClient -p hcloud -o zz_primary_ip_client_iface.go
//go:generate go run github.com/vburenin/ifacemaker -f zone.go -f zone_rrset.go -s ZoneClient -i IZoneClient -p hcloud -o zz_zone_client_iface.go
//go:generate go run github.com/vburenin/ifacemaker -f storage_box_type.go -s StorageBoxTypeClient -i IStorageBoxTypeClient -p hcloud -o zz_storage_box_type_client_iface.go
//go:generate go run github.com/vburenin/ifacemaker -f storage_box.go -f storage_box_snapshot.go -f storage_box_subaccount.go -s StorageBoxClient -i IStorageBoxClient -p hcloud -o zz_storage_box_client_iface.go
