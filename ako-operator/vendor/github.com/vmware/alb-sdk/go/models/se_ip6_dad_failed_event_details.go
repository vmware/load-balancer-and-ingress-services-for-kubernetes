// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SeIp6DadFailedEventDetails se ip6 dad failed event details
// swagger:model SeIP6DadFailedEventDetails
type SeIp6DadFailedEventDetails struct {

	// IPv6 address. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DadIP *IPAddr `json:"dad_ip,omitempty"`

	// Vnic name. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	IfName *string `json:"if_name,omitempty"`

	// UUID of the SE responsible for this event. It is a reference to an object of type ServiceEngine. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeRef *string `json:"se_ref,omitempty"`
}
