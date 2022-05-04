package es

import (
	"context"
	"errors"
	"fmt"
	"testing"

	elastic7 "github.com/olivere/elastic/v7"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccElasticsearchDataStream(t *testing.T) {
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
	case *elastic7.Client:
		allowed = true
	default:
		allowed = false
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)

			if !allowed {
				t.Skip("/_data_stream endpoint only supported on ES >= 7.9")
			}
		},
		Providers:    testAccProviders,
		CheckDestroy: testCheckElasticsearchDataStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccElasticsearchDataStream,
				Check: resource.ComposeTestCheckFunc(
					testCheckElasticsearchDataStreamExists("elasticsearch_data_stream.foo"),
				),
			},
		},
	})
}

func testCheckElasticsearchDataStreamExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No data stream ID is set")
		}

		meta := testAccProvider.Meta()

		var err error
		esClient, err := getClient(meta.(*ProviderConf))
		if err != nil {
			return err
		}
		switch client := esClient.(type) {
		case *elastic7.Client:
			err = elastic7GetDataStream(client, rs.Primary.ID)
		default:
			return errors.New("Elasticsearch version not supported")
		}

		if err != nil {
			return err
		}

		return nil
	}
}

func testCheckElasticsearchDataStreamDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticsearch_data_stream" {
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
			err = elastic7GetDataStream(client, rs.Primary.ID)
		default:
			return errors.New("Elasticsearch version not supported")
		}

		if err != nil {
			return nil // should be not found error
		}

		return fmt.Errorf("Data stream %q still exists", rs.Primary.ID)
	}

	return nil
}

var testAccElasticsearchDataStream = `
resource "elasticsearch_composable_index_template" "foo" {
  name = "foo-template"
  body = <<EOF
{
  "index_patterns": ["foo-data-stream*"],
  "data_stream": {}
}
EOF
}

resource "elasticsearch_data_stream" "foo" {
  name       = "foo-data-stream"
  depends_on = [elasticsearch_composable_index_template.foo]
}
`
