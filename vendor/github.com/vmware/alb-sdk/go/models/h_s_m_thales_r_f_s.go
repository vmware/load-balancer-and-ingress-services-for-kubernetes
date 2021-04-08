package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// HSMThalesRFS h s m thales r f s
// swagger:model HSMThalesRFS
type HSMThalesRFS struct {

	// IP address of the RFS server from where to sync the Thales encrypted private key.
	// Required: true
	IP *IPAddr `json:"ip"`

	// Port at which the RFS server accepts the sync request from clients for Thales encrypted private key. Allowed values are 1-65535.
	Port *int32 `json:"port,omitempty"`
}
