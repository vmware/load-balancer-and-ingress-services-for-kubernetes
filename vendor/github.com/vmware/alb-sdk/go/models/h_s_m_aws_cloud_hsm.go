package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// HSMAwsCloudHsm h s m aws cloud hsm
// swagger:model HSMAwsCloudHsm
type HSMAwsCloudHsm struct {

	//  Field introduced in 17.2.7.
	// Read Only: true
	ClientConfig *string `json:"client_config,omitempty"`

	// AWS CloudHSM Cluster Certificate. Field introduced in 17.2.7.
	ClusterCert *string `json:"cluster_cert,omitempty"`

	// Username of the Crypto User. This will be used to access the keys on the HSM . Field introduced in 17.2.7.
	CryptoUserName *string `json:"crypto_user_name,omitempty"`

	// Password of the Crypto User. This will be used to access the keys on the HSM . Field introduced in 17.2.7.
	CryptoUserPassword *string `json:"crypto_user_password,omitempty"`

	// IP address of the HSM in the cluster. If there are more than one HSMs, only one is sufficient. Field introduced in 17.2.7.
	HsmIP []string `json:"hsm_ip,omitempty"`

	//  Field introduced in 17.2.7.
	// Read Only: true
	MgmtConfig *string `json:"mgmt_config,omitempty"`
}
