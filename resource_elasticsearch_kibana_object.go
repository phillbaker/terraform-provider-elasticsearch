package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
	elastic "gopkg.in/olivere/elastic.v5"
)

func resourceElasticsearchKibanaObject() *schema.Resource {
	return &schema.Resource{
		Create: resourceElasticsearchKibanaObjectCreate,
		Read:   resourceElasticsearchKibanaObjectRead,
		Update: resourceElasticsearchKibanaObjectUpdate,
		Delete: resourceElasticsearchKibanaObjectDelete,
		Schema: map[string]*schema.Schema{
			"body": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				// ForceNew: true,
				// DiffSuppressFunc: diffSuppressKibanaObject,
			},
			"index": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  ".kibana",
			},
		},
	}
}

func resourceElasticsearchKibanaObjectCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*elastic.Client)
	// Use the IndexExists service to check if a specified index exists.
	exists, err := client.IndexExists(d.Get("index").(string)).Do(context.TODO())
	if err != nil {
		return err
	}
	if !exists {
		log.Printf("[INFO] Creating new kibana index")
		mapping := `{
      "search": {
        "properties": {
          "hits": {
            "type": "integer",
          },
          "version": {
            "type": "integer",
          }
        }
      }
    }`
		createIndex, err := client.CreateIndex(d.Get("index").(string) + "/_mapping/search").Body(mapping).Do(context.TODO())
		if err != nil {
			log.Printf("[INFO] Failed to creating new kibana index: %+v", err)
			return err
		}
		if !createIndex.Acknowledged {
			return fmt.Errorf("fail to create the Elasticsearch index")
		}
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

func resourceElasticsearchKibanaObjectRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*elastic.Client)
	// res, err := client.IndexGetTemplate(d.Id()).Do(context.TODO())
	bodyString := d.Get("body").(string)
	var body []map[string]interface{}
	if err := json.Unmarshal([]byte(bodyString), &body); err != nil {
		return err
	}
	// TODO handle multiple objects in json
	id := body[0]["_id"].(string)
	objectType := body[0]["_type"].(string)

	// termQuery := elastic.Query(elastic.NewTermQuery("title", id))
	result, err := client.Get().Index(d.Get("index").(string)).Type(objectType).Id(id).Do(context.TODO())
	if err != nil {
		if elastic.IsNotFound(err) {
			log.Printf("[WARN] Kibana Object (%s) not found, removing from state", id)
			d.SetId("")
			return nil
		}

		return err
	}
	if result.Found {
		// search.Hits.Hits.Fields
		// search.Hits.Hits.Source
		d.Set("index", d.Get("index").(string))
		d.Set("body", result.Source)
		// already exists
	} else {
		return fmt.Errorf("Object not found.")
	}

	// t := res[d.Id()]
	// tj, err := json.Marshal(t)
	// if err != nil {
	//   return err
	// }
	// d.Set("name", d.Id())
	// d.Set("body", string(tj))
	return nil
}

func resourceElasticsearchKibanaObjectUpdate(d *schema.ResourceData, meta interface{}) error {
	_, err := resourceElasticsearchPutKibanaObject(d, meta)
	return err
}

func resourceElasticsearchKibanaObjectDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*elastic.Client)
	// _, err := client.IndexDeleteTemplate(d.Id()).Do(context.TODO())

	bodyString := d.Get("body").(string)
	var body []map[string]interface{}
	if err := json.Unmarshal([]byte(bodyString), &body); err != nil {
		return err
	}
	// TODO handle multiple objects in json
	id := body[0]["_id"].(string)
	objectType := body[0]["_type"].(string)

	res, err := client.Delete().
		Index(d.Get("index").(string)).
		Type(objectType).
		Id(id).
		Do(context.TODO())

	if err != nil {
		return err
	}
	if !res.Found {
		// fmt.Print("Document deleted from from index\n")
		return fmt.Errorf("failed to delete the object")
	}

	return nil
}

func resourceElasticsearchPutKibanaObject(d *schema.ResourceData, meta interface{}) (string, error) {
	client := meta.(*elastic.Client)
	// name := d.Get("name").(string)
	// body := d.Get("body").(string)
	// _, err := client.IndexPutTemplate(name).BodyString(body).Create(create).Do(context.TODO())

	bodyString := d.Get("body").(string)
	var body []map[string]interface{}
	if err := json.Unmarshal([]byte(bodyString), &body); err != nil {
		log.Printf("[INFO] Failed to unmarshal: %+v", err)
		return "", err
	}
	// TODO handle multiple objects in json
	id := body[0]["_id"].(string)
	objectType := body[0]["_type"].(string)
	data := body[0]["_source"]

	// will PUT with the given id
	_, err := client.Index().
		Index(d.Get("index").(string)).
		Type(objectType).
		Id(id).
		BodyJson(&data).
		Do(context.TODO())

	if err != nil {
		return "", err
	}

	return id, nil
}
