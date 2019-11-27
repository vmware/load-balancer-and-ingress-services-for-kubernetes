package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// PoolDeploymentUpdateInfo pool deployment update info
// swagger:model PoolDeploymentUpdateInfo
type PoolDeploymentUpdateInfo struct {

	// Pool deployment state used with the PG deployment policy. Enum options - EVALUATION_IN_PROGRESS, IN_SERVICE, OUT_OF_SERVICE, EVALUATION_FAILED.
	DeploymentState *string `json:"deployment_state,omitempty"`

	// Evaluation period for deployment update.
	EvaluationDuration *int32 `json:"evaluation_duration,omitempty"`

	// Operational traffic ratio for the pool.
	Ratio *int32 `json:"ratio,omitempty"`

	// List of results for each deployment rule.
	Results []*PGDeploymentRuleResult `json:"results,omitempty"`

	// Member Pool's ID.
	UUID *string `json:"uuid,omitempty"`

	// Reason returned in webhook callback when configured.
	WebhookReason *string `json:"webhook_reason,omitempty"`

	// Result of webhook callback when configured.
	WebhookResult *bool `json:"webhook_result,omitempty"`
}
