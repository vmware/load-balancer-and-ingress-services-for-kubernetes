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
	"k8s.io/apimachinery/pkg/util/sets"

	akogatewayapilib "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/objects"
	akogatewayapiobjects "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/objects"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/nodes"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
)

func (o *AviObjectGraph) ProcessL7Routes(key string, routeModel RouteModel, parentNsName string, childVSes map[string]struct{}, fullsync bool) {
	httpRouteConfig := routeModel.ParseRouteConfig()
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

	found, hosts := akogatewayapiobjects.GatewayApiLister().GetGatewayRouteToHostname(parentNsName)
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
		Gateway: parentNsName,
	}
	childNode.ApplicationProfile = utils.DEFAULT_L7_APP_PROFILE
	childNode.ServiceEngineGroup = lib.GetSEGName()
	childNode.VrfContext = lib.GetVrf()
	childNode.AviMarkers = utils.AviObjectMarkers{
		GatewayName: parentName,
		Namespace:   parentNs,
		Host:        hosts,
	}
	for _, host := range hosts {
		if !strings.Contains(host, utils.WILDCARD) && !utils.HasElem(parentNode[0].VSVIPRefs[0].FQDNs, host) {
			parentNode[0].VSVIPRefs[0].FQDNs = append(parentNode[0].VSVIPRefs[0].FQDNs, host)
		}
	}

	routeTypeNsName := lib.HTTPRoute + "/" + routeModel.GetNamespace() + "/" + routeModel.GetName()
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
		}
		t1LR := lib.GetT1LRPath()
		if t1LR == "" {
			poolNode.VrfContext = lib.GetVrf()
		} else {
			poolNode.T1Lr = t1LR
			poolNode.VrfContext = ""
			utils.AviLog.Infof("key: %s, msg: setting t1LR: %s for pool node.", key, t1LR)
		}
		poolNode.NetworkPlacementSettings = lib.GetNodeNetworkMap()
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
		if childVsNode.CheckPoolNChecksum(poolNode.Name, poolNode.GetCheckSum()) {
			// Replace the poolNode.
			childVsNode.ReplaceEvhPoolInEVHNode(poolNode, key)
		}

		// TODO: Check Backend filter code. This is creating an issue in checksum calculation of object which is result in failure in deletion
		o.BuildBackendFiltersModel(key, poolName, httpbackend, childVsNode)
		pool_ref := fmt.Sprintf("/api/pool?name=%s", poolNode.Name)
		ratio := uint32(httpbackend.Backend.Weight)
		PG.Members = append(PG.Members, &models.PoolGroupMember{PoolRef: &pool_ref, Ratio: &ratio})
	}
	if len(PG.Members) > 0 {
		childVsNode.PoolGroupRefs = []*nodes.AviPoolGroupNode{PG}
		childVsNode.DefaultPoolGroup = PG.Name
	}
}

