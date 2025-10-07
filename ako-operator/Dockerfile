ARG golang_src_repo=golang:latest
ARG ubi_src_repo=registry.access.redhat.com/ubi9/ubi:latest

# Build the ako-operator binary
FROM ${golang_src_repo} as builder

ENV BUILD_PATH="github.com/vmware/load-balancer-and-ingress-services-for-kubernetes"
RUN mkdir -p $GOPATH/src/$BUILD_PATH

COPY ../ $GOPATH/src/$BUILD_PATH
WORKDIR $GOPATH/src/$BUILD_PATH/ako-operator

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -buildvcs=false -o $GOPATH/bin/ako-operator -mod=vendor ./cmd/main.go


# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM ${ubi_src_repo}

LABEL name="ako-operator"
LABEL summary="An AKO operator to deploy AKO controller"
LABEL version="v1.13.2"
LABEL release="1"
LABEL description="Manage configmap, statefulset and other artifacts for deploying AKO controller"
LABEL vendor="VMware by Broadcom"
LABEL maintainer="VMware by Broadcom"

RUN yum -y update-minimal --security --sec-severity=Important --sec-severity=Critical

WORKDIR /
COPY LICENSE /licenses/
COPY --from=builder /go/bin/ako-operator .
RUN mkdir -p /var/crds
COPY ../helm/ako/crds/ako.vmware.com_hostrules.yaml /var/crds/
COPY ../helm/ako/crds/ako.vmware.com_httprules.yaml /var/crds/
COPY ../helm/ako/crds/ako.vmware.com_aviinfrasettings.yaml /var/crds/
COPY ../helm/ako/crds/ako.vmware.com_l4rules.yaml /var/crds/
COPY ../helm/ako/crds/ako.vmware.com_ssorules.yaml /var/crds/
COPY ../helm/ako/crds/ako_vmware.com_l7rules.yaml /var/crds/
RUN adduser nonroot
USER nonroot:nonroot

ENTRYPOINT ["/ako-operator"]
