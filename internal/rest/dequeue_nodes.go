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
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/vmware/alb-sdk/go/clients"
	avimodels "github.com/vmware/alb-sdk/go/models"
	"github.com/vmware/alb-sdk/go/session"

	avicache "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/cache"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/nodes"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/status"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/api/models"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
)

type RestOperations struct {
	cache        *avicache.AviObjCache
	restOperator RestOperator
}

func NewRestOperations(cache *avicache.AviObjCache, overrideLeaderFlag ...bool) RestOperations {
	restOp := RestOperations{}
	restOp.cache = cache
	restOp.restOperator = NewRestOperator(&restOp, overrideLeaderFlag...)
	return restOp
}

func (rest *RestOperations) CleanupVS(key string, skipVS bool) {
	namespace, name := utils.ExtractNamespaceObjectName(key)
	vsKey := avicache.NamespaceName{Namespace: namespace, Name: name}
	vs_cache_obj := rest.getVsCacheObj(vsKey, key)
	utils.AviLog.Infof("key: %s, msg: cleanup mode, removing all stale objects", key)
	rest.DeleteVSOper(vsKey, vs_cache_obj, namespace, key, skipVS, false)
	utils.AviLog.Infof("key: %s, msg: cleanup mode, stale object removal done", key)
}

func (rest *RestOperations) DequeueNodes(key string) {
	utils.AviLog.Infof("key: %s, msg: start rest layer sync.", key)
	lib.DecrementQueueCounter(utils.GraphLayer)
	// Got the key from the Graph Layer - let's fetch the model
	ok, avimodelIntf := objects.SharedAviGraphLister().Get(key)
	if !ok {
		utils.AviLog.Warnf("key: %s, msg: no model found for the key", key)
	}
	if key == lib.IstioModel {
		avimodel := avimodelIntf.(*nodes.AviObjectGraph)
		utils.AviLog.Infof("key: %s, msg: processing istio object", key)
		rest.IstioCU(key, avimodel)
		return
	}
	namespace, name := utils.ExtractNamespaceObjectName(key)
	vsKey := avicache.NamespaceName{Namespace: namespace, Name: name}
	vs_cache_obj := rest.getVsCacheObj(vsKey, key)
	if !ok || avimodelIntf == nil {
		if avimodelIntf != nil {
			avimodel, ok := avimodelIntf.(*nodes.AviObjectGraph)
			if ok && key == lib.IstioModel {
				utils.AviLog.Infof("key: %s, msg: processing istio object", key)
				rest.IstioCU(key, avimodel)
				return
			}
		}
		if lib.StaticRouteSyncChan != nil {
			close(lib.StaticRouteSyncChan)
			lib.StaticRouteSyncChan = nil
		}
		if vs_cache_obj != nil {
			utils.AviLog.Infof("key: %s, msg: nil model found, this is a vs deletion case", key)
			rest.DeleteVSOper(vsKey, vs_cache_obj, namespace, key, false, false)
		}
	} else if ok && avimodelIntf != nil {
		avimodel := avimodelIntf.(*nodes.AviObjectGraph)
		if avimodel == nil {
			utils.AviLog.Debugf("Empty Model found, skipping")
			return
		}
		avimodel, ok = avimodel.GetCopy(key)
		if !ok {
			utils.AviLog.Warnf("key: %s, failed to get process model", key)
			return
		}
		if avimodel.IsVrf {
			utils.AviLog.Infof("key: %s, msg: processing vrf object", key)
			rest.vrfCU(key, name, avimodel)
			return
		}
		if strings.Contains(name, "StringGroup") {
			utils.AviLog.Infof("key: %s, msg: processing stringgroup object", key)
			rest.stringGroupCU(key, name, avimodel)
			return
		}
		utils.AviLog.Debugf("key: %s, msg: VS create/update.", key)
		if strings.Contains(name, "-EVH") && lib.IsEvhEnabled() {
			if len(avimodel.GetAviEvhVS()) != 1 {
				utils.AviLog.Warnf("key: %s, msg: virtualservice in the model is not equal to 1:%v", key, avimodel.GetAviEvhVS())
				return
			}
			rest.RestOperationForEvh(name, namespace, avimodel, false, vs_cache_obj, key)

		} else {
			if len(avimodel.GetAviVS()) != 1 {
				utils.AviLog.Warnf("key: %s, msg: virtualservice in the model is not equal to 1:%v", key, avimodel.GetAviVS())
				return
			}
			rest.RestOperation(name, namespace, avimodel, vs_cache_obj, key)
		}

	}

}
func (rest *RestOperations) IstioCU(key string, avimodel *nodes.AviObjectGraph) (bool, bool) {
	var restOps []*utils.RestOp
	var pkiSuccess, sslSuccess bool

	pkiNode, sslNode := avimodel.GetIstioNodes()
	pkiKey := avicache.NamespaceName{Namespace: lib.GetTenant(), Name: pkiNode.Name}
	sslKey := avicache.NamespaceName{Namespace: lib.GetTenant(), Name: sslNode.Name}

	pkiCacheObj, ok := rest.cache.PKIProfileCache.AviCacheGet(pkiKey)
	if !ok {
		restOp := rest.AviPkiProfileBuild(pkiNode, nil)
		restOps = []*utils.RestOp{restOp}
		pkiSuccess, _ = rest.ExecuteRestAndPopulateCache(restOps, pkiKey, avimodel, key, false)
	} else {
		pkiCache := pkiCacheObj.(*avicache.AviPkiProfileCache)
		if pkiCache.CloudConfigCksum != pkiNode.GetCheckSum() {
			restOp := rest.AviPkiProfileBuild(pkiNode, pkiCache)
			restOps = []*utils.RestOp{restOp}
			pkiSuccess, _ = rest.ExecuteRestAndPopulateCache(restOps, pkiKey, avimodel, key, false)
		}
	}

	sslCacheObj, ok := rest.cache.SSLKeyCache.AviCacheGet(sslKey)
	if !ok {
		restOp := rest.AviSSLBuild(sslNode, nil)
		restOps = []*utils.RestOp{restOp}
		sslSuccess, _ = rest.ExecuteRestAndPopulateCache(restOps, sslKey, avimodel, key, false)
	} else {
		sslCache := sslCacheObj.(*avicache.AviSSLCache)
		if sslCache.CloudConfigCksum != sslNode.GetCheckSum() {
			restOp := rest.AviSSLBuild(sslNode, sslCache)
			restOps = []*utils.RestOp{restOp}
			sslSuccess, _ = rest.ExecuteRestAndPopulateCache(restOps, sslKey, avimodel, key, false)
		}
	}
	return pkiSuccess, sslSuccess
}
func (rest *RestOperations) vrfCU(key, vrfName string, avimodel *nodes.AviObjectGraph) {
	if lib.GetDisableStaticRoute() {
		utils.AviLog.Debugf("key: %s, msg: static route sync disabled", key)
		if lib.StaticRouteSyncChan != nil {
			close(lib.StaticRouteSyncChan)
			lib.StaticRouteSyncChan = nil
		}
		return
	}
	// Disable static route sync if ako is in  NodePort mode
	if lib.IsNodePortMode() {
		utils.AviLog.Debugf("key: %s, msg: static route sync disabled in NodePort Mode", key)
		return
	}
	vrfNode := avimodel.GetAviVRF()
	if len(vrfNode) != 1 {
		utils.AviLog.Warnf("key: %s, msg: Number of vrf nodes is not one", key)
		if lib.StaticRouteSyncChan != nil {
			close(lib.StaticRouteSyncChan)
			lib.StaticRouteSyncChan = nil
		}
		return
	}
	aviVrfNode := vrfNode[0]
	vrfCacheObj := rest.getVrfCacheObj(vrfName)
	if vrfCacheObj == nil {
		utils.AviLog.Warnf("key: %s, vrf %s not found in cache, exiting", key, vrfName)
		if lib.StaticRouteSyncChan != nil {
			close(lib.StaticRouteSyncChan)
			lib.StaticRouteSyncChan = nil
		}
		return
	}
	if vrfCacheObj.CloudConfigCksum == aviVrfNode.CloudConfigCksum {
		utils.AviLog.Debugf("key: %s, msg: checksum for vrf %s has not changed, skipping", key, vrfName)
		if lib.StaticRouteSyncChan != nil {
			close(lib.StaticRouteSyncChan)
			lib.StaticRouteSyncChan = nil
		}
		return
	}
	var restOps []*utils.RestOp
	restOp := rest.AviVrfBuild(key, aviVrfNode, vrfCacheObj.Uuid)
	if restOp == nil {
		utils.AviLog.Debugf("key: %s, no rest operation for vrf %s", key, vrfName)
		if lib.StaticRouteSyncChan != nil {
			close(lib.StaticRouteSyncChan)
			lib.StaticRouteSyncChan = nil
		}
		return
	}
	restOps = append(restOps, restOp)
	vrfKey := avicache.NamespaceName{Namespace: lib.GetTenant(), Name: vrfName}
	utils.AviLog.Debugf("key: %s, msg: Executing rest for vrf %s", key, vrfName)
	utils.AviLog.Debugf("key: %s, msg: restops %v", key, *restOp)
	success, _ := rest.ExecuteRestAndPopulateCache(restOps, vrfKey, avimodel, key, false)
	if success && lib.ConfigDeleteSyncChan != nil {
		vsKeysPending := rest.cache.VsCacheMeta.AviGetAllKeys()
		utils.AviLog.Infof("key: %s, msg: Number of VS deletion pending: %d", key, len(vsKeysPending))
		if len(vsKeysPending) == 0 {
			utils.AviLog.Debugf("key: %s, msg: sending signal for vs deletion notification", key)
			close(lib.ConfigDeleteSyncChan)
			lib.ConfigDeleteSyncChan = nil
		}
	}
}

// CheckAndPublishForRetry : Check if the error is of type 401, has string "Rest request error" or was timed out,
// then publish the key to retry layer. These error do not depend on the object state, hence cache refresh is not required.
func (rest *RestOperations) CheckAndPublishForRetry(err error, publishKey avicache.NamespaceName, key string, avimodel *nodes.AviObjectGraph) bool {
	if err == nil {
		return false
	}
	if webSyncErr, ok := err.(*utils.WebSyncError); ok {
		if aviError, ok := webSyncErr.GetWebAPIError().(session.AviError); ok {
			switch aviError.HttpStatusCode {
			case 401:
				if strings.Contains(*aviError.Message, "Invalid credentials") {
					if utils.IsVCFCluster() {
						lib.WaitForInitSecretRecreateAndReboot()
						return true
					}
					utils.AviLog.Errorf("key: %s, msg: Invalid credentials error, Shutting down API Server", key)
					lib.ShutdownApi()
				} else if avimodel != nil && avimodel.GetRetryCounter() != 0 {
					utils.AviLog.Warnf("key: %s, msg: got 401 error while executing rest request, adding to fast retry queue", key)
					rest.PublishKeyToRetryLayer(publishKey, key)
				} else {
					utils.AviLog.Warnf("key: %s, msg: got 401 error while executing rest request, adding to slow retry queue", key)
					rest.PublishKeyToSlowRetryLayer(publishKey, key)
				}
				return true
			case 400:
				if strings.Contains(*aviError.Message, lib.NoFreeIPError) {
					utils.AviLog.Warnf("key: %s, msg: no Free IP available, adding to slow retry queue", key)
					rest.PublishKeyToSlowRetryLayer(publishKey, key)
					return true
				}
				if strings.Contains(*aviError.Message, lib.VrfContextNotFoundError) || strings.Contains(*aviError.Message, lib.VrfContextObjectNotFoundError) {
					utils.AviLog.Warnf("key: %s, msg: VrfContext not found, adding to slow retry queue", key)
					rest.PublishKeyToSlowRetryLayer(publishKey, key)
					return true
				}
				if strings.Contains(*aviError.Message, lib.NetworkNotFoundError) {
					utils.AviLog.Warnf("key: %s, msg: Network object not found, adding to slow retry queue", key)
					rest.PublishKeyToSlowRetryLayer(publishKey, key)
					return true
				}
			case 403:
				if strings.Contains(*aviError.Message, lib.ConfigDisallowedDuringUpgradeError) {
					utils.AviLog.Warnf("key: %s, msg: controller upgrade in progress, adding to slow retry queue", key)
					rest.PublishKeyToSlowRetryLayer(publishKey, key)
					return true
				}
			}
		}
	}
	if strings.Contains(err.Error(), "Rest request error") || strings.Contains(err.Error(), "timed out waiting for rest response") {
		utils.AviLog.Warnf("key: %s, msg: got error while executing rest request: %s, adding to slow retry queue", key, err.Error())
		rest.PublishKeyToSlowRetryLayer(publishKey, key)
		return true
	}
	return false
}

