package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"log"

	elastic6 "gopkg.in/coveo/elasticsearch-client-go.v6"
	elastic5 "gopkg.in/olivere/elastic.v5"
)

func resourceElasticsearchXpackRoleMapping() *schema.Resource {
	log.Printf("Started the Resource making")
	return &schema.Resource{
		Create: resourceElasticsearchXpackRoleMappingCreate,
		Read:   resourceElasticsearchXpackRoleMappingRead,
		Update: resourceElasticsearchXpackRoleMappingUpdate,
		Delete: resourceElasticsearchXpackRoleMappingDelete,

		Schema: map[string]*schema.Schema{
			"role_mapping_name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"enabled": &schema.Schema{
				Type:     schema.TypeBool,
				Default:  true,
				Optional: true,
			},
			"rules": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"roles": &schema.Schema{
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Required: true,
			},
		},
	}
}

func resourceElasticsearchXpackRoleMappingCreate(d *schema.ResourceData, m interface{}) error {
	var err error
	name := d.Get("role_mapping_name").(string)
	err = xpackPutRoleMapping(d, m, name, []byte("reqBody"))
	//_ , err = buildPutRoleMappingBody(d, m)
	if err != nil {
		return err
	}
	/*

		d.SetId(name)
	*/
	return resourceElasticsearchXpackRoleMappingRead(d, m)
}

func resourceElasticsearchXpackRoleMappingRead(d *schema.ResourceData, m interface{}) error {
	return nil
}

func resourceElasticsearchXpackRoleMappingUpdate(d *schema.ResourceData, m interface{}) error {
	return resourceElasticsearchXpackRoleMappingRead(d, m)
}

func resourceElasticsearchXpackRoleMappingDelete(d *schema.ResourceData, m interface{}) error {
	return nil
}

func buildPutRoleMappingBody(d *schema.ResourceData, m interface{}) ([]byte, error) {
	enabled := d.Get("enabled").(bool)
	rules := d.Get("rules").(string)
	roles := d.Get("roles").([]string)

	body, err := json.Marshal(PutRoleMappingBody{roles, enabled, rules, ""})
	if err != nil {
		err = errors.New(fmt.Sprintf("Body Error : %s", body))
	}
	return body, err
}

func xpackPutRoleMapping(d *schema.ResourceData, m interface{}, name string, body []byte) error {

	var err error
	client := m.(*elastic6.Client)
	err = elastic6PutRoleMapping(client, name, body)

	switch m.(type) {
	case *elastic6.Client:
		client := m.(*elastic6.Client)
		err = elastic6PutRoleMapping(client, name, body)
	case *elastic5.Client:
		client := m.(*elastic5.Client)
		err = elastic5PutRoleMapping(client, name, body)
	default:
		err = errors.New("unhandled client type") //TODO: Add verbosity
	}

	return err
}

func elastic5PutRoleMapping(client *elastic5.Client, name string, body []byte) error {
	err := errors.New("unsupported in elasticv5 client")
	return err
}

func elastic6PutRoleMapping(client *elastic6.Client, name string, body []byte) error {
	_, err := client.XPackSecurityPutRoleMapping(name).Body(body).Do(context.TODO())
	return err
}

type PutRoleMappingBody struct {
	Roles    []string    `json:"roles"`
	Enabled  bool        `json:"enabled"`
	Rules    interface{} `json:"rules"`
	Metadata interface{} `json:"metadata"`
}
