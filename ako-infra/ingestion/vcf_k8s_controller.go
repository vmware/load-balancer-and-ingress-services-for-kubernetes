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

package ingestion

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-infra/avirest"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-infra/webhook"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	"github.com/vmware/alb-sdk/go/clients"
	"github.com/vmware/alb-sdk/go/session"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
)

var controllerInstance *VCFK8sController
var ctrlonce sync.Once
var tzonce sync.Once
var transportZone string

var WorkloadNamespaceCount int = 0
var countLock sync.RWMutex

type VCFK8sController struct {
	worker_id        uint32
	informers        *utils.Informers
	dynamicInformers *lib.DynamicInformers
	//workqueue        []workqueue.RateLimitingInterface
	DisableSync bool
	NetHandler  avirest.NetworkingHandler
}

func SharedVCFK8sController() *VCFK8sController {
	ctrlonce.Do(func() {
		controllerInstance = &VCFK8sController{
			worker_id:        (uint32(1) << utils.NumWorkersIngestion) - 1,
			informers:        utils.GetInformers(),
			dynamicInformers: lib.GetDynamicInformers(),
			DisableSync:      true,
		}
	})
	return controllerInstance
}

func (c *VCFK8sController) AddNamespaceEventHandler(stopCh <-chan struct{}) {
	// Saves the initial workload namespace count during reboot,
	// before the config handlers are started.
	if err := c.addWorkloadNamespaceCount(); err != nil {
		utils.AviLog.Fatalf("Unable to list Namespaces: %s", err.Error())
	}

	namespaceHandler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			utils.AviLog.Infof("Namespace ADD Event")
			c.handleNamespaceAdd()
		},
		UpdateFunc: func(old, obj interface{}) {
			utils.AviLog.Infof("Namespace Update Event")
		},
		DeleteFunc: func(obj interface{}) {
			utils.AviLog.Infof("Namespace Delete Event")
			_, ok := obj.(*corev1.Namespace)
			if !ok {
				crd, _ := obj.(*unstructured.Unstructured)
				_, found, err := unstructured.NestedStringMap(crd.UnstructuredContent(), "spec")
				if err != nil || !found {
					utils.AviLog.Warnf("Namespace spec not found: %+v", err)
					return
				}
			}
			c.handleNamespaceDelete()
		},
	}
	c.informers.NSInformer.Informer().AddEventHandler(namespaceHandler)

	go c.informers.NSInformer.Informer().Run(stopCh)
	if !cache.WaitForCacheSync(stopCh, c.informers.NSInformer.Informer().HasSynced) {
		runtime.HandleError(fmt.Errorf("timed out waiting for caches to sync"))
	} else {
		utils.AviLog.Infof("Caches synced for Namespace informer")
	}
}

func (c *VCFK8sController) addWorkloadNamespaceCount() error {
	clientSet := c.informers.ClientSet
	nsList, err := clientSet.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	count := 0
	for _, ns := range nsList.Items {
		if _, ok := ns.Labels[VSphereClusterIDLabelKey]; ok {
			count += 1
		}
	}
	WorkloadNamespaceCount = count
	utils.AviLog.Infof("Initial number of workload namespaces: %d", WorkloadNamespaceCount)
	return nil
}

func (c *VCFK8sController) handleNamespaceAdd() {
	countLock.Lock()
	defer countLock.Unlock()
	count, err := c.getWorkloadNamespaceCount()
	if err != nil {
		return
	}

	// Only when before the addition, the count was 0, (and now it becomes more than 0),
	// we must reconfigure the SEG, by rebooting AKO. On reboot AKO ensures SEG configuration.
	if WorkloadNamespaceCount == 0 && count > 0 {
		utils.AviLog.Fatalf("First Workload Namespace added in cluster. Rebooting AKO for infra configuration.")
	}
	WorkloadNamespaceCount = count
}

func (c *VCFK8sController) handleNamespaceDelete() {
	countLock.Lock()
	count, err := c.getWorkloadNamespaceCount()
	if err != nil {
		countLock.Unlock()
		return
	}

	WorkloadNamespaceCount = count
	if count > 0 {
		utils.AviLog.Infof("%d Workload Namespace exist in the cluster. Skipping deconfiguration.", count)
		countLock.Unlock()
		return
	}
	countLock.Unlock()

	utils.AviLog.Infof("No Workload Namespace exist, proceeding with Avi infra deconfiguraiton.")

	// Fetch all service engines and delete them.
	if err := avirest.DeleteServiceEngines(); err != nil {
		utils.AviLog.Errorf("Unable to remove SEs %s", err.Error())
		return
	}

	// Delete service engine group.
	if err := avirest.DeleteServiceEngineGroup(); err != nil {
		utils.AviLog.Errorf("Unable to remove SEG %s", err.Error())
		return
	}
}

