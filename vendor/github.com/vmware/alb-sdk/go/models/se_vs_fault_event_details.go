// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SeVsFaultEventDetails se vs fault event details
// swagger:model SeVsFaultEventDetails
type SeVsFaultEventDetails struct {

	// Name of the object responsible for the fault. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	FaultObject *string `json:"fault_object,omitempty"`

	// Reason for the fault. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	FaultReason *string `json:"fault_reason,omitempty"`

	// SE uuid. It is a reference to an object of type ServiceEngine. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ServiceEngine *string `json:"service_engine,omitempty"`

	// VS name. It is a reference to an object of type VirtualService. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VirtualService *string `json:"virtual_service,omitempty"`
}
