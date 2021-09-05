package es

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestAccElasticsearchDataSourceHost_basic(t *testing.T) {
	var providers []*schema.Provider
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviderFactories(&providers),
		Steps: []resource.TestStep{
			{
				Config: testAccElasticsearchDataSourceHost,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticsearch_host.test", "id"),
					resource.TestCheckResourceAttrSet("data.elasticsearch_host.test", "url"),
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
