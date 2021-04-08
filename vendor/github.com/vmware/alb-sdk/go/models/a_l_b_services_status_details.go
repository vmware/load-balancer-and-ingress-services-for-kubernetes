package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ALBServicesStatusDetails a l b services status details
// swagger:model ALBServicesStatusDetails
type ALBServicesStatusDetails struct {

	// Connection status of the controller cluster to ALBServices. Enum options - ALBSERVICES_CONNECTIVITY_UNKNOWN, ALBSERVICES_DISCONNECTED, ALBSERVICES_CONNECTED. Field introduced in 18.2.6.
	Connectivity *string `json:"connectivity,omitempty"`

	// Status change reason. Field introduced in 18.2.6.
	Reason *string `json:"reason,omitempty"`

	// Registration status of the controller cluster to ALBServices. Enum options - ALBSERVICES_REGISTRATION_UNKNOWN, ALBSERVICES_REGISTERED, ALBSERVICES_DEREGISTERED. Field introduced in 18.2.6.
	Registration *string `json:"registration,omitempty"`
}
