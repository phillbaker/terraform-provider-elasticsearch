package main

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	elastic7 "github.com/olivere/elastic/v7"
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
				ForceNew: true,
			},
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

func resourceElasticsearchWatchCreate(d *schema.ResourceData, m interface{}) error {
	// Determine whether the watch already exists.
	watchID := d.Get("watch_id").(string)
	_, err := resourceElasticsearchGetWatch(watchID, m)
	if !elastic6.IsNotFound(err) && !elastic7.IsNotFound(err) {
		log.Printf("[INFO] watch exists: %+v", err)
		return fmt.Errorf("watch already exists with ID: %v", watchID)
	}

	watchID, err = resourceElasticsearchPutWatch(d, m)

	if err != nil {
		log.Printf("[INFO] Failed to put watch: %+v", err)
		return err
	}

	d.SetId(watchID)
	log.Printf("[INFO] Object ID: %s", d.Id())

	return resourceElasticsearchWatchRead(d, m)
}

func resourceElasticsearchWatchRead(d *schema.ResourceData, m interface{}) error {
	res, err := resourceElasticsearchGetWatch(d.Id(), m)

	if elastic6.IsNotFound(err) || elastic7.IsNotFound(err) {
		log.Printf("[WARN] Watch (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return err
	}

	switch m.(type) {
	case *elastic7.Client:
		watchResponse := res.(*elastic7.XPackWatcherGetWatchResponse)
		d.Set("body", watchResponse.Watch)
	case *elastic6.Client:
		watchResponse := res.(*elastic6.XPackWatcherGetWatchResponse)
		d.Set("body", watchResponse.Watch)
	}

	// response := new(watcherGetWatchResponse)
	// if err := json.Unmarshal(res.Body, response); err != nil {
	// 	return fmt.Errorf("error unmarshalling watch body: %+v: %+v", err, res.Body)
	// }

	d.Set("watch_id", d.Id())

	return nil
}

func resourceElasticsearchWatchUpdate(d *schema.ResourceData, m interface{}) error {
	_, err := resourceElasticsearchPutWatch(d, m)

	if err != nil {
		return err
	}

	return resourceElasticsearchWatchRead(d, m)
}

func resourceElasticsearchWatchDelete(d *schema.ResourceData, m interface{}) error {
	var err error
	switch m.(type) {
	case *elastic7.Client:
		client := m.(*elastic7.Client)
		_, err = client.XPackWatchDelete(d.Id()).Do(context.TODO())
	case *elastic6.Client:
		client := m.(*elastic6.Client)
		_, err = client.XPackWatchDelete(d.Id()).Do(context.TODO())
	default:
		err = errors.New("watch resource not implemented prior to Elastic v6")
	}

	return err
}

func resourceElasticsearchGetWatch(watchID string, m interface{}) (interface{}, error) {
	var res interface{}
	var err error
	switch m.(type) {
	case *elastic7.Client:
		client := m.(*elastic7.Client)
		res, err = client.XPackWatchGet(watchID).Do(context.TODO())
	case *elastic6.Client:
		client := m.(*elastic6.Client)
		res, err = client.XPackWatchGet(watchID).Do(context.TODO())
	default:
		err = errors.New("watch resource not implemented prior to Elastic v6")
	}

	return res, err
}

func resourceElasticsearchPutWatch(d *schema.ResourceData, m interface{}) (string, error) {
	watchID := d.Get("watch_id").(string)
	watchJSON := d.Get("body").(string)

	var err error
	switch m.(type) {
	case *elastic7.Client:
		client := m.(*elastic7.Client)
		_, err = client.XPackWatchPut(watchID).
			Body(watchJSON).
			Do(context.TODO())
	case *elastic6.Client:
		client := m.(*elastic6.Client)
		_, err = client.XPackWatchPut(watchID).
			Body(watchJSON).
			Do(context.TODO())
	default:
		err = errors.New("watch resource not implemented prior to Elastic v6")
	}

	if err != nil {
		return "", err
	}

	return watchID, nil
}
