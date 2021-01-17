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
	"strconv"

	avicache "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/cache"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/nodes"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	avimodels "github.com/avinetworks/sdk/go/models"

	"github.com/davecgh/go-spew/spew"
)

func (rest *RestOperations) AviHttpPSBuild(hps_meta *nodes.AviHttpPolicySetNode, cache_obj *avicache.AviHTTPPolicyCache, key string) *utils.RestOp {
	name := hps_meta.Name
	cksum := hps_meta.CloudConfigCksum
	cksumString := strconv.Itoa(int(cksum))
	tenant := fmt.Sprintf("/api/tenant/?name=%s", hps_meta.Tenant)
	cr := lib.AKOUser

	http_req_pol := avimodels.HTTPRequestPolicy{}
	hps := avimodels.HTTPPolicySet{Name: &name, CloudConfigCksum: &cksumString,
		CreatedBy: &cr, TenantRef: &tenant, HTTPRequestPolicy: &http_req_pol}

	var idx int32
	idx = 0
	for _, hppmap := range hps_meta.HppMap {
		enable := true
		name := fmt.Sprintf("%s-%d", hps_meta.Name, idx)
		match_target := avimodels.MatchTarget{}
		if hppmap.Host != "" {
			var host []string
			host = append(host, hppmap.Host)
			match_crit := "HDR_EQUALS"
			host_hdr_match := avimodels.HostHdrMatch{
				MatchCriteria: &match_crit,
				Value:         host,
			}
			match_target.HostHdr = &host_hdr_match
		}

		if len(hppmap.Path) > 0 {
			match_crit := hppmap.MatchCriteria
			// always match case sensitive
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
		match_target := avimodels.MatchTarget{}
		if len(hppmap.Hosts) > 0 {
			match_crit := "HDR_EQUALS"
			host_hdr_match := avimodels.HostHdrMatch{MatchCriteria: &match_crit,
				Value: hppmap.Hosts}
			match_target.HostHdr = &host_hdr_match
			port_match_crit := "IS_IN"
			match_target.VsPort = &avimodels.PortMatch{MatchCriteria: &port_match_crit, Ports: []int64{int64(hppmap.VsPort)}}
		}
		redirect_action := avimodels.HTTPRedirectAction{}
		protocol := "HTTPS"
		redirect_action.StatusCode = &hppmap.StatusCode
		redirect_action.Protocol = &protocol
		redirect_action.Port = &hppmap.RedirectPort
		var j int32
		j = idx
		rule := avimodels.HTTPRequestRule{Enable: &enable, Index: &j,
			Name: &name, Match: &match_target, RedirectAction: &redirect_action}
		http_req_pol.Rules = append(http_req_pol.Rules, &rule)
		idx = idx + 1
	}

	macro := utils.AviRestObjMacro{ModelName: "HTTPPolicySet", Data: hps}
	var path string
	var rest_op utils.RestOp
	if cache_obj != nil {
		path = "/api/httppolicyset/" + cache_obj.Uuid
		rest_op = utils.RestOp{Path: path, Method: utils.RestPut, Obj: hps,
			Tenant: hps_meta.Tenant, Model: "HTTPPolicySet", Version: utils.CtrlVersion}

	} else {
		// Patch an existing http policy set object if it exists in the cache but not associated with this VS.
		httppol_key := avicache.NamespaceName{Namespace: hps_meta.Tenant, Name: hps_meta.Name}
		hps_cache, ok := rest.cache.HTTPPolicyCache.AviCacheGet(httppol_key)
		if ok {
			hps_cache_obj, _ := hps_cache.(*avicache.AviHTTPPolicyCache)
			path = "/api/httppolicyset/" + hps_cache_obj.Uuid
			rest_op = utils.RestOp{Path: path, Method: utils.RestPut, Obj: hps,
				Tenant: hps_meta.Tenant, Model: "HTTPPolicySet", Version: utils.CtrlVersion}
		} else {
			path = "/api/macro"
			rest_op = utils.RestOp{Path: path, Method: utils.RestPost, Obj: macro,
				Tenant: hps_meta.Tenant, Model: "HTTPPolicySet", Version: utils.CtrlVersion}
		}
	}

	utils.AviLog.Debug(spew.Sprintf("HTTPPolicySet Restop %v AviHttpPolicySetMeta %v\n",
		rest_op, *hps_meta))
	return &rest_op
}

func (rest *RestOperations) AviHttpPolicyDel(uuid string, tenant string, key string) *utils.RestOp {
	path := "/api/httppolicyset/" + uuid
	rest_op := utils.RestOp{Path: path, Method: "DELETE",
		Tenant: tenant, Model: "HTTPPolicySet", Version: utils.CtrlVersion}
	utils.AviLog.Debug(spew.Sprintf("HTTP Policy Set DELETE Restop %v \n",
		utils.Stringify(rest_op)))
	return &rest_op
}

func (rest *RestOperations) AviHTTPPolicyCacheAdd(rest_op *utils.RestOp, vsKey avicache.NamespaceName, key string) error {
	if (rest_op.Err != nil) || (rest_op.Response == nil) {
		utils.AviLog.Warnf("key: %s, rest_op has err or no response for httppolicyset, err: %s, response: %s", key, rest_op.Err, rest_op.Response)
		return errors.New("Errored rest_op")
	}

	resp_elems, ok := RestRespArrToObjByType(rest_op, "httppolicyset", key)
	if ok != nil || resp_elems == nil {
		utils.AviLog.Warnf("Unable to find HTTP Policy Set obj in resp %v", rest_op.Response)
		return errors.New("HTTP Policy Set object not found")
	}

	for _, resp := range resp_elems {
		name, ok := resp["name"].(string)
		if !ok {
			utils.AviLog.Warnf("Name not present in response %v", resp)
			continue
		}

		uuid, ok := resp["uuid"].(string)
		if !ok {
			utils.AviLog.Warnf("Uuid not present in response %v", resp)
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
		if resp["http_request_policy"] != nil {
			rules, rulessOk := resp["http_request_policy"].(map[string]interface{})
			if rulessOk {
				rulesArr := rules["rules"].([]interface{})
				for _, ruleIntf := range rulesArr {
					rulemap, _ := ruleIntf.(map[string]interface{})
					if rulemap["switching_action"] != nil {
						switchAction := rulemap["switching_action"].(map[string]interface{})
						pgUuid := avicache.ExtractUuid(switchAction["pool_group_ref"].(string), "poolgroup-.*.#")
						// Search the poolName using this Uuid in the poolcache.
						pgName, found := rest.cache.PgCache.AviCacheGetNameByUuid(pgUuid)
						if found {
							pgMembers = append(pgMembers, pgName.(string))
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
			utils.AviLog.Debug(spew.Sprintf("Added VS cache key during http policy update %v val %v\n", vsKey,
				vs_cache_obj))
		}
		utils.AviLog.Debug(spew.Sprintf("Added Http Policy Set cache k %v val %v\n", k,
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
