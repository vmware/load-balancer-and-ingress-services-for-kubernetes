// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// IPAMDNSTencentProfile ipam Dns tencent profile
// swagger:model IpamDnsTencentProfile
type IPAMDNSTencentProfile struct {

	// Credentials to access Tencent cloud. It is a reference to an object of type CloudConnectorUser. Field introduced in 18.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CloudCredentialsRef *string `json:"cloud_credentials_ref,omitempty"`

	// VPC region. Field introduced in 18.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Region *string `json:"region"`

	// Usable networks for Virtual IP. If VirtualService does not specify a network and auto_allocate_ip is set, then the first available network from this list will be chosen for IP allocation. Field introduced in 18.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UsableSubnetIds []string `json:"usable_subnet_ids,omitempty"`

	// VPC ID. Field introduced in 18.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	VpcID *string `json:"vpc_id"`

	// Network configuration for Virtual IP per AZ. Field introduced in 18.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Zones []*TencentZoneNetwork `json:"zones,omitempty"`
}
