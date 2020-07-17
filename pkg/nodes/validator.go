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
	"strings"

	"github.com/avinetworks/container-lib/utils"
	"k8s.io/api/networking/v1beta1"
)

type Validator struct {
	subDomains []string
}

func NewNodesValidator() *Validator {
	validator := &Validator{}
	validator.subDomains = GetDefaultSubDomain()
	return validator
}

func (v *Validator) IsValiddHostName(hostname string) bool {
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

func validateSpecFromHostnameCache(key, ns, ingName string, ingSpec v1beta1.IngressSpec) {
	nsIngress := ns + "/" + ingName
	for _, rule := range ingSpec.Rules {
		for _, svcPath := range rule.IngressRuleValue.HTTP.Paths {
			found, val := SharedHostNameLister().GetHostPathStoreIngresses(rule.Host, svcPath.Path)
			if found && len(val) > 0 && utils.HasElem(val, nsIngress) && len(val) > 1 {
				// TODO: push in ako apiserver
				utils.AviLog.Warnf("key: %s, msg: Duplicate entries found for hostpath %s%s: %s in ingresses: %+v", key, nsIngress, rule.Host, svcPath.Path, utils.Stringify(val))
			}
		}
	}
	return
}

func sslKeyCertHostRulePresent(key, host string) (bool, string) {
	if lib.GetShardScheme() == "namespace" {
		return false, ""
	}

	// from host check if hostrule is present
	found, hrNSNameStr := objects.SharedCRDLister().GetFQDNToHostruleMapping(host)
	if !found {
		utils.AviLog.Debugf("key: %s, msg: Couldn't find fqdn %s to hostrule mapping in cache", key, host)
		return false, ""
	}

	hrNSName := strings.Split(hrNSNameStr, "/")
	// from hostrule check if hostrule.TLS.SSLKeyCertificate is not null
	hostRuleObj, err := lib.GetCRDInformers().HostRuleInformer.Lister().HostRules(hrNSName[0]).Get(hrNSName[1])
	if err != nil {
		utils.AviLog.Warnf("key: %s, msg: Couldn't find hostrule %s", key, hrNSNameStr)
		return false, ""
	} else if hostRuleObj.Status.Status == lib.StatusRejected {
		utils.AviLog.Warnf("key: %s, msg: rejected hostrule %s", key, hrNSNameStr)
		return false, ""
	}

	if hostRuleObj.Spec.VirtualHost.TLS.SSLKeyCertificate.Name != "" {
		utils.AviLog.Infof("key: %s, msg: secret %s found for host %s in hostrule.ako.vmware.com %s",
			key, host, hostRuleObj.Spec.VirtualHost.TLS.SSLKeyCertificate.Name, hostRuleObj.Name)
		return true, lib.DummySecret + "/" + hostRuleObj.Spec.VirtualHost.TLS.SSLKeyCertificate.Name
	}

	return false, ""
}

// ParseHostPathForIngress handling for hostrule: if the host has a hostrule, and that hostrule has a tls.sslkeycertref then
// move that host in the tls.hosts, this should be only in case of hostname sharding
func (v *Validator) ParseHostPathForIngress(ns string, ingName string, ingSpec v1beta1.IngressSpec, key string) IngressConfig {
	// Figure out the service names that are part of this ingress

	ingressConfig := IngressConfig{}
	hostMap := make(IngressHostMap)
	additionalSecureHostMap := make(IngressHostMap)
	secretHostsMap := make(map[string][]string)
	subDomains := GetDefaultSubDomain()

	for _, rule := range ingSpec.Rules {
		var hostPathMapSvcList []IngressHostPathSvc
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
			if !v.IsValiddHostName(rule.Host) {
				continue
			}
			hostName = rule.Host
		}

		if len(hostMap[hostName]) > 0 {
			hostPathMapSvcList = hostMap[hostName]
		}

		// check if this host has a valid hostrule with sslkeycertref present
		useHostRuleSSL, secretName := sslKeyCertHostRulePresent(key, hostName)
		if useHostRuleSSL && len(additionalSecureHostMap[hostName]) > 0 {
			hostPathMapSvcList = additionalSecureHostMap[hostName]
		}
		if _, ok := secretHostsMap[secretName]; !ok {
			secretHostsMap[secretName] = []string{hostName}
		} else {
			secretHostsMap[secretName] = append(secretHostsMap[secretName], hostName)
		}

		for _, path := range rule.IngressRuleValue.HTTP.Paths {
			hostPathMapSvc := IngressHostPathSvc{
				Path:        path.Path,
				ServiceName: path.Backend.ServiceName,
				Port:        path.Backend.ServicePort.IntVal,
				PortName:    path.Backend.ServicePort.StrVal,
			}
			if hostPathMapSvc.Port == 0 {
				// Default to port 80 if not set in the ingress object
				hostPathMapSvc.Port = 80
			}
			hostPathMapSvcList = append(hostPathMapSvcList, hostPathMapSvc)
		}

		if useHostRuleSSL {
			additionalSecureHostMap[hostName] = hostPathMapSvcList
		} else {
			hostMap[hostName] = hostPathMapSvcList
		}
	}

	var tlsConfigs []TlsSettings
	for _, tlsSettings := range ingSpec.TLS {
		tlsHostSvcMap := make(IngressHostMap)
		tls := TlsSettings{}
		tls.SecretName = tlsSettings.SecretName
		for _, host := range tlsSettings.Hosts {
			if _, ok := additionalSecureHostMap[host]; ok {
				continue
			}
			if !v.IsValiddHostName(host) {
				continue
			}
			hostSvcMap, ok := hostMap[host]
			if ok {
				tlsHostSvcMap[host] = hostSvcMap
				delete(hostMap, host)
			}
		}
		if len(tlsHostSvcMap) > 0 {
			tls.Hosts = tlsHostSvcMap
			tlsConfigs = append(tlsConfigs, tls)
		}
	}

	for aviSecret, securedHostNames := range secretHostsMap {
		additionalTLS := TlsSettings{}
		additionalTLS.SecretName = aviSecret
		additionalTLSHostSvcMap := make(IngressHostMap)
		for _, host := range securedHostNames {
			if hostSvcMap, ok := additionalSecureHostMap[host]; ok {
				additionalTLSHostSvcMap[host] = hostSvcMap
			}
		}
		if len(additionalTLSHostSvcMap) > 0 {
			additionalTLS.Hosts = additionalTLSHostSvcMap
			tlsConfigs = append(tlsConfigs, additionalTLS)
		}
	}

	ingressConfig.TlsCollection = tlsConfigs
	ingressConfig.IngressHostMap = hostMap
	utils.AviLog.Infof("key: %s, msg: host path config from ingress: %+v", key, utils.Stringify(ingressConfig))
	return ingressConfig
}
