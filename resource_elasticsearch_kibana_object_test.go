package main

import (
	"context"
	"fmt"
	"os"
	"testing"

	elastic7 "github.com/olivere/elastic/v7"
	elastic5 "gopkg.in/olivere/elastic.v5"
	elastic6 "gopkg.in/olivere/elastic.v6"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccElasticsearchKibanaObject(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			// skip tests on ES > 7 until saved object API is supported
			if v := os.Getenv("ES_VERSION"); v >= "7.0.0" {
				t.Skip("Need to implement saved object API on ES >= 6")
			}
		},
		Providers:    testAccProviders,
		CheckDestroy: testCheckElasticsearchKibanaObjectDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccElasticsearchKibanaObject,
				Check: resource.ComposeTestCheckFunc(
					testCheckElasticsearchKibanaObjectExists("elasticsearch_kibana_object.test_visualization"),
				),
			},
		},
	})
}

func testCheckElasticsearchKibanaObjectExists(name string) resource.TestCheckFunc {
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
		switch meta.(type) {
		case *elastic7.Client:
			// not implemented
		case *elastic6.Client:
			client := meta.(*elastic6.Client)
			_, err = client.Get().Index(".kibana").Type("visualization").Id("response-time-percentile").Do(context.TODO())
		default:
			client := meta.(*elastic5.Client)
			_, err = client.Get().Index(".kibana").Type("visualization").Id("response-time-percentile").Do(context.TODO())
		}

		if err != nil {
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
		switch meta.(type) {
		case *elastic7.Client:
			// not implemented
		case *elastic6.Client:
			client := meta.(*elastic6.Client)
			_, err = client.Get().Index(".kibana").Type("visualization").Id("response-time-percentile").Do(context.TODO())
		default:
			client := meta.(*elastic5.Client)
			_, err = client.Get().Index(".kibana").Type("visualization").Id("response-time-percentile").Do(context.TODO())
		}

		if err != nil {
			return nil // should be not found error
		}

		return fmt.Errorf("Kibana object %q still exists", rs.Primary.ID)
	}

	return nil
}

var testAccElasticsearchKibanaObject = `
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
