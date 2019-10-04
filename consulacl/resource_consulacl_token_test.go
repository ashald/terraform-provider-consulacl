package consulacl_test

import (
	"fmt"
	"testing"
	consul "github.com/hashicorp/consul/api"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

const aclTokenConfig = `
resource "consulacl_token" "token" {
  name  = "A demo token"
  token = "my-custom-token"
  type  = "client"

  rule {
	scope="key"
	policy="write"
	prefix="foo/bar/baz"
  }
  rule {
    scope="operator"
	policy="read"
  }
  rule {
	scope="service"
	policy="read"
	prefix=""
  }
}
`

const aclTokenConfigUpdate = `
resource "consulacl_token" "token" {
  name  = "Updated token"
  token = "my-custom-token"
  type  = "management"

  rule {
	scope="key"
	policy="write"
	prefix=""
  }
  rule {
	scope="keyring"
	policy="write"
  }
  rule {
	scope="service"
	policy="read"
	prefix="some/path"
  }
}
`

const rulesOriginal = `key "foo/bar/baz" { policy = "write" }
operator = "read"
service "" { policy = "read" }
`
const rulesUpdated = `key "" { policy = "write" }
keyring = "write"
service "some/path" { policy = "read" }
`

func TestIntegrationToken(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testResourcePreConfig(t) },
		Providers: testProviders,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testConsulAclTokenAbsent("my-custom-token"),
		),
		Steps: []resource.TestStep{
			{
				Config: aclTokenConfig,
				Check: resource.ComposeTestCheckFunc(
					checkTokenConfig("my-custom-token", "name", "A demo token"),
					checkTokenConfig("my-custom-token", "type", "client"),
					checkTokenConfig("my-custom-token", "rules", rulesOriginal),
				),
			},
			{
				Config:             aclTokenConfig,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					// Change token in Consul bypassing Terraform
					mutateRealToken("my-custom-token", "name", "Mutated Name"),
				),
			},
			{
				Config:             aclTokenConfig,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					// Change token in Consul bypassing Terraform
					mutateRealToken("my-custom-token", "type", "management"),
				),
			},
			{
				Config:             aclTokenConfig,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					// Change token in Consul bypassing Terraform
					mutateRealToken("my-custom-token", "rules", `key "" { policy = "read" }`),
				),
			},
			{
				Config:             aclTokenConfig,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					// Change token in Consul bypassing Terraform
					deleteToken("my-custom-token"),
				),
			},
			{
				Config: aclTokenConfigUpdate,
				Check: resource.ComposeTestCheckFunc(
					checkTokenConfig("my-custom-token", "name", "Updated token"),
					checkTokenConfig("my-custom-token", "type", "management"),
					checkTokenConfig("my-custom-token", "rules", rulesUpdated),
				),
			},
		},
	})
}

const aclTokenImportConfig = `
resource "consulacl_token" "imported" {
  name  = "Imported"
  token = "my-imported-token"
  type  = "client"

  rule {
	scope="operator"
	policy="read"
  }
}
`

func TestIntegrationTokenImport(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testResourcePreConfig(t) },
		Providers: testProviders,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testConsulAclTokenAbsent("my-imported-token"),
		),
		Steps: []resource.TestStep{
			{
				Config: aclTokenImportConfig,
			},
			{
				ResourceName:      "consulacl_token.imported",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     "my-imported-token",
			},
		},
	})
}

func testConsulAclTokenAbsent(token string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		entry, _, err := testClient.ACL().Info(token, nil)
		if err != nil {
			return err
		}
		if entry != nil {
			return fmt.Errorf("ACL token '%s' exists, but shouldn't", token)
		}
		return nil
	}
}

func mutateRealToken(token, field, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		entry, _, err := testClient.ACL().Info(token, nil)
		if err != nil {
			return err
		}

		tokenMap := aclEntryToMap(entry)
		tokenMap[field] = value
		_, err = testClient.ACL().Update(aclEntryFromMap(tokenMap), nil)
		return err
	}
}

func deleteToken(token string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, err := testClient.ACL().Destroy(token, nil)
		return err
	}
}

func checkTokenConfig(token, field, expected string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		entry, _, err := testClient.ACL().Info(token, nil)
		if err != nil {
			return err
		}
		if entry == nil {
			return fmt.Errorf("ACL token %v doesn't exist, but should", token)
		}
		entryMap := aclEntryToMap(entry)
		actual := entryMap[field]
		if actual != expected {
			return fmt.Errorf("ACL token '%#v', field '%s' has value '%v'; expected '%v'", entryMap, field, actual, expected)
		}
		return nil
	}
}

func aclEntryToMap(entry *consul.ACLEntry) map[string]string {
	return map[string]string{
		"id":    entry.ID,
		"name":  entry.Name,
		"type":  entry.Type,
		"rules": entry.Rules,
	}
}

func aclEntryFromMap(entry map[string]string) *consul.ACLEntry {
	return &consul.ACLEntry{
		ID:    entry["id"],
		Name:  entry["name"],
		Type:  entry["type"],
		Rules: entry["rules"],
	}
}
