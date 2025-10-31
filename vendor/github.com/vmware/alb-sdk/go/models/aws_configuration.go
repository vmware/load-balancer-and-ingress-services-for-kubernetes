// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// AwsConfiguration aws configuration
// swagger:model AwsConfiguration
type AwsConfiguration struct {

	// AWS access key ID. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AccessKeyID *string `json:"access_key_id,omitempty"`

	// Time interval between periodic polling of all Auto Scaling Groups. Allowed values are 60-1800. Field introduced in 17.1.3. Unit is SEC. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AsgPollInterval *uint32 `json:"asg_poll_interval,omitempty"`

	// EBS encryption mode and the master key to be used for encrypting SE AMI, Volumes, and Snapshots. Field introduced in 17.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	EbsEncryption *AwsEncryption `json:"ebs_encryption,omitempty"`

	// Free unused elastic IP addresses. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	FreeElasticips *bool `json:"free_elasticips,omitempty"`

	// IAM assume role for cross-account access. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	IamAssumeRole *string `json:"iam_assume_role,omitempty"`

	// If enabled and the virtual service is not floating ip capable, vip will be published to both private and public zones. Field introduced in 17.2.10. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PublishVipToPublicZone *bool `json:"publish_vip_to_public_zone,omitempty"`

	// AWS region. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Region *string `json:"region,omitempty"`

	// If enabled, create/update DNS entries in Amazon Route 53 zones. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Route53Integration *bool `json:"route53_integration,omitempty"`

	// S3 encryption mode and the master key to be used for encrypting S3 buckets during SE AMI upload. Only SSE-KMS mode is supported. Field introduced in 17.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	S3Encryption *AwsEncryption `json:"s3_encryption,omitempty"`

	// AWS secret access key. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SecretAccessKey *string `json:"secret_access_key,omitempty"`

	// Server Side Encryption to be used for encrypting SQS Queues. Field introduced in 17.2.8. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SqsEncryption *AwsEncryption `json:"sqs_encryption,omitempty"`

	// Default TTL for all records. Allowed values are 1-172800. Field introduced in 17.1.3. Unit is SEC. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	TTL *uint32 `json:"ttl,omitempty"`

	// Use IAM roles instead of access and secret key. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UseIamRoles *bool `json:"use_iam_roles,omitempty"`

	// Use SNS/SQS based notifications for monitoring Auto Scaling Groups. Field introduced in 17.1.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UseSnsSqs *bool `json:"use_sns_sqs,omitempty"`

	// VPC name. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Vpc *string `json:"vpc,omitempty"`

	// VPC ID. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	VpcID *string `json:"vpc_id"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Zones []*AwsZoneConfig `json:"zones,omitempty"`
}
