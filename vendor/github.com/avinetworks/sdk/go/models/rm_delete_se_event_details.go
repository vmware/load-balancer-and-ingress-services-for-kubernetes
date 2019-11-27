package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// RmDeleteSeEventDetails rm delete se event details
// swagger:model RmDeleteSeEventDetails
type RmDeleteSeEventDetails struct {

	// cloud_name of RmDeleteSeEventDetails.
	CloudName *string `json:"cloud_name,omitempty"`

	// Unique object identifier of cloud.
	CloudUUID *string `json:"cloud_uuid,omitempty"`

	// host_name of RmDeleteSeEventDetails.
	HostName *string `json:"host_name,omitempty"`

	// Unique object identifier of host.
	HostUUID *string `json:"host_uuid,omitempty"`

	// reason of RmDeleteSeEventDetails.
	Reason *string `json:"reason,omitempty"`

	// se_cookie of RmDeleteSeEventDetails.
	SeCookie *string `json:"se_cookie,omitempty"`

	// se_grp_name of RmDeleteSeEventDetails.
	SeGrpName *string `json:"se_grp_name,omitempty"`

	// Unique object identifier of se_grp.
	SeGrpUUID *string `json:"se_grp_uuid,omitempty"`

	// se_name of RmDeleteSeEventDetails.
	SeName *string `json:"se_name,omitempty"`

	// Unique object identifier of se.
	SeUUID *string `json:"se_uuid,omitempty"`

	// Number of status_code.
	StatusCode *int64 `json:"status_code,omitempty"`
}
