package main

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

var testAccProviders map[string]terraform.ResourceProvider
var testAccProvider *schema.Provider

var testAccXPackProviders map[string]terraform.ResourceProvider
var testAccXPackProvider *schema.Provider

var testAccOpendistroProviders map[string]terraform.ResourceProvider
var testAccOpendistroProvider *schema.Provider

func init() {
	testAccProvider = Provider().(*schema.Provider)
	testAccProviders = map[string]terraform.ResourceProvider{
		"elasticsearch": testAccProvider,
	}

	testAccXPackProvider = Provider().(*schema.Provider)
	testAccXPackProviders = map[string]terraform.ResourceProvider{
		"elasticsearch": testAccXPackProvider,
	}

	xPackOriginalConfigureFunc := testAccXPackProvider.ConfigureFunc
	testAccXPackProvider.ConfigureFunc = func(d *schema.ResourceData) (interface{}, error) {
		d.Set("url", "http://elastic:elastic@127.0.0.1:9210")
		return xPackOriginalConfigureFunc(d)
	}

	testAccOpendistroProvider = Provider().(*schema.Provider)
	testAccOpendistroProviders = map[string]terraform.ResourceProvider{
		"elasticsearch": testAccOpendistroProvider,
	}

	opendistroOriginalConfigureFunc := testAccOpendistroProvider.ConfigureFunc
	testAccOpendistroProvider.ConfigureFunc = func(d *schema.ResourceData) (interface{}, error) {
		d.Set("url", "http://elastic:elastic@127.0.0.1:9220")
		return opendistroOriginalConfigureFunc(d)
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
	testAccPreCheck(t)
	if x := os.Getenv("XPACK"); x != "true" {
		t.Skip("Xpack acceptance tests must be run on an Xpack activated cluster")
	}
	if u := os.Getenv("ELASTICSEARCH_USERNAME"); u == "" {
		t.Fatal("ELASTICSEARCH_USERNAME must be set for acceptance tests on xpack cluster")
	}
	if p := os.Getenv("ELASTICSEARCH_PASSWORD"); p == "" {
		t.Fatal("ELASTICSEARCH_PASSWORD must be set for acceptance tests on xpack cluster")
	}
}
