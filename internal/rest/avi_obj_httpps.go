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

package rest

import (
	"errors"
	"fmt"
	"sort"
	"strconv"

	avicache "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/cache"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/nodes"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	avimodels "github.com/vmware/alb-sdk/go/models"

	"github.com/davecgh/go-spew/spew"
	"google.golang.org/protobuf/proto"
	"k8s.io/apimachinery/pkg/util/sets"
)

func (rest *RestOperations) AviHttpPSBuild(hps_meta *nodes.AviHttpPolicySetNode, cache_obj *avicache.AviHTTPPolicyCache, key string) *utils.RestOp {

	if lib.CheckObjectNameLength(hps_meta.Name, lib.HTTPPS) {
		utils.AviLog.Warnf("key: %s not processing HTTPS object", key)
		return nil
	}
	name := hps_meta.Name
	var httpPresentPaths []string
	httpPresentIng := sets.NewString()

	tenant := fmt.Sprintf("/api/tenant/?name=%s", lib.GetEscapedValue(hps_meta.Tenant))
	cr := lib.AKOUser

	http_req_pol := avimodels.HTTPRequestPolicy{}
	http_sec_pol := avimodels.HttpsecurityPolicy{}
	hps := avimodels.HTTPPolicySet{Name: &name,
		CreatedBy: &cr, TenantRef: &tenant, HTTPRequestPolicy: &http_req_pol, HTTPSecurityPolicy: &http_sec_pol}

	var hppmapWithPath []nodes.AviHostPathPortPoolPG
	var hppmapWithoutPath []nodes.AviHostPathPortPoolPG
	var hppmapAllPaths []nodes.AviHostPathPortPoolPG
	for _, hppmap := range hps_meta.HppMap {
		if hppmap.Path != nil {
			hppmapWithPath = append(hppmapWithPath, hppmap)
			httpPresentPaths = append(httpPresentPaths, hppmap.Path[0])
		} else {
			hppmapWithoutPath = append(hppmapWithoutPath, hppmap)
		}
		httpPresentIng.Insert(hppmap.IngName)
	}
	sort.Slice(hppmapWithPath, func(i, j int) bool {
		return len(hppmapWithPath[i].Path[0]) > len(hppmapWithPath[j].Path[0])
	})
	hppmapAllPaths = append(hppmapAllPaths, hppmapWithPath...)
	hppmapAllPaths = append(hppmapAllPaths, hppmapWithoutPath...)

	if !hps_meta.AttachedToSharedVS {
		hps_meta.AviMarkers.Path = httpPresentPaths
		hps_meta.AviMarkers.IngressName = httpPresentIng.List()
		hps.Markers = lib.GetAllMarkers(hps_meta.AviMarkers)
	} else {
		hps.Markers = lib.GetMarkers()
	}

	hps_meta.CalculateCheckSum()
	cksum := hps_meta.CloudConfigCksum
	cksumString := strconv.Itoa(int(cksum))

	hps.CloudConfigCksum = &cksumString
	var idx int32
	idx = 0
	for _, sec_rule := range hps_meta.SecurityRules {
		name := fmt.Sprintf("%s-%d", hps_meta.Name, idx)
		if lib.CheckObjectNameLength(name, lib.HTTPSecurityRule) {
			utils.AviLog.Warnf("key: %s not adding rule to HTTPS object", key)
			continue
		}
		action := avimodels.HttpsecurityAction{
			Action: &sec_rule.Action,
		}
		portMatch := avimodels.PortMatch{
			MatchCriteria: &sec_rule.MatchCriteria,
			Ports:         []int64{sec_rule.Port},
		}
		match := avimodels.MatchTarget{
			VsPort: &portMatch,
		}
		var j int32
		j = idx
		rule := avimodels.HttpsecurityRule{
			Action: &action,
			Enable: &sec_rule.Enable,
			Index:  &j,
			Match:  &match,
			Name:   &name,
		}
		http_sec_pol.Rules = append(http_sec_pol.Rules, &rule)
		idx = idx + 1
	}
	for _, hppmap := range hppmapAllPaths {
		enable := true
		name := fmt.Sprintf("%s-%d", hps_meta.Name, idx)
		if lib.CheckObjectNameLength(name, lib.HTTPRequestRule) {
			utils.AviLog.Warnf("key: %s not adding request rule to HTTPS object", key)
			continue
		}
		match_target := avimodels.MatchTarget{}

		if hppmap.StringGroupRefs != nil && len(hppmap.StringGroupRefs) > 0 {
			match_crit := hppmap.MatchCriteria
			var match_case string
			if hppmap.MatchCase != "" {
				match_case = hppmap.MatchCase
			} else {
				match_case = "SENSITIVE"
			}
			path_match := avimodels.PathMatch{
				MatchCriteria:   &match_crit,
				MatchCase:       &match_case,
				StringGroupRefs: hppmap.StringGroupRefs,
			}
			match_target.Path = &path_match
		} else if hppmap.Path != nil && len(hppmap.Path) > 0 {
			match_crit := hppmap.MatchCriteria
			match_case := "SENSITIVE"
			path_match := avimodels.PathMatch{
				MatchCriteria: &match_crit,
				MatchCase:     &match_case,
				MatchStr:      hppmap.Path,
			}
			match_target.Path = &path_match
		}

		if hppmap.Port != 0 {
			match_crit := "IS_IN"
			vsport_match := avimodels.PortMatch{
				MatchCriteria: &match_crit,
				Ports:         []int64{int64(hppmap.Port)},
			}
			match_target.VsPort = &vsport_match
		}

		sw_action := avimodels.HttpswitchingAction{}
		if hppmap.Pool != "" {
			action := "HTTP_SWITCHING_SELECT_POOL"
			sw_action.Action = &action
			pool_ref := fmt.Sprintf("/api/pool/?name=%s", hppmap.Pool)
			sw_action.PoolRef = &pool_ref
		} else if hppmap.PoolGroup != "" {
			action := "HTTP_SWITCHING_SELECT_POOLGROUP"
			sw_action.Action = &action
			pg_ref := fmt.Sprintf("/api/poolgroup/?name=%s", hppmap.PoolGroup)
			sw_action.PoolGroupRef = &pg_ref
		}

		var j int32
		j = idx
		rule := avimodels.HTTPRequestRule{
			Index:           &j,
			Enable:          &enable,
			Name:            &name,
			Match:           &match_target,
			SwitchingAction: &sw_action,
		}
		http_req_pol.Rules = append(http_req_pol.Rules, &rule)
		idx = idx + 1

	}

	for _, hppmap := range hps_meta.RedirectPorts {
		enable := true
		name := fmt.Sprintf("%s-%d", hps_meta.Name, idx)
		if lib.CheckObjectNameLength(name, lib.HTTPRedirectRule) {
			utils.AviLog.Warnf("key: %s not adding rule to HTTPS object", key)
			continue
		}
		match_target := avimodels.MatchTarget{}
		if len(hppmap.Hosts) > 0 {
			match_crit := "HDR_EQUALS"
			host_hdr_match := avimodels.HostHdrMatch{MatchCriteria: &match_crit,
				Value: hppmap.Hosts}
			match_target.HostHdr = &host_hdr_match
			port_match_crit := "IS_IN"
			match_target.VsPort = &avimodels.PortMatch{MatchCriteria: &port_match_crit, Ports: []int64{int64(hppmap.VsPort)}}
		}
		if hppmap.Path != "" && hppmap.MatchCriteria != "" {
			match_case := "SENSITIVE"
			path_match := avimodels.PathMatch{
				MatchCriteria: &hppmap.MatchCriteria,
				MatchCase:     &match_case,
				MatchStr:      []string{hppmap.Path},
			}
			match_target.Path = &path_match
		}
		redirect_action := avimodels.HTTPRedirectAction{}
		protocol := "HTTPS"
		if hppmap.Protocol != "" {
			protocol = hppmap.Protocol
		}
		redirect_action.StatusCode = &hppmap.StatusCode
		redirect_action.Protocol = &protocol
		port := uint32(hppmap.RedirectPort)
		redirect_action.Port = &port
		if hppmap.RedirectPath != "" {
			uriParamToken := &avimodels.URIParamToken{
				StrValue: &hppmap.RedirectPath,
				Type:     proto.String("URI_TOKEN_TYPE_STRING"),
			}
			redirect_action.Path = &avimodels.URIParam{
				Tokens: []*avimodels.URIParamToken{uriParamToken},
				Type:   proto.String("URI_PARAM_TYPE_TOKENIZED"),
			}
		}
		var j int32
		j = idx
		rule := avimodels.HTTPRequestRule{Enable: &enable, Index: &j,
			Name: &name, Match: &match_target, RedirectAction: &redirect_action}
		http_req_pol.Rules = append(http_req_pol.Rules, &rule)
		idx = idx + 1

	}
	if hps_meta.HeaderReWrite != nil {
		name := fmt.Sprintf("%s-%d", hps_meta.Name, idx)
		if lib.CheckObjectNameLength(name, lib.HTTPRewriteRule) {
			utils.AviLog.Warnf("key: %s not adding rule to HTTPS object", key)
		} else {
			var hostHdrActionArr []*avimodels.HTTPHdrAction
			enable := true

			match_crit := "HDR_EQUALS"
			host_hdr_match := avimodels.HostHdrMatch{MatchCriteria: &match_crit,
				Value: []string{hps_meta.HeaderReWrite.SourceHost}}
			match_target := avimodels.MatchTarget{}
			match_target.HostHdr = &host_hdr_match
			replaceHeaderLiteral := "HTTP_REPLACE_HDR"
			host := "Host"
			headerVal := avimodels.HTTPHdrValue{Val: &hps_meta.HeaderReWrite.TargetHost}
			headerData := avimodels.HTTPHdrData{Name: &host, Value: &headerVal}
			rewriteHeader := avimodels.HTTPHdrAction{}
			rewriteHeader.Action = &replaceHeaderLiteral
			rewriteHeader.Hdr = &headerData
			hostHdrActionArr = append(hostHdrActionArr, &rewriteHeader)
			var j int32
			j = idx
			rule := avimodels.HTTPRequestRule{
				Index:     &j,
				Enable:    &enable,
				Name:      &name,
				Match:     &match_target,
				HdrAction: hostHdrActionArr,
			}
			http_req_pol.Rules = append(http_req_pol.Rules, &rule)
		}

	}

	if hps_meta.RequestRules != nil {
		hps.HTTPRequestPolicy = &avimodels.HTTPRequestPolicy{
			Rules: hps_meta.RequestRules,
		}
	}

	if hps_meta.ResponseRules != nil {
		hps.HTTPResponsePolicy = &avimodels.HTTPResponsePolicy{
			Rules: hps_meta.ResponseRules,
		}
	}

	var path string
	var rest_op utils.RestOp
	if cache_obj != nil {
		path = "/api/httppolicyset/" + cache_obj.Uuid
		rest_op = utils.RestOp{
			ObjName: hps_meta.Name,
			Path:    path,
			Method:  utils.RestPut,
			Obj:     hps,
			Tenant:  hps_meta.Tenant,
			Model:   "HTTPPolicySet",
		}

	} else {
		// Patch an existing http policy set object if it exists in the cache but not associated with this VS.
		httppol_key := avicache.NamespaceName{Namespace: hps_meta.Tenant, Name: hps_meta.Name}
		hps_cache, ok := rest.cache.HTTPPolicyCache.AviCacheGet(httppol_key)
		if ok {
			hps_cache_obj, _ := hps_cache.(*avicache.AviHTTPPolicyCache)
			path = "/api/httppolicyset/" + hps_cache_obj.Uuid
			rest_op = utils.RestOp{
				ObjName: hps_meta.Name,
				Path:    path,
				Method:  utils.RestPut,
				Obj:     hps,
				Tenant:  hps_meta.Tenant,
				Model:   "HTTPPolicySet",
			}
		} else {
			path = "/api/httppolicyset/"
			rest_op = utils.RestOp{
				ObjName: hps_meta.Name,
				Path:    path,
				Method:  utils.RestPost,
				Obj:     hps,
				Tenant:  hps_meta.Tenant,
				Model:   "HTTPPolicySet",
			}
		}
	}

	utils.AviLog.Debug(spew.Sprintf("HTTPPolicySet Restop %v AviHttpPolicySetMeta %v",
		rest_op, *hps_meta))
	return &rest_op
}

