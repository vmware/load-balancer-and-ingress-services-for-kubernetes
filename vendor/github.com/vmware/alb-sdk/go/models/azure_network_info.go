// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// AzureNetworkInfo azure network info
// swagger:model AzureNetworkInfo
type AzureNetworkInfo struct {

	// Id of the Azure subnet used as management network for the Service Engines. If set, Service Engines will have a dedicated management NIC, otherwise, they operate in inband mode. Field introduced in 18.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ManagementNetworkID *string `json:"management_network_id,omitempty"`

	// Id of the Azure subnet where Avi Controller will create the Service Engines. . Field introduced in 17.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeNetworkID *string `json:"se_network_id,omitempty"`

	// Virtual network where Virtual IPs will belong. Field introduced in 17.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VirtualNetworkID *string `json:"virtual_network_id,omitempty"`
}
