/*
Copyright 2017 The Kubernetes Authors.

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

package build

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"

	"github.com/kubernetes-incubator/apiserver-builder/cmd/apiserver-boot/boot/util"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var Name, Namespace string
var Versions []schema.GroupVersion
var ResourceConfigDir string
var ControllerArgs []string
var ApiserverArgs []string
var ControllerSecret string
var ControllerSecretMount string
var ControllerSecretEnv []string
var ImagePullSecrets []string
var ServiceAccount string

var LocalMinikube bool
var LocalIp string

var buildResourceConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Create kubernetes resource config files to launch the apiserver.",
	Long:  `Create kubernetes resource config files to launch the apiserver.`,
	Example: `
# Build yaml resource config into the config/ directory for running the apiserver and
# controller-manager as an aggregated service in a Kubernetes cluster as a container.
# Generates CA and apiserver certificates.
apiserver-boot build config --name nameofservice --namespace mysystemnamespace --image gcr.io/myrepo/myimage:mytag

# Build yaml resource config into the config/ directory for running the apiserver and
# controller-manager locally, but registered through aggregation into a local minikube cluster
# Generates CA and apiserver certificates.
apiserver-boot build config --name nameofservice --namespace mysystemnamespace --local-minikube
`,
	Run: RunBuildResourceConfig,
}

func AddBuildResourceConfig(cmd *cobra.Command) {
	cmd.AddCommand(buildResourceConfigCmd)
	AddBuildResourceConfigFlags(buildResourceConfigCmd)
}

func AddBuildResourceConfigFlags(cmd *cobra.Command) {
	cmd.Flags().StringSliceVar(&ControllerSecretEnv, "controller-env", []string{}, "")
	cmd.Flags().StringVar(&ControllerSecret, "controller-secret", "", "")
	cmd.Flags().StringVar(&ControllerSecretMount, "controller-secret-mount", "", "")
	cmd.Flags().StringSliceVar(&ControllerArgs, "controller-args", []string{}, "")
	cmd.Flags().StringSliceVar(&ApiserverArgs, "apiserver-args", []string{}, "")
	cmd.Flags().StringVar(&Name, "name", "", "")
	cmd.Flags().StringVar(&Namespace, "namespace", "", "")
	cmd.Flags().StringSliceVar(&ImagePullSecrets, "image-pull-secrets", []string{}, "List of secret names for docker registry")
	cmd.Flags().StringVar(&ServiceAccount, "service-account", "", "Name of service account that will be attached to deployed pod")
	cmd.Flags().StringVar(&Image, "image", "", "name of the apiserver Image with tag")
	cmd.Flags().StringVar(&ResourceConfigDir, "output", "config", "directory to output resourceconfig")

	cmd.Flags().BoolVar(&LocalMinikube, "local-minikube", false, "if true, generate config to run locally but aggregate through minikube.")
	cmd.Flags().StringVar(&LocalIp, "local-ip", "10.0.2.2", "if using --local-minikube, this is the ip address minikube will look for the aggregated server at.")
}

func RunBuildResourceConfig(cmd *cobra.Command, args []string) {
	if len(Name) == 0 {
		log.Fatalf("must specify --name")
	}
	if len(Namespace) == 0 {
		log.Fatalf("must specify --namespace")
	}
	if len(Image) == 0 && !LocalMinikube {
		log.Fatalf("Must specify --image")
	}
	util.GetDomain()

	if _, err := os.Stat("pkg"); err != nil {
		log.Fatalf("could not find 'pkg' directory.  must run apiserver-boot init before generating config")
	}

	createCerts()
	buildResourceConfig()
}

func getBase64(file string) string {
	//out, err := exec.Command("bash", "-c",
	//	fmt.Sprintf("base64 %s | awk 'BEGIN{ORS=\"\";} {print}'", file)).CombinedOutput()
	//if err != nil {
	//	log.Fatalf("Could not base64 encode file: %v", err)
	//}

	buff := bytes.Buffer{}
	enc := base64.NewEncoder(base64.StdEncoding, &buff)
	data, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatalf("Could not read file %s: %v", file, err)
	}

	_, err = enc.Write(data)
	if err != nil {
		log.Fatalf("Could not write bytes: %v", err)
	}
	enc.Close()
	return buff.String()

	//if string(out) != buff.String() {
	//	fmt.Printf("\nNot Equal\n")
	//}
	//
	//return string(out)
}

func buildResourceConfig() {
	initVersionedApis()
	dir := filepath.Join(ResourceConfigDir, "certificates")

	a := resourceConfigTemplateArgs{
		Name:                  Name,
		Namespace:             Namespace,
		Image:                 Image,
		Domain:                util.Domain,
		Versions:              Versions,
		ClientKey:             getBase64(filepath.Join(dir, "apiserver.key")),
		CACert:                getBase64(filepath.Join(dir, "apiserver_ca.crt")),
		ClientCert:            getBase64(filepath.Join(dir, "apiserver.crt")),
		ApiserverArgs:         ApiserverArgs,
		ControllerArgs:        ControllerArgs,
		ControllerSecretMount: ControllerSecretMount,
		ControllerSecret:      ControllerSecret,
		ControllerSecretEnv:   ControllerSecretEnv,
		LocalIp:               LocalIp,
		ImagePullSecrets:      ImagePullSecrets,
		ServiceAccount:        ServiceAccount,
	}
	path := filepath.Join(ResourceConfigDir, "apiserver.yaml")

	temp := resourceConfigTemplate
	if LocalMinikube {
		temp = localConfigTemplate
	}
	created := util.WriteIfNotFound(path, "config-template", temp, a)
	if !created {
		log.Fatalf("Resource config already exists.")
	}
}

func createCerts() {
	dir := filepath.Join(ResourceConfigDir, "certificates")
	os.MkdirAll(dir, 0700)

	if _, err := os.Stat(filepath.Join(dir, "apiserver_ca.crt")); os.IsNotExist(err) {
		util.DoCmd("openssl", "req", "-x509",
			"-newkey", "rsa:2048",
			"-keyout", filepath.Join(dir, "apiserver_ca.key"),
			"-out", filepath.Join(dir, "apiserver_ca.crt"),
			"-days", "365",
			"-nodes",
			"-subj", fmt.Sprintf("/C=un/ST=st/L=l/O=o/OU=ou/CN=%s-certificate-authority", Name),
		)
	} else {
		log.Printf("Skipping generate CA cert.  File already exists.")
	}

	if _, err := os.Stat(filepath.Join(dir, "apiserver.csr")); os.IsNotExist(err) {
		// Use <service-Name>.<Namespace>.svc as the domain Name for the certificate
		util.DoCmd("openssl", "req",
			"-out", filepath.Join(dir, "apiserver.csr"),
			"-new",
			"-newkey", "rsa:2048",
			"-nodes",
			"-keyout", filepath.Join(dir, "apiserver.key"),
			"-subj", fmt.Sprintf("/C=un/ST=st/L=l/O=o/OU=ou/CN=%s.%s.svc", Name, Namespace),
		)
	} else {
		log.Printf("Skipping generate apiserver csr.  File already exists.")
	}

	if _, err := os.Stat(filepath.Join(dir, "apiserver.crt")); os.IsNotExist(err) {
		util.DoCmd("openssl", "x509", "-req",
			"-days", "365",
			"-in", filepath.Join(dir, "apiserver.csr"),
			"-CA", filepath.Join(dir, "apiserver_ca.crt"),
			"-CAkey", filepath.Join(dir, "apiserver_ca.key"),
			"-CAcreateserial",
			"-out", filepath.Join(dir, "apiserver.crt"),
		)
	} else {
		log.Printf("Skipping signing apiserver crt.  File already exists.")
	}
}

func initVersionedApis() {
	groups, err := ioutil.ReadDir(filepath.Join("pkg", "apis"))
	if err != nil {
		log.Fatalf("could not read pkg/apis directory to find api Versions")
	}
	log.Printf("Adding APIs:")
	for _, g := range groups {
		if g.IsDir() {
			versionFiles, err := ioutil.ReadDir(filepath.Join("pkg", "apis", g.Name()))
			if err != nil {
				log.Fatalf("could not read pkg/apis/%s directory to find api Versions", g.Name())
			}
			versionMatch := regexp.MustCompile("^v\\d+(alpha\\d+|beta\\d+)*$")
			for _, v := range versionFiles {
				if v.IsDir() && versionMatch.MatchString(v.Name()) {
					log.Printf("\t%s.%s", g.Name(), v.Name())
					Versions = append(Versions, schema.GroupVersion{
						Group:   g.Name(),
						Version: v.Name(),
					})
				}
			}
		}
	}
	u := map[string]bool{}
	for _, a := range versionedAPIs {
		u[path.Dir(a)] = true
	}
	for a, _ := range u {
		unversionedAPIs = append(unversionedAPIs, a)
	}
}

type resourceConfigTemplateArgs struct {
	Versions              []schema.GroupVersion
	CACert                string
	ClientCert            string
	ClientKey             string
	Domain                string
	Name                  string
	Namespace             string
	Image                 string
	ApiserverArgs         []string
	ControllerArgs        []string
	ControllerSecret      string
	ControllerSecretMount string
	ControllerSecretEnv   []string
	LocalIp               string
	ServiceAccount        string
	ImagePullSecrets      []string
}

var resourceConfigTemplate = `
{{ $config := . -}}
{{ range $api := .Versions -}}
apiVersion: apiregistration.k8s.io/v1beta1
kind: APIService
metadata:
  name: {{ $api.Version }}.{{ $api.Group }}.{{ $config.Domain }}
  labels:
    api: {{ $config.Name }}
    apiserver: "true"
spec:
  version: {{ $api.Version }}
  group: {{ $api.Group }}.{{ $config.Domain }}
  groupPriorityMinimum: 2000
  priority: 200
  service:
    name: {{ $config.Name }}
    namespace: {{ $config.Namespace }}
  versionPriority: 10
  caBundle: "{{ $config.CACert }}"
---
{{ end -}}
apiVersion: v1
kind: Service
metadata:
  name: {{.Name}}
  namespace: {{.Namespace}}
  labels:
    api: {{.Name}}
    apiserver: "true"
spec:
  ports:
  - port: 443
    protocol: TCP
    targetPort: 443
  selector:
    api: {{ .Name }}
    apiserver: "true"
---
apiVersion: apps/v1beta1
kind: Deployment
metadata:
  name: {{.Name}}
  namespace: {{.Namespace}}
  labels:
    api: {{.Name}}
    apiserver: "true"
spec:
  replicas: 1
  template:
    metadata:
      labels:
        api: {{.Name}}
        apiserver: "true"
    spec:
      {{- if .ImagePullSecrets }}
      imagePullSecrets:
      {{range .ImagePullSecrets }}- name: {{.}}
      {{ end }}
      {{- end -}}
      {{- if .ServiceAccount }}
      serviceAccount: {{.ServiceAccount}}
      {{- end }}
      containers:
      - name: apiserver
        image: {{.Image}}
        volumeMounts:
        - name: apiserver-certs
          mountPath: /apiserver.local.config/certificates
          readOnly: true
        command:
        - "./apiserver"
        args:
        - "--etcd-servers=http://etcd-svc:2379"
        - "--tls-cert-file=/apiserver.local.config/certificates/tls.crt"
        - "--tls-private-key-file=/apiserver.local.config/certificates/tls.key"
        - "--audit-log-path=-"
        - "--audit-log-maxage=0"
        - "--audit-log-maxbackup=0"{{ range $arg := .ApiserverArgs }}
        - "{{ $arg }}"{{ end }}
        resources:
          requests:
            cpu: 100m
            memory: 20Mi
          limits:
            cpu: 100m
            memory: 30Mi
      - name: controller
        image: {{.Image}}
        command:
        - "./controller-manager"
        args:{{ range $arg := .ControllerArgs }}
        - "{{ $arg }}"{{ end }}
        resources:
          requests:
            cpu: 100m
            memory: 20Mi
          limits:
            cpu: 100m
            memory: 30Mi
      volumes:
      - name: apiserver-certs
        secret:
          secretName: {{ .Name }}
---
apiVersion: apps/v1beta1
kind: StatefulSet
metadata:
  name: etcd
  namespace: {{ .Namespace }}
spec:
  serviceName: "etcd"
  replicas: 1
  template:
    metadata:
      labels:
        app: etcd
    spec:
      terminationGracePeriodSeconds: 10
      containers:
      - name: etcd
        image: quay.io/coreos/etcd:latest
        imagePullPolicy: Always
        resources:
          requests:
            cpu: 100m
            memory: 20Mi
          limits:
            cpu: 100m
            memory: 30Mi
        env:
        - name: ETCD_DATA_DIR
          value: /etcd-data-dir
        command:
        - /usr/local/bin/etcd
        - --listen-client-urls
        - http://0.0.0.0:2379
        - --advertise-client-urls
        - http://localhost:2379
        ports:
        - containerPort: 2379
        volumeMounts:
        - name: etcd-data-dir
          mountPath: /etcd-data-dir
        readinessProbe:
          httpGet:
            port: 2379
            path: /health
          failureThreshold: 1
          initialDelaySeconds: 10
          periodSeconds: 10
          successThreshold: 1
          timeoutSeconds: 2
        livenessProbe:
          httpGet:
            port: 2379
            path: /health
          failureThreshold: 3
          initialDelaySeconds: 10
          periodSeconds: 10
          successThreshold: 1
          timeoutSeconds: 2
  volumeClaimTemplates:
  - metadata:
     name: etcd-data-dir
     annotations:
        volume.beta.kubernetes.io/storage-class: standard
    spec:
      accessModes: [ "ReadWriteOnce" ]
      resources:
        requests:
         storage: 10Gi
---
apiVersion: v1
kind: Service
metadata:
  name: etcd-svc
  namespace: {{ .Namespace }}
  labels:
    app: etcd
spec:
  ports:
  - port: 2379
    name: etcd
    targetPort: 2379
  selector:
    app: etcd
---
apiVersion: v1
kind: Secret
type: kubernetes.io/tls
metadata:
  name: {{.Name}}
  namespace: {{.Namespace}}
  labels:
    api: {{.Name}}
    apiserver: "true"
data:
  tls.crt: {{ .ClientCert }}
  tls.key: {{ .ClientKey }}
`

var localConfigTemplate = `
{{ $config := . -}}
{{ range $api := .Versions -}}
apiVersion: apiregistration.k8s.io/v1beta1
kind: APIService
metadata:
  name: {{ $api.Version }}.{{ $api.Group }}.{{ $config.Domain }}
  labels:
    api: {{ $config.Name }}
    apiserver: "true"
spec:
  version: {{ $api.Version }}
  group: {{ $api.Group }}.{{ $config.Domain }}
  groupPriorityMinimum: 2000
  priority: 200
  service:
    name: {{ $config.Name }}
    namespace: {{ $config.Namespace }}
  versionPriority: 10
  caBundle: "{{ $config.CACert }}"
---
{{ end -}}
apiVersion: v1
kind: Service
metadata:
  name: {{.Name}}
  namespace: {{.Namespace}}
  labels:
    api: {{.Name}}
    apiserver: "true"
spec:
  type: ExternalName
  externalName: "{{ .LocalIp }}"
  ports:
  - port: 443
    protocol: TCP
`
