# terraform-provider-elasticsearch

[![Build Status](https://travis-ci.org/phillbaker/terraform-provider-elasticsearch.svg?branch=master)](https://travis-ci.org/phillbaker/terraform-provider-elasticsearch)

This is a terraform provider that lets you provision elasticsearch resources, compatible with v5 of elasticsearch. Based off of an [original PR to Terraform](https://github.com/hashicorp/terraform/pull/13238).

## Installation

[Download a binary](https://github.com/phillbaker/terraform-provider-elasticsearch/releases), and put it in a good spot on your system. Then update your `~/.terraformrc` to refer to the binary:

```hcl
providers {
  elasticsearch = "/path/to/terraform-provider-elasticsearch"
}
```

See [the docs for more information](https://www.terraform.io/docs/plugins/basics.html).

## Usage

```
provider "elasticsearch" {
    url = ""
    aws_access_key = ""
    aws_secret_key = ""
    aws_token = "" # if necessary
}

resource "elasticsearch_index_template" "test" {
  name = "terraform-test"
  template = "tf-*"
  order = 1

  settings {
    index.number_of_shards = 5
  }

  mapping {
    type = "reports"

    date_property {
      name = "created_at"
      format = "EEE MMM dd HH:mm:ss Z YYYY"
    }

    keyword_property {
      name = "subject"
    }
  }
}

resource "elasticsearch_kibana_visualization" "test" {
}

resource "elasticsearch_kibana_dashboard" "test" {
}
```

## Development

```
go get github.com/phillbaker/terraform-provider-elasticsearch
gopkg.in/olivere/elastic.v5/uritemplates cf4e58efdcee2e8e7c18dad44d51ed166fb256c2
gopkg.in/olivere/elastic.v5 f698dfea7c6cb058bee5de042f1ad3387f678ab1
go get github.com/deoxxa/aws_signing_client # c20ee106809eacdffcc81ac7cb984261f8e3067e

cd $GOPATH/src/github.com/phillbaker/terraform-provider-elasticsearch
go build -o /path/to/binary/terraform-provider-elasticsearch
```

## Licence

See LICENSE.

## Contributing

1. Fork it ( https://github.com/phillbaker/terraform-provider-elasticsearch/fork )
2. Create your feature branch (`git checkout -b my-new-feature`)
3. Commit your changes (`git commit -am 'Add some feature'`)
4. Push to the branch (`git push origin my-new-feature`)
5. Create a new Pull Request
