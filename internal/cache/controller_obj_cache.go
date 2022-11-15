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

package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	akov1alpha1 "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/apis/ako/v1alpha1"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/third_party/github.com/vmware/alb-sdk/go/clients"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/third_party/github.com/vmware/alb-sdk/go/session"

	"github.com/vmware/alb-sdk/go/models"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
)

type AviObjCache struct {
	PgCache            *AviCache
	DSCache            *AviCache
	PoolCache          *AviCache
	CloudKeyCache      *AviCache
	HTTPPolicyCache    *AviCache
	L4PolicyCache      *AviCache
	SSLKeyCache        *AviCache
	PKIProfileCache    *AviCache
	VSVIPCache         *AviCache
	VrfCache           *AviCache
	VsCacheMeta        *AviCache
	VsCacheLocal       *AviCache
	ClusterStatusCache *AviCache
}

func NewAviObjCache() *AviObjCache {
	c := AviObjCache{}
	c.VsCacheMeta = NewAviCache()
	c.VsCacheLocal = NewAviCache()
	c.PgCache = NewAviCache()
	c.DSCache = NewAviCache()
	c.PoolCache = NewAviCache()
	c.SSLKeyCache = NewAviCache()
	c.CloudKeyCache = NewAviCache()
	c.HTTPPolicyCache = NewAviCache()
	c.L4PolicyCache = NewAviCache()
	c.VSVIPCache = NewAviCache()
	c.VrfCache = NewAviCache()
	c.PKIProfileCache = NewAviCache()
	c.ClusterStatusCache = NewAviCache()
	return &c
}

var cacheInstance *AviObjCache
var cacheOnce sync.Once

func SharedAviObjCache() *AviObjCache {
	cacheOnce.Do(func() {
		cacheInstance = NewAviObjCache()
	})
	return cacheInstance
}

func (c *AviObjCache) AviRefreshObjectCache(client []*clients.AviClient, cloud string) {
	var wg sync.WaitGroup
	// We want to run 8 go routines which will simultanesouly fetch objects from the controller.
	wg.Add(5)
	go func() {
		defer wg.Done()
		c.PopulateSSLKeyToCache(client[4], cloud)
	}()
	go func() {
		defer wg.Done()
		c.PopulateVsVipDataToCache(client[7], cloud)
	}()
	c.PopulatePkiProfilesToCache(client[0])
	c.PopulatePoolsToCache(client[1], cloud)
	c.PopulatePgDataToCache(client[2], cloud)

	go func() {
		defer wg.Done()
		c.PopulateDSDataToCache(client[3], cloud)
	}()

	go func() {
		defer wg.Done()
		c.PopulateHttpPolicySetToCache(client[5], cloud)
	}()
	go func() {
		defer wg.Done()
		c.PopulateL4PolicySetToCache(client[6], cloud)
	}()

	wg.Wait()
	utils.AviLog.Infof("Finished syncing all objects except virtualservices")
}

func (c *AviObjCache) AviCacheRefresh(client *clients.AviClient, cloud string) {
	c.AviCloudPropertiesPopulate(client, cloud)
}

func (c *AviObjCache) AviObjCachePopulate(client []*clients.AviClient, version string, cloud string) ([]NamespaceName, []NamespaceName, error) {
	vsCacheCopy := []NamespaceName{}
	allVsKeys := []NamespaceName{}
	err := c.AviObjVrfCachePopulate(client[0], cloud)
	if err != nil {
		return vsCacheCopy, allVsKeys, err
	}
	// Populate the VS cache
	utils.AviLog.Infof("Refreshing all object cache")
	c.AviRefreshObjectCache(client, cloud)
	utils.AviLog.Infof("Finished Refreshing all object cache")
	vsCacheCopy = c.VsCacheMeta.AviCacheGetAllParentVSKeys()
	allVsKeys = c.VsCacheMeta.AviGetAllKeys()
	err = c.AviObjVSCachePopulate(client[0], cloud, &allVsKeys)
	if err != nil {
		return vsCacheCopy, allVsKeys, err
	}
	// Populate the SNI VS keys to their respective parents
	c.PopulateVsMetaCache()
	// Delete all the VS keys that are left in the copy.
	for _, key := range allVsKeys {
		utils.AviLog.Debugf("Removing vs key from cache: %s", key)
		// We want to synthesize these keys to layer 3.
		vsCacheCopy = RemoveNamespaceName(vsCacheCopy, key)
		c.VsCacheMeta.AviCacheDelete(key)
	}
	err = c.AviCloudPropertiesPopulate(client[0], cloud)
	if err != nil {
		return vsCacheCopy, allVsKeys, err
	}
	if lib.GetDeleteConfigMap() {
		allParentVsKeys := c.VsCacheMeta.AviCacheGetAllParentVSKeys()
		return vsCacheCopy, allParentVsKeys, err
	}
	//vsCacheCopy at this time, is left with only the deleted keys
	return vsCacheCopy, allVsKeys, nil
}

// TODO: Deperecate this function in future release.
// This function list EVH child VS to be deleted which contain namespace in its un-encoded name.
func (c *AviObjCache) listEVHChildrenToDelete(vs_cache_obj *AviVsCache, childUuids []string) ([]NamespaceName, []string) {
	var childNSNameToDelete []NamespaceName
	var childUuidToDelete []string
	for _, childUuid := range childUuids {
		childKey, childFound := c.VsCacheLocal.AviCacheGetKeyByUuid(childUuid)
		if childFound {
			childVSKey := childKey.(NamespaceName)
			childObj, _ := c.VsCacheLocal.AviCacheGet(childVSKey)
			child_cache_obj, vs_found := childObj.(*AviVsCache)
			if vs_found && !lib.IsNameEncoded(child_cache_obj.Name) {
				//In EVH: Encoding of object names done in 2nd release of EVH
				childNSNameToDelete = append(childNSNameToDelete, childVSKey)
				vs_cache_obj.RemoveFromSNIChildCollection(childUuid)
				childUuidToDelete = append(childUuidToDelete, childUuid)
			}
		}
	}
	return childNSNameToDelete, childUuidToDelete
}

func (c *AviObjCache) PopulateVsMetaCache() {
	// The vh_child_uuids field is used to populate the SNI children during cache population. However, due to the datastore to PG delay - that field may
	// not always be accurate. We would reduce the problem with accuracy by refreshing the SNI cache through reverse mapping sni's to parent
	// Go over the entire VS cache.
	parentVsKeys := c.VsCacheLocal.AviCacheGetAllParentVSKeys()

	isEVHEnabled := lib.IsEvhEnabled()
	var nsNameToDelete []NamespaceName
	var childUuidToDelete []string

	for _, pvsKey := range parentVsKeys {
		// For each parentVs get the SNI children
		sniChildUuids := c.VsCacheLocal.AviCacheGetAllChildVSForParent(pvsKey)
		// Fetch the parent VS cache and update the SNI child
		vsObj, parentFound := c.VsCacheLocal.AviCacheGet(pvsKey)
		if parentFound {
			// Parent cache is already populated, just append the SNI key
			vs_cache_obj, foundvs := vsObj.(*AviVsCache)
			if foundvs {
				vs_cache_obj.ReplaceSNIChildCollection(sniChildUuids)
				if isEVHEnabled {
					curChildNSNameToDelete, curChildUuidToDelete := c.listEVHChildrenToDelete(vs_cache_obj, sniChildUuids)
					childUuidToDelete = append(childUuidToDelete, curChildUuidToDelete...)
					nsNameToDelete = append(nsNameToDelete, curChildNSNameToDelete...)
				}

			}
		}
	}
	childNSNameToDelete := make(map[NamespaceName]bool, len(nsNameToDelete))
	for _, ns := range nsNameToDelete {
		childNSNameToDelete[ns] = true
	}
	// Now write lock and copy over all VsCacheMeta and copy the right cache from local
	allVsKeys := c.VsCacheLocal.AviGetAllKeys()
	for _, vsKey := range allVsKeys {
		deleteVS := childNSNameToDelete[vsKey]
		vsObj, vsFound := c.VsCacheLocal.AviCacheGet(vsKey)
		if vsFound {
			vs_cache_obj, foundvs := vsObj.(*AviVsCache)
			if foundvs {
				if !deleteVS {
					c.MarkReference(vs_cache_obj)
				}
				vsCopy, done := vs_cache_obj.GetVSCopy()
				if done {
					c.VsCacheMeta.AviCacheAdd(vsKey, vsCopy)
					c.VsCacheLocal.AviCacheDelete(vsKey)
				}
			}
		}
	}
	c.DeleteUnmarked(childUuidToDelete)
}

// MarkReference : check objects referred by a VS and mark they they have reference
// so that they are not deleted during clean up stage
func (c *AviObjCache) MarkReference(vsCacheObj *AviVsCache) {
	for _, objKey := range vsCacheObj.DSKeyCollection {
		if intf, found := c.DSCache.AviCacheGet(objKey); found {
			if obj, ok := intf.(*AviDSCache); ok {
				obj.HasReference = true
			}
		}
	}

	for _, objKey := range vsCacheObj.HTTPKeyCollection {
		if intf, found := c.HTTPPolicyCache.AviCacheGet(objKey); found {
			if obj, ok := intf.(*AviHTTPPolicyCache); ok {
				obj.HasReference = true
			}
		}
	}

	for _, objKey := range vsCacheObj.L4PolicyCollection {
		if intf, found := c.L4PolicyCache.AviCacheGet(objKey); found {
			if obj, ok := intf.(*AviL4PolicyCache); ok {
				obj.HasReference = true
			}
		}
	}

	for _, objKey := range vsCacheObj.PGKeyCollection {
		if intf, found := c.PgCache.AviCacheGet(objKey); found {
			if obj, ok := intf.(*AviPGCache); ok {
				obj.HasReference = true
			}
		}
	}

	for _, objKey := range vsCacheObj.PoolKeyCollection {
		if intf, found := c.PoolCache.AviCacheGet(objKey); found {
			if obj, ok := intf.(*AviPoolCache); ok {
				obj.HasReference = true
			}
		}
	}

	for _, objKey := range vsCacheObj.SSLKeyCertCollection {
		if intf, found := c.SSLKeyCache.AviCacheGet(objKey); found {
			if obj, ok := intf.(*AviSSLCache); ok {
				obj.HasReference = true
			}
		}
	}

	for _, objKey := range vsCacheObj.VSVipKeyCollection {
		if intf, found := c.VSVIPCache.AviCacheGet(objKey); found {
			if obj, ok := intf.(*AviVSVIPCache); ok {
				obj.HasReference = true
			}
		}
	}
}

// DeleteUnmarked : Adds non referenced cached objects to a Dummy VS, which
// would be used later to delete these objects from AVI Controller
func (c *AviObjCache) DeleteUnmarked(childCollection []string) {

	var dsKeys, vsVipKeys, httpKeys, sslKeys []NamespaceName
	var pgKeys, poolKeys, l4Keys []NamespaceName
	for _, objkey := range c.DSCache.AviGetAllKeys() {
		intf, _ := c.DSCache.AviCacheGet(objkey)
		if obj, ok := intf.(*AviDSCache); ok {
			if obj.HasReference == false {
				utils.AviLog.Infof("Reference Not found for datascript: %s", objkey)
				dsKeys = append(dsKeys, objkey)
			}
		}
	}

	for _, objkey := range c.HTTPPolicyCache.AviGetAllKeys() {
		intf, _ := c.HTTPPolicyCache.AviCacheGet(objkey)
		if obj, ok := intf.(*AviHTTPPolicyCache); ok {
			if obj.HasReference == false {
				utils.AviLog.Infof("Reference Not found for http policy: %s", objkey)
				httpKeys = append(httpKeys, objkey)
			}
		}
	}

	for _, objkey := range c.L4PolicyCache.AviGetAllKeys() {
		intf, _ := c.L4PolicyCache.AviCacheGet(objkey)
		if obj, ok := intf.(*AviL4PolicyCache); ok {
			if obj.HasReference == false {
				utils.AviLog.Infof("Reference Not found for l4 policy: %s", objkey)
				l4Keys = append(l4Keys, objkey)
			}
		}
	}

	for _, objkey := range c.PgCache.AviGetAllKeys() {
		intf, _ := c.PgCache.AviCacheGet(objkey)
		if obj, ok := intf.(*AviPGCache); ok {
			if obj.HasReference == false {
				utils.AviLog.Infof("Reference Not found for poolgroup: %s", objkey)
				pgKeys = append(pgKeys, objkey)
			}
		}

	}

	for _, objkey := range c.PoolCache.AviGetAllKeys() {
		intf, _ := c.PoolCache.AviCacheGet(objkey)
		if obj, ok := intf.(*AviPoolCache); ok {
			if obj.HasReference == false {
				utils.AviLog.Infof("Reference Not found for pool: %s", objkey)
				poolKeys = append(poolKeys, objkey)
			}
		}
	}

	for _, objkey := range c.SSLKeyCache.AviGetAllKeys() {
		intf, _ := c.SSLKeyCache.AviCacheGet(objkey)
		if obj, ok := intf.(*AviSSLCache); ok {
			if obj.HasReference == false && obj.Name != lib.GetIstioWorkloadCertificateName() {
				utils.AviLog.Infof("Reference Not found for ssl key: %s", objkey)
				sslKeys = append(sslKeys, objkey)
			}
		}
	}

	for _, objkey := range c.VSVIPCache.AviGetAllKeys() {
		intf, _ := c.VSVIPCache.AviCacheGet(objkey)
		if obj, ok := intf.(*AviVSVIPCache); ok {
			if lib.IsShardVS(obj.Name) {
				utils.AviLog.Infof("Retaining the vsvip: %s", obj.Name)
				continue
			}
			if obj.HasReference == false {
				utils.AviLog.Infof("Reference Not found for vsvip: %s", objkey)
				vsVipKeys = append(vsVipKeys, objkey)
			}
		}
	}

	// Only add this if we have stale data
	vsMetaObj := AviVsCache{
		Name:                 lib.DummyVSForStaleData,
		VSVipKeyCollection:   vsVipKeys,
		HTTPKeyCollection:    httpKeys,
		DSKeyCollection:      dsKeys,
		SSLKeyCertCollection: sslKeys,
		PGKeyCollection:      pgKeys,
		PoolKeyCollection:    poolKeys,
		L4PolicyCollection:   l4Keys,
		SNIChildCollection:   childCollection,
	}
	vsKey := NamespaceName{
		Namespace: lib.GetTenant(),
		Name:      lib.DummyVSForStaleData,
	}
	utils.AviLog.Infof("Dummy VS for stale objects Deletion %s", utils.Stringify(&vsMetaObj))
	c.VsCacheMeta.AviCacheAdd(vsKey, &vsMetaObj)

}

func (c *AviObjCache) AviPopulateAllPGs(client *clients.AviClient, cloud string, pgData *[]AviPGCache, overrideUri ...NextPage) (*[]AviPGCache, int, error) {
	var uri string
	akoUser := lib.AKOUser

	if len(overrideUri) == 1 {
		uri = overrideUri[0].Next_uri
	} else {
		uri = "/api/poolgroup/?" + "include_name=true&cloud_ref.name=" + cloud + "&created_by=" + akoUser + "&page_size=100"
	}

	result, err := lib.AviGetCollectionRaw(client, uri)
	if err != nil {
		utils.AviLog.Warnf("Get uri %v returned err for pg %v", uri, err)
		return nil, 0, err
	}
	elems := make([]json.RawMessage, result.Count)
	err = json.Unmarshal(result.Results, &elems)
	if err != nil {
		utils.AviLog.Warnf("Failed to unmarshal pg data, err: %v", err)
		return nil, 0, err
	}
	for i := 0; i < len(elems); i++ {
		pg := models.PoolGroup{}
		err = json.Unmarshal(elems[i], &pg)
		if err != nil {
			utils.AviLog.Warnf("Failed to unmarshal pg data, err: %v", err)
			continue
		}

		if pg.Name == nil || pg.UUID == nil || pg.CloudConfigCksum == nil {
			utils.AviLog.Warnf("Incomplete pg data unmarshalled, %s", utils.Stringify(pg))
			continue
		}

		var pools []string
		for _, member := range pg.Members {
			// Parse each pool and populate inside pools.
			// Find out the uuid of the pool and then corresponding name
			poolUuid := ExtractUuid(*member.PoolRef, "pool-.*.#")
			// Search the poolName using this Uuid in the poolcache.
			poolName, found := c.PoolCache.AviCacheGetNameByUuid(poolUuid)
			if found {
				pools = append(pools, poolName.(string))
			}
		}
		pgCacheObj := AviPGCache{
			Name:             *pg.Name,
			Uuid:             *pg.UUID,
			CloudConfigCksum: *pg.CloudConfigCksum,
			LastModified:     *pg.LastModified,
			Members:          pools,
		}
		*pgData = append(*pgData, pgCacheObj)
	}
	if result.Next != "" {
		// It has a next page, let's recursively call the same method.
		next_uri := strings.Split(result.Next, "/api/poolgroup")
		if len(next_uri) > 1 {
			overrideUri := "/api/poolgroup" + next_uri[1]
			nextPage := NextPage{Next_uri: overrideUri}
			_, _, err := c.AviPopulateAllPGs(client, cloud, pgData, nextPage)
			if err != nil {
				return nil, 0, err
			}
		}
	}
	return pgData, result.Count, nil
}