func (rest *RestOperations) RestOperation(vsName string, namespace string, avimodel *nodes.AviObjectGraph, vs_cache_obj *avicache.AviVsCache, key string) {
	var pools_to_delete []avicache.NamespaceName
	var pgs_to_delete []avicache.NamespaceName
	var ds_to_delete []avicache.NamespaceName
	var vsvip_to_delete []avicache.NamespaceName
	var sni_to_delete []avicache.NamespaceName
	var httppol_to_delete []avicache.NamespaceName
	var l4pol_to_delete []avicache.NamespaceName
	var sslkey_cert_delete []avicache.NamespaceName
	var string_groups_to_delete []avicache.NamespaceName
	var vsvipErr error
	var publishKey string

	vsKey := avicache.NamespaceName{Namespace: namespace, Name: vsName}
	aviVsNode := avimodel.GetAviVS()[0]
	if avimodel != nil && len(avimodel.GetAviVS()) > 0 {
		publishKey = avimodel.GetAviVS()[0].Name
	}
	if publishKey == "" {
		// This is a delete case for the virtualservice. Derive the virtualservice from the 'key'
		splitKeys := strings.Split(key, "/")
		if len(splitKeys) == 2 {
			publishKey = splitKeys[1]
		}
	}
	nsPublishKey := avicache.NamespaceName{Namespace: namespace, Name: publishKey}
	// Order would be this: 1. Pools 2. PGs  3. DS. 4. SSLKeyCert 5. VS
	if vs_cache_obj != nil {
		var rest_ops []*utils.RestOp
		vsvip_to_delete, rest_ops, vsvipErr = rest.VSVipCU(aviVsNode.VSVIPRefs, vs_cache_obj, namespace, rest_ops, key)
		if vsvipErr != nil {
			if rest.CheckAndPublishForRetry(vsvipErr, nsPublishKey, key, avimodel) {
				return
			}
		}
		if aviVsNode.Dedicated {
			// CAcerts have to be created first, as they are referred by the keycerts
			sslkey_cert_delete, rest_ops = rest.CACertCU(aviVsNode.CACertRefs, vs_cache_obj.SSLKeyCertCollection, namespace, rest_ops, key)
			// SSLKeyCertCollection which did not match cacerts are present in the list sslkey_cert_delete,
			// which shuld be the new SSLKeyCertCollection
			sslkey_cert_delete, rest_ops = rest.SSLKeyCertCU(aviVsNode.SSLKeyCertRefs, sslkey_cert_delete, namespace, rest_ops, key)
		}
		pools_to_delete, rest_ops = rest.PoolCU(aviVsNode.PoolRefs, vs_cache_obj, namespace, rest_ops, key)
		pgs_to_delete, rest_ops = rest.PoolGroupCU(aviVsNode.PoolGroupRefs, vs_cache_obj, namespace, rest_ops, key)
		string_groups_to_delete, rest_ops = rest.StringGroupVsCU(aviVsNode.StringGroupRefs, vs_cache_obj, namespace, rest_ops, key)
		httppol_to_delete, rest_ops = rest.HTTPPolicyCU(aviVsNode.HttpPolicyRefs, vs_cache_obj, namespace, rest_ops, key)
		ds_to_delete, rest_ops = rest.DatascriptCU(aviVsNode.HTTPDSrefs, vs_cache_obj, namespace, rest_ops, key)
		l4pol_to_delete, rest_ops = rest.L4PolicyCU(aviVsNode.L4PolicyRefs, vs_cache_obj, namespace, rest_ops, key)
		utils.AviLog.Debugf("key: %s, msg: stored checksum for VS: %s, model checksum: %s", key, vs_cache_obj.CloudConfigCksum, strconv.Itoa(int(aviVsNode.GetCheckSum())))
		if vs_cache_obj.CloudConfigCksum == strconv.Itoa(int(aviVsNode.GetCheckSum())) {
			utils.AviLog.Debugf("key: %s, msg: the checksums are same for vs %s, not doing anything", key, vs_cache_obj.Name)
		} else {
			utils.AviLog.Debugf("key: %s, msg: the stored checksum for vs is %v, and the obtained checksum for VS is: %v", key, vs_cache_obj.CloudConfigCksum, strconv.Itoa(int(aviVsNode.GetCheckSum())))
			// The checksums are different, so it should be a PUT call.
			restOp := rest.AviVsBuild(aviVsNode, utils.RestPut, vs_cache_obj, key)
			if restOp != nil {
				rest_ops = append(rest_ops, restOp...)
			}

		}
		if success, _ := rest.ExecuteRestAndPopulateCache(rest_ops, vsKey, avimodel, key, false); !success {
			return
		}
	} else {
		var rest_ops []*utils.RestOp
		_, rest_ops, vsvipErr = rest.VSVipCU(aviVsNode.VSVIPRefs, nil, namespace, rest_ops, key)
		if vsvipErr != nil {
			if rest.CheckAndPublishForRetry(vsvipErr, nsPublishKey, key, avimodel) {
				return
			}
		}
		if aviVsNode.Dedicated {
			_, rest_ops = rest.CACertCU(aviVsNode.CACertRefs, []avicache.NamespaceName{}, namespace, rest_ops, key)
			_, rest_ops = rest.SSLKeyCertCU(aviVsNode.SSLKeyCertRefs, nil, namespace, rest_ops, key)
		}
		_, rest_ops = rest.PoolCU(aviVsNode.PoolRefs, nil, namespace, rest_ops, key)
		_, rest_ops = rest.PoolGroupCU(aviVsNode.PoolGroupRefs, nil, namespace, rest_ops, key)
		_, rest_ops = rest.StringGroupVsCU(aviVsNode.StringGroupRefs, nil, namespace, rest_ops, key)
		_, rest_ops = rest.HTTPPolicyCU(aviVsNode.HttpPolicyRefs, nil, namespace, rest_ops, key)
		_, rest_ops = rest.L4PolicyCU(aviVsNode.L4PolicyRefs, nil, namespace, rest_ops, key)
		_, rest_ops = rest.DatascriptCU(aviVsNode.HTTPDSrefs, nil, namespace, rest_ops, key)
		// The cache was not found - it's a POST call.
		restOp := rest.AviVsBuild(aviVsNode, utils.RestPost, nil, key)
		if restOp != nil {
			rest_ops = append(rest_ops, restOp...)
		}
		utils.AviLog.Debugf("POST key: %s, vsKey: %s", key, vsKey)
		utils.AviLog.Debugf("POST restops %s", utils.Stringify(rest_ops))
		if success, _ := rest.ExecuteRestAndPopulateCache(rest_ops, vsKey, avimodel, key, false); !success {
			return
		}
	}
	if vs_cache_obj != nil {
		for _, sni_uuid := range vs_cache_obj.SNIChildCollection {
			sni_vs_key, ok := rest.cache.VsCacheMeta.AviCacheGetKeyByUuid(sni_uuid)
			if ok {
				sni_to_delete = append(sni_to_delete, sni_vs_key.(avicache.NamespaceName))
			} else {
				utils.AviLog.Debugf("key: %s, msg: Couldn't get SNI key for uuid: %v", key, sni_uuid)
			}
		}
	}
	var rest_ops []*utils.RestOp
	vsKey = avicache.NamespaceName{Namespace: namespace, Name: vsName}
	rest_ops = rest.VSVipDelete(vsvip_to_delete, namespace, rest_ops, key)
	if aviVsNode.Dedicated {
		rest_ops = rest.SSLKeyCertDelete(sslkey_cert_delete, namespace, rest_ops, key)
	}
	rest_ops = rest.HTTPPolicyDelete(httppol_to_delete, namespace, rest_ops, key)
	rest_ops = rest.StringGroupDelete(string_groups_to_delete, namespace, rest_ops, key)
	rest_ops = rest.L4PolicyDelete(l4pol_to_delete, namespace, rest_ops, key)
	rest_ops = rest.DSDelete(ds_to_delete, namespace, rest_ops, key)
	rest_ops = rest.PoolGroupDelete(pgs_to_delete, namespace, rest_ops, key)
	rest_ops = rest.PoolDelete(pools_to_delete, namespace, rest_ops, key)
	if success, _ := rest.ExecuteRestAndPopulateCache(rest_ops, vsKey, avimodel, key, false); !success {
		return
	}

	for _, sni_node := range aviVsNode.SniNodes {
		utils.AviLog.Debugf("key: %s, msg: processing sni node: %s", key, sni_node.Name)
		utils.AviLog.Debugf("key: %s, msg: probable SNI delete candidates: %s", key, sni_to_delete)
		var rest_ops []*utils.RestOp
		vsKey = avicache.NamespaceName{Namespace: namespace, Name: sni_node.Name}
		if vs_cache_obj != nil {
			sni_to_delete, rest_ops = rest.SNINodeCU(sni_node, vs_cache_obj, namespace, sni_to_delete, rest_ops, key)
		} else {
			_, rest_ops = rest.SNINodeCU(sni_node, nil, namespace, sni_to_delete, rest_ops, key)
		}
		if success, processNextChild := rest.ExecuteRestAndPopulateCache(rest_ops, vsKey, avimodel, key, false); !success {
			if !processNextChild {
				utils.AviLog.Infof("key: %s, msg: Failure in processing SNI node: %s. Not processing other child nodes.", key, sni_node.Name)
				return
			}
		}
	}

	// Let's populate all the DELETE entries
	if len(sni_to_delete) > 0 {
		utils.AviLog.Infof("key: %s, msg: SNI delete candidates are : %s", key, sni_to_delete)
		var rest_ops []*utils.RestOp
		for _, del_sni := range sni_to_delete {
			rest.SNINodeDelete(del_sni, namespace, rest_ops, avimodel, key)
			if success, _ := rest.ExecuteRestAndPopulateCache(rest_ops, vsKey, avimodel, key, false); !success {
				return
			}
		}
	}

	for _, passChildNode := range aviVsNode.PassthroughChildNodes {
		var rest_ops []*utils.RestOp
		passChildVSKey := avicache.NamespaceName{Namespace: namespace, Name: passChildNode.Name}
		passChildVSCacheObj := rest.getVsCacheObj(passChildVSKey, key)
		utils.AviLog.Debugf("key: %s, msg: processing passthrough node: %s", key, passChildNode)
		vsKey = avicache.NamespaceName{Namespace: namespace, Name: passChildNode.Name}
		if passChildVSCacheObj != nil {
			rest_ops = rest.PassthroughChildCU(passChildNode, passChildVSCacheObj, namespace, rest_ops, key)
		} else {
			rest_ops = rest.PassthroughChildCU(passChildNode, nil, namespace, rest_ops, key)
		}
		if success, _ := rest.ExecuteRestAndPopulateCache(rest_ops, vsKey, avimodel, key, false); !success {
			return
		}
	}
}

func (rest *RestOperations) PassthroughChildCU(passChildNode *nodes.AviVsNode, vsCacheObj *avicache.AviVsCache, namespace string, restOps []*utils.RestOp, key string) []*utils.RestOp {
	var httpPoliciesToDelete []avicache.NamespaceName
	if vsCacheObj != nil {
		utils.AviLog.Debugf("key: %s, msg: Cache Passthrough Node - %s", key, utils.Stringify(vsCacheObj))
		httpPoliciesToDelete, restOps = rest.HTTPPolicyCU(passChildNode.HttpPolicyRefs, vsCacheObj, namespace, restOps, key)

		// The checksums are different, so it should be a PUT call.
		if vsCacheObj.CloudConfigCksum != strconv.Itoa(int(passChildNode.GetCheckSum())) {
			restOp := rest.AviVsBuild(passChildNode, utils.RestPut, vsCacheObj, key)
			if restOp != nil {
				restOps = append(restOps, restOp...)
			}
			utils.AviLog.Debugf("key: %s, msg: the checksums are different for passthrough child %s, operation: PUT", key, passChildNode.Name)
		}
		restOps = rest.HTTPPolicyDelete(httpPoliciesToDelete, namespace, restOps, key)

	} else {
		utils.AviLog.Infof("key: %s, msg: passthrough Child %s not found in cache", key, passChildNode.Name)
		_, restOps = rest.HTTPPolicyCU(passChildNode.HttpPolicyRefs, nil, namespace, restOps, key)

		// Not found - it should be a POST call.
		restOp := rest.AviVsBuild(passChildNode, utils.RestPost, nil, key)
		if restOp != nil {
			restOps = append(restOps, restOp...)
		}
	}
	return restOps
}

func (rest *RestOperations) getVsCacheObj(vsKey avicache.NamespaceName, key string) *avicache.AviVsCache {
	vs_cache, found := rest.cache.VsCacheMeta.AviCacheGet(vsKey)
	if found {
		vs_cache_obj, ok := vs_cache.(*avicache.AviVsCache)
		if !ok {
			utils.AviLog.Warnf("key: %s, msg: invalid vs object found, cannot cast. Not doing anything", key)
			return nil
		}
		return vs_cache_obj
	}
	utils.AviLog.Infof("key: %s, msg: vs cache object NOT found for vskey: %s", key, vsKey)
	return nil
}

