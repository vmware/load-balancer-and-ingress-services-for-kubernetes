// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// HSMAwsCloudHsm h s m aws cloud hsm
// swagger:model HSMAwsCloudHsm
type HSMAwsCloudHsm struct {

	//  Field introduced in 17.2.7. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	// Read Only: true
	ClientConfig *string `json:"client_config,omitempty"`

	// AWS CloudHSM Cluster Certificate. Field introduced in 17.2.7. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ClusterCert *string `json:"cluster_cert,omitempty"`

	// Username of the Crypto User. This will be used to access the keys on the HSM . Field introduced in 17.2.7. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CryptoUserName *string `json:"crypto_user_name,omitempty"`

	// Password of the Crypto User. This will be used to access the keys on the HSM . Field introduced in 17.2.7. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CryptoUserPassword *string `json:"crypto_user_password,omitempty"`

	// IP address of the HSM in the cluster. If there are more than one HSMs, only one is sufficient. Field introduced in 17.2.7. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	HsmIP []string `json:"hsm_ip,omitempty"`

	//  Field introduced in 17.2.7. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	// Read Only: true
	MgmtConfig *string `json:"mgmt_config,omitempty"`
}
