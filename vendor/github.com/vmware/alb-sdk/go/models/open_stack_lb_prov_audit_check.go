package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// OpenStackLbProvAuditCheck open stack lb prov audit check
// swagger:model OpenStackLbProvAuditCheck
type OpenStackLbProvAuditCheck struct {

	// cc_id of OpenStackLbProvAuditCheck.
	CcID *string `json:"cc_id,omitempty"`

	// detail of OpenStackLbProvAuditCheck.
	Detail *string `json:"detail,omitempty"`

	// Number of elapsed.
	Elapsed *int32 `json:"elapsed,omitempty"`

	// id of OpenStackLbProvAuditCheck.
	// Required: true
	ID *string `json:"id"`

	// result of OpenStackLbProvAuditCheck.
	Result *string `json:"result,omitempty"`

	// tenant of OpenStackLbProvAuditCheck.
	// Required: true
	Tenant *string `json:"tenant"`

	// user of OpenStackLbProvAuditCheck.
	// Required: true
	User *string `json:"user"`
}
