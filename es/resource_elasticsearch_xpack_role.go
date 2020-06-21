package es

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	elastic7 "github.com/olivere/elastic/v7"
	elastic5 "gopkg.in/olivere/elastic.v5"
	elastic6 "gopkg.in/olivere/elastic.v6"
)

func resourceElasticsearchXpackRole() *schema.Resource {
	return &schema.Resource{
		Create: resourceElasticsearchXpackRoleCreate,
		Read:   resourceElasticsearchXpackRoleRead,
		Update: resourceElasticsearchXpackRoleUpdate,
		Delete: resourceElasticsearchXpackRoleDelete,

		Schema: map[string]*schema.Schema{
			"role_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"indices": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"names": {
							Type:     schema.TypeSet,
							Required: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"privileges": {
							Type:     schema.TypeSet,
							Required: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"query": {
							Type:             schema.TypeString,
							Optional:         true,
							DiffSuppressFunc: suppressEquivalentJson,
						},
						"field_security": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"grant": {
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
									"except": {
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
								},
							},
						},
					},
				},
			},
			"applications": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"application": {
							Type:     schema.TypeString,
							Required: true,
						},
						"privileges": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"resources": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
			"cluster": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"global": {
				Type:             schema.TypeString,
				Optional:         true,
				DiffSuppressFunc: suppressEquivalentJson,
			},
			"run_as": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"metadata": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          "{}",
				DiffSuppressFunc: suppressEquivalentJson,
			},
		},
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
	}
}

func resourceElasticsearchXpackRoleCreate(d *schema.ResourceData, m interface{}) error {
	name := d.Get("role_name").(string)

	reqBody, err := buildPutRoleBody(d, m)
	if err != nil {
		return err
	}
	err = xpackPutRole(d, m, name, reqBody)
	if err != nil {
		return err
	}
	d.SetId(name)
	return resourceElasticsearchXpackRoleRead(d, m)
}

