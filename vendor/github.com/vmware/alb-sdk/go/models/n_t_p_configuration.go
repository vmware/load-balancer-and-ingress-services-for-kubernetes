// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// NTPConfiguration n t p configuration
// swagger:model NTPConfiguration
type NTPConfiguration struct {

	// NTP Authentication keys.
	NtpAuthenticationKeys []*NTPAuthenticationKey `json:"ntp_authentication_keys,omitempty"`

	// List of NTP server hostnames or IP addresses.
	NtpServerList []*IPAddr `json:"ntp_server_list,omitempty"`

	// List of NTP Servers.
	NtpServers []*NTPServer `json:"ntp_servers,omitempty"`
}
