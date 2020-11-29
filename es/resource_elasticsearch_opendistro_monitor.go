package es

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/olivere/elastic/uritemplates"

	elastic7 "github.com/olivere/elastic/v7"
	elastic6 "gopkg.in/olivere/elastic.v6"
)

var openDistroMonitorSchema = map[string]*schema.Schema{
	"name": {
		Type:          schema.TypeString,
		Optional:      true,
		ConflictsWith: []string{"body"},
	},
	"enabled": {
		Type:          schema.TypeBool,
		Description:   "A boolean that indicates that the monitor should evaluated",
		Default:       true,
		Optional:      true,
		ConflictsWith: []string{"body"},
	},
	"schedule_interval": {
		Type:          schema.TypeString,
		Optional:      true,
		ConflictsWith: []string{"body"},
	},
	"schedule_unit": {
		Type:          schema.TypeString,
		Optional:      true,
		ConflictsWith: []string{"body"},
	},
	"inputs": {
		Type:     schema.TypeSet,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"search_indices": {
					Type:     schema.TypeSet,
					Optional: true,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
					Set: schema.HashString,
				},
				"search_query": {
					Type:     schema.TypeString,
					Required: true,
					StateFunc: func(v interface{}) string {
						json, _ := structure.NormalizeJsonString(v)
						return json
					},
					ValidateFunc: validation.StringIsJSON,
				},
				// OpenDistro seems to add default values to the object after the resource
				// is saved, e.g. adjust_pure_negative, boost values, so we save the
				// response in a separate attribute
				"search_query_response": {
					Type:        schema.TypeString,
					Description: "The value of the search query as returned",
					Computed:    true,
				},
				"terminate_after": {
					Type:     schema.TypeInt,
					Optional: true,
				},
			},
		},
		// Set:           monitorInputsHash,
		ConflictsWith: []string{"body"},
	},
	"triggers": {
		Type:     schema.TypeSet,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"severity": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"condition_script": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"condition_language": {
					Type:     schema.TypeString,
					Optional: true,
					Default:  "painless",
				},
				"actions": {
					Type:     schema.TypeSet,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"name": {
								Type:     schema.TypeString,
								Optional: true,
							},
							"destination_id": {
								Type:     schema.TypeString,
								Optional: true,
							},
							"message_template": {
								Type:     schema.TypeString,
								Optional: true,
							},
							"message_language": {
								Type:     schema.TypeString,
								Optional: true,
								Default:  "mustache",
							},
							"subject_template": {
								Type:     schema.TypeString,
								Optional: true,
							},
							"subject_language": {
								Type:     schema.TypeString,
								Optional: true,
								Default:  "mustache",
							},
							"throttle_enabled": {
								Type:     schema.TypeBool,
								Default:  false,
								Optional: true,
							},
						},
					},
					// Set: monitorActionsHash,
				},
			},
		},
		// Set:           monitorTriggersHash,
		ConflictsWith: []string{"body"},
	},
	"body": {
		Type:             schema.TypeString,
		Optional:         true,
		DiffSuppressFunc: diffSuppressMonitor,
		StateFunc: func(v interface{}) string {
			json, _ := structure.NormalizeJsonString(v)
			return json
		},
		ValidateFunc:  validation.StringIsJSON,
		ConflictsWith: []string{"name", "enabled", "schedule_interval", "schedule_unit", "inputs", "triggers"},
	},
}

func resourceElasticsearchDeprecatedMonitor() *schema.Resource {
	return &schema.Resource{
		Create: resourceElasticsearchOpenDistroMonitorCreate,
		Read:   resourceElasticsearchOpenDistroMonitorRead,
		Update: resourceElasticsearchOpenDistroMonitorUpdate,
		Delete: resourceElasticsearchOpenDistroMonitorDelete,
		Schema: openDistroMonitorSchema,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		DeprecationMessage: "elasticsearch_monitor is deprecated, please use elasticsearch_opendistro_monitor resource instead.",
	}
}

func resourceElasticsearchOpenDistroMonitor() *schema.Resource {
	return &schema.Resource{
		Create: resourceElasticsearchOpenDistroMonitorCreate,
		Read:   resourceElasticsearchOpenDistroMonitorRead,
		Update: resourceElasticsearchOpenDistroMonitorUpdate,
		Delete: resourceElasticsearchOpenDistroMonitorDelete,
		Schema: openDistroMonitorSchema,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
	}
}

