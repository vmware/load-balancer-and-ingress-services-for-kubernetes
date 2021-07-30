// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// AWSLogin a w s login
// swagger:model AWSLogin
type AWSLogin struct {

	// access_key_id of AWSLogin.
	AccessKeyID *string `json:"access_key_id,omitempty"`

	// iam_assume_role of AWSLogin.
	IamAssumeRole *string `json:"iam_assume_role,omitempty"`

	// AWS region.
	Region *string `json:"region,omitempty"`

	// secret_access_key of AWSLogin.
	SecretAccessKey *string `json:"secret_access_key,omitempty"`

	// Placeholder for description of property use_iam_roles of obj type AWSLogin field type str  type boolean
	UseIamRoles *bool `json:"use_iam_roles,omitempty"`

	// vpc_id of AWSLogin.
	VpcID *string `json:"vpc_id,omitempty"`
}
