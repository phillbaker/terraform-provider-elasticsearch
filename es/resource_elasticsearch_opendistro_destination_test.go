package es

import (
	"fmt"
	"testing"

	elastic7 "github.com/olivere/elastic/v7"
	elastic5 "gopkg.in/olivere/elastic.v5"
	elastic6 "gopkg.in/olivere/elastic.v6"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccElasticsearchOpenDistroDestination(t *testing.T) {
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
				t.Skip("Destinations only supported on >= ES 6")
			}
		},
		Providers:    testAccOpendistroProviders,
		CheckDestroy: testCheckElasticsearchOpenDistroDestinationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccElasticsearchOpenDistroDestination,
				Check: resource.ComposeTestCheckFunc(
					testCheckElasticsearchOpenDistroDestinationExists("elasticsearch_opendistro_destination.test_destination"),
				),
			},
		},
	})
}

func TestAccElasticsearchOpenDistroDestination_importBasic(t *testing.T) {
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
				t.Skip("Destinations only supported on >= ES 6")
			}
		},
		Providers:    testAccOpendistroProviders,
		CheckDestroy: testCheckElasticsearchOpenDistroDestinationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccElasticsearchOpenDistroDestination,
			},
			{
				ResourceName:      "elasticsearch_opendistro_destination.test_destination",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testCheckElasticsearchOpenDistroDestinationExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No destination ID is set")
		}

		meta := testAccOpendistroProvider.Meta()

		var err error
		switch client := meta.(type) {
		case *elastic7.Client:
			_, err = resourceElasticsearchOpenDistroGetDestination(rs.Primary.ID, client)
		case *elastic6.Client:
			_, err = resourceElasticsearchOpenDistroGetDestination(rs.Primary.ID, client)
		default:
		}

		if err != nil {
			return err
		}

		return nil
	}
}

func testCheckElasticsearchOpenDistroDestinationDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticsearch_opendistro_destination" {
			continue
		}

		meta := testAccOpendistroProvider.Meta()

		var err error
		switch client := meta.(type) {
		case *elastic7.Client:
			_, err = resourceElasticsearchOpenDistroGetDestination(rs.Primary.ID, client)
		case *elastic6.Client:
			_, err = resourceElasticsearchOpenDistroGetDestination(rs.Primary.ID, client)
		default:
		}

		if err != nil {
			return nil // should be not found error
		}

		return fmt.Errorf("Destination %q still exists", rs.Primary.ID)
	}

	return nil
}

var testAccElasticsearchOpenDistroDestination = `
resource "elasticsearch_opendistro_destination" "test_destination" {
  body = <<EOF
{
  "name": "my-destination",
  "type": "slack",
  "slack": {
    "url": "http://www.example.com"
  }
}
EOF
}
`
