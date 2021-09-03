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

package ingresstests

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
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/tests/integrationtest"
)

func VerifyPoolDeletionFromVsNode(g *gomega.WithT, modelName string) {
	g.Eventually(func() bool {
		if found, aviModel := objects.SharedAviGraphLister().Get(modelName); found {
			if nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS(); len(nodes) > 0 {
				return len(nodes[0].PoolRefs) == 0
			}
		}
		return true
	}, 50*time.Second).Should(gomega.Equal(true))
}

// Ingress - IngressClass mapping tests

func TestWrongClassMappingInIngress(t *testing.T) {
	// create ingclass, ingress
	// update wrong mapping of class in ingress, VS deleted
	// fix class in ingress, VS created
	g := gomega.NewGomegaWithT(t)

	ingClassName, ingressName, ns := "avi-lb", "foo-with-class", "default"
	modelName := "admin/cluster--Shared-L7-1"

	SetUpTestForIngress(t, modelName)
	integrationtest.RemoveDefaultIngressClass()
	defer integrationtest.AddDefaultIngressClass()

	integrationtest.SetupIngressClass(t, ingClassName, lib.AviIngressController, "")
	ingressCreate := (integrationtest.FakeIngress{
		Name:        ingressName,
		Namespace:   ns,
		ClassName:   ingClassName,
		DnsNames:    []string{"bar.com"},
		ServiceName: "avisvc",
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
		ServiceName: "avisvc",
	}).Ingress()
	ingressUpdate.ResourceVersion = "2"
	if _, err := KubeClient.NetworkingV1().Ingresses(ns).Update(context.TODO(), ingressUpdate, metav1.UpdateOptions{}); err != nil {
		t.Fatalf("error in updating Ingress: %v", err)
	}

	VerifyPoolDeletionFromVsNode(g, modelName)

	g.Eventually(func() int {
		ingress, _ := KubeClient.NetworkingV1().Ingresses(ns).Get(context.TODO(), ingressName, metav1.GetOptions{})
		return len(ingress.Status.LoadBalancer.Ingress)
	}, 40*time.Second).Should(gomega.Equal(0))

	ingressUpdate2 := (integrationtest.FakeIngress{
		Name:        ingressName,
		Namespace:   ns,
		ClassName:   ingClassName,
		DnsNames:    []string{"bar.com"},
		ServiceName: "avisvc",
	}).Ingress()
	ingressUpdate.ResourceVersion = "3"
	if _, err := KubeClient.NetworkingV1().Ingresses(ns).Update(context.TODO(), ingressUpdate2, metav1.UpdateOptions{}); err != nil {
		t.Fatalf("error in updating Ingress: %v", err)
	}

	// vsNode must come back up
	g.Eventually(func() int {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS(); len(nodes) > 0 {
			return len(nodes[0].PoolRefs)
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
	TearDownTestForIngress(t, modelName)
	integrationtest.TeardownIngressClass(t, ingClassName)
	VerifyPoolDeletionFromVsNode(g, modelName)
}

func TestDefaultIngressClassChange(t *testing.T) {
	// use default ingress class, change default annotation to false
	// check that ingress status is removed
	// change back default class annotation to true
	// ingress status IP comes back
	g := gomega.NewGomegaWithT(t)

	ingClassName, ingressName, ns := "avi-lb", "foo-with-class2", "default"
	modelName := "admin/cluster--Shared-L7-1"

	SetUpTestForIngress(t, modelName)
	integrationtest.RemoveDefaultIngressClass()
	defer integrationtest.AddDefaultIngressClass()

	ingClass := (integrationtest.FakeIngressClass{
		Name:       ingClassName,
		Controller: lib.AviIngressController,
		Default:    true,
	}).IngressClass()
	if _, err := KubeClient.NetworkingV1().IngressClasses().Create(context.TODO(), ingClass, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding IngressClass: %v", err)
	}

	// ingress with no IngressClass
	ingressCreate := (integrationtest.FakeIngress{
		Name:        ingressName,
		Namespace:   ns,
		DnsNames:    []string{"bar.com"},
		ServiceName: "avisvc",
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

	g.Eventually(func() int {
		ingress, _ := KubeClient.NetworkingV1().Ingresses(ns).Get(context.TODO(), ingressName, metav1.GetOptions{})
		return len(ingress.Status.LoadBalancer.Ingress)
	}, 40*time.Second).Should(gomega.Equal(0))

	err = KubeClient.NetworkingV1().Ingresses(ns).Delete(context.TODO(), ingressName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	TearDownTestForIngress(t, modelName)
	integrationtest.TeardownIngressClass(t, ingClassName)
	VerifyPoolDeletionFromVsNode(g, modelName)
}

// AviInfraSetting CRD
func TestAviInfraSettingNamingConvention(t *testing.T) {
	// create secure and insecure host ingress, connect with infrasetting
	// check for names of all Avi objects
	g := gomega.NewGomegaWithT(t)

	ingClassName, ingressName, ns, settingName := "avi-lb", "foo-with-class", "default", "my-infrasetting"
	secretName := "my-secret"
	modelName := "admin/cluster--Shared-L7-1"

	SetUpTestForIngress(t, modelName)
	integrationtest.RemoveDefaultIngressClass()
	defer integrationtest.AddDefaultIngressClass()

	settingModelName := "admin/cluster--Shared-L7-my-infrasetting-0"
	integrationtest.SetupAviInfraSetting(t, settingName, "SMALL")
	integrationtest.SetupIngressClass(t, ingClassName, lib.AviIngressController, settingName)
	integrationtest.AddSecret(secretName, ns, "tlsCert", "tlsKey")

	ingressCreate := (integrationtest.FakeIngress{
		Name:        ingressName,
		Namespace:   ns,
		ClassName:   ingClassName,
		DnsNames:    []string{"baz.com", "bar.com"},
		ServiceName: "avisvc",
		TlsSecretDNS: map[string][]string{
			secretName: {"baz.com"},
		},
	}).Ingress()
	_, err := KubeClient.NetworkingV1().Ingresses(ns).Create(context.TODO(), ingressCreate, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	shardVsName := "cluster--Shared-L7-my-infrasetting-0"
	sniVsName := "cluster--my-infrasetting-baz.com"
	shardPoolName := "cluster--my-infrasetting-bar.com_foo-default-foo-with-class"
	sniPoolName := "cluster--my-infrasetting-default-baz.com_foo-foo-with-class"

	g.Eventually(func() string {
		if found, aviSettingModel := objects.SharedAviGraphLister().Get(settingModelName); found {
			if settingNodes := aviSettingModel.(*avinodes.AviObjectGraph).GetAviVS(); len(settingNodes) > 0 &&
				len(settingNodes[0].SniNodes) > 0 {
				return settingNodes[0].SniNodes[0].Name
			}
		}
		return ""
	}, 55*time.Second).Should(gomega.Equal(sniVsName))
	_, aviSettingModel := objects.SharedAviGraphLister().Get(settingModelName)
	settingNodes := aviSettingModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(settingNodes[0].PoolRefs[0].Name).Should(gomega.Equal(shardPoolName))
	g.Expect(settingNodes[0].ServiceEngineGroup).Should(gomega.Equal("thisisaviref-my-infrasetting-seGroup"))
	g.Expect(settingNodes[0].PoolGroupRefs[0].Name).Should(gomega.Equal(shardVsName))
	g.Expect(settingNodes[0].HTTPDSrefs[0].Name).Should(gomega.Equal(shardVsName))
	g.Expect(settingNodes[0].HttpPolicyRefs[0].Name).Should(gomega.Equal(shardVsName))
	g.Expect(settingNodes[0].SniNodes[0].PoolRefs[0].Name).Should(gomega.Equal(sniPoolName))
	g.Expect(settingNodes[0].SniNodes[0].PoolGroupRefs[0].Name).Should(gomega.Equal(sniPoolName))
	g.Expect(settingNodes[0].SniNodes[0].SSLKeyCertRefs[0].Name).Should(gomega.Equal(sniVsName))
	g.Expect(settingNodes[0].SniNodes[0].HttpPolicyRefs[0].Name).Should(gomega.Equal(sniPoolName))

	err = KubeClient.NetworkingV1().Ingresses(ns).Delete(context.TODO(), ingressName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	integrationtest.DeleteSecret(secretName, ns)
	integrationtest.TeardownAviInfraSetting(t, settingName)
	TearDownTestForIngress(t, modelName, settingModelName)
	integrationtest.TeardownIngressClass(t, ingClassName)
	VerifyPoolDeletionFromVsNode(g, modelName)
}

// Updating IngressClass
func TestAddRemoveInfraSettingInIngressClass(t *testing.T) {
	// create ingressclass/ingress, add infrasetting ref, model changes
	// remove infrasetting ref, model changes again
	g := gomega.NewGomegaWithT(t)

	ingClassName, ingressName, ns, settingName := "avi-lb", "foo-with-class", "default", "my-infrasetting"
	modelName := "admin/cluster--Shared-L7-1"
	secretName := "my-secret"

	SetUpTestForIngress(t, modelName)
	integrationtest.RemoveDefaultIngressClass()
	defer integrationtest.AddDefaultIngressClass()

	integrationtest.SetupIngressClass(t, ingClassName, lib.AviIngressController, "")
	integrationtest.AddSecret(secretName, ns, "tlsCert", "tlsKey")
	ingressCreate := (integrationtest.FakeIngress{
		Name:        ingressName,
		Namespace:   ns,
		ClassName:   ingClassName,
		DnsNames:    []string{"baz.com", "bar.com"},
		ServiceName: "avisvc",
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
			if nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS(); len(nodes) > 0 {
				return len(nodes[0].PoolRefs) == 1 && len(nodes[0].SniNodes) == 1
			}
		}
		return false
	}, 40*time.Second).Should(gomega.Equal(true))

	integrationtest.SetupAviInfraSetting(t, settingName, "SMALL")
	settingModelName := "admin/cluster--Shared-L7-my-infrasetting-0"

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

	g.Eventually(func() bool {
		if found, aviSettingModel := objects.SharedAviGraphLister().Get(settingModelName); found {
			if settingNodes := aviSettingModel.(*avinodes.AviObjectGraph).GetAviVS(); len(settingNodes) > 0 {
				return len(settingNodes[0].PoolRefs) == 1
			}
		}
		return false
	}, 40*time.Second).Should(gomega.Equal(true))
	_, aviSettingModel := objects.SharedAviGraphLister().Get(settingModelName)
	settingNodes := aviSettingModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(settingNodes[0].PoolRefs).Should(gomega.HaveLen(1))
	g.Expect(settingNodes[0].PoolRefs[0].Name).Should(gomega.Equal("cluster--my-infrasetting-bar.com_foo-default-foo-with-class"))
	g.Eventually(func() int {
		_, aviSettingModel := objects.SharedAviGraphLister().Get(settingModelName)
		settingNodes = aviSettingModel.(*avinodes.AviObjectGraph).GetAviVS()
		return len(settingNodes[0].SniNodes)
	}, 40*time.Second).Should(gomega.Equal(1))
	g.Expect(settingNodes[0].SniNodes[0].Name).Should(gomega.Equal("cluster--my-infrasetting-baz.com"))

	VerifyPoolDeletionFromVsNode(g, modelName)

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

	err = KubeClient.NetworkingV1().Ingresses(ns).Delete(context.TODO(), ingressName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	integrationtest.DeleteSecret(secretName, ns)
	integrationtest.TeardownAviInfraSetting(t, settingName)
	TearDownTestForIngress(t, modelName, settingModelName)
	integrationtest.TeardownIngressClass(t, ingClassName)
	VerifyPoolDeletionFromVsNode(g, modelName)
}

func TestUpdateInfraSettingInIngressClass(t *testing.T) {
	// create ingressclass/ingress/infrasetting
	// update infrasetting ref in ingressclass, model changes
	g := gomega.NewGomegaWithT(t)

	ingClassName, ingressName, ns, settingName1, settingName2 := "avi-lb", "foo-with-class", "default", "my-infrasetting", "my-infrasetting2"
	modelName := "admin/cluster--Shared-L7-1"
	secretName := "my-secret"

	SetUpTestForIngress(t, modelName)
	integrationtest.RemoveDefaultIngressClass()
	defer integrationtest.AddDefaultIngressClass()

	integrationtest.SetupAviInfraSetting(t, settingName1, "SMALL")
	integrationtest.SetupAviInfraSetting(t, settingName2, "SMALL")
	settingModelName1 := "admin/cluster--Shared-L7-my-infrasetting-0"
	settingModelName2 := "admin/cluster--Shared-L7-my-infrasetting2-0"

	integrationtest.SetupIngressClass(t, ingClassName, lib.AviIngressController, settingName1)
	integrationtest.AddSecret(secretName, ns, "tlsCert", "tlsKey")
	ingressCreate := (integrationtest.FakeIngress{
		Name:        ingressName,
		Namespace:   ns,
		ClassName:   ingClassName,
		DnsNames:    []string{"bar.com"},
		ServiceName: "avisvc",
	}).Ingress()
	_, err := KubeClient.NetworkingV1().Ingresses(ns).Create(context.TODO(), ingressCreate, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	g.Eventually(func() string {
		if found, aviSettingModel := objects.SharedAviGraphLister().Get(settingModelName1); found {
			if settingNodes := aviSettingModel.(*avinodes.AviObjectGraph).GetAviVS(); len(settingNodes[0].PoolRefs) == 1 {
				return settingNodes[0].PoolRefs[0].Name
			}
		}
		return ""
	}, 40*time.Second).Should(gomega.Equal("cluster--my-infrasetting-bar.com_foo-default-foo-with-class"))

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

	g.Eventually(func() int {
		if found, aviSettingModel := objects.SharedAviGraphLister().Get(settingModelName2); found {
			if settingNodes := aviSettingModel.(*avinodes.AviObjectGraph).GetAviVS(); len(settingNodes) > 0 {
				return len(settingNodes[0].PoolRefs)
			}
		}
		return 0
	}, 40*time.Second).Should(gomega.Equal(1))
	_, aviSettingModel := objects.SharedAviGraphLister().Get(settingModelName2)
	settingNodes := aviSettingModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(settingNodes[0].PoolRefs[0].Name).Should(gomega.Equal("cluster--my-infrasetting2-bar.com_foo-default-foo-with-class"))

	err = KubeClient.NetworkingV1().Ingresses(ns).Delete(context.TODO(), ingressName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	integrationtest.DeleteSecret(secretName, ns)
	integrationtest.TeardownAviInfraSetting(t, settingName1)
	integrationtest.TeardownAviInfraSetting(t, settingName2)
	TearDownTestForIngress(t, modelName, settingModelName1)
	TearDownTestForIngress(t, modelName, settingModelName2)
	integrationtest.TeardownIngressClass(t, ingClassName)
	VerifyPoolDeletionFromVsNode(g, modelName)
}

// Updating Ingress
func TestAddIngressClassWithInfraSetting(t *testing.T) {
	// add ingress, ingressclass with valid infrasetting,
	// add ingressclass in ingress, delete ingress
	g := gomega.NewGomegaWithT(t)

	ingClassName, ingressName, ns, settingName := "avi-lb", "foo-with-class", "default", "my-infrasetting"
	modelName := "admin/cluster--Shared-L7-1"
	secretName := "my-secret"

	SetUpTestForIngress(t, modelName)
	integrationtest.RemoveDefaultIngressClass()
	defer integrationtest.AddDefaultIngressClass()

	integrationtest.SetupAviInfraSetting(t, settingName, "SMALL")
	settingModelName := "admin/cluster--Shared-L7-my-infrasetting-0"

	integrationtest.SetupIngressClass(t, ingClassName, lib.AviIngressController, settingName)
	integrationtest.AddSecret(secretName, ns, "tlsCert", "tlsKey")
	ingressCreate := (integrationtest.FakeIngress{
		Name:        ingressName,
		Namespace:   ns,
		DnsNames:    []string{"baz.com", "bar.com"},
		ServiceName: "avisvc",
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
		ServiceName: "avisvc",
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
			if settingNodes := aviSettingModel.(*avinodes.AviObjectGraph).GetAviVS(); len(settingNodes) > 0 {
				return len(settingNodes[0].PoolRefs)
			}
		}
		return 0
	}, 40*time.Second).Should(gomega.Equal(1))
	_, aviSettingModel := objects.SharedAviGraphLister().Get(settingModelName)
	settingNodes := aviSettingModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(settingNodes[0].PoolRefs[0].Name).Should(gomega.Equal("cluster--my-infrasetting-bar.com_foo-default-foo-with-class"))

	err = KubeClient.NetworkingV1().Ingresses(ns).Delete(context.TODO(), ingressName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}

	VerifyPoolDeletionFromVsNode(g, modelName)
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	g.Expect(aviModel).Should(gomega.BeNil())

	integrationtest.DeleteSecret(secretName, ns)
	integrationtest.TeardownAviInfraSetting(t, settingName)
	TearDownTestForIngress(t, modelName, settingModelName)
	integrationtest.TeardownIngressClass(t, ingClassName)
	VerifyPoolDeletionFromVsNode(g, settingModelName)
}

func TestUpdateIngressClassWithInfraSetting(t *testing.T) {
	// update from ingressclass with infrasetting to another
	// ingressclass with infrasetting in ingress

	g := gomega.NewGomegaWithT(t)

	ingClassName1, ingClassName2 := "avi-lb1", "avi-lb2"
	ingressName, ns := "foo-with-class", "default"
	settingName1, settingName2 := "my-infrasetting1", "my-infrasetting2"
	modelName := "admin/cluster--Shared-L7-1"
	secretName := "my-secret"

	SetUpTestForIngress(t, modelName)
	integrationtest.RemoveDefaultIngressClass()
	defer integrationtest.AddDefaultIngressClass()

	integrationtest.SetupAviInfraSetting(t, settingName1, "SMALL")
	integrationtest.SetupAviInfraSetting(t, settingName2, "MEDIUM")
	settingModelName1, settingModelName2 := "admin/cluster--Shared-L7-my-infrasetting1-0", "admin/cluster--Shared-L7-my-infrasetting2-1"

	integrationtest.SetupIngressClass(t, ingClassName1, lib.AviIngressController, settingName1)
	integrationtest.SetupIngressClass(t, ingClassName2, lib.AviIngressController, settingName2)
	integrationtest.AddSecret(secretName, ns, "tlsCert", "tlsKey")
	ingressCreate := (integrationtest.FakeIngress{
		Name:        ingressName,
		Namespace:   ns,
		ClassName:   ingClassName1,
		DnsNames:    []string{"baz.com", "bar.com"},
		ServiceName: "avisvc",
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
	_, aviSettingModel1 := objects.SharedAviGraphLister().Get(settingModelName1)
	settingNodes1 := aviSettingModel1.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(settingNodes1[0].PoolRefs).Should(gomega.HaveLen(1))
	g.Expect(settingNodes1[0].ServiceEngineGroup).Should(gomega.Equal("thisisaviref-my-infrasetting1-seGroup"))
	g.Expect(settingNodes1[0].PoolRefs[0].Name).Should(gomega.Equal("cluster--my-infrasetting1-bar.com_foo-default-foo-with-class"))

	ingressUpdate := (integrationtest.FakeIngress{
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
	_, err = KubeClient.NetworkingV1().Ingresses(ns).Update(context.TODO(), ingressUpdate, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating Ingress: %v", err)
	}

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(settingModelName2)
		return found
	}, 40*time.Second).Should(gomega.Equal(true))
	_, aviSettingModel2 := objects.SharedAviGraphLister().Get(settingModelName2)
	settingNodes2 := aviSettingModel2.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(settingNodes2[0].PoolRefs).Should(gomega.HaveLen(1))
	g.Expect(settingNodes2[0].PoolRefs[0].Name).Should(gomega.Equal("cluster--my-infrasetting2-bar.com_foo-default-foo-with-class"))
	g.Expect(settingNodes2[0].ServiceEngineGroup).Should(gomega.Equal("thisisaviref-my-infrasetting2-seGroup"))
	_, aviSettingModel1 = objects.SharedAviGraphLister().Get(settingModelName1)
	settingNodes1 = aviSettingModel1.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(settingNodes1[0].PoolRefs).Should(gomega.HaveLen(0))

	err = KubeClient.NetworkingV1().Ingresses(ns).Delete(context.TODO(), ingressName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	integrationtest.DeleteSecret(secretName, ns)
	integrationtest.TeardownAviInfraSetting(t, settingName1)
	integrationtest.TeardownAviInfraSetting(t, settingName2)
	TearDownTestForIngress(t, modelName, settingModelName1, settingModelName2)
	integrationtest.TeardownIngressClass(t, ingClassName1)
	integrationtest.TeardownIngressClass(t, ingClassName2)
	VerifyPoolDeletionFromVsNode(g, settingModelName2)
}

func TestUpdateWithInfraSetting(t *testing.T) {
	// update from ingressclass with infrasetting to another
	// ingressclass with infrasetting in ingress

	g := gomega.NewGomegaWithT(t)

	ingClassName, ingressName, ns, settingName := "avi-lb", "foo-with-class", "default", "my-infrasetting"
	modelName := "admin/cluster--Shared-L7-1"
	secretName := "my-secret"

	SetUpTestForIngress(t, modelName)
	integrationtest.RemoveDefaultIngressClass()
	defer integrationtest.AddDefaultIngressClass()

	integrationtest.SetupIngressClass(t, ingClassName, lib.AviIngressController, settingName)
	integrationtest.AddSecret(secretName, ns, "tlsCert", "tlsKey")
	ingressCreate := (integrationtest.FakeIngress{
		Name:        ingressName,
		Namespace:   ns,
		ClassName:   ingClassName,
		DnsNames:    []string{"baz.com", "bar.com"},
		ServiceName: "avisvc",
		TlsSecretDNS: map[string][]string{
			secretName: {"baz.com"},
		},
	}).Ingress()
	_, err := KubeClient.NetworkingV1().Ingresses(ns).Create(context.TODO(), ingressCreate, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	settingModelName := "admin/cluster--Shared-L7-my-infrasetting-1"

	settingsUpdate := integrationtest.FakeAviInfraSetting{
		Name:        settingName,
		SeGroupName: "thisisaviref-" + settingName + "-seGroup",
		Networks:    []string{"thisisBADaviref-" + settingName + "-networkName"},
		EnableRhi:   true,
	}

	settingCreate := settingsUpdate.AviInfraSetting()
	settingCreate.ResourceVersion = "2"
	if _, err := lib.GetCRDClientset().AkoV1alpha1().AviInfraSettings().Create(context.TODO(), settingCreate, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding AviInfraSetting: %v", err)
	}

	g.Eventually(func() string {
		setting, _ := CRDClient.AkoV1alpha1().AviInfraSettings().Get(context.TODO(), settingName, metav1.GetOptions{})
		return setting.Status.Status
	}, 40*time.Second).Should(gomega.Equal("Rejected"))
	g.Eventually(func() bool {
		if found, _ := objects.SharedAviGraphLister().Get(modelName); !found {
			return true
		}
		return false
	}, 40*time.Second).Should(gomega.Equal(true))
	settingUpdate := (integrationtest.FakeAviInfraSetting{
		Name:        settingName,
		SeGroupName: "thisisaviref-seGroup",
		Networks:    []string{"thisisaviref-networkName"},
		EnableRhi:   true,
	}).AviInfraSetting()
	settingUpdate.ResourceVersion = "3"
	if _, err := lib.GetCRDClientset().AkoV1alpha1().AviInfraSettings().Update(context.TODO(), settingUpdate, metav1.UpdateOptions{}); err != nil {
		t.Fatalf("error in updating AviInfraSetting: %v", err)
	}

	g.Eventually(func() string {
		setting, _ := lib.GetCRDClientset().AkoV1alpha1().AviInfraSettings().Get(context.TODO(), settingName, metav1.GetOptions{})
		return setting.Status.Status
	}, 40*time.Second).Should(gomega.Equal("Accepted"))
	g.Eventually(func() bool {
		if found, aviModel := objects.SharedAviGraphLister().Get(settingModelName); found && aviModel != nil {
			if nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS(); len(nodes) > 0 {
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
	if _, err := lib.GetCRDClientset().AkoV1alpha1().AviInfraSettings().Update(context.TODO(), settingUpdate, metav1.UpdateOptions{}); err != nil {
		t.Fatalf("error in updating AviInfraSetting: %v", err)
	}

	g.Eventually(func() string {
		setting, _ := lib.GetCRDClientset().AkoV1alpha1().AviInfraSettings().Get(context.TODO(), settingName, metav1.GetOptions{})
		return setting.Status.Status
	}, 40*time.Second).Should(gomega.Equal("Accepted"))

	g.Eventually(func() int {
		if found, aviModel := objects.SharedAviGraphLister().Get(settingModelName); found && aviModel != nil {
			if nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS(); len(nodes) > 0 {
				return len(nodes[0].VSVIPRefs[0].VipNetworks)
			}
		}
		return 0
	}, 40*time.Second).Should(gomega.Equal(3))
	_, aviModel := objects.SharedAviGraphLister().Get(settingModelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
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
	TearDownTestForIngress(t, modelName, settingModelName)
	integrationtest.TeardownIngressClass(t, ingClassName)
}

func TestUpdateIngressClassWithoutInfraSetting(t *testing.T) {
	// update ingressclass (without infrasetting) in ingress
	g := gomega.NewGomegaWithT(t)

	ingClassName1, ingClassName2 := "avi-lb1", "avi-lb2"
	ingressName, ns := "foo-with-class", "default"
	settingName := "my-infrasetting"
	modelName := "admin/cluster--Shared-L7-1"
	settingModelName := "admin/cluster--Shared-L7-my-infrasetting-1"
	secretName := "my-secret"

	SetUpTestForIngress(t, modelName, settingModelName)
	integrationtest.RemoveDefaultIngressClass()
	defer integrationtest.AddDefaultIngressClass()

	integrationtest.SetupAviInfraSetting(t, settingName, "MEDIUM")

	integrationtest.SetupIngressClass(t, ingClassName1, lib.AviIngressController, settingName)
	integrationtest.SetupIngressClass(t, ingClassName2, lib.AviIngressController, "")
	integrationtest.AddSecret(secretName, ns, "tlsCert", "tlsKey")
	ingressCreate := (integrationtest.FakeIngress{
		Name:        ingressName,
		Namespace:   ns,
		ClassName:   ingClassName1,
		DnsNames:    []string{"baz.com", "bar.com"},
		ServiceName: "avisvc",
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
		settingNodes := aviSettingModel.(*avinodes.AviObjectGraph).GetAviVS()
		return settingNodes[0].ServiceEngineGroup
	}, 40*time.Second).Should(gomega.Equal("thisisaviref-my-infrasetting-seGroup"))
	_, aviSettingModel := objects.SharedAviGraphLister().Get(settingModelName)
	settingNodes := aviSettingModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(settingNodes[0].PoolRefs).Should(gomega.HaveLen(1))
	g.Expect(settingNodes[0].PoolRefs[0].Name).Should(gomega.Equal("cluster--my-infrasetting-bar.com_foo-default-foo-with-class"))

	ingressUpdate := (integrationtest.FakeIngress{
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
	_, err = KubeClient.NetworkingV1().Ingresses(ns).Update(context.TODO(), ingressUpdate, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating Ingress: %v", err)
	}

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 40*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes[0].PoolRefs).Should(gomega.HaveLen(1))
	g.Expect(nodes[0].PoolRefs[0].Name).Should(gomega.Equal("cluster--bar.com_foo-default-foo-with-class"))
	g.Expect(nodes[0].ServiceEngineGroup).Should(gomega.Equal("Default-Group"))
	_, aviSettingModel = objects.SharedAviGraphLister().Get(settingModelName)
	settingNodes = aviSettingModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(settingNodes[0].PoolRefs).Should(gomega.HaveLen(0))

	err = KubeClient.NetworkingV1().Ingresses(ns).Delete(context.TODO(), ingressName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	integrationtest.DeleteSecret(secretName, ns)
	integrationtest.TeardownAviInfraSetting(t, settingName)
	TearDownTestForIngress(t, modelName, settingModelName)
	integrationtest.TeardownIngressClass(t, ingClassName1)
	integrationtest.TeardownIngressClass(t, ingClassName2)
	VerifyPoolDeletionFromVsNode(g, modelName)
}

func TestBGPConfigurationWithInfraSetting(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	ingClassName, ingressName, ns, settingName := "avi-lb", "foo-with-class", "default", "my-infrasetting"
	secretName := "my-secret"
	modelName := "admin/cluster--Shared-L7-1"
	settingModelName := "admin/cluster--Shared-L7-my-infrasetting-1"

	SetUpTestForIngress(t, modelName, settingModelName)
	integrationtest.RemoveDefaultIngressClass()
	defer integrationtest.AddDefaultIngressClass()

	integrationtest.SetupAviInfraSetting(t, settingName, "LARGE")
	integrationtest.SetupIngressClass(t, ingClassName, lib.AviIngressController, settingName)
	integrationtest.AddSecret(secretName, ns, "tlsCert", "tlsKey")
	mcache := cache.SharedAviObjCache()
	vsKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--Shared-L7-my-infrasetting-1"}

	ingressCreate := (integrationtest.FakeIngress{
		Name:        ingressName,
		Namespace:   ns,
		ClassName:   ingClassName,
		DnsNames:    []string{"baz.com", "bar.com"},
		ServiceName: "avisvc",
		TlsSecretDNS: map[string][]string{
			secretName: {"baz.com"},
		},
	}).Ingress()
	_, err := KubeClient.NetworkingV1().Ingresses(ns).Create(context.TODO(), ingressCreate, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	sniVsName := "cluster--my-infrasetting-baz.com"
	shardPoolName := "cluster--my-infrasetting-bar.com_foo-default-foo-with-class"

	g.Eventually(func() string {
		if found, aviSettingModel := objects.SharedAviGraphLister().Get(settingModelName); found {
			if settingNodes := aviSettingModel.(*avinodes.AviObjectGraph).GetAviVS(); len(settingNodes) > 0 &&
				len(settingNodes[0].PoolRefs) > 0 &&
				len(settingNodes[0].SniNodes) > 0 {
				return settingNodes[0].SniNodes[0].Name
			}
		}
		return ""
	}, 55*time.Second).Should(gomega.Equal(sniVsName))
	_, aviSettingModel := objects.SharedAviGraphLister().Get(settingModelName)
	settingNodes := aviSettingModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(*settingNodes[0].EnableRhi).Should(gomega.Equal(true))
	g.Expect(settingNodes[0].PoolRefs[0].Name).Should(gomega.Equal(shardPoolName))
	g.Expect(settingNodes[0].VSVIPRefs[0].BGPPeerLabels).Should(gomega.HaveLen(2))
	g.Expect((settingNodes[0].VSVIPRefs[0].BGPPeerLabels)[0]).Should(gomega.ContainSubstring("peer"))

	settingUpdate := (integrationtest.FakeAviInfraSetting{
		Name:          settingName,
		EnableRhi:     false,
		BGPPeerLabels: []string{"peer1", "peer2"},
	}).AviInfraSetting()
	settingUpdate.ResourceVersion = "2"
	if _, err := lib.GetCRDClientset().AkoV1alpha1().AviInfraSettings().Update(context.TODO(), settingUpdate, metav1.UpdateOptions{}); err != nil {
		t.Fatalf("error in updating AviInfraSetting: %v", err)
	}

	// AviInfraSetting is Rejected since enableRhi is false, but the bgpPeerLabels are configured.
	g.Eventually(func() string {
		setting, _ := lib.GetCRDClientset().AkoV1alpha1().AviInfraSettings().Get(context.TODO(), settingName, metav1.GetOptions{})
		return setting.Status.Status
	}, 40*time.Second).Should(gomega.Equal("Rejected"))

	integrationtest.TeardownAviInfraSetting(t, settingName)
	integrationtest.TeardownIngressClass(t, ingClassName)
	err = KubeClient.NetworkingV1().Ingresses(ns).Delete(context.TODO(), ingressName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	integrationtest.DeleteSecret(secretName, ns)
	// Shard VS remains, Pools are moved/removed
	g.Eventually(func() bool {
		sniCache1, found := mcache.VsCacheMeta.AviCacheGet(vsKey)
		sniCacheObj1, _ := sniCache1.(*cache.AviVsCache)
		if found {
			return len(sniCacheObj1.PoolKeyCollection) == 0
		}
		return false
	}, 50*time.Second).Should(gomega.Equal(true))
	TearDownTestForIngress(t, modelName, settingModelName)
}

func TestBGPConfigurationUpdateLabelWithInfraSetting(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	ingClassName, ingressName, ns, settingName := "avi-lb", "foo-with-class", "default", "my-infrasetting"
	secretName := "my-secret"
	modelName := "admin/cluster--Shared-L7-1"
	settingModelName := "admin/cluster--Shared-L7-my-infrasetting-1"
	mcache := cache.SharedAviObjCache()
	vsKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--Shared-L7-my-infrasetting-1"}

	SetUpTestForIngress(t, modelName, settingModelName)
	integrationtest.RemoveDefaultIngressClass()
	defer integrationtest.AddDefaultIngressClass()

	integrationtest.SetupAviInfraSetting(t, settingName, "LARGE")
	integrationtest.SetupIngressClass(t, ingClassName, lib.AviIngressController, settingName)
	integrationtest.AddSecret(secretName, ns, "tlsCert", "tlsKey")

	ingressCreate := (integrationtest.FakeIngress{
		Name:        ingressName,
		Namespace:   ns,
		ClassName:   ingClassName,
		DnsNames:    []string{"baz.com", "bar.com"},
		ServiceName: "avisvc",
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
	if _, err := lib.GetCRDClientset().AkoV1alpha1().AviInfraSettings().Update(context.TODO(), settingUpdate, metav1.UpdateOptions{}); err != nil {
		t.Fatalf("error in updating AviInfraSetting: %v", err)
	}

	g.Eventually(func() int {
		if found, aviSettingModel := objects.SharedAviGraphLister().Get(settingModelName); found {
			if settingNodes := aviSettingModel.(*avinodes.AviObjectGraph).GetAviVS(); len(settingNodes) > 0 && len(settingNodes[0].VSVIPRefs) > 0 {
				return len(settingNodes[0].VSVIPRefs[0].BGPPeerLabels)
			}
		}
		return 0
	}, 55*time.Second).Should(gomega.Equal(3))
	_, aviSettingModel := objects.SharedAviGraphLister().Get(settingModelName)
	settingNodes := aviSettingModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect((settingNodes[0].VSVIPRefs[0].BGPPeerLabels)[0]).Should(gomega.ContainSubstring("peerUPDATE"))

	integrationtest.TeardownAviInfraSetting(t, settingName)
	integrationtest.TeardownIngressClass(t, ingClassName)
	err = KubeClient.NetworkingV1().Ingresses(ns).Delete(context.TODO(), ingressName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	integrationtest.DeleteSecret(secretName, ns)
	// Shard VS remains, Pools are moved/removed
	g.Eventually(func() bool {
		sniCache1, found := mcache.VsCacheMeta.AviCacheGet(vsKey)
		sniCacheObj1, _ := sniCache1.(*cache.AviVsCache)
		if found {
			return len(sniCacheObj1.PoolKeyCollection) == 0
		}
		return false
	}, 50*time.Second).Should(gomega.Equal(true))
	TearDownTestForIngress(t, modelName, settingModelName)
}

func TestCRDWithAviInfraSetting(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	ingClassName, ingressName, ns, settingName := "avi-lb", "foo-with-class", "default", "my-infrasetting"
	secretName := "my-secret"
	modelName := "admin/cluster--Shared-L7-1"
	settingModelName := "admin/cluster--Shared-L7-my-infrasetting-1"
	hrname, rrname := "samplehr-baz", "samplerr-baz"
	mcache := cache.SharedAviObjCache()
	vsKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--Shared-L7-my-infrasetting-1"}
	sniKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--my-infrasetting-baz.com"}
	poolKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--my-infrasetting-default-baz.com_foo-foo-with-class"}

	SetUpTestForIngress(t, modelName, settingModelName)
	integrationtest.RemoveDefaultIngressClass()
	defer integrationtest.AddDefaultIngressClass()

	integrationtest.SetupAviInfraSetting(t, settingName, "LARGE")
	integrationtest.SetupIngressClass(t, ingClassName, lib.AviIngressController, settingName)
	integrationtest.AddSecret(secretName, ns, "tlsCert", "tlsKey")

	integrationtest.SetupHostRule(t, hrname, "baz.com", true)
	integrationtest.SetupHTTPRule(t, rrname, "baz.com", "/")

	ingressCreate := (integrationtest.FakeIngress{
		Name:        ingressName,
		Namespace:   ns,
		ClassName:   ingClassName,
		DnsNames:    []string{"baz.com"},
		ServiceName: "avisvc",
		TlsSecretDNS: map[string][]string{
			secretName: {"baz.com"},
		},
	}).Ingress()
	_, err := KubeClient.NetworkingV1().Ingresses(ns).Create(context.TODO(), ingressCreate, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	g.Eventually(func() string {
		hostrule, _ := CRDClient.AkoV1alpha1().HostRules("default").Get(context.TODO(), hrname, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 30*time.Second).Should(gomega.Equal("Accepted"))
	g.Eventually(func() string {
		httprule, _ := CRDClient.AkoV1alpha1().HTTPRules("default").Get(context.TODO(), rrname, metav1.GetOptions{})
		return httprule.Status.Status
	}, 30*time.Second).Should(gomega.Equal("Accepted"))

	// check for values set in graph layer.
	integrationtest.VerifyMetadataHostRule(g, sniKey, "default/"+hrname, true)
	integrationtest.VerifyMetadataHTTPRule(g, poolKey, "default/"+rrname, true)
	_, aviModel := objects.SharedAviGraphLister().Get(settingModelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(*nodes[0].SniNodes[0].Enabled).To(gomega.Equal(true))
	g.Expect(nodes[0].SniNodes[0].SSLKeyCertAviRef).To(gomega.ContainSubstring("thisisaviref-sslkey"))
	g.Expect(nodes[0].SniNodes[0].WafPolicyRef).To(gomega.ContainSubstring("thisisaviref-waf"))
	g.Expect(nodes[0].SniNodes[0].PoolRefs[0].LbAlgorithm).To(gomega.Equal("LB_ALGORITHM_CONSISTENT_HASH"))
	g.Expect(nodes[0].SniNodes[0].PoolRefs[0].LbAlgorithmHash).To(gomega.Equal("LB_ALGORITHM_CONSISTENT_HASH_SOURCE_IP_ADDRESS"))
	g.Expect(nodes[0].SniNodes[0].PoolRefs[0].SslProfileRef).To(gomega.ContainSubstring("thisisaviref-sslprofile"))

	integrationtest.TeardownHostRule(t, g, sniKey, hrname)
	integrationtest.TeardownHTTPRule(t, rrname)
	integrationtest.TeardownAviInfraSetting(t, settingName)
	integrationtest.TeardownIngressClass(t, ingClassName)
	err = KubeClient.NetworkingV1().Ingresses(ns).Delete(context.TODO(), ingressName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	integrationtest.DeleteSecret(secretName, ns)
	// Shard VS remains, Pools are moved/removed
	g.Eventually(func() bool {
		sniCache1, found := mcache.VsCacheMeta.AviCacheGet(vsKey)
		sniCacheObj1, _ := sniCache1.(*cache.AviVsCache)
		if found {
			return len(sniCacheObj1.PoolKeyCollection) == 0
		}
		return false
	}, 50*time.Second).Should(gomega.Equal(true))
	TearDownTestForIngress(t, modelName, settingModelName)
}
