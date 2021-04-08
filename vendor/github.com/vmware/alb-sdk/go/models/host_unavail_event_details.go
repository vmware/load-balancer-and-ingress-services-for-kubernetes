package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// HostUnavailEventDetails host unavail event details
// swagger:model HostUnavailEventDetails
type HostUnavailEventDetails struct {

	// Name of the object.
	Name *string `json:"name,omitempty"`

	// reasons of HostUnavailEventDetails.
	Reasons []string `json:"reasons,omitempty"`

	// vs_name of HostUnavailEventDetails.
	VsName *string `json:"vs_name,omitempty"`
}
