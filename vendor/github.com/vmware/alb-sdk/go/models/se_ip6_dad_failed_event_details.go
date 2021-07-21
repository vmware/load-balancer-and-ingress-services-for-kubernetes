// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SeIp6DadFailedEventDetails se ip6 dad failed event details
// swagger:model SeIP6DadFailedEventDetails
type SeIp6DadFailedEventDetails struct {

	// IPv6 address.
	DadIP *IPAddr `json:"dad_ip,omitempty"`

	// Vnic name.
	IfName *string `json:"if_name,omitempty"`

	// UUID of the SE responsible for this event. It is a reference to an object of type ServiceEngine.
	SeRef *string `json:"se_ref,omitempty"`
}
