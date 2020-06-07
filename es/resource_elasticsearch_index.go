package es

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"

	elastic7 "github.com/olivere/elastic/v7"
	elastic5 "gopkg.in/olivere/elastic.v5"
	elastic6 "gopkg.in/olivere/elastic.v6"
)

var (
	staticSettingsKeys = []string{
		"number_of_shards",
		"codec",
		"routing_partition_size",
		"load_fixed_bitset_filters_eagerly",
	}
	dynamicsSettingsKeys = []string{
		"number_of_replicas",
		"auto_expand_replicas",
		"refresh_interval",
		//"max_result_window"
		//"max_inner_result_window"
		//"max_rescore_window"
		//...
	}
	settingsKeys = append(staticSettingsKeys, dynamicsSettingsKeys...)
)

var (
	configSchema = map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Description: "Name of the index to create",
			ForceNew:    true,
			Required:    true,
		},
		"force_destroy": {
			Type:        schema.TypeBool,
			Description: "A boolean that indicates that the index should be deleted even if it contains documents.",
			Default:     false,
			Optional:    true,
		},
		// Static settings that can only be set on creation
		"number_of_shards": {
			Type:        schema.TypeString,
			Description: "Number of shards for the index",
			ForceNew:    true, // shards can only set upon creation
			Default:     "1",
			Optional:    true,
		},
		"routing_partition_size": {
			Type:     schema.TypeInt,
			ForceNew: true, // shards can only set upon creation
			Optional: true,
		},
		"load_fixed_bitset_filters_eagerly": {
			Type:     schema.TypeBool,
			ForceNew: true,
			Optional: true,
		},
		"codec": {
			Type:     schema.TypeString,
			ForceNew: true,
			Optional: true,
		},
		// Dynamic settings that can be changed at runtime
		"number_of_replicas": {
			Type:        schema.TypeString,
			Description: "Number of shard replicas",
			Optional:    true,
		},
		"auto_expand_replicas": {
			Type:        schema.TypeString, // 0-5 OR 0-all
			Description: "Set the number of replicas to the node count in the cluster",
			Optional:    true,
		},
		"refresh_interval": {
			Type:     schema.TypeString, // -1 to disable
			Optional: true,
		},
		// Other attributes
		"mappings": {
			Type:     schema.TypeString,
			Optional: true,
			// In order to not handle complexities of field mapping updates, updates
			// are not allowed via this provider. See
			// https://www.elastic.co/guide/en/elasticsearch/reference/6.8/indices-put-mapping.html#updating-field-mappings.
			ForceNew:     true,
			ValidateFunc: validation.StringIsJSON,
		},
		"aliases": {
			Type:     schema.TypeString,
			Optional: true,
			// In order to not handle the separate endpoint of alias updates, updates
			// are not allowed via this provider currently.
			ForceNew:     true,
			ValidateFunc: validation.StringIsJSON,
		},
	}
)

func resourceElasticsearchIndex() *schema.Resource {
	return &schema.Resource{
		Create: resourceElasticsearchIndexCreate,
		Read:   resourceElasticsearchIndexRead,
		Update: resourceElasticsearchIndexUpdate,
		Delete: resourceElasticsearchIndexDelete,
		Schema: configSchema,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
	}
}

func resourceElasticsearchIndexCreate(d *schema.ResourceData, meta interface{}) error {
	var (
		name     = d.Get("name").(string)
		settings = settingsFromIndexResourceData(d)
		body     = make(map[string]interface{})
		ctx      = context.Background()
		err      error
	)
	if len(settings) > 0 {
		body["settings"] = settings

		if aliasJson, ok := d.GetOk("aliases"); ok {
			var aliases map[string]interface{}
			bytes := []byte(aliasJson.(string))
			err = json.Unmarshal(bytes, &aliases)
			if err != nil {
				return fmt.Errorf("fail to unmarshal: %v", err)
			}
			body["aliases"] = aliases
		}

		if mappingsJson, ok := d.GetOk("mappings"); ok {
			var mappings map[string]interface{}
			bytes := []byte(mappingsJson.(string))
			err = json.Unmarshal(bytes, &mappings)
			if err != nil {
				return fmt.Errorf("fail to unmarshal: %v", err)
			}
			body["mappings"] = mappings
		}
	}

	// if date math is used, we need to pass the resolved name along to the read
	// so we can pull the right result from the response
	var resolvedName string

	// Note: the CreateIndex call handles URL encoding under the hood to handle
	// non-URL friendly characters and functionality like date math
	switch client := meta.(type) {
	case *elastic7.Client:
		resp, requestErr := client.CreateIndex(name).BodyJson(body).Do(ctx)
		resolvedName = resp.Index
		err = requestErr

	case *elastic6.Client:
		resp, requestErr := client.CreateIndex(name).BodyJson(body).Do(ctx)
		resolvedName = resp.Index
		err = requestErr

	default:
		elastic5Client := meta.(*elastic5.Client)
		resp, requestErr := elastic5Client.CreateIndex(name).BodyJson(body).Do(ctx)
		resolvedName = resp.Index
		err = requestErr
	}

	if err == nil {
		// Let terraform know the resource was created
		d.SetId(resolvedName)
		return resourceElasticsearchIndexRead(d, meta)
	}
	return err
}

