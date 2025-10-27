/*
 * Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
 * All Rights Reserved.
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at
*   http://www.apache.org/licenses/LICENSE-2.0
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
*/

package lib

const (
	DISABLE_STATIC_ROUTE_SYNC = "DISABLE_STATIC_ROUTE_SYNC"
	ENABLE_RHI                = "ENABLE_RHI"
	ENABLE_EVH                = "ENABLE_EVH"
	CNI_PLUGIN                = "CNI_PLUGIN"
	CALICO_CNI                = "calico"
	ANTREA_CNI                = "antrea"
	NCP_CNI                   = "ncp"
	OPENSHIFT_CNI             = "openshift"
	OVN_KUBERNETES_CNI        = "ovn-kubernetes"
	CILIUM_CNI                = "cilium"
	INGRESS_API               = "INGRESS_API"
	AviConfigMap              = "avi-k8s-config"
	AviSecret                 = "avi-secret"
	AviInitSecret             = "avi-init-secret"
	VLAN_TRANSPORT_ZONE       = "VLAN"
	OVERLAY_TRANSPORT_ZONE    = "OVERLAY"
	IP_FAMILY                 = "IP_FAMILY"
	AVI_REF                   = "AviRef"

	AVI_INGRESS_CLASS                          = "avi"
	NETWORK_NAME                               = "NETWORK_NAME"
	VIP_NETWORK_LIST                           = "VIP_NETWORK_LIST"
	BGP_PEER_LABELS                            = "BGP_PEER_LABELS"
	SEG_NAME                                   = "SEG_NAME"
	BLOCKED_NS_LIST                            = "BLOCKED_NS_LIST"
	DEFAULT_SE_GROUP                           = "Default-Group"
	NODE_NETWORK_LIST                          = "NODE_NETWORK_LIST"
	NODE_NETWORK_MAX_ENTRIES                   = 5
	DEFAULT_DOMAIN                             = "DEFAULT_DOMAIN"
	ADVANCED_L4                                = "ADVANCED_L4"
	SERVICES_API                               = "SERVICES_API"
	CLUSTER_NAME                               = "CLUSTER_NAME"
	CLUSTER_ID                                 = "CLUSTER_ID"
	CLOUD_VCENTER                              = "CLOUD_VCENTER"
	CLOUD_AZURE                                = "CLOUD_AZURE"
	CLOUD_AWS                                  = "CLOUD_AWS"
	CLOUD_OPENSTACK                            = "CLOUD_OPENSTACK"
	CLOUD_GCP                                  = "CLOUD_GCP"
	CLOUD_NONE                                 = "CLOUD_NONE"
	CLOUD_NSXT                                 = "CLOUD_NSXT"
	DEFAULT_SHARD_SCHEME                       = "hostname"
	HOSTNAME_SHARD_SCHEME                      = "hostname"
	NAMESPACE_SHARD_SCHEME                     = "namespace"
	SLOW_RETRY_LAYER                           = "SlowRetryLayer"
	FAST_RETRY_LAYER                           = "FastRetryLayer"
	NOT_FOUND                                  = "HTTP code: 404"
	STATUS_REDIRECT                            = "HTTP_REDIRECT_STATUS_CODE_302"
	CLOSE_CONNECTION                           = "HTTP_SECURITY_ACTION_CLOSE_CONN"
	IS_IN                                      = "IS_IN"
	SLOW_SYNC_TIME                             = 90 // seconds
	LOG_LEVEL                                  = "logLevel"
	EnableEvents                               = "enableEvents"
	LAYER7_ONLY                                = "layer7Only"
	NO_PG_FOR_SNI                              = "noPGForSNI"
	SERVICE_TYPE                               = "SERVICE_TYPE"
	NODE_PORT                                  = "NodePort"
	NODE_KEY                                   = "NODE_KEY"
	NODE_VALUE                                 = "NODE_VALUE"
	ShardVSSubstring                           = "Shared-"
	ShardVSPrefix                              = "Shared-L7"
	ShardEVHVSPrefix                           = "Shared-L7-EVH-"
	AKOPrefix                                  = "ako-"
	AKOGWPrefix                                = "ako-gw-"
	DedicatedSuffix                            = "-L7-dedicated"
	EVHSuffix                                  = "-EVH"
	PassthroughPrefix                          = "Shared-Passthrough-"
	PolicyAllow                                = "ALLOW"
	PolicyNone                                 = "NONE"
	PolicyEdgeTerm                             = "EDGE"
	PolicyRedirect                             = "REDIRECT"
	PolicyPass                                 = "PASSTHROUGH"
	DeleteConfig                               = "deleteConfig"
	NodePort                                   = "NodePort"
	NodePortLocal                              = "NodePortLocal"
	RouteSecretsPrefix                         = "-route-secret"
	CertTypeVS                                 = "SSL_CERTIFICATE_TYPE_VIRTUALSERVICE"
	CertTypeCA                                 = "SSL_CERTIFICATE_TYPE_CA"
	HostRule                                   = "HostRule"
	HTTPRule                                   = "HTTPRule"
	AviInfraSetting                            = "AviInfraSetting"
	SSORule                                    = "SSORule"
	L4Rule                                     = "L4Rule"
	L7Rule                                     = "L7Rule"
	IstioVirtualService                        = "IstioVirtualService"
	IstioDestinationRule                       = "DestinationRule"
	IstioGateway                               = "IstioGateway"
	MultiClusterIngress                        = "MultiClusterIngress"
	ServiceImport                              = "ServiceImport"
	DummySecret                                = "@avisslkeycertrefdummy"
	DummySecretK8s                             = "@k8ssecretdummy"
	StatusRejected                             = "Rejected"
	StatusAccepted                             = "Accepted"
	AllowedL7ApplicationProfile                = "APPLICATION_PROFILE_TYPE_HTTP"
	AllowedL4ApplicationProfile                = "APPLICATION_PROFILE_TYPE_L4"
	AllowedL4SSLApplicationProfile             = "APPLICATION_PROFILE_TYPE_SSL"
	AllowedTCPProxyNetworkProfileType          = "PROTOCOL_TYPE_TCP_PROXY"
	TypeTLSReencrypt                           = "reencrypt"
	DefaultPoolSSLProfile                      = "System-Standard"
	LB_ALGORITHM_CONSISTENT_HASH_CUSTOM_HEADER = "LB_ALGORITHM_CONSISTENT_HASH_CUSTOM_HEADER"
	LB_ALGORITHM_CONSISTENT_HASH               = "LB_ALGORITHM_CONSISTENT_HASH"
	Gateway                                    = "Gateway"
	GatewayClass                               = "GatewayClass"
	HTTPRoute                                  = "HTTPRoute"
	TCPRoute                                   = "TCPRoute"
	TLSRoute                                   = "TLSRoute"
	UDPRoute                                   = "UDPRoute"
	DuplicateBackends                          = "MultipleBackendsWithSameServiceError"
	HostAlreadyClaimed                         = "Host already Claimed"
	DummyVSForStaleData                        = "DummyVSForStaleData"
	ControllerReqWaitTime                      = 300
	PassthroughInsecure                        = "-insecure"
	AviControllerVSVipIDChangeError            = "Changing an existing VIP's vip_id is not supported"
	AviControllerRecreateVIPError              = "If a new preferred IP is needed, please recreate the VIP"
	ClusterStatusCacheKey                      = "cluster-runtime"
	AviObjDeletionTime                         = 30 // Minutes
	AKOStatefulSet                             = "ako"
	ObjectDeletionStartStatus                  = "Started"
	ObjectDeletionDoneStatus                   = "Done"
	ObjectDeletionTimeoutStatus                = "Timeout"
	DefaultRouteCert                           = "router-certs-default"
	autoAnnotateService                        = "AUTO_ANNOTATE_SERVICE"
	ClusterNameLabelKey                        = "clustername"
	UpdateStatus                               = "UpdateStatus"
	DeleteStatus                               = "DeleteStatus"
	NPLService                                 = "NPLService"
	SyncStatusKey                              = "syncstatus"
	NoFreeIPError                              = "No available free IPs"
	ConfigDisallowedDuringUpgradeError         = "Configuration is disallowed during upgrade"
	NeedToReloadObjectDataVsVip                = "Need to reload object data VsVip"
	VSDataScriptNotFoundError                  = "VSDataScriptSet object not found"
	VSVIPNotFoundError                         = "VsVip object not found"
	DataScript                                 = "Vsdatascript"
	EVHVS                                      = "EVH VirtualService"
	HTTPPS                                     = "HTTPPolicySet"
	HPPMAP                                     = "HTTP Policyset Map"
	HTTPSecurityRule                           = "HTTP Security Rule"
	HTTPRequestRule                            = "HTTP Request Rule"
	HTTPRedirectRule                           = "HTTP Redirect Rule"
	HTTPRewriteRule                            = "HTTP Header Rewrite Rule"
	HTTPRedirectPolicy                         = "HTTP Redirect Policy"
	HeaderRewritePolicy                        = "Header Rewrite Policy"
	L4VS                                       = "L4 Virtual Service"
	L4VIP                                      = "L4 VIP"
	L4Pool                                     = "L4 Pool"
	L4AdvPool                                  = "L4 Advance Pool"
	L4PS                                       = "L4 Policyset"
	L4PSRule                                   = "L4 Policyset Rule"
	SNIVS                                      = "SNI VirtualService"
	StringGroup                                = "StringGroup"
	StringGroupNode                            = "StringGroupNode"
	VIP                                        = "VS VIP"
	PG                                         = "Poolgroup"
	ApplicationPersistenceProfile              = "PersistenceProfile"
	ApplicationPersistenceProfileNode          = "ApplicationPersistenceProfileNode"
	PriorityLabel                              = "PriorityLabel"
	SSLKeyCert                                 = "SSLKeyandCertificate"
	PKIProfile                                 = "PKI Profile"
	ApplicationProfile                         = "ApplicationProfile"
	HealthMonitor                              = "HealthMonitor"
	AllowedTCPHealthMonitorType                = "HEALTH_MONITOR_TCP"
	AllowedUDPHealthMonitorType                = "HEALTH_MONITOR_UDP"
	AllowedSCTPHealthMonitorType               = "HEALTH_MONITOR_SCTP"
	PassthroughPG                              = "Passthrough PG"
	Passthroughpool                            = "Passthrough pool"
	PassthroughVS                              = "Passthrough VirtualService"
	Pool                                       = "Pool"
	TLSKeyCert                                 = "TLS KeyCert"
	CACert                                     = "CA Cert"
	IPCIDRRegex                                = `^(\b([01]?[0-9][0-9]?|2[0-4][0-9]|25[0-5])\.){3}([01]?[0-9][0-9]?|2[0-4][0-9]|25[0-5])\/(([0-9]|[1-2][0-9]|3[0-2]))?$`
	IPRegex                                    = `\b(([01]?[0-9][0-9]?|2[0-4][0-9]|25[0-5])(\.|$)){4}\b`
	IPV6CIDRRegex                              = `^(((?:[0-9A-Fa-f]{1,4}))*((?::[0-9A-Fa-f]{1,4}))*::((?:[0-9A-Fa-f]{1,4}))*((?::[0-9A-Fa-f]{1,4}))*|((?:[0-9A-Fa-f]{1,4}))((?::[0-9A-Fa-f]{1,4})){7})(\/([1-9]|[1-9][0-9]|1[0-1][0-9]|12[0-8])){0,1}$`
	AutoFQDNDefault                            = "Default"
	AutoFQDNFlat                               = "Flat"
	AutoFQDNDisabled                           = "Disabled"
	FQDN_SVCNAME_PREFIX                        = "s"
	FQDN_SVCNAMESPACE_PREFIX                   = "n"
	FQDN_SUBDOMAIN_PREFIX                      = "d"
	DNS_LABEL_LENGTH                           = 63
	VCF_NETWORK                                = "vcf-ako-net"
	VIP_PER_NAMESPACE                          = "VIP_PER_NAMESPACE"
	PRIMARY_AKO_FLAG                           = "PRIMARY_AKO_FLAG"
	CRDActive                                  = "ACTIVE"
	CRDInactive                                = "INACTIVE"
	SSLPort                                    = 443
	IPAMProviderInfoblox                       = "IPAMDNS_TYPE_INFOBLOX"
	IPAMProviderCustom                         = "IPAMDNS_TYPE_CUSTOM"
	SharedVipServiceKey                        = "SharedVipService"
	HttpRulePkiAndDestCASetErr                 = "PKIProfile and DestinationCA fields are set in the HTTPRule. Only one of the field should be set."
	IPTypeV4Only                               = "V4_ONLY"
	IPTypeV6Only                               = "V6_ONLY"
	IPTypeV4V6                                 = "V4_V6"
	IstioCertOutputPath                        = "/etc/istio-output-certs"
	IstioSecret                                = "istio-secret"
	IstioModel                                 = "istioModel"
	CTRL_VERSION_21_1_3                        = "21.1.3"
	FullSyncInterval                           = 300
	Namespace                                  = "Namespace"
	VrfContextNotFoundError                    = "VrfContext not found"
	VrfContextNoPermission                     = "Cannot modify VrfContext"
	HTTPMethodGet                              = "GET"
	HTTPMethodPut                              = "PUT"
	VrfContextObjectNotFoundError              = "VrfContext object not found"
	NetworkNotFoundError                       = "Network object not found"
	TenantDoesNotExist                         = "Tenant '%s' does not exist!"
	FQDNReusePolicyStrict                      = "strict"
	FQDNReusePolicyOpen                        = "internamespaceallowed"
	DefaultPSName                              = "default-backend"
	ConcurrentUpdateError                      = "the object has been modified"

	// AKO Event constants
	AKOEventComponent        = "avi-kubernetes-operator"
	AKOShutdown              = "AKOShutdown"
	SyncDisabled             = "SyncDisabled"
	ValidatedUserInput       = "ValidatedUserInput"
	StatusSync               = "StatusSync"
	AKOReady                 = "AKOReady"
	AKOPause                 = "AKOPause"
	DuplicateHostPath        = "DuplicateHostPath"
	DuplicateHost            = "DuplicateHost"
	Removed                  = "Removed"
	Synced                   = "Synced"
	Attached                 = "Attached"
	Detached                 = "Detached"
	PatchFailed              = "PatchFailed"
	InvalidConfiguration     = "InvalidConfiguration"
	AKODeleteConfigSet       = "AKODeleteConfigSet"
	AKODeleteConfigUnset     = "AKODeleteConfigUnset"
	AKODeleteConfigDone      = "AKODeleteConfigDone"
	AKODeleteConfigTimeout   = "AKODeleteConfigTimeout"
	SSLImportError           = "SSLImportError"
	SSLCertImportError       = "Import error"
	SSLCertCommonNameError   = "common_name"
	AKOGatewayEventComponent = "avi-kubernetes-operator-gateway-api"
	IngressAddEvent          = "IngressAddEvent"
	IngressDeleteEvent       = "IngressDeleteEvent"
	IngressUpdateEvent       = "IngressUpdateEvent"
	RouteAddEvent            = "RouteAddEvent"
	RouteUpdateEvent         = "RouteUpdateEvent"

	DefaultIngressClassAnnotation    = "ingressclass.kubernetes.io/is-default-class"
	ExternalDNSAnnotation            = "external-dns.alpha.kubernetes.io/hostname"
	GatewayFinalizer                 = "gateway.ako.vmware.com"
	IngressFinalizer                 = "ingress.ako.vmware.com/finalizer"
	AkoGroup                         = "ako.vmware.com"
	AviIngressController             = "ako.vmware.com/avi-lb"
	AKOConditionType                 = "ako.vmware.com/ObjectDeletionInProgress"
	DefaultSecretEnabled             = "ako.vmware.com/enable-tls"
	VSTrafficDisabled                = "ako.vmware.com/disable-traffic"
	GatewayNameLabelKey              = "service.route.lbapi.run.tanzu.vmware.com/gateway-name"
	GatewayNamespaceLabelKey         = "service.route.lbapi.run.tanzu.vmware.com/gateway-namespace"
	GatewayTypeLabelKey              = "service.route.lbapi.run.tanzu.vmware.com/type"
	AviGatewayController             = "lbapi.run.tanzu.vmware.com/avi-lb"
	SvcApiGatewayNameLabelKey        = "ako.vmware.com/gateway-name"
	SvcApiGatewayNamespaceLabelKey   = "ako.vmware.com/gateway-namespace"
	SvcApiAviGatewayController       = "ako.vmware.com/avi-lb"
	NPLPodAnnotation                 = "nodeportlocal.antrea.io"
	NPLSvcAnnotation                 = "nodeportlocal.antrea.io/enabled"
	InfraSettingNameAnnotation       = "aviinfrasetting.ako.vmware.com/name"
	SkipNodePortAnnotation           = "skipnodeport.ako.vmware.com/enabled"
	PassthroughAnnotation            = "passthrough.ako.vmware.com/enabled"
	StaticRouteAnnotation            = "ako.vmware.com/pod-cidrs"
	OVNNodeSubnetAnnotation          = "k8s.ovn.org/node-subnets"
	WCPSEGroup                       = "ako.vmware.com/wcp-se-group"
	WCPCloud                         = "ako.vmware.com/wcp-cloud-name"
	WCPAKOUserClusterName            = "ako.vmware.com/ako-user-cluster-name"
	VSAnnotation                     = "ako.vmware.com/host-fqdn-vs-uuid-map"
	ControllerAnnotation             = "ako.vmware.com/controller-cluster-uuid"
	SharedVipSvcLBAnnotation         = "ako.vmware.com/enable-shared-vip"
	LoadBalancerIP                   = "ako.vmware.com/load-balancer-ip"
	LBSvcAppProfileAnnotation        = "ako.vmware.com/application-profile"
	L4RuleAnnotation                 = "ako.vmware.com/l4rule"
	CalicoIPv4AddressAnnotation      = "projectcalico.org/IPv4Address"
	CalicoIPv6AddressAnnotation      = "projectcalico.org/IPv6Address"
	AntreaTransportAddressAnnotation = "node.antrea.io/transport-addresses"
	TenantAnnotation                 = "ako.vmware.com/tenant-name"
	GwProxyProtocolEnableAnnotation  = "iaas.vmware.com/proxy-protocol-enabled"

	// Specifies command used in namespace event handler
	NsFilterAdd                    = "ADD"
	NsFilterDelete                 = "DELETE"
	PoolNameSuffixForHttpPolToPool = "policy-to-pool"
	AVI_OBJ_NAME_MAX_LENGTH        = 255
	ACCESS_TOKEN_TYPE_JWT          = "ACCESS_TOKEN_TYPE_JWT"
	ACCESS_TOKEN_TYPE_OPAQUE       = "ACCESS_TOKEN_TYPE_OPAQUE"
	SAML_AUTHN_REQ_ACS_TYPE_INDEX  = "SAML_AUTHN_REQ_ACS_TYPE_INDEX"

	// License types
	LicenseTypeEnterprise              = "ENTERPRISE"
	LicenseTypeEnterpriseCloudServices = "ENTERPRISE_WITH_CLOUD_SERVICES"
)

