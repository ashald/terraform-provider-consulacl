package consulacl

import (
	"fmt"
	"github.com/hashicorp/terraform/config"
	"os"
	"testing"

	consul "github.com/hashicorp/consul/api"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

var testProviders map[string]terraform.ResourceProvider

var aclProvider *schema.Provider
var testClient *consul.Client

func init() {
	aclProvider = Provider().(*schema.Provider)

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

	rawConfig, err := config.NewRawConfig(raw)
	if err != nil {
		panic(fmt.Sprintf("error initializing config for the test provider instance: %s", err))
	}

	err = aclProvider.Configure(terraform.NewResourceConfig(rawConfig))
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
