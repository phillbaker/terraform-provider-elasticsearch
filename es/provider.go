package es

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	awscredentials "github.com/aws/aws-sdk-go/aws/credentials"
	awsstscreds "github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	awssession "github.com/aws/aws-sdk-go/aws/session"
	awssigv4 "github.com/aws/aws-sdk-go/aws/signer/v4"
	awssts "github.com/aws/aws-sdk-go/service/sts"
	"github.com/deoxxa/aws_signing_client"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	elastic7 "github.com/olivere/elastic/v7"
	elastic6 "gopkg.in/olivere/elastic.v6"
)

const (
	// DefaultVersionPingTimeout is the time the ping to check the cluster
	// version waits for a response from Elasticsearch on startup, i.e. when
	// creating a provider.
	DefaultVersionPingTimeout = 5 * time.Second
)

type ServerFlavor int64

// e.g. elasticsearch, opensearch, elasticsearch-oss, etc.
const (
	Unknown ServerFlavor = iota
	Elasticsearch
	ElasticsearchOpenSource
	OpenSearch
)

var awsUrlRegexp = regexp.MustCompile(`([a-z0-9-]+).es.amazonaws.com$`)

type ProviderConf struct {
	rawUrl             string
	insecure           bool
	sniffing           bool
	healthchecking     bool
	cacertFile         string
	username           string
	password           string
	token              string
	tokenName          string
	parsedUrl          *url.URL
	signAWSRequests    bool
	esVersion          string
	awsRegion          string
	awsAssumeRoleArn   string
	awsAccessKeyId     string
	awsSecretAccessKey string
	awsSessionToken    string
	awsSig4Service     string
	awsProfile         string
	certPemPath        string
	keyPemPath         string
	kibanaUrl          string
	hostOverride       string
	// determined after connecting to the server
	flavor ServerFlavor
}

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"url": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("ELASTICSEARCH_URL", nil),
				Description: "Elasticsearch URL",
			},
			"kibana_url": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("KIBANA_URL", nil),
				Description: "URL to reach the Kibana API",
			},
			"sniff": {
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ELASTICSEARCH_SNIFF", false),
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
			"token": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ELASTICSEARCH_TOKEN", nil),
				Description: "A bearer token or ApiKey for an Authorization header, e.g. Active Directory API key.",
			},
			"token_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "ApiKey",
				Description: "The type of token, usually ApiKey or Bearer",
			},
			"aws_assume_role_arn": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "Amazon Resource Name of an IAM Role to assume prior to making AWS API calls.",
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
			"aws_signature_service": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "es",
				Description: "AWS service name used in the credential scope of signed requests to ElasticSearch.",
			},
			"elasticsearch_version": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "ElasticSearch Version",
			},
			"host_override": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "If provided, sets the 'Host' header of requests and the 'ServerName' for certificate validation to this value. See the documentation on connecting to Elasticsearch via an SSH tunnel.",
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"elasticsearch_index":                           resourceElasticsearchIndex(),
			"elasticsearch_index_template":                  resourceElasticsearchIndexTemplate(),
			"elasticsearch_cluster_settings":                resourceElasticsearchClusterSettings(),
			"elasticsearch_composable_index_template":       resourceElasticsearchComposableIndexTemplate(),
			"elasticsearch_component_template":              resourceElasticsearchComponentTemplate(),
			"elasticsearch_data_stream":                     resourceElasticsearchDataStream(),
			"elasticsearch_ingest_pipeline":                 resourceElasticsearchIngestPipeline(),
			"elasticsearch_kibana_alert":                    resourceElasticsearchKibanaAlert(),
			"elasticsearch_kibana_object":                   resourceElasticsearchKibanaObject(),
			"elasticsearch_snapshot_repository":             resourceElasticsearchSnapshotRepository(),
			"elasticsearch_opendistro_destination":          resourceElasticsearchOpenDistroDestination(),
			"elasticsearch_opensearch_destination":          resourceOpenSearchDestination(),
			"elasticsearch_opendistro_ism_policy":           resourceElasticsearchOpenDistroISMPolicy(),
			"elasticsearch_opensearch_ism_policy":           resourceOpenSearchISMPolicy(),
			"elasticsearch_opendistro_ism_policy_mapping":   resourceElasticsearchOpenDistroISMPolicyMapping(),
			"elasticsearch_opensearch_ism_policy_mapping":   resourceOpenSearchISMPolicyMapping(),
			"elasticsearch_opendistro_monitor":              resourceElasticsearchOpenDistroMonitor(),
			"elasticsearch_opensearch_monitor":              resourceOpenSearchMonitor(),
			"elasticsearch_opendistro_roles_mapping":        resourceElasticsearchOpenDistroRolesMapping(),
			"elasticsearch_opensearch_roles_mapping":        resourceOpenSearchRolesMapping(),
			"elasticsearch_opendistro_role":                 resourceElasticsearchOpenDistroRole(),
			"elasticsearch_opensearch_role":                 resourceOpenSearchRole(),
			"elasticsearch_opendistro_user":                 resourceElasticsearchOpenDistroUser(),
			"elasticsearch_opensearch_user":                 resourceOpenSearchUser(),
			"elasticsearch_opendistro_kibana_tenant":        resourceElasticsearchOpenDistroKibanaTenant(),
			"elasticsearch_opensearch_kibana_tenant":        resourceOpenSearchKibanaTenant(),
			"elasticsearch_xpack_index_lifecycle_policy":    resourceElasticsearchXpackIndexLifecyclePolicy(),
			"elasticsearch_xpack_license":                   resourceElasticsearchXpackLicense(),
			"elasticsearch_xpack_role":                      resourceElasticsearchXpackRole(),
			"elasticsearch_xpack_role_mapping":              resourceElasticsearchXpackRoleMapping(),
			"elasticsearch_xpack_snapshot_lifecycle_policy": resourceElasticsearchXpackSnapshotLifecyclePolicy(),
			"elasticsearch_xpack_user":                      resourceElasticsearchXpackUser(),
			"elasticsearch_xpack_watch":                     resourceElasticsearchXpackWatch(),
			"elasticsearch_script":                          resourceElasticsearchScript(),
		},

		DataSourcesMap: map[string]*schema.Resource{
			"elasticsearch_host":                   dataSourceElasticsearchHost(),
			"elasticsearch_opendistro_destination": dataSourceElasticsearchOpenDistroDestination(),
			"elasticsearch_opensearch_destination": dataSourceOpenSearchDestination(),
		},

		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(c context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	rawUrl := d.Get("url").(string)
	parsedUrl, err := url.Parse(rawUrl)
	if err != nil {
		return nil, diag.FromErr(err)
	}

	return &ProviderConf{
		rawUrl:          rawUrl,
		kibanaUrl:       d.Get("kibana_url").(string),
		insecure:        d.Get("insecure").(bool),
		sniffing:        d.Get("sniff").(bool),
		healthchecking:  d.Get("healthcheck").(bool),
		cacertFile:      d.Get("cacert_file").(string),
		username:        d.Get("username").(string),
		password:        d.Get("password").(string),
		token:           d.Get("token").(string),
		tokenName:       d.Get("token_name").(string),
		parsedUrl:       parsedUrl,
		signAWSRequests: d.Get("sign_aws_requests").(bool),
		awsSig4Service:  d.Get("aws_signature_service").(string),
		esVersion:       d.Get("elasticsearch_version").(string),
		awsRegion:       d.Get("aws_region").(string),

		awsAssumeRoleArn:   d.Get("aws_assume_role_arn").(string),
		awsAccessKeyId:     d.Get("aws_access_key").(string),
		awsSecretAccessKey: d.Get("aws_secret_key").(string),
		awsSessionToken:    d.Get("aws_token").(string),
		awsProfile:         d.Get("aws_profile").(string),
		certPemPath:        d.Get("client_cert_path").(string),
		keyPemPath:         d.Get("client_key_path").(string),
		hostOverride:       d.Get("host_override").(string),
	}, nil
}

