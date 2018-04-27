PREFIX ?= /usr/local

TARGET              := prom-puppet-agent-exporter
TARGET_SRCS         := $(shell find . -type f -iname '*.go' -not -path './vendor/*')

GO                  := GO15VENDOREXPERIMENT=1 go
GIT_SUMMARY         := $(shell git describe --tags --always 2>/dev/null)
GIT_BRANCH          := $(shell git branch | sed -n -e 's/^\* \(.*\)/\1/p')
GO_VERSION          := $(shell $(GO) version)
GOPATH              := $(lastword $(subst :, ,$(GOPATH)))
DEP_BIN             :=  $(GOPATH)/bin/dep
LDFLAGS             := -ldflags "-X 'main.version=$(GIT_SUMMARY)' -X 'main.goVersion=$(GO_VERSION)' -X 'main.gitBranch=$(GIT_BRANCH)'"

.PHONY: all fmt vet lint test build install docker

all: test build

fmt:
	@echo ">> checking code style"
	@ret=0 && for d in $$($(GO) list -f '{{.Dir}}' ./... | grep -v /vendor/); do \
		$(GO) fmt $$d/*.go || ret=$$? ; \
		done ; exit $$ret

test: vendor
	@echo ">> running tests"
	@$(GO) test $(shell $(GO) list ./... | grep -v /vendor/)

vet: vendor
	@echo ">> vetting code"
	@$(GO) vet $(shell $(GO) list ./... | grep -v /vendor/)

imports:
	@echo ">> fixing source imports"
	@find . -type f -iname '*.go' -not -path './vendor/*' -not -iname '*pb.go' | xargs -L 1 goimports -w

lint:
	@echo ">> linting source"
	@find . -type f -iname '*.go' -not -path './vendor/*' -not -iname '*pb.go' | xargs -L 1 golint

build: $(TARGET)

$(TARGET): $(TARGET_SRCS) vendor
	@echo ">> building binary..."
	@echo ">> GOOS=$(GOOS) GOARCH=$(GOARCH) $(GO) build $(LDFLAGS) -o $(TARGET)"
	@GOOS=$(GOOS) GOARCH=$(GOARCH) $(GO) build $(LDFLAGS) -o $(TARGET)

docker: GOOS="linux" GOARCH="amd64"
docker: DOCKER_IMAGE_NAME ?= "monitoring-tools/prom-puppet-agent-exporter:$(GIT_SUMMARY)"
docker: Dockerfile build
	@echo ">> building docker image"
	@docker build -t $(DOCKER_IMAGE_NAME) $(DOCKER_BUILD_ARGS) .

install: build
	install -m 0755 -d $(DESTDIR)$(PREFIX)/bin
	install -m 0755 $(TARGET) $(DESTDIR)$(PREFIX)/bin

vendor: $(DEP_BIN) Gopkg.lock
	@echo ">> installing golang dependencies into vendor directory..."
	@$(DEP_BIN) ensure

$(DEP_BIN):
	@echo "Installing golang dependency manager..."
	@$(GO) get -u github.com/golang/dep/cmd/dep

clean:
	@echo ">> removing binary"
	@rm -rf $(TARGET)
	@echo ">> removing vendor directory"
	@rm -rf vendor
