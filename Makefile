GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOGET=$(GOCMD) get
GOTEST=$(GOCMD) test
BINARY_NAME_AKO=ako
AKO_VERSION=v1.3.1
REL_PATH_AKO=github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/cmd/ako-main


.PHONY:all
all: build docker

.PHONY: build
build: 
		$(GOBUILD) -o bin/$(BINARY_NAME_AKO) -ldflags="-X 'main.version=$(AKO_VERSION)'"  -mod=vendor $(REL_PATH_AKO)

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
	$(GOTEST) -mod=vendor -v ./tests/k8stest -failfast

.PHONY: integrationtest
integrationtest:
	$(GOTEST) -mod=vendor -v ./tests/integrationtest -failfast

.PHONY: hostnameshardtests
hostnameshardtests:
	$(GOTEST) -mod=vendor -v ./tests/hostnameshardtests -failfast

.PHONY: oshiftroutetests
oshiftroutetests:
	$(GOTEST) -mod=vendor -v ./tests/oshiftroutetests -failfast

.PHONY: bootuptests
bootuptests:
	$(GOTEST) -mod=vendor -v ./tests/bootuptests -failfast

.PHONY: multicloudtests
multicloudtests:
	$(GOTEST) -mod=vendor -v ./tests/multicloudtests -failfast

.PHONY: advl4tests
advl4tests:
	$(GOTEST) -mod=vendor -v ./tests/advl4tests -failfast

.PHONY: int_test
int_test:
	make -j 1 integrationtest hostnameshardtests oshiftroutetests bootuptests multicloudtests advl4tests

.PHONY: scale_test
scale_test:
	$(GOTEST) -mod=vendor -v ./tests/scaletest -failfast

