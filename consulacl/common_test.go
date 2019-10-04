package consulacl_test

import (
	"os"
	"testing"
)

func testResourcePreConfig(t *testing.T) {
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
}
