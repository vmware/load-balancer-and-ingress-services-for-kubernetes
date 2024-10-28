GOCMD=go
GOBUILD=$(GOCMD) build -buildvcs=false
GOCLEAN=$(GOCMD) clean
GOGET=$(GOCMD) get
GOTEST=$(GOCMD) test
BINARY_NAME_AKO=ako
BINARY_NAME_AKO_INFRA=ako-infra
BINARY_NAME_AKO_GATEWAY_API=ako-gateway-api
PACKAGE_PATH_AKO=github.com/vmware/load-balancer-and-ingress-services-for-kubernetes
REL_PATH_AKO=$(PACKAGE_PATH_AKO)/cmd/ako-main
REL_PATH_AKO_INFRA=$(PACKAGE_PATH_AKO)/cmd/infra-main
REL_PATH_AKO_GATEWAY_API=$(PACKAGE_PATH_AKO)/cmd/gateway-api
AKO_OPERATOR_IMAGE=ako-operator
ENDPOINTSLICES_ENABLED?="false"
define GetSupportabilityMatrix
$(shell node -p "require('./buildsettings.json').$(1)")
endef

AVI_MIN_VERSION:=$(call GetSupportabilityMatrix,avi.minVersion)
AVI_MAX_VERSION:=$(call GetSupportabilityMatrix,avi.maxVersion)
K8S_MIN_VERSION:=$(call GetSupportabilityMatrix,kubernetes.minVersion)
K8S_MAX_VERSION:=$(call GetSupportabilityMatrix,kubernetes.maxVersion)
AKO_VERSION:=v$(call GetSupportabilityMatrix,version)
AKO_LDFLAGS:="-X 'main.version=$(AKO_VERSION)' \
		-X '$(PACKAGE_PATH_AKO)/internal/lib.aviMinVersion=$(AVI_MIN_VERSION)' \
		-X '$(PACKAGE_PATH_AKO)/internal/lib.aviMaxVersion=$(AVI_MAX_VERSION)' \
		-X '$(PACKAGE_PATH_AKO)/internal/lib.k8sMinVersion=$(K8S_MIN_VERSION)' \
		-X '$(PACKAGE_PATH_AKO)/internal/lib.k8sMaxVersion=$(K8S_MAX_VERSION)'"

ifdef GOLANG_SRC_REPO
	BUILD_GO_IMG=$(GOLANG_SRC_REPO)
else
	BUILD_GO_IMG=golang:latest
endif

GO_IMG_TEST=golang:bullseye
.PHONY: glob-vars
glob-vars:
	$(eval BUILD_ARG_AKO_LDFLAGS=--build-arg AKO_LDFLAGS=$(AKO_LDFLAGS))

ifndef BUILD_TAG
	$(eval BUILD_TAG=$(shell ./hack/jenkins/get_build_version.sh "dummy" 0))
endif

ifndef BUILD_TIME
	$(eval BUILD_TIME=$(shell date +%Y-%m-%d_%H:%M:%S_%Z))
endif

ifdef GOLANG_SRC_REPO
	$(eval BUILD_ARG_GOLANG=--build-arg golang_src_repo=$(GOLANG_SRC_REPO))
else
	$(eval BUILD_ARG_GOLANG=)
endif

ifdef PHOTON_SRC_REPO
	$(eval BUILD_ARG_PHOTON=--build-arg photon_src_repo=$(PHOTON_SRC_REPO))
else
	$(eval BUILD_ARG_PHOTON=)
endif

ifdef UBI_SRC_REPO
	$(eval BUILD_ARG_UBI=--build-arg ubi_src_repo=$(UBI_SRC_REPO))
else
	$(eval BUILD_ARG_UBI=)
endif

.PHONY: all
all: build docker

