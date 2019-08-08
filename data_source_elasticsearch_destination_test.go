package main

import (
	"testing"

	elastic7 "github.com/olivere/elastic/v7"
	elastic5 "gopkg.in/olivere/elastic.v5"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccElasticsearchDataSourceDestination_basic(t *testing.T) {
	provider := Provider().(*schema.Provider)
	err := provider.Configure(&terraform.ResourceConfig{})
	if err != nil {
		t.Skipf("err: %s", err)
	}
	meta := provider.Meta()
	var allowed bool
	switch meta.(type) {
	case *elastic7.Client:
		allowed = false
	case *elastic5.Client:
		allowed = false
	default:
		allowed = true
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			if !allowed {
				t.Skip("Destinations only supported on ES 6, https://github.com/opendistro-for-elasticsearch/alerting/issues/66")
			}
		},
		Providers: testAccOpendistroProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccElasticsearchDataSourceDestination,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticsearch_destination.test", "id"),
				),
			},
		},
	})
}

var testAccElasticsearchDataSourceDestination = `
resource "elasticsearch_destination" "test" {
  body = <<EOF
{
  "name": "my-destination",
  "type": "slack",
  "slack": {
    "url": "http://www.example.com"
  }
}
EOF
}

data "elasticsearch_destination" "test" {
  # Ugh, song and dance to get the json value to force dependency
  name = "${element(list("my-destination", "${elasticsearch_destination.test.body}"), 0)}"
}
`
