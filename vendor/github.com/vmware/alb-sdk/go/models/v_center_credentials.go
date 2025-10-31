// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// VCenterCredentials v center credentials
// swagger:model VCenterCredentials
type VCenterCredentials struct {

	// Password to talk to VCenter server. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Password *string `json:"password,omitempty"`

	// Username to talk to VCenter server. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Username *string `json:"username,omitempty"`
}
