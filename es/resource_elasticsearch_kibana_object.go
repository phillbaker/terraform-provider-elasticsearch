package es

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"

	elastic7 "github.com/olivere/elastic/v7"
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
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: func(i interface{}, k string) (warnings []string, errors []error) {
					v, ok := i.(string)
					if !ok {
						errors = append(errors, fmt.Errorf("expected type of %s to be string", k))
						return warnings, errors
					}

					if _, err := structure.NormalizeJsonString(v); err != nil {
						errors = append(errors, fmt.Errorf("%q contains an invalid JSON: %s", k, err))
						return warnings, errors
					}

					var body []interface{}
					if err := json.Unmarshal([]byte(v), &body); err != nil {
						errors = append(errors, fmt.Errorf("%q must be an array of objects: %s", k, err))
						return warnings, errors
					}

					for _, o := range body {
						kibanaObject, ok := o.(map[string]interface{})

						if !ok {
							errors = append(errors, fmt.Errorf("entries must be objects"))
							continue
						}

						for _, k := range requiredKibanaObjectKeys() {
							if kibanaObject[k] == nil {
								errors = append(errors, fmt.Errorf("object must have the %q key", k))
							}
						}
					}

					return warnings, errors
				},
				// DiffSuppressFunc: diffSuppressKibanaObject,
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
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

const deprecatedDocType = "doc"

