// This file was automatically generated. DO NOT EDIT.
// If you have any remark or suggestion do not hesitate to open an issue.

// Package marketplace provides methods and message types of the marketplace v1 API.
package marketplace

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/scaleway/scaleway-sdk-go/internal/errors"
	"github.com/scaleway/scaleway-sdk-go/internal/marshaler"
	"github.com/scaleway/scaleway-sdk-go/internal/parameter"
	"github.com/scaleway/scaleway-sdk-go/namegenerator"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

// always import dependencies
var (
	_ fmt.Stringer
	_ json.Unmarshaler
	_ url.URL
	_ net.IP
	_ http.Header
	_ bytes.Reader
	_ time.Time
	_ = strings.Join

	_ scw.ScalewayRequest
	_ marshaler.Duration
	_ scw.File
	_ = parameter.AddToQuery
	_ = namegenerator.GetRandomName
)

// API: marketplace API
type API struct {
	client *scw.Client
}

// NewAPI returns a API object from a Scaleway client.
func NewAPI(client *scw.Client) *API {
	return &API{
		client: client,
	}
}

type GetImageResponse struct {
	Image *Image `json:"image"`
}

type GetVersionResponse struct {
	Version *Version `json:"version"`
}

// Image: image
type Image struct {
	// ID: UUID of this image
	ID string `json:"id"`
	// Name: name of the image
	Name string `json:"name"`
	// Description: text description of this image
	Description string `json:"description"`
	// Logo: URL of this image's logo
	Logo string `json:"logo"`
	// Categories: list of categories this image belongs to
	Categories []string `json:"categories"`
	// CreationDate: creation date of this image
	CreationDate *time.Time `json:"creation_date"`
	// ModificationDate: date of the last modification of this image
	ModificationDate *time.Time `json:"modification_date"`
	// ValidUntil: expiration date of this image
	ValidUntil *time.Time `json:"valid_until"`
	// Label: label of this image
	//
	// Typically an identifier for a distribution (ex. "ubuntu_focal").
	Label string `json:"label"`
	// Versions: list of versions of this image
	Versions []*Version `json:"versions"`
	// Organization: organization this image belongs to
	Organization *Organization `json:"organization"`

	CurrentPublicVersion string `json:"current_public_version"`
}

type ListImagesResponse struct {
	Images []*Image `json:"images"`

	TotalCount uint32 `json:"total_count"`
}

type ListVersionsResponse struct {
	Versions []*Version `json:"versions"`

	TotalCount uint32 `json:"total_count"`
}

// LocalImage: local image
type LocalImage struct {
	// ID: UUID of this local image
	//
	// Version you will typically use to define an image in an API call.
	ID string `json:"id"`
	// CompatibleCommercialTypes: list of all commercial types that are compatible with this local image
	CompatibleCommercialTypes []string `json:"compatible_commercial_types"`
	// Arch: supported architecture for this local image
	Arch string `json:"arch"`
	// Zone: availability Zone where this local image is available
	Zone scw.Zone `json:"zone"`
}

type Organization struct {
	ID string `json:"id"`

	Name string `json:"name"`
}

// Version: version
type Version struct {
	// ID: UUID of this version
	ID string `json:"id"`
	// Name: name of this version
	Name string `json:"name"`
	// CreationDate: creation date of this image version
	CreationDate *time.Time `json:"creation_date"`
	// ModificationDate: date of the last modification of this version
	ModificationDate *time.Time `json:"modification_date"`
	// LocalImages: list of local images available in this version
	LocalImages []*LocalImage `json:"local_images"`
}

// Service API

type ListImagesRequest struct {
	// PerPage: a positive integer lower or equal to 100 to select the number of items to display
	PerPage *uint32 `json:"-"`
	// Page: a positive integer to choose the page to display
	Page *int32 `json:"-"`
}

// ListImages: list marketplace images
func (s *API) ListImages(req *ListImagesRequest, opts ...scw.RequestOption) (*ListImagesResponse, error) {
	var err error

	defaultPerPage, exist := s.client.GetDefaultPageSize()
	if (req.PerPage == nil || *req.PerPage == 0) && exist {
		req.PerPage = &defaultPerPage
	}

	query := url.Values{}
	parameter.AddToQuery(query, "per_page", req.PerPage)
	parameter.AddToQuery(query, "page", req.Page)

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/marketplace/v1/images",
		Query:   query,
		Headers: http.Header{},
	}

	var resp ListImagesResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type GetImageRequest struct {
	// ImageID: display the image name
	ImageID string `json:"-"`
}

// GetImage: get a specific marketplace image
func (s *API) GetImage(req *GetImageRequest, opts ...scw.RequestOption) (*GetImageResponse, error) {
	var err error

	if fmt.Sprint(req.ImageID) == "" {
		return nil, errors.New("field ImageID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/marketplace/v1/images/" + fmt.Sprint(req.ImageID) + "",
		Headers: http.Header{},
	}

	var resp GetImageResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ListVersionsRequest struct {
	ImageID string `json:"-"`
}

func (s *API) ListVersions(req *ListVersionsRequest, opts ...scw.RequestOption) (*ListVersionsResponse, error) {
	var err error

	if fmt.Sprint(req.ImageID) == "" {
		return nil, errors.New("field ImageID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/marketplace/v1/images/" + fmt.Sprint(req.ImageID) + "/versions",
		Headers: http.Header{},
	}

	var resp ListVersionsResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type GetVersionRequest struct {
	ImageID string `json:"-"`

	VersionID string `json:"-"`
}

func (s *API) GetVersion(req *GetVersionRequest, opts ...scw.RequestOption) (*GetVersionResponse, error) {
	var err error

	if fmt.Sprint(req.ImageID) == "" {
		return nil, errors.New("field ImageID cannot be empty in request")
	}

	if fmt.Sprint(req.VersionID) == "" {
		return nil, errors.New("field VersionID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/marketplace/v1/images/" + fmt.Sprint(req.ImageID) + "/versions/" + fmt.Sprint(req.VersionID) + "",
		Headers: http.Header{},
	}

	var resp GetVersionResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListImagesResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListImagesResponse) UnsafeAppend(res interface{}) (uint32, error) {
	results, ok := res.(*ListImagesResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.Images = append(r.Images, results.Images...)
	r.TotalCount += uint32(len(results.Images))
	return uint32(len(results.Images)), nil
}
