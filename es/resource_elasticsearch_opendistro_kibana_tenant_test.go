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

func TestAccElasticsearchOpenDistroKibanaTenant(t *testing.T) {
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
				t.Skip("Allowed only for ES >= 7")
			}
		},
		Providers:    testAccOpendistroProviders,
		CheckDestroy: testAccCheckElasticsearchOpenDistroKibanaTenantDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOpenDistroKibanaTenantResource(randomName),
				Check: resource.ComposeTestCheckFunc(
					testCheckElasticSearchOpenDistroKibanaTenantExists("elasticsearch_opendistro_kibana_tenant.test"),
					resource.TestCheckResourceAttr(
						"elasticsearch_opendistro_kibana_tenant.test",
						"id",
						randomName,
					),
					resource.TestCheckResourceAttr(
						"elasticsearch_opendistro_kibana_tenant.test",
						"description",
						"test",
					),
				),
			},
			{
				Config: testAccOpenDistroKibanaTenantResourceUpdated(randomName),
				Check: resource.ComposeTestCheckFunc(
					testCheckElasticSearchOpenDistroKibanaTenantExists("elasticsearch_opendistro_kibana_tenant.test"),
					resource.TestCheckResourceAttr(
						"elasticsearch_opendistro_kibana_tenant.test",
						"description",
						"test2",
					),
				),
			},
		},
	})
}

func testAccCheckElasticsearchOpenDistroKibanaTenantDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticsearch_opendistro_kibana_tenant" {
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
			_, err = resourceElasticsearchGetOpenDistroKibanaTenant(rs.Primary.ID, meta.(*ProviderConf))
		default:
		}

		if err != nil {
			return nil // should be not found error
		}

		return fmt.Errorf("KibanaTenant %q still exists", rs.Primary.ID)
	}

	return nil
}
func testCheckElasticSearchOpenDistroKibanaTenantExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "elasticsearch_opendistro_kibana_tenant" {
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
				_, err = resourceElasticsearchGetOpenDistroKibanaTenant(rs.Primary.ID, meta.(*ProviderConf))
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

func testAccOpenDistroKibanaTenantResource(resourceName string) string {
	return fmt.Sprintf(`
resource "elasticsearch_opendistro_kibana_tenant" "test" {
  tenant_name = "%s"
  description = "test"
}
	`, resourceName)
}

func testAccOpenDistroKibanaTenantResourceUpdated(resourceName string) string {
	return fmt.Sprintf(`
resource "elasticsearch_opendistro_kibana_tenant" "test" {
  tenant_name = "%s"
  description = "test2"
}
	`, resourceName)
}
