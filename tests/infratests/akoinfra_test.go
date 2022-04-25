package infratests

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-infra/ingestion"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/tests/integrationtest"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	k8sfake "k8s.io/client-go/kubernetes/fake"
)

var kubeClient *k8sfake.Clientset
var dynamicClient *dynamicfake.FakeDynamicClient
var keyChan chan bool

func TestMain(m *testing.M) {
	os.Setenv("CLOUD_NAME", "CLOUD_NSXT")
	utils.SetCloudName("CLOUD_NSXT")
	keyChan = make(chan bool)

	kubeClient = k8sfake.NewSimpleClientset()
	data := map[string][]byte{
		"username":  []byte("admin"),
		"authtoken": []byte("admin"),
	}
	object := metav1.ObjectMeta{Name: "avi-secret", Namespace: utils.GetAKONamespace()}
	secret := &corev1.Secret{Data: data, ObjectMeta: object}
	kubeClient.CoreV1().Secrets(utils.GetAKONamespace()).Create(context.TODO(), secret, metav1.CreateOptions{})
	kubeClient.CoreV1().Secrets(utils.GetAKONamespace()).Get(context.TODO(), "avi-secret", metav1.GetOptions{})

	os.Exit(m.Run())
}

func waitAndverify(t *testing.T, expectTimeout bool) {
	waitChan := make(chan int)
	go func() {
		time.Sleep(10 * time.Second)
		waitChan <- 1
	}()

	select {
	case _ = <-keyChan:
		if expectTimeout {
			t.Fatalf("expected timeout, but got valid data")
		}
	case _ = <-waitChan:
		if !expectTimeout {
			t.Fatalf("timed out waiting")
		}
	}
}

func initInfraTest(testData []*unstructured.Unstructured) {
	gvrToKind := make(map[schema.GroupVersionResource]string)
	gvrToKind[lib.NetworkInfoGVR] = "namespacenetworkinfosList"

	dynamicClient = dynamicfake.NewSimpleDynamicClientWithCustomListKinds(runtime.NewScheme(), gvrToKind, testData[0], testData[1])
	dynamicClient.Resource(lib.NetworkInfoGVR).Namespace("default").Create(context.TODO(), testData[1], v1.CreateOptions{})
	lib.SetDynamicClientSet(dynamicClient)

	registeredInformers := []string{
		utils.SecretInformer,
	}
	informersArg := make(map[string]interface{})

	utils.NewInformers(utils.KubeClientIntf{ClientSet: kubeClient}, registeredInformers, informersArg)
	lib.NewDynamicInformers(dynamicClient, true)
}

// TestBoostrapNoALBEndpoint adds an akobootstrapconditions object with no albEndpoint.
// In this condition HandleVCF should wait.
func TestBoostrapNoALBEndpoint(t *testing.T) {
	var testData []*unstructured.Unstructured
	testData = append(testData, &unstructured.Unstructured{})
	testData[0].SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "ncp.vmware.com/v1alpha1",
		"kind":       "akobootstrapconditions",
		"metadata": map[string]interface{}{
			"name":      "testbs",
			"namespace": "default",
		},
		"spec": map[string]interface{}{
			"albCredentialSecretRef": map[string]interface{}{
				"name":      "avi-secret",
				"namespace": "avi-system",
			},
			"albTokenProperty": map[string]interface{}{
				"userName": "admin",
			},
		},
		"status": map[string]interface{}{
			/*"albEndpoint": map[string]interface{}{
				"hostUrl": "10.50.63.199",
			},*/
			"transportZone": map[string]interface{}{
				"path": "testpath",
			},
		},
	})

	testData = append(testData, &unstructured.Unstructured{})
	testData[1].SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "nsx.vmware.com/v1alpha1",
		"kind":       "namespacenetworkinfos",
		"metadata": map[string]interface{}{
			"name":      "testnetinfo",
			"namespace": "default",
		},
		"topology": map[string]interface{}{
			"aviSegmentPath": "testSeg",
			"gatewayPath":    "testGW",
			"ingressCIDRs": []interface{}{
				"10.20.30.0/24",
			},
		},
	})

	initInfraTest(testData)
	informers := ingestion.K8sinformers{Cs: kubeClient, DynamicClient: dynamicClient}
	c := ingestion.SharedVCFK8sController()
	stopCh := make(chan struct{})
	ctrlCh := make(chan struct{})

	integrationtest.NewAviFakeClientInstance(kubeClient)
	defer integrationtest.AviFakeClientInstance.Close()

	go func() {
		_ = c.HandleVCF(informers, stopCh, ctrlCh, true)

		lib.VCFInitialized = true
		keyChan <- true
	}()
	waitAndverify(t, true)
}