// Gets number of only the Workload Namespaces. Only the Workload Namespaces
// have the label with key vSphereClusterID in them, which is how we differentiate.
func (c *VCFK8sController) getWorkloadNamespaceCount() (int, error) {
	nsList, err := c.informers.NSInformer.Lister().List(labels.Set(nil).AsSelector())
	if err != nil {
		utils.AviLog.Error(nil, err.Error())
		return 0, err
	}
	count := 0
	for _, ns := range nsList {
		if _, ok := ns.Labels[VSphereClusterIDLabelKey]; ok {
			count += 1
		}
	}
	return count, nil
}

func (c *VCFK8sController) AddConfigMapEventHandler(stopCh <-chan struct{}, startSyncCh chan struct{}) {
	configmapHandler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			utils.AviLog.Infof("ConfigMap Add")
			if c.ValidBootStrapData() && startSyncCh != nil {
				startSyncCh <- struct{}{}
				startSyncCh = nil
			}
		},
		UpdateFunc: func(old, obj interface{}) {
			utils.AviLog.Infof("ConfigMap Update")
			if c.ValidBootStrapData() && startSyncCh != nil {
				startSyncCh <- struct{}{}
				startSyncCh = nil
			}
		},
	}

	c.informers.ConfigMapInformer.Informer().AddEventHandler(configmapHandler)
	go c.informers.ConfigMapInformer.Informer().Run(stopCh)
	if !cache.WaitForCacheSync(stopCh, c.informers.ConfigMapInformer.Informer().HasSynced) {
		runtime.HandleError(fmt.Errorf("timed out waiting for caches to sync"))
	} else {
		utils.AviLog.Infof("Caches synced for ConfigMap informer")
	}
}

func (c *VCFK8sController) AddAvailabilityZoneCREventHandler(stopCh <-chan struct{}) {
	azEventHandler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			utils.AviLog.Infof("Availability Zone ADD Event")
			updateSEGroup()
		},
		DeleteFunc: func(obj interface{}) {
			utils.AviLog.Infof("Availability Zone DELETE Event")
			updateSEGroup()
		},
	}
	c.dynamicInformers.AvailabilityZoneInformer.Informer().AddEventHandler(azEventHandler)
	go c.dynamicInformers.AvailabilityZoneInformer.Informer().Run(stopCh)
	if !cache.WaitForCacheSync(stopCh, c.dynamicInformers.AvailabilityZoneInformer.Informer().HasSynced) {
		runtime.HandleError(fmt.Errorf("timed out waiting for availability zones caches to sync"))
	} else {
		utils.AviLog.Infof("Caches synced for availability zone informer")
	}
}

func (c *VCFK8sController) AddNetworkInfoEventHandler(stopCh <-chan struct{}) {
	c.NetHandler.AddNetworkInfoEventHandler(stopCh)
}

func (c *VCFK8sController) AddVKSCapabilityEventHandler(stopCh <-chan struct{}) {
	capabilityActive := lib.IsVKSCapabilityActivated()
	utils.AviLog.Infof("VKS capability: informer starting, initial state activated=%t", capabilityActive)

	capabilityEventHandler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			utils.AviLog.Infof("SupervisorCapability ADD Event")
			if lib.IsVKSCapabilityActivated() && !capabilityActive {
				utils.AviLog.Infof("VKS capability activated")
				capabilityActive = true
				go webhook.StartVKSWebhook(utils.GetInformers().ClientSet, stopCh)
			}
		},
		UpdateFunc: func(old, obj interface{}) {
			utils.AviLog.Infof("SupervisorCapability UPDATE Event")
			if lib.IsVKSCapabilityActivated() && !capabilityActive {
				utils.AviLog.Infof("VKS capability activated")
				capabilityActive = true
				go webhook.StartVKSWebhook(utils.GetInformers().ClientSet, stopCh)
			}
		},
		DeleteFunc: func(obj interface{}) {
			utils.AviLog.Infof("SupervisorCapability DELETE Event")
		},
	}

	c.dynamicInformers.SupervisorCapabilityInformer.Informer().AddEventHandler(capabilityEventHandler)
	go c.dynamicInformers.SupervisorCapabilityInformer.Informer().Run(stopCh)
	if !cache.WaitForCacheSync(stopCh, c.dynamicInformers.SupervisorCapabilityInformer.Informer().HasSynced) {
		runtime.HandleError(fmt.Errorf("timed out waiting for SupervisorCapability caches to sync"))
	} else {
		utils.AviLog.Infof("VKS capability: caches synced for SupervisorCapability informer")
	}
}

