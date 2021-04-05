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

package hostnameshardtests

import (
	"context"
	"testing"
	"time"

	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	networkingv1beta1 "k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	avinodes "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/nodes"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/tests/integrationtest"
)

type FakeIngressClass struct {
	Name            string
	Controller      string
	AviInfraSetting string
	Default         bool
}

func (ingclass FakeIngressClass) IngressClass() *networkingv1beta1.IngressClass {
	ingressclass := &networkingv1beta1.IngressClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: ingclass.Name,
		},
		Spec: networkingv1beta1.IngressClassSpec{
			Controller: ingclass.Controller,
		},
	}

	if ingclass.Default {
		ingressclass.Annotations = map[string]string{lib.DefaultIngressClassAnnotation: "true"}
	} else {
		ingressclass.Annotations = map[string]string{lib.DefaultIngressClassAnnotation: "false"}
	}

	if ingclass.AviInfraSetting != "" {
		akoGroup := lib.AkoGroup
		ingressclass.Spec.Parameters = &corev1.TypedLocalObjectReference{
			APIGroup: &akoGroup,
			Kind:     lib.AviInfraSetting,
			Name:     ingclass.AviInfraSetting,
		}
	}

	return ingressclass
}

func SetupIngressClass(t *testing.T, ingclassName, controller, infraSetting string) {
	ingclass := FakeIngressClass{
		Name:            ingclassName,
		Controller:      controller,
		Default:         false,
		AviInfraSetting: infraSetting,
	}

	ingClassCreate := ingclass.IngressClass()
	if _, err := KubeClient.NetworkingV1beta1().IngressClasses().Create(context.TODO(), ingClassCreate, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding IngressClass: %v", err)
	}
}

func TeardownIngressClass(t *testing.T, ingClassName string) {
	if err := KubeClient.NetworkingV1beta1().IngressClasses().Delete(context.TODO(), ingClassName, metav1.DeleteOptions{}); err != nil {
		t.Fatalf("error in deleting IngressClass: %v", err)
	}
}

func VerifyVSNodeDeletion(g *gomega.WithT, modelName string) {
	g.Eventually(func() interface{} {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		return aviModel
	}, 30*time.Second).Should(gomega.BeNil())
}

// Ingress - IngressClass mapping tests

