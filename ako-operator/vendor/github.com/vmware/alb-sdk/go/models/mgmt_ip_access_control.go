// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// MgmtIPAccessControl mgmt Ip access control
// swagger:model MgmtIpAccessControl
type MgmtIPAccessControl struct {

	// Configure IP addresses to access controller using API. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	APIAccess *IPAddrMatch `json:"api_access,omitempty"`

	// Configure IP addresses to access controller using CLI Shell. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ShellServerAccess *IPAddrMatch `json:"shell_server_access,omitempty"`

	// Configure IP addresses to access controller using SNMP. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SnmpAccess *IPAddrMatch `json:"snmp_access,omitempty"`

	// Configure IP addresses to access controller using SSH. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SSHAccess *IPAddrMatch `json:"ssh_access,omitempty"`

	// Configure IP addresses to access controller using sysint access. Field introduced in 18.1.3, 18.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SysintAccess *IPAddrMatch `json:"sysint_access,omitempty"`
}
