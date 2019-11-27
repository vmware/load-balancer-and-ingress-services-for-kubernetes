package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SnmpTrapProfileAPIResponse snmp trap profile Api response
// swagger:model SnmpTrapProfileApiResponse
type SnmpTrapProfileAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// results
	// Required: true
	Results []*SnmpTrapProfile `json:"results,omitempty"`
}
