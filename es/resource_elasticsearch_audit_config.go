package es

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	elastic7 "github.com/olivere/elastic/v7"
)

var auditConfigSchema = map[string]*schema.Schema{
	"enabled": {
		Type:     schema.TypeBool,
		Required: true,
	},
	"audit": {
		Type:     schema.TypeSet,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"enable_rest": {
					Type:     schema.TypeBool,
					Optional: true,
					Default:  false,
				},
				"disabled_rest_categories": {
					Type:     schema.TypeSet,
					Optional: true,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
					Set: schema.HashString,
				},
				"enable_transport": {
					Type:     schema.TypeBool,
					Optional: true,
					Default:  false,
				},
				"disabled_transport_categories": {
					Type:     schema.TypeSet,
					Optional: true,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
					Set: schema.HashString,
				},
				"resolve_bulk_requests": {
					Type:     schema.TypeBool,
					Optional: true,
					Default:  false,
				},
				"log_request_body": {
					Type:     schema.TypeBool,
					Optional: true,
					Default:  false,
				},
				"resolve_indices": {
					Type:     schema.TypeBool,
					Optional: true,
					Default:  false,
				},
				"exclude_sensitive_headers": {
					Type:     schema.TypeBool,
					Optional: true,
					Default:  true,
				},
				"ignore_users": {
					Type:     schema.TypeSet,
					Optional: true,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
					Set: schema.HashString,
				},
				"ignore_requests": {
					Type:     schema.TypeSet,
					Optional: true,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
					Set: schema.HashString,
				},
			},
		},
	},
	"compliance": {
		Type:     schema.TypeSet,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"enabled": {
					Type:     schema.TypeBool,
					Optional: true,
					Default:  false,
				},
				"internal_config": {
					Type:     schema.TypeBool,
					Optional: true,
					Default:  true,
				},
				"external_config": {
					Type:     schema.TypeBool,
					Optional: true,
					Default:  false,
				},
				"read_metadata_only": {
					Type:     schema.TypeBool,
					Optional: true,
					Default:  false,
				},
				"read_watched_field": {
					Type:     schema.TypeSet,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"index": {
								Type:     schema.TypeString,
								Required: true,
							},
							"fields": {
								Type:     schema.TypeSet,
								Required: true,
								Elem: &schema.Schema{
									Type: schema.TypeString,
								},
								Set: schema.HashString,
							},
						},
					},
				},
				"read_ignore_users": {
					Type:     schema.TypeSet,
					Optional: true,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
					Set: schema.HashString,
				},
				"write_metadata_only": {
					Type:     schema.TypeBool,
					Optional: true,
					Default:  false,
				},
				"write_log_diffs": {
					Type:     schema.TypeBool,
					Optional: true,
					Default:  false,
				},
				"write_watched_indices": {
					Type:     schema.TypeSet,
					Optional: true,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
					Set: schema.HashString,
				},
				"write_ignore_users": {
					Type:     schema.TypeSet,
					Optional: true,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
					Set: schema.HashString,
				},
			},
		},
	},
}

