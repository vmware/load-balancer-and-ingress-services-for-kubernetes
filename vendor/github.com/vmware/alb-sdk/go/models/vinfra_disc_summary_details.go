package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// VinfraDiscSummaryDetails vinfra disc summary details
// swagger:model VinfraDiscSummaryDetails
type VinfraDiscSummaryDetails struct {

	// Number of num_clusters.
	NumClusters *int64 `json:"num_clusters,omitempty"`

	// Number of num_dcs.
	NumDcs *int64 `json:"num_dcs,omitempty"`

	// Number of num_hosts.
	NumHosts *int64 `json:"num_hosts,omitempty"`

	// Number of num_nws.
	NumNws *int64 `json:"num_nws,omitempty"`

	// Number of num_vms.
	NumVms *int64 `json:"num_vms,omitempty"`

	// vcenter of VinfraDiscSummaryDetails.
	// Required: true
	Vcenter *string `json:"vcenter"`
}
