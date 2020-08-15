package es

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	elastic7 "github.com/olivere/elastic/v7"
	elastic6 "gopkg.in/olivere/elastic.v6"
)

func resourceElasticsearchXpackLicense() *schema.Resource {
	return &schema.Resource{
		Create: resourceElasticsearchLicenseCreate,
		Read:   resourceElasticsearchLicenseRead,
		Update: resourceElasticsearchLicenseUpdate,
		Delete: resourceElasticsearchLicenseDelete,
		Schema: map[string]*schema.Schema{
			"license": {
				Type:             schema.TypeString,
				Optional:         true,
				DiffSuppressFunc: diffSuppressLicense,
			},
			"use_basic_license": {
				Type:     schema.TypeBool,
				Required: true,
			},
			"license_json": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceElasticsearchLicenseCreate(d *schema.ResourceData, meta interface{}) error {
	id, err := resourceElasticsearchCreateXpackLicense(d, meta)
	if err != nil {
		return err
	}

	d.SetId(id)
	return resourceElasticsearchLicenseRead(d, meta)
}

func resourceElasticsearchLicenseUpdate(d *schema.ResourceData, meta interface{}) error {
	_, err := resourceElasticsearchCreateXpackLicense(d, meta)
	if err != nil {
		return err
	}

	return resourceElasticsearchLicenseRead(d, meta)
}

func resourceElasticsearchLicenseRead(d *schema.ResourceData, meta interface{}) error {
	l, err := resourceElasticsearchGetXpackLicense(meta)
	if err != nil {
		return err
	}

	ds := &resourceDataSetter{d: d}
	ds.set("use_basic_license", d.Get("use_basic_license").(bool))
	ds.set("license", d.Get("license").(string))

	out, err := json.Marshal(l)
	if err != nil {
		return err
	}
	ds.set("license_json", string(out))
	return ds.err
}

func resourceElasticsearchLicenseDelete(d *schema.ResourceData, meta interface{}) error {
	var err error

	esClient, err := getClient(meta.(*ProviderConf))
	if err != nil {
		return err
	}
	switch client := esClient.(type) {
	case *elastic7.Client:
		_, err = client.PerformRequest(context.TODO(), elastic7.PerformRequestOptions{
			Method: "DELETE",
			Path:   "/_license",
		})
	case *elastic6.Client:
		_, err = client.PerformRequest(context.TODO(), elastic6.PerformRequestOptions{
			Method: "DELETE",
			Path:   "/_xpack/license",
		})
	default:
		return errors.New("License is only supported by the elasticsearch >= v6!")
	}
	if err != nil {
		return err
	}

	d.SetId("")
	return nil
}

func resourceElasticsearchGetXpackLicense(meta interface{}) (License, error) {
	license := new(License)

	var body json.RawMessage
	var err error

	esClient, err := getClient(meta.(*ProviderConf))
	if err != nil {
		return *license, err
	}
	switch client := esClient.(type) {
	case *elastic7.Client:
		var res *elastic7.Response
		res, err = client.PerformRequest(context.TODO(), elastic7.PerformRequestOptions{
			Method: "GET",
			Path:   "/_license",
		})
		body = res.Body
	case *elastic6.Client:
		var res *elastic6.Response
		res, err = client.PerformRequest(context.TODO(), elastic6.PerformRequestOptions{
			Method: "GET",
			Path:   "/_xpack/license",
		})
		body = res.Body
	default:
		return *license, errors.New("License is only supported by the elasticsearch >= v6!")
	}

	if err != nil {
		return *license, err
	}
	var licenseResponse map[string]License

	if err := json.Unmarshal(body, &licenseResponse); err != nil {
		return *license, fmt.Errorf("Error unmarshalling license body: %+v: %+v", err, body)
	}

	return licenseResponse["license"], err
}

func resourceElasticsearchCreateXpackLicense(d *schema.ResourceData, meta interface{}) (string, error) {
	license := d.Get("license").(string)
	useBasicLicense := d.Get("use_basic_license").(bool)

	var l License
	var err error
	if !useBasicLicense {
		l, err = resourceElasticsearchPutEnterpriseLicense(license, meta)
	} else if d.Id() == "" {
		l, err = resourceElasticsearchPostBasicLicense(meta)
	} else {
		log.Printf("[INFO] skipping creating basic license because already enabled %s", d.Id())
	}

	if err != nil {
		return "", err
	}

	return l.Uid, nil
}

func resourceElasticsearchPutEnterpriseLicense(l string, meta interface{}) (License, error) {
	request := fmt.Sprintf(`{"licenses": [%s]}`, l)

	var emptyLicense License
	var body json.RawMessage
	var err error
	esClient, err := getClient(meta.(*ProviderConf))
	if err != nil {
		return emptyLicense, err
	}
	switch client := esClient.(type) {
	case *elastic7.Client:
		var res *elastic7.Response
		res, err = client.PerformRequest(context.TODO(), elastic7.PerformRequestOptions{
			Method: "PUT",
			Path:   "/_license?acknowledge=true",
			Body:   request,
		})
		body = res.Body
	case *elastic6.Client:
		var res *elastic6.Response
		res, err = client.PerformRequest(context.TODO(), elastic6.PerformRequestOptions{
			Method: "PUT",
			Path:   "/_xpack/license?acknowledge=true",
			Body:   request,
		})
		body = res.Body
	default:
		return emptyLicense, errors.New("License is only supported by the elastic library >= v6!")
	}

	if err != nil {
		return emptyLicense, err
	}
	var licenseResponse map[string][]License

	if err := json.Unmarshal(body, &licenseResponse); err != nil {
		return emptyLicense, fmt.Errorf("Error unmarshalling license body: %+v: %+v", err, body)
	}

	return licenseResponse["licenses"][0], err
}

func resourceElasticsearchPostBasicLicense(meta interface{}) (License, error) {
	var l License
	var err error
	esClient, err := getClient(meta.(*ProviderConf))
	if err != nil {
		return l, err
	}
	switch client := esClient.(type) {
	case *elastic7.Client:
		_, err = client.PerformRequest(context.TODO(), elastic7.PerformRequestOptions{
			Method: "POST",
			Path:   "/_license/start_basic?acknowledge=true",
		})
	case *elastic6.Client:
		_, err = client.PerformRequest(context.TODO(), elastic6.PerformRequestOptions{
			Method: "POST",
			Path:   "/_xpack/license/start_basic?acknowledge=true",
		})
	default:
		return l, errors.New("License is only supported by the elastic library >= v6!")
	}

	if err != nil {
		return l, err
	}
	return resourceElasticsearchGetXpackLicense(meta)
}

type License struct {
	Status             string `json:"status,omitempty"`
	Uid                string `json:"uid,omitempty"`
	Type               string `json:"type,omitempty"`
	IssueDate          string `json:"issue_date,omitempty"`
	IssueDateInMillis  int    `json:"issue_date_in_millis,omitempty"`
	ExpiryDate         string `json:"expiry_date,omitempty"`
	ExpiryDateInMillis int    `json:"expiry_date_in_millis,omitempty"`
	MaxNodes           int    `json:"max_nodes,omitempty"`
	IssuedTo           string `json:"issued_to,omitempty"`
	Issuer             string `json:"issuer,omitempty"`
	StartDateInMillis  int    `json:"start_date_in_millis,omitempty"`
}
