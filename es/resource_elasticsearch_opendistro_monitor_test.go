package es

import (
	"fmt"
	"testing"

	elastic7 "github.com/olivere/elastic/v7"
	elastic6 "gopkg.in/olivere/elastic.v6"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccElasticsearchOpenDistroMonitor(t *testing.T) {
	opensearchVerionConstraints, _ := version.NewConstraint(">= 1.1, < 6")
	var config string
	var check resource.TestCheckFunc
	meta := testAccOpendistroProvider.Meta()
	v, err := version.NewVersion(meta.(*ProviderConf).esVersion)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if opensearchVerionConstraints.Check(v) {
		config = testAccElasticsearchOpenDistroMonitorOpenSearch11
		check = resource.ComposeTestCheckFunc(
			testCheckElasticsearchOpenDistroMonitorExists("elasticsearch_opendistro_monitor.test_monitor1"),
			testCheckElasticsearchOpenDistroMonitorExists("elasticsearch_opendistro_monitor.test_monitor2"),
		)
	} else {
		config = testAccElasticsearchOpenDistroMonitorV7
		check = resource.ComposeTestCheckFunc(
			testCheckElasticsearchOpenDistroMonitorExists("elasticsearch_opendistro_monitor.test_monitor"),
		)
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccOpendistroProviders,
		CheckDestroy: testCheckElasticsearchMonitorDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check:  check,
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
		switch esClient.(type) {
		case *elastic7.Client:
			_, err = resourceElasticsearchOpenDistroGetMonitor(rs.Primary.ID, meta.(*ProviderConf))
		case *elastic6.Client:
			_, err = resourceElasticsearchOpenDistroGetMonitor(rs.Primary.ID, meta.(*ProviderConf))
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
		switch esClient.(type) {
		case *elastic7.Client:
			_, err = resourceElasticsearchOpenDistroGetMonitor(rs.Primary.ID, meta.(*ProviderConf))

		case *elastic6.Client:
			_, err = resourceElasticsearchOpenDistroGetMonitor(rs.Primary.ID, meta.(*ProviderConf))
		default:
		}

		if err != nil {
			return nil // should be not found error
		}

		return fmt.Errorf("Monitor %q still exists", rs.Primary.ID)
	}

	return nil
}

var testAccElasticsearchOpenDistroMonitorV7 = `
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
      "indices": ["*"],
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

var testAccElasticsearchOpenDistroMonitorOpenSearch11 = `
resource "elasticsearch_opendistro_monitor" "test_monitor1" {
  body = <<EOF
{
  "name": "test monitor",
  "type": "monitor",
  "monitor_type": "query_level_monitor",
  "enabled": true,
  "schedule": {
    "period": {
      "unit": "MINUTES",
      "interval": 1
    }
  },
  "inputs": [
    {
      "search": {
        "indices": [
          "*"
        ],
        "query": {
          "size": 0,
          "aggregations": {},
          "query": {
            "bool": {
              "adjust_pure_negative": true,
              "boost": 1,
              "filter": [
                {
                  "range": {
                    "@timestamp": {
                      "boost": 1,
                      "from": "{{period_end}}||-1h",
                      "to": "{{period_end}}",
                      "format": "epoch_millis",
                      "include_lower": true,
                      "include_upper": true
                    }
                  }
                }
              ]
            }
          }
        }
      }
    }
  ],
  "triggers": [
    {
      "query_level_trigger": {
        "name": "test trigger",
        "severity": "1",
        "condition": {
          "script": {
            "source": "ctx.results[0].hits.total.value < 1",
            "lang": "painless"
          }
        },
        "actions": [
          {
            "name": "test action",
            "destination_id": "iLd4fYIB6tDIVsstcu8w",
            "message_template": {
              "source": "Alert message.",
              "lang": "mustache"
            },
            "throttle_enabled": false,
            "subject_template": {
              "source": "Alert subject",
              "lang": "mustache"
            }
          }
        ]
      }
    }
  ]
}
EOF
}


resource "elasticsearch_opendistro_monitor" "test_monitor2" {
  body = <<EOF
{
  "name": "test-bucket-level-monitor",
  "type": "monitor",
  "monitor_type": "bucket_level_monitor",
  "enabled": true,
  "schedule": {
    "period": {
      "unit": "MINUTES",
      "interval": 1
    }
  },
  "inputs": [
    {
      "search": {
        "indices": [
          "*"
        ],
        "query": {
          "size": 0,
          "aggregations": {
            "composite_agg": {
              "composite": {
                "size": 10,
                "sources": [
                  {
                    "httpRequest.clientIp": {
                      "terms": {
                        "field": "httpRequest.clientIp",
                        "missing_bucket": false,
                        "order": "asc"
                      }
                    }
                  }
                ]
              }
            }
          },
          "query": {
            "bool": {
              "adjust_pure_negative": true,
              "boost": 1,
              "filter": [
                {
                  "range": {
                    "timestamp": {
                      "boost": 1,
                      "from": "{{period_end}}||-1h",
                      "to": "{{period_end}}",
                      "format": "epoch_millis",
                      "include_lower": true,
                      "include_upper": true
                    }
                  }
                }
              ]
            }
          }
        }
      }
    }
  ],
  "triggers": [
    {
      "bucket_level_trigger": {
        "name": "test trigger",
        "severity": "1",
        "condition": {
          "buckets_path": {
            "_count": "_count"
          },
          "parent_bucket_path": "composite_agg",
          "script": {
            "source": "params._count > 10000",
            "lang": "painless"
          },
          "gap_policy": "skip"
        },
        "actions": [
          {
            "name": "test action",
            "destination_id": "iLd4fYIB6tDIVsstcu8w",
            "message_template": {
              "source": "Alert message.",
              "lang": "mustache"
            },
            "throttle_enabled": false,
            "subject_template": {
              "source": "Alert subject",
              "lang": "mustache"
            },
            "action_execution_policy": {
              "action_execution_scope": {
                "per_alert": {
                  "actionable_alerts": [
                    "DEDUPED",
                    "NEW"
                  ]
                }
              }
            }
          }
        ]
      }
    }
  ]
}
EOF
}
`
