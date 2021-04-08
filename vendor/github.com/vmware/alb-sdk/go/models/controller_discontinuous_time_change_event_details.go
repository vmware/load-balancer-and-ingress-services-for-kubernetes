package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ControllerDiscontinuousTimeChangeEventDetails controller discontinuous time change event details
// swagger:model ControllerDiscontinuousTimeChangeEventDetails
type ControllerDiscontinuousTimeChangeEventDetails struct {

	// Time stamp before the discontinuous jump in time.
	FromTime *string `json:"from_time,omitempty"`

	// Name of the Controller responsible for this event.
	NodeName *string `json:"node_name,omitempty"`

	// Time stamp to which the time has discontinuously jumped.
	ToTime *string `json:"to_time,omitempty"`
}
