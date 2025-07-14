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

package npltests

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/cache"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/k8s"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	avinodes "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/nodes"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"
	crdfake "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/client/v1alpha1/clientset/versioned/fake"
	v1beta1crdfake "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/client/v1beta1/clientset/versioned/fake"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/tests/integrationtest"

	utils "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	"github.com/onsi/gomega"
	"github.com/vmware/alb-sdk/go/models"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	k8sfake "k8s.io/client-go/kubernetes/fake"
)

var KubeClient *k8sfake.Clientset
var CRDClient *crdfake.Clientset
var V1beta1CRDClient *v1beta1crdfake.Clientset
var ctrl *k8s.AviController

const (
	defaultPodName  = "test-pod"
	defaultNS       = "default"
	defaultNodeName = "test-node"
	defaultHostIP   = "10.10.10.10"
	defaultPodIP    = "192.168.32.10"
	defaultPodPort  = 80
	defaultNodePort = 61000
	defaultLBModel  = "admin/cluster--default-testsvc"
	defaultL7Model  = "admin/cluster--Shared-L7-0"
)

func SetUpTestForIngress(t *testing.T, modelName string) {
	objects.SharedAviGraphLister().Delete(modelName)
}

func createPodWithNPLAnnotation(labels map[string]string) {
	testPod := getTestPod(labels)
	ann := make(map[string]string)
	ann[lib.NPLPodAnnotation] = "[{\"podPort\":8080,\"nodeIP\":\"10.10.10.10\",\"nodePort\":61000}]"
	testPod.Annotations = ann
	KubeClient.CoreV1().Pods(defaultNS).Create(context.TODO(), &testPod, metav1.CreateOptions{})
}

func createNotReadyPodWithNPLAnnotation(labels map[string]string) {
	testPod := getTestPod(labels)
	ann := make(map[string]string)
	ann[lib.NPLPodAnnotation] = "[{\"podPort\":8080,\"nodeIP\":\"10.10.10.10\",\"nodePort\":61000}]"
	testPod.Annotations = ann
	testPod.Status.Conditions = append(testPod.Status.Conditions, corev1.PodCondition{Type: "Ready", Status: "False"})
	KubeClient.CoreV1().Pods(defaultNS).Create(context.TODO(), &testPod, metav1.CreateOptions{})
}

func updateNotReadyPodWithNPLAnnotation(labels map[string]string) {
	testPod := getTestPod(labels)
	ann := make(map[string]string)
	ann[lib.NPLPodAnnotation] = "[{\"podPort\":8080,\"nodeIP\":\"10.10.10.10\",\"nodePort\":61000}]"
	testPod.Annotations = ann
	testPod.ResourceVersion = "3"
	testPod.Status.Conditions = append(testPod.Status.Conditions, corev1.PodCondition{Type: "Ready", Status: "False"})
	KubeClient.CoreV1().Pods(defaultNS).Update(context.TODO(), &testPod, metav1.UpdateOptions{})
}

func createPodWithMultipleNPLAnnotations(labels map[string]string) {
	testPod := getTestPod(labels)
	ann := make(map[string]string)
	ann[lib.NPLPodAnnotation] = "[{\"podPort\":8080,\"nodeIP\":\"10.10.10.10\",\"nodePort\":61000}, {\"podPort\":8081,\"nodeIP\":\"10.10.10.10\",\"nodePort\":61001}]"
	testPod.Annotations = ann
	KubeClient.CoreV1().Pods(defaultNS).Create(context.TODO(), &testPod, metav1.CreateOptions{})
}

func updatePodWithNPLAnnotation(labels map[string]string) {
	testPod := getTestPod(labels)
	ann := make(map[string]string)
	ann[lib.NPLPodAnnotation] = "[{\"podPort\":8080,\"nodeIP\":\"10.10.10.10\",\"nodePort\":61000}]"
	testPod.Annotations = ann
	testPod.ResourceVersion = "2"
	KubeClient.CoreV1().Pods(defaultNS).Update(context.TODO(), &testPod, metav1.UpdateOptions{})
}

func setUpTestForSvcLB(t *testing.T) {
	objects.SharedAviGraphLister().Delete(integrationtest.SINGLEPORTMODEL)
	selectors := make(map[string]string)
	selectors["app"] = "npl"
	svcExample := integrationtest.ConstructService(defaultNS, integrationtest.SINGLEPORTSVC, corev1.ProtocolTCP, corev1.ServiceTypeLoadBalancer, false, selectors, "")
	svcExample.Annotations = make(map[string]string)
	svcExample.Annotations[lib.NPLSvcAnnotation] = "true"
	_, err := KubeClient.CoreV1().Services(defaultNS).Create(context.TODO(), svcExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Service: %v", err)
	}

	integrationtest.PollForCompletion(t, defaultLBModel, 5)
}

func tearDownTestForSvcLB(t *testing.T, g *gomega.GomegaWithT) {
	objects.SharedAviGraphLister().Delete(integrationtest.SINGLEPORTMODEL)
	integrationtest.DelSVC(t, defaultNS, integrationtest.SINGLEPORTSVC)
	mcache := cache.SharedAviObjCache()
	vsKey := cache.NamespaceName{Namespace: integrationtest.AVINAMESPACE, Name: fmt.Sprintf("cluster--%s-%s", integrationtest.SINGLEPORTSVC, defaultNS)}
	g.Eventually(func() bool {
		_, found := mcache.VsCacheMeta.AviCacheGet(vsKey)
		return found
	}, 40*time.Second).Should(gomega.Equal(false))
	KubeClient.CoreV1().Pods(defaultNS).Delete(context.TODO(), defaultPodName, metav1.DeleteOptions{})
}

func verifyIngressDeletion(t *testing.T, g *gomega.WithT, aviModel interface{}, poolCount int) {
	var nodes []*avinodes.AviVsNode
	g.Eventually(func() []*avinodes.AviPoolNode {
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return nodes[0].PoolRefs
	}, 40*time.Second).Should(gomega.HaveLen(poolCount))

	g.Eventually(func() []*models.PoolGroupMember {
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return nodes[0].PoolGroupRefs[0].Members
	}, 40*time.Second).Should(gomega.HaveLen(poolCount))
}

func TearDownTestForIngress(t *testing.T, modelName string) {
	objects.SharedAviGraphLister().Delete(modelName)
	integrationtest.DelSVC(t, "default", "avisvc")
	KubeClient.CoreV1().Pods(defaultNS).Delete(context.TODO(), defaultPodName, metav1.DeleteOptions{})
}

