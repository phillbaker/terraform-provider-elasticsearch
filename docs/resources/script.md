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
resource "elasticsearch_script" "test_script" {
  script_id = "my_script"
  lang      = "painless"
  source    = "Math.log(_score * 2) + params.my_modifier"
}
```

## Argument Reference

The following arguments are supported:

* `script_id` - (Required) The name of the script.
* `lang` - Specifies the language the script is written in. Defaults to painless..
* `source` - (Required) The source of the stored script.

## Attributes Reference

The following attributes are exported:

* `id` - The name of the script.

## Import

Scripts can be imported using the `script_id`, e.g.

```sh
$ terraform import elasticsearch_script.test_script my_script
```
