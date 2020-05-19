package es

import (
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
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
		err := d.Set("url", "http://elastic:elastic@127.0.0.1:9210")
		if err != nil {
			return nil, err
		}
		return xPackOriginalConfigureFunc(d)
	}

	testAccOpendistroProvider = Provider().(*schema.Provider)
	testAccOpendistroProviders = map[string]terraform.ResourceProvider{
		"elasticsearch": testAccOpendistroProvider,
	}

	opendistroOriginalConfigureFunc := testAccOpendistroProvider.ConfigureFunc
	testAccOpendistroProvider.ConfigureFunc = func(d *schema.ResourceData) (interface{}, error) {
		err := d.Set("url", "http://elastic:elastic@127.0.0.1:9220")
		if err != nil {
			return nil, err
		}
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

func TestAWSCreds(t *testing.T) {
	testRegion := "us-east-1"
	manualAccessKeyID := "MANUAL_ACCESS_KEY"
	envAccessKeyID := "ENV_ACCESS_KEY"
	profileAccessKeyID := "PROFILE_ACCESS_KEY"

	os.Setenv("AWS_CONFIG_FILE", "../test_aws_config")
	os.Setenv("AWS_ACCESS_KEY_ID", envAccessKeyID)
	os.Setenv("AWS_SECRET_ACCESS_KEY", "ENV_SECRET")

	// first, check that if we set aws_profile with aws_access_key_id - the latter takes precedence
	testConfig := map[string]interface{}{
		"aws_profile":    "testing",
		"aws_access_key": "MANUAL_ACCESS_KEY",
		"aws_secret_key": "MANUAL_SECRET_KEY",
	}

	creds := getCreds(t, testRegion, testConfig)

	if creds.AccessKeyID != manualAccessKeyID {
		t.Errorf("access key id should have been %s (we got %s)", manualAccessKeyID, creds.AccessKeyID)
	}

	// Now get rid of the manual access keys, and ensure that the profile is used
	testConfig = map[string]interface{}{
		"aws_profile": "testing",
	}

	creds = getCreds(t, testRegion, testConfig)

	if creds.AccessKeyID != profileAccessKeyID {
		t.Errorf("access key id should have been %s (we got %s)", profileAccessKeyID, creds.AccessKeyID)
	}

	// Now try without anything - it should use the default creds provider and pickup the env variables
	testConfig = map[string]interface{}{}

	creds = getCreds(t, testRegion, testConfig)

	if creds.AccessKeyID != envAccessKeyID {
		t.Errorf("access key id should have been %s (we got %s)", envAccessKeyID, creds.AccessKeyID)
	}
}

func getCreds(t *testing.T, region string, config map[string]interface{}) credentials.Value {
	testConfigData := schema.TestResourceDataRaw(t, Provider().(*schema.Provider).Schema, config)
	s := awsSession(region, testConfigData)
	if s == nil {
		t.Fatalf("awsSession returned nil")
	}
	creds, err := s.Config.Credentials.Get()
	if err != nil {
		t.Fatalf("Failed fetching credentials: %v", err)
	}
	return creds
}
