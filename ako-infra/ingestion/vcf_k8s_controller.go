/*
 * Copyright 2021 VMware, Inc.
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
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-infra/avirest"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/third_party/github.com/vmware/alb-sdk/go/clients"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/third_party/github.com/vmware/alb-sdk/go/session"

	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

var controllerInstance *VCFK8sController
var ctrlonce sync.Once
var tzonce sync.Once
var transportZone string

type VCFK8sController struct {
	worker_id        uint32
	informers        *utils.Informers
	dynamicInformers *lib.VCFDynamicInformers
	//workqueue        []workqueue.RateLimitingInterface
	DisableSync bool
}

type K8sinformers struct {
	Cs            kubernetes.Interface
	DynamicClient dynamic.Interface
}

func SharedVCFK8sController() *VCFK8sController {
	ctrlonce.Do(func() {
		controllerInstance = &VCFK8sController{
			worker_id:        (uint32(1) << utils.NumWorkersIngestion) - 1,
			informers:        utils.GetInformers(),
			dynamicInformers: lib.GetVCFDynamicInformers(),
			DisableSync:      true,
		}
	})
	return controllerInstance
}

// Run will set up the event handlers for types we are interested in, as well
// as syncing informer caches and starting workers. It will block until stopCh
// is closed, at which point it will shutdown the workqueue and wait for
// workers to finish processing their current work items.
func (c *VCFK8sController) Run(stopCh <-chan struct{}) error {
	defer runtime.HandleCrash()

	utils.AviLog.Info("Started the Kubernetes Controller")
	<-stopCh
	utils.AviLog.Info("Shutting down the Kubernetes Controller")
	return nil
}
func (c *VCFK8sController) AddNCPSecretEventHandler(k8sinfo K8sinformers, stopCh <-chan struct{}, startSyncCh chan struct{}) {
	NCPSecretHandler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if lib.VCFInitialized {
				return
			}
			data, ok := obj.(*corev1.Secret)
			if !ok || data.Namespace != utils.GetAKONamespace() {
				return
			}
			if c.ValidBootStrapData() && startSyncCh != nil {
				err := c.CreateOrUpdateAviSecret()
				if err != nil {
					utils.AviLog.Warnf("Failed to create or update AVI Secret, AKO would be rebooted")
					lib.ShutdownApi()
				} else {
					startSyncCh <- struct{}{}
					startSyncCh = nil
				}
			}
		},
		UpdateFunc: func(old, obj interface{}) {
			if lib.VCFInitialized {
				return
			}
			data, ok := obj.(*corev1.Secret)
			if !ok || data.Namespace != utils.GetAKONamespace() {
				return
			}
			if c.ValidBootStrapData() && startSyncCh != nil {
				err := c.CreateOrUpdateAviSecret()
				if err != nil {
					utils.AviLog.Warnf("Failed to create or update AVI Secret, AKO would be rebooted")
					lib.ShutdownApi()
				} else {
					startSyncCh <- struct{}{}
					startSyncCh = nil
				}
			}
		},
	}
	c.informers.SecretInformer.Informer().AddEventHandler(NCPSecretHandler)

	go c.informers.SecretInformer.Informer().Run(stopCh)
	if !cache.WaitForCacheSync(stopCh, c.informers.SecretInformer.Informer().HasSynced) {
		runtime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
	} else {
		utils.AviLog.Info("Caches synced for NCP Secret informer")
	}
}

func (c *VCFK8sController) AddNCPBootstrapEventHandler(k8sinfo K8sinformers, stopCh <-chan struct{}, startSyncCh chan struct{}) {
	NCPBootstrapHandler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			utils.AviLog.Infof("NCP Bootstrap ADD Event")
			if c.ValidBootStrapData() && startSyncCh != nil {
				err := c.CreateOrUpdateAviSecret()
				if err != nil {
					utils.AviLog.Warnf("Failed to create or update AVI Secret, AKO would be rebooted")
					lib.ShutdownApi()
				} else {
					startSyncCh <- struct{}{}
					startSyncCh = nil
				}
			}
		},
		UpdateFunc: func(old, obj interface{}) {
			utils.AviLog.Infof("NCP Bootstrap Update Event")
			if c.ValidBootStrapData() && startSyncCh != nil {
				err := c.CreateOrUpdateAviSecret()
				if err != nil {
					utils.AviLog.Warnf("Failed to create or update AVI Secret, AKO would be rebooted")
					lib.ShutdownApi()
				} else {
					startSyncCh <- struct{}{}
					startSyncCh = nil
				}
			}
		},
	}
	c.dynamicInformers.NCPBootstrapInformer.Informer().AddEventHandler(NCPBootstrapHandler)

	go c.dynamicInformers.NCPBootstrapInformer.Informer().Run(stopCh)
	if !cache.WaitForCacheSync(stopCh, c.dynamicInformers.NCPBootstrapInformer.Informer().HasSynced) {
		runtime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
	} else {
		utils.AviLog.Info("Caches synced for NCP Bootstrap informer")
	}
}

func (c *VCFK8sController) AddNetworkInfoEventHandler(k8sinfo K8sinformers, stopCh <-chan struct{}) {
	NetworkinfoHandler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			utils.AviLog.Infof("NCP Network Info ADD Event")
			avirest.AddSegment(obj)
		},
		UpdateFunc: func(old, obj interface{}) {
			utils.AviLog.Infof("NCP Network Info Update Event")
			avirest.AddSegment(obj)
		},
		DeleteFunc: func(obj interface{}) {
			utils.AviLog.Infof("NCP Network Info Delete Event")
			avirest.DeleteSegment(obj)
		},
	}
	c.dynamicInformers.NetworkInfoInformer.Informer().AddEventHandler(NetworkinfoHandler)

	go c.dynamicInformers.NetworkInfoInformer.Informer().Run(stopCh)
	if !cache.WaitForCacheSync(stopCh, c.dynamicInformers.NetworkInfoInformer.Informer().HasSynced) {
		runtime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
	} else {
		utils.AviLog.Info("Caches synced for networkinfo informer")
	}
}

// HandleVCF checks if avi secret used by AKO is already present. If found, the it would try to connect to
// AVI Controller. If there is any failure, we would look at Bootstrap CR used by NCP to communicate with AKO.
// If Bootstrap CR is not found, AKO would wait for it to be created. If the authtoken from Bootstrap CR
// can be used to connect to the AVI Controller, then avi-secret would be created with that token.
func (c *VCFK8sController) HandleVCF(informers K8sinformers, stopCh <-chan struct{}, ctrlCh chan struct{}, skipAviClient ...bool) string {
	cs := c.informers.ClientSet
	aviSecret, err := cs.CoreV1().Secrets(utils.GetAKONamespace()).Get(context.TODO(), lib.AviSecret, metav1.GetOptions{})
	ctrlIP := lib.GetControllerURLFromBootstrapCR()
	if err == nil && ctrlIP != "" {
		lib.SetControllerIP(ctrlIP)
		authToken := aviSecret.Data["authtoken"]
		username := aviSecret.Data["username"]
		var transport *http.Transport
		_, err = clients.NewAviClient(
			ctrlIP, string(username), session.SetAuthToken(string(authToken)),
			session.SetNoControllerStatusCheck, session.SetTransport(transport),
			session.SetInsecure,
		)
		if err == nil || len(skipAviClient) == 1 {
			utils.AviLog.Infof("Successfully connected to AVI controller using existing AKO secret")
			boostrapdata, ok := lib.GetBootstrapCRData()
			if ok {
				return boostrapdata.TZPath
			}
			utils.AviLog.Warnf("Failed to fetch transportzone from bootstrap CR status")
		} else {
			utils.AviLog.Error("AVI controller initialization failed with err: %v", err)
		}
	} else {
		utils.AviLog.Infof("Got error while fetching avi-secret: %v", err)
	}

	if !c.ValidBootStrapData() {
		utils.AviLog.Infof("Running in a VCF Cluster, but valid Bootstrap CR not found, waiting .. ")
		startSyncCh := make(chan struct{})
		c.AddNCPBootstrapEventHandler(informers, stopCh, startSyncCh)
		c.AddNCPSecretEventHandler(informers, stopCh, startSyncCh)
	L:
		for {
			select {
			case <-startSyncCh:
				break L
			case <-ctrlCh:
				return transportZone
			}
		}
	}
	utils.AviLog.Infof("NCP Bootstrap CR found, continuing AKO initialization")
	c.CreateOrUpdateAviSecret()
	return transportZone
}

func (c *VCFK8sController) CreateOrUpdateAviSecret() error {
	boostrapdata, ok := lib.GetBootstrapCRData()
	if !ok {
		utils.AviLog.Infof("Got empty data from for one or more fields from Bootstrap CR")
		return errors.New("Empty field in Bootstrap CR")
	}

	cs := c.informers.ClientSet

	var ncpSecret *corev1.Secret
	var err error
	ncpSecret, err = cs.CoreV1().Secrets(boostrapdata.SecretNamespace).Get(context.TODO(), boostrapdata.SecretName, metav1.GetOptions{})
	if err != nil {
		utils.AviLog.Warnf("Failed to get secret, got err: %v", err)
		return err
	}

	var aviSecret corev1.Secret
	aviSecret.ObjectMeta.Name = lib.AviSecret
	aviSecret.Data = make(map[string][]byte)
	aviSecret.Data["authtoken"] = []byte(ncpSecret.Data["authToken"])
	aviSecret.Data["username"] = []byte(boostrapdata.UserName)

	_, err = cs.CoreV1().Secrets(utils.GetAKONamespace()).Get(context.TODO(), lib.AviSecret, metav1.GetOptions{})
	if k8serrors.IsNotFound(err) {
		_, err = cs.CoreV1().Secrets(utils.GetAKONamespace()).Create(context.TODO(), &aviSecret, metav1.CreateOptions{})
		if err != nil {
			utils.AviLog.Warnf("Failed to create avi-secret, err: %v", err)
			return err
		}
		return nil
	}

	_, err = cs.CoreV1().Secrets(utils.GetAKONamespace()).Update(context.TODO(), &aviSecret, metav1.UpdateOptions{})
	if err != nil {
		utils.AviLog.Warnf("Failed to update avi-secret, err: %v", err)
		return err
	}

	return nil
}

func (c *VCFK8sController) ValidBootStrapData() bool {
	utils.AviLog.Infof("Validating NCP Boostrap data for AKO")
	cs := c.informers.ClientSet
	boostrapdata, ok := lib.GetBootstrapCRData()
	if !ok {
		utils.AviLog.Infof("Got empty data from for one or more fields from Bootstrap CR")
		return false
	}
	utils.AviLog.Infof("Got data from Bootstrap CR, secretName: %s, namespace: %s, username: %s, tansportzone: %s", boostrapdata.SecretName, boostrapdata.SecretNamespace, boostrapdata.UserName, boostrapdata.TZPath)
	setTranzportZone(boostrapdata.TZPath)
	var ncpSecret *corev1.Secret
	var err error
	ncpSecret, err = cs.CoreV1().Secrets(boostrapdata.SecretNamespace).Get(context.TODO(), boostrapdata.SecretName, metav1.GetOptions{})
	if err != nil {
		utils.AviLog.Warnf("Failed to get secret, got err: %v", err)
		return false
	}
	authToken := ncpSecret.Data["authToken"]
	ctrlIP := boostrapdata.AviURL
	lib.SetControllerIP(ctrlIP)
	var transport *http.Transport
	_, err = clients.NewAviClient(
		ctrlIP, boostrapdata.UserName, session.SetAuthToken(string(authToken)),
		session.SetNoControllerStatusCheck, session.SetTransport(transport),
		session.SetInsecure,
	)
	if err != nil {
		utils.AviLog.Infof("Failed to connect to AVI controller using secret provided by NCP, the secret would be deleted, err: %v", err)
		c.deleteNCPSecret(boostrapdata.SecretName, boostrapdata.SecretNamespace)
		return false
	}
	utils.AviLog.Infof("Successfully connected to AVI controller using secret provided by NCP")
	return true
}

func (c *VCFK8sController) deleteNCPSecret(name, ns string) {
	cs := c.informers.ClientSet
	err := cs.CoreV1().Secrets(ns).Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil {
		utils.AviLog.Warnf("Failed to delete NCP secret, got error: %v", err)
	}
}

func setTranzportZone(tzPath string) {
	tzonce.Do(func() {
		transportZone = tzPath
	})
}
