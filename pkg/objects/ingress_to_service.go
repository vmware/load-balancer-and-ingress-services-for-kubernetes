/*
* [2013] - [2019] Avi Networks Incorporated
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

	"github.com/avinetworks/container-lib/utils"
)

//This package gives relationship APIs to manage a kubernetes service object.

var svclisterinstance *SvcLister
var svconce sync.Once

func SharedSvcLister() *SvcLister {
	svconce.Do(func() {
		svcIngStore := NewObjectStore()
		svclisterinstance = &SvcLister{}
		svclisterinstance.svcIngStore = svcIngStore
		ingSvcStore := NewObjectStore()
		svclisterinstance.ingSvcStore = ingSvcStore
		secretIngStore := NewObjectStore()
		svclisterinstance.secretIngStore = secretIngStore
		ingSecretStore := NewObjectStore()
		svclisterinstance.ingSecretStore = ingSecretStore
		ingHostStore := NewObjectStore()
		svclisterinstance.ingHostStore = ingHostStore
	})
	return svclisterinstance
}

type SvcLister struct {
	svcIngStore    *ObjectStore
	secretIngStore *ObjectStore
	ingSvcStore    *ObjectStore
	ingSecretStore *ObjectStore
	ingHostStore   *ObjectStore
}

type SvcNSCache struct {
	namespace       string
	svcIngobjects   *ObjectMapStore
	secretIngObject *ObjectMapStore
	IngressLock     sync.RWMutex
	IngNSCache
	SecretIngNSCache
	IngHostCache
}

type IngNSCache struct {
	ingSvcobjects *ObjectMapStore
}

type SecretIngNSCache struct {
	secretIngobjects *ObjectMapStore
}

type IngHostCache struct {
	ingHostobjects *ObjectMapStore
}

func (v *SvcLister) IngressMappings(ns string) *SvcNSCache {
	namespacedsvcIngObjs := v.svcIngStore.GetNSStore(ns)
	namespacedIngSvcObjs := v.ingSvcStore.GetNSStore(ns)
	namespacedSecretIngObjs := v.secretIngStore.GetNSStore(ns)
	namespacedIngSecretObjs := v.ingSecretStore.GetNSStore(ns)
	namespacedIngHostObjs := v.ingHostStore.GetNSStore(ns)
	return &SvcNSCache{namespace: ns, svcIngobjects: namespacedsvcIngObjs,
		secretIngObject: namespacedSecretIngObjs, IngNSCache: IngNSCache{ingSvcobjects: namespacedIngSvcObjs},
		SecretIngNSCache: SecretIngNSCache{secretIngobjects: namespacedIngSecretObjs}, IngHostCache: IngHostCache{ingHostobjects: namespacedIngHostObjs}}
}

//=====All service to ingress mapping methods are here.

func (v *SvcNSCache) GetSvcToIng(svcName string) (bool, []string) {
	// Need checks if it's found or not?
	found, ingNames := v.svcIngobjects.Get(svcName)
	if !found {
		return false, make([]string, 0)
	}
	return true, ingNames.([]string)
}

func (v *SvcNSCache) DeleteSvcToIngMapping(svcName string) bool {
	// Need checks if it's found or not?
	success := v.svcIngobjects.Delete(svcName)
	utils.AviLog.Info.Printf("Deleted the service mappings for svc: %s", svcName)
	return success
}

func (v *SvcNSCache) UpdateSvcToIngMapping(svcName string, ingressList []string) {
	utils.AviLog.Info.Printf("Updated the service mappings with svc: %s, ingresses: %s", svcName, ingressList)
	v.svcIngobjects.AddOrUpdate(svcName, ingressList)
}

//=====All secret to ingress mapping methods are here.

func (v *SvcNSCache) GetSecretToIng(secretName string) (bool, []string) {
	// Need checks if it's found or not?
	found, ingNames := v.secretIngObject.Get(secretName)
	if !found {
		return false, make([]string, 0)
	}
	return true, ingNames.([]string)
}

func (v *SvcNSCache) DeleteSecretToIngMapping(secretName string) bool {
	// Need checks if it's found or not?
	success := v.secretIngObject.Delete(secretName)
	utils.AviLog.Info.Printf("Deleted the ingress mappings for secret: %s", secretName)
	return success
}

func (v *SvcNSCache) UpdateSecretToIngMapping(secretName string, ingressList []string) {
	utils.AviLog.Info.Printf("Updated the secret mappings with secret: %s, ingresses: %s", secretName, ingressList)
	v.secretIngObject.AddOrUpdate(secretName, ingressList)
}

//=====All ingress to service mapping methods are here.

func (v *IngNSCache) GetIngToSvc(ingName string) (bool, []string) {
	// Need checks if it's found or not?
	found, svcNames := v.ingSvcobjects.Get(ingName)
	if !found {
		return false, make([]string, 0)
	}
	return true, svcNames.([]string)
}

func (v *IngNSCache) DeleteIngToSvcMapping(ingName string) bool {
	// Need checks if it's found or not?
	success := v.ingSvcobjects.Delete(ingName)
	return success
}

func (v *IngNSCache) UpdateIngToSvcMapping(ingName string, svcList []string) {
	utils.AviLog.Info.Printf("Updated the ingress mappings with ingress: %s, svcs: %s", ingName, svcList)
	v.ingSvcobjects.AddOrUpdate(ingName, svcList)
}

func (v *IngNSCache) UpdatedIngressMappings(ingName string, svcList []string) {
	v.UpdateIngToSvcMapping(ingName, svcList)
}

//=====All ingress to secret mapping methods are here.

func (v *SecretIngNSCache) GetIngToSecret(ingName string) (bool, []string) {
	// Need checks if it's found or not?
	found, secretNames := v.secretIngobjects.Get(ingName)
	if !found {
		return false, make([]string, 0)
	}
	return true, secretNames.([]string)
}

func (v *SecretIngNSCache) DeleteIngToSecretMapping(ingName string) bool {
	// Need checks if it's found or not?
	success := v.secretIngobjects.Delete(ingName)
	return success
}

func (v *SecretIngNSCache) UpdateIngToSecretMapping(ingName string, secretList []string) {
	utils.AviLog.Info.Printf("Updated the ingress mappings with ingress: %s, secrets: %s", ingName, secretList)
	v.secretIngobjects.AddOrUpdate(ingName, secretList)
}

//=====All ingress to host mapping methods are here.

func (v *IngHostCache) GetIngToHost(ingName string) (bool, []string) {
	// Need checks if it's found or not?
	found, hosts := v.ingHostobjects.Get(ingName)
	if !found {
		return false, make([]string, 0)
	}
	return true, hosts.([]string)
}

func (v *IngHostCache) DeleteIngToHostMapping(ingName string) bool {
	// Need checks if it's found or not?
	success := v.ingHostobjects.Delete(ingName)
	return success
}

func (v *IngHostCache) UpdateIngToHostMapping(ingName string, hostList []string) {
	utils.AviLog.Info.Printf("Updated the ingress mappings with ingress: %s, hosts: %s", ingName, hostList)
	v.ingHostobjects.AddOrUpdate(ingName, hostList)
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

func (v *SvcNSCache) UpdateIngressSecretsMappings(ingName string, secret string) {
	v.IngressLock.Lock()
	defer v.IngressLock.Unlock()
	_, ingresses := v.GetSecretToIng(secret)
	if !utils.HasElem(ingresses, ingName) {
		ingresses = append(ingresses, ingName)
		v.UpdateSecretToIngMapping(secret, ingresses)
	}
	_, secrets := v.GetIngToSecret(ingName)
	if !utils.HasElem(secrets, secret) {
		secrets = append(secrets, secret)
		utils.AviLog.Info.Printf("Updated the ingress-->secret mapping: %s", secrets)
		v.UpdateIngToSecretMapping(ingName, secrets)
	}
}

func (v *SvcNSCache) RemoveIngressMappings(ingName string) {
	v.IngressLock.Lock()
	defer v.IngressLock.Unlock()
	// Get all the services for the ingress
	ok, svcs := v.GetIngToSvc(ingName)
	// Iterate and remove this ingress from the service mappings
	if ok {
		for _, svc := range svcs {
			found, ingresses := v.GetSvcToIng(svc)
			if found {
				ingresses = Remove(ingresses, ingName)
				// Update the service mapping
				v.UpdateSvcToIngMapping(svc, ingresses)
			}
		}
	}
	// Remove the ingress from the ingress --> service map
	v.DeleteIngToSvcMapping(ingName)
}

func (v *SvcNSCache) RemoveIngressSecretMappings(ingName string) {
	v.IngressLock.Lock()
	defer v.IngressLock.Unlock()
	// Get all the secrets for the ingress
	ok, secrets := v.GetIngToSecret(ingName)
	// Iterate and remove this ingress from the secret mappings
	if ok {
		for _, secret := range secrets {
			found, ingresses := v.GetSecretToIng(secret)
			if found {
				ingresses = Remove(ingresses, ingName)
				// Update the secret mapping
				v.UpdateSecretToIngMapping(secret, ingresses)
			}
		}
	}
	// Remove the ingress from the ingress --> secret map
	v.DeleteIngToSecretMapping(ingName)
}

// Candidate for utils package
func Remove(s []string, r string) []string {
	for i, v := range s {
		if v == r {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
}
