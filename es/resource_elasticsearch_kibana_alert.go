package es

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/olivere/elastic/uritemplates"

	elastic7 "github.com/olivere/elastic/v7"

	"github.com/phillbaker/terraform-provider-elasticsearch/kibana"
)

var minimalKibanaVersion, _ = version.NewVersion("7.7.0")
var notifyWhenKibanaVersion, _ = version.NewVersion("7.11.0")

func resourceElasticsearchKibanaAlert() *schema.Resource {
	return &schema.Resource{
		Create: resourceElasticsearchKibanaAlertCreate,
		Read:   resourceElasticsearchKibanaAlertRead,
		Update: resourceElasticsearchKibanaAlertUpdate,
		Delete: resourceElasticsearchKibanaAlertDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				ForceNew:    true,
				Required:    true,
				Description: "The name of the alert, does not have to be unique, used to identify and find an alert.",
			},
			"tags": {
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "A list of tag names, they appear in the alert listing in the UI which is searchable by tag.",
			},
			"alert_type_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     ".index-threshold",
				Description: "The ID of the alert type that you want to call when the alert is scheduled to run, defaults to `.index-threshold`.",
			},
			"schedule": {
				Type:        schema.TypeList,
				MaxItems:    1,
				MinItems:    1,
				Optional:    true,
				Description: "How frequently the alert conditions are checked. Note that the timing of evaluating alerts is not guaranteed, particularly for intervals of less than 10 seconds",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"interval": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Specifies the interval in seconds, minutes, hours or days at which the alert should execute, e.g. 10s, 5m, 1h.",
						},
					},
				},
			},
			"throttle": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "How often this alert should fire the same action, this reduces repeated notifications.",
			},
			"notify_when": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The condition for throttling the notification: `onActionGroupChange`, `onActiveAlert`, or `onThrottleInterval`. Only available in Kibana >= 7.11",
			},
			"enabled": {
				Type:        schema.TypeBool,
				Default:     true,
				Optional:    true,
				Description: "Whether the alert is scheduled for evaluation.",
			},
			"consumer": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "alerts",
				Description: "The name of the application that owns the alert. This name has to match the Kibana Feature name, as that dictates the required RBAC privileges. Defaults to `alerts`.",
			},
			"conditions": {
				Type:        schema.TypeSet,
				Optional:    true,
				MaxItems:    1,
				MinItems:    1,
				Description: "The conditions under which the alert is active, they create an expression to be evaluated by the alert type executor. These parameters are passed to the executor `params`. There may be specific attributes for different alert types. Either `params_json` or `conditions` must be specified.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"threshold_comparator": {
							Type:     schema.TypeString,
							Required: true,
						},
						"time_window_size": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"time_window_unit": {
							Type:     schema.TypeString,
							Required: true,
						},
						"term_size": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"time_field": {
							Type:     schema.TypeString,
							Required: true,
						},
						"group_by": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"aggregation_field": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"aggregation_type": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"term_field": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"index": {
							Type:        schema.TypeSet,
							Required:    true,
							MinItems:    1,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "",
						},
						"threshold": {
							Type:        schema.TypeSet,
							Required:    true,
							MinItems:    1,
							Elem:        &schema.Schema{Type: schema.TypeInt},
							Description: "",
						},
					},
				},
			},
			"params_json": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          "",
				ValidateFunc:     validation.StringIsJSON,
				Description:      "JSON body of alert `params`. Either `params_json` or `conditions` must be specified.",
				DiffSuppressFunc: suppressEquivalentJson,
			},
			"actions": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "Actions are invocations of Kibana services or integrations with third-party systems, that run as background tasks on the Kibana server when alert conditions are met.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"group": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "default",
							Description: "When to execute the action, e.g. `threshold met` or `recovered`.",
						},
						"id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The identifier of the saved action object, a UUID.",
						},
						"action_type_id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The type of the action, e.g. `.index`, `.webhook`, etc.",
						},
						"params": {
							Type:        schema.TypeMap,
							Optional:    true,
							Description: "Key value pairs passed to the action executor, e.g. a Mustache formatted `message`.",
						},
					},
				},
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "Alerts allow you to define rules to detect conditions and trigger actions when those conditions are met. Alerts work by running checks on a schedule to detect conditions. When a condition is met, the alert tracks it as an alert instance and responds by triggering one or more actions. Actions typically involve interaction with Kibana services or third party integrations. For more see the [docs](https://www.elastic.co/guide/en/kibana/current/alerting-getting-started.html).",
	}
}

