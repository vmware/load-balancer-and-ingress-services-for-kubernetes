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
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"
	akov1beta1 "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/apis/ako/v1beta1"

	pq "github.com/jupp0r/go-priority-queue"
	"github.com/vmware/alb-sdk/go/clients"
	"github.com/vmware/alb-sdk/go/models"
	"github.com/vmware/alb-sdk/go/session"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
)

type AviObjCache struct {
	PgCache             *AviCache
	DSCache             *AviCache
	StringGroupCache    *AviCache
	PoolCache           *AviCache
	CloudKeyCache       *AviCache
	HTTPPolicyCache     *AviCache
	L4PolicyCache       *AviCache
	SSLKeyCache         *AviCache
	PKIProfileCache     *AviCache
	VSVIPCache          *AviCache
	VrfCache            *AviCache
	VsCacheMeta         *AviCache
	VsCacheLocal        *AviCache
	AppPersProfileCache *AviCache
	ClusterStatusCache  *AviCache
}

func NewAviObjCache() *AviObjCache {
	c := AviObjCache{}
	c.VsCacheMeta = NewAviCache()
	c.VsCacheLocal = NewAviCache()
	c.PgCache = NewAviCache()
	c.DSCache = NewAviCache()
	c.StringGroupCache = NewAviCache()
	c.PoolCache = NewAviCache()
	c.SSLKeyCache = NewAviCache()
	c.CloudKeyCache = NewAviCache()
	c.HTTPPolicyCache = NewAviCache()
	c.L4PolicyCache = NewAviCache()
	c.VSVIPCache = NewAviCache()
	c.VrfCache = NewAviCache()
	c.PKIProfileCache = NewAviCache()
	c.AppPersProfileCache = NewAviCache()
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
	// We want to run 9 go routines which will simultanesouly fetch objects from the controller.
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
	c.PopulateAppPersistenceProfileToCache(client[9], cloud) // Using client[0] assuming it's available
	c.PopulatePoolsToCache(client[1], cloud)
	c.PopulatePgDataToCache(client[2], cloud)
	c.PopulateStringGroupDataToCache(client[8], cloud)

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
	err = func(client *clients.AviClient) error {
		setDefaultTenant := session.SetTenant(lib.GetTenant())
		setTenant := session.SetTenant("*")
		setTenant(client.AviSession)
		defer setDefaultTenant(client.AviSession)
		return c.AviObjVSCachePopulate(client, cloud, &allVsKeys)

	}(client[0])
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

// all childuuuids being sent
func (c *AviObjCache) PopulateVsMetaCache() {
	// The vh_child_uuids field is used to populate the SNI children during cache population. However, due to the datastore to PG delay - that field may
	// not always be accurate. We would reduce the problem with accuracy by refreshing the SNI cache through reverse mapping sni's to parent
	// Go over the entire VS cache.
	parentVsKeys := c.VsCacheLocal.AviCacheGetAllParentVSKeys()

	isEVHEnabled := lib.IsEvhEnabled()
	var nsNameToDelete []NamespaceName
	childUuidToDelete := make(map[string][]string)

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
					childUuidToDelete[pvsKey.Namespace] = append(childUuidToDelete[pvsKey.Namespace], curChildUuidToDelete...)
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
func (c *AviObjCache) DeleteUnmarked(childCollection map[string][]string) {
	allTenants := make(map[string]struct{})

	dsKeys := make(map[string][]NamespaceName)
	for _, objkey := range c.DSCache.AviGetAllKeys() {
		if _, ok := allTenants[objkey.Namespace]; !ok {
			allTenants[objkey.Namespace] = struct{}{}
		}
		intf, _ := c.DSCache.AviCacheGet(objkey)
		if obj, ok := intf.(*AviDSCache); ok {
			if !obj.HasReference {
				utils.AviLog.Infof("Reference Not found for datascript: %s", objkey)
				dsKeys[objkey.Namespace] = append(dsKeys[objkey.Namespace], objkey)
			}
		}
	}

	httpKeys := make(map[string][]NamespaceName)
	for _, objkey := range c.HTTPPolicyCache.AviGetAllKeys() {
		if _, ok := allTenants[objkey.Namespace]; !ok {
			allTenants[objkey.Namespace] = struct{}{}
		}
		intf, _ := c.HTTPPolicyCache.AviCacheGet(objkey)
		if obj, ok := intf.(*AviHTTPPolicyCache); ok {
			if !obj.HasReference {
				utils.AviLog.Infof("Reference Not found for http policy: %s", objkey)
				httpKeys[objkey.Namespace] = append(httpKeys[objkey.Namespace], objkey)
			}
		}
	}

	l4Keys := make(map[string][]NamespaceName)
	for _, objkey := range c.L4PolicyCache.AviGetAllKeys() {
		if _, ok := allTenants[objkey.Namespace]; !ok {
			allTenants[objkey.Namespace] = struct{}{}
		}
		intf, _ := c.L4PolicyCache.AviCacheGet(objkey)
		if obj, ok := intf.(*AviL4PolicyCache); ok {
			if !obj.HasReference {
				utils.AviLog.Infof("Reference Not found for l4 policy: %s", objkey)
				l4Keys[objkey.Namespace] = append(l4Keys[objkey.Namespace], objkey)
			}
		}
	}

	pgKeys := make(map[string][]NamespaceName)
	for _, objkey := range c.PgCache.AviGetAllKeys() {
		if _, ok := allTenants[objkey.Namespace]; !ok {
			allTenants[objkey.Namespace] = struct{}{}
		}
		intf, _ := c.PgCache.AviCacheGet(objkey)
		if obj, ok := intf.(*AviPGCache); ok {
			if !obj.HasReference {
				utils.AviLog.Infof("Reference Not found for poolgroup: %s", objkey)
				pgKeys[objkey.Namespace] = append(pgKeys[objkey.Namespace], objkey)
			}
		}

	}

	poolKeys := make(map[string][]NamespaceName)
	for _, objkey := range c.PoolCache.AviGetAllKeys() {
		if _, ok := allTenants[objkey.Namespace]; !ok {
			allTenants[objkey.Namespace] = struct{}{}
		}
		intf, _ := c.PoolCache.AviCacheGet(objkey)
		if obj, ok := intf.(*AviPoolCache); ok {
			if !obj.HasReference {
				utils.AviLog.Infof("Reference Not found for pool: %s", objkey)
				poolKeys[objkey.Namespace] = append(poolKeys[objkey.Namespace], objkey)
			}
		}
	}

	sslKeys := make(map[string][]NamespaceName)
	for _, objkey := range c.SSLKeyCache.AviGetAllKeys() {
		if _, ok := allTenants[objkey.Namespace]; !ok {
			allTenants[objkey.Namespace] = struct{}{}
		}
		intf, _ := c.SSLKeyCache.AviCacheGet(objkey)
		if obj, ok := intf.(*AviSSLCache); ok {
			if !obj.HasReference {
				// if deleteConfig is false and istio is enabled, do not delete istio sslkeycert
				if obj.Name == lib.GetIstioWorkloadCertificateName() &&
					lib.IsIstioEnabled() && !lib.GetDeleteConfigMap() {
					continue
				}
				utils.AviLog.Infof("Reference Not found for ssl key: %s", objkey)
				sslKeys[objkey.Namespace] = append(sslKeys[objkey.Namespace], objkey)
			}
		}
	}

	vsVipKeys := make(map[string][]NamespaceName)
	for _, objkey := range c.VSVIPCache.AviGetAllKeys() {
		if _, ok := allTenants[objkey.Namespace]; !ok {
			allTenants[objkey.Namespace] = struct{}{}
		}
		intf, _ := c.VSVIPCache.AviCacheGet(objkey)
		if obj, ok := intf.(*AviVSVIPCache); ok {
			if lib.IsShardVS(obj.Name) {
				utils.AviLog.Infof("Retaining the vsvip: %s", obj.Name)
				continue
			}
			if !obj.HasReference {
				utils.AviLog.Infof("Reference Not found for vsvip: %s", objkey)
				vsVipKeys[objkey.Namespace] = append(vsVipKeys[objkey.Namespace], objkey)
			}
		}
	}

	sgKeys := make(map[string][]NamespaceName)
	for _, objkey := range c.StringGroupCache.AviGetAllKeys() {
		if _, ok := allTenants[objkey.Namespace]; !ok {
			allTenants[objkey.Namespace] = struct{}{}
		}
		intf, _ := c.StringGroupCache.AviCacheGet(objkey)
		if obj, ok := intf.(*AviStringGroupCache); ok {
			if !obj.HasReference {
				utils.AviLog.Infof("Reference Not found for stringgroup: %s", objkey)
				sgKeys[objkey.Namespace] = append(sgKeys[objkey.Namespace], objkey)
			}
		}

	}

	for tenant := range allTenants {
		// Only add this if we have stale data
		vsMetaObj := AviVsCache{
			Name:                     lib.DummyVSForStaleData,
			VSVipKeyCollection:       vsVipKeys[tenant],
			HTTPKeyCollection:        httpKeys[tenant],
			DSKeyCollection:          dsKeys[tenant],
			SSLKeyCertCollection:     sslKeys[tenant],
			PGKeyCollection:          pgKeys[tenant],
			PoolKeyCollection:        poolKeys[tenant],
			L4PolicyCollection:       l4Keys[tenant],
			StringGroupKeyCollection: sgKeys[tenant],
			SNIChildCollection:       childCollection[tenant],
		}
		vsKey := NamespaceName{
			Namespace: tenant,
			Name:      lib.DummyVSForStaleData,
		}
		utils.AviLog.Infof("Dummy VS in tenant %s for stale objects Deletion %s", tenant, utils.Stringify(&vsMetaObj))
		c.VsCacheMeta.AviCacheAdd(vsKey, &vsMetaObj)
	}

}

func (c *AviObjCache) AviPopulateAllPGs(client *clients.AviClient, cloud string, pgData *[]AviPGCache, overrideUri ...NextPage) (*[]AviPGCache, int, error) {
	var uri string
	akoUser := lib.AKOUser

	if len(overrideUri) == 1 {
		uri = overrideUri[0].NextURI
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
			poolUuid := ExtractUUID(*member.PoolRef, "pool-.*.#")
			// Search the poolName using this Uuid in the poolcache.
			poolName, found := c.PoolCache.AviCacheGetNameByUuid(poolUuid)
			if found {
				pools = append(pools, poolName.(string))
			}
		}
		pgCacheObj := AviPGCache{
			Name:             *pg.Name,
			Tenant:           getTenantFromTenantRef(*pg.TenantRef),
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
			nextPage := NextPage{NextURI: overrideUri}
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
	setDefaultTenant := session.SetTenant(lib.GetTenant())
	setTenant := session.SetTenant("*")
	setTenant(client.AviSession)
	defer setDefaultTenant(client.AviSession)
	c.AviPopulateAllPGs(client, cloud, &pgData)

	// Get all the PG cache data and copy them.
	pgCacheData := c.PgCache.ShallowCopy()
	for i, pgCacheObj := range pgData {
		k := NamespaceName{Namespace: pgCacheObj.Tenant, Name: pgCacheObj.Name}
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
		_, ok := key.(NamespaceName)
		if !ok {
			continue
		}
		utils.AviLog.Debugf("Deleting key from pg cache :%s", key)
		c.PgCache.AviCacheDelete(key)
	}
}

func (c *AviObjCache) AviPopulateAllPkiPRofiles(client *clients.AviClient, pkiData *[]AviPkiProfileCache, overrideUri ...NextPage) (*[]AviPkiProfileCache, int, error) {
	var uri string
	akoUser := lib.AKOUser

	if len(overrideUri) == 1 {
		uri = overrideUri[0].NextURI
	} else {
		uri = "/api/pkiprofile/?" + "&include_name=true&created_by=" + akoUser + "&page_size=100"
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
			Tenant:           getTenantFromTenantRef(*pki.TenantRef),
			CloudConfigCksum: lib.SSLKeyCertChecksum(*pki.Name, string(*pki.CaCerts[0].Certificate), "", emptyIngestionMarkers, pki.Markers, true),
		}
		*pkiData = append(*pkiData, pkiCacheObj)

	}
	if result.Next != "" {
		// It has a next page, let's recursively call the same method.
		next_uri := strings.Split(result.Next, "/api/pkiprofile")
		if len(next_uri) > 1 {
			overrideUri := "/api/pkiprofile" + next_uri[1]
			nextPage := NextPage{NextURI: overrideUri}
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
		uri = overrideUri[0].NextURI
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

		tenant := getTenantFromTenantRef(*pool.TenantRef)
		var pkiKey NamespaceName
		if pool.PkiProfileRef != nil {
			pkiUuid := ExtractUUID(*pool.PkiProfileRef, "pkiprofile-.*.#")
			pkiName, foundPki := c.PKIProfileCache.AviCacheGetNameByUuid(pkiUuid)
			if foundPki {
				pkiKey = NamespaceName{Namespace: tenant, Name: pkiName.(string)}
			}
		}
		var persistentProfileKey NamespaceName
		if pool.ApplicationPersistenceProfileRef != nil {
			persistentProfileUuid := ExtractUUID(*pool.ApplicationPersistenceProfileRef, "applicationpersistenceprofile-.*.#")
			persistentProfileName, found := c.AppPersProfileCache.AviCacheGetNameByUuid(persistentProfileUuid)
			if found {
				persistentProfileKey = NamespaceName{Namespace: tenant, Name: persistentProfileName.(string)}
			}
		}
		poolCacheObj := AviPoolCache{
			Name:                 *pool.Name,
			Tenant:               tenant,
			Uuid:                 *pool.UUID,
			CloudConfigCksum:     *pool.CloudConfigCksum,
			PkiProfileCollection: pkiKey,
			PersistenceProfile:   persistentProfileKey,
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
			nextPage := NextPage{NextURI: overrideUri}
			_, _, err := c.AviPopulateAllPools(client, cloud, poolData, nextPage)
			if err != nil {
				return nil, 0, err
			}
		}
	}

	return poolData, result.Count, nil
}

func (c *AviObjCache) PopulatePkiProfilesToCache(client *clients.AviClient) {
	var pkiProfData []AviPkiProfileCache
	setDefaultTenant := session.SetTenant(lib.GetTenant())
	setTenant := session.SetTenant("*")
	setTenant(client.AviSession)
	defer setDefaultTenant(client.AviSession)
	c.AviPopulateAllPkiPRofiles(client, &pkiProfData)

	pkiCacheData := c.PKIProfileCache.ShallowCopy()
	for i, pkiCacheObj := range pkiProfData {
		k := NamespaceName{Namespace: pkiCacheObj.Tenant, Name: pkiCacheObj.Name}
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
		_, ok := key.(NamespaceName)
		if !ok {
			continue
		}
		utils.AviLog.Infof("Deleting key from pki cache :%s", key)
		c.PKIProfileCache.AviCacheDelete(key)
	}
}

func (c *AviObjCache) PopulatePoolsToCache(client *clients.AviClient, cloud string) {
	var poolsData []AviPoolCache
	setDefaultTenant := session.SetTenant(lib.GetTenant())
	setTenant := session.SetTenant("*")
	setTenant(client.AviSession)
	defer setDefaultTenant(client.AviSession)
	c.AviPopulateAllPools(client, cloud, &poolsData)

	poolCacheData := c.PoolCache.ShallowCopy()
	for i, poolCacheObj := range poolsData {
		k := NamespaceName{Namespace: poolCacheObj.Tenant, Name: poolCacheObj.Name}
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
		_, ok := key.(NamespaceName)
		if !ok {
			continue
		}
		utils.AviLog.Debugf("Deleting key from pool cache :%s", key)
		c.PoolCache.AviCacheDelete(key)
	}
}

func (c *AviObjCache) AviPopulateAllVSVips(client *clients.AviClient, cloud string, vsVipData *[]AviVSVIPCache, nextPage ...NextPage) (*[]AviVSVIPCache, error) {
	var uri string

	if len(nextPage) == 1 {
		uri = nextPage[0].NextURI
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
		var v6ips []string
		var networkNames []string
		for _, vip := range vsvip.Vip {
			if vip.IPAddress != nil {
				vips = append(vips, *vip.IPAddress.Addr)
			}
			if vip.FloatingIP != nil {
				fips = append(fips, *vip.FloatingIP.Addr)
			}
			if vip.Ip6Address != nil {
				v6ips = append(v6ips, *vip.Ip6Address.Addr)
			}
			if ipamNetworkSubnet := vip.IPAMNetworkSubnet; ipamNetworkSubnet != nil && ipamNetworkSubnet.NetworkRef != nil {
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
			Tenant:           getTenantFromTenantRef(*vsvip.TenantRef),
			Uuid:             *vsvip.UUID,
			FQDNs:            fqdns,
			NetworkNames:     networkNames,
			LastModified:     *vsvip.LastModified,
			Vips:             vips,
			Fips:             fips,
			V6IPs:            v6ips,
			CloudConfigCksum: checksum,
		}
		*vsVipData = append(*vsVipData, vsVipCacheObj)
	}
	if result.Next != "" {
		// It has a next page, let's recursively call the same method.
		next_uri := strings.Split(result.Next, "/api/vsvip")
		if len(next_uri) > 1 {
			overrideUri := "/api/vsvip" + next_uri[1]
			nextPage := NextPage{NextURI: overrideUri}
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
	setDefaultTenant := session.SetTenant(lib.GetTenant())
	setTenant := session.SetTenant("*")
	setTenant(client.AviSession)
	defer setDefaultTenant(client.AviSession)
	c.AviPopulateAllVSVips(client, cloud, &vsVipData)

	vsVipCacheData := c.VSVIPCache.ShallowCopy()
	for i, vsVipCacheObj := range vsVipData {
		k := NamespaceName{Namespace: vsVipCacheObj.Tenant, Name: vsVipCacheObj.Name}
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
		_, ok := key.(NamespaceName)
		if !ok {
			continue
		}
		utils.AviLog.Debugf("Deleting key from vsvip cache :%s", key)
		c.VSVIPCache.AviCacheDelete(key)
	}
}

func (c *AviObjCache) AviPopulateAllDSs(client *clients.AviClient, cloud string, DsData *[]AviDSCache, nextPage ...NextPage) (*[]AviDSCache, int, error) {
	var uri string
	akoUser := lib.AKOUser

	if len(nextPage) == 1 {
		uri = nextPage[0].NextURI
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
			pgUuid := ExtractUUID(pg, "poolgroup-.*.#")
			// Search the poolName using this Uuid in the poolcache.
			pgName, found := c.PgCache.AviCacheGetNameByUuid(pgUuid)
			if found {
				pgs = append(pgs, pgName.(string))
			}
		}
		dsCacheObj := AviDSCache{
			Name:       *ds.Name,
			Tenant:     getTenantFromTenantRef(*ds.TenantRef),
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
			nextPage := NextPage{NextURI: overrideUri}
			_, _, err := c.AviPopulateAllDSs(client, cloud, DsData, nextPage)
			if err != nil {
				return nil, 0, err
			}
		}
	}
	return DsData, result.Count, nil
}

func (c *AviObjCache) PopulateDSDataToCache(client *clients.AviClient, cloud string) {
	var DsData []AviDSCache
	setDefaultTenant := session.SetTenant(lib.GetTenant())
	setTenant := session.SetTenant("*")
	setTenant(client.AviSession)
	defer setDefaultTenant(client.AviSession)
	c.AviPopulateAllDSs(client, cloud, &DsData)
	dsCacheData := c.DSCache.ShallowCopy()
	for i, DsCacheObj := range DsData {
		k := NamespaceName{Namespace: DsCacheObj.Tenant, Name: DsCacheObj.Name}
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
		_, ok := key.(NamespaceName)
		if !ok {
			continue
		}
		utils.AviLog.Debugf("Deleting key from ds cache :%s", key)
		c.DSCache.AviCacheDelete(key)
	}
}

func (c *AviObjCache) AviPopulateAllSSLKeys(client *clients.AviClient, cloud string, SslData *[]AviSSLCache, nextPage ...NextPage) (*[]AviSSLCache, int, error) {
	var uri string
	akoUser := lib.AKOUser

	if len(nextPage) == 1 {
		uri = nextPage[0].NextURI
	} else {
		uri = "/api/sslkeyandcertificate/?" + "&include_name=true" + "&created_by=" + akoUser + "&page_size=100"
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
				cacertUUID = ExtractUUIDWithoutHash(*sslkey.CaCerts[0].CaRef, "sslkeyandcertificate-.*.")
				cacertIntf, found := c.SSLKeyCache.AviCacheGetNameByUuid(cacertUUID)
				if found {
					cacert = cacertIntf.(string)
				}
			}
		}
		emptyIngestionMarkers := utils.AviObjectMarkers{}
		sslCacheObj := AviSSLCache{
			Name:             *sslkey.Name,
			Tenant:           getTenantFromTenantRef(*sslkey.TenantRef),
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
			nextPage := NextPage{NextURI: overrideUri}
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

	uri = "/api/sslkeyandcertificate?name=" + objName + "&include_name=true" + "&created_by=" + akoUser

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
				cacertUUID := ExtractUUIDWithoutHash(*sslkey.CaCerts[0].CaRef, "sslkeyandcertificate-.*.")
				cacertIntf, found := c.SSLKeyCache.AviCacheGetNameByUuid(cacertUUID)
				if found {
					cacert = cacertIntf.(string)
				}
			}
		}
		emptyIngestionMarkers := utils.AviObjectMarkers{}
		tenant := getTenantFromTenantRef(*sslkey.TenantRef)
		sslCacheObj := AviSSLCache{
			Name:             *sslkey.Name,
			Tenant:           tenant,
			Uuid:             *sslkey.UUID,
			CloudConfigCksum: lib.SSLKeyCertChecksum(*sslkey.Name, *sslkey.Certificate.Certificate, cacert, emptyIngestionMarkers, sslkey.Markers, true),
			HasCARef:         hasCA,
		}
		k := NamespaceName{Namespace: tenant, Name: *sslkey.Name}
		c.SSLKeyCache.AviCacheAdd(k, &sslCacheObj)
		utils.AviLog.Debugf("Adding sslkey to Cache during refresh %s", k)
	}
	return nil
}

func (c *AviObjCache) AviPopulateOnePKICache(client *clients.AviClient,
	cloud string, objName string) error {
	var uri string
	akoUser := lib.AKOUser

	uri = "/api/pkiprofile?name=" + objName + "&include_name=true" + "&created_by=" + akoUser

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
		tenant := getTenantFromTenantRef(*pkikey.TenantRef)
		sslCacheObj := AviSSLCache{
			Name:             *pkikey.Name,
			Tenant:           tenant,
			Uuid:             *pkikey.UUID,
			CloudConfigCksum: lib.SSLKeyCertChecksum(*pkikey.Name, *pkikey.CaCerts[0].Certificate, "", emptyIngestionMarkers, pkikey.Markers, true),
		}
		k := NamespaceName{Namespace: tenant, Name: *pkikey.Name}
		c.SSLKeyCache.AviCacheAdd(k, &sslCacheObj)
		utils.AviLog.Debugf("Adding pkikey to Cache during refresh %s", k)
	}
	return nil
}

func (c *AviObjCache) AviPopulateOnePoolCache(client *clients.AviClient,
	cloud string, objName string) error {
	var uri string
	akoUser := lib.AKOUser

	uri = "/api/pool?name=" + objName + "&include_name=true" + "&created_by=" + akoUser

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

		tenant := getTenantFromTenantRef(*pool.TenantRef)
		var pkiKey NamespaceName
		if pool.PkiProfileRef != nil {
			pkiUuid := ExtractUUID(*pool.PkiProfileRef, "pkiprofile-.*.#")
			pkiName, foundPki := c.PKIProfileCache.AviCacheGetNameByUuid(pkiUuid)
			if foundPki {
				pkiKey = NamespaceName{Namespace: tenant, Name: pkiName.(string)}
			}
		}

		var persistenceKey NamespaceName
		if pool.ApplicationPersistenceProfileRef != nil {
			persistenceProfileUuid := ExtractUUID(*pool.ApplicationPersistenceProfileRef, "applicationpersistenceprofile-.*.#")
			persistenceProfileName, foundPersistenceProfile := c.AppPersProfileCache.AviCacheGetNameByUuid(persistenceProfileUuid)
			if foundPersistenceProfile {
				pkiKey = NamespaceName{Namespace: tenant, Name: persistenceProfileName.(string)}
			}
		}

		poolCacheObj := AviPoolCache{
			Name:                 *pool.Name,
			Tenant:               tenant,
			Uuid:                 *pool.UUID,
			CloudConfigCksum:     *pool.CloudConfigCksum,
			PkiProfileCollection: pkiKey,
			PersistenceProfile:   persistenceKey,
			ServiceMetadataObj:   svc_mdata_obj,
			LastModified:         *pool.LastModified,
		}
		k := NamespaceName{Namespace: tenant, Name: *pool.Name}
		c.PoolCache.AviCacheAdd(k, &poolCacheObj)
		utils.AviLog.Debugf("Adding pool to Cache during refresh %s", k)
	}
	return nil
}