func (o *AviObjectGraph) BuildBackendFiltersModel(key, poolName string, httpbackend *HTTPBackend, vsNode *nodes.AviEvhVsNode) {
	for _, filter := range httpbackend.Filters {

		var addRequestString string
		for _, addRequestFilter := range filter.RequestFilter.Add {
			addRequestString = addRequestString + addRequestFilter.Name + ":" + addRequestFilter.Value + ","
		}
		addRequestString = strings.TrimSuffix(addRequestString, ",")
		name := lib.GetAKOUser() + "-" + akogatewayapilib.AddHeaderStringGroup
		description := "StringGroup to support ADDRequestHeaderModifier from BackendRef Filters in AKO Gateway API"
		o.AddOrUpdateStringGroupNode(key, name, description, poolName, addRequestString)

		var setRequestString string
		for _, setRequestFilter := range filter.RequestFilter.Set {
			setRequestString = setRequestString + setRequestFilter.Name + ":" + setRequestFilter.Value + ","
		}
		setRequestString = strings.TrimSuffix(setRequestString, ",")
		name = lib.GetAKOUser() + "-" + akogatewayapilib.UpdateHeaderStringGroup
		description = "StringGroup to support UpdateRequestHeaderModifier from BackendRef Filters in AKO Gateway API"
		o.AddOrUpdateStringGroupNode(key, name, description, poolName, setRequestString)

		var removeRequestString string
		for _, removeRequestKey := range filter.RequestFilter.Remove {
			removeRequestString = removeRequestString + removeRequestKey + ","
		}
		removeRequestString = strings.TrimSuffix(removeRequestString, ",")
		name = lib.GetAKOUser() + "-" + akogatewayapilib.DeleteHeaderStringGroup
		description = "StringGroup to support DeleteRequestHeaderModifier from BackendRef Filters in AKO Gateway API"
		o.AddOrUpdateStringGroupNode(key, name, description, poolName, removeRequestString)
	}
	if len(httpbackend.Filters) == 0 {
		o.UpdateStringGroupsOnRouteDeletion(key, poolName)

		//Remove datascript reference from vs if it already exists
		dsScriptNode := o.ConstructBackendFilterDataScript(key)
		var updatedHTTPDSrefs []*nodes.AviHTTPDataScriptNode
		for _, httpDsRef := range vsNode.HTTPDSrefs {
			if httpDsRef != dsScriptNode {
				updatedHTTPDSrefs = append(updatedHTTPDSrefs, httpDsRef)
			}
		}
		vsNode.HTTPDSrefs = updatedHTTPDSrefs
	}

	if httpbackend.Filters != nil && len(httpbackend.Filters) > 0 {
		dataScriptRefExists := false
		dsScriptNode := o.ConstructBackendFilterDataScript(key)
		for _, httpDsRef := range vsNode.HTTPDSrefs {
			if httpDsRef == dsScriptNode {
				dataScriptRefExists = true
			}
		}
		if !dataScriptRefExists {
			vsNode.HTTPDSrefs = append(vsNode.HTTPDSrefs, dsScriptNode)
		}
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
		o.BuildBackendFiltersModel(key, poolName, backend, parentVsNode)
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
	routeTypeNsName := fmt.Sprintf("%s/%s", httpRouteNamespace, httpRouteName)
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

func (o *AviObjectGraph) ConstructBackendFilterDataScript(key string) *nodes.AviHTTPDataScriptNode {
	datascripts := o.GetAviHTTPDSNode()
	datascriptName := akogatewayapilib.GetDataScriptName()
	for _, datascript := range datascripts {
		if datascript.Name == datascriptName {
			return datascript
		}
	}
	dsScriptNode := &nodes.AviHTTPDataScriptNode{
		Name:   datascriptName,
		Tenant: lib.GetTenant(),
		DataScript: &nodes.DataScript{
			Script: akogatewayapilib.BackendRefFilterDatascript,
			Evt:    "VS_DATASCRIPT_EVT_HTTP_LB_DONE",
		},
	}

	dsScriptNode.Script = strings.Replace(dsScriptNode.Script, "NAMEPREFIX", lib.GetAKOUser(), 3)
	dsScriptNode.StringGroups = append(dsScriptNode.StringGroups, lib.GetAKOUser()+"-"+akogatewayapilib.AddHeaderStringGroup, lib.GetAKOUser()+"-"+akogatewayapilib.UpdateHeaderStringGroup, lib.GetAKOUser()+"-"+akogatewayapilib.DeleteHeaderStringGroup)
	o.AddModelNode(dsScriptNode)
	sharedQueue := utils.SharedWorkQueue().GetQueueByName(utils.GraphLayer)
	dataScriptNamespaceName := lib.GetTenant() + "/" + datascriptName
	ok := saveAviModel(dataScriptNamespaceName, o.AviObjectGraph, key)
	if ok {
		nodes.PublishKeyToRestLayer(dataScriptNamespaceName, key, sharedQueue)
	}

	return dsScriptNode
}

func (o *AviObjectGraph) UpdateStringGroupsOnRouteDeletion(key string, poolName string) {
	addStringGroupName := lib.GetAKOUser() + "-" + akogatewayapilib.AddHeaderStringGroup
	addStringGroupDescription := "StringGroup to support ADDRequestHeaderModifier from BackendRef Filters in AKO Gateway API"
	setStringGroupName := lib.GetAKOUser() + "-" + akogatewayapilib.UpdateHeaderStringGroup
	setStringGroupDescription := "StringGroup to support UpdateRequestHeaderModifier from BackendRef Filters in AKO Gateway API"
	removeStringGroupName := lib.GetAKOUser() + "-" + akogatewayapilib.DeleteHeaderStringGroup
	removeStringGroupDescription := "StringGroup to support DeleteRequestHeaderModifier from BackendRef Filters in AKO Gateway API"

	o.AddOrUpdateStringGroupNode(key, addStringGroupName, addStringGroupDescription, poolName, "")
	o.AddOrUpdateStringGroupNode(key, setStringGroupName, setStringGroupDescription, poolName, "")
	o.AddOrUpdateStringGroupNode(key, removeStringGroupName, removeStringGroupDescription, poolName, "")
}

func (o *AviObjectGraph) RemovePoolNameFromStringGroups(currentEvhNodeName string, modelEvhNodes []*nodes.AviEvhVsNode, key string) {
	if len(modelEvhNodes) > 0 && len(modelEvhNodes[0].EvhNodes) > 0 {
		for _, modelEvhNode := range modelEvhNodes[0].EvhNodes {
			if currentEvhNodeName == modelEvhNode.Name {
				utils.AviLog.Infof("key: %s, msg: Updating stringgroups for model: %s", key, currentEvhNodeName)
				if len(modelEvhNode.PoolRefs) > 0 {
					poolname := modelEvhNode.PoolRefs[0].Name
					o.UpdateStringGroupsOnRouteDeletion(key, poolname)
				}
				return
			}
		}
	}
}
