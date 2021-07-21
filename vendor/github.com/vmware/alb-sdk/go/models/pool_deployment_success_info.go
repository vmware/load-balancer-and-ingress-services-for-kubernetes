// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// PoolDeploymentSuccessInfo pool deployment success info
// swagger:model PoolDeploymentSuccessInfo
type PoolDeploymentSuccessInfo struct {

	// Previous pool in service. Field introduced in 18.1.1.
	PrevInServicePoolName *string `json:"prev_in_service_pool_name,omitempty"`

	// Previous pool in service. It is a reference to an object of type Pool.
	PrevInServicePoolRef *string `json:"prev_in_service_pool_ref,omitempty"`

	// Operational traffic ratio for the pool.
	Ratio *int32 `json:"ratio,omitempty"`

	// List of results for each deployment rule.
	Results []*PGDeploymentRuleResult `json:"results,omitempty"`

	// Pool's ID.
	UUID *string `json:"uuid,omitempty"`

	// Reason returned in webhook callback when configured.
	WebhookReason *string `json:"webhook_reason,omitempty"`

	// Result of webhook callback when configured.
	WebhookResult *bool `json:"webhook_result,omitempty"`
}
