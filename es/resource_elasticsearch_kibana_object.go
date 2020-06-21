package es

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	elastic7 "github.com/olivere/elastic/v7"
	elastic5 "gopkg.in/olivere/elastic.v5"
	elastic6 "gopkg.in/olivere/elastic.v6"
)

func resourceElasticsearchKibanaObject() *schema.Resource {
	return &schema.Resource{
		Create: resourceElasticsearchKibanaObjectCreate,
		Read:   resourceElasticsearchKibanaObjectRead,
		Update: resourceElasticsearchKibanaObjectUpdate,
		Delete: resourceElasticsearchKibanaObjectDelete,
		Schema: map[string]*schema.Schema{
			"body": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsJSON,
			},
			"index": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  ".kibana",
			},
		},
	}
}

const (
	INDEX_CREATED int = iota
	INDEX_EXISTS
	INDEX_CREATION_FAILED
)

const deprecatedDocType = "_doc"

func resourceElasticsearchKibanaObjectCreate(d *schema.ResourceData, meta interface{}) error {
	index := d.Get("index").(string)
	mapping_index := d.Get("index").(string)

	var success int
	var err error
	switch client := meta.(type) {
	case *elastic7.Client:
		success, err = elastic7CreateIndexIfNotExists(client, index, mapping_index)
	case *elastic6.Client:
		success, err = elastic6CreateIndexIfNotExists(client, index, mapping_index)
	default:
		elastic5Client := meta.(*elastic5.Client)
		success, err = elastic5CreateIndexIfNotExists(elastic5Client, index, mapping_index)
	}

	if err != nil {
		log.Printf("[INFO] Failed to creating new kibana index: %+v", err)
		return err
	}

	if success == INDEX_CREATED {
		log.Printf("[INFO] Created new kibana index")
	} else if success == INDEX_CREATION_FAILED {
		return fmt.Errorf("fail to create the Elasticsearch index")
	}

	id, err := resourceElasticsearchPutKibanaObject(d, meta)

	if err != nil {
		log.Printf("[INFO] Failed to put kibana object: %+v", err)
		return err
	}

	d.SetId(id)
	log.Printf("[INFO] Object ID: %s", d.Id())

	return nil
}

func elastic7CreateIndexIfNotExists(client *elastic7.Client, index string, mappingIndex string) (int, error) {
	log.Printf("[INFO] elastic7CreateIndexIfNotExists %s", index)

	// Use the IndexExists service to check if a specified index exists.
	exists, err := client.IndexExists(index).Do(context.TODO())
	if err != nil {
		return INDEX_CREATION_FAILED, err
	}
	if !exists {
		createIndex, err := client.CreateIndex(mappingIndex).Body(`{"mappings":{}}`).Do(context.TODO())
		if createIndex.Acknowledged {
			return INDEX_CREATED, err
		}
		return INDEX_CREATION_FAILED, err
	}

	return INDEX_EXISTS, nil
}

func elastic6CreateIndexIfNotExists(client *elastic6.Client, index string, mapping_index string) (int, error) {
	log.Printf("[INFO] elastic6CreateIndexIfNotExists")

	// Use the IndexExists service to check if a specified index exists.
	exists, err := client.IndexExists(index).Do(context.TODO())
	if err != nil {
		return INDEX_CREATION_FAILED, err
	}
	if !exists {
		createIndex, err := client.CreateIndex(mapping_index).Body(`{"mappings":{}}`).Do(context.TODO())
		if createIndex.Acknowledged {
			return INDEX_CREATED, err
		} else {
			return INDEX_CREATION_FAILED, err
		}
	}

	return INDEX_EXISTS, nil
}

func elastic5CreateIndexIfNotExists(client *elastic5.Client, index string, mapping_index string) (int, error) {
	mapping := `{
    "mappings": {
      "search": {
        "properties": {
          "hits": {
            "type": "integer"
          },
          "version": {
            "type": "integer"
          }
        }
      }
    }
  }`

	// Use the IndexExists service to check if a specified index exists.
	exists, err := client.IndexExists(index).Do(context.TODO())
	if err != nil {
		return INDEX_CREATION_FAILED, err
	}
	if !exists {
		createIndex, err := client.CreateIndex(mapping_index).Body(mapping).Do(context.TODO())
		if createIndex.Acknowledged {
			return INDEX_CREATED, err
		} else {
			return INDEX_CREATION_FAILED, err
		}
	}

	return INDEX_EXISTS, nil
}