func (rest *RestOperations) DeleteVSOper(vsKey avicache.NamespaceName, vs_cache_obj *avicache.AviVsCache, namespace string, key string, skipVS, skipVSVip bool) bool {
	var rest_ops []*utils.RestOp
	if vs_cache_obj != nil {
		sni_vs_keys := make([]string, len(vs_cache_obj.SNIChildCollection))
		copy(sni_vs_keys, vs_cache_obj.SNIChildCollection)

		// VS delete should delete everything together.
		passthroughChild := vs_cache_obj.ServiceMetadataObj.PassthroughChildRef
		if passthroughChild != "" {
			passthroughChildKey := avicache.NamespaceName{
				Namespace: namespace,
				Name:      passthroughChild,
			}
			passthroughChildCache := rest.getVsCacheObj(passthroughChildKey, key)
			if success := rest.DeleteVSOper(passthroughChildKey, passthroughChildCache, namespace, key, skipVS, true); !success {
				return false
			}
		}
		for _, sni_uuid := range sni_vs_keys {
			sniVsKey, ok := rest.cache.VsCacheMeta.AviCacheGetKeyByUuid(sni_uuid)
			if ok {
				delSNI := sniVsKey.(avicache.NamespaceName)
				if !rest.SNINodeDelete(delSNI, namespace, rest_ops, nil, key) {
					return false
				}
			}
		}
		if !skipVS {
			rest_op, ok := rest.AviVSDel(vs_cache_obj.Uuid, namespace, key)
			if ok {
				rest_ops = append(rest_ops, rest_op)
			}
		}
		if !skipVSVip {
			rest_ops = rest.VSVipDelete(vs_cache_obj.VSVipKeyCollection, namespace, rest_ops, key)
		}
		rest_ops = rest.DSDelete(vs_cache_obj.DSKeyCollection, namespace, rest_ops, key)
		rest_ops = rest.SSLKeyCertDelete(vs_cache_obj.SSLKeyCertCollection, namespace, rest_ops, key)
		rest_ops = rest.HTTPPolicyDelete(vs_cache_obj.HTTPKeyCollection, namespace, rest_ops, key)
		rest_ops = rest.StringGroupDelete(vs_cache_obj.StringGroupKeyCollection, namespace, rest_ops, key)
		rest_ops = rest.L4PolicyDelete(vs_cache_obj.L4PolicyCollection, namespace, rest_ops, key)
		rest_ops = rest.PoolGroupDelete(vs_cache_obj.PGKeyCollection, namespace, rest_ops, key)
		rest_ops = rest.PoolDelete(vs_cache_obj.PoolKeyCollection, namespace, rest_ops, key)
		success, _ := rest.ExecuteRestAndPopulateCache(rest_ops, vsKey, nil, key, false)
		if success {
			vsKeysPending := rest.cache.VsCacheMeta.AviGetAllKeys()
			utils.AviLog.Infof("key: %s, msg: Number of VS deletion pending: %d", key, len(vsKeysPending))
			if len(vsKeysPending) == 0 {
				// All VSes got deleted, done with deleteConfig operation. Now notify the user
				if lib.ConfigDeleteSyncChan != nil {
					utils.AviLog.Debugf("key: %s, msg: sending signal for vs deletion notification", key)
					close(lib.ConfigDeleteSyncChan)
					lib.ConfigDeleteSyncChan = nil
				}
			}
		}
		return success
	}

	// All VSes got deleted, done with deleteConfig operation. Now notify the user
	if lib.ConfigDeleteSyncChan != nil {
		utils.AviLog.Debugf("key: %s, msg: sending signal for vs deletion notification", key)
		close(lib.ConfigDeleteSyncChan)
		lib.ConfigDeleteSyncChan = nil
	}

	return true
}

func (rest *RestOperations) deleteSniVs(vsKey avicache.NamespaceName, vs_cache_obj *avicache.AviVsCache, avimodel *nodes.AviObjectGraph, namespace, key string) bool {
	var rest_ops []*utils.RestOp

	if vs_cache_obj != nil {
		rest_op, ok := rest.AviVSDel(vs_cache_obj.Uuid, namespace, key)
		if ok {
			rest_ops = append(rest_ops, rest_op)
		}
		rest_ops = rest.DSDelete(vs_cache_obj.DSKeyCollection, namespace, rest_ops, key)
		rest_ops = rest.SSLKeyCertDelete(vs_cache_obj.SSLKeyCertCollection, namespace, rest_ops, key)
		rest_ops = rest.HTTPPolicyDelete(vs_cache_obj.HTTPKeyCollection, namespace, rest_ops, key)
		rest_ops = rest.PoolGroupDelete(vs_cache_obj.PGKeyCollection, namespace, rest_ops, key)
		rest_ops = rest.PoolDelete(vs_cache_obj.PoolKeyCollection, namespace, rest_ops, key)
		rest_ops = rest.StringGroupDelete(vs_cache_obj.StringGroupKeyCollection, namespace, rest_ops, key)
		success, _ := rest.ExecuteRestAndPopulateCache(rest_ops, vsKey, avimodel, key, false)
		return success
	}
	return true
}

func (rest *RestOperations) ExecuteRestAndPopulateCache(rest_ops []*utils.RestOp, aviObjKey avicache.NamespaceName, avimodel *nodes.AviObjectGraph, key string, isEvh bool, sslKey ...utils.NamespaceName) (bool, bool) {
	// Choose a avi client based on the model name hash. This would ensure that the same worker queue processes updates for a given VS all the time.
	shardSize := lib.GetshardSize()
	if shardSize == 0 {
		// Dedicated VS case
		shardSize = 8
	}
	var retry, fastRetry, processNextObj bool
	bkt := utils.Bkt(key, shardSize)
	aviRestPoolClient := avicache.SharedAVIClients(aviObjKey.Namespace)
	if len(aviRestPoolClient.AviClient) > 0 && len(rest_ops) > 0 {
		utils.AviLog.Infof("key: %s, msg: processing in rest queue number: %v", key, bkt)
		aviclient := aviRestPoolClient.AviClient[bkt]
		err := rest.AviRestOperateWrapper(aviclient, rest_ops, key)
		if err == nil {
			models.RestStatus.UpdateAviApiRestStatus(utils.AVIAPI_CONNECTED, nil)
			utils.AviLog.Debugf("key: %s, msg: rest call executed successfully, will update cache", key)

			// Add to local obj caches
			for _, rest_op := range rest_ops {
				rest.PopulateOneCache(rest_op, aviObjKey, key)
			}

		} else if aviObjKey.Name == lib.DummyVSForStaleData {
			utils.AviLog.Warnf("key: %s, msg: error in rest request %v, for %s, won't retry", key, err.Error(), aviObjKey.Name)
			return false, processNextObj
		} else {
			var publishKey string
			if avimodel != nil && isEvh && len(avimodel.GetAviEvhVS()) > 0 {
				publishKey = avimodel.GetAviEvhVS()[0].Name
			} else if avimodel != nil && !isEvh && len(avimodel.GetAviVS()) > 0 {
				publishKey = avimodel.GetAviVS()[0].Name
			}

			if publishKey == "" {
				// This is a delete case for the virtualservice. Derive the virtualservice from the 'key'
				splitKeys := strings.Split(key, "/")
				if len(splitKeys) == 2 {
					publishKey = splitKeys[1]
				}
			}
			nsPublishKey := avicache.NamespaceName{Namespace: aviObjKey.Namespace, Name: publishKey}

			if rest.restOperator.isRetryRequired(key, err) {
				rest.PublishKeyToRetryLayer(nsPublishKey, key)
				return false, processNextObj
			}

			if rest.CheckAndPublishForRetry(err, nsPublishKey, key, avimodel) {
				return false, processNextObj
			}
			utils.AviLog.Warnf("key: %s, msg: there was an error sending the macro %v", key, err.Error())
			models.RestStatus.UpdateAviApiRestStatus("", err)

			for i := len(rest_ops) - 1; i >= 0; i-- {
				// Go over each of the failed requests and enqueue them to the worker queue for retry.
				if rest_ops[i].Err != nil {
					// check for VSVIP errors for blocked IP address updates
					if checkVsVipUpdateErrors(key, rest_ops[i]) {
						rest.PopulateOneCache(rest_ops[i], aviObjKey, key)
						continue
					}
					if rest_ops[i].Obj != nil && rest_ops[i].Caller != "" {
						updateGatewayStatusWithVsError(key, rest_ops[i])
					}
					// If it's for a SNI child, publish the parent VS's key
					refreshCacheForRetry := false
					if avimodel != nil && isEvh && len(avimodel.GetAviEvhVS()) > 0 {
						refreshCacheForRetry = true
					} else if avimodel != nil && !isEvh && len(avimodel.GetAviVS()) > 0 {
						refreshCacheForRetry = true
					}
					if refreshCacheForRetry {
						utils.AviLog.Warnf("key: %s, msg: Retrieved key for Retry:%s, object: %s", key, publishKey, rest_ops[i].ObjName)
						aviError, ok := rest_ops[i].Err.(session.AviError)
						if !ok {
							utils.AviLog.Infof("key: %s, msg: Error is not of type AviError, err: %v, %T", key, rest_ops[i].Err, rest_ops[i].Err)
							continue
						}
						aviclient := aviRestPoolClient.AviClient[bkt]
						retryable, fastRetryable, nextObj := rest.RefreshCacheForRetryLayer(publishKey, aviObjKey, rest_ops[i], aviError, aviclient, avimodel, key, isEvh)
						retry = retry || retryable
						processNextObj = processNextObj || nextObj
						if avimodel.GetRetryCounter() != 0 {
							fastRetry = fastRetry || fastRetryable
						} else {
							fastRetry = false
							utils.AviLog.Warnf("key: %s, msg: retry count exhausted, would be added to slow retry queue", key)
						}
					} else {
						utils.AviLog.Warnf("key: %s, msg: Avi model not set, possibly a DELETE call", key)
						aviError, ok := rest_ops[i].Err.(session.AviError)
						// If it's 404, don't retry
						if ok {
							statuscode := aviError.HttpStatusCode
							if statuscode != 404 {
								if statuscode == 412 {
									// concurrent update scenario currently happens for VRFContext only
									rest.PublishKeyToRetryLayer(nsPublishKey, key)
								} else {
									rest.PublishKeyToSlowRetryLayer(nsPublishKey, key)
								}
								return false, true
							} else {
								rest.AviVsCacheDel(rest_ops[i], aviObjKey, key)
							}
						}
					}
				} else {
					rest.PopulateOneCache(rest_ops[i], aviObjKey, key)
				}
			}

			if retry {
				if fastRetry {
					rest.PublishKeyToRetryLayer(nsPublishKey, key)
				} else {
					rest.PublishKeyToSlowRetryLayer(nsPublishKey, key)
				}
			}
			return false, processNextObj
		}
	}
	return true, true
}

func updateGatewayStatusWithVsError(key string, rest_op *utils.RestOp) {
	if rest_op.Model == "VirtualService" {
		var vs avimodels.VirtualService
		switch rest_op.Obj.(type) {
		case avimodels.VirtualService:
			vs = rest_op.Obj.(avimodels.VirtualService)
		case *avimodels.VirtualService:
			vs = *(rest_op.Obj.(*avimodels.VirtualService))
		default:
			return
		}
		if vs.Type != nil && *vs.Type == utils.VS_TYPE_VH_PARENT {
			if vs.ServiceMetadata != nil {
				var svc_mdata_obj lib.ServiceMetadataObj
				utils.AviLog.Infof("key:%s, msg: Service Metadata: %s", key, *vs.ServiceMetadata)
				if err := json.Unmarshal([]byte(*vs.ServiceMetadata),
					&svc_mdata_obj); err != nil {
					utils.AviLog.Warnf("Error parsing service metadata :%v", err)
					return
				}
				if svc_mdata_obj.Gateway != "" {
					updateOptions := status.UpdateOptions{
						ServiceMetadata: svc_mdata_obj,
						Key:             key,
						VSName:          rest_op.ObjName,
						Message:         rest_op.Err.Error(),
						Tenant:          rest_op.Tenant,
					}
					statusOption := status.StatusOptions{
						ObjType: lib.Gateway,
						Op:      lib.UpdateStatus,
						Key:     key,
						Options: &updateOptions,
					}
					status.PublishToStatusQueue(updateOptions.ServiceMetadata.Gateway, statusOption)
				}
			}
		}
	}
}

func checkVsVipUpdateErrors(key string, rest_op *utils.RestOp) bool {
	if aviError, ok := rest_op.Err.(session.AviError); ok {
		if aviError.HttpStatusCode == 400 &&
			rest_op.Model == "VsVip" &&
			(strings.Contains(rest_op.Err.Error(), lib.AviControllerVSVipIDChangeError) ||
				strings.Contains(rest_op.Err.Error(), lib.AviControllerRecreateVIPError)) {
			utils.AviLog.Warnf("key: %s, msg: Unsupported call for vsvip %v", key, rest_op.Err.Error())
			// this adds error as a message, useful for sending Avi errors to k8s object statuses, if required
			rest_op.Message = *aviError.Message
			return true
		}
	}
	return false
}

