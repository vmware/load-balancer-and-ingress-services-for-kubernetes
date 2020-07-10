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
	"ako/pkg/status"
	"fmt"
	"regexp"
	"strings"

	"github.com/avinetworks/container-lib/utils"
	"k8s.io/api/networking/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
)

var (
	Service = GraphSchema{
		Type:               "Service",
		GetParentIngresses: SvcToIng,
	}
	Ingress = GraphSchema{
		Type:               "Ingress",
		GetParentIngresses: IngressChanges,
	}
	Endpoint = GraphSchema{
		Type:               "Endpoints",
		GetParentIngresses: EPToIng,
	}
	Secret = GraphSchema{
		Type:               "Secret",
		GetParentIngresses: SecretToIng,
	}
	HostRule = GraphSchema{
		Type:               "HostRule",
		GetParentIngresses: HostRuleToIng,
	}
	HTTPRule = GraphSchema{
		Type:               "HTTPRule",
		GetParentIngresses: HTTPRuleToIng,
	}
	SupportedGraphTypes = GraphDescriptor{
		Ingress,
		Service,
		Endpoint,
		Secret,
		HostRule,
		HTTPRule,
	}
)

type GraphSchema struct {
	Type               string
	GetParentIngresses func(string, string, string) ([]string, bool)
}

type GraphDescriptor []GraphSchema

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

func HostRuleToIng(hrname string, namespace string, key string) ([]string, bool) {
	var err error
	var oldFqdn, fqdn string
	var oldFound bool
	hrDelete := false

	hostrule, err := lib.GetCRDInformers().HostRuleInformer.Lister().HostRules(namespace).Get(hrname)
	if k8serror.IsNotFound(err) {
		utils.AviLog.Debugf("key: %s, msg: HostRule Deleted\n", key)
		_, fqdn = objects.SharedCRDLister().GetHostruleToFQDNMapping(namespace + "/" + hrname)
		objects.SharedCRDLister().DeleteHostruleFQDNMapping(namespace + "/" + hrname)
		hrDelete = true
	} else if err != nil {
		utils.AviLog.Errorf("key: %s, msg: Error getting hostrule: %v\n", key, err)
		return nil, false
	} else {
		fqdn = hostrule.Spec.VirtualHost.Fqdn
		oldFound, oldFqdn = objects.SharedCRDLister().GetHostruleToFQDNMapping(namespace + "/" + hrname)
		if oldFound {
			objects.SharedCRDLister().DeleteHostruleFQDNMapping(namespace + "/" + hrname)
		}
		objects.SharedCRDLister().UpdateFQDNHostruleMapping(fqdn, namespace+"/"+hrname)
	}

	found, pathRules := objects.SharedCRDLister().GetHostHTTPRulesMapping(namespace + "/" + hrname)
	if found {
		for _, ruleNSName := range pathRules {
			rulensn := strings.Split(ruleNSName, "/")
			httprule, err := lib.GetCRDInformers().HTTPRuleInformer.Lister().HTTPRules(rulensn[0]).Get(rulensn[1])
			if err == nil {
				if hrDelete {
					status.UpdateHTTPRuleStatus(httprule, status.UpdateCRDStatusOptions{
						Status: lib.StatusRejected,
						Error:  fmt.Sprintf("hostrules.ako.k8s.io %s not found or is invalid", hrname),
					})
				} else if !hrDelete && httprule.Status.Status == lib.StatusRejected {
					status.UpdateHTTPRuleStatus(httprule, status.UpdateCRDStatusOptions{Status: lib.StatusAccepted})
				}
			}
		}
	}

	allIngresses := make([]string, 0)
	// find ingresses with host==fqdn, across all namespaces
	ok, obj := SharedHostNameLister().GetHostPathStore(fqdn)
	if !ok {
		utils.AviLog.Warnf("key: %s, msg: Couldn't find hostpath info for host: %s in cache", key, fqdn)
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
			utils.AviLog.Warnf("key: %s, msg: Couldn't find hostpath info for host: %s in cache", key, oldFqdn)
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

	return allIngresses, true
}

func HTTPRuleToIng(rrname string, namespace string, key string) ([]string, bool) {
	var err error
	httprule, err := lib.GetCRDInformers().HTTPRuleInformer.Lister().HTTPRules(namespace).Get(rrname)
	var hostrule string
	pathRules := make(map[string]string)
	var ok bool
	deleteCase := false

	if k8serror.IsNotFound(err) {
		utils.AviLog.Debugf("key: %s, msg: HTTPRule Deleted\n", key)
		_, hostrule = objects.SharedCRDLister().GetHTTPHostRuleMapping(namespace + "/" + rrname)
		_, pathRules = objects.SharedCRDLister().GetHostHTTPRulesMapping(hostrule)
		deleteCase = true
	} else if err != nil {
		utils.AviLog.Errorf("key: %s, msg: Error getting httprule: %v\n", key, err)
		return nil, false
	} else {
		utils.AviLog.Debugf("key: %s, HTTPRule %v\n", key, httprule)
		hostrule = httprule.Spec.HostRule
		found, _ := objects.SharedCRDLister().GetHostruleToFQDNMapping(hostrule)
		if !found {
			err = fmt.Errorf("hostrules.ako.k8s.io %s not found or is invalid", hostrule)
			status.UpdateHTTPRuleStatus(httprule, status.UpdateCRDStatusOptions{
				Status: lib.StatusRejected,
				Error:  err.Error(),
			})
			utils.AviLog.Errorf("key: %s, msg: %v", key, err)
			// do not return, let the httprule cache sync in for the benefit of
			// new hostrule coming after httprule
		} else {
			status.UpdateHTTPRuleStatus(httprule, status.UpdateCRDStatusOptions{Status: lib.StatusAccepted})
		}
		objects.SharedCRDLister().RemoveHostHTTPRulesMappings(namespace + "/" + rrname)

		for _, path := range httprule.Spec.Paths {
			objects.SharedCRDLister().UpdateHostHTTPRulesMappings(hostrule, path.Target, namespace+"/"+rrname)
		}
		ok, pathRules = objects.SharedCRDLister().GetHostHTTPRulesMapping(hostrule)
		if !ok {
			utils.AviLog.Warnf("key: %s, msg: Couldn't find httprules for hostrule %s in cache", key, hostrule)
		}
	}

	// find ingresses with host==fqdn, across all namespaces, and path.targetpath
	found, fqdn := objects.SharedCRDLister().GetHostruleToFQDNMapping(hostrule)
	if !found {
		utils.AviLog.Warnf("key: %s, msg: Couldn't find fqdn mapping for hostrule: %s in cache", key, hostrule)
		if deleteCase {
			objects.SharedCRDLister().RemoveHostHTTPRulesMappings(namespace + "/" + rrname)
		}
		return nil, false
	}

	// pathprefix match
	// lets say path: / and available paths registered in the cache could be keyed to /foo, /bar
	// in that case pathprefix match must account for both paths
	allIngresses := make([]string, 0)
	ok, pathIngs := SharedHostNameLister().GetHostPathStore(fqdn)
	if !ok {
		utils.AviLog.Warnf("key %s, msg: Couldn't find hostpath info for host: %s in cache", key, fqdn)
		return nil, false
	}
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

	if deleteCase {
		objects.SharedCRDLister().RemoveHostHTTPRulesMappings(namespace + "/" + rrname)
	}

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
