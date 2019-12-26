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
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"net/http"
	"net/http/httptest"

	"github.com/onsi/gomega"
	"gitlab.eng.vmware.com/orion/akc/pkg/cache"
	"gitlab.eng.vmware.com/orion/akc/pkg/k8s"
)

func TestCacheGETOKStatus(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if strings.Contains(r.URL.EscapedPath(), "virtualservice") {
			data, _ := ioutil.ReadFile("avimockobjects/shared_vs_mock.json")

			fmt.Fprintln(w, string(data))
		} else if strings.Contains(r.URL.EscapedPath(), "poolgroup") {
			data, _ := ioutil.ReadFile("avimockobjects/poolgroups_mock.json")

			fmt.Fprintln(w, string(data))
		} else if strings.Contains(r.URL.EscapedPath(), "pool") {
			data, _ := ioutil.ReadFile("avimockobjects/pool_mock.json")

			fmt.Fprintln(w, string(data))
		} else if strings.Contains(r.URL.EscapedPath(), "vsdatascript") {
			data, _ := ioutil.ReadFile("avimockobjects/datascript_http_mock.json")
			fmt.Fprintln(w, string(data))
		} else {
			// This is used for /login --> first request to controller
			fmt.Fprintln(w, string(`{"dummy" :"data"}`))
		}

	}))
	defer ts.Close()
	url := strings.Split(ts.URL, "https://")[1]
	fmt.Println(url)
	os.Setenv("CTRL_USERNAME", "admin")
	os.Setenv("CTRL_PASSWORD", "admin")
	os.Setenv("CTRL_IPADDRESS", url)
	k8s.PopulateCache()
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
}
