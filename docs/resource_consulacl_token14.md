# resource "consulacl_token14"

## Overview
Exactly as the [official `consul_acl_token` resource](https://www.terraform.io/docs/providers/consul/r/acl_token.html)
but allows setting `accessor` and/or `secret` fields when you need it. **Requires Consul `1.5.0+`.**

## Arguments

The following arguments are supported:

* `accessor` - (Optional) String, the accessor ID for the token - generated if not set
* `secret` - (Optional) String, the secret ID for the token - generated if not set
* `description` - (Optional) String, the description of the token - generated if not set
* `policies` - (Optional) Set of strings, associated policy names - defaults to empty set
* `local` - (Optional) Boolean, a flag to restrict token to the local datacenter - defaults to `false` 

## Usage Example

### Configure

```hcl
resource "consulacl_token14" "test" {
    accessor = "a288508c-372c-4257-b641-5ad37b136b60"
    secret = "8897c608-a2d7-48b6-8db9-65389608eba9"
    description = "Test Token"
    policies = ["global-management"]
    local = true
}
```

### Apply

```bash
$ terraform apply
  consulacl_policy_binding.make-anonymous-poweruser: Refreshing state... [id=33b322809e70be3e888aa9925bd0e98a12455d63b3d40d48dfd73799245d0a9f]
  
  An execution plan has been generated and is shown below.
  Resource actions are indicated with the following symbols:
    + create
  
  Terraform will perform the following actions:
  
    # consulacl_token14.test will be created
    + resource "consulacl_token14" "test" {
        + accessor    = "a288508c-372c-4257-b641-5ad37b136b60"
        + description = "Test Token"
        + id          = (known after apply)
        + local       = true
        + policies    = [
            + "global-management",
          ]
        + secret      = (sensitive value)
      }
  
  Plan: 1 to add, 0 to change, 0 to destroy.
  
  Do you want to perform these actions?
    Terraform will perform the actions described above.
    Only 'yes' will be accepted to approve.
  
    Enter a value: yes
  
  consulacl_token14.test: Creating...
  consulacl_token14.test: Creation complete after 0s [id=a288508c-372c-4257-b641-5ad37b136b60]
  
  Apply complete! Resources: 1 added, 0 changed, 0 destroyed.

```

### Import

```bash
$ terraform import consulacl_token14.test a288508c-372c-4257-b641-5ad37b136b60
  consulacl_token14.test: Importing from ID "a288508c-372c-4257-b641-5ad37b136b60"...
  consulacl_token14.test: Import prepared!
    Prepared consulacl_token14 for import
  consulacl_token14.test: Refreshing state... [id=a288508c-372c-4257-b641-5ad37b136b60]
  
  Import successful!
  
  The resources that were imported are shown above. These resources are now in
  your Terraform state and will henceforth be managed by Terraform.
```
