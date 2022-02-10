package es

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	elastic7 "github.com/olivere/elastic/v7"
	elastic6 "gopkg.in/olivere/elastic.v6"
)

var (
	stringClusterSettings = []string{
		"cluster.persistent_tasks.allocation.enable",
		"cluster.persistent_tasks.allocation.recheck_interval",
		"cluster.info.update.interval",
		"cluster.routing.allocation.allow_rebalance",
		"cluster.routing.allocation.awareness.attributes",
		"cluster.routing.allocation.disk.watermark.high",
		"cluster.routing.allocation.disk.watermark.low",
		"cluster.routing.rebalance.enable",
		"cluster.no_master_block",
		"indices.breaker.fielddata.limit",
		"indices.breaker.request.limit",
		"indices.breaker.total.limit",
		"indices.recovery.max_bytes_per_sec",
		"network.breaker.inflight_requests.limit",
		"script.max_compilations_rate",
		"search.default_search_timeout",
		"action.auto_create_index",
	}
	intClusterSettings = []string{
		"cluster.max_shards_per_node",
		"cluster.max_shards_per_node.frozen",
		"cluster.routing.allocation.cluster_concurrent_rebalance",
		"cluster.routing.allocation.node_concurrent_incoming_recoveries",
		"cluster.routing.allocation.node_concurrent_outgoing_recoveries",
		"cluster.routing.allocation.node_concurrent_recoveries",
		"cluster.routing.allocation.node_initial_primaries_recoveries",
		"cluster.routing.allocation.total_shards_per_node",
	}
	floatClusterSettings = []string{
		"cluster.routing.allocation.balance.index",
		"cluster.routing.allocation.balance.shard",
		"cluster.routing.allocation.balance.threshold",
		"indices.breaker.fielddata.overhead",
		"indices.breaker.request.overhead",
		"network.breaker.inflight_requests.overhead",
	}
	boolClusterSettings = []string{
		"cluster.blocks.read_only",
		"cluster.blocks.read_only_allow_delete",
		"cluster.indices.close.enable",
		"cluster.routing.allocation.disk.include_relocations",
		"cluster.routing.allocation.disk.threshold_enabled",
		"cluster.routing.allocation.enable",
		"cluster.routing.allocation.same_shard.host",
		"action.destructive_requires_name",
	}
	dynamicClusterSettings = concatStringSlice(stringClusterSettings, intClusterSettings, floatClusterSettings, boolClusterSettings)
)

