/*
Copyright 2016 The Kubernetes Authors.

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

package vfsclientset

import (
	"bytes"
	"fmt"
	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	kops "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/v1alpha2"
	"k8s.io/kops/util/pkg/vfs"
	"os"
	"reflect"
	"sort"
	"time"
)

var StoreVersion = v1alpha2.SchemeGroupVersion

type ValidationFunction func(o runtime.Object) error

type commonVFS struct {
	kind               string
	basePath           vfs.Path
	decoder            runtime.Decoder
	encoder            runtime.Encoder
	defaultReadVersion *schema.GroupVersionKind
	validate           ValidationFunction
}

func (c *commonVFS) init(kind string, basePath vfs.Path, storeVersion runtime.GroupVersioner) {
	yaml, ok := runtime.SerializerInfoForMediaType(kops.Codecs.SupportedMediaTypes(), "application/yaml")
	if !ok {
		glog.Fatalf("no YAML serializer registered")
	}
	c.encoder = kops.Codecs.EncoderForVersion(yaml.Serializer, storeVersion)
	c.decoder = kops.Codecs.DecoderToVersion(yaml.Serializer, kops.SchemeGroupVersion)

	c.kind = kind
	c.basePath = basePath
}

func (c *commonVFS) get(name string) (runtime.Object, error) {
	o, err := c.readConfig(c.basePath.Join(name))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("error reading %s %q: %v", c.kind, name, err)
	}
	return o, nil
}

func (c *commonVFS) list(items interface{}, options metav1.ListOptions) (interface{}, error) {
	return c.readAll(items)
}

func (c *commonVFS) create(i runtime.Object) error {
	objectMeta, err := meta.Accessor(i)
	if err != nil {
		return err
	}

	if c.validate != nil {
		err = c.validate(i)
		if err != nil {
			return err
		}
	}

	creationTimestamp := objectMeta.GetCreationTimestamp()
	if creationTimestamp.IsZero() {
		objectMeta.SetCreationTimestamp(v1.NewTime(time.Now().UTC()))
	}

	err = c.writeConfig(c.basePath.Join(objectMeta.GetName()), i, vfs.WriteOptionCreate)
	if err != nil {
		if os.IsExist(err) {
			return err
		}
		return fmt.Errorf("error writing %s: %v", c.kind, err)
	}

	return nil
}

func (c *commonVFS) serialize(o runtime.Object) ([]byte, error) {
	var b bytes.Buffer
	err := c.encoder.Encode(o, &b)
	if err != nil {
		return nil, fmt.Errorf("error encoding object: %v", err)
	}

	return b.Bytes(), nil
}

func (c *commonVFS) readConfig(configPath vfs.Path) (runtime.Object, error) {
	data, err := configPath.ReadFile()
	if err != nil {
		if os.IsNotExist(err) {
			return nil, err
		}
		return nil, fmt.Errorf("error reading %s: %v", configPath, err)
	}

	object, _, err := c.decoder.Decode(data, c.defaultReadVersion, nil)
	if err != nil {
		return nil, fmt.Errorf("error parsing %s: %v", configPath, err)
	}
	return object, nil
}

func (c *commonVFS) writeConfig(configPath vfs.Path, o runtime.Object, writeOptions ...vfs.WriteOption) error {
	data, err := c.serialize(o)
	if err != nil {
		return fmt.Errorf("error marshalling object: %v", err)
	}

	create := false
	for _, writeOption := range writeOptions {
		switch writeOption {
		case vfs.WriteOptionCreate:
			create = true
		case vfs.WriteOptionOnlyIfExists:
			_, err = configPath.ReadFile()
			if err != nil {
				if os.IsNotExist(err) {
					return fmt.Errorf("cannot update configuration file %s: does not exist", configPath)
				}
				return fmt.Errorf("error checking if configuration file %s exists already: %v", configPath, err)
			}
		default:
			return fmt.Errorf("unknown write option: %q", writeOption)
		}
	}

	if create {
		err = configPath.CreateFile(data)
	} else {
		err = configPath.WriteFile(data)
	}
	if err != nil {
		if create && os.IsExist(err) {
			glog.Warningf("failed to create file as already exists: %v", configPath)
			return err
		}
		return fmt.Errorf("error writing configuration file %s: %v", configPath, err)
	}
	return nil
}

func (c *commonVFS) update(i runtime.Object) error {
	objectMeta, err := meta.Accessor(i)
	if err != nil {
		return err
	}

	if c.validate != nil {
		err = c.validate(i)
		if err != nil {
			return err
		}
	}

	creationTimestamp := objectMeta.GetCreationTimestamp()
	if creationTimestamp.IsZero() {
		objectMeta.SetCreationTimestamp(v1.NewTime(time.Now().UTC()))
	}

	err = c.writeConfig(c.basePath.Join(objectMeta.GetName()), i, vfs.WriteOptionOnlyIfExists)
	if err != nil {
		return fmt.Errorf("error writing %s: %v", c.kind, err)
	}

	return nil
}

func (c *commonVFS) delete(name string, options *metav1.DeleteOptions) error {
	p := c.basePath.Join(name)
	err := p.Remove()
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("error deleting %s configuration %q: %v", c.kind, name, err)
	}
	return nil
}

func (c *commonVFS) listNames() ([]string, error) {
	keys, err := listChildNames(c.basePath)
	if err != nil {
		return nil, fmt.Errorf("error listing %s in state store: %v", c.kind, err)
	}

	// Seems to be an assumption in k8s APIs that items are always returned sorted
	sort.Strings(keys)

	return keys, nil
}

func (c *commonVFS) readAll(items interface{}) (interface{}, error) {
	sliceValue := reflect.ValueOf(items)
	sliceType := reflect.TypeOf(items)
	if sliceType.Kind() != reflect.Slice {
		return nil, fmt.Errorf("expected slice, got %T", items)
	}

	names, err := c.listNames()
	if err != nil {
		return nil, err
	}

	for _, name := range names {
		o, err := c.get(name)
		if err != nil {
			return nil, err
		}

		if o == nil {
			return nil, fmt.Errorf("%s was listed, but then not found %q", c.kind, name)
		}

		sliceValue = reflect.Append(sliceValue, reflect.ValueOf(o).Elem())
	}

	return sliceValue.Interface(), nil
}
