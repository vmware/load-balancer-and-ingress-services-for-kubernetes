// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// VsScaleinParams vs scalein params
// swagger:model VsScaleinParams
type VsScaleinParams struct {

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AdminDown *bool `json:"admin_down,omitempty"`

	//  It is a reference to an object of type ServiceEngine. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	FromSeRef *string `json:"from_se_ref,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ScaleinPrimary *bool `json:"scalein_primary,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UUID *string `json:"uuid,omitempty"`

	//  Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	VipID *string `json:"vip_id"`
}
