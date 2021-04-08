package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// HardwareSecurityModule hardware security module
// swagger:model HardwareSecurityModule
type HardwareSecurityModule struct {

	// AWS CloudHSM specific configuration. Field introduced in 17.2.7.
	Cloudhsm *HSMAwsCloudHsm `json:"cloudhsm,omitempty"`

	// Thales netHSM specific configuration.
	Nethsm []*HSMThalesNetHsm `json:"nethsm,omitempty"`

	// Thales Remote File Server (RFS), used for the netHSMs, configuration.
	Rfs *HSMThalesRFS `json:"rfs,omitempty"`

	// Safenet/Gemalto Luna/Gem specific configuration.
	Sluna *HSMSafenetLuna `json:"sluna,omitempty"`

	// HSM type to use. Enum options - HSM_TYPE_THALES_NETHSM, HSM_TYPE_SAFENET_LUNA, HSM_TYPE_AWS_CLOUDHSM.
	// Required: true
	Type *string `json:"type"`
}
