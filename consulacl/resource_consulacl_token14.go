package consulacl

import (
	"fmt"

	consul "github.com/hashicorp/consul/api"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceConsulAclToken14() *schema.Resource {
	return &schema.Resource{
		Create: resourceConsulACLToken14Create,
		Read:   resourceConsulAclToken14Read,
		Update: resourceConsulAclToken14Update,
		Delete: resourceConsulAclToken14Delete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			FieldAccessor: {
				Type:     schema.TypeString,
				ForceNew: true,
				Computed: true,
				Optional: true,
			},
			FieldSecret: {
				Type:      schema.TypeString,
				ForceNew:  true,
				Computed:  true,
				Sensitive: true,
				Optional:  true,
			},
			FieldDescription: {
				Type:     schema.TypeString,
				Optional: true,
			},
			FieldPolicies: {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			FieldLocal: {
				Type:     schema.TypeBool,
				ForceNew: true,
				Optional: true,
				Default:  false,
			},
		},
	}
}

func resourceConsulACLToken14Create(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*consul.Client)

	aclToken := consul.ACLToken{
		AccessorID:  d.Get(FieldAccessor).(string),
		SecretID:    d.Get(FieldSecret).(string),
		Description: d.Get(FieldDescription).(string),
		Local:       d.Get(FieldLocal).(bool),
	}

	iPolicies := d.Get(FieldPolicies).(*schema.Set).List()
	policyLinks := make([]*consul.ACLTokenPolicyLink, 0, len(iPolicies))
	for _, iPolicy := range iPolicies {
		policyLinks = append(policyLinks, &consul.ACLTokenPolicyLink{
			Name: iPolicy.(string),
		})
	}

	if len(policyLinks) > 0 {
		aclToken.Policies = policyLinks
	}

	token, _, err := client.ACL().TokenCreate(&aclToken, nil)
	if err != nil {
		return fmt.Errorf("error creating ACL token: %s", err)
	}

	d.SetId(token.AccessorID)

	return resourceConsulAclToken14Read(d, meta)
}

func resourceConsulAclToken14Read(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*consul.Client)

	id := d.Id()

	aclToken, _, err := client.ACL().TokenRead(id, nil)
	if err != nil {
		d.SetId("")
		return nil
	}

	if err = d.Set(FieldAccessor, aclToken.AccessorID); err != nil {
		return fmt.Errorf("error while setting %q: %s", FieldAccessor, err)
	}

	if err = d.Set(FieldSecret, aclToken.SecretID); err != nil {
		return fmt.Errorf("error while setting %q: %s", FieldAccessor, err)
	}

	if err = d.Set(FieldDescription, aclToken.Description); err != nil {
		return fmt.Errorf("error while setting %q: %s", FieldDescription, err)
	}

	policies := make([]string, 0, len(aclToken.Policies))
	for _, policyLink := range aclToken.Policies {
		policies = append(policies, policyLink.Name)
	}

	if err = d.Set(FieldPolicies, policies); err != nil {
		return fmt.Errorf("error while setting %q: %s", FieldPolicies, err)
	}
	if err = d.Set(FieldLocal, aclToken.Local); err != nil {
		return fmt.Errorf("error while setting %q: %s", FieldLocal, err)
	}

	return nil
}

func resourceConsulAclToken14Update(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*consul.Client)

	id := d.Id()

	aclToken := consul.ACLToken{
		AccessorID:  id,
		SecretID:    d.Get(FieldSecret).(string),
		Description: d.Get(FieldDescription).(string),
		Local:       d.Get(FieldLocal).(bool),
	}

	if v, ok := d.GetOk(FieldPolicies); ok {
		vs := v.(*schema.Set).List()
		s := make([]*consul.ACLTokenPolicyLink, len(vs))
		for i, raw := range vs {
			s[i] = &consul.ACLTokenPolicyLink{
				Name: raw.(string),
			}
		}
		aclToken.Policies = s
	}

	_, _, err := client.ACL().TokenUpdate(&aclToken, nil)
	if err != nil {
		return fmt.Errorf("error updating ACL token %q: %s", id, err)
	}

	return resourceConsulAclToken14Read(d, meta)
}

func resourceConsulAclToken14Delete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*consul.Client)

	id := d.Id()

	_, err := client.ACL().TokenDelete(id, nil)
	if err != nil {
		return fmt.Errorf("error deleting ACL token %q: %s", id, err)
	}

	return nil
}
