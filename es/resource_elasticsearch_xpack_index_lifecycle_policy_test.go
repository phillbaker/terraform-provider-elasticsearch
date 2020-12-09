package es

import (
	"context"
	"errors"
	"fmt"
	"testing"

	elastic7 "github.com/olivere/elastic/v7"
	elastic5 "gopkg.in/olivere/elastic.v5"
	elastic6 "gopkg.in/olivere/elastic.v6"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccElasticsearchXpackIndexLifecyclePolicy(t *testing.T) {
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

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			if !allowed {
				t.Skip("Index lifecycles only supported on ES >= 6")
			}
		},
		Providers:    testAccXPackProviders,
		CheckDestroy: testCheckElasticsearchXpackIndexLifecyclePolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccElasticsearchXpackIndexLifecyclePolicy,
				Check: resource.ComposeTestCheckFunc(
					testCheckElasticsearchXpackIndexLifecyclePolicyExists("elasticsearch_xpack_index_lifecycle_policy.test"),
				),
			},
		},
	})
}

func TestAccElasticsearchXpackIndexLifecyclePolicy_importBasic(t *testing.T) {
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

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			if !allowed {
				t.Skip("Index lifecycles only supported on ES >= 6")
			}
		},
		Providers:    testAccXPackProviders,
		CheckDestroy: testCheckElasticsearchXpackIndexLifecyclePolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccElasticsearchXpackIndexLifecyclePolicy,
			},
			{
				ResourceName:      "elasticsearch_xpack_index_lifecycle_policy.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testCheckElasticsearchXpackIndexLifecyclePolicyExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No index lifecycle policy ID is set")
		}

		meta := testAccXPackProvider.Meta()

		var err error
		esClient, err := getClient(meta.(*ProviderConf))
		if err != nil {
			return err
		}
		switch client := esClient.(type) {
		case *elastic7.Client:
			_, err = client.XPackIlmGetLifecycle().Policy(rs.Primary.ID).Do(context.TODO())
		case *elastic6.Client:
			_, err = client.XPackIlmGetLifecycle().Policy(rs.Primary.ID).Do(context.TODO())
		default:
			err = errors.New("Index Lifecycle Management is only supported by the elastic library >= v6!")
		}

		if err != nil {
			return err
		}

		return nil
	}
}

func testCheckElasticsearchXpackIndexLifecyclePolicyDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticsearch_xpack_index_lifecycle_policy" {
			continue
		}

		meta := testAccXPackProvider.Meta()

		var err error
		esClient, err := getClient(meta.(*ProviderConf))
		if err != nil {
			return err
		}
		switch client := esClient.(type) {
		case *elastic7.Client:
			_, err = client.XPackIlmGetLifecycle().Policy(rs.Primary.ID).Do(context.TODO())
		case *elastic6.Client:
			_, err = client.XPackIlmGetLifecycle().Policy(rs.Primary.ID).Do(context.TODO())
		default:
			err = errors.New("Index Lifecycle Management is only supported by the elastic library >= v6!")
		}

		if err != nil {
			return nil // should be not found error
		}

		return fmt.Errorf("Index lifecycle policy %q still exists", rs.Primary.ID)
	}

	return nil
}

var testAccElasticsearchXpackIndexLifecyclePolicy = `
resource "elasticsearch_xpack_index_lifecycle_policy" "test" {
  name = "terraform-test"
  body = <<EOF
{
  "policy": {
    "phases": {
      "warm": {
        "min_age": "10d",
        "actions": {
          "forcemerge": {
            "max_num_segments": 1
          }
        }
      },
      "delete": {
        "min_age": "30d",
        "actions": {
          "delete": {}
        }
      }
    }
  }
}
EOF
}
`