func getClient(conf *ProviderConf) (interface{}, error) {
	opts := []elastic7.ClientOptionFunc{
		elastic7.SetURL(conf.rawUrl),
		elastic7.SetScheme(conf.parsedUrl.Scheme),
		elastic7.SetSniff(conf.sniffing),
		elastic7.SetHealthcheck(conf.healthchecking),
	}

	if conf.parsedUrl.User.Username() != "" {
		p, _ := conf.parsedUrl.User.Password()
		opts = append(opts, elastic7.SetBasicAuth(conf.parsedUrl.User.Username(), p))
	}
	if conf.username != "" && conf.password != "" {
		opts = append(opts, elastic7.SetBasicAuth(conf.username, conf.password))
	}

	if m := awsUrlRegexp.FindStringSubmatch(conf.parsedUrl.Hostname()); m != nil && conf.signAWSRequests {
		log.Printf("[INFO] Using AWS: %+v", m[1])
		opts = append(opts, elastic7.SetHttpClient(awsHttpClient(m[1], conf, map[string]string{})), elastic7.SetSniff(false))
	} else if awsRegion := conf.awsRegion; conf.awsRegion != "" && conf.signAWSRequests {
		log.Printf("[INFO] Using AWS: %+v", awsRegion)
		opts = append(opts, elastic7.SetHttpClient(awsHttpClient(awsRegion, conf, map[string]string{})), elastic7.SetSniff(false))
	} else if conf.insecure || conf.cacertFile != "" {
		opts = append(opts, elastic7.SetHttpClient(tlsHttpClient(conf, map[string]string{})), elastic7.SetSniff(false))
	} else if conf.token != "" {
		opts = append(opts, elastic7.SetHttpClient(tokenHttpClient(conf, map[string]string{})), elastic7.SetSniff(false))
	} else {
		opts = append(opts, elastic7.SetHttpClient(defaultHttpClient(conf, map[string]string{})))
	}

	logProviderLevel, ok := os.LookupEnv("TF_LOG_PROVIDER")
	if !ok {
		logProviderLevel = "ERROR"
	}
	logProviderLevel = strings.ToUpper(logProviderLevel)

	esLogger := hclog.New(&hclog.LoggerOptions{
		Level:      hclog.LevelFromString(logProviderLevel),
		Output:     os.Stderr,
		JSONFormat: true,
	})
	switch logProviderLevel {
	case "TRACE":
		traceLogger := esLogger.StandardLogger(&hclog.StandardLoggerOptions{
			ForceLevel: hclog.LevelFromString("TRACE"),
		})
		opts = append(opts, elastic7.SetTraceLog(traceLogger))
		fallthrough
	case "INFO":
		infoLogger := esLogger.StandardLogger(&hclog.StandardLoggerOptions{
			ForceLevel: hclog.LevelFromString("INFO"),
		})
		opts = append(opts, elastic7.SetInfoLog(infoLogger))
		fallthrough
	default:
		errorLogger := esLogger.StandardLogger(&hclog.StandardLoggerOptions{
			ForceLevel: hclog.LevelFromString("ERROR"),
		})
		opts = append(opts, elastic7.SetErrorLog(errorLogger))
	}

	var relevantClient interface{}
	client, err := elastic7.NewClient(opts...)
	if err != nil {
		if errors.Is(err, elastic7.ErrNoClient) {
			log.Printf("[INFO] couldn't create client: %T, %s, %T", err, err.Error(), errors.Unwrap(err))
			return nil, errors.New("HEAD healthcheck failed: This is usually due to network or permission issues. The underlying error isn't accessible, please debug by disabling healthchecks.")
		}
		return nil, err
	}
	relevantClient = client

	// Use the v7 client to ping the cluster to determine the version if one was not provided
	if conf.esVersion == "" {
		log.Printf("[INFO] Pinging url to determine version %+v", conf.rawUrl)
		ctx, cancel := context.WithTimeout(context.Background(), DefaultVersionPingTimeout)
		defer cancel()
		info, httpStatus, err := client.Ping(conf.rawUrl).Do(ctx)
		if httpStatus == http.StatusForbidden {
			return nil, errors.New("HTTP 403 Forbidden: Permission denied. Please ensure that the correct credentials are being used to access the cluster.")
		}
		if err != nil {
			return nil, err
		}
		conf.esVersion = info.Version.Number

		// if upstream library exposes support for OpenSearch's distribution
		// param, we can use that as well
		log.Printf("[INFO] ES version %+v", info.Version)
		switch info.Version.BuildFlavor {
		case "default":
			conf.flavor = Elasticsearch
		case "oss":
			conf.flavor = ElasticsearchOpenSource
		}
	}

	if conf.esVersion < "7.0.0" && conf.esVersion >= "6.0.0" {
		log.Printf("[INFO] Using ES 6")
		opts := []elastic6.ClientOptionFunc{
			elastic6.SetURL(conf.rawUrl),
			elastic6.SetScheme(conf.parsedUrl.Scheme),
			elastic6.SetSniff(conf.sniffing),
			elastic6.SetHealthcheck(conf.healthchecking),
		}

		if conf.parsedUrl.User.Username() != "" {
			p, _ := conf.parsedUrl.User.Password()
			opts = append(opts, elastic6.SetBasicAuth(conf.parsedUrl.User.Username(), p))
		}
		if conf.username != "" && conf.password != "" {
			opts = append(opts, elastic6.SetBasicAuth(conf.username, conf.password))
		}

		if m := awsUrlRegexp.FindStringSubmatch(conf.parsedUrl.Hostname()); m != nil && conf.signAWSRequests {
			log.Printf("[INFO] Using AWS: %+v", m[1])
			opts = append(opts, elastic6.SetHttpClient(awsHttpClient(m[1], conf, map[string]string{})), elastic6.SetSniff(false))
		} else if awsRegion := conf.awsRegion; conf.awsRegion != "" && conf.signAWSRequests {
			log.Printf("[INFO] Using AWS: %+v", conf.awsRegion)
			opts = append(opts, elastic6.SetHttpClient(awsHttpClient(awsRegion, conf, map[string]string{})), elastic6.SetSniff(false))
		} else if conf.insecure || conf.cacertFile != "" {
			opts = append(opts, elastic6.SetHttpClient(tlsHttpClient(conf, map[string]string{})), elastic6.SetSniff(false))
		} else if conf.token != "" {
			opts = append(opts, elastic6.SetHttpClient(tokenHttpClient(conf, map[string]string{})), elastic6.SetSniff(false))
		} else {
			opts = append(opts, elastic6.SetHttpClient(defaultHttpClient(conf, map[string]string{})))
		}

		switch logProviderLevel {
		case "TRACE":
			traceLogger := esLogger.StandardLogger(&hclog.StandardLoggerOptions{
				ForceLevel: hclog.LevelFromString("TRACE"),
			})
			opts = append(opts, elastic6.SetTraceLog(traceLogger))
			fallthrough
		case "INFO":
			infoLogger := esLogger.StandardLogger(&hclog.StandardLoggerOptions{
				ForceLevel: hclog.LevelFromString("INFO"),
			})
			opts = append(opts, elastic6.SetInfoLog(infoLogger))
			fallthrough
		default:
			errorLogger := esLogger.StandardLogger(&hclog.StandardLoggerOptions{
				ForceLevel: hclog.LevelFromString("ERROR"),
			})
			opts = append(opts, elastic6.SetErrorLog(errorLogger))
		}

		relevantClient, err = elastic6.NewClient(opts...)
		if err != nil {
			return nil, err
		}
	} else if conf.flavor == Unknown && conf.esVersion < "2.0.0" && conf.esVersion >= "1.0.0" {
		// Version 1.x of OpenSearch very likely. Nothing to do since it's API
		// compatible with 7.x of ES. If elastic client library supports detecting
		// flavor, update to Opensearch.
	} else if conf.esVersion < "6.0.0" {
		return nil, fmt.Errorf("ElasticSearch version %s is older than 6.0.0 and is not supported, flavor: %v.", conf.esVersion, conf.flavor)
	}

	return relevantClient, nil
}

