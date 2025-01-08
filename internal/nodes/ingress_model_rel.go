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
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/status"

	akov1beta1 "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/apis/ako/v1beta1"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	routev1 "github.com/openshift/api/route/v1"
	advl4v1alpha1pre1 "github.com/vmware-tanzu/service-apis/apis/v1alpha1pre1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	servicesapi "sigs.k8s.io/service-apis/apis/v1alpha1"
	svcapiv1alpha1 "sigs.k8s.io/service-apis/apis/v1alpha1"
)

var (
	Service = GraphSchema{
		Type:                           "Service",
		GetParentIngresses:             SvcToIng,
		GetParentRoutes:                SvcToRoute,
		GetParentGateways:              SvcToGateway,
		GetParentMultiClusterIngresses: SvcToMultiClusterIng,
	}
	SharedVipService = GraphSchema{
		Type:              "SharedVipService",
		GetParentServices: ServiceChanges,
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

	EndpointSlices = GraphSchema{
		Type:               utils.Endpointslices,
		GetParentIngresses: EPToIng,
		GetParentRoutes:    EPToRoute,
		GetParentGateways:  EPToGateway,
	}

	Pod = GraphSchema{
		Type:               "Pod",
		GetParentIngresses: PodToIng,
	}
	Node = GraphSchema{
		Type:               "Node",
		GetParentIngresses: NodeToIng,
		GetParentRoutes:    NodeToRoute,
	}
	Secret = GraphSchema{
		Type:                           "Secret",
		GetParentIngresses:             SecretToIng,
		GetParentRoutes:                SecretToRoute,
		GetParentGateways:              SecretToGateway,
		GetParentMultiClusterIngresses: SecretToMultiClusterIng,
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
	AviInfraSetting = GraphSchema{
		Type:               "AviInfraSetting",
		GetParentIngresses: AviSettingToIng,
		GetParentGateways:  AviSettingToGateway,
		GetParentServices:  AviSettingToSvc,
		GetParentRoutes:    AviSettingToRoute,
	}
	MultiClusterIngress = GraphSchema{
		Type:                           lib.MultiClusterIngress,
		GetParentMultiClusterIngresses: MultiClusterIngressChanges,
	}
	ServiceImport = GraphSchema{
		Type:                           lib.ServiceImport,
		GetParentMultiClusterIngresses: ServiceImportToMultiClusterIng,
	}
	SSORule = GraphSchema{
		Type:               lib.SSORule,
		GetParentIngresses: SSORuleToIng,
		GetParentRoutes:    SSORuleToIng,
	}

	SupportedGraphTypes = GraphDescriptor{
		Ingress,
		IngressClass,
		Service,
		SharedVipService,
		Pod,
		Endpoint,
		EndpointSlices,
		Secret,
		Route,
		Node,
		HostRule,
		HTTPRule,
		Gateway,
		GatewayClass,
		AviInfraSetting,
		MultiClusterIngress,
		ServiceImport,
		SSORule,
		L4Rule,
	}
)

type GraphSchema struct {
	Type                           string
	GetParentIngresses             func(string, string, string) ([]string, bool)
	GetParentRoutes                func(string, string, string) ([]string, bool)
	GetParentGateways              func(string, string, string) ([]string, bool)
	GetParentServices              func(string, string, string) ([]string, bool)
	GetParentMultiClusterIngresses func(string, string, string) ([]string, bool)
}

type GraphDescriptor []GraphSchema

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
			utils.AviLog.Debugf("key: %s, msg: updating route relationship for service: %s", key, svc)
			objects.OshiftRouteSvcLister().IngressMappings(namespace).UpdateIngressMappings(routeName, svc)
		}
		if routeObj.Spec.TLS != nil {
			akoNS := utils.GetAKONamespace()
			secret := lib.RouteSecretsPrefix + routeName
			if routeObj.Spec.TLS.Certificate == "" || routeObj.Spec.TLS.Key == "" {
				secret = lib.GetDefaultSecretForRoutes()
			}
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
	utils.AviLog.Debugf("key: %s, msg: Routes retrieved %s", key, routes)
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

	utils.AviLog.Debugf("key: %s, msg: Gateways retrieved %s", key, allGateways)
	return allGateways, true
}

func EPToRoute(epName string, namespace string, key string) ([]string, bool) {
	routes, found := SvcToRoute(epName, namespace, key)
	utils.AviLog.Debugf("key: %s, msg: Routes retrieved %s", key, routes)
	return routes, found
}

func EPToGateway(epName string, namespace string, key string) ([]string, bool) {
	gateways, found := SvcToGateway(epName, namespace, key)
	utils.AviLog.Debugf("key: %s, msg: Gateways retrieved %s", key, gateways)
	return gateways, found
}

func GatewayChanges(gwName string, namespace string, key string) ([]string, bool) {
	var allGateways []string
	allGateways = append(allGateways, namespace+"/"+gwName)
	if utils.IsWCP() {
		gateway, err := lib.AKOControlConfig().AdvL4Informers().GatewayInformer.Lister().Gateways(namespace).Get(gwName)
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
	} else if lib.UseServicesAPI() {
		gateway, err := lib.AKOControlConfig().SvcAPIInformers().GatewayInformer.Lister().Gateways(namespace).Get(gwName)
		if err != nil && k8serrors.IsNotFound(err) {
			// Remove all the Gateway to Services mapping.
			objects.ServiceGWLister().DeleteGWListeners(namespace + "/" + gwName)
			objects.ServiceGWLister().RemoveGatewayGWclassMappings(namespace + "/" + gwName)
		} else {
			if gwListeners := parseSvcApiGatewayForListeners(gateway, key); len(gwListeners) > 0 {
				objects.ServiceGWLister().UpdateGWListeners(namespace+"/"+gwName, gwListeners)
				objects.ServiceGWLister().UpdateGatewayGWclassMappings(namespace+"/"+gwName, gateway.Spec.GatewayClassName)
			} else {
				objects.ServiceGWLister().RemoveGatewayGWclassMappings(namespace + "/" + gwName)
				objects.ServiceGWLister().DeleteGWListeners(namespace + "/" + gwName)
			}
		}
	}
	return allGateways, true
}

func GWClassToGateway(gwClassName string, namespace string, key string) ([]string, bool) {
	found, gateways := objects.ServiceGWLister().GetGWclassToGateways(gwClassName)
	utils.AviLog.Debugf("key: %s, msg: Gateways retrieved %s", key, gateways)
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
			svcToDel := objects.SharedSvcLister().IngressMappings(namespace).RemoveIngressMappings(ingName)
			if lib.AutoAnnotateNPLSvc() {
				for _, svc := range svcToDel {
					statusOption := status.StatusOptions{
						ObjType:   lib.NPLService,
						Op:        lib.DeleteStatus,
						ObjName:   svc,
						Namespace: namespace,
						Key:       key,
					}
					status.PublishToStatusQueue(svc, statusOption)
				}
			}
			objects.SharedSvcLister().IngressMappings(metav1.NamespaceAll).RemoveIngressClassMappings(namespace + "/" + ingName)
		}
	} else {
		// simple validator check for duplicate hostpaths, logs Warning if duplicates found
		success := validateSpecFromHostnameCache(key, ingObj)
		if !success {
			return ingresses, false
		}

		var ingClassName string
		if ingObj.Spec.IngressClassName != nil {
			ingClassName = *ingObj.Spec.IngressClassName
		} else if aviIngClassName, found := lib.IsAviLBDefaultIngressClass(); found {
			// check if default IngressClass is present, and it is Avi's IngressClass, in which case add mapping for that.
			ingClassName = aviIngClassName
		}

		if ingClassName != "" {
			objects.SharedSvcLister().IngressMappings(metav1.NamespaceAll).UpdateIngressClassMappings(namespace+"/"+ingName, ingClassName)
		} else {
			objects.SharedSvcLister().IngressMappings(metav1.NamespaceAll).RemoveIngressClassMappings(namespace + "/" + ingName)
		}

		// If the Ingress Class is not found or is not valid, then return.
		// When the correct Ingress Class is added, then the Ingress would be processed again.
		if !lib.ValidateIngressForClass(key, ingObj) {
			svcToDel := objects.SharedSvcLister().IngressMappings(namespace).RemoveIngressMappings(ingName)
			if lib.AutoAnnotateNPLSvc() {
				for _, svc := range svcToDel {
					statusOption := status.StatusOptions{
						ObjType:   lib.NPLService,
						Op:        lib.DeleteStatus,
						ObjName:   svc,
						Namespace: namespace,
						Key:       key,
					}
					status.PublishToStatusQueue(svc, statusOption)
				}
			}
			return ingresses, true
		}

		_, oldSvcs := objects.SharedSvcLister().IngressMappings(namespace).GetIngToSvc(ingName)
		currSvcs := parseServicesForIngress(ingObj.Spec, key)

		svcToDel := lib.Difference(oldSvcs, currSvcs)
		for _, svc := range svcToDel {
			_, ingrforSvc := objects.SharedSvcLister().IngressMappings(namespace).GetSvcToIng(svc)
			ingrforSvc = utils.Remove(ingrforSvc, ingName)
			if lib.AutoAnnotateNPLSvc() && len(ingrforSvc) == 0 {
				statusOption := status.StatusOptions{
					ObjType:   lib.NPLService,
					Op:        lib.DeleteStatus,
					ObjName:   svc,
					Namespace: namespace,
					Key:       key,
				}
				status.PublishToStatusQueue(svc, statusOption)
			}
			objects.SharedSvcLister().IngressMappings(namespace).RemoveSvcFromIngressMappings(ingName, svc)
		}

		svcToAdd := lib.Difference(currSvcs, oldSvcs)
		for _, svc := range svcToAdd {
			utils.AviLog.Debugf("key: %s, msg: updating ingress relationship for service:  %s", key, svc)
			objects.SharedSvcLister().IngressMappings(namespace).UpdateIngressMappings(ingName, svc)
			// Check and update NPl annotation for svc
			if lib.AutoAnnotateNPLSvc() {
				if !status.CheckNPLSvcAnnotation(key, namespace, svc) {
					statusOption := status.StatusOptions{
						ObjType:   lib.NPLService,
						Op:        lib.UpdateStatus,
						ObjName:   svc,
						Namespace: namespace,
						Key:       key,
					}
					status.PublishToStatusQueue(svc, statusOption)
				}
			}
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
	found, ingresses := objects.SharedSvcLister().IngressMappings(metav1.NamespaceAll).GetClassToIng(ingClassName)
	// Go through the list of ingresses again to populate the ingress Service mapping and annotate services if needed
	for _, namespacedIngr := range ingresses {
		ns, ingr := utils.ExtractNamespaceObjectName(namespacedIngr)
		IngressChanges(ingr, ns, key)
	}
	return ingresses, found
}

func SvcToIng(svcName string, namespace string, key string) ([]string, bool) {
	svc, err := utils.GetInformers().ServiceInformer.Lister().Services(namespace).Get(svcName)
	if err != nil {
		// Detect a delete condition here.
		if k8serrors.IsNotFound(err) {
			// Garbage collect the service if no ingress references exist
			_, ingresses := objects.SharedSvcLister().IngressMappings(namespace).GetSvcToIng(svcName)
			if len(ingresses) == 0 {
				objects.SharedSvcLister().IngressMappings(namespace).DeleteSvcToIngMapping(svcName)
			}
			return ingresses, true
		}
		return nil, false
	}

	if lib.AutoAnnotateNPLSvc() && svc.Spec.Type == corev1.ServiceTypeNodePort {
		statusOption := status.StatusOptions{
			ObjType:   lib.NPLService,
			Op:        lib.DeleteStatus,
			ObjName:   svcName,
			Namespace: namespace,
			Key:       key,
		}
		status.PublishToStatusQueue(svcName, statusOption)
	}

	_, ingresses := objects.SharedSvcLister().IngressMappings(namespace).GetSvcToIng(svcName)
	if len(ingresses) == 0 {
		return nil, false
	}

	// Check if the svc has the NPL Annotation. If not, annotate and exit without returning any ingress
	if lib.AutoAnnotateNPLSvc() {
		if !status.CheckNPLSvcAnnotation(key, namespace, svcName) {
			statusOption := status.StatusOptions{
				ObjType:   lib.NPLService,
				Op:        lib.UpdateStatus,
				ObjName:   svcName,
				Namespace: namespace,
				Key:       key,
			}
			status.PublishToStatusQueue(svcName, statusOption)
			return nil, false
		}
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

// PodToIng fetches the list of impacted Ingresses from Pod update.
// First fetch list of Services for the Pod.
// Then get list of Ingresses for the Services.
func PodToIng(podName string, namespace string, key string) ([]string, bool) {
	var allIngresses []string
	podKey := namespace + "/" + podName
	ok, servicesIntf := objects.SharedPodToSvcLister().Get(podKey)
	if !ok {
		return allIngresses, false
	}
	services := servicesIntf.([]string)
	utils.AviLog.Debugf("key: %s, msg: Services retrieved:  %s", key, services)
	for _, svc := range services {
		_, svcName := utils.ExtractNamespaceObjectName(svc)
		ingresses, _ := SvcToIng(svcName, namespace, key)
		allIngresses = append(allIngresses, ingresses...)
	}
	utils.AviLog.Debugf("key: %s, msg: Ingresses retrieved:  %s", key, allIngresses)
	return allIngresses, true
}

func EPToIng(epName string, namespace string, key string) ([]string, bool) {
	ingresses, found := SvcToIng(epName, namespace, key)
	utils.AviLog.Debugf("key: %s, msg: Ingresses retrieved %s", key, ingresses)
	return ingresses, found
}

func SecretToIng(secretName string, namespace string, key string) ([]string, bool) {
	ok, ingNames := objects.SharedSvcLister().IngressMappings(namespace).GetSecretToIng(secretName)
	utils.AviLog.Debugf("key: %s, msg: Ingresses retrieved %s", key, ingNames)
	if ok {
		return ingNames, true
	}
	return nil, false
}

func SecretToRoute(secretName string, namespace string, key string) ([]string, bool) {
	ok, ingNames := objects.OshiftRouteSvcLister().IngressMappings(namespace).GetSecretToIng(secretName)
	utils.AviLog.Debugf("key: %s, msg: Ingresses retrieved %s", key, ingNames)
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
	var fqdnType, oldFqdnType string
	var oldFound bool

	allIngresses := make([]string, 0)
	hostrule, err := lib.AKOControlConfig().CRDInformers().HostRuleInformer.Lister().HostRules(namespace).Get(hrname)
	if k8serrors.IsNotFound(err) {
		utils.AviLog.Debugf("key: %s, msg: HostRule Deleted", key)
		oldFound, oldFqdn = objects.SharedCRDLister().GetHostruleToFQDNMapping(namespace + "/" + hrname)
		if !strings.Contains(oldFqdn, lib.ShardVSSubstring) {
			objects.SharedCRDLister().DeleteHostruleFQDNMapping(namespace + "/" + hrname)
			oldFqdnType = objects.SharedCRDLister().GetFQDNFQDNTypeMapping(oldFqdn)
			objects.SharedCRDLister().DeleteFQDNFQDNTypeMapping(oldFqdn)
		}
	} else if err != nil {
		utils.AviLog.Errorf("key: %s, msg: Error getting hostrule: %v", key, err)
		return nil, false
	} else {
		if hostrule.Status.Status != lib.StatusAccepted {
			utils.AviLog.Errorf("key: %s, msg: Hostrule is not in accepted state", key)
			return []string{}, false
		}
		fqdn = hostrule.Spec.VirtualHost.Fqdn
		oldFound, oldFqdn = objects.SharedCRDLister().GetHostruleToFQDNMapping(namespace + "/" + hrname)
		if oldFound && !strings.Contains(oldFqdn, lib.ShardVSSubstring) {
			objects.SharedCRDLister().DeleteHostruleFQDNMapping(namespace + "/" + hrname)
			oldFqdnType = objects.SharedCRDLister().GetFQDNFQDNTypeMapping(oldFqdn)
		}
		if !strings.Contains(fqdn, lib.ShardVSSubstring) {
			objects.SharedCRDLister().UpdateFQDNHostruleMapping(fqdn, namespace+"/"+hrname)
			fqdnType = string(hostrule.Spec.VirtualHost.FqdnType)
			if fqdnType == "" {
				fqdnType = string(akov1beta1.Exact)
			}
			objects.SharedCRDLister().UpdateFQDNFQDNTypeMapping(fqdn, fqdnType)
		}
	}

	// in case the hostname is updated, we need to find ingresses for the old ones as well to recompute
	if oldFound {
		allOldFqdns := SharedHostNameLister().GetHostsFromHostPathStore(oldFqdn, oldFqdnType)
		for _, i := range allOldFqdns {
			ok, oldobj := SharedHostNameLister().GetHostPathStore(i)
			if !ok {
				utils.AviLog.Debugf("key: %s, msg: Couldn't find hostpath info for host: %s in cache", key, i)
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
	}

	// find ingresses with host==fqdn, across all namespaces
	allFqdns := SharedHostNameLister().GetHostsFromHostPathStore(fqdn, fqdnType)
	for _, i := range allFqdns {
		ok, obj := SharedHostNameLister().GetHostPathStore(i)
		if !ok {
			utils.AviLog.Debugf("key: %s, msg: Couldn't find hostpath info for host: %s in cache", key, i)
		} else {
			for _, ingresses := range obj {
				for _, ing := range ingresses {
					if !utils.HasElem(allIngresses, ing) {
						allIngresses = append(allIngresses, ing)
					}
				}
			}
		}
	}

	utils.AviLog.Infof("key: %s, msg: Ingresses retrieved %s", key, allIngresses)
	return allIngresses, true
}

func SSORuleToIng(srname string, namespace string, key string) ([]string, bool) {
	var err error
	var oldFqdn, fqdn string
	var fqdnType, oldFqdnType = string(akov1beta1.Exact), string(akov1beta1.Exact)
	var oldFound bool

	allIngresses := make([]string, 0)
	ssoRule, err := lib.AKOControlConfig().CRDInformers().SSORuleInformer.Lister().SSORules(namespace).Get(srname)
	if k8serrors.IsNotFound(err) {
		utils.AviLog.Debugf("key: %s, msg: SSORule Deleted", key)
		oldFound, oldFqdn = objects.SharedCRDLister().GetSSORuleToFQDNMapping(namespace + "/" + srname)
		if !strings.Contains(oldFqdn, lib.ShardVSSubstring) {
			objects.SharedCRDLister().DeleteSSORuleFQDNMapping(namespace + "/" + srname)
		}
	} else if err != nil {
		utils.AviLog.Errorf("key: %s, msg: Error getting SSORule: %v", key, err)
		return nil, false
	} else {
		if ssoRule.Status.Status != lib.StatusAccepted {
			return []string{}, false
		}
		fqdn = *ssoRule.Spec.Fqdn
		oldFound, oldFqdn = objects.SharedCRDLister().GetSSORuleToFQDNMapping(namespace + "/" + srname)
		if oldFound && !strings.Contains(oldFqdn, lib.ShardVSSubstring) {
			objects.SharedCRDLister().DeleteSSORuleFQDNMapping(namespace + "/" + srname)
		}
		if !strings.Contains(fqdn, lib.ShardVSSubstring) {
			objects.SharedCRDLister().UpdateFQDNSSORuleMapping(fqdn, namespace+"/"+srname)
		}
	}

	// in case the hostname is updated, we need to find ingresses for the old ones as well to recompute
	if oldFound {
		allOldFqdns := SharedHostNameLister().GetHostsFromHostPathStore(oldFqdn, oldFqdnType)
		for _, i := range allOldFqdns {
			ok, oldobj := SharedHostNameLister().GetHostPathStore(i)
			if !ok {
				utils.AviLog.Debugf("key: %s, msg: Couldn't find hostpath info for host: %s in cache", key, i)
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
	}

	// find ingresses with host==fqdn, across all namespaces
	allFqdns := SharedHostNameLister().GetHostsFromHostPathStore(fqdn, fqdnType)
	for _, i := range allFqdns {
		ok, obj := SharedHostNameLister().GetHostPathStore(i)
		if !ok {
			utils.AviLog.Debugf("key: %s, msg: Couldn't find hostpath info for host: %s in cache", key, i)
		} else {
			for _, ingresses := range obj {
				for _, ing := range ingresses {
					if !utils.HasElem(allIngresses, ing) {
						allIngresses = append(allIngresses, ing)
					}
				}
			}
		}
	}

	utils.AviLog.Infof("key: %s, msg: Ingresses retrieved %s", key, allIngresses)
	return allIngresses, true
}

func HTTPRuleToIng(rrname string, namespace string, key string) ([]string, bool) {
	var err error
	allIngresses := make([]string, 0)
	httprule, err := lib.AKOControlConfig().CRDInformers().HTTPRuleInformer.Lister().HTTPRules(namespace).Get(rrname)

	var hostrule string
	var oldFqdn, fqdn string
	oldPathRules := make(map[string]string)
	pathRules := make(map[string]string)
	var ok bool

	if k8serrors.IsNotFound(err) {
		utils.AviLog.Debugf("key: %s, msg: HTTPRule Deleted", key)
		_, oldFqdn = objects.SharedCRDLister().GetHTTPRuleFqdnMapping(namespace + "/" + rrname)
		_, x := objects.SharedCRDLister().GetFqdnHTTPRulesMapping(oldFqdn)
		for i, elem := range x {
			oldPathRules[i] = elem
		}
		objects.SharedCRDLister().RemoveFqdnHTTPRulesMappings(namespace + "/" + rrname)
	} else if err != nil {
		utils.AviLog.Errorf("key: %s, msg: Error getting httprule: %v", key, err)
		return nil, false
	} else {
		if httprule.Status.Status != lib.StatusAccepted {
			utils.AviLog.Errorf("key: %s, msg: HTTPRule is not in accepted state", key)
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
		for pathPrefix := range pathRules {
			re := regexp.MustCompile(fmt.Sprintf(`^%s.*`, strings.ReplaceAll(pathPrefix, `/`, `\/`)))
			for path, ingresses := range pathIngs {
				if path != "" && !re.MatchString(path) {
					continue
				}
				utils.AviLog.Debugf("key: %s, msg: Computing for path %s in ingresses %v", key, path, ingresses)
				for _, ing := range ingresses {
					ing_namespace, _, _ := lib.ExtractTypeNameNamespace(ing)
					// httprule is namespace specific. So only add those ingresses which are in same namespace of rule.
					if namespace == ing_namespace && !utils.HasElem(allIngresses, ing) {
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
		for oldPathPrefix := range oldPathRules {
			re := regexp.MustCompile(fmt.Sprintf(`^%s.*`, strings.ReplaceAll(oldPathPrefix, `/`, `\/`)))
			for oldPath, oldIngresses := range oldPathIngs {
				if oldPath != "" && !re.MatchString(oldPath) {
					continue
				}
				utils.AviLog.Debugf("key: %s, msg: Computing for oldPath %s in oldIngresses %v", key, oldPath, oldIngresses)
				for _, oldIng := range oldIngresses {
					ing_namespace, _, _ := lib.ExtractTypeNameNamespace(oldIng)
					// httprule is namespace specific. So only add those ingresses which are in same namespace of rule.
					if namespace == ing_namespace && !utils.HasElem(allIngresses, oldIng) {
						allIngresses = append(allIngresses, oldIng)
					}
				}
			}
		}
	}

	utils.AviLog.Debugf("key: %s, msg: Ingresses retrieved %s", key, allIngresses)
	return allIngresses, true
}

func AviSettingToIng(infraSettingName, namespace, key string) ([]string, bool) {
	allIngresses := make([]string, 0)

	// Get all IngressClasses from AviInfraSetting.
	ingClasses, err := utils.GetInformers().IngressClassInformer.Informer().GetIndexer().ByIndex(lib.AviSettingIngClassIndex, lib.AkoGroup+"/"+lib.AviInfraSetting+"/"+infraSettingName)
	if err != nil {
		utils.AviLog.Warnf("key: %s, msg: Unable to fetch IngressClasses corresponding to AviInfraSetting %s", key, infraSettingName)
		return allIngresses, false
	}

	for _, ingClass := range ingClasses {
		if ingClassObj, isIngClass := ingClass.(*networkingv1.IngressClass); isIngClass {
			if ingresses, found := IngClassToIng(ingClassObj.Name, namespace, key); found {
				allIngresses = append(allIngresses, ingresses...)
			}
		}
	}

	if nsIngresses, found := infraSettingNSToIngress(infraSettingName, key); found {
		// Go through the list of ingresses again to populate the ingress Service mapping and annotate services if needed
		for _, ingress := range nsIngresses {
			ns, ingr := utils.ExtractNamespaceObjectName(ingress)
			IngressChanges(ingr, ns, key)
		}
		allIngresses = append(allIngresses, nsIngresses...)
	}

	utils.AviLog.Infof("key: %s, msg: Ingresses retrieved %s", key, allIngresses)
	return allIngresses, true
}

func AviSettingToRoute(infraSettingName, namespace, key string) ([]string, bool) {
	allRoutes := make([]string, 0)

	// Get all Routes from AviInfraSetting via annotation.
	routes, err := utils.GetInformers().RouteInformer.Informer().GetIndexer().ByIndex(lib.AviSettingRouteIndex, infraSettingName)
	if err != nil {
		utils.AviLog.Warnf("key: %s, msg: Unable to fetch Routes corresponding to AviInfraSetting %s", key, infraSettingName)
		return allRoutes, false
	}

	if nsRoutes, found := infraSettingNSToRoutes(infraSettingName, key); found {
		routes = append(routes, nsRoutes...)
	}

	for _, route := range routes {
		if routeObj, isRoute := route.(*routev1.Route); isRoute {
			RouteChanges(routeObj.Name, routeObj.Namespace, key)
			routeNSName := routeObj.Namespace + "/" + routeObj.Name
			allRoutes = append(allRoutes, routeNSName)
		}
	}

	utils.AviLog.Infof("key: %s, msg: Routes retrieved %s", key, allRoutes)
	return allRoutes, true
}

func AviSettingToGateway(infraSettingName string, namespace string, key string) ([]string, bool) {
	allGateways := make([]string, 0)

	// Get all GatewayClasses from AviInfraSetting.
	if lib.UseServicesAPI() {
		gwClasses, err := lib.AKOControlConfig().SvcAPIInformers().GatewayClassInformer.Informer().GetIndexer().ByIndex(lib.AviSettingGWClassIndex, lib.AkoGroup+"/"+lib.AviInfraSetting+"/"+infraSettingName)
		if err != nil {
			utils.AviLog.Warnf("key: %s, msg: Unable to fetch GatewayClasses corresponding to AviInfraSetting %s", key, infraSettingName)
			return allGateways, false
		}

		for _, gwClass := range gwClasses {
			// Get all Gateways from GatewayClass.
			gwClassObj, isGwClass := gwClass.(*servicesapi.GatewayClass)
			if isGwClass {
				gateways, err := lib.AKOControlConfig().SvcAPIInformers().GatewayInformer.Informer().GetIndexer().ByIndex(lib.GatewayClassGatewayIndex, gwClassObj.Name)
				if err != nil {
					utils.AviLog.Warnf("key: %s, msg: Unable to fetch Gateways %v", key, err)
					continue
				}
				for _, gateway := range gateways {
					if gatewayObj, isGw := gateway.(*servicesapi.Gateway); isGw {
						allGateways = append(allGateways, gatewayObj.Namespace+"/"+gatewayObj.Name)
					}
				}
			}
		}
	}

	if nsGateways, found := infraSettingNSToGateway(infraSettingName, key); found {
		allGateways = append(allGateways, nsGateways...)
	}

	utils.AviLog.Debugf("key: %s, msg: Gateways retrieved %s", key, allGateways)
	return allGateways, true
}

func AviSettingToSvc(infraSettingName string, namespace string, key string) ([]string, bool) {
	allSvcs := make([]string, 0)

	// get all services that are affected by this infrasetting
	services, err := utils.GetInformers().ServiceInformer.Informer().GetIndexer().ByIndex(lib.AviSettingServicesIndex, infraSettingName)
	if err != nil {
		return allSvcs, false
	}

	for _, svc := range services {
		svcObj, isSvc := svc.(*corev1.Service)
		if isSvc {
			allSvcs = append(allSvcs, svcObj.Namespace+"/"+svcObj.Name)
		}
	}

	if nsServices, found := infraSettingNSToServices(infraSettingName, key); found {
		allSvcs = append(allSvcs, nsServices...)
	}

	utils.AviLog.Debugf("key: %s, msg: total services retrieved from AviInfraSettings: %s", key, allSvcs)
	return allSvcs, true
}

func ServiceChanges(serviceName, namespace, key string) ([]string, bool) {
	var vipKeys []string
	serviceNamespaceName := namespace + "/" + serviceName
	serviceObj, err := utils.GetInformers().ServiceInformer.Lister().Services(namespace).Get(serviceName)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			if found, oldKey := objects.SharedlbLister().GetServiceToSharedVipKey(serviceNamespaceName); found {
				vipKeys = append(vipKeys, oldKey)
			}
			objects.SharedlbLister().RemoveSharedVipKeyServiceMappings(serviceNamespaceName)
		}
	} else {
		found, oldKey := objects.SharedlbLister().GetServiceToSharedVipKey(serviceNamespaceName)
		if found {
			vipKeys = append(vipKeys, oldKey)
			objects.SharedlbLister().RemoveSharedVipKeyServiceMappings(serviceNamespaceName)
		}

		if currentKey, ok := serviceObj.Annotations[lib.SharedVipSvcLBAnnotation]; ok {
			if currentKey != oldKey {
				vipKeys = append(vipKeys, serviceObj.Namespace+"/"+currentKey)
			}
			objects.SharedlbLister().UpdateSharedVipKeyServiceMappings(serviceObj.Namespace+"/"+currentKey, serviceNamespaceName)
		}
	}
	return vipKeys, true
}

func parseServicesForIngress(ingSpec networkingv1.IngressSpec, key string) []string {
	// Figure out the service names that are part of this ingress
	var services []string
	for _, rule := range ingSpec.Rules {
		if rule.IngressRuleValue.HTTP != nil {
			for _, path := range rule.IngressRuleValue.HTTP.Paths {
				services = append(services, path.Backend.Service.Name)
			}
		}
	}
	utils.AviLog.Debugf("key: %s, msg: total services retrieved from corev1: %s", key, services)
	return services
}

func parseSecretsForIngress(ingSpec networkingv1.IngressSpec, key string) []string {
	// Figure out the service names that are part of this ingress
	var secrets []string
	for _, tlsSettings := range ingSpec.TLS {
		secrets = append(secrets, tlsSettings.SecretName)
	}
	utils.AviLog.Debugf("key: %s, msg: total secrets retrieved from corev1: %s", key, secrets)
	return secrets
}

func ParseL4ServiceForGateway(svc *corev1.Service, key string) (string, []string) {
	var gateway string
	var portProtocols []string
	if lib.UseServicesAPI() && svc.Spec.Type == corev1.ServiceTypeLoadBalancer {
		utils.AviLog.Infof("key: %s, msg: Service of Type LoadBalancer is not supported with Gateway APIs, will create dedicated VSes", key)
		return gateway, portProtocols
	}

	var gwNameLabel, gwNamespaceLabel string
	if utils.IsWCP() {
		gwNameLabel = lib.GatewayNameLabelKey
		gwNamespaceLabel = lib.GatewayNamespaceLabelKey
	} else if lib.UseServicesAPI() {
		gwNameLabel = lib.SvcApiGatewayNameLabelKey
		gwNamespaceLabel = lib.SvcApiGatewayNamespaceLabelKey
	}

	labels := svc.GetLabels()
	if name, ok := labels[gwNameLabel]; ok {
		if namespace, ok := labels[gwNamespaceLabel]; ok {
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

func parseSvcApiGatewayForListeners(gateway *svcapiv1alpha1.Gateway, key string) []string {
	var listeners []string
	for _, listener := range gateway.Spec.Listeners {
		gwName, nameOk := listener.Routes.Selector.MatchLabels[lib.SvcApiGatewayNameLabelKey]
		gwNamespace, nsOk := listener.Routes.Selector.MatchLabels[lib.SvcApiGatewayNamespaceLabelKey]
		if nameOk && nsOk && gwName == gateway.Name && gwNamespace == gateway.Namespace {
			listeners = append(listeners, fmt.Sprintf("%s/%d", listener.Protocol, listener.Port))
		}
	}
	return listeners
}

func validateGatewayForClass(key string, gateway *advl4v1alpha1pre1.Gateway) error {
	gwClassObj, err := lib.AKOControlConfig().AdvL4Informers().GatewayClassInformer.Lister().Get(gateway.Spec.Class)
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

func validateSvcApiGatewayForClass(key string, gateway *svcapiv1alpha1.Gateway) error {
	gwClassObj, err := lib.AKOControlConfig().SvcAPIInformers().GatewayClassInformer.Lister().Get(gateway.Spec.GatewayClassName)
	if err != nil {
		utils.AviLog.Errorf("key: %s, msg: Unable to fetch corresponding networking.x-k8s.io/gatewayclass %s %v",
			key, gateway.Spec.GatewayClassName, err)
		return err
	}

	for _, listener := range gateway.Spec.Listeners {
		gwName, nameOk := listener.Routes.Selector.MatchLabels[lib.SvcApiGatewayNameLabelKey]
		gwNamespace, nsOk := listener.Routes.Selector.MatchLabels[lib.SvcApiGatewayNamespaceLabelKey]
		if !nameOk || !nsOk ||
			(nameOk && gwName != gateway.Name) ||
			(nsOk && gwNamespace != gateway.Namespace) {
			return errors.New("Incorrect gateway matchLabels configuration")
		}
	}

	// Additional check to see if the gatewayclass is a valid avi gateway class or not.
	if gwClassObj.Spec.Controller != lib.SvcApiAviGatewayController {
		// Return an error since this is not our object.
		return errors.New("Unexpected controller")
	}

	return nil
}

func infraSettingNSToIngress(infraSettingName, key string) ([]string, bool) {
	allIngresses := make([]string, 0)
	namespaces, err := utils.GetInformers().NSInformer.Informer().GetIndexer().ByIndex(lib.AviSettingNamespaceIndex, infraSettingName)
	if err != nil {
		utils.AviLog.Errorf("key: %s, msg: Failed to fetch the namespace corresponding to the AviInfraSetting %s with error %s", key, infraSettingName, err.Error())
		return allIngresses, false
	}
	for _, ns := range namespaces {
		namespace, _ := ns.(*v1.Namespace)
		ingresses, err := utils.GetInformers().IngressInformer.Lister().Ingresses(namespace.GetName()).List(labels.Set(nil).AsSelector())
		if err != nil {
			utils.AviLog.Warnf("key: %s, msg: Failed to list Ingresses in the namespace %s", key, namespace)
			return allIngresses, false
		}
		for _, ing := range ingresses {
			key := ing.GetNamespace() + "/" + ing.GetName()
			allIngresses = append(allIngresses, key)
		}
	}
	return allIngresses, true
}

func infraSettingNSToGateway(infraSettingName, key string) ([]string, bool) {
	allGateways := make([]string, 0)
	namespaces, err := utils.GetInformers().NSInformer.Informer().GetIndexer().ByIndex(lib.AviSettingNamespaceIndex, infraSettingName)
	if err != nil {
		utils.AviLog.Errorf("key: %s, msg: Failed to fetch the namespace corresponding to the AviInfraSetting %s with error %s", key, infraSettingName, err.Error())
		return allGateways, false
	}
	for _, ns := range namespaces {
		namespace, _ := ns.(*v1.Namespace)
		gateways, err := lib.AKOControlConfig().AdvL4Informers().GatewayInformer.Lister().Gateways(namespace.GetName()).List(labels.Set(nil).AsSelector())
		if err != nil {
			utils.AviLog.Warnf("key: %s, msg: Failed to list Gateways in the namespace %s", key, namespace)
			return allGateways, false
		}
		for _, gw := range gateways {
			key := gw.GetNamespace() + "/" + gw.GetName()
			allGateways = append(allGateways, key)
		}
	}
	return allGateways, true
}

func infraSettingNSToServices(infraSettingName, key string) ([]string, bool) {
	allServices := make([]string, 0)
	namespaces, err := utils.GetInformers().NSInformer.Informer().GetIndexer().ByIndex(lib.AviSettingNamespaceIndex, infraSettingName)
	if err != nil {
		utils.AviLog.Errorf("key: %s, msg: Failed to fetch the namespace corresponding to the AviInfraSetting %s with error %s", key, infraSettingName, err.Error())
		return allServices, false
	}
	for _, ns := range namespaces {
		namespace, _ := ns.(*v1.Namespace)
		services, err := utils.GetInformers().ServiceInformer.Lister().Services(namespace.GetName()).List(labels.Set(nil).AsSelector())
		if err != nil {
			utils.AviLog.Warnf("key: %s, msg: Failed to list Services in the namespace %s", key, namespace)
			return allServices, false
		}
		for _, svc := range services {
			if svc.Spec.Type != "LoadBalancer" {
				continue
			}
			key := svc.GetNamespace() + "/" + svc.GetName()
			allServices = append(allServices, key)
		}
	}
	return allServices, true
}

func infraSettingNSToRoutes(infraSettingName, key string) ([]interface{}, bool) {
	allRoutes := make([]interface{}, 0)
	namespaces, err := utils.GetInformers().NSInformer.Informer().GetIndexer().ByIndex(lib.AviSettingNamespaceIndex, infraSettingName)
	if err != nil {
		utils.AviLog.Errorf("key: %s, msg: Failed to fetch the namespace corresponding to the AviInfraSetting %s with error %s", key, infraSettingName, err.Error())
		return allRoutes, false
	}
	for _, ns := range namespaces {
		namespace, _ := ns.(*v1.Namespace)
		routes, err := utils.GetInformers().RouteInformer.Lister().Routes(namespace.GetName()).List(labels.Set(nil).AsSelector())
		if err != nil {
			utils.AviLog.Warnf("key: %s, msg: Failed to list Routes in the namespace %s", key, namespace)
			return allRoutes, false
		}
		for _, route := range routes {
			allRoutes = append(allRoutes, route)
		}
	}
	return allRoutes, true
}
