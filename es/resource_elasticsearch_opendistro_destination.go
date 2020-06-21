package es

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/olivere/elastic/uritemplates"

	elastic7 "github.com/olivere/elastic/v7"
	elastic6 "gopkg.in/olivere/elastic.v6"
)

const DESTINATION_TYPE = "_doc"
const DESTINATION_INDEX = ".opendistro-alerting-config"

var openDistroDestinationSchema = map[string]*schema.Schema{
	"body": {
		Type:             schema.TypeString,
		Required:         true,
		DiffSuppressFunc: diffSuppressDestination,
		ValidateFunc:     validation.StringIsJSON,
	},
}

func resourceElasticsearchDeprecatedDestination() *schema.Resource {
	return &schema.Resource{
		Create: resourceElasticsearchOpenDistroDestinationCreate,
		Read:   resourceElasticsearchOpenDistroDestinationRead,
		Update: resourceElasticsearchOpenDistroDestinationUpdate,
		Delete: resourceElasticsearchOpenDistroDestinationDelete,
		Schema: openDistroDestinationSchema,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		DeprecationMessage: "elasticsearch_destination is deprecated, please use elasticsearch_opendistro_destination resource instead.",
	}
}

func resourceElasticsearchOpenDistroDestination() *schema.Resource {
	return &schema.Resource{
		Create: resourceElasticsearchOpenDistroDestinationCreate,
		Read:   resourceElasticsearchOpenDistroDestinationRead,
		Update: resourceElasticsearchOpenDistroDestinationUpdate,
		Delete: resourceElasticsearchOpenDistroDestinationDelete,
		Schema: openDistroDestinationSchema,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
	}
}

func resourceElasticsearchOpenDistroDestinationCreate(d *schema.ResourceData, m interface{}) error {
	res, err := resourceElasticsearchOpenDistroPostDestination(d, m)

	if err != nil {
		log.Printf("[INFO] Failed to put destination: %+v", err)
		return err
	}

	d.SetId(res.ID)
	destination, err := json.Marshal(res.Destination)
	if err != nil {
		return err
	}
	err = d.Set("body", string(destination))
	return err
}

func resourceElasticsearchOpenDistroDestinationRead(d *schema.ResourceData, m interface{}) error {
	res, err := resourceElasticsearchOpenDistroGetDestination(d.Id(), m)

	if elastic6.IsNotFound(err) || elastic7.IsNotFound(err) {
		log.Printf("[WARN] Destination (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return err
	}

	err = d.Set("body", res)
	return err
}

func resourceElasticsearchOpenDistroDestinationUpdate(d *schema.ResourceData, m interface{}) error {
	_, err := resourceElasticsearchOpenDistroPutDestination(d, m)

	if err != nil {
		return err
	}

	return resourceElasticsearchOpenDistroDestinationRead(d, m)
}

func resourceElasticsearchOpenDistroDestinationDelete(d *schema.ResourceData, m interface{}) error {
	var err error

	path, err := uritemplates.Expand("/_opendistro/_alerting/destinations/{id}", map[string]string{
		"id": d.Id(),
	})
	if err != nil {
		return fmt.Errorf("error building URL path for destination: %+v", err)
	}

	switch client := m.(type) {
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
		err = errors.New("destination resource not implemented prior to Elastic v6")
	}

	return err
}

func resourceElasticsearchOpenDistroGetDestination(destinationID string, m interface{}) (string, error) {
	var err error
	response := new(destinationResponse)

	// See https://github.com/opendistro-for-elasticsearch/alerting/issues/56, no API endpoint for retrieving destination
	var body *json.RawMessage
	switch client := m.(type) {
	case *elastic7.Client:
		body, err = elastic7GetObject(client, DESTINATION_INDEX, destinationID)
	case *elastic6.Client:
		body, err = elastic6GetObject(client, DESTINATION_TYPE, DESTINATION_INDEX, destinationID)
	default:
		err = errors.New("destination resource not implemented prior to Elastic v6")
	}

	if err != nil {
		return "", err
	}

	if err := json.Unmarshal(*body, response); err != nil {
		return "", fmt.Errorf("error unmarshalling destination body: %+v: %+v", err, body)
	}

	tj, err := json.Marshal(response.Destination)
	if err != nil {
		return "", err
	}

	return string(tj), err
}

func resourceElasticsearchOpenDistroPostDestination(d *schema.ResourceData, m interface{}) (*destinationResponse, error) {
	destinationJSON := d.Get("body").(string)

	var err error
	response := new(destinationResponse)

	path := "/_opendistro/_alerting/destinations/"

	var body json.RawMessage
	switch client := m.(type) {
	case *elastic7.Client:
		var res *elastic7.Response
		res, err = client.PerformRequest(context.TODO(), elastic7.PerformRequestOptions{
			Method: "POST",
			Path:   path,
			Body:   destinationJSON,
		})
		body = res.Body
	case *elastic6.Client:
		var res *elastic6.Response
		res, err = client.PerformRequest(context.TODO(), elastic6.PerformRequestOptions{
			Method: "POST",
			Path:   path,
			Body:   destinationJSON,
		})
		body = res.Body
	default:
		err = errors.New("destination resource not implemented prior to Elastic v6")
	}

	if err != nil {
		return response, err
	}

	if err := json.Unmarshal(body, response); err != nil {
		return response, fmt.Errorf("error unmarshalling destination body: %+v: %+v", err, body)
	}

	return response, nil
}

func resourceElasticsearchOpenDistroPutDestination(d *schema.ResourceData, m interface{}) (*destinationResponse, error) {
	destinationJSON := d.Get("body").(string)

	var err error
	response := new(destinationResponse)

	path, err := uritemplates.Expand("/_opendistro/_alerting/destinations/{id}", map[string]string{
		"id": d.Id(),
	})
	if err != nil {
		return response, fmt.Errorf("error building URL path for destination: %+v", err)
	}

	var body json.RawMessage
	switch client := m.(type) {
	case *elastic7.Client:
		var res *elastic7.Response
		res, err = client.PerformRequest(context.TODO(), elastic7.PerformRequestOptions{
			Method: "PUT",
			Path:   path,
			Body:   destinationJSON,
		})
		body = res.Body
	case *elastic6.Client:
		var res *elastic6.Response
		res, err = client.PerformRequest(context.TODO(), elastic6.PerformRequestOptions{
			Method: "PUT",
			Path:   path,
			Body:   destinationJSON,
		})
		body = res.Body
	default:
		err = errors.New("destination resource not implemented prior to Elastic v6")
	}

	if err != nil {
		return response, err
	}

	if err := json.Unmarshal(body, response); err != nil {
		return response, fmt.Errorf("error unmarshalling destination body: %+v: %+v", err, body)
	}

	return response, nil
}

type destinationResponse struct {
	Version     int         `json:"_version"`
	ID          string      `json:"_id"`
	Destination interface{} `json:"destination"`
}
