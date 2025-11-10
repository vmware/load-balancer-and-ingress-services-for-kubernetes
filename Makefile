GOCMD=go
SHELL=/bin/bash -o pipefail -e
GOBUILD=$(GOCMD) build -buildvcs=false
GOCLEAN=$(GOCMD) clean
GOGET=$(GOCMD) get
GOTEST=$(GOCMD) test
BINARY_NAME_AKO=ako
BINARY_NAME_AKO_INFRA=ako-infra
BINARY_NAME_AKO_GATEWAY_API=ako-gateway-api
BINARY_NAME_AKO_CRD_OPERATOR=ako-crd-operator
PACKAGE_PATH_AKO=github.com/vmware/load-balancer-and-ingress-services-for-kubernetes
REL_PATH_AKO=$(PACKAGE_PATH_AKO)/cmd/ako-main
REL_PATH_AKO_INFRA=$(PACKAGE_PATH_AKO)/cmd/infra-main
REL_PATH_AKO_GATEWAY_API=$(PACKAGE_PATH_AKO)/cmd/gateway-api
AKO_OPERATOR_IMAGE=ako-operator
INFORMERS_PACKAGES := $(shell go list ./tests/... | grep informers)
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

# builds
.PHONY: build
build: glob-vars
		sudo docker run \
		-w=/go/src/$(PACKAGE_PATH_AKO) \
		-v $(PWD):/go/src/$(PACKAGE_PATH_AKO) $(BUILD_GO_IMG) \
		$(GOBUILD) \
		-o /go/src/$(PACKAGE_PATH_AKO)/bin/$(BINARY_NAME_AKO) \
		-ldflags $(AKO_LDFLAGS) \
		-mod=vendor \
		/go/src/$(REL_PATH_AKO)

.PHONY: build-infra
build-infra: glob-vars
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
build-local:
		$(GOBUILD) \
		-o bin/$(BINARY_NAME_AKO) \
		-ldflags $(AKO_LDFLAGS) \
		-mod=vendor \
		./cmd/ako-main

.PHONY: build-local-infra
build-local-infra:
		$(GOBUILD) \
		-o bin/$(BINARY_NAME_AKO_INFRA) \
		-ldflags $(AKO_LDFLAGS) \
		-mod=vendor \
		./cmd/infra-main

.PHONY: build-local-gateway-api
build-local-gateway-api:
		$(GOBUILD) \
		-o bin/$(BINARY_NAME_AKO_GATEWAY_API) \
		-ldflags $(AKO_LDFLAGS) \
		-mod=vendor \
		./cmd/gateway-api

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
	-f ako-operator/Dockerfile .

.PHONY: ako-crd-operator-build-all
ako-crd-operator-build:
	make -C ako-crd-operator all

.PHONY: ako-crd-operator-docker-build
ako-crd-operator-docker-build: glob-vars
#	echo Main Makefile: BUILD_ARG_GOLANG=$(BUILD_ARG_GOLANG), BUILD_ARG_PHOTON=$(BUILD_ARG_PHOTON)
#	export IMG=$(BINARY_NAME_AKO_CRD_OPERATOR):latest
#	export BUILD_ARG_GOLANG=$(BUILD_ARG_GOLANG)
#	export BUILD_ARG_PHOTON=$(BUILD_ARG_PHOTON)
	make -C ako-crd-operator docker-build IMG="$(BINARY_NAME_AKO_CRD_OPERATOR):latest" \
    BUILD_ARG_GOLANG="$(BUILD_ARG_GOLANG)" \
    BUILD_ARG_PHOTON="$(BUILD_ARG_PHOTON)" \
    BUILD_TAG="$(BUILD_TAG)" \
    BUILD_TIME="$(BUILD_TIME)"


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
	@> k8s_test.log
	(sudo docker run \
	-w=/go/src/$(PACKAGE_PATH_AKO) \
	-v $(PWD):/go/src/$(PACKAGE_PATH_AKO) $(GO_IMG_TEST) \
	$(GOTEST) -v -mod=vendor $(PACKAGE_PATH_AKO)/tests/k8stest -failfast -timeout 0 \
	-coverprofile cover-1.out -coverpkg=./... > k8s_test.log 2>&1 && echo "k8stest passed") || (echo "k8stest failed" && cat k8s_test.log && exit 1 )
	

