package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// AlertRuleEvent alert rule event
// swagger:model AlertRuleEvent
type AlertRuleEvent struct {

	// Placeholder for description of property event_details of obj type AlertRuleEvent field type str  type object
	EventDetails []*EventDetailsFilter `json:"event_details,omitempty"`

	// When the selected event occurs, trigger this alert. Enum options - VINFRA_DISC_DC, VINFRA_DISC_HOST, VINFRA_DISC_CLUSTER, VINFRA_DISC_VM, VINFRA_DISC_NW, MGMT_NW_NAME_CHANGED, DISCOVERY_DATACENTER_DEL, VM_ADDED, VM_REMOVED, VINFRA_DISC_COMPLETE, VCENTER_ADDRESS_ERROR, SE_GROUP_CLUSTER_DEL, SE_GROUP_MGMT_NW_DEL, MGMT_NW_DEL, VCENTER_BAD_CREDENTIALS, ESX_HOST_UNREACHABLE, SERVER_DELETED, SE_GROUP_HOST_DEL, VINFRA_DISC_FAILURE, ESX_HOST_POWERED_DOWN...
	EventID *string `json:"event_id,omitempty"`

	// Placeholder for description of property not_cond of obj type AlertRuleEvent field type str  type boolean
	NotCond *bool `json:"not_cond,omitempty"`
}
