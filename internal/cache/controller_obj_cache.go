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
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	"github.com/avinetworks/sdk/go/clients"
	"github.com/avinetworks/sdk/go/models"
	"github.com/avinetworks/sdk/go/session"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func (c *AviObjCache) AviRefreshObjectCache(client *clients.AviClient, cloud string) {
	c.PopulatePkiProfilesToCache(client)
	c.PopulatePoolsToCache(client, cloud)
	c.PopulatePgDataToCache(client, cloud)
	c.PopulateDSDataToCache(client, cloud)
	c.PopulateSSLKeyToCache(client, cloud)
	c.PopulateHttpPolicySetToCache(client, cloud)
	c.PopulateL4PolicySetToCache(client, cloud)
	c.PopulateVsVipDataToCache(client, cloud)
	// Switch to the below method once the go sdk is fixed for the DBExtensions fields.
	//c.PopulateVsKeyToCache(client, cloud)
}

func (c *AviObjCache) AviCacheRefresh(client *clients.AviClient, cloud string) {
	c.AviCloudPropertiesPopulate(client, cloud)
}

func (c *AviObjCache) AviObjCachePopulate(client *clients.AviClient, version string, cloud string) ([]NamespaceName, []NamespaceName, error) {
	SetTenant := session.SetTenant(lib.GetTenant())
	SetTenant(client.AviSession)
	SetVersion := session.SetVersion(version)
	SetVersion(client.AviSession)
	vsCacheCopy := []NamespaceName{}
	allVsKeys := []NamespaceName{}
	err := c.AviObjVrfCachePopulate(client, cloud)
	if err != nil {
		return vsCacheCopy, allVsKeys, err
	}
	// Populate the VS cache
	utils.AviLog.Infof("Refreshing all object cache")
	c.AviRefreshObjectCache(client, cloud)
	vsCacheCopy = c.VsCacheMeta.AviCacheGetAllParentVSKeys()
	allVsKeys = c.VsCacheMeta.AviGetAllKeys()
	err = c.AviObjVSCachePopulate(client, cloud, &allVsKeys)
	if err != nil {
		return vsCacheCopy, allVsKeys, err
	}
	// Populate the SNI VS keys to their respective parents
	c.PopulateVsMetaCache()
	// Delete all the VS keys that are left in the copy.
	for _, key := range allVsKeys {
		utils.AviLog.Debugf("Removing vs key from cache: %s", key)
		// We want to synthesize these keys to layer 3.
		vsCacheCopy = Remove(vsCacheCopy, key)
		c.VsCacheMeta.AviCacheDelete(key)
	}
	err = c.AviCloudPropertiesPopulate(client, cloud)
	if err != nil {
		return vsCacheCopy, allVsKeys, err
	}
	//vsCacheCopy at this time, is left with only the deleted keys
	return vsCacheCopy, allVsKeys, nil
}

func (c *AviObjCache) PopulateVsMetaCache() {
	// The vh_child_uuids field is used to populate the SNI children during cache population. However, due to the datastore to PG delay - that field may
	// not always be accurate. We would reduce the problem with accuracy by refreshing the SNI cache through reverse mapping sni's to parent
	// Go over the entire VS cache.
	parentVsKeys := c.VsCacheLocal.AviCacheGetAllParentVSKeys()
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
			}
		}
	}
	// Now write lock and copy over all VsCacheMeta and copy the right cache from local
	allVsKeys := c.VsCacheLocal.AviGetAllKeys()
	for _, vsKey := range allVsKeys {
		vsObj, vsFound := c.VsCacheLocal.AviCacheGet(vsKey)
		if vsFound {
			vs_cache_obj, foundvs := vsObj.(*AviVsCache)
			if foundvs {
				c.MarkReference(vs_cache_obj)
				vsCopy, done := vs_cache_obj.GetVSCopy()
				if done {
					c.VsCacheMeta.AviCacheAdd(vsKey, vsCopy)
					c.VsCacheLocal.AviCacheDelete(vsKey)
				}
			}
		}
	}
	c.DeleteUnmarked()
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
func (c *AviObjCache) DeleteUnmarked() {

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
			if obj.HasReference == false {
				utils.AviLog.Infof("Reference Not found for ssl key: %s", objkey)
				sslKeys = append(sslKeys, objkey)
			}
		}
	}

	for _, objkey := range c.VSVIPCache.AviGetAllKeys() {
		intf, _ := c.VSVIPCache.AviCacheGet(objkey)
		if obj, ok := intf.(*AviVSVIPCache); ok {
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
	}
	vsKey := NamespaceName{
		Namespace: lib.GetTenant(),
		Name:      lib.DummyVSForStaleData,
	}
	utils.AviLog.Infof("Dummy VS for stale objects Deletion %s", utils.Stringify(vsMetaObj))
	c.VsCacheMeta.AviCacheAdd(vsKey, &vsMetaObj)

}

