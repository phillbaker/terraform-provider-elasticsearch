package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	elastic7 "github.com/olivere/elastic/v7"
	elastic6 "gopkg.in/olivere/elastic.v6"
)

const DESTINATION_NAME_FIELD = "destination.name.keyword"

func dataSourceElasticsearchDestination() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceElasticsearchDestinationRead,

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"body": &schema.Schema{
				Type:     schema.TypeMap,
				Computed: true,
			},
		},
	}
}

func dataSourceElasticsearchDestinationRead(d *schema.ResourceData, m interface{}) error {
	destinationName := d.Get("name").(string)

	response := new(destinationResponse)

	// See https://github.com/opendistro-for-elasticsearch/alerting/issues/70, no tags or API endpoint for searching destination
	var id string
	var body *json.RawMessage
	var err error
	switch m.(type) {
	case *elastic7.Client:
		client := m.(*elastic7.Client)
		id, body, err = elastic7Search(client, DESTINATION_INDEX, destinationName)
	case *elastic6.Client:
		client := m.(*elastic6.Client)
		id, body, err = elastic6Search(client, DESTINATION_INDEX, destinationName)
	default:
		err = errors.New("destination resource not implemented prior to Elastic v6")
	}

	if err != nil {
		return err
	} else if id == "" {
		// short circuit
		return nil
	}

	if err := json.Unmarshal(*body, response); err != nil {
		return fmt.Errorf("error unmarshalling destination body: %+v: %+v", err, body)
	}

	d.SetId(id)

	// we get a non-uniform map[string]interface{} back for the body, terraform
	// only accepts a mapping of string to primitive values
	simplifiedBody := map[string]string{}
	for key, value := range response.Destination.(map[string]interface{}) {
		if stringified, ok := value.(string); ok {
			simplifiedBody[key] = stringified
		} else {
			log.Printf("[INFO] couldn't simplify: %+v", value)
		}
	}
	d.Set("body", simplifiedBody)

	return err
}

func elastic7Search(client *elastic7.Client, index string, name string) (string, *json.RawMessage, error) {
	termQuery := elastic7.NewTermQuery(DESTINATION_NAME_FIELD, name)
	result, err := client.Search().
		Index(index).
		Query(termQuery).
		Do(context.TODO())

	if err != nil {
		return "", nil, err
	}
	if result.TotalHits() == 1 {
		return result.Hits.Hits[0].Id, &result.Hits.Hits[0].Source, nil
	} else if result.TotalHits() < 1 {
		return "", nil, err
	} else {
		return "", nil, fmt.Errorf("1 result expected, found %d.", result.TotalHits())
	}
}

func elastic6Search(client *elastic6.Client, index string, name string) (string, *json.RawMessage, error) {
	termQuery := elastic6.NewTermQuery(DESTINATION_NAME_FIELD, name)
	result, err := client.Search().
		Index(index).
		Query(termQuery).
		Do(context.TODO())

	if err != nil {
		return "", nil, err
	}
	if result.TotalHits() == 1 {
		return result.Hits.Hits[0].Id, result.Hits.Hits[0].Source, nil
	} else if result.TotalHits() < 1 {
		return "", nil, err
	} else {
		return "", nil, fmt.Errorf("1 result expected, found %d.", result.TotalHits())
	}
}
