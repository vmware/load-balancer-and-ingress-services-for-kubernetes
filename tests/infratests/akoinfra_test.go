package infratests

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
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

var (
	kubeClient       *k8sfake.Clientset
	dynamicClient    *dynamicfake.FakeDynamicClient
	keyChan          chan bool
	V1beta1CRDClient *v1beta1crdfake.Clientset
	objNameMap       integrationtest.ObjectNameMap
)

var gvrToKind = map[schema.GroupVersionResource]string{
	lib.NetworkInfoGVR:             "namespacenetworkinfosList",
	lib.ClusterNetworkGVR:          "clusternetworkinfosList",
	lib.VPCNetworkConfigurationGVR: "vpcnetworkconfigurationsList",
	lib.AvailabilityZoneVR:         "availabilityzonesList",
	lib.SupervisorCapabilityGVR:    "capabilitiesList",
}

func annotateNamespaceWithVpcNetworkConfigCR(t *testing.T, ns, vpcNetConfigCR string) {
	namespace, err := kubeClient.CoreV1().Namespaces().Get(context.TODO(), ns, metav1.GetOptions{})
	if err != nil {
		namespace := (integrationtest.FakeNamespace{
			Name:   ns,
			Labels: map[string]string{},
		}).Namespace()
		namespace.ResourceVersion = "1"
		namespace.Annotations = map[string]string{
			"nsx.vmware.com/vpc_network_config": vpcNetConfigCR,
		}
		_, err = kubeClient.CoreV1().Namespaces().Create(context.TODO(), namespace, metav1.CreateOptions{})
		if err != nil {
			t.Fatalf("Error occurred while Adding namespace: %v", err)
		}
	} else {
		n, err := strconv.Atoi(namespace.ResourceVersion)
		if err != nil {
			t.Fatalf("Failed to convert version from string to int: %v", err)
		}
		namespace.ResourceVersion = strconv.Itoa(n + 1)
		anns := namespace.Annotations
		anns["nsx.vmware.com/vpc_network_config"] = vpcNetConfigCR
		_, err = kubeClient.CoreV1().Namespaces().Update(context.TODO(), namespace, metav1.UpdateOptions{})
		if err != nil {
			t.Fatalf("Error occurred while Updating namespace: %v", err)
		}
	}
}

func annotateNamespaceWithCloud(t *testing.T, ns, cloudName string) {
	namespace, err := kubeClient.CoreV1().Namespaces().Get(context.TODO(), ns, metav1.GetOptions{})
	if err != nil {
		namespace := (integrationtest.FakeNamespace{
			Name:   ns,
			Labels: map[string]string{},
		}).Namespace()
		namespace.ResourceVersion = "1"
		namespace.Annotations = map[string]string{
			"ako.vmware.com/wcp-cloud-name": cloudName,
		}
		_, err = kubeClient.CoreV1().Namespaces().Create(context.TODO(), namespace, metav1.CreateOptions{})
		if err != nil {
			t.Fatalf("Error occurred while Adding namespace: %v", err)
		}
	} else {
		n, err := strconv.Atoi(namespace.ResourceVersion)
		if err != nil {
			t.Fatalf("Failed to convert version from string to int: %v", err)
		}
		namespace.ResourceVersion = strconv.Itoa(n + 1)
		anns := namespace.Annotations
		anns["ako.vmware.com/wcp-cloud-name"] = cloudName
		_, err = kubeClient.CoreV1().Namespaces().Update(context.TODO(), namespace, metav1.UpdateOptions{})
		if err != nil {
			t.Fatalf("Error occurred while Updating namespace: %v", err)
		}
	}
}

