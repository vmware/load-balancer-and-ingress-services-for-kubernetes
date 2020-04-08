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
}

func NewAviObjCache() *AviObjCache {
	c := AviObjCache{}
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
	c.PopulatePgDataToCache(client, cloud)
	c.PopulateDSDataToCache(client, cloud)
	c.PopulateSSLKeyToCache(client, cloud)
	c.PopulateHttpPolicySetToCache(client, cloud)
	c.PopulateVsVipDataToCache(client, cloud)
	c.PopulatePoolsToCache(client, cloud)
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
	vsCacheCopy := c.VsCache.ShallowCopy()
	var allKeys []interface{}
	for k := range vsCacheCopy {
		allKeys = append(allKeys, k)
	}
	err := c.AviObjVSCachePopulate(client, cloud, vsCacheCopy)
	// Delete all the VS keys that are left in the copy.
	if err != nil {
		for key := range vsCacheCopy {
			utils.AviLog.Info.Printf("Removing vs key from cache: %s", key)
			// We want to synthesize these keys to layer 3.
			deletedKeys = append(deletedKeys, key)
			c.VsCache.AviCacheDelete(key)
		}
	}
	c.AviCloudPropertiesPopulate(client, cloud)
	c.AviObjVrfCachePopulate(client, cloud)
	return deletedKeys, allKeys
}

func (c *AviObjCache) AviPopulateAllPGs(client *clients.AviClient,
	cloud string, pgData *[]AviPGCache, override_uri ...NextPage) (*[]AviPGCache, error) {
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
		return nil, err
	}
	elems := make([]json.RawMessage, result.Count)
	err = json.Unmarshal(result.Results, &elems)
	if err != nil {
		utils.AviLog.Warning.Printf("Failed to unmarshal pg data, err: %v", err)
		return nil, err
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

		pgCacheObj := AviPGCache{
			Name:             *pg.Name,
			Uuid:             *pg.UUID,
			CloudConfigCksum: *pg.CloudConfigCksum,
			LastModified:     *pg.LastModified,
		}
		*pgData = append(*pgData, pgCacheObj)
	}
	if result.Next != "" {
		// It has a next page, let's recursively call the same method.
		next_uri := strings.Split(result.Next, "/api/poolgroup")
		if len(next_uri) > 1 {
			override_uri := "/api/poolgroup" + next_uri[1]
			nextPage := NextPage{Next_uri: override_uri}
			c.AviPopulateAllPGs(client, cloud, pgData, nextPage)
		}
	}
	return pgData, nil
}

func (c *AviObjCache) PopulatePgDataToCache(client *clients.AviClient,
	cloud string) {

	var pgData []AviPGCache
	_, err := c.AviPopulateAllPGs(client, cloud, &pgData)
	if err != nil {
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
	cloud string, poolData *[]AviPoolCache, override_uri ...NextPage) (*[]AviPoolCache, error) {
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
		return nil, err
	}
	elems := make([]json.RawMessage, result.Count)
	err = json.Unmarshal(result.Results, &elems)
	if err != nil {
		return nil, err
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
			c.AviPopulateAllPools(client, cloud, poolData, nextPage)
		}
	}

	return poolData, nil
}

