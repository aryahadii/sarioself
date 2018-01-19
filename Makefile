ROOT := github.com/aryahadii/sarioself
GO_VARS ?= CGO_ENABLED=0 GOOS=darwin GOARCH=amd64
GO ?= go
GIT ?= git
COMMIT := $(shell $(GIT) rev-parse HEAD)
VERSION ?= $(shell $(GIT) describe --tags ${COMMIT} 2> /dev/null || echo "$(COMMIT)")
BUILD_TIME := $(shell LANG=en_US date +"%F_%T_%z")
LD_FLAGS := -X $(ROOT).Version=$(VERSION) -X $(ROOT).Commit=$(COMMIT) -X $(ROOT).BuildTime=$(BUILD_TIME) -X $(ROOT).Title=sarioself

.PHONY: help clean update-dependencies dependencies

sarioselfd: *.go */*.go */*/*.go glide.lock
	$(GO_VARS) $(GO) build -o="sarioselfd" -ldflags="$(LD_FLAGS)" $(ROOT)/cmd/sarioself

help:
	@echo "Please use \`make <ROOT>' where <ROOT> is one of"
	@echo "  update-dependencies    to update glide.lock (refs to dependencies)"
	@echo "  dependencies           to install the dependencies"
	@echo "  sarioselfd             to build the binary"
	@echo "  clean                  to remove generated files"

clean:
	rm -rf sarioselfd

update-dependencies:
	glide up

dependencies:
	glide install
