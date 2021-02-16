package es

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"

	elastic7 "github.com/olivere/elastic/v7"
	elastic5 "gopkg.in/olivere/elastic.v5"
	elastic6 "gopkg.in/olivere/elastic.v6"
)

const (
	testAccElasticsearchIndex = `
resource "elasticsearch_index" "test" {
  name = "terraform-test"
  number_of_shards = 1
  number_of_replicas = 1
}
`
	testAccElasticsearchIndexUpdate1 = `
resource "elasticsearch_index" "test" {
  name = "terraform-test"
  number_of_shards = 1
  number_of_replicas = 2
  routing_partition_size = 1
  refresh_interval = "10s"
	max_result_window = 1000
	max_refresh_listeners = 10
	blocks_read_only = false
	blocks_read = false
	blocks_write = false
	blocks_metadata = false
	search_slowlog_threshold_query_warn = "5s"
	search_slowlog_threshold_fetch_warn = "5s"
	search_slowlog_level = "warn"
	indexing_slowlog_threshold_index_warn = "5s"
	indexing_slowlog_level = "warn"
}
`
	testAccElasticsearchIndexInvalid = `
resource "elasticsearch_index" "test" {
  name = "terraform-test"
  number_of_shards = 1
  number_of_replicas = 1
  mappings = <<EOF
{
  "people": {
    "_all": {
      "enabled": "true"
    },
    "properties": {
      "email": {
        "type": "text"
      }
    }
  }
}
EOF
}
`
	testAccElasticsearchIndexUpdateForceDestroy = `
resource "elasticsearch_index" "test" {
  name = "terraform-test"
  number_of_shards = 1
  number_of_replicas = 2
  force_destroy = true
}
`
	testAccElasticsearchIndexDateMath = `
resource "elasticsearch_index" "test_date_math" {
  name = "<terraform-test-{now/y{yyyy}}-000001>"
  # name = "%3Ctest-%7Bnow%2Fy%7Byyyy%7D%7D-000001%3E"
  number_of_shards = 1
  number_of_replicas = 1
}
`
	testAccElasticsearchIndexRolloverAliasXpack = `
resource "elasticsearch_index_lifecycle_policy" "test" {
  name = "terraform-test"
  body = <<EOF
{
  "policy": {
    "phases": {
      "hot": {
        "min_age": "0ms",
        "actions": {
          "rollover": {
            "max_size": "50gb"
          }
        }
      }
    }
  }
}
EOF
}

resource "elasticsearch_index_template" "test" {
  name = "terraform-test"
  body = <<EOF
{
  "index_patterns": ["terraform-test-*"],
  "settings": {
    "index": {
      "lifecycle": {
        "name": "${elasticsearch_index_lifecycle_policy.test.name}",
        "rollover_alias": "terraform-test"
      }
    }
  }
}
EOF
}

resource "elasticsearch_index" "test" {
  name = "terraform-test-000001"
  number_of_shards = 1
  number_of_replicas = 1
  aliases = jsonencode({
    "terraform-test" = {
      "is_write_index" = true
    }
  })

  depends_on = [elasticsearch_index_template.test]
}
`
	testAccElasticsearchIndexRolloverAliasOpendistro = `
resource elasticsearch_opendistro_ism_policy "test" {
  policy_id = "test"
  body      = <<EOF
{
  "policy": {
    "description": "Terraform Test",
    "default_state": "hot",
    "states": [
      {
        "name": "hot",
        "actions": [
          {
            "rollover": {
              "min_size": "50gb"
            }
          }
        ],
        "transitions": []
      }
    ]
  }
}
  EOF
}

resource "elasticsearch_index_template" "test" {
  name = "terraform-test"
  body = <<EOF
{
  "index_patterns": ["terraform-test-*"],
  "settings": {
    "index": {
      "opendistro": {
        "index_state_management": {
          "policy_id": "${elasticsearch_opendistro_ism_policy.test.policy_id}",
          "rollover_alias": "terraform-test"
        }
      }
    }
  }
}
EOF
}

resource "elasticsearch_index" "test" {
  name = "terraform-test-000001"
  number_of_shards = 1
  number_of_replicas = 1
  aliases = jsonencode({
    "terraform-test" = {
      "is_write_index" = true
    }
  })

  depends_on = [elasticsearch_index_template.test]
}
`
)

