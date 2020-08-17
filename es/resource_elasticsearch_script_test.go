package es

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"

	elastic7 "github.com/olivere/elastic/v7"
	elastic5 "gopkg.in/olivere/elastic.v5"
	elastic6 "gopkg.in/olivere/elastic.v6"
)

const (
	testAccElasticsearchScript = `
resource "elasticsearch_script" "test" {
  name = "terraform-test"
  lang = "painless"
  source = "ctx._source.message = params.message"
}
`
	testAccElasticsearchScriptUpdate1 = `
resource "elasticsearch_script" "test" {
  name = "terraform-test"
  lang = "expression"
  source = "ctx._source.message = params.new_message"
}
`
)

func TestAccElasticsearchScript(t *testing.T) {
	provider := Provider().(*schema.Provider)
	err := provider.Configure(&terraform.ResourceConfig{})
	if err != nil {
		t.Skipf("err: %s", err)
	}
	meta := provider.Meta()
	client, err := getClient(meta.(*ProviderConf))
	if err != nil {
		t.Skipf("err: %s", err)
	}
	var allowed bool

	switch client.(type) {
	case *elastic5.Client:
		allowed = false
	default:
		allowed = true
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			if !allowed {
				t.Skip("Script resouce only supported on >= ES 6")
			}
		},
		Providers:    testAccProviders,
		CheckDestroy: checkElasticsearchScriptDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccElasticsearchScript,
				Check: resource.ComposeTestCheckFunc(
					checkElasticsearchScriptExists("elasticsearch_script.test"),
				),
			},
			{
				Config: testAccElasticsearchScriptUpdate1,
				Check: resource.ComposeTestCheckFunc(
					checkElasticsearchScriptUpdated("elasticsearch_script.test"),
				),
			},
		},
	})
}

func TestAccElasticsearchScript_importBasic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: checkElasticsearchScriptDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccElasticsearchScript,
			},
			{
				ResourceName:            "elasticsearch_script.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
		},
	})
}

func checkElasticsearchScriptExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("script ID not set")
		}

		meta := testAccProvider.Meta()

		var err error
		esClient, err := getClient(meta.(*ProviderConf))
		if err != nil {
			return err
		}
		switch client := esClient.(type) {
		case *elastic7.Client:
			_, err = client.GetScript().Id(rs.Primary.ID).Do(context.Background())
		case *elastic6.Client:
			_, err = client.GetScript().Id(rs.Primary.ID).Do(context.Background())
		default:
		}

		return err
	}
}

func checkElasticsearchScriptUpdated(name string) resource.TestCheckFunc {
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
		var scriptJson json.RawMessage
		esClient, err := getClient(meta.(*ProviderConf))
		if err != nil {
			return err
		}
		switch client := esClient.(type) {
		case *elastic7.Client:
			var res *elastic7.GetScriptResponse
			res, err = client.GetScript().Id(rs.Primary.ID).Do(context.Background())
			if err != nil {
				return err
			}
			scriptJson = res.Script

		case *elastic6.Client:
			var res *elastic6.GetScriptResponse
			res, err = client.GetScript().Id(rs.Primary.ID).Do(context.Background())
			if err != nil {
				return err
			}
			scriptJson = res.Script

		default:
		}

		var script map[string]interface{}
		if err := json.Unmarshal(scriptJson, &script); err != nil {
			return err
		}

		l, ok := script["lang"]
		if ok {
			if lang := l.(string); lang != "expression" {
				return fmt.Errorf("expected expression got %s", lang)
			}
			return nil
		}

		return errors.New("field not found")
	}
}

func checkElasticsearchScriptDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticsearch_script" {
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
			_, err = client.GetScript().Id(rs.Primary.ID).Do(context.Background())
		case *elastic6.Client:
			_, err = client.GetScript().Id(rs.Primary.ID).Do(context.Background())
		default:
		}

		if err != nil {
			return nil // should be not found error
		}

		return fmt.Errorf("index %q still exists", rs.Primary.ID)
	}

	return nil
}