func settingsFromIndexResourceData(d *schema.ResourceData) map[string]interface{} {
	settings := make(map[string]interface{})
	for _, key := range settingsKeys {
		if raw, ok := d.GetOk(key); ok {
			settings[key] = raw
		}
	}
	return settings
}

func indexResourceDataFromSettings(settings map[string]interface{}, d *schema.ResourceData) {
	for _, key := range settingsKeys {
		err := d.Set(key, settings[key])
		if err != nil {
			log.Printf("[INFO] indexResourceDataFromSettings: %+v", err)
		}
	}
}

func resourceElasticsearchIndexDelete(d *schema.ResourceData, meta interface{}) error {
	var (
		name = d.Id()
		ctx  = context.Background()
		err  error
	)

	// check to see if there are documents in the index
	allowed := allowIndexDestroy(name, d, meta)
	if !allowed {
		return fmt.Errorf("There are documents in the index (or the index could not be , set force_destroy to true to allow destroying.")
	}

	switch client := meta.(type) {
	case *elastic7.Client:
		_, err = client.DeleteIndex(name).Do(ctx)

	case *elastic6.Client:
		_, err = client.DeleteIndex(name).Do(ctx)

	default:
		elastic5Client := meta.(*elastic5.Client)
		_, err = elastic5Client.DeleteIndex(name).Do(ctx)
	}

	return err
}

func allowIndexDestroy(indexName string, d *schema.ResourceData, meta interface{}) bool {
	force := d.Get("force_destroy").(bool)

	var (
		ctx   = context.Background()
		count int64
		err   error
	)
	switch client := meta.(type) {
	case *elastic7.Client:
		count, err = client.Count(indexName).Do(ctx)

	case *elastic6.Client:
		count, err = client.Count(indexName).Do(ctx)

	default:
		elastic5Client := meta.(*elastic5.Client)
		count, err = elastic5Client.Count(indexName).Do(ctx)
	}

	if err != nil {
		log.Printf("[INFO] allowIndexDestroy: %+v", err)
		return false
	}

	if count > 0 && !force {
		return false
	}
	return true
}

func resourceElasticsearchIndexUpdate(d *schema.ResourceData, meta interface{}) error {
	settings := make(map[string]interface{})
	for _, key := range settingsKeys {
		if d.HasChange(key) {
			settings[key] = d.Get(key)
		}
	}

	// if we're not changing any settings, no-op this function
	if len(settings) == 0 {
		return resourceElasticsearchIndexRead(d, meta)
	}

	body := map[string]interface{}{
		"settings": settings,
	}

	var (
		name = d.Id()
		ctx  = context.Background()
		err  error
	)

	switch client := meta.(type) {
	case *elastic7.Client:
		_, err = client.IndexPutSettings(name).BodyJson(body).Do(ctx)

	case *elastic6.Client:
		_, err = client.IndexPutSettings(name).BodyJson(body).Do(ctx)

	default:
		elastic5Client := meta.(*elastic5.Client)
		_, err = elastic5Client.IndexPutSettings(name).BodyJson(body).Do(ctx)
	}

	if err == nil {
		return resourceElasticsearchIndexRead(d, meta)
	}
	return err
}

func resourceElasticsearchIndexRead(d *schema.ResourceData, meta interface{}) error {
	var (
		index    = d.Id()
		ctx      = context.Background()
		settings map[string]interface{}
	)

	// The logic is repeated strictly because of the types
	switch client := meta.(type) {
	case *elastic7.Client:
		r, err := client.IndexGet(index).Do(ctx)
		if err != nil {
			return err
		}

		if resp, ok := r[index]; ok {
			settings = resp.Settings["index"].(map[string]interface{})
		}
	case *elastic6.Client:
		r, err := client.IndexGet(index).Do(ctx)
		if err != nil {
			return err
		}

		if resp, ok := r[index]; ok {
			settings = resp.Settings["index"].(map[string]interface{})
		}
	default:
		elastic5Client := meta.(*elastic5.Client)
		r, err := elastic5Client.IndexGet(index).Do(ctx)
		if err != nil {
			return err
		}

		if resp, ok := r[index]; ok {
			settings = resp.Settings["index"].(map[string]interface{})
		}
	}

	name := index
	if providedName, ok := settings["provided_name"].(string); ok {
		name = providedName
	}
	err := d.Set("name", name)
	if err != nil {
		return err
	}

	indexResourceDataFromSettings(settings, d)

	return nil
}
