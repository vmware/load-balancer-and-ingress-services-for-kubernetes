// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SaasLicensingStatus saas licensing status
// swagger:model SaasLicensingStatus
type SaasLicensingStatus struct {

	// Portal connectivity status. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Connected *bool `json:"connected,omitempty"`

	// Status of saas licensing subscription. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Enabled *bool `json:"enabled,omitempty"`

	// Saas license expiry status. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Expired *bool `json:"expired,omitempty"`

	// Message. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Message *string `json:"message,omitempty"`

	// Name. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Name *string `json:"name,omitempty"`

	// Public key. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	PublicKey *string `json:"public_key,omitempty"`

	// Service units reserved on controller. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ReserveServiceUnits *float64 `json:"reserve_service_units,omitempty"`
}
