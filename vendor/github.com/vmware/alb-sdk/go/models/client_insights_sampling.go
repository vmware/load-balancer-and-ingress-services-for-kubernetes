// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ClientInsightsSampling client insights sampling
// swagger:model ClientInsightsSampling
type ClientInsightsSampling struct {

	// Client IP addresses to check when inserting RUM script. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ClientIP *IPAddrMatch `json:"client_ip,omitempty"`

	// URL patterns to check when inserting RUM script. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SampleUris *StringMatch `json:"sample_uris,omitempty"`

	// URL patterns to avoid when inserting RUM script. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SkipUris *StringMatch `json:"skip_uris,omitempty"`
}
