package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// IPPersistenceProfile IP persistence profile
// swagger:model IPPersistenceProfile
type IPPersistenceProfile struct {

	// Mask to be applied on client IP. This may be used to persist clients from a subnet to the same server. When set to 0, all requests are sent to the same server. Allowed values are 0-128. Field introduced in 18.2.7.
	IPMask *int32 `json:"ip_mask,omitempty"`

	// The length of time after a client's connections have closed before expiring the client's persistence to a server. Allowed values are 1-720.
	IPPersistentTimeout *int32 `json:"ip_persistent_timeout,omitempty"`
}
