package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// VinfraCntlrHostUnreachableList vinfra cntlr host unreachable list
// swagger:model VinfraCntlrHostUnreachableList
type VinfraCntlrHostUnreachableList struct {

	// host_name of VinfraCntlrHostUnreachableList.
	HostName []string `json:"host_name,omitempty"`

	// vcenter of VinfraCntlrHostUnreachableList.
	// Required: true
	Vcenter *string `json:"vcenter"`
}