func getTestPod(labels map[string]string) corev1.Pod {
	testPod := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      defaultPodName,
			Namespace: defaultNS,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			NodeName: defaultNodeName,
			Containers: []corev1.Container{
				{
					Ports: []corev1.ContainerPort{
						{
							ContainerPort: int32(defaultPodPort),
						},
					},
				},
			},
		},
		Status: corev1.PodStatus{
			HostIP: defaultHostIP,
			PodIP:  defaultPodIP,
		},
	}
	return testPod
}

func TestMain(m *testing.M) {
	os.Setenv("SERVICE_TYPE", "NodePortLocal")
	os.Setenv("INGRESS_API", "extensionv1")
	os.Setenv("VIP_NETWORK_LIST", `[{"networkName":"net123"}]`)
	os.Setenv("CLUSTER_NAME", "cluster")
	os.Setenv("CLOUD_NAME", "CLOUD_VCENTER")
	os.Setenv("SEG_NAME", "Default-Group")
	os.Setenv("NODE_NETWORK_LIST", `[{"networkName":"net123","cidrs":["10.79.168.0/22"]}]`)
	os.Setenv("CNI_PLUGIN", "antrea")
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
	akoControlConfig.SetDefaultLBController(true)
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
		utils.IngressInformer,
		utils.IngressClassInformer,
		utils.SecretInformer,
		utils.NSInformer,
		utils.NodeInformer,
		utils.ConfigMapInformer,
		utils.PodInformer,
	}

	registeredInformers = append(registeredInformers, utils.EndpointSlicesInformer)
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
	integrationtest.AddDefaultNamespace()

	go ctrl.InitController(informers, registeredInformers, ctrlCh, stopCh, quickSyncCh, waitGroupMap)
	os.Exit(m.Run())
}

// TestIngressAddPod creates a POD with NPL annotation and corresponding Service and Ingress, then verifies the model.
func TestIngressAddPod(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	SetUpTestForIngress(t, defaultL7Model)
	selectors := make(map[string]string)
	selectors["app"] = "npl"
	integrationtest.CreateServiceWithSelectors(t, defaultNS, "avisvc", corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false, selectors)
	createPodWithNPLAnnotation(selectors)

	integrationtest.PollForCompletion(t, defaultL7Model, 10)
	found, _ := objects.SharedAviGraphLister().Get(defaultL7Model)
	if found {
		t.Fatalf("Couldn't find Model for DELETE event %v", defaultL7Model)
	}
	ingrFake := (integrationtest.FakeIngress{
		Name:        "foo-with-targets",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		ServiceName: "avisvc",
	}).Ingress()

	_, err := KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	integrationtest.PollForCompletion(t, defaultL7Model, 10)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(defaultL7Model)
		return found
	}, 40*time.Second).Should(gomega.Equal(true))

	_, aviModel := objects.SharedAviGraphLister().Get(defaultL7Model)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(len(nodes)).To(gomega.Equal(1))
	g.Expect(nodes[0].PoolRefs).To(gomega.HaveLen(1))

	g.Eventually(func() int {
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return len(nodes[0].PoolRefs[0].Servers)
	}, 40*time.Second).Should(gomega.Equal(1))
	g.Expect(nodes[0].PoolRefs[0].Servers).To(gomega.HaveLen(1))
	g.Expect(*nodes[0].PoolRefs[0].Servers[0].Ip.Addr).To(gomega.Equal(defaultHostIP))
	g.Expect(nodes[0].PoolRefs[0].Servers[0].Port).To(gomega.Equal(int32(defaultNodePort)))

	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), "foo-with-targets", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	verifyIngressDeletion(t, g, aviModel, 0)
	TearDownTestForIngress(t, defaultL7Model)
}

// TestIngressDelPod creates a POD with NPL annotation and corresponding Service and Ingress, then verifies the model.
// Then the Pod is deleted and it is verified that the server is removed from the model
func TestIngressDelPod(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	SetUpTestForIngress(t, defaultL7Model)
	selectors := make(map[string]string)
	selectors["app"] = "npl"
	integrationtest.CreateServiceWithSelectors(t, defaultNS, "avisvc", corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false, selectors)
	createPodWithNPLAnnotation(selectors)

	integrationtest.PollForCompletion(t, defaultL7Model, 10)
	found, _ := objects.SharedAviGraphLister().Get(defaultL7Model)
	if found {
		t.Fatalf("Couldn't find Model for DELETE event %v", defaultL7Model)
	}
	ingrFake := (integrationtest.FakeIngress{
		Name:        "foo-with-targets",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		ServiceName: "avisvc",
	}).Ingress()

	_, err := KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	integrationtest.PollForCompletion(t, defaultL7Model, 10)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(defaultL7Model)
		return found
	}, 40*time.Second).Should(gomega.Equal(true))

	_, aviModel := objects.SharedAviGraphLister().Get(defaultL7Model)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(len(nodes)).To(gomega.Equal(1))
	g.Expect(nodes[0].PoolRefs).To(gomega.HaveLen(1))

	g.Eventually(func() int {
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return len(nodes[0].PoolRefs[0].Servers)
	}, 40*time.Second).Should(gomega.Equal(1))
	g.Expect(nodes[0].PoolRefs[0].Servers).To(gomega.HaveLen(1))
	g.Expect(*nodes[0].PoolRefs[0].Servers[0].Ip.Addr).To(gomega.Equal(defaultHostIP))
	g.Expect(nodes[0].PoolRefs[0].Servers[0].Port).To(gomega.Equal(int32(defaultNodePort)))

	err = KubeClient.CoreV1().Pods(defaultNS).Delete(context.TODO(), defaultPodName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("error in deleting Pod: %v", err)
	}
	g.Eventually(func() int {
		_, aviModel = objects.SharedAviGraphLister().Get(defaultL7Model)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return len(nodes[0].PoolRefs[0].Servers)
	}, 40*time.Second).Should(gomega.Equal(0))

	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), "foo-with-targets", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	verifyIngressDeletion(t, g, aviModel, 0)
	TearDownTestForIngress(t, defaultL7Model)
}

