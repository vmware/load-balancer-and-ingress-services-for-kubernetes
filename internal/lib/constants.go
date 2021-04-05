/*
 * Copyright 2019-2020 VMware, Inc.
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
	OPENSHIFT_CNI             = "openshift"
	INGRESS_API               = "INGRESS_API"
	AviConfigMap              = "avi-k8s-config"
	AviSecret                 = "avi-secret"
	AviNS                     = "avi-system"
	VMwareNS                  = "vmware-system-ako"

	AVI_INGRESS_CLASS                          = "avi"
	SUBNET_IP                                  = "SUBNET_IP"
	SUBNET_PREFIX                              = "SUBNET_PREFIX"
	NETWORK_NAME                               = "NETWORK_NAME"
	VIP_NETWORK_LIST                           = "VIP_NETWORK_LIST"
	SEG_NAME                                   = "SEG_NAME"
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
	CLOUD_GCP                                  = "CLOUD_GCP"
	CLOUD_NONE                                 = "CLOUD_NONE"
	DEFAULT_SHARD_SCHEME                       = "hostname"
	HOSTNAME_SHARD_SCHEME                      = "hostname"
	NAMESPACE_SHARD_SCHEME                     = "namespace"
	SLOW_RETRY_LAYER                           = "SlowRetryLayer"
	FAST_RETRY_LAYER                           = "FastRetryLayer"
	NOT_FOUND                                  = "HTTP code: 404"
	STATUS_REDIRECT                            = "HTTP_REDIRECT_STATUS_CODE_302"
	SLOW_SYNC_TIME                             = 90 // seconds
	LOG_LEVEL                                  = "logLevel"
	LAYER7_ONLY                                = "layer7Only"
	NO_PG_FOR_SNI                              = "noPGForSNI"
	SERVICE_TYPE                               = "SERVICE_TYPE"
	NODE_PORT                                  = "NodePort"
	NODE_KEY                                   = "NODE_KEY"
	NODE_VALUE                                 = "NODE_VALUE"
	ShardVSPrefix                              = "Shared-L7"
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
	VSVIPDELCTRLVER                            = "20.1.1"
	Advl4ControllerVersion                     = "20.1.2"
	ControllerVersion2014                      = "20.1.4"
	HostRule                                   = "HostRule"
	HTTPRule                                   = "HTTPRule"
	AviInfraSetting                            = "AviInfraSetting"
	DummySecret                                = "@avisslkeycertrefdummy"
	StatusRejected                             = "Rejected"
	StatusAccepted                             = "Accepted"
	AllowedApplicationProfile                  = "APPLICATION_PROFILE_TYPE_HTTP"
	TypeTLSReencrypt                           = "reencrypt"
	DefaultPoolSSLProfile                      = "System-Standard"
	LB_ALGORITHM_CONSISTENT_HASH_CUSTOM_HEADER = "LB_ALGORITHM_CONSISTENT_HASH_CUSTOM_HEADER"
	LB_ALGORITHM_CONSISTENT_HASH               = "LB_ALGORITHM_CONSISTENT_HASH"
	Gateway                                    = "Gateway"
	GatewayClass                               = "GatewayClass"
	DuplicateBackends                          = "MultipleBackendsWithSameServiceError"
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
	SeGroupLabelKey                            = "clustername"

	INGRESS_CLASS_ANNOT            = "kubernetes.io/ingress.class"
	DefaultIngressClassAnnotation  = "ingressclass.kubernetes.io/is-default-class"
	ExternalDNSAnnotation          = "external-dns.alpha.kubernetes.io/hostname"
	GatewayFinalizer               = "gateway.ako.vmware.com"
	AkoGroup                       = "ako.vmware.com"
	AviIngressController           = "ako.vmware.com/avi-lb"
	AKOConditionType               = "ako.vmware.com/ObjectDeletionInProgress"
	DefaultSecretEnabled           = "ako.vmware.com/enable-tls"
	GatewayNameLabelKey            = "service.route.lbapi.run.tanzu.vmware.com/gateway-name"
	GatewayNamespaceLabelKey       = "service.route.lbapi.run.tanzu.vmware.com/gateway-namespace"
	GatewayTypeLabelKey            = "service.route.lbapi.run.tanzu.vmware.com/type"
	AviGatewayController           = "lbapi.run.tanzu.vmware.com/avi-lb"
	SvcApiGatewayNameLabelKey      = "ako.vmware.com/gateway-name"
	SvcApiGatewayNamespaceLabelKey = "ako.vmware.com/gateway-namespace"
	SvcApiAviGatewayController     = "ako.vmware.com/avi-lb"
	NPLPodAnnotation               = "nodeportlocal.antrea.io"
	NPLSvcAnnotation               = "nodeportlocal.antrea.io/enabled"
	InfraSettingNameAnnotation     = "aviinfrasetting.ako.vmware.com/name"
	SkipNodePortAnnotation         = "skipnodeport.ako.vmware.com/enabled"

	// Specifies command used in namespace event handler
	NsFilterAdd                    = "ADD"
	NsFilterDelete                 = "DELETE"
	PoolNameSuffixForHttpPolToPool = "policy-to-pool"
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
)

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
		  pg_name = "CLUSTER--"..sname
		  avi.poolgroup.select(pg_name)
	   end
	else
	   avi.vs.close_conn()
	end
	avi.l4.ds_done()
	avi_tls = nil`
)
