package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SePersistenceEventDetails se persistence event details
// swagger:model SePersistenceEventDetails
type SePersistenceEventDetails struct {

	// Current number of entries in the client ip persistence table.
	Entries *int32 `json:"entries,omitempty"`

	// Name of pool whose persistence table limits were reached. It is a reference to an object of type Pool.
	Pool *string `json:"pool,omitempty"`

	// Type of persistence. Enum options - PERSISTENCE_TYPE_CLIENT_IP_ADDRESS, PERSISTENCE_TYPE_HTTP_COOKIE, PERSISTENCE_TYPE_TLS, PERSISTENCE_TYPE_CLIENT_IPV6_ADDRESS, PERSISTENCE_TYPE_CUSTOM_HTTP_HEADER, PERSISTENCE_TYPE_APP_COOKIE, PERSISTENCE_TYPE_GSLB_SITE.
	Type *string `json:"type,omitempty"`
}
