// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// TaskJournal task journal
// swagger:model TaskJournal
type TaskJournal struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// List of errors in the process. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Errors []*JournalError `json:"errors,omitempty"`

	// Image uuid for identifying the current base image. It is a reference to an object of type Image. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ImageRef *string `json:"image_ref,omitempty"`

	// Detailed Information of Journal. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Info *JournalInfo `json:"info,omitempty"`

	// Name for the task journal. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Name *string `json:"name,omitempty"`

	// Cloud that this object belongs to. It is a reference to an object of type Cloud. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ObjCloudRef *string `json:"obj_cloud_ref,omitempty"`

	// Operation for which the task journal created. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Operation *string `json:"operation,omitempty"`

	// Image uuid for identifying the current patch. It is a reference to an object of type Image. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	PatchImageRef *string `json:"patch_image_ref,omitempty"`

	// Summary of Journal. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	Summary *JournalSummary `json:"summary"`

	// Tenant UUID associated with the Object. It is a reference to an object of type Tenant. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// UUID Identifier for the task journal. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	UUID *string `json:"uuid,omitempty"`
}
