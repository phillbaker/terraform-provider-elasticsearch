---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "elasticsearch_xpack_role_mapping Resource - terraform-provider-elasticsearch"
subcategory: "Elasticsearch Xpack"
description: |-
  Provides an Elasticsearch XPack role mapping resource. Role mappings define which roles are assigned to each user. Each mapping has rules that identify users and a list of roles that are granted to those users. See the upstream docs https://www.elastic.co/guide/en/elasticsearch/reference/current/security-api.html for more details.
---

# elasticsearch_xpack_role_mapping (Resource)

Provides an Elasticsearch XPack role mapping resource. Role mappings define which roles are assigned to each user. Each mapping has rules that identify users and a list of roles that are granted to those users. See the upstream [docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/security-api.html) for more details.

## Example Usage

```terraform
resource "elasticsearch_xpack_role_mapping" "test" {
  role_mapping_name = "test"
  roles = [
    "admin",
    "user",
  ]
  enabled = true
  rules   = <<-EOF
  {
    "any": [
      {
        "field": {
          "username": "esadmin"
        }
      },
      {
        "field": {
          "groups": "cn=admins,dc=example,dc=com"
        }
      }
    ]
  }
  EOF
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- **role_mapping_name** (String) The distinct name that identifies the role mapping, used solely as an identifier.
- **roles** (Set of String) A list of role names that are granted to the users that match the role mapping rules.
- **rules** (String) A list of mustache templates that will be evaluated to determine the roles names that should granted to the users that match the role mapping rules. This matches fields of users, rules can be grouped into `all` and `any` top level keys.

### Optional

- **enabled** (Boolean) Mappings that have `enabled` set to `false` are ignored when role mapping is performed.
- **metadata** (String) Additional metadata that helps define which roles are assigned to each user. Keys beginning with `_` are reserved for system usage.

### Read-Only

- **id** (String) The ID of this resource.

## Import

Import is supported using the following syntax:

```shell
# Import by name
terraform import elasticsearch_xpack_role_mapping.test test
```
