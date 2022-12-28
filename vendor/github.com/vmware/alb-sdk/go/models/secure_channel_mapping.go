// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SecureChannelMapping secure channel mapping
// swagger:model SecureChannelMapping
type SecureChannelMapping struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// Auth_token used for SE authorization. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	AuthToken *string `json:"auth_token,omitempty"`

	// IP of SE. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	IP *string `json:"ip,omitempty"`

	// Whether this entry used for controller. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	IsController *bool `json:"is_controller,omitempty"`

	// Local ip on controller side reserved for SE. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	LocalIP *string `json:"local_ip,omitempty"`

	// Whether this entry is marked for delete (first step of deletion). Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MarkedForDelete *bool `json:"marked_for_delete,omitempty"`

	// Metadata associated with auth_token. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Metadata []*SecureChannelMetadata `json:"metadata,omitempty"`

	// Uuid of SE. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Name *string `json:"name"`

	// Public key of SE. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PubKey *string `json:"pub_key,omitempty"`

	// Public key pem of SE. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PubKeyPem *string `json:"pub_key_pem,omitempty"`

	// Authorization status of current secure channel. Enum options - SECURE_CHANNEL_NONE, SECURE_CHANNEL_CONNECTED, SECURE_CHANNEL_AUTH_SSH_SUCCESS, SECURE_CHANNEL_AUTH_SSH_FAILED, SECURE_CHANNEL_AUTH_TOKEN_SUCCESS, SECURE_CHANNEL_AUTH_TOKEN_FAILED, SECURE_CHANNEL_AUTH_ERRORS, SECURE_CHANNEL_AUTH_IGNORED. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Status *string `json:"status,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// Uuid of SE. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UUID *string `json:"uuid,omitempty"`
}