func (c *AviObjCache) AviPopulateOneVsDSCache(client *clients.AviClient,
	cloud string, objName string) error {
	var uri string
	akoUser := lib.AKOUser

	uri = "/api/vsdatascriptset?name=" + objName + "&include_name=true" + "&created_by=" + akoUser

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
			pgUuid := ExtractUUID(pg, "poolgroup-.*.#")
			// Search the poolName using this Uuid in the poolcache.
			pgName, found := c.PgCache.AviCacheGetNameByUuid(pgUuid)
			if found {
				pgs = append(pgs, pgName.(string))
			}
		}
		tenant := getTenantFromTenantRef(*ds.TenantRef)
		dsCacheObj := AviDSCache{
			Name:       *ds.Name,
			Tenant:     tenant,
			Uuid:       *ds.UUID,
			PoolGroups: pgs,
		}
		checksum := lib.DSChecksum(dsCacheObj.PoolGroups, ds.Markers, true)
		if len(ds.Datascript) == 1 {
			checksum += utils.Hash(*ds.Datascript[0].Script)
		}
		dsCacheObj.CloudConfigCksum = checksum
		k := NamespaceName{Namespace: tenant, Name: *ds.Name}
		c.DSCache.AviCacheAdd(k, &dsCacheObj)
		utils.AviLog.Debugf("Adding ds to Cache during refresh %s", k)
	}
	return nil
}

func (c *AviObjCache) AviPopulateOnePGCache(client *clients.AviClient,
	cloud string, objName string) error {
	var uri string
	akoUser := lib.AKOUser

	uri = "/api/poolgroup?name=" + objName + "&include_name=true" + "&created_by=" + akoUser

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
			poolUuid := ExtractUUID(*member.PoolRef, "pool-.*.#")
			// Search the poolName using this Uuid in the poolcache.
			poolName, found := c.PoolCache.AviCacheGetNameByUuid(poolUuid)
			if found {
				pools = append(pools, poolName.(string))
			}
		}
		tenant := getTenantFromTenantRef(*pg.TenantRef)
		pgCacheObj := AviPGCache{
			Name:             *pg.Name,
			Tenant:           tenant,
			Uuid:             *pg.UUID,
			CloudConfigCksum: *pg.CloudConfigCksum,
			LastModified:     *pg.LastModified,
			Members:          pools,
		}
		k := NamespaceName{Namespace: tenant, Name: *pg.Name}
		c.PgCache.AviCacheAdd(k, &pgCacheObj)
		utils.AviLog.Debugf("Adding pg to Cache during refresh %s", k)
	}
	return nil
}

