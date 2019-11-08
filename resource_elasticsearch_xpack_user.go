package main

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

func onlyDiffOnCreate(_, _, _ string, d *schema.ResourceData) bool {
	return d.Id() != ""
}

func resourceElasticsearchXpackUser() *schema.Resource {
	return &schema.Resource{
		Create: resourceElasticsearchXpackUserCreate,
		Read:   resourceElasticsearchXpackUserRead,
		Update: resourceElasticsearchXpackUserUpdate,
		Delete: resourceElasticsearchXpackUserDelete,

		Schema: map[string]*schema.Schema{
			"username": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"fullname": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Required: false,
			},
			"email": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Required: false,
			},
			"enabled": &schema.Schema{
				Type:     schema.TypeBool,
				Default:  true,
				Optional: true,
				Required: false,
			},
			"password": &schema.Schema{
				Type:             schema.TypeString,
				Sensitive:        true,
				Required:         false,
				Optional:         true,
				DiffSuppressFunc: onlyDiffOnCreate,
			},
			"password_hash": &schema.Schema{
				Type:             schema.TypeString,
				Required:         false,
				Sensitive:        true,
				Optional:         true,
				DiffSuppressFunc: onlyDiffOnCreate,
			},
			"roles": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: false,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"metadata": &schema.Schema{
				Type:             schema.TypeString,
				Default:          "{}",
				Optional:         true,
				DiffSuppressFunc: suppressEquivalentJson,
			},
		},
	}
}

func resourceElasticsearchXpackUserCreate(d *schema.ResourceData, m interface{}) error {
	name := d.Get("username").(string)

	reqBody, err := buildPutUserBody(d, m)

	err = xpackPutUser(d, m, name, reqBody)
	if err != nil {
		return err
	}
	d.SetId(name)
	return resourceElasticsearchXpackUserRead(d, m)
}

