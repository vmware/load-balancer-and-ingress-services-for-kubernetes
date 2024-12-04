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
	"strings"

	lib2 "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/apis/ako/v1alpha1"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/apis/ako/v1beta1"
	akov1beta1 "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/apis/ako/v1beta1"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	routev1 "github.com/openshift/api/route/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type Validator struct {
	subDomains []string
}

func NewNodesValidator() *Validator {
	validator := &Validator{}
	if !lib.IsWCP() {
		validator.subDomains = lib2.GetDefaultSubDomain()
	}
	return validator
}

func (v *Validator) IsValidHostName(hostname string) bool {
	// Check if a hostname is valid or not by verifying if it has a prefix that
	// matches any of the sub-domains.
	if v.subDomains == nil {
		// No IPAM DNS configured, we simply pass the hostname
		return true
	} else {
		for _, subd := range v.subDomains {
			if strings.HasSuffix(hostname, subd) {
				return true
			}
		}
	}
	utils.AviLog.Warnf("Didn't find match for hostname :%s Available sub-domains:%s", hostname, v.subDomains)
	return false
}

func validateSpecFromHostnameCache(key string, ingress *networkingv1.Ingress) bool {
	nsIngress := ingress.Namespace + "/" + ingress.Name
	for _, rule := range ingress.Spec.Rules {
		if rule.IngressRuleValue.HTTP != nil {
			for _, svcPath := range rule.IngressRuleValue.HTTP.Paths {
				found, val := SharedHostNameLister().GetHostPathStoreIngresses(rule.Host, svcPath.Path)
				if found && len(val) > 1 && utils.HasElem(val, nsIngress) {
					lib.AKOControlConfig().EventRecorder().Eventf(ingress, corev1.EventTypeWarning, lib.DuplicateHostPath, "Duplicate entries found for hostpath %s: %s%s in ingresses: %+v", nsIngress, rule.Host, svcPath.Path, utils.Stringify(val))
					utils.AviLog.Warnf("key: %s, msg: Duplicate entries found for hostpath %s: %s%s in ingresses: %+v", key, nsIngress, rule.Host, svcPath.Path, utils.Stringify(val))
				}
			}
			// In VCF, we use VIP per Namespace, hence same hostname should not be present across multiple namespaces
			if lib.VIPPerNamespace() {
				if ok, oldNS := SharedHostNameLister().GetNamespace(rule.Host); ok {
					if oldNS != ingress.Namespace {
						lib.AKOControlConfig().EventRecorder().Eventf(ingress, corev1.EventTypeWarning, lib.DuplicateHost, "Duplicate entries found for hostname %s in multiple namespaces %s and %s", rule.Host, oldNS, ingress.Namespace)
						utils.AviLog.Warnf("key: %s, msg: Duplicate entries found for hostname %s in multiple namespaces: %s and %s", key, rule.Host, oldNS, ingress.Namespace)
					}
				}
			}
		} else {
			utils.AviLog.Warnf("key: %s, msg: Found Ingress: %s without service backends. Not going to process.", key, ingress.Name)
			return false
		}
	}
	return true
}

func validateRouteSpecFromHostnameCache(key, ns, routeName string, routeSpec routev1.RouteSpec) {
	nsRoute := ns + "/" + routeName
	found, val := SharedHostNameLister().GetHostPathStoreIngresses(routeSpec.Host, routeSpec.Path)
	if found && len(val) > 1 && utils.HasElem(val, nsRoute) {
		utils.AviLog.Warnf("key: %s, msg: Duplicate entries found for hostpath %s%s: %s in routes: %+v", key, nsRoute, routeSpec.Host, routeSpec.Path, utils.Stringify(val))
	}
}

