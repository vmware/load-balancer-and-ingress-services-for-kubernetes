/*
 * Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
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
	"reflect"
	"sync"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
)

type NamespaceName struct {
	Namespace string
	Name      string
}

/*
 * Obj cache
 */

type AviPoolCache struct {
	Name                 string
	Tenant               string
	Uuid                 string
	CloudConfigCksum     string
	ServiceMetadataObj   lib.ServiceMetadataObj
	PkiProfileCollection NamespaceName
	PersistenceProfile   NamespaceName
	LastModified         string
	InvalidData          bool
	HasReference         bool
}

type AviDSCache struct {
	Name             string
	Tenant           string
	Uuid             string
	PoolGroups       []string
	LastModified     string
	InvalidData      bool
	CloudConfigCksum uint32
	HasReference     bool
}

type AviCloudPropertyCache struct {
	Name      string
	VType     string
	IPAMType  string
	NSIpamDNS []string
}

type AviClusterRuntimeCache struct {
	Name    string
	UpSince string
}

type AviVsCache struct {
	Name                     string
	Tenant                   string
	Uuid                     string
	CloudConfigCksum         string
	PGKeyCollection          []NamespaceName
	VSVipKeyCollection       []NamespaceName
	PoolKeyCollection        []NamespaceName
	DSKeyCollection          []NamespaceName
	HTTPKeyCollection        []NamespaceName
	SSLKeyCertCollection     []NamespaceName
	L4PolicyCollection       []NamespaceName
	SNIChildCollection       []string
	ParentVSRef              NamespaceName
	PassthroughParentRef     NamespaceName
	PassthroughChildRef      NamespaceName
	ServiceMetadataObj       lib.ServiceMetadataObj
	LastModified             string
	EnableRhi                bool
	InvalidData              bool
	VSCacheLock              sync.RWMutex
	StringGroupKeyCollection []NamespaceName
}

func (c *AviCache) AviCacheAddVS(k NamespaceName) *AviVsCache {
	c.cache_lock.Lock()
	defer c.cache_lock.Unlock()
	val, found := c.cache[k]
	if found {
		aviVS, ok := val.(*AviVsCache)
		if ok {
			return aviVS
		}
	}
	vsObj := AviVsCache{Name: k.Name, Tenant: k.Namespace}
	c.cache[k] = &vsObj
	return &vsObj
}

func (c *AviCache) AviCacheAddPool(k NamespaceName) *AviPoolCache {
	c.cache_lock.Lock()
	defer c.cache_lock.Unlock()
	val, found := c.cache[k]
	if found {
		aviPool, ok := val.(*AviPoolCache)
		if ok {
			return aviPool
		}
	}
	poolObj := AviPoolCache{Name: k.Name, Tenant: k.Namespace}
	c.cache[k] = &poolObj
	return &poolObj
}

func (v *AviVsCache) SetPGKeyCollection(keyCollection []NamespaceName) {
	v.VSCacheLock.Lock()
	defer v.VSCacheLock.Unlock()
	v.PGKeyCollection = keyCollection
}

func RemoveNamespaceName(s []NamespaceName, r NamespaceName) []NamespaceName {
	n := 0
	for _, v := range s {
		if v != r {
			s[n] = v
			n++
		}
	}
	return s[:n]
}

func (v *AviVsCache) AddToPGKeyCollection(k NamespaceName) {
	if v.PGKeyCollection == nil {
		v.PGKeyCollection = []NamespaceName{k}
	}
	if !utils.HasElem(v.PGKeyCollection, k) {
		v.PGKeyCollection = append(v.PGKeyCollection, k)
	}
}

func (v *AviVsCache) RemoveFromPGKeyCollection(k NamespaceName) {
	if v.PGKeyCollection == nil {
		return
	}
	v.PGKeyCollection = RemoveNamespaceName(v.PGKeyCollection, k)
}

func (v *AviVsCache) AddToVSVipKeyCollection(k NamespaceName) {
	if v.VSVipKeyCollection == nil {
		v.VSVipKeyCollection = []NamespaceName{k}
	}
	if !utils.HasElem(v.VSVipKeyCollection, k) {
		v.VSVipKeyCollection = append(v.VSVipKeyCollection, k)
	}
}

func (v *AviVsCache) RemoveFromVSVipKeyCollection(k NamespaceName) {
	if v.VSVipKeyCollection == nil {
		return
	}
	v.VSVipKeyCollection = RemoveNamespaceName(v.VSVipKeyCollection, k)
}

