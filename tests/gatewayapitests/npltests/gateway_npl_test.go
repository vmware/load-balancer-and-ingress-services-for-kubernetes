/*
 * Copyright 2023-2024 VMware, Inc.
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
	"os"
	"sync"
	"testing"
	"time"

	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	gatewayfake "sigs.k8s.io/gateway-api/pkg/client/clientset/versioned/fake"

	akogatewayapik8s "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/k8s"
	akogatewayapilib "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/k8s"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	avinodes "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/nodes"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/status"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
	tests "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/tests/gatewayapitests"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/tests/integrationtest"
)

var ctrl *akogatewayapik8s.GatewayController

const (
	DEFAULT_NAMESPACE = "default"
)

func TestMain(m *testing.M) {

	tests.KubeClient = k8sfake.NewSimpleClientset()
	tests.GatewayClient = gatewayfake.NewSimpleClientset()
	integrationtest.KubeClient = tests.KubeClient

	// Sets the environment variables
	os.Setenv("CLUSTER_NAME", "cluster")
	os.Setenv("CLOUD_NAME", "CLOUD_VCENTER")
	os.Setenv("SEG_NAME", "Default-Group")
	os.Setenv("POD_NAMESPACE", utils.AKO_DEFAULT_NS)
	os.Setenv("FULL_SYNC_INTERVAL", utils.AKO_DEFAULT_NS)
	os.Setenv("ENABLE_EVH", "true")
	os.Setenv("TENANT", "admin")
	os.Setenv("POD_NAME", "ako-0")
	os.Setenv("SERVICE_TYPE", "NodePortLocal")

	// Set the user with prefix
	_ = lib.AKOControlConfig()
	lib.SetAKOUser(akogatewayapilib.Prefix)
	lib.SetNamePrefix(akogatewayapilib.Prefix)
	lib.AKOControlConfig().SetIsLeaderFlag(true)
	akoControlConfig := akogatewayapilib.AKOControlConfig()
	akoControlConfig.SetEventRecorder(lib.AKOGatewayEventComponent, tests.KubeClient, true)
	registeredInformers := []string{
		utils.ServiceInformer,
		utils.SecretInformer,
		utils.NSInformer,
		utils.PodInformer,
	}
	utils.NewInformers(utils.KubeClientIntf{ClientSet: tests.KubeClient}, registeredInformers, make(map[string]interface{}))
	data := map[string][]byte{
		"username": []byte("admin"),
		"password": []byte("admin"),
	}
	object := metav1.ObjectMeta{Name: "avi-secret", Namespace: utils.GetAKONamespace()}
	secret := &corev1.Secret{Data: data, ObjectMeta: object}
	tests.KubeClient.CoreV1().Secrets(utils.GetAKONamespace()).Create(context.TODO(), secret, metav1.CreateOptions{})

	akoApi := integrationtest.InitializeFakeAKOAPIServer()
	defer akoApi.ShutDown()

	tests.NewAviFakeClientInstance(tests.KubeClient)
	defer integrationtest.AviFakeClientInstance.Close()

	ctrl = akogatewayapik8s.SharedGatewayController()
	ctrl.DisableSync = false
	ctrl.InitGatewayAPIInformers(tests.GatewayClient)
	akoControlConfig.SetGatewayAPIClientset(tests.GatewayClient)

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

	integrationtest.AddConfigMap(tests.KubeClient)
	go ctrl.InitController(k8s.K8sinformers{Cs: tests.KubeClient}, registeredInformers, ctrlCh, stopCh, quickSyncCh, waitGroupMap)
	os.Exit(m.Run())

}

const (
	defaultPodName  = "test-pod"
	defaultNodeName = "test-node"
	defaultPodPort  = 80
	defaultHostIP   = "10.10.10.10"
	defaultPodIP    = "192.168.32.10"
	defaultNodePort = 60001
)

func createPodWithNPLAnnotation(labels map[string]string) {
	testPod := getTestPod(labels)
	ann := make(map[string]string)
	ann[lib.NPLPodAnnotation] = "[{\"podPort\":8080,\"nodeIP\":\"10.10.10.10\",\"nodePort\":60001}]"
	testPod.Annotations = ann
	tests.KubeClient.CoreV1().Pods(DEFAULT_NAMESPACE).Create(context.TODO(), &testPod, metav1.CreateOptions{})
}

func updatePodWithNPLAnnotation(labels map[string]string) {
	testPod := getTestPod(labels)
	ann := make(map[string]string)
	ann[lib.NPLPodAnnotation] = "[{\"podPort\":8080,\"nodeIP\":\"10.10.10.10\",\"nodePort\":60001}]"
	testPod.Annotations = ann
	testPod.ResourceVersion = "2"
	tests.KubeClient.CoreV1().Pods(DEFAULT_NAMESPACE).Update(context.TODO(), &testPod, metav1.UpdateOptions{})
}
func getTestPod(labels map[string]string) corev1.Pod {
	testPod := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      defaultPodName,
			Namespace: DEFAULT_NAMESPACE,
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

func setupAndVerifyGatewayForNPL(t *testing.T, g *gomega.WithT, gatewayClassName, gatewayName, httpRouteName, modelName string, ports []int32) {
	tests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	listeners := tests.GetListenersV1(ports, false, false)
	tests.SetupGateway(t, gatewayName, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 25*time.Second).Should(gomega.Equal(true))

	parentRefs := tests.GetParentReferencesV1([]string{gatewayName}, DEFAULT_NAMESPACE, ports)
	rule := tests.GetHTTPRouteRuleV1([]string{"/foo"}, []string{},
		map[string][]string{"RequestHeaderModifier": {"add"}},
		[][]string{{"avisvc", "default", "8080", "1"}}, nil)
	rules := []gatewayv1.HTTPRouteRule{rule}
	hostnames := []gatewayv1.Hostname{"foo-8080.com"}
	tests.SetupHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE, parentRefs, hostnames, rules)
	g.Eventually(func() int {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)

		if !found {
			return 0
		}
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return len(nodes[0].EvhNodes)
	}, 25*time.Second).Should(gomega.Equal(1))
}

func cleanupGatewayForNPL(t *testing.T, gatewayClassName, gatewayName, httpRouteName string) {
	tests.KubeClient.CoreV1().Pods(DEFAULT_NAMESPACE).Delete(context.TODO(), defaultPodName, metav1.DeleteOptions{})
	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, "avisvc")
	tests.TeardownHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE)
	tests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	tests.TeardownGatewayClass(t, gatewayClassName)
}

func TestAddPod(t *testing.T) {
	gatewayName := "gateway-npl-01"
	gatewayClassName := "gateway-class-npl-01"
	httpRouteName := "http-route-npl-01"
	ports := []int32{8080}
	modelName, _ := tests.GetModelName(DEFAULT_NAMESPACE, gatewayName)

	g := gomega.NewGomegaWithT(t)

	selectors := make(map[string]string)
	selectors["app"] = "npl"
	integrationtest.CreateServiceWithSelectors(t, DEFAULT_NAMESPACE, "avisvc", corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false, selectors)
	createPodWithNPLAnnotation(selectors)

	setupAndVerifyGatewayForNPL(t, g, gatewayClassName, gatewayName, httpRouteName, modelName, ports)

	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	childNode := nodes[0].EvhNodes[0]
	g.Expect(childNode.PoolRefs).To(gomega.HaveLen(1))

	g.Eventually(func() int {
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		childNode = nodes[0].EvhNodes[0]
		return len(childNode.PoolRefs[0].Servers)
	}, 40*time.Second).Should(gomega.Equal(1))
	g.Expect(childNode.PoolRefs[0].Servers).To(gomega.HaveLen(1))
	g.Expect(*childNode.PoolRefs[0].Servers[0].Ip.Addr).To(gomega.Equal(defaultHostIP))
	g.Expect(childNode.PoolRefs[0].Servers[0].Port).To(gomega.Equal(int32(defaultNodePort)))

	cleanupGatewayForNPL(t, gatewayClassName, gatewayName, httpRouteName)
}

func TestDelPod(t *testing.T) {
	gatewayName := "gateway-npl-02"
	gatewayClassName := "gateway-class-npl-02"
	httpRouteName := "http-route-npl-02"
	ports := []int32{8080}
	modelName, _ := tests.GetModelName(DEFAULT_NAMESPACE, gatewayName)

	g := gomega.NewGomegaWithT(t)

	selectors := make(map[string]string)
	selectors["app"] = "npl"
	integrationtest.CreateServiceWithSelectors(t, DEFAULT_NAMESPACE, "avisvc", corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false, selectors)
	createPodWithNPLAnnotation(selectors)

	setupAndVerifyGatewayForNPL(t, g, gatewayClassName, gatewayName, httpRouteName, modelName, ports)

	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	childNode := nodes[0].EvhNodes[0]
	g.Expect(childNode.PoolRefs).To(gomega.HaveLen(1))

	g.Eventually(func() int {
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		childNode = nodes[0].EvhNodes[0]
		return len(childNode.PoolRefs[0].Servers)
	}, 40*time.Second).Should(gomega.Equal(1))
	g.Expect(childNode.PoolRefs[0].Servers).To(gomega.HaveLen(1))
	g.Expect(*childNode.PoolRefs[0].Servers[0].Ip.Addr).To(gomega.Equal(defaultHostIP))
	g.Expect(childNode.PoolRefs[0].Servers[0].Port).To(gomega.Equal(int32(defaultNodePort)))

	err := tests.KubeClient.CoreV1().Pods(DEFAULT_NAMESPACE).Delete(context.TODO(), defaultPodName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("error in deleting Pod: %v", err)
	}
	g.Eventually(func() int {
		_, aviModel = objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		childNode = nodes[0].EvhNodes[0]
		return len(childNode.PoolRefs[0].Servers)
	}, 40*time.Second).Should(gomega.Equal(0))

	cleanupGatewayForNPL(t, gatewayClassName, gatewayName, httpRouteName)
}

func TestAddPodWithoutLabel(t *testing.T) {
	gatewayName := "gateway-npl-03"
	gatewayClassName := "gateway-class-npl-03"
	httpRouteName := "http-route-npl-03"
	ports := []int32{8080}
	modelName, _ := tests.GetModelName(DEFAULT_NAMESPACE, gatewayName)

	g := gomega.NewGomegaWithT(t)

	selectors := make(map[string]string)
	integrationtest.CreateServiceWithSelectors(t, DEFAULT_NAMESPACE, "avisvc", corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false, selectors)
	createPodWithNPLAnnotation(selectors)

	setupAndVerifyGatewayForNPL(t, g, gatewayClassName, gatewayName, httpRouteName, modelName, ports)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 40*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(len(nodes)).To(gomega.Equal(1))
	childNode := nodes[0].EvhNodes[0]
	g.Expect(childNode.PoolRefs).To(gomega.HaveLen(1))
	g.Expect(childNode.PoolRefs[0].Servers).To(gomega.HaveLen(0))

	cleanupGatewayForNPL(t, gatewayClassName, gatewayName, httpRouteName)
}

func TestUpdatePodWithLabel(t *testing.T) {
	gatewayName := "gateway-npl-04"
	gatewayClassName := "gateway-class-npl-04"
	httpRouteName := "http-route-npl-04"
	ports := []int32{8080}
	modelName, _ := tests.GetModelName(DEFAULT_NAMESPACE, gatewayName)

	g := gomega.NewGomegaWithT(t)

	selectors := make(map[string]string)
	selectors["app"] = "npl"
	integrationtest.CreateServiceWithSelectors(t, DEFAULT_NAMESPACE, "avisvc", corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false, selectors)
	labels := make(map[string]string)
	createPodWithNPLAnnotation(labels)

	setupAndVerifyGatewayForNPL(t, g, gatewayClassName, gatewayName, httpRouteName, modelName, ports)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 40*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(len(nodes)).To(gomega.Equal(1))

	childNode := nodes[0].EvhNodes[0]
	g.Expect(childNode.PoolRefs).To(gomega.HaveLen(1))
	g.Expect(childNode.PoolRefs[0].Servers).To(gomega.HaveLen(0))

	labels["app"] = "npl"
	updatePodWithNPLAnnotation(labels)
	g.Eventually(func() int {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		childNode = nodes[0].EvhNodes[0]
		return len(childNode.PoolRefs[0].Servers)
	}, 40*time.Second).Should(gomega.Equal(1))
	g.Expect(childNode.PoolRefs[0].Servers).To(gomega.HaveLen(1))
	g.Expect(*childNode.PoolRefs[0].Servers[0].Ip.Addr).To(gomega.Equal(defaultHostIP))
	g.Expect(childNode.PoolRefs[0].Servers[0].Port).To(gomega.Equal(int32(defaultNodePort)))

	cleanupGatewayForNPL(t, gatewayClassName, gatewayName, httpRouteName)
}

func TestUpdatePodWithoutLabel(t *testing.T) {
	gatewayName := "gateway-npl-05"
	gatewayClassName := "gateway-class-npl-05"
	httpRouteName := "http-route-npl-05"
	ports := []int32{8080}
	modelName, _ := tests.GetModelName(DEFAULT_NAMESPACE, gatewayName)

	g := gomega.NewGomegaWithT(t)

	selectors := make(map[string]string)
	selectors["app"] = "npl"
	integrationtest.CreateServiceWithSelectors(t, DEFAULT_NAMESPACE, "avisvc", corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false, selectors)
	createPodWithNPLAnnotation(selectors)

	setupAndVerifyGatewayForNPL(t, g, gatewayClassName, gatewayName, httpRouteName, modelName, ports)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 40*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(len(nodes)).To(gomega.Equal(1))
	childNode := nodes[0].EvhNodes[0]
	g.Expect(childNode.Tenant).To(gomega.Equal("admin"))
	g.Expect(childNode.PoolRefs).To(gomega.HaveLen(1))
	g.Eventually(func() int {
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		childNode = nodes[0].EvhNodes[0]
		return len(childNode.PoolRefs[0].Servers)
	}, 40*time.Second).Should(gomega.Equal(1))
	g.Expect(childNode.PoolRefs[0].Servers).To(gomega.HaveLen(1))
	g.Expect(*childNode.PoolRefs[0].Servers[0].Ip.Addr).To(gomega.Equal(defaultHostIP))
	g.Expect(childNode.PoolRefs[0].Servers[0].Port).To(gomega.Equal(int32(defaultNodePort)))

	labels := make(map[string]string)
	updatePodWithNPLAnnotation(labels)
	g.Eventually(func() int {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		childNode = nodes[0].EvhNodes[0]
		return len(childNode.PoolRefs[0].Servers)
	}, 40*time.Second).Should(gomega.Equal(0))

	cleanupGatewayForNPL(t, gatewayClassName, gatewayName, httpRouteName)
}

func TestDelSvc(t *testing.T) {
	gatewayName := "gateway-npl-06"
	gatewayClassName := "gateway-class-npl-06"
	httpRouteName := "http-route-npl-06"
	ports := []int32{8080}
	modelName, _ := tests.GetModelName(DEFAULT_NAMESPACE, gatewayName)

	g := gomega.NewGomegaWithT(t)

	selectors := make(map[string]string)
	selectors["app"] = "npl"
	integrationtest.CreateServiceWithSelectors(t, DEFAULT_NAMESPACE, "avisvc", corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false, selectors)
	createPodWithNPLAnnotation(selectors)

	setupAndVerifyGatewayForNPL(t, g, gatewayClassName, gatewayName, httpRouteName, modelName, ports)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 40*time.Second).Should(gomega.Equal(true))

	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(len(nodes)).To(gomega.Equal(1))
	childNode := nodes[0].EvhNodes[0]
	g.Expect(childNode.PoolRefs).To(gomega.HaveLen(1))

	g.Eventually(func() int {
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		childNode = nodes[0].EvhNodes[0]
		return len(childNode.PoolRefs[0].Servers)
	}, 40*time.Second).Should(gomega.Equal(1))
	g.Expect(childNode.PoolRefs[0].Servers).To(gomega.HaveLen(1))
	g.Expect(*childNode.PoolRefs[0].Servers[0].Ip.Addr).To(gomega.Equal(defaultHostIP))
	g.Expect(childNode.PoolRefs[0].Servers[0].Port).To(gomega.Equal(int32(defaultNodePort)))

	integrationtest.DelSVC(t, "default", "avisvc")
	g.Eventually(func() int {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		if len(nodes) > 0 {
			return len(nodes[0].EvhNodes[0].PoolRefs)
		}
		return -1
	}, 40*time.Second).Should(gomega.Equal(0))

	cleanupGatewayForNPL(t, gatewayClassName, gatewayName, httpRouteName)
}

func TestSvcAutoAnnotate(t *testing.T) {
	// create gateway
	// create http route 1
	// create service
	// check annotation present
	// delete route 1
	// check annotation not present
	// create route 1
	// check annotation present

	// create route 2
	// delete route 1
	// check annotation present

	// delete route 2
	// check annotation not present

	gatewayName := "gateway-npl-06"
	gatewayClassName := "gateway-class-npl-06"
	httpRouteName := "http-route-npl-06a"
	httpRouteName2 := "http-route-npl-06b"
	ports := []int32{8080}
	modelName, _ := tests.GetModelName(DEFAULT_NAMESPACE, gatewayName)

	g := gomega.NewGomegaWithT(t)

	tests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	listeners := tests.GetListenersV1(ports, false, false)

	tests.SetupGateway(t, gatewayName, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 25*time.Second).Should(gomega.Equal(true))

	parentRefs := tests.GetParentReferencesV1([]string{gatewayName}, DEFAULT_NAMESPACE, ports)
	rule := tests.GetHTTPRouteRuleV1([]string{"/foo"}, []string{},
		map[string][]string{"RequestHeaderModifier": {"add"}},
		[][]string{{"avisvc", "default", "8080", "1"}}, nil)
	rules := []gatewayv1.HTTPRouteRule{rule}
	hostnames := []gatewayv1.Hostname{"foo-8080.com"}
	tests.SetupHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE, parentRefs, hostnames, rules)
	g.Eventually(func() int {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)

		if !found {
			return 0
		}
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return len(nodes[0].EvhNodes)
	}, 25*time.Second).Should(gomega.Equal(1))

	selectors := make(map[string]string)
	selectors["app"] = "npl"
	integrationtest.CreateServiceWithSelectors(t, DEFAULT_NAMESPACE, "avisvc", corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false, selectors)
	g.Eventually(func() bool {
		if !status.CheckNPLSvcAnnotation(modelName, DEFAULT_NAMESPACE, "avisvc") {
			return false
		}
		return true
	}, 5*time.Second).Should(gomega.Equal(true))

	tests.TeardownHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE)

	g.Eventually(func() bool {
		if status.CheckNPLSvcAnnotation(modelName, DEFAULT_NAMESPACE, "avisvc") {
			return false
		}
		return true
	}, 5*time.Second).Should(gomega.Equal(true))

	tests.SetupHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE, parentRefs, hostnames, rules)

	g.Eventually(func() bool {
		if !status.CheckNPLSvcAnnotation(modelName, DEFAULT_NAMESPACE, "avisvc") {
			return false
		}
		return true
	}, 5*time.Second).Should(gomega.Equal(true))

	// create route 2
	tests.SetupHTTPRoute(t, httpRouteName2, DEFAULT_NAMESPACE, parentRefs, hostnames, rules)
	g.Eventually(func() int {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)

		if !found {
			return 0
		}
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return len(nodes[0].EvhNodes)
	}, 25*time.Second).Should(gomega.Equal(1))

	tests.TeardownHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE)
	g.Eventually(func() bool {
		if !status.CheckNPLSvcAnnotation(modelName, DEFAULT_NAMESPACE, "avisvc") {
			return false
		}
		return true
	}, 5*time.Second).Should(gomega.Equal(true))

	tests.TeardownHTTPRoute(t, httpRouteName2, DEFAULT_NAMESPACE)

	g.Eventually(func() bool {
		if status.CheckNPLSvcAnnotation(modelName, DEFAULT_NAMESPACE, "avisvc") {
			return false
		}
		return true
	}, 5*time.Second).Should(gomega.Equal(true))

	integrationtest.DelSVC(t, "default", "avisvc")

	tests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	tests.TeardownGatewayClass(t, gatewayClassName)
}

func TestSvcUpdateAutoAnnotate(t *testing.T) {
	gatewayName := "gateway-npl-06"
	gatewayClassName := "gateway-class-npl-06"
	httpRouteName := "http-route-npl-06a"
	ports := []int32{8080}
	modelName, _ := tests.GetModelName(DEFAULT_NAMESPACE, gatewayName)

	g := gomega.NewGomegaWithT(t)

	tests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	listeners := tests.GetListenersV1(ports, false, false)

	tests.SetupGateway(t, gatewayName, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 25*time.Second).Should(gomega.Equal(true))

	parentRefs := tests.GetParentReferencesV1([]string{gatewayName}, DEFAULT_NAMESPACE, ports)
	rule := tests.GetHTTPRouteRuleV1([]string{"/foo"}, []string{},
		map[string][]string{"RequestHeaderModifier": {"add"}},
		[][]string{{"avisvc", "default", "8080", "1"}}, nil)
	rules := []gatewayv1.HTTPRouteRule{rule}
	hostnames := []gatewayv1.Hostname{"foo-8080.com"}
	tests.SetupHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE, parentRefs, hostnames, rules)
	g.Eventually(func() int {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)

		if !found {
			return 0
		}
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return len(nodes[0].EvhNodes)
	}, 25*time.Second).Should(gomega.Equal(1))

	selectors := make(map[string]string)
	selectors["app"] = "npl"
	integrationtest.CreateServiceWithSelectors(t, DEFAULT_NAMESPACE, "avisvc", corev1.ProtocolTCP, corev1.ServiceTypeNodePort, false, selectors)
	g.Eventually(func() bool {
		if status.CheckNPLSvcAnnotation(modelName, DEFAULT_NAMESPACE, "avisvc") {
			return false
		}
		return true
	}, 5*time.Second).Should(gomega.Equal(true))

	integrationtest.UpdateSVC(t, DEFAULT_NAMESPACE, "avisvc", corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false)
	g.Eventually(func() bool {
		if !status.CheckNPLSvcAnnotation(modelName, DEFAULT_NAMESPACE, "avisvc") {
			return false
		}
		return true
	}, 5*time.Second).Should(gomega.Equal(true))

	cleanupGatewayForNPL(t, gatewayClassName, gatewayName, httpRouteName)
}
