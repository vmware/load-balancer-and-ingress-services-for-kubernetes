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
	"fmt"
	"regexp"
	"strings"

	"github.com/avinetworks/ako/internal/lib"
	"github.com/avinetworks/ako/internal/objects"

	"github.com/avinetworks/ako/pkg/utils"

	routev1 "github.com/openshift/api/route/v1"
	advl4v1alpha1pre1 "github.com/vmware-tanzu/service-apis/apis/v1alpha1pre1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/api/networking/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
)

var (
	Service = GraphSchema{
		Type:               "Service",
		GetParentIngresses: SvcToIng,
		GetParentRoutes:    SvcToRoute,
		GetParentGateways:  SvcToGateway,
	}
	Ingress = GraphSchema{
		Type:               "Ingress",
		GetParentIngresses: IngressChanges,
	}
	Endpoint = GraphSchema{
		Type:               "Endpoints",
		GetParentIngresses: EPToIng,
		GetParentRoutes:    EPToRoute,
		GetParentGateways:  EPToGateway,
	}
	Node = GraphSchema{
		Type:               "Node",
		GetParentIngresses: NodeToIng,
		GetParentRoutes:    NodeToRoute,
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
	HostRule = GraphSchema{
		Type:               "HostRule",
		GetParentIngresses: HostRuleToIng,
		GetParentRoutes:    HostRuleToIng,
	}
	HTTPRule = GraphSchema{
		Type:               "HTTPRule",
		GetParentIngresses: HTTPRuleToIng,
		GetParentRoutes:    HTTPRuleToIng,
	}
	Gateway = GraphSchema{
		Type:              "Gateway",
		GetParentGateways: GatewayChanges,
	}
	GatewayClass = GraphSchema{
		Type:              "GatewayClass",
		GetParentGateways: GatewayClassChanges,
	}
	SupportedGraphTypes = GraphDescriptor{
		Ingress,
		Service,
		Endpoint,
		Secret,
		Route,
		Node,
		HostRule,
		HTTPRule,
		Gateway,
		GatewayClass,
	}
)

type GraphSchema struct {
	Type               string
	GetParentIngresses func(string, string, string) ([]string, bool)
	GetParentRoutes    func(string, string, string) ([]string, bool)
	GetParentGateways  func(string, string, string) ([]string, bool)
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
		validateRouteSpecFromHostnameCache(key, namespace, routeName, routeObj.Spec)
		services := parseServicesForRoute(routeObj.Spec, key)
		for _, svc := range services {
			utils.AviLog.Debugf("key: %s, msg: updating route relationship for service:  %s", key, svc)
			objects.OshiftRouteSvcLister().IngressMappings(namespace).UpdateIngressMappings(routeName, svc)
		}
		if routeObj.Spec.TLS != nil {
			secret := lib.RouteSecretsPrefix + routeName
			objects.OshiftRouteSvcLister().IngressMappings(namespace).UpdateIngressSecretsMappings(routeName, secret)
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

func SvcToGateway(svcName string, namespace string, key string) ([]string, bool) {
	var gateway string
	var portProtocols []string

	myService, err := utils.GetInformers().ServiceInformer.Lister().Services(namespace).Get(svcName)
	if err != nil && errors.IsNotFound(err) {
		// Garbage collect the svc if no route references exist
		_, gateway = objects.ServiceGWLister().GetSvcToGw(namespace + "/" + svcName)
		if gateway == "" {
			return nil, false
		}
		objects.ServiceGWLister().RemoveGatewayMappings(gateway, namespace+"/"+svcName)
	} else {
		gateway, portProtocols = parseL4ServiceForGateway(myService, key)
		if gateway != "" {
			objects.ServiceGWLister().UpdateGatewayMappings(gateway, namespace+"/"+svcName, portProtocols[0])
		} else {
			return nil, false
		}
	}

	utils.AviLog.Debugf("key: %s, msg: Gateway retrieved %s", key, gateway)
	return []string{gateway}, true
}

func EPToRoute(epName string, namespace string, key string) ([]string, bool) {
	routes, found := SvcToRoute(epName, namespace, key)
	utils.AviLog.Debugf("key: %s, msg: total Routes retrieved:  %v", key, routes)
	return routes, found
}

func EPToGateway(epName string, namespace string, key string) ([]string, bool) {
	gateways, found := SvcToGateway(epName, namespace, key)
	utils.AviLog.Debugf("key: %s, msg: total Gateways retrieved:  %v", key, gateways)
	return gateways, found
}

func GatewayChanges(gwName string, namespace string, key string) ([]string, bool) {
	var allGateways []string
	allGateways = append(allGateways, namespace+"/"+gwName)
	gateway, err := lib.GetAdvL4Informers().GatewayInformer.Lister().Gateways(namespace).Get(gwName)
	if err != nil && errors.IsNotFound(err) {
		// Remove all the Gateway to Services mapping.
		objects.ServiceGWLister().DeleteGWListeners(namespace + "/" + gwName)
		objects.ServiceGWLister().RemoveGatewayGWclassMappings(namespace + "/" + gwName)
	} else {
		if err = validateGatewayObj(key, gateway); err != nil {
			return allGateways, false
		}

		if gwListeners := parseGatewayForListeners(gateway, key); len(gwListeners) > 0 {
			objects.ServiceGWLister().UpdateGWListeners(namespace+"/"+gwName, gwListeners)
		}
		gatewayClass := gateway.Spec.Class
		objects.ServiceGWLister().UpdateGatewayGWclassMappings(namespace+"/"+gwName, gatewayClass)
	}
	return allGateways, true
}

func GatewayClassChanges(gwClassName string, namespace string, key string) ([]string, bool) {
	found, gateways := objects.ServiceGWLister().GetGWclassToGateways(gwClassName)
	return gateways, found
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

func NodeToIng(nodeName string, namespace string, key string) ([]string, bool) {
	// Get all Ingress in the system, as node create/update affects all ingresses in the system
	if !lib.IsNodePortMode() || utils.GetInformers().IngressInformer == nil {
		return nil, false
	}
	ingresses := []string{}
	ingressObjs, err := utils.GetInformers().IngressInformer.Lister().ByNamespace("").List(labels.Set(nil).AsSelector())
	if err != nil {
		utils.AviLog.Errorf("Unable to retrieve the ingress: %s", err)
		return nil, false
	}

	for _, ingObj := range ingressObjs {
		ingObjValidated, ok := utils.ToNetworkingIngress(ingObj)
		if !ok {
			utils.AviLog.Errorf("Unable to convert obj type interface to networking/v1beta1 ingress")
		}
		ingresses = append(ingresses, ingObjValidated.Name)
	}
	if len(ingresses) == 0 {
		return nil, false
	}
	utils.AviLog.Debugf("key: %s, msg: total ingresses retrieved:  %s", key, ingresses)
	return ingresses, true
}

func NodeToRoute(nodeName string, namespace string, key string) ([]string, bool) {
	// Get all routes in the system, as node create/update affects all routes in the system
	if !lib.IsNodePortMode() || utils.GetInformers().RouteInformer == nil {
		return nil, false
	}
	routes := []string{}
	routeObjs, err := utils.GetInformers().RouteInformer.Lister().List(labels.Set(nil).AsSelector())
	if err != nil {
		utils.AviLog.Errorf("Unable to retrieve the routes: %s", err)
		return nil, false
	}

	for _, routeObj := range routeObjs {
		routes = append(routes, routeObj.Name)
	}
	if len(routes) == 0 {
		return nil, false
	}
	utils.AviLog.Debugf("key: %s, msg: total routes retrieved:  %s", key, routes)
	return routes, true
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

func HostRuleToIng(hrname string, namespace string, key string) ([]string, bool) {
	var err error
	var oldFqdn, fqdn string
	var oldFound bool

	allIngresses := make([]string, 0)
	hostrule, err := lib.GetCRDInformers().HostRuleInformer.Lister().HostRules(namespace).Get(hrname)
	if errors.IsNotFound(err) {
		utils.AviLog.Debugf("key: %s, msg: HostRule Deleted\n", key)
		_, fqdn = objects.SharedCRDLister().GetHostruleToFQDNMapping(namespace + "/" + hrname)
		objects.SharedCRDLister().DeleteHostruleFQDNMapping(namespace + "/" + hrname)
	} else if err != nil {
		utils.AviLog.Errorf("key: %s, msg: Error getting hostrule: %v\n", key, err)
		return nil, false
	} else {
		if err = validateHostRuleObj(key, hostrule); err != nil {
			return allIngresses, false
		}

		fqdn = hostrule.Spec.VirtualHost.Fqdn
		oldFound, oldFqdn = objects.SharedCRDLister().GetHostruleToFQDNMapping(namespace + "/" + hrname)
		if oldFound {
			objects.SharedCRDLister().DeleteHostruleFQDNMapping(namespace + "/" + hrname)
		}
		objects.SharedCRDLister().UpdateFQDNHostruleMapping(fqdn, namespace+"/"+hrname)
	}

	// find ingresses with host==fqdn, across all namespaces
	ok, obj := SharedHostNameLister().GetHostPathStore(fqdn)
	if !ok {
		utils.AviLog.Debugf("key: %s, msg: Couldn't find hostpath info for host: %s in cache", key, fqdn)
	} else {
		for _, ingresses := range obj {
			for _, ing := range ingresses {
				if !utils.HasElem(allIngresses, ing) {
					allIngresses = append(allIngresses, ing)
				}
			}
		}
	}

	// in case the hostname is updated, we need to find ingresses for the old ones as well to recompute
	if oldFound {
		ok, oldobj := SharedHostNameLister().GetHostPathStore(oldFqdn)
		if !ok {
			utils.AviLog.Debugf("key: %s, msg: Couldn't find hostpath info for host: %s in cache", key, oldFqdn)
		} else {
			for _, ingresses := range oldobj {
				for _, ing := range ingresses {
					if !utils.HasElem(allIngresses, ing) {
						allIngresses = append(allIngresses, ing)
					}
				}
			}
		}
	}

	utils.AviLog.Infof("key: %s, msg: ingresses to compute: %v via hostrule %s",
		key, allIngresses, namespace+"/"+hrname)
	return allIngresses, true
}

func HTTPRuleToIng(rrname string, namespace string, key string) ([]string, bool) {
	var err error
	allIngresses := make([]string, 0)
	httprule, err := lib.GetCRDInformers().HTTPRuleInformer.Lister().HTTPRules(namespace).Get(rrname)

	var hostrule string
	var oldFqdn, fqdn string
	oldPathRules := make(map[string]string)
	pathRules := make(map[string]string)
	var ok bool

	if errors.IsNotFound(err) {
		utils.AviLog.Debugf("key: %s, msg: HTTPRule Deleted\n", key)
		_, oldFqdn = objects.SharedCRDLister().GetHTTPRuleFqdnMapping(namespace + "/" + rrname)
		_, x := objects.SharedCRDLister().GetFqdnHTTPRulesMapping(oldFqdn)
		for i, elem := range x {
			oldPathRules[i] = elem
		}
		objects.SharedCRDLister().RemoveFqdnHTTPRulesMappings(namespace + "/" + rrname)
	} else if err != nil {
		utils.AviLog.Errorf("key: %s, msg: Error getting httprule: %v\n", key, err)
		return nil, false
	} else {
		utils.AviLog.Debugf("key: %s, HTTPRule %v\n", key, httprule)
		if err = validateHTTPRuleObj(key, httprule); err != nil {
			return allIngresses, false
		}

		_, oldFqdn = objects.SharedCRDLister().GetHTTPRuleFqdnMapping(namespace + "/" + rrname)
		_, x := objects.SharedCRDLister().GetFqdnHTTPRulesMapping(oldFqdn)
		for i, elem := range x {
			oldPathRules[i] = elem
		}

		fqdn = httprule.Spec.Fqdn
		objects.SharedCRDLister().RemoveFqdnHTTPRulesMappings(namespace + "/" + rrname)
		for _, path := range httprule.Spec.Paths {
			objects.SharedCRDLister().UpdateFqdnHTTPRulesMappings(fqdn, path.Target, namespace+"/"+rrname)
		}

		ok, pathRules = objects.SharedCRDLister().GetFqdnHTTPRulesMapping(fqdn)
		if !ok {
			utils.AviLog.Debugf("key: %s, msg: Couldn't find httprules for hostrule %s in cache", key, hostrule)
		}
	}

	// pathprefix match
	// lets say path: / and available paths registered in the cache could be keyed to /foo, /bar
	// in that case pathprefix match must account for both paths
	ok, pathIngs := SharedHostNameLister().GetHostPathStore(fqdn)
	if !ok {
		utils.AviLog.Debugf("key %s, msg: Couldn't find hostpath info for host: %s in cache", key, fqdn)
	} else {
		for pathPrefix, _ := range pathRules {
			re := regexp.MustCompile(fmt.Sprintf(`^%s.*`, strings.ReplaceAll(pathPrefix, `/`, `\/`)))
			for path, ingresses := range pathIngs {
				if !re.MatchString(path) {
					continue
				}
				utils.AviLog.Debugf("key: %s, msg: Computing for path %s in ingresses %v", key, path, ingresses)
				for _, ing := range ingresses {
					if !utils.HasElem(allIngresses, ing) {
						allIngresses = append(allIngresses, ing)
					}
				}
			}
		}
	}

	ok, oldPathIngs := SharedHostNameLister().GetHostPathStore(oldFqdn)
	if !ok {
		utils.AviLog.Debugf("key %s, msg: Couldn't find hostpath info for host: %s in cache", key, oldFqdn)
	} else {
		for oldPathPrefix, _ := range oldPathRules {
			re := regexp.MustCompile(fmt.Sprintf(`^%s.*`, strings.ReplaceAll(oldPathPrefix, `/`, `\/`)))
			for oldPath, oldIngresses := range oldPathIngs {
				if !re.MatchString(oldPath) {
					continue
				}
				utils.AviLog.Debugf("key: %s, msg: Computing for oldPath %s in oldIngresses %v", key, oldPath, oldIngresses)
				for _, oldIng := range oldIngresses {
					if !utils.HasElem(allIngresses, oldIng) {
						allIngresses = append(allIngresses, oldIng)
					}
				}
			}
		}
	}

	utils.AviLog.Infof("key: %s, msg: ingresses to compute: %v via httprule %s",
		key, allIngresses, namespace+"/"+rrname)
	return allIngresses, true
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

func parseL4ServiceForGateway(svc *corev1.Service, key string) (string, []string) {
	var gateway string
	var portProtocols []string

	labels := svc.GetLabels()
	if name, ok := labels[lib.GatewayNameLabelKey]; ok {
		if namespace, ok := labels[lib.GatewayNamespaceLabelKey]; ok {
			gateway = namespace + "/" + name
		}
	}
	if gateway != "" {
		for _, listener := range svc.Spec.Ports {
			if listener.Port != 0 && listener.Protocol != "" {
				portProtocols = append(portProtocols, fmt.Sprintf("%s/%d", listener.Protocol, listener.Port))
			}
		}
	}

	return gateway, portProtocols
}

func parseGatewayForListeners(gateway *advl4v1alpha1pre1.Gateway, key string) []string {
	var listeners []string
	for _, listener := range gateway.Spec.Listeners {
		gwName, nameOk := listener.Routes.RouteSelector.MatchLabels[lib.GatewayNameLabelKey]
		gwNamespace, nsOk := listener.Routes.RouteSelector.MatchLabels[lib.GatewayNamespaceLabelKey]
		if nameOk && nsOk && gwName == gateway.Name && gwNamespace == gateway.Namespace {
			listeners = append(listeners, fmt.Sprintf("%s/%d", listener.Protocol, listener.Port))
		}
	}
	return listeners
}

func filterIngressOnClass(ingress *v1beta1.Ingress) bool {
	// If Avi is not the default ingress, then filter on ingress class.
	if !lib.GetDefaultIngController() {
		annotations := ingress.GetAnnotations()
		ingClass, ok := annotations[lib.INGRESS_CLASS_ANNOT]
		if ok && ingClass == lib.AVI_INGRESS_CLASS {
			return true
		} else {
			utils.AviLog.Infof("AKO is not running as the default ingress controller. Not processing the ingress :%s . Please annotate the ingress class as 'avi'", ingress.Name)
			return false
		}
	} else {
		// If Avi is the default ingress controller, sync everything than the ones that are annotated with ingress class other than 'avi'
		annotations := ingress.GetAnnotations()
		ingClass, ok := annotations[lib.INGRESS_CLASS_ANNOT]
		if ok && ingClass != lib.AVI_INGRESS_CLASS {
			utils.AviLog.Infof("AKO is the default ingress controller but not processing the ingress :%s since ingress class is set to : %s", ingress.Name, ingClass)
			return false
		} else {
			return true
		}
	}
}
