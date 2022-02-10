package es

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mitchellh/go-homedir"
	elastic7 "github.com/olivere/elastic/v7"
	elastic6 "gopkg.in/olivere/elastic.v6"
)

var (
	errObjNotFound = fmt.Errorf("object not found")
)

func elastic7GetObject(client *elastic7.Client, index string, id string) (*elastic7.GetResult, error) {
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

	return result, nil
}

func elastic6GetObject(client *elastic6.Client, objectType string, index string, id string) (*elastic6.GetResult, error) {
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

	return result, nil
}

func normalizeDestination(tpl map[string]interface{}) {
	delete(tpl, "id")
	delete(tpl, "last_update_time")
	delete(tpl, "schema_version")
}

func normalizeMonitor(tpl map[string]interface{}) {
	if triggers, ok := tpl["triggers"].([]interface{}); ok {
		normalizeMonitorTriggers(triggers)
	}

	delete(tpl, "id")
	delete(tpl, "last_update_time")
	delete(tpl, "enabled_time")
	delete(tpl, "schema_version")
	delete(tpl, "user")
}

func normalizeMonitorTriggers(triggers []interface{}) {
	for _, t := range triggers {
		if trigger, ok := t.(map[string]interface{}); ok {
			delete(trigger, "id")

			if actions, ok := trigger["actions"].([]interface{}); ok {
				normalizeMonitorTriggerActions(actions)
			}
		}
	}
}

func normalizeMonitorTriggerActions(actions []interface{}) {
	for _, a := range actions {
		action := a.(map[string]interface{})
		delete(action, "id")
	}
}

func normalizePolicy(tpl map[string]interface{}) {
	delete(tpl, "last_updated_time")
	delete(tpl, "policy_id")
	delete(tpl, "schema_version")
	if ism_template, ok := tpl["ism_template"]; ok {
		if ism_template == nil {
			delete(tpl, "ism_template")
		}

		switch templates := ism_template.(type) {
		case map[string]interface{}:
			delete(templates, "last_updated_time")
		case []interface{}:
			for _, t := range templates {
				if template, ok := t.(map[string]interface{}); ok {
					delete(template, "last_updated_time")
				}
			}
		default:
			log.Printf("[INFO] normalizePolicy unknown type: %T", ism_template)
		}
	}
	// ignore if set to null in response (ie not specified)
	if error_notification, ok := tpl["error_notification"]; ok {
		if error_notification == nil {
			delete(tpl, "error_notification")
		}
	}
}

func normalizeIndexTemplate(tpl map[string]interface{}) {
	delete(tpl, "version")
	if settings, ok := tpl["settings"]; ok {
		if settingsMap, ok := settings.(map[string]interface{}); ok {
			tpl["settings"] = normalizedIndexSettings(settingsMap)
		}
	}
}

/*
normalizeComposableIndexTemplate normalizes an index_template (ES >= 7.8) Index template definition.
For legacy index templates (ES < 7.8) or /_template endpoint on ES >= 7.8 see normalizeIndexTemplate.
*/
func normalizeComposableIndexTemplate(tpl map[string]interface{}) {
	delete(tpl, "version")

	// data_stream accepts only the attribute "hidden", but can return additional attributes, so
	// remove them
	if dataStream, ok := tpl["data_stream"].(map[string]interface{}); ok {
		for k := range dataStream {
			if k != "hidden" {
				delete(dataStream, k)
			}
		}
	}

	if innerTpl, ok := tpl["template"]; ok {
		if innerTplMap, ok := innerTpl.(map[string]interface{}); ok {
			if settings, ok := innerTplMap["settings"]; ok {
				if settingsMap, ok := settings.(map[string]interface{}); ok {
					innerTplMap["settings"] = normalizedIndexSettings(settingsMap)
				}
			}
		}
	}
}