func TestAccElasticsearchIndex(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: checkElasticsearchIndexDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccElasticsearchIndex,
				Check: resource.ComposeTestCheckFunc(
					checkElasticsearchIndexExists("elasticsearch_index.test"),
				),
			},
			{
				Config: testAccElasticsearchIndexUpdate1,
				Check: resource.ComposeTestCheckFunc(
					checkElasticsearchIndexUpdated("elasticsearch_index.test"),
				),
			},
			{
				Config: testAccElasticsearchIndexUpdateForceDestroy,
				Check: resource.ComposeTestCheckFunc(
					checkElasticsearchIndexUpdated("elasticsearch_index.test"),
				),
			},
		},
	})
}

func TestAccElasticsearchIndex_handleInvalid(t *testing.T) {
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
				t.Skip("Only tested on ES >= 6")
			}
		},
		Providers:    testAccProviders,
		CheckDestroy: checkElasticsearchIndexDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccElasticsearchIndexInvalid,
				ExpectError: regexp.MustCompile("Failed to parse mapping"),
			},
		},
	})
}

func TestAccElasticsearchIndex_importBasic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: checkElasticsearchIndexDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccElasticsearchIndex,
			},
			{
				ResourceName:      "elasticsearch_index.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					// not returned from the API
					"force_destroy",
				},
			},
		},
	})
}

func TestAccElasticsearchIndex_dateMath(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: checkElasticsearchIndexDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccElasticsearchIndexDateMath,
				Check: resource.ComposeTestCheckFunc(
					checkElasticsearchIndexExists("elasticsearch_index.test_date_math"),
				),
			},
		},
	})
}

func TestAccElasticsearchIndex_rolloverAliasXpack(t *testing.T) {
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
				t.Skip("Index lifecycles only supported on ES >= 6")
			}
		},
		Providers:    testAccXPackProviders,
		CheckDestroy: checkElasticsearchIndexRolloverAliasDestroy(testAccXPackProvider, "terraform-test"),
		Steps: []resource.TestStep{
			{
				Config: testAccElasticsearchIndexRolloverAliasXpack,
				Check: resource.ComposeTestCheckFunc(
					checkElasticsearchIndexRolloverAliasExists(testAccXPackProvider, "terraform-test"),
				),
			},
			{
				ResourceName:      "elasticsearch_index.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"aliases",       // not handled by this provider
					"force_destroy", // not returned from the API
				},
				ImportStateCheck: checkElasticsearchIndexRolloverAliasState("terraform-test"),
			},
		},
	})
}

func TestAccElasticsearchIndex_rolloverAliasOpendistro(t *testing.T) {
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
	case *elastic6.Client:
		allowed = false
	case *elastic5.Client:
		allowed = false
	default:
		allowed = true
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			if !allowed {
				t.Skip("Opendistro index policies only supported on ES 7")
			}
		},
		Providers:    testAccOpendistroProviders,
		CheckDestroy: checkElasticsearchIndexRolloverAliasDestroy(testAccOpendistroProvider, "terraform-test"),
		Steps: []resource.TestStep{
			{
				Config: testAccElasticsearchIndexRolloverAliasOpendistro,
				Check: resource.ComposeTestCheckFunc(
					checkElasticsearchIndexRolloverAliasExists(testAccOpendistroProvider, "terraform-test"),
				),
			},
			{
				ResourceName:      "elasticsearch_index.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"aliases",       // not handled by this provider
					"force_destroy", // not returned from the API
				},
				ImportStateCheck: checkElasticsearchIndexRolloverAliasState("terraform-test"),
			},
		},
	})
}

func checkElasticsearchIndexExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("index ID not set")
		}

		meta := testAccProvider.Meta()

		var err error
		esClient, err := getClient(meta.(*ProviderConf))
		if err != nil {
			return err
		}
		switch client := esClient.(type) {
		case *elastic7.Client:
			_, err = client.IndexGetSettings(rs.Primary.ID).Do(context.TODO())
		case *elastic6.Client:
			_, err = client.IndexGetSettings(rs.Primary.ID).Do(context.TODO())
		default:
			elastic5Client := client.(*elastic5.Client)
			_, err = elastic5Client.IndexGetSettings(rs.Primary.ID).Do(context.TODO())
		}

		return err
	}
}

