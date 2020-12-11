package es

import (
	"context"
	"fmt"
	"testing"

	elastic7 "github.com/olivere/elastic/v7"
	elastic5 "gopkg.in/olivere/elastic.v5"
	elastic6 "gopkg.in/olivere/elastic.v6"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccElasticsearchWatch(t *testing.T) {
	provider := Provider().(*schema.Provider)
	err := provider.Configure(&terraform.ResourceConfig{})
	if err != nil {
		t.Skipf("err: %s", err)
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
			if !allowed {
				t.Skip("Watches only supported on ES >= 6")
			}
		},
		Providers:    testAccXPackProviders,
		CheckDestroy: testCheckElasticsearchWatchDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccElasticsearchWatch,
				Check: resource.ComposeTestCheckFunc(
					testCheckElasticsearchWatchExists("elasticsearch_xpack_watch.test_watch"),
				),
			},
		},
	})
}

func testCheckElasticsearchWatchExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No watch ID is set")
		}

		meta := testAccXPackProvider.Meta()

		var err error
		esClient, err := getClient(meta.(*ProviderConf))
		if err != nil {
			return err
		}
		switch client := esClient.(type) {
		case *elastic7.Client:
			_, err = client.XPackWatchGet("my_watch").Do(context.TODO())
		case *elastic6.Client:
			_, err = client.XPackWatchGet("my_watch").Do(context.TODO())
		default:
		}

		if err != nil {
			return err
		}

		return nil
	}
}

func testCheckElasticsearchWatchDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticsearch_xpack_watch" {
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
			_, err = client.XPackWatchGet("my_watch").Do(context.TODO())
		case *elastic6.Client:
			_, err = client.XPackWatchGet("my_watch").Do(context.TODO())
		default:
		}

		if err != nil {
			return nil // should be not found error
		}

		return fmt.Errorf("Watch %q still exists", rs.Primary.ID)
	}

	return nil
}

var testAccElasticsearchWatch = `
resource "elasticsearch_xpack_watch" "test_watch" {
  watch_id = "my_watch"
  body = <<EOF
{
  "input": {
    "simple": {
      "payload": {
        "send": "yes"
      }
    }
  },
  "condition": {
    "always": {}
  },
  "trigger": {
    "schedule": {
      "hourly": {
        "minute": [0, 5]
      }
    }
  },
  "actions": {
    "test_log": {
	    "logging": {
	    	"level": "info",
	      "text": "executed at {{ctx.execution_time}}"
	    }
	  }
  }
}
EOF
}
`
