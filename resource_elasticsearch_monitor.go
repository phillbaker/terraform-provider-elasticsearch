package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/olivere/elastic/uritemplates"

	elastic7 "github.com/olivere/elastic/v7"
	elastic6 "gopkg.in/olivere/elastic.v6"
)

func resourceElasticsearchMonitor() *schema.Resource {
	return &schema.Resource{
		Create: resourceElasticsearchMonitorCreate,
		Read:   resourceElasticsearchMonitorRead,
		Update: resourceElasticsearchMonitorUpdate,
		Delete: resourceElasticsearchMonitorDelete,
		Schema: map[string]*schema.Schema{
			"body": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
		},
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
	}
}

func resourceElasticsearchMonitorCreate(d *schema.ResourceData, m interface{}) error {
	res, err := resourceElasticsearchPostMonitor(d, m)

	if err != nil {
		log.Printf("[INFO] Failed to put monitor: %+v", err)
		return err
	}

	d.SetId(res.ID)
	d.Set("body", res.Monitor)
	log.Printf("[INFO] Object ID: %s", d.Id())

	return nil
}

func resourceElasticsearchMonitorRead(d *schema.ResourceData, m interface{}) error {
	res, err := resourceElasticsearchGetMonitor(d.Id(), m)

	if elastic6.IsNotFound(err) || elastic7.IsNotFound(err) {
		log.Printf("[WARN] Monitor (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return err
	}

	d.Set("body", res.Monitor)
	d.SetId(res.ID)

	return nil
}

func resourceElasticsearchMonitorUpdate(d *schema.ResourceData, m interface{}) error {
	_, err := resourceElasticsearchPutMonitor(d, m)

	if err != nil {
		return err
	}

	return resourceElasticsearchMonitorRead(d, m)
}

func resourceElasticsearchMonitorDelete(d *schema.ResourceData, m interface{}) error {
	var err error

	path, err := uritemplates.Expand("/_opendistro/_alerting/monitors/{id}", map[string]string{
		"id": d.Id(),
	})
	if err != nil {
		return fmt.Errorf("error building URL path for monitor: %+v", err)
	}

	switch m.(type) {
	case *elastic7.Client:
		client := m.(*elastic7.Client)
		_, err = client.PerformRequest(context.TODO(), elastic7.PerformRequestOptions{
			Method: "DELETE",
			Path:   path,
		})
	case *elastic6.Client:
		client := m.(*elastic6.Client)
		_, err = client.PerformRequest(context.TODO(), elastic6.PerformRequestOptions{
			Method: "DELETE",
			Path:   path,
		})
	default:
		err = errors.New("monitor resource not implemented prior to Elastic v6")
	}

	return err
}

func resourceElasticsearchGetMonitor(monitorID string, m interface{}) (*monitorResponse, error) {
	var err error
	response := new(monitorResponse)

	path, err := uritemplates.Expand("/_opendistro/_alerting/monitors/{id}", map[string]string{
		"id": monitorID,
	})
	if err != nil {
		return response, fmt.Errorf("error building URL path for monitor: %+v", err)
	}

	var body json.RawMessage
	switch m.(type) {
	case *elastic7.Client:
		client := m.(*elastic7.Client)
		var res *elastic7.Response
		res, err = client.PerformRequest(context.TODO(), elastic7.PerformRequestOptions{
			Method: "GET",
			Path:   path,
		})
		body = res.Body
	case *elastic6.Client:
		client := m.(*elastic6.Client)
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

	return response, err
}

func resourceElasticsearchPostMonitor(d *schema.ResourceData, m interface{}) (*monitorResponse, error) {
	monitorJSON := d.Get("body").(string)

	var err error
	response := new(monitorResponse)

	path := "/_opendistro/_alerting/monitors/"

	var body json.RawMessage
	switch m.(type) {
	case *elastic7.Client:
		client := m.(*elastic7.Client)
		var res *elastic7.Response
		res, err = client.PerformRequest(context.TODO(), elastic7.PerformRequestOptions{
			Method: "POST",
			Path:   path,
			Body:   monitorJSON,
		})
		body = res.Body
	case *elastic6.Client:
		client := m.(*elastic6.Client)
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

	return response, nil
}

func resourceElasticsearchPutMonitor(d *schema.ResourceData, m interface{}) (*monitorResponse, error) {
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
	switch m.(type) {
	case *elastic7.Client:
		client := m.(*elastic7.Client)
		var res *elastic7.Response
		res, err = client.PerformRequest(context.TODO(), elastic7.PerformRequestOptions{
			Method: "PUT",
			Path:   path,
			Body:   monitorJSON,
		})
		body = res.Body
	case *elastic6.Client:
		client := m.(*elastic6.Client)
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
	Version int         `json:"_version"`
	ID      string      `json:"_id"`
	Monitor interface{} `json:"monitor"`
}