func checkElasticsearchIndexUpdated(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("index ID not set")
		}

		meta := testAccProvider.Meta()
		var settings map[string]interface{}

		var err error
		esClient, err := getClient(meta.(*ProviderConf))
		if err != nil {
			return err
		}
		switch client := esClient.(type) {
		case *elastic7.Client:
			resp, err := client.IndexGetSettings(rs.Primary.ID).Do(context.TODO())
			if err != nil {
				return err
			}
			settings = resp[rs.Primary.ID].Settings["index"].(map[string]interface{})

		case *elastic6.Client:
			resp, err := client.IndexGetSettings(rs.Primary.ID).Do(context.TODO())
			if err != nil {
				return err
			}
			settings = resp[rs.Primary.ID].Settings["index"].(map[string]interface{})

		default:
			elastic5Client := client.(*elastic5.Client)
			resp, err := elastic5Client.IndexGetSettings(rs.Primary.ID).Do(context.TODO())
			if err != nil {
				return err
			}
			settings = resp[rs.Primary.ID].Settings["index"].(map[string]interface{})

		}

		r, ok := settings["number_of_replicas"]
		if ok {
			if ir := r.(string); ir != "2" {
				return fmt.Errorf("expected 2 got %s", ir)
			}
			return nil
		}

		return errors.New("field not found")
	}
}

func checkElasticsearchIndexDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticsearch_index" {
			continue
		}

		meta := testAccProvider.Meta()

		var err error
		esClient, err := getClient(meta.(*ProviderConf))
		if err != nil {
			return err
		}
		switch client := esClient.(type) {
		case *elastic7.Client:
			_, err = client.IndexGetSettings(rs.Primary.ID).Do(context.TODO())
		case *elastic6.Client:
			_, err = client.IndexGetSettings(rs.Primary.ID).Do(context.TODO())
		default:
			elastic5Client := client.(*elastic5.Client)
			_, err = elastic5Client.IndexGetSettings(rs.Primary.ID).Do(context.TODO())
		}

		if err != nil {
			return nil // should be not found error
		}

		return fmt.Errorf("index %q still exists", rs.Primary.ID)
	}

	return nil
}

func checkElasticsearchIndexRolloverAliasExists(provider *schema.Provider, alias string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		meta := provider.Meta()

		var count int
		esClient, err := getClient(meta.(*ProviderConf))
		if err != nil {
			return err
		}
		switch client := esClient.(type) {
		case *elastic7.Client:
			r, err := client.CatAliases().Alias(alias).Do(context.TODO())
			if err != nil {
				return err
			}
			count = len(r)
		case *elastic6.Client:
			r, err := client.CatAliases().Alias(alias).Do(context.TODO())
			if err != nil {
				return err
			}
			count = len(r)
		default:
			elastic5Client := client.(*elastic5.Client)
			r, err := elastic5Client.CatAliases().Alias(alias).Do(context.TODO())
			if err != nil {
				return err
			}
			count = len(r)
		}

		if count == 0 {
			return fmt.Errorf("rollover alias %q not found", alias)
		}

		return nil
	}
}

func checkElasticsearchIndexRolloverAliasState(alias string) resource.ImportStateCheckFunc {
	return func(s []*terraform.InstanceState) error {
		if len(s) != 1 {
			return fmt.Errorf("expected 1 state: %+v", s)
		}
		rs := s[0]
		if rs.Attributes["rollover_alias"] != alias {
			return fmt.Errorf("expected rollover alias %q got %q", alias, rs.Attributes["rollover_alias"])
		}

		return nil
	}
}

func checkElasticsearchIndexRolloverAliasDestroy(provider *schema.Provider, alias string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		meta := provider.Meta()

		var count int
		esClient, err := getClient(meta.(*ProviderConf))
		if err != nil {
			return err
		}
		switch client := esClient.(type) {
		case *elastic7.Client:
			r, err := client.CatAliases().Alias(alias).Do(context.TODO())
			if err != nil {
				return err
			}
			count = len(r)
		case *elastic6.Client:
			r, err := client.CatAliases().Alias(alias).Do(context.TODO())
			if err != nil {
				return err
			}
			count = len(r)
		default:
			elastic5Client := client.(*elastic5.Client)
			r, err := elastic5Client.CatAliases().Alias(alias).Do(context.TODO())
			if err != nil {
				return err
			}
			count = len(r)
		}

		if count > 0 {
			return fmt.Errorf("rollover alias %q still exists", alias)
		}

		return nil
	}
}
