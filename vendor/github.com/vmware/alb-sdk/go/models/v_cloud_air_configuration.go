// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// VCloudAirConfiguration v cloud air configuration
// swagger:model vCloudAirConfiguration
type VCloudAirConfiguration struct {

	// vCloudAir access mode. Enum options - NO_ACCESS, READ_ACCESS, WRITE_ACCESS. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Privilege *string `json:"privilege"`

	// vCloudAir host address. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	VcaHost *string `json:"vca_host"`

	// vCloudAir instance ID. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	VcaInstance *string `json:"vca_instance"`

	// vCloudAir management network. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	VcaMgmtNetwork *string `json:"vca_mgmt_network"`

	// vCloudAir orgnization ID. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	VcaOrgnization *string `json:"vca_orgnization"`

	// vCloudAir password. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	VcaPassword *string `json:"vca_password"`

	// vCloudAir username. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	VcaUsername *string `json:"vca_username"`

	// vCloudAir virtual data center name. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	VcaVdc *string `json:"vca_vdc"`
}
