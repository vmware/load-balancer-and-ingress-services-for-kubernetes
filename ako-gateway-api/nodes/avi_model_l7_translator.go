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
	"slices"
	"strconv"
	"strings"

	"github.com/vmware/alb-sdk/go/models"
	"google.golang.org/protobuf/proto"
	"k8s.io/apimachinery/pkg/util/sets"

	akogatewayapilib "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/objects"
	akogatewayapiobjects "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/objects"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/nodes"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
)

func (o *AviObjectGraph) AddDefaultHTTPPolicySet(key string) {
	parentVS := o.GetAviEvhVS()[0]

	policyRefName := lib.GetClusterName() + "-" + utils.CloudName + "-" + akogatewayapilib.DefaultPSName
	// find default backend, if found make sure it is at last index
	for i, policyRef := range parentVS.HttpPolicyRefs {
		if policyRef.Name == policyRefName {
			if i != len(parentVS.HttpPolicyRefs)-1 {
				utils.AviLog.Debugf("key: %s msg: Found %s httpref at non last position", key, policyRefName)
				temp := parentVS.HttpPolicyRefs[i]
				parentVS.HttpPolicyRefs = append(parentVS.HttpPolicyRefs[:i], parentVS.HttpPolicyRefs[i+1:]...)
				parentVS.HttpPolicyRefs = append(parentVS.HttpPolicyRefs, temp)
			}
			return
		}
	}
	// if not found add it to last index

	utils.AviLog.Debugf("key: %s msg: %s httpref not found. Adding", key, policyRefName)
	defaultPolicyRef := &nodes.AviHttpPolicySetNode{Name: policyRefName, Tenant: lib.GetTenant()}
	defaultPolicyRef.RequestRules = []*models.HTTPRequestRule{
		{
			Name:   proto.String("default-backend-rule"),
			Enable: proto.Bool(true),
			Index:  proto.Int32(0),
			Match: &models.MatchTarget{
				Path: &models.PathMatch{
					MatchCriteria: proto.String("BEGINS_WITH"),
					MatchStr:      []string{"/"},
				},
			},
			SwitchingAction: &models.HttpswitchingAction{
				Action:     proto.String("HTTP_SWITCHING_SELECT_LOCAL"),
				StatusCode: proto.String("HTTP_LOCAL_RESPONSE_STATUS_CODE_404"),
			},
		},
	}
	parentVS.HttpPolicyRefs = append(parentVS.HttpPolicyRefs, defaultPolicyRef)
}

func (o *AviObjectGraph) ProcessL7Routes(key string, routeModel RouteModel, parentNsName string, childVSes map[string]struct{}, fullsync bool) {
	httpRouteConfig := routeModel.ParseRouteConfig(key)
	noHostsOnRoute := len(httpRouteConfig.Hosts) == 0
	httpRouteRules := httpRouteConfig.Rules
	if noHostsOnRoute {
		// Add rules on parent VS
		o.BuildParentPGPoolHTTPPS(key, routeModel, parentNsName, httpRouteRules, childVSes, fullsync)
		return
	}
	for _, rule := range httpRouteRules {
		// TODO: add the scenarios where we will not create child VS here.
		if rule.Matches == nil {
			continue
		}
		o.BuildChildVS(key, routeModel, parentNsName, rule, childVSes, fullsync)
	}
}