func (c *AviObjCache) AviPopulateOneVsVipCache(client *clients.AviClient,
	cloud string, objName string) error {
	var uri string

	uri = "/api/vsvip?name=" + objName + "&cloud_ref.name=" + cloud + "&include_name=true"

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
		var v6ips []string
		var networkNames []string
		for _, vip := range vsvip.Vip {
			vips = append(vips, *vip.IPAddress.Addr)
			if vip.FloatingIP != nil {
				fips = append(fips, *vip.FloatingIP.Addr)
			}
			if vip.Ip6Address != nil {
				v6ips = append(v6ips, *vip.Ip6Address.Addr)
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
		tenant := getTenantFromTenantRef(*vsvip.TenantRef)
		vsVipCacheObj := AviVSVIPCache{
			Name:             *vsvip.Name,
			Tenant:           tenant,
			Uuid:             *vsvip.UUID,
			FQDNs:            fqdns,
			LastModified:     *vsvip.LastModified,
			Vips:             vips,
			Fips:             fips,
			V6IPs:            v6ips,
			NetworkNames:     networkNames,
			CloudConfigCksum: checksum,
		}
		k := NamespaceName{Namespace: tenant, Name: *vsvip.Name}
		c.VSVIPCache.AviCacheAdd(k, &vsVipCacheObj)
		utils.AviLog.Debugf("Adding vsvip to Cache during refresh %s", k)
	}
	return nil
}

func (c *AviObjCache) AviPopulateOneVsHttpPolCache(client *clients.AviClient,
	cloud string, objName string) error {
	var uri string
	akoUser := lib.AKOUser

	uri = "/api/httppolicyset?name=" + objName + "&include_name=true" + "&created_by=" + akoUser

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
		var stringGroupRefs []string
		if httppol.HTTPRequestPolicy != nil {
			for _, rule := range httppol.HTTPRequestPolicy.Rules {
				if rule.SwitchingAction != nil {
					val := reflect.ValueOf(rule.SwitchingAction)
					if !val.Elem().FieldByName("PoolGroupRef").IsNil() {
						pgUuid := ExtractUUID(*rule.SwitchingAction.PoolGroupRef, "poolgroup-.*.#")
						pgName, found := c.PgCache.AviCacheGetNameByUuid(pgUuid)
						if found {
							poolGroups = append(poolGroups, pgName.(string))
						}
					} else if !val.Elem().FieldByName("PoolRef").IsNil() {
						poolUuid := ExtractUUID(*rule.SwitchingAction.PoolRef, "pool-.*.#")
						poolName, found := c.PoolCache.AviCacheGetNameByUuid(poolUuid)
						if found {
							pools = append(pools, poolName.(string))
						}
					}
				}
				if rule.Match != nil && rule.Match.Path != nil {
					for _, sg := range rule.Match.Path.StringGroupRefs {
						sgUuid := ExtractUUID(sg, "stringgroup-.*.#")
						// Search the string group name using this Uuid in the string group cache.
						sgName, found := c.StringGroupCache.AviCacheGetNameByUuid(sgUuid)
						if found {
							stringGroupRefs = append(stringGroupRefs, sgName.(string))
						}
					}
				}
			}
		}

		tenant := getTenantFromTenantRef(*httppol.TenantRef)
		httpPolCacheObj := AviHTTPPolicyCache{
			Name:             *httppol.Name,
			Tenant:           tenant,
			Uuid:             *httppol.UUID,
			CloudConfigCksum: *httppol.CloudConfigCksum,
			PoolGroups:       poolGroups,
			Pools:            pools,
			LastModified:     *httppol.LastModified,
			StringGroupRefs:  stringGroupRefs,
		}
		k := NamespaceName{Namespace: tenant, Name: *httppol.Name}
		c.HTTPPolicyCache.AviCacheAdd(k, &httpPolCacheObj)
		utils.AviLog.Debugf("Adding httppolicy to Cache during refresh %s", k)
	}
	return nil
}

func (c *AviObjCache) AviPopulateOneVsL4PolCache(client *clients.AviClient,
	cloud string, objName string) error {
	var uri string
	akoUser := lib.AKOUser

	uri = "/api/l4policyset?name=" + objName + "&include_name=true" + "&created_by=" + akoUser

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
					poolUuid := ExtractUUID(*rule.Action.SelectPool.PoolRef, "pool-.*.#")
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
		cksum := lib.L4PolicyChecksum(ports, protocols, pools, emptyIngestionMarkers, l4pol.Markers, true)
		tenant := getTenantFromTenantRef(*l4pol.TenantRef)
		l4PolCacheObj := AviL4PolicyCache{
			Name:             *l4pol.Name,
			Tenant:           tenant,
			Uuid:             *l4pol.UUID,
			Pools:            pools,
			LastModified:     *l4pol.LastModified,
			CloudConfigCksum: cksum,
		}
		k := NamespaceName{Namespace: tenant, Name: *l4pol.Name}
		c.L4PolicyCache.AviCacheAdd(k, &l4PolCacheObj)
		utils.AviLog.Infof("Adding l4pol to Cache during refresh %s", utils.Stringify(l4PolCacheObj))
	}
	return nil
}

func (c *AviObjCache) PopulateSSLKeyToCache(client *clients.AviClient, cloud string) {
	var SslKeyData []AviSSLCache
	setDefaultTenant := session.SetTenant(lib.GetTenant())
	setTenant := session.SetTenant("*")
	setTenant(client.AviSession)
	defer setDefaultTenant(client.AviSession)
	c.AviPopulateAllSSLKeys(client, cloud, &SslKeyData)
	sslCacheData := c.SSLKeyCache.ShallowCopy()
	for i, SslKeyCacheObj := range SslKeyData {
		k := NamespaceName{Namespace: SslKeyCacheObj.Tenant, Name: SslKeyCacheObj.Name}
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
		_, ok := key.(NamespaceName)
		if !ok {
			continue
		}
		utils.AviLog.Debugf("Deleting key from sslkey cache :%s", key)
		c.SSLKeyCache.AviCacheDelete(key)
	}
}

func (c *AviObjCache) AviPopulateAllHttpPolicySets(client *clients.AviClient, cloud string, httpPolicyData *[]AviHTTPPolicyCache, nextPage ...NextPage) (*[]AviHTTPPolicyCache, int, error) {
	var uri string
	akoUser := lib.AKOUser

	if len(nextPage) == 1 {
		uri = nextPage[0].NextURI
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

		// Fetch the pgs and string group refs associated with the http policyset object
		var poolGroups []string
		var pools []string
		var stringGroupRefs []string
		if httppol.HTTPRequestPolicy != nil {
			for _, rule := range httppol.HTTPRequestPolicy.Rules {
				if rule.SwitchingAction != nil {
					val := reflect.ValueOf(rule.SwitchingAction)
					if !val.Elem().FieldByName("PoolGroupRef").IsNil() {
						pgUuid := ExtractUUID(*rule.SwitchingAction.PoolGroupRef, "poolgroup-.*.#")
						pgName, found := c.PgCache.AviCacheGetNameByUuid(pgUuid)
						if found {
							poolGroups = append(poolGroups, pgName.(string))
						}
					} else if !val.Elem().FieldByName("PoolRef").IsNil() {
						poolUuid := ExtractUUID(*rule.SwitchingAction.PoolRef, "pool-.*.#")
						poolName, found := c.PoolCache.AviCacheGetNameByUuid(poolUuid)
						if found {
							pools = append(pools, poolName.(string))
						}
					}
				}
				if rule.Match != nil && rule.Match.Path != nil {
					for _, sg := range rule.Match.Path.StringGroupRefs {
						sgUuid := ExtractUUID(sg, "stringgroup-.*.#")
						// Search the string group name using this Uuid in the string group cache.
						sgName, found := c.StringGroupCache.AviCacheGetNameByUuid(sgUuid)
						if found {
							stringGroupRefs = append(stringGroupRefs, sgName.(string))
						}
					}
				}
			}
		}
		tenant := getTenantFromTenantRef(*httppol.TenantRef)
		httpPolCacheObj := AviHTTPPolicyCache{
			Name:             *httppol.Name,
			Tenant:           tenant,
			Uuid:             *httppol.UUID,
			CloudConfigCksum: *httppol.CloudConfigCksum,
			PoolGroups:       poolGroups,
			Pools:            pools,
			LastModified:     *httppol.LastModified,
			StringGroupRefs:  stringGroupRefs,
		}
		*httpPolicyData = append(*httpPolicyData, httpPolCacheObj)
	}
	if result.Next != "" {
		// It has a next page, let's recursively call the same method.
		next_uri := strings.Split(result.Next, "/api/httppolicyset")
		if len(next_uri) > 1 {
			overrideUri := "/api/httppolicyset" + next_uri[1]
			nextPage := NextPage{NextURI: overrideUri}
			_, _, err := c.AviPopulateAllHttpPolicySets(client, cloud, httpPolicyData, nextPage)
			if err != nil {
				return nil, 0, err
			}
		}
	}
	return httpPolicyData, result.Count, nil
}
func (c *AviObjCache) AviPopulateHttpPolicySetbyUUID(client *clients.AviClient, uuid string) error {

	uri := "/api/httppolicyset/" + uuid
	rawData, err := lib.AviGetRaw(client, uri)
	if err != nil {
		utils.AviLog.Warnf("Get uri %v returned err for httppolicyset %v", uri, err)
		return err
	}
	httppol := models.HTTPPolicySet{}
	err = json.Unmarshal(rawData, &httppol)
	if err != nil {
		utils.AviLog.Warnf("Failed to unmarshal httppolicyset data, err: %v", err)
		return err
	}

	if httppol.Name == nil || httppol.UUID == nil {
		utils.AviLog.Warnf("Incomplete http policy data unmarshalled, %s", utils.Stringify(httppol))
		return errors.New("incomplete http policy data unmarshalled")
	}

	// Fetch the pgs associated with the http policyset object
	var poolGroups []string
	var pools []string
	if httppol.HTTPRequestPolicy != nil {
		for _, rule := range httppol.HTTPRequestPolicy.Rules {
			if rule.SwitchingAction != nil {
				val := reflect.ValueOf(rule.SwitchingAction)
				if !val.Elem().FieldByName("PoolGroupRef").IsNil() {
					pgUuid := ExtractUUID(*rule.SwitchingAction.PoolGroupRef, "poolgroup-.*.#")
					pgName, found := c.PgCache.AviCacheGetNameByUuid(pgUuid)
					if found {
						poolGroups = append(poolGroups, pgName.(string))
					}
				} else if !val.Elem().FieldByName("PoolRef").IsNil() {
					poolUuid := ExtractUUID(*rule.SwitchingAction.PoolRef, "pool-.*.#")
					poolName, found := c.PoolCache.AviCacheGetNameByUuid(poolUuid)
					if found {
						pools = append(pools, poolName.(string))
					}
				}
			}

		}
	}
	tenant := getTenantFromTenantRef(*httppol.TenantRef)
	httpPolCacheObj := AviHTTPPolicyCache{
		Name:         *httppol.Name,
		Uuid:         *httppol.UUID,
		PoolGroups:   poolGroups,
		Pools:        pools,
		LastModified: *httppol.LastModified,
	}
	key := NamespaceName{Namespace: tenant, Name: httpPolCacheObj.Name}
	c.HTTPPolicyCache.AviCacheAdd(key, httpPolCacheObj)
	utils.AviLog.Debugf("added policy with key %s and policyset %v", key, httpPolCacheObj)
	return nil
}

func (c *AviObjCache) PopulateHttpPolicySetToCache(client *clients.AviClient, cloud string) {
	var HttPolData []AviHTTPPolicyCache
	setDefaultTenant := session.SetTenant(lib.GetTenant())
	setTenant := session.SetTenant("*")
	setTenant(client.AviSession)
	defer setDefaultTenant(client.AviSession)
	_, count, err := c.AviPopulateAllHttpPolicySets(client, cloud, &HttPolData)
	if err != nil || len(HttPolData) != count {
		return
	}
	httpCacheData := c.HTTPPolicyCache.ShallowCopy()
	for i, HttpPolCacheObj := range HttPolData {
		k := NamespaceName{Namespace: HttpPolCacheObj.Tenant, Name: HttpPolCacheObj.Name}
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
		_, ok := key.(NamespaceName)
		if !ok {
			continue
		}
		utils.AviLog.Debugf("Deleting key from httppol cache :%s", key)
		c.HTTPPolicyCache.AviCacheDelete(key)
	}
}

func (c *AviObjCache) AviPopulateAllL4PolicySets(client *clients.AviClient, cloud string, l4PolicyData *[]AviL4PolicyCache, nextPage ...NextPage) (*[]AviL4PolicyCache, int, error) {
	var uri string
	akoUser := lib.AKOUser

	if len(nextPage) == 1 {
		uri = nextPage[0].NextURI
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
					poolUuid := ExtractUUID(*rule.Action.SelectPool.PoolRef, "pool-.*.#")
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
		cksum := lib.L4PolicyChecksum(ports, protocols, pools, emptyIngestionMarkers, l4pol.Markers, true)
		l4PolCacheObj := AviL4PolicyCache{
			Name:             *l4pol.Name,
			Tenant:           getTenantFromTenantRef(*l4pol.TenantRef),
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
			nextPage := NextPage{NextURI: overrideUri}
			_, _, err := c.AviPopulateAllL4PolicySets(client, cloud, l4PolicyData, nextPage)
			if err != nil {
				return nil, 0, err
			}
		}
	}
	return l4PolicyData, result.Count, nil
}

func (c *AviObjCache) PopulateL4PolicySetToCache(client *clients.AviClient, cloud string) {
	var l4PolData []AviL4PolicyCache
	setDefaultTenant := session.SetTenant(lib.GetTenant())
	setTenant := session.SetTenant("*")
	setTenant(client.AviSession)
	defer setDefaultTenant(client.AviSession)
	_, count, err := c.AviPopulateAllL4PolicySets(client, cloud, &l4PolData)
	if err != nil || len(l4PolData) != count {
		return
	}
	l4CacheData := c.L4PolicyCache.ShallowCopy()
	for i, l4PolCacheObj := range l4PolData {
		k := NamespaceName{Namespace: l4PolCacheObj.Tenant, Name: l4PolCacheObj.Name}
		utils.AviLog.Debugf("Adding key to l4 cache :%s", utils.Stringify(l4PolCacheObj))
		c.L4PolicyCache.AviCacheAdd(k, &l4PolData[i])
		delete(l4CacheData, k)
	}
	// // The data that is left in httpCacheData should be explicitly removed
	for key := range l4CacheData {
		_, ok := key.(NamespaceName)
		if !ok {
			continue
		}
		utils.AviLog.Debugf("Deleting key from l4policy cache :%s", key)
		c.L4PolicyCache.AviCacheDelete(key)
	}
}

