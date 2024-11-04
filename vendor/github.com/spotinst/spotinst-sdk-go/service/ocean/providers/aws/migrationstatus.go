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

type MigrationStatus struct {
	ID                   *string            `json:"Id,omitempty"`
	OceanId              *string            `json:"oceanId,omitempty"`
	Status               *string            `json:"status,omitempty"`
	NewInstances         []*InstanceDetails `json:"newInstances,omitempty"`
	OldInstances         []*InstanceDetails `json:"oldInstances,omitempty"`
	UnscheduledPodIds    *string            `json:"unscheduledPodIds,omitempty"`
	NewUnscheduledPodIds *string            `json:"newUnscheduledPodIds,omitempty"`
	MigrationConfig      *MigrationConfig   `json:"migrationConfig,omitempty"`
	CreatedAt            *string            `json:"createdAt,omitempty"`
	ErroredAt            *string            `json:"erroredAt,omitempty"`
	StoppedAt            *string            `json:"stoppedAt,omitempty"`
	CompletedAt          *string            `json:"completedAt,omitempty"`

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
type MigrationConfig struct {
	ShouldTerminateDrainedNodes *bool `json:"shouldTerminateDrainedNodes,omitempty"`
	ShouldEvictStandAlonePods   *bool `json:"shouldEvictStandAlonePods,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type InstanceDetails struct {
	InstanceId  *string       `json:"instanceId,omitempty"`
	K8sNodeName *string       `json:"k8sNodeName,omitempty"`
	AsgName     *string       `json:"asgName,omitempty"`
	State       *string       `json:"state,omitempty"`
	RunningPods *int64        `json:"runningPods,omitempty"`
	PodDetails  []*PodDetails `json:"podDetails,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type PodDetails struct {
	ID   *string `json:"id,omitempty"`
	Name *string `json:"name,omitempty"`
	Kind *string `json:"kind,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type ReadMigrationStatusInput struct {
	ClusterID   *string `json:"clusterId,omitempty"`
	MigrationID *string `json:"migrationId,omitempty"`
}

type ReadMigrationStatusOutput struct {
	MigrationStatus []*MigrationStatus `json:"migrationStatus,omitempty"`
}

func (s *ServiceOp) MigrationStatus(ctx context.Context, input *ReadMigrationStatusInput) (*ReadMigrationStatusOutput, error) {
	path, err := uritemplates.Expand("/ocean/aws/k8s/cluster/{oceanClusterId}/migration/{migrationId}", uritemplates.Values{
		"oceanClusterId": spotinst.StringValue(input.ClusterID),
		"migrationId":    spotinst.StringValue(input.MigrationID),
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
	gs, err := migrationStatusFromHttpResponse(resp)
	if err != nil {
		return nil, err
	}

	return &ReadMigrationStatusOutput{MigrationStatus: gs}, nil
}

func migrationStatusFromHttpResponse(resp *http.Response) ([]*MigrationStatus, error) {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return migrationStatusFromJSON(body)
}

func migrationStatusFromJSON(in []byte) ([]*MigrationStatus, error) {
	var rw client.Response
	if err := json.Unmarshal(in, &rw); err != nil {
		return nil, err
	}
	out := make([]*MigrationStatus, len(rw.Response.Items))
	if len(out) == 0 {
		return out, nil
	}
	for i, rb := range rw.Response.Items {
		b, err := statusFromJSON(rb)
		if err != nil {
			return nil, err
		}
		out[i] = b
	}
	return out, nil
}

func statusFromJSON(in []byte) (*MigrationStatus, error) {
	b := new(MigrationStatus)
	if err := json.Unmarshal(in, b); err != nil {
		return nil, err
	}
	return b, nil
}