func findHostRuleMappingForFqdn(key, host string) (bool, *v1beta1.HostRule) {
	// from host check if hostrule is present
	found, hrNSNameStr := objects.SharedCRDLister().GetFQDNToHostruleMappingWithType(host)
	if !found {
		utils.AviLog.Debugf("key: %s, msg: Couldn't find fqdn %s to hostrule mapping in cache", key, host)
		return false, nil
	}

	hrNSName := strings.Split(hrNSNameStr, "/")
	// from hostrule check if hostrule.TLS.SSLKeyCertificate is not null
	hostRuleObj, err := lib.AKOControlConfig().CRDInformers().HostRuleInformer.Lister().HostRules(hrNSName[0]).Get(hrNSName[1])
	if err != nil {
		utils.AviLog.Warnf("key: %s, msg: Couldn't find hostrule %s: %v", key, hrNSNameStr, err)
		return false, nil
	} else if hostRuleObj.Status.Status == lib.StatusRejected {
		utils.AviLog.Warnf("key: %s, msg: rejected hostrule %s", key, hrNSNameStr)
		return false, nil
	} else {
		return true, hostRuleObj
	}
}

func sslKeyCertHostRulePresent(hostRuleObj *v1beta1.HostRule, key string) (bool, []string) {
	var sslKeyCerts []string
	if hostRuleObj.Spec.VirtualHost.TLS.SSLKeyCertificate.Name != "" {
		utils.AviLog.Infof("key: %s, msg: secret %s found for host %s in hostrule.ako.vmware.com %s",
			key, hostRuleObj.Spec.VirtualHost.TLS.SSLKeyCertificate.Name, hostRuleObj.Spec.VirtualHost.Fqdn, hostRuleObj.Name)
		if hostRuleObj.Spec.VirtualHost.TLS.SSLKeyCertificate.Type == akov1beta1.HostRuleSecretTypeSecretReference {
			sslKeyCerts = append(sslKeyCerts, lib.DummySecretK8s+"/"+hostRuleObj.Namespace+"/"+hostRuleObj.Spec.VirtualHost.TLS.SSLKeyCertificate.Name)
		} else if hostRuleObj.Spec.VirtualHost.TLS.SSLKeyCertificate.Type == akov1beta1.HostRuleSecretTypeAviReference {
			sslKeyCerts = append(sslKeyCerts, lib.DummySecret+"/"+hostRuleObj.Spec.VirtualHost.TLS.SSLKeyCertificate.Name)
		}
	}
	if hostRuleObj.Spec.VirtualHost.TLS.SSLKeyCertificate.AlternateCertificate.Name != "" {
		utils.AviLog.Infof("key: %s, msg: alternate secret %s found for host %s in hostrule.ako.vmware.com %s",
			key, hostRuleObj.Spec.VirtualHost.TLS.SSLKeyCertificate.AlternateCertificate.Name, hostRuleObj.Spec.VirtualHost.Fqdn, hostRuleObj.Name)
		if hostRuleObj.Spec.VirtualHost.TLS.SSLKeyCertificate.AlternateCertificate.Type == akov1beta1.HostRuleSecretTypeSecretReference {
			sslKeyCerts = append(sslKeyCerts, lib.DummySecretK8s+"/"+hostRuleObj.Namespace+"/"+hostRuleObj.Spec.VirtualHost.TLS.SSLKeyCertificate.AlternateCertificate.Name)
		} else if hostRuleObj.Spec.VirtualHost.TLS.SSLKeyCertificate.AlternateCertificate.Type == akov1beta1.HostRuleSecretTypeAviReference {
			sslKeyCerts = append(sslKeyCerts, lib.DummySecret+"/"+hostRuleObj.Spec.VirtualHost.TLS.SSLKeyCertificate.AlternateCertificate.Name)
		}
	}
	if len(sslKeyCerts) != 0 {
		return true, sslKeyCerts
	}
	return false, sslKeyCerts
}

func getGslbFqdnFromHostRule(hostRuleObj *v1beta1.HostRule) (bool, string) {
	if hostRuleObj.Spec.VirtualHost.Gslb.Fqdn != "" {
		return true, hostRuleObj.Spec.VirtualHost.Gslb.Fqdn
	}
	return false, ""
}

