package consulacl

import (
	"fmt"
	"testing"

	"os"

	consul "github.com/hashicorp/consul/api"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

const aclTokenConfig = `
resource "consulacl_token" "token" {
  name  = "A demo token"
  token = "my-custom-token"
  type  = "client"

  rule { scope="key"      policy="read"  prefix="foo/bar"  		 }
  rule { scope="key"      policy="write" prefix="foo/bar/baz"  }
  rule { scope="operator" policy="read"                        }
  rule { scope="service"  policy="read"  prefix=""             }
}

resource "consulacl_token" "second" {
	name = "Second"
	type = "client"

	rule { scope="key" policy="read" prefix="second" }
}

resource "consulacl_token" "inherited" {
	name  = "Inherited token"
	type  = "client"
	token = "my-inherited-token"

	rule { scope="key"	policy="write" prefix="foo/bar"      }
	rule { scope="key"  policy="read"  prefix="foo/bar/baz"  }

	inherits = [  
							"${consulacl_token.token.rule}",
							"${consulacl_token.second.rule}"
							]
}
`

const aclTokenConfigUpdate = `
resource "consulacl_token" "token" {
  name  = "Updated token"
	token = "my-custom-token"
	type  = "management"

  rule { scope="key"     policy="write" prefix=""          }
  rule { scope="keyring" policy="write"                    }
  rule { scope="service" policy="read"  prefix="some/path" }
}

resource "consulacl_token" "inherited" {
	name  = "Inherited token"
	type  = "client"
	token = "my-inherited-token"

	rule { scope="key"	policy="write" prefix="foo/bar"      }
	rule { scope="key"  policy="read"  prefix="foo/bar/baz"  }

	inherits = [ "${consulacl_token.token.rule}" ]
}
`

const rulesOriginal = `key "foo/bar" { policy = "read" }
key "foo/bar/baz" { policy = "write" }
operator = "read"
service "" { policy = "read" }
`
const rulesUpdated = `key "" { policy = "write" }
keyring = "write"
service "some/path" { policy = "read" }
`

func TestIntegrationToken(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testPreCheck(t) },
		Providers: testProviders,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testConsulACLTokenAbsent("my-custom-token"),
		),
		Steps: []resource.TestStep{
			{
				Config: aclTokenConfig,
				//ExpectNonEmptyPlan: true,
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
				Config:             aclTokenConfig,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					// Change token in Consul bypassing Terraform
					checkTokenConfig("my-inherited-token", "type", "client"),
					mutateRealToken("my-inherited-token", "rules", `key "foo/bar" { policy = "write" }`),
				),
			},
			{
				Config: aclTokenConfigUpdate,
				Check: resource.ComposeTestCheckFunc(
					checkTokenConfig("my-custom-token", "name", "Updated token"),
					checkTokenConfig("my-custom-token", "type", "management"),
					checkTokenConfig("my-custom-token", "rules", rulesUpdated),
					checkTokenConfig("my-inherited-token", "type", "client"),
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

  rule { scope="operator" policy="read" }
}
`

func TestIntegrationTokenImport(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testPreCheck(t) },
		Providers: testProviders,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testConsulACLTokenAbsent("my-imported-token"),
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

func testConsulACLTokenAbsent(token string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		acl := aclProvider.Meta().(*consul.Client).ACL()
		entry, _, err := acl.Info(token, nil)
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
		acl := aclProvider.Meta().(*consul.Client).ACL()

		entry, _, err := acl.Info(token, nil)
		if err != nil {
			return err
		}

		tokenMap := entryToMap(entry)
		tokenMap[field] = value
		_, err = acl.Update(mapToEntry(tokenMap), nil)
		return err
	}
}

func deleteToken(token string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		acl := aclProvider.Meta().(*consul.Client).ACL()

		_, err := acl.Destroy(token, nil)
		return err
	}
}

func checkTokenConfig(token, field, expected string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		acl := aclProvider.Meta().(*consul.Client).ACL()

		entry, _, err := acl.Info(token, nil)
		if err != nil {
			return err
		}
		if entry == nil {
			return fmt.Errorf("ACL token %v doesn't exist, but should", token)
		}
		entryMap := entryToMap(entry)
		actual := entryMap[field]
		if actual != expected {
			return fmt.Errorf("ACL token '%#v', field '%s' has value '%v'; expected '%v'", entryMap, field, actual, expected)
		}
		return nil
	}
}

func testPreCheck(t *testing.T) {
	if v := os.Getenv("CONSUL_TOKEN"); v != "" {
		return
	}
	if v := os.Getenv("CONSUL_HTTP_TOKEN"); v != "" {
		return
	}
	t.Fatal("Either CONSUL_TOKEN or CONSUL_HTTP_TOKEN must be set for integration tests")
}

func entryToMap(entry *consul.ACLEntry) map[string]string {
	return map[string]string{
		"id":    entry.ID,
		"name":  entry.Name,
		"type":  entry.Type,
		"rules": entry.Rules,
	}
}

func mapToEntry(entry map[string]string) *consul.ACLEntry {
	return &consul.ACLEntry{
		ID:    entry["id"],
		Name:  entry["name"],
		Type:  entry["type"],
		Rules: entry["rules"],
	}
}
