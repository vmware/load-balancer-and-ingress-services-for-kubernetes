// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// GCPConfiguration g c p configuration
// swagger:model GCPConfiguration
type GCPConfiguration struct {

	// Credentials to access Google Cloud Platform APIs. It is a reference to an object of type CloudConnectorUser. Field introduced in 18.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CloudCredentialsRef *string `json:"cloud_credentials_ref,omitempty"`

	// Encryption Keys for Google Cloud Services. Field introduced in 18.2.10, 20.1.2. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	EncryptionKeys *GCPEncryptionKeys `json:"encryption_keys,omitempty"`

	// Firewall rule network target tags which will be applied on Service Engines to allow ingress and egress traffic for Service Engines. Field introduced in 18.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	FirewallTargetTags []string `json:"firewall_target_tags,omitempty"`

	// Email of GCP Service Account to be associated to the Service Engines. Field introduced in 20.1.7, 21.1.2. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	GcpServiceAccountEmail *string `json:"gcp_service_account_email,omitempty"`

	// Google Cloud Storage Bucket Name where Service Engine image will be uploaded. This image will be deleted once the image is created in Google compute images. By default, a bucket will be created if this field is not specified. Field introduced in 18.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	GcsBucketName *string `json:"gcs_bucket_name,omitempty"`

	// Google Cloud Storage Project ID where Service Engine image will be uploaded. This image will be deleted once the image is created in Google compute images. By default, Service Engine Project ID will be used. Field introduced in 18.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	GcsProjectID *string `json:"gcs_project_id,omitempty"`

	// Google Cloud Platform VPC Network configuration for the Service Engines. Field introduced in 18.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	NetworkConfig *GCPNetworkConfig `json:"network_config"`

	// Google Cloud Platform Region Name where Service Engines will be spawned. Field introduced in 18.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	RegionName *string `json:"region_name"`

	// Google Cloud Platform Project ID where Service Engines will be spawned. Field introduced in 18.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	SeProjectID *string `json:"se_project_id"`

	// VIP allocation strategy defines how the VIPs will be created in Google Cloud. Field introduced in 18.2.9, 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	VipAllocationStrategy *GCPVIPAllocation `json:"vip_allocation_strategy"`

	// Google Cloud Platform Zones where Service Engines will be distributed for HA. Field introduced in 18.2.1. Minimum of 1 items required. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Zones []string `json:"zones,omitempty"`
}
