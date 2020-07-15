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
	"encoding/json"
	"errors"
	"os"
	"regexp"
	"strings"
	"sync"

	"ako/pkg/lib"

	apimodels "github.com/avinetworks/container-lib/api/models"
	"github.com/avinetworks/container-lib/utils"
	"github.com/avinetworks/sdk/go/clients"
	"github.com/avinetworks/sdk/go/models"
	"github.com/avinetworks/sdk/go/session"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type AviObjCache struct {
	PgCache         *AviCache
	DSCache         *AviCache
	PoolCache       *AviCache
	CloudKeyCache   *AviCache
	HTTPPolicyCache *AviCache
	L4PolicyCache   *AviCache
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
	c.PgCache = NewAviCache()
	c.DSCache = NewAviCache()
	c.PoolCache = NewAviCache()
	c.SSLKeyCache = NewAviCache()
	c.CloudKeyCache = NewAviCache()
	c.HTTPPolicyCache = NewAviCache()
	c.L4PolicyCache = NewAviCache()
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

func (c *AviObjCache) AviRefreshObjectCache(client *clients.AviClient, cloud string) {
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

func (c *AviObjCache) AviObjCachePopulate(client *clients.AviClient, version string, cloud string) ([]NamespaceName, []NamespaceName) {
	SetTenant := session.SetTenant(lib.GetTenant())
	SetTenant(client.AviSession)
	SetVersion := session.SetVersion(version)
	SetVersion(client.AviSession)
	c.AviObjVrfCachePopulate(client, cloud)
	// Populate the VS cache
	utils.AviLog.Infof("Refreshing all object cache")
	c.AviRefreshObjectCache(client, cloud)
	vsCacheCopy := c.VsCacheMeta.AviCacheGetAllParentVSKeys()
	allVsKeys := c.VsCacheMeta.AviGetAllVSKeys()
	c.AviObjVSCachePopulate(client, cloud, &allVsKeys)
	// Populate the SNI VS keys to their respective parents
	c.PopulateVsMetaCache()
	// Delete all the VS keys that are left in the copy.
	for _, key := range allVsKeys {
		utils.AviLog.Debugf("Removing vs key from cache: %s", key)
		// We want to synthesize these keys to layer 3.
		vsCacheCopy = Remove(vsCacheCopy, key)
		c.VsCacheMeta.AviCacheDelete(key)
	}
	c.AviCloudPropertiesPopulate(client, cloud)
	//vsCacheCopy at this time, is left with only the deleted keys
	return vsCacheCopy, allVsKeys
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

func (c *AviObjCache) AviPopulateAllPGs(client *clients.AviClient, cloud string, pgData *[]AviPGCache, override_uri ...NextPage) (*[]AviPGCache, int, error) {
	var uri string
	akoUser := lib.AKOUser

	if len(override_uri) == 1 {
		uri = override_uri[0].Next_uri
	} else {
		uri = "/api/poolgroup/?" + "include_name=true&cloud_ref.name=" + cloud + "&created_by=" + akoUser + "&page_size=100"
	}

	result, err := AviGetCollectionRaw(client, uri)
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

func (c *AviObjCache) AviPopulateAllPools(client *clients.AviClient, cloud string, poolData *[]AviPoolCache, override_uri ...NextPage) (*[]AviPoolCache, int, error) {
	var uri string
	akoUser := lib.AKOUser

	if len(override_uri) == 1 {
		uri = override_uri[0].Next_uri
	} else {
		uri = "/api/pool/?" + "&include_name=true&cloud_ref.name=" + cloud + "&created_by=" + akoUser + "&page_size=100"
	}

	result, err := AviGetCollectionRaw(client, uri)
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

		poolCacheObj := AviPoolCache{
			Name:               *pool.Name,
			Uuid:               *pool.UUID,
			CloudConfigCksum:   *pool.CloudConfigCksum,
			ServiceMetadataObj: svc_mdata_obj,
			LastModified:       *pool.LastModified,
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
		uri = "/api/vsvip/?" + "name.contains=" + lib.GetNamePrefix() + "&include_name=true" + "&cloud_ref.name=" + cloud
	}

	result, err := AviGetCollectionRaw(client, uri)
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

	result, err := AviGetCollectionRaw(client, uri)
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

	result, err := AviGetCollectionRaw(client, uri)
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

	result, err := AviGetCollectionRaw(client, uri)
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
		k := NamespaceName{Namespace: utils.ADMIN_NS, Name: *sslkey.Name}
		c.SSLKeyCache.AviCacheAdd(k, &sslCacheObj)
		utils.AviLog.Debugf("Adding sslkey to Cache during refresh %s\n", k)
	}
	return nil
}

func (c *AviObjCache) AviPopulateOnePoolCache(client *clients.AviClient,
	cloud string, objName string) error {
	var uri string
	akoUser := lib.AKOUser

	uri = "/api/pool?name=" + objName + "&created_by=" + akoUser

	result, err := AviGetCollectionRaw(client, uri)
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

		poolCacheObj := AviPoolCache{
			Name:               *pool.Name,
			Uuid:               *pool.UUID,
			CloudConfigCksum:   *pool.CloudConfigCksum,
			ServiceMetadataObj: svc_mdata_obj,
			LastModified:       *pool.LastModified,
		}
		k := NamespaceName{Namespace: utils.ADMIN_NS, Name: *pool.Name}
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

	result, err := AviGetCollectionRaw(client, uri)
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
		k := NamespaceName{Namespace: utils.ADMIN_NS, Name: *ds.Name}
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

	result, err := AviGetCollectionRaw(client, uri)
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
		k := NamespaceName{Namespace: utils.ADMIN_NS, Name: *pg.Name}
		c.PgCache.AviCacheAdd(k, &pgCacheObj)
		utils.AviLog.Debugf("Adding pg to Cache during refresh %s\n", k)
	}
	return nil
}

func (c *AviObjCache) AviPopulateOneVsVipCache(client *clients.AviClient,
	cloud string, objName string) error {
	var uri string

	uri = "/api/vsvip?name=" + objName + "&cloud_ref.name=" + cloud

	result, err := AviGetCollectionRaw(client, uri)
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

		vsVipCacheObj := AviVSVIPCache{
			Name:         *vsvip.Name,
			Uuid:         *vsvip.UUID,
			FQDNs:        fqdns,
			LastModified: *vsvip.LastModified,
		}
		k := NamespaceName{Namespace: utils.ADMIN_NS, Name: *vsvip.Name}
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

	result, err := AviGetCollectionRaw(client, uri)
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
		k := NamespaceName{Namespace: utils.ADMIN_NS, Name: *httppol.Name}
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

	result, err := AviGetCollectionRaw(client, uri)
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
		if l4pol.L4ConnectionPolicy != nil {
			for _, rule := range l4pol.L4ConnectionPolicy.Rules {
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
			CloudConfigCksum: lib.L4PolicyChecksum(ports),
		}
		k := NamespaceName{Namespace: utils.ADMIN_NS, Name: *l4pol.Name}
		c.L4PolicyCache.AviCacheAdd(k, &l4PolCacheObj)
		utils.AviLog.Debugf("Adding l4pol to Cache during refresh %s\n", k)
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

	result, err := AviGetCollectionRaw(client, uri)
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

	result, err := AviGetCollectionRaw(client, uri)
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

	result, err := AviGetCollectionRaw(client, uri)
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
		if l4pol.L4ConnectionPolicy != nil {
			for _, rule := range l4pol.L4ConnectionPolicy.Rules {
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
			CloudConfigCksum: lib.L4PolicyChecksum(ports),
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
		utils.AviLog.Infof("Adding key to l4 cache :%s", utils.Stringify(l4PolCacheObj))
		c.L4PolicyCache.AviCacheAdd(k, &l4PolData[i])
		delete(l4CacheData, k)
	}
	// // The data that is left in httpCacheData should be explicitly removed
	for key := range l4CacheData {
		utils.AviLog.Debugf("Deleting key from l4policy cache :%s", key)
		c.L4PolicyCache.AviCacheDelete(key)
	}
}

func (c *AviObjCache) AviObjVrfCachePopulate(client *clients.AviClient, cloud string) {
	disableStaticRoute := os.Getenv(lib.DISABLE_STATIC_ROUTE_SYNC)
	if disableStaticRoute == "true" {
		utils.AviLog.Debugf("Static route sync disabled, skipping vrf cache population")
		return
	}
	// Disable static route sync if ako is in  NodePort mode
	if lib.IsNodePortMode() {
		utils.AviLog.Infof("Static route sync disabled in NodePort Mode")
		return
	}
	uri := "/api/vrfcontext?name=" + lib.GetVrf() + "&include_name=true&cloud_ref.name=" + cloud
	vrfList := []*models.VrfContext{}

	result, err := AviGetCollectionRaw(client, uri)
	if err != nil {
		utils.AviLog.Warnf("Get uri %v returned err %v", uri, err)
		return
	}
	elems := make([]json.RawMessage, result.Count)
	err = json.Unmarshal(result.Results, &elems)
	if err != nil {
		utils.AviLog.Warnf("Failed to unmarshal data, err: %v", err)
		return
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

	err := AviGet(client, uri, &rest_response)
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
				k := NamespaceName{Namespace: utils.ADMIN_NS, Name: vs["name"].(string)}
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
				if vs["vip"] != nil && len(vs["vip"].([]interface{})) > 0 {
					vip = (vs["vip"].([]interface{})[0].(map[string]interface{})["ip_address"]).(map[string]interface{})["addr"].(string)
				}
				// Populate the VSVIP cache
				if vs["vsvip_ref"] != nil {
					// find the vsvip name from the vsvip cache
					vsVipUuid := ExtractUuid(vs["vsvip_ref"].(string), "vsvip-.*.#")
					vsVipName, foundVip := c.VSVIPCache.AviCacheGetNameByUuid(vsVipUuid)
					if foundVip {
						vipKey := NamespaceName{Namespace: lib.GetTenant(), Name: vsVipName.(string)}
						vsVipKey = append(vsVipKey, vipKey)
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
									poolKeys = c.AviPGPoolCachePopulate(client, cloud, pgName)
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
								poolKeys = c.AviPGPoolCachePopulate(client, cloud, pgName.(string))
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
									poolKeys = c.AviPGPoolCachePopulate(client, cloud, pgName)
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
				utils.AviLog.Infof("Added VS cache key :%s", utils.Stringify(vsMetaObj))

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
	err := AviGet(client, uri, &rest_response)
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
		k := NamespaceName{Namespace: utils.ADMIN_NS, Name: vsName}
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
				if vs["vip"] != nil && len(vs["vip"].([]interface{})) > 0 {
					vip = (vs["vip"].([]interface{})[0].(map[string]interface{})["ip_address"]).(map[string]interface{})["addr"].(string)
				}
				// Populate the VSVIP cache
				if vs["vsvip_ref"] != nil {
					// find the vsvip name from the vsvip cache
					vsVipUuid := ExtractUuidWithoutHash(vs["vsvip_ref"].(string), "vsvip-.*.")
					vsVipName, foundVip := c.VSVIPCache.AviCacheGetNameByUuid(vsVipUuid)
					if foundVip {
						vipKey := NamespaceName{Namespace: utils.ADMIN_NS, Name: vsVipName.(string)}
						vsVipKey = append(vsVipKey, vipKey)
					}
				}
				if vs["ssl_key_and_certificate_refs"] != nil {
					for _, ssl := range vs["ssl_key_and_certificate_refs"].([]interface{}) {
						// find the sslkey name from the ssl key cache
						sslUuid := ExtractUuidWithoutHash(ssl.(string), "sslkeyandcertificate-.*.")
						sslName, foundssl := c.SSLKeyCache.AviCacheGetNameByUuid(sslUuid)
						if foundssl {
							sslKey := NamespaceName{Namespace: utils.ADMIN_NS, Name: sslName.(string)}
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
								dsKey := NamespaceName{Namespace: utils.ADMIN_NS, Name: dsName.(string)}
								// Fetch the associated PGs with the DS.
								dsObj, _ := c.DSCache.AviCacheGet(dsKey)
								for _, pgName := range dsObj.(*AviDSCache).PoolGroups {
									// For each PG, formulate the key and then populate the pg collection cache
									pgKey := NamespaceName{Namespace: utils.ADMIN_NS, Name: pgName}
									poolgroupKeys = append(poolgroupKeys, pgKey)
									poolKeys = c.AviPGPoolCachePopulate(client, cloud, pgName)
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
								pgKey := NamespaceName{Namespace: utils.ADMIN_NS, Name: pgName.(string)}
								poolgroupKeys = append(poolgroupKeys, pgKey)
								poolKeys = c.AviPGPoolCachePopulate(client, cloud, pgName.(string))
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
								httpKey := NamespaceName{Namespace: utils.ADMIN_NS, Name: httpName.(string)}
								httpObj, _ := c.HTTPPolicyCache.AviCacheGet(httpKey)
								for _, pgName := range httpObj.(*AviHTTPPolicyCache).PoolGroups {
									// For each PG, formulate the key and then populate the pg collection cache
									pgKey := NamespaceName{Namespace: utils.ADMIN_NS, Name: pgName}
									poolgroupKeys = append(poolgroupKeys, pgKey)
									poolKeys = c.AviPGPoolCachePopulate(client, cloud, pgName)
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

func (c *AviObjCache) AviCloudPropertiesPopulate(client *clients.AviClient, cloudName string) {
	uri := "/api/cloud/?name=" + cloudName
	result, err := AviGetCollectionRaw(client, uri)
	if err != nil {
		utils.AviLog.Warnf("CloudProperties Get uri %v returned err %v", uri, err)
		return
	}

	elems := make([]json.RawMessage, result.Count)
	err = json.Unmarshal(result.Results, &elems)
	if err != nil {
		utils.AviLog.Warnf("Failed to unmarshal data, err: %v", err)
		return
	}

	if result.Count != 1 {
		utils.AviLog.Errorf("Cloud details not found for cloud name: %s", cloudName)
		return
	}

	cloud := models.Cloud{}
	if err = json.Unmarshal(elems[0], &cloud); err != nil {
		utils.AviLog.Warnf("Failed to unmarshal cloud data, err: %v", err)
	}

	vtype := *cloud.Vtype
	cloud_obj := &AviCloudPropertyCache{Name: cloudName, VType: vtype}
	if cloud.DNSProviderRef == nil {
		utils.AviLog.Warnf("Cloud does not have a dns_provider_ref configured %v", cloudName)
		return
	}
	dns_uuid := ExtractPattern(*cloud.DNSProviderRef, "ipamdnsproviderprofile-.*")
	subdomains := c.AviDNSPropertyPopulate(client, dns_uuid)
	if subdomains != nil {
		cloud_obj.NSIpamDNS = subdomains
	}

	if cloud.IPAMProviderRef == nil {
		utils.AviLog.Warnf("Cloud does not have a ipam_provider_ref configured %v", cloudName)
		return
	}
	ipam_uuid := ExtractPattern(*cloud.IPAMProviderRef, "ipamdnsproviderprofile-.*")
	ipam := c.AviIPAMPropertyPopulate(client, ipam_uuid)
	if ipam != "" {
		cloud_obj.NSIpam = ipam
	}

	c.CloudKeyCache.AviCacheAdd(cloudName, cloud_obj)
	utils.AviLog.Infof("Added CloudKeyCache cache key %v val %v", cloudName, cloud_obj)
	return
}

func (c *AviObjCache) AviIPAMPropertyPopulate(client *clients.AviClient, ipamUUID string) string {
	var ipamProvider models.IPAMDNSProviderProfile
	uri := "/api/ipamdnsproviderprofile/" + ipamUUID

	err := AviGet(client, uri, &ipamProvider)
	if err != nil {
		utils.AviLog.Warnf("IPAMProperty Get uri %v returned err %v", uri, err)
		return ""
	}

	ipamName := *ipamProvider.Name
	return ipamName
}

func (c *AviObjCache) AviDNSPropertyPopulate(client *clients.AviClient, dnsUUID string) []string {
	var dnsProvider models.IPAMDNSProviderProfile
	var dnsSubDomains []string
	uri := "/api/ipamdnsproviderprofile/" + dnsUUID

	err := AviGet(client, uri, &dnsProvider)
	if err != nil {
		utils.AviLog.Warnf("DNSProperty Get uri %v returned err %v", uri, err)
		return nil
	}

	utils.AviLog.Debugf("DNSProperty Get uri %v returned %v ", uri, dnsProvider.Name)

	dnsProfile := dnsProvider.InternalProfile
	// Support multiple dns profiles.
	for _, dnsProf := range dnsProfile.DNSServiceDomain {
		utils.AviLog.Debugf("Found DNS Domain name: %v", *dnsProf.DomainName)
		dnsSubDomains = append(dnsSubDomains, *dnsProf.DomainName)
	}

	return dnsSubDomains
}

func ValidateUserInput(client *clients.AviClient) bool {
	// add other step0 validation logics here -> isValid := check1 && check2 && ...
	isValid := CheckAndSetCloudType(client) && checkRequiredValuesYaml() && CheckAndSetVRFFromNetwork(client) && CheckPublicCloud(client)
	if !isValid {
		utils.AviLog.Warn("Invalid input detected, syncing will be disabled.")
	}
	return isValid
}

func checkRequiredValuesYaml() bool {
	clusterName := lib.GetClusterName()
	re := regexp.MustCompile("^[a-zA-Z0-9-_]*$")
	if clusterName == "" {
		utils.AviLog.Error("Required param clusterName not specified, syncing will be disabled")
		return false
	} else if len(clusterName) > 32 || !re.MatchString(clusterName) {
		utils.AviLog.Error("clusterName must consist of alphanumeric characters or '-'/'_' (max 32 chars), syncing will be disabled")
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
	aviCMNamespace := lib.AviNS
	if lib.GetNamespaceToSync() != "" {
		aviCMNamespace = lib.GetNamespaceToSync()
	}
	_, err := k8sClient.CoreV1().ConfigMaps(aviCMNamespace).Get(lib.AviConfigMap, metav1.GetOptions{})
	if err != nil {
		utils.AviLog.Errorf("Configmap %s/%s not found, error: %v, syncing will be disabled", lib.AviNS, lib.AviConfigMap, err)
		return false
	}

	return true
}

func CheckAndSetCloudType(client *clients.AviClient) bool {

	uri := "/api/cloud/?include_name&name=" + utils.CloudName
	result, err := AviGetCollectionRaw(client, uri)
	if err != nil {
		utils.AviLog.Warnf("Get uri %v returned err %v", uri, err)
		return false
	}

	if result.Count != 1 {
		utils.AviLog.Warnf("Cloud details not found for cloud name: %s", utils.CloudName)
		return false
	}

	elems := make([]json.RawMessage, result.Count)
	err = json.Unmarshal(result.Results, &elems)
	if err != nil {
		utils.AviLog.Warnf("Failed to unmarshal data, err: %v", err)
		return false
	}

	cloud := models.Cloud{}
	err = json.Unmarshal(elems[0], &cloud)
	if err != nil {
		utils.AviLog.Warnf("Failed to unmarshal data, err: %v", err)
		return false
	}
	vType := *cloud.Vtype

	if vType != lib.CLOUD_VCENTER && !lib.IsNodePortMode() {
		utils.AviLog.Warnf("%v not allowed in ClusterIP mode.", vType)
		return false
	}

	utils.AviLog.Infof("Setting cloud vType: %v", vType)
	lib.SetCloudType(vType)
	return true
}

func CheckPublicCloud(client *clients.AviClient) bool {
	if lib.GetCloudType() == lib.CLOUD_AZURE {
		// Handle all azure cloud validations here
		networkName := lib.GetNetworkName()
		if networkName == "" {
			utils.AviLog.Error("Required param networkName not specified, syncing will be disabled.")
			return false
		}
	}

	return true
}

func CheckAndSetVRFFromNetwork(client *clients.AviClient) bool {

	if lib.GetCloudType() != lib.CLOUD_VCENTER {
		// Need not set VRFContext for public clouds.
		return true
	}

	networkName := lib.GetNetworkName()
	if networkName == "" {
		utils.AviLog.Error("Required param networkName not specified, syncing will be disabled.")
		return false
	}

	uri := "/api/network/?include_name&name=" + networkName + "&cloud_ref.name=" + utils.CloudName
	result, err := AviGetCollectionRaw(client, uri)
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

func ExtractUuidWithoutHash(word, pattern string) string {
	r, _ := regexp.Compile(pattern)
	result := r.FindAllString(word, -1)
	if len(result) == 1 {
		return result[0][:len(result[0])]
	}
	return ""
}

func AviGetCollectionRaw(client *clients.AviClient, uri string) (session.AviCollectionResult, error) {
	result, err := client.AviSession.GetCollectionRaw(uri)
	if err != nil {
		apimodels.RestStatus.UpdateAviApiRestStatus("", err)
		return session.AviCollectionResult{}, err
	}

	apimodels.RestStatus.UpdateAviApiRestStatus(utils.AVIAPI_CONNECTED, nil)
	return result, nil
}

func AviGet(client *clients.AviClient, uri string, response interface{}) error {
	err := client.AviSession.Get(uri, &response)
	if err != nil {
		apimodels.RestStatus.UpdateAviApiRestStatus("", err)
		return err
	}

	apimodels.RestStatus.UpdateAviApiRestStatus(utils.AVIAPI_CONNECTED, nil)
	return nil
}