func TestAdvL4WrongClassMappingInIngress(t *testing.T) {
	// create ingclass, ingress
	// update wrong mapping of class in ingress, VS deleted
	// fix class in ingress, VS created
	g := gomega.NewGomegaWithT(t)

	ingClassName, ingressName, ns := "avi-lb", "foo-with-class", "default"
	modelName := "admin/cluster--Shared-L7-1"

	SetUpTestForIngress(t, modelName)
	integrationtest.RemoveDefaultIngressClass()
	defer integrationtest.AddDefaultIngressClass()

	SetupIngressClass(t, ingClassName, lib.AviIngressController, "")
	ingressCreate := (integrationtest.FakeIngress{
		Name:        ingressName,
		Namespace:   ns,
		ClassName:   ingClassName,
		DnsNames:    []string{"bar.com"},
		ServiceName: "avisvc",
	}).Ingress()
	_, err := KubeClient.NetworkingV1beta1().Ingresses(ns).Create(context.TODO(), ingressCreate, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 25*time.Second).Should(gomega.Equal(true))

	g.Eventually(func() int {
		ingress, _ := KubeClient.NetworkingV1beta1().Ingresses(ns).Get(context.TODO(), ingressName, metav1.GetOptions{})
		return len(ingress.Status.LoadBalancer.Ingress)
	}, 20*time.Second).Should(gomega.Equal(1))

	ingressUpdate := (integrationtest.FakeIngress{
		Name:        ingressName,
		Namespace:   ns,
		ClassName:   "xyz",
		DnsNames:    []string{"bar.com"},
		ServiceName: "avisvc",
	}).Ingress()
	ingressUpdate.ResourceVersion = "2"
	if _, err := KubeClient.NetworkingV1beta1().Ingresses(ns).Update(context.TODO(), ingressUpdate, metav1.UpdateOptions{}); err != nil {
		t.Fatalf("error in updating Ingress: %v", err)
	}

	g.Eventually(func() int {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return len(nodes[0].PoolRefs)
	}, 60*time.Second).Should(gomega.Equal(0))

	g.Eventually(func() int {
		ingress, _ := KubeClient.NetworkingV1beta1().Ingresses(ns).Get(context.TODO(), ingressName, metav1.GetOptions{})
		return len(ingress.Status.LoadBalancer.Ingress)
	}, 20*time.Second).Should(gomega.Equal(0))

	ingressUpdate2 := (integrationtest.FakeIngress{
		Name:        ingressName,
		Namespace:   ns,
		ClassName:   ingClassName,
		DnsNames:    []string{"bar.com"},
		ServiceName: "avisvc",
	}).Ingress()
	ingressUpdate.ResourceVersion = "3"
	if _, err := KubeClient.NetworkingV1beta1().Ingresses(ns).Update(context.TODO(), ingressUpdate2, metav1.UpdateOptions{}); err != nil {
		t.Fatalf("error in updating Ingress: %v", err)
	}

	// vsNode must come back up
	g.Eventually(func() int {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return len(nodes[0].PoolRefs)
	}, 60*time.Second).Should(gomega.Equal(1))

	g.Eventually(func() int {
		ingress, _ := KubeClient.NetworkingV1beta1().Ingresses(ns).Get(context.TODO(), ingressName, metav1.GetOptions{})
		return len(ingress.Status.LoadBalancer.Ingress)
	}, 25*time.Second).Should(gomega.Equal(1))

	err = KubeClient.NetworkingV1beta1().Ingresses(ns).Delete(context.TODO(), ingressName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	TearDownTestForIngress(t, modelName)
	TeardownIngressClass(t, ingClassName)
	VerifyVSNodeDeletion(g, modelName)
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

	ingClass := (FakeIngressClass{
		Name:       ingClassName,
		Controller: lib.AviIngressController,
		Default:    true,
	}).IngressClass()
	if _, err := KubeClient.NetworkingV1beta1().IngressClasses().Create(context.TODO(), ingClass, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding IngressClass: %v", err)
	}

	// ingress with no IngressClass
	ingressCreate := (integrationtest.FakeIngress{
		Name:        ingressName,
		Namespace:   ns,
		DnsNames:    []string{"bar.com"},
		ServiceName: "avisvc",
	}).Ingress()
	_, err := KubeClient.NetworkingV1beta1().Ingresses(ns).Create(context.TODO(), ingressCreate, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	g.Eventually(func() int {
		ingress, _ := KubeClient.NetworkingV1beta1().Ingresses(ns).Get(context.TODO(), ingressName, metav1.GetOptions{})
		return len(ingress.Status.LoadBalancer.Ingress)
	}, 25*time.Second).Should(gomega.Equal(1))

	ingClass.Annotations = map[string]string{lib.DefaultIngressClassAnnotation: "false"}
	ingClass.ResourceVersion = "2"
	_, err = KubeClient.NetworkingV1beta1().IngressClasses().Update(context.TODO(), ingClass, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating IngressClass: %v", err)
	}

	g.Eventually(func() int {
		ingress, _ := KubeClient.NetworkingV1beta1().Ingresses(ns).Get(context.TODO(), ingressName, metav1.GetOptions{})
		return len(ingress.Status.LoadBalancer.Ingress)
	}, 35*time.Second).Should(gomega.Equal(0))

	err = KubeClient.NetworkingV1beta1().Ingresses(ns).Delete(context.TODO(), ingressName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	TearDownTestForIngress(t, modelName)
	TeardownIngressClass(t, ingClassName)
	VerifyVSNodeDeletion(g, modelName)
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
	SetupIngressClass(t, ingClassName, lib.AviIngressController, settingName)
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
	_, err := KubeClient.NetworkingV1beta1().Ingresses(ns).Create(context.TODO(), ingressCreate, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	shardVsName := "cluster--Shared-L7-my-infrasetting-0"
	sniVsName := "cluster--my-infrasetting-baz.com"
	shardPoolName := "cluster--my-infrasetting-bar.com_foo-default-foo-with-class"
	sniPoolName := "cluster--my-infrasetting-default-baz.com_foo-foo-with-class"

	g.Eventually(func() string {
		found, aviSettingModel := objects.SharedAviGraphLister().Get(settingModelName)
		if found {
			settingNodes := aviSettingModel.(*avinodes.AviObjectGraph).GetAviVS()
			return settingNodes[0].SniNodes[0].Name
		} else {
			return ""
		}
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

	err = KubeClient.NetworkingV1beta1().Ingresses(ns).Delete(context.TODO(), ingressName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	integrationtest.DeleteSecret(secretName, ns)
	integrationtest.TeardownAviInfraSetting(t, settingName)
	TearDownTestForIngress(t, modelName, settingModelName)
	TeardownIngressClass(t, ingClassName)
	VerifyVSNodeDeletion(g, modelName)
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

	SetupIngressClass(t, ingClassName, lib.AviIngressController, "")
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
	_, err := KubeClient.NetworkingV1beta1().Ingresses(ns).Create(context.TODO(), ingressCreate, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 25*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes[0].PoolRefs).Should(gomega.HaveLen(1))
	g.Expect(nodes[0].SniNodes).Should(gomega.HaveLen(1))

	integrationtest.SetupAviInfraSetting(t, settingName, "SMALL")
	settingModelName := "admin/cluster--Shared-L7-my-infrasetting-0"

	ingClassUpdate := (FakeIngressClass{
		Name:            ingClassName,
		Controller:      lib.AviIngressController,
		AviInfraSetting: settingName,
		Default:         false,
	}).IngressClass()
	ingClassUpdate.ResourceVersion = "2"
	_, err = KubeClient.NetworkingV1beta1().IngressClasses().Update(context.TODO(), ingClassUpdate, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error updating IngressClass")
	}

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(settingModelName)
		return found
	}, 25*time.Second).Should(gomega.Equal(true))

	_, aviSettingModel := objects.SharedAviGraphLister().Get(settingModelName)
	settingNodes := aviSettingModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(settingNodes[0].PoolRefs).Should(gomega.HaveLen(1))
	g.Expect(settingNodes[0].PoolRefs[0].Name).Should(gomega.Equal("cluster--my-infrasetting-bar.com_foo-default-foo-with-class"))
	g.Expect(settingNodes[0].SniNodes).Should(gomega.HaveLen(1))
	g.Expect(settingNodes[0].SniNodes[0].Name).Should(gomega.Equal("cluster--my-infrasetting-baz.com"))
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes[0].PoolRefs).Should(gomega.HaveLen(0))

	ingClassUpdate = (FakeIngressClass{
		Name:       ingClassName,
		Controller: lib.AviIngressController,
		Default:    false,
	}).IngressClass()
	ingClassUpdate.ResourceVersion = "3"
	_, err = KubeClient.NetworkingV1beta1().IngressClasses().Update(context.TODO(), ingClassUpdate, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error updating IngressClass")
	}

	err = KubeClient.NetworkingV1beta1().Ingresses(ns).Delete(context.TODO(), ingressName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	integrationtest.DeleteSecret(secretName, ns)
	integrationtest.TeardownAviInfraSetting(t, settingName)
	TearDownTestForIngress(t, modelName, settingModelName)
	TeardownIngressClass(t, ingClassName)
	VerifyVSNodeDeletion(g, modelName)
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

	SetupIngressClass(t, ingClassName, lib.AviIngressController, settingName1)
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
	_, err := KubeClient.NetworkingV1beta1().Ingresses(ns).Create(context.TODO(), ingressCreate, metav1.CreateOptions{})
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
	}, 25*time.Second).Should(gomega.Equal("cluster--my-infrasetting-bar.com_foo-default-foo-with-class"))

	ingClassUpdate := (FakeIngressClass{
		Name:            ingClassName,
		Controller:      lib.AviIngressController,
		Default:         false,
		AviInfraSetting: settingName2,
	}).IngressClass()
	ingClassUpdate.ResourceVersion = "2"
	_, err = KubeClient.NetworkingV1beta1().IngressClasses().Update(context.TODO(), ingClassUpdate, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error updating IngressClass")
	}

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(settingModelName2)
		return found
	}, 25*time.Second).Should(gomega.Equal(true))
	_, aviSettingModel := objects.SharedAviGraphLister().Get(settingModelName2)
	settingNodes := aviSettingModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(settingNodes[0].PoolRefs).Should(gomega.HaveLen(1))
	g.Expect(settingNodes[0].PoolRefs[0].Name).Should(gomega.Equal("cluster--my-infrasetting2-bar.com_foo-default-foo-with-class"))

	err = KubeClient.NetworkingV1beta1().Ingresses(ns).Delete(context.TODO(), ingressName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	integrationtest.DeleteSecret(secretName, ns)
	integrationtest.TeardownAviInfraSetting(t, settingName1)
	integrationtest.TeardownAviInfraSetting(t, settingName2)
	TearDownTestForIngress(t, modelName, settingModelName1)
	TearDownTestForIngress(t, modelName, settingModelName2)
	TeardownIngressClass(t, ingClassName)
	VerifyVSNodeDeletion(g, modelName)
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

	SetupIngressClass(t, ingClassName, lib.AviIngressController, settingName)
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
	_, err := KubeClient.NetworkingV1beta1().Ingresses(ns).Create(context.TODO(), ingressCreate, metav1.CreateOptions{})
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
	_, err = KubeClient.NetworkingV1beta1().Ingresses(ns).Update(context.TODO(), ingressUpdate, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating Ingress: %v", err)
	}

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(settingModelName)
		return found
	}, 25*time.Second).Should(gomega.Equal(true))

	_, aviSettingModel := objects.SharedAviGraphLister().Get(settingModelName)
	settingNodes := aviSettingModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Eventually(func() int {
		found, _ := objects.SharedAviGraphLister().Get(settingModelName)
		if found {
			settingNodes := aviSettingModel.(*avinodes.AviObjectGraph).GetAviVS()
			return len(settingNodes[0].PoolRefs)
		} else {
			return 0
		}
	}, 25*time.Second).Should(gomega.Equal(1))
	g.Expect(settingNodes[0].PoolRefs[0].Name).Should(gomega.Equal("cluster--my-infrasetting-bar.com_foo-default-foo-with-class"))

	err = KubeClient.NetworkingV1beta1().Ingresses(ns).Delete(context.TODO(), ingressName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}

	g.Eventually(func() int {
		found, aviSettingModel := objects.SharedAviGraphLister().Get(settingModelName)
		nodes := aviSettingModel.(*avinodes.AviObjectGraph).GetAviVS()
		if found && len(nodes) > 0 {
			return len(nodes[0].PoolRefs)
		}
		return -1
	}, 25*time.Second).Should(gomega.Equal(0))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	g.Expect(aviModel).Should(gomega.BeNil())

	integrationtest.DeleteSecret(secretName, ns)
	integrationtest.TeardownAviInfraSetting(t, settingName)
	TearDownTestForIngress(t, modelName, settingModelName)
	TeardownIngressClass(t, ingClassName)
	VerifyVSNodeDeletion(g, modelName)
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

	SetupIngressClass(t, ingClassName1, lib.AviIngressController, settingName1)
	SetupIngressClass(t, ingClassName2, lib.AviIngressController, settingName2)
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
	_, err := KubeClient.NetworkingV1beta1().Ingresses(ns).Create(context.TODO(), ingressCreate, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(settingModelName1)
		return found
	}, 15*time.Second).Should(gomega.Equal(true))
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
	_, err = KubeClient.NetworkingV1beta1().Ingresses(ns).Update(context.TODO(), ingressUpdate, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating Ingress: %v", err)
	}

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(settingModelName2)
		return found
	}, 15*time.Second).Should(gomega.Equal(true))
	_, aviSettingModel2 := objects.SharedAviGraphLister().Get(settingModelName2)
	settingNodes2 := aviSettingModel2.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(settingNodes2[0].PoolRefs).Should(gomega.HaveLen(1))
	g.Expect(settingNodes2[0].PoolRefs[0].Name).Should(gomega.Equal("cluster--my-infrasetting2-bar.com_foo-default-foo-with-class"))
	g.Expect(settingNodes2[0].ServiceEngineGroup).Should(gomega.Equal("thisisaviref-my-infrasetting2-seGroup"))
	_, aviSettingModel1 = objects.SharedAviGraphLister().Get(settingModelName1)
	settingNodes1 = aviSettingModel1.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(settingNodes1[0].PoolRefs).Should(gomega.HaveLen(0))

	err = KubeClient.NetworkingV1beta1().Ingresses(ns).Delete(context.TODO(), ingressName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	integrationtest.DeleteSecret(secretName, ns)
	integrationtest.TeardownAviInfraSetting(t, settingName1)
	integrationtest.TeardownAviInfraSetting(t, settingName2)
	TearDownTestForIngress(t, modelName, settingModelName1, settingModelName2)
	TeardownIngressClass(t, ingClassName1)
	TeardownIngressClass(t, ingClassName2)
	VerifyVSNodeDeletion(g, settingModelName2)
}

