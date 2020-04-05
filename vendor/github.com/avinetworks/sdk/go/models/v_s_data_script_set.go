package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// VSDataScriptSet v s data script set
// swagger:model VSDataScriptSet
type VSDataScriptSet struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// Creator name. Field introduced in 17.1.11,17.2.4.
	CreatedBy *string `json:"created_by,omitempty"`

	// DataScripts to execute.
	Datascript []*VSDataScript `json:"datascript,omitempty"`

	// User defined description for the object.
	Description *string `json:"description,omitempty"`

	// UUID of IP Groups that could be referred by VSDataScriptSet objects. It is a reference to an object of type IpAddrGroup.
	IpgroupRefs []string `json:"ipgroup_refs,omitempty"`

	// Name for the virtual service datascript collection.
	// Required: true
	Name *string `json:"name"`

	// UUID of pool groups that could be referred by VSDataScriptSet objects. It is a reference to an object of type PoolGroup.
	PoolGroupRefs []string `json:"pool_group_refs,omitempty"`

	// UUID of pools that could be referred by VSDataScriptSet objects. It is a reference to an object of type Pool.
	PoolRefs []string `json:"pool_refs,omitempty"`

	// List of protocol parsers that could be referred by VSDataScriptSet objects. It is a reference to an object of type ProtocolParser. Field introduced in 18.2.3.
	ProtocolParserRefs []string `json:"protocol_parser_refs,omitempty"`

	// UUID of String Groups that could be referred by VSDataScriptSet objects. It is a reference to an object of type StringGroup.
	StringGroupRefs []string `json:"string_group_refs,omitempty"`

	//  It is a reference to an object of type Tenant.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// UUID of the virtual service datascript collection.
	UUID *string `json:"uuid,omitempty"`
}