func (o *AviObjectGraph) BuildParentPGPoolHTTPPS(key string, routeModel RouteModel, parentNsName string, rules []*Rule, childVSes map[string]struct{}, fullsync bool) {

	parentNode := o.GetAviEvhVS()
	routeTypeNsName := lib.HTTPRoute + "/" + routeModel.GetNamespace() + "/" + routeModel.GetName()
	parentNs, _, parentName := lib.ExtractTypeNameNamespace(parentNsName)
	// For each GW + httproute, one HTTPPSPGPool
	var locaHTTTPPSPGPool objects.HTTPPSPGPool

	gwHTTPRouteKey := parentNsName + "/" + routeTypeNsName
	found, prevObj := objects.GatewayApiLister().GetGatewayRouteToHTTPSPGPool(gwHTTPRouteKey)
	if found {
		locaHTTTPPSPGPool = prevObj
	} else {
		locaHTTTPPSPGPool.HTTPPS = make([]string, 0)
		locaHTTTPPSPGPool.Pool = make([]string, 0)
		locaHTTTPPSPGPool.PoolGroup = make([]string, 0)
	}
	// TODO(Akshay): with empty hostname at httproute, there will be no hostname attached. So need to fetch non wildcard fqdns from gw listeners

	for _, rule := range rules {
		// Each rule has to be converted to one httpPS
		httpPSName := akogatewayapilib.GetChildName(parentNs, parentName, routeModel.GetNamespace(), routeModel.GetName(), utils.Stringify(rule.Matches)+utils.Stringify(rule.Filters))
		if len(rule.Filters) != 0 {
			// build HTTPPS redirect/header modifier rules first
			o.BuildHTTPPolicySet(key, parentNode[0], routeModel, rule, 0, httpPSName, &locaHTTTPPSPGPool)
		}
		if len(rule.Matches) != 0 {
			// build HTTPPS, PG and pools
			o.BuildParentHTTPPS(key, parentNsName, parentNode[0], routeModel, rule, 0, httpPSName, &locaHTTTPPSPGPool)
		}
		if len(rule.Matches) == 0 && len(rule.Backends) != 0 {
			// Default PG. Empty match name
			o.BuildDefaultPGPoolForParentVS(key, parentNsName, "", parentNode[0], routeModel, rule, &locaHTTTPPSPGPool)
		}

	}
	objects.GatewayApiLister().UpdateGatewayRouteToHTTPPSPGPool(gwHTTPRouteKey, locaHTTTPPSPGPool)
}
func (o *AviObjectGraph) BuildChildVS(key string, routeModel RouteModel, parentNsName string, rule *Rule, childVSes map[string]struct{}, fullsync bool) {

	parentNode := o.GetAviEvhVS()
	parentNs, _, parentName := lib.ExtractTypeNameNamespace(parentNsName)
	routeTypeNsName := lib.HTTPRoute + "/" + routeModel.GetNamespace() + "/" + routeModel.GetName()

	gwRouteNsName := fmt.Sprintf("%s/%s", parentNsName, routeTypeNsName)
	found, hosts := akogatewayapiobjects.GatewayApiLister().GetGatewayRouteToHostname(gwRouteNsName)
	if !found {
		utils.AviLog.Warnf("key: %s, msg: No hosts mapped to the route %s/%s/%s", key, routeModel.GetType(), routeModel.GetNamespace(), routeModel.GetName())
		return
	}

	childVSName := akogatewayapilib.GetChildName(parentNs, parentName, routeModel.GetNamespace(), routeModel.GetName(), utils.Stringify(rule.Matches))
	childVSes[childVSName] = struct{}{}

	childNode := parentNode[0].GetEvhNodeForName(childVSName)
	if childNode == nil {
		childNode = &nodes.AviEvhVsNode{}
	}
	childNode.Name = childVSName
	childNode.VHParentName = parentNode[0].Name
	childNode.Tenant = lib.GetTenant()
	childNode.EVHParent = false

	childNode.ServiceMetadata = lib.ServiceMetadataObj{
		Gateway:   parentNsName,
		HTTPRoute: routeModel.GetNamespace() + "/" + routeModel.GetName(),
	}
	childNode.ApplicationProfile = utils.DEFAULT_L7_APP_PROFILE
	childNode.ServiceEngineGroup = lib.GetSEGName()
	childNode.VrfContext = lib.GetVrf()
	childNode.AviMarkers = utils.AviObjectMarkers{
		GatewayName: parentName,
		Namespace:   parentNs,
		Host:        hosts,
	}
	updateHostname(key, parentNsName, parentNode[0])

	// create vhmatch from the match
	o.BuildVHMatch(key, parentNsName, routeTypeNsName, childNode, rule, hosts)

	if len(childNode.VHMatches) == 0 {
		utils.AviLog.Warnf("key: %s, msg: No valid domain name added for child virtual service", key)
		o.ProcessRouteDeletion(key, parentNsName, routeModel, fullsync)
		return
	}

	// create pg pool from the backend
	o.BuildPGPool(key, parentNsName, childNode, routeModel, rule)

	// create the httppolicyset if the filter is present
	o.BuildHTTPPolicySet(key, childNode, routeModel, rule, 0, childVSName, &objects.HTTPPSPGPool{})

	foundEvhModel := nodes.FindAndReplaceEvhInModel(childNode, parentNode, key)
	if !foundEvhModel {
		parentNode[0].EvhNodes = append(parentNode[0].EvhNodes, childNode)
		utils.AviLog.Debugf("key: %s, msg: added child vs %s to the parent vs %s", key, utils.Stringify(parentNode[0].EvhNodes), childNode.VHParentName)
		akogatewayapiobjects.GatewayApiLister().UpdateRouteChildVSMappings(routeModel.GetType()+"/"+routeModel.GetNamespace()+"/"+routeModel.GetName(), childVSName)
	}
	utils.AviLog.Infof("key: %s, msg: processing of child vs %s attached to parent vs %s completed", key, childNode.Name, childNode.VHParentName)
}

