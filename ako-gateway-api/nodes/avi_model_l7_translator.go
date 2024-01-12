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

	akogatewayapilib "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/lib"
	akogatewayapiobjects "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/objects"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/nodes"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
)

func (o *AviObjectGraph) ProcessL7Routes(key string, routeModel RouteModel, parentNsName string, childVSes map[string]struct{}) {
	for _, rule := range routeModel.ParseRouteRules().Rules {

		// TODO: add the scenarios where we will not create child VS here.
		if rule.Matches == nil || rule.Backends == nil {
			continue
		}
		o.BuildChildVS(key, routeModel, parentNsName, rule, childVSes)
	}
}

func (o *AviObjectGraph) BuildChildVS(key string, routeModel RouteModel, parentNsName string, rule *Rule, childVSes map[string]struct{}) {

	parentNode := o.GetAviEvhVS()
	parentNs, _, parentName := lib.ExtractTypeNameNamespace(parentNsName)

	found, hosts := akogatewayapiobjects.GatewayApiLister().GetGatewayRouteToHostname(parentNs, parentName)
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
		Gateway: parentName,
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
		if !utils.HasElem(parentNode[0].VSVIPRefs[0].FQDNs, host) {
			parentNode[0].VSVIPRefs[0].FQDNs = append(parentNode[0].VSVIPRefs[0].FQDNs, host)
		}
	}

	// create pg pool from the backend
	o.BuildPGPool(key, parentNsName, childNode, routeModel, rule)

	// create vhmatch from the match
	o.BuildVHMatch(key, childNode, routeModel, rule)

	// create the httppolicyset if the filter is present
	o.BuildHTTPPolicySet(key, childNode, routeModel, rule)

	foundEvhModel := nodes.FindAndReplaceEvhInModel(childNode, parentNode, key)
	if !foundEvhModel {
		parentNode[0].EvhNodes = append(parentNode[0].EvhNodes, childNode)
		utils.AviLog.Debugf("key: %s, msg: added child vs %s to the parent vs %s", key, utils.Stringify(parentNode[0].EvhNodes), childNode.VHParentName)
		akogatewayapiobjects.GatewayApiLister().UpdateRouteChildVSMappings(routeModel.GetType()+"/"+routeModel.GetNamespace()+"/"+routeModel.GetName(), childVSName)
	}
	utils.AviLog.Infof("key: %s, msg: processing of child vs %s attached to parent vs %s completed", key, childNode.Name, childNode.VHParentName)
}

func (o *AviObjectGraph) BuildPGPool(key, parentNsName string, childVsNode *nodes.AviEvhVsNode, routeModel RouteModel, rule *Rule) {

	// create the PG from backends
	routeTypeNsName := lib.HTTPRoute + "/" + routeModel.GetNamespace() + "/" + routeModel.GetName()
	parentNs, _, parentName := lib.ExtractTypeNameNamespace(parentNsName)
	_, listeners := akogatewayapiobjects.GatewayApiLister().GetRouteToGatewayListener(routeTypeNsName)
	//ListenerName/port/protocol/allowedRouteSpec
	listenerSlice := strings.Split(listeners[0], "/")
	listenerProtocol := listenerSlice[2]
	PGName := akogatewayapilib.GetPoolGroupName(parentNs, parentName,
		routeModel.GetNamespace(), routeModel.GetName(),
		utils.Stringify(rule.Matches))
	PG := &nodes.AviPoolGroupNode{
		Name:   PGName,
		Tenant: lib.GetTenant(),
	}
	for _, backend := range rule.Backends {
		poolName := akogatewayapilib.GetPoolName(parentNs, parentName,
			routeModel.GetNamespace(), routeModel.GetName(),
			utils.Stringify(rule.Matches),
			backend.Namespace, backend.Name, strconv.Itoa(int(backend.Port)))
		svcObj, err := utils.GetInformers().ServiceInformer.Lister().Services(backend.Namespace).Get(backend.Name)
		if err != nil {
			utils.AviLog.Debugf("key: %s, msg: there was an error in retrieving the service", key)
			o.RemovePoolRefsFromPG(poolName, o.GetPoolGroupByName(PGName))
			continue
		}
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
		pool_ref := fmt.Sprintf("/api/pool?name=%s", poolNode.Name)
		ratio := uint32(backend.Weight)
		PG.Members = append(PG.Members, &models.PoolGroupMember{PoolRef: &pool_ref, Ratio: &ratio})
	}
	childVsNode.PoolGroupRefs = []*nodes.AviPoolGroupNode{PG}
	childVsNode.DefaultPoolGroup = PG.Name
}

