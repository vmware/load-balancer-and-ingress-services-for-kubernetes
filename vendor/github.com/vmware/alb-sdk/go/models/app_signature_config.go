// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// AppSignatureConfig app signature config
// swagger:model AppSignatureConfig
type AppSignatureConfig struct {

	// Application Signature db sync interval in minutes. Allowed values are 60-10080. Field introduced in 20.1.4. Unit is MIN. Allowed in Basic edition, Essentials edition, Enterprise edition. Special default for Basic edition is 1440, Essentials edition is 1440, Enterprise is 1440.
	AppSignatureSyncInterval *int32 `json:"app_signature_sync_interval,omitempty"`
}
