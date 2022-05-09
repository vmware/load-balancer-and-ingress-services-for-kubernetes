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

	"github.com/vmware/alb-sdk/go/models"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"
	akov1alpha1 "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/apis/ako/v1alpha1"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
)

func BuildL7HostRule(host, key string, vsNode AviVsEvhSniModel) {
	// use host to find out HostRule CRD if it exists
	// The host that comes here will have a proper FQDN, either from the Ingress/Route (foo.com)
	// or the SharedVS FQDN (Shared-L7-1.com).
	found, hrNamespaceName := objects.SharedCRDLister().GetFQDNToHostruleMappingWithType(host)
	deleteCase := false
	if !found {
		utils.AviLog.Debugf("key: %s, msg: No HostRule found for virtualhost: %s in Cache", key, host)
		deleteCase = true
	}

	var err error
	var hrNSName []string
	var hostrule *akov1alpha1.HostRule
	if !deleteCase {
		hrNSName = strings.Split(hrNamespaceName, "/")
		hostrule, err = lib.AKOControlConfig().CRDInformers().HostRuleInformer.Lister().HostRules(hrNSName[0]).Get(hrNSName[1])
		if err != nil {
			utils.AviLog.Debugf("key: %s, msg: No HostRule found for virtualhost: %s msg: %v", key, host, err)
			deleteCase = true
		} else if hostrule.Status.Status == lib.StatusRejected {
			// do not apply a rejected hostrule, this way the VS would retain
			return
		}
	}

	// host specific
	var vsWafPolicy, vsAppProfile, vsErrorPageProfile, vsAnalyticsProfile, vsSslProfile, lbIP string
	var vsSslKeyCertificates []string
	var vsEnabled *bool
	var crdStatus lib.CRDMetadata

	// Initializing the values of vsHTTPPolicySets and vsDatascripts, using a nil value would impact the value of VS checksum
	vsHTTPPolicySets := []string{}
	vsDatascripts := []string{}
	var analyticsPolicy *models.AnalyticsPolicy

	// Get the existing VH domain names and then manipulate it based on the aliases in Hostrule CRD.
	VHDomainNames := vsNode.GetVHDomainNames()

	portProtocols := []AviPortHostProtocol{
		{Port: 80, Protocol: utils.HTTP},
		{Port: 443, Protocol: utils.HTTP, EnableSSL: true},
	}

	if !deleteCase {
		if hostrule.Spec.VirtualHost.TLS.SSLKeyCertificate.Type == akov1alpha1.HostRuleSecretTypeAviReference &&
			hostrule.Spec.VirtualHost.TLS.SSLKeyCertificate.Name != "" {
			vsSslKeyCertificates = append(vsSslKeyCertificates, fmt.Sprintf("/api/sslkeyandcertificate?name=%s", hostrule.Spec.VirtualHost.TLS.SSLKeyCertificate.Name))
			vsNode.SetSSLKeyCertRefs([]*AviTLSKeyCertNode{})
		}

		if hostrule.Spec.VirtualHost.TLS.SSLKeyCertificate.AlternateCertificate.Type == akov1alpha1.HostRuleSecretTypeAviReference &&
			hostrule.Spec.VirtualHost.TLS.SSLKeyCertificate.AlternateCertificate.Name != "" {
			vsSslKeyCertificates = append(vsSslKeyCertificates, fmt.Sprintf("/api/sslkeyandcertificate?name=%s", hostrule.Spec.VirtualHost.TLS.SSLKeyCertificate.AlternateCertificate.Name))
			vsNode.SetSSLKeyCertRefs([]*AviTLSKeyCertNode{})
		}

		if hostrule.Spec.VirtualHost.TLS.SSLProfile != "" {
			vsSslProfile = fmt.Sprintf("/api/sslprofile?name=%s", hostrule.Spec.VirtualHost.TLS.SSLProfile)
		}

		if hostrule.Spec.VirtualHost.WAFPolicy != "" {
			vsWafPolicy = fmt.Sprintf("/api/wafpolicy?name=%s", hostrule.Spec.VirtualHost.WAFPolicy)
		}

		if hostrule.Spec.VirtualHost.ApplicationProfile != "" {
			vsAppProfile = fmt.Sprintf("/api/applicationprofile?name=%s", hostrule.Spec.VirtualHost.ApplicationProfile)
		}

		if hostrule.Spec.VirtualHost.ErrorPageProfile != "" {
			vsErrorPageProfile = fmt.Sprintf("/api/errorpageprofile?name=%s", hostrule.Spec.VirtualHost.ErrorPageProfile)
		}

		if hostrule.Spec.VirtualHost.AnalyticsProfile != "" {
			vsAnalyticsProfile = fmt.Sprintf("/api/analyticsprofile?name=%s", hostrule.Spec.VirtualHost.AnalyticsProfile)
		}

		for _, policy := range hostrule.Spec.VirtualHost.HTTPPolicy.PolicySets {
			if !utils.HasElem(vsHTTPPolicySets, fmt.Sprintf("/api/httppolicyset?name=%s", policy)) {
				vsHTTPPolicySets = append(vsHTTPPolicySets, fmt.Sprintf("/api/httppolicyset?name=%s", policy))
			}
		}

		// delete all auto-created HttpPolicySets by AKO if override is set
		if hostrule.Spec.VirtualHost.HTTPPolicy.Overwrite {
			vsNode.SetHttpPolicyRefs([]*AviHttpPolicySetNode{})
		}

		for _, script := range hostrule.Spec.VirtualHost.Datascripts {
			if !utils.HasElem(vsDatascripts, fmt.Sprintf("/api/vsdatascriptset?name=%s", script)) {
				vsDatascripts = append(vsDatascripts, fmt.Sprintf("/api/vsdatascriptset?name=%s", script))
			}
		}

		if hostrule.Spec.VirtualHost.TCPSettings != nil {
			if vsNode.IsSharedVS() || vsNode.IsDedicatedVS() {
				portProtocols = []AviPortHostProtocol{}
				for _, listener := range hostrule.Spec.VirtualHost.TCPSettings.Listeners {
					portProtocol := AviPortHostProtocol{
						Port:     int32(listener.Port),
						Protocol: utils.HTTP,
					}
					if listener.EnableSSL {
						portProtocol.EnableSSL = listener.EnableSSL
					}
					portProtocols = append(portProtocols, portProtocol)
				}

				// L7 StaticIP
				if hostrule.Spec.VirtualHost.TCPSettings.LoadBalancerIP != "" {
					lbIP = hostrule.Spec.VirtualHost.TCPSettings.LoadBalancerIP
				}
			}
		}

		vsEnabled = hostrule.Spec.VirtualHost.EnableVirtualHost
		crdStatus = lib.CRDMetadata{
			Type:   "HostRule",
			Value:  hostrule.Namespace + "/" + hostrule.Name,
			Status: lib.CRDActive,
		}

		if hostrule.Spec.VirtualHost.AnalyticsPolicy != nil {
			var infinite int32 = 0 // Special value to set log duration as infinite
			analyticsPolicy = &models.AnalyticsPolicy{
				FullClientLogs: &models.FullClientLogs{
					Duration: &infinite,
					Enabled:  hostrule.Spec.VirtualHost.AnalyticsPolicy.FullClientLogs.Enabled,
					Throttle: lib.GetThrottle(hostrule.Spec.VirtualHost.AnalyticsPolicy.FullClientLogs.Throttle),
				},
				AllHeaders: hostrule.Spec.VirtualHost.AnalyticsPolicy.LogAllHeaders,
			}
		}

		for _, alias := range hostrule.Spec.VirtualHost.Aliases {
			if !utils.HasElem(VHDomainNames, alias) {
				VHDomainNames = append(VHDomainNames, alias)
			}
		}

		utils.AviLog.Infof("key: %s, Successfully attached hostrule %s on vsNode %s", key, hrNamespaceName, vsNode.GetName())
	} else {
		if vsNode.GetServiceMetadata().CRDStatus.Value != "" {
			crdStatus = vsNode.GetServiceMetadata().CRDStatus
			crdStatus.Status = lib.CRDInactive
		}
		if hrNamespaceName != "" {
			utils.AviLog.Infof("key: %s, Successfully detached hostrule %s from vsNode %s", key, hrNamespaceName, vsNode.GetName())
		}
	}

	vsNode.SetSSLKeyCertAviRef(vsSslKeyCertificates)
	vsNode.SetWafPolicyRef(vsWafPolicy)
	vsNode.SetHttpPolicySetRefs(vsHTTPPolicySets)
	vsNode.SetAppProfileRef(vsAppProfile)
	vsNode.SetAnalyticsProfileRef(vsAnalyticsProfile)
	vsNode.SetErrorPageProfileRef(vsErrorPageProfile)
	vsNode.SetSSLProfileRef(vsSslProfile)
	vsNode.SetVsDatascriptRefs(vsDatascripts)
	vsNode.SetEnabled(vsEnabled)
	vsNode.SetAnalyticsPolicy(analyticsPolicy)
	vsNode.SetPortProtocols(portProtocols)
	vsNode.SetVSVIPLoadBalancerIP(lbIP)
	vsNode.SetVHDomainNames(VHDomainNames)

	serviceMetadataObj := vsNode.GetServiceMetadata()
	serviceMetadataObj.CRDStatus = crdStatus
	vsNode.SetServiceMetadata(serviceMetadataObj)
}

