package es

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/olivere/elastic/uritemplates"

	elastic7 "github.com/olivere/elastic/v7"
	elastic6 "gopkg.in/olivere/elastic.v6"
)

var openDistroISMPolicySchema = map[string]*schema.Schema{
	"policy_id": {
		Type:     schema.TypeString,
		Required: true,
		ForceNew: true,
	},
	"body": {
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
}

func resourceOpenSearchISMPolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceElasticsearchOpenDistroISMPolicyCreate,
		Read:   resourceElasticsearchOpenDistroISMPolicyRead,
		Update: resourceElasticsearchOpenDistroISMPolicyUpdate,
		Delete: resourceElasticsearchOpenDistroISMPolicyDelete,
		Schema: openDistroISMPolicySchema,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceElasticsearchOpenDistroISMPolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceElasticsearchOpenDistroISMPolicyCreate,
		Read:   resourceElasticsearchOpenDistroISMPolicyRead,
		Update: resourceElasticsearchOpenDistroISMPolicyUpdate,
		Delete: resourceElasticsearchOpenDistroISMPolicyDelete,
		Schema: openDistroISMPolicySchema,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		DeprecationMessage: "elasticsearch_opendistro_ism_policy is deprecated, please use elasticsearch_opensearch_ism_policy resource instead.",
	}
}

func resourceElasticsearchOpenDistroISMPolicyCreate(d *schema.ResourceData, m interface{}) error {
	if _, err := resourceElasticsearchPutOpenDistroISMPolicy(d, m); err != nil {
		log.Printf("[INFO] Failed to create OpenDistroPolicy: %+v", err)
		return err
	}

	policyID := d.Get("policy_id").(string)
	d.SetId(policyID)
	return resourceElasticsearchOpenDistroISMPolicyRead(d, m)
}

func resourceElasticsearchOpenDistroISMPolicyRead(d *schema.ResourceData, m interface{}) error {
	policyResponse, err := resourceElasticsearchGetOpenDistroISMPolicy(d.Id(), m)

	if err != nil {
		if elastic6.IsNotFound(err) || elastic7.IsNotFound(err) {
			log.Printf("[WARN] OpenDistroPolicy (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	bodyString, err := json.Marshal(policyResponse.Policy)
	if err != nil {
		return err
	}

	// Need encapsulation as the response from the GET is different than the one in the PUT
	bodyStringNormalized, _ := structure.NormalizeJsonString(fmt.Sprintf("{\"policy\": %+s}", string(bodyString)))

	if err := d.Set("policy_id", policyResponse.PolicyID); err != nil {
		return fmt.Errorf("error setting policy_id: %s", err)
	}
	if err := d.Set("body", bodyStringNormalized); err != nil {
		return fmt.Errorf("error setting body: %s", err)
	}
	if err := d.Set("primary_term", policyResponse.PrimaryTerm); err != nil {
		return fmt.Errorf("error setting primary_term: %s", err)
	}
	if err := d.Set("seq_no", policyResponse.SeqNo); err != nil {
		return fmt.Errorf("error setting seq_no: %s", err)
	}

	return nil
}

func resourceElasticsearchOpenDistroISMPolicyUpdate(d *schema.ResourceData, m interface{}) error {
	if _, err := resourceElasticsearchPutOpenDistroISMPolicy(d, m); err != nil {
		return err
	}

	return resourceElasticsearchOpenDistroISMPolicyRead(d, m)
}

func resourceElasticsearchOpenDistroISMPolicyDelete(d *schema.ResourceData, m interface{}) error {
	path, err := uritemplates.Expand("/_opendistro/_ism/policies/{policy_id}", map[string]string{
		"policy_id": d.Id(),
	})
	if err != nil {
		return fmt.Errorf("error building URL path for policy: %+v", err)
	}

	esClient, err := getClient(m.(*ProviderConf))
	if err != nil {
		return err
	}
	switch client := esClient.(type) {
	case *elastic7.Client:
		_, err = client.PerformRequest(context.TODO(), elastic7.PerformRequestOptions{
			Method:           "DELETE",
			Path:             path,
			RetryStatusCodes: []int{http.StatusConflict},
			Retrier: elastic7.NewBackoffRetrier(
				elastic7.NewExponentialBackoff(100*time.Millisecond, 30*time.Second),
			),
		})

		if err != nil {
			return fmt.Errorf("error deleting policy: %+v : %+v", path, err)
		}
	case *elastic6.Client:
		_, err = client.PerformRequest(context.TODO(), elastic6.PerformRequestOptions{
			Method: "DELETE",
			Path:   path,
		})

		if err != nil {
			return fmt.Errorf("error deleting policy: %+v : %+v", path, err)
		}
	default:
		err = errors.New("policy resource not implemented prior to Elastic v6")
	}

	return err
}

func resourceElasticsearchGetOpenDistroISMPolicy(policyID string, m interface{}) (GetPolicyResponse, error) {
	var err error
	response := new(GetPolicyResponse)

	path, err := uritemplates.Expand("/_opendistro/_ism/policies/{policy_id}", map[string]string{
		"policy_id": policyID,
	})

	if err != nil {
		return *response, fmt.Errorf("error building URL path for policy: %+v", err)
	}

	var body *json.RawMessage
	esClient, err := getClient(m.(*ProviderConf))
	if err != nil {
		return *response, err
	}
	switch client := esClient.(type) {
	case *elastic7.Client:
		var res *elastic7.Response
		res, err = client.PerformRequest(context.TODO(), elastic7.PerformRequestOptions{
			Method: "GET",
			Path:   path,
		})

		if err != nil {
			return *response, fmt.Errorf("error getting policy: %+v : %+v", path, err)
		}
		body = &res.Body
	case *elastic6.Client:
		var res *elastic6.Response
		res, err = client.PerformRequest(context.TODO(), elastic6.PerformRequestOptions{
			Method: "GET",
			Path:   path,
		})

		if err != nil {
			return *response, fmt.Errorf("error getting policy: %+v : %+v", path, err)
		}
		body = &res.Body
	default:
		err = errors.New("policy resource not implemented prior to Elastic v6")
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

func resourceElasticsearchPutOpenDistroISMPolicy(d *schema.ResourceData, m interface{}) (*PutPolicyResponse, error) {
	response := new(PutPolicyResponse)
	policyJSON := d.Get("body").(string)
	seq := d.Get("seq_no").(int)
	primTerm := d.Get("primary_term").(int)
	params := url.Values{}

	if seq >= 0 && primTerm > 0 {
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
	esClient, err := getClient(m.(*ProviderConf))
	if err != nil {
		return nil, err
	}
	switch client := esClient.(type) {
	case *elastic7.Client:
		var res *elastic7.Response
		res, err = client.PerformRequest(context.TODO(), elastic7.PerformRequestOptions{
			Method:           "PUT",
			Path:             path,
			Params:           params,
			Body:             string(policyJSON),
			RetryStatusCodes: []int{http.StatusConflict},
			Retrier: elastic7.NewBackoffRetrier(
				elastic7.NewExponentialBackoff(100*time.Millisecond, 30*time.Second),
			),
		})
		if err != nil {
			return response, fmt.Errorf("error putting policy: %+v : %+v : %+v", path, policyJSON, err)
		}
		body = &res.Body
	case *elastic6.Client:
		var res *elastic6.Response
		res, err = client.PerformRequest(context.TODO(), elastic6.PerformRequestOptions{
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
		err = errors.New("policy resource not implemented prior to Elastic v6")
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
