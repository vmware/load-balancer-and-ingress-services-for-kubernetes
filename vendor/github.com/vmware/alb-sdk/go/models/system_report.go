// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SystemReport system report
// swagger:model SystemReport
type SystemReport struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// Relative path to the report archive file on filesystem.The archive includes exported system configuration and current object as json. Field introduced in 22.1.6, 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ArchiveRef *string `json:"archive_ref,omitempty"`

	// Controller Patch Image associated with the report. It is a reference to an object of type Image. Field introduced in 22.1.6, 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ControllerPatchImageRef *string `json:"controller_patch_image_ref,omitempty"`

	// Indicates whether this report is downloadable as an archive. Field introduced in 22.1.6, 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Downloadable *bool `json:"downloadable,omitempty"`

	// List of events associated with the report. Field introduced in 22.1.6, 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Events []*ReportEvent `json:"events,omitempty"`

	// System Image associated with the report. It is a reference to an object of type Image. Field introduced in 22.1.6, 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ImageRef *string `json:"image_ref,omitempty"`

	// Name of the report derived from operation in a readable format. Ex  upgrade_system_1a5c. Field introduced in 22.1.6, 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Name *string `json:"name,omitempty"`

	// Readiness state of the system. Ex  Upgrade Pre-check Results. Field introduced in 22.1.6, 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ReadinessReports []*ReportDetail `json:"readiness_reports,omitempty"`

	// SE Patch Image associated with the report. It is a reference to an object of type Image. Field introduced in 22.1.6, 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SePatchImageRef *string `json:"se_patch_image_ref,omitempty"`

	// Report state combines all applicable states. Ex  readiness_reports.system_readiness.state. Field introduced in 22.1.6, 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	State *ReportOpsState `json:"state,omitempty"`

	// Summary of the report. Field introduced in 22.1.6, 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Summary *ReportSummary `json:"summary,omitempty"`

	// List of tasks associated with the report. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Tasks []*ReportTask `json:"tasks,omitempty"`

	// Tenant UUID associated with the Object. It is a reference to an object of type Tenant. Field introduced in 22.1.6, 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// UUID Identifier for the report. Field introduced in 22.1.6, 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	UUID *string `json:"uuid,omitempty"`
}
