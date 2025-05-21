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

package evhtests

import (
	"context"
	"sort"
	"testing"
	"time"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/cache"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	avinodes "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/nodes"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"
	utils "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/tests/integrationtest"

	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func VerifyEvhPoolDeletion(t *testing.T, g *gomega.WithT, aviModel interface{}, poolCount int) {
	var nodes []*avinodes.AviEvhVsNode
	g.Eventually(func() []*avinodes.AviPoolNode {
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return nodes[0].PoolRefs
	}, 50*time.Second).Should(gomega.HaveLen(poolCount))
}

func VerifyEvhIngressDeletion(t *testing.T, g *gomega.WithT, aviModel interface{}, evhCount int) {
	var nodes []*avinodes.AviEvhVsNode
	g.Eventually(func() []*avinodes.AviEvhVsNode {
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return nodes[0].EvhNodes
	}, 50*time.Second).Should(gomega.HaveLen(evhCount))
}

func VerifyEvhVsCacheChildDeletion(t *testing.T, g *gomega.WithT, vsKey cache.NamespaceName) {
	mcache := cache.SharedAviObjCache()
	g.Eventually(func() bool {
		evhCache, found := mcache.VsCacheMeta.AviCacheGet(vsKey)
		evhCacheObj, _ := evhCache.(*cache.AviVsCache)
		if found {
			return len(evhCacheObj.SNIChildCollection) == 0
		}
		return true
	}, 50*time.Second).Should(gomega.Equal(true))
}

func SetUpTestForIngressInNodePortMode(t *testing.T, svcName, model_Name, externalTrafficPolicy string) {
	objects.SharedAviGraphLister().Delete(model_Name)
	if externalTrafficPolicy == "" {
		integrationtest.CreateSVC(t, "default", svcName, corev1.ProtocolTCP, corev1.ServiceTypeNodePort, false)
	} else {
		integrationtest.CreateSvcWithExternalTrafficPolicy(t, "default", svcName, corev1.ProtocolTCP, corev1.ServiceTypeNodePort, false, externalTrafficPolicy)
	}
}

func TearDownTestForIngressInNodePortMode(t *testing.T, svcName, model_Name string) {
	objects.SharedAviGraphLister().Delete(model_Name)
	integrationtest.DelSVC(t, "default", svcName)
}

func TestL7ModelForEvh(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName, _ := GetModelName("foo.com", "default")
	ingressName := objNameMap.GenerateName("foo-with-targets")
	svcName := objNameMap.GenerateName("avisvc")
	SetUpTestForIngress(t, svcName, modelName)

	integrationtest.PollForCompletion(t, modelName, 5)
	// This check is moot since we were deleting the model earlier,
	// right before this check. Commenting out.
	// found, _ := objects.SharedAviGraphLister().Get(modelName)
	// if found {
	// 	// We shouldn't get an update for this update since it neither belongs to an ingress nor a L4 LB service
	// 	t.Fatalf("Couldn't find Model for DELETE event %v", modelName)
	// }
	ingrFake := (integrationtest.FakeIngress{
		Name:        ingressName,
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		ServiceName: svcName,
	}).Ingress()

	_, err := KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	integrationtest.PollForCompletion(t, modelName, 5)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 5*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(len(nodes)).To(gomega.Equal(1))
	g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
	g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
	g.Expect(len(nodes[0].EvhNodes)).To(gomega.Equal(1))

	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), ingressName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyEvhPoolDeletion(t, g, aviModel, 0)
	VerifyEvhIngressDeletion(t, g, aviModel, 0)
	VerifyEvhVsCacheChildDeletion(t, g, cache.NamespaceName{Namespace: "admin", Name: modelName})
	TearDownTestForIngress(t, svcName, modelName)
}

func TestL7ModelForEvhNodePort(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	nodeName := "testNodeNP"
	integrationtest.SetNodePortMode()
	defer integrationtest.SetClusterIPMode()
	nodeIP := "10.1.1.2"
	integrationtest.CreateNode(t, nodeName, nodeIP)
	defer integrationtest.DeleteNode(t, nodeName)

	modelName, _ := GetModelName("foo.com", "default")
	ingressName := objNameMap.GenerateName("foo-with-targets")
	svcName := objNameMap.GenerateName("avisvc")
	SetUpTestForIngressInNodePortMode(t, svcName, modelName, "")

	integrationtest.PollForCompletion(t, modelName, 5)

	ingrFake := (integrationtest.FakeIngress{
		Name:        ingressName,
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		ServiceName: svcName,
	}).Ingress()

	_, err := KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	integrationtest.PollForCompletion(t, modelName, 5)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 5*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(len(nodes)).To(gomega.Equal(1))
	g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
	g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
	g.Expect(len(nodes[0].EvhNodes)).To(gomega.Equal(1))
	g.Expect(nodes[0].EvhNodes[0].PoolRefs).To(gomega.HaveLen(1))
	// pool server is added for testNodeNP node even though endpointslice/endpoint does not exist
	g.Eventually(func() int {
		return len(nodes[0].EvhNodes[0].PoolRefs[0].Servers)
	}, 30*time.Second).Should(gomega.Equal(1))
	g.Expect(*nodes[0].EvhNodes[0].PoolRefs[0].Servers[0].Ip.Addr).To(gomega.Equal(nodeIP))
	g.Expect(len(nodes[0].EvhNodes[0].PoolRefs[0].NetworkPlacementSettings)).To(gomega.Equal(1))
	_, ok := nodes[0].EvhNodes[0].PoolRefs[0].NetworkPlacementSettings["net123"]
	g.Expect(ok).To(gomega.Equal(true))

	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), ingressName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyEvhPoolDeletion(t, g, aviModel, 0)
	VerifyEvhIngressDeletion(t, g, aviModel, 0)
	VerifyEvhVsCacheChildDeletion(t, g, cache.NamespaceName{Namespace: "admin", Name: modelName})
	TearDownTestForIngressInNodePortMode(t, svcName, modelName)
}

// TestL7ModelForEvhNodePortExternalTrafficPolicyLocal checks if pool servers are populated in model only for nodes that are running the app pod.
func TestL7ModelForEvhNodePortExternalTrafficPolicyLocal(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	//Set os.Setenv("ENDPOINTSLICES_ENABLED", "true") //in TestMain to run this locally without make
	nodeName := "testNodeNP"
	integrationtest.SetNodePortMode()
	defer integrationtest.SetClusterIPMode()
	nodeIP := "10.1.1.2"
	integrationtest.CreateNode(t, nodeName, nodeIP)
	defer integrationtest.DeleteNode(t, nodeName)

	modelName, _ := GetModelName("foo.com", "default")
	ingressName := objNameMap.GenerateName("foo-with-targets")
	svcName := objNameMap.GenerateName("avisvc")
	SetUpTestForIngressInNodePortMode(t, svcName, modelName, "Local")

	integrationtest.PollForCompletion(t, modelName, 5)

	ingrFake := (integrationtest.FakeIngress{
		Name:        ingressName,
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		ServiceName: svcName,
	}).Ingress()

	_, err := KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	integrationtest.PollForCompletion(t, modelName, 5)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 5*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(len(nodes)).To(gomega.Equal(1))
	g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
	g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
	g.Expect(len(nodes[0].EvhNodes)).To(gomega.Equal(1))
	g.Expect(nodes[0].EvhNodes[0].PoolRefs).To(gomega.HaveLen(1))
	// No pool server is added as endpointslice/endpoint does not exist
	g.Expect(nodes[0].EvhNodes[0].PoolRefs[0].Servers).To(gomega.HaveLen(0))
	g.Expect(len(nodes[0].EvhNodes[0].PoolRefs[0].NetworkPlacementSettings)).To(gomega.Equal(1))
	_, ok := nodes[0].EvhNodes[0].PoolRefs[0].NetworkPlacementSettings["net123"]
	g.Expect(ok).To(gomega.Equal(true))

	integrationtest.CreateEPorEPSNodeName(t, "default", svcName, false, false, "1.1.1", nodeName)
	// After creating the endpointslice/endpoint, pool server should be added for testNodeNP node
	g.Eventually(func() int {
		return len(nodes[0].EvhNodes[0].PoolRefs[0].Servers)
	}, 30*time.Second).Should(gomega.Equal(1))
	g.Expect(*nodes[0].EvhNodes[0].PoolRefs[0].Servers[0].Ip.Addr).To(gomega.Equal(nodeIP))

	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), ingressName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	integrationtest.DelEPorEPS(t, "default", svcName)
	VerifyEvhPoolDeletion(t, g, aviModel, 0)
	VerifyEvhIngressDeletion(t, g, aviModel, 0)
	VerifyEvhVsCacheChildDeletion(t, g, cache.NamespaceName{Namespace: "admin", Name: modelName})
	TearDownTestForIngressInNodePortMode(t, svcName, modelName)
}

