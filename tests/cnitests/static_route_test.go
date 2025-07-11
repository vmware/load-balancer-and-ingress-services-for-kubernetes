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

package cnitests

import (
	"context"
	"flag"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/k8s"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	avinodes "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/nodes"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"
	crdfake "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/client/v1alpha1/clientset/versioned/fake"
	v1beta1crdfake "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/client/v1beta1/clientset/versioned/fake"

	v1alpha2crdfake "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/client/v1alpha2/clientset/versioned/fake"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/tests/integrationtest"

	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	k8sfake "k8s.io/client-go/kubernetes/fake"
)

var KubeClient *k8sfake.Clientset
var CRDClient *crdfake.Clientset
var v1alpha2CRDClient *v1alpha2crdfake.Clientset
var v1beta1CRDClient *v1beta1crdfake.Clientset
var ctrl *k8s.AviController
var DynamicClient *dynamicfake.FakeDynamicClient

var cniPlugin = flag.String("cniPlugin", "", "cni plugin for the setup")

//var calico = "calico"
//var cniPlugin = &calico

func TestMain(m *testing.M) {
	flag.Parse()
	os.Setenv("VIP_NETWORK_LIST", `[{"networkName":"net123"}]`)
	os.Setenv("CLUSTER_NAME", "cluster")
	os.Setenv("CLOUD_NAME", "CLOUD_VCENTER")
	os.Setenv("SEG_NAME", "Default-Group")
	os.Setenv("NODE_NETWORK_LIST", `[{"networkName":"net123","cidrs":["10.79.168.0/22"]}]`)
	os.Setenv("SERVICE_TYPE", "ClusterIP")
	os.Setenv("AUTO_L4_FQDN", "disable")
	os.Setenv("POD_NAMESPACE", utils.AKO_DEFAULT_NS)
	os.Setenv("SHARD_VS_SIZE", "LARGE")
	if *cniPlugin != "" {
		os.Setenv("CNI_PLUGIN", *cniPlugin)
	}
	os.Setenv("POD_NAME", "ako-0")

	akoControlConfig := lib.AKOControlConfig()
	KubeClient = k8sfake.NewSimpleClientset()
	CRDClient = crdfake.NewSimpleClientset()
	v1beta1CRDClient = v1beta1crdfake.NewSimpleClientset()
	v1alpha2CRDClient = v1alpha2crdfake.NewSimpleClientset()

	gvrToKind := make(map[schema.GroupVersionResource]string)
	var testData unstructured.Unstructured
	if *cniPlugin == "calico" {
		testData.SetUnstructuredContent(map[string]interface{}{
			"apiVersion": "crd.projectcalico.org/v1",
			"kind":       "blockaffinities",
			"metadata": map[string]interface{}{
				"name":      "testblockaffinity",
				"namespace": "default",
			},
			"spec": map[string]interface{}{
				"cidr":    "10.133.0.0/26",
				"deleted": "false",
				"node":    "testNodeCalico",
				"state":   "confirmed",
			},
		})
		gvrToKind[lib.CalicoBlockaffinityGVR] = "blockaffinitiesList"
	} else if *cniPlugin == "cilium" {
		specAddress1 := map[string]interface{}{
			"ip":   "10.1.1.2",
			"type": "InternalIP",
		}
		specAddress2 := map[string]interface{}{
			"ip":   "10.128.0.33",
			"type": "CiliumInternalIP",
		}
		specAddressList := []interface{}{
			specAddress1,
			specAddress2,
		}
		testData.SetUnstructuredContent(map[string]interface{}{
			"apiVersion": "cilium.io/v2",
			"kind":       "ciliumnodes",
			"metadata": map[string]interface{}{
				"name":      "testNodeCilium",
				"namespace": "default",
			},
			"spec": map[string]interface{}{
				"addresses": specAddressList,
				"ipam": map[string]interface{}{
					"podCIDRs": []interface{}{"10.128.0.0/23"},
				},
			},
		})
		gvrToKind[lib.CiliumNodeGVR] = "ciliumnodesList"
	}

	DynamicClient = dynamicfake.NewSimpleDynamicClientWithCustomListKinds(runtime.NewScheme(), gvrToKind, &testData)
	//DynamicClient = dynamicfake.NewSimpleDynamicClient(runtime.NewScheme())
	lib.SetDynamicClientSet(DynamicClient)
	akoControlConfig.SetCRDClientset(CRDClient)
	akoControlConfig.Setv1alpha2CRDClientset(v1alpha2CRDClient)
	akoControlConfig.Setv1beta1CRDClientset(v1beta1CRDClient)
	akoControlConfig.SetAKOInstanceFlag(true)
	akoControlConfig.SetEventRecorder(lib.AKOEventComponent, KubeClient, true)
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
	informers := k8s.K8sinformers{Cs: KubeClient, DynamicClient: DynamicClient}
	k8s.NewCRDInformers()

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

	go ctrl.InitController(informers, registeredInformers, ctrlCh, stopCh, quickSyncCh, waitGroupMap)
	os.Exit(m.Run())
}

func TestBlockAffinity(t *testing.T) {
	if *cniPlugin != "calico" {
		t.Skip("Skipping BlockAffinity test since CNI plugin is not Calico")
	}
	g := gomega.NewGomegaWithT(t)
	modelName := "admin/global"
	nodeip := "10.1.1.2"
	nodeName := "testNodeCalico"
	objects.SharedAviGraphLister().Delete(modelName)

	// mimicking actual scenario where the node will have atleast one BlockAffinity object created from start
	var testData unstructured.Unstructured
	testData.SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "crd.projectcalico.org/v1",
		"kind":       "blockaffinities",
		"metadata": map[string]interface{}{
			"name":      "testblockaffinity",
			"namespace": "default",
		},
		"spec": map[string]interface{}{
			"cidr":    "10.133.0.0/26",
			"deleted": "false",
			"node":    nodeName,
			"state":   "confirmed",
		},
	})

	DynamicClient.Resource(lib.CalicoBlockaffinityGVR).Namespace("default").Create(context.TODO(), &testData, v1.CreateOptions{})

	nodeExample := (integrationtest.FakeNode{
		Name:    nodeName,
		Version: "1",
		NodeIP:  nodeip,
	}).NodeCalico()

	_, err := KubeClient.CoreV1().Nodes().Create(context.TODO(), nodeExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Node: %v", err)
	}

	integrationtest.PollForCompletion(t, modelName, 5)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 10*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	g.Expect(aviModel.(*avinodes.AviObjectGraph).IsVrf).To(gomega.Equal(true))
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVRF()
	g.Expect(len(nodes)).To(gomega.Equal(1))

	g.Expect(len(nodes[0].StaticRoutes)).To(gomega.Equal(1))
	g.Expect(*(nodes[0].StaticRoutes[0].NextHop.Addr)).To(gomega.Equal(nodeip))
	g.Expect(*(nodes[0].StaticRoutes[0].Prefix.IPAddr.Addr)).To(gomega.Equal("10.133.0.0"))
	g.Expect(*(nodes[0].StaticRoutes[0].Prefix.Mask)).To(gomega.Equal(int32(26)))

	// creating a new BlockAffinity object for the node
	var testData2 unstructured.Unstructured
	testData2.SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "crd.projectcalico.org/v1",
		"kind":       "blockaffinities",
		"metadata": map[string]interface{}{
			"name":      "testblockaffinity2",
			"namespace": "default",
		},
		"spec": map[string]interface{}{
			"cidr":    "10.134.0.0/26",
			"deleted": "false",
			"node":    nodeName,
			"state":   "confirmed",
		},
	})

	DynamicClient.Resource(lib.CalicoBlockaffinityGVR).Namespace("default").Create(context.TODO(), &testData2, v1.CreateOptions{})
	// waiting for routes to get populated after BlockAffinity object creation
	time.Sleep(5 * time.Second)

	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	g.Expect(aviModel.(*avinodes.AviObjectGraph).IsVrf).To(gomega.Equal(true))
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVRF()
	g.Expect(len(nodes)).To(gomega.Equal(1))

	g.Expect(len(nodes[0].StaticRoutes)).To(gomega.Equal(2))
	g.Expect(*(nodes[0].StaticRoutes[1].NextHop.Addr)).To(gomega.Equal(nodeip))
	g.Expect(*(nodes[0].StaticRoutes[1].Prefix.IPAddr.Addr)).To(gomega.Equal("10.134.0.0"))
	g.Expect(*(nodes[0].StaticRoutes[1].Prefix.Mask)).To(gomega.Equal(int32(26)))

	time.Sleep(5 * time.Second)
	// deleting the second BlockAffinity object for the node
	DynamicClient.Resource(lib.CalicoBlockaffinityGVR).Namespace("default").Delete(context.TODO(), "testblockaffinity2", v1.DeleteOptions{})
	g.Eventually(func() int {
		_, aviModel = objects.SharedAviGraphLister().Get(modelName)
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVRF()
		num_routes := len(nodes[0].StaticRoutes)
		return num_routes
	}, 10*time.Second).Should(gomega.Equal(1))

	err = KubeClient.CoreV1().Nodes().Delete(context.TODO(), nodeExample.Name, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("error in deleting Node: %v", err)
	}
	g.Eventually(func() int {
		_, aviModel = objects.SharedAviGraphLister().Get(modelName)
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVRF()
		num_routes := len(nodes[0].StaticRoutes)
		return num_routes

	}, 10*time.Second).Should(gomega.Equal(0))
}

func TestNodeAnnotationUpdate(t *testing.T) {
	if *cniPlugin != "calico" {
		t.Skip("Skipping BlockAffinity test since CNI plugin is not Calico")
	}

	g := gomega.NewGomegaWithT(t)

	modelName := "admin/global"
	nodeip := "10.1.1.3"
	nodeName := "testNodeCalicoWithAnnotation"
	objects.SharedAviGraphLister().Delete(modelName)

	// mimicking actual scenario where the node will have atleast one BlockAffinity object created from start
	var testData unstructured.Unstructured
	testData.SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "crd.projectcalico.org/v1",
		"kind":       "blockaffinities",
		"metadata": map[string]interface{}{
			"name":      "testblockaffinityWithAnnotation",
			"namespace": "default",
		},
		"spec": map[string]interface{}{
			"cidr":    "10.135.0.0/26",
			"deleted": "false",
			"node":    nodeName,
			"state":   "confirmed",
		},
	})

	DynamicClient.Resource(lib.CalicoBlockaffinityGVR).Namespace("default").Create(context.TODO(), &testData, v1.CreateOptions{})

	nodeExample := (integrationtest.FakeNode{
		Name:    nodeName,
		Version: "1",
		NodeIP:  nodeip,
	}).NodeCalico()

	nodeV4IP := "10.0.0.1"
	nodeV4Mask := "24"
	nodeExample.ObjectMeta.Annotations = map[string]string{}
	nodeExample.ObjectMeta.Annotations[lib.CalicoIPv4AddressAnnotation] = nodeV4IP + "/" + nodeV4Mask

	newNode, err := KubeClient.CoreV1().Nodes().Create(context.TODO(), nodeExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Node: %v", err)
	}

	integrationtest.PollForCompletion(t, modelName, 5)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 10*time.Second).Should(gomega.Equal(true))

	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	g.Expect(aviModel.(*avinodes.AviObjectGraph).IsVrf).To(gomega.Equal(true))
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVRF()
	g.Expect(len(nodes)).To(gomega.Equal(1))

	g.Expect(len(nodes[0].StaticRoutes)).To(gomega.Equal(1))
	g.Expect(*(nodes[0].StaticRoutes[0].NextHop.Addr)).To(gomega.Equal(nodeV4IP))
	g.Expect(*(nodes[0].StaticRoutes[0].Prefix.IPAddr.Addr)).To(gomega.Equal("10.135.0.0"))
	g.Expect(*(nodes[0].StaticRoutes[0].Prefix.Mask)).To(gomega.Equal(int32(26)))

	nodeV4IP = "10.0.0.2"
	newNode.ObjectMeta.Annotations[lib.CalicoIPv4AddressAnnotation] = nodeV4IP + "/" + nodeV4Mask
	newNode, err = KubeClient.CoreV1().Nodes().Update(context.TODO(), newNode, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating Node: %v", err)
	}
	// waiting for routes to get populated after BlockAffinity object creation
	time.Sleep(30 * time.Second)

	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	g.Expect(aviModel.(*avinodes.AviObjectGraph).IsVrf).To(gomega.Equal(true))
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVRF()
	g.Expect(len(nodes)).To(gomega.Equal(1))

	g.Expect(len(nodes[0].StaticRoutes)).To(gomega.Equal(1))
	g.Expect(*(nodes[0].StaticRoutes[0].NextHop.Addr)).To(gomega.Equal(nodeV4IP))
	g.Expect(*(nodes[0].StaticRoutes[0].Prefix.IPAddr.Addr)).To(gomega.Equal("10.135.0.0"))
	g.Expect(*(nodes[0].StaticRoutes[0].Prefix.Mask)).To(gomega.Equal(int32(26)))
	err = KubeClient.CoreV1().Nodes().Delete(context.TODO(), newNode.Name, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("error in deleting Node: %v", err)
	}
}

