package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SeFault se fault
// swagger:model SeFault
type SeFault struct {

	// Optional 64 bit unsigned integer that can be used within the enabled fault. Field introduced in 18.2.7.
	Arg *int64 `json:"arg,omitempty"`

	// The name of the target fault. Field introduced in 18.2.7.
	// Required: true
	FaultName *string `json:"fault_name"`

	// The name of the function that contains the target fault. Field introduced in 18.2.7.
	FunctionName *string `json:"function_name,omitempty"`

	// Number of times the fault should be executed. Allowed values are 1-4294967295. Field introduced in 18.2.7.
	NumExecutions *int32 `json:"num_executions,omitempty"`
}
