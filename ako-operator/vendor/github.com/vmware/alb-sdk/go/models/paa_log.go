// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// PaaLog paa log
// swagger:model PaaLog
type PaaLog struct {

	// PingAccess Agent cache was used for authentication. Field introduced in 18.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CacheHit *bool `json:"cache_hit,omitempty"`

	// The PingAccess server required the client request body for authentication. Field introduced in 18.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ClientRequestBodySent *bool `json:"client_request_body_sent,omitempty"`

	// Logs for each request sent to PA server to completeauthentication for the initial request. Field introduced in 18.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	RequestLogs []*PaaRequestLog `json:"request_logs,omitempty"`
}
