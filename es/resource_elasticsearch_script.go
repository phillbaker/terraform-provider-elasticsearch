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

	elastic7 "github.com/olivere/elastic/v7"
	elastic6 "gopkg.in/olivere/elastic.v6"
)

var scriptSchema = map[string]*schema.Schema{
	"script_id": {
		Type:     schema.TypeString,
		Required: true,
		ForceNew: true,
	},
	"body": {
		Type:             schema.TypeString,
		Required:         true,
		ValidateFunc:     validation.StringIsJSON,
		DiffSuppressFunc: suppressEquivalentJson,
		StateFunc: func(v interface{}) string {
			json, _ := structure.NormalizeJsonString(v)
			return json
		},
	},
}

func resourceElasticsearchScript() *schema.Resource {
	return &schema.Resource{
		Create: resourceElasticsearchScriptCreate,
		Read:   resourceElasticsearchScriptRead,
		Update: resourceElasticsearchScriptUpdate,
		Delete: resourceElasticsearchScriptDelete,
		Schema: scriptSchema,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceElasticsearchScriptCreate(d *schema.ResourceData, m interface{}) error {
	// Determine whether the script already exists, otherwise the API will
	// override an existing script with the name.
	scriptID := d.Get("script_id").(string)
	_, err := resourceElasticsearchGetScript(scriptID, m)

	if err == nil {
		log.Printf("[INFO] script exists: %+v", err)
		return fmt.Errorf("script already exists with ID: %v", scriptID)
	} else if err != nil && !elastic6.IsNotFound(err) && !elastic7.IsNotFound(err) {
		return err
	}

	scriptID, err = resourceElasticsearchPutScript(d, m)

	if err != nil {
		log.Printf("[INFO] Failed to put script: %+v", err)
		return err
	}

	d.SetId(scriptID)
	log.Printf("[INFO] Object ID: %s", d.Id())

	return resourceElasticsearchScriptRead(d, m)
}

func resourceElasticsearchScriptRead(d *schema.ResourceData, m interface{}) error {
	res, err := resourceElasticsearchGetScript(d.Id(), m)

	if elastic6.IsNotFound(err) || elastic7.IsNotFound(err) {
		log.Printf("[WARN] Script (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return err
	}

	var script []byte

	esClient, err := getClient(m.(*ProviderConf))
	if err != nil {
		return err
	}
	switch esClient.(type) {
	case *elastic7.Client:
		scriptResponse := res.(*elastic7.GetScriptResponse)
		script, err = json.Marshal(scriptResponse.Script)
	case *elastic6.Client:
		scriptResponse := res.(*elastic6.GetScriptResponse)
		script, err = json.Marshal(scriptResponse.Script)
	}

	if err != nil {
		return err
	}

	ds := &resourceDataSetter{d: d}
	ds.set("body", string(script))
	ds.set("script_id", d.Id())

	return ds.err
}

func resourceElasticsearchScriptUpdate(d *schema.ResourceData, m interface{}) error {
	_, err := resourceElasticsearchPutScript(d, m)

	if err != nil {
		return err
	}

	return resourceElasticsearchScriptRead(d, m)
}

func resourceElasticsearchScriptDelete(d *schema.ResourceData, m interface{}) error {
	var err error
	esClient, err := getClient(m.(*ProviderConf))
	if err != nil {
		return err
	}
	switch client := esClient.(type) {
	case *elastic7.Client:
		_, err = client.DeleteScript().Id(d.Id()).Do(context.TODO())
	case *elastic6.Client:
		_, err = client.DeleteScript().Id(d.Id()).Do(context.TODO())
	default:
		err = errors.New("script resource not implemented prior to Elastic v6")
	}

	return err
}

func resourceElasticsearchGetScript(scriptID string, m interface{}) (interface{}, error) {
	var res interface{}
	var err error
	esClient, err := getClient(m.(*ProviderConf))
	if err != nil {
		return "", err
	}
	switch client := esClient.(type) {
	case *elastic7.Client:
		res, err = client.GetScript().Id(scriptID).Do(context.TODO())
	case *elastic6.Client:
		res, err = client.GetScript().Id(scriptID).Do(context.TODO())
	default:
		err = errors.New("script resource not implemented prior to Elastic v6")
	}

	return res, err
}

func resourceElasticsearchPutScript(d *schema.ResourceData, m interface{}) (string, error) {
	scriptID := d.Get("script_id").(string)
	scriptJSON := d.Get("body").(string)

	var err error
	esClient, err := getClient(m.(*ProviderConf))
	if err != nil {
		return "", err
	}
	switch client := esClient.(type) {
	case *elastic7.Client:
		_, err = client.PutScript().
			Id(scriptID).
			BodyJson(scriptJSON).
			Do(context.TODO())
	case *elastic6.Client:
		_, err = client.PutScript().
			Id(scriptID).
			BodyJson(scriptJSON).
			Do(context.TODO())
	default:
		err = errors.New("script resource not implemented prior to Elastic v6")
	}

	if err != nil {
		return "", err
	}

	return scriptID, nil
}
