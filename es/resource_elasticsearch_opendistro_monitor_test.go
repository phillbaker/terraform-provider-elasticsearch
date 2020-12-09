package es

import (
	"fmt"
	"testing"

	elastic7 "github.com/olivere/elastic/v7"
	elastic5 "gopkg.in/olivere/elastic.v5"
	elastic6 "gopkg.in/olivere/elastic.v6"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccElasticsearchOpenDistroMonitor(t *testing.T) {
	provider := Provider().(*schema.Provider)
	err := provider.Configure(&terraform.ResourceConfig{})
	if err != nil {
		t.Skipf("err: %s", err)
	}
	meta := provider.Meta()
	var allowed bool

	switch meta.(type) {
	case *elastic5.Client:
		allowed = false
	default:
		allowed = true
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			if !allowed {
				t.Skip("Monitors only supported on >= ES 6")
			}
		},
		Providers:    testAccOpendistroProviders,
		CheckDestroy: testCheckElasticsearchMonitorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccElasticsearchOpenDistroMonitor,
				Check: resource.ComposeTestCheckFunc(
					testCheckElasticsearchOpenDistroMonitorExists("elasticsearch_opendistro_monitor.test_monitor"),
				),
			},
		},
	})
}

func testCheckElasticsearchOpenDistroMonitorExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No monitor ID is set")
		}

		meta := testAccOpendistroProvider.Meta()

		var err error
		esClient, err := getClient(meta.(*ProviderConf))
		if err != nil {
			return err
		}
		switch client := esClient.(type) {
		case *elastic7.Client:
			_, err = resourceElasticsearchOpenDistroGetMonitor(rs.Primary.ID, client)
		case *elastic6.Client:
			_, err = resourceElasticsearchOpenDistroGetMonitor(rs.Primary.ID, client)
		default:
		}

		if err != nil {
			return err
		}

		return nil
	}
}

func testCheckElasticsearchMonitorDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticsearch_opendistro_monitor" {
			continue
		}

		meta := testAccOpendistroProvider.Meta()

		var err error
		esClient, err := getClient(meta.(*ProviderConf))
		if err != nil {
			return err
		}
		switch client := esClient.(type) {
		case *elastic7.Client:
			_, err = resourceElasticsearchOpenDistroGetMonitor(rs.Primary.ID, client)

		case *elastic6.Client:
			_, err = resourceElasticsearchOpenDistroGetMonitor(rs.Primary.ID, client)
		default:
		}

		if err != nil {
			return nil // should be not found error
		}

		return fmt.Errorf("Monitor %q still exists", rs.Primary.ID)
	}

	return nil
}

var testAccElasticsearchOpenDistroMonitor = `
resource "elasticsearch_opendistro_monitor" "test_monitor" {
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
  "triggers": []
}
EOF
}
`