// TestBoostrapNoTransportZone adds an akobootstrapconditions object with no transportZone.
// In this condition HandleVCF should wait.
func TestBoostrapNoTransportZone(t *testing.T) {
	var testData []*unstructured.Unstructured
	testData = append(testData, &unstructured.Unstructured{})
	testData[0].SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "ncp.vmware.com/v1alpha1",
		"kind":       "akobootstrapconditions",
		"metadata": map[string]interface{}{
			"name":      "testbs",
			"namespace": "default",
		},
		"spec": map[string]interface{}{
			"albCredentialSecretRef": map[string]interface{}{
				"name":      "avi-secret",
				"namespace": "avi-system",
			},
			"albTokenProperty": map[string]interface{}{
				"userName": "admin",
			},
		},
		"status": map[string]interface{}{
			"albEndpoint": map[string]interface{}{
				"hostUrl": "10.50.63.199",
			},
			/*"transportZone": map[string]interface{}{
				"path": "testpath",
			},*/
		},
	})

	testData = append(testData, &unstructured.Unstructured{})
	testData[1].SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "nsx.vmware.com/v1alpha1",
		"kind":       "namespacenetworkinfos",
		"metadata": map[string]interface{}{
			"name":      "testnetinfo",
			"namespace": "default",
		},
		"topology": map[string]interface{}{
			"aviSegmentPath": "testSeg",
			"gatewayPath":    "testGW",
			"ingressCIDRs": []interface{}{
				"10.20.30.0/24",
			},
		},
	})

	initInfraTest(testData)
	informers := ingestion.K8sinformers{Cs: kubeClient, DynamicClient: dynamicClient}
	c := ingestion.SharedVCFK8sController()
	stopCh := make(chan struct{})

	ctrlCh := make(chan struct{})

	integrationtest.NewAviFakeClientInstance(kubeClient)
	defer integrationtest.AviFakeClientInstance.Close()

	go func() {
		_ = c.HandleVCF(informers, stopCh, ctrlCh, true)

		lib.VCFInitialized = true
		keyChan <- true
	}()
	waitAndverify(t, true)
}

// TestBoostrapNoSecretRef adds an akobootstrapconditions object with no albCredentialSecretRef.
// In this condition HandleVCF should wait.
func TestBoostrapNoSecretRef(t *testing.T) {
	var testData []*unstructured.Unstructured
	testData = append(testData, &unstructured.Unstructured{})
	testData[0].SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "ncp.vmware.com/v1alpha1",
		"kind":       "akobootstrapconditions",
		"metadata": map[string]interface{}{
			"name":      "testbs",
			"namespace": "default",
		},
		"spec": map[string]interface{}{
			/*"albCredentialSecretRef": map[string]interface{}{
				"name":      "avi-secret",
				"namespace": "avi-system",
			},*/
			"albTokenProperty": map[string]interface{}{
				"userName": "admin",
			},
		},
		"status": map[string]interface{}{
			"albEndpoint": map[string]interface{}{
				"hostUrl": "10.50.63.199",
			},
			"transportZone": map[string]interface{}{
				"path": "testpath",
			},
		},
	})

	testData = append(testData, &unstructured.Unstructured{})
	testData[1].SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "nsx.vmware.com/v1alpha1",
		"kind":       "namespacenetworkinfos",
		"metadata": map[string]interface{}{
			"name":      "testnetinfo",
			"namespace": "default",
		},
		"topology": map[string]interface{}{
			"aviSegmentPath": "testSeg",
			"gatewayPath":    "testGW",
			"ingressCIDRs": []interface{}{
				"10.20.30.0/24",
			},
		},
	})

	initInfraTest(testData)
	informers := ingestion.K8sinformers{Cs: kubeClient, DynamicClient: dynamicClient}
	c := ingestion.SharedVCFK8sController()
	stopCh := make(chan struct{})

	ctrlCh := make(chan struct{})

	integrationtest.NewAviFakeClientInstance(kubeClient)
	defer integrationtest.AviFakeClientInstance.Close()

	go func() {
		_ = c.HandleVCF(informers, stopCh, ctrlCh, true)

		lib.VCFInitialized = true
		keyChan <- true
	}()
	waitAndverify(t, true)
}

