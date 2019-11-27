package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// UserActivity user activity
// swagger:model UserActivity
type UserActivity struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// Number of concurrent user sessions open.
	ConcurrentSessions *int32 `json:"concurrent_sessions,omitempty"`

	// Number of failed login attempts before a successful login.
	FailedLoginAttempts *int32 `json:"failed_login_attempts,omitempty"`

	// IP of the machine the user was last logged in from.
	LastLoginIP *string `json:"last_login_ip,omitempty"`

	// Timestamp of last login.
	LastLoginTimestamp *string `json:"last_login_timestamp,omitempty"`

	// Timestamp of last password update.
	LastPasswordUpdate *string `json:"last_password_update,omitempty"`

	// Indicates whether the user is logged in or not.
	LoggedIn *bool `json:"logged_in,omitempty"`

	// Name of the user this object refers to.
	Name *string `json:"name,omitempty"`

	// Stores the previous n passwords  where n is ControllerProperties.max_password_history_count. .
	PreviousPassword []string `json:"previous_password,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// Unique object identifier of the object.
	UUID *string `json:"uuid,omitempty"`
}
