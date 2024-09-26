package multitenancytests

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/cache"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/k8s"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	avinodes "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/nodes"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/api"
	crdfake "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/client/v1alpha1/clientset/versioned/fake"
	v1beta1crdfake "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/client/v1beta1/clientset/versioned/fake"
	utils "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/tests/ingresstests"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/tests/integrationtest"

	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sfake "k8s.io/client-go/kubernetes/fake"
)

var (
	KubeClient       *k8sfake.Clientset
	CRDClient        *crdfake.Clientset
	v1beta1CRDClient *v1beta1crdfake.Clientset
	ctrl             *k8s.AviController
	akoApiServer     *api.FakeApiServer
	keyChan          chan string
	objNameMap       integrationtest.ObjectNameMap
)

func waitAndVerify(t *testing.T, key string) {
	select {
	case data := <-keyChan:
		if data != key {
			t.Fatalf("error in match expected: %v, got: %v", key, data)
		}
	case <-time.After(40 * time.Second):
		t.Fatalf("timed out waiting for %v", key)
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
	avinodes.DequeueIngestion(keyStr, false)
	return nil
}

func verifyPoolDeletionFromVsNode(g *gomega.WithT, modelName string) {
	g.Eventually(func() bool {
		if found, aviModel := objects.SharedAviGraphLister().Get(modelName); found {
			if nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS(); len(nodes) > 0 {
				return len(nodes[0].PoolRefs) == 0
			}
		}
		return true
	}, 50*time.Second).Should(gomega.Equal(true))
}

func TestMain(m *testing.M) {
	os.Setenv("INGRESS_API", "extensionv1")
	os.Setenv("VIP_NETWORK_LIST", `[{"networkName":"net123"}]`)
	os.Setenv("CLUSTER_NAME", "cluster")
	os.Setenv("CLOUD_NAME", "CLOUD_VCENTER")
	os.Setenv("SEG_NAME", "Default-Group")
	os.Setenv("NODE_NETWORK_LIST", `[{"networkName":"net123","cidrs":["10.79.168.0/22"]}]`)
	os.Setenv("POD_NAMESPACE", utils.AKO_DEFAULT_NS)
	os.Setenv("SHARD_VS_SIZE", "LARGE")
	os.Setenv("AUTO_L4_FQDN", "default")
	os.Setenv("POD_NAME", "ako-0")

	os.Setenv("SERVICE_TYPE", "ClusterIP")
	os.Setenv("DEFAULT_LB_CONTROLLER", "true")

	akoControlConfig := lib.AKOControlConfig()
	KubeClient = k8sfake.NewSimpleClientset()
	CRDClient = crdfake.NewSimpleClientset()
	v1beta1CRDClient = v1beta1crdfake.NewSimpleClientset()
	akoControlConfig.SetCRDClientset(CRDClient)
	akoControlConfig.Setv1beta1CRDClientset(v1beta1CRDClient)
	akoControlConfig.SetAKOInstanceFlag(true)
	akoControlConfig.SetEventRecorder(lib.AKOEventComponent, KubeClient, true)
	akoControlConfig.SetDefaultLBController(true)
	data := map[string][]byte{
		"username": []byte("admin"),
		"password": []byte("admin"),
	}
	object := metav1.ObjectMeta{Name: "avi-secret", Namespace: utils.GetAKONamespace()}
	secret := &corev1.Secret{Data: data, ObjectMeta: object}
	KubeClient.CoreV1().Secrets(utils.GetAKONamespace()).Create(context.TODO(), secret, metav1.CreateOptions{})

	registeredInformers := []string{
		utils.ServiceInformer,
		utils.EndpointInformer,
		utils.IngressInformer,
		utils.IngressClassInformer,
		utils.SecretInformer,
		utils.NSInformer,
		utils.NodeInformer,
		utils.ConfigMapInformer,
	}
	utils.NewInformers(utils.KubeClientIntf{ClientSet: KubeClient}, registeredInformers)
	informers := k8s.K8sinformers{Cs: KubeClient}
	k8s.NewCRDInformers()

	mcache := cache.SharedAviObjCache()
	cloudObj := &cache.AviCloudPropertyCache{Name: "Default-Cloud", VType: "mock"}
	subdomains := []string{"avi.internal", ".com"}
	cloudObj.NSIpamDNS = subdomains
	mcache.CloudKeyCache.AviCacheAdd("Default-Cloud", cloudObj)

	akoApiServer = integrationtest.InitializeFakeAKOAPIServer()

	integrationtest.NewAviFakeClientInstance(KubeClient, true)
	defer integrationtest.AviFakeClientInstance.Close()

	ctrl = k8s.SharedAviController()
	stopCh := utils.SetupSignalHandler()
	ctrlCh := make(chan struct{})
	quickSyncCh := make(chan struct{})
	waitGroupMap := make(map[string]*sync.WaitGroup)
	wgIngestion := &sync.WaitGroup{}
	waitGroupMap["ingestion"] = wgIngestion
	wgFastRetry := &sync.WaitGroup{}
	waitGroupMap["fastretry"] = wgFastRetry
	wgSlowRetry := &sync.WaitGroup{}
	waitGroupMap["slowretry"] = wgSlowRetry
	wgGraph := &sync.WaitGroup{}
	waitGroupMap["graph"] = wgGraph
	wgStatus := &sync.WaitGroup{}
	waitGroupMap["status"] = wgStatus
	wgLeaderElection := &sync.WaitGroup{}
	waitGroupMap["leaderElection"] = wgLeaderElection

	integrationtest.AddConfigMap(KubeClient)
	keyChan = make(chan string)
	ctrl.HandleConfigMap(informers, ctrlCh, stopCh, quickSyncCh)
	integrationtest.PollForSyncStart(ctrl, 10)
	integrationtest.KubeClient = KubeClient
	integrationtest.AddDefaultIngressClass()
	ctrl.SetSEGroupCloudNameFromNSAnnotations()
	integrationtest.AddDefaultNamespace()
	integrationtest.AddDefaultNamespace("red")
	integrationtest.AddDefaultNamespace("red-ns")

	go ctrl.InitController(informers, registeredInformers, ctrlCh, stopCh, quickSyncCh, waitGroupMap)
	objNameMap.InitMap()
	os.Exit(m.Run())
}

func TestMultiTenancyWithNSAviInfraSettingForIngress(t *testing.T) {
	// create secure and insecure host ingress, connect with infrasetting
	// check for names of all Avi objects
	g := gomega.NewGomegaWithT(t)

	ingClassName := objNameMap.GenerateName("avi-lb")
	ingressName := objNameMap.GenerateName("foo-with-class")
	ns := "default"
	settingName := objNameMap.GenerateName("my-infrasetting")
	secretName := objNameMap.GenerateName("my-secret")
	modelName := "nonadmin/cluster--Shared-L7-1"

	svcName := objNameMap.GenerateName("avisvc")
	time.Sleep(time.Second * 5)
	ingestionQueue := utils.SharedWorkQueue().GetQueueByName(utils.ObjectIngestionLayer)
	ingestionQueue.SyncFunc = syncFromIngestionLayerWrapper
	defer func() { ingestionQueue.SyncFunc = k8s.SyncFromIngestionLayer }()

	ingresstests.SetUpTestForIngress(t, svcName, modelName)

	settingModelName := "nonadmin/cluster--Shared-L7-0"
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

	shardVsName := "cluster--Shared-L7-0"
	sniVsName := "cluster--baz.com"
	shardPoolName := "cluster--bar.com_foo-default-" + ingressName
	sniPoolName := "cluster--default-baz.com_foo-" + ingressName

	g.Eventually(func() bool {
		if found, aviSettingModel := objects.SharedAviGraphLister().Get(settingModelName); found {
			if settingNodes := aviSettingModel.(*avinodes.AviObjectGraph).GetAviVS(); len(settingNodes) > 0 &&
				len(settingNodes[0].SniNodes) > 0 {
				return settingNodes[0].SniNodes[0].Name == sniVsName && len(settingNodes[0].PoolRefs) == 1
			}
		}
		return false
	}, 55*time.Second).Should(gomega.Equal(true))
	_, aviSettingModel := objects.SharedAviGraphLister().Get(settingModelName)
	settingNodes := aviSettingModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(settingNodes[0].Tenant).Should(gomega.Equal("nonadmin"))
	g.Expect(settingNodes[0].PoolRefs[0].Name).Should(gomega.Equal(shardPoolName))
	g.Expect(settingNodes[0].ServiceEngineGroup).Should(gomega.Equal("thisisaviref-" + settingName + "-seGroup"))
	g.Expect(settingNodes[0].PoolGroupRefs[0].Name).Should(gomega.Equal(shardVsName))
	g.Expect(settingNodes[0].HTTPDSrefs[0].Name).Should(gomega.Equal(shardVsName))
	g.Expect(settingNodes[0].HttpPolicyRefs).Should(gomega.HaveLen(1))
	g.Expect(settingNodes[0].HttpPolicyRefs[0].Name).Should(gomega.Equal(shardVsName))
	g.Expect(settingNodes[0].SniNodes[0].PoolRefs[0].Name).Should(gomega.Equal(sniPoolName))
	g.Expect(settingNodes[0].SniNodes[0].PoolGroupRefs[0].Name).Should(gomega.Equal(sniPoolName))
	g.Expect(settingNodes[0].SniNodes[0].SSLKeyCertRefs[0].Name).Should(gomega.Equal(sniVsName))
	g.Expect(settingNodes[0].SniNodes[0].HttpPolicyRefs[0].HppMap[0].Name).Should(gomega.Equal(sniPoolName))
	g.Expect(settingNodes[0].VSVIPRefs[0].T1Lr).Should(gomega.Equal("avi-domain-c9:1234"))

	err = KubeClient.NetworkingV1().Ingresses(ns).Delete(context.TODO(), ingressName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	integrationtest.DeleteSecret(secretName, ns)
	ingresstests.TearDownTestForIngress(t, modelName, settingModelName)
	integrationtest.TeardownIngressClass(t, ingClassName)
	waitAndVerify(t, ingClassName)
	verifyPoolDeletionFromVsNode(g, modelName)
	integrationtest.RemoveAnnotateAKONamespaceWithInfraSetting(t, ns)
	integrationtest.TeardownAviInfraSetting(t, settingName)
}

func TestMultiTenancyWithIngressClassAviInfraSetting(t *testing.T) {
	// create ingress, ingressclas, infrasetting, add infrasetting to ingressclass
	// graph layer objects should come up in the right tenant
	// delete the ingress, graph layer nodes should get deleted
	g := gomega.NewGomegaWithT(t)

	ingClassName := objNameMap.GenerateName("avi-lb")
	ingressName := objNameMap.GenerateName("foo-with-class")
	ns := "default"
	settingName := objNameMap.GenerateName("my-infrasetting")
	secretName := objNameMap.GenerateName("my-secret")
	modelName := "nonadmin/cluster--Shared-L7-1"
	nsSettingName := "ns-my-infrasetting"

	svcName := objNameMap.GenerateName("avisvc")
	time.Sleep(time.Second * 5)
	ingestionQueue := utils.SharedWorkQueue().GetQueueByName(utils.ObjectIngestionLayer)
	ingestionQueue.SyncFunc = syncFromIngestionLayerWrapper
	defer func() { ingestionQueue.SyncFunc = k8s.SyncFromIngestionLayer }()

	ingresstests.SetUpTestForIngress(t, svcName, modelName)

	settingModelName := "nonadmin/cluster--Shared-L7-" + settingName + "-0"
	integrationtest.SetupAviInfraSetting(t, settingName, "SMALL")
	integrationtest.SetupAviInfraSetting(t, nsSettingName, "DEDICATED")

	integrationtest.AnnotateAKONamespaceWithInfraSetting(t, ns, nsSettingName)
	integrationtest.AnnotateNamespaceWithTenant(t, ns, "nonadmin")
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

	shardVsName := "cluster--Shared-L7-" + settingName + "-0"
	sniVsName := "cluster--" + settingName + "-baz.com"
	shardPoolName := "cluster--" + settingName + "-bar.com_foo-default-" + ingressName
	sniPoolName := "cluster--" + settingName + "-default-baz.com_foo-" + ingressName

	g.Eventually(func() bool {
		if found, aviSettingModel := objects.SharedAviGraphLister().Get(settingModelName); found {
			if settingNodes := aviSettingModel.(*avinodes.AviObjectGraph).GetAviVS(); len(settingNodes) > 0 &&
				len(settingNodes[0].SniNodes) > 0 {
				return settingNodes[0].SniNodes[0].Name == sniVsName && len(settingNodes[0].PoolRefs) == 1
			}
		}
		return false
	}, 55*time.Second).Should(gomega.Equal(true))
	_, aviSettingModel := objects.SharedAviGraphLister().Get(settingModelName)
	settingNodes := aviSettingModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(settingNodes[0].PoolRefs[0].Name).Should(gomega.Equal(shardPoolName))
	g.Expect(settingNodes[0].ServiceEngineGroup).Should(gomega.Equal("thisisaviref-" + settingName + "-seGroup"))
	g.Expect(settingNodes[0].PoolGroupRefs[0].Name).Should(gomega.Equal(shardVsName))
	g.Expect(settingNodes[0].HTTPDSrefs[0].Name).Should(gomega.Equal(shardVsName))
	g.Expect(settingNodes[0].HttpPolicyRefs).Should(gomega.HaveLen(1))
	g.Expect(settingNodes[0].HttpPolicyRefs[0].Name).Should(gomega.Equal(shardVsName))
	g.Expect(settingNodes[0].SniNodes[0].PoolRefs[0].Name).Should(gomega.Equal(sniPoolName))
	g.Expect(settingNodes[0].SniNodes[0].PoolGroupRefs[0].Name).Should(gomega.Equal(sniPoolName))
	g.Expect(settingNodes[0].SniNodes[0].SSLKeyCertRefs[0].Name).Should(gomega.Equal(sniVsName))
	g.Expect(settingNodes[0].SniNodes[0].HttpPolicyRefs[0].HppMap[0].Name).Should(gomega.Equal(sniPoolName))

	err = KubeClient.NetworkingV1().Ingresses(ns).Delete(context.TODO(), ingressName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	integrationtest.DeleteSecret(secretName, ns)
	integrationtest.RemoveAnnotateAKONamespaceWithInfraSetting(t, ns)
	integrationtest.TeardownAviInfraSetting(t, nsSettingName)
	integrationtest.TeardownAviInfraSetting(t, settingName)
	ingresstests.TearDownTestForIngress(t, modelName, settingModelName)
	integrationtest.TeardownIngressClass(t, ingClassName)
	waitAndVerify(t, ingClassName)
	verifyPoolDeletionFromVsNode(g, modelName)
}

func TestMultiTenancyWithInfraSettingAdditionForIngress(t *testing.T) {
	// create infrasettings, update infrasetting with a tenant
	// new model creation should happen, old model should get deleted
	g := gomega.NewGomegaWithT(t)

	ingClassName := objNameMap.GenerateName("avi-lb")
	ingressName := objNameMap.GenerateName("foo-with-class")
	ns := "default"
	settingName := objNameMap.GenerateName("my-infrasetting")
	secretName := objNameMap.GenerateName("my-secret")
	modelName := "admin/cluster--Shared-L7-1"

	svcName := objNameMap.GenerateName("avisvc")
	time.Sleep(time.Second * 5)
	ingestionQueue := utils.SharedWorkQueue().GetQueueByName(utils.ObjectIngestionLayer)
	ingestionQueue.SyncFunc = syncFromIngestionLayerWrapper
	defer func() { ingestionQueue.SyncFunc = k8s.SyncFromIngestionLayer }()

	ingresstests.SetUpTestForIngress(t, svcName, modelName)

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

	shardVsName := "cluster--Shared-L7-1"
	sniVsName := "cluster--baz.com"
	shardPoolName := "cluster--bar.com_foo-default-" + ingressName
	sniPoolName := "cluster--default-baz.com_foo-" + ingressName

	g.Eventually(func() bool {
		if found, aviSettingModel := objects.SharedAviGraphLister().Get(modelName); found {
			if settingNodes := aviSettingModel.(*avinodes.AviObjectGraph).GetAviVS(); len(settingNodes) > 0 &&
				len(settingNodes[0].SniNodes) > 0 {
				return settingNodes[0].SniNodes[0].Name == sniVsName && len(settingNodes[0].PoolRefs) == 1
			}
		}
		return false
	}, 55*time.Second).Should(gomega.Equal(true))
	_, aviSettingModel := objects.SharedAviGraphLister().Get(modelName)
	settingNodes := aviSettingModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(settingNodes[0].Tenant).Should(gomega.Equal("admin"))
	g.Expect(settingNodes[0].PoolRefs[0].Name).Should(gomega.Equal(shardPoolName))
	g.Expect(settingNodes[0].PoolGroupRefs[0].Name).Should(gomega.Equal(shardVsName))
	g.Expect(settingNodes[0].HTTPDSrefs[0].Name).Should(gomega.Equal(shardVsName))
	g.Expect(settingNodes[0].HttpPolicyRefs).Should(gomega.HaveLen(1))
	g.Expect(settingNodes[0].HttpPolicyRefs[0].Name).Should(gomega.Equal(shardVsName))
	g.Expect(settingNodes[0].SniNodes[0].PoolRefs[0].Name).Should(gomega.Equal(sniPoolName))
	g.Expect(settingNodes[0].SniNodes[0].PoolGroupRefs[0].Name).Should(gomega.Equal(sniPoolName))
	g.Expect(settingNodes[0].SniNodes[0].SSLKeyCertRefs[0].Name).Should(gomega.Equal(sniVsName))
	g.Expect(settingNodes[0].SniNodes[0].HttpPolicyRefs[0].HppMap[0].Name).Should(gomega.Equal(sniPoolName))

	integrationtest.SetupAviInfraSetting(t, settingName, "SMALL")
	integrationtest.AnnotateAKONamespaceWithInfraSetting(t, ns, settingName)
	integrationtest.AnnotateNamespaceWithTenant(t, ns, "nonadmin")

	g.Eventually(func() bool {
		if found, aviSettingModel := objects.SharedAviGraphLister().Get(modelName); found {
			if settingNodes := aviSettingModel.(*avinodes.AviObjectGraph).GetAviVS(); len(settingNodes) > 0 {
				return len(settingNodes[0].SniNodes) == 0 && len(settingNodes[0].PoolRefs) == 0
			}
		}
		return false
	}, 55*time.Second).Should(gomega.Equal(true))

	settingModelName := "nonadmin/cluster--Shared-L7-0"
	shardVsName = "cluster--Shared-L7-0"
	g.Eventually(func() bool {
		if found, aviSettingModel := objects.SharedAviGraphLister().Get(settingModelName); found {
			if settingNodes := aviSettingModel.(*avinodes.AviObjectGraph).GetAviVS(); len(settingNodes) > 0 &&
				len(settingNodes[0].SniNodes) > 0 {
				return settingNodes[0].SniNodes[0].Name == sniVsName && len(settingNodes[0].PoolRefs) == 1
			}
		}
		return false
	}, 55*time.Second).Should(gomega.Equal(true))
	_, aviSettingModel = objects.SharedAviGraphLister().Get(settingModelName)
	settingNodes = aviSettingModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(settingNodes[0].Tenant).Should(gomega.Equal("nonadmin"))
	g.Expect(settingNodes[0].PoolRefs[0].Name).Should(gomega.Equal(shardPoolName))
	g.Expect(settingNodes[0].ServiceEngineGroup).Should(gomega.Equal("thisisaviref-" + settingName + "-seGroup"))
	g.Expect(settingNodes[0].PoolGroupRefs[0].Name).Should(gomega.Equal(shardVsName))
	g.Expect(settingNodes[0].HTTPDSrefs[0].Name).Should(gomega.Equal(shardVsName))
	g.Expect(settingNodes[0].HttpPolicyRefs).Should(gomega.HaveLen(1))
	g.Expect(settingNodes[0].HttpPolicyRefs[0].Name).Should(gomega.Equal(shardVsName))
	g.Expect(settingNodes[0].SniNodes[0].PoolRefs[0].Name).Should(gomega.Equal(sniPoolName))
	g.Expect(settingNodes[0].SniNodes[0].PoolGroupRefs[0].Name).Should(gomega.Equal(sniPoolName))
	g.Expect(settingNodes[0].SniNodes[0].SSLKeyCertRefs[0].Name).Should(gomega.Equal(sniVsName))
	g.Expect(settingNodes[0].SniNodes[0].HttpPolicyRefs[0].HppMap[0].Name).Should(gomega.Equal(sniPoolName))
	g.Expect(settingNodes[0].VSVIPRefs[0].T1Lr).Should(gomega.Equal("avi-domain-c9:1234"))

	err = KubeClient.NetworkingV1().Ingresses(ns).Delete(context.TODO(), ingressName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	integrationtest.DeleteSecret(secretName, ns)
	ingresstests.TearDownTestForIngress(t, modelName, settingModelName)
	integrationtest.TeardownIngressClass(t, ingClassName)
	waitAndVerify(t, ingClassName)
	verifyPoolDeletionFromVsNode(g, modelName)
	integrationtest.RemoveAnnotateAKONamespaceWithInfraSetting(t, ns)
	integrationtest.TeardownAviInfraSetting(t, settingName)
}

func TestMultiTenancyWithTenantDeannotationInNSForIngress(t *testing.T) {
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
	modelName := "nonadmin/cluster--Shared-L7-1"
	svcName := objNameMap.GenerateName("avisvc")
	time.Sleep(time.Second * 5)
	ingestionQueue := utils.SharedWorkQueue().GetQueueByName(utils.ObjectIngestionLayer)
	ingestionQueue.SyncFunc = syncFromIngestionLayerWrapper
	defer func() { ingestionQueue.SyncFunc = k8s.SyncFromIngestionLayer }()

	ingresstests.SetUpTestForIngress(t, svcName, modelName)

	settingModelName := "nonadmin/cluster--Shared-L7-0"
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

	shardVsName := "cluster--Shared-L7-0"
	sniVsName := "cluster--baz.com"
	shardPoolName := "cluster--bar.com_foo-default-" + ingressName
	sniPoolName := "cluster--default-baz.com_foo-" + ingressName

	g.Eventually(func() bool {
		if found, aviSettingModel := objects.SharedAviGraphLister().Get(settingModelName); found {
			if settingNodes := aviSettingModel.(*avinodes.AviObjectGraph).GetAviVS(); len(settingNodes) > 0 &&
				len(settingNodes[0].SniNodes) > 0 {
				return settingNodes[0].SniNodes[0].Name == sniVsName && len(settingNodes[0].PoolRefs) == 1
			}
		}
		return false
	}, 55*time.Second).Should(gomega.Equal(true))
	_, aviSettingModel := objects.SharedAviGraphLister().Get(settingModelName)
	settingNodes := aviSettingModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(settingNodes[0].PoolRefs[0].Name).Should(gomega.Equal(shardPoolName))
	g.Expect(settingNodes[0].ServiceEngineGroup).Should(gomega.Equal("thisisaviref-" + settingName + "-seGroup"))
	g.Expect(settingNodes[0].PoolGroupRefs[0].Name).Should(gomega.Equal(shardVsName))
	g.Expect(settingNodes[0].HTTPDSrefs[0].Name).Should(gomega.Equal(shardVsName))
	g.Expect(settingNodes[0].HttpPolicyRefs).Should(gomega.HaveLen(1))
	g.Expect(settingNodes[0].HttpPolicyRefs[0].Name).Should(gomega.Equal(shardVsName))
	g.Expect(settingNodes[0].SniNodes[0].PoolRefs[0].Name).Should(gomega.Equal(sniPoolName))
	g.Expect(settingNodes[0].SniNodes[0].PoolGroupRefs[0].Name).Should(gomega.Equal(sniPoolName))
	g.Expect(settingNodes[0].SniNodes[0].SSLKeyCertRefs[0].Name).Should(gomega.Equal(sniVsName))
	g.Expect(settingNodes[0].SniNodes[0].HttpPolicyRefs[0].HppMap[0].Name).Should(gomega.Equal(sniPoolName))
	g.Expect(settingNodes[0].VSVIPRefs[0].T1Lr).Should(gomega.Equal("avi-domain-c9:1234"))

	integrationtest.RemoveAnnotateAKONamespaceWithInfraSetting(t, ns)
	verifyPoolDeletionFromVsNode(g, settingModelName)

	modelName = "admin/cluster--Shared-L7-1"
	shardVsName = "cluster--Shared-L7-1"

	g.Eventually(func() bool {
		if found, aviSettingModel := objects.SharedAviGraphLister().Get(modelName); found {
			if settingNodes := aviSettingModel.(*avinodes.AviObjectGraph).GetAviVS(); len(settingNodes) > 0 &&
				len(settingNodes[0].SniNodes) > 0 {
				return settingNodes[0].SniNodes[0].Name == sniVsName && len(settingNodes[0].PoolRefs) == 1
			}
		}
		return false
	}, 55*time.Second).Should(gomega.Equal(true))
	_, aviSettingModel = objects.SharedAviGraphLister().Get(modelName)
	settingNodes = aviSettingModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(settingNodes[0].Tenant).Should(gomega.Equal("admin"))
	g.Expect(settingNodes[0].PoolRefs[0].Name).Should(gomega.Equal(shardPoolName))
	g.Expect(settingNodes[0].PoolGroupRefs[0].Name).Should(gomega.Equal(shardVsName))
	g.Expect(settingNodes[0].HTTPDSrefs[0].Name).Should(gomega.Equal(shardVsName))
	g.Expect(settingNodes[0].HttpPolicyRefs).Should(gomega.HaveLen(1))
	g.Expect(settingNodes[0].HttpPolicyRefs[0].Name).Should(gomega.Equal(shardVsName))
	g.Expect(settingNodes[0].SniNodes[0].PoolRefs[0].Name).Should(gomega.Equal(sniPoolName))
	g.Expect(settingNodes[0].SniNodes[0].PoolGroupRefs[0].Name).Should(gomega.Equal(sniPoolName))
	g.Expect(settingNodes[0].SniNodes[0].SSLKeyCertRefs[0].Name).Should(gomega.Equal(sniVsName))
	g.Expect(settingNodes[0].SniNodes[0].HttpPolicyRefs[0].HppMap[0].Name).Should(gomega.Equal(sniPoolName))

	err = KubeClient.NetworkingV1().Ingresses(ns).Delete(context.TODO(), ingressName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	integrationtest.DeleteSecret(secretName, ns)
	ingresstests.TearDownTestForIngress(t, modelName, settingModelName)
	integrationtest.TeardownIngressClass(t, ingClassName)
	waitAndVerify(t, ingClassName)
	integrationtest.TeardownAviInfraSetting(t, settingName)
}
