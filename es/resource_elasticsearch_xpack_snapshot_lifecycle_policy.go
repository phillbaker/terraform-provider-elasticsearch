package es

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	elastic7 "github.com/olivere/elastic/v7"
)

func resourceElasticsearchXpackSnapshotLifecyclePolicy() *schema.Resource {
	return &schema.Resource{
		Description: "Provides an Elasticsearch XPack snapshot lifecycle management policy. These automatically take snapshots and control how long they are retained. See the upstream [docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/snapshot-lifecycle-management-api.html) for more details.",
		Create:      resourceElasticsearchXpackSnapshotLifecyclePolicyCreate,
		Read:        resourceElasticsearchXpackSnapshotLifecyclePolicyRead,
		Update:      resourceElasticsearchXpackSnapshotLifecyclePolicyUpdate,
		Delete:      resourceElasticsearchXpackSnapshotLifecyclePolicyDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				ForceNew:    true,
				Required:    true,
				Description: "ID for the snapshot lifecycle policy",
			},
			"body": {
				Type:             schema.TypeString,
				Required:         true,
				DiffSuppressFunc: diffSuppressSnapshotLifecyclePolicy,
				ValidateFunc:     validation.StringIsJSON,
				Description:      "See the policy definition defined in the [docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/slm-api-put-policy.html#slm-api-put-request-body)",
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceElasticsearchXpackSnapshotLifecyclePolicyCreate(d *schema.ResourceData, meta interface{}) error {
	err := resourceElasticsearchPutSnapshotLifecyclePolicy(d, meta)
	if err != nil {
		return err
	}
	d.SetId(d.Get("name").(string))
	return nil
}

func resourceElasticsearchXpackSnapshotLifecyclePolicyRead(d *schema.ResourceData, meta interface{}) error {
	id := d.Id()

	var result string
	var err error
	esClient, err := getClient(meta.(*ProviderConf))
	if err != nil {
		return err
	}
	switch client := esClient.(type) {
	case *elastic7.Client:
		result, err = elastic7SnapshotGetLifecyclePolicy(client, id)
	default:
		err = errors.New("Snapshot Lifecycle Management is only supported by the elastic library >= v7!")
	}
	if err != nil {
		return err
	}

	ds := &resourceDataSetter{d: d}
	ds.set("name", d.Id())
	ds.set("body", result)
	return ds.err
}

func elastic7SnapshotGetLifecyclePolicy(client *elastic7.Client, id string) (string, error) {
	res, err := client.PerformRequest(context.TODO(), elastic7.PerformRequestOptions{
		Method: http.MethodGet,
		Path:   "/_slm/policy/" + id,
	})
	if err != nil {
		return "", err
	}

	// GET /_slm/policy/{id} returns a more unique object than other similar
	// API endpoints:
	// this returns a map[policyname]{policy}
	// EVEN IF we specify the id
	// https://www.elastic.co/guide/en/elasticsearch/reference/7.x/slm-api-get-policy.html

	// so we need to do our part to reduce this to just the policy object
	// which is the equivalent of our "body" elsewhere
	var resp map[string]interface{}
	if err := json.Unmarshal(res.Body, &resp); err != nil {
		return "", err
	}

	policy, ok := resp[id]
	if !ok {
		return "", errors.New("Snapshot Lifecycle Management unsuccessfully parsed")
	}

	typedPolicy, ok := policy.(map[string]interface{})
	if !ok {
		return "", errors.New("Snapshot Lifecycle Management unsuccessfully parsed")
	}

	tj, err := json.Marshal(typedPolicy["policy"])
	if err != nil {
		return "", err
	}

	return string(tj), nil
}

func resourceElasticsearchXpackSnapshotLifecyclePolicyUpdate(d *schema.ResourceData, meta interface{}) error {
	return resourceElasticsearchPutSnapshotLifecyclePolicy(d, meta)
}

func resourceElasticsearchXpackSnapshotLifecyclePolicyDelete(d *schema.ResourceData, meta interface{}) error {
	id := d.Id()

	var err error
	esClient, err := getClient(meta.(*ProviderConf))
	if err != nil {
		return err
	}
	switch client := esClient.(type) {
	case *elastic7.Client:
		err = elastic7SnapshotDeleteLifecyclePolicy(client, id)
	default:
		err = errors.New("Snapshot Lifecycle Management is only supported by the elastic library >= v7!")
	}

	if err != nil {
		return err
	}
	d.SetId("")
	return nil
}

func elastic7SnapshotDeleteLifecyclePolicy(client *elastic7.Client, id string) error {
	_, err := client.PerformRequest(context.TODO(), elastic7.PerformRequestOptions{
		Method: http.MethodDelete,
		Path:   "/_slm/policy/" + id,
	})
	return err
}

func resourceElasticsearchPutSnapshotLifecyclePolicy(d *schema.ResourceData, meta interface{}) error {
	name := d.Get("name").(string)
	body := d.Get("body").(string)

	var err error
	esClient, err := getClient(meta.(*ProviderConf))
	if err != nil {
		return err
	}
	switch client := esClient.(type) {
	case *elastic7.Client:
		err = elastic7SnapshotPutLifecyclePolicy(client, name, body)
	default:
		err = errors.New("resourceElasticsearchPutSnapshotLifecyclePolicy Snapshot Lifecycle Management is only supported by the elastic library >= v7!")
	}

	return err
}

func elastic7SnapshotPutLifecyclePolicy(client *elastic7.Client, name string, body string) error {
	_, err := client.PerformRequest(context.TODO(), elastic7.PerformRequestOptions{
		Method: http.MethodPut,
		Path:   "/_slm/policy/" + name,
		Body:   body,
	})
	return err
}
