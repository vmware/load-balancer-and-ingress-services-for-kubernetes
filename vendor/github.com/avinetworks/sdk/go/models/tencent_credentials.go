package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// TencentCredentials tencent credentials
// swagger:model TencentCredentials
type TencentCredentials struct {

	// Tencent secret ID. Field introduced in 18.2.3.
	// Required: true
	SecretID *string `json:"secret_id"`

	// Tencent secret key. Field introduced in 18.2.3.
	// Required: true
	SecretKey *string `json:"secret_key"`
}
