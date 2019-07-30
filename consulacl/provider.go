package consulacl

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	"github.com/mitchellh/mapstructure"
)

func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			// Destination
			"address": {
				Type:     schema.TypeString,
				Optional: true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"CONSUL_ADDRESS",
					"CONSUL_HTTP_ADDR",
				}, "localhost:8500"),
			},

			// Auth

			"token": {
				Type:     schema.TypeString,
				Optional: true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"CONSUL_TOKEN",
					"CONSUL_HTTP_TOKEN",
				}, ""),
			},

			// TLS

			"scheme": {
				Type:     schema.TypeString,
				Optional: true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"CONSUL_SCHEME",
					"CONSUL_HTTP_SCHEME",
				}, "http"),
			},

			"ca_file": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("CONSUL_CA_FILE", ""),
			},

			"cert_file": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("CONSUL_CERT_FILE", ""),
			},

			"key_file": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("CONSUL_KEY_FILE", ""),
			},

			"tls_skip_verify": {
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("CONSUL_TLS_SKIP_VERIFY", false),
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"consulacl_token":          resourceConsulAclToken(),
			"consulacl_policy_binding": resourceConsulAclPolicyBinding(),
		},

		DataSourcesMap: map[string]*schema.Resource{
			"consulacl_token": dataSourceConsulAclToken(),
		},

		ConfigureFunc: configure,
	}
}

func configure(d *schema.ResourceData) (interface{}, error) {
	configRaw := d.Get("").(map[string]interface{})

	var config Config
	if err := mapstructure.Decode(configRaw, &config); err != nil {
		return nil, err
	}

	return config.Client()
}
