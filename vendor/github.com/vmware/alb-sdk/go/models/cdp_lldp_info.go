// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// CdpLldpInfo cdp lldp info
// swagger:model CdpLldpInfo
type CdpLldpInfo struct {

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Chassis *string `json:"chassis,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Device *string `json:"device,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Mgmtaddr *string `json:"mgmtaddr,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Port *string `json:"port,omitempty"`

	//  Enum options - CDP, LLDP, NOT_APPLICABLE. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SwitchInfoType *string `json:"switch_info_type,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SystemName *string `json:"system_name,omitempty"`
}
