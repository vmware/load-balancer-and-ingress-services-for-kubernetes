// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// AppSignatureEventData app signature event data
// swagger:model AppSignatureEventData
type AppSignatureEventData struct {

	// Last Successful updated time of the AppSignature. Field introduced in 20.1.4.
	LastSuccessfulUpdatedTime *string `json:"last_successful_updated_time,omitempty"`

	// Reason for AppSignature transaction failure. Field introduced in 20.1.4.
	Reason *string `json:"reason,omitempty"`

	// Status of AppSignature transaction. Field introduced in 20.1.4.
	Status *string `json:"status,omitempty"`
}
