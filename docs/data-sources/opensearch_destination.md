---
page_title: "elasticsearch_opensearch_destination Data Source - terraform-provider-elasticsearch"
subcategory: ""
description: |-
  elasticsearch_opensearch_destination can be used to retrieve the destination ID by name.
---

# Data Source `elasticsearch_opensearch_destination`

`elasticsearch_opensearch_destination` can be used to retrieve the destination ID by name.

## Example Usage

```terraform
# Example destination in other terraform plan
# resource "elasticsearch_opensearch_destination" "test" {
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

data "elasticsearch_opensearch_destination" "test" {
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


