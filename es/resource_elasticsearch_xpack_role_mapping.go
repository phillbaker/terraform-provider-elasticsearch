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

func resourceElasticsearchXpackRoleMapping() *schema.Resource {
	return &schema.Resource{
		Description: "Provides an Elasticsearch XPack role mapping resource. Role mappings define which roles are assigned to each user. Each mapping has rules that identify users and a list of roles that are granted to those users. See the upstream [docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/security-api.html) for more details.",
		Create:      resourceElasticsearchXpackRoleMappingCreate,
		Read:        resourceElasticsearchXpackRoleMappingRead,
		Update:      resourceElasticsearchXpackRoleMappingUpdate,
		Delete:      resourceElasticsearchXpackRoleMappingDelete,

		Schema: map[string]*schema.Schema{
			"role_mapping_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The distinct name that identifies the role mapping, used solely as an identifier.",
			},
			"enabled": {
				Type:        schema.TypeBool,
				Default:     true,
				Optional:    true,
				Description: "Mappings that have `enabled` set to `false` are ignored when role mapping is performed.",
			},
			"rules": {
				Type:             schema.TypeString,
				Required:         true,
				DiffSuppressFunc: suppressEquivalentJson,
				Description:      "A list of mustache templates that will be evaluated to determine the roles names that should granted to the users that match the role mapping rules. This matches fields of users, rules can be grouped into `all` and `any` top level keys.",
			},
			"roles": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Required:    true,
				Description: "A list of role names that are granted to the users that match the role mapping rules.",
			},
			"metadata": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          "{}",
				DiffSuppressFunc: suppressEquivalentJson,
				Description:      "Additional metadata that helps define which roles are assigned to each user. Keys beginning with `_` are reserved for system usage.",
			},
		},
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
	}
}

func resourceElasticsearchXpackRoleMappingCreate(d *schema.ResourceData, m interface{}) error {
	name := d.Get("role_mapping_name").(string)

	reqBody, err := buildPutRoleMappingBody(d, m)
	if err != nil {
		return err
	}
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
		if elasticErr, ok := err.(*elastic7.Error); ok && elastic7.IsNotFound(elasticErr) {
			fmt.Printf("[WARN] Role mapping %s not found. Removing from state\n", d.Id())
			d.SetId("")
			return nil
		}
		if elasticErr, ok := err.(*elastic6.Error); ok && elastic6.IsNotFound(elasticErr) {
			fmt.Printf("[WARN] Role mapping %s not found. Removing from state\n", d.Id())
			d.SetId("")
			return nil
		}
		if elasticErr, ok := err.(*elastic5.Error); ok && elastic5.IsNotFound(elasticErr) {
			fmt.Printf("[WARN] Role mapping %s not found. Removing from state\n", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	ds := &resourceDataSetter{d: d}
	ds.set("role_mapping_name", roleMapping.Name)
	ds.set("roles", roleMapping.Roles)
	ds.set("enabled", roleMapping.Enabled)
	ds.set("rules", roleMapping.Rules)
	ds.set("metadata", roleMapping.Metadata)
	return ds.err
}

func resourceElasticsearchXpackRoleMappingUpdate(d *schema.ResourceData, m interface{}) error {
	name := d.Get("role_mapping_name").(string)

	reqBody, err := buildPutRoleMappingBody(d, m)
	if err != nil {
		return err
	}
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
		if elasticErr, ok := err.(*elastic7.Error); ok && elastic7.IsNotFound(elasticErr) {
			fmt.Printf("[WARN] Role mapping %s not found. Resource removed from state\n", d.Id())
			d.SetId("")
			return nil
		}
		if elasticErr, ok := err.(*elastic6.Error); ok && elastic6.IsNotFound(elasticErr) {
			fmt.Printf("[WARN] Role mapping %s not found. Resource removed from state\n", d.Id())
			d.SetId("")
			return nil
		}
		if elasticErr, ok := err.(*elastic5.Error); ok && elastic5.IsNotFound(elasticErr) {
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
		err = fmt.Errorf("Body Error : %s", body)
	}
	return string(body[:]), err
}

func xpackPutRoleMapping(d *schema.ResourceData, m interface{}, name string, body string) error {
	esClient, err := getClient(m.(*ProviderConf))
	if err != nil {
		return err
	}
	switch client := esClient.(type) {
	case *elastic7.Client:
		return elastic7PutRoleMapping(client, name, body)
	case *elastic6.Client:
		return elastic6PutRoleMapping(client, name, body)
	case *elastic5.Client:
		return elastic5PutRoleMapping(client, name, body)
	default:
		return errors.New("unhandled client type")
	}
}

func xpackGetRoleMapping(d *schema.ResourceData, m interface{}, name string) (XPackSecurityRoleMapping, error) {
	esClient, err := getClient(m.(*ProviderConf))
	if err != nil {
		return XPackSecurityRoleMapping{}, err
	}
	switch client := esClient.(type) {
	case *elastic7.Client:
		return elastic7GetRoleMapping(client, name)
	case *elastic6.Client:
		return elastic6GetRoleMapping(client, name)
	case *elastic5.Client:
		return elastic5GetRoleMapping(client, name)
	default:
		return XPackSecurityRoleMapping{}, errors.New("unhandled client type")
	}
}

func xpackDeleteRoleMapping(d *schema.ResourceData, m interface{}, name string) error {
	esClient, err := getClient(m.(*ProviderConf))
	if err != nil {
		return err
	}
	switch client := esClient.(type) {
	case *elastic7.Client:
		return elastic7DeleteRoleMapping(client, name)
	case *elastic6.Client:
		return elastic6DeleteRoleMapping(client, name)
	case *elastic5.Client:
		return elastic5DeleteRoleMapping(client, name)
	default:
		return errors.New("unhandled client type")
	}
}

func elastic5PutRoleMapping(client *elastic5.Client, name string, body string) error {
	return errors.New("unsupported in elasticv5 client")
}

func elastic6PutRoleMapping(client *elastic6.Client, name string, body string) error {
	resp, err := client.XPackSecurityPutRoleMapping(name).Body(body).Do(context.Background())
	log.Printf("[INFO] put error: %+v, %+v", resp, err)
	return err
}

func elastic7PutRoleMapping(client *elastic7.Client, name string, body string) error {
	resp, err := client.XPackSecurityPutRoleMapping(name).Body(body).Do(context.Background())
	log.Printf("[INFO] put error: %+v, %+v", resp, err)
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

func elastic7GetRoleMapping(client *elastic7.Client, name string) (XPackSecurityRoleMapping, error) {
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

func elastic7DeleteRoleMapping(client *elastic7.Client, name string) error {
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
