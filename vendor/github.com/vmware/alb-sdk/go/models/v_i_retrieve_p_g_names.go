// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// VIRetrievePGNames v i retrieve p g names
// swagger:model VIRetrievePGNames
type VIRetrievePGNames struct {

	// Unique object identifier of cloud.
	CloudUUID *string `json:"cloud_uuid,omitempty"`

	// datacenter of VIRetrievePGNames.
	Datacenter *string `json:"datacenter,omitempty"`

	// password of VIRetrievePGNames.
	Password *string `json:"password,omitempty"`

	// username of VIRetrievePGNames.
	Username *string `json:"username,omitempty"`

	// vcenter_url of VIRetrievePGNames.
	VcenterURL *string `json:"vcenter_url,omitempty"`
}
