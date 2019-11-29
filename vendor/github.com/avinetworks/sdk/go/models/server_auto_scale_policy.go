package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ServerAutoScalePolicy server auto scale policy
// swagger:model ServerAutoScalePolicy
type ServerAutoScalePolicy struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// User defined description for the object.
	Description *string `json:"description,omitempty"`

	// Use Avi intelligent autoscale algorithm where autoscale is performed by comparing load on the pool against estimated capacity of all the servers.
	IntelligentAutoscale *bool `json:"intelligent_autoscale,omitempty"`

	// Maximum extra capacity as percentage of load used by the intelligent scheme. Scalein is triggered when available capacity is more than this margin. Allowed values are 1-99.
	IntelligentScaleinMargin *int32 `json:"intelligent_scalein_margin,omitempty"`

	// Minimum extra capacity as percentage of load used by the intelligent scheme. Scaleout is triggered when available capacity is less than this margin. Allowed values are 1-99.
	IntelligentScaleoutMargin *int32 `json:"intelligent_scaleout_margin,omitempty"`

	// Maximum number of servers to scalein simultaneously. The actual number of servers to scalein is chosen such that target number of servers is always more than or equal to the min_size.
	MaxScaleinAdjustmentStep *int32 `json:"max_scalein_adjustment_step,omitempty"`

	// Maximum number of servers to scaleout simultaneously. The actual number of servers to scaleout is chosen such that target number of servers is always less than or equal to the max_size.
	MaxScaleoutAdjustmentStep *int32 `json:"max_scaleout_adjustment_step,omitempty"`

	// Maximum number of servers after scaleout. Allowed values are 0-400.
	MaxSize *int32 `json:"max_size,omitempty"`

	// No scale-in happens once number of operationally up servers reach min_servers. Allowed values are 0-400.
	MinSize *int32 `json:"min_size,omitempty"`

	// Name of the object.
	// Required: true
	Name *string `json:"name"`

	// Trigger scalein when alerts due to any of these Alert configurations are raised. It is a reference to an object of type AlertConfig.
	ScaleinAlertconfigRefs []string `json:"scalein_alertconfig_refs,omitempty"`

	// Cooldown period during which no new scalein is triggered to allow previous scalein to successfully complete.
	ScaleinCooldown *int32 `json:"scalein_cooldown,omitempty"`

	// Trigger scaleout when alerts due to any of these Alert configurations are raised. It is a reference to an object of type AlertConfig.
	ScaleoutAlertconfigRefs []string `json:"scaleout_alertconfig_refs,omitempty"`

	// Cooldown period during which no new scaleout is triggered to allow previous scaleout to successfully complete.
	ScaleoutCooldown *int32 `json:"scaleout_cooldown,omitempty"`

	//  It is a reference to an object of type Tenant.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// Use predicted load rather than current load.
	UsePredictedLoad *bool `json:"use_predicted_load,omitempty"`

	// Unique object identifier of the object.
	UUID *string `json:"uuid,omitempty"`
}