func (c *AviObjCache) PopulatePgDataToCache(client *clients.AviClient, cloud string) {
	var pgData []AviPGCache
	c.AviPopulateAllPGs(client, cloud, &pgData)

	// Get all the PG cache data and copy them.
	pgCacheData := c.PgCache.ShallowCopy()
	for i, pgCacheObj := range pgData {
		k := NamespaceName{Namespace: lib.GetTenant(), Name: pgCacheObj.Name}
		oldPGIntf, found := c.PgCache.AviCacheGet(k)
		if found {
			oldPGData, ok := oldPGIntf.(*AviPGCache)
			if ok {
				if oldPGData.InvalidData || oldPGData.LastModified != pgData[i].LastModified {
					pgData[i].InvalidData = true
					utils.AviLog.Warnf("Invalid cache data for pg: %s", k)
				}
			} else {
				utils.AviLog.Warnf("Wrong data type for pg: %s in cache", k)
			}
		}
		utils.AviLog.Debugf("Adding key to pg cache :%s value :%s", k, pgCacheObj.Uuid)
		// Add the actual pg cache data
		// Only replace this data if the lastmodifed field varies
		c.PgCache.AviCacheAdd(k, &pgData[i])
		delete(pgCacheData, k)
	}
	// The data that is left in pgCacheData should be explicitly removed
	for key := range pgCacheData {
		utils.AviLog.Debugf("Deleting key from pg cache :%s", key)
		c.PgCache.AviCacheDelete(key)
	}
}

func (c *AviObjCache) AviPopulateAllPkiPRofiles(client *clients.AviClient, pkiData *[]AviPkiProfileCache, overrideUri ...NextPage) (*[]AviPkiProfileCache, int, error) {
	var uri string
	akoUser := lib.AKOUser

	if len(overrideUri) == 1 {
		uri = overrideUri[0].Next_uri
	} else {
		uri = "/api/pkiprofile/?" + "&include_name=true&" + "&created_by=" + akoUser + "&page_size=100"
	}

	result, err := lib.AviGetCollectionRaw(client, uri)
	if err != nil {
		utils.AviLog.Warnf("Get uri %v returned err for pool %v", uri, err)
		return nil, 0, err
	}
	elems := make([]json.RawMessage, result.Count)
	err = json.Unmarshal(result.Results, &elems)
	if err != nil {
		return nil, 0, err
	}
	for i := 0; i < len(elems); i++ {
		pki := models.PKIprofile{}
		err = json.Unmarshal(elems[i], &pki)
		if err != nil {
			utils.AviLog.Warnf("Failed to unmarshal pki data, err: %v", err)
			continue
		}

		if pki.Name == nil || pki.UUID == nil {
			utils.AviLog.Warnf("Incomplete pki data unmarshalled, %s", utils.Stringify(pki))
			continue
		}
		emptyIngestionMarkers := utils.AviObjectMarkers{}
		pkiCacheObj := AviPkiProfileCache{
			Name:             *pki.Name,
			Uuid:             *pki.UUID,
			Tenant:           lib.GetTenant(),
			CloudConfigCksum: lib.SSLKeyCertChecksum(*pki.Name, string(*pki.CaCerts[0].Certificate), "", emptyIngestionMarkers, pki.Markers, true),
		}
		*pkiData = append(*pkiData, pkiCacheObj)

	}
	if result.Next != "" {
		// It has a next page, let's recursively call the same method.
		next_uri := strings.Split(result.Next, "/api/pkiprofile")
		if len(next_uri) > 1 {
			overrideUri := "/api/pkiprofile" + next_uri[1]
			nextPage := NextPage{Next_uri: overrideUri}
			_, _, err := c.AviPopulateAllPkiPRofiles(client, pkiData, nextPage)
			if err != nil {
				return nil, 0, err
			}
		}
	}

	return pkiData, result.Count, nil
}

func (c *AviObjCache) AviPopulateAllPools(client *clients.AviClient, cloud string, poolData *[]AviPoolCache, overrideUri ...NextPage) (*[]AviPoolCache, int, error) {
	var uri string
	akoUser := lib.AKOUser

	if len(overrideUri) == 1 {
		uri = overrideUri[0].Next_uri
	} else {
		uri = "/api/pool/?" + "&include_name=true&cloud_ref.name=" + cloud + "&created_by=" + akoUser + "&page_size=100"
	}

	result, err := lib.AviGetCollectionRaw(client, uri)
	if err != nil {
		utils.AviLog.Warnf("Get uri %v returned err for pool %v", uri, err)
		return nil, 0, err
	}
	elems := make([]json.RawMessage, result.Count)
	err = json.Unmarshal(result.Results, &elems)
	if err != nil {
		return nil, 0, err
	}
	for i := 0; i < len(elems); i++ {
		pool := models.Pool{}
		err = json.Unmarshal(elems[i], &pool)
		if err != nil {
			utils.AviLog.Warnf("Failed to unmarshal pool data, err: %v", err)
			continue
		}

		if pool.Name == nil || pool.UUID == nil || pool.CloudConfigCksum == nil {
			utils.AviLog.Warnf("Incomplete pool data unmarshalled, %s", utils.Stringify(pool))
			continue
		}

		svc_mdata_intf := *pool.ServiceMetadata
		var svc_mdata_obj lib.ServiceMetadataObj
		if err := json.Unmarshal([]byte(svc_mdata_intf), &svc_mdata_obj); err != nil {
			utils.AviLog.Warnf("Error parsing service metadata during pool cache :%v", err)
		}

		var pkiKey NamespaceName
		if pool.PkiProfileRef != nil {
			pkiUuid := ExtractUuid(*pool.PkiProfileRef, "pkiprofile-.*.#")
			pkiName, foundPki := c.PKIProfileCache.AviCacheGetNameByUuid(pkiUuid)
			if foundPki {
				pkiKey = NamespaceName{Namespace: lib.GetTenant(), Name: pkiName.(string)}
			}
		}

		poolCacheObj := AviPoolCache{
			Name:                 *pool.Name,
			Uuid:                 *pool.UUID,
			CloudConfigCksum:     *pool.CloudConfigCksum,
			PkiProfileCollection: pkiKey,
			ServiceMetadataObj:   svc_mdata_obj,
			LastModified:         *pool.LastModified,
		}
		*poolData = append(*poolData, poolCacheObj)
	}
	if result.Next != "" {
		// It has a next page, let's recursively call the same method.
		next_uri := strings.Split(result.Next, "/api/pool")
		if len(next_uri) > 1 {
			overrideUri := "/api/pool" + next_uri[1]
			nextPage := NextPage{Next_uri: overrideUri}
			_, _, err := c.AviPopulateAllPools(client, cloud, poolData, nextPage)
			if err != nil {
				return nil, 0, err
			}
		}
	}

	return poolData, result.Count, nil
}

func (c *AviObjCache) PopulatePkiProfilesToCache(client *clients.AviClient, overrideUri ...NextPage) {
	var pkiProfData []AviPkiProfileCache
	c.AviPopulateAllPkiPRofiles(client, &pkiProfData)

	pkiCacheData := c.PKIProfileCache.ShallowCopy()
	for i, pkiCacheObj := range pkiProfData {
		k := NamespaceName{Namespace: lib.GetTenant(), Name: pkiCacheObj.Name}
		oldPkiIntf, found := c.PKIProfileCache.AviCacheGet(k)
		if found {
			oldPkiData, ok := oldPkiIntf.(*AviPkiProfileCache)
			if ok {
				if oldPkiData.InvalidData {
					pkiProfData[i].InvalidData = true
					utils.AviLog.Infof("Invalid cache data for pki: %s", k)
				}
			} else {
				utils.AviLog.Infof("Wrong data type for pki: %s in cache", k)
			}
		}
		utils.AviLog.Infof("Adding key to pki cache :%s value :%s", k, pkiCacheObj.Uuid)
		c.PKIProfileCache.AviCacheAdd(k, &pkiProfData[i])
		delete(pkiCacheData, k)
	}
	// The data that is left in pkiCacheData should be explicitly removed
	for key := range pkiCacheData {
		utils.AviLog.Infof("Deleting key from pki cache :%s", key)
		c.PKIProfileCache.AviCacheDelete(key)
	}
}

func (c *AviObjCache) PopulatePoolsToCache(client *clients.AviClient, cloud string, overrideUri ...NextPage) {
	var poolsData []AviPoolCache
	c.AviPopulateAllPools(client, cloud, &poolsData)

	poolCacheData := c.PoolCache.ShallowCopy()
	for i, poolCacheObj := range poolsData {
		k := NamespaceName{Namespace: lib.GetTenant(), Name: poolCacheObj.Name}
		oldPoolIntf, found := c.PoolCache.AviCacheGet(k)
		if found {
			oldPoolData, ok := oldPoolIntf.(*AviPoolCache)
			if ok {
				if oldPoolData.InvalidData || oldPoolData.LastModified != poolsData[i].LastModified {
					poolsData[i].InvalidData = true
					utils.AviLog.Warnf("Invalid cache data for pool: %s", k)
				}
			} else {
				utils.AviLog.Warnf("Wrong data type for pool: %s in cache", k)
			}
		}
		utils.AviLog.Debugf("Adding key to pool cache :%s value :%s", k, poolCacheObj.Uuid)
		c.PoolCache.AviCacheAdd(k, &poolsData[i])
		delete(poolCacheData, k)
	}
	// The data that is left in poolCacheData should be explicitly removed
	for key := range poolCacheData {
		utils.AviLog.Debugf("Deleting key from pool cache :%s", key)
		c.PoolCache.AviCacheDelete(key)
	}
}

func (c *AviObjCache) AviPopulateAllVSVips(client *clients.AviClient, cloud string, vsVipData *[]AviVSVIPCache, nextPage ...NextPage) (*[]AviVSVIPCache, error) {
	var uri string

	if len(nextPage) == 1 {
		uri = nextPage[0].Next_uri
	} else {
		uri = "/api/vsvip/?" + "name.contains=" + lib.GetNamePrefix() + "&include_name=true" + "&cloud_ref.name=" + cloud + "&page_size=100"
	}

	result, err := lib.AviGetCollectionRaw(client, uri)
	if err != nil {
		utils.AviLog.Warnf("Get uri %v returned err for vsvip %v", uri, err)
		return nil, err
	}
	elems := make([]json.RawMessage, result.Count)
	err = json.Unmarshal(result.Results, &elems)
	if err != nil {
		utils.AviLog.Warnf("Failed to unmarshal vsvip data, err: %v", err)
		return nil, err
	}
	for i := 0; i < len(elems); i++ {
		vsvip := models.VsVip{}
		err = json.Unmarshal(elems[i], &vsvip)
		if err != nil {
			utils.AviLog.Warnf("Failed to unmarshal vsvip data, err: %v", err)
			continue
		}

		if vsvip.Name == nil || vsvip.UUID == nil {
			utils.AviLog.Warnf("Incomplete vsvip data unmarshalled, %s", utils.Stringify(vsvip))
			continue
		}

		var fqdns []string
		var checksum string
		for _, dnsinfo := range vsvip.DNSInfo {
			fqdns = append(fqdns, *dnsinfo.Fqdn)
		}
		var vips []string
		var fips []string
		var v6ip string
		var networkNames []string
		for _, vip := range vsvip.Vip {
			vips = append(vips, *vip.IPAddress.Addr)
			if vip.FloatingIP != nil {
				fips = append(fips, *vip.FloatingIP.Addr)
			}
			if vip.Ip6Address != nil {
				v6ip = *vip.Ip6Address.Addr
			}
			if ipamNetworkSubnet := vip.IPAMNetworkSubnet; ipamNetworkSubnet != nil {
				if networkRef := *ipamNetworkSubnet.NetworkRef; networkRef != "" {
					if networkRefName := strings.Split(networkRef, "#"); len(networkRefName) == 2 {
						networkNames = append(networkNames, networkRefName[1])
					}
					if lib.UsesNetworkRef() {
						networkRefStr := strings.Split(networkRef, "/")
						networkNames = append(networkNames, networkRefStr[len(networkRefStr)-1])
					}
				}
			}
		}

		if vsvip.VsvipCloudConfigCksum != nil {
			checksum = *vsvip.VsvipCloudConfigCksum
		}

		vsVipCacheObj := AviVSVIPCache{
			Name:             *vsvip.Name,
			Uuid:             *vsvip.UUID,
			FQDNs:            fqdns,
			NetworkNames:     networkNames,
			LastModified:     *vsvip.LastModified,
			Vips:             vips,
			Fips:             fips,
			V6IP:             v6ip,
			CloudConfigCksum: checksum,
		}
		*vsVipData = append(*vsVipData, vsVipCacheObj)
	}
	if result.Next != "" {
		// It has a next page, let's recursively call the same method.
		next_uri := strings.Split(result.Next, "/api/vsvip")
		if len(next_uri) > 1 {
			overrideUri := "/api/vsvip" + next_uri[1]
			nextPage := NextPage{Next_uri: overrideUri}
			_, err := c.AviPopulateAllVSVips(client, cloud, vsVipData, nextPage)
			if err != nil {
				return nil, err
			}
		}
	}
	return vsVipData, nil
}

func (c *AviObjCache) PopulateVsVipDataToCache(client *clients.AviClient, cloud string) {
	var vsVipData []AviVSVIPCache
	c.AviPopulateAllVSVips(client, cloud, &vsVipData)

	vsVipCacheData := c.VSVIPCache.ShallowCopy()
	for i, vsVipCacheObj := range vsVipData {
		k := NamespaceName{Namespace: lib.GetTenant(), Name: vsVipCacheObj.Name}
		oldVsvipIntf, found := c.VSVIPCache.AviCacheGet(k)
		if found {
			oldVsvipData, ok := oldVsvipIntf.(*AviVSVIPCache)
			if ok {
				if oldVsvipData.InvalidData || oldVsvipData.LastModified != vsVipData[i].LastModified {
					vsVipData[i].InvalidData = true
					utils.AviLog.Warnf("Invalid cache data for vsvip: %s", k)
				}
			} else {
				utils.AviLog.Warnf("Wrong data type for vsvip: %s in cache", k)
			}
		}
		utils.AviLog.Debugf("Adding key to vsvip cache: %s, fqdns: %v", k, vsVipData[i].FQDNs)
		c.VSVIPCache.AviCacheAdd(k, &vsVipData[i])
		delete(vsVipCacheData, k)
	}
	// The data that is left in vsVipCacheData should be explicitly removed
	for key := range vsVipCacheData {
		utils.AviLog.Debugf("Deleting key from vsvip cache :%s", key)
		c.VSVIPCache.AviCacheDelete(key)
	}
}

func (c *AviObjCache) AviPopulateAllDSs(client *clients.AviClient, cloud string, DsData *[]AviDSCache, nextPage ...NextPage) (*[]AviDSCache, int, error) {
	var uri string
	akoUser := lib.AKOUser

	if len(nextPage) == 1 {
		uri = nextPage[0].Next_uri
	} else {
		uri = "/api/vsdatascriptset/?" + "&include_name=true&created_by=" + akoUser
	}

	result, err := lib.AviGetCollectionRaw(client, uri)
	if err != nil {
		utils.AviLog.Warnf("Get uri %v returned err for datascript %v", uri, err)
		return nil, 0, err
	}
	elems := make([]json.RawMessage, result.Count)
	err = json.Unmarshal(result.Results, &elems)
	if err != nil {
		utils.AviLog.Warnf("Failed to unmarshal datascript data, err: %v", err)
		return nil, 0, err
	}
	for i := 0; i < len(elems); i++ {
		ds := models.VSDataScriptSet{}
		err = json.Unmarshal(elems[i], &ds)
		if err != nil {
			utils.AviLog.Warnf("Failed to unmarshal datascript data, err: %v", err)
			continue
		}
		if ds.Name == nil || ds.UUID == nil {
			utils.AviLog.Warnf("Incomplete Datascript data unmarshalled, %s", utils.Stringify(ds))
			continue
		}

		var pgs []string
		for _, pg := range ds.PoolGroupRefs {
			// Parse each pool and populate inside pools.
			// Find out the uuid of the pool and then corresponding name
			pgUuid := ExtractUuid(pg, "poolgroup-.*.#")
			// Search the poolName using this Uuid in the poolcache.
			pgName, found := c.PgCache.AviCacheGetNameByUuid(pgUuid)
			if found {
				pgs = append(pgs, pgName.(string))
			}
		}
		dsCacheObj := AviDSCache{
			Name:       *ds.Name,
			Uuid:       *ds.UUID,
			PoolGroups: pgs,
		}
		checksum := lib.DSChecksum(dsCacheObj.PoolGroups, ds.Markers, true)
		if len(ds.Datascript) == 1 {
			checksum += utils.Hash(*ds.Datascript[0].Script)
		}
		dsCacheObj.CloudConfigCksum = checksum
		*DsData = append(*DsData, dsCacheObj)
	}
	if result.Next != "" {
		// It has a next page, let's recursively call the same method.
		next_uri := strings.Split(result.Next, "/api/vsdatascriptset")
		if len(next_uri) > 1 {
			overrideUri := "/api/vsdatascriptset" + next_uri[1]
			nextPage := NextPage{Next_uri: overrideUri}
			_, _, err := c.AviPopulateAllDSs(client, cloud, DsData, nextPage)
			if err != nil {
				return nil, 0, err
			}
		}
	}
	return DsData, result.Count, nil
}