func updateHostname(key, parentNsName string, parentNode *nodes.AviEvhVsNode) {
	ok, routeNsNames := akogatewayapiobjects.GatewayApiLister().GetGatewayToRoute(parentNsName)
	if !ok || len(routeNsNames) == 0 {
		utils.AviLog.Warnf("key: %s, msg: No routes from gateway, removing all FQDNs", key)
		parentNode.VSVIPRefs[0].FQDNs = []string{}
		return
	}
	uniqueHostnamesSet := sets.NewString()
	for _, routeNsName := range routeNsNames {
		gwRouteNsName := fmt.Sprintf("%s/%s", parentNsName, routeNsName)
		ok, hostnames := akogatewayapiobjects.GatewayApiLister().GetGatewayRouteToHostname(gwRouteNsName)
		if !ok {
			utils.AviLog.Warnf("key: %s, msg: Unable to fetch hostname from route: %s", key, routeNsName)
		} else {
			uniqueHostnamesSet.Insert(hostnames...)
		}
	}

	uniqueHostnames := slices.DeleteFunc(uniqueHostnamesSet.List(), func(s string) bool {
		return strings.Contains(s, utils.WILDCARD)
	})

	utils.AviLog.Debugf("key: %s, unique hostnames %v found for gatewayNs: %s", uniqueHostnames, parentNsName)
	parentNode.VSVIPRefs[0].FQDNs = uniqueHostnames
}

func (o *AviObjectGraph) BuildPGPool(key, parentNsName string, childVsNode *nodes.AviEvhVsNode, routeModel RouteModel, rule *Rule) {
	//reset pool, poolgroupreferences
	childVsNode.PoolGroupRefs = nil
	childVsNode.DefaultPoolGroup = ""
	childVsNode.PoolRefs = nil
	// create the PG from backends
	routeTypeNsName := lib.HTTPRoute + "/" + routeModel.GetNamespace() + "/" + routeModel.GetName()
	parentNs, _, parentName := lib.ExtractTypeNameNamespace(parentNsName)
	allListeners := akogatewayapiobjects.GatewayApiLister().GetRouteToGatewayListener(routeTypeNsName)
	listeners := []akogatewayapiobjects.GatewayListenerStore{}
	for _, listener := range allListeners {
		if listener.Gateway == parentNsName {
			listeners = append(listeners, listener)
		}
	}
	if len(listeners) == 0 {
		utils.AviLog.Warnf("key: %s, msg: No matching listener available for the route : %s", key, routeTypeNsName)
		return
	}
	//ListenerName/port/protocol/allowedRouteSpec
	listenerProtocol := listeners[0].Protocol
	PGName := akogatewayapilib.GetPoolGroupName(parentNs, parentName,
		routeModel.GetNamespace(), routeModel.GetName(),
		utils.Stringify(rule.Matches))
	PG := &nodes.AviPoolGroupNode{
		Name:   PGName,
		Tenant: lib.GetTenant(),
	}

	for _, httpbackend := range rule.Backends {
		poolName := akogatewayapilib.GetPoolName(parentNs, parentName,
			routeModel.GetNamespace(), routeModel.GetName(),
			utils.Stringify(rule.Matches),
			httpbackend.Backend.Namespace, httpbackend.Backend.Name, strconv.Itoa(int(httpbackend.Backend.Port)))
		svcObj, err := utils.GetInformers().ServiceInformer.Lister().Services(httpbackend.Backend.Namespace).Get(httpbackend.Backend.Name)
		if err != nil {
			utils.AviLog.Debugf("key: %s, msg: there was an error in retrieving the service", key)
			o.RemovePoolRefsFromPG(poolName, o.GetPoolGroupByName(PGName))
			continue
		}
		poolNode := &nodes.AviPoolNode{
			Name:       poolName,
			Tenant:     lib.GetTenant(),
			Protocol:   listenerProtocol,
			PortName:   akogatewayapilib.FindPortName(httpbackend.Backend.Name, httpbackend.Backend.Namespace, httpbackend.Backend.Port, key),
			TargetPort: akogatewayapilib.FindTargetPort(httpbackend.Backend.Name, httpbackend.Backend.Namespace, httpbackend.Backend.Port, key),
			Port:       httpbackend.Backend.Port,
			ServiceMetadata: lib.ServiceMetadataObj{
				NamespaceServiceName: []string{httpbackend.Backend.Namespace + "/" + httpbackend.Backend.Name},
			},
			VrfContext: lib.GetVrf(),
		}

		t1LR := lib.GetT1LRPath()
		if t1LR != "" {
			poolNode.T1Lr = t1LR
			poolNode.VrfContext = ""
			utils.AviLog.Infof("key: %s, msg: setting t1LR: %s for pool node.", key, t1LR)
		}
		poolNode.NetworkPlacementSettings = lib.GetNodeNetworkMap()
		serviceType := lib.GetServiceType()
		if serviceType == lib.NodePortLocal {
			servers := nodes.PopulateServersForNPL(poolNode, svcObj.ObjectMeta.Namespace, svcObj.ObjectMeta.Name, false, key)
			if servers != nil {
				poolNode.Servers = servers
			}
		} else if serviceType == lib.NodePort {
			servers := nodes.PopulateServersForNodePort(poolNode, svcObj.ObjectMeta.Namespace, svcObj.ObjectMeta.Name, false, key)
			if servers != nil {
				poolNode.Servers = servers
			}
		} else {
			servers := nodes.PopulateServers(poolNode, svcObj.ObjectMeta.Namespace, svcObj.ObjectMeta.Name, false, key)
			if servers != nil {
				poolNode.Servers = servers
			}
		}
		if childVsNode.CheckPoolNChecksum(poolNode.Name, poolNode.GetCheckSum()) {
			// Replace the poolNode.
			childVsNode.ReplaceEvhPoolInEVHNode(poolNode, key)
		}

		pool_ref := fmt.Sprintf("/api/pool?name=%s", poolNode.Name)
		ratio := uint32(httpbackend.Backend.Weight)
		PG.Members = append(PG.Members, &models.PoolGroupMember{PoolRef: &pool_ref, Ratio: &ratio})
	}
	if len(PG.Members) > 0 {
		childVsNode.PoolGroupRefs = []*nodes.AviPoolGroupNode{PG}
		childVsNode.DefaultPoolGroup = PG.Name
	}
}

