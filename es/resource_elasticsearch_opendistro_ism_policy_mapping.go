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
		Create: resourceElasticsearchOpenDistroISMPolicyMappingCreate,
		Read:   resourceElasticsearchOpenDistroISMPolicyMappingRead,
		Update: resourceElasticsearchOpenDistroISMPolicyMappingUpdate,
		Delete: resourceElasticsearchOpenDistroISMPolicyMappingDelete,
		Schema: map[string]*schema.Schema{
			"policy_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"indexes": {
				Type:     schema.TypeString,
				Required: true,
			},
			"state": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			"include": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeMap},
			},
			"is_safe": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
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
	indexesList, err := resourceElasticsearchGetOpendistroPolicyMapping(d, m)
	concernIndexes := []string{}
	policyName := d.Get("policy_id").(string)

	if err != nil {
		return err
	}

	// If there is no managed indexes we can remove that resource
	for indexName, parameters := range indexesList {
		if parameters.(map[string]interface{})["index.opendistro.index_state_management.policy_id"] == policyName {
			concernIndexes = append(concernIndexes, indexName)
		}
	}

	log.Printf("[INFO] %+v", concernIndexes)

	if len(concernIndexes) == 0 {
		d.SetId("")
		return nil
	}

	d.SetId(d.Get("indexes").(string))
	err = d.Set("managed_indexes", concernIndexes)
	return err
}

func resourceElasticsearchOpenDistroISMPolicyMappingUpdate(d *schema.ResourceData, m interface{}) error {
	if _, err := resourceElasticsearchPostOpendistroPolicyMapping(d, m, "update_policy"); err != nil {
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
	switch client := m.(type) {
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

func resourceElasticsearchGetOpendistroPolicyMapping(d *schema.ResourceData, m interface{}) (map[string]interface{}, error) {

	response := new(map[string]interface{})
	path, err := uritemplates.Expand("/_opendistro/_ism/explain/{indexes}", map[string]string{
		"indexes": d.Get("indexes").(string),
	})
	if err != nil {
		return *response, fmt.Errorf("error building URL path for policy mapping: %+v", err)
	}

	var body *json.RawMessage
	switch client := m.(type) {
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

	log.Printf("[INFO] %+v", response)

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