func TestUpdateIngressClassWithoutInfraSetting(t *testing.T) {
	// update ingressclass (without infrasetting) in ingress
	g := gomega.NewGomegaWithT(t)

	ingClassName1, ingClassName2 := "avi-lb1", "avi-lb2"
	ingressName, ns := "foo-with-class", "default"
	settingName := "my-infrasetting"
	modelName := "admin/cluster--Shared-L7-1"
	secretName := "my-secret"

	SetUpTestForIngress(t, modelName)
	integrationtest.RemoveDefaultIngressClass()
	defer integrationtest.AddDefaultIngressClass()

	integrationtest.SetupAviInfraSetting(t, settingName, "MEDIUM")
	settingModelName := "admin/cluster--Shared-L7-my-infrasetting-1"

	SetupIngressClass(t, ingClassName1, lib.AviIngressController, settingName)
	SetupIngressClass(t, ingClassName2, lib.AviIngressController, "")
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
	_, err := KubeClient.NetworkingV1beta1().Ingresses(ns).Create(context.TODO(), ingressCreate, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(settingModelName)
		return found
	}, 15*time.Second).Should(gomega.Equal(true))
	_, aviSettingModel := objects.SharedAviGraphLister().Get(settingModelName)
	settingNodes := aviSettingModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(settingNodes[0].PoolRefs).Should(gomega.HaveLen(1))
	g.Expect(settingNodes[0].ServiceEngineGroup).Should(gomega.Equal("thisisaviref-my-infrasetting-seGroup"))
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
	_, err = KubeClient.NetworkingV1beta1().Ingresses(ns).Update(context.TODO(), ingressUpdate, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating Ingress: %v", err)
	}

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 15*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes[0].PoolRefs).Should(gomega.HaveLen(1))
	g.Expect(nodes[0].PoolRefs[0].Name).Should(gomega.Equal("cluster--bar.com_foo-default-foo-with-class"))
	g.Expect(nodes[0].ServiceEngineGroup).Should(gomega.Equal("Default-Group"))
	_, aviSettingModel = objects.SharedAviGraphLister().Get(settingModelName)
	settingNodes = aviSettingModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(settingNodes[0].PoolRefs).Should(gomega.HaveLen(0))

	err = KubeClient.NetworkingV1beta1().Ingresses(ns).Delete(context.TODO(), ingressName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	integrationtest.DeleteSecret(secretName, ns)
	integrationtest.TeardownAviInfraSetting(t, settingName)
	TearDownTestForIngress(t, modelName, settingModelName)
	TeardownIngressClass(t, ingClassName1)
	TeardownIngressClass(t, ingClassName2)
	VerifyVSNodeDeletion(g, modelName)
}
