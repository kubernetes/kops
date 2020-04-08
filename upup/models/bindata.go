// Code generated for package models by go-bindata DO NOT EDIT. (@generated)
// sources:
// upup/models/BUILD.bazel
// upup/models/cloudup/resources/addons/OWNERS
// upup/models/cloudup/resources/addons/authentication.aws/k8s-1.10.yaml.template
// upup/models/cloudup/resources/addons/authentication.aws/k8s-1.12.yaml.template
// upup/models/cloudup/resources/addons/authentication.kope.io/k8s-1.12.yaml
// upup/models/cloudup/resources/addons/authentication.kope.io/k8s-1.8.yaml
// upup/models/cloudup/resources/addons/core.addons.k8s.io/addon.yaml
// upup/models/cloudup/resources/addons/core.addons.k8s.io/k8s-1.12.yaml.template
// upup/models/cloudup/resources/addons/core.addons.k8s.io/k8s-1.7.yaml.template
// upup/models/cloudup/resources/addons/core.addons.k8s.io/v1.4.0.yaml
// upup/models/cloudup/resources/addons/coredns.addons.k8s.io/k8s-1.12.yaml.template
// upup/models/cloudup/resources/addons/coredns.addons.k8s.io/k8s-1.6.yaml.template
// upup/models/cloudup/resources/addons/digitalocean-cloud-controller.addons.k8s.io/k8s-1.8.yaml.template
// upup/models/cloudup/resources/addons/dns-controller.addons.k8s.io/k8s-1.12.yaml.template
// upup/models/cloudup/resources/addons/dns-controller.addons.k8s.io/k8s-1.6.yaml.template
// upup/models/cloudup/resources/addons/external-dns.addons.k8s.io/README.md
// upup/models/cloudup/resources/addons/external-dns.addons.k8s.io/k8s-1.12.yaml.template
// upup/models/cloudup/resources/addons/external-dns.addons.k8s.io/k8s-1.6.yaml.template
// upup/models/cloudup/resources/addons/kops-controller.addons.k8s.io/k8s-1.16.yaml.template
// upup/models/cloudup/resources/addons/kube-dns.addons.k8s.io/k8s-1.12.yaml.template
// upup/models/cloudup/resources/addons/kube-dns.addons.k8s.io/k8s-1.6.yaml.template
// upup/models/cloudup/resources/addons/kubelet-api.rbac.addons.k8s.io/k8s-1.9.yaml
// upup/models/cloudup/resources/addons/limit-range.addons.k8s.io/addon.yaml
// upup/models/cloudup/resources/addons/limit-range.addons.k8s.io/v1.5.0.yaml
// upup/models/cloudup/resources/addons/metadata-proxy.addons.k8s.io/addon.yaml
// upup/models/cloudup/resources/addons/metadata-proxy.addons.k8s.io/v0.1.12.yaml
// upup/models/cloudup/resources/addons/networking.amazon-vpc-routed-eni/k8s-1.10.yaml.template
// upup/models/cloudup/resources/addons/networking.amazon-vpc-routed-eni/k8s-1.12.yaml.template
// upup/models/cloudup/resources/addons/networking.amazon-vpc-routed-eni/k8s-1.16.yaml.template
// upup/models/cloudup/resources/addons/networking.amazon-vpc-routed-eni/k8s-1.8.yaml.template
// upup/models/cloudup/resources/addons/networking.cilium.io/k8s-1.12.yaml.template
// upup/models/cloudup/resources/addons/networking.cilium.io/k8s-1.7.yaml.template
// upup/models/cloudup/resources/addons/networking.flannel/k8s-1.12.yaml.template
// upup/models/cloudup/resources/addons/networking.flannel/k8s-1.6.yaml.template
// upup/models/cloudup/resources/addons/networking.kope.io/k8s-1.12.yaml
// upup/models/cloudup/resources/addons/networking.kope.io/k8s-1.6.yaml
// upup/models/cloudup/resources/addons/networking.kuberouter/k8s-1.12.yaml.template
// upup/models/cloudup/resources/addons/networking.kuberouter/k8s-1.6.yaml.template
// upup/models/cloudup/resources/addons/networking.projectcalico.org/k8s-1.12.yaml.template
// upup/models/cloudup/resources/addons/networking.projectcalico.org/k8s-1.16.yaml.template
// upup/models/cloudup/resources/addons/networking.projectcalico.org/k8s-1.7-v3.yaml.template
// upup/models/cloudup/resources/addons/networking.projectcalico.org/k8s-1.7.yaml.template
// upup/models/cloudup/resources/addons/networking.projectcalico.org.canal/k8s-1.12.yaml.template
// upup/models/cloudup/resources/addons/networking.projectcalico.org.canal/k8s-1.15.yaml.template
// upup/models/cloudup/resources/addons/networking.projectcalico.org.canal/k8s-1.16.yaml.template
// upup/models/cloudup/resources/addons/networking.projectcalico.org.canal/k8s-1.9.yaml.template
// upup/models/cloudup/resources/addons/networking.romana/k8s-1.12.yaml.template
// upup/models/cloudup/resources/addons/networking.romana/k8s-1.7.yaml.template
// upup/models/cloudup/resources/addons/networking.weave/k8s-1.12.yaml.template
// upup/models/cloudup/resources/addons/networking.weave/k8s-1.8.yaml.template
// upup/models/cloudup/resources/addons/node-authorizer.addons.k8s.io/k8s-1.10.yaml.template
// upup/models/cloudup/resources/addons/node-authorizer.addons.k8s.io/k8s-1.12.yaml.template
// upup/models/cloudup/resources/addons/openstack.addons.k8s.io/BUILD.bazel
// upup/models/cloudup/resources/addons/openstack.addons.k8s.io/k8s-1.11.yaml.template
// upup/models/cloudup/resources/addons/openstack.addons.k8s.io/k8s-1.13.yaml.template
// upup/models/cloudup/resources/addons/podsecuritypolicy.addons.k8s.io/k8s-1.10.yaml.template
// upup/models/cloudup/resources/addons/podsecuritypolicy.addons.k8s.io/k8s-1.12.yaml.template
// upup/models/cloudup/resources/addons/podsecuritypolicy.addons.k8s.io/k8s-1.9.yaml.template
// upup/models/cloudup/resources/addons/rbac.addons.k8s.io/k8s-1.8.yaml
// upup/models/cloudup/resources/addons/scheduler.addons.k8s.io/v1.7.0.yaml
// upup/models/cloudup/resources/addons/spotinst-kubernetes-cluster-controller.addons.k8s.io/v1.14.0.yaml.template
// upup/models/cloudup/resources/addons/spotinst-kubernetes-cluster-controller.addons.k8s.io/v1.9.0.yaml.template
// upup/models/cloudup/resources/addons/storage-aws.addons.k8s.io/v1.15.0.yaml
// upup/models/cloudup/resources/addons/storage-aws.addons.k8s.io/v1.7.0.yaml
// upup/models/cloudup/resources/addons/storage-gce.addons.k8s.io/v1.7.0.yaml
// upup/models/nodeup/_automatic_upgrades/_debian_family/files/etc/apt/apt.conf.d/20auto-upgrades
// upup/models/nodeup/_automatic_upgrades/_debian_family/packages/unattended-upgrades
// upup/models/nodeup/resources/_lyft_vpc_cni/files/etc/cni/net.d/10-cni-ipvlan-vpc-k8s.conflist.template
package models

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type asset struct {
	bytes []byte
	info  os.FileInfo
}

type bindataFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
}

// Name return file name
func (fi bindataFileInfo) Name() string {
	return fi.name
}

// Size return file size
func (fi bindataFileInfo) Size() int64 {
	return fi.size
}

// Mode return file mode
func (fi bindataFileInfo) Mode() os.FileMode {
	return fi.mode
}

// Mode return file modify time
func (fi bindataFileInfo) ModTime() time.Time {
	return fi.modTime
}

// IsDir return file whether a directory
func (fi bindataFileInfo) IsDir() bool {
	return fi.mode&os.ModeDir != 0
}

// Sys return file is sys mode
func (fi bindataFileInfo) Sys() interface{} {
	return nil
}

var _buildBazel = []byte(`load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "bindata.go",
        "vfs.go",
    ],
    importpath = "k8s.io/kops/upup/models",
    visibility = ["//visibility:public"],
    deps = ["//util/pkg/vfs:go_default_library"],
)

genrule(
    name = "bindata",
    srcs = glob(
        [
            "cloudup/**",
            "nodeup/**",
        ],
    ),
    outs = ["bindata.go"],
    cmd = """
$(location //vendor/github.com/go-bindata/go-bindata/go-bindata:go-bindata) \
  -o "$(OUTS)" -pkg models \
  -nometadata \
  -nocompress \
  -prefix $$(pwd) \
  -prefix upup/models $(SRCS)
""",
    tools = [
        "//vendor/github.com/go-bindata/go-bindata/go-bindata",
    ],
)
`)

func buildBazelBytes() ([]byte, error) {
	return _buildBazel, nil
}

func buildBazel() (*asset, error) {
	bytes, err := buildBazelBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "BUILD.bazel", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _cloudupResourcesAddonsOwners = []byte(`# See the OWNERS docs at https://go.k8s.io/owners
labels:
- area/addons
`)

func cloudupResourcesAddonsOwnersBytes() ([]byte, error) {
	return _cloudupResourcesAddonsOwners, nil
}

func cloudupResourcesAddonsOwners() (*asset, error) {
	bytes, err := cloudupResourcesAddonsOwnersBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "cloudup/resources/addons/OWNERS", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _cloudupResourcesAddonsAuthenticationAwsK8s110YamlTemplate = []byte(`---
apiVersion: extensions/v1beta1
kind: DaemonSet
metadata:
  namespace: kube-system
  name: aws-iam-authenticator
  labels:
    k8s-app: aws-iam-authenticator
    does-this: "pass ci?"
spec:
  updateStrategy:
    type: RollingUpdate
  template:
    metadata:
      annotations:
        scheduler.alpha.kubernetes.io/critical-pod: ""
      labels:
        k8s-app: aws-iam-authenticator
    spec:
      # run on the host network (don't depend on CNI)
      hostNetwork: true

      # run on each master node
      nodeSelector:
        node-role.kubernetes.io/master: ""
      tolerations:
      - effect: NoSchedule
        key: node-role.kubernetes.io/master
      - key: CriticalAddonsOnly
        operator: Exists

      # run ` + "`" + `aws-iam-authenticator server` + "`" + ` with three volumes
      # - config (mounted from the ConfigMap at /etc/aws-iam-authenticator/config.yaml)
      # - state (persisted TLS certificate and keys, mounted from the host)
      # - output (output kubeconfig to plug into your apiserver configuration, mounted from the host)
      containers:
      - name: aws-iam-authenticator
        image: {{ or .Authentication.Aws.Image "602401143452.dkr.ecr.us-west-2.amazonaws.com/amazon/aws-iam-authenticator:v0.4.0" }}
        args:
        - server
        - --config=/etc/aws-iam-authenticator/config.yaml
        - --state-dir=/var/aws-iam-authenticator
        - --kubeconfig-pregenerated=true

        resources:
          requests:
            memory: {{ or .Authentication.Aws.MemoryRequest "20Mi" }}
            cpu: {{ or .Authentication.Aws.CPURequest "10m" }}
          limits:
            memory: {{ or .Authentication.Aws.MemoryLimit "20Mi" }}
            cpu: {{ or .Authentication.Aws.CPULimit "100m" }}

        volumeMounts:
        - name: config
          mountPath: /etc/aws-iam-authenticator/
        - name: state
          mountPath: /var/aws-iam-authenticator/
        - name: output
          mountPath: /etc/kubernetes/aws-iam-authenticator/

      volumes:
      - name: config
        configMap:
          name: aws-iam-authenticator
      - name: output
        hostPath:
          path: /srv/kubernetes/aws-iam-authenticator/
      - name: state
        hostPath:
          path: /srv/kubernetes/aws-iam-authenticator/
`)

func cloudupResourcesAddonsAuthenticationAwsK8s110YamlTemplateBytes() ([]byte, error) {
	return _cloudupResourcesAddonsAuthenticationAwsK8s110YamlTemplate, nil
}

func cloudupResourcesAddonsAuthenticationAwsK8s110YamlTemplate() (*asset, error) {
	bytes, err := cloudupResourcesAddonsAuthenticationAwsK8s110YamlTemplateBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "cloudup/resources/addons/authentication.aws/k8s-1.10.yaml.template", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _cloudupResourcesAddonsAuthenticationAwsK8s112YamlTemplate = []byte(`---
apiVersion: apps/v1
kind: DaemonSet
metadata:
  namespace: kube-system
  name: aws-iam-authenticator
  labels:
    k8s-app: aws-iam-authenticator
spec:
  updateStrategy:
    type: RollingUpdate
  selector:
    matchLabels:
      k8s-app: aws-iam-authenticator
  template:
    metadata:
      annotations:
        scheduler.alpha.kubernetes.io/critical-pod: ""
      labels:
        k8s-app: aws-iam-authenticator
    spec:
      # run on the host network (don't depend on CNI)
      hostNetwork: true

      # run on each master node
      nodeSelector:
        node-role.kubernetes.io/master: ""
      priorityClassName: system-node-critical
      tolerations:
      - effect: NoSchedule
        key: node-role.kubernetes.io/master
      - key: CriticalAddonsOnly
        operator: Exists

      # run ` + "`" + `aws-iam-authenticator server` + "`" + ` with three volumes
      # - config (mounted from the ConfigMap at /etc/aws-iam-authenticator/config.yaml)
      # - state (persisted TLS certificate and keys, mounted from the host)
      # - output (output kubeconfig to plug into your apiserver configuration, mounted from the host)
      containers:
      - name: aws-iam-authenticator
        image: {{ or .Authentication.Aws.Image "602401143452.dkr.ecr.us-west-2.amazonaws.com/amazon/aws-iam-authenticator:v0.4.0" }}
        args:
        - server
        - --config=/etc/aws-iam-authenticator/config.yaml
        - --state-dir=/var/aws-iam-authenticator
        - --kubeconfig-pregenerated=true

        resources:
          requests:
            memory: {{ or .Authentication.Aws.MemoryRequest "20Mi" }}
            cpu: {{ or .Authentication.Aws.CPURequest "10m" }}
          limits:
            memory: {{ or .Authentication.Aws.MemoryLimit "20Mi" }}
            cpu: {{ or .Authentication.Aws.CPULimit "100m" }}

        volumeMounts:
        - name: config
          mountPath: /etc/aws-iam-authenticator/
        - name: state
          mountPath: /var/aws-iam-authenticator/
        - name: output
          mountPath: /etc/kubernetes/aws-iam-authenticator/

      volumes:
      - name: config
        configMap:
          name: aws-iam-authenticator
      - name: output
        hostPath:
          path: /srv/kubernetes/aws-iam-authenticator/
      - name: state
        hostPath:
          path: /srv/kubernetes/aws-iam-authenticator/
`)

func cloudupResourcesAddonsAuthenticationAwsK8s112YamlTemplateBytes() ([]byte, error) {
	return _cloudupResourcesAddonsAuthenticationAwsK8s112YamlTemplate, nil
}

func cloudupResourcesAddonsAuthenticationAwsK8s112YamlTemplate() (*asset, error) {
	bytes, err := cloudupResourcesAddonsAuthenticationAwsK8s112YamlTemplateBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "cloudup/resources/addons/authentication.aws/k8s-1.12.yaml.template", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _cloudupResourcesAddonsAuthenticationKopeIoK8s112Yaml = []byte(`apiVersion: v1
kind: Namespace
metadata:
  name: kopeio-auth
  labels:
    k8s-addon: authentication.kope.io
    role.kubernetes.io/authentication: "1"

---

apiVersion: v1
kind: Service
metadata:
  name: auth-api
  namespace: kopeio-auth
  labels:
    k8s-addon: authentication.kope.io
    role.kubernetes.io/authentication: "1"
spec:
  selector:
    app: auth-api
  ports:
  - port: 443
    targetPort: 9002

---

apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: auth-api
  namespace: kopeio-auth
  labels:
    k8s-addon: authentication.kope.io
    role.kubernetes.io/authentication: "1"
spec:
  selector:
    matchLabels:
      app: auth-api
  template:
    metadata:
      labels:
        app: auth-api
      annotations:
        scheduler.alpha.kubernetes.io/critical-pod: ''
    spec:
      serviceAccountName: auth-api
      hostNetwork: true
      nodeSelector:
        node-role.kubernetes.io/master: ""
      priorityClassName: system-node-critical
      tolerations:
      - effect: NoSchedule
        key: node-role.kubernetes.io/master
      - key: "CriticalAddonsOnly"
        operator: "Exists"
      containers:
      - name: auth-api
        image: kopeio/auth-api:1.0.20171125
        imagePullPolicy: Always
        ports:
        - containerPort: 9001
        command:
        - /auth-api
        - --listen=127.0.0.1:9001
        - --secure-port=9002
        - --etcd-servers=http://127.0.0.1:4001
        - --v=8
        - --storage-backend=etcd2

---

apiVersion: apiregistration.k8s.io/v1
kind: APIService
metadata:
  name: v1alpha1.auth.kope.io
  labels:
    k8s-addon: authentication.kope.io
    role.kubernetes.io/authentication: "1"
spec:
  insecureSkipTLSVerify: true
  group: auth.kope.io
  groupPriorityMinimum: 1000
  versionPriority: 15
  service:
    name: auth-api
    namespace: kopeio-auth
  version: v1alpha1

---

apiVersion: apiregistration.k8s.io/v1
kind: APIService
metadata:
  name: v1alpha1.config.auth.kope.io
  labels:
    k8s-addon: authentication.kope.io
    role.kubernetes.io/authentication: "1"
spec:
  insecureSkipTLSVerify: true
  group: config.auth.kope.io
  groupPriorityMinimum: 1000
  versionPriority: 15
  service:
    name: auth-api
    namespace: kopeio-auth
  version: v1alpha1

---

kind: ServiceAccount
apiVersion: v1
metadata:
  name: auth-api
  namespace: kopeio-auth
  labels:
    k8s-addon: authentication.kope.io
    role.kubernetes.io/authentication: "1"

---

apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: kopeio-auth:auth-api:auth-reader
  namespace: kube-system
  labels:
    k8s-addon: authentication.kope.io
    role.kubernetes.io/authentication: "1"
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: extension-apiserver-authentication-reader
subjects:
- kind: ServiceAccount
  name: auth-api
  namespace: kopeio-auth

---

apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: kopeio-auth:system:auth-delegator
  labels:
    k8s-addon: authentication.kope.io
    role.kubernetes.io/authentication: "1"
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: system:auth-delegator
subjects:
- kind: ServiceAccount
  name: auth-api
  namespace: kopeio-auth

---

kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: auth-api
  namespace: kopeio-auth
  labels:
    k8s-addon: authentication.kope.io
    role.kubernetes.io/authentication: "1"
rules:
- apiGroups: ["auth.kope.io"]
  resources: ["users"]
  verbs: ["get", "list", "watch"]

---

apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: auth-api
  namespace: kopeio-auth
  labels:
    k8s-addon: authentication.kope.io
    role.kubernetes.io/authentication: "1"
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: auth-api
subjects:
- kind: ServiceAccount
  name: auth-api
  namespace: kopeio-auth
`)

func cloudupResourcesAddonsAuthenticationKopeIoK8s112YamlBytes() ([]byte, error) {
	return _cloudupResourcesAddonsAuthenticationKopeIoK8s112Yaml, nil
}

func cloudupResourcesAddonsAuthenticationKopeIoK8s112Yaml() (*asset, error) {
	bytes, err := cloudupResourcesAddonsAuthenticationKopeIoK8s112YamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "cloudup/resources/addons/authentication.kope.io/k8s-1.12.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _cloudupResourcesAddonsAuthenticationKopeIoK8s18Yaml = []byte(`apiVersion: v1
kind: Namespace
metadata:
  name: kopeio-auth
  labels:
    k8s-addon: authentication.kope.io
    role.kubernetes.io/authentication: "1"

---

apiVersion: v1
kind: Service
metadata:
  name: auth-api
  namespace: kopeio-auth
  labels:
    k8s-addon: authentication.kope.io
    role.kubernetes.io/authentication: "1"
spec:
  selector:
    app: auth-api
  ports:
  - port: 443
    targetPort: 9002

---

apiVersion: extensions/v1beta1
kind: DaemonSet
metadata:
  name: auth-api
  namespace: kopeio-auth
  labels:
    k8s-addon: authentication.kope.io
    role.kubernetes.io/authentication: "1"
spec:
  template:
    metadata:
      labels:
        app: auth-api
      annotations:
        scheduler.alpha.kubernetes.io/critical-pod: ''
    spec:
      serviceAccountName: auth-api
      hostNetwork: true
      nodeSelector:
        node-role.kubernetes.io/master: ""
      tolerations:
      - effect: NoSchedule
        key: node-role.kubernetes.io/master
      - key: "CriticalAddonsOnly"
        operator: "Exists"
      containers:
      - name: auth-api
        image: kopeio/auth-api:1.0.20171125
        imagePullPolicy: Always
        ports:
        - containerPort: 9001
        command:
        - /auth-api
        - --listen=127.0.0.1:9001
        - --secure-port=9002
        - --etcd-servers=http://127.0.0.1:4001
        - --v=8
        - --storage-backend=etcd2

---

apiVersion: apiregistration.k8s.io/v1beta1
kind: APIService
metadata:
  name: v1alpha1.auth.kope.io
  labels:
    k8s-addon: authentication.kope.io
    role.kubernetes.io/authentication: "1"
spec:
  insecureSkipTLSVerify: true
  group: auth.kope.io
  groupPriorityMinimum: 1000
  versionPriority: 15
  service:
    name: auth-api
    namespace: kopeio-auth
  version: v1alpha1

---

apiVersion: apiregistration.k8s.io/v1beta1
kind: APIService
metadata:
  name: v1alpha1.config.auth.kope.io
  labels:
    k8s-addon: authentication.kope.io
    role.kubernetes.io/authentication: "1"
spec:
  insecureSkipTLSVerify: true
  group: config.auth.kope.io
  groupPriorityMinimum: 1000
  versionPriority: 15
  service:
    name: auth-api
    namespace: kopeio-auth
  version: v1alpha1

---

kind: ServiceAccount
apiVersion: v1
metadata:
  name: auth-api
  namespace: kopeio-auth
  labels:
    k8s-addon: authentication.kope.io
    role.kubernetes.io/authentication: "1"

---

apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: kopeio-auth:auth-api:auth-reader
  namespace: kube-system
  labels:
    k8s-addon: authentication.kope.io
    role.kubernetes.io/authentication: "1"
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: extension-apiserver-authentication-reader
subjects:
- kind: ServiceAccount
  name: auth-api
  namespace: kopeio-auth

---

apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: kopeio-auth:system:auth-delegator
  labels:
    k8s-addon: authentication.kope.io
    role.kubernetes.io/authentication: "1"
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: system:auth-delegator
subjects:
- kind: ServiceAccount
  name: auth-api
  namespace: kopeio-auth

---

kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: auth-api
  namespace: kopeio-auth
  labels:
    k8s-addon: authentication.kope.io
    role.kubernetes.io/authentication: "1"
rules:
- apiGroups: ["auth.kope.io"]
  resources: ["users"]
  verbs: ["get", "list", "watch"]

---

apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: auth-api
  namespace: kopeio-auth
  labels:
    k8s-addon: authentication.kope.io
    role.kubernetes.io/authentication: "1"
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: auth-api
subjects:
- kind: ServiceAccount
  name: auth-api
  namespace: kopeio-auth
`)

func cloudupResourcesAddonsAuthenticationKopeIoK8s18YamlBytes() ([]byte, error) {
	return _cloudupResourcesAddonsAuthenticationKopeIoK8s18Yaml, nil
}

func cloudupResourcesAddonsAuthenticationKopeIoK8s18Yaml() (*asset, error) {
	bytes, err := cloudupResourcesAddonsAuthenticationKopeIoK8s18YamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "cloudup/resources/addons/authentication.kope.io/k8s-1.8.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _cloudupResourcesAddonsCoreAddonsK8sIoAddonYaml = []byte(`kind: Addons
metadata:
  name: core
spec:
  addons:
  - version: 1.4.0
    selector:
      k8s-addon: core.addons.k8s.io
    manifest: v1.4.0.yaml

`)

func cloudupResourcesAddonsCoreAddonsK8sIoAddonYamlBytes() ([]byte, error) {
	return _cloudupResourcesAddonsCoreAddonsK8sIoAddonYaml, nil
}

func cloudupResourcesAddonsCoreAddonsK8sIoAddonYaml() (*asset, error) {
	bytes, err := cloudupResourcesAddonsCoreAddonsK8sIoAddonYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "cloudup/resources/addons/core.addons.k8s.io/addon.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _cloudupResourcesAddonsCoreAddonsK8sIoK8s112YamlTemplate = []byte(`apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  annotations:
    rbac.authorization.kubernetes.io/autoupdate: "true"
  labels:
    kubernetes.io/bootstrapping: rbac-defaults
  name: system:cloud-controller-manager
rules:
- apiGroups:
  - ""
  resources:
  - events
  verbs:
  - create
  - patch
  - update
- apiGroups:
  - ""
  resources:
  - nodes
  verbs:
  - '*'
- apiGroups:
  - ""
  resources:
  - services
  verbs:
  - list
  - patch
  - update
  - watch
- apiGroups:
  - ""
  resources:
  - events
  verbs:
  - create
  - patch
  - update
- apiGroups:
  - ""
  resources:
  - endpoints
  verbs:
  - create
  - get
  - list
  - update
  - watch
- apiGroups:
  - ""
  resources:
  - serviceaccounts
  verbs:
  - create
- apiGroups:
  - ""
  resources:
  - persistentvolumes
  verbs:
  - get
  - list
  - update
  - watch
- apiGroups:
  - ""
  resources:
  - secrets
  verbs:
  - list

---

apiVersion: v1
kind: ServiceAccount
metadata:
  name: cloud-controller-manager
  namespace: kube-system

---

kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: system:cloud-controller-manager
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: system:cloud-controller-manager
subjects:
- kind: ServiceAccount
  name: cloud-controller-manager
  namespace: kube-system

---

apiVersion: apps/v1
kind: DaemonSet
metadata:
  labels:
    k8s-app: cloud-controller-manager
  name: cloud-controller-manager
  namespace: kube-system
spec:
  selector:
    matchLabels:
      k8s-app: cloud-controller-manager
  template:
    metadata:
      labels:
        k8s-app: cloud-controller-manager
    spec:
      nodeSelector:
        node-role.kubernetes.io/master: ""
      priorityClassName: system-node-critical
      serviceAccountName: cloud-controller-manager
      containers:
      - name: cloud-controller-manager
        # for in-tree providers we use k8s.gcr.io/cloud-controller-manager
        # this can be replaced with any other image for out-of-tree providers
        image: k8s.gcr.io/cloud-controller-manager:v{{ .KubernetesVersion }}  # Reviewers: Will this work?
        command:
        - /usr/local/bin/cloud-controller-manager
        - --cloud-provider={{ .CloudProvider }}
        - --leader-elect=true
        - --use-service-account-credentials
        # these flags will vary for every cloud provider
        - --allocate-node-cidrs=true
        - --configure-cloud-routes=true
        - --cluster-cidr={{ .KubeControllerManager.ClusterCIDR }}
        volumeMounts:
        - name: ca-certificates
          mountPath: /etc/ssl/certs
      hostNetwork: true
      dnsPolicy: Default
      volumes:
      - name: ca-certificates
        hostPath:
          path: /etc/ssl/certs
      tolerations:
      # this is required so CCM can bootstrap itself
      - key: node.cloudprovider.kubernetes.io/uninitialized
        value: "true"
        effect: NoSchedule
      # this is to have the daemonset runnable on master nodes
      # the taint may vary depending on your cluster setup
      - key: node-role.kubernetes.io/master
        effect: NoSchedule
      # this is to restrict CCM to only run on master nodes
      # the node selector may vary depending on your cluster setup
      - key: "CriticalAddonsOnly"
        operator: "Exists"

`)

func cloudupResourcesAddonsCoreAddonsK8sIoK8s112YamlTemplateBytes() ([]byte, error) {
	return _cloudupResourcesAddonsCoreAddonsK8sIoK8s112YamlTemplate, nil
}

func cloudupResourcesAddonsCoreAddonsK8sIoK8s112YamlTemplate() (*asset, error) {
	bytes, err := cloudupResourcesAddonsCoreAddonsK8sIoK8s112YamlTemplateBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "cloudup/resources/addons/core.addons.k8s.io/k8s-1.12.yaml.template", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _cloudupResourcesAddonsCoreAddonsK8sIoK8s17YamlTemplate = []byte(`apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  annotations:
    rbac.authorization.kubernetes.io/autoupdate: "true"
  labels:
    kubernetes.io/bootstrapping: rbac-defaults
  name: system:cloud-controller-manager
rules:
- apiGroups:
  - ""
  resources:
  - events
  verbs:
  - create
  - patch
  - update
- apiGroups:
  - ""
  resources:
  - nodes
  verbs:
  - '*'
- apiGroups:
  - ""
  resources:
  - services
  verbs:
  - list
  - patch
  - update
  - watch
- apiGroups:
  - ""
  resources:
  - events
  verbs:
  - create
  - patch
  - update
- apiGroups:
  - ""
  resources:
  - endpoints
  verbs:
  - create
  - get
  - list
  - update
  - watch
- apiGroups:
  - ""
  resources:
  - serviceaccounts
  verbs:
  - create
- apiGroups:
  - ""
  resources:
  - persistentvolumes
  verbs:
  - get
  - list
  - update
  - watch
- apiGroups:
  - ""
  resources:
  - secrets
  verbs:
  - list

---

apiVersion: v1
kind: ServiceAccount
metadata:
  name: cloud-controller-manager
  namespace: kube-system

---

kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: system:cloud-controller-manager
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: system:cloud-controller-manager
subjects:
- kind: ServiceAccount
  name: cloud-controller-manager
  namespace: kube-system

---

apiVersion: extensions/v1beta1
kind: DaemonSet
metadata:
  labels:
    k8s-app: cloud-controller-manager
  name: cloud-controller-manager
  namespace: kube-system
spec:
  selector:
    matchLabels:
      k8s-app: cloud-controller-manager
  template:
    metadata:
      labels:
        k8s-app: cloud-controller-manager
    spec:
      nodeSelector:
        node-role.kubernetes.io/master: ""
      serviceAccountName: cloud-controller-manager
      containers:
      - name: cloud-controller-manager
        # for in-tree providers we use k8s.gcr.io/cloud-controller-manager
        # this can be replaced with any other image for out-of-tree providers
        image: k8s.gcr.io/cloud-controller-manager:v{{ .KubernetesVersion }}  # Reviewers: Will this work?
        command:
        - /usr/local/bin/cloud-controller-manager
        - --cloud-provider={{ .CloudProvider }}
        - --leader-elect=true
        - --use-service-account-credentials
        # these flags will vary for every cloud provider
        - --allocate-node-cidrs=true
        - --configure-cloud-routes=true
        - --cluster-cidr={{ .KubeControllerManager.ClusterCIDR }}
        volumeMounts:
        - name: ca-certificates
          mountPath: /etc/ssl/certs
      hostNetwork: true
      dnsPolicy: Default
      volumes:
      - name: ca-certificates
        hostPath:
          path: /etc/ssl/certs
      tolerations:
      # this is required so CCM can bootstrap itself
      - key: node.cloudprovider.kubernetes.io/uninitialized
        value: "true"
        effect: NoSchedule
      # this is to have the daemonset runnable on master nodes
      # the taint may vary depending on your cluster setup
      - key: node-role.kubernetes.io/master
        effect: NoSchedule
      # this is to restrict CCM to only run on master nodes
      # the node selector may vary depending on your cluster setup
      - key: "CriticalAddonsOnly"
        operator: "Exists"

`)

func cloudupResourcesAddonsCoreAddonsK8sIoK8s17YamlTemplateBytes() ([]byte, error) {
	return _cloudupResourcesAddonsCoreAddonsK8sIoK8s17YamlTemplate, nil
}

func cloudupResourcesAddonsCoreAddonsK8sIoK8s17YamlTemplate() (*asset, error) {
	bytes, err := cloudupResourcesAddonsCoreAddonsK8sIoK8s17YamlTemplateBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "cloudup/resources/addons/core.addons.k8s.io/k8s-1.7.yaml.template", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _cloudupResourcesAddonsCoreAddonsK8sIoV140Yaml = []byte(`---
apiVersion: v1
kind: Namespace
metadata:
  name: kube-system
`)

func cloudupResourcesAddonsCoreAddonsK8sIoV140YamlBytes() ([]byte, error) {
	return _cloudupResourcesAddonsCoreAddonsK8sIoV140Yaml, nil
}

func cloudupResourcesAddonsCoreAddonsK8sIoV140Yaml() (*asset, error) {
	bytes, err := cloudupResourcesAddonsCoreAddonsK8sIoV140YamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "cloudup/resources/addons/core.addons.k8s.io/v1.4.0.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _cloudupResourcesAddonsCorednsAddonsK8sIoK8s112YamlTemplate = []byte(`apiVersion: v1
kind: ServiceAccount
metadata:
  name: coredns
  namespace: kube-system
  labels:
      kubernetes.io/cluster-service: "true"
      k8s-addon: coredns.addons.k8s.io
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  labels:
    kubernetes.io/bootstrapping: rbac-defaults
    k8s-addon: coredns.addons.k8s.io
  name: system:coredns
rules:
- apiGroups:
  - ""
  resources:
  - endpoints
  - services
  - pods
  - namespaces
  verbs:
  - list
  - watch
- apiGroups:
  - ""
  resources:
  - nodes
  verbs:
  - get
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  annotations:
    rbac.authorization.kubernetes.io/autoupdate: "true"
  labels:
    kubernetes.io/bootstrapping: rbac-defaults
    k8s-addon: coredns.addons.k8s.io
  name: system:coredns
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: system:coredns
subjects:
- kind: ServiceAccount
  name: coredns
  namespace: kube-system
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: coredns
  namespace: kube-system
  labels:
      addonmanager.kubernetes.io/mode: EnsureExists
data:
  Corefile: |
  {{- if KubeDNS.ExternalCoreFile }}
{{ KubeDNS.ExternalCoreFile | indent 4 }}
  {{- else }}
    .:53 {
        errors
        health {
          lameduck 5s
        }
        kubernetes {{ KubeDNS.Domain }}. in-addr.arpa ip6.arpa {
          pods insecure
          upstream
          fallthrough in-addr.arpa ip6.arpa
        }
        prometheus :9153
        forward . /etc/resolv.conf
        loop
        cache 30
        loadbalance
        reload
    }
  {{- end }}
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: coredns-autoscaler
  namespace: kube-system
  labels:
    k8s-addon: coredns.addons.k8s.io
    k8s-app: coredns-autoscaler
    kubernetes.io/cluster-service: "true"
spec:
  selector:
    matchLabels:
      k8s-app: coredns-autoscaler
  template:
    metadata:
      labels:
        k8s-app: coredns-autoscaler
      annotations:
        scheduler.alpha.kubernetes.io/critical-pod: ''
    spec:
      containers:
      - name: autoscaler
        image: k8s.gcr.io/cluster-proportional-autoscaler-{{Arch}}:1.4.0
        resources:
            requests:
                cpu: "20m"
                memory: "10Mi"
        command:
          - /cluster-proportional-autoscaler
          - --namespace=kube-system
          - --configmap=coredns-autoscaler
          - --target=Deployment/coredns
          # When cluster is using large nodes(with more cores), "coresPerReplica" should dominate.
          # If using small nodes, "nodesPerReplica" should dominate.
          - --default-params={"linear":{"coresPerReplica":256,"nodesPerReplica":16,"preventSinglePointFailure":true}}
          - --logtostderr=true
          - --v=2
      priorityClassName: system-cluster-critical
      tolerations:
      - key: "CriticalAddonsOnly"
        operator: "Exists"
      serviceAccountName: coredns-autoscaler
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: coredns
  namespace: kube-system
  labels:
    k8s-app: kube-dns
    k8s-addon: coredns.addons.k8s.io
    kubernetes.io/cluster-service: "true"
spec:
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxUnavailable: 1
      maxSurge: 10%
  selector:
    matchLabels:
      k8s-app: kube-dns
  template:
    metadata:
      labels:
        k8s-app: kube-dns
    spec:
      affinity:
        podAntiAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
            - weight: 1
              podAffinityTerm:
                labelSelector:
                  matchExpressions:
                    - key: k8s-app
                      operator: In
                      values:
                        - kube-dns
                topologyKey: kubernetes.io/hostname
      priorityClassName: system-cluster-critical
      serviceAccountName: coredns
      tolerations:
        - key: "CriticalAddonsOnly"
          operator: "Exists"
      nodeSelector:
          beta.kubernetes.io/os: linux
      containers:
      - name: coredns
        image: {{ if KubeDNS.CoreDNSImage }}{{ KubeDNS.CoreDNSImage }}{{ else }}k8s.gcr.io/coredns:1.6.7{{ end }}
        imagePullPolicy: IfNotPresent
        resources:
          limits:
            memory: {{ KubeDNS.MemoryLimit }}
          requests:
            cpu: {{ KubeDNS.CPURequest }}
            memory: {{ KubeDNS.MemoryRequest }}
        args: [ "-conf", "/etc/coredns/Corefile" ]
        volumeMounts:
        - name: config-volume
          mountPath: /etc/coredns
          readOnly: true
        ports:
        - containerPort: 53
          name: dns
          protocol: UDP
        - containerPort: 53
          name: dns-tcp
          protocol: TCP
        - containerPort: 9153
          name: metrics
          protocol: TCP
        securityContext:
          allowPrivilegeEscalation: false
          capabilities:
            add:
            - NET_BIND_SERVICE
            drop:
            - all
          readOnlyRootFilesystem: true
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
            scheme: HTTP
          initialDelaySeconds: 60
          timeoutSeconds: 5
          successThreshold: 1
          failureThreshold: 5
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
            scheme: HTTP
      dnsPolicy: Default
      volumes:
        - name: config-volume
          configMap:
            name: coredns
            items:
            - key: Corefile
              path: Corefile
---
apiVersion: v1
kind: Service
metadata:
  name: kube-dns
  namespace: kube-system
  annotations:
    prometheus.io/port: "9153"
    prometheus.io/scrape: "true"
  labels:
    k8s-addon: coredns.addons.k8s.io
    k8s-app: kube-dns
    kubernetes.io/cluster-service: "true"
    kubernetes.io/name: "CoreDNS"
  # Without this resourceVersion value, an update of the Service between versions will yield:
  #   Service "kube-dns" is invalid: metadata.resourceVersion: Invalid value: "": must be specified for an update
  resourceVersion: "0"
spec:
  selector:
    k8s-app: kube-dns
  clusterIP: {{ KubeDNS.ServerIP }}
  ports:
  - name: dns
    port: 53
    protocol: UDP
  - name: dns-tcp
    port: 53
    protocol: TCP
  - name: metrics
    port: 9153
    protocol: TCP

---


apiVersion: v1
kind: ServiceAccount
metadata:
  name: coredns-autoscaler
  namespace: kube-system
  labels:
    k8s-addon: coredns.addons.k8s.io

---

apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  labels:
    k8s-addon: coredns.addons.k8s.io
  name: coredns-autoscaler
rules:
  - apiGroups: [""]
    resources: ["nodes"]
    verbs: ["list","watch"]
  - apiGroups: [""]
    resources: ["replicationcontrollers/scale"]
    verbs: ["get", "update"]
  - apiGroups: ["extensions", "apps"]
    resources: ["deployments/scale", "replicasets/scale"]
    verbs: ["get", "update"]
# Remove the configmaps rule once below issue is fixed:
# kubernetes-incubator/cluster-proportional-autoscaler#16
  - apiGroups: [""]
    resources: ["configmaps"]
    verbs: ["get", "create"]

---

apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  labels:
    k8s-addon: coredns.addons.k8s.io
  name: coredns-autoscaler
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: coredns-autoscaler
subjects:
- kind: ServiceAccount
  name: coredns-autoscaler
  namespace: kube-system

---

apiVersion: policy/v1beta1
kind: PodDisruptionBudget
metadata:
  name: kube-dns
  namespace: kube-system
spec:
  selector:
    matchLabels:
      k8s-app: kube-dns
  minAvailable: 1

`)

func cloudupResourcesAddonsCorednsAddonsK8sIoK8s112YamlTemplateBytes() ([]byte, error) {
	return _cloudupResourcesAddonsCorednsAddonsK8sIoK8s112YamlTemplate, nil
}

func cloudupResourcesAddonsCorednsAddonsK8sIoK8s112YamlTemplate() (*asset, error) {
	bytes, err := cloudupResourcesAddonsCorednsAddonsK8sIoK8s112YamlTemplateBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "cloudup/resources/addons/coredns.addons.k8s.io/k8s-1.12.yaml.template", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _cloudupResourcesAddonsCorednsAddonsK8sIoK8s16YamlTemplate = []byte(`apiVersion: v1
kind: ServiceAccount
metadata:
  name: coredns
  namespace: kube-system
  labels:
      kubernetes.io/cluster-service: "true"
      k8s-addon: coredns.addons.k8s.io
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  labels:
    kubernetes.io/bootstrapping: rbac-defaults
    k8s-addon: coredns.addons.k8s.io
  name: system:coredns
rules:
- apiGroups:
  - ""
  resources:
  - endpoints
  - services
  - pods
  - namespaces
  verbs:
  - list
  - watch
- apiGroups:
  - ""
  resources:
  - nodes
  verbs:
  - get
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  annotations:
    rbac.authorization.kubernetes.io/autoupdate: "true"
  labels:
    kubernetes.io/bootstrapping: rbac-defaults
    k8s-addon: coredns.addons.k8s.io
  name: system:coredns
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: system:coredns
subjects:
- kind: ServiceAccount
  name: coredns
  namespace: kube-system
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: coredns
  namespace: kube-system
  labels:
      addonmanager.kubernetes.io/mode: EnsureExists
data:
  Corefile: |
  {{- if KubeDNS.ExternalCoreFile }}
{{ KubeDNS.ExternalCoreFile | indent 4 }}
  {{- else }}
    .:53 {
        errors
        health {
          lameduck 5s
        }
        kubernetes {{ KubeDNS.Domain }}. in-addr.arpa ip6.arpa {
          pods insecure
          upstream
          fallthrough in-addr.arpa ip6.arpa
        }
        prometheus :9153
        forward . /etc/resolv.conf
        loop
        cache 30
        loadbalance
        reload
    }
  {{- end }}
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: coredns
  namespace: kube-system
  labels:
    k8s-app: kube-dns
    k8s-addon: coredns.addons.k8s.io
    kubernetes.io/cluster-service: "true"
spec:
  replicas: 2
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxUnavailable: 1
  selector:
    matchLabels:
      k8s-app: kube-dns
  template:
    metadata:
      labels:
        k8s-app: kube-dns
    spec:
      priorityClassName: system-cluster-critical
      serviceAccountName: coredns
      tolerations:
        - key: "CriticalAddonsOnly"
          operator: "Exists"
      nodeSelector:
          beta.kubernetes.io/os: linux
      containers:
      - name: coredns
        image: {{ if KubeDNS.CoreDNSImage }}{{ KubeDNS.CoreDNSImage }}{{ else }}k8s.gcr.io/coredns:1.6.7{{ end }}
        imagePullPolicy: IfNotPresent
        resources:
          limits:
            memory: {{ KubeDNS.MemoryLimit }}
          requests:
            cpu: {{ KubeDNS.CPURequest }}
            memory: {{ KubeDNS.MemoryRequest }}
        args: [ "-conf", "/etc/coredns/Corefile" ]
        volumeMounts:
        - name: config-volume
          mountPath: /etc/coredns
          readOnly: true
        ports:
        - containerPort: 53
          name: dns
          protocol: UDP
        - containerPort: 53
          name: dns-tcp
          protocol: TCP
        - containerPort: 9153
          name: metrics
          protocol: TCP
        securityContext:
          allowPrivilegeEscalation: false
          capabilities:
            add:
            - NET_BIND_SERVICE
            drop:
            - all
          readOnlyRootFilesystem: true
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
            scheme: HTTP
          initialDelaySeconds: 60
          timeoutSeconds: 5
          successThreshold: 1
          failureThreshold: 5
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
            scheme: HTTP
      dnsPolicy: Default
      volumes:
        - name: config-volume
          configMap:
            name: coredns
            items:
            - key: Corefile
              path: Corefile
---
apiVersion: v1
kind: Service
metadata:
  name: kube-dns
  namespace: kube-system
  annotations:
    prometheus.io/port: "9153"
    prometheus.io/scrape: "true"
  labels:
    k8s-addon: coredns.addons.k8s.io
    k8s-app: kube-dns
    kubernetes.io/cluster-service: "true"
    kubernetes.io/name: "CoreDNS"
spec:
  selector:
    k8s-app: kube-dns
  clusterIP: {{ KubeDNS.ServerIP }}
  ports:
  - name: dns
    port: 53
    protocol: UDP
  - name: dns-tcp
    port: 53
    protocol: TCP
  - name: metrics
    port: 9153
    protocol: TCP
`)

func cloudupResourcesAddonsCorednsAddonsK8sIoK8s16YamlTemplateBytes() ([]byte, error) {
	return _cloudupResourcesAddonsCorednsAddonsK8sIoK8s16YamlTemplate, nil
}

func cloudupResourcesAddonsCorednsAddonsK8sIoK8s16YamlTemplate() (*asset, error) {
	bytes, err := cloudupResourcesAddonsCorednsAddonsK8sIoK8s16YamlTemplateBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "cloudup/resources/addons/coredns.addons.k8s.io/k8s-1.6.yaml.template", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _cloudupResourcesAddonsDigitaloceanCloudControllerAddonsK8sIoK8s18YamlTemplate = []byte(`---
apiVersion: v1
kind: Secret
metadata:
  name: digitalocean
  namespace: kube-system
stringData:
  # insert your DO access token here
  access-token: {{ DO_TOKEN }}

---
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: digitalocean-cloud-controller-manager
  namespace: kube-system
spec:
  selector:
    matchLabels:
      k8s-app: digitalocean-cloud-controller-manager
  template:
    metadata:
      labels:
        k8s-app: digitalocean-cloud-controller-manager
      annotations:
        scheduler.alpha.kubernetes.io/critical-pod: ''
    spec:
      nodeSelector:
        node-role.kubernetes.io/master: ""
      serviceAccountName: cloud-controller-manager
      dnsPolicy: Default
      hostNetwork: true
      priorityClassName: system-node-critical
      tolerations:
        - key: "node.cloudprovider.kubernetes.io/uninitialized"
          value: "true"
          effect: "NoSchedule"
        - key: "CriticalAddonsOnly"
          operator: "Exists"
        - key: "node-role.kubernetes.io/master"
          effect: NoSchedule
        - effect: NoExecute
          key: node.kubernetes.io/not-ready
          operator: Exists
          tolerationSeconds: 300
        - effect: NoExecute
          key: node.kubernetes.io/unreachable
          operator: Exists
          tolerationSeconds: 300
      containers:
      - image: digitalocean/digitalocean-cloud-controller-manager:v0.1.20
        name: digitalocean-cloud-controller-manager
        command:
          - "/bin/digitalocean-cloud-controller-manager"
          - "--leader-elect=true"
        resources:
          requests:
            cpu: 100m
            memory: 50Mi
        env:
          - name: KUBERNETES_SERVICE_HOST
            value: "127.0.0.1"
          - name: KUBERNETES_SERVICE_PORT
            value: "443"
          - name: DO_ACCESS_TOKEN
            valueFrom:
              secretKeyRef:
                name: digitalocean
                key: access-token

---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: cloud-controller-manager
  namespace: kube-system
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  annotations:
    rbac.authorization.kubernetes.io/autoupdate: "true"
  name: system:cloud-controller-manager
rules:
- apiGroups:
  - ""
  resources:
  - events
  verbs:
  - create
  - patch
  - update
- apiGroups:
  - ""
  resources:
  - nodes
  verbs:
  - '*'
- apiGroups:
  - ""
  resources:
  - nodes/status
  verbs:
  - patch
- apiGroups:
  - ""
  resources:
  - services
  verbs:
  - list
  - patch
  - update
  - watch
- apiGroups:
  - ""
  resources:
  - services/status
  verbs:
  - list
  - patch
  - update
  - watch
- apiGroups:
  - ""
  resources:
  - serviceaccounts
  verbs:
  - create
- apiGroups:
  - ""
  resources:
  - persistentvolumes
  verbs:
  - get
  - list
  - update
  - watch
- apiGroups:
  - ""
  resources:
  - endpoints
  verbs:
  - create
  - get
  - list
  - watch
  - update
---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: system:cloud-controller-manager
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: system:cloud-controller-manager
subjects:
- kind: ServiceAccount
  name: cloud-controller-manager
  namespace: kube-system
`)

func cloudupResourcesAddonsDigitaloceanCloudControllerAddonsK8sIoK8s18YamlTemplateBytes() ([]byte, error) {
	return _cloudupResourcesAddonsDigitaloceanCloudControllerAddonsK8sIoK8s18YamlTemplate, nil
}

func cloudupResourcesAddonsDigitaloceanCloudControllerAddonsK8sIoK8s18YamlTemplate() (*asset, error) {
	bytes, err := cloudupResourcesAddonsDigitaloceanCloudControllerAddonsK8sIoK8s18YamlTemplateBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "cloudup/resources/addons/digitalocean-cloud-controller.addons.k8s.io/k8s-1.8.yaml.template", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _cloudupResourcesAddonsDnsControllerAddonsK8sIoK8s112YamlTemplate = []byte(`kind: Deployment
apiVersion: apps/v1
metadata:
  name: dns-controller
  namespace: kube-system
  labels:
    k8s-addon: dns-controller.addons.k8s.io
    k8s-app: dns-controller
    version: v1.18.0-alpha.2
spec:
  replicas: 1
  selector:
    matchLabels:
      k8s-app: dns-controller
  template:
    metadata:
      labels:
        k8s-addon: dns-controller.addons.k8s.io
        k8s-app: dns-controller
        version: v1.18.0-alpha.2
      annotations:
        scheduler.alpha.kubernetes.io/critical-pod: ''
    spec:
      priorityClassName: system-cluster-critical
      tolerations:
      - operator: Exists
      nodeSelector:
        node-role.kubernetes.io/master: ""
      dnsPolicy: Default  # Don't use cluster DNS (we are likely running before kube-dns)
      hostNetwork: true
      serviceAccount: dns-controller
      containers:
      - name: dns-controller
        image: kope/dns-controller:1.18.0-alpha.2
        command:
{{ range $arg := DnsControllerArgv }}
        - "{{ $arg }}"
{{ end }}
        env:
        - name: KUBERNETES_SERVICE_HOST
          value: "127.0.0.1"
{{- if .EgressProxy }}
{{ range $name, $value := ProxyEnv }}
        - name: {{ $name }}
          value: {{ $value }}
{{ end }}
{{- end }}
{{- if eq .CloudProvider "digitalocean" }}
        - name: DIGITALOCEAN_ACCESS_TOKEN
          valueFrom:
            secretKeyRef:
              name: digitalocean
              key: access-token
{{- end }}
        resources:
          requests:
            cpu: 50m
            memory: 50Mi
        securityContext:
          runAsNonRoot: true

---

apiVersion: v1
kind: ServiceAccount
metadata:
  name: dns-controller
  namespace: kube-system
  labels:
    k8s-addon: dns-controller.addons.k8s.io

---

apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  labels:
    k8s-addon: dns-controller.addons.k8s.io
  name: kops:dns-controller
rules:
- apiGroups:
  - ""
  resources:
  - endpoints
  - services
  - pods
  - ingress
  - nodes
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - "extensions"
  resources:
  - ingresses
  verbs:
  - get
  - list
  - watch

---

apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  labels:
    k8s-addon: dns-controller.addons.k8s.io
  name: kops:dns-controller
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: kops:dns-controller
subjects:
- apiGroup: rbac.authorization.k8s.io
  kind: User
  name: system:serviceaccount:kube-system:dns-controller
`)

func cloudupResourcesAddonsDnsControllerAddonsK8sIoK8s112YamlTemplateBytes() ([]byte, error) {
	return _cloudupResourcesAddonsDnsControllerAddonsK8sIoK8s112YamlTemplate, nil
}

func cloudupResourcesAddonsDnsControllerAddonsK8sIoK8s112YamlTemplate() (*asset, error) {
	bytes, err := cloudupResourcesAddonsDnsControllerAddonsK8sIoK8s112YamlTemplateBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "cloudup/resources/addons/dns-controller.addons.k8s.io/k8s-1.12.yaml.template", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _cloudupResourcesAddonsDnsControllerAddonsK8sIoK8s16YamlTemplate = []byte(`kind: Deployment
apiVersion: extensions/v1beta1
metadata:
  name: dns-controller
  namespace: kube-system
  labels:
    k8s-addon: dns-controller.addons.k8s.io
    k8s-app: dns-controller
    version: v1.18.0-alpha.2
spec:
  replicas: 1
  selector:
    matchLabels:
      k8s-app: dns-controller
  template:
    metadata:
      labels:
        k8s-addon: dns-controller.addons.k8s.io
        k8s-app: dns-controller
        version: v1.18.0-alpha.2
      annotations:
        scheduler.alpha.kubernetes.io/critical-pod: ''
        # For 1.6, we keep the old tolerations in case of a downgrade to 1.5
        scheduler.alpha.kubernetes.io/tolerations: '[{"key": "dedicated", "value": "master"}]'
    spec:
      tolerations:
      - key: "node-role.kubernetes.io/master"
        effect: NoSchedule
      nodeSelector:
        node-role.kubernetes.io/master: ""
      dnsPolicy: Default  # Don't use cluster DNS (we are likely running before kube-dns)
      hostNetwork: true
      serviceAccount: dns-controller
      containers:
      - name: dns-controller
        image: kope/dns-controller:1.18.0-alpha.2
        command:
{{ range $arg := DnsControllerArgv }}
        - "{{ $arg }}"
{{ end }}
{{- if .EgressProxy }}
        env:
{{ range $name, $value := ProxyEnv }}
        - name: {{ $name }}
          value: {{ $value }}
{{ end }}
{{- end }}
{{- if eq .CloudProvider "digitalocean" }}
        env:
        - name: DIGITALOCEAN_ACCESS_TOKEN
          valueFrom:
            secretKeyRef:
              name: digitalocean
              key: access-token
{{- end }}
        resources:
          requests:
            cpu: 50m
            memory: 50Mi

---

apiVersion: v1
kind: ServiceAccount
metadata:
  name: dns-controller
  namespace: kube-system
  labels:
    k8s-addon: dns-controller.addons.k8s.io

---

apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  labels:
    k8s-addon: dns-controller.addons.k8s.io
  name: kops:dns-controller
rules:
- apiGroups:
  - ""
  resources:
  - endpoints
  - services
  - pods
  - ingress
  - nodes
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - "extensions"
  resources:
  - ingresses
  verbs:
  - get
  - list
  - watch

---

apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  labels:
    k8s-addon: dns-controller.addons.k8s.io
  name: kops:dns-controller
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: kops:dns-controller
subjects:
- apiGroup: rbac.authorization.k8s.io
  kind: User
  name: system:serviceaccount:kube-system:dns-controller
`)

func cloudupResourcesAddonsDnsControllerAddonsK8sIoK8s16YamlTemplateBytes() ([]byte, error) {
	return _cloudupResourcesAddonsDnsControllerAddonsK8sIoK8s16YamlTemplate, nil
}

func cloudupResourcesAddonsDnsControllerAddonsK8sIoK8s16YamlTemplate() (*asset, error) {
	bytes, err := cloudupResourcesAddonsDnsControllerAddonsK8sIoK8s16YamlTemplateBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "cloudup/resources/addons/dns-controller.addons.k8s.io/k8s-1.6.yaml.template", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _cloudupResourcesAddonsExternalDnsAddonsK8sIoReadmeMd = []byte(`# ExternalDNS

ExternalDNS synchronizes exposed Kubernetes Services and Ingresses with DNS providers.

## What it does

Inspired by [Kubernetes DNS](https://github.com/kubernetes/dns), Kubernetes' cluster-internal DNS server, ExternalDNS makes Kubernetes resources discoverable via public DNS servers. Like KubeDNS, it retrieves a list of resources (Services, Ingresses, etc.) from the [Kubernetes API](https://kubernetes.io/docs/api/) to determine a desired list of DNS records. *Unlike* KubeDNS, however, it's not a DNS server itself, but merely configures other DNS providers accordingly—e.g. [AWS Route 53](https://aws.amazon.com/route53/) or [Google CloudDNS](https://cloud.google.com/dns/docs/).

In a broader sense, ExternalDNS allows you to control DNS records dynamically via Kubernetes resources in a DNS provider-agnostic way.

## Deploying to a Cluster

The following tutorials are provided:

* [AWS](https://github.com/kubernetes-sigs/external-dns/blob/master/docs/tutorials/aws.md)
* [Azure](https://github.com/kubernetes-incubator/external-dns/blob/master/docs/tutorials/azure.md)
* [Cloudflare](https://github.com/kubernetes-incubator/external-dns/blob/master/docs/tutorials/cloudflare.md)
* [DigitalOcean](https://github.com/kubernetes-incubator/external-dns/blob/master/docs/tutorials/digitalocean.md)
* Google Container Engine
	* [Using Google's Default Ingress Controller](https://github.com/kubernetes-incubator/external-dns/blob/master/docs/tutorials/gke.md)
	* [Using the Nginx Ingress Controller](https://github.com/kubernetes-incubator/external-dns/blob/master/docs/tutorials/nginx-ingress.md)
* [FAQ](https://github.com/kubernetes-incubator/external-dns/blob/master/docs/faq.md)

## Github repository

Source code is managed under kubernetes-incubator at [external-dns](https://github.com/kubernetes-incubator/external-dns).`)

func cloudupResourcesAddonsExternalDnsAddonsK8sIoReadmeMdBytes() ([]byte, error) {
	return _cloudupResourcesAddonsExternalDnsAddonsK8sIoReadmeMd, nil
}

func cloudupResourcesAddonsExternalDnsAddonsK8sIoReadmeMd() (*asset, error) {
	bytes, err := cloudupResourcesAddonsExternalDnsAddonsK8sIoReadmeMdBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "cloudup/resources/addons/external-dns.addons.k8s.io/README.md", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _cloudupResourcesAddonsExternalDnsAddonsK8sIoK8s112YamlTemplate = []byte(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: external-dns
  namespace: kube-system
  labels:
    k8s-addon: external-dns.addons.k8s.io
    k8s-app: external-dns
    version: v0.4.4
spec:
  replicas: 1
  selector:
    matchLabels:
      k8s-app: external-dns
  template:
    metadata:
      labels:
        k8s-addon: external-dns.addons.k8s.io
        k8s-app: external-dns
        version: v0.4.4
      annotations:
        scheduler.alpha.kubernetes.io/critical-pod: ''
    spec:
      priorityClassName: system-cluster-critical
      serviceAccount: external-dns
      tolerations:
      - key: "node-role.kubernetes.io/master"
        effect: NoSchedule
      nodeSelector:
        node-role.kubernetes.io/master: ""
      dnsPolicy: Default  # Don't use cluster DNS (we are likely running before kube-dns)
      hostNetwork: true
      containers:
      - name: external-dns
        image: registry.opensource.zalan.do/teapot/external-dns:v0.4.4
        args:
{{ range $arg := ExternalDnsArgv }}
        - "{{ $arg }}"
{{ end }}
        resources:
          requests:
            cpu: 50m
            memory: 50Mi
---

apiVersion: v1
kind: ServiceAccount
metadata:
  name: external-dns
  namespace: kube-system
  labels:
    k8s-addon: external-dns.addons.k8s.io

---

apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  labels:
    k8s-addon: external-dns.addons.k8s.io
  name: kops:external-dns
rules:
- apiGroups:
  - ""
  resources:
  - services
  verbs:
  - list
- apiGroups:
  - extensions
  resources:
  - ingresses
  verbs:
  - list

---

apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  labels:
    k8s-addon: external-dns.addons.k8s.io
  name: kops:external-dns
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: kops:external-dns
subjects:
- kind: ServiceAccount
  name: external-dns
  namespace: kube-system
`)

func cloudupResourcesAddonsExternalDnsAddonsK8sIoK8s112YamlTemplateBytes() ([]byte, error) {
	return _cloudupResourcesAddonsExternalDnsAddonsK8sIoK8s112YamlTemplate, nil
}

func cloudupResourcesAddonsExternalDnsAddonsK8sIoK8s112YamlTemplate() (*asset, error) {
	bytes, err := cloudupResourcesAddonsExternalDnsAddonsK8sIoK8s112YamlTemplateBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "cloudup/resources/addons/external-dns.addons.k8s.io/k8s-1.12.yaml.template", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _cloudupResourcesAddonsExternalDnsAddonsK8sIoK8s16YamlTemplate = []byte(`apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: external-dns
  namespace: kube-system
  labels:
    k8s-addon: external-dns.addons.k8s.io
    k8s-app: external-dns
    version: v0.4.4
spec:
  replicas: 1
  selector:
    matchLabels:
      k8s-app: external-dns
  template:
    metadata:
      labels:
        k8s-addon: external-dns.addons.k8s.io
        k8s-app: external-dns
        version: v0.4.4
      annotations:
        scheduler.alpha.kubernetes.io/critical-pod: ''
        # For 1.6, we keep the old tolerations in case of a downgrade to 1.5
        scheduler.alpha.kubernetes.io/tolerations: '[{"key": "dedicated", "value": "master"}]'
    spec:
      serviceAccount: external-dns
      tolerations:
      - key: "node-role.kubernetes.io/master"
        effect: NoSchedule
      nodeSelector:
        node-role.kubernetes.io/master: ""
      dnsPolicy: Default  # Don't use cluster DNS (we are likely running before kube-dns)
      hostNetwork: true
      containers:
      - name: external-dns
        image: registry.opensource.zalan.do/teapot/external-dns:v0.4.4
        args:
{{ range $arg := ExternalDnsArgv }}
        - "{{ $arg }}"
{{ end }}
        resources:
          requests:
            cpu: 50m
            memory: 50Mi
---

apiVersion: v1
kind: ServiceAccount
metadata:
  name: external-dns
  namespace: kube-system
  labels:
    k8s-addon: external-dns.addons.k8s.io

---

apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  labels:
    k8s-addon: external-dns.addons.k8s.io
  name: kops:external-dns
rules:
- apiGroups:
  - ""
  resources:
  - services
  verbs:
  - list
- apiGroups:
  - extensions
  resources:
  - ingresses
  verbs:
  - list

---

apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  labels:
    k8s-addon: external-dns.addons.k8s.io
  name: kops:external-dns
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: kops:external-dns
subjects:
- kind: ServiceAccount
  name: external-dns
  namespace: kube-system
`)

func cloudupResourcesAddonsExternalDnsAddonsK8sIoK8s16YamlTemplateBytes() ([]byte, error) {
	return _cloudupResourcesAddonsExternalDnsAddonsK8sIoK8s16YamlTemplate, nil
}

func cloudupResourcesAddonsExternalDnsAddonsK8sIoK8s16YamlTemplate() (*asset, error) {
	bytes, err := cloudupResourcesAddonsExternalDnsAddonsK8sIoK8s16YamlTemplateBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "cloudup/resources/addons/external-dns.addons.k8s.io/k8s-1.6.yaml.template", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _cloudupResourcesAddonsKopsControllerAddonsK8sIoK8s116YamlTemplate = []byte(`apiVersion: v1
kind: ConfigMap
metadata:
  name: kops-controller
  namespace: kube-system
  labels:
    k8s-addon: kops-controller.addons.k8s.io
data:
  config.yaml: |
    {{ KopsControllerConfig }}

---

kind: DaemonSet
apiVersion: apps/v1
metadata:
  name: kops-controller
  namespace: kube-system
  labels:
    k8s-addon: kops-controller.addons.k8s.io
    k8s-app: kops-controller
    version: v1.18.0-alpha.2
spec:
  selector:
    matchLabels:
      k8s-app: kops-controller
  updateStrategy:
    type: RollingUpdate
    rollingUpdate:
      maxUnavailable: 1
  template:
    metadata:
      labels:
        k8s-addon: kops-controller.addons.k8s.io
        k8s-app: kops-controller
        version: v1.18.0-alpha.2
    spec:
      priorityClassName: system-node-critical
      tolerations:
      - key: "node-role.kubernetes.io/master"
        operator: Exists
      nodeSelector:
        node-role.kubernetes.io/master: ""
      dnsPolicy: Default  # Don't use cluster DNS (we are likely running before kube-dns)
      hostNetwork: true
      serviceAccount: kops-controller
      containers:
      - name: kops-controller
        image: kope/kops-controller:1.18.0-alpha.2
        volumeMounts:
{{ if .UseHostCertificates }}
        - mountPath: /etc/ssl/certs
          name: etc-ssl-certs
          readOnly: true
{{ end }}
        - mountPath: /etc/kubernetes/kops-controller/
          name: kops-controller-config
        command:
{{ range $arg := KopsControllerArgv }}
        - "{{ $arg }}"
{{ end }}
{{- if KopsSystemEnv }}
        env:
{{ range $var := KopsSystemEnv }}
        - name: {{ $var.Name }}
          value: {{ $var.Value }}
{{ end }}
{{- end }}
        resources:
          requests:
            cpu: 50m
            memory: 50Mi
        securityContext:
          runAsNonRoot: true
      volumes:
{{ if .UseHostCertificates }}
      - hostPath:
          path: /etc/ssl/certs
          type: DirectoryOrCreate
        name: etc-ssl-certs
{{ end }}
      - name: kops-controller-config
        configMap:
          name: kops-controller

---

apiVersion: v1
kind: ServiceAccount
metadata:
  name: kops-controller
  namespace: kube-system
  labels:
    k8s-addon: kops-controller.addons.k8s.io

---

apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  labels:
    k8s-addon: kops-controller.addons.k8s.io
  name: kops-controller
rules:
- apiGroups:
  - ""
  resources:
  - nodes
  verbs:
  - get
  - list
  - watch
  - patch

---

apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  labels:
    k8s-addon: kops-controller.addons.k8s.io
  name: kops-controller
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: kops-controller
subjects:
- apiGroup: rbac.authorization.k8s.io
  kind: User
  name: system:serviceaccount:kube-system:kops-controller

---

apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  labels:
    k8s-addon: kops-controller.addons.k8s.io
  name: kops-controller
  namespace: kube-system
rules:
- apiGroups:
  - ""
  resources:
  - events
  verbs:
  - get
  - list
  - watch
  - create
- apiGroups:
  - ""
  resources:
  - configmaps
  resourceNames:
  - kops-controller-leader
  verbs:
  - get
  - list
  - watch
  - patch
  - update
  - delete
# Workaround for https://github.com/kubernetes/kubernetes/issues/80295
# We can't restrict creation of objects by name
- apiGroups:
  - ""
  resources:
  - configmaps
  verbs:
  - create

---

apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  labels:
    k8s-addon: kops-controller.addons.k8s.io
  name: kops-controller
  namespace: kube-system
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: kops-controller
subjects:
- apiGroup: rbac.authorization.k8s.io
  kind: User
  name: system:serviceaccount:kube-system:kops-controller
`)

func cloudupResourcesAddonsKopsControllerAddonsK8sIoK8s116YamlTemplateBytes() ([]byte, error) {
	return _cloudupResourcesAddonsKopsControllerAddonsK8sIoK8s116YamlTemplate, nil
}

func cloudupResourcesAddonsKopsControllerAddonsK8sIoK8s116YamlTemplate() (*asset, error) {
	bytes, err := cloudupResourcesAddonsKopsControllerAddonsK8sIoK8s116YamlTemplateBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "cloudup/resources/addons/kops-controller.addons.k8s.io/k8s-1.16.yaml.template", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _cloudupResourcesAddonsKubeDnsAddonsK8sIoK8s112YamlTemplate = []byte(`# Copyright 2019 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

{{- if or (.KubeDNS.UpstreamNameservers) (.KubeDNS.StubDomains) }}
apiVersion: v1
kind: ConfigMap
metadata:
  name: kube-dns
  namespace: kube-system
data:
  {{- if .KubeDNS.UpstreamNameservers }}
  upstreamNameservers: |
    {{ ToJSON .KubeDNS.UpstreamNameservers }}
  {{- end }}
  {{- if .KubeDNS.StubDomains }}
  stubDomains: |
    {{ ToJSON .KubeDNS.StubDomains }}
  {{- end }}

---
{{- end }}

apiVersion: apps/v1
kind: Deployment
metadata:
  name: kube-dns-autoscaler
  namespace: kube-system
  labels:
    k8s-addon: kube-dns.addons.k8s.io
    k8s-app: kube-dns-autoscaler
    kubernetes.io/cluster-service: "true"
spec:
  selector:
    matchLabels:
      k8s-app: kube-dns-autoscaler
  template:
    metadata:
      labels:
        k8s-app: kube-dns-autoscaler
      annotations:
        scheduler.alpha.kubernetes.io/critical-pod: ''
    spec:
      containers:
      - name: autoscaler
        image: k8s.gcr.io/cluster-proportional-autoscaler-{{Arch}}:1.4.0
        resources:
            requests:
                cpu: "20m"
                memory: "10Mi"
        command:
          - /cluster-proportional-autoscaler
          - --namespace=kube-system
          - --configmap=kube-dns-autoscaler
          # Should keep target in sync with cluster/addons/dns/kubedns-controller.yaml.base
          - --target=Deployment/kube-dns
          # When cluster is using large nodes(with more cores), "coresPerReplica" should dominate.
          # If using small nodes, "nodesPerReplica" should dominate.
          - --default-params={"linear":{"coresPerReplica":256,"nodesPerReplica":16,"preventSinglePointFailure":true}}
          - --logtostderr=true
          - --v=2
      priorityClassName: system-cluster-critical
      tolerations:
      - key: "CriticalAddonsOnly"
        operator: "Exists"
      serviceAccountName: kube-dns-autoscaler

---

apiVersion: apps/v1
kind: Deployment
metadata:
  name: kube-dns
  namespace: kube-system
  labels:
    k8s-addon: kube-dns.addons.k8s.io
    k8s-app: kube-dns
    kubernetes.io/cluster-service: "true"
spec:
  # replicas: not specified here:
  # 1. In order to make Addon Manager do not reconcile this replicas parameter.
  # 2. Default is 1.
  # 3. Will be tuned in real time if DNS horizontal auto-scaling is turned on.
  strategy:
    rollingUpdate:
      maxSurge: 10%
      maxUnavailable: 0
  selector:
    matchLabels:
      k8s-app: kube-dns
  template:
    metadata:
      labels:
        k8s-app: kube-dns
      annotations:
        scheduler.alpha.kubernetes.io/critical-pod: ''
        prometheus.io/scrape: 'true'
        prometheus.io/port: '10055'
    spec:
      affinity:
        podAntiAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
            - weight: 1
              podAffinityTerm:
                labelSelector:
                  matchExpressions:
                    - key: k8s-app
                      operator: In
                      values:
                        - kube-dns
                topologyKey: kubernetes.io/hostname
      dnsPolicy: Default  # Don't use cluster DNS.
      priorityClassName: system-cluster-critical
      serviceAccountName: kube-dns
      volumes:
      - name: kube-dns-config
        configMap:
          name: kube-dns
          optional: true

      containers:
      - name: kubedns
        image: k8s.gcr.io/k8s-dns-kube-dns-{{Arch}}:1.14.13
        resources:
          # TODO: Set memory limits when we've profiled the container for large
          # clusters, then set request = limit to keep this container in
          # guaranteed class. Currently, this container falls into the
          # "burstable" category so the kubelet doesn't backoff from restarting it.
          limits:
            memory: 170Mi
          requests:
            cpu: 100m
            memory: 70Mi
        livenessProbe:
          httpGet:
            path: /healthcheck/kubedns
            port: 10054
            scheme: HTTP
          initialDelaySeconds: 60
          timeoutSeconds: 5
          successThreshold: 1
          failureThreshold: 5
        readinessProbe:
          httpGet:
            path: /readiness
            port: 8081
            scheme: HTTP
          # we poll on pod startup for the Kubernetes master service and
          # only setup the /readiness HTTP server once that's available.
          initialDelaySeconds: 3
          timeoutSeconds: 5
        args:
        - --config-dir=/kube-dns-config
        - --dns-port=10053
        - --domain={{ KubeDNS.Domain }}.
        - --v=2
        env:
        - name: PROMETHEUS_PORT
          value: "10055"
        ports:
        - containerPort: 10053
          name: dns-local
          protocol: UDP
        - containerPort: 10053
          name: dns-tcp-local
          protocol: TCP
        - containerPort: 10055
          name: metrics
          protocol: TCP
        volumeMounts:
        - name: kube-dns-config
          mountPath: /kube-dns-config

      - name: dnsmasq
        image: k8s.gcr.io/k8s-dns-dnsmasq-nanny-{{Arch}}:1.14.13
        livenessProbe:
          httpGet:
            path: /healthcheck/dnsmasq
            port: 10054
            scheme: HTTP
          initialDelaySeconds: 60
          timeoutSeconds: 5
          successThreshold: 1
          failureThreshold: 5
        args:
        - -v=2
        - -logtostderr
        - -configDir=/etc/k8s/dns/dnsmasq-nanny
        - -restartDnsmasq=true
        - --
        - -k
        - --cache-size={{ KubeDNS.CacheMaxSize }}
        - --dns-forward-max={{ KubeDNS.CacheMaxConcurrent }}
        - --no-negcache
        - --log-facility=-
        - --server=/{{ KubeDNS.Domain }}/127.0.0.1#10053
        - --server=/in-addr.arpa/127.0.0.1#10053
        - --server=/in6.arpa/127.0.0.1#10053
        - --min-port=1024
        ports:
        - containerPort: 53
          name: dns
          protocol: UDP
        - containerPort: 53
          name: dns-tcp
          protocol: TCP
        # see: https://github.com/kubernetes/kubernetes/issues/29055 for details
        resources:
          requests:
            cpu: 150m
            memory: 20Mi
        volumeMounts:
        - name: kube-dns-config
          mountPath: /etc/k8s/dns/dnsmasq-nanny

      - name: sidecar
        image: k8s.gcr.io/k8s-dns-sidecar-amd64:1.14.13
        livenessProbe:
          httpGet:
            path: /metrics
            port: 10054
            scheme: HTTP
          initialDelaySeconds: 60
          timeoutSeconds: 5
          successThreshold: 1
          failureThreshold: 5
        args:
        - --v=2
        - --logtostderr
        - --probe=kubedns,127.0.0.1:10053,kubernetes.default.svc.{{ KubeDNS.Domain }},5,A
        - --probe=dnsmasq,127.0.0.1:53,kubernetes.default.svc.{{ KubeDNS.Domain }},5,A
        ports:
        - containerPort: 10054
          name: metrics
          protocol: TCP
        resources:
          requests:
            memory: 20Mi
            cpu: 10m

---

apiVersion: v1
kind: Service
metadata:
  name: kube-dns
  namespace: kube-system
  labels:
    k8s-addon: kube-dns.addons.k8s.io
    k8s-app: kube-dns
    kubernetes.io/cluster-service: "true"
    kubernetes.io/name: "KubeDNS"
spec:
  selector:
    k8s-app: kube-dns
  clusterIP: {{ KubeDNS.ServerIP }}
  ports:
  - name: dns
    port: 53
    protocol: UDP
  - name: dns-tcp
    port: 53
    protocol: TCP

---

apiVersion: v1
kind: ServiceAccount
metadata:
  name: kube-dns-autoscaler
  namespace: kube-system
  labels:
    k8s-addon: kube-dns.addons.k8s.io

---

apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  labels:
    k8s-addon: kube-dns.addons.k8s.io
  name: kube-dns-autoscaler
rules:
  - apiGroups: [""]
    resources: ["nodes"]
    verbs: ["list","watch"]
  - apiGroups: [""]
    resources: ["replicationcontrollers/scale"]
    verbs: ["get", "update"]
  - apiGroups: ["extensions", "apps"]
    resources: ["deployments/scale", "replicasets/scale"]
    verbs: ["get", "update"]
# Remove the configmaps rule once below issue is fixed:
# kubernetes-incubator/cluster-proportional-autoscaler#16
  - apiGroups: [""]
    resources: ["configmaps"]
    verbs: ["get", "create"]

---

apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  labels:
    k8s-addon: kube-dns.addons.k8s.io
  name: kube-dns-autoscaler
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: kube-dns-autoscaler
subjects:
- kind: ServiceAccount
  name: kube-dns-autoscaler
  namespace: kube-system

---

apiVersion: policy/v1beta1
kind: PodDisruptionBudget
metadata:
  name: kube-dns
  namespace: kube-system
spec:
  selector:
    matchLabels:
      k8s-app: kube-dns
  minAvailable: 1
`)

func cloudupResourcesAddonsKubeDnsAddonsK8sIoK8s112YamlTemplateBytes() ([]byte, error) {
	return _cloudupResourcesAddonsKubeDnsAddonsK8sIoK8s112YamlTemplate, nil
}

func cloudupResourcesAddonsKubeDnsAddonsK8sIoK8s112YamlTemplate() (*asset, error) {
	bytes, err := cloudupResourcesAddonsKubeDnsAddonsK8sIoK8s112YamlTemplateBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "cloudup/resources/addons/kube-dns.addons.k8s.io/k8s-1.12.yaml.template", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _cloudupResourcesAddonsKubeDnsAddonsK8sIoK8s16YamlTemplate = []byte(`# Copyright 2019 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

{{- if or (.KubeDNS.UpstreamNameservers) (.KubeDNS.StubDomains) }}
apiVersion: v1
kind: ConfigMap
metadata:
  name: kube-dns
  namespace: kube-system
data:
  {{- if .KubeDNS.UpstreamNameservers }}
  upstreamNameservers: |
    {{ ToJSON .KubeDNS.UpstreamNameservers }}
  {{- end }}
  {{- if .KubeDNS.StubDomains }}
  stubDomains: |
    {{ ToJSON .KubeDNS.StubDomains }}
  {{- end }}

---
{{- end }}

apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: kube-dns-autoscaler
  namespace: kube-system
  labels:
    k8s-addon: kube-dns.addons.k8s.io
    k8s-app: kube-dns-autoscaler
    kubernetes.io/cluster-service: "true"
spec:
  template:
    metadata:
      labels:
        k8s-app: kube-dns-autoscaler
      annotations:
        scheduler.alpha.kubernetes.io/critical-pod: ''
        # For 1.6, we keep the old tolerations in case of a downgrade to 1.5
        scheduler.alpha.kubernetes.io/tolerations: '[{"key":"CriticalAddonsOnly", "operator":"Exists"}]'
    spec:
      containers:
      - name: autoscaler
        image: k8s.gcr.io/cluster-proportional-autoscaler-{{Arch}}:1.1.2-r2
        resources:
            requests:
                cpu: "20m"
                memory: "10Mi"
        command:
          - /cluster-proportional-autoscaler
          - --namespace=kube-system
          - --configmap=kube-dns-autoscaler
          # Should keep target in sync with cluster/addons/dns/kubedns-controller.yaml.base
          - --target=Deployment/kube-dns
          # When cluster is using large nodes(with more cores), "coresPerReplica" should dominate.
          # If using small nodes, "nodesPerReplica" should dominate.
          - --default-params={"linear":{"coresPerReplica":256,"nodesPerReplica":16,"preventSinglePointFailure":true}}
          - --logtostderr=true
          - --v=2
      tolerations:
      - key: "CriticalAddonsOnly"
        operator: "Exists"
      serviceAccountName: kube-dns-autoscaler

---

apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: kube-dns
  namespace: kube-system
  labels:
    k8s-addon: kube-dns.addons.k8s.io
    k8s-app: kube-dns
    kubernetes.io/cluster-service: "true"
spec:
  # replicas: not specified here:
  # 1. In order to make Addon Manager do not reconcile this replicas parameter.
  # 2. Default is 1.
  # 3. Will be tuned in real time if DNS horizontal auto-scaling is turned on.
  strategy:
    rollingUpdate:
      maxSurge: 10%
      maxUnavailable: 0
  selector:
    matchLabels:
      k8s-app: kube-dns
  template:
    metadata:
      labels:
        k8s-app: kube-dns
      annotations:
        scheduler.alpha.kubernetes.io/critical-pod: ''
        # For 1.6, we keep the old tolerations in case of a downgrade to 1.5
        scheduler.alpha.kubernetes.io/tolerations: '[{"key":"CriticalAddonsOnly", "operator":"Exists"}]'
        prometheus.io/scrape: 'true'
        prometheus.io/port: '10055'
    spec:
      dnsPolicy: Default  # Don't use cluster DNS.
      serviceAccountName: kube-dns
      volumes:
      - name: kube-dns-config
        configMap:
          name: kube-dns
          optional: true

      containers:
      - name: kubedns
        image: k8s.gcr.io/k8s-dns-kube-dns-{{Arch}}:1.14.10
        resources:
          # TODO: Set memory limits when we've profiled the container for large
          # clusters, then set request = limit to keep this container in
          # guaranteed class. Currently, this container falls into the
          # "burstable" category so the kubelet doesn't backoff from restarting it.
          limits:
            memory: 170Mi
          requests:
            cpu: 100m
            memory: 70Mi
        livenessProbe:
          httpGet:
            path: /healthcheck/kubedns
            port: 10054
            scheme: HTTP
          initialDelaySeconds: 60
          timeoutSeconds: 5
          successThreshold: 1
          failureThreshold: 5
        readinessProbe:
          httpGet:
            path: /readiness
            port: 8081
            scheme: HTTP
          # we poll on pod startup for the Kubernetes master service and
          # only setup the /readiness HTTP server once that's available.
          initialDelaySeconds: 3
          timeoutSeconds: 5
        args:
        - --config-dir=/kube-dns-config
        - --dns-port=10053
        - --domain={{ KubeDNS.Domain }}.
        - --v=2
        env:
        - name: PROMETHEUS_PORT
          value: "10055"
        ports:
        - containerPort: 10053
          name: dns-local
          protocol: UDP
        - containerPort: 10053
          name: dns-tcp-local
          protocol: TCP
        - containerPort: 10055
          name: metrics
          protocol: TCP
        volumeMounts:
        - name: kube-dns-config
          mountPath: /kube-dns-config

      - name: dnsmasq
        image: k8s.gcr.io/k8s-dns-dnsmasq-nanny-{{Arch}}:1.14.10
        livenessProbe:
          httpGet:
            path: /healthcheck/dnsmasq
            port: 10054
            scheme: HTTP
          initialDelaySeconds: 60
          timeoutSeconds: 5
          successThreshold: 1
          failureThreshold: 5
        args:
        - -v=2
        - -logtostderr
        - -configDir=/etc/k8s/dns/dnsmasq-nanny
        - -restartDnsmasq=true
        - --
        - -k
        - --cache-size={{ KubeDNS.CacheMaxSize }}
        - --dns-forward-max={{ KubeDNS.CacheMaxConcurrent }}
        - --no-negcache
        - --log-facility=-
        - --server=/{{ KubeDNS.Domain }}/127.0.0.1#10053
        - --server=/in-addr.arpa/127.0.0.1#10053
        - --server=/in6.arpa/127.0.0.1#10053
        - --min-port=1024
        ports:
        - containerPort: 53
          name: dns
          protocol: UDP
        - containerPort: 53
          name: dns-tcp
          protocol: TCP
        # see: https://github.com/kubernetes/kubernetes/issues/29055 for details
        resources:
          requests:
            cpu: 150m
            memory: 20Mi
        volumeMounts:
        - name: kube-dns-config
          mountPath: /etc/k8s/dns/dnsmasq-nanny

      - name: sidecar
        image: k8s.gcr.io/k8s-dns-sidecar-amd64:1.14.10
        livenessProbe:
          httpGet:
            path: /metrics
            port: 10054
            scheme: HTTP
          initialDelaySeconds: 60
          timeoutSeconds: 5
          successThreshold: 1
          failureThreshold: 5
        args:
        - --v=2
        - --logtostderr
        - --probe=kubedns,127.0.0.1:10053,kubernetes.default.svc.{{ KubeDNS.Domain }},5,A
        - --probe=dnsmasq,127.0.0.1:53,kubernetes.default.svc.{{ KubeDNS.Domain }},5,A
        ports:
        - containerPort: 10054
          name: metrics
          protocol: TCP
        resources:
          requests:
            memory: 20Mi
            cpu: 10m

---

apiVersion: v1
kind: Service
metadata:
  name: kube-dns
  namespace: kube-system
  labels:
    k8s-addon: kube-dns.addons.k8s.io
    k8s-app: kube-dns
    kubernetes.io/cluster-service: "true"
    kubernetes.io/name: "KubeDNS"
spec:
  selector:
    k8s-app: kube-dns
  clusterIP: {{ KubeDNS.ServerIP }}
  ports:
  - name: dns
    port: 53
    protocol: UDP
  - name: dns-tcp
    port: 53
    protocol: TCP

---

apiVersion: v1
kind: ServiceAccount
metadata:
  name: kube-dns-autoscaler
  namespace: kube-system
  labels:
    k8s-addon: kube-dns.addons.k8s.io

---

apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  labels:
    k8s-addon: kube-dns.addons.k8s.io
  name: kube-dns-autoscaler
rules:
  - apiGroups: [""]
    resources: ["nodes"]
    verbs: ["list"]
  - apiGroups: [""]
    resources: ["replicationcontrollers/scale"]
    verbs: ["get", "update"]
  - apiGroups: ["extensions"]
    resources: ["deployments/scale", "replicasets/scale"]
    verbs: ["get", "update"]
# Remove the configmaps rule once below issue is fixed:
# kubernetes-incubator/cluster-proportional-autoscaler#16
  - apiGroups: [""]
    resources: ["configmaps"]
    verbs: ["get", "create"]

---

apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  labels:
    k8s-addon: kube-dns.addons.k8s.io
  name: kube-dns-autoscaler
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: kube-dns-autoscaler
subjects:
- kind: ServiceAccount
  name: kube-dns-autoscaler
  namespace: kube-system
`)

func cloudupResourcesAddonsKubeDnsAddonsK8sIoK8s16YamlTemplateBytes() ([]byte, error) {
	return _cloudupResourcesAddonsKubeDnsAddonsK8sIoK8s16YamlTemplate, nil
}

func cloudupResourcesAddonsKubeDnsAddonsK8sIoK8s16YamlTemplate() (*asset, error) {
	bytes, err := cloudupResourcesAddonsKubeDnsAddonsK8sIoK8s16YamlTemplateBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "cloudup/resources/addons/kube-dns.addons.k8s.io/k8s-1.6.yaml.template", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _cloudupResourcesAddonsKubeletApiRbacAddonsK8sIoK8s19Yaml = []byte(`---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: kops:system:kubelet-api-admin
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: system:kubelet-api-admin
subjects:
# TODO: perhaps change the client cerificate, place into a group and using a group selector instead?
- apiGroup: rbac.authorization.k8s.io
  kind: User
  name: kubelet-api
`)

func cloudupResourcesAddonsKubeletApiRbacAddonsK8sIoK8s19YamlBytes() ([]byte, error) {
	return _cloudupResourcesAddonsKubeletApiRbacAddonsK8sIoK8s19Yaml, nil
}

func cloudupResourcesAddonsKubeletApiRbacAddonsK8sIoK8s19Yaml() (*asset, error) {
	bytes, err := cloudupResourcesAddonsKubeletApiRbacAddonsK8sIoK8s19YamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "cloudup/resources/addons/kubelet-api.rbac.addons.k8s.io/k8s-1.9.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _cloudupResourcesAddonsLimitRangeAddonsK8sIoAddonYaml = []byte(`kind: Addons
metadata:
  name: limit-range
spec:
  addons:
  - version: 1.5.0
    selector:
      k8s-addon: limit-range.addons.k8s.io
    manifest: v1.5.0.yaml
`)

func cloudupResourcesAddonsLimitRangeAddonsK8sIoAddonYamlBytes() ([]byte, error) {
	return _cloudupResourcesAddonsLimitRangeAddonsK8sIoAddonYaml, nil
}

func cloudupResourcesAddonsLimitRangeAddonsK8sIoAddonYaml() (*asset, error) {
	bytes, err := cloudupResourcesAddonsLimitRangeAddonsK8sIoAddonYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "cloudup/resources/addons/limit-range.addons.k8s.io/addon.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _cloudupResourcesAddonsLimitRangeAddonsK8sIoV150Yaml = []byte(`apiVersion: "v1"
kind: "LimitRange"
metadata:
  name: "limits"
  namespace: default
spec:
  limits:
    - type: "Container"
      defaultRequest:
        cpu: "100m"
`)

func cloudupResourcesAddonsLimitRangeAddonsK8sIoV150YamlBytes() ([]byte, error) {
	return _cloudupResourcesAddonsLimitRangeAddonsK8sIoV150Yaml, nil
}

func cloudupResourcesAddonsLimitRangeAddonsK8sIoV150Yaml() (*asset, error) {
	bytes, err := cloudupResourcesAddonsLimitRangeAddonsK8sIoV150YamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "cloudup/resources/addons/limit-range.addons.k8s.io/v1.5.0.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _cloudupResourcesAddonsMetadataProxyAddonsK8sIoAddonYaml = []byte(`kind: Addons
metadata:
  name: metadata-proxy
spec:
  addons:
  - version: 0.1.12
    selector:
      k8s-addon: metadata-proxy.addons.k8s.io
    manifest: v0.12.yaml

`)

func cloudupResourcesAddonsMetadataProxyAddonsK8sIoAddonYamlBytes() ([]byte, error) {
	return _cloudupResourcesAddonsMetadataProxyAddonsK8sIoAddonYaml, nil
}

func cloudupResourcesAddonsMetadataProxyAddonsK8sIoAddonYaml() (*asset, error) {
	bytes, err := cloudupResourcesAddonsMetadataProxyAddonsK8sIoAddonYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "cloudup/resources/addons/metadata-proxy.addons.k8s.io/addon.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _cloudupResourcesAddonsMetadataProxyAddonsK8sIoV0112Yaml = []byte(`# Borrowed from https://github.com/kubernetes/kubernetes/tree/master/cluster/addons/metadata-proxy

apiVersion: v1
kind: ServiceAccount
metadata:
  name: metadata-proxy
  namespace: kube-system
  labels:
    k8s-app: metadata-proxy
    kubernetes.io/cluster-service: "true"
    addonmanager.kubernetes.io/mode: Reconcile
---
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: metadata-proxy-v0.12
  namespace: kube-system
  labels:
    k8s-app: metadata-proxy
    kubernetes.io/cluster-service: "true"
    addonmanager.kubernetes.io/mode: Reconcile
    version: v0.12
spec:
  selector:
    matchLabels:
      k8s-app: metadata-proxy
      version: v0.12
  updateStrategy:
    type: RollingUpdate
  template:
    metadata:
      labels:
        k8s-app: metadata-proxy
        kubernetes.io/cluster-service: "true"
        version: v0.12
    spec:
      priorityClassName: system-node-critical
      serviceAccountName: metadata-proxy
      hostNetwork: true
      dnsPolicy: Default
      tolerations:
      - operator: "Exists"
        effect: "NoExecute"
      - operator: "Exists"
        effect: "NoSchedule"
      hostNetwork: true
      initContainers:
      - name: update-ipdtables
        securityContext:
          privileged: true
        image: gcr.io/google_containers/k8s-custom-iptables:1.0
        imagePullPolicy: Always
        command: [ "/bin/sh", "-c", "/sbin/iptables -t nat -I PREROUTING -p tcp --dport 80 -d 169.254.169.254 -j DNAT --to-destination 127.0.0.1:988" ]
        volumeMounts:
        - name: host
          mountPath: /host
      volumes:
      - name: host
        hostPath:
          path: /
          type: Directory
      containers:
      - name: metadata-proxy
        image: k8s.gcr.io/metadata-proxy:v0.1.12
        securityContext:
          privileged: true
        # Request and limit resources to get guaranteed QoS.
        resources:
          requests:
            memory: "25Mi"
            cpu: "30m"
          limits:
            memory: "25Mi"
            cpu: "30m"
      # BEGIN_PROMETHEUS_TO_SD
      - name: prometheus-to-sd-exporter
        image: k8s.gcr.io/prometheus-to-sd:v0.5.0
        # Request and limit resources to get guaranteed QoS.
        resources:
          requests:
            memory: "20Mi"
            cpu: "2m"
          limits:
            memory: "20Mi"
            cpu: "2m"
        command:
          - /monitor
          - --stackdriver-prefix=custom.googleapis.com/addons
          - --source=metadata_proxy:http://127.0.0.1:989?whitelisted=request_count
          - --pod-id=$(POD_NAME)
          - --namespace-id=$(POD_NAMESPACE)
        env:
          - name: POD_NAME
            valueFrom:
              fieldRef:
                fieldPath: metadata.name
          - name: POD_NAMESPACE
            valueFrom:
              fieldRef:
                fieldPath: metadata.namespace
      # END_PROMETHEUS_TO_SD
      nodeSelector:
        cloud.google.com/metadata-proxy-ready: "true"
        beta.kubernetes.io/os: linux
      terminationGracePeriodSeconds: 30
`)

func cloudupResourcesAddonsMetadataProxyAddonsK8sIoV0112YamlBytes() ([]byte, error) {
	return _cloudupResourcesAddonsMetadataProxyAddonsK8sIoV0112Yaml, nil
}

func cloudupResourcesAddonsMetadataProxyAddonsK8sIoV0112Yaml() (*asset, error) {
	bytes, err := cloudupResourcesAddonsMetadataProxyAddonsK8sIoV0112YamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "cloudup/resources/addons/metadata-proxy.addons.k8s.io/v0.1.12.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _cloudupResourcesAddonsNetworkingAmazonVpcRoutedEniK8s110YamlTemplate = []byte(`# Vendored from https://github.com/aws/amazon-vpc-cni-k8s/blob/v1.3.3/config/v1.3/aws-k8s-cni.yaml

apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: aws-node
rules:
- apiGroups:
  - crd.k8s.amazonaws.com
  resources:
  - "*"
  - namespaces
  verbs:
  - "*"
- apiGroups: [""]
  resources:
  - pods
  - nodes
  - namespaces
  verbs: ["list", "watch", "get"]
- apiGroups: ["extensions"]
  resources:
  - daemonsets
  verbs: ["list", "watch"]
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: aws-node
  namespace: kube-system
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: aws-node
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: aws-node
subjects:
- kind: ServiceAccount
  name: aws-node
  namespace: kube-system
---
kind: DaemonSet
apiVersion: extensions/v1beta1
metadata:
  name: aws-node
  namespace: kube-system
  labels:
    k8s-app: aws-node
spec:
  updateStrategy:
    type: RollingUpdate
  selector:
    matchLabels:
      k8s-app: aws-node
  template:
    metadata:
      labels:
        k8s-app: aws-node
      annotations:
        scheduler.alpha.kubernetes.io/critical-pod: ''
    spec:
      serviceAccountName: aws-node
      hostNetwork: true
      priorityClassName: system-node-critical
      tolerations:
      - operator: Exists
      containers:
      - image: "{{- or .Networking.AmazonVPC.ImageName "602401143452.dkr.ecr.us-west-2.amazonaws.com/amazon-k8s-cni:1.3.3" }}"
        ports:
        - containerPort: 61678
          name: metrics
        name: aws-node
        env:
          - name: CLUSTER_NAME
            value: {{ ClusterName }}
          - name: AWS_VPC_K8S_CNI_LOGLEVEL
            value: DEBUG
          - name: MY_NODE_NAME
            valueFrom:
              fieldRef:
                fieldPath: spec.nodeName
          - name: WATCH_NAMESPACE
            valueFrom:
              fieldRef:
                fieldPath: metadata.namespace
          {{- range .Networking.AmazonVPC.Env }}
          - name: {{ .Name }}
            value: "{{ .Value }}"
          {{- end }}
        resources:
          requests:
            cpu: 10m
        securityContext:
          privileged: true
        volumeMounts:
        - mountPath: /host/opt/cni/bin
          name: cni-bin-dir
        - mountPath: /host/etc/cni/net.d
          name: cni-net-dir
        - mountPath: /host/var/log
          name: log-dir
        - mountPath: /var/run/docker.sock
          name: dockersock
      volumes:
      - name: cni-bin-dir
        hostPath:
          path: /opt/cni/bin
      - name: cni-net-dir
        hostPath:
          path: /etc/cni/net.d
      - name: log-dir
        hostPath:
          path: /var/log
      - name: dockersock
        hostPath:
          path: /var/run/docker.sock
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: eniconfigs.crd.k8s.amazonaws.com
spec:
  scope: Cluster
  group: crd.k8s.amazonaws.com
  version: v1alpha1
  names:
    plural: eniconfigs
    singular: eniconfig
    kind: ENIConfig
`)

func cloudupResourcesAddonsNetworkingAmazonVpcRoutedEniK8s110YamlTemplateBytes() ([]byte, error) {
	return _cloudupResourcesAddonsNetworkingAmazonVpcRoutedEniK8s110YamlTemplate, nil
}

func cloudupResourcesAddonsNetworkingAmazonVpcRoutedEniK8s110YamlTemplate() (*asset, error) {
	bytes, err := cloudupResourcesAddonsNetworkingAmazonVpcRoutedEniK8s110YamlTemplateBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "cloudup/resources/addons/networking.amazon-vpc-routed-eni/k8s-1.10.yaml.template", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _cloudupResourcesAddonsNetworkingAmazonVpcRoutedEniK8s112YamlTemplate = []byte(`# Vendored from https://raw.githubusercontent.com/aws/amazon-vpc-cni-k8s/v1.5.5/config/v1.5/aws-k8s-cni.yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: aws-node
rules:
- apiGroups:
  - crd.k8s.amazonaws.com
  resources:
  - "*"
  - namespaces
  verbs:
  - "*"
- apiGroups: [""]
  resources:
  - pods
  - nodes
  - namespaces
  verbs: ["list", "watch", "get"]
- apiGroups: ["extensions"]
  resources:
  - daemonsets
  verbs: ["list", "watch"]
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: aws-node
  namespace: kube-system
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: aws-node
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: aws-node
subjects:
- kind: ServiceAccount
  name: aws-node
  namespace: kube-system
---
kind: DaemonSet
apiVersion: apps/v1
metadata:
  name: aws-node
  namespace: kube-system
  labels:
    k8s-app: aws-node
spec:
  updateStrategy:
    type: RollingUpdate
  selector:
    matchLabels:
      k8s-app: aws-node
  template:
    metadata:
      labels:
        k8s-app: aws-node
    spec:
      priorityClassName: system-node-critical
      affinity:
        nodeAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
              - matchExpressions:
                  - key: "beta.kubernetes.io/os"
                    operator: In
                    values:
                      - linux
                  - key: "beta.kubernetes.io/arch"
                    operator: In
                    values:
                      - amd64
      serviceAccountName: aws-node
      hostNetwork: true
      tolerations:
      - operator: Exists
      containers:
      - image: "{{- or .Networking.AmazonVPC.ImageName "602401143452.dkr.ecr.us-west-2.amazonaws.com/amazon-k8s-cni:v1.5.5" }}"
        imagePullPolicy: Always
        ports:
        - containerPort: 61678
          name: metrics
        name: aws-node
        env:
          - name: CLUSTER_NAME
            value: {{ ClusterName }}
          - name: AWS_VPC_K8S_CNI_LOGLEVEL
            value: DEBUG
          - name: MY_NODE_NAME
            valueFrom:
              fieldRef:
                fieldPath: spec.nodeName
          - name: WATCH_NAMESPACE
            valueFrom:
              fieldRef:
                fieldPath: metadata.namespace
          {{- range .Networking.AmazonVPC.Env }}
          - name: {{ .Name }}
            value: "{{ .Value }}"
          {{- end }}
        resources:
          requests:
            cpu: 10m
        securityContext:
          privileged: true
        volumeMounts:
        - mountPath: /host/opt/cni/bin
          name: cni-bin-dir
        - mountPath: /host/etc/cni/net.d
          name: cni-net-dir
        - mountPath: /host/var/log
          name: log-dir
        - mountPath: /var/run/docker.sock
          name: dockersock
      volumes:
      - name: cni-bin-dir
        hostPath:
          path: /opt/cni/bin
      - name: cni-net-dir
        hostPath:
          path: /etc/cni/net.d
      - name: log-dir
        hostPath:
          path: /var/log
      - name: dockersock
        hostPath:
          path: /var/run/docker.sock
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: eniconfigs.crd.k8s.amazonaws.com
spec:
  scope: Cluster
  group: crd.k8s.amazonaws.com
  versions:
  - name: v1alpha1
    served: true
    storage: true
  names:
    plural: eniconfigs
    singular: eniconfig
    kind: ENIConfig
`)

func cloudupResourcesAddonsNetworkingAmazonVpcRoutedEniK8s112YamlTemplateBytes() ([]byte, error) {
	return _cloudupResourcesAddonsNetworkingAmazonVpcRoutedEniK8s112YamlTemplate, nil
}

func cloudupResourcesAddonsNetworkingAmazonVpcRoutedEniK8s112YamlTemplate() (*asset, error) {
	bytes, err := cloudupResourcesAddonsNetworkingAmazonVpcRoutedEniK8s112YamlTemplateBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "cloudup/resources/addons/networking.amazon-vpc-routed-eni/k8s-1.12.yaml.template", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _cloudupResourcesAddonsNetworkingAmazonVpcRoutedEniK8s116YamlTemplate = []byte(`# Vendored from https://raw.githubusercontent.com/aws/amazon-vpc-cni-k8s/master/config/v1.6/aws-k8s-cni.yaml

---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: aws-node
rules:
- apiGroups:
  - crd.k8s.amazonaws.com
  resources:
  - "*"
  - namespaces
  verbs:
  - "*"
- apiGroups: [""]
  resources:
  - pods
  - nodes
  - namespaces
  verbs: ["list", "watch", "get"]
- apiGroups: ["extensions"]
  resources:
  - daemonsets
  verbs: ["list", "watch"]

---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: aws-node
  namespace: kube-system

---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: aws-node
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: aws-node
subjects:
- kind: ServiceAccount
  name: aws-node
  namespace: kube-system

---
kind: DaemonSet
apiVersion: apps/v1
metadata:
  name: aws-node
  namespace: kube-system
  labels:
    k8s-app: aws-node
spec:
  updateStrategy:
    type: RollingUpdate
    rollingUpdate:
      maxUnavailable: "10%"
  selector:
    matchLabels:
      k8s-app: aws-node
  template:
    metadata:
      labels:
        k8s-app: aws-node
    spec:
      priorityClassName: system-node-critical
      affinity:
        nodeAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
            - matchExpressions:
              - key: "beta.kubernetes.io/os"
                operator: In
                values:
                - linux
              - key: "beta.kubernetes.io/arch"
                operator: In
                values:
                - amd64
              - key: eks.amazonaws.com/compute-type
                operator: NotIn
                values:
                - fargate
      serviceAccountName: aws-node
      hostNetwork: true
      tolerations:
      - operator: Exists
      containers:
      - image: "{{- or .Networking.AmazonVPC.ImageName "602401143452.dkr.ecr.us-west-2.amazonaws.com/amazon-k8s-cni:v1.6.0" }}"
        imagePullPolicy: Always
        ports:
        - containerPort: 61678
          name: metrics
        name: aws-node
        readinessProbe:
          exec:
            command: ["/app/grpc-health-probe", "-addr=:50051"]
          initialDelaySeconds: 35
        livenessProbe:
          exec:
            command: ["/app/grpc-health-probe", "-addr=:50051"]
          initialDelaySeconds: 35
        env:
        - name: CLUSTER_NAME
          value: {{ ClusterName }}
        - name: AWS_VPC_K8S_CNI_LOGLEVEL
          value: DEBUG
        - name: AWS_VPC_K8S_CNI_VETHPREFIX
          value: eni
        - name: AWS_VPC_ENI_MTU
          value: "9001"
        - name: MY_NODE_NAME
          valueFrom:
            fieldRef:
              fieldPath: spec.nodeName
        {{- range .Networking.AmazonVPC.Env }}
        - name: {{ .Name }}
          value: "{{ .Value }}"
        {{- end }}
        resources:
          requests:
            cpu: 10m
        securityContext:
          privileged: true
        volumeMounts:
        - mountPath: /host/opt/cni/bin
          name: cni-bin-dir
        - mountPath: /host/etc/cni/net.d
          name: cni-net-dir
        - mountPath: /host/var/log
          name: log-dir
        - mountPath: /var/run/docker.sock
          name: dockersock
        - mountPath: /var/run/dockershim.sock
          name: dockershim
      volumes:
      - name: cni-bin-dir
        hostPath:
          path: /opt/cni/bin
      - name: cni-net-dir
        hostPath:
          path: /etc/cni/net.d
      - name: log-dir
        hostPath:
          path: /var/log
      - name: dockersock
        hostPath:
          path: /var/run/docker.sock
      - name: dockershim
        hostPath:
          path: /var/run/dockershim.sock

---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: eniconfigs.crd.k8s.amazonaws.com
spec:
  scope: Cluster
  group: crd.k8s.amazonaws.com
  versions:
  - name: v1alpha1
    served: true
    storage: true
  names:
    plural: eniconfigs
    singular: eniconfig
    kind: ENIConfig
`)

func cloudupResourcesAddonsNetworkingAmazonVpcRoutedEniK8s116YamlTemplateBytes() ([]byte, error) {
	return _cloudupResourcesAddonsNetworkingAmazonVpcRoutedEniK8s116YamlTemplate, nil
}

func cloudupResourcesAddonsNetworkingAmazonVpcRoutedEniK8s116YamlTemplate() (*asset, error) {
	bytes, err := cloudupResourcesAddonsNetworkingAmazonVpcRoutedEniK8s116YamlTemplateBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "cloudup/resources/addons/networking.amazon-vpc-routed-eni/k8s-1.16.yaml.template", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _cloudupResourcesAddonsNetworkingAmazonVpcRoutedEniK8s18YamlTemplate = []byte(`# Vendored from https://github.com/aws/amazon-vpc-cni-k8s/blob/v1.3.3/config/v1.3/aws-k8s-cni.yaml

apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: aws-node
rules:
- apiGroups:
  - crd.k8s.amazonaws.com
  resources:
  - "*"
  - namespaces
  verbs:
  - "*"
- apiGroups: [""]
  resources:
  - pods
  - nodes
  - namespaces
  verbs: ["list", "watch", "get"]
- apiGroups: ["extensions"]
  resources:
  - daemonsets
  verbs: ["list", "watch"]
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: aws-node
  namespace: kube-system
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: aws-node
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: aws-node
subjects:
- kind: ServiceAccount
  name: aws-node
  namespace: kube-system
---
kind: DaemonSet
apiVersion: extensions/v1beta1
metadata:
  name: aws-node
  namespace: kube-system
  labels:
    k8s-app: aws-node
spec:
  updateStrategy:
    type: RollingUpdate
  selector:
    matchLabels:
      k8s-app: aws-node
  template:
    metadata:
      labels:
        k8s-app: aws-node
      annotations:
        scheduler.alpha.kubernetes.io/critical-pod: ''
    spec:
      serviceAccountName: aws-node
      hostNetwork: true
      tolerations:
      - operator: Exists
      containers:
      - image: "{{- or .Networking.AmazonVPC.ImageName "602401143452.dkr.ecr.us-west-2.amazonaws.com/amazon-k8s-cni:1.3.3" }}"
        ports:
        - containerPort: 61678
          name: metrics
        name: aws-node
        env:
          - name: CLUSTER_NAME
            value: {{ ClusterName }}
          - name: AWS_VPC_K8S_CNI_LOGLEVEL
            value: DEBUG
          - name: MY_NODE_NAME
            valueFrom:
              fieldRef:
                fieldPath: spec.nodeName
          - name: WATCH_NAMESPACE
            valueFrom:
              fieldRef:
                fieldPath: metadata.namespace
          {{- range .Networking.AmazonVPC.Env }}
          - name: {{ .Name }}
            value: "{{ .Value }}"
          {{- end }}
        resources:
          requests:
            cpu: 10m
        securityContext:
          privileged: true
        volumeMounts:
        - mountPath: /host/opt/cni/bin
          name: cni-bin-dir
        - mountPath: /host/etc/cni/net.d
          name: cni-net-dir
        - mountPath: /host/var/log
          name: log-dir
        - mountPath: /var/run/docker.sock
          name: dockersock
      volumes:
      - name: cni-bin-dir
        hostPath:
          path: /opt/cni/bin
      - name: cni-net-dir
        hostPath:
          path: /etc/cni/net.d
      - name: log-dir
        hostPath:
          path: /var/log
      - name: dockersock
        hostPath:
          path: /var/run/docker.sock
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: eniconfigs.crd.k8s.amazonaws.com
spec:
  scope: Cluster
  group: crd.k8s.amazonaws.com
  version: v1alpha1
  names:
    plural: eniconfigs
    singular: eniconfig
    kind: ENIConfig

---

kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: k8s-ec2-srcdst
  labels:
    role.kubernetes.io/networking: "1"
rules:
- apiGroups:
  - ""
  resources:
  - nodes
  verbs:
  - get
  - list
  - watch
  - update
  - patch

---

apiVersion: v1
kind: ServiceAccount
metadata:
  name: k8s-ec2-srcdst
  namespace: kube-system
  labels:
    role.kubernetes.io/networking: "1"
---

kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: k8s-ec2-srcdst
  labels:
    role.kubernetes.io/networking: "1"
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: k8s-ec2-srcdst
subjects:
- kind: ServiceAccount
  name: k8s-ec2-srcdst
  namespace: kube-system

---

apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: k8s-ec2-srcdst
  namespace: kube-system
  labels:
    k8s-app: k8s-ec2-srcdst
    role.kubernetes.io/networking: "1"
spec:
  replicas: 1
  selector:
    matchLabels:
      k8s-app: k8s-ec2-srcdst
  template:
    metadata:
      labels:
        k8s-app: k8s-ec2-srcdst
        role.kubernetes.io/networking: "1"
      annotations:
        scheduler.alpha.kubernetes.io/critical-pod: ''
    spec:
      hostNetwork: true
      tolerations:
      - key: node-role.kubernetes.io/master
        effect: NoSchedule
      - key: CriticalAddonsOnly
        operator: Exists
      serviceAccountName: k8s-ec2-srcdst
      containers:
        - image: ottoyiu/k8s-ec2-srcdst:v0.2.0-3-gc0c26eca
          name: k8s-ec2-srcdst
          resources:
            requests:
              cpu: 10m
              memory: 64Mi
          env:
            - name: AWS_REGION
              value: {{ Region }}
          volumeMounts:
            - name: ssl-certs
              mountPath: "/etc/ssl/certs/ca-certificates.crt"
              readOnly: true
          imagePullPolicy: "Always"
      volumes:
        - name: ssl-certs
          hostPath:
            path: "/etc/ssl/certs/ca-certificates.crt"
      nodeSelector:
        node-role.kubernetes.io/master: ""
`)

func cloudupResourcesAddonsNetworkingAmazonVpcRoutedEniK8s18YamlTemplateBytes() ([]byte, error) {
	return _cloudupResourcesAddonsNetworkingAmazonVpcRoutedEniK8s18YamlTemplate, nil
}

func cloudupResourcesAddonsNetworkingAmazonVpcRoutedEniK8s18YamlTemplate() (*asset, error) {
	bytes, err := cloudupResourcesAddonsNetworkingAmazonVpcRoutedEniK8s18YamlTemplateBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "cloudup/resources/addons/networking.amazon-vpc-routed-eni/k8s-1.8.yaml.template", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _cloudupResourcesAddonsNetworkingCiliumIoK8s112YamlTemplate = []byte(`apiVersion: v1
kind: ConfigMap
metadata:
  name: cilium-config
  namespace: kube-system
  labels:
    role.kubernetes.io/networking: "1"
data:
{{ with .Networking.Cilium }}

{{- if .EtcdManaged }}
  kvstore: etcd
  kvstore-opt: '{"etcd.config": "/var/lib/etcd-config/etcd.config", "etcd.operator": "true"}'

  etcd-config: |-
    ---
    endpoints:
      - https://cilium-etcd-client.kube-system.svc:2379

    trusted-ca-file: '/var/lib/etcd-secrets/etcd-client-ca.crt'
    key-file: '/var/lib/etcd-secrets/etcd-client.key'
    cert-file: '/var/lib/etcd-secrets/etcd-client.crt'
{{ end }}

  # Identity allocation mode selects how identities are shared between cilium
  # nodes by setting how they are stored. The options are "crd" or "kvstore".
  # - "crd" stores identities in kubernetes as CRDs (custom resource definition).
  #   These can be queried with:
  #     kubectl get ciliumid
  # - "kvstore" stores identities in a kvstore, etcd or consul, that is
  #   configured below. Cilium versions before 1.6 supported only the kvstore
  #   backend. Upgrades from these older cilium versions should continue using
  #   the kvstore by commenting out the identity-allocation-mode below, or
  #   setting it to "kvstore".
  identity-allocation-mode: crd
  # If you want to run cilium in debug mode change this value to true
  debug: "{{- if .Debug -}}true{{- else -}}false{{- end -}}"
  {{ if .EnablePrometheusMetrics }}
  # If you want metrics enabled in all of your Cilium agents, set the port for
  # which the Cilium agents will have their metrics exposed.
  # This option deprecates the "prometheus-serve-addr" in the
  # "cilium-metrics-config" ConfigMap
  # NOTE that this will open the port on ALL nodes where Cilium pods are
  # scheduled.
  prometheus-serve-addr: ":{{- or .AgentPrometheusPort "9090" }}"
  {{ end }}
  # Enable IPv4 addressing. If enabled, all endpoints are allocated an IPv4
  # address.
  enable-ipv4: "{{- if or  (.EnableIpv4) (and (not (.EnableIpv4)) (not (.EnableIpv6))) -}}true{{- else -}}false{{- end -}}"
  # Enable IPv6 addressing. If enabled, all endpoints are allocated an IPv6
  # address.
  enable-ipv6: "{{- if .EnableIpv6 -}}true{{- else -}}false{{- end -}}"
  # If you want cilium monitor to aggregate tracing for packets, set this level
  # to "low", "medium", or "maximum". The higher the level, the less packets
  # that will be seen in monitor output.
  monitor-aggregation: "{{- if eq .MonitorAggregation "" -}}medium{{- else -}}{{ .MonitorAggregation }}{{- end -}}"
  # ct-global-max-entries-* specifies the maximum number of connections
  # supported across all endpoints, split by protocol: tcp or other. One pair
  # of maps uses these values for IPv4 connections, and another pair of maps
  # use these values for IPv6 connections.
  #
  # If these values are modified, then during the next Cilium startup the
  # tracking of ongoing connections may be disrupted. This may lead to brief
  # policy drops or a change in loadbalancing decisions for a connection.
  #
  # For users upgrading from Cilium 1.2 or earlier, to minimize disruption
  # during the upgrade process, comment out these options.
  bpf-ct-global-tcp-max: "{{- if eq .BPFCTGlobalTCPMax 0 -}}524288{{- else -}}{{ .BPFCTGlobalTCPMax}}{{- end -}}"
  bpf-ct-global-any-max: "{{- if eq .BPFCTGlobalAnyMax 0 -}}262144{{- else -}}{{ .BPFCTGlobalAnyMax}}{{- end -}}"

  # Pre-allocation of map entries allows per-packet latency to be reduced, at
  # the expense of up-front memory allocation for the entries in the maps. The
  # default value below will minimize memory usage in the default installation;
  # users who are sensitive to latency may consider setting this to "true".
  #
  # This option was introduced in Cilium 1.4. Cilium 1.3 and earlier ignore
  # this option and behave as though it is set to "true".
  #
  # If this value is modified, then during the next Cilium startup the restore
  # of existing endpoints and tracking of ongoing connections may be disrupted.
  # This may lead to policy drops or a change in loadbalancing decisions for a
  # connection for some time. Endpoints may need to be recreated to restore
  # connectivity.
  #
  # If this option is set to "false" during an upgrade from 1.3 or earlier to
  # 1.4 or later, then it may cause one-time disruptions during the upgrade.
  preallocate-bpf-maps: "{{- if .PreallocateBPFMaps -}}true{{- else -}}false{{- end -}}"
  # Regular expression matching compatible Istio sidecar istio-proxy
  # container image names
  sidecar-istio-proxy-image: "{{- if eq .SidecarIstioProxyImage "" -}}cilium/istio_proxy{{- else -}}{{ .SidecarIstioProxyImage }}{{- end -}}"
  # Encapsulation mode for communication between nodes
  # Possible values:
  #   - disabled
  #   - vxlan (default)
  #   - geneve
  tunnel: "{{- if eq .Tunnel "" -}}vxlan{{- else -}}{{ .Tunnel }}{{- end -}}"

  # Name of the cluster. Only relevant when building a mesh of clusters.
  cluster-name: "{{- if eq .ClusterName "" -}}default{{- else -}}{{ .ClusterName}}{{- end -}}"

  # DNS response code for rejecting DNS requests,
  # available options are "nameError" and "refused"
  tofqdns-dns-reject-response-code: "{{- if eq .ToFqdnsDNSRejectResponseCode  "" -}}refused{{- else -}}{{ .ToFqdnsDNSRejectResponseCode }}{{- end -}}"
  # This option is disabled by default starting from version 1.4.x in favor
  # of a more powerful DNS proxy-based implementation, see [0] for details.
  # Enable this option if you want to use FQDN policies but do not want to use
  # the DNS proxy.
  #
  # To ease upgrade, users may opt to set this option to "true".
  # Otherwise please refer to the Upgrade Guide [1] which explains how to
  # prepare policy rules for upgrade.
  #
  # [0] http://docs.cilium.io/en/stable/policy/language/#dns-based
  # [1] http://docs.cilium.io/en/stable/install/upgrade/#changes-that-may-require-action
  tofqdns-enable-poller: "{{- if .ToFqdnsEnablePoller -}}true{{- else -}}false{{- end -}}"
  # wait-bpf-mount makes init container wait until bpf filesystem is mounted
  wait-bpf-mount: "false"
  # Enable fetching of container-runtime specific metadata
  #
  # By default, the Kubernetes pod and namespace labels are retrieved and
  # associated with endpoints for identification purposes. By integrating
  # with the container runtime, container runtime specific labels can be
  # retrieved, such labels will be prefixed with container:
  #
  # CAUTION: The container runtime labels can include information such as pod
  # annotations which may result in each pod being associated a unique set of
  # labels which can result in excessive security identities being allocated.
  # Please review the labels filter when enabling container runtime labels.
  #
  # Supported values:
  # - containerd
  # - crio
  # - docker
  # - none
  # - auto (automatically detect the container runtime)
  #
  container-runtime: "{{- if eq .ContainerRuntimeLabels "" -}}none{{- else -}}{{ .ContainerRuntimeLabels }}{{- end -}}"
  masquerade: "{{- if .DisableMasquerade -}}false{{- else -}}true{{- end -}}"
  install-iptables-rules: "{{- if .IPTablesRulesNoinstall -}}false{{- else -}}true{{- end -}}"
  auto-direct-node-routes: "{{- if .AutoDirectNodeRoutes -}}true{{- else -}}false{{- end -}}"
  enable-node-port: "{{- if .EnableNodePort -}}true{{- else -}}false{{- end -}}"
  kube-proxy-replacement: "{{- if .EnableNodePort -}}strict{{- else -}}partial{{- end -}}"
  enable-remote-node-identity: "{{- if .EnableRemoteNodeIdentity -}}true{{- else -}}false{{- end -}}"
  {{ with .Ipam }}
  ipam: {{ . }}
  {{ if eq . "eni" }}
  enable-endpoint-routes: "true"
  auto-create-cilium-node-resource: "true"
  blacklist-conflicting-routes: "false"
  {{ end }}
  {{ end }}
{{ end }} # With .Networking.Cilium end
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: cilium
  namespace: kube-system
  labels:
    role.kubernetes.io/networking: "1"
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: cilium-operator
  namespace: kube-system
  labels:
    role.kubernetes.io/networking: "1"
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: cilium
  labels:
    role.kubernetes.io/networking: "1"
rules:
- apiGroups:
  - networking.k8s.io
  resources:
  - networkpolicies
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - discovery.k8s.io
  resources:
  - endpointslices
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - ""
  resources:
  - namespaces
  - services
  - nodes
  - endpoints
  - componentstatuses
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - ""
  resources:
  - pods
  - nodes
  verbs:
  - get
  - list
  - watch
  - update
- apiGroups:
  - ""
  resources:
  - nodes
  - nodes/status
  verbs:
  - patch
- apiGroups:
  - extensions
  resources:
  - ingresses
  verbs:
  - create
  - get
  - list
  - watch
- apiGroups:
  - apiextensions.k8s.io
  resources:
  - customresourcedefinitions
  verbs:
  - create
  - get
  - list
  - watch
  - update
- apiGroups:
  - cilium.io
  resources:
  - ciliumnetworkpolicies
  - ciliumnetworkpolicies/status
  - ciliumclusterwidenetworkpolicies
  - ciliumclusterwidenetworkpolicies/status
  - ciliumendpoints
  - ciliumendpoints/status
  - ciliumnodes
  - ciliumnodes/status
  - ciliumidentities
  - ciliumidentities/status
  verbs:
  - '*'
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: cilium-operator
  labels:
    role.kubernetes.io/networking: "1"
rules:
- apiGroups:
  - ""
  resources:
  # to automatically delete [core|kube]dns pods so that are starting to being
  # managed by Cilium
  - pods
  verbs:
  - get
  - list
  - watch
  - delete
- apiGroups:
  - discovery.k8s.io
  resources:
  - endpointslices
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - ""
  resources:
  # to automatically read from k8s and import the node's pod CIDR to cilium's
  # etcd so all nodes know how to reach another pod running in a different
  # node.
  - nodes
  # to perform the translation of a CNP that contains ` + "`" + `ToGroup` + "`" + ` to its endpoints
  - services
  - endpoints
  # to check apiserver connectivity
  - namespaces
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - cilium.io
  resources:
  - ciliumnetworkpolicies
  - ciliumnetworkpolicies/status
  - ciliumclusterwidenetworkpolicies
  - ciliumclusterwidenetworkpolicies/status
  - ciliumendpoints
  - ciliumendpoints/status
  - ciliumnodes
  - ciliumnodes/status
  - ciliumidentities
  - ciliumidentities/status
  verbs:
  - '*'
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: cilium
  labels:
    role.kubernetes.io/networking: "1"
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cilium
subjects:
- kind: ServiceAccount
  name: cilium
  namespace: kube-system
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: cilium-operator
  labels:
    role.kubernetes.io/networking: "1"
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cilium-operator
subjects:
- kind: ServiceAccount
  name: cilium-operator
  namespace: kube-system
---
apiVersion: apps/v1
kind: DaemonSet
metadata:
  labels:
    k8s-app: cilium
    kubernetes.io/cluster-service: "true"
    role.kubernetes.io/networking: "1"
  name: cilium
  namespace: kube-system
spec:
  selector:
    matchLabels:
      k8s-app: cilium
      kubernetes.io/cluster-service: "true"
  template:
    metadata:
      annotations:
        # This annotation plus the CriticalAddonsOnly toleration makes
        # cilium to be a critical pod in the cluster, which ensures cilium
        # gets priority scheduling.
        # https://kubernetes.io/docs/tasks/administer-cluster/guaranteed-scheduling-critical-addon-pods/
        scheduler.alpha.kubernetes.io/critical-pod: ""
      labels:
        k8s-app: cilium
        kubernetes.io/cluster-service: "true"
    spec:
      containers:
      - args:
        - --config-dir=/tmp/cilium/config-map
        command:
        - cilium-agent
        env:
        - name: K8S_NODE_NAME
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: spec.nodeName
        - name: CILIUM_K8S_NAMESPACE
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: metadata.namespace
        - name: CILIUM_FLANNEL_MASTER_DEVICE
          valueFrom:
            configMapKeyRef:
              key: flannel-master-device
              name: cilium-config
              optional: true
        - name: CILIUM_FLANNEL_UNINSTALL_ON_EXIT
          valueFrom:
            configMapKeyRef:
              key: flannel-uninstall-on-exit
              name: cilium-config
              optional: true
        - name: CILIUM_CLUSTERMESH_CONFIG
          value: /var/lib/cilium/clustermesh/
        - name: CILIUM_CNI_CHAINING_MODE
          valueFrom:
            configMapKeyRef:
              key: cni-chaining-mode
              name: cilium-config
              optional: true
        - name: CILIUM_CUSTOM_CNI_CONF
          valueFrom:
            configMapKeyRef:
              key: custom-cni-conf
              name: cilium-config
              optional: true
        - name: KUBERNETES_SERVICE_HOST
          value: "{{.MasterInternalName}}"
        - name: KUBERNETES_SERVICE_PORT
          value: "443"
        {{ with .Networking.Cilium.EnablePolicy }}
        - name: CILIUM_ENABLE_POLICY
          value: {{ . }}
        {{ end }}
{{ with .Networking.Cilium }}
        image: "docker.io/cilium/cilium:{{- or .Version "v1.7.1" }}"
        imagePullPolicy: IfNotPresent
        lifecycle:
          postStart:
            exec:
              command:
              - /cni-install.sh
          preStop:
            exec:
              command:
              - /cni-uninstall.sh
        livenessProbe:
          exec:
            command:
            - cilium
            - status
            - --brief
          failureThreshold: 10
          # The initial delay for the liveness probe is intentionally large to
          # avoid an endless kill & restart cycle if in the event that the initial
          # bootstrapping takes longer than expected.
          initialDelaySeconds: 120
          periodSeconds: 30
          successThreshold: 1
          timeoutSeconds: 5
        name: cilium-agent
        {{ if .EnablePrometheusMetrics }}
        ports:
        - containerPort: {{ or .AgentPrometheusPort "9090" }}
          hostPort: {{ or .AgentPrometheusPort "9090" }}
          name: prometheus
          protocol: TCP
        {{ end }}
        readinessProbe:
          exec:
            command:
            - cilium
            - status
            - --brief
          failureThreshold: 3
          initialDelaySeconds: 5
          periodSeconds: 30
          successThreshold: 1
          timeoutSeconds: 5
        securityContext:
          capabilities:
            add:
            - NET_ADMIN
            - SYS_MODULE
          privileged: true
        volumeMounts:
        - mountPath: /sys/fs/bpf
          name: bpf-maps
          mountPropagation: HostToContainer
        - mountPath: /var/run/cilium
          name: cilium-run
        - mountPath: /host/opt/cni/bin
          name: cni-path
        - mountPath: /host/etc/cni/net.d
          name: etc-cni-netd
{{ if .EtcdManaged }}
        - mountPath: /var/lib/etcd-config
          name: etcd-config-path
          readOnly: true
        - mountPath: /var/lib/etcd-secrets
          name: etcd-secrets
          readOnly: true
{{ end }}
        - mountPath: /var/lib/cilium/clustermesh
          name: clustermesh-secrets
          readOnly: true
        - mountPath: /tmp/cilium/config-map
          name: cilium-config-path
          readOnly: true
          # Needed to be able to load kernel modules
        - mountPath: /lib/modules
          name: lib-modules
          readOnly: true
        - mountPath: /run/xtables.lock
          name: xtables-lock
      hostNetwork: true
      initContainers:
      - command:
        - /init-container.sh
        env:
        - name: CILIUM_ALL_STATE
          valueFrom:
            configMapKeyRef:
              key: clean-cilium-state
              name: cilium-config
              optional: true
        - name: CILIUM_BPF_STATE
          valueFrom:
            configMapKeyRef:
              key: clean-cilium-bpf-state
              name: cilium-config
              optional: true
        - name: CILIUM_WAIT_BPF_MOUNT
          valueFrom:
            configMapKeyRef:
              key: wait-bpf-mount
              name: cilium-config
              optional: true
        image: "docker.io/cilium/cilium:{{- or .Version "v1.7.1" }}"
## end of ` + "`" + `with .Networking.Cilium` + "`" + `
#{{ end }}
        imagePullPolicy: IfNotPresent
        name: clean-cilium-state
        securityContext:
          capabilities:
            add:
            - NET_ADMIN
          privileged: true
        volumeMounts:
        - mountPath: /sys/fs/bpf
          name: bpf-maps
        - mountPath: /var/run/cilium
          name: cilium-run
      priorityClassName: system-node-critical
      restartPolicy: Always
      serviceAccount: cilium
      serviceAccountName: cilium
      terminationGracePeriodSeconds: 1
      tolerations:
      - operator: Exists
      volumes:
        # To keep state between restarts / upgrades
      - hostPath:
          path: /var/run/cilium
          type: DirectoryOrCreate
        name: cilium-run
        # To keep state between restarts / upgrades for bpf maps
      - hostPath:
          path: /sys/fs/bpf
          type: DirectoryOrCreate
        name: bpf-maps
      # To install cilium cni plugin in the host
      - hostPath:
          path:  /opt/cni/bin
          type: DirectoryOrCreate
        name: cni-path
        # To install cilium cni configuration in the host
      - hostPath:
          path: /etc/cni/net.d
          type: DirectoryOrCreate
        name: etc-cni-netd
        # To be able to load kernel modules
      - hostPath:
          path: /lib/modules
        name: lib-modules
        # To access iptables concurrently with other processes (e.g. kube-proxy)
      - hostPath:
          path: /run/xtables.lock
          type: FileOrCreate
        name: xtables-lock
        # To read the clustermesh configuration
{{- if .Networking.Cilium.EtcdManaged }}
        # To read the etcd config stored in config maps
      - configMap:
          defaultMode: 420
          items:
          - key: etcd-config
            path: etcd.config
          name: cilium-config
        name: etcd-config-path
        # To read the Cilium etcd secrets in case the user might want to use TLS
      - name: etcd-secrets
        secret:
          defaultMode: 420
          optional: true
          secretName: cilium-etcd-secrets
{{- end }}
      - name: clustermesh-secrets
        secret:
          defaultMode: 420
          optional: true
          secretName: cilium-clustermesh
        # To read the configuration from the config map
      - configMap:
          name: cilium-config
        name: cilium-config-path
  updateStrategy:
    rollingUpdate:
      maxUnavailable: 2
    type: RollingUpdate
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    io.cilium/app: operator
    name: cilium-operator
    role.kubernetes.io/networking: "1"
  name: cilium-operator
  namespace: kube-system
spec:
  replicas: 1
  selector:
    matchLabels:
      io.cilium/app: operator
      name: cilium-operator
  strategy:
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 1
    type: RollingUpdate
  template:
    metadata:
      labels:
        io.cilium/app: operator
        name: cilium-operator
    spec:
      containers:
      - args:
        - --debug=$(CILIUM_DEBUG)
        - --identity-allocation-mode=$(CILIUM_IDENTITY_ALLOCATION_MODE)
{{ with .Networking.Cilium }}
        {{ if .EnablePrometheusMetrics }}
        - --enable-metrics
        {{ end }}
{{ end }}
        command:
        - cilium-operator
        env:
        - name: CILIUM_IDENTITY_ALLOCATION_MODE
          valueFrom:
            configMapKeyRef:
              key: identity-allocation-mode
              name: cilium-config
              optional: true
        - name: CILIUM_K8S_NAMESPACE
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: metadata.namespace
        - name: K8S_NODE_NAME
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: spec.nodeName
        - name: CILIUM_DEBUG
          valueFrom:
            configMapKeyRef:
              key: debug
              name: cilium-config
              optional: true
        - name: CILIUM_CLUSTER_NAME
          valueFrom:
            configMapKeyRef:
              key: cluster-name
              name: cilium-config
              optional: true
        - name: CILIUM_CLUSTER_ID
          valueFrom:
            configMapKeyRef:
              key: cluster-id
              name: cilium-config
              optional: true
        - name: CILIUM_IPAM
          valueFrom:
            configMapKeyRef:
              key: ipam
              name: cilium-config
              optional: true
        - name: CILIUM_DISABLE_ENDPOINT_CRD
          valueFrom:
            configMapKeyRef:
              key: disable-endpoint-crd
              name: cilium-config
              optional: true
        - name: CILIUM_KVSTORE
          valueFrom:
            configMapKeyRef:
              key: kvstore
              name: cilium-config
              optional: true
        - name: CILIUM_KVSTORE_OPT
          valueFrom:
            configMapKeyRef:
              key: kvstore-opt
              name: cilium-config
              optional: true
        - name: AWS_ACCESS_KEY_ID
          valueFrom:
            secretKeyRef:
              key: AWS_ACCESS_KEY_ID
              name: cilium-aws
              optional: true
        - name: AWS_SECRET_ACCESS_KEY
          valueFrom:
            secretKeyRef:
              key: AWS_SECRET_ACCESS_KEY
              name: cilium-aws
              optional: true
        - name: AWS_DEFAULT_REGION
          valueFrom:
            secretKeyRef:
              key: AWS_DEFAULT_REGION
              name: cilium-aws
              optional: true
        - name: KUBERNETES_SERVICE_HOST
          value: "{{.MasterInternalName}}"
        - name: KUBERNETES_SERVICE_PORT
          value: "443"
{{ with .Networking.Cilium }}
        image: "docker.io/cilium/operator:{{- if eq .Version "" -}}v1.7.1{{- else -}}{{ .Version }}{{- end -}}"
        imagePullPolicy: IfNotPresent
        name: cilium-operator
        {{ if .EnablePrometheusMetrics }}
        ports:
        - containerPort: 6942
          hostPort: 6942
          name: prometheus
          protocol: TCP
        {{ end }}
        livenessProbe:
          httpGet:
            host: "127.0.0.1"
            path: /healthz
            port: 9234
            scheme: HTTP
          initialDelaySeconds: 60
          periodSeconds: 10
          timeoutSeconds: 3
{{- if .EtcdManaged }}
        volumeMounts:
        - mountPath: /var/lib/etcd-config
          name: etcd-config-path
          readOnly: true
        - mountPath: /var/lib/etcd-secrets
          name: etcd-secrets
          readOnly: true
{{- end }}
      hostNetwork: true
      priorityClassName: system-cluster-critical
      restartPolicy: Always
      serviceAccount: cilium-operator
      serviceAccountName: cilium-operator
{{- if .EtcdManaged }}
      volumes:
      # To read the etcd config stored in config maps
      - configMap:
          defaultMode: 420
          items:
          - key: etcd-config
            path: etcd.config
          name: cilium-config
        name: etcd-config-path
        # To read the k8s etcd secrets in case the user might want to use TLS
      - name: etcd-secrets
        secret:
          defaultMode: 420
          optional: true
          secretName: cilium-etcd-secrets
{{- end }}

      {{ if eq .Ipam "eni" }}
      nodeSelector:
        node-role.kubernetes.io/master: ""
      tolerations:
      - effect: NoSchedule
        key: node-role.kubernetes.io/master
      - effect: NoExecute
        key: node.kubernetes.io/not-ready
        operator: Exists
        tolerationSeconds: 300
      - effect: NoExecute
        key: node.kubernetes.io/unreachable
        operator: Exists
        tolerationSeconds: 300
      {{ end }}
{{ end }}

{{ if .Networking.Cilium.EtcdManaged }}
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  labels:
    role.kubernetes.io/networking: "1"
  name: cilium-etcd-operator
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cilium-etcd-operator
subjects:
- kind: ServiceAccount
  name: cilium-etcd-operator
  namespace: kube-system
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: cilium-etcd-operator
rules:
- apiGroups:
  - etcd.database.coreos.com
  resources:
  - etcdclusters
  verbs:
  - get
  - delete
  - create
  - update
- apiGroups:
  - apiextensions.k8s.io
  resources:
  - customresourcedefinitions
  verbs:
  - delete
  - get
  - create
- apiGroups:
  - ""
  resources:
  - deployments
  verbs:
  - delete
  - create
  - get
  - update
- apiGroups:
  - ""
  resources:
  - pods
  verbs:
  - list
  - get
  - delete
- apiGroups:
  - apps
  resources:
  - deployments
  verbs:
  - delete
  - create
  - get
  - update
- apiGroups:
  - ""
  resources:
  - componentstatuses
  verbs:
  - get
- apiGroups:
  - extensions
  resources:
  - deployments
  verbs:
  - delete
  - create
  - get
  - update
- apiGroups:
  - ""
  resources:
  - secrets
  verbs:
  - get
  - create
  - delete
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    io.cilium/app: etcd-operator
    name: cilium-etcd-operator
  name: cilium-etcd-operator
  namespace: kube-system
spec:
  replicas: 1
  selector:
    matchLabels:
      io.cilium/app: etcd-operator
      name: cilium-etcd-operator
  strategy:
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 1
    type: RollingUpdate
  template:
    metadata:
      labels:
        io.cilium/app: etcd-operator
        name: cilium-etcd-operator
    spec:
      containers:
      - command:
        - /usr/bin/cilium-etcd-operator
        env:
        - name: CILIUM_ETCD_OPERATOR_CLUSTER_DOMAIN
          value: "{{ $.ClusterDNSDomain }}"
        - name: CILIUM_ETCD_OPERATOR_ETCD_CLUSTER_SIZE
          value: "3"
        - name: CILIUM_ETCD_OPERATOR_NAMESPACE
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: metadata.namespace
        - name: CILIUM_ETCD_OPERATOR_POD_NAME
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: metadata.name
        - name: CILIUM_ETCD_OPERATOR_POD_UID
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: metadata.uid
        - name: CILIUM_ETCD_META_ETCD_AUTO_COMPACTION_MODE
          value: "revision"
        - name: CILIUM_ETCD_META_ETCD_AUTO_COMPACTION_RETENTION
          value: "25000"
        image: "cilium/cilium-etcd-operator:v2.0.7"
        name: cilium-etcd-operator
      dnsPolicy: ClusterFirst
      hostNetwork: true
      priorityClassName: system-cluster-critical
      restartPolicy: Always
      serviceAccount: cilium-etcd-operator
      serviceAccountName: cilium-etcd-operator
      tolerations:
      - operator: Exists
---
apiVersion: v1
kind: ServiceAccount
metadata:
  labels:
    role.kubernetes.io/networking: "1"
  name: cilium-etcd-operator
  namespace: kube-system
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  labels:
    role.kubernetes.io/networking: "1"
  name: etcd-operator
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: etcd-operator
subjects:
- kind: ServiceAccount
  name: cilium-etcd-sa
  namespace: kube-system
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  labels:
    role.kubernetes.io/networking: "1"
  name: etcd-operator
rules:
- apiGroups:
  - etcd.database.coreos.com
  resources:
  - etcdclusters
  - etcdbackups
  - etcdrestores
  verbs:
  - '*'
- apiGroups:
  - apiextensions.k8s.io
  resources:
  - customresourcedefinitions
  verbs:
  - '*'
- apiGroups:
  - ""
  resources:
  - pods
  - services
  - endpoints
  - persistentvolumeclaims
  - events
  - deployments
  verbs:
  - '*'
- apiGroups:
  - apps
  resources:
  - deployments
  verbs:
  - '*'
- apiGroups:
  - extensions
  resources:
  - deployments
  verbs:
  - create
  - get
  - list
  - patch
  - update
- apiGroups:
  - ""
  resources:
  - secrets
  verbs:
  - get
---
apiVersion: v1
kind: ServiceAccount
metadata:
  labels:
    role.kubernetes.io/networking: "1"
  name: cilium-etcd-sa
  namespace: kube-system
{{ end }}
`)

func cloudupResourcesAddonsNetworkingCiliumIoK8s112YamlTemplateBytes() ([]byte, error) {
	return _cloudupResourcesAddonsNetworkingCiliumIoK8s112YamlTemplate, nil
}

func cloudupResourcesAddonsNetworkingCiliumIoK8s112YamlTemplate() (*asset, error) {
	bytes, err := cloudupResourcesAddonsNetworkingCiliumIoK8s112YamlTemplateBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "cloudup/resources/addons/networking.cilium.io/k8s-1.12.yaml.template", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _cloudupResourcesAddonsNetworkingCiliumIoK8s17YamlTemplate = []byte(`apiVersion: v1
kind: ConfigMap
metadata:
  name: cilium-config
  namespace: kube-system
  labels:
    role.kubernetes.io/networking: "1"
data:
{{ with .Networking.Cilium }}
  # Identity allocation mode selects how identities are shared between cilium
  # nodes by setting how they are stored. The options are "crd" or "kvstore".
  # - "crd" stores identities in kubernetes as CRDs (custom resource definition).
  #   These can be queried with:
  #     kubectl get ciliumid
  # - "kvstore" stores identities in a kvstore, etcd or consul, that is
  #   configured below. Cilium versions before 1.6 supported only the kvstore
  #   backend. Upgrades from these older cilium versions should continue using
  #   the kvstore by commenting out the identity-allocation-mode below, or
  #   setting it to "kvstore".
  identity-allocation-mode: crd
  # If you want to run cilium in debug mode change this value to true
  debug: "{{- if .Debug -}}true{{- else -}}false{{- end -}}"
  {{ if .EnablePrometheusMetrics }}
  # If you want metrics enabled in all of your Cilium agents, set the port for
  # which the Cilium agents will have their metrics exposed.
  # This option deprecates the "prometheus-serve-addr" in the
  # "cilium-metrics-config" ConfigMap
  # NOTE that this will open the port on ALL nodes where Cilium pods are
  # scheduled.
  prometheus-serve-addr: ":{{- or .AgentPrometheusPort "9090" }}"
  {{ end }}
  # Enable IPv4 addressing. If enabled, all endpoints are allocated an IPv4
  # address.
  enable-ipv4: "{{- if or  (.EnableIpv4) (and (not (.EnableIpv4)) (not (.EnableIpv6))) -}}true{{- else -}}false{{- end -}}"
  # Enable IPv6 addressing. If enabled, all endpoints are allocated an IPv6
  # address.
  enable-ipv6: "{{- if .EnableIpv6 -}}true{{- else -}}false{{- end -}}"
  # If you want cilium monitor to aggregate tracing for packets, set this level
  # to "low", "medium", or "maximum". The higher the level, the less packets
  # that will be seen in monitor output.
  monitor-aggregation: "{{- if eq .MonitorAggregation "" -}}medium{{- else -}}{{ .MonitorAggregation }}{{- end -}}"
  # ct-global-max-entries-* specifies the maximum number of connections
  # supported across all endpoints, split by protocol: tcp or other. One pair
  # of maps uses these values for IPv4 connections, and another pair of maps
  # use these values for IPv6 connections.
  #
  # If these values are modified, then during the next Cilium startup the
  # tracking of ongoing connections may be disrupted. This may lead to brief
  # policy drops or a change in loadbalancing decisions for a connection.
  #
  # For users upgrading from Cilium 1.2 or earlier, to minimize disruption
  # during the upgrade process, comment out these options.
  bpf-ct-global-tcp-max: "{{- if eq .BPFCTGlobalTCPMax 0 -}}524288{{- else -}}{{ .BPFCTGlobalTCPMax}}{{- end -}}"
  bpf-ct-global-any-max: "{{- if eq .BPFCTGlobalAnyMax 0 -}}262144{{- else -}}{{ .BPFCTGlobalAnyMax}}{{- end -}}"

  # Pre-allocation of map entries allows per-packet latency to be reduced, at
  # the expense of up-front memory allocation for the entries in the maps. The
  # default value below will minimize memory usage in the default installation;
  # users who are sensitive to latency may consider setting this to "true".
  #
  # This option was introduced in Cilium 1.4. Cilium 1.3 and earlier ignore
  # this option and behave as though it is set to "true".
  #
  # If this value is modified, then during the next Cilium startup the restore
  # of existing endpoints and tracking of ongoing connections may be disrupted.
  # This may lead to policy drops or a change in loadbalancing decisions for a
  # connection for some time. Endpoints may need to be recreated to restore
  # connectivity.
  #
  # If this option is set to "false" during an upgrade from 1.3 or earlier to
  # 1.4 or later, then it may cause one-time disruptions during the upgrade.
  preallocate-bpf-maps: "{{- if .PreallocateBPFMaps -}}true{{- else -}}false{{- end -}}"
  # Regular expression matching compatible Istio sidecar istio-proxy
  # container image names
  sidecar-istio-proxy-image: "{{- if eq .SidecarIstioProxyImage "" -}}cilium/istio_proxy{{- else -}}{{ .SidecarIstioProxyImage }}{{- end -}}"
  # Encapsulation mode for communication between nodes
  # Possible values:
  #   - disabled
  #   - vxlan (default)
  #   - geneve
  tunnel: "{{- if eq .Tunnel "" -}}vxlan{{- else -}}{{ .Tunnel }}{{- end -}}"

  # Name of the cluster. Only relevant when building a mesh of clusters.
  cluster-name: "{{- if eq .ClusterName "" -}}default{{- else -}}{{ .ClusterName}}{{- end -}}"

  # This option is disabled by default starting from version 1.4.x in favor
  # of a more powerful DNS proxy-based implementation, see [0] for details.
  # Enable this option if you want to use FQDN policies but do not want to use
  # the DNS proxy.
  #
  # To ease upgrade, users may opt to set this option to "true".
  # Otherwise please refer to the Upgrade Guide [1] which explains how to
  # prepare policy rules for upgrade.
  #
  # [0] http://docs.cilium.io/en/stable/policy/language/#dns-based
  # [1] http://docs.cilium.io/en/stable/install/upgrade/#changes-that-may-require-action
  tofqdns-enable-poller: "{{- if .ToFqdnsEnablePoller -}}true{{- else -}}false{{- end -}}"
  # wait-bpf-mount makes init container wait until bpf filesystem is mounted
  wait-bpf-mount: "false"
  # Enable fetching of container-runtime specific metadata
  #
  # By default, the Kubernetes pod and namespace labels are retrieved and
  # associated with endpoints for identification purposes. By integrating
  # with the container runtime, container runtime specific labels can be
  # retrieved, such labels will be prefixed with container:
  #
  # CAUTION: The container runtime labels can include information such as pod
  # annotations which may result in each pod being associated a unique set of
  # labels which can result in excessive security identities being allocated.
  # Please review the labels filter when enabling container runtime labels.
  #
  # Supported values:
  # - containerd
  # - crio
  # - docker
  # - none
  # - auto (automatically detect the container runtime)
  #
  container-runtime: "{{- if eq .ContainerRuntimeLabels "" -}}none{{- else -}}{{ .ContainerRuntimeLabels }}{{- end -}}"
  masquerade: "{{- if .DisableMasquerade -}}false{{- else -}}true{{- end -}}"
  install-iptables-rules: "{{- if .IPTablesRulesNoinstall -}}false{{- else -}}true{{- end -}}"
  auto-direct-node-routes: "{{- if .AutoDirectNodeRoutes -}}true{{- else -}}false{{- end -}}"
  enable-node-port: "{{- if .EnableNodePort -}}true{{- else -}}false{{- end -}}"
  {{ with .Ipam }}
  ipam: {{ . }}
  {{ if eq . "eni" }}
  enable-endpoint-routes: "true"
  auto-create-cilium-node-resource: "true"
  blacklist-conflicting-routes: "false"
  {{ end }}
  {{ end }}
{{ end }} # With .Networking.Cilium end
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: cilium
  namespace: kube-system
  labels:
    role.kubernetes.io/networking: "1"
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: cilium-operator
  namespace: kube-system
  labels:
    role.kubernetes.io/networking: "1"
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: cilium
  labels:
    role.kubernetes.io/networking: "1"
rules:
- apiGroups:
  - networking.k8s.io
  resources:
  - networkpolicies
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - ""
  resources:
  - namespaces
  - services
  - nodes
  - endpoints
  - componentstatuses
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - ""
  resources:
  - pods
  - nodes
  verbs:
  - get
  - list
  - watch
  - update
- apiGroups:
  - ""
  resources:
  - nodes
  - nodes/status
  verbs:
  - patch
- apiGroups:
  - extensions
  resources:
  - ingresses
  verbs:
  - create
  - get
  - list
  - watch
- apiGroups:
  - apiextensions.k8s.io
  resources:
  - customresourcedefinitions
  verbs:
  - create
  - get
  - list
  - watch
  - update
- apiGroups:
  - cilium.io
  resources:
  - ciliumnetworkpolicies
  - ciliumnetworkpolicies/status
  - ciliumendpoints
  - ciliumendpoints/status
  - ciliumnodes
  - ciliumnodes/status
  - ciliumidentities
  - ciliumidentities/status
  verbs:
  - '*'
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: cilium-operator
  labels:
    role.kubernetes.io/networking: "1"
rules:
- apiGroups:
  - ""
  resources:
  # to automatically delete [core|kube]dns pods so that are starting to being
  # managed by Cilium
  - pods
  verbs:
  - get
  - list
  - watch
  - delete
- apiGroups:
  - ""
  resources:
  # to automatically read from k8s and import the node's pod CIDR to cilium's
  # etcd so all nodes know how to reach another pod running in a different
  # node.
  - nodes
  # to perform the translation of a CNP that contains ` + "`" + `ToGroup` + "`" + ` to its endpoints
  - services
  - endpoints
  # to check apiserver connectivity
  - namespaces
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - cilium.io
  resources:
  - ciliumnetworkpolicies
  - ciliumnetworkpolicies/status
  - ciliumendpoints
  - ciliumendpoints/status
  - ciliumnodes
  - ciliumnodes/status
  - ciliumidentities
  - ciliumidentities/status
  verbs:
  - '*'
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: cilium
  labels:
    role.kubernetes.io/networking: "1"
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cilium
subjects:
- kind: ServiceAccount
  name: cilium
  namespace: kube-system
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: cilium-operator
  labels:
    role.kubernetes.io/networking: "1"
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cilium-operator
subjects:
- kind: ServiceAccount
  name: cilium-operator
  namespace: kube-system
---
apiVersion: apps/v1
kind: DaemonSet
metadata:
  labels:
    k8s-app: cilium
    kubernetes.io/cluster-service: "true"
    role.kubernetes.io/networking: "1"
  name: cilium
  namespace: kube-system
spec:
  selector:
    matchLabels:
      k8s-app: cilium
      kubernetes.io/cluster-service: "true"
  template:
    metadata:
      annotations:
        # This annotation plus the CriticalAddonsOnly toleration makes
        # cilium to be a critical pod in the cluster, which ensures cilium
        # gets priority scheduling.
        # https://kubernetes.io/docs/tasks/administer-cluster/guaranteed-scheduling-critical-addon-pods/
        scheduler.alpha.kubernetes.io/critical-pod: ""
        scheduler.alpha.kubernetes.io/tolerations: '[{"key":"dedicated","operator":"Equal","value":"master","effect":"NoSchedule"}]'
      labels:
        k8s-app: cilium
        kubernetes.io/cluster-service: "true"
    spec:
      containers:
      - args:
        - --config-dir=/tmp/cilium/config-map
        command:
        - cilium-agent
        env:
        - name: K8S_NODE_NAME
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: spec.nodeName
        - name: CILIUM_K8S_NAMESPACE
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: metadata.namespace
        - name: CILIUM_FLANNEL_MASTER_DEVICE
          valueFrom:
            configMapKeyRef:
              key: flannel-master-device
              name: cilium-config
              optional: true
        - name: CILIUM_FLANNEL_UNINSTALL_ON_EXIT
          valueFrom:
            configMapKeyRef:
              key: flannel-uninstall-on-exit
              name: cilium-config
              optional: true
        - name: CILIUM_CLUSTERMESH_CONFIG
          value: /var/lib/cilium/clustermesh/
        - name: CILIUM_CNI_CHAINING_MODE
          valueFrom:
            configMapKeyRef:
              key: cni-chaining-mode
              name: cilium-config
              optional: true
        - name: CILIUM_CUSTOM_CNI_CONF
          valueFrom:
            configMapKeyRef:
              key: custom-cni-conf
              name: cilium-config
              optional: true
        - name: KUBERNETES_SERVICE_HOST
          value: "{{ .MasterInternalName }}"
        - name: KUBERNETES_SERVICE_PORT
          value: "443"
        {{ with .Networking.Cilium.EnablePolicy }}
        - name: CILIUM_ENABLE_POLICY
          value: {{ . }}
        {{ end }}
{{ with .Networking.Cilium }}
        image: "docker.io/cilium/cilium:{{- or .Version "v1.6.6" }}"
        imagePullPolicy: IfNotPresent
        lifecycle:
          postStart:
            exec:
              command:
              - /cni-install.sh
          preStop:
            exec:
              command:
              - /cni-uninstall.sh
        livenessProbe:
          exec:
            command:
            - cilium
            - status
            - --brief
          failureThreshold: 10
          # The initial delay for the liveness probe is intentionally large to
          # avoid an endless kill & restart cycle if in the event that the initial
          # bootstrapping takes longer than expected.
          initialDelaySeconds: 120
          periodSeconds: 30
          successThreshold: 1
          timeoutSeconds: 5
        name: cilium-agent
        {{ if .EnablePrometheusMetrics }}
        ports:
        - containerPort: {{ or .AgentPrometheusPort "9090" }}
          hostPort: {{ or .AgentPrometheusPort "9090" }}
          name: prometheus
          protocol: TCP
        {{ end }}
        readinessProbe:
          exec:
            command:
            - cilium
            - status
            - --brief
          failureThreshold: 3
          initialDelaySeconds: 5
          periodSeconds: 30
          successThreshold: 1
          timeoutSeconds: 5
        securityContext:
          capabilities:
            add:
            - NET_ADMIN
            - SYS_MODULE
          privileged: true
        volumeMounts:
        - mountPath: /sys/fs/bpf
          name: bpf-maps
        - mountPath: /var/run/cilium
          name: cilium-run
        - mountPath: /host/opt/cni/bin
          name: cni-path
        - mountPath: /host/etc/cni/net.d
          name: etc-cni-netd
        - mountPath: /var/lib/cilium/clustermesh
          name: clustermesh-secrets
          readOnly: true
        - mountPath: /tmp/cilium/config-map
          name: cilium-config-path
          readOnly: true
          # Needed to be able to load kernel modules
        - mountPath: /lib/modules
          name: lib-modules
          readOnly: true
        - mountPath: /run/xtables.lock
          name: xtables-lock
      hostNetwork: true
      initContainers:
      - command:
        - /init-container.sh
        env:
        - name: CILIUM_ALL_STATE
          valueFrom:
            configMapKeyRef:
              key: clean-cilium-state
              name: cilium-config
              optional: true
        - name: CILIUM_BPF_STATE
          valueFrom:
            configMapKeyRef:
              key: clean-cilium-bpf-state
              name: cilium-config
              optional: true
        - name: CILIUM_WAIT_BPF_MOUNT
          valueFrom:
            configMapKeyRef:
              key: wait-bpf-mount
              name: cilium-config
              optional: true
        image: "docker.io/cilium/cilium:{{ .Version }}"
## end of ` + "`" + `with .Networking.Cilium` + "`" + `
#{{ end }}
        imagePullPolicy: IfNotPresent
        name: clean-cilium-state
        securityContext:
          capabilities:
            add:
            - NET_ADMIN
          privileged: true
        volumeMounts:
        - mountPath: /sys/fs/bpf
          name: bpf-maps
        - mountPath: /var/run/cilium
          name: cilium-run
      restartPolicy: Always
      serviceAccount: cilium
      serviceAccountName: cilium
      terminationGracePeriodSeconds: 1
      tolerations:
      - operator: Exists
      volumes:
        # To keep state between restarts / upgrades
      - hostPath:
          path: /var/run/cilium
          type: DirectoryOrCreate
        name: cilium-run
        # To keep state between restarts / upgrades for bpf maps
      - hostPath:
          path: /sys/fs/bpf
          type: DirectoryOrCreate
        name: bpf-maps
      # To install cilium cni plugin in the host
      - hostPath:
          path:  /opt/cni/bin
          type: DirectoryOrCreate
        name: cni-path
        # To install cilium cni configuration in the host
      - hostPath:
          path: /etc/cni/net.d
          type: DirectoryOrCreate
        name: etc-cni-netd
        # To be able to load kernel modules
      - hostPath:
          path: /lib/modules
        name: lib-modules
        # To access iptables concurrently with other processes (e.g. kube-proxy)
      - hostPath:
          path: /run/xtables.lock
          type: FileOrCreate
        name: xtables-lock
        # To read the clustermesh configuration
      - name: clustermesh-secrets
        secret:
          defaultMode: 420
          optional: true
          secretName: cilium-clustermesh
        # To read the configuration from the config map
      - configMap:
          name: cilium-config
        name: cilium-config-path
  updateStrategy:
    rollingUpdate:
      maxUnavailable: 2
    type: RollingUpdate
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    io.cilium/app: operator
    name: cilium-operator
    role.kubernetes.io/networking: "1"
  name: cilium-operator
  namespace: kube-system
spec:
  replicas: 1
  selector:
    matchLabels:
      io.cilium/app: operator
      name: cilium-operator
  strategy:
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 1
    type: RollingUpdate
  template:
    metadata:
      labels:
        io.cilium/app: operator
        name: cilium-operator
    spec:
      containers:
      - args:
        - --debug=$(CILIUM_DEBUG)
{{ with .Networking.Cilium }}
        {{ if .EnablePrometheusMetrics }}
        - --enable-metrics
        {{ end }}
{{ end }}
        command:
        - cilium-operator
        env:
        - name: CILIUM_K8S_NAMESPACE
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: metadata.namespace
        - name: K8S_NODE_NAME
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: spec.nodeName
        - name: CILIUM_DEBUG
          valueFrom:
            configMapKeyRef:
              key: debug
              name: cilium-config
              optional: true
        - name: CILIUM_CLUSTER_NAME
          valueFrom:
            configMapKeyRef:
              key: cluster-name
              name: cilium-config
              optional: true
        - name: CILIUM_CLUSTER_ID
          valueFrom:
            configMapKeyRef:
              key: cluster-id
              name: cilium-config
              optional: true
        - name: CILIUM_IPAM
          valueFrom:
            configMapKeyRef:
              key: ipam
              name: cilium-config
              optional: true
        - name: CILIUM_DISABLE_ENDPOINT_CRD
          valueFrom:
            configMapKeyRef:
              key: disable-endpoint-crd
              name: cilium-config
              optional: true
        - name: CILIUM_KVSTORE
          valueFrom:
            configMapKeyRef:
              key: kvstore
              name: cilium-config
              optional: true
        - name: CILIUM_KVSTORE_OPT
          valueFrom:
            configMapKeyRef:
              key: kvstore-opt
              name: cilium-config
              optional: true
        - name: AWS_ACCESS_KEY_ID
          valueFrom:
            secretKeyRef:
              key: AWS_ACCESS_KEY_ID
              name: cilium-aws
              optional: true
        - name: AWS_SECRET_ACCESS_KEY
          valueFrom:
            secretKeyRef:
              key: AWS_SECRET_ACCESS_KEY
              name: cilium-aws
              optional: true
        - name: AWS_DEFAULT_REGION
          valueFrom:
            secretKeyRef:
              key: AWS_DEFAULT_REGION
              name: cilium-aws
              optional: true
        - name: KUBERNETES_SERVICE_HOST
          value: "{{ .MasterInternalName }}"
        - name: KUBERNETES_SERVICE_PORT
          value: "443"
{{ with .Networking.Cilium }}
        image: "docker.io/cilium/operator:{{- or .Version "v1.6.6" }}"
        imagePullPolicy: IfNotPresent
        name: cilium-operator
        {{ if .EnablePrometheusMetrics }}
        ports:
        - containerPort: 6942
          hostPort: 6942
          name: prometheus
          protocol: TCP
        {{ end }}
        livenessProbe:
          httpGet:
            path: /healthz
            port: 9234
            scheme: HTTP
          initialDelaySeconds: 60
          periodSeconds: 10
          timeoutSeconds: 3
      hostNetwork: true
      restartPolicy: Always
      serviceAccount: cilium-operator
      serviceAccountName: cilium-operator
      {{if eq .Ipam "eni" }}
      nodeSelector:
        node-role.kubernetes.io/master: ""
      tolerations:
      - effect: NoSchedule
        key: node-role.kubernetes.io/master
      - effect: NoExecute
        key: node.kubernetes.io/not-ready
        operator: Exists
        tolerationSeconds: 300
      - effect: NoExecute
        key: node.kubernetes.io/unreachable
        operator: Exists
        tolerationSeconds: 300
       {{ end }}
{{ end }}`)

func cloudupResourcesAddonsNetworkingCiliumIoK8s17YamlTemplateBytes() ([]byte, error) {
	return _cloudupResourcesAddonsNetworkingCiliumIoK8s17YamlTemplate, nil
}

func cloudupResourcesAddonsNetworkingCiliumIoK8s17YamlTemplate() (*asset, error) {
	bytes, err := cloudupResourcesAddonsNetworkingCiliumIoK8s17YamlTemplateBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "cloudup/resources/addons/networking.cilium.io/k8s-1.7.yaml.template", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _cloudupResourcesAddonsNetworkingFlannelK8s112YamlTemplate = []byte(`kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: flannel
  labels:
    role.kubernetes.io/networking: "1"
rules:
  - apiGroups:
      - ""
    resources:
      - pods
    verbs:
      - get
  - apiGroups:
      - ""
    resources:
      - nodes
    verbs:
      - list
      - watch
  - apiGroups:
      - ""
    resources:
      - nodes/status
    verbs:
      - patch
---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: flannel
  labels:
    role.kubernetes.io/networking: "1"
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: flannel
subjects:
- kind: ServiceAccount
  name: flannel
  namespace: kube-system
---
kind: ServiceAccount
apiVersion: v1
metadata:
  name: flannel
  namespace: kube-system
  labels:
    role.kubernetes.io/networking: "1"
---
kind: ConfigMap
apiVersion: v1
metadata:
  name: kube-flannel-cfg
  namespace: kube-system
  labels:
    k8s-app: flannel
    role.kubernetes.io/networking: "1"
data:
  cni-conf.json: |
    {
      "cniVersion": "0.2.0",
      "name": "cbr0",
      "plugins": [
        {
          "type": "flannel",
          "delegate": {
            "forceAddress": true,
            "isDefaultGateway": true,
            "hairpinMode": true
          }
        },
        {
          "type": "portmap",
          "capabilities": {
            "portMappings": true
          }
        }
      ]
    }
  net-conf.json: |
    {
      "Network": "{{ .NonMasqueradeCIDR }}",
      "Backend": {
        "Type": "{{ FlannelBackendType }}"
      }
    }
---
kind: DaemonSet
apiVersion: apps/v1
metadata:
  name: kube-flannel-ds
  namespace: kube-system
  labels:
    k8s-app: flannel
    role.kubernetes.io/networking: "1"
spec:
  selector:
    matchLabels:
      tier: node
      app: flannel
      role.kubernetes.io/networking: "1"
  template:
    metadata:
      labels:
        tier: node
        app: flannel
        role.kubernetes.io/networking: "1"
    spec:
      priorityClassName: system-node-critical
      hostNetwork: true
      nodeSelector:
        beta.kubernetes.io/arch: amd64
      serviceAccountName: flannel
      tolerations:
      - operator: Exists
      initContainers:
      - name: install-cni
        image: quay.io/coreos/flannel:v0.11.0-amd64
        command:
        - cp
        args:
        - -f
        - /etc/kube-flannel/cni-conf.json
        - /etc/cni/net.d/10-flannel.conflist
        volumeMounts:
        - name: cni
          mountPath: /etc/cni/net.d
        - name: flannel-cfg
          mountPath: /etc/kube-flannel/
      containers:
      - name: kube-flannel
        image: quay.io/coreos/flannel:v0.11.0-amd64
        command:
        - "/opt/bin/flanneld"
        - "--ip-masq"
        - "--kube-subnet-mgr"
        - "--iptables-resync={{- or .Networking.Flannel.IptablesResyncSeconds "5" }}"
        securityContext:
          privileged: true
        env:
        - name: POD_NAME
          valueFrom:
            fieldRef:
              fieldPath: metadata.name
        - name: POD_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        resources:
          limits:
            memory: 100Mi
          requests:
            cpu: 100m
            memory: 100Mi
        volumeMounts:
        - name: run
          mountPath: /run
        - name: flannel-cfg
          mountPath: /etc/kube-flannel/
      volumes:
        - name: run
          hostPath:
            path: /run
        - name: cni
          hostPath:
            path: /etc/cni/net.d
        - name: flannel-cfg
          configMap:
            name: kube-flannel-cfg
`)

func cloudupResourcesAddonsNetworkingFlannelK8s112YamlTemplateBytes() ([]byte, error) {
	return _cloudupResourcesAddonsNetworkingFlannelK8s112YamlTemplate, nil
}

func cloudupResourcesAddonsNetworkingFlannelK8s112YamlTemplate() (*asset, error) {
	bytes, err := cloudupResourcesAddonsNetworkingFlannelK8s112YamlTemplateBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "cloudup/resources/addons/networking.flannel/k8s-1.12.yaml.template", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _cloudupResourcesAddonsNetworkingFlannelK8s16YamlTemplate = []byte(`kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: flannel
  labels:
    role.kubernetes.io/networking: "1"
rules:
  - apiGroups:
      - ""
    resources:
      - pods
    verbs:
      - get
  - apiGroups:
      - ""
    resources:
      - nodes
    verbs:
      - list
      - watch
  - apiGroups:
      - ""
    resources:
      - nodes/status
    verbs:
      - patch
---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: flannel
  labels:
    role.kubernetes.io/networking: "1"
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: flannel
subjects:
- kind: ServiceAccount
  name: flannel
  namespace: kube-system
---
kind: ServiceAccount
apiVersion: v1
metadata:
  name: flannel
  namespace: kube-system
  labels:
    role.kubernetes.io/networking: "1"
---
kind: ConfigMap
apiVersion: v1
metadata:
  name: kube-flannel-cfg
  namespace: kube-system
  labels:
    k8s-app: flannel
    role.kubernetes.io/networking: "1"
data:
  cni-conf.json: |
    {
      "name": "cbr0",
      "plugins": [
        {
          "type": "flannel",
          "delegate": {
            "forceAddress": true,
            "isDefaultGateway": true,
            "hairpinMode": true
          }
        },
        {
          "type": "portmap",
          "capabilities": {
            "portMappings": true
          }
        }
      ]
    }
  net-conf.json: |
    {
      "Network": "{{ .NonMasqueradeCIDR }}",
      "Backend": {
        "Type": "{{ FlannelBackendType }}"
      }
    }
---
kind: DaemonSet
apiVersion: extensions/v1beta1
metadata:
  name: kube-flannel-ds
  namespace: kube-system
  labels:
    k8s-app: flannel
    role.kubernetes.io/networking: "1"
spec:
  template:
    metadata:
      labels:
        tier: node
        app: flannel
        role.kubernetes.io/networking: "1"
    spec:
      hostNetwork: true
      nodeSelector:
        beta.kubernetes.io/arch: amd64
      serviceAccountName: flannel
      tolerations:
      - operator: Exists
      initContainers:
      - name: install-cni
        image: quay.io/coreos/flannel:v0.11.0-amd64
        command:
        - cp
        args:
        - -f
        - /etc/kube-flannel/cni-conf.json
        - /etc/cni/net.d/10-flannel.conflist
        volumeMounts:
        - name: cni
          mountPath: /etc/cni/net.d
        - name: flannel-cfg
          mountPath: /etc/kube-flannel/
      containers:
      - name: kube-flannel
        image: quay.io/coreos/flannel:v0.11.0-amd64
        command:
        - "/opt/bin/flanneld"
        - "--ip-masq"
        - "--kube-subnet-mgr"
        - "--iptables-resync={{- or .Networking.Flannel.IptablesResyncSeconds "5" }}"
        securityContext:
          privileged: true
        env:
        - name: POD_NAME
          valueFrom:
            fieldRef:
              fieldPath: metadata.name
        - name: POD_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        resources:
          limits:
            memory: 100Mi
          requests:
            cpu: 100m
            memory: 100Mi
        volumeMounts:
        - name: run
          mountPath: /run
        - name: flannel-cfg
          mountPath: /etc/kube-flannel/
      volumes:
        - name: run
          hostPath:
            path: /run
        - name: cni
          hostPath:
            path: /etc/cni/net.d
        - name: flannel-cfg
          configMap:
            name: kube-flannel-cfg
`)

func cloudupResourcesAddonsNetworkingFlannelK8s16YamlTemplateBytes() ([]byte, error) {
	return _cloudupResourcesAddonsNetworkingFlannelK8s16YamlTemplate, nil
}

func cloudupResourcesAddonsNetworkingFlannelK8s16YamlTemplate() (*asset, error) {
	bytes, err := cloudupResourcesAddonsNetworkingFlannelK8s16YamlTemplateBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "cloudup/resources/addons/networking.flannel/k8s-1.6.yaml.template", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _cloudupResourcesAddonsNetworkingKopeIoK8s112Yaml = []byte(`apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: kopeio-networking-agent
  namespace: kube-system
  labels:
    k8s-addon: networking.kope.io
    role.kubernetes.io/networking: "1"
spec:
  selector:
    matchLabels:
      name: kopeio-networking-agent
      role.kubernetes.io/networking: "1"
  template:
    metadata:
      labels:
        name: kopeio-networking-agent
        role.kubernetes.io/networking: "1"
      annotations:
        scheduler.alpha.kubernetes.io/critical-pod: ''
        scheduler.alpha.kubernetes.io/tolerations: '[{"key":"CriticalAddonsOnly", "operator":"Exists"}]'
    spec:
      hostPID: true
      hostIPC: true
      hostNetwork: true
      containers:
        - resources:
            requests:
              cpu: 50m
              memory: 100Mi
            limits:
              memory: 100Mi
          securityContext:
            privileged: true
          image: kopeio/networking-agent:1.0.20181028
          name: networking-agent
          volumeMounts:
            - name: lib-modules
              mountPath: /lib/modules
              readOnly: true
          env:
          - name: NODE_NAME
            valueFrom:
              fieldRef:
                fieldPath: spec.nodeName
      serviceAccountName: kopeio-networking-agent
      priorityClassName: system-node-critical
      tolerations:
      - effect: NoSchedule
        operator: Exists
      - effect: NoExecute
        operator: Exists
      volumes:
        - name: lib-modules
          hostPath:
            path: /lib/modules

---

apiVersion: v1
kind: ServiceAccount
metadata:
  name: kopeio-networking-agent
  namespace: kube-system
  labels:
    k8s-addon: networking.kope.io
    role.kubernetes.io/networking: "1"

---

apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  labels:
    k8s-addon: networking.kope.io
  name: kopeio:networking-agent
rules:
- apiGroups:
  - ""
  resources:
  - nodes
  verbs:
  - get
  - list
  - watch
  - patch
- apiGroups:
  - ""
  resources:
  - nodes/status
  verbs:
  - patch

---

apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  labels:
    k8s-addon: networking.kope.io
  name: kopeio:networking-agent
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: kopeio:networking-agent
subjects:
- apiGroup: rbac.authorization.k8s.io
  kind: User
  name: system:serviceaccount:kube-system:kopeio-networking-agent
`)

func cloudupResourcesAddonsNetworkingKopeIoK8s112YamlBytes() ([]byte, error) {
	return _cloudupResourcesAddonsNetworkingKopeIoK8s112Yaml, nil
}

func cloudupResourcesAddonsNetworkingKopeIoK8s112Yaml() (*asset, error) {
	bytes, err := cloudupResourcesAddonsNetworkingKopeIoK8s112YamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "cloudup/resources/addons/networking.kope.io/k8s-1.12.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _cloudupResourcesAddonsNetworkingKopeIoK8s16Yaml = []byte(`apiVersion: extensions/v1beta1
kind: DaemonSet
metadata:
  name: kopeio-networking-agent
  namespace: kube-system
  labels:
    k8s-addon: networking.kope.io
    role.kubernetes.io/networking: "1"
spec:
  template:
    metadata:
      labels:
        name: kopeio-networking-agent
        role.kubernetes.io/networking: "1"
      annotations:
        scheduler.alpha.kubernetes.io/critical-pod: ''
        scheduler.alpha.kubernetes.io/tolerations: '[{"key":"CriticalAddonsOnly", "operator":"Exists"}]'
    spec:
      hostPID: true
      hostIPC: true
      hostNetwork: true
      containers:
        - resources:
            requests:
              cpu: 50m
              memory: 100Mi
            limits:
              memory: 100Mi
          securityContext:
            privileged: true
          image: kopeio/networking-agent:1.0.20181028
          name: networking-agent
          volumeMounts:
            - name: lib-modules
              mountPath: /lib/modules
              readOnly: true
          env:
          - name: NODE_NAME
            valueFrom:
              fieldRef:
                fieldPath: spec.nodeName
      serviceAccountName: kopeio-networking-agent
      tolerations:
      - effect: NoSchedule
        operator: Exists
      - effect: NoExecute
        operator: Exists
      volumes:
        - name: lib-modules
          hostPath:
            path: /lib/modules

---

apiVersion: v1
kind: ServiceAccount
metadata:
  name: kopeio-networking-agent
  namespace: kube-system
  labels:
    k8s-addon: networking.kope.io
    role.kubernetes.io/networking: "1"

---

apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  labels:
    k8s-addon: networking.kope.io
  name: kopeio:networking-agent
rules:
- apiGroups:
  - ""
  resources:
  - nodes
  verbs:
  - get
  - list
  - watch
  - patch
- apiGroups:
  - ""
  resources:
  - nodes/status
  verbs:
  - patch

---

apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  labels:
    k8s-addon: networking.kope.io
  name: kopeio:networking-agent
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: kopeio:networking-agent
subjects:
- apiGroup: rbac.authorization.k8s.io
  kind: User
  name: system:serviceaccount:kube-system:kopeio-networking-agent
`)

func cloudupResourcesAddonsNetworkingKopeIoK8s16YamlBytes() ([]byte, error) {
	return _cloudupResourcesAddonsNetworkingKopeIoK8s16Yaml, nil
}

func cloudupResourcesAddonsNetworkingKopeIoK8s16Yaml() (*asset, error) {
	bytes, err := cloudupResourcesAddonsNetworkingKopeIoK8s16YamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "cloudup/resources/addons/networking.kope.io/k8s-1.6.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _cloudupResourcesAddonsNetworkingKuberouterK8s112YamlTemplate = []byte(`# Pulled and modified from https://github.com/cloudnativelabs/kube-router/blob/v0.4.0/daemonset/kubeadm-kuberouter.yaml

apiVersion: v1
kind: ConfigMap
metadata:
  name: kube-router-cfg
  namespace: kube-system
  labels:
    tier: node
    k8s-app: kube-router
data:
  cni-conf.json: |
    {
       "cniVersion":"0.3.0",
       "name":"mynet",
       "plugins":[
          {
             "name":"kubernetes",
             "type":"bridge",
             "bridge":"kube-bridge",
             "isDefaultGateway":true,
             "ipam":{
                "type":"host-local"
             }
          }
       ]
    }
---
apiVersion: apps/v1
kind: DaemonSet
metadata:
  labels:
    k8s-app: kube-router
    tier: node
  name: kube-router
  namespace: kube-system
spec:
  selector:
    matchLabels:
      k8s-app: kube-router
      tier: node
  template:
    metadata:
      labels:
        k8s-app: kube-router
        tier: node
    spec:
      priorityClassName: system-node-critical
      serviceAccountName: kube-router
      containers:
      - name: kube-router
        image: docker.io/cloudnativelabs/kube-router:v0.4.0
        args:
        - --run-router=true
        - --run-firewall=true
        - --run-service-proxy=true
        - --kubeconfig=/var/lib/kube-router/kubeconfig
        - --metrics-port=12013
        env:
        - name: NODE_NAME
          valueFrom:
            fieldRef:
              fieldPath: spec.nodeName
        - name: KUBE_ROUTER_CNI_CONF_FILE
          value: /etc/cni/net.d/10-kuberouter.conflist
        livenessProbe:
          httpGet:
            path: /healthz
            port: 20244
          initialDelaySeconds: 10
          periodSeconds: 3
        resources:
          requests:
            cpu: 100m
            memory: 250Mi
        securityContext:
          privileged: true
        volumeMounts:
        - name: lib-modules
          mountPath: /lib/modules
          readOnly: true
        - name: cni-conf-dir
          mountPath: /etc/cni/net.d
        - name: kubeconfig
          mountPath: /var/lib/kube-router/kubeconfig
          readOnly: true
      initContainers:
      - name: install-cni
        image: busybox
        command:
        - /bin/sh
        - -c
        - set -e -x;
          if [ ! -f /etc/cni/net.d/10-kuberouter.conflist ]; then
            if [ -f /etc/cni/net.d/*.conf ]; then
              rm -f /etc/cni/net.d/*.conf;
            fi;
            TMP=/etc/cni/net.d/.tmp-kuberouter-cfg;
            cp /etc/kube-router/cni-conf.json ${TMP};
            mv ${TMP} /etc/cni/net.d/10-kuberouter.conflist;
          fi
        volumeMounts:
        - mountPath: /etc/cni/net.d
          name: cni-conf-dir
        - mountPath: /etc/kube-router
          name: kube-router-cfg
      hostNetwork: true
      tolerations:
      - key: CriticalAddonsOnly
        operator: Exists
      - effect: NoSchedule
        key: node-role.kubernetes.io/master
        operator: Exists
      - effect: NoSchedule
        key: node.kubernetes.io/not-ready
        operator: Exists
      volumes:
      - name: lib-modules
        hostPath:
          path: /lib/modules
      - name: cni-conf-dir
        hostPath:
          path: /etc/cni/net.d
      - name: kube-router-cfg
        configMap:
          name: kube-router-cfg
      - name: kubeconfig
        hostPath:
          path: /var/lib/kube-router/kubeconfig
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: kube-router
  namespace: kube-system
---
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: kube-router
  namespace: kube-system
rules:
  - apiGroups:
    - ""
    resources:
      - namespaces
      - pods
      - services
      - nodes
      - endpoints
    verbs:
      - list
      - get
      - watch
  - apiGroups:
    - "networking.k8s.io"
    resources:
      - networkpolicies
    verbs:
      - list
      - get
      - watch
  - apiGroups:
    - extensions
    resources:
      - networkpolicies
    verbs:
      - get
      - list
      - watch
---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: kube-router
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: kube-router
subjects:
- kind: ServiceAccount
  name: kube-router
  namespace: kube-system
- kind: User
  name: system:kube-router
`)

func cloudupResourcesAddonsNetworkingKuberouterK8s112YamlTemplateBytes() ([]byte, error) {
	return _cloudupResourcesAddonsNetworkingKuberouterK8s112YamlTemplate, nil
}

func cloudupResourcesAddonsNetworkingKuberouterK8s112YamlTemplate() (*asset, error) {
	bytes, err := cloudupResourcesAddonsNetworkingKuberouterK8s112YamlTemplateBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "cloudup/resources/addons/networking.kuberouter/k8s-1.12.yaml.template", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _cloudupResourcesAddonsNetworkingKuberouterK8s16YamlTemplate = []byte(`apiVersion: v1
kind: ConfigMap
metadata:
  name: kube-router-cfg
  namespace: kube-system
  labels:
    tier: node
    k8s-app: kube-router
data:
  cni-conf.json: |
    {
      "name":"kubernetes",
      "type":"bridge",
      "bridge":"kube-bridge",
      "isDefaultGateway":true,
      "ipam": {
        "type":"host-local"
      }
    }
---
apiVersion: extensions/v1beta1
kind: DaemonSet
metadata:
  labels:
    k8s-app: kube-router
    tier: node
  name: kube-router
  namespace: kube-system
spec:
  template:
    metadata:
      labels:
        k8s-app: kube-router
        tier: node
      annotations:
        scheduler.alpha.kubernetes.io/critical-pod: ''
    spec:
      containers:
      - name: kube-router
        image: cloudnativelabs/kube-router:v0.3.1
        args:
        - --run-router=true
        - --run-firewall=true
        - --run-service-proxy=true
        - --metrics-port=12013
        - --kubeconfig=/var/lib/kube-router/kubeconfig
        env:
        - name: NODE_NAME
          valueFrom:
            fieldRef:
              fieldPath: spec.nodeName
        livenessProbe:
          httpGet:
            path: /healthz
            port: 20244
          initialDelaySeconds: 10
          periodSeconds: 3
        resources:
          requests:
            cpu: 100m
            memory: 250Mi
        securityContext:
          privileged: true
        volumeMounts:
        - name: lib-modules
          mountPath: /lib/modules
          readOnly: true
        - name: cni-conf-dir
          mountPath: /etc/cni/net.d
        - name: kubeconfig
          mountPath: /var/lib/kube-router/kubeconfig
          readOnly: true
      initContainers:
      - name: install-cni
        image: busybox
        command:
        - /bin/sh
        - -c
        - set -e -x;
          if [ ! -f /etc/cni/net.d/10-kuberouter.conf ]; then
            TMP=/etc/cni/net.d/.tmp-kuberouter-cfg;
            cp /etc/kube-router/cni-conf.json ${TMP};
            mv ${TMP} /etc/cni/net.d/10-kuberouter.conf;
          fi
        volumeMounts:
        - name: cni-conf-dir
          mountPath: /etc/cni/net.d
        - name: kube-router-cfg
          mountPath: /etc/kube-router
      hostNetwork: true
      serviceAccountName: kube-router
      tolerations:
      - key: CriticalAddonsOnly
        operator: Exists
      - effect: NoSchedule
        operator: Exists
      volumes:
      - hostPath:
          path: /lib/modules
        name: lib-modules
      - hostPath:
          path: /etc/cni/net.d
        name: cni-conf-dir
      - name: kubeconfig
        hostPath:
          path: /var/lib/kube-router/kubeconfig
      - name: kube-router-cfg
        configMap:
          name: kube-router-cfg
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: kube-router
  namespace: kube-system
---
# Kube-router roles
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: kube-router
  namespace: kube-system
rules:
  - apiGroups: [""]
    resources:
      - namespaces
      - pods
      - services
      - nodes
      - endpoints
    verbs:
      - get
      - list
      - watch
  - apiGroups: ["networking.k8s.io"]
    resources:
      - networkpolicies
    verbs:
      - get
      - list
      - watch
  - apiGroups: ["extensions"]
    resources:
      - networkpolicies
    verbs:
      - get
      - list
      - watch
---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: kube-router
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: kube-router
subjects:
- kind: ServiceAccount
  name: kube-router
  namespace: kube-system
- kind: User
  name: system:kube-router
`)

func cloudupResourcesAddonsNetworkingKuberouterK8s16YamlTemplateBytes() ([]byte, error) {
	return _cloudupResourcesAddonsNetworkingKuberouterK8s16YamlTemplate, nil
}

func cloudupResourcesAddonsNetworkingKuberouterK8s16YamlTemplate() (*asset, error) {
	bytes, err := cloudupResourcesAddonsNetworkingKuberouterK8s16YamlTemplateBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "cloudup/resources/addons/networking.kuberouter/k8s-1.6.yaml.template", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _cloudupResourcesAddonsNetworkingProjectcalicoOrgK8s112YamlTemplate = []byte(`# Pulled and modified from: https://docs.projectcalico.org/v3.9/manifests/calico-typha.yaml

---
# Source: calico/templates/calico-config.yaml
# This ConfigMap is used to configure a self-hosted Calico installation.
kind: ConfigMap
apiVersion: v1
metadata:
  name: calico-config
  namespace: kube-system
  labels:
    role.kubernetes.io/networking: "1"
data:
  # You must set a non-zero value for Typha replicas below.
  typha_service_name: "{{- if .Networking.Calico.TyphaReplicas -}}calico-typha{{- else -}}none{{- end -}}"
  # Configure the backend to use.
  calico_backend: "bird"

  # Configure the MTU to use
  {{- if .Networking.Calico.MTU }}
  veth_mtu: "{{ .Networking.Calico.MTU }}"
  {{- else }}
  veth_mtu: "{{- if eq .CloudProvider "openstack" -}}1430{{- else -}}1440{{- end -}}"
  {{- end }}

  # The CNI network configuration to install on each node.  The special
  # values in this config will be automatically populated.
  cni_network_config: |-
    {
      "name": "k8s-pod-network",
      "cniVersion": "0.3.1",
      "plugins": [
        {
          "type": "calico",
          "log_level": "info",
          "datastore_type": "kubernetes",
          "nodename": "__KUBERNETES_NODE_NAME__",
          "mtu": __CNI_MTU__,
          "ipam": {
              "type": "calico-ipam"
          },
          "policy": {
              "type": "k8s"
          },
          "kubernetes": {
              "kubeconfig": "__KUBECONFIG_FILEPATH__"
          }
        },
        {
          "type": "portmap",
          "snat": true,
          "capabilities": {"portMappings": true}
        }
      ]
    }

---
# Source: calico/templates/kdd-crds.yaml
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: felixconfigurations.crd.projectcalico.org
  labels:
    role.kubernetes.io/networking: "1"
spec:
  scope: Cluster
  group: crd.projectcalico.org
  version: v1
  names:
    kind: FelixConfiguration
    plural: felixconfigurations
    singular: felixconfiguration
---

apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: ipamblocks.crd.projectcalico.org
  labels:
    role.kubernetes.io/networking: "1"
spec:
  scope: Cluster
  group: crd.projectcalico.org
  version: v1
  names:
    kind: IPAMBlock
    plural: ipamblocks
    singular: ipamblock

---

apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: blockaffinities.crd.projectcalico.org
  labels:
    role.kubernetes.io/networking: "1"
spec:
  scope: Cluster
  group: crd.projectcalico.org
  version: v1
  names:
    kind: BlockAffinity
    plural: blockaffinities
    singular: blockaffinity

---

apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: ipamhandles.crd.projectcalico.org
  labels:
    role.kubernetes.io/networking: "1"
spec:
  scope: Cluster
  group: crd.projectcalico.org
  version: v1
  names:
    kind: IPAMHandle
    plural: ipamhandles
    singular: ipamhandle

---

apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: ipamconfigs.crd.projectcalico.org
  labels:
    role.kubernetes.io/networking: "1"
spec:
  scope: Cluster
  group: crd.projectcalico.org
  version: v1
  names:
    kind: IPAMConfig
    plural: ipamconfigs
    singular: ipamconfig

---

apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: bgppeers.crd.projectcalico.org
  labels:
    role.kubernetes.io/networking: "1"
spec:
  scope: Cluster
  group: crd.projectcalico.org
  version: v1
  names:
    kind: BGPPeer
    plural: bgppeers
    singular: bgppeer

---

apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: bgpconfigurations.crd.projectcalico.org
  labels:
    role.kubernetes.io/networking: "1"
spec:
  scope: Cluster
  group: crd.projectcalico.org
  version: v1
  names:
    kind: BGPConfiguration
    plural: bgpconfigurations
    singular: bgpconfiguration

---

apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: ippools.crd.projectcalico.org
  labels:
    role.kubernetes.io/networking: "1"
spec:
  scope: Cluster
  group: crd.projectcalico.org
  version: v1
  names:
    kind: IPPool
    plural: ippools
    singular: ippool

---

apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: hostendpoints.crd.projectcalico.org
  labels:
    role.kubernetes.io/networking: "1"
spec:
  scope: Cluster
  group: crd.projectcalico.org
  version: v1
  names:
    kind: HostEndpoint
    plural: hostendpoints
    singular: hostendpoint

---

apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: clusterinformations.crd.projectcalico.org
  labels:
    role.kubernetes.io/networking: "1"
spec:
  scope: Cluster
  group: crd.projectcalico.org
  version: v1
  names:
    kind: ClusterInformation
    plural: clusterinformations
    singular: clusterinformation

---

apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: globalnetworkpolicies.crd.projectcalico.org
  labels:
    role.kubernetes.io/networking: "1"
spec:
  scope: Cluster
  group: crd.projectcalico.org
  version: v1
  names:
    kind: GlobalNetworkPolicy
    plural: globalnetworkpolicies
    singular: globalnetworkpolicy

---

apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: globalnetworksets.crd.projectcalico.org
  labels:
    role.kubernetes.io/networking: "1"
spec:
  scope: Cluster
  group: crd.projectcalico.org
  version: v1
  names:
    kind: GlobalNetworkSet
    plural: globalnetworksets
    singular: globalnetworkset

---

apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: networkpolicies.crd.projectcalico.org
  labels:
    role.kubernetes.io/networking: "1"
spec:
  scope: Namespaced
  group: crd.projectcalico.org
  version: v1
  names:
    kind: NetworkPolicy
    plural: networkpolicies
    singular: networkpolicy

---

apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: networksets.crd.projectcalico.org
  labels:
    role.kubernetes.io/networking: "1"
spec:
  scope: Namespaced
  group: crd.projectcalico.org
  version: v1
  names:
    kind: NetworkSet
    plural: networksets
    singular: networkset
---
# Source: calico/templates/rbac.yaml

# Include a clusterrole for the kube-controllers component,
# and bind it to the calico-kube-controllers serviceaccount.
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: calico-kube-controllers
  labels:
    role.kubernetes.io/networking: "1"
rules:
  # Nodes are watched to monitor for deletions.
  - apiGroups: [""]
    resources:
      - nodes
    verbs:
      - watch
      - list
      - get
  # Pods are queried to check for existence.
  - apiGroups: [""]
    resources:
      - pods
    verbs:
      - get
  # IPAM resources are manipulated when nodes are deleted.
  - apiGroups: ["crd.projectcalico.org"]
    resources:
      - ippools
    verbs:
      - list
  - apiGroups: ["crd.projectcalico.org"]
    resources:
      - blockaffinities
      - ipamblocks
      - ipamhandles
    verbs:
      - get
      - list
      - create
      - update
      - delete
  # Needs access to update clusterinformations.
  - apiGroups: ["crd.projectcalico.org"]
    resources:
      - clusterinformations
    verbs:
      - get
      - create
      - update
---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: calico-kube-controllers
  labels:
    role.kubernetes.io/networking: "1"
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: calico-kube-controllers
subjects:
- kind: ServiceAccount
  name: calico-kube-controllers
  namespace: kube-system
---
# Include a clusterrole for the calico-node DaemonSet,
# and bind it to the calico-node serviceaccount.
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: calico-node
  labels:
    role.kubernetes.io/networking: "1"
rules:
  # The CNI plugin needs to get pods, nodes, and namespaces.
  - apiGroups: [""]
    resources:
      - pods
      - nodes
      - namespaces
    verbs:
      - get
  - apiGroups: [""]
    resources:
      - endpoints
      - services
    verbs:
      # Used to discover service IPs for advertisement.
      - watch
      - list
      # Used to discover Typhas.
      - get
  - apiGroups: [""]
    resources:
      - nodes/status
    verbs:
      # Needed for clearing NodeNetworkUnavailable flag.
      - patch
      # Calico stores some configuration information in node annotations.
      - update
  # Watch for changes to Kubernetes NetworkPolicies.
  - apiGroups: ["networking.k8s.io"]
    resources:
      - networkpolicies
    verbs:
      - watch
      - list
  # Used by Calico for policy information.
  - apiGroups: [""]
    resources:
      - pods
      - namespaces
      - serviceaccounts
    verbs:
      - list
      - watch
  # The CNI plugin patches pods/status.
  - apiGroups: [""]
    resources:
      - pods/status
    verbs:
      - patch
  # Calico monitors various CRDs for config.
  - apiGroups: ["crd.projectcalico.org"]
    resources:
      - globalfelixconfigs
      - felixconfigurations
      - bgppeers
      - globalbgpconfigs
      - bgpconfigurations
      - ippools
      - ipamblocks
      - globalnetworkpolicies
      - globalnetworksets
      - networkpolicies
      - networksets
      - clusterinformations
      - hostendpoints
      - blockaffinities
    verbs:
      - get
      - list
      - watch
  # Calico must create and update some CRDs on startup.
  - apiGroups: ["crd.projectcalico.org"]
    resources:
      - ippools
      - felixconfigurations
      - clusterinformations
    verbs:
      - create
      - update
  # Calico stores some configuration information on the node.
  - apiGroups: [""]
    resources:
      - nodes
    verbs:
      - get
      - list
      - watch
  # These permissions are only required for upgrade from v2.6, and can
  # be removed after upgrade or on fresh installations.
  - apiGroups: ["crd.projectcalico.org"]
    resources:
      - bgpconfigurations
      - bgppeers
    verbs:
      - create
      - update
  # These permissions are required for Calico CNI to perform IPAM allocations.
  - apiGroups: ["crd.projectcalico.org"]
    resources:
      - blockaffinities
      - ipamblocks
      - ipamhandles
    verbs:
      - get
      - list
      - create
      - update
      - delete
  - apiGroups: ["crd.projectcalico.org"]
    resources:
      - ipamconfigs
    verbs:
      - get
  # Block affinities must also be watchable by confd for route aggregation.
  - apiGroups: ["crd.projectcalico.org"]
    resources:
      - blockaffinities
    verbs:
      - watch
  # The Calico IPAM migration needs to get daemonsets. These permissions can be
  # removed if not upgrading from an installation using host-local IPAM.
  - apiGroups: ["apps"]
    resources:
      - daemonsets
    verbs:
      - get
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: calico-node
  labels:
    role.kubernetes.io/networking: "1"
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: calico-node
subjects:
- kind: ServiceAccount
  name: calico-node
  namespace: kube-system

{{ if .Networking.Calico.TyphaReplicas -}}
---
# Source: calico/templates/calico-typha.yaml
# This manifest creates a Service, which will be backed by Calico's Typha daemon.
# Typha sits in between Felix and the API server, reducing Calico's load on the API server.

apiVersion: v1
kind: Service
metadata:
  name: calico-typha
  namespace: kube-system
  labels:
    k8s-app: calico-typha
    role.kubernetes.io/networking: "1"
spec:
  ports:
    - port: 5473
      protocol: TCP
      targetPort: calico-typha
      name: calico-typha
  selector:
    k8s-app: calico-typha

---

# This manifest creates a Deployment of Typha to back the above service.

apiVersion: apps/v1
kind: Deployment
metadata:
  name: calico-typha
  namespace: kube-system
  labels:
    k8s-app: calico-typha
    role.kubernetes.io/networking: "1"
spec:
  # Number of Typha replicas.  To enable Typha, set this to a non-zero value *and* set the
  # typha_service_name variable in the calico-config ConfigMap above.
  #
  # We recommend using Typha if you have more than 50 nodes.  Above 100 nodes it is essential
  # (when using the Kubernetes datastore).  Use one replica for every 100-200 nodes.  In
  # production, we recommend running at least 3 replicas to reduce the impact of rolling upgrade.
  replicas: {{ or .Networking.Calico.TyphaReplicas "0" }}
  revisionHistoryLimit: 2
  selector:
    matchLabels:
      k8s-app: calico-typha
  template:
    metadata:
      labels:
        k8s-app: calico-typha
        role.kubernetes.io/networking: "1"
      annotations:
        # This, along with the CriticalAddonsOnly toleration below, marks the pod as a critical
        # add-on, ensuring it gets priority scheduling and that its resources are reserved
        # if it ever gets evicted.
        scheduler.alpha.kubernetes.io/critical-pod: ''
        cluster-autoscaler.kubernetes.io/safe-to-evict: 'true'
    spec:
      nodeSelector:
        beta.kubernetes.io/os: linux
      hostNetwork: true
      tolerations:
        # Mark the pod as a critical add-on for rescheduling.
        - key: CriticalAddonsOnly
          operator: Exists
      # Since Calico can't network a pod until Typha is up, we need to run Typha itself
      # as a host-networked pod.
      serviceAccountName: calico-node
      priorityClassName: system-cluster-critical
      containers:
      - image: calico/typha:v3.9.5
        name: calico-typha
        ports:
        - containerPort: 5473
          name: calico-typha
          protocol: TCP
        env:
          # Enable "info" logging by default.  Can be set to "debug" to increase verbosity.
          - name: TYPHA_LOGSEVERITYSCREEN
            value: "info"
          # Disable logging to file and syslog since those don't make sense in Kubernetes.
          - name: TYPHA_LOGFILEPATH
            value: "none"
          - name: TYPHA_LOGSEVERITYSYS
            value: "none"
          # Monitor the Kubernetes API to find the number of running instances and rebalance
          # connections.
          - name: TYPHA_CONNECTIONREBALANCINGMODE
            value: "kubernetes"
          - name: TYPHA_DATASTORETYPE
            value: "kubernetes"
          - name: TYPHA_HEALTHENABLED
            value: "true"
          # Uncomment these lines to enable prometheus metrics.  Since Typha is host-networked,
          # this opens a port on the host, which may need to be secured.
          - name: TYPHA_PROMETHEUSMETRICSENABLED
            value: "{{- or .Networking.Calico.TyphaPrometheusMetricsEnabled "false" }}"
          - name: TYPHA_PROMETHEUSMETRICSPORT
            value: "{{- or .Networking.Calico.TyphaPrometheusMetricsPort "9093" }}"
        livenessProbe:
          httpGet:
            path: /liveness
            port: 9098
            host: localhost
          periodSeconds: 30
          initialDelaySeconds: 30
        readinessProbe:
          httpGet:
            path: /readiness
            port: 9098
            host: localhost
          periodSeconds: 10

---

# This manifest creates a Pod Disruption Budget for Typha to allow K8s Cluster Autoscaler to evict

apiVersion: policy/v1beta1
kind: PodDisruptionBudget
metadata:
  name: calico-typha
  namespace: kube-system
  labels:
    k8s-app: calico-typha
    role.kubernetes.io/networking: "1"
spec:
  maxUnavailable: 1
  selector:
    matchLabels:
      k8s-app: calico-typha
{{- end -}}
---
# Source: calico/templates/calico-node.yaml
# This manifest installs the calico-node container, as well
# as the CNI plugins and network config on
# each master and worker node in a Kubernetes cluster.
kind: DaemonSet
apiVersion: apps/v1
metadata:
  name: calico-node
  namespace: kube-system
  labels:
    k8s-app: calico-node
    role.kubernetes.io/networking: "1"
spec:
  selector:
    matchLabels:
      k8s-app: calico-node
  updateStrategy:
    type: RollingUpdate
    rollingUpdate:
      maxUnavailable: 1
  template:
    metadata:
      labels:
        k8s-app: calico-node
        role.kubernetes.io/networking: "1"
      annotations:
        # This, along with the CriticalAddonsOnly toleration below,
        # marks the pod as a critical add-on, ensuring it gets
        # priority scheduling and that its resources are reserved
        # if it ever gets evicted.
        scheduler.alpha.kubernetes.io/critical-pod: ''
    spec:
      nodeSelector:
        beta.kubernetes.io/os: linux
      hostNetwork: true
      tolerations:
        # Make sure calico-node gets scheduled on all nodes.
        - effect: NoSchedule
          operator: Exists
        # Mark the pod as a critical add-on for rescheduling.
        - key: CriticalAddonsOnly
          operator: Exists
        - effect: NoExecute
          operator: Exists
      serviceAccountName: calico-node
      # Minimize downtime during a rolling upgrade or deletion; tell Kubernetes to do a "force
      # deletion": https://kubernetes.io/docs/concepts/workloads/pods/pod/#termination-of-pods.
      terminationGracePeriodSeconds: 0
      priorityClassName: system-node-critical
      initContainers:
        # This container performs upgrade from host-local IPAM to calico-ipam.
        # It can be deleted if this is a fresh installation, or if you have already
        # upgraded to use calico-ipam.
        - name: upgrade-ipam
          image: calico/cni:v3.9.5
          command: ["/opt/cni/bin/calico-ipam", "-upgrade"]
          env:
            - name: KUBERNETES_NODE_NAME
              valueFrom:
                fieldRef:
                  fieldPath: spec.nodeName
            - name: CALICO_NETWORKING_BACKEND
              valueFrom:
                configMapKeyRef:
                  name: calico-config
                  key: calico_backend
          volumeMounts:
            - mountPath: /var/lib/cni/networks
              name: host-local-net-dir
            - mountPath: /host/opt/cni/bin
              name: cni-bin-dir
        # This container installs the CNI binaries
        # and CNI network config file on each node.
        - name: install-cni
          image: calico/cni:v3.9.5
          command: ["/install-cni.sh"]
          env:
            # Name of the CNI config file to create.
            - name: CNI_CONF_NAME
              value: "10-calico.conflist"
            # The CNI network config to install on each node.
            - name: CNI_NETWORK_CONFIG
              valueFrom:
                configMapKeyRef:
                  name: calico-config
                  key: cni_network_config
            # Set the hostname based on the k8s node name.
            - name: KUBERNETES_NODE_NAME
              valueFrom:
                fieldRef:
                  fieldPath: spec.nodeName
            # CNI MTU Config variable
            - name: CNI_MTU
              valueFrom:
                configMapKeyRef:
                  name: calico-config
                  key: veth_mtu
            # Prevents the container from sleeping forever.
            - name: SLEEP
              value: "false"
          volumeMounts:
            - mountPath: /host/opt/cni/bin
              name: cni-bin-dir
            - mountPath: /host/etc/cni/net.d
              name: cni-net-dir
        # Adds a Flex Volume Driver that creates a per-pod Unix Domain Socket to allow Dikastes
        # to communicate with Felix over the Policy Sync API.
        - name: flexvol-driver
          image: calico/pod2daemon-flexvol:v3.9.5
          volumeMounts:
          - name: flexvol-driver-host
            mountPath: /host/driver
      containers:
        # Runs calico-node container on each Kubernetes node.  This
        # container programs network policy and routes on each
        # host.
        - name: calico-node
          image: calico/node:v3.9.5
          env:
            # Use Kubernetes API as the backing datastore.
            - name: DATASTORE_TYPE
              value: "kubernetes"
            # Typha support: controlled by the ConfigMap.
            - name: FELIX_TYPHAK8SSERVICENAME
              valueFrom:
                configMapKeyRef:
                  name: calico-config
                  key: typha_service_name
            # Wait for the datastore.
            - name: WAIT_FOR_DATASTORE
              value: "true"
            # Set based on the k8s node name.
            - name: NODENAME
              valueFrom:
                fieldRef:
                  fieldPath: spec.nodeName
            # Choose the backend to use.
            - name: CALICO_NETWORKING_BACKEND
              valueFrom:
                configMapKeyRef:
                  name: calico-config
                  key: calico_backend
            # Cluster type to identify the deployment type
            - name: CLUSTER_TYPE
              # was value: "k8s,bgp"
              value: "kops,bgp"
            # Auto-detect the BGP IP address.
            - name: IP
              value: "autodetect"
            # Enable IPIP
            - name: CALICO_IPV4POOL_IPIP
              value: "{{- if and (eq .CloudProvider "aws") (.Networking.Calico.CrossSubnet) -}}CrossSubnet{{- else -}} {{- or .Networking.Calico.IPIPMode "Always" -}} {{- end -}}"
            # Set MTU for tunnel device used if ipip is enabled
            - name: FELIX_IPINIPMTU
              valueFrom:
                configMapKeyRef:
                  name: calico-config
                  key: veth_mtu
            # The default IPv4 pool to create on startup if none exists. Pod IPs will be
            # chosen from this range. Changing this value after installation will have
            # no effect. This should fall within ` + "`" + `--cluster-cidr` + "`" + `.
            - name: CALICO_IPV4POOL_CIDR
              value: "{{ .KubeControllerManager.ClusterCIDR }}"
            # Disable file logging so ` + "`" + `kubectl logs` + "`" + ` works.
            - name: CALICO_DISABLE_FILE_LOGGING
              value: "true"
            # Set Felix endpoint to host default action to ACCEPT.
            - name: FELIX_DEFAULTENDPOINTTOHOSTACTION
              value: "ACCEPT"
            # Disable IPv6 on Kubernetes.
            - name: FELIX_IPV6SUPPORT
              value: "false"
            # Set Felix logging to the desired level
            - name: FELIX_LOGSEVERITYSCREEN
              value: "{{- or .Networking.Calico.LogSeverityScreen "info" }}"
            - name: FELIX_HEALTHENABLED
              value: "true"

            # kops additions
            # Set Felix iptables binary variant, Legacy or NFT
            - name: FELIX_IPTABLESBACKEND
              value: "{{- or .Networking.Calico.IptablesBackend "Legacy" }}"
            # Set to enable the experimental Prometheus metrics server
            - name: FELIX_PROMETHEUSMETRICSENABLED
              value: "{{- or .Networking.Calico.PrometheusMetricsEnabled "false" }}"
            # TCP port that the Prometheus metrics server should bind to
            - name: FELIX_PROMETHEUSMETRICSPORT
              value: "{{- or .Networking.Calico.PrometheusMetricsPort "9091" }}"
            # Enable Prometheus Go runtime metrics collection
            - name: FELIX_PROMETHEUSGOMETRICSENABLED
              value: "{{- or .Networking.Calico.PrometheusGoMetricsEnabled "true" }}"
            # Enable Prometheus process metrics collection
            - name: FELIX_PROMETHEUSPROCESSMETRICSENABLED
              value: "{{- or .Networking.Calico.PrometheusProcessMetricsEnabled "true" }}"
          securityContext:
            privileged: true
          resources:
            requests:
              cpu: 90m
          livenessProbe:
            exec:
              command:
              - /bin/calico-node
              - -felix-live
            periodSeconds: 10
            initialDelaySeconds: 10
            failureThreshold: 6
          readinessProbe:
            exec:
              command:
              - /bin/calico-node
              - -felix-ready
              - -bird-ready
            periodSeconds: 10
          volumeMounts:
            - mountPath: /lib/modules
              name: lib-modules
              readOnly: true
            - mountPath: /run/xtables.lock
              name: xtables-lock
              readOnly: false
            - mountPath: /var/run/calico
              name: var-run-calico
              readOnly: false
            - mountPath: /var/lib/calico
              name: var-lib-calico
              readOnly: false
            - name: policysync
              mountPath: /var/run/nodeagent
      volumes:
        # Used by calico-node.
        - name: lib-modules
          hostPath:
            path: /lib/modules
        - name: var-run-calico
          hostPath:
            path: /var/run/calico
        - name: var-lib-calico
          hostPath:
            path: /var/lib/calico
        - name: xtables-lock
          hostPath:
            path: /run/xtables.lock
            type: FileOrCreate
        # Used to install CNI.
        - name: cni-bin-dir
          hostPath:
            path: /opt/cni/bin
        - name: cni-net-dir
          hostPath:
            path: /etc/cni/net.d
        # Mount in the directory for host-local IPAM allocations. This is
        # used when upgrading from host-local to calico-ipam, and can be removed
        # if not using the upgrade-ipam init container.
        - name: host-local-net-dir
          hostPath:
            path: /var/lib/cni/networks
        # Used to create per-pod Unix Domain Sockets
        - name: policysync
          hostPath:
            type: DirectoryOrCreate
            path: /var/run/nodeagent
        # Used to install Flex Volume Driver
        - name: flexvol-driver-host
          hostPath:
            type: DirectoryOrCreate
            path: "{{- or .Kubelet.VolumePluginDirectory "/usr/libexec/kubernetes/kubelet-plugins/volume/exec/" }}nodeagent~uds"
---

apiVersion: v1
kind: ServiceAccount
metadata:
  name: calico-node
  namespace: kube-system
  labels:
    role.kubernetes.io/networking: "1"

---
# Source: calico/templates/calico-kube-controllers.yaml

# See https://github.com/projectcalico/kube-controllers
apiVersion: apps/v1
kind: Deployment
metadata:
  name: calico-kube-controllers
  namespace: kube-system
  labels:
    k8s-app: calico-kube-controllers
    role.kubernetes.io/networking: "1"
spec:
  # The controllers can only have a single active instance.
  replicas: 1
  selector:
    matchLabels:
      k8s-app: calico-kube-controllers
  strategy:
    type: Recreate
  template:
    metadata:
      name: calico-kube-controllers
      namespace: kube-system
      labels:
        k8s-app: calico-kube-controllers
        role.kubernetes.io/networking: "1"
      annotations:
        scheduler.alpha.kubernetes.io/critical-pod: ''
    spec:
      nodeSelector:
        beta.kubernetes.io/os: linux
      tolerations:
        # Mark the pod as a critical add-on for rescheduling.
        - key: CriticalAddonsOnly
          operator: Exists
        - key: node-role.kubernetes.io/master
          effect: NoSchedule
      serviceAccountName: calico-kube-controllers
      priorityClassName: system-cluster-critical
      containers:
        - name: calico-kube-controllers
          image: calico/kube-controllers:v3.9.5
          env:
            # Choose which controllers to run.
            - name: ENABLED_CONTROLLERS
              value: node
            - name: DATASTORE_TYPE
              value: kubernetes
          readinessProbe:
            exec:
              command:
              - /usr/bin/check-status
              - -r

---

apiVersion: v1
kind: ServiceAccount
metadata:
  name: calico-kube-controllers
  namespace: kube-system
  labels:
    role.kubernetes.io/networking: "1"

{{ if and (eq .CloudProvider "aws") (.Networking.Calico.CrossSubnet) -}}
# This manifest installs the k8s-ec2-srcdst container, which disables
# src/dst ip checks to allow BGP to function for calico for hosts within subnets
# This only applies for AWS environments.
---

kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: k8s-ec2-srcdst
  labels:
    role.kubernetes.io/networking: "1"
rules:
- apiGroups:
  - ""
  resources:
  - nodes
  verbs:
  - get
  - list
  - watch
  - update
  - patch

---

apiVersion: v1
kind: ServiceAccount
metadata:
  name: k8s-ec2-srcdst
  namespace: kube-system
  labels:
    role.kubernetes.io/networking: "1"
---

kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: k8s-ec2-srcdst
  labels:
    role.kubernetes.io/networking: "1"
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: k8s-ec2-srcdst
subjects:
- kind: ServiceAccount
  name: k8s-ec2-srcdst
  namespace: kube-system

---

apiVersion: apps/v1
kind: Deployment
metadata:
  name: k8s-ec2-srcdst
  namespace: kube-system
  labels:
    k8s-app: k8s-ec2-srcdst
    role.kubernetes.io/networking: "1"
spec:
  replicas: 1
  selector:
    matchLabels:
      k8s-app: k8s-ec2-srcdst
  template:
    metadata:
      labels:
        k8s-app: k8s-ec2-srcdst
        role.kubernetes.io/networking: "1"
      annotations:
        scheduler.alpha.kubernetes.io/critical-pod: ''
    spec:
      hostNetwork: true
      tolerations:
      - key: node-role.kubernetes.io/master
        effect: NoSchedule
      - key: CriticalAddonsOnly
        operator: Exists
      serviceAccountName: k8s-ec2-srcdst
      containers:
        - image: ottoyiu/k8s-ec2-srcdst:v0.2.2
          name: k8s-ec2-srcdst
          resources:
            requests:
              cpu: 10m
              memory: 64Mi
          env:
            - name: AWS_REGION
              value: {{ Region }}
          volumeMounts:
            - name: ssl-certs
              mountPath: "/etc/ssl/certs/ca-certificates.crt"
              readOnly: true
          imagePullPolicy: "Always"
      volumes:
        - name: ssl-certs
          hostPath:
            path: "/etc/ssl/certs/ca-certificates.crt"
      nodeSelector:
        node-role.kubernetes.io/master: ""
{{- end -}}
`)

func cloudupResourcesAddonsNetworkingProjectcalicoOrgK8s112YamlTemplateBytes() ([]byte, error) {
	return _cloudupResourcesAddonsNetworkingProjectcalicoOrgK8s112YamlTemplate, nil
}

func cloudupResourcesAddonsNetworkingProjectcalicoOrgK8s112YamlTemplate() (*asset, error) {
	bytes, err := cloudupResourcesAddonsNetworkingProjectcalicoOrgK8s112YamlTemplateBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "cloudup/resources/addons/networking.projectcalico.org/k8s-1.12.yaml.template", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _cloudupResourcesAddonsNetworkingProjectcalicoOrgK8s116YamlTemplate = []byte(`# Pulled and modified from: https://docs.projectcalico.org/v3.13/manifests/calico-typha.yaml

---
# Source: calico/templates/calico-config.yaml
# This ConfigMap is used to configure a self-hosted Calico installation.
kind: ConfigMap
apiVersion: v1
metadata:
  name: calico-config
  namespace: kube-system
  labels:
    role.kubernetes.io/networking: "1"
data:
  # You must set a non-zero value for Typha replicas below.
  typha_service_name: "{{- if .Networking.Calico.TyphaReplicas -}}calico-typha{{- else -}}none{{- end -}}"
  # Configure the backend to use.
  calico_backend: "bird"

  # Configure the MTU to use
  {{- if .Networking.Calico.MTU }}
  veth_mtu: "{{ .Networking.Calico.MTU }}"
  {{- else }}
  veth_mtu: "{{- if eq .CloudProvider "openstack" -}}1430{{- else -}}1440{{- end -}}"
  {{- end }}

  # The CNI network configuration to install on each node.  The special
  # values in this config will be automatically populated.
  cni_network_config: |-
    {
      "name": "k8s-pod-network",
      "cniVersion": "0.3.1",
      "plugins": [
        {
          "type": "calico",
          "log_level": "info",
          "datastore_type": "kubernetes",
          "nodename": "__KUBERNETES_NODE_NAME__",
          "mtu": __CNI_MTU__,
          "ipam": {
              "type": "calico-ipam"
          },
          "policy": {
              "type": "k8s"
          },
          "kubernetes": {
              "kubeconfig": "__KUBECONFIG_FILEPATH__"
          }
        },
        {
          "type": "portmap",
          "snat": true,
          "capabilities": {"portMappings": true}
        },
        {
          "type": "bandwidth",
          "capabilities": {"bandwidth": true}
        }
      ]
    }

---
# Source: calico/templates/kdd-crds.yaml

apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: bgpconfigurations.crd.projectcalico.org
  labels:
    role.kubernetes.io/networking: "1"
spec:
  scope: Cluster
  group: crd.projectcalico.org
  version: v1
  names:
    kind: BGPConfiguration
    plural: bgpconfigurations
    singular: bgpconfiguration

---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: bgppeers.crd.projectcalico.org
  labels:
    role.kubernetes.io/networking: "1"
spec:
  scope: Cluster
  group: crd.projectcalico.org
  version: v1
  names:
    kind: BGPPeer
    plural: bgppeers
    singular: bgppeer

---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: blockaffinities.crd.projectcalico.org
  labels:
    role.kubernetes.io/networking: "1"
spec:
  scope: Cluster
  group: crd.projectcalico.org
  version: v1
  names:
    kind: BlockAffinity
    plural: blockaffinities
    singular: blockaffinity

---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: clusterinformations.crd.projectcalico.org
  labels:
    role.kubernetes.io/networking: "1"
spec:
  scope: Cluster
  group: crd.projectcalico.org
  version: v1
  names:
    kind: ClusterInformation
    plural: clusterinformations
    singular: clusterinformation

---

apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: felixconfigurations.crd.projectcalico.org
  labels:
    role.kubernetes.io/networking: "1"
spec:
  scope: Cluster
  group: crd.projectcalico.org
  version: v1
  names:
    kind: FelixConfiguration
    plural: felixconfigurations
    singular: felixconfiguration

---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: globalnetworkpolicies.crd.projectcalico.org
  labels:
    role.kubernetes.io/networking: "1"
spec:
  scope: Cluster
  group: crd.projectcalico.org
  version: v1
  names:
    kind: GlobalNetworkPolicy
    plural: globalnetworkpolicies
    singular: globalnetworkpolicy

---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: globalnetworksets.crd.projectcalico.org
  labels:
    role.kubernetes.io/networking: "1"
spec:
  scope: Cluster
  group: crd.projectcalico.org
  version: v1
  names:
    kind: GlobalNetworkSet
    plural: globalnetworksets
    singular: globalnetworkset

---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: hostendpoints.crd.projectcalico.org
  labels:
    role.kubernetes.io/networking: "1"
spec:
  scope: Cluster
  group: crd.projectcalico.org
  version: v1
  names:
    kind: HostEndpoint
    plural: hostendpoints
    singular: hostendpoint

---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: ipamblocks.crd.projectcalico.org
  labels:
    role.kubernetes.io/networking: "1"
spec:
  scope: Cluster
  group: crd.projectcalico.org
  version: v1
  names:
    kind: IPAMBlock
    plural: ipamblocks
    singular: ipamblock

---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: ipamconfigs.crd.projectcalico.org
  labels:
    role.kubernetes.io/networking: "1"
spec:
  scope: Cluster
  group: crd.projectcalico.org
  version: v1
  names:
    kind: IPAMConfig
    plural: ipamconfigs
    singular: ipamconfig

---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: ipamhandles.crd.projectcalico.org
  labels:
    role.kubernetes.io/networking: "1"
spec:
  scope: Cluster
  group: crd.projectcalico.org
  version: v1
  names:
    kind: IPAMHandle
    plural: ipamhandles
    singular: ipamhandle

---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: ippools.crd.projectcalico.org
  labels:
    role.kubernetes.io/networking: "1"
spec:
  scope: Cluster
  group: crd.projectcalico.org
  version: v1
  names:
    kind: IPPool
    plural: ippools
    singular: ippool

---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: networkpolicies.crd.projectcalico.org
  labels:
    role.kubernetes.io/networking: "1"
spec:
  scope: Namespaced
  group: crd.projectcalico.org
  version: v1
  names:
    kind: NetworkPolicy
    plural: networkpolicies
    singular: networkpolicy

---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: networksets.crd.projectcalico.org
  labels:
    role.kubernetes.io/networking: "1"
spec:
  scope: Namespaced
  group: crd.projectcalico.org
  version: v1
  names:
    kind: NetworkSet
    plural: networksets
    singular: networkset

---
# Source: calico/templates/rbac.yaml

# Include a clusterrole for the kube-controllers component,
# and bind it to the calico-kube-controllers serviceaccount.
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: calico-kube-controllers
  labels:
    role.kubernetes.io/networking: "1"
rules:
  # Nodes are watched to monitor for deletions.
  - apiGroups: [""]
    resources:
      - nodes
    verbs:
      - watch
      - list
      - get
  # Pods are queried to check for existence.
  - apiGroups: [""]
    resources:
      - pods
    verbs:
      - get
  # IPAM resources are manipulated when nodes are deleted.
  - apiGroups: ["crd.projectcalico.org"]
    resources:
      - ippools
    verbs:
      - list
  - apiGroups: ["crd.projectcalico.org"]
    resources:
      - blockaffinities
      - ipamblocks
      - ipamhandles
    verbs:
      - get
      - list
      - create
      - update
      - delete
  # Needs access to update clusterinformations.
  - apiGroups: ["crd.projectcalico.org"]
    resources:
      - clusterinformations
    verbs:
      - get
      - create
      - update
---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: calico-kube-controllers
  labels:
    role.kubernetes.io/networking: "1"
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: calico-kube-controllers
subjects:
- kind: ServiceAccount
  name: calico-kube-controllers
  namespace: kube-system
---
# Include a clusterrole for the calico-node DaemonSet,
# and bind it to the calico-node serviceaccount.
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: calico-node
  labels:
    role.kubernetes.io/networking: "1"
rules:
  # The CNI plugin needs to get pods, nodes, and namespaces.
  - apiGroups: [""]
    resources:
      - pods
      - nodes
      - namespaces
    verbs:
      - get
  - apiGroups: [""]
    resources:
      - endpoints
      - services
    verbs:
      # Used to discover service IPs for advertisement.
      - watch
      - list
      # Used to discover Typhas.
      - get
  # Pod CIDR auto-detection on kubeadm needs access to config maps.
  - apiGroups: [""]
    resources:
      - configmaps
    verbs:
      - get
  - apiGroups: [""]
    resources:
      - nodes/status
    verbs:
      # Needed for clearing NodeNetworkUnavailable flag.
      - patch
      # Calico stores some configuration information in node annotations.
      - update
  # Watch for changes to Kubernetes NetworkPolicies.
  - apiGroups: ["networking.k8s.io"]
    resources:
      - networkpolicies
    verbs:
      - watch
      - list
  # Used by Calico for policy information.
  - apiGroups: [""]
    resources:
      - pods
      - namespaces
      - serviceaccounts
    verbs:
      - list
      - watch
  # The CNI plugin patches pods/status.
  - apiGroups: [""]
    resources:
      - pods/status
    verbs:
      - patch
  # Calico monitors various CRDs for config.
  - apiGroups: ["crd.projectcalico.org"]
    resources:
      - globalfelixconfigs
      - felixconfigurations
      - bgppeers
      - globalbgpconfigs
      - bgpconfigurations
      - ippools
      - ipamblocks
      - globalnetworkpolicies
      - globalnetworksets
      - networkpolicies
      - networksets
      - clusterinformations
      - hostendpoints
      - blockaffinities
    verbs:
      - get
      - list
      - watch
  # Calico must create and update some CRDs on startup.
  - apiGroups: ["crd.projectcalico.org"]
    resources:
      - ippools
      - felixconfigurations
      - clusterinformations
    verbs:
      - create
      - update
  # Calico stores some configuration information on the node.
  - apiGroups: [""]
    resources:
      - nodes
    verbs:
      - get
      - list
      - watch
  # These permissions are only required for upgrade from v2.6, and can
  # be removed after upgrade or on fresh installations.
  - apiGroups: ["crd.projectcalico.org"]
    resources:
      - bgpconfigurations
      - bgppeers
    verbs:
      - create
      - update
  # These permissions are required for Calico CNI to perform IPAM allocations.
  - apiGroups: ["crd.projectcalico.org"]
    resources:
      - blockaffinities
      - ipamblocks
      - ipamhandles
    verbs:
      - get
      - list
      - create
      - update
      - delete
  - apiGroups: ["crd.projectcalico.org"]
    resources:
      - ipamconfigs
    verbs:
      - get
  # Block affinities must also be watchable by confd for route aggregation.
  - apiGroups: ["crd.projectcalico.org"]
    resources:
      - blockaffinities
    verbs:
      - watch
  # The Calico IPAM migration needs to get daemonsets. These permissions can be
  # removed if not upgrading from an installation using host-local IPAM.
  - apiGroups: ["apps"]
    resources:
      - daemonsets
    verbs:
      - get

---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: calico-node
  labels:
    role.kubernetes.io/networking: "1"
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: calico-node
subjects:
- kind: ServiceAccount
  name: calico-node
  namespace: kube-system

{{ if .Networking.Calico.TyphaReplicas -}}
---
# Source: calico/templates/calico-typha.yaml
# This manifest creates a Service, which will be backed by Calico's Typha daemon.
# Typha sits in between Felix and the API server, reducing Calico's load on the API server.

apiVersion: v1
kind: Service
metadata:
  name: calico-typha
  namespace: kube-system
  labels:
    k8s-app: calico-typha
    role.kubernetes.io/networking: "1"
spec:
  ports:
    - port: 5473
      protocol: TCP
      targetPort: calico-typha
      name: calico-typha
  selector:
    k8s-app: calico-typha

---

# This manifest creates a Deployment of Typha to back the above service.

apiVersion: apps/v1
kind: Deployment
metadata:
  name: calico-typha
  namespace: kube-system
  labels:
    k8s-app: calico-typha
    role.kubernetes.io/networking: "1"
spec:
  # Number of Typha replicas.  To enable Typha, set this to a non-zero value *and* set the
  # typha_service_name variable in the calico-config ConfigMap above.
  #
  # We recommend using Typha if you have more than 50 nodes.  Above 100 nodes it is essential
  # (when using the Kubernetes datastore).  Use one replica for every 100-200 nodes.  In
  # production, we recommend running at least 3 replicas to reduce the impact of rolling upgrade.
  replicas: {{ or .Networking.Calico.TyphaReplicas "0" }}
  revisionHistoryLimit: 2
  selector:
    matchLabels:
      k8s-app: calico-typha
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxUnavailable: 1
  template:
    metadata:
      labels:
        k8s-app: calico-typha
        role.kubernetes.io/networking: "1"
      annotations:
        cluster-autoscaler.kubernetes.io/safe-to-evict: 'true'
    spec:
      nodeSelector:
        kubernetes.io/os: linux
        kubernetes.io/role: master
      hostNetwork: true
      tolerations:
        # Mark the pod as a critical add-on for rescheduling.
        - key: CriticalAddonsOnly
          operator: Exists
        - key: "node-role.kubernetes.io/master"
          effect: NoSchedule
      # Since Calico can't network a pod until Typha is up, we need to run Typha itself
      # as a host-networked pod.
      serviceAccountName: calico-node
      priorityClassName: system-cluster-critical
      # fsGroup allows using projected serviceaccount tokens as described here kubernetes/kubernetes#82573
      securityContext:
        fsGroup: 65534
      containers:
      - image: calico/typha:v3.13.2
        name: calico-typha
        ports:
        - containerPort: 5473
          name: calico-typha
          protocol: TCP
        env:
          # Enable "info" logging by default.  Can be set to "debug" to increase verbosity.
          - name: TYPHA_LOGSEVERITYSCREEN
            value: "info"
          # Disable logging to file and syslog since those don't make sense in Kubernetes.
          - name: TYPHA_LOGFILEPATH
            value: "none"
          - name: TYPHA_LOGSEVERITYSYS
            value: "none"
          # Monitor the Kubernetes API to find the number of running instances and rebalance
          # connections.
          - name: TYPHA_CONNECTIONREBALANCINGMODE
            value: "kubernetes"
          - name: TYPHA_DATASTORETYPE
            value: "kubernetes"
          - name: TYPHA_HEALTHENABLED
            value: "true"
          - name: TYPHA_PROMETHEUSMETRICSENABLED
            value: "{{- or .Networking.Calico.TyphaPrometheusMetricsEnabled "false" }}"
          - name: TYPHA_PROMETHEUSMETRICSPORT
            value: "{{- or .Networking.Calico.TyphaPrometheusMetricsPort "9093" }}"
        livenessProbe:
          httpGet:
            path: /liveness
            port: 9098
            host: localhost
          periodSeconds: 30
          initialDelaySeconds: 30
        securityContext:
          runAsNonRoot: true
          allowPrivilegeEscalation: false
        readinessProbe:
          httpGet:
            path: /readiness
            port: 9098
            host: localhost
          periodSeconds: 10

---

# This manifest creates a Pod Disruption Budget for Typha to allow K8s Cluster Autoscaler to evict

apiVersion: policy/v1beta1
kind: PodDisruptionBudget
metadata:
  name: calico-typha
  namespace: kube-system
  labels:
    k8s-app: calico-typha
    role.kubernetes.io/networking: "1"
spec:
  maxUnavailable: 1
  selector:
    matchLabels:
      k8s-app: calico-typha
{{- end }}

---
# Source: calico/templates/calico-node.yaml
# This manifest installs the calico-node container, as well
# as the CNI plugins and network config on
# each master and worker node in a Kubernetes cluster.
kind: DaemonSet
apiVersion: apps/v1
metadata:
  name: calico-node
  namespace: kube-system
  labels:
    k8s-app: calico-node
    role.kubernetes.io/networking: "1"
spec:
  selector:
    matchLabels:
      k8s-app: calico-node
  updateStrategy:
    type: RollingUpdate
    rollingUpdate:
      maxUnavailable: 1
  template:
    metadata:
      labels:
        k8s-app: calico-node
        role.kubernetes.io/networking: "1"
    spec:
      nodeSelector:
        kubernetes.io/os: linux
      hostNetwork: true
      tolerations:
        # Make sure calico-node gets scheduled on all nodes.
        - effect: NoSchedule
          operator: Exists
        # Mark the pod as a critical add-on for rescheduling.
        - key: CriticalAddonsOnly
          operator: Exists
        - effect: NoExecute
          operator: Exists
      serviceAccountName: calico-node
      # Minimize downtime during a rolling upgrade or deletion; tell Kubernetes to do a "force
      # deletion": https://kubernetes.io/docs/concepts/workloads/pods/pod/#termination-of-pods.
      terminationGracePeriodSeconds: 0
      priorityClassName: system-node-critical
      initContainers:
        # This container performs upgrade from host-local IPAM to calico-ipam.
        # It can be deleted if this is a fresh installation, or if you have already
        # upgraded to use calico-ipam.
        - name: upgrade-ipam
          image: calico/cni:v3.13.2
          command: ["/opt/cni/bin/calico-ipam", "-upgrade"]
          env:
            - name: KUBERNETES_NODE_NAME
              valueFrom:
                fieldRef:
                  fieldPath: spec.nodeName
            - name: CALICO_NETWORKING_BACKEND
              valueFrom:
                configMapKeyRef:
                  name: calico-config
                  key: calico_backend
          volumeMounts:
            - mountPath: /var/lib/cni/networks
              name: host-local-net-dir
            - mountPath: /host/opt/cni/bin
              name: cni-bin-dir
          securityContext:
            privileged: true
        # This container installs the CNI binaries
        # and CNI network config file on each node.
        - name: install-cni
          image: calico/cni:v3.13.2
          command: ["/install-cni.sh"]
          env:
            # Name of the CNI config file to create.
            - name: CNI_CONF_NAME
              value: "10-calico.conflist"
            # The CNI network config to install on each node.
            - name: CNI_NETWORK_CONFIG
              valueFrom:
                configMapKeyRef:
                  name: calico-config
                  key: cni_network_config
            # Set the hostname based on the k8s node name.
            - name: KUBERNETES_NODE_NAME
              valueFrom:
                fieldRef:
                  fieldPath: spec.nodeName
            # CNI MTU Config variable
            - name: CNI_MTU
              valueFrom:
                configMapKeyRef:
                  name: calico-config
                  key: veth_mtu
            # Prevents the container from sleeping forever.
            - name: SLEEP
              value: "false"
          volumeMounts:
            - mountPath: /host/opt/cni/bin
              name: cni-bin-dir
            - mountPath: /host/etc/cni/net.d
              name: cni-net-dir
          securityContext:
            privileged: true
        # Adds a Flex Volume Driver that creates a per-pod Unix Domain Socket to allow Dikastes
        # to communicate with Felix over the Policy Sync API.
        - name: flexvol-driver
          image: calico/pod2daemon-flexvol:v3.13.2
          volumeMounts:
          - name: flexvol-driver-host
            mountPath: /host/driver
          securityContext:
            privileged: true
      containers:
        # Runs calico-node container on each Kubernetes node.  This
        # container programs network policy and routes on each
        # host.
        - name: calico-node
          image: calico/node:v3.13.2
          env:
            # Use Kubernetes API as the backing datastore.
            - name: DATASTORE_TYPE
              value: "kubernetes"
            {{- if .Networking.Calico.TyphaReplicas }}
            # Typha support: controlled by the ConfigMap.
            - name: FELIX_TYPHAK8SSERVICENAME
              valueFrom:
                configMapKeyRef:
                  name: calico-config
                  key: typha_service_name
            {{- end }}
            # Wait for the datastore.
            - name: WAIT_FOR_DATASTORE
              value: "true"
            # Set based on the k8s node name.
            - name: NODENAME
              valueFrom:
                fieldRef:
                  fieldPath: spec.nodeName
            # Choose the backend to use.
            - name: CALICO_NETWORKING_BACKEND
              valueFrom:
                configMapKeyRef:
                  name: calico-config
                  key: calico_backend
            # Cluster type to identify the deployment type
            - name: CLUSTER_TYPE
              value: "kops,bgp"
            # Auto-detect the BGP IP address.
            - name: IP
              value: "autodetect"
            # Enable IPIP
            - name: CALICO_IPV4POOL_IPIP
              value: "{{- if and (eq .CloudProvider "aws") (.Networking.Calico.CrossSubnet) -}}CrossSubnet{{- else -}} {{- or .Networking.Calico.IPIPMode "Always" -}} {{- end -}}"
            # Set MTU for tunnel device used if ipip is enabled
            - name: FELIX_IPINIPMTU
              valueFrom:
                configMapKeyRef:
                  name: calico-config
                  key: veth_mtu
            # The default IPv4 pool to create on startup if none exists. Pod IPs will be
            # chosen from this range. Changing this value after installation will have
            # no effect. This should fall within ` + "`" + `--cluster-cidr` + "`" + `.
            - name: CALICO_IPV4POOL_CIDR
              value: "{{ .KubeControllerManager.ClusterCIDR }}"
            # Disable file logging so ` + "`" + `kubectl logs` + "`" + ` works.
            - name: CALICO_DISABLE_FILE_LOGGING
              value: "true"
            # Set Felix endpoint to host default action to ACCEPT.
            - name: FELIX_DEFAULTENDPOINTTOHOSTACTION
              value: "ACCEPT"
            # Disable IPv6 on Kubernetes.
            - name: FELIX_IPV6SUPPORT
              value: "false"
            # Set Felix logging to "info"
            - name: FELIX_LOGSEVERITYSCREEN
              value: "{{- or .Networking.Calico.LogSeverityScreen "info" }}"
            - name: FELIX_HEALTHENABLED
              value: "true"

            # kops additions
            # Set Felix iptables binary variant, Legacy or NFT
            - name: FELIX_IPTABLESBACKEND
              value: "{{- or .Networking.Calico.IptablesBackend "Auto" }}"
            # Set to enable the experimental Prometheus metrics server
            - name: FELIX_PROMETHEUSMETRICSENABLED
              value: "{{- or .Networking.Calico.PrometheusMetricsEnabled "false" }}"
            # TCP port that the Prometheus metrics server should bind to
            - name: FELIX_PROMETHEUSMETRICSPORT
              value: "{{- or .Networking.Calico.PrometheusMetricsPort "9091" }}"
            # Enable Prometheus Go runtime metrics collection
            - name: FELIX_PROMETHEUSGOMETRICSENABLED
              value: "{{- or .Networking.Calico.PrometheusGoMetricsEnabled "true" }}"
            # Enable Prometheus process metrics collection
            - name: FELIX_PROMETHEUSPROCESSMETRICSENABLED
              value: "{{- or .Networking.Calico.PrometheusProcessMetricsEnabled "true" }}"
          securityContext:
            privileged: true
          resources:
            requests:
              cpu: 90m
          livenessProbe:
            exec:
              command:
              - /bin/calico-node
              - -felix-live
              - -bird-live
            periodSeconds: 10
            initialDelaySeconds: 10
            failureThreshold: 6
          readinessProbe:
            exec:
              command:
              - /bin/calico-node
              - -felix-ready
              - -bird-ready
            periodSeconds: 10
          volumeMounts:
            - mountPath: /lib/modules
              name: lib-modules
              readOnly: true
            - mountPath: /run/xtables.lock
              name: xtables-lock
              readOnly: false
            - mountPath: /var/run/calico
              name: var-run-calico
              readOnly: false
            - mountPath: /var/lib/calico
              name: var-lib-calico
              readOnly: false
            - name: policysync
              mountPath: /var/run/nodeagent
      volumes:
        # Used by calico-node.
        - name: lib-modules
          hostPath:
            path: /lib/modules
        - name: var-run-calico
          hostPath:
            path: /var/run/calico
        - name: var-lib-calico
          hostPath:
            path: /var/lib/calico
        - name: xtables-lock
          hostPath:
            path: /run/xtables.lock
            type: FileOrCreate
        # Used to install CNI.
        - name: cni-bin-dir
          hostPath:
            path: /opt/cni/bin
        - name: cni-net-dir
          hostPath:
            path: /etc/cni/net.d
        # Mount in the directory for host-local IPAM allocations. This is
        # used when upgrading from host-local to calico-ipam, and can be removed
        # if not using the upgrade-ipam init container.
        - name: host-local-net-dir
          hostPath:
            path: /var/lib/cni/networks
        # Used to create per-pod Unix Domain Sockets
        - name: policysync
          hostPath:
            type: DirectoryOrCreate
            path: /var/run/nodeagent
        # Used to install Flex Volume Driver
        - name: flexvol-driver-host
          hostPath:
            type: DirectoryOrCreate
            path: "{{- or .Kubelet.VolumePluginDirectory "/usr/libexec/kubernetes/kubelet-plugins/volume/exec/" }}nodeagent~uds"
---

apiVersion: v1
kind: ServiceAccount
metadata:
  name: calico-node
  namespace: kube-system
  labels:
    role.kubernetes.io/networking: "1"

---
# Source: calico/templates/calico-kube-controllers.yaml

# See https://github.com/projectcalico/kube-controllers
apiVersion: apps/v1
kind: Deployment
metadata:
  name: calico-kube-controllers
  namespace: kube-system
  labels:
    k8s-app: calico-kube-controllers
    role.kubernetes.io/networking: "1"
spec:
  # The controllers can only have a single active instance.
  replicas: 1
  selector:
    matchLabels:
      k8s-app: calico-kube-controllers
  strategy:
    type: Recreate
  template:
    metadata:
      name: calico-kube-controllers
      namespace: kube-system
      labels:
        k8s-app: calico-kube-controllers
        role.kubernetes.io/networking: "1"
      annotations:
        scheduler.alpha.kubernetes.io/critical-pod: ''
    spec:
      nodeSelector:
        kubernetes.io/os: linux
      tolerations:
        # Mark the pod as a critical add-on for rescheduling.
        - key: CriticalAddonsOnly
          operator: Exists
        - key: node-role.kubernetes.io/master
          effect: NoSchedule
      serviceAccountName: calico-kube-controllers
      priorityClassName: system-cluster-critical
      containers:
        - name: calico-kube-controllers
          image: calico/kube-controllers:v3.13.2
          env:
            # Choose which controllers to run.
            - name: ENABLED_CONTROLLERS
              value: node
            - name: DATASTORE_TYPE
              value: kubernetes
          readinessProbe:
            exec:
              command:
              - /usr/bin/check-status
              - -r

---

apiVersion: v1
kind: ServiceAccount
metadata:
  name: calico-kube-controllers
  namespace: kube-system
  labels:
    role.kubernetes.io/networking: "1"

{{ if and (eq .CloudProvider "aws") (.Networking.Calico.CrossSubnet) -}}
# This manifest installs the k8s-ec2-srcdst container, which disables
# src/dst ip checks to allow BGP to function for calico for hosts within subnets
# This only applies for AWS environments.
---

kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: k8s-ec2-srcdst
  labels:
    role.kubernetes.io/networking: "1"
rules:
- apiGroups:
  - ""
  resources:
  - nodes
  verbs:
  - get
  - list
  - watch
  - update
  - patch

---

apiVersion: v1
kind: ServiceAccount
metadata:
  name: k8s-ec2-srcdst
  namespace: kube-system
  labels:
    role.kubernetes.io/networking: "1"
---

kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: k8s-ec2-srcdst
  labels:
    role.kubernetes.io/networking: "1"
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: k8s-ec2-srcdst
subjects:
- kind: ServiceAccount
  name: k8s-ec2-srcdst
  namespace: kube-system

---

apiVersion: apps/v1
kind: Deployment
metadata:
  name: k8s-ec2-srcdst
  namespace: kube-system
  labels:
    k8s-app: k8s-ec2-srcdst
    role.kubernetes.io/networking: "1"
spec:
  replicas: 1
  selector:
    matchLabels:
      k8s-app: k8s-ec2-srcdst
  template:
    metadata:
      labels:
        k8s-app: k8s-ec2-srcdst
        role.kubernetes.io/networking: "1"
      annotations:
        scheduler.alpha.kubernetes.io/critical-pod: ''
    spec:
      hostNetwork: true
      tolerations:
      - key: node-role.kubernetes.io/master
        effect: NoSchedule
      - key: CriticalAddonsOnly
        operator: Exists
      serviceAccountName: k8s-ec2-srcdst
      priorityClassName: system-cluster-critical
      containers:
        - image: ottoyiu/k8s-ec2-srcdst:v0.2.2
          name: k8s-ec2-srcdst
          resources:
            requests:
              cpu: 10m
              memory: 64Mi
          env:
            - name: AWS_REGION
              value: {{ Region }}
          volumeMounts:
            - name: ssl-certs
              mountPath: "/etc/ssl/certs/ca-certificates.crt"
              readOnly: true
          imagePullPolicy: "Always"
      volumes:
        - name: ssl-certs
          hostPath:
            path: "/etc/ssl/certs/ca-certificates.crt"
      nodeSelector:
        node-role.kubernetes.io/master: ""
{{ end -}}
`)

func cloudupResourcesAddonsNetworkingProjectcalicoOrgK8s116YamlTemplateBytes() ([]byte, error) {
	return _cloudupResourcesAddonsNetworkingProjectcalicoOrgK8s116YamlTemplate, nil
}

func cloudupResourcesAddonsNetworkingProjectcalicoOrgK8s116YamlTemplate() (*asset, error) {
	bytes, err := cloudupResourcesAddonsNetworkingProjectcalicoOrgK8s116YamlTemplateBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "cloudup/resources/addons/networking.projectcalico.org/k8s-1.16.yaml.template", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _cloudupResourcesAddonsNetworkingProjectcalicoOrgK8s17V3YamlTemplate = []byte(`{{- $etcd_scheme := EtcdScheme }}
# This ConfigMap is used to configure a self-hosted Calico installation.
kind: ConfigMap
apiVersion: v1
metadata:
  name: calico-config
  namespace: kube-system
data:
  # The calico-etcd PetSet service IP:port
  etcd_endpoints: "{{ $cluster := index .EtcdClusters 0 -}}
                      {{- range $j, $member := $cluster.Members -}}
                          {{- if $j }},{{ end -}}
                          {{ $etcd_scheme }}://etcd-{{ $member.Name }}.internal.{{ ClusterName }}:4001
                      {{- end }}"

  # Configure the Calico backend to use.
  calico_backend: "bird"

  # The CNI network configuration to install on each node.
  cni_network_config: |-
    {
      "name": "k8s-pod-network",
      "cniVersion": "0.3.1",
      "plugins": [
        {
          "type": "calico",
          "etcd_endpoints": "__ETCD_ENDPOINTS__",
          {{- if eq $etcd_scheme "https" }}
          "etcd_ca_cert_file": "/srv/kubernetes/calico/ca.pem",
          "etcd_cert_file": "/srv/kubernetes/calico/calico-client.pem",
          "etcd_key_file": "/srv/kubernetes/calico/calico-client-key.pem",
          "etcd_scheme": "https",
          {{- end }}
          "log_level": "info",
          "ipam": {
            "type": "calico-ipam"
          },
          "policy": {
            "type": "k8s"
          },
          "kubernetes": {
            "kubeconfig": "/etc/cni/net.d/__KUBECONFIG_FILENAME__"
          }
        },
        {
          "type": "portmap",
          "snat": true,
          "capabilities": {"portMappings": true}
        }
      ]
    }

---

kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: calico-node
  labels:
    role.kubernetes.io/networking: "1"
rules:
  - apiGroups: [""]
    resources:
      - pods
      - nodes
      - namespaces
    verbs:
      - get
---

apiVersion: v1
kind: ServiceAccount
metadata:
  name: calico-node
  namespace: kube-system
  labels:
    role.kubernetes.io/networking: "1"
---

kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: calico-node
  labels:
    role.kubernetes.io/networking: "1"
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: calico-node
subjects:
- kind: ServiceAccount
  name: calico-node
  namespace: kube-system
---

apiVersion: v1
kind: ServiceAccount
metadata:
  name: calico-kube-controllers
  namespace: kube-system
  labels:
    role.kubernetes.io/networking: "1"
---

kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: calico-kube-controllers
  labels:
    role.kubernetes.io/networking: "1"
rules:
  - apiGroups:
    - ""
    - extensions
    resources:
      - pods
      - namespaces
      - networkpolicies
      - nodes
    verbs:
      - watch
      - list
  - apiGroups:
    - networking.k8s.io
    resources:
      - networkpolicies
    verbs:
      - watch
      - list
---

kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: calico-kube-controllers
  labels:
    role.kubernetes.io/networking: "1"
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: calico-kube-controllers
subjects:
- kind: ServiceAccount
  name: calico-kube-controllers
  namespace: kube-system

---

# This manifest installs the calico/node container, as well
# as the Calico CNI plugins and network config on
# each master and worker node in a Kubernetes cluster.
kind: DaemonSet
apiVersion: extensions/v1beta1
metadata:
  name: calico-node
  namespace: kube-system
  labels:
    k8s-app: calico-node
    role.kubernetes.io/networking: "1"
spec:
  selector:
    matchLabels:
      k8s-app: calico-node
  updateStrategy:
    type: RollingUpdate
    rollingUpdate:
      maxUnavailable: 1
  template:
    metadata:
      labels:
        k8s-app: calico-node
        role.kubernetes.io/networking: "1"
      annotations:
        scheduler.alpha.kubernetes.io/critical-pod: ''
    spec:
      hostNetwork: true
      tolerations:
        # Make sure calico/node gets scheduled on all nodes.
        - effect: NoSchedule
          operator: Exists
        # Mark the pod as a critical add-on for rescheduling.
        - key: CriticalAddonsOnly
          operator: Exists
        - effect: NoExecute
          operator: Exists
      serviceAccountName: calico-node
      # Minimize downtime during a rolling upgrade or deletion; tell Kubernetes to do a "force
      # deletion": https://kubernetes.io/docs/concepts/workloads/pods/pod/#termination-of-pods.
      terminationGracePeriodSeconds: 0
      containers:
        # Runs calico/node container on each Kubernetes node.  This
        # container programs network policy and routes on each
        # host.
        - name: calico-node
          image: calico/node:v3.8.0
          env:
            # The location of the Calico etcd cluster.
            - name: ETCD_ENDPOINTS
              valueFrom:
                configMapKeyRef:
                  name: calico-config
                  key: etcd_endpoints
            {{- if eq $etcd_scheme "https" }}
            - name: ETCD_CERT_FILE
              value: /certs/calico-client.pem
            - name: ETCD_KEY_FILE
              value: /certs/calico-client-key.pem
            - name: ETCD_CA_CERT_FILE
              value: /certs/ca.pem
            {{- end }}
            # Choose the backend to use.
            - name: CALICO_NETWORKING_BACKEND
              valueFrom:
                configMapKeyRef:
                  name: calico-config
                  key: calico_backend
            # Cluster type to identify the deployment type
            - name: CLUSTER_TYPE
              value: "kops,bgp"
            # Disable file logging so ` + "`" + `kubectl logs` + "`" + ` works.
            - name: CALICO_DISABLE_FILE_LOGGING
              value: "true"
            # Set noderef for node controller.
            - name: CALICO_K8S_NODE_REF
              valueFrom:
                fieldRef:
                  fieldPath: spec.nodeName
            # Set Felix endpoint to host default action to ACCEPT.
            - name: FELIX_DEFAULTENDPOINTTOHOSTACTION
              value: "ACCEPT"
            # The default IPv4 pool to create on startup if none exists. Pod IPs will be
            # chosen from this range. Changing this value after installation will have
            # no effect. This should fall within ` + "`" + `--cluster-cidr` + "`" + `.
            # Configure the IP Pool from which Pod IPs will be chosen.
            - name: CALICO_IPV4POOL_CIDR
              value: "{{ .KubeControllerManager.ClusterCIDR }}"
            - name: CALICO_IPV4POOL_IPIP
              value: "{{- if and (eq .CloudProvider "aws") (.Networking.Calico.CrossSubnet) -}}CrossSubnet{{- else -}}Always{{- end -}}"
            # Disable IPv6 on Kubernetes.
            - name: FELIX_IPV6SUPPORT
              value: "false"
            # Set Felix logging to the desired level
            - name: FELIX_LOGSEVERITYSCREEN
              value: "{{- or .Networking.Calico.LogSeverityScreen "info" }}"
            # Set to enable the experimental Prometheus metrics server
            - name: FELIX_PROMETHEUSMETRICSENABLED
              value: "{{- or .Networking.Calico.PrometheusMetricsEnabled "false" }}"
            # TCP port that the Prometheus metrics server should bind to
            - name: FELIX_PROMETHEUSMETRICSPORT
              value: "{{- or .Networking.Calico.PrometheusMetricsPort "9091" }}"
            # Enable Prometheus Go runtime metrics collection
            - name: FELIX_PROMETHEUSGOMETRICSENABLED
              value: "{{- or .Networking.Calico.PrometheusGoMetricsEnabled "true" }}"
            # Enable Prometheus process metrics collection
            - name: FELIX_PROMETHEUSPROCESSMETRICSENABLED
              value: "{{- or .Networking.Calico.PrometheusProcessMetricsEnabled "true" }}"
            # Auto-detect the BGP IP address.
            - name: IP
              value: "autodetect"
            - name: FELIX_HEALTHENABLED
              value: "true"
          securityContext:
            privileged: true
          resources:
            requests:
              cpu: 10m
          livenessProbe:
            httpGet:
              path: /liveness
              port: 9099
              host: localhost
            periodSeconds: 10
            initialDelaySeconds: 10
            failureThreshold: 6
          readinessProbe:
            exec:
              command:
              - /bin/calico-node
              - -bird-ready
              - -felix-ready
            periodSeconds: 10
          volumeMounts:
            - mountPath: /lib/modules
              name: lib-modules
              readOnly: true
            - mountPath: /var/run/calico
              name: var-run-calico
              readOnly: false
            - mountPath: /var/lib/calico
              name: var-lib-calico
              readOnly: false
            # Necessary for gossip based DNS
            - mountPath: /etc/hosts
              name: etc-hosts
              readOnly: true
            {{- if eq $etcd_scheme "https" }}
            - mountPath: /certs
              name: calico
              readOnly: true
            {{- end }}
        # This container installs the Calico CNI binaries
        # and CNI network config file on each node.
        - name: install-cni
          image: calico/cni:v3.8.0
          command: ["/install-cni.sh"]
          env:
            # Name of the CNI config file to create.
            - name: CNI_CONF_NAME
              value: "10-calico.conflist"
            # The location of the Calico etcd cluster.
            - name: ETCD_ENDPOINTS
              valueFrom:
                configMapKeyRef:
                  name: calico-config
                  key: etcd_endpoints
            # The CNI network config to install on each node.
            - name: CNI_NETWORK_CONFIG
              valueFrom:
                configMapKeyRef:
                  name: calico-config
                  key: cni_network_config
          volumeMounts:
            - mountPath: /host/opt/cni/bin
              name: cni-bin-dir
            - mountPath: /host/etc/cni/net.d
              name: cni-net-dir
            # Necessary for gossip based DNS
            - mountPath: /etc/hosts
              name: etc-hosts
              readOnly: true
          resources:
            requests:
              cpu: 10m
      initContainers:
        - name: migrate
          image: calico/upgrade:v1.0.5
          command: ['/bin/sh', '-c', '/node-init-container.sh']
          env:
            - name: CALICO_ETCD_ENDPOINTS
              valueFrom:
                configMapKeyRef:
                  name: calico-config
                  key: etcd_endpoints
            - name: CALICO_APIV1_DATASTORE_TYPE
              value: "etcdv2"
            - name: CALICO_APIV1_ETCD_ENDPOINTS
              valueFrom:
                configMapKeyRef:
                  name: calico-config
                  key: etcd_endpoints
          {{- if eq $etcd_scheme "https" }}
            - name: CALICO_ETCD_CERT_FILE
              value: /certs/calico-client.pem
            - name: CALICO_ETCD_KEY_FILE
              value: /certs/calico-client-key.pem
            - name: CALICO_ETCD_CA_CERT_FILE
              value: /certs/ca.pem
            - name: CALICO_APIV1_ETCD_CERT_FILE
              value: /certs/calico-client.pem
            - name: CALICO_APIV1_ETCD_KEY_FILE
              value: /certs/calico-client-key.pem
            - name: CALICO_APIV1_ETCD_CA_CERT_FILE
              value: /certs/ca.pem
          {{- end }}
          volumeMounts:
            # Necessary for gossip based DNS
            - mountPath: /etc/hosts
              name: etc-hosts
              readOnly: true
          {{- if eq $etcd_scheme "https" }}
            - mountPath: /certs
              name: calico
              readOnly: true
          {{- end }}
      volumes:
        # Used by calico/node.
        - name: lib-modules
          hostPath:
            path: /lib/modules
        - name: var-run-calico
          hostPath:
            path: /var/run/calico
        - name: var-lib-calico
          hostPath:
            path: /var/lib/calico
        # Used to install CNI.
        - name: cni-bin-dir
          hostPath:
            path: /opt/cni/bin
        - name: cni-net-dir
          hostPath:
            path: /etc/cni/net.d
        # Necessary for gossip based DNS
        - name: etc-hosts
          hostPath:
            path: /etc/hosts
        {{- if eq $etcd_scheme "https" }}
        - name: calico
          hostPath:
            path: /srv/kubernetes/calico
        {{- end }}

---

# This manifest deploys the Calico Kubernetes controllers.
# See https://github.com/projectcalico/kube-controllers
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: calico-kube-controllers
  namespace: kube-system
  labels:
    k8s-app: calico-kube-controllers
    role.kubernetes.io/networking: "1"
  annotations:
    scheduler.alpha.kubernetes.io/critical-pod: ''
spec:
  # The controllers can only have a single active instance.
  replicas: 1
  strategy:
    type: Recreate
  template:
    metadata:
      name: calico-kube-controllers
      namespace: kube-system
      labels:
        k8s-app: calico-kube-controllers
        role.kubernetes.io/networking: "1"
    spec:
      # The controllers must run in the host network namespace so that
      # it isn't governed by policy that would prevent it from working.
      hostNetwork: true
      tolerations:
        # Mark the pod as a critical add-on for rescheduling.
        - key: CriticalAddonsOnly
          operator: Exists
        - key: node-role.kubernetes.io/master
          effect: NoSchedule
      serviceAccountName: calico-kube-controllers
      containers:
        - name: calico-kube-controllers
          image: calico/kube-controllers:v3.8.0
          resources:
            requests:
              cpu: 10m
          env:
            # The location of the Calico etcd cluster.
            - name: ETCD_ENDPOINTS
              valueFrom:
                configMapKeyRef:
                  name: calico-config
                  key: etcd_endpoints
            # Choose which controllers to run.
            - name: ENABLED_CONTROLLERS
              value: policy,profile,workloadendpoint,node
            {{- if eq $etcd_scheme "https" }}
            - name: ETCD_CERT_FILE
              value: /certs/calico-client.pem
            - name: ETCD_KEY_FILE
              value: /certs/calico-client-key.pem
            - name: ETCD_CA_CERT_FILE
              value: /certs/ca.pem
          volumeMounts:
            - mountPath: /certs
              name: calico
              readOnly: true
            {{- end }}
          readinessProbe:
            exec:
              command:
              - /usr/bin/check-status
              - -r
      initContainers:
        - name: migrate
          image: calico/upgrade:v1.0.5
          command: ['/bin/sh', '-c', '/controller-init.sh']
          env:
            - name: CALICO_ETCD_ENDPOINTS
              valueFrom:
                configMapKeyRef:
                  name: calico-config
                  key: etcd_endpoints
            - name: CALICO_APIV1_DATASTORE_TYPE
              value: "etcdv2"
            - name: CALICO_APIV1_ETCD_ENDPOINTS
              valueFrom:
                configMapKeyRef:
                  name: calico-config
                  key: etcd_endpoints
          {{- if eq $etcd_scheme "https" }}
            - name: CALICO_ETCD_CERT_FILE
              value: /certs/calico-client.pem
            - name: CALICO_ETCD_KEY_FILE
              value: /certs/calico-client-key.pem
            - name: CALICO_ETCD_CA_CERT_FILE
              value: /certs/ca.pem
            - name: CALICO_APIV1_ETCD_CERT_FILE
              value: /certs/calico-client.pem
            - name: CALICO_APIV1_ETCD_KEY_FILE
              value: /certs/calico-client-key.pem
            - name: CALICO_APIV1_ETCD_CA_CERT_FILE
              value: /certs/ca.pem
          {{- end }}
          volumeMounts:
            # Necessary for gossip based DNS
            - mountPath: /etc/hosts
              name: etc-hosts
              readOnly: true
          {{- if eq $etcd_scheme "https" }}
            - mountPath: /certs
              name: calico
              readOnly: true
          {{- end }}
      volumes:
        # Necessary for gossip based DNS
        - name: etc-hosts
          hostPath:
            path: /etc/hosts
        {{- if eq $etcd_scheme "https" }}
        - name: calico
          hostPath:
            path: /srv/kubernetes/calico
        {{- end }}

# This manifest runs the Migration complete container that monitors for the
# completion of the calico-node Daemonset rollout and when it finishes
# successfully rolling out it will mark the migration complete and allow pods
# to be created again.
---

apiVersion: v1
kind: ServiceAccount
metadata:
  name: calico-upgrade-job
  namespace: kube-system
  labels:
    role.kubernetes.io/networking: "1"
---

kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: calico-upgrade-job
  labels:
    role.kubernetes.io/networking: "1"
rules:
  - apiGroups:
      - extensions
    resources:
      - daemonsets
      - daemonsets/status
    verbs:
      - get
      - list
      - watch
---

apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: calico-upgrade-job
  labels:
    role.kubernetes.io/networking: "1"
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: calico-upgrade-job
subjects:
- kind: ServiceAccount
  name: calico-upgrade-job
  namespace: kube-system
---
# If anything in this job is changed then the name of the job
# should be changed because Jobs cannot be updated, so changing
# the name would run a different Job if the previous version had been
# created before and it does not hurt to rerun this job.

apiVersion: batch/v1
kind: Job
metadata:
  name: calico-complete-upgrade-v331
  namespace: kube-system
  labels:
    role.kubernetes.io/networking: "1"
spec:
  template:
    metadata:
      labels:
        role.kubernetes.io/networking: "1"
    spec:
      hostNetwork: true
      serviceAccountName: calico-upgrade-job
      restartPolicy: OnFailure
      containers:
        - name: migrate-completion
          image: calico/upgrade:v1.0.5
          command: ['/bin/sh', '-c', '/completion-job.sh']
          env:
            - name: EXPECTED_NODE_IMAGE
              value: quay.io/calico/node:v3.7.4
            # The location of the Calico etcd cluster.
            - name: CALICO_ETCD_ENDPOINTS
              valueFrom:
                configMapKeyRef:
                  name: calico-config
                  key: etcd_endpoints
            - name: CALICO_APIV1_DATASTORE_TYPE
              value: "etcdv2"
            - name: CALICO_APIV1_ETCD_ENDPOINTS
              valueFrom:
                configMapKeyRef:
                  name: calico-config
                  key: etcd_endpoints
          {{- if eq $etcd_scheme "https" }}
            - name: CALICO_ETCD_CERT_FILE
              value: /certs/calico-client.pem
            - name: CALICO_ETCD_KEY_FILE
              value: /certs/calico-client-key.pem
            - name: CALICO_ETCD_CA_CERT_FILE
              value: /certs/ca.pem
            - name: CALICO_APIV1_ETCD_CERT_FILE
              value: /certs/calico-client.pem
            - name: CALICO_APIV1_ETCD_KEY_FILE
              value: /certs/calico-client-key.pem
            - name: CALICO_APIV1_ETCD_CA_CERT_FILE
              value: /certs/ca.pem
          {{- end }}
          volumeMounts:
            # Necessary for gossip based DNS
            - mountPath: /etc/hosts
              name: etc-hosts
              readOnly: true
          {{- if eq $etcd_scheme "https" }}
            - mountPath: /certs
              name: calico
              readOnly: true
          {{- end }}
      volumes:
        - name: etc-hosts
          hostPath:
            path: /etc/hosts
        {{- if eq $etcd_scheme "https" }}
        - name: calico
          hostPath:
            path: /srv/kubernetes/calico
        {{- end }}

{{ if and (eq .CloudProvider "aws") (.Networking.Calico.CrossSubnet) -}}
# This manifest installs the k8s-ec2-srcdst container, which disables
# src/dst ip checks to allow BGP to function for calico for hosts within subnets
# This only applies for AWS environments.
---

kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: k8s-ec2-srcdst
  labels:
    role.kubernetes.io/networking: "1"
rules:
- apiGroups:
  - ""
  resources:
  - nodes
  verbs:
  - get
  - list
  - watch
  - update
  - patch

---

apiVersion: v1
kind: ServiceAccount
metadata:
  name: k8s-ec2-srcdst
  namespace: kube-system
  labels:
    role.kubernetes.io/networking: "1"
---

kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: k8s-ec2-srcdst
  labels:
    role.kubernetes.io/networking: "1"
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: k8s-ec2-srcdst
subjects:
- kind: ServiceAccount
  name: k8s-ec2-srcdst
  namespace: kube-system

---

apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: k8s-ec2-srcdst
  namespace: kube-system
  labels:
    k8s-app: k8s-ec2-srcdst
    role.kubernetes.io/networking: "1"
spec:
  replicas: 1
  selector:
    matchLabels:
      k8s-app: k8s-ec2-srcdst
  template:
    metadata:
      labels:
        k8s-app: k8s-ec2-srcdst
        role.kubernetes.io/networking: "1"
      annotations:
        scheduler.alpha.kubernetes.io/critical-pod: ''
    spec:
      hostNetwork: true
      tolerations:
      - key: node-role.kubernetes.io/master
        effect: NoSchedule
      - key: CriticalAddonsOnly
        operator: Exists
      serviceAccountName: k8s-ec2-srcdst
      containers:
        - image: ottoyiu/k8s-ec2-srcdst:v0.2.1
          name: k8s-ec2-srcdst
          resources:
            requests:
              cpu: 10m
              memory: 64Mi
          env:
            - name: AWS_REGION
              value: {{ Region }}
          volumeMounts:
            - name: ssl-certs
              mountPath: "/etc/ssl/certs/ca-certificates.crt"
              readOnly: true
          imagePullPolicy: "Always"
      volumes:
        - name: ssl-certs
          hostPath:
            path: "/etc/ssl/certs/ca-certificates.crt"
      nodeSelector:
        node-role.kubernetes.io/master: ""
{{- end -}}
`)

func cloudupResourcesAddonsNetworkingProjectcalicoOrgK8s17V3YamlTemplateBytes() ([]byte, error) {
	return _cloudupResourcesAddonsNetworkingProjectcalicoOrgK8s17V3YamlTemplate, nil
}

func cloudupResourcesAddonsNetworkingProjectcalicoOrgK8s17V3YamlTemplate() (*asset, error) {
	bytes, err := cloudupResourcesAddonsNetworkingProjectcalicoOrgK8s17V3YamlTemplateBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "cloudup/resources/addons/networking.projectcalico.org/k8s-1.7-v3.yaml.template", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _cloudupResourcesAddonsNetworkingProjectcalicoOrgK8s17YamlTemplate = []byte(`{{- $etcd_scheme := EtcdScheme }}
# This ConfigMap is used to configure a self-hosted Calico installation.
kind: ConfigMap
apiVersion: v1
metadata:
  name: calico-config
  namespace: kube-system
data:
  # The calico-etcd PetSet service IP:port
  etcd_endpoints: "{{ $cluster := index .EtcdClusters 0 -}}
                      {{- range $j, $member := $cluster.Members -}}
                          {{- if $j }},{{ end -}}
                          {{ $etcd_scheme }}://etcd-{{ $member.Name }}.internal.{{ ClusterName }}:4001
                      {{- end }}"

  # Configure the Calico backend to use.
  calico_backend: "bird"

  # The CNI network configuration to install on each node.
  cni_network_config: |-
    {
      "name": "k8s-pod-network",
      "cniVersion": "0.3.0",
      "plugins": [
        {
          "type": "calico",
          "etcd_endpoints": "__ETCD_ENDPOINTS__",
          {{- if eq $etcd_scheme "https" }}
          "etcd_ca_cert_file": "/srv/kubernetes/calico/ca.pem",
          "etcd_cert_file": "/srv/kubernetes/calico/calico-client.pem",
          "etcd_key_file": "/srv/kubernetes/calico/calico-client-key.pem",
          "etcd_scheme": "https",
          {{- end }}
          "log_level": "info",
          {{- if .Networking.Calico.MTU }}
          "mtu": {{- or .Networking.Calico.MTU }},
          {{- end }}
          "ipam": {
            "type": "calico-ipam"
          },
          "policy": {
            "type": "k8s",
            "k8s_api_root": "https://__KUBERNETES_SERVICE_HOST__:__KUBERNETES_SERVICE_PORT__",
            "k8s_auth_token": "__SERVICEACCOUNT_TOKEN__"
          },
          "kubernetes": {
            "kubeconfig": "/etc/cni/net.d/__KUBECONFIG_FILENAME__"
          }
        },
        {
          "type": "portmap",
          "snat": true,
          "capabilities": {"portMappings": true}
        }
      ]
    }

---

kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: calico
  labels:
    role.kubernetes.io/networking: "1"
rules:
- apiGroups:
  - ""
  resources:
  - pods
  - namespaces
  - nodes
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - extensions
  resources:
  - networkpolicies
  verbs:
  - get
  - list
  - watch

---

apiVersion: v1
kind: ServiceAccount
metadata:
  name: calico
  namespace: kube-system
  labels:
    role.kubernetes.io/networking: "1"
---

kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: calico
  labels:
    role.kubernetes.io/networking: "1"
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: calico
subjects:
- kind: ServiceAccount
  name: calico
  namespace: kube-system

---

# This manifest installs the calico/node container, as well
# as the Calico CNI plugins and network config on
# each master and worker node in a Kubernetes cluster.
kind: DaemonSet
apiVersion: extensions/v1beta1
metadata:
  name: calico-node
  namespace: kube-system
  labels:
    k8s-app: calico-node
    role.kubernetes.io/networking: "1"
spec:
  selector:
    matchLabels:
      k8s-app: calico-node
  updateStrategy:
    rollingUpdate:
      maxUnavailable: 1
    type: RollingUpdate
  template:
    metadata:
      labels:
        k8s-app: calico-node
        role.kubernetes.io/networking: "1"
      annotations:
        scheduler.alpha.kubernetes.io/critical-pod: ''
    spec:
      hostNetwork: true
      serviceAccountName: calico
      tolerations:
      - key: CriticalAddonsOnly
        operator: Exists
      - effect: NoExecute
        operator: Exists
      - effect: NoSchedule
        operator: Exists
      # Minimize downtime during a rolling upgrade or deletion; tell Kubernetes to do a "force
      # deletion": https://kubernetes.io/docs/concepts/workloads/pods/pod/#termination-of-pods.
      terminationGracePeriodSeconds: 0
      containers:
        # Runs calico/node container on each Kubernetes node.  This
        # container programs network policy and routes on each
        # host.
        - name: calico-node
          image: quay.io/calico/node:v2.6.12
          resources:
            requests:
              cpu: 10m
          env:
            # The location of the Calico etcd cluster.
            - name: ETCD_ENDPOINTS
              valueFrom:
                configMapKeyRef:
                  name: calico-config
                  key: etcd_endpoints
            {{- if eq $etcd_scheme "https" }}
            - name: ETCD_CERT_FILE
              value: /certs/calico-client.pem
            - name: ETCD_KEY_FILE
              value: /certs/calico-client-key.pem
            - name: ETCD_CA_CERT_FILE
              value: /certs/ca.pem
            {{- end }}
            # Enable BGP.  Disable to enforce policy only.
            - name: CALICO_NETWORKING_BACKEND
              valueFrom:
                configMapKeyRef:
                  name: calico-config
                  key: calico_backend
            # Configure the IP Pool from which Pod IPs will be chosen.
            - name: CALICO_IPV4POOL_CIDR
              value: "{{ .KubeControllerManager.ClusterCIDR }}"
            - name: CALICO_IPV4POOL_IPIP
              value: "{{- if and (eq .CloudProvider "aws") (.Networking.Calico.CrossSubnet) -}}cross-subnet{{- else -}}always{{- end -}}"
            # Cluster type to identify the deployment type
            - name: CLUSTER_TYPE
              value: "kops,bgp"
            # Disable file logging so ` + "`" + `kubectl logs` + "`" + ` works.
            - name: CALICO_DISABLE_FILE_LOGGING
              value: "true"
            # Set noderef for node controller.
            - name: CALICO_K8S_NODE_REF
              valueFrom:
                fieldRef:
                  fieldPath: spec.nodeName
            # Auto-detect the BGP IP address.
            - name: IP
              value: ""
            # Disable IPv6 on Kubernetes.
            - name: FELIX_IPV6SUPPORT
              value: "false"
            # Set Felix logging to the desired level
            - name: FELIX_LOGSEVERITYSCREEN
              value: "{{- or .Networking.Calico.LogSeverityScreen "info" }}"
            # Set to enable the experimental Prometheus metrics server
            - name: FELIX_PROMETHEUSMETRICSENABLED
              value: "{{- or .Networking.Calico.PrometheusMetricsEnabled "false" }}"
            # TCP port that the Prometheus metrics server should bind to
            - name: FELIX_PROMETHEUSMETRICSPORT
              value: "{{- or .Networking.Calico.PrometheusMetricsPort "9091" }}"
            # Enable Prometheus Go runtime metrics collection
            - name: FELIX_PROMETHEUSGOMETRICSENABLED
              value: "{{- or .Networking.Calico.PrometheusGoMetricsEnabled "true" }}"
            # Enable Prometheus process metrics collection
            - name: FELIX_PROMETHEUSPROCESSMETRICSENABLED
              value: "{{- or .Networking.Calico.PrometheusProcessMetricsEnabled "true" }}"
            - name: FELIX_HEALTHENABLED
              value: "true"
            {{- if .Networking.Calico.MTU }}
            - name: FELIX_IPINIPMTU
              value: "{{- or .Networking.Calico.MTU }}"
            {{- end}}
          securityContext:
            privileged: true
          volumeMounts:
            - mountPath: /lib/modules
              name: lib-modules
              readOnly: true
            - mountPath: /var/run/calico
              name: var-run-calico
              readOnly: false
            # Necessary for gossip based DNS
            - mountPath: /etc/hosts
              name: etc-hosts
              readOnly: true
            {{- if eq $etcd_scheme "https" }}
            - mountPath: /certs
              name: calico
              readOnly: true
            {{- end }}
        # This container installs the Calico CNI binaries
        # and CNI network config file on each node.
        - name: install-cni
          image: quay.io/calico/cni:v1.11.8
          resources:
            requests:
              cpu: 10m
          imagePullPolicy: Always
          command: ["/install-cni.sh"]
          env:
            # The name of calico config file
            - name: CNI_CONF_NAME
              value: 10-calico.conflist
            # The location of the Calico etcd cluster.
            - name: ETCD_ENDPOINTS
              valueFrom:
                configMapKeyRef:
                  name: calico-config
                  key: etcd_endpoints
            # The CNI network config to install on each node.
            - name: CNI_NETWORK_CONFIG
              valueFrom:
                configMapKeyRef:
                  name: calico-config
                  key: cni_network_config
          volumeMounts:
            - mountPath: /host/opt/cni/bin
              name: cni-bin-dir
            - mountPath: /host/etc/cni/net.d
              name: cni-net-dir
            # Necessary for gossip based DNS
            - mountPath: /etc/hosts
              name: etc-hosts
              readOnly: true
      volumes:
        # Used by calico/node.
        - name: lib-modules
          hostPath:
            path: /lib/modules
        - name: var-run-calico
          hostPath:
            path: /var/run/calico
        # Used to install CNI.
        - name: cni-bin-dir
          hostPath:
            path: /opt/cni/bin
        - name: cni-net-dir
          hostPath:
            path: /etc/cni/net.d
        - name: etc-hosts
          hostPath:
            path: /etc/hosts
        {{- if eq $etcd_scheme "https" }}
        - name: calico
          hostPath:
            path: /srv/kubernetes/calico
        {{- end }}

---

# This manifest deploys the Calico Kubernetes controllers.
# See https://github.com/projectcalico/kube-controllers
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: calico-kube-controllers
  namespace: kube-system
  labels:
    k8s-app: calico-kube-controllers
    role.kubernetes.io/networking: "1"
spec:
  # The controllers can only have a single active instance.
  replicas: 1
  strategy:
    type: Recreate
  template:
    metadata:
      name: calico-kube-controllers
      namespace: kube-system
      labels:
        k8s-app: calico-kube-controllers
        role.kubernetes.io/networking: "1"
      annotations:
        scheduler.alpha.kubernetes.io/critical-pod: ''
    spec:
      # The controllers must run in the host network namespace so that
      # it isn't governed by policy that would prevent it from working.
      hostNetwork: true
      serviceAccountName: calico
      tolerations:
      - key: node-role.kubernetes.io/master
        effect: NoSchedule
      - key: CriticalAddonsOnly
        operator: Exists
      containers:
        - name: calico-kube-controllers
          image: quay.io/calico/kube-controllers:v1.0.5
          resources:
            requests:
              cpu: 10m
          env:
            # By default only policy, profile, workloadendpoint are turned
            # on, node controller will decommission nodes that do not exist anymore
            # this and CALICO_K8S_NODE_REF in calico-node fixes #3224, but invalid nodes that are
            # already registered in calico needs to be deleted manually, see
            # https://docs.projectcalico.org/v2.6/usage/decommissioning-a-node
            - name: ENABLED_CONTROLLERS
              value: policy,profile,workloadendpoint,node
            # The location of the Calico etcd cluster.
            - name: ETCD_ENDPOINTS
              valueFrom:
                configMapKeyRef:
                  name: calico-config
                  key: etcd_endpoints
            {{- if eq $etcd_scheme "https" }}
            - name: ETCD_CERT_FILE
              value: /certs/calico-client.pem
            - name: ETCD_KEY_FILE
              value: /certs/calico-client-key.pem
            - name: ETCD_CA_CERT_FILE
              value: /certs/ca.pem
            {{- end }}
          volumeMounts:
            # Necessary for gossip based DNS
            - mountPath: /etc/hosts
              name: etc-hosts
              readOnly: true
            {{- if eq $etcd_scheme "https" }}
            - mountPath: /certs
              name: calico
              readOnly: true
            {{- end }}
      volumes:
        - name: etc-hosts
          hostPath:
            path: /etc/hosts
        {{- if eq $etcd_scheme "https" }}
        - name: calico
          hostPath:
            path: /srv/kubernetes/calico
        {{- end }}
---

# This deployment turns off the old "policy-controller". It should remain at 0 replicas, and then
# be removed entirely once the new kube-controllers deployment has been deployed above.
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: calico-policy-controller
  namespace: kube-system
  labels:
    k8s-app: calico-policy
spec:
  # Turn this deployment off in favor of the kube-controllers deployment above.
  replicas: 0
  strategy:
    type: Recreate
  template:
    metadata:
      name: calico-policy-controller
      namespace: kube-system
      labels:
        k8s-app: calico-policy
    spec:
      hostNetwork: true
      serviceAccountName: calico
      containers:
        - name: calico-policy-controller
          # This shouldn't get updated, since this is the last version we shipped that should be used.
          image: quay.io/calico/kube-policy-controller:v0.7.0
          env:
            # The location of the Calico etcd cluster.
            - name: ETCD_ENDPOINTS
              valueFrom:
                configMapKeyRef:
                  name: calico-config
                  key: etcd_endpoints
            {{- if eq $etcd_scheme "https" }}
            - name: ETCD_CERT_FILE
              value: /certs/calico-client.pem
            - name: ETCD_KEY_FILE
              value: /certs/calico-client-key.pem
            - name: ETCD_CA_CERT_FILE
              value: /certs/ca.pem
            {{- end }}
          volumeMounts:
            # Necessary for gossip based DNS
            - mountPath: /etc/hosts
              name: etc-hosts
              readOnly: true
            {{- if eq $etcd_scheme "https" }}
            - mountPath: /certs
              name: calico
              readOnly: true
            {{ end }}
      volumes:
        - name: etc-hosts
          hostPath:
            path: /etc/hosts
        {{- if eq $etcd_scheme "https" }}
        - name: calico
          hostPath:
            path: /srv/kubernetes/calico
        {{- end }}

{{ if and (eq .CloudProvider "aws") (.Networking.Calico.CrossSubnet) -}}
# This manifest installs the k8s-ec2-srcdst container, which disables
# src/dst ip checks to allow BGP to function for calico for hosts within subnets
# This only applies for AWS environments.
---

kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: k8s-ec2-srcdst
  labels:
    role.kubernetes.io/networking: "1"
rules:
- apiGroups:
  - ""
  resources:
  - nodes
  verbs:
  - get
  - list
  - watch
  - update
  - patch

---

apiVersion: v1
kind: ServiceAccount
metadata:
  name: k8s-ec2-srcdst
  namespace: kube-system
  labels:
    role.kubernetes.io/networking: "1"
---

kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: k8s-ec2-srcdst
  labels:
    role.kubernetes.io/networking: "1"
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: k8s-ec2-srcdst
subjects:
- kind: ServiceAccount
  name: k8s-ec2-srcdst
  namespace: kube-system

---

apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: k8s-ec2-srcdst
  namespace: kube-system
  labels:
    k8s-app: k8s-ec2-srcdst
    role.kubernetes.io/networking: "1"
spec:
  replicas: 1
  selector:
    matchLabels:
      k8s-app: k8s-ec2-srcdst
  template:
    metadata:
      labels:
        k8s-app: k8s-ec2-srcdst
        role.kubernetes.io/networking: "1"
      annotations:
        scheduler.alpha.kubernetes.io/critical-pod: ''
    spec:
      hostNetwork: true
      tolerations:
      - key: node-role.kubernetes.io/master
        effect: NoSchedule
      - key: CriticalAddonsOnly
        operator: Exists
      serviceAccountName: k8s-ec2-srcdst
      containers:
        - image: ottoyiu/k8s-ec2-srcdst:v0.2.2
          name: k8s-ec2-srcdst
          resources:
            requests:
              cpu: 10m
              memory: 64Mi
          env:
            - name: AWS_REGION
              value: {{ Region }}
          volumeMounts:
            - name: ssl-certs
              mountPath: "/etc/ssl/certs/ca-certificates.crt"
              readOnly: true
          imagePullPolicy: "Always"
      volumes:
        - name: ssl-certs
          hostPath:
            path: "/etc/ssl/certs/ca-certificates.crt"
      nodeSelector:
        node-role.kubernetes.io/master: ""
{{- end -}}
`)

func cloudupResourcesAddonsNetworkingProjectcalicoOrgK8s17YamlTemplateBytes() ([]byte, error) {
	return _cloudupResourcesAddonsNetworkingProjectcalicoOrgK8s17YamlTemplate, nil
}

func cloudupResourcesAddonsNetworkingProjectcalicoOrgK8s17YamlTemplate() (*asset, error) {
	bytes, err := cloudupResourcesAddonsNetworkingProjectcalicoOrgK8s17YamlTemplateBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "cloudup/resources/addons/networking.projectcalico.org/k8s-1.7.yaml.template", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _cloudupResourcesAddonsNetworkingProjectcalicoOrgCanalK8s112YamlTemplate = []byte(`# Pulled and modified from: https://docs.projectcalico.org/v3.7/manifests/canal.yaml

---
# Source: calico/templates/calico-config.yaml
# This ConfigMap is used to configure a self-hosted Canal installation.
kind: ConfigMap
apiVersion: v1
metadata:
  name: canal-config
  namespace: kube-system
data:
  # Typha is disabled.
  typha_service_name: "none"
  # The interface used by canal for host <-> host communication.
  # If left blank, then the interface is chosen using the node's
  # default route.
  canal_iface: ""

  # Whether or not to masquerade traffic to destinations not within
  # the pod network.
  masquerade: "true"

  # MTU default is 1500, can be overridden
  veth_mtu: "{{- or .Networking.Canal.MTU "1500" }}"

  # The CNI network configuration to install on each node.  The special
  # values in this config will be automatically populated.
  cni_network_config: |-
    {
      "name": "k8s-pod-network",
      "cniVersion": "0.3.0",
      "plugins": [
        {
          "type": "calico",
          "log_level": "info",
          "datastore_type": "kubernetes",
          "mtu": __CNI_MTU__,
          "nodename": "__KUBERNETES_NODE_NAME__",
          "ipam": {
              "type": "host-local",
              "subnet": "usePodCidr"
          },
          "policy": {
              "type": "k8s"
          },
          "kubernetes": {
              "kubeconfig": "__KUBECONFIG_FILEPATH__"
          }
        },
        {
          "type": "portmap",
          "snat": true,
          "capabilities": {"portMappings": true}
        }
      ]
    }

  # Flannel network configuration. Mounted into the flannel container.
  net-conf.json: |
    {
      "Network": "{{ .NonMasqueradeCIDR }}",
      "Backend": {
        "Type": "vxlan"
      }
    }

---

# Source: calico/templates/kdd-crds.yaml
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
   name: felixconfigurations.crd.projectcalico.org
spec:
  scope: Cluster
  group: crd.projectcalico.org
  version: v1
  names:
    kind: FelixConfiguration
    plural: felixconfigurations
    singular: felixconfiguration
---

apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: bgpconfigurations.crd.projectcalico.org
spec:
  scope: Cluster
  group: crd.projectcalico.org
  version: v1
  names:
    kind: BGPConfiguration
    plural: bgpconfigurations
    singular: bgpconfiguration

---

apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: ippools.crd.projectcalico.org
spec:
  scope: Cluster
  group: crd.projectcalico.org
  version: v1
  names:
    kind: IPPool
    plural: ippools
    singular: ippool

---

apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: hostendpoints.crd.projectcalico.org
spec:
  scope: Cluster
  group: crd.projectcalico.org
  version: v1
  names:
    kind: HostEndpoint
    plural: hostendpoints
    singular: hostendpoint

---

apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: clusterinformations.crd.projectcalico.org
spec:
  scope: Cluster
  group: crd.projectcalico.org
  version: v1
  names:
    kind: ClusterInformation
    plural: clusterinformations
    singular: clusterinformation

---

apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: globalnetworkpolicies.crd.projectcalico.org
spec:
  scope: Cluster
  group: crd.projectcalico.org
  version: v1
  names:
    kind: GlobalNetworkPolicy
    plural: globalnetworkpolicies
    singular: globalnetworkpolicy

---

apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: globalnetworksets.crd.projectcalico.org
spec:
  scope: Cluster
  group: crd.projectcalico.org
  version: v1
  names:
    kind: GlobalNetworkSet
    plural: globalnetworksets
    singular: globalnetworkset

---

apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: networkpolicies.crd.projectcalico.org
spec:
  scope: Namespaced
  group: crd.projectcalico.org
  version: v1
  names:
    kind: NetworkPolicy
    plural: networkpolicies
    singular: networkpolicy

---

apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: networksets.crd.projectcalico.org
spec:
  scope: Namespaced
  group: crd.projectcalico.org
  version: v1
  names:
    kind: NetworkSet
    plural: networksets
    singular: networkset

---

# Include a clusterrole for the calico-node DaemonSet,
# and bind it to the canal serviceaccount.
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: calico
rules:
  # The CNI plugin needs to get pods, nodes, and namespaces.
  - apiGroups: [""]
    resources:
      - pods
      - nodes
      - namespaces
    verbs:
      - get
  - apiGroups: [""]
    resources:
      - endpoints
      - services
    verbs:
      # Used to discover service IPs for advertisement.
      - watch
      - list
      # Used to discover Typhas.
      - get
  - apiGroups: [""]
    resources:
      - nodes/status
    verbs:
      # Needed for clearing NodeNetworkUnavailable flag.
      - patch
      # Calico stores some configuration information in node annotations.
      - update
  # Watch for changes to Kubernetes NetworkPolicies.
  - apiGroups: ["networking.k8s.io"]
    resources:
      - networkpolicies
    verbs:
      - watch
      - list
  # Used by Calico for policy information.
  - apiGroups: [""]
    resources:
      - pods
      - namespaces
      - serviceaccounts
    verbs:
      - list
      - watch
  # The CNI plugin patches pods/status.
  - apiGroups: [""]
    resources:
      - pods/status
    verbs:
      - patch
  # Calico monitors various CRDs for config.
  - apiGroups: ["crd.projectcalico.org"]
    resources:
      - globalfelixconfigs
      - felixconfigurations
      - bgppeers
      - globalbgpconfigs
      - bgpconfigurations
      - ippools
      - ipamblocks
      - globalnetworkpolicies
      - globalnetworksets
      - networkpolicies
      - networksets
      - clusterinformations
      - hostendpoints
    verbs:
      - get
      - list
      - watch
  # Calico must create and update some CRDs on startup.
  - apiGroups: ["crd.projectcalico.org"]
    resources:
      - ippools
      - felixconfigurations
      - clusterinformations
    verbs:
      - create
      - update
  # Calico stores some configuration information on the node.
  - apiGroups: [""]
    resources:
      - nodes
    verbs:
      - get
      - list
      - watch
  # These permissions are only required for upgrade from v2.6, and can
  # be removed after upgrade or on fresh installations.
  - apiGroups: ["crd.projectcalico.org"]
    resources:
      - bgpconfigurations
      - bgppeers
    verbs:
      - create
      - update
---
# Flannel ClusterRole
# Pulled from https://github.com/coreos/flannel/blob/master/Documentation/k8s-manifests/kube-flannel-rbac.yml
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: flannel
rules:
  - apiGroups: [""]
    resources:
      - pods
    verbs:
      - get
  - apiGroups: [""]
    resources:
      - nodes
    verbs:
      - list
      - watch
  - apiGroups: [""]
    resources:
      - nodes/status
    verbs:
      - patch
---
# Bind the flannel ClusterRole to the canal ServiceAccount.
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: canal-flannel
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: flannel
subjects:
- kind: ServiceAccount
  name: canal
  namespace: kube-system
---
# Bind the Calico ClusterRole to the canal ServiceAccount.
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: canal-calico
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: calico
subjects:
- kind: ServiceAccount
  name: canal
  namespace: kube-system

---

# This manifest installs the calico/node container, as well
# as the Calico CNI plugins and network config on
# each master and worker node in a Kubernetes cluster.
kind: DaemonSet
apiVersion: apps/v1
metadata:
  name: canal
  namespace: kube-system
  labels:
    k8s-app: canal
spec:
  selector:
    matchLabels:
      k8s-app: canal
  updateStrategy:
    type: RollingUpdate
    rollingUpdate:
      maxUnavailable: 1
  template:
    metadata:
      labels:
        k8s-app: canal
      annotations:
        # This, along with the CriticalAddonsOnly toleration below,
        # marks the pod as a critical add-on, ensuring it gets
        # priority scheduling and that its resources are reserved
        # if it ever gets evicted.
        scheduler.alpha.kubernetes.io/critical-pod: ''
    spec:
      priorityClassName: system-node-critical
      nodeSelector:
        beta.kubernetes.io/os: linux
      hostNetwork: true
      tolerations:
        # Make sure canal gets scheduled on all nodes.
        - effect: NoSchedule
          operator: Exists
        # Mark the pod as a critical add-on for rescheduling.
        - key: CriticalAddonsOnly
          operator: Exists
        - effect: NoExecute
          operator: Exists
      serviceAccountName: canal
      # Minimize downtime during a rolling upgrade or deletion; tell Kubernetes to do a "force
      # deletion": https://kubernetes.io/docs/concepts/workloads/pods/pod/#termination-of-pods.
      terminationGracePeriodSeconds: 0
      initContainers:
        # This container installs the Calico CNI binaries
        # and CNI network config file on each node.
        - name: install-cni
          image: calico/cni:v3.7.5
          command: ["/install-cni.sh"]
          env:
            # Name of the CNI config file to create.
            - name: CNI_CONF_NAME
              value: "10-canal.conflist"
            # CNI MTU Config variable
            - name: CNI_MTU
              valueFrom:
                configMapKeyRef:
                  name: canal-config
                  key: veth_mtu
            # The CNI network config to install on each node.
            - name: CNI_NETWORK_CONFIG
              valueFrom:
                configMapKeyRef:
                  name: canal-config
                  key: cni_network_config
            # Set the hostname based on the k8s node name.
            - name: KUBERNETES_NODE_NAME
              valueFrom:
                fieldRef:
                  fieldPath: spec.nodeName
            # Prevents the container from sleeping forever.
            - name: SLEEP
              value: "false"
          volumeMounts:
            - mountPath: /host/opt/cni/bin
              name: cni-bin-dir
            - mountPath: /host/etc/cni/net.d
              name: cni-net-dir
      containers:
        # Runs calico/node container on each Kubernetes node.  This
        # container programs network policy and routes on each
        # host.
        - name: calico-node
          image: calico/node:v3.7.5
          env:
            # Use Kubernetes API as the backing datastore.
            - name: DATASTORE_TYPE
              value: "kubernetes"
            # Configure route aggregation based on pod CIDR.
            - name: USE_POD_CIDR
              value: "true"
            # Wait for the datastore.
            - name: WAIT_FOR_DATASTORE
              value: "true"
            # Set based on the k8s node name.
            - name: NODENAME
              valueFrom:
                fieldRef:
                  fieldPath: spec.nodeName
            # Don't enable BGP.
            - name: CALICO_NETWORKING_BACKEND
              value: "none"
            # Cluster type to identify the deployment type
            - name: CLUSTER_TYPE
              value: "k8s,canal"
            # Period, in seconds, at which felix re-applies all iptables state
            - name: FELIX_IPTABLESREFRESHINTERVAL
              value: "60"
            # No IP address needed.
            - name: IP
              value: ""
            # Disable file logging so ` + "`" + `kubectl logs` + "`" + ` works.
            - name: CALICO_DISABLE_FILE_LOGGING
              value: "true"
            - name: FELIX_IPINIPMTU
              valueFrom:
                configMapKeyRef:
                  name: canal-config
                  key: veth_mtu
            # Disable IPv6 on Kubernetes.
            - name: FELIX_IPV6SUPPORT
              value: "false"
            # Set Felix logging to "INFO"
            - name: FELIX_LOGSEVERITYSCREEN
              value: "{{- or .Networking.Canal.LogSeveritySys "INFO" }}"
            # Set Felix endpoint to host default action to ACCEPT.
            - name: FELIX_DEFAULTENDPOINTTOHOSTACTION
              value: "{{- or .Networking.Canal.DefaultEndpointToHostAction "ACCEPT" }}"
            # Controls whether Felix inserts rules to the top of iptables chains, or appends to the bottom
            - name: FELIX_CHAININSERTMODE
              value: "{{- or .Networking.Canal.ChainInsertMode "insert" }}"
            # Set to enable the experimental Prometheus metrics server
            - name: FELIX_PROMETHEUSMETRICSENABLED
              value: "{{- or .Networking.Canal.PrometheusMetricsEnabled "false" }}"
            # TCP port that the Prometheus metrics server should bind to
            - name: FELIX_PROMETHEUSMETRICSPORT
              value: "{{- or .Networking.Canal.PrometheusMetricsPort "9091" }}"
            # Enable Prometheus Go runtime metrics collection
            - name: FELIX_PROMETHEUSGOMETRICSENABLED
              value: "{{- or .Networking.Canal.PrometheusGoMetricsEnabled "true" }}"
            # Enable Prometheus process metrics collection
            - name: FELIX_PROMETHEUSPROCESSMETRICSENABLED
              value: "{{- or .Networking.Canal.PrometheusProcessMetricsEnabled "true" }}"
            - name: FELIX_HEALTHENABLED
              value: "true"
          securityContext:
            privileged: true
          resources:
            requests:
              cpu: 100m
          livenessProbe:
            httpGet:
              path: /liveness
              port: 9099
              host: localhost
            periodSeconds: 10
            initialDelaySeconds: 10
            failureThreshold: 6
          readinessProbe:
            httpGet:
              path: /readiness
              port: 9099
              host: localhost
            periodSeconds: 10
          volumeMounts:
            - mountPath: /lib/modules
              name: lib-modules
              readOnly: true
            - mountPath: /run/xtables.lock
              name: xtables-lock
              readOnly: false
            - mountPath: /var/run/calico
              name: var-run-calico
              readOnly: false
            - mountPath: /var/lib/calico
              name: var-lib-calico
              readOnly: false
        # This container runs flannel using the kube-subnet-mgr backend
        # for allocating subnets.
        - name: kube-flannel
          image: quay.io/coreos/flannel:v0.11.0
          command: [ "/opt/bin/flanneld", "--ip-masq", "--kube-subnet-mgr" ]
          securityContext:
            privileged: true
          env:
            - name: POD_NAME
              valueFrom:
                fieldRef:
                  fieldPath: metadata.name
            - name: POD_NAMESPACE
              valueFrom:
                fieldRef:
                  fieldPath: metadata.namespace
            - name: FLANNELD_IFACE
              valueFrom:
                configMapKeyRef:
                  name: canal-config
                  key: canal_iface
            - name: FLANNELD_IP_MASQ
              valueFrom:
                configMapKeyRef:
                  name: canal-config
                  key: masquerade
            {{- if eq .Networking.Canal.DisableFlannelForwardRules true }}
            - name: FLANNELD_IPTABLES_FORWARD_RULES
              value: "false"
            {{- end }}
          volumeMounts:
          - mountPath: /run/xtables.lock
            name: xtables-lock
            readOnly: false
          - name: flannel-cfg
            mountPath: /etc/kube-flannel/
      volumes:
        # Used by calico/node.
        - name: lib-modules
          hostPath:
            path: /lib/modules
        - name: var-run-calico
          hostPath:
            path: /var/run/calico
        - name: var-lib-calico
          hostPath:
            path: /var/lib/calico
        - name: xtables-lock
          hostPath:
            path: /run/xtables.lock
            type: FileOrCreate
        # Used by flannel.
        - name: flannel-cfg
          configMap:
            name: canal-config
        # Used to install CNI.
        - name: cni-bin-dir
          hostPath:
            path: /opt/cni/bin
        - name: cni-net-dir
          hostPath:
            path: /etc/cni/net.d
---

apiVersion: v1
kind: ServiceAccount
metadata:
  name: canal
  namespace: kube-system
`)

func cloudupResourcesAddonsNetworkingProjectcalicoOrgCanalK8s112YamlTemplateBytes() ([]byte, error) {
	return _cloudupResourcesAddonsNetworkingProjectcalicoOrgCanalK8s112YamlTemplate, nil
}

func cloudupResourcesAddonsNetworkingProjectcalicoOrgCanalK8s112YamlTemplate() (*asset, error) {
	bytes, err := cloudupResourcesAddonsNetworkingProjectcalicoOrgCanalK8s112YamlTemplateBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "cloudup/resources/addons/networking.projectcalico.org.canal/k8s-1.12.yaml.template", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _cloudupResourcesAddonsNetworkingProjectcalicoOrgCanalK8s115YamlTemplate = []byte(`# Pulled and modified from: https://docs.projectcalico.org/v3.12/manifests/canal.yaml

---
# Source: calico/templates/calico-config.yaml
# This ConfigMap is used to configure a self-hosted Canal installation.
kind: ConfigMap
apiVersion: v1
metadata:
  name: canal-config
  namespace: kube-system
  labels:
    role.kubernetes.io/networking: "1"
data:
  # Typha is disabled.
  typha_service_name: "{{ if .Networking.Canal.TyphaReplicas }}calico-typha{{ else }}none{{ end }}"
  # The interface used by canal for host <-> host communication.
  # If left blank, then the interface is chosen using the node's
  # default route.
  canal_iface: ""

  # Whether or not to masquerade traffic to destinations not within
  # the pod network.
  masquerade: "true"

  # Configure the MTU to use
  {{- if .Networking.Canal.MTU }}
  veth_mtu: "{{ .Networking.Canal.MTU }}"
  {{- else }}
  veth_mtu: "{{- if eq .CloudProvider "openstack" -}}1430{{- else -}}1440{{- end -}}"
  {{- end }}

  # The CNI network configuration to install on each node.  The special
  # values in this config will be automatically populated.
  cni_network_config: |-
    {
      "name": "k8s-pod-network",
      "cniVersion": "0.3.1",
      "plugins": [
        {
          "type": "calico",
          "log_level": "info",
          "datastore_type": "kubernetes",
          "nodename": "__KUBERNETES_NODE_NAME__",
          "mtu": __CNI_MTU__,
          "ipam": {
              "type": "host-local",
              "subnet": "usePodCidr"
          },
          "policy": {
              "type": "k8s"
          },
          "kubernetes": {
              "kubeconfig": "__KUBECONFIG_FILEPATH__"
          }
        },
        {
          "type": "portmap",
          "snat": true,
          "capabilities": {"portMappings": true}
        },
        {
          "type": "bandwidth",
          "capabilities": {"bandwidth": true}
        }
      ]
    }

  # Flannel network configuration. Mounted into the flannel container.
  net-conf.json: |
    {
      "Network": "{{ .NonMasqueradeCIDR }}",
      "Backend": {
        "Type": "vxlan"
      }
    }

---
# Source: calico/templates/kdd-crds.yaml
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: felixconfigurations.crd.projectcalico.org
  labels:
    role.kubernetes.io/networking: "1"
spec:
  scope: Cluster
  group: crd.projectcalico.org
  version: v1
  names:
    kind: FelixConfiguration
    plural: felixconfigurations
    singular: felixconfiguration
---

apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: ipamblocks.crd.projectcalico.org
  labels:
    role.kubernetes.io/networking: "1"
spec:
  scope: Cluster
  group: crd.projectcalico.org
  version: v1
  names:
    kind: IPAMBlock
    plural: ipamblocks
    singular: ipamblock

---

apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: blockaffinities.crd.projectcalico.org
  labels:
    role.kubernetes.io/networking: "1"
spec:
  scope: Cluster
  group: crd.projectcalico.org
  version: v1
  names:
    kind: BlockAffinity
    plural: blockaffinities
    singular: blockaffinity

---

apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: ipamhandles.crd.projectcalico.org
  labels:
    role.kubernetes.io/networking: "1"
spec:
  scope: Cluster
  group: crd.projectcalico.org
  version: v1
  names:
    kind: IPAMHandle
    plural: ipamhandles
    singular: ipamhandle

---

apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: ipamconfigs.crd.projectcalico.org
  labels:
    role.kubernetes.io/networking: "1"
spec:
  scope: Cluster
  group: crd.projectcalico.org
  version: v1
  names:
    kind: IPAMConfig
    plural: ipamconfigs
    singular: ipamconfig

---

apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: bgppeers.crd.projectcalico.org
  labels:
    role.kubernetes.io/networking: "1"
spec:
  scope: Cluster
  group: crd.projectcalico.org
  version: v1
  names:
    kind: BGPPeer
    plural: bgppeers
    singular: bgppeer

---

apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: bgpconfigurations.crd.projectcalico.org
  labels:
    role.kubernetes.io/networking: "1"
spec:
  scope: Cluster
  group: crd.projectcalico.org
  version: v1
  names:
    kind: BGPConfiguration
    plural: bgpconfigurations
    singular: bgpconfiguration

---

apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: ippools.crd.projectcalico.org
  labels:
    role.kubernetes.io/networking: "1"
spec:
  scope: Cluster
  group: crd.projectcalico.org
  version: v1
  names:
    kind: IPPool
    plural: ippools
    singular: ippool

---

apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: hostendpoints.crd.projectcalico.org
  labels:
    role.kubernetes.io/networking: "1"
spec:
  scope: Cluster
  group: crd.projectcalico.org
  version: v1
  names:
    kind: HostEndpoint
    plural: hostendpoints
    singular: hostendpoint

---

apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: clusterinformations.crd.projectcalico.org
  labels:
    role.kubernetes.io/networking: "1"
spec:
  scope: Cluster
  group: crd.projectcalico.org
  version: v1
  names:
    kind: ClusterInformation
    plural: clusterinformations
    singular: clusterinformation

---

apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: globalnetworkpolicies.crd.projectcalico.org
  labels:
    role.kubernetes.io/networking: "1"
spec:
  scope: Cluster
  group: crd.projectcalico.org
  version: v1
  names:
    kind: GlobalNetworkPolicy
    plural: globalnetworkpolicies
    singular: globalnetworkpolicy

---

apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: globalnetworksets.crd.projectcalico.org
  labels:
    role.kubernetes.io/networking: "1"
spec:
  scope: Cluster
  group: crd.projectcalico.org
  version: v1
  names:
    kind: GlobalNetworkSet
    plural: globalnetworksets
    singular: globalnetworkset

---

apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: networkpolicies.crd.projectcalico.org
  labels:
    role.kubernetes.io/networking: "1"
spec:
  scope: Namespaced
  group: crd.projectcalico.org
  version: v1
  names:
    kind: NetworkPolicy
    plural: networkpolicies
    singular: networkpolicy

---

apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: networksets.crd.projectcalico.org
  labels:
    role.kubernetes.io/networking: "1"
spec:
  scope: Namespaced
  group: crd.projectcalico.org
  version: v1
  names:
    kind: NetworkSet
    plural: networksets
    singular: networkset
---
# Source: calico/templates/rbac.yaml

# Include a clusterrole for the calico-node DaemonSet,
# and bind it to the calico-node serviceaccount.
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: calico
  labels:
    role.kubernetes.io/networking: "1"
rules:
  # The CNI plugin needs to get pods, nodes, and namespaces.
  - apiGroups: [""]
    resources:
      - pods
      - nodes
      - namespaces
    verbs:
      - get
  - apiGroups: [""]
    resources:
      - endpoints
      - services
    verbs:
      # Used to discover service IPs for advertisement.
      - watch
      - list
      # Used to discover Typhas.
      - get
  - apiGroups: [""]
    resources:
      - nodes/status
    verbs:
      # Needed for clearing NodeNetworkUnavailable flag.
      - patch
      # Calico stores some configuration information in node annotations.
      - update
  # Watch for changes to Kubernetes NetworkPolicies.
  - apiGroups: ["networking.k8s.io"]
    resources:
      - networkpolicies
    verbs:
      - watch
      - list
  # Used by Calico for policy information.
  - apiGroups: [""]
    resources:
      - pods
      - namespaces
      - serviceaccounts
    verbs:
      - list
      - watch
  # The CNI plugin patches pods/status.
  - apiGroups: [""]
    resources:
      - pods/status
    verbs:
      - patch
  # Calico monitors various CRDs for config.
  - apiGroups: ["crd.projectcalico.org"]
    resources:
      - globalfelixconfigs
      - felixconfigurations
      - bgppeers
      - globalbgpconfigs
      - bgpconfigurations
      - ippools
      - ipamblocks
      - globalnetworkpolicies
      - globalnetworksets
      - networkpolicies
      - networksets
      - clusterinformations
      - hostendpoints
      - blockaffinities
    verbs:
      - get
      - list
      - watch
  # Calico must create and update some CRDs on startup.
  - apiGroups: ["crd.projectcalico.org"]
    resources:
      - ippools
      - felixconfigurations
      - clusterinformations
    verbs:
      - create
      - update
  # Calico stores some configuration information on the node.
  - apiGroups: [""]
    resources:
      - nodes
    verbs:
      - get
      - list
      - watch
  # These permissions are only required for upgrade from v2.6, and can
  # be removed after upgrade or on fresh installations.
  - apiGroups: ["crd.projectcalico.org"]
    resources:
      - bgpconfigurations
      - bgppeers
    verbs:
      - create
      - update
---
# Flannel ClusterRole
# Pulled from https://github.com/coreos/flannel/blob/master/Documentation/kube-flannel-rbac.yml
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: flannel
  labels:
    role.kubernetes.io/networking: "1"
rules:
  - apiGroups: [""]
    resources:
      - pods
    verbs:
      - get
  - apiGroups: [""]
    resources:
      - nodes
    verbs:
      - list
      - watch
  - apiGroups: [""]
    resources:
      - nodes/status
    verbs:
      - patch
---
# Bind the flannel ClusterRole to the canal ServiceAccount.
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: canal-flannel
  labels:
    role.kubernetes.io/networking: "1"
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: flannel
subjects:
- kind: ServiceAccount
  name: canal
  namespace: kube-system
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: canal-calico
  labels:
    role.kubernetes.io/networking: "1"
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: calico
subjects:
- kind: ServiceAccount
  name: canal
  namespace: kube-system

{{ if .Networking.Canal.TyphaReplicas -}}
---
# Source: calico/templates/calico-typha.yaml
# This manifest creates a Service, which will be backed by Calico's Typha daemon.
# Typha sits in between Felix and the API server, reducing Calico's load on the API server.

apiVersion: v1
kind: Service
metadata:
  name: calico-typha
  namespace: kube-system
  labels:
    k8s-app: calico-typha
    role.kubernetes.io/networking: "1"
spec:
  ports:
    - port: 5473
      protocol: TCP
      targetPort: calico-typha
      name: calico-typha
  selector:
    k8s-app: calico-typha

---

# This manifest creates a Deployment of Typha to back the above service.

apiVersion: apps/v1
kind: Deployment
metadata:
  name: calico-typha
  namespace: kube-system
  labels:
    k8s-app: calico-typha
    role.kubernetes.io/networking: "1"
spec:
  # Number of Typha replicas.  To enable Typha, set this to a non-zero value *and* set the
  # typha_service_name variable in the canal-config ConfigMap above.
  #
  # We recommend using Typha if you have more than 50 nodes.  Above 100 nodes it is essential
  # (when using the Kubernetes datastore).  Use one replica for every 100-200 nodes.  In
  # production, we recommend running at least 3 replicas to reduce the impact of rolling upgrade.
  replicas: {{ or .Networking.Canal.TyphaReplicas 0 }}
  revisionHistoryLimit: 2
  selector:
    matchLabels:
      k8s-app: calico-typha
  template:
    metadata:
      labels:
        k8s-app: calico-typha
        role.kubernetes.io/networking: "1"
      annotations:
        # This, along with the CriticalAddonsOnly toleration below, marks the pod as a critical
        # add-on, ensuring it gets priority scheduling and that its resources are reserved
        # if it ever gets evicted.
        scheduler.alpha.kubernetes.io/critical-pod: ''
        cluster-autoscaler.kubernetes.io/safe-to-evict: 'true'
    spec:
      nodeSelector:
        kubernetes.io/os: linux
        kubernetes.io/role: master
      hostNetwork: true
      tolerations:
        # Mark the pod as a critical add-on for rescheduling.
        - key: CriticalAddonsOnly
          operator: Exists
        - key: "node-role.kubernetes.io/master"
          effect: NoSchedule
      # Since Calico can't network a pod until Typha is up, we need to run Typha itself
      # as a host-networked pod.
      serviceAccountName: canal
      priorityClassName: system-cluster-critical
      # fsGroup allows using projected serviceaccount tokens as described here kubernetes/kubernetes#82573
      securityContext:
        fsGroup: 65534
      containers:
      - image: calico/typha:v3.12.0
        name: calico-typha
        ports:
        - containerPort: 5473
          name: calico-typha
          protocol: TCP
        env:
          # Enable "info" logging by default.  Can be set to "debug" to increase verbosity.
          - name: TYPHA_LOGSEVERITYSCREEN
            value: "info"
          # Disable logging to file and syslog since those don't make sense in Kubernetes.
          - name: TYPHA_LOGFILEPATH
            value: "none"
          - name: TYPHA_LOGSEVERITYSYS
            value: "none"
          # Monitor the Kubernetes API to find the number of running instances and rebalance
          # connections.
          - name: TYPHA_CONNECTIONREBALANCINGMODE
            value: "kubernetes"
          - name: TYPHA_DATASTORETYPE
            value: "kubernetes"
          - name: TYPHA_HEALTHENABLED
            value: "true"
          - name: TYPHA_PROMETHEUSMETRICSENABLED
            value: "{{- or .Networking.Canal.TyphaPrometheusMetricsEnabled "false" }}"
          - name: TYPHA_PROMETHEUSMETRICSPORT
            value: "{{- or .Networking.Canal.TyphaPrometheusMetricsPort "9093" }}"
        livenessProbe:
          httpGet:
            path: /liveness
            port: 9098
            host: localhost
          periodSeconds: 30
          initialDelaySeconds: 30
        securityContext:
          runAsNonRoot: true
          allowPrivilegeEscalation: false
        readinessProbe:
          httpGet:
            path: /readiness
            port: 9098
            host: localhost
          periodSeconds: 10

---

# This manifest creates a Pod Disruption Budget for Typha to allow K8s Cluster Autoscaler to evict

apiVersion: policy/v1beta1
kind: PodDisruptionBudget
metadata:
  name: calico-typha
  namespace: kube-system
  labels:
    k8s-app: calico-typha
    role.kubernetes.io/networking: "1"
spec:
  maxUnavailable: 1
  selector:
    matchLabels:
      k8s-app: calico-typha
{{- end }}

---
# Source: calico/templates/calico-node.yaml
# This manifest installs the canal container, as well
# as the CNI plugins and network config on
# each master and worker node in a Kubernetes cluster.
kind: DaemonSet
apiVersion: apps/v1
metadata:
  name: canal
  namespace: kube-system
  labels:
    k8s-app: canal
    role.kubernetes.io/networking: "1"
spec:
  selector:
    matchLabels:
      k8s-app: canal
  updateStrategy:
    type: RollingUpdate
    rollingUpdate:
      maxUnavailable: 1
  template:
    metadata:
      labels:
        k8s-app: canal
        role.kubernetes.io/networking: "1"
      annotations:
        # This, along with the CriticalAddonsOnly toleration below,
        # marks the pod as a critical add-on, ensuring it gets
        # priority scheduling and that its resources are reserved
        # if it ever gets evicted.
        scheduler.alpha.kubernetes.io/critical-pod: ''
    spec:
      nodeSelector:
        kubernetes.io/os: linux
      hostNetwork: true
      tolerations:
        # Make sure canal gets scheduled on all nodes.
        - effect: NoSchedule
          operator: Exists
        # Mark the pod as a critical add-on for rescheduling.
        - key: CriticalAddonsOnly
          operator: Exists
        - effect: NoExecute
          operator: Exists
      serviceAccountName: canal
      # Minimize downtime during a rolling upgrade or deletion; tell Kubernetes to do a "force
      # deletion": https://kubernetes.io/docs/concepts/workloads/pods/pod/#termination-of-pods.
      terminationGracePeriodSeconds: 0
      priorityClassName: system-node-critical
      initContainers:
        # This container installs the CNI binaries
        # and CNI network config file on each node.
        - name: install-cni
          image: calico/cni:v3.12.0
          command: ["/install-cni.sh"]
          env:
            # Name of the CNI config file to create.
            - name: CNI_CONF_NAME
              value: "10-canal.conflist"
            # The CNI network config to install on each node.
            - name: CNI_NETWORK_CONFIG
              valueFrom:
                configMapKeyRef:
                  name: canal-config
                  key: cni_network_config
            # Set the hostname based on the k8s node name.
            - name: KUBERNETES_NODE_NAME
              valueFrom:
                fieldRef:
                  fieldPath: spec.nodeName
            # CNI MTU Config variable
            - name: CNI_MTU
              valueFrom:
                configMapKeyRef:
                  name: canal-config
                  key: veth_mtu
            # Prevents the container from sleeping forever.
            - name: SLEEP
              value: "false"
          volumeMounts:
            - mountPath: /host/opt/cni/bin
              name: cni-bin-dir
            - mountPath: /host/etc/cni/net.d
              name: cni-net-dir
          securityContext:
            privileged: true
        # Adds a Flex Volume Driver that creates a per-pod Unix Domain Socket to allow Dikastes
        # to communicate with Felix over the Policy Sync API.
        - name: flexvol-driver
          image: calico/pod2daemon-flexvol:v3.12.0
          volumeMounts:
          - name: flexvol-driver-host
            mountPath: /host/driver
          securityContext:
            privileged: true
      containers:
        # Runs canal container on each Kubernetes node.  This
        # container programs network policy and routes on each
        # host.
        - name: calico-node
          image: calico/node:v3.12.0
          env:
            # Use Kubernetes API as the backing datastore.
            - name: DATASTORE_TYPE
              value: "kubernetes"
            # Configure route aggregation based on pod CIDR.
            - name: USE_POD_CIDR
              value: "true"
            {{- if .Networking.Canal.TyphaReplicas }}
            # Typha support: controlled by the ConfigMap.
            - name: FELIX_TYPHAK8SSERVICENAME
              valueFrom:
                configMapKeyRef:
                  name: canal-config
                  key: typha_service_name
            {{- end }}
            # Wait for the datastore.
            - name: WAIT_FOR_DATASTORE
              value: "true"
            # Set based on the k8s node name.
            - name: NODENAME
              valueFrom:
                fieldRef:
                  fieldPath: spec.nodeName
            # Don't enable BGP.
            - name: CALICO_NETWORKING_BACKEND
              value: "none"
            # Cluster type to identify the deployment type
            - name: CLUSTER_TYPE
              # was value: "k8s,bgp"
              value: "k8s,canal"
            # Period, in seconds, at which felix re-applies all iptables state
            - name: FELIX_IPTABLESREFRESHINTERVAL
              value: "60"
            # No IP address needed.
            - name: IP
              value: ""
            # Set MTU for tunnel device used if ipip is enabled
            - name: FELIX_IPINIPMTU
              valueFrom:
                configMapKeyRef:
                  name: canal-config
                  key: veth_mtu
            # Disable file logging so ` + "`" + `kubectl logs` + "`" + ` works.
            - name: CALICO_DISABLE_FILE_LOGGING
              value: "true"
            # Set Felix endpoint to host default action to ACCEPT.
            - name: FELIX_DEFAULTENDPOINTTOHOSTACTION
              value: "{{- or .Networking.Canal.DefaultEndpointToHostAction "ACCEPT" }}"
            # Disable IPv6 on Kubernetes.
            - name: FELIX_IPV6SUPPORT
              value: "false"
            # Set Felix logging to "INFO"
            - name: FELIX_LOGSEVERITYSCREEN
              value: "{{- or .Networking.Canal.LogSeveritySys "INFO" }}"
            - name: FELIX_HEALTHENABLED
              value: "true"

            # kops additions
            # Controls whether Felix inserts rules to the top of iptables chains, or appends to the bottom
            - name: FELIX_CHAININSERTMODE
              value: "{{- or .Networking.Canal.ChainInsertMode "insert" }}"
            # Set Felix iptables binary variant, Legacy or NFT
            - name: FELIX_IPTABLESBACKEND
              value: "{{- or .Networking.Canal.IptablesBackend "Auto" }}"
            # Set to enable the experimental Prometheus metrics server
            - name: FELIX_PROMETHEUSMETRICSENABLED
              value: "{{- or .Networking.Canal.PrometheusMetricsEnabled "false" }}"
            # TCP port that the Prometheus metrics server should bind to
            - name: FELIX_PROMETHEUSMETRICSPORT
              value: "{{- or .Networking.Canal.PrometheusMetricsPort "9091" }}"
            # Enable Prometheus Go runtime metrics collection
            - name: FELIX_PROMETHEUSGOMETRICSENABLED
              value: "{{- or .Networking.Canal.PrometheusGoMetricsEnabled "true" }}"
            # Enable Prometheus process metrics collection
            - name: FELIX_PROMETHEUSPROCESSMETRICSENABLED
              value: "{{- or .Networking.Canal.PrometheusProcessMetricsEnabled "true" }}"
          securityContext:
            privileged: true
          resources:
            requests:
              cpu: 250m
          livenessProbe:
            exec:
              command:
              - /bin/calico-node
              - -felix-live
            periodSeconds: 10
            initialDelaySeconds: 10
            failureThreshold: 6
          readinessProbe:
            httpGet:
              path: /readiness
              port: 9099
              host: localhost
            periodSeconds: 10
          volumeMounts:
            - mountPath: /lib/modules
              name: lib-modules
              readOnly: true
            - mountPath: /run/xtables.lock
              name: xtables-lock
              readOnly: false
            - mountPath: /var/run/calico
              name: var-run-calico
              readOnly: false
            - mountPath: /var/lib/calico
              name: var-lib-calico
              readOnly: false
            - name: policysync
              mountPath: /var/run/nodeagent
        # This container runs flannel using the kube-subnet-mgr backend
        # for allocating subnets.
        - name: kube-flannel
          image: quay.io/coreos/flannel:v0.11.0
          command: [ "/opt/bin/flanneld", "--ip-masq", "--kube-subnet-mgr" ]
          securityContext:
            privileged: true
          env:
            - name: POD_NAME
              valueFrom:
                fieldRef:
                  fieldPath: metadata.name
            - name: POD_NAMESPACE
              valueFrom:
                fieldRef:
                  fieldPath: metadata.namespace
            - name: FLANNELD_IFACE
              valueFrom:
                configMapKeyRef:
                  name: canal-config
                  key: canal_iface
            - name: FLANNELD_IP_MASQ
              valueFrom:
                configMapKeyRef:
                  name: canal-config
                  key: masquerade
            {{- if eq .Networking.Canal.DisableFlannelForwardRules true }}
            - name: FLANNELD_IPTABLES_FORWARD_RULES
              value: "false"
            {{- end }}
          volumeMounts:
          - mountPath: /run/xtables.lock
            name: xtables-lock
            readOnly: false
          - name: flannel-cfg
            mountPath: /etc/kube-flannel/
      volumes:
        # Used by canal.
        - name: lib-modules
          hostPath:
            path: /lib/modules
        - name: var-run-calico
          hostPath:
            path: /var/run/calico
        - name: var-lib-calico
          hostPath:
            path: /var/lib/calico
        - name: xtables-lock
          hostPath:
            path: /run/xtables.lock
            type: FileOrCreate
        # Used by flannel.
        - name: flannel-cfg
          configMap:
            name: canal-config
        # Used to install CNI.
        - name: cni-bin-dir
          hostPath:
            path: /opt/cni/bin
        - name: cni-net-dir
          hostPath:
            path: /etc/cni/net.d
        # Used to create per-pod Unix Domain Sockets
        - name: policysync
          hostPath:
            type: DirectoryOrCreate
            path: /var/run/nodeagent
        # Used to install Flex Volume Driver
        - name: flexvol-driver-host
          hostPath:
            type: DirectoryOrCreate
            path: "{{- or .Kubelet.VolumePluginDirectory "/usr/libexec/kubernetes/kubelet-plugins/volume/exec/" }}nodeagent~uds"
---

apiVersion: v1
kind: ServiceAccount
metadata:
  name: canal
  namespace: kube-system
  labels:
    role.kubernetes.io/networking: "1"
`)

func cloudupResourcesAddonsNetworkingProjectcalicoOrgCanalK8s115YamlTemplateBytes() ([]byte, error) {
	return _cloudupResourcesAddonsNetworkingProjectcalicoOrgCanalK8s115YamlTemplate, nil
}

func cloudupResourcesAddonsNetworkingProjectcalicoOrgCanalK8s115YamlTemplate() (*asset, error) {
	bytes, err := cloudupResourcesAddonsNetworkingProjectcalicoOrgCanalK8s115YamlTemplateBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "cloudup/resources/addons/networking.projectcalico.org.canal/k8s-1.15.yaml.template", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _cloudupResourcesAddonsNetworkingProjectcalicoOrgCanalK8s116YamlTemplate = []byte(`# Pulled and modified from: https://docs.projectcalico.org/v3.13/manifests/canal.yaml

---
# Source: calico/templates/calico-config.yaml
# This ConfigMap is used to configure a self-hosted Canal installation.
kind: ConfigMap
apiVersion: v1
metadata:
  name: canal-config
  namespace: kube-system
  labels:
    role.kubernetes.io/networking: "1"
data:
  # Typha is disabled.
  typha_service_name: "{{ if .Networking.Canal.TyphaReplicas }}calico-typha{{ else }}none{{ end }}"
  # The interface used by canal for host <-> host communication.
  # If left blank, then the interface is chosen using the node's
  # default route.
  canal_iface: ""

  # Whether or not to masquerade traffic to destinations not within
  # the pod network.
  masquerade: "true"

  # Configure the MTU to use
  {{- if .Networking.Canal.MTU }}
  veth_mtu: "{{ .Networking.Canal.MTU }}"
  {{- else }}
  veth_mtu: "{{- if eq .CloudProvider "openstack" -}}1430{{- else -}}1440{{- end -}}"
  {{- end }}

  # The CNI network configuration to install on each node.  The special
  # values in this config will be automatically populated.
  cni_network_config: |-
    {
      "name": "k8s-pod-network",
      "cniVersion": "0.3.1",
      "plugins": [
        {
          "type": "calico",
          "log_level": "info",
          "datastore_type": "kubernetes",
          "nodename": "__KUBERNETES_NODE_NAME__",
          "mtu": __CNI_MTU__,
          "ipam": {
              "type": "host-local",
              "subnet": "usePodCidr"
          },
          "policy": {
              "type": "k8s"
          },
          "kubernetes": {
              "kubeconfig": "__KUBECONFIG_FILEPATH__"
          }
        },
        {
          "type": "portmap",
          "snat": true,
          "capabilities": {"portMappings": true}
        },
        {
          "type": "bandwidth",
          "capabilities": {"bandwidth": true}
        }
      ]
    }

  # Flannel network configuration. Mounted into the flannel container.
  net-conf.json: |
    {
      "Network": "{{ .NonMasqueradeCIDR }}",
      "Backend": {
        "Type": "vxlan"
      }
    }

---
# Source: calico/templates/kdd-crds.yaml

apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: bgpconfigurations.crd.projectcalico.org
  labels:
    role.kubernetes.io/networking: "1"
spec:
  scope: Cluster
  group: crd.projectcalico.org
  version: v1
  names:
    kind: BGPConfiguration
    plural: bgpconfigurations
    singular: bgpconfiguration

---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: bgppeers.crd.projectcalico.org
  labels:
    role.kubernetes.io/networking: "1"
spec:
  scope: Cluster
  group: crd.projectcalico.org
  version: v1
  names:
    kind: BGPPeer
    plural: bgppeers
    singular: bgppeer

---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: blockaffinities.crd.projectcalico.org
  labels:
    role.kubernetes.io/networking: "1"
spec:
  scope: Cluster
  group: crd.projectcalico.org
  version: v1
  names:
    kind: BlockAffinity
    plural: blockaffinities
    singular: blockaffinity

---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: clusterinformations.crd.projectcalico.org
  labels:
    role.kubernetes.io/networking: "1"
spec:
  scope: Cluster
  group: crd.projectcalico.org
  version: v1
  names:
    kind: ClusterInformation
    plural: clusterinformations
    singular: clusterinformation

---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: felixconfigurations.crd.projectcalico.org
  labels:
    role.kubernetes.io/networking: "1"
spec:
  scope: Cluster
  group: crd.projectcalico.org
  version: v1
  names:
    kind: FelixConfiguration
    plural: felixconfigurations
    singular: felixconfiguration

---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: globalnetworkpolicies.crd.projectcalico.org
  labels:
    role.kubernetes.io/networking: "1"
spec:
  scope: Cluster
  group: crd.projectcalico.org
  version: v1
  names:
    kind: GlobalNetworkPolicy
    plural: globalnetworkpolicies
    singular: globalnetworkpolicy

---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: globalnetworksets.crd.projectcalico.org
  labels:
    role.kubernetes.io/networking: "1"
spec:
  scope: Cluster
  group: crd.projectcalico.org
  version: v1
  names:
    kind: GlobalNetworkSet
    plural: globalnetworksets
    singular: globalnetworkset

---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: hostendpoints.crd.projectcalico.org
  labels:
    role.kubernetes.io/networking: "1"
spec:
  scope: Cluster
  group: crd.projectcalico.org
  version: v1
  names:
    kind: HostEndpoint
    plural: hostendpoints
    singular: hostendpoint

---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: ipamblocks.crd.projectcalico.org
  labels:
    role.kubernetes.io/networking: "1"
spec:
  scope: Cluster
  group: crd.projectcalico.org
  version: v1
  names:
    kind: IPAMBlock
    plural: ipamblocks
    singular: ipamblock

---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: ipamconfigs.crd.projectcalico.org
  labels:
    role.kubernetes.io/networking: "1"
spec:
  scope: Cluster
  group: crd.projectcalico.org
  version: v1
  names:
    kind: IPAMConfig
    plural: ipamconfigs
    singular: ipamconfig

---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: ipamhandles.crd.projectcalico.org
  labels:
    role.kubernetes.io/networking: "1"
spec:
  scope: Cluster
  group: crd.projectcalico.org
  version: v1
  names:
    kind: IPAMHandle
    plural: ipamhandles
    singular: ipamhandle

---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: ippools.crd.projectcalico.org
  labels:
    role.kubernetes.io/networking: "1"
spec:
  scope: Cluster
  group: crd.projectcalico.org
  version: v1
  names:
    kind: IPPool
    plural: ippools
    singular: ippool

---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: networkpolicies.crd.projectcalico.org
  labels:
    role.kubernetes.io/networking: "1"
spec:
  scope: Namespaced
  group: crd.projectcalico.org
  version: v1
  names:
    kind: NetworkPolicy
    plural: networkpolicies
    singular: networkpolicy

---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: networksets.crd.projectcalico.org
  labels:
    role.kubernetes.io/networking: "1"
spec:
  scope: Namespaced
  group: crd.projectcalico.org
  version: v1
  names:
    kind: NetworkSet
    plural: networksets
    singular: networkset

---
# Source: calico/templates/rbac.yaml

# Include a clusterrole for the calico-node DaemonSet,
# and bind it to the calico-node serviceaccount.
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: calico
  labels:
    role.kubernetes.io/networking: "1"
rules:
  # The CNI plugin needs to get pods, nodes, and namespaces.
  - apiGroups: [""]
    resources:
      - pods
      - nodes
      - namespaces
    verbs:
      - get
  - apiGroups: [""]
    resources:
      - endpoints
      - services
    verbs:
      # Used to discover service IPs for advertisement.
      - watch
      - list
      # Used to discover Typhas.
      - get
  # Pod CIDR auto-detection on kubeadm needs access to config maps.
  - apiGroups: [""]
    resources:
      - configmaps
    verbs:
      - get
  - apiGroups: [""]
    resources:
      - nodes/status
    verbs:
      # Needed for clearing NodeNetworkUnavailable flag.
      - patch
      # Calico stores some configuration information in node annotations.
      - update
  # Watch for changes to Kubernetes NetworkPolicies.
  - apiGroups: ["networking.k8s.io"]
    resources:
      - networkpolicies
    verbs:
      - watch
      - list
  # Used by Calico for policy information.
  - apiGroups: [""]
    resources:
      - pods
      - namespaces
      - serviceaccounts
    verbs:
      - list
      - watch
  # The CNI plugin patches pods/status.
  - apiGroups: [""]
    resources:
      - pods/status
    verbs:
      - patch
  # Calico monitors various CRDs for config.
  - apiGroups: ["crd.projectcalico.org"]
    resources:
      - globalfelixconfigs
      - felixconfigurations
      - bgppeers
      - globalbgpconfigs
      - bgpconfigurations
      - ippools
      - ipamblocks
      - globalnetworkpolicies
      - globalnetworksets
      - networkpolicies
      - networksets
      - clusterinformations
      - hostendpoints
      - blockaffinities
    verbs:
      - get
      - list
      - watch
  # Calico must create and update some CRDs on startup.
  - apiGroups: ["crd.projectcalico.org"]
    resources:
      - ippools
      - felixconfigurations
      - clusterinformations
    verbs:
      - create
      - update
  # Calico stores some configuration information on the node.
  - apiGroups: [""]
    resources:
      - nodes
    verbs:
      - get
      - list
      - watch
  # These permissions are only required for upgrade from v2.6, and can
  # be removed after upgrade or on fresh installations.
  - apiGroups: ["crd.projectcalico.org"]
    resources:
      - bgpconfigurations
      - bgppeers
    verbs:
      - create
      - update

---
# Flannel ClusterRole
# Pulled from https://github.com/coreos/flannel/blob/master/Documentation/kube-flannel-rbac.yml
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: flannel
  labels:
    role.kubernetes.io/networking: "1"
rules:
  - apiGroups: [""]
    resources:
      - pods
    verbs:
      - get
  - apiGroups: [""]
    resources:
      - nodes
    verbs:
      - list
      - watch
  - apiGroups: [""]
    resources:
      - nodes/status
    verbs:
      - patch
---
# Bind the flannel ClusterRole to the canal ServiceAccount.
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: canal-flannel
  labels:
    role.kubernetes.io/networking: "1"
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: flannel
subjects:
- kind: ServiceAccount
  name: canal
  namespace: kube-system
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: canal-calico
  labels:
    role.kubernetes.io/networking: "1"
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: calico
subjects:
- kind: ServiceAccount
  name: canal
  namespace: kube-system

{{ if .Networking.Canal.TyphaReplicas -}}
---
# Source: calico/templates/calico-typha.yaml
# This manifest creates a Service, which will be backed by Calico's Typha daemon.
# Typha sits in between Felix and the API server, reducing Calico's load on the API server.

apiVersion: v1
kind: Service
metadata:
  name: calico-typha
  namespace: kube-system
  labels:
    k8s-app: calico-typha
    role.kubernetes.io/networking: "1"
spec:
  ports:
    - port: 5473
      protocol: TCP
      targetPort: calico-typha
      name: calico-typha
  selector:
    k8s-app: calico-typha

---

# This manifest creates a Deployment of Typha to back the above service.

apiVersion: apps/v1
kind: Deployment
metadata:
  name: calico-typha
  namespace: kube-system
  labels:
    k8s-app: calico-typha
    role.kubernetes.io/networking: "1"
spec:
  # Number of Typha replicas.  To enable Typha, set this to a non-zero value *and* set the
  # typha_service_name variable in the canal-config ConfigMap above.
  #
  # We recommend using Typha if you have more than 50 nodes.  Above 100 nodes it is essential
  # (when using the Kubernetes datastore).  Use one replica for every 100-200 nodes.  In
  # production, we recommend running at least 3 replicas to reduce the impact of rolling upgrade.
  replicas: {{ or .Networking.Canal.TyphaReplicas 0 }}
  revisionHistoryLimit: 2
  selector:
    matchLabels:
      k8s-app: calico-typha
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxUnavailable: 1
  template:
    metadata:
      labels:
        k8s-app: calico-typha
        role.kubernetes.io/networking: "1"
      annotations:
        cluster-autoscaler.kubernetes.io/safe-to-evict: 'true'
    spec:
      nodeSelector:
        kubernetes.io/os: linux
        kubernetes.io/role: master
      hostNetwork: true
      tolerations:
        # Mark the pod as a critical add-on for rescheduling.
        - key: CriticalAddonsOnly
          operator: Exists
        - key: "node-role.kubernetes.io/master"
          effect: NoSchedule
      # Since Calico can't network a pod until Typha is up, we need to run Typha itself
      # as a host-networked pod.
      serviceAccountName: canal
      priorityClassName: system-cluster-critical
      # fsGroup allows using projected serviceaccount tokens as described here kubernetes/kubernetes#82573
      securityContext:
        fsGroup: 65534
      containers:
      - image: calico/typha:v3.13.2
        name: calico-typha
        ports:
        - containerPort: 5473
          name: calico-typha
          protocol: TCP
        env:
          # Enable "info" logging by default.  Can be set to "debug" to increase verbosity.
          - name: TYPHA_LOGSEVERITYSCREEN
            value: "info"
          # Disable logging to file and syslog since those don't make sense in Kubernetes.
          - name: TYPHA_LOGFILEPATH
            value: "none"
          - name: TYPHA_LOGSEVERITYSYS
            value: "none"
          # Monitor the Kubernetes API to find the number of running instances and rebalance
          # connections.
          - name: TYPHA_CONNECTIONREBALANCINGMODE
            value: "kubernetes"
          - name: TYPHA_DATASTORETYPE
            value: "kubernetes"
          - name: TYPHA_HEALTHENABLED
            value: "true"
          - name: TYPHA_PROMETHEUSMETRICSENABLED
            value: "{{- or .Networking.Canal.TyphaPrometheusMetricsEnabled "false" }}"
          - name: TYPHA_PROMETHEUSMETRICSPORT
            value: "{{- or .Networking.Canal.TyphaPrometheusMetricsPort "9093" }}"
        livenessProbe:
          httpGet:
            path: /liveness
            port: 9098
            host: localhost
          periodSeconds: 30
          initialDelaySeconds: 30
        securityContext:
          runAsNonRoot: true
          allowPrivilegeEscalation: false
        readinessProbe:
          httpGet:
            path: /readiness
            port: 9098
            host: localhost
          periodSeconds: 10

---

# This manifest creates a Pod Disruption Budget for Typha to allow K8s Cluster Autoscaler to evict

apiVersion: policy/v1beta1
kind: PodDisruptionBudget
metadata:
  name: calico-typha
  namespace: kube-system
  labels:
    k8s-app: calico-typha
    role.kubernetes.io/networking: "1"
spec:
  maxUnavailable: 1
  selector:
    matchLabels:
      k8s-app: calico-typha
{{- end }}

---
# Source: calico/templates/calico-node.yaml
# This manifest installs the canal container, as well
# as the CNI plugins and network config on
# each master and worker node in a Kubernetes cluster.
kind: DaemonSet
apiVersion: apps/v1
metadata:
  name: canal
  namespace: kube-system
  labels:
    k8s-app: canal
    role.kubernetes.io/networking: "1"
spec:
  selector:
    matchLabels:
      k8s-app: canal
  updateStrategy:
    type: RollingUpdate
    rollingUpdate:
      maxUnavailable: 1
  template:
    metadata:
      labels:
        k8s-app: canal
        role.kubernetes.io/networking: "1"
    spec:
      nodeSelector:
        kubernetes.io/os: linux
      hostNetwork: true
      tolerations:
        # Make sure canal gets scheduled on all nodes.
        - effect: NoSchedule
          operator: Exists
        # Mark the pod as a critical add-on for rescheduling.
        - key: CriticalAddonsOnly
          operator: Exists
        - effect: NoExecute
          operator: Exists
      serviceAccountName: canal
      # Minimize downtime during a rolling upgrade or deletion; tell Kubernetes to do a "force
      # deletion": https://kubernetes.io/docs/concepts/workloads/pods/pod/#termination-of-pods.
      terminationGracePeriodSeconds: 0
      priorityClassName: system-node-critical
      initContainers:
        # This container installs the CNI binaries
        # and CNI network config file on each node.
        - name: install-cni
          image: calico/cni:v3.13.2
          command: ["/install-cni.sh"]
          env:
            # Name of the CNI config file to create.
            - name: CNI_CONF_NAME
              value: "10-canal.conflist"
            # The CNI network config to install on each node.
            - name: CNI_NETWORK_CONFIG
              valueFrom:
                configMapKeyRef:
                  name: canal-config
                  key: cni_network_config
            # Set the hostname based on the k8s node name.
            - name: KUBERNETES_NODE_NAME
              valueFrom:
                fieldRef:
                  fieldPath: spec.nodeName
            # CNI MTU Config variable
            - name: CNI_MTU
              valueFrom:
                configMapKeyRef:
                  name: canal-config
                  key: veth_mtu
            # Prevents the container from sleeping forever.
            - name: SLEEP
              value: "false"
          volumeMounts:
            - mountPath: /host/opt/cni/bin
              name: cni-bin-dir
            - mountPath: /host/etc/cni/net.d
              name: cni-net-dir
          securityContext:
            privileged: true
        # Adds a Flex Volume Driver that creates a per-pod Unix Domain Socket to allow Dikastes
        # to communicate with Felix over the Policy Sync API.
        - name: flexvol-driver
          image: calico/pod2daemon-flexvol:v3.13.2
          volumeMounts:
          - name: flexvol-driver-host
            mountPath: /host/driver
          securityContext:
            privileged: true
      containers:
        # Runs canal container on each Kubernetes node.  This
        # container programs network policy and routes on each
        # host.
        - name: calico-node
          image: calico/node:v3.13.2
          env:
            # Use Kubernetes API as the backing datastore.
            - name: DATASTORE_TYPE
              value: "kubernetes"
            # Configure route aggregation based on pod CIDR.
            - name: USE_POD_CIDR
              value: "true"
            {{- if .Networking.Canal.TyphaReplicas }}
            # Typha support: controlled by the ConfigMap.
            - name: FELIX_TYPHAK8SSERVICENAME
              valueFrom:
                configMapKeyRef:
                  name: canal-config
                  key: typha_service_name
            {{- end }}
            # Wait for the datastore.
            - name: WAIT_FOR_DATASTORE
              value: "true"
            # Set based on the k8s node name.
            - name: NODENAME
              valueFrom:
                fieldRef:
                  fieldPath: spec.nodeName
            # Don't enable BGP.
            - name: CALICO_NETWORKING_BACKEND
              value: "none"
            # Cluster type to identify the deployment type
            - name: CLUSTER_TYPE
              value: "k8s,canal"
            # Period, in seconds, at which felix re-applies all iptables state
            - name: FELIX_IPTABLESREFRESHINTERVAL
              value: "60"
            # No IP address needed.
            - name: IP
              value: ""
            # Set MTU for tunnel device used if ipip is enabled
            - name: FELIX_IPINIPMTU
              valueFrom:
                configMapKeyRef:
                  name: canal-config
                  key: veth_mtu
            # Disable file logging so ` + "`" + `kubectl logs` + "`" + ` works.
            - name: CALICO_DISABLE_FILE_LOGGING
              value: "true"
            # Set Felix endpoint to host default action to ACCEPT.
            - name: FELIX_DEFAULTENDPOINTTOHOSTACTION
              value: "{{- or .Networking.Canal.DefaultEndpointToHostAction "ACCEPT" }}"
            # Disable IPv6 on Kubernetes.
            - name: FELIX_IPV6SUPPORT
              value: "false"
            # Set Felix logging to "info"
            - name: FELIX_LOGSEVERITYSCREEN
              value: "{{- or .Networking.Canal.LogSeveritySys "info" }}"
            - name: FELIX_HEALTHENABLED
              value: "true"

            # kops additions
            # Controls whether Felix inserts rules to the top of iptables chains, or appends to the bottom
            - name: FELIX_CHAININSERTMODE
              value: "{{- or .Networking.Canal.ChainInsertMode "insert" }}"
            # Set Felix iptables binary variant, Legacy or NFT
            - name: FELIX_IPTABLESBACKEND
              value: "{{- or .Networking.Canal.IptablesBackend "Auto" }}"
            # Set to enable the experimental Prometheus metrics server
            - name: FELIX_PROMETHEUSMETRICSENABLED
              value: "{{- or .Networking.Canal.PrometheusMetricsEnabled "false" }}"
            # TCP port that the Prometheus metrics server should bind to
            - name: FELIX_PROMETHEUSMETRICSPORT
              value: "{{- or .Networking.Canal.PrometheusMetricsPort "9091" }}"
            # Enable Prometheus Go runtime metrics collection
            - name: FELIX_PROMETHEUSGOMETRICSENABLED
              value: "{{- or .Networking.Canal.PrometheusGoMetricsEnabled "true" }}"
            # Enable Prometheus process metrics collection
            - name: FELIX_PROMETHEUSPROCESSMETRICSENABLED
              value: "{{- or .Networking.Canal.PrometheusProcessMetricsEnabled "true" }}"
          securityContext:
            privileged: true
          resources:
            requests:
              cpu: 90m
          livenessProbe:
            exec:
              command:
              - /bin/calico-node
              - -felix-live
            periodSeconds: 10
            initialDelaySeconds: 10
            failureThreshold: 6
          readinessProbe:
            httpGet:
              path: /readiness
              port: 9099
              host: localhost
            periodSeconds: 10
          volumeMounts:
            - mountPath: /lib/modules
              name: lib-modules
              readOnly: true
            - mountPath: /run/xtables.lock
              name: xtables-lock
              readOnly: false
            - mountPath: /var/run/calico
              name: var-run-calico
              readOnly: false
            - mountPath: /var/lib/calico
              name: var-lib-calico
              readOnly: false
            - name: policysync
              mountPath: /var/run/nodeagent
        # This container runs flannel using the kube-subnet-mgr backend
        # for allocating subnets.
        - name: kube-flannel
          image: quay.io/coreos/flannel:v0.11.0
          command: [ "/opt/bin/flanneld", "--ip-masq", "--kube-subnet-mgr" ]
          securityContext:
            privileged: true
          env:
            - name: POD_NAME
              valueFrom:
                fieldRef:
                  fieldPath: metadata.name
            - name: POD_NAMESPACE
              valueFrom:
                fieldRef:
                  fieldPath: metadata.namespace
            - name: FLANNELD_IFACE
              valueFrom:
                configMapKeyRef:
                  name: canal-config
                  key: canal_iface
            - name: FLANNELD_IP_MASQ
              valueFrom:
                configMapKeyRef:
                  name: canal-config
                  key: masquerade
            {{- if eq .Networking.Canal.DisableFlannelForwardRules true }}
            - name: FLANNELD_IPTABLES_FORWARD_RULES
              value: "false"
            {{- end }}
          volumeMounts:
          - mountPath: /run/xtables.lock
            name: xtables-lock
            readOnly: false
          - name: flannel-cfg
            mountPath: /etc/kube-flannel/
      volumes:
        # Used by canal.
        - name: lib-modules
          hostPath:
            path: /lib/modules
        - name: var-run-calico
          hostPath:
            path: /var/run/calico
        - name: var-lib-calico
          hostPath:
            path: /var/lib/calico
        - name: xtables-lock
          hostPath:
            path: /run/xtables.lock
            type: FileOrCreate
        # Used by flannel.
        - name: flannel-cfg
          configMap:
            name: canal-config
        # Used to install CNI.
        - name: cni-bin-dir
          hostPath:
            path: /opt/cni/bin
        - name: cni-net-dir
          hostPath:
            path: /etc/cni/net.d
        # Used to create per-pod Unix Domain Sockets
        - name: policysync
          hostPath:
            type: DirectoryOrCreate
            path: /var/run/nodeagent
        # Used to install Flex Volume Driver
        - name: flexvol-driver-host
          hostPath:
            type: DirectoryOrCreate
            path: "{{- or .Kubelet.VolumePluginDirectory "/usr/libexec/kubernetes/kubelet-plugins/volume/exec/" }}nodeagent~uds"
---

apiVersion: v1
kind: ServiceAccount
metadata:
  name: canal
  namespace: kube-system
  labels:
    role.kubernetes.io/networking: "1"
`)

func cloudupResourcesAddonsNetworkingProjectcalicoOrgCanalK8s116YamlTemplateBytes() ([]byte, error) {
	return _cloudupResourcesAddonsNetworkingProjectcalicoOrgCanalK8s116YamlTemplate, nil
}

func cloudupResourcesAddonsNetworkingProjectcalicoOrgCanalK8s116YamlTemplate() (*asset, error) {
	bytes, err := cloudupResourcesAddonsNetworkingProjectcalicoOrgCanalK8s116YamlTemplateBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "cloudup/resources/addons/networking.projectcalico.org.canal/k8s-1.16.yaml.template", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _cloudupResourcesAddonsNetworkingProjectcalicoOrgCanalK8s19YamlTemplate = []byte(`# Canal Version v3.2.3
# https://docs.projectcalico.org/v3.2/releases#v3.2.3
# This manifest includes the following component versions:
#   calico/node:v3.2.3
#   calico/cni:v3.2.3
#   coreos/flannel:v0.9.0

# This ConfigMap is used to configure a self-hosted Canal installation.
kind: ConfigMap
apiVersion: v1
metadata:
  name: canal-config
  namespace: kube-system
data:
  # The interface used by canal for host <-> host communication.
  # If left blank, then the interface is chosen using the node's
  # default route.
  canal_iface: ""

  # Whether or not to masquerade traffic to destinations not within
  # the pod network.
  masquerade: "true"

  # The CNI network configuration to install on each node.  The special
  # values in this config will be automatically populated.
  cni_network_config: |-
    {
      "name": "k8s-pod-network",
      "cniVersion": "0.3.0",
      "plugins": [
        {
          "type": "calico",
          "log_level": "info",
          "datastore_type": "kubernetes",
          "nodename": "__KUBERNETES_NODE_NAME__",
          "ipam": {
            "type": "host-local",
            "subnet": "usePodCidr"
          },
          "policy": {
              "type": "k8s"
          },
          "kubernetes": {
              "kubeconfig": "__KUBECONFIG_FILEPATH__"
          }
        },
        {
          "type": "portmap",
          "snat": true,
          "capabilities": {"portMappings": true}
        }
      ]
    }

  # Flannel network configuration. Mounted into the flannel container.
  net-conf.json: |
    {
      "Network": "{{ .NonMasqueradeCIDR }}",
      "Backend": {
        "Type": "vxlan"
      }
    }

---



# This manifest installs the calico/node container, as well
# as the Calico CNI plugins and network config on
# each master and worker node in a Kubernetes cluster.
kind: DaemonSet
apiVersion: apps/v1
metadata:
  name: canal
  namespace: kube-system
  labels:
    k8s-app: canal
spec:
  selector:
    matchLabels:
      k8s-app: canal
  updateStrategy:
    type: RollingUpdate
    rollingUpdate:
      maxUnavailable: 1
  template:
    metadata:
      labels:
        k8s-app: canal
      annotations:
        # This, along with the CriticalAddonsOnly toleration below,
        # marks the pod as a critical add-on, ensuring it gets
        # priority scheduling and that its resources are reserved
        # if it ever gets evicted.
        scheduler.alpha.kubernetes.io/critical-pod: ''
    spec:
      nodeSelector:
        beta.kubernetes.io/os: linux
      hostNetwork: true
      tolerations:
        # Make sure canal gets scheduled on all nodes.
        - effect: NoSchedule
          operator: Exists
        # Mark the pod as a critical add-on for rescheduling.
        - key: CriticalAddonsOnly
          operator: Exists
        - effect: NoExecute
          operator: Exists
      serviceAccountName: canal
      # Minimize downtime during a rolling upgrade or deletion; tell Kubernetes to do a "force
      # deletion": https://kubernetes.io/docs/concepts/workloads/pods/pod/#termination-of-pods.
      terminationGracePeriodSeconds: 0
      containers:
        # Runs calico/node container on each Kubernetes node.  This
        # container programs network policy and routes on each
        # host.
        - name: calico-node
          image: quay.io/calico/node:v3.2.3
          env:
            # Use Kubernetes API as the backing datastore.
            - name: DATASTORE_TYPE
              value: "kubernetes"
            # Wait for the datastore.
            - name: WAIT_FOR_DATASTORE
              value: "true"
            # Set based on the k8s node name.
            - name: NODENAME
              valueFrom:
                fieldRef:
                  fieldPath: spec.nodeName
            # Don't enable BGP.
            - name: CALICO_NETWORKING_BACKEND
              value: "none"
            # Cluster type to identify the deployment type
            - name: CLUSTER_TYPE
              value: "k8s,canal"
            # Period, in seconds, at which felix re-applies all iptables state
            - name: FELIX_IPTABLESREFRESHINTERVAL
              value: "60"
            # No IP address needed.
            - name: IP
              value: ""
            # Disable file logging so ` + "`" + `kubectl logs` + "`" + ` works.
            - name: CALICO_DISABLE_FILE_LOGGING
              value: "true"
            # Disable IPv6 on Kubernetes.
            - name: FELIX_IPV6SUPPORT
              value: "false"
            # Set Felix logging to "info"
            - name: FELIX_LOGSEVERITYSCREEN
              value: "{{- or .Networking.Canal.LogSeveritySys "INFO" }}"
            # Set Felix endpoint to host default action to ACCEPT.
            - name: FELIX_DEFAULTENDPOINTTOHOSTACTION
              value: "{{- or .Networking.Canal.DefaultEndpointToHostAction "ACCEPT" }}"
            # Controls whether Felix inserts rules to the top of iptables chains, or appends to the bottom
            - name: FELIX_CHAININSERTMODE
              value: "{{- or .Networking.Canal.ChainInsertMode "insert" }}"
            # Set to enable the experimental Prometheus metrics server
            - name: FELIX_PROMETHEUSMETRICSENABLED
              value: "{{- or .Networking.Canal.PrometheusMetricsEnabled "false" }}"
            # TCP port that the Prometheus metrics server should bind to
            - name: FELIX_PROMETHEUSMETRICSPORT
              value: "{{- or .Networking.Canal.PrometheusMetricsPort "9091" }}"
            # Enable Prometheus Go runtime metrics collection
            - name: FELIX_PROMETHEUSGOMETRICSENABLED
              value: "{{- or .Networking.Canal.PrometheusGoMetricsEnabled "true" }}"
            # Enable Prometheus process metrics collection
            - name: FELIX_PROMETHEUSPROCESSMETRICSENABLED
              value: "{{- or .Networking.Canal.PrometheusProcessMetricsEnabled "true" }}"
            - name: FELIX_HEALTHENABLED
              value: "true"
          securityContext:
            privileged: true
          resources:
            requests:
              cpu: 250m
          livenessProbe:
            httpGet:
              path: /liveness
              port: 9099
              host: localhost
            periodSeconds: 10
            initialDelaySeconds: 10
            failureThreshold: 6
          readinessProbe:
            httpGet:
              path: /readiness
              port: 9099
              host: localhost
            periodSeconds: 10
          volumeMounts:
            - mountPath: /lib/modules
              name: lib-modules
              readOnly: true
            - mountPath: /var/run/calico
              name: var-run-calico
              readOnly: false
            - mountPath: /var/lib/calico
              name: var-lib-calico
              readOnly: false
        # This container installs the Calico CNI binaries
        # and CNI network config file on each node.
        - name: install-cni
          image: quay.io/calico/cni:v3.2.3
          command: ["/install-cni.sh"]
          env:
            # Name of the CNI config file to create.
            - name: CNI_CONF_NAME
              value: "10-canal.conflist"
            # Set the hostname based on the k8s node name.
            - name: KUBERNETES_NODE_NAME
              valueFrom:
                fieldRef:
                  fieldPath: spec.nodeName
            # The CNI network config to install on each node.
            - name: CNI_NETWORK_CONFIG
              valueFrom:
                configMapKeyRef:
                  name: canal-config
                  key: cni_network_config
          volumeMounts:
            - mountPath: /host/opt/cni/bin
              name: cni-bin-dir
            - mountPath: /host/etc/cni/net.d
              name: cni-net-dir
        # This container runs flannel using the kube-subnet-mgr backend
        # for allocating subnets.
        - name: kube-flannel
          image: quay.io/coreos/flannel:v0.9.0
          command: [ "/opt/bin/flanneld", "--ip-masq", "--kube-subnet-mgr" ]
          securityContext:
            privileged: true
          env:
            - name: POD_NAME
              valueFrom:
                fieldRef:
                  fieldPath: metadata.name
            - name: POD_NAMESPACE
              valueFrom:
                fieldRef:
                  fieldPath: metadata.namespace
            - name: FLANNELD_IFACE
              valueFrom:
                configMapKeyRef:
                  name: canal-config
                  key: canal_iface
            - name: FLANNELD_IP_MASQ
              valueFrom:
                configMapKeyRef:
                  name: canal-config
                  key: masquerade
          volumeMounts:
          - name: run
            mountPath: /run
          - name: flannel-cfg
            mountPath: /etc/kube-flannel/
      volumes:
        # Used by calico/node.
        - name: lib-modules
          hostPath:
            path: /lib/modules
        - name: var-run-calico
          hostPath:
            path: /var/run/calico
        - name: var-lib-calico
          hostPath:
            path: /var/lib/calico
        # Used by flannel.
        - name: run
          hostPath:
            path: /run
        - name: flannel-cfg
          configMap:
            name: canal-config
        # Used to install CNI.
        - name: cni-bin-dir
          hostPath:
            path: /opt/cni/bin
        - name: cni-net-dir
          hostPath:
            path: /etc/cni/net.d
---

apiVersion: v1
kind: ServiceAccount
metadata:
  name: canal
  namespace: kube-system

---

kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: calico
rules:
  - apiGroups: [""]
    resources:
      - namespaces
      - serviceaccounts
    verbs:
      - get
      - list
      - watch
  - apiGroups: [""]
    resources:
      - pods/status
    verbs:
      - update
  - apiGroups: [""]
    resources:
      - pods
    verbs:
      - get
      - list
      - watch
      - patch
  - apiGroups: [""]
    resources:
      - services
    verbs:
      - get
  - apiGroups: [""]
    resources:
      - endpoints
    verbs:
      - get
  - apiGroups: [""]
    resources:
      - nodes
    verbs:
      - get
      - list
      - update
      - watch
  - apiGroups: ["networking.k8s.io"]
    resources:
      - networkpolicies
    verbs:
      - get
      - list
      - watch
  - apiGroups: ["crd.projectcalico.org"]
    resources:
      - globalfelixconfigs
      - felixconfigurations
      - bgppeers
      - globalbgpconfigs
      - globalnetworksets
      - hostendpoints
      - bgpconfigurations
      - ippools
      - globalnetworkpolicies
      - networkpolicies
      - clusterinformations
    verbs:
      - create
      - get
      - list
      - update
      - watch

---

# Flannel roles
# Pulled from https://github.com/coreos/flannel/blob/master/Documentation/kube-flannel-rbac.yml
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: flannel
rules:
  - apiGroups:
      - ""
    resources:
      - pods
    verbs:
      - get
  - apiGroups:
      - ""
    resources:
      - nodes
    verbs:
      - list
      - watch
  - apiGroups:
      - ""
    resources:
      - nodes/status
    verbs:
      - patch
---

# Bind the flannel ClusterRole to the canal ServiceAccount.
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: canal-flannel
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: flannel
subjects:
- kind: ServiceAccount
  name: canal
  namespace: kube-system

---

# Bind the ClusterRole to the canal ServiceAccount.
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: canal-calico
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: calico
subjects:
- kind: ServiceAccount
  name: canal
  namespace: kube-system

---

# Create all the CustomResourceDefinitions needed for
# Calico policy and networking mode.

apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
   name: felixconfigurations.crd.projectcalico.org
spec:
  scope: Cluster
  group: crd.projectcalico.org
  version: v1
  names:
    kind: FelixConfiguration
    plural: felixconfigurations
    singular: felixconfiguration
---

apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: bgpconfigurations.crd.projectcalico.org
spec:
  scope: Cluster
  group: crd.projectcalico.org
  version: v1
  names:
    kind: BGPConfiguration
    plural: bgpconfigurations
    singular: bgpconfiguration

---

apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: ippools.crd.projectcalico.org
spec:
  scope: Cluster
  group: crd.projectcalico.org
  version: v1
  names:
    kind: IPPool
    plural: ippools
    singular: ippool

---

apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: hostendpoints.crd.projectcalico.org
spec:
  scope: Cluster
  group: crd.projectcalico.org
  version: v1
  names:
    kind: HostEndpoint
    plural: hostendpoints
    singular: hostendpoint

---

apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: clusterinformations.crd.projectcalico.org
spec:
  scope: Cluster
  group: crd.projectcalico.org
  version: v1
  names:
    kind: ClusterInformation
    plural: clusterinformations
    singular: clusterinformation

---

apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: globalnetworkpolicies.crd.projectcalico.org
spec:
  scope: Cluster
  group: crd.projectcalico.org
  version: v1
  names:
    kind: GlobalNetworkPolicy
    plural: globalnetworkpolicies
    singular: globalnetworkpolicy

---

apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: globalnetworksets.crd.projectcalico.org
spec:
  scope: Cluster
  group: crd.projectcalico.org
  version: v1
  names:
    kind: GlobalNetworkSet
    plural: globalnetworksets
    singular: globalnetworkset

---

apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: networkpolicies.crd.projectcalico.org
spec:
  scope: Namespaced
  group: crd.projectcalico.org
  version: v1
  names:
    kind: NetworkPolicy
    plural: networkpolicies
    singular: networkpolicy
`)

func cloudupResourcesAddonsNetworkingProjectcalicoOrgCanalK8s19YamlTemplateBytes() ([]byte, error) {
	return _cloudupResourcesAddonsNetworkingProjectcalicoOrgCanalK8s19YamlTemplate, nil
}

func cloudupResourcesAddonsNetworkingProjectcalicoOrgCanalK8s19YamlTemplate() (*asset, error) {
	bytes, err := cloudupResourcesAddonsNetworkingProjectcalicoOrgCanalK8s19YamlTemplateBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "cloudup/resources/addons/networking.projectcalico.org.canal/k8s-1.9.yaml.template", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _cloudupResourcesAddonsNetworkingRomanaK8s112YamlTemplate = []byte(`---
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: romana-listener
rules:
- apiGroups:
  - "*"
  resources:
  - pods
  - namespaces
  - nodes
  - endpoints
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - extensions
  resources:
  - networkpolicies
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - "*"
  resources:
  - services
  verbs:
  - update
  - list
  - watch
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: romana-listener
  namespace: kube-system
---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: romana-listener
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: romana-listener
subjects:
- kind: ServiceAccount
  name: romana-listener
  namespace: kube-system
---
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: romana-agent
rules:
- apiGroups:
  - "*"
  resources:
  - pods
  - nodes
  verbs:
  - get
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: romana-agent
  namespace: kube-system
---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: romana-agent
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: romana-agent
subjects:
- kind: ServiceAccount
  name: romana-agent
  namespace: kube-system
---
apiVersion: v1
kind: Service
metadata:
  name: romana-etcd
  namespace: kube-system
spec:
  clusterIP: {{ .Networking.Romana.EtcdServiceIP }}
  ports:
  - name: etcd
    port: 12379
    protocol: TCP
    targetPort: 4001
  selector:
    k8s-app: etcd-server
  sessionAffinity: None
  type: ClusterIP
---
apiVersion: v1
kind: Service
metadata:
  name: romana
  namespace: kube-system
spec:
  clusterIP: {{ .Networking.Romana.DaemonServiceIP }}
  ports:
  - name: daemon
    port: 9600
    protocol: TCP
    targetPort: 9600
  selector:
    romana-app: daemon
  sessionAffinity: None
  type: ClusterIP
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: romana-daemon
  namespace: kube-system
  labels:
    romana-app: daemon
spec:
  replicas: 1
  selector:
    matchLabels:
      romana-app: daemon
  template:
    metadata:
      labels:
        romana-app: daemon
    spec:
      nodeSelector:
        node-role.kubernetes.io/master: ""
      hostNetwork: true
      priorityClassName: system-cluster-critical
      tolerations:
      - key: node-role.kubernetes.io/master
        effect: NoSchedule
      containers:
      - name: romana-daemon
        image: quay.io/romana/daemon:v2.0.2
        imagePullPolicy: Always
        resources:
          requests:
            cpu: 10m
            memory: 64Mi
          limits:
            memory: 64Mi
        args:
        - --cloud=aws
        - --network-cidr-overrides=romana-network={{ .KubeControllerManager.ClusterCIDR }}
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: romana-listener
  namespace: kube-system
spec:
  replicas: 1
  selector:
    matchLabels:
      romana-app: listener
  template:
    metadata:
      labels:
        romana-app: listener
    spec:
      nodeSelector:
        node-role.kubernetes.io/master: ""
      hostNetwork: true
      priorityClassName: system-cluster-critical
      serviceAccountName: romana-listener
      tolerations:
      - key: node-role.kubernetes.io/master
        effect: NoSchedule
      containers:
      - name: romana-listener
        image: quay.io/romana/listener:v2.0.2
        imagePullPolicy: Always
        resources:
          requests:
            cpu: 10m
            memory: 64Mi
          limits:
            memory: 64Mi
---
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: romana-agent
  namespace: kube-system
  labels:
    romana-app: agent
spec:
  updateStrategy:
    type: RollingUpdate
  selector:
    matchLabels:
      romana-app: agent
  template:
    metadata:
      labels:
        romana-app: agent
    spec:
      hostNetwork: true
      priorityClassName: system-node-critical
      securityContext:
        seLinuxOptions:
          type: spc_t
      serviceAccountName: romana-agent
      tolerations:
      - effect: NoSchedule
        operator: Exists
      - effect: NoExecute
        operator: Exists
      containers:
      - name: romana-agent
        image: quay.io/romana/agent:v2.0.2
        imagePullPolicy: Always
        resources:
          requests:
            cpu: 25m
            memory: 128Mi
          limits:
            memory: 128Mi
        env:
        - name: NODENAME
          valueFrom:
            fieldRef:
              fieldPath: spec.nodeName
        - name: NODEIP
          valueFrom:
            fieldRef:
              fieldPath: status.hostIP
        args:
        - --service-cluster-ip-range={{ .ServiceClusterIPRange }}
        securityContext:
          privileged: true
        volumeMounts:
        - name: host-usr-local-bin
          mountPath: /host/usr/local/bin
        - name: host-etc-romana
          mountPath: /host/etc/romana
        - name: host-cni-bin
          mountPath: /host/opt/cni/bin
        - name: host-cni-net-d
          mountPath: /host/etc/cni/net.d
        - name: run-path
          mountPath: /var/run/romana
      volumes:
      - name: host-usr-local-bin
        hostPath:
          path: /usr/local/bin
      - name: host-etc-romana
        hostPath:
          path: /etc/romana
      - name: host-cni-bin
        hostPath:
          path: /opt/cni/bin
      - name: host-cni-net-d
        hostPath:
          path: /etc/cni/net.d
      - name: run-path
        hostPath:
          path: /var/run/romana
---
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: romana-aws
rules:
- apiGroups:
  - "*"
  resources:
  - nodes
  verbs:
  - get
  - list
  - watch
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: romana-aws
  namespace: kube-system
---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: romana-aws
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: romana-aws
subjects:
- kind: ServiceAccount
  name: romana-aws
  namespace: kube-system
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: romana-aws
  namespace: kube-system
  labels:
    romana-app: aws
spec:
  replicas: 1
  selector:
    matchLabels:
      romana-app: aws
  template:
    metadata:
      labels:
        romana-app: aws
    spec:
      nodeSelector:
        node-role.kubernetes.io/master: ""
      hostNetwork: true
      priorityClassName: system-cluster-critical
      serviceAccountName: romana-aws
      tolerations:
      - key: node-role.kubernetes.io/master
        effect: NoSchedule
      containers:
      - name: romana-aws
        image: quay.io/romana/aws:v2.0.2
        imagePullPolicy: Always
        resources:
          requests:
            cpu: 10m
            memory: 64Mi
          limits:
            memory: 64Mi
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: romana-vpcrouter
  namespace: kube-system
  labels:
    romana-app: vpcrouter
spec:
  replicas: 1
  selector:
    matchLabels:
      romana-app: vpcrouter
  template:
    metadata:
      labels:
        romana-app: vpcrouter
    spec:
      nodeSelector:
        node-role.kubernetes.io/master: ""
      hostNetwork: true
      priorityClassName: system-cluster-critical
      tolerations:
      - key: node-role.kubernetes.io/master
        effect: NoSchedule
      containers:
      - name: romana-vpcrouter
        image: quay.io/romana/vpcrouter-romana-plugin:1.1.17
        imagePullPolicy: Always
        resources:
          requests:
            cpu: 45m
            memory: 128Mi
          limits:
            memory: 128Mi
        args:
        - --etcd_use_v2
        - --etcd_addr={{ .Networking.Romana.EtcdServiceIP }}
        - --etcd_port=12379
`)

func cloudupResourcesAddonsNetworkingRomanaK8s112YamlTemplateBytes() ([]byte, error) {
	return _cloudupResourcesAddonsNetworkingRomanaK8s112YamlTemplate, nil
}

func cloudupResourcesAddonsNetworkingRomanaK8s112YamlTemplate() (*asset, error) {
	bytes, err := cloudupResourcesAddonsNetworkingRomanaK8s112YamlTemplateBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "cloudup/resources/addons/networking.romana/k8s-1.12.yaml.template", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _cloudupResourcesAddonsNetworkingRomanaK8s17YamlTemplate = []byte(`---
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: romana-listener
rules:
- apiGroups:
  - "*"
  resources:
  - pods
  - namespaces
  - nodes
  - endpoints
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - extensions
  resources:
  - networkpolicies
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - "*"
  resources:
  - services
  verbs:
  - update
  - list
  - watch
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: romana-listener
  namespace: kube-system
---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: romana-listener
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: romana-listener
subjects:
- kind: ServiceAccount
  name: romana-listener
  namespace: kube-system
---
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: romana-agent
rules:
- apiGroups:
  - "*"
  resources:
  - pods
  - nodes
  verbs:
  - get
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: romana-agent
  namespace: kube-system
---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: romana-agent
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: romana-agent
subjects:
- kind: ServiceAccount
  name: romana-agent
  namespace: kube-system
---
apiVersion: v1
kind: Service
metadata:
  name: romana-etcd
  namespace: kube-system
spec:
  clusterIP: {{ .Networking.Romana.EtcdServiceIP }}
  ports:
  - name: etcd
    port: 12379
    protocol: TCP
    targetPort: 4001
  selector:
    k8s-app: etcd-server
  sessionAffinity: None
  type: ClusterIP
---
apiVersion: v1
kind: Service
metadata:
  name: romana
  namespace: kube-system
spec:
  clusterIP: {{ .Networking.Romana.DaemonServiceIP }}
  ports:
  - name: daemon
    port: 9600
    protocol: TCP
    targetPort: 9600
  selector:
    romana-app: daemon
  sessionAffinity: None
  type: ClusterIP
---
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: romana-daemon
  namespace: kube-system
spec:
  replicas: 1
  template:
    metadata:
      labels:
        romana-app: daemon
    spec:
      nodeSelector:
        node-role.kubernetes.io/master: ""
      hostNetwork: true
      tolerations:
      - key: node-role.kubernetes.io/master
        effect: NoSchedule
      containers:
      - name: romana-daemon
        image: quay.io/romana/daemon:v2.0.2
        imagePullPolicy: Always
        resources:
          requests:
            cpu: 10m
            memory: 64Mi
          limits:
            memory: 64Mi
        args:
        - --cloud=aws
        - --network-cidr-overrides=romana-network={{ .KubeControllerManager.ClusterCIDR }}
---
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: romana-listener
  namespace: kube-system
spec:
  replicas: 1
  template:
    metadata:
      labels:
        romana-app: listener
    spec:
      nodeSelector:
        node-role.kubernetes.io/master: ""
      hostNetwork: true
      serviceAccountName: romana-listener
      tolerations:
      - key: node-role.kubernetes.io/master
        effect: NoSchedule
      containers:
      - name: romana-listener
        image: quay.io/romana/listener:v2.0.2
        imagePullPolicy: Always
        resources:
          requests:
            cpu: 10m
            memory: 64Mi
          limits:
            memory: 64Mi
---
apiVersion: extensions/v1beta1
kind: DaemonSet
metadata:
  name: romana-agent
  namespace: kube-system
spec:
  updateStrategy:
    type: RollingUpdate
  template:
    metadata:
      labels:
        romana-app: agent
    spec:
      hostNetwork: true
      securityContext:
        seLinuxOptions:
          type: spc_t
      serviceAccountName: romana-agent
      tolerations:
      - effect: NoSchedule
        operator: Exists
      - effect: NoExecute
        operator: Exists
      containers:
      - name: romana-agent
        image: quay.io/romana/agent:v2.0.2
        imagePullPolicy: Always
        resources:
          requests:
            cpu: 25m
            memory: 128Mi
          limits:
            memory: 128Mi
        env:
        - name: NODENAME
          valueFrom:
            fieldRef:
              fieldPath: spec.nodeName
        - name: NODEIP
          valueFrom:
            fieldRef:
              fieldPath: status.hostIP
        args:
        - --service-cluster-ip-range={{ .ServiceClusterIPRange }}
        securityContext:
          privileged: true
        volumeMounts:
        - name: host-usr-local-bin
          mountPath: /host/usr/local/bin
        - name: host-etc-romana
          mountPath: /host/etc/romana
        - name: host-cni-bin
          mountPath: /host/opt/cni/bin
        - name: host-cni-net-d
          mountPath: /host/etc/cni/net.d
        - name: run-path
          mountPath: /var/run/romana
      volumes:
      - name: host-usr-local-bin
        hostPath:
          path: /usr/local/bin
      - name: host-etc-romana
        hostPath:
          path: /etc/romana
      - name: host-cni-bin
        hostPath:
          path: /opt/cni/bin
      - name: host-cni-net-d
        hostPath:
          path: /etc/cni/net.d
      - name: run-path
        hostPath:
          path: /var/run/romana
---
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: romana-aws
rules:
- apiGroups:
  - "*"
  resources:
  - nodes
  verbs:
  - get
  - list
  - watch
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: romana-aws
  namespace: kube-system
---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: romana-aws
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: romana-aws
subjects:
- kind: ServiceAccount
  name: romana-aws
  namespace: kube-system
---
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: romana-aws
  namespace: kube-system
spec:
  replicas: 1
  template:
    metadata:
      labels:
        romana-app: aws
    spec:
      nodeSelector:
        node-role.kubernetes.io/master: ""
      hostNetwork: true
      serviceAccountName: romana-aws
      tolerations:
      - key: node-role.kubernetes.io/master
        effect: NoSchedule
      containers:
      - name: romana-aws
        image: quay.io/romana/aws:v2.0.2
        imagePullPolicy: Always
        resources:
          requests:
            cpu: 10m
            memory: 64Mi
          limits:
            memory: 64Mi
---
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: romana-vpcrouter
  namespace: kube-system
spec:
  replicas: 1
  template:
    metadata:
      labels:
        romana-app: vpcrouter
    spec:
      nodeSelector:
        node-role.kubernetes.io/master: ""
      hostNetwork: true
      tolerations:
      - key: node-role.kubernetes.io/master
        effect: NoSchedule
      containers:
      - name: romana-vpcrouter
        image: quay.io/romana/vpcrouter-romana-plugin:1.1.17
        imagePullPolicy: Always
        resources:
          requests:
            cpu: 45m
            memory: 128Mi
          limits:
            memory: 128Mi
        args:
        - --etcd_use_v2
        - --etcd_addr={{ .Networking.Romana.EtcdServiceIP }}
        - --etcd_port=12379
`)

func cloudupResourcesAddonsNetworkingRomanaK8s17YamlTemplateBytes() ([]byte, error) {
	return _cloudupResourcesAddonsNetworkingRomanaK8s17YamlTemplate, nil
}

func cloudupResourcesAddonsNetworkingRomanaK8s17YamlTemplate() (*asset, error) {
	bytes, err := cloudupResourcesAddonsNetworkingRomanaK8s17YamlTemplateBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "cloudup/resources/addons/networking.romana/k8s-1.7.yaml.template", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _cloudupResourcesAddonsNetworkingWeaveK8s112YamlTemplate = []byte(`{{- if WeaveSecret }}
apiVersion: v1
kind: Secret
metadata:
  name: weave-net
  namespace: kube-system
stringData:
  network-password: {{ WeaveSecret }}
---
{{- end }}

apiVersion: v1
kind: ServiceAccount
metadata:
  name: weave-net
  namespace: kube-system
  labels:
    name: weave-net
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: weave-net
  namespace: kube-system
  labels:
    name: weave-net
    role.kubernetes.io/networking: "1"
rules:
  - apiGroups:
      - ''
    resources:
      - pods
      - namespaces
      - nodes
    verbs:
      - get
      - list
      - watch
  - apiGroups:
      - 'networking.k8s.io'
    resources:
      - networkpolicies
    verbs:
      - get
      - list
      - watch
  - apiGroups:
      - ''
    resources:
      - nodes/status
    verbs:
      - patch
      - update
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: weave-net
  namespace: kube-system
  labels:
    name: weave-net
    role.kubernetes.io/networking: "1"
roleRef:
  kind: ClusterRole
  name: weave-net
  apiGroup: rbac.authorization.k8s.io
subjects:
  - kind: ServiceAccount
    name: weave-net
    namespace: kube-system
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: weave-net
  namespace: kube-system
  labels:
    name: weave-net
rules:
  - apiGroups:
      - ''
    resources:
      - configmaps
    resourceNames:
      - weave-net
    verbs:
      - get
      - update
  - apiGroups:
      - ''
    resources:
      - configmaps
    verbs:
      - create
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: weave-net
  namespace: kube-system
  labels:
    name: weave-net
roleRef:
  kind: Role
  name: weave-net
  apiGroup: rbac.authorization.k8s.io
subjects:
  - kind: ServiceAccount
    name: weave-net
    namespace: kube-system
---
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: weave-net
  namespace: kube-system
  labels:
    name: weave-net
    role.kubernetes.io/networking: "1"
spec:
  # Wait 5 seconds to let pod connect before rolling next pod
  minReadySeconds: 5
  selector:
    matchLabels:
      name: weave-net
      role.kubernetes.io/networking: "1"
  template:
    metadata:
      labels:
        name: weave-net
        role.kubernetes.io/networking: "1"
      annotations:
        prometheus.io/scrape: "true"
        scheduler.alpha.kubernetes.io/critical-pod: ''
    spec:
      containers:
        - name: weave
          command:
            - /home/weave/launch.sh
          env:
            - name: HOSTNAME
              valueFrom:
                fieldRef:
                  apiVersion: v1
                  fieldPath: spec.nodeName
            - name: IPALLOC_RANGE
              value: {{ .KubeControllerManager.ClusterCIDR }}
            {{- if .Networking.Weave.MTU }}
            - name: WEAVE_MTU
              value: "{{ .Networking.Weave.MTU }}"
            {{- end }}
            {{- if .Networking.Weave.NoMasqLocal }}
            - name: NO_MASQ_LOCAL
              value: "{{ .Networking.Weave.NoMasqLocal }}"
            {{- end }}
            {{- if .Networking.Weave.ConnLimit }}
            - name: CONN_LIMIT
              value: "{{ .Networking.Weave.ConnLimit }}"
            {{- end }}
            {{- if .Networking.Weave.NetExtraArgs }}
            - name: EXTRA_ARGS
              value: "{{ .Networking.Weave.NetExtraArgs }}"
            {{- end }}
            {{- if WeaveSecret }}
            - name: WEAVE_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: weave-net
                  key: network-password
            {{- end }}
          image: 'weaveworks/weave-kube:2.6.2'
          ports:
            - name: metrics
              containerPort: 6782
          readinessProbe:
            httpGet:
              host: 127.0.0.1
              path: /status
              port: 6784
          resources:
            requests:
              cpu: {{ or .Networking.Weave.CPURequest "50m" }}
              memory: {{ or .Networking.Weave.MemoryRequest "200Mi" }}
            limits:
              {{- if .Networking.Weave.CPULimit }}
              cpu: {{ .Networking.Weave.CPULimit }}
              {{- end }}
              memory: {{ or .Networking.Weave.MemoryLimit "200Mi" }}
          securityContext:
            privileged: true
          volumeMounts:
            - name: weavedb
              mountPath: /weavedb
            - name: cni-bin
              mountPath: /host/opt
            - name: cni-bin2
              mountPath: /host/home
            - name: cni-conf
              mountPath: /host/etc
            - name: dbus
              mountPath: /host/var/lib/dbus
            - name: lib-modules
              mountPath: /lib/modules
            - name: xtables-lock
              mountPath: /run/xtables.lock
              readOnly: false
        - name: weave-npc
          args: []
          env:
            - name: HOSTNAME
              valueFrom:
                fieldRef:
                  apiVersion: v1
                  fieldPath: spec.nodeName
            {{- if .Networking.Weave.NPCExtraArgs }}
            - name: EXTRA_ARGS
              value: "{{ .Networking.Weave.NPCExtraArgs }}"
            {{- end }}
          image: 'weaveworks/weave-npc:2.6.2'
          ports:
            - name: metrics
              containerPort: 6781
          resources:
            requests:
              cpu: {{ or .Networking.Weave.NPCCPURequest "50m" }}
              memory: {{ or .Networking.Weave.NPCMemoryRequest "200Mi" }}
            limits:
              {{- if .Networking.Weave.NPCCPULimit }}
              cpu: {{ .Networking.Weave.NPCCPULimit }}
              {{- end }}
              memory: {{ or .Networking.Weave.NPCMemoryLimit "200Mi" }}
          securityContext:
            privileged: true
          volumeMounts:
            - name: xtables-lock
              mountPath: /run/xtables.lock
              readOnly: false
      hostNetwork: true
      dnsPolicy: ClusterFirstWithHostNet
      hostPID: true
      restartPolicy: Always
      securityContext:
        seLinuxOptions: {}
      serviceAccountName: weave-net
      tolerations:
        - effect: NoSchedule
          operator: Exists
        - key: CriticalAddonsOnly
          operator: Exists
      volumes:
        - name: weavedb
          hostPath:
            path: /var/lib/weave
        - name: cni-bin
          hostPath:
            path: /opt
        - name: cni-bin2
          hostPath:
            path: /home
        - name: cni-conf
          hostPath:
            path: /etc
        - name: dbus
          hostPath:
            path: /var/lib/dbus
        - name: lib-modules
          hostPath:
            path: /lib/modules
        - name: xtables-lock
          hostPath:
            path: /run/xtables.lock
            type: FileOrCreate
      priorityClassName: system-node-critical
  updateStrategy:
    type: RollingUpdate
`)

func cloudupResourcesAddonsNetworkingWeaveK8s112YamlTemplateBytes() ([]byte, error) {
	return _cloudupResourcesAddonsNetworkingWeaveK8s112YamlTemplate, nil
}

func cloudupResourcesAddonsNetworkingWeaveK8s112YamlTemplate() (*asset, error) {
	bytes, err := cloudupResourcesAddonsNetworkingWeaveK8s112YamlTemplateBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "cloudup/resources/addons/networking.weave/k8s-1.12.yaml.template", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _cloudupResourcesAddonsNetworkingWeaveK8s18YamlTemplate = []byte(`{{- if WeaveSecret }}
apiVersion: v1
kind: Secret
metadata:
  name: weave-net
  namespace: kube-system
stringData:
  network-password: {{ WeaveSecret }}
---
{{- end }}

apiVersion: v1
kind: ServiceAccount
metadata:
  name: weave-net
  namespace: kube-system
  labels:
    name: weave-net
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  name: weave-net
  namespace: kube-system
  labels:
    name: weave-net
    role.kubernetes.io/networking: "1"
rules:
  - apiGroups:
      - ''
    resources:
      - pods
      - namespaces
      - nodes
    verbs:
      - get
      - list
      - watch
  - apiGroups:
      - 'networking.k8s.io'
    resources:
      - networkpolicies
    verbs:
      - get
      - list
      - watch
  - apiGroups:
      - ''
    resources:
      - nodes/status
    verbs:
      - patch
      - update
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: weave-net
  namespace: kube-system
  labels:
    name: weave-net
    role.kubernetes.io/networking: "1"
roleRef:
  kind: ClusterRole
  name: weave-net
  apiGroup: rbac.authorization.k8s.io
subjects:
  - kind: ServiceAccount
    name: weave-net
    namespace: kube-system
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: Role
metadata:
  name: weave-net
  namespace: kube-system
  labels:
    name: weave-net
rules:
  - apiGroups:
      - ''
    resources:
      - configmaps
    resourceNames:
      - weave-net
    verbs:
      - get
      - update
  - apiGroups:
      - ''
    resources:
      - configmaps
    verbs:
      - create
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: RoleBinding
metadata:
  name: weave-net
  namespace: kube-system
  labels:
    name: weave-net
roleRef:
  kind: Role
  name: weave-net
  apiGroup: rbac.authorization.k8s.io
subjects:
  - kind: ServiceAccount
    name: weave-net
    namespace: kube-system
---
apiVersion: extensions/v1beta1
kind: DaemonSet
metadata:
  name: weave-net
  namespace: kube-system
  labels:
    name: weave-net
    role.kubernetes.io/networking: "1"
spec:
  # Wait 5 seconds to let pod connect before rolling next pod
  minReadySeconds: 5
  template:
    metadata:
      labels:
        name: weave-net
        role.kubernetes.io/networking: "1"
      annotations:
        prometheus.io/scrape: "true"
        scheduler.alpha.kubernetes.io/critical-pod: ''
    spec:
      containers:
        - name: weave
          command:
            - /home/weave/launch.sh
          env:
            - name: HOSTNAME
              valueFrom:
                fieldRef:
                  apiVersion: v1
                  fieldPath: spec.nodeName
            - name: IPALLOC_RANGE
              value: {{ .KubeControllerManager.ClusterCIDR }}
            {{- if .Networking.Weave.MTU }}
            - name: WEAVE_MTU
              value: "{{ .Networking.Weave.MTU }}"
            {{- end }}
            {{- if .Networking.Weave.NoMasqLocal }}
            - name: NO_MASQ_LOCAL
              value: "{{ .Networking.Weave.NoMasqLocal }}"
            {{- end }}
            {{- if .Networking.Weave.ConnLimit }}
            - name: CONN_LIMIT
              value: "{{ .Networking.Weave.ConnLimit }}"
            {{- end }}
            {{- if .Networking.Weave.NetExtraArgs }}
            - name: EXTRA_ARGS
              value: "{{ .Networking.Weave.NetExtraArgs }}"
            {{- end }}
            {{- if WeaveSecret }}
            - name: WEAVE_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: weave-net
                  key: network-password
            {{- end }}
          image: 'weaveworks/weave-kube:2.6.2'
          ports:
            - name: metrics
              containerPort: 6782
          readinessProbe:
            httpGet:
              host: 127.0.0.1
              path: /status
              port: 6784
          resources:
            requests:
              cpu: {{ or .Networking.Weave.CPURequest "50m" }}
              memory: {{ or .Networking.Weave.MemoryRequest "200Mi" }}
            limits:
              {{- if .Networking.Weave.CPULimit }}
              cpu: {{ .Networking.Weave.CPULimit }}
              {{- end }}
              memory: {{ or .Networking.Weave.MemoryLimit "200Mi" }}
          securityContext:
            privileged: true
          volumeMounts:
            - name: weavedb
              mountPath: /weavedb
            - name: cni-bin
              mountPath: /host/opt
            - name: cni-bin2
              mountPath: /host/home
            - name: cni-conf
              mountPath: /host/etc
            - name: dbus
              mountPath: /host/var/lib/dbus
            - name: lib-modules
              mountPath: /lib/modules
            - name: xtables-lock
              mountPath: /run/xtables.lock
              readOnly: false
        - name: weave-npc
          args: []
          env:
            - name: HOSTNAME
              valueFrom:
                fieldRef:
                  apiVersion: v1
                  fieldPath: spec.nodeName
            {{- if .Networking.Weave.NPCExtraArgs }}
            - name: EXTRA_ARGS
              value: "{{ .Networking.Weave.NPCExtraArgs }}"
            {{- end }}
          image: 'weaveworks/weave-npc:2.6.2'
          ports:
            - name: metrics
              containerPort: 6781
          resources:
            requests:
              cpu: {{ or .Networking.Weave.NPCCPURequest "50m" }}
              memory: {{ or .Networking.Weave.NPCMemoryRequest "200Mi" }}
            limits:
              {{- if .Networking.Weave.NPCCPULimit }}
              cpu: {{ .Networking.Weave.NPCCPULimit }}
              {{- end }}
              memory: {{ or .Networking.Weave.NPCMemoryLimit "200Mi" }}
          securityContext:
            privileged: true
          volumeMounts:
            - name: xtables-lock
              mountPath: /run/xtables.lock
              readOnly: false
      hostNetwork: true
      hostPID: true
      restartPolicy: Always
      securityContext:
        seLinuxOptions: {}
      serviceAccountName: weave-net
      tolerations:
        - effect: NoSchedule
          operator: Exists
        - key: CriticalAddonsOnly
          operator: Exists
      volumes:
        - name: weavedb
          hostPath:
            path: /var/lib/weave
        - name: cni-bin
          hostPath:
            path: /opt
        - name: cni-bin2
          hostPath:
            path: /home
        - name: cni-conf
          hostPath:
            path: /etc
        - name: dbus
          hostPath:
            path: /var/lib/dbus
        - name: lib-modules
          hostPath:
            path: /lib/modules
        - name: xtables-lock
          hostPath:
            path: /run/xtables.lock
            type: FileOrCreate
  updateStrategy:
    type: RollingUpdate
`)

func cloudupResourcesAddonsNetworkingWeaveK8s18YamlTemplateBytes() ([]byte, error) {
	return _cloudupResourcesAddonsNetworkingWeaveK8s18YamlTemplate, nil
}

func cloudupResourcesAddonsNetworkingWeaveK8s18YamlTemplate() (*asset, error) {
	bytes, err := cloudupResourcesAddonsNetworkingWeaveK8s18YamlTemplateBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "cloudup/resources/addons/networking.weave/k8s-1.8.yaml.template", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _cloudupResourcesAddonsNodeAuthorizerAddonsK8sIoK8s110YamlTemplate = []byte(`{{- $proxy := .EgressProxy }}
{{- $na := .NodeAuthorization.NodeAuthorizer }}
{{- $name := "node-authorizer" }}
{{- $namespace := "kube-system" }}
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: {{ $name }}
  namespace: {{ $namespace }}
  labels:
    k8s-app: {{ $name }}
    k8s-addon: {{ $name }}.addons.k8s.io
---
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: kops:{{ $name }}:nodes-viewer
  labels:
    k8s-app: {{ $name }}
    k8s-addon: {{ $name }}.addons.k8s.io
rules:
- apiGroups:
  - "*"
  resources:
  - nodes
  verbs:
  - get
  - list
---
# permits the node access to create a CSR
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: kops:{{ $name }}:system:bootstrappers
  labels:
    k8s-app: {{ $name }}
    k8s-addon: {{ $name }}.addons.k8s.io
roleRef:
  kind: ClusterRole
  name: system:node-bootstrapper
  apiGroup: rbac.authorization.k8s.io
subjects:
- kind: Group
  name: system:bootstrappers
  apiGroup: rbac.authorization.k8s.io
---
# indicates to the controller to auto-sign the CSR for this group
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: kops:{{ $name }}:approval
  labels:
    k8s-app: {{ $name }}
    k8s-addon: {{ $name }}.addons.k8s.io
roleRef:
  kind: ClusterRole
  name: system:certificates.k8s.io:certificatesigningrequests:nodeclient
  apiGroup: rbac.authorization.k8s.io
subjects:
- kind: Group
  name: system:bootstrappers
  apiGroup: rbac.authorization.k8s.io
---
# the service permission requires to create the bootstrap tokens
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: Role
metadata:
  name: kops:{{ $namespace }}:{{ $name }}
  namespace: {{ $namespace }}
  labels:
    k8s-app: {{ $name }}
    k8s-addon: {{ $name }}.addons.k8s.io
rules:
- apiGroups:
  - "*"
  resources:
  - secrets
  verbs:
  - create
  - list
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: RoleBinding
metadata:
  name: kops:{{ $namespace }}:{{ $name }}
  namespace: {{ $namespace }}
  labels:
    k8s-app: {{ $name }}
    k8s-addon: {{ $name }}.addons.k8s.io
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: kops:{{ $namespace }}:{{ $name }}
subjects:
- kind: ServiceAccount
  name: {{ $name }}
  namespace: {{ $namespace }}
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: kops:{{ $name }}:nodes-viewer
  labels:
    k8s-app: {{ $name }}
    k8s-addon: {{ $name }}.addons.k8s.io
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: kops:{{ $name }}:nodes-viewer
subjects:
- kind: ServiceAccount
  name: {{ $name }}
  namespace: {{ $namespace }}
---
kind: DaemonSet
apiVersion: extensions/v1beta1
metadata:
  name: {{ $name }}
  namespace: {{ $namespace }}
  labels:
    k8s-app: {{ $name }}
    k8s-addon: {{ $name }}.addons.k8s.io
spec:
  selector:
    matchLabels:
      k8s-app: {{ $name }}
  template:
    metadata:
      labels:
        k8s-app: {{ $name }}
      annotations:
        dns.alpha.kubernetes.io/internal: {{ $name }}-internal.{{ ClusterName }}
        prometheus.io/port: "{{ $na.Port }}"
        prometheus.io/scheme: "https"
        prometheus.io/scrape: "true"
        scheduler.alpha.kubernetes.io/critical-pod: ''
    spec:
      hostNetwork: true
      nodeSelector:
        kubernetes.io/role: master
      serviceAccount: {{ $name }}
      securityContext:
        fsGroup: 1000
      tolerations:
        - key: "node-role.kubernetes.io/master"
          effect: NoSchedule
      volumes:
        - name: config
          hostPath:
            path: /srv/kubernetes/node-authorizer
            type: DirectoryOrCreate
      containers:
        - name: {{ $name }}
          image: {{ $na.Image }}
          args:
            - server
            - --authorization-timeout={{ $na.Timeout.Duration }}
            - --authorizer={{ $na.Authorizer }}
            - --cluster-name={{ ClusterName }}
            {{- range $na.Features }}
            - --feature={{ . }}
            {{- end }}
            - --listen=0.0.0.0:{{ $na.Port }}
            - --tls-cert=/config/tls.pem
            - --tls-client-ca=/config/ca.pem
            - --tls-private-key=/config/tls-key.pem
            - --token-ttl={{ $na.TokenTTL.Duration }}
          {{- if $proxy }}
          env:
            - name: http_proxy
              value: {{ $proxy.HTTPProxy.Host }}:{{ $proxy.HTTPProxy.Port }}
            {{- if $proxy.ProxyExcludes }}
            - name: no_proxy
              value: {{ $proxy.ProxyExcludes }}
            {{- end }}
          {{- end }}
          resources:
            limits:
              cpu: 100m
              memory: 64Mi
            requests:
              cpu: 10m
              memory: 10Mi
          volumeMounts:
            - mountPath: /config
              readOnly: true
              name: config
`)

func cloudupResourcesAddonsNodeAuthorizerAddonsK8sIoK8s110YamlTemplateBytes() ([]byte, error) {
	return _cloudupResourcesAddonsNodeAuthorizerAddonsK8sIoK8s110YamlTemplate, nil
}

func cloudupResourcesAddonsNodeAuthorizerAddonsK8sIoK8s110YamlTemplate() (*asset, error) {
	bytes, err := cloudupResourcesAddonsNodeAuthorizerAddonsK8sIoK8s110YamlTemplateBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "cloudup/resources/addons/node-authorizer.addons.k8s.io/k8s-1.10.yaml.template", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _cloudupResourcesAddonsNodeAuthorizerAddonsK8sIoK8s112YamlTemplate = []byte(`{{- $proxy := .EgressProxy -}}
{{- $na := .NodeAuthorization.NodeAuthorizer -}}
{{- $name := "node-authorizer" -}}
{{- $namespace := "kube-system" -}}
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: {{ $name }}
  namespace: {{ $namespace }}
  labels:
    k8s-app: {{ $name }}
    k8s-addon: {{ $name }}.addons.k8s.io
---
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: kops:{{ $name }}:nodes-viewer
  labels:
    k8s-app: {{ $name }}
    k8s-addon: {{ $name }}.addons.k8s.io
rules:
- apiGroups:
  - "*"
  resources:
  - nodes
  verbs:
  - get
  - list
---
# permits the node access to create a CSR
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: kops:{{ $name }}:system:bootstrappers
  labels:
    k8s-app: {{ $name }}
    k8s-addon: {{ $name }}.addons.k8s.io
roleRef:
  kind: ClusterRole
  name: system:node-bootstrapper
  apiGroup: rbac.authorization.k8s.io
subjects:
- kind: Group
  name: system:bootstrappers
  apiGroup: rbac.authorization.k8s.io
---
# indicates to the controller to auto-sign the CSR for this group
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: kops:{{ $name }}:approval
  labels:
    k8s-app: {{ $name }}
    k8s-addon: {{ $name }}.addons.k8s.io
roleRef:
  kind: ClusterRole
  name: system:certificates.k8s.io:certificatesigningrequests:nodeclient
  apiGroup: rbac.authorization.k8s.io
subjects:
- kind: Group
  name: system:bootstrappers
  apiGroup: rbac.authorization.k8s.io
---
# the service permission requires to create the bootstrap tokens
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: kops:{{ $namespace }}:{{ $name }}
  namespace: {{ $namespace }}
  labels:
    k8s-app: {{ $name }}
    k8s-addon: {{ $name }}.addons.k8s.io
rules:
- apiGroups:
  - "*"
  resources:
  - secrets
  verbs:
  - create
  - list
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: kops:{{ $namespace }}:{{ $name }}
  namespace: {{ $namespace }}
  labels:
    k8s-app: {{ $name }}
    k8s-addon: {{ $name }}.addons.k8s.io
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: kops:{{ $namespace }}:{{ $name }}
subjects:
- kind: ServiceAccount
  name: {{ $name }}
  namespace: {{ $namespace }}
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: kops:{{ $name }}:nodes-viewer
  labels:
    k8s-app: {{ $name }}
    k8s-addon: {{ $name }}.addons.k8s.io
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: kops:{{ $name }}:nodes-viewer
subjects:
- kind: ServiceAccount
  name: {{ $name }}
  namespace: {{ $namespace }}
---
kind: DaemonSet
apiVersion: apps/v1
metadata:
  name: {{ $name }}
  namespace: {{ $namespace }}
  labels:
    k8s-app: {{ $name }}
    k8s-addon: {{ $name }}.addons.k8s.io
spec:
  selector:
    matchLabels:
      k8s-app: {{ $name }}
  template:
    metadata:
      labels:
        k8s-app: {{ $name }}
      annotations:
        dns.alpha.kubernetes.io/internal: {{ $name }}-internal.{{ ClusterName }}
        prometheus.io/port: "{{ $na.Port }}"
        prometheus.io/scheme: "https"
        prometheus.io/scrape: "true"
        scheduler.alpha.kubernetes.io/critical-pod: ''
    spec:
      hostNetwork: true
      nodeSelector:
        kubernetes.io/role: master
      priorityClassName: system-node-critical
      serviceAccount: {{ $name }}
      securityContext:
        fsGroup: 1000
      tolerations:
        - key: "node-role.kubernetes.io/master"
          effect: NoSchedule
      volumes:
        - name: config
          hostPath:
            path: /srv/kubernetes/node-authorizer
            type: DirectoryOrCreate
      containers:
        - name: {{ $name }}
          image: {{ $na.Image }}
          args:
            - server
            - --authorization-timeout={{ $na.Timeout.Duration }}
            - --authorizer={{ $na.Authorizer }}
            - --cluster-name={{ ClusterName }}
            {{- range $na.Features }}
            - --feature={{ . }}
            {{- end }}
            - --listen=0.0.0.0:{{ $na.Port }}
            - --tls-cert=/config/tls.pem
            - --tls-client-ca=/config/ca.pem
            - --tls-private-key=/config/tls-key.pem
            - --token-ttl={{ $na.TokenTTL.Duration }}
          {{- if $proxy }}
          env:
            - name: http_proxy
              value: {{ $proxy.HTTPProxy.Host }}:{{ $proxy.HTTPProxy.Port }}
            {{- if $proxy.ProxyExcludes }}
            - name: no_proxy
              value: {{ $proxy.ProxyExcludes }}
            {{- end }}
          {{- end }}
          resources:
            limits:
              cpu: 100m
              memory: 64Mi
            requests:
              cpu: 10m
              memory: 10Mi
          volumeMounts:
            - mountPath: /config
              readOnly: true
              name: config
`)

func cloudupResourcesAddonsNodeAuthorizerAddonsK8sIoK8s112YamlTemplateBytes() ([]byte, error) {
	return _cloudupResourcesAddonsNodeAuthorizerAddonsK8sIoK8s112YamlTemplate, nil
}

func cloudupResourcesAddonsNodeAuthorizerAddonsK8sIoK8s112YamlTemplate() (*asset, error) {
	bytes, err := cloudupResourcesAddonsNodeAuthorizerAddonsK8sIoK8s112YamlTemplateBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "cloudup/resources/addons/node-authorizer.addons.k8s.io/k8s-1.12.yaml.template", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _cloudupResourcesAddonsOpenstackAddonsK8sIoBuildBazel = []byte(`filegroup(
    name = "exported_testdata",
    srcs = glob(["**"]),
    visibility = ["//visibility:public"],
)
`)

func cloudupResourcesAddonsOpenstackAddonsK8sIoBuildBazelBytes() ([]byte, error) {
	return _cloudupResourcesAddonsOpenstackAddonsK8sIoBuildBazel, nil
}

func cloudupResourcesAddonsOpenstackAddonsK8sIoBuildBazel() (*asset, error) {
	bytes, err := cloudupResourcesAddonsOpenstackAddonsK8sIoBuildBazelBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "cloudup/resources/addons/openstack.addons.k8s.io/BUILD.bazel", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _cloudupResourcesAddonsOpenstackAddonsK8sIoK8s111YamlTemplate = []byte(`---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: cloud-controller-manager
  namespace: kube-system
  labels:
    k8s-addon: openstack.addons.k8s.io
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: system:cloud-node-controller
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: system:cloud-node-controller
subjects:
- kind: ServiceAccount
  name: cloud-node-controller
  namespace: kube-system
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: system:pvl-controller
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: system:pvl-controller
subjects:
- kind: ServiceAccount
  name: pvl-controller
  namespace: kube-system
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: system:cloud-controller-manager
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: system:cloud-controller-manager
subjects:
- kind: ServiceAccount
  name: cloud-controller-manager
  namespace: kube-system
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: system:cloud-controller-manager
rules:
- apiGroups:
  - ""
  resources:
  - events
  verbs:
  - create
  - patch
  - update
- apiGroups:
  - ""
  resources:
  - nodes
  verbs:
  - '*'
- apiGroups:
  - ""
  resources:
  - nodes/status
  verbs:
  - patch
- apiGroups:
  - ""
  resources:
  - services
  verbs:
  - list
  - patch
  - update
  - watch
- apiGroups:
  - ""
  resources:
  - serviceaccounts
  verbs:
  - create
  - get
- apiGroups:
  - ""
  resources:
  - persistentvolumes
  verbs:
  - '*'
- apiGroups:
  - ""
  resources:
  - endpoints
  verbs:
  - create
  - get
  - list
  - watch
  - update
- apiGroups:
  - ""
  resources:
  - configmaps
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - ""
  resources:
  - secrets
  verbs:
  - list
  - get
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: system:cloud-node-controller
rules:
- apiGroups:
  - ""
  resources:
  - nodes
  verbs:
  - '*'
- apiGroups:
  - ""
  resources:
  - nodes/status
  verbs:
  - patch
- apiGroups:
  - ""
  resources:
  - events
  verbs:
  - create
  - patch
  - update
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: system:pvl-controller
rules:
- apiGroups:
  - ""
  resources:
  - persistentvolumes
  verbs:
  - '*'
- apiGroups:
  - ""
  resources:
  - events
  verbs:
  - create
  - patch
  - update
---
apiVersion: apps/v1
kind: DaemonSet
metadata:
  namespace: kube-system
  name: openstack-cloud-provider
  labels:
    k8s-app: openstack-cloud-provider
    k8s-addon: openstack.addons.k8s.io
  annotations:
    scheduler.alpha.kubernetes.io/critical-pod: ""
spec:
  updateStrategy:
    type: RollingUpdate
  selector:
    matchLabels:
      name: openstack-cloud-provider
  template:
    metadata:
      labels:
        name: openstack-cloud-provider
    spec:
      # run on the host network (don't depend on CNI)
      hostNetwork: true
      # run on each master node
      nodeSelector:
        node-role.kubernetes.io/master: ""
      securityContext:
        runAsUser: 1001
      serviceAccountName: cloud-controller-manager
      tolerations:
      - effect: NoSchedule
        operator: Exists
      - key: CriticalAddonsOnly
        operator: Exists
      containers:
      - name: openstack-cloud-controller-manager
        image: "{{- .ExternalCloudControllerManager.Image }}"
        args:
          - /bin/openstack-cloud-controller-manager
{{- range $arg := CloudControllerConfigArgv }}
          - {{ $arg }}
{{- end }}
          - --cloud-config=/etc/kubernetes/cloud.config
          - --address=127.0.0.1
        volumeMounts:
        - mountPath: /etc/kubernetes/cloud.config
          name: cloudconfig
          readOnly: true
{{ if .UseHostCertificates }}
        - mountPath: /etc/ssl/certs
          name: etc-ssl-certs
          readOnly: true
{{ end }}
      volumes:
      - hostPath:
          path: /etc/kubernetes/cloud.config
        name: cloudconfig
{{ if .UseHostCertificates }}
      - hostPath:
          path: /etc/ssl/certs
          type: DirectoryOrCreate
        name: etc-ssl-certs
{{ end }}
`)

func cloudupResourcesAddonsOpenstackAddonsK8sIoK8s111YamlTemplateBytes() ([]byte, error) {
	return _cloudupResourcesAddonsOpenstackAddonsK8sIoK8s111YamlTemplate, nil
}

func cloudupResourcesAddonsOpenstackAddonsK8sIoK8s111YamlTemplate() (*asset, error) {
	bytes, err := cloudupResourcesAddonsOpenstackAddonsK8sIoK8s111YamlTemplateBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "cloudup/resources/addons/openstack.addons.k8s.io/k8s-1.11.yaml.template", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _cloudupResourcesAddonsOpenstackAddonsK8sIoK8s113YamlTemplate = []byte(`---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: cloud-controller-manager
  namespace: kube-system
  labels:
    k8s-app: openstack-cloud-provider
    k8s-addon: openstack.addons.k8s.io
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: system:cloud-node-controller
  labels:
    k8s-app: openstack-cloud-provider
    k8s-addon: openstack.addons.k8s.io
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: system:cloud-node-controller
subjects:
- kind: ServiceAccount
  name: cloud-node-controller
  namespace: kube-system
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: system:cloud-controller-manager
  labels:
    k8s-app: openstack-cloud-provider
    k8s-addon: openstack.addons.k8s.io
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: system:cloud-controller-manager
subjects:
- kind: ServiceAccount
  name: cloud-controller-manager
  namespace: kube-system
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: system:cloud-controller-manager
  labels:
    k8s-app: openstack-cloud-provider
    k8s-addon: openstack.addons.k8s.io
rules:
- apiGroups:
  - coordination.k8s.io
  resources:
  - leases
  verbs:
  - get
  - create
  - update
- apiGroups:
  - ""
  resources:
  - events
  verbs:
  - create
  - patch
  - update
- apiGroups:
  - ""
  resources:
  - nodes
  verbs:
  - '*'
- apiGroups:
  - ""
  resources:
  - nodes/status
  verbs:
  - patch
- apiGroups:
  - ""
  resources:
  - services
  verbs:
  - list
  - patch
  - update
  - watch
- apiGroups:
  - ""
  resources:
  - serviceaccounts
  verbs:
  - create
  - get
- apiGroups:
  - ""
  resources:
  - persistentvolumes
  verbs:
  - '*'
- apiGroups:
  - ""
  resources:
  - endpoints
  verbs:
  - create
  - get
  - list
  - watch
  - update
- apiGroups:
  - ""
  resources:
  - configmaps
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - ""
  resources:
  - secrets
  verbs:
  - list
  - get
  - watch
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: system:cloud-node-controller
  labels:
    k8s-app: openstack-cloud-provider
    k8s-addon: openstack.addons.k8s.io
rules:
- apiGroups:
  - ""
  resources:
  - nodes
  verbs:
  - '*'
- apiGroups:
  - ""
  resources:
  - nodes/status
  verbs:
  - patch
- apiGroups:
  - ""
  resources:
  - events
  verbs:
  - create
  - patch
  - update
---
apiVersion: apps/v1
kind: DaemonSet
metadata:
  namespace: kube-system
  name: openstack-cloud-provider
  labels:
    k8s-app: openstack-cloud-provider
    k8s-addon: openstack.addons.k8s.io
  annotations:
    scheduler.alpha.kubernetes.io/critical-pod: ""
spec:
  updateStrategy:
    type: RollingUpdate
  selector:
    matchLabels:
      name: openstack-cloud-provider
  template:
    metadata:
      labels:
        name: openstack-cloud-provider
    spec:
      # run on the host network (don't depend on CNI)
      hostNetwork: true
      # run on each master node
      nodeSelector:
        node-role.kubernetes.io/master: ""
      priorityClassName: system-node-critical
      securityContext:
        runAsUser: 1001
      serviceAccountName: cloud-controller-manager
      tolerations:
      - effect: NoSchedule
        operator: Exists
      - key: CriticalAddonsOnly
        operator: Exists
      containers:
      - name: openstack-cloud-controller-manager
        image: "{{- if .ExternalCloudControllerManager.Image -}} {{ .ExternalCloudControllerManager.Image }} {{- else -}} {{OpenStackCCM}} {{- end -}}"
        args:
          - /bin/openstack-cloud-controller-manager
{{- range $arg := CloudControllerConfigArgv }}
          - {{ $arg }}
{{- end }}
          - --cloud-config=/etc/kubernetes/cloud.config
          - --address=127.0.0.1
        resources:
          requests:
            cpu: 200m
        volumeMounts:
        - mountPath: /etc/kubernetes/cloud.config
          name: cloudconfig
          readOnly: true
{{ if .UseHostCertificates }}
        - mountPath: /etc/ssl/certs
          name: etc-ssl-certs
          readOnly: true
{{ end }}
      volumes:
      - hostPath:
          path: /etc/kubernetes/cloud.config
        name: cloudconfig
{{ if .UseHostCertificates }}
      - hostPath:
          path: /etc/ssl/certs
          type: DirectoryOrCreate
        name: etc-ssl-certs
{{ end }}
`)

func cloudupResourcesAddonsOpenstackAddonsK8sIoK8s113YamlTemplateBytes() ([]byte, error) {
	return _cloudupResourcesAddonsOpenstackAddonsK8sIoK8s113YamlTemplate, nil
}

func cloudupResourcesAddonsOpenstackAddonsK8sIoK8s113YamlTemplate() (*asset, error) {
	bytes, err := cloudupResourcesAddonsOpenstackAddonsK8sIoK8s113YamlTemplateBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "cloudup/resources/addons/openstack.addons.k8s.io/k8s-1.13.yaml.template", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _cloudupResourcesAddonsPodsecuritypolicyAddonsK8sIoK8s110YamlTemplate = []byte(`---
apiVersion: extensions/v1beta1
kind: PodSecurityPolicy
metadata:
  annotations:
    k8s-addon: podsecuritypolicy.addons.k8s.io
  name: kube-system
spec:
  allowedCapabilities:
  - '*'
  fsGroup:
    rule: RunAsAny
  hostPID: true
  hostIPC: true
  hostNetwork: true
  hostPorts:
  - min: 1
    max: 65536
  privileged: true
  runAsUser:
    rule: RunAsAny
  seLinux:
    rule: RunAsAny
  supplementalGroups:
    rule: RunAsAny
  volumes:
  - '*'
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  annotations:
    k8s-addon: podsecuritypolicy.addons.k8s.io
  name: kops:kube-system:psp
rules:
- apiGroups:
  - policy
  resources:
  - podsecuritypolicies
  resourceNames:
  - kube-system
  verbs:
  - use
---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: kops:kube-system:psp
roleRef:
  kind: ClusterRole
  name: kops:kube-system:psp
  apiGroup: rbac.authorization.k8s.io
subjects:
- kind: Group
  name: system:masters
  apiGroup: rbac.authorization.k8s.io
# permit the kubelets to access this policy (used for manifests)
- kind: User
  name: kubelet
  apiGroup: rbac.authorization.k8s.io
{{- if UseBootstrapTokens }}
- kind: Group
  name: system:nodes
  apiGroup: rbac.authorization.k8s.io
{{- end }}
---
kind: RoleBinding
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  annotations:
    k8s-addon: podsecuritypolicy.addons.k8s.io
  name: kops:kube-system:psp
  namespace: kube-system
roleRef:
  kind: ClusterRole
  name: kops:kube-system:psp
  apiGroup: rbac.authorization.k8s.io
subjects:
# permit the cluster wise admin to use this policy
- kind: Group
  name: system:serviceaccounts:kube-system
  apiGroup: rbac.authorization.k8s.io
`)

func cloudupResourcesAddonsPodsecuritypolicyAddonsK8sIoK8s110YamlTemplateBytes() ([]byte, error) {
	return _cloudupResourcesAddonsPodsecuritypolicyAddonsK8sIoK8s110YamlTemplate, nil
}

func cloudupResourcesAddonsPodsecuritypolicyAddonsK8sIoK8s110YamlTemplate() (*asset, error) {
	bytes, err := cloudupResourcesAddonsPodsecuritypolicyAddonsK8sIoK8s110YamlTemplateBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "cloudup/resources/addons/podsecuritypolicy.addons.k8s.io/k8s-1.10.yaml.template", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _cloudupResourcesAddonsPodsecuritypolicyAddonsK8sIoK8s112YamlTemplate = []byte(`---
apiVersion: policy/v1beta1
kind: PodSecurityPolicy
metadata:
  annotations:
    k8s-addon: podsecuritypolicy.addons.k8s.io
  name: kube-system
spec:
  allowedCapabilities:
  - '*'
  fsGroup:
    rule: RunAsAny
  hostPID: true
  hostIPC: true
  hostNetwork: true
  hostPorts:
  - min: 1
    max: 65536
  privileged: true
  runAsUser:
    rule: RunAsAny
  seLinux:
    rule: RunAsAny
  supplementalGroups:
    rule: RunAsAny
  volumes:
  - '*'
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  annotations:
    k8s-addon: podsecuritypolicy.addons.k8s.io
  name: kops:kube-system:psp
rules:
- apiGroups:
  - policy
  resources:
  - podsecuritypolicies
  resourceNames:
  - kube-system
  verbs:
  - use
---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: kops:kube-system:psp
roleRef:
  kind: ClusterRole
  name: kops:kube-system:psp
  apiGroup: rbac.authorization.k8s.io
subjects:
- kind: Group
  name: system:masters
  apiGroup: rbac.authorization.k8s.io
# permit the kubelets to access this policy (used for manifests)
- kind: User
  name: kubelet
  apiGroup: rbac.authorization.k8s.io
{{- if UseBootstrapTokens }}
- kind: Group
  name: system:nodes
  apiGroup: rbac.authorization.k8s.io
{{- end }}
---
kind: RoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  annotations:
    k8s-addon: podsecuritypolicy.addons.k8s.io
  name: kops:kube-system:psp
  namespace: kube-system
roleRef:
  kind: ClusterRole
  name: kops:kube-system:psp
  apiGroup: rbac.authorization.k8s.io
subjects:
# permit the cluster wise admin to use this policy
- kind: Group
  name: system:serviceaccounts:kube-system
  apiGroup: rbac.authorization.k8s.io
`)

func cloudupResourcesAddonsPodsecuritypolicyAddonsK8sIoK8s112YamlTemplateBytes() ([]byte, error) {
	return _cloudupResourcesAddonsPodsecuritypolicyAddonsK8sIoK8s112YamlTemplate, nil
}

func cloudupResourcesAddonsPodsecuritypolicyAddonsK8sIoK8s112YamlTemplate() (*asset, error) {
	bytes, err := cloudupResourcesAddonsPodsecuritypolicyAddonsK8sIoK8s112YamlTemplateBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "cloudup/resources/addons/podsecuritypolicy.addons.k8s.io/k8s-1.12.yaml.template", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _cloudupResourcesAddonsPodsecuritypolicyAddonsK8sIoK8s19YamlTemplate = []byte(`---
apiVersion: extensions/v1beta1
kind: PodSecurityPolicy
metadata:
  name: kube-system
spec:
  allowedCapabilities:
  - '*'
  fsGroup:
    rule: RunAsAny
  hostPID: true
  hostIPC: true
  hostNetwork: true
  hostPorts:
  - min: 1
    max: 65536
  privileged: true
  runAsUser:
    rule: RunAsAny
  seLinux:
    rule: RunAsAny
  supplementalGroups:
    rule: RunAsAny
  volumes:
  - '*'
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  name: kops:kube-system:psp
rules:
- apiGroups:
  - extensions
  resources:
  - podsecuritypolicies
  resourceNames:
  - kube-system
  verbs:
  - use
---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: kops:kube-system:psp
roleRef:
  kind: ClusterRole
  name: kops:kube-system:psp
  apiGroup: rbac.authorization.k8s.io
subjects:
# permit the cluster wise admin to use this policy
- kind: Group
  name: system:masters
  apiGroup: rbac.authorization.k8s.io
# permit the kubelets to access this policy (used for manifests)
- kind: User
  name: kubelet
  apiGroup: rbac.authorization.k8s.io
## TODO: need to question whether this can move into a rolebinding?
{{- if UseBootstrapTokens }}
- kind: Group
  name: system:nodes
  apiGroup: rbac.authorization.k8s.io
{{- end }}
---
kind: RoleBinding
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: kops:kube-system:psp
  namespace: kube-system
roleRef:
  kind: ClusterRole
  name: kops:kube-system:psp
  apiGroup: rbac.authorization.k8s.io
subjects:
# permit the cluster wise admin to use this policy
- kind: Group
  name: system:serviceaccounts:kube-system
  apiGroup: rbac.authorization.k8s.io
`)

func cloudupResourcesAddonsPodsecuritypolicyAddonsK8sIoK8s19YamlTemplateBytes() ([]byte, error) {
	return _cloudupResourcesAddonsPodsecuritypolicyAddonsK8sIoK8s19YamlTemplate, nil
}

func cloudupResourcesAddonsPodsecuritypolicyAddonsK8sIoK8s19YamlTemplate() (*asset, error) {
	bytes, err := cloudupResourcesAddonsPodsecuritypolicyAddonsK8sIoK8s19YamlTemplateBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "cloudup/resources/addons/podsecuritypolicy.addons.k8s.io/k8s-1.9.yaml.template", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _cloudupResourcesAddonsRbacAddonsK8sIoK8s18Yaml = []byte(`# Source: https://raw.githubusercontent.com/kubernetes/kubernetes/master/cluster/addons/rbac/kubelet-binding.yaml
# The GKE environments don't have kubelets with certificates that
# identify the system:nodes group.  They use the kubelet identity
# TODO: remove this once new nodes are granted individual identities and the
# NodeAuthorizer is enabled.
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: kubelet-cluster-admin
  labels:
    k8s-addon: rbac.addons.k8s.io
    kubernetes.io/cluster-service: "true"
    addonmanager.kubernetes.io/mode: Reconcile
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: system:node
subjects:
- apiGroup: rbac.authorization.k8s.io
  kind: User
  name: kubelet
`)

func cloudupResourcesAddonsRbacAddonsK8sIoK8s18YamlBytes() ([]byte, error) {
	return _cloudupResourcesAddonsRbacAddonsK8sIoK8s18Yaml, nil
}

func cloudupResourcesAddonsRbacAddonsK8sIoK8s18Yaml() (*asset, error) {
	bytes, err := cloudupResourcesAddonsRbacAddonsK8sIoK8s18YamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "cloudup/resources/addons/rbac.addons.k8s.io/k8s-1.8.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _cloudupResourcesAddonsSchedulerAddonsK8sIoV170Yaml = []byte(`kind: ConfigMap
apiVersion: v1
metadata:
  name: scheduler-policy
  namespace: kube-system
  labels:
    k8s-addon: scheduler.addons.k8s.io
data:
  policy.cfg: |
    {
      "kind" : "Policy",
      "apiVersion" : "v1",
      "predicates" : [
        {"name": "NoDiskConflict"},
        {"name": "NoVolumeZoneConflict"},
        {"name": "MaxEBSVolumeCount"},
        {"name": "MaxGCEPDVolumeCount"},
        {"name": "MaxAzureDiskVolumeCount"},
        {"name": "MatchInterPodAffinity"},
        {"name": "NoDiskConflict"},
        {"name": "GeneralPredicates"},
        {"name": "CheckNodeMemoryPressure"},
        {"name": "CheckNodeDiskPressure"},
        {"name": "CheckNodeCondition"},
        {"name": "PodToleratesNodeTaints"},
        {"name": "NoVolumeNodeConflict"}
      ],
      "priorities" : [
        {"name": "SelectorSpreadPriority", "weight" : 1},
        {"name": "LeastRequestedPriority", "weight" : 1},
        {"name": "BalancedResourceAllocation", "weight" : 1},
        {"name": "NodePreferAvoidPodsPriority", "weight" : 1},
        {"name": "NodeAffinityPriority", "weight" : 1},
        {"name": "TaintTolerationPriority", "weight" : 1},
        {"name": "InterPodAffinityPriority", "weight" : 1}
      ],
      "hardPodAffinitySymmetricWeight" : 1
    }`)

func cloudupResourcesAddonsSchedulerAddonsK8sIoV170YamlBytes() ([]byte, error) {
	return _cloudupResourcesAddonsSchedulerAddonsK8sIoV170Yaml, nil
}

func cloudupResourcesAddonsSchedulerAddonsK8sIoV170Yaml() (*asset, error) {
	bytes, err := cloudupResourcesAddonsSchedulerAddonsK8sIoV170YamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "cloudup/resources/addons/scheduler.addons.k8s.io/v1.7.0.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _cloudupResourcesAddonsSpotinstKubernetesClusterControllerAddonsK8sIoV1140YamlTemplate = []byte(`# ------------------------------------------------------------------------------
# Config Map
# ------------------------------------------------------------------------------
apiVersion: v1
kind: ConfigMap
metadata:
  name: spotinst-kubernetes-cluster-controller-config
  namespace: kube-system
data:
  spotinst.token: {{ SpotinstToken }}
  spotinst.account: {{ SpotinstAccount }}
  spotinst.cluster-identifier: {{ ClusterName }}
---
# ------------------------------------------------------------------------------
# Service Account
# ------------------------------------------------------------------------------
apiVersion: v1
kind: ServiceAccount
metadata:
  name: spotinst-kubernetes-cluster-controller
  namespace: kube-system
---
# ------------------------------------------------------------------------------
# Cluster Role
# ------------------------------------------------------------------------------
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: spotinst-kubernetes-cluster-controller
rules:
  # ----------------------------------------------------------------------------
  # Required for functional operation (read-only).
  # ----------------------------------------------------------------------------
- apiGroups: [""]
  resources: ["pods", "nodes", "services", "namespaces", "replicationcontrollers", "limitranges", "events", "persistentvolumes", "persistentvolumeclaims"]
  verbs: ["get", "list"]
- apiGroups: ["apps"]
  resources: ["deployments", "daemonsets", "statefulsets", "replicasets"]
  verbs: ["get","list"]
- apiGroups: ["storage.k8s.io"]
  resources: ["storageclasses"]
  verbs: ["get", "list"]
- apiGroups: ["batch"]
  resources: ["jobs"]
  verbs: ["get", "list"]
- apiGroups: ["extensions"]
  resources: ["replicasets", "daemonsets"]
  verbs: ["get","list"]
- apiGroups: ["policy"]
  resources: ["poddisruptionbudgets"]
  verbs: ["get", "list"]
- apiGroups: ["metrics.k8s.io"]
  resources: ["pods"]
  verbs: ["get", "list"]
- apiGroups: ["autoscaling"]
  resources: ["horizontalpodautoscalers"]
  verbs: ["get", "list"]
- nonResourceURLs: ["/version/", "/version"]
  verbs: ["get"]
  # ----------------------------------------------------------------------------
  # Required by the draining feature and for functional operation.
  # ----------------------------------------------------------------------------
- apiGroups: [""]
  resources: ["nodes"]
  verbs: ["patch", "update"]
- apiGroups: [""]
  resources: ["pods"]
  verbs: ["delete"]
- apiGroups: [""]
  resources: ["pods/eviction"]
  verbs: ["create"]
  # ----------------------------------------------------------------------------
  # Required by the Spotinst Cleanup feature.
  # ----------------------------------------------------------------------------
- apiGroups: [""]
  resources: ["nodes"]
  verbs: ["delete"]
  # ----------------------------------------------------------------------------
  # Required by the Spotinst CSR Approval feature.
  # ----------------------------------------------------------------------------
- apiGroups: ["certificates.k8s.io"]
  resources: ["certificatesigningrequests"]
  verbs: ["get", "list"]
- apiGroups: ["certificates.k8s.io"]
  resources: ["certificatesigningrequests/approval"]
  verbs: ["patch", "update"]
  # ----------------------------------------------------------------------------
  # Required by the Spotinst Auto Update feature.
  # ----------------------------------------------------------------------------
- apiGroups: ["rbac.authorization.k8s.io"]
  resources: ["clusterroles"]
  resourceNames: ["spotinst-kubernetes-cluster-controller"]
  verbs: ["patch", "update", "escalate"]
- apiGroups: ["apps"]
  resources: ["deployments"]
  resourceNames: ["spotinst-kubernetes-cluster-controller"]
  verbs: ["patch","update"]
  # ----------------------------------------------------------------------------
  # Required by the Spotinst Apply feature.
  # ----------------------------------------------------------------------------
- apiGroups: ["apps"]
  resources: ["deployments", "daemonsets"]
  verbs: ["get", "list", "patch","update","create","delete"]
- apiGroups: ["extensions"]
  resources: ["daemonsets"]
  verbs: ["get", "list", "patch","update","create","delete"]
- apiGroups: [""]
  resources: ["pods"]
  verbs: ["get", "list", "patch", "update", "create", "delete"]
---
# ------------------------------------------------------------------------------
# Cluster Role Binding
# ------------------------------------------------------------------------------
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: spotinst-kubernetes-cluster-controller
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: spotinst-kubernetes-cluster-controller
subjects:
- kind: ServiceAccount
  name: spotinst-kubernetes-cluster-controller
  namespace: kube-system
---
# ------------------------------------------------------------------------------
# Deployment
# ------------------------------------------------------------------------------
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    k8s-addon: spotinst-kubernetes-cluster-controller.addons.k8s.io
  name: spotinst-kubernetes-cluster-controller
  namespace: kube-system
spec:
  replicas: 1
  revisionHistoryLimit: 10
  selector:
    matchLabels:
      k8s-addon: spotinst-kubernetes-cluster-controller.addons.k8s.io
  template:
    metadata:
      labels:
        k8s-addon: spotinst-kubernetes-cluster-controller.addons.k8s.io
    spec:
      priorityClassName: system-cluster-critical
      affinity:
        nodeAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
          - weight: 100
            preference:
              matchExpressions:
              - key: node-role.kubernetes.io/master
                operator: Exists
        podAntiAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
          - weight: 50
            podAffinityTerm:
              labelSelector:
                matchExpressions:
                - key: k8s-addon
                  operator: In
                  values:
                  - spotinst-kubernetes-cluster-controller.addons.k8s.io
              topologyKey: kubernetes.io/hostname
      containers:
      - name: spotinst-kubernetes-cluster-controller
        imagePullPolicy: Always
        image: spotinst/kubernetes-cluster-controller:1.0.57
        livenessProbe:
          httpGet:
            path: /healthcheck
            port: 4401
          initialDelaySeconds: 300
          periodSeconds: 20
          timeoutSeconds: 2
          successThreshold: 1
          failureThreshold: 3
        env:
        - name: SPOTINST_TOKEN
          valueFrom:
            configMapKeyRef:
              name: spotinst-kubernetes-cluster-controller-config
              key: spotinst.token
        - name: SPOTINST_ACCOUNT
          valueFrom:
            configMapKeyRef:
              name: spotinst-kubernetes-cluster-controller-config
              key: spotinst.account
        - name: CLUSTER_IDENTIFIER
          valueFrom:
            configMapKeyRef:
              name: spotinst-kubernetes-cluster-controller-config
              key: spotinst.cluster-identifier
        - name: DISABLE_AUTO_UPDATE
          valueFrom:
            configMapKeyRef:
              name: spotinst-kubernetes-cluster-controller-config
              key: disable-auto-update
              optional: true
        - name: ENABLE_CSR_APPROVAL
          valueFrom:
            configMapKeyRef:
              name: spotinst-kubernetes-cluster-controller-config
              key: enable-csr-approval
              optional: true
        - name: PROXY_URL
          valueFrom:
            configMapKeyRef:
              name: spotinst-kubernetes-cluster-controller-config
              key: proxy-url
              optional: true
        - name: BASE_SPOTINST_URL
          valueFrom:
            configMapKeyRef:
              name: spotinst-kubernetes-cluster-controller-config
              key: base-url
              optional: true
        - name: POD_ID
          valueFrom:
            fieldRef:
              fieldPath: metadata.uid
        - name: POD_NAME
          valueFrom:
            fieldRef:
              fieldPath: metadata.name
        - name: POD_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
      serviceAccountName: spotinst-kubernetes-cluster-controller
      tolerations:
      - key: node.kubernetes.io/not-ready
        effect: NoExecute
        operator: Exists
        tolerationSeconds: 150
      - key: node.kubernetes.io/unreachable
        effect: NoExecute
        operator: Exists
        tolerationSeconds: 150
      - key: node-role.kubernetes.io/master
        operator: Exists
---
`)

func cloudupResourcesAddonsSpotinstKubernetesClusterControllerAddonsK8sIoV1140YamlTemplateBytes() ([]byte, error) {
	return _cloudupResourcesAddonsSpotinstKubernetesClusterControllerAddonsK8sIoV1140YamlTemplate, nil
}

func cloudupResourcesAddonsSpotinstKubernetesClusterControllerAddonsK8sIoV1140YamlTemplate() (*asset, error) {
	bytes, err := cloudupResourcesAddonsSpotinstKubernetesClusterControllerAddonsK8sIoV1140YamlTemplateBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "cloudup/resources/addons/spotinst-kubernetes-cluster-controller.addons.k8s.io/v1.14.0.yaml.template", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _cloudupResourcesAddonsSpotinstKubernetesClusterControllerAddonsK8sIoV190YamlTemplate = []byte(`# ------------------------------------------
# Config Map
# ------------------------------------------
apiVersion: v1
kind: ConfigMap
metadata:
  name: spotinst-kubernetes-cluster-controller-config
  namespace: kube-system
data:
  spotinst.token: {{ SpotinstToken }}
  spotinst.account: {{ SpotinstAccount }}
  spotinst.cluster-identifier: {{ ClusterName }}
---
# ------------------------------------------
# Secret
# ------------------------------------------
apiVersion: v1
kind: Secret
metadata:
  name: spotinst-kubernetes-cluster-controller-certs
  namespace: kube-system
type: Opaque
---
# ------------------------------------------
# Service Account
# ------------------------------------------
apiVersion: v1
kind: ServiceAccount
metadata:
  name: spotinst-kubernetes-cluster-controller
  namespace: kube-system
---
# ------------------------------------------
# Cluster Role
# ------------------------------------------
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: spotinst-kubernetes-cluster-controller
  namespace: kube-system
rules:
- apiGroups: [""]
  resources: ["pods", "nodes", "replicationcontrollers", "events", "limitranges", "services", "persistentvolumes", "persistentvolumeclaims", "namespaces"]
  verbs: ["get", "delete", "list", "patch", "update"]
- apiGroups: ["apps"]
  resources: ["deployments"]
  verbs: ["get","list","patch"]
- apiGroups: ["extensions"]
  resources: ["replicasets"]
  verbs: ["get","list"]
- apiGroups: ["rbac.authorization.k8s.io"]
  resources: ["clusterroles"]
  verbs: ["patch", "update", "escalate"]
- apiGroups: ["policy"]
  resources: ["poddisruptionbudgets"]
  verbs: ["list"]
- apiGroups: ["metrics.k8s.io"]
  resources: ["pods"]
  verbs: ["list"]
- nonResourceURLs: ["/version/", "/version"]
  verbs: ["get"]
---
# ------------------------------------------
# Cluster Role Binding
# ------------------------------------------
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: spotinst-kubernetes-cluster-controller
  namespace: kube-system
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: spotinst-kubernetes-cluster-controller
subjects:
- kind: ServiceAccount
  name: spotinst-kubernetes-cluster-controller
  namespace: kube-system
---
# ------------------------------------------
# Deployment
# ------------------------------------------
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    k8s-addon: spotinst-kubernetes-cluster-controller.addons.k8s.io
  name: spotinst-kubernetes-cluster-controller
  namespace: kube-system
spec:
  replicas: 1
  revisionHistoryLimit: 10
  selector:
    matchLabels:
      k8s-addon: spotinst-kubernetes-cluster-controller.addons.k8s.io
  template:
    metadata:
      labels:
        k8s-addon: spotinst-kubernetes-cluster-controller.addons.k8s.io
    spec:
      containers:
      - name: spotinst-kubernetes-cluster-controller
        imagePullPolicy: Always
        image: spotinst/kubernetes-cluster-controller:1.0.39
        volumeMounts:
        - name: spotinst-kubernetes-cluster-controller-certs
          mountPath: /certs
        livenessProbe:
          httpGet:
            path: /healthcheck
            port: 4401
          initialDelaySeconds: 300
          periodSeconds: 30
        env:
        - name: SPOTINST_TOKEN
          valueFrom:
            configMapKeyRef:
              name: spotinst-kubernetes-cluster-controller-config
              key: spotinst.token
        - name: SPOTINST_ACCOUNT
          valueFrom:
            configMapKeyRef:
              name: spotinst-kubernetes-cluster-controller-config
              key: spotinst.account
        - name: CLUSTER_IDENTIFIER
          valueFrom:
            configMapKeyRef:
              name: spotinst-kubernetes-cluster-controller-config
              key: spotinst.cluster-identifier
      volumes:
      - name: spotinst-kubernetes-cluster-controller-certs
        secret:
          secretName: spotinst-kubernetes-cluster-controller-certs
      serviceAccountName: spotinst-kubernetes-cluster-controller
      tolerations:
      - key: node-role.kubernetes.io/master
        effect: NoSchedule
---
`)

func cloudupResourcesAddonsSpotinstKubernetesClusterControllerAddonsK8sIoV190YamlTemplateBytes() ([]byte, error) {
	return _cloudupResourcesAddonsSpotinstKubernetesClusterControllerAddonsK8sIoV190YamlTemplate, nil
}

func cloudupResourcesAddonsSpotinstKubernetesClusterControllerAddonsK8sIoV190YamlTemplate() (*asset, error) {
	bytes, err := cloudupResourcesAddonsSpotinstKubernetesClusterControllerAddonsK8sIoV190YamlTemplateBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "cloudup/resources/addons/spotinst-kubernetes-cluster-controller.addons.k8s.io/v1.9.0.yaml.template", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _cloudupResourcesAddonsStorageAwsAddonsK8sIoV1150Yaml = []byte(`apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: default
  labels:
    k8s-addon: storage-aws.addons.k8s.io
provisioner: kubernetes.io/aws-ebs
parameters:
  type: gp2

---

apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: gp2
  labels:
    k8s-addon: storage-aws.addons.k8s.io
provisioner: kubernetes.io/aws-ebs
parameters:
  type: gp2

---

apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: kops-ssd-1-17
  annotations:
    storageclass.beta.kubernetes.io/is-default-class: "true"
  labels:
    k8s-addon: storage-aws.addons.k8s.io
provisioner: kubernetes.io/aws-ebs
parameters:
  type: gp2
  encrypted: "true"
allowVolumeExpansion: true
volumeBindingMode: WaitForFirstConsumer

---

apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  labels:
    k8s-addon: storage-aws.addons.k8s.io
  name: system:aws-cloud-provider
rules:
- apiGroups:
  - ""
  resources:
  - nodes
  verbs:
  - get
  - list
  - patch
- apiGroups:
  - ""
  resources:
  - events
  verbs:
  - create
  - patch
  - update

---

apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  labels:
    k8s-addon: storage-aws.addons.k8s.io
  name: system:aws-cloud-provider
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: system:aws-cloud-provider
subjects:
- kind: ServiceAccount
  name: aws-cloud-provider
  namespace: kube-system
`)

func cloudupResourcesAddonsStorageAwsAddonsK8sIoV1150YamlBytes() ([]byte, error) {
	return _cloudupResourcesAddonsStorageAwsAddonsK8sIoV1150Yaml, nil
}

func cloudupResourcesAddonsStorageAwsAddonsK8sIoV1150Yaml() (*asset, error) {
	bytes, err := cloudupResourcesAddonsStorageAwsAddonsK8sIoV1150YamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "cloudup/resources/addons/storage-aws.addons.k8s.io/v1.15.0.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _cloudupResourcesAddonsStorageAwsAddonsK8sIoV170Yaml = []byte(`apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: default
  labels:
    k8s-addon: storage-aws.addons.k8s.io
provisioner: kubernetes.io/aws-ebs
parameters:
  type: gp2

---

apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: gp2
  annotations:
    storageclass.beta.kubernetes.io/is-default-class: "true"
  labels:
    k8s-addon: storage-aws.addons.k8s.io
provisioner: kubernetes.io/aws-ebs
parameters:
  type: gp2
`)

func cloudupResourcesAddonsStorageAwsAddonsK8sIoV170YamlBytes() ([]byte, error) {
	return _cloudupResourcesAddonsStorageAwsAddonsK8sIoV170Yaml, nil
}

func cloudupResourcesAddonsStorageAwsAddonsK8sIoV170Yaml() (*asset, error) {
	bytes, err := cloudupResourcesAddonsStorageAwsAddonsK8sIoV170YamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "cloudup/resources/addons/storage-aws.addons.k8s.io/v1.7.0.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _cloudupResourcesAddonsStorageGceAddonsK8sIoV170Yaml = []byte(`apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: standard
  annotations:
    storageclass.beta.kubernetes.io/is-default-class: "true"
  labels:
    kubernetes.io/cluster-service: "true"
    k8s-addon: storage-gce.addons.k8s.io
    addonmanager.kubernetes.io/mode: EnsureExists
provisioner: kubernetes.io/gce-pd
parameters:
  type: pd-standard
`)

func cloudupResourcesAddonsStorageGceAddonsK8sIoV170YamlBytes() ([]byte, error) {
	return _cloudupResourcesAddonsStorageGceAddonsK8sIoV170Yaml, nil
}

func cloudupResourcesAddonsStorageGceAddonsK8sIoV170Yaml() (*asset, error) {
	bytes, err := cloudupResourcesAddonsStorageGceAddonsK8sIoV170YamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "cloudup/resources/addons/storage-gce.addons.k8s.io/v1.7.0.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _nodeup_automatic_upgrades_debian_familyFilesEtcAptAptConfD20autoUpgrades = []byte(`APT::Periodic::Update-Package-Lists "1";
APT::Periodic::Unattended-Upgrade "1";

APT::Periodic::AutocleanInterval "7";
`)

func nodeup_automatic_upgrades_debian_familyFilesEtcAptAptConfD20autoUpgradesBytes() ([]byte, error) {
	return _nodeup_automatic_upgrades_debian_familyFilesEtcAptAptConfD20autoUpgrades, nil
}

func nodeup_automatic_upgrades_debian_familyFilesEtcAptAptConfD20autoUpgrades() (*asset, error) {
	bytes, err := nodeup_automatic_upgrades_debian_familyFilesEtcAptAptConfD20autoUpgradesBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "nodeup/_automatic_upgrades/_debian_family/files/etc/apt/apt.conf.d/20auto-upgrades", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _nodeup_automatic_upgrades_debian_familyPackagesUnattendedUpgrades = []byte(``)

func nodeup_automatic_upgrades_debian_familyPackagesUnattendedUpgradesBytes() ([]byte, error) {
	return _nodeup_automatic_upgrades_debian_familyPackagesUnattendedUpgrades, nil
}

func nodeup_automatic_upgrades_debian_familyPackagesUnattendedUpgrades() (*asset, error) {
	bytes, err := nodeup_automatic_upgrades_debian_familyPackagesUnattendedUpgradesBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "nodeup/_automatic_upgrades/_debian_family/packages/unattended-upgrades", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _nodeupResources_lyft_vpc_cniFilesEtcCniNetD10CniIpvlanVpcK8sConflistTemplate = []byte(`{
  "cniVersion": "0.3.1",
  "name": "cni-ipvlan-vpc-k8s",
  "plugins": [
  {
    "cniVersion": "0.3.1",
    "type": "cni-ipvlan-vpc-k8s-ipam",
    "interfaceIndex": 1,
    "skipDeallocation": true,
    "subnetTags":  {{ SubnetTags }},
    "secGroupIds": {{ NodeSecurityGroups }}
  },
  {
    "cniVersion": "0.3.1",
    "type": "cni-ipvlan-vpc-k8s-ipvlan",
    "mode": "l2"
  },
  {
    "cniVersion": "0.3.1",
    "type": "cni-ipvlan-vpc-k8s-unnumbered-ptp",
    "hostInterface": "eth0",
    "containerInterface": "veth0",
    "ipMasq": true
  }
  ]
}
`)

func nodeupResources_lyft_vpc_cniFilesEtcCniNetD10CniIpvlanVpcK8sConflistTemplateBytes() ([]byte, error) {
	return _nodeupResources_lyft_vpc_cniFilesEtcCniNetD10CniIpvlanVpcK8sConflistTemplate, nil
}

func nodeupResources_lyft_vpc_cniFilesEtcCniNetD10CniIpvlanVpcK8sConflistTemplate() (*asset, error) {
	bytes, err := nodeupResources_lyft_vpc_cniFilesEtcCniNetD10CniIpvlanVpcK8sConflistTemplateBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "nodeup/resources/_lyft_vpc_cni/files/etc/cni/net.d/10-cni-ipvlan-vpc-k8s.conflist.template", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("Asset %s can't read by error: %v", name, err)
		}
		return a.bytes, nil
	}
	return nil, fmt.Errorf("Asset %s not found", name)
}

// MustAsset is like Asset but panics when Asset would return an error.
// It simplifies safe initialization of global variables.
func MustAsset(name string) []byte {
	a, err := Asset(name)
	if err != nil {
		panic("asset: Asset(" + name + "): " + err.Error())
	}

	return a
}

// AssetInfo loads and returns the asset info for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func AssetInfo(name string) (os.FileInfo, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("AssetInfo %s can't read by error: %v", name, err)
		}
		return a.info, nil
	}
	return nil, fmt.Errorf("AssetInfo %s not found", name)
}

// AssetNames returns the names of the assets.
func AssetNames() []string {
	names := make([]string, 0, len(_bindata))
	for name := range _bindata {
		names = append(names, name)
	}
	return names
}

// _bindata is a table, holding each asset generator, mapped to its name.
var _bindata = map[string]func() (*asset, error){
	"BUILD.bazel":                     buildBazel,
	"cloudup/resources/addons/OWNERS": cloudupResourcesAddonsOwners,
	"cloudup/resources/addons/authentication.aws/k8s-1.10.yaml.template":                                  cloudupResourcesAddonsAuthenticationAwsK8s110YamlTemplate,
	"cloudup/resources/addons/authentication.aws/k8s-1.12.yaml.template":                                  cloudupResourcesAddonsAuthenticationAwsK8s112YamlTemplate,
	"cloudup/resources/addons/authentication.kope.io/k8s-1.12.yaml":                                       cloudupResourcesAddonsAuthenticationKopeIoK8s112Yaml,
	"cloudup/resources/addons/authentication.kope.io/k8s-1.8.yaml":                                        cloudupResourcesAddonsAuthenticationKopeIoK8s18Yaml,
	"cloudup/resources/addons/core.addons.k8s.io/addon.yaml":                                              cloudupResourcesAddonsCoreAddonsK8sIoAddonYaml,
	"cloudup/resources/addons/core.addons.k8s.io/k8s-1.12.yaml.template":                                  cloudupResourcesAddonsCoreAddonsK8sIoK8s112YamlTemplate,
	"cloudup/resources/addons/core.addons.k8s.io/k8s-1.7.yaml.template":                                   cloudupResourcesAddonsCoreAddonsK8sIoK8s17YamlTemplate,
	"cloudup/resources/addons/core.addons.k8s.io/v1.4.0.yaml":                                             cloudupResourcesAddonsCoreAddonsK8sIoV140Yaml,
	"cloudup/resources/addons/coredns.addons.k8s.io/k8s-1.12.yaml.template":                               cloudupResourcesAddonsCorednsAddonsK8sIoK8s112YamlTemplate,
	"cloudup/resources/addons/coredns.addons.k8s.io/k8s-1.6.yaml.template":                                cloudupResourcesAddonsCorednsAddonsK8sIoK8s16YamlTemplate,
	"cloudup/resources/addons/digitalocean-cloud-controller.addons.k8s.io/k8s-1.8.yaml.template":          cloudupResourcesAddonsDigitaloceanCloudControllerAddonsK8sIoK8s18YamlTemplate,
	"cloudup/resources/addons/dns-controller.addons.k8s.io/k8s-1.12.yaml.template":                        cloudupResourcesAddonsDnsControllerAddonsK8sIoK8s112YamlTemplate,
	"cloudup/resources/addons/dns-controller.addons.k8s.io/k8s-1.6.yaml.template":                         cloudupResourcesAddonsDnsControllerAddonsK8sIoK8s16YamlTemplate,
	"cloudup/resources/addons/external-dns.addons.k8s.io/README.md":                                       cloudupResourcesAddonsExternalDnsAddonsK8sIoReadmeMd,
	"cloudup/resources/addons/external-dns.addons.k8s.io/k8s-1.12.yaml.template":                          cloudupResourcesAddonsExternalDnsAddonsK8sIoK8s112YamlTemplate,
	"cloudup/resources/addons/external-dns.addons.k8s.io/k8s-1.6.yaml.template":                           cloudupResourcesAddonsExternalDnsAddonsK8sIoK8s16YamlTemplate,
	"cloudup/resources/addons/kops-controller.addons.k8s.io/k8s-1.16.yaml.template":                       cloudupResourcesAddonsKopsControllerAddonsK8sIoK8s116YamlTemplate,
	"cloudup/resources/addons/kube-dns.addons.k8s.io/k8s-1.12.yaml.template":                              cloudupResourcesAddonsKubeDnsAddonsK8sIoK8s112YamlTemplate,
	"cloudup/resources/addons/kube-dns.addons.k8s.io/k8s-1.6.yaml.template":                               cloudupResourcesAddonsKubeDnsAddonsK8sIoK8s16YamlTemplate,
	"cloudup/resources/addons/kubelet-api.rbac.addons.k8s.io/k8s-1.9.yaml":                                cloudupResourcesAddonsKubeletApiRbacAddonsK8sIoK8s19Yaml,
	"cloudup/resources/addons/limit-range.addons.k8s.io/addon.yaml":                                       cloudupResourcesAddonsLimitRangeAddonsK8sIoAddonYaml,
	"cloudup/resources/addons/limit-range.addons.k8s.io/v1.5.0.yaml":                                      cloudupResourcesAddonsLimitRangeAddonsK8sIoV150Yaml,
	"cloudup/resources/addons/metadata-proxy.addons.k8s.io/addon.yaml":                                    cloudupResourcesAddonsMetadataProxyAddonsK8sIoAddonYaml,
	"cloudup/resources/addons/metadata-proxy.addons.k8s.io/v0.1.12.yaml":                                  cloudupResourcesAddonsMetadataProxyAddonsK8sIoV0112Yaml,
	"cloudup/resources/addons/networking.amazon-vpc-routed-eni/k8s-1.10.yaml.template":                    cloudupResourcesAddonsNetworkingAmazonVpcRoutedEniK8s110YamlTemplate,
	"cloudup/resources/addons/networking.amazon-vpc-routed-eni/k8s-1.12.yaml.template":                    cloudupResourcesAddonsNetworkingAmazonVpcRoutedEniK8s112YamlTemplate,
	"cloudup/resources/addons/networking.amazon-vpc-routed-eni/k8s-1.16.yaml.template":                    cloudupResourcesAddonsNetworkingAmazonVpcRoutedEniK8s116YamlTemplate,
	"cloudup/resources/addons/networking.amazon-vpc-routed-eni/k8s-1.8.yaml.template":                     cloudupResourcesAddonsNetworkingAmazonVpcRoutedEniK8s18YamlTemplate,
	"cloudup/resources/addons/networking.cilium.io/k8s-1.12.yaml.template":                                cloudupResourcesAddonsNetworkingCiliumIoK8s112YamlTemplate,
	"cloudup/resources/addons/networking.cilium.io/k8s-1.7.yaml.template":                                 cloudupResourcesAddonsNetworkingCiliumIoK8s17YamlTemplate,
	"cloudup/resources/addons/networking.flannel/k8s-1.12.yaml.template":                                  cloudupResourcesAddonsNetworkingFlannelK8s112YamlTemplate,
	"cloudup/resources/addons/networking.flannel/k8s-1.6.yaml.template":                                   cloudupResourcesAddonsNetworkingFlannelK8s16YamlTemplate,
	"cloudup/resources/addons/networking.kope.io/k8s-1.12.yaml":                                           cloudupResourcesAddonsNetworkingKopeIoK8s112Yaml,
	"cloudup/resources/addons/networking.kope.io/k8s-1.6.yaml":                                            cloudupResourcesAddonsNetworkingKopeIoK8s16Yaml,
	"cloudup/resources/addons/networking.kuberouter/k8s-1.12.yaml.template":                               cloudupResourcesAddonsNetworkingKuberouterK8s112YamlTemplate,
	"cloudup/resources/addons/networking.kuberouter/k8s-1.6.yaml.template":                                cloudupResourcesAddonsNetworkingKuberouterK8s16YamlTemplate,
	"cloudup/resources/addons/networking.projectcalico.org/k8s-1.12.yaml.template":                        cloudupResourcesAddonsNetworkingProjectcalicoOrgK8s112YamlTemplate,
	"cloudup/resources/addons/networking.projectcalico.org/k8s-1.16.yaml.template":                        cloudupResourcesAddonsNetworkingProjectcalicoOrgK8s116YamlTemplate,
	"cloudup/resources/addons/networking.projectcalico.org/k8s-1.7-v3.yaml.template":                      cloudupResourcesAddonsNetworkingProjectcalicoOrgK8s17V3YamlTemplate,
	"cloudup/resources/addons/networking.projectcalico.org/k8s-1.7.yaml.template":                         cloudupResourcesAddonsNetworkingProjectcalicoOrgK8s17YamlTemplate,
	"cloudup/resources/addons/networking.projectcalico.org.canal/k8s-1.12.yaml.template":                  cloudupResourcesAddonsNetworkingProjectcalicoOrgCanalK8s112YamlTemplate,
	"cloudup/resources/addons/networking.projectcalico.org.canal/k8s-1.15.yaml.template":                  cloudupResourcesAddonsNetworkingProjectcalicoOrgCanalK8s115YamlTemplate,
	"cloudup/resources/addons/networking.projectcalico.org.canal/k8s-1.16.yaml.template":                  cloudupResourcesAddonsNetworkingProjectcalicoOrgCanalK8s116YamlTemplate,
	"cloudup/resources/addons/networking.projectcalico.org.canal/k8s-1.9.yaml.template":                   cloudupResourcesAddonsNetworkingProjectcalicoOrgCanalK8s19YamlTemplate,
	"cloudup/resources/addons/networking.romana/k8s-1.12.yaml.template":                                   cloudupResourcesAddonsNetworkingRomanaK8s112YamlTemplate,
	"cloudup/resources/addons/networking.romana/k8s-1.7.yaml.template":                                    cloudupResourcesAddonsNetworkingRomanaK8s17YamlTemplate,
	"cloudup/resources/addons/networking.weave/k8s-1.12.yaml.template":                                    cloudupResourcesAddonsNetworkingWeaveK8s112YamlTemplate,
	"cloudup/resources/addons/networking.weave/k8s-1.8.yaml.template":                                     cloudupResourcesAddonsNetworkingWeaveK8s18YamlTemplate,
	"cloudup/resources/addons/node-authorizer.addons.k8s.io/k8s-1.10.yaml.template":                       cloudupResourcesAddonsNodeAuthorizerAddonsK8sIoK8s110YamlTemplate,
	"cloudup/resources/addons/node-authorizer.addons.k8s.io/k8s-1.12.yaml.template":                       cloudupResourcesAddonsNodeAuthorizerAddonsK8sIoK8s112YamlTemplate,
	"cloudup/resources/addons/openstack.addons.k8s.io/BUILD.bazel":                                        cloudupResourcesAddonsOpenstackAddonsK8sIoBuildBazel,
	"cloudup/resources/addons/openstack.addons.k8s.io/k8s-1.11.yaml.template":                             cloudupResourcesAddonsOpenstackAddonsK8sIoK8s111YamlTemplate,
	"cloudup/resources/addons/openstack.addons.k8s.io/k8s-1.13.yaml.template":                             cloudupResourcesAddonsOpenstackAddonsK8sIoK8s113YamlTemplate,
	"cloudup/resources/addons/podsecuritypolicy.addons.k8s.io/k8s-1.10.yaml.template":                     cloudupResourcesAddonsPodsecuritypolicyAddonsK8sIoK8s110YamlTemplate,
	"cloudup/resources/addons/podsecuritypolicy.addons.k8s.io/k8s-1.12.yaml.template":                     cloudupResourcesAddonsPodsecuritypolicyAddonsK8sIoK8s112YamlTemplate,
	"cloudup/resources/addons/podsecuritypolicy.addons.k8s.io/k8s-1.9.yaml.template":                      cloudupResourcesAddonsPodsecuritypolicyAddonsK8sIoK8s19YamlTemplate,
	"cloudup/resources/addons/rbac.addons.k8s.io/k8s-1.8.yaml":                                            cloudupResourcesAddonsRbacAddonsK8sIoK8s18Yaml,
	"cloudup/resources/addons/scheduler.addons.k8s.io/v1.7.0.yaml":                                        cloudupResourcesAddonsSchedulerAddonsK8sIoV170Yaml,
	"cloudup/resources/addons/spotinst-kubernetes-cluster-controller.addons.k8s.io/v1.14.0.yaml.template": cloudupResourcesAddonsSpotinstKubernetesClusterControllerAddonsK8sIoV1140YamlTemplate,
	"cloudup/resources/addons/spotinst-kubernetes-cluster-controller.addons.k8s.io/v1.9.0.yaml.template":  cloudupResourcesAddonsSpotinstKubernetesClusterControllerAddonsK8sIoV190YamlTemplate,
	"cloudup/resources/addons/storage-aws.addons.k8s.io/v1.15.0.yaml":                                     cloudupResourcesAddonsStorageAwsAddonsK8sIoV1150Yaml,
	"cloudup/resources/addons/storage-aws.addons.k8s.io/v1.7.0.yaml":                                      cloudupResourcesAddonsStorageAwsAddonsK8sIoV170Yaml,
	"cloudup/resources/addons/storage-gce.addons.k8s.io/v1.7.0.yaml":                                      cloudupResourcesAddonsStorageGceAddonsK8sIoV170Yaml,
	"nodeup/_automatic_upgrades/_debian_family/files/etc/apt/apt.conf.d/20auto-upgrades":                  nodeup_automatic_upgrades_debian_familyFilesEtcAptAptConfD20autoUpgrades,
	"nodeup/_automatic_upgrades/_debian_family/packages/unattended-upgrades":                              nodeup_automatic_upgrades_debian_familyPackagesUnattendedUpgrades,
	"nodeup/resources/_lyft_vpc_cni/files/etc/cni/net.d/10-cni-ipvlan-vpc-k8s.conflist.template":          nodeupResources_lyft_vpc_cniFilesEtcCniNetD10CniIpvlanVpcK8sConflistTemplate,
}

// AssetDir returns the file names below a certain
// directory embedded in the file by go-bindata.
// For example if you run go-bindata on data/... and data contains the
// following hierarchy:
//     data/
//       foo.txt
//       img/
//         a.png
//         b.png
// then AssetDir("data") would return []string{"foo.txt", "img"}
// AssetDir("data/img") would return []string{"a.png", "b.png"}
// AssetDir("foo.txt") and AssetDir("notexist") would return an error
// AssetDir("") will return []string{"data"}.
func AssetDir(name string) ([]string, error) {
	node := _bintree
	if len(name) != 0 {
		cannonicalName := strings.Replace(name, "\\", "/", -1)
		pathList := strings.Split(cannonicalName, "/")
		for _, p := range pathList {
			node = node.Children[p]
			if node == nil {
				return nil, fmt.Errorf("Asset %s not found", name)
			}
		}
	}
	if node.Func != nil {
		return nil, fmt.Errorf("Asset %s not found", name)
	}
	rv := make([]string, 0, len(node.Children))
	for childName := range node.Children {
		rv = append(rv, childName)
	}
	return rv, nil
}

type bintree struct {
	Func     func() (*asset, error)
	Children map[string]*bintree
}

var _bintree = &bintree{nil, map[string]*bintree{
	"BUILD.bazel": {buildBazel, map[string]*bintree{}},
	"cloudup": {nil, map[string]*bintree{
		"resources": {nil, map[string]*bintree{
			"addons": {nil, map[string]*bintree{
				"OWNERS": {cloudupResourcesAddonsOwners, map[string]*bintree{}},
				"authentication.aws": {nil, map[string]*bintree{
					"k8s-1.10.yaml.template": {cloudupResourcesAddonsAuthenticationAwsK8s110YamlTemplate, map[string]*bintree{}},
					"k8s-1.12.yaml.template": {cloudupResourcesAddonsAuthenticationAwsK8s112YamlTemplate, map[string]*bintree{}},
				}},
				"authentication.kope.io": {nil, map[string]*bintree{
					"k8s-1.12.yaml": {cloudupResourcesAddonsAuthenticationKopeIoK8s112Yaml, map[string]*bintree{}},
					"k8s-1.8.yaml":  {cloudupResourcesAddonsAuthenticationKopeIoK8s18Yaml, map[string]*bintree{}},
				}},
				"core.addons.k8s.io": {nil, map[string]*bintree{
					"addon.yaml":             {cloudupResourcesAddonsCoreAddonsK8sIoAddonYaml, map[string]*bintree{}},
					"k8s-1.12.yaml.template": {cloudupResourcesAddonsCoreAddonsK8sIoK8s112YamlTemplate, map[string]*bintree{}},
					"k8s-1.7.yaml.template":  {cloudupResourcesAddonsCoreAddonsK8sIoK8s17YamlTemplate, map[string]*bintree{}},
					"v1.4.0.yaml":            {cloudupResourcesAddonsCoreAddonsK8sIoV140Yaml, map[string]*bintree{}},
				}},
				"coredns.addons.k8s.io": {nil, map[string]*bintree{
					"k8s-1.12.yaml.template": {cloudupResourcesAddonsCorednsAddonsK8sIoK8s112YamlTemplate, map[string]*bintree{}},
					"k8s-1.6.yaml.template":  {cloudupResourcesAddonsCorednsAddonsK8sIoK8s16YamlTemplate, map[string]*bintree{}},
				}},
				"digitalocean-cloud-controller.addons.k8s.io": {nil, map[string]*bintree{
					"k8s-1.8.yaml.template": {cloudupResourcesAddonsDigitaloceanCloudControllerAddonsK8sIoK8s18YamlTemplate, map[string]*bintree{}},
				}},
				"dns-controller.addons.k8s.io": {nil, map[string]*bintree{
					"k8s-1.12.yaml.template": {cloudupResourcesAddonsDnsControllerAddonsK8sIoK8s112YamlTemplate, map[string]*bintree{}},
					"k8s-1.6.yaml.template":  {cloudupResourcesAddonsDnsControllerAddonsK8sIoK8s16YamlTemplate, map[string]*bintree{}},
				}},
				"external-dns.addons.k8s.io": {nil, map[string]*bintree{
					"README.md":              {cloudupResourcesAddonsExternalDnsAddonsK8sIoReadmeMd, map[string]*bintree{}},
					"k8s-1.12.yaml.template": {cloudupResourcesAddonsExternalDnsAddonsK8sIoK8s112YamlTemplate, map[string]*bintree{}},
					"k8s-1.6.yaml.template":  {cloudupResourcesAddonsExternalDnsAddonsK8sIoK8s16YamlTemplate, map[string]*bintree{}},
				}},
				"kops-controller.addons.k8s.io": {nil, map[string]*bintree{
					"k8s-1.16.yaml.template": {cloudupResourcesAddonsKopsControllerAddonsK8sIoK8s116YamlTemplate, map[string]*bintree{}},
				}},
				"kube-dns.addons.k8s.io": {nil, map[string]*bintree{
					"k8s-1.12.yaml.template": {cloudupResourcesAddonsKubeDnsAddonsK8sIoK8s112YamlTemplate, map[string]*bintree{}},
					"k8s-1.6.yaml.template":  {cloudupResourcesAddonsKubeDnsAddonsK8sIoK8s16YamlTemplate, map[string]*bintree{}},
				}},
				"kubelet-api.rbac.addons.k8s.io": {nil, map[string]*bintree{
					"k8s-1.9.yaml": {cloudupResourcesAddonsKubeletApiRbacAddonsK8sIoK8s19Yaml, map[string]*bintree{}},
				}},
				"limit-range.addons.k8s.io": {nil, map[string]*bintree{
					"addon.yaml":  {cloudupResourcesAddonsLimitRangeAddonsK8sIoAddonYaml, map[string]*bintree{}},
					"v1.5.0.yaml": {cloudupResourcesAddonsLimitRangeAddonsK8sIoV150Yaml, map[string]*bintree{}},
				}},
				"metadata-proxy.addons.k8s.io": {nil, map[string]*bintree{
					"addon.yaml":   {cloudupResourcesAddonsMetadataProxyAddonsK8sIoAddonYaml, map[string]*bintree{}},
					"v0.1.12.yaml": {cloudupResourcesAddonsMetadataProxyAddonsK8sIoV0112Yaml, map[string]*bintree{}},
				}},
				"networking.amazon-vpc-routed-eni": {nil, map[string]*bintree{
					"k8s-1.10.yaml.template": {cloudupResourcesAddonsNetworkingAmazonVpcRoutedEniK8s110YamlTemplate, map[string]*bintree{}},
					"k8s-1.12.yaml.template": {cloudupResourcesAddonsNetworkingAmazonVpcRoutedEniK8s112YamlTemplate, map[string]*bintree{}},
					"k8s-1.16.yaml.template": {cloudupResourcesAddonsNetworkingAmazonVpcRoutedEniK8s116YamlTemplate, map[string]*bintree{}},
					"k8s-1.8.yaml.template":  {cloudupResourcesAddonsNetworkingAmazonVpcRoutedEniK8s18YamlTemplate, map[string]*bintree{}},
				}},
				"networking.cilium.io": {nil, map[string]*bintree{
					"k8s-1.12.yaml.template": {cloudupResourcesAddonsNetworkingCiliumIoK8s112YamlTemplate, map[string]*bintree{}},
					"k8s-1.7.yaml.template":  {cloudupResourcesAddonsNetworkingCiliumIoK8s17YamlTemplate, map[string]*bintree{}},
				}},
				"networking.flannel": {nil, map[string]*bintree{
					"k8s-1.12.yaml.template": {cloudupResourcesAddonsNetworkingFlannelK8s112YamlTemplate, map[string]*bintree{}},
					"k8s-1.6.yaml.template":  {cloudupResourcesAddonsNetworkingFlannelK8s16YamlTemplate, map[string]*bintree{}},
				}},
				"networking.kope.io": {nil, map[string]*bintree{
					"k8s-1.12.yaml": {cloudupResourcesAddonsNetworkingKopeIoK8s112Yaml, map[string]*bintree{}},
					"k8s-1.6.yaml":  {cloudupResourcesAddonsNetworkingKopeIoK8s16Yaml, map[string]*bintree{}},
				}},
				"networking.kuberouter": {nil, map[string]*bintree{
					"k8s-1.12.yaml.template": {cloudupResourcesAddonsNetworkingKuberouterK8s112YamlTemplate, map[string]*bintree{}},
					"k8s-1.6.yaml.template":  {cloudupResourcesAddonsNetworkingKuberouterK8s16YamlTemplate, map[string]*bintree{}},
				}},
				"networking.projectcalico.org": {nil, map[string]*bintree{
					"k8s-1.12.yaml.template":   {cloudupResourcesAddonsNetworkingProjectcalicoOrgK8s112YamlTemplate, map[string]*bintree{}},
					"k8s-1.16.yaml.template":   {cloudupResourcesAddonsNetworkingProjectcalicoOrgK8s116YamlTemplate, map[string]*bintree{}},
					"k8s-1.7-v3.yaml.template": {cloudupResourcesAddonsNetworkingProjectcalicoOrgK8s17V3YamlTemplate, map[string]*bintree{}},
					"k8s-1.7.yaml.template":    {cloudupResourcesAddonsNetworkingProjectcalicoOrgK8s17YamlTemplate, map[string]*bintree{}},
				}},
				"networking.projectcalico.org.canal": {nil, map[string]*bintree{
					"k8s-1.12.yaml.template": {cloudupResourcesAddonsNetworkingProjectcalicoOrgCanalK8s112YamlTemplate, map[string]*bintree{}},
					"k8s-1.15.yaml.template": {cloudupResourcesAddonsNetworkingProjectcalicoOrgCanalK8s115YamlTemplate, map[string]*bintree{}},
					"k8s-1.16.yaml.template": {cloudupResourcesAddonsNetworkingProjectcalicoOrgCanalK8s116YamlTemplate, map[string]*bintree{}},
					"k8s-1.9.yaml.template":  {cloudupResourcesAddonsNetworkingProjectcalicoOrgCanalK8s19YamlTemplate, map[string]*bintree{}},
				}},
				"networking.romana": {nil, map[string]*bintree{
					"k8s-1.12.yaml.template": {cloudupResourcesAddonsNetworkingRomanaK8s112YamlTemplate, map[string]*bintree{}},
					"k8s-1.7.yaml.template":  {cloudupResourcesAddonsNetworkingRomanaK8s17YamlTemplate, map[string]*bintree{}},
				}},
				"networking.weave": {nil, map[string]*bintree{
					"k8s-1.12.yaml.template": {cloudupResourcesAddonsNetworkingWeaveK8s112YamlTemplate, map[string]*bintree{}},
					"k8s-1.8.yaml.template":  {cloudupResourcesAddonsNetworkingWeaveK8s18YamlTemplate, map[string]*bintree{}},
				}},
				"node-authorizer.addons.k8s.io": {nil, map[string]*bintree{
					"k8s-1.10.yaml.template": {cloudupResourcesAddonsNodeAuthorizerAddonsK8sIoK8s110YamlTemplate, map[string]*bintree{}},
					"k8s-1.12.yaml.template": {cloudupResourcesAddonsNodeAuthorizerAddonsK8sIoK8s112YamlTemplate, map[string]*bintree{}},
				}},
				"openstack.addons.k8s.io": {nil, map[string]*bintree{
					"BUILD.bazel":            {cloudupResourcesAddonsOpenstackAddonsK8sIoBuildBazel, map[string]*bintree{}},
					"k8s-1.11.yaml.template": {cloudupResourcesAddonsOpenstackAddonsK8sIoK8s111YamlTemplate, map[string]*bintree{}},
					"k8s-1.13.yaml.template": {cloudupResourcesAddonsOpenstackAddonsK8sIoK8s113YamlTemplate, map[string]*bintree{}},
				}},
				"podsecuritypolicy.addons.k8s.io": {nil, map[string]*bintree{
					"k8s-1.10.yaml.template": {cloudupResourcesAddonsPodsecuritypolicyAddonsK8sIoK8s110YamlTemplate, map[string]*bintree{}},
					"k8s-1.12.yaml.template": {cloudupResourcesAddonsPodsecuritypolicyAddonsK8sIoK8s112YamlTemplate, map[string]*bintree{}},
					"k8s-1.9.yaml.template":  {cloudupResourcesAddonsPodsecuritypolicyAddonsK8sIoK8s19YamlTemplate, map[string]*bintree{}},
				}},
				"rbac.addons.k8s.io": {nil, map[string]*bintree{
					"k8s-1.8.yaml": {cloudupResourcesAddonsRbacAddonsK8sIoK8s18Yaml, map[string]*bintree{}},
				}},
				"scheduler.addons.k8s.io": {nil, map[string]*bintree{
					"v1.7.0.yaml": {cloudupResourcesAddonsSchedulerAddonsK8sIoV170Yaml, map[string]*bintree{}},
				}},
				"spotinst-kubernetes-cluster-controller.addons.k8s.io": {nil, map[string]*bintree{
					"v1.14.0.yaml.template": {cloudupResourcesAddonsSpotinstKubernetesClusterControllerAddonsK8sIoV1140YamlTemplate, map[string]*bintree{}},
					"v1.9.0.yaml.template":  {cloudupResourcesAddonsSpotinstKubernetesClusterControllerAddonsK8sIoV190YamlTemplate, map[string]*bintree{}},
				}},
				"storage-aws.addons.k8s.io": {nil, map[string]*bintree{
					"v1.15.0.yaml": {cloudupResourcesAddonsStorageAwsAddonsK8sIoV1150Yaml, map[string]*bintree{}},
					"v1.7.0.yaml":  {cloudupResourcesAddonsStorageAwsAddonsK8sIoV170Yaml, map[string]*bintree{}},
				}},
				"storage-gce.addons.k8s.io": {nil, map[string]*bintree{
					"v1.7.0.yaml": {cloudupResourcesAddonsStorageGceAddonsK8sIoV170Yaml, map[string]*bintree{}},
				}},
			}},
		}},
	}},
	"nodeup": {nil, map[string]*bintree{
		"_automatic_upgrades": {nil, map[string]*bintree{
			"_debian_family": {nil, map[string]*bintree{
				"files": {nil, map[string]*bintree{
					"etc": {nil, map[string]*bintree{
						"apt": {nil, map[string]*bintree{
							"apt.conf.d": {nil, map[string]*bintree{
								"20auto-upgrades": {nodeup_automatic_upgrades_debian_familyFilesEtcAptAptConfD20autoUpgrades, map[string]*bintree{}},
							}},
						}},
					}},
				}},
				"packages": {nil, map[string]*bintree{
					"unattended-upgrades": {nodeup_automatic_upgrades_debian_familyPackagesUnattendedUpgrades, map[string]*bintree{}},
				}},
			}},
		}},
		"resources": {nil, map[string]*bintree{
			"_lyft_vpc_cni": {nil, map[string]*bintree{
				"files": {nil, map[string]*bintree{
					"etc": {nil, map[string]*bintree{
						"cni": {nil, map[string]*bintree{
							"net.d": {nil, map[string]*bintree{
								"10-cni-ipvlan-vpc-k8s.conflist.template": {nodeupResources_lyft_vpc_cniFilesEtcCniNetD10CniIpvlanVpcK8sConflistTemplate, map[string]*bintree{}},
							}},
						}},
					}},
				}},
			}},
		}},
	}},
}}

// RestoreAsset restores an asset under the given directory
func RestoreAsset(dir, name string) error {
	data, err := Asset(name)
	if err != nil {
		return err
	}
	info, err := AssetInfo(name)
	if err != nil {
		return err
	}
	err = os.MkdirAll(_filePath(dir, filepath.Dir(name)), os.FileMode(0755))
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(_filePath(dir, name), data, info.Mode())
	if err != nil {
		return err
	}
	err = os.Chtimes(_filePath(dir, name), info.ModTime(), info.ModTime())
	if err != nil {
		return err
	}
	return nil
}

// RestoreAssets restores an asset under the given directory recursively
func RestoreAssets(dir, name string) error {
	children, err := AssetDir(name)
	// File
	if err != nil {
		return RestoreAsset(dir, name)
	}
	// Dir
	for _, child := range children {
		err = RestoreAssets(dir, filepath.Join(name, child))
		if err != nil {
			return err
		}
	}
	return nil
}

func _filePath(dir, name string) string {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	return filepath.Join(append([]string{dir}, strings.Split(cannonicalName, "/")...)...)
}
