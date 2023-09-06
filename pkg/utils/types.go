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

import (
	"fmt"
	"os"
	"sync"
	"time"

	oshiftclientset "github.com/openshift/client-go/route/clientset/versioned"
	oshiftinformers "github.com/openshift/client-go/route/informers/externalversions/route/v1"
	avimodels "github.com/vmware/alb-sdk/go/models"
	coreinformers "k8s.io/client-go/informers/core/v1"
	netinformers "k8s.io/client-go/informers/networking/v1"
	"k8s.io/client-go/kubernetes"

	akoinformers "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/client/v1alpha1/informers/externalversions/ako/v1alpha1"
)

type EvType string

var InformerDefaultResync = 12 * time.Hour

const (
	CreateEv            EvType = "CREATE"
	UpdateEv            EvType = "UPDATE"
	DeleteEv            EvType = "DELETE"
	NumWorkersIngestion uint32 = 2
	NumWorkersGraph     uint32 = 2
)

const (
	OSHIFT_K8S_CLOUD_CONNECTOR string = "amc-k8s-cloud-connector"
)

const (
	AVI_DEFAULT_TCP_HM  string = "System-TCP"
	AVI_DEFAULT_UDP_HM  string = "System-UDP"
	AVI_DEFAULT_SCTP_HM string = "System-SCTP"
)

const (
	PatchAddOp     string = "add"
	PatchReplaceOp string = "replace"
	PatchDeleteOp  string = "delete"
)

const (
	INFORMERS_INSTANTIATE_ONCE string = "instantiateOnce"
	INFORMERS_OPENSHIFT_CLIENT string = "oshiftClient"
	INFORMERS_AKO_CLIENT       string = "akoClient"
	INFORMERS_NAMESPACE        string = "namespace"
	INFORMERS_ADVANCED_L4      string = "informersAdvL4"
	VMWARE_SYSTEM_AKO          string = "vmware-system-ako"
	AKO_DEFAULT_NS             string = "avi-system"
)

type KubeClientIntf struct {
	ClientSet kubernetes.Interface
}

type Informers struct {
	ConfigMapInformer           coreinformers.ConfigMapInformer
	ServiceInformer             coreinformers.ServiceInformer
	EpInformer                  coreinformers.EndpointsInformer
	PodInformer                 coreinformers.PodInformer
	NSInformer                  coreinformers.NamespaceInformer
	SecretInformer              coreinformers.SecretInformer
	RouteInformer               oshiftinformers.RouteInformer
	NodeInformer                coreinformers.NodeInformer
	IngressInformer             netinformers.IngressInformer
	IngressClassInformer        netinformers.IngressClassInformer
	MultiClusterIngressInformer akoinformers.MultiClusterIngressInformer
	ServiceImportInformer       akoinformers.ServiceImportInformer
	OshiftClient                oshiftclientset.Interface
	IngressVersion              string
	KubeClientIntf
}

type AviRestObjMacro struct {
	ModelName string      `json:"model_name"`
	Data      interface{} `json:"data"`
}

type RestMethod string

const (
	RestPost   RestMethod = "POST"
	RestPut    RestMethod = "PUT"
	RestDelete RestMethod = "DELETE"
	RestPatch  RestMethod = "PATCH"
	RestGet    RestMethod = "GET"
)

type RestOp struct {
	Path     string
	Method   RestMethod
	Obj      interface{}
	Tenant   string
	PatchOp  string
	Response interface{}
	Err      error
	Message  string // Optional field - can be used to carry forward err/msgs to k8s objects
	Model    string
	Version  string
	ObjName  string // Optional field - right only to be used for delete.
}

type ServiceMetadataObj struct {
	CrudHashKey string `json:"crud_hash_key"`
}

type NamespaceName struct {
	Namespace string
	Name      string
}

/*
* Meta data passed to Avi Rest Crud by Ep Crud
 */

type AviPoolMetaServer struct {
	Ip         avimodels.IPAddr
	ServerNode string
}

type K8sAviPoolMeta struct {
	Name             string
	Tenant           string
	ServiceMetadata  ServiceMetadataObj
	CloudConfigCksum string
	Port             int32
	Servers          []AviPoolMetaServer
	Protocol         string
}

