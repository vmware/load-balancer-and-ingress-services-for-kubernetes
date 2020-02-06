/*
* [2013] - [2019] Avi Networks Incorporated
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

package nodes

import (
	"fmt"
	"sort"

	avimodels "github.com/avinetworks/sdk/go/models"
	avicache "gitlab.eng.vmware.com/orion/akc/pkg/cache"
	"gitlab.eng.vmware.com/orion/container-lib/utils"
)

type AviModelNode interface {
	//Each AVIModelNode represents a AVI API object.
	GetCheckSum() uint32
	CalculateCheckSum()
	GetNodeType() string
}

type AviObjectGraphIntf interface {
	GetOrderedNodes() []AviModelNode
}

type AviObjectGraph struct {
	modelNodes    []AviModelNode
	Name          string
	GraphChecksum uint32
	IsVrf         bool
}

func (v *AviObjectGraph) GetCheckSum() uint32 {
	// Calculate checksum and return
	v.CalculateCheckSum()
	return v.GraphChecksum
}

func (v *AviObjectGraph) CalculateCheckSum() {
	// A sum of fields for this model.
	v.GraphChecksum = 0
	for _, model := range v.modelNodes {
		v.GraphChecksum = v.GraphChecksum + model.GetCheckSum()
	}
}

func NewAviObjectGraph() *AviObjectGraph {
	return &AviObjectGraph{}
}

func (o *AviObjectGraph) AddModelNode(node AviModelNode) {
	o.modelNodes = append(o.modelNodes, node)
}

func (o *AviObjectGraph) RemovePoolNodeRefs(poolName string) {
	for _, node := range o.modelNodes {
		if node.GetNodeType() == "VirtualServiceNode" {
			for i, pool := range node.(*AviVsNode).PoolRefs {
				if pool.Name == poolName {
					utils.AviLog.Info.Printf("Removing poolref: %s", poolName)
					utils.AviLog.Info.Printf("Before removing the pool nodes are: %s", utils.Stringify(node.(*AviVsNode).PoolRefs))
					node.(*AviVsNode).PoolRefs = append(node.(*AviVsNode).PoolRefs[:i], node.(*AviVsNode).PoolRefs[i+1:]...)
					break
				}
			}
			utils.AviLog.Info.Printf("After removing the pool nodes are: %s", utils.Stringify(node.(*AviVsNode).PoolRefs))
		}
	}
}

func (o *AviObjectGraph) GetOrderedNodes() []AviModelNode {
	return o.modelNodes
}

type AviVrfNode struct {
	Name             string
	StaticRoutes     []*avimodels.StaticRoute
	CloudConfigCksum uint32
}

func (v *AviVrfNode) GetCheckSum() uint32 {
	// Calculate checksum and return
	v.CalculateCheckSum()
	return v.CloudConfigCksum
}

func (v *AviVrfNode) GetNodeType() string {
	return "VrfNode"
}

func (v *AviVrfNode) CalculateCheckSum() {
	// A sum of fields for this vrf.
	checksum := utils.Hash(v.Name) + utils.Hash(utils.Stringify(v.StaticRoutes))
	v.CloudConfigCksum = checksum
}

func (o *AviObjectGraph) GetAviVRF() []*AviVrfNode {
	var aviVrf []*AviVrfNode
	for _, model := range o.modelNodes {
		vrf, ok := model.(*AviVrfNode)
		if ok {
			aviVrf = append(aviVrf, vrf)
		}
	}
	return aviVrf
}

type AviVsNode struct {
	Name               string
	Tenant             string
	ApplicationProfile string
	NetworkProfile     string
	PortProto          []AviPortHostProtocol // for listeners
	DefaultPool        string
	EastWest           bool
	CloudConfigCksum   uint32
	DefaultPoolGroup   string
	HTTPChecksum       uint32
	SNIParent          bool
	PoolGroupRefs      []*AviPoolGroupNode
	PoolRefs           []*AviPoolNode
	TCPPoolGroupRefs   []*AviPoolGroupNode
	HTTPDSrefs         []*AviHTTPDataScriptNode
	SniNodes           []*AviVsNode
	SharedVS           bool
	SSLKeyCertRefs     []*AviTLSKeyCertNode
	HttpPolicyRefs     []*AviHttpPolicySetNode
	VSVIPRefs          []*AviVSVIPNode
	VHParentName       string
	VHDomainNames      []string
	TLSType            string
	IsSNIChild         bool
	ServiceMetadata    avicache.LBServiceMetadataObj
	VrfContext         string
}

func (o *AviObjectGraph) GetAviVS() []*AviVsNode {
	var aviVs []*AviVsNode
	for _, model := range o.modelNodes {
		vs, ok := model.(*AviVsNode)
		if ok {
			aviVs = append(aviVs, vs)
		}
	}
	return aviVs
}

func (v *AviVsNode) GetCheckSum() uint32 {
	// Calculate checksum and return
	v.CalculateCheckSum()
	return v.CloudConfigCksum
}

func (v *AviVsNode) GetNodeType() string {
	// Calculate checksum and return
	return "VirtualServiceNode"
}

func (v *AviVsNode) CalculateCheckSum() {
	// A sum of fields for this VS.
	checksum := utils.Hash(v.ApplicationProfile) + utils.Hash(v.NetworkProfile) + utils.Hash(utils.Stringify(v.PortProto)) + utils.Hash(utils.Stringify(v.HTTPDSrefs)) + utils.Hash(utils.Stringify(v.SniNodes)) + utils.Hash(utils.Stringify(v.ServiceMetadata))
	v.CloudConfigCksum = checksum
}

type AviHttpPolicySetNode struct {
	Name             string
	Tenant           string
	CloudConfigCksum uint32
	HppMap           []AviHostPathPortPoolPG
	RedirectPorts    []AviRedirectPort
}

func (v *AviHttpPolicySetNode) GetCheckSum() uint32 {
	// Calculate checksum and return
	v.CalculateCheckSum()
	return v.CloudConfigCksum
}

func (v *AviHttpPolicySetNode) CalculateCheckSum() {
	// A sum of fields for this VS.
	var checksum uint32
	for _, hpp := range v.HppMap {
		sort.Strings(hpp.Host)
		sort.Strings(hpp.Path)
		checksum = checksum + utils.Hash(utils.Stringify(hpp))
	}
	for _, redir := range v.RedirectPorts {
		sort.Strings(redir.Hosts)
		checksum = checksum + utils.Hash(utils.Stringify(redir.Hosts))
	}
	utils.AviLog.Info.Printf("The HTTP rules during checksum calculation is: %s with checksum: %v", utils.Stringify(v.HppMap), checksum)
	v.CloudConfigCksum = checksum
}

func (v *AviHttpPolicySetNode) GetNodeType() string {
	// Calculate checksum and return
	return "HTTPPolicyNode"
}

type AviHostPathPortPoolPG struct {
	Host          []string
	Path          []string
	Port          uint32
	Pool          string
	PoolGroup     string
	MatchCriteria string
}

type AviRedirectPort struct {
	Hosts        []string
	RedirectPort int32
	StatusCode   string
	VsPort       int32
}

type AviTLSKeyCertNode struct {
	Name             string
	Tenant           string
	CloudConfigCksum uint32
	Key              []byte
	Cert             []byte
	Port             int32
}

func (v *AviTLSKeyCertNode) CalculateCheckSum() {
	// A sum of fields for this SSL cert.
	checksum := utils.Hash(string(v.Key)) + utils.Hash(string(v.Cert))
	v.CloudConfigCksum = checksum
}

func (v *AviTLSKeyCertNode) GetCheckSum() uint32 {
	v.CalculateCheckSum()
	return v.CloudConfigCksum
}

func (v *AviTLSKeyCertNode) GetNodeType() string {
	// Calculate checksum and return
	return "TLSCertNode"
}

type AviPortHostProtocol struct {
	PortMap     map[string][]int32
	Port        int32
	Protocol    string
	Hosts       []string
	Secret      string
	Passthrough bool
	Redirect    bool
	EnableSSL   bool
	Name        string
}

type AviVSVIPNode struct {
	Name             string
	Tenant           string
	CloudConfigCksum uint32
	FQDNs            []string
	EastWest         bool
	VrfContext       string
}

func (v *AviVSVIPNode) GetCheckSum() uint32 {
	// Calculate checksum and return
	v.CalculateCheckSum()
	return v.CloudConfigCksum
}

func (v *AviVSVIPNode) CalculateCheckSum() {
	// A sum of fields for this VS.
	checksum := utils.Hash(utils.Stringify(v.FQDNs))
	v.CloudConfigCksum = checksum
}

func (v *AviVSVIPNode) GetNodeType() string {
	return "VSVIPNode"
}

type AviPoolGroupNode struct {
	Name                  string
	Tenant                string
	CloudConfigCksum      uint32
	Members               []*avimodels.PoolGroupMember
	Port                  string
	ImplicitPriorityLabel bool
}

func (v *AviPoolGroupNode) GetCheckSum() uint32 {
	// Calculate checksum and return
	v.CalculateCheckSum()
	return v.CloudConfigCksum
}

func (v *AviPoolGroupNode) CalculateCheckSum() {
	// A sum of fields for this VS.
	checksum := utils.Hash(utils.Stringify(v.Members))
	v.CloudConfigCksum = checksum
}

func (o *AviObjectGraph) GetPoolGroupByName(pgName string) *AviPoolGroupNode {
	for _, model := range o.modelNodes {
		pg, ok := model.(*AviPoolGroupNode)
		if ok {
			if pg.Name == pgName {
				utils.AviLog.Info.Printf("Found PG with name: %s", pg.Name)
				return pg
			}
		}
	}
	return nil
}

func (v *AviPoolGroupNode) GetNodeType() string {
	return "PoolGroupNode"
}

type AviHTTPDataScriptNode struct {
	Name             string
	Tenant           string
	CloudConfigCksum uint32
	PoolGroupRefs    []string
	*DataScript
}

func (v *AviHTTPDataScriptNode) GetCheckSum() uint32 {
	// Calculate checksum and return
	v.CalculateCheckSum()
	return v.CloudConfigCksum
}

func (v *AviHTTPDataScriptNode) CalculateCheckSum() {
	// A sum of fields for this VS.
	checksum := utils.Hash(utils.Stringify(v.PoolGroupRefs))
	v.CloudConfigCksum = checksum
}

func (v *AviHTTPDataScriptNode) GetNodeType() string {
	return "HTTPDataScript"
}

func (o *AviObjectGraph) GetAviHTTPDSNode() []*AviHTTPDataScriptNode {
	var aviDS []*AviHTTPDataScriptNode
	for _, model := range o.modelNodes {
		ds, ok := model.(*AviHTTPDataScriptNode)
		if ok {
			aviDS = append(aviDS, ds)
		}
	}
	return aviDS
}

type DataScript struct {
	Evt    string
	Script string
}

type AviPoolNode struct {
	Name             string
	Tenant           string
	CloudConfigCksum uint32
	Port             int32
	PortName         string
	Servers          []AviPoolMetaServer
	Protocol         string
	LbAlgorithm      string
	ServerClientCert string
	PkiProfile       string
	SSLProfileRef    string
	IngressName      string
	PriorityLabel    string
	ServiceMetadata  avicache.ServiceMetadataObj
	VrfContext       string
}

func (v *AviPoolNode) GetCheckSum() uint32 {
	// Calculate checksum and return
	v.CalculateCheckSum()
	return v.CloudConfigCksum
}

func (v *AviPoolNode) CalculateCheckSum() {
	// A sum of fields for this VS.
	checksum := utils.Hash(v.Protocol) + utils.Hash(fmt.Sprint(v.Port)) + utils.Hash(v.PortName) + utils.Hash(utils.Stringify(v.Servers)) + utils.Hash(utils.Stringify(v.LbAlgorithm)) + utils.Hash(utils.Stringify(v.SSLProfileRef)) + utils.Hash(utils.Stringify(v.ServerClientCert)) + utils.Hash(utils.Stringify(v.PkiProfile)) + utils.Hash(utils.Stringify(v.PriorityLabel))
	v.CloudConfigCksum = checksum
}

func (v *AviPoolNode) GetNodeType() string {
	return "PoolNode"
}

func (o *AviObjectGraph) GetAviPoolNodes() []*AviPoolNode {
	var aviPool []*AviPoolNode
	for _, model := range o.modelNodes {
		pool, ok := model.(*AviPoolNode)
		if ok {
			aviPool = append(aviPool, pool)
		}
	}
	return aviPool
}

func (o *AviObjectGraph) GetAviPoolNodesByIngress(tenant string, ingName string) []*AviPoolNode {
	var aviPool []*AviPoolNode
	for _, model := range o.modelNodes {
		if model.GetNodeType() == "VirtualServiceNode" {
			for _, pool := range model.(*AviVsNode).PoolRefs {
				if pool.IngressName == ingName && tenant == pool.Tenant {
					utils.AviLog.Info.Printf("Found Pool with name: %s Adding...", pool.IngressName)
					aviPool = append(aviPool, pool)
				}
			}
		}
	}
	return aviPool
}

type AviPoolMetaServer struct {
	Ip         avimodels.IPAddr
	ServerNode string
}

type IngressHostPathSvc struct {
	ServiceName string
	Path        string
	Port        int32
}

type IngressHostMap map[string][]IngressHostPathSvc

type TlsSettings struct {
	Hosts      map[string][]IngressHostPathSvc
	SecretName string
}

type IngressConfig struct {
	TlsCollection []TlsSettings
	IngressHostMap
}
