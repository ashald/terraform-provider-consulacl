package consulacl

import (
	"crypto/sha256"
	"fmt"
	consul "github.com/hashicorp/consul/api"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/hcl"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"reflect"
	"sort"
	"strings"
)

const FieldName = "name"
const FieldToken = "token"
const FieldAccessor = "accessor"
const FieldSecret = "secret"
const FieldType = "type"

const FieldRule = "rule"

const FieldScope = "scope"
const FieldPrefix = "prefix"
const FieldPolicy = "policy"

var prefixedScopes = []string{"agent", "event", "key", "node", "query", "service", "session"}
var singletonScopes = []string{"keyring", "operator"}

func resourceConsulAclToken() *schema.Resource {
	var allScopes []string
	allScopes = append(allScopes, prefixedScopes...)
	allScopes = append(allScopes, singletonScopes...)

	return &schema.Resource{
		Create: resourceConsulAclTokenCreate,
		Update: resourceConsulAclTokenUpdate,
		Read:   resourceConsulAclTokenRead,
		Delete: resourceConsulAclTokenDelete,

		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				_ = d.Set(FieldToken, d.Id())
				d.SetId(getSHA256(d.Id()))
				return []*schema.ResourceData{d}, nil
			},
		},

		CustomizeDiff: diffResource,

		Schema: map[string]*schema.Schema{
			FieldName: {
				Type:     schema.TypeString,
				Required: true,
			},

			FieldRule: {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						FieldScope: {
							Type: schema.TypeString,
							// it's required but we have to enforce otherwise due to a bug in terraform
							// when injecting rules as
							// rule = ["${data.null_data_source.policy.*.outputs}"]
							Optional:     true,
							ValidateFunc: validation.StringInSlice(allScopes, true),
						},
						FieldPrefix: {
							Type:     schema.TypeString,
							Optional: true,
						},
						FieldPolicy: {
							Type: schema.TypeString,
							// it's required but we have to enforce otherwise due to a bug in terraform
							// when injecting rules as
							// rule = ["${data.null_data_source.policy.*.outputs}"]
							Optional: true,
						},
					},
				},
			},

			FieldToken: {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
				Computed:  true,
			},

			FieldType: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"client", "management"}, true),
			},
		},
	}
}

func resourceConsulAclTokenCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*consul.Client)

	rules, err := extractRules(d.Get(FieldRule).(*schema.Set).List())
	if err != nil {
		return err
	}

	acl := &consul.ACLEntry{
		ID:    d.Get(FieldToken).(string),
		Name:  d.Get(FieldName).(string),
		Type:  d.Get(FieldType).(string),
		Rules: encodeRules(rules),
	}

	token, _, err := client.ACL().Create(acl, nil)
	if err != nil {
		return err
	}

	d.SetId(getSHA256(token))
	d.Set(FieldToken, token)
	return resourceConsulAclTokenRead(d, meta)
}

func resourceConsulAclTokenRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*consul.Client)

	_, err := extractRules(d.Get(FieldRule).(*schema.Set).List())
	if err != nil {
		return err
	}

	acl, _, err := client.ACL().Info(d.Get(FieldToken).(string), nil)
	if err != nil {
		return err
	}

	if acl == nil {
		d.SetId("")
		return nil
	}

	d.Set(FieldName, acl.Name)
	d.Set(FieldType, acl.Type)

	rules, err := decodeRules(acl.Rules)
	if err != nil {
		return err
	}

	d.Set(FieldRule, rules)

	return nil
}

func resourceConsulAclTokenUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*consul.Client)

	rules, err := extractRules(d.Get(FieldRule).(*schema.Set).List())
	if err != nil {
		return err
	}

	acl := &consul.ACLEntry{
		ID:    d.Get(FieldToken).(string),
		Name:  d.Get(FieldName).(string),
		Type:  d.Get(FieldType).(string),
		Rules: encodeRules(rules),
	}

	_, err = client.ACL().Update(acl, nil)
	if err != nil {
		return err
	}

	return resourceConsulAclTokenRead(d, meta)
}

func resourceConsulAclTokenDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*consul.Client)

	_, err := client.ACL().Destroy(d.Get(FieldToken).(string), nil)
	if err != nil {
		return err
	}

	d.SetId("")
	return nil
}