func (c *AviObjCache) PopulateDSDataToCache(client *clients.AviClient, cloud string, overrideUri ...NextPage) {
	var DsData []AviDSCache
	c.AviPopulateAllDSs(client, cloud, &DsData)
	dsCacheData := c.DSCache.ShallowCopy()
	for i, DsCacheObj := range DsData {
		k := NamespaceName{Namespace: lib.GetTenant(), Name: DsCacheObj.Name}
		oldDSIntf, found := c.DSCache.AviCacheGet(k)
		if found {
			oldDSData, ok := oldDSIntf.(*AviDSCache)
			if ok {
				if oldDSData.InvalidData || oldDSData.LastModified != DsData[i].LastModified {
					DsData[i].InvalidData = true
					utils.AviLog.Warnf("Invalid cache data for datascript: %s", k)
				}
			} else {
				utils.AviLog.Warnf("Wrong data type for datascript: %s in cache", k)
			}
		}
		utils.AviLog.Debugf("Adding key to ds cache :%s", k)
		c.DSCache.AviCacheAdd(k, &DsData[i])
		delete(dsCacheData, k)
	}
	// The data that is left in dsCacheData should be explicitly removed
	for key := range dsCacheData {
		utils.AviLog.Debugf("Deleting key from ds cache :%s", key)
		c.DSCache.AviCacheDelete(key)
	}
}

func (c *AviObjCache) AviPopulateAllSSLKeys(client *clients.AviClient, cloud string, SslData *[]AviSSLCache, nextPage ...NextPage) (*[]AviSSLCache, int, error) {
	var uri string
	akoUser := lib.AKOUser

	if len(nextPage) == 1 {
		uri = nextPage[0].Next_uri
	} else {
		uri = "/api/sslkeyandcertificate/?" + "&created_by=" + akoUser + "&page_size=100"
	}

	result, err := lib.AviGetCollectionRaw(client, uri)
	if err != nil {
		utils.AviLog.Warnf("Get uri %v returned err for sslkeyandcertificate %v", uri, err)
		return nil, 0, err
	}
	elems := make([]json.RawMessage, result.Count)
	err = json.Unmarshal(result.Results, &elems)
	if err != nil {
		utils.AviLog.Warnf("Failed to unmarshal sslkeyandcertificate data, err: %v", err)
		return nil, 0, err
	}
	for i := 0; i < len(elems); i++ {
		sslkey := models.SSLKeyAndCertificate{}
		err = json.Unmarshal(elems[i], &sslkey)
		if err != nil {
			utils.AviLog.Warnf("Failed to unmarshal sslkeyandcertificate data, err: %v", err)
			continue
		}
		if sslkey.Name == nil || sslkey.UUID == nil {
			utils.AviLog.Warnf("Incomplete sslkey data unmarshalled, %s", utils.Stringify(sslkey))
			continue
		}

		// No support for checksum in the SSLKeyCert object, so we have to synthesize it.
		var cacertUUID, cacert string
		hasCA := false
		// find amnd store UUID of the CA cert which would be used later to calculate checksum
		if len(sslkey.CaCerts) != 0 {
			if sslkey.CaCerts[0].CaRef != nil {
				hasCA = true
				cacertUUID = ExtractUuidWithoutHash(*sslkey.CaCerts[0].CaRef, "sslkeyandcertificate-.*.")
				cacertIntf, found := c.SSLKeyCache.AviCacheGetNameByUuid(cacertUUID)
				if found {
					cacert = cacertIntf.(string)
				}
			}
		}
		emptyIngestionMarkers := utils.AviObjectMarkers{}
		sslCacheObj := AviSSLCache{
			Name:             *sslkey.Name,
			Uuid:             *sslkey.UUID,
			Cert:             *sslkey.Certificate.Certificate,
			HasCARef:         hasCA,
			CACertUUID:       cacertUUID,
			CloudConfigCksum: lib.SSLKeyCertChecksum(*sslkey.Name, *sslkey.Certificate.Certificate, cacert, emptyIngestionMarkers, sslkey.Markers, true),
		}
		*SslData = append(*SslData, sslCacheObj)
	}
	if result.Next != "" {
		// It has a next page, let's recursively call the same method.
		next_uri := strings.Split(result.Next, "/api/sslkeyandcertificate")
		if len(next_uri) > 1 {
			overrideUri := "/api/sslkeyandcertificate" + next_uri[1]
			nextPage := NextPage{Next_uri: overrideUri}
			_, _, err := c.AviPopulateAllSSLKeys(client, cloud, SslData, nextPage)
			if err != nil {
				return nil, 0, err
			}
		}
	}
	return SslData, result.Count, nil
}

func (c *AviObjCache) AviPopulateOneSSLCache(client *clients.AviClient,
	cloud string, objName string) error {
	var uri string
	akoUser := lib.AKOUser

	uri = "/api/sslkeyandcertificate?name=" + objName + "&created_by=" + akoUser

	result, err := lib.AviGetCollectionRaw(client, uri)
	if err != nil {
		utils.AviLog.Warnf("Get uri %v returned err for sslkeyandcertificate %v", uri, err)
		return err
	}
	elems := make([]json.RawMessage, result.Count)
	err = json.Unmarshal(result.Results, &elems)
	if err != nil {
		utils.AviLog.Warnf("Failed to unmarshal sslkeyandcertificate data, err: %v", err)
		return err
	}

	for i := 0; i < len(elems); i++ {
		sslkey := models.SSLKeyAndCertificate{}
		err = json.Unmarshal(elems[i], &sslkey)
		if err != nil {
			utils.AviLog.Warnf("Failed to unmarshal sslkeyandcertificate data, err: %v", err)
			continue
		}
		if sslkey.Name == nil || sslkey.UUID == nil {
			utils.AviLog.Warnf("Incomplete sslkey data unmarshalled, %s", utils.Stringify(sslkey))
			continue
		}

		//Only cache a SSL keys that belongs to this AKO.
		if !strings.HasPrefix(*sslkey.Name, lib.GetNamePrefix()) {
			continue
		}
		var cacert string
		hasCA := false
		if len(sslkey.CaCerts) != 0 {
			if sslkey.CaCerts[0].CaRef != nil {
				hasCA = true
				cacertUUID := ExtractUuidWithoutHash(*sslkey.CaCerts[0].CaRef, "sslkeyandcertificate-.*.")
				cacertIntf, found := c.SSLKeyCache.AviCacheGetNameByUuid(cacertUUID)
				if found {
					cacert = cacertIntf.(string)
				}
			}
		}
		emptyIngestionMarkers := utils.AviObjectMarkers{}
		sslCacheObj := AviSSLCache{
			Name:             *sslkey.Name,
			Uuid:             *sslkey.UUID,
			CloudConfigCksum: lib.SSLKeyCertChecksum(*sslkey.Name, *sslkey.Certificate.Certificate, cacert, emptyIngestionMarkers, sslkey.Markers, true),
			HasCARef:         hasCA,
		}
		k := NamespaceName{Namespace: lib.GetTenant(), Name: *sslkey.Name}
		c.SSLKeyCache.AviCacheAdd(k, &sslCacheObj)
		utils.AviLog.Debugf("Adding sslkey to Cache during refresh %s", k)
	}
	return nil
}

func (c *AviObjCache) AviPopulateOnePKICache(client *clients.AviClient,
	cloud string, objName string) error {
	var uri string
	akoUser := lib.AKOUser

	uri = "/api/pkiprofile?name=" + objName + "&created_by=" + akoUser

	result, err := lib.AviGetCollectionRaw(client, uri)
	if err != nil {
		utils.AviLog.Warnf("Get uri %v returned err for pkiprofile %v", uri, err)
		return err
	}
	elems := make([]json.RawMessage, result.Count)
	err = json.Unmarshal(result.Results, &elems)
	if err != nil {
		utils.AviLog.Warnf("Failed to unmarshal pkiprofile data, err: %v", err)
		return err
	}
	for i := 0; i < len(elems); i++ {
		pkikey := models.PKIprofile{}
		err = json.Unmarshal(elems[i], &pkikey)
		if err != nil {
			utils.AviLog.Warnf("Failed to unmarshal pkiprofile data, err: %v", err)
			continue
		}
		if pkikey.Name == nil || pkikey.UUID == nil {
			utils.AviLog.Warnf("Incomplete pkikey data unmarshalled, %s", utils.Stringify(pkikey))
			continue
		}
		//Only cache a SSL keys that belongs to this AKO.
		if !strings.HasPrefix(*pkikey.Name, lib.GetNamePrefix()) {
			continue
		}
		emptyIngestionMarkers := utils.AviObjectMarkers{}
		sslCacheObj := AviSSLCache{
			Name:             *pkikey.Name,
			Uuid:             *pkikey.UUID,
			CloudConfigCksum: lib.SSLKeyCertChecksum(*pkikey.Name, *pkikey.CaCerts[0].Certificate, "", emptyIngestionMarkers, pkikey.Markers, true),
		}
		k := NamespaceName{Namespace: lib.GetTenant(), Name: *pkikey.Name}
		c.SSLKeyCache.AviCacheAdd(k, &sslCacheObj)
		utils.AviLog.Debugf("Adding pkikey to Cache during refresh %s", k)
	}
	return nil
}

func (c *AviObjCache) AviPopulateOnePoolCache(client *clients.AviClient,
	cloud string, objName string) error {
	var uri string
	akoUser := lib.AKOUser

	uri = "/api/pool?name=" + objName + "&created_by=" + akoUser

	result, err := lib.AviGetCollectionRaw(client, uri)
	if err != nil {
		utils.AviLog.Warnf("Get uri %v returned err for pool %v", uri, err)
		return err
	}
	elems := make([]json.RawMessage, result.Count)
	err = json.Unmarshal(result.Results, &elems)
	if err != nil {
		utils.AviLog.Warnf("Failed to unmarshal pool data, err: %v", err)
		return err
	}
	for i := 0; i < len(elems); i++ {
		pool := models.Pool{}
		err = json.Unmarshal(elems[i], &pool)
		if err != nil {
			utils.AviLog.Warnf("Failed to unmarshal pool data, err: %v", err)
			continue
		}

		if pool.Name == nil || pool.UUID == nil || pool.CloudConfigCksum == nil {
			utils.AviLog.Warnf("Incomplete pool data unmarshalled, %s", utils.Stringify(pool))
			continue
		}
		//Only cache a Pool that belongs to this AKO.
		if !strings.HasPrefix(*pool.Name, lib.GetNamePrefix()) {
			continue
		}

		svc_mdata_intf := *pool.ServiceMetadata
		var svc_mdata_obj lib.ServiceMetadataObj
		if err := json.Unmarshal([]byte(svc_mdata_intf), &svc_mdata_obj); err != nil {
			utils.AviLog.Warnf("Error parsing service metadata during pool cache :%v", err)
		}

		var pkiKey NamespaceName
		if pool.PkiProfileRef != nil {
			pkiUuid := ExtractUuid(*pool.PkiProfileRef, "pkiprofile-.*.#")
			pkiName, foundPki := c.PKIProfileCache.AviCacheGetNameByUuid(pkiUuid)
			if foundPki {
				pkiKey = NamespaceName{Namespace: lib.GetTenant(), Name: pkiName.(string)}
			}
		}

		poolCacheObj := AviPoolCache{
			Name:                 *pool.Name,
			Uuid:                 *pool.UUID,
			CloudConfigCksum:     *pool.CloudConfigCksum,
			PkiProfileCollection: pkiKey,
			ServiceMetadataObj:   svc_mdata_obj,
			LastModified:         *pool.LastModified,
		}
		k := NamespaceName{Namespace: lib.GetTenant(), Name: *pool.Name}
		c.PoolCache.AviCacheAdd(k, &poolCacheObj)
		utils.AviLog.Debugf("Adding pool to Cache during refresh %s", k)
	}
	return nil
}

func (c *AviObjCache) AviPopulateOneVsDSCache(client *clients.AviClient,
	cloud string, objName string) error {
	var uri string
	akoUser := lib.AKOUser

	uri = "/api/vsdatascriptset?name=" + objName + "&created_by=" + akoUser

	result, err := lib.AviGetCollectionRaw(client, uri)
	if err != nil {
		utils.AviLog.Warnf("Get uri %v returned err for vsdatascript %v", uri, err)
		return err
	}
	elems := make([]json.RawMessage, result.Count)
	err = json.Unmarshal(result.Results, &elems)
	if err != nil {
		utils.AviLog.Warnf("Failed to unmarshal vsdatascript data, err: %v", err)
		return err
	}
	for i := 0; i < len(elems); i++ {
		ds := models.VSDataScriptSet{}
		err = json.Unmarshal(elems[i], &ds)
		if err != nil {
			utils.AviLog.Warnf("Failed to unmarshal datascript data, err: %v", err)
			continue
		}
		if ds.Name == nil || ds.UUID == nil {
			utils.AviLog.Warnf("Incomplete Datascript data unmarshalled, %s", utils.Stringify(ds))
			continue
		}
		//Only cache a DS that belongs to this AKO.
		if !strings.HasPrefix(*ds.Name, lib.GetNamePrefix()) {
			continue
		}
		var pgs []string
		for _, pg := range ds.PoolGroupRefs {
			// Parse each pool and populate inside pools.
			// Find out the uuid of the pool and then corresponding name
			pgUuid := ExtractUuid(pg, "poolgroup-.*.#")
			// Search the poolName using this Uuid in the poolcache.
			pgName, found := c.PgCache.AviCacheGetNameByUuid(pgUuid)
			if found {
				pgs = append(pgs, pgName.(string))
			}
		}
		dsCacheObj := AviDSCache{
			Name:       *ds.Name,
			Uuid:       *ds.UUID,
			PoolGroups: pgs,
		}
		checksum := lib.DSChecksum(dsCacheObj.PoolGroups, ds.Markers, true)
		if len(ds.Datascript) == 1 {
			checksum += utils.Hash(*ds.Datascript[0].Script)
		}
		dsCacheObj.CloudConfigCksum = checksum
		k := NamespaceName{Namespace: lib.GetTenant(), Name: *ds.Name}
		c.DSCache.AviCacheAdd(k, &dsCacheObj)
		utils.AviLog.Debugf("Adding ds to Cache during refresh %s", k)
	}
	return nil
}

func (c *AviObjCache) AviPopulateOnePGCache(client *clients.AviClient,
	cloud string, objName string) error {
	var uri string
	akoUser := lib.AKOUser

	uri = "/api/poolgroup?name=" + objName + "&created_by=" + akoUser

	result, err := lib.AviGetCollectionRaw(client, uri)
	if err != nil {
		utils.AviLog.Warnf("Get uri %v returned err for poolgroup %v", uri, err)
		return err
	}
	elems := make([]json.RawMessage, result.Count)
	err = json.Unmarshal(result.Results, &elems)
	if err != nil {
		utils.AviLog.Warnf("Failed to unmarshal poolgroup data, err: %v", err)
		return err
	}
	for i := 0; i < len(elems); i++ {
		pg := models.PoolGroup{}
		err = json.Unmarshal(elems[i], &pg)
		if err != nil {
			utils.AviLog.Warnf("Failed to unmarshal pg data, err: %v", err)
			continue
		}

		if pg.Name == nil || pg.UUID == nil || pg.CloudConfigCksum == nil {
			utils.AviLog.Warnf("Incomplete pg data unmarshalled, %s", utils.Stringify(pg))
			continue
		}
		//Only cache a PG that belongs to this AKO.
		if !strings.HasPrefix(*pg.Name, lib.GetNamePrefix()) {
			continue
		}
		var pools []string
		for _, member := range pg.Members {
			// Parse each pool and populate inside pools.
			// Find out the uuid of the pool and then corresponding name
			poolUuid := ExtractUuid(*member.PoolRef, "pool-.*.#")
			// Search the poolName using this Uuid in the poolcache.
			poolName, found := c.PoolCache.AviCacheGetNameByUuid(poolUuid)
			if found {
				pools = append(pools, poolName.(string))
			}
		}
		pgCacheObj := AviPGCache{
			Name:             *pg.Name,
			Uuid:             *pg.UUID,
			CloudConfigCksum: *pg.CloudConfigCksum,
			LastModified:     *pg.LastModified,
			Members:          pools,
		}
		k := NamespaceName{Namespace: lib.GetTenant(), Name: *pg.Name}
		c.PgCache.AviCacheAdd(k, &pgCacheObj)
		utils.AviLog.Debugf("Adding pg to Cache during refresh %s", k)
	}
	return nil
}

