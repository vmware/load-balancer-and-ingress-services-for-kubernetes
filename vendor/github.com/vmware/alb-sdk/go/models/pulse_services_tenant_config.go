// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// PulseServicesTenantConfig pulse services tenant config
// swagger:model PulseServicesTenantConfig
type PulseServicesTenantConfig struct {

	// Heartbeat Interval duration. Field introduced in 30.2.1. Unit is MIN. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	HeartbeatInterval *uint32 `json:"heartbeat_interval,omitempty"`

	// License Escrow Interval duration. Field introduced in 30.2.1. Unit is MIN. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	LicenseEscrowInterval *uint32 `json:"license_escrow_interval,omitempty"`

	// License Expiry Interval duration. Allowed values are 1-1440. Field introduced in 30.2.1. Unit is MIN. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	LicenseExpiryInterval *uint32 `json:"license_expiry_interval,omitempty"`

	// License Reconcile Interval duration. Field introduced in 30.2.1. Unit is MIN. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	LicenseReconcileInterval *uint32 `json:"license_reconcile_interval,omitempty"`

	// License Refresh Interval duration. Field introduced in 30.2.1. Unit is MIN. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	LicenseRefreshInterval *uint32 `json:"license_refresh_interval,omitempty"`

	// License Renewal Interval duration. Allowed values are 1-1440. Field introduced in 30.2.1. Unit is MIN. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	LicenseRenewalInterval *uint32 `json:"license_renewal_interval,omitempty"`

	// Token Refresh Interval duration. Allowed values are 1-240. Field introduced in 30.2.1. Unit is MIN. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	TokenRefreshInterval *uint32 `json:"token_refresh_interval,omitempty"`
}