// This tests the different objects associated in the evh model for ingress
func TestShardObjectsForEvh(t *testing.T) {
	// checks naming convention of all generated nodes

	g := gomega.NewGomegaWithT(t)

	modelName, vsName := GetModelName("foo.com", "default")
	secretName := objNameMap.GenerateName("my-secret")
	ingressName := objNameMap.GenerateName("foo-with-targets")
	svcName := objNameMap.GenerateName("avisvc")
	SetUpTestForIngress(t, svcName, modelName)
	integrationtest.AddSecret(secretName, "default", "tlsCert", "tlsKey")

	// foo.com and noo.com compute the same hashed shard vs num
	ingrFake := (integrationtest.FakeIngress{
		Name:      ingressName,
		Namespace: "default",
		DnsNames:  []string{"foo.com", "noo.com"},
		Ips:       []string{"8.8.8.8"},
		Paths:     []string{"/foo/bar"},
		HostNames: []string{"v1"},
		TlsSecretDNS: map[string][]string{
			secretName: {"foo.com"},
		},
		ServiceName: svcName,
	}).IngressMultiPath()

	_, err := KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	integrationtest.PollForCompletion(t, modelName, 5)

	verifyIng, _ := KubeClient.NetworkingV1().Ingresses("default").Get(context.TODO(), ingressName, metav1.GetOptions{})
	for i, host := range []string{"foo.com", "noo.com"} {
		if verifyIng.Spec.Rules[i].Host == host {
			g.Expect(verifyIng.Spec.Rules[i].Host).To(gomega.Equal(host))
			g.Expect(verifyIng.Spec.Rules[i].HTTP.Paths[0].Path).To(gomega.Equal("/foo/bar"))
		}
	}

	time.Sleep(10 * time.Second)
	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 5*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(nodes[0].Name).To(gomega.Equal(vsName))
	// Shared VS in EVH will not have any pool or pool group unlike the normal VS
	g.Expect(len(nodes[0].PoolGroupRefs)).To(gomega.Equal(0))
	g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(0))
	g.Expect(nodes[0].HTTPDSrefs).Should(gomega.HaveLen(0))
	g.Expect(nodes[0].VSVIPRefs[0].Name).To(gomega.Equal(vsName))
	// the certs will be associated to parent evh vs
	g.Expect(nodes[0].SSLKeyCertRefs).Should(gomega.HaveLen(1))
	// There will be 2 evh node one for each host
	g.Expect(nodes[0].EvhNodes).Should(gomega.HaveLen(2))
	g.Expect(nodes[0].EvhNodes[0].Name).To(gomega.Equal(lib.Encode("cluster--noo.com", lib.EVHVS)))
	g.Expect(nodes[0].EvhNodes[0].PoolGroupRefs[0].Name).To(gomega.Equal(lib.Encode("cluster--default-noo.com_foo_bar-"+ingressName, lib.PG)))
	g.Expect(nodes[0].EvhNodes[0].PoolRefs[0].Name).To(gomega.Equal(lib.Encode("cluster--default-noo.com_foo_bar-"+ingressName+"-"+svcName, lib.Pool)))
	// Shared VS in EVH will not have any certificates and httppolicy
	g.Expect(nodes[0].EvhNodes[0].SSLKeyCertRefs).Should(gomega.HaveLen(0))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs).Should(gomega.HaveLen(1))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].Name).To(gomega.Equal(lib.Encode("cluster--default-noo.com", lib.HTTPPS)))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[0].Name).To(gomega.Equal(lib.Encode("cluster--default-noo.com_foo_bar-"+ingressName, lib.HPPMAP)))

	g.Expect(nodes[0].EvhNodes[1].Name).To(gomega.Equal(lib.Encode("cluster--foo.com", lib.EVHVS)))
	g.Expect(nodes[0].EvhNodes[1].PoolGroupRefs[0].Name).To(gomega.Equal(lib.Encode("cluster--default-foo.com_foo_bar-"+ingressName, lib.PG)))
	g.Expect(nodes[0].EvhNodes[1].PoolRefs[0].Name).To(gomega.Equal(lib.Encode("cluster--default-foo.com_foo_bar-"+ingressName+"-"+svcName, lib.Pool)))
	// since foo is bound with cert this node will have the cert bound to it
	g.Expect(nodes[0].EvhNodes[1].SSLKeyCertRefs).Should(gomega.HaveLen(0))
	g.Expect(nodes[0].EvhNodes[1].HttpPolicyRefs).Should(gomega.HaveLen(2))
	g.Expect(nodes[0].EvhNodes[1].HttpPolicyRefs[0].Name).To(gomega.Equal(lib.Encode("cluster--default-foo.com", lib.HTTPPS)))
	g.Expect(nodes[0].EvhNodes[1].HttpPolicyRefs[0].HppMap[0].Name).To(gomega.Equal(lib.Encode("cluster--default-foo.com_foo_bar-"+ingressName, lib.HPPMAP)))
	g.Expect(nodes[0].EvhNodes[1].HttpPolicyRefs[1].Name).To(gomega.Equal(lib.Encode("cluster--foo.com", lib.HTTPPS)))

	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), ingressName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyEvhVsCacheChildDeletion(t, g, cache.NamespaceName{Namespace: "admin", Name: modelName})
	TearDownTestForIngress(t, svcName, modelName)
}

func TestNoBackendL7ModelForEvh(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName, _ := GetModelName("foo.com", "default")
	ingressName := objNameMap.GenerateName("foo-with-targets")
	svcName := objNameMap.GenerateName("avisvc")
	SetUpTestForIngress(t, svcName, modelName)
	objects.SharedAviGraphLister().Delete(modelName)

	integrationtest.PollForCompletion(t, modelName, 5)
	// found, _ := objects.SharedAviGraphLister().Get(modelName)
	// if found {
	// 	// We shouldn't get an update for this update since it neither belongs to an ingress nor a L4 LB service
	// 	t.Fatalf("Couldn't find Model for DELETE event %v", modelName)
	// }
	ingrFake := (integrationtest.FakeIngress{
		Name:      ingressName,
		Namespace: "default",
		DnsNames:  []string{"foo.com"},
		Paths:     []string{"/"},
	}).IngressOnlyHostNoBackend()

	_, err := KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	integrationtest.PollForCompletion(t, modelName, 5)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 15*time.Second).Should(gomega.Equal(false))

	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), ingressName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyEvhVsCacheChildDeletion(t, g, cache.NamespaceName{Namespace: "admin", Name: modelName})
	TearDownTestForIngress(t, svcName, modelName)
}

func TestMultiIngressToSameSvcForEvh(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	modelName, _ := GetModelName("foo.com", "default")
	objects.SharedAviGraphLister().Delete(modelName)
	svcName := objNameMap.GenerateName("avisvc")
	ingressName1 := objNameMap.GenerateName("foo-with-targets")
	ingressName2 := objNameMap.GenerateName("foo-with-targets")
	svcExample := (integrationtest.FakeService{
		Name:         svcName,
		Namespace:    "default",
		Type:         corev1.ServiceTypeClusterIP,
		ServicePorts: []integrationtest.Serviceport{{PortName: "foo", Protocol: "TCP", PortNumber: 8080, TargetPort: intstr.FromInt(8080)}},
	}).Service()

	_, err := KubeClient.CoreV1().Services("default").Create(context.TODO(), svcExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Service: %v", err)
	}
	integrationtest.CreateEPorEPS(t, "default", svcName, false, false, "1.1.1")
	ingrFake1 := (integrationtest.FakeIngress{
		Name:        ingressName1,
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		ServiceName: svcName,
	}).Ingress()

	_, err = KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake1, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	ingrFake2 := (integrationtest.FakeIngress{
		Name:        ingressName2,
		Namespace:   "default",
		DnsNames:    []string{"bar.com"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		ServiceName: svcName,
	}).Ingress()

	_, err = KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake2, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	integrationtest.PollForCompletion(t, modelName, 5)
	found, aviModel := objects.SharedAviGraphLister().Get(modelName)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		g.Expect(nodes[0].SharedVS).To(gomega.Equal(true))
		dsNodes := aviModel.(*avinodes.AviObjectGraph).GetAviHTTPDSNode()
		g.Expect(len(dsNodes)).To(gomega.Equal(0))
		g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(0))
		// Delete the model.
		objects.SharedAviGraphLister().Delete(modelName)
	} else {
		t.Fatalf("Could not find model on ingress delete: %v", err)
	}
	//====== VERIFICATION OF SERVICE DELETE
	// Now we have cleared the layer 2 queue for both the models. Let's delete the service.
	integrationtest.DelEPorEPS(t, "default", svcName)
	err = KubeClient.CoreV1().Services("default").Delete(context.TODO(), svcName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Service %v", err)
	}
	// We should be able to get one model now in the queue
	integrationtest.PollForCompletion(t, modelName, 5)
	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 15*time.Second).Should(gomega.Equal(true))
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(len(nodes)).To(gomega.Equal(1))
	g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
	g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
	dsNodes := aviModel.(*avinodes.AviObjectGraph).GetAviHTTPDSNode()
	g.Expect(len(dsNodes)).To(gomega.Equal(0))
	g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(0))

	integrationtest.CreateEPorEPS(t, "default", svcName, false, false, "1.1.1")
	//====== VERIFICATION OF ONE INGRESS DELETE
	// Now let's delete one ingress and expect the update for that.
	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), ingressName1, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	integrationtest.DetectModelChecksumChange(t, modelName, 5)
	found, aviModel = objects.SharedAviGraphLister().Get(modelName)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(0))

		// Delete the model.
		objects.SharedAviGraphLister().Delete(modelName)
	} else {
		t.Fatalf("Could not find model on ingress delete: %v", err)
	}
	//====== VERIFICATION OF SERVICE ADD
	// Let's add the service back now - the ingress's associated with this service should be returned
	_, err = KubeClient.CoreV1().Services("default").Create(context.TODO(), svcExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Service: %v", err)
	}
	modelName, _ = GetModelName("bar.com", "default")
	integrationtest.PollForCompletion(t, modelName, 5)
	// We should be able to get one model now in the queue
	found, aviModel = objects.SharedAviGraphLister().Get(modelName)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(0))

		objects.SharedAviGraphLister().Delete(modelName)
	} else {
		t.Fatalf("Could not find model on service ADD: %v", err)
	}
	//====== VERIFICATION OF ONE ENDPOINT DELETE
	integrationtest.DelEPorEPS(t, "default", svcName)
	integrationtest.PollForCompletion(t, modelName, 5)
	// Deletion should also give us the affected ingress objects
	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 20*time.Second).Should(gomega.Equal(true))
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(len(nodes)).To(gomega.Equal(1))
	g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
	g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
	// Delete the model.
	objects.SharedAviGraphLister().Delete(modelName)

	err = KubeClient.CoreV1().Services("default").Delete(context.TODO(), svcName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Service %v", err)
	}
	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), ingressName2, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
}

// TestMultiPathIngressForEvh in evh mode will validate if 2 evh nodes with host + path are created
func TestMultiPathIngressForEvh(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	var err error

	modelName, _ := GetModelName("foo.com", "default")
	ingressName := objNameMap.GenerateName("foo-with-targets")
	svcName := objNameMap.GenerateName("avisvc")
	SetUpTestForIngress(t, svcName, modelName)

	ingrFake := (integrationtest.FakeIngress{
		Name:        ingressName,
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/foo", "/bar"},
		ServiceName: svcName,
	}).IngressMultiPath()

	_, err = KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	integrationtest.PollForCompletion(t, modelName, 5)
	found, aviModel := objects.SharedAviGraphLister().Get(modelName)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		g.Expect(nodes[0].PoolRefs).Should(gomega.HaveLen(0))

		g.Expect(nodes[0].EvhNodes).Should(gomega.HaveLen(1))
		g.Expect(nodes[0].EvhNodes[0].Name).To(gomega.Equal(lib.Encode("cluster--foo.com", lib.EVHVS)))
		g.Expect(nodes[0].EvhNodes[0].EvhHostName).To(gomega.Equal("foo.com"))
		g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs).Should(gomega.HaveLen(1))
		g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap).Should(gomega.HaveLen(2))
		g.Expect(len(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[0].Path), gomega.Equal(1))
		g.Expect(len(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[1].Path), gomega.Equal(1))
		g.Expect(func() []string {
			p := []string{
				nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[0].Path[0],
				nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[1].Path[0]}
			sort.Strings(p)
			return p
		}, gomega.Equal([]string{"/bar", "/foo"}))
		g.Expect(func() []string {
			p := []string{
				nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[0].Name,
				nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[1].Name}
			sort.Strings(p)
			return p
		}, gomega.Equal([]string{lib.Encode("cluster--default-foo.com_bar-"+ingressName, lib.HPPMAP),
			lib.Encode("cluster--default-foo.com_foo-"+ingressName, lib.HPPMAP)}))
		g.Expect(len(nodes[0].EvhNodes[0].PoolGroupRefs)).To(gomega.Equal(2))
		g.Expect(len(nodes[0].EvhNodes[0].PoolGroupRefs[0].Members)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].EvhNodes[0].PoolGroupRefs[1].Members)).To(gomega.Equal(1))

	} else {
		t.Fatalf("Could not find model: %s", modelName)
	}
	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), ingressName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyEvhPoolDeletion(t, g, aviModel, 0)
	VerifyEvhIngressDeletion(t, g, aviModel, 0)
	VerifyEvhVsCacheChildDeletion(t, g, cache.NamespaceName{Namespace: "admin", Name: modelName})
	TearDownTestForIngress(t, svcName, modelName)
}