func (c *AviObjCache) AviPopulateOneVsVipCache(client *clients.AviClient,
	cloud string, objName string) error {
	var uri string

	uri = "/api/vsvip?name=" + objName + "&cloud_ref.name=" + cloud

	result, err := lib.AviGetCollectionRaw(client, uri)
	if err != nil {
		utils.AviLog.Warnf("Get uri %v returned err for vsvip %v", uri, err)
		return err
	}
	elems := make([]json.RawMessage, result.Count)
	err = json.Unmarshal(result.Results, &elems)
	if err != nil {
		utils.AviLog.Warnf("Failed to unmarshal vsvip data, err: %v", err)
		return err
	}
	for i := 0; i < len(elems); i++ {
		vsvip := models.VsVip{}
		err = json.Unmarshal(elems[i], &vsvip)
		if err != nil {
			utils.AviLog.Warnf("Failed to unmarshal vsvip data, err: %v", err)
			continue
		}

		if vsvip.Name == nil || vsvip.UUID == nil {
			utils.AviLog.Warnf("Incomplete vsvip data unmarshalled, %s", utils.Stringify(vsvip))
			continue
		}
		if !strings.HasPrefix(*vsvip.Name, lib.GetNamePrefix()) {
			continue
		}
		var fqdns []string
		var checksum string
		for _, dnsinfo := range vsvip.DNSInfo {
			fqdns = append(fqdns, *dnsinfo.Fqdn)
		}

		var vips []string
		var fips []string
		var v6ip string
		var networkNames []string
		for _, vip := range vsvip.Vip {
			vips = append(vips, *vip.IPAddress.Addr)
			if vip.FloatingIP != nil {
				fips = append(fips, *vip.FloatingIP.Addr)
			}
			if vip.Ip6Address != nil {
				v6ip = *vip.Ip6Address.Addr
			}
			if ipamNetworkSubnet := vip.IPAMNetworkSubnet; ipamNetworkSubnet != nil {
				if networkRef := *ipamNetworkSubnet.NetworkRef; networkRef != "" {
					if networkRefName := strings.Split(networkRef, "#"); len(networkRefName) == 2 {
						networkNames = append(networkNames, networkRefName[1])
					}
					if lib.UsesNetworkRef() {
						networkRefStr := strings.Split(networkRef, "/")
						networkNames = append(networkNames, networkRefStr[len(networkRefStr)-1])
					}
				}
			}
		}

		if vsvip.VsvipCloudConfigCksum != nil {
			checksum = *vsvip.VsvipCloudConfigCksum
		}
		vsVipCacheObj := AviVSVIPCache{
			Name:             *vsvip.Name,
			Uuid:             *vsvip.UUID,
			FQDNs:            fqdns,
			LastModified:     *vsvip.LastModified,
			Vips:             vips,
			Fips:             fips,
			V6IP:             v6ip,
			NetworkNames:     networkNames,
			CloudConfigCksum: checksum,
		}
		k := NamespaceName{Namespace: lib.GetTenant(), Name: *vsvip.Name}
		c.VSVIPCache.AviCacheAdd(k, &vsVipCacheObj)
		utils.AviLog.Debugf("Adding vsvip to Cache during refresh %s", k)
	}
	return nil
}

func (c *AviObjCache) AviPopulateOneVsHttpPolCache(client *clients.AviClient,
	cloud string, objName string) error {
	var uri string
	akoUser := lib.AKOUser

	uri = "/api/httppolicyset?name=" + objName + "&created_by=" + akoUser

	result, err := lib.AviGetCollectionRaw(client, uri)
	if err != nil {
		utils.AviLog.Warnf("Get uri %v returned err for httppol %v", uri, err)
		return err
	}
	elems := make([]json.RawMessage, result.Count)
	err = json.Unmarshal(result.Results, &elems)
	if err != nil {
		utils.AviLog.Warnf("Failed to unmarshal httppol data, err: %v", err)
		return err
	}
	for i := 0; i < len(elems); i++ {
		httppol := models.HTTPPolicySet{}
		err = json.Unmarshal(elems[i], &httppol)
		if err != nil {
			utils.AviLog.Warnf("Failed to unmarshal httppolicyset data, err: %v", err)
			continue
		}
		if httppol.Name == nil || httppol.UUID == nil || httppol.CloudConfigCksum == nil {
			utils.AviLog.Warnf("Incomplete http policy data unmarshalled, %s", utils.Stringify(httppol))
			continue
		}
		//Only cache a http policies that belongs to this AKO.
		if !strings.HasPrefix(*httppol.Name, lib.GetNamePrefix()) {
			continue
		}
		// Fetch the pgs associated with the http policyset object
		var poolGroups []string
		var pools []string
		if httppol.HTTPRequestPolicy != nil {
			for _, rule := range httppol.HTTPRequestPolicy.Rules {
				if rule.SwitchingAction != nil {
					val := reflect.ValueOf(rule.SwitchingAction)
					if !val.Elem().FieldByName("PoolGroupRef").IsNil() {
						pgUuid := ExtractUuid(*rule.SwitchingAction.PoolGroupRef, "poolgroup-.*.#")
						pgName, found := c.PgCache.AviCacheGetNameByUuid(pgUuid)
						if found {
							poolGroups = append(poolGroups, pgName.(string))
						}
					} else if !val.Elem().FieldByName("PoolRef").IsNil() {
						poolUuid := ExtractUuid(*rule.SwitchingAction.PoolRef, "pool-.*.#")
						poolName, found := c.PoolCache.AviCacheGetNameByUuid(poolUuid)
						if found {
							pools = append(pools, poolName.(string))
						}
					}
				}
			}
		}

		httpPolCacheObj := AviHTTPPolicyCache{
			Name:             *httppol.Name,
			Uuid:             *httppol.UUID,
			CloudConfigCksum: *httppol.CloudConfigCksum,
			PoolGroups:       poolGroups,
			Pools:            pools,
			LastModified:     *httppol.LastModified,
		}
		k := NamespaceName{Namespace: lib.GetTenant(), Name: *httppol.Name}
		c.HTTPPolicyCache.AviCacheAdd(k, &httpPolCacheObj)
		utils.AviLog.Debugf("Adding httppolicy to Cache during refresh %s", k)
	}
	return nil
}

func (c *AviObjCache) AviPopulateOneVsL4PolCache(client *clients.AviClient,
	cloud string, objName string) error {
	var uri string
	akoUser := lib.AKOUser

	uri = "/api/l4policyset?name=" + objName + "&created_by=" + akoUser

	result, err := lib.AviGetCollectionRaw(client, uri)
	if err != nil {
		utils.AviLog.Warnf("Get uri %v returned err for l4pol %v", uri, err)
		return err
	}
	elems := make([]json.RawMessage, result.Count)
	err = json.Unmarshal(result.Results, &elems)
	if err != nil {
		utils.AviLog.Warnf("Failed to unmarshal l4pol data, err: %v", err)
		return err
	}
	for i := 0; i < len(elems); i++ {
		l4pol := models.L4PolicySet{}
		err = json.Unmarshal(elems[i], &l4pol)
		if err != nil {
			utils.AviLog.Warnf("Failed to unmarshal l4polset data, err: %v", err)
			continue
		}
		if l4pol.Name == nil || l4pol.UUID == nil {
			utils.AviLog.Warnf("Incomplete l4 policy data unmarshalled, %s", utils.Stringify(l4pol))
			continue
		}
		//Only cache a l4 policies that belongs to this AKO.
		if !strings.HasPrefix(*l4pol.Name, lib.GetNamePrefix()) {
			continue
		}
		// Fetch the pools associated with the l4 policyset object
		var pools []string
		var ports []int64
		var protocols []string
		if l4pol.L4ConnectionPolicy != nil {
			for _, rule := range l4pol.L4ConnectionPolicy.Rules {
				protocols = append(protocols, *rule.Match.Protocol.Protocol)
				if rule.Action != nil {
					poolUuid := ExtractUuid(*rule.Action.SelectPool.PoolRef, "pool-.*.#")
					poolName, found := c.PoolCache.AviCacheGetNameByUuid(poolUuid)
					if found {
						pools = append(pools, poolName.(string))
					}
				}
				if rule.Match != nil {
					ports = append(ports, rule.Match.Port.Ports...)
				}
			}
		}
		emptyIngestionMarkers := utils.AviObjectMarkers{}
		cksum := lib.L4PolicyChecksum(ports, protocols, emptyIngestionMarkers, l4pol.Markers, true)
		l4PolCacheObj := AviL4PolicyCache{
			Name:             *l4pol.Name,
			Uuid:             *l4pol.UUID,
			Pools:            pools,
			LastModified:     *l4pol.LastModified,
			CloudConfigCksum: cksum,
		}
		k := NamespaceName{Namespace: lib.GetTenant(), Name: *l4pol.Name}
		c.L4PolicyCache.AviCacheAdd(k, &l4PolCacheObj)
		utils.AviLog.Infof("Adding l4pol to Cache during refresh %s", utils.Stringify(l4PolCacheObj))
	}
	return nil
}

func (c *AviObjCache) PopulateSSLKeyToCache(client *clients.AviClient, cloud string, overrideUri ...NextPage) {
	var SslKeyData []AviSSLCache
	c.AviPopulateAllSSLKeys(client, cloud, &SslKeyData)
	sslCacheData := c.SSLKeyCache.ShallowCopy()
	for i, SslKeyCacheObj := range SslKeyData {
		k := NamespaceName{Namespace: lib.GetTenant(), Name: SslKeyCacheObj.Name}
		oldSslkeyIntf, found := c.SSLKeyCache.AviCacheGet(k)
		if found {
			oldSslkeyData, ok := oldSslkeyIntf.(*AviSSLCache)
			if ok {
				if oldSslkeyData.InvalidData || oldSslkeyData.LastModified != SslKeyData[i].LastModified {
					SslKeyData[i].InvalidData = true
					utils.AviLog.Warnf("Invalid cache data for ssl key: %s", k)
				}
			} else {
				utils.AviLog.Warnf("Wrong data type for ssl key: %s in cache", k)
			}
		}
		utils.AviLog.Debugf("Adding key to sslkey cache :%s", k)
		c.SSLKeyCache.AviCacheAdd(k, &SslKeyData[i])
		delete(sslCacheData, k)
	}
	//The data that is left in sslCacheData should be explicitly removed
	for key := range sslCacheData {
		utils.AviLog.Debugf("Deleting key from sslkey cache :%s", key)
		c.SSLKeyCache.AviCacheDelete(key)
	}
}

func (c *AviObjCache) AviPopulateAllHttpPolicySets(client *clients.AviClient, cloud string, httpPolicyData *[]AviHTTPPolicyCache, nextPage ...NextPage) (*[]AviHTTPPolicyCache, int, error) {
	var uri string
	akoUser := lib.AKOUser

	if len(nextPage) == 1 {
		uri = nextPage[0].Next_uri
	} else {
		uri = "/api/httppolicyset/?" + "&include_name=true" + "&created_by=" + akoUser + "&page_size=100"
	}

	result, err := lib.AviGetCollectionRaw(client, uri)
	if err != nil {
		utils.AviLog.Warnf("Get uri %v returned err for httppolicyset %v", uri, err)
		return nil, 0, err
	}
	elems := make([]json.RawMessage, result.Count)
	err = json.Unmarshal(result.Results, &elems)
	if err != nil {
		utils.AviLog.Warnf("Failed to unmarshal httppolicyset data, err: %v", err)
		return nil, 0, err
	}
	for i := 0; i < len(elems); i++ {
		httppol := models.HTTPPolicySet{}
		err = json.Unmarshal(elems[i], &httppol)
		if err != nil {
			utils.AviLog.Warnf("Failed to unmarshal httppolicyset data, err: %v", err)
			continue
		}
		if httppol.Name == nil || httppol.UUID == nil || httppol.CloudConfigCksum == nil {
			utils.AviLog.Warnf("Incomplete http policy data unmarshalled, %s", utils.Stringify(httppol))
			continue
		}

		// Fetch the pgs associated with the http policyset object
		var poolGroups []string
		var pools []string
		if httppol.HTTPRequestPolicy != nil {
			for _, rule := range httppol.HTTPRequestPolicy.Rules {
				if rule.SwitchingAction != nil {
					val := reflect.ValueOf(rule.SwitchingAction)
					if !val.Elem().FieldByName("PoolGroupRef").IsNil() {
						pgUuid := ExtractUuid(*rule.SwitchingAction.PoolGroupRef, "poolgroup-.*.#")
						pgName, found := c.PgCache.AviCacheGetNameByUuid(pgUuid)
						if found {
							poolGroups = append(poolGroups, pgName.(string))
						}
					} else if !val.Elem().FieldByName("PoolRef").IsNil() {
						poolUuid := ExtractUuid(*rule.SwitchingAction.PoolRef, "pool-.*.#")
						poolName, found := c.PoolCache.AviCacheGetNameByUuid(poolUuid)
						if found {
							pools = append(pools, poolName.(string))
						}
					}
				}

			}
		}
		httpPolCacheObj := AviHTTPPolicyCache{
			Name:             *httppol.Name,
			Uuid:             *httppol.UUID,
			CloudConfigCksum: *httppol.CloudConfigCksum,
			PoolGroups:       poolGroups,
			Pools:            pools,
			LastModified:     *httppol.LastModified,
		}
		*httpPolicyData = append(*httpPolicyData, httpPolCacheObj)
	}
	if result.Next != "" {
		// It has a next page, let's recursively call the same method.
		next_uri := strings.Split(result.Next, "/api/httppolicyset")
		if len(next_uri) > 1 {
			overrideUri := "/api/httppolicyset" + next_uri[1]
			nextPage := NextPage{Next_uri: overrideUri}
			_, _, err := c.AviPopulateAllHttpPolicySets(client, cloud, httpPolicyData, nextPage)
			if err != nil {
				return nil, 0, err
			}
		}
	}
	return httpPolicyData, result.Count, nil
}

func (c *AviObjCache) PopulateHttpPolicySetToCache(client *clients.AviClient, cloud string, overrideUri ...NextPage) {
	var HttPolData []AviHTTPPolicyCache
	_, count, err := c.AviPopulateAllHttpPolicySets(client, cloud, &HttPolData)
	if err != nil || len(HttPolData) != count {
		return
	}
	httpCacheData := c.HTTPPolicyCache.ShallowCopy()
	for i, HttpPolCacheObj := range HttPolData {
		k := NamespaceName{Namespace: lib.GetTenant(), Name: HttpPolCacheObj.Name}
		oldHttppolIntf, found := c.HTTPPolicyCache.AviCacheGet(k)
		if found {
			oldHttppolData, ok := oldHttppolIntf.(*AviHTTPPolicyCache)
			if ok {
				if oldHttppolData.InvalidData || oldHttppolData.LastModified != HttPolData[i].LastModified {
					HttPolData[i].InvalidData = true
					utils.AviLog.Warnf("Invalid cache data for http policy: %s", k)
				}
			} else {
				utils.AviLog.Warnf("Wrong data type for http policy: %s in cache", k)
			}
		}
		utils.AviLog.Debugf("Adding key to httppol cache :%s", k)
		c.HTTPPolicyCache.AviCacheAdd(k, &HttPolData[i])
		delete(httpCacheData, k)
	}
	// // The data that is left in httpCacheData should be explicitly removed
	for key := range httpCacheData {
		utils.AviLog.Debugf("Deleting key from httppol cache :%s", key)
		c.HTTPPolicyCache.AviCacheDelete(key)
	}
}

