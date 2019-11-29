package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// NsxConfiguration nsx configuration
// swagger:model NsxConfiguration
type NsxConfiguration struct {

	// This prefix will be added to the names of all NSX objects created by Avi Controller. It should be unique across all the Avi Controller clusters. Field introduced in 17.1.1.
	// Required: true
	AviNsxPrefix *string `json:"avi_nsx_prefix"`

	// The hostname or IP address of the NSX MGr. Field introduced in 17.1.1.
	// Required: true
	NsxManagerName *string `json:"nsx_manager_name"`

	// The password Avi Vantage will use when authenticating with NSX Mgr. Field introduced in 17.1.1.
	// Required: true
	NsxManagerPassword *string `json:"nsx_manager_password"`

	// The username Avi Vantage will use when authenticating with NSX Mgr. Field introduced in 17.1.1.
	// Required: true
	NsxManagerUsername *string `json:"nsx_manager_username"`

	// The interval (in secs) with which Avi Controller polls the NSX Manager for updates. Field introduced in 17.1.1.
	// Required: true
	NsxPollTime *int32 `json:"nsx_poll_time"`
}