func TestMultiPortServiceIngressForEvh(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	var err error

	modelName, _ := GetModelName("foo.com", "default")
	objects.SharedAviGraphLister().Delete(modelName)
	ingressName := objNameMap.GenerateName("foo-with-targets")
	svcName := objNameMap.GenerateName("avisvc")
	integrationtest.CreateSVC(t, "default", svcName, corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, true)
	integrationtest.CreateEPorEPS(t, "default", svcName, true, true, "1.1.1")
	ingrFake := (integrationtest.FakeIngress{
		Name:        ingressName,
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/foo"},
		ServiceName: svcName,
	}).IngressMultiPort()

	_, err = KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	integrationtest.PollForCompletion(t, modelName, 5)
	found, aviModel := objects.SharedAviGraphLister().Get(modelName)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(0))

		g.Expect(len(nodes[0].EvhNodes)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].EvhNodes[0].PoolRefs)).To(gomega.Equal(2))
		for _, pool := range nodes[0].EvhNodes[0].PoolRefs {
			if pool.Name == lib.Encode("cluster--default-foo.com_foo-"+ingressName+"-"+svcName, lib.Pool) {
				g.Expect(pool.Port).To(gomega.Equal(int32(8080)))
				g.Expect(len(pool.Servers)).To(gomega.Equal(3))
			} else if pool.Name == lib.Encode("cluster--default-foo.com_bar-"+ingressName+"-"+svcName, lib.Pool) {
				g.Expect(pool.Port).To(gomega.Equal(int32(8081)))
				g.Expect(len(pool.Servers)).To(gomega.Equal(2))
			} else {
				t.Fatalf("unexpected pool: %s", pool.Name)
			}
		}

		g.Expect(nodes[0].EvhNodes).Should(gomega.HaveLen(1))
		g.Expect(nodes[0].EvhNodes[0].Name).To(gomega.Equal(lib.Encode("cluster--foo.com", lib.EVHVS)))
		g.Expect(nodes[0].EvhNodes[0].EvhHostName).To(gomega.Equal("foo.com"))
		g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs).Should(gomega.HaveLen(1))
		g.Expect(len(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap), gomega.Equal(2))
		g.Expect(len(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[0].Path), gomega.Equal(1))
		g.Expect(len(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[1].Path), gomega.Equal(1))
		g.Expect(func() []string {
			p := []string{
				nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[0].Path[0],
				nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[1].Path[0]}
			sort.Strings(p)
			return p
		}, gomega.Equal([]string{"/bar", "/foo"}))
		g.Expect(func() []string {
			p := []string{
				nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[0].Name,
				nodes[0].EvhNodes[0].HttpPolicyRefs[1].HppMap[0].Name}
			sort.Strings(p)
			return p
		}, gomega.Equal([]string{"cluster--default-foo.com_bar-" + ingressName,
			"cluster--default-foo.com_foo-" + ingressName}))
		g.Expect(len(nodes[0].EvhNodes[0].PoolGroupRefs)).To(gomega.Equal(2))

	} else {
		t.Fatalf("Could not find model: %s", modelName)
	}
	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), ingressName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyEvhPoolDeletion(t, g, aviModel, 0)
	VerifyEvhIngressDeletion(t, g, aviModel, 0)
	time.Sleep(15 * time.Second)
	TearDownTestForIngress(t, svcName, modelName)
	VerifyEvhVsCacheChildDeletion(t, g, cache.NamespaceName{Namespace: "admin", Name: modelName})
}

func TestMultiIngressSameHostForEvh(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName, _ := GetModelName("foo.com", "default")
	svcName := objNameMap.GenerateName("avisvc")
	ingressName1 := objNameMap.GenerateName("foo-with-targets")
	ingressName2 := objNameMap.GenerateName("foo-with-targets")
	SetUpTestForIngress(t, svcName, modelName)

	ingrFake1 := (integrationtest.FakeIngress{
		Name:        ingressName1,
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/foo"},
		ServiceName: svcName,
	}).Ingress()

	_, err := KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake1, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	ingrFake2 := (integrationtest.FakeIngress{
		Name:        ingressName2,
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/bar"},
		ServiceName: svcName,
	}).Ingress()

	_, err = KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake2, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	integrationtest.PollForCompletion(t, modelName, 5)
	found, aviModel := objects.SharedAviGraphLister().Get(modelName)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(0))

		g.Eventually(func() int {
			_, aviModel := objects.SharedAviGraphLister().Get(modelName)
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
			return len(nodes[0].EvhNodes)
		}, 10*time.Second).Should(gomega.Equal(1))
		g.Expect(nodes[0].EvhNodes[0].Name).To(gomega.Equal(lib.Encode("cluster--foo.com", lib.EVHVS)))
		g.Expect(nodes[0].EvhNodes[0].EvhHostName).To(gomega.Equal("foo.com"))
		g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs).Should(gomega.HaveLen(1))
		g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap).Should(gomega.HaveLen(2))
		g.Expect(len(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[0].Path), gomega.Equal(1))
		g.Expect(len(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[1].Path), gomega.Equal(1))
		g.Expect(func() []string {
			p := []string{
				nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[0].Path[0],
				nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[1].Path[0]}
			sort.Strings(p)
			return p
		}, gomega.Equal([]string{"/bar", "/foo"}))
		g.Expect(func() []string {
			p := []string{
				nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[0].Name,
				nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[1].Name}
			sort.Strings(p)
			return p
		}, gomega.Equal([]string{lib.Encode("cluster--default-foo.com_bar-"+ingressName1, lib.HPPMAP),
			lib.Encode("cluster--default-foo.com_foo-"+ingressName2, lib.HPPMAP)}))
		g.Expect(len(nodes[0].EvhNodes[0].PoolGroupRefs)).To(gomega.Equal(2))
	} else {
		t.Fatalf("Could not find model: %s", modelName)
	}
	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), ingressName1, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyEvhPoolDeletion(t, g, aviModel, 0)
	VerifyEvhIngressDeletion(t, g, aviModel, 1)
	integrationtest.DetectModelChecksumChange(t, modelName, 5)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(len(nodes)).To(gomega.Equal(1))
	g.Expect(len(nodes[0].EvhNodes)).To(gomega.Equal(1))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs).Should(gomega.HaveLen(1))

	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), ingressName2, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyEvhPoolDeletion(t, g, aviModel, 0)
	VerifyEvhIngressDeletion(t, g, aviModel, 0)
	VerifyEvhVsCacheChildDeletion(t, g, cache.NamespaceName{Namespace: "admin", Name: modelName})
	TearDownTestForIngress(t, svcName, modelName)
}

func TestDeleteBackendServiceForEvh(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName, _ := GetModelName("foo.com", "default")
	svcName := objNameMap.GenerateName("avisvc")
	ingressName1 := objNameMap.GenerateName("foo-with-targets")
	ingressName2 := objNameMap.GenerateName("foo-with-targets")
	SetUpTestForIngress(t, svcName, modelName)

	ingrFake1 := (integrationtest.FakeIngress{
		Name:        ingressName1,
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/foo"},
		ServiceName: svcName,
	}).Ingress()

	_, err := KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake1, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	ingrFake2 := (integrationtest.FakeIngress{
		Name:        ingressName2,
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/bar"},
		ServiceName: svcName,
	}).Ingress()

	_, err = KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake2, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	integrationtest.PollForCompletion(t, modelName, 5)
	found, aviModel := objects.SharedAviGraphLister().Get(modelName)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(0))
		g.Eventually(func() int {
			_, aviModel := objects.SharedAviGraphLister().Get(modelName)
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
			if len(nodes[0].EvhNodes) > 0 && len(nodes[0].EvhNodes[0].PoolRefs) > 0 {
				return len(nodes[0].EvhNodes[0].PoolRefs[0].Servers)
			}
			return 0
		}, 10*time.Second).Should(gomega.Equal(1))

	} else {
		t.Fatalf("Could not find model: %s", modelName)
	}
	// Delete the service
	integrationtest.DelSVC(t, "default", svcName)
	integrationtest.DelEPorEPS(t, "default", svcName)
	g.Eventually(func() bool {
		found, _ = objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 20*time.Second).Should(gomega.Equal(true))
	found, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(len(nodes)).To(gomega.Equal(1))
	g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
	g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
	g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(0))
	g.Expect(len(nodes[0].EvhNodes[0].HttpPolicyRefs)).To(gomega.Equal(1))
	g.Eventually(func() int {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return len(nodes[0].EvhNodes[0].PoolRefs[0].Servers)
	}, 30*time.Second).Should(gomega.Equal(0))

	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), ingressName1, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	integrationtest.DetectModelChecksumChange(t, modelName, 5)
	VerifyEvhPoolDeletion(t, g, aviModel, 0)
	VerifyEvhIngressDeletion(t, g, aviModel, 1)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(len(nodes)).To(gomega.Equal(1))
	g.Expect(nodes[0].EvhNodes[0].PoolRefs[0].Name).To(gomega.Equal(lib.Encode("cluster--default-foo.com_bar-"+ingressName2+"-"+svcName, lib.Pool)))

	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), ingressName2, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyEvhPoolDeletion(t, g, aviModel, 0)
	VerifyEvhIngressDeletion(t, g, aviModel, 0)
	VerifyEvhVsCacheChildDeletion(t, g, cache.NamespaceName{Namespace: "admin", Name: modelName})
	TearDownTestForIngress(t, svcName, modelName)
}

