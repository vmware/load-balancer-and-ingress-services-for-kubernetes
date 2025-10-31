// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// HardwareSecurityModule hardware security module
// swagger:model HardwareSecurityModule
type HardwareSecurityModule struct {

	// AWS CloudHSM specific configuration. Field introduced in 17.2.7. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Cloudhsm *HSMAwsCloudHsm `json:"cloudhsm,omitempty"`

	// Thales netHSM specific configuration. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Nethsm []*HSMThalesNetHsm `json:"nethsm,omitempty"`

	// Thales Remote File Server (RFS), used for the netHSMs, configuration. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Rfs *HSMThalesRFS `json:"rfs,omitempty"`

	// Thales Luna HSM/Gem specific configuration. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Sluna *HSMSafenetLuna `json:"sluna,omitempty"`

	// HSM type to use. Enum options - HSM_TYPE_THALES_NETHSM, HSM_TYPE_SAFENET_LUNA, HSM_TYPE_AWS_CLOUDHSM. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Type *string `json:"type"`
}
