## kube-apiserver-healthcheck

This is a small sidecar container that allows for health-checking the
kube-apiserver without enabling anonymous authentication and without
enabling the unauthenticated port.

It listens on port 8080 (http), and proxies a few known-safe requests
to the real apiserver listening on 443.  It uses a client certificate
to authenticate itself to the apiserver.

This lets us turn off the unauthenticated kube-apiserver endpoint, but
it also lets us have better load-balancer health-checks.

Because it runs as a sidecar next to kube-apiserver, it is in the same
network namespace, and thus it can reach apiserver on
https://127.0.0.1 .  The kube-apiserver-healthcheck process listens on
8080, but the health checks for the apiserver container are configured
for :8080 and actually go via the sidecar.
