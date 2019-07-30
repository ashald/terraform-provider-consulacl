NAME := terraform-provider-consulacl
PLATFORMS ?= darwin/amd64 linux/amd64 windows/amd64
VERSION = $(shell git describe 1>/dev/null 2>/dev/null && echo "_$$(git describe)")

temp = $(subst /, ,$@)
os = $(word 1, $(temp))
arch = $(word 2, $(temp))

BASE := $(NAME)$(VERSION)
RELEASE_DIR := ./release

CONSUL_ADDRESS ?= 127.0.0.1:8500
CONSUL_LOCAL_CONFIG ?= {"acl_datacenter": "dc1", "acl_master_token": "secret", "bootstrap_expect": 1, "server": true, "ui": true}
CONSUL_VERSION ?= latest
CONSUL_TOKEN ?= secret

.PHONY: all
all: clean test release

.PHONY: clean
clean:
	rm -rf $(RELEASE_DIR) ./$(NAME)*

.PHONY: format
format:
	GOPROXY="off" GOFLAGS="-mod=vendor" go fmt ./...

.PHONY: test
test:
	GOPROXY="off" GOFLAGS="-mod=vendor" go test -v ./...
	GOPROXY="off" GOFLAGS="-mod=vendor" go vet ./...

.PHONY: test-server
test-server:
	@docker pull 'consul:$(CONSUL_VERSION)'
	docker run --rm -p $(CONSUL_ADDRESS):8500 -e CONSUL_LOCAL_CONFIG='$(CONSUL_LOCAL_CONFIG)' 'consul:$(CONSUL_VERSION)'

.PHONY: test-integration
test-integration:
	TF_ACC=1 CONSUL_TOKEN=$(CONSUL_TOKEN) go test -v ./... -timeout 1m

.PHONY: build
build:
	GOPROXY="off" GOFLAGS="-mod=vendor" go build -o $(BASE)

.PHONY:
release: $(PLATFORMS)

.PHONY: $(PLATFORMS)
$(PLATFORMS):
	GOPROXY="off" GOFLAGS="-mod=vendor" GOOS=$(os) GOARCH=$(arch) go build -o '$(RELEASE_DIR)/$(BASE)-$(os)-$(arch)'

.PHONY: compress
compress:
	upx $(RELEASE_DIR)/*

.PHONY: sums
sums:
	cd $(RELEASE_DIR); shasum -a 256 para* > SHA256SUMS
