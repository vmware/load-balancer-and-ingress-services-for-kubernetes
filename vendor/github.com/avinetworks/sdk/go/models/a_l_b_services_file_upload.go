package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ALBServicesFileUpload a l b services file upload
// swagger:model ALBServicesFileUpload
type ALBServicesFileUpload struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// Salesforce alphanumeric CaseID to attach uploaded file to. Field introduced in 18.2.6.
	CaseID *string `json:"case_id,omitempty"`

	// Error reported during file upload. Field introduced in 18.2.6.
	// Read Only: true
	Error *string `json:"error,omitempty"`

	// Stores output file path, for upload to AWS S3. Field introduced in 18.2.6.
	// Required: true
	FilePath *string `json:"file_path"`

	//  Field introduced in 18.2.6.
	// Required: true
	Name *string `json:"name"`

	// Custom AWS S3 Directory Path to upload file. Field introduced in 18.2.6.
	S3Directory *string `json:"s3_directory,omitempty"`

	// Captures status for file upload. Enum options - SYSERR_SUCCESS, SYSERR_FAILURE, SYSERR_OUT_OF_MEMORY, SYSERR_NO_ENT, SYSERR_INVAL, SYSERR_ACCESS, SYSERR_FAULT, SYSERR_IO, SYSERR_TIMEOUT, SYSERR_NOT_SUPPORTED, SYSERR_NOT_READY, SYSERR_UPGRADE_IN_PROGRESS, SYSERR_WARM_START_IN_PROGRESS, SYSERR_TRY_AGAIN, SYSERR_NOT_UPGRADING, SYSERR_PENDING, SYSERR_EVENT_GEN_FAILURE, SYSERR_CONFIG_PARAM_MISSING, SYSERR_RANGE, SYSERR_BAD_REQUEST.... Field introduced in 18.2.6.
	// Read Only: true
	Status *string `json:"status,omitempty"`

	//  It is a reference to an object of type Tenant. Field introduced in 18.2.6.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// Unique object identifier of the object.
	UUID *string `json:"uuid,omitempty"`
}
