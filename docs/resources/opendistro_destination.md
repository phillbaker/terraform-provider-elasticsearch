---
page_title: "elasticsearch_opendistro_destination Resource - terraform-provider-elasticsearch"
subcategory: "Elasticsearch Open Distro"
description: |-
  Provides an Elasticsearch OpenDistro destination, a reusable communication channel for an action, such as email, Slack, or a webhook URL. Please refer to the OpenDistro destination documentation https://opendistro.github.io/for-elasticsearch-docs/docs/alerting/monitors/#create-destinations for details.
---

# Resource `elasticsearch_opendistro_destination`

Provides an Elasticsearch OpenDistro destination, a reusable communication channel for an action, such as email, Slack, or a webhook URL. Please refer to the OpenDistro [destination documentation](https://opendistro.github.io/for-elasticsearch-docs/docs/alerting/monitors/#create-destinations) for details.

## Example Usage

```terraform
resource "elasticsearch_opendistro_destination" "test_destination" {
  body = <<EOF
{
  "name": "my-destination",
  "type": "slack",
  "slack": {
    "url": "http://www.example.com"
  }
}
EOF
}
```

## Schema

### Required

- **body** (String) The JSON body of the destination.

### Optional

- **id** (String) The ID of this resource.


