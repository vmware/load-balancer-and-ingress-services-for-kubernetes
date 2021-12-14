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
	"strings"
	"sync"

	akov1alpha1 "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/apis/ako/v1alpha1"
)

var CRDinstance *CRDLister
var crdonce sync.Once

// SharedCRDLister store is used to keep CRD mappings with relevant objects
func SharedCRDLister() *CRDLister {
	crdonce.Do(func() {
		CRDinstance = &CRDLister{
			FqdnHostRuleCache:      NewObjectMapStore(),
			HostRuleFQDNCache:      NewObjectMapStore(),
			FqdnHTTPRulesCache:     NewObjectMapStore(),
			HTTPRuleFqdnCache:      NewObjectMapStore(),
			FqdnToGSFQDNCache:      NewObjectMapStore(),
			FqdnSharedVSModelCache: NewObjectMapStore(),
			SharedVSModelFqdnCache: NewObjectMapStore(),
			FqdnFqdnTypeCache:      NewObjectMapStore(),
		}
	})
	return CRDinstance
}

type CRDLister struct {
	// since the stored values can be from separate namespaces
	// this struct is locked
	NSLock sync.RWMutex

	// TODO: can be removed once we move to indexers
	// fqdn.com: hr1
	FqdnHostRuleCache *ObjectMapStore

	// hr1: fqdn.com - required for httprule
	HostRuleFQDNCache *ObjectMapStore

	// hr1: gsfqdn.com
	FqdnToGSFQDNCache *ObjectMapStore

	// TODO: can be removed once we move to indexers
	// fqdn.com: {path1: rr1, path2: rr1, path3: rr2}
	FqdnHTTPRulesCache *ObjectMapStore

	// rr1: fqdn1.com, rr2: fqdn2.com
	HTTPRuleFqdnCache *ObjectMapStore

	// shared-vs1-fqdn.com: Shared-VS-L7-1, shared-vs2-fqdn.com: SharedVS-L7-2
	FqdnSharedVSModelCache *ObjectMapStore

	// Shared-VS-L7-1: shared-vs1-fqdn.com, SharedVS-L7-2: shared-vs2-fqdn.com
	SharedVSModelFqdnCache *ObjectMapStore

	// shared-vs1-fqdn: contains, foo.com: exact
	FqdnFqdnTypeCache *ObjectMapStore
}

// FqdnHostRuleCache
func (c *CRDLister) GetFQDNToHostruleMapping(fqdn string) (bool, string) {
	found, hostrule := c.FqdnHostRuleCache.Get(fqdn)
	if !found {
		return false, ""
	}
	return true, hostrule.(string)
}

func (c *CRDLister) GetFQDNToHostruleMappingWithType(fqdn string) (bool, string) {
	// not exact fqdns
	allFqdns := c.FqdnHostRuleCache.GetAllKeys()
	returnHostrules := []string{}
	for _, mFqdn := range allFqdns {
		oktype, fqdnType := c.FqdnFqdnTypeCache.Get(mFqdn)
		if !oktype || fqdnType == "" {
			fqdnType = string(akov1alpha1.Exact)
		}

		if fqdnType == string(akov1alpha1.Exact) && mFqdn == fqdn {
			if found, hostrule := c.FqdnHostRuleCache.Get(mFqdn); found {
				returnHostrules = append(returnHostrules, hostrule.(string))
				break
			}
		} else if fqdnType == string(akov1alpha1.Contains) && strings.Contains(fqdn, mFqdn) {
			if found, hostrule := c.FqdnHostRuleCache.Get(mFqdn); found {
				returnHostrules = append(returnHostrules, hostrule.(string))
				break
			}
		} else if fqdnType == string(akov1alpha1.Wildcard) && strings.HasPrefix(mFqdn, "*") {
			wildcardFqdn := strings.Split(mFqdn, "*")[1]
			if strings.HasSuffix(fqdn, wildcardFqdn) {
				if found, hostrule := c.FqdnHostRuleCache.Get(mFqdn); found {
					returnHostrules = append(returnHostrules, hostrule.(string))
				}
				break
			}
		}
	}

	if len(returnHostrules) > 0 {
		return true, returnHostrules[0]
	}
	return false, ""
}

func (c *CRDLister) GetHostruleToFQDNMapping(hostrule string) (bool, string) {
	found, fqdn := c.HostRuleFQDNCache.Get(hostrule)
	if !found {
		return false, ""
	}
	return true, fqdn.(string)
}

func (c *CRDLister) GetLocalFqdnToGSFQDNMapping(fqdn string) (bool, string) {
	found, gsfqdn := c.FqdnToGSFQDNCache.Get(fqdn)
	if !found {
		return false, ""
	}
	return true, gsfqdn.(string)
}

