// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// VIMgrIPSubnetRuntime v i mgr IP subnet runtime
// swagger:model VIMgrIPSubnetRuntime
type VIMgrIPSubnetRuntime struct {

	// If true, capable of floating/elastic IP association. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	FipAvailable *bool `json:"fip_available,omitempty"`

	// If fip_available is True, this is list of supported FIP subnets, possibly empty if Cloud does not support such a network list. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	FipSubnetUuids []string `json:"fip_subnet_uuids,omitempty"`

	// If fip_available is True, the list of associated FloatingIP subnets, possibly empty if unsupported or implictly defined by the Cloud. Field introduced in 17.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	FloatingipSubnets []*FloatingIPSubnet `json:"floatingip_subnets,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	IPSubnet *string `json:"ip_subnet,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Name *string `json:"name,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Prefix *IPAddrPrefix `json:"prefix"`

	// True if prefix is primary IP on interface, else false. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Primary *bool `json:"primary,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	RefCount *int32 `json:"ref_count,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeRefCount *int32 `json:"se_ref_count,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UUID *string `json:"uuid,omitempty"`
}
