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
	"net/url"
	"strings"

	avicache "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/cache"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/nodes"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	avimodels "github.com/vmware/alb-sdk/go/models"

	"github.com/davecgh/go-spew/spew"
)

func (rest *RestOperations) AviL4PSBuild(hps_meta *nodes.AviL4PolicyNode, cache_obj *avicache.AviL4PolicyCache, key string) *utils.RestOp {

	if lib.CheckObjectNameLength(hps_meta.Name, lib.L4PS) {
		utils.AviLog.Warnf("key: %s not processing L4 policyset object", key)
		return nil
	}
	name := hps_meta.Name
	tenant := fmt.Sprintf("/api/tenant/?name=%s", url.QueryEscape(hps_meta.Tenant))
	cr := lib.AKOUser

	hps := avimodels.L4PolicySet{
		Name:      &name,
		CreatedBy: &cr,
		TenantRef: &tenant,
	}

	hps.Markers = lib.GetAllMarkers(hps_meta.AviMarkers)

	var idx int32
	idx = 0
	var l4Policy avimodels.L4ConnectionPolicy
	var l4rules []*avimodels.L4Rule
	for _, hppmap := range hps_meta.PortPool {
		if hppmap.Port != 0 {
			// Keep the l4 policy rule name similar to the Pool name it corresponds to.
			ruleName := hppmap.Pool
			if lib.CheckObjectNameLength(ruleName, lib.L4PSRule) {
				utils.AviLog.Warnf("key: %s not adding L4 PolicyRule to Policyset object", key)
				continue
			}
			var ports []int64
			l4rule := &avimodels.L4Rule{}

			l4rule.Name = &ruleName
			ports = append(ports, int64(hppmap.Port))
			l4action := &avimodels.L4RuleAction{}
			actionSelect := &avimodels.L4RuleActionSelectPool{}
			poolName := hppmap.Pool
			actionSelect.PoolRef = &poolName
			poolSelect := "L4_RULE_ACTION_SELECT_POOL"
			actionSelect.ActionType = &poolSelect
			l4action.SelectPool = actionSelect
			l4rule.Action = l4action
			j := idx
			l4rule.Index = &j
			portMatch := &avimodels.L4RulePortMatch{}
			portMatch.Ports = ports
			matchCriteria := "IS_IN"
			portMatch.MatchCriteria = &matchCriteria
			ruleMatchTarget := &avimodels.L4RuleMatchTarget{}

			l4Protocol := &avimodels.L4RuleProtocolMatch{}
			l4Protocol.MatchCriteria = &matchCriteria
			if hppmap.Protocol == utils.TCP {
				tcpString := "PROTOCOL_TCP"
				l4Protocol.Protocol = &tcpString
			} else if hppmap.Protocol == utils.UDP {
				udpString := "PROTOCOL_UDP"
				l4Protocol.Protocol = &udpString
			} else if hppmap.Protocol == utils.SCTP {
				sctpString := "PROTOCOL_SCTP"
				l4Protocol.Protocol = &sctpString
			}
			ruleMatchTarget.Port = portMatch
			ruleMatchTarget.Protocol = l4Protocol
			l4rule.Match = ruleMatchTarget
			l4rules = append(l4rules, l4rule)
			l4Policy.Rules = l4rules
			idx = idx + 1

		}
	}
	hps.L4ConnectionPolicy = &l4Policy
	var path string
	var rest_op utils.RestOp
	if cache_obj != nil {
		path = "/api/l4policyset/" + cache_obj.Uuid
		rest_op = utils.RestOp{
			ObjName: hps_meta.Name,
			Path:    path,
			Method:  utils.RestPut,
			Obj:     hps,
			Tenant:  hps_meta.Tenant,
			Model:   "L4PolicySet",
		}

	} else {
		// Patch an existing l4 policy set object if it exists in the cache but not associated with this VS.
		l4pol_key := avicache.NamespaceName{Namespace: hps_meta.Tenant, Name: hps_meta.Name}
		hps_cache, ok := rest.cache.L4PolicyCache.AviCacheGet(l4pol_key)
		if ok {
			hps_cache_obj, _ := hps_cache.(*avicache.AviL4PolicyCache)
			path = "/api/l4policyset/" + hps_cache_obj.Uuid
			rest_op = utils.RestOp{
				ObjName: hps_meta.Name,
				Path:    path,
				Method:  utils.RestPut,
				Obj:     hps,
				Tenant:  hps_meta.Tenant,
				Model:   "L4PolicySet",
			}
		} else {
			path = "/api/l4policyset/"
			rest_op = utils.RestOp{
				ObjName: hps_meta.Name,
				Path:    path,
				Method:  utils.RestPost,
				Obj:     hps,
				Tenant:  hps_meta.Tenant,
				Model:   "L4PolicySet",
			}
		}
	}

	utils.AviLog.Debug(spew.Sprintf("L4PolicySet Restop %v AviHttpPolicySetMeta %v",
		rest_op, utils.Stringify(hps_meta)))
	return &rest_op
}

