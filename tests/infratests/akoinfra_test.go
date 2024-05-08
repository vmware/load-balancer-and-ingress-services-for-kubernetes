package infratests

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-infra/ingestion"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/k8s"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	v1beta1crdfake "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/client/v1beta1/clientset/versioned/fake"
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

var kubeClient *k8sfake.Clientset
var dynamicClient *dynamicfake.FakeDynamicClient
var keyChan chan bool
var V1beta1CRDClient *v1beta1crdfake.Clientset

var gvrToKind = map[schema.GroupVersionResource]string{
	lib.NetworkInfoGVR:     "namespacenetworkinfosList",
	lib.ClusterNetworkGVR:  "clusternetworkinfosList",
	lib.VPCGVR:             "vpcsList",
	lib.AvailabilityZoneVR: "availabilityzonesList",
}

func TestMain(m *testing.M) {
	os.Setenv("CLOUD_NAME", "CLOUD_NSXT")
	os.Setenv("POD_NAMESPACE", "vmware-system-ako")
	os.Setenv("VCF_CLUSTER", "true")
	utils.SetCloudName("CLOUD_NSXT")
	lib.SetClusterID("domain-c10:9d4c5eaa-7ddd-40c8-aadf-2cff7b4bee82")
	keyChan = make(chan bool)

	kubeClient = k8sfake.NewSimpleClientset()
	integrationtest.KubeClient = kubeClient
	data := map[string][]byte{
		"username":  []byte("admin"),
		"authtoken": []byte("admin"),
	}
	object := metav1.ObjectMeta{Name: "avi-secret", Namespace: utils.GetAKONamespace()}
	secret := &corev1.Secret{Data: data, ObjectMeta: object}
	kubeClient.CoreV1().Secrets(utils.GetAKONamespace()).Create(context.TODO(), secret, metav1.CreateOptions{})
	kubeClient.CoreV1().Secrets(utils.GetAKONamespace()).Get(context.TODO(), "avi-secret", metav1.GetOptions{})

	integrationtest.NewAviFakeClientInstance(kubeClient, true)
	defer integrationtest.AviFakeClientInstance.Close()

	aviCM := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: utils.GetAKONamespace(),
			Name:      lib.AviConfigMap,
		},
		Data: map[string]string{
			"cloudName":                  "CLOUD_NSXT",
			"clusterID":                  "domain-c10:9d4c5eaa-7ddd-40c8-aadf-2cff7b4bee82",
			"controllerIP":               strings.Split(integrationtest.AviFakeClientInstance.URL, "https://")[1],
			"credentialsSecretName":      "avi-secret",
			"credentialsSecretNamespace": "vmware-system-ako",
		},
	}
	kubeClient.CoreV1().ConfigMaps(utils.GetAKONamespace()).Create(context.TODO(), aviCM, metav1.CreateOptions{})
	integrationtest.AddDefaultNamespace(utils.GetAKONamespace())
	integrationtest.AddDefaultNamespace()

	akoControlConfig := lib.AKOControlConfig()
	V1beta1CRDClient = v1beta1crdfake.NewSimpleClientset()
	akoControlConfig.SetAKOInstanceFlag(true)
	akoControlConfig.SetCRDClientsetAndEnableInfraSettingParam(V1beta1CRDClient)

	registeredInformers := []string{
		utils.ConfigMapInformer,
		utils.NSInformer,
		utils.SecretInformer,
	}
	utils.NewInformers(utils.KubeClientIntf{ClientSet: kubeClient}, registeredInformers)
	k8s.NewInfraSettingCRDInformer()
	a := ingestion.NewAviControllerInfra(kubeClient)

	c := ingestion.SharedVCFK8sController()
	stopCh := make(chan struct{})
	ctrlCh := make(chan struct{})
	c.HandleVCF(stopCh, ctrlCh)

	lib.RunAviInfraSettingInformer(stopCh)
	c.AddSecretEventHandler(stopCh)

	a.AnnotateSystemNamespace(lib.GetClusterID(), utils.CloudName)
	c.AddNamespaceEventHandler(stopCh)

	os.Exit(m.Run())
}

func setupInfraTest(testData []*unstructured.Unstructured) {
	dynamicClient = dynamicfake.NewSimpleDynamicClientWithCustomListKinds(runtime.NewScheme(), gvrToKind, testData[0])
	lib.SetDynamicClientSet(dynamicClient)
	lib.NewDynamicInformers(dynamicClient, true)
}

