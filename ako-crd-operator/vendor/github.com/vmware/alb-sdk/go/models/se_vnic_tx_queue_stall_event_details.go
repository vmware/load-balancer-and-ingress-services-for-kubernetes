// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SeVnicTxQueueStallEventDetails se vnic tx queue stall event details
// swagger:model SeVnicTxQueueStallEventDetails
type SeVnicTxQueueStallEventDetails struct {

	// Vnic name. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	IfName *string `json:"if_name,omitempty"`

	// Vnic Linux name. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	LinuxName *string `json:"linux_name,omitempty"`

	// Queue number. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Queue uint32 `json:"queue,omitempty"`

	// UUID of the SE responsible for this event. It is a reference to an object of type ServiceEngine. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeRef *string `json:"se_ref,omitempty"`
}
