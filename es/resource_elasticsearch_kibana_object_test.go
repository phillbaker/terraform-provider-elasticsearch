package es

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"testing"

	elastic7 "github.com/olivere/elastic/v7"
	elastic5 "gopkg.in/olivere/elastic.v5"
	elastic6 "gopkg.in/olivere/elastic.v6"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccElasticsearchKibanaObject(t *testing.T) {

	provider := Provider().(*schema.Provider)
	err := provider.Configure(&terraform.ResourceConfig{})
	if err != nil {
		t.Skipf("err: %s", err)
	}

	var visualizationConfig string
	var indexPatternConfig string
	meta := testAccProvider.Meta()

	switch meta.(type) {
	case *elastic7.Client:
		visualizationConfig = testAccElasticsearch7KibanaVisualization
		indexPatternConfig = testAccElasticsearch7KibanaIndexPattern
	case *elastic6.Client:
		visualizationConfig = testAccElasticsearch6KibanaVisualization
		indexPatternConfig = testAccElasticsearch6KibanaIndexPattern
	default:
		visualizationConfig = testAccElasticsearchKibanaVisualization
		indexPatternConfig = testAccElasticsearchKibanaIndexPattern
	}

	resource.Test(t, resource.TestCase{
		Providers:    testAccProviders,
		CheckDestroy: testCheckElasticsearchKibanaObjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: visualizationConfig,
				Check: resource.ComposeTestCheckFunc(
					testCheckElasticsearchKibanaObjectExists("elasticsearch_kibana_object.test_visualization", "visualization", "response-time-percentile"),
				),
			},
			{
				Config: indexPatternConfig,
				Check: resource.ComposeTestCheckFunc(
					testCheckElasticsearchKibanaObjectExists("elasticsearch_kibana_object.test_pattern", "index-pattern", "index-pattern:cloudwatch"),
				),
			},
		},
	})
}

func TestAccElasticsearchKibanaObject_ProviderFormatInvalid(t *testing.T) {
	provider := Provider().(*schema.Provider)
	err := provider.Configure(&terraform.ResourceConfig{})
	if err != nil {
		t.Skipf("err: %s", err)
	}

	resource.Test(t, resource.TestCase{
		Providers:    testAccProviders,
		CheckDestroy: testCheckElasticsearchKibanaObjectDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccElasticsearchFormatInvalid,
				ExpectError: regexp.MustCompile("must be an array of objects"),
			},
		},
	})
}

func TestAccElasticsearchKibanaObject_Rejected(t *testing.T) {
	provider := Provider().(*schema.Provider)
	err := provider.Configure(&terraform.ResourceConfig{})
	if err != nil {
		t.Skipf("err: %s", err)
	}
	meta := provider.Meta()
	var allowed bool

	switch meta.(type) {
	case *elastic6.Client:
		allowed = true
	default:
		allowed = false
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			if !allowed {
				t.Skip("Only >= ES 6 has index type restrictions")
			}
		},
		Providers:    testAccProviders,
		CheckDestroy: testCheckElasticsearchKibanaObjectDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccElasticsearchKibanaIndexPattern,
				ExpectError: regexp.MustCompile("Error 400"),
			},
		},
	})
}

func testCheckElasticsearchKibanaObjectExists(name string, objectType string, id string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No kibana object ID is set")
		}

		meta := testAccProvider.Meta()

		var err error
		esClient, err := getClient(meta.(*ProviderConf))
		if err != nil {
			return err
		}
		switch client := esClient.(type) {
		case *elastic7.Client:
			_, err = client.Get().Index(".kibana").Id(id).Do(context.TODO())
		case *elastic6.Client:
			_, err = client.Get().Index(".kibana").Type(deprecatedDocType).Id(id).Do(context.TODO())
		default:
			elastic5Client := client.(*elastic5.Client)
			_, err = elastic5Client.Get().Index(".kibana").Type(objectType).Id(id).Do(context.TODO())
		}

		if err != nil {
			log.Printf("[INFO] testCheckElasticsearchKibanaObjectExists: %+v", err)
			return err
		}

		return nil
	}
}

func testCheckElasticsearchKibanaObjectDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticsearch_kibana_object" {
			continue
		}

		meta := testAccProvider.Meta()

		var err error
		esClient, err := getClient(meta.(*ProviderConf))
		if err != nil {
			return err
		}
		switch client := esClient.(type) {
		case *elastic7.Client:
			_, err = client.Get().Index(".kibana").Id("response-time-percentile").Do(context.TODO())
		case *elastic6.Client:
			_, err = client.Get().Index(".kibana").Type("visualization").Id("response-time-percentile").Do(context.TODO())
		default:
			elastic5Client := client.(*elastic5.Client)
			_, err = elastic5Client.Get().Index(".kibana").Type("visualization").Id("response-time-percentile").Do(context.TODO())
		}

		if err != nil {
			if elastic7.IsNotFound(err) || elastic6.IsNotFound(err) || elastic5.IsNotFound(err) {
				return nil // should be not found error
			}

			// Fail on any other error
			return fmt.Errorf("Unexpected error %s", err)
		}

		return fmt.Errorf("Kibana object %q still exists", rs.Primary.ID)
	}

	return nil
}

