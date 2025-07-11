// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// FalsePositiveResult false positive result
// swagger:model FalsePositiveResult
type FalsePositiveResult struct {

	// This flag indicates whether this result is identifying an attack. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Attack *bool `json:"attack,omitempty"`

	// Confidence on false positive detection. Allowed values are 0-100. Field introduced in 21.1.1. Unit is PERCENT. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Confidence *float32 `json:"confidence,omitempty"`

	// This flag indicates whether this result is identifying a false positive. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	FalsePositive *bool `json:"false_positive,omitempty"`

	// Meta data for this false positive result. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	FpResultHeader *FalsePositiveResultHeader `json:"fp_result_header"`

	// HTTP method for URIs did false positive detection. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	HTTPMethod *string `json:"http_method,omitempty"`

	// HTTP request header info if URI hit signature rule and match element is REQUEST_HEADERS. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	HTTPRequestHeaderInfo *HeaderInfoInURI `json:"http_request_header_info,omitempty"`

	// Params info if URI hit signature rule and match element is ARGS. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ParamsInfo *ParamsInURI `json:"params_info,omitempty"`

	// Signature rule info hitted by URI. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	RuleInfo *RuleInfo `json:"rule_info,omitempty"`

	// URIs did false positive detection. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	URI *string `json:"uri,omitempty"`

	// What failing mode that false positive detected as for current URI. Enum options - ALWAYS_FAIL, SOMETIMES_FAIL, NOT_SURE. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	URIResultMode *string `json:"uri_result_mode"`
}
