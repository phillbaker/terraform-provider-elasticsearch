---
layout: "elasticsearch"
page_title: "Elasticsearch: elasticsearch_opensearch_kibana_tenant"
subcategory: "OpenSearch"
description: |-
  Provides an Elasticsearch OpenSearch Kibana tenant resource.
---

# elasticsearch_opensearch_kibana_tenant

Provides an Elasticsearch OpenSearch Kibana tenant resource.
Please refer to the OpenSearch [documentation][1] for details.

## Example Usage

```hcl
# Create a tenant
resource "elasticsearch_opensearch_kibana_tenant" "test" {
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

Elasticsearch OpenSearch tenant can be imported using the `tenant_name`, e.g.

```sh
$ terraform import elasticsearch_opensearch_kibana_tenant.writer test
```

<!-- External links -->
[1]: https://opendistro.github.io/for-elasticsearch-docs/docs/security/access-control/multi-tenancy/