// Cache Indexer constants.
const (
	// AviSettingGWClassIndex maintains a map of AviInfraSetting Name to
	// GatewayClass Objects. This helps in fetching all GatewayClasses with a
	// given AviinfraSetting Name.
	AviSettingGWClassIndex = "aviSettingGWClass"

	// GatewayClassGatewayIndex maintains a map of GatewayClass Name to
	// Gateway Objects. This helps in fetching all Gateways with a
	// given GatewayClass Name.
	GatewayClassGatewayIndex = "gatewayClassGateway"

	// SeGroupAviSettingIndex maintains a map of SeGroup Name to
	// AviInfraSetting Objects. This helps in fetching all AviInfraSettings
	// with a given SeGroup Name.
	SeGroupAviSettingIndex = "seGroupAviSetting"

	// AviSettingServicesIndex maintains a map of AviInfraSetting Objects to
	// Service Namespace/Name. This helps in fettching all Services
	// with a given AviInfraSetting.
	AviSettingServicesIndex = "aviSettingServices"

	// AviSettingIngClassIndex maintains a map of AviInfraSetting Name to
	// IngressClass Objects. This helps in fetching all IngressClasses with a
	// given AviinfraSetting Name.
	AviSettingIngClassIndex = "aviSettingIngClass"

	// v maintains a map of AviInfraSetting Name to
	// Route Objects. This helps in fetching all Routes with a
	// given AviinfraSetting Name.
	AviSettingRouteIndex = "aviSettingRoute"

	// L4RuleToServicesIndex maintains a map of L4Rule CRD to
	// Service objects. This helps in fetching all Services
	// with a given L4Rule CRD name.
	L4RuleToServicesIndex = "l4RuleToServicesIndex"

	// AviSettingNamespaceIndex maintains a map of AviInfraSetting Objects to
	// Namespace objects. This helps in fettching a Namespace with a given
	// AviInfraSetting.
	AviSettingNamespaceIndex = "aviSettingNamespaces"
)

// Passthrough deployment same in EVH and SNI. Not changing log messages.
const (
	PassthroughDatascript = `local avi_tls = require "Default-TLS"
	buffered = avi.l4.collect(20)
	payload = avi.l4.read()
	len = avi_tls.get_req_buffer_size(payload)
	if ( buffered < len ) then
	  avi.l4.collect(len)
	end
	if ( avi_tls.sanity_check(payload) ) then
	   local h = avi_tls.parse_record(payload)
	   local sname = avi_tls.get_sni(h)
	   if sname == nil then
		  avi.vs.log('SNI not present')
		  avi.vs.close_conn()
	   else
		  avi.vs.log("SNI=".. sname)
		  pg_name = "CLUSTER--AVIINFRA"..sname
		  avi.poolgroup.select(pg_name)
	   end
	else
	   avi.vs.close_conn()
	end
	avi.l4.ds_done()
	avi_tls = nil`
)
