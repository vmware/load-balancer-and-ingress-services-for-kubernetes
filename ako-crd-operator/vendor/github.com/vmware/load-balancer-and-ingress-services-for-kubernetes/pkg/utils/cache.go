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

package utils

import (
	"sync"
)

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

/*
 * AviMultiCache provides a one to many cache
 * AviMultiCache for storing objects such as:
 * 1) Service to E/W Pools and Route/Ingress Pools. Of the form:
 * map[{namespace: string, name: string}]map[pool_name_prefix:string]bool
 * 2) Route host name to all Routes with same host name. Of the form:
 * map[host:string]map[{namespace: string, name: string}]bool
 */

type AviMultiCache struct {
	cache_lock sync.RWMutex
	cache      map[interface{}]map[interface{}]bool
}

func NewAviMultiCache() *AviMultiCache {
	c := AviMultiCache{}
	c.cache = make(map[interface{}]map[interface{}]bool)
	return &c
}

func (c *AviMultiCache) AviMultiCacheGetKey(k interface{}) (map[interface{}]bool, bool) {
	c.cache_lock.RLock()
	defer c.cache_lock.RUnlock()
	val, ok := c.cache[k]
	return val, ok
}

func (c *AviMultiCache) AviMultiCacheLookup(k interface{}, lval interface{}) bool {
	c.cache_lock.RLock()
	defer c.cache_lock.RUnlock()
	val, ok := c.cache[k]
	if !ok {
		return ok
	} else {
		_, ok := val[lval]
		return ok
	}
}

func (c *AviMultiCache) AviMultiCacheAdd(k interface{}, val interface{}) {
	c.cache_lock.Lock()
	defer c.cache_lock.Unlock()
	l1val, ok := c.cache[k]
	if ok {
		l1val[val] = true
	} else {
		c.cache[k] = make(map[interface{}]bool)
		c.cache[k][val] = true
	}
}

func (c *AviMultiCache) AviMultiCacheDeleteVal(k interface{}, dval interface{}) {
	c.cache_lock.Lock()
	defer c.cache_lock.Unlock()
	l1val, ok := c.cache[k]
	if ok {
		delete(l1val, dval)
		if len(l1val) == 0 {
			delete(c.cache, k)
		}
	}
}

func (c *AviMultiCache) AviMultiCacheDeleteKey(k interface{}) {
	c.cache_lock.Lock()
	defer c.cache_lock.Unlock()
	delete(c.cache, k)
}
