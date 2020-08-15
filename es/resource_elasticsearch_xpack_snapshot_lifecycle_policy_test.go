package es

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

	elastic7 "github.com/olivere/elastic/v7"
	elastic5 "gopkg.in/olivere/elastic.v5"
	elastic6 "gopkg.in/olivere/elastic.v6"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccElasticsearchXpackSnapshotLifecyclePolicy(t *testing.T) {
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
	case *elastic5.Client, *elastic6.Client:
		allowed = false
	default:
		allowed = true
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			if !allowed {
				t.Skip("Snapshot lifecycles only supported on ES >= 7")
			}
		},
		Providers:    testAccXPackProviders,
		CheckDestroy: testCheckElasticsearchXpackSnapshotLifecyclePolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccElasticsearchSnapshotRepository,
			},
			{
				Config: testAccElasticsearchXpackSnapshotLifecyclePolicy,
				Check: resource.ComposeTestCheckFunc(
					testCheckElasticsearchXpackSnapshotLifecyclePolicyExists("elasticsearch_xpack_snapshot_lifecycle_policy.terraform-test"),
				),
			},
		},
	})
}

func TestAccElasticsearchXpackSnapshotLifecyclePolicy_importBasic(t *testing.T) {
	provider := Provider()
	diags := provider.Configure(context.Background(), &terraform.ResourceConfig{})
	if diags.HasError() {
		t.Skipf("err: %#v", diags)
	}
	meta := provider.Meta()
	var allowed bool
	esClient, err := getClient(meta.(*ProviderConf))
	if err != nil {
		t.Skipf("err: %s", err)
	}
	switch esClient.(type) {
	case *elastic5.Client, *elastic6.Client:
		allowed = false
	default:
		allowed = true
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			if !allowed {
				t.Skip("Snapshot lifecycles only supported on ES >= 7")
			}
		},
		Providers:    testAccXPackProviders,
		CheckDestroy: testCheckElasticsearchXpackSnapshotLifecyclePolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccElasticsearchSnapshotRepository,
			},
			{
				Config: testAccElasticsearchXpackSnapshotLifecyclePolicy,
			},
			{
				ResourceName:      "elasticsearch_xpack_snapshot_lifecycle_policy.terraform-test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testCheckElasticsearchXpackSnapshotLifecyclePolicyExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No snapshot lifecycle policy ID is set")
		}

		meta := testAccXPackProvider.Meta()

		var err error
		esClient, err := getClient(meta.(*ProviderConf))
		if err != nil {
			return err
		}
		switch client := esClient.(type) {
		case *elastic7.Client:
			_, err = client.PerformRequest(context.TODO(), elastic7.PerformRequestOptions{Method: http.MethodGet, Path: "/_slm/policy/" + rs.Primary.ID})
		default:
			err = errors.New("Snapshot Lifecycle Management is only supported by the elastic library >= v7!")
		}

		if err != nil {
			return err
		}

		return nil
	}
}

func testCheckElasticsearchXpackSnapshotLifecyclePolicyDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticsearch_xpack_snapshot_lifecycle_policy" {
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
			_, err = client.PerformRequest(context.TODO(), elastic7.PerformRequestOptions{Method: http.MethodDelete, Path: "/_slm/policy/" + rs.Primary.ID})
		default:
			err = errors.New("Snapshot Lifecycle Management is only supported by the elastic library >= v7!")
		}

		if err != nil {
			return nil // should be not found error
		}

		return fmt.Errorf("Snapshot lifecycle policy %q still exists", rs.Primary.ID)
	}

	return nil
}

var testAccElasticsearchXpackSnapshotLifecyclePolicy = `
resource "elasticsearch_xpack_snapshot_lifecycle_policy" "terraform-test" {
  name = "terraformtest"
  body = <<EOF
{
  "schedule": "0 30 1 * * ?", 
  "name": "<daily-snap-{now/d}>", 
  "repository": "terraform-test",
  "config": { 
    "indices": ["data-*", "important"], 
    "ignore_unavailable": false,
    "include_global_state": false
  },
  "retention": { 
    "expire_after": "30d", 
    "min_count": 5, 
    "max_count": 50 
  }
}
EOF
}
`
