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

package multiclusteringresstests

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/onsi/gomega"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/cache"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/k8s"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	avinodes "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/nodes"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"
	crdfake "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/client/v1alpha1/clientset/versioned/fake"
	v1beta1crdfake "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/client/v1beta1/clientset/versioned/fake"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/tests/integrationtest"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/api"
	utils "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sfake "k8s.io/client-go/kubernetes/fake"
)

var KubeClient *k8sfake.Clientset
var CRDClient *crdfake.Clientset
var V1beta1CRDClient *v1beta1crdfake.Clientset
var ctrl *k8s.AviController
var akoApiServer *api.FakeApiServer

func TestMain(m *testing.M) {
	os.Setenv("INGRESS_API", "extensionv1")
	os.Setenv("VIP_NETWORK_LIST", `[{"networkName":"net123"}]`)
	os.Setenv("CLUSTER_NAME", "cluster")
	os.Setenv("CLOUD_NAME", "CLOUD_VCENTER")
	os.Setenv("SEG_NAME", "Default-Group")
	os.Setenv("NODE_NETWORK_LIST", `[{"networkName":"net123","cidrs":["10.79.168.0/22"]}]`)
	os.Setenv("POD_NAMESPACE", utils.AKO_DEFAULT_NS)
	os.Setenv("AUTO_L4_FQDN", "default")
	os.Setenv("SERVICE_TYPE", "NodePort")
	os.Setenv("ENABLE_EVH", "true")
	os.Setenv("MCI_ENABLED", "true")
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
		utils.EndpointSlicesInformer,
		utils.IngressInformer,
		utils.IngressClassInformer,
		utils.SecretInformer,
		utils.NSInformer,
		utils.NodeInformer,
		utils.ConfigMapInformer,
		utils.MultiClusterIngressInformer,
		utils.ServiceImportInformer,
	}
	args := make(map[string]interface{})
	args[utils.INFORMERS_AKO_CLIENT] = CRDClient
	utils.NewInformers(utils.KubeClientIntf{ClientSet: KubeClient}, registeredInformers, args)
	informers := k8s.K8sinformers{Cs: KubeClient}
	k8s.NewCRDInformers()

	mcache := cache.SharedAviObjCache()
	cloudObj := &cache.AviCloudPropertyCache{Name: "Default-Cloud", VType: "mock"}
	subdomains := []string{"avi.internal", ".com"}
	cloudObj.NSIpamDNS = subdomains
	mcache.CloudKeyCache.AviCacheAdd("Default-Cloud", cloudObj)

	akoApiServer = integrationtest.InitializeFakeAKOAPIServer()

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
	integrationtest.PollForSyncStart(ctrl, 10)

	ctrl.HandleConfigMap(informers, ctrlCh, stopCh, quickSyncCh)
	integrationtest.KubeClient = KubeClient
	integrationtest.AddDefaultIngressClass()
	ctrl.SetSEGroupCloudNameFromNSAnnotations()

	go ctrl.InitController(informers, registeredInformers, ctrlCh, stopCh, quickSyncCh, waitGroupMap)
	os.Exit(m.Run())
}

func SetupDomain() {
	mcache := cache.SharedAviObjCache()
	cloudObj := &cache.AviCloudPropertyCache{Name: "Default-Cloud", VType: "mock"}
	subdomains := []string{"avi.internal", ".com"}
	cloudObj.NSIpamDNS = subdomains
	mcache.CloudKeyCache.AviCacheAdd("Default-Cloud", cloudObj)
}

func cleanupModels(modelNames ...string) {
	for _, model := range modelNames {
		objects.SharedAviGraphLister().Delete(model)
	}
}

func getServiceName(str string) string {
	return fmt.Sprintf("svc-%s", str)
}

func getClusterName(str string) string {
	return fmt.Sprintf("cluster-%s", str)
}

func getMultiClusterIngressName(str string) string {
	return fmt.Sprintf("mci-%s", str)
}

func getServiceImportName(str string) string {
	return fmt.Sprintf("si-%s", getServiceName(str))
}

