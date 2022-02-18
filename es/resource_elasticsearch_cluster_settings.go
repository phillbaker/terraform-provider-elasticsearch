package es

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	elastic7 "github.com/olivere/elastic/v7"
	elastic6 "gopkg.in/olivere/elastic.v6"
)

func resourceElasticsearchClusterSettings() *schema.Resource {
	return &schema.Resource{
		Create: resourceElasticsearchClusterSettingsCreate,
		Read:   resourceElasticsearchClusterSettingsRead,
		Update: resourceElasticsearchClusterSettingsUpdate,
		Delete: resourceElasticsearchClusterSettingsDelete,
		Schema: map[string]*schema.Schema{
			"body": {
				Type:             schema.TypeString,
				Required:         true,
				DiffSuppressFunc: diffSuppressIndexTemplate,
				ValidateFunc:     validation.StringIsJSON,
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceElasticsearchClusterSettingsCreate(d *schema.ResourceData, meta interface{}) error {
	err := resourceElasticsearchPutClusterSettings(d, meta)
	if err != nil {
		return err
	}
	d.SetId("settings")
	return nil
}

func resourceElasticsearchPutClusterSettings(d *schema.ResourceData, meta interface{}) error {

	var statusCode *int

	var err error

	esClient, err := getClient(meta.(*ProviderConf))
	if err != nil {
		return err
	}
	body := d.Get("body").(string)

	switch client := esClient.(type) {
	case *elastic7.Client:
		var res *elastic7.Response
		res, err = client.PerformRequest(context.TODO(), elastic7.PerformRequestOptions{
			Method: "PUT",
			Path:   "/_cluster/settings",
			Body:   body,
		})
		if err != nil {
			return err
		}
		statusCode = &res.StatusCode
	case *elastic6.Client:
		var res *elastic6.Response
		res, err = client.PerformRequest(context.TODO(), elastic6.PerformRequestOptions{
			Method: "PUT",
			Path:   "/_cluster/settings",
			Body:   body,
		})
		if err != nil {
			return err
		}
		statusCode = &res.StatusCode
	default:
		return errors.New("elasticsearch version not supported")
	}

	if *statusCode != 200 {
		s := fmt.Sprintf("can't query elastic, status code: %d", *statusCode)
		err = errors.New(s)
	}
	return err
}

func resourceElasticsearchClusterSettingsRead(d *schema.ResourceData, meta interface{}) error {
	var statusCode *int
	var err error
	var responce *json.RawMessage
	esClient, err := getClient(meta.(*ProviderConf))
	if err != nil {
		return err
	}
	switch client := esClient.(type) {
	case *elastic7.Client:
		var res *elastic7.Response

		res, err = client.PerformRequest(context.TODO(), elastic7.PerformRequestOptions{
			Method: "GET",
			Path:   "/_cluster/settings",
		})
		if err != nil {
			return err
		}
		statusCode = &res.StatusCode
		responce = &res.Body
	case *elastic6.Client:
		var res *elastic6.Response

		res, err = client.PerformRequest(context.TODO(), elastic6.PerformRequestOptions{
			Method: "GET",
			Path:   "/_cluster/settings",
		})
		if err != nil {
			return err
		}
		statusCode = &res.StatusCode
		responce = &res.Body
	default:
		return errors.New("elasticsearch version not supported")
	}

	responce_j, err := json.Marshal(responce)
	if err != nil {
		return err
	}
	if *statusCode != 200 {
		s := fmt.Sprintf("can't query elastic, status code: %d", *statusCode)
		return errors.New(s)
	}
	ds := &resourceDataSetter{d: d}
	ds.set("body", string(responce_j))
	return ds.err
}

func resourceElasticsearchClusterSettingsUpdate(d *schema.ResourceData, meta interface{}) error {
	return resourceElasticsearchPutClusterSettings(d, meta)
}

func resourceElasticsearchClusterSettingsDelete(d *schema.ResourceData, meta interface{}) error {

	var statusCode *int

	var err error

	esClient, err := getClient(meta.(*ProviderConf))
	if err != nil {
		return err
	}
	body := `{
		"persistent" : {
			"cluster.*" : null,
			"indices.*" : null
		},
		"transient" : {
			"cluster.*" : null,
			"indices.*" : null
		}
	  }`

	switch client := esClient.(type) {
	case *elastic7.Client:
		var res *elastic7.Response
		res, err = client.PerformRequest(context.TODO(), elastic7.PerformRequestOptions{
			Method: "PUT",
			Path:   "/_cluster/settings",
			Body:   body,
		})
		if err != nil {
			return err
		}
		statusCode = &res.StatusCode
	case *elastic6.Client:
		var res *elastic6.Response
		res, err = client.PerformRequest(context.TODO(), elastic6.PerformRequestOptions{
			Method: "PUT",
			Path:   "/_cluster/settings",
			Body:   body,
		})
		if err != nil {
			return err
		}
		statusCode = &res.StatusCode
	default:
		return errors.New("elasticsearch version not supported")
	}

	if *statusCode != 200 {
		s := fmt.Sprintf("can't query elastic, status code: %d", *statusCode)
		err = errors.New(s)
	}
	d.SetId("")
	return err
}