func getKibanaClient(conf *ProviderConf) (interface{}, error) {
	// use either the provided version of elasticsearch or the version of
	// elasticsearch determined by pinging the cluster. Base AWS or other auth
	// off of the same ES config
	esClient, err := getClient(conf)
	if err != nil {
		return nil, err
	}

	switch esClient.(type) {
	case *elastic7.Client:
		opts := []elastic7.ClientOptionFunc{
			elastic7.SetURL(conf.kibanaUrl),
			elastic7.SetScheme(conf.parsedUrl.Scheme),
			// kibana api does not support sniff/health check
			elastic7.SetSniff(false),
			elastic7.SetHealthcheck(false),
		}

		if conf.parsedUrl.User.Username() != "" {
			p, _ := conf.parsedUrl.User.Password()
			opts = append(opts, elastic7.SetBasicAuth(conf.parsedUrl.User.Username(), p))
		}
		if conf.username != "" && conf.password != "" {
			opts = append(opts, elastic7.SetBasicAuth(conf.username, conf.password))
		}

		headers := map[string]string{"kbn-xsrf": "true"}

		if m := awsUrlRegexp.FindStringSubmatch(conf.parsedUrl.Hostname()); m != nil && conf.signAWSRequests {
			log.Printf("[INFO] Using AWS: %+v", m[1])
			opts = append(opts, elastic7.SetHttpClient(awsHttpClient(m[1], conf, headers)), elastic7.SetSniff(false))
		} else if awsRegion := conf.awsRegion; conf.awsRegion != "" && conf.signAWSRequests {
			log.Printf("[INFO] Using AWS: %+v", awsRegion)
			opts = append(opts, elastic7.SetHttpClient(awsHttpClient(awsRegion, conf, headers)), elastic7.SetSniff(false))
		} else if conf.insecure || conf.cacertFile != "" {
			opts = append(opts, elastic7.SetHttpClient(tlsHttpClient(conf, headers)))
		} else if conf.token != "" {
			opts = append(opts, elastic7.SetHttpClient(tokenHttpClient(conf, headers)), elastic7.SetSniff(false))
		} else {
			opts = append(opts, elastic7.SetHttpClient(defaultHttpClient(conf, headers)))
		}

		return elastic7.NewClient(opts...)
	case *elastic6.Client:
		return nil, errors.New("ElasticSearch is older than 6.0.0!")
	default:
		return nil, errors.New("ElasticSearch is older than 5.0.0!")
	}
}