func SetUpServices(t *testing.T, paths []string) {
	for _, path := range paths {
		serviceName := getServiceName(path)
		integrationtest.CreateSVC(t, "default", serviceName, corev1.ProtocolTCP, corev1.ServiceTypeNodePort, false)
	}
}

func TearDownServices(t *testing.T, paths []string) {
	for _, path := range paths {
		serviceName := getServiceName(path)
		integrationtest.DelSVC(t, "default", serviceName)
	}
}

func SetUpServiceImport(t *testing.T, paths []string) {

	for _, path := range paths {
		siObj := integrationtest.FakeServiceImport{
			Name:          getServiceImportName(path),
			Cluster:       getClusterName(path),
			Namespace:     "default",
			ServiceName:   getServiceName(path),
			EndPointIPs:   []string{"100.1.1.1", "100.1.1.2"},
			EndPointPorts: []int32{31030, 31030},
		}
		fakeSI := siObj.Create()
		if _, err := CRDClient.AkoV1alpha1().ServiceImports(utils.GetAKONamespace()).Create(context.TODO(), fakeSI, metav1.CreateOptions{}); err != nil {
			t.Fatalf("error in adding service import: %v", err)
		}
	}
}

func TearDownServiceImport(t *testing.T, paths []string) {

	for _, path := range paths {
		siName := getServiceImportName(path)
		if err := CRDClient.AkoV1alpha1().ServiceImports(utils.GetAKONamespace()).Delete(context.TODO(), siName, metav1.DeleteOptions{}); err != nil {
			t.Fatalf("error in deleting service imports: %v", err)
		}
	}
}

func SetUpMultiClusterIngress(t *testing.T, paths []string) {

	ingressObject := integrationtest.FakeMultiClusterIngress{
		Name:       getMultiClusterIngressName(paths[0]),
		HostName:   fmt.Sprintf("%s.com", paths[0]),
		SecretName: "my-secret",
	}
	for i, path := range paths {
		cluster := getClusterName(path)
		weight := 10 + i*10
		serviceName := getServiceName(path)
		ingressObject.Namespaces = append(ingressObject.Namespaces, "default")
		ingressObject.Ports = append(ingressObject.Ports, 8080)
		ingressObject.Clusters = append(ingressObject.Clusters, cluster)
		ingressObject.Weights = append(ingressObject.Weights, weight)
		ingressObject.Paths = append(ingressObject.Paths, path)
		ingressObject.ServiceNames = append(ingressObject.ServiceNames, serviceName)
	}

	fakeMCI := ingressObject.Create()
	if _, err := CRDClient.AkoV1alpha1().MultiClusterIngresses(utils.GetAKONamespace()).Create(context.TODO(), fakeMCI, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding multi-cluster Ingress: %v", err)
	}
}

func UpdateMultiClusterIngress(t *testing.T, paths []string) {

	ingressObject := integrationtest.FakeMultiClusterIngress{
		Name:       getMultiClusterIngressName(paths[0]),
		HostName:   fmt.Sprintf("%s.com", paths[0]),
		SecretName: "my-secret",
	}
	for i, path := range paths {
		cluster := getClusterName(path)
		weight := 10 + i*10
		serviceName := getServiceName(path)
		ingressObject.Namespaces = append(ingressObject.Namespaces, "default")
		ingressObject.Ports = append(ingressObject.Ports, 8080)
		ingressObject.Clusters = append(ingressObject.Clusters, cluster)
		ingressObject.Weights = append(ingressObject.Weights, weight)
		ingressObject.Paths = append(ingressObject.Paths, path)
		ingressObject.ServiceNames = append(ingressObject.ServiceNames, serviceName)
	}

	fakeMCI := ingressObject.Create()
	if _, err := CRDClient.AkoV1alpha1().MultiClusterIngresses(utils.GetAKONamespace()).Update(context.TODO(), fakeMCI, metav1.UpdateOptions{}); err != nil {
		t.Fatalf("error occurred while updating multi-cluster Ingress: %v", err)
	}

}

