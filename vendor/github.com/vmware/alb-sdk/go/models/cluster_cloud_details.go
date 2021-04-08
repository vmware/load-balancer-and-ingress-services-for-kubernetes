package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ClusterCloudDetails cluster cloud details
// swagger:model ClusterCloudDetails
type ClusterCloudDetails struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// Azure info to configure cluster_vip on the controller. Field introduced in 17.2.5.
	AzureInfo *AzureClusterInfo `json:"azure_info,omitempty"`

	//  Field introduced in 17.2.5.
	// Required: true
	Name *string `json:"name"`

	//  It is a reference to an object of type Tenant. Field introduced in 17.2.5.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	//  Field introduced in 17.2.5.
	UUID *string `json:"uuid,omitempty"`
}
