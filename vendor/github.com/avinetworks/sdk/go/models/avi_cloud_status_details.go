package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// AviCloudStatusDetails avi cloud status details
// swagger:model AviCloudStatusDetails
type AviCloudStatusDetails struct {

	// Connection status of the controller cluster to Avi Cloud. Enum options - AVICLOUD_CONNECTIVITY_UNKNOWN, AVICLOUD_DISCONNECTED, AVICLOUD_CONNECTED. Field introduced in 18.2.6.
	Connectivity *string `json:"connectivity,omitempty"`

	// Status change reason. Field introduced in 18.2.6.
	Reason *string `json:"reason,omitempty"`

	// Registration status of the controller cluster to Avi Cloud. Enum options - AVICLOUD_REGISTRATION_UNKNOWN, AVICLOUD_REGISTERED, AVICLOUD_DEREGISTERED. Field introduced in 18.2.6.
	Registration *string `json:"registration,omitempty"`
}
