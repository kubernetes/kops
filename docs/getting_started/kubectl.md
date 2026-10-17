# kubectl cluster admin configuration

When you run `kops create cluster --yes`, you automatically get a kubectl configuration for accessing the cluster. This configuration gives you full admin access to the cluster for 18 hours.

If you want to create this configuration on other machine, you can run the following as long as you have access to the kOps state store:

```
export KOPS_STATE_STORE=<location of the kops state store>
NAME=<kubernetes.mydomain.com>
kops export kubeconfig ${NAME} --admin
```

Warning: Note that the exported configuration gives you full admin privileges for 18 hours. For regular kubectl usage, you should consider using another method for authenticating to the cluster.

To create the kubeconfig configuration settings without the admin credential:

```
export KOPS_STATE_STORE=<location of the kops state store>
NAME=<kubernetes.mydomain.com>
kops export kubeconfig ${NAME}
```

## Custom API endpoints and TLS

`--api-server=https://endpoint.example.com:6443` changes the destination URL used by kOps or written by `kops export kubeconfig`. Generated connections still use `api.internal.<cluster-name>` as the TLS server name. This name is included in the API server certificate and allows certificate verification when connecting through a load balancer DNS name, IP address or tunnel address that is absent from the certificate. A DNS TLS server name controls both the SNI sent during the TLS handshake and the name checked in the server certificate. It does not change the network destination or the HTTP Host header.

For a proxy that needs a different TLS identity, CA or authentication method, configure and verify a working local kubeconfig, then use `--use-kubeconfig`:

```bash
kubectl --context="${NAME}" get nodes
kops validate cluster --name="${NAME}" --use-kubeconfig
kops rolling-update cluster --name="${NAME}" --use-kubeconfig
kops reconcile cluster --name="${NAME}" --use-kubeconfig
```

kOps respects `KUBECONFIG` and selects the context named for the kOps cluster, regardless of the current context. It reuses that context's endpoint, TLS, proxy and authentication settings. When combined with `--api-server`, only the URL is overridden:

| Configuration | TLS identity used for verification and SNI |
| --- | --- |
| Generated, with or without `--api-server` | The cluster's internal API name |
| `--use-kubeconfig` with `tls-server-name` | The explicit kubeconfig TLS name, even with `--api-server` |
| `--use-kubeconfig` without `tls-server-name` | The selected endpoint hostname, including an `--api-server` override |

When the TLS identity is an IP literal, it must match an IP SAN in the certificate; no SNI is sent. For a proxy that routes using the endpoint's DNS name, remove an inherited `tls-server-name` from the kubeconfig's cluster entry or set it to the name required by the proxy. The proxy must present a certificate valid for that name and trusted by the kubeconfig's certificate authority configuration (or system roots when no CA is configured).

Changing the endpoint or TLS name does not add trust for a proxy's CA, supply proxy credentials, or change Kubernetes authentication. Use the proxy's required CA and credentials in the local kubeconfig; keep certificate verification enabled. `proxy-url` configures a forward proxy and is separate from the API server URL.

`kops export kubeconfig` generates connection settings from cluster state and overwrites the endpoint and TLS server name in its target cluster entry. Reapply custom settings after exporting, or keep the proxy kubeconfig in a separate file selected with `KUBECONFIG`.