func destinationCAHTTPRulePresent(key, host, path string) (bool, string) {
	// from host check if httprule is present
	found, pathRules := objects.SharedCRDLister().GetFqdnHTTPRulesMapping(host)
	if !found {
		utils.AviLog.Debugf("key: %s, msg: Couldn't find fqdn %s to httprule mapping in cache", key, host)
		return false, ""
	}

	rule, isValid := pathRules[path]
	if !isValid {
		utils.AviLog.Debugf("key: %s, msg: Couldn't find path %s to httprule mapping in cache", key, path)
		return false, ""
	}

	ruleNSName := strings.Split(rule, "/")
	// from hostrule check if hostrule.TLS.SSLKeyCertificate is not null
	httpRuleObj, err := lib.AKOControlConfig().CRDInformers().HTTPRuleInformer.Lister().HTTPRules(ruleNSName[0]).Get(ruleNSName[1])
	if err != nil {
		utils.AviLog.Warnf("key: %s, msg: Couldn't find httprule %s: %v", key, rule, err)
		return false, ""
	} else if httpRuleObj.Status.Status == lib.StatusRejected {
		utils.AviLog.Warnf("key: %s, msg: rejected httprule %s", key, rule)
		return false, ""
	}

	for _, rulePath := range httpRuleObj.Spec.Paths {
		if rulePath.Target == path && rulePath.TLS.DestinationCA != "" {
			utils.AviLog.Infof("key: %s, msg: destinationCA found for hostpath %s %s in httprule.ako.vmware.com %s",
				key, host, rulePath.Target, httpRuleObj.Name)
			return true, rulePath.TLS.DestinationCA
		}
	}

	return false, ""
}

