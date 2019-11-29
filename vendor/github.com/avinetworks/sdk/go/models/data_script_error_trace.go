package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// DataScriptErrorTrace data script error trace
// swagger:model DataScriptErrorTrace
type DataScriptErrorTrace struct {

	// error of DataScriptErrorTrace.
	Error *string `json:"error,omitempty"`

	// event of DataScriptErrorTrace.
	Event *string `json:"event,omitempty"`

	// stack_trace of DataScriptErrorTrace.
	StackTrace *string `json:"stack_trace,omitempty"`
}
