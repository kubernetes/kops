// Copyright 2023 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package report provides an API to the configfs/tsm/report subsystem for collecting
// attestation reports and associated certificates.
package report

import (
	"errors"
	"fmt"

	"github.com/google/go-configfs-tsm/configfs/configfsi"
	"go.uber.org/multierr"
)

const (
	subsystem           = "report"
	subsystemPath       = configfsi.TsmPrefix + "/" + subsystem
	numberAttributeBase = 10
)

// Privilege represents the requested privilege information at which a report should
// be created.
type Privilege struct {
	Level uint
}

// Request represents an open request for an attestation report.
type Request struct {
	InBlob                 []byte
	Privilege              *Privilege
	GetAuxBlob             bool
	ServiceProvider        string
	ServiceGuid            string
	ServiceManifestVersion string
}

// OpenReport represents a created tsm report subtree with internal expectations for the generation.
type OpenReport struct {
	InBlob                 []byte
	Privilege              *Privilege
	GetAuxBlob             bool
	ServiceProvider        string
	ServiceGuid            string
	ServiceManifestVersion string
	entry                  *configfsi.TsmPath
	expectedGeneration     uint64
	client                 configfsi.Client
}

// Response represents a common case response for getting at attestation report to avoid
// multiple attribute access calls.
type Response struct {
	Provider     string
	OutBlob      []byte
	AuxBlob      []byte
	ManifestBlob []byte
}

// GenerationErr is returned when an attribute's value is invalid due to mismatched expectations
// on the number of writes to a report entry.
type GenerationErr struct {
	Got       uint64
	Want      uint64
	Attribute string
}

// Error returns the human-readable explanation for the error.
func (e *GenerationErr) Error() string {
	return fmt.Sprintf("report generation was %d when expecting %d while reading property %q",
		e.Got, e.Want, e.Attribute)
}

// GetGenerationErr returns the GenerationErr contained in an error with 0 or 1 wraps.
func GetGenerationErr(err error) *GenerationErr {
	var result *GenerationErr
	if err != nil && (errors.As(err, &result) || errors.As(errors.Unwrap(err), &result)) {
		return result
	}
	return nil
}

func (r *OpenReport) attribute(subtree string) string {
	a := *r.entry
	a.Attribute = subtree
	return a.String()
}

func readUint64File(client configfsi.Client, p string) (uint64, error) {
	data, err := client.ReadFile(p)
	if err != nil {
		return 0, fmt.Errorf("could not read %q: %v", p, err)
	}
	return configfsi.Kstrtouint(data, numberAttributeBase, 64)
}

// CreateOpenReport returns a newly-created entry in the configfs-tsm report subtree with an initial
// expected generation value.
func CreateOpenReport(client configfsi.Client) (*OpenReport, error) {
	entry, err := client.MkdirTemp(subsystemPath, "entry")
	if err != nil {
		return nil, fmt.Errorf("could not create report entry in configfs: %v", err)
	}
	return UnsafeWrap(client, entry)
}

// UnsafeWrap returns a new OpenReport for a given report entry.
func UnsafeWrap(client configfsi.Client, entryPath string) (r *OpenReport, err error) {
	p, _ := configfsi.ParseTsmPath(entryPath)
	r = &OpenReport{
		client: client,
		entry:  &configfsi.TsmPath{Subsystem: subsystem, Entry: p.Entry},
	}
	r.expectedGeneration, err = readUint64File(client, r.attribute("generation"))
	if err != nil {
		// The report was created but couldn't be properly initialized.
		return nil, multierr.Combine(r.Destroy(), err)
	}
	return r, nil
}

// Create returns a newly-created entry in the configfs-tsm report subtree with common inputs
// for the Get() method initialized from the request.
func Create(client configfsi.Client, req *Request) (*OpenReport, error) {
	r, err := CreateOpenReport(client)
	if err != nil {
		return nil, err
	}
	r.InBlob = req.InBlob // InBlob is not a copy!
	r.Privilege = req.Privilege
	r.GetAuxBlob = req.GetAuxBlob
	r.ServiceProvider = req.ServiceProvider
	r.ServiceGuid = req.ServiceGuid
	r.ServiceManifestVersion = req.ServiceManifestVersion
	return r, nil
}