func TestAKOInfraAviInfraSettingCreationT1(t *testing.T) {
	// create ns, namespacenetworkinfos CR, AKO should create an AviInfraSetting
	// ns should be annotated with InfraSetting
	g := gomega.NewGomegaWithT(t)
	var testData []*unstructured.Unstructured
	testData = append(testData, &unstructured.Unstructured{})
	testData[0].SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "nsx.vmware.com/v1alpha1",
		"kind":       "namespacenetworkinfos",
		"metadata": map[string]interface{}{
			"name":      "testnetinfo",
			"namespace": "default",
		},
		"topology": map[string]interface{}{
			"aviSegmentPath": "testSeg",
			"gatewayPath":    "testGW",
		},
	})

	setupInfraTest(testData)

	c := ingestion.SharedVCFK8sController()
	c.InitNetworkingHandler()
	c.AddNetworkInfoEventHandler(make(chan struct{}))
	worker := c.InitFullSyncWorker()
	go worker.Run()

	_, err := dynamicClient.Resource(lib.NetworkInfoGVR).Namespace("default").Create(context.TODO(), testData[0], v1.CreateOptions{})
	if err != nil {
		t.Fatalf("failed to create namespacenetworkinfos CR, error: %s", err.Error())
	}

	g.Eventually(func() bool {
		if infraSetting, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().AviInfraSettings().Get(context.TODO(), "testGW", metav1.GetOptions{}); err != nil {
			return false
		} else {
			return *infraSetting.Spec.NSXSettings.T1LR == "testGW" &&
				infraSetting.Spec.Network.VipNetworks[0].NetworkName == lib.GetVCFNetworkName() &&
				infraSetting.Spec.SeGroup.Name == lib.GetClusterID()
		}
	}, 30*time.Second).Should(gomega.Equal(true))

	g.Eventually(func() bool {
		if ns, err := kubeClient.CoreV1().Namespaces().Get(context.TODO(), "default", metav1.GetOptions{}); err != nil {
			return false
		} else {
			return ns.Annotations[lib.InfraSettingNameAnnotation] == "testGW"
		}
	}, 10*time.Second).Should(gomega.Equal(true))

	err = dynamicClient.Resource(lib.NetworkInfoGVR).Namespace("default").Delete(context.TODO(), "testnetinfo", v1.DeleteOptions{})
	if err != nil {
		t.Fatalf("failed to delete namespacenetworkinfo CR, error: %s", err.Error())
	}

	g.Eventually(func() bool {
		if _, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().AviInfraSettings().Get(context.TODO(), "testGW", metav1.GetOptions{}); err != nil {
			return false
		}
		return true
	}, 10*time.Second).Should(gomega.Equal(false))

	g.Eventually(func() bool {
		if ns, err := kubeClient.CoreV1().Namespaces().Get(context.TODO(), "default", metav1.GetOptions{}); err != nil {
			return false
		} else {
			return ns.Annotations[lib.InfraSettingNameAnnotation] != "testGW"
		}
	}, 10*time.Second).Should(gomega.Equal(true))

	worker.Shutdown()
}

