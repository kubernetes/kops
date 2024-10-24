/*
Package attributestags manages Tags on Resources created by the OpenStack Neutron Service.

This enables tagging via a standard interface for resources types which support it.

See https://developer.openstack.org/api-ref/network/v2/#standard-attributes-tag-extension for more information on the underlying API.

Example to ReplaceAll Resource Tags

	network, err := networks.Create(context.TODO(), client, createOpts).Extract()

	tagReplaceAllOpts := attributestags.ReplaceAllOpts{
	    Tags:         []string{"abc", "123"},
	}
	attributestags.ReplaceAll(context.TODO(), client, "networks", network.ID, tagReplaceAllOpts)

Example to List all Resource Tags

	tags, err = attributestags.List(context.TODO(), client, "networks", network.ID).Extract()

Example to Delete all Resource Tags

	err = attributestags.DeleteAll(context.TODO(), client, "networks", network.ID).ExtractErr()

Example to Add a tag to a Resource

	err = attributestags.Add(context.TODO(), client, "networks", network.ID, "atag").ExtractErr()

Example to Delete a tag from a Resource

	err = attributestags.Delete(context.TODO(), client, "networks", network.ID, "atag").ExtractErr()

Example to confirm if a tag exists on a resource

	exists, _ := attributestags.Confirm(context.TODO(), client, "networks", network.ID, "atag").Extract()
*/
package attributestags
