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
	"sort"
	"strconv"
	"strings"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	"github.com/vmware/alb-sdk/go/models"
	v1 "k8s.io/api/core/v1"
)

func GetStaticRoutesForOtherNodes(aviVrfNode *AviVrfNode, routeId string) []*models.StaticRoute {
	var staticRouteCopy []*models.StaticRoute
	nodePrefix := lib.GetClusterName() + "-" + routeId
	for i := 0; i < len(aviVrfNode.StaticRoutes); i++ {
		if routeId == "" || !strings.HasPrefix(*aviVrfNode.StaticRoutes[i].RouteID, nodePrefix) {
			staticRouteCopy = append(staticRouteCopy, aviVrfNode.StaticRoutes[i])
		}
	}
	return staticRouteCopy
}

func (o *AviObjectGraph) CheckAndDeduplicateRecords(key string) {
	aviVrfNodes := o.GetAviVRF()
	if len(aviVrfNodes) == 0 {
		return
	}
	// Each AKO should have single VRF node as it deals with single cluster only.
	aviVrfNode := aviVrfNodes[0]

	podCidrNextHopMap := make(map[string]string)
	hasDuplicateRecords := false
	// Check if duplicate records for staticroutes exist
	for i := 0; i < len(aviVrfNode.StaticRoutes); i++ {
		_, ok := podCidrNextHopMap[*aviVrfNode.StaticRoutes[i].Prefix.IPAddr.Addr]
		if ok {
			utils.AviLog.Warnf("key: %s, VRFContext has duplicate records.", key)
			hasDuplicateRecords = true
			break
		} else {
			podCidrNextHopMap[*aviVrfNode.StaticRoutes[i].Prefix.IPAddr.Addr] = *aviVrfNode.StaticRoutes[i].NextHop.Addr
		}
	}
	if !hasDuplicateRecords {
		return
	}

	utils.AviLog.Infof("key: %s, Starting deduplication of records in VRFContext", key)

	// Clean VRFCache
	aviVrfNode.Nodes = nil
	aviVrfNode.StaticRoutes = nil
	aviVrfNode.NodeStaticRoutes = nil

	// send sorted list of nodes from here.
	allNodes := objects.SharedNodeLister().CopyAllObjects()
	var nodeNames []string
	for k := range allNodes {
		nodeNames = append(nodeNames, k)
	}
	sort.Strings(nodeNames)
	for _, nodeKey := range nodeNames {
		o.BuildVRFGraph(key, aviVrfNode.Name, nodeKey, false)
	}

	utils.AviLog.Infof("key: %s, Deduplication of records in VRFContext finished", key)
}