func normalizeComponentTemplate(tpl map[string]interface{}) {
	delete(tpl, "version")
	if innerTpl, ok := tpl["template"]; ok {
		if innerTplMap, ok := innerTpl.(map[string]interface{}); ok {
			if settings, ok := innerTplMap["settings"]; ok {
				if settingsMap, ok := settings.(map[string]interface{}); ok {
					innerTplMap["settings"] = normalizedIndexSettings(settingsMap)
				}
			}
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

func normalizeSnapshotLifecyclePolicy(pol map[string]interface{}) {
	delete(pol, "version")
	delete(pol, "modified_date")
	delete(pol, "modified_date_millis")
	delete(pol, "stats")
	delete(pol, "next_execution")
	delete(pol, "next_execution_millis")
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

func concatStringSlice(args ...[]string) []string {
	merged := make([]string, 0)
	for _, slice := range args {
		merged = append(merged, slice...)
	}
	return merged
}

func containsString(h []string, n string) bool {
	for _, e := range h {
		if e == n {
			return true
		}
	}
	return false
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

func flattenStringList(list []string) []interface{} {
	vs := make([]interface{}, 0, len(list))
	for _, v := range list {
		vs = append(vs, v)
	}
	return vs
}

func flattenStringSet(list []string) *schema.Set {
	return flattenStringAsInterfaceSet(flattenStringList(list))
}

func flattenStringAsInterfaceSet(list []interface{}) *schema.Set {
	return schema.NewSet(schema.HashString, list)
}

func flattenFloatSet(list []interface{}) *schema.Set {
	hashFloat := func(v interface{}) int {
		return hashcode(strconv.FormatFloat(v.(float64), 'f', -1, 64))
	}

	return schema.NewSet(hashFloat, list)
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

func flattenIndexPermissions(permissions []IndexPermissions, d *schema.ResourceData) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(permissions))
	for _, permission := range permissions {
		p := make(map[string]interface{})

		if len(permission.IndexPatterns) > 0 {
			p["index_patterns"] = flattenStringSet(permission.IndexPatterns)
		}
		if len(permission.DocumentLevelSecurity) > 0 {
			p["document_level_security"] = permission.DocumentLevelSecurity
		}

		if len(permission.FieldLevelSecurity) > 0 {
			p["field_level_security"] = flattenStringSet(permission.FieldLevelSecurity)
		}

		if len(permission.MaskedFields) > 0 {
			p["masked_fields"] = flattenStringSet(permission.MaskedFields)
		}
		if len(permission.AllowedActions) > 0 {
			p["allowed_actions"] = flattenStringSet(permission.AllowedActions)
		}

		result = append(result, p)
	}

	return result
}

func expandIndexPermissionsSet(resourcesArray []interface{}) ([]IndexPermissions, error) {
	vperm := make([]IndexPermissions, 0, len(resourcesArray))
	for _, item := range resourcesArray {
		data, ok := item.(map[string]interface{})
		if !ok {
			return vperm, fmt.Errorf("Error asserting data as type []byte : %v", item)
		}

		fls := data["field_level_security"]
		flsList := fls.(*schema.Set).List()

		obj := IndexPermissions{
			IndexPatterns:         expandStringList(data["index_patterns"].(*schema.Set).List()),
			DocumentLevelSecurity: data["document_level_security"].(string),
			FieldLevelSecurity:    expandStringList(flsList),
			MaskedFields:          expandStringList(data["masked_fields"].(*schema.Set).List()),
			AllowedActions:        expandStringList(data["allowed_actions"].(*schema.Set).List()),
		}
		vperm = append(vperm, obj)
	}
	return vperm, nil
}

func flattenTenantPermissions(permissions []TenantPermissions) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(permissions))
	for _, permission := range permissions {
		p := make(map[string]interface{})

		if len(permission.TenantPatterns) > 0 {
			p["tenant_patterns"] = flattenStringSet(permission.TenantPatterns)
		}
		if len(permission.AllowedActions) > 0 {
			p["allowed_actions"] = flattenStringSet(permission.AllowedActions)
		}

		result = append(result, p)
	}

	return result
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

func hashSum(contents interface{}) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(contents.(string))))
}

func indexPermissionsHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})

	// We need to make sure to sort the strings below so that we always
	// generate the same hash code no matter what is in the set.
	if v, ok := m["index_patterns"]; ok {
		vs := v.(*schema.Set).List()
		s := make([]string, len(vs))
		for i, raw := range vs {
			s[i] = raw.(string)
		}
		sort.Strings(s)

		for _, v := range s {
			buf.WriteString(fmt.Sprintf("%s-", v))
		}
	}

	if v, ok := m["document_level_security"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	if v, ok := m["fls"]; ok {
		vs := v.(*schema.Set).List()
		s := make([]string, len(vs))
		for i, raw := range vs {
			s[i] = raw.(string)
		}
		sort.Strings(s)

		for _, v := range s {
			buf.WriteString(fmt.Sprintf("%s-", v))
		}
	}

	if v, ok := m["field_level_security"]; ok {
		vs := v.(*schema.Set).List()
		s := make([]string, len(vs))
		for i, raw := range vs {
			s[i] = raw.(string)
		}
		sort.Strings(s)

		for _, v := range s {
			buf.WriteString(fmt.Sprintf("%s-", v))
		}
	}
	if v, ok := m["masked_fields"]; ok {
		vs := v.(*schema.Set).List()
		s := make([]string, len(vs))
		for i, raw := range vs {
			s[i] = raw.(string)
		}
		sort.Strings(s)

		for _, v := range s {
			buf.WriteString(fmt.Sprintf("%s-", v))
		}
	}
	if v, ok := m["allowed_actions"]; ok {
		vs := v.(*schema.Set).List()
		s := make([]string, len(vs))
		for i, raw := range vs {
			s[i] = raw.(string)
		}
		sort.Strings(s)

		for _, v := range s {
			buf.WriteString(fmt.Sprintf("%s-", v))
		}
	}

	return hashcode(buf.String())
}

func tenantPermissionsHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})

	// We need to make sure to sort the strings below so that we always
	// generate the same hash code no matter what is in the set.
	if v, ok := m["tenant_patterns"]; ok {
		vs := v.(*schema.Set).List()
		s := make([]string, len(vs))
		for i, raw := range vs {
			s[i] = raw.(string)
		}
		sort.Strings(s)

		for _, v := range s {
			buf.WriteString(fmt.Sprintf("%s-", v))
		}
	}
	if v, ok := m["allowed_actions"]; ok {
		vs := v.(*schema.Set).List()
		s := make([]string, len(vs))
		for i, raw := range vs {
			s[i] = raw.(string)
		}
		sort.Strings(s)

		for _, v := range s {
			buf.WriteString(fmt.Sprintf("%s-", v))
		}
	}

	return hashcode(buf.String())
}

// hashcode hashes a string to a unique hash code.
//
// crc32 returns a uint32, but for our use we need
// and non negative integer. Here we cast to an integer
// and invert it if the result is negative.
func hashcode(s string) int {
	v := int(crc32.ChecksumIEEE([]byte(s)))
	if v >= 0 {
		return v
	}
	if -v >= 0 {
		return -v
	}
	// v == MinInt
	return 0
}

// If the argument is a path, readPathOrContent loads it and returns the contents,
// otherwise the argument is assumed to be the desired contents and is simply
// returned.
//
// The boolean second return value can be called `wasPath` - it indicates if a
// path was detected and a file loaded.
func readPathOrContent(poc string) (string, bool, error) {
	if len(poc) == 0 {
		return poc, false, nil
	}

	path := poc
	if path[0] == '~' {
		var err error
		path, err = homedir.Expand(path)
		if err != nil {
			return path, true, err
		}
	}

	if _, err := os.Stat(path); err == nil {
		contents, err := ioutil.ReadFile(path)
		if err != nil {
			return string(contents), true, err
		}
		return string(contents), true, nil
	}

	return poc, false, nil
}

func toCamelCase(underScored string, startUpperCased bool) (camelCased string) {
	isToUpper := false

	for k, v := range underScored {
		if k == 0 {
			if startUpperCased {
				camelCased = strings.ToUpper(string(underScored[0]))
			} else {
				camelCased = strings.ToLower(string(underScored[0]))
			}
		} else {
			if isToUpper {
				camelCased += strings.ToUpper(string(v))
				isToUpper = false
			} else {
				if v == '_' {
					isToUpper = true
				} else {
					camelCased += string(v)
				}
			}
		}
	}
	return
}

func toUnderscore(s string) string {
	var res = make([]rune, 0, len(s))
	var p = '_'
	for i, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			res = append(res, '_')
		} else if unicode.IsUpper(r) && i > 0 {
			if unicode.IsLetter(p) && !unicode.IsUpper(p) || unicode.IsDigit(p) {
				res = append(res, '_', unicode.ToLower(r))
			} else {
				res = append(res, unicode.ToLower(r))
			}
		} else {
			res = append(res, unicode.ToLower(r))
		}

		p = r
	}
	return string(res)
}

// borrowed from upstream terraform, this isn't exported though
type diagnosticsAsError struct {
	diag.Diagnostics
}

func (dae diagnosticsAsError) Error() string {
	diags := dae.Diagnostics
	switch {
	case len(diags) == 0:
		// should never happen, since we don't create this wrapper if
		// there are no diagnostics in the list.
		return "no errors"
	case len(diags) == 1:
		if diags[0].Detail == "" {
			return diags[0].Summary
		}
		return fmt.Sprintf("%s: %s", diags[0].Summary, diags[0].Detail)
	default:
		var ret bytes.Buffer
		fmt.Fprintf(&ret, "%d problems:\n", len(diags))
		for _, diag := range dae.Diagnostics {
			if diag.Detail == "" {
				fmt.Fprintf(&ret, "\n- %s", diag.Summary)
			} else {
				fmt.Fprintf(&ret, "\n- %s: %s", diag.Summary, diag.Detail)
			}
		}
		return ret.String()
	}
}
