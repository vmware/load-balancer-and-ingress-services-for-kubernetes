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

package nodes

import (
	"errors"
	"regexp"
	"strconv"
	"strings"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	"github.com/vmware/alb-sdk/go/models"
	v1 "k8s.io/api/core/v1"
)

// BuildVRFGraph : build vrf graph from k8s nodes
func (o *AviObjectGraph) BuildVRFGraph(key, vrfName, nodeName string, deleteFlag bool) error {
	o.Lock.Lock()
	defer o.Lock.Unlock()
	//fetch vrf Node
	aviVrfNodes := o.GetAviVRF()
	if len(aviVrfNodes) == 0 {
		vrfNode := &AviVrfNode{
			Name: vrfName,
		}
		o.AddModelNode(vrfNode)
		aviVrfNodes = append(aviVrfNodes, vrfNode)
	}
	// Each AKO should have single VRF node as it deals with single cluster only.
	aviVrfNode := aviVrfNodes[0]

	routeid := 1
	if len(aviVrfNode.StaticRoutes) != 0 {
		routeid = len(aviVrfNode.StaticRoutes) + 1
	} else {
		aviVrfNode.NodeStaticRoutes = make(map[string]StaticRouteDetails)
	}
	var nodeRoutes []*models.StaticRoute
	if !deleteFlag {
		node, err := utils.GetInformers().NodeInformer.Lister().Get(nodeName)
		if err != nil {
			utils.AviLog.Errorf("key: %s, Error in fetching node details: %s: %v", key, nodeName, err)
			return err
		}
		nodeRoutes, err = o.addRouteForNode(node, vrfName, routeid)
		if err != nil {
			utils.AviLog.Errorf("key: %s, Error Adding vrf for node %s: %v", key, nodeName, err)
			return err
		}
	}
	// For new node addition (coming from ingestion layer), nodes static routes will be attahced at the end
	// During reboot of AKO, nodes will be sorted. So this will give rest call
	// HA case has to be checked: As Active and passive node, after failover, will not have same ordering of static
	// routes
	nodeStaticRouteDetails, ok := aviVrfNode.NodeStaticRoutes[nodeName]
	if !ok {
		//node not found, check overlapping and then add case
		if !findRoutePrefix(nodeRoutes, aviVrfNode.StaticRoutes, key) {
			// node is not present and no overlapping of cidr, append at last
			aviVrfNode.StaticRoutes = append(aviVrfNode.StaticRoutes, nodeRoutes...)
			nodeStaticRoute := StaticRouteDetails{}
			// start index shows at what index of StaticRoutes, nodes routes start (index based zero)
			nodeStaticRoute.StartIndex = routeid - 1
			nodeStaticRoute.Count = len(nodeRoutes)
			aviVrfNode.NodeStaticRoutes[nodeName] = nodeStaticRoute
			aviVrfNode.Nodes = append(aviVrfNode.Nodes, nodeName)
		}
	} else if !deleteFlag {
		// update case
		// Assumption: updated routes (values) for given node will not overlap with other nodes.
		// So only updating existing routes of that node.
		startIndex := nodeStaticRouteDetails.StartIndex
		lenNewNodeRoutes := len(nodeRoutes)
		diff := lenNewNodeRoutes - nodeStaticRouteDetails.Count

		var staticRouteCopy []*models.StaticRoute
		staticRouteCopy = append(staticRouteCopy, aviVrfNode.StaticRoutes[:startIndex]...)
		staticRouteCopy = append(staticRouteCopy, nodeRoutes...)

		staticRouteCopy = append(staticRouteCopy, aviVrfNode.StaticRoutes[startIndex+nodeStaticRouteDetails.Count:]...)
		aviVrfNode.StaticRoutes = staticRouteCopy

		//if diff is 0, there is no change in number of routes previously exist and newly created.
		if diff != 0 {
			updateNodeStaticRoutes(aviVrfNode, deleteFlag, nodeName, lenNewNodeRoutes, diff)
		}

	} else {
		//delete case
		startIndex := nodeStaticRouteDetails.StartIndex
		count := nodeStaticRouteDetails.Count
		var staticRouteCopy []*models.StaticRoute
		staticRouteCopy = append(staticRouteCopy, aviVrfNode.StaticRoutes[:startIndex]...)
		staticRouteCopy = append(staticRouteCopy, aviVrfNode.StaticRoutes[startIndex+count:]...)
		aviVrfNode.StaticRoutes = staticRouteCopy
		countToSubstract := aviVrfNode.NodeStaticRoutes[nodeName].Count
		updateNodeStaticRoutes(aviVrfNode, deleteFlag, nodeName, 0, -countToSubstract)
	}
	aviVrfNode.CalculateCheckSum()
	utils.AviLog.Infof("key: %s, Added vrf node %s", key, vrfName)
	utils.AviLog.Infof("key: %s, Number of static routes %v", key, len(aviVrfNode.StaticRoutes))
	return nil
}

