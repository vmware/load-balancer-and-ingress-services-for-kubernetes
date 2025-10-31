// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// AlertRuleEvent alert rule event
// swagger:model AlertRuleEvent
type AlertRuleEvent struct {

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	EventDetails []*EventDetailsFilter `json:"event_details,omitempty"`

	// When the selected event occurs, trigger this alert. Enum options - VINFRA_DISC_DC, VINFRA_DISC_HOST, VINFRA_DISC_CLUSTER, VINFRA_DISC_VM, VINFRA_DISC_NW, MGMT_NW_NAME_CHANGED, DISCOVERY_DATACENTER_DEL, VM_ADDED, VM_REMOVED, VINFRA_DISC_COMPLETE, VCENTER_ADDRESS_ERROR, SE_GROUP_CLUSTER_DEL, SE_GROUP_MGMT_NW_DEL, MGMT_NW_DEL, VCENTER_BAD_CREDENTIALS, ESX_HOST_UNREACHABLE, SERVER_DELETED, SE_GROUP_HOST_DEL, VINFRA_DISC_FAILURE, ESX_HOST_POWERED_DOWN.... Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	EventID *string `json:"event_id,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NotCond *bool `json:"not_cond,omitempty"`
}