func TestAKOInfraAviInfraSettingCreationVPC(t *testing.T) {
	// create ns, VPC CR, AKO should create an AviInfraSetting
	// ns should be annotated with InfraSetting
	g := gomega.NewGomegaWithT(t)
	var testData []*unstructured.Unstructured
	testData = append(testData, &unstructured.Unstructured{})
	testData[0].SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "nsx.vmware.com/v1alpha1",
		"kind":       "vpcs",
		"metadata": map[string]interface{}{
			"name":      "testvpcinfo",
			"namespace": "default",
		},
		"status": map[string]interface{}{
			"lbSubnetPath":    "/orgs/default/projects/test-project/vpcs/testGW/subnets/_AVI_SUBNET--LB",
			"nsxResourcePath": "/orgs/default/projects/test-project/vpcs/testGW",
		},
	})

	setupInfraTest(testData)

	os.Setenv("VPC_MODE", "true")

	c := ingestion.SharedVCFK8sController()
	c.InitNetworkingHandler()
	c.AddNetworkInfoEventHandler(make(chan struct{}))
	worker := c.InitFullSyncWorker()
	go worker.Run()

	_, err := dynamicClient.Resource(lib.VPCGVR).Namespace("default").Create(context.TODO(), testData[0], v1.CreateOptions{})
	if err != nil {
		t.Fatalf("failed to create namespacenetworkinfos CR, error: %s", err.Error())
	}

	g.Eventually(func() bool {
		if infraSetting, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().AviInfraSettings().Get(context.TODO(), "testGW", metav1.GetOptions{}); err != nil {
			return false
		} else {
			return *infraSetting.Spec.NSXSettings.T1LR == "/orgs/default/projects/test-project/vpcs/testGW" &&
				*infraSetting.Spec.NSXSettings.Project == "test-project" &&
				len(infraSetting.Spec.Network.VipNetworks) == 0 &&
				infraSetting.Spec.SeGroup.Name == lib.GetClusterID()
		}
	}, 30*time.Second).Should(gomega.Equal(true))

	g.Eventually(func() bool {
		if ns, err := kubeClient.CoreV1().Namespaces().Get(context.TODO(), "default", metav1.GetOptions{}); err != nil {
			return false
		} else {
			return ns.Annotations[lib.InfraSettingNameAnnotation] == "testGW"
		}
	}, 10*time.Second).Should(gomega.Equal(true))

	err = dynamicClient.Resource(lib.VPCGVR).Namespace("default").Delete(context.TODO(), "testvpcinfo", v1.DeleteOptions{})
	if err != nil {
		t.Fatalf("failed to delete namespacenetworkinfo CR, error: %s", err.Error())
	}

	g.Eventually(func() bool {
		if _, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().AviInfraSettings().Get(context.TODO(), "testGW", metav1.GetOptions{}); err != nil {
			return false
		}
		return true
	}, 10*time.Second).Should(gomega.Equal(false))

	g.Eventually(func() bool {
		if ns, err := kubeClient.CoreV1().Namespaces().Get(context.TODO(), "default", metav1.GetOptions{}); err != nil {
			return false
		} else {
			return ns.Annotations[lib.InfraSettingNameAnnotation] != "testGW"
		}
	}, 10*time.Second).Should(gomega.Equal(true))

	worker.Shutdown()
	os.Unsetenv("VPC_MODE")
}

