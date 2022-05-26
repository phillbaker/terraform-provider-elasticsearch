---
layout: "elasticsearch"
page_title: "Elasticsearch: elasticsearch_xpack_role"
subcategory: "Elasticsearch Xpack"
description: |-
  Provides an Elasticsearch XPack role resource.
---

# elasticsearch_xpack_role

Provides an Elasticsearch XPack role resource. See the upstream [docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/security-api-put-role.html) for more details.

## Example Usage

```tf
# Create an xpack role
resource "elasticsearch_xpack_role" "test" {
  role_name = "tes"
  indices {
    names 	   = ["testIndice"]
    privileges = ["read"]
    field_security {
      grant = ["testField", "testField2"]
    }
  }
  indices {
    names 	   = ["testIndice2"]
    privileges = ["write"]
    field_security {
      grant  = ["*"]
      except = ["testField3"]
    }
  }
  cluster = [
    "all"
  ]
  applications {
    application = "testapp"
    privileges  = [
      "write",
      "read"
    ]
    resources   = [
      "*"
    ]
  }
}
```


## Argument Reference

The following arguments are supported:

* `role_name` - (Required) The name of the xpack role.
* `indices` - (Optional) A configuration of index objects (see below).
* `applications` - (Optional) A configuration of application objects (see below).
* `global` - (Optional) A JSON string of an object defining global privileges. A global privilege is a form of cluster privilege that is request-aware.
* `run_as` - (Optional) A list of users that the owners of this role can impersonate
* `metadata` - (Optional) A JSON string of arbitrary key value pairs, keys cannot start with `_`.


The `indices` object supports the following:

* `names` - (Required) A list of index names.
* `privileges` - (Required) The index level privileges that the owners of the role have on the specified indices.
* `query` - (Optional) A search query that defines the documents the owners of the role have read access to. A document within the specified indices must match this query in order for it to be accessible by the owners of the role.
* `field_security` - (Optional) A configuration of field security objects (see below). The absence of field_security in a role is equivalent to * access.


The `field_security` object supports the following:

* `grant` - (Optional) Specifies the fields that a role can access as part of the indices permissions. Wildcards are supported.
* `except` - (Optional) Specify denied fields for a role. The denied fields must be a subset of the fields to which permissions were granted. Defining denied and granted fields implies access to all granted fields except those which match the pattern in the denied fields.


The `applications` object supports the following:

* `application` - (Required) The name of the application to which this entry applies
* `privileges` - (Optional) A list of strings, where each element is the name of an application privilege.
* `resources` - (Optional) A list resources to which the privileges are applied


## Attributes Reference

The following attributes are exported:

* `id` - The name of the xpack role.

## Import

XPack roles can be imported using the `role_name`, e.g.

```sh
$ terraform import elasticsearch_xpack_role.test tes
```
