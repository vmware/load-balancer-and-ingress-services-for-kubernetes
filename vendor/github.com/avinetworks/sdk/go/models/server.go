package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// Server server
// swagger:model Server
type Server struct {

	// Name of autoscaling group this server belongs to. Field introduced in 17.1.2.
	AutoscalingGroupName *string `json:"autoscaling_group_name,omitempty"`

	// Availability-zone of the server VM.
	AvailabilityZone *string `json:"availability_zone,omitempty"`

	// A description of the Server.
	Description *string `json:"description,omitempty"`

	// (internal-use) Discovered network for this server. This field is deprecated. It is a reference to an object of type Network. Field deprecated in 17.1.1.
	DiscoveredNetworkRef []string `json:"discovered_network_ref,omitempty"`

	// (internal-use) Discovered networks providing reachability for server IP. This field is used internally by Avi, not editable by the user.
	DiscoveredNetworks []*DiscoveredNetwork `json:"discovered_networks,omitempty"`

	// (internal-use) Discovered subnet for this server. This field is deprecated. Field deprecated in 17.1.1.
	DiscoveredSubnet []*IPAddrPrefix `json:"discovered_subnet,omitempty"`

	// Enable, Disable or Graceful Disable determine if new or existing connections to the server are allowed.
	Enabled *bool `json:"enabled,omitempty"`

	// UID of server in external orchestration systems.
	ExternalOrchestrationID *string `json:"external_orchestration_id,omitempty"`

	// UUID identifying VM in OpenStack and other external compute.
	ExternalUUID *string `json:"external_uuid,omitempty"`

	// DNS resolvable name of the server.  May be used in place of the IP address.
	Hostname *string `json:"hostname,omitempty"`

	// IP Address of the server.  Required if there is no resolvable host name.
	// Required: true
	IP *IPAddr `json:"ip"`

	// (internal-use) Geographic location of the server.Currently only for internal usage. Field introduced in 17.1.1.
	Location *GeoLocation `json:"location,omitempty"`

	// MAC address of server.
	MacAddress *string `json:"mac_address,omitempty"`

	// (internal-use) This field is used internally by Avi, not editable by the user. It is a reference to an object of type VIMgrNWRuntime.
	NwRef *string `json:"nw_ref,omitempty"`

	// Optionally specify the servers port number.  This will override the pool's default server port attribute. Allowed values are 1-65535. Special values are 0- 'use backend port in pool'.
	Port *int32 `json:"port,omitempty"`

	// Header value for custom header persistence. .
	PrstHdrVal *string `json:"prst_hdr_val,omitempty"`

	// Ratio of selecting eligible servers in the pool. Allowed values are 1-20.
	Ratio *int32 `json:"ratio,omitempty"`

	// Auto resolve server's IP using DNS name.
	ResolveServerByDNS *bool `json:"resolve_server_by_dns,omitempty"`

	// Rewrite incoming Host Header to server name.
	RewriteHostHeader *bool `json:"rewrite_host_header,omitempty"`

	// Hostname of the node where the server VM or container resides.
	ServerNode *string `json:"server_node,omitempty"`

	// If statically learned.
	Static *bool `json:"static,omitempty"`

	// Verify server belongs to a discovered network or reachable via a discovered network. Verify reachable network isn't the OpenStack management network.
	VerifyNetwork *bool `json:"verify_network,omitempty"`

	// (internal-use) This field is used internally by Avi, not editable by the user. It is a reference to an object of type VIMgrVMRuntime.
	VMRef *string `json:"vm_ref,omitempty"`
}
