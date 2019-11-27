package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SeBootupCompressionProperties se bootup compression properties
// swagger:model SeBootupCompressionProperties
type SeBootupCompressionProperties struct {

	// Number of buffers to use for compression output.
	BufNum *int32 `json:"buf_num,omitempty"`

	// Size of each buffer used for compression output, this should ideally be a multiple of pagesize.
	BufSize *int32 `json:"buf_size,omitempty"`

	// hash size used by compression, rounded to the last power of 2.
	HashSize *int32 `json:"hash_size,omitempty"`

	// Level of compression to apply on content selected for aggressive compression.
	LevelAggressive *int32 `json:"level_aggressive,omitempty"`

	// Level of compression to apply on content selected for normal compression.
	LevelNormal *int32 `json:"level_normal,omitempty"`

	// window size used by compression, rounded to the last power of 2.
	WindowSize *int32 `json:"window_size,omitempty"`
}
