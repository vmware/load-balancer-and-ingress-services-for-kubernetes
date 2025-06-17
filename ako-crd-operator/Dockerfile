# Build the manager binary
ARG golang_src_repo=golang:latest
ARG photon_src_repo=photon:latest
FROM ${golang_src_repo} as build

WORKDIR /workspace
COPY . /workspace/

# Build
# the GOARCH has not a default value to allow the binary be built according to the host where the command
# was called. For example, if we call make docker-build in a local env which has the Apple Silicon M1 SO
# the docker BUILDPLATFORM arg will be linux/arm64 when for Apple x86 it will be linux/amd64. Therefore,
# by leaving it empty we can ensure that the container and binary shipped on it will have the same platform.
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build  -buildvcs=false -o manager cmd/main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM ${photon_src_repo}
WORKDIR /
COPY --from=build /workspace/manager .
COPY ./scripts/pre_stop_hook.sh /var/pre_stop_hook.sh
RUN chmod +x /var/pre_stop_hook.sh
USER 65532:65532

ENTRYPOINT ["/manager"]
