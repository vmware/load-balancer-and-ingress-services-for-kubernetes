package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SSHSeDeployment SSH se deployment
// swagger:model SSHSeDeployment
type SSHSeDeployment struct {

	// Host OS distribution e.g. COREOS, UBUNTU, REDHAT. Field deprecated in 17.1.1.
	HostOs *string `json:"host_os,omitempty"`

	// Password for ssh and/or sudo. Field deprecated in 17.1.1.
	Password *string `json:"password,omitempty"`

	// Username for SSH access to hosts. Field deprecated in 17.1.1.
	SSHUser *string `json:"ssh_user,omitempty"`

	// Username for sudo. Field deprecated in 17.1.1.
	SudoUser *string `json:"sudo_user,omitempty"`
}