// HandleVCF checks if avi secret used by AKO is already present. If found, then it would try to connect to
// AVI Controller. If there is any failure, we would look at Bootstrap CR used by NCP to communicate with AKO.
// If Bootstrap CR is not found, AKO would wait for it to be created. If the authtoken from Bootstrap CR
// can be used to connect to the AVI Controller, then avi-secret would be created with that token.
func (c *VCFK8sController) HandleVCF(stopCh <-chan struct{}, ctrlCh chan struct{}, skipAviClient ...bool) string {
	startSyncCh := make(chan struct{})
	if !c.ValidBootStrapData() {
		c.AddConfigMapEventHandler(stopCh, startSyncCh)
		utils.AviLog.Infof("Running in a VCF Cluster, but valid ConfigMap/Secret not found, waiting ..")
		ticker := time.NewTicker(lib.FullSyncInterval * time.Second)
	L:
		for {
			select {
			case <-startSyncCh:
				ticker.Stop()
				break L
			case <-ctrlCh:
				return transportZone
			case <-ticker.C:
				if c.ValidBootStrapData() {
					ticker.Stop()
					break L
				}
			}
		}
	}

	c.AddConfigMapEventHandler(stopCh, nil)
	utils.AviLog.Infof("Bootstrap information found, continuing AKO initialization")
	return transportZone
}

func (c *VCFK8sController) ValidBootStrapData() bool {
	cs := c.informers.ClientSet
	configmap, err := cs.CoreV1().ConfigMaps(utils.VMWARE_SYSTEM_AKO).Get(context.TODO(), "avi-k8s-config", metav1.GetOptions{})
	if err != nil {
		utils.AviLog.Warnf("Failed to get ConfigMap, got err: %v", err)
		lib.AKOControlConfig().PodEventf(corev1.EventTypeWarning, "ConfigMapNotFound", err.Error())
		return false
	}

	clusterID := configmap.Data["clusterID"]
	controllerIP := configmap.Data["controllerIP"]
	secretName := configmap.Data["credentialsSecretName"]
	secretNamespace := configmap.Data["credentialsSecretNamespace"]

	// The transport zone is used in order to identify the cloud in Avi controller.
	// We take the cloudName from the configmap in case of a VCF cluster which would
	// in fact have the transportZone information.
	transportzone := configmap.Data["cloudName"]
	utils.AviLog.Infof("Got data from ConfigMap %v", utils.Stringify(configmap.Data))
	if clusterID == "" || controllerIP == "" || secretName == "" || secretNamespace == "" {
		utils.AviLog.Infof("ConfigMap data insufficient")
		lib.AKOControlConfig().PodEventf(corev1.EventTypeWarning, "ConfigMapDataInsufficient", "ConfigMap data insufficient")
		return false
	}
	// In case of VCF cluster in VPC mode, transport zone is not defined in the config map
	if transportzone == "" && !lib.GetVPCMode() {
		utils.AviLog.Infof("ConfigMap data insufficient, transport zone not present")
		lib.AKOControlConfig().PodEventf(corev1.EventTypeWarning, "ConfigMapDataInsufficient", "ConfigMap data insufficient")
		return false
	}

	lib.SetClusterID(clusterID)
	setTranzportZone(transportzone)
	return c.ValidBootstrapSecretData(controllerIP, secretName, secretNamespace)
}

