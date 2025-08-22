package nodes

import (
	"fmt"

	"github.com/vmware/alb-sdk/go/models"

	"google.golang.org/protobuf/proto"

	akogatewayapilib "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/lib"
	akogatewayapiobjects "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/objects"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/nodes"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
)

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

	// First, remove existing pools and pool groups for this HTTPRoute to replace them
	o.RemovePoolsAndPoolGroupsForHTTPRoute(key, dedicatedVS, routeModel)
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

// RemovePoolsAndPoolGroupsForHTTPRoute removes existing pools and pool groups for a specific HTTPRoute
// This ensures that when an HTTPRoute is updated, old pools/pool groups are replaced instead of accumulated
func (o *AviObjectGraph) RemovePoolsAndPoolGroupsForHTTPRoute(key string, vsNode *nodes.AviEvhVsNode, routeModel RouteModel) {
	// Remove pool groups for this HTTPRoute
	var updatedPoolGroupRefs []*nodes.AviPoolGroupNode
	for _, poolGroup := range vsNode.PoolGroupRefs {
		if poolGroup.AviMarkers.HTTPRouteName == routeModel.GetName() &&
			poolGroup.AviMarkers.HTTPRouteNamespace == routeModel.GetNamespace() {
			utils.AviLog.Debugf("key: %s, msg: Removing existing Pool Group %s for HTTPRoute %s/%s replacement",
				key, poolGroup.Name, routeModel.GetNamespace(), routeModel.GetName())
		} else {
			updatedPoolGroupRefs = append(updatedPoolGroupRefs, poolGroup)
		}
	}
	vsNode.PoolGroupRefs = updatedPoolGroupRefs

	// Remove pools for this HTTPRoute
	var updatedPoolRefs []*nodes.AviPoolNode
	for _, pool := range vsNode.PoolRefs {
		if pool.AviMarkers.HTTPRouteName == routeModel.GetName() &&
			pool.AviMarkers.HTTPRouteNamespace == routeModel.GetNamespace() {
			utils.AviLog.Debugf("key: %s, msg: Removing existing Pool %s for HTTPRoute %s/%s replacement",
				key, pool.Name, routeModel.GetNamespace(), routeModel.GetName())
		} else {
			updatedPoolRefs = append(updatedPoolRefs, pool)
		}
	}
	vsNode.PoolRefs = updatedPoolRefs

	utils.AviLog.Debugf("key: %s, msg: Completed removal of existing pools/pool groups for HTTPRoute %s/%s",
		key, routeModel.GetNamespace(), routeModel.GetName())
}

// BuildSingleHTTPPolicySetForDedicatedMode creates a single HTTP PolicySet for all HTTPRoute rules in dedicated mode
// Each match gets its own request and response rules with specific match criteria
func (o *AviObjectGraph) BuildSingleHTTPPolicySetForDedicatedMode(key string, vsNode *nodes.AviEvhVsNode, routeModel RouteModel, httpRouteRules []*Rule) {
	if len(httpRouteRules) == 0 {
		return
	}

	// Create HTTP PolicySet name for the entire HTTPRoute with encoding
	httpPSName := akogatewayapilib.GetHttpPolicySetName(vsNode.AviMarkers.GatewayNamespace, vsNode.AviMarkers.GatewayName, routeModel.GetNamespace(), routeModel.GetName())

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
			o.BuildHTTPRequestRuleWithMatch(key, policy, routeModel, rule, ruleIndex, matchIndex, match, &requestRuleIndex, vsNode.AviMarkers.GatewayNamespace, vsNode.AviMarkers.GatewayName)
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
func (o *AviObjectGraph) BuildHTTPRequestRuleWithMatch(key string, policy *nodes.AviHttpPolicySetNode, routeModel RouteModel, rule *Rule, ruleIndex, matchIndex int, match *Match, requestRuleIndex *int32, gatewayNamespace, gatewayName string) {
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
		poolGroupName := o.GetPoolGroupNameForRule(routeModel, rule, ruleIndex, gatewayNamespace, gatewayName)
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
func (o *AviObjectGraph) GetPoolGroupNameForRule(routeModel RouteModel, rule *Rule, ruleIndex int, gatewayNamespace, gatewayName string) string {
	if rule.Name == "" {
		return akogatewayapilib.GetPoolGroupName(gatewayNamespace, gatewayName,
			routeModel.GetNamespace(), routeModel.GetName(),
			utils.Stringify(rule.Matches))
	} else {
		return akogatewayapilib.GetPoolGroupName(gatewayNamespace, gatewayName,
			routeModel.GetNamespace(), routeModel.GetName(),
			rule.Name)
	}
}

// BuildPGPoolForDedicatedMode creates pool groups and pools for dedicated mode
// In dedicated mode, pools are attached to HTTP PolicySets, not directly to VS
func (o *AviObjectGraph) BuildPGPoolForDedicatedMode(key string, vsNode *nodes.AviEvhVsNode, routeModel RouteModel, rule *Rule, ruleIndex int) {
	if len(rule.Backends) == 0 {
		utils.AviLog.Debugf("key: %s, msg: No backends found for rule %d in HTTPRoute %s/%s", key, ruleIndex, routeModel.GetNamespace(), routeModel.GetName())
		return
	}

	// Create pool group name using the same logic as GetPoolGroupNameForRule
	// This ensures the HTTP Policy Set references will match the actual pool group names
	poolGroupName := o.GetPoolGroupNameForRule(routeModel, rule, ruleIndex, vsNode.AviMarkers.GatewayNamespace, vsNode.AviMarkers.GatewayName)

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
		poolName := akogatewayapilib.GetDedicatedPoolName(poolGroupName, backend.Namespace, backend.Name, backend.Port, backendIndex)

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
