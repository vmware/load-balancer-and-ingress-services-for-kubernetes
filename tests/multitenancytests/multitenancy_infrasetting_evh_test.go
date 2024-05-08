package multitenancytests

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/cache"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	avinodes "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/nodes"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/tests/ingresstests"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/tests/integrationtest"

	"github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func tearDownTestForIngress(t *testing.T) {
	integrationtest.DelSVC(t, "default", "avisvc")
	integrationtest.DelEP(t, "default", "avisvc")
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

	ingClassName, ingressName, ns, settingName := "avi-lb", "foo-with-class", "default", "my-infrasetting"
	secretName := "my-secret"
	modelName := "admin/cluster--Shared-L7-EVH-1"
	settingModelName := "nonadmin/cluster--Shared-L7-EVH-0"

	ingresstests.SetUpTestForIngress(t, modelName)

	vsKey := cache.NamespaceName{Namespace: "nonadmin", Name: "cluster--Shared-L7-EVH-0"}
	evhKey := cache.NamespaceName{Namespace: "nonadmin", Name: lib.Encode("cluster--baz.com", lib.EVHVS)}
	integrationtest.SetupAviInfraSetting(t, settingName, "SMALL", "nonadmin")
	integrationtest.AnnotateAKONamespaceWithInfraSetting(t, ns, settingName)
	integrationtest.SetupIngressClass(t, ingClassName, lib.AviIngressController, "")
	waitAndVerify(t, ingClassName)
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

	secureVsName := "cluster--baz.com"
	insecureVsName := "cluster--bar.com"
	insecurePoolName := "cluster--default-bar.com_foo-foo-with-class-avisvc"
	securePoolName := "cluster--default-baz.com_foo-foo-with-class-avisvc"
	insecurePGName := "cluster--default-bar.com_foo-foo-with-class"
	securePGName := "cluster--default-baz.com_foo-foo-with-class"

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
		g.Expect(settingNodes[0].VSVIPRefs[0].T1Lr).Should(gomega.Equal("avi-domain-c9:1234"))
	}

	err = KubeClient.NetworkingV1().Ingresses(ns).Delete(context.TODO(), ingressName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	integrationtest.DeleteSecret(secretName, ns)
	integrationtest.RemoveAnnotateAKONamespaceWithInfraSetting(t, ns)
	integrationtest.TeardownAviInfraSetting(t, settingName)
	tearDownTestForIngress(t)
	integrationtest.TeardownIngressClass(t, ingClassName)
	waitAndVerify(t, ingClassName)
	verifyEvhNodeDeletionFromVsNode(g, settingModelName, vsKey, evhKey)
}

