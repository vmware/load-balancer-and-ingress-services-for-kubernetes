package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SummarizedSubnetInfo summarized subnet info
// swagger:model SummarizedSubnetInfo
type SummarizedSubnetInfo struct {

	// cidr_prefix of SummarizedSubnetInfo.
	// Required: true
	CidrPrefix *string `json:"cidr_prefix"`

	// network of SummarizedSubnetInfo.
	// Required: true
	Network *string `json:"network"`
}
