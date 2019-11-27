package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// AwsConfiguration aws configuration
// swagger:model AwsConfiguration
type AwsConfiguration struct {

	// AWS access key ID.
	AccessKeyID *string `json:"access_key_id,omitempty"`

	// Time interval between periodic polling of all Auto Scaling Groups. Allowed values are 60-1800. Field introduced in 17.1.3.
	AsgPollInterval *int32 `json:"asg_poll_interval,omitempty"`

	// EBS encryption mode and the master key to be used for encrypting SE AMI, Volumes, and Snapshots. Field introduced in 17.2.3.
	EbsEncryption *AwsEncryption `json:"ebs_encryption,omitempty"`

	// Free unused elastic IP addresses.
	FreeElasticips *bool `json:"free_elasticips,omitempty"`

	// IAM assume role for cross-account access.
	IamAssumeRole *string `json:"iam_assume_role,omitempty"`

	// If enabled and the virtual service is not floating ip capable, vip will be published to both private and public zones. Field introduced in 17.2.10.
	PublishVipToPublicZone *bool `json:"publish_vip_to_public_zone,omitempty"`

	// AWS region.
	Region *string `json:"region,omitempty"`

	// If enabled, create/update DNS entries in Amazon Route 53 zones.
	Route53Integration *bool `json:"route53_integration,omitempty"`

	// S3 encryption mode and the master key to be used for encrypting S3 buckets during SE AMI upload. Only SSE-KMS mode is supported. Field introduced in 17.2.3.
	S3Encryption *AwsEncryption `json:"s3_encryption,omitempty"`

	// AWS secret access key.
	SecretAccessKey *string `json:"secret_access_key,omitempty"`

	// Server Side Encryption to be used for encrypting SQS Queues. Field introduced in 17.2.8.
	SqsEncryption *AwsEncryption `json:"sqs_encryption,omitempty"`

	// Default TTL for all records. Allowed values are 1-172800. Field introduced in 17.1.3.
	TTL *int32 `json:"ttl,omitempty"`

	// Use IAM roles instead of access and secret key.
	UseIamRoles *bool `json:"use_iam_roles,omitempty"`

	// Use SNS/SQS based notifications for monitoring Auto Scaling Groups. Field introduced in 17.1.3.
	UseSnsSqs *bool `json:"use_sns_sqs,omitempty"`

	// VPC name.
	Vpc *string `json:"vpc,omitempty"`

	// VPC ID.
	// Required: true
	VpcID *string `json:"vpc_id"`

	// If enabled, program SE security group with ingress rule to allow SSH (port 22) access from 0.0.0.0/0. Field deprecated in 17.1.5. Field introduced in 17.1.3.
	WildcardAccess *bool `json:"wildcard_access,omitempty"`

	// Placeholder for description of property zones of obj type AwsConfiguration field type str  type object
	Zones []*AwsZoneConfig `json:"zones,omitempty"`
}
