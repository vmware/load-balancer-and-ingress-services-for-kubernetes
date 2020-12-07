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
	CNI_PLUGIN                = "CNI_PLUGIN"
	CALICO_CNI                = "calico"
	OPENSHIFT_CNI             = "openshift"
	INGRESS_API               = "INGRESS_API"
	AviConfigMap              = "avi-k8s-config"
	AviSecret                 = "avi-secret"
	AviNS                     = "avi-system"
	VMwareNS                  = "vmware-system-ako"

	INGRESS_CLASS_ANNOT                        = "kubernetes.io/ingress.class"
	AVI_INGRESS_CLASS                          = "avi"
	SUBNET_IP                                  = "SUBNET_IP"
	SUBNET_PREFIX                              = "SUBNET_PREFIX"
	NETWORK_NAME                               = "NETWORK_NAME"
	SEG_NAME                                   = "SEG_NAME"
	DEFAULT_GROUP                              = "Default-Group"
	NODE_NETWORK_LIST                          = "NODE_NETWORK_LIST"
	NODE_NETWORK_MAX_ENTRIES                   = 5
	L7_SHARD_SCHEME                            = "L7_SHARD_SCHEME"
	DEFAULT_DOMAIN                             = "DEFAULT_DOMAIN"
	ADVANCED_L4                                = "ADVANCED_L4"
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
	RouteSecretsPrefix                         = "-route-secret"
	CertTypeVS                                 = "SSL_CERTIFICATE_TYPE_VIRTUALSERVICE"
	CertTypeCA                                 = "SSL_CERTIFICATE_TYPE_CA"
	VSVIPDELCTRLVER                            = "20.1.1"
	Advl4ControllerVersion                     = "20.1.2"
	HostRule                                   = "HostRule"
	HTTPRule                                   = "HTTPRule"
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
	GatewayNameLabelKey                        = "service.route.lbapi.run.tanzu.vmware.com/gateway-name"
	GatewayNamespaceLabelKey                   = "service.route.lbapi.run.tanzu.vmware.com/gateway-namespace"
	GatewayTypeLabelKey                        = "service.route.lbapi.run.tanzu.vmware.com/type"
	AviGatewayController                       = "lbapi.run.tanzu.vmware.com/avi-lb"
	AviIngressController                       = "ako.vmware.com/avi-lb"
	DummyVSForStaleData                        = "DummyVSForStaleData"
	ControllerReqWaitTime                      = 300
	PassthroughInsecure                        = "-insecure"
	AviControllerVSVipIDChangeError            = "Changing an existing VIP's vip_id is not supported"
	AviControllerRecreateVIPError              = "If a new preferred IP is needed, please recreate the VIP"
	DefaultSEGroup                             = "Default-Group"
	GatewayFinalizer                           = "gateway.ako.vmware.com"
	ClusterStatusCacheKey                      = "cluster-runtime"
	AviObjDeletionTime                         = 30 // Minutes
	AKOConditionType                           = "akoStatus"
	AKOStatefulSet                             = "ako"
	ObjectDeletionStartStatus                  = "objDeletionStarted"
	ObjectDeletionDoneStatus                   = "objDeletionDone"
	ObjectDeletionTimeoutStatus                = "objDeletionTimeout"
	DefaultIngressClassAnnotation              = "ingressclass.kubernetes.io/is-default-class"
	DefaultRouteCert                           = "router-certs-default"

	//Specifies command used in namespace event handler
	NsFilterAdd    = "ADD"
	NsFilterDelete = "DELETE"
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
