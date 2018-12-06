package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
)

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
