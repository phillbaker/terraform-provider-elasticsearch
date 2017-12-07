# terraform-provider-elasticsearch

[![Build Status](https://travis-ci.org/phillbaker/terraform-provider-elasticsearch.svg?branch=master)](https://travis-ci.org/phillbaker/terraform-provider-elasticsearch)

This is a terraform provider that lets you provision elasticsearch resources, compatible with v5 of elasticsearch. Based off of an [original PR to Terraform](https://github.com/hashicorp/terraform/pull/13238).

## Installation

[Download a binary](https://github.com/phillbaker/terraform-provider-elasticsearch/releases), and put it in a good spot on your system. Then update your `~/.terraformrc` to refer to the binary:

```hcl
providers {
  elasticsearch = "/path/to/terraform-provider-elasticsearch"
}
```

See [the docs for more information](https://www.terraform.io/docs/plugins/basics.html).

## Usage

```tf
provider "elasticsearch" {
    url = "https://search-foo-bar-pqrhr4w3u4dzervg41frow4mmy.us-east-1.es.amazonaws.com" # Don't include port at the end for aws
    aws_access_key = ""
    aws_secret_key = ""
    aws_token = "" # if necessary
    insecure = true # to bypass certificate check
    cacert_file = "/path/to/ca.crt" # when connecting to elastic with self-signed certificate
}

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

resource "elasticsearch_kibana_object" "test_visualization" {
  body = <<EOF
[
  {
    "_id": "response-time-percentile",
    "_type": "visualization",
    "_source": {
      "title": "Total response time percentiles",
      "visState": "{\"title\":\"Total response time percentiles\",\"type\":\"line\",\"params\":{\"addTooltip\":true,\"addLegend\":true,\"legendPosition\":\"right\",\"showCircles\":true,\"interpolate\":\"linear\",\"scale\":\"linear\",\"drawLinesBetweenPoints\":true,\"radiusRatio\":9,\"times\":[],\"addTimeMarker\":false,\"defaultYExtents\":false,\"setYExtents\":false},\"aggs\":[{\"id\":\"1\",\"enabled\":true,\"type\":\"percentiles\",\"schema\":\"metric\",\"params\":{\"field\":\"app.total_time\",\"percents\":[50,90,95]}},{\"id\":\"2\",\"enabled\":true,\"type\":\"date_histogram\",\"schema\":\"segment\",\"params\":{\"field\":\"@timestamp\",\"interval\":\"auto\",\"customInterval\":\"2h\",\"min_doc_count\":1,\"extended_bounds\":{}}},{\"id\":\"3\",\"enabled\":true,\"type\":\"terms\",\"schema\":\"group\",\"params\":{\"field\":\"system.syslog.program\",\"size\":5,\"order\":\"desc\",\"orderBy\":\"_term\"}}],\"listeners\":{}}",
      "uiStateJSON": "{}",
      "description": "",
      "version": 1,
      "kibanaSavedObjectMeta": {
        "searchSourceJSON": "{\"index\":\"filebeat-*\",\"query\":{\"query_string\":{\"query\":\"*\",\"analyze_wildcard\":true}},\"filter\":[]}"
      }
    }
  }
]
EOF
}
```

## Development

### Requirements

* [Golang](https://golang.org/dl/) >= 1.7
* [Glide](https://github.com/Masterminds/glide)


```
# Ensure that this folder is at the following location: `${GOPATH}/src/github.com/phillbaker/terraform-provider-elasticsearch`
cd $GOPATH/src/github.com/phillbaker/terraform-provider-elasticsearch

glide install
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