func (c *VCFK8sController) ValidBootstrapSecretData(controllerIP, secretName, secretNamespace string) bool {
	cs := c.informers.ClientSet
	ncpSecret, err := cs.CoreV1().Secrets(secretNamespace).Get(context.TODO(), secretName, metav1.GetOptions{})
	if err != nil {
		utils.AviLog.Warnf("Failed to get Secret, got err: %v", err)
		lib.AKOControlConfig().PodEventf(corev1.EventTypeWarning, "AviSecretNotFound", err.Error())
		return false
	}

	authToken := string(ncpSecret.Data["authtoken"])
	username := string(ncpSecret.Data["username"])
	caData := string(ncpSecret.Data["certificateAuthorityData"])
	lib.SetControllerIP(controllerIP)

	transport, isSecure := utils.GetHTTPTransportWithCert(caData)
	options := []func(*session.AviSession) error{
		session.SetAuthToken(string(authToken)),
		session.DisableControllerStatusCheckOnFailure(true),
		session.SetTransport(transport),
		session.SetTimeout(120 * time.Second),
	}
	if !isSecure {
		options = append(options, session.SetInsecure)
	}
	aviClient, err := clients.NewAviClient(controllerIP, username, options...)
	if err != nil {
		utils.AviLog.Errorf("Failed to connect to AVI controller using secret provided by NCP, the secret would be deleted, err: %v", err)
		c.deleteAviSecret(lib.AviInitSecret, secretNamespace)
		c.deleteAviSecret(secretName, secretNamespace)
		lib.AKOControlConfig().PodEventf(corev1.EventTypeWarning, "InvalidSecret", err.Error())
		return false
	}

	ctrlVersion := lib.GetControllerVersion()
	if ctrlVersion == "" {
		version, err := aviClient.AviSession.GetControllerVersion()
		if err != nil {
			utils.AviLog.Infof("Failed to get controller version from Avi session, err: %s", err)
			return false
		}
		maxVersion, err := utils.NewVersion(utils.MaxAviVersion)
		if err != nil {
			utils.AviLog.Errorf("Failed to create Version object, err: %s", err)
			return false
		}
		curVersion, err := utils.NewVersion(version)
		if err != nil {
			utils.AviLog.Errorf("Failed to create Version object, err: %s", err)
			return false
		}
		if curVersion.Compare(maxVersion) > 0 {
			utils.AviLog.Infof("Overwriting the controller version %s to max Avi version %s", version, utils.MaxAviVersion)
			version = utils.MaxAviVersion
		}
		ctrlVersion = version
	}
	SetVersion := session.SetVersion(ctrlVersion)
	SetVersion(aviClient.AviSession)

	avirest.InfraAviClientInstance(aviClient)
	utils.AviLog.Infof("Successfully connected to AVI controller using secret provided by NCP")
	return true
}

func (c *VCFK8sController) deleteAviSecret(name, ns string) {
	cs := c.informers.ClientSet
	err := cs.CoreV1().Secrets(ns).Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil {
		utils.AviLog.Warnf("Failed to delete secret: %s, namespace: %s, err: %v", name, ns, err.Error())
	}
}

func setTranzportZone(tzPath string) {
	tzonce.Do(func() {
		utils.AviLog.Infof("TransportZone to use for AKO is set to %s", tzPath)
		transportZone = tzPath
	})
}

func (c *VCFK8sController) AddSecretEventHandler(stopCh <-chan struct{}) {
	secretEventHandler := cache.ResourceEventHandlerFuncs{
		DeleteFunc: func(obj interface{}) {
			secret, ok := obj.(*corev1.Secret)
			if !ok {
				tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
				if !ok {
					utils.AviLog.Errorf("couldn't get object from tombstone %#v", obj)
					return
				}
				secret, ok = tombstone.Obj.(*corev1.Secret)
				if !ok {
					utils.AviLog.Errorf("Tombstone contained object that is not a Secret: %#v", obj)
					return
				}
			}
			if secret.Namespace == utils.GetAKONamespace() && secret.Name == lib.AviSecret {
				utils.AviLog.Fatalf("Avi Secret object %s/%s updated/deleted, shutting down AKO", secret.Namespace, secret.Name)
			}
		},
		UpdateFunc: func(old, cur interface{}) {
			oldobj := old.(*corev1.Secret)
			secret := cur.(*corev1.Secret)
			if oldobj.ResourceVersion != secret.ResourceVersion && !reflect.DeepEqual(secret.Data, oldobj.Data) {
				if secret.Namespace == utils.GetAKONamespace() && secret.Name == lib.AviSecret {
					utils.AviLog.Fatalf("Avi Secret object %s/%s updated/deleted, shutting down AKO", secret.Namespace, secret.Name)
				}
			}
		},
	}

	if c.informers.SecretInformer != nil {
		c.informers.SecretInformer.Informer().AddEventHandler(secretEventHandler)
	}
	go c.informers.SecretInformer.Informer().Run(stopCh)
	if !cache.WaitForCacheSync(stopCh, c.informers.SecretInformer.Informer().HasSynced) {
		runtime.HandleError(fmt.Errorf("timed out waiting for caches to sync"))
	} else {
		utils.AviLog.Infof("Caches synced for Secret informer")
	}
}

func (c *VCFK8sController) Sync() {
	c.NetHandler.SyncLSLRNetwork()
}

func (c *VCFK8sController) InitFullSyncWorker() *utils.FullSyncThread {
	worker := c.NetHandler.NewLRLSFullSyncWorker()
	return worker
}

func (c *VCFK8sController) InitNetworkingHandler() {
	if lib.GetVPCMode() {
		c.NetHandler = &avirest.VPCHandler{}
	} else {
		c.NetHandler = &avirest.T1LRNetworking{}
	}
}
