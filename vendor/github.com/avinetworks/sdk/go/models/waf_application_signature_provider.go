package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// WafApplicationSignatureProvider waf application signature provider
// swagger:model WafApplicationSignatureProvider
type WafApplicationSignatureProvider struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// Available application names and the ruleset version, when the rules for an application changed the last time. Field introduced in 20.1.1.
	// Read Only: true
	AvailableApplications []*WafApplicationSignatureAppVersion `json:"available_applications,omitempty"`

	// The error message indicating why the last update check failed. Field introduced in 20.1.1.
	// Read Only: true
	LastCheckForUpdatesError *string `json:"last_check_for_updates_error,omitempty"`

	// The last time that we checked for updates but did not get a result because of an error. Field introduced in 20.1.1.
	// Read Only: true
	LastFailedCheckForUpdates *TimeStamp `json:"last_failed_check_for_updates,omitempty"`

	// The last time that we checked for updates sucessfully. Field introduced in 20.1.1.
	// Read Only: true
	LastSuccessfulCheckForUpdates *TimeStamp `json:"last_successful_check_for_updates,omitempty"`

	// Name of Application Specific Ruleset Provider. Field introduced in 20.1.1.
	Name *string `json:"name,omitempty"`

	// Version of signatures. Field introduced in 20.1.1.
	// Read Only: true
	RulesetVersion *string `json:"ruleset_version,omitempty"`

	// The WAF rules. Not visible in the API. Field introduced in 20.1.1.
	// Read Only: true
	Signatures []*WafRule `json:"signatures,omitempty"`

	//  It is a reference to an object of type Tenant. Field introduced in 20.1.1.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	//  Field introduced in 20.1.1.
	UUID *string `json:"uuid,omitempty"`
}
