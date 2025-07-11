// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// CloudStackConfiguration cloud stack configuration
// swagger:model CloudStackConfiguration
type CloudStackConfiguration struct {

	// CloudStack API Key. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	AccessKeyID *string `json:"access_key_id"`

	// CloudStack API URL. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	APIURL *string `json:"api_url"`

	// If controller's management IP is in a private network, a publicly accessible IP to reach the controller. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CntrPublicIP *string `json:"cntr_public_ip,omitempty"`

	// Default hypervisor type. Enum options - DEFAULT, VMWARE_ESX, KVM, VMWARE_VSAN, XEN. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Hypervisor *string `json:"hypervisor,omitempty"`

	// Avi Management network name. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	MgmtNetworkName *string `json:"mgmt_network_name"`

	// Avi Management network name. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MgmtNetworkUUID *string `json:"mgmt_network_uuid,omitempty"`

	// CloudStack Secret Key. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	SecretAccessKey *string `json:"secret_access_key"`
}
