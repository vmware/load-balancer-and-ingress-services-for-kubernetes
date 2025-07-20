// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SamlMetadataUpdateFailedDetails saml metadata update failed details
// swagger:model SamlMetadataUpdateFailedDetails
type SamlMetadataUpdateFailedDetails struct {

	// Name of Auth Profile. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Name *string `json:"name,omitempty"`

	// Reason for Update Failure. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Reason *string `json:"reason,omitempty"`
}
