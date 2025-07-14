/*
 * Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
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

package namespacesynctests

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/tests/integrationtest"

	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sfake "k8s.io/client-go/kubernetes/fake"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/cache"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/k8s"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	avinodes "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/nodes"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"
	crdfake "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/client/v1alpha1/clientset/versioned/fake"
	v1beta1crdfake "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/client/v1beta1/clientset/versioned/fake"
)

var KubeClient *k8sfake.Clientset
var ctrl *k8s.AviController
var CRDClient *crdfake.Clientset
var V1beta1CRDClient *v1beta1crdfake.Clientset

func TestMain(m *testing.M) {
	os.Setenv("INGRESS_API", "extensionv1")
	os.Setenv("VIP_NETWORK_LIST", `[{"networkName":"net123"}]`)
	os.Setenv("CLUSTER_NAME", "cluster")
	os.Setenv("CLOUD_NAME", "CLOUD_VCENTER")
	os.Setenv("SEG_NAME", "Default-Group")
	os.Setenv("NODE_NETWORK_LIST", `[{"networkName":"net123","cidrs":["10.79.168.0/22"]}]`)
	os.Setenv("SERVICE_TYPE", "ClusterIP")
	os.Setenv("POD_NAMESPACE", utils.AKO_DEFAULT_NS)
	os.Setenv("SHARD_VS_SIZE", "LARGE")
	os.Setenv("POD_NAME", "ako-0")

	akoControlConfig := lib.AKOControlConfig()
	KubeClient = k8sfake.NewSimpleClientset()
	CRDClient = crdfake.NewSimpleClientset()
	V1beta1CRDClient = v1beta1crdfake.NewSimpleClientset()
	akoControlConfig.SetCRDClientset(CRDClient)
	akoControlConfig.Setv1beta1CRDClientset(V1beta1CRDClient)
	akoControlConfig.SetEventRecorder(lib.AKOEventComponent, KubeClient, true)
	akoControlConfig.SetAKOInstanceFlag(true)
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

	integrationtest.InitializeFakeAKOAPIServer()

	integrationtest.NewAviFakeClientInstance(KubeClient)
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
	ctrl.SetSEGroupCloudNameFromNSAnnotations()

	integrationtest.PollForSyncStart(ctrl, 10)

	ctrl.HandleConfigMap(informers, ctrlCh, stopCh, quickSyncCh)
	integrationtest.KubeClient = KubeClient
	integrationtest.AddDefaultIngressClass()

	SetupNamespaceSync("app", "migrate")
	go ctrl.InitController(informers, registeredInformers, ctrlCh, stopCh, quickSyncCh, waitGroupMap)
	os.Exit(m.Run())
}

func SetupNamespaceSync(key, value string) {
	os.Setenv("NAMESPACE_SYNC_LABEL_KEY", key)
	os.Setenv("NAMESPACE_SYNC_LABEL_VALUE", value)
	ctrl.InitializeNamespaceSync()
}

func UpdateIngress(t *testing.T, modelName, namespace string) {
	integrationtest.CreateSVC(t, namespace, "avisvc1", corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEP(t, namespace, "avisvc1", false, false, "2.2.2")
	integrationtest.PollForCompletion(t, modelName, 5)
	ingressObject := (integrationtest.FakeIngress{
		Name:        "foo-with-targets",
		Namespace:   namespace,
		DnsNames:    []string{"bar.com"},
		Paths:       []string{"/foo"},
		ServiceName: "avisvc1",
	})
	ingrFake := ingressObject.Ingress()
	ingrFake.ResourceVersion = "22"
	if _, err := KubeClient.NetworkingV1().Ingresses(namespace).Update(context.TODO(), ingrFake, metav1.UpdateOptions{}); err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	integrationtest.PollForCompletion(t, modelName, 5)
}

func SetupIngress(t *testing.T, modelName, namespace string, withSecret, tlsIngress bool) {

	objects.SharedAviGraphLister().Delete(modelName)
	integrationtest.CreateSVC(t, namespace, "avisvc", corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEP(t, namespace, "avisvc", false, false, "1.1.1")
	integrationtest.PollForCompletion(t, modelName, 5)

	ingressObject := (integrationtest.FakeIngress{
		Name:        "foo-with-targets",
		Namespace:   namespace,
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/foo"},
		ServiceName: "avisvc",
	})
	if withSecret {
		integrationtest.AddSecret("my-secret", namespace, "tlsCert", "tlsKey")
	}
	if tlsIngress {
		ingressObject.TlsSecretDNS = map[string][]string{
			"my-secret": {"foo.com"},
		}
	}

	ingrFake := ingressObject.Ingress()
	if _, err := KubeClient.NetworkingV1().Ingresses(namespace).Create(context.TODO(), ingrFake, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	integrationtest.PollForCompletion(t, modelName, 5)
}

func TearDownTestForIngressNamespace(t *testing.T, modelName, namespace string, g *gomega.GomegaWithT) {

	if err := KubeClient.NetworkingV1().Ingresses(namespace).Delete(context.TODO(), "foo-with-targets", metav1.DeleteOptions{}); err != nil {
		t.Fatalf("Couldn't Delete Ingress: %v", err)
	}
	g.Eventually(func() error {
		_, err := KubeClient.NetworkingV1().Ingresses(namespace).Get(context.TODO(), "foo-with-targets", metav1.GetOptions{})
		return err
	}, 30*time.Second).Should(gomega.Not(gomega.BeNil()))

	objects.SharedAviGraphLister().Delete(modelName)
	integrationtest.DelSVC(t, namespace, "avisvc")
	integrationtest.DelEP(t, namespace, "avisvc")
	integrationtest.DeleteNamespace(namespace)
	integrationtest.PollForCompletion(t, modelName, 10)
}

func VerifyModelDeleted(g *gomega.WithT, modelName string) {
	g.Eventually(func() interface{} {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if aviModel != nil {
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
			return nodes[0].PoolRefs
		}
		return aviModel
	}, 30*time.Second).Should(gomega.BeNil())
}

func TestNamespaceSyncFeatureWithCorrectEnvParameters(t *testing.T) {

	var nsLabel map[string]string
	nsLabel = map[string]string{
		"app": "migrate",
	}

	g := gomega.NewGomegaWithT(t)
	var found bool
	//Valid Namespace
	namespace1 := "rednsmig"
	err := integrationtest.AddNamespace(t, namespace1, nsLabel)
	modelName1 := "admin/cluster--Shared-L7-0"
	if err != nil {
		t.Fatal("Error while adding namespace")
	}

	SetupIngress(t, modelName1, namespace1, false, false)
	g.Eventually(func() bool {
		found, _ = objects.SharedAviGraphLister().Get(modelName1)
		return found
	}, 30*time.Second).Should(gomega.Equal(true))

	mcache := cache.SharedAviObjCache()

	poolName := fmt.Sprintf("cluster--foo.com_foo-%s-foo-with-targets", namespace1)
	poolKey := cache.NamespaceName{Namespace: "admin", Name: poolName}

	g.Eventually(func() bool {
		_, found := mcache.PoolCache.AviCacheGet(poolKey)
		return found
	}, 30*time.Second).Should(gomega.Equal(true))

	utils.AviLog.Debug("Update Valid Ingress")
	UpdateIngress(t, modelName1, namespace1)
	poolName = fmt.Sprintf("cluster--bar.com_foo-%s-foo-with-targets", namespace1)
	poolKey = cache.NamespaceName{Namespace: "admin", Name: poolName}

	g.Eventually(func() bool {
		_, found := mcache.PoolCache.AviCacheGet(poolKey)
		return found
	}, 30*time.Second).Should(gomega.Equal(true))

	TearDownTestForIngressNamespace(t, modelName1, namespace1, g)
	VerifyModelDeleted(g, modelName1)

	//Invalid Namespace
	utils.AviLog.Debug("Adding namespace with wrong label")

	namespace := "greenns"
	nsLabel = map[string]string{
		"app": "migrate1",
	}

	err = integrationtest.AddNamespace(t, namespace, nsLabel)
	modelName := "admin/cluster--Shared-L7-0"
	if err != nil {
		t.Fatal("Error while adding namespace")
	}

	SetupIngress(t, modelName, namespace, false, false)
	g.Eventually(func() bool {
		found, _ = objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 30*time.Second).Should(gomega.Equal(false))

	TearDownTestForIngressNamespace(t, modelName, namespace, g)
	VerifyModelDeleted(g, modelName)
}

func checkNSTransition(t *testing.T, oldLabels, newLabels map[string]string, oldFlag, newFlag bool, namespace, modelName string) {

	g := gomega.NewGomegaWithT(t)
	var found bool

	err := integrationtest.AddNamespace(t, namespace, oldLabels)
	if err != nil {
		t.Fatal("Error while adding namespace")
	}

	SetupIngress(t, modelName, namespace, false, false)

	poolName := fmt.Sprintf("cluster--foo.com_foo-%s-foo-with-targets", namespace)

	mcache := cache.SharedAviObjCache()
	poolKey := cache.NamespaceName{Namespace: "admin", Name: poolName}

	if !oldFlag {
		g.Eventually(func() bool {
			found, _ = objects.SharedAviGraphLister().Get(modelName)

			return found
		}, 30*time.Second).Should(gomega.Equal(oldFlag))
	} else {
		g.Eventually(func() bool {
			_, found := mcache.PoolCache.AviCacheGet(poolKey)
			return found
		}, 30*time.Second).Should(gomega.Equal(oldFlag))
	}

	err = integrationtest.UpdateNamespace(t, namespace, newLabels)
	integrationtest.PollForCompletion(t, modelName, 5)
	if err != nil {
		t.Fatal("Error occurred while updating namespace")
	}

	g.Eventually(func() bool {
		_, found := mcache.PoolCache.AviCacheGet(poolKey)
		return found
	}, 30*time.Second).Should(gomega.Equal(newFlag))

	TearDownTestForIngressNamespace(t, modelName, namespace, g)
	VerifyModelDeleted(g, modelName)
}

func TestNSTransitionValidToInvalid(t *testing.T) {
	oldLabels := map[string]string{
		"app": "migrate",
	}
	newLabels := map[string]string{
		"app": "migrate2",
	}
	namespace := "bluemigns"
	modelName := "admin/cluster--Shared-L7-0"
	checkNSTransition(t, oldLabels, newLabels, true, false, namespace, modelName)
}

func TestNSTransitionInvalidToValid(t *testing.T) {
	oldLabels := map[string]string{
		"app": "migrate2",
	}
	newLabels := map[string]string{
		"app": "migrate",
	}
	namespace := "purple"
	modelName := "admin/cluster--Shared-L7-0"

	checkNSTransition(t, oldLabels, newLabels, false, true, namespace, modelName)
}

func TestNSTransitionInvalidToInvalid(t *testing.T) {
	oldLabels := map[string]string{
		"app": "migrate2",
	}
	newLabels := map[string]string{
		"app": "migrate1",
	}
	namespace := "magenta"
	modelName := "admin/cluster--Shared-L7-0"

	checkNSTransition(t, oldLabels, newLabels, false, false, namespace, modelName)
}

// Hostname ShardScheme test case
func TestNSTransitionValidToInvalidHostName(t *testing.T) {
	oldLabels := map[string]string{
		"app": "migrate",
	}
	newLabels := map[string]string{
		"app": "migrate2",
	}
	namespace := "whitemigns"
	modelName := "admin/cluster--Shared-L7-0"

	checkNSTransition(t, oldLabels, newLabels, true, false, namespace, modelName)
}
