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

package nodes

import (
	"fmt"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/apis/ako/v1alpha1"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
)

func MultiClusterIngressChanges(ingName string, namespace string, key string) ([]string, bool) {
	var ingresses []string
	ingresses = append(ingresses, ingName)
	ingObj, err := utils.GetInformers().MultiClusterIngressInformer.Lister().MultiClusterIngresses(namespace).Get(ingName)
	if err != nil {
		utils.AviLog.Warnf("key: %s, msg: getting multi-cluster ingress with name: %s", key, ingName)
		// Detect a delete condition here.
		if k8serrors.IsNotFound(err) {
			// Garbage collect the service if no ingress references exist
			_, svcNames := objects.SharedMultiClusterIngressSvcLister().MultiClusterIngressMappings(namespace).GetIngToSvc(ingName)
			for _, svcName := range svcNames {
				objects.SharedMultiClusterIngressSvcLister().MultiClusterIngressMappings(namespace).RemoveSvcFromIngressMappings(ingName, svcName)
			}
			objects.SharedMultiClusterIngressSvcLister().MultiClusterIngressMappings(namespace).DeleteIngToSvcMapping(ingName)
			objects.SharedMultiClusterIngressSvcLister().MultiClusterIngressMappings(namespace).RemoveIngressSecretMappings(ingName)
		}
	} else {

		// TODO: Validation of host + path across MCIs

		_, oldSvcs := objects.SharedMultiClusterIngressSvcLister().MultiClusterIngressMappings(namespace).GetIngToSvc(ingName)
		currSvcs := parseServicesForMulticlusterIngress(ingObj, key)

		svcToDel := lib.Difference(oldSvcs, currSvcs)
		for _, svc := range svcToDel {
			utils.AviLog.Debugf("key: %s, msg: removing multi-cluster ingress relationship for service:  multi-cluster ingress %s, service %s", key, ingName, svc)
			// Remove the ing to svc and svc to ing mappings
			objects.SharedMultiClusterIngressSvcLister().MultiClusterIngressMappings(namespace).RemoveSvcFromIngressMappings(ingName, svc)
		}

		svcToAdd := lib.Difference(currSvcs, oldSvcs)
		for _, svc := range svcToAdd {
			utils.AviLog.Debugf("key: %s, msg: updating multi-cluster ingress relationship for service:  multi-cluster ingress %s, service %s", key, ingName, svc)
			// Update the ing to svc and svc to ing mappings
			objects.SharedMultiClusterIngressSvcLister().MultiClusterIngressMappings(namespace).UpdateIngressMappings(ingName, svc)
		}
		secret := ingObj.Spec.SecretName
		objects.SharedMultiClusterIngressSvcLister().MultiClusterIngressMappings(namespace).AddIngressToSecretsMappings(namespace, ingName, secret)
		objects.SharedMultiClusterIngressSvcLister().MultiClusterIngressMappings(namespace).AddSecretsToIngressMappings(namespace, ingName, secret)
	}
	return ingresses, true
}

func parseServicesForMulticlusterIngress(ingObj *v1alpha1.MultiClusterIngress, key string) []string {
	// Figure out the service names that are part of this multi-cluster ingress
	var services []string
	for _, config := range ingObj.Spec.Config {
		svcName := generateMultiClusterKey(config.ClusterContext, config.Service.Namespace, config.Service.Name)
		services = append(services, svcName)
	}
	utils.AviLog.Debugf("key: %s, msg: total services retrieved from multi-cluster ingress: %s", key, services)
	return services
}

func SvcToMultiClusterIng(svcName string, namespace string, key string) ([]string, bool) {
	_, err := utils.GetInformers().ServiceInformer.Lister().Services(namespace).Get(svcName)
	if err != nil {
		utils.AviLog.Warnf("key: %s, msg: getting service with name: %s", key, svcName)
		return []string{}, false
	}

	_, ingresses := objects.SharedMultiClusterIngressSvcLister().MultiClusterIngressMappings(namespace).GetSvcToIng(svcName)
	if len(ingresses) == 0 {
		return nil, false
	}

	return ingresses, true
}

func SecretToMultiClusterIng(secretName string, namespace string, key string) ([]string, bool) {
	ok, ingNames := objects.SharedMultiClusterIngressSvcLister().MultiClusterIngressMappings(namespace).GetSecretToIng(secretName)
	utils.AviLog.Debugf("key: %s, msg: Multi-cluster ingresses retrieved %s", key, ingNames)
	if !ok {
		return []string{}, false
	}
	return ingNames, true
}

func ServiceImportToMultiClusterIng(siName string, namespace string, key string) ([]string, bool) {

	serviceImport, err := utils.GetInformers().ServiceImportInformer.Lister().ServiceImports(namespace).Get(siName)
	if err != nil {
		utils.AviLog.Warnf("key: %s, msg: error getting service import with name: %s", key, siName)
		// if it is not a delete condition, then do nothing.
		if !k8serrors.IsNotFound(err) {
			return []string{}, false
		}
		found, svcNames := objects.SharedMultiClusterIngressSvcLister().MultiClusterIngressMappings(namespace).GetSIToSvc(siName)
		utils.AviLog.Debugf("key: %s, msg: services retrieved for service import with name: %s, services %s", key, siName, svcNames)
		if !found {
			return []string{}, false
		}
		svcName := svcNames[0]
		objects.SharedMultiClusterIngressSvcLister().MultiClusterIngressMappings(namespace).DeleteSvcToSIMapping(svcName)
		objects.SharedMultiClusterIngressSvcLister().MultiClusterIngressMappings(namespace).DeleteSIToSvcMapping(siName)
		found, mciNames := objects.SharedMultiClusterIngressSvcLister().MultiClusterIngressMappings(namespace).GetSvcToIng(svcName)
		utils.AviLog.Debugf("key: %s, msg: Multi-cluster ingresses retrieved for service with name: %s, multi-cluster ingresses", key, svcName, mciNames)
		return mciNames, found
	}

	svcName := generateMultiClusterKey(serviceImport.Spec.Cluster, serviceImport.Spec.Namespace, serviceImport.Spec.Service)

	// Add SI mappings
	objects.SharedMultiClusterIngressSvcLister().MultiClusterIngressMappings(namespace).UpdateSvcToSIMapping(svcName, []string{siName})
	objects.SharedMultiClusterIngressSvcLister().MultiClusterIngressMappings(namespace).UpdateSIToSvcMapping(siName, []string{svcName})

	found, mciNames := objects.SharedMultiClusterIngressSvcLister().MultiClusterIngressMappings(namespace).GetSvcToIng(svcName)
	if !found {
		utils.AviLog.Warnf("key: %s, msg: Multi-cluster ingresses not found for service with name: %s", key, svcName)
		return []string{}, false
	}

	utils.AviLog.Debugf("key: %s, msg: Multi-cluster ingresses retrieved %s", key, mciNames)
	return mciNames, true
}

func generateMultiClusterKey(cluster, namespace, objName string) string {
	return fmt.Sprintf("%s/%s/%s", cluster, namespace, objName)
}
