/*
 * [2013] - [2018] Avi Networks Incorporated
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
	"encoding/json"
	"errors"
	"os"
	"regexp"
	"strings"
	"sync"

	"ako/pkg/lib"

	"github.com/avinetworks/container-lib/utils"
	"github.com/avinetworks/sdk/go/clients"
	"github.com/avinetworks/sdk/go/models"
	"github.com/avinetworks/sdk/go/session"
)

type AviObjCache struct {
	VsCache         *AviCache
	PgCache         *AviCache
	DSCache         *AviCache
	PoolCache       *AviCache
	CloudKeyCache   *AviCache
	HTTPPolicyCache *AviCache
	SSLKeyCache     *AviCache
	VSVIPCache      *AviCache
	VrfCache        *AviCache
	VsCacheMeta     *AviCache
	VsCacheLocal    *AviCache
}

func NewAviObjCache() *AviObjCache {
	c := AviObjCache{}
	c.VsCacheMeta = NewAviCache()
	c.VsCacheLocal = NewAviCache()
	c.VsCache = NewAviCache()
	c.PgCache = NewAviCache()
	c.DSCache = NewAviCache()
	c.PoolCache = NewAviCache()
	c.SSLKeyCache = NewAviCache()
	c.CloudKeyCache = NewAviCache()
	c.HTTPPolicyCache = NewAviCache()
	c.VSVIPCache = NewAviCache()
	c.VrfCache = NewAviCache()
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

func VrfChecksum(vrfName string, staticRoutes []*models.StaticRoute) uint32 {
	return (utils.Hash(vrfName) + utils.Hash(utils.Stringify(staticRoutes)))
}

func (c *AviObjCache) AviRefreshObjectCache(client *clients.AviClient,
	cloud string) {
	c.PopulatePoolsToCache(client, cloud)
	c.PopulatePgDataToCache(client, cloud)
	c.PopulateDSDataToCache(client, cloud)
	c.PopulateSSLKeyToCache(client, cloud)
	c.PopulateHttpPolicySetToCache(client, cloud)
	c.PopulateVsVipDataToCache(client, cloud)
	// Switch to the below method once the go sdk is fixed for the DBExtensions fields.
	//c.PopulateVsKeyToCache(client, cloud)
}

func (c *AviObjCache) AviObjCachePopulate(client *clients.AviClient,
	version string, cloud string) ([]interface{}, []interface{}) {
	// Populate the VS cache
	var deletedKeys []interface{}
	SetTenant := session.SetTenant(utils.ADMIN_NS)
	SetTenant(client.AviSession)
	SetVersion := session.SetVersion(version)
	SetVersion(client.AviSession)
	utils.AviLog.Info.Printf("Refreshing all object cache")
	c.AviRefreshObjectCache(client, cloud)
	vsCacheCopy := c.VsCacheMeta.ShallowCopyVSes()
	var allKeys []interface{}
	for k := range vsCacheCopy {
		allKeys = append(allKeys, k)
	}
	err := c.AviObjVSCachePopulate(client, cloud, vsCacheCopy)
	// Populate the SNI VS keys to their respective parents
	c.PopulateVsMetaCache()
	// Delete all the VS keys that are left in the copy.
	if err != nil {
		for key := range vsCacheCopy {
			utils.AviLog.Info.Printf("Removing vs key from cache: %s", key)
			// We want to synthesize these keys to layer 3.
			deletedKeys = append(deletedKeys, key)
			c.VsCacheMeta.AviCacheDelete(key)
		}
	}
	c.AviCloudPropertiesPopulate(client, cloud)
	c.AviObjVrfCachePopulate(client, cloud)
	return deletedKeys, allKeys
}

func (c *AviObjCache) PopulateVsMetaCache() {
	// The vh_child_uuids field is used to populate the SNI children during cache population. However, due to the datastore to PG delay - that field may
	// not always be accurate. We would reduce the problem with accuracy by refreshing the SNI cache through reverse mapping sni's to parent
	// Go over the entire VS cache.
	parentVsKeys := c.VsCacheLocal.AviCacheGetAllParentVSKeys()
	for _, pvsKey := range parentVsKeys {
		// For each parentVs get the SNI children
		sniChildUuids := c.VsCacheLocal.AviCacheGetAllChildVSForParent(pvsKey.(NamespaceName))
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
	allVsKeys := c.VsCacheLocal.AviGetAllVSKeys()
	for _, vsKey := range allVsKeys {
		vsObj, vsFound := c.VsCacheLocal.AviCacheGet(vsKey)
		if vsFound {
			vs_cache_obj, foundvs := vsObj.(*AviVsCache)
			if foundvs {
				vsCopy, done := vs_cache_obj.GetVSCopy()
				if done {
					c.VsCacheMeta.AviCacheAdd(vsKey, vsCopy)
					c.VsCacheLocal.AviCacheDelete(vsKey)
				}
			}
		}
	}
}

func (c *AviObjCache) AviPopulateAllPGs(client *clients.AviClient,
	cloud string, pgData *[]AviPGCache, override_uri ...NextPage) (*[]AviPGCache, int, error) {
	var uri string
	akcUser := utils.OSHIFT_K8S_CLOUD_CONNECTOR
	if len(override_uri) == 1 {
		uri = override_uri[0].Next_uri
	} else {
		uri = "/api/poolgroup?include_name=true&cloud_ref.name=" + cloud + "&created_by=" + akcUser
	}
	result, err := client.AviSession.GetCollectionRaw(uri)
	if err != nil {
		utils.AviLog.Warning.Printf("Get uri %v returned err for pg %v", uri, err)
		return nil, 0, err
	}
	elems := make([]json.RawMessage, result.Count)
	err = json.Unmarshal(result.Results, &elems)
	if err != nil {
		utils.AviLog.Warning.Printf("Failed to unmarshal pg data, err: %v", err)
		return nil, 0, err
	}
	for i := 0; i < len(elems); i++ {
		pg := models.PoolGroup{}
		err = json.Unmarshal(elems[i], &pg)
		if err != nil {
			utils.AviLog.Warning.Printf("Failed to unmarshal pg data, err: %v", err)
			continue
		}

		if pg.Name == nil || pg.UUID == nil || pg.CloudConfigCksum == nil {
			utils.AviLog.Warning.Printf("Incomplete pg data unmarshalled, %s", utils.Stringify(pg))
			continue
		}
		//Only cache a PG that belongs to this AKO.
		if !strings.HasPrefix(*pg.Name, lib.GetVrf()+"--") {
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

func (c *AviObjCache) PopulatePgDataToCache(client *clients.AviClient,
	cloud string) {

	var pgData []AviPGCache
	_, count, err := c.AviPopulateAllPGs(client, cloud, &pgData)
	if err != nil || len(pgData) != count {
		utils.AviLog.Warning.Printf("Incomplete data received, skipping full sync for pgs. Count: %v, pgsdata: %v", count, len(pgData))
		return
	}
	// Get all the PG cache data and copy them.
	pgCacheData := c.PgCache.ShallowCopy()
	for i, pgCacheObj := range pgData {
		k := NamespaceName{Namespace: utils.ADMIN_NS, Name: pgCacheObj.Name}
		oldPGIntf, found := c.PgCache.AviCacheGet(k)
		if found {
			oldPGData, ok := oldPGIntf.(*AviPGCache)
			if ok {
				if oldPGData.InvalidData || oldPGData.LastModified != pgData[i].LastModified {
					pgData[i].InvalidData = true
					utils.AviLog.Warning.Printf("Invalid cache data for pg: %s", k)
				}
			} else {
				utils.AviLog.Warning.Printf("Wrong data type for pg: %s in cache", k)
			}
		}
		utils.AviLog.Info.Printf("Adding key to pg cache :%s value :%s", k, pgCacheObj.Uuid)
		// Add the actual pg cache data
		// Only replace this data if the lastmodifed field varies
		c.PgCache.AviCacheAdd(k, &pgData[i])
		delete(pgCacheData, k)
	}
	// The data that is left in pgCacheData should be explicitly removed
	for key := range pgCacheData {
		utils.AviLog.Info.Printf("Deleting key from pg cache :%s", key)
		c.PgCache.AviCacheDelete(key)
	}
}

func (c *AviObjCache) AviPopulateAllPools(client *clients.AviClient,
	cloud string, poolData *[]AviPoolCache, override_uri ...NextPage) (*[]AviPoolCache, int, error) {
	var uri string
	akcUser := utils.OSHIFT_K8S_CLOUD_CONNECTOR

	if len(override_uri) == 1 {
		uri = override_uri[0].Next_uri
	} else {
		uri = "/api/pool?include_name=true&cloud_ref.name=" + cloud + "&created_by=" + akcUser
	}
	result, err := client.AviSession.GetCollectionRaw(uri)
	if err != nil {
		utils.AviLog.Warning.Printf("Get uri %v returned err for pool %v", uri, err)
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
			utils.AviLog.Warning.Printf("Failed to unmarshal pool data, err: %v", err)
			continue
		}

		if pool.Name == nil || pool.UUID == nil || pool.CloudConfigCksum == nil {
			utils.AviLog.Warning.Printf("Incomplete pool data unmarshalled, %s", utils.Stringify(pool))
			continue
		}
		//Only cache a Pool that belongs to this AKO.
		if !strings.HasPrefix(*pool.Name, lib.GetVrf()+"--") {
			continue
		}
		poolCacheObj := AviPoolCache{
			Name:             *pool.Name,
			Uuid:             *pool.UUID,
			CloudConfigCksum: *pool.CloudConfigCksum,
			LastModified:     *pool.LastModified,
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

func (c *AviObjCache) PopulatePoolsToCache(client *clients.AviClient,
	cloud string, override_uri ...NextPage) {
	var poolsData []AviPoolCache
	_, count, err := c.AviPopulateAllPools(client, cloud, &poolsData)
	if err != nil || len(poolsData) != count {
		utils.AviLog.Warning.Printf("Incomplete data received, skipping full sync for pools. Count: %v, poolsdata: %v", count, len(poolsData))
		return
	}
	poolCacheData := c.PoolCache.ShallowCopy()
	for i, poolCacheObj := range poolsData {
		k := NamespaceName{Namespace: utils.ADMIN_NS, Name: poolCacheObj.Name}
		oldPoolIntf, found := c.PoolCache.AviCacheGet(k)
		if found {
			oldPoolData, ok := oldPoolIntf.(*AviPoolCache)
			if ok {
				if oldPoolData.InvalidData || oldPoolData.LastModified != poolsData[i].LastModified {
					poolsData[i].InvalidData = true
					utils.AviLog.Warning.Printf("Invalid cache data for pool: %s", k)
				}
			} else {
				utils.AviLog.Warning.Printf("Wrong data type for pool: %s in cache", k)
			}
		}
		c.PoolCache.AviCacheAdd(k, &poolsData[i])
		delete(poolCacheData, k)
	}
	// The data that is left in poolCacheData should be explicitly removed
	for key := range poolCacheData {
		utils.AviLog.Info.Printf("Deleting key from pool cache :%s", key)
		c.PoolCache.AviCacheDelete(key)
	}
}

func (c *AviObjCache) AviPopulateAllVSVips(client *clients.AviClient,
	cloud string, vsVipData *[]AviVSVIPCache, nextPage ...NextPage) (*[]AviVSVIPCache, int, error) {
	var uri string
	if len(nextPage) == 1 {
		uri = nextPage[0].Next_uri
	} else {
		uri = "/api/vsvip?include_name=true&cloud_ref.name=" + cloud
	}
	result, err := client.AviSession.GetCollectionRaw(uri)
	if err != nil {
		utils.AviLog.Warning.Printf("Get uri %v returned err for vsvip %v", uri, err)
		return nil, 0, err
	}
	elems := make([]json.RawMessage, result.Count)
	err = json.Unmarshal(result.Results, &elems)
	if err != nil {
		utils.AviLog.Warning.Printf("Failed to unmarshal vsvip data, err: %v", err)
		return nil, 0, err
	}
	for i := 0; i < len(elems); i++ {
		vsvip := models.VsVip{}
		err = json.Unmarshal(elems[i], &vsvip)
		if err != nil {
			utils.AviLog.Warning.Printf("Failed to unmarshal vsvip data, err: %v", err)
			continue
		}

		if vsvip.Name == nil || vsvip.UUID == nil {
			utils.AviLog.Warning.Printf("Incomplete vsvip data unmarshalled, %s", utils.Stringify(vsvip))
			continue
		}
		if !strings.HasPrefix(*vsvip.Name, lib.GetVrf()+"--") {
			continue
		}
		var fqdns []string
		for _, dnsinfo := range vsvip.DNSInfo {
			fqdns = append(fqdns, *dnsinfo.Fqdn)
		}

		vsVipCacheObj := AviVSVIPCache{
			Name:         *vsvip.Name,
			Uuid:         *vsvip.UUID,
			FQDNs:        fqdns,
			LastModified: *vsvip.LastModified,
		}
		*vsVipData = append(*vsVipData, vsVipCacheObj)
	}
	if result.Next != "" {
		// It has a next page, let's recursively call the same method.
		next_uri := strings.Split(result.Next, "/api/vsvip")
		if len(next_uri) > 1 {
			override_uri := "/api/vsvip" + next_uri[1]
			nextPage := NextPage{Next_uri: override_uri}
			_, _, err := c.AviPopulateAllVSVips(client, cloud, vsVipData, nextPage)
			if err != nil {
				return nil, 0, err
			}
		}
	}
	return vsVipData, result.Count, nil
}

func (c *AviObjCache) PopulateVsVipDataToCache(client *clients.AviClient,
	cloud string) {
	var vsVipData []AviVSVIPCache
	_, count, err := c.AviPopulateAllVSVips(client, cloud, &vsVipData)
	if err != nil {
		utils.AviLog.Warning.Printf("Incomplete data received, skipping full sync for vsvip. Count: %v, vsvipdata: %v", count, len(vsVipData))
		return
	}
	vsVipCacheData := c.VSVIPCache.ShallowCopy()
	for i, vsVipCacheObj := range vsVipData {
		k := NamespaceName{Namespace: utils.ADMIN_NS, Name: vsVipCacheObj.Name}
		oldVsvipIntf, found := c.VSVIPCache.AviCacheGet(k)
		if found {
			oldVsvipData, ok := oldVsvipIntf.(*AviVSVIPCache)
			if ok {
				if oldVsvipData.InvalidData || oldVsvipData.LastModified != vsVipData[i].LastModified {
					vsVipData[i].InvalidData = true
					utils.AviLog.Warning.Printf("Invalid cache data for vsvip: %s", k)
				}
			} else {
				utils.AviLog.Warning.Printf("Wrong data type for vsvip: %s in cache", k)
			}
		}
		utils.AviLog.Info.Printf("Adding key to vsvip cache: %s, fqdns: %v", k, vsVipData[i].FQDNs)
		c.VSVIPCache.AviCacheAdd(k, &vsVipData[i])
		delete(vsVipCacheData, k)
	}
	// The data that is left in vsVipCacheData should be explicitly removed
	for key := range vsVipCacheData {
		utils.AviLog.Info.Printf("Deleting key from vsvip cache :%s", key)
		c.VSVIPCache.AviCacheDelete(key)
	}
}

func (c *AviObjCache) AviPopulateAllDSs(client *clients.AviClient,
	cloud string, DsData *[]AviDSCache, nextPage ...NextPage) (*[]AviDSCache, int, error) {
	var uri string
	akcUser := utils.OSHIFT_K8S_CLOUD_CONNECTOR
	if len(nextPage) == 1 {
		uri = nextPage[0].Next_uri
	} else {
		uri = "/api/vsdatascriptset?include_name=true&created_by=" + akcUser
	}
	result, err := client.AviSession.GetCollectionRaw(uri)
	if err != nil {
		utils.AviLog.Warning.Printf("Get uri %v returned err for datascript %v", uri, err)
		return nil, 0, err
	}
	elems := make([]json.RawMessage, result.Count)
	err = json.Unmarshal(result.Results, &elems)
	if err != nil {
		utils.AviLog.Warning.Printf("Failed to unmarshal datascript data, err: %v", err)
		return nil, 0, err
	}
	for i := 0; i < len(elems); i++ {
		ds := models.VSDataScriptSet{}
		err = json.Unmarshal(elems[i], &ds)
		if err != nil {
			utils.AviLog.Warning.Printf("Failed to unmarshal datascript data, err: %v", err)
			continue
		}
		if ds.Name == nil || ds.UUID == nil {
			utils.AviLog.Warning.Printf("Incomplete Datascript data unmarshalled, %s", utils.Stringify(ds))
			continue
		}
		//Only cache a DS that belongs to this AKO.
		if !strings.HasPrefix(*ds.Name, lib.GetVrf()+"--") {
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

func (c *AviObjCache) PopulateDSDataToCache(client *clients.AviClient,
	cloud string, override_uri ...NextPage) {
	var DsData []AviDSCache
	_, count, err := c.AviPopulateAllDSs(client, cloud, &DsData)
	dsCacheData := c.DSCache.ShallowCopy()
	if err != nil || len(DsData) != count {
		return
	}
	for i, DsCacheObj := range DsData {
		k := NamespaceName{Namespace: utils.ADMIN_NS, Name: DsCacheObj.Name}
		oldDSIntf, found := c.DSCache.AviCacheGet(k)
		if found {
			oldDSData, ok := oldDSIntf.(*AviDSCache)
			if ok {
				if oldDSData.InvalidData || oldDSData.LastModified != DsData[i].LastModified {
					DsData[i].InvalidData = true
					utils.AviLog.Warning.Printf("Invalid cache data for datascript: %s", k)
				}
			} else {
				utils.AviLog.Warning.Printf("Wrong data type for datascript: %s in cache", k)
			}
		}
		utils.AviLog.Info.Printf("Adding key to ds cache :%s", k)
		c.DSCache.AviCacheAdd(k, &DsData[i])
		delete(dsCacheData, k)
	}
	// The data that is left in dsCacheData should be explicitly removed
	for key := range dsCacheData {
		utils.AviLog.Info.Printf("Deleting key from ds cache :%s", key)
		c.DSCache.AviCacheDelete(key)
	}
}

func (c *AviObjCache) AviPopulateAllSSLKeys(client *clients.AviClient,
	cloud string, SslData *[]AviSSLCache, nextPage ...NextPage) (*[]AviSSLCache, int, error) {
	var uri string
	akcUser := utils.OSHIFT_K8S_CLOUD_CONNECTOR
	if len(nextPage) == 1 {
		uri = nextPage[0].Next_uri
	} else {
		uri = "/api/sslkeyandcertificate?include_name=true" + "&created_by=" + akcUser
	}
	result, err := client.AviSession.GetCollectionRaw(uri)
	if err != nil {
		utils.AviLog.Warning.Printf("Get uri %v returned err for sslkeyandcertificate %v", uri, err)
		return nil, 0, err
	}
	elems := make([]json.RawMessage, result.Count)
	err = json.Unmarshal(result.Results, &elems)
	if err != nil {
		utils.AviLog.Warning.Printf("Failed to unmarshal sslkeyandcertificate data, err: %v", err)
		return nil, 0, err
	}
	for i := 0; i < len(elems); i++ {
		sslkey := models.SSLKeyAndCertificate{}
		err = json.Unmarshal(elems[i], &sslkey)
		if err != nil {
			utils.AviLog.Warning.Printf("Failed to unmarshal sslkeyandcertificate data, err: %v", err)
			continue
		}
		if sslkey.Name == nil || sslkey.UUID == nil {
			utils.AviLog.Warning.Printf("Incomplete sslkey data unmarshalled, %s", utils.Stringify(sslkey))
			continue
		}
		//Only cache a SSL keys that belongs to this AKO.
		if !strings.HasPrefix(*sslkey.Name, lib.GetVrf()+"--") {
			continue
		}
		sslCacheObj := AviSSLCache{
			Name: *sslkey.Name,
			Uuid: *sslkey.UUID,
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

// This method is just added for future here. We can't use it until they expose the DB extensions on the virtualservice object
func (c *AviObjCache) AviPopulateAllVSMeta(client *clients.AviClient,
	cloud string, vsData *[]AviVsCache, nextPage ...NextPage) (*[]AviVsCache, error) {
	var uri string
	akcUser := utils.OSHIFT_K8S_CLOUD_CONNECTOR
	if len(nextPage) == 1 {
		uri = nextPage[0].Next_uri
	} else {
		uri = "/api/virtualservice?include_name=true&cloud_ref.name=" + cloud + "&vrf_context_ref.name=" + lib.GetVrf() + "&created_by=" + akcUser
	}
	result, err := client.AviSession.GetCollectionRaw(uri)
	if err != nil {
		utils.AviLog.Warning.Printf("Get uri %v returned err for virtualservice %v", uri, err)
		return nil, err
	}
	elems := make([]json.RawMessage, result.Count)
	err = json.Unmarshal(result.Results, &elems)
	if err != nil {
		utils.AviLog.Warning.Printf("Failed to unmarshal virtualservice data, err: %v", err)
		return nil, err
	}
	for i := 0; i < len(elems); i++ {
		vsModel := models.VirtualService{}
		err = json.Unmarshal(elems[i], &vsModel)
		if err != nil {
			utils.AviLog.Warning.Printf("Failed to unmarshal virtualservice data, err: %v", err)
			continue
		}
		if vsModel.Name == nil || vsModel.UUID == nil {
			utils.AviLog.Warning.Printf("Incomplete sslkey data unmarshalled, %s", utils.Stringify(vsModel))
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
				vipKey := NamespaceName{Namespace: utils.ADMIN_NS, Name: vsVipName.(string)}
				vsVipKey = append(vsVipKey, vipKey)
			}
		}
		if vsModel.SslKeyAndCertificateRefs != nil {
			for _, ssl := range vsModel.SslKeyAndCertificateRefs {
				// find the sslkey name from the ssl key cache
				sslUuid := ExtractUuid(ssl, "sslkeyandcertificate-.*.#")
				sslName, foundssl := c.SSLKeyCache.AviCacheGetNameByUuid(sslUuid)
				if foundssl {
					sslKey := NamespaceName{Namespace: utils.ADMIN_NS, Name: sslName.(string)}
					sslKeys = append(sslKeys, sslKey)
				}
			}
		}
		if vsModel.VsDatascripts != nil {
			for _, dataScript := range vsModel.VsDatascripts {
				dsUuid := ExtractUuid(*dataScript.VsDatascriptSetRef, "vsdatascriptset-.*.#")

				dsName, foundDs := c.DSCache.AviCacheGetNameByUuid(dsUuid)
				if foundDs {
					dsKey := NamespaceName{Namespace: utils.ADMIN_NS, Name: dsName.(string)}
					// Fetch the associated PGs with the DS.
					dsObj, _ := c.DSCache.AviCacheGet(dsKey)
					for _, pgName := range dsObj.(*AviDSCache).PoolGroups {
						// For each PG, formulate the key and then populate the pg collection cache
						pgKey := NamespaceName{Namespace: utils.ADMIN_NS, Name: pgName}
						poolgroupKeys = append(poolgroupKeys, pgKey)
						poolKeys = c.AviPGPoolCachePopulate(pgName)
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
					pgKey := NamespaceName{Namespace: utils.ADMIN_NS, Name: pgName.(string)}
					poolgroupKeys = append(poolgroupKeys, pgKey)
					poolKeys = c.AviPGPoolCachePopulate(pgName.(string))
				}

			}
		}
		if vsModel.HTTPPolicies != nil {
			for _, http_intf := range vsModel.HTTPPolicies {

				httpUuid := ExtractUuid(*http_intf.HTTPPolicySetRef, "httppolicyset-.*.#")

				httpName, foundhttp := c.HTTPPolicyCache.AviCacheGetNameByUuid(httpUuid)
				if foundhttp {
					httpKey := NamespaceName{Namespace: utils.ADMIN_NS, Name: httpName.(string)}
					httpObj, _ := c.HTTPPolicyCache.AviCacheGet(httpKey)
					for _, pgName := range httpObj.(*AviHTTPPolicyCache).PoolGroups {
						// For each PG, formulate the key and then populate the pg collection cache
						pgKey := NamespaceName{Namespace: utils.ADMIN_NS, Name: pgName}
						poolgroupKeys = append(poolgroupKeys, pgKey)
						poolKeys = c.AviPGPoolCachePopulate(pgName)
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
func (c *AviObjCache) PopulateVsKeyToCache(client *clients.AviClient,
	cloud string, override_uri ...NextPage) {
	var vsCacheMeta []AviVsCache
	_, err := c.AviPopulateAllVSMeta(client, cloud, &vsCacheMeta)
	if err != nil {
		return
	}
}

func (c *AviObjCache) PopulateSSLKeyToCache(client *clients.AviClient,
	cloud string, override_uri ...NextPage) {
	var SslKeyData []AviSSLCache
	_, count, err := c.AviPopulateAllSSLKeys(client, cloud, &SslKeyData)
	if err != nil || len(SslKeyData) != count {
		utils.AviLog.Warning.Printf("Incomplete data received, skipping full sync for sslkeycert. Count: %v, ssldata: %v", count, len(SslKeyData))
		return
	}
	sslCacheData := c.SSLKeyCache.ShallowCopy()
	for i, SslKeyCacheObj := range SslKeyData {
		k := NamespaceName{Namespace: utils.ADMIN_NS, Name: SslKeyCacheObj.Name}
		oldSslkeyIntf, found := c.SSLKeyCache.AviCacheGet(k)
		if found {
			oldSslkeyData, ok := oldSslkeyIntf.(*AviSSLCache)
			if ok {
				if oldSslkeyData.InvalidData || oldSslkeyData.LastModified != SslKeyData[i].LastModified {
					SslKeyData[i].InvalidData = true
					utils.AviLog.Warning.Printf("Invalid cache data for ssl key: %s", k)
				}
			} else {
				utils.AviLog.Warning.Printf("Wrong data type for ssl key: %s in cache", k)
			}
		}
		utils.AviLog.Info.Printf("Adding key to sslkey cache :%s", k)
		c.SSLKeyCache.AviCacheAdd(k, &SslKeyData[i])
		delete(sslCacheData, k)
	}
	//The data that is left in sslCacheData should be explicitly removed
	for key := range sslCacheData {
		utils.AviLog.Info.Printf("Deleting key from sslkey cache :%s", key)
		c.SSLKeyCache.AviCacheDelete(key)
	}
}

func (c *AviObjCache) AviPopulateAllHttpPolicySets(client *clients.AviClient,
	cloud string, httpPolicyData *[]AviHTTPPolicyCache, nextPage ...NextPage) (*[]AviHTTPPolicyCache, int, error) {
	var uri string
	akcUser := utils.OSHIFT_K8S_CLOUD_CONNECTOR
	if len(nextPage) == 1 {
		uri = nextPage[0].Next_uri
	} else {
		uri = "/api/httppolicyset?include_name=true" + "&created_by=" + akcUser
	}
	result, err := client.AviSession.GetCollectionRaw(uri)
	utils.AviLog.Info.Printf("Http policy set returned :%v, results", result.Count)
	if err != nil {
		utils.AviLog.Warning.Printf("Get uri %v returned err for httppolicyset %v", uri, err)
		return nil, 0, err
	}
	elems := make([]json.RawMessage, result.Count)
	err = json.Unmarshal(result.Results, &elems)
	if err != nil {
		utils.AviLog.Warning.Printf("Failed to unmarshal httppolicyset data, err: %v", err)
		return nil, 0, err
	}
	for i := 0; i < len(elems); i++ {
		httppol := models.HTTPPolicySet{}
		err = json.Unmarshal(elems[i], &httppol)
		if err != nil {
			utils.AviLog.Warning.Printf("Failed to unmarshal httppolicyset data, err: %v", err)
			continue
		}
		if httppol.Name == nil || httppol.UUID == nil || httppol.CloudConfigCksum == nil {
			utils.AviLog.Warning.Printf("Incomplete http policy data unmarshalled, %s", utils.Stringify(httppol))
			continue
		}
		//Only cache a http policies that belongs to this AKO.
		if !strings.HasPrefix(*httppol.Name, lib.GetVrf()+"--") {
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

func (c *AviObjCache) PopulateHttpPolicySetToCache(client *clients.AviClient,
	cloud string, override_uri ...NextPage) {
	var HttPolData []AviHTTPPolicyCache
	_, count, err := c.AviPopulateAllHttpPolicySets(client, cloud, &HttPolData)
	if err != nil || len(HttPolData) != count {
		return
	}
	httpCacheData := c.HTTPPolicyCache.ShallowCopy()
	for i, HttpPolCacheObj := range HttPolData {
		k := NamespaceName{Namespace: utils.ADMIN_NS, Name: HttpPolCacheObj.Name}
		oldHttppolIntf, found := c.HTTPPolicyCache.AviCacheGet(k)
		if found {
			oldHttppolData, ok := oldHttppolIntf.(*AviHTTPPolicyCache)
			if ok {
				if oldHttppolData.InvalidData || oldHttppolData.LastModified != HttPolData[i].LastModified {
					HttPolData[i].InvalidData = true
					utils.AviLog.Warning.Printf("Invalid cache data for http policy: %s", k)
				}
			} else {
				utils.AviLog.Warning.Printf("Wrong data type for http policy: %s in cache", k)
			}
		}
		utils.AviLog.Info.Printf("Adding key to httppol cache :%s", k)
		c.HTTPPolicyCache.AviCacheAdd(k, &HttPolData[i])
		delete(httpCacheData, k)
	}
	// // The data that is left in httpCacheData should be explicitly removed
	for key := range httpCacheData {
		utils.AviLog.Info.Printf("Deleting key from httppol cache :%s", key)
		c.HTTPPolicyCache.AviCacheDelete(key)
	}
}

func (c *AviObjCache) AviObjVrfCachePopulate(client *clients.AviClient, cloud string) {
	disableStaticRoute := os.Getenv(lib.DISABLE_STATIC_ROUTE_SYNC)
	if disableStaticRoute == "true" {
		utils.AviLog.Info.Printf("Static route sync disabled, skipping vrf cache population")
		return
	}
	uri := "/api/vrfcontext?name=" + lib.GetVrf() + "&include_name=true&cloud_ref.name=" + cloud

	vrfList := []*models.VrfContext{}
	result, err := client.AviSession.GetCollectionRaw(uri)
	if err != nil {
		utils.AviLog.Warning.Printf("Get uri %v returned err %v", uri, err)
		return
	}
	elems := make([]json.RawMessage, result.Count)
	err = json.Unmarshal(result.Results, &elems)
	if err != nil {
		utils.AviLog.Warning.Printf("Failed to unmarshal data, err: %v", err)
		return
	}
	for i := 0; i < result.Count; i++ {
		vrf := models.VrfContext{}
		err = json.Unmarshal(elems[i], &vrf)
		if err != nil {
			utils.AviLog.Warning.Printf("Failed to unmarshal data, err: %v", err)
			continue
		}
		vrfList = append(vrfList, &vrf)

		vrfName := *vrf.Name
		checksum := VrfChecksum(vrfName, vrf.StaticRoutes)
		vrfCacheObj := AviVrfCache{
			Name:             vrfName,
			Uuid:             *vrf.UUID,
			CloudConfigCksum: checksum,
		}
		utils.AviLog.Info.Printf("Adding vrf to Cache %s\n", vrfName)
		c.VrfCache.AviCacheAdd(vrfName, &vrfCacheObj)
	}
}

func (c *AviObjCache) AviObjVSCachePopulate(client *clients.AviClient,
	cloud string, vsCacheCopy map[interface{}]interface{}, override_uri ...NextPage) error {
	var rest_response interface{}
	akcUser := utils.OSHIFT_K8S_CLOUD_CONNECTOR
	var uri string
	if len(override_uri) == 1 {
		uri = override_uri[0].Next_uri
	} else {
		uri = "/api/virtualservice?include_name=true&cloud_ref.name=" + cloud + "&vrf_context_ref.name=" + lib.GetVrf() + "&created_by=" + akcUser
	}
	err := client.AviSession.Get(uri, &rest_response)

	if err != nil {
		utils.AviLog.Warning.Printf("Vs Get uri %v returned err %v", uri, err)
		return err
	} else {
		resp, ok := rest_response.(map[string]interface{})
		if !ok {
			utils.AviLog.Warning.Printf("Vs Get uri %v returned %v type %T", uri,
				rest_response, rest_response)
			return errors.New("VS type is wrong")
		}
		utils.AviLog.Info.Printf("Vs Get uri %v returned %v vses", uri,
			resp["count"])
		results, ok := resp["results"].([]interface{})
		if !ok {
			utils.AviLog.Warning.Printf("results not of type []interface{} Instead of type %T", resp["results"])
			return errors.New("Results are not of right type for VS")
		}
		for _, vs_intf := range results {

			vs, ok := vs_intf.(map[string]interface{})
			if !ok {
				utils.AviLog.Warning.Printf("vs_intf not of type map[string] interface{}. Instead of type %T", vs_intf)
				continue
			}
			svc_mdata_intf, ok := vs["service_metadata"]
			var svc_mdata_obj ServiceMetadataObj
			if ok {
				if err := json.Unmarshal([]byte(svc_mdata_intf.(string)),
					&svc_mdata_obj); err != nil {
					utils.AviLog.Warning.Printf("Error parsing service metadata during vs cache :%v", err)
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
				utils.AviLog.Info.Printf("extracted the vs uuid from parent ref during cache population: %s", vs_uuid)
				// Now let's get the VS key from this uuid
				vsKey, gotVS := c.VsCacheLocal.AviCacheGetKeyByUuid(vs_uuid)
				if gotVS {
					parentVSKey = vsKey.(NamespaceName)
				}

			}
			if vs["cloud_config_cksum"] != nil {
				k := NamespaceName{Namespace: utils.ADMIN_NS, Name: vs["name"].(string)}
				delete(vsCacheCopy, k)
				var vip string
				var vsVipKey []NamespaceName
				var sslKeys []NamespaceName
				var dsKeys []NamespaceName
				var httpKeys []NamespaceName
				var poolgroupKeys []NamespaceName
				var poolKeys []NamespaceName
				if vs["vip"] != nil && len(vs["vip"].([]interface{})) > 0 {
					vip = (vs["vip"].([]interface{})[0].(map[string]interface{})["ip_address"]).(map[string]interface{})["addr"].(string)
				}
				// Populate the VSVIP cache
				if vs["vsvip_ref"] != nil {
					// find the vsvip name from the vsvip cache
					vsVipUuid := ExtractUuid(vs["vsvip_ref"].(string), "vsvip-.*.#")
					vsVipName, foundVip := c.VSVIPCache.AviCacheGetNameByUuid(vsVipUuid)
					if foundVip {
						vipKey := NamespaceName{Namespace: utils.ADMIN_NS, Name: vsVipName.(string)}
						vsVipKey = append(vsVipKey, vipKey)
					}
				}
				if vs["ssl_key_and_certificate_refs"] != nil {
					for _, ssl := range vs["ssl_key_and_certificate_refs"].([]interface{}) {
						// find the sslkey name from the ssl key cache
						sslUuid := ExtractUuid(ssl.(string), "sslkeyandcertificate-.*.#")
						sslName, foundssl := c.SSLKeyCache.AviCacheGetNameByUuid(sslUuid)
						if foundssl {
							sslKey := NamespaceName{Namespace: utils.ADMIN_NS, Name: sslName.(string)}
							sslKeys = append(sslKeys, sslKey)
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
								dsKey := NamespaceName{Namespace: utils.ADMIN_NS, Name: dsName.(string)}
								// Fetch the associated PGs with the DS.
								dsObj, _ := c.DSCache.AviCacheGet(dsKey)
								for _, pgName := range dsObj.(*AviDSCache).PoolGroups {
									// For each PG, formulate the key and then populate the pg collection cache
									pgKey := NamespaceName{Namespace: utils.ADMIN_NS, Name: pgName}
									poolgroupKeys = append(poolgroupKeys, pgKey)
									poolKeys = c.AviPGPoolCachePopulate(pgName)
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
							pgUuid := ExtractUuid(pgmap["service_pool_group_ref"].(string), "poolgroup-.*.#")

							pgName, foundpg := c.PgCache.AviCacheGetNameByUuid(pgUuid)
							if foundpg {
								pgKey := NamespaceName{Namespace: utils.ADMIN_NS, Name: pgName.(string)}
								poolgroupKeys = append(poolgroupKeys, pgKey)
								poolKeys = c.AviPGPoolCachePopulate(pgName.(string))
							}
						}
					}
				}
				if vs["http_policies"] != nil {
					for _, http_intf := range vs["http_policies"].([]interface{}) {
						// find the sslkey name from the ssl key cache
						httpmap, ok := http_intf.(map[string]interface{})
						if ok {
							httpUuid := ExtractUuid(httpmap["http_policy_set_ref"].(string), "httppolicyset-.*.#")

							httpName, foundhttp := c.HTTPPolicyCache.AviCacheGetNameByUuid(httpUuid)
							if foundhttp {
								httpKey := NamespaceName{Namespace: utils.ADMIN_NS, Name: httpName.(string)}
								httpObj, _ := c.HTTPPolicyCache.AviCacheGet(httpKey)
								for _, pgName := range httpObj.(*AviHTTPPolicyCache).PoolGroups {
									// For each PG, formulate the key and then populate the pg collection cache
									pgKey := NamespaceName{Namespace: utils.ADMIN_NS, Name: pgName}
									poolgroupKeys = append(poolgroupKeys, pgKey)
									poolKeys = c.AviPGPoolCachePopulate(pgName)
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
					LastModified:         vs["_last_modified"].(string),
				}
				c.VsCacheLocal.AviCacheAdd(k, &vsMetaObj)
				utils.AviLog.Info.Printf("Added VS cache key :%s", k)

			}
		}
		if resp["next"] != nil {
			// It has a next page, let's recursively call the same method.
			next_uri := strings.Split(resp["next"].(string), "/api/virtualservice")
			utils.AviLog.Info.Printf("Found next page for vs, uri: %s", next_uri)
			if len(next_uri) > 1 {
				override_uri := "/api/virtualservice" + next_uri[1]
				utils.AviLog.Info.Printf("Next page uri for vs: %s", override_uri)
				nextPage := NextPage{Next_uri: override_uri}
				c.AviObjVSCachePopulate(client, cloud, vsCacheCopy, nextPage)
			}
		}
	}
	return nil
}

func (c *AviObjCache) AviObjOneVSCachePopulate(client *clients.AviClient,
	cloud string, vs_uuid string) error {
	var vs_intf interface{}
	akcUser := utils.OSHIFT_K8S_CLOUD_CONNECTOR
	var uri string

	uri = "/api/virtualservice/" + vs_uuid + "?include_name=true&cloud_ref.name=" + cloud + "&vrf_context_ref.name=" + lib.GetVrf() + "&created_by=" + akcUser

	err := client.AviSession.Get(uri, &vs_intf)

	if err != nil {
		utils.AviLog.Warning.Printf("Vs Get uri %v returned err %v", uri, err)
		return err
	} else {
		vs, ok := vs_intf.(map[string]interface{})
		if !ok {
			utils.AviLog.Warning.Printf("vs_intf not of type map[string] interface{}. Instead of type %T", vs_intf)
			return errors.New("VS object is corrupted")
		}
		svc_mdata_intf, ok := vs["service_metadata"]
		var svc_mdata_obj ServiceMetadataObj
		if ok {
			if err := json.Unmarshal([]byte(svc_mdata_intf.(string)),
				&svc_mdata_obj); err != nil {
				utils.AviLog.Warning.Printf("Error parsing service metadata during vs cache :%v", err)
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
			utils.AviLog.Info.Printf("extracted the vs uuid from parent ref during cache population: %s", vs_uuid)
			// Now let's get the VS key from this uuid
			vsKey, gotVS := c.VsCacheLocal.AviCacheGetKeyByUuid(vs_uuid)
			if gotVS {
				parentVSKey = vsKey.(NamespaceName)
			}

		}
		if vs["cloud_config_cksum"] != nil {
			k := NamespaceName{Namespace: utils.ADMIN_NS, Name: vs["name"].(string)}
			var vip string
			var vsVipKey []NamespaceName
			var sslKeys []NamespaceName
			var dsKeys []NamespaceName
			var httpKeys []NamespaceName
			var poolgroupKeys []NamespaceName
			var poolKeys []NamespaceName
			if vs["vip"] != nil && len(vs["vip"].([]interface{})) > 0 {
				vip = (vs["vip"].([]interface{})[0].(map[string]interface{})["ip_address"]).(map[string]interface{})["addr"].(string)
			}
			// Populate the VSVIP cache
			if vs["vsvip_ref"] != nil {
				// find the vsvip name from the vsvip cache
				vsVipUuid := ExtractUuid(vs["vsvip_ref"].(string), "vsvip-.*.#")
				vsVipName, foundVip := c.VSVIPCache.AviCacheGetNameByUuid(vsVipUuid)
				if foundVip {
					vipKey := NamespaceName{Namespace: utils.ADMIN_NS, Name: vsVipName.(string)}
					vsVipKey = append(vsVipKey, vipKey)
				}
			}
			if vs["ssl_key_and_certificate_refs"] != nil {
				for _, ssl := range vs["ssl_key_and_certificate_refs"].([]interface{}) {
					// find the sslkey name from the ssl key cache
					sslUuid := ExtractUuid(ssl.(string), "sslkeyandcertificate-.*.#")
					sslName, foundssl := c.SSLKeyCache.AviCacheGetNameByUuid(sslUuid)
					if foundssl {
						sslKey := NamespaceName{Namespace: utils.ADMIN_NS, Name: sslName.(string)}
						sslKeys = append(sslKeys, sslKey)
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
							dsKey := NamespaceName{Namespace: utils.ADMIN_NS, Name: dsName.(string)}
							// Fetch the associated PGs with the DS.
							dsObj, _ := c.DSCache.AviCacheGet(dsKey)
							for _, pgName := range dsObj.(*AviDSCache).PoolGroups {
								// For each PG, formulate the key and then populate the pg collection cache
								pgKey := NamespaceName{Namespace: utils.ADMIN_NS, Name: pgName}
								poolgroupKeys = append(poolgroupKeys, pgKey)
								poolKeys = c.AviPGPoolCachePopulate(pgName)
							}
							dsKeys = append(dsKeys, dsKey)
						}
					}
				}
			}
			if vs["http_policies"] != nil {
				for _, http_intf := range vs["http_policies"].([]interface{}) {
					// find the sslkey name from the ssl key cache
					httpmap, ok := http_intf.(map[string]interface{})
					if ok {
						httpUuid := ExtractUuid(httpmap["http_policy_set_ref"].(string), "httppolicyset-.*.#")

						httpName, foundhttp := c.HTTPPolicyCache.AviCacheGetNameByUuid(httpUuid)
						if foundhttp {
							httpKey := NamespaceName{Namespace: utils.ADMIN_NS, Name: httpName.(string)}
							httpObj, _ := c.HTTPPolicyCache.AviCacheGet(httpKey)
							for _, pgName := range httpObj.(*AviHTTPPolicyCache).PoolGroups {
								// For each PG, formulate the key and then populate the pg collection cache
								pgKey := NamespaceName{Namespace: utils.ADMIN_NS, Name: pgName}
								poolgroupKeys = append(poolgroupKeys, pgKey)
								poolKeys = c.AviPGPoolCachePopulate(pgName)
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
			}
			c.VsCacheLocal.AviCacheAdd(k, &vsMetaObj)
			utils.AviLog.Info.Printf("Added VS cache key :%s", k)
		}
	}
	return nil
}

func (c *AviObjCache) AviPGPoolCachePopulate(pgName string) []NamespaceName {
	var poolKeyCollection []NamespaceName

	k := NamespaceName{Namespace: utils.ADMIN_NS, Name: pgName}
	// Find the pools associated with this PG and populate them
	pgObj, ok := c.PgCache.AviCacheGet(k)
	// Get the members from this and populate the VS ref
	if ok {
		for _, poolName := range pgObj.(*AviPGCache).Members {
			k := NamespaceName{Namespace: utils.ADMIN_NS, Name: poolName}
			poolKeyCollection = append(poolKeyCollection, k)
		}
	}
	return poolKeyCollection
}

func (c *AviObjCache) AviCloudPropertiesPopulate(client *clients.AviClient,
	cloud string) {
	vtype := os.Getenv("CLOUD_VTYPE")
	if vtype == "" {
		// Default to vcenter.
		vtype = "CLOUD_VCENTER"
	}
	var rest_response interface{}
	uri := "/api/cloud"
	err := client.AviSession.Get(uri, &rest_response)
	if err != nil {
		utils.AviLog.Warning.Printf("CloudProperties Get uri %v returned err %v", uri, err)
	} else {
		resp, ok := rest_response.(map[string]interface{})
		if !ok {
			utils.AviLog.Warning.Printf("CloudProperties Get uri %v returned %v type %T", uri,
				rest_response, rest_response)
			return
		}
		utils.AviLog.Info.Printf("CloudProperties Get uri %v returned %v ", uri,
			resp["count"])
		results, ok := resp["results"].([]interface{})
		if !ok {
			utils.AviLog.Warning.Printf("results not of type []interface{} Instead of type %T ", resp["results"])
			return
		}
		for _, cloud_intf := range results {
			cloud_pol, ok := cloud_intf.(map[string]interface{})
			if !ok {
				utils.AviLog.Warning.Printf("cloud_intf not of type map[string] interface{}. Instead of type %T", cloud_intf)
				continue
			}

			if cloud == cloud_pol["name"] {

				cloud_obj := &AviCloudPropertyCache{Name: cloud, VType: vtype}
				if cloud_pol["dns_provider_ref"] != nil {
					dns_uuid := ExtractPattern(cloud_pol["dns_provider_ref"].(string), "ipamdnsproviderprofile-.*")
					subdomains := c.AviDNSPropertyPopulate(client, dns_uuid)
					if subdomains != nil {
						cloud_obj.NSIpamDNS = subdomains
					}

				} else {
					utils.AviLog.Warning.Printf("Cloud does not have a dns_provider_ref configured %v", cloud)
				}
				c.CloudKeyCache.AviCacheAdd(cloud, cloud_obj)
				utils.AviLog.Info.Printf("Added CloudKeyCache cache key %v val %v",
					cloud, cloud_obj)
			}

		}
	}
}

func (c *AviObjCache) AviDNSPropertyPopulate(client *clients.AviClient,
	nsDNSIpam string) []string {
	var rest_response interface{}
	var dnsSubDomains []string
	uri := "/api/ipamdnsproviderprofile/"
	err := client.AviSession.Get(uri, &rest_response)
	if err != nil {
		utils.AviLog.Warning.Printf("DNSProperty Get uri %v returned err %v", uri, err)
		return nil
	} else {
		resp, ok := rest_response.(map[string]interface{})
		if !ok {
			utils.AviLog.Warning.Printf("DNSProperty Get uri %v returned %v type %T", uri,
				rest_response, rest_response)
			return nil
		}
		utils.AviLog.Info.Printf("DNSProperty Get uri %v returned %v ", uri,
			resp["count"])
		results, ok := resp["results"].([]interface{})
		if !ok {
			utils.AviLog.Warning.Printf("results not of type []interface{} Instead of type %T ", resp["results"])
			return nil
		}
		for _, dns_intf := range results {
			dns_pol, ok := dns_intf.(map[string]interface{})
			if !ok {
				utils.AviLog.Warning.Printf("dns_intf not of type map[string] interface{}. Instead of type %T", dns_intf)
				continue
			}
			if dns_pol["uuid"] == nsDNSIpam {

				dns_profile := dns_pol["internal_profile"]
				dns_profile_pol, dns_found := dns_profile.(map[string]interface{})
				if dns_found {
					dns_ipam := dns_profile_pol["dns_service_domain"].([]interface{})[0].(map[string]interface{})
					// Pick the first dns profile
					utils.AviLog.Info.Printf("Found DNS_IPAM: %v", dns_ipam["domain_name"])
					dnsSubDomains = append(dnsSubDomains, dns_ipam["domain_name"].(string))
				}

			}

		}
	}
	return dnsSubDomains
}

func ExtractPattern(word string, pattern string) string {
	r, _ := regexp.Compile(pattern)
	result := r.FindAllString(word, -1)
	if len(result) == 1 {
		return result[0][:len(result[0])]
	}
	return ""
}

func ExtractUuid(word, pattern string) string {
	r, _ := regexp.Compile(pattern)
	result := r.FindAllString(word, -1)
	if len(result) == 1 {
		return result[0][:len(result[0])-1]
	}
	return ""
}
