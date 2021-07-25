package es

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	elastic7 "github.com/olivere/elastic/v7"
	elastic6 "gopkg.in/olivere/elastic.v6"
)

func resourceElasticsearchSnapshotRepository() *schema.Resource {
	return &schema.Resource{
		Create: resourceElasticsearchSnapshotRepositoryCreate,
		Read:   resourceElasticsearchSnapshotRepositoryRead,
		Update: resourceElasticsearchSnapshotRepositoryUpdate,
		Delete: resourceElasticsearchSnapshotRepositoryDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
			},
			"settings": {
				Type:     schema.TypeMap,
				Optional: true,
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceElasticsearchSnapshotRepositoryCreate(d *schema.ResourceData, meta interface{}) error {
	err := resourceElasticsearchSnapshotRepositoryUpdate(d, meta)
	if err != nil {
		return err
	}
	d.SetId(d.Get("name").(string))
	return nil
}

func resourceElasticsearchSnapshotRepositoryRead(d *schema.ResourceData, meta interface{}) error {
	id := d.Id()

	var repositoryType string
	var settings map[string]interface{}
	var err error
	esClient, err := getClient(meta.(*ProviderConf))
	if err != nil {
		return err
	}
	switch client := esClient.(type) {
	case *elastic7.Client:
		repositoryType, settings, err = elastic7SnapshotGetRepository(client, id)
	case *elastic6.Client:
		repositoryType, settings, err = elastic6SnapshotGetRepository(client, id)
	default:
		return errors.New("Elasticsearch version not supported")
	}

	if err != nil {
		return err
	}

	ds := &resourceDataSetter{d: d}
	ds.set("name", id)
	ds.set("type", repositoryType)
	ds.set("settings", settings)
	return ds.err
}

func elastic7SnapshotGetRepository(client *elastic7.Client, id string) (string, map[string]interface{}, error) {
	repos, err := client.SnapshotGetRepository(id).Do(context.TODO())
	if err != nil {
		return "", make(map[string]interface{}), err
	}

	return repos[id].Type, repos[id].Settings, nil
}

func elastic6SnapshotGetRepository(client *elastic6.Client, id string) (string, map[string]interface{}, error) {
	repos, err := client.SnapshotGetRepository(id).Do(context.TODO())
	if err != nil {
		return "", make(map[string]interface{}), err
	}

	return repos[id].Type, repos[id].Settings, nil
}

func resourceElasticsearchSnapshotRepositoryUpdate(d *schema.ResourceData, meta interface{}) error {
	repositoryType := d.Get("type").(string)
	name := d.Get("name").(string)

	var settings map[string]interface{}

	if v, ok := d.GetOk("settings"); ok {
		settings = v.(map[string]interface{})
	}

	var err error
	esClient, err := getClient(meta.(*ProviderConf))
	if err != nil {
		return err
	}
	switch client := esClient.(type) {
	case *elastic7.Client:
		err = elastic7SnapshotCreateRepository(client, name, repositoryType, settings)
	case *elastic6.Client:
		err = elastic6SnapshotCreateRepository(client, name, repositoryType, settings)
	default:
		return errors.New("Elasticsearch version not supported")
	}

	return err
}

func elastic7SnapshotCreateRepository(client *elastic7.Client, name string, repositoryType string, settings map[string]interface{}) error {
	repo := elastic7.SnapshotRepositoryMetaData{
		Type:     repositoryType,
		Settings: settings,
	}

	_, err := client.SnapshotCreateRepository(name).BodyJson(&repo).Do(context.TODO())
	return err
}

func elastic6SnapshotCreateRepository(client *elastic6.Client, name string, repositoryType string, settings map[string]interface{}) error {
	repo := elastic6.SnapshotRepositoryMetaData{
		Type:     repositoryType,
		Settings: settings,
	}

	_, err := client.SnapshotCreateRepository(name).BodyJson(&repo).Do(context.TODO())
	return err
}

func resourceElasticsearchSnapshotRepositoryDelete(d *schema.ResourceData, meta interface{}) error {
	id := d.Id()

	var err error
	esClient, err := getClient(meta.(*ProviderConf))
	if err != nil {
		return err
	}
	switch client := esClient.(type) {
	case *elastic7.Client:
		err = elastic7SnapshotDeleteRepository(client, id)
	case *elastic6.Client:
		err = elastic6SnapshotDeleteRepository(client, id)
	default:
		return errors.New("Elasticsearch version not supported")
	}

	if err != nil {
		return err
	}
	d.SetId("")
	return nil
}

func elastic7SnapshotDeleteRepository(client *elastic7.Client, id string) error {
	_, err := client.SnapshotDeleteRepository(id).Do(context.TODO())
	return err
}

func elastic6SnapshotDeleteRepository(client *elastic6.Client, id string) error {
	_, err := client.SnapshotDeleteRepository(id).Do(context.TODO())
	return err
}
