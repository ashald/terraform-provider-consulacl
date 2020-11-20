# Change Log

## 1.6.1 - 2020-11-20

### Added

- Use `goreleaser` to publish to the Terraform registry.

## 1.6.0 - 2020-03-31

### Added

- Resource `consulacl_token` now resets `anonymous` token permissions upon deletion as it cannot be removed 

## 1.5.0 - 2019-10-04

### Added

- Resource `consulacl_token14` to manage post-Consul 1.4 tokens and their `accessor`/`secret` fields

## 1.4.0 - 2019-07-30

### Added

- Resource `consulacl_policy_binding` to assign post-Consul 1.4 ACL policies to tokens by their accessor IDs

## 1.3.0 - 2019-05-26

### Added

- Integration with Terraform `v0.12`

## 1.2.0 - 2019-05-21

### Added

- Data source `consulacl_token` to retrieve Consul ACL secret id by its accessor id

## 1.1.1 - 2018-08-01

### Fixed

- Mark `token` as computed attribute so that it's really optional ([#1])

## 1.1.0 - 2018-07-06

### Added

- Import functionality for `consulacl_token`


## 1.0.0 - 2018-07-05

### Added

- Initial implementation for `consulacl_token` resource

[#1]: https://github.com/Ashald/terraform-provider-consulacl/issues/1