// BuildVRFGraph : build vrf graph from k8s nodes
func (o *AviObjectGraph) BuildVRFGraph(key, vrfName, nodeName string, deleteFlag bool) error {
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

	if len(aviVrfNode.StaticRoutes) == 0 {
		aviVrfNode.NodeStaticRoutes = make(map[string]StaticRouteDetails)
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
			aviVrfNode.StaticRoutes = GetStaticRoutesForOtherNodes(aviVrfNode, nodeStaticRouteDetails.RouteIDPrefix)
			processNodeStaticRouteAndNodeDeletion(nodeName, aviVrfNode)
			return err
		}
		var routeIdPrefix string
		if ok {
			routeIdPrefix = nodeStaticRouteDetails.RouteIDPrefix
		} else {
			routeIdPrefix = lib.Uuid4()
		}

		nodeRoutes, err = o.addRouteForNode(node, vrfName, routeIdPrefix)
		if err != nil {
			utils.AviLog.Errorf("key: %s, Error Adding vrf for node %s: %v", key, nodeName, err)
			aviVrfNode.StaticRoutes = GetStaticRoutesForOtherNodes(aviVrfNode, routeIdPrefix)
			processNodeStaticRouteAndNodeDeletion(nodeName, aviVrfNode)
			return err
		}
		if !ok {
			//node not found, check overlapping and then add case
			if len(nodeRoutes) > 0 && !findRoutePrefix(nodeRoutes, aviVrfNode.StaticRoutes, key) {
				// node is not present and no overlapping of cidr, append at last
				aviVrfNode.StaticRoutes = append(aviVrfNode.StaticRoutes, nodeRoutes...)
				nodeStaticRoute := StaticRouteDetails{}
				nodeStaticRoute.Count = len(nodeRoutes)
				nodeStaticRoute.RouteIDPrefix = routeIdPrefix
				aviVrfNode.NodeStaticRoutes[nodeName] = nodeStaticRoute
				aviVrfNode.Nodes = append(aviVrfNode.Nodes, nodeName)
			} else {
				if len(nodeRoutes) == 0 {
					//delete all the routes and details of this node
					aviVrfNode.StaticRoutes = GetStaticRoutesForOtherNodes(aviVrfNode, routeIdPrefix)
					processNodeStaticRouteAndNodeDeletion(nodeName, aviVrfNode)
				}
			}
		} else {
			// update case
			// Assumption: updated routes (values) for given node will not overlap with other nodes
			// So only updating existing routes of that node.
			utils.AviLog.Infof("key: %s, StaticRoutes before updation/deletion: [%v]", key, utils.Stringify(aviVrfNode.StaticRoutes))
			lenNewNodeRoutes := len(nodeRoutes)
			diff := lenNewNodeRoutes - nodeStaticRouteDetails.Count

			staticRouteCopy := GetStaticRoutesForOtherNodes(aviVrfNode, routeIdPrefix)
			staticRouteCopy = append(staticRouteCopy, nodeRoutes...)
			aviVrfNode.StaticRoutes = staticRouteCopy

			//if diff is 0, there is no change in number of routes previously exist and newly created.
			if diff != 0 {
				//update all the routes of this node
				updateNodeStaticRoutes(aviVrfNode, deleteFlag, nodeName, lenNewNodeRoutes)
			}
			if lenNewNodeRoutes == 0 {
				//delete all the routes of this node
				processNodeStaticRouteAndNodeDeletion(nodeName, aviVrfNode)
			}
			utils.AviLog.Infof("key: %s, StaticRoutes after updation/deletion: [%v]", key, utils.Stringify(aviVrfNode.StaticRoutes))
		}
	} else {
		//delete flag is turned on and node is deleted
		utils.AviLog.Infof("key: %s, StaticRoutes before deletion: [%v]", key, utils.Stringify(aviVrfNode.StaticRoutes))
		aviVrfNode.StaticRoutes = GetStaticRoutesForOtherNodes(aviVrfNode, nodeStaticRouteDetails.RouteIDPrefix)
		processNodeStaticRouteAndNodeDeletion(nodeName, aviVrfNode)
		utils.AviLog.Infof("key: %s, StaticRoutes after deletion: [%v]", key, utils.Stringify(aviVrfNode.StaticRoutes))
	}
	aviVrfNode.CalculateCheckSum()
	utils.AviLog.Infof("key: %s, Added vrf node %s", key, vrfName)
	utils.AviLog.Infof("key: %s, Number of static routes %v", key, len(aviVrfNode.StaticRoutes))
	utils.AviLog.Infof("key: %s, vrf node: [%v]", key, utils.Stringify(aviVrfNode))
	return nil
}
func processNodeStaticRouteAndNodeDeletion(nodeName string, aviVrfNode *AviVrfNode) {
	delete(aviVrfNode.NodeStaticRoutes, nodeName)
	nodesCopy := []string{}
	for _, node := range aviVrfNode.Nodes {
		if node != nodeName {
			nodesCopy = append(nodesCopy, node)
		}
	}
	aviVrfNode.Nodes = nodesCopy
}

func updateNodeStaticRoutes(aviVrfNode *AviVrfNode, isDelete bool, nodeNameToUpdate string, lenNewNodeRoutes int) {
	//get index of nodename in node array
	indexOfNodeUnderUpdation := -1

	for i := 0; i < len(aviVrfNode.Nodes); i++ {
		if aviVrfNode.Nodes[i] == nodeNameToUpdate {
			indexOfNodeUnderUpdation = i
			if !isDelete {
				nodeDetails := aviVrfNode.NodeStaticRoutes[nodeNameToUpdate]
				nodeDetails.Count = lenNewNodeRoutes
				aviVrfNode.NodeStaticRoutes[nodeNameToUpdate] = nodeDetails
			}
			break
		}
	}
	if indexOfNodeUnderUpdation != -1 {
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

func (o *AviObjectGraph) addRouteForNode(node *v1.Node, vrfName string, routeIdPrefix string) ([]*models.StaticRoute, error) {
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
	for index, podCIDR := range podCIDRs {
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
		routeIDString := clusterName + "-" + routeIdPrefix + "-" + strconv.Itoa(index)
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
	}
	return nodeRoutes, nil
}
