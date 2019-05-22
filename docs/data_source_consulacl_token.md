# data "consulacl_token"

## Overview
Retrieves post-Consul 1.4 ACL token's secret ID by its accessor ID

## Arguments

The following arguments are supported:

* `accessor` - (Required) Accessor ID to fetch token by

## Attributes

The following attribute is exported:

* `secret` - String, the ACL token's secret value. Sensitive.

## Usage Example

### Configure
```hcl
data "consulacl_token" "test" {
  accessor = "b09503a5-906b-2b70-45e3-caeef43bba3f"
}

output "result" {
  value = "${data.consulacl_token.test.secret}"
}
```

### Apply
```bash
$ terraform apply
  data.consulacl_token.test: Refreshing state...
  
  Apply complete! Resources: 0 added, 0 changed, 0 destroyed.
  
  Outputs:
  
  result = 62a31ebc-0249-f327-dd3b-81d1f9f3bba5
```
