package es

import (
	"context"
	"fmt"
	"testing"

	elastic7 "github.com/olivere/elastic/v7"
	elastic6 "gopkg.in/olivere/elastic.v6"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccElasticsearchOpenDistroRole(t *testing.T) {
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
		CheckDestroy: testAccCheckElasticsearchOpenDistroRoleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOpenDistroRoleResource(randomName),
				Check: resource.ComposeTestCheckFunc(
					testCheckElasticSearchOpenDistroRoleExists("elasticsearch_opendistro_role.test"),
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
				Config: testAccOpenDistroRoleResourceUpdated(randomName),
				Check: resource.ComposeTestCheckFunc(
					testCheckElasticSearchOpenDistroRoleExists("elasticsearch_opendistro_role.test"),
					resource.TestCheckResourceAttr(
						"elasticsearch_opendistro_role.test",
						"tenant_permissions.#",
						"2",
					),
				),
			},
			{
				Config: testAccOpenDistroRoleResourceWithoutTenantPermissions(randomName),
				Check: resource.ComposeTestCheckFunc(
					testCheckElasticSearchOpenDistroRoleExists("elasticsearch_opendistro_role.test"),
					resource.TestCheckResourceAttr(
						"elasticsearch_opendistro_role.test",
						"tenant_permissions.#",
						"0",
					),
				),
			},
			{
				Config: testAccOpenDistroRoleResourceFieldLevelSecurity(randomName),
				Check: resource.ComposeTestCheckFunc(
					testCheckElasticSearchOpenDistroRoleExists("elasticsearch_opendistro_role.test"),
					resource.TestCheckResourceAttr(
						"elasticsearch_opendistro_role.test",
						"index_permissions.#",
						"1",
					),
					resource.TestCheckTypeSetElemNestedAttrs(
						"elasticsearch_opendistro_role.test",
						"index_permissions.*",
						map[string]string{
							"field_level_security.#": "2",
						},
					),
				),
			},
		},
	})
}

func TestAccElasticsearchOpenDistroRole_importBasic(t *testing.T) {
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
		CheckDestroy: testAccCheckElasticsearchOpenDistroRoleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOpenDistroRoleResource(randomName),
			},
			{
				ResourceName:      "elasticsearch_opendistro_role.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckElasticsearchOpenDistroRoleDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticsearch_opendistro_role" {
			continue
		}

		meta := testAccOpendistroProvider.Meta()

		var err error
		esClient, err := getClient(meta.(*ProviderConf))
		if err != nil {
			return err
		}
		switch esClient.(type) {
		case *elastic7.Client:
			_, err = resourceElasticsearchGetOpenDistroRole(rs.Primary.ID, meta.(*ProviderConf))
		default:
		}

		if err != nil {
			return nil // should be not found error
		}

		return fmt.Errorf("Role %q still exists", rs.Primary.ID)
	}

	return nil
}
func testCheckElasticSearchOpenDistroRoleExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "elasticsearch_opendistro_role" {
				continue
			}

			meta := testAccOpendistroProvider.Meta()

			var err error
			esClient, err := getClient(meta.(*ProviderConf))
			if err != nil {
				return err
			}
			switch esClient.(type) {
			case *elastic7.Client:
				_, err = resourceElasticsearchGetOpenDistroRole(rs.Primary.ID, meta.(*ProviderConf))
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

func testAccOpenDistroRoleResource(resourceName string) string {
	return fmt.Sprintf(`
resource "elasticsearch_opendistro_role" "test" {
  role_name   = "%s"
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

func testAccOpenDistroRoleResourceUpdated(resourceName string) string {
	return fmt.Sprintf(`
resource "elasticsearch_opendistro_role" "test" {
  role_name   = "%s"
  description = "test"
  index_permissions {
    index_patterns = [
      "test*",
    ]

    allowed_actions = [
      "read",
    ]
  }

  index_permissions {
    index_patterns = [
      "?kibana",
    ]

    allowed_actions = [
      "indices_all",
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

func testAccOpenDistroRoleResourceWithoutTenantPermissions(resourceName string) string {
	return fmt.Sprintf(`
resource "elasticsearch_opendistro_role" "test" {
  role_name   = "%s"
  description = "test"
  index_permissions {
    index_patterns = [
      "test*",
    ]
    allowed_actions = [
      "read",
    ]
  }
  index_permissions {
    index_patterns = [
      "?kibana",
    ]
    allowed_actions = [
      "indices_all",
    ]
  }
  cluster_permissions = ["*"]
}
	`, resourceName)
}

func testAccOpenDistroRoleResourceFieldLevelSecurity(resourceName string) string {
	return fmt.Sprintf(`
resource "elasticsearch_opendistro_role" "test" {
  role_name   = "%s"
  description = "test"

  index_permissions {
    index_patterns       = ["pub*"]
    allowed_actions      = ["read"]
    field_level_security = ["fielda", "myfieldb"]
  }

  cluster_permissions = ["*"]
}
	`, resourceName)
}
