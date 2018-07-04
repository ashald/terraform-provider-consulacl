package main

import (
	"github.com/ashald/terraform-provider-consulacl/consulacl"
	"github.com/hashicorp/terraform/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: consulacl.Provider})
}
