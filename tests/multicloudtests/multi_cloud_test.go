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
package k8stest

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/k8s"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	crdfake "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/client/v1alpha1/clientset/versioned/fake"
	v1beta1crdfake "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/client/v1beta1/clientset/versioned/fake"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/tests/integrationtest"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	k8sfake "k8s.io/client-go/kubernetes/fake"

	// To Do: add test for openshift route
	//oshiftfake "github.com/openshift/client-go/route/clientset/versioned/fake"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
)

const defaultMockFilePath = "../avimockobjects"
const invalidFilePath = "invalidmock"

var kubeClient *k8sfake.Clientset
var crdClient *crdfake.Clientset
var v1beta1CRDClient *v1beta1crdfake.Clientset
var dynamicClient *dynamicfake.FakeDynamicClient
var keyChan chan string
var ctrl *k8s.AviController
var RegisteredInformers = []string{
	utils.ServiceInformer,
	utils.EndpointInformer,
	utils.IngressInformer,
	utils.IngressClassInformer,
	utils.SecretInformer,
	utils.NSInformer,
	utils.NodeInformer,
	utils.ConfigMapInformer,
}

func syncFuncForTest(key interface{}, wg *sync.WaitGroup) error {
	keyStr, ok := key.(string)
	if !ok {
		return nil
	}
	keyChan <- keyStr
	return nil
}

// empty key ("") means we are not expecting the key
func waitAndverify(t *testing.T, key string) {
	waitChan := make(chan int)
	go func() {
		time.Sleep(20 * time.Second)
		waitChan <- 1
	}()

	select {
	case data := <-keyChan:
		if key == "" {
			t.Fatalf("unpexpected key: %v", data)
		} else if data != key {
			t.Fatalf("error in match expected: %v, got: %v", key, data)
		}
	case _ = <-waitChan:
		if key != "" {
			t.Fatalf("timed out waiting for %v", key)
		}
	}
}

func setupQueue(stopCh <-chan struct{}) {
	ingestionQueue := utils.SharedWorkQueue().GetQueueByName(utils.ObjectIngestionLayer)
	wgIngestion := &sync.WaitGroup{}

	ingestionQueue.SyncFunc = syncFuncForTest
	ingestionQueue.Run(stopCh, wgIngestion)
}

func addConfigMap(t *testing.T) {
	integrationtest.AddConfigMap(kubeClient)
	time.Sleep(10 * time.Second)
}

func AddCMap() {
	aviCM := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: utils.GetAKONamespace(),
			Name:      "avi-k8s-config",
		},
	}
	kubeClient.CoreV1().ConfigMaps(utils.GetAKONamespace()).Create(context.TODO(), aviCM, metav1.CreateOptions{})
}

func DeleteConfigMap(t *testing.T) {
	err := kubeClient.CoreV1().ConfigMaps(utils.GetAKONamespace()).Delete(context.TODO(), "avi-k8s-config", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("error in deleting configmap: %v", err)
	}
	time.Sleep(10 * time.Second)
}

func ValidateIngress(t *testing.T) {

	integrationtest.AddDefaultIngressClass()
	waitAndverify(t, "IngressClass/avi-lb")

	// validate svc first
	svcExample := &corev1.Service{
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeLoadBalancer,
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "red-ns",
			Name:      "testsvc",
		},
	}
	_, err := kubeClient.CoreV1().Services("red-ns").Create(context.TODO(), svcExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Service: %v", err)
	}
	waitAndverify(t, "L4LBService/red-ns/testsvc")

	ingrExample := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "red-ns",
			Name:      "testingr",
		},
		Spec: networkingv1.IngressSpec{
			DefaultBackend: &networkingv1.IngressBackend{
				Service: &networkingv1.IngressServiceBackend{
					Name: "testsvc",
				},
			},
		},
	}
	_, err = kubeClient.NetworkingV1().Ingresses("red-ns").Create(context.TODO(), ingrExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	waitAndverify(t, "Ingress/red-ns/testingr")
	// delete svc and ingress
	err = kubeClient.CoreV1().Services("red-ns").Delete(context.TODO(), "testsvc", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("error in deleting Service: %v", err)
	}
	waitAndverify(t, "L4LBService/red-ns/testsvc")
	err = kubeClient.NetworkingV1().Ingresses("red-ns").Delete(context.TODO(), "testingr", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	waitAndverify(t, "Ingress/red-ns/testingr")

	integrationtest.RemoveDefaultIngressClass()
	waitAndverify(t, "IngressClass/avi-lb")
}

func ValidateNode(t *testing.T) {
	nodeExample := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:            "testnode",
			ResourceVersion: "1",
		},
		Spec: corev1.NodeSpec{
			PodCIDR:  "10.244.0.0/24",
			PodCIDRs: []string{"10.244.0.0/24"},
		},
		Status: corev1.NodeStatus{
			Addresses: []corev1.NodeAddress{
				{
					Type:    "InternalIP",
					Address: "10.1.1.2",
				},
			},
		},
	}
	_, err := kubeClient.CoreV1().Nodes().Create(context.TODO(), nodeExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Node: %v", err)
	}
	waitAndverify(t, utils.NodeObj+"/testnode")

	nodeExample.ObjectMeta.ResourceVersion = "2"
	nodeExample.Spec.PodCIDR = "10.230.0.0/24"
	_, err = kubeClient.CoreV1().Nodes().Update(context.TODO(), nodeExample, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating Node: %v", err)
	}
	waitAndverify(t, utils.NodeObj+"/testnode")

	err = kubeClient.CoreV1().Nodes().Delete(context.TODO(), "testnode", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("error in Deleting Node: %v", err)
	}
	waitAndverify(t, utils.NodeObj+"/testnode")
}

