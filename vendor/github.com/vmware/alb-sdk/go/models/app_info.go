package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// AppInfo app info
// swagger:model AppInfo
type AppInfo struct {

	// app_hdr_name of AppInfo.
	// Required: true
	AppHdrName *string `json:"app_hdr_name"`

	// app_hdr_value of AppInfo.
	// Required: true
	AppHdrValue *string `json:"app_hdr_value"`
}