func TestUpdateBackendServiceForEvh(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	modelName, _ := GetModelName("foo.com", "default")

	ingressName := objNameMap.GenerateName("ingress-multipath")
	svcName1 := objNameMap.GenerateName("avisvc")
	svcName2 := objNameMap.GenerateName("avisvc")

	SetUpTestForIngress(t, svcName1, modelName)
	ingrFake1 := (integrationtest.FakeIngress{
		Name:        ingressName,
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/foo"},
		ServiceName: svcName1,
	}).Ingress()
	_, err := KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake1, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	integrationtest.PollForCompletion(t, modelName, 5)
	g.Eventually(func() string {
		if found, aviModel := objects.SharedAviGraphLister().Get(modelName); found {
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
			if len(nodes[0].EvhNodes) > 0 && len(nodes[0].EvhNodes[0].PoolRefs) > 0 {
				return *nodes[0].EvhNodes[0].PoolRefs[0].Servers[0].Ip.Addr
			}
		}
		return ""
	}, 15*time.Second).Should(gomega.Equal("1.1.1.1"))

	// Update the service
	integrationtest.CreateSVC(t, "default", svcName2, corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPorEPS(t, "default", svcName2, false, false, "2.2.2")

	_, err = (integrationtest.FakeIngress{
		Name:        ingressName,
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/foo"},
		ServiceName: svcName2,
	}).UpdateIngress()
	if err != nil {
		t.Fatalf("error in updating ingress %s", err)
	}

	found, aviModel := objects.SharedAviGraphLister().Get(modelName)
	if found {
		g.Eventually(func() string {
			_, aviModel := objects.SharedAviGraphLister().Get(modelName)
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
			if len(nodes) > 0 && len(nodes[0].EvhNodes) > 0 && len(nodes[0].EvhNodes[0].PoolRefs) > 0 && len(nodes[0].EvhNodes[0].PoolRefs[0].Servers) > 0 {
				return *nodes[0].EvhNodes[0].PoolRefs[0].Servers[0].Ip.Addr
			}
			return ""
		}, 15*time.Second).Should(gomega.Equal("2.2.2.1"))
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		g.Expect(len(nodes[0].EvhNodes)).To(gomega.Equal(1))
	} else {
		t.Fatalf("Could not find model: %s", modelName)
	}
	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), ingressName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	integrationtest.DelSVC(t, "default", svcName2)
	integrationtest.DelEPorEPS(t, "default", svcName2)
	VerifyEvhPoolDeletion(t, g, aviModel, 0)
	VerifyEvhIngressDeletion(t, g, aviModel, 0)
	VerifyEvhVsCacheChildDeletion(t, g, cache.NamespaceName{Namespace: "admin", Name: modelName})
	TearDownTestForIngress(t, svcName1, modelName)
}

func TestL2ChecksumsUpdateForEvh(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName, _ := GetModelName("foo.com", "default")
	secretName1 := objNameMap.GenerateName("my-secret")
	ingressName := objNameMap.GenerateName("foo-with-targets")
	svcName1 := objNameMap.GenerateName("avisvc")
	SetUpTestForIngress(t, svcName1, modelName)
	integrationtest.AddSecret(secretName1, "default", "tlsCert", "tlsKey")
	//create ingress with tls secret
	ingrFake1 := (integrationtest.FakeIngress{
		Name:      ingressName,
		Namespace: "default",
		DnsNames:  []string{"foo.com"},
		Ips:       []string{"8.8.8.8"},
		Paths:     []string{"/foo"},
		HostNames: []string{"v1"},
		TlsSecretDNS: map[string][]string{
			secretName1: {"foo.com"},
		},
		ServiceName: svcName1,
	}).Ingress()
	_, err := KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake1, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	initCheckSums := make(map[string]uint32)
	integrationtest.PollForCompletion(t, modelName, 5)
	g.Eventually(func() int {
		if found, aviModel := objects.SharedAviGraphLister().Get(modelName); found {
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
			return len(nodes[0].EvhNodes)
		}
		return 0
	}, 15*time.Second).Should(gomega.Equal(1))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	t.Logf("nodes %s", utils.Stringify(nodes))
	initCheckSums["nodes[0]"] = nodes[0].CloudConfigCksum

	g.Expect(len(nodes[0].EvhNodes)).To(gomega.Equal(1))
	initCheckSums["nodes[0].EvhNodes[0]"] = nodes[0].EvhNodes[0].CloudConfigCksum

	g.Expect(len(nodes[0].EvhNodes[0].PoolRefs)).To(gomega.Equal(1))
	initCheckSums["nodes[0].EvhNodes[0].PoolRefs[0]"] = nodes[0].EvhNodes[0].PoolRefs[0].CloudConfigCksum

	g.Expect(len(nodes[0].EvhNodes[0].SSLKeyCertRefs)).To(gomega.Equal(0))
	g.Expect(len(nodes[0].SSLKeyCertRefs)).To(gomega.Equal(1))
	initCheckSums["nodes[0].SSLKeyCertRefs[0]"] = nodes[0].SSLKeyCertRefs[0].CloudConfigCksum

	g.Expect(len(nodes[0].HttpPolicyRefs)).To(gomega.Equal(0))

	secretName2, svcName2 := "my-secret-71", "avisvc-71"

	integrationtest.CreateSVC(t, "default", svcName2, corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPorEPS(t, "default", svcName2, false, false, "2.2.2")
	integrationtest.AddSecret(secretName2, "default", "tlsCert-new", "tlsKey")

	_, err = (integrationtest.FakeIngress{
		Name:      ingressName,
		Namespace: "default",
		DnsNames:  []string{"foo.com"},
		Ips:       []string{"8.8.8.8"},
		//to update httppolicyref checksum
		Paths:     []string{"/bar"},
		HostNames: []string{"v1"},
		TlsSecretDNS: map[string][]string{
			//to update tls secret checksum
			secretName2: {"foo.com"},
		},
		//to update poolref checksum
		ServiceName: svcName2,
	}).UpdateIngress()
	if err != nil {
		t.Fatalf("error in updating ingress %s", err)
	}
	time.Sleep(15 * time.Second)
	integrationtest.PollForCompletion(t, modelName, 5)
	integrationtest.DetectModelChecksumChange(t, modelName, 10)
	found, aviModel := objects.SharedAviGraphLister().Get(modelName)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		t.Logf("nodes: %v", utils.Stringify(nodes))
		g.Eventually(len(nodes), 5*time.Second).Should(gomega.Equal(1))

		g.Expect(len(nodes[0].EvhNodes)).To(gomega.Equal(1))

		g.Expect(len(nodes[0].EvhNodes[0].PoolRefs)).To(gomega.Equal(1))
		g.Eventually(func() uint32 {
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
			return nodes[0].EvhNodes[0].PoolRefs[0].CloudConfigCksum
		}, 5*time.Second).ShouldNot(gomega.Equal(initCheckSums["nodes[0].EvhNodes[0].PoolRefs[0]"]))

		g.Expect(len(nodes[0].EvhNodes[0].SSLKeyCertRefs)).To(gomega.Equal(0))
		g.Eventually(func() uint32 {
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
			return nodes[0].SSLKeyCertRefs[0].CloudConfigCksum
		}, 5*time.Second).ShouldNot(gomega.Equal(initCheckSums["nodes[0].SSLKeyCertRefs[0]"]))

		g.Expect(len(nodes[0].EvhNodes[0].HttpPolicyRefs)).To(gomega.Equal(2))

	} else {
		t.Fatalf("Could not find model: %s", modelName)
	}
	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), ingressName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	integrationtest.DelSVC(t, "default", svcName2)
	integrationtest.DelEPorEPS(t, "default", svcName2)
	KubeClient.CoreV1().Secrets("default").Delete(context.TODO(), secretName1, metav1.DeleteOptions{})
	KubeClient.CoreV1().Secrets("default").Delete(context.TODO(), secretName2, metav1.DeleteOptions{})
	VerifyEvhIngressDeletion(t, g, aviModel, 0)
	VerifyEvhVsCacheChildDeletion(t, g, cache.NamespaceName{Namespace: "admin", Name: modelName})
	TearDownTestForIngress(t, svcName1, modelName)
}

func TestMultiHostSameHostNameIngressForEvh(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	modelName, _ := GetModelName("foo.com", "default")
	ingressName := objNameMap.GenerateName("foo-with-targets")
	svcName := objNameMap.GenerateName("avisvc")
	SetUpTestForIngress(t, svcName, modelName)

	ingrFake := (integrationtest.FakeIngress{
		Name:        ingressName,
		Namespace:   "default",
		DnsNames:    []string{"foo.com", "foo.com"},
		Paths:       []string{"/foo", "/bar"},
		ServiceName: svcName,
	}).Ingress()

	_, err := KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	integrationtest.PollForCompletion(t, modelName, 5)
	g.Eventually(func() int {
		if found, aviModel := objects.SharedAviGraphLister().Get(modelName); found {
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
			return len(nodes[0].EvhNodes)
		}
		return 0
	}, 15*time.Second).Should(gomega.Equal(1))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(len(nodes)).To(gomega.Equal(1))
	g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
	g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
	g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(0))
	g.Expect(len(nodes[0].PoolGroupRefs)).To(gomega.Equal(0))
	g.Expect(len(nodes[0].EvhNodes)).To(gomega.Equal(1))
	g.Expect(nodes[0].EvhNodes[0].Name).To(gomega.Equal(lib.Encode("cluster--foo.com", lib.EVHVS)))
	g.Expect(len(nodes[0].EvhNodes[0].HttpPolicyRefs)).To(gomega.Equal(1))
	g.Expect(len(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap)).To(gomega.Equal(2))

	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), ingressName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyEvhPoolDeletion(t, g, aviModel, 0)
	VerifyEvhIngressDeletion(t, g, aviModel, 0)
	VerifyEvhVsCacheChildDeletion(t, g, cache.NamespaceName{Namespace: "admin", Name: modelName})
	TearDownTestForIngress(t, svcName, modelName)
}

