package instanceactions

/*
Package instanceactions provides the ability to list or get a server instance-action.

Example to List and Get actions:

	pages, err := instanceactions.List(client, "server-id", nil).AllPages(context.TODO())
	if err != nil {
		panic("fail to get actions pages")
	}

	actions, err := instanceactions.ExtractInstanceActions(pages)
	if err != nil {
		panic("fail to list instance actions")
	}

	for _, action := range actions {
		action, err = instanceactions.Get(context.TODO(), client, "server-id", action.RequestID).Extract()
		if err != nil {
			panic("fail to get instance action")
		}

		fmt.Println(action)
	}
*/
