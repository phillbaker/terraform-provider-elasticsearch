package main

import (
	"context"
	"fmt"
	"testing"

	elastic7 "github.com/olivere/elastic/v7"
	elastic5 "gopkg.in/olivere/elastic.v5"
	elastic6 "gopkg.in/olivere/elastic.v6"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccElasticsearchSnapshotRepository(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckElasticsearchSnapshotRepositoryDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccElasticsearchSnapshotRepository,
				Check: resource.ComposeTestCheckFunc(
					testCheckElasticsearchSnapshotRepositoryExists("elasticsearch_snapshot_repository.test"),
				),
			},
		},
	})
}

func testCheckElasticsearchSnapshotRepositoryExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No snapshot repository ID is set")
		}

		meta := testAccProvider.Meta()

		var err error
		switch meta.(type) {
		case *elastic7.Client:
			client := meta.(*elastic7.Client)
			_, err = client.SnapshotGetRepository(rs.Primary.ID).Do(context.TODO())
		case *elastic6.Client:
			client := meta.(*elastic6.Client)
			_, err = client.SnapshotGetRepository(rs.Primary.ID).Do(context.TODO())
		default:
			client := meta.(*elastic5.Client)
			_, err = client.SnapshotGetRepository(rs.Primary.ID).Do(context.TODO())
		}

		if err != nil {
			return err
		}

		return nil
	}
}

func testCheckElasticsearchSnapshotRepositoryDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticsearch_snapshot_repository" {
			continue
		}

		meta := testAccProvider.Meta()

		var err error
		switch meta.(type) {
		case *elastic7.Client:
			client := meta.(*elastic7.Client)
			_, err = client.SnapshotGetRepository(rs.Primary.ID).Do(context.TODO())
		case *elastic6.Client:
			client := meta.(*elastic6.Client)
			_, err = client.SnapshotGetRepository(rs.Primary.ID).Do(context.TODO())
		default:
			client := meta.(*elastic5.Client)
			_, err = client.SnapshotGetRepository(rs.Primary.ID).Do(context.TODO())
		}

		if err != nil {
			return nil // should be not found error
		}

		return fmt.Errorf("Snapshot repository %q still exists", rs.Primary.ID)
	}

	return nil
}

var testAccElasticsearchSnapshotRepository = `
resource "elasticsearch_snapshot_repository" "test" {
  name = "terraform-test"
  type = "fs"

  settings = {
    location = "/tmp/elasticsearch"
  }
}
`
