package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// AutoScaleLaunchConfig auto scale launch config
// swagger:model AutoScaleLaunchConfig
type AutoScaleLaunchConfig struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// User defined description for the object.
	Description *string `json:"description,omitempty"`

	// Unique ID of the Amazon Machine Image (AMI)  or OpenStack VM ID.
	ImageID *string `json:"image_id,omitempty"`

	// Placeholder for description of property mesos of obj type AutoScaleLaunchConfig field type str  type object
	Mesos *AutoScaleMesosSettings `json:"mesos,omitempty"`

	// Name of the object.
	// Required: true
	Name *string `json:"name"`

	// Placeholder for description of property openstack of obj type AutoScaleLaunchConfig field type str  type object
	Openstack *AutoScaleOpenStackSettings `json:"openstack,omitempty"`

	//  It is a reference to an object of type Tenant.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// If set to True, ServerAutoscalePolicy will use the autoscaling group (external_autoscaling_groups) from Pool to perform scale up and scale down. Pool should have single autoscaling group configured. Field introduced in 17.2.3.
	UseExternalAsg *bool `json:"use_external_asg,omitempty"`

	// Unique object identifier of the object.
	UUID *string `json:"uuid,omitempty"`
}