// TestIngressAddPodWithoutLabel creates a Service, an Ingress, and a Pod without matching label,
// then verifies in the model that no server is added.
func TestIngressAddPodWithoutLabel(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	SetUpTestForIngress(t, defaultL7Model)
	selectors := make(map[string]string)
	integrationtest.CreateServiceWithSelectors(t, defaultNS, "avisvc", corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false, selectors)
	createPodWithNPLAnnotation(selectors)

	integrationtest.PollForCompletion(t, defaultL7Model, 10)
	found, _ := objects.SharedAviGraphLister().Get(defaultL7Model)
	if found {
		t.Fatalf("Couldn't find Model for DELETE event %v", defaultL7Model)
	}
	ingrFake := (integrationtest.FakeIngress{
		Name:        "foo-with-targets",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		ServiceName: "avisvc",
	}).Ingress()

	_, err := KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	integrationtest.PollForCompletion(t, defaultL7Model, 10)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(defaultL7Model)
		return found
	}, 40*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(defaultL7Model)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(len(nodes)).To(gomega.Equal(1))
	g.Expect(nodes[0].PoolRefs).To(gomega.HaveLen(1))
	g.Expect(nodes[0].PoolRefs[0].Servers).To(gomega.HaveLen(0))

	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), "foo-with-targets", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	verifyIngressDeletion(t, g, aviModel, 0)
	TearDownTestForIngress(t, defaultL7Model)
}

// TestIngressUpdatePodWithLabel creates a Service, an Ingress, and a Pod without matching label.
// Then the Pod is updated with correct label and then the model is verified.
func TestIngressUpdatePodWithLabel(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	SetUpTestForIngress(t, defaultL7Model)
	selectors := make(map[string]string)
	selectors["app"] = "npl"
	integrationtest.CreateServiceWithSelectors(t, defaultNS, "avisvc", corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false, selectors)
	labels := make(map[string]string)
	createPodWithNPLAnnotation(labels)

	integrationtest.PollForCompletion(t, defaultL7Model, 10)
	found, _ := objects.SharedAviGraphLister().Get(defaultL7Model)
	if found {
		t.Fatalf("Couldn't find Model for DELETE event %v", defaultL7Model)
	}
	ingrFake := (integrationtest.FakeIngress{
		Name:        "foo-with-targets",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		ServiceName: "avisvc",
	}).Ingress()

	_, err := KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	integrationtest.PollForCompletion(t, defaultL7Model, 10)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(defaultL7Model)
		return found
	}, 40*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(defaultL7Model)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(len(nodes)).To(gomega.Equal(1))
	g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
	g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
	g.Expect(nodes[0].PoolRefs).To(gomega.HaveLen(1))
	g.Expect(nodes[0].PoolRefs[0].Servers).To(gomega.HaveLen(0))

	labels["app"] = "npl"
	updatePodWithNPLAnnotation(labels)
	g.Eventually(func() int {
		_, aviModel := objects.SharedAviGraphLister().Get(defaultL7Model)
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return len(nodes[0].PoolRefs[0].Servers)
	}, 40*time.Second).Should(gomega.Equal(1))
	g.Expect(nodes[0].PoolRefs[0].Servers).To(gomega.HaveLen(1))
	g.Expect(*nodes[0].PoolRefs[0].Servers[0].Ip.Addr).To(gomega.Equal(defaultHostIP))
	g.Expect(nodes[0].PoolRefs[0].Servers[0].Port).To(gomega.Equal(int32(defaultNodePort)))

	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), "foo-with-targets", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	verifyIngressDeletion(t, g, aviModel, 0)
	TearDownTestForIngress(t, defaultL7Model)
}

// TestIngressUpdatePodWithoutLabel creates a Service, an Ingress, and a Pod with matching label.
// Then the Pod is updated with the label removed and then the model is verified.
func TestIngressUpdatePodWithoutLabel(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	SetUpTestForIngress(t, defaultL7Model)
	selectors := make(map[string]string)
	selectors["app"] = "npl"
	integrationtest.CreateServiceWithSelectors(t, defaultNS, "avisvc", corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false, selectors)
	createPodWithNPLAnnotation(selectors)

	integrationtest.PollForCompletion(t, defaultL7Model, 5)
	found, _ := objects.SharedAviGraphLister().Get(defaultL7Model)
	if found {
		t.Fatalf("Couldn't find Model for DELETE event %v", defaultL7Model)
	}
	ingrFake := (integrationtest.FakeIngress{
		Name:        "foo-with-targets",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		ServiceName: "avisvc",
	}).Ingress()

	_, err := KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	integrationtest.PollForCompletion(t, defaultL7Model, 10)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(defaultL7Model)
		return found
	}, 40*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(defaultL7Model)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(len(nodes)).To(gomega.Equal(1))
	g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
	g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
	g.Expect(nodes[0].PoolRefs).To(gomega.HaveLen(1))
	g.Eventually(func() int {
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return len(nodes[0].PoolRefs[0].Servers)
	}, 40*time.Second).Should(gomega.Equal(1))
	g.Expect(nodes[0].PoolRefs[0].Servers).To(gomega.HaveLen(1))
	g.Expect(*nodes[0].PoolRefs[0].Servers[0].Ip.Addr).To(gomega.Equal(defaultHostIP))
	g.Expect(nodes[0].PoolRefs[0].Servers[0].Port).To(gomega.Equal(int32(defaultNodePort)))

	labels := make(map[string]string)
	updatePodWithNPLAnnotation(labels)
	g.Eventually(func() int {
		_, aviModel := objects.SharedAviGraphLister().Get(defaultL7Model)
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return len(nodes[0].PoolRefs[0].Servers)
	}, 40*time.Second).Should(gomega.Equal(0))

	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), "foo-with-targets", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	verifyIngressDeletion(t, g, aviModel, 0)
	TearDownTestForIngress(t, defaultL7Model)
}

// TestIngressDelSvc creates a POD with NPL annotation and corresponding Service and Ingress, then verifies the model.
// Then the Service is deleted and it is verified that corresponding servers are deleted from the model.
func TestIngressDelSvc(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	SetUpTestForIngress(t, defaultL7Model)
	selectors := make(map[string]string)
	selectors["app"] = "npl"
	integrationtest.CreateServiceWithSelectors(t, defaultNS, "avisvc", corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false, selectors)
	createPodWithNPLAnnotation(selectors)

	integrationtest.PollForCompletion(t, defaultL7Model, 10)
	found, _ := objects.SharedAviGraphLister().Get(defaultL7Model)
	if found {
		t.Fatalf("Couldn't find Model for DELETE event %v", defaultL7Model)
	}
	ingrFake := (integrationtest.FakeIngress{
		Name:        "foo-with-targets",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		ServiceName: "avisvc",
	}).Ingress()

	_, err := KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	integrationtest.PollForCompletion(t, defaultL7Model, 10)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(defaultL7Model)
		return found
	}, 40*time.Second).Should(gomega.Equal(true))

	_, aviModel := objects.SharedAviGraphLister().Get(defaultL7Model)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(len(nodes)).To(gomega.Equal(1))
	g.Expect(nodes[0].PoolRefs).To(gomega.HaveLen(1))

	g.Eventually(func() int {
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return len(nodes[0].PoolRefs[0].Servers)
	}, 40*time.Second).Should(gomega.Equal(1))
	g.Expect(nodes[0].PoolRefs[0].Servers).To(gomega.HaveLen(1))
	g.Expect(*nodes[0].PoolRefs[0].Servers[0].Ip.Addr).To(gomega.Equal(defaultHostIP))
	g.Expect(nodes[0].PoolRefs[0].Servers[0].Port).To(gomega.Equal(int32(defaultNodePort)))

	integrationtest.DelSVC(t, "default", "avisvc")
	g.Eventually(func() int {
		_, aviModel := objects.SharedAviGraphLister().Get(defaultL7Model)
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return len(nodes[0].PoolRefs[0].Servers)
	}, 40*time.Second).Should(gomega.Equal(0))

	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), "foo-with-targets", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	verifyIngressDeletion(t, g, aviModel, 0)
	TearDownTestForIngress(t, defaultL7Model)
}

