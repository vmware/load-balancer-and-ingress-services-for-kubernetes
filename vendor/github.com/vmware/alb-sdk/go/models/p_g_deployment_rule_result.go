package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// PGDeploymentRuleResult p g deployment rule result
// swagger:model PGDeploymentRuleResult
type PGDeploymentRuleResult struct {

	// Metric value that is used as the pass fail. If it is not provided then it will simply compare it with current pool vs new pool.
	MetricValue *float64 `json:"metric_value,omitempty"`

	// Whether rule passed or failed.
	Result *bool `json:"result,omitempty"`

	// Rule used for evaluation.
	// Required: true
	Rule *PGDeploymentRule `json:"rule"`
}
