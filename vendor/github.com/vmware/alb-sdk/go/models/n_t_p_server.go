package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// NTPServer n t p server
// swagger:model NTPServer
type NTPServer struct {

	// Key number from the list of trusted keys used to authenticate this server. Allowed values are 1-65534.
	KeyNumber *int32 `json:"key_number,omitempty"`

	// IP Address of the NTP Server.
	// Required: true
	Server *IPAddr `json:"server"`
}
