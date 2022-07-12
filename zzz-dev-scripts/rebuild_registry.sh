#!/usr/bin/zsh

KOPS_PATH=$HOME/Desktop/kops
ETCD_MANAGER_PATH=$HOME/Desktop/etcdadm/etcd-manager
PROFILE=normal

export REGISTRY_NAME=kops
export DOCKER_REGISTRY=rg.fr-par.scw.cloud
export DOCKER_IMAGE_PREFIX=$REGISTRY_NAME/
export DOCKER_TAG=1.25.0-beta.1

if [[ $1 == "-r" ]]
then
    echo "Recreating registry"
    scw registry namespace create name=$REGISTRY_NAME is-public=true description="Stores images needed by kops (things like etcd-manager, dns-controller, kops-controller, etc)" -p $PROFILE
fi

docker login rg.fr-par.scw.cloud/$REGISTRY_NAME -u nologin --password $SCW_SECRET_KEY

cd "$KOPS_PATH" || exit
printf "\nKOPS-CONTROLLER\n"
make kops-controller-push
printf "\nDNS-CONTROLLER\n"
make dns-controller-push
printf "\nKUBE-API-SERVER-HEALTHCHECK\n"
make kube-apiserver-healthcheck-push

cd "$ETCD_MANAGER_PATH" || exit
printf "\nETCD-MANAGER\n"
make push-etcd-manager
#printf "\nETCD-MANAGER MANIFESTS\n"
#make push-etcd-manager-manifest
