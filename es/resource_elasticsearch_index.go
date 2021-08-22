package es

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	elastic7 "github.com/olivere/elastic/v7"
	elastic6 "gopkg.in/olivere/elastic.v6"
)

var (
	staticSettingsKeys = []string{
		"number_of_shards",
		"codec",
		"routing_partition_size",
		"load_fixed_bitset_filters_eagerly",
		"shard.check_on_startup",
	}
	dynamicsSettingsKeys = []string{
		"number_of_replicas",
		"auto_expand_replicas",
		"refresh_interval",
		"search.idle.after",
		"max_result_window",
		"max_inner_result_window",
		"max_rescore_window",
		"max_docvalue_fields_search",
		"max_script_fields",
		"max_ngram_diff",
		"max_shingle_diff",
		"blocks.read_only",
		"blocks.read_only_allow_delete",
		"blocks.read",
		"blocks.write",
		"blocks.metadata",
		"max_refresh_listeners",
		"analyze.max_token_count",
		"highlight.max_analyzed_offset",
		"max_terms_count",
		"max_regex_length",
		"routing.allocation.enable",
		"routing.rebalance.enable",
		"gc_deletes",
		"default_pipeline",
		"search.slowlog.threshold.query.warn",
		"search.slowlog.threshold.query.info",
		"search.slowlog.threshold.query.debug",
		"search.slowlog.threshold.query.trace",
		"search.slowlog.threshold.fetch.warn",
		"search.slowlog.threshold.fetch.info",
		"search.slowlog.threshold.fetch.debug",
		"search.slowlog.threshold.fetch.trace",
		"search.slowlog.level",
		"indexing.slowlog.threshold.index.warn",
		"indexing.slowlog.threshold.index.info",
		"indexing.slowlog.threshold.index.debug",
		"indexing.slowlog.threshold.index.trace",
		"indexing.slowlog.level",
		"indexing.slowlog.source",
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
		"include_type_name": {
			Type:        schema.TypeString,
			Description: "A string that indicates if and what we should pass to include_type_name parameter. Set to `\"false\"` when trying to create an index on a v6 cluster without a doc type or set to `\"true\"` when trying to create an index on a v7 cluster with a doc type. Since mapping updates are not currently supported, this applies only on index create.",
			Default:     "",
			Optional:    true,
		},
		// Static settings that can only be set on creation
		"number_of_shards": {
			Type:        schema.TypeString,
			Description: "Number of shards for the index. This can be set only on creation.",
			ForceNew:    true,
			Default:     "1",
			Optional:    true,
		},
		"routing_partition_size": {
			Type:        schema.TypeString,
			Description: "The number of shards a custom routing value can go to. A stringified number. This can be set only on creation.",
			ForceNew:    true,
			Optional:    true,
		},
		"load_fixed_bitset_filters_eagerly": {
			Type:        schema.TypeBool,
			Description: "Indicates whether cached filters are pre-loaded for nested queries. This can be set only on creation.",
			ForceNew:    true,
			Optional:    true,
		},
		"codec": {
			Type:        schema.TypeString,
			Description: "The `default` value compresses stored data with LZ4 compression, but this can be set to `best_compression` which uses DEFLATE for a higher compression ratio. This can be set only on creation.",
			ForceNew:    true,
			Optional:    true,
		},
		"shard_check_on_startup": {
			Type:        schema.TypeString,
			Description: "Whether or not shards should be checked for corruption before opening. When corruption is detected, it will prevent the shard from being opened. Accepts `false`, `true`, `checksum`.",
			ForceNew:    true,
			Optional:    true,
		},
		// Dynamic settings that can be changed at runtime
		"number_of_replicas": {
			Type:        schema.TypeString,
			Description: "Number of shard replicas. A stringified number.",
			Optional:    true,
		},
		"auto_expand_replicas": {
			Type:        schema.TypeString,
			Description: "Set the number of replicas to the node count in the cluster. Set to a dash delimited lower and upper bound (e.g. 0-5) or use all for the upper bound (e.g. 0-all)",
			Optional:    true,
		},
		"refresh_interval": {
			Type:        schema.TypeString,
			Description: "How often to perform a refresh operation, which makes recent changes to the index visible to search. Can be set to `-1` to disable refresh.",
			Optional:    true,
		},
		"search_idle_after": {
			Type:        schema.TypeString,
			Description: "How long a shard can not receive a search or get request until it’s considered search idle.",
			Optional:    true,
		},
		"max_result_window": {
			Type:        schema.TypeString,
			Description: "The maximum value of `from + size` for searches to this index. A stringified number.",
			Optional:    true,
		},
		"max_inner_result_window": {
			Type:        schema.TypeString,
			Description: "The maximum value of `from + size` for inner hits definition and top hits aggregations to this index. A stringified number.",
			Optional:    true,
		},
		"max_rescore_window": {
			Type:        schema.TypeString,
			Description: "The maximum value of `window_size` for `rescore` requests in searches of this index. A stringified number.",
			Optional:    true,
		},
		"max_docvalue_fields_search": {
			Type:        schema.TypeString,
			Description: "The maximum number of `docvalue_fields` that are allowed in a query. A stringified number.",
			Optional:    true,
		},
		"max_script_fields": {
			Type:        schema.TypeString,
			Description: "The maximum number of `script_fields` that are allowed in a query. A stringified number.",
			Optional:    true,
		},
		"max_ngram_diff": {
			Type:        schema.TypeString,
			Description: "The maximum allowed difference between min_gram and max_gram for NGramTokenizer and NGramTokenFilter. A stringified number.",
			Optional:    true,
		},
		"max_shingle_diff": {
			Type:        schema.TypeString,
			Description: "The maximum allowed difference between max_shingle_size and min_shingle_size for ShingleTokenFilter. A stringified number.",
			Optional:    true,
		},
		"max_refresh_listeners": {
			Type:        schema.TypeString,
			Description: "Maximum number of refresh listeners available on each shard of the index. A stringified number.",
			Optional:    true,
		},
		"analyze_max_token_count": {
			Type:        schema.TypeString,
			Description: "The maximum number of tokens that can be produced using _analyze API. A stringified number.",
			Optional:    true,
		},
		"highlight_max_analyzed_offset": {
			Type:        schema.TypeString,
			Description: "The maximum number of characters that will be analyzed for a highlight request. A stringified number.",
			Optional:    true,
		},
		"max_terms_count": {
			Type:        schema.TypeString,
			Description: "The maximum number of terms that can be used in Terms Query. A stringified number.",
			Optional:    true,
		},
		"max_regex_length": {
			Type:        schema.TypeString,
			Description: "The maximum length of regex that can be used in Regexp Query. A stringified number.",
			Optional:    true,
		},
		"blocks_read_only": {
			Type:        schema.TypeBool,
			Description: "Set to `true` to make the index and index metadata read only, `false` to allow writes and metadata changes.",
			Optional:    true,
		},
		"blocks_read_only_allow_delete": {
			Type:        schema.TypeBool,
			Description: "Identical to `index.blocks.read_only` but allows deleting the index to free up resources.",
			Optional:    true,
		},
		"blocks_read": {
			Type:        schema.TypeBool,
			Description: "Set to `true` to disable read operations against the index.",
			Optional:    true,
		},
		"blocks_write": {
			Type:        schema.TypeBool,
			Description: "Set to `true` to disable data write operations against the index. This setting does not affect metadata.",
			Optional:    true,
		},
		"blocks_metadata": {
			Type:        schema.TypeBool,
			Description: "Set to `true` to disable index metadata reads and writes.",
			Optional:    true,
		},
		"routing_allocation_enable": {
			Type:        schema.TypeString,
			Description: "Controls shard allocation for this index. It can be set to: `all` , `primaries` , `new_primaries` , `none`.",
			Optional:    true,
		},
		"routing_rebalance_enable": {
			Type:        schema.TypeString,
			Description: "Enables shard rebalancing for this index. It can be set to: `all`, `primaries` , `replicas` , `none`.",
			Optional:    true,
		},
		"gc_deletes": {
			Type:        schema.TypeString,
			Description: "The length of time that a deleted document's version number remains available for further versioned operations.",
			Optional:    true,
		},
		"default_pipeline": {
			Type:        schema.TypeString,
			Description: "The default ingest node pipeline for this index. Index requests will fail if the default pipeline is set and the pipeline does not exist.",
			Optional:    true,
		},
		"search_slowlog_threshold_query_warn": {
			Type:        schema.TypeString,
			Description: "Set the cutoff for shard level slow search logging of slow searches in the query phase, in time units, e.g. `10s`",
			Optional:    true,
		},
		"search_slowlog_threshold_query_info": {
			Type:        schema.TypeString,
			Description: "Set the cutoff for shard level slow search logging of slow searches in the query phase, in time units, e.g. `5s`",
			Optional:    true,
		},
		"search_slowlog_threshold_query_debug": {
			Type:        schema.TypeString,
			Description: "Set the cutoff for shard level slow search logging of slow searches in the query phase, in time units, e.g. `2s`",
			Optional:    true,
		},
		"search_slowlog_threshold_query_trace": {
			Type:        schema.TypeString,
			Description: "Set the cutoff for shard level slow search logging of slow searches in the query phase, in time units, e.g. `500ms`",
			Optional:    true,
		},
		"search_slowlog_threshold_fetch_warn": {
			Type:        schema.TypeString,
			Description: "Set the cutoff for shard level slow search logging of slow searches in the fetch phase, in time units, e.g. `10s`",
			Optional:    true,
		},
		"search_slowlog_threshold_fetch_info": {
			Type:        schema.TypeString,
			Description: "Set the cutoff for shard level slow search logging of slow searches in the fetch phase, in time units, e.g. `5s`",
			Optional:    true,
		},
		"search_slowlog_threshold_fetch_debug": {
			Type:        schema.TypeString,
			Description: "Set the cutoff for shard level slow search logging of slow searches in the fetch phase, in time units, e.g. `2s`",
			Optional:    true,
		},
		"search_slowlog_threshold_fetch_trace": {
			Type:        schema.TypeString,
			Description: "Set the cutoff for shard level slow search logging of slow searches in the fetch phase, in time units, e.g. `500ms`",
			Optional:    true,
		},
		"search_slowlog_level": {
			Type:        schema.TypeString,
			Description: "Set which logging level to use for the search slow log, can be: `warn`, `info`, `debug`, `trace`",
			Optional:    true,
		},
		"indexing_slowlog_threshold_index_warn": {
			Type:        schema.TypeString,
			Description: "Set the cutoff for shard level slow search logging of slow searches for indexing queries, in time units, e.g. `10s`",
			Optional:    true,
		},
		"indexing_slowlog_threshold_index_info": {
			Type:        schema.TypeString,
			Description: "Set the cutoff for shard level slow search logging of slow searches for indexing queries, in time units, e.g. `5s`",
			Optional:    true,
		},
		"indexing_slowlog_threshold_index_debug": {
			Type:        schema.TypeString,
			Description: "Set the cutoff for shard level slow search logging of slow searches for indexing queries, in time units, e.g. `2s`",
			Optional:    true,
		},
		"indexing_slowlog_threshold_index_trace": {
			Type:        schema.TypeString,
			Description: "Set the cutoff for shard level slow search logging of slow searches for indexing queries, in time units, e.g. `500ms`",
			Optional:    true,
		},
		"indexing_slowlog_level": {
			Type:        schema.TypeString,
			Description: "Set which logging level to use for the search slow log, can be: `warn`, `info`, `debug`, `trace`",
			Optional:    true,
		},
		"indexing_slowlog_source": {
			Type:        schema.TypeString,
			Description: "Set the number of characters of the `_source` to include in the slowlog lines, `false` or `0` will skip logging the source entirely and setting it to `true` will log the entire source regardless of size. The original `_source` is reformatted by default to make sure that it fits on a single log line.",
			Optional:    true,
		},
		// Other attributes
		"mappings": {
			Type:         schema.TypeString,
			Description:  "A JSON string defining how documents in the index, and the fields they contain, are stored and indexed. To avoid the complexities of field mapping updates, updates of this field are not allowed via this provider. See the upstream [Elasticsearch docs](https://www.elastic.co/guide/en/elasticsearch/reference/6.8/indices-put-mapping.html#updating-field-mappings) for more details.",
			Optional:     true,
			ForceNew:     true,
			ValidateFunc: validation.StringIsJSON,
		},
		"aliases": {
			Type:        schema.TypeString,
			Description: "A JSON string describing a set of aliases. The index aliases API allows aliasing an index with a name, with all APIs automatically converting the alias name to the actual index name. An alias can also be mapped to more than one index, and when specifying it, the alias will automatically expand to the aliased indices.",
			Optional:    true,
			// In order to not handle the separate endpoint of alias updates, updates
			// are not allowed via this provider currently.
			ForceNew:     true,
			ValidateFunc: validation.StringIsJSON,
		},
		"analysis_analyzer": {
			Type:         schema.TypeString,
			Description:  "A JSON string describing the analyzers applied to the index.",
			Optional:     true,
			ForceNew:     true, // To add an analyzer, the index must be closed, updated, and then reopened; we can't handle that here.
			ValidateFunc: validation.StringIsJSON,
		},
		"analysis_tokenizer": {
			Type:         schema.TypeString,
			Description:  "A JSON string describing the tokenizers applied to the index.",
			Optional:     true,
			ForceNew:     true, // To add a tokenizer, the index must be closed, updated, and then reopened; we can't handle that here.
			ValidateFunc: validation.StringIsJSON,
		},
		"analysis_filter": {
			Type:         schema.TypeString,
			Description:  "A JSON string describing the filters applied to the index.",
			Optional:     true,
			ForceNew:     true, // To add a filter, the index must be closed, updated, and then reopened; we can't handle that here.
			ValidateFunc: validation.StringIsJSON,
		},
		"analysis_normalizer": {
			Type:         schema.TypeString,
			Description:  "A JSON string describing the normalizers applied to the index.",
			Optional:     true,
			ForceNew:     true, // To add a normalizer, the index must be closed, updated, and then reopened; we can't handle that here.
			ValidateFunc: validation.StringIsJSON,
		},
		// Computed attributes
		"rollover_alias": {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
		},
	}
)

