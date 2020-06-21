package es

import (
	"encoding/json"
	"reflect"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func diffSuppressIndexTemplate(k, old, new string, d *schema.ResourceData) bool {
	var oo, no interface{}
	if err := json.Unmarshal([]byte(old), &oo); err != nil {
		return false
	}
	if err := json.Unmarshal([]byte(new), &no); err != nil {
		return false
	}

	if om, ok := oo.(map[string]interface{}); ok {
		normalizeIndexTemplate(om)
	}

	if nm, ok := no.(map[string]interface{}); ok {
		normalizeIndexTemplate(nm)
	}

	return reflect.DeepEqual(oo, no)
}

func diffSuppressDestination(k, old, new string, d *schema.ResourceData) bool {
	var oo, no interface{}
	if err := json.Unmarshal([]byte(old), &oo); err != nil {
		return false
	}
	if err := json.Unmarshal([]byte(new), &no); err != nil {
		return false
	}

	if om, ok := oo.(map[string]interface{}); ok {
		normalizeDestination(om)
	}

	if nm, ok := no.(map[string]interface{}); ok {
		normalizeDestination(nm)
	}

	return reflect.DeepEqual(oo, no)
}

func diffSuppressMonitor(k, old, new string, d *schema.ResourceData) bool {
	var oo, no interface{}
	if err := json.Unmarshal([]byte(old), &oo); err != nil {
		return false
	}
	if err := json.Unmarshal([]byte(new), &no); err != nil {
		return false
	}

	if om, ok := oo.(map[string]interface{}); ok {
		normalizeMonitor(om)
	}

	if nm, ok := no.(map[string]interface{}); ok {
		normalizeMonitor(nm)
	}

	return reflect.DeepEqual(oo, no)
}

func suppressEquivalentJson(k, old, new string, d *schema.ResourceData) bool {
	var oldObj, newObj interface{}
	if err := json.Unmarshal([]byte(old), &oldObj); err != nil {
		return false
	}
	if err := json.Unmarshal([]byte(new), &newObj); err != nil {
		return false
	}
	return reflect.DeepEqual(oldObj, newObj)
}

func diffSuppressIndexLifecyclePolicy(k, old, new string, d *schema.ResourceData) bool {
	var oo, no interface{}
	if err := json.Unmarshal([]byte(old), &oo); err != nil {
		return false
	}
	if err := json.Unmarshal([]byte(new), &no); err != nil {
		return false
	}

	if om, ok := oo.(map[string]interface{}); ok {
		normalizeIndexLifecyclePolicy(om)
	}

	if nm, ok := no.(map[string]interface{}); ok {
		normalizeIndexLifecyclePolicy(nm)
	}

	return reflect.DeepEqual(oo, no)
}

func diffSuppressSnapshotLifecyclePolicy(k, old, new string, d *schema.ResourceData) bool {
	var oo, no interface{}
	if err := json.Unmarshal([]byte(old), &oo); err != nil {
		return false
	}
	if err := json.Unmarshal([]byte(new), &no); err != nil {
		return false
	}

	if om, ok := oo.(map[string]interface{}); ok {
		normalizeSnapshotLifecyclePolicy(om)
	}

	if nm, ok := no.(map[string]interface{}); ok {
		normalizeSnapshotLifecyclePolicy(nm)
	}

	return reflect.DeepEqual(oo, no)
}

func diffSuppressIngestPipeline(k, old, new string, d *schema.ResourceData) bool {
	var oo, no interface{}
	if err := json.Unmarshal([]byte(old), &oo); err != nil {
		return false
	}
	if err := json.Unmarshal([]byte(new), &no); err != nil {
		return false
	}

	return reflect.DeepEqual(oo, no)
}

func diffSuppressPolicy(k, old, new string, d *schema.ResourceData) bool {
	var oo, no interface{}
	if err := json.Unmarshal([]byte(old), &oo); err != nil {
		return false
	}
	if err := json.Unmarshal([]byte(new), &no); err != nil {
		return false
	}

	if om, ok := oo.(map[string]interface{}); ok {
		normalizePolicy(om)
	}

	if nm, ok := no.(map[string]interface{}); ok {
		normalizePolicy(nm)
	}

	return reflect.DeepEqual(oo, no)
}