// ParseHostPathForIngress handling for hostrule: if the host has a hostrule, and that hostrule has a tls.sslkeycertref then
// move that host in the tls.hosts, this should be only in case of hostname sharding
func (v *Validator) ParseHostPathForIngress(ns string, ingName string, ingSpec networkingv1.IngressSpec, annotations map[string]string, key string) IngressConfig {
	// Figure out the service names that are part of this ingress

	ingressConfig := IngressConfig{}
	hostMap := make(IngressHostMap)
	additionalSecureHostMap := make(IngressHostMap)
	secretHostsMap := make(map[string][]string)
	subDomains := lib2.GetDefaultSubDomain()

	var useDefaultSecret bool
	if val, found := annotations[lib.DefaultSecretEnabled]; found {
		useDefaultSecret = strings.EqualFold(val, "true")
	}

	var passthroughEnabled bool
	pass := PassthroughSettings{}
	passConfig := make(map[string]PassthroughSettings)
	if val, found := annotations[lib.PassthroughAnnotation]; found {
		passthroughEnabled = strings.EqualFold(val, "true")
	}

	var tlsConfigs []TlsSettings
	for _, rule := range ingSpec.Rules {
		var hostPathMapSvcList HostMetadata
		var hostName string
		if rule.Host == "" {
			if subDomains == nil {
				utils.AviLog.Warnf("No sub-domain configured in cloud")
				continue
			} else {
				// The Host field is empty. Generate a hostName using the sub-domain info
				if strings.HasPrefix(subDomains[0], ".") {
					hostName = ingName + "." + ns + subDomains[0]
				} else {
					hostName = ingName + "." + ns + "." + subDomains[0]
				}
			}
		} else {
			if !v.IsValidHostName(rule.Host) {
				continue
			}
			hostName = rule.Host
		}

		if len(hostMap[hostName].ingressHPSvc) > 0 {
			hostPathMapSvcList = hostMap[hostName]
		}
		foundHR, hrObj := findHostRuleMappingForFqdn(key, hostName)
		if foundHR {
			// Fetch the GSLB FQDN and update the hostmap
			foundGs, gslbFqdn := getGslbFqdnFromHostRule(hrObj)
			if foundGs {
				hostPathMapSvcList.gslbHostHeader = gslbFqdn
			}
		}
		// check if this host has a valid hostrule with sslkeycertref present
		var useHostRuleSSL bool
		var secretNames []string
		if foundHR {
			useHostRuleSSL, secretNames = sslKeyCertHostRulePresent(hrObj, key)
		}
		if useHostRuleSSL && len(additionalSecureHostMap[hostName].ingressHPSvc) > 0 {
			hostPathMapSvcList = additionalSecureHostMap[hostName]
		}
		for _, secretName := range secretNames {
			if _, ok := secretHostsMap[secretName]; !ok {
				secretHostsMap[secretName] = []string{hostName}
			} else {
				secretHostsMap[secretName] = append(secretHostsMap[secretName], hostName)
			}
		}
		if rule.IngressRuleValue.HTTP != nil {
			for _, path := range rule.IngressRuleValue.HTTP.Paths {
				pathType := networkingv1.PathTypeImplementationSpecific
				if path.PathType != nil {
					pathType = *path.PathType
				}
				hostPathMapSvc := IngressHostPathSvc{
					Path:        path.Path,
					PathType:    pathType,
					ServiceName: path.Backend.Service.Name,
					Port:        path.Backend.Service.Port.Number,
					PortName:    path.Backend.Service.Port.Name,
					TargetPort:  v.findTargetPort(path.Backend.Service.Name, ns, &path.Backend.Service.Port, key),
				}
				if hostPathMapSvc.PortName == "" {
					// fill the port name as the port name is not given in the ingress
					hostPathMapSvc.PortName = v.findPortName(path.Backend.Service.Name, ns, path.Backend.Service.Port.Number, key)
				}
				if hostPathMapSvc.Port == 0 {
					// Default to port 80 if not set in the ingress object
					hostPathMapSvc.Port = 80
				}
				// for ingress use 100 as default weight
				hostPathMapSvc.weight = 100
				hostPathMapSvcList.ingressHPSvc = append(hostPathMapSvcList.ingressHPSvc, hostPathMapSvc)
			}
		}

		if passthroughEnabled {
			pass.host = hostName
			pass.PathSvc = hostPathMapSvcList.ingressHPSvc
			// For secure ingress redirect is enabled, hence enabling this for passthrough ingresses too
			pass.redirect = true
			passConfig[hostName] = pass
		} else if useHostRuleSSL {
			additionalSecureHostMap[hostName] = hostPathMapSvcList
		} else if useDefaultSecret {
			defaultTLSHostSvcMap := make(IngressHostMap)
			defaultTLSHostSvcMap[hostName] = hostPathMapSvcList
			defaultTLS := TlsSettings{
				SecretName: lib.GetDefaultSecretForRoutes(),
				SecretNS:   utils.GetAKONamespace(),
				Hosts:      defaultTLSHostSvcMap,
				redirect:   true,
			}

			tlsConfigs = append(tlsConfigs, defaultTLS)
			if ok, _ := objects.SharedSvcLister().IngressMappings(ns).GetIngToSecret(ingName); !ok {
				akoNS := utils.GetAKONamespace()
				objects.SharedSvcLister().IngressMappings(ns).AddIngressToSecretsMappings(akoNS, ingName, defaultTLS.SecretName)
				objects.SharedSvcLister().IngressMappings(akoNS).AddSecretsToIngressMappings(ns, ingName, defaultTLS.SecretName)
			}
		} else {
			hostMap[hostName] = hostPathMapSvcList
		}
	}

	if passthroughEnabled {
		ingressConfig.PassthroughCollection = passConfig
		utils.AviLog.Infof("key: %s, msg: host path config from passthrough enabled ingress: %+v", key, utils.Stringify(ingressConfig))
		return ingressConfig
	}

	for _, tlsSettings := range ingSpec.TLS {
		tlsHostSvcMap := make(IngressHostMap)
		tls := TlsSettings{}
		tls.SecretName = tlsSettings.SecretName
		tls.SecretNS = ns
		for _, host := range tlsSettings.Hosts {
			if _, ok := additionalSecureHostMap[host]; ok {
				continue
			}
			if !v.IsValidHostName(host) {
				continue
			}
			hostSvcMap, ok := hostMap[host]
			if ok {
				tlsHostSvcMap[host] = hostSvcMap
				delete(hostMap, host)
			}
		}
		tls.Hosts = tlsHostSvcMap
		// Always add http -> https redirect rule for secure ingress
		tls.redirect = true
		tlsConfigs = append(tlsConfigs, tls)
		// If svc for an ingress gets processed before the ingress itself,
		// then secret mapping may not be updated, update it here.
		if ok, _ := objects.SharedSvcLister().IngressMappings(ns).GetIngToSecret(ingName); !ok {
			objects.SharedSvcLister().IngressMappings(ns).AddIngressToSecretsMappings(ns, ingName, tlsSettings.SecretName)
			objects.SharedSvcLister().IngressMappings(ns).AddSecretsToIngressMappings(ns, ingName, tlsSettings.SecretName)
		}
	}

	for aviSecret, securedHostNames := range secretHostsMap {
		isCertRef := false
		if lib.IsSecretAviCertRef(aviSecret) {
			isCertRef = true
		}

		additionalTLSHostSvcMap := make(IngressHostMap)
		for _, host := range securedHostNames {
			if hostSvcMap, ok := additionalSecureHostMap[host]; ok {
				additionalTLSHostSvcMap[host] = hostSvcMap
			}
		}

		var additionalTLS TlsSettings
		if !isCertRef {
			if len(additionalTLSHostSvcMap) > 0 {
				secretNS := strings.Split(aviSecret, "/")[1]
				additionalTLS = TlsSettings{
					SecretName: aviSecret,
					SecretNS:   secretNS,
					Hosts:      additionalTLSHostSvcMap,
					redirect:   true,
				}
				if ok, _ := objects.SharedSvcLister().IngressMappings(ns).GetIngToSecret(ingName); !ok {
					objects.SharedSvcLister().IngressMappings(ns).AddIngressToSecretsMappings(secretNS, ingName, additionalTLS.SecretName)
					objects.SharedSvcLister().IngressMappings(secretNS).AddSecretsToIngressMappings(ns, ingName, additionalTLS.SecretName)
				}
			}
		} else {
			if len(additionalTLSHostSvcMap) > 0 {
				// Always add http -> https redirect rule for secure ingress
				// for sni VS created using hostrule
				additionalTLS = TlsSettings{
					SecretName: aviSecret,
					Hosts:      additionalTLSHostSvcMap,
					redirect:   true,
				}
			}
		}
		tlsConfigs = append(tlsConfigs, additionalTLS)
	}

	ingressConfig.TlsCollection = tlsConfigs
	ingressConfig.IngressHostMap = hostMap
	utils.AviLog.Infof("key: %s, msg: host path config from ingress: %+v", key, utils.Stringify(ingressConfig))
	return ingressConfig
}