func TestAKOInfraMultiAviInfraSettingCreationT1(t *testing.T) {
	// create multipl ns and namespacenetworkinfos CR with same lr,
	// AKO should create an AviInfraSetting for each unique combination of lr and ingress cidr
	// ns should be annotated with InfraSetting
	g := gomega.NewGomegaWithT(t)
	integrationtest.AddDefaultNamespace("red")
	integrationtest.AddDefaultNamespace("red-ns")
	var testData []*unstructured.Unstructured
	testData = append(testData, &unstructured.Unstructured{})
	testData[0].SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "nsx.vmware.com/v1alpha1",
		"kind":       "namespacenetworkinfos",
		"metadata": map[string]interface{}{
			"name":      "testnetinfo",
			"namespace": "default",
		},
		"topology": map[string]interface{}{
			"aviSegmentPath": "testSeg",
			"gatewayPath":    "testGW",
		},
	})
	testData = append(testData, &unstructured.Unstructured{})
	testData[1].SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "nsx.vmware.com/v1alpha1",
		"kind":       "namespacenetworkinfos",
		"metadata": map[string]interface{}{
			"name":      "testnetinfo",
			"namespace": "red",
		},
		"topology": map[string]interface{}{
			"aviSegmentPath": "testSeg",
			"gatewayPath":    "testGW",
			"ingressCIDRs": []interface{}{
				"10.20.30.0/24",
			},
		},
	})
	testData = append(testData, &unstructured.Unstructured{})
	testData[2].SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "nsx.vmware.com/v1alpha1",
		"kind":       "namespacenetworkinfos",
		"metadata": map[string]interface{}{
			"name":      "testnetinfo",
			"namespace": "red-ns",
		},
		"topology": map[string]interface{}{
			"aviSegmentPath": "testSeg",
			"gatewayPath":    "testGW",
		},
	})

	setupInfraTest(testData)

	c := ingestion.SharedVCFK8sController()
	c.InitNetworkingHandler()
	c.AddNetworkInfoEventHandler(make(chan struct{}))
	worker := c.InitFullSyncWorker()
	go worker.Run()

	_, err := dynamicClient.Resource(lib.NetworkInfoGVR).Namespace("default").Create(context.TODO(), testData[0], v1.CreateOptions{})
	if err != nil {
		t.Fatalf("failed to create namespacenetworkinfos CR, error: %s", err.Error())
	}

	_, err = dynamicClient.Resource(lib.NetworkInfoGVR).Namespace("red").Create(context.TODO(), testData[1], v1.CreateOptions{})
	if err != nil {
		t.Fatalf("failed to create namespacenetworkinfos CR, error: %s", err.Error())
	}
	_, err = dynamicClient.Resource(lib.NetworkInfoGVR).Namespace("red-ns").Create(context.TODO(), testData[2], v1.CreateOptions{})
	if err != nil {
		t.Fatalf("failed to create namespacenetworkinfos CR, error: %s", err.Error())
	}

	g.Eventually(func() bool {
		if infraSetting, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().AviInfraSettings().Get(context.TODO(), "testGW", metav1.GetOptions{}); err != nil {
			return false
		} else {
			return *infraSetting.Spec.NSXSettings.T1LR == "testGW" &&
				infraSetting.Spec.Network.VipNetworks[0].NetworkName == lib.GetVCFNetworkName() &&
				infraSetting.Spec.SeGroup.Name == lib.GetClusterID()
		}
	}, 30*time.Second).Should(gomega.Equal(true))

	g.Eventually(func() bool {
		if infraSetting, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().AviInfraSettings().Get(context.TODO(), "testGW-red", metav1.GetOptions{}); err != nil {
			return false
		} else {
			return *infraSetting.Spec.NSXSettings.T1LR == "testGW" &&
				infraSetting.Spec.Network.VipNetworks[0].NetworkName == lib.GetVCFNetworkNameWithNS("red") &&
				infraSetting.Spec.SeGroup.Name == lib.GetClusterID()
		}
	}, 30*time.Second).Should(gomega.Equal(true))

	g.Eventually(func() bool {
		if ns, err := kubeClient.CoreV1().Namespaces().Get(context.TODO(), "default", metav1.GetOptions{}); err != nil {
			return false
		} else {
			return ns.Annotations[lib.InfraSettingNameAnnotation] == "testGW"
		}
	}, 10*time.Second).Should(gomega.Equal(true))

	g.Eventually(func() bool {
		if ns, err := kubeClient.CoreV1().Namespaces().Get(context.TODO(), "red", metav1.GetOptions{}); err != nil {
			return false
		} else {
			return ns.Annotations[lib.InfraSettingNameAnnotation] == "testGW-red"
		}
	}, 10*time.Second).Should(gomega.Equal(true))

	g.Eventually(func() bool {
		if ns, err := kubeClient.CoreV1().Namespaces().Get(context.TODO(), "red-ns", metav1.GetOptions{}); err != nil {
			return false
		} else {
			return ns.Annotations[lib.InfraSettingNameAnnotation] == "testGW"
		}
	}, 10*time.Second).Should(gomega.Equal(true))

	err = dynamicClient.Resource(lib.NetworkInfoGVR).Namespace("default").Delete(context.TODO(), "testnetinfo", v1.DeleteOptions{})
	if err != nil {
		t.Fatalf("failed to delete namespacenetworkinfo CR, error: %s", err.Error())
	}

	err = dynamicClient.Resource(lib.NetworkInfoGVR).Namespace("red").Delete(context.TODO(), "testnetinfo", v1.DeleteOptions{})
	if err != nil {
		t.Fatalf("failed to delete namespacenetworkinfo CR, error: %s", err.Error())
	}

	err = dynamicClient.Resource(lib.NetworkInfoGVR).Namespace("red-ns").Delete(context.TODO(), "testnetinfo", v1.DeleteOptions{})
	if err != nil {
		t.Fatalf("failed to delete namespacenetworkinfo CR, error: %s", err.Error())
	}

	g.Eventually(func() bool {
		if _, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().AviInfraSettings().Get(context.TODO(), "testGW", metav1.GetOptions{}); err != nil {
			return false
		}
		return true
	}, 10*time.Second).Should(gomega.Equal(false))

	g.Eventually(func() bool {
		if _, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().AviInfraSettings().Get(context.TODO(), "testGW-red", metav1.GetOptions{}); err != nil {
			return false
		}
		return true
	}, 10*time.Second).Should(gomega.Equal(false))

	g.Eventually(func() bool {
		if ns, err := kubeClient.CoreV1().Namespaces().Get(context.TODO(), "default", metav1.GetOptions{}); err != nil {
			return false
		} else {
			return ns.Annotations[lib.InfraSettingNameAnnotation] != "testGW"
		}
	}, 10*time.Second).Should(gomega.Equal(true))
	g.Eventually(func() bool {
		if ns, err := kubeClient.CoreV1().Namespaces().Get(context.TODO(), "red", metav1.GetOptions{}); err != nil {
			return false
		} else {
			return ns.Annotations[lib.InfraSettingNameAnnotation] != "testGW-red"
		}
	}, 10*time.Second).Should(gomega.Equal(true))
	g.Eventually(func() bool {
		if ns, err := kubeClient.CoreV1().Namespaces().Get(context.TODO(), "red-ns", metav1.GetOptions{}); err != nil {
			return false
		} else {
			return ns.Annotations[lib.InfraSettingNameAnnotation] != "testGW"
		}
	}, 10*time.Second).Should(gomega.Equal(true))

	worker.Shutdown()
}

