package es

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

func resourceElasticsearchOpenDistroUser() *schema.Resource {
	return &schema.Resource{
		Create: resourceElasticsearchOpenDistroUserCreate,
		Read:   resourceElasticsearchOpenDistroUserRead,
		Update: resourceElasticsearchOpenDistroUserUpdate,
		Delete: resourceElasticsearchOpenDistroUserDelete,
		Schema: map[string]*schema.Schema{
			"username": {
				Type:     schema.TypeString,
				Required: true,
			},
			"password": {
				Type:             schema.TypeString,
				Optional:         true,
				Sensitive:        true,
				DiffSuppressFunc: onlyDiffOnCreate,
				ConflictsWith:    []string{"password_hash"},
			},
			"password_hash": {
				Type:             schema.TypeString,
				Optional:         true,
				Sensitive:        true,
				DiffSuppressFunc: onlyDiffOnCreate,
				ConflictsWith:    []string{"password"},
			},
			"backend_roles": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"attributes": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
	}
}

func resourceElasticsearchOpenDistroUserCreate(d *schema.ResourceData, m interface{}) error {
	_, err := resourceElasticsearchPutOpenDistroUser(d, m)

	if err != nil {
		return err
	}

	name := d.Get("username").(string)
	d.SetId(name)
	return resourceElasticsearchOpenDistroUserRead(d, m)
}

func resourceElasticsearchOpenDistroUserRead(d *schema.ResourceData, m interface{}) error {
	res, err := resourceElasticsearchGetOpenDistroUser(d.Id(), m)

	if err != nil {
		if elastic7.IsNotFound(err) {
			log.Printf("[WARN] OdfeUser (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	ds := &resourceDataSetter{d: d}
	ds.set("backend_roles", res.BackendRoles)
	ds.set("attributes", res.Attributes)
	ds.set("description", res.Description)
	return ds.err
}

func resourceElasticsearchOpenDistroUserUpdate(d *schema.ResourceData, m interface{}) error {
	if _, err := resourceElasticsearchPutOpenDistroUser(d, m); err != nil {
		return err
	}

	return resourceElasticsearchOpenDistroUserRead(d, m)
}

func resourceElasticsearchOpenDistroUserDelete(d *schema.ResourceData, m interface{}) error {
	var err error

	path, err := uritemplates.Expand("/_opendistro/_security/api/internalusers/{name}", map[string]string{
		"name": d.Get("username").(string),
	})
	if err != nil {
		return fmt.Errorf("Error building URL path for user: %+v", err)
	}

	switch client := m.(type) {
	case *elastic7.Client:
		_, err = client.PerformRequest(context.TODO(), elastic7.PerformRequestOptions{
			Method: "DELETE",
			Path:   path,
		})
	default:
		err = errors.New("Role resource not implemented prior to Elastic v7")
	}

	return err
}

func resourceElasticsearchGetOpenDistroUser(userID string, m interface{}) (UserBody, error) {
	var err error
	user := new(UserBody)

	path, err := uritemplates.Expand("/_opendistro/_security/api/internalusers/{name}", map[string]string{
		"name": userID,
	})

	if err != nil {
		return *user, fmt.Errorf("Error building URL path for user: %+v", err)
	}

	var body json.RawMessage
	switch client := m.(type) {
	case *elastic7.Client:
		var res *elastic7.Response
		res, err = client.PerformRequest(context.TODO(), elastic7.PerformRequestOptions{
			Method: "GET",
			Path:   path,
		})
		body = res.Body
	default:
		err = errors.New("Role resource not implemented prior to Elastic v7")
	}

	if err != nil {
		return *user, err
	}
	var userDefinition map[string]UserBody

	if err := json.Unmarshal(body, &userDefinition); err != nil {
		return *user, fmt.Errorf("Error unmarshalling user body: %+v: %+v", err, body)
	}

	*user = userDefinition[userID]

	return *user, err
}

func resourceElasticsearchPutOpenDistroUser(d *schema.ResourceData, m interface{}) (*UserResponse, error) {
	response := new(UserResponse)

	userDefinition := UserBody{
		BackendRoles: d.Get("backend_roles").(*schema.Set).List(),
		Description:  d.Get("description").(string),
		Attributes:   d.Get("attributes").(map[string]interface{}),
		Password:     d.Get("password").(string),
		PasswordHash: d.Get("password_hash").(string),
	}

	userJSON, err := json.Marshal(userDefinition)
	if err != nil {
		return response, fmt.Errorf("Body Error : %s", userJSON)
	}

	path, err := uritemplates.Expand("/_opendistro/_security/api/internalusers/{name}", map[string]string{
		"name": d.Get("username").(string),
	})
	if err != nil {
		return response, fmt.Errorf("Error building URL path for user: %+v", err)
	}

	var body json.RawMessage
	switch client := m.(type) {
	case *elastic7.Client:
		var res *elastic7.Response
		res, err = client.PerformRequest(context.TODO(), elastic7.PerformRequestOptions{
			Method: "PUT",
			Path:   path,
			Body:   string(userJSON),
		})
		body = res.Body
	default:
		err = errors.New("User resource not implemented prior to Elastic v7")
	}

	if err != nil {
		return response, fmt.Errorf("Error creating user mapping: %+v: %+v: %+v", err, body, string(userJSON))
	}

	if err := json.Unmarshal(body, response); err != nil {
		return response, fmt.Errorf("Error unmarshalling user body: %+v: %+v", err, body)
	}

	return response, nil
}

// UserBody used by the odfe's API
type UserBody struct {
	BackendRoles []interface{}          `json:"backend_roles"`
	Attributes   map[string]interface{} `json:"attributes"`
	Description  string                 `json:"description"`
	Password     string                 `json:"password"`
	PasswordHash string                 `json:"hash"`
}

// UserResponse sent by the odfe's API
type UserResponse struct {
	Message string `json:"message"`
	Status  string `json:"status"`
}
