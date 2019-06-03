# terraform-provider-elasticsearch

[![Build Status](https://travis-ci.org/phillbaker/terraform-provider-elasticsearch.svg?branch=master)](https://travis-ci.org/phillbaker/terraform-provider-elasticsearch)

This is a terraform provider that lets you provision elasticsearch resources, compatible with v5, v6 and v7 of elasticsearch. Based off of an [original PR to Terraform](https://github.com/hashicorp/terraform/pull/13238).

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

resource "elasticsearch_kibana_object" "test_visualization_v5" {
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

resource "elasticsearch_kibana_object" "test_visualization_v6" {
  body = <<EOF
[
  {
    "_id": "visualization:response-time-percentile",
    "_type": "doc",
    "_source": {
      "type": "visualization",
      "visualization": {
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
  }
]
EOF
}
```

### For use with AWS Elasticsearch domains

The Elasticsearch provider is flexible in the means of providing credentials for authentication with AWS Elasticsearch domains. The following methods are supported, in this order, and explained below:

- Static credentials
- Environment variables
- Shared credentials file

#### Static credentials

Static credentials can be provided by adding an `aws_access_key` and `aws_secret_key` in-line in the Elasticsearch provider block. If applicable, you may also specify a `aws_token` value.

Example usage:

```tf
provider "elasticsearch" {
    url = "https://search-foo-bar-pqrhr4w3u4dzervg41frow4mmy.us-east-1.es.amazonaws.com"
    aws_access_key = "anaccesskey"
    aws_secret_key = "asecretkey"
    aws_token = "" # if necessary
}
```

#### Environment variables

You can provide your credentials via the `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY`, environment variables, representing your AWS Access Key and AWS Secret Key. If applicable, the `AWS_SESSION_TOKEN` environment variables is also supported.

Example usage:

```shell
$ export AWS_ACCESS_KEY_ID="anaccesskey"
$ export AWS_SECRET_ACCESS_KEY="asecretkey"
$ terraform plan
```

#### Shared Credentials file

You can use an AWS credentials file to specify your credentials. The default location is `$HOME/.aws/credentials` on Linux and macOS, or `%USERPROFILE%\.aws\credentials` for Windows users.

Please refer to the official [userguide](https://docs.aws.amazon.com/cli/latest/userguide/cli-config-files.html) for instructions on how to create the credentials file.

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
