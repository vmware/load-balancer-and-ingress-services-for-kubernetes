package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// LearningLogPolicy learning log policy
// swagger:model LearningLogPolicy
type LearningLogPolicy struct {

	// Determine whether app learning logging is enabled. Field introduced in 20.1.3.
	Enabled *bool `json:"enabled,omitempty"`

	// Host name where learning logs will be sent to. Field introduced in 20.1.3.
	Host *string `json:"host,omitempty"`

	// Port number for the service listening for learning logs. Field introduced in 20.1.3.
	Port *int32 `json:"port,omitempty"`
}