func (c *AviObjCache) PopulatePoolsToCache(client *clients.AviClient,
	cloud string, override_uri ...NextPage) {
	var poolsData []AviPoolCache
	_, err := c.AviPopulateAllPools(client, cloud, &poolsData)
	if err != nil {
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
	cloud string, vsVipData *[]AviVSVIPCache, nextPage ...NextPage) (*[]AviVSVIPCache, error) {
	var uri string
	if len(nextPage) == 1 {
		uri = nextPage[0].Next_uri
	} else {
		uri = "/api/vsvip?include_name=true&cloud_ref.name=" + cloud
	}
	result, err := client.AviSession.GetCollectionRaw(uri)
	if err != nil {
		utils.AviLog.Warning.Printf("Get uri %v returned err for vsvip %v", uri, err)
		return nil, err
	}
	elems := make([]json.RawMessage, result.Count)
	err = json.Unmarshal(result.Results, &elems)
	if err != nil {
		utils.AviLog.Warning.Printf("Failed to unmarshal vsvip data, err: %v", err)
		return nil, err
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

		vsVipCacheObj := AviVSVIPCache{
			Name:             *vsvip.Name,
			Uuid:             *vsvip.UUID,
			CloudConfigCksum: *vsvip.VsvipCloudConfigCksum,
			LastModified:     *vsvip.LastModified,
		}
		*vsVipData = append(*vsVipData, vsVipCacheObj)

	}
	if result.Next != "" {
		// It has a next page, let's recursively call the same method.
		next_uri := strings.Split(result.Next, "/api/vsvip")
		if len(next_uri) > 1 {
			override_uri := "/api/vsvip" + next_uri[1]
			nextPage := NextPage{Next_uri: override_uri}
			c.AviPopulateAllVSVips(client, cloud, vsVipData, nextPage)
		}
	}
	return vsVipData, nil
}

func (c *AviObjCache) PopulateVsVipDataToCache(client *clients.AviClient,
	cloud string) {
	var vsVipData []AviVSVIPCache
	_, err := c.AviPopulateAllVSVips(client, cloud, &vsVipData)
	if err != nil {
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
		utils.AviLog.Info.Printf("Adding key to vsvip cache :%s", k)
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
	cloud string, DsData *[]AviDSCache, nextPage ...NextPage) (*[]AviDSCache, error) {
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
		return nil, err
	}
	elems := make([]json.RawMessage, result.Count)
	err = json.Unmarshal(result.Results, &elems)
	if err != nil {
		utils.AviLog.Warning.Printf("Failed to unmarshal datascript data, err: %v", err)
		return nil, err
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
		dsCacheObj := AviDSCache{
			Name: *ds.Name,
			Uuid: *ds.UUID,
		}
		*DsData = append(*DsData, dsCacheObj)
	}
	if result.Next != "" {
		// It has a next page, let's recursively call the same method.
		next_uri := strings.Split(result.Next, "/api/vsdatascriptset")
		if len(next_uri) > 1 {
			override_uri := "/api/vsdatascriptset" + next_uri[1]
			nextPage := NextPage{Next_uri: override_uri}
			c.AviPopulateAllDSs(client, cloud, DsData, nextPage)
		}
	}
	return DsData, nil
}

func (c *AviObjCache) PopulateDSDataToCache(client *clients.AviClient,
	cloud string, override_uri ...NextPage) {
	var DsData []AviDSCache
	_, err := c.AviPopulateAllDSs(client, cloud, &DsData)
	dsCacheData := c.DSCache.ShallowCopy()
	if err != nil {
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
	cloud string, SslData *[]AviSSLCache, nextPage ...NextPage) (*[]AviSSLCache, error) {
	var uri string
	akcUser := utils.OSHIFT_K8S_CLOUD_CONNECTOR
	if len(nextPage) == 1 {
		uri = nextPage[0].Next_uri
	} else {
		uri = "/api/sslkeyandcertificate?include_name=true&cloud_ref.name=" + cloud + "&created_by=" + akcUser
	}
	result, err := client.AviSession.GetCollectionRaw(uri)
	if err != nil {
		utils.AviLog.Warning.Printf("Get uri %v returned err for sslkeyandcertificate %v", uri, err)
		return nil, err
	}
	elems := make([]json.RawMessage, result.Count)
	err = json.Unmarshal(result.Results, &elems)
	if err != nil {
		utils.AviLog.Warning.Printf("Failed to unmarshal sslkeyandcertificate data, err: %v", err)
		return nil, err
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
			c.AviPopulateAllSSLKeys(client, cloud, SslData, nextPage)
		}
	}
	return SslData, nil
}

func (c *AviObjCache) PopulateSSLKeyToCache(client *clients.AviClient,
	cloud string, override_uri ...NextPage) {
	var SslKeyData []AviSSLCache
	_, err := c.AviPopulateAllSSLKeys(client, cloud, &SslKeyData)
	if err != nil {
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
	// The data that is left in sslCacheData should be explicitly removed
	for key := range sslCacheData {
		utils.AviLog.Info.Printf("Deleting key from sslkey cache :%s", key)
		c.SSLKeyCache.AviCacheDelete(key)
	}
}

func (c *AviObjCache) AviPopulateAllHttpPolicySets(client *clients.AviClient,
	cloud string, httpPolicyData *[]AviHTTPPolicyCache, nextPage ...NextPage) (*[]AviHTTPPolicyCache, error) {
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
		return nil, err
	}
	elems := make([]json.RawMessage, result.Count)
	err = json.Unmarshal(result.Results, &elems)
	if err != nil {
		utils.AviLog.Warning.Printf("Failed to unmarshal httppolicyset data, err: %v", err)
		return nil, err
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
		httpPolCacheObj := AviHTTPPolicyCache{
			Name:             *httppol.Name,
			Uuid:             *httppol.UUID,
			CloudConfigCksum: *httppol.CloudConfigCksum,
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
			c.AviPopulateAllHttpPolicySets(client, cloud, httpPolicyData, nextPage)
		}
	}
	return httpPolicyData, nil
}

func (c *AviObjCache) PopulateHttpPolicySetToCache(client *clients.AviClient,
	cloud string, override_uri ...NextPage) {
	var HttPolData []AviHTTPPolicyCache
	_, err := c.AviPopulateAllHttpPolicySets(client, cloud, &HttPolData)
	if err != nil {
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
	// The data that is left in httpCacheData should be explicitly removed
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
	uri := "/api/vrfcontext?include_name=true&cloud_ref.name=" + cloud

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

// TODO (sudswas): Should this be run inside a go routine for parallel population
// to reduce bootup time when the system is loaded. Variable duplication expected.
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
			var svc_mdata interface{}
			var svc_mdata_map map[string]interface{}
			if ok {
				if err := json.Unmarshal([]byte(svc_mdata_intf.(string)),
					&svc_mdata); err == nil {
					svc_mdata_map, ok = svc_mdata.(map[string]interface{})
					if !ok {
						utils.AviLog.Warning.Printf(`resp %v svc_mdata %T has invalid
								 service_metadata type for vs`, vs, svc_mdata)
					} else {
						svcName, ok := svc_mdata_map["svc_name"]
						if ok {
							svc_mdata_obj.ServiceName = svcName.(string)
						} else {
							utils.AviLog.Warning.Printf(`service_metadata %v 
									  malformed for vs`, svc_mdata_map)
						}
						namespace, ok := svc_mdata_map["namespace"]
						if ok {
							svc_mdata_obj.Namespace = namespace.(string)
						} else {
							utils.AviLog.Warning.Printf(`service_metadata %v 
									  malformed for vs`, svc_mdata_map)
						}
					}
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
				vs_uuid := ExtractVsUuid(vs_parent_ref.(string))
				utils.AviLog.Info.Printf("extracted the vs uuid from parent ref during cache population: %s", vs_uuid)
				// Now let's get the VS key from this uuid
				vsKey, gotVS := c.VsCache.AviCacheGetKeyByUuid(vs_uuid)
				if gotVS {
					parentVSKey = vsKey.(NamespaceName)
				}

			}
			if vs["cloud_config_cksum"] != nil {
				k := NamespaceName{Namespace: utils.ADMIN_NS, Name: vs["name"].(string)}
				delete(vsCacheCopy, k)
				var vip string
				if vs["vip"] != nil && len(vs["vip"].([]interface{})) > 0 {
					vip = (vs["vip"].([]interface{})[0].(map[string]interface{})["ip_address"]).(map[string]interface{})["addr"].(string)
				}
				vs_cache, found := c.VsCache.AviCacheGet(k)
				if found {
					vs_cache_obj, ok := vs_cache.(*AviVsCache)
					if ok {
						if vs_cache_obj.Uuid == vs["uuid"].(string) {
							// Same object - let's just refresh the values.
							vs_cache_obj.CloudConfigCksum = vs["cloud_config_cksum"].(string)
							vs_cache_obj.SNIChildCollection = sni_child_collection
							vs_cache_obj.ParentVSRef = parentVSKey
							newLastModified := vs["_last_modified"].(string)
							if vs_cache_obj.LastModified != "" && vs_cache_obj.LastModified != newLastModified {
								utils.AviLog.Warning.Printf("Invalid cache data for vs: %s", k)
								vs_cache_obj.InvalidData = true
							}
							vs_cache_obj.LastModified = newLastModified
							utils.AviLog.Info.Printf("Updated Vs cache k %v val %v",
								k, vs_cache_obj)
						} else {
							// New object
							vs_cache_obj := AviVsCache{Name: vs["name"].(string),
								Tenant: utils.ADMIN_NS, Uuid: vs["uuid"].(string), Vip: vip,
								CloudConfigCksum:   vs["cloud_config_cksum"].(string),
								SNIChildCollection: sni_child_collection,
								ParentVSRef:        parentVSKey,
								ServiceMetadataObj: svc_mdata_obj}

							c.VsCache.AviCacheAdd(k, &vs_cache_obj)
							utils.AviLog.Info.Printf("Added Vs cache k %v val %v",
								k, vs_cache_obj)
						}
					} else {
						// New object
						vs_cache_obj := AviVsCache{Name: vs["name"].(string),
							Tenant: utils.ADMIN_NS, Uuid: vs["uuid"].(string), Vip: vip,
							CloudConfigCksum:   vs["cloud_config_cksum"].(string),
							SNIChildCollection: sni_child_collection,
							ParentVSRef:        parentVSKey,
							ServiceMetadataObj: svc_mdata_obj}

						c.VsCache.AviCacheAdd(k, &vs_cache_obj)
						utils.AviLog.Info.Printf("Added Vs cache k %v val %v",
							k, vs_cache_obj)
					}
				} else {
					vs_cache_obj := AviVsCache{Name: vs["name"].(string),
						Tenant: utils.ADMIN_NS, Uuid: vs["uuid"].(string), Vip: vip,
						CloudConfigCksum:   vs["cloud_config_cksum"].(string),
						SNIChildCollection: sni_child_collection,
						ParentVSRef:        parentVSKey,
						ServiceMetadataObj: svc_mdata_obj}

					c.VsCache.AviCacheAdd(k, &vs_cache_obj)
					utils.AviLog.Info.Printf("Added Vs cache k %v val %v",
						k, vs_cache_obj)
				}
				c.AviHTTPolicyCachePopulate(client, cloud, vs["uuid"].(string), utils.ADMIN_NS, k)
				c.AviPGCachePopulate(client, cloud, vs["uuid"].(string), utils.ADMIN_NS, k)
				c.AviPoolCachePopulate(client, cloud, vs["uuid"].(string), utils.ADMIN_NS, k)
				c.AviDataScriptPopulate(client, cloud, vs["uuid"].(string), utils.ADMIN_NS, k)
				c.AviSSLKeyCachePopulate(client, cloud, vs["uuid"].(string), utils.ADMIN_NS, k)
				c.AviVSVIPCachePopulate(client, cloud, vs["uuid"].(string), utils.ADMIN_NS, k)
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
		var svc_mdata interface{}
		var svc_mdata_map map[string]interface{}
		if ok {
			if err := json.Unmarshal([]byte(svc_mdata_intf.(string)),
				&svc_mdata); err == nil {
				svc_mdata_map, ok = svc_mdata.(map[string]interface{})
				if !ok {
					utils.AviLog.Warning.Printf(`resp %v svc_mdata %T has invalid
								 service_metadata type for vs`, vs, svc_mdata)
				} else {
					svcName, ok := svc_mdata_map["svc_name"]
					if ok {
						svc_mdata_obj.ServiceName = svcName.(string)
					} else {
						utils.AviLog.Warning.Printf(`service_metadata %v 
									  malformed for vs`, svc_mdata_map)
					}
					namespace, ok := svc_mdata_map["namespace"]
					if ok {
						svc_mdata_obj.Namespace = namespace.(string)
					} else {
						utils.AviLog.Warning.Printf(`service_metadata %v 
									  malformed for vs`, svc_mdata_map)
					}
				}
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
			vs_uuid := ExtractVsUuid(vs_parent_ref.(string))
			utils.AviLog.Info.Printf("extracted the vs uuid from parent ref during cache population: %s", vs_uuid)
			// Now let's get the VS key from this uuid
			vsKey, gotVS := c.VsCache.AviCacheGetKeyByUuid(vs_uuid)
			if gotVS {
				parentVSKey = vsKey.(NamespaceName)
			}

		}
		if vs["cloud_config_cksum"] != nil {
			k := NamespaceName{Namespace: utils.ADMIN_NS, Name: vs["name"].(string)}
			var vip string
			if vs["vip"] != nil && len(vs["vip"].([]interface{})) > 0 {
				vip = (vs["vip"].([]interface{})[0].(map[string]interface{})["ip_address"]).(map[string]interface{})["addr"].(string)
			}
			vs_cache, found := c.VsCache.AviCacheGet(k)
			if found {
				vs_cache_obj, ok := vs_cache.(*AviVsCache)
				if ok {
					if vs_cache_obj.Uuid == vs["uuid"].(string) {
						// Same object - let's just refresh the values.
						vs_cache_obj.CloudConfigCksum = vs["cloud_config_cksum"].(string)
						vs_cache_obj.SNIChildCollection = sni_child_collection
						vs_cache_obj.ParentVSRef = parentVSKey
						utils.AviLog.Info.Printf("Updated Vs cache k %v val %v",
							k, vs_cache_obj)
					} else {
						// New object
						vs_cache_obj := AviVsCache{Name: vs["name"].(string),
							Tenant: utils.ADMIN_NS, Uuid: vs["uuid"].(string), Vip: vip,
							CloudConfigCksum:   vs["cloud_config_cksum"].(string),
							SNIChildCollection: sni_child_collection,
							ParentVSRef:        parentVSKey,
							ServiceMetadataObj: svc_mdata_obj}

						c.VsCache.AviCacheAdd(k, &vs_cache_obj)
						utils.AviLog.Info.Printf("Added Vs cache k %v val %v",
							k, vs_cache_obj)
					}
				} else {
					// New object
					vs_cache_obj := AviVsCache{Name: vs["name"].(string),
						Tenant: utils.ADMIN_NS, Uuid: vs["uuid"].(string), Vip: vip,
						CloudConfigCksum:   vs["cloud_config_cksum"].(string),
						SNIChildCollection: sni_child_collection,
						ParentVSRef:        parentVSKey,
						ServiceMetadataObj: svc_mdata_obj}

					c.VsCache.AviCacheAdd(k, &vs_cache_obj)
					utils.AviLog.Info.Printf("Added Vs cache k %v val %v",
						k, vs_cache_obj)
				}
			} else {
				vs_cache_obj := AviVsCache{Name: vs["name"].(string),
					Tenant: utils.ADMIN_NS, Uuid: vs["uuid"].(string), Vip: vip,
					CloudConfigCksum:   vs["cloud_config_cksum"].(string),
					SNIChildCollection: sni_child_collection,
					ParentVSRef:        parentVSKey,
					ServiceMetadataObj: svc_mdata_obj}

				c.VsCache.AviCacheAdd(k, &vs_cache_obj)
				utils.AviLog.Info.Printf("Added Vs cache k %v val %v",
					k, vs_cache_obj)
			}
			c.AviHTTPolicyCachePopulate(client, cloud, vs["uuid"].(string), utils.ADMIN_NS, k)
			c.AviPGCachePopulate(client, cloud, vs["uuid"].(string), utils.ADMIN_NS, k)
			c.AviPoolCachePopulate(client, cloud, vs["uuid"].(string), utils.ADMIN_NS, k)
			c.AviDataScriptPopulate(client, cloud, vs["uuid"].(string), utils.ADMIN_NS, k)
			c.AviSSLKeyCachePopulate(client, cloud, vs["uuid"].(string), utils.ADMIN_NS, k)
			c.AviVSVIPCachePopulate(client, cloud, vs["uuid"].(string), utils.ADMIN_NS, k)
		}
	}
	return nil
}

//Design library methods to remove repeatation of code.
func (c *AviObjCache) AviPGCachePopulate(client *clients.AviClient,
	cloud string, vs_uuid string, tenant string, vsKey NamespaceName, nextPage ...NextPage) {
	var rest_response interface{}
	var pg_key_collection []NamespaceName
	var uri string
	akcUser := utils.OSHIFT_K8S_CLOUD_CONNECTOR

	if len(nextPage) == 1 {
		uri = nextPage[0].Next_uri
		pg_key_collection = nextPage[0].Collection.([]NamespaceName)
	} else {
		uri = "/api/poolgroup?include_name=true&cloud_ref.name=" + cloud + "&referred_by=virtualservice:" + vs_uuid + "&created_by=" + akcUser
	}
	err := client.AviSession.Get(uri, &rest_response)
	if err != nil {
		utils.AviLog.Warning.Printf("PG Get uri %v returned err %v", uri, err)
		return
	} else {
		resp, ok := rest_response.(map[string]interface{})
		if !ok {
			utils.AviLog.Warning.Printf("PG Get uri %v returned %v type %T", uri,
				rest_response, rest_response)
			return
		}
		utils.AviLog.Info.Printf("PG Get uri %v returned %v PGs", uri,
			resp["count"])
		results, ok := resp["results"].([]interface{})
		if !ok {
			utils.AviLog.Warning.Printf("results not of type []interface{} Instead of type %T for PGs", resp["results"])
			return
		}
		for _, pg_intf := range results {
			pg, ok := pg_intf.(map[string]interface{})
			if !ok {
				utils.AviLog.Warning.Printf("pg_intf not of type map[string] interface{}. Instead of type %T", pg_intf)
				continue
			}
			if pg["cloud_config_cksum"] != nil {
				k := NamespaceName{Namespace: tenant, Name: pg["name"].(string)}
				utils.AviLog.Info.Printf("Added PG cache key %v to VS", k)
				pg_key_collection = append(pg_key_collection, k)
			}

		}
		vs_cache, found := c.VsCache.AviCacheGet(vsKey)
		if found {
			vs_cache_obj, ok := vs_cache.(*AviVsCache)
			if !ok {
				utils.AviLog.Warning.Printf("Unable to cast to VS object: %v", vsKey)
				return
			}
			vs_cache_obj.PGKeyCollection = pg_key_collection
		} else {
			utils.AviLog.Warning.Printf("VS cache not found for key: %v . Unable to update PG collection", vsKey)
		}
		if resp["next"] != nil {
			// It has a next page, let's recursively call the same method.
			next_uri := strings.Split(resp["next"].(string), "/api/poolgroup")
			utils.AviLog.Info.Printf("Found next page for pg, uri: %s", next_uri)
			if len(next_uri) > 1 {
				override_uri := "/api/poolgroup" + next_uri[1]
				utils.AviLog.Info.Printf("Next page uri for pg: %s", override_uri)
				nextPage := NextPage{Next_uri: override_uri}
				c.AviPGCachePopulate(client, cloud, vs_uuid, tenant, vsKey, nextPage)
			}
		}
	}
}

func (c *AviObjCache) AviVSVIPCachePopulate(client *clients.AviClient,
	cloud string, vs_uuid string, tenant string, vsKey NamespaceName, nextPage ...NextPage) {
	var rest_response interface{}
	var vsvip_key_collection []NamespaceName
	var uri string
	if len(nextPage) == 1 {
		uri = nextPage[0].Next_uri
		vsvip_key_collection = nextPage[0].Collection.([]NamespaceName)
	} else {
		uri = "/api/vsvip?include_name=true&cloud_ref.name=" + cloud + "&referred_by=virtualservice:" + vs_uuid
	}
	err := client.AviSession.Get(uri, &rest_response)
	if err != nil {
		utils.AviLog.Warning.Printf("VSVIP Get uri %v returned err %v", uri, err)
		return
	} else {
		resp, ok := rest_response.(map[string]interface{})
		if !ok {
			utils.AviLog.Warning.Printf("VSVIP Get uri %v returned %v type %T", uri,
				rest_response, rest_response)
			return
		}
		utils.AviLog.Info.Printf("VSVIP Get uri %v returned %v VSVIPs", uri,
			resp["count"])
		results, ok := resp["results"].([]interface{})
		if !ok {
			utils.AviLog.Warning.Printf("results not of type []interface{} Instead of type %T for VSVIP", resp["results"])
			return
		}
		for _, vsvip_intf := range results {
			vsvip, ok := vsvip_intf.(map[string]interface{})
			var fqdns []string
			if !ok {
				utils.AviLog.Warning.Printf("vsvip_intf not of type map[string] interface{}. Instead of type %T", vsvip_intf)
				continue
			}
			if vsvip["dns_info"] != nil {
				for _, aRecord := range vsvip["dns_info"].([]interface{}) {
					aRecordMap, success := aRecord.(map[string]interface{})
					if success {
						fqdn, ok := aRecordMap["fqdn"].(string)
						if ok {
							fqdns = append(fqdns, fqdn)
						}
					}
				}
			}
			vsvip_cache_obj := AviVSVIPCache{Name: vsvip["name"].(string),
				Tenant: tenant, Uuid: vsvip["uuid"].(string), FQDNs: fqdns,
			}
			k := NamespaceName{Namespace: tenant, Name: vsvip["name"].(string)}
			c.VSVIPCache.AviCacheAdd(k, &vsvip_cache_obj)
			utils.AviLog.Info.Printf("Added VSVIP cache key %v val %v",
				k, vsvip_cache_obj)
			vsvip_key_collection = append(vsvip_key_collection, k)

		}
		vs_cache, found := c.VsCache.AviCacheGet(vsKey)
		if found {
			vs_cache_obj, ok := vs_cache.(*AviVsCache)
			if !ok {
				utils.AviLog.Warning.Printf("Unable to cast to VS object: %v", vsKey)
				return
			}
			vs_cache_obj.VSVipKeyCollection = vsvip_key_collection
		} else {
			utils.AviLog.Warning.Printf("VS cache not found for key: %v . Unable to update VSVIP collection", vsKey)
			return
		}
		if resp["next"] != nil {
			// It has a next page, let's recursively call the same method.
			next_uri := strings.Split(resp["next"].(string), "/api/vsvip")
			utils.AviLog.Info.Printf("Found next page for vsvip, uri: %s", next_uri)
			if len(next_uri) > 1 {
				override_uri := "/api/vsvip" + next_uri[1]
				utils.AviLog.Info.Printf("Next page uri for vsvip: %s", override_uri)
				nextPage := NextPage{Next_uri: override_uri}
				c.AviVSVIPCachePopulate(client, cloud, vs_uuid, tenant, vsKey, nextPage)
			}
		}
	}
}

func (c *AviObjCache) AviHTTPolicyCachePopulate(client *clients.AviClient,
	cloud string, vs_uuid string, tenant string, vsKey NamespaceName, nextPage ...NextPage) {
	var rest_response interface{}
	var http_pol_key_collection []NamespaceName
	var uri string
	akcUser := utils.OSHIFT_K8S_CLOUD_CONNECTOR

	if len(nextPage) == 1 {
		uri = nextPage[0].Next_uri
		http_pol_key_collection = nextPage[0].Collection.([]NamespaceName)
	} else {
		uri = "/api/httppolicyset?include_name=true&referred_by=virtualservice:" + vs_uuid + "&created_by=" + akcUser
	}
	err := client.AviSession.Get(uri, &rest_response)
	if err != nil {
		utils.AviLog.Warning.Printf("HTTP Policy Get uri %v returned err %v", uri, err)
		return
	} else {
		resp, ok := rest_response.(map[string]interface{})
		if !ok {
			utils.AviLog.Warning.Printf("HTTP Policy Get uri %v returned %v type %T", uri,
				rest_response, rest_response)
			return
		}
		utils.AviLog.Info.Printf("HTTP Policy Get uri %v returned %v HTTP policy objects", uri,
			resp["count"])
		results, ok := resp["results"].([]interface{})
		if !ok {
			utils.AviLog.Warning.Printf("results not of type []interface{} Instead of type %T for HTTP Policy", resp["results"])
			return
		}
		for _, pol_intf := range results {
			pol, ok := pol_intf.(map[string]interface{})
			if !ok {
				utils.AviLog.Warning.Printf("http_intf not of type map[string] interface{}. Instead of type %T", pol_intf)
				continue
			}
			if pol["cloud_config_cksum"] == nil {
				utils.AviLog.Warning.Printf("http policy object has no checksum: %s", pol)
				continue
			}
			k := NamespaceName{Namespace: tenant, Name: pol["name"].(string)}
			utils.AviLog.Info.Printf("Added HTTP Policy cache key %v",
				k)
			http_pol_key_collection = append(http_pol_key_collection, k)

		}
		vs_cache, found := c.VsCache.AviCacheGet(vsKey)
		if found {
			vs_cache_obj, ok := vs_cache.(*AviVsCache)
			if !ok {
				utils.AviLog.Warning.Printf("Unable to cast to VS object: %v", vsKey)
				return
			}
			vs_cache_obj.HTTPKeyCollection = http_pol_key_collection
		} else {
			utils.AviLog.Warning.Printf("VS cache not found for key: %v . Unable to update HTTP Policy collection", vsKey)
		}
		if resp["next"] != nil {
			// It has a next page, let's recursively call the same method.
			next_uri := strings.Split(resp["next"].(string), "/api/httppolicyset")
			utils.AviLog.Info.Printf("Found next page for http policyset objs, uri: %s", next_uri)
			if len(next_uri) > 1 {
				override_uri := "/api/httppolicyset" + next_uri[1]
				utils.AviLog.Info.Printf("Next page uri for http policyset objs: %s", override_uri)
				nextPage := NextPage{Next_uri: override_uri}
				c.AviHTTPolicyCachePopulate(client, cloud, vs_uuid, tenant, vsKey, nextPage)
			}
		}
	}
}

func (c *AviObjCache) AviSSLKeyCachePopulate(client *clients.AviClient,
	cloud string, vs_uuid string, tenant string, vsKey NamespaceName, nextPage ...NextPage) {
	var rest_response interface{}
	var ssl_key_collection []NamespaceName
	var uri string
	akcUser := utils.OSHIFT_K8S_CLOUD_CONNECTOR

	if len(nextPage) == 1 {
		uri = nextPage[0].Next_uri
		ssl_key_collection = nextPage[0].Collection.([]NamespaceName)
	} else {
		uri = "/api/sslkeyandcertificate?include_name=true&referred_by=virtualservice:" + vs_uuid + "&created_by=" + akcUser
	}
	err := client.AviSession.Get(uri, &rest_response)
	if err != nil {
		utils.AviLog.Warning.Printf("SSL Keys Get uri %v returned err %v", uri, err)
		return
	} else {
		resp, ok := rest_response.(map[string]interface{})
		if !ok {
			utils.AviLog.Warning.Printf("SSL Keys Get uri %v returned %v type %T", uri,
				rest_response, rest_response)
			return
		}
		utils.AviLog.Info.Printf("SSL Keys Get uri %v returned %v SSL key", uri,
			resp["count"])
		results, ok := resp["results"].([]interface{})
		if !ok {
			utils.AviLog.Warning.Printf("results not of type []interface{} Instead of type %T for SSL Keys", resp["results"])
			return
		}
		for _, ssl_intf := range results {
			ssl, ok := ssl_intf.(map[string]interface{})
			if !ok {
				utils.AviLog.Warning.Printf("ssl_intf not of type map[string] interface{}. Instead of type %T", ssl_intf)
				continue
			}

			k := NamespaceName{Namespace: tenant, Name: ssl["name"].(string)}
			utils.AviLog.Info.Printf("Added SSL Key cache key %v",
				k)
			ssl_key_collection = append(ssl_key_collection, k)

		}
		vs_cache, found := c.VsCache.AviCacheGet(vsKey)
		if found {
			vs_cache_obj, ok := vs_cache.(*AviVsCache)
			if !ok {
				utils.AviLog.Warning.Printf("Unable to cast to VS object: %v", vsKey)
				return
			}
			vs_cache_obj.SSLKeyCertCollection = ssl_key_collection
		} else {
			utils.AviLog.Warning.Printf("VS cache not found for key: %v . Unable to update HTTP Policy collection", vsKey)
		}
		if resp["next"] != nil {
			// It has a next page, let's recursively call the same method.
			next_uri := strings.Split(resp["next"].(string), "/api/sslkeyandcertificate")
			utils.AviLog.Info.Printf("Found next page for ssl key objs, uri: %s", next_uri)
			if len(next_uri) > 1 {
				override_uri := "/api/sslkeyandcertificate" + next_uri[1]
				utils.AviLog.Info.Printf("Next page uri for ssl key objs: %s", override_uri)
				nextPage := NextPage{Next_uri: override_uri}
				c.AviHTTPolicyCachePopulate(client, cloud, vs_uuid, tenant, vsKey, nextPage)
			}
		}
	}
}

func (c *AviObjCache) AviPoolCachePopulate(client *clients.AviClient,
	cloud string, vs_uuid string, tenant string, vsKey NamespaceName, nextPage ...NextPage) {
	var rest_response interface{}
	var err error
	var pool_key_collection []NamespaceName
	var uri string
	akcUser := utils.OSHIFT_K8S_CLOUD_CONNECTOR

	if len(nextPage) == 1 {
		uri = nextPage[0].Next_uri
		pool_key_collection = nextPage[0].Collection.([]NamespaceName)
	} else {
		uri = "/api/pool?include_name=true&cloud_ref.name=" + cloud + "&referred_by=virtualservice:" + vs_uuid + "&created_by=" + akcUser
	}
	err = client.AviSession.Get(uri, &rest_response)

	if err != nil {
		utils.AviLog.Warning.Printf("Pool Get uri %v returned err %v", uri, err)
	} else {
		resp, ok := rest_response.(map[string]interface{})
		if !ok {
			utils.AviLog.Warning.Printf("Pool Get uri %v returned %v type %T", uri,
				rest_response, rest_response)
			return
		}
		utils.AviLog.Info.Printf("Pool Get uri %v returned %v pools", uri,
			resp["count"])
		results, ok := resp["results"].([]interface{})
		if !ok {
			utils.AviLog.Warning.Printf("results not of type []interface{} Instead of type %T", resp["results"])
			return
		}
		for _, pool_intf := range results {
			pool, ok := pool_intf.(map[string]interface{})
			if !ok {
				utils.AviLog.Warning.Printf("pool_intf not of type map[string] interface{}. Instead of type %T", pool_intf)
				continue
			}
			svc_mdata_intf, ok := pool["service_metadata"]
			var svc_mdata_obj ServiceMetadataObj
			var svc_mdata interface{}
			var svc_mdata_map map[string]interface{}
			if ok {
				if err := json.Unmarshal([]byte(svc_mdata_intf.(string)),
					&svc_mdata); err == nil {
					svc_mdata_map, ok = svc_mdata.(map[string]interface{})
					if !ok {
						utils.AviLog.Warning.Printf(`resp %v svc_mdata %T has invalid
								 service_metadata type`, pool, svc_mdata)
					} else {
						ingressName, ok := svc_mdata_map["ingress_name"]
						if ok {
							svc_mdata_obj.IngressName = ingressName.(string)
						} else {
							utils.AviLog.Warning.Printf(`service_metadata %v 
									  malformed`, svc_mdata_map)
						}
						namespace, ok := svc_mdata_map["namespace"]
						if ok {
							svc_mdata_obj.Namespace = namespace.(string)
						} else {
							utils.AviLog.Warning.Printf(`service_metadata %v 
									  malformed`, svc_mdata_map)
						}
					}
				}
			} else {
				utils.AviLog.Warning.Printf("service_metadata %v malformed", pool)
				// Not caching a pool with malformed metadata?
				continue
			}
			if pool["cloud_config_cksum"] == nil {
				utils.AviLog.Warning.Printf("pool object has no checksum: %v", pool)
				continue
			}
			k := NamespaceName{Namespace: tenant, Name: pool["name"].(string)}
			pool_key_collection = append(pool_key_collection, k)
			utils.AviLog.Info.Printf("Added Pool cache key %v val.",
				k)
		}
		vs_cache, found := c.VsCache.AviCacheGet(vsKey)
		if found {
			vs_cache_obj, ok := vs_cache.(*AviVsCache)
			if !ok {
				utils.AviLog.Warning.Printf("Unable to cast to VS object: %v", vsKey)
				return
			}
			vs_cache_obj.PoolKeyCollection = pool_key_collection
		} else {
			utils.AviLog.Warning.Printf("VS cache not found for key: %v . Unable to update Pool collection", vsKey)
		}
		if resp["next"] != nil {
			// It has a next page, let's recursively call the same method.
			next_uri := strings.Split(resp["next"].(string), "/api/pool")
			utils.AviLog.Info.Printf("Found next page, uri for pool during VS cache population: %s", next_uri)
			if len(next_uri) > 1 {
				override_uri := "/api/pool" + next_uri[1]
				nextPage := NextPage{Next_uri: override_uri, Collection: pool_key_collection}
				c.AviPoolCachePopulate(client, cloud, vs_uuid, tenant, vsKey, nextPage)
			}
		}
	}
	return
}

func (c *AviObjCache) AviDataScriptPopulate(client *clients.AviClient,
	cloud string, vs_uuid string, tenant string, vsKey NamespaceName, nextPage ...NextPage) {
	var rest_response interface{}
	var err error
	var ds_key_collection []NamespaceName
	akcUser := utils.OSHIFT_K8S_CLOUD_CONNECTOR
	// TODO Retrieve just fields we care about
	uri := "/api/vsdatascriptset?referred_by=virtualservice:" + vs_uuid + "&created_by=" + akcUser
	err = client.AviSession.Get(uri, &rest_response)
	if len(nextPage) == 1 {
		uri = nextPage[0].Next_uri
		ds_key_collection = nextPage[0].Collection.([]NamespaceName)
	} else {
		uri = "/api/vsdatascriptset?referred_by=virtualservice:" + vs_uuid + "&created_by=" + akcUser
	}
	if err != nil {
		utils.AviLog.Warning.Printf("DS Get uri %v returned err %v", uri, err)
	} else {
		resp, ok := rest_response.(map[string]interface{})
		if !ok {
			utils.AviLog.Warning.Printf("DS Get uri %v returned %v type %T", uri,
				rest_response, rest_response)
			return
		}
		utils.AviLog.Info.Printf("DS Get uri %v returned %v DSes", uri,
			resp["count"])
		results, ok := resp["results"].([]interface{})
		if !ok {
			utils.AviLog.Warning.Printf("results not of type []interface{} Instead of type %T", resp["results"])
			return
		}
		for _, ds_intf := range results {
			ds, ok := ds_intf.(map[string]interface{})
			if !ok {
				utils.AviLog.Warning.Printf("ds_intf not of type map[string] interface{}. Instead of type %T", ds_intf)
				continue
			}

			k := NamespaceName{Namespace: tenant, Name: ds["name"].(string)}

			ds_key_collection = append(ds_key_collection, k)
			utils.AviLog.Info.Printf("Added DS cache key %v ", k)
		}
		vs_cache, found := c.VsCache.AviCacheGet(vsKey)
		if found {
			vs_cache_obj, ok := vs_cache.(*AviVsCache)
			if !ok {
				utils.AviLog.Warning.Printf("Unable to cast to VS object: %v", vsKey)
				return
			}
			vs_cache_obj.DSKeyCollection = ds_key_collection
		} else {
			utils.AviLog.Warning.Printf("VS cache not found for key: %v . Unable to update DS collection", vsKey)
		}
		if resp["next"] != nil {
			// It has a next page, let's recursively call the same method.
			next_uri := strings.Split(resp["next"].(string), "/api/vsdatascriptset")
			utils.AviLog.Info.Printf("Found next page, uri for ds during VS cache population: %s", next_uri)
			if len(next_uri) > 1 {
				override_uri := "/api/vsdatascriptset" + next_uri[1]
				nextPage := NextPage{Next_uri: override_uri, Collection: ds_key_collection}
				c.AviDataScriptPopulate(client, cloud, vs_uuid, tenant, vsKey, nextPage)
			}
		}
	}
	return
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
					cloud_obj.NSIpamDNS = c.AviDNSPropertyPopulate(client, dns_uuid)

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
	nsDNSIpam string) string {
	var rest_response interface{}
	uri := "/api/ipamdnsproviderprofile/"
	err := client.AviSession.Get(uri, &rest_response)
	if err != nil {
		utils.AviLog.Warning.Printf("DNSProperty Get uri %v returned err %v", uri, err)
		return ""
	} else {
		resp, ok := rest_response.(map[string]interface{})
		if !ok {
			utils.AviLog.Warning.Printf("DNSProperty Get uri %v returned %v type %T", uri,
				rest_response, rest_response)
			return ""
		}
		utils.AviLog.Info.Printf("DNSProperty Get uri %v returned %v ", uri,
			resp["count"])
		results, ok := resp["results"].([]interface{})
		if !ok {
			utils.AviLog.Warning.Printf("results not of type []interface{} Instead of type %T ", resp["results"])
			return ""
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
					return dns_ipam["domain_name"].(string)
				}

			}

		}
	}
	return ""
}

func ExtractPattern(word string, pattern string) string {
	r, _ := regexp.Compile(pattern)
	result := r.FindAllString(word, -1)
	if len(result) == 1 {
		return result[0][:len(result[0])]
	}
	return ""
}

func ExtractVsUuid(word string) string {
	r, _ := regexp.Compile("virtualservice-.*.#")
	result := r.FindAllString(word, -1)
	if len(result) == 1 {
		return result[0][:len(result[0])-1]
	}
	return ""
}