func (o *AviObjectGraph) BuildDefaultPGPoolForParentVS(key, parentNsName, matchName string, parentVsNode *nodes.AviEvhVsNode, routeModel RouteModel, rule *Rule, httpPSPGPool *objects.HTTPPSPGPool) bool {
	// create the PG from backends
	pgAttachedToVS := false
	httpRouteNamespace := routeModel.GetNamespace()
	httpRouteName := routeModel.GetName()

	routeTypeNsName := lib.HTTPRoute + "/" + httpRouteNamespace + "/" + httpRouteName
	parentNs, _, parentName := lib.ExtractTypeNameNamespace(parentNsName)

	PGName := akogatewayapilib.GetPoolGroupName(parentNs, parentName,
		httpRouteNamespace, httpRouteName, matchName)

	// Check default PG name is same as that of already assigned
	if matchName == "" && parentVsNode.DefaultPoolGroup != "" && parentVsNode.DefaultPoolGroup != PGName {
		utils.AviLog.Warnf("key: %s, msg: Parent VS already has default PG. HttpRoute %s/%s is not attached to Gateway %s", key, httpRouteNamespace, httpRouteName, parentNsName)
		//TODO(Akshay): add condition here.
		return pgAttachedToVS
	}

	allListeners := akogatewayapiobjects.GatewayApiLister().GetRouteToGatewayListener(routeTypeNsName)
	listeners := []akogatewayapiobjects.GatewayListenerStore{}
	for _, listener := range allListeners {
		if listener.Gateway == parentNsName {
			listeners = append(listeners, listener)
		}
	}
	//ListenerName/port/protocol/allowedRouteSpec
	listenerProtocol := listeners[0].Protocol

	PG := &nodes.AviPoolGroupNode{
		Name:   PGName,
		Tenant: lib.GetTenant(),
	}
	for _, backend := range rule.Backends {
		poolName := akogatewayapilib.GetPoolName(parentNs, parentName,
			routeModel.GetNamespace(), routeModel.GetName(),
			matchName,
			backend.Backend.Namespace, backend.Backend.Name, strconv.Itoa(int(backend.Backend.Port)))
		svcObj, err := utils.GetInformers().ServiceInformer.Lister().Services(backend.Backend.Namespace).Get(backend.Backend.Name)
		if err != nil {
			utils.AviLog.Debugf("key: %s, msg: there was an error in retrieving the service", key)
			o.RemovePoolRefsFromPG(poolName, o.GetPoolGroupByName(PGName))
			continue
		}
		poolNode := &nodes.AviPoolNode{
			Name:       poolName,
			Tenant:     lib.GetTenant(),
			Protocol:   listenerProtocol,
			PortName:   akogatewayapilib.FindPortName(backend.Backend.Name, backend.Backend.Namespace, backend.Backend.Port, key),
			TargetPort: akogatewayapilib.FindTargetPort(backend.Backend.Name, backend.Backend.Namespace, backend.Backend.Port, key),
			Port:       backend.Backend.Port,
			ServiceMetadata: lib.ServiceMetadataObj{
				NamespaceServiceName: []string{backend.Backend.Namespace + "/" + backend.Backend.Name},
			},
			VrfContext: lib.GetVrf(),
		}
		poolNode.NetworkPlacementSettings = lib.GetNodeNetworkMap()
		t1LR := lib.GetT1LRPath()
		if t1LR != "" {
			poolNode.T1Lr = t1LR
			poolNode.VrfContext = ""
			utils.AviLog.Infof("key: %s, msg: setting t1LR: %s for pool node.", key, t1LR)
		}
		serviceType := lib.GetServiceType()
		if serviceType == lib.NodePort {
			servers := nodes.PopulateServersForNodePort(poolNode, svcObj.ObjectMeta.Namespace, svcObj.ObjectMeta.Name, false, key)
			if servers != nil {
				poolNode.Servers = servers
			}
		} else {
			servers := nodes.PopulateServers(poolNode, svcObj.ObjectMeta.Namespace, svcObj.ObjectMeta.Name, false, key)
			if servers != nil {
				poolNode.Servers = servers
			}
		}
		if parentVsNode.CheckPoolNChecksum(poolNode.Name, poolNode.GetCheckSum()) {
			// Replace the poolNode.
			parentVsNode.ReplaceEvhPoolInEVHNode(poolNode, key)
			// Add pool to list of Pools attached to Parent VS
			httpPSPGPool.Pool = append(httpPSPGPool.Pool, poolNode.Name)
			uniquePools := sets.NewString(httpPSPGPool.Pool...)
			httpPSPGPool.Pool = uniquePools.List()
		}
		pool_ref := fmt.Sprintf("/api/pool?name=%s", poolNode.Name)
		ratio := uint32(backend.Backend.Weight)
		PG.Members = append(PG.Members, &models.PoolGroupMember{PoolRef: &pool_ref, Ratio: &ratio})
	}
	if len(PG.Members) > 0 {
		parentVsNode.ReplaceEvhPGInEVHNode(PG, key)
		// Add PG to list of PG attached to Parent VS
		httpPSPGPool.PoolGroup = append(httpPSPGPool.PoolGroup, PG.Name)
		uniquePGs := sets.NewString(httpPSPGPool.PoolGroup...)
		httpPSPGPool.PoolGroup = uniquePGs.List()
		if matchName == "" {
			parentVsNode.DefaultPoolGroup = PG.Name
		}
		pgAttachedToVS = true
	}
	return pgAttachedToVS
}

