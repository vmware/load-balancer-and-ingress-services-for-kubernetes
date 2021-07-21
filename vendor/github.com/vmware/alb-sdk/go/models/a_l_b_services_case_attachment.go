// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ALBServicesCaseAttachment a l b services case attachment
// swagger:model ALBServicesCaseAttachment
type ALBServicesCaseAttachment struct {

	//  Field introduced in 18.2.6.
	// Required: true
	AttachmentName *string `json:"attachment_name"`

	//  Field introduced in 18.2.6.
	// Required: true
	AttachmentSize *string `json:"attachment_size"`

	//  Field introduced in 18.2.6.
	// Required: true
	AttachmentURL *string `json:"attachment_url"`
}