func TestAKOInfraMultiAviInfraSettingCreationVPC(t *testing.T) {
	// create multipl ns and vpc CRs,
	// AKO should create an AviInfraSetting for each unique vpcpath
	// ns should be annotated with InfraSetting
	g := gomega.NewGomegaWithT(t)
	integrationtest.AddDefaultNamespace("red")
	integrationtest.AddDefaultNamespace("red-ns")
	var testData []*unstructured.Unstructured
	testData = append(testData, &unstructured.Unstructured{})
	testData[0].SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "nsx.vmware.com/v1alpha1",
		"kind":       "vpcs",
		"metadata": map[string]interface{}{
			"name":      "testvpcinfo",
			"namespace": "red",
		},
		"status": map[string]interface{}{
			"lbSubnetPath":    "/orgs/default/projects/test-project/vpcs/testGW/subnets/_AVI_SUBNET--LB",
			"nsxResourcePath": "/orgs/default/projects/test-project/vpcs/testGW-red",
		},
	})
	testData = append(testData, &unstructured.Unstructured{})
	testData[1].SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "nsx.vmware.com/v1alpha1",
		"kind":       "vpcs",
		"metadata": map[string]interface{}{
			"name":      "testvpcinfo",
			"namespace": "red-ns",
		},
		"status": map[string]interface{}{
			"lbSubnetPath":    "/orgs/default/projects/test-project/vpcs/testGW/subnets/_AVI_SUBNET--LB",
			"nsxResourcePath": "/orgs/default/projects/test-project/vpcs/testGW",
		},
	})

	namespace, err := kubeClient.CoreV1().Namespaces().Get(context.TODO(), "default", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Error occurred while fetching default namespace: %v", err)
	}
	namespace.ResourceVersion = "2"
	namespace.Annotations = map[string]string{
		"nsx.vmware.com/vpc_name": "red-ns/testvpcinfo",
	}
	_, err = kubeClient.CoreV1().Namespaces().Update(context.TODO(), namespace, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("Error occurred while Updating namespace: %v", err)
	}

	setupInfraTest(testData)

	os.Setenv("VPC_MODE", "true")

	c := ingestion.SharedVCFK8sController()
	c.InitNetworkingHandler()
	c.AddNetworkInfoEventHandler(make(chan struct{}))
	worker := c.InitFullSyncWorker()
	go worker.Run()

	_, err = dynamicClient.Resource(lib.VPCGVR).Namespace("red").Create(context.TODO(), testData[0], v1.CreateOptions{})
	if err != nil {
		t.Fatalf("failed to create namespacenetworkinfos CR, error: %s", err.Error())
	}
	_, err = dynamicClient.Resource(lib.VPCGVR).Namespace("red-ns").Create(context.TODO(), testData[1], v1.CreateOptions{})
	if err != nil {
		t.Fatalf("failed to create namespacenetworkinfos CR, error: %s", err.Error())
	}

	g.Eventually(func() bool {
		if infraSetting, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().AviInfraSettings().Get(context.TODO(), "testGW", metav1.GetOptions{}); err != nil {
			return false
		} else {
			return *infraSetting.Spec.NSXSettings.T1LR == "/orgs/default/projects/test-project/vpcs/testGW" &&
				*infraSetting.Spec.NSXSettings.Project == "test-project" &&
				len(infraSetting.Spec.Network.VipNetworks) == 0 &&
				infraSetting.Spec.SeGroup.Name == lib.GetClusterID()
		}
	}, 30*time.Second).Should(gomega.Equal(true))

	g.Eventually(func() bool {
		if infraSetting, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().AviInfraSettings().Get(context.TODO(), "testGW-red", metav1.GetOptions{}); err != nil {
			return false
		} else {
			return *infraSetting.Spec.NSXSettings.T1LR == "/orgs/default/projects/test-project/vpcs/testGW-red" &&
				*infraSetting.Spec.NSXSettings.Project == "test-project" &&
				len(infraSetting.Spec.Network.VipNetworks) == 0 &&
				infraSetting.Spec.SeGroup.Name == lib.GetClusterID()
		}
	}, 30*time.Second).Should(gomega.Equal(true))

	g.Eventually(func() bool {
		if ns, err := kubeClient.CoreV1().Namespaces().Get(context.TODO(), "default", metav1.GetOptions{}); err != nil {
			return false
		} else {
			return ns.Annotations[lib.InfraSettingNameAnnotation] == "testGW"
		}
	}, 10*time.Second).Should(gomega.Equal(true))

	g.Eventually(func() bool {
		if ns, err := kubeClient.CoreV1().Namespaces().Get(context.TODO(), "red", metav1.GetOptions{}); err != nil {
			return false
		} else {
			return ns.Annotations[lib.InfraSettingNameAnnotation] == "testGW-red"
		}
	}, 10*time.Second).Should(gomega.Equal(true))

	g.Eventually(func() bool {
		if ns, err := kubeClient.CoreV1().Namespaces().Get(context.TODO(), "red-ns", metav1.GetOptions{}); err != nil {
			return false
		} else {
			return ns.Annotations[lib.InfraSettingNameAnnotation] == "testGW"
		}
	}, 10*time.Second).Should(gomega.Equal(true))

	err = dynamicClient.Resource(lib.VPCGVR).Namespace("red").Delete(context.TODO(), "testvpcinfo", v1.DeleteOptions{})
	if err != nil {
		t.Fatalf("failed to delete namespacenetworkinfo CR, error: %s", err.Error())
	}

	err = dynamicClient.Resource(lib.VPCGVR).Namespace("red-ns").Delete(context.TODO(), "testvpcinfo", v1.DeleteOptions{})
	if err != nil {
		t.Fatalf("failed to delete namespacenetworkinfo CR, error: %s", err.Error())
	}

	g.Eventually(func() bool {
		if _, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().AviInfraSettings().Get(context.TODO(), "testGW", metav1.GetOptions{}); err != nil {
			return false
		}
		return true
	}, 10*time.Second).Should(gomega.Equal(false))

	g.Eventually(func() bool {
		if _, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().AviInfraSettings().Get(context.TODO(), "testGW-red", metav1.GetOptions{}); err != nil {
			return false
		}
		return true
	}, 10*time.Second).Should(gomega.Equal(false))

	g.Eventually(func() bool {
		if ns, err := kubeClient.CoreV1().Namespaces().Get(context.TODO(), "default", metav1.GetOptions{}); err != nil {
			return false
		} else {
			return ns.Annotations[lib.InfraSettingNameAnnotation] != "testGW"
		}
	}, 10*time.Second).Should(gomega.Equal(true))
	g.Eventually(func() bool {
		if ns, err := kubeClient.CoreV1().Namespaces().Get(context.TODO(), "red", metav1.GetOptions{}); err != nil {
			return false
		} else {
			return ns.Annotations[lib.InfraSettingNameAnnotation] != "testGW-red"
		}
	}, 10*time.Second).Should(gomega.Equal(true))
	g.Eventually(func() bool {
		if ns, err := kubeClient.CoreV1().Namespaces().Get(context.TODO(), "red-ns", metav1.GetOptions{}); err != nil {
			return false
		} else {
			return ns.Annotations[lib.InfraSettingNameAnnotation] != "testGW"
		}
	}, 10*time.Second).Should(gomega.Equal(true))

	worker.Shutdown()
	os.Unsetenv("VPC_MODE")
}
