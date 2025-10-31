// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// PoolDeploymentSuccessInfo pool deployment success info
// swagger:model PoolDeploymentSuccessInfo
type PoolDeploymentSuccessInfo struct {

	// Previous pool in service. Field introduced in 18.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PrevInServicePoolName *string `json:"prev_in_service_pool_name,omitempty"`

	// Previous pool in service. It is a reference to an object of type Pool. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PrevInServicePoolRef *string `json:"prev_in_service_pool_ref,omitempty"`

	// Operational traffic ratio for the pool. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Ratio *uint32 `json:"ratio,omitempty"`

	// List of results for each deployment rule. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Results []*PGDeploymentRuleResult `json:"results,omitempty"`

	// Pool's ID. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UUID *string `json:"uuid,omitempty"`

	// Reason returned in webhook callback when configured. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	WebhookReason *string `json:"webhook_reason,omitempty"`

	// Result of webhook callback when configured. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	WebhookResult *bool `json:"webhook_result,omitempty"`
}
