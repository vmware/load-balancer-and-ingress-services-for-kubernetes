GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOGET=$(GOCMD) get
GOTEST=$(GOCMD) test
BINARY_NAME_AKO=ako
AKO_VERSION=v1.3.2
PACKAGE_PATH_AKO=github.com/vmware/load-balancer-and-ingress-services-for-kubernetes
REL_PATH_AKO=$(PACKAGE_PATH_AKO)/cmd/ako-main

ifdef GOLANG_SRC_REPO
	BUILD_GO_IMG=$(GOLANG_SRC_REPO)
else
	BUILD_GO_IMG=golang:latest
endif

.PHONY:all
all: build docker

.PHONY: build
build: 
		sudo docker run -v $(PWD):/go/src/$(PACKAGE_PATH_AKO) $(BUILD_GO_IMG) \
		$(GOBUILD) -o /go/src/$(PACKAGE_PATH_AKO)/bin/$(BINARY_NAME_AKO) -ldflags="-X 'main.version=$(AKO_VERSION)'" /go/src/$(REL_PATH_AKO)

.PHONY: clean
clean: 
		$(GOCLEAN) -mod=vendor $(REL_PATH_AKO)
		rm -f bin/$(BINARY_NAME_AKO)

.PHONY: deps
deps:
	dep ensure -v

.PHONY: docker
docker:
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
	sudo docker build -t $(BINARY_NAME_AKO):latest --label "BUILD_TAG=$(BUILD_TAG)" --label "BUILD_TIME=$(BUILD_TIME)" $(BUILD_ARG_GOLANG) $(BUILD_ARG_PHOTON) -f Dockerfile.ako .

.PHONY: test
test:
	sudo docker run -v $(PWD):/go/src/$(PACKAGE_PATH_AKO) $(BUILD_GO_IMG) \
	$(GOTEST) -v $(PACKAGE_PATH_AKO)/tests/k8stest -failfast

.PHONY: integrationtest
integrationtest:
	sudo docker run -v $(PWD):/go/src/$(PACKAGE_PATH_AKO) $(BUILD_GO_IMG) \
	$(GOTEST) -v $(PACKAGE_PATH_AKO)/tests/integrationtest -failfast

.PHONY: hostnameshardtests
hostnameshardtests:
	sudo docker run -v $(PWD):/go/src/$(PACKAGE_PATH_AKO) $(BUILD_GO_IMG) \
	$(GOTEST) -v $(PACKAGE_PATH_AKO)/tests/hostnameshardtests -failfast

.PHONY: oshiftroutetests
oshiftroutetests:
	sudo docker run -v $(PWD):/go/src/$(PACKAGE_PATH_AKO) $(BUILD_GO_IMG) \
	$(GOTEST) -v $(PACKAGE_PATH_AKO)/tests/oshiftroutetests -failfast

.PHONY: bootuptests
bootuptests:
	sudo docker run -v $(PWD):/go/src/$(PACKAGE_PATH_AKO) $(BUILD_GO_IMG) \
	$(GOTEST) -v $(PACKAGE_PATH_AKO)/tests/bootuptests -failfast

.PHONY: multicloudtests
multicloudtests:
	sudo docker run -v $(PWD):/go/src/$(PACKAGE_PATH_AKO) $(BUILD_GO_IMG) \
	$(GOTEST) -v $(PACKAGE_PATH_AKO)/tests/multicloudtests -failfast

.PHONY: advl4tests
advl4tests:
	sudo docker run -v $(PWD):/go/src/$(PACKAGE_PATH_AKO) $(BUILD_GO_IMG) \
	$(GOTEST) -v $(PACKAGE_PATH_AKO)/tests/advl4tests -failfast

.PHONY: int_test
int_test:
	sudo docker run -v $(PWD):/go/src/$(PACKAGE_PATH_AKO) $(BUILD_GO_IMG) \
	$(GOTEST) -v $(PACKAGE_PATH_AKO)/tests/k8stest -failfast && \
	$(GOTEST) -v $(PACKAGE_PATH_AKO)/tests/integrationtest -failfast && \
	$(GOTEST) -v $(PACKAGE_PATH_AKO)/tests/hostnameshardtests -failfast && \
	$(GOTEST) -v $(PACKAGE_PATH_AKO)/tests/oshiftroutetests -failfast && \
	$(GOTEST) -v $(PACKAGE_PATH_AKO)/tests/bootuptests -failfast && \
	$(GOTEST) -v $(PACKAGE_PATH_AKO)/tests/multicloudtests -failfast && \
	$(GOTEST) -v $(PACKAGE_PATH_AKO)/tests/advl4tests -failfast

.PHONY: scale_test
scale_test:
	sudo docker run -v $(PWD):/go/src/$(PACKAGE_PATH_AKO) $(BUILD_GO_IMG) \
	$(GOTEST) -mod=vendor -v ./tests/scaletest -failfast -timeout $(Timeout) $(NumGoRoutines) $(TestbedFilePath)
