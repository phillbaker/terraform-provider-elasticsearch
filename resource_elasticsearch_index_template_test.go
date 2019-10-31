package main

import (
	"context"
	"fmt"
	"testing"

	elastic7 "github.com/olivere/elastic/v7"
	elastic5 "gopkg.in/olivere/elastic.v5"
	elastic6 "gopkg.in/olivere/elastic.v6"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccElasticsearchIndexTemplate(t *testing.T) {
	provider := Provider().(*schema.Provider)
	err := provider.Configure(&terraform.ResourceConfig{})
	if err != nil {
		t.Skipf("err: %s", err)
	}
	meta := provider.Meta()
	var config string
	switch meta.(type) {
	case *elastic7.Client:
		config = testAccElasticsearchIndexTemplateV7
	case *elastic6.Client:
		config = testAccElasticsearchIndexTemplateV6
	default:
		config = testAccElasticsearchIndexTemplateV5
	}
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testCheckElasticsearchIndexTemplateDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckElasticsearchIndexTemplateExists("elasticsearch_index_template.test"),
				),
			},
		},
	})
}

func testCheckElasticsearchIndexTemplateExists(name string) resource.TestCheckFunc {
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
			_, err = client.IndexGetTemplate(rs.Primary.ID).Do(context.TODO())
		case *elastic6.Client:
			client := meta.(*elastic6.Client)
			_, err = client.IndexGetTemplate(rs.Primary.ID).Do(context.TODO())
		default:
			client := meta.(*elastic5.Client)
			_, err = client.IndexGetTemplate(rs.Primary.ID).Do(context.TODO())
		}

		if err != nil {
			return err
		}

		return nil
	}
}

func testCheckElasticsearchIndexTemplateDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticsearch_index_template" {
			continue
		}

		meta := testAccProvider.Meta()

		var err error
		switch meta.(type) {
		case *elastic7.Client:
			client := meta.(*elastic7.Client)
			_, err = client.IndexGetTemplate(rs.Primary.ID).Do(context.TODO())
		case *elastic6.Client:
			client := meta.(*elastic6.Client)
			_, err = client.IndexGetTemplate(rs.Primary.ID).Do(context.TODO())
		default:
			client := meta.(*elastic5.Client)
			_, err = client.IndexGetTemplate(rs.Primary.ID).Do(context.TODO())
		}

		if err != nil {
			return nil // should be not found error
		}

		return fmt.Errorf("Index template %q still exists", rs.Primary.ID)
	}

	return nil
}

var testAccElasticsearchIndexTemplateV5 = `
resource "elasticsearch_index_template" "test" {
  name = "terraform-test"
  body = <<EOF
{
  "template": "te*",
  "settings": {
    "index": {
      "number_of_shards": 1
    }
  },
  "mappings": {
    "type1": {
      "_source": {
        "enabled": false
      },
      "properties": {
        "host_name": {
          "type": "keyword"
        },
        "created_at": {
          "type": "date",
          "format": "EEE MMM dd HH:mm:ss Z YYYY"
        }
      }
    }
  }
}
EOF
}
`

var testAccElasticsearchIndexTemplateV6 = `
resource "elasticsearch_index_template" "test" {
  name = "terraform-test"
  body = <<EOF
{
  "index_patterns": ["te*", "bar*"],
  "settings": {
    "index": {
      "number_of_shards": 1
    }
  },
  "mappings": {
    "type1": {
      "_source": {
        "enabled": false
      },
      "properties": {
        "host_name": {
          "type": "keyword"
        },
        "created_at": {
          "type": "date",
          "format": "EEE MMM dd HH:mm:ss Z YYYY"
        }
      }
    }
  }
}
EOF
}
`

var testAccElasticsearchIndexTemplateV7 = `
resource "elasticsearch_index_template" "test" {
  name = "terraform-test"
  body = <<EOF
{
  "index_patterns": ["te*", "bar*"],
  "settings": {
    "index": {
      "number_of_shards": 1
    }
  },
  "mappings": {
    "_source": {
      "enabled": false
    },
    "properties": {
      "host_name": {
        "type": "keyword"
      },
      "created_at": {
        "type": "date",
        "format": "EEE MMM dd HH:mm:ss Z YYYY"
      }
    }
  }
}
EOF
}
`
