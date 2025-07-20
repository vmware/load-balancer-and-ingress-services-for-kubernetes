package informers

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/cache"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/k8s"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	avinodes "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/nodes"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"
	crdfake "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/client/v1alpha1/clientset/versioned/fake"
	v1beta1crdfake "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/client/v1beta1/clientset/versioned/fake"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/tests/advl4tests"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/tests/integrationtest"
	advl4fake "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/third_party/service-apis/client/clientset/versioned/fake"
)

var stopCh, ctrlCh chan struct{}
var ctrl *k8s.AviController

func init() {
	stopCh = make(chan struct{})
	ctrlCh = make(chan struct{})
}

func TestMain(m *testing.M) {
	os.Setenv("CLUSTER_ID", "abc:cluster")
	os.Setenv("CLOUD_NAME", "Default-Cloud")
	os.Setenv("ADVANCED_L4", "true")
	os.Setenv("POD_NAMESPACE", utils.AKO_DEFAULT_NS)
	os.Setenv("SHARD_VS_SIZE", "LARGE")
	os.Setenv("POD_NAME", "ako-0")

	akoControlConfig := lib.AKOControlConfig()
	advl4tests.KubeClient = k8sfake.NewSimpleClientset()
	advl4tests.AdvL4Client = advl4fake.NewSimpleClientset()
	advl4tests.CRDClient = crdfake.NewSimpleClientset()
	advl4tests.V1beta1CRDClient = v1beta1crdfake.NewSimpleClientset()
	akoControlConfig.SetAKOInstanceFlag(true)
	akoControlConfig.SetAdvL4Clientset(advl4tests.AdvL4Client)
	akoControlConfig.Setv1beta1CRDClientset(advl4tests.V1beta1CRDClient)
	akoControlConfig.SetCRDClientsetAndEnableInfraSettingParam(advl4tests.V1beta1CRDClient)
	akoControlConfig.SetEventRecorder(lib.AKOEventComponent, advl4tests.KubeClient, true)
	akoControlConfig.SetDefaultLBController(true)
	data := map[string][]byte{
		"username": []byte("admin"),
		"password": []byte("admin"),
	}
	object := metav1.ObjectMeta{Name: "avi-secret", Namespace: utils.GetAKONamespace()}
	secret := &corev1.Secret{Data: data, ObjectMeta: object}
	advl4tests.KubeClient.CoreV1().Secrets(utils.GetAKONamespace()).Create(context.TODO(), secret, metav1.CreateOptions{})

	registeredInformers := []string{
		utils.ServiceInformer,
		utils.EndpointInformer,
		utils.SecretInformer,
		utils.NSInformer,
		utils.ConfigMapInformer,
	}
	utils.NewInformers(utils.KubeClientIntf{ClientSet: advl4tests.KubeClient}, registeredInformers)
	informers := k8s.K8sinformers{Cs: advl4tests.KubeClient}
	k8s.NewCRDInformers()
	k8s.NewAdvL4Informers(advl4tests.AdvL4Client)

	mcache := cache.SharedAviObjCache()
	cloudObj := &cache.AviCloudPropertyCache{Name: "Default-Cloud", VType: "mock"}
	subdomains := []string{"avi.internal", ".com"}
	cloudObj.NSIpamDNS = subdomains
	mcache.CloudKeyCache.AviCacheAdd("Default-Cloud", cloudObj)

	integrationtest.KubeClient = advl4tests.KubeClient
	integrationtest.DefaultMockFilePath = "../../avimockobjects"
	integrationtest.InitializeFakeAKOAPIServer()

	integrationtest.NewAviFakeClientInstance(advl4tests.KubeClient)
	defer integrationtest.AviFakeClientInstance.Close()

	ctrl = k8s.SharedAviController()
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

	integrationtest.AddConfigMap(advl4tests.KubeClient)
	ctrl.SetSEGroupCloudNameFromNSAnnotations()

	integrationtest.PollForSyncStart(ctrl, 10)

	ctrl.HandleConfigMap(informers, ctrlCh, stopCh, quickSyncCh)
	integrationtest.AddDefaultNamespace()
	go ctrl.InitController(informers, registeredInformers, ctrlCh, stopCh, quickSyncCh, waitGroupMap)
	os.Exit(m.Run())
}

