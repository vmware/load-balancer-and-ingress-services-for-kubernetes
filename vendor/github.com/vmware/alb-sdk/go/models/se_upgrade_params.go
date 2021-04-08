package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SeUpgradeParams se upgrade params
// swagger:model SeUpgradeParams
type SeUpgradeParams struct {

	// This field is used to disable scale-in/scale out operations during upgrade operations. .
	Disruptive *bool `json:"disruptive,omitempty"`

	//  Field deprecated in 18.2.10, 20.1.1.
	Force *bool `json:"force,omitempty"`

	// Upgrade System with patch upgrade. Field introduced in 17.2.2.
	Patch *bool `json:"patch,omitempty"`

	// Rollback System with patch upgrade.
	PatchRollback *bool `json:"patch_rollback,omitempty"`

	// Resume from suspended state.
	ResumeFromSuspend *bool `json:"resume_from_suspend,omitempty"`

	// It is used in rollback operations. .
	Rollback *bool `json:"rollback,omitempty"`

	//  It is a reference to an object of type ServiceEngineGroup. Field introduced in 17.2.2.
	SeGroupRefs []string `json:"se_group_refs,omitempty"`

	// When set, this will skip upgrade on the Service Engine which is upgrade suspended state.
	SkipSuspended *bool `json:"skip_suspended,omitempty"`

	// When set to true, if there is any failure during the SE upgrade, upgrade will be suspended for this SE group and manual intervention would be needed to resume the upgrade. Field introduced in 17.1.4.
	SuspendOnFailure *bool `json:"suspend_on_failure,omitempty"`

	//  Field deprecated in 18.2.10, 20.1.1.
	Test *bool `json:"test,omitempty"`

	// Unique object identifier of the object.
	UUID *string `json:"uuid,omitempty"`
}
