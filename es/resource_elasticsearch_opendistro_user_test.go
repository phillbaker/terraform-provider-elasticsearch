package es

import (
	"context"
	"fmt"
	"os"
	"testing"

	elastic7 "github.com/olivere/elastic/v7"
	elastic5 "gopkg.in/olivere/elastic.v5"
	elastic6 "gopkg.in/olivere/elastic.v6"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccElasticsearchOpendistroUser(t *testing.T) {

	provider := Provider().(*schema.Provider)
	err := provider.Configure(&terraform.ResourceConfig{})
	if err != nil {
		t.Skipf("err: %s", err)
	}
	meta := provider.Meta()
	var allowed bool
	switch meta.(type) {
	case *elastic5.Client:
		allowed = false
	case *elastic6.Client:
		allowed = false
	default:
		allowed = true
	}

	randomName := "test" + acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			if !allowed {
				t.Skip("Users only supported on ES >= 7")
			}
		},
		Providers:    testAccOpendistroProviders,
		CheckDestroy: testAccCheckElasticsearchOpendistroUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOpendistroUserResource(randomName),
				Check: resource.ComposeTestCheckFunc(
					testCheckElasticSearchOpendistroUserExists("elasticsearch_opendistro_user.test"),
					resource.TestCheckResourceAttr(
						"elasticsearch_opendistro_user.test",
						"id",
						randomName,
					),
					resource.TestCheckResourceAttr(
						"elasticsearch_opendistro_user.test",
						"backend_roles.#",
						"1",
					),
					resource.TestCheckResourceAttr(
						"elasticsearch_opendistro_user.test",
						"attributes.some_attribute",
						"alpha",
					),
					resource.TestCheckResourceAttr(
						"elasticsearch_opendistro_user.test",
						"description",
						"test",
					),
				),
			},
			{
				Config: testAccOpendistroUserResourceUpdated(randomName),
				Check: resource.ComposeTestCheckFunc(
					testCheckElasticSearchOpendistroUserExists("elasticsearch_opendistro_user.test"),
					testCheckElasticSearchOpendistroUserConnects("elasticsearch_opendistro_user.test"),
					resource.TestCheckResourceAttr(
						"elasticsearch_opendistro_user.test",
						"backend_roles.#",
						"2",
					),
				),
			},
			{
				Config: testAccOpendistroUserResourceMinimal(randomName),
				Check: resource.ComposeTestCheckFunc(
					testCheckElasticSearchOpendistroUserExists("elasticsearch_opendistro_user.test"),
					resource.TestCheckResourceAttr(
						"elasticsearch_opendistro_user.test",
						"backend_roles.#",
						"0",
					),
				),
			},
		},
	})
}

func testAccCheckElasticsearchOpendistroUserDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticsearch_opendistro_user" {
			continue
		}

		meta := testAccOpendistroProvider.Meta()

		var err error
		switch meta.(type) {
		case *elastic7.Client:
			client := meta.(*elastic7.Client)
			_, err = resourceElasticsearchGetOpendistroUser(rs.Primary.ID, client)
		default:
		}

		if err != nil {
			return nil // should be not found error
		}

		return fmt.Errorf("User %q still exists", rs.Primary.ID)
	}

	return nil
}

func testCheckElasticSearchOpendistroUserExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "elasticsearch_opendistro_user" {
				continue
			}

			meta := testAccOpendistroProvider.Meta()

			var err error
			switch meta.(type) {
			case *elastic7.Client:
				client := meta.(*elastic7.Client)
				_, err = resourceElasticsearchGetOpendistroUser(rs.Primary.ID, client)
			default:
			}

			if err != nil {
				return err
			}

			return nil
		}

		return nil
	}
}

func testCheckElasticSearchOpendistroUserConnects(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "elasticsearch_opendistro_user" {
				continue
			}

			username := rs.Primary.Attributes["username"]
			password := rs.Primary.Attributes["password"]
			meta := testAccOpendistroProvider.Meta()

			var err error
			switch meta.(type) {
			case *elastic7.Client:
				var client *elastic7.Client
				client, err = elastic7.NewClient(
					elastic7.SetURL(os.Getenv("ELASTICSEARCH_URL")),
					elastic7.SetBasicAuth(username, password))

				if err == nil {
					_, err = client.ClusterHealth().Do(context.TODO())
				}
			}

			if err != nil {
				return err
			}

			return nil
		}

		return nil
	}
}

func testAccOpendistroUserResource(resourceName string) string {
	return fmt.Sprintf(`
	resource "elasticsearch_opendistro_user" "test" {
		username      = "%s"
		password      = "passw0rd"
		description   = "test"
		backend_roles = ["some_role"]

		attributes = {
			some_attribute = "alpha"
		}
	}
	`, resourceName)
}

func testAccOpendistroUserResourceUpdated(resourceName string) string {
	return fmt.Sprintf(`
	resource "elasticsearch_opendistro_user" "test" {
		username      = "%s"
		password      = "passw0rd"
		description   = "test"
		backend_roles = ["some_role", "monitor_role"]

		attributes = {
			some_attribute  = "alpha"
			other_attribute = "beta"
		}
	}

	resource "elasticsearch_opendistro_role" "monitor_role" {
		role_name           = "monitor_role"
		cluster_permissions = ["cluster_monitor"]
	}

	resource "elasticsearch_opendistro_roles_mapping" "monitor_role" {
		role_name = "monitor_role"
	}
	`, resourceName)
}

func testAccOpendistroUserResourceMinimal(resourceName string) string {
	return fmt.Sprintf(`
	resource "elasticsearch_opendistro_user" "test" {
		username = "%s"
		password = "passw0rd"
	}
	`, resourceName)
}