func resourceElasticsearchKibanaObjectRead(d *schema.ResourceData, meta interface{}) error {
	bodyString := d.Get("body").(string)
	var body []map[string]interface{}
	if err := json.Unmarshal([]byte(bodyString), &body); err != nil {
		log.Printf("[WARN] Failed to unmarshal: %+v", bodyString)
		return err
	}
	// TODO handle multiple objects in json
	id := body[0]["_id"].(string)
	objectType := objectTypeOrDefault(body[0])
	index := d.Get("index").(string)

	var result *json.RawMessage
	var err error
	switch client := meta.(type) {
	case *elastic7.Client:
		result, err = elastic7GetObject(client, index, id)
	case *elastic6.Client:
		result, err = elastic6GetObject(client, objectType, index, id)
	default:
		elastic5Client := meta.(*elastic5.Client)
		result, err = elastic5GetObject(elastic5Client, objectType, index, id)
	}

	if err != nil {
		if elastic7.IsNotFound(err) || elastic6.IsNotFound(err) || elastic5.IsNotFound(err) {
			log.Printf("[WARN] Kibana Object (%s) not found, removing from state", id)
			d.SetId("")
			return nil
		}

		return err
	}

	ds := &resourceDataSetter{d: d}
	ds.set("index", index)
	d.Set("body", result)

	return ds.err
}

func resourceElasticsearchKibanaObjectUpdate(d *schema.ResourceData, meta interface{}) error {
	_, err := resourceElasticsearchPutKibanaObject(d, meta)
	return err
}

func resourceElasticsearchKibanaObjectDelete(d *schema.ResourceData, meta interface{}) error {
	bodyString := d.Get("body").(string)
	var body []map[string]interface{}
	if err := json.Unmarshal([]byte(bodyString), &body); err != nil {
		log.Printf("[WARN] Failed to unmarshal: %+v", bodyString)
		return err
	}
	// TODO handle multiple objects in json
	id := body[0]["_id"].(string)
	objectType := objectTypeOrDefault(body[0])
	index := d.Get("index").(string)

	var err error
	switch client := meta.(type) {
	case *elastic7.Client:
		err = elastic7DeleteIndex(client, index, id)
	case *elastic6.Client:
		err = elastic6DeleteIndex(client, objectType, index, id)
	default:
		elastic5Client := meta.(*elastic5.Client)
		err = elastic5DeleteIndex(elastic5Client, objectType, index, id)
	}

	if err != nil {
		return err
	}

	return nil
}

func elastic7DeleteIndex(client *elastic7.Client, index string, id string) error {
	_, err := client.Delete().
		Index(index).
		Id(id).
		Do(context.TODO())

	// we'll get an error if it's not found
	return err
}

func elastic6DeleteIndex(client *elastic6.Client, objectType string, index string, id string) error {
	_, err := client.Delete().
		Index(index).
		Type(objectType).
		Id(id).
		Do(context.TODO())

	// we'll get an error if it's not found: https://github.com/olivere/elastic/blob/v6.1.26/delete.go#L207-L210
	return err
}

func elastic5DeleteIndex(client *elastic5.Client, objectType string, index string, id string) error {
	_, err := client.Delete().
		Index(index).
		Type(objectType).
		Id(id).
		Do(context.TODO())

	// we'll get an error if it's not found: https://github.com/olivere/elastic/blob/v5.0.70/delete.go#L201-L203
	return err
}

func resourceElasticsearchPutKibanaObject(d *schema.ResourceData, meta interface{}) (string, error) {
	bodyString := d.Get("body").(string)
	var body []map[string]interface{}
	if err := json.Unmarshal([]byte(bodyString), &body); err != nil {
		log.Printf("[WARN] Failed to unmarshal: %+v", bodyString)
		return "", err
	}
	// TODO handle multiple objects in json
	id := body[0]["_id"].(string)
	objectType := objectTypeOrDefault(body[0])
	data := body[0]["_source"]
	index := d.Get("index").(string)

	var err error
	switch client := meta.(type) {
	case *elastic7.Client:
		err = elastic7PutIndex(client, index, id, data)
	case *elastic6.Client:
		err = elastic6PutIndex(client, objectType, index, id, data)
	default:
		elastic5Client := meta.(*elastic5.Client)
		err = elastic5PutIndex(elastic5Client, objectType, index, id, data)
	}

	if err != nil {
		return "", err
	}

	return id, nil
}

func elastic7PutIndex(client *elastic7.Client, index string, id string, data interface{}) error {
	_, err := client.Index().
		Index(index).
		Id(id).
		BodyJson(&data).
		Do(context.TODO())

	return err
}

func elastic6PutIndex(client *elastic6.Client, objectType string, index string, id string, data interface{}) error {
	_, err := client.Index().
		Index(index).
		Type(objectType).
		Id(id).
		BodyJson(&data).
		Do(context.TODO())

	return err
}

func elastic5PutIndex(client *elastic5.Client, objectType string, index string, id string, data interface{}) error {
	_, err := client.Index().
		Index(index).
		Type(objectType).
		Id(id).
		BodyJson(&data).
		Do(context.TODO())

	return err
}

// objectType is deprecated
func objectTypeOrDefault(document map[string]interface{}) string {
	if document["_type"] != nil {
		return document["_type"].(string)
	}

	return deprecatedDocType
}
