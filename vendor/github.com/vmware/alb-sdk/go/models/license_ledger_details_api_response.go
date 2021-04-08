package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

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
