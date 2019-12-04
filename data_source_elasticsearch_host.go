package main

import (
	"reflect"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	elastic7 "github.com/olivere/elastic/v7"
	elastic5 "gopkg.in/olivere/elastic.v5"
	elastic6 "gopkg.in/olivere/elastic.v6"
)

func dataSourceElasticsearchHost() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceElasticsearchHostRead,

		Schema: map[string]*schema.Schema{
			"active": &schema.Schema{
				Type:     schema.TypeBool,
				Required: true,
			},
			"url": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceElasticsearchHostRead(d *schema.ResourceData, m interface{}) error {

	// The upstream elastic client does not export the property for the urls
	// it's using. Presumably the URLS would be available where the client is
	// intantiated, but in terraform, that's not always practicable.

	var err error
	switch m.(type) {
	case *elastic7.Client:
		client := m.(*elastic7.Client)

		urls := reflect.ValueOf(client).Elem().FieldByName("urls")
		if urls.Len() > 0 {
			d.SetId(urls.Index(0).String())
		}
	case *elastic6.Client:
		client := m.(*elastic6.Client)

		urls := reflect.ValueOf(client).Elem().FieldByName("urls")
		if urls.Len() > 0 {
			d.SetId(urls.Index(0).String())
		}
	default:
		client := m.(*elastic5.Client)

		urls := reflect.ValueOf(client).Elem().FieldByName("urls")
		if urls.Len() > 0 {
			d.SetId(urls.Index(0).String())
		}
	}

	return err
}
