package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SeAgentProperties se agent properties
// swagger:model SeAgentProperties
type SeAgentProperties struct {

	// Aggressive Count of HB misses for se health check failure. Allowed values are 1-1000.
	ControllerEchoMissAggressiveLimit *int32 `json:"controller_echo_miss_aggressive_limit,omitempty"`

	// Count of HB misses for se health check failure. Allowed values are 1-1000.
	ControllerEchoMissLimit *int32 `json:"controller_echo_miss_limit,omitempty"`

	// Aggressive Timeout used for se health check.
	ControllerEchoRPCAggressiveTimeout *int32 `json:"controller_echo_rpc_aggressive_timeout,omitempty"`

	// Timeout used for se health check.
	ControllerEchoRPCTimeout *int32 `json:"controller_echo_rpc_timeout,omitempty"`

	//  Allowed values are 1-20.
	ControllerHeartbeatMissLimit *int32 `json:"controller_heartbeat_miss_limit,omitempty"`

	//  Allowed values are 1-60.
	ControllerHeartbeatTimeoutSec *int32 `json:"controller_heartbeat_timeout_sec,omitempty"`

	// Number of controller_registration_timeout_sec.
	ControllerRegistrationTimeoutSec *int32 `json:"controller_registration_timeout_sec,omitempty"`

	// Number of controller_rpc_timeout.
	ControllerRPCTimeout *int32 `json:"controller_rpc_timeout,omitempty"`

	// Number of cpustats_interval.
	CpustatsInterval *int32 `json:"cpustats_interval,omitempty"`

	// Max time to wait for ctrl registration before assert. Allowed values are 1-1000.
	CtrlRegPendingMaxWaitTime *int32 `json:"ctrl_reg_pending_max_wait_time,omitempty"`

	// Placeholder for description of property debug_mode of obj type SeAgentProperties field type str  type boolean
	DebugMode *bool `json:"debug_mode,omitempty"`

	//  Allowed values are 1-1000.
	DpAggressiveDeqIntervalMsec *int32 `json:"dp_aggressive_deq_interval_msec,omitempty"`

	//  Allowed values are 1-1000.
	DpAggressiveEnqIntervalMsec *int32 `json:"dp_aggressive_enq_interval_msec,omitempty"`

	// Number of dp_batch_size.
	DpBatchSize *int32 `json:"dp_batch_size,omitempty"`

	//  Allowed values are 1-1000.
	DpDeqIntervalMsec *int32 `json:"dp_deq_interval_msec,omitempty"`

	//  Allowed values are 1-1000.
	DpEnqIntervalMsec *int32 `json:"dp_enq_interval_msec,omitempty"`

	// Number of dp_max_wait_rsp_time_sec.
	DpMaxWaitRspTimeSec *int32 `json:"dp_max_wait_rsp_time_sec,omitempty"`

	// Max time to wait for dp registration before assert.
	DpRegPendingMaxWaitTime *int32 `json:"dp_reg_pending_max_wait_time,omitempty"`

	// Number of headless_timeout_sec.
	HeadlessTimeoutSec *int32 `json:"headless_timeout_sec,omitempty"`

	// Placeholder for description of property ignore_docker_mac_change of obj type SeAgentProperties field type str  type boolean
	IgnoreDockerMacChange *bool `json:"ignore_docker_mac_change,omitempty"`

	// Dequeue interval for receive queue from NS HELPER. Allowed values are 1-1000. Field introduced in 17.2.13, 18.1.3, 18.2.1.
	NsHelperDeqIntervalMsec *int32 `json:"ns_helper_deq_interval_msec,omitempty"`

	// SDB pipeline flush interval. Allowed values are 1-10000.
	SdbFlushInterval *int32 `json:"sdb_flush_interval,omitempty"`

	// SDB pipeline size. Allowed values are 1-10000.
	SdbPipelineSize *int32 `json:"sdb_pipeline_size,omitempty"`

	// SDB scan count. Allowed values are 1-1000.
	SdbScanCount *int32 `json:"sdb_scan_count,omitempty"`

	// Timeout for sending SE_READY without NS HELPER registration completion. Allowed values are 10-600. Field introduced in 17.2.13, 18.1.3, 18.2.1.
	SendSeReadyTimeout *int32 `json:"send_se_ready_timeout,omitempty"`

	// Interval for update of operational states to controller. Allowed values are 1-10000. Field introduced in 18.2.1, 17.2.14, 18.1.5.
	StatesFlushInterval *int32 `json:"states_flush_interval,omitempty"`

	// DHCP ip check interval. Allowed values are 1-1000.
	VnicDhcpIPCheckInterval *int32 `json:"vnic_dhcp_ip_check_interval,omitempty"`

	// DHCP ip max retries.
	VnicDhcpIPMaxRetries *int32 `json:"vnic_dhcp_ip_max_retries,omitempty"`

	// wait interval before deleting IP.
	VnicIPDeleteInterval *int32 `json:"vnic_ip_delete_interval,omitempty"`

	// Probe vnic interval.
	VnicProbeInterval *int32 `json:"vnic_probe_interval,omitempty"`
}
