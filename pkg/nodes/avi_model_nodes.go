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

	avimodels "github.com/avinetworks/sdk/go/models"
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
}

func (v *AviObjectGraph) GetCheckSum() uint32 {
	// Calculate checksum and return
	v.CalculateCheckSum()
	return v.GraphChecksum
}

func (v *AviObjectGraph) CalculateCheckSum() {
	// A sum of fields for this model.
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

func (o *AviObjectGraph) RemoveModelNode(nodeName string) {
	for i, node := range o.modelNodes {
		if node.GetNodeType() == "PoolNode" {
			if node.(*AviPoolNode).Name == nodeName {
				utils.AviLog.Info.Printf("Removed PoolNode: %s", nodeName)
				o.modelNodes = append(o.modelNodes[:i], o.modelNodes[i+1:]...)
			}
		}
	}
}

func (o *AviObjectGraph) GetOrderedNodes() []AviModelNode {
	return o.modelNodes
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
	// This field will detect if the HTTP policy set rules have changed.
	HTTPChecksum  uint32
	SNIParent     bool
	PoolGroupRefs []*AviPoolGroupNode
	// TODO(sudswas): Can this be a part of the PG object itself?
	PoolRefs         []*AviPoolNode
	TCPPoolGroupRefs []*AviPoolGroupNode
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
	checksum := utils.Hash(v.ApplicationProfile) + utils.Hash(v.NetworkProfile) + utils.Hash(utils.Stringify(v.PortProto))
	v.CloudConfigCksum = checksum
}

type AviPortHostProtocol struct {
	PortMap     map[string][]int32
	Port        int32
	Protocol    string
	Hosts       []string
	Secret      string
	Passthrough bool
	Redirect    bool
}

type AviPoolGroupNode struct {
	Name             string
	Tenant           string
	CloudConfigCksum uint32
	Members          []*avimodels.PoolGroupMember
	Port             string
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
}

func (v *AviPoolNode) GetCheckSum() uint32 {
	// Calculate checksum and return
	v.CalculateCheckSum()
	return v.CloudConfigCksum
}

func (v *AviPoolNode) CalculateCheckSum() {
	// A sum of fields for this VS.
	checksum := utils.Hash(v.Protocol) + utils.Hash(fmt.Sprint(v.Port)) + utils.Hash(v.PortName) + utils.Hash(utils.Stringify(v.Servers)) + utils.Hash(utils.Stringify(v.LbAlgorithm)) + utils.Hash(utils.Stringify(v.SSLProfileRef)) + utils.Hash(utils.Stringify(v.ServerClientCert)) + utils.Hash(utils.Stringify(v.PkiProfile))
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
		pool, ok := model.(*AviPoolNode)
		if ok {
			if pool.IngressName == ingName && tenant == pool.Tenant {
				utils.AviLog.Info.Printf("Found Pool with name: %s Adding...", pool.IngressName)
				aviPool = append(aviPool, pool)
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
	Host        string
	ServiceName string
	Path        string
	Port        int32
}
