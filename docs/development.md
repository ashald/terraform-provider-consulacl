# Development

## Go

In order to work on the provider, [Go](http://www.golang.org) should be installed first (version 1.8+ is *required*).
[goenv](https://github.com/syndbg/goenv) and [gvm](https://github.com/moovweb/gvm) are great utilities that can help a
lot with that and simplify setup tremendously. 
[GOPATH](http://golang.org/doc/code.html#GOPATH) should be setup correctly and as long as `$GOPATH/bin` should be
added `$PATH`.

## Source Code

Source code can be retrieved either with `go get`

```bash
$ go get -u -d github.com/ashald/terraform-provider-consulacl
```

or with `git`
```bash
$ mkdir -p ${GOPATH}/src/github.com/ashald/terraform-provider-consulacl
$ cd ${GOPATH}/src/github.com/ashald/terraform-provider-consulacl
$ git clone git@github.com:ashald/terraform-provider-consulacl.git .
```

## Dependencies

This project uses `govendor` to manage its dependencies. When adding a dependency on a new package it should be fetched
with:
```bash
$ govendor fetch +o
```

## Test

### Unit Tests

```bash
$ make test
  go test -v ./...
  ?   	github.com/ashald/terraform-provider-consulacl	[no test files]
  === RUN   TestIntegrationDataSourceToken
  --- SKIP: TestIntegrationDataSourceToken (0.00s)
      testing.go:461: Acceptance tests skipped unless env 'TF_ACC' set
  === RUN   TestProvider
  --- PASS: TestProvider (0.00s)
  === RUN   TestIntegrationToken
  --- SKIP: TestIntegrationToken (0.00s)
      testing.go:461: Acceptance tests skipped unless env 'TF_ACC' set
  === RUN   TestIntegrationTokenImport
  --- SKIP: TestIntegrationTokenImport (0.00s)
      testing.go:461: Acceptance tests skipped unless env 'TF_ACC' set
  PASS
  ok  	github.com/ashald/terraform-provider-consulacl/consulacl	0.017s
  go vet ./...
```

### Integration Tests

This requires a running Consul agent locally.

```bash
$ CONSUL_TOKEN=secret make test-integration
  TF_ACC=1 go test -v ./... -timeout 1m
  ?   	github.com/ashald/terraform-provider-consulacl	[no test files]
  === RUN   TestIntegrationDataSourceToken
  --- PASS: TestIntegrationDataSourceToken (0.03s)
  === RUN   TestProvider
  --- PASS: TestProvider (0.00s)
  === RUN   TestIntegrationToken
  --- PASS: TestIntegrationToken (0.18s)
  === RUN   TestIntegrationTokenImport
  --- PASS: TestIntegrationTokenImport (0.04s)
  PASS
  ok  	github.com/ashald/terraform-provider-consulacl/consulacl	(cached)
```

If you have [Docker](https://docs.docker.com/install/) installed, you can run Consul with the following command:
```bash
$ make test-server
  latest: Pulling from library/consul
  Digest: sha256:ae2c9409a77533485982c00f5c1eab89c090889318cb2f4276d64a7d125f83f8
  Status: Image is up to date for consul:latest
  docker run --rm -p 127.0.0.1:8500:8500 -e CONSUL_LOCAL_CONFIG='{"acl_datacenter": "dc1", "acl_master_token": "secret", "bootstrap_expect": 1, "server": true, "ui": true}' 'consul:latest'
  ...
```

By default, this will use the
[latest version of Consul based on the latest image in the Docker repository](https://hub.docker.com/_/consul/).
You can specify a version using `CONSUL_VERSION` environment variable:
```bash
$ CONSUL_VERSION=1.2.0 make test-server
```

This command will run in foreground and will stop Consul when interrupted.
Images will be cached locally by Docker so it is quick to restart the server as necessary.
This will expose Consul on the default address `127.0.0.1:8500` but this can be changed with `CONSUL_ADDRESS`
environment variable.

## Build
In order to build plugin for the current platform use [GNU]make:
```bash
$ make build
  go build -o terraform-provider-consulacl_v1.2.0

```

it will build provider from sources and put it into current working directory.

If Terraform was installed (as a binary) or via `go get -u github.com/hashicorp/terraform` it'll pick up the plugin if 
executed against a configuration in the same directory.

## Release

In order to prepare provider binaries for all platforms:
```bash
$ make release
  GOOS=darwin GOARCH=amd64 go build -o './release/terraform-provider-consulacl_v1.2.0-darwin-amd64'
  GOOS=linux GOARCH=amd64 go build -o './release/terraform-provider-consulacl_v1.2.0-linux-amd64'
```

## Versioning

This project follow [Semantic Versioning](https://semver.org/)

## Changelog

This project follows [keep a changelog](https://keepachangelog.com/en/1.0.0/) guidelines for changelog.