func (c *AviObjCache) AviPopulateAllL4PolicySets(client *clients.AviClient, cloud string, l4PolicyData *[]AviL4PolicyCache, nextPage ...NextPage) (*[]AviL4PolicyCache, int, error) {
	var uri string
	akoUser := lib.AKOUser

	if len(nextPage) == 1 {
		uri = nextPage[0].Next_uri
	} else {
		uri = "/api/l4policyset/?" + "&include_name=true" + "&created_by=" + akoUser + "&page_size=100"
	}

	result, err := lib.AviGetCollectionRaw(client, uri)
	if err != nil {
		utils.AviLog.Warnf("Get uri %v returned err for httppolicyset %v", uri, err)
		return nil, 0, err
	}
	elems := make([]json.RawMessage, result.Count)
	err = json.Unmarshal(result.Results, &elems)
	if err != nil {
		utils.AviLog.Warnf("Failed to unmarshal httppolicyset data, err: %v", err)
		return nil, 0, err
	}
	for i := 0; i < len(elems); i++ {
		l4pol := models.L4PolicySet{}
		err = json.Unmarshal(elems[i], &l4pol)
		if err != nil {
			utils.AviLog.Warnf("Failed to unmarshal httppolicyset data, err: %v", err)
			continue
		}
		if l4pol.Name == nil || l4pol.UUID == nil {
			utils.AviLog.Warnf("Incomplete http policy data unmarshalled, %s", utils.Stringify(l4pol))
			continue
		}

		// Fetch the pgs associated with the http policyset object
		// Fetch the pools associated with the l4 policyset object
		var pools []string
		var ports []int64
		var protocols []string
		if l4pol.L4ConnectionPolicy != nil {
			for _, rule := range l4pol.L4ConnectionPolicy.Rules {
				if rule.Action != nil {
					protocol := *rule.Match.Protocol.Protocol
					if strings.Contains(protocol, utils.TCP) {
						protocol = utils.TCP
					} else {
						protocol = utils.UDP
					}
					protocols = append(protocols, protocol)
					poolUuid := ExtractUuid(*rule.Action.SelectPool.PoolRef, "pool-.*.#")
					poolName, found := c.PoolCache.AviCacheGetNameByUuid(poolUuid)
					if found {
						pools = append(pools, poolName.(string))
					}
				}
				if rule.Match != nil {
					ports = append(ports, rule.Match.Port.Ports...)
				}
			}
		}

		emptyIngestionMarkers := utils.AviObjectMarkers{}
		cksum := lib.L4PolicyChecksum(ports, protocols, emptyIngestionMarkers, l4pol.Markers, true)
		l4PolCacheObj := AviL4PolicyCache{
			Name:             *l4pol.Name,
			Uuid:             *l4pol.UUID,
			Pools:            pools,
			LastModified:     *l4pol.LastModified,
			CloudConfigCksum: cksum,
		}

		*l4PolicyData = append(*l4PolicyData, l4PolCacheObj)
	}

	if result.Next != "" {
		// It has a next page, let's recursively call the same method.
		next_uri := strings.Split(result.Next, "/api/l4policyset")
		if len(next_uri) > 1 {
			overrideUri := "/api/l4policyset" + next_uri[1]
			nextPage := NextPage{Next_uri: overrideUri}
			_, _, err := c.AviPopulateAllL4PolicySets(client, cloud, l4PolicyData, nextPage)
			if err != nil {
				return nil, 0, err
			}
		}
	}
	return l4PolicyData, result.Count, nil
}

func (c *AviObjCache) PopulateL4PolicySetToCache(client *clients.AviClient, cloud string, overrideUri ...NextPage) {
	var l4PolData []AviL4PolicyCache
	_, count, err := c.AviPopulateAllL4PolicySets(client, cloud, &l4PolData)
	if err != nil || len(l4PolData) != count {
		return
	}
	l4CacheData := c.L4PolicyCache.ShallowCopy()
	for i, l4PolCacheObj := range l4PolData {
		k := NamespaceName{Namespace: lib.GetTenant(), Name: l4PolCacheObj.Name}
		utils.AviLog.Debugf("Adding key to l4 cache :%s", utils.Stringify(l4PolCacheObj))
		c.L4PolicyCache.AviCacheAdd(k, &l4PolData[i])
		delete(l4CacheData, k)
	}
	// // The data that is left in httpCacheData should be explicitly removed
	for key := range l4CacheData {
		utils.AviLog.Debugf("Deleting key from l4policy cache :%s", key)
		c.L4PolicyCache.AviCacheDelete(key)
	}
}

func (c *AviObjCache) AviObjVrfCachePopulate(client *clients.AviClient, cloud string) error {
	if lib.GetDisableStaticRoute() {
		utils.AviLog.Debugf("Static route sync disabled, skipping vrf cache population")
		return nil
	}
	// Disable static route sync if ako is in  NodePort mode
	if lib.IsNodePortMode() {
		utils.AviLog.Infof("Static route sync disabled in NodePort Mode")
		return nil
	}
	isPrimaryAKO := lib.AKOControlConfig().GetAKOInstanceFlag()
	if !isPrimaryAKO {
		utils.AviLog.Warnf("AKO is not primary instance, not populating vrf cache.")
		return nil
	}
	uri := "/api/vrfcontext?name=" + lib.GetVrf() + "&include_name=true&cloud_ref.name=" + cloud
	result, err := lib.AviGetCollectionRaw(client, uri)
	if err != nil {
		utils.AviLog.Warnf("Get uri %v returned err %v", uri, err)
		return err
	}
	elems := make([]json.RawMessage, result.Count)
	err = json.Unmarshal(result.Results, &elems)
	if err != nil {
		utils.AviLog.Warnf("Failed to unmarshal data, err: %v", err)
		return err
	}
	for i := 0; i < result.Count; i++ {
		vrf := models.VrfContext{}
		err = json.Unmarshal(elems[i], &vrf)
		if err != nil {
			utils.AviLog.Warnf("Failed to unmarshal data, err: %v", err)
			continue
		}

		vrfName := *vrf.Name
		checksum := lib.VrfChecksum(vrfName, vrf.StaticRoutes)
		vrfCacheObj := AviVrfCache{
			Name:             vrfName,
			Uuid:             *vrf.UUID,
			CloudConfigCksum: checksum,
		}
		// set the vrf context. The result shouldn't be more than 1.
		lib.SetVrfUuid(*vrf.UUID)
		utils.AviLog.Debugf("Adding vrf to Cache %s", vrfName)
		c.VrfCache.AviCacheAdd(vrfName, &vrfCacheObj)
	}
	return nil
}

func (c *AviObjCache) AviObjVSCachePopulate(client *clients.AviClient, cloud string, vsCacheCopy *[]NamespaceName, overrideUri ...NextPage) error {
	var rest_response interface{}
	akoUser := lib.AKOUser
	var uri string
	httpCacheRefreshCount := 1 // Refresh count for http cache is attempted once per page
	if len(overrideUri) == 1 {
		uri = overrideUri[0].Next_uri
	} else {
		uri = "/api/virtualservice/?" + "include_name=true" + "&cloud_ref.name=" + cloud + "&created_by=" + akoUser + "&page_size=100"
	}

	err := lib.AviGet(client, uri, &rest_response)
	if err != nil {
		utils.AviLog.Warnf("Vs Get uri %v returned err %v", uri, err)
		return err
	} else {
		resp, ok := rest_response.(map[string]interface{})
		if !ok {
			utils.AviLog.Warnf("Vs Get uri %v returned %v type %T", uri,
				rest_response, rest_response)
			return errors.New("VS type is wrong")
		}
		utils.AviLog.Debugf("Vs Get uri %v returned %v vses", uri,
			resp["count"])
		results, ok := resp["results"].([]interface{})
		if !ok {
			utils.AviLog.Warnf("results not of type []interface{} Instead of type %T", resp["results"])
			return errors.New("Results are not of right type for VS")
		}
		for _, vs_intf := range results {

			vs, ok := vs_intf.(map[string]interface{})
			if !ok {
				utils.AviLog.Warnf("vs_intf not of type map[string] interface{}. Instead of type %T", vs_intf)
				continue
			}
			svc_mdata_intf, ok := vs["service_metadata"]
			var svc_mdata_obj lib.ServiceMetadataObj
			if ok {
				if err := json.Unmarshal([]byte(svc_mdata_intf.(string)),
					&svc_mdata_obj); err != nil {
					utils.AviLog.Warnf("Error parsing service metadata during vs cache :%v", err)
				}
			}
			var sni_child_collection []string
			vh_child, found := vs["vh_child_vs_uuid"]
			if found {
				for _, child := range vh_child.([]interface{}) {
					sni_child_collection = append(sni_child_collection, child.(string))
				}

			}
			vs_parent_ref, foundParent := vs["vh_parent_vs_ref"]
			var parentVSKey NamespaceName
			if foundParent {
				vs_uuid := ExtractUuid(vs_parent_ref.(string), "virtualservice-.*.#")
				utils.AviLog.Debugf("extracted the vs uuid from parent ref during cache population: %s", vs_uuid)
				// Now let's get the VS key from this uuid
				vsKey, gotVS := c.VsCacheLocal.AviCacheGetKeyByUuid(vs_uuid)
				if gotVS {
					parentVSKey = vsKey.(NamespaceName)
				}

			}
			if vs["cloud_config_cksum"] != nil {
				k := NamespaceName{Namespace: lib.GetTenant(), Name: vs["name"].(string)}
				*vsCacheCopy = RemoveNamespaceName(*vsCacheCopy, k)
				var vsVipKey []NamespaceName
				var sslKeys []NamespaceName
				var dsKeys []NamespaceName
				var httpKeys []NamespaceName
				var l4Keys []NamespaceName
				var poolgroupKeys []NamespaceName
				var poolKeys []NamespaceName
				var sharedVsOrL4 bool

				// Populate the VSVIP cache
				if vs["vsvip_ref"] != nil {
					// find the vsvip name from the vsvip cache
					vsVipUuid := ExtractUuid(vs["vsvip_ref"].(string), "vsvip-.*.#")
					objKey, objFound := c.VSVIPCache.AviCacheGetKeyByUuid(vsVipUuid)
					if objFound {
						vsVip, foundVip := c.VSVIPCache.AviCacheGet(objKey)
						if foundVip {
							vsVipData, ok := vsVip.(*AviVSVIPCache)
							if ok {
								vipKey := NamespaceName{Namespace: lib.GetTenant(), Name: vsVipData.Name}
								vsVipKey = append(vsVipKey, vipKey)
							}
						}
					}
				}

				if vs["ssl_key_and_certificate_refs"] != nil {
					for _, ssl := range vs["ssl_key_and_certificate_refs"].([]interface{}) {
						// find the sslkey name from the ssl key cache
						sslUuid := ExtractUuid(ssl.(string), "sslkeyandcertificate-.*.#")
						sslName, foundssl := c.SSLKeyCache.AviCacheGetNameByUuid(sslUuid)
						if foundssl {
							sslKey := NamespaceName{Namespace: lib.GetTenant(), Name: sslName.(string)}
							sslKeys = append(sslKeys, sslKey)

							sslIntf, _ := c.SSLKeyCache.AviCacheGet(sslKey)
							sslData := sslIntf.(*AviSSLCache)
							// Populate CAcert if available
							if sslData.CACertUUID != "" {
								caName, found := c.SSLKeyCache.AviCacheGetNameByUuid(sslData.CACertUUID)
								if found {
									caCertKey := NamespaceName{Namespace: lib.GetTenant(), Name: caName.(string)}
									sslKeys = append(sslKeys, caCertKey)
								}
							}
						}
					}
				}
				if vs["vs_datascripts"] != nil {
					for _, ds_intf := range vs["vs_datascripts"].([]interface{}) {
						// find the sslkey name from the ssl key cache
						dsmap, ok := ds_intf.(map[string]interface{})
						if ok {
							dsUuid := ExtractUuid(dsmap["vs_datascript_set_ref"].(string), "vsdatascriptset-.*.#")

							dsName, foundDs := c.DSCache.AviCacheGetNameByUuid(dsUuid)
							if foundDs {
								dsKey := NamespaceName{Namespace: lib.GetTenant(), Name: dsName.(string)}
								// Fetch the associated PGs with the DS.
								dsObj, _ := c.DSCache.AviCacheGet(dsKey)
								for _, pgName := range dsObj.(*AviDSCache).PoolGroups {
									// For each PG, formulate the key and then populate the pg collection cache
									pgKey := NamespaceName{Namespace: lib.GetTenant(), Name: pgName}
									poolgroupKeys = append(poolgroupKeys, pgKey)
									pgpoolKeys := c.AviPGPoolCachePopulate(client, cloud, pgName)
									poolKeys = append(poolKeys, pgpoolKeys...)
								}
								dsKeys = append(dsKeys, dsKey)
								sharedVsOrL4 = true
							}
						}
					}
				}
				// Handle L4 vs - pg references
				if vs["service_pool_select"] != nil {
					for _, pg_intf := range vs["service_pool_select"].([]interface{}) {
						// find the sslkey name from the ssl key cache
						pgmap, ok := pg_intf.(map[string]interface{})
						if ok {
							pgUuid := ExtractUuid(pgmap["service_pool_group_ref"].(string), "poolgroup-.*.#")

							pgName, foundpg := c.PgCache.AviCacheGetNameByUuid(pgUuid)
							if foundpg {
								pgKey := NamespaceName{Namespace: lib.GetTenant(), Name: pgName.(string)}
								poolgroupKeys = append(poolgroupKeys, pgKey)
								pgpoolKeys := c.AviPGPoolCachePopulate(client, cloud, pgName.(string))
								poolKeys = append(poolKeys, pgpoolKeys...)
								sharedVsOrL4 = true
							}
						}
					}
				}
				if vs["l4_policies"] != nil {
					for _, l4_intf := range vs["l4_policies"].([]interface{}) {
						l4map, ok := l4_intf.(map[string]interface{})
						if ok {
							l4PolUuid := ExtractUuid(l4map["l4_policy_set_ref"].(string), "l4policyset-.*.#")
							l4Name, foundl4pol := c.L4PolicyCache.AviCacheGetNameByUuid(l4PolUuid)
							if foundl4pol {
								sharedVsOrL4 = true
								l4key := NamespaceName{Namespace: lib.GetTenant(), Name: l4Name.(string)}
								l4Obj, _ := c.L4PolicyCache.AviCacheGet(l4key)
								for _, poolName := range l4Obj.(*AviL4PolicyCache).Pools {
									poolKey := NamespaceName{Namespace: lib.GetTenant(), Name: poolName}
									poolKeys = append(poolKeys, poolKey)
								}
								l4Keys = append(l4Keys, l4key)
							}
						}
					}
				}
				if vs["http_policies"] != nil {
					for _, http_intf := range vs["http_policies"].([]interface{}) {
						httpmap, ok := http_intf.(map[string]interface{})
						if ok {
							httpUuid := ExtractUuid(httpmap["http_policy_set_ref"].(string), "httppolicyset-.*.#")
							httpName, foundhttp := c.HTTPPolicyCache.AviCacheGetNameByUuid(httpUuid)
							// If the httppol is not found in the cache, do an explicit get
							if !foundhttp && !sharedVsOrL4 && httpCacheRefreshCount > 0 {
								// We do a full refresh of the httpcache once per page, if we detect a data discrepancy
								httpCacheRefreshCount = httpCacheRefreshCount - 1
								c.PopulateHttpPolicySetToCache(client, cloud)
								httpName, foundhttp = c.HTTPPolicyCache.AviCacheGetNameByUuid(httpUuid)
								if !foundhttp {
									// If still the httpName is not found. Log an error saying, this VS may not behave appropriately.
									utils.AviLog.Warnf("HTTPPolicySet not found in Avi for VS: %s for httpUUID: %s", vs["name"].(string), httpUuid)
								}
							}
							if foundhttp {
								httpKey := NamespaceName{Namespace: lib.GetTenant(), Name: httpName.(string)}
								httpObj, _ := c.HTTPPolicyCache.AviCacheGet(httpKey)
								for _, pgName := range httpObj.(*AviHTTPPolicyCache).PoolGroups {
									// For each PG, formulate the key and then populate the pg collection cache
									pgKey := NamespaceName{Namespace: lib.GetTenant(), Name: pgName}
									poolgroupKeys = append(poolgroupKeys, pgKey)
									pgpoolKeys := c.AviPGPoolCachePopulate(client, cloud, pgName)
									poolKeys = append(poolKeys, pgpoolKeys...)
								}
								httpKeys = append(httpKeys, httpKey)
							}
						}
					}
				}
				// Populate the vscache meta object here.
				vsMetaObj := AviVsCache{
					Name:                 vs["name"].(string),
					Uuid:                 vs["uuid"].(string),
					VSVipKeyCollection:   vsVipKey,
					HTTPKeyCollection:    httpKeys,
					DSKeyCollection:      dsKeys,
					SSLKeyCertCollection: sslKeys,
					PGKeyCollection:      poolgroupKeys,
					PoolKeyCollection:    poolKeys,
					CloudConfigCksum:     vs["cloud_config_cksum"].(string),
					SNIChildCollection:   sni_child_collection,
					ParentVSRef:          parentVSKey,
					ServiceMetadataObj:   svc_mdata_obj,
					L4PolicyCollection:   l4Keys,
					LastModified:         vs["_last_modified"].(string),
				}
				if val, ok := vs["enable_rhi"]; ok {
					vsMetaObj.EnableRhi = val.(bool)
				}
				c.VsCacheLocal.AviCacheAdd(k, &vsMetaObj)
				utils.AviLog.Debugf("Added VS cache key :%s", utils.Stringify(&vsMetaObj))

			}
		}
		if resp["next"] != nil {
			// It has a next page, let's recursively call the same method.
			next_uri := strings.Split(resp["next"].(string), "/api/virtualservice")
			utils.AviLog.Debugf("Found next page for vs, uri: %s", next_uri)
			if len(next_uri) > 1 {
				overrideUri := "/api/virtualservice" + next_uri[1]
				utils.AviLog.Debugf("Next page uri for vs: %s", overrideUri)
				nextPage := NextPage{Next_uri: overrideUri}
				c.AviObjVSCachePopulate(client, cloud, vsCacheCopy, nextPage)
			}
		}
	}
	return nil
}