func resourceElasticsearchKibanaAlertCreate(d *schema.ResourceData, meta interface{}) error {
	err := resourceElasticsearchKibanaAlertCheckVersion(meta)
	if err != nil {
		return err
	}

	id, err := resourceElasticsearchPostKibanaAlert(d, meta)
	if err != nil {
		return err
	}

	log.Printf("[INFO] Kibana Alert (%s) created", id)
	d.SetId(id)

	return nil
}

func resourceElasticsearchKibanaAlertRead(d *schema.ResourceData, meta interface{}) error {
	err := resourceElasticsearchKibanaAlertCheckVersion(meta)
	if err != nil {
		return err
	}

	id := d.Id()
	spaceID := ""

	var alert kibana.Alert

	providerConf := meta.(*ProviderConf)
	esClient, err := getKibanaClient(providerConf)
	if err != nil {
		return err
	}

	switch client := esClient.(type) {
	case *elastic7.Client:
		alert, err = kibanaGetAlert(client, id, spaceID)
	default:
		err = fmt.Errorf("Kibana Alert endpoint only available from Kibana >= 7.7, got version < 7.0.0")
	}

	if err != nil {
		if elastic7.IsNotFound(err) {
			log.Printf("[WARN] Kibana Alert (%s) not found, removing from state", id)
			d.SetId("")
			return nil
		}

		return err
	}

	schedule := make([]map[string]interface{}, 0, 1)
	schedule = append(schedule, map[string]interface{}{"interval": alert.Schedule.Interval})

	ds := &resourceDataSetter{d: d}
	ds.set("name", alert.Name)
	ds.set("tags", alert.Tags)
	ds.set("alert_type_id", alert.AlertTypeID)
	ds.set("schedule", schedule)
	ds.set("throttle", alert.Throttle)
	ds.set("notify_when", alert.NotifyWhen)
	ds.set("enabled", alert.Enabled)
	ds.set("consumer", alert.Consumer)
	if _, ok := d.GetOk("params_json"); ok {
		pj, err := json.Marshal(alert.Params)
		if err != nil {
			return err
		}
		ds.set("params_json", string(pj))
	} else {
		ds.set("conditions", flattenKibanaAlertConditions(alert.Params))
	}
	ds.set("actions", flattenKibanaAlertActions(alert.Actions))

	return ds.err
}

func resourceElasticsearchKibanaAlertUpdate(d *schema.ResourceData, meta interface{}) error {
	err := resourceElasticsearchKibanaAlertCheckVersion(meta)
	if err != nil {
		return err
	}

	return resourceElasticsearchPutKibanaAlert(d, meta)
}

func resourceElasticsearchKibanaAlertDelete(d *schema.ResourceData, meta interface{}) error {
	err := resourceElasticsearchKibanaAlertCheckVersion(meta)
	if err != nil {
		return err
	}

	id := d.Id()
	spaceID := ""

	providerConf := meta.(*ProviderConf)
	kibanaClient, err := getKibanaClient(providerConf)
	if err != nil {
		return err
	}

	switch client := kibanaClient.(type) {
	case *elastic7.Client:
		err = kibanaDeleteAlert(client, id, spaceID)
	default:
		err = fmt.Errorf("Kibana Alert endpoint only available from ElasticSearch >= 7.7, got version < 7.0.0")
	}

	if err != nil {
		return err
	}
	d.SetId("")
	return nil
}