.PHONY: integrationtest
integrationtest:
	@> integrationtest.log
	(sudo docker run \
	-w=/go/src/$(PACKAGE_PATH_AKO) \
	-v $(PWD):/go/src/$(PACKAGE_PATH_AKO) $(GO_IMG_TEST) \
	$(GOTEST) -v -mod=vendor $(PACKAGE_PATH_AKO)/tests/integrationtest -failfast -timeout 0 -coverprofile cover-2.out -coverpkg=./...  > integrationtest.log 2>&1 && echo "integrationtest passed") || (echo "integrationtest failed" && cat integrationtest.log && exit 1)

.PHONY: ingresstests
ingresstests:
	@> ingresstests.log
	(sudo docker run \
	-w=/go/src/$(PACKAGE_PATH_AKO) \
	-v $(PWD):/go/src/$(PACKAGE_PATH_AKO) $(GO_IMG_TEST) \
	$(GOTEST) -v -mod=vendor $(PACKAGE_PATH_AKO)/tests/ingresstests -failfast -timeout 0 -coverprofile cover-3.out -coverpkg=./... > ingresstests.log 2>&1 && echo "ingresstests passed") || (echo "ingresstests failed" && cat ingresstests.log && exit 1)

.PHONY: oshiftroutetests
oshiftroutetests:
	@> oshiftroutetests.log
	(sudo docker run \
	-w=/go/src/$(PACKAGE_PATH_AKO) \
	-v $(PWD):/go/src/$(PACKAGE_PATH_AKO) $(GO_IMG_TEST) \
	$(GOTEST) -v -mod=vendor $(PACKAGE_PATH_AKO)/tests/oshiftroutetests -failfast -timeout 0 \
	-coverprofile cover-4.out -coverpkg=./... > oshiftroutetests.log 2>&1 && echo "oshiftroutetests passed") || (echo "oshiftroutetests failed" && cat oshiftroutetests.log && exit 1)
	

.PHONY: bootuptests
bootuptests:
	@> bootuptests.log
	(sudo docker run \
	-w=/go/src/$(PACKAGE_PATH_AKO) \
	-v $(PWD):/go/src/$(PACKAGE_PATH_AKO) $(GO_IMG_TEST) \
	$(GOTEST) -v -mod=vendor $(PACKAGE_PATH_AKO)/tests/bootuptests -failfast -timeout 0 \
	-coverprofile cover-5.out -coverpkg=./... > bootuptests.log 2>&1 && echo "bootuptests passed") || (echo "bootuptests failed" && cat bootuptests.log && exit 1)

.PHONY: multicloudtests
multicloudtests:
	@> multicloudtests.log
	(sudo docker run \
	-w=/go/src/$(PACKAGE_PATH_AKO) \
	-v $(PWD):/go/src/$(PACKAGE_PATH_AKO) $(GO_IMG_TEST) \
	$(GOTEST) -v -mod=vendor $(PACKAGE_PATH_AKO)/tests/multicloudtests -failfast -timeout 0 \
	-coverprofile cover-6.out -coverpkg=./... > multicloudtests.log 2>&1 && echo "multicloudtests passed") || (echo "multicloudtests failed" && cat multicloudtests.log && exit 1)

.PHONY: servicesapitests
servicesapitests:
	@> servicesapitests.log
	(sudo docker run \
	-w=/go/src/$(PACKAGE_PATH_AKO) \
	-v $(PWD):/go/src/$(PACKAGE_PATH_AKO) $(GO_IMG_TEST) \
	$(GOTEST) -v -mod=vendor $(PACKAGE_PATH_AKO)/tests/servicesapitests -failfast -timeout 0 \
	-coverprofile cover-7.out -coverpkg=./... > servicesapitests.log 2>&1 && echo "servicesapitests passed") || (echo "servicesapitests failed" && cat servicesapitests.log && exit 1)

.PHONY: advl4tests
advl4tests:
	@> advl4tests.log
	(sudo docker run \
	-w=/go/src/$(PACKAGE_PATH_AKO) \
	-v $(PWD):/go/src/$(PACKAGE_PATH_AKO) $(GO_IMG_TEST) \
	$(GOTEST) -v -mod=vendor $(PACKAGE_PATH_AKO)/tests/advl4tests -failfast -timeout 0 \
	-coverprofile cover-8.out -coverpkg=./... > advl4tests.log 2>&1 && echo "advl4tests passed") || (echo "advl4tests failed" && cat advl4tests.log && exit 1)

