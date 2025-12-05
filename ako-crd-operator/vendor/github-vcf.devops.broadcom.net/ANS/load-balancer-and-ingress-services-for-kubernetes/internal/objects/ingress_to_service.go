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

package objects

import (
	"sync"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
)

//This package gives relationship APIs to manage a kubernetes service object.

var svclisterinstance *SvcLister
var svconce sync.Once

var oshiftroutesvclister *SvcLister
var routesvconce sync.Once

func SharedSvcLister() *SvcLister {
	svconce.Do(func() {
		svclisterinstance = &SvcLister{
			svcIngStore:         NewObjectStore(),
			ingSvcStore:         NewObjectStore(),
			secretIngStore:      NewObjectStore(),
			ingSecretStore:      NewObjectStore(),
			secretHostNameStore: NewObjectStore(),
			ingHostStore:        NewObjectStore(),
			classIngStore:       NewObjectStore(),
			ingClassStore:       NewObjectStore(),
		}
	})
	return svclisterinstance
}

func OshiftRouteSvcLister() *SvcLister {
	routesvconce.Do(func() {
		oshiftroutesvclister = &SvcLister{
			svcIngStore:         NewObjectStore(),
			ingSvcStore:         NewObjectStore(),
			secretIngStore:      NewObjectStore(),
			ingSecretStore:      NewObjectStore(),
			secretHostNameStore: NewObjectStore(),
			ingHostStore:        NewObjectStore(),
			classIngStore:       NewObjectStore(),
			ingClassStore:       NewObjectStore(),
		}
	})
	return oshiftroutesvclister
}

type SvcLister struct {
	svcIngStore         *ObjectStore
	secretIngStore      *ObjectStore
	ingSvcStore         *ObjectStore
	ingSecretStore      *ObjectStore
	ingHostStore        *ObjectStore
	secretHostNameStore *ObjectStore
	ingClassStore       *ObjectStore
	classIngStore       *ObjectStore
	svcSIStore          *ObjectStore
	SISvcStore          *ObjectStore
}

type SvcNSCache struct {
	namespace       string
	svcIngObject    *ObjectMapStore
	secretIngObject *ObjectMapStore
	classIngObject  *ObjectMapStore
	IngressLock     sync.RWMutex
	IngNSCache
	SecretIngNSCache
	IngHostCache
	SecretHostNameNSCache
	IngClassNSCache
	mciNSCache
}

// stores path/policies for a hostname
// policies can be of various types for both secure and insecure hosts
type RouteIngrhost struct {
	Hostname       string
	InsecurePolicy string // none, redirect, allow
	SecurePolicy   string // edge, reencrypt, passthrough
	Paths          []string
	PathSvc        map[string][]string //list of services for a path, used for alternate backend
}

type IngNSCache struct {
	ingSvcObjects *ObjectMapStore
}

type SecretIngNSCache struct {
	ingSecretObjects *ObjectMapStore
}

type SecretHostNameNSCache struct {
	SecretLock            sync.RWMutex
	secretHostNameObjects *ObjectMapStore
}

type IngClassNSCache struct {
	ingClassObjects *ObjectMapStore
}

type IngHostCache struct {
	ingHostObjects *ObjectMapStore
}

func (v *SvcLister) IngressMappings(ns string) *SvcNSCache {
	return &SvcNSCache{
		namespace:       ns,
		svcIngObject:    v.svcIngStore.GetNSStore(ns),
		secretIngObject: v.secretIngStore.GetNSStore(ns),
		classIngObject:  v.classIngStore.GetNSStore(ns),
		IngNSCache: IngNSCache{
			ingSvcObjects: v.ingSvcStore.GetNSStore(ns),
		},
		SecretHostNameNSCache: SecretHostNameNSCache{
			secretHostNameObjects: v.secretHostNameStore.GetNSStore(ns),
		},
		SecretIngNSCache: SecretIngNSCache{
			ingSecretObjects: v.ingSecretStore.GetNSStore(ns),
		},
		IngHostCache: IngHostCache{
			ingHostObjects: v.ingHostStore.GetNSStore(ns),
		},
		IngClassNSCache: IngClassNSCache{
			ingClassObjects: v.ingClassStore.GetNSStore(ns),
		},
	}
}

//=====All service to ingress mapping methods are here.

