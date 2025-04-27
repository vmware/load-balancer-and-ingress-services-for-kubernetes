// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SSLKeyAndCertificateAPIResponse s s l key and certificate Api response
// swagger:model SSLKeyAndCertificateApiResponse
type SSLKeyAndCertificateAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*SSLKeyAndCertificate `json:"results,omitempty"`
}
