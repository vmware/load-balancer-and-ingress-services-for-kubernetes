// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// AlertTestEmailParams alert test email params
// swagger:model AlertTestEmailParams
type AlertTestEmailParams struct {

	// The Subject line of the originating email from  Avi Controller.
	// Required: true
	Subject *string `json:"subject"`

	// The email context.
	// Required: true
	Text *string `json:"text"`

	// Unique object identifier of the object.
	UUID *string `json:"uuid,omitempty"`
}
