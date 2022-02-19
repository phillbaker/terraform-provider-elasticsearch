package es

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccElasticsearchDataSourceDestination_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccOpendistroProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccElasticsearchDataSourceDestination,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticsearch_opendistro_destination.test", "id"),
					resource.TestCheckResourceAttrSet("data.elasticsearch_opendistro_destination.test", "body.type"),
				),
			},
		},
	})
}

var testAccElasticsearchDataSourceDestination = `
resource "elasticsearch_opendistro_destination" "test" {
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

data "elasticsearch_opendistro_destination" "test" {
  # Ugh, song and dance to get the json value to force dependency
  name = "${element(tolist(["my-destination", "${elasticsearch_opendistro_destination.test.body}"]), 0)}"
}
`
