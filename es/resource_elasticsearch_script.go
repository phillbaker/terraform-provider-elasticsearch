package es

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	elastic7 "github.com/olivere/elastic/v7"
	elastic6 "gopkg.in/olivere/elastic.v6"
)

var scriptSchema = map[string]*schema.Schema{
	"script_id": {
		Type:        schema.TypeString,
		Description: "Identifier for the stored script. Must be unique within the cluster.",
		Required:    true,
		ForceNew:    true,
	},
	"source": {
		Type:        schema.TypeString,
		Description: "The source of the stored script",
		Required:    true,
	},
	"lang": {
		Type:        schema.TypeString,
		Description: "Specifies the language the script is written in. Defaults to painless.",
		Default:     "painless",
		Optional:    true,
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

func buildScriptJSONBody(d *schema.ResourceData) (string, error) {
	var err error

	body := make(map[string]interface{})
	script := ScriptBody{
		Language: d.Get("lang").(string),
		Source:   d.Get("source").(string),
	}
	body["script"] = script

	data, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	return string(data), nil
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
	scriptBody, err := resourceElasticsearchGetScript(d.Id(), m)

	if elastic6.IsNotFound(err) || elastic7.IsNotFound(err) {
		log.Printf("[WARN] Script (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return err
	}

	ds := &resourceDataSetter{d: d}
	ds.set("script_id", d.Id())
	ds.set("source", scriptBody.Source)
	ds.set("lang", scriptBody.Language)

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

func resourceElasticsearchGetScript(scriptID string, m interface{}) (ScriptBody, error) {
	var scriptBody json.RawMessage
	var err error
	esClient, err := getClient(m.(*ProviderConf))
	if err != nil {
		return ScriptBody{}, err
	}
	switch client := esClient.(type) {
	case *elastic7.Client:
		var res *elastic7.GetScriptResponse
		res, err = client.GetScript().Id(scriptID).Do(context.TODO())
		if err != nil {
			return ScriptBody{}, err
		}
		scriptBody = res.Script
	case *elastic6.Client:
		var res *elastic6.GetScriptResponse
		res, err = client.GetScript().Id(scriptID).Do(context.TODO())
		if err != nil {
			return ScriptBody{}, err
		}
		scriptBody = res.Script
	default:
		err = errors.New("script resource not implemented prior to Elastic v6")
	}

	var script ScriptBody

	if err := json.Unmarshal(scriptBody, &script); err != nil {
		return ScriptBody{}, fmt.Errorf("error unmarshalling destination body: %+v: %+v", err, scriptBody)
	}

	return script, err
}

func resourceElasticsearchPutScript(d *schema.ResourceData, m interface{}) (string, error) {
	var err error
	scriptID := d.Get("script_id").(string)
	scriptBody, err := buildScriptJSONBody(d)

	if err != nil {
		return "", err
	}

	esClient, err := getClient(m.(*ProviderConf))
	if err != nil {
		return "", err
	}
	switch client := esClient.(type) {
	case *elastic7.Client:
		_, err = client.PutScript().
			Id(scriptID).
			BodyJson(scriptBody).
			Do(context.TODO())
	case *elastic6.Client:
		_, err = client.PutScript().
			Id(scriptID).
			BodyJson(scriptBody).
			Do(context.TODO())
	default:
		err = errors.New("script resource not implemented prior to Elastic v6")
	}

	if err != nil {
		return "", err
	}

	return scriptID, nil
}

type ScriptBody struct {
	Language string `json:"lang"`
	Source   string `json:"source"`
}

type Script struct {
	Name   string     `json:"name"`
	Script ScriptBody `json:"script"`
}
