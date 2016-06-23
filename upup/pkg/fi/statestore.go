package fi

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/golang/glog"
	"k8s.io/kube-deploy/upup/pkg/fi/utils"
	"k8s.io/kube-deploy/upup/pkg/fi/vfs"
	"net/url"
	"os"
	"strings"
)

type StateStore interface {
	// VFSPath returns the path where the StateStore is stored
	VFSPath() vfs.Path

	CA() CAStore
	Secrets() SecretStore

	ReadConfig(config interface{}) error
	WriteConfig(config interface{}) error
}

type VFSStateStore struct {
	location vfs.Path
	ca       CAStore
	secrets  SecretStore
}

var _ StateStore = &VFSStateStore{}

func NewVFSStateStore(location vfs.Path, dryrun bool) (*VFSStateStore, error) {
	s := &VFSStateStore{
		location: location,
	}
	var err error
	s.ca, err = NewVFSCAStore(location.Join("pki"), dryrun)
	if err != nil {
		return nil, fmt.Errorf("error building CA store: %v", err)
	}
	s.secrets, err = NewVFSSecretStore(location.Join("secrets"))
	if err != nil {
		return nil, fmt.Errorf("error building secret store: %v", err)
	}

	return s, nil
}

func (s *VFSStateStore) CA() CAStore {
	return s.ca
}

func (s *VFSStateStore) VFSPath() vfs.Path {
	return s.location
}

func (s *VFSStateStore) Secrets() SecretStore {
	return s.secrets
}

func (s *VFSStateStore) ReadConfig(config interface{}) error {
	configPath := s.location.Join("config")
	data, err := configPath.ReadFile()
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("error reading configuration file %s: %v", configPath, err)
	}

	// Yaml can't parse empty strings
	configString := string(data)
	configString = strings.TrimSpace(configString)

	if configString != "" {
		err = utils.YamlUnmarshal([]byte(configString), config)
		if err != nil {
			return fmt.Errorf("error parsing configuration: %v", err)
		}
	}

	return nil
}

func (s *VFSStateStore) WriteConfig(config interface{}) error {
	configPath := s.location.Join("config")

	data, err := utils.YamlMarshal(config)
	if err != nil {
		return fmt.Errorf("error marshalling configuration: %v", err)
	}

	err = configPath.WriteFile(data)
	if err != nil {
		return fmt.Errorf("error writing configuration file %s: %v", configPath, err)
	}
	return nil
}

func BuildVfsPath(p string) (vfs.Path, error) {
	if strings.HasPrefix(p, "s3://") {
		u, err := url.Parse(p)
		if err != nil {
			return nil, fmt.Errorf("invalid s3 path: %q", err)
		}

		var region string
		{
			config := aws.NewConfig().WithRegion("us-east-1")
			session := session.New()
			s3Client := s3.New(session, config)

			bucket := strings.TrimSuffix(u.Host, "/")
			request := &s3.GetBucketLocationInput{}
			request.Bucket = aws.String(bucket)

			response, err := s3Client.GetBucketLocation(request)
			if err != nil {
				// TODO: Auto-create bucket?
				return nil, fmt.Errorf("error getting location for S3 bucket %q: %v", bucket, err)
			}
			if response.LocationConstraint == nil {
				// US Classic does not return a region
				region = "us-east-1"
			} else {
				region = *response.LocationConstraint
				// Another special case: "EU" can mean eu-west-1
				if region == "EU" {
					region = "eu-west-1"
				}
			}
			glog.V(2).Infof("Found bucket %q in region %q", bucket, region)
		}

		{
			config := aws.NewConfig().WithRegion(region)
			session := session.New()
			s3Client := s3.New(session, config)

			s3path := vfs.NewS3Path(s3Client, u.Host, u.Path)
			return s3path, nil
		}
	}

	return nil, fmt.Errorf("unknown / unhandled path type: %q", p)
}