func resourceOpenSearchAuditConfig() *schema.Resource {
	return &schema.Resource{
		Create: resourceElasticsearchAuditConfigCreate,
		Read:   resourceElasticsearchAuditConfigRead,
		Update: resourceElasticsearchAuditConfigUpdate,
		Delete: resourceElasticsearchAuditConfigDelete,
		Schema: auditConfigSchema,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceElasticsearchAuditConfigCheckVersion(meta interface{}) error {
	providerConf := meta.(*ProviderConf)
	if _, err := getClient(providerConf); err != nil {
		return err
	}

	elasticVersion, err := version.NewVersion(providerConf.esVersion)
	if err != nil {
		return err
	}

	if providerConf.flavor != Unknown && elasticVersion.Segments()[0] != 1 {
		return fmt.Errorf("audit config only available from OpenSearch >= 1.0, got version %s", elasticVersion.String())
	}

	return nil
}

func resourceElasticsearchAuditConfigCreate(d *schema.ResourceData, m interface{}) error {
	if err := resourceElasticsearchAuditConfigCheckVersion(m); err != nil {
		return err
	}

	if _, err := resourceElasticsearchPutAuditConfig(d, m); err != nil {
		return err
	}

	d.SetId("audit_config")
	return resourceElasticsearchAuditConfigRead(d, m)
}

func resourceElasticsearchAuditConfigRead(d *schema.ResourceData, m interface{}) error {
	if err := resourceElasticsearchAuditConfigCheckVersion(m); err != nil {
		return err
	}

	res, err := resourceElasticsearchGetAuditConfig(m)
	if err != nil {
		if elastic7.IsNotFound(err) {
			log.Printf("[WARN] audit config (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	err = d.Set("enabled", res.Config.Enabled)
	if err != nil {
		return err
	}

	err = d.Set("audit", flattenAudit(res.Config.Audit))
	if err != nil {
		return err
	}

	err = d.Set("compliance", flattenCompliance(res.Config.Compliance))
	if err != nil {
		return err
	}

	return nil
}

func flattenCompliance(com auditConfig_compliance) []map[string]interface{} {
	return []map[string]interface{}{{
		"enabled":               com.Enabled,
		"internal_config":       com.InternalConfig,
		"external_config":       com.ExternalConfig,
		"read_metadata_only":    com.ReadMetadataOnly,
		"read_ignore_users":     com.ReadIgnoreUsers,
		"read_watched_field":    flattenReadWatchedFields(com),
		"write_metadata_only":   com.WriteMetadataOnly,
		"write_log_diffs":       com.WriteLogDiffs,
		"write_watched_indices": com.WriteWatchedIndices,
		"write_ignore_users":    com.WriteIgnoreUsers,
	}}
}

func flattenReadWatchedFields(com auditConfig_compliance) []map[string]interface{} {
	result := []map[string]interface{}{}

	for k, v := range com.ReadWatchedFields {
		item := map[string]interface{}{
			"index":  k,
			"fields": v,
		}
		result = append(result, item)
	}
	return result
}

func flattenAudit(aud auditConfig_audit) []map[string]interface{} {
	return []map[string]interface{}{{
		"enable_rest":                   aud.EnableRest,
		"disabled_rest_categories":      aud.DisabledRestCategories,
		"enable_transport":              aud.EnableTransport,
		"disabled_transport_categories": aud.DisabledTransportCategories,
		"resolve_bulk_requests":         aud.ResolveBulkRequests,
		"log_request_body":              aud.LogRequestBody,
		"resolve_indices":               aud.ResolveIndices,
		"exclude_sensitive_headers":     aud.ExcludeSensitiveHeaders,
		"ignore_users":                  aud.IgnoreUsers,
		"ignore_requests":               aud.IgnoreRequests,
	}}
}

func resourceElasticsearchAuditConfigUpdate(d *schema.ResourceData, m interface{}) error {
	if err := resourceElasticsearchAuditConfigCheckVersion(m); err != nil {
		return err
	}

	if _, err := resourceElasticsearchPutAuditConfig(d, m); err != nil {
		return err
	}

	return resourceElasticsearchAuditConfigRead(d, m)
}

func resourceElasticsearchAuditConfigDelete(d *schema.ResourceData, m interface{}) error {
	if err := resourceElasticsearchAuditConfigCheckVersion(m); err != nil {
		return err
	}

	return nil
}

func resourceElasticsearchGetAuditConfig(m interface{}) (getAuditConfigResponse, error) {
	var err error
	audit := new(getAuditConfigResponse)

	var body json.RawMessage
	esClient, err := getClient(m.(*ProviderConf))
	if err != nil {
		return *audit, err
	}
	switch client := esClient.(type) {
	case *elastic7.Client:
		var res *elastic7.Response
		res, err = client.PerformRequest(context.TODO(), elastic7.PerformRequestOptions{
			Method: "GET",
			Path:   "/_plugins/_security/api/audit",
		})
		if err != nil {
			return *audit, err
		}
		body = res.Body
	default:
		return *audit, errors.New("audit config resource not implemented prior to OpenSearch v1")
	}

	if err := json.Unmarshal(body, &audit); err != nil {
		return *audit, fmt.Errorf("Error unmarshalling user body: %+v: %+v", err, body)
	}
	log.Printf("[INFO] get audit config response: %+v", *audit)
	return *audit, err
}

func expandAudit(d *schema.ResourceData) auditConfig_audit {
	aud, ok := d.GetOk("audit")
	if !ok || len(aud.(*schema.Set).List()) == 0 {
		return auditConfig_audit{
			ExcludeSensitiveHeaders:     true,
			IgnoreUsers:                 []string{},
			IgnoreRequests:              []string{},
			DisabledRestCategories:      []string{},
			DisabledTransportCategories: []string{},
		}
	}

	m := aud.(*schema.Set).List()[0].(map[string]interface{})
	return auditConfig_audit{
		EnableRest:                  m["enable_rest"].(bool),
		EnableTransport:             m["enable_transport"].(bool),
		ExcludeSensitiveHeaders:     m["exclude_sensitive_headers"].(bool),
		ResolveBulkRequests:         m["resolve_bulk_requests"].(bool),
		LogRequestBody:              m["log_request_body"].(bool),
		ResolveIndices:              m["resolve_indices"].(bool),
		IgnoreUsers:                 expandStringList(m["ignore_users"].(*schema.Set).List()),
		IgnoreRequests:              expandStringList(m["ignore_requests"].(*schema.Set).List()),
		DisabledRestCategories:      expandStringList(m["disabled_rest_categories"].(*schema.Set).List()),
		DisabledTransportCategories: expandStringList(m["disabled_transport_categories"].(*schema.Set).List()),
	}
}

func expandCompliance(d *schema.ResourceData) auditConfig_compliance {
	comp, ok := d.GetOk("compliance")
	if !ok || len(comp.(*schema.Set).List()) == 0 {
		return auditConfig_compliance{
			InternalConfig:      true,
			ExternalConfig:      false,
			ReadWatchedFields:   map[string][]string{},
			ReadIgnoreUsers:     []string{},
			WriteWatchedIndices: []string{},
			WriteIgnoreUsers:    []string{},
		}
	}

	m := comp.(*schema.Set).List()[0].(map[string]interface{})

	return auditConfig_compliance{
		Enabled:             m["enabled"].(bool),
		InternalConfig:      m["internal_config"].(bool),
		ExternalConfig:      m["external_config"].(bool),
		WriteMetadataOnly:   m["write_metadata_only"].(bool),
		ReadMetadataOnly:    m["read_metadata_only"].(bool),
		WriteLogDiffs:       m["write_log_diffs"].(bool),
		ReadWatchedFields:   expandReadWatchedFields(m["read_watched_field"].(*schema.Set).List()),
		ReadIgnoreUsers:     expandStringList(m["read_ignore_users"].(*schema.Set).List()),
		WriteWatchedIndices: expandStringList(m["write_watched_indices"].(*schema.Set).List()),
		WriteIgnoreUsers:    expandStringList(m["write_ignore_users"].(*schema.Set).List()),
	}
}

func expandReadWatchedFields(fields []interface{}) map[string][]string {
	result := map[string][]string{}

	for _, field := range fields {
		key := field.(map[string]interface{})["index"].(string)
		values := field.(map[string]interface{})["fields"].(*schema.Set).List()
		result[key] = []string{}
		for _, v := range values {
			result[key] = append(result[key], v.(string))
		}
	}
	return result
}

func resourceElasticsearchPutAuditConfig(d *schema.ResourceData, m interface{}) (*putAuditConfigResponse, error) {
	response := new(putAuditConfigResponse)
	auditConfig := auditConfig{
		Enabled:    d.Get("enabled").(bool),
		Audit:      expandAudit(d),
		Compliance: expandCompliance(d),
	}

	auditConfigJSON, err := json.Marshal(auditConfig)
	if err != nil {
		return response, fmt.Errorf("body Error : %s", auditConfigJSON)
	}

	var body json.RawMessage
	esClient, err := getClient(m.(*ProviderConf))
	if err != nil {
		return nil, err
	}
	switch client := esClient.(type) {
	case *elastic7.Client:
		var res *elastic7.Response
		log.Printf("[INFO] put audit config: %+v", auditConfig)
		res, err = client.PerformRequest(context.TODO(), elastic7.PerformRequestOptions{
			Method:           "PUT",
			Path:             "/_plugins/_security/api/audit/config",
			Body:             string(auditConfigJSON),
			RetryStatusCodes: []int{http.StatusInternalServerError},
			Retrier: elastic7.NewBackoffRetrier(
				elastic7.NewExponentialBackoff(100*time.Millisecond, 30*time.Second),
			),
		})
		if err != nil {
			e, ok := err.(*elastic7.Error)
			if !ok {
				log.Printf("[ERROR] expected error to be of type *elastic.Error")
			} else {
				log.Printf("[ERROR] error creating audit config: %v %v %v", res, res.Body, e)
			}
			return response, err
		}

		body = res.Body
	default:
		return response, errors.New("audit config resource not implemented prior to OpenSearch v1")
	}

	if err := json.Unmarshal(body, response); err != nil {
		return response, fmt.Errorf("failed to unmarshal audit config body: %+v: %+v", err, body)
	}

	return response, nil
}

// Response used by the security plugin API (GET method)
type getAuditConfigResponse struct {
	Config auditConfig `json:"config"`
}

// Response sent by the security plugin API (PUT method)
type putAuditConfigResponse struct {
	Message string `json:"message"`
	Status  string `json:"status"`
}

// Payload used by the security plugin API (PUT method)
type auditConfig struct {
	Enabled    bool                   `json:"enabled"`
	Audit      auditConfig_audit      `json:"audit"`
	Compliance auditConfig_compliance `json:"compliance"`
}

type auditConfig_audit struct {
	EnableRest                  bool     `json:"enable_rest"`
	DisabledRestCategories      []string `json:"disabled_rest_categories"`
	EnableTransport             bool     `json:"enable_transport"`
	DisabledTransportCategories []string `json:"disabled_transport_categories"`
	ResolveBulkRequests         bool     `json:"resolve_bulk_requests"`
	LogRequestBody              bool     `json:"log_request_body"`
	ResolveIndices              bool     `json:"resolve_indices"`
	ExcludeSensitiveHeaders     bool     `json:"exclude_sensitive_headers"`
	IgnoreUsers                 []string `json:"ignore_users"`
	IgnoreRequests              []string `json:"ignore_requests"`
}

type auditConfig_compliance struct {
	Enabled             bool                `json:"enabled"`
	InternalConfig      bool                `json:"internal_config"`
	ExternalConfig      bool                `json:"external_config"`
	ReadMetadataOnly    bool                `json:"read_metadata_only"`
	ReadWatchedFields   map[string][]string `json:"read_watched_fields"`
	ReadIgnoreUsers     []string            `json:"read_ignore_users"`
	WriteMetadataOnly   bool                `json:"write_metadata_only"`
	WriteLogDiffs       bool                `json:"write_log_diffs"`
	WriteWatchedIndices []string            `json:"write_watched_indices"`
	WriteIgnoreUsers    []string            `json:"write_ignore_users"`
}
