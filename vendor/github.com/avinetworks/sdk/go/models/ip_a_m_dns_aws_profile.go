package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// IPAMDNSAwsProfile ipam Dns aws profile
// swagger:model IpamDnsAwsProfile
type IPAMDNSAwsProfile struct {

	// AWS access key ID.
	AccessKeyID *string `json:"access_key_id,omitempty"`

	// IAM assume role for cross-account access. Field introduced in 17.1.1.
	IamAssumeRole *string `json:"iam_assume_role,omitempty"`

	// If enabled and the virtual service is not floating ip capable, vip will be published to both private and public zones. This flag is applicable only for AWS DNS profile. Field introduced in 17.2.10.
	PublishVipToPublicZone *bool `json:"publish_vip_to_public_zone,omitempty"`

	// AWS region.
	Region *string `json:"region,omitempty"`

	// AWS secret access key.
	SecretAccessKey *string `json:"secret_access_key,omitempty"`

	// Default TTL for all records. Allowed values are 1-172800. Field introduced in 17.1.3.
	TTL *int32 `json:"ttl,omitempty"`

	// Usable domains to pick from Amazon Route 53. Field introduced in 17.1.1.
	UsableDomains []string `json:"usable_domains,omitempty"`

	// Usable networks for Virtual IP. If VirtualService does not specify a network and auto_allocate_ip is set, then the first available network from this list will be chosen for IP allocation. Field introduced in 17.1.1.
	UsableNetworkUuids []string `json:"usable_network_uuids,omitempty"`

	// Use IAM roles instead of access and secret key.
	UseIamRoles *bool `json:"use_iam_roles,omitempty"`

	// VPC name.
	Vpc *string `json:"vpc,omitempty"`

	// VPC ID.
	// Required: true
	VpcID *string `json:"vpc_id"`

	// Network configuration for Virtual IP per AZ. Field introduced in 17.1.3.
	Zones []*AwsZoneNetwork `json:"zones,omitempty"`
}