func (rest *RestOperations) PopulateOneCache(rest_op *utils.RestOp, aviObjKey avicache.NamespaceName, key string) {
	aviErr, ok := rest_op.Err.(session.AviError)
	if !ok && rest_op.Err != nil {
		utils.AviLog.Warnf("key: %s, msg: Error in rest operation is not of type AviError, err: %v, %T", key, rest_op.Err, rest_op.Err)
	}
	if (rest_op.Err == nil || rest_op.Message != "") &&
		(rest_op.Method == utils.RestPost ||
			rest_op.Method == utils.RestPut ||
			rest_op.Method == utils.RestPatch) {
		utils.AviLog.Infof("key: %s, msg: creating/updating %s cache, method: %s", key, rest_op.Model, rest_op.Method)
		if rest_op.Model == "PKIprofile" {
			rest.AviPkiProfileAdd(rest_op, aviObjKey, key)
		} else if rest_op.Model == "Pool" {
			rest.AviPoolCacheAdd(rest_op, aviObjKey, key)
		} else if rest_op.Model == "VirtualService" {
			rest.AviVsCacheAdd(rest_op, key)
		} else if rest_op.Model == "PoolGroup" {
			rest.AviPGCacheAdd(rest_op, aviObjKey, key)
		} else if rest_op.Model == "VSDataScriptSet" {
			rest.AviDSCacheAdd(rest_op, aviObjKey, key)
		} else if rest_op.Model == "HTTPPolicySet" {
			rest.AviHTTPPolicyCacheAdd(rest_op, aviObjKey, key)
		} else if rest_op.Model == "SSLKeyAndCertificate" {
			rest.AviSSLKeyCertAdd(rest_op, aviObjKey, key)
		} else if rest_op.Model == "L4PolicySet" {
			rest.AviL4PolicyCacheAdd(rest_op, aviObjKey, key)
		} else if rest_op.Model == "VrfContext" {
			rest.AviVrfCacheAdd(rest_op, aviObjKey, key)
		} else if rest_op.Model == "VsVip" {
			rest.AviVsVipCacheAdd(rest_op, aviObjKey, key)
		} else if rest_op.Model == "StringGroup" {
			rest.AviStringGroupCacheAdd(rest_op, aviObjKey, key)
		}

	} else if (rest_op.Err == nil || aviErr.HttpStatusCode == 404) &&
		rest_op.Method == utils.RestDelete {
		utils.AviLog.Infof("key: %s, msg: deleting %s cache", key, rest_op.Model)
		if rest_op.Model == "PKIprofile" {
			rest.AviPkiProfileCacheDel(rest_op, aviObjKey, key)
		} else if rest_op.Model == "Pool" {
			rest.AviPoolCacheDel(rest_op, aviObjKey, key)
		} else if rest_op.Model == "VirtualService" {
			rest.AviVsCacheDel(rest_op, aviObjKey, key)
		} else if rest_op.Model == "PoolGroup" {
			rest.AviPGCacheDel(rest_op, aviObjKey, key)
		} else if rest_op.Model == "HTTPPolicySet" {
			rest.AviHTTPPolicyCacheDel(rest_op, aviObjKey, key)
		} else if rest_op.Model == "SSLKeyAndCertificate" {
			rest.AviSSLCacheDel(rest_op, aviObjKey, key)
		} else if rest_op.Model == "L4PolicySet" {
			rest.AviL4PolicyCacheDel(rest_op, aviObjKey, key)
		} else if rest_op.Model == "VsVip" {
			rest.AviVsVipCacheDel(rest_op, aviObjKey, key)
		} else if rest_op.Model == "VSDataScriptSet" {
			rest.AviDSCacheDel(rest_op, aviObjKey, key)
		} else if rest_op.Model == "StringGroup" {
			rest.AviStringGroupCacheDel(rest_op, aviObjKey, key)
		}
	}
}

func (rest *RestOperations) PublishKeyToRetryLayer(parentVsKey avicache.NamespaceName, key string) {
	fastRetryQueue := utils.SharedWorkQueue().GetQueueByName(lib.FAST_RETRY_LAYER)
	fastRetryQueue.Workqueue[0].AddRateLimited(fmt.Sprintf("%s/%s", parentVsKey.Namespace, parentVsKey.Name))
	lib.IncrementQueueCounter(lib.FAST_RETRY_LAYER)
	utils.AviLog.Infof("key: %s, msg: Published key with vs_key to fast path retry queue: %s", key, parentVsKey)
}

func (rest *RestOperations) PublishKeyToSlowRetryLayer(parentVsKey avicache.NamespaceName, key string) {
	slowRetryQueue := utils.SharedWorkQueue().GetQueueByName(lib.SLOW_RETRY_LAYER)
	slowRetryQueue.Workqueue[0].AddRateLimited(fmt.Sprintf("%s/%s", parentVsKey.Namespace, parentVsKey.Name))
	lib.IncrementQueueCounter(lib.SLOW_RETRY_LAYER)
	utils.AviLog.Infof("key: %s, msg: Published key with vs_key to slow path retry queue: %s", key, parentVsKey)
}

func (rest *RestOperations) AviRestOperateWrapper(aviClient *clients.AviClient, rest_ops []*utils.RestOp, key string) error {
	restTimeoutChan := make(chan error, 1)
	go func() {
		err := rest.restOperator.AviRestOperate(aviClient, rest_ops, key)
		restTimeoutChan <- err
	}()
	select {
	case err := <-restTimeoutChan:
		return err
	case <-time.After(lib.ControllerReqWaitTime * time.Second):
		utils.AviLog.Warnf("timed out waiting for rest response after %d seconds", lib.ControllerReqWaitTime)
		return errors.New("timed out waiting for rest response")
	}
}

func (rest *RestOperations) RefreshCacheForRetryLayer(parentVsKey string, aviObjKey avicache.NamespaceName, rest_op *utils.RestOp, aviError session.AviError, c *clients.AviClient, avimodel *nodes.AviObjectGraph, key string, isEvh bool) (bool, bool, bool) {
	var fastRetry bool
	statuscode := aviError.HttpStatusCode
	errorStr := aviError.Error()
	retry := true
	processNextObj := true
	utils.AviLog.Warnf("key: %s, msg: problem in processing request for: %s", key, rest_op.Model)
	utils.AviLog.Infof("key: %s, msg: error str: %s", key, errorStr)
	aviObjCache := avicache.SharedAviObjCache()

	if statuscode >= 500 && statuscode < 599 {
		fastRetry = true
		processNextObj = false
	} else if statuscode >= 400 && statuscode < 499 { // Will account for more error codes.*/
		fastRetry = true
		// 404 means the object exists in our cache but not on the controller.
		if statuscode == 404 {
			switch rest_op.Model {
			case "Pool":
				var poolObjName string
				switch rest_op.Obj.(type) {
				case utils.AviRestObjMacro:
					poolObjName = *rest_op.Obj.(utils.AviRestObjMacro).Data.(avimodels.Pool).Name
				case avimodels.Pool:
					poolObjName = *rest_op.Obj.(avimodels.Pool).Name
				}
				if poolObjName != "" {
					rest_op.ObjName = poolObjName
				}
				rest.AviPoolCacheDel(rest_op, aviObjKey, key)
			case "PoolGroup":
				var pgObjName string
				switch rest_op.Obj.(type) {
				case utils.AviRestObjMacro:
					pgObjName = *rest_op.Obj.(utils.AviRestObjMacro).Data.(avimodels.PoolGroup).Name
				case avimodels.PoolGroup:
					pgObjName = *rest_op.Obj.(avimodels.PoolGroup).Name
				}
				if pgObjName != "" {
					rest_op.ObjName = pgObjName
				}
				if strings.Contains(errorStr, "Pool object not found!") {
					// PG error with pool object not found.
					aviObjCache.AviPopulateOnePGCache(c, utils.CloudName, pgObjName)
					// After the refresh - get the members
					pgKey := avicache.NamespaceName{Namespace: aviObjKey.Namespace, Name: pgObjName}
					pgCache, ok := rest.cache.PgCache.AviCacheGet(pgKey)
					if ok {
						pgCacheObj, _ := pgCache.(*avicache.AviPGCache)
						// Iterate the pools
						var poolGroupRefs []*nodes.AviPoolGroupNode
						if isEvh {
							evhVsNode := avimodel.GetAviEvhVS()[0]
							poolGroupRefs = evhVsNode.PoolGroupRefs
						} else {
							vsNode := avimodel.GetAviVS()[0]
							poolGroupRefs = vsNode.PoolGroupRefs
						}

						var pools []string
						for _, pgNode := range poolGroupRefs {
							if pgNode.Name == pgObjName {
								for _, poolInModel := range pgNode.Members {
									poolToken := strings.Split(*poolInModel.PoolRef, "?name=")
									if len(poolToken) > 1 {
										pools = append(pools, poolToken[1])
									}
								}
							}
						}
						utils.AviLog.Debugf("key: %s, msg: pools in model during retry: %s", key, pools)
						// Find out pool members that exist in the model but do not exist in the cache and delete them.

						poolsCopy := make([]string, len(pools))
						copy(poolsCopy, pools)
						for _, poolName := range pgCacheObj.Members {
							if utils.HasElem(pools, poolName) {
								poolsCopy = utils.Remove(poolsCopy, poolName)
							}
						}
						// Whatever is left it in poolsCopy - remove them from the avi pools cache
						for _, poolsToDel := range poolsCopy {
							rest_op.ObjName = poolsToDel
							utils.AviLog.Debugf("key: %s, msg: deleting pool from cache due to pool not found %s", key, poolsToDel)
							rest.AviPoolCacheDel(rest_op, aviObjKey, key)
						}
					} else {
						utils.AviLog.Infof("key: %s, msg: PG object not found during retry pgname: %s", key, pgObjName)
					}
				}
				rest.AviPGCacheDel(rest_op, aviObjKey, key)
			case "VsVip":
				var VsVip string
				switch rest_op.Obj.(type) {
				case utils.AviRestObjMacro:
					VsVip = *rest_op.Obj.(utils.AviRestObjMacro).Data.(avimodels.VsVip).Name
				case avimodels.VsVip:
					VsVip = *rest_op.Obj.(avimodels.VsVip).Name
				}
				if VsVip != "" {
					rest_op.ObjName = VsVip
				}
				rest.AviVsVipCacheDel(rest_op, aviObjKey, key)
			case "StringGroup":
				var StringGroup string
				switch rest_op.Obj.(type) {
				case utils.AviRestObjMacro:
					StringGroup = *rest_op.Obj.(utils.AviRestObjMacro).Data.(avimodels.StringGroup).Name
				case avimodels.StringGroup:
					StringGroup = *rest_op.Obj.(avimodels.StringGroup).Name
				}
				if StringGroup != "" {
					rest_op.ObjName = StringGroup
				}
				rest.AviStringGroupCacheDel(rest_op, aviObjKey, key)
			case "HTTPPolicySet":
				var HTTPPolicySet string
				switch rest_op.Obj.(type) {
				case utils.AviRestObjMacro:
					HTTPPolicySet = *rest_op.Obj.(utils.AviRestObjMacro).Data.(avimodels.HTTPPolicySet).Name
				case avimodels.HTTPPolicySet:
					HTTPPolicySet = *rest_op.Obj.(avimodels.HTTPPolicySet).Name
				}
				if HTTPPolicySet != "" {
					rest_op.ObjName = HTTPPolicySet
				}
				rest.AviHTTPPolicyCacheDel(rest_op, aviObjKey, key)
			case "L4PolicySet":
				var L4PolicySet string
				switch rest_op.Obj.(type) {
				case utils.AviRestObjMacro:
					L4PolicySet = *rest_op.Obj.(utils.AviRestObjMacro).Data.(avimodels.L4PolicySet).Name
				case avimodels.L4PolicySet:
					L4PolicySet = *rest_op.Obj.(avimodels.L4PolicySet).Name
				}
				if L4PolicySet != "" {
					rest_op.ObjName = L4PolicySet
				}
				rest.AviL4PolicyCacheDel(rest_op, aviObjKey, key)
			case "SSLKeyAndCertificate":
				var SSLKeyAndCertificate string
				switch rest_op.Obj.(type) {
				case utils.AviRestObjMacro:
					SSLKeyAndCertificate = *rest_op.Obj.(utils.AviRestObjMacro).Data.(avimodels.SSLKeyAndCertificate).Name
				case avimodels.SSLKeyAndCertificate:
					SSLKeyAndCertificate = *rest_op.Obj.(avimodels.SSLKeyAndCertificate).Name
				}
				if SSLKeyAndCertificate != "" {
					rest_op.ObjName = SSLKeyAndCertificate
				}
				rest.AviSSLCacheDel(rest_op, aviObjKey, key)
			case "PKIprofile":
				var PKIprofile string
				switch rest_op.Obj.(type) {
				case utils.AviRestObjMacro:
					PKIprofile = *rest_op.Obj.(utils.AviRestObjMacro).Data.(avimodels.PKIprofile).Name
				case avimodels.PKIprofile:
					PKIprofile = *rest_op.Obj.(avimodels.PKIprofile).Name
				}
				if PKIprofile != "" {
					rest_op.ObjName = PKIprofile
				}
				rest.AviPkiProfileCacheDel(rest_op, aviObjKey, key)
			case "VirtualService":
				rest.AviVsCacheDel(rest_op, aviObjKey, key)
			case "VSDataScriptSet":
				var VSDataScriptSet string
				switch rest_op.Obj.(type) {
				case utils.AviRestObjMacro:
					VSDataScriptSet = *rest_op.Obj.(utils.AviRestObjMacro).Data.(avimodels.VSDataScriptSet).Name
				case avimodels.VSDataScriptSet:
					VSDataScriptSet = *rest_op.Obj.(avimodels.VSDataScriptSet).Name
				}
				if VSDataScriptSet != "" {
					rest_op.ObjName = VSDataScriptSet
				}
				rest.AviDSCacheDel(rest_op, aviObjKey, key)
			}
		} else if statuscode == 409 {

			// TODO (sudswas): if error code 400 happens, it means layer 2's model has issue - can re-trigger a model eval in that case?
			// If it's 409 it refers to a conflict. That means the cache should be refreshed for the particular object.

			utils.AviLog.Infof("key: %s, msg: Conflict for object: %s of type :%s", key, rest_op.ObjName, rest_op.Model)
			switch rest_op.Model {
			case "Pool":
				var poolObjName string
				switch rest_op.Obj.(type) {
				case utils.AviRestObjMacro:
					poolObjName = *rest_op.Obj.(utils.AviRestObjMacro).Data.(avimodels.Pool).Name
				case avimodels.Pool:
					poolObjName = *rest_op.Obj.(avimodels.Pool).Name
				}
				aviObjCache.AviPopulateOnePoolCache(c, utils.CloudName, poolObjName)
			case "PoolGroup":
				var pgObjName string
				switch rest_op.Obj.(type) {
				case utils.AviRestObjMacro:
					pgObjName = *rest_op.Obj.(utils.AviRestObjMacro).Data.(avimodels.PoolGroup).Name
				case avimodels.PoolGroup:
					pgObjName = *rest_op.Obj.(avimodels.PoolGroup).Name
				}
				aviObjCache.AviPopulateOnePGCache(c, utils.CloudName, pgObjName)
			case "StringGroup":
				var stringGroupObjName string
				switch rest_op.Obj.(type) {
				case utils.AviRestObjMacro:
					stringGroupObjName = *rest_op.Obj.(utils.AviRestObjMacro).Data.(avimodels.StringGroup).Name
				case avimodels.StringGroup:
					stringGroupObjName = *rest_op.Obj.(avimodels.StringGroup).Name
				}
				aviObjCache.AviPopulateOneStringGroupCache(c, utils.CloudName, stringGroupObjName)
			case "VsVip":
				var VsVip string
				switch rest_op.Obj.(type) {
				case utils.AviRestObjMacro:
					VsVip = *rest_op.Obj.(utils.AviRestObjMacro).Data.(avimodels.VsVip).Name
				case avimodels.VsVip:
					VsVip = *rest_op.Obj.(avimodels.VsVip).Name
				}
				aviObjCache.AviPopulateOneVsVipCache(c, utils.CloudName, VsVip)
			case "HTTPPolicySet":
				var HTTPPolicySet string
				switch rest_op.Obj.(type) {
				case utils.AviRestObjMacro:
					HTTPPolicySet = *rest_op.Obj.(utils.AviRestObjMacro).Data.(avimodels.HTTPPolicySet).Name
				case avimodels.HTTPPolicySet:
					HTTPPolicySet = *rest_op.Obj.(avimodels.HTTPPolicySet).Name
				}
				aviObjCache.AviPopulateOneVsHttpPolCache(c, utils.CloudName, HTTPPolicySet)
			case "L4PolicySet":
				var L4PolicySet string
				switch rest_op.Obj.(type) {
				case utils.AviRestObjMacro:
					L4PolicySet = *rest_op.Obj.(utils.AviRestObjMacro).Data.(avimodels.L4PolicySet).Name
				case avimodels.L4PolicySet:
					L4PolicySet = *rest_op.Obj.(avimodels.L4PolicySet).Name
				}
				aviObjCache.AviPopulateOneVsL4PolCache(c, utils.CloudName, L4PolicySet)
			case "SSLKeyAndCertificate":
				var SSLKeyAndCertificate string
				switch rest_op.Obj.(type) {
				case utils.AviRestObjMacro:
					SSLKeyAndCertificate = *rest_op.Obj.(utils.AviRestObjMacro).Data.(avimodels.SSLKeyAndCertificate).Name
				case avimodels.SSLKeyAndCertificate:
					SSLKeyAndCertificate = *rest_op.Obj.(avimodels.SSLKeyAndCertificate).Name
				}
				aviObjCache.AviPopulateOneSSLCache(c, utils.CloudName, SSLKeyAndCertificate)
			case "PKIprofile":
				var PKIprofile string
				switch rest_op.Obj.(type) {
				case utils.AviRestObjMacro:
					PKIprofile = *rest_op.Obj.(utils.AviRestObjMacro).Data.(avimodels.PKIprofile).Name
				case avimodels.PKIprofile:
					PKIprofile = *rest_op.Obj.(avimodels.PKIprofile).Name
				}
				aviObjCache.AviPopulateOnePKICache(c, utils.CloudName, PKIprofile)
			case "VirtualService":
				aviObjCache.AviObjOneVSCachePopulate(c, utils.CloudName, aviObjKey.Name, aviObjKey.Namespace)
				vsObjMeta, ok := rest.cache.VsCacheMeta.AviCacheGet(aviObjKey)
				if !ok {
					// Object deleted
					utils.AviLog.Warnf("key: %s, msg: VS object already deleted during retry", key)
				} else {
					vsCopy, done := vsObjMeta.(*avicache.AviVsCache).GetVSCopy()
					if done {
						rest.cache.VsCacheMeta.AviCacheAdd(aviObjKey, vsCopy)
						rest.StatusUpdateForVS(rest_op.Method, vsCopy, key)
					}
				}
			case "VSDataScriptSet":
				var VSDataScriptSet string
				switch rest_op.Obj.(type) {
				case utils.AviRestObjMacro:
					VSDataScriptSet = *rest_op.Obj.(utils.AviRestObjMacro).Data.(avimodels.VSDataScriptSet).Name
				case avimodels.VSDataScriptSet:
					VSDataScriptSet = *rest_op.Obj.(avimodels.VSDataScriptSet).Name
				}
				aviObjCache.AviPopulateOneVsDSCache(c, utils.CloudName, VSDataScriptSet)
			}
		} else if statuscode == 408 {
			// This status code refers to a problem with the controller timeouts. We need to re-init the session object.
			utils.AviLog.Infof("key: %s, msg: Controller request timed out, will re-init session by retrying", key)
			processNextObj = false
		} else if statuscode == 400 && strings.Contains(*aviError.Message, lib.NoFreeIPError) {
			utils.AviLog.Infof("key: %s, msg:  msg: Got no free IP error, would be added to slow retry queue", key)
			fastRetry = false
			processNextObj = false
		} else if statuscode == 403 && strings.Contains(*aviError.Message, lib.ConfigDisallowedDuringUpgradeError) {
			utils.AviLog.Infof("key: %s, msg: Controller upgrade in progress, would be added to slow retry queue", key)
			fastRetry = false
			processNextObj = false
		} else {
			// We don't want to handle any other error code like 400 etc.
			utils.AviLog.Infof("key: %s, msg: Detected error code %d that we don't support, not going to retry", key, statuscode)
			retry = false
		}
	}

	return retry, fastRetry, processNextObj
}

