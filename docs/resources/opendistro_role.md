---
layout: "elasticsearch"
page_title: "Elasticsearch: elasticsearch_opendistro_role"
subcategory: "Elasticsearch Opensource"
description: |-
  Provides an Elasticsearch Open Distro security role resource.
---

# elasticsearch_opendistro_role

Provides an Elasticsearch Open Distro security role resource.
Please refer to the Open Distro [Access Control documentation][1] for details.

## Example Usage

```hcl
# Create a role
resource "elasticsearch_opendistro_role" "writer" {
  role_name   = "logs_writer"
  description = "Logs writer role"

  cluster_permissions = ["*"]

  index_permissions {
    index_patterns  = ["logstash-*"]
    allowed_actions = ["write"]
  }

  tenant_permissions {
    tenant_patterns = ["logstash-*"]
    allowed_actions = ["kibana_all_write"]
  }
}
```

## Argument Reference

The following arguments are supported:

* `role_name` -
    (Required) The name of the security role.
* `description` -
    (Optional) Description of the role.
* `cluster_permissions` -
    (Optional) A list of cluster permissions.
* `index_permissions` -
    (Optional) A configuration of index permissions (documented below).
* `tenant_permissions` -
    (Optional) A configuration of tenant permissions (documented below).

The `index_permissions` object supports the following:

* `index_patterns` -
    (Optional) A list of glob patterns for the index names.
* `fls` -
    (Optional) A list of selectors for [field-level security][2].
* `masked_fields` -
    (Optional) A list of [masked fields][3].
* `allowed_actions` -
    (Optional) A list of allowed actions.

The `tenant_permissions` object supports the following:

* `tenant_patterns` -
    (Optional) A list of glob patterns for the [tenant][4] names.
* `allowed_actions` -
    (Optional) A list of allowed actions.

## Attributes Reference

The following attributes are exported:

* `id` -
    The name of the security role.

## Import

Elasticsearch Open Distro security role can be imported using the `role_name`, e.g.

```
$ terraform import elasticsearch_opendistro_role.writer logs_writer
```

<!-- External links -->
[1]: https://opendistro.github.io/for-elasticsearch-docs/docs/security/access-control/
[2]: https://opendistro.github.io/for-elasticsearch-docs/docs/security/access-control/field-level-security/
[3]: https://opendistro.github.io/for-elasticsearch-docs/docs/security/access-control/field-masking/
[4]: https://opendistro.github.io/for-elasticsearch-docs/docs/security/access-control/multi-tenancy/
