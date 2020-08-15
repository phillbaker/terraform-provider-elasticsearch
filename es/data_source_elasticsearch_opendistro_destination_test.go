package es

import (
	"context"
	"testing"

	elastic5 "gopkg.in/olivere/elastic.v5"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccElasticsearchDataSourceDestination_basic(t *testing.T) {
	provider := Provider()
	diags := provider.Configure(context.Background(), &terraform.ResourceConfig{})
	if diags.HasError() {
		t.Skipf("err: %#v", diags)
	}
	meta := provider.Meta()
	esClient, err := getClient(meta.(*ProviderConf))
	if err != nil {
		t.Skipf("err: %s", err)
	}
	var allowed bool
	switch esClient.(type) {
	case *elastic5.Client:
		allowed = false
	default:
		allowed = true
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			if !allowed {
				t.Skip("Destinations only supported on >= ES 6")
			}
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
  name = "${element(list("my-destination", "${elasticsearch_opendistro_destination.test.body}"), 0)}"
}
`
