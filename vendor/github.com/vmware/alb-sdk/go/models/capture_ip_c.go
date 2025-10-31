// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// CaptureIPC capture IP c
// swagger:model CaptureIPC
type CaptureIPC struct {

	// Flow del probe filter for SE IPC. Field introduced in 18.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	FlowDelProbe *bool `json:"flow_del_probe,omitempty"`

	// Flow mirror add filter for SE IPC. Field introduced in 18.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	FlowMirrorAdd *bool `json:"flow_mirror_add,omitempty"`

	// Filter for all flow mirror SE IPC. Field introduced in 18.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	FlowMirrorAll *bool `json:"flow_mirror_all,omitempty"`

	// Flow mirror del filter for SE IPC. Field introduced in 18.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	FlowMirrorDel *bool `json:"flow_mirror_del,omitempty"`

	// Flow probe filter for SE IPC. Field introduced in 18.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	FlowProbe *bool `json:"flow_probe,omitempty"`

	// Filter for all flow probe SE IPC. Field introduced in 18.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	FlowProbeAll *bool `json:"flow_probe_all,omitempty"`

	// IPC batched filter for SE IPC. Field introduced in 18.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	IpcBatched *bool `json:"ipc_batched,omitempty"`

	// Filter for incoming IPC request. Field introduced in 18.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	IpcRxReq *bool `json:"ipc_rx_req,omitempty"`

	// Filter for incoming IPC response. Field introduced in 18.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	IpcRxRes *bool `json:"ipc_rx_res,omitempty"`

	// Filter for outgoing IPC request. Field introduced in 18.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	IpcTxReq *bool `json:"ipc_tx_req,omitempty"`

	// Filter for outgoing IPC response. Field introduced in 18.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	IpcTxRes *bool `json:"ipc_tx_res,omitempty"`

	// Vs heart beat filter for SE IPC. Field introduced in 18.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VsHb *bool `json:"vs_hb,omitempty"`
}
