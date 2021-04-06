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
		if vsSvcMetadataObj.Gateway != "" {
			// gateway based VSes
			allGatewayUpdateOptions = append(allGatewayUpdateOptions,
				status.UpdateOptions{
					Vip:             vsCacheObj.Vip,
					ServiceMetadata: vsSvcMetadataObj,
					Key:             "syncstatus",
				})
		} else if parentVsKey != (avicache.NamespaceName{}) {
			// secure VSes handler
			parentVs, found := rest.cache.VsCacheMeta.AviCacheGet(parentVsKey)
			if !found {
				continue
			}

			parentVsObj, _ := parentVs.(*avicache.AviVsCache)
			if (vsSvcMetadataObj.IngressName != "" || len(vsSvcMetadataObj.NamespaceIngressName) > 0) && vsSvcMetadataObj.Namespace != "" && parentVsObj != nil {
				allIngressUpdateOptions = append(allIngressUpdateOptions,
					status.UpdateOptions{
						Vip:                parentVsObj.Vip,
						ServiceMetadata:    vsSvcMetadataObj,
						Key:                "syncstatus",
						VirtualServiceUUID: vsCacheObj.Uuid,
					})
			}
		} else if len(vsSvcMetadataObj.NamespaceServiceName) > 0 {
			// serviceLB
			allServiceLBUpdateOptions = append(allServiceLBUpdateOptions,
				status.UpdateOptions{
					Vip:                vsCacheObj.Vip,
					ServiceMetadata:    vsSvcMetadataObj,
					Key:                "syncstatus",
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
					allIngressUpdateOptions = append(allIngressUpdateOptions,
						status.UpdateOptions{
							Vip:                vsCacheObj.Vip,
							ServiceMetadata:    poolCacheObj.ServiceMetadataObj,
							Key:                "syncstatus",
							VirtualServiceUUID: vsCacheObj.Uuid,
						})
				}
			}
		}
	}

	if lib.GetAdvancedL4() {
		for i := range allGatewayUpdateOptions {
			statusOption := status.StatusOptions{
				ObjType: lib.Gateway,
				Op:      lib.UpdateStatus,
				Options: &allGatewayUpdateOptions[i],
			}
			status.PublishToStatusQueue(allGatewayUpdateOptions[i].ServiceMetadata.Gateway, statusOption)
		}
	} else {
		for i := range allIngressUpdateOptions {
			statusOption := status.StatusOptions{
				ObjType: utils.Ingress,
				Op:      lib.UpdateStatus,
				Options: &allIngressUpdateOptions[i],
				IsVSDel: true,
			}
			if utils.GetInformers().RouteInformer != nil {
				statusOption.ObjType = utils.OshiftRoute
			}
			status.PublishToStatusQueue(allIngressUpdateOptions[i].ServiceMetadata.IngressName, statusOption)
		}
		for i := range allServiceLBUpdateOptions {
			statusOption := status.StatusOptions{
				ObjType: utils.L4LBService,
				Op:      lib.UpdateStatus,
				Options: &allServiceLBUpdateOptions[i],
			}
			status.PublishToStatusQueue(allServiceLBUpdateOptions[i].ServiceMetadata.NamespaceServiceName[0], statusOption)
		}
	}
	utils.AviLog.Infof("Status syncing completed")
	return
}
