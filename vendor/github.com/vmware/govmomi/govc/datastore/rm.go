/*
Copyright (c) 2014-2015 VMware, Inc. All Rights Reserved.

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

package datastore

import (
	"context"
	"errors"
	"flag"

	"github.com/vmware/govmomi/govc/cli"
	"github.com/vmware/govmomi/govc/flags"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/types"
)

type rm struct {
	*flags.DatastoreFlag

	force       bool
	isNamespace bool
}

func init() {
	cli.Register("datastore.rm", &rm{})
	cli.Alias("datastore.rm", "datastore.delete")
}

func (cmd *rm) Register(ctx context.Context, f *flag.FlagSet) {
	cmd.DatastoreFlag, ctx = flags.NewDatastoreFlag(ctx)
	cmd.DatastoreFlag.Register(ctx, f)

	f.BoolVar(&cmd.force, "f", false, "Force; ignore nonexistent files and arguments")
	f.BoolVar(&cmd.isNamespace, "namespace", false, "Path is uuid of namespace on vsan datastore")
}

func (cmd *rm) Process(ctx context.Context) error {
	if err := cmd.DatastoreFlag.Process(ctx); err != nil {
		return err
	}
	return nil
}

func (cmd *rm) Usage() string {
	return "FILE"
}

func (cmd *rm) Run(ctx context.Context, f *flag.FlagSet) error {
	args := f.Args()
	if len(args) == 0 {
		return errors.New("missing operand")
	}

	c, err := cmd.Client()
	if err != nil {
		return err
	}

	var dc *object.Datacenter
	dc, err = cmd.Datacenter()
	if err != nil {
		return err
	}

	if cmd.isNamespace {
		path := args[0]

		nm := object.NewDatastoreNamespaceManager(c)
		err = nm.DeleteDirectory(ctx, dc, path)
	} else {
		var path string
		var task *object.Task

		// TODO(PN): Accept multiple args
		path, err = cmd.DatastorePath(args[0])
		if err != nil {
			return err
		}

		m := object.NewFileManager(c)
		task, err = m.DeleteDatastoreFile(ctx, path, dc)
		if err != nil {
			return err
		}

		err = task.Wait(ctx)
	}

	if err != nil {
		if types.IsFileNotFound(err) && cmd.force {
			// Ignore error
			return nil
		}
	}

	return err
}