func annotateNamespaceWithSEG(t *testing.T, ns, segName string) {
	namespace, err := kubeClient.CoreV1().Namespaces().Get(context.TODO(), ns, metav1.GetOptions{})
	if err != nil {
		namespace := (integrationtest.FakeNamespace{
			Name:   ns,
			Labels: map[string]string{},
		}).Namespace()
		namespace.ResourceVersion = "1"
		namespace.Annotations = map[string]string{
			"ako.vmware.com/wcp-se-group": segName,
		}
		_, err = kubeClient.CoreV1().Namespaces().Create(context.TODO(), namespace, metav1.CreateOptions{})
		if err != nil {
			t.Fatalf("Error occurred while Adding namespace: %v", err)
		}
	} else {
		n, err := strconv.Atoi(namespace.ResourceVersion)
		if err != nil {
			t.Fatalf("Failed to convert version from string to int: %v", err)
		}
		namespace.ResourceVersion = strconv.Itoa(n + 1)
		anns := namespace.Annotations
		anns["ako.vmware.com/wcp-se-group"] = segName
		_, err = kubeClient.CoreV1().Namespaces().Update(context.TODO(), namespace, metav1.UpdateOptions{})
		if err != nil {
			t.Fatalf("Error occurred while Updating namespace: %v", err)
		}
	}
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

	a.AnnotateSystemNamespace(lib.GetClusterID(), utils.CloudName, lib.GetClusterName())
	c.AddNamespaceEventHandler(stopCh)
	objNameMap.InitMap()

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
	netInfoName := objNameMap.GenerateName("testnetinfo")
	segPathName := objNameMap.GenerateName("testSeg")
	gatewayPathName := objNameMap.GenerateName("testGW")
	testData[0].SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "nsx.vmware.com/v1alpha1",
		"kind":       "namespacenetworkinfos",
		"metadata": map[string]interface{}{
			"name":      netInfoName,
			"namespace": "default",
		},
		"topology": map[string]interface{}{
			"aviSegmentPath": segPathName,
			"gatewayPath":    gatewayPathName,
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

	infraSettingName := lib.GetAviInfraSettingName(gatewayPathName)

	g.Eventually(func() bool {
		if infraSetting, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().AviInfraSettings().Get(context.TODO(), infraSettingName, metav1.GetOptions{}); err != nil {
			return false
		} else {
			return *infraSetting.Spec.NSXSettings.T1LR == gatewayPathName &&
				infraSetting.Spec.Network.VipNetworks[0].NetworkName == lib.GetVCFNetworkName() &&
				infraSetting.Spec.SeGroup.Name == lib.GetClusterID()
		}
	}, 30*time.Second).Should(gomega.Equal(true))

	g.Eventually(func() bool {
		if ns, err := kubeClient.CoreV1().Namespaces().Get(context.TODO(), "default", metav1.GetOptions{}); err != nil {
			return false
		} else {
			return ns.Annotations[lib.InfraSettingNameAnnotation] == infraSettingName
		}
	}, 10*time.Second).Should(gomega.Equal(true))

	err = dynamicClient.Resource(lib.NetworkInfoGVR).Namespace("default").Delete(context.TODO(), netInfoName, v1.DeleteOptions{})
	if err != nil {
		t.Fatalf("failed to delete namespacenetworkinfo CR, error: %s", err.Error())
	}

	g.Eventually(func() bool {
		if _, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().AviInfraSettings().Get(context.TODO(), infraSettingName, metav1.GetOptions{}); err != nil {
			return false
		}
		return true
	}, 10*time.Second).Should(gomega.Equal(false))

	g.Eventually(func() bool {
		if ns, err := kubeClient.CoreV1().Namespaces().Get(context.TODO(), "default", metav1.GetOptions{}); err != nil {
			return false
		} else {
			return ns.Annotations[lib.InfraSettingNameAnnotation] != infraSettingName
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
	gatewayPathName := objNameMap.GenerateName("testGW")
	vpcNetConfig := objNameMap.GenerateName("testvpcnetworkconfig")
	testData[0].SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "crd.nsx.vmware.com/v1alpha1",
		"kind":       "vpcnetworkconfigurations",
		"metadata": map[string]interface{}{
			"name": vpcNetConfig,
		},
		"status": map[string]interface{}{
			"vpcs": []interface{}{
				map[string]interface{}{
					"name":         "vpc1",
					"lbSubnetPath": "/orgs/default/projects/test-project/vpcs/" + gatewayPathName + "/subnets/_AVI_SUBNET--LB",
				},
			},
		},
	})

	setupInfraTest(testData)
	annotateNamespaceWithVpcNetworkConfigCR(t, "default", vpcNetConfig)

	os.Setenv("VPC_MODE", "true")

	c := ingestion.SharedVCFK8sController()
	c.InitNetworkingHandler()
	c.AddNetworkInfoEventHandler(make(chan struct{}))
	worker := c.InitFullSyncWorker()
	go worker.Run()

	_, err := dynamicClient.Resource(lib.VPCNetworkConfigurationGVR).Create(context.TODO(), testData[0], v1.CreateOptions{})
	if err != nil {
		t.Fatalf("failed to create namespacenetworkinfos CR, error: %s", err.Error())
	}

	infraSettingName := lib.GetAviInfraSettingName("test-project" + gatewayPathName)

	g.Eventually(func() bool {
		if infraSetting, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().AviInfraSettings().Get(context.TODO(), infraSettingName, metav1.GetOptions{}); err != nil {
			return false
		} else {
			return *infraSetting.Spec.NSXSettings.T1LR == "/orgs/default/projects/test-project/vpcs/"+gatewayPathName &&
				len(infraSetting.Spec.Network.VipNetworks) == 0 &&
				infraSetting.Spec.SeGroup.Name == "Default-Group"
		}
	}, 30*time.Second).Should(gomega.Equal(true))

	g.Eventually(func() bool {
		if ns, err := kubeClient.CoreV1().Namespaces().Get(context.TODO(), "default", metav1.GetOptions{}); err != nil {
			return false
		} else {
			return ns.Annotations[lib.InfraSettingNameAnnotation] == infraSettingName &&
				ns.Annotations[lib.TenantAnnotation] == "test-project"
		}
	}, 10*time.Second).Should(gomega.Equal(true))

	err = dynamicClient.Resource(lib.VPCNetworkConfigurationGVR).Delete(context.TODO(), vpcNetConfig, v1.DeleteOptions{})
	if err != nil {
		t.Fatalf("failed to delete namespacenetworkinfo CR, error: %s", err.Error())
	}

	g.Eventually(func() bool {
		if _, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().AviInfraSettings().Get(context.TODO(), infraSettingName, metav1.GetOptions{}); err != nil {
			return false
		}
		return true
	}, 10*time.Second).Should(gomega.Equal(false))

	g.Eventually(func() bool {
		if ns, err := kubeClient.CoreV1().Namespaces().Get(context.TODO(), "default", metav1.GetOptions{}); err != nil {
			return false
		} else {
			return ns.Annotations[lib.InfraSettingNameAnnotation] != infraSettingName
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
	netInfoName := objNameMap.GenerateName("testnetinfo")
	segPathName := objNameMap.GenerateName("testSeg")
	gatewayPathName := objNameMap.GenerateName("testGW")
	testData[0].SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "nsx.vmware.com/v1alpha1",
		"kind":       "namespacenetworkinfos",
		"metadata": map[string]interface{}{
			"name":      netInfoName,
			"namespace": "default",
		},
		"topology": map[string]interface{}{
			"aviSegmentPath": segPathName,
			"gatewayPath":    gatewayPathName,
		},
	})
	testData = append(testData, &unstructured.Unstructured{})
	testData[1].SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "nsx.vmware.com/v1alpha1",
		"kind":       "namespacenetworkinfos",
		"metadata": map[string]interface{}{
			"name":      netInfoName,
			"namespace": "red",
		},
		"topology": map[string]interface{}{
			"aviSegmentPath": segPathName,
			"gatewayPath":    gatewayPathName,
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
			"name":      netInfoName,
			"namespace": "red-ns",
		},
		"topology": map[string]interface{}{
			"aviSegmentPath": segPathName,
			"gatewayPath":    gatewayPathName,
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

	infraSettingName1 := lib.GetAviInfraSettingName(gatewayPathName)
	infraSettingName2 := "red-" + lib.GetAviInfraSettingName(gatewayPathName)

	g.Eventually(func() bool {
		if infraSetting, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().AviInfraSettings().Get(context.TODO(), infraSettingName1, metav1.GetOptions{}); err != nil {
			return false
		} else {
			return *infraSetting.Spec.NSXSettings.T1LR == gatewayPathName &&
				infraSetting.Spec.Network.VipNetworks[0].NetworkName == lib.GetVCFNetworkName() &&
				infraSetting.Spec.SeGroup.Name == lib.GetClusterID()
		}
	}, 30*time.Second).Should(gomega.Equal(true))

	g.Eventually(func() bool {
		if infraSetting, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().AviInfraSettings().Get(context.TODO(), infraSettingName2, metav1.GetOptions{}); err != nil {
			t.Logf("setting not found %+v", err)
			return false
		} else {
			t.Logf("found but %+v", *infraSetting.Spec.NSXSettings.T1LR)
			t.Logf("found but %+v", infraSetting.Spec.Network.VipNetworks[0].NetworkName)
			t.Logf("found but %+v", infraSetting.Spec.SeGroup.Name)
			return *infraSetting.Spec.NSXSettings.T1LR == gatewayPathName &&
				infraSetting.Spec.Network.VipNetworks[0].NetworkName == lib.GetVCFNetworkNameWithNS("red") &&
				infraSetting.Spec.SeGroup.Name == lib.GetClusterID()
		}
	}, 30*time.Second).Should(gomega.Equal(true))

	g.Eventually(func() bool {
		if ns, err := kubeClient.CoreV1().Namespaces().Get(context.TODO(), "default", metav1.GetOptions{}); err != nil {
			return false
		} else {
			return ns.Annotations[lib.InfraSettingNameAnnotation] == infraSettingName1
		}
	}, 10*time.Second).Should(gomega.Equal(true))

	g.Eventually(func() bool {
		if ns, err := kubeClient.CoreV1().Namespaces().Get(context.TODO(), "red", metav1.GetOptions{}); err != nil {
			return false
		} else {
			return ns.Annotations[lib.InfraSettingNameAnnotation] == infraSettingName2
		}
	}, 10*time.Second).Should(gomega.Equal(true))

	g.Eventually(func() bool {
		if ns, err := kubeClient.CoreV1().Namespaces().Get(context.TODO(), "red-ns", metav1.GetOptions{}); err != nil {
			return false
		} else {
			return ns.Annotations[lib.InfraSettingNameAnnotation] == infraSettingName1
		}
	}, 10*time.Second).Should(gomega.Equal(true))

	err = dynamicClient.Resource(lib.NetworkInfoGVR).Namespace("default").Delete(context.TODO(), netInfoName, v1.DeleteOptions{})
	if err != nil {
		t.Fatalf("failed to delete namespacenetworkinfo CR, error: %s", err.Error())
	}

	err = dynamicClient.Resource(lib.NetworkInfoGVR).Namespace("red").Delete(context.TODO(), netInfoName, v1.DeleteOptions{})
	if err != nil {
		t.Fatalf("failed to delete namespacenetworkinfo CR, error: %s", err.Error())
	}

	err = dynamicClient.Resource(lib.NetworkInfoGVR).Namespace("red-ns").Delete(context.TODO(), netInfoName, v1.DeleteOptions{})
	if err != nil {
		t.Fatalf("failed to delete namespacenetworkinfo CR, error: %s", err.Error())
	}

	g.Eventually(func() bool {
		if _, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().AviInfraSettings().Get(context.TODO(), infraSettingName1, metav1.GetOptions{}); err != nil {
			return false
		}
		return true
	}, 10*time.Second).Should(gomega.Equal(false))

	g.Eventually(func() bool {
		if _, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().AviInfraSettings().Get(context.TODO(), infraSettingName2, metav1.GetOptions{}); err != nil {
			return false
		}
		return true
	}, 10*time.Second).Should(gomega.Equal(false))

	g.Eventually(func() bool {
		if ns, err := kubeClient.CoreV1().Namespaces().Get(context.TODO(), "default", metav1.GetOptions{}); err != nil {
			return false
		} else {
			return ns.Annotations[lib.InfraSettingNameAnnotation] != infraSettingName1
		}
	}, 10*time.Second).Should(gomega.Equal(true))
	g.Eventually(func() bool {
		if ns, err := kubeClient.CoreV1().Namespaces().Get(context.TODO(), "red", metav1.GetOptions{}); err != nil {
			return false
		} else {
			return ns.Annotations[lib.InfraSettingNameAnnotation] != infraSettingName2
		}
	}, 10*time.Second).Should(gomega.Equal(true))
	g.Eventually(func() bool {
		if ns, err := kubeClient.CoreV1().Namespaces().Get(context.TODO(), "red-ns", metav1.GetOptions{}); err != nil {
			return false
		} else {
			return ns.Annotations[lib.InfraSettingNameAnnotation] != infraSettingName1
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
	gatewayPathName := objNameMap.GenerateName("testGW")
	gatewayRedPathName := gatewayPathName + "-red"
	vpcNetConfig := objNameMap.GenerateName("testvpcnetworkconfig")
	vpcRedNetConfig := vpcNetConfig + "-red"
	testData[0].SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "crd.nsx.vmware.com/v1alpha1",
		"kind":       "vpcnetworkconfigurations",
		"metadata": map[string]interface{}{
			"name": vpcNetConfig,
		},
		"status": map[string]interface{}{
			"vpcs": []interface{}{
				map[string]interface{}{
					"name":         "vpc1",
					"lbSubnetPath": "/orgs/default/projects/test-project/vpcs/" + gatewayPathName + "/subnets/_AVI_SUBNET--LB",
				},
			},
		},
	})
	testData = append(testData, &unstructured.Unstructured{})
	testData[1].SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "crd.nsx.vmware.com/v1alpha1",
		"kind":       "vpcnetworkconfigurations",
		"metadata": map[string]interface{}{
			"name": vpcRedNetConfig,
		},
		"status": map[string]interface{}{
			"vpcs": []interface{}{
				map[string]interface{}{
					"name":         "vpc1",
					"lbSubnetPath": "/orgs/default/projects/test-project/vpcs/" + gatewayRedPathName + "/subnets/_AVI_SUBNET--LB",
				},
			},
		},
	})

	annotateNamespaceWithVpcNetworkConfigCR(t, "default", vpcNetConfig)
	annotateNamespaceWithVpcNetworkConfigCR(t, "red-ns", vpcNetConfig)
	annotateNamespaceWithVpcNetworkConfigCR(t, "red", vpcRedNetConfig)

	setupInfraTest(testData)

	os.Setenv("VPC_MODE", "true")

	c := ingestion.SharedVCFK8sController()
	c.InitNetworkingHandler()
	c.AddNetworkInfoEventHandler(make(chan struct{}))
	worker := c.InitFullSyncWorker()
	go worker.Run()

	_, err := dynamicClient.Resource(lib.VPCNetworkConfigurationGVR).Create(context.TODO(), testData[0], v1.CreateOptions{})
	if err != nil {
		t.Fatalf("failed to create namespacenetworkinfos CR, error: %s", err.Error())
	}
	_, err = dynamicClient.Resource(lib.VPCNetworkConfigurationGVR).Create(context.TODO(), testData[1], v1.CreateOptions{})
	if err != nil {
		t.Fatalf("failed to create namespacenetworkinfos CR, error: %s", err.Error())
	}

	infraSettingName1 := lib.GetAviInfraSettingName("test-project" + gatewayPathName)
	infraSettingName2 := lib.GetAviInfraSettingName("test-project" + gatewayRedPathName)

	g.Eventually(func() bool {
		if infraSetting, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().AviInfraSettings().Get(context.TODO(), infraSettingName1, metav1.GetOptions{}); err != nil {
			return false
		} else {
			return *infraSetting.Spec.NSXSettings.T1LR == "/orgs/default/projects/test-project/vpcs/"+gatewayPathName &&
				len(infraSetting.Spec.Network.VipNetworks) == 0 &&
				infraSetting.Spec.SeGroup.Name == "Default-Group"
		}
	}, 30*time.Second).Should(gomega.Equal(true))

	g.Eventually(func() bool {
		if infraSetting, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().AviInfraSettings().Get(context.TODO(), infraSettingName2, metav1.GetOptions{}); err != nil {
			return false
		} else {
			return *infraSetting.Spec.NSXSettings.T1LR == "/orgs/default/projects/test-project/vpcs/"+gatewayRedPathName &&
				len(infraSetting.Spec.Network.VipNetworks) == 0 &&
				infraSetting.Spec.SeGroup.Name == "Default-Group"
		}
	}, 30*time.Second).Should(gomega.Equal(true))

	g.Eventually(func() bool {
		if ns, err := kubeClient.CoreV1().Namespaces().Get(context.TODO(), "default", metav1.GetOptions{}); err != nil {
			return false
		} else {
			return ns.Annotations[lib.InfraSettingNameAnnotation] == infraSettingName1 &&
				ns.Annotations[lib.TenantAnnotation] == "test-project"
		}
	}, 10*time.Second).Should(gomega.Equal(true))

	g.Eventually(func() bool {
		if ns, err := kubeClient.CoreV1().Namespaces().Get(context.TODO(), "red", metav1.GetOptions{}); err != nil {
			return false
		} else {
			return ns.Annotations[lib.InfraSettingNameAnnotation] == infraSettingName2 &&
				ns.Annotations[lib.TenantAnnotation] == "test-project"
		}
	}, 10*time.Second).Should(gomega.Equal(true))

	g.Eventually(func() bool {
		if ns, err := kubeClient.CoreV1().Namespaces().Get(context.TODO(), "red-ns", metav1.GetOptions{}); err != nil {
			return false
		} else {
			return ns.Annotations[lib.InfraSettingNameAnnotation] == infraSettingName1 &&
				ns.Annotations[lib.TenantAnnotation] == "test-project"
		}
	}, 10*time.Second).Should(gomega.Equal(true))

	err = dynamicClient.Resource(lib.VPCNetworkConfigurationGVR).Delete(context.TODO(), vpcNetConfig, v1.DeleteOptions{})
	if err != nil {
		t.Fatalf("failed to delete namespacenetworkinfo CR, error: %s", err.Error())
	}

	err = dynamicClient.Resource(lib.VPCNetworkConfigurationGVR).Delete(context.TODO(), vpcRedNetConfig, v1.DeleteOptions{})
	if err != nil {
		t.Fatalf("failed to delete namespacenetworkinfo CR, error: %s", err.Error())
	}

	g.Eventually(func() bool {
		if _, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().AviInfraSettings().Get(context.TODO(), infraSettingName1, metav1.GetOptions{}); err != nil {
			return false
		}
		return true
	}, 10*time.Second).Should(gomega.Equal(false))

	g.Eventually(func() bool {
		if _, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().AviInfraSettings().Get(context.TODO(), infraSettingName2, metav1.GetOptions{}); err != nil {
			return false
		}
		return true
	}, 10*time.Second).Should(gomega.Equal(false))

	g.Eventually(func() bool {
		if ns, err := kubeClient.CoreV1().Namespaces().Get(context.TODO(), "default", metav1.GetOptions{}); err != nil {
			return false
		} else {
			return ns.Annotations[lib.InfraSettingNameAnnotation] != infraSettingName1
		}
	}, 10*time.Second).Should(gomega.Equal(true))
	g.Eventually(func() bool {
		if ns, err := kubeClient.CoreV1().Namespaces().Get(context.TODO(), "red", metav1.GetOptions{}); err != nil {
			return false
		} else {
			return ns.Annotations[lib.InfraSettingNameAnnotation] != infraSettingName2
		}
	}, 10*time.Second).Should(gomega.Equal(true))
	g.Eventually(func() bool {
		if ns, err := kubeClient.CoreV1().Namespaces().Get(context.TODO(), "red-ns", metav1.GetOptions{}); err != nil {
			return false
		} else {
			return ns.Annotations[lib.InfraSettingNameAnnotation] != infraSettingName1
		}
	}, 10*time.Second).Should(gomega.Equal(true))

	worker.Shutdown()
	os.Unsetenv("VPC_MODE")
}

func TestAKOInfraAviInfraSettingCreationVPCSEG(t *testing.T) {
	// create NS and annotate it with SEG and VPC CR,
	// AKO should create an AviInfraSetting for each unique vpcpath and SEG combination
	// NS should be annotated with AviInfraSetting
	// Update NS annotation for new SEG,
	// AKO should create an AviInfraSetting for each unique vpcpath and SEG combination
	// NS should be annotated with new AviInfraSetting
	// Older AviInfraSetting should get deleted
	// Update NS annotation for SEG  to empty string,
	// AKO should create an AviInfraSetting for each unique vpcpath and SEG combination
	// NS should be annotated with new AviInfraSetting
	// Older AviInfraSetting should get deleted
	g := gomega.NewGomegaWithT(t)
	var testData []*unstructured.Unstructured
	testData = append(testData, &unstructured.Unstructured{})
	gatewayPathName := objNameMap.GenerateName("testGW")
	vpcNetConfig := objNameMap.GenerateName("testvpcnetworkconfig")
	testData[0].SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "crd.nsx.vmware.com/v1alpha1",
		"kind":       "vpcnetworkconfigurations",
		"metadata": map[string]interface{}{
			"name": vpcNetConfig,
		},
		"status": map[string]interface{}{
			"vpcs": []interface{}{
				map[string]interface{}{
					"name":         "vpc1",
					"lbSubnetPath": "/orgs/default/projects/test-project/vpcs/" + gatewayPathName + "/subnets/_AVI_SUBNET--LB",
				},
			},
		},
	})

	setupInfraTest(testData)
	annotateNamespaceWithSEG(t, "default", "AVISEG1")
	annotateNamespaceWithVpcNetworkConfigCR(t, "default", vpcNetConfig)

	os.Setenv("VPC_MODE", "true")

	c := ingestion.SharedVCFK8sController()
	c.InitNetworkingHandler()
	c.AddNetworkInfoEventHandler(make(chan struct{}))
	worker := c.InitFullSyncWorker()
	go worker.Run()

	_, err := dynamicClient.Resource(lib.VPCNetworkConfigurationGVR).Create(context.TODO(), testData[0], v1.CreateOptions{})
	if err != nil {
		t.Fatalf("failed to create namespacenetworkinfos CR, error: %s", err.Error())
	}

	infraSettingName := lib.GetAviInfraSettingName("test-project" + gatewayPathName + "AVISEG1")

	g.Eventually(func() bool {
		if infraSetting, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().AviInfraSettings().Get(context.TODO(), infraSettingName, metav1.GetOptions{}); err != nil {
			return false
		} else {
			return *infraSetting.Spec.NSXSettings.T1LR == "/orgs/default/projects/test-project/vpcs/"+gatewayPathName &&
				len(infraSetting.Spec.Network.VipNetworks) == 0 &&
				infraSetting.Spec.SeGroup.Name == "AVISEG1"
		}
	}, 30*time.Second).Should(gomega.Equal(true))

	g.Eventually(func() bool {
		if ns, err := kubeClient.CoreV1().Namespaces().Get(context.TODO(), "default", metav1.GetOptions{}); err != nil {
			return false
		} else {
			return ns.Annotations[lib.InfraSettingNameAnnotation] == infraSettingName &&
				ns.Annotations[lib.TenantAnnotation] == "test-project"
		}
	}, 10*time.Second).Should(gomega.Equal(true))

	annotateNamespaceWithSEG(t, "default", "AVISEG2")
	infraSettingName2 := lib.GetAviInfraSettingName("test-project" + gatewayPathName + "AVISEG2")

	g.Eventually(func() bool {
		if infraSetting, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().AviInfraSettings().Get(context.TODO(), infraSettingName2, metav1.GetOptions{}); err != nil {
			return false
		} else {
			return *infraSetting.Spec.NSXSettings.T1LR == "/orgs/default/projects/test-project/vpcs/"+gatewayPathName &&
				len(infraSetting.Spec.Network.VipNetworks) == 0 &&
				infraSetting.Spec.SeGroup.Name == "AVISEG2"
		}
	}, 30*time.Second).Should(gomega.Equal(true))

	g.Eventually(func() bool {
		if ns, err := kubeClient.CoreV1().Namespaces().Get(context.TODO(), "default", metav1.GetOptions{}); err != nil {
			return false
		} else {
			return ns.Annotations[lib.InfraSettingNameAnnotation] == infraSettingName2 &&
				ns.Annotations[lib.TenantAnnotation] == "test-project"
		}
	}, 10*time.Second).Should(gomega.Equal(true))

	g.Eventually(func() bool {
		if _, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().AviInfraSettings().Get(context.TODO(), infraSettingName, metav1.GetOptions{}); err != nil {
			return false
		}
		return true
	}, 10*time.Second).Should(gomega.Equal(false))

	annotateNamespaceWithSEG(t, "default", "")
	infraSettingName3 := lib.GetAviInfraSettingName("test-project" + gatewayPathName)

	g.Eventually(func() bool {
		if infraSetting, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().AviInfraSettings().Get(context.TODO(), infraSettingName3, metav1.GetOptions{}); err != nil {
			return false
		} else {
			return *infraSetting.Spec.NSXSettings.T1LR == "/orgs/default/projects/test-project/vpcs/"+gatewayPathName &&
				len(infraSetting.Spec.Network.VipNetworks) == 0 &&
				infraSetting.Spec.SeGroup.Name == "Default-Group"
		}
	}, 30*time.Second).Should(gomega.Equal(true))

	g.Eventually(func() bool {
		if ns, err := kubeClient.CoreV1().Namespaces().Get(context.TODO(), "default", metav1.GetOptions{}); err != nil {
			return false
		} else {
			return ns.Annotations[lib.InfraSettingNameAnnotation] == infraSettingName3 &&
				ns.Annotations[lib.TenantAnnotation] == "test-project"
		}
	}, 10*time.Second).Should(gomega.Equal(true))

	g.Eventually(func() bool {
		if _, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().AviInfraSettings().Get(context.TODO(), infraSettingName2, metav1.GetOptions{}); err != nil {
			return false
		}
		return true
	}, 10*time.Second).Should(gomega.Equal(false))

	err = dynamicClient.Resource(lib.VPCNetworkConfigurationGVR).Delete(context.TODO(), vpcNetConfig, v1.DeleteOptions{})
	if err != nil {
		t.Fatalf("failed to delete namespacenetworkinfo CR, error: %s", err.Error())
	}

	g.Eventually(func() bool {
		if _, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().AviInfraSettings().Get(context.TODO(), infraSettingName3, metav1.GetOptions{}); err != nil {
			return false
		}
		return true
	}, 10*time.Second).Should(gomega.Equal(false))

	g.Eventually(func() bool {
		if ns, err := kubeClient.CoreV1().Namespaces().Get(context.TODO(), "default", metav1.GetOptions{}); err != nil {
			return false
		} else {
			_, exist := ns.Annotations[lib.InfraSettingNameAnnotation]
			return !exist
		}
	}, 10*time.Second).Should(gomega.Equal(true))

	worker.Shutdown()
	os.Unsetenv("VPC_MODE")
}

func TestAKOInfraDeriveCloudNameT1(t *testing.T) {
	origCloud := ""
	lib.SetAKOUser(lib.AKOPrefix)
	tz := "/infra/sites/default/enforcement-points/default/transport-zones/1b3a2f36-bfd1-443e-a0f6-4de01abc963e"
	a := ingestion.NewAviControllerInfra(kubeClient)
	if ns, err := kubeClient.CoreV1().Namespaces().Get(context.TODO(), utils.GetAKONamespace(), metav1.GetOptions{}); err != nil {
		t.Fatalf("Error occurred while GET namespace: %v", err)
	} else {
		origCloud = ns.Annotations[lib.WCPCloud]
	}
	annotateNamespaceWithCloud(t, utils.GetAKONamespace(), "CLOUD_NSXT")
	if ns, err := kubeClient.CoreV1().Namespaces().Get(context.TODO(), utils.GetAKONamespace(), metav1.GetOptions{}); err != nil {
		t.Fatalf("Error occurred while GET namespace: %v", err)
	} else {
		if ns.Annotations[lib.WCPCloud] != "CLOUD_NSXT" {
			t.Fatalf("NS cloud annotation update to CLOUD_NSXT failed")
		}
	}
	aviCloud, err := a.DeriveCloudMappedToTZ(tz)
	if err != nil || *aviCloud.Name != "CLOUD_NSXT" {
		t.Fatalf("Cloud name derivation using NS annotation failed: %v", err)
	}
	annotateNamespaceWithCloud(t, utils.GetAKONamespace(), "CLOUD_VCENTER")
	if ns, err := kubeClient.CoreV1().Namespaces().Get(context.TODO(), utils.GetAKONamespace(), metav1.GetOptions{}); err != nil {
		t.Fatalf("Error occurred while GET namespace: %v", err)
	} else {
		if ns.Annotations[lib.WCPCloud] != "CLOUD_VCENTER" {
			t.Fatalf("NS cloud annotation update to CLOUD_VCENTER failed")
		}
	}
	aviCloud, err = a.DeriveCloudMappedToTZ(tz)
	if err == nil {
		t.Fatalf("Cloud name derivation using NS annotation passed for wrong cloud")
	}

	annotateNamespaceWithCloud(t, utils.GetAKONamespace(), "")
	if ns, err := kubeClient.CoreV1().Namespaces().Get(context.TODO(), utils.GetAKONamespace(), metav1.GetOptions{}); err != nil {
		t.Fatalf("Error occurred while GET namespace: %v", err)
	} else if ns.Annotations[lib.WCPCloud] != "" {
		t.Fatalf("NS cloud annotation update to empty failed")
	}

	aviCloud, err = a.DeriveCloudMappedToTZ(tz)
	if err != nil || *aviCloud.Name != "CLOUD_NSXT" {
		t.Fatalf("Cloud name derivation using VS failed: %v", err)
	}

	annotateNamespaceWithCloud(t, utils.GetAKONamespace(), origCloud)
	if ns, err := kubeClient.CoreV1().Namespaces().Get(context.TODO(), utils.GetAKONamespace(), metav1.GetOptions{}); err != nil {
		t.Fatalf("Error occurred while GET namespace: %v", err)
	} else {
		if ns.Annotations[lib.WCPCloud] != origCloud {
			t.Fatalf("NS cloud annotation update to %s failed", origCloud)
		}
	}
}

func injectMWNSXVPCCoud() {
	integrationtest.AddMiddleware(func(w http.ResponseWriter, r *http.Request) {
		url := r.URL.EscapedPath()
		mockFilePath := integrationtest.DefaultMockFilePath
		if r.Method == "GET" && strings.Contains(url, "/api/cloud/") {
			data, _ := os.ReadFile(fmt.Sprintf("%s/%s_mock.json", mockFilePath, "CLOUD_NSXT1"))
			w.WriteHeader(http.StatusOK)
			w.Write(data)
		}
	})
}

func TestAKOInfraDeriveCloudNameVPC(t *testing.T) {
	// Add Cloud Annotation to the  AKO NS
	// AKO should pick the cloud from AKO NS annotation
	// Add wrong cloud annotation to AKO NS
	// AKO should fail to pick the cloud
	// Remove the NS cloud annotation
	// AKO should pick the cloud from kube lbservice VS
	origCloud := ""
	os.Setenv("VPC_MODE", "true")
	lib.SetAKOUser(lib.AKOPrefix)
	a := ingestion.NewAviControllerInfra(kubeClient)
	if ns, err := kubeClient.CoreV1().Namespaces().Get(context.TODO(), utils.GetAKONamespace(), metav1.GetOptions{}); err != nil {
		t.Fatalf("Error occurred while GET namespace: %v", err)
	} else {
		origCloud = ns.Annotations[lib.WCPCloud]
	}
	annotateNamespaceWithCloud(t, utils.GetAKONamespace(), "CLOUD_NSXT1")
	if ns, err := kubeClient.CoreV1().Namespaces().Get(context.TODO(), utils.GetAKONamespace(), metav1.GetOptions{}); err != nil {
		t.Fatalf("Error occurred while GET namespace: %v", err)
	} else {
		if ns.Annotations[lib.WCPCloud] != "CLOUD_NSXT1" {
			t.Fatalf("NS cloud annotation update to CLOUD_NSXT1 failed")
		}
	}
	aviCloud, err := a.DeriveCloudMappedToTZ("")
	if err != nil || *aviCloud.Name != "CLOUD_NSXT1" {
		t.Fatalf("Cloud name derivation using NS annotation failed: %v", err)
	}
	annotateNamespaceWithCloud(t, utils.GetAKONamespace(), "CLOUD_VCENTER")
	if ns, err := kubeClient.CoreV1().Namespaces().Get(context.TODO(), utils.GetAKONamespace(), metav1.GetOptions{}); err != nil {
		t.Fatalf("Error occurred while GET namespace: %v", err)
	} else {
		if ns.Annotations[lib.WCPCloud] != "CLOUD_VCENTER" {
			t.Fatalf("NS cloud annotation update to CLOUD_VCENTER failed")
		}
	}
	aviCloud, err = a.DeriveCloudMappedToTZ("")
	if err == nil {
		t.Fatalf("Cloud name derivation using NS annotation passed for wrong cloud")
	}

	annotateNamespaceWithCloud(t, utils.GetAKONamespace(), "")
	if ns, err := kubeClient.CoreV1().Namespaces().Get(context.TODO(), utils.GetAKONamespace(), metav1.GetOptions{}); err != nil {
		t.Fatalf("Error occurred while GET namespace: %v", err)
	} else if ns.Annotations[lib.WCPCloud] != "" {
		t.Fatalf("NS cloud annotation update to empty failed")
	}

	injectMWNSXVPCCoud()

	aviCloud, err = a.DeriveCloudMappedToTZ("")
	if err != nil || *aviCloud.Name != "CLOUD_NSXT1" {
		t.Fatalf("Cloud name derivation failed: %v", err)
	}

	annotateNamespaceWithCloud(t, utils.GetAKONamespace(), origCloud)
	if ns, err := kubeClient.CoreV1().Namespaces().Get(context.TODO(), utils.GetAKONamespace(), metav1.GetOptions{}); err != nil {
		t.Fatalf("Error occurred while GET namespace: %v", err)
	} else {
		if ns.Annotations[lib.WCPCloud] != origCloud {
			t.Fatalf("NS cloud annotation update to %s failed", origCloud)
		}
	}
	os.Unsetenv("VPC_MODE")
	integrationtest.ResetMiddleware()
}
