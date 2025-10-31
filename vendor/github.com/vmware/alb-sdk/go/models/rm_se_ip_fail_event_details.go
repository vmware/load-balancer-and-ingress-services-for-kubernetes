// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// RmSeIPFailEventDetails rm se Ip fail event details
// swagger:model RmSeIpFailEventDetails
type RmSeIPFailEventDetails struct {

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	HostName *string `json:"host_name,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Networks []*RmAddVnic `json:"networks,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Reason *string `json:"reason,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeName *string `json:"se_name,omitempty"`
}
