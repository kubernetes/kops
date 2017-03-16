package members

import (
	"fmt"

	"github.com/rackspace/gophercloud"
	"github.com/rackspace/gophercloud/pagination"
)

// Create member for specific image
//
// Preconditions
//    The specified images must exist.
//    You can only add a new member to an image which 'visibility' attribute is private.
//    You must be the owner of the specified image.
// Synchronous Postconditions
//    With correct permissions, you can see the member status of the image as pending through API calls.
//
// More details here: http://developer.openstack.org/api-ref-image-v2.html#createImageMember-v2
func Create(client *gophercloud.ServiceClient, id string, member string) CreateMemberResult {
	var res CreateMemberResult
	body := map[string]interface{}{}
	body["member"] = member

	response, err := client.Post(imageMembersURL(client, id), body, &res.Body,
		&gophercloud.RequestOpts{OkCodes: []int{200, 409, 403}})

	//some problems in http stack or lower
	if err != nil {
		res.Err = err
		return res
	}

	// membership conflict
	if response.StatusCode == 409 {
		res.Err = fmt.Errorf("Given tenant '%s' is already member for image '%s'.", member, id)
		return res
	}

	// visibility conflict
	if response.StatusCode == 403 {
		res.Err = fmt.Errorf("You can only add a new member to an image "+
			"which 'visibility' attribute is private (image '%s')", id)
		return res
	}

	return res
}

// List members returns list of members for specifed image id
// More details: http://developer.openstack.org/api-ref-image-v2.html#listImageMembers-v2
func List(client *gophercloud.ServiceClient, id string) pagination.Pager {
	createPage := func(r pagination.PageResult) pagination.Page {
		return MemberPage{pagination.SinglePageBase(r)}
	}

	return pagination.NewPager(client, listMembersURL(client, id), createPage)
}

// Get image member details.
// More details: http://developer.openstack.org/api-ref-image-v2.html#getImageMember-v2
func Get(client *gophercloud.ServiceClient, imageID string, memberID string) MemberDetailsResult {
	var res MemberDetailsResult
	_, res.Err = client.Get(imageMemberURL(client, imageID, memberID), &res.Body, &gophercloud.RequestOpts{OkCodes: []int{200}})
	return res
}

// Delete membership for given image.
// Callee should be image owner
// More details: http://developer.openstack.org/api-ref-image-v2.html#deleteImageMember-v2
func Delete(client *gophercloud.ServiceClient, imageID string, memberID string) MemberDeleteResult {
	var res MemberDeleteResult
	response, err := client.Delete(imageMemberURL(client, imageID, memberID), &gophercloud.RequestOpts{OkCodes: []int{204, 403}})

	//some problems in http stack or lower
	if err != nil {
		res.Err = err
		return res
	}

	// Callee is not owner of specified image
	if response.StatusCode == 403 {
		res.Err = fmt.Errorf("You must be the owner of the specified image. "+
			"(image '%s')", imageID)
		return res
	}
	return res
}

// UpdateOptsBuilder allows extensions to add additional attributes to the Update request.
type UpdateOptsBuilder interface {
	ToMemberUpdateMap() map[string]interface{}
}

// UpdateOpts implements UpdateOptsBuilder
type UpdateOpts struct {
	Status string
}

// ToMemberUpdateMap formats an UpdateOpts structure into a request body.
func (opts UpdateOpts) ToMemberUpdateMap() map[string]interface{} {
	m := make(map[string]interface{})

	if opts.Status != "" {
		m["status"] = opts.Status
	}

	return m
}

// Update function updates member
// More details: http://developer.openstack.org/api-ref-image-v2.html#updateImageMember-v2
func Update(client *gophercloud.ServiceClient, imageID string, memberID string, opts UpdateOptsBuilder) MemberUpdateResult {
	var res MemberUpdateResult
	body := opts.ToMemberUpdateMap()
	_, res.Err = client.Put(imageMemberURL(client, imageID, memberID), body, &res.Body,
		&gophercloud.RequestOpts{OkCodes: []int{200}})
	return res
}