func resourceElasticsearchPostKibanaAlert(d *schema.ResourceData, meta interface{}) (string, error) {
	spaceID := ""

	providerConf := meta.(*ProviderConf)
	kibanaClient, err := getKibanaClient(providerConf)
	if err != nil {
		return "", err
	}

	alertSchedule := kibana.AlertSchedule{}
	schedule := d.Get("schedule").([]interface{})
	if len(schedule) > 0 {
		scheduleEntry := schedule[0].(map[string]interface{})
		alertSchedule.Interval = scheduleEntry["interval"].(string)
	}
	actions, err := expandKibanaActionsList(d.Get("actions").(*schema.Set).List())
	if err != nil {
		return "", err
	}

	tags := expandStringList(d.Get("tags").(*schema.Set).List())

	var params map[string]interface{}
	if conditions, ok := d.GetOk("conditions"); ok {
		c := conditions.(*schema.Set).List()[0].(map[string]interface{})
		params = expandKibanaAlertConditions(c)
	} else if pj, ok := d.GetOk("params_json"); ok {
		bytes := []byte(pj.(string))
		err = json.Unmarshal(bytes, &params)
		if err != nil {
			return "", fmt.Errorf("fail to unmarshal: %v", err)
		}
	}

	alert := kibana.Alert{
		Name:        d.Get("name").(string),
		Tags:        tags,
		AlertTypeID: d.Get("alert_type_id").(string),
		Schedule:    alertSchedule,
		Throttle:    d.Get("throttle").(string),
		Enabled:     d.Get("enabled").(bool),
		Consumer:    d.Get("consumer").(string),
		Params:      params,
		Actions:     actions,
	}

	version, _ := resourceElasticsearchKibanaGetVersion(meta)
	if version.GreaterThanOrEqual(notifyWhenKibanaVersion) {
		alert.NotifyWhen = d.Get("notify_when").(string)
	}

	var id string
	switch client := kibanaClient.(type) {
	case *elastic7.Client:
		id, err = kibanaPostAlert(client, spaceID, alert)
	default:
		err = fmt.Errorf("Kibana Alert endpoint only available from ElasticSearch >= 7.7, got version < 7.0.0")
	}

	return id, err
}

func expandKibanaActionsList(resourcesArray []interface{}) ([]kibana.AlertAction, error) {
	actions := make([]kibana.AlertAction, 0, len(resourcesArray))
	for _, resource := range resourcesArray {
		data, ok := resource.(map[string]interface{})
		if !ok {
			return actions, fmt.Errorf("Error asserting data: %+v, %T", resource, resource)
		}
		action := kibana.AlertAction{
			ID:           data["id"].(string),
			Group:        data["group"].(string),
			ActionTypeId: data["action_type_id"].(string),
			Params:       data["params"].(map[string]interface{}),
		}
		actions = append(actions, action)
	}

	return actions, nil
}

func flattenKibanaAlertActions(actions []kibana.AlertAction) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(actions))

	for _, a := range actions {
		m := make(map[string]interface{})
		m["id"] = a.ID
		m["group"] = a.Group
		m["action_type_id"] = a.ActionTypeId
		m["params"] = flattenMap(a.Params)
		result = append(result, m)
	}
	return result
}

func expandKibanaAlertConditions(raw map[string]interface{}) map[string]interface{} {
	conditions := make(map[string]interface{})

	// convert cases
	for k := range raw {
		camelCasedKey := toCamelCase(k, false)
		conditions[camelCasedKey] = raw[k]
		if camelCasedKey != k {
			delete(conditions, k)
		}
	}

	// override nested objects
	conditions["index"] = raw["index"].(*schema.Set).List()
	conditions["threshold"] = raw["threshold"].(*schema.Set).List()

	// convert abbreviated fields
	conditions["aggField"] = conditions["aggregationField"]
	delete(conditions, "aggregationField")
	conditions["aggType"] = conditions["aggregationType"]
	delete(conditions, "aggregationType")

	return conditions
}