func (o *AviObjectGraph) BuildVHMatch(key string, vsNode *nodes.AviEvhVsNode, routeModel RouteModel, rule *Rule) {
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
				headerName := headerMatch.Name
				hdrMatch := &models.HdrMatch{
					MatchCase:     proto.String("SENSITIVE"),
					MatchCriteria: proto.String("HDR_EQUALS"),
					Hdr:           &headerName,
					Value:         []string{headerMatch.Value},
				}
				rule.Matches.Hdrs = append(rule.Matches.Hdrs, hdrMatch)
			}

			vhMatch.Rules = append(vhMatch.Rules, rule)
		}
		vhMatches = append(vhMatches, vhMatch)
	}
	vsNode.VHMatches = vhMatches
	utils.AviLog.Debugf("key: %s, msg: Attached match criteria %s to vs %s", key, utils.Stringify(vsNode.VHMatches), vsNode.Name)
	utils.AviLog.Infof("key: %s, msg: Attached match criteria to vs %s", key, vsNode.Name)
}

func (o *AviObjectGraph) BuildHTTPPolicySet(key string, vsNode *nodes.AviEvhVsNode, routeModel RouteModel, rule *Rule) {

	if len(rule.Filters) == 0 {
		vsNode.HttpPolicyRefs = nil
		return
	}

	policy := &nodes.AviHttpPolicySetNode{Name: vsNode.Name, Tenant: lib.GetTenant()}
	vsNode.HttpPolicyRefs = []*nodes.AviHttpPolicySetNode{policy}

	o.BuildHTTPPolicySetHTTPRequestRedirectRules(key, vsNode, routeModel, rule.Filters)
	if len(vsNode.HttpPolicyRefs[0].RequestRules) == 1 {
		// When the RedirectAction is specified the Request and Response Modify Header Action
		// won't have any effect, hence returning.
		utils.AviLog.Infof("key: %s, msg: Attached HTTP redirect policy to vs %s", key, vsNode.Name)
		return
	}
	o.BuildHTTPPolicySetHTTPRequestRules(key, vsNode, routeModel, rule.Filters)
	o.BuildHTTPPolicySetHTTPResponseRules(key, vsNode, routeModel, rule.Filters)
	utils.AviLog.Infof("key: %s, msg: Attached HTTP policies to vs %s", key, vsNode.Name)
}

func (o *AviObjectGraph) BuildHTTPPolicySetHTTPRequestRules(key string, vsNode *nodes.AviEvhVsNode, routeModel RouteModel, filters []*Filter) {
	requestRule := &models.HTTPRequestRule{Name: &vsNode.Name, Enable: proto.Bool(true), Index: proto.Int32(1)}
	for _, filter := range filters {
		if filter.RequestFilter != nil {
			for i := range filter.RequestFilter.Add {
				action := o.BuildHTTPPolicySetHTTPRuleHdrAction(key, "HTTP_ADD_HDR", filter.RequestFilter.Add[i])
				requestRule.HdrAction = append(requestRule.HdrAction, action)
			}

			for i := range filter.RequestFilter.Set {
				action := o.BuildHTTPPolicySetHTTPRuleHdrAction(key, "HTTP_REPLACE_HDR", filter.RequestFilter.Set[i])
				requestRule.HdrAction = append(requestRule.HdrAction, action)
			}

			for i := range filter.RequestFilter.Remove {
				action := o.BuildHTTPPolicySetHTTPRuleHdrAction(key, "HTTP_REMOVE_HDR", &Header{Name: filter.RequestFilter.Remove[i]})
				requestRule.HdrAction = append(requestRule.HdrAction, action)
			}
		}
	}
	if len(requestRule.HdrAction) != 0 {
		vsNode.HttpPolicyRefs[0].RequestRules = []*models.HTTPRequestRule{requestRule}
		utils.AviLog.Debugf("key: %s, msg: Attached HTTP request policies %s to vs %s", key, utils.Stringify(vsNode.HttpPolicyRefs[0].RequestRules), vsNode.Name)
	}
}