.PHONY: namespacesynctests 
namespacesynctests:
	@> namespacesynctests.log
	(sudo docker run \
	-w=/go/src/$(PACKAGE_PATH_AKO) \
	-v $(PWD):/go/src/$(PACKAGE_PATH_AKO) $(GO_IMG_TEST) \
	$(GOTEST) -v -mod=vendor $(PACKAGE_PATH_AKO)/tests/namespacesynctests -failfast -timeout 0 \
	-coverprofile cover-9.out -coverpkg=./... > namespacesynctests.log 2>&1 && echo "namespacesynctests passed") || (echo "namespacesynctests failed" && cat namespacesynctests.log && exit 1)


.PHONY: npltests 
npltests:
	@> npltests.log
	(sudo docker run \
	-w=/go/src/$(PACKAGE_PATH_AKO) \
	-v $(PWD):/go/src/$(PACKAGE_PATH_AKO) $(GO_IMG_TEST) \
	$(GOTEST) -v -mod=vendor $(PACKAGE_PATH_AKO)/tests/npltests -failfast -timeout 0 \
	-coverprofile cover-10.out -coverpkg=./... > npltests.log 2>&1 && echo "npltests passed") || (echo "npltests failed" && cat npltests.log && exit 1)

.PHONY: evhtests 
evhtests:
	@> evhtests.log
	(sudo docker run \
	-w=/go/src/$(PACKAGE_PATH_AKO) \
	-v $(PWD):/go/src/$(PACKAGE_PATH_AKO) $(GO_IMG_TEST) \
	$(GOTEST) -v -mod=vendor $(PACKAGE_PATH_AKO)/tests/evhtests -failfast -timeout 0 \
	-coverprofile cover-11.out -coverpkg=./... > evhtests.log 2>&1 && echo "evhtests passed") || (echo "evhtests failed" && cat evhtests.log && exit 1)

.PHONY: vippernstests
vippernstests:
	@> vippernstests.log
	(sudo docker run \
	-w=/go/src/$(PACKAGE_PATH_AKO) \
	-v $(PWD):/go/src/$(PACKAGE_PATH_AKO) $(GO_IMG_TEST) \
	$(GOTEST) -v -mod=vendor $(PACKAGE_PATH_AKO)/tests/evhtests -failfast -timeout 0 -isVipPerNS=true \
	-coverprofile cover-12.out -coverpkg=./... > vippernstests.log 2>&1 && echo "vippernstests passed") || (echo "vippernstests failed" && cat vippernstests.log && exit 1)

.PHONY: dedicatedevhtests
dedicatedevhtests:
	@> dedicatedevhtests.log
	(sudo docker run \
	-w=/go/src/$(PACKAGE_PATH_AKO) \
	-v $(PWD):/go/src/$(PACKAGE_PATH_AKO) $(GO_IMG_TEST) \
	$(GOTEST) -v -mod=vendor $(PACKAGE_PATH_AKO)/tests/dedicatedevhtests -failfast -timeout 0 \
	-coverprofile cover-13.out -coverpkg=./... > dedicatedevhtests.log 2>&1 && echo "dedicatedevhtests passed") || (echo "dedicatedevhtests failed" && cat dedicatedevhtests.log && exit 1)

.PHONY: dedicatedvippernstests
dedicatedvippernstests:
	@> dedicatedvippernstests.log
	(sudo docker run \
	-w=/go/src/$(PACKAGE_PATH_AKO) \
	-v $(PWD):/go/src/$(PACKAGE_PATH_AKO) $(GO_IMG_TEST) \
	$(GOTEST) -v -mod=vendor $(PACKAGE_PATH_AKO)/tests/dedicatedevhtests -failfast -timeout 0 -isVipPerNS=true \
	-coverprofile cover-14.out -coverpkg=./... > dedicatedvippernstests.log 2>&1 && echo "dedicatedvippernstests passed") || (echo "dedicatedvippernstests failed" && cat dedicatedvippernstests.log && exit 1)