// So this one is really ugly. But it's still more convenient that native HCL struct de-serialization
func decodeRules(raw string) ([]map[string]string, error) {
	var result []map[string]string

	var policies map[string]interface{}

	err := hcl.Decode(&policies, raw)
	if err != nil {
		return nil, err
	}

	for scope, scopeDefinitions := range policies {
		// scopeDefinitions is either of:
		// {"foo/bar":[{"policy":"write"}]}
		// "operator":"read"

		defRef := reflect.ValueOf(scopeDefinitions)

		if defRef.Kind() == reflect.String {
			simplePolicyValue := defRef.String()
			simplePolicy := map[string]string{FieldScope: scope, FieldPolicy: simplePolicyValue}
			result = append(result, simplePolicy)
		} else {

			for i := 0; i < defRef.Len(); i++ {
				scopePolicyRef := defRef.Index(i)

				prefixRef := scopePolicyRef.MapKeys()[0]
				// "foo/bar"
				prefix := prefixRef.String()

				// {"policy":"write"}
				policyMapRef := scopePolicyRef.MapIndex(prefixRef).Elem().Index(0)

				policyMap := make(map[string]string)
				for _, k := range policyMapRef.MapKeys() {
					policyMap[k.String()] = policyMapRef.MapIndex(k).Elem().String()
				}

				policy, ok := policyMap["policy"]
				if ok {
					decodedPolicy := map[string]string{FieldScope: scope, FieldPrefix: prefix, FieldPolicy: policy}
					result = append(result, decodedPolicy)
				}
			}
		}
	}

	return result, nil
}

// HCL lib does not provide Marshal/Serialize functionality :/
func encodeRules(rules []map[string]string) string {
	var result []string

	for _, rule := range rules {
		policy := strings.ToLower(rule[FieldPolicy])
		scope := strings.ToLower(rule[FieldScope])
		prefix, ok := rule[FieldPrefix]

		var ruleStr string

		if ok {
			ruleStr = fmt.Sprintf("%s \"%s\" { policy = \"%s\" }", scope, strings.ToLower(prefix), policy)
		} else {
			ruleStr = fmt.Sprintf("%s = \"%s\"", scope, policy)
		}
		result = append(result, ruleStr)

	}
	sort.Strings(result)
	result = append(result, "")
	return strings.Join(result, "\n")
}

func extractRules(rawRules []interface{}) ([]map[string]string, error) {
	var allErrors *multierror.Error

	var result []map[string]string
	for _, raw := range rawRules {
		definition := raw.(map[string]interface{})

		scope := definition[FieldScope].(string)
		if scope == "" {
			err := fmt.Errorf("the '%s' field is required in: '%v'", FieldScope, definition)
			allErrors = multierror.Append(allErrors, err)
		}

		policy := definition[FieldPolicy].(string)
		if policy == "" {
			err := fmt.Errorf("the '%s' field is required in: '%v'", FieldPolicy, definition)
			allErrors = multierror.Append(allErrors, err)
		}

		prefix := definition[FieldPrefix].(string)
		rule := map[string]string{FieldScope: scope, FieldPolicy: policy}

		if stringInSlice(scope, prefixedScopes) {
			rule[FieldPrefix] = strings.ToLower(prefix)
		} else if prefix != "" {
			err := fmt.Errorf("the 'prefix' field is not allowed on scopes %s: %v", strings.Join(singletonScopes, ", "), definition)
			allErrors = multierror.Append(allErrors, err)
		}

		result = append(result, rule)
	}

	return result, allErrors.ErrorOrNil()
}

func stringInSlice(str string, list []string) bool {
	for _, elem := range list {
		if elem == str {
			return true
		}
	}
	return false
}

// We only need this to run manual validation on fields
func diffResource(d *schema.ResourceDiff, m interface{}) error {
	_, newRules := d.GetChange(FieldRule)

	_, err := extractRules(newRules.(*schema.Set).List())
	if err != nil {
		return err
	}

	return nil
}

func getSHA256(src string) string {
	h := sha256.New()
	h.Write([]byte(src))
	return fmt.Sprintf("%x", h.Sum(nil))
}
