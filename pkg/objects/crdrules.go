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

package objects

import (
	"sync"
)

var CRDinstance *CRDLister
var crdonce sync.Once

// SharedCRDLister store is used to keep CRD mappings with relevant objects
func SharedCRDLister() *CRDLister {
	crdonce.Do(func() {
		CRDinstance = &CRDLister{
			FqdnHostRuleCache:  NewObjectMapStore(),
			HostRuleFQDNCache:  NewObjectMapStore(),
			FqdnHTTPRulesCache: NewObjectMapStore(),
			HTTPRuleFqdnCache:  NewObjectMapStore(),
		}
	})
	return CRDinstance
}

type CRDLister struct {
	// since the stored values can be from separate namespaces
	// this struct is locked
	NSLock sync.RWMutex

	// fqdn.com: hr1
	FqdnHostRuleCache *ObjectMapStore

	// hr1: fqdn.com - required for httprule
	HostRuleFQDNCache *ObjectMapStore

	// fqdn.com: {path1: rr1, path2: rr1, path3: rr2}
	FqdnHTTPRulesCache *ObjectMapStore

	// rr1: fqdn1.com, rr2: fqdn2.com
	HTTPRuleFqdnCache *ObjectMapStore
}

// FqdnHostRuleCache

func (c *CRDLister) GetFQDNToHostruleMapping(fqdn string) (bool, string) {
	found, hostrule := c.FqdnHostRuleCache.Get(fqdn)
	if !found {
		return false, ""
	}
	return true, hostrule.(string)
}

func (c *CRDLister) GetHostruleToFQDNMapping(hostrule string) (bool, string) {
	found, fqdn := c.HostRuleFQDNCache.Get(hostrule)
	if !found {
		return false, ""
	}
	return true, fqdn.(string)
}

func (c *CRDLister) DeleteHostruleFQDNMapping(hostrule string) bool {
	c.NSLock.Lock()
	defer c.NSLock.Unlock()
	found, fqdn := c.HostRuleFQDNCache.Get(hostrule)
	if found {
		success1 := c.HostRuleFQDNCache.Delete(hostrule)
		success2 := c.FqdnHostRuleCache.Delete(fqdn.(string))
		// utils.AviLog.Infof("Deleted the ingress mappings for hostrule: %s, fqdn: %s", hostrule, fqdn)
		return success1 && success2
	}
	return true
}

func (c *CRDLister) UpdateFQDNHostruleMapping(fqdn string, hostrule string) {
	c.NSLock.Lock()
	defer c.NSLock.Unlock()
	// utils.AviLog.Infof("Updated the Hostrule.fqdn mappings with fqdn: %s, hostrule: %s", fqdn, hostrule)
	c.FqdnHostRuleCache.AddOrUpdate(fqdn, hostrule)
	c.HostRuleFQDNCache.AddOrUpdate(hostrule, fqdn)
}

// FqdnHTTPRulesCache

func (c *CRDLister) GetFqdnHTTPRulesMapping(fqdn string) (bool, map[string]string) {
	found, pathRules := c.FqdnHTTPRulesCache.Get(fqdn)
	if !found {
		return false, make(map[string]string)
	}
	return true, pathRules.(map[string]string)
}

func (c *CRDLister) GetHTTPRuleFqdnMapping(httprule string) (bool, string) {
	found, fqdn := c.HTTPRuleFqdnCache.Get(httprule)
	if !found {
		return false, ""
	}
	return true, fqdn.(string)
}

func (c *CRDLister) RemoveFqdnHTTPRulesMappings(httprule string) bool {
	c.NSLock.Lock()
	defer c.NSLock.Unlock()
	_, fqdn := c.GetHTTPRuleFqdnMapping(httprule)
	success := c.HTTPRuleFqdnCache.Delete(httprule)
	if !success {
		return false
	}

	_, pathRules := c.GetFqdnHTTPRulesMapping(fqdn)
	for path, rule := range pathRules {
		if rule == httprule {
			delete(pathRules, path)
		}
	}
	if len(pathRules) == 0 {
		return c.FqdnHTTPRulesCache.Delete(fqdn)
	}
	c.FqdnHTTPRulesCache.AddOrUpdate(fqdn, pathRules)
	return true
}

func (c *CRDLister) UpdateFqdnHTTPRulesMappings(fqdn, path, httprule string) {
	c.NSLock.Lock()
	defer c.NSLock.Unlock()
	c.HTTPRuleFqdnCache.AddOrUpdate(httprule, fqdn)
	_, pathRules := c.GetFqdnHTTPRulesMapping(fqdn)
	pathRules[path] = httprule
	c.FqdnHTTPRulesCache.AddOrUpdate(fqdn, pathRules)
}
