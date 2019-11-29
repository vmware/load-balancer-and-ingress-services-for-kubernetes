package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// GCPConfiguration g c p configuration
// swagger:model GCPConfiguration
type GCPConfiguration struct {

	// Credentials to access Google Cloud Platform APIs. It is a reference to an object of type CloudConnectorUser. Field introduced in 18.2.1.
	CloudCredentialsRef *string `json:"cloud_credentials_ref,omitempty"`

	// Firewall rule network target tags which will be applied on Service Engines to allow ingress and egress traffic for Service Engines. Field introduced in 18.2.1.
	FirewallTargetTags []string `json:"firewall_target_tags,omitempty"`

	// Google Cloud Storage Bucket Name where Service Engine image will be uploaded. This image will be deleted once the image is created in Google compute images. By default, a bucket will be created if this field is not specified. Field introduced in 18.2.1.
	GcsBucketName *string `json:"gcs_bucket_name,omitempty"`

	// Google Cloud Storage Project ID where Service Engine image will be uploaded. This image will be deleted once the image is created in Google compute images. By default, se_project_id will be used. Field introduced in 18.2.1.
	GcsProjectID *string `json:"gcs_project_id,omitempty"`

	// Match SE group subnets for VIP placement. Default is to not match SE group subnets. Field introduced in 18.2.1.
	MatchSeGroupSubnet *bool `json:"match_se_group_subnet,omitempty"`

	// Google Cloud Platform VPC Network configuration for the Service Engines. Field introduced in 18.2.1.
	// Required: true
	NetworkConfig *GCPNetworkConfig `json:"network_config"`

	// Google Cloud Platform Region Name where Service Engines will be spawned. Field introduced in 18.2.1.
	// Required: true
	RegionName *string `json:"region_name"`

	// Google Cloud Platform Project ID where Service Engines will be spawned. Field introduced in 18.2.1.
	// Required: true
	SeProjectID *string `json:"se_project_id"`

	// Google Cloud Platform Zones where Service Engines will be distributed for HA. By default, all the zones in a region will be used. Field introduced in 18.2.1.
	Zones []string `json:"zones,omitempty"`
}
