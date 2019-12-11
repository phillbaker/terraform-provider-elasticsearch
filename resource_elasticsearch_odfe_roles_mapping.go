package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/olivere/elastic/uritemplates"

	elastic7 "github.com/olivere/elastic/v7"
)

func resourceElasticsearchOdfeRolesMapping() *schema.Resource {
	return &schema.Resource{
		Create: resourceElasticsearchOdfeRolesMappingCreate,
		Read:   resourceElasticsearchOdfeRolesMappingRead,
		Update: resourceElasticsearchOdfeRolesMappingUpdate,
		Delete: resourceElasticsearchOdfeRolesMappingDelete,
		Schema: map[string]*schema.Schema{
			"role_name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"backend_roles": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"hosts": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"users": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"description": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"and_backend_roles": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
	}
}

func resourceElasticsearchOdfeRolesMappingCreate(d *schema.ResourceData, m interface{}) error {
	if _, err := resourceElasticsearchPutOdfeRolesMapping(d, m); err != nil {
		log.Printf("[INFO] Failed to put role mapping: %+v", err)
		return err
	}

	name := d.Get("role_name").(string)

	d.SetId(name)
	return resourceElasticsearchOdfeRolesMappingRead(d, m)
}

func resourceElasticsearchOdfeRolesMappingRead(d *schema.ResourceData, m interface{}) error {
	res, err := resourceElasticsearchGetOdfeRolesMapping(d.Id(), m)

	if err != nil {
		if elastic7.IsNotFound(err) {
			log.Printf("[WARN] OdfeRolesMapping (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	d.Set("backend_roles", res.BackendRoles)
	d.Set("hosts", res.Hosts)
	d.Set("users", res.Users)
	d.Set("description", res.Description)
	d.Set("and_backend_roles", res.AndBackendRoles)

	return nil
}

func resourceElasticsearchOdfeRolesMappingUpdate(d *schema.ResourceData, m interface{}) error {
	if _, err := resourceElasticsearchPutOdfeRolesMapping(d, m); err != nil {
		return err
	}

	return resourceElasticsearchOdfeRolesMappingRead(d, m)
}

func resourceElasticsearchOdfeRolesMappingDelete(d *schema.ResourceData, m interface{}) error {
	path, err := uritemplates.Expand("/_opendistro/_security/api/rolesmapping/{name}", map[string]string{
		"name": d.Get("role_name").(string),
	})
	if err != nil {
		return fmt.Errorf("error building URL path for role mapping: %+v", err)
	}

	switch m.(type) {
	case *elastic7.Client:
		client := m.(*elastic7.Client)
		_, err = client.PerformRequest(context.TODO(), elastic7.PerformRequestOptions{
			Method: "DELETE",
			Path:   path,
		})
	default:
		err = errors.New("role mapping resource not implemented prior to Elastic v7")
	}

	return err
}

func resourceElasticsearchGetOdfeRolesMapping(roleID string, m interface{}) (RolesMapping, error) {
	var err error
	var roleMapping = new(RolesMapping)

	path, err := uritemplates.Expand("/_opendistro/_security/api/rolesmapping/{name}", map[string]string{
		"name": roleID,
	})

	if err != nil {
		return *roleMapping, fmt.Errorf("error building URL path for role mapping: %+v", err)
	}
	var body json.RawMessage
	switch m.(type) {
	case *elastic7.Client:
		client := m.(*elastic7.Client)
		var res *elastic7.Response
		res, err = client.PerformRequest(context.TODO(), elastic7.PerformRequestOptions{
			Method: "GET",
			Path:   path,
		})
		body = res.Body
	default:
		err = errors.New("role mapping  resource not implemented prior to Elastic v7")
	}

	if err != nil {
		return *roleMapping, err
	}
	var rolesMappingDefinition map[string]RolesMapping

	if err := json.Unmarshal(body, &rolesMappingDefinition); err != nil {
		return *roleMapping, fmt.Errorf("error unmarshalling role mapping body: %+v: %+v", err, body)
	}

	*roleMapping = rolesMappingDefinition[roleID]

	return *roleMapping, err
}

func resourceElasticsearchPutOdfeRolesMapping(d *schema.ResourceData, m interface{}) (*RoleMappingResponse, error) {
	var err error
	response := new(RoleMappingResponse)

	rolesMappingDefinition := RolesMapping{
		BackendRoles:    expandStringList(d.Get("backend_roles").(*schema.Set).List()),
		Hosts:           expandStringList(d.Get("hosts").(*schema.Set).List()),
		Users:           expandStringList(d.Get("users").(*schema.Set).List()),
		Description:     d.Get("description").(string),
		AndBackendRoles: expandStringList(d.Get("and_backend_roles").(*schema.Set).List()),
	}
	roleJSON, err := json.Marshal(rolesMappingDefinition)

	if err != nil {
		return response, fmt.Errorf("Body Error : %s", roleJSON)
	}

	path, err := uritemplates.Expand("/_opendistro/_security/api/rolesmapping/{name}", map[string]string{
		"name": d.Get("role_name").(string),
	})

	if err != nil {
		return response, fmt.Errorf("error building URL path for role mapping: %+v", err)
	}

	var body json.RawMessage
	switch m.(type) {
	case *elastic7.Client:
		client := m.(*elastic7.Client)
		var res *elastic7.Response
		res, err = client.PerformRequest(context.TODO(), elastic7.PerformRequestOptions{
			Method: "PUT",
			Path:   path,
			Body:   string(roleJSON),
		})
		body = res.Body
	default:
		err = errors.New("role mapping resource not implemented prior to Elastic v7")
	}

	if err != nil {
		return response, fmt.Errorf("error creating role mapping: %+v: %+v", err, body)
	}

	if err := json.Unmarshal(body, response); err != nil {
		return response, fmt.Errorf("error unmarshalling role mapping body: %+v: %+v", err, body)
	}

	return response, nil
}

type RoleMappingResponse struct {
	Message string `json:"message"`
	Status  string `json:"status"`
}

type RolesMapping struct {
	BackendRoles    []string `json:"backend_roles"`
	Hosts           []string `json:"hosts"`
	Users           []string `json:"users"`
	Description     string   `json:"description"`
	AndBackendRoles []string `json:"and_backend_roles"`
}
