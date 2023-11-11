#!/bin/bash
set -eux -o pipefail

go get k8s.io/code-generator/cmd/go-to-protobuf@v0.16.7-beta.0

bash ${GOPATH}/pkg/mod/k8s.io/code-generator@v0.16.7-beta.0/generate-groups.sh \
  "deepcopy,client,informer,lister" \
  github.com/kubeTasker/kubeTasker/pkg/client github.com/kubeTasker/kubeTasker/pkg/apis \
  taskflow:v1alpha1 \
  --go-header-file ./hack/custom-boilerplate.go.txt
