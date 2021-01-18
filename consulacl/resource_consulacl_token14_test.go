package consulacl_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/ashald/terraform-provider-consulacl/consulacl"
)

const (
	resourceAclToken14Accessor      = "86bfeed0-bc8f-4bd6-8dcb-e4f404409a4d"
	resourceAclToken14Secret        = "fadfa76c-e894-4bea-8ac7-5249356e148b"
	resourceAclToken14ConfigInitial = `
resource "consulacl_token14" "test" {
	accessor = "86bfeed0-bc8f-4bd6-8dcb-e4f404409a4d"
	secret = "fadfa76c-e894-4bea-8ac7-5249356e148b"
	description = "Test Token Initial"
}
`
)

const resourceAclToken14ConfigUpdated = `
resource "consulacl_token14" "test" {
	accessor = "86bfeed0-bc8f-4bd6-8dcb-e4f404409a4d"
	secret = "fadfa76c-e894-4bea-8ac7-5249356e148b"
	description = "Test Token Updated"
	policies = ["global-management"]
}
`

func TestIntegrationResourceToken14(t *testing.T) {
	resource.Test(t, resource.TestCase{
		IsUnitTest:   false,
		Providers:    testProviders,
		PreCheck:     func() { testResourcePreConfig(t) },
		CheckDestroy: testResourceToken14Absent(resourceAclToken14Accessor),
		Steps: []resource.TestStep{
			{
				Config: resourceAclToken14ConfigInitial,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("consulacl_token14.test", consulacl.FieldAccessor, resourceAclToken14Accessor),
					resource.TestCheckResourceAttr("consulacl_token14.test", consulacl.FieldSecret, resourceAclToken14Secret),
					resource.TestCheckResourceAttr("consulacl_token14.test", consulacl.FieldDescription, "Test Token Initial"),
					resource.TestCheckResourceAttr("consulacl_token14.test", "policies.#", "0"),
				),
			},
			{
				Config: resourceAclToken14ConfigUpdated,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("consulacl_token14.test", consulacl.FieldAccessor, resourceAclToken14Accessor),
					resource.TestCheckResourceAttr("consulacl_token14.test", consulacl.FieldSecret, resourceAclToken14Secret),
					resource.TestCheckResourceAttr("consulacl_token14.test", consulacl.FieldDescription, "Test Token Updated"),
					resource.TestCheckResourceAttr("consulacl_token14.test", "policies.#", "1"),
					// 'policies' is a set and '938696404' appears to be a hash-value of the string 'global-management' ¯\_(ツ)_/¯
					resource.TestCheckResourceAttr("consulacl_token14.test", "policies.938696404", "global-management"),
				),
			},
		},
	})
}

func testResourceToken14Absent(accessor string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		tokens, _, err := testClient.ACL().TokenList(nil)
		if err != nil {
			return fmt.Errorf("error listing post-1.4 ACL tokens in Consul: %s", err)
		}
		for _, entry := range tokens {
			if entry.AccessorID == accessor {
				return fmt.Errorf("a test token with accessor %q wasn't deleted from Consul", accessor)
			}
		}
		return nil
	}
}
