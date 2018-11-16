package main

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

var testAccProviders map[string]terraform.ResourceProvider
var testAccProvider *schema.Provider

func init() {
	testAccProvider = Provider().(*schema.Provider)
	testAccProviders = map[string]terraform.ResourceProvider{
		"elasticsearch": testAccProvider,
	}
}

func TestProvider(t *testing.T) {
	if err := Provider().(*schema.Provider).InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_impl(t *testing.T) {
	var _ terraform.ResourceProvider = Provider()
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("ELASTICSEARCH_URL"); v == "" {
		t.Fatal("ELASTICSEARCH_URL must be set for acceptance tests")
	}
}

func testAccPreCheckXpack(t *testing.T) {
	if v := os.Getenv("ELASTICSEARCH_URL"); v == "" {
		t.Fatal("ELASTICSEARCH_URL must be set for acceptance tests")
	}
	if u := os.Getenv("ELASTICSEARCH_USERNAME"); u == "" {
		t.Fatal("ELASTICSEARCH_USERNAME must be set for acceptance tests on xpack cluster")
	}
	if p := os.Getenv("ELASTICSEARCH_PASSWORD"); p == "" {
		t.Fatal("ELASTICSEARCH_PASSWORD must be set for acceptance tests on xpack cluster")
	}
}
