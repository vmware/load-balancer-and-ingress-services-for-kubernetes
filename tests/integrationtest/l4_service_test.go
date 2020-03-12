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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"ako/pkg/cache"
	"ako/pkg/k8s"
	avinodes "ako/pkg/nodes"
	"ako/pkg/objects"
	"github.com/avinetworks/container-lib/utils"
	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	SINGLEPORTSVC   = "testsvc"
	MULTIPORTSVC    = "testsvcmulti"
	NAMESPACE       = "red-ns"
	AVINAMESPACE    = "admin"
	SINGLEPORTMODEL = "admin/testsvc--red-ns"
	MULTIPORTMODEL  = "admin/testsvcmulti--red-ns"
)

func TestMain(m *testing.M) {
	SetUp()
	os.Exit(m.Run())
}

func returnTestServerMacro() (ts *httptest.Server) {
	ts = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if strings.Contains(r.URL.EscapedPath(), "macro") {
			// POST macro APIs for vs, pg, pool, ds creation on controller

			// copying request payload into response body
			data, _ := ioutil.ReadAll(r.Body)
			var s map[string]interface{}
			_ = json.Unmarshal(data, &s)
			sData, sModelName := s["data"].(map[string]interface{}), strings.ToLower(s["model_name"].(string))
			theURL := "https://localhost/api/" + sModelName + "/" + sModelName + "-random-uuid#" + sData["name"].(string)

			// and adding uuid and url (read-only) fields in the response
			sData["url"] = theURL
			sData["uuid"] = "random-uuid"
			respData := []interface{}{s["data"]}
			out, _ := json.Marshal(respData)
			fmt.Fprintln(w, string(out))
		} else {
			// This is used for /login --> first request to controller
			fmt.Fprintln(w, string(`{"dummy" :"data"}`))
		}
	}))

	url := strings.Split(ts.URL, "https://")[1]
	os.Setenv("CTRL_USERNAME", "admin")
	os.Setenv("CTRL_PASSWORD", "admin")
	os.Setenv("CTRL_IPADDRESS", url)
	return ts
}

func SetUpTestForSvcLB(t *testing.T) {
	objects.SharedAviGraphLister().Delete(SINGLEPORTMODEL)
	CreateSVC(t, NAMESPACE, SINGLEPORTSVC, corev1.ServiceTypeLoadBalancer, false)
	CreateEP(t, NAMESPACE, SINGLEPORTSVC, false, false)
	PollForCompletion(t, SINGLEPORTMODEL, 5)
}

func TearDownTestForSvcLB(t *testing.T) {
	objects.SharedAviGraphLister().Delete(SINGLEPORTMODEL)
	DelSVC(t, NAMESPACE, SINGLEPORTSVC)
	DelEP(t, NAMESPACE, SINGLEPORTSVC)
}

func SetUpTestForSvcLBMultiport(t *testing.T) {
	objects.SharedAviGraphLister().Delete(MULTIPORTMODEL)
	CreateSVC(t, NAMESPACE, MULTIPORTSVC, corev1.ServiceTypeLoadBalancer, true)
	CreateEP(t, NAMESPACE, MULTIPORTSVC, true, true)
	PollForCompletion(t, MULTIPORTMODEL, 10)
}

func TearDownTestForSvcLBMultiport(t *testing.T) {
	objects.SharedAviGraphLister().Delete(MULTIPORTMODEL)
	DelSVC(t, NAMESPACE, MULTIPORTSVC)
	DelEP(t, NAMESPACE, MULTIPORTSVC)
}

func TestAviNodeCreationSinglePort(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	name, namespace, aviNamespace := "testsvc", "red-ns", "admin"
	modelName := SINGLEPORTMODEL

	SetUpTestForSvcLB(t)

	found, aviModel := objects.SharedAviGraphLister().Get(modelName)
	if !found {
		t.Fatalf("Couldn't find model %v", modelName)
	} else {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.Equal(fmt.Sprintf("%s--%s", name, namespace)))
		g.Expect(nodes[0].Tenant).To(gomega.Equal(aviNamespace))
		g.Expect(nodes[0].EastWest).To(gomega.Equal(false))
		g.Expect(nodes[0].PortProto[0].Port).To(gomega.Equal(int32(8080)))

		// Check for the pools
		g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(1))
		address := "1.1.1.1"
		g.Expect(nodes[0].PoolRefs[0].Servers[0].Ip.Addr).To(gomega.Equal(&address))
		g.Expect(len(nodes[0].TCPPoolGroupRefs)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].PoolGroupRefs)).To(gomega.Equal(1))
	}

	TearDownTestForSvcLB(t)
}

