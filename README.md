# Terraform Consul ACL Provider

## Overview

This provider defines Terraform resources and data sources related to Consul ACL subsystem that are missing from the
official one.

**PLEASE NOTE THAT USING THIS PROVIDER WOULD EXPOSE SENSITIVE TOKEN ID VALUES IN YOUR STATE.**

### Resources:  
* [resource "consulacl_token"](./docs/resource_consulacl_token.md) - manages a single Consul ACL token (legacy API, pre
Consul 1.4)
* [resource "consulacl_token14"](./docs/resource_consulacl_token14.md) - manages a single post-Consul 1.4 ACL token
(like the official one, but allows setting `accessor` and/or `secret`)
* [resource "consulacl_policy_binding"](./docs/resource_consulacl_policy_binding.md) - manages bindings between
post-Consul 1.4 ACL policies and tokens by their accessor IDs

### Data Sources:
* [data "consulacl_token"](./docs/data_source_consulacl_token.md) - retrieves post-Consul 1.4 ACL token's secret ID by
its accessor ID

## Installation

> Terraform automatically discovers the Providers when it parses configuration files.
> This only occurs when the init command is executed.

Currently Terraform is able to automatically download only [official plugins distributed by HashiCorp](https://github.com/terraform-providers).

The provider plugin can be installed automatically via [Para - 3rd-party plugin manager for Terraform](https://github.com/paraterraform/para)
or it can be downloaded and installed manually.  

### Para

This plugin is available via [default index](https://github.com/paraterraform/index) for [Para](https://github.com/paraterraform/para).
If you use Para or Para Launcher you can just skip to the [Usage](#usage) section below assuming you'd wrap all calls to Terraform with Para:
```bash
$ curl -Lo para https://raw.githubusercontent.com/paraterraform/para/master/para && chmod +x para 

$ ./para terraform init
Para Launcher Activated!
- Checking para.cfg.yaml in current directory for 'version: X.Y.Z'
- Desired version: latest (latest is used when no version specified)
- Executing '$TMPDIR/para-501/para/latest/para_v0.4.3_darwin-amd64'

------------------------------------------------------------------------

Para is being initialized...
- Cache Dir: $TMPDIR/para-501
- Terraform: downloading to $TMPDIR/para-501/terraform/0.12.9/darwin_amd64
- Plugin Dir: terraform.d/plugins
- Primary Index: https://raw.githubusercontent.com/paraterraform/index/master/para.idx.yaml as of 2019-10-04T12:18:32-04:00 (providers: 16)
- Index Extensions: para.idx.d (0/0), ~/.para/para.idx.d (0/0), /etc/para/para.idx.d (0/0)
- Command: terraform init

------------------------------------------------------------------------


Initializing the backend...

Initializing provider plugins...
- Para provides 3rd-party Terraform provider plugin 'consulacl' version 'v1.5.0' for 'darwin_amd64' (downloading)


The following providers do not have any version constraints in configuration,
so the latest version was installed.

To prevent automatic upgrades to new major versions that may contain breaking
changes, it is recommended to add version = "..." constraints to the
corresponding provider blocks in configuration, with the constraint strings
suggested below.

* provider.consulacl: version = "~> 1.5"

Terraform has been successfully initialized!
```  

If you use Para but don't use the [default index](https://github.com/paraterraform/index) you can make the plugin
available by including index extension for this plugin: either add [`provider.consulacl.yaml`](./provider.consulacl.yaml)
from this repo to your [Para index extensions dir](https://github.com/paraterraform/para#extensions) to fix currently
available versions or create `provider.consulacl.yaml` as an empty file and put the URL to the aforementioned file
inside to automatically get updates:
```yaml
https://raw.githubusercontent.com/ashald/terraform-provider-consulacl/master/provider.consulacl.yaml
```

### Manual

> Terraform will search for matching Providers via a
> [Discovery](https://www.terraform.io/docs/extend/how-terraform-works.html#discovery) process, **including the current
> local directory**.

This means that the plugin should either be placed into current working directory where Terraform will be executed from
or it can be [installed system-wide](https://www.terraform.io/docs/configuration/providers.html#third-party-plugins).

```bash
wget "https://github.com/ashald/terraform-provider-consulacl/releases/download/1.5.0/terraform-provider-consulacl_v1.5.0-$(uname -s | tr '[:upper:]' '[:lower:]')-amd64"
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