.PHONY: sync-crd-files
sync-crd-files:
		cp ./helm/ako/crds/* ./ako-operator/helm/ako-operator/crds/

.PHONY: pre-build
pre-build: sync-crd-files

# builds
.PHONY: build
build: pre-build glob-vars
		sudo docker run \
		-w=/go/src/$(PACKAGE_PATH_AKO) \
		-v $(PWD):/go/src/$(PACKAGE_PATH_AKO) $(BUILD_GO_IMG) \
		$(GOBUILD) \
		-o /go/src/$(PACKAGE_PATH_AKO)/bin/$(BINARY_NAME_AKO) \
		-ldflags $(AKO_LDFLAGS) \
		-mod=vendor \
		/go/src/$(REL_PATH_AKO)

.PHONY: build-infra
build-infra: pre-build glob-vars
		sudo docker run \
		-w=/go/src/$(PACKAGE_PATH_AKO) \
		-v $(PWD):/go/src/$(PACKAGE_PATH_AKO) $(BUILD_GO_IMG) \
		$(GOBUILD) \
		-o /go/src/$(PACKAGE_PATH_AKO)/bin/$(BINARY_NAME_AKO_INFRA) \
		-ldflags $(AKO_LDFLAGS) \
		-mod=vendor \
		/go/src/$(REL_PATH_AKO_INFRA)

.PHONY: build-gateway-api
build-gateway-api: glob-vars
		sudo docker run \
		-w=/go/src/$(PACKAGE_PATH_AKO) \
		-v $(PWD):/go/src/$(PACKAGE_PATH_AKO) $(BUILD_GO_IMG) \
		$(GOBUILD) \
		-o /go/src/$(PACKAGE_PATH_AKO)/bin/$(BINARY_NAME_AKO_GATEWAY_API) \
		-ldflags $(AKO_LDFLAGS) \
		-mod=vendor \
		/go/src/$(REL_PATH_AKO_GATEWAY_API)

.PHONY: build-local
build-local: pre-build
		$(GOBUILD) \
		-o bin/$(BINARY_NAME_AKO) \
		-ldflags $(AKO_LDFLAGS) \
		-mod=vendor \
		./cmd/ako-main

.PHONY: build-local-infra
build-local-infra: pre-build
		$(GOBUILD) \
		-o bin/$(BINARY_NAME_AKO_INFRA) \
		-ldflags $(AKO_LDFLAGS) \
		-mod=vendor \
		./cmd/infra-main

.PHONY: clean
clean:
		$(GOCLEAN) -mod=vendor $(REL_PATH_AKO)
		rm -f bin/$(BINARY_NAME_AKO)

.PHONY: deps
deps:
	dep ensure -v

# docker images
.PHONY: docker
docker: glob-vars
	sudo docker build \
	-t $(BINARY_NAME_AKO):latest \
	--label "BUILD_TAG=$(BUILD_TAG)" \
	--label "BUILD_TIME=$(BUILD_TIME)" \
	$(BUILD_ARG_GOLANG) $(BUILD_ARG_PHOTON) $(BUILD_ARG_AKO_LDFLAGS) \
	-f Dockerfile.ako .

.PHONY: ako-infra-docker
ako-infra-docker: glob-vars
	sudo docker build \
	-t $(BINARY_NAME_AKO_INFRA):latest \
	--label "BUILD_TAG=$(BUILD_TAG)" \
	--label "BUILD_TIME=$(BUILD_TIME)" \
	$(BUILD_ARG_GOLANG) $(BUILD_ARG_PHOTON) $(BUILD_ARG_AKO_LDFLAGS) \
	-f Dockerfile.ako-infra .

.PHONY: ako-operator-docker
ako-operator-docker: glob-vars
	sudo docker build \
	-t $(AKO_OPERATOR_IMAGE):latest \
	--label "BUILD_TAG=$(BUILD_TAG)" \
	--label "BUILD_TIME=$(BUILD_TIME)" \
	$(BUILD_ARG_GOLANG) $(BUILD_ARG_UBI) \
	-f Dockerfile.ako-operator .

.PHONY: ako-gateway-api-docker
ako-gateway-api-docker: glob-vars
	sudo docker build \
	-t $(BINARY_NAME_AKO_GATEWAY_API):latest \
	--label "BUILD_TAG=$(BUILD_TAG)" \
	--label "BUILD_TIME=$(BUILD_TIME)" \
	$(BUILD_ARG_GOLANG) $(BUILD_ARG_PHOTON) $(BUILD_ARG_AKO_LDFLAGS) \
	-f Dockerfile.ako-gateway-api .

# tests
.PHONY: k8stest
k8stest:
	sudo docker run \
	-e ENDPOINTSLICES_ENABLED=$(ENDPOINTSLICES_ENABLED) \
	-w=/go/src/$(PACKAGE_PATH_AKO) \
	-v $(PWD):/go/src/$(PACKAGE_PATH_AKO) $(GO_IMG_TEST) \
	$(GOTEST) -v -mod=vendor $(PACKAGE_PATH_AKO)/tests/k8stest -failfast -timeout 0 \
	-coverprofile cover-1.out -coverpkg=./... 2>&1 | \
	awk '{print "[k8stest]"$$0; if ($$0 ~ /^FAIL/) exit 1;}'

.PHONY: integrationtest
integrationtest:
	sudo docker run \
	-e ENDPOINTSLICES_ENABLED=$(ENDPOINTSLICES_ENABLED) \
	-w=/go/src/$(PACKAGE_PATH_AKO) \
	-v $(PWD):/go/src/$(PACKAGE_PATH_AKO) $(GO_IMG_TEST) \
	$(GOTEST) -v -mod=vendor $(PACKAGE_PATH_AKO)/tests/integrationtest -failfast -coverprofile cover-2.out -coverpkg=./... | buffer -u 500

.PHONY: ingresstests
ingresstests:
	sudo docker run \
	-e ENDPOINTSLICES_ENABLED=$(ENDPOINTSLICES_ENABLED) \
	-w=/go/src/$(PACKAGE_PATH_AKO) \
	-v $(PWD):/go/src/$(PACKAGE_PATH_AKO) $(GO_IMG_TEST) \
	$(GOTEST) -v -mod=vendor $(PACKAGE_PATH_AKO)/tests/ingresstests -failfast -timeout 0 -coverprofile cover-3.out -coverpkg=./... | buffer -u 500

.PHONY: oshiftroutetests
oshiftroutetests:
	sudo docker run \
	-e ENDPOINTSLICES_ENABLED=$(ENDPOINTSLICES_ENABLED) \
	-w=/go/src/$(PACKAGE_PATH_AKO) \
	-v $(PWD):/go/src/$(PACKAGE_PATH_AKO) $(GO_IMG_TEST) \
	$(GOTEST) -v -mod=vendor $(PACKAGE_PATH_AKO)/tests/oshiftroutetests -failfast -timeout 0 \
	-coverprofile cover-4.out -coverpkg=./... 2>&1 | \
	awk '{print "[oshiftroutetests]"$$0; if ($$0 ~ /^FAIL/) exit 1;}'

.PHONY: bootuptests
bootuptests:
	sudo docker run \
	-w=/go/src/$(PACKAGE_PATH_AKO) \
	-v $(PWD):/go/src/$(PACKAGE_PATH_AKO) $(GO_IMG_TEST) \
	$(GOTEST) -v -mod=vendor $(PACKAGE_PATH_AKO)/tests/bootuptests -failfast -timeout 0 \
	-coverprofile cover-5.out -coverpkg=./... 2>&1 | \
	awk '{print "[bootuptests]"$$0; if ($$0 ~ /^FAIL/) exit 1;}'

.PHONY: multicloudtests
multicloudtests:
	sudo docker run \
	-w=/go/src/$(PACKAGE_PATH_AKO) \
	-v $(PWD):/go/src/$(PACKAGE_PATH_AKO) $(GO_IMG_TEST) \
	$(GOTEST) -v -mod=vendor $(PACKAGE_PATH_AKO)/tests/multicloudtests -failfast -timeout 0 \
	-coverprofile cover-6.out -coverpkg=./... 2>&1 | \
	awk '{print "[multicloudtests]"$$0; if ($$0 ~ /^FAIL/) exit 1;}'

.PHONY: servicesapitests
servicesapitests:
	sudo docker run \
	-w=/go/src/$(PACKAGE_PATH_AKO) \
	-v $(PWD):/go/src/$(PACKAGE_PATH_AKO) $(GO_IMG_TEST) \
	$(GOTEST) -v -mod=vendor $(PACKAGE_PATH_AKO)/tests/servicesapitests -failfast -timeout 0 \
	-coverprofile cover-7.out -coverpkg=./... 2>&1 | \
	awk '{print "[servicesapitests]"$$0; if ($$0 ~ /^FAIL/) exit 1;}'

.PHONY: advl4tests
advl4tests:
	sudo docker run \
	-w=/go/src/$(PACKAGE_PATH_AKO) \
	-v $(PWD):/go/src/$(PACKAGE_PATH_AKO) $(GO_IMG_TEST) \
	$(GOTEST) -v -mod=vendor $(PACKAGE_PATH_AKO)/tests/advl4tests -failfast -timeout 0 \
	-coverprofile cover-8.out -coverpkg=./... 2>&1 | \
	awk '{print "[advl4tests]"$$0; if ($$0 ~ /^FAIL/) exit 1;}'

.PHONY: namespacesynctests 
namespacesynctests:
	sudo docker run \
	-w=/go/src/$(PACKAGE_PATH_AKO) \
	-v $(PWD):/go/src/$(PACKAGE_PATH_AKO) $(GO_IMG_TEST) \
	$(GOTEST) -v -mod=vendor $(PACKAGE_PATH_AKO)/tests/namespacesynctests -failfast -timeout 0 \
	-coverprofile cover-9.out -coverpkg=./... 2>&1 | \
	awk '{print "[namespacesynctests]"$$0; if ($$0 ~ /^FAIL/) exit 1;}'

.PHONY: misc 
temp:
	sudo docker run \
	-w=/go/src/$(PACKAGE_PATH_AKO) \
	-v $(PWD):/go/src/$(PACKAGE_PATH_AKO) $(GO_IMG_TEST) \
	$(GOTEST) -v -mod=vendor $(PACKAGE_PATH_AKO)/tests/temp -failfast 2>&1 | \
	awk '{print "[temp]"$$0; if ($$0 ~ /^FAIL/) exit 1;}'

.PHONY: npltests 
npltests:
	sudo docker run \
	-e ENDPOINTSLICES_ENABLED=$(ENDPOINTSLICES_ENABLED) \
	-w=/go/src/$(PACKAGE_PATH_AKO) \
	-v $(PWD):/go/src/$(PACKAGE_PATH_AKO) $(GO_IMG_TEST) \
	$(GOTEST) -v -mod=vendor $(PACKAGE_PATH_AKO)/tests/npltests -failfast -timeout 0 \
	-coverprofile cover-10.out -coverpkg=./... 2>&1 | \
	awk '{print "[npltests]"$$0; if ($$0 ~ /^FAIL/) exit 1;}'

.PHONY: evhtests 
evhtests:
	sudo docker run \
	-e ENDPOINTSLICES_ENABLED=$(ENDPOINTSLICES_ENABLED) \
	-w=/go/src/$(PACKAGE_PATH_AKO) \
	-v $(PWD):/go/src/$(PACKAGE_PATH_AKO) $(GO_IMG_TEST) \
	$(GOTEST) -v -mod=vendor $(PACKAGE_PATH_AKO)/tests/evhtests -failfast -timeout 0 \
	-coverprofile cover-11.out -coverpkg=./... 2>&1 | \
	awk '{print "[evhtests]"$$0; if ($$0 ~ /^FAIL/) exit 1;}'

.PHONY: vippernstests
vippernstests:
	sudo docker run \
	-w=/go/src/$(PACKAGE_PATH_AKO) \
	-v $(PWD):/go/src/$(PACKAGE_PATH_AKO) $(GO_IMG_TEST) \
	$(GOTEST) -v -mod=vendor $(PACKAGE_PATH_AKO)/tests/evhtests -failfast -timeout 0 -isVipPerNS=true \
	-coverprofile cover-12.out -coverpkg=./... 2>&1 | \
	awk '{print "[vippernstests]"$$0; if ($$0 ~ /^FAIL/) exit 1;}'

.PHONY: dedicatedevhtests
dedicatedevhtests:
	sudo docker run \
	-e ENDPOINTSLICES_ENABLED=$(ENDPOINTSLICES_ENABLED) \
	-w=/go/src/$(PACKAGE_PATH_AKO) \
	-v $(PWD):/go/src/$(PACKAGE_PATH_AKO) $(GO_IMG_TEST) \
	$(GOTEST) -v -mod=vendor $(PACKAGE_PATH_AKO)/tests/dedicatedevhtests -failfast -timeout 0 \
	-coverprofile cover-13.out -coverpkg=./... 2>&1 | \
	awk '{print "[dedicatedevhtests]"$$0; if ($$0 ~ /^FAIL/) exit 1;}'

.PHONY: dedicatedvippernstests
dedicatedvippernstests:
	sudo docker run \
	-w=/go/src/$(PACKAGE_PATH_AKO) \
	-v $(PWD):/go/src/$(PACKAGE_PATH_AKO) $(GO_IMG_TEST) \
	$(GOTEST) -v -mod=vendor $(PACKAGE_PATH_AKO)/tests/dedicatedevhtests -failfast -timeout 0 -isVipPerNS=true \
	-coverprofile cover-14.out -coverpkg=./... 2>&1 | \
	awk '{print "[dedicatedvippernstests]"$$0; if ($$0 ~ /^FAIL/) exit 1;}'

.PHONY: dedicatedvstests
dedicatedvstests:
	sudo docker run \
	-e ENDPOINTSLICES_ENABLED=$(ENDPOINTSLICES_ENABLED) \
	-w=/go/src/$(PACKAGE_PATH_AKO) \
	-v $(PWD):/go/src/$(PACKAGE_PATH_AKO) $(GO_IMG_TEST) \
	$(GOTEST) -v -mod=vendor $(PACKAGE_PATH_AKO)/tests/dedicatedvstests -failfast -timeout 0 \
	-coverprofile cover-15.out -coverpkg=./... 2>&1 | \
	awk '{print "[dedicatedvstests]"$$0; if ($$0 ~ /^FAIL/) exit 1;}'

.PHONY: infratests
infratests:
	sudo docker run \
	-w=/go/src/$(PACKAGE_PATH_AKO) \
	-v $(PWD):/go/src/$(PACKAGE_PATH_AKO) $(GO_IMG_TEST) \
	$(GOTEST) -v -mod=vendor $(PACKAGE_PATH_AKO)/tests/infratests -failfast -timeout 0 2>&1 | \
	awk '{print "[infratests]"$$0; if ($$0 ~ /^FAIL/) exit 1;}'


# .PHONY: multiclusteringresstests
# multiclusteringresstests:
#   sudo docker run \
#   -w=/go/src/$(PACKAGE_PATH_AKO) \
#   -v $(PWD):/go/src/$(PACKAGE_PATH_AKO) $(BUILD_GO_IMG) \
#   $(GOTEST) -v -mod=vendor $(PACKAGE_PATH_AKO)/tests/multiclusteringresstests -failfast -coverprofile cover-16.out -coverpkg=./...


.PHONY: hatests
hatests:
	sudo docker run \
	-w=/go/src/$(PACKAGE_PATH_AKO) \
	-v $(PWD):/go/src/$(PACKAGE_PATH_AKO) $(GO_IMG_TEST) \
	$(GOTEST) -v -mod=vendor $(PACKAGE_PATH_AKO)/tests/hatests -failfast -timeout 0 \
	-coverprofile cover-17.out -coverpkg=./... 2>&1 | \
	awk '{print "[hatests]"$$0; if ($$0 ~ /^FAIL/) exit 1;}'

.PHONY: calicotests
calicotests:
	sudo docker run \
	-w=/go/src/$(PACKAGE_PATH_AKO) \
	-v $(PWD):/go/src/$(PACKAGE_PATH_AKO) $(GO_IMG_TEST) \
	$(GOTEST) -v -mod=vendor $(PACKAGE_PATH_AKO)/tests/cnitests -failfast -timeout 0 -cniPlugin=calico \
	-coverprofile cover-18.out -coverpkg=./... 2>&1 | \
	awk '{print "[calicotests]"$$0; if ($$0 ~ /^FAIL/) exit 1;}'

.PHONY: ciliumtests
ciliumtests:
	sudo docker run \
	-w=/go/src/$(PACKAGE_PATH_AKO) \
	-v $(PWD):/go/src/$(PACKAGE_PATH_AKO) $(GO_IMG_TEST) \
	$(GOTEST) -v -mod=vendor $(PACKAGE_PATH_AKO)/tests/cnitests -failfast -timeout 0 -cniPlugin=cilium \
	-coverprofile cover-19.out -coverpkg=./... 2>&1 | \
	awk '{print "[ciliumtests]"$$0; if ($$0 ~ /^FAIL/) exit 1;}'

.PHONY: helmtests
helmtests:
	sudo docker run \
	-u root:root \
	-v $(PWD)/helm/ako:/apps \
	-v $(PWD)/tests/helmtests:/apps/tests \
	avi-buildops-docker-registry-02.avilb.broadcom.net:5000/avi-buildops/helmunittest/helm-unittest:3.11.1-0.3.0 .

.PHONY: gatewayapi_ingestiontests
gatewayapi_ingestiontests:
	sudo docker run \
	-e ENDPOINTSLICES_ENABLED=$(ENDPOINTSLICES_ENABLED) \
	-w=/go/src/$(PACKAGE_PATH_AKO) \
	-v $(PWD):/go/src/$(PACKAGE_PATH_AKO) $(GO_IMG_TEST) \
	$(GOTEST) -v -mod=vendor $(PACKAGE_PATH_AKO)/tests/gatewayapitests/ingestion -failfast -timeout 0 \
	-coverprofile cover-20.out -coverpkg=./... 2>&1 | \
	awk '{print "[gatewayapi_ingestiontests]"$$0; if ($$0 ~ /^FAIL/) exit 1;}'

.PHONY: gatewayapi_graphlayertests
gatewayapi_graphlayertests:
	sudo docker run \
	-w=/go/src/$(PACKAGE_PATH_AKO) \
	-v $(PWD):/go/src/$(PACKAGE_PATH_AKO) $(GO_IMG_TEST) \
	$(GOTEST) -v -mod=vendor $(PACKAGE_PATH_AKO)/tests/gatewayapitests/graphlayer -failfast -timeout 0 \
	-coverprofile cover-21.out -coverpkg=./... 2>&1 | \
	awk '{print "[gatewayapi_graphlayertests]"$$0; if ($$0 ~ /^FAIL/) exit 1;}'

.PHONY: gatewayapi_statustests
gatewayapi_statustests:
	sudo docker run \
	-w=/go/src/$(PACKAGE_PATH_AKO) \
	-v $(PWD):/go/src/$(PACKAGE_PATH_AKO) $(GO_IMG_TEST) \
	$(GOTEST) -v -mod=vendor $(PACKAGE_PATH_AKO)/tests/gatewayapitests/status -failfast -timeout 0 \
	-coverprofile cover-22.out -coverpkg=./... 2>&1 | \
	awk '{print "[gatewayapi_statustests]"$$0; if ($$0 ~ /^FAIL/) exit 1;}'

.PHONY: multitenancytests
multitenancytests:
	sudo docker run \
	-w=/go/src/$(PACKAGE_PATH_AKO) \
	-v $(PWD):/go/src/$(PACKAGE_PATH_AKO) $(GO_IMG_TEST) \
	$(GOTEST) -v -mod=vendor $(PACKAGE_PATH_AKO)/tests/multitenancytests -failfast -timeout 0 -coverprofile cover-21.out -coverpkg=./... | buffer -u 500

.PHONY: urltests
urltests:
	sudo docker run \
	-w=/go/src/$(PACKAGE_PATH_AKO) \
	-v $(PWD):/go/src/$(PACKAGE_PATH_AKO) $(GO_IMG_TEST) \
	$(GOTEST) -v -mod=vendor $(PACKAGE_PATH_AKO)/tests/urltests -failfast -coverprofile cover-22.out -coverpkg=./... | sed "s/^/[urltests]/"

.PHONY: int_test
int_test:
	make -j 1 k8stest integrationtest ingresstests evhtests vippernstests dedicatedevhtests dedicatedvippernstests oshiftroutetests bootuptests multicloudtests advl4tests namespacesynctests servicesapitests npltests misc dedicatedvstests hatests calicotests ciliumtests helmtests infratests urltests multitenancytests gatewayapitests

.PHONY: eps_enabled
eps_enabled:
	make -j 1 k8stest integrationtest ingresstests evhtests dedicatedevhtests oshiftroutetests npltests dedicatedvstests gatewayapitests ENDPOINTSLICES_ENABLED="true"  

.PHONY: scale_test
scale_test:
	sudo docker run \
	-w=/go/src/$(PACKAGE_PATH_AKO) \
	--mount type=bind,source=$(TestbedFilePath),target=$(TestbedFilePath) \
	--mount type=bind,source=$(KubeConfigFileName),target=$(KubeConfigFileName) \
	-v $(PWD):/go/src/$(PACKAGE_PATH_AKO) $(GO_IMG_TEST) \
	$(GOTEST) -v -mod=vendor $(PACKAGE_PATH_AKO)/tests/scaletest -failfast -test.timeout=$(Timeout) \
	-kubeConfigFileName=$(KubeConfigFileName) -testbedFileName=$(TestbedFilePath) -numGoRoutines=$(NumGoRoutines) \
	-numOfLBSvc=$(NumOfLBSvc) -numOfIng=$(NumOfIng)

# linting and formatting
GO_FILES := $(shell find . -type d -path ./vendor -prune -o -type f -name '*.go' -print)
.PHONY: fmt
fmt:
	@echo
	@echo "Formatting Go files"
	@gofmt -s -l -w $(GO_FILES)

.golangci-bin:
	@echo "Installing Golangci-lint"
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $@ v1.55.2

.PHONY: golangci
golangci: .golangci-bin
	@echo "Running golangci"
	@GOOS=linux GOGC=1 .golangci-bin/golangci-lint run -c .golangci.yml

.PHONY: golangci-fix
golangci-fix: .golangci-bin
	@echo "Running golangci-fix"
	@GOOS=linux GOGC=1 .golangci-bin/golangci-lint run -c .golangci.yml --fix