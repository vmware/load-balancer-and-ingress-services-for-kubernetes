// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SSLExportDetails s s l export details
// swagger:model SSLExportDetails
type SSLExportDetails struct {

	// Request user. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	User *string `json:"user,omitempty"`
}
