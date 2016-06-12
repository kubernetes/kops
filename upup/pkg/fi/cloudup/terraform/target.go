package terraform

import (
	"encoding/json"
	"fmt"
	"github.com/golang/glog"
	hcl_parser "github.com/hashicorp/hcl/json/parser"
	"io/ioutil"
	"k8s.io/kube-deploy/upup/pkg/fi"
	"os"
	"path"
	"strings"
)

type TerraformTarget struct {
	Cloud   fi.Cloud
	Region  string
	Project string

	resources []*terraformResource

	files  map[string][]byte
	outDir string
}

func NewTerraformTarget(cloud fi.Cloud, region, project string, outDir string) *TerraformTarget {
	return &TerraformTarget{
		Cloud:   cloud,
		Region:  region,
		Project: project,
		outDir:  outDir,
		files:   make(map[string][]byte),
	}
}

var _ fi.Target = &TerraformTarget{}

type terraformResource struct {
	ResourceType string
	ResourceName string
	Item         interface{}
}

// A TF name can't have dots in it (if we want to refer to it from a literal),
// so we replace them
func tfSanitize(name string) string {
	name = strings.Replace(name, ".", "-", -1)
	name = strings.Replace(name, "/", "--", -1)
	return name
}

func (t *TerraformTarget) AddFile(resourceType string, resourceName string, key string, r fi.Resource) (*Literal, error) {
	id := resourceType + "_" + resourceName + "_" + key

	d, err := fi.ResourceAsBytes(r)
	if err != nil {
		return nil, fmt.Errorf("error rending resource %s %v", id, err)
	}

	p := path.Join("data", id)
	t.files[p] = d

	l := LiteralExpression(fmt.Sprintf("${file(%q)}", p))
	return l, nil
}

func (t *TerraformTarget) RenderResource(resourceType string, resourceName string, e interface{}) error {
	res := &terraformResource{
		ResourceType: resourceType,
		ResourceName: resourceName,
		Item:         e,
	}

	t.resources = append(t.resources, res)

	return nil
}

func (t *TerraformTarget) Finish(taskMap map[string]fi.Task) error {
	resourcesByType := make(map[string]map[string]interface{})

	for _, res := range t.resources {
		resources := resourcesByType[res.ResourceType]
		if resources == nil {
			resources = make(map[string]interface{})
			resourcesByType[res.ResourceType] = resources
		}

		tfName := tfSanitize(res.ResourceName)

		if resources[tfName] != nil {
			return fmt.Errorf("duplicate resource found: %s.%s", res.ResourceType, tfName)
		}

		resources[tfName] = res.Item
	}

	providersByName := make(map[string]map[string]interface{})
	if t.Cloud.ProviderID() == fi.CloudProviderGCE {
		providerGoogle := make(map[string]interface{})
		providerGoogle["project"] = t.Project
		providerGoogle["region"] = t.Region
		providersByName["google"] = providerGoogle
	}

	data := make(map[string]interface{})
	data["resource"] = resourcesByType
	if len(providersByName) != 0 {
		data["provider"] = providersByName
	}

	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling terraform data to json: %v", err)
	}

	useJson := false

	if useJson {
		t.files["kubernetes.tf"] = jsonBytes
	} else {
		f, err := hcl_parser.Parse(jsonBytes)
		if err != nil {
			return fmt.Errorf("error parsing terraform json: %v", err)
		}

		b, err := hclPrint(f)
		if err != nil {
			return fmt.Errorf("error writing terraform data to output: %v", err)
		}

		t.files["kubernetes.tf"] = b

	}

	for relativePath, contents := range t.files {
		p := path.Join(t.outDir, relativePath)

		err = os.MkdirAll(path.Dir(p), os.FileMode(0755))
		if err != nil {
			return fmt.Errorf("error creating output directory %q: %v", path.Dir(p), err)
		}

		err = ioutil.WriteFile(p, contents, os.FileMode(0644))
		if err != nil {
			return fmt.Errorf("error writing terraform data to output file %q: %v", p, err)
		}
	}

	glog.Infof("Terraform output is in %s", t.outDir)

	return nil
}
