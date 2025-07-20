// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SeAgentStateCacheProperties se agent state cache properties
// swagger:model SeAgentStateCacheProperties
type SeAgentStateCacheProperties struct {

	// Max elements to flush in one shot from the internal buffer by the statecache thread. Allowed values are 1-10000. Field introduced in 18.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ScBatchBufferFlushLimit *uint32 `json:"sc_batch_buffer_flush_limit,omitempty"`

	// Max elements to dequeue in one shot from the Q by the statecache thread. Allowed values are 1-10000. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ScCfgQBatchDequeueLimit *uint32 `json:"sc_cfg_q_batch_dequeue_limit,omitempty"`

	// Max elements in the config queue between seagent main and the statecache thread. Allowed values are 1-10000. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ScCfgQMaxSize *uint32 `json:"sc_cfg_q_max_size,omitempty"`

	// Max elements to dequeue in one shot from the Q by the statecache thread. Allowed values are 1-10000. Field introduced in 20.1.5. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ScDNSQBatchDequeueLimit *uint32 `json:"sc_dns_q_batch_dequeue_limit,omitempty"`

	// Max elements in the dns queue between seagent main and the statecache thread. Allowed values are 1-10000. Field introduced in 20.1.5. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ScDNSQMaxSize *uint32 `json:"sc_dns_q_max_size,omitempty"`

	// Max time to wait by the statecache thread before cleaning up connection to the controller shard. Allowed values are 1-1000000. Field introduced in 18.2.5. Unit is SECONDS. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ScShardCleanupMaxTime *uint32 `json:"sc_shard_cleanup_max_time,omitempty"`

	// Max elements to dequeue in one shot from the state_ring by the statecache thread. Allowed values are 1-10000. Field introduced in 18.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ScStateRingBatchDequeueLimit *uint32 `json:"sc_state_ring_batch_dequeue_limit,omitempty"`

	// Interval for update of operational states to controller. Allowed values are 1-10000. Field introduced in 18.2.5. Unit is MILLISECONDS. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ScStatesFlushInterval *uint32 `json:"sc_states_flush_interval,omitempty"`

	// Interval for checking health of grpc streams to statecache_mgr. Allowed values are 1-90000. Field introduced in 18.2.5. Unit is MILLISECONDS. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ScStreamCheckInterval *uint32 `json:"sc_stream_check_interval,omitempty"`

	// Max elements to dequeue in one shot from the Q by the statecache thread. Allowed values are 1-10000. Field introduced in 18.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ScThreadQBatchDequeueLimit *uint32 `json:"sc_thread_q_batch_dequeue_limit,omitempty"`

	// Max elements in the Q between seagent main and the statecache thread. Allowed values are 1-10000. Field introduced in 18.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ScThreadQMaxSize *uint32 `json:"sc_thread_q_max_size,omitempty"`

	// Interval for grpc thread to sleep between doing work. Allowed values are 1-10000. Field introduced in 18.2.5. Unit is MILLISECONDS. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ScThreadSleepInterval *uint32 `json:"sc_thread_sleep_interval,omitempty"`
}
