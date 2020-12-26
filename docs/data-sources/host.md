---
page_title: "elasticsearch_host Data Source - terraform-provider-elasticsearch"
subcategory: ""
description: |-
  elasticsearch_host can be used to retrieve the host URL for the provider's current elasticsearch cluster.
---

# Data Source `elasticsearch_host`

`elasticsearch_host` can be used to retrieve the host URL for the provider's current elasticsearch cluster.

## Example Usage

```terraform
data "elasticsearch_host" "test" {
  active = true
}
```

## Schema

### Required

- **active** (Boolean)

### Optional

- **id** (String) The ID of this resource.

### Read-only

- **url** (String)