var testAccElasticsearchKibanaVisualization = `
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
`

var testAccElasticsearch6KibanaVisualization = `
resource "elasticsearch_kibana_object" "test_visualization" {
  body = <<EOF
[
  {
    "_id": "response-time-percentile",
    "_type": "doc",
    "_source": {
      "title": "Total response time percentiles",
      "visState": "{\"title\":\"Total response time percentiles\",\"type\":\"line\",\"params\":{\"addTooltip\":true,\"addLegend\":true,\"legendPosition\":\"right\",\"showCircles\":true,\"interpolate\":\"linear\",\"scale\":\"linear\",\"drawLinesBetweenPoints\":true,\"radiusRatio\":9,\"times\":[],\"addTimeMarker\":false,\"defaultYExtents\":false,\"setYExtents\":false},\"aggs\":[{\"id\":\"1\",\"enabled\":true,\"type\":\"percentiles\",\"schema\":\"metric\",\"params\":{\"field\":\"app.total_time\",\"percents\":[50,90,95]}},{\"id\":\"2\",\"enabled\":true,\"type\":\"date_histogram\",\"schema\":\"segment\",\"params\":{\"field\":\"@timestamp\",\"interval\":\"auto\",\"customInterval\":\"2h\",\"min_doc_count\":1,\"extended_bounds\":{}}},{\"id\":\"3\",\"enabled\":true,\"type\":\"terms\",\"schema\":\"group\",\"params\":{\"field\":\"system.syslog.program\",\"size\":5,\"order\":\"desc\",\"orderBy\":\"_term\"}}],\"listeners\":{}}",
      "uiStateJSON": "{}",
      "description": "",
      "version": 1,
      "kibanaSavedObjectMeta": {
        "searchSourceJSON": "{\"index\":\"filebeat-*\",\"query\":{\"query_string\":{\"query\":\"*\",\"analyze_wildcard\":true}},\"filter\":[]}"
      },
      "type": "visualization"
    }
  }
]
EOF
}
`

var testAccElasticsearch7KibanaVisualization = `
resource "elasticsearch_kibana_object" "test_visualization" {
  body = <<EOF
[
  {
    "_id": "response-time-percentile",
    "_source": {
      "title": "Total response time percentiles",
      "visState": "{\"title\":\"Total response time percentiles\",\"type\":\"line\",\"params\":{\"addTooltip\":true,\"addLegend\":true,\"legendPosition\":\"right\",\"showCircles\":true,\"interpolate\":\"linear\",\"scale\":\"linear\",\"drawLinesBetweenPoints\":true,\"radiusRatio\":9,\"times\":[],\"addTimeMarker\":false,\"defaultYExtents\":false,\"setYExtents\":false},\"aggs\":[{\"id\":\"1\",\"enabled\":true,\"type\":\"percentiles\",\"schema\":\"metric\",\"params\":{\"field\":\"app.total_time\",\"percents\":[50,90,95]}},{\"id\":\"2\",\"enabled\":true,\"type\":\"date_histogram\",\"schema\":\"segment\",\"params\":{\"field\":\"@timestamp\",\"interval\":\"auto\",\"customInterval\":\"2h\",\"min_doc_count\":1,\"extended_bounds\":{}}},{\"id\":\"3\",\"enabled\":true,\"type\":\"terms\",\"schema\":\"group\",\"params\":{\"field\":\"system.syslog.program\",\"size\":5,\"order\":\"desc\",\"orderBy\":\"_term\"}}],\"listeners\":{}}",
      "uiStateJSON": "{}",
      "description": "",
      "version": 1,
      "kibanaSavedObjectMeta": {
        "searchSourceJSON": "{\"index\":\"filebeat-*\",\"query\":{\"query_string\":{\"query\":\"*\",\"analyze_wildcard\":true}},\"filter\":[]}"
      },
      "type": "visualization"
    }
  }
]
EOF
}
`

var testAccElasticsearchKibanaIndexPattern = `
resource "elasticsearch_kibana_object" "test_pattern" {
  body = <<EOF
[
  {
		"_id": "index-pattern:cloudwatch",
		"_type": "index-pattern",
		"_source": {
			"title": "cloudwatch-*",
			"timeFieldName": "timestamp"
		}
	}
]
EOF
}
`

var testAccElasticsearch6KibanaIndexPattern = `
resource "elasticsearch_kibana_object" "test_pattern" {
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
`

var testAccElasticsearch7KibanaIndexPattern = `
resource "elasticsearch_kibana_object" "test_pattern" {
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
`

var testAccElasticsearchFormatInvalid = `
resource "elasticsearch_kibana_object" "test_invalid" {
  body = <<EOF
{
  "test": "yes"
}
EOF
}
`