// Destroy returns an error if the configfs report subtree cannot be removed. Will not error for
// partially initialized or already-destroyed reports.
func (r *OpenReport) Destroy() error {
	if r.entry != nil {
		if err := r.client.RemoveAll(r.entry.String()); err != nil {
			return err
		}
		r.entry = nil
	}
	return nil
}

// PrivilegeLevelFloor returns the privlevel_floor attribute interpreted as the uint type it is.
func (r *OpenReport) PrivilegeLevelFloor() (uint, error) {
	data, err := r.ReadOption("privlevel_floor")
	if err != nil {
		return 0, err
	}
	i, err := configfsi.Kstrtouint(data, numberAttributeBase, 32)
	if err != nil {
		return 0, fmt.Errorf("could not parse privlevel_floor data %v as int: %v", data, err)
	}
	return uint(i), nil
}

// WriteOption sets a configfs report option to the provided data and internally tracks
// the generation that should be expected on the next ReadOption.
func (r *OpenReport) WriteOption(subtree string, data []byte) error {
	if err := r.client.WriteFile(r.attribute(subtree), data); err != nil {
		return fmt.Errorf("could not write report %s: %v", subtree, err)
	}
	r.expectedGeneration += 1
	return nil
}

// ReadOption is a safe accessor to a readable attribute of a report. Returns an error if there is
// any detected tampering to the ongoing request.
func (r *OpenReport) ReadOption(subtree string) ([]byte, error) {
	data, err := r.client.ReadFile(r.attribute(subtree))
	if err != nil {
		return nil, fmt.Errorf("could not read report property %q: %v", subtree, err)
	}
	gotGeneration, err := readUint64File(r.client, r.attribute("generation"))
	if err != nil {
		return nil, err
	}
	if gotGeneration != r.expectedGeneration {
		return nil, &GenerationErr{Got: gotGeneration, Want: r.expectedGeneration, Attribute: subtree}
	}
	return data, nil
}

// Get returns the requested report data after initializing the context to the expected
// parameters. Returns an error if the kernel reports an error or there is a difference in expected
// generation value.
func (r *OpenReport) Get() (*Response, error) {
	var err error
	if err := r.WriteOption("inblob", r.InBlob); err != nil {
		return nil, err
	}
	if r.Privilege != nil {
		if err := r.WriteOption("privlevel", []byte(fmt.Sprintf("%d", r.Privilege.Level))); err != nil {
			return nil, err
		}
	}
	if r.ServiceProvider != "" {
		if err := r.WriteOption("service_provider", []byte(r.ServiceProvider)); err != nil {
			return nil, err
		}
	}
	if r.ServiceGuid != "" {
		if err := r.WriteOption("service_guid", []byte(r.ServiceGuid)); err != nil {
			return nil, err
		}
	}
	if r.ServiceManifestVersion != "" {
		if err := r.WriteOption("service_manifest_version", []byte(r.ServiceManifestVersion)); err != nil {
			return nil, err
		}
	}
	resp := &Response{}

	if r.GetAuxBlob {
		resp.AuxBlob, err = r.ReadOption("auxblob")
		if err != nil {
			return nil, fmt.Errorf("could not read report auxblob: %w", err)
		}
	}
	resp.OutBlob, err = r.ReadOption("outblob")
	if err != nil {
		return nil, fmt.Errorf("could not read report outblob: %w", err)
	}
	providerData, err := r.ReadOption("provider")
	if err != nil {
		return nil, err
	}
	resp.Provider = string(providerData)
	if r.ServiceProvider != "" {
		manifest, err := r.ReadOption("manifestblob")
		if err != nil {
			return nil, fmt.Errorf("could not read report manifestblob: %w", err)
		}
		resp.ManifestBlob = manifest
	}
	return resp, nil
}

// Get returns a one-shot configfs-tsm report given a report request.
func Get(client configfsi.Client, req *Request) (*Response, error) {
	var err error
	r, err := Create(client, req)
	if err != nil {
		return nil, err
	}
	response, err := r.Get()
	return response, multierr.Combine(r.Destroy(), err)
}
