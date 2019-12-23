GOCMD=/usr/local/go/bin/go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOGET=$(GOCMD) get
BINARY_NAME_AMC=avi-k8s-controller
REL_PATH_AMC=gitlab.eng.vmware.com/orion/akc/cmd/akc-main

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
	docker build -t $(BINARY_NAME_AMC):latest -f Dockerfile.akc .

.PHONY: test
test:
	/usr/local/go/bin/go test -v ./pkg/k8s
.PHONY: int_test
int_test:
	/usr/local/go/bin/go test -v ./integrationtest
