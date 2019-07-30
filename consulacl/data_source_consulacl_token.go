package consulacl

import (
	consul "github.com/hashicorp/consul/api"
	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceConsulAclToken() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceConsulAclTokenRead,

		Schema: map[string]*schema.Schema{
			FieldAccessor: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			FieldSecret: {
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
	_ = d.Set(FieldSecret, acl.SecretID)

	rules, err := decodeRules(acl.Rules)
	if err != nil {
		return err
	}

	_ = d.Set(FieldRule, rules)

	return nil
}
