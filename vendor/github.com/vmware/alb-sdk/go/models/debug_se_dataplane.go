package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// DebugSeDataplane debug se dataplane
// swagger:model DebugSeDataplane
type DebugSeDataplane struct {

	//  Enum options - DEBUG_DISPATCHER_FLOW. DEBUG_DISPATCHER_FLOW_DETAIL. DEBUG_DISPATCHER_FLOW_ALL. DEBUG_CONFIG. DEBUG_IP. DEBUG_IP_PKT_IN. DEBUG_IP_PKT_OUT. DEBUG_ARP. DEBUG_ARP_PKT_IN. DEBUG_ARP_PKT_OUT. DEBUG_ETHERNET. DEBUG_ETHERNET_PKT_IN. DEBUG_ETHERNET_PKT_OUT. DEBUG_ICMP. DEBUG_PCAP_RX. DEBUG_PCAP_TX. DEBUG_PCAP_DROP. DEBUG_PCAP_ALL. DEBUG_MISC. DEBUG_CRUD...
	// Required: true
	Flag *string `json:"flag"`
}
