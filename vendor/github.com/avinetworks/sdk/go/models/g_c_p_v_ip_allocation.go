package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// GCPVIPAllocation g c p v IP allocation
// swagger:model GCPVIPAllocation
type GCPVIPAllocation struct {

	// Configure Google Cloud Internal LoadBalancer for VIP. The VIP will be auto allocated from a Google Cloud VPC Subnet. Field introduced in 20.1.1.
	Ilb *GCPVIPILB `json:"ilb,omitempty"`

	// VIP Allocation Mode. Enum options - ROUTES, ILB. Field introduced in 20.1.1.
	// Required: true
	Mode *string `json:"mode"`

	// Configure Google Cloud VPC Routes for VIP. The VIP can either be a static IP or auto allocted from AVI Internal Network. The VIP should not overlap with any of the subnet ranges in Google Cloud VPC. Field introduced in 20.1.1.
	Routes *GCPVIPRoutes `json:"routes,omitempty"`
}