// TestBoostrapNoUserName adds an akobootstrapconditions object with no albTokenProperty.
// In this condition HandleVCF should wait.
func TestBoostrapNoUserName(t *testing.T) {
	var testData []*unstructured.Unstructured
	testData = append(testData, &unstructured.Unstructured{})
	testData[0].SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "ncp.vmware.com/v1alpha1",
		"kind":       "akobootstrapconditions",
		"metadata": map[string]interface{}{
			"name":      "testbs",
			"namespace": "default",
		},
		"spec": map[string]interface{}{
			"albCredentialSecretRef": map[string]interface{}{
				"name":      "avi-secret",
				"namespace": "avi-system",
			},
			/*"albTokenProperty": map[string]interface{}{
				"userName": "admin",
			},*/
		},
		"status": map[string]interface{}{
			"albEndpoint": map[string]interface{}{
				"hostUrl": "10.50.63.199",
			},
			"transportZone": map[string]interface{}{
				"path": "testpath",
			},
		},
	})

	testData = append(testData, &unstructured.Unstructured{})
	testData[1].SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "nsx.vmware.com/v1alpha1",
		"kind":       "namespacenetworkinfos",
		"metadata": map[string]interface{}{
			"name":      "testnetinfo",
			"namespace": "default",
		},
		"topology": map[string]interface{}{
			"aviSegmentPath": "testSeg",
			"gatewayPath":    "testGW",
			"ingressCIDRs": []interface{}{
				"10.20.30.0/24",
			},
		},
	})

	initInfraTest(testData)
	informers := ingestion.K8sinformers{Cs: kubeClient, DynamicClient: dynamicClient}
	c := ingestion.SharedVCFK8sController()
	stopCh := make(chan struct{})

	ctrlCh := make(chan struct{})

	integrationtest.NewAviFakeClientInstance(kubeClient)
	defer integrationtest.AviFakeClientInstance.Close()

	go func() {
		_ = c.HandleVCF(informers, stopCh, ctrlCh, true)

		lib.VCFInitialized = true
		keyChan <- true
	}()
	waitAndverify(t, true)
}

func TestValidBootstrapData(t *testing.T) {
	var testData []*unstructured.Unstructured
	testData = append(testData, &unstructured.Unstructured{})
	testData[0].SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "ncp.vmware.com/v1alpha1",
		"kind":       "akobootstrapconditions",
		"metadata": map[string]interface{}{
			"name":      "testbs",
			"namespace": "default",
		},
		"spec": map[string]interface{}{
			"albCredentialSecretRef": map[string]interface{}{
				"name":      "avi-secret",
				"namespace": "avi-system",
			},
			"albTokenProperty": map[string]interface{}{
				"userName": "admin",
			},
		},
		"status": map[string]interface{}{
			"albEndpoint": map[string]interface{}{
				"hostUrl": "10.50.63.199",
			},
			"transportZone": map[string]interface{}{
				"path": "testpath",
			},
		},
	})

	testData = append(testData, &unstructured.Unstructured{})
	testData[1].SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "nsx.vmware.com/v1alpha1",
		"kind":       "namespacenetworkinfos",
		"metadata": map[string]interface{}{
			"name":      "testnetinfo",
			"namespace": "default",
		},
		"topology": map[string]interface{}{
			"aviSegmentPath": "testSeg",
			"gatewayPath":    "testGW",
			"ingressCIDRs": []interface{}{
				"10.20.30.0/24",
			},
		},
	})

	initInfraTest(testData)
	informers := ingestion.K8sinformers{Cs: kubeClient, DynamicClient: dynamicClient}
	c := ingestion.SharedVCFK8sController()
	stopCh := make(chan struct{})

	ctrlCh := make(chan struct{})

	integrationtest.NewAviFakeClientInstance(kubeClient)
	defer integrationtest.AviFakeClientInstance.Close()

	go func() {
		_ = c.HandleVCF(informers, stopCh, ctrlCh, true)

		lib.VCFInitialized = true
		keyChan <- true
	}()
	time.Sleep(20 * time.Second)
	waitAndverify(t, false)
}
