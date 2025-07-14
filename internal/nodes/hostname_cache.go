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

package nodes

import (
	"strings"
	"sync"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"
	akov1beta1 "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/apis/ako/v1beta1"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
)

var hostNameListerInstance *HostNameLister
var hsonce sync.Once

func SharedHostNameLister() *HostNameLister {
	hsonce.Do(func() {
		hostNameListerInstance = &HostNameLister{
			secureHostNameStore: objects.NewObjectMapStore(),
			namespaceStore:      objects.NewObjectMapStore(),
			HostNamePathStore: HostNamePathStore{
				hostNamePathStore: objects.NewObjectMapStore(),
			},
		}
	})
	return hostNameListerInstance
}

type HostNameLister struct {
	secureHostNameStore *objects.ObjectMapStore
	namespaceStore      *objects.ObjectMapStore
	HostNamePathStore
}

func (a *HostNameLister) Save(hostname string, hsGraph SecureHostNameMapProp) {
	utils.AviLog.Infof("Saving hostname map :%s", hostname)
	a.secureHostNameStore.AddOrUpdate(hostname, hsGraph)
}

func (a *HostNameLister) Get(hostname string) (bool, SecureHostNameMapProp) {
	ok, obj := a.secureHostNameStore.Get(hostname)
	if !ok {
		return ok, SecureHostNameMapProp{}
	}
	return ok, obj.(SecureHostNameMapProp)
}

func (a *HostNameLister) Delete(hostname string) {
	a.secureHostNameStore.Delete(hostname)
}

func (a *HostNameLister) SaveNamespace(hostname string, namespace string) {
	a.namespaceStore.AddOrUpdate(hostname, namespace)
}

func (a *HostNameLister) GetNamespace(hostname string) (bool, string) {
	found, obj := a.namespaceStore.Get(hostname)
	if !found {
		return false, ""
	}

	ns, ok := obj.(string)
	if !ok {
		utils.AviLog.Warnf("Wrong object type for namespace for the hostname %s: T", hostname, obj)
		return false, ""
	}
	return true, ns
}

func (a *HostNameLister) DeleteNamespace(hostname string) {
	a.namespaceStore.Delete(hostname)
}

// thread safe for namespace based sharding in case of same hostname in different namespaces
// cache sample: foo.com -> {path1: [ns1/ingress1], path2: [ns2/ingress2]}
type HostNamePathStore struct {
	sync.RWMutex
	hostNamePathStore *objects.ObjectMapStore
}

func (h *HostNamePathStore) GetHostPathStore(host string) (bool, map[string][]string) {
	ok, obj := h.hostNamePathStore.Get(host)
	if !ok {
		return false, make(map[string][]string)
	}
	return true, obj.(map[string][]string)
}

func (h *HostNamePathStore) GetHostsFromHostPathStore(host, fqdnMatchType string) []string {
	allHosts := h.hostNamePathStore.GetAllKeys()
	returnHosts := []string{}

	for _, mHost := range allHosts {
		if fqdnMatchType == string(akov1beta1.Exact) && mHost == host {
			returnHosts = append(returnHosts, mHost)
		} else if fqdnMatchType == string(akov1beta1.Contains) && strings.Contains(host, mHost) {
			returnHosts = append(returnHosts, mHost)
		} else if fqdnMatchType == string(akov1beta1.Wildcard) && strings.HasPrefix(host, "*") {
			wildcardFqdn := strings.Split(host, "*")[1]
			if strings.HasSuffix(mHost, wildcardFqdn) {
				returnHosts = append(returnHosts, mHost)
			}
		}
	}
	return returnHosts
}

func (h *HostNamePathStore) GetHostPathStoreIngresses(host, path string) (bool, []string) {
	ok, obj := h.hostNamePathStore.Get(host)
	if !ok {
		return false, []string{}
	}
	mmap := obj.(map[string][]string)
	if _, ok := mmap[path]; !ok {
		return false, []string{}
	}
	return true, mmap[path]
}

func (h *HostNamePathStore) SaveHostPathStore(host, path string, ing string) {
	h.Lock()
	defer h.Unlock()
	found, pathings := h.GetHostPathStore(host)
	if found {
		if _, ok := pathings[path]; ok && !utils.HasElem(pathings[path], ing) {
			pathings[path] = append(pathings[path], ing)
		} else {
			pathings[path] = []string{ing}
		}
	} else {
		pathings = make(map[string][]string)
		pathings[path] = []string{ing}
	}

	h.hostNamePathStore.AddOrUpdate(host, pathings)
}

func (h *HostNamePathStore) RemoveHostPathStore(host, path string, ing string) {
	h.Lock()
	defer h.Unlock()
	found, pathings := h.GetHostPathStore(host)
	if found {
		if _, ok := pathings[path]; ok && utils.HasElem(pathings[path], ing) {
			pathings[path] = utils.Remove(pathings[path], ing)
			if len(pathings[path]) == 0 {
				delete(pathings, path)
			}
			h.hostNamePathStore.AddOrUpdate(host, pathings)
		}
	}

	if len(pathings) == 0 {
		h.hostNamePathStore.Delete(host)
	}
}

func (h *HostNamePathStore) DeleteHostPathStore(host string) {
	h.hostNamePathStore.Delete(host)
}

// added multiple layer of validation. Useful in multihost ingress
func PopulateIngHostMap(namespace, hostName, ingName, secretName string, pathsvcMap HostMetadata) {
	hostMap := HostNamePathSecrets{paths: getPaths(pathsvcMap.ingressHPSvc), secretName: secretName}
	found, ingressHostMap := SharedHostNameLister().Get(hostName)
	if found {
		ingressHostMap.HostNameMap[namespace+"/"+ingName] = hostMap
	} else {
		// Create the map
		ingressHostMap = NewSecureHostNameMapProp()
		ingressHostMap.HostNameMap[namespace+"/"+ingName] = hostMap
	}
	SharedHostNameLister().Save(hostName, ingressHostMap)
}

func EnqueueIng(key, namespace, hostName, ingName string) bool {
	if lib.AKOControlConfig().GetAKOFQDNReusePolicy() != lib.FQDNReusePolicyStrict {
		return true
	}
	return !hostnameExistInDifferentNamespace(key, hostName, namespace)
}

func hostnameExistInDifferentNamespace(key, hostName, namespace string) bool {
	found, ingressHostMap := SharedHostNameLister().Get(hostName)
	if found {
		ingresses := ingressHostMap.GetIngressesForHostName()
		for _, ingress := range ingresses {
			nsIngressName := strings.Split(ingress, "/")
			if len(nsIngressName) > 0 {
				if nsIngressName[0] != namespace {
					utils.AviLog.Debugf("key: %s, msg: Host %s already claimed. Not processing ingress.", key, hostName)
					return true
				}
			}
		}
	}
	return false
}
