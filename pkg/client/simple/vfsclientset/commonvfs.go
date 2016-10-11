package vfsclientset

import (
	k8sapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"time"
	"fmt"
	"os"
	"k8s.io/kops/util/pkg/vfs"
	"k8s.io/kops/pkg/apis/kops/registry"
	"reflect"
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kubernetes/pkg/runtime/serializer/json"
	"bytes"
	"k8s.io/kubernetes/pkg/runtime"
)

type commonVFS struct {
	key      string
	basePath vfs.Path
	encoder runtime.Encoder
}

func (c*commonVFS) init(key string, basePath vfs.Path, storeVersion runtime.GroupVersioner) {
	yamlSerde := json.NewYAMLSerializer(json.DefaultMetaFactory, k8sapi.Scheme, k8sapi.Scheme)
	encoder := k8sapi.Codecs.EncoderForVersion(yamlSerde, storeVersion)

	c.key = key
	c.basePath = basePath
	c.encoder = encoder
}

func (c *commonVFS) get(name string, dest interface{}) (bool, error) {
	err := registry.ReadConfig(c.basePath.Join(name), dest)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("error reading %s %q: %v", c.key, name, err)
	}
	return true, nil
}

func (c *commonVFS) list(items interface{}, options k8sapi.ListOptions) (interface{}, error) {
	return c.readAll(items)
}

func (c *commonVFS) create(i api.ApiType) (error) {
	objectMeta, err := k8sapi.ObjectMetaFor(i)
	if err != nil {
		return err
	}

	err = i.Validate()
	if err != nil {
		return err
	}

	if objectMeta.CreationTimestamp.IsZero() {
		objectMeta.CreationTimestamp = unversioned.NewTime(time.Now().UTC())
	}

	err = c.writeConfig(c.basePath.Join(objectMeta.Name), i, vfs.WriteOptionCreate)
	if err != nil {
		if os.IsExist(err) {
			return err
		}
		return fmt.Errorf("error writing %s: %v", c.key, err)
	}

	return nil
}


func (c*commonVFS) serialize(o runtime.Object) ([]byte, error) {
	var b bytes.Buffer
	err := c.encoder.Encode(o, &b)
	if err != nil {
		return nil, fmt.Errorf("error encoding object: %v", err)
	}

	return b.Bytes(), nil
}


func (c*commonVFS) writeConfig(configPath vfs.Path, o runtime.Object, writeOptions ...vfs.WriteOption) error {
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
			return err
		}
		return fmt.Errorf("error writing configuration file %s: %v", configPath, err)
	}
	return nil
}


func (c *commonVFS) update(i api.ApiType) (error) {
	objectMeta, err := k8sapi.ObjectMetaFor(i)
	if err != nil {
		return err
	}

	err = i.Validate()
	if err != nil {
		return err
	}

	if objectMeta.CreationTimestamp.IsZero() {
		objectMeta.CreationTimestamp = unversioned.NewTime(time.Now().UTC())
	}

	err = registry.WriteConfig(c.basePath.Join(objectMeta.Name), i, vfs.WriteOptionOnlyIfExists)
	if err != nil {
		return fmt.Errorf("error writing %s: %v", c.key, err)
	}

	return nil
}

func (c *commonVFS) delete(name string, options *k8sapi.DeleteOptions) (error) {
	p := c.basePath.Join(name)
	err := p.Remove()
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("error deleting %s configuration %q: %v", c.key, name, err)
	}
	return nil
}

func (c *commonVFS) listNames() ([]string, error) {
	keys, err := listChildNames(c.basePath)
	if err != nil {
		return nil, fmt.Errorf("error listing %s in state store: %v", c.key, err)
	}
	return keys, nil
}

func (c *commonVFS) readAll(items interface{}) (interface{}, error) {
	sliceValue := reflect.ValueOf(items)
	sliceType := reflect.TypeOf(items)
	if sliceType.Kind() != reflect.Slice {
		return nil, fmt.Errorf("expected slice, got %T", items)
	}

	elemType := sliceType.Elem()

	names, err := c.listNames()
	if err != nil {
		return nil, err
	}

	for _, name := range names {
		elemValue := reflect.New(elemType)

		found, err := c.get(name, elemValue.Interface())
		if err != nil {
			return nil, err
		}

		if !found {
			return nil, fmt.Errorf("%s was listed, but then not found %q", c.key, name)
		}

		sliceValue = reflect.Append(sliceValue, elemValue.Elem())
	}

	return sliceValue.Interface(), nil
}