func (v *Validator) findTargetPort(serviceName, ns string, serviceBackendPort *networkingv1.ServiceBackendPort, key string) intstr.IntOrString {
	// Query the service and obtain the targetPort
	svcObj, err := utils.GetInformers().ServiceInformer.Lister().Services(ns).Get(serviceName)
	if err != nil {
		utils.AviLog.Warnf("key: %s, msg: error while fetching service object: %s", key, err)
		return intstr.IntOrString{}
	}
	if svcObj.Spec.Type == "NodePort" {
		// Service of type NodePorts are not supported with tagertPort info. In such a case, the ports in the ingress must be strings
		return intstr.IntOrString{}
	}
	for _, port := range svcObj.Spec.Ports {
		// Iterate the ports and find the match for targetPort
		if serviceBackendPort.Number == port.Port ||
			serviceBackendPort.Name == port.Name {
			utils.AviLog.Infof("key: %s, msg: Found targetPort %v for Port: %v", key, port.TargetPort.String(), serviceBackendPort)
			return port.TargetPort
		}
	}
	return intstr.IntOrString{}
}

func (v *Validator) findPortName(serviceName, ns string, servicePort int32, key string) string {
	// Query the service and obtain the port name
	svcObj, err := utils.GetInformers().ServiceInformer.Lister().Services(ns).Get(serviceName)
	if err != nil {
		utils.AviLog.Warnf("key: %s, msg: error while fetching service object: %s", key, err)
		return ""
	}
	for _, port := range svcObj.Spec.Ports {
		// Iterate the ports and find the match for targetPort
		if servicePort == port.Port {
			utils.AviLog.Debugf("key: %s, msg: Found port name %s for Port: %v", key, port.Name, servicePort)
			return port.Name
		}
	}
	utils.AviLog.Warnf("key: %s, msg: Port name not found in service obj: %v", key, svcObj)
	return ""
}