func flattenKibanaAlertConditions(raw map[string]interface{}) []map[string]interface{} {
	conditions := make(map[string]interface{})

	// convert cases
	for k := range raw {
		underscoredKey := toUnderscore(k)
		conditions[underscoredKey] = raw[k]
		if underscoredKey != k {
			delete(conditions, k)
		}
	}
	log.Printf("[INFO] flattenKibanaAlertConditions: %+v", conditions)
	// override nested objects
	conditions["index"] = flattenStringAsInterfaceSet(conditions["index"].([]interface{}))
	conditions["threshold"] = flattenFloatSet(conditions["threshold"].([]interface{}))

	// convert abbreviated fields
	conditions["aggregation_field"] = conditions["agg_field"]
	delete(conditions, "agg_field")
	conditions["aggregation_type"] = conditions["agg_type"]
	delete(conditions, "agg_type")

	return []map[string]interface{}{conditions}
}

func resourceElasticsearchPutKibanaAlert(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceElasticsearchKibanaGetVersion(meta interface{}) (*version.Version, error) {
	providerConf := meta.(*ProviderConf)
	esClient, err := getClient(providerConf)
	if err != nil {
		return nil, err
	}

	switch esClient.(type) {
	case *elastic7.Client:
		return version.NewVersion(providerConf.esVersion)
	default:
		return nil, fmt.Errorf("Kibana Alert endpoint only available from ElasticSearch >= 7.7, got version < 7.0.0")
	}
}

func resourceElasticsearchKibanaAlertCheckVersion(meta interface{}) error {
	elasticVersion, err := resourceElasticsearchKibanaGetVersion(meta)
	if err != nil {
		return err
	}

	if elasticVersion.LessThan(minimalKibanaVersion) {
		return fmt.Errorf("Kibana Alert endpoint only available from ElasticSearch >= 7.7, got version %s", elasticVersion.String())
	}

	return err
}

func kibanaGetAlert(client *elastic7.Client, id, spaceID string) (kibana.Alert, error) {
	path, err := uritemplates.Expand("/api/alerts/alert/{id}", map[string]string{
		"id": id,
	})
	if err != nil {
		return kibana.Alert{}, fmt.Errorf("error building URL path for alert: %+v", err)
	}

	var body json.RawMessage
	var res *elastic7.Response
	res, err = client.PerformRequest(context.TODO(), elastic7.PerformRequestOptions{
		Method: "GET",
		Path:   path,
	})
	if err != nil {
		return kibana.Alert{}, err
	}

	body = res.Body

	alert := new(kibana.Alert)
	if err := json.Unmarshal(body, alert); err != nil {
		return *alert, fmt.Errorf("error unmarshalling alert body: %+v: %+v", err, body)
	}

	return *alert, nil
}

func kibanaPostAlert(client *elastic7.Client, spaceID string, alert kibana.Alert) (string, error) {
	path, err := uritemplates.Expand("/api/alerts/alert", map[string]string{})
	if err != nil {
		return "", fmt.Errorf("error building URL path for alert: %+v", err)
	}

	body, err := json.Marshal(alert)
	if err != nil {
		log.Printf("[INFO] kibanaPostAlert: %+v %+v %+v", path, alert, err)
		return "", fmt.Errorf("Body Error: %s", err)
	}

	var res *elastic7.Response
	res, err = client.PerformRequest(context.TODO(), elastic7.PerformRequestOptions{
		Method: "POST",
		Path:   path,
		Body:   string(body[:]),
	})

	if err != nil {
		log.Printf("[INFO] kibanaPostAlert: %+v %+v %+v", path, alert, string(body[:]))
		return "", err
	}

	if err := json.Unmarshal(res.Body, &alert); err != nil {
		return "", fmt.Errorf("error unmarshalling alert body: %+v: %+v", err, body)
	}

	return alert.ID, nil
}

func kibanaDeleteAlert(client *elastic7.Client, id, spaceID string) error {
	path, err := uritemplates.Expand("/api/alerts/alert/{id}", map[string]string{
		"id": id,
	})
	if err != nil {
		return fmt.Errorf("error building URL path for alert: %+v", err)
	}

	_, err = client.PerformRequest(context.TODO(), elastic7.PerformRequestOptions{
		Method: "DELETE",
		Path:   path,
	})

	if err != nil {
		return err
	}

	return nil
}
