---
page_title: "Provider: Elasticsearch"
description: |-
  The provider is used to interact with the resources supported by Elasticsearch/Opensearch. The provider needs to be configured with an endpoint URL before it can be used.
---

# Elasticsearch/OpenSearch Provider

The provider is used to interact with the resources supported by
Elasticsearch/Opensearch. The provider needs to be configured with an endpoint
URL before it can be used.

AWS Opensearch Service domains and OpenSearch clusters deployed on Kubernetes and other infrastructure are supported.

Use the navigation to the left to read about the available resources.

## Example Usage

```tf
# Configure the Elasticsearch provider
provider "elasticsearch" {
  url = "http://127.0.0.1:9200"
}

# Create an index template
resource "elasticsearch_index_template" "template_1" {
  name = "template_1"
  body = <<EOF
{
  "template": "te*",
  "settings": {
    "number_of_shards": 1
  },
  "mappings": {
    "type1": {
      "_source": {
        "enabled": false
      },
      "properties": {
        "host_name": {
          "type": "keyword"
        },
        "created_at": {
          "type": "date",
          "format": "EEE MMM dd HH:mm:ss Z YYYY"
        }
      }
    }
  }
}
EOF
}
```

## Argument Reference

The following arguments are supported:

* `url` (Required) - Elasticsearch URL. Defaults to `ELASTICSEARCH_URL` from the environment.
* `kibana_url` (Optional) - URL to reach the Kibana API. Required if using elasticsearch_kibana_* resources.
* `sniff` (Optional) - Set the node sniffing option for the elastic client. Client won't work with sniffing if nodes are not routable. Defaults to `ELASTICSEARCH_SNIFF` from the environment or false.
* `healthcheck` (Optional) - Set the client healthcheck option for the elastic client. Healthchecking is designed for direct access to the cluster. Defaults to `ELASTICSEARCH_HEALTH` from the environment, or true.
* `username` (Optional) - Username to use to connect to elasticsearch using basic auth. Defaults to `ELASTICSEARCH_USERNAME` from the environment
* `password` (Optional) - Password to use to connect to elasticsearch using basic auth. Defaults to `ELASTICSEARCH_PASSWORD` from the environment
* `aws_assume_role_arn` (Optional) - ARN of role to assume when using AWS Elasticsearch Service domains.
* `aws_assume_role_external_id` (Optional) - External ID configured in the IAM policy of the IAM Role to assume prior to using AWS Elasticsearch Service domains.
* `aws_assume_role_session_name` - AWS IAM session name to use when assuming a role.
* `aws_access_key` (Optional) - The access key for use with AWS Elasticsearch Service domains. It can also be sourced from the `AWS_ACCESS_KEY_ID` environment variable.
* `aws_secret_key` (Optional) - The secret key for use with AWS Elasticsearch Service domains. It can also be sourced from the `AWS_SECRET_ACCESS_KEY` environment variable.
* `aws_token` (Optional) - The session token for use with AWS Elasticsearch Service domains. It can also be sourced from the `AWS_SESSION_TOKEN` environment variable.
* `aws_profile` (Optional) - The AWS profile for use with AWS Elasticsearch Service domains
* `aws_region` (Optional) - The AWS region for use in signing of AWS elasticsearch requests. Must be specified in order to use AWS URL signing with AWS ElasticSearch endpoint exposed on a custom DNS domain.
* `token` (Optional) - A bearer token or ApiKey for an Authorization header, e.g. Active Directory API key. See the [docs](https://www.elastic.co/guide/en/elasticsearch/reference/master/token-authentication-services.html). Defaults to `ELASTICSEARCH_TOKEN` from the environment
* `token_name` (Optional) - The type of token, usually ApiKey or Bearer. Defaults to ApiKey.
* `cacert_file` (Optional) - a custom CA certificate when communicating over SSL. You can specify either a path to the file or the contents of the certificate.
* `insecure` (Optional) - Disable SSL verification of API calls (defaults to `false`)
* `client_cert_path` (Optional) - A X509 certificate to connect to elasticsearch. Defaults to `ES_CLIENT_CERTIFICATE_PATH` from the environment
* `client_key_path` (Optional) - A X509 key to connect to elasticsearch. Defaults to `ES_CLIENT_KEY_PATH`
* `sign_aws_requests` (Optional) - Enable signing of AWS elasticsearch requests (defaults to `true`). The `url` must refer to AWS ES domain (`*.<region>.es.amazonaws.com`), or `aws_region` must be specified explicitly.
* `aws_signature_service` (Optional) - AWS service name (e.g. `execute-api` for IAM secured API Gateways) used in the [credential scope](https://docs.aws.amazon.com/general/latest/gr/sigv4_elements.html) of signed requests to ElasticSearch.
* `elasticsearch_version` (Optional) - ElasticSearch Version, if set, skips the version detection at provider start.
* `host_override` (Optional) - If provided, sets the 'Host' header of requests and the 'ServerName' for certificate validation to this value. See the documentation on connecting to Elasticsearch via an SSH tunnel.

### AWS authentication

The provider is flexible in the means of providing credentials for authentication with AWS OpenSearch domains. The following methods are supported, in this order, and explained below:

- Static credentials
- Assume role configuration
- Environment variables
- Shared credentials file

If a [custom domain](https://docs.aws.amazon.com/elasticsearch-service/latest/developerguide/es-customendpoint.html) is being used (instead of the default, of the form `https://search-mydomain-1a2a3a4a5a6a7a8a9a0a9a8a7a.us-east-1.es.amazonaws.com`), please make sure to set `aws_region` in the provider configuration.

#### Static credentials

Static credentials can be provided by adding an `aws_access_key` and `aws_secret_key` in-line in the provider block. If applicable, you may also specify a `aws_token` value.

Example usage:

```tf
provider "elasticsearch" {
    url            = "https://search-foo-bar-pqrhr4w3u4dzervg41frow4mmy.us-east-1.es.amazonaws.com"
    aws_access_key = "anaccesskey"
    aws_secret_key = "asecretkey"
    aws_token      = "" # if necessary
}
```

#### Assume role configuration

You can instruct the provider to assume a role in AWS before interacting with the cluster by setting the `aws_assume_role_arn` variable.
Optionnaly, you can configure the [External ID](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_create_for-user_externalid.html) of IAM role trust policy by setting the `aws_assume_role_external_id` variable.

Example usage:

```tf
provider "elasticsearch" {
    url                         = "https://search-foo-bar-pqrhr4w3u4dzervg41frow4mmy.us-east-1.es.amazonaws.com"
    aws_assume_role_arn         = "arn:aws:iam::012345678901:role/rolename"
    aws_assume_role_external_id = "Unique ID"
}
```

#### Environment variables

You can provide your credentials via the `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY`, environment variables, representing your AWS Access Key and AWS Secret Key. If applicable, the `AWS_SESSION_TOKEN` environment variables is also supported.

Example usage:

```shell
$ export AWS_ACCESS_KEY_ID="anaccesskey"
$ export AWS_SECRET_ACCESS_KEY="asecretkey"
$ terraform plan
```

#### AWS profile

You can specify a named profile that will be used for credentials (either static, or sts assumed role creds).  eg:

```tf
provider "elasticsearch" {
    url         = "https://search-foo-bar-pqrhr4w3u4dzervg41frow4mmy.us-east-1.es.amazonaws.com"
    aws_profile = "profilename"
}
```

#### Shared Credentials file

You can use an AWS credentials file to specify your credentials. The default location is `$HOME/.aws/credentials` on Linux and macOS, or `%USERPROFILE%\.aws\credentials` for Windows users.

Please refer to the official [userguide](https://docs.aws.amazon.com/cli/latest/userguide/cli-config-files.html) for instructions on how to create the credentials file.

### Connecting to a cluster via an SSH Tunnel

If you need to connect to a cluster via an SSH tunnel (for example, to an AWS VPC Cluster), set the following configuration options in your provider:

```tf
provider "elasticsearch" {
  url           = "https://localhost:9999" # Replace 9999 with the port your SSH tunnel is running on
  host_override = "vpc-<******>.us-east-1.es.amazonaws.com"
}
```

The `host_override` flag will set the `Host` header of requests to the cluster and the `ServerName` used for certificate validation. It is recommended to set this flag instead of `insecure = true`, which causes certificate validation to be skipped. Note that if both `host_override` and `insecure = true` are set, certificate validation will be skipped and the `Host` header will be overridden.