func resourceElasticsearchXpackRoleRead(d *schema.ResourceData, m interface{}) error {

	role, err := xpackGetRole(d, m, d.Id())

	if err != nil {
		log.Print("Error during read")
		if elasticErr, ok := err.(*elastic7.Error); ok && elasticErr.Status == 404 {
			log.Printf("[WARN] Role %s not found. Removing from state\n", d.Id())
			d.SetId("")
			return nil
		}
		if elasticErr, ok := err.(*elastic6.Error); ok && elasticErr.Status == 404 {
			log.Printf("[WARN] Role %s not found. Removing from state\n", d.Id())
			d.SetId("")
			return nil
		}
		if elasticErr, ok := err.(*elastic5.Error); ok && elasticErr.Status == 404 {
			log.Printf("[WARN] Role %s not found. Removing from state\n", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	ds := &resourceDataSetter{d: d}
	ds.set("role_name", d.Id())

	if len(role.Indices) > 0 {
		indices := make([]map[string]interface{}, 0, len(role.Indices))
		for _, v := range role.Indices {
			ip := map[string]interface{}{
				"names":          v.Names,
				"privileges":     v.Privileges,
				"field_security": v.FieldSecurity,
				"query":          v.Query,
			}
			indices = append(indices, ip)
		}
		ds.set("indices", indices)
	}

	ds.set("cluster", role.Cluster)

	if len(role.Applications) > 0 {
		applications := make([]map[string]interface{}, 0, len(role.Applications))
		for _, va := range role.Applications {
			ap := map[string]interface{}{
				"application": va.Application,
				"privileges":  va.Privileges,
				"resources":   va.Resources,
			}
			applications = append(applications, ap)
		}
		ds.set("applications", applications)
	}

	ds.set("global", role.Global)
	ds.set("run_as", role.RunAs)
	ds.set("metadata", role.Metadata)
	return ds.err
}

func resourceElasticsearchXpackRoleUpdate(d *schema.ResourceData, m interface{}) error {
	name := d.Get("role_name").(string)

	reqBody, err := buildPutRoleBody(d, m)
	if err != nil {
		return err
	}
	err = xpackPutRole(d, m, name, reqBody)
	if err != nil {
		return err
	}
	return resourceElasticsearchXpackRoleRead(d, m)
}

func resourceElasticsearchXpackRoleDelete(d *schema.ResourceData, m interface{}) error {

	err := xpackDeleteRole(d, m, d.Id())
	if err != nil {
		log.Print("Error during destroy")
		if elasticErr, ok := err.(*elastic7.Error); ok && elasticErr.Status == 404 {
			log.Printf("[WARN] Role %s not found. Resource removed from state\n", d.Id())
			d.SetId("")
			return nil
		}
		if elasticErr, ok := err.(*elastic6.Error); ok && elasticErr.Status == 404 {
			log.Printf("[WARN] Role %s not found. Resource removed from state\n", d.Id())
			d.SetId("")
			return nil
		}
		if elasticErr, ok := err.(*elastic5.Error); ok && elasticErr.Status == 404 {
			log.Printf("[WARN] Role %s not found. Resource removed from state\n", d.Id())
			d.SetId("")
			return nil
		}
	}
	d.SetId("")
	return nil
}

func buildPutRoleBody(d *schema.ResourceData, m interface{}) (string, error) {
	clusterPrivileges := expandStringList(d.Get("cluster").(*schema.Set).List())
	applications, err := expandApplicationPermissionSet(d.Get("applications").(*schema.Set).List())
	if err != nil {
		log.Printf("Error in application get : %v", err)
		return "", err
	}
	var applicationsBody []PutRoleApplicationPrivileges
	for _, app := range applications {
		putApp := PutRoleApplicationPrivileges(app)
		applicationsBody = append(applicationsBody, putApp)
	}

	indicesPrivileges, err := expandIndicesPermissionSet(d.Get("indices").(*schema.Set).List())
	if err != nil {
		log.Printf("Error in indices get : %v", err)
		return "", err
	}

	var indicesBody []PutRoleIndicesPermissions
	for _, indice := range indicesPrivileges {
		putIndex := PutRoleIndicesPermissions{
			Names:         indice.Names,
			Privileges:    indice.Privileges,
			FieldSecurity: indice.FieldSecurity,
			Query:         optionalInterfaceJson(indice.Query.(string)),
		}
		indicesBody = append(indicesBody, putIndex)
	}

	runAs := expandStringList(d.Get("run_as").(*schema.Set).List())
	global := d.Get("global").(string)
	metadata := d.Get("metadata").(string)

	role := PutRoleBody{
		Cluster:      clusterPrivileges,
		Applications: applicationsBody,
		Indices:      indicesBody,
		RunAs:        runAs,
		Global:       optionalInterfaceJson(global),
		Metadata:     optionalInterfaceJson(metadata),
	}

	body, err := json.Marshal(role)
	if err != nil {
		log.Printf("Body : %s", body)
		err = fmt.Errorf("Body Error : %s", body)
	}
	return string(body[:]), err
}

func xpackPutRole(d *schema.ResourceData, m interface{}, name string, body string) error {
	if client, ok := m.(*elastic7.Client); ok {
		return elastic7PutRole(client, name, body)
	}
	if client, ok := m.(*elastic6.Client); ok {
		return elastic6PutRole(client, name, body)
	}
	if client, ok := m.(*elastic5.Client); ok {
		return elastic5PutRole(client, name, body)
	}
	return errors.New("unhandled client type")
}

func xpackGetRole(d *schema.ResourceData, m interface{}, name string) (XPackSecurityRole, error) {
	if client, ok := m.(*elastic7.Client); ok {
		return elastic7GetRole(client, name)
	}
	if client, ok := m.(*elastic6.Client); ok {
		return elastic6GetRole(client, name)
	}
	if client, ok := m.(*elastic5.Client); ok {
		return elastic5GetRole(client, name)
	}
	return XPackSecurityRole{}, errors.New("unhandled client type")
}

func xpackDeleteRole(d *schema.ResourceData, m interface{}, name string) error {
	if client, ok := m.(*elastic7.Client); ok {
		return elastic7DeleteRole(client, name)
	}
	if client, ok := m.(*elastic6.Client); ok {
		return elastic6DeleteRole(client, name)
	}
	if client, ok := m.(*elastic5.Client); ok {
		return elastic5DeleteRole(client, name)
	}
	return errors.New("unhandled client type")
}

func elastic5PutRole(client *elastic5.Client, name string, body string) error {
	return errors.New("unsupported in elasticv5 client")
}

func elastic6PutRole(client *elastic6.Client, name string, body string) error {
	_, err := client.XPackSecurityPutRole(name).Body(body).Do(context.Background())
	log.Printf("[INFO] put error: %+v", err)
	return err
}

func elastic7PutRole(client *elastic7.Client, name string, body string) error {
	_, err := client.XPackSecurityPutRole(name).Body(body).Do(context.Background())
	log.Printf("[INFO] put error: %+v", err)
	return err
}

func elastic5GetRole(client *elastic5.Client, name string) (XPackSecurityRole, error) {
	err := errors.New("unsupported in elasticv5 client")
	return XPackSecurityRole{}, err
}

func elastic6GetRole(client *elastic6.Client, name string) (XPackSecurityRole, error) {
	res, err := client.XPackSecurityGetRole(name).Do(context.Background())
	if err != nil {
		return XPackSecurityRole{}, err
	}
	obj := (*res)[name]
	role := XPackSecurityRole{}
	role.Name = name
	role.Cluster = obj.Cluster

	// if we have field security settings, we have to flatten them for tf
	if len(obj.Indices) > 0 {
		if data, err := flattenIndicesPermissionSetv6(obj.Indices); err == nil {
			role.Indices = data
		} else {
			log.Printf("[INFO] Data: %+v", data)
			return role, err
		}
	}

	if data, err := json.Marshal(obj.Applications); err == nil {
		if err := json.Unmarshal(data, &role.Applications); err != nil {
			log.Printf("[INFO] Data: %+v", data)
			return role, err
		}
	}
	if global, err := json.Marshal(obj.Global); err != nil {
		return role, err
	} else {
		// The Elastic API will not return the field unless it exists, which force us to check for null compared to Metadata
		if string(global) == "null" {
			role.Global = ""
		} else {
			role.Global = string(global)
		}
	}
	if metadata, err := json.Marshal(obj.Metadata); err != nil {
		return role, err
	} else {
		role.Metadata = string(metadata)
	}
	return role, err
}

func elastic7GetRole(client *elastic7.Client, name string) (XPackSecurityRole, error) {
	res, err := client.XPackSecurityGetRole(name).Do(context.Background())
	if err != nil {
		return XPackSecurityRole{}, err
	}
	obj := (*res)[name]
	role := XPackSecurityRole{}
	role.Name = name
	role.Cluster = obj.Cluster

	// if we have field security settings, we have to flatten them for tf
	if len(obj.Indices) > 0 {
		if data, err := flattenIndicesPermissionSetv7(obj.Indices); err == nil {
			role.Indices = data
		} else {
			log.Printf("Data: %v\n", data)

			return role, err
		}
	}

	if data, err := json.Marshal(obj.Applications); err == nil {
		if err := json.Unmarshal(data, &role.Applications); err != nil {
			log.Printf("Data : %s\n", data)
			return role, err
		}
	}
	if global, err := json.Marshal(obj.Global); err != nil {
		return role, err
	} else {
		// The Elastic API will not return the field unless it exists, which force us to check for null compared to Metadata
		if string(global) == "null" {
			role.Global = ""
		} else {
			role.Global = string(global)
		}
	}
	if metadata, err := json.Marshal(obj.Metadata); err != nil {
		return role, err
	} else {
		role.Metadata = string(metadata)
	}
	return role, err
}

func elastic5DeleteRole(client *elastic5.Client, name string) error {
	err := errors.New("unsupported in elasticv5 client")
	return err
}

func elastic6DeleteRole(client *elastic6.Client, name string) error {
	_, err := client.XPackSecurityDeleteRole(name).Do(context.Background())
	return err
}

func elastic7DeleteRole(client *elastic7.Client, name string) error {
	_, err := client.XPackSecurityDeleteRole(name).Do(context.Background())
	return err
}

type PutRoleBody struct {
	Cluster      []string                       `json:"cluster"`
	Applications []PutRoleApplicationPrivileges `json:"applications,omitempty"`
	Indices      []PutRoleIndicesPermissions    `json:"indices,omitempty"`
	RunAs        []string                       `json:"run_as,omitempty"`
	Global       interface{}                    `json:"global,omitempty"`
	Metadata     interface{}                    `json:"metadata,omitempty"`
}

type PutRoleApplicationPrivileges struct {
	Application string   `json:"application"`
	Privileges  []string `json:"privileges,omitempty"`
	Resources   []string `json:"resources,omitempty"`
}

type PutRoleIndicesPermissions struct {
	Names         []string            `json:"names"`
	Privileges    []string            `json:"privileges"`
	FieldSecurity map[string][]string `json:"field_security,omitempty"`
	Query         interface{}         `json:"query,omitempty"`
}

type XPackSecurityRole struct {
	Name         string                               `json:"name"`
	Cluster      []string                             `json:"cluster"`
	Indices      []XPackSecurityIndicesPermissions    `json:"indices"`
	Applications []XPackSecurityApplicationPrivileges `json:"applications"`
	RunAs        []string                             `json:"run_as"`
	Global       string                               `json:"global"`
	Metadata     string                               `json:"metadata"`
}

// XPackSecurityApplicationPrivileges is the application privileges object of Elasticsearch
type XPackSecurityApplicationPrivileges struct {
	Application string   `json:"application"`
	Privileges  []string `json:"privileges"`
	Resources   []string `json:"resources"`
}

// XPackSecurityIndicesPermissions is the indices permission object of Elasticsearch
type XPackSecurityIndicesPermissions struct {
	Names         []string                 `json:"names"`
	Privileges    []string                 `json:"privileges"`
	FieldSecurity []map[string]interface{} `json:"field_security"`
	Query         string                   `json:"query"`
}
