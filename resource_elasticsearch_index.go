package main

import (
	"context"

	"github.com/hashicorp/terraform/helper/schema"

	elastic7 "github.com/olivere/elastic/v7"
	elastic5 "gopkg.in/olivere/elastic.v5"
	elastic6 "gopkg.in/olivere/elastic.v6"
)

var (
	staticConfigKeys = []string{
		"number_of_shards",
		"codec",
		"routing_partition_size",
		"load_fixed_bitset_filters_eagerly",
	}
	dynamicConfigKeys = []string{
		"number_of_replicas",
		"auto_expand_replicas",
		"refresh_interval",
		//"max_result_window"
		//"max_inner_result_window"
		//"max_rescore_window"
		//...
	}
	configKeys = append(staticConfigKeys, dynamicConfigKeys...)
)

var (
	configSchema = map[string]*schema.Schema{
		// Static settings that can only be set on creation
		"name": &schema.Schema{
			Type:        schema.TypeString,
			Description: "Name of the index to create",
			ForceNew:    true,
			Required:    true,
		},
		"number_of_shards": &schema.Schema{
			Type:        schema.TypeInt,
			Description: "Number of shards for the index",
			ForceNew:    true, // shards can only set upon creation
			Default:     1,
			Optional:    true,
		},
		// "check_on_startup": &schema.Schema{
		// 	Type:     schema.TypeString, // false,checksum,true
		// 	ForceNew: true,
		// 	// Default:  "false",
		// 	Optional: true,
		// },
		"routing_partition_size": &schema.Schema{
			Type:     schema.TypeInt,
			ForceNew: true, // shards can only set upon creation
			// Default:  1,
			Optional: true,
		},
		"load_fixed_bitset_filters_eagerly": &schema.Schema{
			Type:     schema.TypeBool,
			ForceNew: true, // false,checksum,true
			// Default:  true,
			Optional: true,
		},
		"codec": &schema.Schema{
			Type:     schema.TypeString,
			ForceNew: true,
			// Default:  "default",
			Optional: true,
		},
		// Dynamic settings that can be changed at runtime
		"number_of_replicas": &schema.Schema{
			Type:        schema.TypeInt,
			Description: "Number of shard replicas",
			// Default:  1,
			Optional: true,
		},
		"auto_expand_replicas": &schema.Schema{
			Type:        schema.TypeString, // 0-5 OR 0-all
			Description: "Set the number of replicas to the node count in the cluster",
			// Default:  "false",
			Optional: true,
		},
		"refresh_interval": &schema.Schema{
			Type: schema.TypeString, // -1 to disable
			// Default:  "1s",
			Optional: true,
		},
	}
)

// this is a check to ensure consistency when new settings are supported.
// It ensures the listed config keys match the schema as new keys must
// be registered in both places
func init() {
	if len(configKeys) != len(configSchema)-1 {
		panic("declared keys do not match the schema")
	}
}

func resourceElasticsearchIndex() *schema.Resource {
	return &schema.Resource{
		Create: resourceElasticsearchIndexCreate,
		Read:   resourceElasticsearchIndexRead,
		Update: resourceElasticsearchIndexUpdate,
		Delete: resourceElasticsearchIndexDelete,
		Schema: configSchema,
	}
}

func resourceElasticsearchIndexCreate(d *schema.ResourceData, meta interface{}) error {
	var (
		name     = d.Get("name").(string)
		settings = settingsFromResourceData(d)
		body     = make(map[string]interface{})
		ctx      = context.Background()
		err      error
	)

	if len(settings) > 0 {
		body["settings"] = settings
	}

	switch meta.(type) {
	case *elastic7.Client:
		client := meta.(*elastic7.Client)
		_, err = client.CreateIndex(name).BodyJson(body).Do(ctx)

	case *elastic6.Client:
		client := meta.(*elastic6.Client)
		_, err = client.CreateIndex(name).BodyJson(body).Do(ctx)

	default:
		client := meta.(*elastic5.Client)
		_, err = client.CreateIndex(name).BodyJson(body).Do(ctx)
	}

	// return err

	if err == nil {
		// Let terraform know the resource was created
		d.SetId(name)
		return resourceElasticsearchIndexRead(d, meta)
	}
	return err
}

