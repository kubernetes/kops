package aws

import (
	"context"
	"encoding/json"
	"github.com/spotinst/spotinst-sdk-go/spotinst"
	"github.com/spotinst/spotinst-sdk-go/spotinst/client"
	"github.com/spotinst/spotinst-sdk-go/spotinst/util/uritemplates"
	"io/ioutil"
	"net/http"
)

type Migration struct {
	ID          *string `json:"Id,omitempty"`
	State       *string `json:"state,omitempty"`
	CreatedAt   *string `json:"createdAt,omitempty"`
	CompletedAt *string `json:"completedAt,omitempty"`

	// forceSendFields is a list of field names (e.g. "Keys") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	forceSendFields []string

	// nullFields is a list of field names (e.g. "Keys") to include in API
	// requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	nullFields []string
}
type ReadMigrationInput struct {
	ClusterID *string `json:"clusterId,omitempty"`
}

type ReadMigrationOutput struct {
	Migration []*Migration `json:"migration,omitempty"`
}

func (s *ServiceOp) ListMigrations(ctx context.Context, input *ReadMigrationInput) (*ReadMigrationOutput, error) {
	path, err := uritemplates.Expand("/ocean/aws/k8s/cluster/{oceanClusterId}/migration", uritemplates.Values{
		"oceanClusterId": spotinst.StringValue(input.ClusterID),
	})
	if err != nil {
		return nil, err
	}
	r := client.NewRequest(http.MethodGet, path)
	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	gs, err := migrationsFromHttpResponse(resp)
	if err != nil {
		return nil, err
	}

	return &ReadMigrationOutput{Migration: gs}, nil
}

func migrationsFromHttpResponse(resp *http.Response) ([]*Migration, error) {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return migrationsFromJSON(body)
}

func migrationsFromJSON(in []byte) ([]*Migration, error) {
	var rw client.Response
	if err := json.Unmarshal(in, &rw); err != nil {
		return nil, err
	}
	out := make([]*Migration, len(rw.Response.Items))
	if len(out) == 0 {
		return out, nil
	}
	for i, rb := range rw.Response.Items {
		b, err := migrationFromJSON(rb)
		if err != nil {
			return nil, err
		}
		out[i] = b
	}
	return out, nil
}

func migrationFromJSON(in []byte) (*Migration, error) {
	b := new(Migration)
	if err := json.Unmarshal(in, b); err != nil {
		return nil, err
	}
	return b, nil
}
