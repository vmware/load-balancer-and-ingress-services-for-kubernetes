package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// VIRetrievePGNames v i retrieve p g names
// swagger:model VIRetrievePGNames
type VIRetrievePGNames struct {

	// Unique object identifier of cloud.
	CloudUUID *string `json:"cloud_uuid,omitempty"`

	// datacenter of VIRetrievePGNames.
	Datacenter *string `json:"datacenter,omitempty"`

	// password of VIRetrievePGNames.
	Password *string `json:"password,omitempty"`

	// username of VIRetrievePGNames.
	Username *string `json:"username,omitempty"`

	// vcenter_url of VIRetrievePGNames.
	VcenterURL *string `json:"vcenter_url,omitempty"`
}
