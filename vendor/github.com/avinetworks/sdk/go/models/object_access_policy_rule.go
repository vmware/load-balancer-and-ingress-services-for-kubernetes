package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ObjectAccessPolicyRule object access policy rule
// swagger:model ObjectAccessPolicyRule
type ObjectAccessPolicyRule struct {

	// Match criteria for the rule. Field introduced in 18.2.7, 20.1.1.
	Matches []*ObjectAccessMatchTarget `json:"matches,omitempty"`

	// Name of the rule. Field introduced in 18.2.7, 20.1.1.
	Name *string `json:"name,omitempty"`

	// Object types that this rule applies to. Enum options - VIRTUALSERVICE, POOL, HEALTHMONITOR, NETWORKPROFILE, APPLICATIONPROFILE, HTTPPOLICYSET, DNSPOLICY, SECURITYPOLICY, IPADDRGROUP, STRINGGROUP, SSLPROFILE, SSLKEYANDCERTIFICATE, NETWORKSECURITYPOLICY, APPLICATIONPERSISTENCEPROFILE, ANALYTICSPROFILE, VSDATASCRIPTSET, TENANT, PKIPROFILE, AUTHPROFILE, CLOUD.... Field introduced in 18.2.7, 20.1.1.
	// Required: true
	ObjTypes []string `json:"obj_types,omitempty"`

	// Privilege granted for objects matched by the rule. Enum options - NO_ACCESS, READ_ACCESS, WRITE_ACCESS. Field introduced in 18.2.7, 20.1.1.
	Privilege *string `json:"privilege,omitempty"`
}
