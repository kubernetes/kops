/*
Copyright 2019 The Kubernetes Authors.

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

package dotasks

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/digitalocean/godo"
	"k8s.io/kops/pkg/resources/digitalocean"
	"k8s.io/kops/upup/pkg/fi"
)

type fakeStorageClient struct {
	listFn           func(context.Context, *godo.ListVolumeParams) ([]godo.Volume, *godo.Response, error)
	getFn            func(context.Context, string) (*godo.Volume, *godo.Response, error)
	createFn         func(context.Context, *godo.VolumeCreateRequest) (*godo.Volume, *godo.Response, error)
	deleteFn         func(context.Context, string) (*godo.Response, error)
	listSnapshotFn   func(ctx context.Context, volumeID string, opts *godo.ListOptions) ([]godo.Snapshot, *godo.Response, error)
	getSnapshotFn    func(context.Context, string) (*godo.Snapshot, *godo.Response, error)
	createSnapshotFn func(context.Context, *godo.SnapshotCreateRequest) (*godo.Snapshot, *godo.Response, error)
	deleteSnapshotFn func(context.Context, string) (*godo.Response, error)
}

func (f fakeStorageClient) ListVolumes(ctx context.Context, listOpts *godo.ListVolumeParams) ([]godo.Volume, *godo.Response, error) {
	return f.listFn(ctx, listOpts)
}

func (f fakeStorageClient) GetVolume(ctx context.Context, id string) (*godo.Volume, *godo.Response, error) {
	return f.getFn(ctx, id)
}

func (f fakeStorageClient) CreateVolume(ctx context.Context, req *godo.VolumeCreateRequest) (*godo.Volume, *godo.Response, error) {
	return f.createFn(ctx, req)
}

func (f fakeStorageClient) DeleteVolume(ctx context.Context, id string) (*godo.Response, error) {
	return f.deleteFn(ctx, id)
}

func (f fakeStorageClient) ListSnapshots(ctx context.Context, volumeID string, opts *godo.ListOptions) ([]godo.Snapshot, *godo.Response, error) {
	return f.listSnapshotFn(ctx, volumeID, opts)
}

func (f fakeStorageClient) GetSnapshot(ctx context.Context, id string) (*godo.Snapshot, *godo.Response, error) {
	return f.getSnapshotFn(ctx, id)
}

func (f fakeStorageClient) CreateSnapshot(ctx context.Context, req *godo.SnapshotCreateRequest) (*godo.Snapshot, *godo.Response, error) {
	return f.createSnapshotFn(ctx, req)
}

func (f fakeStorageClient) DeleteSnapshot(ctx context.Context, id string) (*godo.Response, error) {
	return f.deleteSnapshotFn(ctx, id)
}

func newCloud(client *godo.Client) *digitalocean.Cloud {
	return &digitalocean.Cloud{
		Client:     client,
		RegionName: "nyc1",
	}
}

func newContext(cloud fi.Cloud) *fi.Context {
	return &fi.Context{
		Cloud: cloud,
	}
}

func Test_Find(t *testing.T) {
	testcases := []struct {
		name      string
		storage   fakeStorageClient
		inVolume  *Volume
		outVolume *Volume
		err       error
	}{
		{
			"successfully found volume",
			fakeStorageClient{
				listFn: func(context.Context, *godo.ListVolumeParams) (
					[]godo.Volume, *godo.Response, error) {
					return []godo.Volume{
						{
							Name:          "test0",
							ID:            "100",
							SizeGigaBytes: int64(100),
							Region:        &godo.Region{Slug: "nyc1"},
						},
					}, nil, nil
				},
			},
			&Volume{
				Name:   fi.String("test0"),
				SizeGB: fi.Int64(int64(100)),
				Region: fi.String("nyc1"),
			},
			&Volume{
				Name:   fi.String("test0"),
				ID:     fi.String("100"),
				SizeGB: fi.Int64(int64(100)),
				Region: fi.String("nyc1"),
			},
			nil,
		},
		{
			"no volume found",
			fakeStorageClient{
				listFn: func(context.Context, *godo.ListVolumeParams) (
					[]godo.Volume, *godo.Response, error) {
					return []godo.Volume{}, nil, nil
				},
			},
			&Volume{
				Name:   fi.String("test1"),
				SizeGB: fi.Int64(int64(100)),
				Region: fi.String("nyc1"),
			},
			nil,
			nil,
		},
		{
			"error from server",
			fakeStorageClient{
				listFn: func(context.Context, *godo.ListVolumeParams) (
					[]godo.Volume, *godo.Response, error) {
					return []godo.Volume{}, nil, errors.New("error!")
				},
			},
			&Volume{
				Name:   fi.String("test1"),
				SizeGB: fi.Int64(int64(100)),
				Region: fi.String("nyc1"),
			},
			nil,
			errors.New("error!"),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			cloud := newCloud(godo.NewClient(nil))
			cloud.Client.Storage = tc.storage
			ctx := newContext(cloud)

			actualVolume, err := tc.inVolume.Find(ctx)
			if !reflect.DeepEqual(actualVolume, tc.outVolume) {
				t.Error("unexpected volume")
				t.Logf("actual volume: %v", actualVolume)
				t.Logf("expected volume: %v", tc.outVolume)
			}

			if !reflect.DeepEqual(err, tc.err) {
				t.Error("unexpected error")
				t.Logf("actual err: %v", err)
				t.Logf("expected err: %v", tc.err)
			}
		})
	}
}
