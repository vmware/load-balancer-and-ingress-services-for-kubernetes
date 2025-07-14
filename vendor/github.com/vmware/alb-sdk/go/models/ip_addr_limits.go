// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// IPAddrLimits IP addr limits
// swagger:model IPAddrLimits
type IPAddrLimits struct {

	// Number of IP address groups for match criteria. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	IPAddressGroupPerMatchCriteria *int32 `json:"ip_address_group_per_match_criteria,omitempty"`

	// Number of IP address prefixes for match criteria. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	IPAddressPrefixPerMatchCriteria *int32 `json:"ip_address_prefix_per_match_criteria,omitempty"`

	// Number of IP address ranges for match criteria. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	IPAddressRangePerMatchCriteria *int32 `json:"ip_address_range_per_match_criteria,omitempty"`

	// Number of IP addresses for match criteria. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	IPAddressesPerMatchCriteria *int32 `json:"ip_addresses_per_match_criteria,omitempty"`
}
