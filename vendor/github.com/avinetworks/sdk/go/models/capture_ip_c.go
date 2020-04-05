package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// CaptureIPC capture IP c
// swagger:model CaptureIPC
type CaptureIPC struct {

	// Flow del probe filter for SE IPC. Field introduced in 18.2.5.
	FlowDelProbe *bool `json:"flow_del_probe,omitempty"`

	// Flow mirror add filter for SE IPC. Field introduced in 18.2.5.
	FlowMirrorAdd *bool `json:"flow_mirror_add,omitempty"`

	// Filter for all flow mirror SE IPC. Field introduced in 18.2.5.
	FlowMirrorAll *bool `json:"flow_mirror_all,omitempty"`

	// Flow mirror del filter for SE IPC. Field introduced in 18.2.5.
	FlowMirrorDel *bool `json:"flow_mirror_del,omitempty"`

	// Flow probe filter for SE IPC. Field introduced in 18.2.5.
	FlowProbe *bool `json:"flow_probe,omitempty"`

	// Filter for all flow probe SE IPC. Field introduced in 18.2.5.
	FlowProbeAll *bool `json:"flow_probe_all,omitempty"`

	// IPC batched filter for SE IPC. Field introduced in 18.2.5.
	IpcBatched *bool `json:"ipc_batched,omitempty"`

	// Filter for incoming IPC request. Field introduced in 18.2.5.
	IpcRxReq *bool `json:"ipc_rx_req,omitempty"`

	// Filter for incoming IPC response. Field introduced in 18.2.5.
	IpcRxRes *bool `json:"ipc_rx_res,omitempty"`

	// Filter for outgoing IPC request. Field introduced in 18.2.5.
	IpcTxReq *bool `json:"ipc_tx_req,omitempty"`

	// Filter for outgoing IPC response. Field introduced in 18.2.5.
	IpcTxRes *bool `json:"ipc_tx_res,omitempty"`

	// Vs heart beat filter for SE IPC. Field introduced in 18.2.5.
	VsHb *bool `json:"vs_hb,omitempty"`
}