func assumeRoleCredentials(region, roleARN, profile string) *awscredentials.Credentials {
	sessOpts := awsSessionOptions(region)
	sessOpts.Profile = profile

	sess := awssession.Must(awssession.NewSessionWithOptions(sessOpts))
	stsClient := awssts.New(sess)
	assumeRoleProvider := &awsstscreds.AssumeRoleProvider{
		Client:  stsClient,
		RoleARN: roleARN,
	}

	return awscredentials.NewChainCredentials([]awscredentials.Provider{assumeRoleProvider})
}

func awsSessionOptions(region string) awssession.Options {
	return awssession.Options{
		Config: aws.Config{
			Region:   aws.String(region),
			LogLevel: aws.LogLevel(aws.LogDebugWithHTTPBody),
			Logger: aws.LoggerFunc(func(args ...interface{}) {
				log.Print(append([]interface{}{"[DEBUG] "}, args...))
			}),
			CredentialsChainVerboseErrors: aws.Bool(true),
			MaxRetries:                    aws.Int(1),
			// HTTP client is required to fetch EC2 metadata values
			// having zero timeout on the default HTTP client sometimes makes
			// it fail with Credential error
			// https://github.com/aws/aws-sdk-go/issues/2914
			HTTPClient: &http.Client{Timeout: 10 * time.Second},
		},
		SharedConfigState: awssession.SharedConfigEnable,
	}
}