// BuildPoolHTTPRule notes
// when we get an ingress update and we are building the corresponding pools of that ingress
// we need to get all httprules which match ingress's host/path
func BuildPoolHTTPRule(host, poolPath, ingName, namespace, infraSettingName, key string, vsNode AviVsEvhSniModel, isSNI, isDedicated, isIngr bool) {
	found, pathRules := objects.SharedCRDLister().GetFqdnHTTPRulesMapping(host)
	if !found {
		utils.AviLog.Debugf("key: %s, msg: HTTPRules for fqdn %s not found", key, host)
		return
	}

	// finds unique httprules to fetch from the client call
	var getHTTPRules []string
	for _, rule := range pathRules {
		if !utils.HasElem(getHTTPRules, rule) {
			getHTTPRules = append(getHTTPRules, rule)
		}
	}

	// maintains map of rrname+path: rrobj.spec.paths, prefetched for compute ahead
	httpruleNameObjMap := make(map[string]akov1alpha1.HTTPRulePaths)
	for _, httprule := range getHTTPRules {
		pathNSName := strings.Split(httprule, "/")
		httpRuleObj, err := lib.AKOControlConfig().CRDInformers().HTTPRuleInformer.Lister().HTTPRules(pathNSName[0]).Get(pathNSName[1])
		if err != nil {
			utils.AviLog.Debugf("key: %s, msg: httprule not found err: %+v", key, err)
			continue
		} else if httpRuleObj.Status.Status == lib.StatusRejected {
			continue
		}
		for _, path := range httpRuleObj.Spec.Paths {
			httpruleNameObjMap[httprule+path.Target] = path
		}
	}

	// iterate through httpRule which we get from GetFqdnHTTPRulesMapping
	// must contain fqdn.com: {path1: rr1, path2: rr1, path3: rr2}
	for path, rule := range pathRules {
		rrNamespace := strings.Split(rule, "/")[0]
		httpRulePath, ok := httpruleNameObjMap[rule+path]
		if !ok {
			continue
		}
		if httpRulePath.TLS.Type != "" && httpRulePath.TLS.Type != lib.TypeTLSReencrypt {
			continue
		}

		for _, pool := range vsNode.GetPoolRefs() {
			isPathSniEnabled := pool.SniEnabled
			pathSslProfile := pool.SslProfileRef
			pathPkiProfile := pool.PkiProfileRef
			destinationCertNode := pool.PkiProfile
			pathHMs := pool.HealthMonitors
			if poolPath == "" && path == "/" {
				// In case of openfhit Route, the path could be empty, in that case, treat
				// httprule targt path / as that of empty path, to match the pool appropriately.
				path = ""
			}
			// pathprefix match
			// lets say path: / and available pools are cluster--namespace-host_foo-ingName, cluster--namespace-host_bar-ingName
			// then cluster--namespace-host_-ingName should qualify for both pools
			// basic path prefix regex: ^<path_entered>.*
			pathPrefix := strings.ReplaceAll(path, "/", "_")
			var secureRgx, insecureRgx *regexp.Regexp
			if isIngr {
				// sni poolname match regex
				secureRgx = regexp.MustCompile(fmt.Sprintf(`^%s%s-%s.*-%s`, lib.GetNamePrefix(), rrNamespace, host+pathPrefix, ingName))
				// sharedvs poolname match regex
				insecureRgx = regexp.MustCompile(fmt.Sprintf(`^%s%s.*-%s-%s`, lib.GetNamePrefix(), host+pathPrefix, rrNamespace, ingName))
			} else {
				// in case of route, the poolName for secure and insecure ends with service name.
				secureRgx = regexp.MustCompile(fmt.Sprintf(`^%s%s-%s.*-%s-%s`, lib.GetNamePrefix(), rrNamespace, host+pathPrefix, ingName, pool.AviMarkers.ServiceName))
				insecureRgx = regexp.MustCompile(fmt.Sprintf(`^%s%s.*-%s-%s-%s`, lib.GetNamePrefix(), host+pathPrefix, rrNamespace, ingName, pool.AviMarkers.ServiceName))
			}
			if infraSettingName != "" {
				if isIngr {
					secureRgx = regexp.MustCompile(fmt.Sprintf(`^%s%s-%s-%s.*-%s`, lib.GetNamePrefix(), infraSettingName, rrNamespace, host+pathPrefix, ingName))
					insecureRgx = regexp.MustCompile(fmt.Sprintf(`^%s%s-%s.*-%s-%s`, lib.GetNamePrefix(), infraSettingName, host+pathPrefix, rrNamespace, ingName))
				} else {
					// in case of route, the poolName for secure and insecure ends with service name.
					secureRgx = regexp.MustCompile(fmt.Sprintf(`^%s%s-%s.*-%s-%s`, lib.GetNamePrefix(), rrNamespace, host+pathPrefix, ingName, pool.AviMarkers.ServiceName))
					insecureRgx = regexp.MustCompile(fmt.Sprintf(`^%s%s.*-%s-%s-%s`, lib.GetNamePrefix(), host+pathPrefix, rrNamespace, ingName, pool.AviMarkers.ServiceName))
				}
			}
			var poolName string
			// FOR EVH: Build poolname using marker fields.
			if lib.IsEvhEnabled() && pool.AviMarkers.Namespace != "" {
				poolName = lib.GetEvhPoolNameNoEncoding(pool.AviMarkers.IngressName[0], pool.AviMarkers.Namespace, pool.AviMarkers.Host[0],
					pool.AviMarkers.Path[0], pool.AviMarkers.InfrasettingName, pool.AviMarkers.ServiceName, isDedicated)
			} else {
				poolName = pool.Name
			}
			if (secureRgx.MatchString(poolName) && isSNI) || (insecureRgx.MatchString(poolName) && !isSNI) {
				utils.AviLog.Debugf("key: %s, msg: computing poolNode %s for httprule.paths.target %s", key, poolName, path)
				// pool tls
				if httpRulePath.TLS.Type != "" {
					isPathSniEnabled = true
					if httpRulePath.TLS.SSLProfile != "" {
						pathSslProfile = fmt.Sprintf("/api/sslprofile?name=%s", httpRulePath.TLS.SSLProfile)
					} else {
						pathSslProfile = fmt.Sprintf("/api/sslprofile?name=%s", lib.DefaultPoolSSLProfile)
					}

					if httpRulePath.TLS.DestinationCA != "" {
						destinationCertNode = &AviPkiProfileNode{
							Name:   lib.GetPoolPKIProfileName(poolName),
							Tenant: lib.GetTenant(),
							CACert: httpRulePath.TLS.DestinationCA,
						}
						destinationCertNode.AviMarkers = lib.PopulatePoolNodeMarkers(namespace, host, "", pool.AviMarkers.ServiceName, []string{ingName}, []string{path})
					} else {
						destinationCertNode = nil
					}

					if httpRulePath.TLS.PKIProfile != "" {
						pathPkiProfile = fmt.Sprintf("/api/pkiprofile?name=%s", httpRulePath.TLS.PKIProfile)
					}
				}

				var persistenceProfile string
				if httpRulePath.ApplicationPersistence != "" {
					persistenceProfile = fmt.Sprintf("/api/applicationpersistenceprofile?name=%s", httpRulePath.ApplicationPersistence)
				}

				for _, hm := range httpRulePath.HealthMonitors {
					if !utils.HasElem(pathHMs, fmt.Sprintf("/api/healthmonitor?name=%s", hm)) {
						pathHMs = append(pathHMs, fmt.Sprintf("/api/healthmonitor?name=%s", hm))
					}
				}

				pool.SniEnabled = isPathSniEnabled
				pool.SslProfileRef = pathSslProfile
				pool.PkiProfileRef = pathPkiProfile
				pool.PkiProfile = destinationCertNode
				pool.HealthMonitors = pathHMs
				pool.ApplicationPersistence = persistenceProfile

				// from this path, generate refs to this pool node
				pool.LbAlgorithm = httpRulePath.LoadBalancerPolicy.Algorithm
				if pool.LbAlgorithm == lib.LB_ALGORITHM_CONSISTENT_HASH {
					pool.LbAlgorithmHash = httpRulePath.LoadBalancerPolicy.Hash
					if pool.LbAlgorithmHash == lib.LB_ALGORITHM_CONSISTENT_HASH_CUSTOM_HEADER {
						if httpRulePath.LoadBalancerPolicy.HostHeader != "" {
							pool.LbAlgoHostHeader = httpRulePath.LoadBalancerPolicy.HostHeader
						} else {
							utils.AviLog.Warnf("key: %s, HostHeader is not provided for LB_ALGORITHM_CONSISTENT_HASH_CUSTOM_HEADER", key)
						}
					} else if httpRulePath.LoadBalancerPolicy.HostHeader != "" {
						utils.AviLog.Warnf("key: %s, HostHeader is only applicable for LB_ALGORITHM_CONSISTENT_HASH_CUSTOM_HEADER", key)
					}
				}

				// There is no need to convert the servicemetadata CRDStatus to INACTIVE, in case
				// no appropriate HttpRule is found, since we build the PoolNodes from scrach every time
				// while building the graph. If no HttpRule is found, the CRDStatus will remain empty.
				// For the same reason, we cannot track ACTIVE -> INACTIVE transitions in HTTPRule.
				pool.ServiceMetadata.CRDStatus = lib.CRDMetadata{
					Type:   "HTTPRule",
					Value:  rule + "/" + path,
					Status: lib.CRDActive,
				}
				utils.AviLog.Infof("key: %s, Attached httprule %s on pool %s", key, rule, pool.Name)
			}
		}
	}

}
