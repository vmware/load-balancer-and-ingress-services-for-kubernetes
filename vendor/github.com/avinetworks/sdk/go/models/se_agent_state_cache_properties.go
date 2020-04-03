package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SeAgentStateCacheProperties se agent state cache properties
// swagger:model SeAgentStateCacheProperties
type SeAgentStateCacheProperties struct {

	// Max elements to flush in one shot from the internal buffer by the statecache thread. Allowed values are 1-10000. Field introduced in 18.2.5.
	ScBatchBufferFlushLimit *int32 `json:"sc_batch_buffer_flush_limit,omitempty"`

	// Max time to wait by the statecache thread before cleaning up connection to the controller shard. Allowed values are 1-1000000. Field introduced in 18.2.5.
	ScShardCleanupMaxTime *int32 `json:"sc_shard_cleanup_max_time,omitempty"`

	// Max elements to dequeue in one shot from the state_ring by the statecache thread. Allowed values are 1-10000. Field introduced in 18.2.5.
	ScStateRingBatchDequeueLimit *int32 `json:"sc_state_ring_batch_dequeue_limit,omitempty"`

	// Interval for update of operational states to controller. Allowed values are 1-10000. Field introduced in 18.2.5.
	ScStatesFlushInterval *int32 `json:"sc_states_flush_interval,omitempty"`

	// Interval for checking health of grpc streams to statecache_mgr. Allowed values are 1-90000. Field introduced in 18.2.5.
	ScStreamCheckInterval *int32 `json:"sc_stream_check_interval,omitempty"`

	// Max elements to dequeue in one shot from the Q by the statecache thread. Allowed values are 1-10000. Field introduced in 18.2.5.
	ScThreadQBatchDequeueLimit *int32 `json:"sc_thread_q_batch_dequeue_limit,omitempty"`

	// Max elements in the Q between seagent main and the statecache thread. Allowed values are 1-10000. Field introduced in 18.2.5.
	ScThreadQMaxSize *int32 `json:"sc_thread_q_max_size,omitempty"`

	// Interval for grpc thread to sleep between doing work. Allowed values are 1-10000. Field introduced in 18.2.5.
	ScThreadSleepInterval *int32 `json:"sc_thread_sleep_interval,omitempty"`
}