.PHONY: dedicatedvstests
dedicatedvstests:
	@> dedicatedvstests.log
	(sudo docker run \
	-w=/go/src/$(PACKAGE_PATH_AKO) \
	-v $(PWD):/go/src/$(PACKAGE_PATH_AKO) $(GO_IMG_TEST) \
	$(GOTEST) -v -mod=vendor $(PACKAGE_PATH_AKO)/tests/dedicatedvstests -failfast -timeout 0 \
	-coverprofile cover-15.out -coverpkg=./... > dedicatedvstests.log 2>&1 && echo "dedicatedvstests passed") || (echo "dedicatedvstests failed" && cat dedicatedvstests.log && exit 1)

.PHONY: infratests
infratests:
	@> infratests.log
	(sudo docker run \
	-w=/go/src/$(PACKAGE_PATH_AKO) \
	-v $(PWD):/go/src/$(PACKAGE_PATH_AKO) $(GO_IMG_TEST) \
	$(GOTEST) -v -mod=vendor $(PACKAGE_PATH_AKO)/tests/infratests -failfast -timeout 0 \
	-coverprofile cover-16.out -coverpkg=./... > infratests.log 2>&1 && echo "infratests passed") || (echo "infratests failed" && cat infratests.log && exit 1)

.PHONY: hatests
hatests:
	@> hatests.log
	(sudo docker run \
	-w=/go/src/$(PACKAGE_PATH_AKO) \
	-v $(PWD):/go/src/$(PACKAGE_PATH_AKO) $(GO_IMG_TEST) \
	$(GOTEST) -v -mod=vendor $(PACKAGE_PATH_AKO)/tests/hatests -failfast -timeout 0 \
	-coverprofile cover-17.out -coverpkg=./... > hatests.log 2>&1 && echo "hatests passed") || (echo "hatests failed" && cat hatests.log && exit 1)

.PHONY: calicotests
calicotests:
	@> calicotests.log
	(sudo docker run \
	-w=/go/src/$(PACKAGE_PATH_AKO) \
	-v $(PWD):/go/src/$(PACKAGE_PATH_AKO) $(GO_IMG_TEST) \
	$(GOTEST) -v -mod=vendor $(PACKAGE_PATH_AKO)/tests/cnitests -failfast -timeout 0 -cniPlugin=calico \
	-coverprofile cover-18.out -coverpkg=./... > calicotests.log 2>&1 && echo "calicotests passed") || (echo "calicotests failed" && cat calicotests.log && exit 1)

.PHONY: ciliumtests
ciliumtests:
	@> ciliumtests.log
	(sudo docker run \
	-w=/go/src/$(PACKAGE_PATH_AKO) \
	-v $(PWD):/go/src/$(PACKAGE_PATH_AKO) $(GO_IMG_TEST) \
	$(GOTEST) -v -mod=vendor $(PACKAGE_PATH_AKO)/tests/cnitests -failfast -timeout 0 -cniPlugin=cilium \
	-coverprofile cover-19.out -coverpkg=./... > ciliumtests.log 2>&1 && echo "ciliumtests passed") || (echo "ciliumtests failed" && cat ciliumtests.log && exit 1)

.PHONY: helmtests
helmtests:
	@> helmtests.log
	(sudo docker run \
	-u root:root \
	-v $(PWD)/helm/ako:/apps \
	-v $(PWD)/tests/helmtests:/apps/tests \
	avi-buildops-docker-registry-02-lv.avilb.broadcom.net:8080/avi-buildops/helmunittest/helm-unittest:3.11.1-0.3.0 . > helmtests.log 2>&1 && echo "helmtests passed") || (echo "helmtests failed" && cat helmtests.log && exit 1)

.PHONY: gatewayapi_ingestiontests
gatewayapi_ingestiontests:
	@> gatewayapi_ingestiontests.log
	(sudo docker run \
	-w=/go/src/$(PACKAGE_PATH_AKO) \
	-v $(PWD):/go/src/$(PACKAGE_PATH_AKO) $(GO_IMG_TEST) \
	$(GOTEST) -v -mod=vendor $(PACKAGE_PATH_AKO)/tests/gatewayapitests/ingestion -failfast -timeout 0 \
	-coverprofile cover-20.out -coverpkg=./ako-gateway-api/k8s/... > gatewayapi_ingestiontests.log 2>&1 && echo "gatewayapi_ingestiontests passed") || (echo "gatewayapi_ingestiontests failed" && cat gatewayapi_ingestiontests.log && exit 1)

