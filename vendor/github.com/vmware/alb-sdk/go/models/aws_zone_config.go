// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// AwsZoneConfig aws zone config
// swagger:model AwsZoneConfig
type AwsZoneConfig struct {

	// Availability zone.
	// Required: true
	AvailabilityZone *string `json:"availability_zone"`

	// Name or CIDR of the network in the Availability Zone that will be used as management network.
	// Required: true
	MgmtNetworkName *string `json:"mgmt_network_name"`

	// UUID of the network in the Availability Zone that will be used as management network.
	MgmtNetworkUUID *string `json:"mgmt_network_uuid,omitempty"`
}