func (c *CRDLister) DeleteHostruleFQDNMapping(hostrule string) bool {
	c.NSLock.Lock()
	defer c.NSLock.Unlock()
	found, fqdn := c.HostRuleFQDNCache.Get(hostrule)
	if found {
		success1 := c.HostRuleFQDNCache.Delete(hostrule)
		success2 := c.FqdnHostRuleCache.Delete(fqdn.(string))
		return success1 && success2
	}
	return true
}

func (c *CRDLister) DeleteLocalFqdnToGsFqdnMap(fqdn string) bool {
	c.NSLock.Lock()
	defer c.NSLock.Unlock()
	found, _ := c.FqdnToGSFQDNCache.Get(fqdn)
	if found {
		success := c.FqdnToGSFQDNCache.Delete(fqdn)
		return success
	}
	return true
}

func (c *CRDLister) UpdateLocalFQDNToGSFqdnMapping(fqdn string, gsFqdn string) {
	c.NSLock.Lock()
	defer c.NSLock.Unlock()
	c.FqdnToGSFQDNCache.AddOrUpdate(fqdn, gsFqdn)
}

func (c *CRDLister) UpdateFQDNHostruleMapping(fqdn string, hostrule string) {
	c.NSLock.Lock()
	defer c.NSLock.Unlock()
	c.FqdnHostRuleCache.AddOrUpdate(fqdn, hostrule)
	c.HostRuleFQDNCache.AddOrUpdate(hostrule, fqdn)
}

func (c *CRDLister) GetFQDNFQDNTypeMapping(fqdn string) (bool, string) {
	found, fqdnType := c.FqdnFqdnTypeCache.Get(fqdn)
	if !found {
		return false, ""
	}
	return true, fqdnType.(string)
}

func (c *CRDLister) DeleteFQDNFQDNTypeMapping(fqdn string) bool {
	return c.FqdnFqdnTypeCache.Delete(fqdn)
}

func (c *CRDLister) UpdateFQDNFQDNTypeMapping(fqdn, fqdnType string) {
	c.FqdnFqdnTypeCache.AddOrUpdate(fqdn, fqdnType)
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

// FqdnSharedVSModelCache/SharedVSModelFqdnCache
func (c *CRDLister) GetFQDNToSharedVSModelMapping(fqdn string) (bool, []string) {
	oktype, fqdnType := c.FqdnFqdnTypeCache.Get(fqdn)
	if !oktype || fqdnType == "" {
		fqdnType = string(akov1alpha1.Exact)
	}

	allFqdns := c.FqdnSharedVSModelCache.GetAllKeys()
	returnModelNames := []string{}
	for _, mFqdn := range allFqdns {
		if fqdnType == string(akov1alpha1.Exact) && mFqdn == fqdn {
			if found, modelName := c.FqdnSharedVSModelCache.Get(mFqdn); found {
				returnModelNames = append(returnModelNames, modelName.(string))
				break
			}
		} else if fqdnType == string(akov1alpha1.Contains) && strings.Contains(mFqdn, fqdn) {
			if found, modelName := c.FqdnSharedVSModelCache.Get(mFqdn); found {
				returnModelNames = append(returnModelNames, modelName.(string))
			}
		} else if fqdnType == string(akov1alpha1.Wildcard) && strings.HasPrefix(fqdn, "*") {
			wildcardFqdn := strings.Split(fqdn, "*")[1]
			if strings.HasSuffix(mFqdn, wildcardFqdn) {
				if found, modelName := c.FqdnSharedVSModelCache.Get(mFqdn); found {
					returnModelNames = append(returnModelNames, modelName.(string))
				}
			}
		}
	}

	if len(returnModelNames) > 0 {
		return true, returnModelNames
	}
	return false, returnModelNames
}

func (c *CRDLister) GetSharedVSModelFQDNMapping(modelName string) (bool, string) {
	found, fqdn := c.SharedVSModelFqdnCache.Get(modelName)
	if !found {
		return false, ""
	}
	return true, fqdn.(string)
}

func (c *CRDLister) UpdateFQDNSharedVSModelMappings(fqdn, modelName string) {
	c.NSLock.Lock()
	defer c.NSLock.Unlock()
	c.FqdnSharedVSModelCache.AddOrUpdate(fqdn, modelName)
	c.SharedVSModelFqdnCache.AddOrUpdate(modelName, fqdn)
}

func (c *CRDLister) DeleteFQDNSharedVSModelMapping(fqdn string) bool {
	return c.FqdnSharedVSModelCache.Delete(fqdn)
}

func (c *CRDLister) DeleteSharedVSModelFQDNMapping(modelName string) bool {
	return c.SharedVSModelFqdnCache.Delete(modelName)
}