.PHONY: gatewayapi_graphlayertests
gatewayapi_graphlayertests:
	@> gatewayapi_graphlayertests.log
	(sudo docker run \
	-w=/go/src/$(PACKAGE_PATH_AKO) \
	-v $(PWD):/go/src/$(PACKAGE_PATH_AKO) $(GO_IMG_TEST) \
	$(GOTEST) -v -mod=vendor $(PACKAGE_PATH_AKO)/tests/gatewayapitests/graphlayer -failfast -timeout 0 \
	-coverprofile cover-21.out -coverpkg=./ako-gateway-api/nodes/...,./ako-gateway-api/lib/...,./ako-gateway-api/objects/... > gatewayapi_graphlayertests.log 2>&1 && echo "gatewayapi_graphlayertests passed") || (echo "gatewayapi_graphlayertests failed" && cat gatewayapi_graphlayertests.log && exit 1)

.PHONY: gatewayapi_statustests
gatewayapi_statustests:
	@> gatewayapi_statustests.log
	(sudo docker run \
	-w=/go/src/$(PACKAGE_PATH_AKO) \
	-v $(PWD):/go/src/$(PACKAGE_PATH_AKO) $(GO_IMG_TEST) \
	$(GOTEST) -v -mod=vendor $(PACKAGE_PATH_AKO)/tests/gatewayapitests/status -failfast -timeout 0 \
	-coverprofile cover-22.out -coverpkg=./ako-gateway-api/status/...  > gatewayapi_statustests.log 2>&1 && echo "gatewayapi_statustests passed") || (echo "gatewayapi_statustests failed" && cat gatewayapi_statustests.log && exit 1)

.PHONY: gatewayapi_npltests
gatewayapi_npltests:
	@> gatewayapi_npltests.log
	(sudo docker run \
	-w=/go/src/$(PACKAGE_PATH_AKO) \
	-v $(PWD):/go/src/$(PACKAGE_PATH_AKO) $(GO_IMG_TEST) \
	$(GOTEST) -v -mod=vendor $(PACKAGE_PATH_AKO)/tests/gatewayapitests/npltests -failfast -timeout 0 \
	-coverprofile cover-23.out -coverpkg=./ako-gateway-api/...  > gatewayapi_npltests.log 2>&1 && echo "gatewayapi_npltests passed") || (echo "gatewayapi_npltests failed" && cat gatewayapi_npltests.log && exit 1)

.PHONY: gatewayapi_infrasettingtests
gatewayapi_infrasettingtests:
	@> gatewayapi_infrasettingtests.log
	(sudo docker run \
	-w=/go/src/$(PACKAGE_PATH_AKO) \
	-v $(PWD):/go/src/$(PACKAGE_PATH_AKO) $(GO_IMG_TEST) \
	$(GOTEST) -v -mod=vendor $(PACKAGE_PATH_AKO)/tests/gatewayapitests/crd -failfast -timeout 0 \
	-coverprofile cover-24.out -coverpkg=./ako-gateway-api/...  > gatewayapi_infrasettingtests.log 2>&1 && echo "gatewayapi_infrasettingtests passed") || (echo "gatewayapi_infrasettingtests failed" && cat gatewayapi_infrasettingtests.log && exit 1)

.PHONY: gatewayapi_multitenancytests
gatewayapi_multitenancytests:
	@> gatewayapi_multitenancytests.log
	(sudo docker run \
	-w=/go/src/$(PACKAGE_PATH_AKO) \
	-v $(PWD):/go/src/$(PACKAGE_PATH_AKO) $(GO_IMG_TEST) \
	$(GOTEST) -v -mod=vendor $(PACKAGE_PATH_AKO)/tests/gatewayapitests/multitenancy -failfast -timeout 0 \
	-coverprofile cover-25.out -coverpkg=./ako-gateway-api/...  > gatewayapi_multitenancytests.log 2>&1 && echo "gatewayapi_multitenancytests passed") || (echo "gatewayapi_multitenancytests failed" && cat gatewayapi_multitenancytests.log && exit 1)

