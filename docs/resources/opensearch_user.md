---
layout: "elasticsearch"
page_title: "Elasticsearch: elasticsearch_opensearch_user"
subcategory: "OpenSearch"
description: |-
  Provides an Elasticsearch OpenSearch security user.
---

# elasticsearch_opensearch_user

Provides an Elasticsearch OpenSearch security user. Please refer to the OpenSearch [Access Control documentation][1] for details.

## Example Usage

```hcl
# Create a user
resource "elasticsearch_opensearch_user" "mapper" {
  username    = "app-reader"
  password    = "supersekret123!"
  description = "a reader role for our app"
}
```

And a full user, role and role mapping example:

```hcl
resource "elasticsearch_opensearch_role" "reader" {
  role_name   = "app_reader"
  description = "App Reader Role"

  index_permissions {
    index_patterns  = ["app-*"]
    allowed_actions = ["get", "read", "search"]
  }
}

resource "elasticsearch_opensearch_user" "reader" {
  username = "app-reader"
  password = var.password
}

resource "elasticsearch_opensearch_roles_mapping" "reader" {
  role_name   = elasticsearch_opensearch_role.reader.id
  description = "App Reader Role"
  users       = [elasticsearch_opensearch_user.reader.id]
}
```

## Argument Reference

The following arguments are supported:

* `username` -
    (Required) The name of the security role.
* `description` -
    (Optional) Description of the user.
* `backend_roles` -
    (Optional) A list of backend roles.
* `password` -
    (Optional) The plain text password for the user, cannot be specified with `password_hash`.
* `password_hash` -
    (Optional) The pre-hashed password for the user, cannot be specified with `password`.
* `attributes` -
    (Optional) A map of arbitrary key value string pairs stored alongside of users.

## Attributes Reference

The following attributes are exported:

* `id` -
    The name of the security user.

## Import

Elasticsearch OpenSearch user can be imported using the `username`, e.g.

```sh
$ terraform import elasticsearch_opensearch_user.reader app_reader
```

<!-- External links -->
[1]: https://opendistro.github.io/for-elasticsearch-docs/docs/security/access-control/
