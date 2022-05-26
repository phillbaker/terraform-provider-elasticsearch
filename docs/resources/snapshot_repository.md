---
layout: "elasticsearch"
page_title: "Elasticsearch: elasticsearch_snapshot_repository"
subcategory: "Elasticsearch Opensource"
description: |-
  Provides an Elasticsearch snapshot repository resource.
---

# elasticsearch_snapshot_repository

Provides an Elasticsearch snapshot repository resource.

## Example Usage

```hcl
# Create a snapshot repository
resource "elasticsearch_snapshot_repository" "repo" {
  name = "es-index-backups"
  type = "s3"
  settings = {
    bucket   = "es-index-backups"
    region   = "us-east-1"
    role_arn = "arn:aws:iam::123456789012:role/MyElasticsearchRole"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the repository.
* `type` - (Required) The name of the repository backend (required plugins must be installed).
* `settings` - (Optional) The settings map applicable for the backend (documented [here](https://www.elastic.co/guide/en/elasticsearch/reference/current/modules-snapshots.html) for official plugins).

## Attributes Reference

The following attributes are exported:

* `id` - The name of the snapshot repository.

## Import

Snapshot repositories can be imported using the `name`, e.g.

```sh
$ terraform import elasticsearch_snapshot_repository.repo es-index-backups
```
