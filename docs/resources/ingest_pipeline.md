---
layout: "elasticsearch"
page_title: "Elasticsearch: elasticsearch_ingest_pipeline"
subcategory: "Elasticsearch Opensource"
description: |-
  Provides an Elasticsearch ingest pipeline resource.
---

# elasticsearch_ingest_pipeline

Provides an Elasticsearch ingest pipeline resource.

## Example Usage

```tf
# Create a simple ingest pipeline
resource "elasticsearch_ingest_pipeline" "test" {
  name = "terraform-test"
  body = <<EOF
{
  "description" : "describe pipeline",
  "version": 123,
  "processors" : [
    {
      "set" : {
        "field": "foo",
        "value": "bar"
      }
    }
  ]
}
EOF
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the ingest pipeline
* `body` - (Required) The JSON body of the ingest pipeline

## Attributes Reference

The following attributes are exported:

* `id` - The name of the ingest pipeline.

## Import

Ingest pipelines can be imported using the `name`, e.g.

```sh
$ terraform import elasticsearch_ingest_pipeline.test terraform-test
```
