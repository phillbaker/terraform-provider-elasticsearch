package main

import (
	"context"
	"fmt"
	"testing"

	elastic "gopkg.in/olivere/elastic.v5"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccElasticsearchKibanaObject(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
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

		conn := testAccProvider.Meta().(*elastic.Client)
		// _, err := //conn.IndexGetTemplate(rs.Primary.ID).Do(context.TODO())
		_, err := conn.Get().Index(".kibana").Type("visualization").Id("response-time-percentile").Do(context.TODO())
		if err != nil {
			return err
		}

		return nil
	}
}

func testCheckElasticsearchKibanaObjectDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*elastic.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticsearch_kibana_object" {
			continue
		}

		// _, err := // conn.IndexGetTemplate(rs.Primary.ID).Do(context.TODO())
		_, err := conn.Get().Index(".kibana").Type("visualization").Id("response-time-percentile").Do(context.TODO())
		if err != nil {
			return nil
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
