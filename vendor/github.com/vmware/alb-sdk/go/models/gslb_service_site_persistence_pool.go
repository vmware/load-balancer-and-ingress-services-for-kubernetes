package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// GslbServiceSitePersistencePool gslb service site persistence pool
// swagger:model GslbServiceSitePersistencePool
type GslbServiceSitePersistencePool struct {

	// Site persistence pool's name. . Field introduced in 17.2.2.
	Name *string `json:"name,omitempty"`

	// Number of servers configured in the pool. . Field introduced in 17.2.2.
	NumServers *int64 `json:"num_servers,omitempty"`

	// Number of servers operationally up in the pool. . Field introduced in 17.2.2.
	NumServersUp *int64 `json:"num_servers_up,omitempty"`

	// Detailed information of the servers in the pool. . Field introduced in 17.2.8.
	Servers []*ServerConfig `json:"servers,omitempty"`

	// Site persistence pool's uuid. . Field introduced in 17.2.2.
	UUID *string `json:"uuid,omitempty"`
}
