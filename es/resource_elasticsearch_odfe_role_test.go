package es

import (
	"fmt"
	"testing"

	elastic7 "github.com/olivere/elastic/v7"
	elastic5 "gopkg.in/olivere/elastic.v5"
	elastic6 "gopkg.in/olivere/elastic.v6"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccElasticsearchOpendistroRole(t *testing.T) {

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
				t.Skip("Roles only supported on ES >= 7")
			}
		},
		Providers:    testAccOpendistroProviders,
		CheckDestroy: testAccCheckElasticsearchOpendistroRoleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOpendistroRoleResource(randomName),
				Check: resource.ComposeTestCheckFunc(
					testCheckElasticSearchOpendistroRoleExists("elasticsearch_opendistro_role.test"),
					resource.TestCheckResourceAttr(
						"elasticsearch_opendistro_role.test",
						"id",
						randomName,
					),
					resource.TestCheckResourceAttr(
						"elasticsearch_opendistro_role.test",
						"cluster_permissions.#",
						"1",
					),
					resource.TestCheckResourceAttr(
						"elasticsearch_opendistro_role.test",
						"tenant_permissions.#",
						"1",
					),
					resource.TestCheckResourceAttr(
						"elasticsearch_opendistro_role.test",
						"index_permissions.#",
						"1",
					),
					resource.TestCheckResourceAttr(
						"elasticsearch_opendistro_role.test",
						"description",
						"test",
					),
				),
			},
			{
				Config: testAccOpendistroRoleResourceUpdated(randomName),
				Check: resource.ComposeTestCheckFunc(
					testCheckElasticSearchOpendistroRoleExists("elasticsearch_opendistro_role.test"),
					resource.TestCheckResourceAttr(
						"elasticsearch_opendistro_role.test",
						"tenant_permissions.#",
						"2",
					),
				),
			},
		},
	})
}

func testAccCheckElasticsearchOpendistroRoleDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticsearch_opendistro_role" {
			continue
		}

		meta := testAccOpendistroProvider.Meta()

		var err error
		switch meta.(type) {
		case *elastic7.Client:
			client := meta.(*elastic7.Client)
			_, err = resourceElasticsearchGetOpendistroRole(rs.Primary.ID, client)
		default:
		}

		if err != nil {
			return nil // should be not found error
		}

		return fmt.Errorf("Role %q still exists", rs.Primary.ID)
	}

	return nil
}
func testCheckElasticSearchOpendistroRoleExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "elasticsearch_opendistro_role" {
				continue
			}

			meta := testAccOpendistroProvider.Meta()

			var err error
			switch meta.(type) {
			case *elastic7.Client:
				client := meta.(*elastic7.Client)
				_, err = resourceElasticsearchGetOpendistroRole(rs.Primary.ID, client)
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

func testAccOpendistroRoleResource(resourceName string) string {
	return fmt.Sprintf(` 
	resource "elasticsearch_opendistro_role" "test" {
		role_name = "%s"
		description = "test"
		index_permissions {
			index_patterns = [
				"*",
			]
		
			allowed_actions = [
				"*",
			]
		}
		
		tenant_permissions {
			tenant_patterns = [
				"*",
			]
		
			allowed_actions = [
				"kibana_all_write",
			]
		}
	
		cluster_permissions = ["*"]
	}
	`, resourceName)
}

func testAccOpendistroRoleResourceUpdated(resourceName string) string {
	return fmt.Sprintf(` 
	resource "elasticsearch_opendistro_role" "test" {
		role_name = "%s"
		description = "test"
		index_permissions {
			index_patterns = [
				"*",
			]
		
			allowed_actions = [
				"*",
			]
		}
		
		tenant_permissions {
			tenant_patterns = [
				"*",
			]
		
			allowed_actions = [
				"kibana_all_write",
			]
		}

		tenant_permissions {
			tenant_patterns = [
				"test*",
			]
		
			allowed_actions = [
				"kibana_all_write",
			]
		}
	
		cluster_permissions = ["*"]
	}
	`, resourceName)
}
