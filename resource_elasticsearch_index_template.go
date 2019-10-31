package main

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	elastic7 "github.com/olivere/elastic/v7"
	elastic5 "gopkg.in/olivere/elastic.v5"
	elastic6 "gopkg.in/olivere/elastic.v6"
)

func resourceElasticsearchIndexTemplate() *schema.Resource {
	return &schema.Resource{
		Create: resourceElasticsearchIndexTemplateCreate,
		Read:   resourceElasticsearchIndexTemplateRead,
		Update: resourceElasticsearchIndexTemplateUpdate,
		Delete: resourceElasticsearchIndexTemplateDelete,
		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"body": &schema.Schema{
				Type:             schema.TypeString,
				Required:         true,
				DiffSuppressFunc: diffSuppressIndexTemplate,
			},
		},
	}
}

func resourceElasticsearchIndexTemplateCreate(d *schema.ResourceData, meta interface{}) error {
	err := resourceElasticsearchPutIndexTemplate(d, meta, true)
	if err != nil {
		return err
	}
	d.SetId(d.Get("name").(string))
	return nil
}

func resourceElasticsearchIndexTemplateRead(d *schema.ResourceData, meta interface{}) error {
	id := d.Id()

	var result string
	var err error
	switch meta.(type) {
	case *elastic7.Client:
		client := meta.(*elastic7.Client)
		result, err = elastic7IndexGetTemplate(client, id)
	case *elastic6.Client:
		client := meta.(*elastic6.Client)
		result, err = elastic6IndexGetTemplate(client, id)
	default:
		client := meta.(*elastic5.Client)
		result, err = elastic5IndexGetTemplate(client, id)
	}
	if err != nil {
		return err
	}

	d.Set("name", d.Id())
	d.Set("body", result)
	return nil
}

func elastic7IndexGetTemplate(client *elastic7.Client, id string) (string, error) {
	res, err := client.IndexGetTemplate(id).Do(context.TODO())
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

func elastic6IndexGetTemplate(client *elastic6.Client, id string) (string, error) {
	res, err := client.IndexGetTemplate(id).Do(context.TODO())
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

func elastic5IndexGetTemplate(client *elastic5.Client, id string) (string, error) {
	res, err := client.IndexGetTemplate(id).Do(context.TODO())
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

func resourceElasticsearchIndexTemplateUpdate(d *schema.ResourceData, meta interface{}) error {
	return resourceElasticsearchPutIndexTemplate(d, meta, false)
}

func resourceElasticsearchIndexTemplateDelete(d *schema.ResourceData, meta interface{}) error {
	id := d.Id()

	var err error
	switch meta.(type) {
	case *elastic7.Client:
		client := meta.(*elastic7.Client)
		err = elastic7IndexDeleteTemplate(client, id)
	case *elastic6.Client:
		client := meta.(*elastic6.Client)
		err = elastic6IndexDeleteTemplate(client, id)
	default:
		client := meta.(*elastic5.Client)
		err = elastic5IndexDeleteTemplate(client, id)
	}

	if err != nil {
		return err
	}
	d.SetId("")
	return nil
}

func elastic7IndexDeleteTemplate(client *elastic7.Client, id string) error {
	_, err := client.IndexDeleteTemplate(id).Do(context.TODO())
	return err
}

func elastic6IndexDeleteTemplate(client *elastic6.Client, id string) error {
	_, err := client.IndexDeleteTemplate(id).Do(context.TODO())
	return err
}

func elastic5IndexDeleteTemplate(client *elastic5.Client, id string) error {
	_, err := client.IndexDeleteTemplate(id).Do(context.TODO())
	return err
}

func resourceElasticsearchPutIndexTemplate(d *schema.ResourceData, meta interface{}, create bool) error {
	name := d.Get("name").(string)
	body := d.Get("body").(string)

	var err error
	switch meta.(type) {
	case *elastic7.Client:
		client := meta.(*elastic7.Client)
		err = elastic7IndexPutTemplate(client, name, body, create)
	case *elastic6.Client:
		client := meta.(*elastic6.Client)
		err = elastic6IndexPutTemplate(client, name, body, create)
	default:
		client := meta.(*elastic5.Client)
		err = elastic5IndexPutTemplate(client, name, body, create)
	}

	return err
}

func elastic7IndexPutTemplate(client *elastic7.Client, name string, body string, create bool) error {
	_, err := client.IndexPutTemplate(name).BodyString(body).Create(create).Do(context.TODO())
	return err
}

func elastic6IndexPutTemplate(client *elastic6.Client, name string, body string, create bool) error {
	_, err := client.IndexPutTemplate(name).BodyString(body).Create(create).Do(context.TODO())
	return err
}

func elastic5IndexPutTemplate(client *elastic5.Client, name string, body string, create bool) error {
	_, err := client.IndexPutTemplate(name).BodyString(body).Create(create).Do(context.TODO())
	return err
}
