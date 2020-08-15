package es

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	elastic7 "github.com/olivere/elastic/v7"
	elastic5 "gopkg.in/olivere/elastic.v5"
	elastic6 "gopkg.in/olivere/elastic.v6"
)

func resourceElasticsearchIngestPipeline() *schema.Resource {
	return &schema.Resource{
		Create: resourceElasticsearchIngestPipelineCreate,
		Read:   resourceElasticsearchIngestPipelineRead,
		Update: resourceElasticsearchIngestPipelineUpdate,
		Delete: resourceElasticsearchIngestPipelineDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"body": {
				Type:             schema.TypeString,
				DiffSuppressFunc: diffSuppressIngestPipeline,
				Required:         true,
				ValidateFunc:     validation.StringIsJSON,
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceElasticsearchIngestPipelineCreate(d *schema.ResourceData, meta interface{}) error {

	err := resourceElasticsearchPutIngestPipeline(d, meta)
	if err != nil {
		return err
	}
	d.SetId(d.Get("name").(string))
	return nil
}

func resourceElasticsearchIngestPipelineRead(d *schema.ResourceData, meta interface{}) error {
	id := d.Id()

	var result string
	var err error
	esClient, err := getClient(meta.(*ProviderConf))
	if err != nil {
		return err
	}
	switch client := esClient.(type) {
	case *elastic7.Client:
		result, err = elastic7IngestGetPipeline(client, id)
	case *elastic6.Client:
		result, err = elastic6IngestGetPipeline(client, id)
	default:
		elastic5Client := client.(*elastic5.Client)
		result, err = elastic5IngestGetPipeline(elastic5Client, id)
	}
	if err != nil {
		return err
	}

	ds := &resourceDataSetter{d: d}
	ds.set("name", d.Id())
	ds.set("body", result)
	return ds.err
}

func elastic7IngestGetPipeline(client *elastic7.Client, id string) (string, error) {

	res, err := client.IngestGetPipeline().Pretty(false).Do(context.TODO())
	if err != nil {
		return "", err
	}

	t := res[id]

	tj, err := json.Marshal(t)
	if err != nil {
		return "", err
	}

	return string(tj), nil
}

func elastic6IngestGetPipeline(client *elastic6.Client, id string) (string, error) {
	res, err := client.IngestGetPipeline(id).Do(context.TODO())
	if err != nil {
		return "", err
	}

	t := res[id]
	tj, err := json.Marshal(t)
	if err != nil {
		return "", err
	}
	return string(tj), nil
}

func elastic5IngestGetPipeline(client *elastic5.Client, id string) (string, error) {
	res, err := client.IngestGetPipeline(id).Do(context.TODO())
	if err != nil {
		return "", err
	}

	t := res[id]
	tj, err := json.Marshal(t)
	if err != nil {
		return "", err
	}

	return string(tj), nil
}

func resourceElasticsearchIngestPipelineUpdate(d *schema.ResourceData, meta interface{}) error {
	return resourceElasticsearchPutIngestPipeline(d, meta)
}

func resourceElasticsearchIngestPipelineDelete(d *schema.ResourceData, meta interface{}) error {
	id := d.Id()

	var err error
	esClient, err := getClient(meta.(*ProviderConf))
	if err != nil {
		return err
	}
	switch client := esClient.(type) {
	case *elastic7.Client:
		_, err = client.IngestDeletePipeline(id).Do(context.TODO())
	case *elastic6.Client:
		_, err = client.IngestDeletePipeline(id).Do(context.TODO())
	default:
		elastic5Client := client.(*elastic5.Client)
		_, err = elastic5Client.IngestDeletePipeline(id).Do(context.TODO())
	}

	if err != nil {
		return err
	}
	d.SetId("")
	return nil
}

func resourceElasticsearchPutIngestPipeline(d *schema.ResourceData, meta interface{}) error {
	name := d.Get("name").(string)
	body := d.Get("body").(string)

	var err error
	esClient, err := getClient(meta.(*ProviderConf))
	if err != nil {
		return err
	}
	switch client := esClient.(type) {
	case *elastic7.Client:
		_, err = client.IngestPutPipeline(name).BodyString(body).Do(context.TODO())
	case *elastic6.Client:
		_, err = client.IngestPutPipeline(name).BodyString(body).Do(context.TODO())
	default:
		elastic5Client := client.(*elastic5.Client)
		_, err = elastic5Client.IngestPutPipeline(name).BodyString(body).Do(context.TODO())
	}

	return err
}
