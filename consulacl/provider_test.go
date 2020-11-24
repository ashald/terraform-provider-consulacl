package consulacl_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	consul "github.com/hashicorp/consul/api"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/ashald/terraform-provider-consulacl/consulacl"
)

var testProviders map[string]*schema.Provider

var (
	aclProvider *schema.Provider
	testClient  *consul.Client
)

func init() {
	aclProvider = consulacl.Provider()

	testProviders = map[string]*schema.Provider{
		"consulacl": aclProvider,
	}

	ok := os.Getenv("TF_ACC") == ""

	if os.Getenv("CONSUL_TOKEN") != "" {
		ok = true
	}
	if os.Getenv("CONSUL_HTTP_TOKEN") != "" {
		ok = true
	}
	if !ok {
		panic("neither CONSUL_TOKEN or CONSUL_HTTP_TOKEN is set for integration tests!")
	}

	raw := map[string]interface{}{}

	diags := aclProvider.Configure(context.TODO(), terraform.NewResourceConfigRaw(raw))
	if diags != nil {
		panic(fmt.Sprintf("error configuring the test provider instance: %#v", diags))
	}

	testClient = aclProvider.Meta().(*consul.Client)
}

func TestProvider(t *testing.T) {
	if err := aclProvider.InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}
