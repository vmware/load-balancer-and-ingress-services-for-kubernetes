GOCMD=/usr/local/go/bin/go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOGET=$(GOCMD) get
BINARY_NAME_AMC=ako
REL_PATH_AMC=ako/cmd/akc-main


.PHONY:all
all: build docker

.PHONY: build
build: 
		$(GOBUILD) -o bin/$(BINARY_NAME_AMC)  -mod=vendor $(REL_PATH_AMC)

.PHONY: clean
clean: 
		$(GOCLEAN)
		rm -f $(BINARY_NAME)

.PHONY: deps
deps:
	dep ensure -v

.PHONY: docker
docker:
ifndef BUILD_TAG
	$(eval BUILD_TAG=$(shell ./tests/jenkins/get_build_version.sh "dummy" 0))
endif

ifndef BUILD_TIME
	$(eval BUILD_TIME=$(shell date +%Y-%m-%d_%H:%M:%S_%Z))
endif

	sudo docker build -t $(BINARY_NAME_AMC):latest --label "BUILD_TAG=$(BUILD_TAG)" --label "BUILD_TIME=$(BUILD_TIME)" -f Dockerfile.ako .


.PHONY: test
test:
	/usr/local/go/bin/go test -mod=vendor -v ./pkg/k8s -failfast
.PHONY: int_test
int_test:
	/usr/local/go/bin/go test -mod=vendor -v ./tests/integrationtest -failfast
	/usr/local/go/bin/go test -mod=vendor -v ./tests/hostnameshardtests -failfast
