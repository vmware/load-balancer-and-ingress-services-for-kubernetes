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
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	routev1 "github.com/openshift/api/route/v1"
	advl4v1alpha1pre1 "github.com/vmware-tanzu/service-apis/apis/v1alpha1pre1"
	corev1 "k8s.io/api/core/v1"
	networkingv1beta1 "k8s.io/api/networking/v1beta1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
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
	IngressClass = GraphSchema{
		Type:               "IngressClass",
		GetParentIngresses: IngClassToIng,
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
		GetParentRoutes:    SecretToRoute,
		GetParentGateways:  SecretToGateway,
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
		GetParentGateways: GWClassToGateway,
	}
	SupportedGraphTypes = GraphDescriptor{
		Ingress,
		IngressClass,
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
		if k8serrors.IsNotFound(err) {
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
			if routeObj.Spec.TLS.Certificate == "" || routeObj.Spec.TLS.Key == "" {
				secret = lib.GetDefaultSecretForRoutes()
			}
			akoNS := utils.GetAKONamespace()
			objects.OshiftRouteSvcLister().IngressMappings(namespace).AddIngressToSecretsMappings(akoNS, routeName, secret)
			objects.OshiftRouteSvcLister().IngressMappings(akoNS).AddSecretsToIngressMappings(namespace, routeName, secret)
		}
	}
	return routes, true
}

func SvcToRoute(svcName string, namespace string, key string) ([]string, bool) {
	_, err := utils.GetInformers().ServiceInformer.Lister().Services(namespace).Get(svcName)
	if err != nil && k8serrors.IsNotFound(err) {
		// Garbage collect the svc if no route references exist
		_, routes := objects.OshiftRouteSvcLister().IngressMappings(namespace).GetSvcToIng(svcName)
		if len(routes) == 0 {
			objects.OshiftRouteSvcLister().IngressMappings(namespace).DeleteSvcToIngMapping(svcName)
		}
	}
	_, routes := objects.OshiftRouteSvcLister().IngressMappings(namespace).GetSvcToIng(svcName)
	utils.AviLog.Debugf("key: %s, msg: total Routes retrieved: %v", key, routes)
	if len(routes) == 0 {
		return nil, false
	}
	return routes, true
}

func SvcToGateway(svcName string, namespace string, key string) ([]string, bool) {
	var allGateways []string
	svcNSName := namespace + "/" + svcName

	myService, err := utils.GetInformers().ServiceInformer.Lister().Services(namespace).Get(svcName)
	if err != nil && k8serrors.IsNotFound(err) {
		// Garbage collect the svc if no route references exist
		found, gateway := objects.ServiceGWLister().GetSvcToGw(svcNSName)
		if found {
			objects.ServiceGWLister().RemoveGatewayMappings(gateway, svcNSName)
			allGateways = append(allGateways, gateway)
		}
	} else {
		foundOld, oldGateway := objects.ServiceGWLister().GetSvcToGw(svcNSName)
		if foundOld {
			objects.ServiceGWLister().RemoveGatewayMappings(oldGateway, svcNSName)
			if !utils.HasElem(allGateways, oldGateway) {
				allGateways = append(allGateways, oldGateway)
			}
		}

		if gateway, svcPortProtocols := ParseL4ServiceForGateway(myService, key); gateway != "" {
			_, svcListeners := objects.ServiceGWLister().GetGwToSvcs(gateway)
			newSvcListeners := svcListeners
			for _, portProto := range svcPortProtocols {
				if val, ok := newSvcListeners[portProto]; ok && !utils.HasElem(val, svcNSName) {
					newSvcListeners[portProto] = append(val, svcNSName)
				} else {
					newSvcListeners[portProto] = []string{svcNSName}
				}
			}

			for portProto, svcs := range svcListeners {
				if utils.HasElem(svcs, svcNSName) && !utils.HasElem(svcPortProtocols, portProto) {
					svcs = utils.Remove(svcs, svcNSName)
					newSvcListeners[portProto] = svcs
				}
			}

			objects.ServiceGWLister().UpdateGatewayMappings(gateway, newSvcListeners, svcNSName)
			if !utils.HasElem(allGateways, gateway) {
				allGateways = append(allGateways, gateway)
			}
		}
	}

	utils.AviLog.Infof("key: %s, msg: Gateways retrieved %v", key, allGateways)
	return allGateways, true
}

