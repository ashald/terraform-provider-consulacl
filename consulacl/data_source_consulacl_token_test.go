package consulacl_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	consul "github.com/hashicorp/consul/api"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/ashald/terraform-provider-consulacl/consulacl"
)

const (
	dataSourceAclTokenAccessor = "65150ab9-1ba8-4538-a1cd-b0f64780ffb6"
	dataSourceAclTokenSecret   = "b7723bf9-cf63-4c69-96bf-dccb924e4734"
	dataSourceAclTokenConfig   = `
data "consulacl_token" "test" {
  accessor = "65150ab9-1ba8-4538-a1cd-b0f64780ffb6"
}
`
)

func TestIntegrationDataSourceToken(t *testing.T) {
	resource.Test(t, resource.TestCase{
		IsUnitTest: false,
		Providers:  testProviders,
		PreCheck:   func() { testDataSourceConsulAclTokenPreConfig(t) },
		Steps: []resource.TestStep{
			{
				Config: dataSourceAclTokenConfig,
				Check:  resource.TestCheckResourceAttr("data.consulacl_token.test", "secret", dataSourceAclTokenSecret),
			},
			{
				Config: "locals {}",
				Check: resource.ComposeTestCheckFunc(
					testDataSourceConsulAclTokenAbsent(dataSourceAclTokenAccessor),
				),
			},
		},
	})
}

func testDataSourceConsulAclTokenPreConfig(t *testing.T) {
	ok := false

	if v := os.Getenv("CONSUL_TOKEN"); v != "" {
		ok = true
	}
	if v := os.Getenv("CONSUL_HTTP_TOKEN"); v != "" {
		ok = true
	}
	if !ok {
		t.Fatal("Either CONSUL_TOKEN or CONSUL_HTTP_TOKEN must be set for integration tests")
	}

	rp := consulacl.Provider()

	raw := map[string]interface{}{}

	diags := rp.Configure(context.TODO(), terraform.NewResourceConfigRaw(raw))
	if diags.HasError() {
		t.Fatalf("err: %#v", diags)
	}

	token := &consul.ACLToken{
		AccessorID: dataSourceAclTokenAccessor,
		SecretID:   dataSourceAclTokenSecret,
	}

	_, _, err := testClient.ACL().TokenCreate(token, nil)
	if err != nil {
		t.Fatal("Cannot provision a test token for consulacl_token datasource test", err)
	}
}

func testDataSourceConsulAclTokenAbsent(accessor string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, err := testClient.ACL().TokenDelete(accessor, nil)
		if err != nil {
			return fmt.Errorf("error deleting test ACL token with accessor '%s': %s", accessor, err)
		}
		return nil
	}
}
