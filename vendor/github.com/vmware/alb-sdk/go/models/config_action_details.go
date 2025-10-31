// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ConfigActionDetails config action details
// swagger:model ConfigActionDetails
type ConfigActionDetails struct {

	// Name of the action. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ActionName *string `json:"action_name,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ClientIP *string `json:"client_ip,omitempty"`

	// Error message if request failed. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ErrorMessage *string `json:"error_message,omitempty"`

	// Parameter data. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ParameterData *string `json:"parameter_data,omitempty"`

	// API path. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Path *string `json:"path,omitempty"`

	// Name of the resource. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ResourceName *string `json:"resource_name,omitempty"`

	// Config type of the resource. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ResourceType *string `json:"resource_type,omitempty"`

	// Status. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Status *string `json:"status,omitempty"`

	// Request user. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	User *string `json:"user,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UserAgent *string `json:"user_agent,omitempty"`
}
