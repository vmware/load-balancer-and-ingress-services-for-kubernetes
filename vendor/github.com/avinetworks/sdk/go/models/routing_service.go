package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// RoutingService routing service
// swagger:model RoutingService
type RoutingService struct {

	// Advertise reachability of backend server networks via ADC through BGP for default gateway feature. Field introduced in 18.2.5.
	AdvertiseBackendNetworks *bool `json:"advertise_backend_networks,omitempty"`

	// Service Engine acts as Default Gateway for this service. Field introduced in 18.2.5.
	EnableRouting *bool `json:"enable_routing,omitempty"`

	// Enable VIP on all interfaces of this service. Field introduced in 18.2.5.
	EnableVipOnAllInterfaces *bool `json:"enable_vip_on_all_interfaces,omitempty"`

	// Use Virtual MAC address for interfaces on which floating interface IPs are placed. Field introduced in 18.2.5.
	EnableVMAC *bool `json:"enable_vmac,omitempty"`

	// Floating Interface IPs for the RoutingService. Field introduced in 18.2.5.
	FloatingIntfIP []*IPAddr `json:"floating_intf_ip,omitempty"`

	// If ServiceEngineGroup is configured for Legacy 1+1 Active Standby HA Mode, Floating IP's will be advertised only by the Active SE in the Pair. Virtual Services in this group must be disabled/enabled for any changes to the Floating IP's to take effect. Only active SE hosting VS tagged with Active Standby SE 2 Tag will advertise this floating IP when manual load distribution is enabled. Field introduced in 18.2.5.
	FloatingIntfIPSe2 []*IPAddr `json:"floating_intf_ip_se_2,omitempty"`

	// Routing Service related Flow profile information. Field introduced in 18.2.5.
	FlowtableProfile *FlowtableProfile `json:"flowtable_profile,omitempty"`

	// Enable graceful restart feature in routing service. For example, BGP. Field introduced in 18.2.6.
	GracefulRestart *bool `json:"graceful_restart,omitempty"`

	// NAT policy for outbound NAT functionality. This is done in post-routing. It is a reference to an object of type NatPolicy. Field introduced in 18.2.5.
	NatPolicyRef *string `json:"nat_policy_ref,omitempty"`

	// For IP Routing feature, enabling this knob will fallback to routing through Linux, by default routing is done via Service Engine data-path. Field introduced in 18.2.5.
	RoutingByLinuxIpstack *bool `json:"routing_by_linux_ipstack,omitempty"`
}
