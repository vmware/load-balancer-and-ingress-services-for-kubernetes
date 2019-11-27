package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// PoolDeploymentFailureInfo pool deployment failure info
// swagger:model PoolDeploymentFailureInfo
type PoolDeploymentFailureInfo struct {

	// Curent in-service pool. It is a reference to an object of type Pool.
	CurrInServicePoolName *string `json:"curr_in_service_pool_name,omitempty"`

	// Curent in service pool. It is a reference to an object of type Pool.
	CurrInServicePoolRef *string `json:"curr_in_service_pool_ref,omitempty"`

	// Operational traffic ratio for the pool.
	Ratio *int32 `json:"ratio,omitempty"`

	// Placeholder for description of property results of obj type PoolDeploymentFailureInfo field type str  type object
	Results []*PGDeploymentRuleResult `json:"results,omitempty"`

	// Pool's ID.
	UUID *string `json:"uuid,omitempty"`

	// Reason returned in webhook callback when configured.
	WebhookReason *string `json:"webhook_reason,omitempty"`

	// Result of webhook callback when configured.
	WebhookResult *bool `json:"webhook_result,omitempty"`
}
