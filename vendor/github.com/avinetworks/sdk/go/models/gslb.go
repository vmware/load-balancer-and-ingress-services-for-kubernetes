package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// Gslb gslb
// swagger:model Gslb
type Gslb struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// Frequency with which messages are propagated to Vs Mgr. Value of 0 disables async behavior and RPC are sent inline. Allowed values are 0-5. Field introduced in 18.2.3.
	AsyncInterval *int32 `json:"async_interval,omitempty"`

	// Max retries after which the remote site is treated as a fresh start. In fresh start all the configs are downloaded. Allowed values are 1-1024.
	ClearOnMaxRetries *int32 `json:"clear_on_max_retries,omitempty"`

	// Group to specify if the client ip addresses are public or private. Field introduced in 17.1.2.
	ClientIPAddrGroup *GslbClientIPAddrGroup `json:"client_ip_addr_group,omitempty"`

	// User defined description for the object.
	Description *string `json:"description,omitempty"`

	// Sub domain configuration for the GSLB.  GSLB service's FQDN must be a match one of these subdomains. .
	DNSConfigs []*DNSConfig `json:"dns_configs,omitempty"`

	// Frequency with which errored messages are resynced to follower sites. Value of 0 disables resync behavior. Allowed values are 60-3600. Special values are 0 - 'Disable'. Field introduced in 18.2.3.
	ErrorResyncInterval *int32 `json:"error_resync_interval,omitempty"`

	// This field indicates that this object is replicated across GSLB federation. Field introduced in 17.1.3.
	IsFederated *bool `json:"is_federated,omitempty"`

	// Mark this Site as leader of GSLB configuration. This site is the one among the Avi sites.
	// Required: true
	LeaderClusterUUID *string `json:"leader_cluster_uuid"`

	// This field disables the configuration operations on the leader for all federated objects.  CUD operations on Gslb, GslbService, GslbGeoDbProfile and other federated objects will be rejected. The rest-api disabling helps in upgrade scenarios where we don't want configuration sync operations to the Gslb member when the member is being upgraded.  This configuration programmatically blocks the leader from accepting new Gslb configuration when member sites are undergoing upgrade. . Field introduced in 17.2.1.
	MaintenanceMode *bool `json:"maintenance_mode,omitempty"`

	// Name for the GSLB object.
	// Required: true
	Name *string `json:"name"`

	// Frequency with which group members communicate. Allowed values are 1-3600.
	SendInterval *int32 `json:"send_interval,omitempty"`

	// The user can specify a send-interval while entering maintenance mode. The validity of this 'maintenance send-interval' is only during maintenance mode. When the user leaves maintenance mode, the original send-interval is reinstated. This internal variable is used to store the original send-interval. . Field introduced in 18.2.3.
	SendIntervalPriorToMaintenanceMode *int32 `json:"send_interval_prior_to_maintenance_mode,omitempty"`

	// Select Avi site member belonging to this Gslb.
	Sites []*GslbSite `json:"sites,omitempty"`

	//  It is a reference to an object of type Tenant.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// Third party site member belonging to this Gslb. Field introduced in 17.1.1.
	ThirdPartySites []*GslbThirdPartySite `json:"third_party_sites,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// UUID of the GSLB object.
	UUID *string `json:"uuid,omitempty"`

	// The view-id is used in change-leader mode to differentiate partitioned groups while they have the same GSLB namespace. Each partitioned group will be able to operate independently by using the view-id.
	ViewID *int64 `json:"view_id,omitempty"`
}