func (o *AviObjectGraph) BuildHTTPPolicySetHTTPResponseRules(key string, vsNode *nodes.AviEvhVsNode, routeModel RouteModel, filters []*Filter) {
	responseRule := &models.HTTPResponseRule{Name: &vsNode.Name, Enable: proto.Bool(true), Index: proto.Int32(1)}
	for _, filter := range filters {
		if filter.ResponseFilter != nil {
			for i := range filter.ResponseFilter.Add {
				action := o.BuildHTTPPolicySetHTTPRuleHdrAction(key, "HTTP_ADD_HDR", filter.ResponseFilter.Add[i])
				responseRule.HdrAction = append(responseRule.HdrAction, action)
			}

			for i := range filter.ResponseFilter.Set {
				action := o.BuildHTTPPolicySetHTTPRuleHdrAction(key, "HTTP_REPLACE_HDR", filter.ResponseFilter.Set[i])
				responseRule.HdrAction = append(responseRule.HdrAction, action)
			}

			for i := range filter.ResponseFilter.Remove {
				action := o.BuildHTTPPolicySetHTTPRuleHdrAction(key, "HTTP_REMOVE_HDR", &Header{Name: filter.ResponseFilter.Remove[i]})
				responseRule.HdrAction = append(responseRule.HdrAction, action)
			}
		}
	}
	if len(responseRule.HdrAction) != 0 {
		vsNode.HttpPolicyRefs[0].ResponseRules = []*models.HTTPResponseRule{responseRule}
		utils.AviLog.Debugf("key: %s, msg: Attached HTTP response policies %s to vs %s", key, utils.Stringify(vsNode.HttpPolicyRefs[0].RequestRules), vsNode.Name)
	}
}

func (o *AviObjectGraph) BuildHTTPPolicySetHTTPRuleHdrAction(key string, action string, header *Header) *models.HTTPHdrAction {
	hdrAction := &models.HTTPHdrAction{}
	hdrAction.Action = proto.String(action)
	hdrAction.Hdr = &models.HTTPHdrData{}
	hdrAction.Hdr.Name = proto.String(header.Name)
	if header.Value != "" {
		hdrAction.Hdr.Value = &models.HTTPHdrValue{}
		hdrAction.Hdr.Value.IsSensitive = proto.Bool(false)
		hdrAction.Hdr.Value.Val = proto.String(header.Value)
	}
	return hdrAction
}

func (o *AviObjectGraph) BuildHTTPPolicySetHTTPRequestRedirectRules(key string, vsNode *nodes.AviEvhVsNode, routeModel RouteModel, filters []*Filter) {
	redirectAction := &models.HTTPRedirectAction{}
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
			requestRule := &models.HTTPRequestRule{Name: &vsNode.Name, Enable: proto.Bool(true), Index: proto.Int32(1), RedirectAction: redirectAction}
			vsNode.HttpPolicyRefs[0].RequestRules = []*models.HTTPRequestRule{requestRule}
			utils.AviLog.Debugf("key: %s, msg: Attached HTTP request redirect policies %s to vs %s", key, utils.Stringify(vsNode.HttpPolicyRefs[0].RequestRules), vsNode.Name)
			break
		}
	}
}
