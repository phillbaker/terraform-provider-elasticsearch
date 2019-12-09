package main

import (
	"context"
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

func TestAccElasticsearchXpackRoleMapping(t *testing.T) {

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
	default:
		allowed = true
	}

	randomName := "test" + acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			if !allowed {
				t.Skip("Role Mapping only supported on ES >= 6")
			}
		},
		Providers:    testAccXPackProviders,
		CheckDestroy: testAccCheckRoleMappingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoleMappingResource(randomName),
				Check: resource.ComposeTestCheckFunc(
					testCheckRoleMappingExists("elasticsearch_xpack_role_mapping.test"),
					resource.TestCheckResourceAttr(
						"elasticsearch_xpack_role_mapping.test",
						"id",
						randomName,
					),
					resource.TestCheckResourceAttr(
						"elasticsearch_xpack_role_mapping.test",
						"enabled",
						"true",
					),
					resource.TestCheckResourceAttr(
						"elasticsearch_xpack_role_mapping.test",
						"roles.#",
						"2",
					),
					resource.TestCheckResourceAttr(
						"elasticsearch_xpack_role_mapping.test",
						"rules",
						`{"any":[{"field":{"username":"esadmin"}},{"field":{"groups":"cn=admins,dc=example,dc=com"}}]}`,
					),
					resource.TestCheckResourceAttr(
						"elasticsearch_xpack_role_mapping.test",
						"metadata",
						`{}`,
					),
				),
			},
			{
				Config: testAccRoleMappingResource_Updated(randomName),
				Check: resource.ComposeTestCheckFunc(
					testCheckRoleMappingExists("elasticsearch_xpack_role_mapping.test"),
					resource.TestCheckResourceAttr(
						"elasticsearch_xpack_role_mapping.test",
						"roles.#",
						"3",
					),
					resource.TestCheckResourceAttr(
						"elasticsearch_xpack_role_mapping.test",
						"rules",
						`{"any":[{"field":{"username":"esadmin"}},{"field":{"groups":"cn=admins,dc=corp,dc=example,dc=com"}}]}`,
					),
				),
			},
		},
	})
}

func testAccCheckRoleMappingDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticsearch_xpack_role_mapping" {
			continue
		}

		meta := testAccXPackProvider.Meta()

		if client, ok := meta.(*elastic7.Client); ok {
			if _, err := client.XPackSecurityGetRoleMapping(rs.Primary.ID).Do(context.TODO()); err != nil {
				if elasticErr, ok := err.(*elastic7.Error); ok && elasticErr.Status == 404 {
					return nil
				} else {
					return fmt.Errorf("Role mapping %q still exists", rs.Primary.ID)
				}
			} else {
				return err
			}

		} else if client, ok := meta.(*elastic6.Client); ok {
			if _, err := client.XPackSecurityGetRoleMapping(rs.Primary.ID).Do(context.TODO()); err != nil {
				if elasticErr, ok := err.(*elastic6.Error); ok && elasticErr.Status == 404 {
					return nil
				} else {
					return fmt.Errorf("Role mapping %q still exists", rs.Primary.ID)
				}
			} else {
				return err
			}

		} else {
			return fmt.Errorf("Unsupported client type : %v", meta)
		}
	}
	return nil
}

func testCheckRoleMappingExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No role mapping ID is set")
		}

		meta := testAccXPackProvider.Meta()
		var err error
		if client, ok := meta.(*elastic7.Client); ok {
			_, err = client.XPackSecurityGetRoleMapping(rs.Primary.ID).Do(context.TODO())
		} else {
			client := meta.(*elastic6.Client)
			_, err = client.XPackSecurityGetRoleMapping(rs.Primary.ID).Do(context.TODO())
		}

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccRoleMappingResource(resourceName string) string {
	return fmt.Sprintf(`
resource "elasticsearch_xpack_role_mapping" "test" {
  role_mapping_name = "%s"
  roles = [
    "admin",
    "user",
  ]
  rules = <<-EOF
  {
    "any": [
      {
        "field": {
          "username": "esadmin"
        }
      },
      {
        "field": {
          "groups": "cn=admins,dc=example,dc=com"
        }
      }
    ]
  }
  EOF
  enabled = true
}
`, resourceName)
}

func testAccRoleMappingResource_Updated(resourceName string) string {
	return fmt.Sprintf(`
resource "elasticsearch_xpack_role_mapping" "test" {
  role_mapping_name = "%s"
  roles = [
    "admin",
	"user",
	"guest",
  ]
  rules = <<-EOF
  {
    "any": [
      {
        "field": {
          "username": "esadmin"
        }
      },
      {
        "field": {
          "groups": "cn=admins,dc=corp,dc=example,dc=com"
        }
      }
    ]
  }
  EOF
  enabled = false
}
`, resourceName)
}
