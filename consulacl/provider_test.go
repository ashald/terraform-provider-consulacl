package consulacl

import (
	"testing"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

var testProviders map[string]terraform.ResourceProvider

var aclProvider *schema.Provider

func init() {
	aclProvider = Provider().(*schema.Provider)

	testProviders = map[string]terraform.ResourceProvider{
		"consulacl": aclProvider,
	}
}

func TestProvider(t *testing.T) {
	if err := aclProvider.InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}