func TestCiliumNodeAddUpdate(t *testing.T) {
	if *cniPlugin != "cilium" {
		t.Skip("Skipping CiliumNode test since CNI plugin is not Cilium")
	}
	g := gomega.NewGomegaWithT(t)
	modelName := "admin/global"
	nodeip := "10.1.1.2"
	nodeName := "testNodeCilium-1"
	objects.SharedAviGraphLister().Delete(modelName)

	// mimicking actual scenario where the node will have atleast one CiliumNode object created from start
	var testData unstructured.Unstructured
	specAddress1 := map[string]interface{}{
		"ip":   nodeip,
		"type": "InternalIP",
	}
	specAddress2 := map[string]interface{}{
		"ip":   "10.128.0.33",
		"type": "CiliumInternalIP",
	}
	specAddressList := []interface{}{
		specAddress1,
		specAddress2,
	}
	testData.SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "cilium.io/v2",
		"kind":       "ciliumnodes",
		"metadata": map[string]interface{}{
			"name":      nodeName,
			"namespace": "default",
		},
		"spec": map[string]interface{}{
			"addresses": specAddressList,
			"ipam": map[string]interface{}{
				"podCIDRs": []interface{}{"10.128.0.0/23"},
			},
		},
	})

	DynamicClient.Resource(lib.CiliumNodeGVR).Namespace("default").Create(context.TODO(), &testData, v1.CreateOptions{})

	nodeExample := (integrationtest.FakeNode{
		Name:     nodeName,
		Version:  "1",
		NodeIP:   nodeip,
		PodCIDR:  "10.133.0.0/23",
		PodCIDRs: []string{"10.133.0.0/23"},
	}).Node()

	_, err := KubeClient.CoreV1().Nodes().Create(context.TODO(), nodeExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Node: %v", err)
	}

	integrationtest.PollForCompletion(t, modelName, 5)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 10*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	g.Expect(aviModel.(*avinodes.AviObjectGraph).IsVrf).To(gomega.Equal(true))
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVRF()
	g.Expect(len(nodes)).To(gomega.Equal(1))

	g.Expect(len(nodes[0].StaticRoutes)).To(gomega.Equal(1))
	g.Expect(*(nodes[0].StaticRoutes[0].NextHop.Addr)).To(gomega.Equal(nodeip))
	g.Expect(*(nodes[0].StaticRoutes[0].Prefix.IPAddr.Addr)).To(gomega.Equal("10.128.0.0"))
	g.Expect(*(nodes[0].StaticRoutes[0].Prefix.Mask)).To(gomega.Equal(int32(23)))

	// adding a new podcidr to the ciliumnode object for the node
	testData.SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "cilium.io/v2",
		"kind":       "ciliumnodes",
		"metadata": map[string]interface{}{
			"name":      nodeName,
			"namespace": "default",
		},
		"spec": map[string]interface{}{
			"addresses": specAddressList,
			"ipam": map[string]interface{}{
				"podCIDRs": []interface{}{"10.128.0.0/23", "10.129.0.0/23"},
			},
		},
	})

	DynamicClient.Resource(lib.CiliumNodeGVR).Namespace("default").Update(context.TODO(), &testData, v1.UpdateOptions{})
	// waiting for routes to get populated after CiliumNode object creation
	time.Sleep(5 * time.Second)

	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	g.Expect(aviModel.(*avinodes.AviObjectGraph).IsVrf).To(gomega.Equal(true))
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVRF()
	g.Expect(len(nodes)).To(gomega.Equal(1))

	g.Expect(len(nodes[0].StaticRoutes)).To(gomega.Equal(2))
	g.Expect(*(nodes[0].StaticRoutes[1].NextHop.Addr)).To(gomega.Equal(nodeip))
	g.Expect(*(nodes[0].StaticRoutes[1].Prefix.IPAddr.Addr)).To(gomega.Equal("10.129.0.0"))
	g.Expect(*(nodes[0].StaticRoutes[1].Prefix.Mask)).To(gomega.Equal(int32(23)))

	// deleting the new podcidr from the ciliumnode object for the node
	testData.SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "cilium.io/v2",
		"kind":       "ciliumnodes",
		"metadata": map[string]interface{}{
			"name":      nodeName,
			"namespace": "default",
		},
		"spec": map[string]interface{}{
			"addresses": specAddressList,
			"ipam": map[string]interface{}{
				"podCIDRs": []interface{}{"10.128.0.0/23"},
			},
		},
	})

	time.Sleep(5 * time.Second)
	// deleting the second podcidr from CiliumNode object for the node
	DynamicClient.Resource(lib.CiliumNodeGVR).Namespace("default").Update(context.TODO(), &testData, v1.UpdateOptions{})
	g.Eventually(func() int {
		num_routes := len(nodes[0].StaticRoutes)
		return num_routes
	}, 10*time.Second).Should(gomega.Equal(1))
}

