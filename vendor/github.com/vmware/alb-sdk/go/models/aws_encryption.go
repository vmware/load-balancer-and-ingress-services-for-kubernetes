// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// AwsEncryption aws encryption
// swagger:model AwsEncryption
type AwsEncryption struct {

	// AWS KMS ARN ID of the master key for encryption. Field introduced in 17.2.3.
	MasterKey *string `json:"master_key,omitempty"`

	// AWS encryption mode. Enum options - AWS_ENCRYPTION_MODE_NONE, AWS_ENCRYPTION_MODE_SSE_KMS. Field introduced in 17.2.3.
	Mode *string `json:"mode,omitempty"`
}
