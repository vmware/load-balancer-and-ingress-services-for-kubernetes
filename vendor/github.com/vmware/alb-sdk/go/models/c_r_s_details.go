// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// CRSDetails c r s details
// swagger:model CRSDetails
type CRSDetails struct {

	// Name of the CRS release. Field introduced in 18.2.6. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Name *string `json:"name,omitempty"`

	// CRS release date. Field introduced in 18.2.6. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ReleaseDate *string `json:"release_date,omitempty"`

	// Download link of the CRS release. Field introduced in 18.2.6. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	URL *string `json:"url,omitempty"`

	// Version of the CRS release. Field introduced in 18.2.6. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Version *string `json:"version,omitempty"`
}
