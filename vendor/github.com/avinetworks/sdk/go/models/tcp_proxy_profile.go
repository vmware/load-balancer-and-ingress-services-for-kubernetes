package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// TCPProxyProfile TCP proxy profile
// swagger:model TCPProxyProfile
type TCPProxyProfile struct {

	// Controls the our congestion window to send, normally it's 1 mss, If this option is turned on, we use 10 msses.
	AggressiveCongestionAvoidance *bool `json:"aggressive_congestion_avoidance,omitempty"`

	// Dynamically pick the relevant parameters for connections.
	Automatic *bool `json:"automatic,omitempty"`

	// Controls the congestion control algorithm we use. Enum options - CC_ALGO_NEW_RENO, CC_ALGO_CUBIC, CC_ALGO_HTCP.
	CcAlgo *string `json:"cc_algo,omitempty"`

	// Congestion window scaling factor after recovery. Allowed values are 0-8. Field introduced in 17.2.12, 18.1.3, 18.2.1.
	CongestionRecoveryScalingFactor *int32 `json:"congestion_recovery_scaling_factor,omitempty"`

	// The duration for keepalive probes or session idle timeout. Max value is 3600 seconds, min is 5.  Set to 0 to allow infinite idle time. Allowed values are 5-14400. Special values are 0 - 'infinite'.
	IDLEConnectionTimeout *int32 `json:"idle_connection_timeout,omitempty"`

	// Controls the behavior of idle connections. Enum options - KEEP_ALIVE, CLOSE_IDLE.
	IDLEConnectionType *string `json:"idle_connection_type,omitempty"`

	// A new SYN is accepted from the same 4-tuple even if there is already a connection in TIME_WAIT state.  This is equivalent of setting Time Wait Delay to 0.
	IgnoreTimeWait *bool `json:"ignore_time_wait,omitempty"`

	// Controls the value of the Differentiated Services Code Point field inserted in the IP header.  This has two options   Set to a specific value, or Pass Through, which uses the incoming DSCP value. Allowed values are 0-63. Special values are MAX - 'Passthrough'.
	IPDscp *int32 `json:"ip_dscp,omitempty"`

	// Controls whether to keep the connection alive with keepalive messages in the TCP half close state. The interval for sending keepalive messages is 30s. If a timeout is already configured in the network profile, this will not override it. Field introduced in 18.2.6.
	KeepaliveInHalfcloseState *bool `json:"keepalive_in_halfclose_state,omitempty"`

	// The number of attempts at retransmit before closing the connection. Allowed values are 3-8.
	MaxRetransmissions *int32 `json:"max_retransmissions,omitempty"`

	// Maximum TCP segment size. Allowed values are 512-9000. Special values are 0 - 'Use Interface MTU'.
	MaxSegmentSize *int32 `json:"max_segment_size,omitempty"`

	// The maximum number of attempts at retransmitting a SYN packet before giving up. Allowed values are 3-8.
	MaxSynRetransmissions *int32 `json:"max_syn_retransmissions,omitempty"`

	// The minimum wait time (in millisec) to retransmit packet. Allowed values are 50-5000. Field introduced in 17.2.8.
	MinRexmtTimeout *int32 `json:"min_rexmt_timeout,omitempty"`

	// Consolidates small data packets to send clients fewer but larger packets.  Adversely affects real time protocols such as telnet or SSH.
	NaglesAlgorithm *bool `json:"nagles_algorithm,omitempty"`

	// Maximum number of TCP segments that can be queued for reassembly. Configuring this to 0 disables the feature and provides unlimited queuing. Field introduced in 17.2.13, 18.1.4, 18.2.1.
	ReassemblyQueueSize *int32 `json:"reassembly_queue_size,omitempty"`

	// Size of the receive window. Allowed values are 2-65536.
	ReceiveWindow *int32 `json:"receive_window,omitempty"`

	// Controls the number of duplicate acks required to trigger retransmission. Setting a higher value reduces retransmission caused by packet reordering. A larger value is recommended in public cloud environments where packet reordering is quite common. The default value is 8 in public cloud platforms (AWS, Azure, GCP), and 3 in other environments. Allowed values are 1-100. Field introduced in 17.2.7.
	ReorderThreshold *int32 `json:"reorder_threshold,omitempty"`

	// Congestion window scaling factor during slow start. Allowed values are 0-8. Field introduced in 17.2.12, 18.1.3, 18.2.1.
	SlowStartScalingFactor *int32 `json:"slow_start_scaling_factor,omitempty"`

	// The time (in millisec) to wait before closing a connection in the TIME_WAIT state. Allowed values are 500-2000. Special values are 0 - 'immediate'.
	TimeWaitDelay *int32 `json:"time_wait_delay,omitempty"`

	// Use the interface MTU to calculate the TCP max segment size.
	UseInterfaceMtu *bool `json:"use_interface_mtu,omitempty"`
}
