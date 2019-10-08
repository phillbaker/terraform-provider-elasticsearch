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

func TestAccElasticsearchXpackUser(t *testing.T) {
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
				t.Skip("Users only supported on ES >= 6")
			}
		},
		Providers:    testAccXPackProviders,
		CheckDestroy: testAccCheckUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserResource(randomName),
				Check: resource.ComposeTestCheckFunc(
					testCheckUserExists("elasticsearch_xpack_user.test"),
					resource.TestCheckResourceAttr(
						"elasticsearch_xpack_user.test",
						"id",
						randomName,
					),
					resource.TestCheckResourceAttr(
						"elasticsearch_xpack_user.test",
						"roles.#",
						"1",
					),
					resource.TestCheckResourceAttr(
						"elasticsearch_xpack_user.test",
						"metadata",
						"{}",
					),
				),
			},
			{
				Config: testAccUserResource_Updated(randomName),
				Check: resource.ComposeTestCheckFunc(
					testCheckUserExists("elasticsearch_xpack_user.test"),
					resource.TestCheckResourceAttr(
						"elasticsearch_xpack_user.test",
						"metadata",
						`{"foo":"bar"}`,
					),
				),
			},
			{
				Config: testAccUserResource_Global(randomName),
				Check: resource.ComposeTestCheckFunc(
					testCheckUserExists("elasticsearch_xpack_user.test"),
					resource.TestCheckResourceAttr(
						"elasticsearch_xpack_user.test",
						"username",
						randomName,
					),
				),
			},
		},
	})
}

func testAccCheckUserDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticsearch_xpack_user" {
			continue
		}

		meta := testAccXPackProvider.Meta()

		if client, ok := meta.(*elastic7.Client); ok {
			if _, err := client.XPackSecurityGetUser(rs.Primary.ID).Do(context.TODO()); err != nil {
				if elasticErr, ok := err.(*elastic7.Error); ok && elasticErr.Status == 404 {
					return nil
				} else {
					return fmt.Errorf("User %q still exists", rs.Primary.ID)
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

func testCheckUserExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No user mapping ID is set")
		}

		meta := testAccXPackProvider.Meta()

		var err error
		if client, ok := meta.(*elastic7.Client); ok {
			_, err = client.XPackSecurityGetUser(rs.Primary.ID).Do(context.TODO())
		}
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccUserResource(resourceName string) string {
	return fmt.Sprintf(`
resource "elasticsearch_xpack_user" "test" {
	username = "%s"
	fullname = "John Do"
	email    = "john@do.com"
	password = "secret"
	roles    = ["admin"]
}
`, resourceName)
}

func testAccUserResource_Updated(resourceName string) string {
	return fmt.Sprintf(`
resource "elasticsearch_xpack_user" "test" {
	username = "%s"
	fullname = "John DoDo"
	email    = "john@do.com"
	password = "secret"
	roles    = ["admin"]
  metadata = <<-EOF
  {
    "foo": "bar"
  }
  EOF
}
`, resourceName)
}

func testAccUserResource_Global(resourceName string) string {
	return fmt.Sprintf(`
resource "elasticsearch_xpack_user" "test" {
	username = "%s"
	password = "secret"
	roles    = []
}
`, resourceName)
}
