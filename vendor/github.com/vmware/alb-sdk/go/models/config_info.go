package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ConfigInfo config info
// swagger:model ConfigInfo
type ConfigInfo struct {

	// Placeholder for description of property queue of obj type ConfigInfo field type str  type object
	Queue []*VersionInfo `json:"queue,omitempty"`

	// Number of reader_count.
	ReaderCount *int32 `json:"reader_count,omitempty"`

	//  Enum options - REPL_NONE, REPL_ENABLED, REPL_DISABLED.
	State *string `json:"state,omitempty"`

	// Number of writer_count.
	WriterCount *int32 `json:"writer_count,omitempty"`
}
