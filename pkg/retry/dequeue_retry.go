/*
 * [2013] - [2020] Avi Networks Incorporated
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
package retry

import (
	avicache "ako/pkg/cache"
	"ako/pkg/lib"
	"ako/pkg/nodes"
	"strings"

	"github.com/avinetworks/container-lib/utils"
)

func DequeueSlowRetry(vsKey string) {
	// Retrieve the Key and note the time.
	utils.AviLog.Info.Printf("Retrieved the key: %s", vsKey)
	// Fetch the cache for this VS key
	aviObjCache := avicache.SharedAviObjCache()
	// Fetch the VS UUID from the cache. If the VS UUID is not found, it's a POST call retry,
	// in which case cache refresh is not required.
	vsCacheKey := avicache.NamespaceName{Namespace: utils.ADMIN_NS, Name: vsKey}
	vsCache, ok := aviObjCache.VsCache.AviCacheGet(vsCacheKey)
	var deletedKeys []avicache.NamespaceName
	if ok {
		avi_rest_client_pool := avicache.SharedAVIClients()
		// Randomly pickup a client.
		if len(avi_rest_client_pool.AviClient) > 0 {
			vsCacheObj, found := vsCache.(*avicache.AviVsCache)
			if found {
				utils.AviLog.Info.Printf("Refreshing cache for: %s", vsCacheObj.Uuid)
				// Let's check if this VS also has a SNI Child - in which we will refresh that cache as well.
				err := aviObjCache.AviObjOneVSCachePopulate(avi_rest_client_pool.AviClient[0], utils.CloudName, vsCacheObj.Uuid)
				if err != nil && strings.Contains(err.Error(), lib.NOT_FOUND) {
					// Assume something really bad has happened, run a refresh on the entire cache.
					utils.AviLog.Warning.Printf("VS not found, something bad happened, refreshing the entire cache: %s", vsCacheObj.Uuid)
					deletedKeys, _ = aviObjCache.AviObjCachePopulate(avi_rest_client_pool.AviClient[0], utils.CtrlVersion, utils.CloudName)
					utils.AviLog.Info.Printf("Deleting cache for: %s, object not present in controller", vsCacheObj.Uuid)
					aviObjCache.VsCache.AviCacheDelete(vsCacheKey)
				}
			}
		}
	}
	// At this point, we can re-enqueue the key back to the rest layer.
	sharedQueue := utils.SharedWorkQueue().GetQueueByName(utils.GraphLayer)
	modelName := utils.ADMIN_NS + "/" + vsKey
	nodes.PublishKeyToRestLayer(modelName, vsKey, sharedQueue)
	// Let's publish the deleted keys as well
	for _, key := range deletedKeys {
		// Don't want to re-enqueue the same key again
		if key.Name != vsKey {
			utils.AviLog.Info.Printf("Found deleted keys in the cache during full sync, re-publishing them to the REST layer: :%s", utils.Stringify(key))
			modelName := utils.ADMIN_NS + "/" + key.Name
			nodes.PublishKeyToRestLayer(modelName, key.Name, sharedQueue)
		}
	}
}

func DequeueFastRetry(vsKey string) {
	// Identical to the slow retry for now, we can make them different as we test out more scenarios.
	// Retrieve the Key and note the time.
	utils.AviLog.Info.Printf("Retrieved the key for fast retry: %s", vsKey)
	// Fetch the cache for this VS key
	aviObjCache := avicache.SharedAviObjCache()
	// Fetch the VS UUID from the cache. If the VS UUID is not found, it's a POST call retry,
	// in which case cache refresh is not required.
	vsCacheKey := avicache.NamespaceName{Namespace: utils.ADMIN_NS, Name: vsKey}
	vsCache, ok := aviObjCache.VsCache.AviCacheGet(vsCacheKey)
	var deletedKeys []avicache.NamespaceName
	if ok {
		avi_rest_client_pool := avicache.SharedAVIClients()
		// Randomly pickup a client.
		if len(avi_rest_client_pool.AviClient) > 0 {
			vsCacheObj, found := vsCache.(*avicache.AviVsCache)
			if found {
				// If we are here, refresh the Pool/PG/DS/SSL cache
				aviObjCache.AviRefreshObjectCache(avi_rest_client_pool.AviClient[0], utils.CloudName)
				utils.AviLog.Info.Printf("Refreshing cache for: %s", vsCacheObj.Uuid)
				// Let's check if this VS also has a SNI Child - in which we will refresh that cache as well.
				err := aviObjCache.AviObjOneVSCachePopulate(avi_rest_client_pool.AviClient[0], utils.CloudName, vsCacheObj.Uuid)
				if err != nil && strings.Contains(err.Error(), lib.NOT_FOUND) {
					// Assume something really bad has happened, run a refresh on the entire cache.
					utils.AviLog.Warning.Printf("VS not found, something bad happened, refreshing the entire cache: %s", vsCacheObj.Uuid)
					deletedKeys, _ = aviObjCache.AviObjCachePopulate(avi_rest_client_pool.AviClient[0], utils.CtrlVersion, utils.CloudName)
					utils.AviLog.Info.Printf("Deleting cache for: %s, object not present in controller", vsCacheObj.Uuid)
					aviObjCache.VsCache.AviCacheDelete(vsCacheKey)
				}
			}
		}
	}
	// At this point, we can re-enqueue the key back to the rest layer.
	sharedQueue := utils.SharedWorkQueue().GetQueueByName(utils.GraphLayer)
	modelName := utils.ADMIN_NS + "/" + vsKey
	nodes.PublishKeyToRestLayer(modelName, vsKey, sharedQueue)
	// Let's publish the deleted keys as well
	for _, key := range deletedKeys {
		// Don't want to re-enqueue the same key again
		if key.Name != vsKey {
			utils.AviLog.Info.Printf("Found deleted keys in the cache, re-publishing them to the REST layer: :%s", utils.Stringify(key))
			modelName := utils.ADMIN_NS + "/" + key.Name
			nodes.PublishKeyToRestLayer(modelName, vsKey, sharedQueue)
		}
	}
}
