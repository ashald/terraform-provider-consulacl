# resource "consulacl_token"

## Overview
Manages bindings post-Consul 1.4 ACL policies and tokens.
This resource is useful when controlling the entire token configuration is impossible or undesirables. For instance, 
that's often the case with pre-generated tokens such as `anonymous` or "master token".

WARNING: This resource cannot be used together with the official [`consul_acl_token`](https://www.terraform.io/docs/providers/consul/r/acl_token.html)
resource as both of them control the `policies` field of the token and will conflict with each other.     

## Arguments

The following arguments are supported:

* `accessor` - (Required) String, accessor ID to fetch token by
* `policy` - (Required) String, policy name to bind toke to  

## Attributes

The following attribute is exported:

* `id` - String, SHA256 hash derived fromm `accessor` and `policy` fields

## Usage Example

### Configure

```hcl
resource "consulacl_policy_binding" "make-anonymous-power-user" {
  accessor = "00000000-0000-0000-0000-000000000002"
  policy = "global-management"
}
```

### Apply

```bash
$ terraform apply
  
  An execution plan has been generated and is shown below.
  Resource actions are indicated with the following symbols:
    + create
  
  Terraform will perform the following actions:
  
    # consulacl_policy_binding.make-anonymous-power-user will be created
    + resource "consulacl_policy_binding" "make-anonymous-power-user" {
        + accessor = "00000000-0000-0000-0000-000000000002"
        + id       = (known after apply)
        + policy   = "global-management"
      }
  
  Plan: 1 to add, 0 to change, 0 to destroy.
  
  Do you want to perform these actions?
    Terraform will perform the actions described above.
    Only 'yes' will be accepted to approve.
  
    Enter a value: yes
  
  consulacl_policy_binding.make-anonymous-power-user: Creating...
  consulacl_policy_binding.make-anonymous-power-user: Creation complete after 0s [id=33b322809e70be3e888aa9925bd0e98a12455d63b3d40d48dfd73799245d0a9f]
  
  Apply complete! Resources: 1 added, 0 changed, 0 destroyed.

```

### Import

Not supported.
Instead just define the binding resource. If binding is already present the resource will assume control over it.