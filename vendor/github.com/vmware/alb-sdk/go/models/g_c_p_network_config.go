// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// GCPNetworkConfig g c p network config
// swagger:model GCPNetworkConfig
type GCPNetworkConfig struct {

	// Config Mode for Google Cloud network configuration. Enum options - INBAND_MANAGEMENT, ONE_ARM_MODE, TWO_ARM_MODE. Field introduced in 18.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Config *string `json:"config"`

	// Configure InBand Management as Google Cloud network configuration. In this configuration the data network and management network for Service Engines will be same. Field introduced in 18.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Inband *GCPInBandManagement `json:"inband,omitempty"`

	// Configure One Arm Mode as Google Cloud network configuration. In this configuration the data network and the management network for the Service Engines will be separated. Field introduced in 18.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	OneArm *GCPOneArmMode `json:"one_arm,omitempty"`

	// Configure Two Arm Mode as Google Cloud network configuration. In this configuration the frontend data network, backend data network and the management network for the Service Engines will be separated. Field introduced in 18.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	TwoArm *GCPTwoArmMode `json:"two_arm,omitempty"`
}
