// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// AWSSetup a w s setup
// swagger:model AWSSetup
type AWSSetup struct {

	// access_key_id of AWSSetup.
	AccessKeyID *string `json:"access_key_id,omitempty"`

	// cc_id of AWSSetup.
	CcID *string `json:"cc_id,omitempty"`

	// error_string of AWSSetup.
	ErrorString *string `json:"error_string,omitempty"`

	//  Enum options - NO_ACCESS, READ_ACCESS, WRITE_ACCESS.
	Privilege *string `json:"privilege,omitempty"`

	// region of AWSSetup.
	// Required: true
	Region *string `json:"region"`

	//  Field introduced in 17.1.3.
	VpcID *string `json:"vpc_id,omitempty"`
}
