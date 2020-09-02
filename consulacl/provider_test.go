package consulacl_test

import (
	"fmt"
	"github.com/ashald/terraform-provider-consulacl/consulacl"
	"os"
	"testing"

	consul "github.com/hashicorp/consul/api"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

var testProviders map[string]terraform.ResourceProvider

var aclProvider *schema.Provider
var testClient *consul.Client

func init() {
	aclProvider = consulacl.Provider().(*schema.Provider)

	testProviders = map[string]terraform.ResourceProvider{
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

	err = aclProvider.Configure(terraform.NewResourceConfigRaw(raw))
	if err != nil {
		panic(fmt.Sprintf("error configuring the test provider instance: %s", err))
	}

	testClient = aclProvider.Meta().(*consul.Client)
}

func TestProvider(t *testing.T) {
	if err := aclProvider.InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}
