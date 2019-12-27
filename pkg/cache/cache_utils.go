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
	"sync"

	avimodels "github.com/avinetworks/sdk/go/models"
)

type NamespaceName struct {
	Namespace string
	Name      string
}

/*
 * Obj cache
 */

type AviPoolCache struct {
	Name             string
	Tenant           string
	Uuid             string
	CloudConfigCksum string
}

type AviDSCache struct {
	Name   string
	Tenant string
	Uuid   string
}

type AviVsCache struct {
	Name               string
	Tenant             string
	Uuid               string
	Vip                []*avimodels.Vip
	CloudConfigCksum   string
	PGKeyCollection    []NamespaceName
	PoolKeyCollection  []NamespaceName
	DSKeyCollection    []NamespaceName
	SNIChildCollection []string
}

type AviPGCache struct {
	Name             string
	Tenant           string
	Uuid             string
	CloudConfigCksum string
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

func (c *AviCache) AviCacheGetKeyByUuid(uuid string) (interface{}, bool) {
	c.cache_lock.RLock()
	defer c.cache_lock.RUnlock()
	for key, value := range c.cache {
		switch value.(type) {
		case *AviVsCache:
			if value.(*AviVsCache).Uuid == uuid {
				return key, true
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
