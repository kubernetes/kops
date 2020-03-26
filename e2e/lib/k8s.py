from kubernetes import client, config

from os import path

class KubernetesClient(object):
  def __init__(self, api_client=None):
    if not api_client:
      api_client = config.new_client_from_config()

    self.api_client = api_client
    self.v1 = client.CoreV1Api(api_client=api_client)
    self.storagev1 = client.StorageV1Api(api_client=api_client)
    self.appsv1 = client.AppsV1Api(api_client=api_client)

  def version(self):
    versionapi = client.VersionApi(api_client=self.api_client)
    return versionapi.get_code()

  def pods(self):
    return list(map(lambda obj: Pod(self, obj), self.v1.list_namespaced_pod(namespace="").items))

  def namespaces(self):
    return [Namespace(self, o) for o in self.v1.list_namespace().items]

  def storage_classes(self):
    return [StorageClass(self, o) for o in self.storagev1.list_storage_class().items]

  def daemon_sets(self):
    return [DaemonSet(self, o) for o in self.appsv1.list_namespaced_daemon_set(namespace="").items]

  def stateful_sets(self):
    return [StatefulSet(self, o) for o in self.appsv1.list_namespaced_stateful_set(namespace="").items]

  def nodes(self):
    return list(map(lambda obj: Node(self, obj), self.v1.list_node().items))

  def namespace(self, name):
    obj = self.v1.read_namespace(name)
    return Namespace(self, obj)

class KubernetesObject(object):
  def __init__(self, k8s, obj, kind):
    self.k8s = k8s
    self.obj = obj
    self.kind = kind
    self.namespace = self.obj.metadata.namespace
    self.name = self.obj.metadata.name

  def __repr__(self):
    return self.kind + ":" + self.name

  def _repr_json_(self):
    j = {
        "kind": self.kind,
        "name": self.obj.metadata.name,
    }
    if self.obj.metadata.namespace:
      j["namepace"] = self.obj.metadata.namespace
    return j

  def get_annotation(self, key):
    annotations = self.obj.metadata.annotations or {}
    return annotations.get(key)

class Node(KubernetesObject):
  def __init__(self, k8s, obj):
    super().__init__(k8s, obj, "Node")

  def bad_conditions(self):
    return [c for c in self.obj.status.conditions if not _condition_is_ok(c)]

  def ready(self):
    return len(self.bad_conditions()) == 0

def _condition_is_ok(c):
  if c.type == "Ready":
    return c.status == "True"
  return c.status == "False"

class Namespace(KubernetesObject):
  def __init__(self, k8s, obj):
    super().__init__(k8s, obj, "Namespace")

  def pods(self):
    response = self.k8s.v1.list_namespaced_pod(self.obj.metadata.name)
    return list(map(lambda obj: Pod(self.k8s, obj), response.items))

class StorageClass(KubernetesObject):
  def __init__(self, k8s, obj):
    super().__init__(k8s, obj, "StorageClass")

  def is_default_class(self):
    return self.get_annotation('storageclass.kubernetes.io/is-default-class') == 'true' or self.get_annotation('storageclass.beta.kubernetes.io/is-default-class') == 'true'

class StatefulSet(KubernetesObject):
  def __init__(self, k8s, obj):
    super().__init__(k8s, obj, "StatefulSet")

class DaemonSet(KubernetesObject):
  def __init__(self, k8s, obj):
    super().__init__(k8s, obj, "DaemonSet")

class Pod(KubernetesObject):
  def __init__(self, k8s, obj):
    super().__init__(k8s, obj, "Pod")
