// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// AsyncPatchState async patch state
// swagger:model AsyncPatchState
type AsyncPatchState struct {

	// Error message if request failed. Field introduced in 22.1.6,30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ErrorMessage *string `json:"error_message,omitempty"`

	// Error status code if request failed. Field introduced in 22.1.6,30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ErrorStatusCode *uint32 `json:"error_status_code,omitempty"`

	// Merged patch id. Field introduced in 22.1.6,30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	MergedPatchID *uint64 `json:"merged_patch_id,omitempty"`

	// List of patch ids. Field introduced in 22.1.6,30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	PatchIds *string `json:"patch_ids,omitempty"`

	// API path. Field introduced in 22.1.6,30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Path *string `json:"path,omitempty"`

	// Request data. Field introduced in 22.1.6,30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	RequestData *string `json:"request_data,omitempty"`

	// Async Patch Queue data for which status is updated. Field introduced in 22.1.6,30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ResourceData *string `json:"resource_data,omitempty"`

	// Name of the resource. Field introduced in 22.1.6,30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ResourceName *string `json:"resource_name,omitempty"`

	// Config type of the resource. Field introduced in 22.1.6,30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ResourceType *string `json:"resource_type,omitempty"`

	// Status of Async Patch. Field introduced in 22.1.6,30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Status *string `json:"status,omitempty"`

	// Request user. Field introduced in 22.1.6,30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	User *string `json:"user,omitempty"`
}