func resourceElasticsearchXpackUserRead(d *schema.ResourceData, m interface{}) error {

	user, err := xpackGetUser(d, m, d.Id())
	if err != nil {
		fmt.Println("Error during read")
		if elasticErr, ok := err.(*elastic7.Error); ok && elasticErr.Status == 404 {
			fmt.Printf("[WARN] User %s not found. Removing from state\n", d.Id())
			d.SetId("")
			return nil
		}
		if elasticErr, ok := err.(*elastic6.Error); ok && elasticErr.Status == 404 {
			fmt.Printf("[WARN] User %s not found. Removing from state\n", d.Id())
			d.SetId("")
			return nil
		}
		if elasticErr, ok := err.(*elastic5.Error); ok && elasticErr.Status == 404 {
			fmt.Printf("[WARN] User %s not found. Removing from state\n", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}
	d.Set("username", user.Username)
	d.Set("roles", user.Roles)
	d.Set("fullname", user.Fullname)
	d.Set("email", user.Email)
	d.Set("metadata", user.Metadata)
	d.Set("enabled", user.Enabled)
	return nil
}

func resourceElasticsearchXpackUserUpdate(d *schema.ResourceData, m interface{}) error {
	name := d.Get("username").(string)

	reqBody, err := buildPutUserBody(d, m)
	err = xpackPutUser(d, m, name, reqBody)
	if err != nil {
		return err
	}
	return resourceElasticsearchXpackUserRead(d, m)
}

func resourceElasticsearchXpackUserDelete(d *schema.ResourceData, m interface{}) error {

	err := xpackDeleteUser(d, m, d.Id())
	if err != nil {
		fmt.Println("Error during destroy")
		if elasticErr, ok := err.(*elastic7.Error); ok && elasticErr.Status == 404 {
			fmt.Printf("[WARN] User %s not found. Resource removed from state\n", d.Id())
			d.SetId("")
			return nil
		}
		if elasticErr, ok := err.(*elastic6.Error); ok && elasticErr.Status == 404 {
			fmt.Printf("[WARN] User %s not found. Resource removed from state\n", d.Id())
			d.SetId("")
			return nil
		}
		if elasticErr, ok := err.(*elastic5.Error); ok && elasticErr.Status == 404 {
			fmt.Printf("[WARN] User %s not found. Resource removed from state\n", d.Id())
			d.SetId("")
			return nil
		}
	}
	d.SetId("")
	return nil
}

func buildPutUserBody(d *schema.ResourceData, m interface{}) (string, error) {
	roles := expandStringList(d.Get("roles").(*schema.Set).List())
	username := d.Get("username").(string)
	fullname := d.Get("fullname").(string)
	password := d.Get("password").(string)
	passwordHash := d.Get("password_hash").(string)
	email := d.Get("email").(string)
	enabled := d.Get("enabled").(bool)
	metadata := d.Get("metadata").(string)

	user := XPackSecurityUser{
		Username: username,
		Roles:    roles,
		Fullname: fullname,
		Password: password,
		Email:    email,
		Enabled:  enabled,
		Metadata: optionalInterfaceJson(metadata),
	}
	if password == "" {
		user.PasswordHash = passwordHash
	}

	body, err := json.Marshal(user)
	if err != nil {
		fmt.Printf("Body : %s", body)
		err = errors.New(fmt.Sprintf("Body Error : %s", body))
	}
	log.Printf("[INFO] put body: %+v", body)
	return string(body[:]), err
}

func xpackPutUser(d *schema.ResourceData, m interface{}, name string, body string) error {
	if client, ok := m.(*elastic7.Client); ok {
		return elastic7PutUser(client, name, body)
	}
	if client, ok := m.(*elastic6.Client); ok {
		return elastic6PutUser(client, name, body)
	}
	if client, ok := m.(*elastic5.Client); ok {
		return elastic5PutUser(client, name, body)
	}
	return errors.New("unhandled client type")
}

func xpackGetUser(d *schema.ResourceData, m interface{}, name string) (XPackSecurityUser, error) {
	if client, ok := m.(*elastic7.Client); ok {
		return elastic7GetUser(client, name)
	}
	if client, ok := m.(*elastic6.Client); ok {
		return elastic6GetUser(client, name)
	}
	if client, ok := m.(*elastic5.Client); ok {
		return elastic5GetUser(client, name)
	}
	return XPackSecurityUser{}, errors.New("unhandled client type")
}

func xpackDeleteUser(d *schema.ResourceData, m interface{}, name string) error {
	if client, ok := m.(*elastic7.Client); ok {
		return elastic7DeleteUser(client, name)
	}
	if client, ok := m.(*elastic6.Client); ok {
		return elastic6DeleteUser(client, name)
	}
	if client, ok := m.(*elastic5.Client); ok {
		return elastic5DeleteUser(client, name)
	}
	return errors.New("unhandled client type")
}

func elastic5PutUser(client *elastic5.Client, name string, body string) error {
	return errors.New("unsupported in elasticv5 client")
}

func elastic6PutUser(client *elastic6.Client, name string, body string) error {
	return errors.New("unsupported in elasticv6 client")
}

func elastic7PutUser(client *elastic7.Client, name string, body string) error {
	_, err := client.XPackSecurityPutUser(name).Body(body).Do(context.Background())
	log.Printf("[INFO] put error: %+v", err)
	return err
}

func elastic5GetUser(client *elastic5.Client, name string) (XPackSecurityUser, error) {
	err := errors.New("unsupported in elasticv5 client")
	return XPackSecurityUser{}, err
}

func elastic6GetUser(client *elastic6.Client, name string) (XPackSecurityUser, error) {
	err := errors.New("unsupported in elasticv6 client")
	return XPackSecurityUser{}, err
}

func elastic7GetUser(client *elastic7.Client, name string) (XPackSecurityUser, error) {
	res, err := client.XPackSecurityGetUser(name).Do(context.Background())
	if err != nil {
		return XPackSecurityUser{}, err
	}
	obj := (*res)[name]
	user := XPackSecurityUser{}
	user.Username = name
	user.Roles = obj.Roles
	user.Fullname = obj.Fullname
	user.Email = obj.Email
	user.Enabled = obj.Enabled
	if metadata, err := json.Marshal(obj.Metadata); err != nil {
		return user, err
	} else {
		user.Metadata = string(metadata)
	}
	return user, err
}

func elastic5DeleteUser(client *elastic5.Client, name string) error {
	err := errors.New("unsupported in elasticv5 client")
	return err
}

func elastic6DeleteUser(client *elastic6.Client, name string) error {
	err := errors.New("unsupported in elasticv5 client")
	return err
}

func elastic7DeleteUser(client *elastic7.Client, name string) error {
	_, err := client.XPackSecurityDeleteUser(name).Do(context.Background())
	return err
}

// XPackSecurityUser is the user object.
//
// we want to define a new struct as the one from elastic has metadata as
// a map[string]interface{} but we want to manage string only
type XPackSecurityUser struct {
	Username     string      `json:"username"`
	Roles        []string    `json:"roles"`
	Fullname     string      `json:"full_name,omitempty"`
	Email        string      `json:"email,omitempty"`
	Metadata     interface{} `json:"metadata,omitempty"`
	Enabled      bool        `json:"enabled,omitempty"`
	Password     string      `json:"password,omitempty"`
	PasswordHash string      `json:"password_hash,omitempty"`
}