func (v *SvcNSCache) GetSvcToIng(svcName string) (bool, []string) {
	found, ingNames := v.svcIngObject.Get(svcName)
	if !found {
		return false, make([]string, 0)
	}
	return true, ingNames.([]string)
}

func (v *SvcNSCache) DeleteSvcToIngMapping(svcName string) bool {
	success := v.svcIngObject.Delete(svcName)
	utils.AviLog.Debugf("Deleted the service mappings for svc: %s", svcName)
	return success
}

func (v *SvcNSCache) UpdateSvcToIngMapping(svcName string, ingressList []string) {
	utils.AviLog.Debugf("Updated the service mappings with svc: %s, ingresses: %s", svcName, ingressList)
	v.svcIngObject.AddOrUpdate(svcName, ingressList)
}

//=====All secret to ingress mapping methods are here.

func (v *SvcNSCache) GetSecretToIng(secretName string) (bool, []string) {
	found, ingNames := v.secretIngObject.Get(secretName)
	if !found {
		return false, make([]string, 0)
	}
	return true, ingNames.([]string)
}

func (v *SvcNSCache) DeleteSecretToIngMapping(secretName string) bool {
	success := v.secretIngObject.Delete(secretName)
	utils.AviLog.Debugf("Deleted the ingress mappings for secret: %s", secretName)
	return success
}

func (v *SvcNSCache) UpdateSecretToIngMapping(secretName string, ingressList []string) {
	utils.AviLog.Debugf("Updated the secret mappings with secret: %s, ingresses: %s", secretName, ingressList)
	v.secretIngObject.AddOrUpdate(secretName, ingressList)
}

//=====All ingress class to ingress mapping methods are here.

func (v *SvcNSCache) GetClassToIng(className string) (bool, []string) {
	found, ingNSNames := v.classIngObject.Get(className)
	if !found {
		return false, make([]string, 0)
	}
	return true, ingNSNames.([]string)
}

func (v *SvcNSCache) DeleteClassToIngMapping(className string) bool {
	success := v.classIngObject.Delete(className)
	utils.AviLog.Debugf("Deleted the ingress mappings for class: %s", className)
	return success
}

func (v *SvcNSCache) UpdateClassToIngMapping(className string, ingressList []string) {
	utils.AviLog.Debugf("Updated the class mappings with class: %s, ingresses: %s", className, ingressList)
	v.classIngObject.AddOrUpdate(className, ingressList)
}

//=====All ingress to service mapping methods are here.

func (v *IngNSCache) GetIngToSvc(ingName string) (bool, []string) {
	found, svcNames := v.ingSvcObjects.Get(ingName)
	if !found {
		return false, make([]string, 0)
	}
	return true, svcNames.([]string)
}

func (v *IngNSCache) DeleteIngToSvcMapping(ingName string) bool {
	success := v.ingSvcObjects.Delete(ingName)
	return success
}

func (v *IngNSCache) UpdateIngToSvcMapping(ingName string, svcList []string) {
	utils.AviLog.Debugf("Updated the ingress mappings with ingress: %s, svcs: %s", ingName, svcList)
	v.ingSvcObjects.AddOrUpdate(ingName, svcList)
}

//=====All ingress to secret mapping methods are here

func (v *SecretIngNSCache) GetIngToSecret(ingName string) (bool, []string) {
	found, secretNames := v.ingSecretObjects.Get(ingName)
	if !found {
		return false, make([]string, 0)
	}
	return true, secretNames.([]string)
}

func (v *SecretIngNSCache) DeleteIngToSecretMapping(ingName string) bool {
	success := v.ingSecretObjects.Delete(ingName)
	return success
}

func (v *SecretIngNSCache) UpdateIngToSecretMapping(ingName string, secretList []string) {
	utils.AviLog.Debugf("Updated the ingress mappings with ingress: %s, secrets: %s", ingName, secretList)
	v.ingSecretObjects.AddOrUpdate(ingName, secretList)
}

//=====All ingress to ingress class mapping methods are here.

func (v *IngClassNSCache) GetIngToClass(ingName string) (bool, string) {
	found, class := v.ingClassObjects.Get(ingName)
	if !found {
		return false, ""
	}
	return true, class.(string)
}

func (v *IngClassNSCache) DeleteIngToClassMapping(ingName string) bool {
	success := v.ingClassObjects.Delete(ingName)
	return success
}

