package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// TestSeDatastoreLevel1 test se datastore level1
// swagger:model TestSeDatastoreLevel1
type TestSeDatastoreLevel1 struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// Name of the object.
	// Required: true
	Name *string `json:"name"`

	//  It is a reference to an object of type Tenant. Field introduced in 18.2.6.
	TenantRef *string `json:"tenant_ref,omitempty"`

	//  It is a reference to an object of type TestSeDatastoreLevel2. Field introduced in 18.2.6.
	TestSeDatastoreLevel2Ref *string `json:"test_se_datastore_level_2_ref,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// Unique object identifier of the object.
	UUID *string `json:"uuid,omitempty"`
}
