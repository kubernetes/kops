CLUSTER_NAME=$1

go run -v ./cmd/kops replace -f "$CLUSTER_NAME"_extra_masters.yaml
go run -v ./cmd/kops/ create instancegroup -v10 --name=$CLUSTER_NAME master2 --role master --subnet fr-par-1
go run -v ./cmd/kops/ create instancegroup -v10 --name=$CLUSTER_NAME master3 --role master --subnet fr-par-1
go run -v ./cmd/kops/ update cluster -v10 --name=$CLUSTER_NAME --yes