.PHONY: multitenancytests
multitenancytests:
	@> multitenancytests.log
	(sudo docker run \
	-w=/go/src/$(PACKAGE_PATH_AKO) \
	-v $(PWD):/go/src/$(PACKAGE_PATH_AKO) $(GO_IMG_TEST) \
	$(GOTEST) -v -mod=vendor $(PACKAGE_PATH_AKO)/tests/multitenancytests -failfast -timeout 0 -coverprofile cover-24.out -coverpkg=./... > multitenancytests.log 2>&1 && echo "multitenancytests passed") || (echo "multitenancytests failed" && cat multitenancytests.log && exit 1)

.PHONY: urltests
urltests:
	@> urltests.log
	(sudo docker run \
	-w=/go/src/$(PACKAGE_PATH_AKO) \
	-v $(PWD):/go/src/$(PACKAGE_PATH_AKO) $(GO_IMG_TEST) \
	$(GOTEST) -v -mod=vendor $(PACKAGE_PATH_AKO)/tests/urltests -failfast -coverprofile cover-25.out -coverpkg=./... > urltests.log 2>&1 && echo "urltests passed") || (echo "urltests failed" && cat urltests.log && exit 1)

.PHONY: gatewayapi_tests
gatewayapi_tests:
	@> gatewayapi_tests.log
	(make -j 4 --output-sync=target gatewayapi_ingestiontests gatewayapi_graphlayertests gatewayapi_statustests gatewayapi_npltests gatewayapi_infrasettingtests gatewayapi_multitenancytests > gatewayapi_tests.log 2>&1 && echo "gatewayapi_tests passed") || (echo "gatewayapi_tests failed" && cat gatewayapi_tests.log && exit 1)

.PHONY: informers_tests
informers_tests:
	@> informers_tests.log
	(sudo docker run \
	-w=/go/src/$(PACKAGE_PATH_AKO) \
	-v $(PWD):/go/src/$(PACKAGE_PATH_AKO) $(GO_IMG_TEST) \
	$(GOTEST) -v -mod=vendor $(INFORMERS_PACKAGES)  -failfast -timeout 0 \
	-coverprofile cover-26.out -coverpkg=./... > informers_tests.log 2>&1 && echo "informers_tests passed") || (echo "informers_tests failed" && cat informers_tests.log && exit 1)

.PHONY: vks_addon_controller_tests
vks_addon_controller_tests:
	@> vks_addon_controller_tests.log
	(sudo docker run \
	-w=/go/src/$(PACKAGE_PATH_AKO) \
	-v $(PWD):/go/src/$(PACKAGE_PATH_AKO) $(GO_IMG_TEST) \
	$(GOTEST) -v -mod=vendor $(PACKAGE_PATH_AKO)/ako-infra/addon -failfast -timeout 0 \
	-coverprofile cover-27.out -coverpkg=./... > vks_addon_controller_tests.log 2>&1 && echo "vks_addon_controller_tests passed") || (echo "vks_addon_controller_tests failed" && cat vks_addon_controller_tests.log && exit 1)

.PHONY: vks_cluster_webhook_tests
vks_cluster_webhook_tests:
	@> vks_cluster_webhook_tests.log
	(sudo docker run \
	-w=/go/src/$(PACKAGE_PATH_AKO) \
	-v $(PWD):/go/src/$(PACKAGE_PATH_AKO) $(GO_IMG_TEST) \
	$(GOTEST) -v -mod=vendor $(PACKAGE_PATH_AKO)/ako-infra/webhook -failfast -timeout 0 \
	-coverprofile cover-28.out -coverpkg=./... > vks_cluster_webhook_tests.log 2>&1 && echo "vks_cluster_webhook_tests passed") || (echo "vks_cluster_webhook_tests failed" && cat vks_cluster_webhook_tests.log && exit 1)

