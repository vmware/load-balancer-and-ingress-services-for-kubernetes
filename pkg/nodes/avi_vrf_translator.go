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
	"errors"
	"strconv"
	"strings"

	"github.com/avinetworks/sdk/go/models"
	"gitlab.eng.vmware.com/orion/akc/pkg/objects"
	"gitlab.eng.vmware.com/orion/container-lib/utils"
	v1 "k8s.io/api/core/v1"
)

// BuildVRFGraph : build vrf graph from k8s nodes
func (o *AviObjectGraph) BuildVRFGraph(key string, vrfName string) error {
	aviVrfNode := &AviVrfNode{
		Name: vrfName,
	}
	allNodes := objects.SharedNodeLister().GetAllObjectNames()
	utils.AviLog.Trace.Printf("key: %s, All Nodes %v\n", key, allNodes)
	routeid := 1
	for _, n := range allNodes {
		node := n.(*v1.Node)
		nodeRoute, err := o.addRouteForNode(node, vrfName, strconv.Itoa(routeid))
		if err != nil {
			utils.AviLog.Error.Printf("key: %s, Error Adding vrf for node %s: %v\n", key, node.Name, err)
			continue
		}
		routeid += 1
		aviVrfNode.StaticRoutes = append(aviVrfNode.StaticRoutes, nodeRoute)
	}
	o.AddModelNode(aviVrfNode)
	utils.AviLog.Info.Printf("key: %s, Added vrf node %s\n", key, vrfName)
	utils.AviLog.Info.Printf("key: %s, Number of static routes %v\n", key, len(aviVrfNode.StaticRoutes))
	return nil
}

func (o *AviObjectGraph) addRouteForNode(node *v1.Node, vrfName, routeid string) (*models.StaticRoute, error) {
	var nodeIP string
	var nodeRoute models.StaticRoute
	nodeRoute = models.StaticRoute{
		RouteID: &routeid,
	}
	nodeAddrs := node.Status.Addresses
	for _, addr := range nodeAddrs {
		if addr.Type == "InternalIP" {
			nodeIP = addr.Address
			break
		}
	}
	if nodeIP == "" {
		utils.AviLog.Error.Printf("Error in fetching nodeIP for %v", node.ObjectMeta.Name)
		return nil, errors.New("nodeip not found")
	}
	podCIDR := node.Spec.PodCIDR
	if podCIDR == "" {
		utils.AviLog.Error.Printf("Error in fetching Pod CIDR for %v", node.ObjectMeta.Name)
		return nil, errors.New("podcidr not found")
	}
	nodeipType := "V4"
	nodeRoute.NextHop = &models.IPAddr{
		Addr: &nodeIP,
		Type: &nodeipType,
	}
	s := strings.Split(podCIDR, "/")
	if len(s) != 2 {
		utils.AviLog.Error.Printf("Error in splitting Pod CIDR for %v", node.ObjectMeta.Name)
		return nil, errors.New("wrong podcidr")
	}
	m, err := strconv.Atoi(s[1])
	if err != nil {
		utils.AviLog.Error.Printf("Error in getting mask %v", err)
		return nil, err
	}
	prefixipType := "V4"
	mask := int32(m)
	nodeRoute.Prefix = &models.IPAddrPrefix{
		IPAddr: &models.IPAddr{
			Addr: &s[0],
			Type: &prefixipType,
		},
		Mask: &mask,
	}
	return &nodeRoute, nil
}