func (o *AviObjectGraph) BuildVHMatch(key string, parentNsName string, routeTypeNsName string, vsNode *nodes.AviEvhVsNode, rule *Rule, hosts []string) {
	var vhMatches []*models.VHMatch

	allListeners := objects.GatewayApiLister().GetRouteToGatewayListener(routeTypeNsName)
	listeners := []akogatewayapiobjects.GatewayListenerStore{}
	for _, listener := range allListeners {
		if listener.Gateway == parentNsName {
			listeners = append(listeners, listener)
		}
	}

	for _, host := range hosts {
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
				headerName := headerMatch.Name
				hdrMatch := &models.HdrMatch{
					MatchCase:     proto.String("SENSITIVE"),
					MatchCriteria: proto.String("HDR_EQUALS"),
					Hdr:           &headerName,
					Value:         []string{headerMatch.Value},
				}
				rule.Matches.Hdrs = append(rule.Matches.Hdrs, hdrMatch)
			}

			//port match from listener
			matchCriteria := "IS_IN"
			rule.Matches.VsPort = &models.PortMatch{
				MatchCriteria: &matchCriteria,
			}
			for _, listener := range listeners {
				rule.Matches.VsPort.Ports = append(rule.Matches.VsPort.Ports, int64(listener.Port))
			}
			//TODO correctly add protocol
			//rule.Matches.Protocol.Protocols = &listeners[0].Protocol

			vhMatch.Rules = append(vhMatch.Rules, rule)
		}
		vhMatches = append(vhMatches, vhMatch)
	}
	vsNode.VHMatches = vhMatches
	utils.AviLog.Debugf("key: %s, msg: Attached match criteria %s to vs %s", key, utils.Stringify(vsNode.VHMatches), vsNode.Name)
	utils.AviLog.Infof("key: %s, msg: Attached match criteria to vs %s", key, vsNode.Name)
}