func EPToRoute(epName string, namespace string, key string) ([]string, bool) {
	routes, found := SvcToRoute(epName, namespace, key)
	utils.AviLog.Debugf("key: %s, msg: total Routes retrieved: %v", key, routes)
	return routes, found
}

func EPToGateway(epName string, namespace string, key string) ([]string, bool) {
	gateways, found := SvcToGateway(epName, namespace, key)
	utils.AviLog.Debugf("key: %s, msg: total Gateways retrieved: %v", key, gateways)
	return gateways, found
}

func GatewayChanges(gwName string, namespace string, key string) ([]string, bool) {
	var allGateways []string
	allGateways = append(allGateways, namespace+"/"+gwName)
	gateway, err := lib.GetAdvL4Informers().GatewayInformer.Lister().Gateways(namespace).Get(gwName)
	if err != nil && k8serrors.IsNotFound(err) {
		// Remove all the Gateway to Services mapping.
		objects.ServiceGWLister().DeleteGWListeners(namespace + "/" + gwName)
		objects.ServiceGWLister().RemoveGatewayGWclassMappings(namespace + "/" + gwName)
	} else {
		if gwListeners := parseGatewayForListeners(gateway, key); len(gwListeners) > 0 {
			objects.ServiceGWLister().UpdateGWListeners(namespace+"/"+gwName, gwListeners)
			objects.ServiceGWLister().UpdateGatewayGWclassMappings(namespace+"/"+gwName, gateway.Spec.Class)
		} else {
			objects.ServiceGWLister().RemoveGatewayGWclassMappings(namespace + "/" + gwName)
			objects.ServiceGWLister().DeleteGWListeners(namespace + "/" + gwName)
		}
	}

	utils.AviLog.Debugf("key: %s, msg: total Gateways retrieved: %v", key, allGateways)
	return allGateways, true
}

func GWClassToGateway(gwClassName string, namespace string, key string) ([]string, bool) {
	found, gateways := objects.ServiceGWLister().GetGWclassToGateways(gwClassName)
	return gateways, found
}

func IngressChanges(ingName string, namespace string, key string) ([]string, bool) {
	var ingresses []string
	ingresses = append(ingresses, ingName)
	ingObj, err := utils.GetInformers().IngressInformer.Lister().Ingresses(namespace).Get(ingName)

	if err != nil {
		// Detect a delete condition here.
		if k8serrors.IsNotFound(err) {
			// Remove all the Ingress to Services mapping.
			// Remove the references of this ingress from the Services
			objects.SharedSvcLister().IngressMappings(namespace).RemoveIngressMappings(ingName)
			objects.SharedSvcLister().IngressMappings("").RemoveIngressClassMappings(ingName)
		}
	} else {
		// simple validator check for duplicate hostpaths, logs Warning if duplicates found
		success := validateSpecFromHostnameCache(key, ingObj.Namespace, ingObj.Name, ingObj.Spec)
		if !success {
			return ingresses, false
		}

		if ingObj.Spec.IngressClassName != nil {
			objects.SharedSvcLister().IngressMappings("").UpdateIngressClassMappings(namespace+"/"+ingName, *ingObj.Spec.IngressClassName)
		} else {
			objects.SharedSvcLister().IngressMappings("").RemoveIngressClassMappings(ingName)
		}

		services := parseServicesForIngress(ingObj.Spec, key)
		for _, svc := range services {
			utils.AviLog.Debugf("key: %s, msg: updating ingress relationship for service:  %s", key, svc)
			objects.SharedSvcLister().IngressMappings(namespace).UpdateIngressMappings(ingName, svc)
		}
		secrets := parseSecretsForIngress(ingObj.Spec, key)
		if len(secrets) > 0 {
			for _, secret := range secrets {
				objects.SharedSvcLister().IngressMappings(namespace).AddIngressToSecretsMappings(namespace, ingName, secret)
				objects.SharedSvcLister().IngressMappings(namespace).AddSecretsToIngressMappings(namespace, ingName, secret)
			}
		}
	}
	return ingresses, true
}

