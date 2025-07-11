// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ALBServicesCaseAttachment a l b services case attachment
// swagger:model ALBServicesCaseAttachment
type ALBServicesCaseAttachment struct {

	//  Field introduced in 18.2.6. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	AttachmentName *string `json:"attachment_name"`

	//  Field introduced in 18.2.6. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	AttachmentSize *string `json:"attachment_size"`

	//  Field introduced in 18.2.6. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	AttachmentURL *string `json:"attachment_url"`
}
