package nodes

import (
	"fmt"
	"sort"

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

	ruleToPoolGroupIndex := make(map[*Rule]int)
	poolGroupIndex := 0
	for _, rule := range httpRouteRules {
		if _, exists := ruleToPoolGroupIndex[rule]; !exists {
			ruleToPoolGroupIndex[rule] = poolGroupIndex
			poolGroupIndex++
		}
	}

	dedicatedVS.StringGroupRefs = nil
	// Create a single HTTP PolicySet for all HTTPRoute rules
	o.BuildHTTPPolicySetsForDedicatedMode(key, dedicatedVS, routeModel, httpRouteRules, ruleToPoolGroupIndex)

	// cleanup before processing
	dedicatedVS.PoolGroupRefs = nil
	dedicatedVS.PoolRefs = nil
	dedicatedVS.DefaultPoolGroup = ""

	// Create pool groups for unique rules using the same mapping
	for rule, poolGroupIndex := range ruleToPoolGroupIndex {
		// Create pools for the backends in this rule
		o.BuildPGPoolForDedicatedMode(key, dedicatedVS, routeModel, rule, poolGroupIndex)
		utils.AviLog.Debugf("key: %s, msg: Processed rule %d for HTTPRoute %s/%s in dedicated mode", key, poolGroupIndex, routeModel.GetNamespace(), routeModel.GetName())
	}
	// Update route to VS mapping for dedicated mode
	akogatewayapiobjects.GatewayApiLister().UpdateRouteChildVSMappings(routeModel.GetType()+"/"+routeModel.GetNamespace()+"/"+routeModel.GetName(), dedicatedVSName)
	utils.AviLog.Infof("key: %s, msg: Completed processing HTTPRoute %s/%s in dedicated mode for VS %s", key, routeModel.GetNamespace(), routeModel.GetName(), dedicatedVSName)
}

// BuildHTTPPolicySetsForDedicatedMode creates HTTP PolicySets for all HTTPRoute rules in dedicated mode
func (o *AviObjectGraph) BuildHTTPPolicySetsForDedicatedMode(key string, vsNode *nodes.AviEvhVsNode, routeModel RouteModel, httpRouteRules []*Rule, ruleToPoolGroupIndex map[*Rule]int) {
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

	// Set AviMarkers for the policy to ensure proper checksum calculation
	policy.AviMarkers = utils.AviObjectMarkers{
		GatewayName:        vsNode.AviMarkers.GatewayName,
		GatewayNamespace:   vsNode.AviMarkers.GatewayNamespace,
		HTTPRouteName:      routeModel.GetName(),
		HTTPRouteNamespace: routeModel.GetNamespace(),
	}

	// Clear existing rules to rebuild them completely
	policy.RequestRules = []*models.HTTPRequestRule{}
	policy.ResponseRules = []*models.HTTPResponseRule{}

	var requestRuleIndex int32 = 0
	var responseRuleIndex int32 = 1000 // Response rules start at higher indices

	type MatchWithMetadata struct {
		Match          *Match
		Rule           *Rule
		RuleIndex      int
		MatchIndex     int
		PoolGroupIndex int
	}

	rootPathPresent := false
	var allMatches []MatchWithMetadata
	for ruleIndex, rule := range httpRouteRules {
		for matchIndex, match := range rule.Matches {
			if match.PathMatch.Path == "/" {
				rootPathPresent = true
			}
			allMatches = append(allMatches, MatchWithMetadata{
				Match:          match,
				Rule:           rule,
				RuleIndex:      ruleIndex,
				MatchIndex:     matchIndex,
				PoolGroupIndex: ruleToPoolGroupIndex[rule],
			})
		}
	}

	sort.SliceStable(allMatches, func(i, j int) bool {
		pathI := allMatches[i].Match.PathMatch.Path
		pathJ := allMatches[j].Match.PathMatch.Path

		// path type priority
		priorityI := getPathTypePriority(allMatches[i].Match.PathMatch.Type)
		priorityJ := getPathTypePriority(allMatches[j].Match.PathMatch.Type)

		if priorityI != priorityJ {
			return priorityI > priorityJ
		}

		// if path type is equal, check length
		if len(pathI) != len(pathJ) {
			return len(pathI) > len(pathJ)
		}

		// maintain rule index
		if allMatches[i].RuleIndex != allMatches[j].RuleIndex {
			return allMatches[i].RuleIndex < allMatches[j].RuleIndex
		}
		// maintain match index
		return allMatches[i].MatchIndex < allMatches[j].MatchIndex
	})

	utils.AviLog.Debugf("key: %s, msg: Sorted %d matches for longest prefix matching in dedicated mode %v", key, len(allMatches), allMatches)

	for _, matchMeta := range allMatches {
		o.BuildHTTPRequestRuleWithMatch(key, policy, routeModel, matchMeta.Rule,
			matchMeta.Match, matchMeta.PoolGroupIndex, &requestRuleIndex, vsNode.AviMarkers.GatewayNamespace, vsNode.AviMarkers.GatewayName, vsNode)
	}

	for _, matchMeta := range allMatches {
		responseFilters := o.GetResponseFilters(matchMeta.Rule.Filters)
		if len(responseFilters) > 0 {
			o.BuildHTTPResponseRuleWithMatch(key, policy, routeModel, responseFilters,
				matchMeta.Match, &responseRuleIndex, vsNode)
		}
	}

	if !rootPathPresent {
		o.AddDefaultRule(key, policy, &requestRuleIndex)
	}
	utils.AviLog.Debugf("key: %s, msg: Built HTTP PolicySets for dedicated mode with %d request rules and %d response rules",
		key, httpPSName, len(policy.RequestRules), len(policy.ResponseRules))
}

