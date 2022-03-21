---
layout: "elasticsearch"
page_title: "Elasticsearch: elasticsearch_xpack_watch"
subcategory: "Elasticsearch Xpack"
description: |-
  Provides an Elasticsearch xpack watch resource.
---

# elasticsearch_xpack_watch

Provides an Elasticsearch xpack watch resource.

## Example Usage

```tf
# Create an xpack watch
resource "elasticsearch_xpack_watch" "watch_1" {
  watch_id = "watch_1"
  active = true
  body = <<EOF
{
  "trigger": {
    "schedule": {
      "interval": "10m"
    }
  },
  "input" : {
    "search" : {
      "request" : {
        "indices" : [
          "filebeat*"
        ],
        "body" : {
          "query" : {
            "bool" : {
              "must" : {
                "match": {
                   "http.response": 500
                }
              },
              "filter" : {
                "range": {
                  "@timestamp": {
                    "from": "{{ctx.trigger.scheduled_time}}||-10m",
                    "to": "{{ctx.trigger.triggered_time}}"
                  }
                }
              }
            }
          }
        }
      }
    }
  },
  "condition" : {
    "compare" : { "ctx.payload.hits.total" : { "gt" : 100 }}
  },
  "actions" : {
    "email_ops" : {
      "email" : {
        "to" : "ops@example.com",
        "subject" : "High 500s detected"
      }
    }
  },
  "metadata": {
    "xpack": {
      "type": "json"
    },
    "name": "http error 500 warning"
  }
}
EOF
}
```

Note: Watches using basic authentication should define a basic authorization header as part of `headers` json block, rather than using the watch basic auth stanza. 
With the watch basic auth stanza, the value of the `password` field return by the get watch api will be `::es_redacted::`, not the plain text password. This will cause the provider to continuously re-apply watches as the passwords do not match.


## Argument Reference

The following arguments are supported:

* `watch_id` - (Required) The name of the xpack watch.
* `body` - (Required) The JSON body of the xpack watch.
* `active` - (Optional) Boolean to activate the xpack watcher, defaults `true`

## Attributes Reference

The following attributes are exported:

* `id` - The name of the xpack watch.
