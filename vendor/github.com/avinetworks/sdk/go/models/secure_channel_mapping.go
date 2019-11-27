package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SecureChannelMapping secure channel mapping
// swagger:model SecureChannelMapping
type SecureChannelMapping struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// ip of SecureChannelMapping.
	IP *string `json:"ip,omitempty"`

	// Placeholder for description of property is_controller of obj type SecureChannelMapping field type str  type boolean
	IsController *bool `json:"is_controller,omitempty"`

	// local_ip of SecureChannelMapping.
	LocalIP *string `json:"local_ip,omitempty"`

	// Placeholder for description of property marked_for_delete of obj type SecureChannelMapping field type str  type boolean
	MarkedForDelete *bool `json:"marked_for_delete,omitempty"`

	// Name of the object.
	// Required: true
	Name *string `json:"name"`

	// pub_key of SecureChannelMapping.
	PubKey *string `json:"pub_key,omitempty"`

	// pub_key_pem of SecureChannelMapping.
	PubKeyPem *string `json:"pub_key_pem,omitempty"`

	//  Enum options - SECURE_CHANNEL_NONE, SECURE_CHANNEL_CONNECTED, SECURE_CHANNEL_AUTH_SSH_SUCCESS, SECURE_CHANNEL_AUTH_SSH_FAILED, SECURE_CHANNEL_AUTH_TOKEN_SUCCESS, SECURE_CHANNEL_AUTH_TOKEN_FAILED, SECURE_CHANNEL_AUTH_ERRORS, SECURE_CHANNEL_AUTH_IGNORED.
	Status *string `json:"status,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// Unique object identifier of the object.
	UUID *string `json:"uuid,omitempty"`
}