func (v *AviVsCache) AddToPoolKeyCollection(k NamespaceName) {
	if v.PoolKeyCollection == nil {
		v.PoolKeyCollection = []NamespaceName{k}
		return
	}
	if !utils.HasElem(v.PoolKeyCollection, k) {
		v.PoolKeyCollection = append(v.PoolKeyCollection, k)
	}
}

func (v *AviVsCache) RemoveFromPoolKeyCollection(k NamespaceName) {
	if v.PoolKeyCollection == nil {
		return
	}
	v.PoolKeyCollection = RemoveNamespaceName(v.PoolKeyCollection, k)
}

func (v *AviVsCache) AddToDSKeyCollection(k NamespaceName) {
	if v.DSKeyCollection == nil {
		v.DSKeyCollection = []NamespaceName{k}
	}
	if !utils.HasElem(v.DSKeyCollection, k) {
		v.DSKeyCollection = append(v.DSKeyCollection, k)
	}
}

func (v *AviVsCache) RemoveFromDSKeyCollection(k NamespaceName) {
	if v.DSKeyCollection == nil {
		return
	}
	v.DSKeyCollection = RemoveNamespaceName(v.DSKeyCollection, k)
}

func (v *AviVsCache) AddToHTTPKeyCollection(k NamespaceName) {
	if v.HTTPKeyCollection == nil {
		v.HTTPKeyCollection = []NamespaceName{k}
	}
	if !utils.HasElem(v.HTTPKeyCollection, k) {
		v.HTTPKeyCollection = append(v.HTTPKeyCollection, k)
	}
}

func (v *AviVsCache) RemoveFromHTTPKeyCollection(k NamespaceName) {
	if v.HTTPKeyCollection == nil {
		return
	}
	v.HTTPKeyCollection = RemoveNamespaceName(v.HTTPKeyCollection, k)
}

func (v *AviVsCache) AddToStringGroupKeyCollection(k NamespaceName) {
	if v.StringGroupKeyCollection == nil {
		v.StringGroupKeyCollection = []NamespaceName{k}
	}
	if !utils.HasElem(v.StringGroupKeyCollection, k) {
		v.StringGroupKeyCollection = append(v.StringGroupKeyCollection, k)
	}
}

func (v *AviVsCache) RemoveFromStringGroupKeyCollection(k NamespaceName) {
	if v.StringGroupKeyCollection == nil {
		return
	}
	v.StringGroupKeyCollection = RemoveNamespaceName(v.StringGroupKeyCollection, k)
}

func (v *AviVsCache) AddToSSLKeyCertCollection(k NamespaceName) {
	if v.SSLKeyCertCollection == nil {
		v.SSLKeyCertCollection = []NamespaceName{k}
	}
	if !utils.HasElem(v.SSLKeyCertCollection, k) {
		v.SSLKeyCertCollection = append(v.SSLKeyCertCollection, k)
	}
}

func (v *AviVsCache) RemoveFromSSLKeyCertCollection(k NamespaceName) {
	if v.SSLKeyCertCollection == nil {
		return
	}
	v.SSLKeyCertCollection = RemoveNamespaceName(v.SSLKeyCertCollection, k)
}

func (v *AviVsCache) AddToL4PolicyCollection(k NamespaceName) {
	if v.L4PolicyCollection == nil {
		v.L4PolicyCollection = []NamespaceName{k}
	}
	if !utils.HasElem(v.L4PolicyCollection, k) {
		v.L4PolicyCollection = append(v.L4PolicyCollection, k)
	}
}

func (v *AviVsCache) RemoveFromL4PolicyCollection(k NamespaceName) {
	if v.L4PolicyCollection == nil {
		return
	}
	v.L4PolicyCollection = RemoveNamespaceName(v.L4PolicyCollection, k)
}

func (v *AviVsCache) AddToSNIChildCollection(k string) {
	if v.SNIChildCollection == nil {
		v.SNIChildCollection = []string{k}
	}
	if !utils.HasElem(v.SNIChildCollection, k) {
		v.SNIChildCollection = append(v.SNIChildCollection, k)
	}
}

func (v *AviVsCache) ReplaceSNIChildCollection(k []string) {
	v.SNIChildCollection = k
}

func (v *AviVsCache) RemoveFromSNIChildCollection(k string) {
	if v.SNIChildCollection == nil {
		return
	}
	v.SNIChildCollection = utils.Remove(v.SNIChildCollection, k)
}

type AviSSLCache struct {
	Name             string
	Tenant           string
	Uuid             string
	CloudConfigCksum uint32
	LastModified     string
	InvalidData      bool
	Cert             string
	HasCARef         bool
	CACertUUID       string
	HasReference     bool
}

