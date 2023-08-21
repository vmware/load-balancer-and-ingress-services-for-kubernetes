/*
 * Copyright 2023-2024 VMware, Inc.
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
	"strconv"
	"strings"

	"github.com/vmware/alb-sdk/go/models"
	"google.golang.org/protobuf/proto"

	avimodels "github.com/vmware/alb-sdk/go/models"
	akogatewayapilib "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/lib"
	akogatewayapiobjects "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/objects"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/nodes"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
)

func (o *AviObjectGraph) ProcessL7Routes(key string, routeModel RouteModel, parentNsName string, listeners []string) {
	for _, rule := range routeModel.ParseRouteRules().Rules {

		// TODO: add the scenarios where we will not create child VS here.
		if rule.Matches == nil || rule.Backends == nil {
			continue
		}

		o.BuildChildVS(key, routeModel, parentNsName, rule, listeners)
	}
}

func (o *AviObjectGraph) BuildChildVS(key string, routeModel RouteModel, parentNsName string, rule *Rule, listeners []string) {
	parentNode := o.GetAviEvhVS()
	parentSlice := strings.Split(parentNsName, "/")
	parentNs, parentName := parentSlice[0], parentSlice[1]

	childVSName := akogatewayapilib.GetChildName(parentNs, parentName, routeModel.GetNamespace(), routeModel.GetName(), utils.Stringify(rule.Matches))
	//childVSName := "child-vs-name" //TODO: logic to get the child name (done)

	childNode := parentNode[0].GetEvhNodeForName(childVSName)
	if childNode == nil {
		childNode = &nodes.AviEvhVsNode{}
	}
	childNode.Name = childVSName
	childNode.VHParentName = parentNode[0].Name
	childNode.Tenant = lib.GetTenant()
	childNode.EVHParent = false

	//_, _, parentName := lib.ExtractTypeNameNamespace(parentNsName)
	childNode.ServiceMetadata = lib.ServiceMetadataObj{
		Gateway: parentName,
	}
	childNode.ApplicationProfile = utils.DEFAULT_L7_APP_PROFILE //TODO: move this to the routeModel
	childNode.ServiceEngineGroup = lib.GetSEGName()
	childNode.VrfContext = lib.GetVrf()
	//childNode.VHDomainNames = routeModel.ParseRouteRules().Hosts

	// TODO: add markers
	//	evhNode.AviMarkers = lib.PopulateVSNodeMarkers(namespace, host, infraSettingName)
	childNode.AviMarkers = utils.AviObjectMarkers{
		GatewayName: parentName,
		Namespace:   parentNs,
	}
	found, hosts := akogatewayapiobjects.GatewayApiLister().GetGatewayRouteToHostname(parentNs, parentName)
	if found {
		childNode.VHDomainNames = hosts
		childNode.AviMarkers.Host = hosts
	} else {
		return
	}

	// create pg pool from the backend
	o.BuildPGPool(key, childNode, routeModel, parentNsName, rule, listeners)

	// create vhmatch from the match
	o.BuildVHMatch(key, childNode, routeModel, rule)

	// create the httppolicyset if the filter is mentioned
	o.BuildHTTPPolicySet(key, childNode, routeModel, rule)
	parentNode[0].EvhNodes = append(parentNode[0].EvhNodes, childNode)
}

func (o *AviObjectGraph) BuildPGPool(key string, childVsNode *nodes.AviEvhVsNode, routeModel RouteModel, parentNsName string, rule *Rule, listeners []string) {

	parentSlice := strings.Split(parentNsName, "/")
	parentNs, parentName := parentSlice[0], parentSlice[1]
	listenerSlice := strings.Split(listeners[0], "/")
	listenerProtocol := listenerSlice[2]
	PGName := akogatewayapilib.GetPoolGroupName(parentNs, parentName,
		routeModel.GetNamespace(), routeModel.GetName(),
		utils.Stringify(rule.Matches))
	PG := nodes.AviPoolGroupNode{
		Name:   PGName,
		Tenant: lib.GetTenant(),
	}
	for _, backend := range rule.Backends {
		svcObj, err := utils.GetInformers().ServiceInformer.Lister().Services(backend.Namespace).Get(backend.Name)
		if err != nil {
			utils.AviLog.Debugf("key: %s, msg: there was an error in retrieving the service", key)
			return
		}
		poolName := akogatewayapilib.GetPoolName(parentNs, parentName,
			routeModel.GetNamespace(), routeModel.GetName(),
			utils.Stringify(rule.Matches),
			backend.Namespace, backend.Name, strconv.Itoa(int(backend.Port)))
		poolNode := &nodes.AviPoolNode{
			Name:     poolName,
			Tenant:   lib.GetTenant(),
			Protocol: listenerProtocol,
			PortName: "",
			ServiceMetadata: lib.ServiceMetadataObj{
				NamespaceServiceName: []string{backend.Namespace + "/" + backend.Name},
			},
			VrfContext: lib.GetVrf(),
		}
		poolNode.NetworkPlacementSettings, _ = lib.GetNodeNetworkMap()
		serviceType := lib.GetServiceType()
		if serviceType == lib.NodePortLocal {
			//TODO: support NPL
		} else if serviceType == lib.NodePort {
			//TODO: support nodeport
		} else {
			if servers := nodes.PopulateServers(poolNode, svcObj.ObjectMeta.Namespace, svcObj.ObjectMeta.Name, false, key); servers != nil {
				poolNode.Servers = servers
			}
		}
		childVsNode.PoolRefs = append(childVsNode.PoolRefs, poolNode)
		pool_ref := fmt.Sprintf("/api/pool?name=%s", poolNode.Name)
		PG.Members = append(PG.Members, &avimodels.PoolGroupMember{PoolRef: &pool_ref, Ratio: &backend.Weight})
	}
	childVsNode.PoolGroupRefs = append(childVsNode.PoolGroupRefs, &PG)
}

func (o *AviObjectGraph) BuildVHMatch(key string, childVsNode *nodes.AviEvhVsNode, routeModel RouteModel, rule *Rule) {
	var vhMatches []*models.VHMatch

	for _, host := range routeModel.ParseRouteRules().Hosts {
		hostname := host
		vhMatch := &models.VHMatch{
			Host: &hostname,
		}

		for i, match := range rule.Matches {
			ruleName := fmt.Sprintf("rule-%d", i)
			rule := &models.VHMatchRule{
				Name:    &ruleName,
				Matches: &models.MatchTarget{},
			}

			// path match
			if match.PathMatch != nil {
				rule.Matches.Path = &models.PathMatch{
					MatchCase: proto.String("SENSITIVE"),
					MatchStr:  []string{match.PathMatch.Path},
				}
				if match.PathMatch.Type == "Exact" {
					rule.Matches.Path.MatchCriteria = proto.String("EQUALS")
				} else if match.PathMatch.Type == "PathPrefix" {
					rule.Matches.Path.MatchCriteria = proto.String("BEGINS_WITH")
				}
			}

			// header match
			rule.Matches.Hdrs = make([]*models.HdrMatch, 0, len(match.HeaderMatch))
			for _, headerMatch := range match.HeaderMatch {
				hdrMatch := &models.HdrMatch{
					MatchCase:     proto.String("SENSITIVE"),
					MatchCriteria: proto.String("HDR_EQUALS"),
					Value:         []string{headerMatch.Value},
				}
				rule.Matches.Hdrs = append(rule.Matches.Hdrs, hdrMatch)
			}

			vhMatch.Rules = append(vhMatch.Rules, rule)
		}
		vhMatches = append(vhMatches, vhMatch)
	}
	childVsNode.VHMatches = vhMatches
}

func (o *AviObjectGraph) BuildHTTPPolicySet(key string, childVsNode *nodes.AviEvhVsNode, routeModel RouteModel, rule *Rule) {

	// create the httppolicyset from filters

}
