---
layout: "elasticsearch"
page_title: "Elasticsearch: elasticsearch_opensearch_roles_mapping"
subcategory: "OpenSearch"
description: |-
  Provides an Elasticsearch OpenSearch security role mapping.
---

# elasticsearch_opensearch_roles_mapping

Provides an Elasticsearch OpenSearch security role mapping.
Please refer to the OpenSearch [Access Control documentation][1] for details.

## Example Usage

```hcl
# Create a role mapping
resource "elasticsearch_opensearch_roles_mapping" "mapper" {
  role_name     = "logs_writer"
  description   = "Mapping AWS IAM roles to ES role"
  backend_roles = [
    "arn:aws:iam::123456789012:role/lambda-call-elasticsearch",
    "arn:aws:iam::123456789012:role/run-containers",
  ]
}
```

## Argument Reference

The following arguments are supported:

* `role_name` -
    (Required) The name of the security role.
* `description` -
    (Optional) Description of the role mapping.
* `backend_roles` -
    (Optional) A list of backend roles.
* `hosts` -
    (Optional) A list of host names.
* `users` -
    (Optional) A list of users.
* `and_backend_roles` -
    (Optional) A list of backend roles.

## Attributes Reference

The following attributes are exported:

* `id` -
    The name of the security role.

## Import

Elasticsearch OpenSearch security role mapping can be imported using the `role_name`, e.g.

```sh
$ terraform import elasticsearch_opensearch_roles_mapping.mapper logs_writer
```

<!-- External links -->
[1]: https://opendistro.github.io/for-elasticsearch-docs/docs/security/access-control/