func TestCiliumNodeAddDelete(t *testing.T) {
	if *cniPlugin != "cilium" {
		t.Skip("Skipping CiliumNode test since CNI plugin is not Cilium")
	}
	g := gomega.NewGomegaWithT(t)
	modelName := "admin/global"
	nodeip1 := "10.1.1.2"
	nodeName1 := "testNodeCilium-2"
	nodeip2 := "10.1.1.3"
	nodeName2 := "testNodeCilium-3"
	objects.SharedAviGraphLister().Delete(modelName)

	// mimicking actual scenario where the node will have atleast one CiliumNode object created from start
	var testData unstructured.Unstructured
	specAddress1 := map[string]interface{}{
		"ip":   nodeip1,
		"type": "InternalIP",
	}
	specAddress2 := map[string]interface{}{
		"ip":   "10.128.0.33",
		"type": "CiliumInternalIP",
	}
	specAddressList := []interface{}{
		specAddress1,
		specAddress2,
	}
	testData.SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "cilium.io/v2",
		"kind":       "ciliumnodes",
		"metadata": map[string]interface{}{
			"name":      nodeName1,
			"namespace": "default",
		},
		"spec": map[string]interface{}{
			"addresses": specAddressList,
			"ipam": map[string]interface{}{
				"podCIDRs": []interface{}{"10.128.0.0/23"},
			},
		},
	})

	DynamicClient.Resource(lib.CiliumNodeGVR).Namespace("default").Create(context.TODO(), &testData, v1.CreateOptions{})

	nodeExample := (integrationtest.FakeNode{
		Name:     nodeName1,
		Version:  "1",
		NodeIP:   nodeip1,
		PodCIDR:  "10.133.0.0/23",
		PodCIDRs: []string{"10.133.0.0/23"},
	}).Node()

	_, err := KubeClient.CoreV1().Nodes().Create(context.TODO(), nodeExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Node: %v", err)
	}

	integrationtest.PollForCompletion(t, modelName, 5)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 10*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	g.Expect(aviModel.(*avinodes.AviObjectGraph).IsVrf).To(gomega.Equal(true))
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVRF()
	g.Expect(len(nodes)).To(gomega.Equal(1))

	g.Expect(len(nodes[0].StaticRoutes)).To(gomega.Equal(1))
	g.Expect(*(nodes[0].StaticRoutes[0].NextHop.Addr)).To(gomega.Equal(nodeip1))
	g.Expect(*(nodes[0].StaticRoutes[0].Prefix.IPAddr.Addr)).To(gomega.Equal("10.128.0.0"))
	g.Expect(*(nodes[0].StaticRoutes[0].Prefix.Mask)).To(gomega.Equal(int32(23)))

	//creating second node
	nodeExample1 := (integrationtest.FakeNode{
		Name:     nodeName2,
		Version:  "1",
		NodeIP:   nodeip2,
		PodCIDR:  "10.132.0.0/23",
		PodCIDRs: []string{"10.132.0.0/23"},
	}).Node()

	_, err = KubeClient.CoreV1().Nodes().Create(context.TODO(), nodeExample1, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Node: %v", err)
	}

	// mimicking actual scenario where the node will have atleast one CiliumNode object created from start
	var testData1 unstructured.Unstructured
	specAddress3 := map[string]interface{}{
		"ip":   nodeip2,
		"type": "InternalIP",
	}
	specAddress4 := map[string]interface{}{
		"ip":   "10.129.0.33",
		"type": "CiliumInternalIP",
	}
	specAddressList2 := []interface{}{
		specAddress3,
		specAddress4,
	}
	// adding a new ciliumnode object for the second node
	testData1.SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "cilium.io/v2",
		"kind":       "ciliumnodes",
		"metadata": map[string]interface{}{
			"name":      nodeName2,
			"namespace": "default",
		},
		"spec": map[string]interface{}{
			"addresses": specAddressList2,
			"ipam": map[string]interface{}{
				"podCIDRs": []interface{}{"10.129.0.0/23"},
			},
		},
	})

	DynamicClient.Resource(lib.CiliumNodeGVR).Namespace("default").Create(context.TODO(), &testData1, v1.CreateOptions{})
	// waiting for routes to get populated after CiliumNode object creation
	time.Sleep(5 * time.Second)

	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	g.Expect(aviModel.(*avinodes.AviObjectGraph).IsVrf).To(gomega.Equal(true))
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVRF()
	g.Expect(len(nodes)).To(gomega.Equal(1))

	g.Expect(len(nodes[0].StaticRoutes)).To(gomega.Equal(2))
	g.Expect(*(nodes[0].StaticRoutes[1].NextHop.Addr)).To(gomega.Equal(nodeip2))
	g.Expect(*(nodes[0].StaticRoutes[1].Prefix.IPAddr.Addr)).To(gomega.Equal("10.129.0.0"))
	g.Expect(*(nodes[0].StaticRoutes[1].Prefix.Mask)).To(gomega.Equal(int32(23)))

	// deleting ciliumnode for second node
	DynamicClient.Resource(lib.CiliumNodeGVR).Namespace("default").Delete(context.TODO(), nodeName2, v1.DeleteOptions{})
	g.Eventually(func() int {
		num_routes := len(nodes[0].StaticRoutes)
		return num_routes
	}, 10*time.Second).Should(gomega.Equal(1))
}
func TestMultipleBlockAffinityAddition(t *testing.T) {
	if *cniPlugin != "calico" {
		t.Skip("Skipping BlockAffinity test since CNI plugin is not Calico")
	}
	g := gomega.NewGomegaWithT(t)
	modelName := "admin/global"
	nodeip1 := "10.102.99.232"
	nodeName1 := "testNodeCalico1"
	objects.SharedAviGraphLister().Delete(modelName)
	time.Sleep(10 * time.Second)

	// mimicking actual scenario where the node will have atleast one BlockAffinity object created from start
	var testData1 unstructured.Unstructured
	testData1.SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "crd.projectcalico.org/v1",
		"kind":       "blockaffinities",
		"metadata": map[string]interface{}{
			"name":      "testblockaffinity1",
			"namespace": "default",
		},
		"spec": map[string]interface{}{
			"cidr":    "192.168.29.64/26",
			"deleted": "false",
			"node":    nodeName1,
			"state":   "confirmed",
		},
	})

	DynamicClient.Resource(lib.CalicoBlockaffinityGVR).Namespace("default").Create(context.TODO(), &testData1, v1.CreateOptions{})

	nodeExample := (integrationtest.FakeNode{
		Name:    nodeName1,
		Version: "1",
		NodeIP:  nodeip1,
	}).NodeCalico()

	_, err := KubeClient.CoreV1().Nodes().Create(context.TODO(), nodeExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Node: %v", err)
	}

	// creating a new BlockAffinity object for the node
	var testData2 unstructured.Unstructured
	testData2.SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "crd.projectcalico.org/v1",
		"kind":       "blockaffinities",
		"metadata": map[string]interface{}{
			"name":      "testblockaffinity2",
			"namespace": "default",
		},
		"spec": map[string]interface{}{
			"cidr":    "192.168.30.64/26",
			"deleted": "false",
			"node":    nodeName1,
			"state":   "confirmed",
		},
	})

	DynamicClient.Resource(lib.CalicoBlockaffinityGVR).Namespace("default").Create(context.TODO(), &testData2, v1.CreateOptions{})

	nodeip2 := "10.102.99.146"
	nodeName2 := "testNodeCalico2"

	// mimicking actual scenario where the node will have atleast one BlockAffinity object created from start
	var testData3 unstructured.Unstructured
	testData3.SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "crd.projectcalico.org/v1",
		"kind":       "blockaffinities",
		"metadata": map[string]interface{}{
			"name":      "testblockaffinity3",
			"namespace": "default",
		},
		"spec": map[string]interface{}{
			"cidr":    "192.168.247.0/26",
			"deleted": "false",
			"node":    nodeName2,
			"state":   "confirmed",
		},
	})

	DynamicClient.Resource(lib.CalicoBlockaffinityGVR).Namespace("default").Create(context.TODO(), &testData3, v1.CreateOptions{})

	nodeExample2 := (integrationtest.FakeNode{
		Name:    nodeName2,
		Version: "1",
		NodeIP:  nodeip2,
	}).NodeCalico()

	_, err = KubeClient.CoreV1().Nodes().Create(context.TODO(), nodeExample2, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Node: %v", err)
	}

	integrationtest.PollForCompletion(t, modelName, 5)

	// creating a new BlockAffinity object for the node
	var testData4 unstructured.Unstructured
	testData4.SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "crd.projectcalico.org/v1",
		"kind":       "blockaffinities",
		"metadata": map[string]interface{}{
			"name":      "testblockaffinity4",
			"namespace": "default",
		},
		"spec": map[string]interface{}{
			"cidr":    "192.168.246.0/26",
			"deleted": "false",
			"node":    nodeName2,
			"state":   "confirmed",
		},
	})

	DynamicClient.Resource(lib.CalicoBlockaffinityGVR).Namespace("default").Create(context.TODO(), &testData4, v1.CreateOptions{})

	g.Eventually(func() bool {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVRF()
		if len(nodes[0].StaticRoutes) < 4 {
			found = false
		}
		return found
	}, 10*time.Second).Should(gomega.Equal(true))

	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	g.Expect(aviModel.(*avinodes.AviObjectGraph).IsVrf).To(gomega.Equal(true))
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVRF()
	g.Expect(len(nodes)).To(gomega.Equal(1))

	g.Expect(len(nodes[0].StaticRoutes)).To(gomega.Equal(4))
	g.Expect(*(nodes[0].StaticRoutes[3].NextHop.Addr)).To(gomega.Equal(nodeip2))
	g.Expect(*(nodes[0].StaticRoutes[3].Prefix.IPAddr.Addr)).To(gomega.Equal("192.168.246.0"))
	g.Expect(*(nodes[0].StaticRoutes[3].Prefix.Mask)).To(gomega.Equal(int32(26)))

	var testData5 unstructured.Unstructured
	testData5.SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "crd.projectcalico.org/v1",
		"kind":       "blockaffinities",
		"metadata": map[string]interface{}{
			"name":      "testblockaffinity5",
			"namespace": "default",
		},
		"spec": map[string]interface{}{
			"cidr":    "192.168.31.64/26",
			"deleted": "false",
			"node":    nodeName1,
			"state":   "confirmed",
		},
	})

	DynamicClient.Resource(lib.CalicoBlockaffinityGVR).Namespace("default").Create(context.TODO(), &testData5, v1.CreateOptions{})

	g.Eventually(func() bool {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVRF()
		if len(nodes[0].StaticRoutes) < 5 {
			found = false
		}
		return found
	}, 10*time.Second).Should(gomega.Equal(true))

	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	g.Expect(aviModel.(*avinodes.AviObjectGraph).IsVrf).To(gomega.Equal(true))
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVRF()
	g.Expect(len(nodes)).To(gomega.Equal(1))

	g.Expect(len(nodes[0].StaticRoutes)).To(gomega.Equal(5))

	g.Expect(*nodes[0].StaticRoutes[0].RouteID).To(gomega.Equal("cluster-" + nodes[0].NodeStaticRoutes["testNodeCalico2"].RouteIDPrefix + "-0"))
	g.Expect(*nodes[0].StaticRoutes[1].RouteID).To(gomega.Equal("cluster-" + nodes[0].NodeStaticRoutes["testNodeCalico2"].RouteIDPrefix + "-1"))
	g.Expect(*nodes[0].StaticRoutes[2].RouteID).To(gomega.Equal("cluster-" + nodes[0].NodeStaticRoutes["testNodeCalico1"].RouteIDPrefix + "-0"))
	g.Expect(*nodes[0].StaticRoutes[3].RouteID).To(gomega.Equal("cluster-" + nodes[0].NodeStaticRoutes["testNodeCalico1"].RouteIDPrefix + "-1"))
	g.Expect(*nodes[0].StaticRoutes[4].RouteID).To(gomega.Equal("cluster-" + nodes[0].NodeStaticRoutes["testNodeCalico1"].RouteIDPrefix + "-2"))
	time.Sleep(10 * time.Second)

	// deleting the BlockAffinity objects for the node
	DynamicClient.Resource(lib.CalicoBlockaffinityGVR).Namespace("default").Delete(context.TODO(), "testblockaffinity2", v1.DeleteOptions{})
	DynamicClient.Resource(lib.CalicoBlockaffinityGVR).Namespace("default").Delete(context.TODO(), "testblockaffinity5", v1.DeleteOptions{})
	g.Eventually(func() int {
		num_routes := len(nodes[0].StaticRoutes)
		return num_routes
	}, 10*time.Second).Should(gomega.Equal(3))
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVRF()
	g.Expect(*nodes[0].StaticRoutes[0].RouteID).To(gomega.Equal("cluster-" + nodes[0].NodeStaticRoutes["testNodeCalico2"].RouteIDPrefix + "-0"))
	g.Expect(*nodes[0].StaticRoutes[1].RouteID).To(gomega.Equal("cluster-" + nodes[0].NodeStaticRoutes["testNodeCalico2"].RouteIDPrefix + "-1"))
	g.Expect(*nodes[0].StaticRoutes[2].RouteID).To(gomega.Equal("cluster-" + nodes[0].NodeStaticRoutes["testNodeCalico1"].RouteIDPrefix + "-0"))

	KubeClient.CoreV1().Nodes().Delete(context.TODO(), nodeName1, metav1.DeleteOptions{})
	DynamicClient.Resource(lib.CalicoBlockaffinityGVR).Namespace("default").Delete(context.TODO(), "testblockaffinity1", v1.DeleteOptions{})
	g.Eventually(func() int {
		num_routes := len(nodes[0].StaticRoutes)
		return num_routes
	}, 10*time.Second).Should(gomega.Equal(2))
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVRF()
	g.Expect(*nodes[0].StaticRoutes[0].RouteID).To(gomega.Equal("cluster-" + nodes[0].NodeStaticRoutes["testNodeCalico2"].RouteIDPrefix + "-0"))
	g.Expect(*nodes[0].StaticRoutes[1].RouteID).To(gomega.Equal("cluster-" + nodes[0].NodeStaticRoutes["testNodeCalico2"].RouteIDPrefix + "-1"))

	DynamicClient.Resource(lib.CalicoBlockaffinityGVR).Namespace("default").Delete(context.TODO(), "testblockaffinity4", v1.DeleteOptions{})
	g.Eventually(func() int {
		num_routes := len(nodes[0].StaticRoutes)
		return num_routes
	}, 10*time.Second).Should(gomega.Equal(1))
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVRF()
	g.Expect(*nodes[0].StaticRoutes[0].RouteID).To(gomega.Equal("cluster-" + nodes[0].NodeStaticRoutes["testNodeCalico2"].RouteIDPrefix + "-0"))

	KubeClient.CoreV1().Nodes().Delete(context.TODO(), nodeName2, metav1.DeleteOptions{})
	DynamicClient.Resource(lib.CalicoBlockaffinityGVR).Namespace("default").Delete(context.TODO(), "testblockaffinity3", v1.DeleteOptions{})
	g.Eventually(func() int {
		num_routes := len(nodes[0].StaticRoutes)
		return num_routes
	}, 10*time.Second).Should(gomega.Equal(0))
}
func TestStaticRoutesWithMultipleNodesUpdate(t *testing.T) {
	if *cniPlugin != "calico" {
		t.Skip("Skipping BlockAffinity test since CNI plugin is not Calico")
	}
	g := gomega.NewGomegaWithT(t)
	modelName := "admin/global"
	nodesNameList := []string{"testNodeCalico1", "testNodeCalico2", "testNodeCalico3", "testNodeCalico4"}
	nodeIPList := []string{"10.64.29.32", "10.64.29.33", "10.64.29.34", "10.64.29.35"}
	objects.SharedAviGraphLister().Delete(modelName)
	time.Sleep(10 * time.Second)

	// mimicking actual scenario where the node will have atleast one BlockAffinity object created from start
	var testData1 unstructured.Unstructured
	testData1.SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "crd.projectcalico.org/v1",
		"kind":       "blockaffinities",
		"metadata": map[string]interface{}{
			"name":      "testblockaffinity1",
			"namespace": "default",
		},
		"spec": map[string]interface{}{
			"cidr":    "10.244.255.236/30",
			"deleted": "false",
			"node":    nodesNameList[0],
			"state":   "confirmed",
		},
	})

	DynamicClient.Resource(lib.CalicoBlockaffinityGVR).Namespace("default").Create(context.TODO(), &testData1, v1.CreateOptions{})

	nodeExample := (integrationtest.FakeNode{
		Name:    nodesNameList[0],
		Version: "1",
		NodeIP:  nodeIPList[0],
	}).NodeCalico()

	_, err := KubeClient.CoreV1().Nodes().Create(context.TODO(), nodeExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Node: %v", err)
	}

	// mimicking actual scenario where the node will have atleast one BlockAffinity object created from start
	var testData2 unstructured.Unstructured
	testData2.SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "crd.projectcalico.org/v1",
		"kind":       "blockaffinities",
		"metadata": map[string]interface{}{
			"name":      "testblockaffinity2",
			"namespace": "default",
		},
		"spec": map[string]interface{}{
			"cidr":    "10.244.255.72/30",
			"deleted": "false",
			"node":    nodesNameList[1],
			"state":   "confirmed",
		},
	})

	DynamicClient.Resource(lib.CalicoBlockaffinityGVR).Namespace("default").Create(context.TODO(), &testData2, v1.CreateOptions{})

	nodeExample2 := (integrationtest.FakeNode{
		Name:    nodesNameList[1],
		Version: "1",
		NodeIP:  nodeIPList[1],
	}).NodeCalico()

	_, err = KubeClient.CoreV1().Nodes().Create(context.TODO(), nodeExample2, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Node: %v", err)
	}

	integrationtest.PollForCompletion(t, modelName, 5)

	// mimicking actual scenario where the node will have atleast one BlockAffinity object created from start
	var testData3 unstructured.Unstructured
	testData3.SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "crd.projectcalico.org/v1",
		"kind":       "blockaffinities",
		"metadata": map[string]interface{}{
			"name":      "testblockaffinity3",
			"namespace": "default",
		},
		"spec": map[string]interface{}{
			"cidr":    "10.244.115.128/25",
			"deleted": "false",
			"node":    nodesNameList[2],
			"state":   "confirmed",
		},
	})

	DynamicClient.Resource(lib.CalicoBlockaffinityGVR).Namespace("default").Create(context.TODO(), &testData3, v1.CreateOptions{})

	nodeExample3 := (integrationtest.FakeNode{
		Name:    nodesNameList[2],
		Version: "1",
		NodeIP:  nodeIPList[2],
	}).NodeCalico()

	_, err = KubeClient.CoreV1().Nodes().Create(context.TODO(), nodeExample3, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Node: %v", err)
	}

	integrationtest.PollForCompletion(t, modelName, 5)

	// mimicking actual scenario where the node will have atleast one BlockAffinity object created from start
	var testData4 unstructured.Unstructured
	testData4.SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "crd.projectcalico.org/v1",
		"kind":       "blockaffinities",
		"metadata": map[string]interface{}{
			"name":      "testblockaffinity4",
			"namespace": "default",
		},
		"spec": map[string]interface{}{
			"cidr":    "10.244.125.128/25",
			"deleted": "false",
			"node":    nodesNameList[3],
			"state":   "confirmed",
		},
	})

	DynamicClient.Resource(lib.CalicoBlockaffinityGVR).Namespace("default").Create(context.TODO(), &testData4, v1.CreateOptions{})

	nodeExample4 := (integrationtest.FakeNode{
		Name:    nodesNameList[3],
		Version: "1",
		NodeIP:  nodeIPList[3],
	}).NodeCalico()

	_, err = KubeClient.CoreV1().Nodes().Create(context.TODO(), nodeExample4, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Node: %v", err)
	}

	integrationtest.PollForCompletion(t, modelName, 5)

	g.Eventually(func() bool {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVRF()
		if len(nodes[0].StaticRoutes) < 4 {
			found = false
		}
		return found
	}, 10*time.Second).Should(gomega.Equal(true))

	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	g.Expect(aviModel.(*avinodes.AviObjectGraph).IsVrf).To(gomega.Equal(true))
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVRF()
	g.Expect(len(nodes)).To(gomega.Equal(1))

	DynamicClient.Resource(lib.CalicoBlockaffinityGVR).Namespace("default").Delete(context.TODO(), "testblockaffinity2", v1.DeleteOptions{})
	g.Eventually(func() int {
		num_routes := len(nodes[0].StaticRoutes)
		return num_routes
	}, 10*time.Second).Should(gomega.Equal(3))
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVRF()

	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	g.Expect(aviModel.(*avinodes.AviObjectGraph).IsVrf).To(gomega.Equal(true))
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVRF()
	g.Expect(len(nodes)).To(gomega.Equal(1))

	KubeClient.CoreV1().Nodes().Delete(context.TODO(), nodesNameList[1], metav1.DeleteOptions{})
	g.Eventually(func() int {
		num_routes := len(nodes[0].NodeStaticRoutes)
		return num_routes
	}, 10*time.Second).Should(gomega.Equal(3))
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVRF()

	g.Expect(len(nodes)).To(gomega.Equal(1))
	g.Expect(len(nodes[0].NodeStaticRoutes)).To(gomega.Equal(3))
	g.Expect(len(nodes[0].Nodes)).To(gomega.Equal(3))

	DynamicClient.Resource(lib.CalicoBlockaffinityGVR).Namespace("default").Delete(context.TODO(), "testblockaffinity1", v1.DeleteOptions{})
	DynamicClient.Resource(lib.CalicoBlockaffinityGVR).Namespace("default").Delete(context.TODO(), "testblockaffinity3", v1.DeleteOptions{})
	DynamicClient.Resource(lib.CalicoBlockaffinityGVR).Namespace("default").Delete(context.TODO(), "testblockaffinity4", v1.DeleteOptions{})
	g.Eventually(func() int {
		num_routes := len(nodes[0].StaticRoutes)
		return num_routes
	}, 10*time.Second).Should(gomega.Equal(0))
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVRF()

	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	g.Expect(aviModel.(*avinodes.AviObjectGraph).IsVrf).To(gomega.Equal(true))
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVRF()
	g.Expect(len(nodes)).To(gomega.Equal(1))

	KubeClient.CoreV1().Nodes().Delete(context.TODO(), nodesNameList[0], metav1.DeleteOptions{})
	KubeClient.CoreV1().Nodes().Delete(context.TODO(), nodesNameList[2], metav1.DeleteOptions{})
	KubeClient.CoreV1().Nodes().Delete(context.TODO(), nodesNameList[3], metav1.DeleteOptions{})
	g.Eventually(func() int {
		num_routes := len(nodes[0].NodeStaticRoutes)
		return num_routes
	}, 10*time.Second).Should(gomega.Equal(0))
}

