package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// LogControllerMapping log controller mapping
// swagger:model LogControllerMapping
type LogControllerMapping struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// controller_ip of LogControllerMapping.
	ControllerIP *string `json:"controller_ip,omitempty"`

	//  Enum options - METRICS_MGR_PORT_0, METRICS_MGR_PORT_1, METRICS_MGR_PORT_2, METRICS_MGR_PORT_3.
	MetricsMgrPort *string `json:"metrics_mgr_port,omitempty"`

	// Unique object identifier of node.
	NodeUUID *string `json:"node_uuid,omitempty"`

	// prev_controller_ip of LogControllerMapping.
	PrevControllerIP *string `json:"prev_controller_ip,omitempty"`

	//  Enum options - METRICS_MGR_PORT_0, METRICS_MGR_PORT_1, METRICS_MGR_PORT_2, METRICS_MGR_PORT_3.
	PrevMetricsMgrPort *string `json:"prev_metrics_mgr_port,omitempty"`

	// Placeholder for description of property static_mapping of obj type LogControllerMapping field type str  type boolean
	StaticMapping *bool `json:"static_mapping,omitempty"`

	//  It is a reference to an object of type Tenant.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// Unique object identifier of the object.
	UUID *string `json:"uuid,omitempty"`

	// Unique object identifier of vs.
	// Required: true
	VsUUID *string `json:"vs_uuid"`
}
