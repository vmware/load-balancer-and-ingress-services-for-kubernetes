/*
 * Copyright 2020-2021 VMware, Inc.
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

package k8s

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/third_party/github.com/vmware/alb-sdk/go/clients"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/third_party/github.com/vmware/alb-sdk/go/session"

	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
)

func (c *AviController) AddNCPSecretEventHandler(k8sinfo K8sinformers, stopCh <-chan struct{}, startSyncCh chan struct{}) {
	cs := k8sinfo.Cs
	utils.AviLog.Infof("Creating event broadcaster for NCP Secret Handling")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(utils.AviLog.Debugf)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: cs.CoreV1().Events("")})
	NCPSecretHandler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if lib.VCFInitialized {
				return
			}
			if c.ValidBSData() && startSyncCh != nil {
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
			if c.ValidBSData() && startSyncCh != nil {
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

func (c *AviController) AddNCPBootstrapEventHandler(k8sinfo K8sinformers, stopCh <-chan struct{}, startSyncCh chan struct{}) {
	cs := k8sinfo.Cs
	utils.AviLog.Debugf("Creating event broadcaster for NCP Bootstrap CRD")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(utils.AviLog.Debugf)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: cs.CoreV1().Events("")})
	NCPBootstrapHandler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			utils.AviLog.Infof("NCP Bootstrap ADD Event")
			if c.ValidBSData() && startSyncCh != nil {
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
			if c.ValidBSData() && startSyncCh != nil {
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

// HandleVCF checks if avi secret used by AKO is already present. If found, the it would try to connect to
// AVI Controller. If there is any failure, we would look at Bootstrap CR used by NCP to communicate with AKO.
// If Bootstrap CR is not found, AKO would wait for it to be created. If the authtoken from Bootstrap CR
// can be used to connect to the AVI Controller, then avi-secret would be created with that token.
func (c *AviController) HandleVCF(informers K8sinformers, stopCh <-chan struct{}, ctrlCh chan struct{}) {
	cs := c.informers.ClientSet
	aviSecret, err := cs.CoreV1().Secrets(utils.GetAKONamespace()).Get(context.TODO(), lib.AviSecret, metav1.GetOptions{})
	if err == nil {
		ctrlIP := os.Getenv(utils.ENV_CTRL_IPADDRESS)
		authToken := aviSecret.Data["authtoken"]
		username := aviSecret.Data["username"]
		var transport *http.Transport
		_, err = clients.NewAviClient(
			ctrlIP, string(username), session.SetAuthToken(string(authToken)),
			session.SetNoControllerStatusCheck, session.SetTransport(transport),
			session.SetInsecure,
		)
		if err == nil {
			utils.AviLog.Infof("Successfully connected to AVI controller using existing AKO secret")
			return
		} else {
			utils.AviLog.Error("AVI controller initialization failed with err: %v", err)
		}
	} else {
		utils.AviLog.Infof("Got error while fetching avi-secret: %v", err)
	}

	if !c.ValidBSData() {
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
				return
			}
		}
	}
	utils.AviLog.Infof("NCP Bootstrap CR found, continuing AKO initialization")
	c.CreateOrUpdateAviSecret()
}

func (c *AviController) CreateOrUpdateAviSecret() error {
	secretName, ns, username := lib.GetBootstrapCRData()
	if secretName == "" || ns == "" || username == "" {
		utils.AviLog.Infof("Got empty data from for one or more fields from Bootstrap CR, secretName: %s, namespace: %s, username: %s",
			secretName, ns, username)
		return errors.New("Empty field in Bootstrap CR")
	}

	cs := c.informers.ClientSet

	var ncpSecret *corev1.Secret
	var err error
	ncpSecret, err = cs.CoreV1().Secrets(ns).Get(context.TODO(), secretName, metav1.GetOptions{})
	if err != nil {
		utils.AviLog.Warnf("Failed to get secret, got err: %v", err)
		return err
	}

	var aviSecret corev1.Secret
	aviSecret.ObjectMeta.Name = lib.AviSecret
	aviSecret.Data = make(map[string][]byte)
	aviSecret.Data["authtoken"] = []byte(ncpSecret.Data["authToken"])
	aviSecret.Data["username"] = []byte(username)

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

func (c *AviController) ValidBSData() bool {
	utils.AviLog.Infof("Validating NCP Boostrap data for AKO")
	cs := c.informers.ClientSet
	secretName, ns, username := lib.GetBootstrapCRData()
	if secretName == "" || ns == "" || username == "" {
		utils.AviLog.Infof("Got empty data from for one or more fields from Bootstrap CR, secretName: %s, namespace: %s, username: %s", secretName, ns, username)
		return false
	}
	utils.AviLog.Infof("Got data from Bootstrap CR, secretName: %s, namespace: %s, username: %s", secretName, ns, username)
	var ncpSecret *corev1.Secret
	var err error
	ncpSecret, err = cs.CoreV1().Secrets(ns).Get(context.TODO(), secretName, metav1.GetOptions{})
	if err != nil {
		utils.AviLog.Warnf("Failed to get secret, got err: %v", err)
		return false
	}
	authToken := ncpSecret.Data["authToken"]
	ctrlIP := os.Getenv(utils.ENV_CTRL_IPADDRESS)
	var transport *http.Transport
	_, err = clients.NewAviClient(
		ctrlIP, username, session.SetAuthToken(string(authToken)),
		session.SetNoControllerStatusCheck, session.SetTransport(transport),
		session.SetInsecure,
	)
	if err != nil {
		utils.AviLog.Infof("Failed to connect to AVI controller using secret provided by NCP, the secret would be deleted, err: %v", err)
		c.deleteNCPSecret(secretName, ns)
		return false
	}
	utils.AviLog.Infof("Successfully connected to AVI controller using secret provided by NCP")
	return true
}

func (c *AviController) deleteNCPSecret(name, ns string) {
	cs := c.informers.ClientSet
	err := cs.CoreV1().Secrets(ns).Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil {
		utils.AviLog.Warnf("Failed to delete NCP secret, got error: %v", err)
	}
}
