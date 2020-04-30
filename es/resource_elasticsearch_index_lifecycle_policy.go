package es

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	elastic7 "github.com/olivere/elastic/v7"
	elastic6 "gopkg.in/olivere/elastic.v6"
)

func resourceElasticsearchIndexLifecyclePolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceElasticsearchIndexLifecyclePolicyCreate,
		Read:   resourceElasticsearchIndexLifecyclePolicyRead,
		Update: resourceElasticsearchIndexLifecyclePolicyUpdate,
		Delete: resourceElasticsearchIndexLifecyclePolicyDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"body": {
				Type:             schema.TypeString,
				Required:         true,
				DiffSuppressFunc: diffSuppressIndexLifecyclePolicy,
				ValidateFunc:     validation.ValidateJsonString,
			},
		},
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
	}
}

func resourceElasticsearchIndexLifecyclePolicyCreate(d *schema.ResourceData, meta interface{}) error {
	err := resourceElasticsearchPutIndexLifecyclePolicy(d, meta)
	if err != nil {
		return err
	}
	d.SetId(d.Get("name").(string))
	return nil
}

func resourceElasticsearchIndexLifecyclePolicyRead(d *schema.ResourceData, meta interface{}) error {
	id := d.Id()

	var result string
	var err error
	switch client := meta.(type) {
	case *elastic7.Client:
		result, err = elastic7IndexGetLifecyclePolicy(client, id)
	case *elastic6.Client:
		result, err = elastic6IndexGetLifecyclePolicy(client, id)
	default:
		err = errors.New("Index Lifecycle Management is only supported by the elastic library >= v6!")
	}
	if err != nil {
		return err
	}

	ds := &resourceDataSetter{d: d}
	ds.set("name", d.Id())
	ds.set("body", result)
	return ds.err
}

func elastic7IndexGetLifecyclePolicy(client *elastic7.Client, id string) (string, error) {
	res, err := client.XPackIlmGetLifecycle().Policy(id).Do(context.TODO())
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

func elastic6IndexGetLifecyclePolicy(client *elastic6.Client, id string) (string, error) {
	res, err := client.XPackIlmGetLifecycle().Policy(id).Do(context.TODO())
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

func resourceElasticsearchIndexLifecyclePolicyUpdate(d *schema.ResourceData, meta interface{}) error {
	return resourceElasticsearchPutIndexLifecyclePolicy(d, meta)
}

func resourceElasticsearchIndexLifecyclePolicyDelete(d *schema.ResourceData, meta interface{}) error {
	id := d.Id()

	var err error
	switch client := meta.(type) {
	case *elastic7.Client:
		err = elastic7IndexDeleteLifecyclePolicy(client, id)
	case *elastic6.Client:
		err = elastic6IndexDeleteLifecyclePolicy(client, id)
	default:
		err = errors.New("Index Lifecycle Management is only supported by the elastic library >= v6!")
	}

	if err != nil {
		return err
	}
	d.SetId("")
	return nil
}

func elastic7IndexDeleteLifecyclePolicy(client *elastic7.Client, id string) error {
	_, err := client.XPackIlmDeleteLifecycle().Policy(id).Do(context.TODO())
	return err
}

func elastic6IndexDeleteLifecyclePolicy(client *elastic6.Client, id string) error {
	_, err := client.XPackIlmDeleteLifecycle().Policy(id).Do(context.TODO())
	return err
}

func resourceElasticsearchPutIndexLifecyclePolicy(d *schema.ResourceData, meta interface{}) error {
	name := d.Get("name").(string)
	body := d.Get("body").(string)

	var err error
	switch client := meta.(type) {
	case *elastic7.Client:
		err = elastic7IndexPutLifecyclePolicy(client, name, body)
	case *elastic6.Client:
		err = elastic6IndexPutLifecyclePolicy(client, name, body)
	default:
		err = errors.New("resourceElasticsearchPutIndexLifecyclePolicy Index Lifecycle Management is only supported by the elastic library >= v6!")
	}

	return err
}

func elastic7IndexPutLifecyclePolicy(client *elastic7.Client, name string, body string) error {
	_, err := client.XPackIlmPutLifecycle().Policy(name).BodyString(body).Do(context.TODO())
	return err
}

func elastic6IndexPutLifecyclePolicy(client *elastic6.Client, name string, body string) error {
	_, err := client.XPackIlmPutLifecycle().Policy(name).BodyString(body).Do(context.TODO())
	return err
}
