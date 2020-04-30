package es

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

var (
	errObjNotFound = fmt.Errorf("object not found")
)

func elastic7GetObject(client *elastic7.Client, index string, id string) (*json.RawMessage, error) {
	result, err := client.Get().
		Index(index).
		Id(id).
		Do(context.TODO())

	if err != nil {
		return nil, err
	}
	if !result.Found {
		return nil, errObjNotFound
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
		return nil, errObjNotFound
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
		return nil, errObjNotFound
	}

	return result.Source, nil
}

func normalizeDestination(tpl map[string]interface{}) {
	delete(tpl, "last_update_time")
}

func normalizePolicy(tpl map[string]interface{}) {
	delete(tpl, "last_updated_time")
	delete(tpl, "error_notification")
	delete(tpl, "policy_id")
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

func flattenIndicesFieldSecurity(rawSettings map[string]interface{}) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, 1)
	out = append(out, rawSettings)
	return out
}

func flattenIndicesPermissionSetv6(resourcesArray []elastic6.XPackSecurityIndicesPermissions) ([]XPackSecurityIndicesPermissions, error) {
	vperm := make([]XPackSecurityIndicesPermissions, 0, len(resourcesArray))
	for _, item := range resourcesArray {
		if item.FieldSecurity != nil {
			obj := XPackSecurityIndicesPermissions{
				Names:         item.Names,
				Privileges:    item.Privileges,
				FieldSecurity: flattenIndicesFieldSecurity(item.FieldSecurity.(map[string]interface{})),
				Query:         item.Query,
			}
			vperm = append(vperm, obj)
		} else {
			obj := XPackSecurityIndicesPermissions{
				Names:      item.Names,
				Privileges: item.Privileges,
				Query:      item.Query,
			}
			vperm = append(vperm, obj)
		}
	}

	return vperm, nil
}

func flattenIndicesPermissionSetv7(resourcesArray []elastic7.XPackSecurityIndicesPermissions) ([]XPackSecurityIndicesPermissions, error) {
	vperm := make([]XPackSecurityIndicesPermissions, 0, len(resourcesArray))
	for _, item := range resourcesArray {
		if item.FieldSecurity != nil {
			obj := XPackSecurityIndicesPermissions{
				Names:         item.Names,
				Privileges:    item.Privileges,
				FieldSecurity: flattenIndicesFieldSecurity(item.FieldSecurity.(map[string]interface{})),
				Query:         item.Query,
			}
			vperm = append(vperm, obj)
		} else {
			obj := XPackSecurityIndicesPermissions{
				Names:      item.Names,
				Privileges: item.Privileges,
				Query:      item.Query,
			}
			vperm = append(vperm, obj)
		}
	}

	return vperm, nil
}

func expandIndicesFieldSecurity(collapsedSettings []interface{}) map[string][]string {
	out := make(map[string][]string)

	if len(collapsedSettings) > 0 {
		if collapsedSettings[0].(map[string]interface{})["grant"] != nil {
			out["grant"] = expandStringList(collapsedSettings[0].(map[string]interface{})["grant"].(*schema.Set).List())
		}

		if collapsedSettings[0].(map[string]interface{})["except"] != nil {
			out["except"] = expandStringList(collapsedSettings[0].(map[string]interface{})["except"].(*schema.Set).List())
		}
	}

	return out
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

func expandIndicesPermissionSet(resourcesArray []interface{}) ([]PutRoleIndicesPermissions, error) {
	vperm := make([]PutRoleIndicesPermissions, 0, len(resourcesArray))
	for _, item := range resourcesArray {
		data, ok := item.(map[string]interface{})
		if !ok {
			return vperm, fmt.Errorf("Error asserting data as type []byte : %v", item)
		}

		if len(data["names"].(*schema.Set).List()) > 0 && len(data["privileges"].(*schema.Set).List()) > 0 {
			obj := PutRoleIndicesPermissions{
				Names:         expandStringList(data["names"].(*schema.Set).List()),
				Privileges:    expandStringList(data["privileges"].(*schema.Set).List()),
				FieldSecurity: expandIndicesFieldSecurity(data["field_security"].([]interface{})),
				Query:         data["query"].(string),
			}
			vperm = append(vperm, obj)
		}
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

type resourceDataSetter struct {
	d   *schema.ResourceData
	err error
}

func (ds *resourceDataSetter) set(key string, value interface{}) {
	if ds.err != nil {
		return
	}
	ds.err = ds.d.Set(key, value)
}