func (c *AviObjCache) AviPopulateAllPGs(client *clients.AviClient, cloud string, pgData *[]AviPGCache, override_uri ...NextPage) (*[]AviPGCache, int, error) {
	var uri string
	akoUser := lib.AKOUser

	if len(override_uri) == 1 {
		uri = override_uri[0].Next_uri
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
			override_uri := "/api/poolgroup" + next_uri[1]
			nextPage := NextPage{Next_uri: override_uri}
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

func (c *AviObjCache) AviPopulateAllPkiPRofiles(client *clients.AviClient, pkiData *[]AviPkiProfileCache, override_uri ...NextPage) (*[]AviPkiProfileCache, int, error) {
	var uri string
	akoUser := lib.AKOUser

	if len(override_uri) == 1 {
		uri = override_uri[0].Next_uri
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

		pkiCacheObj := AviPkiProfileCache{
			Name:             *pki.Name,
			Uuid:             *pki.UUID,
			Tenant:           lib.GetTenant(),
			CloudConfigCksum: lib.SSLKeyCertChecksum(*pki.Name, string(*pki.CaCerts[0].Certificate), ""),
		}
		*pkiData = append(*pkiData, pkiCacheObj)

	}
	if result.Next != "" {
		// It has a next page, let's recursively call the same method.
		next_uri := strings.Split(result.Next, "/api/pkiprofile")
		if len(next_uri) > 1 {
			override_uri := "/api/pkiprofile" + next_uri[1]
			nextPage := NextPage{Next_uri: override_uri}
			_, _, err := c.AviPopulateAllPkiPRofiles(client, pkiData, nextPage)
			if err != nil {
				return nil, 0, err
			}
		}
	}

	return pkiData, result.Count, nil
}

func (c *AviObjCache) AviPopulateAllPools(client *clients.AviClient, cloud string, poolData *[]AviPoolCache, override_uri ...NextPage) (*[]AviPoolCache, int, error) {
	var uri string
	akoUser := lib.AKOUser

	if len(override_uri) == 1 {
		uri = override_uri[0].Next_uri
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
		var svc_mdata_obj ServiceMetadataObj
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
			override_uri := "/api/pool" + next_uri[1]
			nextPage := NextPage{Next_uri: override_uri}
			_, _, err := c.AviPopulateAllPools(client, cloud, poolData, nextPage)
			if err != nil {
				return nil, 0, err
			}
		}
	}

	return poolData, result.Count, nil
}

func (c *AviObjCache) PopulatePkiProfilesToCache(client *clients.AviClient, override_uri ...NextPage) {
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

func (c *AviObjCache) PopulatePoolsToCache(client *clients.AviClient, cloud string, override_uri ...NextPage) {
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
		for _, dnsinfo := range vsvip.DNSInfo {
			fqdns = append(fqdns, *dnsinfo.Fqdn)
		}
		var vips []string
		for _, vip := range vsvip.Vip {
			vips = append(vips, *vip.IPAddress.Addr)
		}

		vsVipCacheObj := AviVSVIPCache{
			Name:         *vsvip.Name,
			Uuid:         *vsvip.UUID,
			FQDNs:        fqdns,
			LastModified: *vsvip.LastModified,
			Vips:         vips,
		}
		*vsVipData = append(*vsVipData, vsVipCacheObj)
	}
	if result.Next != "" {
		// It has a next page, let's recursively call the same method.
		next_uri := strings.Split(result.Next, "/api/vsvip")
		if len(next_uri) > 1 {
			override_uri := "/api/vsvip" + next_uri[1]
			nextPage := NextPage{Next_uri: override_uri}
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
		dsCacheObj.CloudConfigCksum = lib.DSChecksum(dsCacheObj.PoolGroups)
		*DsData = append(*DsData, dsCacheObj)
	}
	if result.Next != "" {
		// It has a next page, let's recursively call the same method.
		next_uri := strings.Split(result.Next, "/api/vsdatascriptset")
		if len(next_uri) > 1 {
			override_uri := "/api/vsdatascriptset" + next_uri[1]
			nextPage := NextPage{Next_uri: override_uri}
			_, _, err := c.AviPopulateAllDSs(client, cloud, DsData, nextPage)
			if err != nil {
				return nil, 0, err
			}
		}
	}
	return DsData, result.Count, nil
}

func (c *AviObjCache) PopulateDSDataToCache(client *clients.AviClient, cloud string, override_uri ...NextPage) {
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
		checksum := lib.SSLKeyCertChecksum(*sslkey.Name, *sslkey.Certificate.Certificate, cacert)
		sslCacheObj := AviSSLCache{
			Name:             *sslkey.Name,
			Uuid:             *sslkey.UUID,
			Cert:             *sslkey.Certificate.Certificate,
			HasCARef:         hasCA,
			CACertUUID:       cacertUUID,
			CloudConfigCksum: checksum,
		}
		*SslData = append(*SslData, sslCacheObj)
	}
	if result.Next != "" {
		// It has a next page, let's recursively call the same method.
		next_uri := strings.Split(result.Next, "/api/sslkeyandcertificate")
		if len(next_uri) > 1 {
			override_uri := "/api/sslkeyandcertificate" + next_uri[1]
			nextPage := NextPage{Next_uri: override_uri}
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
		checksum := lib.SSLKeyCertChecksum(*sslkey.Name, *sslkey.Certificate.Certificate, cacert)
		sslCacheObj := AviSSLCache{
			Name:             *sslkey.Name,
			Uuid:             *sslkey.UUID,
			CloudConfigCksum: checksum,
			HasCARef:         hasCA,
		}
		k := NamespaceName{Namespace: lib.GetTenant(), Name: *sslkey.Name}
		c.SSLKeyCache.AviCacheAdd(k, &sslCacheObj)
		utils.AviLog.Debugf("Adding sslkey to Cache during refresh %s\n", k)
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
		checksum := lib.SSLKeyCertChecksum(*pkikey.Name, *pkikey.CaCerts[0].Certificate, "")
		sslCacheObj := AviSSLCache{
			Name:             *pkikey.Name,
			Uuid:             *pkikey.UUID,
			CloudConfigCksum: checksum,
		}
		k := NamespaceName{Namespace: lib.GetTenant(), Name: *pkikey.Name}
		c.SSLKeyCache.AviCacheAdd(k, &sslCacheObj)
		utils.AviLog.Debugf("Adding pkikey to Cache during refresh %s\n", k)
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
		var svc_mdata_obj ServiceMetadataObj
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
		utils.AviLog.Debugf("Adding pool to Cache during refresh %s\n", k)
	}
	return nil
}

func (c *AviObjCache) AviPopulateOneVsDSCache(client *clients.AviClient,
	cloud string, objName string) error {
	var uri string
	akoUser := lib.AKOUser

	uri = "/api/vsdatascript?name=" + objName + "&created_by=" + akoUser

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
		dsCacheObj.CloudConfigCksum = lib.DSChecksum(dsCacheObj.PoolGroups)
		k := NamespaceName{Namespace: lib.GetTenant(), Name: *ds.Name}
		c.DSCache.AviCacheAdd(k, &dsCacheObj)
		utils.AviLog.Debugf("Adding ds to Cache during refresh %s\n", k)
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
		utils.AviLog.Debugf("Adding pg to Cache during refresh %s\n", k)
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
		for _, dnsinfo := range vsvip.DNSInfo {
			fqdns = append(fqdns, *dnsinfo.Fqdn)
		}

		var vips []string
		for _, vip := range vsvip.Vip {
			vips = append(vips, *vip.IPAddress.Addr)
		}

		vsVipCacheObj := AviVSVIPCache{
			Name:         *vsvip.Name,
			Uuid:         *vsvip.UUID,
			FQDNs:        fqdns,
			LastModified: *vsvip.LastModified,
			Vips:         vips,
		}
		k := NamespaceName{Namespace: lib.GetTenant(), Name: *vsvip.Name}
		c.VSVIPCache.AviCacheAdd(k, &vsVipCacheObj)
		utils.AviLog.Debugf("Adding vsvip to Cache during refresh %s\n", k)
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
		if httppol.HTTPRequestPolicy != nil {
			for _, rule := range httppol.HTTPRequestPolicy.Rules {
				if rule.SwitchingAction != nil {
					pgUuid := ExtractUuid(*rule.SwitchingAction.PoolGroupRef, "poolgroup-.*.#")
					pgName, found := c.PgCache.AviCacheGetNameByUuid(pgUuid)
					if found {
						poolGroups = append(poolGroups, pgName.(string))
					}
				}
			}
		}
		httpPolCacheObj := AviHTTPPolicyCache{
			Name:             *httppol.Name,
			Uuid:             *httppol.UUID,
			CloudConfigCksum: *httppol.CloudConfigCksum,
			PoolGroups:       poolGroups,
			LastModified:     *httppol.LastModified,
		}
		k := NamespaceName{Namespace: lib.GetTenant(), Name: *httppol.Name}
		c.HTTPPolicyCache.AviCacheAdd(k, &httpPolCacheObj)
		utils.AviLog.Debugf("Adding httppolicy to Cache during refresh %s\n", k)
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
		var protocol string
		if l4pol.L4ConnectionPolicy != nil {
			for _, rule := range l4pol.L4ConnectionPolicy.Rules {
				protocol = *rule.Match.Protocol.Protocol
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
		l4PolCacheObj := AviL4PolicyCache{
			Name:             *l4pol.Name,
			Uuid:             *l4pol.UUID,
			Pools:            pools,
			LastModified:     *l4pol.LastModified,
			CloudConfigCksum: lib.L4PolicyChecksum(ports, protocol),
		}
		k := NamespaceName{Namespace: lib.GetTenant(), Name: *l4pol.Name}
		c.L4PolicyCache.AviCacheAdd(k, &l4PolCacheObj)
		utils.AviLog.Infof("Adding l4pol to Cache during refresh %s\n", lib.L4PolicyChecksum(ports, protocol))
	}
	return nil
}

// This method is just added for future here. We can't use it until they expose the DB extensions on the virtualservice object
func (c *AviObjCache) AviPopulateAllVSMeta(client *clients.AviClient, cloud string, vsData *[]AviVsCache, nextPage ...NextPage) (*[]AviVsCache, error) {
	var uri string
	akoUser := lib.AKOUser

	if len(nextPage) == 1 {
		uri = nextPage[0].Next_uri
	} else {
		uri = "/api/virtualservice/?" + "include_name=true" + "&cloud_ref.name=" + cloud + "&created_by=" + akoUser + "&page_size=100"
	}

	result, err := lib.AviGetCollectionRaw(client, uri)
	if err != nil {
		utils.AviLog.Warnf("Get uri %v returned err for virtualservice %v", uri, err)
		return nil, err
	}
	elems := make([]json.RawMessage, result.Count)
	err = json.Unmarshal(result.Results, &elems)
	if err != nil {
		utils.AviLog.Warnf("Failed to unmarshal virtualservice data, err: %v", err)
		return nil, err
	}
	for i := 0; i < len(elems); i++ {
		vsModel := models.VirtualService{}
		err = json.Unmarshal(elems[i], &vsModel)
		if err != nil {
			utils.AviLog.Warnf("Failed to unmarshal virtualservice data, err: %v", err)
			continue
		}
		if vsModel.Name == nil || vsModel.UUID == nil {
			utils.AviLog.Warnf("Incomplete sslkey data unmarshalled, %s", utils.Stringify(vsModel))
			continue
		}
		var vsVipKey []NamespaceName
		var sslKeys []NamespaceName
		var dsKeys []NamespaceName
		var httpKeys []NamespaceName
		var poolgroupKeys []NamespaceName
		var poolKeys []NamespaceName
		var virtualIp string
		if vsModel.Vip != nil {
			for _, vip := range vsModel.Vip {
				virtualIp = *vip.IPAddress.Addr
			}
		}
		if vsModel.VsvipRef != nil {
			// find the vsvip name from the vsvip cache
			vsVipUuid := ExtractUuid(*vsModel.VsvipRef, "vsvip-.*.#")
			vsVipName, foundVip := c.VSVIPCache.AviCacheGetNameByUuid(vsVipUuid)
			if foundVip {
				vipKey := NamespaceName{Namespace: lib.GetTenant(), Name: vsVipName.(string)}
				vsVipKey = append(vsVipKey, vipKey)
			}
		}
		if vsModel.SslKeyAndCertificateRefs != nil {
			for _, ssl := range vsModel.SslKeyAndCertificateRefs {
				// find the sslkey name from the ssl key cache
				sslUuid := ExtractUuid(ssl, "sslkeyandcertificate-.*.#")
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
		if vsModel.VsDatascripts != nil {
			for _, dataScript := range vsModel.VsDatascripts {
				dsUuid := ExtractUuid(*dataScript.VsDatascriptSetRef, "vsdatascriptset-.*.#")

				dsName, foundDs := c.DSCache.AviCacheGetNameByUuid(dsUuid)
				if foundDs {
					dsKey := NamespaceName{Namespace: lib.GetTenant(), Name: dsName.(string)}
					// Fetch the associated PGs with the DS.
					dsObj, _ := c.DSCache.AviCacheGet(dsKey)
					for _, pgName := range dsObj.(*AviDSCache).PoolGroups {
						// For each PG, formulate the key and then populate the pg collection cache
						pgKey := NamespaceName{Namespace: lib.GetTenant(), Name: pgName}
						poolgroupKeys = append(poolgroupKeys, pgKey)
						poolKeys = c.AviPGPoolCachePopulate(client, cloud, pgName)
					}
					dsKeys = append(dsKeys, dsKey)
				}

			}
		}
		// Handle L4 vs - pg references
		if vsModel.ServicePoolSelect != nil {
			for _, pg_intf := range vsModel.ServicePoolSelect {
				pgUuid := ExtractUuid(*pg_intf.ServicePoolGroupRef, "poolgroup-.*.#")

				pgName, foundpg := c.PgCache.AviCacheGetNameByUuid(pgUuid)
				if foundpg {
					pgKey := NamespaceName{Namespace: lib.GetTenant(), Name: pgName.(string)}
					poolgroupKeys = append(poolgroupKeys, pgKey)
					poolKeys = c.AviPGPoolCachePopulate(client, cloud, pgName.(string))
				}

			}
		}
		if vsModel.HTTPPolicies != nil {
			for _, http_intf := range vsModel.HTTPPolicies {

				httpUuid := ExtractUuid(*http_intf.HTTPPolicySetRef, "httppolicyset-.*.#")

				httpName, foundhttp := c.HTTPPolicyCache.AviCacheGetNameByUuid(httpUuid)
				if foundhttp {
					httpKey := NamespaceName{Namespace: lib.GetTenant(), Name: httpName.(string)}
					httpObj, _ := c.HTTPPolicyCache.AviCacheGet(httpKey)
					for _, pgName := range httpObj.(*AviHTTPPolicyCache).PoolGroups {
						// For each PG, formulate the key and then populate the pg collection cache
						pgKey := NamespaceName{Namespace: lib.GetTenant(), Name: pgName}
						poolgroupKeys = append(poolgroupKeys, pgKey)
						poolKeys = c.AviPGPoolCachePopulate(client, cloud, pgName)
					}
					httpKeys = append(httpKeys, httpKey)
				}

			}
		}
		vsMetaObj := AviVsCache{
			Name:                 *vsModel.Name,
			Uuid:                 *vsModel.UUID,
			VSVipKeyCollection:   vsVipKey,
			HTTPKeyCollection:    httpKeys,
			DSKeyCollection:      dsKeys,
			SSLKeyCertCollection: sslKeys,
			PGKeyCollection:      poolgroupKeys,
			PoolKeyCollection:    poolKeys,
			Vip:                  virtualIp,
			CloudConfigCksum:     *vsModel.CloudConfigCksum,
		}
		*vsData = append(*vsData, vsMetaObj)
	}
	if result.Next != "" {
		// It has a next page, let's recursively call the same method.
		next_uri := strings.Split(result.Next, "/api/virtualservice")
		if len(next_uri) > 1 {
			override_uri := "/api/virtualservice" + next_uri[1]
			nextPage := NextPage{Next_uri: override_uri}
			_, err := c.AviPopulateAllVSMeta(client, cloud, vsData, nextPage)
			if err != nil {
				return nil, err
			}
		}
	}
	return vsData, nil
}

// This method is just added for future here. We can't use it until they expose the DB extensions on the virtualservice object
func (c *AviObjCache) PopulateVsKeyToCache(client *clients.AviClient, cloud string, override_uri ...NextPage) {
	var vsCacheMeta []AviVsCache
	_, err := c.AviPopulateAllVSMeta(client, cloud, &vsCacheMeta)
	if err != nil {
		return
	}
}

func (c *AviObjCache) PopulateSSLKeyToCache(client *clients.AviClient, cloud string, override_uri ...NextPage) {
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
		var cacert string
		// Find CA Cert name from the cache for checksum calculation.
		if SslKeyData[i].HasCARef {
			ca, found := c.SSLKeyCache.AviCacheGetNameByUuid(SslKeyCacheObj.CACertUUID)
			if !found {
				utils.AviLog.Warnf("cacertUUID %s for keycert %s not found in cache", SslKeyCacheObj.CACertUUID, SslKeyCacheObj.Name)
			} else {
				cacert = ca.(string)
			}
		}
		SslKeyData[i].CloudConfigCksum = lib.SSLKeyCertChecksum(SslKeyCacheObj.Name, SslKeyCacheObj.Cert, cacert)
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
		if httppol.HTTPRequestPolicy != nil {
			for _, rule := range httppol.HTTPRequestPolicy.Rules {
				if rule.SwitchingAction != nil {
					pgUuid := ExtractUuid(*rule.SwitchingAction.PoolGroupRef, "poolgroup-.*.#")
					pgName, found := c.PgCache.AviCacheGetNameByUuid(pgUuid)
					if found {
						poolGroups = append(poolGroups, pgName.(string))
					}
				}
			}
		}
		httpPolCacheObj := AviHTTPPolicyCache{
			Name:             *httppol.Name,
			Uuid:             *httppol.UUID,
			CloudConfigCksum: *httppol.CloudConfigCksum,
			PoolGroups:       poolGroups,
			LastModified:     *httppol.LastModified,
		}
		*httpPolicyData = append(*httpPolicyData, httpPolCacheObj)

	}
	if result.Next != "" {
		// It has a next page, let's recursively call the same method.
		next_uri := strings.Split(result.Next, "/api/httppolicyset")
		if len(next_uri) > 1 {
			override_uri := "/api/httppolicyset" + next_uri[1]
			nextPage := NextPage{Next_uri: override_uri}
			_, _, err := c.AviPopulateAllHttpPolicySets(client, cloud, httpPolicyData, nextPage)
			if err != nil {
				return nil, 0, err
			}
		}
	}
	return httpPolicyData, result.Count, nil
}

func (c *AviObjCache) PopulateHttpPolicySetToCache(client *clients.AviClient, cloud string, override_uri ...NextPage) {
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
		var protocol string
		if l4pol.L4ConnectionPolicy != nil {
			for _, rule := range l4pol.L4ConnectionPolicy.Rules {
				if rule.Action != nil {
					protocol = *rule.Match.Protocol.Protocol
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
		if strings.Contains(protocol, utils.TCP) {
			protocol = utils.TCP
		} else {
			protocol = utils.UDP
		}
		l4PolCacheObj := AviL4PolicyCache{
			Name:             *l4pol.Name,
			Uuid:             *l4pol.UUID,
			Pools:            pools,
			LastModified:     *l4pol.LastModified,
			CloudConfigCksum: lib.L4PolicyChecksum(ports, protocol),
		}

		*l4PolicyData = append(*l4PolicyData, l4PolCacheObj)
	}

	if result.Next != "" {
		// It has a next page, let's recursively call the same method.
		next_uri := strings.Split(result.Next, "/api/l4policyset")
		if len(next_uri) > 1 {
			override_uri := "/api/l4policyset" + next_uri[1]
			nextPage := NextPage{Next_uri: override_uri}
			_, _, err := c.AviPopulateAllL4PolicySets(client, cloud, l4PolicyData, nextPage)
			if err != nil {
				return nil, 0, err
			}
		}
	}
	return l4PolicyData, result.Count, nil
}

func (c *AviObjCache) PopulateL4PolicySetToCache(client *clients.AviClient, cloud string, override_uri ...NextPage) {
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
	uri := "/api/vrfcontext?name=" + lib.GetVrf() + "&include_name=true&cloud_ref.name=" + cloud
	vrfList := []*models.VrfContext{}

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
		vrfList = append(vrfList, &vrf)

		vrfName := *vrf.Name
		checksum := lib.VrfChecksum(vrfName, vrf.StaticRoutes)
		vrfCacheObj := AviVrfCache{
			Name:             vrfName,
			Uuid:             *vrf.UUID,
			CloudConfigCksum: checksum,
		}
		// set the vrf context. The result shouldn't be more than 1.
		lib.SetVrfUuid(*vrf.UUID)
		utils.AviLog.Debugf("Adding vrf to Cache %s\n", vrfName)
		c.VrfCache.AviCacheAdd(vrfName, &vrfCacheObj)
	}
	return nil
}

func (c *AviObjCache) AviObjVSCachePopulate(client *clients.AviClient, cloud string, vsCacheCopy *[]NamespaceName, override_uri ...NextPage) error {
	var rest_response interface{}
	akoUser := lib.AKOUser
	var uri string
	httpCacheRefreshCount := 1 // Refresh count for http cache is attempted once per page
	if len(override_uri) == 1 {
		uri = override_uri[0].Next_uri
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
			var svc_mdata_obj ServiceMetadataObj
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
				*vsCacheCopy = Remove(*vsCacheCopy, k)
				var vip string
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
								if len(vsVipData.Vips) > 0 {
									vip = vsVipData.Vips[0]
								}
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
					Vip:                  vip,
					CloudConfigCksum:     vs["cloud_config_cksum"].(string),
					SNIChildCollection:   sni_child_collection,
					ParentVSRef:          parentVSKey,
					ServiceMetadataObj:   svc_mdata_obj,
					L4PolicyCollection:   l4Keys,
					LastModified:         vs["_last_modified"].(string),
				}
				c.VsCacheLocal.AviCacheAdd(k, &vsMetaObj)
				utils.AviLog.Debugf("Added VS cache key :%s", utils.Stringify(vsMetaObj))

			}
		}
		if resp["next"] != nil {
			// It has a next page, let's recursively call the same method.
			next_uri := strings.Split(resp["next"].(string), "/api/virtualservice")
			utils.AviLog.Debugf("Found next page for vs, uri: %s", next_uri)
			if len(next_uri) > 1 {
				override_uri := "/api/virtualservice" + next_uri[1]
				utils.AviLog.Debugf("Next page uri for vs: %s", override_uri)
				nextPage := NextPage{Next_uri: override_uri}
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
			vs, ok := vs_intf.(map[string]interface{})
			svc_mdata_intf, ok := vs["service_metadata"]
			var svc_mdata_obj ServiceMetadataObj
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
				var vip string
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
							if len(vsVipData.Vips) > 0 {
								vip = vsVipData.Vips[0]
							}
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
					Vip:                  vip,
					CloudConfigCksum:     vs["cloud_config_cksum"].(string),
					SNIChildCollection:   sni_child_collection,
					ParentVSRef:          parentVSKey,
					L4PolicyCollection:   l4Keys,
					ServiceMetadataObj:   svc_mdata_obj,
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
				utils.AviLog.Debugf("Added VS during refresh with cache key :%v", utils.Stringify(vsMetaObj))
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
			if runtimeCache != nil &&
				(runtimeCache.UpSince != nodeObj["up_since"].(string) || runtimeCache.Name != nodeObj["name"].(string)) {
				// reboot AKO
				utils.AviLog.Warnf("Avi controller leader node or leader uptime changed, shutting down AKO")
				lib.ShutdownApi()
				return nil
			}

			setCacheVal := &AviClusterRuntimeCache{
				Name:    nodeObj["name"].(string),
				UpSince: nodeObj["up_since"].(string),
			}
			c.ClusterStatusCache.AviCacheAdd(lib.ClusterStatusCacheKey, setCacheVal)
			utils.AviLog.Infof("Added ClusterStatusCache cache key %v val %v", lib.ClusterStatusCacheKey, setCacheVal)
			break
		}
	}

	return nil
}

func (c *AviObjCache) AviCloudPropertiesPopulate(client *clients.AviClient, cloudName string) error {
	uri := "/api/cloud/?name=" + cloudName
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
	cloud_obj := &AviCloudPropertyCache{Name: cloudName, VType: vtype}

	subdomains := c.AviDNSPropertyPopulate(client, *cloud.UUID)
	if len(subdomains) == 0 {
		utils.AviLog.Warnf("Cloud: %v does not have a dns provider configured", cloudName)
		return nil
	}

	if subdomains != nil {
		cloud_obj.NSIpamDNS = subdomains
	}

	c.CloudKeyCache.AviCacheAdd(cloudName, cloud_obj)
	utils.AviLog.Infof("Added CloudKeyCache cache key %v val %v", cloudName, cloud_obj)
	return nil
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

func ValidateUserInput(client *clients.AviClient) bool {
	// add other step0 validation logics here -> isValid := check1 && check2 && ...

	isTenantValid := checkTenant(client)
	isCloudValid := checkAndSetCloudType(client)
	isRequiredValuesValid := checkRequiredValuesYaml()
	if lib.GetAdvancedL4() && isTenantValid && isCloudValid && isRequiredValuesValid {
		utils.AviLog.Info("All values verified for advanced L4, proceeding with bootup")
		return true
	}

	isSegroupValid := isCloudValid && checkSegroupLabels(client)
	isNodeNetworkValid := isCloudValid && checkNodeNetwork(client)
	isValid := isTenantValid &&
		isCloudValid &&
		isSegroupValid &&
		isNodeNetworkValid &&
		checkPublicCloud(client) &&
		isRequiredValuesValid &&
		checkAndSetVRFFromNetwork(client)

	if !isValid {
		if !isCloudValid || !isSegroupValid || !isNodeNetworkValid {
			utils.AviLog.Warn("Invalid input detected, AKO will be rebooted to retry")
			lib.ShutdownApi()
		}
		utils.AviLog.Warn("Invalid input detected, sync will be disabled.")
	}
	return isValid
}

func checkRequiredValuesYaml() bool {
	if !lib.IsClusterNameValid() {
		return false
	}
	lib.SetNamePrefix()

	// after clusterName validation, set AKO User to be used in created_by fields for Avi Objects
	lib.SetAKOUser()

	cloudName := os.Getenv("CLOUD_NAME")
	if cloudName == "" {
		utils.AviLog.Error("Required param cloudName not specified, syncing will be disabled")
		return false
	}

	// check if config map exists
	k8sClient := utils.GetInformers().ClientSet
	aviCMNamespace := utils.GetAKONamespace()
	if lib.GetNamespaceToSync() != "" {
		aviCMNamespace = lib.GetNamespaceToSync()
	}
	_, err := k8sClient.CoreV1().ConfigMaps(aviCMNamespace).Get(context.TODO(), lib.AviConfigMap, metav1.GetOptions{})
	if err != nil {
		utils.AviLog.Errorf("Configmap %s/%s not found, error: %v, syncing will be disabled", aviCMNamespace, lib.AviConfigMap, err)
		return false
	}

	return true
}

func checkSegroupLabels(client *clients.AviClient) bool {

	// Not applicable for NodePort mode / disable route is set as True
	if lib.GetDisableStaticRoute() {
		utils.AviLog.Infof("Skipping the check for SE group labels ")
		return true
	}
	// validate SE Group labels
	segName := lib.GetSEGName()
	if segName == "" {
		utils.AviLog.Warnf("Service Engine Group: serviceEngineGroupName not set in values.yaml, skipping sync.")
		return false
	}
	uri := "/api/serviceenginegroup/?include_name&name=" + segName + "&cloud_ref.name=" + utils.CloudName
	result, err := lib.AviGetCollectionRaw(client, uri)
	if err != nil {
		utils.AviLog.Warnf("Get uri %v returned err %v", uri, err)
		return false
	}

	if result.Count != 1 {
		utils.AviLog.Warnf("Service Engine Group details not found with serviceEngineGroupName: %s", segName)
		return false
	}

	elems := make([]json.RawMessage, result.Count)
	err = json.Unmarshal(result.Results, &elems)
	if err != nil {
		utils.AviLog.Warnf("Failed to unmarshal data, err: %v", err)
		return false
	}

	seg := models.ServiceEngineGroup{}
	err = json.Unmarshal(elems[0], &seg)
	if err != nil {
		utils.AviLog.Warnf("Failed to unmarshal data, err: %v", err)
		return false
	}

	if seg.UUID == nil {
		utils.AviLog.Warnf("Failed to get UUID for Service Engine Group: %s", segName)
		return false
	}

	labels := seg.Labels
	if len(labels) == 0 {
		uri = "/api/serviceenginegroup/" + *seg.UUID
		seg.Labels = lib.GetLabels()
		response := models.ServiceEngineGroupAPIResponse{}
		// If tenants per cluster is enabled then the X-Avi-Tenant needs to be set to admin for vrfcontext and segroup updates
		if lib.GetTenantsPerCluster() && lib.IsCloudInAdminTenant {
			SetAdminTenant := session.SetTenant(lib.GetAdminTenant())
			SetTenant := session.SetTenant(lib.GetTenant())
			SetAdminTenant(client.AviSession)
			defer SetTenant(client.AviSession)
		}

		err = lib.AviPut(client, uri, seg, response)
		if err != nil {
			utils.AviLog.Warnf("Setting labels on Service Engine Group :%v failed with error :%v. Expected Labels: %v", segName, err.Error(), utils.Stringify(lib.GetLabels()))
			return false
		}
		utils.AviLog.Infof("labels: %v set on Service Engine Group :%v", utils.Stringify(lib.GetLabels()), segName)
		return true

	}

	segLabelEq := reflect.DeepEqual(labels, lib.GetLabels())
	if !segLabelEq {
		utils.AviLog.Warnf("Labels does not match with cluster name for SE group :%v. Expected Labels: %v", segName, utils.Stringify(lib.GetLabels()))
		return false
	}

	return true
}

func checkTenant(client *clients.AviClient) bool {

	uri := "/api/tenant/?name=" + lib.GetTenant()
	if lib.GetTenantsPerCluster() {
		SetAdminTenant := session.SetTenant(lib.GetAdminTenant())
		SetTenant := session.SetTenant(lib.GetTenant())
		SetAdminTenant(client.AviSession)
		defer SetTenant(client.AviSession)
	}
	result, err := lib.AviGetCollectionRaw(client, uri)
	if err != nil {
		utils.AviLog.Errorf("Get uri %v returned err %v", uri, err)
		return false
	}

	if result.Count != 1 {
		utils.AviLog.Errorf("Tenant details not found for the tenant: %s", lib.GetTenant())
		return false
	}
	return true
}

func checkAndSetCloudType(client *clients.AviClient) bool {

	uri := "/api/cloud/?include_name&name=" + utils.CloudName
	result, err := lib.AviGetCollectionRaw(client, uri)
	if err != nil {
		utils.AviLog.Errorf("Get uri %v returned err %v", uri, err)
		return false
	}

	if result.Count != 1 {
		utils.AviLog.Errorf("Cloud details not found for cloud name: %s", utils.CloudName)
		return false
	}

	elems := make([]json.RawMessage, result.Count)
	err = json.Unmarshal(result.Results, &elems)
	if err != nil {
		utils.AviLog.Errorf("Failed to unmarshal data, err: %v", err)
		return false
	}

	cloud := models.Cloud{}
	err = json.Unmarshal(elems[0], &cloud)
	if err != nil {
		utils.AviLog.Errorf("Failed to unmarshal data, err: %v", err)
		return false
	}
	vType := *cloud.Vtype
	tenantRef := *cloud.TenantRef
	lib.SetIsCloudInAdminTenant(strings.HasSuffix(tenantRef, lib.GetAdminTenant()))

	utils.AviLog.Infof("Setting cloud vType: %v", vType)
	lib.SetCloudType(vType)

	// IPAM is mandatory for vcenter and noaccess cloud
	if !lib.IsPublicCloud() && cloud.IPAMProviderRef == nil {
		utils.AviLog.Errorf("Cloud does not have a ipam_provider_ref configured")
		return false
	}

	return true
}

func checkPublicCloud(client *clients.AviClient) bool {
	if lib.IsPublicCloud() {
		// Handle all public cloud validations here
		networkName := lib.GetNetworkName()
		if networkName == "" && lib.GetCloudType() != lib.CLOUD_GCP {
			// networkName is required param for AWS and Azure Clouds
			utils.AviLog.Errorf("Required param networkName not specified, syncing will be disabled.")
			return false
		}
	}

	return true
}

func checkNodeNetwork(client *clients.AviClient) bool {

	// Not applicable for NodePort mode and non vcenter clouds
	if lib.IsNodePortMode() || lib.GetCloudType() != lib.CLOUD_VCENTER {
		utils.AviLog.Infof("Skipping the check for Node Network ")
		return true
	}

	// check if node network and cidr's are valid
	nodeNetworkMap, err := lib.GetNodeNetworkMap()
	if err != nil {
		utils.AviLog.Errorf("Fetching node network list failed with error: %s, syncing will be disabled.", err.Error())
		return false
	}

	for nodeNetworkName, nodeNetworkCIDRs := range nodeNetworkMap {

		uri := "/api/network/?include_name&name=" + nodeNetworkName + "&cloud_ref.name=" + utils.CloudName
		result, err := lib.AviGetCollectionRaw(client, uri)
		if err != nil {
			utils.AviLog.Errorf("Get uri %v returned err %v", uri, err)
			return false
		}
		elems := make([]json.RawMessage, result.Count)
		err = json.Unmarshal(result.Results, &elems)
		if err != nil {
			utils.AviLog.Warnf("Failed to unmarshal data, err: %s", err.Error())
			return false
		}

		if result.Count == 0 {
			utils.AviLog.Errorf("No networks found for networkName: %s", nodeNetworkName)
			return false
		}

		for _, cidr := range nodeNetworkCIDRs {
			_, _, err := net.ParseCIDR(cidr)
			if err != nil {
				utils.AviLog.Errorf("The value of CIDR couldn't be parsed. Failed with error: %v.", err.Error())
				return false
			}
			mask := strings.Split(cidr, "/")[1]
			_, err = strconv.ParseInt(mask, 10, 32)
			if err != nil {
				utils.AviLog.Errorf("The value of CIDR couldn't be converted to int32")
				return false
			}
		}
	}

	return true
}

func checkAndSetVRFFromNetwork(client *clients.AviClient) bool {

	if lib.IsPublicCloud() {
		// Need not set VRFContext for public clouds.
		return true
	}

	networkName := lib.GetNetworkName()
	if networkName == "" {
		utils.AviLog.Warnf("Param networkName not specified, skipping fetching of the VRF setting from network")
		return true
	}

	uri := "/api/network/?include_name&name=" + networkName + "&cloud_ref.name=" + utils.CloudName
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
		utils.AviLog.Warnf("No networks found for networkName: %s", networkName)
		return false
	}

	network := models.Network{}
	err = json.Unmarshal(elems[0], &network)
	if err != nil {
		utils.AviLog.Warnf("Failed to unmarshal data, err: %v", err)
		return false
	}

	if lib.IsNodePortMode() {
		utils.AviLog.Infof("Using global VRF for NodePort mode")
		return true
	}

	vrfRef := *network.VrfContextRef
	vrfName := strings.Split(vrfRef, "#")[1]
	utils.AviLog.Infof("Setting VRF %s found from network %s", vrfName, networkName)
	lib.SetVrf(vrfName)
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
