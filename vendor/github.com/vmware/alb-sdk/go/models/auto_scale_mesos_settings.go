package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// AutoScaleMesosSettings auto scale mesos settings
// swagger:model AutoScaleMesosSettings
type AutoScaleMesosSettings struct {

	// Apply scaleout even when there are deployments inprogress.
	Force *bool `json:"force,omitempty"`
}
