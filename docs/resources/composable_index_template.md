---
layout: "elasticsearch"
page_title: "Elasticsearch: elasticsearch_composable_index_template"
subcategory: "Elasticsearch Opensource"
description: |-
  Provides an Elasticsearch Composable index template resource.
---

# elasticsearch_composable_index_template

Provides an Elasticsearch Composable index template resource. This resource uses the `/_index_template`
endpoint of Elasticsearch API that is available since version 7.8. Use `elasticsearch_index_template` if
you are using older versions of Elasticsearch or if you want to keep using legacy Index Templates in Elasticsearch 7.8+.

## Example Usage

```tf
# Create an index template
resource "elasticsearch_composable_index_template" "template_1" {
  name = "template_1"
  body = <<EOF
{
  "index_patterns": ["te*", "bar*"],
  "template": {
    "settings": {
      "index": {
        "number_of_shards": 1
      }
    },
    "mappings": {
      "properties": {
        "host_name": {
          "type": "keyword"
        },
        "created_at": {
          "type": "date",
          "format": "EEE MMM dd HH:mm:ss Z yyyy"
        }
      }
    },
    "aliases": {
      "mydata": { }
    }
  },
  "priority": 200,
  "version": 3
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

Composable index templates can be imported using the `name`, e.g.

```sh
$ terraform import elasticsearch_composable_index_template.template_1 template_1
```
