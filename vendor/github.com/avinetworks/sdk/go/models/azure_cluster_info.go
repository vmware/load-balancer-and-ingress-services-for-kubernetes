package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// AzureClusterInfo azure cluster info
// swagger:model AzureClusterInfo
type AzureClusterInfo struct {

	//  It is a reference to an object of type CloudConnectorUser. Field introduced in 17.2.5.
	// Required: true
	CloudCredentialRef *string `json:"cloud_credential_ref"`

	//  Field introduced in 17.2.5.
	// Required: true
	SubscriptionID *string `json:"subscription_id"`
}
