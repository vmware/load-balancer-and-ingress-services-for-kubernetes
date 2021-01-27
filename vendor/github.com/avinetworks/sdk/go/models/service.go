package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// Service service
// swagger:model Service
type Service struct {

	// Enable HTTP2 on this port. Field introduced in 20.1.1. Allowed in Basic(Allowed values- false) edition, Essentials(Allowed values- false) edition, Enterprise edition.
	EnableHttp2 *bool `json:"enable_http2,omitempty"`

	// Enable SSL termination and offload for traffic from clients.
	EnableSsl *bool `json:"enable_ssl,omitempty"`

	// Enable application layer specific features for the this specific service. It is a reference to an object of type ApplicationProfile. Field introduced in 17.2.4. Allowed in Basic edition, Essentials edition, Enterprise edition.
	OverrideApplicationProfileRef *string `json:"override_application_profile_ref,omitempty"`

	// Override the network profile for this specific service port. It is a reference to an object of type NetworkProfile.
	OverrideNetworkProfileRef *string `json:"override_network_profile_ref,omitempty"`

	// The Virtual Service's port number. Allowed values are 0-65535.
	// Required: true
	Port *int32 `json:"port"`

	// The end of the Virtual Service's port number range. Allowed values are 1-65535. Special values are 0- 'single port'.
	PortRangeEnd *int32 `json:"port_range_end,omitempty"`
}
