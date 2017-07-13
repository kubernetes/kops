go get github.com/golang/protobuf/protoc-gen-go

pushd $GOPATH/src/github.com/googleapis/gnostic/generator 
go build
cd ..
./generator/generator
popd

pushd $GOPATH/src/github.com/googleapis/gnostic/generator
go install
popd

pushd $GOPATH/src/github.com/googleapis/gnostic/OpenAPIv2
protoc \
--go_out=Mgoogle/protobuf/any.proto=github.com/golang/protobuf/ptypes/any:. OpenAPIv2.proto 
go build
go install
popd