func (v *Validator) ParseHostPathForRoute(ns string, routeName string, routeSpec routev1.RouteSpec, key string) IngressConfig {
	ingressConfig := IngressConfig{}
	hostMap := make(IngressHostMap)
	hostName := routeSpec.Host

	if hostName == "" {
		hostName = lib.GetHostnameforSubdomain(routeSpec.Subdomain)
	}

	if !v.IsValidHostName(hostName) {
		return ingressConfig
	}

	defaultWeight := uint32(100)
	var hostPathMapSvcList HostMetadata

	hostPathMapSvc := IngressHostPathSvc{}
	hostPathMapSvc.Path = routeSpec.Path
	hostPathMapSvc.ServiceName = routeSpec.To.Name
	hostPathMapSvc.weight = defaultWeight
	if routeSpec.To.Weight != nil {
		hostPathMapSvc.weight = uint32(*routeSpec.To.Weight)
	}

	if routeSpec.Port != nil {
		if routeSpec.Port.TargetPort.Type == intstr.Int {
			hostPathMapSvc.TargetPort = routeSpec.Port.TargetPort
		} else if routeSpec.Port.TargetPort.Type == intstr.String {
			hostPathMapSvc.PortName = routeSpec.Port.TargetPort.StrVal
		}
	} else {
		utils.AviLog.Infof("key: %s, msg: no port specified for route, all ports would be used", key)
	}

	hostPathMapSvcList.ingressHPSvc = append(hostPathMapSvcList.ingressHPSvc, hostPathMapSvc)

	for _, backend := range routeSpec.AlternateBackends {
		hostPathMapSvc := IngressHostPathSvc{}
		hostPathMapSvc.Path = routeSpec.Path
		hostPathMapSvc.ServiceName = backend.Name
		hostPathMapSvc.weight = defaultWeight
		if backend.Weight != nil {
			hostPathMapSvc.weight = uint32(*backend.Weight)
		}
		hostPathMapSvcList.ingressHPSvc = append(hostPathMapSvcList.ingressHPSvc, hostPathMapSvc)
	}

	var tlsConfigs []TlsSettings

	// check if this host has a valid hostrule with sslkeycertref present
	var useHostRuleSSL bool
	var secretNames []string
	foundHR, hrObj := findHostRuleMappingForFqdn(key, hostName)
	if foundHR {
		// Fetch the GSLB FQDN and update the hostmap
		foundGs, gslbFqdn := getGslbFqdnFromHostRule(hrObj)
		if foundGs {
			hostPathMapSvcList.gslbHostHeader = gslbFqdn
		}
	}
	hostMap[hostName] = hostPathMapSvcList

	if foundHR {
		useHostRuleSSL, secretNames = sslKeyCertHostRulePresent(hrObj, key)
	}
	if routeSpec.TLS != nil && !useHostRuleSSL {
		secretNames = []string{lib.RouteSecretsPrefix + routeName}
		if routeSpec.TLS.Certificate == "" || routeSpec.TLS.Key == "" {
			secretNames = []string{lib.GetDefaultSecretForRoutes()}
		}
	}

	if routeSpec.TLS != nil && routeSpec.TLS.Termination == routev1.TLSTerminationPassthrough {
		pass := PassthroughSettings{
			host:    hostName,
			PathSvc: hostPathMapSvcList.ingressHPSvc,
		}
		if routeSpec.TLS.InsecureEdgeTerminationPolicy == routev1.InsecureEdgeTerminationPolicyRedirect {
			pass.redirect = true
		}
		passConfig := make(map[string]PassthroughSettings)
		passConfig[hostName] = pass
		ingressConfig.PassthroughCollection = passConfig
	} else if len(secretNames) != 0 {
		for _, secretName := range secretNames {
			tls := TlsSettings{
				Hosts:      hostMap,
				SecretName: secretName,
			}

			isCertRef := false
			if lib.IsSecretAviCertRef(secretName) {
				isCertRef = true
			}
			if useHostRuleSSL {
				var additionalTLS TlsSettings
				if !isCertRef {
					secretNS := strings.Split(secretName, "/")[1]
					additionalTLS = TlsSettings{
						SecretName: secretName,
						SecretNS:   secretNS,
						Hosts:      hostMap,
					}
					if ok, _ := objects.SharedSvcLister().IngressMappings(ns).GetIngToSecret(routeName); !ok {
						objects.SharedSvcLister().IngressMappings(ns).AddIngressToSecretsMappings(secretNS, routeName, additionalTLS.SecretName)
						objects.SharedSvcLister().IngressMappings(secretNS).AddSecretsToIngressMappings(ns, routeName, additionalTLS.SecretName)
					}
				} else {
					// Always add http -> https redirect rule for secure ingress
					// for sni VS created using hostrule
					additionalTLS = TlsSettings{
						SecretName: secretName,
						Hosts:      hostMap,
					}
				}
				tlsConfigs = append(tlsConfigs, additionalTLS)
			} else if routeSpec.TLS != nil {
				// build edge cert data for termination: edge and reencrypt
				if routeSpec.TLS.Termination == routev1.TLSTerminationEdge ||
					routeSpec.TLS.Termination == routev1.TLSTerminationReencrypt {
					if routeSpec.TLS.Certificate == "" || routeSpec.TLS.Key == "" {
						tls.SecretName = secretName
						tls.SecretNS = utils.GetAKONamespace()
					} else {
						tls.cert = routeSpec.TLS.Certificate
						tls.key = routeSpec.TLS.Key
					}
					tls.cacert = routeSpec.TLS.CACertificate
					if routeSpec.TLS.InsecureEdgeTerminationPolicy == routev1.InsecureEdgeTerminationPolicyRedirect {
						tls.redirect = true
					} else if routeSpec.TLS.InsecureEdgeTerminationPolicy == routev1.InsecureEdgeTerminationPolicyNone {
						tls.blockHTTPTraffic = true
					}
				}

				// reencrypt specific
				if routeSpec.TLS.Termination == routev1.TLSTerminationReencrypt {
					tls.reencrypt = true
					if routeSpec.TLS.DestinationCACertificate != "" {
						tls.destCA = routeSpec.TLS.DestinationCACertificate
					}

					// overwrite with httprule
					if useHttpRuleCA, caCert := destinationCAHTTPRulePresent(key, hostName, routeSpec.Path); useHttpRuleCA {
						tls.destCA = caCert
					}
				}
				tlsConfigs = append(tlsConfigs, tls)
			}
			if !useHostRuleSSL {
				// If svc for a route gets processed before the route itself,
				// then secret mapping may not be updated, update it here.
				if ok, _ := objects.OshiftRouteSvcLister().IngressMappings(ns).GetIngToSecret(routeName); !ok {
					akoNS := utils.GetAKONamespace()
					objects.OshiftRouteSvcLister().IngressMappings(ns).AddIngressToSecretsMappings(akoNS, routeName, secretName)
					objects.OshiftRouteSvcLister().IngressMappings(akoNS).AddSecretsToIngressMappings(ns, routeName, secretName)
				}
			}
		}
		ingressConfig.TlsCollection = tlsConfigs
	}

	if len(secretNames) == 0 || (routeSpec.TLS != nil && routeSpec.TLS.InsecureEdgeTerminationPolicy == routev1.InsecureEdgeTerminationPolicyAllow) {
		ingressConfig.IngressHostMap = hostMap
		if routeSpec.TLS != nil && routeSpec.TLS.InsecureEdgeTerminationPolicy == routev1.InsecureEdgeTerminationPolicyAllow {
			ingressConfig.InsecureEdgeTermAllow = true
		}
	}

	utils.AviLog.Infof("key: %s, msg: host path config from routes: %+v", key, utils.Stringify(ingressConfig))
	return ingressConfig
}

