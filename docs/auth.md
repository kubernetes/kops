# Auth

## Authentication

### OpenID Provider (OIDC)

`kops` allows you to set the relevant kube-apiserver options through the cluster yaml (`kops edit cluster`).

Here is an example setting up OIDC with Google:

```
spec:
  kubeAPIServer:
    oidcClientID: 766499101111-4coma6ekm47gg21k1xxjd9gbuh541q5r.apps.googleusercontent.com
    oidcIssuerURL: https://accounts.google.com
    oidcUsernameClaim: email
```

You can find more here - [Kubernetes - Authenticating](http://kubernetes.io/docs/admin/authentication/)

## Authorization

### RBAC

`kops` allows you to set the relevant kube-apiserver options through the cluster yaml (`kops edit cluster`).

Here is an example setting up RBAC:

```
spec:
  kubeAPIServer:
    runtimeConfig:
      - rbac.authorization.k8s.io/v1alpha1: true
    authorizationMode: RBAC
```

You can find more here - [Kubernetes - Using Authorization Plugins](http://kubernetes.io/docs/admin/authorization/)