func (c *AviObjCache) AviPopulateAllStringGroups(client *clients.AviClient, cloud string, StringGroupData *[]AviStringGroupCache, nextPage ...NextPage) (*[]AviStringGroupCache, int, error) {
	var uri string

	if len(nextPage) == 1 {
		uri = nextPage[0].NextURI
	} else {
		//Fetching container specific StringGroups
		uri = "/api/stringgroup?&include_name=true&label_key=created_by&label_value=" + lib.GetAKOUser() + "&page_size=100"
	}

	result, err := lib.AviGetCollectionRaw(client, uri)
	if err != nil {
		utils.AviLog.Warnf("Get uri %v returned err for stringgroup %v", uri, err)
		return nil, 0, err
	}
	elems := make([]json.RawMessage, result.Count)
	err = json.Unmarshal(result.Results, &elems)
	if err != nil {
		utils.AviLog.Warnf("Failed to unmarshal stringgroup data, err: %v", err)
		return nil, 0, err
	}
	for i := 0; i < len(elems); i++ {
		sg := models.StringGroup{}
		err = json.Unmarshal(elems[i], &sg)
		if err != nil {
			utils.AviLog.Warnf("Failed to unmarshal stringgroup data, err: %v", err)
			continue
		}
		if sg.Name == nil || sg.UUID == nil {
			utils.AviLog.Warnf("Incomplete stringgroup data unmarshalled, %s", utils.Stringify(sg))
			continue
		}

		stringGroupCacheObj := AviStringGroupCache{
			Name:   *sg.Name,
			Uuid:   *sg.UUID,
			Tenant: getTenantFromTenantRef(*sg.TenantRef),
		}
		if sg.Description != nil {
			stringGroupCacheObj.Description = *sg.Description
		}
		if sg.LongestMatch != nil {
			stringGroupCacheObj.LongestMatch = *sg.LongestMatch
		}
		checksum := lib.StringGroupChecksum(sg.Kv, sg.Markers, sg.LongestMatch, true)

		stringGroupCacheObj.CloudConfigCksum = checksum
		*StringGroupData = append(*StringGroupData, stringGroupCacheObj)
	}
	if result.Next != "" {
		// It has a next page, let's recursively call the same method.
		next_uri := strings.Split(result.Next, "/api/stringgroup")
		if len(next_uri) > 1 {
			overrideUri := "/api/stringgroup" + next_uri[1]
			nextPage := NextPage{NextURI: overrideUri}
			_, _, err := c.AviPopulateAllStringGroups(client, cloud, StringGroupData, nextPage)
			if err != nil {
				return nil, 0, err
			}
		}
	}
	return StringGroupData, result.Count, nil
}

func (c *AviObjCache) PopulateStringGroupDataToCache(client *clients.AviClient, cloud string) {
	var StringGroupData []AviStringGroupCache
	setDefaultTenant := session.SetTenant(lib.GetTenant())
	setTenant := session.SetTenant("*")
	setTenant(client.AviSession)
	defer setDefaultTenant(client.AviSession)
	c.AviPopulateAllStringGroups(client, cloud, &StringGroupData)
	stringGroupCacheData := c.StringGroupCache.ShallowCopy()
	for i, stringGroupCacheObj := range StringGroupData {
		k := NamespaceName{Namespace: stringGroupCacheObj.Tenant, Name: stringGroupCacheObj.Name}
		oldSGIntf, found := c.StringGroupCache.AviCacheGet(k)
		if found {
			oldSGData, ok := oldSGIntf.(*AviStringGroupCache)
			if ok {
				if oldSGData.InvalidData || oldSGData.LastModified != StringGroupData[i].LastModified {
					StringGroupData[i].InvalidData = true
					utils.AviLog.Warnf("Invalid cache data for stringgroup: %s", k)
				}
			} else {
				utils.AviLog.Warnf("Wrong data type for stringgroup: %s in cache", k)
			}
		}
		utils.AviLog.Debugf("Adding key to stringgroup cache :%s", k)
		c.StringGroupCache.AviCacheAdd(k, &StringGroupData[i])
		delete(stringGroupCacheData, k)
	}
	// The data that is left in stringGroupCacheData should be explicitly removed
	for key := range stringGroupCacheData {
		_, ok := key.(NamespaceName)
		if !ok {
			continue
		}
		utils.AviLog.Debugf("Deleting key from stringgroup cache :%s", key)
		c.StringGroupCache.AviCacheDelete(key)
	}
}

func (c *AviObjCache) AviPopulateOneStringGroupCache(client *clients.AviClient,
	cloud string, objName string) error {
	uri := "/api/stringgroup?name=" + objName + "&include_name=true&label_key=created_by&label_value=" + lib.GetAKOUser()

	result, err := lib.AviGetCollectionRaw(client, uri)
	if err != nil {
		utils.AviLog.Warnf("Get uri %v returned err for stringgroup %v", uri, err)
		return err
	}
	elems := make([]json.RawMessage, result.Count)
	err = json.Unmarshal(result.Results, &elems)
	if err != nil {
		utils.AviLog.Warnf("Failed to unmarshal stringgroup data, err: %v", err)
		return err
	}
	for i := 0; i < len(elems); i++ {
		sg := models.StringGroup{}
		err = json.Unmarshal(elems[i], &sg)
		if err != nil {
			utils.AviLog.Warnf("Failed to unmarshal stringgroup data, err: %v", err)
			continue
		}

		if sg.Name == nil || sg.UUID == nil {
			utils.AviLog.Warnf("Incomplete stringgroup data unmarshalled, %s", utils.Stringify(sg))
			continue
		}

		tenant := getTenantFromTenantRef(*sg.TenantRef)
		stringGroupCacheObj := AviStringGroupCache{
			Name:   *sg.Name,
			Uuid:   *sg.UUID,
			Tenant: tenant,
		}
		if sg.Description != nil {
			stringGroupCacheObj.Description = *sg.Description
		}
		if sg.LongestMatch != nil {
			stringGroupCacheObj.LongestMatch = *sg.LongestMatch
		}
		checksum := lib.StringGroupChecksum(sg.Kv, sg.Markers, sg.LongestMatch, true)
		stringGroupCacheObj.CloudConfigCksum = checksum

		k := NamespaceName{Namespace: tenant, Name: *sg.Name}
		c.StringGroupCache.AviCacheAdd(k, &stringGroupCacheObj)
		utils.AviLog.Debugf("Adding stringgroup to Cache during refresh %s", k)
	}
	return nil
}

func (c *AviObjCache) AviPopulateAllAppPersistenceProfiles(client *clients.AviClient, cloud string, appPersProfileData *[]AviPersistenceProfileCache, nextPage ...NextPage) (*[]AviPersistenceProfileCache, int, error) {
	var uri string
	if len(nextPage) == 1 {
		uri = nextPage[0].NextURI
	} else {
		uri = "/api/applicationpersistenceprofile/?" + "name.contains=" + lib.GetNamePrefix() + "&include_name=true" + "&page_size=100"
	}
	utils.AviLog.Debugf("Get uri %v for applicationpersistenceprofile: ", uri)

	result, err := lib.AviGetCollectionRaw(client, uri)
	if err != nil {
		utils.AviLog.Warnf("Get uri %v returned err for applicationpersistenceprofile %v", uri, err)
		return nil, 0, err
	}
	elems := make([]json.RawMessage, result.Count)
	err = json.Unmarshal(result.Results, &elems)
	if err != nil {
		utils.AviLog.Warnf("Failed to unmarshal applicationpersistenceprofile data, err: %v", err)
		return nil, 0, err
	}
	for i := 0; i < len(elems); i++ {
		appPersProfile := models.ApplicationPersistenceProfile{}
		err = json.Unmarshal(elems[i], &appPersProfile)
		if err != nil {
			utils.AviLog.Warnf("Failed to unmarshal applicationpersistenceprofile data, err: %v", err)
			continue
		}
		if appPersProfile.Name == nil || appPersProfile.UUID == nil {
			utils.AviLog.Warnf("Incomplete applicationpersistenceprofile data unmarshalled, %s", utils.Stringify(appPersProfile))
			continue
		}
		emptyIngestionMarkers := utils.AviObjectMarkers{}
		chksum := lib.PersistenceProfileChecksum(*appPersProfile.Name, *appPersProfile.PersistenceType, emptyIngestionMarkers, appPersProfile.Markers, true)
		if appPersProfile.HTTPCookiePersistenceProfile != nil {
			chksum += lib.HTTPCookiePersistenceProfileChecksum(*appPersProfile.HTTPCookiePersistenceProfile.CookieName, appPersProfile.HTTPCookiePersistenceProfile.Timeout, appPersProfile.HTTPCookiePersistenceProfile.IsPersistentCookie)
		}

		appPersProfileCacheObj := AviPersistenceProfileCache{
			Name:             *appPersProfile.Name,
			Tenant:           getTenantFromTenantRef(*appPersProfile.TenantRef),
			Uuid:             *appPersProfile.UUID,
			CloudConfigCksum: chksum,
			LastModified:     *appPersProfile.LastModified,
			Type:             *appPersProfile.PersistenceType,
		}
		*appPersProfileData = append(*appPersProfileData, appPersProfileCacheObj)
	}

	if result.Next != "" {
		next_uri := strings.Split(result.Next, "/api/applicationpersistenceprofile")
		if len(next_uri) > 1 {
			overrideUri := "/api/applicationpersistenceprofile" + next_uri[1]
			nextPage := NextPage{NextURI: overrideUri}
			_, _, err := c.AviPopulateAllAppPersistenceProfiles(client, cloud, appPersProfileData, nextPage)
			if err != nil {
				return nil, 0, err
			}
		}
	}
	return appPersProfileData, result.Count, nil
}