func (rest *RestOperations) AviL4PolicyDel(uuid string, tenant string, key string) *utils.RestOp {
	path := "/api/l4policyset/" + uuid
	rest_op := utils.RestOp{
		Path:   path,
		Method: "DELETE",
		Tenant: tenant,
		Model:  "L4PolicySet",
	}
	utils.AviLog.Infof(spew.Sprintf("L4 Policy Set DELETE Restop %v ",
		utils.Stringify(rest_op)))
	return &rest_op
}

func (rest *RestOperations) AviL4PolicyCacheAdd(rest_op *utils.RestOp, vsKey avicache.NamespaceName, key string) error {
	if (rest_op.Err != nil) || (rest_op.Response == nil) {
		utils.AviLog.Warnf("key: %s, rest_op has err or no response for l4policyset, err: %s, response: %s", key, rest_op.Err, rest_op.Response)
		return errors.New("Errored rest_op")
	}

	resp_elems := rest.restOperator.RestRespArrToObjByType(rest_op, "l4policyset", key)
	if resp_elems == nil {
		utils.AviLog.Warnf("Unable to find L4 Policy Set obj in resp %v", rest_op.Response)
		return errors.New("L4 Policy Set object not found")
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

		var l4policyset avimodels.L4PolicySet
		var protocols []string
		var ports []int64
		var pools []string
		switch rest_op.Obj.(type) {
		case utils.AviRestObjMacro:
			l4policyset = rest_op.Obj.(utils.AviRestObjMacro).Data.(avimodels.L4PolicySet)
		case avimodels.L4PolicySet:
			l4policyset = rest_op.Obj.(avimodels.L4PolicySet)
		}
		for _, rule := range l4policyset.L4ConnectionPolicy.Rules {
			// cannot create an external load balancer with mix protocol - hence just caching the protocol once
			protocols = append(protocols, *rule.Match.Protocol.Protocol)
			ports = rule.Match.Port.Ports
			pool := strings.TrimPrefix(*rule.Action.SelectPool.PoolRef, "/api/pool?name=")
			pools = append(pools, pool)
		}
		emptyIngestionMarkers := utils.AviObjectMarkers{}
		//This is fetching data from response send at avi controller.
		cksum := lib.L4PolicyChecksum(ports, protocols, emptyIngestionMarkers, l4policyset.Markers, true)
		l4_cache_obj := avicache.AviL4PolicyCache{Name: name, Tenant: rest_op.Tenant,
			Uuid:             uuid,
			LastModified:     lastModifiedStr,
			Pools:            pools,
			CloudConfigCksum: cksum,
		}

		k := avicache.NamespaceName{Namespace: rest_op.Tenant, Name: name}
		rest.cache.L4PolicyCache.AviCacheAdd(k, &l4_cache_obj)
		vs_cache, ok := rest.cache.VsCacheMeta.AviCacheGet(vsKey)
		if ok {
			vs_cache_obj, found := vs_cache.(*avicache.AviVsCache)
			if found {
				vs_cache_obj.AddToL4PolicyCollection(k)
				utils.AviLog.Infof("Modified the VS cache for l4s object. The cache now is :%v", utils.Stringify(vs_cache_obj))
			}

		} else {
			vs_cache_obj := rest.cache.VsCacheMeta.AviCacheAddVS(vsKey)
			vs_cache_obj.AddToL4PolicyCollection(k)
			utils.AviLog.Infof(spew.Sprintf("Added VS cache key during l4 policy update %v val %v", vsKey,
				vs_cache_obj))
		}
		utils.AviLog.Infof(spew.Sprintf("Added L4 Policy Set cache k %v val %v", k,
			l4_cache_obj))
	}

	return nil
}

func (rest *RestOperations) AviL4PolicyCacheDel(rest_op *utils.RestOp, vsKey avicache.NamespaceName, key string) error {
	l4key := avicache.NamespaceName{Namespace: rest_op.Tenant, Name: rest_op.ObjName}
	rest.cache.L4PolicyCache.AviCacheDelete(l4key)
	vs_cache, ok := rest.cache.VsCacheMeta.AviCacheGet(vsKey)
	if ok {
		vs_cache_obj, found := vs_cache.(*avicache.AviVsCache)
		if found {
			vs_cache_obj.RemoveFromL4PolicyCollection(l4key)
		}
	}

	return nil
}