.PHONY: vks_cluster_watcher_tests
vks_cluster_watcher_tests:
	@> vks_cluster_watcher_tests.log
	(sudo docker run \
	-w=/go/src/$(PACKAGE_PATH_AKO) \
	-v $(PWD):/go/src/$(PACKAGE_PATH_AKO) $(GO_IMG_TEST) \
	$(GOTEST) -v -mod=vendor $(PACKAGE_PATH_AKO)/ako-infra/ingestion -failfast -timeout 0 \
	-coverprofile cover-29.out -coverpkg=./... > vks_cluster_watcher_tests.log 2>&1 && echo "vks_cluster_watcher_tests passed") || (echo "vks_cluster_watcher_tests failed" && cat vks_cluster_watcher_tests.log && exit 1)

.PHONY: avi_rbac_tests
avi_rbac_tests:
	@> avi_rbac_tests.log
	(sudo docker run \
	-w=/go/src/$(PACKAGE_PATH_AKO) \
	-v $(PWD):/go/src/$(PACKAGE_PATH_AKO) $(GO_IMG_TEST) \
	$(GOTEST) -v -mod=vendor $(PACKAGE_PATH_AKO)/internal/lib -run "Test.*" -failfast -timeout 0 \
	-coverprofile cover-30.out -coverpkg=./... > avi_rbac_tests.log 2>&1 && echo "avi_rbac_tests passed") || (echo "avi_rbac_tests failed" && cat avi_rbac_tests.log && exit 1)

.PHONY: misc 
misc:
	@> misc.log
	(sudo docker run \
	-w=/go/src/$(PACKAGE_PATH_AKO) \
	-v $(PWD):/go/src/$(PACKAGE_PATH_AKO) $(GO_IMG_TEST) \
	$(GOTEST) -v -mod=vendor $(PACKAGE_PATH_AKO)/tests/miscellaneous -failfast -timeout 0 \
	-coverprofile cover-31.out -coverpkg=./... > misc.log 2>&1 && echo "misc passed") || (echo "misc failed" && cat misc.log && exit 1)

.PHONY: multiclusteringresstests
multiclusteringresstests:
	@> multiclusteringresstests.log
	(sudo docker run \
	-w=/go/src/$(PACKAGE_PATH_AKO) \
	-v $(PWD):/go/src/$(PACKAGE_PATH_AKO) $(GO_IMG_TEST) \
	$(GOTEST) -v -mod=vendor $(PACKAGE_PATH_AKO)/tests/multiclusteringresstests -failfast -coverprofile cover-32.out \
	-coverpkg=./... > multiclusteringresstests.log 2>&1 && echo "multiclusteringresstests passed") || (echo "multiclusteringresstests failed" && cat multiclusteringresstests.log && exit 1)

.PHONY: vks_tests
vks_tests:
	@> vks_tests.log
	(make -j 4 --output-sync=target vks_addon_controller_tests vks_cluster_webhook_tests vks_cluster_watcher_tests avi_rbac_tests > vks_tests.log 2>&1 && echo "vks_tests passed") || (echo "vks_tests failed" && cat vks_tests.log && exit 1)

.PHONY: int_test
int_test:
	@> int_test.log
	(make -j 8 --output-sync=target k8stest integrationtest ingresstests \
	evhtests vippernstests dedicatedevhtests dedicatedvippernstests \
	oshiftroutetests bootuptests multicloudtests advl4tests \
	namespacesynctests servicesapitests npltests misc \
	dedicatedvstests hatests calicotests ciliumtests \
	helmtests infratests urltests multitenancytests gatewayapi_ingestiontests gatewayapi_graphlayertests \
	gatewayapi_statustests gatewayapi_npltests gatewayapi_infrasettingtests gatewayapi_multitenancytests \
	informers_tests vks_tests > int_test.log 2>&1 \
	&& echo "int_test succeeded" && buffer -i int_test.log -u 1000 -z 1k) \
	|| (echo "int_test failed" && (buffer -i int_test.log -u 2000 -z 1b || \
	echo "Dumping the whole log failed; here are the last 100 lines" && tail -n100 int_test.log ) && exit 1)

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
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $@ v1.64.7

.PHONY: golangci
golangci: .golangci-bin
	@echo "Running golangci"
	@GOOS=linux GOGC=1 .golangci-bin/golangci-lint run -c .golangci.yml

.PHONY: golangci-fix
golangci-fix: .golangci-bin
	@echo "Running golangci-fix"
	@GOOS=linux GOGC=1 .golangci-bin/golangci-lint run -c .golangci.yml --fix
