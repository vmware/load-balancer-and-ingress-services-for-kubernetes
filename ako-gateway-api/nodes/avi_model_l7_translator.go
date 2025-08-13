/*
 * Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
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
	"time"

	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/vmware/alb-sdk/go/models"
	"google.golang.org/protobuf/proto"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/sets"

	akogatewayapilib "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/lib"
	akogatewayapiobjects "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/objects"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/nodes"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
)

func (o *AviObjectGraph) AddDefaultHTTPPolicySet(key string) {

	parentVS := o.GetAviEvhVS()[0]

	policyRefName := akogatewayapilib.GetDefaultHTTPPSName()
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
	defaultPolicyRef := &nodes.AviHttpPolicySetNode{Name: policyRefName, Tenant: parentVS.Tenant}
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
	httpRouteRules := httpRouteConfig.Rules
	if akogatewayapilib.IsGatewayInDedicatedMode(routeModel.GetNamespace()) {
		o.ProcessL7RoutesForDedicatedGateway(key, routeModel, httpRouteRules, parentNsName, childVSes, fullsync)
		return
	}
	for _, rule := range httpRouteRules {
		// TODO: add the scenarios where we will not create child VS here.
		if rule.Matches == nil {
			continue
		}
		if httpRouteConfig.Rejected {
			utils.AviLog.Warnf("key: %s, msg: route %s is rejected", key, routeModel.GetName())
			continue
		}
		o.BuildChildVS(key, routeModel, parentNsName, rule, childVSes, fullsync)
	}
}

func (o *AviObjectGraph) ProcessL7RoutesForDedicatedGateway(key string, routeModel RouteModel, httpRouteRules []*Rule, parentNsName string, childVSes map[string]struct{}, fullsync bool) {
	// In dedicated mode, the Gateway VS is not EVH parent but a single dedicated VS
	// All HTTPRoute rules are translated to a single HTTP PolicySet attached to this VS

	gatewayVSes := o.GetAviEvhVS()
	if len(gatewayVSes) == 0 {
		utils.AviLog.Errorf("key: %s, msg: No Gateway VS found for dedicated mode processing", key)
		return
	}

	dedicatedVS := gatewayVSes[0]
	if dedicatedVS.EVHParent {
		utils.AviLog.Errorf("key: %s, msg: Gateway VS is still in EVH mode, cannot process in dedicated mode", key)
		return
	}

	dedicatedVSName := dedicatedVS.Name
	childVSes[dedicatedVSName] = struct{}{}

	// Update metadata to include HTTPRoute information
	if dedicatedVS.ServiceMetadata.HTTPRoute == "" {
		dedicatedVS.ServiceMetadata.HTTPRoute = routeModel.GetNamespace() + "/" + routeModel.GetName()
	}

	// Update AVI markers for the HTTPRoute
	dedicatedVS.AviMarkers.HTTPRouteName = routeModel.GetName()
	dedicatedVS.AviMarkers.HTTPRouteNamespace = routeModel.GetNamespace()

	// // Get hostname mapping for this route
	// gwRouteNsName := fmt.Sprintf("%s/%s", parentNsName, routeTypeNsName)
	// found, hosts := akogatewayapiobjects.GatewayApiLister().GetGatewayRouteToHostname(gwRouteNsName)
	// if !found {
	// 	utils.AviLog.Warnf("key: %s, msg: No hosts mapped to the route %s/%s/%s", key, routeModel.GetType(), routeModel.GetNamespace(), routeModel.GetName())
	// 	return
	// }
	// dedicatedVS.AviMarkers.Host = hosts

	// Create a single HTTP PolicySet for all HTTPRoute rules
	o.BuildSingleHTTPPolicySetForDedicatedMode(key, dedicatedVS, routeModel, httpRouteRules)

	// Process each rule to create pools for the backends
	for ruleIndex, rule := range httpRouteRules {
		if rule.Matches == nil {
			utils.AviLog.Warnf("key: %s, msg: Skipping rule %d with no matches for HTTPRoute %s/%s", key, ruleIndex, routeModel.GetNamespace(), routeModel.GetName())
			continue
		}

		// Create pools for the backends in this rule
		o.BuildPGPoolForDedicatedMode(key, dedicatedVS, routeModel, rule, ruleIndex)

		utils.AviLog.Debugf("key: %s, msg: Processed rule %d for HTTPRoute %s/%s in dedicated mode", key, ruleIndex, routeModel.GetNamespace(), routeModel.GetName())
	}

	// Update route to VS mapping for dedicated mode
	akogatewayapiobjects.GatewayApiLister().UpdateRouteChildVSMappings(routeModel.GetType()+"/"+routeModel.GetNamespace()+"/"+routeModel.GetName(), dedicatedVSName)

	utils.AviLog.Infof("key: %s, msg: Completed processing HTTPRoute %s/%s in dedicated mode for VS %s", key, routeModel.GetNamespace(), routeModel.GetName(), dedicatedVSName)
}

// // BuildDedicatedModeHTTPPolicySet creates HTTP PolicySet for dedicated mode
// func (o *AviObjectGraph) BuildDedicatedModeHTTPPolicySet(key string, vsNode *nodes.AviEvhVsNode, routeModel RouteModel, rule *Rule, ruleIndex int) {
// 	// Create HTTP PolicySet name for this rule
// 	httpPSName := fmt.Sprintf("%s-rule-%d", vsNode.Name, ruleIndex)
// 	if rule.Name != "" {
// 		httpPSName = fmt.Sprintf("%s-rule-%s", vsNode.Name, rule.Name)
// 	}

// 	// Create matching criteria based on the rule
// 	var policy *nodes.AviHttpPolicySetNode
// 	for i, http := range vsNode.HttpPolicyRefs {
// 		if http.Name == httpPSName {
// 			policy = vsNode.HttpPolicyRefs[i]
// 			break
// 		}
// 	}
// 	if policy == nil {
// 		policy = &nodes.AviHttpPolicySetNode{Name: httpPSName, Tenant: vsNode.Tenant}
// 		vsNode.HttpPolicyRefs = append(vsNode.HttpPolicyRefs, policy)
// 	}

// 	// Build switching rule based on matches
// 	o.BuildDedicatedModeSwitchingRule(key, vsNode, policy, routeModel, rule, ruleIndex)

// 	// Handle filters - redirect, rewrite, header modification, etc.
// 	if len(rule.Filters) > 0 {
// 		policyIndex := len(vsNode.HttpPolicyRefs) - 1
// 		o.BuildHTTPPolicySetHTTPRequestRedirectRules(key, httpPSName, vsNode, routeModel, rule.Filters, policyIndex)
// 		o.BuildHTTPPolicySetHTTPRequestRules(key, httpPSName, vsNode, routeModel, rule.Filters, policyIndex)
// 		o.BuildHTTPPolicySetHTTPRequestUrlRewriteRules(key, httpPSName, vsNode, routeModel, rule.Filters, policyIndex)
// 		o.BuildHTTPPolicySetHTTPResponseRules(key, vsNode, routeModel, rule.Filters, policyIndex)
// 	}

// 	utils.AviLog.Debugf("key: %s, msg: Built HTTP PolicySet %s for dedicated mode", key, httpPSName)
// }

// BuildSingleHTTPPolicySetForDedicatedMode creates a single HTTP PolicySet for all HTTPRoute rules in dedicated mode
// Each match gets its own request and response rules with specific match criteria
func (o *AviObjectGraph) BuildSingleHTTPPolicySetForDedicatedMode(key string, vsNode *nodes.AviEvhVsNode, routeModel RouteModel, httpRouteRules []*Rule) {
	if len(httpRouteRules) == 0 {
		return
	}

	// Create HTTP PolicySet name for the entire HTTPRoute
	httpPSName := fmt.Sprintf("%s-%s-%s-httproute", vsNode.Name, routeModel.GetNamespace(), routeModel.GetName())

	// Check if policy already exists, if so, reuse it
	var policy *nodes.AviHttpPolicySetNode
	for i, http := range vsNode.HttpPolicyRefs {
		if http.Name == httpPSName {
			policy = vsNode.HttpPolicyRefs[i]
			break
		}
	}
	if policy == nil {
		policy = &nodes.AviHttpPolicySetNode{Name: httpPSName, Tenant: vsNode.Tenant}
		vsNode.HttpPolicyRefs = append(vsNode.HttpPolicyRefs, policy)
	}

	// Clear existing rules to rebuild them completely
	policy.RequestRules = []*models.HTTPRequestRule{}
	policy.ResponseRules = []*models.HTTPResponseRule{}

	var requestRuleIndex int32 = 0
	var responseRuleIndex int32 = 1000 // Response rules start at higher indices

	// Process each rule and create request/response rules for each match
	for ruleIndex, rule := range httpRouteRules {
		if rule.Matches == nil {
			utils.AviLog.Warnf("key: %s, msg: Skipping rule %d with no matches for HTTPRoute %s/%s", key, ruleIndex, routeModel.GetNamespace(), routeModel.GetName())
			continue
		}

		// Create HTTP request rules with match criteria for each match
		for matchIndex, match := range rule.Matches {
			// Build HTTP request rule with match, filters, and switching action
			o.BuildHTTPRequestRuleWithMatch(key, policy, routeModel, rule, ruleIndex, matchIndex, match, &requestRuleIndex)
		}

		// Create HTTP response rules with match criteria for each match (if response filters exist)
		responseFilters := o.GetResponseFilters(rule.Filters)
		if len(responseFilters) > 0 {
			for matchIndex, match := range rule.Matches {
				// Build HTTP response rule with match and response filters
				o.BuildHTTPResponseRuleWithMatch(key, policy, routeModel, responseFilters, ruleIndex, matchIndex, match, &responseRuleIndex)
			}
		}
	}

	utils.AviLog.Debugf("key: %s, msg: Built single HTTP PolicySet %s for dedicated mode with %d request rules and %d response rules",
		key, httpPSName, len(policy.RequestRules), len(policy.ResponseRules))
}

// GetResponseFilters extracts response filters from the filter list
func (o *AviObjectGraph) GetResponseFilters(filters []*Filter) []*Filter {
	var responseFilters []*Filter
	for _, filter := range filters {
		if filter.ResponseFilter != nil {
			responseFilters = append(responseFilters, filter)
		}
	}
	return responseFilters
}

// BuildHTTPRequestRuleWithMatch creates an HTTP request rule with specific match criteria, filters, and switching action
func (o *AviObjectGraph) BuildHTTPRequestRuleWithMatch(key string, policy *nodes.AviHttpPolicySetNode, routeModel RouteModel, rule *Rule, ruleIndex, matchIndex int, match *Match, requestRuleIndex *int32) {
	ruleName := fmt.Sprintf("rule-%d-match-%d", ruleIndex, matchIndex)

	// Create the HTTP request rule
	httpRequestRule := &models.HTTPRequestRule{
		Name:   &ruleName,
		Enable: proto.Bool(true),
		Index:  proto.Int32(*requestRuleIndex),
		Match:  &models.MatchTarget{},
	}

	// Build match criteria
	o.BuildMatchTarget(httpRequestRule.Match, match)

	// Apply request filters (redirects, header modifications, URL rewrites)
	for _, filter := range rule.Filters {
		// Handle redirect filter
		if filter.RedirectFilter != nil {
			redirect := &models.HTTPRedirectAction{}
			if filter.RedirectFilter.Host != "" {
				redirect.Host = &models.URIParam{
					Type:   proto.String("HTTP_TYPE_HOST"),
					Tokens: []*models.URIParamToken{{Type: proto.String("URI_TOKEN_TYPE_STRING"), StrValue: &filter.RedirectFilter.Host}},
				}
			}
			if filter.RedirectFilter.StatusCode != 0 {
				redirect.StatusCode = proto.String(fmt.Sprintf("HTTP_REDIRECT_%d", filter.RedirectFilter.StatusCode))
			}
			httpRequestRule.RedirectAction = redirect
		}

		// Handle request header modifications
		if filter.RequestFilter != nil {
			var j uint32 = 0
			for i := range filter.RequestFilter.Add {
				action := o.BuildHTTPPolicySetHTTPRuleHdrAction(key, "HTTP_ADD_HDR", filter.RequestFilter.Add[i], j)
				httpRequestRule.HdrAction = append(httpRequestRule.HdrAction, action)
				j++
			}
			for i := range filter.RequestFilter.Set {
				action := o.BuildHTTPPolicySetHTTPRuleHdrAction(key, "HTTP_REPLACE_HDR", filter.RequestFilter.Set[i], j)
				httpRequestRule.HdrAction = append(httpRequestRule.HdrAction, action)
				j++
			}
			for i := range filter.RequestFilter.Remove {
				action := o.BuildHTTPPolicySetHTTPRuleHdrAction(key, "HTTP_REMOVE_HDR", &Header{Name: filter.RequestFilter.Remove[i]}, j)
				httpRequestRule.HdrAction = append(httpRequestRule.HdrAction, action)
				j++
			}
		}

		// Handle URL rewrite
		if filter.UrlRewriteFilter != nil {
			rewriteAction := &models.HTTPRewriteURLAction{}
			if filter.UrlRewriteFilter.path != nil {
				var pathValue string
				if filter.UrlRewriteFilter.path.ReplaceFullPath != nil {
					pathValue = *filter.UrlRewriteFilter.path.ReplaceFullPath
				} else if filter.UrlRewriteFilter.path.ReplacePrefixMatch != nil {
					pathValue = *filter.UrlRewriteFilter.path.ReplacePrefixMatch
				}
				if pathValue != "" {
					rewriteAction.Path = &models.URIParam{
						Type:   proto.String("HTTP_TYPE_PATH"),
						Tokens: []*models.URIParamToken{{Type: proto.String("URI_TOKEN_TYPE_STRING"), StrValue: &pathValue}},
					}
				}
			}
			if filter.UrlRewriteFilter.hostname != "" {
				rewriteAction.HostHdr = &models.URIParam{
					Type:   proto.String("HTTP_TYPE_HOST"),
					Tokens: []*models.URIParamToken{{Type: proto.String("URI_TOKEN_TYPE_STRING"), StrValue: &filter.UrlRewriteFilter.hostname}},
				}
			}
			httpRequestRule.RewriteURLAction = rewriteAction
		}
	}

	// Add switching action to backends
	if len(rule.Backends) > 0 {
		poolGroupName := o.GetPoolGroupNameForRule(routeModel, rule, ruleIndex)
		switchAction := &models.HttpswitchingAction{
			Action:       proto.String("HTTP_SWITCHING_SELECT_POOLGROUP"),
			PoolGroupRef: proto.String(fmt.Sprintf("/api/poolgroup?name=%s", poolGroupName)),
		}
		httpRequestRule.SwitchingAction = switchAction
	}

	// Add the rule to the policy
	policy.RequestRules = append(policy.RequestRules, httpRequestRule)
	*requestRuleIndex++

	utils.AviLog.Debugf("key: %s, msg: Created HTTP request rule %s with match criteria", key, ruleName)
}

// BuildHTTPResponseRuleWithMatch creates an HTTP response rule with specific match criteria and response filters
func (o *AviObjectGraph) BuildHTTPResponseRuleWithMatch(key string, policy *nodes.AviHttpPolicySetNode, routeModel RouteModel, responseFilters []*Filter, ruleIndex, matchIndex int, match *Match, responseRuleIndex *int32) {
	ruleName := fmt.Sprintf("response-rule-%d-match-%d", ruleIndex, matchIndex)

	// Create the HTTP response rule
	httpResponseRule := &models.HTTPResponseRule{
		Name:   &ruleName,
		Enable: proto.Bool(true),
		Index:  proto.Int32(*responseRuleIndex),
		Match:  &models.ResponseMatchTarget{},
	}

	// Build match criteria for response rule (converting request match to response match)
	o.BuildResponseMatchTarget(httpResponseRule.Match, match)

	// Apply response filters
	for _, filter := range responseFilters {
		if filter.ResponseFilter != nil {
			var j uint32 = 0
			// Handle response header additions
			for i := range filter.ResponseFilter.Add {
				action := o.BuildHTTPPolicySetHTTPRuleHdrAction(key, "HTTP_ADD_HDR", filter.ResponseFilter.Add[i], j)
				httpResponseRule.HdrAction = append(httpResponseRule.HdrAction, action)
				j++
			}
			// Handle response header replacements
			for i := range filter.ResponseFilter.Set {
				action := o.BuildHTTPPolicySetHTTPRuleHdrAction(key, "HTTP_REPLACE_HDR", filter.ResponseFilter.Set[i], j)
				httpResponseRule.HdrAction = append(httpResponseRule.HdrAction, action)
				j++
			}
			// Handle response header removals
			for i := range filter.ResponseFilter.Remove {
				action := o.BuildHTTPPolicySetHTTPRuleHdrAction(key, "HTTP_REMOVE_HDR", &Header{Name: filter.ResponseFilter.Remove[i]}, j)
				httpResponseRule.HdrAction = append(httpResponseRule.HdrAction, action)
				j++
			}
		}
	}

	// Only add the rule if there are actual response actions
	if len(httpResponseRule.HdrAction) > 0 {
		policy.ResponseRules = append(policy.ResponseRules, httpResponseRule)
		*responseRuleIndex++
		utils.AviLog.Debugf("key: %s, msg: Created HTTP response rule %s with match criteria", key, ruleName)
	}
}

// BuildMatchTarget builds the MatchTarget for HTTP request rules
func (o *AviObjectGraph) BuildMatchTarget(matchTarget *models.MatchTarget, match *Match) {
	// Handle path matching
	if match.PathMatch != nil {
		matchTarget.Path = &models.PathMatch{
			MatchCase: proto.String("SENSITIVE"),
			MatchStr:  []string{match.PathMatch.Path},
		}
		if match.PathMatch.Type == akogatewayapilib.EXACT {
			matchTarget.Path.MatchCriteria = proto.String("EQUALS")
		} else if match.PathMatch.Type == akogatewayapilib.PATHPREFIX {
			matchTarget.Path.MatchCriteria = proto.String("BEGINS_WITH")
		} else if match.PathMatch.Type == akogatewayapilib.REGULAREXPRESSION {
			matchTarget.Path.MatchCriteria = proto.String("REGEX")
		}
	}

	// Handle header matching
	if match.HeaderMatch != nil {
		for _, header := range match.HeaderMatch {
			hdrMatch := &models.HdrMatch{
				Hdr:       &header.Name,
				MatchCase: proto.String("SENSITIVE"),
			}
			if header.Type == akogatewayapilib.EXACT {
				hdrMatch.MatchCriteria = proto.String("HDR_EQUALS")
				hdrMatch.Value = []string{header.Value}
			} else if header.Type == akogatewayapilib.REGULAREXPRESSION {
				hdrMatch.MatchCriteria = proto.String("HDR_REGEX")
				hdrMatch.Value = []string{header.Value}
			}
			matchTarget.Hdrs = append(matchTarget.Hdrs, hdrMatch)
		}
	}

}

// BuildResponseMatchTarget builds the ResponseMatchTarget for HTTP response rules
// This converts request match criteria to response match criteria where applicable
func (o *AviObjectGraph) BuildResponseMatchTarget(responseMatchTarget *models.ResponseMatchTarget, match *Match) {
	// Copy over applicable match criteria that can be used for response matching

	// Path matching (applicable to response)
	if match.PathMatch != nil {
		responseMatchTarget.Path = &models.PathMatch{
			MatchCase: proto.String("SENSITIVE"),
			MatchStr:  []string{match.PathMatch.Path},
		}
		if match.PathMatch.Type == akogatewayapilib.EXACT {
			responseMatchTarget.Path.MatchCriteria = proto.String("EQUALS")
		} else if match.PathMatch.Type == akogatewayapilib.PATHPREFIX {
			responseMatchTarget.Path.MatchCriteria = proto.String("BEGINS_WITH")
		} else if match.PathMatch.Type == akogatewayapilib.REGULAREXPRESSION {
			responseMatchTarget.Path.MatchCriteria = proto.String("REGEX")
		}
	}

	// Header matching (applicable to response - can match request headers)
	if match.HeaderMatch != nil {
		for _, header := range match.HeaderMatch {
			hdrMatch := &models.HdrMatch{
				Hdr:       &header.Name,
				MatchCase: proto.String("SENSITIVE"),
			}
			if header.Type == akogatewayapilib.EXACT {
				hdrMatch.MatchCriteria = proto.String("HDR_EQUALS")
				hdrMatch.Value = []string{header.Value}
			} else if header.Type == akogatewayapilib.REGULAREXPRESSION {
				hdrMatch.MatchCriteria = proto.String("HDR_REGEX")
				hdrMatch.Value = []string{header.Value}
			}
			responseMatchTarget.Hdrs = append(responseMatchTarget.Hdrs, hdrMatch)
		}
	}
}

// GetPoolGroupNameForRule generates the pool group name for a rule
func (o *AviObjectGraph) GetPoolGroupNameForRule(routeModel RouteModel, rule *Rule, ruleIndex int) string {
	if rule.Name == "" {
		return akogatewayapilib.GetPoolGroupName("", "",
			routeModel.GetNamespace(), routeModel.GetName(),
			utils.Stringify(rule.Matches))
	} else {
		return akogatewayapilib.GetPoolGroupName("", "",
			routeModel.GetNamespace(), routeModel.GetName(),
			rule.Name)
	}
}

// BuildDedicatedModeSwitchingRule creates the switching rule with path/header matching
func (o *AviObjectGraph) BuildDedicatedModeSwitchingRule(key string, vsNode *nodes.AviEvhVsNode, policy *nodes.AviHttpPolicySetNode, routeModel RouteModel, rule *Rule, ruleIndex int) {
	// Build pool group name for this rule's backends
	poolGroupName := fmt.Sprintf("%s-rule-%d-pg", vsNode.Name, ruleIndex)
	if rule.Name != "" {
		poolGroupName = fmt.Sprintf("%s-rule-%s-pg", vsNode.Name, rule.Name)
	}

	for matchIndex, match := range rule.Matches {
		requestRule := &models.HTTPRequestRule{
			Name:   proto.String(fmt.Sprintf("rule-%d-match-%d", ruleIndex, matchIndex)),
			Enable: proto.Bool(true),
			Index:  proto.Int32(int32(ruleIndex*100 + matchIndex)), // Ensure proper ordering
		}

		// Build match criteria
		matchTarget := &models.MatchTarget{}

		// Path matching
		if match.PathMatch != nil {
			pathMatch := &models.PathMatch{}
			switch match.PathMatch.Type {
			case akogatewayapilib.EXACT:
				pathMatch.MatchCriteria = proto.String("EQUALS")
			case akogatewayapilib.PATHPREFIX:
				pathMatch.MatchCriteria = proto.String("BEGINS_WITH")
			case akogatewayapilib.REGULAREXPRESSION:
				pathMatch.MatchCriteria = proto.String("REGEX_MATCH")
			default:
				pathMatch.MatchCriteria = proto.String("BEGINS_WITH")
			}
			pathMatch.MatchStr = []string{match.PathMatch.Path}
			matchTarget.Path = pathMatch
		}

		// Header matching
		if len(match.HeaderMatch) > 0 {
			var hdrs []*models.HdrMatch
			for _, hdrMatch := range match.HeaderMatch {
				hdr := &models.HdrMatch{
					MatchCase:     proto.String("SENSITIVE"),
					Hdr:           &hdrMatch.Name,
					MatchCriteria: proto.String("HDR_EQUALS"),
					Value:         []string{hdrMatch.Value},
				}
				hdrs = append(hdrs, hdr)
			}
			matchTarget.Hdrs = hdrs
		}

		requestRule.Match = matchTarget

		// Set switching action to pool group
		poolGroupRef := fmt.Sprintf("/api/poolgroup?name=%s", poolGroupName)
		requestRule.SwitchingAction = &models.HttpswitchingAction{
			Action:       proto.String("HTTP_SWITCHING_SELECT_POOLGROUP"),
			PoolGroupRef: &poolGroupRef,
		}

		policy.RequestRules = append(policy.RequestRules, requestRule)
	}
}

// BuildDedicatedModeSwitchingRuleForSinglePolicy creates the switching rule with path/header matching for single policy set
func (o *AviObjectGraph) BuildDedicatedModeSwitchingRuleForSinglePolicy(key string, vsNode *nodes.AviEvhVsNode, policy *nodes.AviHttpPolicySetNode, routeModel RouteModel, rule *Rule, ruleIndex int) {
	// Build pool group name for this rule's backends
	poolGroupName := fmt.Sprintf("%s-rule-%d-pg", vsNode.Name, ruleIndex)
	if rule.Name != "" {
		poolGroupName = fmt.Sprintf("%s-rule-%s-pg", vsNode.Name, rule.Name)
	}

	for matchIndex, match := range rule.Matches {
		requestRule := &models.HTTPRequestRule{
			Name:   proto.String(fmt.Sprintf("rule-%d-match-%d", ruleIndex, matchIndex)),
			Enable: proto.Bool(true),
			Index:  proto.Int32(int32(ruleIndex*100 + matchIndex)), // Ensure proper ordering
		}

		// Build match criteria
		matchTarget := &models.MatchTarget{}

		// Path matching
		if match.PathMatch != nil {
			pathMatch := &models.PathMatch{}
			switch match.PathMatch.Type {
			case akogatewayapilib.EXACT:
				pathMatch.MatchCriteria = proto.String("EQUALS")
			case akogatewayapilib.PATHPREFIX:
				pathMatch.MatchCriteria = proto.String("BEGINS_WITH")
			case akogatewayapilib.REGULAREXPRESSION:
				pathMatch.MatchCriteria = proto.String("REGEX_MATCH")
			default:
				pathMatch.MatchCriteria = proto.String("BEGINS_WITH")
			}
			pathMatch.MatchStr = []string{match.PathMatch.Path}
			matchTarget.Path = pathMatch
		}

		// Header matching
		if len(match.HeaderMatch) > 0 {
			var hdrs []*models.HdrMatch
			for _, hdrMatch := range match.HeaderMatch {
				hdr := &models.HdrMatch{
					MatchCase:     proto.String("SENSITIVE"),
					Hdr:           &hdrMatch.Name,
					MatchCriteria: proto.String("HDR_EQUALS"),
					Value:         []string{hdrMatch.Value},
				}
				hdrs = append(hdrs, hdr)
			}
			matchTarget.Hdrs = hdrs
		}

		requestRule.Match = matchTarget

		// Set switching action to pool group
		poolGroupRef := fmt.Sprintf("/api/poolgroup?name=%s", poolGroupName)
		requestRule.SwitchingAction = &models.HttpswitchingAction{
			Action:       proto.String("HTTP_SWITCHING_SELECT_POOLGROUP"),
			PoolGroupRef: &poolGroupRef,
		}

		policy.RequestRules = append(policy.RequestRules, requestRule)
	}
}

// BuildPGPoolForDedicatedMode creates pool groups and pools for dedicated mode
// In dedicated mode, pools are attached to HTTP PolicySets, not directly to VS
func (o *AviObjectGraph) BuildPGPoolForDedicatedMode(key string, vsNode *nodes.AviEvhVsNode, routeModel RouteModel, rule *Rule, ruleIndex int) {
	if len(rule.Backends) == 0 {
		utils.AviLog.Debugf("key: %s, msg: No backends found for rule %d in HTTPRoute %s/%s", key, ruleIndex, routeModel.GetNamespace(), routeModel.GetName())
		return
	}

	// Create pool group name for this rule (same naming as in BuildDedicatedModeSwitchingRule)
	poolGroupName := fmt.Sprintf("%s-rule-%d-pg", vsNode.Name, ruleIndex)
	if rule.Name != "" {
		poolGroupName = fmt.Sprintf("%s-rule-%s-pg", vsNode.Name, rule.Name)
	}

	// Create pool group for this rule's backends
	poolGroupNode := &nodes.AviPoolGroupNode{
		Name:   poolGroupName,
		Tenant: vsNode.Tenant,
	}

	// Set pool group markers
	poolGroupNode.AviMarkers = utils.AviObjectMarkers{
		GatewayName:        vsNode.AviMarkers.GatewayName,
		GatewayNamespace:   vsNode.AviMarkers.GatewayNamespace,
		HTTPRouteName:      routeModel.GetName(),
		HTTPRouteNamespace: routeModel.GetNamespace(),
	}
	if rule.Name != "" {
		poolGroupNode.AviMarkers.HTTPRouteRuleName = rule.Name
	}

	// Calculate total weight for ratio calculation
	totalWeight := int32(0)
	for _, httpbackend := range rule.Backends {
		totalWeight += httpbackend.Backend.Weight
	}

	// Create pools for each backend
	for backendIndex, httpbackend := range rule.Backends {
		backend := httpbackend.Backend
		poolName := fmt.Sprintf("%s-backend-%d", poolGroupName, backendIndex)
		if backend.Name != "" {
			poolName = fmt.Sprintf("%s-%s-%d", poolGroupName, backend.Name, backend.Port)
		}

		// Get service object for pool population
		svcObj, err := utils.GetInformers().ServiceInformer.Lister().Services(backend.Namespace).Get(backend.Name)
		if err != nil {
			utils.AviLog.Warnf("key: %s, msg: Failed to retrieve service %s/%s for pool %s, err: %v",
				key, backend.Namespace, backend.Name, poolName, err)
			continue
		}

		// Create pool node with full service details
		poolNode := &nodes.AviPoolNode{
			Name:       poolName,
			Tenant:     vsNode.Tenant,
			VrfContext: vsNode.VrfContext,
			Protocol:   "HTTP", // Default for L7
			Port:       backend.Port,
			PortName:   akogatewayapilib.FindPortName(backend.Name, backend.Namespace, backend.Port, key),
			TargetPort: akogatewayapilib.FindTargetPort(backend.Name, backend.Namespace, backend.Port, key),
			ServiceMetadata: lib.ServiceMetadataObj{
				NamespaceServiceName: []string{backend.Namespace + "/" + backend.Name},
			},
		}

		// Set pool markers
		poolNode.AviMarkers = utils.AviObjectMarkers{
			GatewayName:        vsNode.AviMarkers.GatewayName,
			GatewayNamespace:   vsNode.AviMarkers.GatewayNamespace,
			HTTPRouteName:      routeModel.GetName(),
			HTTPRouteNamespace: routeModel.GetNamespace(),
			BackendNs:          backend.Namespace,
			BackendName:        backend.Name,
		}
		if rule.Name != "" {
			poolNode.AviMarkers.HTTPRouteRuleName = rule.Name
		}

		// Populate servers for the pool
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

		// Apply infrastructure settings (T1LR, etc.)
		t1LR := lib.GetT1LRPath()
		if found, infraSettingName := akogatewayapiobjects.GatewayApiLister().GetGatewayToAviInfraSetting(vsNode.AviMarkers.GatewayNamespace + "/" + vsNode.AviMarkers.GatewayName); found {
			if infraSetting, err := akogatewayapilib.AKOControlConfig().AviInfraSettingInformer().Lister().Get(infraSettingName); err != nil {
				utils.AviLog.Warnf("key: %s, msg: failed to retrieve AviInfraSetting %s, err: %s", key, infraSettingName, err.Error())
			} else if infraSetting != nil && infraSetting.Status.Status == lib.StatusAccepted && infraSetting.Spec.NSXSettings.T1LR != nil {
				t1LR = *infraSetting.Spec.NSXSettings.T1LR
			}
		}

		if t1LR != "" {
			poolNode.T1Lr = t1LR
			poolNode.VrfContext = ""
			utils.AviLog.Infof("key: %s, msg: setting t1LR: %s for pool node %s", key, t1LR, poolName)
		}

		// Set network placement settings
		poolNode.NetworkPlacementSettings = lib.GetNodeNetworkMap()

		// Calculate pool ratio based on weight
		ratio := uint32(1) // Default ratio
		if totalWeight > 0 {
			ratio = uint32(backend.Weight)
		}

		// Add pool reference to pool group
		poolRef := fmt.Sprintf("/api/pool?name=%s", poolName)
		poolGroupNode.Members = append(poolGroupNode.Members, &models.PoolGroupMember{
			PoolRef: &poolRef,
			Ratio:   &ratio,
		})

		// Add pool to VS (pools are still attached to VS for management)
		vsNode.PoolRefs = append(vsNode.PoolRefs, poolNode)

		utils.AviLog.Debugf("key: %s, msg: Created pool %s with ratio %d for dedicated mode rule %d",
			key, poolName, ratio, ruleIndex)
	}

	// NOTE: Pool group is NOT attached to VS directly in dedicated mode
	// It will be referenced by HTTP PolicySet switching rules created in BuildDedicatedModeSwitchingRule

	// Instead, we add the pool group to a temporary collection for HTTP PolicySet reference
	// The HTTP PolicySet switching rules will reference this pool group by name
	vsNode.PoolGroupRefs = append(vsNode.PoolGroupRefs, poolGroupNode)

	utils.AviLog.Debugf("key: %s, msg: Built pool group %s for dedicated mode rule %d (attached via HTTP PolicySet)",
		key, poolGroupName, ruleIndex)
}

// ApplyRuleExtensionRefsForDedicatedMode handles extension refs for dedicated mode
func (o *AviObjectGraph) ApplyRuleExtensionRefsForDedicatedMode(key string, vsNode *nodes.AviEvhVsNode, routeModel RouteModel, rule *Rule) {
	// This would handle any extension refs similar to how it's done for child VS
	// For now, just log that extension refs are not yet fully implemented for dedicated mode
	if len(rule.Filters) > 0 {
		for _, filter := range rule.Filters {
			if filter.ExtensionRef != nil {
				utils.AviLog.Debugf("key: %s, msg: Extension ref %s/%s found for dedicated mode (implementation pending)",
					key, filter.ExtensionRef.Kind, filter.ExtensionRef.Name)
			}
		}
	}
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
	var childVSName string
	if rule.Name == "" {
		childVSName = akogatewayapilib.GetChildName(parentNs, parentName, routeModel.GetNamespace(), routeModel.GetName(), utils.Stringify(rule.Matches))
	} else {
		childVSName = akogatewayapilib.GetChildName(parentNs, parentName, routeModel.GetNamespace(), routeModel.GetName(), rule.Name)
	}
	childVSes[childVSName] = struct{}{}

	childNode := parentNode[0].GetEvhNodeForName(childVSName)
	if childNode == nil {
		childNode = &nodes.AviEvhVsNode{}
	}
	childNode.Name = childVSName
	childNode.VHParentName = parentNode[0].Name
	childNode.Tenant = parentNode[0].Tenant
	childNode.EVHParent = false

	childNode.ServiceMetadata = lib.ServiceMetadataObj{
		Gateway:   parentNsName,
		HTTPRoute: routeModel.GetNamespace() + "/" + routeModel.GetName(),
	}
	childNode.ApplicationProfileRef = proto.String(fmt.Sprintf("/api/applicationprofile/?name=%s", utils.DEFAULT_L7_APP_PROFILE))
	childNode.ServiceEngineGroup = lib.GetSEGName()
	childNode.VrfContext = lib.GetVrf()
	childNode.AviMarkers = utils.AviObjectMarkers{
		GatewayName:        parentName,
		GatewayNamespace:   parentNs,
		HTTPRouteName:      routeModel.GetName(),
		HTTPRouteNamespace: routeModel.GetNamespace(),
		Host:               hosts,
	}
	if rule.Name != "" {
		childNode.AviMarkers.HTTPRouteRuleName = rule.Name
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
	o.BuildHTTPPolicySet(key, childNode, routeModel, rule, 0, childVSName)
	// Apply Extension Ref
	o.ApplyRuleExtensionRefs(key, childNode, routeModel, rule)
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

	utils.AviLog.Debugf("key: %s, unique hostnames %v found for gatewayNs: %s", key, uniqueHostnames, parentNsName)
	parentNode.VSVIPRefs[0].FQDNs = uniqueHostnames
}

// parseGatewayDurationToMinutes converts Gateway API Duration string to minutes (int32).
// Returns nil if duration is nil.
// Returns 0 (Avi's representation for infinite) if parsing fails
// Returns 1 if duration is positive but < 1 minute
// Avi specific: Timeout for HTTPCookiePersistenceProfile: Allowed values are 1-14400. Special values are 0- No Timeout.
func parseGatewayDurationToMinutes(key string, gwDuration *gatewayv1.Duration) *int32 {
	if gwDuration == nil {
		return nil
	}
	zeroTimeout := int32(0)
	d, err := time.ParseDuration(string(*gwDuration))
	if err != nil {
		utils.AviLog.Warnf("key: %s, msg: failed to parse duration string '%s': %v. Defaulting to Avi's infinite timeout.", key, *gwDuration, err)
		return &zeroTimeout
	}

	if d < 0 {
		utils.AviLog.Warnf("key: %s, msg: negative duration '%s' is not supported. Defaulting to Avi's infinite timeout.", key, *gwDuration)
		return &zeroTimeout
	}

	minutes := int32(d.Minutes())
	// If duration is positive but less than 1 minute, Avi cannot represent it precisely as it takes minutes so we will translate it to 1 minute
	if d > 0 && minutes == 0 {
		utils.AviLog.Warnf("key: %s, msg: duration %s is positive but less than 1 minute. Avi's minimum granularity for persistence timeout is 1 minute. Defaulting to Avi's 1 minute).", key, *gwDuration)
		oneTimeout := int32(1)
		return &oneTimeout
	}
	// Avi's timeout range is 1-14400 minutes, or 0 for infinite.
	if minutes > 14400 {
		utils.AviLog.Warnf("key: %s, msg: duration %s (%d minutes) exceeds Avi's max persistence timeout of 14400 minutes. Clamping to 14400.", key, *gwDuration, minutes)
		maxTimeout := int32(14400)
		return &maxTimeout
	}

	return &minutes
}

func (o *AviObjectGraph) BuildApplicationPersistenceProfile(key string, rule *Rule, routeModel RouteModel, parentNs, parentName string, markers utils.AviObjectMarkers) *nodes.AviApplicationPersistenceProfileNode {
	sp := rule.SessionPersistence
	persistProfileNode := &nodes.AviApplicationPersistenceProfileNode{
		AviMarkers: markers,
	}

	persistenceType := gatewayv1.CookieBasedSessionPersistence // Default as per Gateway API spec
	if sp.Type != nil {
		persistenceType = *sp.Type
	}

	switch persistenceType {
	case gatewayv1.CookieBasedSessionPersistence:
		persistProfileNode.PersistenceType = "PERSISTENCE_TYPE_HTTP_COOKIE"
		httpCookiePersistenceProfileNode := &nodes.HTTPCookiePersistenceProfileNode{
			CookieName: *sp.SessionName,
		}
		httpCookiePersistenceProfileNode.Timeout = parseGatewayDurationToMinutes(key, sp.AbsoluteTimeout)
		if sp.CookieConfig != nil && sp.CookieConfig.LifetimeType != nil &&
			*sp.CookieConfig.LifetimeType == gatewayv1.PermanentCookieLifetimeType {
			httpCookiePersistenceProfileNode.IsPersistentCookie = proto.Bool(true)
		} else { // Default LifetimeType is Session
			httpCookiePersistenceProfileNode.IsPersistentCookie = proto.Bool(false)
		}
		persistProfileNode.HTTPCookiePersistenceProfile = httpCookiePersistenceProfileNode

	default:
		utils.AviLog.Errorf("key: %s, msg: unsupported session persistence type: %s in route %s/%s. No persistence profile will be applied.", key, persistenceType, routeModel.GetNamespace(), routeModel.GetName())
		return nil
	}
	var appPersistProfileName string
	if rule.Name == "" {
		appPersistProfileName = akogatewayapilib.GetPersistenceProfileName(parentNs, parentName,
			routeModel.GetNamespace(), routeModel.GetName(),
			utils.Stringify(rule.Matches), persistProfileNode.PersistenceType)
	} else {
		appPersistProfileName = akogatewayapilib.GetPersistenceProfileName(parentNs, parentName,
			routeModel.GetNamespace(), routeModel.GetName(),
			rule.Name, persistProfileNode.PersistenceType)
	}
	persistProfileNode.Name = appPersistProfileName
	return persistProfileNode
}

func (o *AviObjectGraph) ApplyRuleExtensionRefs(key string, childNode *nodes.AviEvhVsNode, routeModel RouteModel, rule *Rule) {

	if rule != nil && rule.Filters != nil {
		isFilterAppProfSet := false
		for _, filter := range rule.Filters {
			if filter.ExtensionRef != nil {
				// validations are already done. Here we need to just apply Extension refs
				// Priortity: Individual ako-crd-oprator CRD have higher priority over
				// same kind of the object present in AKO defined CRD.
				// Application Profile CRD is to be given higher priority.

				switch filter.ExtensionRef.Kind {
				case lib.ApplicationProfile:
					appProfRef := fmt.Sprintf("/api/applicationprofile/?name=%s-%s-%s", lib.GetClusterName(), routeModel.GetNamespace(), filter.ExtensionRef.Name)
					childNode.ApplicationProfileRef = &appProfRef
					isFilterAppProfSet = true
				case lib.L7Rule:
					err := akogatewayapilib.ParseL7CRD(key, routeModel.GetNamespace(), filter.ExtensionRef.Name, childNode, isFilterAppProfSet)
					if err != nil {
						resetChildNodeFields(key, err, childNode, isFilterAppProfSet)
					}
				}
			}
		}
	}
}
func resetChildNodeFields(key string, err error, childNode *nodes.AviEvhVsNode, isFilterAppProfSet bool) {
	utils.AviLog.Warnf("key: %s, msg: Error while parsing extension ref: %s. Resetting child VS %s fields", key, err.Error(), childNode.Name)
	generatedFields := childNode.GetGeneratedFields()
	generatedFields.ConvertL7RuleFieldsToNil()
	childNode.SetAnalyticsPolicy(nil)
	childNode.SetAnalyticsProfileRef(nil)
	if !isFilterAppProfSet {
		childNode.SetAppProfileRef(nil)
	}
	childNode.SetErrorPageProfileRef("")
	childNode.SetICAPProfileRefs([]string{})
	childNode.SetWafPolicyRef(nil)
	childNode.SetHttpPolicySetRefs([]string{})
}

func resetPoolNodeFields(key string, err error, poolNode *nodes.AviPoolNode, isFilterHMSet bool) {
	utils.AviLog.Warnf("key: %s, msg: Error while parsing extension ref: %s. Resetting pool %s fields", key, err.Error(), poolNode.Name)
	poolNode.LbAlgorithm = nil
	poolNode.LbAlgorithmHash = nil
	poolNode.LbAlgorithmConsistentHashHdr = nil
	if !isFilterHMSet {
		poolNode.HealthMonitorRefs = nil
	}
}

func buildPoolWithBackendExtensionRefs(key string, poolNode *nodes.AviPoolNode, namespace string, backend *HTTPBackend) {
	healthMonitorRefsSet := sets.NewString()
	if backend == nil || backend.Filters == nil || len(backend.Filters) == 0 {
		return
	}
	for _, filter := range backend.Filters {
		if filter.ExtensionRef != nil {
			if filter.ExtensionRef.Kind == akogatewayapilib.HealthMonitorKind {
				obj, err := akogatewayapilib.GetDynamicInformers().HealthMonitorInformer.Lister().ByNamespace(namespace).Get(filter.ExtensionRef.Name)
				if err != nil {
					utils.AviLog.Warnf("key: %s, msg: error: HealthMonitor %s/%s will not be processed by gateway-container. err: %s", key, namespace, filter.ExtensionRef.Name, err)
					continue
				}
				unstructuredObj := obj.(*unstructured.Unstructured)
				status, found, err := unstructured.NestedMap(unstructuredObj.UnstructuredContent(), "status")
				if err != nil || !found {
					utils.AviLog.Warnf("key: %s, msg: error: HealthMonitor %s/%s status not found. err: %s", key, namespace, filter.ExtensionRef.Name, err)
					continue
				}
				uuid, ok := status["uuid"]
				if !ok {
					utils.AviLog.Warnf("key: %s, msg: error: HealthMonitor %s/%s uuid not found. err: %s", key, namespace, filter.ExtensionRef.Name, err)
					continue
				}
				healthMonitorRefsSet.Insert(uuid.(string))
			} else if filter.ExtensionRef.Kind == akogatewayapilib.RouteBackendExtensionKind {
				isFilterHMSet := healthMonitorRefsSet.Len() > 0
				err := akogatewayapilib.ParseRouteBackendExtensionCR(key, namespace, filter.ExtensionRef.Name, poolNode, isFilterHMSet)
				if err != nil {
					resetPoolNodeFields(key, err, poolNode, isFilterHMSet)
				}
			}
		}
	}
	if healthMonitorRefsSet.Len() > 0 {
		poolNode.HealthMonitorRefs = healthMonitorRefsSet.List()
	}
}

func (o *AviObjectGraph) BuildPGPool(key, parentNsName string, childVsNode *nodes.AviEvhVsNode, routeModel RouteModel, rule *Rule) {
	//reset pool, poolgroupreferences
	childVsNode.PoolGroupRefs = nil
	childVsNode.DefaultPoolGroup = ""
	childVsNode.PoolRefs = nil
	// create the PG from backends
	routeTypeNsName := lib.HTTPRoute + "/" + routeModel.GetNamespace() + "/" + routeModel.GetName()
	parentNs, _, parentName := lib.ExtractTypeNameNamespace(parentNsName)
	listeners := akogatewayapiobjects.GatewayApiLister().GetRouteToGatewayListener(routeTypeNsName, parentNsName)
	if len(listeners) == 0 {
		utils.AviLog.Warnf("key: %s, msg: No matching listener available for the route : %s", key, routeTypeNsName)
		return
	}
	//ListenerName/port/protocol/allowedRouteSpec
	listenerProtocol := listeners[0].Protocol
	var PGName string
	if rule.Name == "" {
		PGName = akogatewayapilib.GetPoolGroupName(parentNs, parentName,
			routeModel.GetNamespace(), routeModel.GetName(),
			utils.Stringify(rule.Matches))
	} else {
		PGName = akogatewayapilib.GetPoolGroupName(parentNs, parentName,
			routeModel.GetNamespace(), routeModel.GetName(),
			rule.Name)
	}
	PG := &nodes.AviPoolGroupNode{
		Name:   PGName,
		Tenant: childVsNode.Tenant,
	}
	PG.AviMarkers = utils.AviObjectMarkers{
		GatewayName:        parentName,
		GatewayNamespace:   parentNs,
		HTTPRouteName:      routeModel.GetName(),
		HTTPRouteNamespace: routeModel.GetNamespace(),
	}
	if rule.Name != "" {
		PG.AviMarkers.HTTPRouteRuleName = rule.Name
	}
	var persistenceProfile *nodes.AviApplicationPersistenceProfileNode
	if rule.SessionPersistence != nil {
		persistenceProfile = o.BuildApplicationPersistenceProfile(key, rule, routeModel, parentNs, parentName, PG.AviMarkers)
		if persistenceProfile != nil {
			persistenceProfile.Tenant = childVsNode.Tenant
		}
	}
	for _, httpbackend := range rule.Backends {
		var poolName string
		if rule.Name == "" {
			poolName = akogatewayapilib.GetPoolName(parentNs, parentName,
				routeModel.GetNamespace(), routeModel.GetName(),
				utils.Stringify(rule.Matches),
				httpbackend.Backend.Namespace, httpbackend.Backend.Name, strconv.Itoa(int(httpbackend.Backend.Port)))
		} else {
			poolName = akogatewayapilib.GetPoolName(parentNs, parentName,
				routeModel.GetNamespace(), routeModel.GetName(),
				rule.Name,
				httpbackend.Backend.Namespace, httpbackend.Backend.Name, strconv.Itoa(int(httpbackend.Backend.Port)))
		}
		svcObj, err := utils.GetInformers().ServiceInformer.Lister().Services(httpbackend.Backend.Namespace).Get(httpbackend.Backend.Name)
		if err != nil {
			utils.AviLog.Debugf("key: %s, msg: there was an error in retrieving the service", key)
			o.RemovePoolRefsFromPG(poolName, o.GetPoolGroupByName(PGName))
			continue
		}
		poolNode := &nodes.AviPoolNode{
			Name:       poolName,
			Tenant:     childVsNode.Tenant,
			Protocol:   listenerProtocol,
			PortName:   akogatewayapilib.FindPortName(httpbackend.Backend.Name, httpbackend.Backend.Namespace, httpbackend.Backend.Port, key),
			TargetPort: akogatewayapilib.FindTargetPort(httpbackend.Backend.Name, httpbackend.Backend.Namespace, httpbackend.Backend.Port, key),
			Port:       httpbackend.Backend.Port,
			ServiceMetadata: lib.ServiceMetadataObj{
				NamespaceServiceName: []string{httpbackend.Backend.Namespace + "/" + httpbackend.Backend.Name},
			},
			VrfContext: lib.GetVrf(),
		}
		poolNode.AviMarkers = utils.AviObjectMarkers{
			GatewayName:        parentName,
			GatewayNamespace:   parentNs,
			HTTPRouteName:      routeModel.GetName(),
			HTTPRouteNamespace: routeModel.GetNamespace(),
			BackendNs:          httpbackend.Backend.Namespace,
			BackendName:        httpbackend.Backend.Name,
		}
		if rule.SessionPersistence != nil {
			poolNode.ApplicationPersistenceProfile = persistenceProfile
		}
		if rule.Name != "" {
			poolNode.AviMarkers.HTTPRouteRuleName = rule.Name
		}

		if lib.IsIstioEnabled() {
			poolNode.UpdatePoolNodeForIstio()
		}

		t1LR := lib.GetT1LRPath()
		if found, infraSettingName := akogatewayapiobjects.GatewayApiLister().GetGatewayToAviInfraSetting(parentNsName); found {
			if infraSetting, err := akogatewayapilib.AKOControlConfig().AviInfraSettingInformer().Lister().Get(infraSettingName); err != nil {
				utils.AviLog.Warnf("key: %s, msg: failed to retrieve AviInfraSetting %s, err: %s", key, infraSettingName, err.Error())
			} else if infraSetting != nil && infraSetting.Status.Status == lib.StatusAccepted && infraSetting.Spec.NSXSettings.T1LR != nil {
				t1LR = *infraSetting.Spec.NSXSettings.T1LR
			}
		}

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
		buildPoolWithBackendExtensionRefs(key, poolNode, routeModel.GetNamespace(), httpbackend)
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

func (o *AviObjectGraph) BuildVHMatch(key string, parentNsName string, routeTypeNsName string, vsNode *nodes.AviEvhVsNode, rule *Rule, hosts []string) {
	var vhMatches []*models.VHMatch

	listeners := akogatewayapiobjects.GatewayApiLister().GetRouteToGatewayListener(routeTypeNsName, parentNsName)

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
				if match.PathMatch.Type == akogatewayapilib.EXACT {
					rule.Matches.Path.MatchCriteria = proto.String("EQUALS")
				} else if match.PathMatch.Type == akogatewayapilib.PATHPREFIX {
					rule.Matches.Path.MatchCriteria = proto.String("BEGINS_WITH")
				} else if match.PathMatch.Type == akogatewayapilib.REGULAREXPRESSION {
					rule.Matches.Path.MatchCriteria = proto.String("REGEX_MATCH")
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

func (o *AviObjectGraph) BuildHTTPPolicySet(key string, vsNode *nodes.AviEvhVsNode, routeModel RouteModel, rule *Rule, index int, httpPSName string) {

	if len(rule.Filters) == 0 {
		vsNode.HttpPolicyRefs = nil
		return
	}
	// go through filters
	var otherFiltersPresent bool

	for _, filter := range rule.Filters {
		if filter.RedirectFilter != nil || filter.UrlRewriteFilter != nil || filter.ResponseFilter != nil || filter.RequestFilter != nil {
			otherFiltersPresent = true
			break
		}
	}
	if !otherFiltersPresent {
		// do not require HTTPPolicy ref if above filters are not present
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
		policy = &nodes.AviHttpPolicySetNode{Name: httpPSName, Tenant: vsNode.Tenant}
		vsNode.HttpPolicyRefs = append(vsNode.HttpPolicyRefs, policy)
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
	o.BuildHTTPPolicySetHTTPRequestUrlRewriteRules(key, httpPSName, vsNode, routeModel, rule.Filters, index)
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
			vsNode.HttpPolicyRefs[index].RequestRules = []*models.HTTPRequestRule{requestRule}
			isRedirectPresent = true
			utils.AviLog.Debugf("key: %s, msg: Attached HTTP request redirect policies %s to vs %s", key, utils.Stringify(vsNode.HttpPolicyRefs[index].RequestRules), vsNode.Name)
			break
		}
	}
	return isRedirectPresent
}

func (o *AviObjectGraph) BuildHTTPPolicySetHTTPRequestUrlRewriteRules(key, httpPSname string, vsNode *nodes.AviEvhVsNode, routeModel RouteModel, filters []*Filter, index int) {
	urlRewriteAction := &models.HTTPRewriteURLAction{}
	for _, filter := range filters {
		// considering only the first reWriteFilter
		if filter.UrlRewriteFilter != nil {
			if filter.UrlRewriteFilter.hostname != "" {
				urlRewriteAction.HostHdr = &models.URIParam{
					Tokens: []*models.URIParamToken{
						{
							StrValue: &filter.UrlRewriteFilter.hostname,
							Type:     proto.String("URI_TOKEN_TYPE_STRING"),
						},
					},
					Type: proto.String("URI_PARAM_TYPE_TOKENIZED"),
				}
			}
			if filter.UrlRewriteFilter.path != nil {
				urlRewriteAction.Path = &models.URIParam{
					Tokens: []*models.URIParamToken{{
						StrValue: filter.UrlRewriteFilter.path.ReplaceFullPath,
						Type:     proto.String("URI_TOKEN_TYPE_STRING"),
					}},
					Type: proto.String("URI_PARAM_TYPE_TOKENIZED"),
				}
			}
			urlRewriteAction.Query = &models.URIParamQuery{
				AddString: nil,
				KeepQuery: proto.Bool(true),
			}

			if len(vsNode.HttpPolicyRefs[index].RequestRules) == 0 {
				requestRule := &models.HTTPRequestRule{Name: &httpPSname, Enable: proto.Bool(true), RewriteURLAction: urlRewriteAction, Index: proto.Int32(int32(index + 1))}
				vsNode.HttpPolicyRefs[index].RequestRules = []*models.HTTPRequestRule{requestRule}
			} else {
				vsNode.HttpPolicyRefs[index].RequestRules[0].RewriteURLAction = urlRewriteAction
			}
			utils.AviLog.Debugf("key: %s, msg: Attached HTTP request redirect policies %s to vs %s", key, utils.Stringify(vsNode.HttpPolicyRefs[index].RequestRules), vsNode.Name)
			break
		}
	}
}
