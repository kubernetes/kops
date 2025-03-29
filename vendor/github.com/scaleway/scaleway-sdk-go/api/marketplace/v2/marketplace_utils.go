package marketplace

import (
	"strings"

	"github.com/scaleway/scaleway-sdk-go/errors"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

// FindByLabel returns the first image with the given label in the image list
// Cannot find an image if it is not in the ListImagesResponse struct
// Use scw.WithAllPages when listing image to get all images
func (r *ListImagesResponse) FindByLabel(label string) *Image {
	for _, image := range r.Images {
		if image.Label == label {
			return image
		}
	}
	return nil
}

type GetImageByLabelRequest struct {
	Label string
}

// GetImageByLabel returns the image with the given label
func (s *API) GetImageByLabel(req *GetImageByLabelRequest, opts ...scw.RequestOption) (*Image, error) {
	listImagesRequest := &ListImagesRequest{}
	opts = append(opts, scw.WithAllPages())

	listImagesResponse, err := s.ListImages(listImagesRequest, opts...)
	if err != nil {
		return nil, err
	}

	image := listImagesResponse.FindByLabel(req.Label)
	if image == nil {
		return nil, errors.New("couldn't find a matching image for the given label (%s)", req.Label)
	}

	return image, nil
}

type GetLocalImageByLabelRequest struct {
	ImageLabel     string
	Zone           scw.Zone
	CommercialType string
	Type           LocalImageType
}

// GetLocalImageByLabel returns the local image for the given image label in the given zone and compatible with given commercial type
func (s *API) GetLocalImageByLabel(req *GetLocalImageByLabelRequest, opts ...scw.RequestOption) (*LocalImage, error) {
	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}
	req.CommercialType = strings.ToUpper(req.CommercialType)

	resp, err := s.ListLocalImages(&ListLocalImagesRequest{
		ImageLabel: scw.StringPtr(req.ImageLabel),
		Zone:       &req.Zone,
		Type:       req.Type,
	}, opts...)
	if err != nil {
		return nil, err
	}
	for _, localImage := range resp.LocalImages {
		if localImage.IsCompatible(req.CommercialType) {
			return localImage, nil
		}
	}

	return nil, errors.New("couldn't find a local image for the given zone (%s) and commercial type (%s)", req.Zone, req.CommercialType)
}

// IsCompatible returns true if a local image is compatible with the given instance type
// commercialType should be an uppercase string ex: DEV1-S
func (li *LocalImage) IsCompatible(commercialType string) bool {
	for _, compatibleCommercialType := range li.CompatibleCommercialTypes {
		if compatibleCommercialType == commercialType {
			return true
		}
	}
	return false
}
