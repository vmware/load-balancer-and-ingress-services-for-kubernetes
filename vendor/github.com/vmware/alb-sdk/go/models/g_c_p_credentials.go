// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// GCPCredentials g c p credentials
// swagger:model GCPCredentials
type GCPCredentials struct {

	// Google Cloud Platform Service Account keyfile data in JSON format. Field introduced in 18.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ServiceAccountKeyfileData *string `json:"service_account_keyfile_data,omitempty"`
}
