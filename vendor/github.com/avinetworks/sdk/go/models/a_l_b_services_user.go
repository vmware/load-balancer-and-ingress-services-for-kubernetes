package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ALBServicesUser a l b services user
// swagger:model ALBServicesUser
type ALBServicesUser struct {

	// ID of primary account of the portal user. Field introduced in 20.1.1.
	AccountID *string `json:"account_id,omitempty"`

	// Name of primary account of the portal user. Field introduced in 20.1.1.
	AccountName *string `json:"account_name,omitempty"`

	// Email ID of the portal user. Field introduced in 20.1.1.
	// Required: true
	Email *string `json:"email"`

	// Information about all the accounts managed by user in the customer portal. Field introduced in 20.1.1.
	ManagedAccounts []*ALBServicesAccount `json:"managed_accounts,omitempty"`

	// Name of the portal user. Field introduced in 20.1.1.
	Name *string `json:"name,omitempty"`

	// Phone number of the user. Field introduced in 20.1.1.
	Phone *string `json:"phone,omitempty"`
}
