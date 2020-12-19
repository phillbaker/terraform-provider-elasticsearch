---
layout: "elasticsearch"
page_title: "Elasticsearch: elasticsearch_xpack_index_lifecycle_policy"
subcategory: "Elasticsearch Xpack"
description: |-
  Provides an Elasticsearch xpack index lifecycle policy resource.
---

# elasticsearch_xpack_index_lifecycle_policy

Provides an Elasticsearch xpack index lifecycle policy resource. Please see [docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/snapshot-lifecycle-management-api.html) for more details on usage.

## Example Usage

```tf
# Create an xpack index_lifecycle_policy
resource "elasticsearch_xpack_index_lifecycle_policy" "test" {
  name = "terraform-test"
  body = <<EOF
{
  "policy": {
    "phases": {
      "warm": {
        "min_age": "10d",
        "actions": {
          "forcemerge": {
            "max_num_segments": 1
          }
        }
      },
      "delete": {
        "min_age": "30d",
        "actions": {
          "delete": {
          	"delete_searchable_snapshot": true
          }
        }
      }
    }
  }
}
EOF
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the xpack index_lifecycle_policy.
* `body` - (Required) The JSON body of the xpack index_lifecycle_policy.

## Attributes Reference

The following attributes are exported:

* `id` - The name of the xpack index_lifecycle_policy.