func (c *AviObjCache) AviObjOneVSCachePopulate(client *clients.AviClient, cloud string, vsName string) error {
	// This method should be called only from layer-3 during a retry.
	var rest_response interface{}
	akoUser := lib.AKOUser
	var uri string

	uri = "/api/virtualservice?name=" + vsName + "&cloud_ref.name=" + cloud + "&created_by=" + akoUser

	utils.AviLog.Debugf("Refreshing cache for vs uri: %s", uri)
	err := lib.AviGet(client, uri, &rest_response)
	if err != nil {
		utils.AviLog.Warnf("Vs Get uri %v returned err %v", uri, err)
		return err
	} else {
		resp, ok := rest_response.(map[string]interface{})
		if !ok {
			utils.AviLog.Warnf("Vs Get uri %v returned %v type %T", uri,
				rest_response, rest_response)
			return errors.New("VS type is wrong")
		}
		utils.AviLog.Debugf("Vs Get uri %v returned %v vses", uri,
			resp["count"])
		k := NamespaceName{Namespace: lib.GetTenant(), Name: vsName}
		objCount, _ := resp["count"]
		if objCount == 0.0 {
			utils.AviLog.Debugf("Empty response removing VS meta :%s", k)
			// Count is 0 delete the VS.
			c.VsCacheMeta.AviCacheDelete(k)
			return nil
		}
		results, ok := resp["results"].([]interface{})
		if !ok {
			utils.AviLog.Warnf("results not of type []interface{} Instead of type %T", resp["results"])
			return errors.New("Results are not of right type for VS")
		}
		for _, vs_intf := range results {
			vs := vs_intf.(map[string]interface{})
			svc_mdata_intf, ok := vs["service_metadata"]
			var svc_mdata_obj lib.ServiceMetadataObj
			if ok {
				if err := json.Unmarshal([]byte(svc_mdata_intf.(string)),
					&svc_mdata_obj); err != nil {
					utils.AviLog.Warnf("Error parsing service metadata during vs cache :%v", err)
				}
			}
			var sni_child_collection []string
			vh_child, found := vs["vh_child_vs_uuid"]
			if found {
				for _, child := range vh_child.([]interface{}) {
					sni_child_collection = append(sni_child_collection, child.(string))
				}

			}
			vs_parent_ref, foundParent := vs["vh_parent_vs_ref"]
			var parentVSKey NamespaceName
			if foundParent {
				vs_uuid := ExtractUuidWithoutHash(vs_parent_ref.(string), "virtualservice-.*.")
				utils.AviLog.Debugf("extracted the vs uuid from parent ref during cache population: %s", vs_uuid)
				// Now let's get the VS key from this uuid
				vsKey, gotVS := c.VsCacheMeta.AviCacheGetKeyByUuid(vs_uuid)
				if gotVS {
					parentVSKey = vsKey.(NamespaceName)
				}

			}
			if vs["cloud_config_cksum"] != nil {
				var vsVipKey []NamespaceName
				var sslKeys []NamespaceName
				var dsKeys []NamespaceName
				var httpKeys []NamespaceName
				var poolgroupKeys []NamespaceName
				var poolKeys []NamespaceName
				var l4Keys []NamespaceName

				// Populate the VSVIP cache
				if vs["vsvip_ref"] != nil {
					// find the vsvip name from the vsvip cache
					vsVipUuid := ExtractUuidWithoutHash(vs["vsvip_ref"].(string), "vsvip-.*.")
					vsVip, foundVip := c.VSVIPCache.AviCacheGet(vsVipUuid)

					if foundVip {
						vsVipData, ok := vsVip.(*AviVSVIPCache)
						if ok {
							vipKey := NamespaceName{Namespace: lib.GetTenant(), Name: vsVipData.Name}
							vsVipKey = append(vsVipKey, vipKey)
						}
					}
				}

				if vs["ssl_key_and_certificate_refs"] != nil {
					for _, ssl := range vs["ssl_key_and_certificate_refs"].([]interface{}) {
						// find the sslkey name from the ssl key cache
						sslUuid := ExtractUuidWithoutHash(ssl.(string), "sslkeyandcertificate-.*.")
						sslName, foundssl := c.SSLKeyCache.AviCacheGetNameByUuid(sslUuid)
						if foundssl {
							sslKey := NamespaceName{Namespace: lib.GetTenant(), Name: sslName.(string)}
							sslKeys = append(sslKeys, sslKey)

							sslIntf, _ := c.SSLKeyCache.AviCacheGet(sslKey)
							sslData := sslIntf.(*AviSSLCache)
							// Populate CAcert if available
							if sslData.CACertUUID != "" {
								caName, found := c.SSLKeyCache.AviCacheGetNameByUuid(sslData.CACertUUID)
								if found {
									caCertKey := NamespaceName{Namespace: lib.GetTenant(), Name: caName.(string)}
									sslKeys = append(sslKeys, caCertKey)
								}
							}
						}
					}
				}
				if vs["vs_datascripts"] != nil {
					for _, ds_intf := range vs["vs_datascripts"].([]interface{}) {
						// find the sslkey name from the ssl key cache
						dsmap, ok := ds_intf.(map[string]interface{})
						if ok {
							dsUuid := ExtractUuidWithoutHash(dsmap["vs_datascript_set_ref"].(string), "vsdatascriptset-.*.")

							dsName, foundDs := c.DSCache.AviCacheGetNameByUuid(dsUuid)
							if foundDs {
								dsKey := NamespaceName{Namespace: lib.GetTenant(), Name: dsName.(string)}
								// Fetch the associated PGs with the DS.
								dsObj, _ := c.DSCache.AviCacheGet(dsKey)
								for _, pgName := range dsObj.(*AviDSCache).PoolGroups {
									// For each PG, formulate the key and then populate the pg collection cache
									pgKey := NamespaceName{Namespace: lib.GetTenant(), Name: pgName}
									poolgroupKeys = append(poolgroupKeys, pgKey)
									pgpoolKeys := c.AviPGPoolCachePopulate(client, cloud, pgName)
									poolKeys = append(poolKeys, pgpoolKeys...)
								}
								dsKeys = append(dsKeys, dsKey)
							}
						}
					}
				}
				// Handle L4 vs - pg references
				if vs["service_pool_select"] != nil {
					for _, pg_intf := range vs["service_pool_select"].([]interface{}) {
						// find the sslkey name from the ssl key cache
						pgmap, ok := pg_intf.(map[string]interface{})
						if ok {
							pgUuid := ExtractUuidWithoutHash(pgmap["service_pool_group_ref"].(string), "poolgroup-.*.")

							pgName, foundpg := c.PgCache.AviCacheGetNameByUuid(pgUuid)
							if foundpg {
								pgKey := NamespaceName{Namespace: lib.GetTenant(), Name: pgName.(string)}
								poolgroupKeys = append(poolgroupKeys, pgKey)
								pgpoolKeys := c.AviPGPoolCachePopulate(client, cloud, pgName.(string))
								poolKeys = append(poolKeys, pgpoolKeys...)
							}
						}
					}
				}
				if vs["l4_policies"] != nil {
					for _, l4_intf := range vs["l4_policies"].([]interface{}) {
						l4map, ok := l4_intf.(map[string]interface{})
						if ok {
							l4PolUuid := ExtractUuid(l4map["l4_policy_set_ref"].(string), "l4policyset-.*.#")
							l4Name, foundl4pol := c.L4PolicyCache.AviCacheGetNameByUuid(l4PolUuid)
							if foundl4pol {
								l4key := NamespaceName{Namespace: lib.GetTenant(), Name: l4Name.(string)}
								l4Obj, _ := c.L4PolicyCache.AviCacheGet(l4key)
								for _, poolName := range l4Obj.(*AviL4PolicyCache).Pools {
									poolKey := NamespaceName{Namespace: lib.GetTenant(), Name: poolName}
									poolKeys = append(poolKeys, poolKey)
								}
								l4Keys = append(l4Keys, l4key)
							}
						}
					}
				}
				if vs["http_policies"] != nil {
					for _, http_intf := range vs["http_policies"].([]interface{}) {
						// find the sslkey name from the ssl key cache
						httpmap, ok := http_intf.(map[string]interface{})
						if ok {
							httpUuid := ExtractUuidWithoutHash(httpmap["http_policy_set_ref"].(string), "httppolicyset-.*.")

							httpName, foundhttp := c.HTTPPolicyCache.AviCacheGetNameByUuid(httpUuid)
							if foundhttp {
								httpKey := NamespaceName{Namespace: lib.GetTenant(), Name: httpName.(string)}
								httpObj, _ := c.HTTPPolicyCache.AviCacheGet(httpKey)
								for _, pgName := range httpObj.(*AviHTTPPolicyCache).PoolGroups {
									// For each PG, formulate the key and then populate the pg collection cache
									pgKey := NamespaceName{Namespace: lib.GetTenant(), Name: pgName}
									poolgroupKeys = append(poolgroupKeys, pgKey)
									pgpoolKeys := c.AviPGPoolCachePopulate(client, cloud, pgName)
									poolKeys = append(poolKeys, pgpoolKeys...)
								}
								httpKeys = append(httpKeys, httpKey)
							}
						}
					}
				}
				// Populate the vscache meta object here.
				vsMetaObj := AviVsCache{
					Name:                 vs["name"].(string),
					Uuid:                 vs["uuid"].(string),
					VSVipKeyCollection:   vsVipKey,
					HTTPKeyCollection:    httpKeys,
					DSKeyCollection:      dsKeys,
					SSLKeyCertCollection: sslKeys,
					PGKeyCollection:      poolgroupKeys,
					PoolKeyCollection:    poolKeys,
					CloudConfigCksum:     vs["cloud_config_cksum"].(string),
					SNIChildCollection:   sni_child_collection,
					ParentVSRef:          parentVSKey,
					L4PolicyCollection:   l4Keys,
					ServiceMetadataObj:   svc_mdata_obj,
				}
				if val, ok := vs["enable_rhi"]; ok {
					vsMetaObj.EnableRhi = val.(bool)
				}
				c.VsCacheMeta.AviCacheAdd(k, &vsMetaObj)
				vs_cache, found := c.VsCacheMeta.AviCacheGet(parentVSKey)
				if found {
					parentVsObj, ok := vs_cache.(*AviVsCache)
					if !ok {
						utils.AviLog.Warnf("key: %s, msg: invalid vs object found.", parentVSKey)
					} else {
						parentVsObj.AddToSNIChildCollection(vs["uuid"].(string))
						utils.AviLog.Infof("Updated Parents VS :%s", utils.Stringify(parentVsObj))
					}
				}
				utils.AviLog.Debugf("Added VS during refresh with cache key :%v", utils.Stringify(&vsMetaObj))
			}
		}
	}
	return nil
}

func (c *AviObjCache) AviPGPoolCachePopulate(client *clients.AviClient, cloud string, pgName string) []NamespaceName {
	var poolKeyCollection []NamespaceName

	k := NamespaceName{Namespace: lib.GetTenant(), Name: pgName}
	// Find the pools associated with this PG and populate them
	pgObj, ok := c.PgCache.AviCacheGet(k)
	// Get the members from this and populate the VS ref
	if ok {
		for _, poolName := range pgObj.(*AviPGCache).Members {
			k := NamespaceName{Namespace: lib.GetTenant(), Name: poolName}
			poolKeyCollection = append(poolKeyCollection, k)
		}
	} else {
		// PG not found in the cache. Let's try a refresh explicitly
		c.AviPopulateOnePGCache(client, cloud, pgName)
		pgObj, ok = c.PgCache.AviCacheGet(k)
		if ok {
			utils.AviLog.Debugf("Found PG on refresh: %s", pgName)
			for _, poolName := range pgObj.(*AviPGCache).Members {
				k := NamespaceName{Namespace: lib.GetTenant(), Name: poolName}
				poolKeyCollection = append(poolKeyCollection, k)
			}
		} else {
			utils.AviLog.Warnf("PG not found on cache refresh: %s", pgName)
		}
	}
	return poolKeyCollection
}

func IsAviClusterActive(client *clients.AviClient) bool {
	uri := "/api/cluster/runtime"
	var response map[string]interface{}
	err := lib.AviGet(client, uri, &response)
	if err != nil {
		utils.AviLog.Warnf("Cluster status Get uri %v returned err %v", uri, err)
		return false
	}

	clusterStateMap, ok := response["cluster_state"].(map[string]interface{})
	if !ok {
		utils.AviLog.Warnf("Unexpected type for cluster_state map %T", response["cluster_states"])
		return false
	}

	clusterState, ok := clusterStateMap["state"].(string)
	if !ok {
		utils.AviLog.Warnf("Unexpected type for cluster state %T", clusterStateMap["state"])
		return false
	}

	utils.AviLog.Infof("Avi cluster state is %s", clusterState)
	if clusterState == "CLUSTER_UP_NO_HA" || clusterState == "CLUSTER_UP_HA_ACTIVE" || clusterState == "CLUSTER_UP_HA_COMPROMISED" {
		return true
	}

	return false
}

func (c *AviObjCache) AviClusterStatusPopulate(client *clients.AviClient) error {
	uri := "/api/cluster/runtime"
	var response map[string]interface{}
	err := lib.AviGet(client, uri, &response)
	if err != nil {
		utils.AviLog.Warnf("Cluster status Get uri %v returned err %v", uri, err)
		return err
	}

	// clusterRuntime would be nil on bootup, and must not change after that, if it does
	// AKO apiserver is shutdown and AKO restarted
	var runtimeCache *AviClusterRuntimeCache
	clusterRuntime, ok := c.ClusterStatusCache.AviCacheGet(lib.ClusterStatusCacheKey)
	if ok {
		runtimeCache, ok = clusterRuntime.(*AviClusterRuntimeCache)
		if !ok {
			utils.AviLog.Warnf("ClusterRuntime is not of type AviClusterRuntimeCache")
			return fmt.Errorf("ClusterRuntime is not of type AviClusterRuntimeCache")
		}
	} else {
		utils.AviLog.Infof("Unable to find ClusterRuntime in cache")
	}

	nodeStates := response["node_states"].([]interface{})
	for _, node := range nodeStates {
		nodeObj := node.(map[string]interface{})
		if nodeObj["role"].(string) == "CLUSTER_LEADER" {
			nodeObjUpSince, okUpSince := nodeObj["up_since"].(string)
			nodeObjName, okName := nodeObj["name"].(string)
			if runtimeCache != nil &&
				okUpSince && okName &&
				(runtimeCache.UpSince != nodeObjUpSince || runtimeCache.Name != nodeObjName) {
				// reboot AKO
				utils.AviLog.Warnf("Avi controller leader node or leader uptime changed, shutting down AKO")
				lib.ShutdownApi()
				return nil
			}

			setCacheVal := &AviClusterRuntimeCache{
				Name:    nodeObjName,
				UpSince: nodeObjUpSince,
			}
			c.ClusterStatusCache.AviCacheAdd(lib.ClusterStatusCacheKey, setCacheVal)
			utils.AviLog.Infof("Added ClusterStatusCache cache key %v val %v", lib.ClusterStatusCacheKey, setCacheVal)
			break
		}
	}

	return nil
}

func (c *AviObjCache) AviCloudPropertiesPopulate(client *clients.AviClient, cloudName string) error {
	uri := "/api/cloud/?include_name&name=" + cloudName
	result, err := lib.AviGetCollectionRaw(client, uri)
	if err != nil {
		utils.AviLog.Warnf("CloudProperties Get uri %v returned err %v", uri, err)
		return err
	}

	elems := make([]json.RawMessage, result.Count)
	err = json.Unmarshal(result.Results, &elems)
	if err != nil {
		utils.AviLog.Warnf("Failed to unmarshal data, err: %v", err)
		return err
	}

	if result.Count != 1 {
		utils.AviLog.Errorf("Cloud details not found for cloud name: %s", cloudName)
		return fmt.Errorf("Cloud details not found for cloud name: %s", cloudName)
	}

	cloud := models.Cloud{}
	if err = json.Unmarshal(elems[0], &cloud); err != nil {
		utils.AviLog.Warnf("Failed to unmarshal cloud data, err: %v", err)
		return err
	}

	vtype := *cloud.Vtype
	if vtype == lib.CLOUD_NSXT {
		// Check the transport zone type.
		if cloud.NsxtConfiguration != nil {
			if cloud.NsxtConfiguration.DataNetworkConfig != nil {
				tz := *cloud.NsxtConfiguration.DataNetworkConfig.TzType
				lib.SetNSXTTransportZone(tz)
			}
		}

	}
	cloud_obj := &AviCloudPropertyCache{Name: cloudName, VType: vtype}

	ipamType := ""
	if cloud.IPAMProviderRef != nil && *cloud.IPAMProviderRef != "" {
		ipamType = c.AviIPAMPropertyPopulate(client, *cloud.IPAMProviderRef)
	}
	cloud_obj.IPAMType = ipamType
	utils.AviLog.Infof("IPAM Provider type configured as %s for Cloud %s", cloud_obj.IPAMType, cloud_obj.Name)

	subdomains := c.AviDNSPropertyPopulate(client, *cloud.UUID)
	if len(subdomains) == 0 {
		utils.AviLog.Warnf("Cloud: %v does not have a dns provider configured", cloudName)
	}
	if subdomains != nil {
		cloud_obj.NSIpamDNS = subdomains
	}

	c.CloudKeyCache.AviCacheAdd(cloudName, cloud_obj)
	utils.AviLog.Infof("Added CloudKeyCache cache key %v val %v", cloudName, cloud_obj)
	return nil
}

