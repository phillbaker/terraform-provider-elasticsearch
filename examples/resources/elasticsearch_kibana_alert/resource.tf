resource "elasticsearch_kibana_alert" "test" {
  name = "terraform-alert"
  schedule {
  	interval = "1m"
  }
  conditions {
    aggregation_type = "avg"
    term_size = 6
    threshold_comparator = ">"
    time_window_size = 5
    time_window_unit = "m"
    group_by = "top"
    threshold = [1000]
    index = [".test-index"]
    time_field = "@timestamp"
    aggregation_field = "sheet.version"
    term_field = "name.keyword"
  }
  actions {
  	id = "c87f0dc6-c301-4988-aee9-95d391359a39"
  	action_type_id = ".index"
  	params = {
  		level = "info"
  		message = "alert '{{alertName}}' is active for group '{{context.group}}':\n\n- Value: {{context.value}}\n- Conditions Met: {{context.conditions}} over {{params.timeWindowSize}}{{params.timeWindowUnit}}\n- Timestamp: {{context.date}}"
  	}
  }
}