func TestStaticRoutesWithMultipleBlockAffinityDeletion(t *testing.T) {
	if *cniPlugin != "calico" {
		t.Skip("Skipping BlockAffinity test since CNI plugin is not Calico")
	}
	g := gomega.NewGomegaWithT(t)
	modelName := "admin/global"
	nodesNameList := []string{"testNodeCalico1", "testNodeCalico2", "testNodeCalico3", "testNodeCalico4"}
	nodeIPList := []string{"10.64.29.32", "10.64.29.33", "10.64.29.34", "10.64.29.35"}
	objects.SharedAviGraphLister().Delete(modelName)
	time.Sleep(10 * time.Second)

	// mimicking actual scenario where the node will have atleast one BlockAffinity object created from start
	var testData1 unstructured.Unstructured
	testData1.SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "crd.projectcalico.org/v1",
		"kind":       "blockaffinities",
		"metadata": map[string]interface{}{
			"name":      "testblockaffinity1",
			"namespace": "default",
		},
		"spec": map[string]interface{}{
			"cidr":    "10.244.255.236/30",
			"deleted": "false",
			"node":    nodesNameList[0],
			"state":   "confirmed",
		},
	})

	DynamicClient.Resource(lib.CalicoBlockaffinityGVR).Namespace("default").Create(context.TODO(), &testData1, v1.CreateOptions{})

	nodeExample := (integrationtest.FakeNode{
		Name:    nodesNameList[0],
		Version: "1",
		NodeIP:  nodeIPList[0],
	}).NodeCalico()

	_, err := KubeClient.CoreV1().Nodes().Create(context.TODO(), nodeExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Node: %v", err)
	}
	integrationtest.PollForCompletion(t, modelName, 5)
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVRF()

	var banode1 unstructured.Unstructured
	banode1.SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "crd.projectcalico.org/v1",
		"kind":       "blockaffinities",
		"metadata": map[string]interface{}{
			"name":      "testblockaffinity11",
			"namespace": "default",
		},
		"spec": map[string]interface{}{
			"cidr":    "10.244.255.239/30",
			"deleted": "false",
			"node":    nodesNameList[0],
			"state":   "confirmed",
		},
	})

	DynamicClient.Resource(lib.CalicoBlockaffinityGVR).Namespace("default").Create(context.TODO(), &banode1, v1.CreateOptions{})

	g.Eventually(func() bool {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVRF()
		if len(nodes[0].StaticRoutes) < 2 {
			found = false
		}
		return found
	}, 30*time.Second).Should(gomega.Equal(true))
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVRF()

	// mimicking actual scenario where the node will have atleast one BlockAffinity object created from start
	var testData2 unstructured.Unstructured
	testData2.SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "crd.projectcalico.org/v1",
		"kind":       "blockaffinities",
		"metadata": map[string]interface{}{
			"name":      "testblockaffinity2",
			"namespace": "default",
		},
		"spec": map[string]interface{}{
			"cidr":    "10.244.255.72/30",
			"deleted": "false",
			"node":    nodesNameList[1],
			"state":   "confirmed",
		},
	})

	DynamicClient.Resource(lib.CalicoBlockaffinityGVR).Namespace("default").Create(context.TODO(), &testData2, v1.CreateOptions{})

	nodeExample2 := (integrationtest.FakeNode{
		Name:    nodesNameList[1],
		Version: "1",
		NodeIP:  nodeIPList[1],
	}).NodeCalico()

	_, err = KubeClient.CoreV1().Nodes().Create(context.TODO(), nodeExample2, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Node: %v", err)
	}

	integrationtest.PollForCompletion(t, modelName, 5)
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVRF()

	g.Eventually(func() int {
		num_routes := len(nodes[0].NodeStaticRoutes)
		return num_routes
	}, 10*time.Second).Should(gomega.Equal(2))
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVRF()

	g.Eventually(func() bool {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVRF()
		if len(nodes[0].StaticRoutes) < 3 {
			found = false
		}
		return found
	}, 10*time.Second).Should(gomega.Equal(true))

	// mimicking actual scenario where the node will have atleast one BlockAffinity object created from start
	var testData3 unstructured.Unstructured
	testData3.SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "crd.projectcalico.org/v1",
		"kind":       "blockaffinities",
		"metadata": map[string]interface{}{
			"name":      "testblockaffinity3",
			"namespace": "default",
		},
		"spec": map[string]interface{}{
			"cidr":    "10.244.115.128/25",
			"deleted": "false",
			"node":    nodesNameList[2],
			"state":   "confirmed",
		},
	})

	DynamicClient.Resource(lib.CalicoBlockaffinityGVR).Namespace("default").Create(context.TODO(), &testData3, v1.CreateOptions{})

	nodeExample3 := (integrationtest.FakeNode{
		Name:    nodesNameList[2],
		Version: "1",
		NodeIP:  nodeIPList[2],
	}).NodeCalico()

	_, err = KubeClient.CoreV1().Nodes().Create(context.TODO(), nodeExample3, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Node: %v", err)
	}

	g.Eventually(func() bool {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVRF()
		if len(nodes[0].StaticRoutes) < 4 {
			found = false
		}
		return found
	}, 10*time.Second).Should(gomega.Equal(true))
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVRF()

	g.Eventually(func() int {
		num_routes := len(nodes[0].NodeStaticRoutes)
		return num_routes
	}, 10*time.Second).Should(gomega.Equal(3))
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVRF()

	DynamicClient.Resource(lib.CalicoBlockaffinityGVR).Namespace("default").Delete(context.TODO(), "testblockaffinity2", v1.DeleteOptions{})
	g.Eventually(func() bool {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVRF()
		if len(nodes[0].StaticRoutes) > 3 {
			found = false
		}
		return found
	}, 10*time.Second).Should(gomega.Equal(true))
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVRF()

	DynamicClient.Resource(lib.CalicoBlockaffinityGVR).Namespace("default").Delete(context.TODO(), "testblockaffinity3", v1.DeleteOptions{})
	g.Eventually(func() bool {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVRF()
		if len(nodes[0].StaticRoutes) > 2 {
			found = false
		}
		return found
	}, 10*time.Second).Should(gomega.Equal(true))
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVRF()

	// mimicking actual scenario where the node will have atleast one BlockAffinity object created from start
	var testData4 unstructured.Unstructured
	testData4.SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "crd.projectcalico.org/v1",
		"kind":       "blockaffinities",
		"metadata": map[string]interface{}{
			"name":      "testblockaffinity4",
			"namespace": "default",
		},
		"spec": map[string]interface{}{
			"cidr":    "10.244.125.128/25",
			"deleted": "false",
			"node":    nodesNameList[3],
			"state":   "confirmed",
		},
	})

	DynamicClient.Resource(lib.CalicoBlockaffinityGVR).Namespace("default").Create(context.TODO(), &testData4, v1.CreateOptions{})

	nodeExample4 := (integrationtest.FakeNode{
		Name:    nodesNameList[3],
		Version: "1",
		NodeIP:  nodeIPList[3],
	}).NodeCalico()

	_, err = KubeClient.CoreV1().Nodes().Create(context.TODO(), nodeExample4, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Node: %v", err)
	}

	integrationtest.PollForCompletion(t, modelName, 5)

	g.Eventually(func() int {
		num_routes := len(nodes[0].NodeStaticRoutes)
		return num_routes
	}, 10*time.Second).Should(gomega.Equal(2))
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVRF()

	g.Eventually(func() bool {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVRF()
		if len(nodes[0].StaticRoutes) < 3 {
			found = false
		}
		return found
	}, 10*time.Second).Should(gomega.Equal(true))

	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	g.Expect(aviModel.(*avinodes.AviObjectGraph).IsVrf).To(gomega.Equal(true))
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVRF()

	var banode4 unstructured.Unstructured
	banode4.SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "crd.projectcalico.org/v1",
		"kind":       "blockaffinities",
		"metadata": map[string]interface{}{
			"name":      "testblockaffinity44",
			"namespace": "default",
		},
		"spec": map[string]interface{}{
			"cidr":    "10.244.127.128/25",
			"deleted": "false",
			"node":    nodesNameList[3],
			"state":   "confirmed",
		},
	})

	DynamicClient.Resource(lib.CalicoBlockaffinityGVR).Namespace("default").Create(context.TODO(), &banode4, v1.CreateOptions{})

	g.Eventually(func() bool {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVRF()
		if len(nodes[0].StaticRoutes) < 4 {
			found = false
		}
		return found
	}, 10*time.Second).Should(gomega.Equal(true))

	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	g.Expect(aviModel.(*avinodes.AviObjectGraph).IsVrf).To(gomega.Equal(true))
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVRF()
	g.Expect(len(nodes)).To(gomega.Equal(1))

	DynamicClient.Resource(lib.CalicoBlockaffinityGVR).Namespace("default").Delete(context.TODO(), "testblockaffinity44", v1.DeleteOptions{})
	DynamicClient.Resource(lib.CalicoBlockaffinityGVR).Namespace("default").Delete(context.TODO(), "testblockaffinity4", v1.DeleteOptions{})
	DynamicClient.Resource(lib.CalicoBlockaffinityGVR).Namespace("default").Delete(context.TODO(), "testblockaffinity3", v1.DeleteOptions{})
	DynamicClient.Resource(lib.CalicoBlockaffinityGVR).Namespace("default").Delete(context.TODO(), "testblockaffinity2", v1.DeleteOptions{})
	DynamicClient.Resource(lib.CalicoBlockaffinityGVR).Namespace("default").Delete(context.TODO(), "testblockaffinity11", v1.DeleteOptions{})
	DynamicClient.Resource(lib.CalicoBlockaffinityGVR).Namespace("default").Delete(context.TODO(), "testblockaffinity1", v1.DeleteOptions{})
	g.Eventually(func() bool {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVRF()
		if len(nodes[0].StaticRoutes) > 0 {
			found = false
		}
		return found
	}, 10*time.Second).Should(gomega.Equal(true))
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVRF()

	KubeClient.CoreV1().Nodes().Delete(context.TODO(), nodesNameList[0], metav1.DeleteOptions{})
	KubeClient.CoreV1().Nodes().Delete(context.TODO(), nodesNameList[1], metav1.DeleteOptions{})
	KubeClient.CoreV1().Nodes().Delete(context.TODO(), nodesNameList[2], metav1.DeleteOptions{})
	KubeClient.CoreV1().Nodes().Delete(context.TODO(), nodesNameList[3], metav1.DeleteOptions{})

	g.Eventually(func() int {
		num_routes := len(nodes[0].NodeStaticRoutes)
		return num_routes
	}, 10*time.Second).Should(gomega.Equal(0))
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVRF()
	g.Expect(len(nodes[0].StaticRoutes)).To(gomega.Equal(0))
	g.Expect(len(nodes[0].NodeStaticRoutes)).To(gomega.Equal(0))
	g.Expect(len(nodes[0].Nodes)).To(gomega.Equal(0))
}
func TestNodeWithoutBlockAffinityAddition(t *testing.T) {
	if *cniPlugin != "calico" {
		t.Skip("Skipping BlockAffinity test since CNI plugin is not Calico")
	}
	g := gomega.NewGomegaWithT(t)
	modelName := "admin/global"
	nodesNameList := []string{"testNodeCalico1", "testNodeCalico2", "testNodeCalico3", "testNodeCalico4"}
	nodeIPList := []string{"10.64.29.32", "10.64.29.33", "10.64.29.34", "10.64.29.35"}
	objects.SharedAviGraphLister().Delete(modelName)
	time.Sleep(10 * time.Second)

	// mimicking actual scenario where the node will have atleast one BlockAffinity object created from start
	var testData1 unstructured.Unstructured
	testData1.SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "crd.projectcalico.org/v1",
		"kind":       "blockaffinities",
		"metadata": map[string]interface{}{
			"name":      "testblockaffinity1",
			"namespace": "default",
		},
		"spec": map[string]interface{}{
			"cidr":    "10.244.255.236/30",
			"deleted": "false",
			"node":    nodesNameList[0],
			"state":   "confirmed",
		},
	})

	DynamicClient.Resource(lib.CalicoBlockaffinityGVR).Namespace("default").Create(context.TODO(), &testData1, v1.CreateOptions{})

	nodeExample := (integrationtest.FakeNode{
		Name:    nodesNameList[0],
		Version: "1",
		NodeIP:  nodeIPList[0],
	}).NodeCalico()

	_, err := KubeClient.CoreV1().Nodes().Create(context.TODO(), nodeExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Node: %v", err)
	}
	integrationtest.PollForCompletion(t, modelName, 5)
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVRF()

	var banode1 unstructured.Unstructured
	banode1.SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "crd.projectcalico.org/v1",
		"kind":       "blockaffinities",
		"metadata": map[string]interface{}{
			"name":      "testblockaffinity11",
			"namespace": "default",
		},
		"spec": map[string]interface{}{
			"cidr":    "10.244.255.239/30",
			"deleted": "false",
			"node":    nodesNameList[0],
			"state":   "confirmed",
		},
	})

	DynamicClient.Resource(lib.CalicoBlockaffinityGVR).Namespace("default").Create(context.TODO(), &banode1, v1.CreateOptions{})

	g.Eventually(func() bool {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVRF()
		if len(nodes[0].StaticRoutes) < 2 {
			found = false
		}
		return found
	}, 30*time.Second).Should(gomega.Equal(true))
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVRF()

	// mimicking actual scenario where the node will have atleast one BlockAffinity object created from start
	var testData2 unstructured.Unstructured
	testData2.SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "crd.projectcalico.org/v1",
		"kind":       "blockaffinities",
		"metadata": map[string]interface{}{
			"name":      "testblockaffinity2",
			"namespace": "default",
		},
		"spec": map[string]interface{}{
			"cidr":    "10.244.255.72/30",
			"deleted": "false",
			"node":    nodesNameList[1],
			"state":   "confirmed",
		},
	})

	nodeExample2 := (integrationtest.FakeNode{
		Name:    nodesNameList[1],
		Version: "1",
		NodeIP:  nodeIPList[1],
	}).NodeCalico()

	_, err = KubeClient.CoreV1().Nodes().Create(context.TODO(), nodeExample2, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Node: %v", err)
	}

	integrationtest.PollForCompletion(t, modelName, 5)
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVRF()

	g.Eventually(func() int {
		num_routes := len(nodes[0].NodeStaticRoutes)
		return num_routes
	}, 10*time.Second).Should(gomega.Equal(1))
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVRF()
	g.Expect(len(nodes)).To(gomega.Equal(1))
	g.Expect(len(nodes[0].StaticRoutes)).To(gomega.Equal(2))
	g.Expect(len(nodes[0].NodeStaticRoutes)).To(gomega.Equal(1))
	g.Expect(*nodes[0].StaticRoutes[0].RouteID).To(gomega.Equal("cluster-" + nodes[0].NodeStaticRoutes["testNodeCalico1"].RouteIDPrefix + "-0"))
	g.Expect(*nodes[0].StaticRoutes[1].RouteID).To(gomega.Equal("cluster-" + nodes[0].NodeStaticRoutes["testNodeCalico1"].RouteIDPrefix + "-1"))

	g.Eventually(func() bool {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVRF()
		if len(nodes[0].StaticRoutes) < 2 {
			found = false
		}
		return found
	}, 10*time.Second).Should(gomega.Equal(true))

	// mimicking actual scenario where the node will have atleast one BlockAffinity object created from start
	var testData3 unstructured.Unstructured
	testData3.SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "crd.projectcalico.org/v1",
		"kind":       "blockaffinities",
		"metadata": map[string]interface{}{
			"name":      "testblockaffinity3",
			"namespace": "default",
		},
		"spec": map[string]interface{}{
			"cidr":    "10.244.115.128/25",
			"deleted": "false",
			"node":    nodesNameList[2],
			"state":   "confirmed",
		},
	})

	DynamicClient.Resource(lib.CalicoBlockaffinityGVR).Namespace("default").Create(context.TODO(), &testData3, v1.CreateOptions{})

	nodeExample3 := (integrationtest.FakeNode{
		Name:    nodesNameList[2],
		Version: "1",
		NodeIP:  nodeIPList[2],
	}).NodeCalico()

	_, err = KubeClient.CoreV1().Nodes().Create(context.TODO(), nodeExample3, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Node: %v", err)
	}

	g.Eventually(func() bool {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVRF()
		if len(nodes[0].StaticRoutes) < 3 {
			found = false
		}
		return found
	}, 10*time.Second).Should(gomega.Equal(true))
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVRF()

	g.Eventually(func() int {
		num_routes := len(nodes[0].NodeStaticRoutes)
		return num_routes
	}, 10*time.Second).Should(gomega.Equal(2))
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVRF()
	g.Expect(len(nodes)).To(gomega.Equal(1))
	g.Expect(len(nodes[0].StaticRoutes)).To(gomega.Equal(3))
	g.Expect(len(nodes[0].NodeStaticRoutes)).To(gomega.Equal(2))
	g.Expect(*nodes[0].StaticRoutes[0].RouteID).To(gomega.Equal("cluster-" + nodes[0].NodeStaticRoutes["testNodeCalico1"].RouteIDPrefix + "-0"))
	g.Expect(*nodes[0].StaticRoutes[1].RouteID).To(gomega.Equal("cluster-" + nodes[0].NodeStaticRoutes["testNodeCalico1"].RouteIDPrefix + "-1"))
	g.Expect(*nodes[0].StaticRoutes[2].RouteID).To(gomega.Equal("cluster-" + nodes[0].NodeStaticRoutes["testNodeCalico3"].RouteIDPrefix + "-0"))

	// mimicking actual scenario where the node will have atleast one BlockAffinity object created from start
	var testData4 unstructured.Unstructured
	testData4.SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "crd.projectcalico.org/v1",
		"kind":       "blockaffinities",
		"metadata": map[string]interface{}{
			"name":      "testblockaffinity4",
			"namespace": "default",
		},
		"spec": map[string]interface{}{
			"cidr":    "10.244.125.128/25",
			"deleted": "false",
			"node":    nodesNameList[3],
			"state":   "confirmed",
		},
	})

	DynamicClient.Resource(lib.CalicoBlockaffinityGVR).Namespace("default").Create(context.TODO(), &testData4, v1.CreateOptions{})

	nodeExample4 := (integrationtest.FakeNode{
		Name:    nodesNameList[3],
		Version: "1",
		NodeIP:  nodeIPList[3],
	}).NodeCalico()

	_, err = KubeClient.CoreV1().Nodes().Create(context.TODO(), nodeExample4, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Node: %v", err)
	}

	integrationtest.PollForCompletion(t, modelName, 5)

	g.Eventually(func() int {
		num_routes := len(nodes[0].NodeStaticRoutes)
		return num_routes
	}, 10*time.Second).Should(gomega.Equal(3))
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVRF()

	g.Eventually(func() bool {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVRF()
		if len(nodes[0].StaticRoutes) < 4 {
			found = false
		}
		return found
	}, 10*time.Second).Should(gomega.Equal(true))

	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	g.Expect(aviModel.(*avinodes.AviObjectGraph).IsVrf).To(gomega.Equal(true))
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVRF()
	g.Expect(len(nodes)).To(gomega.Equal(1))
	g.Expect(len(nodes[0].StaticRoutes)).To(gomega.Equal(4))
	g.Expect(len(nodes[0].NodeStaticRoutes)).To(gomega.Equal(3))
	g.Expect(*nodes[0].StaticRoutes[0].RouteID).To(gomega.Equal("cluster-" + nodes[0].NodeStaticRoutes["testNodeCalico1"].RouteIDPrefix + "-0"))
	g.Expect(*nodes[0].StaticRoutes[1].RouteID).To(gomega.Equal("cluster-" + nodes[0].NodeStaticRoutes["testNodeCalico1"].RouteIDPrefix + "-1"))
	g.Expect(*nodes[0].StaticRoutes[2].RouteID).To(gomega.Equal("cluster-" + nodes[0].NodeStaticRoutes["testNodeCalico3"].RouteIDPrefix + "-0"))
	g.Expect(*nodes[0].StaticRoutes[3].RouteID).To(gomega.Equal("cluster-" + nodes[0].NodeStaticRoutes["testNodeCalico4"].RouteIDPrefix + "-0"))

	var banode4 unstructured.Unstructured
	banode4.SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "crd.projectcalico.org/v1",
		"kind":       "blockaffinities",
		"metadata": map[string]interface{}{
			"name":      "testblockaffinity44",
			"namespace": "default",
		},
		"spec": map[string]interface{}{
			"cidr":    "10.244.127.128/25",
			"deleted": "false",
			"node":    nodesNameList[3],
			"state":   "confirmed",
		},
	})

	DynamicClient.Resource(lib.CalicoBlockaffinityGVR).Namespace("default").Create(context.TODO(), &banode4, v1.CreateOptions{})

	g.Eventually(func() bool {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVRF()
		if len(nodes[0].StaticRoutes) < 5 {
			found = false
		}
		return found
	}, 10*time.Second).Should(gomega.Equal(true))

	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	g.Expect(aviModel.(*avinodes.AviObjectGraph).IsVrf).To(gomega.Equal(true))
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVRF()
	g.Expect(len(nodes)).To(gomega.Equal(1))

	g.Expect(len(nodes)).To(gomega.Equal(1))
	g.Expect(len(nodes[0].StaticRoutes)).To(gomega.Equal(5))
	g.Expect(len(nodes[0].NodeStaticRoutes)).To(gomega.Equal(3))
	g.Expect(*nodes[0].StaticRoutes[0].RouteID).To(gomega.Equal("cluster-" + nodes[0].NodeStaticRoutes["testNodeCalico1"].RouteIDPrefix + "-0"))
	g.Expect(*nodes[0].StaticRoutes[1].RouteID).To(gomega.Equal("cluster-" + nodes[0].NodeStaticRoutes["testNodeCalico1"].RouteIDPrefix + "-1"))
	g.Expect(*nodes[0].StaticRoutes[2].RouteID).To(gomega.Equal("cluster-" + nodes[0].NodeStaticRoutes["testNodeCalico3"].RouteIDPrefix + "-0"))
	g.Expect(*nodes[0].StaticRoutes[3].RouteID).To(gomega.Equal("cluster-" + nodes[0].NodeStaticRoutes["testNodeCalico4"].RouteIDPrefix + "-0"))
	g.Expect(*nodes[0].StaticRoutes[4].RouteID).To(gomega.Equal("cluster-" + nodes[0].NodeStaticRoutes["testNodeCalico4"].RouteIDPrefix + "-1"))

	DynamicClient.Resource(lib.CalicoBlockaffinityGVR).Namespace("default").Create(context.TODO(), &testData2, v1.CreateOptions{})
	g.Eventually(func() bool {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVRF()
		if len(nodes[0].StaticRoutes) < 6 {
			found = false
		}
		return found
	}, 10*time.Second).Should(gomega.Equal(true))

	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	g.Expect(aviModel.(*avinodes.AviObjectGraph).IsVrf).To(gomega.Equal(true))
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVRF()
	g.Expect(len(nodes)).To(gomega.Equal(1))
	g.Expect(len(nodes[0].StaticRoutes)).To(gomega.Equal(6))
	g.Expect(len(nodes[0].NodeStaticRoutes)).To(gomega.Equal(4))
	g.Expect(*nodes[0].StaticRoutes[0].RouteID).To(gomega.Equal("cluster-" + nodes[0].NodeStaticRoutes["testNodeCalico1"].RouteIDPrefix + "-0"))
	g.Expect(*nodes[0].StaticRoutes[1].RouteID).To(gomega.Equal("cluster-" + nodes[0].NodeStaticRoutes["testNodeCalico1"].RouteIDPrefix + "-1"))
	g.Expect(*nodes[0].StaticRoutes[2].RouteID).To(gomega.Equal("cluster-" + nodes[0].NodeStaticRoutes["testNodeCalico3"].RouteIDPrefix + "-0"))
	g.Expect(*nodes[0].StaticRoutes[3].RouteID).To(gomega.Equal("cluster-" + nodes[0].NodeStaticRoutes["testNodeCalico4"].RouteIDPrefix + "-0"))
	g.Expect(*nodes[0].StaticRoutes[4].RouteID).To(gomega.Equal("cluster-" + nodes[0].NodeStaticRoutes["testNodeCalico4"].RouteIDPrefix + "-1"))
	g.Expect(*nodes[0].StaticRoutes[5].RouteID).To(gomega.Equal("cluster-" + nodes[0].NodeStaticRoutes["testNodeCalico2"].RouteIDPrefix + "-0"))

	DynamicClient.Resource(lib.CalicoBlockaffinityGVR).Namespace("default").Delete(context.TODO(), "testblockaffinity2", v1.DeleteOptions{})
	g.Eventually(func() bool {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVRF()
		if len(nodes[0].StaticRoutes) > 5 {
			found = false
		}
		return found
	}, 10*time.Second).Should(gomega.Equal(true))
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVRF()

	KubeClient.CoreV1().Nodes().Delete(context.TODO(), nodesNameList[1], metav1.DeleteOptions{})
	g.Eventually(func() bool {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVRF()
		if len(nodes[0].NodeStaticRoutes) > 3 {
			found = false
		}
		return found
	}, 10*time.Second).Should(gomega.Equal(true))
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVRF()

	g.Expect(len(nodes)).To(gomega.Equal(1))
	g.Expect(len(nodes[0].StaticRoutes)).To(gomega.Equal(5))
	g.Expect(len(nodes[0].NodeStaticRoutes)).To(gomega.Equal(3))
	g.Expect(*nodes[0].StaticRoutes[0].RouteID).To(gomega.Equal("cluster-" + nodes[0].NodeStaticRoutes["testNodeCalico1"].RouteIDPrefix + "-0"))
	g.Expect(*nodes[0].StaticRoutes[1].RouteID).To(gomega.Equal("cluster-" + nodes[0].NodeStaticRoutes["testNodeCalico1"].RouteIDPrefix + "-1"))
	g.Expect(*nodes[0].StaticRoutes[2].RouteID).To(gomega.Equal("cluster-" + nodes[0].NodeStaticRoutes["testNodeCalico3"].RouteIDPrefix + "-0"))
	g.Expect(*nodes[0].StaticRoutes[3].RouteID).To(gomega.Equal("cluster-" + nodes[0].NodeStaticRoutes["testNodeCalico4"].RouteIDPrefix + "-0"))
	g.Expect(*nodes[0].StaticRoutes[4].RouteID).To(gomega.Equal("cluster-" + nodes[0].NodeStaticRoutes["testNodeCalico4"].RouteIDPrefix + "-1"))

	_, err = KubeClient.CoreV1().Nodes().Create(context.TODO(), nodeExample2, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Node: %v", err)
	}

	integrationtest.PollForCompletion(t, modelName, 5)
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVRF()

	g.Eventually(func() int {
		num_routes := len(nodes[0].NodeStaticRoutes)
		return num_routes
	}, 10*time.Second).Should(gomega.Equal(3))
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVRF()
	g.Expect(len(nodes)).To(gomega.Equal(1))
	g.Expect(len(nodes[0].StaticRoutes)).To(gomega.Equal(5))
	g.Expect(len(nodes[0].NodeStaticRoutes)).To(gomega.Equal(3))
	g.Expect(*nodes[0].StaticRoutes[0].RouteID).To(gomega.Equal("cluster-" + nodes[0].NodeStaticRoutes["testNodeCalico1"].RouteIDPrefix + "-0"))
	g.Expect(*nodes[0].StaticRoutes[1].RouteID).To(gomega.Equal("cluster-" + nodes[0].NodeStaticRoutes["testNodeCalico1"].RouteIDPrefix + "-1"))
	g.Expect(*nodes[0].StaticRoutes[2].RouteID).To(gomega.Equal("cluster-" + nodes[0].NodeStaticRoutes["testNodeCalico3"].RouteIDPrefix + "-0"))
	g.Expect(*nodes[0].StaticRoutes[3].RouteID).To(gomega.Equal("cluster-" + nodes[0].NodeStaticRoutes["testNodeCalico4"].RouteIDPrefix + "-0"))
	g.Expect(*nodes[0].StaticRoutes[4].RouteID).To(gomega.Equal("cluster-" + nodes[0].NodeStaticRoutes["testNodeCalico4"].RouteIDPrefix + "-1"))

	DynamicClient.Resource(lib.CalicoBlockaffinityGVR).Namespace("default").Delete(context.TODO(), "testblockaffinity11", v1.DeleteOptions{})
	DynamicClient.Resource(lib.CalicoBlockaffinityGVR).Namespace("default").Delete(context.TODO(), "testblockaffinity1", v1.DeleteOptions{})
	g.Eventually(func() bool {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVRF()
		if len(nodes[0].StaticRoutes) > 3 {
			found = false
		}
		return found
	}, 10*time.Second).Should(gomega.Equal(true))
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVRF()
	KubeClient.CoreV1().Nodes().Delete(context.TODO(), nodesNameList[0], metav1.DeleteOptions{})
	g.Eventually(func() int {
		num_routes := len(nodes[0].NodeStaticRoutes)
		return num_routes
	}, 10*time.Second).Should(gomega.Equal(2))
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVRF()
	g.Expect(len(nodes[0].StaticRoutes)).To(gomega.Equal(3))
	g.Expect(len(nodes[0].NodeStaticRoutes)).To(gomega.Equal(2))
	g.Expect(len(nodes[0].Nodes)).To(gomega.Equal(2))
	g.Expect(*nodes[0].StaticRoutes[0].RouteID).To(gomega.Equal("cluster-" + nodes[0].NodeStaticRoutes["testNodeCalico3"].RouteIDPrefix + "-0"))
	g.Expect(*nodes[0].StaticRoutes[1].RouteID).To(gomega.Equal("cluster-" + nodes[0].NodeStaticRoutes["testNodeCalico4"].RouteIDPrefix + "-0"))
	g.Expect(*nodes[0].StaticRoutes[2].RouteID).To(gomega.Equal("cluster-" + nodes[0].NodeStaticRoutes["testNodeCalico4"].RouteIDPrefix + "-1"))

	KubeClient.CoreV1().Nodes().Delete(context.TODO(), nodesNameList[1], metav1.DeleteOptions{})
	g.Eventually(func() bool {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVRF()
		if len(nodes[0].StaticRoutes) > 3 {
			found = false
		}
		return found
	}, 10*time.Second).Should(gomega.Equal(true))
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVRF()

	g.Eventually(func() int {
		num_routes := len(nodes[0].NodeStaticRoutes)
		return num_routes
	}, 10*time.Second).Should(gomega.Equal(2))
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVRF()
	g.Expect(len(nodes[0].StaticRoutes)).To(gomega.Equal(3))
	g.Expect(len(nodes[0].NodeStaticRoutes)).To(gomega.Equal(2))
	g.Expect(len(nodes[0].Nodes)).To(gomega.Equal(2))
	g.Expect(*nodes[0].StaticRoutes[0].RouteID).To(gomega.Equal("cluster-" + nodes[0].NodeStaticRoutes["testNodeCalico3"].RouteIDPrefix + "-0"))
	g.Expect(*nodes[0].StaticRoutes[1].RouteID).To(gomega.Equal("cluster-" + nodes[0].NodeStaticRoutes["testNodeCalico4"].RouteIDPrefix + "-0"))
	g.Expect(*nodes[0].StaticRoutes[2].RouteID).To(gomega.Equal("cluster-" + nodes[0].NodeStaticRoutes["testNodeCalico4"].RouteIDPrefix + "-1"))

	DynamicClient.Resource(lib.CalicoBlockaffinityGVR).Namespace("default").Delete(context.TODO(), "testblockaffinity3", v1.DeleteOptions{})

	g.Eventually(func() bool {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVRF()
		if len(nodes[0].StaticRoutes) > 2 {
			found = false
		}
		return found
	}, 10*time.Second).Should(gomega.Equal(true))
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVRF()

	KubeClient.CoreV1().Nodes().Delete(context.TODO(), nodesNameList[2], metav1.DeleteOptions{})
	g.Eventually(func() int {
		num_routes := len(nodes[0].NodeStaticRoutes)
		return num_routes
	}, 10*time.Second).Should(gomega.Equal(1))
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVRF()
	g.Expect(len(nodes[0].StaticRoutes)).To(gomega.Equal(2))
	g.Expect(len(nodes[0].NodeStaticRoutes)).To(gomega.Equal(1))
	g.Expect(len(nodes[0].Nodes)).To(gomega.Equal(1))
	g.Expect(*nodes[0].StaticRoutes[0].RouteID).To(gomega.Equal("cluster-" + nodes[0].NodeStaticRoutes["testNodeCalico4"].RouteIDPrefix + "-0"))
	g.Expect(*nodes[0].StaticRoutes[1].RouteID).To(gomega.Equal("cluster-" + nodes[0].NodeStaticRoutes["testNodeCalico4"].RouteIDPrefix + "-1"))

	DynamicClient.Resource(lib.CalicoBlockaffinityGVR).Namespace("default").Delete(context.TODO(), "testblockaffinity44", v1.DeleteOptions{})
	DynamicClient.Resource(lib.CalicoBlockaffinityGVR).Namespace("default").Delete(context.TODO(), "testblockaffinity4", v1.DeleteOptions{})

	g.Eventually(func() bool {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVRF()
		if len(nodes[0].StaticRoutes) > 0 {
			found = false
		}
		return found
	}, 10*time.Second).Should(gomega.Equal(true))
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVRF()

	KubeClient.CoreV1().Nodes().Delete(context.TODO(), nodesNameList[3], metav1.DeleteOptions{})

	g.Eventually(func() int {
		num_routes := len(nodes[0].NodeStaticRoutes)
		return num_routes
	}, 10*time.Second).Should(gomega.Equal(0))
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVRF()
	g.Expect(len(nodes[0].StaticRoutes)).To(gomega.Equal(0))
	g.Expect(len(nodes[0].NodeStaticRoutes)).To(gomega.Equal(0))
	g.Expect(len(nodes[0].Nodes)).To(gomega.Equal(0))
}
func TestStaticRoutesRecordsDeduplication(t *testing.T) {
	if *cniPlugin != "calico" {
		t.Skip("Skipping BlockAffinity test since CNI plugin is not Calico")
	}
	g := gomega.NewGomegaWithT(t)
	modelName := "admin/global"
	nodeip1 := "10.102.99.232"
	nodeName1 := "testNodeCalico1"
	objects.SharedAviGraphLister().Delete(modelName)
	time.Sleep(10 * time.Second)

	// mimicking actual scenario where the node will have atleast one BlockAffinity object created from start
	var testData1 unstructured.Unstructured
	testData1.SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "crd.projectcalico.org/v1",
		"kind":       "blockaffinities",
		"metadata": map[string]interface{}{
			"name":      "testblockaffinity1",
			"namespace": "default",
		},
		"spec": map[string]interface{}{
			"cidr":    "192.168.29.64/26",
			"deleted": "false",
			"node":    nodeName1,
			"state":   "confirmed",
		},
	})

	DynamicClient.Resource(lib.CalicoBlockaffinityGVR).Namespace("default").Create(context.TODO(), &testData1, v1.CreateOptions{})

	nodeExample := (integrationtest.FakeNode{
		Name:    nodeName1,
		Version: "1",
		NodeIP:  nodeip1,
	}).NodeCalico()

	_, err := KubeClient.CoreV1().Nodes().Create(context.TODO(), nodeExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Node: %v", err)
	}

	// creating a new BlockAffinity object for the node
	var testData2 unstructured.Unstructured
	testData2.SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "crd.projectcalico.org/v1",
		"kind":       "blockaffinities",
		"metadata": map[string]interface{}{
			"name":      "testblockaffinity2",
			"namespace": "default",
		},
		"spec": map[string]interface{}{
			"cidr":    "192.168.30.64/26",
			"deleted": "false",
			"node":    nodeName1,
			"state":   "confirmed",
		},
	})

	DynamicClient.Resource(lib.CalicoBlockaffinityGVR).Namespace("default").Create(context.TODO(), &testData2, v1.CreateOptions{})

	nodeip2 := "10.102.99.146"
	nodeName2 := "testNodeCalico2"

	// mimicking actual scenario where the node will have atleast one BlockAffinity object created from start
	var testData3 unstructured.Unstructured
	testData3.SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "crd.projectcalico.org/v1",
		"kind":       "blockaffinities",
		"metadata": map[string]interface{}{
			"name":      "testblockaffinity3",
			"namespace": "default",
		},
		"spec": map[string]interface{}{
			"cidr":    "192.168.247.0/26",
			"deleted": "false",
			"node":    nodeName2,
			"state":   "confirmed",
		},
	})

	DynamicClient.Resource(lib.CalicoBlockaffinityGVR).Namespace("default").Create(context.TODO(), &testData3, v1.CreateOptions{})

	nodeExample2 := (integrationtest.FakeNode{
		Name:    nodeName2,
		Version: "1",
		NodeIP:  nodeip2,
	}).NodeCalico()

	_, err = KubeClient.CoreV1().Nodes().Create(context.TODO(), nodeExample2, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Node: %v", err)
	}

	integrationtest.PollForCompletion(t, modelName, 5)

	// creating a new BlockAffinity object for the node
	var testData4 unstructured.Unstructured
	testData4.SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "crd.projectcalico.org/v1",
		"kind":       "blockaffinities",
		"metadata": map[string]interface{}{
			"name":      "testblockaffinity4",
			"namespace": "default",
		},
		"spec": map[string]interface{}{
			"cidr":    "192.168.246.0/26",
			"deleted": "false",
			"node":    nodeName2,
			"state":   "confirmed",
		},
	})

	DynamicClient.Resource(lib.CalicoBlockaffinityGVR).Namespace("default").Create(context.TODO(), &testData4, v1.CreateOptions{})

	g.Eventually(func() bool {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVRF()
		if len(nodes[0].StaticRoutes) < 4 {
			found = false
		}
		return found
	}, 10*time.Second).Should(gomega.Equal(true))

	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	g.Expect(aviModel.(*avinodes.AviObjectGraph).IsVrf).To(gomega.Equal(true))
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVRF()
	g.Expect(len(nodes)).To(gomega.Equal(1))

	g.Expect(len(nodes[0].StaticRoutes)).To(gomega.Equal(4))
	g.Expect(*(nodes[0].StaticRoutes[3].NextHop.Addr)).To(gomega.Equal(nodeip2))
	g.Expect(*(nodes[0].StaticRoutes[3].Prefix.IPAddr.Addr)).To(gomega.Equal("192.168.246.0"))
	g.Expect(*(nodes[0].StaticRoutes[3].Prefix.Mask)).To(gomega.Equal(int32(26)))

	//Manipulate internal cache to mimic duplicate records
	nodes[0].StaticRoutes[3] = nodes[0].StaticRoutes[2]

	var testData5 unstructured.Unstructured
	testData5.SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "crd.projectcalico.org/v1",
		"kind":       "blockaffinities",
		"metadata": map[string]interface{}{
			"name":      "testblockaffinity5",
			"namespace": "default",
		},
		"spec": map[string]interface{}{
			"cidr":    "192.168.31.64/26",
			"deleted": "false",
			"node":    nodeName1,
			"state":   "confirmed",
		},
	})

	DynamicClient.Resource(lib.CalicoBlockaffinityGVR).Namespace("default").Create(context.TODO(), &testData5, v1.CreateOptions{})

	g.Eventually(func() bool {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVRF()
		if len(nodes[0].StaticRoutes) < 5 {
			found = false
		}
		return found
	}, 10*time.Second).Should(gomega.Equal(true))

	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	g.Expect(aviModel.(*avinodes.AviObjectGraph).IsVrf).To(gomega.Equal(true))
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVRF()
	g.Expect(len(nodes)).To(gomega.Equal(1))

	g.Expect(len(nodes[0].StaticRoutes)).To(gomega.Equal(5))

	g.Expect(*nodes[0].StaticRoutes[0].RouteID).To(gomega.Equal("cluster-" + nodes[0].NodeStaticRoutes["testNodeCalico1"].RouteIDPrefix + "-0"))
	g.Expect(*nodes[0].StaticRoutes[1].RouteID).To(gomega.Equal("cluster-" + nodes[0].NodeStaticRoutes["testNodeCalico1"].RouteIDPrefix + "-1"))
	g.Expect(*nodes[0].StaticRoutes[2].RouteID).To(gomega.Equal("cluster-" + nodes[0].NodeStaticRoutes["testNodeCalico1"].RouteIDPrefix + "-2"))
	g.Expect(*nodes[0].StaticRoutes[3].RouteID).To(gomega.Equal("cluster-" + nodes[0].NodeStaticRoutes["testNodeCalico2"].RouteIDPrefix + "-0"))
	g.Expect(*nodes[0].StaticRoutes[4].RouteID).To(gomega.Equal("cluster-" + nodes[0].NodeStaticRoutes["testNodeCalico2"].RouteIDPrefix + "-1"))
	g.Expect(*(nodes[0].StaticRoutes[0].NextHop.Addr)).To(gomega.Equal(nodeip1))
	g.Expect(*(nodes[0].StaticRoutes[0].Prefix.IPAddr.Addr)).To(gomega.Equal("192.168.29.64"))
	g.Expect(*(nodes[0].StaticRoutes[0].Prefix.Mask)).To(gomega.Equal(int32(26)))
	g.Expect(*(nodes[0].StaticRoutes[1].NextHop.Addr)).To(gomega.Equal(nodeip1))
	g.Expect(*(nodes[0].StaticRoutes[1].Prefix.IPAddr.Addr)).To(gomega.Equal("192.168.30.64"))
	g.Expect(*(nodes[0].StaticRoutes[1].Prefix.Mask)).To(gomega.Equal(int32(26)))
	g.Expect(*(nodes[0].StaticRoutes[2].NextHop.Addr)).To(gomega.Equal(nodeip1))
	g.Expect(*(nodes[0].StaticRoutes[2].Prefix.IPAddr.Addr)).To(gomega.Equal("192.168.31.64"))
	g.Expect(*(nodes[0].StaticRoutes[2].Prefix.Mask)).To(gomega.Equal(int32(26)))
	g.Expect(*(nodes[0].StaticRoutes[3].NextHop.Addr)).To(gomega.Equal(nodeip2))
	g.Expect(*(nodes[0].StaticRoutes[3].Prefix.IPAddr.Addr)).To(gomega.Equal("192.168.247.0"))
	g.Expect(*(nodes[0].StaticRoutes[3].Prefix.Mask)).To(gomega.Equal(int32(26)))
	g.Expect(*(nodes[0].StaticRoutes[4].NextHop.Addr)).To(gomega.Equal(nodeip2))
	g.Expect(*(nodes[0].StaticRoutes[4].Prefix.IPAddr.Addr)).To(gomega.Equal("192.168.246.0"))
	g.Expect(*(nodes[0].StaticRoutes[4].Prefix.Mask)).To(gomega.Equal(int32(26)))
	time.Sleep(10 * time.Second)

	// deleting the BlockAffinity objects for the node
	DynamicClient.Resource(lib.CalicoBlockaffinityGVR).Namespace("default").Delete(context.TODO(), "testblockaffinity2", v1.DeleteOptions{})
	DynamicClient.Resource(lib.CalicoBlockaffinityGVR).Namespace("default").Delete(context.TODO(), "testblockaffinity5", v1.DeleteOptions{})
	g.Eventually(func() int {
		num_routes := len(nodes[0].StaticRoutes)
		return num_routes
	}, 10*time.Second).Should(gomega.Equal(3))
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVRF()
	g.Expect(*nodes[0].StaticRoutes[0].RouteID).To(gomega.Equal("cluster-" + nodes[0].NodeStaticRoutes["testNodeCalico2"].RouteIDPrefix + "-0"))
	g.Expect(*nodes[0].StaticRoutes[1].RouteID).To(gomega.Equal("cluster-" + nodes[0].NodeStaticRoutes["testNodeCalico2"].RouteIDPrefix + "-1"))
	g.Expect(*nodes[0].StaticRoutes[2].RouteID).To(gomega.Equal("cluster-" + nodes[0].NodeStaticRoutes["testNodeCalico1"].RouteIDPrefix + "-0"))

	KubeClient.CoreV1().Nodes().Delete(context.TODO(), nodeName1, metav1.DeleteOptions{})
	DynamicClient.Resource(lib.CalicoBlockaffinityGVR).Namespace("default").Delete(context.TODO(), "testblockaffinity1", v1.DeleteOptions{})
	g.Eventually(func() int {
		num_routes := len(nodes[0].StaticRoutes)
		return num_routes
	}, 10*time.Second).Should(gomega.Equal(2))
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVRF()
	g.Expect(*nodes[0].StaticRoutes[0].RouteID).To(gomega.Equal("cluster-" + nodes[0].NodeStaticRoutes["testNodeCalico2"].RouteIDPrefix + "-0"))
	g.Expect(*nodes[0].StaticRoutes[1].RouteID).To(gomega.Equal("cluster-" + nodes[0].NodeStaticRoutes["testNodeCalico2"].RouteIDPrefix + "-1"))

	DynamicClient.Resource(lib.CalicoBlockaffinityGVR).Namespace("default").Delete(context.TODO(), "testblockaffinity4", v1.DeleteOptions{})
	g.Eventually(func() int {
		num_routes := len(nodes[0].StaticRoutes)
		return num_routes
	}, 10*time.Second).Should(gomega.Equal(1))
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVRF()
	g.Expect(*nodes[0].StaticRoutes[0].RouteID).To(gomega.Equal("cluster-" + nodes[0].NodeStaticRoutes["testNodeCalico2"].RouteIDPrefix + "-0"))

	KubeClient.CoreV1().Nodes().Delete(context.TODO(), nodeName2, metav1.DeleteOptions{})
	DynamicClient.Resource(lib.CalicoBlockaffinityGVR).Namespace("default").Delete(context.TODO(), "testblockaffinity3", v1.DeleteOptions{})
	g.Eventually(func() int {
		num_routes := len(nodes[0].StaticRoutes)
		return num_routes
	}, 10*time.Second).Should(gomega.Equal(0))
}
func TestStaticRoutesWithIncompleteNodeData(t *testing.T) {
	if *cniPlugin != "calico" {
		t.Skip("Skipping BlockAffinity test since CNI plugin is not Calico")
	}
	g := gomega.NewGomegaWithT(t)
	modelName := "admin/global"
	nodeip1 := "10.102.99.232"
	nodeName1 := "testNodeCalico1"
	objects.SharedAviGraphLister().Delete(modelName)
	time.Sleep(10 * time.Second)

	// mimicking actual scenario where the node will have atleast one BlockAffinity object created from start
	var testData1 unstructured.Unstructured
	testData1.SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "crd.projectcalico.org/v1",
		"kind":       "blockaffinities",
		"metadata": map[string]interface{}{
			"name":      "testblockaffinity1",
			"namespace": "default",
		},
		"spec": map[string]interface{}{
			"cidr":    "192.168.29.64/26",
			"deleted": "false",
			"node":    nodeName1,
			"state":   "confirmed",
		},
	})

	DynamicClient.Resource(lib.CalicoBlockaffinityGVR).Namespace("default").Create(context.TODO(), &testData1, v1.CreateOptions{})

	nodeExample := (integrationtest.FakeNode{
		Name:    nodeName1,
		Version: "1",
		NodeIP:  nodeip1,
	}).NodeCalico()

	_, err := KubeClient.CoreV1().Nodes().Create(context.TODO(), nodeExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Node: %v", err)
	}

	// creating a new BlockAffinity object for the node
	var testData2 unstructured.Unstructured
	testData2.SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "crd.projectcalico.org/v1",
		"kind":       "blockaffinities",
		"metadata": map[string]interface{}{
			"name":      "testblockaffinity2",
			"namespace": "default",
		},
		"spec": map[string]interface{}{
			"cidr":    "192.168.30.64/26",
			"deleted": "false",
			"node":    nodeName1,
			"state":   "confirmed",
		},
	})

	DynamicClient.Resource(lib.CalicoBlockaffinityGVR).Namespace("default").Create(context.TODO(), &testData2, v1.CreateOptions{})

	nodeip2 := "10.102.99.146"
	nodeName2 := "testNodeCalico2"

	// mimicking actual scenario where the node will have atleast one BlockAffinity object created from start
	var testData3 unstructured.Unstructured
	testData3.SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "crd.projectcalico.org/v1",
		"kind":       "blockaffinities",
		"metadata": map[string]interface{}{
			"name":      "testblockaffinity3",
			"namespace": "default",
		},
		"spec": map[string]interface{}{
			"cidr":    "192.168.247.0/26",
			"deleted": "false",
			"node":    nodeName2,
			"state":   "confirmed",
		},
	})

	DynamicClient.Resource(lib.CalicoBlockaffinityGVR).Namespace("default").Create(context.TODO(), &testData3, v1.CreateOptions{})

	nodeExample2 := (integrationtest.FakeNode{
		Name:    nodeName2,
		Version: "1",
		NodeIP:  nodeip2,
	}).NodeCalico()

	_, err = KubeClient.CoreV1().Nodes().Create(context.TODO(), nodeExample2, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Node: %v", err)
	}

	integrationtest.PollForCompletion(t, modelName, 5)

	// creating a new BlockAffinity object for the node
	var testData4 unstructured.Unstructured
	testData4.SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "crd.projectcalico.org/v1",
		"kind":       "blockaffinities",
		"metadata": map[string]interface{}{
			"name":      "testblockaffinity4",
			"namespace": "default",
		},
		"spec": map[string]interface{}{
			"cidr":    "192.168.246.0/26",
			"deleted": "false",
			"node":    nodeName2,
			"state":   "confirmed",
		},
	})

	DynamicClient.Resource(lib.CalicoBlockaffinityGVR).Namespace("default").Create(context.TODO(), &testData4, v1.CreateOptions{})

	g.Eventually(func() bool {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVRF()
		if len(nodes[0].StaticRoutes) < 4 {
			found = false
		}
		return found
	}, 10*time.Second).Should(gomega.Equal(true))

	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	g.Expect(aviModel.(*avinodes.AviObjectGraph).IsVrf).To(gomega.Equal(true))
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVRF()
	g.Expect(len(nodes)).To(gomega.Equal(1))

	g.Expect(*nodes[0].StaticRoutes[0].RouteID).To(gomega.Equal("cluster-" + nodes[0].NodeStaticRoutes["testNodeCalico1"].RouteIDPrefix + "-0"))
	g.Expect(*nodes[0].StaticRoutes[1].RouteID).To(gomega.Equal("cluster-" + nodes[0].NodeStaticRoutes["testNodeCalico1"].RouteIDPrefix + "-1"))
	g.Expect(*nodes[0].StaticRoutes[2].RouteID).To(gomega.Equal("cluster-" + nodes[0].NodeStaticRoutes["testNodeCalico2"].RouteIDPrefix + "-0"))
	g.Expect(*nodes[0].StaticRoutes[3].RouteID).To(gomega.Equal("cluster-" + nodes[0].NodeStaticRoutes["testNodeCalico2"].RouteIDPrefix + "-1"))
	g.Expect(*(nodes[0].StaticRoutes[0].NextHop.Addr)).To(gomega.Equal(nodeip1))
	g.Expect(*(nodes[0].StaticRoutes[0].Prefix.IPAddr.Addr)).To(gomega.Equal("192.168.29.64"))
	g.Expect(*(nodes[0].StaticRoutes[0].Prefix.Mask)).To(gomega.Equal(int32(26)))
	g.Expect(*(nodes[0].StaticRoutes[1].NextHop.Addr)).To(gomega.Equal(nodeip1))
	g.Expect(*(nodes[0].StaticRoutes[1].Prefix.IPAddr.Addr)).To(gomega.Equal("192.168.30.64"))
	g.Expect(*(nodes[0].StaticRoutes[1].Prefix.Mask)).To(gomega.Equal(int32(26)))
	g.Expect(*(nodes[0].StaticRoutes[2].NextHop.Addr)).To(gomega.Equal(nodeip2))
	g.Expect(*(nodes[0].StaticRoutes[2].Prefix.IPAddr.Addr)).To(gomega.Equal("192.168.247.0"))
	g.Expect(*(nodes[0].StaticRoutes[2].Prefix.Mask)).To(gomega.Equal(int32(26)))
	g.Expect(*(nodes[0].StaticRoutes[3].NextHop.Addr)).To(gomega.Equal(nodeip2))
	g.Expect(*(nodes[0].StaticRoutes[3].Prefix.IPAddr.Addr)).To(gomega.Equal("192.168.246.0"))
	g.Expect(*(nodes[0].StaticRoutes[3].Prefix.Mask)).To(gomega.Equal(int32(26)))
	time.Sleep(10 * time.Second)

	// Update node data to not have any nodeIP
	nodeExample = (integrationtest.FakeNode{
		Name:    nodeName1,
		Version: "1",
	}).NodeCalico()

	KubeClient.CoreV1().Nodes().Update(context.TODO(), nodeExample, metav1.UpdateOptions{})

	g.Eventually(func() bool {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVRF()
		if len(nodes[0].StaticRoutes) > 2 {
			found = false
		}
		return found
	}, 10*time.Second).Should(gomega.Equal(true))

	// Adding NodeIp back
	nodeExample = (integrationtest.FakeNode{
		Name:    nodeName1,
		Version: "1",
		NodeIP:  nodeip1,
	}).NodeCalico()

	KubeClient.CoreV1().Nodes().Update(context.TODO(), nodeExample, metav1.UpdateOptions{})

	g.Eventually(func() bool {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVRF()
		if len(nodes[0].StaticRoutes) != 4 {
			found = false
		}
		return found
	}, 10*time.Second).Should(gomega.Equal(true))

	DynamicClient.Resource(lib.CalicoBlockaffinityGVR).Namespace("default").Delete(context.TODO(), "testblockaffinity3", v1.DeleteOptions{})
	g.Eventually(func() int {
		num_routes := len(nodes[0].StaticRoutes)
		return num_routes
	}, 10*time.Second).Should(gomega.Equal(3))

	// Updating PODCidr to not have any IP
	testData3.SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "crd.projectcalico.org/v1",
		"kind":       "blockaffinities",
		"metadata": map[string]interface{}{
			"name":      "testblockaffinity3",
			"namespace": "default",
		},
		"spec": map[string]interface{}{
			"cidr":    "",
			"deleted": "false",
			"node":    nodeName2,
			"state":   "confirmed",
		},
	})

	DynamicClient.Resource(lib.CalicoBlockaffinityGVR).Namespace("default").Create(context.TODO(), &testData3, v1.CreateOptions{})

	g.Eventually(func() bool {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVRF()
		if len(nodes[0].StaticRoutes) != 2 {
			found = false
		}
		return found
	}, 10*time.Second).Should(gomega.Equal(true))

	DynamicClient.Resource(lib.CalicoBlockaffinityGVR).Namespace("default").Delete(context.TODO(), "testblockaffinity3", v1.DeleteOptions{})
	g.Eventually(func() int {
		num_routes := len(nodes[0].StaticRoutes)
		return num_routes
	}, 10*time.Second).Should(gomega.Equal(3))

	// Updating PODCidr with IP
	testData3.SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "crd.projectcalico.org/v1",
		"kind":       "blockaffinities",
		"metadata": map[string]interface{}{
			"name":      "testblockaffinity3",
			"namespace": "default",
		},
		"spec": map[string]interface{}{
			"cidr":    "192.168.247.0/26",
			"deleted": "false",
			"node":    nodeName2,
			"state":   "confirmed",
		},
	})

	DynamicClient.Resource(lib.CalicoBlockaffinityGVR).Namespace("default").Create(context.TODO(), &testData3, v1.CreateOptions{})

	g.Eventually(func() bool {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVRF()
		if len(nodes[0].StaticRoutes) != 4 {
			found = false
		}
		return found
	}, 10*time.Second).Should(gomega.Equal(true))

	DynamicClient.Resource(lib.CalicoBlockaffinityGVR).Namespace("default").Delete(context.TODO(), "testblockaffinity3", v1.DeleteOptions{})
	g.Eventually(func() int {
		num_routes := len(nodes[0].StaticRoutes)
		return num_routes
	}, 10*time.Second).Should(gomega.Equal(3))

	// Updating PODCidr to have malformed IP
	testData3.SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "crd.projectcalico.org/v1",
		"kind":       "blockaffinities",
		"metadata": map[string]interface{}{
			"name":      "testblockaffinity3",
			"namespace": "default",
		},
		"spec": map[string]interface{}{
			"cidr":    "malformedIp",
			"deleted": "false",
			"node":    nodeName2,
			"state":   "confirmed",
		},
	})

	DynamicClient.Resource(lib.CalicoBlockaffinityGVR).Namespace("default").Create(context.TODO(), &testData3, v1.CreateOptions{})

	g.Eventually(func() bool {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVRF()
		if len(nodes[0].StaticRoutes) != 2 {
			found = false
		}
		return found
	}, 10*time.Second).Should(gomega.Equal(true))

	DynamicClient.Resource(lib.CalicoBlockaffinityGVR).Namespace("default").Delete(context.TODO(), "testblockaffinity3", v1.DeleteOptions{})
	g.Eventually(func() int {
		num_routes := len(nodes[0].StaticRoutes)
		return num_routes
	}, 10*time.Second).Should(gomega.Equal(3))

	// Updating PODCidr with IP
	testData3.SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "crd.projectcalico.org/v1",
		"kind":       "blockaffinities",
		"metadata": map[string]interface{}{
			"name":      "testblockaffinity3",
			"namespace": "default",
		},
		"spec": map[string]interface{}{
			"cidr":    "192.168.247.0/26",
			"deleted": "false",
			"node":    nodeName2,
			"state":   "confirmed",
		},
	})

	DynamicClient.Resource(lib.CalicoBlockaffinityGVR).Namespace("default").Create(context.TODO(), &testData3, v1.CreateOptions{})

	g.Eventually(func() bool {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVRF()
		if len(nodes[0].StaticRoutes) != 4 {
			found = false
		}
		return found
	}, 10*time.Second).Should(gomega.Equal(true))

	DynamicClient.Resource(lib.CalicoBlockaffinityGVR).Namespace("default").Delete(context.TODO(), "testblockaffinity3", v1.DeleteOptions{})
	g.Eventually(func() int {
		num_routes := len(nodes[0].StaticRoutes)
		return num_routes
	}, 10*time.Second).Should(gomega.Equal(3))

	// Updating PODCidr to have malformed mask
	testData3.SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "crd.projectcalico.org/v1",
		"kind":       "blockaffinities",
		"metadata": map[string]interface{}{
			"name":      "testblockaffinity3",
			"namespace": "default",
		},
		"spec": map[string]interface{}{
			"cidr":    "ip/nonintegermask",
			"deleted": "false",
			"node":    nodeName2,
			"state":   "confirmed",
		},
	})

	DynamicClient.Resource(lib.CalicoBlockaffinityGVR).Namespace("default").Create(context.TODO(), &testData3, v1.CreateOptions{})

	g.Eventually(func() bool {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVRF()
		if len(nodes[0].StaticRoutes) != 2 {
			found = false
		}
		return found
	}, 10*time.Second).Should(gomega.Equal(true))

	DynamicClient.Resource(lib.CalicoBlockaffinityGVR).Namespace("default").Delete(context.TODO(), "testblockaffinity3", v1.DeleteOptions{})
	g.Eventually(func() int {
		num_routes := len(nodes[0].StaticRoutes)
		return num_routes
	}, 10*time.Second).Should(gomega.Equal(3))

	// Updating PODCidr with IP
	testData3.SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "crd.projectcalico.org/v1",
		"kind":       "blockaffinities",
		"metadata": map[string]interface{}{
			"name":      "testblockaffinity3",
			"namespace": "default",
		},
		"spec": map[string]interface{}{
			"cidr":    "192.168.247.0/26",
			"deleted": "false",
			"node":    nodeName2,
			"state":   "confirmed",
		},
	})

	DynamicClient.Resource(lib.CalicoBlockaffinityGVR).Namespace("default").Create(context.TODO(), &testData3, v1.CreateOptions{})

	g.Eventually(func() bool {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVRF()
		if len(nodes[0].StaticRoutes) != 4 {
			found = false
		}
		return found
	}, 10*time.Second).Should(gomega.Equal(true))

	// deleting the BlockAffinity objects for the node
	DynamicClient.Resource(lib.CalicoBlockaffinityGVR).Namespace("default").Delete(context.TODO(), "testblockaffinity1", v1.DeleteOptions{})
	DynamicClient.Resource(lib.CalicoBlockaffinityGVR).Namespace("default").Delete(context.TODO(), "testblockaffinity2", v1.DeleteOptions{})
	DynamicClient.Resource(lib.CalicoBlockaffinityGVR).Namespace("default").Delete(context.TODO(), "testblockaffinity3", v1.DeleteOptions{})
	DynamicClient.Resource(lib.CalicoBlockaffinityGVR).Namespace("default").Delete(context.TODO(), "testblockaffinity4", v1.DeleteOptions{})

	KubeClient.CoreV1().Nodes().Delete(context.TODO(), nodeName1, metav1.DeleteOptions{})
	KubeClient.CoreV1().Nodes().Delete(context.TODO(), nodeName2, metav1.DeleteOptions{})

	g.Eventually(func() int {
		num_routes := len(nodes[0].StaticRoutes)
		return num_routes
	}, 10*time.Second).Should(gomega.Equal(0))
}
