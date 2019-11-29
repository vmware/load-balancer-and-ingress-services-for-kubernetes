package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// IPPersistenceProfile IP persistence profile
// swagger:model IPPersistenceProfile
type IPPersistenceProfile struct {

	// The length of time after a client's connections have closed before expiring the client's persistence to a server. Allowed values are 1-720.
	IPPersistentTimeout *int32 `json:"ip_persistent_timeout,omitempty"`
}
