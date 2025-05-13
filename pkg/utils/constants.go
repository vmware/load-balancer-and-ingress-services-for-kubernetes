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

package utils

import "time"

const (
	GraphLayer                    = "GraphLayer"
	ObjectIngestionLayer          = "ObjectIngestionLayer"
	StatusQueue                   = "StatusQueue"
	LeastConnection               = "LB_ALGORITHM_LEAST_CONNECTIONS"
	RandomConnection              = "RANDOM_CONN"
	PassthroughConnection         = "PASSTHROUGH_CONN"
	RoundRobinConnection          = "LB_ALGORITHM_ROUND_ROBIN"
	ServiceInformer               = "ServiceInformer"
	PodInformer                   = "PodInformer"
	SecretInformer                = "SecretInformer"
	NodeInformer                  = "NodeInformer"
	EndpointInformer              = "EndpointInformer"
	EndpointSlicesInformer        = "EndpointSlicesInformer"
	ConfigMapInformer             = "ConfigMapInformer"
	MultiClusterIngressInformer   = "MultiClusterIngressInformer"
	ServiceImportInformer         = "ServiceImportInformer"
	K8S_TLS_SECRET_CERT           = "tls.crt"
	K8S_TLS_SECRET_KEY            = "tls.key"
	K8S_TLS_SECRET_ALT_CERT       = "alt.crt"
	K8S_TLS_SECRET_ALT_KEY        = "alt.key"
	IngressInformer               = "IngressInformer"
	RouteInformer                 = "RouteInformer"
	IngressClassInformer          = "IngressClassInformer"
	NSInformer                    = "NamespaceInformer"
	L4LBService                   = "L4LBService"
	LoadBalancer                  = "LoadBalancer"
	Pod                           = "Pod"
	Endpoints                     = "Endpoints"
	Endpointslices                = "Endpointslices"
	Ingress                       = "Ingress"
	IngressClass                  = "IngressClass"
	OshiftRoute                   = "OshiftRoute"
	Service                       = "Service"
	Secret                        = "Secret"
	HTTP                          = "HTTP"
	HTTPRoute                     = "HTTPRoute"
	HeaderMethod                  = ":method"
	HeaderAuthority               = ":authority"
	HeaderScheme                  = ":scheme"
	TLS                           = "TLS"
	HTTPS                         = "HTTPS"
	TCP                           = "TCP"
	UDP                           = "UDP"
	SCTP                          = "SCTP"
	SYSTEM_UDP_FAST_PATH          = "System-UDP-Fast-Path"
	TCP_NW_FAST_PATH              = "System-TCP-Fast-Path"
	DEFAULT_TCP_NW_PROFILE        = "System-TCP-Proxy"
	SYSTEM_SCTP_PROXY             = "System-SCTP-Proxy"
	MIXED_NET_PROFILE             = "Mixed-Network-Profile-Internal"
	DEFAULT_L4_APP_PROFILE        = "System-L4-Application"
	DEFAULT_L4_SSL_APP_PROFILE    = "System-SSL-Application"
	DEFAULT_L7_APP_PROFILE        = "System-HTTP"
	DEFAULT_L7_SECURE_APP_PROFILE = "System-Secure-HTTP"
	DEFAULT_SHARD_VS_PREFIX       = "Shard-VS-"
	L7_PG_PREFIX                  = "-PG-l7"
	VS_DATASCRIPT_EVT_HTTP_REQ    = "VS_DATASCRIPT_EVT_HTTP_REQ"
	HTTP_DS_SCRIPT                = "host = avi.http.get_host_tokens(1)\npath = avi.http.get_path_tokens(1)\nif host and path then\nlbl = host..\"/\"..path\nelse\nlbl = host..\"/\"\nend\navi.poolgroup.select(\"%s\", string.lower(lbl) )"
	HTTP_DS_SCRIPT_MODIFIED       = "host = avi.http.get_host_tokens(\"MODIFIED\", 1)\npath = avi.http.get_path_tokens(1)\nif string.contains(host, \":\") then\nfor match in string.gmatch(host, \".*:\") do\nhost = string.sub(match,0,-2)\nend\nend\nif host and path then\nlbl = host..\"/\"..path\nelse\nlbl = host..\"/\"\nend\navi.poolgroup.select(\"%s\", string.lower(lbl) )"
	ADMIN_NS                      = "admin"
	TLS_PASSTHROUGH               = "TLS_PASSTHROUGH"
	VS_TYPE_VH_PARENT             = "VS_TYPE_VH_PARENT"
	VS_TYPE_NORMAL                = "VS_TYPE_NORMAL"
	VS_TYPE_VH_CHILD              = "VS_TYPE_VH_CHILD"
	VS_TYPE_VH_ENHANCED           = "VS_TYPE_VH_ENHANCED"
	GATEWAY_API                   = "GATEWAY_API_V2"
	NodeObj                       = "Node"
	GlobalVRF                     = "global"
	VRF_CONTEXT                   = "VRF_CONTEXT"
	FULL_SYNC_INTERVAL            = "FULL_SYNC_INTERVAL"
	DEFAULT_FILE_SUFFIX           = "avi.log"
	K8S_ETIMEDOUT                 = "timed out"
	K8S_UNAUTHORIZED              = "Unauthorized"
	ADVANCED_L4                   = "ADVANCED_L4"
	SERVICES_API                  = "SERVICES_API"
	ENV_CTRL_USERNAME             = "CTRL_USERNAME"
	ENV_CTRL_PASSWORD             = "CTRL_PASSWORD"
	ENV_CTRL_AUTHTOKEN            = "CTRL_AUTHTOKEN"
	ENV_CTRL_IPADDRESS            = "CTRL_IPADDRESS"
	ENV_CTRL_CADATA               = "CTRL_CA_DATA"
	POD_NAMESPACE                 = "POD_NAMESPACE"
	VCF_CLUSTER                   = "VCF_CLUSTER"
	VPC_MODE                      = "VPC_MODE"
	MCI_ENABLED                   = "MCI_ENABLED"
	USE_DEFAULT_SECRETS_ONLY      = "USE_DEFAULT_SECRETS_ONLY"
	Namespace                     = "Namespace"
	MaxAviVersion                 = "30.2.1"
	ControllerAPIHeader           = "userHeader"
	ControllerAPIScheme           = "scheme"
	XAviUserAgentHeader           = "X-Avi-UserAgent"

	RefreshAuthTokenInterval = 12  //hours
	AuthTokenExpiry          = 240 //hours
	RefreshAuthTokenPeriod   = 0.5 //ratio

	// container-lib/api constants
	AVIAPI_INITIATING   = "INITIATING"
	AVIAPI_CONNECTED    = "CONNECTED"
	AVIAPI_DISCONNECTED = "DISCONNECTED"

	// Constants used for leader election
	leaseDuration = 15 * time.Second
	renewDeadline = 10 * time.Second
	retryPeriod   = 2 * time.Second
	leaseLockName = "ako-lease-lock"

	// Constants used in Gateway context
	WILDCARD         = "*"
	FQDN_LABEL_REGEX = "([a-z0-9-]{1,})"
)
