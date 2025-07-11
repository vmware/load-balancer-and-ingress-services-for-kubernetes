// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// DockerConfiguration docker configuration
// swagger:model DockerConfiguration
type DockerConfiguration struct {

	// Sync frequency in seconds with frameworks. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AppSyncFrequency *uint32 `json:"app_sync_frequency,omitempty"`

	// UUID of the UCP CA TLS cert and key. It is a reference to an object of type SSLKeyAndCertificate. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CaTLSKeyAndCertificateRef *string `json:"ca_tls_key_and_certificate_ref,omitempty"`

	// UUID of the client TLS cert and key. It is a reference to an object of type SSLKeyAndCertificate. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ClientTLSKeyAndCertificateRef *string `json:"client_tls_key_and_certificate_ref,omitempty"`

	// Perform container port matching to create a HTTP Virtualservice instead of a TCP/UDP VirtualService. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ContainerPortMatchHTTPService *bool `json:"container_port_match_http_service,omitempty"`

	// Directory to mount to check for core dumps on Service Engines. This will be mapped read only to /var/crash on any new Service Engines. This is a disruptive change. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CoredumpDirectory *string `json:"coredump_directory,omitempty"`

	// Disable auto service sync for back end services. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DisableAutoBackendServiceSync *bool `json:"disable_auto_backend_service_sync,omitempty"`

	// Disable auto service sync for front end services. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DisableAutoFrontendServiceSync *bool `json:"disable_auto_frontend_service_sync,omitempty"`

	// Disable SE creation. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DisableAutoSeCreation *bool `json:"disable_auto_se_creation,omitempty"`

	// Docker registry for ServiceEngine image. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DockerRegistrySe *DockerRegistry `json:"docker_registry_se,omitempty"`

	// Match against this prefix when placing east-west VSs on SEs . Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	EastWestPlacementSubnet *IPAddrPrefix `json:"east_west_placement_subnet,omitempty"`

	// Enable Docker event subscription. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	EnableEventSubscription *bool `json:"enable_event_subscription,omitempty"`

	// For Front End proxies, use container port as service port. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	FeproxyContainerPortAsService *bool `json:"feproxy_container_port_as_service,omitempty"`

	// Enable proxy ARP from Host interface for Front End  proxies. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	FeproxyVipsEnableProxyArp *bool `json:"feproxy_vips_enable_proxy_arp,omitempty"`

	// Optional fleet remote endpoint if fleet is used for SE deployment. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	FleetEndpoint *string `json:"fleet_endpoint,omitempty"`

	// List of container ports that create a HTTP Virtualservice instead of a TCP/UDP VirtualService. Defaults to 80. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	HTTPContainerPorts []int64 `json:"http_container_ports,omitempty,omitempty"`

	// Use Fleet/SSH for SE deployment. Enum options - SE_CREATE_FLEET, SE_CREATE_SSH, SE_CREATE_POD. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeDeploymentMethod *string `json:"se_deployment_method,omitempty"`

	// Exclude hosts with attributes for SE creation. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeExcludeAttributes []*MesosAttribute `json:"se_exclude_attributes,omitempty"`

	// Create SEs just on hosts with include attributes. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeIncludeAttributes []*MesosAttribute `json:"se_include_attributes,omitempty"`

	// New SE spawn rate per minute. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeSpawnRate *uint32 `json:"se_spawn_rate,omitempty"`

	// Host volume to be used as a disk for Avi SE, This is a disruptive change. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeVolume *string `json:"se_volume,omitempty"`

	// Make service ports accessible on all Host interfaces in addition to East-West VIP and/or bridge IP. Usually enabled AWS clusters to export East-West services on Host interface. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ServicesAccessibleAllInterfaces *bool `json:"services_accessible_all_interfaces,omitempty"`

	// Cloud connector user uuid for SSH to hosts. It is a reference to an object of type CloudConnectorUser. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SSHUserRef *string `json:"ssh_user_ref,omitempty"`

	// List of Docker UCP nodes; In case of a load balanced UCP cluster, use Virtual IP of the cluster. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UcpNodes []string `json:"ucp_nodes,omitempty"`

	// Use container IP address port for pool instead of host IP address hostport. This mode is applicable if the container IP is reachable (not a private NATed IP) from other hosts in a routed environment for containers. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UseContainerIPPort *bool `json:"use_container_ip_port,omitempty"`

	// If true, use controller generated SE docker image via fileservice, else use docker repository image as defined by docker_registry_se. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UseControllerImage *bool `json:"use_controller_image,omitempty"`
}