func (rest *RestOperations) AviHttpPolicyDel(uuid string, tenant string, key string) *utils.RestOp {
	path := "/api/httppolicyset/" + uuid
	rest_op := utils.RestOp{
		Path:   path,
		Method: "DELETE",
		Tenant: tenant,
		Model:  "HTTPPolicySet",
	}
	utils.AviLog.Debug(spew.Sprintf("HTTP Policy Set DELETE Restop %v ",
		utils.Stringify(rest_op)))
	return &rest_op
}

func (rest *RestOperations) AviHTTPPolicyCacheAdd(rest_op *utils.RestOp, vsKey avicache.NamespaceName, key string) error {
	if (rest_op.Err != nil) || (rest_op.Response == nil) {
		utils.AviLog.Warnf("key: %s, rest_op has err or no response for httppolicyset, err: %s, response: %s", key, rest_op.Err, rest_op.Response)
		return errors.New("errored rest_op")
	}

	resp_elems := rest.restOperator.RestRespArrToObjByType(rest_op, "httppolicyset", key)
	if resp_elems == nil {
		utils.AviLog.Warnf("Unable to find HTTP Policy Set obj in resp %v", rest_op.Response)
		return errors.New("HTTP Policy Set object not found")
	}

	for _, resp := range resp_elems {
		name, ok := resp["name"].(string)
		if !ok {
			utils.AviLog.Warnf("key: %s, Name not present in response %v", key, resp)
			continue
		}

		uuid, ok := resp["uuid"].(string)
		if !ok {
			utils.AviLog.Warnf("key: %s, Uuid not present in response %v", key, resp)
			continue
		}

		cksum := resp["cloud_config_cksum"].(string)

		var lastModifiedStr string
		lastModifiedIntf, ok := resp["_last_modified"]
		if !ok {
			utils.AviLog.Warnf("key: %s, msg: last_modified not present in response %v", key, resp)
		} else {
			lastModifiedStr, ok = lastModifiedIntf.(string)
			if !ok {
				utils.AviLog.Warnf("key: %s, msg: last_modified is not of type string", key)
			}
		}
		var pgMembers []string
		var poolMembers []string
		var stringGroupRefs []string
		if resp["http_request_policy"] != nil {
			if rules, rulesOk := resp["http_request_policy"].(map[string]interface{}); rulesOk {
				if rulesArr, rulesArrOk := rules["rules"].([]interface{}); rulesArrOk {
					for _, ruleIntf := range rulesArr {
						rulemap, _ := ruleIntf.(map[string]interface{})
						if rulemap["switching_action"] != nil {
							switchAction := rulemap["switching_action"].(map[string]interface{})
							if switchAction["pool_group_ref"] != nil {
								pgUuid := avicache.ExtractUUID(switchAction["pool_group_ref"].(string), "poolgroup-.*.#")
								// Search the poolName using this Uuid in the poolcache.
								pgName, found := rest.cache.PgCache.AviCacheGetNameByUuid(pgUuid)
								if found {
									pgMembers = append(pgMembers, pgName.(string))
								}
							} else if switchAction["pool_ref"] != nil {
								poolUuid := avicache.ExtractUUID(switchAction["pool_ref"].(string), "pool-.*.#")
								poolName, found := rest.cache.PoolCache.AviCacheGetNameByUuid(poolUuid)
								if found {
									poolMembers = append(poolMembers, poolName.(string))
								}
							}
						}
						if rulemap["match"] != nil {
							matchMap, _ := rulemap["match"].(map[string]interface{})
							if matchMap["path"] != nil {
								pathMap, _ := matchMap["path"].(map[string]interface{})
								if pathMap["string_group_refs"] != nil {
									sgRefs, _ := pathMap["string_group_refs"].([]interface{})
									for _, sg := range sgRefs {
										sgUuid := avicache.ExtractUUID(sg.(string), "stringgroup-.*.#")
										// Search the string group name using this Uuid in the string group cache.
										sgName, found := rest.cache.StringGroupCache.AviCacheGetNameByUuid(sgUuid)
										if found {
											stringGroupRefs = append(stringGroupRefs, sgName.(string))
										}
									}

								}
							}
						}
					}
				}
			}
		}
		http_cache_obj := avicache.AviHTTPPolicyCache{Name: name, Tenant: rest_op.Tenant,
			Uuid:             uuid,
			CloudConfigCksum: cksum,
			LastModified:     lastModifiedStr,
			PoolGroups:       pgMembers,
			Pools:            poolMembers,
			StringGroupRefs:  stringGroupRefs,
		}
		if lastModifiedStr == "" {
			http_cache_obj.InvalidData = true
		}

		k := avicache.NamespaceName{Namespace: rest_op.Tenant, Name: name}
		rest.cache.HTTPPolicyCache.AviCacheAdd(k, &http_cache_obj)
		vs_cache, ok := rest.cache.VsCacheMeta.AviCacheGet(vsKey)
		if ok {
			vs_cache_obj, found := vs_cache.(*avicache.AviVsCache)
			if found {
				vs_cache_obj.AddToHTTPKeyCollection(k)
				utils.AviLog.Debugf("Modified the VS cache for https object. The cache now is :%v", utils.Stringify(vs_cache_obj))
			}

		} else {
			vs_cache_obj := rest.cache.VsCacheMeta.AviCacheAddVS(vsKey)
			vs_cache_obj.AddToHTTPKeyCollection(k)
			utils.AviLog.Debug(spew.Sprintf("Added VS cache key during http policy update %v val %v", vsKey,
				vs_cache_obj))
		}
		utils.AviLog.Debug(spew.Sprintf("Added Http Policy Set cache k %v val %v", k,
			http_cache_obj))
	}

	return nil
}

func (rest *RestOperations) AviHTTPPolicyCacheDel(rest_op *utils.RestOp, vsKey avicache.NamespaceName, key string) error {
	httpkey := avicache.NamespaceName{Namespace: rest_op.Tenant, Name: rest_op.ObjName}
	rest.cache.HTTPPolicyCache.AviCacheDelete(httpkey)
	vs_cache, ok := rest.cache.VsCacheMeta.AviCacheGet(vsKey)
	if ok {
		vs_cache_obj, found := vs_cache.(*avicache.AviVsCache)
		if found {
			vs_cache_obj.RemoveFromHTTPKeyCollection(httpkey)
		}
	}

	return nil
}
