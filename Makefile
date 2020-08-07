GOCMD=/usr/local/go/bin/go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOGET=$(GOCMD) get
GOTEST=$(GOCMD) test
BINARY_NAME_AKO=ako
REL_PATH_AKO=github.com/avinetworks/ako/cmd/ako-main


.PHONY:all
all: build docker

.PHONY: build
build: 
		$(GOBUILD) -o bin/$(BINARY_NAME_AKO)  -mod=vendor $(REL_PATH_AKO)

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
	sudo docker build -t $(BINARY_NAME_AKO):latest --label "BUILD_TAG=$(BUILD_TAG)" --label "BUILD_TIME=$(BUILD_TIME)" -f Dockerfile.ako .


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

.PHONY: int_test
int_test:
	make -j 3 integrationtest hostnameshardtests oshiftroutetests

.PHONY: scale_test
scale_test:
	$(GOTEST) -mod=vendor -v ./tests/scaletest -failfast