func TestAviNodeCreationMultiPort(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	name, namespace, aviNamespace := "testsvcmulti", "red-ns", "admin"
	modelName := fmt.Sprintf("%s/%s--%s", aviNamespace, name, namespace)

	SetUpTestForSvcLBMultiport(t)

	found, aviModel := objects.SharedAviGraphLister().Get(modelName)
	if !found {
		t.Fatalf("Couldn't find model %v", modelName)
	} else {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.Equal(fmt.Sprintf("%s--%s", name, namespace)))
		g.Expect(nodes[0].Tenant).To(gomega.Equal(aviNamespace))
		g.Expect(nodes[0].EastWest).To(gomega.Equal(false))
		g.Expect(nodes[0].PortProto[0].Port).To(gomega.Equal(int32(8080)))

		// Check for the pools
		g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(3))
		for _, node := range nodes[0].PoolRefs {
			if node.Port == 8080 {
				address := "1.1.1.1"
				g.Expect(len(node.Servers)).To(gomega.Equal(3))
				g.Expect(node.Servers[0].Ip.Addr).To(gomega.Equal(&address))
			} else if node.Port == 8081 {
				address := "1.1.1.4"
				g.Expect(len(node.Servers)).To(gomega.Equal(2))
				g.Expect(node.Servers[0].Ip.Addr).To(gomega.Equal(&address))
			} else {
				address := "1.1.1.6"
				g.Expect(len(node.Servers)).To(gomega.Equal(1))
				g.Expect(node.Servers[0].Ip.Addr).To(gomega.Equal(&address))
			}
		}
		g.Expect(len(nodes[0].TCPPoolGroupRefs)).To(gomega.Equal(3))
		g.Expect(len(nodes[0].PoolGroupRefs)).To(gomega.Equal(3))
		g.Expect(nodes[0].ApplicationProfile).To(gomega.Equal(utils.DEFAULT_L4_APP_PROFILE))
		g.Expect(nodes[0].NetworkProfile).To(gomega.Equal(utils.DEFAULT_TCP_NW_PROFILE))
	}

	TearDownTestForSvcLBMultiport(t)
}

func TestAviNodeMultiPortApplicationProf(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	name, namespace, aviNamespace := "testsvcmulti", "red-ns", "admin"
	modelName := fmt.Sprintf("%s/%s--%s", aviNamespace, name, namespace)

	SetUpTestForSvcLBMultiport(t)

	found, aviModel := objects.SharedAviGraphLister().Get(modelName)
	if !found {
		t.Fatalf("Couldn't find model %v", modelName)
	} else {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.Equal(fmt.Sprintf("%s--%s", name, namespace)))
		g.Expect(nodes[0].Tenant).To(gomega.Equal(aviNamespace))
		g.Expect(nodes[0].EastWest).To(gomega.Equal(false))
		g.Expect(nodes[0].PortProto[0].Port).To(gomega.Equal(int32(8080)))

		// Check for the pools
		g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(3))
		for _, node := range nodes[0].PoolRefs {
			if node.Port == 8080 {
				address := "1.1.1.1"
				g.Expect(len(node.Servers)).To(gomega.Equal(3))
				g.Expect(node.Servers[0].Ip.Addr).To(gomega.Equal(&address))
			} else if node.Port == 8081 {
				address := "1.1.1.4"
				g.Expect(len(node.Servers)).To(gomega.Equal(2))
				g.Expect(node.Servers[0].Ip.Addr).To(gomega.Equal(&address))
			} else if node.Port == 8082 {
				address := "1.1.1.6"
				g.Expect(len(node.Servers)).To(gomega.Equal(1))
				g.Expect(node.Servers[0].Ip.Addr).To(gomega.Equal(&address))
			}
		}
		g.Expect(len(nodes[0].TCPPoolGroupRefs)).To(gomega.Equal(3))
		g.Expect(len(nodes[0].PoolGroupRefs)).To(gomega.Equal(3))
		g.Expect(nodes[0].SharedVS).To(gomega.Equal(false))
		g.Expect(nodes[0].ApplicationProfile).To(gomega.Equal(utils.DEFAULT_L4_APP_PROFILE))
		g.Expect(nodes[0].NetworkProfile).To(gomega.Equal(utils.DEFAULT_TCP_NW_PROFILE))
	}

	TearDownTestForSvcLBMultiport(t)
}

