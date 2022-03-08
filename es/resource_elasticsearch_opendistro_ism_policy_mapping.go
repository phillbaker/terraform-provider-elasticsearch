package es

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/olivere/elastic/uritemplates"

	elastic7 "github.com/olivere/elastic/v7"
)

var openDistroISMPolicyMappingSchema = map[string]*schema.Schema{
	"policy_id": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "The name of the policy.",
	},
	"indexes": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "Name of the index to apply the policy to. You can use an index pattern to update multiple indices at once.",
	},
	"state": {
		Type:        schema.TypeString,
		Optional:    true,
		Default:     "",
		Description: "After a change in policy takes place, specify the state for the index to transition to",
	},
	"include": {
		Type:        schema.TypeSet,
		Optional:    true,
		Elem:        &schema.Schema{Type: schema.TypeMap},
		Description: "When updating multiple indices, you might want to include a state filter to only affect certain managed indices. The background process only applies the change if the index is currently in the state specified.",
	},
	"is_safe": {
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     false,
		Description: "",
	},
	"managed_indexes": {
		Type:     schema.TypeSet,
		Optional: true,
		Computed: true,
		Elem:     &schema.Schema{Type: schema.TypeString},
	},
}

func resourceOpenSearchISMPolicyMapping() *schema.Resource {
	return &schema.Resource{
		Description: "Provides an Elasticsearch Open Distro Index State Management (ISM) policy. Please refer to the Open Distro [ISM documentation](https://opendistro.github.io/for-elasticsearch-docs/docs/ism/) for details.",
		Create:      resourceElasticsearchOpenDistroISMPolicyMappingCreate,
		Read:        resourceElasticsearchOpenDistroISMPolicyMappingRead,
		Update:      resourceElasticsearchOpenDistroISMPolicyMappingUpdate,
		Delete:      resourceElasticsearchOpenDistroISMPolicyMappingDelete,
		Schema:      openDistroISMPolicyMappingSchema,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
		},
		DeprecationMessage: "elasticsearch_opensearch_ism_policy_mapping is deprecated in Opensearch 1.x please use ism_template attribute in policies instead.",
	}
}

func resourceElasticsearchOpenDistroISMPolicyMapping() *schema.Resource {
	return &schema.Resource{
		Description: "Provides an Elasticsearch Open Distro Index State Management (ISM) policy. Please refer to the Open Distro [ISM documentation](https://opendistro.github.io/for-elasticsearch-docs/docs/ism/) for details.",
		Create:      resourceElasticsearchOpenDistroISMPolicyMappingCreate,
		Read:        resourceElasticsearchOpenDistroISMPolicyMappingRead,
		Update:      resourceElasticsearchOpenDistroISMPolicyMappingUpdate,
		Delete:      resourceElasticsearchOpenDistroISMPolicyMappingDelete,
		Schema:      openDistroISMPolicyMappingSchema,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
		},
		DeprecationMessage: "elasticsearch_opendistro_ism_policy_mapping is deprecated in ODFE 1.13.x please use ism_template attribute in policies instead.",
	}
}

func resourceElasticsearchOpenDistroISMPolicyMappingCreate(d *schema.ResourceData, m interface{}) error {
	resp, err := resourceElasticsearchPostOpendistroPolicyMapping(d, m, "add")
	log.Printf("[INFO] resourceElasticsearchOpenDistroISMPolicyMappingCreate %+v", resp)
	if err != nil {
		return err
	}

	indexPattern := d.Get("indexes").(string)
	policyID := d.Get("policy_id").(string)

	return resource.RetryContext(context.TODO(), d.Timeout(schema.TimeoutCreate), resourceElasticsearchOpenDistroISMPolicyMappingRetry(indexPattern, policyID, d, m))
}

