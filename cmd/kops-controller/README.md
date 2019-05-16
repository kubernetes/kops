kops-controller


Created with



kubebuilder init --license apache2 --domain k8s.io --dep=false

kubebuilder create api --group kops --version v1alpha2 --kind=InstanceGroup --namespaced=true --resource=false --controller=true --example=false
