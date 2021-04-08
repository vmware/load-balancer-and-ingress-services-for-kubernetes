package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ControllerPortalAuth controller portal auth
// swagger:model ControllerPortalAuth
type ControllerPortalAuth struct {

	// Access Token to authenticate Customer Portal REST calls. Field introduced in 18.2.6.
	AccessToken *string `json:"access_token,omitempty"`

	// Salesforce instance URL. Field introduced in 18.2.6.
	InstanceURL *string `json:"instance_url,omitempty"`

	// Signed JWT to refresh the access token. Field introduced in 18.2.6.
	JwtToken *string `json:"jwt_token,omitempty"`
}
