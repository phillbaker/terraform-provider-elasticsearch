package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
	elastic5 "gopkg.in/olivere/elastic.v5"
	elastic6 "gopkg.in/olivere/elastic.v6"
)

func resourceElasticsearchXpackRoleMapping() *schema.Resource {
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
				Type:             schema.TypeString,
				Required:         true,
				DiffSuppressFunc: suppressEquivalentJson,
			},
			"roles": &schema.Schema{
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Required: true,
			},
			"metadata": &schema.Schema{
				Type:             schema.TypeString,
				Optional:         true,
				Default:          "{}",
				DiffSuppressFunc: suppressEquivalentJson,
			},
		},
	}
}

func resourceElasticsearchXpackRoleMappingCreate(d *schema.ResourceData, m interface{}) error {
	name := d.Get("role_mapping_name").(string)

	reqBody, err := buildPutRoleMappingBody(d, m)
	err = xpackPutRoleMapping(d, m, name, reqBody)
	if err != nil {
		return err
	}
	d.SetId(name)
	return resourceElasticsearchXpackRoleMappingRead(d, m)
}

func resourceElasticsearchXpackRoleMappingRead(d *schema.ResourceData, m interface{}) error {

	roleMapping, err := xpackGetRoleMapping(d, m, d.Id())
	if err != nil {
		fmt.Println("Error during read")
		if elasticErr, ok := err.(*elastic6.Error); ok && elasticErr.Status == 404 {
			fmt.Printf("[WARN] Role mapping %s not found. Removing from state\n", d.Id())
			d.SetId("")
			return nil
		}
		if elasticErr, ok := err.(*elastic5.Error); ok && elasticErr.Status == 404 {
			fmt.Printf("[WARN] Role mapping %s not found. Removing from state\n", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}
	d.Set("name", roleMapping.Name)
	d.Set("roles", roleMapping.Roles)
	d.Set("enabled", roleMapping.Enabled)
	d.Set("rules", roleMapping.Rules)
	d.Set("metadata", roleMapping.Metadata)
	return nil
}

func resourceElasticsearchXpackRoleMappingUpdate(d *schema.ResourceData, m interface{}) error {
	name := d.Get("role_mapping_name").(string)

	reqBody, err := buildPutRoleMappingBody(d, m)
	err = xpackPutRoleMapping(d, m, name, reqBody)
	if err != nil {
		return err
	}
	return resourceElasticsearchXpackRoleMappingRead(d, m)
}

func resourceElasticsearchXpackRoleMappingDelete(d *schema.ResourceData, m interface{}) error {

	err := xpackDeleteRoleMapping(d, m, d.Id())
	if err != nil {
		fmt.Println("Error during destroy")
		if elasticErr, ok := err.(*elastic6.Error); ok && elasticErr.Status == 404 {
			fmt.Printf("[WARN] Role mapping %s not found. Resource removed from state\n", d.Id())
			d.SetId("")
			return nil
		}
		if elasticErr, ok := err.(*elastic5.Error); ok && elasticErr.Status == 404 {
			fmt.Printf("[WARN] Role mapping %s not found. Resource removed from state\n", d.Id())
			d.SetId("")
			return nil
		}
	}
	d.SetId("")
	return nil
}

func buildPutRoleMappingBody(d *schema.ResourceData, m interface{}) (string, error) {
	enabled := d.Get("enabled").(bool)
	rules := d.Get("rules").(string)
	roles := expandStringList(d.Get("roles").(*schema.Set).List())
	metadata := d.Get("metadata").(string)

	roleMapping := PutRoleMappingBody{
		Roles:    roles,
		Enabled:  enabled,
		Rules:    json.RawMessage(rules),
		Metadata: optionalInterfaceJson(metadata),
	}

	body, err := json.Marshal(roleMapping)
	if err != nil {
		err = errors.New(fmt.Sprintf("Body Error : %s", body))
	}
	return string(body[:]), err
}

func xpackPutRoleMapping(d *schema.ResourceData, m interface{}, name string, body string) error {
	if client, ok := m.(*elastic6.Client); ok {
		return elastic6PutRoleMapping(client, name, body)
	}
	if client, ok := m.(*elastic5.Client); ok {
		return elastic5PutRoleMapping(client, name, body)
	}
	return errors.New("unhandled client type")
}

func xpackGetRoleMapping(d *schema.ResourceData, m interface{}, name string) (XPackSecurityRoleMapping, error) {
	if client, ok := m.(*elastic6.Client); ok {
		return elastic6GetRoleMapping(client, name)
	}
	if client, ok := m.(*elastic5.Client); ok {
		return elastic5GetRoleMapping(client, name)
	}
	return XPackSecurityRoleMapping{}, errors.New("unhandled client type")
}

func xpackDeleteRoleMapping(d *schema.ResourceData, m interface{}, name string) error {
	if client, ok := m.(*elastic6.Client); ok {
		return elastic6DeleteRoleMapping(client, name)
	}
	if client, ok := m.(*elastic5.Client); ok {
		return elastic5DeleteRoleMapping(client, name)
	}
	return errors.New("unhandled client type")
}

func elastic5PutRoleMapping(client *elastic5.Client, name string, body string) error {
	return errors.New("unsupported in elasticv5 client")
}

func elastic6PutRoleMapping(client *elastic6.Client, name string, body string) error {
	_, err := client.XPackSecurityPutRoleMapping(name).Body(body).Do(context.Background())
	return err
}

func elastic5GetRoleMapping(client *elastic5.Client, name string) (XPackSecurityRoleMapping, error) {
	err := errors.New("unsupported in elasticv5 client")
	return XPackSecurityRoleMapping{}, err
}

func elastic6GetRoleMapping(client *elastic6.Client, name string) (XPackSecurityRoleMapping, error) {
	res, err := client.XPackSecurityGetRoleMapping(name).Do(context.Background())
	if err != nil {
		return XPackSecurityRoleMapping{}, err
	}
	obj := (*res)[name]
	roleMapping := XPackSecurityRoleMapping{}
	roleMapping.Name = name
	roleMapping.Roles = obj.Roles
	roleMapping.Enabled = obj.Enabled
	if rules, err := json.Marshal(obj.Rules); err != nil {
		return roleMapping, err
	} else {
		roleMapping.Rules = string(rules)
	}
	if metadata, err := json.Marshal(obj.Metadata); err != nil {
		return roleMapping, err
	} else {
		roleMapping.Metadata = string(metadata)
	}

	return roleMapping, err
}

func elastic5DeleteRoleMapping(client *elastic5.Client, name string) error {
	err := errors.New("unsupported in elasticv5 client")
	return err
}

func elastic6DeleteRoleMapping(client *elastic6.Client, name string) error {
	_, err := client.XPackSecurityDeleteRoleMapping(name).Do(context.Background())
	return err
}

type PutRoleMappingBody struct {
	Roles    []string    `json:"roles"`
	Enabled  bool        `json:"enabled"`
	Rules    interface{} `json:"rules"`
	Metadata interface{} `json:"metadata,omitempty"`
}

type XPackSecurityRoleMapping struct {
	Name     string   `json:"name"`
	Roles    []string `json:"roles"`
	Enabled  bool     `json:"enabled"`
	Rules    string   `json:"rules"`
	Metadata string   `json:"metadata"`
}
