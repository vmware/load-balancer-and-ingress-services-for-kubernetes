// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// CdpLldpInfo cdp lldp info
// swagger:model CdpLldpInfo
type CdpLldpInfo struct {

	// chassis of CdpLldpInfo.
	Chassis *string `json:"chassis,omitempty"`

	// device of CdpLldpInfo.
	Device *string `json:"device,omitempty"`

	// mgmtaddr of CdpLldpInfo.
	Mgmtaddr *string `json:"mgmtaddr,omitempty"`

	// port of CdpLldpInfo.
	Port *string `json:"port,omitempty"`

	//  Enum options - CDP, LLDP, NOT_APPLICABLE.
	SwitchInfoType *string `json:"switch_info_type,omitempty"`

	// system_name of CdpLldpInfo.
	SystemName *string `json:"system_name,omitempty"`
}
