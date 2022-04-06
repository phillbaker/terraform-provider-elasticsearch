---
layout: "elasticsearch"
page_title: "Elasticsearch: elasticsearch_script"
subcategory: "Elasticsearch Opensource"
description: |-
  Provides an Elasticsearch Opensource script resource.
---

# elasticsearch_script

Provides an Elasticsearch script resource.

## Example Usage

```tf
# Create a script
resource "elasticsearch_script" "script_1" {
  script_id = "script_1"
  body = <<EOF
{
  "script": {
	"lang": "painless",
	"source": "Math.log(_score * 2) + params.my_modifier"
  }
}
EOF
}
```

## Argument Reference

The following arguments are supported:

* `script_id` - (Required) The name of the script.
* `body` - (Required) The JSON body of the script.

## Attributes Reference

The following attributes are exported:

* `id` - The name of the script.
