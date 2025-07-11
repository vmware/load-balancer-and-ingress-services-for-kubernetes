// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// GCPInBandManagement g c p in band management
// swagger:model GCPInBandManagement
type GCPInBandManagement struct {

	// Service Engine Network Name. Field introduced in 18.2.2. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	VpcNetworkName *string `json:"vpc_network_name"`

	// Project ID of the Service Engine Network. By default, Service Engine Project ID will be used. Field introduced in 18.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VpcProjectID *string `json:"vpc_project_id,omitempty"`

	// Service Engine Network Subnet Name. Field introduced in 18.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	VpcSubnetName *string `json:"vpc_subnet_name"`
}