func (o *AviObjectGraph) BuildParentHTTPPS(key, parentNsName string, vsNode *nodes.AviEvhVsNode, routeModel RouteModel, rule *Rule, index int, httpPSName string, httpPSPGPool *objects.HTTPPSPGPool) {
	var policy *nodes.AviHttpPolicySetNode
	var req_rule *models.HTTPRequestRule
	rule_index := 0
	httpRouteNamespace := routeModel.GetNamespace()
	httpRouteName := routeModel.GetName()
	// TODO: Common code. Make it function

	for i, http := range vsNode.HttpPolicyRefs {
		if http.Name == httpPSName {
			policy = vsNode.HttpPolicyRefs[i]
			index = i
			break
		}
	}
	if policy == nil {
		policy = &nodes.AviHttpPolicySetNode{Name: httpPSName, Tenant: lib.GetTenant()}
		vsNode.HttpPolicyRefs = append(vsNode.HttpPolicyRefs, policy)
		httpPSPGPool.HTTPPS = append(httpPSPGPool.HTTPPS, httpPSName)
		uniqueHTTPS := sets.NewString(httpPSPGPool.HTTPPS...)
		httpPSPGPool.HTTPPS = uniqueHTTPS.List()
		index = len(vsNode.HttpPolicyRefs) - 1
	}
	for i, requestRule := range policy.RequestRules {
		if *requestRule.Name == httpPSName {
			rule_index = i
			req_rule = requestRule
			break
		}
	}
	if req_rule == nil {
		req_rule = &models.HTTPRequestRule{Name: &httpPSName, Enable: proto.Bool(true)}
		vsNode.HttpPolicyRefs[index].RequestRules = append(vsNode.HttpPolicyRefs[index].RequestRules, req_rule)

		rule_index = len(vsNode.HttpPolicyRefs[index].RequestRules) - 1
	}
	parentNs, _, parentName := lib.ExtractTypeNameNamespace(parentNsName)

	// Code to retrieve ports associated with given httproute name
	routeTypeNsName := fmt.Sprintf("%s/%s/%s", lib.HTTPRoute, httpRouteNamespace, httpRouteName)
	allListeners := objects.GatewayApiLister().GetRouteToGatewayListener(routeTypeNsName)
	listeners := []akogatewayapiobjects.GatewayListenerStore{}
	for _, listener := range allListeners {
		if listener.Gateway == parentNsName {
			listeners = append(listeners, listener)
		}
	}

	for _, match := range rule.Matches {
		//rulename should be combination of parent, httproute and matches
		pgName := akogatewayapilib.GetHTTPRuleName(parentNs, parentName,
			httpRouteNamespace, httpRouteName, utils.Stringify(match))

		match_target := models.MatchTarget{}
		// path match
		if match.PathMatch != nil {
			matchCriteria := ""
			if match.PathMatch.Type == "Exact" {
				matchCriteria = "EQUALS"
			} else if match.PathMatch.Type == "PathPrefix" {
				matchCriteria = "BEGINS_WITH"
			}
			paths := []string{match.PathMatch.Path}
			path_match := models.PathMatch{
				MatchCriteria: proto.String(matchCriteria),
				MatchCase:     proto.String("SENSITIVE"),
				MatchStr:      paths,
			}
			match_target.Path = &path_match
		}
		// Header Match
		match_target.Hdrs = make([]*models.HdrMatch, 0, len(match.HeaderMatch))
		for _, headerMatch := range match.HeaderMatch {
			headerName := headerMatch.Name
			hdrMatch := &models.HdrMatch{
				MatchCase:     proto.String("SENSITIVE"),
				MatchCriteria: proto.String("HDR_EQUALS"),
				Hdr:           &headerName,
				Value:         []string{headerMatch.Value},
			}
			match_target.Hdrs = append(match_target.Hdrs, hdrMatch)
		}

		// attaching ports
		match_target.VsPort = &models.PortMatch{
			MatchCriteria: proto.String("IS_IN"),
		}
		for _, listener := range listeners {
			match_target.VsPort.Ports = append(match_target.VsPort.Ports, int64(listener.Port))
		}
		pgAttachedToVS := o.BuildDefaultPGPoolForParentVS(key, parentNsName, utils.Stringify(match), vsNode, routeModel, rule, httpPSPGPool)
		// switching action to PG
		sw_action := models.HttpswitchingAction{}
		if pgAttachedToVS {
			sw_action.Action = proto.String("HTTP_SWITCHING_SELECT_POOLGROUP")
			pg_ref := fmt.Sprintf("/api/poolgroup/?name=%s", pgName)
			sw_action.PoolGroupRef = proto.String(pg_ref)
		}

		vsNode.HttpPolicyRefs[index].RequestRules[rule_index].Index = proto.Int32(int32(rule_index + 1))
		vsNode.HttpPolicyRefs[index].RequestRules[rule_index].Match = &match_target
		vsNode.HttpPolicyRefs[index].RequestRules[rule_index].SwitchingAction = &sw_action
	}

}
func (o *AviObjectGraph) BuildHTTPPolicySet(key string, vsNode *nodes.AviEvhVsNode, routeModel RouteModel, rule *Rule, index int, httpPSName string, httpPSPGPool *objects.HTTPPSPGPool) {

	if len(rule.Filters) == 0 {
		vsNode.HttpPolicyRefs = nil
		return
	}
	var policy *nodes.AviHttpPolicySetNode
	// add it over here for httppolicyset
	for i, http := range vsNode.HttpPolicyRefs {
		if http.Name == httpPSName {
			policy = vsNode.HttpPolicyRefs[i]
			index = i
			break
		}
	}
	if policy == nil {
		policy = &nodes.AviHttpPolicySetNode{Name: httpPSName, Tenant: lib.GetTenant()}
		vsNode.HttpPolicyRefs = append(vsNode.HttpPolicyRefs, policy)
		httpPSPGPool.HTTPPS = append(httpPSPGPool.HTTPPS, httpPSName)
		uniqueHTTPS := sets.NewString(httpPSPGPool.HTTPPS...)
		httpPSPGPool.HTTPPS = uniqueHTTPS.List()
		index = len(vsNode.HttpPolicyRefs) - 1
	}
	isRedirectPresent := o.BuildHTTPPolicySetHTTPRequestRedirectRules(key, httpPSName, vsNode, routeModel, rule.Filters, index)
	if isRedirectPresent {
		// When the RedirectAction is specified the Request and Response Modify Header Action
		// won't have any effect, hence returning.
		utils.AviLog.Infof("key: %s, msg: Attached HTTP redirect policy to vs %s", key, vsNode.Name)
		return
	}
	o.BuildHTTPPolicySetHTTPRequestRules(key, httpPSName, vsNode, routeModel, rule.Filters, index)
	o.BuildHTTPPolicySetHTTPResponseRules(key, vsNode, routeModel, rule.Filters, index)
	utils.AviLog.Infof("key: %s, msg: Attached HTTP policies to vs %s", key, vsNode.Name)
}

