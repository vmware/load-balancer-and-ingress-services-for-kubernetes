package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ProtocolParser protocol parser
// swagger:model ProtocolParser
type ProtocolParser struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// Description of the protocol parser. Field introduced in 18.2.3.
	Description *string `json:"description,omitempty"`

	// Name of the protocol parser. Field introduced in 18.2.3.
	// Required: true
	Name *string `json:"name"`

	// Command script provided inline. Field introduced in 18.2.3.
	// Required: true
	ParserCode *string `json:"parser_code"`

	// Tenant UUID of the protocol parser. It is a reference to an object of type Tenant. Field introduced in 18.2.3.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// UUID of the protocol parser. Field introduced in 18.2.3.
	UUID *string `json:"uuid,omitempty"`
}
