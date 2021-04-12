package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// JWTProfile j w t profile
// swagger:model JWTProfile
type JWTProfile struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// This field describes the object's replication scope. If the field is set to false, then the object is visible within the controller-cluster.  If the field is set to true, then the object is replicated across the federation.  . Field introduced in 20.1.5. Allowed in Basic(Allowed values- false) edition, Essentials(Allowed values- false) edition, Enterprise edition.
	IsFederated *bool `json:"is_federated,omitempty"`

	// JWK keys used for signing/validating the JWT. Field introduced in 20.1.5. Minimum of 1 items required. Maximum of 1 items allowed.
	JwksKeys []*JWSKey `json:"jwks_keys,omitempty"`

	// JWT auth type for JWT validation. Enum options - JWT_TYPE_JWS. Field introduced in 20.1.5.
	// Required: true
	JwtAuthType *string `json:"jwt_auth_type"`

	// A user friendly name for this jwt profile. Field introduced in 20.1.5.
	// Required: true
	Name *string `json:"name"`

	// UUID of the Tenant. It is a reference to an object of type Tenant. Field introduced in 20.1.5.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// UUID of the jwt profile. Field introduced in 20.1.5.
	UUID *string `json:"uuid,omitempty"`
}
