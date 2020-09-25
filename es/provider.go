package es

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"log"
	"net/http"
	"net/url"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	awscredentials "github.com/aws/aws-sdk-go/aws/credentials"
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

			"aws_profile": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "The AWS profile for use with AWS Elasticsearch Service domains",
			},

			"aws_region": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "The AWS region for use in signing of AWS elasticsearch requests. Must be specified in order to use AWS URL signing with AWS ElasticSearch endpoint exposed on a custom DNS domain.",
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
				Description: "Enable signing of AWS elasticsearch requests. The `url` must refer to AWS ES domain (`*.<region>.es.amazonaws.com`), or `aws_region` must be specified explicitly.",
			},
			"elasticsearch_version": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "ElasticSearch Version",
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"elasticsearch_destination":                     resourceElasticsearchDeprecatedDestination(),
			"elasticsearch_index":                           resourceElasticsearchIndex(),
			"elasticsearch_index_lifecycle_policy":          resourceElasticsearchIndexLifecyclePolicy(),
			"elasticsearch_index_template":                  resourceElasticsearchIndexTemplate(),
			"elasticsearch_ingest_pipeline":                 resourceElasticsearchIngestPipeline(),
			"elasticsearch_kibana_object":                   resourceElasticsearchKibanaObject(),
			"elasticsearch_monitor":                         resourceElasticsearchDeprecatedMonitor(),
			"elasticsearch_snapshot_repository":             resourceElasticsearchSnapshotRepository(),
			"elasticsearch_watch":                           resourceElasticsearchDeprecatedWatch(),
			"elasticsearch_opendistro_destination":          resourceElasticsearchOpenDistroDestination(),
			"elasticsearch_opendistro_ism_policy":           resourceElasticsearchOpenDistroISMPolicy(),
			"elasticsearch_opendistro_ism_policy_mapping":   resourceElasticsearchOpenDistroISMPolicyMapping(),
			"elasticsearch_opendistro_monitor":              resourceElasticsearchOpenDistroMonitor(),
			"elasticsearch_opendistro_roles_mapping":        resourceElasticsearchOpenDistroRolesMapping(),
			"elasticsearch_opendistro_role":                 resourceElasticsearchOpenDistroRole(),
			"elasticsearch_opendistro_user":                 resourceElasticsearchOpenDistroUser(),
			"elasticsearch_xpack_license":                   resourceElasticsearchXpackLicense(),
			"elasticsearch_xpack_role":                      resourceElasticsearchXpackRole(),
			"elasticsearch_xpack_role_mapping":              resourceElasticsearchXpackRoleMapping(),
			"elasticsearch_xpack_snapshot_lifecycle_policy": resourceElasticsearchXpackSnapshotLifecyclePolicy(),
			"elasticsearch_xpack_user":                      resourceElasticsearchXpackUser(),
			"elasticsearch_xpack_watch":                     resourceElasticsearchXpackWatch(),
		},

		DataSourcesMap: map[string]*schema.Resource{
			"elasticsearch_destination":            dataSourceElasticsearchDeprecatedDestination(),
			"elasticsearch_host":                   dataSourceElasticsearchHost(),
			"elasticsearch_opendistro_destination": dataSourceElasticsearchOpenDistroDestination(),
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
	esVersion := d.Get("elasticsearch_version").(string)
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
	} else if awsRegion := d.Get("aws_region").(string); awsRegion != "" && signAWSRequests {
		log.Printf("[INFO] Using AWS: %+v", awsRegion)
		opts = append(opts, elastic7.SetHttpClient(awsHttpClient(awsRegion, d)), elastic7.SetSniff(false))
	} else if insecure || cacertFile != "" {
		opts = append(opts, elastic7.SetHttpClient(tlsHttpClient(d)), elastic7.SetSniff(false))
	}

	var relevantClient interface{}
	client, err := elastic7.NewClient(opts...)
	if err != nil {
		return nil, err
	}
	relevantClient = client

	// Use the v7 client to ping the cluster to determine the version if one was not provided
	if esVersion == "" {
		info, _, err := client.Ping(rawUrl).Do(context.TODO())
		if err != nil {
			return nil, err
		}
		esVersion = info.Version.Number
	}

	if esVersion < "7.0.0" && esVersion >= "6.0.0" {
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
		} else if awsRegion := d.Get("aws_region").(string); awsRegion != "" && signAWSRequests {
			log.Printf("[INFO] Using AWS: %+v", awsRegion)
			opts = append(opts, elastic6.SetHttpClient(awsHttpClient(awsRegion, d)), elastic6.SetSniff(false))
		} else if insecure || cacertFile != "" {
			opts = append(opts, elastic6.SetHttpClient(tlsHttpClient(d)), elastic6.SetSniff(false))
		}
		relevantClient, err = elastic6.NewClient(opts...)
		if err != nil {
			return nil, err
		}
	} else if esVersion < "6.0.0" && esVersion >= "5.0.0" {
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
		} else if awsRegion := d.Get("aws_region").(string); awsRegion != "" && signAWSRequests {
			log.Printf("[INFO] Using AWS: %+v", awsRegion)
			opts = append(opts, elastic5.SetHttpClient(awsHttpClient(awsRegion, d)), elastic5.SetSniff(false))
		} else if insecure || cacertFile != "" {
			opts = append(opts, elastic5.SetHttpClient(tlsHttpClient(d)), elastic5.SetSniff(false))
		}
		relevantClient, err = elastic5.NewClient(opts...)
		if err != nil {
			return nil, err
		}
	} else if esVersion < "5.0.0" {
		return nil, errors.New("ElasticSearch is older than 5.0.0!")
	}

	return relevantClient, nil
}

func awsSession(region string, d *schema.ResourceData) *awssession.Session {
	aws_access_key_id := d.Get("aws_access_key").(string)
	aws_secret_access_key := d.Get("aws_secret_key").(string)
	aws_session_token := d.Get("aws_token").(string)
	aws_profile := d.Get("aws_profile").(string)

	sessOpts := awssession.Options{
		Config: aws.Config{
			Region: aws.String(region),
		},
	}
	// 1. access keys take priority
	// 2. next is a profile (for assume role)
	// 3. let the default credentials provider figure out the rest (env, ec2, etc..)
	//
	// note: if #1 is chosen, then no further providers will be tested, since we've overridden the credentials with just a static provider
	if aws_access_key_id != "" {
		sessOpts.Config.Credentials = awscredentials.NewStaticCredentials(aws_access_key_id, aws_secret_access_key, aws_session_token)
	} else if aws_profile != "" {
		sessOpts.Profile = aws_profile
	}

	return awssession.Must(awssession.NewSessionWithOptions(sessOpts))
}

func awsHttpClient(region string, d *schema.ResourceData) *http.Client {
	signer := awssigv4.NewSigner(awsSession(region, d).Config.Credentials)
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