func RestartController(stopCh, ctrlCh chan struct{}) {
	akoControlConfig := lib.AKOControlConfig()
	advl4tests.KubeClient = k8sfake.NewSimpleClientset()
	advl4tests.AdvL4Client = advl4fake.NewSimpleClientset()

	advl4tests.AdvL4Client.PrependWatchReactor("*", func(action k8stesting.Action) (handled bool, ret watch.Interface, err error) {
		return true, nil, fmt.Errorf("simulated watch error")
	})
	advl4tests.AdvL4Client.PrependReactor("list", "gateways", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, nil, fmt.Errorf("simulated list error")
	})
	advl4tests.AdvL4Client.PrependReactor("list", "gatewayclasses", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, nil, fmt.Errorf("simulated list error")
	})
	advl4tests.CRDClient = crdfake.NewSimpleClientset()
	advl4tests.V1beta1CRDClient = v1beta1crdfake.NewSimpleClientset()
	akoControlConfig.SetAKOInstanceFlag(true)
	akoControlConfig.SetAdvL4Clientset(advl4tests.AdvL4Client)
	akoControlConfig.Setv1beta1CRDClientset(advl4tests.V1beta1CRDClient)
	akoControlConfig.SetCRDClientsetAndEnableInfraSettingParam(advl4tests.V1beta1CRDClient)
	akoControlConfig.SetEventRecorder(lib.AKOEventComponent, advl4tests.KubeClient, true)
	akoControlConfig.SetDefaultLBController(true)

	registeredInformers := []string{
		utils.ServiceInformer,
		utils.EndpointInformer,
		utils.SecretInformer,
		utils.NSInformer,
		utils.ConfigMapInformer,
	}
	utils.NewInformers(utils.KubeClientIntf{ClientSet: advl4tests.KubeClient}, registeredInformers)
	informers := k8s.K8sinformers{Cs: advl4tests.KubeClient}
	k8s.NewCRDInformers()
	k8s.NewAdvL4Informers(advl4tests.AdvL4Client)

	integrationtest.KubeClient = advl4tests.KubeClient

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
	ctrl.State = &k8s.State{}
	go ctrl.InitController(informers, registeredInformers, ctrlCh, stopCh, quickSyncCh, waitGroupMap)
}

func TestAdvL4InformerError(t *testing.T) {

	g := gomega.NewGomegaWithT(t)

	gwClassName, gatewayName, ns := "avi-lb-1", "my-gateway-1", "default"
	modelName := "admin/abc--default-" + gatewayName
	svcName := "svc-1"

	advl4tests.SetupGatewayClass(t, gwClassName, lib.AviGatewayController)
	advl4tests.SetupGateway(t, gatewayName, ns, gwClassName, false)

	advl4tests.SetupAdvLBService(t, svcName, ns, gatewayName, ns)

	g.Eventually(func() string {
		gw, _ := lib.AKOControlConfig().AdvL4Informers().GatewayInformer.Lister().Gateways(ns).Get(gatewayName)
		if len(gw.Status.Addresses) > 0 {
			return gw.Status.Addresses[0].Value
		}
		return ""
	}, 40*time.Second).Should(gomega.Equal("10.250.250.1"))

	g.Eventually(func() string {
		svc, _ := advl4tests.KubeClient.CoreV1().Services(ns).Get(context.TODO(), svcName, metav1.GetOptions{})
		if len(svc.Status.LoadBalancer.Ingress) > 0 {
			return svc.Status.LoadBalancer.Ingress[0].IP
		}
		return ""
	}, 30*time.Second).Should(gomega.Equal("10.250.250.1"))

	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes[0].PortProto[0].Port).To(gomega.Equal(int32(8081)))
	g.Expect(nodes[0].HttpPolicySetRefs).To(gomega.HaveLen(0))
	g.Expect(nodes[0].L4PolicyRefs).To(gomega.HaveLen(1))
	g.Expect(nodes[0].L4PolicyRefs[0].PortPool[0].Port).To(gomega.Equal(uint32(8081)))
	g.Expect(nodes[0].L4PolicyRefs[0].PortPool[0].Protocol).To(gomega.Equal("TCP"))
	g.Expect(nodes[0].ServiceMetadata.NamespaceServiceName[0]).To(gomega.Equal("default/" + svcName))
	g.Expect(nodes[0].ServiceMetadata.Gateway).To(gomega.Equal("default/" + gatewayName))
	g.Expect(nodes[0].PoolRefs[0].Servers).To(gomega.HaveLen(3))

	go func() {
		close(stopCh)
		ctrlCh <- struct{}{}
	}()
	time.Sleep(10 * time.Second)
	stopCh := make(chan struct{})
	go RestartController(stopCh, ctrlCh)
	time.Sleep(10 * time.Second)

	g.Eventually(func() bool {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		return found && aviModel != nil
	}, 30*time.Second).Should(gomega.BeTrue())
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)

	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes[0].PortProto[0].Port).To(gomega.Equal(int32(8081)))
	g.Expect(nodes[0].HttpPolicySetRefs).To(gomega.HaveLen(0))
	g.Expect(nodes[0].L4PolicyRefs).To(gomega.HaveLen(1))
	g.Expect(nodes[0].L4PolicyRefs[0].PortPool[0].Port).To(gomega.Equal(uint32(8081)))
	g.Expect(nodes[0].L4PolicyRefs[0].PortPool[0].Protocol).To(gomega.Equal("TCP"))
	g.Expect(nodes[0].ServiceMetadata.NamespaceServiceName[0]).To(gomega.Equal("default/" + svcName))
	g.Expect(nodes[0].ServiceMetadata.Gateway).To(gomega.Equal("default/" + gatewayName))
	g.Expect(nodes[0].PoolRefs[0].Servers).To(gomega.HaveLen(3))

}
