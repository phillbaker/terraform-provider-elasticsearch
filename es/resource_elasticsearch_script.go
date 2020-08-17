package es

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	elastic7 "github.com/olivere/elastic/v7"
	elastic6 "gopkg.in/olivere/elastic.v6"
)

func resourceElasticsearchScript() *schema.Resource {
	return &schema.Resource{
		Create: resourceElasticsearchScriptCreate,
		Read:   resourceElasticsearchScriptRead,
		Update: resourceElasticsearchScriptUpdate,
		Delete: resourceElasticsearchScriptDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "Name (id) of the stored script to create",
				ForceNew:    true,
				Required:    true,
			},
			"language": {
				Type:        schema.TypeString,
				Description: "Specifies the language the script is written in, defaults to painless by the API if not specified.",
				Default:     "",
				Optional:    true,
			},
			"source": {
				Type:        schema.TypeString,
				Description: "The source of the stored script",
				Default:     "",
				Required:    true,
			},
		},
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
	}
}

func resourceElasticsearchScriptCreate(d *schema.ResourceData, meta interface{}) error {
	return resourceElasticsearchScriptUpdate(d, meta)
}

func resourceElasticsearchScriptDelete(d *schema.ResourceData, meta interface{}) error {
	id := d.Get("name").(string)

	var err error

	esClient, err := getClient(meta.(*ProviderConf))
	if err != nil {
		return err
	}
	switch client := esClient.(type) {
	case *elastic7.Client:
		_, err = client.DeleteScript().
			Id(id).
			Do(context.Background())
	case *elastic6.Client:
		_, err = client.DeleteScript().
			Id(id).
			Do(context.Background())
	default:
		// Not supported by the upstream client and there are some oddities, see
		// https://github.com/olivere/elastic/issues/643
		err = errors.New("script resource not implemented prior to Elastic v6")
	}

	if err != nil {
		return err
	}

	d.SetId("")
	return nil
}

func resourceElasticsearchScriptUpdate(d *schema.ResourceData, meta interface{}) error {
	id := d.Get("name").(string)
	var err error

	body, err := buildScriptJSONBody(d)
	if err != nil {
		return err
	}

	esClient, err := getClient(meta.(*ProviderConf))
	if err != nil {
		return err
	}

	switch client := esClient.(type) {
	case *elastic7.Client:
		_, err = client.PutScript().Id(id).BodyString(body).Do(context.Background())
	case *elastic6.Client:
		_, err = client.PutScript().Id(id).BodyString(body).Do(context.Background())
	default:
		// Not supported by the upstream client and there are some oddities, see
		// https://github.com/olivere/elastic/issues/643
		err = errors.New("script resource not implemented prior to Elastic v6")
	}

	if err != nil {
		return err
	}

	return resourceElasticsearchScriptRead(d, meta)
}

func resourceElasticsearchScriptRead(d *schema.ResourceData, meta interface{}) error {
	id := d.Get("name").(string)

	script, err := resourceElasticsearchScriptGet(id, meta)
	if err != nil {
		d.SetId("")
		return err
	}
	d.SetId(id)

	ds := &resourceDataSetter{d: d}
	ds.set("language", script.Language)
	ds.set("source", script.Source)
	return ds.err
}

func resourceElasticsearchScriptGet(scriptID string, meta interface{}) (StoredScript, error) {
	var scriptJson json.RawMessage
	var err error

	esClient, err := getClient(meta.(*ProviderConf))
	if err != nil {
		return StoredScript{}, err
	}
	switch client := esClient.(type) {
	case *elastic7.Client:
		var res *elastic7.GetScriptResponse
		res, err = client.GetScript().Id(scriptID).Do(context.Background())
		scriptJson = res.Script
	case *elastic6.Client:
		var res *elastic6.GetScriptResponse
		res, err = client.GetScript().Id(scriptID).Do(context.Background())
		scriptJson = res.Script
	default:
		// Not supported by the upstream client and there are some oddities, see
		// https://github.com/olivere/elastic/issues/643
		err = errors.New("script resource not implemented prior to Elastic v6")
	}

	if err != nil {
		return StoredScript{}, err
	}

	var script StoredScript
	if err := json.Unmarshal(scriptJson, &script); err != nil {
		return StoredScript{}, fmt.Errorf("error unmarshalling destination body: %+v: %+v", err, scriptJson)
	}

	return script, nil
}

func buildScriptJSONBody(d *schema.ResourceData) (string, error) {
	var err error
	// esClient, err := getClient(meta.(*ProviderConf))
	// if err != nil {
	// 	return nil, err
	// }

	body := make(map[string]interface{})
	// switch client := esClient.(type) {
	// case *elastic7.Client:
	// 	s := elastic7.NewScriptStored(d.Get("source").(string))
	// 	s.Lang(d.Get("language").(string))
	// 	d, err := s.Source()
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	body["script"] = d
	// case *elastic6.Client:
	// 	s := elastic6.NewScriptStored(d.Get("source").(string))
	// 	s.Lang(d.Get("language").(string))
	// 	d, err := s.Source()
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	body["script"] = d
	// default:
	// 	s := elastic5.NewScriptStored(d.Get("source").(string))
	// 	s.Lang(d.Get("language").(string))
	// 	d, err := s.Source()
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	body["script"] = d
	// }
	script := StoredScript{
		Language: d.Get("language").(string),
		Source:   d.Get("source").(string),
	}
	body["script"] = script

	data, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

type StoredScript struct {
	Name     string `json:"name"`
	Language string `json:"lang"`
	Source   string `json:"source"`
}
