// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// Image image
// swagger:model Image
type Image struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// This field describes the cloud info specific to the base image. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CloudInfoValues []*ImageCloudData `json:"cloud_info_values,omitempty"`

	// Controller package details. Field introduced in 18.2.6. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ControllerInfo *PackageDetails `json:"controller_info,omitempty"`

	// Mandatory Controller patch name that is applied along with this base image. Field introduced in 18.2.10, 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ControllerPatchName *string `json:"controller_patch_name,omitempty"`

	// It references the controller-patch associated with the Uber image. It is a reference to an object of type Image. Field introduced in 18.2.8, 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ControllerPatchRef *string `json:"controller_patch_ref,omitempty"`

	// Time taken to upload the image in seconds. Field introduced in 21.1.3. Unit is SEC. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Duration uint32 `json:"duration,omitempty"`

	// Image upload end time. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	EndTime *string `json:"end_time,omitempty"`

	// Image events for image upload operation. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Events []*ImageEventMap `json:"events,omitempty"`

	// Specifies whether FIPS mode can be enabled on this image. Field introduced in 30.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	FipsModeTransitionApplicable *bool `json:"fips_mode_transition_applicable,omitempty"`

	// Status of the image. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ImgState *ImageUploadOpsStatus `json:"img_state,omitempty"`

	// This field describes the api migration related information. Field introduced in 18.2.6. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Migrations *SupportedMigrations `json:"migrations,omitempty"`

	// Name of the image. Field introduced in 18.2.6. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Name *string `json:"name"`

	// Image upload progress which holds value between 0-100. Allowed values are 0-100. Field introduced in 21.1.3. Unit is PERCENT. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Progress uint32 `json:"progress,omitempty"`

	// SE package details. Field introduced in 18.2.6. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeInfo *PackageDetails `json:"se_info,omitempty"`

	// Mandatory ServiceEngine patch name that is applied along with this base image. Field introduced in 18.2.10, 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SePatchName *string `json:"se_patch_name,omitempty"`

	// It references the Service Engine patch associated with the Uber Image. It is a reference to an object of type Image. Field introduced in 18.2.8, 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SePatchRef *string `json:"se_patch_ref,omitempty"`

	// Image upload start time. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	StartTime *string `json:"start_time,omitempty"`

	// Completed set of tasks for Image upload. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	TasksCompleted *int32 `json:"tasks_completed,omitempty"`

	// Tenant that this object belongs to. It is a reference to an object of type Tenant. Field introduced in 18.2.6. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// Total number of tasks for Image upload. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	TotalTasks *int32 `json:"total_tasks,omitempty"`

	// Type of the image patch/system. Enum options - IMAGE_TYPE_PATCH, IMAGE_TYPE_SYSTEM, IMAGE_TYPE_MUST_CHECK. Field introduced in 18.2.6. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Type *string `json:"type,omitempty"`

	// Status to check if the image is an uber bundle. Field introduced in 18.2.8, 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UberBundle *bool `json:"uber_bundle,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// UUID of the image. Field introduced in 18.2.6. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UUID *string `json:"uuid,omitempty"`
}
