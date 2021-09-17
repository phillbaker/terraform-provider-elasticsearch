package es

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	elastic7 "github.com/olivere/elastic/v7"
)

var componentTemplateMinimalVersion, _ = version.NewVersion("7.8.0")

func resourceElasticsearchComponentTemplate() *schema.Resource {
	return &schema.Resource{
		Create: resourceElasticsearchComponentTemplateCreate,
		Read:   resourceElasticsearchComponentTemplateRead,
		Update: resourceElasticsearchComponentTemplateUpdate,
		Delete: resourceElasticsearchComponentTemplateDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				ForceNew:    true,
				Required:    true,
				Description: "Name of the component template to create.",
			},
			"body": {
				Type:             schema.TypeString,
				Required:         true,
				DiffSuppressFunc: diffSuppressComponentTemplate,
				ValidateFunc:     validation.StringIsJSON,
				Description:      "The JSON body of the template.",
			},
			"body_json": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "Component templates are building blocks for constructing index templates that specify index mappings, settings, and aliases. You cannot directly apply a component template to a data stream or index. To be applied, a component template must be included in an index template’s `composed_of` list.",
	}
}

func resourceElasticsearchComponentTemplateCreate(d *schema.ResourceData, meta interface{}) error {
	err := resourceElasticsearchPutComponentTemplate(d, meta, true)
	if err != nil {
		return err
	}
	d.SetId(d.Get("name").(string))
	return nil
}

func resourceElasticsearchComponentTemplateRead(d *schema.ResourceData, meta interface{}) error {
	id := d.Id()

	var result string
	var elasticVersion *version.Version

	esClient, err := getClient(meta.(*ProviderConf))
	if err != nil {
		return err
	}

	switch client := esClient.(type) {
	case *elastic7.Client:
		elasticVersion, err = elastic7GetVersion(client)
		if err == nil {
			if elasticVersion.LessThan(componentTemplateMinimalVersion) {
				err = fmt.Errorf("component_template endpoint only available from ElasticSearch >= 7.8, got version %s", elasticVersion.String())
			} else {
				result, err = elastic7GetComponentTemplate(client, id)
			}
		}
	default:
		err = fmt.Errorf("component_template endpoint only available from ElasticSearch >= 7.8, got version < 7.0.0")
	}
	if err != nil {
		if elastic7.IsNotFound(err) {
			log.Printf("[WARN] Index template (%s) not found, removing from state", id)
			d.SetId("")
			return nil
		}

		return err
	}

	ds := &resourceDataSetter{d: d}
	ds.set("name", d.Id())
	ds.set("body_json", result)
	ds.set("body", d.Get("body"))
	return ds.err
}

func elastic7GetComponentTemplate(client *elastic7.Client, id string) (string, error) {
	res, err := client.IndexGetComponentTemplate(id).Do(context.TODO())
	if err != nil {
		return "", err
	}

	// No more than 1 element is expected, if the index template is not found, previous call should
	// return a 404 error
	t := res.ComponentTemplates[0].ComponentTemplate
	tj, err := json.Marshal(t)
	if err != nil {
		return "", err
	}
	return string(tj), nil
}

func resourceElasticsearchComponentTemplateUpdate(d *schema.ResourceData, meta interface{}) error {
	return resourceElasticsearchPutComponentTemplate(d, meta, false)
}

func resourceElasticsearchComponentTemplateDelete(d *schema.ResourceData, meta interface{}) error {
	id := d.Id()

	var elasticVersion *version.Version

	esClient, err := getClient(meta.(*ProviderConf))
	if err != nil {
		return err
	}

	switch client := esClient.(type) {
	case *elastic7.Client:
		elasticVersion, err = elastic7GetVersion(client)
		if err == nil {
			if elasticVersion.LessThan(componentTemplateMinimalVersion) {
				err = fmt.Errorf("component_template endpoint only available from ElasticSearch >= 7.8, got version %s", elasticVersion.String())
			} else {
				err = elastic7DeleteComponentTemplate(client, id)
			}
		}
	default:
		err = fmt.Errorf("component_template endpoint only available from ElasticSearch >= 7.8, got version < 7.0.0")
	}

	if err != nil {
		return err
	}
	d.SetId("")
	return nil
}

func elastic7DeleteComponentTemplate(client *elastic7.Client, id string) error {
	_, err := client.IndexDeleteComponentTemplate(id).Do(context.TODO())
	return err
}

func resourceElasticsearchPutComponentTemplate(d *schema.ResourceData, meta interface{}, create bool) error {
	name := d.Get("name").(string)
	body := d.Get("body").(string)

	var elasticVersion *version.Version

	esClient, err := getClient(meta.(*ProviderConf))
	if err != nil {
		return err
	}

	switch client := esClient.(type) {
	case *elastic7.Client:
		elasticVersion, err = elastic7GetVersion(client)
		if err == nil {
			if elasticVersion.LessThan(componentTemplateMinimalVersion) {
				err = fmt.Errorf("component_template endpoint only available from ElasticSearch >= 7.8, got version %s", elasticVersion.String())
			} else {
				err = elastic7PutComponentTemplate(client, name, body, create)
			}
		}
	default:
		err = fmt.Errorf("component_template endpoint only available from ElasticSearch >= 7.8, got version < 7.0.0")
	}

	return err
}

func elastic7PutComponentTemplate(client *elastic7.Client, name string, body string, create bool) error {
	_, err := client.IndexPutComponentTemplate(name).BodyString(body).Create(create).Do(context.TODO())
	return err
}