func (o *AviObjectGraph) AddDefaultRule(key string, policy *nodes.AviHttpPolicySetNode, requestRuleIndex *int32) {
	// Check if default rule already exists
	for index, rule := range policy.RequestRules {
		if *rule.Name == "default-backend-rule" {
			// Update the index to ensure it's the last rule
			rule.Index = proto.Int32(*requestRuleIndex)
			// Move to end if not already there
			if index != len(policy.RequestRules)-1 {
				policy.RequestRules = append(policy.RequestRules[:index], policy.RequestRules[index+1:]...)
				policy.RequestRules = append(policy.RequestRules, rule)
			}
			*requestRuleIndex++
			return
		}
	}

	// Create new default rule with the current index
	defaultRule := &models.HTTPRequestRule{
		Name:   proto.String("default-backend-rule"),
		Enable: proto.Bool(true),
		Index:  proto.Int32(*requestRuleIndex),
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
	}
	policy.RequestRules = append(policy.RequestRules, defaultRule)
	*requestRuleIndex++
	utils.AviLog.Debugf("key: %s, msg: Added default rule to HTTP PolicySet %s", key, policy.Name)
}

// getPathTypePriority returns priority value for path match types for LPM sorting
func getPathTypePriority(pathType string) int {
	switch pathType {
	case akogatewayapilib.EXACT:
		return 3
	case akogatewayapilib.PATHPREFIX:
		return 2
	case akogatewayapilib.REGULAREXPRESSION:
		return 1
	default:
		return 2
	}
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
func (o *AviObjectGraph) BuildHTTPRequestRuleWithMatch(key string, policy *nodes.AviHttpPolicySetNode, routeModel RouteModel, rule *Rule, match *Match, poolGroupIndex int, requestRuleIndex *int32, gatewayNamespace, gatewayName string, vsNode *nodes.AviEvhVsNode) {
	ruleName := fmt.Sprintf("rule-%d", *requestRuleIndex)

	// Create the HTTP request rule
	httpRequestRule := &models.HTTPRequestRule{
		Name:   &ruleName,
		Enable: proto.Bool(true),
		Index:  proto.Int32(*requestRuleIndex),
		Match:  &models.MatchTarget{},
	}

	// Build match criteria
	o.BuildMatchTarget(httpRequestRule.Match, match, vsNode.Tenant, vsNode)

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
			var index uint32 = 0
			for requestFilterAdd := range filter.RequestFilter.Add {
				action := o.BuildHTTPPolicySetHTTPRuleHdrAction(key, "HTTP_ADD_HDR", filter.RequestFilter.Add[requestFilterAdd], index)
				httpRequestRule.HdrAction = append(httpRequestRule.HdrAction, action)
				index++
			}
			for requestFilterSet := range filter.RequestFilter.Set {
				action := o.BuildHTTPPolicySetHTTPRuleHdrAction(key, "HTTP_REPLACE_HDR", filter.RequestFilter.Set[requestFilterSet], index)
				httpRequestRule.HdrAction = append(httpRequestRule.HdrAction, action)
				index++
			}
			for requestFilterRemove := range filter.RequestFilter.Remove {
				action := o.BuildHTTPPolicySetHTTPRuleHdrAction(key, "HTTP_REMOVE_HDR", &Header{Name: filter.RequestFilter.Remove[requestFilterRemove]}, index)
				httpRequestRule.HdrAction = append(httpRequestRule.HdrAction, action)
				index++
			}
		}
		// Handle URL rewrite
		if filter.UrlRewriteFilter != nil {
			rewriteAction := &models.HTTPRewriteURLAction{}
			if filter.UrlRewriteFilter.path != nil {
				var pathValue string
				if filter.UrlRewriteFilter.path.ReplaceFullPath != nil {
					pathValue = *filter.UrlRewriteFilter.path.ReplaceFullPath
				}
				if pathValue != "" {
					rewriteAction.Path = &models.URIParam{
						Type:   proto.String("URI_PARAM_TYPE_TOKENIZED"),
						Tokens: []*models.URIParamToken{{Type: proto.String("URI_TOKEN_TYPE_STRING"), StrValue: &pathValue}},
					}
				}
			}
			if filter.UrlRewriteFilter.hostname != "" {
				rewriteAction.HostHdr = &models.URIParam{
					Type:   proto.String("URI_PARAM_TYPE_TOKENIZED"),
					Tokens: []*models.URIParamToken{{Type: proto.String("URI_TOKEN_TYPE_STRING"), StrValue: &filter.UrlRewriteFilter.hostname}},
				}
			}
			rewriteAction.Query = &models.URIParamQuery{
				AddString: nil,
				KeepQuery: proto.Bool(true),
			}
			httpRequestRule.RewriteURLAction = rewriteAction
		}
	}

	// Add switching action to backends
	if len(rule.Backends) > 0 {
		poolGroupName := o.GetPoolGroupNameForRule(routeModel, rule, poolGroupIndex, gatewayNamespace, gatewayName)
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
func (o *AviObjectGraph) BuildHTTPResponseRuleWithMatch(key string, policy *nodes.AviHttpPolicySetNode, routeModel RouteModel, responseFilters []*Filter, match *Match, responseRuleIndex *int32, vsNode *nodes.AviEvhVsNode) {
	ruleName := fmt.Sprintf("response-rule-%d", *responseRuleIndex)

	// Create the HTTP response rule
	httpResponseRule := &models.HTTPResponseRule{
		Name:   &ruleName,
		Enable: proto.Bool(true),
		Index:  proto.Int32(*responseRuleIndex),
		Match:  &models.ResponseMatchTarget{},
	}

	// Build match criteria for response rule (converting request match to response match)
	o.BuildResponseMatchTarget(httpResponseRule.Match, match, vsNode.Tenant, vsNode)

	// Apply response filters
	for _, filter := range responseFilters {
		if filter.ResponseFilter != nil {
			var index uint32 = 0
			// Handle response header additions
			for responseFilterAdd := range filter.ResponseFilter.Add {
				action := o.BuildHTTPPolicySetHTTPRuleHdrAction(key, "HTTP_ADD_HDR", filter.ResponseFilter.Add[responseFilterAdd], index)
				httpResponseRule.HdrAction = append(httpResponseRule.HdrAction, action)
				index++
			}
			// Handle response header replacements
			for responseFilterSet := range filter.ResponseFilter.Set {
				action := o.BuildHTTPPolicySetHTTPRuleHdrAction(key, "HTTP_REPLACE_HDR", filter.ResponseFilter.Set[responseFilterSet], index)
				httpResponseRule.HdrAction = append(httpResponseRule.HdrAction, action)
				index++
			}
			// Handle response header removals
			for responseFilterRemove := range filter.ResponseFilter.Remove {
				action := o.BuildHTTPPolicySetHTTPRuleHdrAction(key, "HTTP_REMOVE_HDR", &Header{Name: filter.ResponseFilter.Remove[responseFilterRemove]}, index)
				httpResponseRule.HdrAction = append(httpResponseRule.HdrAction, action)
				index++
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
func (o *AviObjectGraph) BuildMatchTarget(matchTarget *models.MatchTarget, match *Match, tenant string, vsNode *nodes.AviEvhVsNode) {
	// Handle path matching
	matchTarget.Path = &models.PathMatch{
		MatchCase: proto.String("SENSITIVE"),
		MatchStr:  []string{match.PathMatch.Path},
	}
	if match.PathMatch.Type == akogatewayapilib.EXACT {
		matchTarget.Path.MatchCriteria = proto.String("EQUALS")
	} else if match.PathMatch.Type == akogatewayapilib.PATHPREFIX {
		matchTarget.Path.MatchCriteria = proto.String("BEGINS_WITH")
	} else if match.PathMatch.Type == akogatewayapilib.REGULAREXPRESSION {
		matchTarget.Path.MatchCriteria = proto.String("REGEX_MATCH")
		// unset MatchStr
		matchTarget.Path.MatchStr = []string{}
		// generate string group to be attached
		regexStringGroupName := lib.GetEncodedStringGroupName("", match.PathMatch.Path)
		matchTarget.Path.StringGroupRefs = []string{"/api/stringgroup?name=" + regexStringGroupName}
		o.addStringGroup(regexStringGroupName, match.PathMatch.Path, tenant, vsNode)
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
func (o *AviObjectGraph) BuildResponseMatchTarget(responseMatchTarget *models.ResponseMatchTarget, match *Match, tenant string, vsNode *nodes.AviEvhVsNode) {
	// Path matching (applicable to response)

	responseMatchTarget.Path = &models.PathMatch{
		MatchCase: proto.String("SENSITIVE"),
		MatchStr:  []string{match.PathMatch.Path},
	}
	if match.PathMatch.Type == akogatewayapilib.EXACT {
		responseMatchTarget.Path.MatchCriteria = proto.String("EQUALS")
	} else if match.PathMatch.Type == akogatewayapilib.PATHPREFIX {
		responseMatchTarget.Path.MatchCriteria = proto.String("BEGINS_WITH")
	} else if match.PathMatch.Type == akogatewayapilib.REGULAREXPRESSION {
		responseMatchTarget.Path.MatchCriteria = proto.String("REGEX_MATCH")
		// unset MatchStr
		responseMatchTarget.Path.MatchStr = []string{}
		regexStringGroupName := lib.GetEncodedStringGroupName("", match.PathMatch.Path)
		responseMatchTarget.Path.StringGroupRefs = []string{"/api/stringgroup?name=" + regexStringGroupName}
		o.addStringGroup(regexStringGroupName, match.PathMatch.Path, tenant, vsNode)
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

// addStringGroup checks if a string group already exists in vsNode and adds it only if it doesn't exist
func (o *AviObjectGraph) addStringGroup(stringGroupName, pathPattern, tenant string, vsNode *nodes.AviEvhVsNode) {
	for _, existingStringGroup := range vsNode.StringGroupRefs {
		if existingStringGroup.StringGroup.Name != nil && *existingStringGroup.StringGroup.Name == stringGroupName {
			return
		}
	}

	kv := &models.KeyValue{
		Key: &pathPattern,
	}
	regexStringGroup := &models.StringGroup{
		TenantRef:    &tenant,
		Type:         proto.String("SG_TYPE_STRING"),
		LongestMatch: proto.Bool(true),
		Name:         &stringGroupName,
		Kv:           []*models.KeyValue{kv},
	}
	stringGroupNode := &nodes.AviStringGroupNode{
		StringGroup: regexStringGroup,
	}
	stringGroupNode.CloudConfigCksum = stringGroupNode.GetCheckSum()
	vsNode.StringGroupRefs = append(vsNode.StringGroupRefs, stringGroupNode)
	utils.AviLog.Debugf("key: %s, msg: Added string group %s to VS %s", stringGroupName, vsNode.Name)
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

		buildPoolWithBackendExtensionRefs(key, poolNode, routeModel.GetNamespace(), httpbackend)
		if vsNode.CheckPoolNChecksum(poolNode.Name, poolNode.GetCheckSum()) {
			// Replace the poolNode.
			vsNode.ReplaceEvhPoolInEVHNode(poolNode, key)
		}
		// Add pool reference to pool group
		poolRef := fmt.Sprintf("/api/pool?name=%s", poolName)
		poolGroupNode.Members = append(poolGroupNode.Members, &models.PoolGroupMember{
			PoolRef: &poolRef,
			Ratio:   &ratio,
		})

		if vsNode.CheckPoolNChecksum(poolNode.Name, poolNode.GetCheckSum()) {
			// Replace the poolNode.
			vsNode.ReplaceEvhPoolInEVHNode(poolNode, key)
		}

		utils.AviLog.Debugf("key: %s, msg: Created pool %s with ratio %d for dedicated mode rule %d",
			key, poolName, ratio, ruleIndex)
	}
	// pool groups are attached to VS for management
	if !utils.HasElem(vsNode.PoolGroupRefs, poolGroupName) {
		vsNode.PoolGroupRefs = append(vsNode.PoolGroupRefs, poolGroupNode)
	}

	utils.AviLog.Debugf("key: %s, msg: Built pool group %s for dedicated mode rule %d (attached via HTTP PolicySet)",
		key, poolGroupName, ruleIndex)
}