func TestMultiTenancyWithIngressClassAviInfraSettingEVH(t *testing.T) {
	// create secure and insecure host ingress, connect with infrasetting
	// check for names of all Avi objects
	g := gomega.NewGomegaWithT(t)

	ingClassName, ingressName, ns, settingName := "avi-lb", "foo-with-class", "default", "my-infrasetting"
	secretName := "my-secret"
	modelName := "admin/cluster--Shared-L7-EVH-1"

	ingresstests.SetUpTestForIngress(t, modelName)

	settingModelName := "nonadmin/cluster--Shared-L7-EVH-my-infrasetting-0"
	vsKey := cache.NamespaceName{Namespace: "nonadmin", Name: "cluster--Shared-L7-EVH-my-infrasetting-0"}
	evhKey := cache.NamespaceName{Namespace: "nonadmin", Name: lib.Encode("cluster--my-infrasetting-baz.com", lib.EVHVS)}
	integrationtest.SetupAviInfraSetting(t, settingName, "SMALL", "nonadmin")
	integrationtest.SetupIngressClass(t, ingClassName, lib.AviIngressController, settingName)
	waitAndVerify(t, ingClassName)
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
	time.Sleep(5 * time.Second)
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

	err = KubeClient.NetworkingV1().Ingresses(ns).Delete(context.TODO(), ingressName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	integrationtest.DeleteSecret(secretName, ns)
	integrationtest.TeardownAviInfraSetting(t, settingName)
	tearDownTestForIngress(t)
	integrationtest.TeardownIngressClass(t, ingClassName)
	waitAndVerify(t, ingClassName)
	verifyEvhNodeDeletionFromVsNode(g, settingModelName, vsKey, evhKey)
}

func TestMultiTenancyWithInfraSettingAdditionEVH(t *testing.T) {
	// create infrasettings, update infrasetting with a tenant
	// new model creation should happen, old model should get deleted
	g := gomega.NewGomegaWithT(t)

	ingClassName, ingressName, ns, settingName := "avi-lb", "foo-with-class", "default", "my-infrasetting"
	secretName := "my-secret"
	modelName := "admin/cluster--Shared-L7-EVH-1"
	settingModelName := "nonadmin/cluster--Shared-L7-EVH-0"

	ingresstests.SetUpTestForIngress(t, modelName, settingModelName)

	integrationtest.SetupIngressClass(t, ingClassName, lib.AviIngressController, "")
	waitAndVerify(t, ingClassName)
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

	secureVsName := "cluster--baz.com"
	insecureVsName := "cluster--bar.com"
	insecurePoolName := "cluster--default-bar.com_foo-foo-with-class-avisvc"
	securePoolName := "cluster--default-baz.com_foo-foo-with-class-avisvc"
	insecurePGName := "cluster--default-bar.com_foo-foo-with-class"
	securePGName := "cluster--default-baz.com_foo-foo-with-class"

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

	integrationtest.SetupAviInfraSetting(t, settingName, "SMALL", "nonadmin")
	integrationtest.AnnotateAKONamespaceWithInfraSetting(t, ns, settingName)

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
	g.Expect(settingNodes[0].VSVIPRefs[0].T1Lr).Should(gomega.Equal("avi-domain-c9:1234"))

	err = KubeClient.NetworkingV1().Ingresses(ns).Delete(context.TODO(), ingressName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	integrationtest.DeleteSecret(secretName, ns)
	tearDownTestForIngress(t)
	integrationtest.TeardownIngressClass(t, ingClassName)
	waitAndVerify(t, ingClassName)
	verifyPoolDeletionFromVsNode(g, modelName)
	integrationtest.RemoveAnnotateAKONamespaceWithInfraSetting(t, ns)
	integrationtest.TeardownAviInfraSetting(t, settingName)
}

func TestMultiTenancyWithInfraSettingSwitchFromNSToIngressClassEVH(t *testing.T) {
	// Aviinfrasetting gets changed from NS scoped to Ingress class scoped
	// New VS creation should happen with Infrasetting in its name.
	// Old Model should get cleaned up
	g := gomega.NewGomegaWithT(t)

	ingClassName, ingressName, ns, settingName := "avi-lb", "foo-with-class", "default", "my-infrasetting"
	secretName := "my-secret"
	modelName := "nonadmin/cluster--Shared-L7-EVH-1"
	settingModelName := "nonadmin/cluster--Shared-L7-EVH-0"

	ingresstests.SetUpTestForIngress(t, modelName, settingModelName)

	integrationtest.SetupAviInfraSetting(t, settingName, "SMALL", "nonadmin")
	integrationtest.AnnotateAKONamespaceWithInfraSetting(t, ns, settingName)
	integrationtest.SetupIngressClass(t, ingClassName, lib.AviIngressController, "")
	waitAndVerify(t, ingClassName)
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

	secureVsName := "cluster--baz.com"
	insecureVsName := "cluster--bar.com"
	insecurePoolName := "cluster--default-bar.com_foo-foo-with-class-avisvc"
	securePoolName := "cluster--default-baz.com_foo-foo-with-class-avisvc"
	insecurePGName := "cluster--default-bar.com_foo-foo-with-class"
	securePGName := "cluster--default-baz.com_foo-foo-with-class"

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

	newSettingName := "my-infrasetting-1"
	integrationtest.SetupAviInfraSetting(t, newSettingName, "SMALL", "nonadmin")
	integrationtest.SetupIngressClass(t, ingClassName, lib.AviIngressController, newSettingName)
	waitAndVerify(t, ingClassName)

	g.Eventually(func() bool {
		if found, aviSettingModel := objects.SharedAviGraphLister().Get(settingModelName); found {
			if settingNodes := aviSettingModel.(*avinodes.AviObjectGraph).GetAviEvhVS(); len(settingNodes) > 0 {
				return len(settingNodes[0].EvhNodes) == 0 && len(settingNodes[0].PoolRefs) == 0
			}
		}
		return true
	}, 55*time.Second).Should(gomega.Equal(true))

	newSettingModelName := "nonadmin/cluster--Shared-L7-EVH-my-infrasetting-1-0"

	secureVsName = "cluster--my-infrasetting-1-baz.com"
	insecureVsName = "cluster--my-infrasetting-1-bar.com"
	insecurePoolName = "cluster--my-infrasetting-1-default-bar.com_foo-foo-with-class-avisvc"
	securePoolName = "cluster--my-infrasetting-1-default-baz.com_foo-foo-with-class-avisvc"
	insecurePGName = "cluster--my-infrasetting-1-default-bar.com_foo-foo-with-class"
	securePGName = "cluster--my-infrasetting-1-default-baz.com_foo-foo-with-class"

	g.Eventually(func() bool {
		if found, aviSettingModel := objects.SharedAviGraphLister().Get(settingModelName); found {
			if settingNodes := aviSettingModel.(*avinodes.AviObjectGraph).GetAviEvhVS(); len(settingNodes) > 0 {
				return len(settingNodes[0].EvhNodes) == 0
			}
		}
		return true
	}, 55*time.Second).Should(gomega.Equal(true))
	time.Sleep(5 * time.Second)
	_, aviSettingModel = objects.SharedAviGraphLister().Get(newSettingModelName)
	settingNodes = aviSettingModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(settingNodes[0].ServiceEngineGroup).Should(gomega.Equal("thisisaviref-my-infrasetting-1-seGroup"))
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
	tearDownTestForIngress(t)
	integrationtest.TeardownIngressClass(t, ingClassName)
	waitAndVerify(t, ingClassName)
	integrationtest.RemoveAnnotateAKONamespaceWithInfraSetting(t, ns)
	integrationtest.TeardownAviInfraSetting(t, settingName)
	integrationtest.TeardownAviInfraSetting(t, newSettingName)
	verifyPoolDeletionFromVsNode(g, newSettingModelName)
}

func TestMultiTenancyWithNSAviInfraSettingDeletionEVH(t *testing.T) {
	// create an ingress, infrasetting and annotate a namespace with infrasetting
	// graph layer objects should come up with correct tenant
	// delete the Infrasetting, old model should be deleted
	// new model creation will not happen as the Infrasetting annotation is still there
	// and AKO will bail out ingress creation due to failure in fetching the infrasetting
	g := gomega.NewGomegaWithT(t)

	ingClassName, ingressName, ns, settingName := "avi-lb", "foo-with-class", "default", "my-infrasetting"
	secretName := "my-secret"
	modelName := "admin/cluster--Shared-L7-EVH-1"
	settingModelName := "nonadmin/cluster--Shared-L7-EVH-0"

	ingresstests.SetUpTestForIngress(t, modelName)

	vsKey := cache.NamespaceName{Namespace: "nonadmin", Name: "cluster--Shared-L7-EVH-0"}
	evhKey := cache.NamespaceName{Namespace: "nonadmin", Name: lib.Encode("cluster--baz.com", lib.EVHVS)}
	integrationtest.SetupAviInfraSetting(t, settingName, "SMALL", "nonadmin")
	integrationtest.AnnotateAKONamespaceWithInfraSetting(t, ns, settingName)
	integrationtest.SetupIngressClass(t, ingClassName, lib.AviIngressController, "")
	waitAndVerify(t, ingClassName)
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

	secureVsName := "cluster--baz.com"
	insecureVsName := "cluster--bar.com"
	insecurePoolName := "cluster--default-bar.com_foo-foo-with-class-avisvc"
	securePoolName := "cluster--default-baz.com_foo-foo-with-class-avisvc"
	insecurePGName := "cluster--default-bar.com_foo-foo-with-class"
	securePGName := "cluster--default-baz.com_foo-foo-with-class"

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
		g.Expect(settingNodes[0].VSVIPRefs[0].T1Lr).Should(gomega.Equal("avi-domain-c9:1234"))
	}

	integrationtest.TeardownAviInfraSetting(t, settingName)
	verifyEvhNodeDeletionFromVsNode(g, settingModelName, vsKey, evhKey)

	err = KubeClient.NetworkingV1().Ingresses(ns).Delete(context.TODO(), ingressName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	integrationtest.DeleteSecret(secretName, ns)
	integrationtest.RemoveAnnotateAKONamespaceWithInfraSetting(t, ns)
	tearDownTestForIngress(t)
	integrationtest.TeardownIngressClass(t, ingClassName)
	waitAndVerify(t, ingClassName)
}

func TestMultiTenancyWithNSAviInfraSettingDeannotationEVH(t *testing.T) {
	// create an ingress, infrasetting and annotate a namespace with infrasetting
	// graph layer objects should come up with correct tenant
	// delete the Infrasetting annotation from the namespace, old model should be deleted
	// new model in default tenant should get created
	g := gomega.NewGomegaWithT(t)

	ingClassName, ingressName, ns, settingName := "avi-lb", "foo-with-class", "default", "my-infrasetting"
	secretName := "my-secret"
	modelName := "admin/cluster--Shared-L7-EVH-1"
	settingModelName := "nonadmin/cluster--Shared-L7-EVH-0"

	ingresstests.SetUpTestForIngress(t, modelName)

	vsKey := cache.NamespaceName{Namespace: "nonadmin", Name: "cluster--Shared-L7-EVH-0"}
	evhKey := cache.NamespaceName{Namespace: "nonadmin", Name: lib.Encode("cluster--baz.com", lib.EVHVS)}
	integrationtest.SetupAviInfraSetting(t, settingName, "SMALL", "nonadmin")
	integrationtest.AnnotateAKONamespaceWithInfraSetting(t, ns, settingName)
	integrationtest.SetupIngressClass(t, ingClassName, lib.AviIngressController, "")
	waitAndVerify(t, ingClassName)
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

	secureVsName := "cluster--baz.com"
	insecureVsName := "cluster--bar.com"
	insecurePoolName := "cluster--default-bar.com_foo-foo-with-class-avisvc"
	securePoolName := "cluster--default-baz.com_foo-foo-with-class-avisvc"
	insecurePGName := "cluster--default-bar.com_foo-foo-with-class"
	securePGName := "cluster--default-baz.com_foo-foo-with-class"

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
	insecurePoolName = "cluster--default-bar.com_foo-foo-with-class-avisvc"
	securePoolName = "cluster--default-baz.com_foo-foo-with-class-avisvc"
	insecurePGName = "cluster--default-bar.com_foo-foo-with-class"
	securePGName = "cluster--default-baz.com_foo-foo-with-class"

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
	tearDownTestForIngress(t)
	integrationtest.TeardownIngressClass(t, ingClassName)
	waitAndVerify(t, ingClassName)
	integrationtest.TeardownAviInfraSetting(t, settingName)
	verifyEvhNodeDeletionFromVsNode(g, settingModelName, vsKey, evhKey)
	os.Unsetenv("ENABLE_EVH")
}
