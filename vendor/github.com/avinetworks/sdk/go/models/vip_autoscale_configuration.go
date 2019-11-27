package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// VipAutoscaleConfiguration vip autoscale configuration
// swagger:model VipAutoscaleConfiguration
type VipAutoscaleConfiguration struct {

	// This is the list of AZ+Subnet in which Vips will be spawned. Field introduced in 17.2.12, 18.1.2.
	Zones []*VipAutoscaleZones `json:"zones,omitempty"`
}
