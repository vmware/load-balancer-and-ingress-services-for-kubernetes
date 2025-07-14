package multitenancytests

import (
	"context"
	"testing"
	"time"

	"github.com/onsi/gomega"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/cache"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	avinodes "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/nodes"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"
	utils "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/tests/integrationtest"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func tearDownTestForSvcLB(t *testing.T, g *gomega.GomegaWithT, tenant, svcName string) {
	modelName := tenant + "/" + "cluster--red-ns-" + svcName
	objects.SharedAviGraphLister().Delete(modelName)
	integrationtest.DelSVC(t, "red-ns", svcName)
	integrationtest.DelEPS(t, "red-ns", svcName)
	mcache := cache.SharedAviObjCache()
	vsKey := cache.NamespaceName{Namespace: tenant, Name: "cluster--red-ns-" + svcName}
	g.Eventually(func() bool {
		_, found := mcache.VsCacheMeta.AviCacheGet(vsKey)
		return found
	}, 5*time.Second).Should(gomega.Equal(false))
}

func TestMultiTenancyWithNSAviInfraSetting(t *testing.T) {
	// create svc, infrasetting, annotate infrasetting to NS
	// graph layer objects should come up in the right tenant
	g := gomega.NewGomegaWithT(t)
	svcName := objNameMap.GenerateName("testsvc")
	objects.SharedAviGraphLister().Delete("admin/cluster--red-ns-" + svcName)

	ns := "red-ns"
	settingName := objNameMap.GenerateName("my-infrasetting")
	integrationtest.SetupAviInfraSetting(t, settingName, "DEDICATED")
	integrationtest.AnnotateAKONamespaceWithInfraSetting(t, ns, settingName)
	integrationtest.AnnotateNamespaceWithTenant(t, ns, "nonadmin")

	modelName := "nonadmin/cluster--red-ns-" + svcName
	svcExample := (integrationtest.FakeService{
		Name:         svcName,
		Namespace:    ns,
		Type:         v1.ServiceTypeLoadBalancer,
		ServicePorts: []integrationtest.Serviceport{{PortName: "foo1", Protocol: "TCP", PortNumber: 8080, TargetPort: intstr.FromInt(8080)}},
	}).Service()
	_, err := KubeClient.CoreV1().Services(ns).Create(context.TODO(), svcExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in creating Service: %v", err)
	}
	integrationtest.CreateEPS(t, ns, svcName, false, false, "1.1.1")
	integrationtest.PollForCompletion(t, modelName, 5)

	g.Eventually(func() bool {
		if found, aviModel := objects.SharedAviGraphLister().Get(modelName); found && aviModel != nil {
			if nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS(); len(nodes) > 0 {
				return nodes[0].ServiceEngineGroup == "thisisaviref-"+settingName+"-seGroup" &&
					len(nodes[0].VSVIPRefs[0].VipNetworks) > 0 &&
					nodes[0].VSVIPRefs[0].VipNetworks[0].NetworkName == "thisisaviref-"+settingName+"-networkName" &&
					*nodes[0].EnableRhi
			}
		}
		return false
	}, 45*time.Second).Should(gomega.Equal(true))

	integrationtest.RemoveAnnotateAKONamespaceWithInfraSetting(t, ns)
	integrationtest.TeardownAviInfraSetting(t, settingName)
	tearDownTestForSvcLB(t, g, "nonadmin", svcName)
}

func TestMultiTenancyWithSvcAnnotatedAviInfraSetting(t *testing.T) {
	// create ingress, ingressclas, infrasetting, add infrasetting to ingressclass
	// graph layer objects should come up in the right tenant
	// delete the ingress, graph layer nodes should get deleted
	g := gomega.NewGomegaWithT(t)
	svcName := objNameMap.GenerateName("testsvc")
	objects.SharedAviGraphLister().Delete("admin/cluster--red-ns-" + svcName)

	ns := "red-ns"
	settingName := objNameMap.GenerateName("my-infrasetting")
	integrationtest.SetupAviInfraSetting(t, settingName, "DEDICATED")
	integrationtest.AnnotateNamespaceWithTenant(t, ns, "nonadmin")

	modelName := "nonadmin/cluster--red-ns-" + svcName
	svcExample := (integrationtest.FakeService{
		Name:         svcName,
		Namespace:    ns,
		Type:         v1.ServiceTypeLoadBalancer,
		ServicePorts: []integrationtest.Serviceport{{PortName: "foo1", Protocol: "TCP", PortNumber: 8080, TargetPort: intstr.FromInt(8080)}},
	}).Service()
	svcExample.Annotations = map[string]string{lib.InfraSettingNameAnnotation: settingName}
	_, err := KubeClient.CoreV1().Services(ns).Create(context.TODO(), svcExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in creating Service: %v", err)
	}
	integrationtest.CreateEPS(t, ns, svcName, false, false, "1.1.1")
	integrationtest.PollForCompletion(t, modelName, 5)

	g.Eventually(func() bool {
		if found, aviModel := objects.SharedAviGraphLister().Get(modelName); found && aviModel != nil {
			if nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS(); len(nodes) > 0 {
				return nodes[0].ServiceEngineGroup == "thisisaviref-"+settingName+"-seGroup" &&
					len(nodes[0].VSVIPRefs[0].VipNetworks) > 0 &&
					nodes[0].VSVIPRefs[0].VipNetworks[0].NetworkName == "thisisaviref-"+settingName+"-networkName" &&
					*nodes[0].EnableRhi
			}
		}
		return false
	}, 45*time.Second).Should(gomega.Equal(true))

	integrationtest.TeardownAviInfraSetting(t, settingName)
	integrationtest.RemoveAnnotateAKONamespaceWithInfraSetting(t, ns)
	tearDownTestForSvcLB(t, g, "nonadmin", svcName)
}

func TestMultiTenancyWithInfraSettingAddition(t *testing.T) {
	// create infrasettings, update infrasetting with a tenant
	// new model creation should happen, old model should get deleted
	g := gomega.NewGomegaWithT(t)
	svcName := objNameMap.GenerateName("testsvc")
	settingName := objNameMap.GenerateName("my-infrasetting")
	objects.SharedAviGraphLister().Delete("admin/cluster--red-ns-" + svcName)

	ns := "red-ns"

	modelName := "admin/cluster--red-ns-" + svcName
	svcExample := (integrationtest.FakeService{
		Name:         svcName,
		Namespace:    ns,
		Type:         v1.ServiceTypeLoadBalancer,
		ServicePorts: []integrationtest.Serviceport{{PortName: "foo1", Protocol: "TCP", PortNumber: 8080, TargetPort: intstr.FromInt(8080)}},
	}).Service()
	_, err := KubeClient.CoreV1().Services(ns).Create(context.TODO(), svcExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in creating Service: %v", err)
	}
	integrationtest.CreateEPS(t, ns, svcName, false, false, "1.1.1")
	integrationtest.PollForCompletion(t, modelName, 5)

	netList := utils.GetVipNetworkList()
	g.Eventually(func() bool {
		if found, aviModel := objects.SharedAviGraphLister().Get(modelName); found && aviModel != nil {
			if nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS(); len(nodes) > 0 {
				return nodes[0].ServiceEngineGroup == lib.GetSEGName() &&
					len(nodes[0].VSVIPRefs[0].VipNetworks) > 0 &&
					nodes[0].VSVIPRefs[0].VipNetworks[0].NetworkName == netList[0].NetworkName &&
					!*nodes[0].EnableRhi
			}
		}
		return false
	}, 45*time.Second).Should(gomega.Equal(true))

	integrationtest.SetupAviInfraSetting(t, settingName, "DEDICATED")
	integrationtest.AnnotateAKONamespaceWithInfraSetting(t, ns, settingName)
	integrationtest.AnnotateNamespaceWithTenant(t, ns, "nonadmin")

	g.Eventually(func() bool {
		if found, aviModel := objects.SharedAviGraphLister().Get(modelName); found && aviModel == nil {
			return true
		}
		return false
	}, 45*time.Second).Should(gomega.Equal(true))

	newModelName := "nonadmin/cluster--red-ns-" + svcName
	g.Eventually(func() bool {
		if found, aviModel := objects.SharedAviGraphLister().Get(newModelName); found && aviModel != nil {
			if nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS(); len(nodes) > 0 {
				return nodes[0].ServiceEngineGroup == "thisisaviref-"+settingName+"-seGroup" &&
					len(nodes[0].VSVIPRefs[0].VipNetworks) > 0 &&
					nodes[0].VSVIPRefs[0].VipNetworks[0].NetworkName == "thisisaviref-"+settingName+"-networkName" &&
					*nodes[0].EnableRhi
			}
		}
		return false
	}, 45*time.Second).Should(gomega.Equal(true))

	integrationtest.TeardownAviInfraSetting(t, settingName)
	integrationtest.RemoveAnnotateAKONamespaceWithInfraSetting(t, ns)
	tearDownTestForSvcLB(t, g, "nonadmin", svcName)
}

func TestMultiTenancyWithTenantDeannotationInNS(t *testing.T) {
	// create a svc, infrasetting and annotate a namespace with infrasetting
	// graph layer objects should come up with correct tenant
	// delete the Infrasetting annotation from the namespace, old model should be deleted
	// new model in default tenant should get created
	g := gomega.NewGomegaWithT(t)
	svcName := objNameMap.GenerateName("testsvc")
	objects.SharedAviGraphLister().Delete("admin/cluster--red-ns-" + svcName)

	ns := "red-ns"
	settingName := objNameMap.GenerateName("my-infrasetting")
	integrationtest.SetupAviInfraSetting(t, settingName, "DEDICATED")
	integrationtest.AnnotateAKONamespaceWithInfraSetting(t, ns, settingName)
	integrationtest.AnnotateNamespaceWithTenant(t, ns, "nonadmin")

	modelName := "nonadmin/cluster--red-ns-" + svcName
	svcExample := (integrationtest.FakeService{
		Name:         svcName,
		Namespace:    ns,
		Type:         v1.ServiceTypeLoadBalancer,
		ServicePorts: []integrationtest.Serviceport{{PortName: "foo1", Protocol: "TCP", PortNumber: 8080, TargetPort: intstr.FromInt(8080)}},
	}).Service()
	_, err := KubeClient.CoreV1().Services(ns).Create(context.TODO(), svcExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in creating Service: %v", err)
	}
	integrationtest.CreateEPS(t, ns, svcName, false, false, "1.1.1")
	integrationtest.PollForCompletion(t, modelName, 5)

	g.Eventually(func() bool {
		if found, aviModel := objects.SharedAviGraphLister().Get(modelName); found && aviModel != nil {
			if nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS(); len(nodes) > 0 {
				return nodes[0].ServiceEngineGroup == "thisisaviref-"+settingName+"-seGroup" &&
					len(nodes[0].VSVIPRefs[0].VipNetworks) > 0 &&
					nodes[0].VSVIPRefs[0].VipNetworks[0].NetworkName == "thisisaviref-"+settingName+"-networkName" &&
					*nodes[0].EnableRhi
			}
		}
		return false
	}, 45*time.Second).Should(gomega.Equal(true))

	integrationtest.RemoveAnnotateAKONamespaceWithInfraSetting(t, ns)
	g.Eventually(func() bool {
		if found, aviModel := objects.SharedAviGraphLister().Get(modelName); found && aviModel == nil {
			return true
		}
		return false
	}, 45*time.Second).Should(gomega.Equal(true))

	modelName = "admin/cluster--red-ns-" + svcName
	netList := utils.GetVipNetworkList()
	g.Eventually(func() bool {
		if found, aviModel := objects.SharedAviGraphLister().Get(modelName); found && aviModel != nil {
			if nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS(); len(nodes) > 0 {
				return nodes[0].ServiceEngineGroup == lib.GetSEGName() &&
					len(nodes[0].VSVIPRefs[0].VipNetworks) > 0 &&
					nodes[0].VSVIPRefs[0].VipNetworks[0].NetworkName == netList[0].NetworkName &&
					!*nodes[0].EnableRhi
			}
		}
		return false
	}, 45*time.Second).Should(gomega.Equal(true))

	integrationtest.TeardownAviInfraSetting(t, settingName)
	tearDownTestForSvcLB(t, g, "admin", svcName)
}