type AviPkiProfileCache struct {
	Name             string
	Tenant           string
	Uuid             string
	CloudConfigCksum uint32
	LastModified     string
	InvalidData      bool
	HasReference     bool
}

type AviPersistenceProfileCache struct {
	Name             string
	Tenant           string
	Uuid             string
	CloudConfigCksum uint32
	LastModified     string
	Type             string
	InvalidData      bool
}

type NextPage struct {
	NextURI    string
	Collection interface{}
}

type AviPGCache struct {
	Name             string
	Tenant           string
	Uuid             string
	Members          []string // Collection of pools referred by this PG.
	CloudConfigCksum string
	LastModified     string
	InvalidData      bool
	HasReference     bool
}

type AviVSVIPCache struct {
	Name             string
	Tenant           string
	Uuid             string
	FQDNs            []string
	CloudConfigCksum string
	LastModified     string
	InvalidData      bool
	V6IPs            []string
	Vips             []string
	Fips             []string
	NetworkNames     []string
	HasReference     bool
}

type AviHTTPPolicyCache struct {
	Name             string
	Tenant           string
	Uuid             string
	CloudConfigCksum string
	PoolGroups       []string
	Pools            []string
	LastModified     string
	InvalidData      bool
	HasReference     bool
	StringGroupRefs  []string
}

type AviL4PolicyCache struct {
	Name             string
	Tenant           string
	Uuid             string
	CloudConfigCksum uint32
	Pools            []string
	LastModified     string
	HasReference     bool
}

type AviVrfCache struct {
	Name             string
	Uuid             string
	CloudConfigCksum uint32
}

type AviStringGroupCache struct {
	Name             string
	Tenant           string
	Uuid             string
	LastModified     string
	InvalidData      bool
	CloudConfigCksum uint32
	HasReference     bool
	Description      string
	LongestMatch     bool
}

func (v *AviVsCache) GetVSCopy() (*AviVsCache, bool) {
	v.VSCacheLock.RLock()
	defer v.VSCacheLock.RUnlock()
	newObj := AviVsCache{}
	bytes, err := json.Marshal(v)
	if err != nil {
		utils.AviLog.Errorf("key: %s, Unable to marshal: %s", err)
		return nil, false
	}
	err = json.Unmarshal(bytes, &newObj)
	if err != nil {
		utils.AviLog.Errorf("key: %s, Unable to Unmarshal src: %s", err)
		return nil, false
	}
	return &newObj, true
}

/*
 * AviCache provides a one to one cache
 * AviCache for storing objects such as:
 * VirtualServices, PoolGroups, Pools, etc.
 */

type AviCache struct {
	cache_lock sync.RWMutex
	cache      map[interface{}]interface{}
}

func NewAviCache() *AviCache {
	c := AviCache{}
	c.cache = make(map[interface{}]interface{})
	return &c
}

func (c *AviCache) AviCacheGet(k interface{}) (interface{}, bool) {
	c.cache_lock.RLock()
	defer c.cache_lock.RUnlock()
	val, ok := c.cache[k]
	return val, ok
}

func (c *AviCache) AviCacheGetAllParentVSKeys() []NamespaceName {
	c.cache_lock.RLock()
	defer c.cache_lock.RUnlock()
	var keys []NamespaceName
	for k, val := range c.cache {
		vsCache := val.(*AviVsCache)
		if vsCache.ParentVSRef == (NamespaceName{}) && vsCache.ServiceMetadataObj.PassthroughParentRef == "" {
			keys = append(keys, k.(NamespaceName))
		}
	}
	return keys
}

func (c *AviCache) AviCacheGetAllChildVSForParent(parentVsKey NamespaceName) []string {
	c.cache_lock.RLock()
	defer c.cache_lock.RUnlock()
	var uuids []string
	for _, val := range c.cache {
		if val.(*AviVsCache).ParentVSRef == parentVsKey {
			uuids = append(uuids, val.(*AviVsCache).Uuid)
		}
	}
	return uuids
}

func (c *AviCache) AviGetAllKeys() []NamespaceName {
	c.cache_lock.RLock()
	defer c.cache_lock.RUnlock()
	var keys []NamespaceName
	for key := range c.cache {
		keys = append(keys, key.(NamespaceName))
	}
	return keys
}

func (c *AviCache) AviCacheGetKeyByUuid(uuid string) (interface{}, bool) {
	c.cache_lock.RLock()
	defer c.cache_lock.RUnlock()
	for key, value := range c.cache {
		switch value.(type) {
		case *AviVsCache:
			if value.(*AviVsCache) == nil {
				utils.AviLog.Warnf("Got nil value in cache for vs key %v", reflect.ValueOf(key))
			} else if value.(*AviVsCache).Uuid == uuid {
				return key, true
			}
		case *AviVSVIPCache:
			if value.(*AviVSVIPCache) == nil {
				utils.AviLog.Warnf("Got nil value in cache for vsvip key %v", reflect.ValueOf(key))
			} else if value.(*AviVSVIPCache).Uuid == uuid {
				return key, true
			}
		}
	}
	return nil, false
}

