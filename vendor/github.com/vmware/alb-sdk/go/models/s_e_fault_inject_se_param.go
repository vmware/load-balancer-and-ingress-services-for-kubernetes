package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SEFaultInjectSeParam s e fault inject se param
// swagger:model SEFaultInjectSeParam
type SEFaultInjectSeParam struct {

	// Inject fault in specific core. Field introduced in 18.1.5,18.2.1.
	Core *int32 `json:"core,omitempty"`

	// Inject fault in random no of cores. Field introduced in 18.1.5,18.2.1.
	RandomCore *bool `json:"random_core,omitempty"`

	// Set se-agent fault type. Enum options - SE_AGENT_FAULT_DISABLED, SE_AGENT_PRE_PROCESS_FAULT, SE_AGENT_POST_PROCESS_FAULT, SE_AGENT_DP_RESPONSE_FAULT, SE_AGENT_RANDOM_PROCESS_FAULT. Field introduced in 18.1.5,18.2.1.
	SeAgentFault *string `json:"se_agent_fault,omitempty"`

	// Set se-dp fault type. Enum options - SE_DP_FAULT_DISABLED. Field introduced in 18.1.5,18.2.1.
	SeDpFault *string `json:"se_dp_fault,omitempty"`
}