func resourceElasticsearchKibanaObjectCreate(d *schema.ResourceData, meta interface{}) error {
	index := d.Get("index").(string)
	mapping_index := d.Get("index").(string)

	var success int
	var err error
	esClient, err := getClient(meta.(*ProviderConf))
	if err != nil {
		return err
	}
	switch client := esClient.(type) {
	case *elastic7.Client:
		success, err = elastic7CreateIndexIfNotExists(client, index, mapping_index)
	case *elastic6.Client:
		success, err = elastic6CreateIndexIfNotExists(client, index, mapping_index)
	default:
		return errors.New("Elasticsearch version not supported")
	}

	if err != nil {
		log.Printf("[INFO] Failed to create new kibana index: %+v", err)
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
		if err == nil && createIndex.Acknowledged {
			return INDEX_CREATED, err
		} else if e, ok := err.(*elastic7.Error); ok && e.Details.Type == "resource_already_exists_exception" {
			// If we have multiple parallel objects creating, we can get a race condition
			return INDEX_CREATED, nil
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
		if err == nil && createIndex.Acknowledged {
			return INDEX_CREATED, err
		} else if e, ok := err.(*elastic6.Error); ok && e.Details.Type == "resource_already_exists_exception" {
			// If we have multiple parallel objects creating, we can get a race condition
			return INDEX_CREATED, nil
		}
		return INDEX_CREATION_FAILED, err
	}

	return INDEX_EXISTS, nil
}

func resourceElasticsearchKibanaObjectRead(d *schema.ResourceData, meta interface{}) error {
	bodyString := d.Get("body").(string)
	var body []interface{}
	if err := json.Unmarshal([]byte(bodyString), &body); err != nil {
		log.Printf("[WARN] Failed to unmarshal on read: %+v", bodyString)
		return err
	}
	kibanaObject, ok := body[0].(map[string]interface{})
	if !ok {
		return fmt.Errorf("expected %v to be an object", body[0])
	}
	id := kibanaObject["_id"].(string)
	objectType := objectTypeOrDefault(kibanaObject)
	index := d.Get("index").(string)

	var resultJSON []byte
	var err error
	esClient, err := getClient(meta.(*ProviderConf))
	if err != nil {
		return err
	}
	switch client := esClient.(type) {
	case *elastic7.Client:
		var result *elastic7.GetResult
		result, err = elastic7GetObject(client, index, id)
		if err == nil {
			resultJSON, err = json.Marshal(result)
		}
	case *elastic6.Client:
		var result *elastic6.GetResult
		result, err = elastic6GetObject(client, objectType, index, id)
		if err == nil {
			resultJSON, err = json.Marshal(result)
		}
	default:
		return errors.New("Elasticsearch version not supported")
	}

	if err != nil {
		if elastic7.IsNotFound(err) || elastic6.IsNotFound(err) {
			log.Printf("[WARN] Kibana Object (%s) not found, removing from state", id)
			d.SetId("")
			return nil
		}

		return err
	}
	log.Printf("[TRACE] body: %s", string(resultJSON))

	ds := &resourceDataSetter{d: d}
	ds.set("index", index)

	// The Kibana object interface was originally built with the notion that
	// multiple kibana objects would be specified in the same resource, however,
	// that's not practical given that the Elasticsearch API is for a single
	// object. We account for that here: use the _source attribute and build a
	// single entry array
	var originalKeys []string
	for k := range kibanaObject {
		originalKeys = append(originalKeys, k)
	}

	result := make(map[string]interface{})
	if err := json.Unmarshal(resultJSON, &result); err != nil {
		log.Printf("[WARN] Failed to unmarshal: %+v", resultJSON)
		return err
	}

	stateObject := []map[string]interface{}{make(map[string]interface{})}
	for _, k := range originalKeys {
		stateObject[0][k] = result[k]
	}
	state, err := json.Marshal(stateObject)
	if err != nil {
		return fmt.Errorf("error marshalling resource data: %+v", err)
	}
	ds.set("body", string(state))

	return ds.err
}

func resourceElasticsearchKibanaObjectUpdate(d *schema.ResourceData, meta interface{}) error {
	_, err := resourceElasticsearchPutKibanaObject(d, meta)
	return err
}

func resourceElasticsearchKibanaObjectDelete(d *schema.ResourceData, meta interface{}) error {
	bodyString := d.Get("body").(string)
	var body []interface{}
	if err := json.Unmarshal([]byte(bodyString), &body); err != nil {
		log.Printf("[WARN] Failed to unmarshal: %+v", bodyString)
		return err
	}
	object, ok := body[0].(map[string]interface{})
	if !ok {
		return fmt.Errorf("expected %v to be an object", body[0])
	}
	id := object["_id"].(string)
	objectType := objectTypeOrDefault(object)
	index := d.Get("index").(string)

	var err error
	esClient, err := getClient(meta.(*ProviderConf))
	if err != nil {
		return err
	}
	switch client := esClient.(type) {
	case *elastic7.Client:
		err = elastic7DeleteIndex(client, index, id)
	case *elastic6.Client:
		err = elastic6DeleteIndex(client, objectType, index, id)
	default:
		return errors.New("Elasticsearch version not supported")
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

func resourceElasticsearchPutKibanaObject(d *schema.ResourceData, meta interface{}) (string, error) {
	bodyString := d.Get("body").(string)
	var body []interface{}
	if err := json.Unmarshal([]byte(bodyString), &body); err != nil {
		log.Printf("[WARN] Failed to unmarshal on put: %+v", bodyString)
		return "", err
	}
	object, ok := body[0].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("expected %v to be an object", body[0])
	}
	id := object["_id"].(string)
	objectType := objectTypeOrDefault(object)
	data := object["_source"]
	index := d.Get("index").(string)

	var err error
	esClient, err := getClient(meta.(*ProviderConf))
	if err != nil {
		return "", err
	}
	switch client := esClient.(type) {
	case *elastic7.Client:
		err = elastic7PutIndex(client, index, id, data)
	case *elastic6.Client:
		err = elastic6PutIndex(client, objectType, index, id, data)
	default:
		err = errors.New("Elasticsearch version not supported")
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

// objectType is deprecated
func objectTypeOrDefault(document map[string]interface{}) string {
	if document["_type"] != nil {
		return document["_type"].(string)
	}

	return deprecatedDocType
}

func requiredKibanaObjectKeys() []string {
	return []string{"_source", "_id"}
}
