/*
 * [2013] - [2018] Avi Networks Incorporated
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

	avimodels "github.com/avinetworks/sdk/go/models"
	oshiftinformers "github.com/openshift/client-go/route/informers/externalversions/route/v1"
	coreinformers "k8s.io/client-go/informers/core/v1"
	extensioninformers "k8s.io/client-go/informers/extensions/v1beta1"
	networking "k8s.io/client-go/informers/networking/v1beta1"
	"k8s.io/client-go/kubernetes"
)

type EvType string

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
	AVI_DEFAULT_TCP_HM string = "System-TCP"
	AVI_DEFAULT_UDP_HM string = "System-UDP"
)

const (
	INFORMERS_INSTANTIATE_ONCE string = "instantiateOnce"
	INFORMERS_OPENSHIFT_CLIENT string = "oshiftClient"
	INFORMERS_NAMESPACE        string = "namespace"
)

type KubeClientIntf struct {
	ClientSet kubernetes.Interface
}

type Informers struct {
	ConfigMapInformer     coreinformers.ConfigMapInformer
	ServiceInformer       coreinformers.ServiceInformer
	EpInformer            coreinformers.EndpointsInformer
	PodInformer           coreinformers.PodInformer
	NSInformer            coreinformers.NamespaceInformer
	SecretInformer        coreinformers.SecretInformer
	ExtV1IngressInformer  extensioninformers.IngressInformer
	RouteInformer         oshiftinformers.RouteInformer
	NodeInformer          coreinformers.NodeInformer
	CoreV1IngressInformer networking.IngressInformer // New ingress API
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
	err       error
	operation string
}

func (e *WebSyncError) Error() string  { return fmt.Sprintf("Error during %s: %v", e.operation, e.err) }
func (e *SkipSyncError) Error() string { return e.Msg }

var CloudName string

func init() {
	CloudName = os.Getenv("CLOUD_NAME")
	if CloudName == "" {
		// If the cloud name is blank - assume it to be Default-Cloud
		CloudName = "Default-Cloud"
	}
}