func ExtractStatusCode(word string) string {
	r, _ := regexp.Compile("HTTP code: .*.;")
	result := r.FindAllString(word, -1)
	if len(result) == 1 {
		return result[0][len(result[0])-4 : len(result[0])-1]
	}
	return ""
}

func (rest *RestOperations) PoolDelete(pools_to_delete []avicache.NamespaceName, namespace string, rest_ops []*utils.RestOp, key string) []*utils.RestOp {
	utils.AviLog.Debugf("key: %s, msg: about to delete the pools %s", key, utils.Stringify(pools_to_delete))
	for _, del_pool := range pools_to_delete {
		// fetch trhe pool uuid from cache
		pool_key := avicache.NamespaceName{Namespace: namespace, Name: del_pool.Name}
		pool_cache, ok := rest.cache.PoolCache.AviCacheGet(pool_key)
		if ok {
			pool_cache_obj, _ := pool_cache.(*avicache.AviPoolCache)
			restOp := rest.AviPoolDel(pool_cache_obj.Uuid, namespace, key)
			restOp.ObjName = del_pool.Name
			rest_ops = append(rest_ops, restOp)

			pkiProfile := pool_cache_obj.PkiProfileCollection
			if pkiProfile.Name != "" {
				rest_ops = rest.PkiProfileDelete([]avicache.NamespaceName{pkiProfile}, namespace, rest_ops, key)
			}
		}
	}
	return rest_ops
}

func (rest *RestOperations) VSVipDelete(vsvip_to_delete []avicache.NamespaceName, namespace string, rest_ops []*utils.RestOp, key string) []*utils.RestOp {
	utils.AviLog.Infof("key: %s, msg: about to delete the vsvips %s", key, utils.Stringify(vsvip_to_delete))
	for _, del_vsvip := range vsvip_to_delete {
		// fetch trhe pool uuid from cache
		vsvip_key := avicache.NamespaceName{Namespace: namespace, Name: del_vsvip.Name}
		vsvip_cache, ok := rest.cache.VSVIPCache.AviCacheGet(vsvip_key)
		if ok {
			vsvip_cache_obj, _ := vsvip_cache.(*avicache.AviVSVIPCache)
			var restOp *utils.RestOp
			if lib.IsShardVS(del_vsvip.Name) && !lib.IsWCP() {
				vsvip_avi, err := rest.AviVsVipGet(key, vsvip_cache_obj.Uuid, del_vsvip.Name, vsvip_cache_obj.Tenant)
				if err != nil {
					utils.AviLog.Errorf("key: %s, msg: failed to get VS VIP %s", key, del_vsvip.Name)
					return rest_ops
				}
				// Retain the DNS info corresponding to the default FQDN.
				for i := 0; i < len(vsvip_avi.DNSInfo); i++ {
					if !strings.Contains(*vsvip_avi.DNSInfo[i].Fqdn, del_vsvip.Name) {
						vsvip_avi.DNSInfo = append(vsvip_avi.DNSInfo[:i], vsvip_avi.DNSInfo[i+1:]...)
						i--
					}
				}
				restOp = rest.AviVsVipPut(vsvip_cache_obj.Uuid, vsvip_avi, namespace, key)
			} else {
				restOp = rest.AviVsVipDel(vsvip_cache_obj.Uuid, namespace, key)
			}
			restOp.ObjName = del_vsvip.Name
			rest_ops = append(rest_ops, restOp)
		}
	}
	return rest_ops
}

func (rest *RestOperations) PoolGroupDelete(pgs_to_delete []avicache.NamespaceName, namespace string, rest_ops []*utils.RestOp, key string) []*utils.RestOp {
	utils.AviLog.Debugf("key: %s, msg: about to delete the PGs %s", key, pgs_to_delete)
	for _, del_pg := range pgs_to_delete {
		// fetch trhe pool uuid from cache
		pg_key := avicache.NamespaceName{Namespace: namespace, Name: del_pg.Name}
		pg_cache, ok := rest.cache.PgCache.AviCacheGet(pg_key)
		if ok {
			pg_cache_obj, _ := pg_cache.(*avicache.AviPGCache)
			restOp := rest.AviPGDel(pg_cache_obj.Uuid, namespace, key)
			restOp.ObjName = del_pg.Name
			rest_ops = append(rest_ops, restOp)
		}
	}
	return rest_ops
}

func (rest *RestOperations) DSDelete(ds_to_delete []avicache.NamespaceName, namespace string, rest_ops []*utils.RestOp, key string) []*utils.RestOp {
	utils.AviLog.Infof("key: %s, msg: about to delete the DS %s", key, ds_to_delete)
	for _, del_ds := range ds_to_delete {
		// fetch trhe pool uuid from cache
		ds_key := avicache.NamespaceName{Namespace: namespace, Name: del_ds.Name}
		ds_cache, ok := rest.cache.DSCache.AviCacheGet(ds_key)
		if ok {
			ds_cache_obj, _ := ds_cache.(*avicache.AviDSCache)
			restOp := rest.AviDSDel(ds_cache_obj.Uuid, namespace, key)
			restOp.ObjName = del_ds.Name
			rest_ops = append(rest_ops, restOp)
		} else {
			utils.AviLog.Debugf("key: %s, msg: ds not found in cache during delete %s", key, ds_to_delete)
		}
	}
	return rest_ops
}

