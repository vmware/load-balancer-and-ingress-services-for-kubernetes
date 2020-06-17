/*
 * Copyright 2019-2020 VMware, Inc.
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
	"ako/pkg/lib"
	"ako/pkg/objects"

	"github.com/avinetworks/container-lib/utils"
	routev1 "github.com/openshift/api/route/v1"
	"k8s.io/api/networking/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
)

var (
	Service = GraphSchema{
		Type:               "Service",
		GetParentIngresses: SvcToIng,
		GetParentRoutes:    SvcToRoute,
	}
	Ingress = GraphSchema{
		Type:               "Ingress",
		GetParentIngresses: IngressChanges,
	}
	Endpoint = GraphSchema{
		Type:               "Endpoints",
		GetParentIngresses: EPToIng,
		GetParentRoutes:    EPToRoute,
	}
	Secret = GraphSchema{
		Type:               "Secret",
		GetParentIngresses: SecretToIng,
		GetParentRoutes:    SecretToIng,
	}
	Route = GraphSchema{
		Type:            utils.OshiftRoute,
		GetParentRoutes: RouteChanges,
	}
	SupportedGraphTypes = GraphDescriptor{
		Ingress,
		Service,
		Endpoint,
		Secret,
		Route,
	}
)

type GraphSchema struct {
	Type               string
	GetParentIngresses func(string, string, string) ([]string, bool)
	GetParentRoutes    func(string, string, string) ([]string, bool)
}

type GraphDescriptor []GraphSchema

var SvcListerForRoutes struct {
	SvcListerForRoute map[string]objects.SvcLister
}

func RouteChanges(routeName string, namespace string, key string) ([]string, bool) {
	var routes []string
	routes = append(routes, routeName)
	routeObj, err := utils.GetInformers().RouteInformer.Lister().Routes(namespace).Get(routeName)
	if err != nil {
		// Detect a delete condition here.
		if errors.IsNotFound(err) {
			// Remove all the Ingress to Services mapping.
			// Remove the references of this ingress from the Services
			objects.OshiftRouteSvcLister().IngressMappings(namespace).RemoveIngressMappings(routeName)
		}
	} else {
		services := parseServicesForRoute(routeObj.Spec, key)
		for _, svc := range services {
			utils.AviLog.Debugf("key: %s, msg: updating route relationship for service:  %s", key, svc)
			objects.OshiftRouteSvcLister().IngressMappings(namespace).UpdateIngressMappings(routeName, svc)
		}
	}
	return routes, true
}

func SvcToRoute(svcName string, namespace string, key string) ([]string, bool) {
	_, err := utils.GetInformers().ServiceInformer.Lister().Services(namespace).Get(svcName)
	if err != nil && errors.IsNotFound(err) {
		// Garbage collect the svc if no route references exist
		_, routes := objects.OshiftRouteSvcLister().IngressMappings(namespace).GetSvcToIng(svcName)
		if len(routes) == 0 {
			objects.SharedSvcLister().IngressMappings(namespace).DeleteSvcToIngMapping(svcName)
		}
	}
	_, routes := objects.OshiftRouteSvcLister().IngressMappings(namespace).GetSvcToIng(svcName)
	utils.AviLog.Debugf("key: %s, msg: total Routes retrieved:  %v", key, routes)
	if len(routes) == 0 {
		return nil, false
	}
	return routes, true
}

func EPToRoute(epName string, namespace string, key string) ([]string, bool) {
	routes, found := SvcToRoute(epName, namespace, key)
	utils.AviLog.Debugf("key: %s, msg: total Routes retrieved:  %v", key, routes)
	return routes, found
}

func IngressChanges(ingName string, namespace string, key string) ([]string, bool) {
	var ingresses []string
	ingresses = append(ingresses, ingName)
	myIng, err := utils.GetInformers().IngressInformer.Lister().ByNamespace(namespace).Get(ingName)

	if err != nil {
		// Detect a delete condition here.
		if errors.IsNotFound(err) {
			// Remove all the Ingress to Services mapping.
			// Remove the references of this ingress from the Services
			objects.SharedSvcLister().IngressMappings(namespace).RemoveIngressMappings(ingName)
		}
	} else {
		ingObj, ok := utils.ToNetworkingIngress(myIng)
		if !ok {
			utils.AviLog.Errorf("Unable to convert obj type interface to networking/v1beta1 ingress")
		}

		// simple validator check for duplicate hostpaths, logs Warning if duplicates found
		validateSpecFromHostnameCache(key, ingObj.Namespace, ingObj.Name, ingObj.Spec)

		services := parseServicesForIngress(ingObj.Spec, key)
		for _, svc := range services {
			utils.AviLog.Debugf("key: %s, msg: updating ingress relationship for service:  %s", key, svc)
			objects.SharedSvcLister().IngressMappings(namespace).UpdateIngressMappings(ingName, svc)
		}
		secrets := parseSecretsForIngress(ingObj.Spec, key)
		if len(secrets) > 0 {
			for _, secret := range secrets {
				objects.SharedSvcLister().IngressMappings(namespace).UpdateIngressSecretsMappings(ingName, secret)
			}
		}
	}
	return ingresses, true
}

func SvcToIng(svcName string, namespace string, key string) ([]string, bool) {
	_, err := utils.GetInformers().ServiceInformer.Lister().Services(namespace).Get(svcName)
	if err != nil {
		// Detect a delete condition here.
		if errors.IsNotFound(err) {
			// Garbage collect the service if no ingress references exist
			_, ingresses := objects.SharedSvcLister().IngressMappings(namespace).GetSvcToIng(svcName)
			if len(ingresses) == 0 {
				objects.SharedSvcLister().IngressMappings(namespace).DeleteSvcToIngMapping(svcName)
			}
		}
	}
	_, ingresses := objects.SharedSvcLister().IngressMappings(namespace).GetSvcToIng(svcName)
	utils.AviLog.Debugf("key: %s, msg: total ingresses retrieved:  %s", key, ingresses)
	if len(ingresses) == 0 {
		return nil, false
	}
	return ingresses, true
}

func EPToIng(epName string, namespace string, key string) ([]string, bool) {
	ingresses, found := SvcToIng(epName, namespace, key)
	utils.AviLog.Debugf("key: %s, msg: total ingresses retrieved:  %s", key, ingresses)
	return ingresses, found
}

func SecretToIng(secretName string, namespace string, key string) ([]string, bool) {
	ok, ingNames := objects.SharedSvcLister().IngressMappings(namespace).GetSecretToIng(secretName)
	utils.AviLog.Debugf("key:%s, msg: Ingresses associated with the secret are: %s", key, ingNames)
	if ok {
		return ingNames, true
	}
	return nil, false
}

func parseServicesForRoute(routeSpec routev1.RouteSpec, key string) []string {
	// Figure out the service names that are part of this route
	var services []string

	services = append(services, routeSpec.To.Name)
	for _, ab := range routeSpec.AlternateBackends {
		services = append(services, ab.Name)
	}

	utils.AviLog.Debugf("key: %s, msg: total services retrieved  from route:  %v", key, services)
	return services
}

func parseServicesForIngress(ingSpec v1beta1.IngressSpec, key string) []string {
	// Figure out the service names that are part of this ingress
	var services []string
	for _, rule := range ingSpec.Rules {
		for _, path := range rule.IngressRuleValue.HTTP.Paths {
			services = append(services, path.Backend.ServiceName)
		}
	}
	utils.AviLog.Debugf("key: %s, msg: total services retrieved  from corev1:  %s", key, services)
	return services
}

func parseSecretsForIngress(ingSpec v1beta1.IngressSpec, key string) []string {
	// Figure out the service names that are part of this ingress
	var secrets []string
	for _, tlsSettings := range ingSpec.TLS {
		secrets = append(secrets, tlsSettings.SecretName)
	}
	utils.AviLog.Debugf("key: %s, msg: total secrets retrieved from corev1:  %s", key, secrets)
	return secrets
}

func filterIngressOnClass(ingress *v1beta1.Ingress) bool {
	// If Avi is not the default ingress, then filter on ingress class.
	if !lib.GetDefaultIngController() {
		annotations := ingress.GetAnnotations()
		ingClass, ok := annotations[lib.INGRESS_CLASS_ANNOT]
		if ok && ingClass == lib.AVI_INGRESS_CLASS {
			return true
		} else {
			return false
		}
	} else {
		// If Avi is the default ingress controller, sync everything than the ones that are annotated with ingress class other than 'avi'
		annotations := ingress.GetAnnotations()
		ingClass, ok := annotations[lib.INGRESS_CLASS_ANNOT]
		if ok && ingClass != lib.AVI_INGRESS_CLASS {
			return false
		} else {
			return true
		}
	}
}
