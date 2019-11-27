package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// TacacsPlusAuthSettings tacacs plus auth settings
// swagger:model TacacsPlusAuthSettings
type TacacsPlusAuthSettings struct {

	// TACACS+ authorization attribute value pairs.
	AuthorizationAttrs []*AuthTacacsPlusAttributeValuePair `json:"authorization_attrs,omitempty"`

	// TACACS+ server shared secret.
	Password *string `json:"password,omitempty"`

	// TACACS+ server listening port.
	Port *int32 `json:"port,omitempty"`

	// TACACS+ server IP address.
	Server []string `json:"server,omitempty"`

	// TACACS+ service. Enum options - AUTH_TACACS_PLUS_SERVICE_NONE, AUTH_TACACS_PLUS_SERVICE_LOGIN, AUTH_TACACS_PLUS_SERVICE_ENABLE, AUTH_TACACS_PLUS_SERVICE_PPP, AUTH_TACACS_PLUS_SERVICE_ARAP, AUTH_TACACS_PLUS_SERVICE_PT, AUTH_TACACS_PLUS_SERVICE_RCMD, AUTH_TACACS_PLUS_SERVICE_X25, AUTH_TACACS_PLUS_SERVICE_NASI, AUTH_TACACS_PLUS_SERVICE_FWPROXY.
	Service *string `json:"service,omitempty"`
}