func (rest *RestOperations) PoolCU(pool_nodes []*nodes.AviPoolNode, vs_cache_obj *avicache.AviVsCache, namespace string, rest_ops []*utils.RestOp, key string) ([]avicache.NamespaceName, []*utils.RestOp) {
	var cache_pool_nodes []avicache.NamespaceName
	var pool_pkiprofile_delete []avicache.NamespaceName
	if vs_cache_obj != nil {
		cache_pool_nodes = make([]avicache.NamespaceName, len(vs_cache_obj.PoolKeyCollection))
		copy(cache_pool_nodes, vs_cache_obj.PoolKeyCollection)
		utils.AviLog.Debugf("key: %s, msg: the cached pools are: %v", key, utils.Stringify(cache_pool_nodes))

		for _, pool := range pool_nodes {
			// check in the pool cache to see if this pool exists in AVI
			pool_key := avicache.NamespaceName{Namespace: namespace, Name: pool.Name}
			found := utils.HasElem(cache_pool_nodes, pool_key)
			utils.AviLog.Debugf("key: %s, msg: processing pool key: %v", key, pool_key)
			if found {
				cache_pool_nodes = avicache.RemoveNamespaceName(cache_pool_nodes, pool_key)
				utils.AviLog.Debugf("key: %s, key: the cache pool nodes are: %v", key, cache_pool_nodes)
				pool_cache, ok := rest.cache.PoolCache.AviCacheGet(pool_key)
				if ok {
					pool_cache_obj, _ := pool_cache.(*avicache.AviPoolCache)
					pool_pkiprofile_delete, rest_ops = rest.PkiProfileCU(pool.PkiProfile, pool_cache_obj, namespace, rest_ops, key)

					// Cache found. Let's compare the checksums
					utils.AviLog.Debugf("key: %s, msg: poolcache: %v", key, pool_cache_obj)
					if pool_cache_obj.CloudConfigCksum == strconv.Itoa(int(pool.GetCheckSum())) {
						utils.AviLog.Debugf("key: %s, msg: the checksums are same for pool %s, not doing anything", key, pool.Name)
					} else {
						utils.AviLog.Debugf("key: %s, msg: the checksums are different for pool %s, operation: PUT", key, pool.Name)
						// The checksums are different, so it should be a PUT call.
						restOp := rest.AviPoolBuild(pool, pool_cache_obj, key)
						if restOp != nil {
							rest_ops = append(rest_ops, restOp)
						}
					}
				}
			} else {
				utils.AviLog.Debugf("key: %s, msg: pool %s not found in cache, operation: POST", key, pool.Name)
				_, rest_ops = rest.PkiProfileCU(pool.PkiProfile, nil, namespace, rest_ops, key)
				// Not found - it should be a POST call.
				restOp := rest.AviPoolBuild(pool, nil, key)
				if restOp != nil {
					rest_ops = append(rest_ops, restOp)
				}
			}
			if len(pool_pkiprofile_delete) > 0 {
				rest_ops = rest.PkiProfileDelete(pool_pkiprofile_delete, namespace, rest_ops, key)
			}
		}

	} else {
		// Everything is a POST call
		for _, pool := range pool_nodes {
			_, rest_ops = rest.PkiProfileCU(pool.PkiProfile, nil, namespace, rest_ops, key)

			utils.AviLog.Debugf("key: %s, msg: pool cache does not exist %s, operation: POST", key, pool.Name)
			restOp := rest.AviPoolBuild(pool, nil, key)
			if restOp != nil {
				rest_ops = append(rest_ops, restOp)
			}
		}

	}
	utils.AviLog.Debugf("key: %s, msg: the POOLS rest_op is %s", key, utils.Stringify(rest_ops))
	utils.AviLog.Debugf("key: %s, msg: the POOLs to be deleted are: %s", key, cache_pool_nodes)
	return cache_pool_nodes, rest_ops
}

func (rest *RestOperations) SNINodeDelete(del_sni avicache.NamespaceName, namespace string, rest_ops []*utils.RestOp, avimodel *nodes.AviObjectGraph, key string) bool {
	utils.AviLog.Infof("key: %s, msg: about to delete the SNI child %s", key, del_sni)
	sni_key := avicache.NamespaceName{Namespace: namespace, Name: del_sni.Name}
	sni_cache_obj := rest.getVsCacheObj(sni_key, key)
	if sni_cache_obj != nil {
		utils.AviLog.Debugf("key: %s, msg: SNI object before delete %s", key, utils.Stringify(sni_cache_obj))
		// Verify that this object has all the related objects, if not do a manual refresh before delete.
		if len(sni_cache_obj.HTTPKeyCollection) < 1 || len(sni_cache_obj.PGKeyCollection) < 1 || len(sni_cache_obj.PoolKeyCollection) < 1 {
			// Some relationships are missing, do a manual refresh of the VS cache.
			aviObjCache := avicache.SharedAviObjCache()
			shardSize := lib.GetshardSize()
			if shardSize == 0 {
				// Dedicated VS case
				shardSize = 8
			}
			if shardSize != 0 {
				bkt := utils.Bkt(key, shardSize)
				utils.AviLog.Warnf("key: %s, msg: corrupted sni cache found, retrying in bkt: %v", key, bkt)
				aviRestPoolClient := avicache.SharedAVIClients(del_sni.Namespace)
				if len(aviRestPoolClient.AviClient) > 0 {
					aviclient := aviRestPoolClient.AviClient[bkt]
					aviObjCache.AviObjOneVSCachePopulate(aviclient, utils.CloudName, del_sni.Name, del_sni.Namespace)
					vsObjMeta, ok := rest.cache.VsCacheMeta.AviCacheGet(sni_key)
					if !ok {
						// Object deleted
						utils.AviLog.Warnf("key: %s, msg: SNI object already deleted")
						return true
					}
					vsCopy, done := vsObjMeta.(*avicache.AviVsCache).GetVSCopy()
					if done {
						rest.cache.VsCacheMeta.AviCacheAdd(sni_key, vsCopy)
					}
				}
				// Retry
				sni_cache_obj = rest.getVsCacheObj(sni_key, key)
			}
		}
		return rest.deleteSniVs(sni_key, sni_cache_obj, avimodel, namespace, key)
	}
	return true

}

func (rest *RestOperations) SNINodeCU(sni_node *nodes.AviVsNode, vs_cache_obj *avicache.AviVsCache, namespace string, cache_sni_nodes []avicache.NamespaceName, rest_ops []*utils.RestOp, key string) ([]avicache.NamespaceName, []*utils.RestOp) {
	var sni_pools_to_delete []avicache.NamespaceName
	var sni_pgs_to_delete []avicache.NamespaceName
	var http_policies_to_delete []avicache.NamespaceName
	var sslkey_cert_delete []avicache.NamespaceName
	var string_groups_to_delete []avicache.NamespaceName
	if vs_cache_obj != nil {
		sni_key := avicache.NamespaceName{Namespace: namespace, Name: sni_node.Name}
		// Search the VS cache and obtain the UUID of this VS. Then see if this UUID is part of the SNIChildCollection or not.
		found := utils.HasElem(cache_sni_nodes, sni_key)
		utils.AviLog.Debugf("key: %s, msg: processing node key: %v", key, sni_key)
		if found && cache_sni_nodes != nil {
			cache_sni_nodes = avicache.RemoveNamespaceName(cache_sni_nodes, sni_key)
			utils.AviLog.Debugf("key: %s, msg: the cache sni nodes are: %v", key, cache_sni_nodes)
			sni_cache_obj := rest.getVsCacheObj(sni_key, key)
			if sni_cache_obj != nil {
				// CAcerts have to be created first, as they are referred by the keycerts
				sslkey_cert_delete, rest_ops = rest.CACertCU(sni_node.CACertRefs, sni_cache_obj.SSLKeyCertCollection, namespace, rest_ops, key)
				// SSLKeyCertCollection which did not match cacerts are present in the list sslkey_cert_delete,
				// which shuld be the new SSLKeyCertCollection
				sslkey_cert_delete, rest_ops = rest.SSLKeyCertCU(sni_node.SSLKeyCertRefs, sslkey_cert_delete, namespace, rest_ops, key)
				sni_pools_to_delete, rest_ops = rest.PoolCU(sni_node.PoolRefs, sni_cache_obj, namespace, rest_ops, key)
				sni_pgs_to_delete, rest_ops = rest.PoolGroupCU(sni_node.PoolGroupRefs, sni_cache_obj, namespace, rest_ops, key)
				string_groups_to_delete, rest_ops = rest.StringGroupVsCU(sni_node.StringGroupRefs, sni_cache_obj, namespace, rest_ops, key)
				http_policies_to_delete, rest_ops = rest.HTTPPolicyCU(sni_node.HttpPolicyRefs, sni_cache_obj, namespace, rest_ops, key)
				// The checksums are different, so it should be a PUT call.
				if sni_cache_obj.CloudConfigCksum != strconv.Itoa(int(sni_node.GetCheckSum())) {
					restOp := rest.AviVsBuild(sni_node, utils.RestPut, sni_cache_obj, key)
					if restOp != nil {
						rest_ops = append(rest_ops, restOp...)
					}
					utils.AviLog.Infof("key: %s, msg: the checksums are different for sni child %s, operation: PUT", key, sni_node.Name)

				}
			}
		} else {
			utils.AviLog.Debugf("key: %s, msg: sni child %s not found in cache, operation: POST", key, sni_node.Name)
			_, rest_ops = rest.CACertCU(sni_node.CACertRefs, []avicache.NamespaceName{}, namespace, rest_ops, key)
			_, rest_ops = rest.SSLKeyCertCU(sni_node.SSLKeyCertRefs, nil, namespace, rest_ops, key)
			_, rest_ops = rest.PoolCU(sni_node.PoolRefs, nil, namespace, rest_ops, key)
			_, rest_ops = rest.PoolGroupCU(sni_node.PoolGroupRefs, nil, namespace, rest_ops, key)
			_, rest_ops = rest.StringGroupVsCU(sni_node.StringGroupRefs, nil, namespace, rest_ops, key)
			_, rest_ops = rest.HTTPPolicyCU(sni_node.HttpPolicyRefs, nil, namespace, rest_ops, key)

			// Not found - it should be a POST call.
			restOp := rest.AviVsBuild(sni_node, utils.RestPost, nil, key)
			if restOp != nil {
				rest_ops = append(rest_ops, restOp...)
			}
		}
		rest_ops = rest.SSLKeyCertDelete(sslkey_cert_delete, namespace, rest_ops, key)
		rest_ops = rest.HTTPPolicyDelete(http_policies_to_delete, namespace, rest_ops, key)
		rest_ops = rest.StringGroupDelete(string_groups_to_delete, namespace, rest_ops, key)
		rest_ops = rest.PoolGroupDelete(sni_pgs_to_delete, namespace, rest_ops, key)
		rest_ops = rest.PoolDelete(sni_pools_to_delete, namespace, rest_ops, key)
		utils.AviLog.Debugf("key: %s, msg: the SNI VSes to be deleted are: %s", key, cache_sni_nodes)
	} else {
		utils.AviLog.Debugf("key: %s, msg: sni child %s not found in cache and SNI parent also does not exist in cache", key, sni_node.Name)
		_, rest_ops = rest.CACertCU(sni_node.CACertRefs, []avicache.NamespaceName{}, namespace, rest_ops, key)
		_, rest_ops = rest.SSLKeyCertCU(sni_node.SSLKeyCertRefs, nil, namespace, rest_ops, key)
		_, rest_ops = rest.PoolCU(sni_node.PoolRefs, nil, namespace, rest_ops, key)
		_, rest_ops = rest.PoolGroupCU(sni_node.PoolGroupRefs, nil, namespace, rest_ops, key)
		_, rest_ops = rest.StringGroupVsCU(sni_node.StringGroupRefs, nil, namespace, rest_ops, key)
		_, rest_ops = rest.HTTPPolicyCU(sni_node.HttpPolicyRefs, nil, namespace, rest_ops, key)

		// Not found - it should be a POST call.
		restOp := rest.AviVsBuild(sni_node, utils.RestPost, nil, key)
		if restOp != nil {
			rest_ops = append(rest_ops, restOp...)
		}
	}
	return cache_sni_nodes, rest_ops
}

func (rest *RestOperations) PoolGroupCU(pg_nodes []*nodes.AviPoolGroupNode, vs_cache_obj *avicache.AviVsCache, namespace string, rest_ops []*utils.RestOp, key string) ([]avicache.NamespaceName, []*utils.RestOp) {
	var cache_pg_nodes []avicache.NamespaceName
	if vs_cache_obj != nil {
		cache_pg_nodes = make([]avicache.NamespaceName, len(vs_cache_obj.PGKeyCollection))
		copy(cache_pg_nodes, vs_cache_obj.PGKeyCollection)
		utils.AviLog.Debugf("key: %s, msg: cached poolgroups before CU :%v", key, cache_pg_nodes)
		// Default is POST

		for _, pg := range pg_nodes {
			pg_key := avicache.NamespaceName{Namespace: namespace, Name: pg.Name}
			found := utils.HasElem(cache_pg_nodes, pg_key)
			if found {
				cache_pg_nodes = avicache.RemoveNamespaceName(cache_pg_nodes, pg_key)
				pg_cache, ok := rest.cache.PgCache.AviCacheGet(pg_key)
				if ok {
					pg_cache_obj, _ := pg_cache.(*avicache.AviPGCache)
					// Cache found. Let's compare the checksums
					if pg_cache_obj.CloudConfigCksum == strconv.Itoa(int(pg.GetCheckSum())) {
						utils.AviLog.Debugf("key: %s, msg: the checksums are same for PG %s, not doing anything", key, pg_cache_obj.Name)
					} else {
						// The checksums are different, so it should be a PUT call.
						restOp := rest.AviPoolGroupBuild(pg, pg_cache_obj, key)
						if restOp != nil {
							rest_ops = append(rest_ops, restOp)
						}
					}
				}
			} else {
				// Not found - it should be a POST call.
				restOp := rest.AviPoolGroupBuild(pg, nil, key)
				if restOp != nil {
					rest_ops = append(rest_ops, restOp)
				}
			}

		}

	} else {
		// Everything is a POST call
		for _, pg := range pg_nodes {
			restOp := rest.AviPoolGroupBuild(pg, nil, key)
			if restOp != nil {
				rest_ops = append(rest_ops, restOp)
			}
		}

	}
	utils.AviLog.Debugf("key: %s, msg: the PGs rest_op is %s", key, utils.Stringify(rest_ops))
	utils.AviLog.Debugf("key: %s, msg: the PGs to be deleted are: %s", key, cache_pg_nodes)
	return cache_pg_nodes, rest_ops
}

