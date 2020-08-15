package es

import (
	"context"
	"errors"
	"fmt"
	"log"
	"testing"

	elastic7 "github.com/olivere/elastic/v7"
	elastic5 "gopkg.in/olivere/elastic.v5"
	elastic6 "gopkg.in/olivere/elastic.v6"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// Note the tests run with a trial license enabled, so this test is
// "destructive" in that once deactivated, a trail license may not be re-
// activated. Restarting the docker compose container doesn't seem to work.
func TestAccElasticsearchXpackLicense_Basic(t *testing.T) {
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
	case *elastic5.Client:
		allowed = false
	default:
		allowed = true
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			if testing.Short() {
				t.Skip("Skipping destructive license test because short is set")
			}
			if !allowed {
				t.Skip("License only supported on ES >= 6")
			}
		},
		Providers:    testAccXPackProviders,
		CheckDestroy: testCheckElasticsearchLicenseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testElasticsearchLicense,
				Check: resource.ComposeTestCheckFunc(
					testCheckElasticsearchLicenseExists("elasticsearch_xpack_license.test"),
				),
			},
		},
	})
}

func testCheckElasticsearchLicenseExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No license ID is set")
		}

		meta := testAccXPackProvider.Meta()

		var err error
		esClient, err := getClient(meta.(*ProviderConf))
		if err != nil {
			return err
		}
		switch client := esClient.(type) {
		case *elastic7.Client:
			var resp *elastic7.XPackInfoServiceResponse
			resp, err = client.XPackInfo().Do(context.TODO())
			log.Printf("[INFO] testCheckElasticsearchLicenseExists %+v", resp)
		case *elastic6.Client:
			var resp *elastic6.XPackInfoServiceResponse
			resp, err = client.XPackInfo().Do(context.TODO())
			log.Printf("[INFO] testCheckElasticsearchLicenseExists %+v", resp)
		default:
			return errors.New("License is only supported by elasticsearch >= v6!")
		}

		if err != nil {
			return err
		}

		return nil
	}
}

func testCheckElasticsearchLicenseDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticsearch_xpack_license" {
			continue
		}

		meta := testAccXPackProvider.Meta()

		esClient, err := getClient(meta.(*ProviderConf))
		if err != nil {
			return err
		}
		switch client := esClient.(type) {
		case *elastic7.Client:
			resp, err := client.XPackInfo().Do(context.TODO())
			log.Printf("[INFO] testCheckElasticsearchLicenseDestroy %+v", resp)

			if err != nil {
				return err
			}

			// See https://github.com/elastic/elasticsearch/pull/52407, deleting a
			// basic license is a no-op
			if resp.License.Type != "basic" && resp.License.UID != "" {
				return fmt.Errorf("License still exists")
			} else if resp.License.Type == "basic" && resp.License.UID == "" {
				return nil
			}
		case *elastic6.Client:
			resp, err := client.XPackInfo().Do(context.TODO())
			log.Printf("[INFO] testCheckElasticsearchLicenseDestroy %+v", resp)
			licenseUID := resp.License.UID

			if err != nil {
				return err
			}

			if licenseUID != "" {
				return fmt.Errorf("License still exists")
			}
		default:
			return errors.New("License is only supported by elasticsearch >= v6!")
		}

		return nil
	}

	return nil
}

var testElasticsearchLicense = `
resource "elasticsearch_xpack_license" "test" {
  use_basic_license = "true"
}
`
