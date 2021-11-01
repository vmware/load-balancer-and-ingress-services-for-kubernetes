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
	"testing"
	"time"

	"github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/cache"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	avinodes "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/nodes"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"
	utils "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/tests/testlib"
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

// Ingress - IngressClass mapping tests

func TestEVHWrongClassMappingInIngress(t *testing.T) {
	// create ingclass, ingress
	// update wrong mapping of class in ingress, VS deleted
	// fix class in ingress, VS created
	g := gomega.NewGomegaWithT(t)

	ingClassName, ingressName, ns := "avi-lb", "foo-with-class", "default"
	modelName, _ := GetModelName("bar.com", "default")
	vsKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--Shared-L7-EVH-1"}
	evhKey := cache.NamespaceName{Namespace: "admin", Name: lib.Encode("cluster--bar.com", lib.EVHVS)}

	SetUpTestForIngress(t, modelName)
	testlib.RemoveDefaultIngressClass()
	defer testlib.AddDefaultIngressClass()

	testlib.SetupIngressClass(t, ingClassName, lib.AviIngressController, "")
	ingressCreate := (testlib.FakeIngress{
		Name:        ingressName,
		Namespace:   ns,
		ClassName:   ingClassName,
		DnsNames:    []string{"bar.com"},
		ServiceName: "avisvc",
	}).Ingress()
	_, err := utils.GetInformers().ClientSet.NetworkingV1().Ingresses(ns).Create(context.TODO(), ingressCreate, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 40*time.Second).Should(gomega.Equal(true))

	g.Eventually(func() int {
		ingress, _ := utils.GetInformers().ClientSet.NetworkingV1().Ingresses(ns).Get(context.TODO(), ingressName, metav1.GetOptions{})
		return len(ingress.Status.LoadBalancer.Ingress)
	}, 40*time.Second).Should(gomega.Equal(1))

	ingressUpdate := (testlib.FakeIngress{
		Name:        ingressName,
		Namespace:   ns,
		ClassName:   "xyz",
		DnsNames:    []string{"bar.com"},
		ServiceName: "avisvc",
	}).Ingress()
	ingressUpdate.ResourceVersion = "2"
	testlib.UpdateObjectOrFail(t, lib.Ingress, ingressUpdate)

	VerifyEvhNodeDeletionFromVsNode(g, modelName, vsKey, evhKey)

	g.Eventually(func() int {
		ingress, _ := utils.GetInformers().ClientSet.NetworkingV1().Ingresses(ns).Get(context.TODO(), ingressName, metav1.GetOptions{})
		return len(ingress.Status.LoadBalancer.Ingress)
	}, 40*time.Second).Should(gomega.Equal(0))

	ingressUpdate2 := (testlib.FakeIngress{
		Name:        ingressName,
		Namespace:   ns,
		ClassName:   ingClassName,
		DnsNames:    []string{"bar.com"},
		ServiceName: "avisvc",
	}).Ingress()
	ingressUpdate2.ResourceVersion = "3"
	testlib.UpdateObjectOrFail(t, lib.Ingress, ingressUpdate2)

	// vsNode must come back up
	g.Eventually(func() int {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS(); len(nodes) > 0 {
			return len(nodes[0].EvhNodes)
		}
		return 0
	}, 60*time.Second).Should(gomega.Equal(1))

	g.Eventually(func() int {
		ingress, _ := utils.GetInformers().ClientSet.NetworkingV1().Ingresses(ns).Get(context.TODO(), ingressName, metav1.GetOptions{})
		return len(ingress.Status.LoadBalancer.Ingress)
	}, 50*time.Second).Should(gomega.Equal(1))

	testlib.DeleteObject(t, lib.Ingress, ingressName, ns)
	TearDownTestForIngress(t, modelName)
	testlib.DeleteObject(t, lib.IngressClass, ingClassName)
	VerifyEvhNodeDeletionFromVsNode(g, modelName, vsKey, evhKey)
}

