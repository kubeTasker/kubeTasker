#!/bin/bash
set -eux -o pipefail

go get k8s.io/code-generator/cmd/go-to-protobuf
go get github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway@v1.16.0
go get github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger
go get github.com/gogo/protobuf/protoc-gen-gogofast@v1.3.2
go get github.com/gogo/protobuf/gogoproto@v1.3.2

go install k8s.io/code-generator/cmd/go-to-protobuf

go-to-protobuf \
    --go-header-file=./hack/custom-boilerplate.go.txt \
    --packages=github.com/kubeTasker/kubeTasker/pkg/apis/workflow/v1alpha1 \
    --apimachinery-packages=+k8s.io/apimachinery/pkg/util/intstr,+k8s.io/apimachinery/pkg/api/resource,k8s.io/apimachinery/pkg/runtime/schema,+k8s.io/apimachinery/pkg/runtime,k8s.io/apimachinery/pkg/apis/meta/v1,k8s.io/api/core/v1

for f in $(find pkg -name '*.proto'); do
    protoc \
        -I /usr/local/include \
        -I . \
        -I ${GOPATH}/src \
        -I ${GOPATH}/pkg/mod/github.com/gogo/protobuf@v1.3.2/gogoproto \
        -I ${GOPATH}/pkg/mod/github.com/grpc-ecosystem/grpc-gateway@v1.16.0/third_party/googleapis \
        --gogofast_out=plugins=grpc:${GOPATH}/src \
        --grpc-gateway_out=logtostderr=true:${GOPATH}/src \
        --swagger_out=logtostderr=true,fqn_for_swagger_name=true:. \
        $f
done
