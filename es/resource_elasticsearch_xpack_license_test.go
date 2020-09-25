package es

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"testing"

	elastic7 "github.com/olivere/elastic/v7"
	elastic5 "gopkg.in/olivere/elastic.v5"
	elastic6 "gopkg.in/olivere/elastic.v6"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

// Note the tests run with a trial license enabled, so we need to ensure that
// we revert to the starting license at the end of the test
func TestAccElasticsearchLicense_Basic(t *testing.T) {
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

	license, err := resourceElasticsearchGetXpackLicense(meta)
	if err != nil {
		t.Skipf("err: %s", err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
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

	out, err := json.Marshal(license)
	if err != nil {
		t.Fatalf("err %s", err)
	}
	_, err = resourceElasticsearchPutEnterpriseLicense(string(out), meta)
	if err != nil {
		t.Fatalf("err %s", err)
	}
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
		switch client := meta.(type) {
		case *elastic7.Client:
			var resp *elastic7.XPackInfoServiceResponse
			resp, err = client.XPackInfo().Do(context.TODO())
			log.Printf("[INFO] %+v", resp)
		case *elastic6.Client:
			var resp *elastic6.XPackInfoServiceResponse
			resp, err = client.XPackInfo().Do(context.TODO())
			log.Printf("[INFO] %+v", resp)
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

		// switch meta.(type) {
		// case *elastic7.Client:
		// 	client := meta.(*elastic7.Client)
		// 	res, err := client.API.License.Get(
		// 		client.API.License.Get.WithContext(context.Background()),
		// 		client.API.License.Get.WithPretty(),
		// 	)
		// 	if err != nil {
		// 		return err
		// 	}
		// 	defer res.Body.Close()
		// 	if res.IsError() {
		// 		if res.StatusCode == 404 {

		// 			err = forceBasicLicense()
		// 			if err != nil {
		// 				return errors.New("Error when enabled trial license for other tests. You need to check by your hand")
		// 			}

		// 			return nil
		// 		} else {
		// 			return err
		// 		}
		// 	}
		// default:
		// 	return errors.New("License is only supported by the elastic library >= v6!")
		// }

		var err error
		switch client := meta.(type) {
		case *elastic7.Client:
			var resp *elastic7.XPackInfoServiceResponse
			resp, err = client.XPackInfo().Do(context.TODO())
			log.Printf("[INFO] %+v", resp)
		case *elastic6.Client:
			var resp *elastic6.XPackInfoServiceResponse
			resp, err = client.XPackInfo().Do(context.TODO())
			log.Printf("[INFO] %+v", resp)
		default:
			return errors.New("License is only supported by elasticsearch >= v6!")
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("License still exists")
	}

	return nil
}

// func forceBasicLicense() error {
// 	meta := testAccProvider.Meta()

// 	switch meta.(type) {
// 	case *elastic7.Client:
// 		client := meta.(*elastic7.Client)
// 		res, err := client.API.License.PostStartBasic(
// 			client.API.License.PostStartBasic.WithContext(context.Background()),
// 			client.API.License.PostStartBasic.WithPretty(),
// 			client.API.License.PostStartBasic.WithAcknowledge(true),
// 		)

// 		if err != nil {
// 			return err
// 		}

// 		if res.IsError() {
// 			return errors.New("Error when enabled basic license")
// 		}
// 	}

// 	return nil
// }

var testElasticsearchLicense = `
resource "elasticsearch_xpack_license" "test" {
  use_basic_license = "true"
}
`
