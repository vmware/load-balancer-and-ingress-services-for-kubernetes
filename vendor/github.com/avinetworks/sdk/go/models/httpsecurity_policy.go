package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// HttpsecurityPolicy httpsecurity policy
// swagger:model HTTPSecurityPolicy
type HttpsecurityPolicy struct {

	// Add rules to the HTTP security policy.
	Rules []*HttpsecurityRule `json:"rules,omitempty"`
}
