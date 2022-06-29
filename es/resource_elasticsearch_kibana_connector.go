package es

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"

	elastic7 "github.com/olivere/elastic/v7"
	elastic6 "gopkg.in/olivere/elastic.v6"
)

func resourceElasticsearchKibanaConnector() *schema.Resource {
	return &schema.Resource{
		Create: resourceElasticsearchKibanaConnectorCreate,
		Read:   resourceElasticsearchKibanaConnectorRead,
		Update: resourceElasticsearchKibanaConnectorUpdate,
		Delete: resourceElasticsearchKibanaConnectorDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				ForceNew:    true,
				Required:    true,
				Description: "The display name for the connector",
			},
			"connector_type_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The connector type ID for the connector",
			},
			"config": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The configuration for the connector. Configuration properties vary depending on the connector type",
			},
		},
	}
}

func resourceElasticsearchKibanaConnectorCreate(d *schema.ResourceData, meta interface{}) error {

	log.Printf("[INFO] Kibana Connector (%s) created", id)
	d.SetId(id)

	return nil
}