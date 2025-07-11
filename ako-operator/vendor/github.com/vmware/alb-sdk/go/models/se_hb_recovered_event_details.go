// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SeHbRecoveredEventDetails se hb recovered event details
// swagger:model SeHbRecoveredEventDetails
type SeHbRecoveredEventDetails struct {

	// Heartbeat Request/Response received. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	HbType *int32 `json:"hb_type,omitempty"`

	// UUID of the remote SE with which dataplane heartbeat recovered. It is a reference to an object of type ServiceEngine. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	RemoteSeRef *string `json:"remote_se_ref,omitempty"`

	// UUID of the SE reporting this event. It is a reference to an object of type ServiceEngine. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ReportingSeRef *string `json:"reporting_se_ref,omitempty"`

	// UUID of a VS which is placed on reporting-SE and remote-SE. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VsUUID *string `json:"vs_uuid,omitempty"`
}
