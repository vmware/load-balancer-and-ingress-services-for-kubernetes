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

	"github.com/vmware/alb-sdk/go/models"
	"google.golang.org/protobuf/proto"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/nodes"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
)

func (o *AviObjectGraph) ProcessL7Routes(key string, routeModel RouteModel, parentNsName string) {
	for _, rule := range routeModel.ParseRouteRules().Rules {

		// TODO: add the scenarios where we will not create child VS here.

		o.BuildChildVS(key, routeModel, parentNsName, rule)
	}
}

func (o *AviObjectGraph) BuildChildVS(key string, routeModel RouteModel, parentNsName string, rule *Rule) {

	parentNode := o.GetAviEvhVS()

	//childVSName := akogatewayapilib.GetChildName(parentNs, parentName, routeModel.GetNamespace(), routeModel.GetName(), akogatewayapilib.Encode(utils.Stringify(match)))
	childVSName := "child-vs-name" //TODO: logic to get the child name

	childNode := parentNode[0].GetEvhNodeForName(childVSName)
	if childNode == nil {
		childNode = &nodes.AviEvhVsNode{}
	}
	childNode.Name = childVSName
	childNode.VHParentName = parentNode[0].Name
	childNode.Tenant = lib.GetTenant()
	childNode.EVHParent = false

	_, _, parentName := lib.ExtractTypeNameNamespace(parentNsName)
	childNode.ServiceMetadata = lib.ServiceMetadataObj{
		Gateway: parentName,
	}
	childNode.ApplicationProfile = utils.DEFAULT_L7_APP_PROFILE //TODO: move this to the routeModel
	childNode.ServiceEngineGroup = lib.GetSEGName()
	childNode.VrfContext = lib.GetVrf()
	childNode.VHDomainNames = routeModel.ParseRouteRules().Hosts

	// TODO: add markers
	//	evhNode.AviMarkers = lib.PopulateVSNodeMarkers(namespace, host, infraSettingName)

	// create pg pool from the backend
	o.BuildPGPool(key, childNode, routeModel, rule)

	// create vhmatch from the match
	o.BuildVHMatch(key, childNode, routeModel, rule)

	// create the httppolicyset if the filter is mentioned
	o.BuildHTTPPolicySet(key, childNode, routeModel, rule)
}

func (o *AviObjectGraph) BuildPGPool(key string, childVsNode *nodes.AviEvhVsNode, routeModel RouteModel, rule *Rule) {

	// create the PG from backends

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
