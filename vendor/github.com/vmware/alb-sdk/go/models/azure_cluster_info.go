// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

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
