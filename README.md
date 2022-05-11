# terraform-provider-elasticsearch

![Test](https://github.com/phillbaker/terraform-provider-elasticsearch/workflows/Test/badge.svg?branch=master)

This is a terraform provider that lets you provision Elasticsearch and Opensearch resources, compatible with v6 and v7 of Elasticsearch and v1 of Opensearch. Based off of an [original PR to Terraform](https://github.com/hashicorp/terraform/pull/13238).

## Using the Provider

### Terraform 0.13 and above

[This package is published on the official Terraform registry](https://registry.terraform.io/providers/phillbaker/elasticsearch/latest). Note, we currently test against the 1.x branch of Terraform - this should continue to work with >= 0.13 versions, however, compatibility is not tested in the >= 2.x version of this provider.

### Terraform 0.12 or manual installation

[Or download a binary](https://github.com/phillbaker/terraform-provider-elasticsearch/releases), and put it in a good spot on your system. Then update your `~/.terraformrc` to refer to the binary:

```hcl
providers {
  elasticsearch = "/path/to/terraform-provider-elasticsearch"
}
```

See [the docs for more on manual installation](https://www.terraform.io/docs/extend/how-terraform-works.html#plugin-locations).

### Terraform 0.11

With version 2.x of this provider, it uses version 2.x of the Terraform Plugin SDK which only supports Terraform 0.12 and higher. Please see the 1.x releases of this provider for Terraform 0.11 support.

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

Examples of resources can be found in the examples directory. The resources currently supported from the: opensource Elasticsearch, XPack and OpenDistro/OpenSearch distributions are described below.

#### Elasticsearch

- [x] [Index](https://www.elastic.co/guide/en/elasticsearch/reference/current/indices.html)
- [x] [Index template](https://www.elastic.co/guide/en/elasticsearch/reference/current/indices.html)
- [x] [Ingest pipeline](https://www.elastic.co/guide/en/elasticsearch/reference/current/ingest-apis.html)
- [x] [Snapshot repository](https://www.elastic.co/guide/en/elasticsearch/reference/current/snapshot-restore-apis.html)
- [ ] [Search template](https://www.elastic.co/guide/en/elasticsearch/reference/current/search-template.html)

#### Kibana

- [x] Kibana Object
  - [ ] Visualization
  - [ ] Search
  - [ ] Dashboard
- [x] Kibana Alerts

#### XPack

- [ ] [Cross cluster replication](https://www.elastic.co/guide/en/elasticsearch/reference/current/ccr-apis.html)
- [ ] [Enrich policies](https://www.elastic.co/guide/en/elasticsearch/reference/current/enrich-apis.html)
- [x] [Index lifecycle management](https://www.elastic.co/guide/en/elasticsearch/reference/current/index-lifecycle-management-api.html)
- [x] [License management](https://www.elastic.co/guide/en/elasticsearch/reference/current/licensing-apis.html)
- [ ] [Rollup jobs](https://www.elastic.co/guide/en/elasticsearch/reference/current/rollup-apis.html)
- [x] [Security](https://www.elastic.co/guide/en/elasticsearch/reference/current/security-api.html) (Role/Role Mapping/User)
- [x] [Snapshot lifecycle policy](https://www.elastic.co/guide/en/elasticsearch/reference/current/snapshot-lifecycle-management-api.html)
- [x] [Watch](https://www.elastic.co/guide/en/elasticsearch/reference/current/watcher-api.html)

#### OpenDistro/OpenSearch

- [x] [Alerting](https://opendistro.github.io/for-elasticsearch-docs/docs/alerting/api/) (Destinations/Monitors)
- [x] [Security](https://opendistro.github.io/for-elasticsearch-docs/docs/security/access-control/api/) (Role/Role Mapping/User)
- [x] [Index State Management](https://opendistro.github.io/for-elasticsearch-docs/docs/ism/api/)
- [x] [Kibana Tenant](https://opendistro.github.io/for-elasticsearch-docs/docs/security/access-control/multi-tenancy/)
- [ ] [Anomaly Detection](https://opendistro.github.io/for-elasticsearch-docs/docs/ad/api/)

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

```hcl
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

### For use with AWS Opensearch domains

Please see [the documentation](./docs/index.md#AWS-authentication) for details.

## Development

### Requirements

* [Golang](https://golang.org/dl/) >= 1.13


```sh
go build -o /path/to/binary/terraform-provider-elasticsearch
```

### Running tests locally

Start an instance of ElasticSearch locally with the following:

```sh
./script/install-tools
export OSS_IMAGE="opensearchproject/opensearch:1.2.0"
export ES_OPENDISTRO_IMAGE="opensearchproject/opensearch:1.2.0"
export ES_COMMAND=""
export ES_KIBANA_IMAGE=""
export OPENSEARCH_PREFIX="plugins.security"
export OSS_ENV_VAR="plugins.security.disabled=true"
export XPACK_IMAGE="docker.elastic.co/elasticsearch/elasticsearch:7.10.1"
docker-compose up -d
docker-compose ps -a
```

When running tests, ensure that your test/debug profile has environmental variables as below:

- `ELASTICSEARCH_URL=http://localhost:9200_`
- `TF_ACC=1`

### Running terraform with a local provider

Build the executable, and start in debug mode:

```console
$ go build
$ ./terraform-provider-elasticsearch -debuggable
{"@level":"debug","@message":"plugin address","@timestamp":"2022-05-17T10:10:04.331668+01:00","address":"/var/folders/32/3mbbgs9x0r5bf991ltrl3p280000gs/T/plugin1346340234","network":"unix"}
Provider started, to attach Terraform set the TF_REATTACH_PROVIDERS env var:

        TF_REATTACH_PROVIDERS='{"registry.terraform.io/phillbaker/elasticsearch":{"Protocol":"grpc","ProtocolVersion":5,"Pid":79075,"Test":true,"Addr":{"Network":"unix","String":"/var/folders/32/3mbbgs9x0r5bf991ltrl3p280000gs/T/plugin1346340234"}}}'
```

In another terminal, you can test your terraform code:

```console
$ cd <my-project/terraform>
$ export TF_REATTACH_PROVIDERS=<env var above>
$ terraform apply
```

The local provider will be used instead, and you should see debug information printed to the terminal.

## Licence

See LICENSE.

## Contributing

1. Fork it ( https://github.com/phillbaker/terraform-provider-elasticsearch/fork )
2. Create your feature branch (`git checkout -b my-new-feature`)
3. Commit your changes (`git commit -am 'Add some feature'`)
4. Push to the branch (`git push origin my-new-feature`)
5. Create a new Pull Request
