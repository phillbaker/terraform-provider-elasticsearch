package es

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"

	elastic7 "github.com/olivere/elastic/v7"
	elastic5 "gopkg.in/olivere/elastic.v5"
	elastic6 "gopkg.in/olivere/elastic.v6"
)

const (
	testAccElasticsearchIndex = `
resource "elasticsearch_index" "test" {
  name = "terraform-test"
  number_of_shards = 1
  number_of_replicas = 1
}
`
	testAccElasticsearchIndexUpdate1 = `
resource "elasticsearch_index" "test" {
  name = "terraform-test"
  number_of_shards = 1
  number_of_replicas = 2
}
`
)

func TestAccElasticsearchIndex(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: checkElasticsearchIndexDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccElasticsearchIndex,
				Check: resource.ComposeTestCheckFunc(
					checkElasticsearchIndexExists("elasticsearch_index.test"),
				),
			},
			{
				Config: testAccElasticsearchIndexUpdate1,
				Check: resource.ComposeTestCheckFunc(
					checkElasticsearchIndexUpdated("elasticsearch_index.test"),
				),
			},
		},
	})
}

func TestAccElasticsearchIndex_importBasic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: checkElasticsearchIndexDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccElasticsearchIndex,
			},
			{
				ResourceName:      "elasticsearch_index.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					// not returned from the API
					"force_destroy",
				},
			},
		},
	})
}

func checkElasticsearchIndexExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("index ID not set")
		}

		meta := testAccProvider.Meta()

		var err error
		switch client := meta.(type) {
		case *elastic7.Client:
			_, err = client.IndexGetSettings(rs.Primary.ID).Do(context.TODO())
		case *elastic6.Client:
			_, err = client.IndexGetSettings(rs.Primary.ID).Do(context.TODO())
		default:
			elastic5Client := meta.(*elastic5.Client)
			_, err = elastic5Client.IndexGetSettings(rs.Primary.ID).Do(context.TODO())
		}

		return err
	}
}

func checkElasticsearchIndexUpdated(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("index ID not set")
		}

		meta := testAccProvider.Meta()
		settings := make(map[string]interface{})

		switch client := meta.(type) {
		case *elastic7.Client:
			resp, err := client.IndexGetSettings(rs.Primary.ID).Do(context.TODO())
			if err != nil {
				return err
			}
			settings = resp[rs.Primary.ID].Settings["index"].(map[string]interface{})

		case *elastic6.Client:
			resp, err := client.IndexGetSettings(rs.Primary.ID).Do(context.TODO())
			if err != nil {
				return err
			}
			settings = resp[rs.Primary.ID].Settings["index"].(map[string]interface{})

		default:
			elastic5Client := meta.(*elastic5.Client)
			resp, err := elastic5Client.IndexGetSettings(rs.Primary.ID).Do(context.TODO())
			if err != nil {
				return err
			}
			settings = resp[rs.Primary.ID].Settings["index"].(map[string]interface{})

		}

		r, ok := settings["number_of_replicas"]
		if ok {
			if ir := r.(string); ir != "2" {
				return fmt.Errorf("expected 2 got %s", ir)
			}
			return nil
		}

		return errors.New("field not found")
	}
}

func checkElasticsearchIndexDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticsearch_index" {
			continue
		}

		meta := testAccProvider.Meta()

		var err error
		switch client := meta.(type) {
		case *elastic7.Client:
			_, err = client.IndexGetSettings(rs.Primary.ID).Do(context.TODO())
		case *elastic6.Client:
			_, err = client.IndexGetSettings(rs.Primary.ID).Do(context.TODO())
		default:
			elastic5Client := meta.(*elastic5.Client)
			_, err = elastic5Client.IndexGetSettings(rs.Primary.ID).Do(context.TODO())
		}

		if err != nil {
			return nil // should be not found error
		}

		return fmt.Errorf("index %q still exists", rs.Primary.ID)
	}

	return nil
}