// TestNPLLBSvc creates a Service type LB and a Pod with matching label, then the model is verified.
func TestNPLLBSvc(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	selectors := make(map[string]string)
	selectors["app"] = "npl"
	createPodWithNPLAnnotation(selectors)
	setUpTestForSvcLB(t)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(defaultLBModel)
		return found
	}, 40*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(defaultLBModel)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].Name).To(gomega.Equal(fmt.Sprintf("cluster--%s-%s", defaultNS, integrationtest.SINGLEPORTSVC)))
	g.Expect(nodes[0].Tenant).To(gomega.Equal(integrationtest.AVINAMESPACE))
	g.Expect(nodes[0].PortProto[0].Port).To(gomega.Equal(int32(8080)))

	g.Expect(nodes[0].PoolRefs).To(gomega.HaveLen(1))
	g.Eventually(func() int {
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return len(nodes[0].PoolRefs[0].Servers)
	}, 40*time.Second).Should(gomega.Equal(1))
	address := defaultHostIP
	g.Expect(nodes[0].PoolRefs[0].Servers[0].Ip.Addr).To(gomega.Equal(&address))

	// If we transition the service from Loadbalancer to ClusterIP - it should get deleted.
	svcExample := (integrationtest.FakeService{
		Name:         integrationtest.SINGLEPORTSVC,
		Namespace:    defaultNS,
		Type:         corev1.ServiceTypeClusterIP,
		ServicePorts: []integrationtest.Serviceport{{PortName: "foo1", Protocol: "TCP", PortNumber: 8080, TargetPort: intstr.FromInt(8080)}},
	}).Service()
	svcExample.Annotations = make(map[string]string)
	svcExample.Annotations[lib.NPLSvcAnnotation] = "true"
	svcExample.ResourceVersion = "2"
	_, err := KubeClient.CoreV1().Services(defaultNS).Update(context.TODO(), svcExample, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in adding Service: %v", err)
	}
	mcache := cache.SharedAviObjCache()
	vsKey := cache.NamespaceName{Namespace: integrationtest.AVINAMESPACE, Name: fmt.Sprintf("cluster--%s-%s", defaultNS, integrationtest.SINGLEPORTSVC)}
	g.Eventually(func() bool {
		_, found := mcache.VsCacheMeta.AviCacheGet(vsKey)
		return found
	}, 40*time.Second).Should(gomega.Equal(false))

	// If we transition the service from clusterIP to Loadbalancer - vs should get created
	svcExample = (integrationtest.FakeService{
		Name:         integrationtest.SINGLEPORTSVC,
		Namespace:    defaultNS,
		Type:         corev1.ServiceTypeLoadBalancer,
		ServicePorts: []integrationtest.Serviceport{{PortName: "foo1", Protocol: "TCP", PortNumber: 8080, TargetPort: intstr.FromInt(8080)}},
	}).Service()
	svcExample.Annotations = make(map[string]string)
	svcExample.Annotations[lib.NPLSvcAnnotation] = "true"
	svcExample.ResourceVersion = "3"
	_, err = KubeClient.CoreV1().Services(defaultNS).Update(context.TODO(), svcExample, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in adding Service: %v", err)
	}

	g.Eventually(func() bool {
		_, found := mcache.VsCacheMeta.AviCacheGet(vsKey)
		return found
	}, 40*time.Second).Should(gomega.Equal(true))
	tearDownTestForSvcLB(t, g)
}

// TestNPLLBSvcDelPod creates a Service type LB and a Pod with matching label and the model is verified.
// Then the Pod is deleted, and it is verified that the Server is deleted from model
func TestNPLLBSvcDelPod(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	selectors := make(map[string]string)
	selectors["app"] = "npl"
	objects.SharedAviGraphLister().Delete(defaultLBModel)
	createPodWithNPLAnnotation(selectors)
	setUpTestForSvcLB(t)

	var aviModel interface{}
	var found bool
	g.Eventually(func() bool {
		found, aviModel = objects.SharedAviGraphLister().Get(defaultLBModel)
		return aviModel == nil
	}, 40*time.Second).Should(gomega.Equal(false))
	if !found {
		t.Fatalf("Couldn't find model %v", defaultLBModel)
	}
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].Name).To(gomega.Equal(fmt.Sprintf("cluster--%s-%s", defaultNS, integrationtest.SINGLEPORTSVC)))
	g.Expect(nodes[0].Tenant).To(gomega.Equal(integrationtest.AVINAMESPACE))
	g.Expect(nodes[0].PortProto[0].Port).To(gomega.Equal(int32(8080)))

	g.Expect(nodes[0].PoolRefs).To(gomega.HaveLen(1))
	g.Eventually(func() int {
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return len(nodes[0].PoolRefs[0].Servers)
	}, 40*time.Second).Should(gomega.Equal(1))
	address := defaultHostIP
	g.Expect(nodes[0].PoolRefs[0].Servers[0].Ip.Addr).To(gomega.Equal(&address))

	// If we delete the Pod, the server should get deleted from model
	err := KubeClient.CoreV1().Pods(defaultNS).Delete(context.TODO(), defaultPodName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("error in deleting Pod: %v", err)
	}
	g.Eventually(func() int {
		_, aviModel = objects.SharedAviGraphLister().Get(defaultLBModel)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return len(nodes[0].PoolRefs[0].Servers)
	}, 40*time.Second).Should(gomega.Equal(0))
	tearDownTestForSvcLB(t, g)
}

