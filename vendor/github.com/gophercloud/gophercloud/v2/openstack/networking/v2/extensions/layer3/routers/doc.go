/*
Package routers enables management and retrieval of Routers from the OpenStack
Networking service.

Example to List Routers

	listOpts := routers.ListOpts{}
	allPages, err := routers.List(networkClient, listOpts).AllPages(context.TODO())
	if err != nil {
		panic(err)
	}

	allRouters, err := routers.ExtractRouters(allPages)
	if err != nil {
		panic(err)
	}

	for _, router := range allRoutes {
		fmt.Printf("%+v\n", router)
	}

Example to Create a Router

	iTrue := true
	gwi := routers.GatewayInfo{
		NetworkID: "8ca37218-28ff-41cb-9b10-039601ea7e6b",
	}

	createOpts := routers.CreateOpts{
		Name:         "router_1",
		AdminStateUp: &iTrue,
		GatewayInfo:  &gwi,
	}

	router, err := routers.Create(context.TODO(), networkClient, createOpts).Extract()
	if err != nil {
		panic(err)
	}

Example to Update a Router

	routerID := "4e8e5957-649f-477b-9e5b-f1f75b21c03c"

	routes := []routers.Route{{
		DestinationCIDR: "40.0.1.0/24",
		NextHop:         "10.1.0.10",
	}}

	updateOpts := routers.UpdateOpts{
		Name:   "new_name",
		Routes: &routes,
	}

	router, err := routers.Update(context.TODO(), networkClient, routerID, updateOpts).Extract()
	if err != nil {
		panic(err)
	}

Example to Update just the Router name, keeping everything else as-is

	routerID := "4e8e5957-649f-477b-9e5b-f1f75b21c03c"

	updateOpts := routers.UpdateOpts{
		Name:   "new_name",
	}

	router, err := routers.Update(context.TODO(), networkClient, routerID, updateOpts).Extract()
	if err != nil {
		panic(err)
	}

Example to Remove all Routes from a Router

	routerID := "4e8e5957-649f-477b-9e5b-f1f75b21c03c"

	routes := []routers.Route{}

	updateOpts := routers.UpdateOpts{
		Routes: &routes,
	}

	router, err := routers.Update(context.TODO(), networkClient, routerID, updateOpts).Extract()
	if err != nil {
		panic(err)
	}

Example to Delete a Router

	routerID := "4e8e5957-649f-477b-9e5b-f1f75b21c03c"
	err := routers.Delete(context.TODO(), networkClient, routerID).ExtractErr()
	if err != nil {
		panic(err)
	}

Example to Add an Interface to a Router

	routerID := "4e8e5957-649f-477b-9e5b-f1f75b21c03c"

	intOpts := routers.AddInterfaceOpts{
		SubnetID: "a2f1f29d-571b-4533-907f-5803ab96ead1",
	}

	interface, err := routers.AddInterface(context.TODO(), networkClient, routerID, intOpts).Extract()
	if err != nil {
		panic(err)
	}

Example to Remove an Interface from a Router

	routerID := "4e8e5957-649f-477b-9e5b-f1f75b21c03c"

	intOpts := routers.RemoveInterfaceOpts{
		SubnetID: "a2f1f29d-571b-4533-907f-5803ab96ead1",
	}

	interface, err := routers.RemoveInterface(context.TODO(), networkClient, routerID, intOpts).Extract()
	if err != nil {
		panic(err)
	}

Example to List an L3 agents for a Router

	routerID := "4e8e5957-649f-477b-9e5b-f1f75b21c03c"

	allPages, err := routers.ListL3Agents(networkClient, routerID).AllPages(context.TODO())
	if err != nil {
		panic(err)
	}

	allL3Agents, err := routers.ExtractL3Agents(allPages)
	if err != nil {
		panic(err)
	}

	for _, agent := range allL3Agents {
		fmt.Printf("%+v\n", agent)
	}

Example to Add External Gateways to a Router: This requires the external-gateway-multihoming extension.

	routerID := "4e8e5957-649f-477b-9e5b-f1f75b21c03c"

	addOpts := routers.AddExternalGatewaysOpts{
		ExternalGateways: []routers.GatewayInfo{
			{
				NetworkID: "8ca37218-28ff-41cb-9b10-039601ea7e6b",
				ExternalFixedIPs: []routers.ExternalFixedIP{
					{SubnetID: "ab561bc4-1a8e-48f2-9fbd-376fcb1a1def"},
				},
			},
		},
	}

	router, err := routers.AddExternalGateways(context.TODO(), networkClient, routerID, addOpts).Extract()
	if err != nil {
		panic(err)
	}

Example to Update External Gateways of a Router: This requires the external-gateway-multihoming extension.

	routerID := "4e8e5957-649f-477b-9e5b-f1f75b21c03c"

	enableSNAT := true
	updateOpts := routers.UpdateExternalGatewaysOpts{
		ExternalGateways: []routers.GatewayInfo{
			{
				NetworkID:  "8ca37218-28ff-41cb-9b10-039601ea7e6b",
				EnableSNAT: &enableSNAT,
				ExternalFixedIPs: []routers.ExternalFixedIP{
					{IPAddress: "192.0.2.17", SubnetID: "ab561bc4-1a8e-48f2-9fbd-376fcb1a1def"},
				},
			},
		},
	}

	router, err := routers.UpdateExternalGateways(context.TODO(), networkClient, routerID, updateOpts).Extract()
	if err != nil {
		panic(err)
	}

Example to Remove External Gateways from a Router: This requires the external-gateway-multihoming extension.

	routerID := "4e8e5957-649f-477b-9e5b-f1f75b21c03c"

	removeOpts := routers.RemoveExternalGatewaysOpts{
		ExternalGateways: []routers.GatewayInfo{
			{
				NetworkID: "8ca37218-28ff-41cb-9b10-039601ea7e6b",
				ExternalFixedIPs: []routers.ExternalFixedIP{
					{IPAddress: "192.0.2.17", SubnetID: "ab561bc4-1a8e-48f2-9fbd-376fcb1a1def"},
				},
			},
		},
	}

	router, err := routers.RemoveExternalGateways(context.TODO(), networkClient, routerID, removeOpts).Extract()
	if err != nil {
		panic(err)
	}
*/
package routers
