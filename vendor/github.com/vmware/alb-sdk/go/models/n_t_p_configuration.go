package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

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