// TestNPLLBSvcNoLabel creates a Service of type LB with no Label and a Pod with NPL annotation.
// Then it is verified that no server is getting added in the model.
func TestNPLLBSvcNoLabel(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	labels := make(map[string]string)
	labels["app"] = "npl"
	createPodWithNPLAnnotation(labels)

	objects.SharedAviGraphLister().Delete(integrationtest.SINGLEPORTMODEL)
	selectors := make(map[string]string)
	integrationtest.CreateServiceWithSelectors(t, defaultNS, integrationtest.SINGLEPORTSVC, corev1.ProtocolTCP, corev1.ServiceTypeLoadBalancer, false, selectors)
	integrationtest.PollForCompletion(t, defaultLBModel, 5)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(defaultLBModel)
		return found
	}, 10*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(defaultLBModel)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].Name).To(gomega.Equal(fmt.Sprintf("cluster--%s-%s", defaultNS, integrationtest.SINGLEPORTSVC)))
	g.Expect(nodes[0].Tenant).To(gomega.Equal(integrationtest.AVINAMESPACE))
	g.Expect(nodes[0].PortProto[0].Port).To(gomega.Equal(int32(8080)))

	g.Eventually(func() int {
		if nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS(); len(nodes) > 0 {
			return len(nodes[0].PoolRefs)
		}
		return 0
	}, 40*time.Second).Should(gomega.Equal(1))
	g.Expect(nodes[0].PoolRefs).To(gomega.HaveLen(1))
	g.Expect(nodes[0].PoolRefs[0].Servers).To(gomega.HaveLen(0))

	tearDownTestForSvcLB(t, g)
}

// TestNPLUpdateLBSvcCorrectSelector creates a Service of type LB with no Label and a Pod with NPL annotation.
// Then the service is updated with required selector and the model is verified.
func TestNPLUpdateLBSvcCorrectSelector(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	labels := make(map[string]string)
	labels["app"] = "npl"
	createPodWithNPLAnnotation(labels)

	objects.SharedAviGraphLister().Delete(integrationtest.SINGLEPORTMODEL)
	selectors := make(map[string]string)
	integrationtest.CreateServiceWithSelectors(t, defaultNS, integrationtest.SINGLEPORTSVC, corev1.ProtocolTCP, corev1.ServiceTypeLoadBalancer, false, selectors)
	integrationtest.PollForCompletion(t, defaultLBModel, 10)

	selectors["app"] = "npl"
	integrationtest.UpdateServiceWithSelectors(t, defaultNS, integrationtest.SINGLEPORTSVC, corev1.ProtocolTCP, corev1.ServiceTypeLoadBalancer, false, selectors)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(defaultLBModel)
		return found
	}, 40*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(defaultLBModel)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].Name).To(gomega.Equal(fmt.Sprintf("cluster--%s-%s", defaultNS, integrationtest.SINGLEPORTSVC)))
	g.Expect(nodes[0].Tenant).To(gomega.Equal(integrationtest.AVINAMESPACE))
	g.Expect(nodes[0].PortProto[0].Port).To(gomega.Equal(int32(8080)))

	g.Expect(nodes[0].PoolRefs).To(gomega.HaveLen(1))
	g.Eventually(func() int {
		_, aviModel = objects.SharedAviGraphLister().Get(defaultLBModel)
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return len(nodes[0].PoolRefs[0].Servers)
	}, 40*time.Second).Should(gomega.Equal(1))
	address := defaultHostIP
	g.Expect(nodes[0].PoolRefs[0].Servers[0].Ip.Addr).To(gomega.Equal(&address))

	tearDownTestForSvcLB(t, g)
}

// TestSvcIngressAddDel creates a Service and an Ingress which uses that Service.
// It verifies that the Service gets annotated with the NPL annotation, and the annotation
// is removed when the service is deleted
func TestNPLSvcIngressAddDel(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	SetUpTestForIngress(t, defaultL7Model)
	selectors := make(map[string]string)
	selectors["app"] = "npl"
	integrationtest.CreateServiceWithSelectors(t, defaultNS, "avisvc", corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false, selectors)
	createPodWithNPLAnnotation(selectors)

	found, _ := objects.SharedAviGraphLister().Get(defaultL7Model)
	if found {
		t.Fatalf("Couldn't find Model %v", defaultL7Model)
	}
	ingrFake := (integrationtest.FakeIngress{
		Name:        "foo-with-targets",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		ServiceName: "avisvc",
	}).Ingress()

	_, err := KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	g.Eventually(func() bool {
		svc, _ := KubeClient.CoreV1().Services(defaultNS).Get(context.TODO(), "avisvc", metav1.GetOptions{})
		ann := svc.GetAnnotations()
		if _, ok := ann[lib.NPLSvcAnnotation]; ok {
			return true
		}
		return false
	}, 40*time.Second).Should(gomega.Equal(true))

	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), "foo-with-targets", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}

	g.Eventually(func() bool {
		svc, _ := KubeClient.CoreV1().Services(defaultNS).Get(context.TODO(), "avisvc", metav1.GetOptions{})
		ann := svc.GetAnnotations()
		if _, ok := ann[lib.NPLSvcAnnotation]; ok {
			return true
		}
		return false
	}, 40*time.Second).Should(gomega.Equal(false))

	TearDownTestForIngress(t, defaultL7Model)
}

// TestSvcIngressUpdate creates a Service and an Ingress which doesn't use that Service.
// Then the ingress is updated with correct Service and annotation of the Service is Verified.
func TestNPLSvcIngressUpdate(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	SetUpTestForIngress(t, defaultL7Model)
	selectors := make(map[string]string)
	selectors["app"] = "npl"
	integrationtest.CreateServiceWithSelectors(t, defaultNS, "avisvc", corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false, selectors)
	createPodWithNPLAnnotation(selectors)

	ingrFake := (integrationtest.FakeIngress{
		Name:        "foo-with-targets",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		ServiceName: "avisvc-wrong",
	}).Ingress()
	_, err := KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	ingrFake = (integrationtest.FakeIngress{
		Name:        "foo-with-targets",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		ServiceName: "avisvc",
	}).Ingress()
	ingrFake.ResourceVersion = "2"
	_, err = KubeClient.NetworkingV1().Ingresses("default").Update(context.TODO(), ingrFake, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating Ingress: %v", err)
	}

	g.Eventually(func() bool {
		svc, _ := KubeClient.CoreV1().Services(defaultNS).Get(context.TODO(), "avisvc", metav1.GetOptions{})
		ann := svc.GetAnnotations()
		if _, ok := ann[lib.NPLSvcAnnotation]; ok {
			return true
		}
		return false
	}, 40*time.Second).Should(gomega.Equal(true))

	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), "foo-with-targets", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}

	//time.Sleep(10)
	g.Eventually(func() bool {
		svc, _ := KubeClient.CoreV1().Services(defaultNS).Get(context.TODO(), "avisvc", metav1.GetOptions{})
		ann := svc.GetAnnotations()
		if _, ok := ann[lib.NPLSvcAnnotation]; ok {
			return true
		}
		return false
	}, 40*time.Second).Should(gomega.Equal(false))

	TearDownTestForIngress(t, defaultL7Model)
}

