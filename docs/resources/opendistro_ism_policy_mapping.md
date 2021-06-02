---
page_title: "elasticsearch_opendistro_ism_policy_mapping Resource - terraform-provider-elasticsearch"
subcategory: "Elasticsearch Open Distro"
description: |-
  Provides an Elasticsearch Open Distro ISM policy. Please refer to the Open Distro [ISM documentation](https://opendistro.github.io/for-elasticsearch-docs/docs/ism/) for details.
---

# Resource `elasticsearch_opendistro_ism_policy_mapping`

Provides an Elasticsearch Open Distro ISM policy. Please refer to the Open Distro [ISM documentation](https://opendistro.github.io/for-elasticsearch-docs/docs/ism/) for details.

## Example Usage

```terraform
resource "elasticsearch_opendistro_ism_policy_mapping" "test" {
	policy_id = "policy_1"
  state = "delete"
  include" = {
    { "state": "searches" }
  }
}
```

## Schema

### Required

- **indexes** (String) Name of the index to apply the policy to.
- **policy_id** (String) The name of the policy.

### Optional

- **id** (String) The ID of this resource.
- **include** (Set of Map of String)
- **is_safe** (Boolean)
- **managed_indexes** (Set of String)
- **state** (String)


