/*
Copyright 2015 Google Inc. All Rights Reserved.

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

package bigtable

import (
	"fmt"
	"regexp"
	"strings"

	"golang.org/x/net/context"
	"google.golang.org/cloud"
	btispb "google.golang.org/cloud/bigtable/internal/instance_service_proto"
	bttdpb "google.golang.org/cloud/bigtable/internal/table_data_proto"
	bttspb "google.golang.org/cloud/bigtable/internal/table_service_proto"
	"google.golang.org/cloud/internal/transport"
	"google.golang.org/grpc"
)

const adminAddr = "bigtableadmin.googleapis.com:443"

// AdminClient is a client type for performing admin operations within a specific instance.
type AdminClient struct {
	conn    *grpc.ClientConn
	tClient bttspb.BigtableTableAdminClient

	project, instance string
}

// NewAdminClient creates a new AdminClient for a given project and instance.
func NewAdminClient(ctx context.Context, project, instance string, opts ...cloud.ClientOption) (*AdminClient, error) {
	o := []cloud.ClientOption{
		cloud.WithEndpoint(adminAddr),
		cloud.WithScopes(AdminScope),
		cloud.WithUserAgent(clientUserAgent),
	}
	o = append(o, opts...)
	conn, err := transport.DialGRPC(ctx, o...)
	if err != nil {
		return nil, fmt.Errorf("dialing: %v", err)
	}
	return &AdminClient{
		conn:     conn,
		tClient:  bttspb.NewBigtableTableAdminClient(conn),
		project:  project,
		instance: instance,
	}, nil
}

// Close closes the AdminClient.
func (ac *AdminClient) Close() error {
	return ac.conn.Close()
}

func (ac *AdminClient) instancePrefix() string {
	return fmt.Sprintf("projects/%s/instances/%s", ac.project, ac.instance)
}

// Tables returns a list of the tables in the instance.
func (ac *AdminClient) Tables(ctx context.Context) ([]string, error) {
	prefix := ac.instancePrefix()
	req := &bttspb.ListTablesRequest{
		Name: prefix,
	}
	res, err := ac.tClient.ListTables(ctx, req)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(res.Tables))
	for _, tbl := range res.Tables {
		names = append(names, strings.TrimPrefix(tbl.Name, prefix+"/tables/"))
	}
	return names, nil
}

// CreateTable creates a new table in the instance.
// This method may return before the table's creation is complete.
func (ac *AdminClient) CreateTable(ctx context.Context, table string) error {
	prefix := ac.instancePrefix()
	req := &bttspb.CreateTableRequest{
		Name:    prefix,
		TableId: table,
	}
	_, err := ac.tClient.CreateTable(ctx, req)
	if err != nil {
		return err
	}
	return nil
}

// CreateColumnFamily creates a new column family in a table.
func (ac *AdminClient) CreateColumnFamily(ctx context.Context, table, family string) error {
	// TODO(dsymonds): Permit specifying gcexpr and any other family settings.
	prefix := ac.instancePrefix()
	req := &bttspb.ModifyColumnFamiliesRequest{
		Name: prefix + "/tables/" + table,
		Modifications: []*bttspb.ModifyColumnFamiliesRequest_Modification{
			{
				Id:  family,
				Mod: &bttspb.ModifyColumnFamiliesRequest_Modification_Create{Create: &bttdpb.ColumnFamily{}},
			},
		},
	}
	_, err := ac.tClient.ModifyColumnFamilies(ctx, req)
	return err
}

// DeleteTable deletes a table and all of its data.
func (ac *AdminClient) DeleteTable(ctx context.Context, table string) error {
	prefix := ac.instancePrefix()
	req := &bttspb.DeleteTableRequest{
		Name: prefix + "/tables/" + table,
	}
	_, err := ac.tClient.DeleteTable(ctx, req)
	return err
}

// DeleteColumnFamily deletes a column family in a table and all of its data.
func (ac *AdminClient) DeleteColumnFamily(ctx context.Context, table, family string) error {
	prefix := ac.instancePrefix()
	req := &bttspb.ModifyColumnFamiliesRequest{
		Name: prefix + "/tables/" + table,
		Modifications: []*bttspb.ModifyColumnFamiliesRequest_Modification{
			{
				Id:  family,
				Mod: &bttspb.ModifyColumnFamiliesRequest_Modification_Drop{Drop: true},
			},
		},
	}
	_, err := ac.tClient.ModifyColumnFamilies(ctx, req)
	return err
}

// TableInfo represents information about a table.
type TableInfo struct {
	Families []string
}

// TableInfo retrieves information about a table.
func (ac *AdminClient) TableInfo(ctx context.Context, table string) (*TableInfo, error) {
	prefix := ac.instancePrefix()
	req := &bttspb.GetTableRequest{
		Name: prefix + "/tables/" + table,
	}
	res, err := ac.tClient.GetTable(ctx, req)
	if err != nil {
		return nil, err
	}
	ti := &TableInfo{}
	for fam := range res.ColumnFamilies {
		ti.Families = append(ti.Families, fam)
	}
	return ti, nil
}

// SetGCPolicy specifies which cells in a column family should be garbage collected.
// GC executes opportunistically in the background; table reads may return data
// matching the GC policy.
func (ac *AdminClient) SetGCPolicy(ctx context.Context, table, family string, policy GCPolicy) error {
	prefix := ac.instancePrefix()
	req := &bttspb.ModifyColumnFamiliesRequest{
		Name: prefix + "/tables/" + table,
		Modifications: []*bttspb.ModifyColumnFamiliesRequest_Modification{
			{
				Id:  family,
				Mod: &bttspb.ModifyColumnFamiliesRequest_Modification_Update{Update: &bttdpb.ColumnFamily{GcRule: policy.proto()}},
			},
		},
	}
	_, err := ac.tClient.ModifyColumnFamilies(ctx, req)
	return err
}

const instanceAdminAddr = "bigtableadmin.googleapis.com:443"

// InstanceAdminClient is a client type for performing admin operations on instances.
// These operations can be substantially more dangerous than those provided by AdminClient.
type InstanceAdminClient struct {
	conn    *grpc.ClientConn
	iClient btispb.BigtableInstanceAdminClient

	project string
}

// NewInstanceAdminClient creates a new InstanceAdminClient for a given project.
func NewInstanceAdminClient(ctx context.Context, project string, opts ...cloud.ClientOption) (*InstanceAdminClient, error) {
	o := []cloud.ClientOption{
		cloud.WithEndpoint(instanceAdminAddr),
		cloud.WithScopes(InstanceAdminScope),
		cloud.WithUserAgent(clientUserAgent),
	}
	o = append(o, opts...)
	conn, err := transport.DialGRPC(ctx, o...)
	if err != nil {
		return nil, fmt.Errorf("dialing: %v", err)
	}
	return &InstanceAdminClient{
		conn:    conn,
		iClient: btispb.NewBigtableInstanceAdminClient(conn),

		project: project,
	}, nil
}

// Close closes the InstanceAdminClient.
func (iac *InstanceAdminClient) Close() error {
	return iac.conn.Close()
}

// InstanceInfo represents information about an instance
type InstanceInfo struct {
	Name        string // name of the instance
	DisplayName string // display name for UIs
}

var instanceNameRegexp = regexp.MustCompile(`^projects/([^/]+)/instances/([a-z][-a-z0-9]*)$`)

// Instances returns a list of instances in the project.
func (cac *InstanceAdminClient) Instances(ctx context.Context) ([]*InstanceInfo, error) {
	req := &btispb.ListInstancesRequest{
		Name: "projects/" + cac.project,
	}
	res, err := cac.iClient.ListInstances(ctx, req)
	if err != nil {
		return nil, err
	}

	var is []*InstanceInfo
	for _, i := range res.Instances {
		m := instanceNameRegexp.FindStringSubmatch(i.Name)
		if m == nil {
			return nil, fmt.Errorf("malformed instance name %q", i.Name)
		}
		is = append(is, &InstanceInfo{
			Name:        m[2],
			DisplayName: i.DisplayName,
		})
	}
	return is, nil
}
