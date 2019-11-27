package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// VipAutoscaleGroup vip autoscale group
// swagger:model VipAutoscaleGroup
type VipAutoscaleGroup struct {

	//  Field introduced in 17.2.12, 18.1.2.
	Configuration *VipAutoscaleConfiguration `json:"configuration,omitempty"`

	//  Field introduced in 17.2.12, 18.1.2.
	Policy *VipAutoscalePolicy `json:"policy,omitempty"`
}
