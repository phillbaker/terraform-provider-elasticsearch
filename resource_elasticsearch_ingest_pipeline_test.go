package main

import (
	"context"
	"fmt"
	"testing"

	elastic7 "github.com/olivere/elastic/v7"
	elastic5 "gopkg.in/olivere/elastic.v5"
	elastic6 "gopkg.in/olivere/elastic.v6"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccElasticsearchIngestPipeline(t *testing.T) {
	provider := Provider().(*schema.Provider)
	err := provider.Configure(&terraform.ResourceConfig{})
	if err != nil {
		t.Skipf("err: %s", err)
	}
	meta := provider.Meta()
	var config string
	switch meta.(type) {
	case *elastic7.Client:
		config = testAccElasticsearchIngestPipelineV7
	case *elastic6.Client:
		config = testAccElasticsearchIngestPipelineV6
	default:
		config = testAccElasticsearchIngestPipelineV5
	}
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testCheckElasticsearchIngestPipelineDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckElasticsearchIngestPipelineExists("elasticsearch_ingest_pipeline.test"),
				),
			},
		},
	})
}

func testCheckElasticsearchIngestPipelineExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No index template ID is set")
		}

		meta := testAccProvider.Meta()

		var err error
		switch meta.(type) {
		case *elastic7.Client:
			client := meta.(*elastic7.Client)
			_, err = client.IngestGetPipeline(rs.Primary.ID).Do(context.TODO())
		case *elastic6.Client:
			client := meta.(*elastic6.Client)
			_, err = client.IngestGetPipeline(rs.Primary.ID).Do(context.TODO())
		default:
			client := meta.(*elastic5.Client)
			_, err = client.IngestGetPipeline(rs.Primary.ID).Do(context.TODO())
		}

		if err != nil {
			return err
		}

		return nil
	}
}

func testCheckElasticsearchIngestPipelineDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticsearch_ingest_pipeline" {
			continue
		}

		meta := testAccProvider.Meta()

		var err error
		switch meta.(type) {
		case *elastic7.Client:
			client := meta.(*elastic7.Client)
			_, err = client.IngestGetPipeline(rs.Primary.ID).Do(context.TODO())
		case *elastic6.Client:
			client := meta.(*elastic6.Client)
			_, err = client.IngestGetPipeline(rs.Primary.ID).Do(context.TODO())
		default:
			client := meta.(*elastic5.Client)
			_, err = client.IngestGetPipeline(rs.Primary.ID).Do(context.TODO())
		}

		if err != nil {
			return nil // should be not found error
		}

		return fmt.Errorf("Index template %q still exists", rs.Primary.ID)
	}

	return nil
}

var testAccElasticsearchIngestPipelineV5 = `
resource "elasticsearch_ingest_pipeline" "test" {
  name = "terraform-test"
  body = <<EOF
{
  "description" : "describe pipeline",
  "processors" : [
    {
      "set" : {
        "field": "foo",
        "value": "bar"
      }
    }
  ]
}
EOF
}
`

var testAccElasticsearchIngestPipelineV6 = `
resource "elasticsearch_ingest_pipeline" "test" {
  name = "terraform-test"
  body = <<EOF
{
  "description" : "describe pipeline",
  "version": 123,
  "processors" : [
    {
      "set" : {
        "field": "foo",
        "value": "bar"
      }
    }
  ]
}
EOF
}
`

var testAccElasticsearchIngestPipelineV7 = `
resource "elasticsearch_ingest_pipeline" "test" {
  name = "terraform-test"
  body = <<EOF
{
  "description" : "describe pipeline",
  "version": 123,
  "processors" : [
    {
      "set" : {
        "field": "foo",
        "value": "bar"
      }
    }
  ]
}
EOF
}
`
