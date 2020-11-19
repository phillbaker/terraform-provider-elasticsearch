---
layout: "elasticsearch"
page_title: "Elasticsearch: elasticsearch_index"
subcategory: "Elasticsearch Opensource"
description: |-
  Provides an Elasticsearch index resource.
---

# elasticsearch_index

Provides an Elasticsearch index resource.

## Example Usage

```tf
# Create a simple index
resource "elasticsearch_index" "test" {
  name = "terraform-test"
  number_of_shards = 1
  number_of_replicas = 1
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the index
* `force_destroy` - (Optional) A boolean that indicates that the index should be deleted even if it contains documents, this exists in Terraform state only.
* `number_of_shards` - (Optional) Passed through to the ES index API. This can be set only on creation.
* `routing_partition_size` - (Optional) Passed through to the ES index API. This can be set only on creation.
* `load_fixed_bitset_filters_eagerly` - (Optional) Passed through to the ES index API. This can be set only on creation.
* `codec` - (Optional) Passed through to the ES index API. This can be set only on creation.
* `number_of_replicas` - (Optional) Passed thorugh to the ES index API. can be changed at runtime.
* `auto_expand_replicas` - (Optional) `0-5` or `0-all`, set the number of replicas to the node count in the cluster, can be changed at runtime.
* `refresh_interval` - (Optional) set to `-1` to disable, can be changed at runtime.
* `mappings` - (Optional) In order to not handle complexities of field mapping updates, updates are not allowed via this provider. See [Elasticsearch docs][1].
* `aliases` - (Optional) In order to not handle the separate endpoint of alias updates, updates are not allowed via this provider currently.


## Attributes Reference

The following attributes are exported:

* `id` - The name of the index.

<!-- External links -->
[1]: https://www.elastic.co/guide/en/elasticsearch/reference/6.8/indices-put-mapping.html#updating-field-mappings