func injectMWForCloud() {
	integrationtest.AddMiddleware(func(w http.ResponseWriter, r *http.Request) {
		url := r.URL.EscapedPath()
		if r.Method == "GET" && strings.Contains(url, "/api/cloud/") {
			integrationtest.FeedMockCollectionData(w, r, invalidFilePath)

		} else if r.Method == "GET" {
			integrationtest.FeedMockCollectionData(w, r, defaultMockFilePath)

		} else if strings.Contains(url, "login") {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"success": "true"}`))
		}
	})
}

func TestMain(m *testing.M) {
	os.Setenv("INGRESS_API", "extensionv1")
	os.Setenv("VIP_NETWORK_LIST", `[{"networkName":"net123"}]`)
	os.Setenv("CLUSTER_NAME", "cluster")
	os.Setenv("SEG_NAME", "Default-Group")
	os.Setenv("NODE_NETWORK_LIST", `[{"networkName":"net123","cidrs":["10.79.168.0/22"]}]`)
	os.Setenv("CLOUD_NAME", "CLOUD_AWS")
	utils.SetCloudName("CLOUD_AWS")
	os.Setenv("SERVICE_TYPE", "NodePort")
	os.Setenv("POD_NAMESPACE", utils.AKO_DEFAULT_NS)
	os.Setenv("SHARD_VS_SIZE", "LARGE")
	os.Setenv("POD_NAME", "ako-0")

	akoControlConfig := lib.AKOControlConfig()
	kubeClient = k8sfake.NewSimpleClientset()
	dynamicClient = dynamicfake.NewSimpleDynamicClient(runtime.NewScheme())
	crdClient = crdfake.NewSimpleClientset()
	v1beta1CRDClient = v1beta1crdfake.NewSimpleClientset()
	data := map[string][]byte{
		"username": []byte("admin"),
		"password": []byte("admin"),
	}
	object := metav1.ObjectMeta{Name: "avi-secret", Namespace: utils.GetAKONamespace()}
	secret := &corev1.Secret{Data: data, ObjectMeta: object}
	kubeClient.CoreV1().Secrets(utils.GetAKONamespace()).Create(context.TODO(), secret, metav1.CreateOptions{})
	akoControlConfig.SetCRDClientset(crdClient)
	akoControlConfig.Setv1beta1CRDClientset(v1beta1CRDClient)
	akoControlConfig.SetAKOInstanceFlag(true)
	akoControlConfig.SetEventRecorder(lib.AKOEventComponent, kubeClient, true)
	akoControlConfig.SetDefaultLBController(true)
	utils.NewInformers(utils.KubeClientIntf{ClientSet: kubeClient}, RegisteredInformers)
	k8s.NewCRDInformers()

	integrationtest.InitializeFakeAKOAPIServer()
	integrationtest.NewAviFakeClientInstance(kubeClient)
	defer integrationtest.AviFakeClientInstance.Close()

	ctrl = k8s.SharedAviController()
	stopCh := utils.SetupSignalHandler()
	ctrl.Start(stopCh)
	keyChan = make(chan string)
	ctrlCh := make(chan struct{})
	quickSyncCh := make(chan struct{})
	AddCMap()
	integrationtest.KubeClient = kubeClient
	ctrl.SetSEGroupCloudNameFromNSAnnotations()

	ctrl.HandleConfigMap(k8s.K8sinformers{Cs: kubeClient, DynamicClient: dynamicClient}, ctrlCh, stopCh, quickSyncCh)
	ctrl.SetupEventHandlers(k8s.K8sinformers{Cs: kubeClient, DynamicClient: dynamicClient})
	setupQueue(stopCh)
	os.Exit(m.Run())
}

// Cloud does not have a ipam_provider_ref configured, sync should be disabled
func TestVcenterCloudNoIpamDuringBootup(t *testing.T) {
	DeleteConfigMap(t)
	os.Setenv("CLOUD_NAME", "CLOUD_VCENTER")
	utils.SetCloudName("CLOUD_VCENTER")
	os.Setenv("SERVICE_TYPE", "ClusterIP")
	injectMWForCloud()

	addConfigMap(t)

	if !ctrl.DisableSync {
		t.Fatalf("Validation for vcenter Cloud for ipam_provider_ref failed")
	}
	integrationtest.ResetMiddleware()
	DeleteConfigMap(t)
}

// TestAWSCloudValidation tests validation in place for public clouds
func TestAWSCloudValidation(t *testing.T) {
	os.Setenv("CLOUD_NAME", "CLOUD_AWS")
	utils.SetCloudName("CLOUD_AWS")
	os.Setenv("SERVICE_TYPE", "NodePort")
	os.Setenv("VIP_NETWORK_LIST", `[]`)

	addConfigMap(t)

	if !ctrl.DisableSync {
		t.Fatalf("CLOUD_AWS should not be allowed if VIP_NETWORK_LIST is empty")
	}
	DeleteConfigMap(t)
	os.Setenv("VIP_NETWORK_LIST", `[{"networkName":"net123"}]`)
}

// TestAzureCloudValidation tests validation in place for public clouds
func TestAzureCloudValidation(t *testing.T) {
	os.Setenv("CLOUD_NAME", "CLOUD_AZURE")
	utils.SetCloudName("CLOUD_AZURE")
	os.Setenv("SERVICE_TYPE", "NodePort")
	os.Setenv("VIP_NETWORK_LIST", `[]`)

	addConfigMap(t)

	if !ctrl.DisableSync {
		t.Fatalf("CLOUD_AZURE should not be allowed if VIP_NETWORK_LIST is empty")
	}
	DeleteConfigMap(t)
	os.Setenv("VIP_NETWORK_LIST", `[{"networkName":"net123"}]`)
}

// TestAWSCloudInClusterIPMode tests case where AWS CLOUD is configured in ClousterIP mode. Sync should be allowed.
func TestAWSCloudInClusterIPMode(t *testing.T) {
	os.Setenv("CLOUD_NAME", "CLOUD_AWS")
	utils.SetCloudName("CLOUD_AWS")
	os.Setenv("SERVICE_TYPE", "ClusterIP")

	addConfigMap(t)

	if ctrl.DisableSync {
		t.Fatalf("CLOUD_AWS should be allowed in ClusterIP mode")
	}
	DeleteConfigMap(t)
}

// TestAzureCloudInClusterIPMode tests case where Azure cloud is configured in ClusterIP mode. Sync should be allowed.
func TestAzureCloudInClusterIPMode(t *testing.T) {
	os.Setenv("CLOUD_NAME", "CLOUD_AZURE")
	utils.SetCloudName("CLOUD_AZURE")
	os.Setenv("SERVICE_TYPE", "ClusterIP")

	addConfigMap(t)

	if ctrl.DisableSync {
		t.Fatalf("CLOUD_AZURE should be allowed in ClusterIP mode")
	}
	DeleteConfigMap(t)

}

// TestGCPCloudInClusterIPMode tests case where GCP cloud is configured in ClusterIP mode. Sync should be allowed.
func TestGCPCloudInClusterIPMode(t *testing.T) {
	os.Setenv("CLOUD_NAME", "CLOUD_GCP")
	utils.SetCloudName("CLOUD_GCP")
	os.Setenv("SERVICE_TYPE", "ClusterIP")

	addConfigMap(t)

	if ctrl.DisableSync {
		t.Fatalf("CLOUD_GCP should  be allowed in ClusterIP mode")
	}
	DeleteConfigMap(t)

}

// TestAzureCloudInNodePortMode tests case where Azure cloud is configured in NodePort mode. Sync should be enabled.
func TestAzureCloudInNodePortMode(t *testing.T) {
	waitAndverify(t, fmt.Sprintf("Secret/%s/avi-secret", utils.GetAKONamespace()))
	os.Setenv("CLOUD_NAME", "CLOUD_AZURE")
	utils.SetCloudName("CLOUD_AZURE")
	os.Setenv("SERVICE_TYPE", "NodePort")

	// add the config and check if the sync is enabled
	addConfigMap(t)

	if ctrl.DisableSync {
		t.Fatalf("CLOUD_AZURE should be allowed in ClusterIP mode")
	}
	// validate the config
	ValidateIngress(t)
	ValidateNode(t)

	DeleteConfigMap(t)

}

// TestAWSCloudInNodePortMode tests case where AWS cloud is configured in NodePort mode. Sync should be enabled.
func TestAWSCloudInNodePortMode(t *testing.T) {
	os.Setenv("CLOUD_NAME", "CLOUD_AWS")
	utils.SetCloudName("CLOUD_AWS")
	os.Setenv("SERVICE_TYPE", "NodePort")

	// add the config and check if the sync is enabled
	addConfigMap(t)

	if ctrl.DisableSync {
		t.Fatalf("CLOUD_AWS should be allowed in ClusterIP mode")
	}
	// validate the config
	ValidateIngress(t)
	ValidateNode(t)
	DeleteConfigMap(t)

}
