CLUSTER_NAME=$1
SPEC_FILES_DIR=zzz-dev-scripts

go run -v ./cmd/kops/ delete instancegroup -v10 --name=$CLUSTER_NAME master2 --yes
go run -v ./cmd/kops/ delete instancegroup -v10 --name=$CLUSTER_NAME master3 --yes
go run -v ./cmd/kops replace -f "$SPEC_FILES_DIR/$CLUSTER_NAME"_simple.yaml
go run -v ./cmd/kops/ update cluster -v10 --name=$CLUSTER_NAME --yes
