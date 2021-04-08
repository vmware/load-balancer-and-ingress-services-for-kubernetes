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

	// Aggressive Timeout used for se health check. Unit is MILLISECONDS.
	ControllerEchoRPCAggressiveTimeout *int32 `json:"controller_echo_rpc_aggressive_timeout,omitempty"`

	// Timeout used for se health check. Unit is MILLISECONDS.
	ControllerEchoRPCTimeout *int32 `json:"controller_echo_rpc_timeout,omitempty"`

	//  Allowed values are 1-20.
	ControllerHeartbeatMissLimit *int32 `json:"controller_heartbeat_miss_limit,omitempty"`

	//  Allowed values are 1-60. Unit is SEC.
	ControllerHeartbeatTimeoutSec *int32 `json:"controller_heartbeat_timeout_sec,omitempty"`

	//  Unit is SEC.
	ControllerRegistrationTimeoutSec *int32 `json:"controller_registration_timeout_sec,omitempty"`

	//  Unit is SEC.
	ControllerRPCTimeout *int32 `json:"controller_rpc_timeout,omitempty"`

	//  Unit is SEC.
	CpustatsInterval *int32 `json:"cpustats_interval,omitempty"`

	// Max time to wait for ctrl registration before assert. Allowed values are 1-1000. Unit is SEC.
	CtrlRegPendingMaxWaitTime *int32 `json:"ctrl_reg_pending_max_wait_time,omitempty"`

	// Placeholder for description of property debug_mode of obj type SeAgentProperties field type str  type boolean
	DebugMode *bool `json:"debug_mode,omitempty"`

	// Deprecated in 21.1.1. Use dp_aggressive_deq_interval_msec in ServiceEngineGroup instead. Allowed values are 1-1000. Unit is MILLISECONDS.
	DpAggressiveDeqIntervalMsec *int32 `json:"dp_aggressive_deq_interval_msec,omitempty"`

	// Deprecated in 21.1.1. Use dp_aggressive_enq_interval_msec in ServiceEngineGroup instead. Allowed values are 1-1000. Unit is MILLISECONDS.
	DpAggressiveEnqIntervalMsec *int32 `json:"dp_aggressive_enq_interval_msec,omitempty"`

	// Number of dp_batch_size.
	DpBatchSize *int32 `json:"dp_batch_size,omitempty"`

	// Deprecated in 21.1.1. Use dp_deq_interval_msec in ServiceEngineGroup instead. Allowed values are 1-1000. Unit is MILLISECONDS.
	DpDeqIntervalMsec *int32 `json:"dp_deq_interval_msec,omitempty"`

	// Deprecated in 21.1.1. Use dp_enq_interval_msec in ServiceEngineGroup instead. Allowed values are 1-1000. Unit is MILLISECONDS.
	DpEnqIntervalMsec *int32 `json:"dp_enq_interval_msec,omitempty"`

	//  Unit is SEC.
	DpMaxWaitRspTimeSec *int32 `json:"dp_max_wait_rsp_time_sec,omitempty"`

	// Max time to wait for dp registration before assert. Unit is SEC.
	DpRegPendingMaxWaitTime *int32 `json:"dp_reg_pending_max_wait_time,omitempty"`

	//  Unit is SEC.
	HeadlessTimeoutSec *int32 `json:"headless_timeout_sec,omitempty"`

	// Placeholder for description of property ignore_docker_mac_change of obj type SeAgentProperties field type str  type boolean
	IgnoreDockerMacChange *bool `json:"ignore_docker_mac_change,omitempty"`

	// Dequeue interval for receive queue from NS HELPER. Deprecated in 21.1.1. Use ns_helper_deq_interval_msec in ServiceEngineGroup instead. Allowed values are 1-1000. Field introduced in 17.2.13, 18.1.3, 18.2.1. Unit is MILLISECONDS.
	NsHelperDeqIntervalMsec *int32 `json:"ns_helper_deq_interval_msec,omitempty"`

	// SDB pipeline flush interval. Allowed values are 1-10000. Unit is MILLISECONDS.
	SdbFlushInterval *int32 `json:"sdb_flush_interval,omitempty"`

	// SDB pipeline size. Allowed values are 1-10000.
	SdbPipelineSize *int32 `json:"sdb_pipeline_size,omitempty"`

	// SDB scan count. Allowed values are 1-1000.
	SdbScanCount *int32 `json:"sdb_scan_count,omitempty"`

	// Internal flag used to decide if SE restart is needed,when the se-group is changed. Field introduced in 18.2.3.
	SeGrpChangeDisruptive *bool `json:"se_grp_change_disruptive,omitempty"`

	// SeAgent properties for State Cache functionality. Field introduced in 18.2.5.
	SeagentStatecacheProperties *SeAgentStateCacheProperties `json:"seagent_statecache_properties,omitempty"`

	// Timeout for sending SE_READY without NS HELPER registration completion. Deprecated in 21.1.1. Use send_se_ready_timeout ServiceEngineGroup instead. Allowed values are 10-600. Field introduced in 17.2.13, 18.1.3, 18.2.1. Unit is SECONDS.
	SendSeReadyTimeout *int32 `json:"send_se_ready_timeout,omitempty"`

	// Interval for update of operational states to controller. Allowed values are 1-10000. Field introduced in 18.2.1, 17.2.14, 18.1.5. Unit is MILLISECONDS.
	StatesFlushInterval *int32 `json:"states_flush_interval,omitempty"`

	// DHCP ip check interval. Deprecated in 21.1.1. Use vnic_dhcp_ip_check_interval instead. Allowed values are 1-1000. Unit is SEC.
	VnicDhcpIPCheckInterval *int32 `json:"vnic_dhcp_ip_check_interval,omitempty"`

	// DHCP ip max retries. Deprecated in 21.1.1. Use vnic_dhcp_ip_max_retries ServiceEngineGroup instead.
	VnicDhcpIPMaxRetries *int32 `json:"vnic_dhcp_ip_max_retries,omitempty"`

	// wait interval before deleting IP. Deprecated in 21.1.1. Use vnic_ip_delete_interval ServiceEngineGroup instead. Unit is SEC.
	VnicIPDeleteInterval *int32 `json:"vnic_ip_delete_interval,omitempty"`

	// Probe vnic interval. Deprecated in 21.1.1. Use vnic_probe_interval ServiceEngineGroup instead. Unit is SEC.
	VnicProbeInterval *int32 `json:"vnic_probe_interval,omitempty"`

	// Time interval for retrying the failed VNIC RPC requestsDeprecated in 21.1.1. Use vnic_rpc_retry_interval ServiceEngineGroup instead. Field introduced in 18.2.6. Unit is SEC.
	VnicRPCRetryInterval *int32 `json:"vnic_rpc_retry_interval,omitempty"`

	// Size of vnicdb command history. Deprecated in 21.1.1. Use vnicdb_cmd_history_size ServiceEngineGroup instead. Allowed values are 0-65535. Field introduced in 18.2.3.
	VnicdbCmdHistorySize *int32 `json:"vnicdb_cmd_history_size,omitempty"`
}