func (o *AviObjectGraph) BuildHTTPPolicySetHTTPRequestRules(key, httpPSName string, vsNode *nodes.AviEvhVsNode, routeModel RouteModel, filters []*Filter, index int) {
	requestRule := &models.HTTPRequestRule{Name: &httpPSName, Enable: proto.Bool(true), Index: proto.Int32(int32(index + 1))}
	vsNode.HttpPolicyRefs[index].RequestRules = []*models.HTTPRequestRule{}
	for _, filter := range filters {
		if filter.RequestFilter != nil {
			var j uint32 = 0
			for i := range filter.RequestFilter.Add {
				action := o.BuildHTTPPolicySetHTTPRuleHdrAction(key, "HTTP_ADD_HDR", filter.RequestFilter.Add[i], j)
				requestRule.HdrAction = append(requestRule.HdrAction, action)
				j += 1
			}

			for i := range filter.RequestFilter.Set {
				action := o.BuildHTTPPolicySetHTTPRuleHdrAction(key, "HTTP_REPLACE_HDR", filter.RequestFilter.Set[i], j)
				requestRule.HdrAction = append(requestRule.HdrAction, action)
				j += 1
			}

			for i := range filter.RequestFilter.Remove {
				action := o.BuildHTTPPolicySetHTTPRuleHdrAction(key, "HTTP_REMOVE_HDR", &Header{Name: filter.RequestFilter.Remove[i]}, j)
				requestRule.HdrAction = append(requestRule.HdrAction, action)
				j += 1
			}
		}
	}
	if len(requestRule.HdrAction) != 0 {
		vsNode.HttpPolicyRefs[index].RequestRules = append(vsNode.HttpPolicyRefs[index].RequestRules, requestRule)
		utils.AviLog.Debugf("key: %s, msg: Attached HTTP request policies %v to vs %s", key, utils.Stringify(vsNode.HttpPolicyRefs[index].RequestRules), vsNode.Name)
	}
}

