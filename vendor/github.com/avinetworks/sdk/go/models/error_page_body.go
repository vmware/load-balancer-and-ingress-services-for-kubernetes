package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ErrorPageBody error page body
// swagger:model ErrorPageBody
type ErrorPageBody struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// Error page body sent to client when match. Field introduced in 17.2.4.
	ErrorPageBody *string `json:"error_page_body,omitempty"`

	// Format of an error page body HTML or JSON. Enum options - ERROR_PAGE_FORMAT_HTML, ERROR_PAGE_FORMAT_JSON. Field introduced in 18.2.3.
	Format *string `json:"format,omitempty"`

	//  Field introduced in 17.2.4.
	// Required: true
	Name *string `json:"name"`

	//  It is a reference to an object of type Tenant. Field introduced in 17.2.4.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	//  Field introduced in 17.2.4.
	UUID *string `json:"uuid,omitempty"`
}