func resourceElasticsearchIndex() *schema.Resource {
	return &schema.Resource{
		Description: "Provides an Elasticsearch index resource.",
		Create:      resourceElasticsearchIndexCreate,
		Read:        resourceElasticsearchIndexRead,
		Update:      resourceElasticsearchIndexUpdate,
		Delete:      resourceElasticsearchIndexDelete,
		Schema:      configSchema,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
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
	}

	if aliasJSON, ok := d.GetOk("aliases"); ok {
		var aliases map[string]interface{}
		bytes := []byte(aliasJSON.(string))
		err = json.Unmarshal(bytes, &aliases)
		if err != nil {
			return fmt.Errorf("fail to unmarshal: %v", err)
		}
		body["aliases"] = aliases
	}

	analysis := map[string]interface{}{}
	settings["analysis"] = analysis

	if analyzerJSON, ok := d.GetOk("analysis_analyzer"); ok {
		var analyzer map[string]interface{}
		bytes := []byte(analyzerJSON.(string))
		err = json.Unmarshal(bytes, &analyzer)
		if err != nil {
			return fmt.Errorf("fail to unmarshal: %v", err)
		}
		analysis["analyzer"] = analyzer
	}
	if tokenizerJSON, ok := d.GetOk("analysis_tokenizer"); ok {
		var tokenizer map[string]interface{}
		bytes := []byte(tokenizerJSON.(string))
		err = json.Unmarshal(bytes, &tokenizer)
		if err != nil {
			return fmt.Errorf("fail to unmarshal: %v", err)
		}
		analysis["tokenizer"] = tokenizer
	}
	if filterJSON, ok := d.GetOk("analysis_filter"); ok {
		var filter map[string]interface{}
		bytes := []byte(filterJSON.(string))
		err = json.Unmarshal(bytes, &filter)
		if err != nil {
			return fmt.Errorf("fail to unmarshal: %v", err)
		}
		analysis["filter"] = filter
	}
	if normalizerJSON, ok := d.GetOk("analysis_normalizer"); ok {
		var normalizer map[string]interface{}
		bytes := []byte(normalizerJSON.(string))
		err = json.Unmarshal(bytes, &normalizer)
		if err != nil {
			return fmt.Errorf("fail to unmarshal: %v", err)
		}
		analysis["normalizer"] = normalizer
	}

	if mappingsJSON, ok := d.GetOk("mappings"); ok {
		var mappings map[string]interface{}
		bytes := []byte(mappingsJSON.(string))
		err = json.Unmarshal(bytes, &mappings)
		if err != nil {
			return fmt.Errorf("fail to unmarshal: %v", err)
		}
		body["mappings"] = mappings
	}

	// if date math is used, we need to pass the resolved name along to the read
	// so we can pull the right result from the response
	var resolvedName string

	// Note: the CreateIndex call handles URL encoding under the hood to handle
	// non-URL friendly characters and functionality like date math
	esClient, err := getClient(meta.(*ProviderConf))
	if err != nil {
		return err
	}
	switch client := esClient.(type) {
	case *elastic7.Client:
		put := client.CreateIndex(name)
		if d.Get("include_type_name").(string) == "true" {
			put = put.IncludeTypeName(true)
		} else if d.Get("include_type_name").(string) == "false" {
			put = put.IncludeTypeName(false)
		}
		resp, requestErr := put.BodyJson(body).Do(ctx)
		err = requestErr
		if err == nil {
			resolvedName = resp.Index
		}

	case *elastic6.Client:
		put := client.CreateIndex(name)
		if d.Get("include_type_name").(string) == "true" {
			put = put.IncludeTypeName(true)
		} else if d.Get("include_type_name").(string) == "false" {
			put = put.IncludeTypeName(false)
		}
		resp, requestErr := put.BodyJson(body).Do(ctx)
		err = requestErr
		if err == nil {
			resolvedName = resp.Index
		}

	default:
		return errors.New("Elasticsearch version not supported")
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
		schemaName := strings.Replace(key, ".", "_", -1)
		if raw, ok := d.GetOk(schemaName); ok {
			log.Printf("[INFO] settingsFromIndexResourceData: key:%+v schemaName:%+v value:%+v, %+v", key, schemaName, raw, settings)
			settings[key] = raw
		}
	}
	return settings
}

