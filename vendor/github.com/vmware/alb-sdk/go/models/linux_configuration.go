// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// LinuxConfiguration linux configuration
// swagger:model LinuxConfiguration
type LinuxConfiguration struct {

	// Banner displayed before login to ssh, and UI. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Banner *string `json:"banner,omitempty"`

	// Enforce CIS benchmark recommendations for Avi Controller and Service Engines. The enforcement is as per CIS DIL 1.0.1 level 2, for applicable controls. Field introduced in 17.2.8. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CisMode *bool `json:"cis_mode,omitempty"`

	// Message of the day, shown to users on login via the command line interface, web interface, or ssh. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Motd *string `json:"motd,omitempty"`
}
