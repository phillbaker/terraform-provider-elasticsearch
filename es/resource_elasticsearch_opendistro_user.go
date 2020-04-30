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

func resourceElasticsearchOpendistroUser() *schema.Resource {
	return &schema.Resource{
		Create: resourceElasticsearchOpendistroUserCreate,
		Read:   resourceElasticsearchOpendistroUserRead,
		Update: resourceElasticsearchOpendistroUserUpdate,
		Delete: resourceElasticsearchOpendistroUserDelete,
		Schema: map[string]*schema.Schema{
			"username": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"password": &schema.Schema{
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},
			"backend_roles": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"attributes": &schema.Schema{
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
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

func resourceElasticsearchOpendistroUserCreate(d *schema.ResourceData, m interface{}) error {
	_, err := resourceElasticsearchPutOpendistroUser(d, m)

	if err != nil {
		return err
	}

	name := d.Get("username").(string)
	d.SetId(name)
	return resourceElasticsearchOpendistroUserRead(d, m)
}

func resourceElasticsearchOpendistroUserRead(d *schema.ResourceData, m interface{}) error {
	res, err := resourceElasticsearchGetOpendistroUser(d.Id(), m)

	if err != nil {
		if elastic7.IsNotFound(err) {
			log.Printf("[WARN] OpendistroUser (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	d.Set("backend_roles", res.BackendRoles)
	d.Set("attributes", res.Attributes)
	d.Set("description", res.Description)

	return nil
}

func resourceElasticsearchOpendistroUserUpdate(d *schema.ResourceData, m interface{}) error {
	if _, err := resourceElasticsearchPutOpendistroUser(d, m); err != nil {
		return err
	}

	return resourceElasticsearchOpendistroUserRead(d, m)
}

func resourceElasticsearchOpendistroUserDelete(d *schema.ResourceData, m interface{}) error {
	var err error

	path, err := uritemplates.Expand("/_opendistro/_security/api/internalusers/{name}", map[string]string{
		"name": d.Get("username").(string),
	})
	if err != nil {
		return fmt.Errorf("Error building URL path for user: %+v", err)
	}

	switch m.(type) {
	case *elastic7.Client:
		client := m.(*elastic7.Client)
		_, err = client.PerformRequest(context.TODO(), elastic7.PerformRequestOptions{
			Method: "DELETE",
			Path:   path,
		})
	default:
		err = errors.New("Role resource not implemented prior to Elastic v7")
	}

	return err
}

func resourceElasticsearchGetOpendistroUser(userID string, m interface{}) (UserBody, error) {
	var err error
	user := new(UserBody)

	path, err := uritemplates.Expand("/_opendistro/_security/api/internalusers/{name}", map[string]string{
		"name": userID,
	})

	if err != nil {
		return *user, fmt.Errorf("Error building URL path for user: %+v", err)
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

func resourceElasticsearchPutOpendistroUser(d *schema.ResourceData, m interface{}) (*UserResponse, error) {
	response := new(UserResponse)

	userDefinition := UserBody{
		BackendRoles: d.Get("backend_roles").(*schema.Set).List(),
		Description:  d.Get("description").(string),
		Attributes:   d.Get("attributes").(map[string]interface{}),
		Password:     d.Get("password").(string),
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
	switch m.(type) {
	case *elastic7.Client:
		client := m.(*elastic7.Client)
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

// UserBody used by the opendistro's API
type UserBody struct {
	BackendRoles []interface{}          `json:"backend_roles"`
	Attributes   map[string]interface{} `json:"attributes"`
	Description  string                 `json:"description"`
	Password     string                 `json:"password"`
}

// UserResponse sent by the opendistro's API
type UserResponse struct {
	Message string `json:"message"`
	Status  string `json:"status"`
}
