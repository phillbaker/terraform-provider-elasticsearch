---
layout: "elasticsearch"
page_title: "Elasticsearch: elasticsearch_index_template"
subcategory: "Elasticsearch Opensource"
description: |-
  Provides an Elasticsearch index template resource.
---

# elasticsearch_index_template

Provides an Elasticsearch index template resource.

## Example Usage

```tf
# Create an index template
resource "elasticsearch_index_template" "template_1" {
  name = "template_1"
  body = <<EOF
{
  "template": "te*",
  "settings": {
    "number_of_shards": 1
  },
  "mappings": {
    "type1": {
      "_source": {
        "enabled": false
      },
      "properties": {
        "host_name": {
          "type": "keyword"
        },
        "created_at": {
          "type": "date",
          "format": "EEE MMM dd HH:mm:ss Z YYYY"
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

* `name` - (Required) The name of the index template.
* `body` - (Required) The JSON body of the index template.

## Attributes Reference

The following attributes are exported:

* `id` - The name of the index template.

## Import

Index templates can be imported using the `name`, e.g.

```sh
$ terraform import elasticsearch_index_template.template_1 template_1
```