func TestEVHDefaultIngressClassChange(t *testing.T) {
	// use default ingress class, change default annotation to false
	// check that ingress status is removed
	// change back default class annotation to true
	// ingress status IP comes back
	g := gomega.NewGomegaWithT(t)

	ingClassName, ingressName, ns := "avi-lb", "foo-with-class", "default"
	modelName, _ := GetModelName("bar.com", "default")
	vsKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--Shared-L7-EVH-1"}
	evhKey := cache.NamespaceName{Namespace: "admin", Name: lib.Encode("cluster--bar.com", lib.EVHVS)}

	SetUpTestForIngress(t, modelName)
	testlib.RemoveDefaultIngressClass()
	defer testlib.AddDefaultIngressClass()

	ingClass := (testlib.FakeIngressClass{
		Name:       ingClassName,
		Controller: lib.AviIngressController,
		Default:    true,
	}).IngressClass()
	if _, err := utils.GetInformers().ClientSet.NetworkingV1().IngressClasses().Create(context.TODO(), ingClass, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding IngressClass: %v", err)
	}

	// ingress with no IngressClass
	ingressCreate := (testlib.FakeIngress{
		Name:        ingressName,
		Namespace:   ns,
		DnsNames:    []string{"bar.com"},
		ServiceName: "avisvc",
	}).Ingress()
	_, err := utils.GetInformers().ClientSet.NetworkingV1().Ingresses(ns).Create(context.TODO(), ingressCreate, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	g.Eventually(func() int {
		ingress, _ := utils.GetInformers().ClientSet.NetworkingV1().Ingresses(ns).Get(context.TODO(), ingressName, metav1.GetOptions{})
		return len(ingress.Status.LoadBalancer.Ingress)
	}, 40*time.Second).Should(gomega.Equal(1))

	ingClass.Annotations = map[string]string{lib.DefaultIngressClassAnnotation: "false"}
	ingClass.ResourceVersion = "2"
	testlib.UpdateObjectOrFail(t, lib.IngressClass, ingClass)

	g.Eventually(func() int {
		ingress, _ := utils.GetInformers().ClientSet.NetworkingV1().Ingresses(ns).Get(context.TODO(), ingressName, metav1.GetOptions{})
		return len(ingress.Status.LoadBalancer.Ingress)
	}, 80*time.Second).Should(gomega.Equal(0))

	testlib.DeleteObject(t, lib.Ingress, ingressName, ns)
	TearDownTestForIngress(t, modelName)
	testlib.DeleteObject(t, lib.IngressClass, ingClassName)
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

	ingClassName, ingressName, ns, settingName := "avi-lb", "foo-with-class", "default", "my-infrasetting"
	secretName := "my-secret"
	modelName := "admin/cluster--Shared-L7-EVH-1"

	SetUpTestForIngress(t, modelName)
	testlib.RemoveDefaultIngressClass()
	defer testlib.AddDefaultIngressClass()

	settingModelName := "admin/cluster--Shared-L7-EVH-my-infrasetting-0"
	vsKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--Shared-L7-EVH-my-infrasetting-0"}
	evhKey := cache.NamespaceName{Namespace: "admin", Name: lib.Encode("cluster--my-infrasetting-baz.com", lib.EVHVS)}
	testlib.SetupAviInfraSetting(t, settingName, "SMALL")
	testlib.SetupIngressClass(t, ingClassName, lib.AviIngressController, settingName)
	testlib.AddSecret(secretName, ns, "tlsCert", "tlsKey")

	ingressCreate := (testlib.FakeIngress{
		Name:        ingressName,
		Namespace:   ns,
		ClassName:   ingClassName,
		DnsNames:    []string{"baz.com", "bar.com"},
		ServiceName: "avisvc",
		TlsSecretDNS: map[string][]string{
			secretName: {"baz.com"},
		},
	}).Ingress()
	_, err := utils.GetInformers().ClientSet.NetworkingV1().Ingresses(ns).Create(context.TODO(), ingressCreate, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	// shardVsName := "cluster--Shared-L7-EVH-my-infrasetting-0"
	secureVsName := "cluster--my-infrasetting-baz.com"
	insecureVsName := "cluster--my-infrasetting-bar.com"
	insecurePoolName := "cluster--my-infrasetting-default-bar.com_foo-foo-with-class-avisvc"
	securePoolName := "cluster--my-infrasetting-default-baz.com_foo-foo-with-class-avisvc"
	insecurePGName := "cluster--my-infrasetting-default-bar.com_foo-foo-with-class"
	securePGName := "cluster--my-infrasetting-default-baz.com_foo-foo-with-class"

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
	g.Expect(settingNodes[0].ServiceEngineGroup).Should(gomega.Equal("thisisaviref-my-infrasetting-seGroup"))
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

	testlib.DeleteObject(t, lib.Ingress, ingressName, ns)
	testlib.DeleteObject(t, lib.Secret, secretName, ns)
	testlib.DeleteObject(t, lib.AviInfraSetting, settingName)
	TearDownTestForIngress(t, modelName, settingModelName)
	testlib.DeleteObject(t, lib.IngressClass, ingClassName)
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

	ingClassName, ingressName, ns, settingName := "avi-lb", "foo-with-class", "default", "my-infrasetting"
	modelName := "admin/cluster--Shared-L7-EVH-1"
	secretName := "my-secret"
	vsKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--Shared-L7-EVH-my-infrasetting-0"}
	evhKey := cache.NamespaceName{Namespace: "admin", Name: lib.Encode("cluster--my-infrasetting-baz.com", lib.EVHVS)}

	SetUpTestForIngress(t, modelName)
	testlib.RemoveDefaultIngressClass()
	defer testlib.AddDefaultIngressClass()

	testlib.SetupIngressClass(t, ingClassName, lib.AviIngressController, "")
	testlib.AddSecret(secretName, ns, "tlsCert", "tlsKey")
	ingressCreate := (testlib.FakeIngress{
		Name:        ingressName,
		Namespace:   ns,
		ClassName:   ingClassName,
		DnsNames:    []string{"baz.com", "bar.com"},
		ServiceName: "avisvc",
		TlsSecretDNS: map[string][]string{
			secretName: {"baz.com"},
		},
	}).Ingress()
	_, err := utils.GetInformers().ClientSet.NetworkingV1().Ingresses(ns).Create(context.TODO(), ingressCreate, metav1.CreateOptions{})
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

	testlib.SetupAviInfraSetting(t, settingName, "SMALL")
	settingModelName := "admin/cluster--Shared-L7-EVH-my-infrasetting-0"

	ingClassUpdate := (testlib.FakeIngressClass{
		Name:            ingClassName,
		Controller:      lib.AviIngressController,
		AviInfraSetting: settingName,
		Default:         false,
	}).IngressClass()
	ingClassUpdate.ResourceVersion = "2"
	testlib.UpdateObjectOrFail(t, lib.IngressClass, ingClassUpdate)

	g.Eventually(func() bool {
		if found, aviSettingModel := objects.SharedAviGraphLister().Get(settingModelName); found {
			if settingNodes := aviSettingModel.(*avinodes.AviObjectGraph).GetAviEvhVS(); len(settingNodes) > 0 {
				return len(settingNodes[0].EvhNodes) == 2
			}
		}
		return false
	}, 40*time.Second).Should(gomega.Equal(true))
	VerifyEvhNodeDeletionFromVsNode(g, modelName, vsKey, evhKey)

	ingClassUpdate = (testlib.FakeIngressClass{
		Name:       ingClassName,
		Controller: lib.AviIngressController,
		Default:    false,
	}).IngressClass()
	ingClassUpdate.ResourceVersion = "3"
	testlib.UpdateObjectOrFail(t, lib.IngressClass, ingClassUpdate)

	testlib.DeleteObject(t, lib.Ingress, ingressName, ns)
	testlib.DeleteObject(t, lib.Secret, secretName, ns)
	testlib.DeleteObject(t, lib.AviInfraSetting, settingName)
	TearDownTestForIngress(t, modelName, settingModelName)
	testlib.DeleteObject(t, lib.IngressClass, ingClassName)
	VerifyEvhNodeDeletionFromVsNode(g, modelName, vsKey, evhKey)
}

func TestEVHUpdateInfraSettingInIngressClass(t *testing.T) {
	if lib.VIPPerNamespace() {
		t.Skip()
	}
	// create ingressclass/ingress/infrasetting
	// update infrasetting ref in ingressclass, model changes
	g := gomega.NewGomegaWithT(t)

	ingClassName, ingressName, ns, settingName1, settingName2 := "avi-lb", "foo-with-class", "default", "my-infrasetting", "my-infrasetting2"
	modelName := "admin/cluster--Shared-L7-EVH-1"
	secretName := "my-secret"

	SetUpTestForIngress(t, modelName)
	testlib.RemoveDefaultIngressClass()
	defer testlib.AddDefaultIngressClass()

	testlib.SetupAviInfraSetting(t, settingName1, "SMALL")
	testlib.SetupAviInfraSetting(t, settingName2, "SMALL")
	settingModelName1 := "admin/cluster--Shared-L7-EVH-my-infrasetting-0"
	settingModelName2 := "admin/cluster--Shared-L7-EVH-my-infrasetting2-0"

	testlib.SetupIngressClass(t, ingClassName, lib.AviIngressController, settingName1)
	testlib.AddSecret(secretName, ns, "tlsCert", "tlsKey")
	ingressCreate := (testlib.FakeIngress{
		Name:        ingressName,
		Namespace:   ns,
		ClassName:   ingClassName,
		DnsNames:    []string{"bar.com"},
		ServiceName: "avisvc",
	}).Ingress()
	_, err := utils.GetInformers().ClientSet.NetworkingV1().Ingresses(ns).Create(context.TODO(), ingressCreate, metav1.CreateOptions{})
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
	}, 40*time.Second).Should(gomega.Equal(lib.Encode("cluster--my-infrasetting-bar.com", lib.EVHVS)))

	ingClassUpdate := (testlib.FakeIngressClass{
		Name:            ingClassName,
		Controller:      lib.AviIngressController,
		Default:         false,
		AviInfraSetting: settingName2,
	}).IngressClass()
	ingClassUpdate.ResourceVersion = "2"
	testlib.UpdateObjectOrFail(t, lib.IngressClass, ingClassUpdate)

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
	g.Expect(settingNodes[0].EvhNodes[0].Name).Should(gomega.Equal(lib.Encode("cluster--my-infrasetting2-bar.com", lib.EVHVS)))

	testlib.DeleteObject(t, lib.Ingress, ingressName, ns)
	testlib.DeleteObject(t, lib.Secret, secretName, ns)
	testlib.DeleteObject(t, lib.AviInfraSetting, settingName1)
	testlib.DeleteObject(t, lib.AviInfraSetting, settingName2)
	testlib.DeleteObject(t, lib.IngressClass, ingClassName)
	TearDownTestForIngress(t, modelName, settingModelName1)
	TearDownTestForIngress(t, modelName, settingModelName2)
	vsKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--Shared-L7-EVH-my-infrasetting2-0"}
	evhKey := cache.NamespaceName{Namespace: "admin", Name: lib.Encode("cluster--my-infrasetting2-bar.com", lib.EVHVS)}
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

	ingClassName, ingressName, ns, settingName := "avi-lb", "foo-with-class", "default", "my-infrasetting"
	modelName := "admin/cluster--Shared-L7-EVH-1"
	secretName := "my-secret"
	vsKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--Shared-L7-EVH-my-infrasetting-0"}
	evhKey := cache.NamespaceName{Namespace: "admin", Name: lib.Encode("cluster--my-infrasetting-bar.com", lib.EVHVS)}

	SetUpTestForIngress(t, modelName)
	testlib.RemoveDefaultIngressClass()
	defer testlib.AddDefaultIngressClass()

	testlib.SetupAviInfraSetting(t, settingName, "SMALL")
	settingModelName := "admin/cluster--Shared-L7-EVH-my-infrasetting-0"

	testlib.SetupIngressClass(t, ingClassName, lib.AviIngressController, settingName)
	testlib.AddSecret(secretName, ns, "tlsCert", "tlsKey")
	ingressCreate := (testlib.FakeIngress{
		Name:        ingressName,
		Namespace:   ns,
		DnsNames:    []string{"baz.com", "bar.com"},
		ServiceName: "avisvc",
		TlsSecretDNS: map[string][]string{
			secretName: {"baz.com"},
		},
	}).Ingress()
	_, err := utils.GetInformers().ClientSet.NetworkingV1().Ingresses(ns).Create(context.TODO(), ingressCreate, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	ingressUpdate := (testlib.FakeIngress{
		Name:        ingressName,
		Namespace:   ns,
		ClassName:   ingClassName,
		DnsNames:    []string{"bar.com"},
		ServiceName: "avisvc",
	}).Ingress()
	ingressUpdate.ResourceVersion = "2"
	testlib.UpdateObjectOrFail(t, lib.Ingress, ingressUpdate)

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
	g.Expect(settingNodes[0].EvhNodes[0].Name).Should(gomega.Equal(lib.Encode("cluster--my-infrasetting-bar.com", lib.EVHVS)))
	g.Expect(settingNodes[0].EvhNodes[0].PoolRefs[0].Name).Should(gomega.Equal(lib.Encode("cluster--my-infrasetting-default-bar.com_foo-foo-with-class-avisvc", lib.Pool)))

	testlib.DeleteObject(t, lib.Ingress, ingressName, ns)

	g.Eventually(func() int {
		found, aviSettingModel := objects.SharedAviGraphLister().Get(settingModelName)
		if nodes := aviSettingModel.(*avinodes.AviObjectGraph).GetAviEvhVS(); found && len(nodes) > 0 {
			return len(nodes[0].EvhNodes)
		}
		return -1
	}, 40*time.Second).Should(gomega.Equal(0))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	g.Expect(aviModel).Should(gomega.BeNil())

	testlib.DeleteObject(t, lib.Secret, secretName, ns)
	testlib.DeleteObject(t, lib.AviInfraSetting, settingName)
	TearDownTestForIngress(t, modelName, settingModelName)
	testlib.DeleteObject(t, lib.IngressClass, ingClassName)
	VerifyEvhNodeDeletionFromVsNode(g, settingModelName, vsKey, evhKey)
}

func TestEVHUpdateIngressClassWithInfraSetting(t *testing.T) {
	if lib.VIPPerNamespace() {
		t.Skip()
	}
	// update from ingressclass with infrasetting to another
	// ingressclass with infrasetting in ingress

	g := gomega.NewGomegaWithT(t)

	ingClassName1, ingClassName2 := "avi-lb1", "avi-lb2"
	ingressName, ns := "foo-with-class", "default"
	settingName1, settingName2 := "my-infrasetting1", "my-infrasetting2"
	modelName := "admin/cluster--Shared-L7-EVH-1"
	secretName := "my-secret"

	SetUpTestForIngress(t, modelName)
	testlib.RemoveDefaultIngressClass()
	defer testlib.AddDefaultIngressClass()

	testlib.SetupAviInfraSetting(t, settingName1, "SMALL")
	testlib.SetupAviInfraSetting(t, settingName2, "MEDIUM")
	settingModelName1, settingModelName2 := "admin/cluster--Shared-L7-EVH-my-infrasetting1-0", "admin/cluster--Shared-L7-EVH-my-infrasetting2-1"

	testlib.SetupIngressClass(t, ingClassName1, lib.AviIngressController, settingName1)
	testlib.SetupIngressClass(t, ingClassName2, lib.AviIngressController, settingName2)
	testlib.AddSecret(secretName, ns, "tlsCert", "tlsKey")
	ingressCreate := (testlib.FakeIngress{
		Name:        ingressName,
		Namespace:   ns,
		ClassName:   ingClassName1,
		DnsNames:    []string{"baz.com", "bar.com"},
		ServiceName: "avisvc",
		TlsSecretDNS: map[string][]string{
			secretName: {"baz.com"},
		},
	}).Ingress()
	_, err := utils.GetInformers().ClientSet.NetworkingV1().Ingresses(ns).Create(context.TODO(), ingressCreate, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(settingModelName1)
		return found
	}, 40*time.Second).Should(gomega.Equal(true))
	_, aviSettingModel1 := objects.SharedAviGraphLister().Get(settingModelName1)
	settingNodes1 := aviSettingModel1.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(settingNodes1[0].EvhNodes).Should(gomega.HaveLen(2))
	g.Expect(settingNodes1[0].ServiceEngineGroup).Should(gomega.Equal("thisisaviref-my-infrasetting1-seGroup"))

	ingressUpdate := (testlib.FakeIngress{
		Name:        ingressName,
		Namespace:   ns,
		ClassName:   ingClassName2,
		DnsNames:    []string{"baz.com", "bar.com"},
		ServiceName: "avisvc",
		TlsSecretDNS: map[string][]string{
			secretName: {"baz.com"},
		},
	}).Ingress()
	ingressUpdate.ResourceVersion = "2"
	testlib.UpdateObjectOrFail(t, lib.Ingress, ingressUpdate)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(settingModelName2)
		return found
	}, 40*time.Second).Should(gomega.Equal(true))
	_, aviSettingModel2 := objects.SharedAviGraphLister().Get(settingModelName2)
	settingNodes2 := aviSettingModel2.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(settingNodes2[0].ServiceEngineGroup).Should(gomega.Equal("thisisaviref-my-infrasetting2-seGroup"))
	g.Expect(settingNodes2[0].EvhNodes).Should(gomega.HaveLen(2))
	_, aviSettingModel1 = objects.SharedAviGraphLister().Get(settingModelName1)
	settingNodes1 = aviSettingModel1.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(settingNodes1[0].EvhNodes).Should(gomega.HaveLen(0))

	testlib.DeleteObject(t, lib.Ingress, ingressName, ns)
	testlib.DeleteObject(t, lib.Secret, secretName, ns)
	testlib.DeleteObject(t, lib.AviInfraSetting, settingName1)
	testlib.DeleteObject(t, lib.AviInfraSetting, settingName2)
	TearDownTestForIngress(t, modelName, settingModelName1, settingModelName2)
	testlib.DeleteObject(t, lib.IngressClass, ingClassName1)
	testlib.DeleteObject(t, lib.IngressClass, ingClassName2)
	vsKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--Shared-L7-EVH-my-infrasetting2-1"}
	evhKey := cache.NamespaceName{Namespace: "admin", Name: lib.Encode("cluster--my-infrasetting2-bar.com", lib.EVHVS)}
	VerifyEvhNodeDeletionFromVsNode(g, settingModelName2, vsKey, evhKey)
}

