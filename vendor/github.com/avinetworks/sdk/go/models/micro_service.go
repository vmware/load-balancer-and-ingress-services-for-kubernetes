package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// MicroService micro service
// swagger:model MicroService
type MicroService struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// application_name of MicroService.
	// Required: true
	ApplicationName *string `json:"application_name"`

	// Checksum of cloud configuration for Microservice. Internally set by cloud connector. Field introduced in 17.2.8.
	CloudConfigCksum *string `json:"cloud_config_cksum,omitempty"`

	// The list of containers for this microservice.
	Containers []*MicroServiceContainer `json:"containers,omitempty"`

	// Creator name.
	CreatedBy *string `json:"created_by,omitempty"`

	// User defined description for the object.
	Description *string `json:"description,omitempty"`

	// Flag to indicate if container IP list is provided by cloud connectorThis is applicable for overlay cases.
	IPList *bool `json:"ip_list,omitempty"`

	// Name of the object.
	// Required: true
	Name *string `json:"name"`

	// orchestrator_name of MicroService.
	// Required: true
	OrchestratorName *string `json:"orchestrator_name"`

	//  It is a reference to an object of type Tenant.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// Unique object identifier of the object.
	UUID *string `json:"uuid,omitempty"`
}
