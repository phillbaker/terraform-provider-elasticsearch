package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/url"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/structure"
	"github.com/olivere/elastic/uritemplates"

	elastic7 "github.com/olivere/elastic/v7"
)

func resourceElasticsearchOpendistroPolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceElasticsearchOpendistroPolicyCreate,
		Read:   resourceElasticsearchOpendistroPolicyRead,
		Update: resourceElasticsearchOpendistroPolicyCreate,
		Delete: resourceElasticsearchOpendistroPolicyDelete,
		Schema: map[string]*schema.Schema{
			"policy_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"body": &schema.Schema{
				Type:             schema.TypeString,
				Required:         true,
				DiffSuppressFunc: diffSuppressPolicy,
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},
			"primary_term": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"seq_no": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
		},
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
	}
}

func resourceElasticsearchOpendistroPolicyCreate(d *schema.ResourceData, m interface{}) error {
	if _, err := resourceElasticsearchPutOpendistroPolicy(d, m); err != nil {
		return err
	}

	return resourceElasticsearchOpendistroPolicyRead(d, m)
}

func resourceElasticsearchOpendistroPolicyRead(d *schema.ResourceData, m interface{}) error {
	policyID := d.Get("policy_id").(string)
	policyResponse, err := resourceElasticsearchGetOpendistroPolicy(policyID, m)

	if err != nil {
		if elastic7.IsNotFound(err) {
			log.Printf("[WARN] OpendistroPolicy (%s) not found, removing from state", policyID)
			d.SetId("")
			return nil
		}
		return err
	}

	bodyString, err := json.Marshal(policyResponse.Policy)
	// Need encapsulation as the reponse from the GET is different than the one in the PUT
	bodyStringNormalized, _ := structure.NormalizeJsonString(fmt.Sprintf("{\"policy\": %+s}", string(bodyString)))

	if err != nil {
		return err
	}
	d.SetId(policyResponse.PolicyID)
	d.Set("body", bodyStringNormalized)
	d.Set("primary_term", policyResponse.PrimaryTerm)
	d.Set("seq_no", policyResponse.SeqNo)

	return nil
}

func resourceElasticsearchOpendistroPolicyDelete(d *schema.ResourceData, m interface{}) error {
	var err error

	path, err := uritemplates.Expand("/_opendistro/_ism/policies/{policy_id}", map[string]string{
		"policy_id": d.Id(),
	})
	if err != nil {
		return fmt.Errorf("error building URL path for policy: %+v", err)
	}

	switch m.(type) {
	case *elastic7.Client:
		client := m.(*elastic7.Client)
		_, err = client.PerformRequest(context.TODO(), elastic7.PerformRequestOptions{
			Method: "DELETE",
			Path:   path,
		})

		if err != nil {
			return fmt.Errorf("error deleting policy: %+v : %+v", path, err)
		}
	default:
		err = errors.New("policy resource not implemented prior to Elastic v7")
	}

	return err
}

func resourceElasticsearchGetOpendistroPolicy(policyID string, m interface{}) (GetPolicyResponse, error) {
	var err error
	response := new(GetPolicyResponse)

	path, err := uritemplates.Expand("/_opendistro/_ism/policies/{policy_id}", map[string]string{
		"policy_id": policyID,
	})

	if err != nil {
		return *response, fmt.Errorf("error building URL path for policy: %+v", err)
	}

	var body *json.RawMessage
	switch m.(type) {
	case *elastic7.Client:
		client := m.(*elastic7.Client)
		var res *elastic7.Response
		res, err = client.PerformRequest(context.TODO(), elastic7.PerformRequestOptions{
			Method: "GET",
			Path:   path,
		})

		if err != nil {
			return *response, fmt.Errorf("error getting policy: %+v : %+v", path, err)
		}
		body = &res.Body
	default:
		err = errors.New("policy resource not implemented prior to Elastic v7")
	}

	if err != nil {
		return *response, err
	}

	if err := json.Unmarshal(*body, &response); err != nil {
		return *response, fmt.Errorf("error unmarshalling policy body: %+v: %+v", err, body)
	}

	normalizePolicy(response.Policy)

	return *response, err
}

func resourceElasticsearchPutOpendistroPolicy(d *schema.ResourceData, m interface{}) (*PutPolicyResponse, error) {
	response := new(PutPolicyResponse)
	policyJSON := d.Get("body").(string)
	seq := d.Get("seq_no").(int)
	primTerm := d.Get("primary_term").(int)
	params := url.Values{}

	if seq > 0 && primTerm > 0 {
		params.Set("if_seq_no", strconv.Itoa(seq))
		params.Set("if_primary_term", strconv.Itoa(primTerm))
	}

	path, err := uritemplates.Expand("/_opendistro/_ism/policies/{policy_id}", map[string]string{
		"policy_id": d.Get("policy_id").(string),
	})
	if err != nil {
		return response, fmt.Errorf("error building URL path for policy: %+v", err)
	}

	var body *json.RawMessage
	switch m.(type) {
	case *elastic7.Client:
		client := m.(*elastic7.Client)
		var res *elastic7.Response
		res, err = client.PerformRequest(context.TODO(), elastic7.PerformRequestOptions{
			Method: "PUT",
			Path:   path,
			Params: params,
			Body:   string(policyJSON),
		})
		if err != nil {
			return response, fmt.Errorf("error putting policy: %+v : %+v : %+v", path, policyJSON, err)
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

type GetPolicyResponse struct {
	PolicyID    string                 `json:"_id"`
	Version     int                    `json:"_version"`
	PrimaryTerm int                    `json:"_primary_term"`
	SeqNo       int                    `json:"_seq_no"`
	Policy      map[string]interface{} `json:"policy"`
}

type PutPolicyResponse struct {
	PolicyID    string `json:"_id"`
	Version     int    `json:"_version"`
	PrimaryTerm int    `json:"_primary_term"`
	SeqNo       int    `json:"_seq_no"`
	Policy      struct {
		Policy map[string]interface{} `json:"policy"`
	} `json:"policy"`
}
