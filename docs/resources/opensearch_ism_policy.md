---
layout: "elasticsearch"
page_title: "Elasticsearch: elasticsearch_opensearch_ism_policy"
subcategory: "OpenSearch"
description: |-
  Provides an Elasticsearch Open Distro ISM policy.
---

# elasticsearch_opensearch_ism_policy

Provides an OpenSearch ISM policy.
Please refer to the Open Distro [ISM documentation][1] for details.

## Example Usage

```hcl
# Create an ISM policy
resource "elasticsearch_opensearch_ism_policy" "cleanup" {
  policy_id = "delete_after_15d"
  body      = file("${path.module}/policies/delete_after_15d.json")
}
```

## Argument Reference

The following arguments are supported:

* `policy_id` -
    (Required) The id of the ISM policy.
* `body` -
    (Required) The policy document.

## Attributes Reference

The following attributes are exported:

* `id` -
    The id of the ISM policy.
* `primary_term` -
    The primary term of the ISM policy version.
* `seq_no` -
    The sequence number of the ISM policy version.

## Import

Elasticsearch Open Distro ISM policy can be imported using the `policy_id`, e.g.

```sh
$ terraform import elasticsearch_opensearch_ism_policy.cleanup delete_after_15d
```

<!-- External links -->
[1]: https://opendistro.github.io/for-elasticsearch-docs/docs/ism/
