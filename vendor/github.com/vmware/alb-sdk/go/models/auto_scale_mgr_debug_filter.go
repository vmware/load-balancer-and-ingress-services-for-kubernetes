package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// AutoScaleMgrDebugFilter auto scale mgr debug filter
// swagger:model AutoScaleMgrDebugFilter
type AutoScaleMgrDebugFilter struct {

	// Enable aws autoscale integration. This is an alpha feature. Field introduced in 17.1.1.
	EnableAwsAutoscaleIntegration *bool `json:"enable_aws_autoscale_integration,omitempty"`

	// period of running intelligent autoscale check.
	IntelligentAutoscalePeriod *int32 `json:"intelligent_autoscale_period,omitempty"`

	// uuid of the Pool. It is a reference to an object of type Pool.
	PoolRef *string `json:"pool_ref,omitempty"`
}