func (v *IngClassNSCache) UpdateIngToClassMapping(ingName string, class string) {
	utils.AviLog.Debugf("Updated the ingress mappings with ingress: %s, class: %s", ingName, class)
	v.ingClassObjects.AddOrUpdate(ingName, class)
}

//=====All secret to hostname mapping goes here.

func (v *SecretHostNameNSCache) GetSecretToHostname(secretName string) (bool, []string) {
	found, hostNames := v.secretHostNameObjects.Get(secretName)
	if !found {
		return false, make([]string, 0)
	}
	return true, hostNames.([]string)
}

func (v *SecretHostNameNSCache) DeleteSecretToHostNameMapping(secretName string) bool {
	success := v.secretHostNameObjects.Delete(secretName)
	return success
}

func (v *SecretHostNameNSCache) UpdateSecretToHostNameMapping(secretName string, hostName string) {
	v.SecretLock.Lock()
	defer v.SecretLock.Unlock()
	var hostnames []string
	found := false
	// Get the list of hostnames for this secret and update the new one.
	found, hostnames = v.GetSecretToHostname(secretName)
	if found {
		if !utils.HasElem(hostnames, hostName) {
			hostnames = append(hostnames, hostName)
		}
	} else {
		hostnames = []string{hostName}
	}
	utils.AviLog.Debugf("Updated the secret mappings for secret: %s, hostnames: %s", secretName, hostnames)
	v.secretHostNameObjects.AddOrUpdate(secretName, hostnames)
}

func (v *SecretHostNameNSCache) DecrementSecretToHostNameMapping(secretName string, hostName string) []string {
	v.SecretLock.Lock()
	defer v.SecretLock.Unlock()
	var hostnames []string
	found := false
	// Get the list of hostnames for this secret and update the new one.
	found, hostnames = v.GetSecretToHostname(secretName)
	if found {
		if utils.HasElem(hostnames, hostName) {
			hostnames = utils.Remove(hostnames, hostName)
		}
	}
	utils.AviLog.Debugf("After Decrement secret: %s, hostnames: %s", secretName, hostnames)

	v.secretHostNameObjects.AddOrUpdate(secretName, hostnames)
	return hostnames
}

func (v *IngHostCache) GetIngToHost(ingName string) (bool, map[string]map[string][]string) {
	found, hosts := v.ingHostObjects.Get(ingName)
	if !found {
		return false, make(map[string]map[string][]string, 0)
	}
	return true, hosts.(map[string]map[string][]string)
}

func (v *IngHostCache) GetRouteIngToHost(ingName string) (bool, map[string]*RouteIngrhost) {
	found, hosts := v.ingHostObjects.Get(ingName)
	if !found {
		return false, make(map[string]*RouteIngrhost)
	}
	return true, hosts.(map[string]*RouteIngrhost)
}

func (v *IngHostCache) UpdateRouteIngToHostMapping(ingName string, hostMap map[string]*RouteIngrhost) {
	utils.AviLog.Debugf("Updated the ingress mappings with ingress: %s, hosts: %v", ingName, utils.Stringify(hostMap))
	v.ingHostObjects.AddOrUpdate(ingName, hostMap)
}

func (v *IngHostCache) DeleteIngToHostMapping(ingName string) bool {
	success := v.ingHostObjects.Delete(ingName)
	return success
}

func (v *IngHostCache) UpdateIngToHostMapping(ingName string, hostMap map[string]map[string][]string) {
	utils.AviLog.Debugf("Updated the ingress mappings with ingress: %s, hosts: %s", ingName, hostMap)
	v.ingHostObjects.AddOrUpdate(ingName, hostMap)
}

//===All cross mapping update methods are here.

func (v *SvcNSCache) UpdateIngressMappings(ingName string, svcName string) {
	v.IngressLock.Lock()
	defer v.IngressLock.Unlock()
	_, ingresses := v.GetSvcToIng(svcName)
	if !utils.HasElem(ingresses, ingName) {
		ingresses = append(ingresses, ingName)
		v.UpdateSvcToIngMapping(svcName, ingresses)
	}
	_, svcs := v.GetIngToSvc(ingName)
	if !utils.HasElem(svcs, svcName) {
		svcs = append(svcs, svcName)
		v.UpdateIngToSvcMapping(ingName, svcs)
	}
}

