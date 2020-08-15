package es

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/olivere/elastic/uritemplates"

	elastic7 "github.com/olivere/elastic/v7"
	elastic6 "gopkg.in/olivere/elastic.v6"
)

var openDistroMonitorSchema = map[string]*schema.Schema{
	"body": {
		Type:             schema.TypeString,
		Required:         true,
		DiffSuppressFunc: diffSuppressMonitor,
		StateFunc: func(v interface{}) string {
			json, _ := structure.NormalizeJsonString(v)
			return json
		},
		ValidateFunc: validation.StringIsJSON,
	},
}

func resourceElasticsearchDeprecatedMonitor() *schema.Resource {
	return &schema.Resource{
		Create: resourceElasticsearchOpenDistroMonitorCreate,
		Read:   resourceElasticsearchOpenDistroMonitorRead,
		Update: resourceElasticsearchOpenDistroMonitorUpdate,
		Delete: resourceElasticsearchOpenDistroMonitorDelete,
		Schema: openDistroMonitorSchema,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		DeprecationMessage: "elasticsearch_monitor is deprecated, please use elasticsearch_opendistro_monitor resource instead.",
	}
}

func resourceElasticsearchOpenDistroMonitor() *schema.Resource {
	return &schema.Resource{
		Create: resourceElasticsearchOpenDistroMonitorCreate,
		Read:   resourceElasticsearchOpenDistroMonitorRead,
		Update: resourceElasticsearchOpenDistroMonitorUpdate,
		Delete: resourceElasticsearchOpenDistroMonitorDelete,
		Schema: openDistroMonitorSchema,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceElasticsearchOpenDistroMonitorCreate(d *schema.ResourceData, m interface{}) error {
	res, err := resourceElasticsearchOpenDistroPostMonitor(d, m)

	if err != nil {
		log.Printf("[INFO] Failed to put monitor: %+v", err)
		return err
	}

	d.SetId(res.ID)
	log.Printf("[INFO] Object ID: %s", d.Id())

	// Although we receive the full monitor in the response to the POST,
	// OpenDistro seems to add default values to the ojbect after the resource
	// is saved, e.g. adjust_pure_negative, boost values
	return resourceElasticsearchOpenDistroMonitorRead(d, m)
}

func resourceElasticsearchOpenDistroMonitorRead(d *schema.ResourceData, m interface{}) error {
	res, err := resourceElasticsearchOpenDistroGetMonitor(d.Id(), m)

	if elastic6.IsNotFound(err) || elastic7.IsNotFound(err) {
		log.Printf("[WARN] Monitor (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return err
	}

	d.SetId(res.ID)

	monitorJson, err := json.Marshal(res.Monitor)
	if err != nil {
		return err
	}
	monitorJsonNormalized, err := structure.NormalizeJsonString(string(monitorJson))
	if err != nil {
		return err
	}
	err = d.Set("body", monitorJsonNormalized)
	return err
}

func resourceElasticsearchOpenDistroMonitorUpdate(d *schema.ResourceData, m interface{}) error {
	_, err := resourceElasticsearchOpenDistroPutMonitor(d, m)

	if err != nil {
		return err
	}

	return resourceElasticsearchOpenDistroMonitorRead(d, m)
}

func resourceElasticsearchOpenDistroMonitorDelete(d *schema.ResourceData, m interface{}) error {
	var err error

	path, err := uritemplates.Expand("/_opendistro/_alerting/monitors/{id}", map[string]string{
		"id": d.Id(),
	})
	if err != nil {
		return fmt.Errorf("error building URL path for monitor: %+v", err)
	}

	esClient, err := getClient(m.(*ProviderConf))
	if err != nil {
		return err
	}
	switch client := esClient.(type) {
	case *elastic7.Client:
		_, err = client.PerformRequest(context.TODO(), elastic7.PerformRequestOptions{
			Method: "DELETE",
			Path:   path,
		})
	case *elastic6.Client:
		_, err = client.PerformRequest(context.TODO(), elastic6.PerformRequestOptions{
			Method: "DELETE",
			Path:   path,
		})
	default:
		err = errors.New("monitor resource not implemented prior to Elastic v6")
	}

	return err
}

func resourceElasticsearchOpenDistroGetMonitor(monitorID string, m interface{}) (*monitorResponse, error) {
	var err error
	response := new(monitorResponse)

	path, err := uritemplates.Expand("/_opendistro/_alerting/monitors/{id}", map[string]string{
		"id": monitorID,
	})
	if err != nil {
		return response, fmt.Errorf("error building URL path for monitor: %+v", err)
	}

	var body json.RawMessage
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
		body = res.Body
	case *elastic6.Client:
		var res *elastic6.Response
		res, err = client.PerformRequest(context.TODO(), elastic6.PerformRequestOptions{
			Method: "GET",
			Path:   path,
		})
		body = res.Body
	default:
		err = errors.New("monitor resource not implemented prior to Elastic v6")
	}

	if err != nil {
		return response, err
	}

	if err := json.Unmarshal(body, response); err != nil {
		return response, fmt.Errorf("error unmarshalling monitor body: %+v: %+v", err, body)
	}
	normalizeMonitor(response.Monitor)
	return response, err
}

func resourceElasticsearchOpenDistroPostMonitor(d *schema.ResourceData, m interface{}) (*monitorResponse, error) {
	monitorJSON := d.Get("body").(string)

	var err error
	response := new(monitorResponse)

	path := "/_opendistro/_alerting/monitors/"

	var body json.RawMessage
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
			Body:   monitorJSON,
		})
		body = res.Body
	case *elastic6.Client:
		var res *elastic6.Response
		res, err = client.PerformRequest(context.TODO(), elastic6.PerformRequestOptions{
			Method: "POST",
			Path:   path,
			Body:   monitorJSON,
		})
		body = res.Body
	default:
		err = errors.New("monitor resource not implemented prior to Elastic v6")
	}

	if err != nil {
		return response, err
	}

	if err := json.Unmarshal(body, response); err != nil {
		return response, fmt.Errorf("error unmarshalling monitor body: %+v: %+v", err, body)
	}
	normalizeMonitor(response.Monitor)
	return response, nil
}

