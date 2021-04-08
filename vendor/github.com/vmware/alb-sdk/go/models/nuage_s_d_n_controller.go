package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// NuageSDNController nuage s d n controller
// swagger:model NuageSDNController
type NuageSDNController struct {

	// nuage_organization of NuageSDNController.
	NuageOrganization *string `json:"nuage_organization,omitempty"`

	// nuage_password of NuageSDNController.
	NuagePassword *string `json:"nuage_password,omitempty"`

	// Number of nuage_port.
	NuagePort *int32 `json:"nuage_port,omitempty"`

	// nuage_username of NuageSDNController.
	NuageUsername *string `json:"nuage_username,omitempty"`

	// Nuage VSD host name or IP address.
	NuageVsdHost *string `json:"nuage_vsd_host,omitempty"`

	// Domain to be used for SE creation.
	SeDomain *string `json:"se_domain,omitempty"`

	// Enterprise to be used for SE creation.
	SeEnterprise *string `json:"se_enterprise,omitempty"`

	// Network to be used for SE creation.
	SeNetwork *string `json:"se_network,omitempty"`

	// Policy Group to be used for SE creation.
	SePolicyGroup *string `json:"se_policy_group,omitempty"`

	// User to be used for SE creation.
	SeUser *string `json:"se_user,omitempty"`

	// Zone to be used for SE creation.
	SeZone *string `json:"se_zone,omitempty"`
}
