/*
Package availabilityzones provides the ability to get lists and detailed
availability zone information and to extend a server result with
availability zone information.

Example of Get Availability Zone Information

		allPages, err := availabilityzones.List(computeClient).AllPages(context.TODO())
		if err != nil {
			panic(err)
		}

		availabilityZoneInfo, err := availabilityzones.ExtractAvailabilityZones(allPages)
		if err != nil {
			panic(err)
		}

		for _, zoneInfo := range availabilityZoneInfo {
	  		fmt.Printf("%+v\n", zoneInfo)
		}

Example of Get Detailed Availability Zone Information

		allPages, err := availabilityzones.ListDetail(computeClient).AllPages(context.TODO())
		if err != nil {
			panic(err)
		}

		availabilityZoneInfo, err := availabilityzones.ExtractAvailabilityZones(allPages)
		if err != nil {
			panic(err)
		}

		for _, zoneInfo := range availabilityZoneInfo {
	  		fmt.Printf("%+v\n", zoneInfo)
		}
*/
package availabilityzones
