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

// BuildVRFGraph : build vrf graph from k8s nodes
func (o *AviObjectGraph) BuildVRFGraph(key string, vrfName string) error {
	o.Lock.Lock()
	defer o.Lock.Unlock()
	aviVrfNode := &AviVrfNode{
		Name: vrfName,
	}

	allNodes := objects.SharedNodeLister().CopyAllObjects()
	// We need to sort the node list so that the staticroutes are in same order always
	var nodeKeys []string
	for k := range allNodes {
		nodeKeys = append(nodeKeys, k)
	}
	sort.Strings(nodeKeys)

	utils.AviLog.Debugf("key: %s, All Nodes %v", key, allNodes)
	routeid := 1
	for _, k := range nodeKeys {
		node := allNodes[k].(*v1.Node)
		nodeRoutes, err := o.addRouteForNode(node, vrfName, routeid)
		if err != nil {
			utils.AviLog.Errorf("key: %s, Error Adding vrf for node %s: %v", key, node.Name, err)
			continue
		}
		if !findRoutePrefix(nodeRoutes, aviVrfNode.StaticRoutes, key) {
			aviVrfNode.StaticRoutes = append(aviVrfNode.StaticRoutes, nodeRoutes...)
			routeid += len(nodeRoutes)
		}
	}
	aviVrfNode.CalculateCheckSum()
	o.AddModelNode(aviVrfNode)
	utils.AviLog.Infof("key: %s, Added vrf node %s", key, vrfName)
	utils.AviLog.Infof("key: %s, Number of static routes %v", key, len(aviVrfNode.StaticRoutes))
	return nil
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
	var nodeIP string
	var nodeRoutes []*models.StaticRoute
	ipFamily := lib.GetIPFamily()

	nodeAddrs := node.Status.Addresses
	for _, addr := range nodeAddrs {
		if addr.Type == "InternalIP" {
			nodeIP = addr.Address
			break
		}
	}
	if nodeIP == "" {
		utils.AviLog.Errorf("Error in fetching nodeIP for %v", node.ObjectMeta.Name)
		return nil, errors.New("nodeip not found")
	}

	podCIDRs, err := lib.GetPodCIDR(node)
	if err != nil {
		utils.AviLog.Errorf("Error in fetching Pod CIDR for %v: %s", node.ObjectMeta.Name, err.Error())
		return nil, errors.New("podcidr not found")
	}

	nodeipType := "V4"
	re := regexp.MustCompile(lib.IPCIDRRegex)
	if re.MatchString(nodeIP + "/32") {
		if ipFamily != "V4" {
			return nil, errors.New("cannot add V4 node for ipFamily")
		}
		nodeipType = "V4"
	} else {
		if lib.GetIPFamily() != "V6" {
			return nil, errors.New("cannot add V6 node for ipFamily")
		}
		nodeipType = "V6"
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
		prefixipType := "V4"
		re := regexp.MustCompile(lib.IPCIDRRegex)
		if !re.MatchString(podCIDR) {
			prefixipType = "V6"
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
				Addr: &nodeIP,
				Type: &nodeipType,
			},
			Labels: labels,
		}

		nodeRoutes = append(nodeRoutes, &nodeRoute)
		routeid++
	}

	return nodeRoutes, nil
}
