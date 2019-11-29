package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// DebugSeFault debug se fault
// swagger:model DebugSeFault
type DebugSeFault struct {

	// Fail SE malloc type at this frequency. Field introduced in 18.1.2.
	SeMallocFailFrequency *int32 `json:"se_malloc_fail_frequency,omitempty"`

	// Fail this SE malloc type. Field introduced in 18.1.2.
	SeMallocFailType *int32 `json:"se_malloc_fail_type,omitempty"`

	// Toggle assert on mbuf cluster sanity check fail. Field introduced in 17.2.13,18.1.3,18.2.1.
	SeMbufClSanity *bool `json:"se_mbuf_cl_sanity,omitempty"`

	// Fail SE SHM malloc type at this frequency. Field introduced in 18.1.2.
	SeShmMallocFailFrequency *int32 `json:"se_shm_malloc_fail_frequency,omitempty"`

	// Fail this SE SHM malloc type. Field introduced in 18.1.2.
	SeShmMallocFailType *int32 `json:"se_shm_malloc_fail_type,omitempty"`

	// Fail SE WAF allocation at this frequency. Field introduced in 18.1.2.
	SeWafAllocFailFrequency *int32 `json:"se_waf_alloc_fail_frequency,omitempty"`

	// Fail SE WAF learning allocation at this frequency. Field introduced in 18.1.2.
	SeWafLearningAllocFailFrequency *int32 `json:"se_waf_learning_alloc_fail_frequency,omitempty"`
}
