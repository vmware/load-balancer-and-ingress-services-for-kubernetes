// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// OShiftK8SConfiguration o shift k8 s configuration
// swagger:model OShiftK8SConfiguration
type OShiftK8SConfiguration struct {

	// Sync frequency in seconds with frameworks. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AppSyncFrequency *uint32 `json:"app_sync_frequency,omitempty"`

	// Auto assign FQDN to a virtual service if a valid FQDN is not configured. Field introduced in 17.2.8. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AutoAssignFqdn *bool `json:"auto_assign_fqdn,omitempty"`

	// Avi Linux bridge subnet on OpenShift/K8s nodes. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AviBridgeSubnet *IPAddrPrefix `json:"avi_bridge_subnet,omitempty"`

	// UUID of the UCP CA TLS cert and key. It is a reference to an object of type SSLKeyAndCertificate. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CaTLSKeyAndCertificateRef *string `json:"ca_tls_key_and_certificate_ref,omitempty"`

	// UUID of the client TLS cert and key instead of service account token. One of client certificate or token is required. It is a reference to an object of type SSLKeyAndCertificate. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ClientTLSKeyAndCertificateRef *string `json:"client_tls_key_and_certificate_ref,omitempty"`

	// Openshift/K8S Cluster ID used to uniquely map same named namespaces as tenants in Avi. In order to use more than one OpenShift/K8S cloud on this controller, cluster_tag has to be unique across these clouds. Changing cluster_tag is disruptive as all virtual services in the cloud will be recreated. Field introduced in 17.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ClusterTag *string `json:"cluster_tag,omitempty"`

	// Perform container port matching to create a HTTP Virtualservice instead of a TCP/UDP VirtualService. By default, ports 80, 8080, 443, 8443 are considered HTTP. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ContainerPortMatchHTTPService *bool `json:"container_port_match_http_service,omitempty"`

	// Directory to mount to check for core dumps on Service Engines. This will be mapped read only to /var/crash on any new Service Engines. This is a disruptive change. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CoredumpDirectory *string `json:"coredump_directory,omitempty"`

	// If there is no explicit east_west_placement field in virtualservice configuration, treat service as a East-West service; default services such a OpenShift API server do not have virtualservice configuration. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DefaultServiceAsEastWestService *bool `json:"default_service_as_east_west_service,omitempty"`

	// Disable auto service sync for back end services. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DisableAutoBackendServiceSync *bool `json:"disable_auto_backend_service_sync,omitempty"`

	// Disable auto service sync for front end services. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DisableAutoFrontendServiceSync *bool `json:"disable_auto_frontend_service_sync,omitempty"`

	// Disable auto sync for GSLB services. Field introduced in 17.1.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DisableAutoGsSync *bool `json:"disable_auto_gs_sync,omitempty"`

	// Disable SE creation. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DisableAutoSeCreation *bool `json:"disable_auto_se_creation,omitempty"`

	// Host Docker server UNIX socket endpoint. Field introduced in 17.2.14, 18.1.5, 18.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DockerEndpoint *string `json:"docker_endpoint,omitempty"`

	// Docker registry for ServiceEngine image. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DockerRegistrySe *DockerRegistry `json:"docker_registry_se,omitempty"`

	// Match against this prefix when placing east-west VSs on SEs . Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	EastWestPlacementSubnet *IPAddrPrefix `json:"east_west_placement_subnet,omitempty"`

	// Enable Kubernetes event subscription. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	EnableEventSubscription *bool `json:"enable_event_subscription,omitempty"`

	// Knob to turn on adding of HTTP drop rules for host and path combinations in incoming request header, specified as part of Ingress/Route spec. The default state is to enable this behavior. Note  Toggling this knob only affects any new routes/ingresses, existing routes/ingresses present in Avi will continue to function as-is. Field introduced in 18.2.6. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	EnableRouteIngressHardening *bool `json:"enable_route_ingress_hardening,omitempty"`

	// Enable proxy ARP from Host interface for Front End  proxies. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	FeproxyVipsEnableProxyArp *bool `json:"feproxy_vips_enable_proxy_arp,omitempty"`

	// List of container ports that create a HTTP Virtualservice instead of a TCP/UDP VirtualService. Defaults to 80, 8080, 443 and 8443. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	HTTPContainerPorts []int64 `json:"http_container_ports,omitempty,omitempty"`

	// Do not sync applications only for ingress that have these exclude attributes configured. Field introduced in 17.2.15, 18.1.5, 18.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	IngExcludeAttributes []*IngAttribute `json:"ing_exclude_attributes,omitempty"`

	// Sync applications only for ingress objects that have these include attributes configured. Default values are populated for this field if not provided. The default value are  'attribute'  'kubernetes.io/ingress.class', 'value' 'avi'. Field introduced in 17.2.15, 18.1.5, 18.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	IngIncludeAttributes []*IngAttribute `json:"ing_include_attributes,omitempty"`

	// Perform Layer4 (TCP/UDP) health monitoring even for Layer7 (HTTP) Pools. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	L4HealthMonitoring *bool `json:"l4_health_monitoring,omitempty"`

	// List of OpenShift/Kubernetes master nodes; In case of a load balanced OpenShift/K8S cluster, use Virtual IP of the cluster. Each node is of the form node 8443 or http //node 8080. If scheme is not provided, https is assumed. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MasterNodes []string `json:"master_nodes,omitempty"`

	// OpenShift/K8S Node label to be used as OpenShift/K8S Node's availability zone in a dual availability zone deployment. ServiceEngines belonging to the availability zone will be rebooted during a manual DR failover. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NodeAvailabilityZoneLabel *string `json:"node_availability_zone_label,omitempty"`

	// Syncing of applications is disabled only for namespaces/projects that have these exclude attributes configured. If there are apps synced already for these namespaces, they will be removed from Avi. Field introduced in 17.1.9,17.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NsExcludeAttributes []*MesosAttribute `json:"ns_exclude_attributes,omitempty"`

	// Sync applications only for namespaces/projects that have these include attributes configured. Field introduced in 17.1.9,17.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NsIncludeAttributes []*MesosAttribute `json:"ns_include_attributes,omitempty"`

	// Enables sharding of Routes and Ingresses to this number (if non zero) of virtual services in the admin tenant per SEGroup. Sharding is done by hashing on the namespace of the Ingress/Route object. This knob is valid only if shared_virtualservice_namespace flag is set. Field introduced in 18.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumShards *uint32 `json:"num_shards,omitempty"`

	// Override Service Ports with well known ports (80/443) for http/https Route/Ingress VirtualServices. Field introduced in 17.2.12,18.1.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	OverrideServicePorts *bool `json:"override_service_ports,omitempty"`

	// Persistent Volume Claim name to be used for persistent storage for Avi service engines. This could be used in scenarios where host based volumes are ephemeral. Refer https //kubernetes.io/docs/concepts/storage/persistent-volumes/#persistentvolumeclaims for more details on the usage of this field. Field introduced in 18.2.6. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PersistentVolumeClaim *string `json:"persistent_volume_claim,omitempty"`

	// Routes support adding routes to a particular namespace routing table in Openshift/K8s cluster. Each route is a combination of subnet and nexthop ip address or nexthop interface name, and a enum type is used to distinguish an entry in the host (default behaviour) or in the container/pod or in other namespace. This knob should be enabled in the following cases  1. Forwarding the network packets to the same network interface from where it came from in the OpenShift/K8s node. 2. OpenShift/K8s Node connected to the Internet via multiple network interfaces on different networks/ISPs.3. Handling North-South traffic originating from with in the node when the default gateway for outgoing traffic of vs is configured.4. Handling the container/pod traffic by adding the routes in the container/pod. Field introduced in 18.2.6. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Routes []*RouteInfo `json:"routes,omitempty"`

	// Cluster uses overlay based SDN. Enable this flag if cluster uses a overlay based SDN for OpenShift, Flannel, Weave, Nuage. Disable for routed mode. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SdnOverlay *bool `json:"sdn_overlay,omitempty"`

	// Use SSH/Pod for SE deployment. Enum options - SE_CREATE_FLEET, SE_CREATE_SSH, SE_CREATE_POD. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeDeploymentMethod *string `json:"se_deployment_method,omitempty"`

	// Exclude hosts with attributes for SE creation. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeExcludeAttributes []*MesosAttribute `json:"se_exclude_attributes,omitempty"`

	// OpenShift/K8S secret name to be used for private docker repos when deploying SE as a Pod. Reference Link  https //kubernetes.io/docs/tasks/configure-pod-container/pull-image-private-registry/. Field introduced in 17.2.13,18.1.3,18.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeImagePullSecret *string `json:"se_image_pull_secret,omitempty"`

	// Create SEs just on hosts with include attributes. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeIncludeAttributes []*MesosAttribute `json:"se_include_attributes,omitempty"`

	// Kubernetes namespace to be used for deploying Avi service engines. This namespace is used to create daemonsets, service accounts, etc. for Avi only use. Setting this value is a disruptive operation and assumes the namespace exists in kubernetes. 'default' namespace is picked if this field is unset. Field introduced in 18.2.6. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeNamespace *string `json:"se_namespace,omitempty"`

	// Match SE Pod tolerations against taints of OpenShift/K8S nodes https //kubernetes.io/docs/concepts/configuration/taint-and-toleration/. Field introduced in 17.2.14, 18.1.5, 18.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SePodTolerations []*PodToleration `json:"se_pod_tolerations,omitempty"`

	// Priority class for AVI SEs when running as pods. User is expected to have the priority class (with this name) created in Kubernetes, beforehand. If the priority class doesn't exist while assigning this field, the SE pods may not start. If empty no priority class will be used for deploying SE pods (default behaviour). Field introduced in 18.2.6. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SePriorityClass *string `json:"se_priority_class,omitempty"`

	// Restart ServiceEngines by batch on ServiceEngineGroup updates (cpu, memory..etc). Field introduced in 17.2.15, 18.1.5, 18.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeRestartBatchSize *uint32 `json:"se_restart_batch_size,omitempty"`

	// Restart ServiceEngines forcely if VirtualServices failed to migrate to another SE. Field introduced in 17.2.15, 18.1.5, 18.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeRestartForce *bool `json:"se_restart_force,omitempty"`

	// Host volume to be used as a disk for Avi SE, This is a disruptive change. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeVolume *string `json:"se_volume,omitempty"`

	// Allow Avi Vantage to create SecurityContextConstraints and ServiceAccounts which allow Egress Pods to run in privileged mode in an Openshift environment. Enabling this would exclude egress services from 'disable_auto_backend_service_sync' (if set) behaviour. Note  Access credentials must have cluster-admin role privileges. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SecureEgressMode *bool `json:"secure_egress_mode,omitempty"`

	// Authorization token for service account instead of client certificate. One of client certificate or token is required. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ServiceAccountToken *string `json:"service_account_token,omitempty"`

	// Prefix to be used for Shard VS name when num_shards knob is non zero. Format for Shard VS name will be <shard_prefix>-<idx>-CloudName-SEGroupName. Field introduced in 18.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ShardPrefix *string `json:"shard_prefix,omitempty"`

	// Projects/Namespaces use a shared virtualservice for http/https Routes and Ingress objects unless overriden by the avi_virtualservice  dedicated|shared annotation. Field introduced in 17.1.9,17.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SharedVirtualserviceNamespace *bool `json:"shared_virtualservice_namespace,omitempty"`

	// Cloud connector user uuid for SSH to hosts. It is a reference to an object of type CloudConnectorUser. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SSHUserRef *string `json:"ssh_user_ref,omitempty"`

	// Allow the not_ready_addresses in the kubernetes endpoint object to be added as servers in the AVI pool object. Field introduced in 18.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SyncNotReadyAddresses *bool `json:"sync_not_ready_addresses,omitempty"`

	// If true, use controller generated SE docker image via fileservice, else use docker repository image as defined by docker_registry_se. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UseControllerImage *bool `json:"use_controller_image,omitempty"`

	// Use OpenShift/Kubernetes resource definition and annotations as single-source-of-truth. Any changes made in Avi Controller via UI or CLI will be overridden by values provided in annotations. Field introduced in 17.2.13, 18.1.4, 18.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UseResourceDefinitionAsSsot *bool `json:"use_resource_definition_as_ssot,omitempty"`

	// Enable VirtualService placement on Service Engines on nodes with scheduling disabled. When false, Service Engines are disabled on nodes where scheduling is disabled. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UseSchedulingDisabledNodes *bool `json:"use_scheduling_disabled_nodes,omitempty"`

	// Use Cluster IP of service as VIP for East-West services; This option requires that kube proxy is disabled on all nodes. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UseServiceClusterIPAsEwVip *bool `json:"use_service_cluster_ip_as_ew_vip,omitempty"`

	// VirtualService default gateway if multiple nics are present in the host. Field introduced in 18.2.2. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VipDefaultGateway *IPAddr `json:"vip_default_gateway,omitempty"`
}
