package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// DebugVsDataplane debug vs dataplane
// swagger:model DebugVsDataplane
type DebugVsDataplane struct {

	//  Enum options - DEBUG_VS_TCP_CONNECTION. DEBUG_VS_TCP_PKT. DEBUG_VS_TCP_APP. DEBUG_VS_TCP_APP_PKT. DEBUG_VS_TCP_RETRANSMIT. DEBUG_VS_TCP_TIMER. DEBUG_VS_TCP_CONN_ERROR. DEBUG_VS_TCP_PKT_ERROR. DEBUG_VS_TCP_REXMT. DEBUG_VS_TCP_ALL. DEBUG_VS_CREDIT. DEBUG_VS_PROXY_CONNECTION. DEBUG_VS_PROXY_PKT. DEBUG_VS_PROXY_ERR. DEBUG_VS_UDP. DEBUG_VS_UDP_PKT. DEBUG_VS_HM. DEBUG_VS_HM_ERR. DEBUG_VS_HM_PKT. DEBUG_VS_HTTP_CORE...
	// Required: true
	Flag *string `json:"flag"`
}
