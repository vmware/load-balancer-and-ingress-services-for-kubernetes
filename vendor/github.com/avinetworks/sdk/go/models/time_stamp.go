package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// TimeStamp time stamp
// swagger:model TimeStamp
type TimeStamp struct {

	// Number of secs.
	// Required: true
	Secs *int64 `json:"secs"`

	// Number of usecs.
	// Required: true
	Usecs *int64 `json:"usecs"`
}