func settingsFromResourceData(d *schema.ResourceData) map[string]interface{} {
	settings := make(map[string]interface{})
	for _, key := range configKeys {
		if raw, ok := d.GetOk(key); ok {
			settings[key] = raw
		}
	}
	return settings
}

func resourceDataFromSettings(settings map[string]interface{}, d *schema.ResourceData) {
	for _, key := range configKeys {
		if val, ok := settings[key]; ok {
			d.Set(key, val)
		}
	}
	// if raw, ok := d.GetOk("check_on_startup"); ok {
	// 	settings["shard.check_on_startup"] = raw.(string)
	// }
}

func elasticIndexCreate(meta interface{}, name string, settings, mappings map[string]interface{}) error {
	var (
		body = make(map[string]interface{})
		ctx  = context.Background()
		err  error
	)

	if len(settings) > 0 {
		body["settings"] = settings
	}

	switch meta.(type) {
	case *elastic7.Client:
		client := meta.(*elastic7.Client)
		_, err = client.CreateIndex(name).BodyJson(body).Do(ctx)

	case *elastic6.Client:
		client := meta.(*elastic6.Client)
		_, err = client.CreateIndex(name).BodyJson(body).Do(ctx)

	default:
		client := meta.(*elastic5.Client)
		_, err = client.CreateIndex(name).BodyJson(body).Do(ctx)
	}

	return err
}

func resourceElasticsearchIndexDelete(d *schema.ResourceData, meta interface{}) error {
	var (
		name = d.Get("name").(string)
		ctx  = context.Background()
		err  error
	)

	switch meta.(type) {
	case *elastic7.Client:
		client := meta.(*elastic7.Client)
		_, err = client.DeleteIndex(name).Do(ctx)

	case *elastic6.Client:
		client := meta.(*elastic6.Client)
		_, err = client.DeleteIndex(name).Do(ctx)

	default:
		client := meta.(*elastic5.Client)
		_, err = client.DeleteIndex(name).Do(ctx)
	}

	return err
}

func resourceElasticsearchIndexUpdate(d *schema.ResourceData, meta interface{}) error {
	settings := make(map[string]interface{})
	for _, key := range configKeys {
		if d.HasChange(key) {
			settings[key] = d.Get(key)
		}
	}
	body := map[string]interface{}{
		"settings": settings,
	}

	var (
		name = d.Get("name").(string)
		ctx  = context.Background()
		err  error
	)

	switch meta.(type) {
	case *elastic7.Client:
		client := meta.(*elastic7.Client)
		_, err = client.IndexPutSettings(name).BodyJson(body).Do(ctx)

	case *elastic6.Client:
		client := meta.(*elastic6.Client)
		_, err = client.IndexPutSettings(name).BodyJson(body).Do(ctx)

	default:
		client := meta.(*elastic5.Client)
		_, err = client.IndexPutSettings(name).BodyJson(body).Do(ctx)
	}

	if err == nil {
		return resourceElasticsearchIndexRead(d, meta)
	}
	return err
}

func resourceElasticsearchIndexRead(d *schema.ResourceData, meta interface{}) error {
	var (
		name     = d.Get("name").(string)
		ctx      = context.Background()
		settings map[string]interface{}
	)

	// The logic is repeated strictly becuase of the types
	switch meta.(type) {
	case *elastic7.Client:
		client := meta.(*elastic7.Client)
		r, err := client.IndexGet(name).Do(ctx)
		if err != nil {
			return err
		}

		resp := r[name]
		// aliases = resp.Aliases
		settings = resp.Settings["index"].(map[string]interface{})

	case *elastic6.Client:
		client := meta.(*elastic6.Client)
		r, err := client.IndexGet(name).Do(ctx)
		if err != nil {
			return err
		}

		resp := r[name]
		// aliases = resp.Aliases
		settings = resp.Settings

	default:
		client := meta.(*elastic5.Client)
		r, err := client.IndexGet(name).Do(ctx)
		if err != nil {
			return err
		}

		resp := r[name]
		// aliases = resp.Aliases
		settings = resp.Settings

	}

	d.Set("name", name)
	resourceDataFromSettings(settings, d)

	return nil
}