// From https://opendistro.github.io/for-elasticsearch-docs/docs/im/ism/api/#update-managed-index-policy
// A policy change is an asynchronous background process. The changes are
// queued and are not executed immediately by the background process. This
// delay in execution protects the currently running managed indices from
// being put into a broken state. If the policy you are changing to has only
// some small configuration changes, then the change takes place immediately.
// If the change modifies the state, actions, or the order of actions of the
// current state the index is in, then the change happens at the end of its
// current state before transitioning to a new state.
func resourceElasticsearchOpenDistroISMPolicyMappingRetry(indexPattern string, policyID string, d *schema.ResourceData, m interface{}) func() *resource.RetryError {
	return func() *resource.RetryError {
		indices, err := resourceElasticsearchOpendistroPolicyIndices(indexPattern, policyID, m)

		if err != nil {
			log.Printf("[INFO] error on retrieving indices %+v", err)
			return resource.NonRetryableError(err)
		}

		// This isn't a great test, index patterns with a glob could in theory
		// match zero indices or more
		if len(indices) == 0 {
			return resource.RetryableError(fmt.Errorf("Expected at least one index to be mapped, but found %d", len(indices)))
		}

		err = resourceElasticsearchOpenDistroISMPolicyMappingRead(d, m)
		log.Printf("[INFO] resourceElasticsearchOpenDistroISMPolicyMappingRetry error %+v", err)
		if err != nil {
			return resource.NonRetryableError(err)
		}
		return nil
	}
}

func resourceElasticsearchOpendistroPolicyIndices(indexPattern string, policyID string, m interface{}) ([]string, error) {
	indices, err := resourceElasticsearchGetOpendistroPolicyMapping(indexPattern, m)
	mappedIndexes := []string{}

	if err != nil {
		return mappedIndexes, err
	}

	for indexName, parameters := range indices {
		p, ok := parameters.(map[string]interface{})
		if ok && p["index.opendistro.index_state_management.policy_id"] == policyID {
			mappedIndexes = append(mappedIndexes, indexName)
		} else if ok && p["index.plugins.index_state_management.policy_id"] == policyID {
			mappedIndexes = append(mappedIndexes, indexName)
		}
	}

	log.Printf("[INFO] resourceElasticsearchOpendistroPolicyIndices %+v %+v %+v", indexPattern, indices, mappedIndexes)
	return mappedIndexes, nil
}

func resourceElasticsearchOpenDistroISMPolicyMappingRead(d *schema.ResourceData, m interface{}) error {
	indexPattern := d.Get("indexes").(string)
	policyID := d.Get("policy_id").(string)

	indices, err := resourceElasticsearchOpendistroPolicyIndices(indexPattern, policyID, m)
	if err != nil {
		log.Printf("[INFO] resourceElasticsearchOpenDistroISMPolicyMappingRead %+v %+v", indices, err)
		return err
	}

	// If there is no managed indices, remove the resource
	if len(indices) == 0 {
		log.Printf("[INFO] no managed indices, removing mapping")
		d.SetId("")
		return nil
	}

	d.SetId(d.Get("indexes").(string))

	ds := &resourceDataSetter{d: d}
	ds.set("managed_indexes", indices)

	return ds.err
}

