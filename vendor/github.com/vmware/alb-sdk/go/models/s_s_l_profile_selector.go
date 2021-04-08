package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SSLProfileSelector s s l profile selector
// swagger:model SSLProfileSelector
type SSLProfileSelector struct {

	// Configure client IP address groups. Field introduced in 18.2.3.
	// Required: true
	ClientIPList *IPAddrMatch `json:"client_ip_list"`

	// SSL profile for the client IP addresses listed. It is a reference to an object of type SSLProfile. Field introduced in 18.2.3.
	// Required: true
	SslProfileRef *string `json:"ssl_profile_ref"`
}