func indexResourceDataFromSettings(settings map[string]interface{}, d *schema.ResourceData) {
	log.Printf("[INFO] indexResourceDataFromSettings: %+v", settings)
	for _, key := range settingsKeys {
		rawValue, okRaw := settings[key]
		rawPrefixedValue, okPrefixed := settings["index."+key]
		var value interface{}
		if !okRaw && !okPrefixed {
			continue
		} else if okRaw {
			value = rawValue
		} else if okPrefixed {
			value = rawPrefixedValue
		}

		schemaName := strings.Replace(key, ".", "_", -1)
		err := d.Set(schemaName, value)
		if err != nil {
			log.Printf("[ERROR] indexResourceDataFromSettings: %+v", err)
		}
	}
}

func resourceElasticsearchIndexDelete(d *schema.ResourceData, meta interface{}) error {
	var (
		name = d.Id()
		ctx  = context.Background()
		err  error
	)

	if alias, ok := d.GetOk("rollover_alias"); ok {
		name = getWriteIndexByAlias(alias.(string), d, meta)
	}

	// check to see if there are documents in the index
	allowed := allowIndexDestroy(name, d, meta)
	if !allowed {
		return fmt.Errorf("There are documents in the index (or the index could not be , set force_destroy to true to allow destroying.")
	}

	esClient, err := getClient(meta.(*ProviderConf))
	if err != nil {
		return err
	}
	switch client := esClient.(type) {
	case *elastic7.Client:
		_, err = client.DeleteIndex(name).Do(ctx)

	case *elastic6.Client:
		_, err = client.DeleteIndex(name).Do(ctx)

	default:
		err = errors.New("Elasticsearch version not supported")
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
	esClient, err := getClient(meta.(*ProviderConf))
	if err != nil {
		return false
	}
	switch client := esClient.(type) {
	case *elastic7.Client:
		count, err = client.Count(indexName).Do(ctx)

	case *elastic6.Client:
		count, err = client.Count(indexName).Do(ctx)

	default:
		err = errors.New("Elasticsearch version not supported")
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
		schemaName := strings.Replace(key, ".", "_", -1)
		if d.HasChange(schemaName) {
			settings[key] = d.Get(schemaName)
		}
	}

	// if we're not changing any settings, no-op this function
	if len(settings) == 0 {
		return resourceElasticsearchIndexRead(d, meta)
	}

	body := map[string]interface{}{
		// Note you do not have to explicitly specify the `index` section inside
		// the `settings` section
		"settings": settings,
	}

	var (
		name = d.Id()
		ctx  = context.Background()
		err  error
	)

	if alias, ok := d.GetOk("rollover_alias"); ok {
		name = getWriteIndexByAlias(alias.(string), d, meta)
	}

	esClient, err := getClient(meta.(*ProviderConf))
	if err != nil {
		return err
	}
	switch client := esClient.(type) {
	case *elastic7.Client:
		_, err = client.IndexPutSettings(name).BodyJson(body).Do(ctx)
	case *elastic6.Client:
		_, err = client.IndexPutSettings(name).BodyJson(body).Do(ctx)
	default:
		return errors.New("Elasticsearch version not supported")
	}

	if err == nil {
		return resourceElasticsearchIndexRead(d, meta.(*ProviderConf))
	}
	return err
}

func getWriteIndexByAlias(alias string, d *schema.ResourceData, meta interface{}) string {
	var (
		index   = d.Id()
		ctx     = context.Background()
		columns = []string{"index", "is_write_index"}
	)

	esClient, err := getClient(meta.(*ProviderConf))
	if err != nil {
		log.Printf("[INFO] getWriteIndexByAlias: %+v", err)
		return index
	}
	switch client := esClient.(type) {
	case *elastic7.Client:
		r, err := client.CatAliases().Alias(alias).Columns(columns...).Do(ctx)
		if err != nil {
			log.Printf("[INFO] getWriteIndexByAlias: %+v", err)
			return index
		}
		for _, column := range r {
			if column.IsWriteIndex == "true" {
				return column.Index
			}
		}

	case *elastic6.Client:
		r, err := client.CatAliases().Alias(alias).Columns(columns...).Do(ctx)
		if err != nil {
			log.Printf("[INFO] getWriteIndexByAlias: %+v", err)
			return index
		}
		for _, column := range r {
			if column.IsWriteIndex == "true" {
				return column.Index
			}
		}

	default:
		log.Printf("[INFO] Elasticsearch version not supported")
	}

	return index
}

func resourceElasticsearchIndexRead(d *schema.ResourceData, meta interface{}) error {
	var (
		index    = d.Id()
		ctx      = context.Background()
		settings map[string]interface{}
	)

	if alias, ok := d.GetOk("rollover_alias"); ok {
		index = getWriteIndexByAlias(alias.(string), d, meta)
	}

	// The logic is repeated strictly because of the types
	esClient, err := getClient(meta.(*ProviderConf))
	if err != nil {
		return err
	}
	switch client := esClient.(type) {
	case *elastic7.Client:
		r, err := client.IndexGetSettings(index).FlatSettings(true).Do(ctx)
		if err != nil {
			if elastic7.IsNotFound(err) {
				log.Printf("[WARN] Index (%s) not found, removing from state", index)
				d.SetId("")
				return nil
			}

			return err
		}

		if resp, ok := r[index]; ok {
			settings = resp.Settings
		}
	case *elastic6.Client:
		r, err := client.IndexGetSettings(index).FlatSettings(true).Do(ctx)
		if err != nil {
			if elastic6.IsNotFound(err) {
				log.Printf("[WARN] Index (%s) not found, removing from state", index)
				d.SetId("")
				return nil
			}
			return err
		}

		if resp, ok := r[index]; ok {
			settings = resp.Settings
		}
	default:
		return errors.New("Elasticsearch version not supported")
	}

	// Don't override name otherwise it will force a replacement
	if _, ok := d.GetOk("name"); !ok {
		name := index
		if providedName, ok := settings["index.provided_name"].(string); ok {
			name = providedName
		}
		err := d.Set("name", name)
		if err != nil {
			return err
		}
	}

	// If index is managed by ILM or ISM set rollover_alias
	if alias, ok := settings["index.lifecycle.rollover_alias"].(string); ok {
		err := d.Set("rollover_alias", alias)
		if err != nil {
			return err
		}
	} else if alias, ok := settings["index.opendistro.index_state_management.rollover_alias"].(string); ok {
		err := d.Set("rollover_alias", alias)
		if err != nil {
			return err
		}
	}

	indexResourceDataFromSettings(settings, d)

	return nil
}
