---
layout: "elasticsearch"
page_title: "Elasticsearch: elasticsearch_opensearch_role"
subcategory: "OpenSearch"
description: |-
  Provides an Elasticsearch OpenSearch security role resource.
---

# elasticsearch_opensearch_role

Provides an Elasticsearch OpenSearch security role resource.
Please refer to the OpenSearch [Access Control documentation][1] for details.

## Example Usage

```hcl
# Create a role
resource "elasticsearch_opensearch_role" "writer" {
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

To set document level permissions:

```hcl
resource "elasticsearch_opensearch_role" "writer" {
  role_name = "foo_writer"

  cluster_permissions = ["*"]

  index_permissions {
    index_patterns          = ["pub*"]
    allowed_actions         = ["read"]
    document_level_security = "{\"term\": { \"readable_by\": \"$${user.name}\"}}"
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
* `document_level_security` -
    (Optional) A selector for [document-level security][2] (json formatted using jsonencode).
* `field_level_security` -
    (Optional) A list of selectors for [field-level security][3].
* `masked_fields` -
    (Optional) A list of [masked fields][4].
* `allowed_actions` -
    (Optional) A list of allowed actions.

The `tenant_permissions` object supports the following:

* `tenant_patterns` -
    (Optional) A list of glob patterns for the [tenant][5] names.
* `allowed_actions` -
    (Optional) A list of allowed actions.

## Attributes Reference

The following attributes are exported:

* `id` -
    The name of the security role.

## Import

Elasticsearch OpenSearch security role can be imported using the `role_name`, e.g.

```sh
$ terraform import elasticsearch_opensearch_role.writer logs_writer
```

<!-- External links -->
[1]: https://opendistro.github.io/for-elasticsearch-docs/docs/security/access-control/
[1]: https://opendistro.github.io/for-elasticsearch-docs/docs/security/access-control/document-level-security/
[3]: https://opendistro.github.io/for-elasticsearch-docs/docs/security/access-control/field-level-security/
[4]: https://opendistro.github.io/for-elasticsearch-docs/docs/security/access-control/field-masking/
[5]: https://opendistro.github.io/for-elasticsearch-docs/docs/security/access-control/multi-tenancy/