func TestEditPathIngressForEvh(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	modelName, _ := GetModelName("foo.com", "default")
	ingressName := objNameMap.GenerateName("foo-with-targets")
	svcName := objNameMap.GenerateName("avisvc")
	SetUpTestForIngress(t, svcName, modelName)

	ingrFake := (integrationtest.FakeIngress{
		Name:        ingressName,
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/foo"},
		ServiceName: svcName,
	}).Ingress()
	ingrFake.ResourceVersion = "1"
	_, err := KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	integrationtest.PollForCompletion(t, modelName, 5)
	g.Eventually(func() int {
		if found, aviModel := objects.SharedAviGraphLister().Get(modelName); found {
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
			return len(nodes[0].EvhNodes)
		}
		return 0
	}, 15*time.Second).Should(gomega.Equal(1))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(len(nodes)).To(gomega.Equal(1))
	g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
	g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
	dsNodes := aviModel.(*avinodes.AviObjectGraph).GetAviHTTPDSNode()
	g.Expect(len(dsNodes)).To(gomega.Equal(0))
	g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(0))
	g.Expect(nodes[0].EvhNodes).Should(gomega.HaveLen(1))
	g.Expect(nodes[0].EvhNodes[0].Name).To(gomega.Equal(lib.Encode("cluster--foo.com", lib.EVHVS)))
	g.Expect(nodes[0].EvhNodes[0].PoolGroupRefs[0].Name).To(gomega.Equal(lib.Encode("cluster--default-foo.com_foo-"+ingressName, lib.PG)))
	g.Expect(nodes[0].EvhNodes[0].PoolRefs[0].Name).To(gomega.Equal(lib.Encode("cluster--default-foo.com_foo-"+ingressName+"-"+svcName, lib.Pool)))
	g.Expect(nodes[0].EvhNodes[0].EvhHostName).To(gomega.Equal("foo.com"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs).Should(gomega.HaveLen(1))
	g.Expect(len(nodes[0].EvhNodes[0].PoolGroupRefs[0].Members)).To(gomega.Equal(1))
	g.Expect(len(nodes[0].EvhNodes[0].PoolRefs[0].Servers)).To(gomega.Equal(1))

	ingrFake = (integrationtest.FakeIngress{
		Name:        ingressName,
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/bar"},
		ServiceName: svcName,
	}).Ingress()
	ingrFake.ResourceVersion = "2"
	_, err = KubeClient.NetworkingV1().Ingresses("default").Update(context.TODO(), ingrFake, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating Ingress: %v", err)
	}
	integrationtest.DetectModelChecksumChange(t, modelName, 5)
	found, aviModel := objects.SharedAviGraphLister().Get(modelName)
	if found {

		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		dsNodes := aviModel.(*avinodes.AviObjectGraph).GetAviHTTPDSNode()
		g.Expect(len(dsNodes)).To(gomega.Equal(0))
		g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(0))
		g.Expect(nodes[0].EvhNodes).Should(gomega.HaveLen(1))
		g.Expect(nodes[0].EvhNodes[0].Name).To(gomega.Equal(lib.Encode("cluster--foo.com", lib.EVHVS)))
		g.Expect(nodes[0].EvhNodes[0].PoolGroupRefs[0].Name).To(gomega.Equal(lib.Encode("cluster--default-foo.com_bar-"+ingressName, lib.PG)))
		g.Expect(nodes[0].EvhNodes[0].PoolRefs[0].Name).To(gomega.Equal(lib.Encode("cluster--default-foo.com_bar-"+ingressName+"-"+svcName, lib.Pool)))
		g.Expect(nodes[0].EvhNodes[0].EvhHostName).To(gomega.Equal("foo.com"))
		g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs).Should(gomega.HaveLen(1))
		g.Expect(len(nodes[0].EvhNodes[0].PoolGroupRefs[0].Members)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].EvhNodes[0].PoolRefs[0].Servers)).To(gomega.Equal(1))

	} else {
		t.Fatalf("Could not find model: %s", modelName)
	}

	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), ingressName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyEvhPoolDeletion(t, g, aviModel, 0)
	VerifyEvhIngressDeletion(t, g, aviModel, 0)
	VerifyEvhVsCacheChildDeletion(t, g, cache.NamespaceName{Namespace: "admin", Name: modelName})
	TearDownTestForIngress(t, svcName, modelName)
}

func TestEditMultiPathIngressForEvh(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	modelName, _ := GetModelName("foo.com", "default")
	ingressName := objNameMap.GenerateName("foo-with-targets")
	svcName := objNameMap.GenerateName("avisvc")
	SetUpTestForIngress(t, svcName, modelName)

	ingrFake := (integrationtest.FakeIngress{
		Name:        ingressName,
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/foo"},
		ServiceName: svcName,
	}).Ingress()
	ingrFake.ResourceVersion = "1"

	_, err := KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	integrationtest.PollForCompletion(t, modelName, 5)
	ingrFake = (integrationtest.FakeIngress{
		Name:        ingressName,
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/foo", "/bar"},
		ServiceName: svcName,
	}).IngressMultiPath()
	ingrFake.ResourceVersion = "2"
	_, err = KubeClient.NetworkingV1().Ingresses("default").Update(context.TODO(), ingrFake, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	integrationtest.DetectModelChecksumChange(t, modelName, 5)
	g.Eventually(func() int {
		if found, aviModel := objects.SharedAviGraphLister().Get(modelName); found {
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
			return len(nodes[0].EvhNodes)
		}
		return 0
	}, 15*time.Second).Should(gomega.Equal(1))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(len(nodes)).To(gomega.Equal(1))
	g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
	g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
	g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(0))
	g.Expect(len(nodes[0].EvhNodes)).To(gomega.Equal(1))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs).Should(gomega.HaveLen(1))
	g.Expect(len(nodes[0].EvhNodes[0].HttpPolicyRefs), gomega.Equal(1))
	g.Expect(len(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap), gomega.Equal(2))
	g.Expect(len(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[0].Path), gomega.Equal(1))
	g.Expect(len(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[1].Path), gomega.Equal(1))
	g.Expect(func() []string {
		p := []string{
			nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[0].Path[0],
			nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[1].Path[0]}
		sort.Strings(p)
		return p
	}, gomega.Equal([]string{"/bar", "/foo"}))
	g.Expect(func() []string {
		p := []string{
			nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[0].Name,
			nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[1].Name}
		sort.Strings(p)
		return p
	}, gomega.Equal([]string{lib.Encode("cluster--default-foo.com_bar-"+ingressName, lib.HPPMAP),
		lib.Encode("cluster--default-foo.com_foo-"+ingressName, lib.HPPMAP)}))
	g.Expect(len(nodes[0].EvhNodes[0].PoolGroupRefs)).To(gomega.Equal(2))
	g.Expect(len(nodes[0].PoolGroupRefs)).To(gomega.Equal(0))

	ingrFake = (integrationtest.FakeIngress{
		Name:        ingressName,
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/foo", "/foobar"},
		ServiceName: svcName,
	}).IngressMultiPath()
	ingrFake.ResourceVersion = "3"
	objects.SharedAviGraphLister().Delete(modelName)
	_, err = KubeClient.NetworkingV1().Ingresses("default").Update(context.TODO(), ingrFake, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	integrationtest.DetectModelChecksumChange(t, modelName, 5)
	found, aviModel := objects.SharedAviGraphLister().Get(modelName)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(0))
		g.Expect(len(nodes[0].EvhNodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs).Should(gomega.HaveLen(1))
		g.Expect(len(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[0].Path), gomega.Equal(1))
		g.Expect(len(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[1].Path), gomega.Equal(1))
		g.Expect(func() []string {
			p := []string{
				nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[0].Path[0],
				nodes[0].EvhNodes[0].HttpPolicyRefs[1].HppMap[0].Path[0]}
			sort.Strings(p)
			return p
		}, gomega.Equal([]string{"/foo", "/foobar"}))
		g.Expect(func() []string {
			p := []string{
				nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[0].Name,
				nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[1].Name}
			sort.Strings(p)
			return p
		}, gomega.Equal([]string{lib.Encode("cluster--default-foo.com_foo-"+ingressName, lib.HPPMAP),
			lib.Encode("cluster--default-foo.com_foobar-"+ingressName, lib.HPPMAP)}))
		g.Expect(len(nodes[0].EvhNodes[0].PoolGroupRefs)).To(gomega.Equal(2))
		g.Expect(len(nodes[0].PoolGroupRefs)).To(gomega.Equal(0))
	} else {
		t.Fatalf("Could not find model: %s", modelName)
	}

	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), ingressName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyEvhPoolDeletion(t, g, aviModel, 0)
	VerifyEvhIngressDeletion(t, g, aviModel, 0)
	VerifyEvhVsCacheChildDeletion(t, g, cache.NamespaceName{Namespace: "admin", Name: modelName})
	TearDownTestForIngress(t, svcName, modelName)
}

func TestEditMultiIngressSameHostForEvh(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName, _ := GetModelName("foo.com", "default")
	svcName := objNameMap.GenerateName("avisvc")
	ingressName1 := objNameMap.GenerateName("foo-with-targets")
	ingressName2 := objNameMap.GenerateName("foo-with-targets")
	SetUpTestForIngress(t, svcName, modelName)

	ingrFake1 := (integrationtest.FakeIngress{
		Name:        ingressName1,
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/foo"},
		ServiceName: svcName,
	}).Ingress()

	_, err := integrationtest.KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake1, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	ingrFake2 := (integrationtest.FakeIngress{
		Name:        ingressName2,
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/bar"},
		ServiceName: svcName,
	}).Ingress()

	_, err = integrationtest.KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake2, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	ingrFake2 = (integrationtest.FakeIngress{
		Name:        ingressName2,
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/foobar"},
		ServiceName: svcName,
	}).Ingress()

	_, err = integrationtest.KubeClient.NetworkingV1().Ingresses("default").Update(context.TODO(), ingrFake2, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	integrationtest.PollForCompletion(t, modelName, 5)
	integrationtest.DetectModelChecksumChange(t, modelName, 15)
	g.Eventually(func() int {
		if found, aviModel := objects.SharedAviGraphLister().Get(modelName); found {
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
			return len(nodes[0].EvhNodes)
		}
		return 0
	}, 15*time.Second).Should(gomega.Equal(1))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(len(nodes)).To(gomega.Equal(1))
	g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
	g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
	g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(0))

	g.Expect(len(nodes[0].EvhNodes)).To(gomega.Equal(1))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs).Should(gomega.HaveLen(1))
	g.Expect(len(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[0].Path), gomega.Equal(1))
	g.Expect(len(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[1].Path), gomega.Equal(1))
	g.Expect(func() []string {
		p := []string{
			nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[0].Path[0],
			nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[1].Path[0]}
		sort.Strings(p)
		return p
	}, gomega.Equal([]string{"/foo", "/foobar"}))
	g.Expect(func() []string {
		p := []string{
			nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[0].Name,
			nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[1].Name}
		sort.Strings(p)
		return p
	}, gomega.Equal([]string{lib.Encode("cluster--default-foo.com_foo-"+ingressName1, lib.HPPMAP),
		lib.Encode("cluster--default-foo.com_foobar-"+ingressName2, lib.HPPMAP)}))
	g.Expect(len(nodes[0].EvhNodes[0].PoolGroupRefs)).To(gomega.Equal(2))
	g.Expect(len(nodes[0].PoolGroupRefs)).To(gomega.Equal(0))

	err = integrationtest.KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), ingressName1, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	integrationtest.DetectModelChecksumChange(t, modelName, 5)
	VerifyEvhIngressDeletion(t, g, aviModel, 1)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(len(nodes)).To(gomega.Equal(1))

	err = integrationtest.KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), ingressName2, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	integrationtest.DetectModelChecksumChange(t, modelName, 5)
	VerifyEvhPoolDeletion(t, g, aviModel, 0)
	VerifyEvhIngressDeletion(t, g, aviModel, 0)
	VerifyEvhVsCacheChildDeletion(t, g, cache.NamespaceName{Namespace: "admin", Name: modelName})
	TearDownTestForIngress(t, svcName, modelName)
}

