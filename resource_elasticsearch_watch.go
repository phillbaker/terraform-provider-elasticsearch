package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/olivere/elastic/uritemplates"
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
			"watch_json": &schema.Schema{
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
	if !elastic6.IsNotFound(err) {
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

	if elastic6.IsNotFound(err) {
		log.Printf("[WARN] Watch (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return err
	}

	response := new(watcherGetWatchResponse)
	if err := json.Unmarshal(res.Body, response); err != nil {
		return fmt.Errorf("error unmarshalling watch body: %+v: %+v", err, res.Body)
	}

	d.Set("watch_id", d.Id())
	d.Set("watch_json", response.Watch)

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
	case *elastic6.Client:
		client := m.(*elastic6.Client)
		_, err = client.XPackWatchDelete().
			Id(d.Id()).
			Do(context.TODO())
	default:
		err = errors.New("watch resource not implemented prior to Elastic v6")
	}

	return err
}

func resourceElasticsearchGetWatch(watchID string, m interface{}) (*elastic6.Response, error) {
	// Build URL for the watch
	path, err := uritemplates.Expand("/_xpack/watcher/watch/{id}", map[string]string{
		"id": watchID,
	})

	if err != nil {
		return nil, fmt.Errorf("error building URL path for watch: %+v", err)
	}

	var res *elastic6.Response
	switch m.(type) {
	case *elastic6.Client:
		client := m.(*elastic6.Client)
		res, err = client.PerformRequest(context.TODO(), elastic6.PerformRequestOptions{
			Method: "GET",
			Path:   path,
		})
	default:
		err = errors.New("watch resource not implemented prior to Elastic v6")
	}

	return res, err
}

func resourceElasticsearchPutWatch(d *schema.ResourceData, m interface{}) (string, error) {
	watchID := d.Get("watch_id").(string)
	watchJSON := d.Get("watch_json").(string)

	var err error
	switch m.(type) {
	case *elastic6.Client:
		client := m.(*elastic6.Client)
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

type watcherGetWatchResponse struct {
	Found  bool        `json:"found"`
	ID     string      `json:"_id"`
	Status watchStatus `json:"status"`
	Watch  interface{} `json:"watch"`
}

type watchStatus struct {
	State   map[string]interface{}            `json:"state"`
	Actions map[string]map[string]interface{} `json:"actions"`
	Version int                               `json:"version"`
}
