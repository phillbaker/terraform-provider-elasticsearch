package es

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/olivere/elastic/uritemplates"
	elastic7 "github.com/olivere/elastic/v7"
	elastic6 "gopkg.in/olivere/elastic.v6"
)

const DESTINATION_NAME_FIELD = "destination.name.keyword"

var datasourceOpenDistroDestinationSchema = map[string]*schema.Schema{
	"name": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "Name of the destrination to retrieve",
	},
	"body": {
		Type:        schema.TypeMap,
		Computed:    true,
		Description: "Map of the attributes of the destination",
	},
}

func dataSourceOpenSearchDestination() *schema.Resource {
	return &schema.Resource{
		Description: "`elasticsearch_opensearch_destination` can be used to retrieve the destination object by name.",
		Read:        dataSourceElasticsearchOpenDistroDestinationRead,
		Schema:      datasourceOpenDistroDestinationSchema,
	}
}

func dataSourceElasticsearchOpenDistroDestination() *schema.Resource {
	return &schema.Resource{
		Description:        "`elasticsearch_opendistro_destination` can be used to retrieve the destination object by name.",
		Read:               dataSourceElasticsearchOpenDistroDestinationRead,
		Schema:             datasourceOpenDistroDestinationSchema,
		DeprecationMessage: "elasticsearch_opendistro_destination is deprecated, please use elasticsearch_opensearch_destination data source instead.",
	}
}

func dataSourceElasticsearchOpenDistroDestinationRead(d *schema.ResourceData, m interface{}) error {
	destinationName := d.Get("name").(string)

	var id string
	var destination map[string]interface{}
	var err error
	esClient, err := getClient(m.(*ProviderConf))
	if err != nil {
		return err
	}
	switch client := esClient.(type) {
	case *elastic7.Client:
		// See https://github.com/opendistro-for-elasticsearch/alerting/issues/70,
		// no tags or API endpoint for searching destination. In ODFE >= 1.11.0,
		// the index has become a "system index", so it cannot be searched:
		// https://opendistro.github.io/for-elasticsearch-docs/docs/alerting/settings/#alerting-indices
		// instead we paginate through all destinations to find the first name match :|
		id, destination, err = destinationElasticsearch7GetAll(client, destinationName)
		if err != nil {
			id, destination, err = destinationElasticsearch7Search(client, DESTINATION_INDEX, destinationName)
		}
	case *elastic6.Client:
		id, destination, err = destinationElasticsearch6Search(client, DESTINATION_INDEX, destinationName)
	default:
		err = errors.New("destination resource not implemented prior to Elastic v6")
	}

	if err != nil {
		return err
	} else if id == "" {
		// short circuit
		return nil
	}

	d.SetId(id)

	// we get a non-uniform map[string]interface{} back for the body, terraform
	// only accepts a mapping of string to primitive values. We want to save
	// this as a map so that attributes are accessible
	simplifiedBody := map[string]string{}
	for key, value := range destination {
		if stringified, ok := value.(string); ok {
			simplifiedBody[key] = stringified
		} else {
			log.Printf("[INFO] couldn't simplify: %+v", value)
		}
	}
	err = d.Set("body", simplifiedBody)
	return err
}

func destinationElasticsearch7Search(client *elastic7.Client, index string, name string) (string, map[string]interface{}, error) {
	termQuery := elastic7.NewTermQuery(DESTINATION_NAME_FIELD, name)
	result, err := client.Search().
		Index(index).
		Query(termQuery).
		Do(context.TODO())

	destination := make(map[string]interface{})
	if err != nil {
		return "", destination, err
	}
	if result.TotalHits() == 1 {
		if err := json.Unmarshal(result.Hits.Hits[0].Source, &destination); err != nil {
			return "", destination, fmt.Errorf("error unmarshalling destination body: %+v", err)
		}

		return result.Hits.Hits[0].Id, destination["destination"].(map[string]interface{}), nil
	} else if result.TotalHits() < 1 {
		return "", destination, err
	} else {
		return "", destination, fmt.Errorf("1 result expected, found %d.", result.TotalHits())
	}
}

func destinationElasticsearch6Search(client *elastic6.Client, index string, name string) (string, map[string]interface{}, error) {
	termQuery := elastic6.NewTermQuery(DESTINATION_NAME_FIELD, name)
	result, err := client.Search().
		Index(index).
		Query(termQuery).
		Do(context.TODO())

	destination := make(map[string]interface{})
	if err != nil {
		return "", destination, err
	}
	if result.TotalHits() == 1 {
		if err := json.Unmarshal(*result.Hits.Hits[0].Source, &destination); err != nil {
			return "", destination, fmt.Errorf("error unmarshalling destination body: %+v", err)
		}

		return result.Hits.Hits[0].Id, destination["destination"].(map[string]interface{}), nil
	} else if result.TotalHits() < 1 {
		return "", destination, err
	} else {
		return "", destination, fmt.Errorf("1 result expected, found %d.", result.TotalHits())
	}
}

func destinationElasticsearch7GetAll(client *elastic7.Client, name string) (string, map[string]interface{}, error) {
	offset := 0
	pageSize := 1000
	destination := make(map[string]interface{})
	for {
		path, err := uritemplates.Expand("/_opendistro/_alerting/destinations?startIndex={startIndex}&size={size}", map[string]string{
			"startIndex": fmt.Sprint(offset),
			"size":       fmt.Sprint(pageSize),
		})
		if err != nil {
			return "", destination, fmt.Errorf("error building URL path for destination: %+v", err)
		}

		httpResponse, err := client.PerformRequest(context.TODO(), elastic7.PerformRequestOptions{
			Method: "GET",
			Path:   path,
		})
		if err != nil {
			return "", destination, err
		}

		var drg destinationResponseGet
		if err := json.Unmarshal(httpResponse.Body, &drg); err != nil {
			return "", destination, fmt.Errorf("error unmarshalling destination body: %+v", err)
		}

		for _, d := range drg.Destinations {
			if d.Name == name {
				j, err := json.Marshal(d)
				if err != nil {
					return "", destination, fmt.Errorf("error marshalling destination: %+v", err)
				}
				if err := json.Unmarshal(j, &destination); err != nil {
					return "", destination, fmt.Errorf("error unmarshalling destination body: %+v", err)
				}
				return d.ID, destination, nil
			}
		}

		if drg.Total > offset {
			offset += pageSize
		} else {
			break
		}
	}

	return "", destination, fmt.Errorf("destination not found")
}
