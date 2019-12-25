package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"log"
	"net/http"
	"net/url"
	"regexp"

	awscredentials "github.com/aws/aws-sdk-go/aws/credentials"
	awsec2rolecreds "github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	awsec2metadata "github.com/aws/aws-sdk-go/aws/ec2metadata"
	awssession "github.com/aws/aws-sdk-go/aws/session"
	awssigv4 "github.com/aws/aws-sdk-go/aws/signer/v4"
	"github.com/deoxxa/aws_signing_client"
	"github.com/hashicorp/terraform-plugin-sdk/helper/pathorcontents"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	elastic7 "github.com/olivere/elastic/v7"
	elastic5 "gopkg.in/olivere/elastic.v5"
	elastic6 "gopkg.in/olivere/elastic.v6"
)

var awsUrlRegexp = regexp.MustCompile(`([a-z0-9-]+).es.amazonaws.com$`)

func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"url": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("ELASTICSEARCH_URL", nil),
				Description: "Elasticsearch URL",
			},
			"sniff": {
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ELASTICSEARCH_SNIFF", true),
				Description: "Set the node sniffing option for the elastic client. Client won't work with sniffing if nodes are not routable.",
			},
			"healthcheck": {
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ELASTICSEARCH_HEALTH", true),
				Description: "Set the client healthcheck option for the elastic client. Healthchecking is designed for direct access to the cluster.",
			},
			"username": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ELASTICSEARCH_USERNAME", nil),
				Description: "Username to use to connect to elasticsearch using basic auth",
			},
			"password": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ELASTICSEARCH_PASSWORD", nil),
				Description: "Password to use to connect to elasticsearch using basic auth",
			},
			"aws_access_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "The access key for use with AWS Elasticsearch Service domains",
			},

			"aws_secret_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "The secret key for use with AWS Elasticsearch Service domains",
			},

			"aws_token": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "The session token for use with AWS Elasticsearch Service domains",
			},

			"cacert_file": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "A Custom CA certificate",
			},

			"insecure": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Disable SSL verification of API calls",
			},
			"client_cert_path": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "A X509 certificate to connect to elasticsearch",
				DefaultFunc: schema.EnvDefaultFunc("ES_CLIENT_CERTIFICATE_PATH", ""),
			},
			"client_key_path": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "A X509 key to connect to elasticsearch",
				DefaultFunc: schema.EnvDefaultFunc("ES_CLIENT_KEY_PATH", ""),
			},
			"sign_aws_requests": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Enable signing of AWS elasticsearch requests",
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"elasticsearch_index_template":         resourceElasticsearchIndexTemplate(),
			"elasticsearch_index_lifecycle_policy": resourceElasticsearchIndexLifecyclePolicy(),
			"elasticsearch_snapshot_repository":    resourceElasticsearchSnapshotRepository(),
			"elasticsearch_kibana_object":          resourceElasticsearchKibanaObject(),
			"elasticsearch_watch":                  resourceElasticsearchWatch(),
			"elasticsearch_monitor":                resourceElasticsearchMonitor(),
			"elasticsearch_destination":            resourceElasticsearchDestination(),
			"elasticsearch_xpack_role_mapping":     resourceElasticsearchXpackRoleMapping(),
			"elasticsearch_xpack_role":             resourceElasticsearchXpackRole(),
			"elasticsearch_ingest_pipeline":        resourceElasticsearchIngestPipeline(),
			"elasticsearch_xpack_user":             resourceElasticsearchXpackUser(),
		},

		DataSourcesMap: map[string]*schema.Resource{
			"elasticsearch_destination": dataSourceElasticsearchDestination(),
			"elasticsearch_host":        dataSourceElasticsearchHost(),
		},

		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	rawUrl := d.Get("url").(string)
	insecure := d.Get("insecure").(bool)
	sniffing := d.Get("sniff").(bool)
	healthchecking := d.Get("healthcheck").(bool)
	cacertFile := d.Get("cacert_file").(string)
	username := d.Get("username").(string)
	password := d.Get("password").(string)
	parsedUrl, err := url.Parse(rawUrl)
	signAWSRequests := d.Get("sign_aws_requests").(bool)
	if err != nil {
		return nil, err
	}

	opts := []elastic7.ClientOptionFunc{
		elastic7.SetURL(rawUrl),
		elastic7.SetScheme(parsedUrl.Scheme),
		elastic7.SetSniff(sniffing),
		elastic7.SetHealthcheck(healthchecking),
	}

	if parsedUrl.User.Username() != "" {
		p, _ := parsedUrl.User.Password()
		opts = append(opts, elastic7.SetBasicAuth(parsedUrl.User.Username(), p))
	}
	if username != "" && password != "" {
		opts = append(opts, elastic7.SetBasicAuth(username, password))
	}

	if m := awsUrlRegexp.FindStringSubmatch(parsedUrl.Hostname()); m != nil && signAWSRequests {
		log.Printf("[INFO] Using AWS: %+v", m[1])
		opts = append(opts, elastic7.SetHttpClient(awsHttpClient(m[1], d)), elastic7.SetSniff(false))
	} else if insecure || cacertFile != "" {
		opts = append(opts, elastic7.SetHttpClient(tlsHttpClient(d)), elastic7.SetSniff(false))
	}

	var relevantClient interface{}
	client, err := elastic7.NewClient(opts...)
	if err != nil {
		return nil, err
	}
	relevantClient = client

	// Use the v7 client to ping the cluster to determine the version
	info, _, err := client.Ping(rawUrl).Do(context.TODO())
	if err != nil {
		return nil, err
	}

	if info.Version.Number < "7.0.0" && info.Version.Number >= "6.0.0" {
		log.Printf("[INFO] Using ES 6")
		opts := []elastic6.ClientOptionFunc{
			elastic6.SetURL(rawUrl),
			elastic6.SetScheme(parsedUrl.Scheme),
			elastic6.SetSniff(sniffing),
			elastic6.SetHealthcheck(healthchecking),
		}

		if parsedUrl.User.Username() != "" {
			p, _ := parsedUrl.User.Password()
			opts = append(opts, elastic6.SetBasicAuth(parsedUrl.User.Username(), p))
		}
		if username != "" && password != "" {
			opts = append(opts, elastic6.SetBasicAuth(username, password))
		}

		if m := awsUrlRegexp.FindStringSubmatch(parsedUrl.Hostname()); m != nil && signAWSRequests {
			log.Printf("[INFO] Using AWS: %+v", m[1])
			opts = append(opts, elastic6.SetHttpClient(awsHttpClient(m[1], d)), elastic6.SetSniff(false))
		} else if insecure || cacertFile != "" {
			opts = append(opts, elastic6.SetHttpClient(tlsHttpClient(d)), elastic6.SetSniff(false))
		}
		relevantClient, err = elastic6.NewClient(opts...)
		if err != nil {
			return nil, err
		}
	} else if info.Version.Number < "6.0.0" && info.Version.Number >= "5.0.0" {
		log.Printf("[INFO] Using ES 5")
		opts := []elastic5.ClientOptionFunc{
			elastic5.SetURL(rawUrl),
			elastic5.SetScheme(parsedUrl.Scheme),
			elastic5.SetSniff(sniffing),
			elastic5.SetHealthcheck(healthchecking),
		}

		if parsedUrl.User.Username() != "" {
			p, _ := parsedUrl.User.Password()
			opts = append(opts, elastic5.SetBasicAuth(parsedUrl.User.Username(), p))
		}
		if username != "" && password != "" {
			opts = append(opts, elastic5.SetBasicAuth(username, password))
		}

		if m := awsUrlRegexp.FindStringSubmatch(parsedUrl.Hostname()); m != nil && signAWSRequests {
			opts = append(opts, elastic5.SetHttpClient(awsHttpClient(m[1], d)), elastic5.SetSniff(false))
		} else if insecure || cacertFile != "" {
			opts = append(opts, elastic5.SetHttpClient(tlsHttpClient(d)), elastic5.SetSniff(false))
		}
		relevantClient, err = elastic5.NewClient(opts...)
		if err != nil {
			return nil, err
		}
	} else if info.Version.Number < "5.0.0" {
		return nil, errors.New("ElasticSearch is older than 5.0.0!")
	}

	return relevantClient, nil
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
		&awscredentials.EnvProvider{},
		&awscredentials.SharedCredentialsProvider{},
		&awsec2rolecreds.EC2RoleProvider{
			Client: awsec2metadata.New(awssession.Must(awssession.NewSession())),
		},
	})
	signer := awssigv4.NewSigner(creds)
	client, _ := aws_signing_client.New(signer, nil, "es", region)

	return client
}

func tlsHttpClient(d *schema.ResourceData) *http.Client {
	insecure := d.Get("insecure").(bool)
	cacertFile := d.Get("cacert_file").(string)
	certPemPath := d.Get("client_cert_path").(string)
	keyPemPath := d.Get("client_key_path").(string)

	// Configure TLS/SSL
	tlsConfig := &tls.Config{}
	if certPemPath != "" && keyPemPath != "" {
		certPem, _, err := pathorcontents.Read(certPemPath)
		if err != nil {
			log.Fatal(err)
		}
		keyPem, _, err := pathorcontents.Read(keyPemPath)
		if err != nil {
			log.Fatal(err)
		}
		cert, err := tls.X509KeyPair([]byte(certPem), []byte(keyPem))
		if err != nil {
			log.Fatal(err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

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
