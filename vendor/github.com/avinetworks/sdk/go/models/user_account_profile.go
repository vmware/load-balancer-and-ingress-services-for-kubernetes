package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// UserAccountProfile user account profile
// swagger:model UserAccountProfile
type UserAccountProfile struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// Lock timeout period (in minutes). Default is 30 minutes. Unit is MIN.
	AccountLockTimeout *int32 `json:"account_lock_timeout,omitempty"`

	// The time period after which credentials expire. Default is 180 days. Unit is DAYS.
	CredentialsTimeoutThreshold *int32 `json:"credentials_timeout_threshold,omitempty"`

	// Maximum number of concurrent sessions allowed. There are unlimited sessions by default.
	MaxConcurrentSessions *int32 `json:"max_concurrent_sessions,omitempty"`

	// Number of login attempts before lockout. Default is 3 attempts. Allowed values are 3-20. Special values are 0 - 'Unlimited login attempts allowed.'.
	MaxLoginFailureCount *int32 `json:"max_login_failure_count,omitempty"`

	// Maximum number of passwords to be maintained in the password history. Default is 4 passwords.
	MaxPasswordHistoryCount *int32 `json:"max_password_history_count,omitempty"`

	// Name of the object.
	// Required: true
	Name *string `json:"name"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// Unique object identifier of the object.
	UUID *string `json:"uuid,omitempty"`
}
