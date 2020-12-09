# terraform-provider-elasticsearch

![Test](https://github.com/phillbaker/terraform-provider-elasticsearch/workflows/Test/badge.svg?branch=master)

This is a terraform provider that lets you provision elasticsearch resources, compatible with v5, v6 and v7 of elasticsearch. Based off of an [original PR to Terraform](https://github.com/hashicorp/terraform/pull/13238).

## Installation

[This package is published on the official Terraform registry](https://registry.terraform.io/providers/phillbaker/elasticsearch/latest).

[Or download a binary](https://github.com/phillbaker/terraform-provider-elasticsearch/releases), and put it in a good spot on your system. Then update your `~/.terraformrc` to refer to the binary:

```hcl
providers {
  elasticsearch = "/path/to/terraform-provider-elasticsearch"
}
```

See [the docs for more on manual installation](https://www.terraform.io/docs/plugins/basics.html).

## Usage

```tf
provider "elasticsearch" {
    url = "https://search-foo-bar-pqrhr4w3u4dzervg41frow4mmy.us-east-1.es.amazonaws.com" # Don't include port at the end for aws
    aws_access_key = ""
    aws_secret_key = ""
    aws_token = "" # if necessary
    insecure = true # to bypass certificate check
    cacert_file = "/path/to/ca.crt" # when connecting to elastic with self-signed certificate
    sign_aws_requests = true # only needs to be true if your domain access policy includes IAM users or roles
}
```

### API Coverage

Examples of resources can be found in the examples directory. The resources currently supported from the: opensource Elasticsearch, XPack and OpenDistro distributions are described below.

#### Elasticsearch

- [x] [Index](https://www.elastic.co/guide/en/elasticsearch/reference/current/indices.html)
- [x] [Index template](https://www.elastic.co/guide/en/elasticsearch/reference/current/indices.html)
- [x] [Ingest pipeline](https://www.elastic.co/guide/en/elasticsearch/reference/current/ingest-apis.html)
- [x] [Snapshot repository](https://www.elastic.co/guide/en/elasticsearch/reference/current/snapshot-restore-apis.html)

#### Kibana

- [x] Kibana Object
  - [ ] Visualization
  - [ ] Search
  - [ ] Dashboard

#### XPack

- [ ] [Cross cluster replication](https://www.elastic.co/guide/en/elasticsearch/reference/current/ccr-apis.html)
- [ ] [Enrich policies](https://www.elastic.co/guide/en/elasticsearch/reference/current/enrich-apis.html)
- [x] [Index lifecycle management](https://www.elastic.co/guide/en/elasticsearch/reference/current/index-lifecycle-management-api.html)
- [ ] [License management](https://www.elastic.co/guide/en/elasticsearch/reference/current/licensing-apis.html)
- [ ] [Rollup jobs](https://www.elastic.co/guide/en/elasticsearch/reference/current/rollup-apis.html)
- [x] [Security](https://www.elastic.co/guide/en/elasticsearch/reference/current/security-api.html) (Role/Role Mapping/User)
- [x] [Snapshot lifecycle policy](https://www.elastic.co/guide/en/elasticsearch/reference/current/snapshot-lifecycle-management-api.html)
- [x] [Watch](https://www.elastic.co/guide/en/elasticsearch/reference/current/watcher-api.html)

#### OpenDistro

- [x] [Alerting](https://opendistro.github.io/for-elasticsearch-docs/docs/alerting/api/) (Destinations/Monitors)
- [x] [Security](https://opendistro.github.io/for-elasticsearch-docs/docs/security/access-control/api/) (Role/Role Mapping/User)
- [x] [Index State Management](https://opendistro.github.io/for-elasticsearch-docs/docs/ism/api/)

### Examples

```tf
resource "elasticsearch_index_template" "test" {
  name = "terraform-test"
  body = <<EOF
{
  "template": "logstash-*",
  "version": 50001,
  "settings": {
    "index.refresh_interval": "5s"
  },
  "mappings": {
    "_default_": {
      "_all": {"enabled": true, "norms": false},
      "dynamic_templates": [ {
        "message_field": {
          "path_match": "message",
          "match_mapping_type": "string",
          "mapping": {
            "type": "text",
            "norms": false
          }
        }
      }, {
        "string_fields": {
          "match": "*",
          "match_mapping_type": "string",
          "mapping": {
            "type": "text", "norms": false,
            "fields": {
              "keyword": { "type": "keyword" }
            }
          }
        }
      } ],
      "properties": {
        "@timestamp": { "type": "date", "include_in_all": false },
        "@version": { "type": "keyword", "include_in_all": false },
        "geoip" : {
          "dynamic": true,
          "properties": {
            "ip": { "type": "ip" },
            "location": { "type": "geo_point" },
            "latitude": { "type": "half_float" },
            "longitude": { "type": "half_float" }
          }
        }
      }
    }
  }
}
EOF
}

# A saved search, visualization or dashboard
resource "elasticsearch_kibana_object" "test_dashboard" {
  body = "${file("dashboard_path.txt")}"
}
```

Example watches (target notification actions must be setup manually before hand)

```
# Monitor cluster status with auth being required
resource "elasticsearch_xpack_watch" "cluster-status-red" {
  watch_id = "cluster-status-red"
  body = <<EOF
{
  "trigger": {
    "schedule": {
      "interval": "1m"
    }
  },
  "input": {
    "http": {
      "request": {
        "scheme": "http",
        "host": "localhost",
        "port": 9200,
        "method": "get",
        "path": "/_cluster/health",
        "params": {},
        "headers": {
          "Authorization": "Basic ${base64encode('username:password')}"
        }
      }
    }
  },
  "condition": {
    "compare": {
      "ctx.payload.status": {
        "eq": "red"
      }
    }
  },
  "actions": {
    "notify-slack": {
      "throttle_period_in_millis": 300000,
      "slack": {
        "account": "monitoring",
        "message": {
          "from": "watcher",
          "to": [
            "#my-slack-channel"
          ],
          "text": "Elasticsearch Monitoring",
          "attachments": [
            {
              "color": "danger",
              "title": "Cluster Health Warning - RED",
              "text": "elasticsearch cluster health is RED"
            }
          ]
        }
      }
    }
  },
  "metadata": {
    "xpack": {
      "type": "json"
    },
    "name": "Cluster Health Red"
  }
}
EOF
}

# Monitor JVM memory usage without auth required
resource "elasticsearch_xpack_watch" "jvm-memory-usage" {
  watch_id = "jvm-memory-usage"
  body = <<EOF
{
  "trigger": {
    "schedule": {
      "interval": "10m"
    }
  },
  "input": {
    "http": {
      "request": {
        "scheme": "http",
        "host": "localhost",
        "port": 9200,
        "method": "get",
        "path": "/_nodes/stats/jvm",
        "params": {
                  "filter_path": "nodes.*.jvm.mem.heap_used_percent"
                },
        "headers": {}
      }
    }
  },
  "condition": {
    "script": {
      "lang": "painless",
      "source": "ctx.payload.nodes.values().stream().anyMatch(node -> node.jvm.mem.heap_used_percent > 75)"
    }
  },
  "actions": {
    "notify-slack": {
      "throttle_period_in_millis": 600000,
      "slack": {
        "account": "monitoring",
        "message": {
          "from": "watcher",
          "to": [
            "#my-slack-channel"
          ],
          "text": "Elasticsearch Monitoring",
          "attachments": [
            {
              "color": "danger",
              "title": "JVM Memory Pressure Warning",
              "text": "JVM Memory Pressure has been > 75% on one or more nodes for the last 5 minutes."
            }
          ]
        }
      }
    }
  },
  "metadata": {
    "xpack": {
      "type": "json"
    },
    "name": "JVM Memory Pressure Warning"
  }
}
EOF
}
```

### For use with AWS Elasticsearch domains

Please see [the documentation](./docs/index.md#AWS-authentication) for details.

## Development

### Requirements

* [Golang](https://golang.org/dl/) >= 1.11


```
go build -o /path/to/binary/terraform-provider-elasticsearch
```

## Licence

See LICENSE.

## Contributing

1. Fork it ( https://github.com/phillbaker/terraform-provider-elasticsearch/fork )
2. Create your feature branch (`git checkout -b my-new-feature`)
3. Commit your changes (`git commit -am 'Add some feature'`)
4. Push to the branch (`git push origin my-new-feature`)
5. Create a new Pull Request
