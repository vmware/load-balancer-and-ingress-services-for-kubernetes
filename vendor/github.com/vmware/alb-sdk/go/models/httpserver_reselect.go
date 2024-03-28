// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// HttpserverReselect httpserver reselect
// swagger:model HTTPServerReselect
type HttpserverReselect struct {

	// Enable HTTP request reselect when server responds with specific response codes. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- false), Basic edition(Allowed values- false), Enterprise with Cloud Services edition.
	// Required: true
	Enabled *bool `json:"enabled"`

	// Number of times to retry an HTTP request when server responds with configured status codes. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumRetries *uint32 `json:"num_retries,omitempty"`

	// Allow retry of non-idempotent HTTP requests. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	RetryNonidempotent *bool `json:"retry_nonidempotent,omitempty"`

	// Timeout per retry attempt, for a given request. Value of 0 indicates default timeout. Allowed values are 0-3600000. Field introduced in 18.1.5,18.2.1. Unit is MILLISECONDS. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	RetryTimeout *uint32 `json:"retry_timeout,omitempty"`

	// Server response codes which will trigger an HTTP request retry. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SvrRespCode *HTTPReselectRespCode `json:"svr_resp_code,omitempty"`
}