func resourceElasticsearchOpenDistroMonitorCreate(d *schema.ResourceData, m interface{}) error {
	res, err := resourceElasticsearchOpenDistroPostMonitor(d, m)

	if err != nil {
		log.Printf("[INFO] Failed to put monitor: %+v", err)
		return err
	}

	d.SetId(res.ID)
	log.Printf("[INFO] Object ID: %s", d.Id())

	// Although we receive the full monitor in the response to the POST,
	// OpenDistro seems to add default values to the ojbect after the resource
	// is saved, e.g. adjust_pure_negative, boost values
	return resourceElasticsearchOpenDistroMonitorRead(d, m)
}

func resourceElasticsearchOpenDistroMonitorRead(d *schema.ResourceData, m interface{}) error {
	res, err := resourceElasticsearchOpenDistroGetMonitor(d.Id(), m)

	if elastic6.IsNotFound(err) || elastic7.IsNotFound(err) {
		log.Printf("[WARN] Monitor (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return err
	}

	d.SetId(res.ID)

	monitorJson, err := json.Marshal(res.Monitor)
	if err != nil {
		return err
	}
	monitorJsonNormalized, err := structure.NormalizeJsonString(string(monitorJson))
	if err != nil {
		return err
	}
	err = d.Set("body", monitorJsonNormalized)
	return err
}

func resourceElasticsearchOpenDistroMonitorUpdate(d *schema.ResourceData, m interface{}) error {
	_, err := resourceElasticsearchOpenDistroPutMonitor(d, m)

	if err != nil {
		return err
	}

	return resourceElasticsearchOpenDistroMonitorRead(d, m)
}

func resourceElasticsearchOpenDistroMonitorDelete(d *schema.ResourceData, m interface{}) error {
	var err error

	path, err := uritemplates.Expand("/_opendistro/_alerting/monitors/{id}", map[string]string{
		"id": d.Id(),
	})
	if err != nil {
		return fmt.Errorf("error building URL path for monitor: %+v", err)
	}

	esClient, err := getClient(m.(*ProviderConf))
	if err != nil {
		return err
	}
	switch client := esClient.(type) {
	case *elastic7.Client:
		_, err = client.PerformRequest(context.TODO(), elastic7.PerformRequestOptions{
			Method: "DELETE",
			Path:   path,
		})
	case *elastic6.Client:
		_, err = client.PerformRequest(context.TODO(), elastic6.PerformRequestOptions{
			Method: "DELETE",
			Path:   path,
		})
	default:
		err = errors.New("monitor resource not implemented prior to Elastic v6")
	}

	return err
}

func resourceElasticsearchOpenDistroGetMonitor(monitorID string, m interface{}) (*monitorResponse, error) {
	var err error
	response := new(monitorResponse)

	path, err := uritemplates.Expand("/_opendistro/_alerting/monitors/{id}", map[string]string{
		"id": monitorID,
	})
	if err != nil {
		return response, fmt.Errorf("error building URL path for monitor: %+v", err)
	}

	var body json.RawMessage
	esClient, err := getClient(m.(*ProviderConf))
	if err != nil {
		return nil, err
	}
	switch client := esClient.(type) {
	case *elastic7.Client:
		var res *elastic7.Response
		res, err = client.PerformRequest(context.TODO(), elastic7.PerformRequestOptions{
			Method: "GET",
			Path:   path,
		})
		body = res.Body
	case *elastic6.Client:
		var res *elastic6.Response
		res, err = client.PerformRequest(context.TODO(), elastic6.PerformRequestOptions{
			Method: "GET",
			Path:   path,
		})
		body = res.Body
	default:
		err = errors.New("monitor resource not implemented prior to Elastic v6")
	}

	if err != nil {
		return response, err
	}

	if err := json.Unmarshal(body, response); err != nil {
		return response, fmt.Errorf("error unmarshalling monitor body: %+v: %+v", err, body)
	}
	normalizeMonitor(response.Monitor)
	return response, err
}

func resourceElasticsearchOpenDistroPostMonitor(d *schema.ResourceData, m interface{}) (*monitorResponse, error) {
	var err error
	response := new(monitorResponse)

	monitorJSON := d.Get("body").(string)
	if monitorJSON == "" {
		// we have to build the json from the attributes
		monitorDefinition := monitor{
			Name:    d.Get("name").(string),
			Enabled: d.Get("enabled").(bool),
			Schedule: monitorSchedule{
				Period: monitorPeriod{
					Interval: d.Get("schedule_interval").(int),
					Unit:     d.Get("schedule_unit").(string),
				},
			},
			// Inputs: ,
			// Triggers: ,
		}

		monitorJSONBytes, err := json.Marshal(monitorDefinition)
		monitorJSON = string(monitorJSONBytes)
		if err != nil {
			return response, fmt.Errorf("Body Error : %s", monitorJSON)
		}
	}

	path := "/_opendistro/_alerting/monitors/"

	var body json.RawMessage
	esClient, err := getClient(m.(*ProviderConf))
	if err != nil {
		return nil, err
	}
	switch client := esClient.(type) {
	case *elastic7.Client:
		var res *elastic7.Response
		res, err = client.PerformRequest(context.TODO(), elastic7.PerformRequestOptions{
			Method: "POST",
			Path:   path,
			Body:   monitorJSON,
		})
		body = res.Body
	case *elastic6.Client:
		var res *elastic6.Response
		res, err = client.PerformRequest(context.TODO(), elastic6.PerformRequestOptions{
			Method: "POST",
			Path:   path,
			Body:   monitorJSON,
		})
		body = res.Body
	default:
		err = errors.New("monitor resource not implemented prior to Elastic v6")
	}

	if err != nil {
		return response, err
	}

	if err := json.Unmarshal(body, response); err != nil {
		return response, fmt.Errorf("error unmarshalling monitor body: %+v: %+v", err, body)
	}
	normalizeMonitor(response.Monitor)
	return response, nil
}

func resourceElasticsearchOpenDistroPutMonitor(d *schema.ResourceData, m interface{}) (*monitorResponse, error) {
	monitorJSON := d.Get("body").(string)

	var err error
	response := new(monitorResponse)

	path, err := uritemplates.Expand("/_opendistro/_alerting/monitors/{id}", map[string]string{
		"id": d.Id(),
	})
	if err != nil {
		return response, fmt.Errorf("error building URL path for monitor: %+v", err)
	}

	var body json.RawMessage
	esClient, err := getClient(m.(*ProviderConf))
	if err != nil {
		return nil, err
	}
	switch client := esClient.(type) {
	case *elastic7.Client:
		var res *elastic7.Response
		res, err = client.PerformRequest(context.TODO(), elastic7.PerformRequestOptions{
			Method: "PUT",
			Path:   path,
			Body:   monitorJSON,
		})
		body = res.Body
	case *elastic6.Client:
		var res *elastic6.Response
		res, err = client.PerformRequest(context.TODO(), elastic6.PerformRequestOptions{
			Method: "PUT",
			Path:   path,
			Body:   monitorJSON,
		})
		body = res.Body
	default:
		err = errors.New("monitor resource not implemented prior to Elastic v6")
	}

	if err != nil {
		return response, err
	}

	if err := json.Unmarshal(body, response); err != nil {
		return response, fmt.Errorf("error unmarshalling monitor body: %+v: %+v", err, body)
	}

	return response, nil
}

type monitorAction struct {
	Name            string `json:"name"`
	DestinationId   string `json:"destination_id"`
	MessageTemplate string `json:"message_template"`
	MessageLanguage string `json:"message_lang"`
	SubjectTemplate string `json:"subject_template"`
	SubjectLanguage string `json:"subject_lang"`
	ThrottleEnabled bool   `json:"throttle_enabled"`
}

type monitorTrigger struct {
	Name              string          `json:"name"`
	Severity          string          `json:"severity"`
	ConditionScript   string          `json:"condition_script"`
	ConditionLanguage string          `json:"condition_lang"`
	Actions           []monitorAction `json:"actions"`
}

type monitorInput struct {
	SearchIndices  []string `json:"indices"`
	SearchQuery    string   `json:"query"`
	TerminateAfter int      `json:"terminate_after"`
}

type monitorPeriod struct {
	Interval int    `json:"interval"`
	Unit     string `json:"unit"`
}

type monitorSchedule struct {
	Period monitorPeriod `json:"period"`
}

type monitor struct {
	Name     string           `json:"name"`
	Enabled  bool             `json:"enabled"`
	Schedule monitorSchedule  `json:"schedule"`
	Inputs   []monitorInput   `json:"inputs"`
	Triggers []monitorTrigger `json:"triggers"`
}

type monitorResponse struct {
	Version int                    `json:"_version"`
	ID      string                 `json:"_id"`
	Monitor map[string]interface{} `json:"monitor"`
}
