package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// RmSpawnSeEventDetails rm spawn se event details
// swagger:model RmSpawnSeEventDetails
type RmSpawnSeEventDetails struct {

	// availability_zone of RmSpawnSeEventDetails.
	AvailabilityZone *string `json:"availability_zone,omitempty"`

	// cloud_name of RmSpawnSeEventDetails.
	CloudName *string `json:"cloud_name,omitempty"`

	// Unique object identifier of cloud.
	CloudUUID *string `json:"cloud_uuid,omitempty"`

	// host_name of RmSpawnSeEventDetails.
	HostName *string `json:"host_name,omitempty"`

	// Unique object identifier of host.
	HostUUID *string `json:"host_uuid,omitempty"`

	// Number of memory.
	Memory *int32 `json:"memory,omitempty"`

	// network_names of RmSpawnSeEventDetails.
	NetworkNames []string `json:"network_names,omitempty"`

	// networks of RmSpawnSeEventDetails.
	Networks []string `json:"networks,omitempty"`

	// reason of RmSpawnSeEventDetails.
	Reason *string `json:"reason,omitempty"`

	// se_cookie of RmSpawnSeEventDetails.
	SeCookie *string `json:"se_cookie,omitempty"`

	// se_grp_name of RmSpawnSeEventDetails.
	SeGrpName *string `json:"se_grp_name,omitempty"`

	// Unique object identifier of se_grp.
	SeGrpUUID *string `json:"se_grp_uuid,omitempty"`

	// se_name of RmSpawnSeEventDetails.
	SeName *string `json:"se_name,omitempty"`

	// Unique object identifier of se.
	SeUUID *string `json:"se_uuid,omitempty"`

	// Number of status_code.
	StatusCode *int64 `json:"status_code,omitempty"`

	// Number of vcpus.
	Vcpus *int32 `json:"vcpus,omitempty"`

	// vs_name of RmSpawnSeEventDetails.
	VsName *string `json:"vs_name,omitempty"`

	// Unique object identifier of vs.
	VsUUID *string `json:"vs_uuid,omitempty"`
}
