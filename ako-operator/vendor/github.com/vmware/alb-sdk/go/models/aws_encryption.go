// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// AwsEncryption aws encryption
// swagger:model AwsEncryption
type AwsEncryption struct {

	// AWS KMS ARN ID of the master key for encryption. Field introduced in 17.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MasterKey *string `json:"master_key,omitempty"`

	// AWS encryption mode. Enum options - AWS_ENCRYPTION_MODE_NONE, AWS_ENCRYPTION_MODE_SSE_KMS. Field introduced in 17.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Mode *string `json:"mode,omitempty"`
}
