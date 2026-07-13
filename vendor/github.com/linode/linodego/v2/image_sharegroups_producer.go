package linodego

import (
	"context"
	"encoding/json"
	"time"

	"github.com/linode/linodego/v2/internal/parseabletime"
)

// ProducerImageShareGroup represents an ImageShareGroup owned by the producer.
type ProducerImageShareGroup struct {
	ID           int        `json:"id"`
	UUID         string     `json:"uuid"`
	Label        string     `json:"label"`
	Description  string     `json:"description"`
	IsSuspended  bool       `json:"is_suspended"`
	ImagesCount  int        `json:"images_count"`
	MembersCount int        `json:"members_count"`
	Created      *time.Time `json:"-"`
	Updated      *time.Time `json:"-"`
	Expiry       *time.Time `json:"-"`
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (isg *ProducerImageShareGroup) UnmarshalJSON(b []byte) error {
	type Mask ProducerImageShareGroup

	p := struct {
		*Mask

		Created *parseabletime.ParseableTime `json:"created"`
		Updated *parseabletime.ParseableTime `json:"updated"`
		Expiry  *parseabletime.ParseableTime `json:"expiry"`
	}{
		Mask: (*Mask)(isg),
	}

	if err := json.Unmarshal(b, &p); err != nil {
		return err
	}

	isg.Created = (*time.Time)(p.Created)
	isg.Updated = (*time.Time)(p.Updated)
	isg.Expiry = (*time.Time)(p.Expiry)

	return nil
}

// ImageShareGroupCreateOptions fields are those accepted by CreateImageShareGroup.
type ImageShareGroupCreateOptions struct {
	Label       string                 `json:"label"`
	Description *string                `json:"description,omitzero"`
	Images      []ImageShareGroupImage `json:"images,omitzero"`
}

// ImageShareGroupUpdateOptions fields are those accepted by UpdateImageShareGroup.
type ImageShareGroupUpdateOptions struct {
	Label       *string `json:"label,omitzero"`
	Description *string `json:"description,omitzero"`
}

// ImageShareGroupAddImagesOptions fields are those accepted by ImageShareGroupAddImages.
type ImageShareGroupAddImagesOptions struct {
	Images []ImageShareGroupImage `json:"images"`
}

// ImageShareGroupUpdateImageOptions fields are those accepted by ImageShareGroupUpdateImage.
type ImageShareGroupUpdateImageOptions struct {
	Label       *string `json:"label,omitzero"`
	Description *string `json:"description,omitzero"`
}

// ImageShareGroupImage represents an Image to be included in a ProducerImageShareGroup.
type ImageShareGroupImage struct {
	ID          string  `json:"id"`
	Label       *string `json:"label,omitzero"`
	Description *string `json:"description,omitzero"`
}

// ImageShareGroupMember represents a Member of an ImageShareGroup owned by the producer.
type ImageShareGroupMember struct {
	TokenUUID string     `json:"token_uuid"`
	Status    string     `json:"status"`
	Label     string     `json:"label"`
	Created   *time.Time `json:"-"`
	Updated   *time.Time `json:"-"`
	Expiry    *time.Time `json:"-"`
}

// ImageShareGroupUpdateMemberOptions fields are those accepted by ImageShareGroupUpdateMember.
type ImageShareGroupUpdateMemberOptions struct {
	Label string `json:"label"`
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (m *ImageShareGroupMember) UnmarshalJSON(b []byte) error {
	type Mask ImageShareGroupMember

	p := struct {
		*Mask

		Created *parseabletime.ParseableTime `json:"created"`
		Updated *parseabletime.ParseableTime `json:"updated"`
		Expiry  *parseabletime.ParseableTime `json:"expiry"`
	}{
		Mask: (*Mask)(m),
	}

	if err := json.Unmarshal(b, &p); err != nil {
		return err
	}

	m.Created = (*time.Time)(p.Created)
	m.Updated = (*time.Time)(p.Updated)
	m.Expiry = (*time.Time)(p.Expiry)

	return nil
}

// ImageShareGroupAddMemberOptions fields are those accepted by ImageShareGroupAddMember.
// The token must be provided to the producer by the consumer via an outside medium.
type ImageShareGroupAddMemberOptions struct {
	Token string `json:"token"`
	Label string `json:"label"`
}

// ListImageShareGroups lists all ImageShareGroups owned by the producer.
// NOTE: May not currently be available to all users and can only be used with v4beta.
func (c *Client) ListImageShareGroups(
	ctx context.Context,
	opts *ListOptions,
) ([]ProducerImageShareGroup, error) {
	return getPaginatedResults[ProducerImageShareGroup](
		ctx,
		c,
		"images/sharegroups",
		opts,
	)
}

// ListImageShareGroupsContainingPrivateImage lists all current ImageShareGroups owned by the producer where
// the given private image is present.
// NOTE: May not currently be available to all users and can only be used with v4beta.
func (c *Client) ListImageShareGroupsContainingPrivateImage(
	ctx context.Context,
	privateImageID string,
	opts *ListOptions,
) ([]ProducerImageShareGroup, error) {
	return getPaginatedResults[ProducerImageShareGroup](
		ctx,
		c,
		formatAPIPath("images/%s/sharegroups", privateImageID),
		opts,
	)
}

// GetImageShareGroup gets the specified ImageShareGroup owned by the producer.
// NOTE: May not currently be available to all users and can only be used with v4beta.
func (c *Client) GetImageShareGroup(
	ctx context.Context,
	imageShareGroupID int,
) (*ProducerImageShareGroup, error) {
	return doGETRequest[ProducerImageShareGroup](
		ctx,
		c,
		formatAPIPath("images/sharegroups/%d", imageShareGroupID),
	)
}

// CreateImageShareGroup allows the producer to create a new ImageShareGroup.
// NOTE: May not currently be available to all users and can only be used with v4beta.
func (c *Client) CreateImageShareGroup(
	ctx context.Context,
	opts ImageShareGroupCreateOptions,
) (*ProducerImageShareGroup, error) {
	return doPOSTRequest[ProducerImageShareGroup](
		ctx,
		c,
		"images/sharegroups",
		opts,
	)
}

// UpdateImageShareGroup allows the producer to update an existing ImageShareGroup's description and label.
// NOTE: May not currently be available to all users and can only be used with v4beta.
func (c *Client) UpdateImageShareGroup(
	ctx context.Context,
	imageShareGroupID int,
	opts ImageShareGroupUpdateOptions,
) (*ProducerImageShareGroup, error) {
	return doPUTRequest[ProducerImageShareGroup](
		ctx,
		c,
		formatAPIPath("images/sharegroups/%d", imageShareGroupID),
		opts,
	)
}

// DeleteImageShareGroup deletes the specified ImageShareGroup owned by the producer.
// NOTE: May not currently be available to all users and can only be used with v4beta.
func (c *Client) DeleteImageShareGroup(ctx context.Context, imageShareGroupID int) error {
	return doDELETERequest(
		ctx,
		c,
		formatAPIPath("images/sharegroups/%d", imageShareGroupID),
	)
}

// ImageShareGroupListImageShareEntries lists the shared image entries of a specified ImageShareGroup owned by the producer.
// NOTE: May not currently be available to all users and can only be used with v4beta.
func (c *Client) ImageShareGroupListImageShareEntries(
	ctx context.Context,
	imageShareGroupID int,
	opts *ListOptions,
) ([]ImageShareEntry, error) {
	return getPaginatedResults[ImageShareEntry](
		ctx,
		c,
		formatAPIPath("images/sharegroups/%d/images", imageShareGroupID),
		opts,
	)
}

// ImageShareGroupAddImages allows the producer to add images to a specific ImageShareGroup.
// NOTE: May not currently be available to all users and can only be used with v4beta.
func (c *Client) ImageShareGroupAddImages(
	ctx context.Context,
	imageShareGroupID int,
	opts ImageShareGroupAddImagesOptions,
) ([]ImageShareEntry, error) {
	return postPaginatedResults[ImageShareEntry, ImageShareGroupAddImagesOptions](
		ctx,
		c,
		formatAPIPath("images/sharegroups/%d/images", imageShareGroupID),
		nil,
		opts,
	)
}

// ImageShareGroupUpdateImageShareEntry allows the producer to update the description and label of a specified ImageShareEntry within the specified ImageShareGroup.
// NOTE: May not currently be available to all users and can only be used with v4beta.
func (c *Client) ImageShareGroupUpdateImageShareEntry(
	ctx context.Context,
	imageShareGroupID int,
	imageID string,
	opts ImageShareGroupUpdateImageOptions,
) (*ImageShareEntry, error) {
	return doPUTRequest[ImageShareEntry](
		ctx,
		c,
		formatAPIPath("images/sharegroups/%d/images/%s", imageShareGroupID, imageID),
		opts,
	)
}

// ImageShareGroupRemoveImage allows the producer to remove access to an image within an ImageShareGroup owned by the producer.
// NOTE: May not currently be available to all users and can only be used with v4beta.
func (c *Client) ImageShareGroupRemoveImage(
	ctx context.Context,
	imageShareGroupID int,
	imageID string,
) error {
	return doDELETERequest(
		ctx,
		c,
		formatAPIPath("images/sharegroups/%d/images/%s", imageShareGroupID, imageID),
	)
}

// ImageShareGroupListMembers lists the ImageShareGroupMembers of the provided ImageShareGroup owned by the producer.
// NOTE: May not currently be available to all users and can only be used with v4beta.
func (c *Client) ImageShareGroupListMembers(
	ctx context.Context,
	imageShareGroupID int,
	opts *ListOptions,
) ([]ImageShareGroupMember, error) {
	return getPaginatedResults[ImageShareGroupMember](
		ctx,
		c,
		formatAPIPath("images/sharegroups/%d/members", imageShareGroupID),
		opts,
	)
}

// ImageShareGroupGetMember gets the details of the specified ImageShareGroupMember in the specified
// ImageShareGroup owned by the producer.
// NOTE: May not currently be available to all users and can only be used with v4beta.
func (c *Client) ImageShareGroupGetMember(
	ctx context.Context,
	imageShareGroupID int,
	tokenUUID string,
) (*ImageShareGroupMember, error) {
	return doGETRequest[ImageShareGroupMember](
		ctx,
		c,
		formatAPIPath("images/sharegroups/%d/members/%s", imageShareGroupID, tokenUUID),
	)
}

// ImageShareGroupAddMember allows the producer to add members to a specific ImageShareGroup.
// NOTE: May not currently be available to all users and can only be used with v4beta.
func (c *Client) ImageShareGroupAddMember(
	ctx context.Context,
	imageShareGroupID int,
	opts ImageShareGroupAddMemberOptions,
) (*ImageShareGroupMember, error) {
	return doPOSTRequest[ImageShareGroupMember](
		ctx,
		c,
		formatAPIPath("images/sharegroups/%d/members", imageShareGroupID),
		opts,
	)
}

// ImageShareGroupUpdateMember allows the producer to update the label associated with the specified
// ImageShareGroupMember in the specified ImageShareGroup owned by the producer.
// NOTE: May not currently be available to all users and can only be used with v4beta.
func (c *Client) ImageShareGroupUpdateMember(
	ctx context.Context,
	imageShareGroupID int,
	tokenUUID string,
	opts ImageShareGroupUpdateMemberOptions,
) (*ImageShareGroupMember, error) {
	return doPUTRequest[ImageShareGroupMember](
		ctx,
		c,
		formatAPIPath("images/sharegroups/%d/members/%s", imageShareGroupID, tokenUUID),
		opts,
	)
}

// ImageShareGroupRemoveMember allows the producer to remove an individual ImageShareGroupMember
// that’s been accepted into the ImageShareGroup owned by the producer.
// NOTE: May not currently be available to all users and can only be used with v4beta.
func (c *Client) ImageShareGroupRemoveMember(
	ctx context.Context,
	imageShareGroupID int,
	tokenUUID string,
) error {
	return doDELETERequest(
		ctx,
		c,
		formatAPIPath("images/sharegroups/%d/members/%s", imageShareGroupID, tokenUUID),
	)
}
