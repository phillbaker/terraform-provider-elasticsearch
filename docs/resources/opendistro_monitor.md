---
layout: "elasticsearch"
page_title: "Elasticsearch: elasticsearch_opendistro_monitor"
subcategory: "Elasticsearch Open Distro"
description: |-
  Provides an Elasticsearch Open Distro monitor.
---

# elasticsearch_opendistro_monitor

Provides an Elasticsearch Open Distro monitor.
Please refer to the Open Distro [monitor documentation][1] for details.

## Example Usage

```hcl
# Create an monitor
resource "elasticsearch_opendistro_monitor" "movies_last_hour" {
  body = <<EOF
{
  "name": "test-monitor",
  "type": "monitor",
  "enabled": true,
  "schedule": {
    "period": {
      "interval": 1,
      "unit": "MINUTES"
    }
  },
  "inputs": [{
    "search": {
      "indices": ["movies"],
      "query": {
        "size": 0,
        "aggregations": {},
        "query": {
          "bool": {
            "adjust_pure_negative":true,
            "boost":1,
            "filter": [{
              "range": {
                "@timestamp": {
                  "boost":1,
                  "from":"||-1h",
                  "to":"",
                  "include_lower":true,
                  "include_upper":true,
                  "format": "epoch_millis"
                }
              }
            }]
          }
        }
      }
    }
  }],
  "triggers": [
    {
      "name" : "Errors",
      "severity" : "1",
      "condition" : {
        "script" : {
          "source" : "ctx.results[0].hits.total.value > 0",
          "lang" : "painless"
        }
      },
      "actions" : [
        {
          "name" : "Slack",
          "destination_id" : "${elasticsearch_opendistro_destination.slack_on_call_channel.id}",
          "message_template" : {
            "source" : "bogus",
            "lang" : "mustache"
          },
          "throttle_enabled" : false,
          "subject_template" : {
            "source" : "Production Errors",
            "lang" : "mustache"
          }
        }
      ]
    }
  ]
}
EOF
}
```

## Argument Reference

The following arguments are supported:

* `body` -
    (Required) The policy document.

## Attributes Reference

The following attributes are exported:

* `id` -
    The id of the monitor.

## Import

Elasticsearch Open Distro monitor can be imported using the `id`, e.g.

```
$ terraform import elasticsearch_opendistro_monitor.alert lgOZb3UB96pyyRQv0ppQ
```

<!-- External links -->
[1]: https://opendistro.github.io/for-elasticsearch-docs/docs/alerting/monitors/