func resourceElasticsearchOpenDistroISMPolicyMappingUpdate(d *schema.ResourceData, m interface{}) error {
	if _, err := resourceElasticsearchPostOpendistroPolicyMapping(d, m, "change_policy"); err != nil {
		if elastic7.IsNotFound(err) {
			log.Printf("[WARN] OpendistroPolicyMapping (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	indexPattern := d.Get("indexes").(string)
	policyID := d.Get("policy_id").(string)

	return resource.RetryContext(context.TODO(), d.Timeout(schema.TimeoutUpdate), resourceElasticsearchOpenDistroISMPolicyMappingRetry(indexPattern, policyID, d, m))
}

func resourceElasticsearchOpenDistroISMPolicyMappingDelete(d *schema.ResourceData, m interface{}) error {
	if _, err := resourceElasticsearchPostOpendistroPolicyMapping(d, m, "remove"); err != nil {
		return err
	}

	d.SetId("")

	return nil
}

func resourceElasticsearchPostOpendistroPolicyMapping(d *schema.ResourceData, m interface{}, action string) (*PolicyMappingResponse, error) {
	response := new(PolicyMappingResponse)
	requestBody := ""

	switch action {
	case "remove":
		requestBody = ""
	case "add":
		mapping, err := json.Marshal(PolicyMapping{
			PolicyID: d.Get("policy_id").(string),
		})
		requestBody = string(mapping)

		if err != nil {
			return response, err
		}
	default:
		include, _ := d.GetOk("include")
		mapping, err := json.Marshal(PolicyMapping{
			PolicyID: d.Get("policy_id").(string),
			State:    d.Get("state").(string),
			IsSafe:   d.Get("is_safe").(bool),
			Include:  include.(*schema.Set).List(),
		})
		requestBody = string(mapping)

		if err != nil {
			return response, err
		}

	}

	path, err := uritemplates.Expand("/_opendistro/_ism/{action}/{indexes}", map[string]string{
		"indexes": d.Get("indexes").(string),
		"action":  action,
	})
	if err != nil {
		return response, fmt.Errorf("error building URL path for policy: %+v", err)
	}

	var body *json.RawMessage
	esClient, err := getClient(m.(*ProviderConf))
	if err != nil {
		return nil, err
	}
	switch client := esClient.(type) {
	case *elastic7.Client:
		var res *elastic7.Response
		res, err = client.PerformRequest(context.TODO(), elastic7.PerformRequestOptions{
			Method: "POST",
			Path:   path,
			Body:   requestBody,
		})
		if err != nil {
			return response, fmt.Errorf("error posting policy attachment: %+v : %+v : %+v", path, requestBody, err)
		}
		body = &res.Body
	default:
		err = errors.New("policy resource not implemented prior to Elastic v7")
	}

	if err != nil {
		return response, fmt.Errorf("error creating policy mapping: %+v", err)
	}

	if err := json.Unmarshal(*body, response); err != nil {
		return response, fmt.Errorf("error unmarshalling policy body: %+v: %+v", err, body)
	}

	return response, nil
}

func resourceElasticsearchGetOpendistroPolicyMapping(indexPattern string, m interface{}) (map[string]interface{}, error) {
	response := new(map[string]interface{})
	path, err := uritemplates.Expand("/_opendistro/_ism/explain/{index_pattern}", map[string]string{
		"index_pattern": indexPattern,
	})
	if err != nil {
		return *response, fmt.Errorf("error building URL path for policy mapping: %+v", err)
	}

	var body *json.RawMessage
	esClient, err := getClient(m.(*ProviderConf))
	if err != nil {
		return nil, err
	}
	switch client := esClient.(type) {
	case *elastic7.Client:
		var res *elastic7.Response
		res, err = client.PerformRequest(context.TODO(), elastic7.PerformRequestOptions{
			Method: "GET",
			Path:   path,
		})
		if err != nil {
			return *response, fmt.Errorf("error getting policy attachment: %+v, %w", path, err)
		}
		body = &res.Body
	default:
		err = errors.New("policy mapping resource not implemented prior to Elastic v7")
	}

	if err != nil {
		return *response, fmt.Errorf("error creating policy mapping: %+v", err)
	}

	if err := json.Unmarshal(*body, response); err != nil {
		return *response, fmt.Errorf("error unmarshalling policy explain body: %+v: %+v", err, body)
	}

	return *response, nil
}

type PolicyMappingResponse struct {
	UpdatedIndices int           `json:"updated_indices"`
	Failures       bool          `json:"failures"`
	FailedIndices  []interface{} `json:"failed_indices"`
}

type PolicyMapping struct {
	PolicyID string        `json:"policy_id"`
	State    string        `json:"state"`
	IsSafe   bool          `json:"is_safe"`
	Include  []interface{} `json:"include"`
}
