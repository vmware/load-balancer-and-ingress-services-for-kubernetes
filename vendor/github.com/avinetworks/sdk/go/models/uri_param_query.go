package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// URIParamQuery URI param query
// swagger:model URIParamQuery
type URIParamQuery struct {

	// Concatenate a *string to the query of the incoming request URI and then use it in the request URI going to the backend server.
	AddString *string `json:"add_string,omitempty"`

	// Use or drop the query of the incoming request URI in the request URI to the backend server.
	KeepQuery *bool `json:"keep_query,omitempty"`
}
