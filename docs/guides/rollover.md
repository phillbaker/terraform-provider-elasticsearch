---
page_title: "Manage indices using rollover and templates"
---

When indices are created using `elasticsearch_index_template` and configured to rollover with Index Lifecycle Management (ILM, part of Xpack) or Index State Management (ISM, part of OpenDistro), this provider can follow the most recent index so that the underlying index is still managed correctly.

This works by:
1. Create an index template to link the ISM or ILM policy with the indices. Then when creating the first index with `elasticsearch_index`, the provider inspects the response for an ILM/ISM lifecycle policy.
1. If the ILM/ISM lifecycle policy contains the setting `rollover_alias`, the provider sets this field in the Terraform state.
1. On the next Terraform run: if the `rollover_alias` is present then it uses this value to get the current `is_write_index`.

For example, using ILM:

1. Create the following `main.tf` file:

```hcl
provider "elasticsearch" {
  url = "http://localhost:9200"
}

resource "elasticsearch_xpack_index_lifecycle_policy" "test" {
  name = "test"
  body = <<EOF
{
  "policy": {
    "phases": {
      "hot": {
        "min_age": "0ms",
        "actions": {
          "rollover": {
            "max_size": "50gb"
          }
        }
      }
    }
  }
}
EOF
}

resource "elasticsearch_index_template" "test" {
  name = "test"
  body = <<EOF
{
  "index_patterns": [
    "test-*"
  ],
  "settings": {
    "index": {
      "lifecycle": {
        "name": "${elasticsearch_xpack_index_lifecycle_policy.test.name}",
        "rollover_alias": "test"
      }
    }
  }
}
EOF
}

resource "elasticsearch_index" "test" {
  name               = "test-000001"
  number_of_shards   = 1
  number_of_replicas = 1
  aliases = jsonencode({
    "test" = {
      "is_write_index" = true
    }
  })

  depends_on = [elasticsearch_index_template.test]
}
```

2. Start a new Elasticsearch Container:

```sh
$ export ES_OSS_IMAGE=elasticsearch:7.9.2
$ docker-compose up -d elasticsearch
Creating network "terraform-provider-elasticsearch_default" with the default driver
Creating terraform-provider-elasticsearch_elasticsearch_1 ... done
```

3. Create the example Index Lifecycle Policy, Index Template and Index:

```sh
$ terraform apply -auto-approve
elasticsearch_xpack_index_lifecycle_policy.test: Creating...
elasticsearch_xpack_index_lifecycle_policy.test: Creation complete after 0s [id=test]
elasticsearch_index_template.test: Creating...
elasticsearch_index_template.test: Creation complete after 0s [id=test]
elasticsearch_index.test: Creating...
elasticsearch_index.test: Creation complete after 0s [id=test-000001]

Apply complete! Resources: 3 added, 0 changed, 0 destroyed.
```

4. Manually rollover the example index:

```sh
$ curl -X POST http://localhost:9200/test/_rollover
{"acknowledged":true,"shards_acknowledged":true,"old_index":"test-000001","new_index":"test-000002","rolled_over":true,"dry_run":false,"conditions":{}}
```

5. Delete the initial created index:

```sh
$ curl -X DELETE http://localhost:9200/test-000001
{"acknowledged":true}
```

6. Now all changes will be made to the index `test-000002`.
