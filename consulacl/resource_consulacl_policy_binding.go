package consulacl

import (
	"fmt"

	consul "github.com/hashicorp/consul/api"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceConsulAclPolicyBinding() *schema.Resource {
	return &schema.Resource{
		Create: resourceConsulAclPolicyBindingCreate,
		Read:   resourceConsulAclPolicyBindingRead,
		Delete: resourceConsulAclPolicyBindingDelete,

		Schema: map[string]*schema.Schema{
			FieldAccessor: {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Token's accessor ID",
			},
			FieldPolicy: {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Policy name",
			},
		},
	}
}

func resourceConsulAclPolicyBindingCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*consul.Client)

	accessor := d.Get(FieldAccessor).(string)
	policy := d.Get(FieldPolicy).(string)

	aclToken, _, err := client.ACL().TokenRead(accessor, nil)
	if err != nil {
		return err
	}

	d.SetId(getSHA256(accessor + policy))

	index := -1
	for i, policyLink := range aclToken.Policies {
		if policyLink.Name == policy {
			index = i
			break
		}
	}

	if index >= 0 {
		return nil
	}

	aclToken.Policies = append(aclToken.Policies, &consul.ACLTokenPolicyLink{
		Name: policy,
	})

	_, _, err = client.ACL().TokenUpdate(aclToken, nil)
	if err != nil {
		return fmt.Errorf("error binding ACL token %q to the policy %q: %s", accessor, policy, err)
	}

	return nil
}

func resourceConsulAclPolicyBindingRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*consul.Client)

	accessor := d.Get(FieldAccessor).(string)
	policy := d.Get(FieldPolicy).(string)

	aclToken, _, err := client.ACL().TokenRead(accessor, nil)
	if err != nil {
		return err
	}

	index := -1
	for i, policyLink := range aclToken.Policies {
		if policyLink.Name == policy {
			index = i
			break
		}
	}

	if index < 0 {
		d.SetId("")
	}

	return nil
}

func resourceConsulAclPolicyBindingDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*consul.Client)

	accessor := d.Get(FieldAccessor).(string)
	policy := d.Get(FieldPolicy).(string)

	aclToken, _, err := client.ACL().TokenRead(accessor, nil)
	if err != nil {
		return nil // token not found but it also means there are no bindings
	}

	index := -1
	for i, policyLink := range aclToken.Policies {
		if policyLink.Name == policy {
			index = i
			break
		}
	}

	if index < 0 {
		return nil // already not present
	}

	// that's how you delete an element from a slice in go T_T
	aclToken.Policies = append(aclToken.Policies[:index], aclToken.Policies[index+1:]...)
	_, _, err = client.ACL().TokenUpdate(aclToken, nil)
	if err != nil {
		return fmt.Errorf("error un-binding ACL token %q from the policy %q: %s", accessor, policy, err)
	}

	return nil
}
