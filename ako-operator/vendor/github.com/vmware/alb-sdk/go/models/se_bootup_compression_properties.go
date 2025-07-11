// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SeBootupCompressionProperties se bootup compression properties
// swagger:model SeBootupCompressionProperties
type SeBootupCompressionProperties struct {

	// Number of buffers to use for compression output. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	BufNum *int32 `json:"buf_num,omitempty"`

	// Size of each buffer used for compression output, this should ideally be a multiple of pagesize. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	BufSize *int32 `json:"buf_size,omitempty"`

	// hash size used by compression, rounded to the last power of 2. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	HashSize *int32 `json:"hash_size,omitempty"`

	// Level of compression to apply on content selected for aggressive compression. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	LevelAggressive *int32 `json:"level_aggressive,omitempty"`

	// Level of compression to apply on content selected for normal compression. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	LevelNormal *int32 `json:"level_normal,omitempty"`

	// window size used by compression, rounded to the last power of 2. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	WindowSize *int32 `json:"window_size,omitempty"`
}
