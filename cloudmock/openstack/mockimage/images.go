/*
Copyright 2020 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package mockimage

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"

	"github.com/google/uuid"
	"github.com/gophercloud/gophercloud/openstack/imageservice/v2/images"
)

type imageListResponse struct {
	Images []images.Image `json:"images"`
}

type imageGetResponse struct {
	Image images.Image `json:"image"`
}

func (m *MockClient) mockImages() {
	re := regexp.MustCompile(`/images/?`)

	handler := func(w http.ResponseWriter, r *http.Request) {
		m.mutex.Lock()
		defer m.mutex.Unlock()

		w.Header().Add("Content-Type", "application/json")

		imageID := re.ReplaceAllString(r.URL.Path, "")
		switch r.Method {
		case http.MethodGet:
			if imageID == "" {
				m.listImages(w)
			} else {
				m.getImage(w, imageID)
			}
		case http.MethodPost:
			m.createImage(w, r)
		case http.MethodDelete:
			m.deleteImage(w, imageID)
		default:
			w.WriteHeader(http.StatusBadRequest)
		}
	}
	m.Mux.HandleFunc("/images/", handler)
	m.Mux.HandleFunc("/images", handler)
}

func (m *MockClient) listImages(w http.ResponseWriter) {

	w.WriteHeader(http.StatusOK)

	images := make([]images.Image, 0)
	for _, image := range m.images {
		images = append(images, image)
	}

	resp := imageListResponse{
		Images: images,
	}
	respB, err := json.Marshal(resp)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal %+v", resp))
	}
	_, err = w.Write(respB)
	if err != nil {
		panic("failed to write body")
	}
}

func (m *MockClient) getImage(w http.ResponseWriter, imageID string) {
	if image, ok := m.images[imageID]; ok {
		resp := imageGetResponse{
			Image: image,
		}
		respB, err := json.Marshal(resp)
		if err != nil {
			panic(fmt.Sprintf("failed to marshal %+v", resp))
		}
		_, err = w.Write(respB)
		if err != nil {
			panic("failed to write body")
		}
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func (m *MockClient) deleteImage(w http.ResponseWriter, imageID string) {
	if _, ok := m.images[imageID]; ok {
		delete(m.images, imageID)
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func (m *MockClient) createImage(w http.ResponseWriter, r *http.Request) {
	var create images.CreateOpts
	err := json.NewDecoder(r.Body).Decode(&create)
	if err != nil {
		panic("error decoding create image request")
	}

	w.WriteHeader(http.StatusAccepted)

	image := images.Image{
		ID:               uuid.New().String(),
		Name:             create.Name,
		MinDiskGigabytes: create.MinDisk,
	}
	m.images[image.ID] = image
	resp := imageGetResponse{
		Image: image,
	}
	respB, err := json.Marshal(resp)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal %+v", resp))
	}
	_, err = w.Write(respB)
	if err != nil {
		panic("failed to write body")
	}
}
