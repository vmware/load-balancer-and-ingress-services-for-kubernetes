package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SeLicensedBandwdithExceededEventDetails se licensed bandwdith exceeded event details
// swagger:model SeLicensedBandwdithExceededEventDetails
type SeLicensedBandwdithExceededEventDetails struct {

	// Number of packets dropped since the last event.
	NumPktsDropped *int32 `json:"num_pkts_dropped,omitempty"`

	// UUID of the SE responsible for this event. It is a reference to an object of type ServiceEngine.
	SeRef *string `json:"se_ref,omitempty"`
}
