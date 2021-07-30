// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ClientInsightsSampling client insights sampling
// swagger:model ClientInsightsSampling
type ClientInsightsSampling struct {

	// Client IP addresses to check when inserting RUM script.
	ClientIP *IPAddrMatch `json:"client_ip,omitempty"`

	// URL patterns to check when inserting RUM script.
	SampleUris *StringMatch `json:"sample_uris,omitempty"`

	// URL patterns to avoid when inserting RUM script.
	SkipUris *StringMatch `json:"skip_uris,omitempty"`
}
