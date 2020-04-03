/*
* [2013] - [2019] Avi Networks Incorporated
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

package integrationtest

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"ako/pkg/cache"

	"github.com/avinetworks/sdk/go/models"
	"github.com/onsi/gomega"
)

func TestCacheGETOKStatus(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	ts := GetAviControllerFakeAPIServer(FeedMockCollectionData)
	defer ts.Close()

	// Verify the cache.
	cacheobj := cache.SharedAviObjCache()
	vsKey := cache.NamespaceName{Namespace: "admin", Name: "Shard-VS-5"}
	vs_cache, found := cacheobj.VsCache.AviCacheGet(vsKey)

	if !found {
		t.Fatalf("Cache not found for VS: %v", vsKey)
	} else {
		vs_cache_obj, ok := vs_cache.(*cache.AviVsCache)
		if !ok {
			t.Fatalf("Invalid VS object. Cannot cast.")
		}
		g.Expect(len(vs_cache_obj.PoolKeyCollection)).To(gomega.Equal(3))
		g.Expect(len(vs_cache_obj.PGKeyCollection)).To(gomega.Equal(1))
		g.Expect(len(vs_cache_obj.DSKeyCollection)).To(gomega.Equal(1))
	}

	// vs with no created_by field should not be present in cache
	vsKey = cache.NamespaceName{Namespace: "admin", Name: "dns-vs"}
	vs_cache, found = cacheobj.VsCache.AviCacheGet(vsKey)
	if found {
		t.Fatalf("Unexpected Cache found for VS: %v", vsKey)
	}

	vrfName := "global"
	vrfCache, found := cacheobj.VrfCache.AviCacheGet(vrfName)
	if !found {
		t.Fatalf("Cache not found for Vrf: %v", vrfName)
	}
	vrfCacheObj, ok := vrfCache.(*cache.AviVrfCache)
	if !ok {
		t.Fatalf("Invalid VS object. Cannot cast.")
	}

	var staticRoutes []*models.StaticRoute
	nodeAddr := "10.52.2.23"
	prefixAddr := "10.244.0.0"
	mask := int32(24)
	routeID := "1"
	staticRoute := GetStaticRoute(nodeAddr, prefixAddr, routeID, mask)
	staticRoutes = append(staticRoutes, staticRoute)

	chksum := cache.VrfChecksum("global", staticRoutes)
	g.Expect(vrfCacheObj.CloudConfigCksum).To(gomega.Equal(chksum))
}

func TestCacheGETControllerUnavailable(t *testing.T) {
	ctrlUnavail := true
	ts := GetAviControllerFakeAPIServer(func(w http.ResponseWriter, r *http.Request) {
		if ctrlUnavail {
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprintln(w, string(`{"error": "Unserviceable"}`))
			ctrlUnavail = false
		}
		FeedMockCollectionData(w, r)
	})
	defer ts.Close()

	// Verify the cache.
	cacheobj := cache.SharedAviObjCache()
	vsKey := cache.NamespaceName{Namespace: "admin", Name: "Shard-VS-5"}
	_, found := cacheobj.VsCache.AviCacheGet(vsKey)
	if !found {
		// The older cache member should be available.
		t.Fatalf("Cache not found for VS: %v", vsKey)
	}
	vsKey = cache.NamespaceName{Namespace: "admin", Name: "Shard-VS-4"}
	_, found = cacheobj.VsCache.AviCacheGet(vsKey)
	if found {
		// The older cache member should be available.
		t.Fatalf("Cache found for VS: %v", vsKey)
	}

	vrfName := "global"
	_, found = cacheobj.VrfCache.AviCacheGet(vrfName)
	if !found {
		t.Fatalf("Cache not found for Vrf: %v", vrfName)
	}
}

func TestCacheGETDependentObjectUnavailable(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	// Verify the state of the cache

	cacheobj := cache.SharedAviObjCache()
	vsKey := cache.NamespaceName{Namespace: "admin", Name: "Shard-VS-5"}
	vs_cache, found := cacheobj.VsCache.AviCacheGet(vsKey)
	if !found {
		// The older cache member should be available.
		t.Fatalf("Cache not found for VS: %v", vsKey)
	} else {
		vs_cache_obj, ok := vs_cache.(*cache.AviVsCache)
		if !ok {
			t.Fatalf("Invalid VS object. Cannot cast.")
		}
		g.Expect(len(vs_cache_obj.PoolKeyCollection)).To(gomega.Equal(3))
		g.Expect(len(vs_cache_obj.PGKeyCollection)).To(gomega.Equal(1))
		g.Expect(len(vs_cache_obj.DSKeyCollection)).To(gomega.Equal(1))
	}

	ts := GetAviControllerFakeAPIServer(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.EscapedPath(), "poolgroup") {
			w.WriteHeader(http.StatusInternalServerError)
			return
		} else if strings.Contains(r.URL.EscapedPath(), "pool") {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}

		FeedMockCollectionData(w, r)
	})
	defer ts.Close()

	// Verify the cache.
	vs_cache, found = cacheobj.VsCache.AviCacheGet(vsKey)
	if !found {
		// The older cache member should be available.
		t.Fatalf("Cache not found for VS: %v", vsKey)
	} else {
		vs_cache_obj, ok := vs_cache.(*cache.AviVsCache)
		if !ok {
			t.Fatalf("Invalid VS object. Cannot cast.")
		}
		g.Expect(len(vs_cache_obj.PoolKeyCollection)).To(gomega.Equal(3))
		// The PG had a problem in GET operation, but we will retain the cache.
		g.Expect(len(vs_cache_obj.PGKeyCollection)).To(gomega.Equal(1))
		g.Expect(len(vs_cache_obj.DSKeyCollection)).To(gomega.Equal(1))
	}
	vsKey = cache.NamespaceName{Namespace: "admin", Name: "Shard-VS-4"}
	_, found = cacheobj.VsCache.AviCacheGet(vsKey)
	if found {
		t.Fatalf("Cache found for VS: %v", vsKey)
	}
}