// TestSvcIngressUpdateWrongSvc creates a Service and an Ingress which uses that Service.
// Then the ingress is updated with a different Service and it is verified that
// annotation of the original Service is deleted.
func TestNPLSvcIngressUpdateWrongSvc(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	SetUpTestForIngress(t, defaultL7Model)
	selectors := make(map[string]string)
	selectors["app"] = "npl"
	integrationtest.CreateServiceWithSelectors(t, defaultNS, "avisvc", corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false, selectors)
	createPodWithNPLAnnotation(selectors)

	found, _ := objects.SharedAviGraphLister().Get(defaultL7Model)
	if found {
		t.Fatalf("Couldn't find Model %v", defaultL7Model)
	}
	ingrFake := (integrationtest.FakeIngress{
		Name:        "foo-with-targets",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		ServiceName: "avisvc",
	}).Ingress()
	_, err := KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	g.Eventually(func() bool {
		svc, _ := KubeClient.CoreV1().Services(defaultNS).Get(context.TODO(), "avisvc", metav1.GetOptions{})
		ann := svc.GetAnnotations()
		if _, ok := ann[lib.NPLSvcAnnotation]; ok {
			return true
		}
		return false
	}, 40*time.Second).Should(gomega.Equal(true))

	ingrFake = (integrationtest.FakeIngress{
		Name:        "foo-with-targets",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		ServiceName: "avisvc-wrong",
	}).Ingress()
	ingrFake.ResourceVersion = "2"
	_, err = KubeClient.NetworkingV1().Ingresses("default").Update(context.TODO(), ingrFake, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating Ingress: %v", err)
	}

	g.Eventually(func() bool {
		svc, _ := KubeClient.CoreV1().Services(defaultNS).Get(context.TODO(), "avisvc", metav1.GetOptions{})
		ann := svc.GetAnnotations()
		if _, ok := ann[lib.NPLSvcAnnotation]; ok {
			return true
		}
		return false
	}, 40*time.Second).Should(gomega.Equal(false))

	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), "foo-with-targets", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	TearDownTestForIngress(t, defaultL7Model)
}

// TestNPLSvcIngressUpdateClass creates a Service and an Ingress which uses that Service.
// Then the Ingress is updated with a wrong Ingress Class name,
// and it is verified that the NPL annotation is removed from the Service.
// Then the Ingress is updated with correct Ingress Class name,
// and it is verified that the NPL annotation is added in the Service.
func TestNPLSvcIngressUpdateClass(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	SetUpTestForIngress(t, defaultL7Model)
	selectors := make(map[string]string)
	selectors["app"] = "npl"
	integrationtest.CreateServiceWithSelectors(t, defaultNS, "avisvc", corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false, selectors)
	createPodWithNPLAnnotation(selectors)

	found, _ := objects.SharedAviGraphLister().Get(defaultL7Model)
	if found {
		t.Fatalf("Couldn't find Model %v", defaultL7Model)
	}
	ingrFake := (integrationtest.FakeIngress{
		Name:        "foo-with-targets",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		ServiceName: "avisvc",
		ClassName:   integrationtest.DefaultIngressClass,
	}).Ingress()

	_, err := KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	g.Eventually(func() bool {
		svc, _ := KubeClient.CoreV1().Services(defaultNS).Get(context.TODO(), "avisvc", metav1.GetOptions{})
		ann := svc.GetAnnotations()
		if _, ok := ann[lib.NPLSvcAnnotation]; ok {
			return true
		}
		return false
	}, 40*time.Second).Should(gomega.Equal(true))

	wrongClass := "wrong-class"
	ingrFake.Spec.IngressClassName = &wrongClass
	ingrFake.ResourceVersion = "2"
	_, err = KubeClient.NetworkingV1().Ingresses("default").Update(context.TODO(), ingrFake, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("Couldn't Update the Ingress %v", err)
	}

	g.Eventually(func() bool {
		svc, _ := KubeClient.CoreV1().Services(defaultNS).Get(context.TODO(), "avisvc", metav1.GetOptions{})
		ann := svc.GetAnnotations()
		if _, ok := ann[lib.NPLSvcAnnotation]; ok {
			return true
		}
		return false
	}, 20*time.Second).Should(gomega.Equal(false))

	defaultClass := integrationtest.DefaultIngressClass
	ingrFake.Spec.IngressClassName = &defaultClass
	ingrFake.ResourceVersion = "3"
	_, err = KubeClient.NetworkingV1().Ingresses("default").Update(context.TODO(), ingrFake, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("Couldn't Update the Ingress %v", err)
	}
	g.Eventually(func() bool {
		svc, _ := KubeClient.CoreV1().Services(defaultNS).Get(context.TODO(), "avisvc", metav1.GetOptions{})
		ann := svc.GetAnnotations()
		if _, ok := ann[lib.NPLSvcAnnotation]; ok {
			return true
		}
		return false
	}, 20*time.Second).Should(gomega.Equal(true))

	KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), "foo-with-targets", metav1.DeleteOptions{})
	TearDownTestForIngress(t, defaultL7Model)
}