func (c *AviObjCache) AviIPAMPropertyPopulate(client *clients.AviClient, ipamRef string) string {
	ipamName := strings.Split(ipamRef, "#")[1]
	uri := "/api/ipamdnsproviderprofile/?include_name&name=" + ipamName
	result, err := lib.AviGetCollectionRaw(client, uri)
	if err != nil {
		utils.AviLog.Warnf("IPAMProvider Get uri %v returned err %v", uri, err)
		return ""
	}

	elems := make([]json.RawMessage, result.Count)
	err = json.Unmarshal(result.Results, &elems)
	if err != nil {
		utils.AviLog.Warnf("Failed to unmarshal data, err: %v", err)
		return ""
	}

	if result.Count != 1 {
		utils.AviLog.Errorf("IPAM details not found for cloud name: %s", ipamName)
	}

	ipam := models.IPAMDNSProviderProfile{}
	if err = json.Unmarshal(elems[0], &ipam); err != nil {
		utils.AviLog.Warnf("Failed to unmarshal ipam provider data, err: %v", err)
		return ""
	}

	if ipam.Type == nil {
		return ""
	}
	return *ipam.Type
}

func (c *AviObjCache) AviDNSPropertyPopulate(client *clients.AviClient, cloudUUID string) []string {
	type IPAMDNSProviderProfileDomainList struct {

		// List of service domains.
		Domains []*string `json:"domains,omitempty"`
	}
	var dnsSubDomains []string
	domainList := IPAMDNSProviderProfileDomainList{}
	uri := "/api/ipamdnsproviderprofiledomainlist?cloud_uuid=" + cloudUUID

	err := lib.AviGet(client, uri, &domainList)
	if err != nil {
		utils.AviLog.Warnf("DNSProperty Get uri %v returned err %v", uri, err)
		return nil
	}

	for _, subdomain := range domainList.Domains {
		utils.AviLog.Debugf("Found DNS Domain name: %v", *subdomain)
		dnsSubDomains = append(dnsSubDomains, *subdomain)
	}

	return dnsSubDomains
}

var controllerClusterUUID string

// SetControllerClusterUUID sets the controller cluster's UUID value which is fetched from
// /api/cluster. If the variable controllerClusterUUID is already set, no API call will be
// made.
func SetControllerClusterUUID(clientPool *utils.AviRestClientPool) error {
	if controllerClusterUUID != "" {
		// controller cluster UUID already set
		return nil
	}
	uri := "/api/cluster"
	var result interface{}
	err := lib.AviGet(clientPool.AviClient[0], uri, &result)
	if err != nil {
		return fmt.Errorf("cluster get uri %s returned error %v", uri, err)
	}
	controllerClusterData, ok := result.(map[string]interface{})
	if !ok {
		return fmt.Errorf("error in parsing controller cluster response: %v", result)
	}
	if clusterUUID, parsed := controllerClusterData["uuid"].(string); parsed {
		controllerClusterUUID = clusterUUID
	} else {
		return fmt.Errorf("error in parsing controller cluster uuid field from %v", controllerClusterData)
	}

	return nil
}

// GetControllerClusterUUID returns the value in controllerClusterUUID variable.
func GetControllerClusterUUID() string {
	return controllerClusterUUID
}

func ValidateUserInput(client *clients.AviClient) (bool, error) {
	// add other step0 validation logics here -> isValid := check1 && check2 && ...

	var err error
	isTenantValid := checkTenant(client, &err)
	isCloudValid := checkAndSetCloudType(client, &err)
	isRequiredValuesValid := checkRequiredValuesYaml(&err)
	isSegroupValid := validateAndConfigureSeGroup(client, &err)
	if lib.GetAdvancedL4() {
		if isTenantValid &&
			isCloudValid &&
			isRequiredValuesValid &&
			isSegroupValid {
			utils.AviLog.Info("All values verified for advanced L4, proceeding with bootup.")
			return true, nil
		}
		return false, err
	}

	isNodeNetworkValid := checkNodeNetwork(client, &err)
	isBGPConfigurationValid := checkBGPParams(&err)
	isPublicCloudConfigValid := checkPublicCloud(client, &err)
	checkedAndSetVRFConfig := checkAndSetVRFFromNetwork(client, &err)
	isCNIConfigValid := lib.IsValidCni(&err)
	lib.SetIPFamily()
	isValidV6Config := lib.IsValidV6Config(&err)

	isValid := isTenantValid &&
		isCloudValid &&
		isSegroupValid &&
		isRequiredValuesValid &&
		isNodeNetworkValid &&
		isPublicCloudConfigValid &&
		checkedAndSetVRFConfig &&
		isCNIConfigValid &&
		isBGPConfigurationValid &&
		isValidV6Config

	if !isValid {
		if !isCloudValid || !isSegroupValid || !isNodeNetworkValid || !isBGPConfigurationValid {
			utils.AviLog.Warnf("Invalid input detected, AKO will be rebooted to retry %s", err.Error())
			lib.AKOControlConfig().PodEventf(corev1.EventTypeWarning, lib.AKOShutdown, "Invalid user input %s", err.Error())
			lib.ShutdownApi()
		} else {
			utils.AviLog.Warn("Invalid input detected, sync will be disabled.")
		}
	}
	return isValid, err
}

func checkRequiredValuesYaml(returnErr *error) bool {
	if _, err := lib.IsClusterNameValid(); err != nil {
		*returnErr = err
		return false
	}

	lib.SetNamePrefix()
	// after clusterName validation, set AKO User to be used in created_by fields for Avi Objects
	lib.SetAKOUser()
	//Set clusterlabel checksum
	lib.SetClusterLabelChecksum()

	cloudName := utils.CloudName
	if cloudName == "" {
		*returnErr = fmt.Errorf("Required param cloudName not specified, syncing will be disabled")
		return false
	}

	if vipList, err := lib.GetVipNetworkListEnv(); err != nil {
		*returnErr = fmt.Errorf("Error in getting VIP network %s, shutting down AKO", err)
		return false
	} else if len(vipList) > 0 {
		lib.SetVipNetworkList(vipList)
		return true
	}

	// check if config map exists
	k8sClient := utils.GetInformers().ClientSet
	aviCMNamespace := utils.GetAKONamespace()
	if lib.GetNamespaceToSync() != "" {
		aviCMNamespace = lib.GetNamespaceToSync()
	}
	_, err := k8sClient.CoreV1().ConfigMaps(aviCMNamespace).Get(context.TODO(), lib.AviConfigMap, metav1.GetOptions{})
	if err != nil {
		*returnErr = fmt.Errorf("Configmap %s/%s not found, error: %v, syncing will be disabled", aviCMNamespace, lib.AviConfigMap, err)
		return false
	}

	return true
}

// validateAndConfigureSeGroup validates SeGroup configuration provided during installation
// and configures labels on the SeGroup if not present already
func validateAndConfigureSeGroup(client *clients.AviClient, returnErr *error) bool {
	// Note: The Name of SEgroup is being set in the function SetSEGroupCloudName during initialisation

	// Not applicable for NodePort mode / disable route is set as True.
	if lib.GetDisableStaticRoute() {
		utils.AviLog.Infof("Skipping the check for SE group labels ")
		return true
	}

	// List all service engine groups, for every SEG check for AviInfraSettings,
	// if AviInfraSetting NOT found, remove label if exists,
	// if AviInfraSetting found, configure label if doesn't exist.
	// This takes care of syncing SeGroup label settings during reboots.
	seGroupSet := sets.NewString()
	if lib.AKOControlConfig().AviInfraSettingEnabled() {
		infraSettingList, err := lib.AKOControlConfig().CRDClientset().AkoV1alpha1().AviInfraSettings().List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			utils.AviLog.Warnf("Unable to list AviInfraSettings %s", err.Error())
		}
		for _, setting := range infraSettingList.Items {
			seGroupSet.Insert(setting.Spec.SeGroup.Name)
		}
	}
	seGroupSet.Insert(lib.GetSEGName())

	SetAdminTenant := session.SetTenant(lib.GetAdminTenant())
	SetTenant := session.SetTenant(lib.GetTenant())

	// This assumes that a single cluster won't use more than 100 distinct SEGroups.
	uri := "/api/serviceenginegroup/?include_name&page_size=100&cloud_ref.name=" + utils.CloudName + "&name.in=" + strings.Join(seGroupSet.List(), ",")
	var result session.AviCollectionResult
	result, err := lib.AviGetCollectionRaw(client, uri)
	if err != nil {
		if aviError, ok := err.(session.AviError); ok && aviError.HttpStatusCode == 403 {
			//SE in provider context no read access
			utils.AviLog.Debugf("Switching to admin context from  %s", lib.GetTenant())
			SetAdminTenant(client.AviSession)
			defer SetTenant(client.AviSession)
			result, err = lib.AviGetCollectionRaw(client, uri)
			if err != nil {
				*returnErr = fmt.Errorf("Get uri %v returned err %v", uri, err)
				return false
			}

		} else {
			*returnErr = fmt.Errorf("Get uri %v returned err %v", uri, err)
			return false
		}
	}

	elems := make([]json.RawMessage, result.Count)
	err = json.Unmarshal(result.Results, &elems)
	if err != nil {
		*returnErr = fmt.Errorf("Failed to unmarshal data, err: %v", err)
		return false
	}

	for _, elem := range elems {
		seg := models.ServiceEngineGroup{}
		if err := json.Unmarshal(elem, &seg); err != nil {
			utils.AviLog.Warnf("Failed to unmarshal data, err: %v", err)
			continue
		}

		if _, found := seGroupSet[*seg.Name]; found {
			if err = ConfigureSeGroupLabels(client, &seg); err != nil {
				*returnErr = err
				return false
			}
		}
	}

	return true
}

// ConfigureSeGroupLabels configures labels on the SeGroup if not present already
func ConfigureSeGroupLabels(client *clients.AviClient, seGroup *models.ServiceEngineGroup) error {
	labels := seGroup.Labels
	segName := *seGroup.Name
	SetAdminTenant := session.SetTenant(lib.GetAdminTenant())
	SetTenant := session.SetTenant(lib.GetTenant())
	if len(labels) == 0 {
		uri := "/api/serviceenginegroup/" + *seGroup.UUID
		seGroup.Labels = lib.GetLabels()
		response := models.ServiceEngineGroupAPIResponse{}
		err := lib.AviPut(client, uri, seGroup, response)
		if err != nil {
			if aviError, ok := err.(session.AviError); ok && aviError.HttpStatusCode == 400 {
				//SE in provider context
				utils.AviLog.Debugf("Switching to admin context from  %s", lib.GetTenant())
				SetAdminTenant(client.AviSession)
				defer SetTenant(client.AviSession)
				err := lib.AviPut(client, uri, seGroup, response)
				if err != nil {
					return fmt.Errorf("Setting labels on Service Engine Group :%v failed with error :%v. Expected Labels: %v", segName, err.Error(), utils.Stringify(lib.GetLabels()))
				}

			} else {
				return fmt.Errorf("Setting labels on Service Engine Group :%v failed with error :%v. Expected Labels: %v", segName, err.Error(), utils.Stringify(lib.GetLabels()))
			}
		}
		utils.AviLog.Infof("labels: %v set on Service Engine Group :%v", utils.Stringify(lib.GetLabels()), segName)
		return nil
	}

	segLabelEq := reflect.DeepEqual(labels, lib.GetLabels())
	if !segLabelEq {
		return fmt.Errorf("Labels does not match with cluster name for SE group :%v. Expected Labels: %v", segName, utils.Stringify(lib.GetLabels()))
	}

	return nil
}

// DeConfigureSeGroupLabels deconfigures labels on the SeGroup.
func DeConfigureSeGroupLabels() {
	if len(lib.GetLabels()) == 0 {
		return
	}
	segName := lib.GetSEGName()
	clients := SharedAVIClients()
	aviClientLen := lib.GetshardSize()
	var index uint32
	if aviClientLen != 0 {
		index = aviClientLen - 1
	}
	client := clients.AviClient[index]
	SetAdminTenant := session.SetTenant(lib.GetAdminTenant())
	SetTenant := session.SetTenant(lib.GetTenant())
	seGroup, err := GetAviSeGroup(client, segName)
	if err != nil {
		utils.AviLog.Errorf("Failed to get SE group. Error: %v", err)
		return
	}
	clusterLabel := lib.GetLabels()[0]
	// Remove the label from the SEG that belongs to this cluster
	for i, label := range seGroup.Labels {
		if *label.Key == *clusterLabel.Key && *label.Value == *clusterLabel.Value {
			seGroup.Labels = append(seGroup.Labels[:i], seGroup.Labels[i+1:]...)
		}
	}
	utils.AviLog.Infof("Updating the following labels: %v, on the SE Group", utils.Stringify(seGroup.Labels))
	uri := "/api/serviceenginegroup/" + *seGroup.UUID
	response := models.ServiceEngineGroupAPIResponse{}

	err = lib.AviPut(client, uri, seGroup, response)
	if err != nil {
		if aviError, ok := err.(session.AviError); ok && aviError.HttpStatusCode == 400 {
			//SE in provider context
			utils.AviLog.Debugf("Switching to admin context from  %s", lib.GetTenant())
			SetAdminTenant(client.AviSession)
			defer SetTenant(client.AviSession)
			err = lib.AviPut(client, uri, seGroup, response)
			if err != nil {
				utils.AviLog.Warnf("Deconfiguring SE Group labels failed on %v with error %v", segName, err.Error())
				return

			}
		} else {
			utils.AviLog.Warnf("Deconfiguring SE Group labels failed on %v with error %v", segName, err.Error())
			return
		}
	}
	utils.AviLog.Infof("Successfully deconfigured SE Group labels  on %v", segName)
}

func GetAviSeGroup(client *clients.AviClient, segName string) (*models.ServiceEngineGroup, error) {
	SetAdminTenant := session.SetTenant(lib.GetAdminTenant())
	SetTenant := session.SetTenant(lib.GetTenant())
	uri := "/api/serviceenginegroup/?include_name&name=" + segName + "&cloud_ref.name=" + utils.CloudName
	var result session.AviCollectionResult
	result, err := lib.AviGetCollectionRaw(client, uri)
	if err != nil {
		if aviError, ok := err.(session.AviError); ok && aviError.HttpStatusCode == 403 {
			//SE in provider context no read access
			utils.AviLog.Debugf("Switching to admin context from  %s", lib.GetTenant())
			SetAdminTenant(client.AviSession)
			defer SetTenant(client.AviSession)
			result, err = lib.AviGetCollectionRaw(client, uri)
			if err != nil {
				return nil, fmt.Errorf("Get uri %v returned err %v", uri, err)

			}
		} else {
			return nil, fmt.Errorf("Get uri %v returned err %v", uri, err)
		}
	}

	if result.Count != 1 {
		return nil, fmt.Errorf("Service Engine Group details not found with serviceEngineGroupName: %s", segName)
	}

	elems := make([]json.RawMessage, result.Count)
	err = json.Unmarshal(result.Results, &elems)
	if err != nil {
		return nil, fmt.Errorf("Failed to unmarshal data, err: %v", err)
	}

	seg := models.ServiceEngineGroup{}
	err = json.Unmarshal(elems[0], &seg)
	if err != nil {
		return nil, fmt.Errorf("Failed to unmarshal data, err: %v", err)
	}

	if seg.UUID == nil {
		return nil, fmt.Errorf("Failed to get UUID for Service Engine Group: %s", segName)
	}

	return &seg, nil
}

func checkTenant(client *clients.AviClient, returnError *error) bool {
	uri := "/api/tenant/?name=" + lib.GetTenant()
	SetAdminTenant := session.SetTenant(lib.GetAdminTenant())
	SetTenant := session.SetTenant(lib.GetTenant())
	SetAdminTenant(client.AviSession)
	defer SetTenant(client.AviSession)
	result, err := lib.AviGetCollectionRaw(client, uri)
	if err != nil {
		*returnError = fmt.Errorf("Get uri %v returned err %v", uri, err)
		return false
	}

	if result.Count != 1 {
		*returnError = fmt.Errorf("Tenant details not found for the tenant: %s", lib.GetTenant())
		return false
	}

	return true
}