func TearDownMultiClusterIngress(t *testing.T, str string) {
	mciName := getMultiClusterIngressName(str)
	if err := CRDClient.AkoV1alpha1().MultiClusterIngresses(utils.GetAKONamespace()).Delete(context.TODO(), mciName, metav1.DeleteOptions{}); err != nil {
		t.Fatalf("Couldn't DELETE the multi-cluster Ingress %v", err)
	}
}

func SetUpTest(t *testing.T, withSecret bool, paths []string, modelName string) {
	SetupDomain()

	cleanupModels(modelName)

	SetUpServices(t, paths)

	if withSecret {
		integrationtest.AddSecret("my-secret", utils.GetAKONamespace(), "tlsCert", "tlsKey")
	}
	SetUpMultiClusterIngress(t, paths)
	SetUpServiceImport(t, paths)
}

func TearDownTest(t *testing.T, paths []string, modelName string) {

	TearDownMultiClusterIngress(t, paths[0])
	TearDownServiceImport(t, paths)
	cleanupModels(modelName)
	TearDownServices(t, paths)

	KubeClient.CoreV1().Secrets(utils.GetAKONamespace()).Delete(context.TODO(), "my-secret", metav1.DeleteOptions{})
}

func GetModelName(hostname, namespace string) (string, string) {
	vsName := "cluster--Shared-L7-EVH-"
	if !lib.VIPPerNamespace() {
		vsName += strconv.Itoa(int(utils.Bkt(hostname, 8)))
		return "admin/" + vsName, vsName
	}
	vsName += "NS-" + namespace
	return "admin/" + vsName, vsName
}

func TestL7ModelForEvh(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName, _ := GetModelName("foo.com", utils.GetAKONamespace())

	paths := []string{"foo"}
	SetUpTest(t, true, paths, modelName)
	integrationtest.PollForCompletion(t, modelName, 5)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 20*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(len(nodes)).To(gomega.Equal(1))
	g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
	g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
	g.Expect(len(nodes[0].EvhNodes)).To(gomega.Equal(1))

	TearDownTest(t, paths, modelName)

	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	if aviModel != nil {
		t.Fatalf("Multi-cluster Ingress model is not properly removed")
	}
}

func TestMultiPathMultiClusterIngress(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName, _ := GetModelName("foo.com", utils.GetAKONamespace())

	paths := []string{"foo", "bar"}
	SetUpTest(t, true, paths, modelName)
	integrationtest.PollForCompletion(t, modelName, 5)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 5*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(len(nodes)).To(gomega.Equal(1))
	g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
	g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
	g.Expect(len(nodes[0].EvhNodes)).To(gomega.Equal(1))

	// Validate pools are populated correctly
	g.Expect(len(nodes[0].EvhNodes[0].PoolRefs)).To(gomega.Equal(2))
	g.Expect(len(nodes[0].EvhNodes[0].PoolRefs[0].Servers)).To(gomega.Equal(2))
	g.Expect(len(nodes[0].EvhNodes[0].PoolRefs[1].Servers)).To(gomega.Equal(2))

	// Validate servers are populated with IPs
	g.Expect(len(*nodes[0].EvhNodes[0].PoolRefs[0].Servers[0].Ip.Addr)).ShouldNot(gomega.BeNil())
	g.Expect(len(*nodes[0].EvhNodes[0].PoolRefs[0].Servers[1].Ip.Addr)).ShouldNot(gomega.BeNil())
	g.Expect(len(*nodes[0].EvhNodes[0].PoolRefs[1].Servers[0].Ip.Addr)).ShouldNot(gomega.BeNil())
	g.Expect(len(*nodes[0].EvhNodes[0].PoolRefs[1].Servers[1].Ip.Addr)).ShouldNot(gomega.BeNil())

	TearDownTest(t, paths, modelName)
}

