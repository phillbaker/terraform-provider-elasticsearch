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

const DESTINATION_TYPE = "_doc"
const DESTINATION_INDEX = ".opendistro-alerting-config"

var openDistroDestinationSchema = map[string]*schema.Schema{
	"body": {
		Type:             schema.TypeString,
		Required:         true,
		DiffSuppressFunc: diffSuppressDestination,
		ValidateFunc:     validation.StringIsJSON,
		StateFunc: func(v interface{}) string {
			json, _ := structure.NormalizeJsonString(v)
			return json
		},
		Description: "The JSON body of the destination.",
	},
}

func resourceOpenSearchDestination() *schema.Resource {
	return &schema.Resource{
		Description: "Provides an OpenSearch destination, a reusable communication channel for an action, such as email, Slack, or a webhook URL. Please refer to the OpenDistro [destination documentation](https://opendistro.github.io/for-elasticsearch-docs/docs/alerting/monitors/#create-destinations) for details.",
		Create:      resourceElasticsearchOpenDistroDestinationCreate,
		Read:        resourceElasticsearchOpenDistroDestinationRead,
		Update:      resourceElasticsearchOpenDistroDestinationUpdate,
		Delete:      resourceElasticsearchOpenDistroDestinationDelete,
		Schema:      openDistroDestinationSchema,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceElasticsearchOpenDistroDestination() *schema.Resource {
	return &schema.Resource{
		Description: "Provides an Elasticsearch OpenDistro destination, a reusable communication channel for an action, such as email, Slack, or a webhook URL. Please refer to the OpenDistro [destination documentation](https://opendistro.github.io/for-elasticsearch-docs/docs/alerting/monitors/#create-destinations) for details.",
		Create:      resourceElasticsearchOpenDistroDestinationCreate,
		Read:        resourceElasticsearchOpenDistroDestinationRead,
		Update:      resourceElasticsearchOpenDistroDestinationUpdate,
		Delete:      resourceElasticsearchOpenDistroDestinationDelete,
		Schema:      openDistroDestinationSchema,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		DeprecationMessage: "elasticsearch_opendistro_destination is deprecated, please use elasticsearch_opensearch_destination resource instead.",
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
	destination, err := resourceElasticsearchOpenDistroQueryOrGetDestination(d.Id(), m)

	if elastic6.IsNotFound(err) || elastic7.IsNotFound(err) {
		log.Printf("[WARN] Destination (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return err
	}

	body, err := json.Marshal(destination)
	if err != nil {
		return err
	}

	err = d.Set("body", string(body))
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
		err = errors.New("destination resource not implemented prior to Elastic v6")
	}

	return err
}

func resourceElasticsearchOpenDistroGetDestination(destinationID string, esClient interface{}) (Destination, error) {
	switch client := esClient.(type) {
	case *elastic7.Client:
		path, err := uritemplates.Expand("/_opendistro/_alerting/destinations/{id}", map[string]string{
			"id": destinationID,
		})
		if err != nil {
			return Destination{}, fmt.Errorf("error building URL path for destination: %+v", err)
		}

		httpResponse, err := client.PerformRequest(context.TODO(), elastic7.PerformRequestOptions{
			Method: "GET",
			Path:   path,
		})
		if err != nil {
			return Destination{}, err
		}

		var drg destinationResponseGet
		if err := json.Unmarshal(httpResponse.Body, &drg); err != nil {
			return Destination{}, fmt.Errorf("error unmarshalling destination body: %+v", err)
		}
		// The response structure from the API is the same for the index and get
		// endpoints :|, and different from the other endpoints. Normalize the
		// response here.
		if len(drg.Destinations) > 0 {
			return drg.Destinations[0], nil
		} else {
			return Destination{}, fmt.Errorf("endpoint returned empty set of destinations: %+v", drg)
		}
	default:
		return Destination{}, errors.New("destination get api not implemented prior to ODFE 1.11.0")
	}
}

func resourceElasticsearchOpenDistroQueryOrGetDestination(destinationID string, m interface{}) (Destination, error) {
	esClient, err := getClient(m.(*ProviderConf))
	if err != nil {
		return Destination{}, err
	}

	var dr destinationResponse
	switch client := esClient.(type) {
	case *elastic7.Client:
		// See https://github.com/opendistro-for-elasticsearch/alerting/issues/56,
		// no API endpoint for retrieving destination prior to ODFE 1.11.0. So do
		// a request, if it 404s, fall back to trying to query the index.
		destination, err := resourceElasticsearchOpenDistroGetDestination(destinationID, client)
		if err == nil {
			return destination, err
		} else {
			result, err := elastic7GetObject(client, DESTINATION_INDEX, destinationID)

			if err != nil {
				return Destination{}, err
			}
			if err := json.Unmarshal(result.Source, &dr); err != nil {
				return Destination{}, fmt.Errorf("error unmarshalling destination body: %+v: %+v", err, result.Source)
			}
			return dr.Destination, nil
		}
	case *elastic6.Client:
		result, err := elastic6GetObject(client, DESTINATION_TYPE, DESTINATION_INDEX, destinationID)
		if err != nil {
			return Destination{}, err
		}
		if err := json.Unmarshal(*result.Source, &dr); err != nil {
			return Destination{}, fmt.Errorf("error unmarshalling destination body: %+v: %+v", err, result.Source)
		}
		return dr.Destination, nil
	default:
		return Destination{}, errors.New("destination resource not implemented prior to Elastic v6")
	}
}

func resourceElasticsearchOpenDistroPostDestination(d *schema.ResourceData, m interface{}) (*destinationResponse, error) {
	destinationJSON := d.Get("body").(string)

	var err error
	response := new(destinationResponse)

	path := "/_opendistro/_alerting/destinations/"

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
			Body:   destinationJSON,
		})
		if err != nil {
			return response, err
		}
		body = res.Body
	case *elastic6.Client:
		var res *elastic6.Response
		res, err = client.PerformRequest(context.TODO(), elastic6.PerformRequestOptions{
			Method: "POST",
			Path:   path,
			Body:   destinationJSON,
		})
		if err != nil {
			return response, err
		}
		body = res.Body
	default:
		return response, errors.New("destination resource not implemented prior to Elastic v6")
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
			Body:   destinationJSON,
		})
		if err != nil {
			return response, err
		}
		body = res.Body
	case *elastic6.Client:
		var res *elastic6.Response
		res, err = client.PerformRequest(context.TODO(), elastic6.PerformRequestOptions{
			Method: "PUT",
			Path:   path,
			Body:   destinationJSON,
		})
		if err != nil {
			return response, err
		}
		body = res.Body
	default:
		return response, errors.New("destination resource not implemented prior to Elastic v6")
	}

	if err := json.Unmarshal(body, response); err != nil {
		return response, fmt.Errorf("error unmarshalling destination body: %+v: %+v", err, body)
	}

	return response, nil
}

type destinationResponse struct {
	Version     int         `json:"_version"`
	ID          string      `json:"_id"`
	Destination Destination `json:"destination"`
}

// When this api endpoint was introduced after the other endpoints, it has a
// different response structure
type destinationResponseGet struct {
	Total        int           `json:"totalDestinations"`
	Destinations []Destination `json:"destinations"`
}

type Destination struct {
	ID            string      `json:"id"`
	Type          string      `json:"type"`
	Name          string      `json:"name"`
	Slack         interface{} `json:"slack,omitempty"`
	CustomWebhook interface{} `json:"custom_webhook,omitempty"`
	Chime         interface{} `json:"chime,omitempty"`
	SNS           interface{} `json:"sns,omitempty"`
	Email         interface{} `json:"email,omitempty"`
}
