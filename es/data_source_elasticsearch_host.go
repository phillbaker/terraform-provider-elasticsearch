package es

import (
	"errors"
	"reflect"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	elastic7 "github.com/olivere/elastic/v7"
	elastic6 "gopkg.in/olivere/elastic.v6"
)

func dataSourceElasticsearchHost() *schema.Resource {
	return &schema.Resource{
		Description: "`elasticsearch_host` can be used to retrieve the host URL for the provider's current elasticsearch cluster.",
		Read:        dataSourceElasticsearchHostRead,

		Schema: map[string]*schema.Schema{
			"active": {
				Type:        schema.TypeBool,
				Required:    true,
				Description: "should be set to `true`",
			},
			"url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "the url of the active elasticsearch cluster",
			},
		},
	}
}

func dataSourceElasticsearchHostRead(d *schema.ResourceData, m interface{}) error {

	// The upstream elastic client does not export the property for the urls
	// it's using. Presumably the URLS would be available where the client is
	// intantiated, but in terraform, that's not always practicable.
	var err error
	esClient, err := getClient(m.(*ProviderConf))
	if err != nil {
		return err
	}

	var url string
	switch client := esClient.(type) {
	case *elastic7.Client:
		urls := reflect.ValueOf(client).Elem().FieldByName("urls")
		if urls.Len() > 0 {
			url = urls.Index(0).String()
		}
	case *elastic6.Client:
		urls := reflect.ValueOf(client).Elem().FieldByName("urls")
		if urls.Len() > 0 {
			url = urls.Index(0).String()
		}
	default:
		return errors.New("this version of Elasticsearch is not supported")
	}
	d.SetId(url)
	err = d.Set("url", url)

	return err
}