func (o *AviObjectGraph) BuildHTTPPolicySetHTTPResponseRules(key string, vsNode *nodes.AviEvhVsNode, routeModel RouteModel, filters []*Filter, index int) {
	responseRule := &models.HTTPResponseRule{Name: &vsNode.Name, Enable: proto.Bool(true), Index: proto.Int32(int32(index + 1))}
	for _, filter := range filters {
		if filter.ResponseFilter != nil {
			var j uint32 = 0
			for i := range filter.ResponseFilter.Add {
				action := o.BuildHTTPPolicySetHTTPRuleHdrAction(key, "HTTP_ADD_HDR", filter.ResponseFilter.Add[i], j)
				responseRule.HdrAction = append(responseRule.HdrAction, action)
				j += 1
			}

			for i := range filter.ResponseFilter.Set {
				action := o.BuildHTTPPolicySetHTTPRuleHdrAction(key, "HTTP_REPLACE_HDR", filter.ResponseFilter.Set[i], j)
				responseRule.HdrAction = append(responseRule.HdrAction, action)
				j += 1
			}

			for i := range filter.ResponseFilter.Remove {
				action := o.BuildHTTPPolicySetHTTPRuleHdrAction(key, "HTTP_REMOVE_HDR", &Header{Name: filter.ResponseFilter.Remove[i]}, j)
				responseRule.HdrAction = append(responseRule.HdrAction, action)
				j += 1
			}
		}
	}
	if len(responseRule.HdrAction) != 0 {
		vsNode.HttpPolicyRefs[index].ResponseRules = []*models.HTTPResponseRule{responseRule}
		utils.AviLog.Debugf("key: %s, msg: Attached HTTP response policies %v to vs %s", key, utils.Stringify(vsNode.HttpPolicyRefs[index].ResponseRules), vsNode.Name)
	}
}

func (o *AviObjectGraph) BuildHTTPPolicySetHTTPRuleHdrAction(key string, action string, header *Header, headerIndex uint32) *models.HTTPHdrAction {
	hdrAction := &models.HTTPHdrAction{}
	hdrAction.Action = proto.String(action)
	hdrAction.Hdr = &models.HTTPHdrData{}
	hdrAction.Hdr.Name = proto.String(header.Name)
	hdrAction.HdrIndex = &headerIndex
	if header.Value != "" {
		hdrAction.Hdr.Value = &models.HTTPHdrValue{}
		hdrAction.Hdr.Value.IsSensitive = proto.Bool(false)
		hdrAction.Hdr.Value.Val = proto.String(header.Value)
	}
	return hdrAction
}

func (o *AviObjectGraph) BuildHTTPPolicySetHTTPRequestRedirectRules(key, httpPSname string, vsNode *nodes.AviEvhVsNode, routeModel RouteModel, filters []*Filter, index int) bool {
	redirectAction := &models.HTTPRedirectAction{}
	isRedirectPresent := false
	for _, filter := range filters {
		// considering only the first RedirectFilter
		if filter.RedirectFilter != nil {
			uriParamToken := &models.URIParamToken{
				StrValue: &filter.RedirectFilter.Host,
				Type:     proto.String("URI_TOKEN_TYPE_STRING"),
			}
			redirectAction.Host = &models.URIParam{
				Tokens: []*models.URIParamToken{uriParamToken},
				Type:   proto.String("URI_PARAM_TYPE_TOKENIZED"),
			}
			redirectAction.Protocol = proto.String("HTTP")
			statusCode := "HTTP_REDIRECT_STATUS_CODE_302"
			switch filter.RedirectFilter.StatusCode {
			case 301, 302, 307:
				statusCode = fmt.Sprintf("HTTP_REDIRECT_STATUS_CODE_%d", filter.RedirectFilter.StatusCode)
			}
			redirectAction.StatusCode = &statusCode
			requestRule := &models.HTTPRequestRule{Name: &httpPSname, Enable: proto.Bool(true), RedirectAction: redirectAction, Index: proto.Int32(int32(index + 1))}
			vsNode.HttpPolicyRefs[index].RequestRules = append(vsNode.HttpPolicyRefs[index].RequestRules, requestRule)
			isRedirectPresent = true
			utils.AviLog.Debugf("key: %s, msg: Attached HTTP request redirect policies %s to vs %s", key, utils.Stringify(vsNode.HttpPolicyRefs[index].RequestRules), vsNode.Name)
			break
		}
	}
	return isRedirectPresent
}
