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
	"gitlab.eng.vmware.com/orion/container-lib/utils"
)

type AviModelNode interface {
	//Each AVIModelNode represents a AVI API object.
	GetCheckSum() uint32
	CalculateCheckSum()
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
	return v.GraphChecksum
}

func NewAviObjectGraph() *AviObjectGraph {
	return &AviObjectGraph{}
}
func (o *AviObjectGraph) AddModelNode(node AviModelNode) {
	o.modelNodes = append(o.modelNodes, node)
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
	HTTPChecksum uint32
	SNIParent    bool
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
	return v.CloudConfigCksum
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
