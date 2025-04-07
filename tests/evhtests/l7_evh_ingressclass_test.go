/*
 * Copyright 2020-2021 VMware, Inc.
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
	"sync"
	"testing"
	"time"

	"github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/cache"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/k8s"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/nodes"
	avinodes "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/nodes"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"
	utils "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/tests/integrationtest"
)

func VerifyEvhNodeDeletionFromVsNode(g *gomega.WithT, modelName string, parentVSKey, evhVsKey cache.NamespaceName) {
	g.Eventually(func() bool {
		if found, aviModel := objects.SharedAviGraphLister().Get(modelName); found {
			if nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS(); len(nodes) > 0 {
				return len(nodes[0].EvhNodes) == 0
			}
		}
		return true
	}, 50*time.Second).Should(gomega.Equal(true))

	if !lib.VIPPerNamespace() {
		g.Eventually(func() bool {
			mcache := cache.SharedAviObjCache()
			vsCache, found := mcache.VsCacheMeta.AviCacheGet(parentVSKey)
			if vsCacheObj, _ := vsCache.(*cache.AviVsCache); found {
				return len(vsCacheObj.SNIChildCollection) == 0
			}
			return false
		}, 30*time.Second).Should(gomega.Equal(true))

		g.Eventually(func() bool {
			mcache := cache.SharedAviObjCache()
			_, found := mcache.VsCacheMeta.AviCacheGet(evhVsKey)
			return found
		}, 30*time.Second).Should(gomega.Equal(false))
	}
}

func syncFromIngestionLayerWrapper(key interface{}, wg *sync.WaitGroup) error {
	keyStr, ok := key.(string)
	if !ok {
		utils.AviLog.Warnf("Unexpected object type: expected string, got %T", key)
		return nil
	}
	objType, _, name := lib.ExtractTypeNameNamespace(keyStr)
	if objType == utils.IngressClass {
		keyChan <- name
	}
	nodes.DequeueIngestion(keyStr, false)
	return nil
}

func waitAndVerify(t *testing.T, key string, opt ...string) {
	select {
	case data := <-keyChan:
		if data != key {
			t.Fatalf("error in match expected: %v, got: %v", key, data)
		}
	case <-time.After(20 * time.Second):
		t.Fatalf("timed out waiting for %v, %v", key, opt)
	}
}

// Ingress - IngressClass mapping tests

func TestEVHWrongClassMappingInIngress(t *testing.T) {
	// create ingclass, ingress
	// update wrong mapping of class in ingress, VS deleted
	// fix class in ingress, VS created
	g := gomega.NewGomegaWithT(t)

	ingClassName := objNameMap.GenerateName("avi-lb")
	ingressName := objNameMap.GenerateName("foo-with-class")
	ns := "default"
	svcName := objNameMap.GenerateName("avisvc")
	modelName, _ := GetModelName("bar.com", "default")
	vsKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--Shared-L7-EVH-1"}
	evhKey := cache.NamespaceName{Namespace: "admin", Name: lib.Encode("cluster--bar.com", lib.EVHVS)}

	// SyncFunc is replaced with a wrapper to make sure that ingressClass
	// is processed first and then ingress.
	time.Sleep(time.Second * 5)
	ingestionQueue := utils.SharedWorkQueue().GetQueueByName(utils.ObjectIngestionLayer)
	ingestionQueue.SyncFunc = syncFromIngestionLayerWrapper
	defer func() { ingestionQueue.SyncFunc = k8s.SyncFromIngestionLayer }()

	SetUpTestForIngress(t, svcName, modelName)
	integrationtest.RemoveDefaultIngressClass()
	defer integrationtest.AddDefaultIngressClass()
	waitAndVerify(t, integrationtest.DefaultIngressClass)

	integrationtest.SetupIngressClass(t, ingClassName, lib.AviIngressController, "")
	waitAndVerify(t, ingClassName)
	ingressCreate := (integrationtest.FakeIngress{
		Name:        ingressName,
		Namespace:   ns,
		ClassName:   ingClassName,
		DnsNames:    []string{"bar.com"},
		ServiceName: svcName,
	}).Ingress()
	_, err := KubeClient.NetworkingV1().Ingresses(ns).Create(context.TODO(), ingressCreate, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 40*time.Second).Should(gomega.Equal(true))

	g.Eventually(func() int {
		ingress, _ := KubeClient.NetworkingV1().Ingresses(ns).Get(context.TODO(), ingressName, metav1.GetOptions{})
		return len(ingress.Status.LoadBalancer.Ingress)
	}, 40*time.Second).Should(gomega.Equal(1))

	ingressUpdate := (integrationtest.FakeIngress{
		Name:        ingressName,
		Namespace:   ns,
		ClassName:   "xyz",
		DnsNames:    []string{"bar.com"},
		ServiceName: svcName,
	}).Ingress()
	ingressUpdate.ResourceVersion = "2"
	if _, err := KubeClient.NetworkingV1().Ingresses(ns).Update(context.TODO(), ingressUpdate, metav1.UpdateOptions{}); err != nil {
		t.Fatalf("error in updating Ingress: %v", err)
	}

	VerifyEvhNodeDeletionFromVsNode(g, modelName, vsKey, evhKey)

	g.Eventually(func() int {
		ingress, _ := KubeClient.NetworkingV1().Ingresses(ns).Get(context.TODO(), ingressName, metav1.GetOptions{})
		return len(ingress.Status.LoadBalancer.Ingress)
	}, 40*time.Second).Should(gomega.Equal(0))

	ingressUpdate2 := (integrationtest.FakeIngress{
		Name:        ingressName,
		Namespace:   ns,
		ClassName:   ingClassName,
		DnsNames:    []string{"bar.com"},
		ServiceName: svcName,
	}).Ingress()
	ingressUpdate.ResourceVersion = "3"
	if _, err := KubeClient.NetworkingV1().Ingresses(ns).Update(context.TODO(), ingressUpdate2, metav1.UpdateOptions{}); err != nil {
		t.Fatalf("error in updating Ingress: %v", err)
	}

	// vsNode must come back up
	g.Eventually(func() int {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS(); len(nodes) > 0 {
			return len(nodes[0].EvhNodes)
		}
		return 0
	}, 60*time.Second).Should(gomega.Equal(1))

	g.Eventually(func() int {
		ingress, _ := KubeClient.NetworkingV1().Ingresses(ns).Get(context.TODO(), ingressName, metav1.GetOptions{})
		return len(ingress.Status.LoadBalancer.Ingress)
	}, 50*time.Second).Should(gomega.Equal(1))

	err = KubeClient.NetworkingV1().Ingresses(ns).Delete(context.TODO(), ingressName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	TearDownTestForIngress(t, svcName, modelName)
	integrationtest.TeardownIngressClass(t, ingClassName)
	waitAndVerify(t, ingClassName)
	VerifyEvhNodeDeletionFromVsNode(g, modelName, vsKey, evhKey)
}

func TestEVHDefaultIngressClassChange(t *testing.T) {
	// use default ingress class, change default annotation to false
	// check that ingress status is removed
	// change back default class annotation to true
	// ingress status IP comes back
	g := gomega.NewGomegaWithT(t)

	ingClassName := objNameMap.GenerateName("avi-lb")
	ingressName := objNameMap.GenerateName("foo-with-class")
	ns := "default"
	svcName := objNameMap.GenerateName("avisvc")
	modelName, vsName := GetModelName("bar.com", "default")
	vsKey := cache.NamespaceName{Namespace: "admin", Name: vsName}
	evhKey := cache.NamespaceName{Namespace: "admin", Name: lib.Encode("cluster--bar.com", lib.EVHVS)}

	mcache := cache.SharedAviObjCache()
	mcache.VsCacheMeta.AviCacheDelete(vsKey)

	integrationtest.RemoveDefaultIngressClass()
	defer integrationtest.AddDefaultIngressClass()

	time.Sleep(time.Second * 5)
	ingestionQueue := utils.SharedWorkQueue().GetQueueByName(utils.ObjectIngestionLayer)
	ingestionQueue.SyncFunc = syncFromIngestionLayerWrapper
	defer func() { ingestionQueue.SyncFunc = k8s.SyncFromIngestionLayer }()

	SetUpTestForIngress(t, svcName, modelName)

	ingClass := (integrationtest.FakeIngressClass{
		Name:       ingClassName,
		Controller: lib.AviIngressController,
		Default:    true,
	}).IngressClass()
	if _, err := KubeClient.NetworkingV1().IngressClasses().Create(context.TODO(), ingClass, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding IngressClass: %v", err)
	}
	waitAndVerify(t, ingClassName)

	// ingress with no IngressClass
	ingressCreate := (integrationtest.FakeIngress{
		Name:        ingressName,
		Namespace:   ns,
		DnsNames:    []string{"bar.com"},
		ServiceName: svcName,
	}).Ingress()
	_, err := KubeClient.NetworkingV1().Ingresses(ns).Create(context.TODO(), ingressCreate, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	g.Eventually(func() int {
		ingress, _ := KubeClient.NetworkingV1().Ingresses(ns).Get(context.TODO(), ingressName, metav1.GetOptions{})
		return len(ingress.Status.LoadBalancer.Ingress)
	}, 40*time.Second).Should(gomega.Equal(1))

	ingClass.Annotations = map[string]string{lib.DefaultIngressClassAnnotation: "false"}
	ingClass.ResourceVersion = "2"
	_, err = KubeClient.NetworkingV1().IngressClasses().Update(context.TODO(), ingClass, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating IngressClass: %v", err)
	}
	waitAndVerify(t, ingClassName)

	g.Eventually(func() int {
		ingress, _ := KubeClient.NetworkingV1().Ingresses(ns).Get(context.TODO(), ingressName, metav1.GetOptions{})
		return len(ingress.Status.LoadBalancer.Ingress)
	}, 80*time.Second).Should(gomega.Equal(0))

	err = KubeClient.NetworkingV1().Ingresses(ns).Delete(context.TODO(), ingressName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	TearDownTestForIngress(t, svcName, modelName)
	integrationtest.TeardownIngressClass(t, ingClassName)
	waitAndVerify(t, ingClassName)
	VerifyEvhNodeDeletionFromVsNode(g, modelName, vsKey, evhKey)
}

// AviInfraSetting CRD
func TestEVHAviInfraSettingNamingConvention(t *testing.T) {
	if lib.VIPPerNamespace() {
		t.Skip()
	}
	// create secure and insecure host ingress, connect with infrasetting
	// check for names of all Avi objects
	g := gomega.NewGomegaWithT(t)

	ingClassName := objNameMap.GenerateName("avi-lb")
	ingressName := objNameMap.GenerateName("foo-with-class")
	ns := "default"
	settingName := objNameMap.GenerateName("my-infrasetting")

	secretName := objNameMap.GenerateName("my-secret")
	svcName := objNameMap.GenerateName("avisvc")
	modelName := "admin/cluster--Shared-L7-EVH-1"

	time.Sleep(time.Second * 5)
	ingestionQueue := utils.SharedWorkQueue().GetQueueByName(utils.ObjectIngestionLayer)
	ingestionQueue.SyncFunc = syncFromIngestionLayerWrapper
	defer func() { ingestionQueue.SyncFunc = k8s.SyncFromIngestionLayer }()

	SetUpTestForIngress(t, svcName, modelName)

	settingModelName := "admin/cluster--Shared-L7-EVH-" + settingName + "-0"
	vsKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--Shared-L7-EVH-" + settingName + "-0"}
	evhKey := cache.NamespaceName{Namespace: "admin", Name: lib.Encode("cluster--"+settingName+"-baz.com", lib.EVHVS)}
	integrationtest.SetupAviInfraSetting(t, settingName, "SMALL")
	integrationtest.SetupIngressClass(t, ingClassName, lib.AviIngressController, settingName)
	waitAndVerify(t, ingClassName)
	integrationtest.AddSecret(secretName, ns, "tlsCert", "tlsKey")

	ingressCreate := (integrationtest.FakeIngress{
		Name:        ingressName,
		Namespace:   ns,
		ClassName:   ingClassName,
		DnsNames:    []string{"baz.com", "bar.com"},
		ServiceName: svcName,
		TlsSecretDNS: map[string][]string{
			secretName: {"baz.com"},
		},
	}).Ingress()
	_, err := KubeClient.NetworkingV1().Ingresses(ns).Create(context.TODO(), ingressCreate, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	// shardVsName := "cluster--Shared-L7-EVH-"+settingName+"-0"
	secureVsName := "cluster--" + settingName + "-baz.com"
	insecureVsName := "cluster--" + settingName + "-bar.com"
	insecurePoolName := "cluster--" + settingName + "-default-bar.com_foo-" + ingressName + "-" + svcName
	securePoolName := "cluster--" + settingName + "-default-baz.com_foo-" + ingressName + "-" + svcName
	insecurePGName := "cluster--" + settingName + "-default-bar.com_foo-" + ingressName
	securePGName := "cluster--" + settingName + "-default-baz.com_foo-" + ingressName

	g.Eventually(func() int {
		if found, aviSettingModel := objects.SharedAviGraphLister().Get(settingModelName); found {
			if settingNodes := aviSettingModel.(*avinodes.AviObjectGraph).GetAviEvhVS(); len(settingNodes) > 0 {
				return len(settingNodes[0].EvhNodes)
			}
		}
		return 0
	}, 55*time.Second).Should(gomega.Equal(2))
	time.Sleep(5 * time.Second)
	_, aviSettingModel := objects.SharedAviGraphLister().Get(settingModelName)
	settingNodes := aviSettingModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(settingNodes[0].ServiceEngineGroup).Should(gomega.Equal("thisisaviref-" + settingName + "-seGroup"))
	g.Expect(settingNodes[0].SSLKeyCertRefs[0].Name).Should(gomega.Equal(lib.Encode(secureVsName, lib.EVHVS)))
	for _, evhnode := range settingNodes[0].EvhNodes {
		if evhnode.Name == lib.Encode(insecureVsName, lib.EVHVS) {
			g.Expect(evhnode.PoolRefs[0].Name).Should(gomega.Equal(lib.Encode(insecurePoolName, lib.Pool)))
			g.Expect(evhnode.PoolGroupRefs[0].Name).Should(gomega.Equal(lib.Encode(insecurePGName, lib.PG)))
			g.Expect(evhnode.HttpPolicyRefs[0].HppMap[0].Name).Should(gomega.Equal(lib.Encode(insecurePGName, lib.HPPMAP)))
		} else if evhnode.Name == lib.Encode(secureVsName, lib.EVHVS) {
			g.Expect(evhnode.PoolRefs[0].Name).Should(gomega.Equal(lib.Encode(securePoolName, lib.Pool)))
			g.Expect(evhnode.PoolGroupRefs[0].Name).Should(gomega.Equal(lib.Encode(securePGName, lib.PG)))
			g.Expect(evhnode.HttpPolicyRefs[0].HppMap[0].Name).Should(gomega.Equal(lib.Encode(securePGName, lib.HPPMAP)))
		} else {
			t.Fatalf("No matching evh node names found, nodes found: %s, expected one of %s, %s", evhnode.Name, secureVsName, insecureVsName)
		}
	}

	err = KubeClient.NetworkingV1().Ingresses(ns).Delete(context.TODO(), ingressName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	integrationtest.DeleteSecret(secretName, ns)
	integrationtest.TeardownAviInfraSetting(t, settingName)
	TearDownTestForIngress(t, svcName, modelName, settingModelName)
	integrationtest.TeardownIngressClass(t, ingClassName)
	waitAndVerify(t, ingClassName)
	VerifyEvhNodeDeletionFromVsNode(g, modelName, vsKey, evhKey)
}

// AviInfraSetting CRD
func TestEVHAviInfraSettingPerNSNamingConvention(t *testing.T) {
	if lib.VIPPerNamespace() {
		t.Skip()
	}
	// create secure and insecure host ingress, connect with infrasetting
	// check for names of all Avi objects
	g := gomega.NewGomegaWithT(t)

	ingClassName := objNameMap.GenerateName("avi-lb")
	ingressName := objNameMap.GenerateName("foo-with-class")
	ns := "default"
	settingName := objNameMap.GenerateName("my-infrasetting")
	secretName := objNameMap.GenerateName("my-secret")
	svcName := objNameMap.GenerateName("avisvc")
	modelName := "admin/cluster--Shared-L7-EVH-1"

	time.Sleep(time.Second * 5)
	ingestionQueue := utils.SharedWorkQueue().GetQueueByName(utils.ObjectIngestionLayer)
	ingestionQueue.SyncFunc = syncFromIngestionLayerWrapper
	defer func() { ingestionQueue.SyncFunc = k8s.SyncFromIngestionLayer }()
	SetUpTestForIngress(t, svcName, modelName)

	settingModelName := "admin/cluster--Shared-L7-EVH-0"
	vsKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--Shared-L7-EVH-0"}
	evhKey := cache.NamespaceName{Namespace: "admin", Name: lib.Encode("cluster--baz.com", lib.EVHVS)}
	integrationtest.AnnotateAKONamespaceWithInfraSetting(t, ns, settingName)
	integrationtest.SetupAviInfraSetting(t, settingName, "SMALL")
	integrationtest.SetupIngressClass(t, ingClassName, lib.AviIngressController, "")
	waitAndVerify(t, ingClassName)
	integrationtest.AddSecret(secretName, ns, "tlsCert", "tlsKey")

	ingressCreate := (integrationtest.FakeIngress{
		Name:        ingressName,
		Namespace:   ns,
		ClassName:   ingClassName,
		DnsNames:    []string{"baz.com", "bar.com"},
		ServiceName: svcName,
		TlsSecretDNS: map[string][]string{
			secretName: {"baz.com"},
		},
	}).Ingress()
	_, err := KubeClient.NetworkingV1().Ingresses(ns).Create(context.TODO(), ingressCreate, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	// shardVsName := "cluster--Shared-L7-EVH-"+settingName+"-0"
	secureVsName := "cluster--baz.com"
	insecureVsName := "cluster--bar.com"
	insecurePoolName := "cluster--default-bar.com_foo-" + ingressName + "-" + svcName
	securePoolName := "cluster--default-baz.com_foo-" + ingressName + "-" + svcName
	insecurePGName := "cluster--default-bar.com_foo-" + ingressName
	securePGName := "cluster--default-baz.com_foo-" + ingressName

	g.Eventually(func() int {
		if found, aviSettingModel := objects.SharedAviGraphLister().Get(settingModelName); found {
			if settingNodes := aviSettingModel.(*avinodes.AviObjectGraph).GetAviEvhVS(); len(settingNodes) > 0 {
				return len(settingNodes[0].EvhNodes)
			}
		}
		return 0
	}, 55*time.Second).Should(gomega.Equal(2))
	time.Sleep(5 * time.Second)
	_, aviSettingModel := objects.SharedAviGraphLister().Get(settingModelName)
	settingNodes := aviSettingModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(settingNodes[0].ServiceEngineGroup).Should(gomega.Equal("thisisaviref-" + settingName + "-seGroup"))
	g.Expect(settingNodes[0].SSLKeyCertRefs[0].Name).Should(gomega.Equal(lib.Encode(secureVsName, lib.EVHVS)))
	for _, evhnode := range settingNodes[0].EvhNodes {
		if evhnode.Name == lib.Encode(insecureVsName, lib.EVHVS) {
			g.Expect(evhnode.PoolRefs[0].Name).Should(gomega.Equal(lib.Encode(insecurePoolName, lib.Pool)))
			g.Expect(evhnode.PoolGroupRefs[0].Name).Should(gomega.Equal(lib.Encode(insecurePGName, lib.PG)))
			g.Expect(evhnode.HttpPolicyRefs[0].HppMap[0].Name).Should(gomega.Equal(lib.Encode(insecurePGName, lib.HPPMAP)))
		} else if evhnode.Name == lib.Encode(secureVsName, lib.EVHVS) {
			g.Expect(evhnode.PoolRefs[0].Name).Should(gomega.Equal(lib.Encode(securePoolName, lib.Pool)))
			g.Expect(evhnode.PoolGroupRefs[0].Name).Should(gomega.Equal(lib.Encode(securePGName, lib.PG)))
			g.Expect(evhnode.HttpPolicyRefs[0].HppMap[0].Name).Should(gomega.Equal(lib.Encode(securePGName, lib.HPPMAP)))
		} else {
			t.Fatalf("No matching evh node names found, nodes found: %s, expected one of %s, %s", evhnode.Name, secureVsName, insecureVsName)
		}
		g.Expect(settingNodes[0].VSVIPRefs[0].T1Lr).Should(gomega.Equal("avi-domain-c9:1234"))
	}

	err = KubeClient.NetworkingV1().Ingresses(ns).Delete(context.TODO(), ingressName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	integrationtest.DeleteSecret(secretName, ns)
	integrationtest.RemoveAnnotateAKONamespaceWithInfraSetting(t, ns)
	integrationtest.TeardownAviInfraSetting(t, settingName)
	TearDownTestForIngress(t, svcName, modelName, settingModelName)
	integrationtest.TeardownIngressClass(t, ingClassName)
	waitAndVerify(t, ingClassName)
	VerifyEvhNodeDeletionFromVsNode(g, modelName, vsKey, evhKey)
}

// Updating IngressClass
func TestEVHAddRemoveInfraSettingInIngressClass(t *testing.T) {
	if lib.VIPPerNamespace() {
		t.Skip()
	}
	// create ingressclass/ingress, add infrasetting ref, model changes
	// remove infrasetting ref, model changes again
	g := gomega.NewGomegaWithT(t)

	ingClassName := objNameMap.GenerateName("avi-lb")
	ingressName := objNameMap.GenerateName("foo-with-class")
	ns := "default"
	settingName := objNameMap.GenerateName("my-infrasetting")
	modelName := "admin/cluster--Shared-L7-EVH-1"
	secretName := objNameMap.GenerateName("my-secret")
	svcName := objNameMap.GenerateName("avisvc")
	vsKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--Shared-L7-EVH-" + settingName + "-0"}
	evhKey := cache.NamespaceName{Namespace: "admin", Name: lib.Encode("cluster--"+settingName+"-baz.com", lib.EVHVS)}

	time.Sleep(time.Second * 5)
	ingestionQueue := utils.SharedWorkQueue().GetQueueByName(utils.ObjectIngestionLayer)
	ingestionQueue.SyncFunc = syncFromIngestionLayerWrapper
	defer func() { ingestionQueue.SyncFunc = k8s.SyncFromIngestionLayer }()

	SetUpTestForIngress(t, svcName, modelName)

	integrationtest.SetupIngressClass(t, ingClassName, lib.AviIngressController, "")
	waitAndVerify(t, ingClassName)
	integrationtest.AddSecret(secretName, ns, "tlsCert", "tlsKey")
	ingressCreate := (integrationtest.FakeIngress{
		Name:        ingressName,
		Namespace:   ns,
		ClassName:   ingClassName,
		DnsNames:    []string{"baz.com", "bar.com"},
		ServiceName: svcName,
		TlsSecretDNS: map[string][]string{
			secretName: {"baz.com"},
		},
	}).Ingress()
	_, err := KubeClient.NetworkingV1().Ingresses(ns).Create(context.TODO(), ingressCreate, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	g.Eventually(func() bool {
		if found, aviModel := objects.SharedAviGraphLister().Get(modelName); found {
			if nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS(); len(nodes) > 0 {
				return len(nodes[0].EvhNodes) == 2
			}
		}
		return false
	}, 40*time.Second).Should(gomega.Equal(true))

	integrationtest.SetupAviInfraSetting(t, settingName, "SMALL")
	settingModelName := "admin/cluster--Shared-L7-EVH-" + settingName + "-0"

	ingClassUpdate := (integrationtest.FakeIngressClass{
		Name:            ingClassName,
		Controller:      lib.AviIngressController,
		AviInfraSetting: settingName,
		Default:         false,
	}).IngressClass()
	ingClassUpdate.ResourceVersion = "2"
	_, err = KubeClient.NetworkingV1().IngressClasses().Update(context.TODO(), ingClassUpdate, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error updating IngressClass")
	}
	waitAndVerify(t, ingClassName)

	g.Eventually(func() bool {
		if found, aviSettingModel := objects.SharedAviGraphLister().Get(settingModelName); found {
			if settingNodes := aviSettingModel.(*avinodes.AviObjectGraph).GetAviEvhVS(); len(settingNodes) > 0 {
				return len(settingNodes[0].EvhNodes) == 2
			}
		}
		return false
	}, 40*time.Second).Should(gomega.Equal(true))

	ingClassUpdate = (integrationtest.FakeIngressClass{
		Name:       ingClassName,
		Controller: lib.AviIngressController,
		Default:    false,
	}).IngressClass()
	ingClassUpdate.ResourceVersion = "3"
	_, err = KubeClient.NetworkingV1().IngressClasses().Update(context.TODO(), ingClassUpdate, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error updating IngressClass")
	}
	waitAndVerify(t, ingClassName)

	err = KubeClient.NetworkingV1().Ingresses(ns).Delete(context.TODO(), ingressName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	integrationtest.DeleteSecret(secretName, ns)
	integrationtest.TeardownAviInfraSetting(t, settingName)
	TearDownTestForIngress(t, svcName, modelName, settingModelName)
	integrationtest.TeardownIngressClass(t, ingClassName)
	waitAndVerify(t, ingClassName)
	VerifyEvhNodeDeletionFromVsNode(g, modelName, vsKey, evhKey)
}

func TestEVHUpdateInfraSettingInIngressClass(t *testing.T) {
	if lib.VIPPerNamespace() {
		t.Skip()
	}
	// create ingressclass/ingress/infrasetting
	// update infrasetting ref in ingressclass, model changes
	g := gomega.NewGomegaWithT(t)

	ingClassName, ingressName, ns, settingName1, settingName2 := "avi-lb-6", "foo-with-class-6", "default", "my-infrasetting-6-1", "my-infrasetting-6-2"
	modelName := "admin/cluster--Shared-L7-EVH-1"
	secretName := objNameMap.GenerateName("my-secret")
	svcName := objNameMap.GenerateName("avisvc")

	time.Sleep(time.Second * 5)
	ingestionQueue := utils.SharedWorkQueue().GetQueueByName(utils.ObjectIngestionLayer)
	ingestionQueue.SyncFunc = syncFromIngestionLayerWrapper
	defer func() { ingestionQueue.SyncFunc = k8s.SyncFromIngestionLayer }()

	SetUpTestForIngress(t, svcName, modelName)

	integrationtest.SetupAviInfraSetting(t, settingName1, "SMALL")
	integrationtest.SetupAviInfraSetting(t, settingName2, "SMALL")
	settingModelName1 := "admin/cluster--Shared-L7-EVH-" + settingName1 + "-0"
	settingModelName2 := "admin/cluster--Shared-L7-EVH-" + settingName2 + "-0"

	integrationtest.SetupIngressClass(t, ingClassName, lib.AviIngressController, settingName1)
	waitAndVerify(t, ingClassName)
	integrationtest.AddSecret(secretName, ns, "tlsCert", "tlsKey")
	ingressCreate := (integrationtest.FakeIngress{
		Name:        ingressName,
		Namespace:   ns,
		ClassName:   ingClassName,
		DnsNames:    []string{"bar.com"},
		ServiceName: svcName,
	}).Ingress()
	_, err := KubeClient.NetworkingV1().Ingresses(ns).Create(context.TODO(), ingressCreate, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	g.Eventually(func() string {
		if found, aviSettingModel := objects.SharedAviGraphLister().Get(settingModelName1); found {
			if settingNodes := aviSettingModel.(*avinodes.AviObjectGraph).GetAviEvhVS(); len(settingNodes[0].EvhNodes) == 1 {
				return settingNodes[0].EvhNodes[0].Name
			}
		}
		return ""
	}, 40*time.Second).Should(gomega.Equal(lib.Encode("cluster--"+settingName1+"-bar.com", lib.EVHVS)))

	ingClassUpdate := (integrationtest.FakeIngressClass{
		Name:            ingClassName,
		Controller:      lib.AviIngressController,
		Default:         false,
		AviInfraSetting: settingName2,
	}).IngressClass()
	ingClassUpdate.ResourceVersion = "2"
	_, err = KubeClient.NetworkingV1().IngressClasses().Update(context.TODO(), ingClassUpdate, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error updating IngressClass")
	}
	waitAndVerify(t, ingClassName)

	g.Eventually(func() int {
		if found, aviSettingModel := objects.SharedAviGraphLister().Get(settingModelName2); found {
			if settingNodes := aviSettingModel.(*avinodes.AviObjectGraph).GetAviEvhVS(); len(settingNodes) > 0 {
				return len(settingNodes[0].EvhNodes)
			}
		}
		return 0
	}, 40*time.Second).Should(gomega.Equal(1))
	_, aviSettingModel := objects.SharedAviGraphLister().Get(settingModelName2)
	settingNodes := aviSettingModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(settingNodes[0].EvhNodes[0].Name).Should(gomega.Equal(lib.Encode("cluster--"+settingName2+"-bar.com", lib.EVHVS)))

	err = KubeClient.NetworkingV1().Ingresses(ns).Delete(context.TODO(), ingressName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	integrationtest.DeleteSecret(secretName, ns)
	integrationtest.TeardownAviInfraSetting(t, settingName1)
	integrationtest.TeardownAviInfraSetting(t, settingName2)
	integrationtest.TeardownIngressClass(t, ingClassName)
	waitAndVerify(t, ingClassName)
	TearDownTestForIngress(t, svcName, modelName, settingModelName1)
	TearDownTestForIngress(t, svcName, modelName, settingModelName2)
	vsKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--Shared-L7-EVH-" + settingName2 + "-0"}
	evhKey := cache.NamespaceName{Namespace: "admin", Name: lib.Encode("cluster--"+settingName2+"-bar.com", lib.EVHVS)}
	VerifyEvhNodeDeletionFromVsNode(g, modelName, vsKey, evhKey)
}

// Updating Ingress
func TestEVHAddIngressClassWithInfraSetting(t *testing.T) {
	if lib.VIPPerNamespace() {
		t.Skip()
	}
	// add ingress, ingressclass with valid infrasetting,
	// add ingressclass in ingress, delete ingress
	g := gomega.NewGomegaWithT(t)

	ingClassName := objNameMap.GenerateName("avi-lb")
	ingressName := objNameMap.GenerateName("foo-with-class")
	ns := "default"
	settingName := objNameMap.GenerateName("my-infrasetting")
	modelName := "admin/cluster--Shared-L7-EVH-1"
	secretName := objNameMap.GenerateName("my-secret")
	svcName := objNameMap.GenerateName("avisvc")
	vsKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--Shared-L7-EVH-" + settingName + "-0"}
	evhKey := cache.NamespaceName{Namespace: "admin", Name: lib.Encode("cluster--"+settingName+"-bar.com", lib.EVHVS)}

	time.Sleep(time.Second * 5)
	ingestionQueue := utils.SharedWorkQueue().GetQueueByName(utils.ObjectIngestionLayer)
	ingestionQueue.SyncFunc = syncFromIngestionLayerWrapper
	defer func() { ingestionQueue.SyncFunc = k8s.SyncFromIngestionLayer }()

	SetUpTestForIngress(t, svcName, modelName)

	integrationtest.SetupAviInfraSetting(t, settingName, "SMALL")
	settingModelName := "admin/cluster--Shared-L7-EVH-" + settingName + "-0"

	integrationtest.SetupIngressClass(t, ingClassName, lib.AviIngressController, settingName)
	waitAndVerify(t, ingClassName)
	integrationtest.AddSecret(secretName, ns, "tlsCert", "tlsKey")
	ingressCreate := (integrationtest.FakeIngress{
		Name:        ingressName,
		Namespace:   ns,
		DnsNames:    []string{"baz.com", "bar.com"},
		ServiceName: svcName,
		TlsSecretDNS: map[string][]string{
			secretName: {"baz.com"},
		},
	}).Ingress()
	_, err := KubeClient.NetworkingV1().Ingresses(ns).Create(context.TODO(), ingressCreate, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	ingressUpdate := (integrationtest.FakeIngress{
		Name:        ingressName,
		Namespace:   ns,
		ClassName:   ingClassName,
		DnsNames:    []string{"bar.com"},
		ServiceName: svcName,
	}).Ingress()
	ingressUpdate.ResourceVersion = "2"
	_, err = KubeClient.NetworkingV1().Ingresses(ns).Update(context.TODO(), ingressUpdate, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating Ingress: %v", err)
	}

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(settingModelName)
		return found
	}, 40*time.Second).Should(gomega.Equal(true))

	g.Eventually(func() int {
		if found, aviSettingModel := objects.SharedAviGraphLister().Get(settingModelName); found {
			if settingNodes := aviSettingModel.(*avinodes.AviObjectGraph).GetAviEvhVS(); len(settingNodes) > 0 {
				return len(settingNodes[0].EvhNodes)
			}
		}
		return 0
	}, 40*time.Second).Should(gomega.Equal(1))
	_, aviSettingModel := objects.SharedAviGraphLister().Get(settingModelName)
	settingNodes := aviSettingModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(settingNodes[0].EvhNodes[0].Name).Should(gomega.Equal(lib.Encode("cluster--"+settingName+"-bar.com", lib.EVHVS)))
	g.Expect(settingNodes[0].EvhNodes[0].PoolRefs[0].Name).Should(gomega.Equal(lib.Encode("cluster--"+settingName+"-default-bar.com_foo-"+ingressName+"-"+svcName, lib.Pool)))

	err = KubeClient.NetworkingV1().Ingresses(ns).Delete(context.TODO(), ingressName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}

	g.Eventually(func() int {
		found, aviSettingModel := objects.SharedAviGraphLister().Get(settingModelName)
		if nodes := aviSettingModel.(*avinodes.AviObjectGraph).GetAviEvhVS(); found && len(nodes) > 0 {
			return len(nodes[0].EvhNodes)
		}
		return -1
	}, 40*time.Second).Should(gomega.Equal(0))

	integrationtest.DeleteSecret(secretName, ns)
	integrationtest.TeardownAviInfraSetting(t, settingName)
	TearDownTestForIngress(t, svcName, modelName, settingModelName)
	integrationtest.TeardownIngressClass(t, ingClassName)
	waitAndVerify(t, ingClassName)
	VerifyEvhNodeDeletionFromVsNode(g, settingModelName, vsKey, evhKey)
}

func TestEVHUpdateIngressClassWithInfraSetting(t *testing.T) {

	settingName1, settingName2 := "my-infrasetting-8-1", "my-infrasetting-8-2"
	settingModelName1, settingModelName2 := "admin/cluster--Shared-L7-EVH-"+settingName1+"-0", "admin/cluster--Shared-L7-EVH-"+settingName2+"-1"
	if lib.VIPPerNamespace() {
		settingModelName1, settingModelName2 = "admin/cluster--Shared-L7-EVH-"+settingName1+"-NS-default", "admin/cluster--Shared-L7-EVH-"+settingName2+"-NS-default"
	}
	objects.SharedAviGraphLister().Delete(settingModelName1)
	objects.SharedAviGraphLister().Delete(settingModelName2)
	// update from ingressclass with infrasetting to another
	// ingressclass with infrasetting in ingress

	g := gomega.NewGomegaWithT(t)

	ingClassName1, ingClassName2 := "avi-lb1-1", "avi-lb2-1"
	ingressName, ns := "foo-with-class-8", "default"

	modelName := "admin/cluster--Shared-L7-EVH-1"
	secretName := objNameMap.GenerateName("my-secret")
	svcName := objNameMap.GenerateName("avisvc")

	time.Sleep(time.Second * 5)
	ingestionQueue := utils.SharedWorkQueue().GetQueueByName(utils.ObjectIngestionLayer)
	ingestionQueue.SyncFunc = syncFromIngestionLayerWrapper
	defer func() { ingestionQueue.SyncFunc = k8s.SyncFromIngestionLayer }()

	SetUpTestForIngress(t, svcName, modelName)

	integrationtest.SetupAviInfraSetting(t, settingName1, "SMALL")
	integrationtest.SetupAviInfraSetting(t, settingName2, "MEDIUM")

	integrationtest.SetupIngressClass(t, ingClassName1, lib.AviIngressController, settingName1)
	waitAndVerify(t, ingClassName1)
	integrationtest.SetupIngressClass(t, ingClassName2, lib.AviIngressController, settingName2)
	waitAndVerify(t, ingClassName2)
	integrationtest.AddSecret(secretName, ns, "tlsCert", "tlsKey")
	ingressCreate := (integrationtest.FakeIngress{
		Name:        ingressName,
		Namespace:   ns,
		ClassName:   ingClassName1,
		DnsNames:    []string{"baz.com", "bar.com"},
		ServiceName: svcName,
		TlsSecretDNS: map[string][]string{
			secretName: {"baz.com"},
		},
	}).Ingress()
	_, err := KubeClient.NetworkingV1().Ingresses(ns).Create(context.TODO(), ingressCreate, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(settingModelName1)
		return found
	}, 40*time.Second).Should(gomega.Equal(true))
	var settingNodes1 []*avinodes.AviEvhVsNode
	g.Eventually(func() int {
		_, aviSettingModel := objects.SharedAviGraphLister().Get(settingModelName1)
		settingNodes1 = aviSettingModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return len(settingNodes1[0].EvhNodes)
	}, 40*time.Second).Should(gomega.Equal(2))
	g.Expect(settingNodes1[0].EvhNodes).Should(gomega.HaveLen(2))
	g.Expect(settingNodes1[0].ServiceEngineGroup).Should(gomega.Equal("thisisaviref-" + settingName1 + "-seGroup"))

	ingressUpdate := (integrationtest.FakeIngress{
		Name:        ingressName,
		Namespace:   ns,
		ClassName:   ingClassName2,
		DnsNames:    []string{"baz.com", "bar.com"},
		ServiceName: svcName,
		TlsSecretDNS: map[string][]string{
			secretName: {"baz.com"},
		},
	}).Ingress()
	ingressUpdate.ResourceVersion = "2"
	_, err = KubeClient.NetworkingV1().Ingresses(ns).Update(context.TODO(), ingressUpdate, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating Ingress: %v", err)
	}

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(settingModelName2)
		return found
	}, 40*time.Second).Should(gomega.Equal(true))
	var settingNodes2 []*avinodes.AviEvhVsNode
	g.Eventually(func() int {
		_, aviSettingModel := objects.SharedAviGraphLister().Get(settingModelName2)
		settingNodes2 = aviSettingModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return len(settingNodes2[0].EvhNodes)
	}, 40*time.Second).Should(gomega.Equal(2))
	g.Expect(settingNodes2[0].ServiceEngineGroup).Should(gomega.Equal("thisisaviref-" + settingName2 + "-seGroup"))
	g.Expect(settingNodes2[0].EvhNodes).Should(gomega.HaveLen(2))
	_, aviSettingModel1 := objects.SharedAviGraphLister().Get(settingModelName1)
	settingNodes1 = aviSettingModel1.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(settingNodes1[0].EvhNodes).Should(gomega.HaveLen(0))

	err = KubeClient.NetworkingV1().Ingresses(ns).Delete(context.TODO(), ingressName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	integrationtest.DeleteSecret(secretName, ns)
	integrationtest.TeardownAviInfraSetting(t, settingName1)
	integrationtest.TeardownAviInfraSetting(t, settingName2)
	TearDownTestForIngress(t, svcName, modelName, settingModelName1, settingModelName2)
	integrationtest.TeardownIngressClass(t, ingClassName1)
	waitAndVerify(t, ingClassName1)
	integrationtest.TeardownIngressClass(t, ingClassName2)
	waitAndVerify(t, ingClassName2)
	vsKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--Shared-L7-EVH-" + settingName2 + "-1"}
	evhKey := cache.NamespaceName{Namespace: "admin", Name: lib.Encode("cluster--"+settingName2+"-bar.com", lib.EVHVS)}
	VerifyEvhNodeDeletionFromVsNode(g, settingModelName2, vsKey, evhKey)
}

func TestEVHUpdateWithInfraSetting(t *testing.T) {
	if lib.VIPPerNamespace() {
		t.Skip()
	}
	// update from ingressclass with infrasetting to another
	// ingressclass with infrasetting in ingress

	g := gomega.NewGomegaWithT(t)

	ingClassName := objNameMap.GenerateName("avi-lb")
	ingressName := objNameMap.GenerateName("foo-with-class")
	ns := "default"
	settingName := objNameMap.GenerateName("my-infrasetting")
	modelName := "admin/cluster--Shared-L7-EVH-1"
	secretName := objNameMap.GenerateName("my-secret")
	svcName := objNameMap.GenerateName("avisvc")

	time.Sleep(time.Second * 5)
	ingestionQueue := utils.SharedWorkQueue().GetQueueByName(utils.ObjectIngestionLayer)
	ingestionQueue.SyncFunc = syncFromIngestionLayerWrapper
	defer func() { ingestionQueue.SyncFunc = k8s.SyncFromIngestionLayer }()

	SetUpTestForIngress(t, svcName, modelName)

	integrationtest.SetupIngressClass(t, ingClassName, lib.AviIngressController, settingName)
	waitAndVerify(t, ingClassName)
	integrationtest.AddSecret(secretName, ns, "tlsCert", "tlsKey")
	ingressCreate := (integrationtest.FakeIngress{
		Name:        ingressName,
		Namespace:   ns,
		ClassName:   ingClassName,
		DnsNames:    []string{"baz.com", "bar.com"},
		ServiceName: svcName,
		TlsSecretDNS: map[string][]string{
			secretName: {"baz.com"},
		},
	}).Ingress()
	_, err := KubeClient.NetworkingV1().Ingresses(ns).Create(context.TODO(), ingressCreate, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	settingModelName := "admin/cluster--Shared-L7-EVH-" + settingName + "-1"

	settingsUpdate := integrationtest.FakeAviInfraSetting{
		Name:        settingName,
		SeGroupName: "thisisaviref-" + settingName + "-seGroup",
		Networks:    []string{"thisisBADaviref-" + settingName + "-networkName"},
		EnableRhi:   true,
	}

	settingCreate := settingsUpdate.AviInfraSetting()
	settingCreate.ResourceVersion = "2"
	if _, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().AviInfraSettings().Create(context.TODO(), settingCreate, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding AviInfraSetting: %v", err)
	}

	g.Eventually(func() string {
		setting, _ := v1beta1CRDClient.AkoV1beta1().AviInfraSettings().Get(context.TODO(), settingName, metav1.GetOptions{})
		return setting.Status.Status
	}, 40*time.Second).Should(gomega.Equal("Rejected"))

	settingUpdate := (integrationtest.FakeAviInfraSetting{
		Name:        settingName,
		SeGroupName: "thisisaviref-seGroup",
		Networks:    []string{"thisisaviref-networkName"},
		EnableRhi:   true,
	}).AviInfraSetting()
	settingUpdate.ResourceVersion = "3"
	if _, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().AviInfraSettings().Update(context.TODO(), settingUpdate, metav1.UpdateOptions{}); err != nil {
		t.Fatalf("error in updating AviInfraSetting: %v", err)
	}

	g.Eventually(func() string {
		setting, _ := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().AviInfraSettings().Get(context.TODO(), settingName, metav1.GetOptions{})
		return setting.Status.Status
	}, 40*time.Second).Should(gomega.Equal("Accepted"))
	g.Eventually(func() bool {
		if found, aviModel := objects.SharedAviGraphLister().Get(settingModelName); found && aviModel != nil {
			if nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS(); len(nodes) > 0 {
				return nodes[0].ServiceEngineGroup == "thisisaviref-seGroup" &&
					len(nodes[0].VSVIPRefs[0].VipNetworks) > 0 &&
					nodes[0].VSVIPRefs[0].VipNetworks[0].NetworkName == "thisisaviref-networkName" &&
					*nodes[0].EnableRhi
			}
		}
		return false
	}, 45*time.Second).Should(gomega.Equal(true))
	settingUpdate = (integrationtest.FakeAviInfraSetting{
		Name:        settingName,
		SeGroupName: "thisisaviref-seGroup",
		Networks:    []string{"multivip-network1", "multivip-network2", "multivip-network3"},
		EnableRhi:   true,
	}).AviInfraSetting()
	settingUpdate.ResourceVersion = "4"
	if _, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().AviInfraSettings().Update(context.TODO(), settingUpdate, metav1.UpdateOptions{}); err != nil {
		t.Fatalf("error in updating AviInfraSetting: %v", err)
	}

	g.Eventually(func() string {
		setting, _ := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().AviInfraSettings().Get(context.TODO(), settingName, metav1.GetOptions{})
		return setting.Status.Status
	}, 40*time.Second).Should(gomega.Equal("Accepted"))

	g.Eventually(func() int {
		if found, aviModel := objects.SharedAviGraphLister().Get(settingModelName); found && aviModel != nil {
			if nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS(); len(nodes) > 0 {
				return len(nodes[0].VSVIPRefs[0].VipNetworks)
			}
		}
		return 0
	}, 40*time.Second).Should(gomega.Equal(3))
	_, aviModel := objects.SharedAviGraphLister().Get(settingModelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(nodes[0].VSVIPRefs[0].VipNetworks[0].NetworkName).Should(gomega.Equal("multivip-network1"))
	g.Expect(nodes[0].VSVIPRefs[0].VipNetworks[1].NetworkName).Should(gomega.Equal("multivip-network2"))
	g.Expect(nodes[0].VSVIPRefs[0].VipNetworks[2].NetworkName).Should(gomega.Equal("multivip-network3"))
	g.Expect(*nodes[0].EnableRhi).Should(gomega.Equal(true))

	err = KubeClient.NetworkingV1().Ingresses(ns).Delete(context.TODO(), ingressName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	integrationtest.DeleteSecret(secretName, ns)
	integrationtest.TeardownAviInfraSetting(t, settingName)
	TearDownTestForIngress(t, svcName, modelName, settingModelName)
	integrationtest.TeardownIngressClass(t, ingClassName)
	waitAndVerify(t, ingClassName)
}

func TestEVHUpdateIngressClassWithoutInfraSetting(t *testing.T) {
	if lib.VIPPerNamespace() {
		t.Skip()
	}
	// update ingressclass (without infrasetting) in ingress
	g := gomega.NewGomegaWithT(t)

	ingClassName1, ingClassName2 := "avi-lb1-2", "avi-lb2-2"
	ingressName, ns := "foo-with-class-10", "default"
	settingName := "my-infrasetting-10"
	modelName := "admin/cluster--Shared-L7-EVH-1"
	settingModelName := "admin/cluster--Shared-L7-EVH-" + settingName + "-1"
	secretName := objNameMap.GenerateName("my-secret")
	svcName := objNameMap.GenerateName("avisvc")
	vsKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--Shared-L7-EVH-" + settingName + "-1"}
	evhKey := cache.NamespaceName{Namespace: "admin", Name: lib.Encode("cluster--"+settingName+"-bar.com", lib.EVHVS)}

	time.Sleep(time.Second * 5)
	ingestionQueue := utils.SharedWorkQueue().GetQueueByName(utils.ObjectIngestionLayer)
	ingestionQueue.SyncFunc = syncFromIngestionLayerWrapper
	defer func() { ingestionQueue.SyncFunc = k8s.SyncFromIngestionLayer }()

	SetUpTestForIngress(t, svcName, modelName, settingModelName)

	integrationtest.SetupAviInfraSetting(t, settingName, "MEDIUM")

	integrationtest.SetupIngressClass(t, ingClassName1, lib.AviIngressController, settingName)
	waitAndVerify(t, ingClassName1)
	integrationtest.SetupIngressClass(t, ingClassName2, lib.AviIngressController, "")
	waitAndVerify(t, ingClassName2)
	integrationtest.AddSecret(secretName, ns, "tlsCert", "tlsKey")
	ingressCreate := (integrationtest.FakeIngress{
		Name:        ingressName,
		Namespace:   ns,
		ClassName:   ingClassName1,
		DnsNames:    []string{"baz.com", "bar.com"},
		ServiceName: svcName,
		TlsSecretDNS: map[string][]string{
			secretName: {"baz.com"},
		},
	}).Ingress()
	_, err := KubeClient.NetworkingV1().Ingresses(ns).Create(context.TODO(), ingressCreate, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(settingModelName)
		return found
	}, 40*time.Second).Should(gomega.Equal(true))

	g.Eventually(func() string {
		_, aviSettingModel := objects.SharedAviGraphLister().Get(settingModelName)
		settingNodes := aviSettingModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return settingNodes[0].ServiceEngineGroup
	}, 40*time.Second).Should(gomega.Equal("thisisaviref-" + settingName + "-seGroup"))

	g.Eventually(func() int {
		_, aviSettingModel := objects.SharedAviGraphLister().Get(settingModelName)
		settingNodes := aviSettingModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return len(settingNodes[0].EvhNodes)
	}, 40*time.Second).Should(gomega.Equal(2))

	ingressUpdate := (integrationtest.FakeIngress{
		Name:        ingressName,
		Namespace:   ns,
		ClassName:   ingClassName2,
		DnsNames:    []string{"baz.com", "bar.com"},
		ServiceName: svcName,
		TlsSecretDNS: map[string][]string{
			secretName: {"baz.com"},
		},
	}).Ingress()
	ingressUpdate.ResourceVersion = "2"
	_, err = KubeClient.NetworkingV1().Ingresses(ns).Update(context.TODO(), ingressUpdate, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating Ingress: %v", err)
	}

	g.Eventually(func() int {
		if found, aviModel := objects.SharedAviGraphLister().Get(modelName); found {
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
			if len(nodes) > 0 {
				return len(nodes[0].EvhNodes)
			}
		}
		return 0
	}, 40*time.Second).Should(gomega.Equal(2))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(nodes[0].EvhNodes).Should(gomega.HaveLen(2))
	g.Expect(nodes[0].ServiceEngineGroup).Should(gomega.Equal("Default-Group"))
	_, aviSettingModel := objects.SharedAviGraphLister().Get(settingModelName)
	settingNodes := aviSettingModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(settingNodes[0].EvhNodes).Should(gomega.HaveLen(0))

	err = KubeClient.NetworkingV1().Ingresses(ns).Delete(context.TODO(), ingressName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	integrationtest.DeleteSecret(secretName, ns)
	integrationtest.TeardownAviInfraSetting(t, settingName)
	TearDownTestForIngress(t, svcName, modelName, settingModelName)
	integrationtest.TeardownIngressClass(t, ingClassName1)
	waitAndVerify(t, ingClassName1)
	integrationtest.TeardownIngressClass(t, ingClassName2)
	waitAndVerify(t, ingClassName2)
	VerifyEvhNodeDeletionFromVsNode(g, modelName, vsKey, evhKey)
}

func TestEVHBGPConfigurationWithInfraSetting(t *testing.T) {
	if lib.VIPPerNamespace() {
		t.Skip()
	}
	g := gomega.NewGomegaWithT(t)

	ingClassName := objNameMap.GenerateName("avi-lb")
	ingressName := objNameMap.GenerateName("foo-with-class")
	ns := "default"
	settingName := objNameMap.GenerateName("my-infrasetting")
	secretName := objNameMap.GenerateName("my-secret")
	svcName := objNameMap.GenerateName("avisvc")
	modelName := "admin/cluster--Shared-L7-EVH-1"
	settingModelName := "admin/cluster--Shared-L7-EVH-" + settingName + "-1"

	time.Sleep(time.Second * 5)
	ingestionQueue := utils.SharedWorkQueue().GetQueueByName(utils.ObjectIngestionLayer)
	ingestionQueue.SyncFunc = syncFromIngestionLayerWrapper
	defer func() { ingestionQueue.SyncFunc = k8s.SyncFromIngestionLayer }()

	SetUpTestForIngress(t, svcName, modelName, settingModelName)

	integrationtest.SetupAviInfraSetting(t, settingName, "LARGE")
	integrationtest.SetupIngressClass(t, ingClassName, lib.AviIngressController, settingName)
	waitAndVerify(t, ingClassName)
	integrationtest.AddSecret(secretName, ns, "tlsCert", "tlsKey")
	mcache := cache.SharedAviObjCache()
	vsKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--Shared-L7-EVH-" + settingName + "-1"}

	ingressCreate := (integrationtest.FakeIngress{
		Name:        ingressName,
		Namespace:   ns,
		ClassName:   ingClassName,
		DnsNames:    []string{"baz.com", "bar.com"},
		ServiceName: svcName,
		TlsSecretDNS: map[string][]string{
			secretName: {"baz.com"},
		},
	}).Ingress()
	_, err := KubeClient.NetworkingV1().Ingresses(ns).Create(context.TODO(), ingressCreate, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	g.Eventually(func() int {
		if found, aviSettingModel := objects.SharedAviGraphLister().Get(settingModelName); found {
			if settingNodes := aviSettingModel.(*avinodes.AviObjectGraph).GetAviEvhVS(); len(settingNodes) > 0 {
				return len(settingNodes[0].EvhNodes)
			}
		}
		return 0
	}, 55*time.Second).Should(gomega.Equal(2))
	_, aviSettingModel := objects.SharedAviGraphLister().Get(settingModelName)
	settingNodes := aviSettingModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(*settingNodes[0].EnableRhi).Should(gomega.Equal(true))
	g.Expect(settingNodes[0].VSVIPRefs[0].BGPPeerLabels).Should(gomega.HaveLen(2))
	g.Expect((settingNodes[0].VSVIPRefs[0].BGPPeerLabels)[0]).Should(gomega.ContainSubstring("peer"))

	settingUpdate := (integrationtest.FakeAviInfraSetting{
		Name:          settingName,
		EnableRhi:     false,
		BGPPeerLabels: []string{"peer1", "peer2"},
	}).AviInfraSetting()
	settingUpdate.ResourceVersion = "2"
	if _, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().AviInfraSettings().Update(context.TODO(), settingUpdate, metav1.UpdateOptions{}); err != nil {
		t.Fatalf("error in updating AviInfraSetting: %v", err)
	}

	// AviInfraSetting is Rejected since enableRhi is false, but the bgpPeerLabels are configured.
	g.Eventually(func() string {
		setting, _ := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().AviInfraSettings().Get(context.TODO(), settingName, metav1.GetOptions{})
		return setting.Status.Status
	}, 40*time.Second).Should(gomega.Equal("Rejected"))

	integrationtest.TeardownAviInfraSetting(t, settingName)
	integrationtest.TeardownIngressClass(t, ingClassName)
	waitAndVerify(t, ingClassName)
	err = KubeClient.NetworkingV1().Ingresses(ns).Delete(context.TODO(), ingressName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	integrationtest.DeleteSecret(secretName, ns)
	// Shard VS remains, Pools are moved/removed
	g.Eventually(func() bool {
		vsCache, found := mcache.VsCacheMeta.AviCacheGet(vsKey)
		if vsCacheObj, _ := vsCache.(*cache.AviVsCache); found {
			return len(vsCacheObj.SNIChildCollection) == 0
		}
		return false
	}, 50*time.Second).Should(gomega.Equal(true))
	TearDownTestForIngress(t, svcName, modelName, settingModelName)
}

func TestEVHBGPConfigurationUpdateLabelWithInfraSetting(t *testing.T) {
	if lib.VIPPerNamespace() {
		t.Skip()
	}
	g := gomega.NewGomegaWithT(t)

	ingClassName := objNameMap.GenerateName("avi-lb")
	ingressName := objNameMap.GenerateName("foo-with-class")
	ns := "default"
	settingName := objNameMap.GenerateName("my-infrasetting")
	secretName := objNameMap.GenerateName("my-secret")
	svcName := objNameMap.GenerateName("avisvc")
	modelName := "admin/cluster--Shared-L7-EVH-1"
	settingModelName := "admin/cluster--Shared-L7-EVH-" + settingName + "-1"
	mcache := cache.SharedAviObjCache()
	vsKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--Shared-L7-EVH-" + settingName + "-1"}

	time.Sleep(time.Second * 5)
	ingestionQueue := utils.SharedWorkQueue().GetQueueByName(utils.ObjectIngestionLayer)
	ingestionQueue.SyncFunc = syncFromIngestionLayerWrapper
	defer func() { ingestionQueue.SyncFunc = k8s.SyncFromIngestionLayer }()

	SetUpTestForIngress(t, svcName, modelName, settingModelName)

	integrationtest.SetupAviInfraSetting(t, settingName, "LARGE")
	integrationtest.SetupIngressClass(t, ingClassName, lib.AviIngressController, settingName)
	waitAndVerify(t, ingClassName)
	integrationtest.AddSecret(secretName, ns, "tlsCert", "tlsKey")

	ingressCreate := (integrationtest.FakeIngress{
		Name:        ingressName,
		Namespace:   ns,
		ClassName:   ingClassName,
		DnsNames:    []string{"baz.com", "bar.com"},
		ServiceName: svcName,
		TlsSecretDNS: map[string][]string{
			secretName: {"baz.com"},
		},
	}).Ingress()
	_, err := KubeClient.NetworkingV1().Ingresses(ns).Create(context.TODO(), ingressCreate, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	settingUpdate := (integrationtest.FakeAviInfraSetting{
		Name:          settingName,
		EnableRhi:     true,
		BGPPeerLabels: []string{"peerUPDATE1", "peerUPDATE2", "peerUPDATE3"},
	}).AviInfraSetting()
	settingUpdate.ResourceVersion = "2"
	if _, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().AviInfraSettings().Update(context.TODO(), settingUpdate, metav1.UpdateOptions{}); err != nil {
		t.Fatalf("error in updating AviInfraSetting: %v", err)
	}

	g.Eventually(func() int {
		if found, aviSettingModel := objects.SharedAviGraphLister().Get(settingModelName); found {
			if settingNodes := aviSettingModel.(*avinodes.AviObjectGraph).GetAviEvhVS(); len(settingNodes) > 0 && len(settingNodes[0].VSVIPRefs) > 0 {
				return len(settingNodes[0].VSVIPRefs[0].BGPPeerLabels)
			}
		}
		return 0
	}, 55*time.Second).Should(gomega.Equal(3))
	_, aviSettingModel := objects.SharedAviGraphLister().Get(settingModelName)
	settingNodes := aviSettingModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect((settingNodes[0].VSVIPRefs[0].BGPPeerLabels)[0]).Should(gomega.ContainSubstring("peerUPDATE"))

	integrationtest.TeardownAviInfraSetting(t, settingName)
	integrationtest.TeardownIngressClass(t, ingClassName)
	waitAndVerify(t, ingClassName)
	err = KubeClient.NetworkingV1().Ingresses(ns).Delete(context.TODO(), ingressName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	integrationtest.DeleteSecret(secretName, ns)
	// Shard VS remains, Pools are moved/removed
	g.Eventually(func() bool {
		vsCache, found := mcache.VsCacheMeta.AviCacheGet(vsKey)
		if vsCacheObj, _ := vsCache.(*cache.AviVsCache); found {
			return len(vsCacheObj.SNIChildCollection) == 0
		}
		return false
	}, 50*time.Second).Should(gomega.Equal(true))
	TearDownTestForIngress(t, svcName, modelName, settingModelName)
}

func TestEVHCRDWithAviInfraSetting(t *testing.T) {

	if lib.VIPPerNamespace() {
		t.Skip()
	}
	g := gomega.NewGomegaWithT(t)

	ingClassName := objNameMap.GenerateName("avi-lb")
	ingressName := objNameMap.GenerateName("foo-with-class")
	ns := "default"
	settingName := objNameMap.GenerateName("my-infrasetting")
	secretName := objNameMap.GenerateName("my-secret")
	svcName := objNameMap.GenerateName("avisvc")
	modelName := "admin/cluster--Shared-L7-EVH-1"
	settingModelName := "admin/cluster--Shared-L7-EVH-" + settingName + "-1"
	hrName := objNameMap.GenerateName("samplehr-foo")
	rrName := objNameMap.GenerateName("samplerr-foo")
	mcache := cache.SharedAviObjCache()
	vsKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--Shared-L7-EVH-" + settingName + "-1"}
	evhKey := cache.NamespaceName{Namespace: "admin", Name: lib.Encode("cluster--"+settingName+"-baz.com", lib.EVHVS)}
	poolKey := cache.NamespaceName{Namespace: "admin", Name: lib.Encode("cluster--"+settingName+"-default-baz.com_foo-"+ingressName+"-"+svcName, lib.Pool)}

	integrationtest.RemoveDefaultIngressClass()

	time.Sleep(time.Second * 5)
	ingestionQueue := utils.SharedWorkQueue().GetQueueByName(utils.ObjectIngestionLayer)
	ingestionQueue.SyncFunc = syncFromIngestionLayerWrapper
	defer func() { ingestionQueue.SyncFunc = k8s.SyncFromIngestionLayer }()

	SetUpTestForIngress(t, svcName, modelName, settingModelName)

	integrationtest.SetupAviInfraSetting(t, settingName, "LARGE")
	integrationtest.SetupIngressClass(t, ingClassName, lib.AviIngressController, settingName)
	waitAndVerify(t, ingClassName)
	integrationtest.AddSecret(secretName, ns, "tlsCert", "tlsKey")

	httpRulePath := "/"
	integrationtest.SetupHostRule(t, hrName, "baz.com", true)
	integrationtest.SetupHTTPRule(t, rrName, "baz.com", httpRulePath)

	ingressCreate := (integrationtest.FakeIngress{
		Name:        ingressName,
		Namespace:   ns,
		ClassName:   ingClassName,
		DnsNames:    []string{"baz.com"},
		ServiceName: svcName,
		TlsSecretDNS: map[string][]string{
			secretName: {"baz.com"},
		},
	}).Ingress()
	_, err := KubeClient.NetworkingV1().Ingresses(ns).Create(context.TODO(), ingressCreate, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	g.Eventually(func() string {
		hostrule, _ := v1beta1CRDClient.AkoV1beta1().HostRules("default").Get(context.TODO(), hrName, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 30*time.Second).Should(gomega.Equal("Accepted"))
	g.Eventually(func() string {
		httprule, _ := v1beta1CRDClient.AkoV1beta1().HTTPRules("default").Get(context.TODO(), rrName, metav1.GetOptions{})
		return httprule.Status.Status
	}, 30*time.Second).Should(gomega.Equal("Accepted"))

	// check for values set in graph layer.
	integrationtest.VerifyMetadataHostRule(t, g, evhKey, "default/"+hrName, true)
	integrationtest.VerifyMetadataHTTPRule(t, g, poolKey, "default/"+rrName+"/"+httpRulePath, true)
	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(settingModelName)
		return found
	}, 40*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(settingModelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(nodes[0].EvhNodes[0].SslKeyAndCertificateRefs).To(gomega.HaveLen(1))
	g.Expect(nodes[0].EvhNodes[0].SslKeyAndCertificateRefs[0]).To(gomega.ContainSubstring("thisisaviref-sslkey"))
	g.Expect(*nodes[0].EvhNodes[0].Enabled).To(gomega.Equal(true))
	g.Expect(*nodes[0].EvhNodes[0].WafPolicyRef).To(gomega.ContainSubstring("thisisaviref-waf"))
	g.Expect(*nodes[0].EvhNodes[0].PoolRefs[0].LbAlgorithm).To(gomega.Equal("LB_ALGORITHM_CONSISTENT_HASH"))
	g.Expect(*nodes[0].EvhNodes[0].PoolRefs[0].LbAlgorithmHash).To(gomega.Equal("LB_ALGORITHM_CONSISTENT_HASH_SOURCE_IP_ADDRESS"))
	g.Expect(*nodes[0].EvhNodes[0].PoolRefs[0].SslProfileRef).To(gomega.ContainSubstring("thisisaviref-sslprofile"))

	integrationtest.TeardownHostRule(t, g, evhKey, hrName)
	integrationtest.TeardownHTTPRule(t, rrName)
	integrationtest.TeardownAviInfraSetting(t, settingName)
	integrationtest.TeardownIngressClass(t, ingClassName)
	waitAndVerify(t, ingClassName)
	err = KubeClient.NetworkingV1().Ingresses(ns).Delete(context.TODO(), ingressName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	integrationtest.DeleteSecret(secretName, ns)
	// Shard VS remains, Pools are moved/removed
	g.Eventually(func() bool {
		vsCache, found := mcache.VsCacheMeta.AviCacheGet(vsKey)
		if vsCacheObj, _ := vsCache.(*cache.AviVsCache); found {
			return len(vsCacheObj.SNIChildCollection) == 0
		}
		return false
	}, 50*time.Second).Should(gomega.Equal(true))
	TearDownTestForIngress(t, svcName, modelName, settingModelName)
	integrationtest.AddDefaultIngressClass()
	waitAndVerify(t, integrationtest.DefaultIngressClass)
}

func TestFQDNsCountForAviInfraSettingWithDedicatedShardSize(t *testing.T) {
	if lib.VIPPerNamespace() {
		t.Skip()
	}
	g := gomega.NewGomegaWithT(t)
	ingClassName := objNameMap.GenerateName("avi-lb")
	ingressName := objNameMap.GenerateName("foo-with-class")
	ns := "default"
	settingName := objNameMap.GenerateName("my-infrasetting")
	modelName := "admin/" + lib.Encode("cluster--"+settingName+"-foo.com-L7-dedicated", lib.EVHVS) + "-EVH"
	secretName := objNameMap.GenerateName("my-secret")
	svcName := objNameMap.GenerateName("avisvc")

	SetUpTestForIngress(t, svcName, modelName)
	integrationtest.RemoveDefaultIngressClass()
	defer integrationtest.AddDefaultIngressClass()

	integrationtest.SetupAviInfraSetting(t, settingName, "DEDICATED")
	integrationtest.SetupIngressClass(t, ingClassName, lib.AviIngressController, settingName)
	integrationtest.AddSecret(secretName, ns, "tlsCert", "tlsKey")

	ingressCreate := (integrationtest.FakeIngress{
		Name:        ingressName,
		Namespace:   ns,
		ClassName:   ingClassName,
		DnsNames:    []string{"foo.com"},
		ServiceName: svcName,
		TlsSecretDNS: map[string][]string{
			secretName: {"foo.com"},
		},
	}).Ingress()
	_, err := KubeClient.NetworkingV1().Ingresses(ns).Create(context.TODO(), ingressCreate, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	g.Eventually(func() int {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if aviModel == nil {
			return 0
		}
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return len(nodes)
	}, 30*time.Second).Should(gomega.Equal(1))

	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	node := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()[0]

	g.Expect(node.VSVIPRefs).To(gomega.HaveLen(1))
	g.Expect(node.VSVIPRefs[0].FQDNs).To(gomega.HaveLen(1))
	for _, fqdn := range node.VSVIPRefs[0].FQDNs {
		g.Expect(fqdn).ShouldNot(gomega.ContainSubstring("L7-dedicated"))
	}
	integrationtest.TeardownAviInfraSetting(t, settingName)
	integrationtest.TeardownIngressClass(t, ingClassName)
	err = KubeClient.NetworkingV1().Ingresses(ns).Delete(context.TODO(), ingressName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	integrationtest.DeleteSecret(secretName, ns)

	mcache := cache.SharedAviObjCache()
	vsKey := cache.NamespaceName{Namespace: "admin", Name: lib.Encode("cluster--"+settingName+"-foo.com-L7-dedicated", lib.EVHVS) + "-EVH"}
	// verify the removal of VS.
	g.Eventually(func() bool {
		_, found := mcache.VsCacheMeta.AviCacheGet(vsKey)
		return found
	}, 50*time.Second).Should(gomega.Equal(false))
	TearDownTestForIngress(t, svcName, modelName)
}

func TestFQDNsCountForAviInfraSettingWithLargeShardSize(t *testing.T) {
	if lib.VIPPerNamespace() {
		t.Skip()
	}
	g := gomega.NewGomegaWithT(t)

	ingClassName := objNameMap.GenerateName("avi-lb")
	ingressName := objNameMap.GenerateName("foo-with-class")
	ns := "default"
	settingName := objNameMap.GenerateName("my-infrasetting")
	modelName := "admin/cluster--Shared-L7-EVH-" + settingName + "-0"
	secretName := objNameMap.GenerateName("my-secret")
	svcName := objNameMap.GenerateName("avisvc")

	SetUpTestForIngress(t, svcName, modelName)
	integrationtest.RemoveDefaultIngressClass()
	defer integrationtest.AddDefaultIngressClass()

	integrationtest.SetupAviInfraSetting(t, settingName, "LARGE")
	integrationtest.SetupIngressClass(t, ingClassName, lib.AviIngressController, settingName)
	integrationtest.AddSecret(secretName, ns, "tlsCert", "tlsKey")

	ingressCreate := (integrationtest.FakeIngress{
		Name:        ingressName,
		Namespace:   ns,
		ClassName:   ingClassName,
		DnsNames:    []string{"foo.com"},
		ServiceName: svcName,
		TlsSecretDNS: map[string][]string{
			secretName: {"foo.com"},
		},
	}).Ingress()
	_, err := KubeClient.NetworkingV1().Ingresses(ns).Create(context.TODO(), ingressCreate, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

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
			g.Expect(fqdn).Should(gomega.ContainSubstring("Shared-L7-EVH"))
		}
	}
	integrationtest.TeardownAviInfraSetting(t, settingName)
	integrationtest.TeardownIngressClass(t, ingClassName)
	err = KubeClient.NetworkingV1().Ingresses(ns).Delete(context.TODO(), ingressName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	integrationtest.DeleteSecret(secretName, ns)

	mcache := cache.SharedAviObjCache()
	vsKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--Shared-L7-EVH-" + settingName + "-0"}
	// Shard VS remains, Pools are moved/removed
	g.Eventually(func() bool {
		sniCache1, found := mcache.VsCacheMeta.AviCacheGet(vsKey)
		sniCacheObj1, _ := sniCache1.(*cache.AviVsCache)
		if found {
			return len(sniCacheObj1.PoolKeyCollection) == 0
		}
		return false
	}, 50*time.Second).Should(gomega.Equal(true))
	TearDownTestForIngress(t, svcName, modelName)
}
