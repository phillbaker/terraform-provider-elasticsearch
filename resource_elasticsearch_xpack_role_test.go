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

func TestAccElasticsearchXpackRole(t *testing.T) {
	
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
				t.Skip("Roles only supported on ES >= 6")
			}
		},
		Providers:    testAccXPackProviders,
		CheckDestroy: testAccCheckRoleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoleResource(randomName),
				Check: resource.ComposeTestCheckFunc(
					testCheckRoleExists("elasticsearch_xpack_role.test"),
					resource.TestCheckResourceAttr(
						"elasticsearch_xpack_role.test",
						"id",
						randomName,
					),
					resource.TestCheckResourceAttr(
						"elasticsearch_xpack_role.test",
						"cluster.#",
						"1",
					),
					resource.TestCheckResourceAttr(
						"elasticsearch_xpack_role.test",
						"metadata",
						"{}",
					),
				),
			},
			{
				Config: testAccRoleResource_Updated(randomName),
				Check: resource.ComposeTestCheckFunc(
					testCheckRoleExists("elasticsearch_xpack_role.test"),
					resource.TestCheckResourceAttr(
						"elasticsearch_xpack_role.test",
						"metadata",
						`{"foo":"bar"}`,
					),
				),
			},
			{
				Config: testAccRoleResource_Global(randomName),
				Check: resource.ComposeTestCheckFunc(
					testCheckRoleExists("elasticsearch_xpack_role.test"),
					resource.TestCheckResourceAttr(
						"elasticsearch_xpack_role.test",
						"global",
						`{"application":{"manage":{"applications":["testapp"]}}}`,
					),
				),
			},
		},
	})
}

func testAccCheckRoleDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticsearch_xpack_role" {
			continue
		}
		meta := testAccXPackProvider.Meta()

		if client, ok := meta.(*elastic7.Client); ok {
			if _, err := client.XPackSecurityGetRole(rs.Primary.ID).Do(context.TODO()); err != nil {
				if elasticErr, ok := err.(*elastic7.Error); ok && elasticErr.Status == 404 {
					return nil
				} else {
					return fmt.Errorf("Role %q still exists", rs.Primary.ID)
				}
			} else {
				return err
			}

		} else if client, ok := meta.(*elastic6.Client); ok {
			if _, err := client.XPackSecurityGetRole(rs.Primary.ID).Do(context.TODO()); err != nil {
				if elasticErr, ok := err.(*elastic6.Error); ok && elasticErr.Status == 404 {
					return nil
				} else {
					return fmt.Errorf("Role %q still exists", rs.Primary.ID)
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
func testCheckRoleExists(name string) resource.TestCheckFunc {
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
			_, err = client.XPackSecurityGetRole(rs.Primary.ID).Do(context.TODO())
		} else {
			client := meta.(*elastic6.Client)
			_, err = client.XPackSecurityGetRole(rs.Primary.ID).Do(context.TODO())
		}

		if err != nil {
			return err
		}

		return nil
	}
}
			
func testAccRoleResource(resourceName string) string {
	return fmt.Sprintf(` 
	resource "elasticsearch_xpack_role" "test" {
		role_name = "%s"
		indices {
			names 	   = ["testIndice"]
			privileges = ["testPrivilege"]
		}
		indices {
			names 	   = ["testIndice2"]
			privileges = ["testPrivilege2"]
		}
		cluster = [
		"all"
		]
		applications {
			application = "testapp"
			privileges = [ 
			"admin", 
			"read" 
			]
			resources = [ 
			"*" 
			]
		}
	}
	`, resourceName)
}
			
func testAccRoleResource_Updated(resourceName string) string {
	return fmt.Sprintf(`
	resource "elasticsearch_xpack_role" "test" {
		role_name = "%s"
		indices {
			names 	 = ["testIndice"]
			privileges = ["testPrivilege"]
		}
		indices {
			names 	 = ["testIndice2"]
			privileges = ["testPrivilege2"]
		}
		cluster = [
		"all"
		]
		applications {
			application = "testapp"
			privileges = [ 
			"admin", 
			"read",
			"delete", 
			]
			resources = [ 
			"*" 
			]
		}
		metadata = <<-EOF
		{
			"foo": "bar"
		}
		EOF
	}
	`, resourceName)
}

func testAccRoleResource_Global(resourceName string) string {
	return fmt.Sprintf(`
	resource "elasticsearch_xpack_role" "test" {
		role_name = "%s"
		indices {
			names 	 = ["testIndice"]
			privileges = ["testPrivilege"]
		}
		indices {
			names 	 = ["testIndice2"]
			privileges = ["testPrivilege2"]
		}
		cluster = [
		"all",
		]
		applications {
			application = "testapp"
			privileges = [ 
			"admin", 
			"read",
			"delete", 
			]
			resources = [ 
			"*" ,
			]
		}
		
		metadata = <<-EOF
		{
			"foo": "bar"
		}
		EOF
		
		
		global = <<-EOF
		{
			"application": {
				"manage": {    
					"applications": ["testapp"] 
				}
			}
		}
		EOF
	}
	`, resourceName)
}
