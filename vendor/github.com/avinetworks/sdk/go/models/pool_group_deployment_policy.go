package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// PoolGroupDeploymentPolicy pool group deployment policy
// swagger:model PoolGroupDeploymentPolicy
type PoolGroupDeploymentPolicy struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// It will automatically disable old production pools once there is a new production candidate.
	AutoDisableOldProdPools *bool `json:"auto_disable_old_prod_pools,omitempty"`

	// User defined description for the object.
	Description *string `json:"description,omitempty"`

	// Duration of evaluation period for automatic deployment. Allowed values are 60-86400.
	EvaluationDuration *int32 `json:"evaluation_duration,omitempty"`

	// The name of the pool group deployment policy.
	// Required: true
	Name *string `json:"name"`

	// Placeholder for description of property rules of obj type PoolGroupDeploymentPolicy field type str  type object
	Rules []*PGDeploymentRule `json:"rules,omitempty"`

	// deployment scheme. Enum options - BLUE_GREEN, CANARY.
	Scheme *string `json:"scheme,omitempty"`

	// Target traffic ratio before pool is made production. Allowed values are 1-100.
	TargetTestTrafficRatio *int32 `json:"target_test_traffic_ratio,omitempty"`

	//  It is a reference to an object of type Tenant.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// Ratio of the traffic that is sent to the pool under test. test ratio of 100 means blue green. Allowed values are 1-100.
	TestTrafficRatioRampup *int32 `json:"test_traffic_ratio_rampup,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// UUID of the pool group deployment policy.
	UUID *string `json:"uuid,omitempty"`

	// Webhook configured with URL that Avi controller will pass back information about pool group, old and new pool information and current deployment rule results. It is a reference to an object of type Webhook. Field introduced in 17.1.1.
	WebhookRef *string `json:"webhook_ref,omitempty"`
}
