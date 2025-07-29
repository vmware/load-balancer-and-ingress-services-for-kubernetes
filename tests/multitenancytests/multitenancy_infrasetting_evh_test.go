package multitenancytests

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/cache"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/k8s"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	avinodes "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/nodes"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"
	utils "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/tests/ingresstests"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/tests/integrationtest"

	"github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func tearDownTestForIngress(t *testing.T, svcname string) {
	integrationtest.DelSVC(t, "default", svcname)
	integrationtest.DelEPS(t, "default", svcname)
}

func verifyEvhNodeDeletionFromVsNode(g *gomega.WithT, modelName string, parentVSKey, evhVsKey cache.NamespaceName) {
	g.Eventually(func() bool {
		if found, aviModel := objects.SharedAviGraphLister().Get(modelName); found {
			if nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS(); len(nodes) > 0 {
				return len(nodes[0].EvhNodes) == 0
			}
		}
		return true
	}, 50*time.Second).Should(gomega.Equal(true))

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

func TestMultiTenancyWithNSAviInfraSettingEVH(t *testing.T) {
	// create secure and insecure host ingress, connect with infrasetting
	// check for names of all Avi objects
	g := gomega.NewGomegaWithT(t)
	os.Setenv("ENABLE_EVH", "true")

	time.Sleep(time.Second * 5)
	ingestionQueue := utils.SharedWorkQueue().GetQueueByName(utils.ObjectIngestionLayer)
	ingestionQueue.SyncFunc = syncFromIngestionLayerWrapper
	defer func() { ingestionQueue.SyncFunc = k8s.SyncFromIngestionLayer }()

	ingClassName := objNameMap.GenerateName("avi-lb")
	ingressName := objNameMap.GenerateName("foo-with-class")
	ns := "default"
	settingName := objNameMap.GenerateName("my-infrasetting")
	secretName := objNameMap.GenerateName("my-secret")
	modelName := "admin/cluster--Shared-L7-EVH-1"
	settingModelName := "nonadmin/cluster--Shared-L7-EVH-0"
	svcName := objNameMap.GenerateName("avisvc")
	ingresstests.SetUpTestForIngress(t, svcName, modelName)

	vsKey := cache.NamespaceName{Namespace: "nonadmin", Name: "cluster--Shared-L7-EVH-0"}
	evhKey := cache.NamespaceName{Namespace: "nonadmin", Name: lib.Encode("cluster--baz.com", lib.EVHVS)}
	integrationtest.SetupAviInfraSetting(t, settingName, "SMALL")
	integrationtest.AnnotateAKONamespaceWithInfraSetting(t, ns, settingName)
	integrationtest.AnnotateNamespaceWithTenant(t, ns, "nonadmin")
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
	tearDownTestForIngress(t, svcName)
	integrationtest.TeardownIngressClass(t, ingClassName)
	waitAndVerify(t, ingClassName)
	verifyEvhNodeDeletionFromVsNode(g, settingModelName, vsKey, evhKey)
}

func TestMultiTenancyWithIngressClassAviInfraSettingEVH(t *testing.T) {
	// create secure and insecure host ingress, connect with infrasetting
	// check for names of all Avi objects
	g := gomega.NewGomegaWithT(t)

	ingClassName := objNameMap.GenerateName("avi-lb")
	ingressName := objNameMap.GenerateName("foo-with-class")
	ns := "default"
	settingName := objNameMap.GenerateName("my-infrasetting")
	secretName := objNameMap.GenerateName("my-secret")
	modelName := "admin/cluster--Shared-L7-EVH-1"

	svcName := objNameMap.GenerateName("avisvc")
	time.Sleep(time.Second * 5)
	ingestionQueue := utils.SharedWorkQueue().GetQueueByName(utils.ObjectIngestionLayer)
	ingestionQueue.SyncFunc = syncFromIngestionLayerWrapper
	defer func() { ingestionQueue.SyncFunc = k8s.SyncFromIngestionLayer }()

	ingresstests.SetUpTestForIngress(t, svcName, modelName)

	settingModelName := "nonadmin/cluster--Shared-L7-EVH-" + settingName + "-0"
	vsKey := cache.NamespaceName{Namespace: "nonadmin", Name: "cluster--Shared-L7-EVH-" + settingName + "-0"}
	evhKey := cache.NamespaceName{Namespace: "nonadmin", Name: lib.Encode("cluster--"+settingName+"-baz.com", lib.EVHVS)}
	integrationtest.SetupAviInfraSetting(t, settingName, "SMALL")
	integrationtest.SetupIngressClass(t, ingClassName, lib.AviIngressController, settingName)
	integrationtest.AnnotateNamespaceWithTenant(t, ns, "nonadmin")
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
	integrationtest.RemoveAnnotateAKONamespaceWithInfraSetting(t, ns)
	tearDownTestForIngress(t, svcName)
	integrationtest.TeardownIngressClass(t, ingClassName)
	waitAndVerify(t, ingClassName)
	verifyEvhNodeDeletionFromVsNode(g, settingModelName, vsKey, evhKey)
}

func TestMultiTenancyWithInfraSettingAdditionEVH(t *testing.T) {
	// create infrasettings, update infrasetting with a tenant
	// new model creation should happen, old model should get deleted
	g := gomega.NewGomegaWithT(t)

	ingClassName := objNameMap.GenerateName("avi-lb")
	ingressName := objNameMap.GenerateName("foo-with-class")
	ns := "default"
	settingName := objNameMap.GenerateName("my-infrasetting")
	secretName := objNameMap.GenerateName("my-secret")
	modelName := "admin/cluster--Shared-L7-EVH-1"
	settingModelName := "nonadmin/cluster--Shared-L7-EVH-0"

	svcName := objNameMap.GenerateName("avisvc")
	time.Sleep(time.Second * 5)
	ingestionQueue := utils.SharedWorkQueue().GetQueueByName(utils.ObjectIngestionLayer)
	ingestionQueue.SyncFunc = syncFromIngestionLayerWrapper
	defer func() { ingestionQueue.SyncFunc = k8s.SyncFromIngestionLayer }()

	ingresstests.SetUpTestForIngress(t, svcName, modelName, settingModelName)

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

	secureVsName := "cluster--baz.com"
	insecureVsName := "cluster--bar.com"
	insecurePoolName := "cluster--default-bar.com_foo-" + ingressName + "-" + svcName
	securePoolName := "cluster--default-baz.com_foo-" + ingressName + "-" + svcName
	insecurePGName := "cluster--default-bar.com_foo-" + ingressName
	securePGName := "cluster--default-baz.com_foo-" + ingressName

	g.Eventually(func() int {
		if found, aviSettingModel := objects.SharedAviGraphLister().Get(modelName); found {
			if settingNodes := aviSettingModel.(*avinodes.AviObjectGraph).GetAviEvhVS(); len(settingNodes) > 0 {
				return len(settingNodes[0].EvhNodes)
			}
		}
		return 0
	}, 55*time.Second).Should(gomega.Equal(2))
	time.Sleep(5 * time.Second)
	_, aviSettingModel := objects.SharedAviGraphLister().Get(modelName)
	settingNodes := aviSettingModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(settingNodes[0].ServiceEngineGroup).Should(gomega.Equal(lib.GetSEGName()))
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

	integrationtest.SetupAviInfraSetting(t, settingName, "SMALL")
	integrationtest.AnnotateAKONamespaceWithInfraSetting(t, ns, settingName)
	integrationtest.AnnotateNamespaceWithTenant(t, ns, "nonadmin")

	g.Eventually(func() bool {
		if found, aviSettingModel := objects.SharedAviGraphLister().Get(modelName); found {
			if settingNodes := aviSettingModel.(*avinodes.AviObjectGraph).GetAviEvhVS(); len(settingNodes) > 0 {
				return len(settingNodes[0].EvhNodes) == 0 && len(settingNodes[0].PoolRefs) == 0
			}
		}
		return false
	}, 55*time.Second).Should(gomega.Equal(true))

	g.Eventually(func() int {
		if found, aviSettingModel := objects.SharedAviGraphLister().Get(settingModelName); found {
			if settingNodes := aviSettingModel.(*avinodes.AviObjectGraph).GetAviEvhVS(); len(settingNodes) > 0 {
				return len(settingNodes[0].EvhNodes)
			}
		}
		return 0
	}, 55*time.Second).Should(gomega.Equal(2))
	time.Sleep(5 * time.Second)
	_, aviSettingModel = objects.SharedAviGraphLister().Get(settingModelName)
	settingNodes = aviSettingModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
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
	g.Expect(settingNodes[0].VSVIPRefs[0].T1Lr).Should(gomega.Equal("avi-domain-c9:1234"))

	err = KubeClient.NetworkingV1().Ingresses(ns).Delete(context.TODO(), ingressName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	integrationtest.DeleteSecret(secretName, ns)
	tearDownTestForIngress(t, svcName)
	integrationtest.TeardownIngressClass(t, ingClassName)
	waitAndVerify(t, ingClassName)
	verifyPoolDeletionFromVsNode(g, modelName)
	integrationtest.RemoveAnnotateAKONamespaceWithInfraSetting(t, ns)
	integrationtest.TeardownAviInfraSetting(t, settingName)
}

func TestMultiTenancyWithTenantDeannotationInNSEVH(t *testing.T) {
	// create an ingress, infrasetting and annotate a namespace with infrasetting
	// graph layer objects should come up with correct tenant
	// delete the Infrasetting annotation from the namespace, old model should be deleted
	// new model in default tenant should get created
	g := gomega.NewGomegaWithT(t)

	ingClassName := objNameMap.GenerateName("avi-lb")
	ingressName := objNameMap.GenerateName("foo-with-class")
	ns := "default"
	settingName := objNameMap.GenerateName("my-infrasetting")
	secretName := objNameMap.GenerateName("my-secret")
	modelName := "admin/cluster--Shared-L7-EVH-1"
	settingModelName := "nonadmin/cluster--Shared-L7-EVH-0"

	svcName := objNameMap.GenerateName("avisvc")
	time.Sleep(time.Second * 5)
	ingestionQueue := utils.SharedWorkQueue().GetQueueByName(utils.ObjectIngestionLayer)
	ingestionQueue.SyncFunc = syncFromIngestionLayerWrapper
	defer func() { ingestionQueue.SyncFunc = k8s.SyncFromIngestionLayer }()

	ingresstests.SetUpTestForIngress(t, svcName, modelName)

	vsKey := cache.NamespaceName{Namespace: "nonadmin", Name: "cluster--Shared-L7-EVH-0"}
	evhKey := cache.NamespaceName{Namespace: "nonadmin", Name: lib.Encode("cluster--baz.com", lib.EVHVS)}
	integrationtest.SetupAviInfraSetting(t, settingName, "SMALL")
	integrationtest.AnnotateAKONamespaceWithInfraSetting(t, ns, settingName)
	integrationtest.AnnotateNamespaceWithTenant(t, ns, "nonadmin")
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

	integrationtest.RemoveAnnotateAKONamespaceWithInfraSetting(t, ns)
	g.Eventually(func() bool {
		if found, aviSettingModel := objects.SharedAviGraphLister().Get(settingModelName); found {
			if settingNodes := aviSettingModel.(*avinodes.AviObjectGraph).GetAviEvhVS(); len(settingNodes) > 0 {
				return len(settingNodes[0].EvhNodes) == 0
			}
		}
		return true
	}, 55*time.Second).Should(gomega.Equal(true))

	secureVsName = "cluster--baz.com"
	insecureVsName = "cluster--bar.com"
	insecurePoolName = "cluster--default-bar.com_foo-" + ingressName + "-" + svcName
	securePoolName = "cluster--default-baz.com_foo-" + ingressName + "-" + svcName
	insecurePGName = "cluster--default-bar.com_foo-" + ingressName
	securePGName = "cluster--default-baz.com_foo-" + ingressName

	g.Eventually(func() int {
		if found, aviSettingModel := objects.SharedAviGraphLister().Get(modelName); found {
			if settingNodes := aviSettingModel.(*avinodes.AviObjectGraph).GetAviEvhVS(); len(settingNodes) > 0 {
				return len(settingNodes[0].EvhNodes)
			}
		}
		return 0
	}, 55*time.Second).Should(gomega.Equal(2))
	time.Sleep(5 * time.Second)
	_, aviSettingModel = objects.SharedAviGraphLister().Get(modelName)
	settingNodes = aviSettingModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(settingNodes[0].ServiceEngineGroup).Should(gomega.Equal(lib.GetSEGName()))
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
	tearDownTestForIngress(t, svcName)
	integrationtest.TeardownIngressClass(t, ingClassName)
	waitAndVerify(t, ingClassName)
	integrationtest.TeardownAviInfraSetting(t, settingName)
	verifyEvhNodeDeletionFromVsNode(g, settingModelName, vsKey, evhKey)
	os.Unsetenv("ENABLE_EVH")
}
