// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// AppLearningParams app learning params
// swagger:model AppLearningParams
type AppLearningParams struct {

	// Learn the params per URI path. Field introduced in 18.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	EnablePerURILearning *bool `json:"enable_per_uri_learning,omitempty"`

	// If true, learning will only be performed on requests from clients who have passed the authentication process configured in the Virtual Service's Auth Profile. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	LearnFromAuthenticatedClientsOnly *bool `json:"learn_from_authenticated_clients_only,omitempty"`

	// Maximum number of params programmed for an application. Allowed values are 10-1000. Field introduced in 18.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MaxParams *int32 `json:"max_params,omitempty"`

	// Maximum number of URI paths programmed for an application. Allowed values are 10-10000. Field introduced in 18.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MaxUris *int32 `json:"max_uris,omitempty"`

	// Minimum number of occurances required for a Param to qualify for learning. Field introduced in 18.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MinHitsToLearn *int64 `json:"min_hits_to_learn,omitempty"`

	// Percent of the requests subjected to Application learning. Allowed values are 1-100. Field introduced in 18.2.3. Unit is PERCENT. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SamplingPercent *int32 `json:"sampling_percent,omitempty"`

	// If configured, learning will only be performed on requests from client IPs within the configured IP Address Group. It is a reference to an object of type IpAddrGroup. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	TrustedIpgroupRef *string `json:"trusted_ipgroup_ref,omitempty"`

	// Frequency with which SE publishes Application learning data to controller. Allowed values are 1-60. Field introduced in 18.2.3. Unit is MIN. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UpdateInterval *int32 `json:"update_interval,omitempty"`
}