func updateNodeStaticRoutes(aviVrfNode *AviVrfNode, isDelete bool, nodeName string, lenNewNodeRoutes, diff int) {
	//get index of nodename in node array
	index := -1
	for i := 0; i < len(aviVrfNode.Nodes); i++ {
		if aviVrfNode.Nodes[i] == nodeName {
			index = i
			if !isDelete {
				nodeNameToUpdate := aviVrfNode.Nodes[index]
				nodeDetails := aviVrfNode.NodeStaticRoutes[nodeNameToUpdate]
				nodeDetails.Count = lenNewNodeRoutes
				aviVrfNode.NodeStaticRoutes[nodeNameToUpdate] = nodeDetails
			}
			break
		}
	}
	if index != -1 {
		//Change nodemap entries till index
		for i := len(aviVrfNode.Nodes) - 1; i > index; i-- {
			nodeNameToUpdate := aviVrfNode.Nodes[i]
			nodeDetails := aviVrfNode.NodeStaticRoutes[nodeNameToUpdate]
			nodeDetails.StartIndex = nodeDetails.StartIndex + diff
			aviVrfNode.NodeStaticRoutes[nodeNameToUpdate] = nodeDetails
		}
		if isDelete {
			// now remove nodename from Nodes list
			updateNodeList := aviVrfNode.Nodes[:index]
			if index+1 < len(aviVrfNode.Nodes) {
				updateNodeList = append(updateNodeList, aviVrfNode.Nodes[index+1:]...)
			}
			aviVrfNode.Nodes = updateNodeList
		}
	}
}

func findRoutePrefix(nodeRoutes, aviRoutes []*models.StaticRoute, key string) bool {
	for _, noderoute := range nodeRoutes {
		for _, vrfroute := range aviRoutes {
			if *vrfroute.Prefix.IPAddr.Addr == *noderoute.Prefix.IPAddr.Addr {
				utils.AviLog.Errorf("key: %s, msg: static route prefix %s already exits", key, *vrfroute.Prefix.IPAddr.Addr)
				return true
			}
		}
	}
	return false
}

func (o *AviObjectGraph) addRouteForNode(node *v1.Node, vrfName string, routeid int) ([]*models.StaticRoute, error) {
	var nodeIP, nodeIP6 string
	var nodeRoutes []*models.StaticRoute
	ipFamily := lib.GetIPFamily()

	v4Type, v6Type := "V4", "V6"
	nodeIP, nodeIP6 = lib.GetIPFromNode(node)

	if ipFamily == v6Type && nodeIP6 == "" {
		utils.AviLog.Errorf("Error in fetching nodeIPv6 for %v", node.ObjectMeta.Name)
		return nil, errors.New("nodeipv6 not found")
	} else if ipFamily == v4Type && nodeIP == "" {
		utils.AviLog.Errorf("Error in fetching nodeIP for %v", node.ObjectMeta.Name)
		return nil, errors.New("nodeip not found")
	}

	podCIDRs, err := lib.GetPodCIDR(node)
	if err != nil {
		utils.AviLog.Errorf("Error in fetching Pod CIDR for %v: %s", node.ObjectMeta.Name, err.Error())
		return nil, errors.New("podcidr not found")
	}
	for _, podCIDR := range podCIDRs {
		s := strings.Split(podCIDR, "/")
		if len(s) != 2 {
			utils.AviLog.Errorf("Error in splitting Pod CIDR for %v", node.ObjectMeta.Name)
			return nil, errors.New("wrong podcidr")
		}

		m, err := strconv.Atoi(s[1])
		if err != nil {
			utils.AviLog.Errorf("Error in getting mask %v", err)
			return nil, err
		}
		clusterName := lib.GetClusterName()
		labels := lib.GetLabels()
		var prefixipType, nextHopIP, nextHopIPType string
		rev4 := regexp.MustCompile(lib.IPCIDRRegex)
		rev6 := regexp.MustCompile(lib.IPV6CIDRRegex)
		if ipFamily == v4Type && rev4.MatchString(podCIDR) {
			prefixipType = v4Type
			nextHopIP = nodeIP
			nextHopIPType = v4Type
		} else if ipFamily == v6Type && rev6.MatchString(podCIDR) {
			prefixipType = v6Type
			nextHopIP = nodeIP6
			nextHopIPType = v6Type
		} else {
			utils.AviLog.Warnf("Skipping PodCIDR %s, ipfamily is %s", podCIDR, ipFamily)
			continue
		}
		mask := int32(m)
		routeIDString := clusterName + "-" + strconv.Itoa(routeid)
		nodeRoute := models.StaticRoute{
			RouteID: &routeIDString,
			Prefix: &models.IPAddrPrefix{
				IPAddr: &models.IPAddr{
					Addr: &s[0],
					Type: &prefixipType,
				},
				Mask: &mask,
			},
			NextHop: &models.IPAddr{
				Addr: &nextHopIP,
				Type: &nextHopIPType,
			},
			Labels: labels,
		}

		nodeRoutes = append(nodeRoutes, &nodeRoute)
		routeid++
	}

	return nodeRoutes, nil
}
