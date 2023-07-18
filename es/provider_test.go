package es

import (
	"context"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var testAccProviders map[string]*schema.Provider
var testAccProviderFactories func(providers *[]*schema.Provider) map[string]func() (*schema.Provider, error)
var testAccProvider *schema.Provider

var testAccXPackProviders map[string]*schema.Provider
var testAccXPackProvider *schema.Provider

var testAccOpendistroProviders map[string]*schema.Provider
var testAccOpendistroProvider *schema.Provider

var testAccKibanaProviders map[string]*schema.Provider
var testAccKibanaProvider *schema.Provider

func init() {
	testAccProvider = Provider()
	testAccProviders = map[string]*schema.Provider{
		"elasticsearch": testAccProvider,
	}
	testAccProviderFactories = func(providers *[]*schema.Provider) map[string]func() (*schema.Provider, error) {
		// this is an SDKV2 compatible hack, the "factory" functions are
		// effectively singletons for the lifecycle of a resource.Test
		var factories = make(map[string]func() (*schema.Provider, error), len(testAccProviders))
		for name, p := range testAccProviders {
			factories[name] = func() (*schema.Provider, error) {
				return p, nil
			}
			*providers = append(*providers, p)
		}
		return factories
	}

	testAccXPackProvider = Provider()
	testAccXPackProviders = map[string]*schema.Provider{
		"elasticsearch": testAccXPackProvider,
	}

	xPackOriginalConfigureFunc := testAccXPackProvider.ConfigureContextFunc
	testAccXPackProvider.ConfigureContextFunc = func(c context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		err := d.Set("url", "http://elastic:elastic@127.0.0.1:9210")
		if err != nil {
			return nil, diag.FromErr(err)
		}
		return xPackOriginalConfigureFunc(c, d)
	}

	testAccOpendistroProvider = Provider()
	testAccOpendistroProviders = map[string]*schema.Provider{
		"elasticsearch": testAccOpendistroProvider,
	}

	opendistroOriginalConfigureFunc := testAccOpendistroProvider.ConfigureContextFunc
	testAccOpendistroProvider.ConfigureContextFunc = func(c context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		err := d.Set("url", "http://admin:admin@127.0.0.1:9220")
		if err != nil {
			return nil, diag.FromErr(err)
		}
		return opendistroOriginalConfigureFunc(c, d)
	}

	testAccKibanaProvider = Provider()
	testAccKibanaProviders = map[string]*schema.Provider{
		"elasticsearch": testAccKibanaProvider,
	}

	kibanaOriginalConfigureFunc := testAccKibanaProvider.ConfigureContextFunc
	testAccKibanaProvider.ConfigureContextFunc = func(c context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		err := d.Set("kibana_url", "http://127.0.0.1:5601")
		if err != nil {
			return nil, diag.FromErr(err)
		}
		return kibanaOriginalConfigureFunc(c, d)
	}
}

func TestProvider(t *testing.T) {
	if err := Provider().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_impl(t *testing.T) {
	var _ = Provider()
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("ELASTICSEARCH_URL"); v == "" {
		t.Fatal("ELASTICSEARCH_URL must be set for acceptance tests")
	}
}

// Given:
// 1. AWS credentials are specified via environment variables
// 2. aws access key and secret access key are specified via the provider configuration
// 3. a named profile is specified via the provider config
//
// this tests that:  the configured provider access key / secret key are used over the other options (ie: #2)
func TestAWSCredsManualKey(t *testing.T) {
	envAccessKeyID := "ENV_ACCESS_KEY"
	testRegion := "us-east-1"
	manualAccessKeyID := "MANUAL_ACCESS_KEY"
	namedProfile := "testing"

	os.Setenv("AWS_ACCESS_KEY_ID", envAccessKeyID)
	os.Setenv("AWS_SECRET_ACCESS_KEY", "ENV_SECRET")

	// first, check that if we set aws_profile with aws_access_key_id - the latter takes precedence
	testConfig := map[string]interface{}{
		"aws_profile":    namedProfile,
		"aws_access_key": manualAccessKeyID,
		"aws_secret_key": "MANUAL_SECRET_KEY",
	}

	creds := getCreds(t, testRegion, testConfig)

	if creds.AccessKeyID != manualAccessKeyID {
		t.Errorf("access key id should have been %s (we got %s)", manualAccessKeyID, creds.AccessKeyID)
	}
}

// Given:
// 1. AWS credentials are specified via environment variables
// 2. a named profile is specified via the provider config
//
// this tests that:  the named profile credentials are used over the env vars
func TestAWSCredsNamedProfile(t *testing.T) {
	envAccessKeyID := "ENV_ACCESS_KEY"
	testRegion := "us-east-1"
	namedProfile := "testing"
	profileAccessKeyID := "PROFILE_ACCESS_KEY"

	os.Setenv("AWS_CONFIG_FILE", "../test_aws_config") // set config file so we can ensure the profile we want to test exists
	os.Setenv("AWS_SDK_LOAD_CONFIG", "1")
	os.Setenv("AWS_ACCESS_KEY_ID", envAccessKeyID)
	os.Setenv("AWS_SECRET_ACCESS_KEY", "ENV_SECRET")

	testConfig := map[string]interface{}{
		"aws_profile": namedProfile,
	}

	creds := getCreds(t, testRegion, testConfig)

	if creds.AccessKeyID != profileAccessKeyID {
		t.Errorf("access key id should have been %s (we got %s)", profileAccessKeyID, creds.AccessKeyID)
	}

	os.Unsetenv("AWS_ACCESS_KEY_ID")
	os.Unsetenv("AWS_SECRET_ACCESS_KEY")
	os.Unsetenv("AWS_CONFIG_FILE")
	os.Unsetenv("AWS_SDK_LOAD_CONFIG")
}

// Given:
// 1. AWS credentials are specified via environment variables
// 2. No configuration provided to the provider
//
// This tests that: we get the credentials from the environment variables (ie: from the default credentials provider chain)

func TestAWSCredsEnv(t *testing.T) {
	envAccessKeyID := "ENV_ACCESS_KEY"
	testRegion := "us-east-1"

	os.Setenv("AWS_ACCESS_KEY_ID", envAccessKeyID)
	os.Setenv("AWS_SECRET_ACCESS_KEY", "ENV_SECRET")

	testConfig := map[string]interface{}{}

	creds := getCreds(t, testRegion, testConfig)

	if creds.AccessKeyID != envAccessKeyID {
		t.Errorf("access key id should have been %s (we got %s)", envAccessKeyID, creds.AccessKeyID)
	}

	os.Unsetenv("AWS_ACCESS_KEY_ID")
	os.Unsetenv("AWS_SECRET_ACCESS_KEY")
}

func TestAWSCredsEnvNamedProfile(t *testing.T) {
	namedProfile := "testing"
	testRegion := "us-east-1"
	profileAccessKeyID := "PROFILE_ACCESS_KEY"

	os.Setenv("AWS_PROFILE", namedProfile)
	os.Setenv("AWS_SDK_LOAD_CONFIG", "1")
	os.Setenv("AWS_CONFIG_FILE", "../test_aws_config") // set config file so we can ensure the profile we want to test exists

	testConfig := map[string]interface{}{}

	creds := getCreds(t, testRegion, testConfig)

	if creds.AccessKeyID != profileAccessKeyID {
		t.Errorf("access key id should have been %s (we got %s)", profileAccessKeyID, creds.AccessKeyID)
	}
	os.Unsetenv("AWS_PROFILE")
	os.Unsetenv("AWS_CONFIG_FILE")
	os.Unsetenv("AWS_SDK_LOAD_CONFIG")
}

// Given:
// 1. An AWS role ARN is specified
// 2. No additional AWS configuration is provided to the provider
//
// This tests that: we can safely generate a session. Note we cannot get the credentials, because that requires connecting to AWS
func TestAWSCredsAssumeRole(t *testing.T) {
	testRegion := "us-east-1"

	testConfig := map[string]interface{}{
		"aws_assume_role_arn":         "test_arn",
		"aws_assume_role_external_id": "secret_id",
	}

	testConfigData := schema.TestResourceDataRaw(t, Provider().Schema, testConfig)

	conf := &ProviderConf{
		awsAssumeRoleArn:        testConfigData.Get("aws_assume_role_arn").(string),
		awsAssumeRoleExternalID: testConfigData.Get("aws_assume_role_external_id").(string),
	}
	s := awsSession(testRegion, conf)
	if s == nil {
		t.Fatalf("awsSession returned nil")
	}
}

func getCreds(t *testing.T, region string, config map[string]interface{}) credentials.Value {
	awsAccessKey := ""
	awsSecretKey := ""
	awsProfile := ""
	if val, ok := config["aws_access_key"]; ok {
		awsAccessKey = val.(string)
	}
	if val, ok := config["aws_secret_key"]; ok {
		awsSecretKey = val.(string)
	}
	if val, ok := config["aws_profile"]; ok {
		awsProfile = val.(string)
	}

	conf := &ProviderConf{
		awsAccessKeyId:     awsAccessKey,
		awsSecretAccessKey: awsSecretKey,
		awsProfile:         awsProfile,
	}
	s := awsSession(region, conf)
	if s == nil {
		t.Fatalf("awsSession returned nil")
	}
	creds, err := s.Config.Credentials.Get()
	if err != nil {
		t.Fatalf("Failed fetching credentials: %v", err)
	}
	return creds
}
