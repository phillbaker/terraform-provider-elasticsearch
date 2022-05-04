package es

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/go-version"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/olivere/elastic/uritemplates"
	elastic7 "github.com/olivere/elastic/v7"
)

var minimalESDataStreamVersion, _ = version.NewVersion("7.9.0")

func resourceElasticsearchDataStream() *schema.Resource {
	return &schema.Resource{
		Description: "A data stream lets you store append-only time series data across multiple (hidden, auto-generated) indices while giving you a single named resource for requests. See the [guide](https://www.elastic.co/guide/en/elasticsearch/reference/7.17/data-streams.html) and [API docs](https://www.elastic.co/guide/en/elasticsearch/reference/7.17/data-stream-apis.html).",
		Create:      resourceElasticsearchDataStreamCreate,
		Read:        resourceElasticsearchDataStreamRead,
		Delete:      resourceElasticsearchDataStreamDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				ForceNew:    true,
				Required:    true,
				Description: "Name of the data stream to create, must have a matching ",
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceElasticsearchDataStreamCreate(d *schema.ResourceData, meta interface{}) error {
	err := resourceElasticsearchPutDataStream(d, meta)
	if err != nil {
		return err
	}
	d.SetId(d.Get("name").(string))
	return resourceElasticsearchDataStreamRead(d, meta)
}

func resourceElasticsearchDataStreamAvailable(v *version.Version, c *ProviderConf) bool {
	return v.GreaterThanOrEqual(minimalESDataStreamVersion) || c.flavor == Unknown
}

func resourceElasticsearchDataStreamRead(d *schema.ResourceData, meta interface{}) error {
	id := d.Id()

	var elasticVersion *version.Version

	providerConf := meta.(*ProviderConf)
	esClient, err := getClient(providerConf)
	if err != nil {
		return err
	}

	switch client := esClient.(type) {
	case *elastic7.Client:
		elasticVersion, err = version.NewVersion(providerConf.esVersion)
		if err == nil {
			if resourceElasticsearchDataStreamAvailable(elasticVersion, providerConf) {
				err = elastic7GetDataStream(client, id)
			} else {
				err = fmt.Errorf("_data_stream endpoint only available from ElasticSearch >= 7.9, got version %s", elasticVersion.String())
			}
		}
	default:
		err = fmt.Errorf("_data_stream endpoint only available from ElasticSearch >= 7.9, got version < 7.0.0")
	}
	if err != nil {
		if elastic7.IsNotFound(err) {
			log.Printf("[WARN] data stream (%s) not found, removing from state", id)
			d.SetId("")
			return nil
		}

		return err
	}

	ds := &resourceDataSetter{d: d}
	ds.set("name", d.Id())
	return ds.err
}

func resourceElasticsearchDataStreamDelete(d *schema.ResourceData, meta interface{}) error {
	id := d.Id()

	var elasticVersion *version.Version

	providerConf := meta.(*ProviderConf)
	esClient, err := getClient(providerConf)
	if err != nil {
		return err
	}

	switch client := esClient.(type) {
	case *elastic7.Client:
		elasticVersion, err = version.NewVersion(providerConf.esVersion)
		if err == nil {
			if resourceElasticsearchDataStreamAvailable(elasticVersion, providerConf) {
				err = elastic7DeleteDataStream(client, id)
			} else {
				err = fmt.Errorf("_data_stream endpoint only available from ElasticSearch >= 7.9, got version %s", elasticVersion.String())
			}
		}
	default:
		err = fmt.Errorf("_data_stream endpoint only available from ElasticSearch >= 7.9, got version < 7.0.0")
	}

	if err != nil {
		return err
	}
	d.SetId("")
	return nil
}

func resourceElasticsearchPutDataStream(d *schema.ResourceData, meta interface{}) error {
	name := d.Get("name").(string)

	var elasticVersion *version.Version

	providerConf := meta.(*ProviderConf)
	esClient, err := getClient(providerConf)
	if err != nil {
		return err
	}

	switch client := esClient.(type) {
	case *elastic7.Client:
		elasticVersion, err = version.NewVersion(providerConf.esVersion)
		if err == nil {
			if resourceElasticsearchDataStreamAvailable(elasticVersion, providerConf) {
				err = elastic7PutDataStream(client, name)
			} else {
				err = fmt.Errorf("_data_stream endpoint only available from ElasticSearch >= 7.9, got version %s", elasticVersion.String())
			}
		}
	default:
		err = fmt.Errorf("_data_stream endpoint only available from ElasticSearch >= 7.9, got version < 7.0.0")
	}

	return err
}

func elastic7GetDataStream(client *elastic7.Client, id string) error {
	path, err := uritemplates.Expand("/_data_stream/{id}", map[string]string{
		"id": id,
	})
	if err != nil {
		return fmt.Errorf("error building URL path for data stream: %+v", err)
	}

	_, err = client.PerformRequest(context.TODO(), elastic7.PerformRequestOptions{
		Method: "GET",
		Path:   path,
	})
	return err
}

func elastic7DeleteDataStream(client *elastic7.Client, id string) error {
	path, err := uritemplates.Expand("/_data_stream/{id}", map[string]string{
		"id": id,
	})
	if err != nil {
		return fmt.Errorf("error building URL path for data stream: %+v", err)
	}

	_, err = client.PerformRequest(context.TODO(), elastic7.PerformRequestOptions{
		Method: "DELETE",
		Path:   path,
	})
	return err
}

func elastic7PutDataStream(client *elastic7.Client, id string) error {
	path, err := uritemplates.Expand("/_data_stream/{id}", map[string]string{
		"id": id,
	})
	if err != nil {
		return fmt.Errorf("error building URL path for data stream: %+v", err)
	}

	_, err = client.PerformRequest(context.TODO(), elastic7.PerformRequestOptions{
		Method: "PUT",
		Path:   path,
	})
	return err
}
