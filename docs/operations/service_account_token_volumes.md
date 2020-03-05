Some services, such as istio and Envoy's Secrect Discovery Service (SDS), take advantage of a new feature in kubernetes 1.13+, [Service Account Token Volume Projection](https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/#service-account-token-volume-projection).


1. In order to enable this feaute for kubernetes 1.12+, add the following config to your cluster spec:

    kubeAPIServer:
        apiAudiences:
        - api
        - istio-ca
        serviceAccountIssuer: kubernetes.default.svc
        serviceAccountKeyFile:
        - /srv/kubernetes/server.key
        serviceAccountSigningKeyFile: /srv/kubernetes/server.key

