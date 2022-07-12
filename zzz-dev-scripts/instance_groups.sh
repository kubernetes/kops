###########################################
#         cluster.leila.sieben.fr         #
###########################################

# NODE
go run -v ./cmd/kops/ create instancegroup -v10 --name=cluster.leila.sieben.fr extra-node --role node
go run -v ./cmd/kops/ delete instancegroup -v10 --name=cluster.leila.sieben.fr extra-node

# MASTER
go run -v ./cmd/kops get cluster -o yaml > mycluster.yaml
go run -v ./cmd/kops replace -f zzz-dev-scripts/cluster.leila.sieben.fr_extra_masters.yaml
#go run -v ./cmd/kops/ edit cluster -v10 --name=cluster.leila.sieben.fr
go run -v ./cmd/kops/ create instancegroup -v10 --name=cluster.leila.sieben.fr master2 --role master --subnet fr-par-1
go run -v ./cmd/kops/ create instancegroup -v10 --name=cluster.leila.sieben.fr master3 --role master --subnet fr-par-1
go run -v ./cmd/kops/ delete instancegroup -v10 --name=cluster.leila.sieben.fr master2
go run -v ./cmd/kops/ delete instancegroup -v10 --name=cluster.leila.sieben.fr master3

###########################################
#            cluster.k8s.local            #
###########################################

# NODE
go run -v ./cmd/kops/ create instancegroup -v10 --name=cluster.k8s.local extra-node --role node
go run -v ./cmd/kops/ delete instancegroup -v10 --name=cluster.k8s.local extra-node

# MASTER
go run -v ./cmd/kops get cluster -o yaml > cluster.k8s.local_simple.yaml
go run -v ./cmd/kops replace -f zzz-dev-scripts/cluster.k8s.local_extra_masters.yaml
#go run -v ./cmd/kops/ edit cluster -v10 --name=cluster.k8s.local
go run -v ./cmd/kops/ create instancegroup -v10 --name=cluster.k8s.local master2 --role master --subnet fr-par-1
go run -v ./cmd/kops/ create instancegroup -v10 --name=cluster.k8s.local master3 --role master --subnet fr-par-1
go run -v ./cmd/kops/ delete instancegroup -v10 --name=cluster.k8s.local master2
go run -v ./cmd/kops/ delete instancegroup -v10 --name=cluster.k8s.local master3