// TestNPLSvcIngressRemoveAddClass creates a Service and an Ingress which uses that Service.
// Then the Ingress Class is removed and it is verified that the NPL annotation is removed from the Service.
// Then the Ingress Class is added back, and it is verified that the NPL annotation is added in the Service.
func TestNPLSvcIngressRemoveAddClass(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	SetUpTestForIngress(t, defaultL7Model)
	selectors := make(map[string]string)
	selectors["app"] = "npl"
	integrationtest.CreateServiceWithSelectors(t, defaultNS, "avisvc", corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false, selectors)
	createPodWithNPLAnnotation(selectors)

	found, _ := objects.SharedAviGraphLister().Get(defaultL7Model)
	if found {
		t.Fatalf("Couldn't find Model %v", defaultL7Model)
	}
	ingrFake := (integrationtest.FakeIngress{
		Name:        "foo-with-targets",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		ServiceName: "avisvc",
		ClassName:   integrationtest.DefaultIngressClass,
	}).Ingress()

	_, err := KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	g.Eventually(func() bool {
		svc, _ := KubeClient.CoreV1().Services(defaultNS).Get(context.TODO(), "avisvc", metav1.GetOptions{})
		ann := svc.GetAnnotations()
		if _, ok := ann[lib.NPLSvcAnnotation]; ok {
			return true
		}
		return false
	}, 40*time.Second).Should(gomega.Equal(true))

	integrationtest.RemoveDefaultIngressClass()
	g.Eventually(func() bool {
		svc, _ := KubeClient.CoreV1().Services(defaultNS).Get(context.TODO(), "avisvc", metav1.GetOptions{})
		ann := svc.GetAnnotations()
		if _, ok := ann[lib.NPLSvcAnnotation]; ok {
			return true
		}
		return false
	}, 20*time.Second).Should(gomega.Equal(false))

	integrationtest.AddDefaultIngressClass()
	g.Eventually(func() bool {
		svc, _ := KubeClient.CoreV1().Services(defaultNS).Get(context.TODO(), "avisvc", metav1.GetOptions{})
		ann := svc.GetAnnotations()
		if _, ok := ann[lib.NPLSvcAnnotation]; ok {
			return true
		}
		return false
	}, 20*time.Second).Should(gomega.Equal(true))

	KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), "foo-with-targets", metav1.DeleteOptions{})
	TearDownTestForIngress(t, defaultL7Model)
}

// TestNPLLBSvcNoLabel creates a Service of type LB with no Label and a Pod with NPL annotation.
// Then it is verified that no server is getting added in the model.
func TestNPLAutoAnnotationLBSvc(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	labels := make(map[string]string)
	labels["app"] = "npl"
	createPodWithNPLAnnotation(labels)

	objects.SharedAviGraphLister().Delete(integrationtest.SINGLEPORTMODEL)
	selectors := make(map[string]string)
	integrationtest.CreateServiceWithSelectors(t, defaultNS, integrationtest.SINGLEPORTSVC, corev1.ProtocolTCP, corev1.ServiceTypeLoadBalancer, false, selectors)
	g.Eventually(func() bool {
		svc, _ := KubeClient.CoreV1().Services(defaultNS).Get(context.TODO(), integrationtest.SINGLEPORTSVC, metav1.GetOptions{})
		ann := svc.GetAnnotations()
		if _, ok := ann[lib.NPLSvcAnnotation]; ok {
			return true
		}
		return false
	}, 40*time.Second).Should(gomega.Equal(true))

	tearDownTestForSvcLB(t, g)
}

// TestNPLSvcNodePort creates a Service and an Ingress which uses that Service.
// Then the service type is changed to NodePort, and it is verified that the NPL annotation is removed from the Service.
func TestNPLSvcNodePort(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	SetUpTestForIngress(t, defaultL7Model)
	defer func() {
		tearDownTestForSvcLB(t, g)
		TearDownTestForIngress(t, defaultL7Model)
	}()
	selectors := make(map[string]string)
	selectors["app"] = "npl"
	integrationtest.CreateServiceWithSelectors(t, defaultNS, "avisvc", corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false, selectors)
	createPodWithNPLAnnotation(selectors)

	found, _ := objects.SharedAviGraphLister().Get(defaultL7Model)
	if found {
		t.Fatalf("Couldn't find Model %v", defaultL7Model)
	}
	ingrFake := (integrationtest.FakeIngress{
		Name:        "foo-with-targets",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		ServiceName: "avisvc",
	}).Ingress()

	_, err := KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	g.Eventually(func() bool {
		svc, _ := KubeClient.CoreV1().Services(defaultNS).Get(context.TODO(), "avisvc", metav1.GetOptions{})
		ann := svc.GetAnnotations()
		if _, ok := ann[lib.NPLSvcAnnotation]; ok {
			return true
		}
		return false
	}, 20*time.Second).Should(gomega.Equal(true))

	svc := integrationtest.ConstructService(defaultNS, "avisvc", corev1.ProtocolTCP, corev1.ServiceTypeNodePort, false, selectors, "")
	ann := make(map[string]string)
	ann[lib.NPLSvcAnnotation] = "true"
	svc.Annotations = ann
	svc.ResourceVersion = "2"
	_, err = KubeClient.CoreV1().Services(defaultNS).Update(context.TODO(), svc, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating Service: %v", err)
	}
	g.Eventually(func() bool {
		svc, _ := KubeClient.CoreV1().Services(defaultNS).Get(context.TODO(), "avisvc", metav1.GetOptions{})
		ann := svc.GetAnnotations()
		if _, ok := ann[lib.NPLSvcAnnotation]; ok {
			return true
		}
		return false
	}, 20*time.Second).Should(gomega.Equal(false))

	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), "foo-with-targets", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
}

// TestIngressAddPodWithMultiportSvc creates a Pod with multiple nodeportlocal.antrea.io annotations, Service with multiport and
// an Ingress which uses that Service. Port number is mentioned instead of port name as backend servicePort.
func TestIngressAddPodWithMultiportSvc(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	SetUpTestForIngress(t, defaultL7Model)
	defer func() {
		tearDownTestForSvcLB(t, g)
		TearDownTestForIngress(t, defaultL7Model)
	}()
	selectors := make(map[string]string)
	selectors["app"] = "npl"
	integrationtest.CreateServiceWithSelectors(t, defaultNS, "avisvc", corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, true, selectors)
	createPodWithMultipleNPLAnnotations(selectors)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(defaultL7Model)
		return found
	}, 20*time.Second, 1*time.Second).Should(gomega.Equal(false))

	ingrFake := (integrationtest.FakeIngress{
		Name:        "foo-with-targets",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		ServiceName: "avisvc",
	}).IngressMultiPort()

	_, err := KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(defaultL7Model)
		return found
	}, 20*time.Second, 1*time.Second).Should(gomega.Equal(true))

	_, aviModel := objects.SharedAviGraphLister().Get(defaultL7Model)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(len(nodes)).To(gomega.Equal(1))
	g.Expect(nodes[0].PoolRefs).To(gomega.HaveLen(2))
	g.Expect(nodes[0].PoolRefs[0].Servers).To(gomega.HaveLen(1))
	g.Expect(nodes[0].PoolRefs[1].Servers).To(gomega.HaveLen(1))
	g.Expect(*nodes[0].PoolRefs[0].Servers[0].Ip.Addr).To(gomega.Equal(defaultHostIP))
	g.Expect(nodes[0].PoolRefs[0].Servers[0].Port).To(gomega.Equal(int32(61000)))
	g.Expect(*nodes[0].PoolRefs[1].Servers[0].Ip.Addr).To(gomega.Equal(defaultHostIP))
	g.Expect(nodes[0].PoolRefs[1].Servers[0].Port).To(gomega.Equal(int32(61001)))

	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), "foo-with-targets", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	verifyIngressDeletion(t, g, aviModel, 0)
}

