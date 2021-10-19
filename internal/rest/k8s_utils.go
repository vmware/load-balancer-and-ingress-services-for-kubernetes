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
	avicache "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/cache"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/status"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
)

// SyncObjectStatuses gets data from L3 cache and does a status update on the ingress objects
// based on the service metadata objects it finds in the cache
// This is executed once AKO is done with populating the L3 cache in reboot scenarios
func (rest *RestOperations) SyncObjectStatuses() {
	vsKeys := rest.cache.VsCacheMeta.AviGetAllKeys()
	utils.AviLog.Debugf("Ingress status sync for vsKeys %+v", utils.Stringify(vsKeys))

	var allIngressUpdateOptions []status.UpdateOptions
	var allServiceLBUpdateOptions []status.UpdateOptions
	var allGatewayUpdateOptions []status.UpdateOptions
	for _, vsKey := range vsKeys {
		if vsKey.Name == lib.DummyVSForStaleData {
			continue
		}
		vsCache, ok := rest.cache.VsCacheMeta.AviCacheGet(vsKey)
		if !ok {
			continue
		}

		vsCacheObj, found := vsCache.(*avicache.AviVsCache)
		if !found {
			continue
		}

		parentVsKey := vsCacheObj.ParentVSRef
		vsSvcMetadataObj := vsCacheObj.ServiceMetadataObj
		var IPAddrs []string
		if vsSvcMetadataObj.Gateway != "" {
			// gateway based VSes
			IPAddrs = append(IPAddrs, vsCacheObj.Vip)
			allGatewayUpdateOptions = append(allGatewayUpdateOptions,
				status.UpdateOptions{
					Vip:             IPAddrs,
					ServiceMetadata: vsSvcMetadataObj,
					Key:             lib.SyncStatusKey,
				})
		} else if parentVsKey != (avicache.NamespaceName{}) {
			// secure VSes handler
			parentVs, found := rest.cache.VsCacheMeta.AviCacheGet(parentVsKey)
			if !found {
				continue
			}

			parentVsObj, _ := parentVs.(*avicache.AviVsCache)
			if (vsSvcMetadataObj.IngressName != "" || len(vsSvcMetadataObj.NamespaceIngressName) > 0) && vsSvcMetadataObj.Namespace != "" && parentVsObj != nil {
				IPAddrs = append(IPAddrs, parentVsObj.Vip)
				allIngressUpdateOptions = append(allIngressUpdateOptions,
					status.UpdateOptions{
						Vip:                IPAddrs,
						ServiceMetadata:    vsSvcMetadataObj,
						Key:                lib.SyncStatusKey,
						VirtualServiceUUID: vsCacheObj.Uuid,
					})
			}
		} else if len(vsSvcMetadataObj.NamespaceServiceName) > 0 {
			// serviceLB
			if vsCacheObj.Fip != "" {
				IPAddrs = append(IPAddrs, vsCacheObj.Fip)
			} else {
				IPAddrs = append(IPAddrs, vsCacheObj.Vip)
			}
			allServiceLBUpdateOptions = append(allServiceLBUpdateOptions,

				status.UpdateOptions{
					Vip:                IPAddrs,
					ServiceMetadata:    vsSvcMetadataObj,
					Key:                lib.SyncStatusKey,
					VirtualServiceUUID: vsCacheObj.Uuid,
				})
		} else {
			// insecure VSes handler
			for _, poolKey := range vsCacheObj.PoolKeyCollection {
				poolCache, ok := rest.cache.PoolCache.AviCacheGet(poolKey)
				if !ok {
					continue
				}

				poolCacheObj, found := poolCache.(*avicache.AviPoolCache)
				if !found {
					continue
				}

				// insecure pools
				if poolCacheObj.ServiceMetadataObj.Namespace != "" {
					if vsCacheObj.Fip != "" {
						IPAddrs = append(IPAddrs, vsCacheObj.Fip)
					} else {
						IPAddrs = []string{vsCacheObj.Vip}
					}
					allIngressUpdateOptions = append(allIngressUpdateOptions,
						status.UpdateOptions{
							Vip:                IPAddrs,
							ServiceMetadata:    poolCacheObj.ServiceMetadataObj,
							Key:                lib.SyncStatusKey,
							VirtualServiceUUID: vsCacheObj.Uuid,
						})
				}
			}
		}
	}

	if lib.GetAdvancedL4() {
		status.UpdateGatewayStatusAddress(allGatewayUpdateOptions, true)
		status.UpdateL4LBStatus(allServiceLBUpdateOptions, true)
	} else {
		if lib.UseServicesAPI() {
			status.UpdateSvcApiGatewayStatusAddress(allGatewayUpdateOptions, true)
			status.UpdateL4LBStatus(allServiceLBUpdateOptions, true)
		}
		status.UpdateRouteIngressStatus(allIngressUpdateOptions, true)
		if !lib.GetLayer7Only() {
			status.UpdateL4LBStatus(allServiceLBUpdateOptions, true)
		}
	}
	utils.AviLog.Infof("Status syncing completed")
}
