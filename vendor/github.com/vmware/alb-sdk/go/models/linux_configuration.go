package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// LinuxConfiguration linux configuration
// swagger:model LinuxConfiguration
type LinuxConfiguration struct {

	// Banner displayed before login to ssh, and UI.
	Banner *string `json:"banner,omitempty"`

	// Enforce CIS benchmark recommendations for Avi Controller and Service Engines. The enforcement is as per CIS DIL 1.0.1 level 2, for applicable controls. Field introduced in 17.2.8.
	CisMode *bool `json:"cis_mode,omitempty"`

	// Message of the day, shown to users on login via the command line interface, web interface, or ssh.
	Motd *string `json:"motd,omitempty"`
}
