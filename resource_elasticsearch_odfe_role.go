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

func resourceElasticsearchOdfeRole() *schema.Resource {
	return &schema.Resource{
		Create: resourceElasticsearchOdfeRoleCreate,
		Read:   resourceElasticsearchOdfeRoleRead,
		Update: resourceElasticsearchOdfeRoleUpdate,
		Delete: resourceElasticsearchOdfeRoleDelete,
		Schema: map[string]*schema.Schema{
			"role_name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"cluster_permissions": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"index_permissions": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"index_patterns": &schema.Schema{
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"fls": &schema.Schema{
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"masked_fields": &schema.Schema{
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"allowed_actions": &schema.Schema{
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
			"tenant_permissions": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"tenant_patterns": &schema.Schema{
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"allowed_actions": &schema.Schema{
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
			"description": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
		},
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
	}
}

func resourceElasticsearchOdfeRoleCreate(d *schema.ResourceData, m interface{}) error {
	_, err := resourceElasticsearchPutOdfeRole(d, m)

	if err != nil {
		return err
	}

	name := d.Get("role_name").(string)
	d.SetId(name)
	return resourceElasticsearchOdfeRoleRead(d, m)
}

func resourceElasticsearchOdfeRoleRead(d *schema.ResourceData, m interface{}) error {
	res, err := resourceElasticsearchGetOdfeRole(d.Id(), m)

	if err != nil {
		if elastic7.IsNotFound(err) {
			log.Printf("[WARN] OdfeRole (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	d.Set("tenant_permissions", res.TenantPermissions)
	d.Set("cluster_permissions", res.ClusterPermissions)
	d.Set("index_permissions", res.IndexPermissions)
	d.Set("description", res.Description)

	return nil
}

func resourceElasticsearchOdfeRoleUpdate(d *schema.ResourceData, m interface{}) error {
	if _, err := resourceElasticsearchPutOdfeRole(d, m); err != nil {
		return err
	}

	return resourceElasticsearchOdfeRoleRead(d, m)
}

func resourceElasticsearchOdfeRoleDelete(d *schema.ResourceData, m interface{}) error {
	var err error

	path, err := uritemplates.Expand("/_opendistro/_security/api/roles/{name}", map[string]string{
		"name": d.Get("role_name").(string),
	})
	if err != nil {
		return fmt.Errorf("error building URL path for role: %+v", err)
	}

	switch m.(type) {
	case *elastic7.Client:
		client := m.(*elastic7.Client)
		_, err = client.PerformRequest(context.TODO(), elastic7.PerformRequestOptions{
			Method: "DELETE",
			Path:   path,
		})
	default:
		err = errors.New("role resource not implemented prior to Elastic v7")
	}

	return err
}

func resourceElasticsearchGetOdfeRole(roleID string, m interface{}) (RoleBody, error) {
	var err error
	role := new(RoleBody)

	path, err := uritemplates.Expand("/_opendistro/_security/api/roles/{name}", map[string]string{
		"name": roleID,
	})

	if err != nil {
		return *role, fmt.Errorf("error building URL path for role: %+v", err)
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
		err = errors.New("role resource not implemented prior to Elastic v7")
	}

	if err != nil {
		return *role, err
	}
	var roleDefinition map[string]RoleBody

	if err := json.Unmarshal(body, &roleDefinition); err != nil {
		return *role, fmt.Errorf("error unmarshalling role body: %+v: %+v", err, body)
	}

	*role = roleDefinition[roleID]

	return *role, err
}

func resourceElasticsearchPutOdfeRole(d *schema.ResourceData, m interface{}) (*RoleResponse, error) {
	response := new(RoleResponse)

	indexPermissions, err := expandIndexPermissionsSet(d.Get("index_permissions").(*schema.Set).List())
	if err != nil {
		fmt.Print("Error in index get : ", err)
	}
	var indexPermissionsBody []IndexPermissions
	for _, idx := range indexPermissions {
		putIdx := IndexPermissions{
			IndexPatterns:  idx.IndexPatterns,
			Fls:            idx.Fls,
			MaskedFields:   idx.MaskedFields,
			AllowedActions: idx.AllowedActions,
		}
		indexPermissionsBody = append(indexPermissionsBody, putIdx)
	}

	tenantPermissions, err := expandTenantPermissionsSet(d.Get("tenant_permissions").(*schema.Set).List())
	if err != nil {
		fmt.Print("Error in tenant get : ", err)
	}
	var tenantPermissionsBody []TenantPermissions
	for _, tenant := range tenantPermissions {
		putTeanant := TenantPermissions{
			TenantPatterns: tenant.TenantPatterns,
			AllowedActions: tenant.AllowedActions,
		}
		tenantPermissionsBody = append(tenantPermissionsBody, putTeanant)
	}

	rolesDefinition := RoleBody{
		ClusterPermissions: expandStringList(d.Get("cluster_permissions").(*schema.Set).List()),
		IndexPermissions:   indexPermissionsBody,
		TenantPermissions:  tenantPermissionsBody,
		Description:        d.Get("description").(string),
	}

	roleJSON, err := json.Marshal(rolesDefinition)
	if err != nil {
		return response, fmt.Errorf("Body Error : %s", roleJSON)
	}

	path, err := uritemplates.Expand("/_opendistro/_security/api/roles/{name}", map[string]string{
		"name": d.Get("role_name").(string),
	})
	if err != nil {
		return response, fmt.Errorf("error building URL path for role: %+v", err)
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
		err = errors.New("role resource not implemented prior to Elastic v7")
	}

	if err != nil {
		return response, fmt.Errorf("error creating role mapping: %+v: %+v", err, body)
	}

	if err := json.Unmarshal(body, response); err != nil {
		return response, fmt.Errorf("error unmarshalling role body: %+v: %+v", err, body)
	}

	return response, nil
}

type RoleResponse struct {
	Message string `json:"message"`
	Status  string `json:"status"`
}

type RoleBody struct {
	Description        string              `json:"description"`
	ClusterPermissions []string            `json:"cluster_permissions"`
	IndexPermissions   []IndexPermissions  `json:"index_permissions"`
	TenantPermissions  []TenantPermissions `json:"tenant_permissions"`
}

type IndexPermissions struct {
	IndexPatterns  []string `json:"index_patterns"`
	Fls            []string `json:"fls"`
	MaskedFields   []string `json:"masked_fields"`
	AllowedActions []string `json:"allowed_actions"`
}

type TenantPermissions struct {
	TenantPatterns []string `json:"tenant_patterns"`
	AllowedActions []string `json:"allowed_actions"`
}
