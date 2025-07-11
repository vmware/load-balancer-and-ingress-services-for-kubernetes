// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// TacacsPlusAuthSettings tacacs plus auth settings
// swagger:model TacacsPlusAuthSettings
type TacacsPlusAuthSettings struct {

	// TACACS+ authorization attribute value pairs. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AuthorizationAttrs []*AuthTacacsPlusAttributeValuePair `json:"authorization_attrs,omitempty"`

	// TACACS+ server shared secret. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Password *string `json:"password,omitempty"`

	// TACACS+ server listening port. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Port *uint32 `json:"port,omitempty"`

	// TACACS+ server IP address or FQDN. Minimum of 1 items required. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Server []string `json:"server,omitempty"`

	// TACACS+ service. Enum options - AUTH_TACACS_PLUS_SERVICE_NONE, AUTH_TACACS_PLUS_SERVICE_LOGIN, AUTH_TACACS_PLUS_SERVICE_ENABLE, AUTH_TACACS_PLUS_SERVICE_PPP, AUTH_TACACS_PLUS_SERVICE_ARAP, AUTH_TACACS_PLUS_SERVICE_PT, AUTH_TACACS_PLUS_SERVICE_RCMD, AUTH_TACACS_PLUS_SERVICE_X25, AUTH_TACACS_PLUS_SERVICE_NASI, AUTH_TACACS_PLUS_SERVICE_FWPROXY. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Service *string `json:"service,omitempty"`
}