func TestNoHostIngressForEvh(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	ingressName := objNameMap.GenerateName("foo-with-targets")
	svcName := objNameMap.GenerateName("avisvc")
	modelName, _ := GetModelName(ingressName+".default.com", "default")

	SetUpTestForIngress(t, svcName, modelName)

	ingrFake := (integrationtest.FakeIngress{
		Name:        ingressName,
		Namespace:   "default",
		Paths:       []string{"/foo"},
		ServiceName: svcName,
	}).IngressNoHost()

	_, err := KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	integrationtest.PollForCompletion(t, modelName, 5)
	g.Eventually(func() int {
		if found, aviModel := objects.SharedAviGraphLister().Get(modelName); found {
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
			return len(nodes[0].EvhNodes)
		}
		return 0
	}, 30*time.Second).Should(gomega.Equal(1))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(len(nodes)).To(gomega.Equal(1))
	g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
	g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
	g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(0))
	g.Expect(len(nodes[0].EvhNodes)).To(gomega.Equal(1))

	g.Expect(nodes[0].EvhNodes[0].Name).To(gomega.Equal(lib.Encode("cluster--"+ingressName+".default.com", lib.EVHVS)))
	g.Expect(nodes[0].EvhNodes[0].PoolGroupRefs[0].Name).To(gomega.Equal(lib.Encode("cluster--default-"+ingressName+".default.com_foo-"+ingressName, lib.PG)))
	g.Expect(nodes[0].EvhNodes[0].PoolRefs[0].Name).To(gomega.Equal(lib.Encode("cluster--default-"+ingressName+".default.com_foo-"+ingressName+"-"+svcName, lib.Pool)))
	g.Expect(nodes[0].EvhNodes[0].EvhHostName).To(gomega.Equal(ingressName + ".default.com"))
	g.Expect(len(nodes[0].EvhNodes[0].PoolGroupRefs[0].Members)).To(gomega.Equal(1))
	g.Expect(len(nodes[0].EvhNodes[0].PoolRefs[0].Servers)).To(gomega.Equal(1))

	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), ingressName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyEvhPoolDeletion(t, g, aviModel, 0)
	VerifyEvhIngressDeletion(t, g, aviModel, 0)
	VerifyEvhVsCacheChildDeletion(t, g, cache.NamespaceName{Namespace: "admin", Name: modelName})
	TearDownTestForIngress(t, svcName, modelName)
}

func TestEditNoHostToHostIngressForEvh(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	ingressName := objNameMap.GenerateName("foo-with-targets")
	svcName := objNameMap.GenerateName("avisvc")
	modelName, _ := GetModelName(ingressName+".default.com", "default")
	SetUpTestForIngress(t, svcName, modelName)

	ingrFake := (integrationtest.FakeIngress{
		Name:        ingressName,
		Namespace:   "default",
		Paths:       []string{"/foo"},
		ServiceName: svcName,
	}).IngressNoHost()

	_, err := KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	integrationtest.PollForCompletion(t, modelName, 5)
	g.Eventually(func() int {
		if found, aviModel := objects.SharedAviGraphLister().Get(modelName); found {
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
			return len(nodes[0].EvhNodes)
		}
		return 0
	}, 15*time.Second).Should(gomega.Equal(1))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(len(nodes)).To(gomega.Equal(1))
	g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
	g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
	g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(0))
	g.Expect(len(nodes[0].EvhNodes)).To(gomega.Equal(1))

	g.Expect(nodes[0].EvhNodes[0].Name).To(gomega.Equal(lib.Encode("cluster--"+ingressName+".default.com", lib.EVHVS)))
	g.Expect(nodes[0].EvhNodes[0].PoolGroupRefs[0].Name).To(gomega.Equal(lib.Encode("cluster--default-"+ingressName+".default.com_foo-"+ingressName, lib.PG)))
	g.Expect(nodes[0].EvhNodes[0].PoolRefs[0].Name).To(gomega.Equal(lib.Encode("cluster--default-"+ingressName+".default.com_foo-"+ingressName+"-"+svcName, lib.Pool)))
	g.Expect(nodes[0].EvhNodes[0].EvhHostName).To(gomega.Equal(ingressName + ".default.com"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs).Should(gomega.HaveLen(1))
	g.Expect(len(nodes[0].EvhNodes[0].PoolGroupRefs[0].Members)).To(gomega.Equal(1))
	g.Expect(len(nodes[0].EvhNodes[0].PoolRefs[0].Servers)).To(gomega.Equal(1))

	ingrFake = (integrationtest.FakeIngress{
		Name:        ingressName,
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/foo"},
		ServiceName: svcName,
	}).Ingress()

	ingrFake.ResourceVersion = "2"
	_, err = KubeClient.NetworkingV1().Ingresses("default").Update(context.TODO(), ingrFake, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in Updating Ingress: %v", err)
	}

	integrationtest.PollForCompletion(t, modelName, 5)
	integrationtest.DetectModelChecksumChange(t, modelName, 5)

	found, aviModel := objects.SharedAviGraphLister().Get(modelName)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(0))
		if !lib.VIPPerNamespace() {
			g.Expect(len(nodes[0].EvhNodes)).To(gomega.Equal(0))
		}

	} else {
		t.Fatalf("Could not find model: %s", modelName)
	}

	modelName, _ = GetModelName("foo.com", "default")
	integrationtest.PollForCompletion(t, modelName, 5)
	integrationtest.DetectModelChecksumChange(t, modelName, 5)

	g.Eventually(func() string {
		if found, aviModel := objects.SharedAviGraphLister().Get(modelName); found {
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
			if len(nodes[0].EvhNodes) > 0 {
				return nodes[0].EvhNodes[0].Name
			}
		}
		return ""
	}, 15*time.Second).Should(gomega.Equal(lib.Encode("cluster--foo.com", lib.EVHVS)))
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(len(nodes)).To(gomega.Equal(1))
	g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
	g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
	g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(0))
	g.Expect(len(nodes[0].EvhNodes)).To(gomega.Equal(1))
	g.Expect(nodes[0].EvhNodes[0].Name).To(gomega.Equal(lib.Encode("cluster--foo.com", lib.EVHVS)))
	g.Expect(nodes[0].EvhNodes[0].PoolGroupRefs[0].Name).To(gomega.Equal(lib.Encode("cluster--default-foo.com_foo-"+ingressName, lib.PG)))
	g.Expect(nodes[0].EvhNodes[0].PoolRefs[0].Name).To(gomega.Equal(lib.Encode("cluster--default-foo.com_foo-"+ingressName+"-"+svcName, lib.Pool)))
	g.Expect(nodes[0].EvhNodes[0].EvhHostName).To(gomega.Equal("foo.com"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs).Should(gomega.HaveLen(1))
	g.Expect(len(nodes[0].EvhNodes[0].PoolGroupRefs[0].Members)).To(gomega.Equal(1))
	g.Expect(len(nodes[0].EvhNodes[0].PoolRefs[0].Servers)).To(gomega.Equal(1))

	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), ingressName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}

	VerifyEvhPoolDeletion(t, g, aviModel, 0)
	VerifyEvhIngressDeletion(t, g, aviModel, 0)
	VerifyEvhVsCacheChildDeletion(t, g, cache.NamespaceName{Namespace: "admin", Name: modelName})
	TearDownTestForIngress(t, svcName, modelName)
}

func TestScaleEndpointsForEvh(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName, _ := GetModelName("foo.com", "default")
	svcName := objNameMap.GenerateName("avisvc")
	ingressName1 := objNameMap.GenerateName("foo-with-targets")
	ingressName2 := objNameMap.GenerateName("foo-with-targets")
	SetUpTestForIngress(t, svcName, modelName)

	ingrFake1 := (integrationtest.FakeIngress{
		Name:        ingressName1,
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/foo"},
		ServiceName: svcName,
	}).Ingress()

	_, err := KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake1, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	ingrFake2 := (integrationtest.FakeIngress{
		Name:        ingressName2,
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/bar"},
		ServiceName: svcName,
	}).Ingress()

	_, err = KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake2, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	integrationtest.PollForCompletion(t, modelName, 5)
	g.Eventually(func() int {
		if found, aviModel := objects.SharedAviGraphLister().Get(modelName); found {
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
			return len(nodes[0].EvhNodes)
		}
		return 0
	}, 15*time.Second).Should(gomega.Equal(1))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(len(nodes)).To(gomega.Equal(1))
	g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
	g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
	g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(0))
	g.Expect(len(nodes[0].EvhNodes)).To(gomega.Equal(1))
	g.Expect(len(nodes[0].PoolGroupRefs)).To(gomega.Equal(0))
	g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(0))
	g.Expect(len(nodes[0].EvhNodes[0].PoolGroupRefs[0].Members)).To(gomega.Equal(1))
	g.Expect(len(nodes[0].EvhNodes[0].PoolRefs[0].Servers)).To(gomega.Equal(1))
	g.Expect(len(nodes[0].EvhNodes[0].PoolGroupRefs[1].Members)).To(gomega.Equal(1))
	g.Expect(len(nodes[0].EvhNodes[0].PoolRefs[1].Servers)).To(gomega.Equal(1))

	integrationtest.ScaleCreateEPorEPS(t, "default", svcName)
	integrationtest.PollForCompletion(t, modelName, 5)
	integrationtest.DetectModelChecksumChange(t, modelName, 5)

	g.Eventually(func() bool {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS(); found && len(nodes) > 0 {
			return len(nodes[0].EvhNodes[0].PoolRefs[0].Servers) == 2
		}
		return false
	}, 10*time.Second).Should(gomega.Equal(true))
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(len(nodes)).To(gomega.Equal(1))
	g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
	g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
	g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(0))
	g.Expect(len(nodes[0].EvhNodes)).To(gomega.Equal(1))
	g.Expect(len(nodes[0].PoolGroupRefs)).To(gomega.Equal(0))
	g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(0))
	g.Expect(len(nodes[0].EvhNodes[0].PoolGroupRefs[0].Members)).To(gomega.Equal(1))
	// Count should be 2 for both the backend pool members of the pool after the scaleout
	g.Expect(len(nodes[0].EvhNodes[0].PoolRefs[0].Servers)).To(gomega.Equal(2))
	g.Expect(len(nodes[0].EvhNodes[0].PoolGroupRefs[1].Members)).To(gomega.Equal(1))
	g.Expect(len(nodes[0].EvhNodes[0].PoolRefs[1].Servers)).To(gomega.Equal(2))

	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), ingressName1, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyEvhPoolDeletion(t, g, aviModel, 0)
	VerifyEvhIngressDeletion(t, g, aviModel, 1)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(len(nodes)).To(gomega.Equal(1))

	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), ingressName2, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyEvhPoolDeletion(t, g, aviModel, 0)
	VerifyEvhIngressDeletion(t, g, aviModel, 0)
	VerifyEvhVsCacheChildDeletion(t, g, cache.NamespaceName{Namespace: "admin", Name: modelName})
	TearDownTestForIngress(t, svcName, modelName)
}

// Additional EVH test cases follow:

