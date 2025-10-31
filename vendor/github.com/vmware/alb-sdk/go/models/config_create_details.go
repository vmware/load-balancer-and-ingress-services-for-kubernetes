// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ConfigCreateDetails config create details
// swagger:model ConfigCreateDetails
type ConfigCreateDetails struct {

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ClientIP *string `json:"client_ip,omitempty"`

	// Error message if request failed. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ErrorMessage *string `json:"error_message,omitempty"`

	// API path. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Path *string `json:"path,omitempty"`

	// Request data if request failed. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	RequestData *string `json:"request_data,omitempty"`

	// Data of the created resource. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ResourceData *string `json:"resource_data,omitempty"`

	// Name of the created resource. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ResourceName *string `json:"resource_name,omitempty"`

	// Config type of the created resource. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ResourceType *string `json:"resource_type,omitempty"`

	// Status. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Status *string `json:"status,omitempty"`

	// Request user. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	User *string `json:"user,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UserAgent *string `json:"user_agent,omitempty"`
}
