package consulacl

import (
	"crypto/sha256"
	"fmt"
	"log"
	"reflect"
	"sort"
	"strings"

	consul "github.com/hashicorp/consul/api"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/hcl"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

const (
	FieldName  = "name"
	FieldToken = "token"
	FieldType  = "type"

	FieldRule = "rule"

	FieldScope  = "scope"
	FieldPrefix = "prefix"
	FieldPolicy = "policy"

	FieldInherits = "inherits"
)

var prefixedScopes = []string{"agent", "event", "key", "node", "query", "service", "session"}
var singletonScopes = []string{"keyring", "operator"}

func resourceConsulACLToken() *schema.Resource {
	var allScopes []string
	allScopes = append(allScopes, prefixedScopes...)
	allScopes = append(allScopes, singletonScopes...)

	return &schema.Resource{
		Create: resourceConsulACLTokenCreate,
		Update: resourceConsulACLTokenUpdate,
		Read:   resourceConsulACLTokenRead,
		Delete: resourceConsulACLTokenDelete,
		Exists: resourceConsulACLTokenExists,

		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				d.Set(FieldToken, d.Id())
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
				Computed: true,
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
				Computed:  true,
				Optional:  true,
				Sensitive: true,
			},

			FieldType: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"client", "management"}, true),
			},

			FieldInherits: &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						FieldScope: {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(allScopes, true),
						},
						FieldPrefix: {
							Type:     schema.TypeString,
							Optional: true,
						},
						FieldPolicy: {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
		},
	}
}

func resourceConsulACLTokenCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*consul.Client)

	rules, err := extractRules(d.Get(FieldRule).(*schema.Set).List())
	if err != nil {
		return err
	}

	var inheritedRules string

	if len(d.Get("inherits").(*schema.Set).List()) > 0 {
		inherits := d.Get("inherits").(*schema.Set).List()

		existingRules, _ := extractRules(inherits)
		inheritedRules = inheritedRules + encodeRules(existingRules)
		rules, _ = dedupeRules(encodeRules(rules), inheritedRules)
	}

	d.Set(FieldRule, sortRules(rules))

	var acl *consul.ACLEntry

	acl = &consul.ACLEntry{
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
	return resourceConsulACLTokenRead(d, meta)
}

func resourceConsulACLTokenRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*consul.Client)

	acl, _, err := client.ACL().Info(d.Get(FieldToken).(string), nil)
	if err != nil {
		return err
	}

	if acl == nil {
		d.SetId("")
		return nil
	}

	rules, err := decodeRules(acl.Rules)
	if err != nil {
		return err
	}

	d.SetId(getSHA256(acl.ID))
	d.Set(FieldToken, acl.ID)
	d.Set(FieldName, acl.Name)
	d.Set(FieldType, acl.Type)
	d.Set(FieldRule, sortRules(rules))

	return nil
}

func resourceConsulACLTokenExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	client := meta.(*consul.Client)

	_, resp, err := client.ACL().Info(d.Get(FieldToken).(string), nil)
	if err != nil {
		if resp != nil {
			log.Printf("[WARN] Token %s not found", d.Get(FieldName).(string))
			d.SetId("")
			return false, nil
		}
		return false, fmt.Errorf("Error retrieving ACL %s", d.Get(FieldName).(string))
	}

	return true, nil

}

func resourceConsulACLTokenUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*consul.Client)

	rules, err := extractRules(d.Get(FieldRule).(*schema.Set).List())
	if err != nil {
		return err
	}

	if len(d.Get("inherits").(*schema.Set).List()) > 0 {
		var inheritedRules string

		inherits := d.Get("inherits").(*schema.Set).List()
		existingRules, _ := extractRules(inherits)

		inheritedRules = inheritedRules + encodeRules(existingRules)
		rules, _ = dedupeRules(encodeRules(rules), inheritedRules)
	}

	d.Set(FieldRule, rules)

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

	return resourceConsulACLTokenRead(d, meta)
}

func resourceConsulACLTokenDelete(d *schema.ResourceData, meta interface{}) error {
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

func dedupeRules(existingRules string, newRules string) ([]map[string]string, error) {
	var allErrors *multierror.Error
	var result []map[string]string

	// ACL rules most permissive to least
	permissions := []string{"write", "read", "deny"}

	newRules = sortString(existingRules + newRules)
	allRules, err := decodeRules(newRules)

	if err != nil {
		err := fmt.Errorf("Couldn't decode all the rules")
		allErrors = multierror.Append(allErrors, err)
	}

	for _, i := range allRules {
		found := false
		for jIndex, j := range result {
			// Search through the ones we've already added
			if strings.ToLower(i["scope"]) == strings.ToLower(j["scope"]) {
				if stringInSlice(i["scope"], prefixedScopes) {
					iprefix, iok := i["prefix"]
					jprefix, jok := j["prefix"]
					if iok && jok {
						if iprefix == jprefix {
							iPermissions := indexInSlice(i["policy"], permissions)
							jPermissions := indexInSlice(j["policy"], permissions)

							// The lower the index, the more permissive
							// -1 if it's not found (it should never be -1)
							if jPermissions > -1 && iPermissions < jPermissions {
								result[jIndex] = i
								found = true
							}
						}
					}
				} else if stringInSlice(i["scope"], singletonScopes) {
					iPermissions := indexInSlice(i["policy"], permissions)
					jPermissions := indexInSlice(j["policy"], permissions)

					if jPermissions > -1 && iPermissions <= jPermissions {
						result[jIndex] = i
						found = true
					}
				}
			}
		}
		if found == false {
			result = append(result, i)
		}
	}

	return result, allErrors.ErrorOrNil()
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

		prefix := ""
		if definition[FieldPrefix] == nil {
			prefix = ""
		} else {
			prefix = definition[FieldPrefix].(string)
		}
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

func indexInSlice(str string, list []string) int {
	for index, key := range list {
		if key == str {
			return index
		}
	}
	return -1
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

	if len(d.Get("inherits").(*schema.Set).List()) > 0 {
		rules, err := extractRules(d.Get(FieldRule).(*schema.Set).List())
		if err != nil {
			return err
		}

		inherits := d.Get("inherits").(*schema.Set).List()
		extractedInherits, err := extractRules(inherits)
		if err != nil {
			return err
		}
		inheritedRules := encodeRules(extractedInherits)

		combinedRules, _ := dedupeRules(encodeRules(rules), inheritedRules)

		d.SetNew(FieldRule, combinedRules)

	} else {
		_, newRules := d.GetChange(FieldRule)

		_, err := extractRules(newRules.(*schema.Set).List())
		if err != nil {
			return err
		}
	}

	return nil
}

func getSHA256(src string) string {
	h := sha256.New()
	h.Write([]byte(src))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func sortString(w string) string {
	s := strings.Split(w, "\n")
	sort.Strings(s)
	var n []string
	for _, str := range s {
		if str != "" && stringInSlice(str, n) == false {
			n = append(n, str)
		}
	}
	return strings.Join(n, "\n")
}

func sortRules(rulesList []map[string]string) []map[string]string {
	rules := encodeRules(rulesList)
	rules = sortString(rules)
	rulesDecoded, _ := decodeRules(rules)
	return rulesDecoded
}
