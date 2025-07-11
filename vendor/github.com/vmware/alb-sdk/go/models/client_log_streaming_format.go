// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ClientLogStreamingFormat client log streaming format
// swagger:model ClientLogStreamingFormat
type ClientLogStreamingFormat struct {

	// Format for the streamed logs. Enum options - LOG_STREAMING_FORMAT_JSON_FULL, LOG_STREAMING_FORMAT_JSON_SELECTED. Field introduced in 18.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Format *string `json:"format"`

	// List of log fields to be streamed, when selective fields (LOG_STREAMING_FORMAT_JSON_SELECTED) option is chosen. Only top-level fields in application or connection logs are supported. Field introduced in 18.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	IncludedFields []string `json:"included_fields,omitempty"`
}
