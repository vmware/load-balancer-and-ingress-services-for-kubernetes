// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// PKIprofileDetails p k iprofile details
// swagger:model PKIProfileDetails
type PKIprofileDetails struct {

	// CRL list. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Crls *string `json:"crls,omitempty"`

	// Name of PKIProfile. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Name *string `json:"name,omitempty"`
}
