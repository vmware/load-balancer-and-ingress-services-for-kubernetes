// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ALBServicesFileDownloadAPIResponse a l b services file download Api response
// swagger:model ALBServicesFileDownloadApiResponse
type ALBServicesFileDownloadAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*ALBServicesFileDownload `json:"results,omitempty"`
}
