---
layout: "elasticsearch"
page_title: "Elasticsearch: elasticsearch_opendistro_kibana_tenant"
subcategory: "Elasticsearch Open Distro"
description: |-
  Provides an Elasticsearch Open Distro Kibana tenant resource.
---

# elasticsearch_opendistro_kibana_tenant

Provides an Elasticsearch Open Distro Kibana tenant resource.
Please refer to the Open Distro [documentation][1] for details.

## Example Usage

```hcl
# Create a tenant
resource "elasticsearch_opendistro_kibana_tenant" "test" {
  tenant_name   = "test"
  description   = "test tenant"
}
```

## Argument Reference

The following arguments are supported:

* `tenant_name` -
    (Required) The name of the tenant.
* `description` -
    (Optional) Description of the tenant.

## Attributes Reference

The following attributes are exported:

* `id` -
    The name of the tenant.

## Import

Elasticsearch Open Distro tenant can be imported using the `tenant_name`, e.g.

```sh
$ terraform import elasticsearch_opendistro_kibana_tenant.writer test
```

<!-- External links -->
[1]: https://opendistro.github.io/for-elasticsearch-docs/docs/security/access-control/multi-tenancy/