// ParseHostPathForMultiClusterIngress extracts the information from multi-cluster ingress object and generates ingress configs required for creating the models.
func (v *Validator) ParseHostPathForMultiClusterIngress(ns string, ingName string, ingSpec *v1alpha1.MultiClusterIngressSpec, key string) IngressConfig {
	// Figure out the service names that are part of this ingress

	hostname := ingSpec.Hostname
	secretName := ingSpec.SecretName

	var hostPathMapSvcList HostMetadata
	for _, config := range ingSpec.Config {
		ingressHPSvc := IngressHostPathSvc{
			ServiceName:    config.Service.Name,
			Path:           config.Path,
			PathType:       networkingv1.PathTypeImplementationSpecific,
			Port:           int32(config.Service.Port),
			weight:         uint32(config.Weight),
			clusterContext: config.ClusterContext,
			svcNamespace:   config.Service.Namespace,
		}
		hostPathMapSvcList.ingressHPSvc = append(hostPathMapSvcList.ingressHPSvc, ingressHPSvc)
	}
	hostMap := make(IngressHostMap, 1)
	hostMap[hostname] = hostPathMapSvcList
	var tlsConfigs []TlsSettings
	if secretName != "" {
		tls := TlsSettings{
			SecretName: secretName,
			SecretNS:   ns,
			key:        key,
			Hosts:      hostMap,
			// Always add http -> https redirect rule for secure ingress
			redirect: true,
		}
		tlsConfigs = append(tlsConfigs, tls)
	}

	// If svc for an multi-cluster ingress gets processed before the ingress itself,
	// then secret mapping may not be updated, update it here.
	if ok, _ := objects.SharedSvcLister().IngressMappings(ns).GetIngToSecret(ingName); !ok {
		objects.SharedSvcLister().IngressMappings(ns).AddIngressToSecretsMappings(ns, ingName, secretName)
		objects.SharedSvcLister().IngressMappings(ns).AddSecretsToIngressMappings(ns, ingName, secretName)
	}

	// TODO: default secret, passthrough, additionalTLS config, subdomain and target port

	ingressConfig := IngressConfig{}
	ingressConfig.TlsCollection = tlsConfigs
	ingressConfig.IngressHostMap = hostMap
	utils.AviLog.Infof("key: %s, msg: host path config from multi-cluster ingress: %+v", key, utils.Stringify(ingressConfig))
	return ingressConfig
}

func getNamespaceAviInfraSetting(key, ns string) (*v1beta1.AviInfraSetting, error) {
	namespace, err := utils.GetInformers().NSInformer.Lister().Get(ns)
	if err != nil {
		return nil, err
	}
	infraSettingCRName, ok := namespace.GetAnnotations()[lib.InfraSettingNameAnnotation]
	if !ok {
		return nil, nil
	}
	infraSetting, err := lib.AKOControlConfig().CRDInformers().AviInfraSettingInformer.Lister().Get(infraSettingCRName)
	if err != nil {
		return nil, err
	}
	if infraSetting != nil && infraSetting.Status.Status != lib.StatusAccepted {
		utils.AviLog.Warnf("key: %s, msg: Referred AviInfraSetting %s is invalid", key, infraSetting.Name)
		return nil, fmt.Errorf("AviInfraSetting %s is invalid", infraSetting.Name)
	}
	return infraSetting, nil
}
