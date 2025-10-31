// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// CertificateManagementProfileAPIResponse certificate management profile Api response
// swagger:model CertificateManagementProfileApiResponse
type CertificateManagementProfileAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*CertificateManagementProfile `json:"results,omitempty"`
}
