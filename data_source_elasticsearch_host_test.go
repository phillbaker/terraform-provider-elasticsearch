package main

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccElasticsearchDataSourceHost_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccElasticsearchDataSourceHost,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticsearch_host.test", "id"),
				),
			},
		},
	})
}

var testAccElasticsearchDataSourceHost = `
data "elasticsearch_host" "test" {
  active = true
}
`