func TestAviNodeUpdateEndpoint(t *testing.T) {
	var err error
	g := gomega.NewGomegaWithT(t)
	name, namespace, aviNamespace := "testsvc", "red-ns", "admin"
	modelName := fmt.Sprintf("%s/%s--%s", aviNamespace, name, namespace)

	SetUpTestForSvcLB(t)

	epExample := &corev1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: name},
		Subsets: []corev1.EndpointSubset{{
			Addresses: []corev1.EndpointAddress{{IP: "1.2.3.14"}, {IP: "1.2.3.24"}},
			Ports:     []corev1.EndpointPort{{Name: "foo", Port: 8080, Protocol: "TCP"}},
		}},
	}
	if _, err = KubeClient.CoreV1().Endpoints(namespace).Update(epExample); err != nil {
		t.Fatalf("Error in updating the Endpoint: %v", err)
	}

	var aviModel interface{}
	g.Eventually(func() []avinodes.AviPoolMetaServer {
		_, aviModel = objects.SharedAviGraphLister().Get(modelName)
		pools := aviModel.(*avinodes.AviObjectGraph).GetAviPoolNodes()
		return pools[0].Servers
	}, 5*time.Second).Should(gomega.HaveLen(2))

	found, aviModel := objects.SharedAviGraphLister().Get(modelName)
	if !found {
		t.Fatalf("Couldn't find model %v", modelName)
	} else {
		pools := aviModel.(*avinodes.AviObjectGraph).GetAviPoolNodes()
		for _, pool := range pools {
			if pool.Port == 8080 {
				address := "1.2.3.24"
				g.Expect(pool.Servers).To(gomega.HaveLen(2))
				g.Expect(pool.Servers[1].Ip.Addr).To(gomega.Equal(&address))
			} else {
				g.Expect(pool.Servers).To(gomega.HaveLen(0))
			}
		}
	}

	TearDownTestForSvcLB(t)
}

func TestCreateServiceLB(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	name, namespace, aviNamespace := "testsvc", "red-ns", "admin"

	ts := returnTestServerMacro()
	defer ts.Close()
	k8s.PopulateCache()

	SetUpTestForSvcLB(t)

	mcache := cache.SharedAviObjCache()
	vsKey := cache.NamespaceName{Namespace: aviNamespace, Name: fmt.Sprintf("%s--%s", name, namespace)}
	vsCache, found := mcache.VsCache.AviCacheGet(vsKey)
	if !found {
		t.Fatalf("Cache not found for VS: %v", vsKey)
	} else {
		vsCacheObj, ok := vsCache.(*cache.AviVsCache)
		if !ok {
			t.Fatalf("Invalid VS object. Cannot cast.")
		}
		g.Expect(vsCacheObj.Name).To(gomega.Equal(fmt.Sprintf("%s--%s", name, namespace)))
		g.Expect(vsCacheObj.Tenant).To(gomega.Equal(aviNamespace))
	}

	TearDownTestForSvcLB(t)
}

func TestCreateMultiportServiceLB(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	name, namespace, aviNamespace := "testsvcmulti", "red-ns", "admin"

	ts := returnTestServerMacro()
	defer ts.Close()
	k8s.PopulateCache()

	SetUpTestForSvcLBMultiport(t)

	mcache := cache.SharedAviObjCache()
	vsKey := cache.NamespaceName{Namespace: aviNamespace, Name: fmt.Sprintf("%s--%s", name, namespace)}
	vsCache, found := mcache.VsCache.AviCacheGet(vsKey)
	if !found {
		t.Fatalf("Cache not found for VS: %v", vsKey)
	} else {
		vsCacheObj, ok := vsCache.(*cache.AviVsCache)
		if !ok {
			t.Fatalf("Invalid VS object. Cannot cast.")
		}
		g.Expect(vsCacheObj.Name).To(gomega.Equal(fmt.Sprintf("%s--%s", name, namespace)))
		g.Expect(vsCacheObj.Tenant).To(gomega.Equal(aviNamespace))

		// TODO: Check for the pools
	}

	TearDownTestForSvcLBMultiport(t)
}

func TestUpdateAndDeleteServiceLB(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	var err error
	name, namespace, aviNamespace := "testsvc", "red-ns", "admin"

	ts := returnTestServerMacro()
	defer ts.Close()
	k8s.PopulateCache()

	SetUpTestForSvcLB(t)

	svcExample := &corev1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: name},
		Subsets: []corev1.EndpointSubset{{
			Addresses: []corev1.EndpointAddress{{IP: "1.2.3.24"}, {IP: "1.2.3.14"}},
			Ports:     []corev1.EndpointPort{{Name: "foo", Port: 8080, Protocol: "TCP"}},
		}},
	}
	if _, err = KubeClient.CoreV1().Endpoints(namespace).Update(svcExample); err != nil {
		t.Fatalf("Error in updating the Endpoint: %v", err)
	}

	mcache := cache.SharedAviObjCache()
	vsKey := cache.NamespaceName{Namespace: aviNamespace, Name: fmt.Sprintf("%s--%s", name, namespace)}
	vsCache, found := mcache.VsCache.AviCacheGet(vsKey)
	if !found {
		t.Fatalf("Cache not found for VS: %v", vsKey)
	} else {
		vsCacheObj, ok := vsCache.(*cache.AviVsCache)
		if !ok {
			t.Fatalf("Invalid VS object. Cannot cast.")
		}
		g.Expect(vsCacheObj.Name).To(gomega.Equal(fmt.Sprintf("%s--%s", name, namespace)))
		g.Expect(vsCacheObj.Tenant).To(gomega.Equal(aviNamespace))
	}

	TearDownTestForSvcLB(t)
}
