// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// WafCRS waf c r s
// swagger:model WafCRS
type WafCRS struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// Allowed request content type character sets in WAF. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	AllowedRequestContentTypeCharsets []string `json:"allowed_request_content_type_charsets,omitempty"`

	// Protobuf versioning for config pbs. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	ConfigpbAttributes *ConfigPbAttributes `json:"configpb_attributes,omitempty"`

	// A short description of this ruleset. Field introduced in 18.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Description *string `json:"description"`

	// List of Data Files used for WAF Rules. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Files []*WafDataFile `json:"files,omitempty"`

	// WAF Rules are sorted in groups based on their characterization. Field introduced in 18.1.1. Maximum of 64 items allowed. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Groups []*WafRuleGroup `json:"groups,omitempty"`

	// Integrity protection value. Field introduced in 18.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Integrity *string `json:"integrity"`

	// Integrity protection values for CRS. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	IntegrityValues []string `json:"integrity_values,omitempty"`

	// List of labels to be used for granular RBAC. Field introduced in 20.1.6. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	Markers []*RoleFilterMatchLabel `json:"markers,omitempty"`

	// The name of this ruleset object. Field introduced in 18.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Name *string `json:"name"`

	// The release date of this version in RFC 3339 / ISO 8601 format. Field introduced in 18.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	ReleaseDate *string `json:"release_date"`

	// WAF Restricted File Extensions. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	RestrictedExtensions []string `json:"restricted_extensions,omitempty"`

	// WAF Restricted HTTP Headers. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	RestrictedHeaders []string `json:"restricted_headers,omitempty"`

	// Tenant that this object belongs to. It is a reference to an object of type Tenant. Field introduced in 18.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	//  Field introduced in 18.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UUID *string `json:"uuid,omitempty"`

	// The version of this ruleset object. Field introduced in 18.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Version *string `json:"version"`
}
