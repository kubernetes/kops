package assets

import (
	"fmt"
	"github.com/golang/glog"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/kubemanifest"
	"k8s.io/kops/util/pkg/vfs"
	"os"
	"strings"
)

// AssetBuilder discovers and remaps assets
type AssetBuilder struct {
	kopsDistroURL string

	Cluster *kops.Cluster
	Assets  []*Asset
}

type Asset struct {
	Name   string
	Hash   string
	Origin string
	Mirror string
}

func NewAssetBuilder(cluster *kops.Cluster) *AssetBuilder {
	kopsDistroURL := findBaseUrl()
	return &AssetBuilder{
		kopsDistroURL: kopsDistroURL,
		Cluster:       cluster,
	}
}

func (a *AssetBuilder) RemapManifest(data []byte) ([]byte, error) {
	manifests, err := kubemanifest.LoadManifestsFrom(data)
	if err != nil {
		return nil, err
	}

	for _, manifest := range manifests {
		err := manifest.RemapImages(a.remapImage)
		if err != nil {
			return nil, fmt.Errorf("error remapping images: %v", err)
		}
		y, err := manifest.ToYAML()
		if err != nil {
			return nil, fmt.Errorf("error re-marshalling manifest: %v", err)
		}

		glog.Infof("manifest: %v", string(y))
	}

	return data, nil
}

func (a *AssetBuilder) remapImage(image string) (string, error) {
	asset := &Asset{}

	asset.Origin = image

	if strings.HasPrefix(image, "kope/dns-controller:") {
		// To use user-defined DNS Controller:
		// 1. DOCKER_REGISTRY=[your docker hub repo] make dns-controller-push
		// 2. export DNSCONTROLLER_IMAGE=[your docker hub repo]
		// 3. make kops and create/apply cluster
		override := os.Getenv("DNSCONTROLLER_IMAGE")
		if override != "" {
			image = override
		}
	}

	asset.Mirror = image

	a.addAsset(asset)

	return image, nil
}

func IsBaseURL(kubernetesVersion string) bool {
	return strings.HasPrefix(kubernetesVersion, "http:") || strings.HasPrefix(kubernetesVersion, "https:")
}

// ComponentImage returns the docker image name for the specified component
func (a *AssetBuilder) ComponentImage(component string) (string, error) {
	if component == "kube-dns" {
		// TODO: Once we are shipping different versions, start to use them
		return a.remapImage("gcr.io/google_containers/kubedns-amd64:1.3")
	}

	clusterSpec := &a.Cluster.Spec
	if !IsBaseURL(clusterSpec.KubernetesVersion) {
		image := "gcr.io/google_containers/" + component + ":" + "v" + clusterSpec.KubernetesVersion
		return a.remapImage(image)
	}

	baseURL := clusterSpec.KubernetesVersion
	baseURL = strings.TrimSuffix(baseURL, "/")

	tagURL := baseURL + "/bin/linux/amd64/" + component + ".docker_tag"
	glog.V(2).Infof("Downloading docker tag for %s from: %s", component, tagURL)

	b, err := vfs.Context.ReadFile(tagURL)
	if err != nil {
		return "", fmt.Errorf("error reading tag file %q: %v", tagURL, err)
	}
	tag := strings.TrimSpace(string(b))
	glog.V(2).Infof("Found tag %q for %q", tag, component)

	imageName := "gcr.io/google_containers/" + component + ":" + tag
	asset := &Asset{
		Name:   imageName,
		Origin: baseURL + "/bin/linux/amd64/" + component,
	}
	a.addAsset(asset)
	return imageName, nil
}

func (a *AssetBuilder) KubernetesAsset(key string) string {
	cluster := a.Cluster

	var baseURL string
	if IsBaseURL(cluster.Spec.KubernetesVersion) {
		baseURL = cluster.Spec.KubernetesVersion
	} else {
		baseURL = "https://storage.googleapis.com/kubernetes-release/release/v" + cluster.Spec.KubernetesVersion
	}
	baseURL = strings.TrimSuffix(baseURL, "/")
	url := baseURL + "/bin/linux/amd64/" + key
	a.addURLAsset(url)
	return url
}

func (a *AssetBuilder) StaticUtils() string {
	url := a.kopsDistroURL + "linux/amd64/utils.tar.gz"
	a.addURLAsset(url)
	return url
}

func (a *AssetBuilder) addURLAsset(url string) {
	asset := &Asset{
		Origin: url,
	}
	a.addAsset(asset)
}

func (a *AssetBuilder) addAsset(asset *Asset) {
	a.Assets = append(a.Assets, asset)
}
