// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// VipScaleDetails vip scale details
// swagger:model VipScaleDetails
type VipScaleDetails struct {

	// availability_zone of VipScaleDetails.
	AvailabilityZone *string `json:"availability_zone,omitempty"`

	// error of VipScaleDetails.
	Error *string `json:"error,omitempty"`

	// Unique object identifier of subnet.
	SubnetUUID *string `json:"subnet_uuid,omitempty"`

	// vip_id of VipScaleDetails.
	VipID *string `json:"vip_id,omitempty"`

	// Unique object identifier of vsvip.
	VsvipUUID *string `json:"vsvip_uuid,omitempty"`
}
