package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// Image image
// swagger:model Image
type Image struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// Controller package details. Field introduced in 18.2.6.
	ControllerInfo *PackageDetails `json:"controller_info,omitempty"`

	// It references the controller-patch associated with the Uber image. Field introduced in 18.2.8.
	ControllerPatchUUID *string `json:"controller_patch_uuid,omitempty"`

	// This field describes the api migration related information. Field introduced in 18.2.6.
	Migrations *SupportedMigrations `json:"migrations,omitempty"`

	// Name of the image. Field introduced in 18.2.6.
	// Required: true
	Name *string `json:"name"`

	// SE package details. Field introduced in 18.2.6.
	SeInfo *PackageDetails `json:"se_info,omitempty"`

	// It references the Service Engine patch associated with the Uber Image. Field introduced in 18.2.8.
	SePatchUUID *string `json:"se_patch_uuid,omitempty"`

	// Status to check if the image is present. Enum options - SYSERR_SUCCESS, SYSERR_FAILURE, SYSERR_OUT_OF_MEMORY, SYSERR_NO_ENT, SYSERR_INVAL, SYSERR_ACCESS, SYSERR_FAULT, SYSERR_IO, SYSERR_TIMEOUT, SYSERR_NOT_SUPPORTED, SYSERR_NOT_READY, SYSERR_UPGRADE_IN_PROGRESS, SYSERR_WARM_START_IN_PROGRESS, SYSERR_TRY_AGAIN, SYSERR_NOT_UPGRADING, SYSERR_PENDING, SYSERR_EVENT_GEN_FAILURE, SYSERR_CONFIG_PARAM_MISSING, SYSERR_BAD_REQUEST, SYSERR_TEST1.... Field introduced in 18.2.6.
	Status *string `json:"status,omitempty"`

	// Tenant that this object belongs to. It is a reference to an object of type Tenant. Field introduced in 18.2.6.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// Type of the image patch/system. Enum options - IMAGE_TYPE_PATCH, IMAGE_TYPE_SYSTEM. Field introduced in 18.2.6.
	Type *string `json:"type,omitempty"`

	// Status to check if the image is an uber bundle. Field introduced in 18.2.8.
	UberBundle *bool `json:"uber_bundle,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// UUID of the image. Field introduced in 18.2.6.
	UUID *string `json:"uuid,omitempty"`
}
