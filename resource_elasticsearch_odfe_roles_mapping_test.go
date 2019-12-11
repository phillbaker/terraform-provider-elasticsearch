package main

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

func TestAccElasticsearchOdfeRolesMapping(t *testing.T) {

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
		CheckDestroy: testAccCheckElasticsearchOdfeRolesMappingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOdfeRolesMappingResource(randomName),
				Check: resource.ComposeTestCheckFunc(
					testCheckElasticSearchOdfeRolesMappingExists("elasticsearch_odfe_roles_mapping.test"),
					resource.TestCheckResourceAttr(
						"elasticsearch_odfe_roles_mapping.test",
						"id",
						randomName,
					),
					resource.TestCheckResourceAttr(
						"elasticsearch_odfe_roles_mapping.test",
						"backend_roles.#",
						"1",
					),
					resource.TestCheckResourceAttr(
						"elasticsearch_odfe_roles_mapping.test",
						"description",
						"test",
					),
				),
			},
			{
				Config: testAccOdfeRoleMappingResourceUpdated(randomName),
				Check: resource.ComposeTestCheckFunc(
					testCheckElasticSearchOdfeRolesMappingExists("elasticsearch_odfe_roles_mapping.test"),
					resource.TestCheckResourceAttr(
						"elasticsearch_odfe_roles_mapping.test",
						"backend_roles.#",
						"2",
					),
				),
			},
		},
	})
}

func testAccCheckElasticsearchOdfeRolesMappingDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticsearch_odfe_roles_mappings_mapping" {
			continue
		}

		meta := testAccOpendistroProvider.Meta()

		var err error
		switch meta.(type) {
		case *elastic7.Client:
			client := meta.(*elastic7.Client)
			_, err = resourceElasticsearchGetOdfeRolesMapping(rs.Primary.ID, client)
		default:
		}

		if err != nil {
			return nil // should be not found error
		}

		return fmt.Errorf("Role %q still exists", rs.Primary.ID)
	}

	return nil
}
func testCheckElasticSearchOdfeRolesMappingExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "elasticsearch_odfe_roles_mapping" {
				continue
			}

			meta := testAccOpendistroProvider.Meta()

			var err error
			switch meta.(type) {
			case *elastic7.Client:
				client := meta.(*elastic7.Client)
				_, err = resourceElasticsearchGetOdfeRolesMapping(rs.Primary.ID, client)
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

func testAccOdfeRolesMappingResource(resourceName string) string {
	return fmt.Sprintf(` 
	resource "elasticsearch_odfe_roles_mapping" "test" {
		role_name = "%s"
		backend_roles = [
			"active_directory",
		]
		
		description = "test"
	}
	`, resourceName)
}

func testAccOdfeRoleMappingResourceUpdated(resourceName string) string {
	return fmt.Sprintf(` 
	resource "elasticsearch_odfe_roles_mapping" "test" {
		role_name = "%s"
		backend_roles = [
			"active_directory",
			"ldap",
		]
		
		description = "test_update"
	}
	`, resourceName)
}
