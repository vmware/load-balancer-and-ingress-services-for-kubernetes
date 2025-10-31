// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// MesosConfiguration mesos configuration
// swagger:model MesosConfiguration
type MesosConfiguration struct {

	// Consider all Virtualservices as Front End Proxies. Front End proxies are placed on specific SEs as opposed to Back End proxies placed on all SEs. Applicable where each service has its own VIP and VIP is reachable from anywhere. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AllVsesAreFeproxy *bool `json:"all_vses_are_feproxy,omitempty"`

	// Sync frequency in seconds with frameworks. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AppSyncFrequency *uint32 `json:"app_sync_frequency,omitempty"`

	// Perform container port matching to create a HTTP Virtualservice instead of a TCP/UDP VirtualService. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ContainerPortMatchHTTPService *bool `json:"container_port_match_http_service,omitempty"`

	// Directory to mount to check for core dumps on Service Engines. This will be mapped read only to /var/crash on any new Service Engines. This is a disruptive change. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CoredumpDirectory *string `json:"coredump_directory,omitempty"`

	// Disable auto service sync for back end services. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DisableAutoBackendServiceSync *bool `json:"disable_auto_backend_service_sync,omitempty"`

	// Disable auto service sync for front end services. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DisableAutoFrontendServiceSync *bool `json:"disable_auto_frontend_service_sync,omitempty"`

	// Disable auto sync for GSLB services. Field introduced in 17.1.2. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DisableAutoGsSync *bool `json:"disable_auto_gs_sync,omitempty"`

	// Disable SE creation. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DisableAutoSeCreation *bool `json:"disable_auto_se_creation,omitempty"`

	// Docker registry for ServiceEngine image. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DockerRegistrySe *DockerRegistry `json:"docker_registry_se,omitempty"`

	// Match against this prefix when placing east-west VSs on SEs (Mesos mode only). Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	EastWestPlacementSubnet *IPAddrPrefix `json:"east_west_placement_subnet,omitempty"`

	// Enable Marathon event subscriptions. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	EnableEventSubscription *bool `json:"enable_event_subscription,omitempty"`

	// Name of second Linux bridge on Host providing connectivity for Front End proxies. This is a disruptive change. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	FeproxyBridgeName *string `json:"feproxy_bridge_name,omitempty"`

	// For Front End proxies, use container port as service port. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	FeproxyContainerPortAsService *bool `json:"feproxy_container_port_as_service,omitempty"`

	// Publish ECMP route to upstream router for VIP. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	FeproxyRoutePublish *FeProxyRoutePublishConfig `json:"feproxy_route_publish,omitempty"`

	// Enable proxy ARP from Host interface for Front End  proxies. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	FeproxyVipsEnableProxyArp *bool `json:"feproxy_vips_enable_proxy_arp,omitempty"`

	// Optional fleet remote endpoint if fleet is used for SE deployment. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	FleetEndpoint *string `json:"fleet_endpoint,omitempty"`

	// List of container ports that create a HTTP Virtualservice instead of a TCP/UDP VirtualService. Defaults to 80. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	HTTPContainerPorts []int64 `json:"http_container_ports,omitempty,omitempty"`

	// List of Marathon frameworks. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MarathonConfigurations []*MarathonConfiguration `json:"marathon_configurations,omitempty"`

	// Options for Marathon SE deployment. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MarathonSeDeployment *MarathonSeDeployment `json:"marathon_se_deployment,omitempty"`

	// Mesos URL of the form http //host port. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MesosURL *string `json:"mesos_url,omitempty"`

	// Mesos Node label to be used as Mesos Node's availability zone in a dual availability zone deployment. ServiceEngines belonging to the availability zone will be rebooted during a manual DR failover. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NodeAvailabilityZoneLabel *string `json:"node_availability_zone_label,omitempty"`

	// Nuage Overlay SDN Controller information. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NuageController *NuageSDNController `json:"nuage_controller,omitempty"`

	// Use Fleet/SSH for deploying Service Engines. Enum options - MESOS_SE_CREATE_FLEET, MESOS_SE_CREATE_SSH, MESOS_SE_CREATE_MARATHON. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeDeploymentMethod *string `json:"se_deployment_method,omitempty"`

	// Exclude hosts with attributes for SE creation. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeExcludeAttributes []*MesosAttribute `json:"se_exclude_attributes,omitempty"`

	// Create SEs just on hosts with include attributes. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeIncludeAttributes []*MesosAttribute `json:"se_include_attributes,omitempty"`

	// Obsolete - ignored. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeResources []*MesosSeResources `json:"se_resources,omitempty"`

	// New SE spawn rate per minute. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeSpawnRate *uint32 `json:"se_spawn_rate,omitempty"`

	// Host volume to be used as a disk for Avi SE, This is a disruptive change. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeVolume *string `json:"se_volume,omitempty"`

	// Make service ports accessible on all Host interfaces in addition to East-West VIP and/or bridge IP. Usually enabled AWS Mesos clusters to export East-West services on Host interface. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ServicesAccessibleAllInterfaces *bool `json:"services_accessible_all_interfaces,omitempty"`

	// Cloud connector user uuid for SSH to hosts. It is a reference to an object of type CloudConnectorUser. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SSHUserRef *string `json:"ssh_user_ref,omitempty"`

	// Use Bridge IP on each Host as VIP. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UseBridgeIPAsVip *bool `json:"use_bridge_ip_as_vip,omitempty"`

	// Use container IP address port for pool instead of host IP address hostport. This mode is applicable if the container IP is reachable (not a private NATed IP) from other hosts in a routed environment for containers. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UseContainerIPPort *bool `json:"use_container_ip_port,omitempty"`

	// If true, use controller generated SE docker image via fileservice, else use docker repository image as defined by docker_registry_se. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UseControllerImage *bool `json:"use_controller_image,omitempty"`

	// Use unique virtual IP address for every east west service in Mesos/Marathon. 'use_bridge_ip_as_vip' and 'vip' fields , if set, will not be used if this field is set. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UseVipsForEastWestServices *bool `json:"use_vips_for_east_west_services,omitempty"`

	// VIP to be used by all East-West apps on all Hosts. Preferrably use an address from outside the subnet. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Vip *IPAddr `json:"vip,omitempty"`
}
