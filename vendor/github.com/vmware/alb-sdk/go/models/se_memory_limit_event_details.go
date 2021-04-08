package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SeMemoryLimitEventDetails se memory limit event details
// swagger:model SeMemoryLimitEventDetails
type SeMemoryLimitEventDetails struct {

	// Current status of config memory. Field introduced in 18.2.2.
	ConfigMemoryStatus *string `json:"config_memory_status,omitempty"`

	// Heap config memory hard limit. Field introduced in 18.2.2.
	HeapConfigHardLimit *int32 `json:"heap_config_hard_limit,omitempty"`

	// Heap config memory soft limit. Field introduced in 18.2.2.
	HeapConfigSoftLimit *int32 `json:"heap_config_soft_limit,omitempty"`

	// Config memory usage in heap memory. Field introduced in 18.2.2.
	HeapConfigUsage *int32 `json:"heap_config_usage,omitempty"`

	// Connection memory usage in heap memory. Field introduced in 18.2.2.
	HeapConnUsage *int32 `json:"heap_conn_usage,omitempty"`

	// UUID of the SE responsible for this event. It is a reference to an object of type ServiceEngine. Field introduced in 18.2.2.
	SeRef *string `json:"se_ref,omitempty"`

	// Current shm config memory hard limit. Field introduced in 18.2.2.
	ShmConfigHardLimit *int32 `json:"shm_config_hard_limit,omitempty"`

	// Current shm config memory soft limit. Field introduced in 18.2.2.
	ShmConfigSoftLimit *int32 `json:"shm_config_soft_limit,omitempty"`

	// Config memory usage in shared memory. Field introduced in 18.2.2.
	ShmConfigUsage *int32 `json:"shm_config_usage,omitempty"`

	// Connection memory usage in shared memory. Field introduced in 18.2.2.
	ShmConnUsage *int32 `json:"shm_conn_usage,omitempty"`
}
