package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// DebugVirtualServiceObjSync debug virtual service obj sync
// swagger:model DebugVirtualServiceObjSync
type DebugVirtualServiceObjSync struct {

	// Triggers Initial Sync on all the SEs of this VS. Field introduced in 20.1.3.
	TriggerInitialSync *bool `json:"trigger_initial_sync,omitempty"`
}
