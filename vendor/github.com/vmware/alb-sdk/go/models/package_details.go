package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// PackageDetails package details
// swagger:model PackageDetails
type PackageDetails struct {

	// This contains build related information. Field introduced in 18.2.6.
	Build *BuildInfo `json:"build,omitempty"`

	// MD5 checksum over the entire package. Field introduced in 18.2.6.
	Hash *string `json:"hash,omitempty"`

	// Patch related necessary information. Field introduced in 18.2.6.
	Patch *PatchInfo `json:"patch,omitempty"`

	// Path of the package in the repository. Field introduced in 18.2.6.
	Path *string `json:"path,omitempty"`
}
