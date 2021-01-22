GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOGET=$(GOCMD) get
GOTEST=$(GOCMD) test
BINARY_NAME_AKO=ako
AKO_VERSION=v1.3.1
PACKAGE_PATH_AKO=github.com/vmware/load-balancer-and-ingress-services-for-kubernetes
REL_PATH_AKO=$(PACKAGE_PATH_AKO)/cmd/ako-main
AKO_OPERATOR_IMAGE=ako-operator


ifdef GOLANG_SRC_REPO
	BUILD_GO_IMG=$(GOLANG_SRC_REPO)
else
	BUILD_GO_IMG=golang:latest
endif

.PHONY: glob-vars
glob-vars:
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
		sudo docker run -w=/go/src/$(PACKAGE_PATH_AKO) -v $(PWD):/go/src/$(PACKAGE_PATH_AKO) $(BUILD_GO_IMG) \
		$(GOBUILD) -o /go/src/$(PACKAGE_PATH_AKO)/bin/$(BINARY_NAME_AKO) -ldflags="-X 'main.version=$(AKO_VERSION)'" -mod=vendor /go/src/$(REL_PATH_AKO)

.PHONY: build-local
build-local:
		$(GOBUILD) -o bin/$(BINARY_NAME_AKO) -ldflags="-X 'main.version=$(AKO_VERSION)'" -mod=vendor ./cmd/ako-main

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
	sudo docker build -t $(BINARY_NAME_AKO):latest --label "BUILD_TAG=$(BUILD_TAG)" --label "BUILD_TIME=$(BUILD_TIME)" $(BUILD_ARG_GOLANG) $(BUILD_ARG_PHOTON) -f Dockerfile.ako .

.PHONY: ako-operator-docker
ako-operator-docker: glob-vars
	sudo docker build -t $(AKO_OPERATOR_IMAGE):latest --label "BUILD_TAG=$(BUILD_TAG)" --label "BUILD_TIME=$(BUILD_TIME)" $(BUILD_ARG_GOLANG) $(BUILD_ARG_UBI)  -f Dockerfile.ako-operator .

# tests
.PHONY: k8stest
k8stest:
	sudo docker run -w=/go/src/$(PACKAGE_PATH_AKO) -v $(PWD):/go/src/$(PACKAGE_PATH_AKO) $(BUILD_GO_IMG) \
	$(GOTEST) -v -mod=vendor $(PACKAGE_PATH_AKO)/tests/k8stest -failfast

.PHONY: integrationtest
integrationtest:
	sudo docker run -w=/go/src/$(PACKAGE_PATH_AKO) -v $(PWD):/go/src/$(PACKAGE_PATH_AKO) $(BUILD_GO_IMG) \
	$(GOTEST) -v -mod=vendor $(PACKAGE_PATH_AKO)/tests/integrationtest -failfast

.PHONY: hostnameshardtests
hostnameshardtests:
	sudo docker run -w=/go/src/$(PACKAGE_PATH_AKO) -v $(PWD):/go/src/$(PACKAGE_PATH_AKO) $(BUILD_GO_IMG) \
	$(GOTEST) -v -mod=vendor $(PACKAGE_PATH_AKO)/tests/hostnameshardtests -failfast

.PHONY: oshiftroutetests
oshiftroutetests:
	sudo docker run -w=/go/src/$(PACKAGE_PATH_AKO) -v $(PWD):/go/src/$(PACKAGE_PATH_AKO) $(BUILD_GO_IMG) \
	$(GOTEST) -v -mod=vendor $(PACKAGE_PATH_AKO)/tests/oshiftroutetests -failfast -timeout 0

.PHONY: bootuptests
bootuptests:
	sudo docker run -w=/go/src/$(PACKAGE_PATH_AKO) -v $(PWD):/go/src/$(PACKAGE_PATH_AKO) $(BUILD_GO_IMG) \
	$(GOTEST) -v -mod=vendor $(PACKAGE_PATH_AKO)/tests/bootuptests -failfast

.PHONY: multicloudtests
multicloudtests:
	sudo docker run -w=/go/src/$(PACKAGE_PATH_AKO) -v $(PWD):/go/src/$(PACKAGE_PATH_AKO) $(BUILD_GO_IMG) \
	$(GOTEST) -v -mod=vendor $(PACKAGE_PATH_AKO)/tests/multicloudtests -failfast

.PHONY: servicesapitests
servicesapitests:
	sudo docker run -w=/go/src/$(PACKAGE_PATH_AKO) -v $(PWD):/go/src/$(PACKAGE_PATH_AKO) $(BUILD_GO_IMG) \
	$(GOTEST) -v -mod=vendor $(PACKAGE_PATH_AKO)/tests/servicesapitests -failfast

.PHONY: advl4tests
advl4tests:
	sudo docker run -w=/go/src/$(PACKAGE_PATH_AKO) -v $(PWD):/go/src/$(PACKAGE_PATH_AKO) $(BUILD_GO_IMG) \
	$(GOTEST) -v -mod=vendor $(PACKAGE_PATH_AKO)/tests/advl4tests -failfast

.PHONY: namespacesynctests 
namespacesynctests:
	sudo docker run -w=/go/src/$(PACKAGE_PATH_AKO) -v $(PWD):/go/src/$(PACKAGE_PATH_AKO) $(BUILD_GO_IMG) \
	$(GOTEST) -v -mod=vendor $(PACKAGE_PATH_AKO)/tests/namespacesynctests -failfast

.PHONY: int_test
int_test:
	make -j 1 k8stest integrationtest hostnameshardtests oshiftroutetests bootuptests multicloudtests advl4tests namespacesynctests servicesapitests

.PHONY: scale_test
scale_test:
	sudo docker run -w=/go/src/$(PACKAGE_PATH_AKO) -v $(PWD):/go/src/$(PACKAGE_PATH_AKO) $(BUILD_GO_IMG) \
	$(GOTEST) -v -mod=vendor ./tests/scaletest -failfast $(Timeout) $(TestbedFilePath) $(NumGoRoutines) $(NumOfLBSvc) $(NumOfIng)

# linting and formatting
GO_FILES := $(shell find . -type d -path ./vendor -prune -o -type f -name '*.go' -print)
.PHONY: fmt
fmt:
	@echo
	@echo "Formatting Go files"
	@gofmt -s -l -w $(GO_FILES)

.golangci-bin:
	@echo "Installing Golangci-lint"
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $@ v1.32.1

.PHONY: golangci
golangci: .golangci-bin
	@echo "Running golangci"
	@GOOS=linux .golangci-bin/golangci-lint run -c .golangci.yml

.PHONY: golangci-fix
golangci-fix: .golangci-bin
	@echo "Running golangci-fix"
	@GOOS=linux .golangci-bin/golangci-lint run -c .golangci.yml --fix

