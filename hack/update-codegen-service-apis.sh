#!/usr/bin/env bash

# Copyright 2020 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -o errexit  # Exits immediately on unexpected errors (does not bypass traps)
set -o nounset  # Errors if variables are used without first being defined
set -o pipefail # Non-zero exit codes in piped commands causes pipeline to fail
                # with that code

go install k8s.io/code-generator/cmd/{client-gen,lister-gen,informer-gen,deepcopy-gen,register-gen}

# Go installs the above commands to get installed in $GOBIN if defined, and $GOPATH/bin otherwise:
GOBIN="$(go env GOBIN)"
gobin="${GOBIN:-$(go env GOPATH)/bin}"

OUTPUT_PKG=github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/third_party/service-apis/client
FQ_APIS=github.com/vmware-tanzu/service-apis/apis/v1alpha1pre1
CLIENTSET_NAME=versioned
CLIENTSET_PKG_NAME=clientset
BOILERPLATE_FILE="${GOPATH}/src/github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/hack/boilerplate.go.txt"

echo "Generating clientset at ${OUTPUT_PKG}/${CLIENTSET_PKG_NAME}"
"${gobin}/client-gen" --clientset-name "${CLIENTSET_NAME}" \
  --input-base "" \
  --input "${FQ_APIS}" \
  --output-package "${OUTPUT_PKG}/${CLIENTSET_PKG_NAME}" \
  --go-header-file "${BOILERPLATE_FILE}" \
  ${COMMON_FLAGS-}

echo "Generating listers at ${OUTPUT_PKG}/listers"
"${gobin}/lister-gen" --input-dirs "${FQ_APIS}" \
  --output-package "${OUTPUT_PKG}/listers" \
  --go-header-file "${BOILERPLATE_FILE}" \
  ${COMMON_FLAGS-}

echo "Generating informers at ${OUTPUT_PKG}/informers"
"${gobin}/informer-gen" \
  --input-dirs "${FQ_APIS}" \
  --versioned-clientset-package "${OUTPUT_PKG}/${CLIENTSET_PKG_NAME}/${CLIENTSET_NAME}" \
  --listers-package "${OUTPUT_PKG}/listers" \
  --output-package "${OUTPUT_PKG}/informers" \
  --go-header-file "${BOILERPLATE_FILE}" \
  ${COMMON_FLAGS-}
