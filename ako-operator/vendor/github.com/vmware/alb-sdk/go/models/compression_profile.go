// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// CompressionProfile compression profile
// swagger:model CompressionProfile
type CompressionProfile struct {

	// Number of buffers to use for compression output. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	BufNum *int32 `json:"buf_num,omitempty"`

	// Size of each buffer used for compression output, this should ideally be a multiple of pagesize. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	BufSize *int32 `json:"buf_size,omitempty"`

	// Compress only content types listed in this *string group. Content types not present in this list are not compressed. It is a reference to an object of type StringGroup. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CompressibleContentRef *string `json:"compressible_content_ref,omitempty"`

	// Compress HTTP response content if it wasn't already compressed. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Compression *bool `json:"compression"`

	// Custom filters used when auto compression is not selected. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Filter []*CompressionFilter `json:"filter,omitempty"`

	// hash size used by compression, rounded to the last power of 2. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	HashSize *int32 `json:"hash_size,omitempty"`

	// Level of compression to apply on content selected for aggressive compression. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	LevelAggressive *int32 `json:"level_aggressive,omitempty"`

	// Level of compression to apply on content selected for normal compression. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	LevelNormal *int32 `json:"level_normal,omitempty"`

	// If client RTT is higher than this threshold, enable normal compression on the response. Field introduced in 21.1.1. Unit is MILLISECONDS. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	MaxLowRtt *int32 `json:"max_low_rtt,omitempty"`

	// If client RTT is higher than this threshold, enable aggressive compression on the response.  . Field introduced in 21.1.1. Unit is MILLISECONDS. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	MinHighRtt *int32 `json:"min_high_rtt,omitempty"`

	// Minimum response content length to enable compression. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	MinLength *int32 `json:"min_length,omitempty"`

	// Values that identify mobile browsers in order to enable aggressive compression. It is a reference to an object of type StringGroup. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	MobileStrRef *string `json:"mobile_str_ref,omitempty"`

	// Offload compression from the servers to AVI. Saves compute cycles on the servers. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	RemoveAcceptEncodingHeader *bool `json:"remove_accept_encoding_header"`

	// Compress content automatically or add custom filters to define compressible content and compression levels. Enum options - AUTO_COMPRESSION, CUSTOM_COMPRESSION. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Type *string `json:"type"`

	// window size used by compression, rounded to the last power of 2. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	WindowSize *int32 `json:"window_size,omitempty"`
}