func (rest *RestOperations) DatascriptCU(ds_nodes []*nodes.AviHTTPDataScriptNode, vs_cache_obj *avicache.AviVsCache, namespace string, rest_ops []*utils.RestOp, key string) ([]avicache.NamespaceName, []*utils.RestOp) {
	var cache_ds_nodes []avicache.NamespaceName

	if vs_cache_obj != nil {
		cache_ds_nodes = make([]avicache.NamespaceName, len(vs_cache_obj.DSKeyCollection))
		copy(cache_ds_nodes, vs_cache_obj.DSKeyCollection)

		// Default is POST

		for _, ds := range ds_nodes {
			// check in the ds cache to see if this ds exists in AVI
			ds_key := avicache.NamespaceName{Namespace: namespace, Name: ds.Name}
			found := utils.HasElem(cache_ds_nodes, ds_key)
			if found {
				cache_ds_nodes = avicache.RemoveNamespaceName(cache_ds_nodes, ds_key)
				ds_cache, ok := rest.cache.DSCache.AviCacheGet(ds_key)
				if !ok {
					// If the DS Is not found - let's do a POST call.
					restOp := rest.AviDSBuild(ds, nil, key)
					if restOp != nil {
						rest_ops = append(rest_ops, restOp)
					}
				} else {
					dsCacheObj := ds_cache.(*avicache.AviDSCache)
					if dsCacheObj.CloudConfigCksum != ds.GetCheckSum() {
						utils.AviLog.Debugf("key: %s, msg: datascript checksum changed, updating - %s", key, ds.Name)
						restOp := rest.AviDSBuild(ds, dsCacheObj, key)
						if restOp != nil {
							rest_ops = append(rest_ops, restOp)
						}
					}
				}
			} else {
				// If the DS Is not found - let's do a POST call.
				for _, ds := range ds_nodes {
					restOp := rest.AviDSBuild(ds, nil, key)
					if restOp != nil {
						rest_ops = append(rest_ops, restOp)
					}
				}
			}
		}

	} else {
		// Everything is a POST call
		for _, ds := range ds_nodes {
			restOp := rest.AviDSBuild(ds, nil, key)
			if restOp != nil {
				rest_ops = append(rest_ops, restOp)
			}
		}

	}
	utils.AviLog.Debugf("key: %s, msg: the DS rest_op is %s", key, utils.Stringify(rest_ops))
	utils.AviLog.Debugf("key: %s, msg: the DS to be deleted are: %s", key, cache_ds_nodes)
	return cache_ds_nodes, rest_ops
}

func (rest *RestOperations) VSVipCU(vsvip_nodes []*nodes.AviVSVIPNode, vs_cache_obj *avicache.AviVsCache, namespace string, rest_ops []*utils.RestOp, key string) ([]avicache.NamespaceName, []*utils.RestOp, error) {
	var cache_vsvip_nodes []avicache.NamespaceName
	if vs_cache_obj != nil {
		cache_vsvip_nodes = make([]avicache.NamespaceName, len(vs_cache_obj.VSVipKeyCollection))
		copy(cache_vsvip_nodes, vs_cache_obj.VSVipKeyCollection)
		// Default is POST

		for _, vsvip := range vsvip_nodes {
			vsvip_key := avicache.NamespaceName{Namespace: namespace, Name: vsvip.Name}
			found := utils.HasElem(cache_vsvip_nodes, vsvip_key)
			if found {
				cache_vsvip_nodes = avicache.RemoveNamespaceName(cache_vsvip_nodes, vsvip_key)
				vsvip_cache, ok := rest.cache.VSVIPCache.AviCacheGet(vsvip_key)
				if ok {
					vsvip_cache_obj, _ := vsvip_cache.(*avicache.AviVSVIPCache)
					sort.Strings(vsvip_cache_obj.FQDNs)
					// Cache found. Let's compare the checksums
					utils.AviLog.Debugf("key: %s, msg: the model FQDNs: %s, cache_FQDNs: %s", key, vsvip.FQDNs, vsvip_cache_obj.FQDNs)

					if vsvip_cache_obj.CloudConfigCksum == strconv.Itoa(int(vsvip.GetCheckSum())) {
						utils.AviLog.Debugf("key: %s, msg: the checksums are same for VSVIP %s, not doing anything", key, vsvip_cache_obj.Name)
					} else {
						// The checksums are different, so it should be a PUT call.
						restOp, err := rest.AviVsVipBuild(vsvip, vs_cache_obj, vsvip_cache_obj, key)
						if err == nil && restOp != nil {
							rest_ops = append(rest_ops, restOp)
						} else {
							return cache_vsvip_nodes, rest_ops, err
						}
					}
				}
			} else {
				// Not found - it should be a POST call.
				restOp, err := rest.AviVsVipBuild(vsvip, vs_cache_obj, nil, key)
				if err == nil && restOp != nil {
					rest_ops = append(rest_ops, restOp)
				} else {
					return cache_vsvip_nodes, rest_ops, err
				}
			}

		}

	} else {
		// Everything is a POST call
		for _, vsvip := range vsvip_nodes {
			restOp, err := rest.AviVsVipBuild(vsvip, vs_cache_obj, nil, key)
			if err == nil && restOp != nil {
				rest_ops = append(rest_ops, restOp)
			} else {
				return cache_vsvip_nodes, rest_ops, err
			}
		}

	}
	utils.AviLog.Debugf("key: %s, msg: the vsvip rest_op is %s", key, utils.Stringify(rest_ops))
	utils.AviLog.Debugf("key: %s, msg: the vsvip to be deleted are: %s", key, cache_vsvip_nodes)
	return cache_vsvip_nodes, rest_ops, nil
}

func (rest *RestOperations) HTTPPolicyCU(http_nodes []*nodes.AviHttpPolicySetNode, vs_cache_obj *avicache.AviVsCache, namespace string, rest_ops []*utils.RestOp, key string) ([]avicache.NamespaceName, []*utils.RestOp) {
	var cache_http_nodes []avicache.NamespaceName
	// Default is POST
	if vs_cache_obj != nil {
		cache_http_nodes = make([]avicache.NamespaceName, len(vs_cache_obj.HTTPKeyCollection))
		copy(cache_http_nodes, vs_cache_obj.HTTPKeyCollection)
		for _, http := range http_nodes {
			http_key := avicache.NamespaceName{Namespace: namespace, Name: http.Name}
			found := utils.HasElem(cache_http_nodes, http_key)
			if found {
				http_cache, ok := rest.cache.HTTPPolicyCache.AviCacheGet(http_key)
				if ok {
					cache_http_nodes = avicache.RemoveNamespaceName(cache_http_nodes, http_key)
					http_cache_obj, _ := http_cache.(*avicache.AviHTTPPolicyCache)
					// Cache found. Let's compare the checksums
					if http_cache_obj.CloudConfigCksum == strconv.Itoa(int(http.GetCheckSum())) {
						utils.AviLog.Debugf("The checksums are same for HTTP cache obj %s, not doing anything", http_cache_obj.Name)
					} else {
						// The checksums are different, so it should be a PUT call.
						restOp := rest.AviHttpPSBuild(http, http_cache_obj, key)
						if restOp != nil {
							rest_ops = append(rest_ops, restOp)
						}
					}
				}
			} else {
				// Not found - it should be a POST call.
				restOp := rest.AviHttpPSBuild(http, nil, key)
				if restOp != nil {
					rest_ops = append(rest_ops, restOp)
				}
			}

		}
	} else {
		// Everything is a POST call
		for _, http := range http_nodes {
			restOp := rest.AviHttpPSBuild(http, nil, key)
			if restOp != nil {
				rest_ops = append(rest_ops, restOp)
			}
		}

	}
	utils.AviLog.Debugf("key: %s, The HTTP Policies rest_op is %s", key, utils.Stringify(rest_ops))
	utils.AviLog.Debugf("key: %s, msg: the http policies to be deleted are: %s", key, cache_http_nodes)
	return cache_http_nodes, rest_ops
}

func (rest *RestOperations) L4PolicyCU(l4_nodes []*nodes.AviL4PolicyNode, vs_cache_obj *avicache.AviVsCache, namespace string, rest_ops []*utils.RestOp, key string) ([]avicache.NamespaceName, []*utils.RestOp) {
	var cache_l4_nodes []avicache.NamespaceName
	// Default is POST
	if vs_cache_obj != nil {
		cache_l4_nodes = make([]avicache.NamespaceName, len(vs_cache_obj.L4PolicyCollection))
		copy(cache_l4_nodes, vs_cache_obj.L4PolicyCollection)
		for _, l4 := range l4_nodes {
			l4_key := avicache.NamespaceName{Namespace: namespace, Name: l4.Name}
			found := utils.HasElem(cache_l4_nodes, l4_key)
			if found {
				l4_cache, ok := rest.cache.L4PolicyCache.AviCacheGet(l4_key)
				if ok {
					cache_l4_nodes = avicache.RemoveNamespaceName(cache_l4_nodes, l4_key)
					l4_cache_obj, _ := l4_cache.(*avicache.AviL4PolicyCache)
					// Cache found. Let's compare the checksums
					if l4_cache_obj.CloudConfigCksum == l4.GetCheckSum() {
						utils.AviLog.Debugf("The checksums are same for l4 cache obj %s, not doing anything", l4_cache_obj.Name)
					} else {
						// The checksums are different, so it should be a PUT call.
						restOp := rest.AviL4PSBuild(l4, l4_cache_obj, key)
						if restOp != nil {
							rest_ops = append(rest_ops, restOp)
						}
					}
				}
			} else {
				// Not found - it should be a POST call.
				restOp := rest.AviL4PSBuild(l4, nil, key)
				if restOp != nil {
					rest_ops = append(rest_ops, restOp)
				}
			}

		}
	} else {
		// Everything is a POST call
		for _, l4 := range l4_nodes {
			restOp := rest.AviL4PSBuild(l4, nil, key)
			if restOp != nil {
				rest_ops = append(rest_ops, restOp)
			}
		}

	}
	utils.AviLog.Debugf("The l4 Policies rest_op is %s", utils.Stringify(rest_ops))
	utils.AviLog.Debugf("key: %s, msg: the l4 policies to be deleted are: %s", key, cache_l4_nodes)
	return cache_l4_nodes, rest_ops
}

func (rest *RestOperations) HTTPPolicyDelete(https_to_delete []avicache.NamespaceName, namespace string, rest_ops []*utils.RestOp, key string) []*utils.RestOp {
	for _, del_http := range https_to_delete {
		// fetch trhe http policyset uuid from cache
		http_key := avicache.NamespaceName{Namespace: namespace, Name: del_http.Name}
		http_cache, ok := rest.cache.HTTPPolicyCache.AviCacheGet(http_key)
		if ok {
			http_cache_obj, _ := http_cache.(*avicache.AviHTTPPolicyCache)
			restOp := rest.AviHttpPolicyDel(http_cache_obj.Uuid, namespace, key)
			restOp.ObjName = del_http.Name
			rest_ops = append(rest_ops, restOp)
		}
	}
	return rest_ops
}

func (rest *RestOperations) CACertCU(caCertNodes []*nodes.AviTLSKeyCertNode, certKeys []avicache.NamespaceName, namespace string, rest_ops []*utils.RestOp, key string) ([]avicache.NamespaceName, []*utils.RestOp) {
	return rest.KeyCertCU(caCertNodes, certKeys, namespace, rest_ops, key)
}

func (rest *RestOperations) SSLKeyCertCU(sslkeyNodes []*nodes.AviTLSKeyCertNode, certKeys []avicache.NamespaceName, namespace string, rest_ops []*utils.RestOp, key string) ([]avicache.NamespaceName, []*utils.RestOp) {
	return rest.KeyCertCU(sslkeyNodes, certKeys, namespace, rest_ops, key)
}

func (rest *RestOperations) L4PolicyDelete(l4_to_delete []avicache.NamespaceName, namespace string, rest_ops []*utils.RestOp, key string) []*utils.RestOp {
	utils.AviLog.Infof("key: %s, msg: about to delete l4 policies %s", key, utils.Stringify(l4_to_delete))
	for _, del_l4 := range l4_to_delete {
		// fetch trhe http policyset uuid from cache
		l4_key := avicache.NamespaceName{Namespace: namespace, Name: del_l4.Name}
		l4_cache, ok := rest.cache.L4PolicyCache.AviCacheGet(l4_key)
		if ok {
			l4_cache_obj, _ := l4_cache.(*avicache.AviL4PolicyCache)
			restOp := rest.AviL4PolicyDel(l4_cache_obj.Uuid, namespace, key)
			restOp.ObjName = del_l4.Name
			rest_ops = append(rest_ops, restOp)
		}
	}
	return rest_ops
}