func TestEVHUpdateWithInfraSetting(t *testing.T) {
	if lib.VIPPerNamespace() {
		t.Skip()
	}
	// update from ingressclass with infrasetting to another
	// ingressclass with infrasetting in ingress

	g := gomega.NewGomegaWithT(t)

	ingClassName, ingressName, ns, settingName := "avi-lb", "foo-with-class", "default", "my-infrasetting"
	modelName := "admin/cluster--Shared-L7-EVH-1"
	secretName := "my-secret"

	SetUpTestForIngress(t, modelName)
	testlib.RemoveDefaultIngressClass()
	defer testlib.AddDefaultIngressClass()

	testlib.SetupIngressClass(t, ingClassName, lib.AviIngressController, settingName)
	testlib.AddSecret(secretName, ns, "tlsCert", "tlsKey")
	ingressCreate := (testlib.FakeIngress{
		Name:        ingressName,
		Namespace:   ns,
		ClassName:   ingClassName,
		DnsNames:    []string{"baz.com", "bar.com"},
		ServiceName: "avisvc",
		TlsSecretDNS: map[string][]string{
			secretName: {"baz.com"},
		},
	}).Ingress()
	_, err := utils.GetInformers().ClientSet.NetworkingV1().Ingresses(ns).Create(context.TODO(), ingressCreate, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	settingModelName := "admin/cluster--Shared-L7-EVH-my-infrasetting-1"

	settingsUpdate := testlib.FakeAviInfraSetting{
		Name:        settingName,
		SeGroupName: "thisisaviref-" + settingName + "-seGroup",
		Networks:    []string{"thisisBADaviref-" + settingName + "-networkName"},
		EnableRhi:   true,
	}

	settingCreate := settingsUpdate.AviInfraSetting()
	settingCreate.ResourceVersion = "2"
	if _, err := lib.AKOControlConfig().CRDClientset().AkoV1alpha1().AviInfraSettings().Create(context.TODO(), settingCreate, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding AviInfraSetting: %v", err)
	}

	g.Eventually(func() string {
		setting, _ := lib.AKOControlConfig().CRDClientset().AkoV1alpha1().AviInfraSettings().Get(context.TODO(), settingName, metav1.GetOptions{})
		return setting.Status.Status
	}, 40*time.Second).Should(gomega.Equal("Rejected"))
	g.Eventually(func() bool {
		if found, _ := objects.SharedAviGraphLister().Get(modelName); !found {
			return true
		}
		return false
	}, 40*time.Second).Should(gomega.Equal(true))
	settingUpdate := (testlib.FakeAviInfraSetting{
		Name:        settingName,
		SeGroupName: "thisisaviref-seGroup",
		Networks:    []string{"thisisaviref-networkName"},
		EnableRhi:   true,
	}).AviInfraSetting()
	settingUpdate.ResourceVersion = "3"
	testlib.UpdateObjectOrFail(t, lib.AviInfraSetting, settingUpdate)

	g.Eventually(func() string {
		setting, _ := lib.AKOControlConfig().CRDClientset().AkoV1alpha1().AviInfraSettings().Get(context.TODO(), settingName, metav1.GetOptions{})
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
	settingUpdate = (testlib.FakeAviInfraSetting{
		Name:        settingName,
		SeGroupName: "thisisaviref-seGroup",
		Networks:    []string{"multivip-network1", "multivip-network2", "multivip-network3"},
		EnableRhi:   true,
	}).AviInfraSetting()
	settingUpdate.ResourceVersion = "4"
	testlib.UpdateObjectOrFail(t, lib.AviInfraSetting, settingUpdate)

	g.Eventually(func() string {
		setting, _ := lib.AKOControlConfig().CRDClientset().AkoV1alpha1().AviInfraSettings().Get(context.TODO(), settingName, metav1.GetOptions{})
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

	testlib.DeleteObject(t, lib.Ingress, ingressName, ns)
	testlib.DeleteObject(t, lib.Secret, secretName, ns)
	testlib.DeleteObject(t, lib.AviInfraSetting, settingName)
	TearDownTestForIngress(t, modelName, settingModelName)
	testlib.DeleteObject(t, lib.IngressClass, ingClassName)
}

func TestEVHUpdateIngressClassWithoutInfraSetting(t *testing.T) {
	if lib.VIPPerNamespace() {
		t.Skip()
	}
	// update ingressclass (without infrasetting) in ingress
	g := gomega.NewGomegaWithT(t)

	ingClassName1, ingClassName2 := "avi-lb1", "avi-lb2"
	ingressName, ns := "foo-with-class", "default"
	settingName := "my-infrasetting"
	modelName := "admin/cluster--Shared-L7-EVH-1"
	settingModelName := "admin/cluster--Shared-L7-EVH-my-infrasetting-1"
	secretName := "my-secret"
	vsKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--Shared-L7-EVH-my-infrasetting-1"}
	evhKey := cache.NamespaceName{Namespace: "admin", Name: lib.Encode("cluster--my-infrasetting-bar.com", lib.EVHVS)}

	SetUpTestForIngress(t, modelName, settingModelName)
	testlib.RemoveDefaultIngressClass()
	defer testlib.AddDefaultIngressClass()

	testlib.SetupAviInfraSetting(t, settingName, "MEDIUM")

	testlib.SetupIngressClass(t, ingClassName1, lib.AviIngressController, settingName)
	testlib.SetupIngressClass(t, ingClassName2, lib.AviIngressController, "")
	testlib.AddSecret(secretName, ns, "tlsCert", "tlsKey")
	ingressCreate := (testlib.FakeIngress{
		Name:        ingressName,
		Namespace:   ns,
		ClassName:   ingClassName1,
		DnsNames:    []string{"baz.com", "bar.com"},
		ServiceName: "avisvc",
		TlsSecretDNS: map[string][]string{
			secretName: {"baz.com"},
		},
	}).Ingress()
	_, err := utils.GetInformers().ClientSet.NetworkingV1().Ingresses(ns).Create(context.TODO(), ingressCreate, metav1.CreateOptions{})
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
	}, 40*time.Second).Should(gomega.Equal("thisisaviref-my-infrasetting-seGroup"))

	g.Eventually(func() int {
		_, aviSettingModel := objects.SharedAviGraphLister().Get(settingModelName)
		settingNodes := aviSettingModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return len(settingNodes[0].EvhNodes)
	}, 40*time.Second).Should(gomega.Equal(2))

	ingressUpdate := (testlib.FakeIngress{
		Name:        ingressName,
		Namespace:   ns,
		ClassName:   ingClassName2,
		DnsNames:    []string{"baz.com", "bar.com"},
		ServiceName: "avisvc",
		TlsSecretDNS: map[string][]string{
			secretName: {"baz.com"},
		},
	}).Ingress()
	ingressUpdate.ResourceVersion = "2"
	testlib.UpdateObjectOrFail(t, lib.Ingress, ingressUpdate)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 40*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(nodes[0].EvhNodes).Should(gomega.HaveLen(2))
	g.Expect(nodes[0].ServiceEngineGroup).Should(gomega.Equal("Default-Group"))
	_, aviSettingModel := objects.SharedAviGraphLister().Get(settingModelName)
	settingNodes := aviSettingModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(settingNodes[0].EvhNodes).Should(gomega.HaveLen(0))

	testlib.DeleteObject(t, lib.Ingress, ingressName, ns)
	testlib.DeleteObject(t, lib.Secret, secretName, ns)
	testlib.DeleteObject(t, lib.AviInfraSetting, settingName)
	TearDownTestForIngress(t, modelName, settingModelName)
	testlib.DeleteObject(t, lib.IngressClass, ingClassName1)
	testlib.DeleteObject(t, lib.IngressClass, ingClassName2)
	VerifyEvhNodeDeletionFromVsNode(g, modelName, vsKey, evhKey)
}

func TestEVHBGPConfigurationWithInfraSetting(t *testing.T) {
	if lib.VIPPerNamespace() {
		t.Skip()
	}
	g := gomega.NewGomegaWithT(t)

	ingClassName, ingressName, ns, settingName := "avi-lb", "foo-with-class", "default", "my-infrasetting"
	secretName := "my-secret"
	modelName := "admin/cluster--Shared-L7-EVH-1"
	settingModelName := "admin/cluster--Shared-L7-EVH-my-infrasetting-1"

	SetUpTestForIngress(t, modelName, settingModelName)
	testlib.RemoveDefaultIngressClass()
	defer testlib.AddDefaultIngressClass()

	testlib.SetupAviInfraSetting(t, settingName, "LARGE")
	testlib.SetupIngressClass(t, ingClassName, lib.AviIngressController, settingName)
	testlib.AddSecret(secretName, ns, "tlsCert", "tlsKey")
	mcache := cache.SharedAviObjCache()
	vsKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--Shared-L7-EVH-my-infrasetting-1"}

	ingressCreate := (testlib.FakeIngress{
		Name:        ingressName,
		Namespace:   ns,
		ClassName:   ingClassName,
		DnsNames:    []string{"baz.com", "bar.com"},
		ServiceName: "avisvc",
		TlsSecretDNS: map[string][]string{
			secretName: {"baz.com"},
		},
	}).Ingress()
	_, err := utils.GetInformers().ClientSet.NetworkingV1().Ingresses(ns).Create(context.TODO(), ingressCreate, metav1.CreateOptions{})
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

	settingUpdate := (testlib.FakeAviInfraSetting{
		Name:          settingName,
		EnableRhi:     false,
		BGPPeerLabels: []string{"peer1", "peer2"},
	}).AviInfraSetting()
	settingUpdate.ResourceVersion = "2"
	testlib.UpdateObjectOrFail(t, lib.AviInfraSetting, settingUpdate)

	// AviInfraSetting is Rejected since enableRhi is false, but the bgpPeerLabels are configured.
	g.Eventually(func() string {
		setting, _ := lib.AKOControlConfig().CRDClientset().AkoV1alpha1().AviInfraSettings().Get(context.TODO(), settingName, metav1.GetOptions{})
		return setting.Status.Status
	}, 40*time.Second).Should(gomega.Equal("Rejected"))

	testlib.DeleteObject(t, lib.AviInfraSetting, settingName)
	testlib.DeleteObject(t, lib.IngressClass, ingClassName)
	testlib.DeleteObject(t, lib.Ingress, ingressName, ns)
	testlib.DeleteObject(t, lib.Secret, secretName, ns)
	// Shard VS remains, Pools are moved/removed
	g.Eventually(func() bool {
		vsCache, found := mcache.VsCacheMeta.AviCacheGet(vsKey)
		if vsCacheObj, _ := vsCache.(*cache.AviVsCache); found {
			return len(vsCacheObj.SNIChildCollection) == 0
		}
		return false
	}, 50*time.Second).Should(gomega.Equal(true))
	TearDownTestForIngress(t, modelName, settingModelName)
}

func TestEVHBGPConfigurationUpdateLabelWithInfraSetting(t *testing.T) {
	if lib.VIPPerNamespace() {
		t.Skip()
	}
	g := gomega.NewGomegaWithT(t)

	ingClassName, ingressName, ns, settingName := "avi-lb", "foo-with-class", "default", "my-infrasetting"
	secretName := "my-secret"
	modelName := "admin/cluster--Shared-L7-EVH-1"
	settingModelName := "admin/cluster--Shared-L7-EVH-my-infrasetting-1"
	mcache := cache.SharedAviObjCache()
	vsKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--Shared-L7-EVH-my-infrasetting-1"}

	SetUpTestForIngress(t, modelName, settingModelName)
	testlib.RemoveDefaultIngressClass()
	defer testlib.AddDefaultIngressClass()

	testlib.SetupAviInfraSetting(t, settingName, "LARGE")
	testlib.SetupIngressClass(t, ingClassName, lib.AviIngressController, settingName)
	testlib.AddSecret(secretName, ns, "tlsCert", "tlsKey")

	ingressCreate := (testlib.FakeIngress{
		Name:        ingressName,
		Namespace:   ns,
		ClassName:   ingClassName,
		DnsNames:    []string{"baz.com", "bar.com"},
		ServiceName: "avisvc",
		TlsSecretDNS: map[string][]string{
			secretName: {"baz.com"},
		},
	}).Ingress()
	_, err := utils.GetInformers().ClientSet.NetworkingV1().Ingresses(ns).Create(context.TODO(), ingressCreate, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	settingUpdate := (testlib.FakeAviInfraSetting{
		Name:          settingName,
		EnableRhi:     true,
		BGPPeerLabels: []string{"peerUPDATE1", "peerUPDATE2", "peerUPDATE3"},
	}).AviInfraSetting()
	settingUpdate.ResourceVersion = "2"
	testlib.UpdateObjectOrFail(t, lib.AviInfraSetting, settingUpdate)

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

	testlib.DeleteObject(t, lib.AviInfraSetting, settingName)
	testlib.DeleteObject(t, lib.IngressClass, ingClassName)
	testlib.DeleteObject(t, lib.Ingress, ingressName, ns)
	testlib.DeleteObject(t, lib.Secret, secretName, ns)
	// Shard VS remains, Pools are moved/removed
	g.Eventually(func() bool {
		vsCache, found := mcache.VsCacheMeta.AviCacheGet(vsKey)
		if vsCacheObj, _ := vsCache.(*cache.AviVsCache); found {
			return len(vsCacheObj.SNIChildCollection) == 0
		}
		return false
	}, 50*time.Second).Should(gomega.Equal(true))
	TearDownTestForIngress(t, modelName, settingModelName)
}

func TestEVHCRDWithAviInfraSetting(t *testing.T) {
	if lib.VIPPerNamespace() {
		t.Skip()
	}
	g := gomega.NewGomegaWithT(t)

	ingClassName, ingressName, ns, settingName := "avi-lb", "foo-with-class", "default", "my-infrasetting"
	secretName := "my-secret"
	modelName := "admin/cluster--Shared-L7-EVH-1"
	settingModelName := "admin/cluster--Shared-L7-EVH-my-infrasetting-1"
	hrname, rrname := "samplehr-baz", "samplerr-baz"
	mcache := cache.SharedAviObjCache()
	vsKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--Shared-L7-EVH-my-infrasetting-1"}
	evhKey := cache.NamespaceName{Namespace: "admin", Name: lib.Encode("cluster--my-infrasetting-baz.com", lib.EVHVS)}
	poolKey := cache.NamespaceName{Namespace: "admin", Name: lib.Encode("cluster--my-infrasetting-default-baz.com_foo-foo-with-class-avisvc", lib.Pool)}

	SetUpTestForIngress(t, modelName, settingModelName)
	testlib.RemoveDefaultIngressClass()
	defer testlib.AddDefaultIngressClass()

	testlib.SetupAviInfraSetting(t, settingName, "LARGE")
	testlib.SetupIngressClass(t, ingClassName, lib.AviIngressController, settingName)
	testlib.AddSecret(secretName, ns, "tlsCert", "tlsKey")

	httpRulePath := "/"
	testlib.SetupHostRule(t, hrname, "baz.com", true)
	testlib.SetupHTTPRule(t, rrname, "baz.com", httpRulePath)

	ingressCreate := (testlib.FakeIngress{
		Name:        ingressName,
		Namespace:   ns,
		ClassName:   ingClassName,
		DnsNames:    []string{"baz.com"},
		ServiceName: "avisvc",
		TlsSecretDNS: map[string][]string{
			secretName: {"baz.com"},
		},
	}).Ingress()
	_, err := utils.GetInformers().ClientSet.NetworkingV1().Ingresses(ns).Create(context.TODO(), ingressCreate, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	g.Eventually(func() string {
		hostrule, _ := lib.AKOControlConfig().CRDClientset().AkoV1alpha1().HostRules("default").Get(context.TODO(), hrname, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 30*time.Second).Should(gomega.Equal("Accepted"))
	g.Eventually(func() string {
		httprule, _ := lib.AKOControlConfig().CRDClientset().AkoV1alpha1().HTTPRules("default").Get(context.TODO(), rrname, metav1.GetOptions{})
		return httprule.Status.Status
	}, 30*time.Second).Should(gomega.Equal("Accepted"))

	// check for values set in graph layer.
	testlib.VerifyMetadataHostRule(t, g, evhKey, "default/"+hrname, true)
	testlib.VerifyMetadataHTTPRule(t, g, poolKey, "default/"+rrname+"/"+httpRulePath, true)
	_, aviModel := objects.SharedAviGraphLister().Get(settingModelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(nodes[0].EvhNodes[0].SSLKeyCertAviRef).To(gomega.ContainSubstring("thisisaviref-sslkey"))
	g.Expect(*nodes[0].EvhNodes[0].Enabled).To(gomega.Equal(true))
	g.Expect(nodes[0].EvhNodes[0].WafPolicyRef).To(gomega.ContainSubstring("thisisaviref-waf"))
	g.Expect(nodes[0].EvhNodes[0].PoolRefs[0].LbAlgorithm).To(gomega.Equal("LB_ALGORITHM_CONSISTENT_HASH"))
	g.Expect(nodes[0].EvhNodes[0].PoolRefs[0].LbAlgorithmHash).To(gomega.Equal("LB_ALGORITHM_CONSISTENT_HASH_SOURCE_IP_ADDRESS"))
	g.Expect(nodes[0].EvhNodes[0].PoolRefs[0].SslProfileRef).To(gomega.ContainSubstring("thisisaviref-sslprofile"))

	testlib.TeardownHostRule(t, g, evhKey, hrname)
	testlib.DeleteObject(t, lib.HTTPRule, rrname, ns)
	testlib.DeleteObject(t, lib.AviInfraSetting, settingName)
	testlib.DeleteObject(t, lib.IngressClass, ingClassName)
	testlib.DeleteObject(t, lib.Ingress, ingressName, ns)
	testlib.DeleteObject(t, lib.Secret, secretName, ns)
	// Shard VS remains, Pools are moved/removed
	g.Eventually(func() bool {
		vsCache, found := mcache.VsCacheMeta.AviCacheGet(vsKey)
		if vsCacheObj, _ := vsCache.(*cache.AviVsCache); found {
			return len(vsCacheObj.SNIChildCollection) == 0
		}
		return false
	}, 50*time.Second).Should(gomega.Equal(true))
	TearDownTestForIngress(t, modelName, settingModelName)
}
