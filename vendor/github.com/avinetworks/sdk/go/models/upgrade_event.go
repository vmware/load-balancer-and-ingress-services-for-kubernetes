package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// UpgradeEvent upgrade event
// swagger:model UpgradeEvent
type UpgradeEvent struct {

	// Time taken to complete upgrade event in seconds. Field introduced in 18.2.6.
	Duration *int32 `json:"duration,omitempty"`

	// Task end time. Field introduced in 18.2.6.
	EndTime *string `json:"end_time,omitempty"`

	// Ip of the node. Field introduced in 18.2.6.
	IP *IPAddr `json:"ip,omitempty"`

	// Upgrade event message if any. Field introduced in 18.2.6.
	Message *string `json:"message,omitempty"`

	// Task start time. Field introduced in 18.2.6.
	StartTime *string `json:"start_time,omitempty"`

	// Upgrade event status. Field introduced in 18.2.6.
	Status *bool `json:"status,omitempty"`

	// Sub tasks executed on each node. Field introduced in 18.2.8.
	SubTasks []string `json:"sub_tasks,omitempty"`
}
