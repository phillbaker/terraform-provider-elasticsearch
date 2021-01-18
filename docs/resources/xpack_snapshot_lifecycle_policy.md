---
page_title: "elasticsearch_xpack_snapshot_lifecycle_policy Resource - terraform-provider-elasticsearch"
subcategory: "Elasticsearch Xpack"
description: |-
  Provides an Elasticsearch XPack snapshot lifecycle management policy. These automatically take snapshots and control how long they are retained. See the upstream docs https://www.elastic.co/guide/en/elasticsearch/reference/current/snapshot-lifecycle-management-api.html for more details.
---

# Resource `elasticsearch_xpack_snapshot_lifecycle_policy`

Provides an Elasticsearch XPack snapshot lifecycle management policy. These automatically take snapshots and control how long they are retained. See the upstream [docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/snapshot-lifecycle-management-api.html) for more details.

## Example Usage

```terraform
resource "elasticsearch_xpack_snapshot_lifecycle_policy" "terraform-test" {
  name = "test"
  body = <<EOF
{
  "schedule": "0 30 1 * * ?",
  "name": "<daily-snap-{now/d}>",
  "repository": "terraform-test",
  "config": {
    "indices": ["data-*", "important"],
    "ignore_unavailable": false,
    "include_global_state": false
  },
  "retention": {
    "expire_after": "30d",
    "min_count": 5,
    "max_count": 50
  }
}
EOF
}
```

## Schema

### Required

- **body** (String) See the policy definition defined in the [docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/slm-api-put-policy.html#slm-api-put-request-body)
- **name** (String) ID for the snapshot lifecycle policy

### Optional

- **id** (String) The ID of this resource.