func resourceElasticsearchOpenDistroPutMonitor(d *schema.ResourceData, m interface{}) (*monitorResponse, error) {
	monitorJSON := d.Get("body").(string)

	var err error
	response := new(monitorResponse)

	path, err := uritemplates.Expand("/_opendistro/_alerting/monitors/{id}", map[string]string{
		"id": d.Id(),
	})
	if err != nil {
		return response, fmt.Errorf("error building URL path for monitor: %+v", err)
	}

	var body json.RawMessage
	esClient, err := getClient(m.(*ProviderConf))
	if err != nil {
		return nil, err
	}
	switch client := esClient.(type) {
	case *elastic7.Client:
		var res *elastic7.Response
		res, err = client.PerformRequest(context.TODO(), elastic7.PerformRequestOptions{
			Method: "PUT",
			Path:   path,
			Body:   monitorJSON,
		})
		body = res.Body
	case *elastic6.Client:
		var res *elastic6.Response
		res, err = client.PerformRequest(context.TODO(), elastic6.PerformRequestOptions{
			Method: "PUT",
			Path:   path,
			Body:   monitorJSON,
		})
		body = res.Body
	default:
		err = errors.New("monitor resource not implemented prior to Elastic v6")
	}

	if err != nil {
		return response, err
	}

	if err := json.Unmarshal(body, response); err != nil {
		return response, fmt.Errorf("error unmarshalling monitor body: %+v: %+v", err, body)
	}

	return response, nil
}

type monitorResponse struct {
	Version int                    `json:"_version"`
	ID      string                 `json:"_id"`
	Monitor map[string]interface{} `json:"monitor"`
}
