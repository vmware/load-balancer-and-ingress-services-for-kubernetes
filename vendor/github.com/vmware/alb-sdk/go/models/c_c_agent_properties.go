// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// CCAgentProperties c c agent properties
// swagger:model CC_AgentProperties
type CCAgentProperties struct {

	// Maximum polls to check for async jobs to finish. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AsyncRetries *int32 `json:"async_retries,omitempty"`

	// Delay between each async job status poll check. Unit is SEC. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AsyncRetriesDelay *int32 `json:"async_retries_delay,omitempty"`

	// Discovery poll target duration; a scale factor of 1+ is computed with the actual discovery (actual/target) and used to tweak slow and fast poll intervals. Unit is SEC. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PollDurationTarget *int32 `json:"poll_duration_target,omitempty"`

	// Fast poll interval. Unit is SEC. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PollFastTarget *int32 `json:"poll_fast_target,omitempty"`

	// Slow poll interval. Unit is SEC. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PollSlowTarget *int32 `json:"poll_slow_target,omitempty"`

	// Vcenter host reachability check interval. Allowed values are 60-3600. Field introduced in 22.1.3. Unit is SEC. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	VcenterHostPingInterval *int32 `json:"vcenter_host_ping_interval,omitempty"`

	// Batch size of vcenter inventory updates. Allowed values are 1-500. Field introduced in 22.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	VcenterInventoryMaxObjectUpdates *int32 `json:"vcenter_inventory_max_object_updates,omitempty"`

	// Max datastore processing go routines for vcenter datastore updates. Allowed values are 1-40. Field introduced in 22.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	VcenterMaxDatastoreGoRoutines *int32 `json:"vcenter_max_datastore_go_routines,omitempty"`

	// Reconcile interval for vcenter inventory. Allowed values are 60-3600. Field introduced in 22.1.3. Unit is SEC. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	VcenterReconcileInterval *int32 `json:"vcenter_reconcile_interval,omitempty"`

	// Maximum polls to check for vnics to be attached to VM. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VnicRetries *int32 `json:"vnic_retries,omitempty"`

	// Delay between each vnic status poll check. Unit is SEC. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VnicRetriesDelay *int32 `json:"vnic_retries_delay,omitempty"`
}