func TestL7ModelNoSecretToSecretForEvh(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	modelName, _ := GetModelName("foo.com", "default")
	secretName := objNameMap.GenerateName("my-secret")
	ingressName := objNameMap.GenerateName("foo-with-targets")
	svcName := objNameMap.GenerateName("avisvc")
	SetUpTestForIngress(t, svcName, modelName)

	integrationtest.PollForCompletion(t, modelName, 5)
	ingrFake := (integrationtest.FakeIngress{
		Name:        ingressName,
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		ServiceName: svcName,
		TlsSecretDNS: map[string][]string{
			secretName: {"foo.com"},
		},
	}).Ingress()
	_, err := KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	integrationtest.PollForCompletion(t, modelName, 5)
	found, aviModel := objects.SharedAviGraphLister().Get(modelName)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		g.Expect(len(nodes[0].EvhNodes)).To(gomega.Equal(0))
		g.Expect(nodes[0].VHDomainNames).To(gomega.HaveLen(0))
		g.Expect(nodes[0].HttpPolicyRefs).To(gomega.HaveLen(0))
	} else {
		t.Fatalf("Could not find Model: %v", err)
	}

	// Now create the secret and verify the models.
	integrationtest.AddSecret(secretName, "default", "tlsCert", "tlsKey")
	found, aviModel = objects.SharedAviGraphLister().Get(modelName)
	if found {
		g.Eventually(func() int {
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
			return len(nodes[0].EvhNodes)
		}, 10*time.Second).Should(gomega.Equal(1))

	} else {
		t.Fatalf("Could not find model: %s", modelName)
	}

	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), ingressName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	KubeClient.CoreV1().Secrets("default").Delete(context.TODO(), secretName, metav1.DeleteOptions{})
	VerifyEvhPoolDeletion(t, g, aviModel, 0)
	VerifyEvhIngressDeletion(t, g, aviModel, 0)
	VerifyEvhVsCacheChildDeletion(t, g, cache.NamespaceName{Namespace: "admin", Name: modelName})
	TearDownTestForIngress(t, svcName, modelName)
}

func TestL7ModelOneSecretToMultiIngForEvh(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	modelName, _ := GetModelName("foo.com", "default")
	secretName := objNameMap.GenerateName("my-secret")
	svcName := objNameMap.GenerateName("avisvc")
	ingressName1 := objNameMap.GenerateName("foo-with-targets")
	ingressName2 := objNameMap.GenerateName("foo-with-targets")
	SetUpTestForIngress(t, svcName, modelName)

	integrationtest.PollForCompletion(t, modelName, 5)
	ingrFake1 := (integrationtest.FakeIngress{
		Name:        ingressName1,
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		ServiceName: svcName,
		TlsSecretDNS: map[string][]string{
			secretName: {"foo.com"},
		},
	}).Ingress()
	_, err := KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake1, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	ingrFake2 := (integrationtest.FakeIngress{
		Name:      ingressName2,
		Namespace: "default",
		DnsNames:  []string{"foo.com"},
		Ips:       []string{"8.8.8.8"},
		HostNames: []string{"v1"},
		TlsSecretDNS: map[string][]string{
			secretName: {"foo.com"},
		},
		ServiceName: svcName,
	}).Ingress()
	_, err = KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake2, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	integrationtest.PollForCompletion(t, modelName, 5)
	found, aviModel := objects.SharedAviGraphLister().Get(modelName)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		g.Expect(nodes[0].SSLKeyCertRefs).Should(gomega.HaveLen(0))
		g.Expect(len(nodes[0].EvhNodes)).To(gomega.Equal(0))
	} else {
		t.Fatalf("Could not find Model: %v", err)
	}

	// Now create the secret and verify the models.
	integrationtest.AddSecret(secretName, "default", "tlsCert", "tlsKey")
	found, aviModel = objects.SharedAviGraphLister().Get(modelName)
	if found {
		// Check if the secret affected both the models.
		g.Eventually(func() int {
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
			return len(nodes[0].EvhNodes)
		}, 30*time.Second).Should(gomega.Equal(1))
	} else {
		t.Fatalf("Could not find model: %s", modelName)
	}
	time.Sleep(10 * time.Second)
	KubeClient.CoreV1().Secrets("default").Delete(context.TODO(), secretName, metav1.DeleteOptions{})

	integrationtest.PollForCompletion(t, modelName, 5)
	integrationtest.DetectModelChecksumChange(t, modelName, 5)

	VerifyEvhIngressDeletion(t, g, aviModel, 0)
	// Since we deleted the secret, both EVH should get removed.
	found, aviModel = objects.SharedAviGraphLister().Get(modelName)
	if found {
		// Check if the secret affected both the models.
		g.Eventually(func() int {
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
			return len(nodes[0].EvhNodes)
		}, 10*time.Second).Should(gomega.Equal(0))

	} else {
		t.Fatalf("Could not find model: %s", modelName)
	}
	// check if the certificate is deleted
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(nodes[0].SSLKeyCertRefs).Should(gomega.HaveLen(0))

	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), ingressName1, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), ingressName2, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyEvhVsCacheChildDeletion(t, g, cache.NamespaceName{Namespace: "admin", Name: modelName})
	TearDownTestForIngress(t, svcName, modelName)
}

func TestL7ModelMultiSNIForEvh(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	secretName := objNameMap.GenerateName("my-secret")
	ingressName := objNameMap.GenerateName("foo-with-targets")
	svcName := objNameMap.GenerateName("avisvc")
	integrationtest.AddSecret(secretName, "default", "tlsCert", "tlsKey")
	modelName, _ := GetModelName("foo.com", "default")
	SetUpTestForIngress(t, svcName, modelName)

	ingrFake := (integrationtest.FakeIngress{
		Name:        ingressName,
		Namespace:   "default",
		DnsNames:    []string{"foo.com", "noo.com"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		ServiceName: svcName,
		TlsSecretDNS: map[string][]string{
			secretName: {"foo.com", "noo.com"},
		},
	}).Ingress()
	_, err := KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	integrationtest.PollForCompletion(t, modelName, 5)
	g.Eventually(func() int {
		if found, aviModel := objects.SharedAviGraphLister().Get(modelName); found {
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
			return len(nodes[0].EvhNodes)
		}
		return 0
	}, 15*time.Second).Should(gomega.Equal(2))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(len(nodes)).To(gomega.Equal(1))
	g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
	g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
	g.Expect(nodes[0].HttpPolicyRefs).To(gomega.HaveLen(0))
	g.Expect(len(nodes[0].EvhNodes)).To(gomega.Equal(2))
	g.Expect(len(nodes[0].EvhNodes[0].PoolGroupRefs)).To(gomega.Equal(1))
	g.Expect(len(nodes[0].EvhNodes[0].HttpPolicyRefs)).To(gomega.Equal(2))
	g.Expect(len(nodes[0].EvhNodes[0].PoolRefs)).To(gomega.Equal(1))
	g.Expect(len(nodes[0].EvhNodes[0].PoolRefs[0].Servers)).To(gomega.Equal(1))
	g.Expect(len(nodes[0].EvhNodes[0].SSLKeyCertRefs)).To(gomega.Equal(0))

	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), ingressName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	KubeClient.CoreV1().Secrets("default").Delete(context.TODO(), secretName, metav1.DeleteOptions{})
	VerifyEvhIngressDeletion(t, g, aviModel, 0)
	VerifyEvhVsCacheChildDeletion(t, g, cache.NamespaceName{Namespace: "admin", Name: modelName})
	TearDownTestForIngress(t, svcName, modelName)
}

func TestL7ModelMultiSNIMultiCreateEditSecretForEvh(t *testing.T) {
	// This test covers creating multiple SNI nodes via multiple secrets.

	g := gomega.NewGomegaWithT(t)
	secretName := objNameMap.GenerateName("my-secret")
	ingressName := objNameMap.GenerateName("foo-with-targets")
	svcName := objNameMap.GenerateName("avisvc")
	integrationtest.AddSecret(secretName, "default", "tlsCert", "tlsKey")
	integrationtest.AddSecret("my-secret2", "default", "tlsCert", "tlsKey")
	// Clean up any earlier models.
	modelName, _ := GetModelName("foo.com", "default")
	objects.SharedAviGraphLister().Delete(modelName)
	modelName, _ = GetModelName("foo.com", "default")
	objects.SharedAviGraphLister().Delete(modelName)
	SetUpTestForIngress(t, svcName, modelName)

	ingrFake := (integrationtest.FakeIngress{
		Name:        ingressName,
		Namespace:   "default",
		DnsNames:    []string{"foo.com", "FOO.com"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		ServiceName: svcName,
		TlsSecretDNS: map[string][]string{
			secretName:   {"foo.com"},
			"my-secret2": {"FOO.com"},
		},
	}).Ingress()

	_, err := KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	integrationtest.PollForCompletion(t, modelName, 5)
	g.Eventually(func() int {
		if found, aviModel := objects.SharedAviGraphLister().Get(modelName); found {
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
			return len(nodes[0].EvhNodes)
		}
		return 0
	}, 15*time.Second).Should(gomega.Equal(2))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(len(nodes)).To(gomega.Equal(1))
	g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
	g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
	g.Expect(nodes[0].HttpPolicyRefs).To(gomega.HaveLen(0))
	g.Expect(len(nodes[0].EvhNodes)).To(gomega.Equal(2))
	g.Expect(len(nodes[0].EvhNodes[0].PoolGroupRefs)).To(gomega.Equal(1))
	g.Expect(len(nodes[0].EvhNodes[0].HttpPolicyRefs)).To(gomega.Equal(2))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[1].RedirectPorts[0].Hosts).To(gomega.HaveLen(1))
	g.Expect(len(nodes[0].EvhNodes[0].PoolRefs)).To(gomega.Equal(1))
	g.Expect(len(nodes[0].EvhNodes[0].PoolRefs[0].Servers)).To(gomega.Equal(1))
	g.Expect(len(nodes[0].EvhNodes[0].SSLKeyCertRefs)).To(gomega.Equal(0))
	g.Expect(nodes[0].EvhNodes[0].VHDomainNames).To(gomega.HaveLen(1))

	ingrFake = (integrationtest.FakeIngress{
		Name:        ingressName,
		Namespace:   "default",
		DnsNames:    []string{"foo.com", "bar.com"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		ServiceName: svcName,
		TlsSecretDNS: map[string][]string{
			secretName:   {"foo.com"},
			"my-secret2": {"bar.com"},
		},
	}).Ingress()
	ingrFake.ResourceVersion = "2"
	_, err = KubeClient.NetworkingV1().Ingresses("default").Update(context.TODO(), ingrFake, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating Ingress: %v", err)
	}

	// Because of change of the hostnames, the SNI nodes should now get distributed to two shared VSes.
	found, aviModel := objects.SharedAviGraphLister().Get(modelName)
	evhNodesLen := 1
	if lib.VIPPerNamespace() {
		evhNodesLen = 2
	}
	if found {
		// Check if the secret affected both the models.
		g.Eventually(func() int {
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
			return len(nodes[0].EvhNodes)
		}, 20*time.Second).Should(gomega.Equal(evhNodesLen))
	} else {
		t.Fatalf("Could not find model: %s", modelName)
	}
	modelName, _ = GetModelName("foo.com", "default")
	found, aviModel = objects.SharedAviGraphLister().Get(modelName)
	if found {
		g.Eventually(func() int {
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
			return len(nodes[0].EvhNodes)
		}, 10*time.Second).Should(gomega.Equal(evhNodesLen))
	} else {
		t.Fatalf("Could not find model: %s", modelName)
	}
	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), ingressName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}

	KubeClient.CoreV1().Secrets("default").Delete(context.TODO(), secretName, metav1.DeleteOptions{})
	KubeClient.CoreV1().Secrets("default").Delete(context.TODO(), "my-secret2", metav1.DeleteOptions{})
	VerifyEvhIngressDeletion(t, g, aviModel, 0)
	VerifyEvhVsCacheChildDeletion(t, g, cache.NamespaceName{Namespace: "admin", Name: "cluster--Shared-L7-EVH-1"})
	TearDownTestForIngress(t, svcName, modelName)
	VerifyEvhVsCacheChildDeletion(t, g, cache.NamespaceName{Namespace: "admin", Name: "cluster--Shared-L7-EVH-0"})
	VerifyEvhVsCacheChildDeletion(t, g, cache.NamespaceName{Namespace: "admin", Name: modelName})
}

