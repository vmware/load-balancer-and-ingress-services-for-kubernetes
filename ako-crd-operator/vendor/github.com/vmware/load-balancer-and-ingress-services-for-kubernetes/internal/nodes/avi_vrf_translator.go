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
	"math"
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
	if len(aviVrfNode.StaticRoutes) == 0 {
		aviVrfNode.NodeStaticRoutes = make(map[string]StaticRouteDetails)
		aviVrfNode.NodeIds = make(map[int]struct{})
	}
	var nodeRoutes []*models.StaticRoute
	// For new node addition (coming from ingestion layer), nodes static routes will be attahced at the end
	// During reboot of AKO, nodes will be sorted. So this will give rest call
	// HA case has to be checked: As Active and passive node, after failover, will not have same ordering of static
	// routes
	nodeStaticRouteDetails, ok := aviVrfNode.NodeStaticRoutes[nodeName]
	if !deleteFlag {
		node, err := utils.GetInformers().NodeInformer.Lister().Get(nodeName)
		if err != nil {
			utils.AviLog.Errorf("key: %s, Error in fetching node details: %s: %v", key, nodeName, err)
			return err
		}
		if ok {
			routeid = nodeStaticRouteDetails.routeID
		} else {
			//O(n) for each node. But it will re-use previous index. So used instead of always incrementing index.
			routeid = findFreeRouteId(aviVrfNode.NodeIds)
		}
		aviVrfNode.NodeIds[routeid] = struct{}{}
		nodeRoutes, err = o.addRouteForNode(node, vrfName, routeid, aviVrfNode.NodeIds)
		if err != nil {
			utils.AviLog.Errorf("key: %s, Error Adding vrf for node %s: %v", key, nodeName, err)
			delete(aviVrfNode.NodeIds, routeid)
			return err
		}
		if !ok {
			//node not found, check overlapping and then add case
			if len(nodeRoutes) > 0 && !findRoutePrefix(nodeRoutes, aviVrfNode.StaticRoutes, key) {
				// node is not present and no overlapping of cidr, append at last
				aviVrfNode.StaticRoutes = append(aviVrfNode.StaticRoutes, nodeRoutes...)
				nodeStaticRoute := StaticRouteDetails{}
				// start index shows at what index of StaticRoutes, nodes routes start (index based zero)
				nodeStaticRoute.StartIndex = routeid - 1
				nodeStaticRoute.Count = len(nodeRoutes)
				nodeStaticRoute.routeID = routeid
				aviVrfNode.NodeStaticRoutes[nodeName] = nodeStaticRoute
				aviVrfNode.Nodes = append(aviVrfNode.Nodes, nodeName)
			} else {
				if len(nodeRoutes) == 0 {
					delete(aviVrfNode.NodeIds, routeid)
				}
			}
		} else {
			// update case
			// Assumption: updated routes (values) for given node will not overlap with other nodes.
			// So only updating existing routes of that node.
			utils.AviLog.Debugf("key: %s, StaticRoutes before updation/deletion: [%v]", key, utils.Stringify(aviVrfNode.StaticRoutes))
			startIndex := nodeStaticRouteDetails.StartIndex
			lenNewNodeRoutes := len(nodeRoutes)
			diff := lenNewNodeRoutes - nodeStaticRouteDetails.Count

			var staticRouteCopy []*models.StaticRoute
			copyTill := int(math.Min(float64(startIndex), float64(len(aviVrfNode.StaticRoutes))))
			staticRouteCopy = append(staticRouteCopy, aviVrfNode.StaticRoutes[:copyTill]...)

			staticRouteCopy = append(staticRouteCopy, nodeRoutes...)

			copyFrom := int(math.Min(float64(startIndex+nodeStaticRouteDetails.Count), float64(len(aviVrfNode.StaticRoutes))))
			staticRouteCopy = append(staticRouteCopy, aviVrfNode.StaticRoutes[copyFrom:]...)

			aviVrfNode.StaticRoutes = staticRouteCopy

			//if diff is 0, there is no change in number of routes previously exist and newly created.
			if diff != 0 {
				updateNodeStaticRoutes(aviVrfNode, deleteFlag, nodeName, lenNewNodeRoutes, diff)
			}
			if lenNewNodeRoutes == 0 {
				processNodeStaticRouteAndNodeIdDeletion(nodeName, aviVrfNode)
			}
			utils.AviLog.Debugf("key: %s, StaticRoutes after updation/deletion: [%v]", key, utils.Stringify(aviVrfNode.StaticRoutes))
		}
	} else {
		//delete case
		utils.AviLog.Debugf("key: %s, StaticRoutes before deletion: [%v]", key, utils.Stringify(aviVrfNode.StaticRoutes))
		startIndex := nodeStaticRouteDetails.StartIndex
		count := nodeStaticRouteDetails.Count
		var staticRouteCopy []*models.StaticRoute
		updateNodeStaticRoutes(aviVrfNode, deleteFlag, nodeName, 0, -count)

		copyTill := int(math.Min(float64(startIndex), float64(len(aviVrfNode.StaticRoutes))))
		staticRouteCopy = append(staticRouteCopy, aviVrfNode.StaticRoutes[:copyTill]...)

		copyFrom := int(math.Min(float64(startIndex+nodeStaticRouteDetails.Count), float64(len(aviVrfNode.StaticRoutes))))
		staticRouteCopy = append(staticRouteCopy, aviVrfNode.StaticRoutes[copyFrom:]...)

		aviVrfNode.StaticRoutes = staticRouteCopy
		processNodeStaticRouteAndNodeIdDeletion(nodeName, aviVrfNode)
		utils.AviLog.Debugf("key: %s, StaticRoutes after deletion: [%v]", key, utils.Stringify(aviVrfNode.StaticRoutes))
	}
	aviVrfNode.CalculateCheckSum()
	utils.AviLog.Infof("key: %s, Added vrf node %s", key, vrfName)
	utils.AviLog.Infof("key: %s, Number of static routes %v", key, len(aviVrfNode.StaticRoutes))
	utils.AviLog.Debugf("key: %s, vrf node: [%v]", key, utils.Stringify(aviVrfNode))
	return nil
}
func processNodeStaticRouteAndNodeIdDeletion(nodeName string, aviVrfNode *AviVrfNode) {
	delete(aviVrfNode.NodeStaticRoutes, nodeName)
	for nodeId := len(aviVrfNode.NodeIds); nodeId > len(aviVrfNode.StaticRoutes); nodeId-- {
		delete(aviVrfNode.NodeIds, nodeId)
	}
}
func findFreeRouteId(routeIdList map[int]struct{}) int {
	for i := 1; i < math.MaxInt32; i++ {
		if _, ok := routeIdList[i]; !ok {
			return i
		}
	}
	return -1
}
func updateNodeStaticRoutes(aviVrfNode *AviVrfNode, isDelete bool, nodeName string, lenNewNodeRoutes, diff int) {
	//get index of nodename in node array
	indexOfNodeUnderUpdation := -1

	for i := 0; i < len(aviVrfNode.Nodes); i++ {
		if aviVrfNode.Nodes[i] == nodeName {
			indexOfNodeUnderUpdation = i
			if !isDelete {
				nodeNameToUpdate := aviVrfNode.Nodes[indexOfNodeUnderUpdation]
				nodeDetails := aviVrfNode.NodeStaticRoutes[nodeNameToUpdate]
				nodeDetails.Count = lenNewNodeRoutes
				aviVrfNode.NodeStaticRoutes[nodeNameToUpdate] = nodeDetails
			}
			break
		}
	}
	if indexOfNodeUnderUpdation != -1 {
		clusterName := lib.GetClusterName()
		//Change nodemap entries till index
		for nodeIndex := len(aviVrfNode.Nodes) - 1; nodeIndex > indexOfNodeUnderUpdation; nodeIndex-- {
			nodeNameToUpdate := aviVrfNode.Nodes[nodeIndex]
			nodeDetails := aviVrfNode.NodeStaticRoutes[nodeNameToUpdate]
			oldStartIndex := nodeDetails.StartIndex
			nodeDetails.StartIndex = nodeDetails.StartIndex + diff
			nodeDetails.routeID = nodeDetails.StartIndex + 1
			aviVrfNode.NodeStaticRoutes[nodeNameToUpdate] = nodeDetails
			newRouteId := nodeDetails.routeID
			var tempStartIndex int
			if isDelete {
				tempStartIndex = oldStartIndex
			} else {
				tempStartIndex = nodeDetails.StartIndex
			}
			for staticRouteIndex := tempStartIndex; staticRouteIndex < nodeDetails.Count+tempStartIndex; staticRouteIndex++ {
				newRouteName := clusterName + "-" + strconv.Itoa(newRouteId)
				if staticRouteIndex > (len(aviVrfNode.StaticRoutes)-1) || staticRouteIndex < 0 {
					utils.AviLog.Warnf("Some StaticRoutes could not be updated.")
					continue
				}
				aviVrfNode.StaticRoutes[staticRouteIndex].RouteID = &newRouteName
				newRouteId++
			}
		}
		// lenNewNodeRoutes will be zero if Node exists without any PodCidr/BloackAffinity attached to it.
		if isDelete || lenNewNodeRoutes == 0 {
			// now remove nodename from Nodes list
			updateNodeList := aviVrfNode.Nodes[:indexOfNodeUnderUpdation]
			if indexOfNodeUnderUpdation+1 < len(aviVrfNode.Nodes) {
				updateNodeList = append(updateNodeList, aviVrfNode.Nodes[indexOfNodeUnderUpdation+1:]...)
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

func (o *AviObjectGraph) addRouteForNode(node *v1.Node, vrfName string, routeid int, routeIdList map[int]struct{}) ([]*models.StaticRoute, error) {
	var nodeIP, nodeIP6 string
	var nodeRoutes []*models.StaticRoute

	v4Type, v6Type := "V4", "V6"
	nodeIP, nodeIP6 = lib.GetIPFromNode(node)

	if nodeIP == "" && nodeIP6 == "" {
		utils.AviLog.Errorf("Error in fetching nodeIPs for %v", node.ObjectMeta.Name)
		return nil, errors.New("nodeip not found")
	}

	podCIDRs, err := lib.GetPodCIDR(node)
	if err != nil {
		utils.AviLog.Errorf("Error in fetching Pod CIDR for %v: %s", node.ObjectMeta.Name, err.Error())
		return nil, errors.New("podcidr not found")
	}
	for _, podCIDR := range podCIDRs {
		podCIDRAndMask := strings.Split(podCIDR, "/")
		if len(podCIDRAndMask) != 2 {
			utils.AviLog.Errorf("Error in splitting Pod CIDR for %v", node.ObjectMeta.Name)
			return nil, errors.New("wrong podcidr")
		}

		mask, err := strconv.Atoi(podCIDRAndMask[1])
		if err != nil {
			utils.AviLog.Errorf("Error in getting mask %v", err)
			return nil, err
		}
		clusterName := lib.GetClusterName()
		labels := lib.GetLabels()
		var prefixipType, nextHopIP, nextHopIPType string
		rev4 := regexp.MustCompile(lib.IPCIDRRegex)
		rev6 := regexp.MustCompile(lib.IPV6CIDRRegex)
		if nodeIP != "" && rev4.MatchString(podCIDR) {
			prefixipType = v4Type
			nextHopIP = nodeIP
			nextHopIPType = v4Type
		} else if nodeIP6 != "" && rev6.MatchString(podCIDR) {
			prefixipType = v6Type
			nextHopIP = nodeIP6
			nextHopIPType = v6Type
		} else {
			utils.AviLog.Warnf("Skipping PodCIDR %s", podCIDR)
			continue
		}
		mask32 := int32(mask)
		routeIDString := clusterName + "-" + strconv.Itoa(routeid)
		nodeRoute := models.StaticRoute{
			RouteID: &routeIDString,
			Prefix: &models.IPAddrPrefix{
				IPAddr: &models.IPAddr{
					Addr: &podCIDRAndMask[0],
					Type: &prefixipType,
				},
				Mask: &mask32,
			},
			NextHop: &models.IPAddr{
				Addr: &nextHopIP,
				Type: &nextHopIPType,
			},
			Labels: labels,
		}

		nodeRoutes = append(nodeRoutes, &nodeRoute)
		routeIdList[routeid] = struct{}{}
		routeid++
	}
	return nodeRoutes, nil
}