func (v *SvcNSCache) RemoveSvcFromIngressMappings(ingName string, svcName string) {
	v.IngressLock.Lock()
	defer v.IngressLock.Unlock()
	_, ingresses := v.GetSvcToIng(svcName)
	if utils.HasElem(ingresses, ingName) {
		ingresses = utils.Remove(ingresses, ingName)
		v.UpdateSvcToIngMapping(svcName, ingresses)
	}
	_, svcs := v.GetIngToSvc(ingName)
	if utils.HasElem(svcs, svcName) {
		svcs = utils.Remove(svcs, svcName)
		v.UpdateIngToSvcMapping(ingName, svcs)
	}
}

func (v *SvcNSCache) AddSecretsToIngressMappings(ingressNS, ingName, secretName string) {
	v.IngressLock.Lock()
	defer v.IngressLock.Unlock()
	nsIngress := ingressNS + "/" + ingName
	_, ingresses := v.GetSecretToIng(secretName)
	if !utils.HasElem(ingresses, nsIngress) {
		ingresses = append(ingresses, nsIngress)
		v.UpdateSecretToIngMapping(secretName, ingresses)
	}
}

func (v *SvcNSCache) AddIngressToSecretsMappings(secretNS, ingName, secretName string) {
	v.IngressLock.Lock()
	defer v.IngressLock.Unlock()
	_, secrets := v.GetIngToSecret(ingName)
	nsSecret := secretNS + "/" + secretName
	if !utils.HasElem(secrets, nsSecret) {
		secrets = append(secrets, nsSecret)
		utils.AviLog.Debugf("Updated the ingress: %s to secrets: %s", ingName, secrets)
		v.UpdateIngToSecretMapping(ingName, secrets)
	}
}

func (v *SvcNSCache) UpdateIngressClassMappings(ingName string, ingClass string) {
	v.IngressLock.Lock()
	defer v.IngressLock.Unlock()
	_, ingresses := v.GetClassToIng(ingClass)
	if !utils.HasElem(ingresses, ingName) {
		ingresses = append(ingresses, ingName)
		v.UpdateClassToIngMapping(ingClass, ingresses)
	}
	v.UpdateIngToClassMapping(ingName, ingClass)
}

func (v *SvcNSCache) RemoveIngressMappings(ingName string) []string {
	v.IngressLock.Lock()
	defer v.IngressLock.Unlock()
	// Get all the services for the ingress
	ok, svcs := v.GetIngToSvc(ingName)
	// Iterate and remove this ingress from the service mappings
	var svcToDel []string
	if ok {
		for _, svc := range svcs {
			found, ingresses := v.GetSvcToIng(svc)
			if found {
				ingresses = utils.Remove(ingresses, ingName)
				// Update the service mapping
				v.UpdateSvcToIngMapping(svc, ingresses)
				if len(ingresses) == 0 {
					svcToDel = append(svcToDel, svc)
				}
			}
		}
	}
	// Remove the ingress from the ingress --> service map
	v.DeleteIngToSvcMapping(ingName)
	return svcToDel
}

func (v *SvcNSCache) RemoveIngressSecretMappings(ingNSName string) {
	v.IngressLock.Lock()
	defer v.IngressLock.Unlock()
	// Get all the secrets for the ingress
	ok, secrets := v.GetIngToSecret(ingNSName)
	// Iterate and remove this ingress from the secret mappings
	if ok {
		for _, secret := range secrets {
			found, ingresses := v.GetSecretToIng(secret)
			if found {
				ingresses = utils.Remove(ingresses, ingNSName)
				// Update the secret mapping
				v.UpdateSecretToIngMapping(secret, ingresses)
			}
		}
	}
	// Remove the ingress from the ingress --> secret map
	v.DeleteIngToSecretMapping(ingNSName)
}

func (v *SvcNSCache) RemoveIngressClassMappings(ingNSName string) {
	v.IngressLock.Lock()
	defer v.IngressLock.Unlock()
	ok, class := v.GetIngToClass(ingNSName)
	if ok {
		found, ingresses := v.GetClassToIng(class)
		if found {
			ingresses = utils.Remove(ingresses, ingNSName)
			v.UpdateClassToIngMapping(class, ingresses)
		}
	}
	v.DeleteIngToClassMapping(ingNSName)
}