func checkAndSetCloudType(client *clients.AviClient, returnErr *error) bool {
	uri := "/api/cloud/?include_name&name=" + utils.CloudName
	result, err := lib.AviGetCollectionRaw(client, uri)
	if err != nil {
		*returnErr = fmt.Errorf("Get uri %v returned err %v", uri, err)
		return false
	}

	if result.Count != 1 {
		*returnErr = fmt.Errorf("Cloud details not found for cloud name: %s", utils.CloudName)
		return false
	}

	elems := make([]json.RawMessage, result.Count)
	err = json.Unmarshal(result.Results, &elems)
	if err != nil {
		*returnErr = fmt.Errorf("Failed to unmarshal data, err: %v", err)
		return false
	}

	cloud := models.Cloud{}
	err = json.Unmarshal(elems[0], &cloud)
	if err != nil {
		*returnErr = fmt.Errorf("Failed to unmarshal data, err: %v", err)
		return false
	}
	vType := *cloud.Vtype
	tenantRef := *cloud.TenantRef
	lib.SetIsCloudInAdminTenant(strings.HasSuffix(tenantRef, lib.GetAdminTenant()))

	utils.AviLog.Infof("Setting cloud vType: %v", vType)
	lib.SetCloudType(vType)

	utils.AviLog.Infof("Setting cloud uuid: %s", *cloud.UUID)
	lib.SetCloudUUID(*cloud.UUID)

	// IPAM is mandatory for vcenter and noaccess cloud
	if !lib.IsPublicCloud() && cloud.IPAMProviderRef == nil {
		*returnErr = fmt.Errorf("Cloud does not have a ipam_provider_ref configured")
		return false
	}

	// if the cloud's ipamprovider profile is set, check in the ipam
	// whether any of the usable networks have a label set. The marker based approach is not valid
	// for public clouds.
	if ipamCheck, err := checkIPAMForUsableNetworkLabels(client, cloud.IPAMProviderRef); !ipamCheck {
		*returnErr = err
		return false
	}

	// If an NSX-T cloud is configured without a T1LR param, we will disable sync.
	if vType == lib.CLOUD_NSXT && lib.GetNSXTTransportZone() == lib.OVERLAY_TRANSPORT_ZONE {
		if lib.GetT1LRPath() == "" {
			*returnErr = fmt.Errorf("Cloud is configured as NSX-T with overlay transport zone but the T1 LR mapping is not provided")
			return false
		}
	} else if lib.GetT1LRPath() != "" && vType != lib.CLOUD_NSXT {
		// If the cloud type is not NSX-T and yet the T1 LR is set then too disable sync
		*returnErr = fmt.Errorf("Cloud is not configured as NSX-T but the T1 LR mapping is provided")
		return false
	} else if lib.GetT1LRPath() != "" && vType == lib.CLOUD_NSXT && lib.GetNSXTTransportZone() == lib.VLAN_TRANSPORT_ZONE {
		*returnErr = fmt.Errorf("Cloud configured as NSX-T with VLAN transport zone but the T1 LR mapping is  provided")
		return false
	}

	return true
}

func checkIPAMForUsableNetworkLabels(client *clients.AviClient, ipamRefUri *string) (bool, error) {
	// Donot check for labels in usable networks if a vipNetwork is provided by the user.
	// In this case, the vipNetwork provided by the user will be used.
	// 1. Prioritize user input vipetworkList, skip marker based selection if provided.
	// 2. If not provided, check for markers in ipam's usable networks.
	// 3. If marker based usable network is not available, keep vipNetworkList empty.
	// 4. vipNetworkList can be empty only in WCP usecases, for all others, mark invalid configuration.

	// 1. User input
	if vipList, err := lib.GetVipNetworkListEnv(); err != nil {
		return false, fmt.Errorf("Error in getting VIP network %s, shutting down AKO", err)
	} else if len(vipList) > 0 {
		lib.SetVipNetworkList(vipList)
		return true, nil
	}

	// 2. Marker based (only advancedL4)
	var err error
	markerNetworkFound := ""
	if lib.GetAdvancedL4() && ipamRefUri != nil {
		// Using clusterID for advl4.
		ipam := models.IPAMDNSProviderProfile{}
		ipamRef := strings.SplitAfter(*ipamRefUri, "/api/")
		ipamRefWithoutName := strings.Split(ipamRef[1], "#")[0]
		if err := lib.AviGet(client, "/api/"+ipamRefWithoutName+"/?include_name", &ipam); err != nil {
			return false, fmt.Errorf("Get uri %v returned err %v", ipamRef, err)
		}

		usableNetworkNames := []string{}
		for _, usableNetwork := range ipam.InternalProfile.UsableNetworks {
			networkRefName := strings.Split(*usableNetwork.NwRef, "#")
			usableNetworkNames = append(usableNetworkNames, networkRefName[1])
		}

		if len(usableNetworkNames) == 0 {
			return false, fmt.Errorf("No usable network configured in configured cloud's ipam profile.")
		}

		err, markerNetworkFound = fetchNetworkWithMarkerSet(client, usableNetworkNames)
		if err != nil {
			return false, err
		}

		if markerNetworkFound != "" {
			lib.SetVipNetworkList([]akov1alpha1.AviInfraSettingVipNetwork{{
				NetworkName: markerNetworkFound,
			}})
			return true, nil
		}

	}

	// 3. Empty VipNetworkList
	if lib.GetAdvancedL4() && markerNetworkFound == "" {
		lib.SetVipNetworkList([]akov1alpha1.AviInfraSettingVipNetwork{})
		return true, nil
	}

	if utils.IsVCFCluster() {
		vipNetList := akov1alpha1.AviInfraSettingVipNetwork{
			NetworkName: lib.GetVCFNetworkName(),
		}
		lib.SetVipNetworkList([]akov1alpha1.AviInfraSettingVipNetwork{vipNetList})
		return true, nil
	}

	return false, fmt.Errorf("No user input detected for vipNetworkList.")
}

func fetchNetworkWithMarkerSet(client *clients.AviClient, usableNetworkNames []string, overrideUri ...NextPage) (error, string) {
	clusterName := lib.GetClusterID()
	var uri string
	if len(overrideUri) == 1 {
		uri = overrideUri[0].Next_uri
	} else {
		uri = "/api/network/?include_name&page_size=100&name.in=" + strings.Join(usableNetworkNames, ",")
	}

	result, err := lib.AviGetCollectionRaw(client, uri)
	if err != nil {
		utils.AviLog.Errorf("Get uri %v returned err %v", uri, err)
		return err, ""
	}

	elems := make([]json.RawMessage, result.Count)
	err = json.Unmarshal(result.Results, &elems)
	if err != nil {
		utils.AviLog.Errorf("Failed to unmarshal data, err: %v", err)
		return err, ""
	}

	for _, elem := range elems {
		network := models.Network{}
		err = json.Unmarshal(elem, &network)
		if err != nil {
			utils.AviLog.Errorf("Get uri %v returned err %v", uri, err)
			return err, ""
		}

		if len(network.Markers) == 1 &&
			*network.Markers[0].Key == lib.ClusterNameLabelKey &&
			len(network.Markers[0].Values) == 1 &&
			network.Markers[0].Values[0] == clusterName {
			utils.AviLog.Infof("Marker configuration found in usable network. Using %s as vipNetworkList.", *network.Name)
			return nil, *network.Name
		}
	}

	if result.Next != "" {
		// It has a next page, let's recursively call the same method.
		next_uri := strings.Split(result.Next, "/api/network")
		if len(next_uri) > 1 {
			overrideUri := "/api/network" + next_uri[1]
			nextPage := NextPage{Next_uri: overrideUri}
			return fetchNetworkWithMarkerSet(client, usableNetworkNames, nextPage)
		}
	}

	utils.AviLog.Infof("No Marker configured usable networks found.")
	return nil, ""
}

func checkPublicCloud(client *clients.AviClient, returnErr *error) bool {
	if lib.IsPublicCloud() {
		// Handle all public cloud validations here
		vipNetworkList := lib.GetVipNetworkList()
		if len(vipNetworkList) == 0 {
			*returnErr = fmt.Errorf("vipNetworkList not specified, syncing will be disabled.")
			return false
		}
	}
	return true
}

func checkNodeNetwork(client *clients.AviClient, returnErr *error) bool {
	// Not applicable for NodePort mode and non vcenter and nsx-t clouds (overlay)
	if lib.IsNodePortMode() || !lib.IsNodeNetworkAllowedCloud() {
		utils.AviLog.Infof("Skipping the check for Node Network ")
		return true
	}

	// check if node network and cidr's are valid
	nodeNetworkMap, err := lib.GetNodeNetworkMap()
	if err != nil {
		*returnErr = fmt.Errorf("Fetching node network list failed with error: %s, syncing will be disabled.", err.Error())
		return false
	}

	for nodeNetworkName, nodeNetworkCIDRs := range nodeNetworkMap {
		uri := "/api/network/?include_name&name=" + nodeNetworkName + "&cloud_ref.name=" + utils.CloudName
		result, err := lib.AviGetCollectionRaw(client, uri)
		if err != nil {
			*returnErr = fmt.Errorf("Get uri %v returned err %v", uri, err)
			return false
		}
		elems := make([]json.RawMessage, result.Count)
		err = json.Unmarshal(result.Results, &elems)
		if err != nil {
			*returnErr = fmt.Errorf("Failed to unmarshal data, err: %s", err.Error())
			return false
		}

		if result.Count == 0 {
			*returnErr = fmt.Errorf("No networks found for networkName: %s", nodeNetworkName)
			return false
		}

		for _, cidr := range nodeNetworkCIDRs {
			_, _, err := net.ParseCIDR(cidr)
			if err != nil {
				*returnErr = fmt.Errorf("The value of CIDR couldn't be parsed. Failed with error: %v.", err.Error())
				return false
			}
			mask := strings.Split(cidr, "/")[1]
			_, err = strconv.ParseInt(mask, 10, 32)
			if err != nil {
				*returnErr = fmt.Errorf("The value of CIDR couldn't be converted to int32")
				return false
			}
		}
	}
	return true
}

func checkAndSetVRFFromNetwork(client *clients.AviClient, returnErr *error) bool {
	if lib.IsPublicCloud() {
		if lib.GetCloudType() == lib.CLOUD_OPENSTACK {
			if lib.GetTenant() == lib.GetAdminTenant() {
				lib.SetVrf(utils.GlobalVRF)
			} else {
				lib.SetVrf(lib.GetTenant() + "-default")
			}
		} else {
			lib.SetVrf(utils.GlobalVRF)
		}
		return true
	}
	if lib.IsNodePortMode() {
		utils.AviLog.Infof("Using global VRF for NodePort mode")
		lib.SetVrf(utils.GlobalVRF)
		return true
	}

	networkList := lib.GetVipNetworkList()
	if len(networkList) == 0 {
		utils.AviLog.Warnf("Network name not specified, skipping fetching of the VRF setting from network")
		return true
	}

	if !validateNetworkNames(client, networkList) {
		*returnErr = fmt.Errorf("Failed to validate Network Names specified in VIP Network List")
		return false
	}

	networkName := networkList[0].NetworkName

	uri := "/api/network/?include_name&name=" + networkName + "&cloud_ref.name=" + utils.CloudName
	result, err := lib.AviGetCollectionRaw(client, uri)
	if err != nil {
		*returnErr = fmt.Errorf("Get uri %v returned err %v", uri, err)
		return false
	}
	elems := make([]json.RawMessage, result.Count)
	err = json.Unmarshal(result.Results, &elems)
	if err != nil {
		*returnErr = fmt.Errorf("Failed to unmarshal data, err: %v", err)
		return false
	}

	if result.Count == 0 {
		*returnErr = fmt.Errorf("No networks found for networkName: %s", networkName)
		return false
	}

	network := models.Network{}
	err = json.Unmarshal(elems[0], &network)
	if err != nil {
		*returnErr = fmt.Errorf("Failed to unmarshal data, err: %v", err)
		return false
	}

	if lib.GetCloudType() == lib.CLOUD_NSXT &&
		lib.GetServiceType() == "ClusterIP" &&
		lib.GetCNIPlugin() != lib.NCP_CNI &&
		lib.GetNSXTTransportZone() != lib.VLAN_TRANSPORT_ZONE {
		// Here we need to determine the right VRF for this T1LR
		// The logic is: Get all the VRF context objects from the controller, figure out the VRF that matches the T1LR
		// Current pagination size is set to 100, this may have to increased if we have more than 100 T1 routers.
		err, foundVrf := fetchAndSetVrf(client)
		if err != nil {
			*returnErr = err
			return false
		}

		if !foundVrf {
			// Fall back on the `global` VRF if there are no attrs are present.
			vrfRef := *network.VrfContextRef
			vrfName := strings.Split(vrfRef, "#")[1]
			utils.AviLog.Infof("Setting VRF %s from the network because no match found for T1Lr: %s", vrfName, lib.GetT1LRPath())
			lib.SetVrf(vrfName)
		}

	} else {
		vrfRef := *network.VrfContextRef
		vrfName := strings.Split(vrfRef, "#")[1]
		utils.AviLog.Infof("Setting VRF %s found from network %s", vrfName, networkName)
		lib.SetVrf(vrfName)
	}
	return true
}

func fetchAndSetVrf(client *clients.AviClient, overrideUri ...NextPage) (error, bool) {
	var uri string
	if len(overrideUri) == 1 {
		uri = overrideUri[0].Next_uri
	} else {
		uri = "/api/vrfcontext?" + "&include_name=true&cloud_ref.name=" + utils.CloudName + "&page_size=100"
	}

	result, err := lib.AviGetCollectionRaw(client, uri)
	if err != nil {
		utils.AviLog.Warnf("Get uri %v returned err %v", uri, err)
		return err, false
	}
	elems := make([]json.RawMessage, result.Count)
	err = json.Unmarshal(result.Results, &elems)
	if err != nil {
		utils.AviLog.Warnf("Failed to unmarshal data, err: %v", err)
		return err, false
	}

	for i := 0; i < result.Count; i++ {
		vrf := models.VrfContext{}
		err = json.Unmarshal(elems[i], &vrf)
		if err != nil {
			utils.AviLog.Warnf("Failed to unmarshal data, err: %v", err)
			continue
		}

		vrfName := *vrf.Name
		if vrf.Attrs != nil {
			for _, v := range vrf.Attrs {
				if *v.Key == "tier1path" && *v.Value == lib.GetT1LRPath() {
					lib.SetVrf(vrfName)
					utils.AviLog.Infof("Setting VRF %s found that matches the T1Lr %s", vrfName, lib.GetT1LRPath())
					return nil, true
				}
			}
		}
	}

	if result.Next != "" {
		// It has a next page, let's recursively call the same method.
		next_uri := strings.Split(result.Next, "/api/vrfcontext")
		if len(next_uri) > 1 {
			overrideUri := "/api/vrfcontext" + next_uri[1]
			nextPage := NextPage{Next_uri: overrideUri}
			return fetchAndSetVrf(client, nextPage)
		}
	}

	return nil, false
}

func checkBGPParams(returnErr *error) bool {
	enableRhi := lib.GetEnableRHI()
	bgpPeerLabels := lib.GetGlobalBgpPeerLabels()
	if !enableRhi && len(bgpPeerLabels) > 0 {
		*returnErr = fmt.Errorf("BGPPeerLabels %s cannot be set if EnableRhi is set to %v.", utils.Stringify(bgpPeerLabels), enableRhi)
		return false
	}
	return true
}

func validateNetworkNames(client *clients.AviClient, vipNetworkList []akov1alpha1.AviInfraSettingVipNetwork) bool {
	for _, vipNetwork := range vipNetworkList {
		if vipNetwork.Cidr != "" {
			re := regexp.MustCompile(lib.IPCIDRRegex)
			if !re.MatchString(vipNetwork.Cidr) {
				utils.AviLog.Errorf("invalid CIDR configuration %s detected for networkName %s in vipNetworkList", vipNetwork.Cidr, vipNetwork.NetworkName)
				return false
			}
		}
		if vipNetwork.V6Cidr != "" {
			re := regexp.MustCompile(lib.IPV6CIDRRegex)
			if !re.MatchString(vipNetwork.V6Cidr) {
				utils.AviLog.Errorf("invalid IPv6 CIDR configuration %s detected for networkName %s in vipNetworkList", vipNetwork.V6Cidr, vipNetwork.NetworkName)
				return false
			}
		}

		uri := "/api/network/?include_name&name=" + vipNetwork.NetworkName + "&cloud_ref.name=" + utils.CloudName
		result, err := lib.AviGetCollectionRaw(client, uri)
		if err != nil {
			utils.AviLog.Warnf("Get uri %v returned err %v", uri, err)
			return false
		}
		elems := make([]json.RawMessage, result.Count)
		err = json.Unmarshal(result.Results, &elems)
		if err != nil {
			utils.AviLog.Warnf("Failed to unmarshal data, err: %v", err)
			return false
		}

		if result.Count == 0 {
			utils.AviLog.Warnf("No networks found for vipNetwork: %s", vipNetwork.NetworkName)
			return false
		}
	}
	return true
}

func ExtractPattern(word string, pattern string) (string, error) {
	r, err := regexp.Compile(pattern)
	if err != nil {
		return "", err
	}
	result := r.FindAllString(word, -1)
	if len(result) == 1 {
		return result[0][:len(result[0])], nil
	}
	return "", nil
}

func ExtractUuid(word, pattern string) string {
	r, _ := regexp.Compile(pattern)
	result := r.FindAllString(word, -1)
	if len(result) == 1 {
		return result[0][:len(result[0])-1]
	}
	return ""
}

func ExtractUuidWithoutHash(word, pattern string) string {
	r, _ := regexp.Compile(pattern)
	result := r.FindAllString(word, -1)
	if len(result) == 1 {
		return result[0][:len(result[0])]
	}
	return ""
}