type K8sAviPoolGroupMeta struct {
	Name             string
	Tenant           string
	ServiceMetadata  ServiceMetadataObj
	CloudConfigCksum string
	Members          []*avimodels.PoolGroupMember
}

type AviPortProtocol struct {
	Port     int32
	Protocol string
}

type AviPortStrProtocol struct {
	Port     string // Can be Port name or int32 string
	Protocol string
}

type AviHostPathPortPoolPG struct {
	Host      string
	Path      string
	Port      uint32
	Pool      string
	PoolGroup string
}

type K8sAviVsMeta struct {
	Name               string
	Tenant             string
	ServiceMetadata    ServiceMetadataObj
	ApplicationProfile string
	NetworkProfile     string
	PortProto          []AviPortProtocol          // for listeners
	PoolGroupMap       map[AviPortProtocol]string // for mapping listener to Pools
	DefaultPool        string
	EastWest           bool
	CloudConfigCksum   string
	DefaultPoolGroup   string
}

/*
 * Structures related with Namespace migration functionality
 */
//stores key and values fetched from values.yaml"
type NamespaceFilter struct {
	key   string
	value string
}

// Stores list of valid namespaces with lock
type K8NamespaceList struct {
	nsList map[string]struct{}
	lock   sync.RWMutex
}
type K8ValidNamespaces struct {
	nsFilter        NamespaceFilter
	EnableMigration bool
	validNSList     K8NamespaceList
}

type AviObjectMarkers struct {
	Namespace        string
	Host             []string
	InfrasettingName string
	ServiceName      string
	Path             []string
	Port             string
	Protocol         string
	IngressName      []string
	GatewayName      string
}

/*
* Obj cache
 */

type AviPoolCache struct {
	Name             string
	Tenant           string
	Uuid             string
	LbAlgorithm      string
	ServiceMetadata  ServiceMetadataObj
	CloudConfigCksum string
}

type AviVsCache struct {
	Name                 string
	Tenant               string
	Uuid                 string
	Vip                  []*avimodels.Vip
	ServiceMetadata      ServiceMetadataObj
	CloudConfigCksum     string
	PGKeyCollection      []NamespaceName
	PoolKeyCollection    []NamespaceName
	HTTPKeyCollection    []NamespaceName
	SSLKeyCertCollection []NamespaceName
	SNIChildCollection   []string
}

type AviPGCache struct {
	Name             string
	Tenant           string
	Uuid             string
	ServiceMetadata  ServiceMetadataObj
	CloudConfigCksum string
}

type AviHTTPCache struct {
	Name             string
	Tenant           string
	Uuid             string
	CloudConfigCksum string
}

type AviSSLCache struct {
	Name   string
	Tenant string
	Uuid   string
	//CloudConfigCksum string
}

type AviPkiProfileCache struct {
	Name   string
	Tenant string
	Uuid   string
	//CloudConfigCksum string
}

type AviCloudPropertyCache struct {
	Name      string
	VType     string
	NSIpam    string
	NSIpamDNS string
}

type AviHttpPolicySetMeta struct {
	Name             string
	Tenant           string
	CloudConfigCksum string
	HppMap           []AviHostPathPortPoolPG
}

type SkipSyncError struct {
	Msg string
}

type WebSyncError struct {
	Err       error
	Operation string
}

func (e *WebSyncError) Error() string         { return fmt.Sprintf("Error during %s: %v", e.Operation, e.Err) }
func (e *SkipSyncError) Error() string        { return e.Msg }
func (e *WebSyncError) GetWebAPIError() error { return e.Err }

var CloudName string

func SetCloudName(cloudName string) {
	CloudName = cloudName
}

func GetCloudRef(tenant string) string {
	return fmt.Sprintf("/api/cloud?tenant=%s&name=%s", tenant, CloudName)
}

func init() {
	CloudName = os.Getenv("CLOUD_NAME")
	if CloudName == "" {
		// If the cloud name is blank - assume it to be Default-Cloud
		CloudName = "Default-Cloud"
	}
}
