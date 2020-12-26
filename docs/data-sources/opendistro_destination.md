---
page_title: "elasticsearch_opendistro_destination Data Source - terraform-provider-elasticsearch"
subcategory: ""
description: |-
  elasticsearch_opendistro_destination can be used to retrieve the destination ID by name.
---

# Data Source `elasticsearch_opendistro_destination`

`elasticsearch_opendistro_destination` can be used to retrieve the destination ID by name.

## Example Usage

```terraform
# Example destination in other terraform plan
# resource "elasticsearch_opendistro_destination" "test" {
#   body = <<EOF
# {
#   "name": "my-destination",
#   "type": "slack",
#   "slack": {
#     "url": "http://www.example.com"
#   }
# }
# EOF
# }

data "elasticsearch_opendistro_destination" "test" {
  name = my-destination"
}
```

## Schema

### Required

- **name** (String)

### Optional

- **id** (String) The ID of this resource.

### Read-only

- **body** (Map of String)


