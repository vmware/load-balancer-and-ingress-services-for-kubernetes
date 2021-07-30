// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SysTestEmailParams sys test email params
// swagger:model SysTestEmailParams
type SysTestEmailParams struct {

	// Alerts are copied to the comma separated list of  email recipients.
	CcEmails *string `json:"cc_emails,omitempty"`

	// The Subject line of the originating email from  Avi Controller.
	// Required: true
	Subject *string `json:"subject"`

	// The email context.
	// Required: true
	Text *string `json:"text"`

	// Alerts are sent to the comma separated list of  email recipients.
	// Required: true
	ToEmails *string `json:"to_emails"`
}