func TestUpdateBackendConfig(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName, _ := GetModelName("foo.com", utils.GetAKONamespace())

	paths := []string{"foo", "bar"}
	SetUpTest(t, true, paths, modelName)
	integrationtest.PollForCompletion(t, modelName, 5)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 5*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(len(nodes)).To(gomega.Equal(1))
	g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
	g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
	g.Expect(len(nodes[0].EvhNodes)).To(gomega.Equal(1))

	// Validate whether pools are populated
	g.Expect(len(nodes[0].EvhNodes[0].PoolRefs)).To(gomega.Equal(2))
	g.Expect(len(nodes[0].EvhNodes[0].PoolRefs[0].Servers)).To(gomega.Equal(2))
	g.Expect(len(nodes[0].EvhNodes[0].PoolRefs[1].Servers)).To(gomega.Equal(2))

	paths = append(paths, "baz")
	UpdateMultiClusterIngress(t, paths)
	SetUpServiceImport(t, []string{"baz"})

	// Validate whether pools are updated with empty servers
	g.Eventually(func() int {
		_, aviModel = objects.SharedAviGraphLister().Get(modelName)
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return len(nodes[0].EvhNodes[0].PoolRefs)
	}, 60*time.Second, 30*time.Second).Should(gomega.Equal(3))

	// Validate whether servers are populated with IPs
	g.Expect(len(*nodes[0].EvhNodes[0].PoolRefs[0].Servers[0].Ip.Addr)).ShouldNot(gomega.BeNil())
	g.Expect(len(*nodes[0].EvhNodes[0].PoolRefs[0].Servers[1].Ip.Addr)).ShouldNot(gomega.BeNil())
	g.Expect(len(*nodes[0].EvhNodes[0].PoolRefs[1].Servers[0].Ip.Addr)).ShouldNot(gomega.BeNil())
	g.Expect(len(*nodes[0].EvhNodes[0].PoolRefs[1].Servers[1].Ip.Addr)).ShouldNot(gomega.BeNil())
	g.Expect(len(*nodes[0].EvhNodes[0].PoolRefs[2].Servers[0].Ip.Addr)).ShouldNot(gomega.BeNil())
	g.Expect(len(*nodes[0].EvhNodes[0].PoolRefs[2].Servers[1].Ip.Addr)).ShouldNot(gomega.BeNil())

	TearDownTest(t, paths, modelName)
}

func TestDeleteBackendConfig(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName, _ := GetModelName("foo.com", utils.GetAKONamespace())

	paths := []string{"foo", "bar"}
	SetUpTest(t, true, paths, modelName)
	integrationtest.PollForCompletion(t, modelName, 5)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 5*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(len(nodes)).To(gomega.Equal(1))
	g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
	g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
	g.Expect(len(nodes[0].EvhNodes)).To(gomega.Equal(1))

	// Validate whether pools are populated
	g.Expect(len(nodes[0].EvhNodes[0].PoolRefs)).To(gomega.Equal(2))
	g.Expect(len(nodes[0].EvhNodes[0].PoolRefs[0].Servers)).To(gomega.Equal(2))
	g.Expect(len(nodes[0].EvhNodes[0].PoolRefs[1].Servers)).To(gomega.Equal(2))

	TearDownServiceImport(t, paths)

	// Validate whether pools are updated with empty servers
	g.Eventually(func() int {
		_, aviModel = objects.SharedAviGraphLister().Get(modelName)
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return len(nodes[0].EvhNodes[0].PoolRefs[0].Servers) +
			len(nodes[0].EvhNodes[0].PoolRefs[1].Servers)
	}, 60*time.Second, 30*time.Second).Should(gomega.Equal(0))

	g.Expect(len(nodes[0].EvhNodes[0].PoolRefs)).To(gomega.Equal(2))

	// Remove the ingress, models, services and secret
	TearDownMultiClusterIngress(t, paths[0])
	cleanupModels(modelName)
	TearDownServices(t, paths)
	KubeClient.CoreV1().Secrets(utils.GetAKONamespace()).Delete(context.TODO(), "my-secret", metav1.DeleteOptions{})
}
