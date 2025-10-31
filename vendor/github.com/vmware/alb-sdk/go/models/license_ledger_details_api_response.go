// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// LicenseLedgerDetailsAPIResponse license ledger details Api response
// swagger:model LicenseLedgerDetailsApiResponse
type LicenseLedgerDetailsAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*LicenseLedgerDetails `json:"results,omitempty"`
}