// TestIngressPodReadiness creates a POD in not ready state with NPL annotation and corresponding Service and Ingress, then verifies the model.
// It also updates the Pod to ready state and verifies the model.
func TestIngressPodReadiness(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	SetUpTestForIngress(t, defaultL7Model)
	selectors := make(map[string]string)
	selectors["app"] = "npl"
	integrationtest.CreateServiceWithSelectors(t, defaultNS, "avisvc", corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false, selectors)
	// creating pod in not ready state
	createNotReadyPodWithNPLAnnotation(selectors)

	integrationtest.PollForCompletion(t, defaultL7Model, 10)
	found, _ := objects.SharedAviGraphLister().Get(defaultL7Model)
	if found {
		t.Fatalf("Model %v exists even after deletion", defaultL7Model)
	}
	ingrFake := (integrationtest.FakeIngress{
		Name:        "foo-with-targets",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		ServiceName: "avisvc",
	}).Ingress()

	_, err := KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	integrationtest.PollForCompletion(t, defaultL7Model, 10)
	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(defaultL7Model)
		return found
	}, 40*time.Second).Should(gomega.Equal(true))

	_, aviModel := objects.SharedAviGraphLister().Get(defaultL7Model)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(len(nodes)).To(gomega.Equal(1))
	g.Expect(nodes[0].PoolRefs).To(gomega.HaveLen(1))

	// verifying the number of pool servers to be zero
	g.Eventually(func() int {
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return len(nodes[0].PoolRefs[0].Servers)
	}, 40*time.Second).Should(gomega.Equal(0))

	// updating the pod to ready state
	updatePodWithNPLAnnotation(selectors)
	time.Sleep(5 * time.Second)
	_, aviModel = objects.SharedAviGraphLister().Get(defaultL7Model)
	g.Eventually(func() int {
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return len(nodes[0].PoolRefs[0].Servers)
	}, 40*time.Second).Should(gomega.Equal(1))
	g.Expect(*nodes[0].PoolRefs[0].Servers[0].Ip.Addr).To(gomega.Equal(defaultHostIP))
	g.Expect(nodes[0].PoolRefs[0].Servers[0].Port).To(gomega.Equal(int32(defaultNodePort)))

	// updating the pod to not ready state again
	updateNotReadyPodWithNPLAnnotation(selectors)
	time.Sleep(5 * time.Second)
	_, aviModel = objects.SharedAviGraphLister().Get(defaultL7Model)
	g.Eventually(func() int {
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return len(nodes[0].PoolRefs[0].Servers)
	}, 40*time.Second).Should(gomega.Equal(0))

	// re-updating the pod to ready state
	updatePodWithNPLAnnotation(selectors)
	time.Sleep(5 * time.Second)
	_, aviModel = objects.SharedAviGraphLister().Get(defaultL7Model)
	g.Eventually(func() int {
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return len(nodes[0].PoolRefs[0].Servers)
	}, 40*time.Second).Should(gomega.Equal(1))
	g.Expect(*nodes[0].PoolRefs[0].Servers[0].Ip.Addr).To(gomega.Equal(defaultHostIP))
	g.Expect(nodes[0].PoolRefs[0].Servers[0].Port).To(gomega.Equal(int32(defaultNodePort)))

	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), "foo-with-targets", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	verifyIngressDeletion(t, g, aviModel, 0)
	TearDownTestForIngress(t, defaultL7Model)
}

// TestNPLLBSvcPodReadiness creates a Service type LB and a not ready Pod with matching label, then the model is verified.
// It also updates the Pod to ready state and verifies the model.
func TestNPLLBSvcPodReadiness(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	selectors := make(map[string]string)
	selectors["app"] = "npl"
	createNotReadyPodWithNPLAnnotation(selectors)
	setUpTestForSvcLB(t)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(defaultLBModel)
		return found
	}, 40*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(defaultLBModel)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].Name).To(gomega.Equal(fmt.Sprintf("cluster--%s-%s", defaultNS, integrationtest.SINGLEPORTSVC)))
	g.Expect(nodes[0].Tenant).To(gomega.Equal(integrationtest.AVINAMESPACE))
	g.Expect(nodes[0].PortProto[0].Port).To(gomega.Equal(int32(8080)))

	// verifying the number of pool servers to be zero
	g.Eventually(func() int {
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(nodes[0].PoolRefs).To(gomega.HaveLen(1))
		return len(nodes[0].PoolRefs[0].Servers)
	}, 40*time.Second).Should(gomega.Equal(0))

	// updating the pod to ready state
	updatePodWithNPLAnnotation(selectors)
	time.Sleep(5 * time.Second)
	_, aviModel = objects.SharedAviGraphLister().Get(defaultLBModel)
	g.Eventually(func() int {
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(nodes[0].PoolRefs).To(gomega.HaveLen(1))
		return len(nodes[0].PoolRefs[0].Servers)
	}, 40*time.Second).Should(gomega.Equal(1))
	address := defaultHostIP
	g.Expect(nodes[0].PoolRefs[0].Servers[0].Ip.Addr).To(gomega.Equal(&address))

	// updating the pod to not ready state again
	updateNotReadyPodWithNPLAnnotation(selectors)
	time.Sleep(5 * time.Second)
	_, aviModel = objects.SharedAviGraphLister().Get(defaultLBModel)
	g.Eventually(func() int {
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(nodes[0].PoolRefs).To(gomega.HaveLen(1))
		return len(nodes[0].PoolRefs[0].Servers)
	}, 40*time.Second).Should(gomega.Equal(0))

	// re-updating the pod to ready state
	updatePodWithNPLAnnotation(selectors)
	time.Sleep(5 * time.Second)
	_, aviModel = objects.SharedAviGraphLister().Get(defaultLBModel)
	g.Eventually(func() int {
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(nodes[0].PoolRefs).To(gomega.HaveLen(1))
		return len(nodes[0].PoolRefs[0].Servers)
	}, 40*time.Second).Should(gomega.Equal(1))
	g.Expect(nodes[0].PoolRefs[0].Servers[0].Ip.Addr).To(gomega.Equal(&address))

	tearDownTestForSvcLB(t, g)
}