func IngClassToIng(ingClassName string, namespace string, key string) ([]string, bool) {
	found, ingresses := objects.SharedSvcLister().IngressMappings("").GetClassToIng(ingClassName)
	return ingresses, found
}

func SvcToIng(svcName string, namespace string, key string) ([]string, bool) {
	_, err := utils.GetInformers().ServiceInformer.Lister().Services(namespace).Get(svcName)
	if err != nil {
		// Detect a delete condition here.
		if k8serrors.IsNotFound(err) {
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
	// As node create/update affects all ingresses in the system
	// Post this, filtered ingresses for each service is fetched for all services.
	if !lib.IsNodePortMode() || utils.GetInformers().IngressInformer == nil {
		return nil, false
	}
	ingresses := []string{}
	return ingresses, true

}

func NodeToRoute(nodeName string, namespace string, key string) ([]string, bool) {
	// As node create/update affects all routes in the system return true in NodePort mode.
	// post this, filtered routes for each service is fetched for all services.
	if !lib.IsNodePortMode() || utils.GetInformers().RouteInformer == nil {
		return nil, false
	}
	routes := []string{}
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

func SecretToRoute(secretName string, namespace string, key string) ([]string, bool) {
	ok, ingNames := objects.OshiftRouteSvcLister().IngressMappings(namespace).GetSecretToIng(secretName)
	utils.AviLog.Debugf("key:%s, msg: Ingresses associated with the secret are: %s", key, ingNames)
	if ok {
		return ingNames, true
	}
	return nil, false
}

func SecretToGateway(secretName string, namespace string, key string) ([]string, bool) {
	return nil, false
}

func parseServicesForRoute(routeSpec routev1.RouteSpec, key string) []string {
	// Figure out the service names that are part of this route
	var services []string

	services = append(services, routeSpec.To.Name)
	for _, ab := range routeSpec.AlternateBackends {
		services = append(services, ab.Name)
	}

	utils.AviLog.Debugf("key: %s, msg: total services retrieved from route: %v", key, services)
	return services
}

func HostRuleToIng(hrname string, namespace string, key string) ([]string, bool) {
	var err error
	var oldFqdn, fqdn string
	var oldFound bool

	allIngresses := make([]string, 0)
	hostrule, err := lib.GetCRDInformers().HostRuleInformer.Lister().HostRules(namespace).Get(hrname)
	if k8serrors.IsNotFound(err) {
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

	if k8serrors.IsNotFound(err) {
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

func parseServicesForIngress(ingSpec networkingv1beta1.IngressSpec, key string) []string {
	// Figure out the service names that are part of this ingress
	var services []string
	for _, rule := range ingSpec.Rules {
		if rule.IngressRuleValue.HTTP != nil {
			for _, path := range rule.IngressRuleValue.HTTP.Paths {
				services = append(services, path.Backend.ServiceName)
			}
		}
	}
	utils.AviLog.Debugf("key: %s, msg: total services retrieved  from corev1:  %s", key, services)
	return services
}

func parseSecretsForIngress(ingSpec networkingv1beta1.IngressSpec, key string) []string {
	// Figure out the service names that are part of this ingress
	var secrets []string
	for _, tlsSettings := range ingSpec.TLS {
		secrets = append(secrets, tlsSettings.SecretName)
	}
	utils.AviLog.Debugf("key: %s, msg: total secrets retrieved from corev1:  %s", key, secrets)
	return secrets
}

func ParseL4ServiceForGateway(svc *corev1.Service, key string) (string, []string) {
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

func validateGatewayForClass(key string, gateway *advl4v1alpha1pre1.Gateway) error {
	gwClassObj, err := lib.GetAdvL4Informers().GatewayClassInformer.Lister().Get(gateway.Spec.Class)
	if err != nil {
		utils.AviLog.Errorf("key: %s, msg: Unable to fetch corresponding networking.x-k8s.io/gatewayclass %s %v",
			key, gateway.Spec.Class, err)
		return err
	}

	for _, listener := range gateway.Spec.Listeners {
		gwName, nameOk := listener.Routes.RouteSelector.MatchLabels[lib.GatewayNameLabelKey]
		gwNamespace, nsOk := listener.Routes.RouteSelector.MatchLabels[lib.GatewayNamespaceLabelKey]
		if !nameOk || !nsOk ||
			(nameOk && gwName != gateway.Name) ||
			(nsOk && gwNamespace != gateway.Namespace) {
			return errors.New("Incorrect gateway matchLabels configuration")
		}
	}

	// Additional check to see if the gatewayclass is a valid avi gateway class or not.
	if gwClassObj.Spec.Controller != lib.AviGatewayController {
		// Return an error since this is not our object.
		return errors.New("Unexpected controller")
	}

	return nil
}

func filterIngressOnClassAnnotation(key string, ingress *networkingv1beta1.Ingress) bool {
	// If Avi is not the default ingress, then filter on ingress class.
	if !lib.GetDefaultIngController() {
		annotations := ingress.GetAnnotations()
		ingClass, ok := annotations[lib.INGRESS_CLASS_ANNOT]
		if ok && ingClass == lib.AVI_INGRESS_CLASS {
			return true
		} else {
			utils.AviLog.Infof("key: %s, msg: AKO is not running as the default ingress controller. Not processing the ingress: %s. Please annotate the ingress class as 'avi'", key, ingress.Name)
			return false
		}
	} else {
		// If Avi is the default ingress controller, sync everything than the ones that are annotated with ingress class other than 'avi'
		annotations := ingress.GetAnnotations()
		ingClass, ok := annotations[lib.INGRESS_CLASS_ANNOT]
		if ok && ingClass != lib.AVI_INGRESS_CLASS {
			utils.AviLog.Infof("key: %s, msg: AKO is the default ingress controller but not processing the ingress: %s since ingress class is set to : %s", key, ingress.Name, ingClass)
			return false
		} else {
			return true
		}
	}
}

func validateIngressForClass(key string, ingress *networkingv1beta1.Ingress) bool {
	// see whether ingress class resources are present or not
	if !utils.GetIngressClassEnabled() {
		return filterIngressOnClassAnnotation(key, ingress)
	}

	if ingress.Spec.IngressClassName == nil {
		// check whether avi-lb ingress class is set as the default ingress class
		if isAviLBDefaultIngressClass() {
			utils.AviLog.Debugf("key: %s, msg: ingress class name is not specified but ako.vmware.com/avi-lb is default ingress controller", key)
			return true
		} else {
			utils.AviLog.Warnf("key: %s, msg: ingress class name not specified for ingress %s and ako.vmware.com/avi-lb is not default ingress controller", key, ingress.Name)
			return false
		}
	}

	ingClassObj, err := utils.GetInformers().IngressClassInformer.Lister().Get(*ingress.Spec.IngressClassName)
	if err != nil {
		utils.AviLog.Warnf("key: %s, msg: Unable to fetch corresponding networking.k8s.io/ingressclass %s %v",
			key, *ingress.Spec.IngressClassName, err)
		return false
	}

	// Additional check to see if the gatewayclass is a valid avi gateway class or not.
	if ingClassObj.Spec.Controller != lib.AviIngressController {
		// Return an error since this is not our object.
		utils.AviLog.Warnf("key: %s, msg: Unexpected controller in ingress class %s", key, *ingress.Spec.IngressClassName)
		return false
	}

	return true
}

func isAviLBDefaultIngressClass() bool {
	ingClassObjs, _ := utils.GetInformers().IngressClassInformer.Lister().List(labels.Set(nil).AsSelector())
	for _, ingClass := range ingClassObjs {
		if ingClass.Spec.Controller == lib.AviIngressController {
			annotations := ingClass.GetAnnotations()
			isDefaultClass, ok := annotations[lib.DefaultIngressClassAnnotation]
			if ok && isDefaultClass == "true" {
				return true
			}
		}
	}

	utils.AviLog.Debugf("IngressClass with controller ako.vmware.com/avi-lb not found in the cluster")
	return false
}
