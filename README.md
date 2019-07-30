# Terraform Consul ACL Provider

## Overview

This provider defines Terraform resources and data sources related to Consul ACL subsystem that are missing from the official one.

**PLEASE NOTE THAT USING THIS PROVIDER WOULD EXPOSE SENSITIVE TOKEN ID VALUES IN YOUR STATE.**

### Resources:  
* [resource "consulacl_token"](./docs/resource_consulacl_token.md) - manages a single Consul ACL token (legacy API, pre Consul 1.4)
* [resource "consulacl_policy_binding"](./docs/resource_consulacl_policy_binding.md) - manages bindings between post-Consul 1.4 ACL policies and tokens by their accessor IDs

### Data Sources:
* [data "consulacl_token"](./docs/data_source_consulacl_token.md) - retrieves post-Consul 1.4 ACL token's secret ID by its accessor ID

## Installation

> Terraform automatically discovers the Providers when it parses configuration files.
> This only occurs when the init command is executed.

Currently Terraform is able to automatically download only
[official plugins distributed by HashiCorp](https://github.com/terraform-providers).

[All other plugins](https://www.terraform.io/docs/providers/type/community-index.html) should be installed manually.

> Terraform will search for matching Providers via a
> [Discovery](https://www.terraform.io/docs/extend/how-terraform-works.html#discovery) process, **including the current
> local directory**.

This means that the plugin should either be placed into current working directory where Terraform will be executed from
or it can be [installed system-wide](https://www.terraform.io/docs/configuration/providers.html#third-party-plugins).

The simplest way to get started is:
```bash
wget "https://github.com/ashald/terraform-provider-consulacl/releases/download/1.3.0/terraform-provider-consulacl_v1.3.0-$(uname -s | tr '[:upper:]' '[:lower:]')-amd64"
chmod +x ./terraform-provider-consulacl*
```

## Configuration

Provider is configurable with number of parameters:

```hcl
provider "consulacl" {
  // Host and port used to connect to Consul.
  // Can be set via environment variables `CONSUL_ADDRESS` or `CONSUL_HTTP_ADDR`. 
  address = "localhost:8500"
  
  // ACL token to use for API calls to Consul. Must be a `management` token to manage ACLs.
  // Can be set via environment variables `CONSUL_TOKEN` or `CONSUL_HTTP_TOKEN`.
  token = ""
  
  // Scheme to use to connect to Consul.
  // Can be set via environment variables `CONSUL_SCHEME` or `CONSUL_HTTP_SCHEME`.
  scheme = "http" // Only "http" and "https" are supported.
  
  // Path to a certificate of a certification authority (CA) that was used to sign Consul's TLS
  // certificate and therefore should be used for TLS validation.
  // Can be set via environment variable `CONSUL_CA_FILE`.
  ca_file = "" // Empty value means use system bundle.
  
  // Path to a client certificate for client-side TLS authentication, if enabled in Consul.
  // Can be set via environment variable `CONSUL_CERT_FILE`.
  cert_file = ""
  
  // Path to a private key for client certificate provided in `cert_file`.
  // Can be set via environment variable `CONSUL_KEY_FILE`.
  key_file = ""
  
  // Whether to skip verification of Consul's TLS certificate.
  // Can be set via environment variable `CONSUL_TLS_SKIP_VERIFY`.
  tls_skip_verify = false
}
``` 

## Development

Provider is written and maintained by [Borys Pierov](https://github.com/Ashald).
Contributions are welcome and should follow [development guidelines](./docs/development.md).
All contributors are honored in [CONTRIBUTORS.md](./CONTRIBUTORS.md).

## License

This is free and unencumbered software released into the public domain. See [LICENSE](./LICENSE)
