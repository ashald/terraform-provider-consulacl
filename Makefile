NAME := terraform-provider-consulacl
PLATFORMS ?= darwin/amd64 linux/amd64
VERSION ?= $(shell git describe &>/dev/null && echo "_$$(git describe)")

temp = $(subst /, ,$@)
os = $(word 1, $(temp))
arch = $(word 2, $(temp))

BASE := $(NAME)$(VERSION)
RELEASE_DIR := ./release

CONSUL_ADDRESS ?= 127.0.0.1:8500
CONSUL_LOCAL_CONFIG ?= {"acl_datacenter": "dc1", "acl_master_token": "secret", "bootstrap_expect": 1, "server": true, "ui": true}
CONSUL_VERSION ?= latest
CONSUL_TOKEN ?= secret

all: clean format test release

clean:
	rm -rf $(RELEASE_DIR) ./$(NAME)*

format:
	go fmt ./...

test:
	go test -v ./...
	go vet ./...

test-server:
	@docker pull 'consul:$(CONSUL_VERSION)'
	docker run --rm -p $(CONSUL_ADDRESS):8500 -e CONSUL_LOCAL_CONFIG='$(CONSUL_LOCAL_CONFIG)' 'consul:$(CONSUL_VERSION)'

test-integration:
	TF_ACC=1 go test -v ./... -timeout 1m

build:
	go build -o $(BASE)

release: $(PLATFORMS)

$(PLATFORMS):
	GOOS=$(os) GOARCH=$(arch) go build -o '$(RELEASE_DIR)/$(BASE)-$(os)-$(arch)'

.PHONY: $(PLATFORMS) release build test fmt clean all