func (c *AviObjCache) PopulateAppPersistenceProfileToCache(client *clients.AviClient, cloud string) {
	var appPersProfileData []AviPersistenceProfileCache
	setDefaultTenant := session.SetTenant(lib.GetTenant())
	setTenant := session.SetTenant("*")
	setTenant(client.AviSession)
	defer setDefaultTenant(client.AviSession)
	c.AviPopulateAllAppPersistenceProfiles(client, cloud, &appPersProfileData)

	persistenceCacheData := c.AppPersProfileCache.ShallowCopy()
	for i, persistenceCacheObj := range appPersProfileData {
		k := NamespaceName{Namespace: persistenceCacheObj.Tenant, Name: persistenceCacheObj.Name}
		oldPersistenceIntf, found := c.AppPersProfileCache.AviCacheGet(k)
		if found {
			oldPersistenceData, ok := oldPersistenceIntf.(*AviPersistenceProfileCache)
			if ok {
				if oldPersistenceData.InvalidData {
					appPersProfileData[i].InvalidData = true
					utils.AviLog.Infof("Invalid cache data for pki: %s", k)
				}
			} else {
				utils.AviLog.Infof("Wrong data type for pki: %s in cache", k)
			}
		}
		utils.AviLog.Infof("Adding key to persistence profile cache :%s value :%s", k, persistenceCacheObj.Uuid)
		c.AppPersProfileCache.AviCacheAdd(k, &appPersProfileData[i])
		delete(persistenceCacheData, k)
	}
	// The data that is left in persistentCache  should be explicitly removed
	for key := range persistenceCacheData {
		_, ok := key.(NamespaceName)
		if !ok {
			continue
		}
		utils.AviLog.Infof("Deleting key from persistence cache :%s", key)
		c.AppPersProfileCache.AviCacheDelete(key)
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
	if len(overrideUri) == 1 {
		uri = overrideUri[0].NextURI
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
				} else if lib.AKOControlConfig().GetAKOFQDNReusePolicy() == lib.FQDNReusePolicyStrict {
					// call this only when FQDN policy is strict
					hostToIngMapping := svc_mdata_obj.HostToNamespaceIngressName
					utils.AviLog.Debugf("HosttoIng mapping is %v", utils.Stringify(hostToIngMapping))
					if hostToIngMapping != nil {
						// Now populate the map
						PopulateHostToIngMapping(hostToIngMapping)
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
				vs_uuid := ExtractUUID(vs_parent_ref.(string), "virtualservice-.*.#")
				utils.AviLog.Debugf("extracted the vs uuid from parent ref during cache population: %s", vs_uuid)
				// Now let's get the VS key from this uuid
				vsKey, gotVS := c.VsCacheLocal.AviCacheGetKeyByUuid(vs_uuid)
				if gotVS {
					parentVSKey = vsKey.(NamespaceName)
				}

			}
			tenantRef, _ := vs["tenant_ref"].(string)
			tenant := getTenantFromTenantRef(tenantRef)
			if vs["cloud_config_cksum"] != nil {
				k := NamespaceName{Namespace: tenant, Name: vs["name"].(string)}
				*vsCacheCopy = RemoveNamespaceName(*vsCacheCopy, k)
				var vsVipKey []NamespaceName
				var sslKeys []NamespaceName
				var dsKeys []NamespaceName
				var httpKeys []NamespaceName
				var l4Keys []NamespaceName
				var poolgroupKeys []NamespaceName
				var poolKeys []NamespaceName
				var sharedVsOrL4 bool
				var stringgroupKeys []NamespaceName

				// Populate the VSVIP cache
				if vs["vsvip_ref"] != nil {
					// find the vsvip name from the vsvip cache
					vsVipUuid := ExtractUUID(vs["vsvip_ref"].(string), "vsvip-.*.#")
					objKey, objFound := c.VSVIPCache.AviCacheGetKeyByUuid(vsVipUuid)
					if objFound {
						vsVip, foundVip := c.VSVIPCache.AviCacheGet(objKey)
						if foundVip {
							vsVipData, ok := vsVip.(*AviVSVIPCache)
							if ok {
								vipKey := NamespaceName{Namespace: tenant, Name: vsVipData.Name}
								vsVipKey = append(vsVipKey, vipKey)
							}
						}
					}
				}

				if vs["ssl_key_and_certificate_refs"] != nil {
					for _, ssl := range vs["ssl_key_and_certificate_refs"].([]interface{}) {
						// find the sslkey name from the ssl key cache
						sslUuid := ExtractUUID(ssl.(string), "sslkeyandcertificate-.*.#")
						sslName, foundssl := c.SSLKeyCache.AviCacheGetNameByUuid(sslUuid)
						if foundssl {
							sslKey := NamespaceName{Namespace: tenant, Name: sslName.(string)}
							sslKeys = append(sslKeys, sslKey)

							sslIntf, _ := c.SSLKeyCache.AviCacheGet(sslKey)
							sslData := sslIntf.(*AviSSLCache)
							// Populate CAcert if available
							if sslData.CACertUUID != "" {
								caName, found := c.SSLKeyCache.AviCacheGetNameByUuid(sslData.CACertUUID)
								if found {
									caCertKey := NamespaceName{Namespace: tenant, Name: caName.(string)}
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
							dsUuid := ExtractUUID(dsmap["vs_datascript_set_ref"].(string), "vsdatascriptset-.*.#")

							dsName, foundDs := c.DSCache.AviCacheGetNameByUuid(dsUuid)
							if foundDs {
								dsKey := NamespaceName{Namespace: tenant, Name: dsName.(string)}
								// Fetch the associated PGs with the DS.
								dsObj, _ := c.DSCache.AviCacheGet(dsKey)
								for _, pgName := range dsObj.(*AviDSCache).PoolGroups {
									// For each PG, formulate the key and then populate the pg collection cache
									pgKey := NamespaceName{Namespace: tenant, Name: pgName}
									poolgroupKeys = append(poolgroupKeys, pgKey)
									pgpoolKeys := c.AviPGPoolCachePopulate(client, cloud, pgName, tenant)
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
							pgUuid := ExtractUUID(pgmap["service_pool_group_ref"].(string), "poolgroup-.*.#")

							pgName, foundpg := c.PgCache.AviCacheGetNameByUuid(pgUuid)
							if foundpg {
								pgKey := NamespaceName{Namespace: tenant, Name: pgName.(string)}
								poolgroupKeys = append(poolgroupKeys, pgKey)
								pgpoolKeys := c.AviPGPoolCachePopulate(client, cloud, pgName.(string), tenant)
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
							l4PolUuid := ExtractUUID(l4map["l4_policy_set_ref"].(string), "l4policyset-.*.#")
							l4Name, foundl4pol := c.L4PolicyCache.AviCacheGetNameByUuid(l4PolUuid)
							if foundl4pol {
								sharedVsOrL4 = true
								l4key := NamespaceName{Namespace: tenant, Name: l4Name.(string)}
								l4Obj, _ := c.L4PolicyCache.AviCacheGet(l4key)
								for _, poolName := range l4Obj.(*AviL4PolicyCache).Pools {
									poolKey := NamespaceName{Namespace: tenant, Name: poolName}
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
							httpUuidStr, ok := httpmap["http_policy_set_ref"].(string)
							if ok {
								httpUuid := ExtractUUID(httpUuidStr, "httppolicyset-.*.#")
								httpName, foundhttp := c.HTTPPolicyCache.AviCacheGetNameByUuid(httpUuid)
								// If the httppol is not found in the cache, do an explicit get
								if !foundhttp && !sharedVsOrL4 {
									err = c.AviPopulateHttpPolicySetbyUUID(client, httpUuid)
									// If still the httpName is not found. Log an error saying, this VS may not behave appropriately.
									if err != nil {
										utils.AviLog.Warnf("HTTPPolicySet not found in Avi for VS: %s for httpUUID: %s", vs["name"].(string), httpUuid)
									} else {
										httpName, foundhttp = c.HTTPPolicyCache.AviCacheGetNameByUuid(httpUuid)
									}
								}
								if foundhttp {
									httpKey := NamespaceName{Namespace: tenant, Name: httpName.(string)}
									httpObj, _ := c.HTTPPolicyCache.AviCacheGet(httpKey)
									for _, pgName := range httpObj.(*AviHTTPPolicyCache).PoolGroups {
										// For each PG, formulate the key and then populate the pg collection cache
										pgKey := NamespaceName{Namespace: tenant, Name: pgName}
										poolgroupKeys = append(poolgroupKeys, pgKey)
										pgpoolKeys := c.AviPGPoolCachePopulate(client, cloud, pgName, tenant)
										poolKeys = append(poolKeys, pgpoolKeys...)
									}
									for _, sgName := range httpObj.(*AviHTTPPolicyCache).StringGroupRefs {
										sgKey := NamespaceName{Namespace: tenant, Name: sgName}
										stringgroupKeys = append(stringgroupKeys, sgKey)
									}
									httpKeys = append(httpKeys, httpKey)
								}
							} else {
								utils.AviLog.Warnf("No httppolicyset UUID found in http_policy_set_ref for VS: %s", vs["name"].(string))
							}
						}
					}
				}
				if vs["pool_ref"] != nil {
					poolRef, ok := vs["pool_ref"].(string)
					if ok {
						poolNameFromRef := strings.Split(poolRef, "#")[1]
						poolUuid := ExtractUUID(poolRef, "pool-.*.#")
						poolNameFromCache, foundPool := c.PoolCache.AviCacheGetNameByUuid(poolUuid)
						if foundPool && poolNameFromCache.(string) == poolNameFromRef {
							poolKey := NamespaceName{Namespace: tenant, Name: poolNameFromCache.(string)}
							poolKeys = append(poolKeys, poolKey)
						}
					}
				}

				// Populate the vscache meta object here.
				vsMetaObj := AviVsCache{
					Name:                     vs["name"].(string),
					Tenant:                   tenant,
					Uuid:                     vs["uuid"].(string),
					VSVipKeyCollection:       vsVipKey,
					HTTPKeyCollection:        httpKeys,
					DSKeyCollection:          dsKeys,
					SSLKeyCertCollection:     sslKeys,
					PGKeyCollection:          poolgroupKeys,
					PoolKeyCollection:        poolKeys,
					CloudConfigCksum:         vs["cloud_config_cksum"].(string),
					SNIChildCollection:       sni_child_collection,
					ParentVSRef:              parentVSKey,
					ServiceMetadataObj:       svc_mdata_obj,
					L4PolicyCollection:       l4Keys,
					LastModified:             vs["_last_modified"].(string),
					StringGroupKeyCollection: stringgroupKeys,
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
				nextPage := NextPage{NextURI: overrideUri}
				c.AviObjVSCachePopulate(client, cloud, vsCacheCopy, nextPage)
			}
		}
	}
	return nil
}

// Upfront populate mapping so that during FullSyncK8s, it will be used to assign ingresses/routes to appropriate hosts list.
func PopulateHostToIngMapping(hostsToIng map[string][]string) {
	isRoute := false
	if utils.GetInformers().RouteInformer != nil {
		isRoute = true
	}
	var routeNamespaceName objects.RouteNamspaceName
	for host, ings := range hostsToIng {
		// append each ingress in active list
		utils.AviLog.Debugf("Populating Ingress mapping for host : %s", host)
		for _, ing := range ings {
			namespace, _, name := lib.ExtractTypeNameNamespace(ing)
			// Fetch ingress using clientset. From informer couldn't fetch it.
			if !isRoute {
				ingObj, err := utils.GetInformers().IngressInformer.Lister().Ingresses(namespace).Get(name)
				if err != nil {
					utils.AviLog.Errorf("Unable to retrieve the ingress %s/%s during populating host to ingress map in populate cache: %s", namespace, name, err)
					continue
				}
				routeNamespaceName = objects.RouteNamspaceName{
					RouteNSRouteName: utils.Ingress + "/" + ing,
					CreationTime:     ingObj.CreationTimestamp,
				}
			} else {
				routeObj, err := utils.GetInformers().RouteInformer.Lister().Routes(namespace).Get(name)
				if err != nil {
					utils.AviLog.Errorf("Unable to retrieve the ingress %s/%s during populating host to ingress map in populate cache: %s", namespace, name, err)
					continue
				}
				routeNamespaceName = objects.RouteNamspaceName{
					RouteNSRouteName: utils.OshiftRoute + "/" + ing,
					CreationTime:     routeObj.CreationTimestamp,
				}
			}

			// Add it to the structure
			objects.SharedUniqueNamespaceLister().UpdateHostnameToRoute(host, routeNamespaceName)

		}
	}
}

func (c *AviObjCache) AviObjOneVSCachePopulate(client *clients.AviClient, cloud string, vsName, tenant string) error {
	// This method should be called only from layer-3 during a retry.
	var rest_response interface{}
	akoUser := lib.AKOUser
	var uri string

	uri = "/api/virtualservice?name=" + vsName + "&cloud_ref.name=" + cloud + "&include_name=true" + "&created_by=" + akoUser

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
		k := NamespaceName{Namespace: tenant, Name: vsName}
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
				} else if lib.AKOControlConfig().GetAKOFQDNReusePolicy() == lib.FQDNReusePolicyStrict {
					// call this only when FQDN policy is strict
					hostToIngMapping := svc_mdata_obj.HostToNamespaceIngressName
					utils.AviLog.Debugf("HosttoIng mapping is %v", utils.Stringify(hostToIngMapping))
					if hostToIngMapping != nil {
						// Now populate the map
						PopulateHostToIngMapping(hostToIngMapping)
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
				vsUUID := ExtractUUID(vs_parent_ref.(string), "virtualservice-.*.#")
				utils.AviLog.Debugf("Extracted the vs uuid from parent ref during cache population: %s", vsUUID)
				// Now let's get the VS key from this uuid
				vsKey, gotVS := c.VsCacheMeta.AviCacheGetKeyByUuid(vsUUID)
				if gotVS {
					parentVSKey = vsKey.(NamespaceName)
				}

			}
			tenantRef, _ := vs["tenant_ref"].(string)
			tenant := getTenantFromTenantRef(tenantRef)
			if vs["cloud_config_cksum"] != nil {
				var vsVipKey []NamespaceName
				var sslKeys []NamespaceName
				var dsKeys []NamespaceName
				var httpKeys []NamespaceName
				var poolgroupKeys []NamespaceName
				var poolKeys []NamespaceName
				var l4Keys []NamespaceName
				var stringgroupKeys []NamespaceName

				// Populate the VSVIP cache
				if vs["vsvip_ref"] != nil {
					// find the vsvip name from the vsvip cache
					vsVipUuid := ExtractUUID(vs["vsvip_ref"].(string), "vsvip-.*.#")
					vsVipName, foundVip := c.VSVIPCache.AviCacheGetNameByUuid(vsVipUuid)

					if foundVip {
						vipKey := NamespaceName{Namespace: tenant, Name: vsVipName.(string)}
						vsVipKey = append(vsVipKey, vipKey)
					}
				}

				if vs["ssl_key_and_certificate_refs"] != nil {
					for _, ssl := range vs["ssl_key_and_certificate_refs"].([]interface{}) {
						// find the sslkey name from the ssl key cache
						sslUuid := ExtractUUID(ssl.(string), "sslkeyandcertificate-.*.#")
						sslName, foundssl := c.SSLKeyCache.AviCacheGetNameByUuid(sslUuid)
						if foundssl {
							sslKey := NamespaceName{Namespace: tenant, Name: sslName.(string)}
							sslKeys = append(sslKeys, sslKey)

							sslIntf, _ := c.SSLKeyCache.AviCacheGet(sslKey)
							sslData := sslIntf.(*AviSSLCache)
							// Populate CAcert if available
							if sslData.CACertUUID != "" {
								caName, found := c.SSLKeyCache.AviCacheGetNameByUuid(sslData.CACertUUID)
								if found {
									caCertKey := NamespaceName{Namespace: tenant, Name: caName.(string)}
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
							dsUuid := ExtractUUID(dsmap["vs_datascript_set_ref"].(string), "vsdatascriptset-.*.#")

							dsName, foundDs := c.DSCache.AviCacheGetNameByUuid(dsUuid)
							if foundDs && !strings.Contains(dsName.(string), "ako-gw") {
								dsKey := NamespaceName{Namespace: tenant, Name: dsName.(string)}
								// Fetch the associated PGs with the DS.
								dsObj, _ := c.DSCache.AviCacheGet(dsKey)
								for _, pgName := range dsObj.(*AviDSCache).PoolGroups {
									// For each PG, formulate the key and then populate the pg collection cache
									pgKey := NamespaceName{Namespace: tenant, Name: pgName}
									poolgroupKeys = append(poolgroupKeys, pgKey)
									pgpoolKeys := c.AviPGPoolCachePopulate(client, cloud, pgName, tenant)
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
							pgUuid := ExtractUUID(pgmap["service_pool_group_ref"].(string), "poolgroup-.*.#")

							pgName, foundpg := c.PgCache.AviCacheGetNameByUuid(pgUuid)
							if foundpg {
								pgKey := NamespaceName{Namespace: tenant, Name: pgName.(string)}
								poolgroupKeys = append(poolgroupKeys, pgKey)
								pgpoolKeys := c.AviPGPoolCachePopulate(client, cloud, pgName.(string), tenant)
								poolKeys = append(poolKeys, pgpoolKeys...)
							}
						}
					}
				}
				if vs["l4_policies"] != nil {
					for _, l4_intf := range vs["l4_policies"].([]interface{}) {
						l4map, ok := l4_intf.(map[string]interface{})
						if ok {
							l4PolUuid := ExtractUUID(l4map["l4_policy_set_ref"].(string), "l4policyset-.*.#")
							l4Name, foundl4pol := c.L4PolicyCache.AviCacheGetNameByUuid(l4PolUuid)
							if foundl4pol {
								l4key := NamespaceName{Namespace: tenant, Name: l4Name.(string)}
								l4Obj, _ := c.L4PolicyCache.AviCacheGet(l4key)
								for _, poolName := range l4Obj.(*AviL4PolicyCache).Pools {
									poolKey := NamespaceName{Namespace: tenant, Name: poolName}
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
							httpUuidStr, ok := httpmap["http_policy_set_ref"].(string)
							if ok {
								httpUuid := ExtractUUID(httpUuidStr, "httppolicyset-.*.#")
								httpName, foundhttp := c.HTTPPolicyCache.AviCacheGetNameByUuid(httpUuid)
								if foundhttp {
									httpKey := NamespaceName{Namespace: tenant, Name: httpName.(string)}
									httpObj, _ := c.HTTPPolicyCache.AviCacheGet(httpKey)
									for _, pgName := range httpObj.(*AviHTTPPolicyCache).PoolGroups {
										// For each PG, formulate the key and then populate the pg collection cache
										pgKey := NamespaceName{Namespace: tenant, Name: pgName}
										poolgroupKeys = append(poolgroupKeys, pgKey)
										pgpoolKeys := c.AviPGPoolCachePopulate(client, cloud, pgName, tenant)
										poolKeys = append(poolKeys, pgpoolKeys...)
									}
									for _, sgName := range httpObj.(*AviHTTPPolicyCache).StringGroupRefs {
										sgKey := NamespaceName{Namespace: tenant, Name: sgName}
										stringgroupKeys = append(stringgroupKeys, sgKey)
									}
									httpKeys = append(httpKeys, httpKey)
								}
							} else {
								utils.AviLog.Warnf("No httppolicyset UUID found in http_policy_set_ref for VS: %s", vs["name"].(string))
							}
						}
					}
				}
				if vs["pool_group_ref"] != nil {
					pgRef, ok := vs["pool_group_ref"].(string)
					if ok {
						pgUuid := ExtractUUID(pgRef, "poolgroup-.*.#")
						pgName, foundpg := c.PgCache.AviCacheGetNameByUuid(pgUuid)
						if foundpg {
							pgKey := NamespaceName{Namespace: tenant, Name: pgName.(string)}
							poolgroupKeys = append(poolgroupKeys, pgKey)
							pgpoolKeys := c.AviPGPoolCachePopulate(client, cloud, pgName.(string), tenant)
							poolKeys = append(poolKeys, pgpoolKeys...)
						}
					}
				}
				// Populate the vscache meta object here.
				vsMetaObj := AviVsCache{
					Name:                     vs["name"].(string),
					Tenant:                   tenant,
					Uuid:                     vs["uuid"].(string),
					VSVipKeyCollection:       vsVipKey,
					HTTPKeyCollection:        httpKeys,
					DSKeyCollection:          dsKeys,
					SSLKeyCertCollection:     sslKeys,
					PGKeyCollection:          poolgroupKeys,
					PoolKeyCollection:        poolKeys,
					CloudConfigCksum:         vs["cloud_config_cksum"].(string),
					SNIChildCollection:       sni_child_collection,
					ParentVSRef:              parentVSKey,
					L4PolicyCollection:       l4Keys,
					ServiceMetadataObj:       svc_mdata_obj,
					StringGroupKeyCollection: stringgroupKeys,
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

func (c *AviObjCache) AviPGPoolCachePopulate(client *clients.AviClient, cloud string, pgName string, tenant string) []NamespaceName {
	var poolKeyCollection []NamespaceName

	k := NamespaceName{Namespace: tenant, Name: pgName}
	// Find the pools associated with this PG and populate them
	pgObj, ok := c.PgCache.AviCacheGet(k)
	// Get the members from this and populate the VS ref
	if ok {
		for _, poolName := range pgObj.(*AviPGCache).Members {
			k := NamespaceName{Namespace: tenant, Name: poolName}
			poolKeyCollection = append(poolKeyCollection, k)
		}
	} else {
		// PG not found in the cache. Let's try a refresh explicitly
		c.AviPopulateOnePGCache(client, cloud, pgName)
		pgObj, ok = c.PgCache.AviCacheGet(k)
		if ok {
			utils.AviLog.Debugf("Found PG on refresh: %s", pgName)
			for _, poolName := range pgObj.(*AviPGCache).Members {
				k := NamespaceName{Namespace: tenant, Name: poolName}
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

	if !utils.IsWCP() {
		subdomains := c.AviDNSPropertyPopulate(client, *cloud.UUID)
		if len(subdomains) == 0 {
			utils.AviLog.Warnf("Cloud: %v does not have a dns provider configured", cloudName)
		}
		if subdomains != nil {
			cloud_obj.NSIpamDNS = subdomains
		}
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

func ValidateUserInput(client *clients.AviClient, isGateway bool) (bool, error) {
	// add other step0 validation logics here -> isValid := check1 && check2 && ...

	var err error
	// default it to true, only for VCenter use this flag
	isVRFValid := true
	isTenantValid := checkTenant(client, &err)
	isCloudValid := checkAndSetCloudType(client, &err)
	// Check VRF for both VCENTER and NO Access cloud
	if lib.GetCloudType() == lib.CLOUD_VCENTER || lib.GetCloudType() == lib.CLOUD_NONE {
		isVRFValid = checkVRF(client, &err)
		if !isVRFValid {
			utils.AviLog.Warnf("Invalid input detected, AKO will be rebooted to retry %s", err.Error())
			lib.AKOControlConfig().PodEventf(corev1.EventTypeWarning, lib.AKOShutdown, "Invalid user input %s", err.Error())
			lib.ShutdownApi()
			return isVRFValid, err
		}
	}
	isRequiredValuesValid := checkRequiredValuesYaml(client, isGateway, &err)
	if utils.IsWCP() {
		if isTenantValid &&
			isCloudValid &&
			isRequiredValuesValid {
			utils.AviLog.Infof("All values verified for advanced L4, proceeding with bootup.")
			return true, nil
		}
		return false, err
	}

	isSegroupValid := validateAndConfigureSeGroup(client, &err)
	isNodeNetworkValid := checkNodeNetwork(client, &err)
	isBGPConfigurationValid := checkBGPParams(&err)
	isPublicCloudConfigValid := checkPublicCloud(client, &err)
	checkedAndSetVRFConfig := checkAndSetVRFFromNetwork(client, &err)
	isCNIConfigValid := lib.IsValidCni(&err)
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

func findHostRefs(client *clients.AviClient, nwUUID string) []string {
	uri := "/api/vimgrnwruntime/" + nwUUID
	mgrRuntime := models.VIMgrNWRuntime{}
	var hostsRef []string
	if nwUUID != "" {
		err := lib.AviGet(client, uri, &mgrRuntime)
		if err != nil {
			utils.AviLog.Warnf("Error while retrieving cloud mgmt network %v", err)
			return hostsRef
		}
		hostsRef = mgrRuntime.HostRefs
	}
	return hostsRef
}

// Function to find max host overlap between mgmt network and vipNetwork/nodenetwoeklist
func findHostWithMaxOverlapping(segMgmtNetwork string, client *clients.AviClient, localNetworkList []models.Network) akov1beta1.AviInfraSettingVipNetwork {
	cloudMgmtNW := lib.GetCloudMgmtNetwork()
	// Use SEG Mgmt network to fetch host refs
	if segMgmtNetwork != "" {
		cloudMgmtNW = segMgmtNetwork
	}
	var matchedNW akov1beta1.AviInfraSettingVipNetwork
	mgmtHostRefs := findHostRefs(client, cloudMgmtNW)
	utils.AviLog.Infof("For Management network:%s, hosts are: %v", cloudMgmtNW, utils.Stringify(mgmtHostRefs))
	mgmtHostsSet := sets.NewString(mgmtHostRefs...)
	//default choice of network
	desiredNW := localNetworkList[0]
	nwHashMap := make(map[string]models.Network)
	//create priority Queue
	pqNetworks := pq.New()

	for _, nw := range localNetworkList {
		hostRefs := findHostRefs(client, *nw.UUID)
		utils.AviLog.Infof("For network %s with uuid %s, hosts are: %v", *nw.Name, *nw.UUID, utils.Stringify(hostRefs))
		hostRefsSet := sets.NewString(hostRefs...)
		matchedHostSet := mgmtHostsSet.Intersection(hostRefsSet)

		// If no overlap of hosts between network and mgmt network do not add
		if matchedHostSet.Len() != 0 {
			// Insert into PQ uuid of network in descending order
			pqNetworks.Insert(*nw.UUID, -float64(matchedHostSet.Len()))
			// Add hashmap entry for that network
			nwHashMap[*nw.UUID] = nw
		}
	}
	nwElement, err := pqNetworks.Pop()
	if err == nil {
		// default desired network of PQ has entries.
		networkUUID := nwElement.(string)
		desiredNW = nwHashMap[networkUUID]
	}
	for err == nil {
		networkUUID := nwElement.(string)
		network := nwHashMap[networkUUID]
		if network.ConfiguredSubnets != nil {
			desiredNW = network
			break
		}
		nwElement, err = pqNetworks.Pop()
	}
	matchedNW.NetworkName = *desiredNW.Name
	matchedNW.NetworkUUID = *desiredNW.UUID
	return matchedNW
}

func FindCIDROverlapping(networks []models.Network, ipNet akov1beta1.AviInfraSettingVipNetwork) (bool, models.Network) {
	localVIPNetwork := models.Network{}
	countOfCidrMatchReq := 0
	if ipNet.Cidr != "" {
		countOfCidrMatchReq += 1
	}
	if ipNet.V6Cidr != "" {
		countOfCidrMatchReq += 1
	}
	utils.AviLog.Infof("Performing CIDR match for Network: %v", utils.Stringify(ipNet))
	networkFound := false
	//Go through fetched network's cidr and match it against cidr given in configmap or aviinfra
	for _, network := range networks {
		matchedCidrCount := 0
		networkFound = false
		// Do cidr match first then do host overlapping
		if countOfCidrMatchReq > 0 {
			//check configured subnets are matching with given cidr.
			// IF matched, use that network.
			utils.AviLog.Infof("For Network %v Configured subnet is: %v", *network.Name, utils.Stringify(network.ConfiguredSubnets))
			for _, cidr := range network.ConfiguredSubnets {
				addr := fmt.Sprintf("%s/%v", *cidr.Prefix.IPAddr.Addr, *cidr.Prefix.Mask)
				if *cidr.Prefix.IPAddr.Type == "V4" {
					if ipNet.Cidr != "" && ipNet.Cidr == addr {
						matchedCidrCount += 1
					}
				} else if *cidr.Prefix.IPAddr.Type == "V6" {
					if ipNet.V6Cidr != "" && ipNet.V6Cidr == addr {
						matchedCidrCount += 1
					}
				}
				//
				if countOfCidrMatchReq == matchedCidrCount {
					networkFound = true
					break
				}
				matchedCidrCount = 0
			}
		}
		if networkFound {
			// If cidr matched network found, reset all list and add only that network
			localVIPNetwork = network
			break
		}
	}
	return networkFound, localVIPNetwork
}

// This is called for Vcenter and No access cloud only
func PopulateVipNetworkwithUUID(segMgmtNetwork string, client *clients.AviClient, vipNetworks []akov1beta1.AviInfraSettingVipNetwork) ([]akov1beta1.AviInfraSettingVipNetwork, error) {
	var ipNetworkList []akov1beta1.AviInfraSettingVipNetwork
	var ipNetwork akov1beta1.AviInfraSettingVipNetwork
	cmVRFName := lib.AKOControlConfig().ControllerVRFContext()
	var retErr error
	// In Public cloud we allow multiple network, so loop.
	for _, vipNet := range vipNetworks {
		// If Network uuid is present, use that.
		if vipNet.NetworkUUID != "" {
			ipNetwork = akov1beta1.AviInfraSettingVipNetwork{
				NetworkName: vipNet.NetworkName,
				NetworkUUID: vipNet.NetworkUUID,
				Cidr:        vipNet.Cidr,
				V6Cidr:      vipNet.V6Cidr,
			}
			// Whether vrfContext is correct or not.
			// This vip list is mentioned in configmap and aviinfrasetting
			// only add this when vrf is mentioned in configmap
			if cmVRFName != "" {
				uri := fmt.Sprintf("/api/network/%s?cloud_uuid=%s&include_name", vipNet.NetworkUUID, lib.GetCloudUUID())
				var rest_response interface{}
				err := lib.AviGet(client, uri, &rest_response)
				if err != nil || rest_response == nil {
					utils.AviLog.Warnf("No network with UUID %s found", vipNet.NetworkUUID)
					retErr = fmt.Errorf("no network with UUID %s found", vipNet.NetworkUUID)
					continue
				}
				result := rest_response.(map[string]interface{})
				tempVrf := result["vrf_context_ref"].(string)
				if tempVrf != "" {
					vrf_uuid_name := strings.Split(tempVrf, "#")
					if len(vrf_uuid_name) != 2 || vrf_uuid_name[1] != cmVRFName {
						utils.AviLog.Warnf("Network %s does not have vrf %s", vipNet.NetworkUUID, cmVRFName)
						retErr = fmt.Errorf("network %s does not have vrf %s", vipNet.NetworkUUID, cmVRFName)
						continue
					}
				}
			}

		} else {
			//default value
			ipNetwork = akov1beta1.AviInfraSettingVipNetwork{
				NetworkName: vipNet.NetworkName,
				Cidr:        vipNet.Cidr,
				V6Cidr:      vipNet.V6Cidr,
			}
			// For Each network from config/aviinfra, perform following set of operations.
			//  of vrfcontext and tenant for vcenter cloud (it will good to fetch vrfuuid and use it)
			// check network against vrf and tenant name
			localVIPNetworkList := []models.Network{}
			networkURI := "/api/network/?include_name=true&name=" + vipNet.NetworkName + "&cloud_ref.name=" + utils.CloudName
			// only add this when vrf is mentioned in configmap. Achieves backward compatibility even if vrf is not mentioned
			if cmVRFName != "" {
				networkURI = networkURI + "&vrf_context_ref.name=" + cmVRFName
			}
			result, err := lib.AviGetCollectionRaw(client, networkURI)
			if err != nil {
				utils.AviLog.Warnf("Error while retrieving network %v details. Error: %v", vipNet.NetworkName, err)
				retErr = fmt.Errorf("error while retrieving network %v details. Error: %v", vipNet.NetworkName, err)
				continue
			}
			elems := make([]json.RawMessage, result.Count)
			err = json.Unmarshal(result.Results, &elems)
			if err != nil {
				utils.AviLog.Warnf("Failed to unmarshal data, err: %v", err)
				retErr = fmt.Errorf("failed to unmarshal data, err: %v", err)
				continue
			}
			if result.Count == 0 {
				//network not found
				retErr = fmt.Errorf("network %s not found under vrf %s", vipNet.NetworkName, cmVRFName)
				continue
			}
			for _, elem := range elems {
				net := models.Network{}
				if err = json.Unmarshal(elem, &net); err != nil {
					utils.AviLog.Warnf("Failed to unmarshal network  data, err: %v", err)
					retErr = fmt.Errorf("failed to unmarshal network  data, err: %v", err)
					continue
				}
				localVIPNetworkList = append(localVIPNetworkList, net)
			}

			// For no access cloud, user manages the network. So there should not be duplicate networks.
			if len(localVIPNetworkList) > 1 {
				//first check cidr matching
				found, netLocal := FindCIDROverlapping(localVIPNetworkList, ipNetwork)
				if found {
					utils.AviLog.Infof("Network found from CIDR overlapping is: %v", utils.Stringify(netLocal))
					ipNetwork = akov1beta1.AviInfraSettingVipNetwork{
						NetworkName: *netLocal.Name,
						NetworkUUID: *netLocal.UUID,
						Cidr:        vipNet.Cidr,
						V6Cidr:      vipNet.V6Cidr,
					}
				} else {
					// Then do host uuid mapping and return max host-uuid overlapping network
					ipNetwork = findHostWithMaxOverlapping(segMgmtNetwork, client, localVIPNetworkList)
					ipNetwork.Cidr = vipNet.Cidr
					ipNetwork.V6Cidr = vipNet.V6Cidr
					utils.AviLog.Infof("Network found from Host overlapping is: %v", utils.Stringify(ipNetwork))
				}
			}
			if len(localVIPNetworkList) == 1 || ipNetwork == (akov1beta1.AviInfraSettingVipNetwork{}) {
				// If empty network returned or len 1, fill with first network
				// with cidr provided in configmap or aviinfra

				ipNetwork = akov1beta1.AviInfraSettingVipNetwork{
					NetworkName: *localVIPNetworkList[0].Name,
					Cidr:        vipNet.Cidr,
					V6Cidr:      vipNet.V6Cidr,
				}
				// do not add uuid if number of networks retrieved are 1. so that cksum will not change
				if len(localVIPNetworkList) > 1 {
					ipNetwork.NetworkUUID = *localVIPNetworkList[0].UUID
				}
			}
		}
		ipNetworkList = append(ipNetworkList, ipNetwork)
	}
	return ipNetworkList, retErr
}

func checkRequiredValuesYaml(client *clients.AviClient, isGateway bool, returnErr *error) bool {
	if _, err := lib.IsClusterNameValid(); err != nil {
		*returnErr = err
		return false
	}

	// Set the ako user with prefix
	// after clusterName validation, set AKO User to be used in created_by fields for Avi Objects
	if !isGateway {
		lib.SetNamePrefix("")
		lib.SetAKOUser(lib.AKOPrefix)
	}
	//Set clusterlabel checksum
	lib.SetClusterLabelChecksum()

	cloudName := utils.CloudName
	if cloudName == "" {
		*returnErr = fmt.Errorf("required param cloudName not specified, syncing will be disabled")
		return false
	}

	if vipList, err := lib.GetVipNetworkListEnv(); err != nil {
		*returnErr = fmt.Errorf("error in getting VIP network %s, shutting down AKO", err)
		return false
	} else if len(vipList) > 0 {

		vipListUpdated := vipList
		var err error
		if lib.GetCloudType() == lib.CLOUD_VCENTER || lib.GetCloudType() == lib.CLOUD_NONE {
			segMgmtNetwork := ""
			if lib.GetCloudType() == lib.CLOUD_VCENTER {
				segMgmtNetwork = GetCMSEGManagementNetwork(client)
			}
			vipListUpdated, err = PopulateVipNetworkwithUUID(segMgmtNetwork, client, vipList)
			if err != nil {
				*returnErr = err
				return false
			}
		}
		utils.SetVipNetworkList(vipListUpdated)
		return true
	}

	// check if config map exists
	// TODO: Check if this code will ever git hit
	k8sClient := utils.GetInformers().ClientSet
	aviCMNamespace := utils.GetAKONamespace()
	if lib.GetNamespaceToSync() != "" {
		aviCMNamespace = lib.GetNamespaceToSync()
	}
	_, err := k8sClient.CoreV1().ConfigMaps(aviCMNamespace).Get(context.TODO(), lib.AviConfigMap, metav1.GetOptions{})
	if err != nil {
		*returnErr = fmt.Errorf("configmap %s/%s not found, error: %v, syncing will be disabled", aviCMNamespace, lib.AviConfigMap, err)
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
		infraSettingList, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().AviInfraSettings().List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			utils.AviLog.Warnf("Unable to list AviInfraSettings %s", err.Error())
		}
		for _, setting := range infraSettingList.Items {
			seGroupSet.Insert(setting.Spec.SeGroup.Name)
		}
	}
	seGroupSet.Insert(lib.GetSEGName())

	// This assumes that a single cluster won't use more than 100 distinct SEGroups.
	uri := "/api/serviceenginegroup/?include_name&page_size=100&cloud_ref.name=" + utils.CloudName + "&name.in=" + strings.Join(seGroupSet.List(), ",")
	var result session.AviCollectionResult
	result, err := lib.AviGetCollectionRaw(client, uri)
	if err != nil {
		if aviError, ok := err.(session.AviError); ok && aviError.HttpStatusCode == 403 {
			//SE in provider context no read access
			utils.AviLog.Debugf("Switching to admin context from  %s", lib.GetTenant())
			//fallback to Admin Tenant
			client = SharedAVIClients(lib.GetAdminTenant()).AviClient[0]
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

func refreshSeGroupDataAndCheckLabel(client *clients.AviClient, seGroup *models.ServiceEngineGroup) error {
	uri := "/api/serviceenginegroup/" + *seGroup.UUID
	response := models.ServiceEngineGroup{}
	err := lib.AviGet(client, uri, &response)
	if err != nil {
		utils.AviLog.Warnf("ServiceEngineGroup Get uri %v returned err %v", uri, err)
		return err
	}
	for _, label := range response.Labels {
		if *label.Key == lib.ClusterNameLabelKey && label.Value != nil && *label.Value == *lib.GetLabels()[0].Value {
			return nil
		}
	}

	return fmt.Errorf("Labels do not match with cluster name for SE group :%v. Expected Labels: %v", seGroup.Name, utils.Stringify(lib.GetLabels()))
}

// ConfigureSeGroupLabels configures labels on the SeGroup if not present already
func ConfigureSeGroupLabels(client *clients.AviClient, seGroup *models.ServiceEngineGroup) error {

	labels := seGroup.Labels
	segName := *seGroup.Name
	if len(labels) == 0 {
		uri := "/api/serviceenginegroup/" + *seGroup.UUID
		seGroup.Labels = lib.GetLabels()
		response := models.ServiceEngineGroupAPIResponse{}
		err := lib.AviPut(client, uri, seGroup, response)
		if err != nil {
			if aviError, ok := err.(session.AviError); ok && aviError.HttpStatusCode == 400 {
				//SE in provider context
				utils.AviLog.Debugf("Switching to admin context from  %s", lib.GetTenant())
				client = SharedAVIClients(lib.GetAdminTenant()).AviClient[0]
				err := lib.AviPut(client, uri, seGroup, response)
				if err != nil {
					utils.AviLog.Warnf("Setting labels on Service Engine Group :%v failed with error :%v. Expected Labels: %v", segName, err.Error(), utils.Stringify(lib.GetLabels()))
					if aviError, ok := err.(session.AviError); ok && aviError.HttpStatusCode == 412 {
						err := refreshSeGroupDataAndCheckLabel(client, seGroup)
						if err != nil {
							return fmt.Errorf("Setting labels on Service Engine Group :%v failed with error :%v. Expected Labels: %v", segName, err.Error(), utils.Stringify(lib.GetLabels()))
						}
					} else {
						return fmt.Errorf("Setting labels on Service Engine Group :%v failed with error :%v. Expected Labels: %v", segName, err.Error(), utils.Stringify(lib.GetLabels()))
					}
				}
			} else if aviError.HttpStatusCode == 412 {
				err := refreshSeGroupDataAndCheckLabel(client, seGroup)
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

	for _, label := range seGroup.Labels {
		if *label.Key == lib.ClusterNameLabelKey && *label.Value == *lib.GetLabels()[0].Value {
			return nil
		}
	}
	return fmt.Errorf("Labels do not match with cluster name for SE group :%v. Expected Labels: %v", segName, utils.Stringify(lib.GetLabels()))
}

// DeConfigureSeGroupLabels deconfigures labels on the SeGroup.
func DeConfigureSeGroupLabels() {

	if !lib.AKOControlConfig().IsLeader() {
		return
	}

	if len(lib.GetLabels()) == 0 {
		return
	}
	segName := lib.GetSEGName()
	clients := SharedAVIClients(lib.GetTenant())
	aviClientLen := lib.GetshardSize()
	var index uint32
	if aviClientLen != 0 {
		index = aviClientLen - 1
	}
	client := clients.AviClient[index]
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
			client = SharedAVIClients(lib.GetAdminTenant()).AviClient[0]
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
	uri := "/api/serviceenginegroup/?include_name&name=" + segName + "&cloud_ref.name=" + utils.CloudName
	var result session.AviCollectionResult
	result, err := lib.AviGetCollectionRaw(client, uri)
	if err != nil {
		if aviError, ok := err.(session.AviError); ok && aviError.HttpStatusCode == 403 {
			//SE in provider context no read access
			utils.AviLog.Debugf("Switching to admin context from  %s", lib.GetTenant())
			client = SharedAVIClients(lib.GetAdminTenant()).AviClient[0]
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
	uri := "/api/tenant/?include_name&name=" + lib.GetEscapedValue(lib.GetTenant())
	result, err := lib.AviGetCollectionRaw(client, uri)
	if err != nil {
		*returnError = fmt.Errorf("get uri %v returned err %v", uri, err)
		return false
	}

	if result.Count != 1 {
		*returnError = fmt.Errorf("tenant details not found for the tenant: %s", lib.GetTenant())
		return false
	}
	return true
}

// Check VRF in given tenant
// IF 403 or 404, switch to Admin tenant
func checkVRF(client *clients.AviClient, returnError *error) bool {
	// Here fetch vrf details for vcenter cloud
	vrfName := lib.AKOControlConfig().ControllerVRFContext()
	if vrfName != "" {
		uri := "/api/vrfcontext/?include_name&name=" + vrfName + "&cloud_ref.name=" + utils.CloudName
		result, err := lib.AviGetCollectionRaw(client, uri)
		if err != nil {
			if aviError, ok := err.(session.AviError); ok && (aviError.HttpStatusCode == 403 || aviError.HttpStatusCode == 404) {
				utils.AviLog.Debugf("Switching to admin context from  %s", lib.GetTenant())
				client := SharedAVIClients(lib.GetAdminTenant()).AviClient[0]
				result, err = lib.AviGetCollectionRaw(client, uri)
				if err != nil {
					*returnError = fmt.Errorf("get uri %v returned err %v", uri, err)
					return false
				}
			} else {
				*returnError = fmt.Errorf("get uri %v returned err %v", uri, err)
				return false
			}
		}
		if result.Count != 1 {
			*returnError = fmt.Errorf("vrf %s details not found", vrfName)
			return false
		}
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
	if cloud.VcenterConfiguration != nil {
		// This set cloud mgmt network in vimgrruntime format
		// TODO: Fetch it from SE Group defined.
		lib.SetCloudMgmtNetwork(*cloud.VcenterConfiguration.ManagementNetwork)
	}
	// IPAM is mandatory for vcenter, nsxt and noaccess cloud but not for public clouds and nsxt cloud in VPC mode
	if !lib.IsPublicCloud() && !lib.GetVPCMode() && cloud.IPAMProviderRef == nil {
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
	var ret_err error
	// 1. User input
	if vipList, err := lib.GetVipNetworkListEnv(); err != nil {
		return false, fmt.Errorf("error in getting VIP network %s, shutting down AKO", err)
	} else if len(vipList) > 0 {

		vipListUpdated := vipList
		if lib.GetCloudType() == lib.CLOUD_VCENTER || lib.GetCloudType() == lib.CLOUD_NONE {
			segMgmtNetwork := ""
			if lib.GetCloudType() == lib.CLOUD_VCENTER {
				segMgmtNetwork = GetCMSEGManagementNetwork(client)
			}
			vipListUpdated, ret_err = PopulateVipNetworkwithUUID(segMgmtNetwork, client, vipList)
			if len(vipListUpdated) == 0 {
				return false, ret_err
			}
		}
		utils.SetVipNetworkList(vipListUpdated)
		return true, nil
	}

	// 2. AKO created VIP network for AKO in VCF
	if utils.IsVCFCluster() {
		vipNetList := []akov1beta1.AviInfraSettingVipNetwork{
			{
				NetworkName: lib.GetVCFNetworkName(),
			},
		}

		utils.SetVipNetworkList(vipNetList)
		return true, nil
	}

	// 3. Marker based (only advancedL4 - AKO in VDS)
	var err error
	markerNetworkFound := ""
	if utils.GetAdvancedL4() && ipamRefUri != nil {
		// Using clusterID for advl4.
		ipam := models.IPAMDNSProviderProfile{}
		ipamRef := strings.SplitAfter(*ipamRefUri, "/api/")
		ipamRefWithoutName := strings.Split(ipamRef[1], "#")[0]
		ipamURI := "/api/" + ipamRefWithoutName + "/?include_name"
		if err := lib.AviGet(client, ipamURI, &ipam); err != nil {
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

			vipList := []akov1beta1.AviInfraSettingVipNetwork{{
				NetworkName: markerNetworkFound,
			}}
			utils.SetVipNetworkList(vipList)
			return true, nil
		}

	}

	// 4. Empty VipNetworkList
	if utils.IsWCP() && markerNetworkFound == "" {
		utils.SetVipNetworkList([]akov1beta1.AviInfraSettingVipNetwork{})
		return true, nil
	}

	return false, fmt.Errorf("No user input detected for vipNetworkList.")
}

func fetchNetworkWithMarkerSet(client *clients.AviClient, usableNetworkNames []string, overrideUri ...NextPage) (error, string) {
	clusterName := lib.GetClusterID()
	var uri string
	if len(overrideUri) == 1 {
		uri = overrideUri[0].NextURI
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
			nextPage := NextPage{NextURI: overrideUri}
			return fetchNetworkWithMarkerSet(client, usableNetworkNames, nextPage)
		}
	}

	utils.AviLog.Infof("No Marker configured usable networks found.")
	return nil, ""
}

func checkPublicCloud(client *clients.AviClient, returnErr *error) bool {
	if lib.IsPublicCloud() {
		// Handle all public cloud validations here
		vipNetworkList := utils.GetVipNetworkList()
		if len(vipNetworkList) == 0 {
			*returnErr = fmt.Errorf("vipNetworkList not specified, syncing will be disabled")
			return false
		}
	}
	return true
}

func FetchNodeNetworks(segMgmtNetwork string, client *clients.AviClient, returnErr *error, nodeNetworkMap map[string]lib.NodeNetworkMap) bool {
	isVcenterorNoAccessCloud := lib.GetCloudType() == lib.CLOUD_VCENTER || lib.GetCloudType() == lib.CLOUD_NONE
	cmVRFName := lib.AKOControlConfig().ControllerVRFContext()
	for nodeNetworkName, nodeNetworkCIDRs := range nodeNetworkMap {
		localNodeNetworkList := []models.Network{}
		uri := ""

		// cidr validations
		for _, cidr := range nodeNetworkCIDRs.Cidrs {
			_, _, err := net.ParseCIDR(cidr)
			if err != nil {
				*returnErr = fmt.Errorf("value of CIDR couldn't be parsed. Failed with error: %v", err.Error())
				return false
			}
			mask := strings.Split(cidr, "/")[1]
			_, err = strconv.ParseInt(mask, 10, 32)
			if err != nil {
				*returnErr = fmt.Errorf("value of CIDR couldn't be converted to int32")
				return false
			}
		}

		// Following validation is happening double time for Aviinfrasetting side entries.
		if nodeNetworkCIDRs.NetworkUUID != "" {
			// This will change once Aviinfrasetting introduces cloud parameters.
			uri = fmt.Sprintf("/api/network/%s?cloud_uuid=%s&include_name", nodeNetworkCIDRs.NetworkUUID, lib.GetCloudUUID())
			var rest_response interface{}
			err := lib.AviGet(client, uri, &rest_response)
			// here validate response field against vrf for vcenter cloud
			if err != nil {
				*returnErr = fmt.Errorf("no networks with UUID %s found", nodeNetworkCIDRs.NetworkUUID)
				return false
			} else if rest_response == nil {
				*returnErr = fmt.Errorf("no networks with UUID %s found", nodeNetworkCIDRs.NetworkUUID)
				return false
			}
			if isVcenterorNoAccessCloud && cmVRFName != "" {
				uri := fmt.Sprintf("/api/network/%s?cloud_uuid=%s&include_name", nodeNetworkCIDRs.NetworkUUID, lib.GetCloudUUID())
				var rest_response interface{}
				err := lib.AviGet(client, uri, &rest_response)
				if err != nil || rest_response == nil {
					utils.AviLog.Warnf("No networks with UUID %s found", nodeNetworkCIDRs.NetworkUUID)
					*returnErr = fmt.Errorf("no networks with UUID %s found", nodeNetworkCIDRs.NetworkUUID)
					continue
				}
				result := rest_response.(map[string]interface{})
				tempVrf := result["vrf_context_ref"].(string)
				if tempVrf != "" {
					vrf_uuid_name := strings.Split(tempVrf, "#")
					if len(vrf_uuid_name) != 2 || vrf_uuid_name[1] != cmVRFName {
						utils.AviLog.Warnf("Network with UUID %s does not have correct vrf %s", nodeNetworkCIDRs.NetworkUUID, cmVRFName)
						*returnErr = fmt.Errorf("network with UUID %s does not have correct vrf %s", nodeNetworkCIDRs.NetworkUUID, cmVRFName)
						continue
					}
				}
			}
		} else {
			var result session.AviCollectionResult
			var err error
			uri = "/api/network/?include_name&name=" + nodeNetworkName + "&cloud_ref.name=" + utils.CloudName
			if isVcenterorNoAccessCloud && cmVRFName != "" {
				uri = uri + "&vrf_context_ref.name=" + cmVRFName
				result, err = lib.AviGetCollectionRawWithTenantSwitch(client, uri)
			} else {
				result, err = lib.AviGetCollectionRaw(client, uri)
			}
			if err != nil {
				*returnErr = fmt.Errorf("get uri %v returned err %v", uri, err)
				return false
			}
			elems := make([]json.RawMessage, result.Count)
			err = json.Unmarshal(result.Results, &elems)
			if err != nil {
				*returnErr = fmt.Errorf("failed to unmarshal data, err: %s", err.Error())
				return false
			}

			if result.Count == 0 {
				*returnErr = fmt.Errorf("no networks found for networkName: %s", nodeNetworkName)
				return false

			}
			// Only for vcenter when networkUUID is empty, then fetch uuid, for remaining types, use as it is.
			// For no access cloud, as user configures the network so no need to run duplicate network check.
			if lib.GetCloudType() == lib.CLOUD_VCENTER {
				//Fetch all network associated with network name-> This will fetch duplicate networks
				for i := 0; i < result.Count; i++ {
					net := models.Network{}
					if err = json.Unmarshal(elems[i], &net); err != nil {
						utils.AviLog.Warnf("Failed to unmarshal network  data, err: %v", err)
						continue
					}
					localNodeNetworkList = append(localNodeNetworkList, net)
				}
				// if networks count is > 1 find network using overlapping host
				if len(localNodeNetworkList) > 1 {
					// Avoiding cidr match for node network list as nodenetwork cidr can have multiple values without
					// providing type of IP and eah network fetched can have multiple entries.
					// This will create O(n2) loop to find overlap
					nodeNetwork := findHostWithMaxOverlapping(segMgmtNetwork, client, localNodeNetworkList)
					utils.AviLog.Infof("Node network after host overlap call is: %v", utils.Stringify(nodeNetwork))
					nodeNetworkMap[nodeNetworkName] = lib.NodeNetworkMap{
						Cidrs:       nodeNetworkCIDRs.Cidrs,
						NetworkUUID: nodeNetwork.NetworkUUID,
					}
				} else {
					if len(localNodeNetworkList) == 1 {
						nodeNetworkMap[nodeNetworkName] = lib.NodeNetworkMap{
							Cidrs:       nodeNetworkCIDRs.Cidrs,
							NetworkUUID: *localNodeNetworkList[0].UUID,
						}
					}
				}
			}
		}
	}
	return true
}

func checkNodeNetwork(client *clients.AviClient, returnErr *error) bool {
	// Not applicable for non vcenter and nsx-t clouds (overlay)
	if !lib.IsNodeNetworkAllowedCloud() {
		utils.AviLog.Infof("Skipping the check for Node Network ")
		return true
	}

	// check if node network and cidr's are valid
	nodeNetworkMap, err := lib.GetNodeNetworkMapEnv()
	if err != nil {
		*returnErr = fmt.Errorf("fetching node network list failed with error: %s, syncing will be disabled", err.Error())
		return false
	}

	segMgmtNetwork := ""
	// No need to fetch seg mgmt network for no access cloud as AKO doesn't do hostoverlap check
	if lib.CloudType == lib.CLOUD_VCENTER {
		segMgmtNetwork = GetCMSEGManagementNetwork(client)
		utils.AviLog.Infof("SEG Management network is: %v", segMgmtNetwork)
	}
	flag := FetchNodeNetworks(segMgmtNetwork, client, returnErr, nodeNetworkMap)
	utils.AviLog.Infof("NodeNetwork list is: %v", nodeNetworkMap)
	lib.SetNodeNetworkMap(nodeNetworkMap)
	return flag
}
func GetCMSEGManagementNetwork(client *clients.AviClient) string {
	mgmtNetwork := ""
	seg, err := GetAviSeGroup(client, lib.GetSEGName())
	if err == nil {
		// seg MgmtNetwork ref contains network-uuid based url.
		if seg.MgmtNetworkRef != nil {
			parts := strings.Split(*seg.MgmtNetworkRef, "/")
			mgmtNetwork = parts[len(parts)-1]
		}
	}
	return mgmtNetwork
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
	cmVrfContext := lib.AKOControlConfig().ControllerVRFContext()
	if lib.IsNodePortMode() {
		//set it from cm
		if cmVrfContext == "" {
			lib.SetVrf(utils.GlobalVRF)
			utils.AviLog.Infof("Using global VRF for NodePort mode")
		} else {
			lib.SetVrf(cmVrfContext)
			utils.AviLog.Infof("Using %s VRF for NodePort mode", cmVrfContext)
		}
		return true
	}

	// Cluster IP mode: vcenter cloud (write/no access) if vrfContext is in CM use that
	if (lib.GetCloudType() == lib.CLOUD_VCENTER || lib.GetCloudType() == lib.CLOUD_NONE) && cmVrfContext != "" {
		lib.SetVrf(cmVrfContext)
		return true
	}

	// validation of vip networklist with vrf in vcenter cloud is already done in checkRequiredValues function
	networkList := utils.GetVipNetworkList()
	if len(networkList) == 0 {
		utils.AviLog.Warnf("Network name not specified, skipping fetching of the VRF setting from network")
		return true
	}

	if !validateNetworkNames(client, networkList) {
		*returnErr = fmt.Errorf("failed to validate Network Names specified in VIP Network List")
		return false
	}

	network := models.Network{}
	networkName := networkList[0].NetworkName
	if networkList[0].NetworkUUID != "" {
		uri := fmt.Sprintf("/api/network/%s?cloud_uuid=%s&include_name", networkList[0].NetworkUUID, lib.GetCloudUUID())
		var rest_response interface{}
		err := lib.AviGet(client, uri, &rest_response)
		if err != nil {
			*returnErr = fmt.Errorf("no networks found for network: %s", networkList[0].NetworkUUID)
			return false
		} else if rest_response == nil {
			*returnErr = fmt.Errorf("no networks found for network: %s", networkList[0].NetworkUUID)
			return false
		}
		result := rest_response.(map[string]interface{})
		tempVrf := result["vrf_context_ref"].(string)
		network.VrfContextRef = &tempVrf
		networkName = result["name"].(string)
		network.Name = &networkName
	} else {
		uri := "/api/network/?include_name&name=" + networkName + "&cloud_ref.name=" + utils.CloudName
		result, err := lib.AviGetCollectionRaw(client, uri)
		if err != nil {
			*returnErr = fmt.Errorf("get uri %v returned err %v", uri, err)
			return false
		}
		elems := make([]json.RawMessage, result.Count)
		err = json.Unmarshal(result.Results, &elems)
		if err != nil {
			*returnErr = fmt.Errorf("failed to unmarshal data, err: %v", err)
			return false
		}

		if result.Count == 0 {
			*returnErr = fmt.Errorf("no networks found for networkName: %s", networkName)
			return false
		}

		err = json.Unmarshal(elems[0], &network)
		if err != nil {
			*returnErr = fmt.Errorf("failed to unmarshal data, err: %v", err)
			return false
		}
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
		uri = overrideUri[0].NextURI
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
			nextPage := NextPage{NextURI: overrideUri}
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

func validateNetworkNames(client *clients.AviClient, vipNetworkList []akov1beta1.AviInfraSettingVipNetwork) bool {
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

		if vipNetwork.NetworkUUID != "" {
			uri := fmt.Sprintf("/api/network/%s?cloud_uuid=%s&include_name", vipNetwork.NetworkUUID, lib.GetCloudUUID())
			var rest_response interface{}
			err := lib.AviGet(client, uri, &rest_response)
			if err != nil {
				utils.AviLog.Warnf("No networks found for network: %s", vipNetwork.NetworkUUID)
				return false
			} else if rest_response == nil {
				utils.AviLog.Warnf("No networks found for network: %s", vipNetwork.NetworkUUID)
				return false
			}
		} else {
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

func ExtractUUID(word, pattern string) string {
	r, _ := regexp.Compile(pattern)
	result := r.FindAllString(word, -1)
	if len(result) == 1 {
		return result[0][:len(result[0])-1]
	}
	utils.AviLog.Debugf("Uid extraction not successful from: %s, will retry without hash pattern", word)
	return ExtractUUIDWithoutHash(word, pattern[:len(pattern)-1])
}

func ExtractUUIDWithoutHash(word, pattern string) string {
	r, _ := regexp.Compile(pattern)
	result := r.FindAllString(word, -1)
	if len(result) == 1 {
		return result[0][:len(result[0])]
	}
	return ""
}

func getTenantFromTenantRef(tenantRef string) string {
	arr := strings.Split(tenantRef, "#")
	if len(arr) == 2 {
		return arr[1]
	}
	if len(arr) == 1 {
		arr = strings.Split(tenantRef, "/")
		return arr[len(arr)-1]
	}
	return tenantRef
}