func awsSession(region string, conf *ProviderConf) *awssession.Session {
	sessOpts := awsSessionOptions(region)

	// 1. access keys take priority
	// 2. next is an assume role configuration
	// 3. followed by a profile (for assume role)
	// 4. let the default credentials provider figure out the rest (env, ec2, etc..)
	//
	// note: if #1 is chosen, then no further providers will be tested, since we've overridden the credentials with just a static provider
	if conf.awsAccessKeyId != "" {
		sessOpts.Config.Credentials = awscredentials.NewStaticCredentials(conf.awsAccessKeyId, conf.awsSecretAccessKey, conf.awsSessionToken)
	} else if conf.awsAssumeRoleArn != "" {
		sessOpts.Config.Credentials = assumeRoleCredentials(region, conf.awsAssumeRoleArn, conf.awsProfile)
	} else if conf.awsProfile != "" {
		sessOpts.Profile = conf.awsProfile
	}

	// If configured as insecure, turn off SSL verification
	if conf.insecure {
		client := &http.Client{Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}}
		sessOpts.Config.HTTPClient = client
	} else if conf.hostOverride != "" {
		// Only use `host_override` to set `ServerName` if we're using a secure connection
		client := &http.Client{Transport: &http.Transport{
			TLSClientConfig: &tls.Config{ServerName: conf.hostOverride},
		}}
		sessOpts.Config.HTTPClient = client
	}

	return awssession.Must(awssession.NewSessionWithOptions(sessOpts))
}

