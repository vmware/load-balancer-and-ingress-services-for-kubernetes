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
	"strings"

	"github.com/avinetworks/container-lib/utils"
	routev1 "github.com/openshift/api/route/v1"
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
			hostpath := rule.Host + svcPath.Path

			found, val := SharedHostNameLister().GetHostPathStore(hostpath)
			if found && len(val) > 0 && val[0] != nsIngress {
				// TODO: push in ako apiserver
				utils.AviLog.Warnf("key: %s, msg: Duplicate entries found for hostpath %s: %s in ingresses: %+v", key, nsIngress, hostpath, utils.Stringify(val))
			}
		}
	}
	return
}

func (v *Validator) ParseHostPathForIngress(ns string, ingName string, ingSpec v1beta1.IngressSpec, key string) IngressConfig {
	// Figure out the service names that are part of this ingress

	ingressConfig := IngressConfig{}
	hostMap := make(IngressHostMap)
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
		for _, path := range rule.IngressRuleValue.HTTP.Paths {
			hostPathMapSvc := IngressHostPathSvc{}

			hostPathMapSvc.Path = path.Path
			hostPathMapSvc.ServiceName = path.Backend.ServiceName
			hostPathMapSvc.Port = path.Backend.ServicePort.IntVal
			hostPathMapSvc.PortName = path.Backend.ServicePort.StrVal
			if hostPathMapSvc.Port == 0 {
				// Default to port 80 if not set in the ingress object
				hostPathMapSvc.Port = 80
			}
			// for ingress use 100 as default weight
			hostPathMapSvc.weight = 100
			hostPathMapSvcList = append(hostPathMapSvcList, hostPathMapSvc)
		}
		hostMap[hostName] = hostPathMapSvcList
	}

	var tlsConfigs []TlsSettings
	for _, tlsSettings := range ingSpec.TLS {
		tlsHostSvcMap := make(IngressHostMap)
		tls := TlsSettings{}
		tls.SecretName = tlsSettings.SecretName
		for _, host := range tlsSettings.Hosts {
			if !v.IsValiddHostName(host) {
				continue
			}
			hostSvcMap, ok := hostMap[host]
			if ok {
				tlsHostSvcMap[host] = hostSvcMap
				delete(hostMap, host)
			}
		}
		tls.Hosts = tlsHostSvcMap
		tlsConfigs = append(tlsConfigs, tls)
	}
	ingressConfig.TlsCollection = tlsConfigs
	ingressConfig.IngressHostMap = hostMap
	utils.AviLog.Infof("key: %s, msg: host path config from ingress: %+v", key, utils.Stringify(ingressConfig))
	return ingressConfig
}

func (v *Validator) ParseHostPathForRoute(ns string, ingName string, routeSpec routev1.RouteSpec, key string) IngressConfig {
	ingressConfig := IngressConfig{}
	hostMap := make(IngressHostMap)
	hostName := routeSpec.Host
	if !v.IsValiddHostName(hostName) {
		return ingressConfig
	}
	defaultWeight := int32(100)
	var hostPathMapSvcList []IngressHostPathSvc

	hostPathMapSvc := IngressHostPathSvc{}
	hostPathMapSvc.Path = routeSpec.Path
	hostPathMapSvc.ServiceName = routeSpec.To.Name
	hostPathMapSvc.weight = defaultWeight
	if routeSpec.To.Weight != nil {
		hostPathMapSvc.weight = *routeSpec.To.Weight
	}

	hostPathMapSvcList = append(hostPathMapSvcList, hostPathMapSvc)

	for _, backend := range routeSpec.AlternateBackends {
		hostPathMapSvc := IngressHostPathSvc{}
		hostPathMapSvc.Path = routeSpec.Path
		hostPathMapSvc.ServiceName = backend.Name
		hostPathMapSvc.weight = defaultWeight
		if routeSpec.To.Weight != backend.Weight {
			hostPathMapSvc.weight = *backend.Weight
		}
		hostPathMapSvcList = append(hostPathMapSvcList, hostPathMapSvc)
	}

	hostMap[hostName] = hostPathMapSvcList

	ingressConfig.IngressHostMap = hostMap
	utils.AviLog.Infof("key: %s, msg: host path config from ingress: %+v", key, utils.Stringify(ingressConfig))
	return ingressConfig
}
