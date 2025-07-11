// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// PackageDetails package details
// swagger:model PackageDetails
type PackageDetails struct {

	// This contains build related information. Field introduced in 18.2.6. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Build *BuildInfo `json:"build,omitempty"`

	// MD5 checksum over the entire package. Field introduced in 18.2.6. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Hash *string `json:"hash,omitempty"`

	// Patch related necessary information. Field introduced in 18.2.6. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Patch *PatchInfo `json:"patch,omitempty"`

	// Path of the package in the repository. Field introduced in 18.2.6. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Path *string `json:"path,omitempty"`
}