func awsHttpClient(region string, conf *ProviderConf, headers map[string]string) *http.Client {
	session := awsSession(region, conf)
	// Call Get() to ensure concurrency safe retrieval of credentials. Since the
	// client is created in many go routines, this synchronizes it.
	_, err := session.Config.Credentials.Get()
	if err != nil {
		log.Fatal(err)
	}
	signer := awssigv4.NewSigner(session.Config.Credentials)
	client, err := aws_signing_client.New(signer, session.Config.HTTPClient, conf.awsSig4Service, region)
	if err != nil {
		log.Fatal(err)
	}

	rt := WithHeader(client.Transport)
	rt.hostOverride = conf.hostOverride
	for k, v := range headers {
		rt.Set(k, v)
	}
	client.Transport = rt

	return client
}

func tokenHttpClient(conf *ProviderConf, headers map[string]string) *http.Client {
	// Setup TLS options
	tlsConfig := &tls.Config{}
	if conf.insecure {
		tlsConfig.InsecureSkipVerify = true
	} else if conf.hostOverride != "" {
		tlsConfig.ServerName = conf.hostOverride
	}

	// Wrapper to inject headers as needed
	transport := &http.Transport{TLSClientConfig: tlsConfig}
	rt := WithHeader(transport)
	rt.hostOverride = conf.hostOverride
	rt.Set("Authorization", fmt.Sprintf("%s %s", conf.tokenName, conf.token))
	for k, v := range headers {
		rt.Set(k, v)
	}

	client := &http.Client{Transport: rt}

	return client
}

func tlsHttpClient(conf *ProviderConf, headers map[string]string) *http.Client {
	// Configure TLS/SSL
	tlsConfig := &tls.Config{}
	if conf.certPemPath != "" && conf.keyPemPath != "" {
		certPem, _, err := readPathOrContent(conf.certPemPath)
		if err != nil {
			log.Fatal(err)
		}
		keyPem, _, err := readPathOrContent(conf.keyPemPath)
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
	if conf.cacertFile != "" {
		caCert, _, _ := readPathOrContent(conf.cacertFile)

		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM([]byte(caCert))
		tlsConfig.RootCAs = caCertPool
	}

	// If configured as insecure, turn off SSL verification
	if conf.insecure {
		tlsConfig.InsecureSkipVerify = true
	} else if conf.hostOverride != "" {
		tlsConfig.ServerName = conf.hostOverride
	}

	transport := &http.Transport{TLSClientConfig: tlsConfig}

	rt := WithHeader(transport)
	rt.hostOverride = conf.hostOverride
	for k, v := range headers {
		rt.Set(k, v)
	}

	client := &http.Client{Transport: rt}

	return client
}

func defaultHttpClient(conf *ProviderConf, headers map[string]string) *http.Client {
	// Setup TLS options
	tlsConfig := &tls.Config{}
	if conf.insecure {
		tlsConfig.InsecureSkipVerify = true
	} else if conf.hostOverride != "" {
		tlsConfig.ServerName = conf.hostOverride
	}

	// Wrapper to inject headers as needed
	transport := &http.Transport{TLSClientConfig: tlsConfig}
	rt := WithHeader(transport)
	rt.hostOverride = conf.hostOverride
	for k, v := range headers {
		rt.Set(k, v)
	}

	client := &http.Client{Transport: rt}
	return client
}
