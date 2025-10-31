// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// VSDataScriptSet v s data script set
// swagger:model VSDataScriptSet
type VSDataScriptSet struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// Protobuf versioning for config pbs. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	ConfigpbAttributes *ConfigPbAttributes `json:"configpb_attributes,omitempty"`

	// Creator name. Field introduced in 17.1.11,17.2.4. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CreatedBy *string `json:"created_by,omitempty"`

	// DataScripts to execute. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Datascript []*VSDataScript `json:"datascript,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Description *string `json:"description,omitempty"`

	// Geo Location Mapping Database used by this DataScriptSet. It is a reference to an object of type GeoDB. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	GeoDbRef *string `json:"geo_db_ref,omitempty"`

	// IP reputation database that can be used by DataScript functions. It is a reference to an object of type IPReputationDB. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	IPReputationDbRef *string `json:"ip_reputation_db_ref,omitempty"`

	// UUID of IP Groups that could be referred by VSDataScriptSet objects. It is a reference to an object of type IpAddrGroup. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	IpgroupRefs []string `json:"ipgroup_refs,omitempty"`

	// List of labels to be used for granular RBAC. Field introduced in 20.1.5. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	Markers []*RoleFilterMatchLabel `json:"markers,omitempty"`

	// Name for the virtual service datascript collection. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Name *string `json:"name"`

	// UUIDs of PKIProfile objects that could be referred by VSDataScriptSet objects. It is a reference to an object of type PKIProfile. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	PkiProfileRefs []string `json:"pki_profile_refs,omitempty"`

	// UUID of pool groups that could be referred by VSDataScriptSet objects. It is a reference to an object of type PoolGroup. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PoolGroupRefs []string `json:"pool_group_refs,omitempty"`

	// UUID of pools that could be referred by VSDataScriptSet objects. It is a reference to an object of type Pool. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	PoolRefs []string `json:"pool_refs,omitempty"`

	// List of protocol parsers that could be referred by VSDataScriptSet objects. It is a reference to an object of type ProtocolParser. Field introduced in 18.2.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ProtocolParserRefs []string `json:"protocol_parser_refs,omitempty"`

	// The Rate Limit definitions needed for this DataScript. The name is composed of the Virtual Service name and the DataScript name. Field introduced in 18.2.9. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	RateLimiters []*RateLimiter `json:"rate_limiters,omitempty"`

	// UUIDs of SSLKeyAndCertificate objects that could be referred by VSDataScriptSet objects. It is a reference to an object of type SSLKeyAndCertificate. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SslKeyCertificateRefs []string `json:"ssl_key_certificate_refs,omitempty"`

	// UUIDs of SSLProfile objects that could be referred by VSDataScriptSet objects. It is a reference to an object of type SSLProfile. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SslProfileRefs []string `json:"ssl_profile_refs,omitempty"`

	// UUID of String Groups that could be referred by VSDataScriptSet objects. It is a reference to an object of type StringGroup. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	StringGroupRefs []string `json:"string_group_refs,omitempty"`

	//  It is a reference to an object of type Tenant. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// UUID of the virtual service datascript collection. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UUID *string `json:"uuid,omitempty"`
}
