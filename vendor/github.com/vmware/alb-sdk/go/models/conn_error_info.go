package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ConnErrorInfo conn error info
// swagger:model ConnErrorInfo
type ConnErrorInfo struct {

	// Number of num_syn_retransmit.
	NumSynRetransmit *int32 `json:"num_syn_retransmit,omitempty"`

	// Number of num_window_shrink.
	NumWindowShrink *int32 `json:"num_window_shrink,omitempty"`

	// Number of out_of_orders.
	OutOfOrders *int32 `json:"out_of_orders,omitempty"`

	// Number of retransmits.
	Retransmits *int32 `json:"retransmits,omitempty"`

	// Number of rx_pkts.
	RxPkts *int64 `json:"rx_pkts,omitempty"`

	// Number of server_num_window_shrink.
	ServerNumWindowShrink *int32 `json:"server_num_window_shrink,omitempty"`

	// Number of server_out_of_orders.
	ServerOutOfOrders *int32 `json:"server_out_of_orders,omitempty"`

	// Number of server_retransmits.
	ServerRetransmits *int32 `json:"server_retransmits,omitempty"`

	// Number of server_rx_pkts.
	ServerRxPkts *int64 `json:"server_rx_pkts,omitempty"`

	// Number of server_timeouts.
	ServerTimeouts *int32 `json:"server_timeouts,omitempty"`

	// Number of server_tx_pkts.
	ServerTxPkts *int64 `json:"server_tx_pkts,omitempty"`

	// Number of server_zero_window_size_events.
	ServerZeroWindowSizeEvents *int64 `json:"server_zero_window_size_events,omitempty"`

	// Number of timeouts.
	Timeouts *int32 `json:"timeouts,omitempty"`

	// Number of tx_pkts.
	TxPkts *int64 `json:"tx_pkts,omitempty"`

	// Number of zero_window_size_events.
	ZeroWindowSizeEvents *int64 `json:"zero_window_size_events,omitempty"`
}
