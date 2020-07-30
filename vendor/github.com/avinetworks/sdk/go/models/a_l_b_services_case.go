package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ALBServicesCase a l b services case
// swagger:model ALBServicesCase
type ALBServicesCase struct {

	//  Field introduced in 18.2.6.
	AssetID *string `json:"asset_id,omitempty"`

	//  Field introduced in 18.2.6.
	CaseAttachments []*ALBServicesCaseAttachment `json:"case_attachments,omitempty"`

	//  Field introduced in 18.2.6.
	CaseCreatedBy *string `json:"case_created_by,omitempty"`

	//  Field introduced in 18.2.6.
	CaseNumber *string `json:"case_number,omitempty"`

	//  Field introduced in 18.2.6.
	CaseStatus *string `json:"case_status,omitempty"`

	// Contact information associated to particular case. Field introduced in 20.1.1.
	ContactInfo *ALBServicesUser `json:"contact_info,omitempty"`

	//  Field introduced in 18.2.6.
	CreatedDate *string `json:"created_date,omitempty"`

	//  Field introduced in 18.2.6.
	CustomTag *string `json:"custom_tag,omitempty"`

	//  Field introduced in 18.2.6.
	DeploymentEnvironment *string `json:"deployment_environment,omitempty"`

	//  Field introduced in 18.2.6.
	Description *string `json:"description,omitempty"`

	// Email of the point of contact for a particular support case. Field introduced in 20.1.1.
	Email *string `json:"email,omitempty"`

	//  Field introduced in 18.2.6.
	Environment *string `json:"environment,omitempty"`

	// Business justification for a feature request. Field introduced in 20.1.1.
	FrBusinessJustification *string `json:"fr_business_justification,omitempty"`

	// Current solution/workaround for a feature request. Field introduced in 20.1.1.
	FrCurrentSolution *string `json:"fr_current_solution,omitempty"`

	// Expected date of delivery for a feature request in YYYY-MM-DD format. Field introduced in 20.1.1.
	FrTiming *string `json:"fr_timing,omitempty"`

	// Possible use cases for a feature request. Field introduced in 20.1.1.
	FrUseCases *string `json:"fr_use_cases,omitempty"`

	//  Field introduced in 18.2.6.
	ID *string `json:"id,omitempty"`

	//  Field introduced in 18.2.6.
	LastModifiedDate *string `json:"last_modified_date,omitempty"`

	//  Field introduced in 18.2.6.
	PatchVersion *string `json:"patch_version,omitempty"`

	//  Field introduced in 18.2.6.
	Severity *string `json:"severity,omitempty"`

	//  Field introduced in 18.2.6.
	Status *string `json:"status,omitempty"`

	//  Field introduced in 18.2.6.
	Subject *string `json:"subject,omitempty"`

	//  Field introduced in 18.2.6.
	Time *string `json:"time,omitempty"`

	//  Field introduced in 18.2.6.
	Type *string `json:"type,omitempty"`

	//  Field introduced in 18.2.6.
	Version *string `json:"version,omitempty"`
}
