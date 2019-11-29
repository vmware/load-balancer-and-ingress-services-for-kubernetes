package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// AWSASGNotifDetails a w s a s g notif details
// swagger:model AWSASGNotifDetails
type AWSASGNotifDetails struct {

	// asg_name of AWSASGNotifDetails.
	AsgName *string `json:"asg_name,omitempty"`

	// cc_id of AWSASGNotifDetails.
	CcID *string `json:"cc_id,omitempty"`

	// error_string of AWSASGNotifDetails.
	ErrorString *string `json:"error_string,omitempty"`

	// event_type of AWSASGNotifDetails.
	EventType *string `json:"event_type,omitempty"`

	// instance_id of AWSASGNotifDetails.
	InstanceID *string `json:"instance_id,omitempty"`

	// Placeholder for description of property instance_ip_addr of obj type AWSASGNotifDetails field type str  type object
	InstanceIPAddr *IPAddr `json:"instance_ip_addr,omitempty"`

	// UUID of the Pool. It is a reference to an object of type Pool. Field introduced in 17.2.3.
	PoolRef *string `json:"pool_ref,omitempty"`

	// vpc_id of AWSASGNotifDetails.
	VpcID *string `json:"vpc_id,omitempty"`
}
