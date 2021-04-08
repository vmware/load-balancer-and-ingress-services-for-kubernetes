package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// JWTValidationVsConfig j w t validation vs config
// swagger:model JWTValidationVsConfig
type JWTValidationVsConfig struct {

	// Uniquely identifies a resource server. This is used to validate against the aud claim. Field introduced in 20.1.3.
	// Required: true
	Audience *string `json:"audience"`

	// Defines where to look for JWT in the request. Enum options - JWT_LOCATION_AUTHORIZATION_HEADER, JWT_LOCATION_QUERY_PARAM. Field introduced in 20.1.3.
	// Required: true
	JwtLocation *string `json:"jwt_location"`

	// Name by which the JWT can be identified if the token is sent as a query param in the request url. Field introduced in 20.1.3.
	JwtName *string `json:"jwt_name,omitempty"`
}
