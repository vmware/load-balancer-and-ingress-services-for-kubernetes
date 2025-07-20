// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// AzureClusterInfo azure cluster info
// swagger:model AzureClusterInfo
type AzureClusterInfo struct {

	//  It is a reference to an object of type CloudConnectorUser. Field introduced in 17.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	CloudCredentialRef *string `json:"cloud_credential_ref"`

	//  Field introduced in 17.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	SubscriptionID *string `json:"subscription_id"`
}
