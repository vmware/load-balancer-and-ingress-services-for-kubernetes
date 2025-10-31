// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SeHBEventDetails se h b event details
// swagger:model SeHBEventDetails
type SeHBEventDetails struct {

	// HB Request/Response not received. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	HbType *int32 `json:"hb_type,omitempty"`

	// UUID of the SE with which Heartbeat failed. It is a reference to an object of type ServiceEngine. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	RemoteSeRef *string `json:"remote_se_ref,omitempty"`

	// UUID of the SE reporting this event. It is a reference to an object of type ServiceEngine. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ReportingSeRef *string `json:"reporting_se_ref,omitempty"`

	// UUID of the virtual service which is placed on reporting-SE and remote-SE. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VsUUID *string `json:"vs_uuid,omitempty"`
}