func (c *AviCache) AviCacheGetNameByUuid(uuid string) (interface{}, bool) {
	c.cache_lock.RLock()
	defer c.cache_lock.RUnlock()
	for key, value := range c.cache {
		switch value.(type) {
		case *AviPoolCache:
			if value.(*AviPoolCache) == nil {
				utils.AviLog.Warnf("Got nil value in cache for pool key %v", reflect.ValueOf(key))
			} else if value.(*AviPoolCache).Uuid == uuid {
				return value.(*AviPoolCache).Name, true
			}
		case *AviVSVIPCache:
			if value.(*AviVSVIPCache) == nil {
				utils.AviLog.Warnf("Got nil value in cache for vsvip key %v", reflect.ValueOf(key))
			} else if value.(*AviVSVIPCache).Uuid == uuid {
				return value.(*AviVSVIPCache).Name, true
			}
		case *AviSSLCache:
			if value.(*AviSSLCache) == nil {
				utils.AviLog.Warnf("Got nil value in cache for ssl key %v", reflect.ValueOf(key))
			} else if value.(*AviSSLCache).Uuid == uuid {
				return value.(*AviSSLCache).Name, true
			}
		case *AviDSCache:
			if value.(*AviDSCache) == nil {
				utils.AviLog.Warnf("Got nil value in cache for DS key %v", reflect.ValueOf(key))
			} else if value.(*AviDSCache).Uuid == uuid {
				return value.(*AviDSCache).Name, true
			}
		case *AviL4PolicyCache:
			if value.(*AviL4PolicyCache) == nil {
				utils.AviLog.Warnf("Got nil value in cache for l4 policy key %v", reflect.ValueOf(key))
			} else if value.(*AviL4PolicyCache).Uuid == uuid {
				return value.(*AviL4PolicyCache).Name, true
			}
		case *AviHTTPPolicyCache:
			if value.(*AviHTTPPolicyCache) == nil {
				utils.AviLog.Warnf("Got nil value in cache for http policy key %v", reflect.ValueOf(key))
			} else if value.(*AviHTTPPolicyCache).Uuid == uuid {
				return value.(*AviHTTPPolicyCache).Name, true
			}
		case *AviPGCache:
			if value.(*AviPGCache) == nil {
				utils.AviLog.Warnf("Got nil value in cache for PG key %v", reflect.ValueOf(key))
			} else if value.(*AviPGCache).Uuid == uuid {
				return value.(*AviPGCache).Name, true
			}
		case *AviPkiProfileCache:
			if value.(*AviPkiProfileCache) == nil {
				utils.AviLog.Warnf("Got nil value in cache for pki profile key %v", reflect.ValueOf(key))
			} else if value.(*AviPkiProfileCache).Uuid == uuid {
				return value.(*AviPkiProfileCache).Name, true
			}
		case *AviStringGroupCache:
			if value.(*AviStringGroupCache) == nil {
				utils.AviLog.Warnf("Got nil value in cache for stringgroup key %v", reflect.ValueOf(key))
			} else if value.(*AviStringGroupCache).Uuid == uuid {
				return value.(*AviStringGroupCache).Name, true
			}
		case *AviPersistenceProfileCache:
			if value.(*AviPersistenceProfileCache) == nil {
				utils.AviLog.Warnf("Got nil value in cache for persistence profile key %v", reflect.ValueOf(key))
			} else if value.(*AviPersistenceProfileCache).Uuid == uuid {
				return value.(*AviPersistenceProfileCache).Name, true
			}
		}
	}
	return nil, false
}

func (c *AviCache) AviCacheAdd(k interface{}, val interface{}) {
	c.cache_lock.Lock()
	defer c.cache_lock.Unlock()
	c.cache[k] = val
}

func (c *AviCache) AviCacheDelete(k interface{}) {
	c.cache_lock.Lock()
	defer c.cache_lock.Unlock()
	delete(c.cache, k)
}

func (c *AviCache) ShallowCopy() map[interface{}]interface{} {
	// Shallow copy, does not dereference the pointers.
	c.cache_lock.Lock()
	defer c.cache_lock.Unlock()
	newMap := make(map[interface{}]interface{})
	for key, value := range c.cache {
		newMap[key] = value
	}
	return newMap
}
