****************
Work in progress
****************


Notes:

* Only works if you haven't made changes to the kube_env.yaml file (which includes assets)


## Procedure

To get the external IPs of all nodes:

```
IPS=`kubectl get nodes -o jsonpath='{.items[*].status.addresses[?(@.type=="ExternalIP")].address}'`
echo $IPS
```

Then to apply updates:

```
for ip in $IPS; do

echo "Updating ${ip}"

echo "Sleeping for 30 seconds first"
sleep 30

cat <<'EOF' | ssh admin@${ip} 'sudo bash -s'
#/bin/bash
set -e
set -x
NODEUP_URL=https://kubeupv2.s3.amazonaws.com/kops/1.4.0/linux/amd64/nodeup

INSTALL_DIR="/var/cache/kubernetes-install"
mkdir -p ${INSTALL_DIR}  
cd ${INSTALL_DIR}

rm -rf nodeup
curl -f --ipv4 -Lo "nodeup" --connect-timeout 20 --retry 6 --retry-delay 10 "${NODEUP_URL}"
chmod +x nodeup

( ./nodeup --conf=/var/cache/kubernetes-install/kube_env.yaml --v=8 )
EOF

done

echo "Done!"
```