func TestL7WrongSubDomainMultiSNIForEvh(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	secretName := objNameMap.GenerateName("my-secret")
	ingressName := objNameMap.GenerateName("foo-with-targets")
	svcName := objNameMap.GenerateName("avisvc")
	integrationtest.AddSecret(secretName, "default", "tlsCert", "tlsKey")
	integrationtest.AddSecret("my-secret2", "default", "tlsCert", "tlsKey")
	modelName, _ := GetModelName("foo.com", "default")
	SetUpTestForIngress(t, svcName, integrationtest.AllModels...)

	ingrFake := (integrationtest.FakeIngress{
		Name:        ingressName,
		Namespace:   "default",
		DnsNames:    []string{"foo.org"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		ServiceName: svcName,
		TlsSecretDNS: map[string][]string{
			secretName: {"foo.org"},
		},
	}).Ingress()
	_, err := KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	ingrFake = (integrationtest.FakeIngress{
		Name:        ingressName,
		Namespace:   "default",
		DnsNames:    []string{"foo.org", "bar.com"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		ServiceName: svcName,
		TlsSecretDNS: map[string][]string{
			secretName:   {"foo.org"},
			"my-secret2": {"bar.com"},
		},
	}).Ingress()
	ingrFake.ResourceVersion = "2"
	_, err = KubeClient.NetworkingV1().Ingresses("default").Update(context.TODO(), ingrFake, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("Couldn't update the Ingress %v", err)
	}

	modelName, _ = GetModelName("bar.com", "default")
	integrationtest.PollForCompletion(t, modelName, 5)
	g.Eventually(func() bool {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if found {
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
			return len(nodes) == 1 && len(nodes[0].EvhNodes) == 1
		}
		return false
	}, 40*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
	g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
	g.Expect(len(nodes[0].EvhNodes[0].PoolGroupRefs)).To(gomega.Equal(1))
	g.Expect(len(nodes[0].EvhNodes[0].HttpPolicyRefs)).To(gomega.Equal(2))
	g.Expect(len(nodes[0].EvhNodes[0].PoolRefs)).To(gomega.Equal(1))
	g.Expect(len(nodes[0].EvhNodes[0].PoolRefs[0].Servers)).To(gomega.Equal(1))
	g.Expect(len(nodes[0].EvhNodes[0].SSLKeyCertRefs)).To(gomega.Equal(0))
	g.Expect(nodes[0].EvhNodes[0].VHDomainNames).To(gomega.HaveLen(1))

	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), ingressName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	KubeClient.CoreV1().Secrets("default").Delete(context.TODO(), secretName, metav1.DeleteOptions{})
	VerifyEvhIngressDeletion(t, g, aviModel, 0)
	VerifyEvhVsCacheChildDeletion(t, g, cache.NamespaceName{Namespace: "admin", Name: modelName})
	TearDownTestForIngress(t, svcName, modelName)
}

func TestFQDNCountInL7Model(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	modelName, _ := GetModelName("foo.com", "default")
	secretName := objNameMap.GenerateName("my-secret")
	ingressName := objNameMap.GenerateName("foo-with-targets")
	svcName := objNameMap.GenerateName("avisvc")
	ingTestObj := IngressTestObject{
		ingressName: ingressName,
		isTLS:       true,
		withSecret:  true,
		secretName:  secretName,
		serviceName: svcName,
		modelNames:  []string{modelName},
	}
	ingTestObj.FillParams()
	SetUpIngressForCacheSyncCheck(t, ingTestObj)
	g.Eventually(func() int {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if aviModel == nil {
			return 0
		}
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return len(nodes)
	}, 10*time.Second).Should(gomega.Equal(1))

	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	node := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()[0]

	fqdnCount := 2
	if lib.VIPPerNamespace() {
		fqdnCount = 1
	}
	g.Expect(node.VSVIPRefs).To(gomega.HaveLen(1))
	g.Expect(node.VSVIPRefs[0].FQDNs).To(gomega.HaveLen(fqdnCount))
	for _, fqdn := range node.VSVIPRefs[0].FQDNs {
		if fqdn == "foo.com" {
			continue
		}
		if !lib.VIPPerNamespace() {
			g.Expect(fqdn).Should(gomega.ContainSubstring("Shared-L7"))
		}
	}

	TearDownIngressForCacheSyncCheck(t, secretName, ingressName, svcName, modelName)
}

func TestPortsForInsecureAndSecureEVH(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	modelName, _ := GetModelName("foo.com", "default")
	secretName := objNameMap.GenerateName("my-secret")
	ingressName := objNameMap.GenerateName("foo-with-targets")
	svcName := objNameMap.GenerateName("avisvc")
	SetUpTestForIngress(t, svcName, modelName)

	// Insecure
	integrationtest.PollForCompletion(t, modelName, 5)
	ingrFake := (integrationtest.FakeIngress{
		Name:        ingressName,
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		ServiceName: svcName,
		TlsSecretDNS: map[string][]string{
			secretName: {"foo.com"},
		},
	}).Ingress()
	_, err := KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	integrationtest.PollForCompletion(t, modelName, 5)
	found, aviModel := objects.SharedAviGraphLister().Get(modelName)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		g.Expect(len(nodes[0].PortProto)).To(gomega.Equal(2))
		var ports []int
		for _, port := range nodes[0].PortProto {
			ports = append(ports, int(port.Port))
			if port.EnableSSL {
				g.Expect(int(port.Port)).To(gomega.Equal(443))
			}
		}
		sort.Ints(ports)
		g.Expect(ports[0]).To(gomega.Equal(80))
		g.Expect(ports[1]).To(gomega.Equal(443))
	}

	// Secure
	integrationtest.AddSecret(secretName, "default", "tlsCert", "tlsKey")
	found, aviModel = objects.SharedAviGraphLister().Get(modelName)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		g.Expect(len(nodes[0].PortProto)).To(gomega.Equal(2))
		var ports []int
		for _, port := range nodes[0].PortProto {
			ports = append(ports, int(port.Port))
			if port.EnableSSL {
				g.Expect(int(port.Port)).To(gomega.Equal(443))
			}
		}
		sort.Ints(ports)
		g.Expect(ports[0]).To(gomega.Equal(80))
		g.Expect(ports[1]).To(gomega.Equal(443))
	} else {
		t.Fatalf("Could not find model: %s", modelName)
	}

	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), ingressName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	KubeClient.CoreV1().Secrets("default").Delete(context.TODO(), secretName, metav1.DeleteOptions{})
	TearDownTestForIngress(t, svcName, modelName)
}

func TestMultiIngressSameHostDifferentNamespaceForEvh(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	lib.AKOControlConfig().SetAKOFQDNReusePolicy("strict")
	modelName, _ := GetModelName("foo.com", "default")
	svcName := objNameMap.GenerateName("avisvc")
	SetUpTestForIngress(t, svcName, modelName)

	ingrFake1 := (integrationtest.FakeIngress{
		Name:        "ingress-multi1",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/foo"},
		ServiceName: svcName,
	}).Ingress()

	_, err := KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake1, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Error in adding Ingress: %v", err)
	}

	// red namespace
	ingrFake2 := (integrationtest.FakeIngress{
		Name:        "ingress-multi2",
		Namespace:   "red",
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/bar"},
		ServiceName: svcName,
	}).Ingress()

	ing_red, err := KubeClient.NetworkingV1().Ingresses("red").Create(context.TODO(), ingrFake2, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Error in adding Ingress: %v", err)
	}
	// status should be empty.
	// TODO: Events can be checked for ingress.
	g.Expect(len(ing_red.Status.LoadBalancer.Ingress)).To(gomega.Equal(0))
	integrationtest.PollForCompletion(t, modelName, 5)
	found, aviModel := objects.SharedAviGraphLister().Get(modelName)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(0))

		g.Eventually(func() int {
			_, aviModel := objects.SharedAviGraphLister().Get(modelName)
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
			return len(nodes[0].EvhNodes)
		}, 10*time.Second).Should(gomega.Equal(1))
		g.Expect(nodes[0].EvhNodes[0].Name).To(gomega.Equal(lib.Encode("cluster--foo.com", lib.EVHVS)))
		g.Expect(nodes[0].EvhNodes[0].EvhHostName).To(gomega.Equal("foo.com"))
		g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs).Should(gomega.HaveLen(1))
		g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap).Should(gomega.HaveLen(1))
		g.Expect(len(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[0].Path), gomega.Equal(1))
		g.Expect(func() []string {
			p := []string{
				nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[0].Path[0],
			}
			sort.Strings(p)
			return p
		}, gomega.Equal([]string{"/foo"}))
		g.Expect(len(nodes[0].EvhNodes[0].PoolGroupRefs)).To(gomega.Equal(1))
	} else {
		t.Fatalf("Could not find model: %s", modelName)
	}
	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), "ingress-multi1", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyEvhPoolDeletion(t, g, aviModel, 0)
	VerifyEvhIngressDeletion(t, g, aviModel, 1)
	integrationtest.DetectModelChecksumChange(t, modelName, 5)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(len(nodes)).To(gomega.Equal(1))
	g.Expect(len(nodes[0].EvhNodes)).To(gomega.Equal(0))

	err = KubeClient.NetworkingV1().Ingresses("red").Delete(context.TODO(), "ingress-multi2", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	lib.AKOControlConfig().SetAKOFQDNReusePolicy("internamespaceallowed")
	VerifyEvhPoolDeletion(t, g, aviModel, 0)
	VerifyEvhIngressDeletion(t, g, aviModel, 0)
	VerifyEvhVsCacheChildDeletion(t, g, cache.NamespaceName{Namespace: "admin", Name: modelName})
	TearDownTestForIngress(t, svcName, modelName)
}
