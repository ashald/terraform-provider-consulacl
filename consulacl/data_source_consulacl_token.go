package consulacl

import (
	consul "github.com/hashicorp/consul/api"
	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceConsulAclToken() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceConsulAclTokenRead,

		Schema: map[string]*schema.Schema{
			"accessor": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"secret": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
		},
	}
}

func dataSourceConsulAclTokenRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*consul.Client)

	id := d.Get(FieldAccessor).(string)
	acl, _, err := client.ACL().TokenRead(id, nil)
	if err != nil {
		return err
	}

	if acl == nil {
		d.SetId("")
		return nil
	}

	d.SetId(id)
	d.Set(FieldSecret, acl.SecretID)

	rules, err := decodeRules(acl.Rules)
	if err != nil {
		return err
	}

	d.Set(FieldRule, rules)

	return nil
}
