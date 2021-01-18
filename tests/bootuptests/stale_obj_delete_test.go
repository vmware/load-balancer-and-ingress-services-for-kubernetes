package bootuptests

import (
	"context"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/cache"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/k8s"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/tests/integrationtest"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sfake "k8s.io/client-go/kubernetes/fake"

	crdfake "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/client/v1alpha1/clientset/versioned/fake"
)

var KubeClient *k8sfake.Clientset
var CRDClient *crdfake.Clientset
var restChan chan bool
var uuidMap map[string]bool

const mockFilePath = "bootupmock"
const invalidFilePath = "invalidmock1"

func TestMain(m *testing.M) {
	os.Setenv("INGRESS_API", "extensionv1")
	os.Setenv("NETWORK_NAME", "net123")
	os.Setenv("CLUSTER_NAME", "cluster")
	os.Setenv("CLOUD_NAME", "CLOUD_VCENTER")
	os.Setenv("SEG_NAME", "Default-Group")
	os.Setenv("NODE_NETWORK_LIST", `[{"networkName":"net123","cidrs":["10.79.168.0/22"]}]`)
	utils.CtrlVersion = "20.1.1"
	restChan = make(chan bool)
	uuidMap = make(map[string]bool)

	KubeClient = k8sfake.NewSimpleClientset()
	CRDClient = crdfake.NewSimpleClientset()
	lib.SetCRDClientset(CRDClient)

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
	k8s.NewCRDInformers(CRDClient)

	mcache := cache.SharedAviObjCache()
	cloudObj := &cache.AviCloudPropertyCache{Name: "Default-Cloud", VType: "mock"}
	subdomains := []string{"avi.internal", ".com"}
	cloudObj.NSIpamDNS = subdomains
	mcache.CloudKeyCache.AviCacheAdd("Default-Cloud", cloudObj)

	integrationtest.InitializeFakeAKOAPIServer()

	integrationtest.NewAviFakeClientInstance(true)
	defer integrationtest.AviFakeClientInstance.Close()

	os.Exit(m.Run())
}

func addConfigMap() {
	aviCM := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "avi-system",
			Name:      "avi-k8s-config",
		},
	}
	KubeClient.CoreV1().ConfigMaps("avi-system").Create(context.TODO(), aviCM, metav1.CreateOptions{})

}

func injectMWForObjDeletion() {
	integrationtest.AddMiddleware(func(w http.ResponseWriter, r *http.Request) {
		var finalResponse []byte
		url := r.URL.EscapedPath()
		object := strings.Split(strings.Trim(url, "/"), "/")
		if r.Method == "DELETE" {
			w.WriteHeader(http.StatusNoContent)
			w.Write(finalResponse)
			uuid := object[2]
			utils.AviLog.Infof("uuid of the object for deletion: %s", uuid)
			if _, found := uuidMap[uuid]; found {
				delete(uuidMap, uuid)
			} else {
				utils.AviLog.Warnf("unexpcted object for deletion: %s", uuid)
				restChan <- false
			}
			// We expect all objects to be deleted in the end
			if len(uuidMap) == 0 {
				restChan <- true
			}
		} else if r.Method == "GET" {
			integrationtest.FeedMockCollectionData(w, r, mockFilePath)

		} else if strings.Contains(url, "login") {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"success": "true"}`))
		} else if strings.Contains(url, "initial-data") {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"version": {"Version": "20.1.2"}}`))
		}
	})
}

func injectMWForCloud() {
	integrationtest.AddMiddleware(func(w http.ResponseWriter, r *http.Request) {
		url := r.URL.EscapedPath()
		if r.Method == "GET" && strings.Contains(url, "/api/cloud/") {
			integrationtest.FeedMockCollectionData(w, r, invalidFilePath)

		} else if strings.Contains(url, "initial-data") {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"version": {"Version": "20.1.2"}}`))
		} else if r.Method == "GET" {
			integrationtest.FeedMockCollectionData(w, r, mockFilePath)

		} else if strings.Contains(url, "login") {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"success": "true"}`))
		}
	})
}

// PopulateCache populates cache and triggers deletion of unused objects.
// In this case two pool and two vsvip objects. Among these, one vsvip is
// refeered by a Virtual Service, so we need to delete 3 objects
func TestObjDeletion(t *testing.T) {
	uuidMap["pool-e3b87aff-a9d7-44eb-9935-6fd9ab81a37c"] = true
	uuidMap["pool-11a38043-e51e-4c93-8187-b390d7d81abd"] = true
	uuidMap["vsvip-a590042a-358f-4693-bfa5-cb9d0c8c1931"] = true
	//uuidMap["vsvip-82b41dd7-5b19-4007-85d4-530acea4d86b"] = true

	injectMWForObjDeletion()
	addConfigMap()
	go k8s.PopulateCache()
	// DeleteConfigMap(t)
	integrationtest.ResetMiddleware()
}

// Injecting middleware to error out cloud properties cache update failure
func TestNetworkIssueCacheValidationDuringBootup(t *testing.T) {
	injectMWForCloud()
	err := k8s.PopulateCache()
	if err == nil {
		t.Fatalf("Cache validation failed.")
	}
	integrationtest.ResetMiddleware()
}
