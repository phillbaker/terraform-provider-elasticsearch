package es

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/olivere/elastic/uritemplates"

	elastic7 "github.com/olivere/elastic/v7"
)

func resourceElasticsearchOpenDistroISMPolicyMapping() *schema.Resource {
	return &schema.Resource{
		Description: "Provides an Elasticsearch Open Distro ISM policy. Please refer to the Open Distro [ISM documentation](https://opendistro.github.io/for-elasticsearch-docs/docs/ism/) for details.",
		Create:      resourceElasticsearchOpenDistroISMPolicyMappingCreate,
		Read:        resourceElasticsearchOpenDistroISMPolicyMappingRead,
		Update:      resourceElasticsearchOpenDistroISMPolicyMappingUpdate,
		Delete:      resourceElasticsearchOpenDistroISMPolicyMappingDelete,
		Schema: map[string]*schema.Schema{
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
		},
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
	}
}

func resourceElasticsearchOpenDistroISMPolicyMappingCreate(d *schema.ResourceData, m interface{}) error {
	if _, err := resourceElasticsearchPostOpendistroPolicyMapping(d, m, "add"); err != nil {
		return err
	}

	return resourceElasticsearchOpenDistroISMPolicyMappingRead(d, m)
}

func resourceElasticsearchOpenDistroISMPolicyMappingRead(d *schema.ResourceData, m interface{}) error {
	indexPattern := d.Get("indexes").(string)
	indices, err := resourceElasticsearchGetOpendistroPolicyMapping(indexPattern, m)
	concernIndexes := []string{}
	policyName := d.Get("policy_id").(string)

	if err != nil {
		return err
	}

	// If there is no managed indexes we can remove that resource
	for indexName, parameters := range indices {
		p, ok := parameters.(map[string]interface{})
		if ok && p["index.opendistro.index_state_management.policy_id"] == policyName {
			concernIndexes = append(concernIndexes, indexName)
		}
	}

	log.Printf("[INFO] resourceElasticsearchOpenDistroISMPolicyMappingRead %+v %+v %+v", indexPattern, indices, concernIndexes)

	// If there is no managed indices, remove the resource
	if len(concernIndexes) == 0 {
		log.Printf("[INFO] no managed indices, removing mapping")
		d.SetId("")
		return nil
	}

	d.SetId(d.Get("indexes").(string))

	ds := &resourceDataSetter{d: d}
	ds.set("managed_indexes", concernIndexes)

	// TODO
	// state
	// include
	// is_safe

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

	return resourceElasticsearchOpenDistroISMPolicyMappingRead(d, m)
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
			return response, fmt.Errorf("error posting policy attachement: %+v : %+v : %+v", path, requestBody, err)
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
			return *response, fmt.Errorf("error getting policy attachement: %+v : %+v", path, err)
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
	Failures      bool          `json:"failures"`
	FailedIndices []interface{} `json:"failed_indices"`
}

type PolicyMapping struct {
	PolicyID string        `json:"policy_id"`
	State    string        `json:"state"`
	IsSafe   bool          `json:"is_safe"`
	Include  []interface{} `json:"include"`
}
