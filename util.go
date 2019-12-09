package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	elastic7 "github.com/olivere/elastic/v7"
	elastic5 "gopkg.in/olivere/elastic.v5"
	elastic6 "gopkg.in/olivere/elastic.v6"
)

func elastic7GetObject(client *elastic7.Client, objectType string, index string, id string) (*json.RawMessage, error) {
	result, err := client.Get().
		Index(index).
		Type(objectType).
		Id(id).
		Do(context.TODO())

	if err != nil {
		return nil, err
	}
	if !result.Found {
		return nil, fmt.Errorf("Object not found.")
	}

	return &result.Source, nil
}

func elastic6GetObject(client *elastic6.Client, objectType string, index string, id string) (*json.RawMessage, error) {
	result, err := client.Get().
		Index(index).
		Type(objectType).
		Id(id).
		Do(context.TODO())

	if err != nil {
		return nil, err
	}
	if !result.Found {
		return nil, fmt.Errorf("Object not found.")
	}

	return result.Source, nil
}

func elastic5GetObject(client *elastic5.Client, objectType string, index string, id string) (*json.RawMessage, error) {
	result, err := client.Get().
		Index(index).
		Type(objectType).
		Id(id).
		Do(context.TODO())

	if err != nil {
		return nil, err
	}
	if !result.Found {
		return nil, fmt.Errorf("Object not found.")
	}

	return result.Source, nil
}

func normalizeDestination(tpl map[string]interface{}) {
	delete(tpl, "last_update_time")
}

func normalizeIndexTemplate(tpl map[string]interface{}) {
	delete(tpl, "version")
	if settings, ok := tpl["settings"]; ok {
		if settingsMap, ok := settings.(map[string]interface{}); ok {
			tpl["settings"] = normalizedIndexSettings(settingsMap)
		}
	}
}

func normalizedIndexSettings(settings map[string]interface{}) map[string]interface{} {
	f := flattenMap(settings)
	for k, v := range f {
		f[k] = fmt.Sprintf("%v", v)
		if !strings.HasPrefix(k, "index.") {
			f["index."+k] = fmt.Sprintf("%v", v)
			delete(f, k)
		}
	}

	return f
}

func normalizeIndexLifecyclePolicy(pol map[string]interface{}) {
	delete(pol, "version")
	delete(pol, "modified_date")
	if policy, ok := pol["policy"]; ok {
		if policyMap, ok := policy.(map[string]interface{}); ok {
			pol["policy"] = normalizedIndexLifecyclePolicy(policyMap)
		}
	}
}

func normalizedIndexLifecyclePolicy(policy map[string]interface{}) map[string]interface{} {
	f := flattenMap(policy)
	for k, v := range f {
		f[k] = fmt.Sprintf("%v", v)
	}

	return f
}

func flattenMap(m map[string]interface{}) map[string]interface{} {
	f := make(map[string]interface{})
	for k, v := range m {
		if vm, ok := v.(map[string]interface{}); ok {
			fm := flattenMap(vm)
			for k2, v2 := range fm {
				f[k+"."+k2] = v2
			}
		} else {
			f[k] = v
		}
	}

	return f
}

func intp(i int) *int {
	return &i
}

func stringp(s string) *string {
	return &s
}

func boolp(b bool) *bool {
	return &b
}

// Takes the result of flatmap.Expand for an array of strings
// and returns a []string
func expandStringList(resourcesArray []interface{}) []string {
	vs := make([]string, 0, len(resourcesArray))
	for _, v := range resourcesArray {
		val, ok := v.(string)
		if ok && val != "" {
			vs = append(vs, v.(string))
		}
	}
	return vs
}

func expandApplicationPermissionSet(resourcesArray []interface{}) ([]XPackSecurityApplicationPrivileges, error) {
	vperm := make([]XPackSecurityApplicationPrivileges, 0, len(resourcesArray))
	for _, item := range resourcesArray {
		data, ok := item.(map[string]interface{})
		if !ok {
			return vperm, fmt.Errorf("Error asserting data as type []byte : %v", item)
		}
		obj := XPackSecurityApplicationPrivileges{
			Application: data["application"].(string),
			Privileges:  expandStringList(data["privileges"].(*schema.Set).List()),
			Resources:   expandStringList(data["resources"].(*schema.Set).List()),
		}
		vperm = append(vperm, obj)
	}
	return vperm, nil
}

func expandIndicesPermissionSet(resourcesArray []interface{}) ([]XPackSecurityIndicesPermissions, error) {
	vperm := make([]XPackSecurityIndicesPermissions, 0, len(resourcesArray))
	for _, item := range resourcesArray {
		data, ok := item.(map[string]interface{})
		if !ok {
			return vperm, fmt.Errorf("Error asserting data as type []byte : %v", item)
		}
		obj := XPackSecurityIndicesPermissions{
			Names:         expandStringList(data["names"].(*schema.Set).List()),
			Privileges:    expandStringList(data["privileges"].(*schema.Set).List()),
			FieldSecurity: data["field_security"].(string),
			Query:         data["query"].(string),
		}
		vperm = append(vperm, obj)
	}
	return vperm, nil
}

func optionalInterfaceJson(input string) interface{} {
	if input == "" || input == "{}" {
		return nil
	} else {
		return json.RawMessage(input)
	}
}

func expandIndexPermissionsSet(resourcesArray []interface{}) ([]IndexPermissions, error) {
	vperm := make([]IndexPermissions, 0, len(resourcesArray))
	for _, item := range resourcesArray {
		data, ok := item.(map[string]interface{})
		if !ok {
			return vperm, fmt.Errorf("Error asserting data as type []byte : %v", item)
		}
		obj := IndexPermissions{
			IndexPatterns:  expandStringList(data["index_patterns"].(*schema.Set).List()),
			Fls:            expandStringList(data["fls"].(*schema.Set).List()),
			MaskedFields:   expandStringList(data["masked_fields"].(*schema.Set).List()),
			AllowedActions: expandStringList(data["allowed_actions"].(*schema.Set).List()),
		}
		vperm = append(vperm, obj)
	}
	return vperm, nil
}

func expandTenantPermissionsSet(resourcesArray []interface{}) ([]TenantPermissions, error) {
	vperm := make([]TenantPermissions, 0, len(resourcesArray))
	for _, item := range resourcesArray {
		data, ok := item.(map[string]interface{})
		if !ok {
			return vperm, fmt.Errorf("Error asserting data as type []byte : %v", item)
		}
		obj := TenantPermissions{
			TenantPatterns: expandStringList(data["tenant_patterns"].(*schema.Set).List()),
			AllowedActions: expandStringList(data["allowed_actions"].(*schema.Set).List()),
		}
		vperm = append(vperm, obj)
	}
	return vperm, nil
}
