package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ProxyConfiguration proxy configuration
// swagger:model ProxyConfiguration
type ProxyConfiguration struct {

	// Proxy hostname or IP address.
	// Required: true
	Host *string `json:"host"`

	// Password for proxy.
	Password *string `json:"password,omitempty"`

	// Proxy port.
	// Required: true
	Port *int32 `json:"port"`

	// Username for proxy.
	Username *string `json:"username,omitempty"`
}
