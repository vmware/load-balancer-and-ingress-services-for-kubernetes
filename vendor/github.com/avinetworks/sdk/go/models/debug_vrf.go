package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// DebugVrf debug vrf
// swagger:model DebugVrf
type DebugVrf struct {

	//  Enum options - DEBUG_VRF_BGP, DEBUG_VRF_QUAGGA, DEBUG_VRF_ALL, DEBUG_VRF_NONE. Field introduced in 17.1.1.
	// Required: true
	Flag *string `json:"flag"`
}
