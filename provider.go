package main

import (
	"crypto/tls"
	"crypto/x509"
	"log"
	"net/http"
	"net/url"
	"regexp"

	awscredentials "github.com/aws/aws-sdk-go/aws/credentials"
	awssigv4 "github.com/aws/aws-sdk-go/aws/signer/v4"
	"github.com/deoxxa/aws_signing_client"
	"github.com/hashicorp/terraform/helper/pathorcontents"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	elastic "gopkg.in/olivere/elastic.v5"
)

var awsUrlRegexp = regexp.MustCompile(`([a-z0-9-]+).es.amazonaws.com$`)

func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"url": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("ELASTICSEARCH_URL", nil),
				Description: "Elasticsearch URL",
			},

			"aws_access_key": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "The access key for use with AWS Elasticsearch Service domains",
			},

			"aws_secret_key": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "The secret key for use with AWS Elasticsearch Service domains",
			},

			"aws_token": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "The session token for use with AWS Elasticsearch Service domains",
			},

			"cacert_file": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "A Custom CA certificate",
			},

			"insecure": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Disable SSL verification of API calls",
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"elasticsearch_index_template":      resourceElasticsearchIndexTemplate(),
			"elasticsearch_snapshot_repository": resourceElasticsearchSnapshotRepository(),
			"elasticsearch_kibana_object":       resourceElasticsearchKibanaObject(),
		},

		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	rawUrl := d.Get("url").(string)
	insecure := d.Get("insecure").(bool)
	cacertFile := d.Get("cacert_file").(string)
	parsedUrl, err := url.Parse(rawUrl)
	if err != nil {
		return nil, err
	}
	opts := []elastic.ClientOptionFunc{
		elastic.SetURL(rawUrl),
		elastic.SetScheme(parsedUrl.Scheme),
	}

	if m := awsUrlRegexp.FindStringSubmatch(parsedUrl.Hostname()); m != nil {
		log.Printf("[INFO] Using AWS: %+v", m[1])
		opts = append(opts, elastic.SetHttpClient(awsHttpClient(m[1], d)), elastic.SetSniff(false))
	} else if insecure || cacertFile != "" {
		opts = append(opts, elastic.SetHttpClient(tlsHttpClient(d)), elastic.SetSniff(false))
	}

	return elastic.NewClient(opts...)
}

func awsHttpClient(region string, d *schema.ResourceData) *http.Client {
	creds := awscredentials.NewChainCredentials([]awscredentials.Provider{
		&awscredentials.StaticProvider{
			Value: awscredentials.Value{
				AccessKeyID:     d.Get("aws_access_key").(string),
				SecretAccessKey: d.Get("aws_secret_key").(string),
				SessionToken:    d.Get("aws_token").(string),
			},
		},
		&awscredentials.SharedCredentialsProvider{},
		&awscredentials.EnvProvider{},
	})
	signer := awssigv4.NewSigner(creds)
	client, _ := aws_signing_client.New(signer, nil, "es", region)

	return client
}

func tlsHttpClient(d *schema.ResourceData) *http.Client {
	insecure := d.Get("insecure").(bool)
	cacertFile := d.Get("cacert_file").(string)

	// Configure TLS/SSL
	tlsConfig := &tls.Config{}

	// If a cacertFile has been specified, use that for cert validation
	if cacertFile != "" {
		caCert, _, _ := pathorcontents.Read(cacertFile)

		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM([]byte(caCert))
		tlsConfig.RootCAs = caCertPool
	}

	// If configured as insecure, turn off SSL verification
	if insecure {
		tlsConfig.InsecureSkipVerify = true
	}

	transport := &http.Transport{TLSClientConfig: tlsConfig}

	client := &http.Client{Transport: transport}

	return client
}