func resourceElasticsearchClusterSettings() *schema.Resource {
	return &schema.Resource{
		Description: "Manages a cluster's (persistent) settings.",
		Create:      resourceElasticsearchClusterSettingsCreate,
		Read:        resourceElasticsearchClusterSettingsRead,
		Update:      resourceElasticsearchClusterSettingsUpdate,
		Delete:      resourceElasticsearchClusterSettingsDelete,
		Schema: map[string]*schema.Schema{
			"cluster_max_shards_per_node": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The total number of primary and replica shards for the cluster, this number is multiplied by the number of non-frozen data nodes; shards for closed indices do not count toward this limit",
			},
			"cluster_max_shards_per_node_frozen": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The total number of primary and replica frozen shards, for the cluster; Ssards for closed indices do not count toward this limit, a cluster with no frozen data nodes is unlimited.",
			},
			"cluster_persistent_tasks_allocation_enable": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Whether allocation for persistent tasks is active (all, none)",
			},
			"cluster_persistent_tasks_allocation_recheck_interval": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "A time string controling how often assignment checks are performed to react to whether persistent tasks can be assigned to nodes",
			},
			"cluster_blocks_read_only": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Make the whole cluster read only and metadata is not allowed to be modified",
			},
			"cluster_blocks_read_only_allow_delete": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Make the whole cluster read only, but allows to delete indices to free up resources",
			},
			"cluster_indices_close_enable": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "If false, you cannot close open indices",
			},
			"cluster_info_update_interval": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "A time string controlling how often Elasticsearch should check on disk usage for each node in the cluster",
			},
			"cluster_routing_allocation_allow_rebalance": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Specify when shard rebalancing is allowed (always, indices_primaries_active, indices_all_active)",
			},
			"cluster_routing_allocation_awareness_attributes": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Use custom node attributes to take hardware configuration into account when allocating shards",
			},
			"cluster_routing_allocation_balance_index": {
				Type:        schema.TypeFloat,
				Optional:    true,
				Description: "Weight factor for the number of shards per index allocated on a node, increasing this raises the tendency to equalize the number of shards per index across all nodes",
			},
			"cluster_routing_allocation_balance_shard": {
				Type:        schema.TypeFloat,
				Optional:    true,
				Description: "Weight factor for the total number of shards allocated on a node, increasing this raises the tendency to equalize the number of shards across all nodes",
			},
			"cluster_routing_allocation_balance_threshold": {
				Type:        schema.TypeFloat,
				Optional:    true,
				Description: "Minimal optimization value of operations that should be performed, raising this will cause the cluster to be less aggressive about optimizing the shard balance",
			},
			"cluster_routing_allocation_cluster_concurrent_rebalance": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "How many concurrent shard rebalances are allowed cluster wide",
			},
			"cluster_routing_allocation_disk_include_relocations": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Whether the allocator will take into account shards that are currently being relocated to the target node when computing a node’s disk usage",
			},
			"cluster_routing_allocation_disk_threshold_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Whether the disk allocation decider is active",
			},
			"cluster_routing_allocation_disk_watermark_high": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Allocator will attempt to relocate shards away from a node whose disk usage is above this percentage disk used",
			},
			"cluster_routing_allocation_disk_watermark_low": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Allocator will not allocate shards to nodes that have more than this percentage disk used",
			},
			"cluster_routing_allocation_enable": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Enable or disable allocation for specific kinds of shards (all, primaries, new_primaries, none)",
			},
			"cluster_routing_allocation_node_concurrent_incoming_recoveries": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "How many incoming recoveries where the target shard (likely the replica unless a shard is relocating) are allocated on the node",
			},
			"cluster_routing_allocation_node_concurrent_outgoing_recoveries": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "How many outgoing recoveries where the source shard (likely the primary unless a shard is relocating) are allocated on the node",
			},
			"cluster_routing_allocation_node_concurrent_recoveries": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "A shortcut to set both incoming and outgoing recoveries",
			},
			"cluster_routing_allocation_node_initial_primaries_recoveries": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Set a (usually) higher rate for primary recovery on node restart (usually from disk, so fast)",
			},
			"cluster_routing_allocation_same_shard_host": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Perform a check to prevent allocation of multiple instances of the same shard on a single host, if multiple nodes are started on the host",
			},
			"cluster_routing_allocation_total_shards_per_node": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Maximum number of primary and replica shards allocated to each node",
			},
			"cluster_routing_rebalance_enable": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Allow rebalancing for specific kinds of shards (all, primaries, replicas, none)",
			},
			"cluster_no_master_block": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Specifies which operations are rejected when there is no active master in a cluster (all, write)",
			},
			"indices_breaker_fielddata_limit": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The percentage of memory above which if loading a field into the field data cache would cause the cache to exceed this limit, an error is returned",
			},
			"indices_breaker_fielddata_overhead": {
				Type:        schema.TypeFloat,
				Optional:    true,
				Description: "A constant that all field data estimations are multiplied by",
			},
			"indices_breaker_request_limit": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The percentabge of memory above which per-request data structures (e.g. calculating aggregations) are prevented from exceeding",
			},
			"indices_breaker_request_overhead": {
				Type:        schema.TypeFloat,
				Optional:    true,
				Description: "A constant that all request estimations are multiplied by",
			},
			"indices_breaker_total_limit": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The percentage of total amount of memory that can be used across all breakers",
			},
			"indices_recovery_max_bytes_per_sec": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Maximum total inbound and outbound recovery traffic for each node, in mb",
			},
			"network_breaker_inflight_requests_limit": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The percentage limit of memory usage on a node of all currently active incoming requests on transport or HTTP level",
			},
			"network_breaker_inflight_requests_overhead": {
				Type:        schema.TypeFloat,
				Optional:    true,
				Description: "A constant that all in flight requests estimations are multiplied by",
			},
			"script_max_compilations_rate": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Limit for the number of unique dynamic scripts within a certain interval that are allowed to be compiled, expressed as compilations divided by a time string",
			},
			"search_default_search_timeout": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "A time string setting a cluster-wide default timeout for all search requests",
			},
			"action_auto_create_index": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^(true|false|([-+]?[a-z0-9][a-z0-9_-]*\*?,?)+)$`), "expected value to be one of: true, false or comma-separated list"),
				Description:  "Whether to automatically create an index if it doesn’t already exist and apply any configured index template",
			},
			"action_destructive_requires_name": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "When set to true, you must specify the index name to delete an index and it is not possible to delete all indices with _all or use wildcards",
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceElasticsearchClusterSettingsCreate(d *schema.ResourceData, meta interface{}) error {
	err := resourceElasticsearchPutClusterSettings(d, meta)
	if err != nil {
		return err
	}
	d.SetId("settings")
	return resourceElasticsearchClusterSettingsRead(d, meta)
}

func resourceElasticsearchPutClusterSettings(d *schema.ResourceData, meta interface{}) error {
	var err error

	esClient, err := getClient(meta.(*ProviderConf))
	if err != nil {
		return err
	}
	settings := make(map[string]interface{})
	settings["persistent"] = clusterSettingsFromResourceData(d)

	body, err := json.Marshal(settings)
	if err != nil {
		return err
	}

	switch client := esClient.(type) {
	case *elastic7.Client:
		// elastic doesn't support PUTing settings: https://github.com/olivere/elastic/issues/1274
		_, err = client.PerformRequest(context.TODO(), elastic7.PerformRequestOptions{
			Method: "PUT",
			Path:   "/_cluster/settings",
			Body:   string(body),
		})
		if err != nil {
			return err
		}
	case *elastic6.Client:
		_, err = client.PerformRequest(context.TODO(), elastic6.PerformRequestOptions{
			Method: "PUT",
			Path:   "/_cluster/settings",
			Body:   string(body),
		})
		if err != nil {
			return err
		}
	default:
		return errors.New("elasticsearch version not supported")
	}

	return err
}

func resourceElasticsearchClusterSettingsRead(d *schema.ResourceData, meta interface{}) error {
	settings, err := resourceElasticsearchClusterSettingsGet(meta)
	if err != nil {
		return err
	}

	return clusterResourceDataFromSettings(settings["persistent"].(map[string]interface{}), d)
}

func resourceElasticsearchClusterSettingsUpdate(d *schema.ResourceData, meta interface{}) error {
	return resourceElasticsearchPutClusterSettings(d, meta)
}

func resourceElasticsearchClusterSettingsDelete(d *schema.ResourceData, meta interface{}) error {
	err := clearAllSettings(meta)
	if err != nil {
		return err
	}

	d.SetId("")
	return err
}

func resourceElasticsearchClusterSettingsGet(meta interface{}) (map[string]interface{}, error) {
	var err error
	var settings map[string]interface{}
	var response *json.RawMessage

	esClient, err := getClient(meta.(*ProviderConf))
	if err != nil {
		return settings, err
	}

	switch client := esClient.(type) {
	case *elastic7.Client:
		var res *elastic7.Response

		res, err := client.PerformRequest(context.TODO(), elastic7.PerformRequestOptions{
			Method: "GET",
			Path:   "/_cluster/settings?flat_settings=true",
		})
		if err != nil {
			return settings, err
		}
		response = &res.Body
	case *elastic6.Client:
		var res *elastic6.Response

		res, err := client.PerformRequest(context.TODO(), elastic6.PerformRequestOptions{
			Method: "GET",
			Path:   "/_cluster/settings?flat_settings=true",
		})
		if err != nil {
			return settings, err
		}
		response = &res.Body
	default:
		return settings, errors.New("elasticsearch version not supported")
	}

	err = json.Unmarshal(*response, &settings)
	if err != nil {
		return settings, fmt.Errorf("fail to unmarshal: %v", err)
	}

	return settings, err
}

func clearAllSettings(meta interface{}) error {
	var err error

	esClient, err := getClient(meta.(*ProviderConf))
	if err != nil {
		return err
	}
	body := `{
		"persistent" : {
			"cluster.*": null,
			"indices.*": null,
			"action.*": null,
			"script.*": null,
			"network.*": null,
			"search.*": null
		}
	  }`

	switch client := esClient.(type) {
	case *elastic7.Client:
		_, err = client.PerformRequest(context.TODO(), elastic7.PerformRequestOptions{
			Method: "PUT",
			Path:   "/_cluster/settings",
			Body:   body,
		})
		if err != nil {
			return err
		}
	case *elastic6.Client:
		_, err = client.PerformRequest(context.TODO(), elastic6.PerformRequestOptions{
			Method: "PUT",
			Path:   "/_cluster/settings",
			Body:   body,
		})
		if err != nil {
			return err
		}
	default:
		return errors.New("elasticsearch version not supported")
	}

	return err
}

func clusterSettingsFromResourceData(d *schema.ResourceData) map[string]interface{} {
	settings := make(map[string]interface{})
	for _, key := range dynamicClusterSettings {
		schemaName := strings.Replace(key, ".", "_", -1)
		if raw, ok := d.GetOk(schemaName); ok {
			log.Printf("[INFO] clusterSettingsFromResourceData: key:%+v schemaName:%+v value:%+v, %+v", key, schemaName, raw, settings)
			settings[key] = raw
		}
	}
	return settings
}

func clusterResourceDataFromSettings(settings map[string]interface{}, d *schema.ResourceData) error {
	log.Printf("[INFO] clusterResourceDataFromSettings: %+v", settings)
	for _, key := range dynamicClusterSettings {
		value, ok := settings[key]
		if !ok {
			continue
		}

		schemaName := strings.Replace(key, ".", "_", -1)
		if containsString(intClusterSettings, key) && reflect.TypeOf(value).String() == "string" {
			var err error
			value, err = strconv.Atoi(value.(string))
			if err != nil {
				return err
			}
		} else if containsString(floatClusterSettings, key) && reflect.TypeOf(value).String() == "string" {
			var err error
			value, err = strconv.ParseFloat(value.(string), 64)
			if err != nil {
				return err
			}
		} else if containsString(boolClusterSettings, key) && reflect.TypeOf(value).String() == "string" {
			var err error
			value, err = strconv.ParseBool(value.(string))
			if err != nil {
				return err
			}
		}
		err := d.Set(schemaName, value)
		if err != nil {
			log.Printf("[ERROR] clusterResourceDataFromSettings: %+v", err)
			return err
		}
	}
	return nil
}
