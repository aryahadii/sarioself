ROOT := github.com/aryahadii/sarioself
GO_VARS ?= CGO_ENABLED=1 GOOS=linux GOARCH=amd64
GO ?= go
GIT ?= git
COMMIT := $(shell $(GIT) rev-parse HEAD)
VERSION ?= $(shell $(GIT) describe --tags ${COMMIT} 2> /dev/null || echo "$(COMMIT)")
BUILD_TIME := $(shell LANG=en_US date +"%F_%T_%z")
LD_FLAGS := -X $(ROOT).Version=$(VERSION) -X $(ROOT).Commit=$(COMMIT) -X $(ROOT).BuildTime=$(BUILD_TIME) -X $(ROOT).Title=sarioself
DOCKER_IMAGE := registry.gitlab.com/arha/sarioself

.PHONY: help clean update-dependencies dependencies docker push

sarioselfd: *.go */*.go */*/*.go Gopkg.lock
	$(GO_VARS) $(GO) build -o="sarioselfd" -ldflags="$(LD_FLAGS)" $(ROOT)/cmd/sarioself

help:
	@echo "Please use \`make <ROOT>' where <ROOT> is one of"
	@echo "  update-dependencies    to update Gopkg.lock (refs to dependencies)"
	@echo "  dependencies           to install the dependencies"
	@echo "  docker     	        to create docker image"
	@echo "  push        	        to push docker image to registry"
	@echo "  sarioselfd             to build the binary"
	@echo "  clean                  to remove generated files"

clean:
	rm -rf sarioselfd

update-dependencies:
	dep ensure -update

dependencies:
	dep ensure

docker: sarioselfd Dockerfile
	docker build -t $(DOCKER_IMAGE):$(VERSION) .
	docker tag $(DOCKER_IMAGE):$(VERSION) $(DOCKER_IMAGE):latest

push:
	docker push $(DOCKER_IMAGE):$(VERSION)
	docker push $(DOCKER_IMAGE):latest
