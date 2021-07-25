---
layout: "elasticsearch"
page_title: "Elasticsearch: elasticsearch_kibana_object"
subcategory: "Elasticsearch Opensource"
description: |-
  Provides an Elasticsearch kibana object resource.
---

# elasticsearch_kibana_object

Provides an Elasticsearch kibana object resource. This resource interacts directly with the underlying Elasticsearch index backing Kibana, so the format must match what Kibana the version of Kibana is expecting. Kibana v5 and v6 will export all objects in Kibana v5 format, so the exported objects cannot be used as a source for `body` in this resource - directly pulling the JSON from a Kibana index of the same version of Elasticsearch targeted by the provider is a workaround.

With the removal of mapping types in Elasticsearch, the Kibana index changed from v5 to >= v6, previously the document mapping type had the Kibana object type, however, the `_type` going forward is `doc` and the type is within the document, see below. Using v5 doc types in v6 and above will result in errors from Elasticsearch after one or more document types are used.

## Example Usage

```tf
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


resource "elasticsearch_kibana_object" "test_visualization_v7" {
  body = <<EOF
[
  {
    "_id": "response-time-percentile",
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

resource "elasticsearch_kibana_object" "test_index_pattern_v6" {
  body = <<EOF
[
  {
    "_id": "index-pattern:cloudwatch",
    "_type": "doc",
    "_source": {
      "type": "index-pattern",
      "index-pattern": {
        "title": "cloudwatch-*",
        "timeFieldName": "timestamp"
      }
    }
  }
]
EOF
}

resource "elasticsearch_kibana_object" "test_index_pattern_v7" {
  body = <<EOF
[
  {
    "_id": "index-pattern:cloudwatch",
    "_type": "doc",
    "_source": {
      "type": "index-pattern",
      "index-pattern": {
        "title": "cloudwatch-*",
        "timeFieldName": "timestamp"
      }
    }
  }
]
EOF
}
```

## Argument Reference

The following arguments are supported:

* `body` - (Required) The JSON body of the kibana object.
* `index` - (Optional) The name of the index where kibana data is stored.

## Attributes Reference

The following attributes are exported:

* `id` - The identifier of the kibana object.
