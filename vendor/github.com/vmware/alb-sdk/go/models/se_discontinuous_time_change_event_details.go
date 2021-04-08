package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SeDiscontinuousTimeChangeEventDetails se discontinuous time change event details
// swagger:model SeDiscontinuousTimeChangeEventDetails
type SeDiscontinuousTimeChangeEventDetails struct {

	// Relative time drift between SE and controller in terms of microseconds.
	DriftTime *int64 `json:"drift_time,omitempty"`

	// Time stamp before the discontinuous jump in time.
	FromTime *string `json:"from_time,omitempty"`

	// Name of the SE responsible for this event. It is a reference to an object of type ServiceEngine.
	SeName *string `json:"se_name,omitempty"`

	// UUID of the SE responsible for this event. It is a reference to an object of type ServiceEngine.
	SeRef *string `json:"se_ref,omitempty"`

	// Time stamp to which the time has discontinuously jumped.
	ToTime *string `json:"to_time,omitempty"`
}