func (rest *RestOperations) KeyCertCU(sslkey_nodes []*nodes.AviTLSKeyCertNode, certKeys []avicache.NamespaceName, namespace string, rest_ops []*utils.RestOp, key string) ([]avicache.NamespaceName, []*utils.RestOp) {
	// Default is POST
	var cache_ssl_nodes []avicache.NamespaceName
	if len(certKeys) != 0 {
		cache_ssl_nodes = make([]avicache.NamespaceName, len(certKeys))
		copy(cache_ssl_nodes, certKeys)
		for _, ssl := range sslkey_nodes {
			ssl_key := avicache.NamespaceName{Namespace: namespace, Name: ssl.Name}
			found := utils.HasElem(cache_ssl_nodes, ssl_key)
			if found {
				ssl_cache, ok := rest.cache.SSLKeyCache.AviCacheGet(ssl_key)
				if ok {
					cache_ssl_nodes = avicache.RemoveNamespaceName(cache_ssl_nodes, ssl_key)
					ssl_cache_obj, _ := ssl_cache.(*avicache.AviSSLCache)
					if ssl_cache_obj.CloudConfigCksum == ssl.GetCheckSum() {
						utils.AviLog.Debugf("The checksums are same for SSL cache obj %s, not doing anything", ssl_cache_obj.Name)
					} else {
						// The checksums are different, so it should be a PUT call.
						restOp := rest.AviSSLBuild(ssl, ssl_cache_obj)
						if restOp != nil {
							rest_ops = append(rest_ops, restOp)
						}
					}
				}
			} else {
				// Not found - it should be a POST call.
				restOp := rest.AviSSLBuild(ssl, nil)
				if restOp != nil {
					rest_ops = append(rest_ops, restOp)
				}
			}

		}
	} else {
		// Everything is a POST call
		for _, ssl := range sslkey_nodes {
			restOp := rest.AviSSLBuild(ssl, nil)
			if restOp != nil {
				rest_ops = append(rest_ops, restOp)
			}
		}

	}
	return cache_ssl_nodes, rest_ops
}

func (rest *RestOperations) SSLKeyCertDelete(ssl_to_delete []avicache.NamespaceName, namespace string, rest_ops []*utils.RestOp, key string) []*utils.RestOp {
	utils.AviLog.Debugf("key: %s, msg: about to delete ssl keycert %s", key, utils.Stringify(ssl_to_delete))
	var noCARefRestOps []*utils.RestOp
	defaultRouteCertName := lib.GetTLSKeyCertNodeName("", "", lib.GetDefaultSecretForRoutes())
	defaultRouteAltCertName := lib.GetTLSKeyCertNodeName("", "", lib.GetDefaultSecretForRoutes()+"-alt")
	for _, del_ssl := range ssl_to_delete {
		// Skip SSL cert deletion if it maps to the Default Router Cert
		if del_ssl.Name == defaultRouteCertName || del_ssl.Name == defaultRouteAltCertName {
			if !strings.HasSuffix(key, lib.DummyVSForStaleData) {
				continue
			}
		}
		ssl_key := avicache.NamespaceName{Namespace: namespace, Name: del_ssl.Name}
		ssl_cache, ok := rest.cache.SSLKeyCache.AviCacheGet(ssl_key)
		if ok {
			// if deleteConfig is false and istio is enabled, do not delete istio sslkeycert
			if ssl_key.Name == lib.GetIstioWorkloadCertificateName() &&
				lib.IsIstioEnabled() && !lib.GetDeleteConfigMap() {
				continue
			}
			ssl_cache_obj, _ := ssl_cache.(*avicache.AviSSLCache)
			restOp := rest.AviSSLKeyCertDel(ssl_cache_obj.Uuid, namespace)
			restOp.ObjName = del_ssl.Name
			//Objects with a CA ref should be deleted first
			if !ssl_cache_obj.HasCARef {
				noCARefRestOps = append(noCARefRestOps, restOp)
			} else {
				rest_ops = append(rest_ops, restOp)
			}
		}
	}
	rest_ops = append(rest_ops, noCARefRestOps...)
	return rest_ops
}

func (rest *RestOperations) PkiProfileCU(pki_node *nodes.AviPkiProfileNode, pool_cache_obj *avicache.AviPoolCache, namespace string, rest_ops []*utils.RestOp, key string) ([]avicache.NamespaceName, []*utils.RestOp) {
	// Default is POST
	var cache_pki_nodes []avicache.NamespaceName
	if pool_cache_obj != nil {
		cache_pki_nodes = make([]avicache.NamespaceName, 1)
		copy(cache_pki_nodes, []avicache.NamespaceName{pool_cache_obj.PkiProfileCollection})

		if pki_node != nil {
			pki_key := avicache.NamespaceName{Namespace: namespace, Name: pki_node.Name}
			found := utils.HasElem(cache_pki_nodes, pki_key)
			if found {
				pki_cache, ok := rest.cache.PKIProfileCache.AviCacheGet(pki_key)
				if ok {
					cache_pki_nodes = avicache.RemoveNamespaceName(cache_pki_nodes, pki_key)
					pki_cache_obj, _ := pki_cache.(*avicache.AviPkiProfileCache)
					if pki_cache_obj.CloudConfigCksum == pki_node.GetCheckSum() {
						utils.AviLog.Debugf("The checksums are same for Pki cache obj %s, not doing anything", pki_cache_obj.Name)
					} else {
						// The checksums are different, so it should be a PUT call.
						restOp := rest.AviPkiProfileBuild(pki_node, pki_cache_obj)
						if restOp != nil {
							rest_ops = append(rest_ops, restOp)
						}
					}
				}
			} else {
				restOp := rest.AviPkiProfileBuild(pki_node, nil)
				if restOp != nil {
					rest_ops = append(rest_ops, restOp)
				}
			}
		}
	} else {
		if pki_node != nil {
			// Everything is a POST call
			restOp := rest.AviPkiProfileBuild(pki_node, nil)
			if restOp != nil {
				rest_ops = append(rest_ops, restOp)
			}
		}

	}

	return cache_pki_nodes, rest_ops
}

func (rest *RestOperations) PkiProfileDelete(pkiProfileDelete []avicache.NamespaceName, namespace string, rest_ops []*utils.RestOp, key string) []*utils.RestOp {
	utils.AviLog.Debugf("key: %s, msg: about to delete pki profile %s", key, utils.Stringify(pkiProfileDelete))
	for _, delPki := range pkiProfileDelete {
		pkiProfile := avicache.NamespaceName{Namespace: namespace, Name: delPki.Name}
		pkiCache, ok := rest.cache.PKIProfileCache.AviCacheGet(pkiProfile)
		if ok {
			// if deleteConfig is false and istio is enabled, do not delete istio pkiprofile
			if pkiProfile.Name == lib.GetIstioPKIProfileName() &&
				lib.IsIstioEnabled() && !lib.GetDeleteConfigMap() {

				continue
			}
			pkiCacheObj, _ := pkiCache.(*avicache.AviPkiProfileCache)
			restOp := rest.AviPkiProfileDel(pkiCacheObj.Uuid, namespace)
			restOp.ObjName = delPki.Name
			rest_ops = append(rest_ops, restOp)

		}
	}
	return rest_ops
}

func (rest *RestOperations) stringGroupCU(key, stringGroupName string, avimodel *nodes.AviObjectGraph) {
	var cache_sg_node avicache.NamespaceName
	var rest_ops []*utils.RestOp
	sg_node := avimodel.GetAviStringGroupNodeByName(stringGroupName)
	if sg_node != nil {
		// Default is POST
		// check in the sg cache to see if this exists in AVI
		sg_key := avicache.NamespaceName{Namespace: lib.GetTenant(), Name: *sg_node.Name}
		found := utils.HasElem(cache_sg_node, sg_key)
		if found {
			sg_cache, ok := rest.cache.StringGroupCache.AviCacheGet(sg_key)
			if !ok {
				// If the StringGroup Is not found - let's do a POST call.
				restOp := rest.AviStringGroupBuild(sg_node, nil, key)
				if restOp != nil {
					rest_ops = append(rest_ops, restOp)
				}
			} else {
				sgCacheObj := sg_cache.(*avicache.AviStringGroupCache)
				if sgCacheObj.CloudConfigCksum != sg_node.GetCheckSum() {
					utils.AviLog.Debugf("key: %s, msg: stringgroup checksum changed, updating - %s", key, sg_node.Name)
					restOp := rest.AviStringGroupBuild(sg_node, sgCacheObj, key)
					if restOp != nil {
						rest_ops = append(rest_ops, restOp)
					}
				}
			}
		} else {
			// If the stringgroup Is not found - let's do a POST call.
			restOp := rest.AviStringGroupBuild(sg_node, nil, key)
			if restOp != nil {
				rest_ops = append(rest_ops, restOp)
			}
		}
		utils.AviLog.Debugf("key: %s, msg: the StringGroup rest_op is %s", key, utils.Stringify(rest_ops))
		utils.AviLog.Debugf("key: %s, msg: Executing rest for stringgroup %s", key, stringGroupName)
		utils.AviLog.Debugf("key: %s, msg: restops %v", key, rest_ops)
		success, _ := rest.ExecuteRestAndPopulateCache(rest_ops, sg_key, avimodel, key, false)
		if success {
			utils.AviLog.Debugf("key: %s, msg: the StringGroup added successfully: %s", key, cache_sg_node)
		}
	}
}

func (rest *RestOperations) StringGroupVsCU(stringGroupNodes []*nodes.AviStringGroupNode, vs_cache_obj *avicache.AviVsCache, namespace string, rest_ops []*utils.RestOp, key string) ([]avicache.NamespaceName, []*utils.RestOp) {
	var cache_string_group_nodes []avicache.NamespaceName
	// Default is POST
	if vs_cache_obj != nil {
		cache_string_group_nodes = make([]avicache.NamespaceName, len(vs_cache_obj.StringGroupKeyCollection))
		copy(cache_string_group_nodes, vs_cache_obj.StringGroupKeyCollection)
		for _, stringGroup := range stringGroupNodes {
			stringGroupKey := avicache.NamespaceName{Namespace: namespace, Name: *stringGroup.Name}
			found := utils.HasElem(cache_string_group_nodes, stringGroupKey)
			if found {
				stringGroupCache, ok := rest.cache.StringGroupCache.AviCacheGet(stringGroupKey)
				if ok {
					cache_string_group_nodes = avicache.RemoveNamespaceName(cache_string_group_nodes, stringGroupKey)
					stringGroupCacheObj, _ := stringGroupCache.(*avicache.AviStringGroupCache)
					// Cache found. Let's compare the checksums
					if stringGroupCacheObj.CloudConfigCksum == stringGroup.GetCheckSum() {
						utils.AviLog.Debugf("The checksums are same for string group cache obj %s, not doing anything", stringGroupCacheObj.Name)
					} else {
						// The checksums are different, so it should be a PUT call.
						restOp := rest.AviStringGroupBuild(stringGroup, stringGroupCacheObj, key)
						if restOp != nil {
							rest_ops = append(rest_ops, restOp)
						}
					}
				}
			} else {
				// Not found - it should be a POST call.
				restOp := rest.AviStringGroupBuild(stringGroup, nil, key)
				if restOp != nil {
					rest_ops = append(rest_ops, restOp)
				}
			}

		}
	} else {
		// Everything is a POST call
		for _, stringGroup := range stringGroupNodes {
			restOp := rest.AviStringGroupBuild(stringGroup, nil, key)
			if restOp != nil {
				rest_ops = append(rest_ops, restOp)
			}
		}

	}
	utils.AviLog.Debugf("key: %s, The StringGroup rest_op is %s", key, utils.Stringify(rest_ops))
	utils.AviLog.Debugf("key: %s, msg: the string groups to be deleted are: %s", key, cache_string_group_nodes)
	return cache_string_group_nodes, rest_ops
}

func (rest *RestOperations) StringGroupDelete(sg_to_delete []avicache.NamespaceName, namespace string, rest_ops []*utils.RestOp, key string) []*utils.RestOp {
	utils.AviLog.Infof("key: %s, msg: about to delete the StringGroup %s", key, sg_to_delete)
	for _, del_sg := range sg_to_delete {
		// fetch trhe stringgroup uuid from cache
		sg_key := avicache.NamespaceName{Namespace: namespace, Name: del_sg.Name}
		sg_cache, ok := rest.cache.StringGroupCache.AviCacheGet(sg_key)
		if ok {
			sg_cache_obj, _ := sg_cache.(*avicache.AviStringGroupCache)
			restOp := rest.AviStringGroupDel(sg_cache_obj.Uuid, namespace, key)
			restOp.ObjName = del_sg.Name
			rest_ops = append(rest_ops, restOp)
		} else {
			utils.AviLog.Debugf("key: %s, msg: stringgroup not found in cache during delete %s", key, sg_to_delete)
		}
	}
	return rest_ops
}
