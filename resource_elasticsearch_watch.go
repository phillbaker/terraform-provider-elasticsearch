package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
	elastic6 "gopkg.in/olivere/elastic.v6"
)

func resourceElasticsearchWatch() *schema.Resource {
	return &schema.Resource{
		Create: resourceElasticsearchWatchCreate,
		Read:   resourceElasticsearchWatchRead,
		Update: resourceElasticsearchWatchUpdate,
		Delete: resourceElasticsearchWatchDelete,
		Schema: map[string]*schema.Schema{
			"watch_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"watch_json": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceElasticsearchWatchCreate(d *schema.ResourceData, meta interface{}) error {
	watchID, err := resourceElasticsearchPutWatch(d, meta)

	if err != nil {
		log.Printf("[INFO] Failed to put watch: %+v", err)
		return err
	}

	d.SetId(watchID)
	log.Printf("[INFO] Object ID: %s", d.Id())

	return nil
}

func resourceElasticsearchWatchRead(d *schema.ResourceData, meta interface{}) error {
	watchID := d.Get("watch_id").(string)

	var err error
	var response *elastic6.XpackWatcherGetWatchResponse
	switch meta.(type) {
	case *elastic6.Client:
		client := meta.(*elastic6.Client)
		response, err = client.XPackWatchGet().
			Id(watchID).
			Do(context.TODO())
	default:
		err = errors.New("watch resource not implemented prior to Elastic v6")
	}

	if err != nil {
		return err
	}

	if !response.Found {
		log.Printf("[WARN] Watch (%s) not found, removing from state", watchID)
		d.SetId("")
		return nil
	}

	watchBytes, err := json.Marshal(response.Watch)
	if err != nil {
		return fmt.Errorf("unable to marshall watch to JSON: %v: %+v", err, response.Watch)
	}

	watchJSON := string(watchBytes)
	d.Set("watch_json", watchJSON)

	return nil
}

func resourceElasticsearchWatchUpdate(d *schema.ResourceData, meta interface{}) error {
	_, err := resourceElasticsearchPutWatch(d, meta)
	return err
}

func resourceElasticsearchWatchDelete(d *schema.ResourceData, meta interface{}) error {
	watchID := d.Get("watch_id").(string)

	var err error
	switch meta.(type) {
	case *elastic6.Client:
		client := meta.(*elastic6.Client)
		_, err = client.XPackWatchDelete().
			Id(watchID).
			Do(context.TODO())
	default:
		err = errors.New("watch resource not implemented prior to Elastic v6")
	}

	return err
}

func resourceElasticsearchPutWatch(d *schema.ResourceData, meta interface{}) (string, error) {
	watchID := d.Get("watch_id").(string)
	watchJSON := d.Get("watch_json").(string)

	var err error
	switch meta.(type) {
	case *elastic6.Client:
		client := meta.(*elastic6.Client)
		_, err = client.XPackWatchPut().
			Id(watchID).
			BodyString(watchJSON).
			Do(context.TODO())
	default:
		err = errors.New("watch resource not implemented prior to Elastic v6")
	}

	if err != nil {
		return "", err
	}

	return watchID, nil
}
