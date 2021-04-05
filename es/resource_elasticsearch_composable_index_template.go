package es

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	elastic7 "github.com/olivere/elastic/v7"
)

var minimalESComposableTemplateVersion, _ = version.NewVersion("7.8.0")

func resourceElasticsearchComposableIndexTemplate() *schema.Resource {
	return &schema.Resource{
		Create: resourceElasticsearchComposableIndexTemplateCreate,
		Read:   resourceElasticsearchComposableIndexTemplateRead,
		Update: resourceElasticsearchComposableIndexTemplateUpdate,
		Delete: resourceElasticsearchComposableIndexTemplateDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"body": {
				Type:             schema.TypeString,
				Required:         true,
				DiffSuppressFunc: diffSuppressComposableIndexTemplate,
				ValidateFunc:     validation.StringIsJSON,
			},
		},
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
	}
}

func resourceElasticsearchComposableIndexTemplateCreate(d *schema.ResourceData, meta interface{}) error {
	err := resourceElasticsearchPutComposableIndexTemplate(d, meta, true)
	if err != nil {
		return err
	}
	d.SetId(d.Get("name").(string))
	return nil
}

func resourceElasticsearchComposableIndexTemplateRead(d *schema.ResourceData, meta interface{}) error {
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
			if elasticVersion.LessThan(minimalESComposableTemplateVersion) {
				err = fmt.Errorf("index_template endpoint only available from ElasticSearch >= 7.8, got version %s", elasticVersion.String())
			} else {
				result, err = elastic7GetIndexTemplate(client, id)
			}
		}
	default:
		err = fmt.Errorf("index_template endpoint only available from ElasticSearch >= 7.8, got version < 7.0.0")
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
	ds.set("body", result)
	return ds.err
}

func elastic7GetIndexTemplate(client *elastic7.Client, id string) (string, error) {
	res, err := client.IndexGetIndexTemplate(id).Do(context.TODO())
	if err != nil {
		return "", err
	}

	// No more than 1 element is expected, if the index template is not found, previous call should
	// return a 404 error
	t := res.IndexTemplates[0].IndexTemplate
	tj, err := json.Marshal(t)
	if err != nil {
		return "", err
	}
	return string(tj), nil
}

func resourceElasticsearchComposableIndexTemplateUpdate(d *schema.ResourceData, meta interface{}) error {
	return resourceElasticsearchPutComposableIndexTemplate(d, meta, false)
}

func resourceElasticsearchComposableIndexTemplateDelete(d *schema.ResourceData, meta interface{}) error {
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
			if elasticVersion.LessThan(minimalESComposableTemplateVersion) {
				err = fmt.Errorf("index_template endpoint only available from ElasticSearch >= 7.8, got version %s", elasticVersion.String())
			} else {
				err = elastic7DeleteIndexTemplate(client, id)
			}
		}
	default:
		err = fmt.Errorf("index_template endpoint only available from ElasticSearch >= 7.8, got version < 7.0.0")
	}

	if err != nil {
		return err
	}
	d.SetId("")
	return nil
}

func elastic7DeleteIndexTemplate(client *elastic7.Client, id string) error {
	_, err := client.IndexDeleteIndexTemplate(id).Do(context.TODO())
	return err
}

func resourceElasticsearchPutComposableIndexTemplate(d *schema.ResourceData, meta interface{}, create bool) error {
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
			if elasticVersion.LessThan(minimalESComposableTemplateVersion) {
				err = fmt.Errorf("index_template endpoint only available from ElasticSearch >= 7.8, got version %s", elasticVersion.String())
			} else {
				err = elastic7PutIndexTemplate(client, name, body, create)
			}
		}
	default:
		err = fmt.Errorf("index_template endpoint only available from ElasticSearch >= 7.8, got version < 7.0.0")
	}

	return err
}

func elastic7PutIndexTemplate(client *elastic7.Client, name string, body string, create bool) error {
	_, err := client.IndexPutIndexTemplate(name).BodyString(body).Create(create).Do(context.TODO())
	return err
}
