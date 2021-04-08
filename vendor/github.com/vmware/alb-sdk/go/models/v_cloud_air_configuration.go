package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// VCloudAirConfiguration v cloud air configuration
// swagger:model vCloudAirConfiguration
type VCloudAirConfiguration struct {

	// vCloudAir access mode. Enum options - NO_ACCESS, READ_ACCESS, WRITE_ACCESS.
	// Required: true
	Privilege *string `json:"privilege"`

	// vCloudAir host address.
	// Required: true
	VcaHost *string `json:"vca_host"`

	// vCloudAir instance ID.
	// Required: true
	VcaInstance *string `json:"vca_instance"`

	// vCloudAir management network.
	// Required: true
	VcaMgmtNetwork *string `json:"vca_mgmt_network"`

	// vCloudAir orgnization ID.
	// Required: true
	VcaOrgnization *string `json:"vca_orgnization"`

	// vCloudAir password.
	// Required: true
	VcaPassword *string `json:"vca_password"`

	// vCloudAir username.
	// Required: true
	VcaUsername *string `json:"vca_username"`

	// vCloudAir virtual data center name.
	// Required: true
	VcaVdc *string `json:"vca_vdc"`
}
