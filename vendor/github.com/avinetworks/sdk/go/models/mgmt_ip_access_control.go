package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// MgmtIPAccessControl mgmt Ip access control
// swagger:model MgmtIpAccessControl
type MgmtIPAccessControl struct {

	// Configure IP addresses to access controller using API.
	APIAccess *IPAddrMatch `json:"api_access,omitempty"`

	// Configure IP addresses to access controller using CLI Shell.
	ShellServerAccess *IPAddrMatch `json:"shell_server_access,omitempty"`

	// Configure IP addresses to access controller using SNMP.
	SnmpAccess *IPAddrMatch `json:"snmp_access,omitempty"`

	// Configure IP addresses to access controller using SSH.
	SSHAccess *IPAddrMatch `json:"ssh_access,omitempty"`

	// Configure IP addresses to access controller using sysint access. Field introduced in 18.1.3, 18.2.1.
	SysintAccess *IPAddrMatch `json:"sysint_access,omitempty"`
}
