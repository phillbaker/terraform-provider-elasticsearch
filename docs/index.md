---
page_title: "Provider: Elasticsearch"
description: |-
  The Elasticsearch provider is used to interact with the resources supported by Elasticsearch. The provider needs to be configured with an endpoint URL before it can be used.
---

# Elasticsearch Provider

The Elasticsearch provider is used to interact with the
resources supported by Elasticsearch. The provider needs
to be configured with an endpoint URL before it can be used.

AWS Elasticsearch Service domains are supported.

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
* `sniff` (Optional) - Set the node sniffing option for the elastic client. Client won't work with sniffing if nodes are not routable. Defaults to `ELASTICSEARCH_SNIFF` from the environment or true.
* `healthcheck` (Optional) - Set the client healthcheck option for the elastic client. Healthchecking is designed for direct access to the cluster. Defaults to `ELASTICSEARCH_HEALTH` from the environment, or true.
* `username` (Optional) - Username to use to connect to elasticsearch using basic auth. Defaults to `ELASTICSEARCH_USERNAME` from the environment
* `password` (Optional) - Password to use to connect to elasticsearch using basic auth. Defaults to `ELASTICSEARCH_PASSWORD` from the environment
* `aws_assume_role_arn` (Optional) - ARN of role to assume when using AWS Elasticsearch Service domains.
* `aws_access_key` (Optional) - The access key for use with AWS Elasticsearch Service domains. It can also be sourced from the `AWS_ACCESS_KEY_ID` environment variable.
* `aws_secret_key` (Optional) - The secret key for use with AWS Elasticsearch Service domains. It can also be sourced from the `AWS_SECRET_ACCESS_KEY` environment variable.
* `aws_token` (Optional) - The session token for use with AWS Elasticsearch Service domains. It can also be sourced from the `AWS_SESSION_TOKEN` environment variable.
* `aws_profile` (Optional) - The AWS profile for use with AWS Elasticsearch Service domains
* `aws_region` (Optional) - The AWS region for use in signing of AWS elasticsearch requests. Must be specified in order to use AWS URL signing with AWS ElasticSearch endpoint exposed on a custom DNS domain.
* `cacert_file` (Optional) - a custom CA certificate when communicating over SSL. You can specify either a path to the file or the contents of the certificate.
* `insecure` (Optional) - Disable SSL verification of API calls (defaults to `false`)
* `client_cert_path` (Optional) - A X509 certificate to connect to elasticsearch. Defaults to `ES_CLIENT_CERTIFICATE_PATH` from the environment
* `client_key_path` (Optional) - A X509 key to connect to elasticsearch. Defaults to `ES_CLIENT_KEY_PATH`
* `sign_aws_requests` (Optional) - Enable signing of AWS elasticsearch requests (defauls to `true`). The `url` must refer to AWS ES domain (`*.<region>.es.amazonaws.com`), or `aws_region` must be specified explicitly.
* `elasticsearch_version` (Optional) - ElasticSearch Version, if set, skips the version detection at provider start.

### AWS authentication

The Elasticsearch provider is flexible in the means of providing credentials for authentication with AWS Elasticsearch domains. The following methods are supported, in this order, and explained below:

- Static credentials
- Assume role configuration
- Environment variables
- Shared credentials file

#### Static credentials

Static credentials can be provided by adding an `aws_access_key` and `aws_secret_key` in-line in the Elasticsearch provider block. If applicable, you may also specify a `aws_token` value.

Example usage:

```tf
provider "elasticsearch" {
    url = "https://search-foo-bar-pqrhr4w3u4dzervg41frow4mmy.us-east-1.es.amazonaws.com"
    aws_access_key = "anaccesskey"
    aws_secret_key = "asecretkey"
    aws_token = "" # if necessary
}
```

#### Assume role configuration

You can instruct the provider to assume a role in AWS before interacting with Elasticsearch by setting the `aws_assume_role_arn` variable.

Example usage:

```tf
provider "elasticsearch" {
    url = "https://search-foo-bar-pqrhr4w3u4dzervg41frow4mmy.us-east-1.es.amazonaws.com"
    aws_assume_role_arn = "arn:aws:iam::012345678901:role/rolename`
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
