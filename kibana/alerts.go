package kibana

type AlertSchedule struct {
	Interval string `json:"interval,omitempty"`
}

type AlertAction struct {
	ID           string                 `json:"id"`
	Group        string                 `json:"group"`
	ActionTypeId string                 `json:"actionTypeId,omitempty"`
	Params       map[string]interface{} `json:"params,omitempty"`
}

type Alert struct {
	ID          string                 `json:"id,omitempty"`
	Name        string                 `json:"name"`
	Tags        []string               `json:"tags,omitempty"`
	AlertTypeID string                 `json:"alertTypeId,omitempty"`
	Schedule    AlertSchedule          `json:"schedule,omitempty"`
	Throttle    string                 `json:"throttle,omitempty"`
	NotifyWhen  string                 `json:"notifyWhen,omitempty"`
	Enabled     bool                   `json:"enabled,omitempty"`
	Consumer    string                 `json:"consumer,omitempty"`
	Params      map[string]interface{} `json:"params,omitempty"`
	Actions     []AlertAction          `json:"actions,omitempty"`
}
