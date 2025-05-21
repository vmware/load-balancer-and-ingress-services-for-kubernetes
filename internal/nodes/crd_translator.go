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

	"github.com/jinzhu/copier"
	"github.com/vmware/alb-sdk/go/models"
	"google.golang.org/protobuf/proto"
	corev1 "k8s.io/api/core/v1"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"
	akov1beta1 "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/apis/ako/v1beta1"

	akov1alpha2 "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/apis/ako/v1alpha2"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
)

func BuildL7HostRule(host, key string, vsNode AviVsEvhSniModel) {
	// use host to find out HostRule CRD if it exists
	// The host that comes here will have a proper FQDN, either from the Ingress/Route (foo.com)
	// or the SharedVS FQDN (Shared-L7-1.com).
	found, hrNamespaceName := objects.SharedCRDLister().GetFQDNToHostruleMappingWithType(host)
	deleteCase := false
	if !found {
		utils.AviLog.Warnf("key: %s, msg: No HostRule found for virtualhost: %s in Cache", key, host)
		deleteCase = true
	}

	var err error
	var hrNSName []string
	var hostrule *akov1beta1.HostRule
	if !deleteCase {
		hrNSName = strings.Split(hrNamespaceName, "/")
		hostrule, err = lib.AKOControlConfig().CRDInformers().HostRuleInformer.Lister().HostRules(hrNSName[0]).Get(hrNSName[1])
		if err != nil {
			utils.AviLog.Debugf("key: %s, msg: No HostRule found for virtualhost: %s msg: %v", key, host, err)
			deleteCase = true
		} else if hostrule.Status.Status == lib.StatusRejected {
			// do not apply a rejected hostrule, this way the VS would retain
			return
		} else {
			if lib.GetTenantInNamespace(hostrule.Namespace) != vsNode.GetTenant() {
				utils.AviLog.Warnf("key: %s, msg: Tenant annotation in hostrule namespace %s does not matches with the tenant of host %s ", key, hostrule.Namespace, host)
				return
			}
		}
	}

	// host specific
	var vsWafPolicy, vsAppProfile, vsAnalyticsProfile, vsSslProfile, vsNetworkSecurityPolicy *string
	var vsErrorPageProfile, lbIP string
	var vsSslKeyCertificates []string
	var vsEnabled *bool
	var crdStatus lib.CRDMetadata
	var vsICAPProfile []string

	// Initializing the values of vsHTTPPolicySets and vsDatascripts, using a nil value would impact the value of VS checksum
	vsHTTPPolicySets := []string{}
	vsDatascripts := []string{}
	var analyticsPolicy *models.AnalyticsPolicy
	var vsStringGroupRefs []*AviStringGroupNode

	// Get the existing VH domain names and then manipulate it based on the aliases in Hostrule CRD.
	VHDomainNames := vsNode.GetVHDomainNames()

	portProtocols := []AviPortHostProtocol{
		{Port: 80, Protocol: utils.HTTP},
	}

	if vsNode.IsSecure() || !vsNode.IsDedicatedVS() {
		portProtocols = append(portProtocols, AviPortHostProtocol{Port: 443, Protocol: utils.HTTP, EnableSSL: true})
	}

	if !deleteCase {
		if hostrule.Spec.VirtualHost.TLS.SSLKeyCertificate.Type == akov1beta1.HostRuleSecretTypeAviReference &&
			hostrule.Spec.VirtualHost.TLS.SSLKeyCertificate.Name != "" {
			vsSslKeyCertificates = append(vsSslKeyCertificates, fmt.Sprintf("/api/sslkeyandcertificate?name=%s", hostrule.Spec.VirtualHost.TLS.SSLKeyCertificate.Name))
			vsNode.SetSSLKeyCertRefs([]*AviTLSKeyCertNode{})
		}

		if hostrule.Spec.VirtualHost.TLS.SSLKeyCertificate.AlternateCertificate.Type == akov1beta1.HostRuleSecretTypeAviReference &&
			hostrule.Spec.VirtualHost.TLS.SSLKeyCertificate.AlternateCertificate.Name != "" {
			vsSslKeyCertificates = append(vsSslKeyCertificates, fmt.Sprintf("/api/sslkeyandcertificate?name=%s", hostrule.Spec.VirtualHost.TLS.SSLKeyCertificate.AlternateCertificate.Name))
			vsNode.SetSSLKeyCertRefs([]*AviTLSKeyCertNode{})
		}

		if hostrule.Spec.VirtualHost.TLS.SSLProfile != "" {
			vsSslProfile = proto.String(fmt.Sprintf("/api/sslprofile?name=%s", hostrule.Spec.VirtualHost.TLS.SSLProfile))
		}

		if hostrule.Spec.VirtualHost.WAFPolicy != "" {
			vsWafPolicy = proto.String(fmt.Sprintf("/api/wafpolicy?name=%s", hostrule.Spec.VirtualHost.WAFPolicy))
		}

		if hostrule.Spec.VirtualHost.ApplicationProfile != "" {
			vsAppProfile = proto.String(fmt.Sprintf("/api/applicationprofile?name=%s", hostrule.Spec.VirtualHost.ApplicationProfile))
		}

		if len(hostrule.Spec.VirtualHost.ICAPProfile) != 0 {
			vsICAPProfile = []string{fmt.Sprintf("/api/icapprofile?name=%s", hostrule.Spec.VirtualHost.ICAPProfile[0])}
		}
		if hostrule.Spec.VirtualHost.ErrorPageProfile != "" {
			vsErrorPageProfile = fmt.Sprintf("/api/errorpageprofile?name=%s", hostrule.Spec.VirtualHost.ErrorPageProfile)
		}

		if hostrule.Spec.VirtualHost.AnalyticsProfile != "" {
			vsAnalyticsProfile = proto.String(fmt.Sprintf("/api/analyticsprofile?name=%s", hostrule.Spec.VirtualHost.AnalyticsProfile))
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
				if hostrule.Spec.VirtualHost.TCPSettings.Listeners != nil {
					portProtocols = []AviPortHostProtocol{}
				}
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
		if hostrule.Spec.VirtualHost.NetworkSecurityPolicy != "" {
			if vsNode.IsSharedVS() || vsNode.IsDedicatedVS() {
				vsNetworkSecurityPolicy = proto.String(fmt.Sprintf("/api/networksecuritypolicy?name=%s", hostrule.Spec.VirtualHost.NetworkSecurityPolicy))
			} else {
				utils.AviLog.Warnf("key: %s, can not associate network security policy with host which is attached to child virtual service. Configuration is ignored", key)
				lib.AKOControlConfig().EventRecorder().Eventf(hostrule, corev1.EventTypeWarning, lib.InvalidConfiguration,
					"can not associate network security policy with host which is attached to child virtual service. Configuration is ignored")
			}
		}
		vsEnabled = hostrule.Spec.VirtualHost.EnableVirtualHost
		crdStatus = lib.CRDMetadata{
			Type:   "HostRule",
			Value:  hostrule.Namespace + "/" + hostrule.Name,
			Status: lib.CRDActive,
		}

		if hostrule.Spec.VirtualHost.AnalyticsPolicy != nil {
			var infinite uint32 = 0 // Special value to set log duration as infinite
			// defaults to 'infinite' if hostrule doesn't specify a duration
			analyticsPolicy = &models.AnalyticsPolicy{
				FullClientLogs: &models.FullClientLogs{
					Duration: &infinite,
				},
				AllHeaders: hostrule.Spec.VirtualHost.AnalyticsPolicy.LogAllHeaders,
			}
			if hostrule.Spec.VirtualHost.AnalyticsPolicy.FullClientLogs != nil {
				analyticsPolicy.FullClientLogs.Enabled = hostrule.Spec.VirtualHost.AnalyticsPolicy.FullClientLogs.Enabled
				analyticsPolicy.FullClientLogs.Throttle = lib.GetThrottle(hostrule.Spec.VirtualHost.AnalyticsPolicy.FullClientLogs.Throttle)

				// only update duration if duration is actually specified in hr
				if hostrule.Spec.VirtualHost.AnalyticsPolicy.FullClientLogs.Duration != nil {
					analyticsPolicy.FullClientLogs.Duration = hostrule.Spec.VirtualHost.AnalyticsPolicy.FullClientLogs.Duration
				}
			}
		}

		for _, alias := range hostrule.Spec.VirtualHost.Aliases {
			if !utils.HasElem(VHDomainNames, alias) {
				VHDomainNames = append(VHDomainNames, alias)
			}
		}
		if lib.IsEvhEnabled() {
			if hostrule.Spec.VirtualHost.L7Rule != "" {
				BuildL7Rule(host, key, hostrule.Spec.VirtualHost.L7Rule, hrNSName[0], vsNode)
			} else {
				vsNode.GetGeneratedFields().ConvertL7RuleFieldsToNil()
			}
		}

		if !hostrule.Spec.VirtualHost.HTTPPolicy.Overwrite && (hostrule.Spec.VirtualHost.UseRegex || hostrule.Spec.VirtualHost.ApplicationRootPath != "") {
			if !vsNode.IsSharedVS() {
				if !lib.IsEvhEnabled() && vsNode.IsDedicatedVS() && !vsNode.IsSecure() {
					utils.AviLog.Debugf("key: %s, Regex and App-root are not supported for insecure SNI virtual service", key)
				} else {
					// BuildRegexAppRootForHostRule applies useRegex and applicationRootPath to vsNode if applicable
					vsStringGroupRefs = BuildRegexAppRootForHostRule(hostrule, vsNode, host, key)
				}
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
		vsNode.GetGeneratedFields().ConvertL7RuleFieldsToNil()
	}

	vsNode.SetSslKeyAndCertificateRefs(vsSslKeyCertificates)
	vsNode.SetWafPolicyRef(vsWafPolicy)
	vsNode.SetHttpPolicySetRefs(vsHTTPPolicySets)
	vsNode.SetICAPProfileRefs(vsICAPProfile)
	vsNode.SetAppProfileRef(vsAppProfile)
	vsNode.SetAnalyticsProfileRef(vsAnalyticsProfile)
	vsNode.SetErrorPageProfileRef(vsErrorPageProfile)
	vsNode.SetSSLProfileRef(vsSslProfile)
	vsNode.SetVsDatascriptRefs(vsDatascripts)
	vsNode.SetEnabled(vsEnabled)
	vsNode.SetAnalyticsPolicy(analyticsPolicy)
	if len(portProtocols) != 0 {
		vsNode.SetPortProtocols(portProtocols)
	}
	vsNode.SetVSVIPLoadBalancerIP(lbIP)
	vsNode.SetVHDomainNames(VHDomainNames)
	vsNode.SetNetworkSecurityPolicyRef(vsNetworkSecurityPolicy)

	serviceMetadataObj := vsNode.GetServiceMetadata()
	serviceMetadataObj.CRDStatus = crdStatus
	vsNode.SetServiceMetadata(serviceMetadataObj)
	vsNode.SetStringGroupRefs(vsStringGroupRefs)
}

// BuildOnlyRegexAppRoot builds only Regex and Approot with HostRule for vs node.
// This is added because in some cases, we apply hostrule first to the vs node followed by aviinfrasetting.
// Also, in some cases the ports added by aviinfrasetting first are over-written by hostrule and then added back.
// Due to these sequences app-root path redirect rules may not be added for listener ports set from aviinfrasetting.
// Hence we process only regex app-root again after all ports are updated.
func BuildOnlyRegexAppRoot(host, key string, vsNode AviVsEvhSniModel) {
	// use host to find out HostRule CRD if it exists
	// The host that comes here will have a proper FQDN, from the Ingress/Route (foo.com)
	found, hrNamespaceName := objects.SharedCRDLister().GetFQDNToHostruleMappingWithType(host)
	if !found {
		utils.AviLog.Debugf("key: %s, msg: No HostRule found for virtualhost: %s in Cache", key, host)
		return
	}

	var err error
	var hrNSName []string
	var hostrule *akov1beta1.HostRule
	hrNSName = strings.Split(hrNamespaceName, "/")
	hostrule, err = lib.AKOControlConfig().CRDInformers().HostRuleInformer.Lister().HostRules(hrNSName[0]).Get(hrNSName[1])
	if err != nil {
		utils.AviLog.Debugf("key: %s, msg: No HostRule found for virtualhost: %s msg: %v", key, host, err)
		return
	} else if hostrule.Status.Status == lib.StatusRejected {
		// do not apply a rejected hostrule, this way the VS would retain
		utils.AviLog.Debugf("key: %s, msg: hostrule %s is in rejected state", key, hrNSName)
		return
	} else {
		if lib.GetTenantInNamespace(hostrule.Namespace) != vsNode.GetTenant() {
			utils.AviLog.Warnf("key: %s, msg: Tenant annotation in hostrule namespace %s does not matches with the tenant of host %s ", key, hostrule.Namespace, host)
			return
		}
	}
	if !hostrule.Spec.VirtualHost.HTTPPolicy.Overwrite && (hostrule.Spec.VirtualHost.UseRegex || hostrule.Spec.VirtualHost.ApplicationRootPath != "") {
		if !vsNode.IsSharedVS() {
			if !lib.IsEvhEnabled() && vsNode.IsDedicatedVS() && !vsNode.IsSecure() {
				utils.AviLog.Debugf("key: %s, Regex and App-root are not supported for insecure SNI virtual service", key)
			} else {
				vsStringGroupRefs := BuildRegexAppRootForHostRule(hostrule, vsNode, host, key)
				vsNode.SetStringGroupRefs(vsStringGroupRefs)
				utils.AviLog.Infof("key: %s, Successfully updated Regex and AppRoot properties with hostrule %s on dedicated vsNode %s", key, hrNamespaceName, vsNode.GetName())
			}
		}
	}
}

func BuildRegexAppRootForHostRule(hostrule *akov1beta1.HostRule, vsNode AviVsEvhSniModel, host, key string) []*AviStringGroupNode {
	var vsStringGroupRefs []*AviStringGroupNode

	httpPolicyRefs := vsNode.GetHttpPolicyRefs()
	for _, httpPolicyRef := range httpPolicyRefs {
		var regexhppMap []AviHostPathPortPoolPG
		var redirectPorts []AviRedirectPort
		for _, hppMap := range httpPolicyRef.HppMap {
			if hostrule.Spec.VirtualHost.ApplicationRootPath != "" {
				if hppMap.Path != nil && len(hppMap.Path) > 0 {
					path := hppMap.Path[0]
					var protocol string
					if path == "/" {
						for _, portProto := range vsNode.GetPortProtocols() {
							if portProto.EnableSSL {
								protocol = "HTTPS"
							} else {
								protocol = "HTTP"
							}
							redirectPort := AviRedirectPort{
								StatusCode:        lib.STATUS_REDIRECT,
								Protocol:          protocol,
								Path:              path,
								RedirectPort:      portProto.Port,
								RedirectPath:      hostrule.Spec.VirtualHost.ApplicationRootPath[1:],
								MatchCriteriaPath: "EQUALS",
								MatchCriteriaPort: "IS_IN",
							}
							redirectPorts = append(redirectPorts, redirectPort)
						}
						hppMap.Path[0] = hostrule.Spec.VirtualHost.ApplicationRootPath
					} else if path == hostrule.Spec.VirtualHost.ApplicationRootPath {
						for _, childPath := range vsNode.GetPaths() {
							if childPath == "/" {
								for _, portProto := range vsNode.GetPortProtocols() {
									if portProto.EnableSSL {
										protocol = "HTTPS"
									} else {
										protocol = "HTTP"
									}
									redirectPort := AviRedirectPort{
										StatusCode:        lib.STATUS_REDIRECT,
										Protocol:          protocol,
										Path:              childPath,
										RedirectPort:      portProto.Port,
										RedirectPath:      hostrule.Spec.VirtualHost.ApplicationRootPath[1:],
										MatchCriteriaPath: "EQUALS",
										MatchCriteriaPort: "IS_IN",
									}
									redirectPorts = append(redirectPorts, redirectPort)
								}
							}
						}
					}
				}
			}
			if hostrule.Spec.VirtualHost.UseRegex {
				if hppMap.Path != nil && len(hppMap.Path) > 0 {
					var regexStringGroupName string
					path := hppMap.Path[0]
					regexStringGroupName = lib.GetEncodedStringGroupName(host, path)
					kv := &models.KeyValue{
						Key: &path,
					}
					hppMap.MatchCase = "INSENSITIVE"
					hppMap.MatchCriteria = "REGEX_MATCH"

					tenant := vsNode.GetTenant()
					regexStringGroup := &models.StringGroup{
						TenantRef:    &tenant,
						Type:         proto.String("SG_TYPE_STRING"),
						LongestMatch: proto.Bool(true),
						Name:         &regexStringGroupName,
						Kv:           []*models.KeyValue{kv},
					}
					aviStringGroupNode := AviStringGroupNode{StringGroup: regexStringGroup}
					aviStringGroupNode.CloudConfigCksum = aviStringGroupNode.GetCheckSum()
					vsStringGroupRefs = append(vsStringGroupRefs, &aviStringGroupNode)
					stringGroupRef := []string{"/api/stringgroup?name=" + regexStringGroupName}
					hppMap.StringGroupRefs = stringGroupRef
				}
				if !lib.IsEvhEnabled() {
					if hppMap.PoolGroup != "" && !lib.IsNameEncoded(hppMap.PoolGroup) {
						hppMap.PoolGroup = lib.GetEncodedSniPGPoolNameforRegex(hppMap.PoolGroup)
					}
					if hppMap.Pool != "" && !lib.IsNameEncoded(hppMap.Pool) {
						hppMap.Pool = lib.GetEncodedSniPGPoolNameforRegex(hppMap.Pool)
					}
				}
				hppMap.CalculateCheckSum()

			}
			regexhppMap = append(regexhppMap, hppMap)
		}
		httpPolicyRef.HppMap = regexhppMap
		if len(redirectPorts) != 0 {
			httpPolicyRef.RedirectPorts = redirectPorts
		}
	}
	if !lib.IsEvhEnabled() && hostrule.Spec.VirtualHost.UseRegex {
		for _, pool := range vsNode.GetPoolRefs() {
			if !lib.IsNameEncoded(pool.Name) {
				pool.Name = lib.GetEncodedSniPGPoolNameforRegex(pool.Name)
			}
		}
		for _, pg := range vsNode.GetPoolGroupRefs() {
			if !lib.IsNameEncoded(pg.Name) {
				pg.Name = lib.GetEncodedSniPGPoolNameforRegex(pg.Name)
				for _, member := range pg.Members {
					poolName := strings.TrimPrefix(*member.PoolRef, "/api/pool?name=")
					encodedPoolName := lib.GetEncodedSniPGPoolNameforRegex(poolName)
					poolRef := "/api/pool?name=" + encodedPoolName
					member.PoolRef = &poolRef
				}
			}
		}
	}
	vsNode.SetHttpPolicyRefs(httpPolicyRefs)
	return vsStringGroupRefs
}

// BuildPoolHTTPRule notes
// when we get an ingress update and we are building the corresponding pools of that ingress
// we need to get all httprules which match ingress's host/path
func BuildPoolHTTPRule(host, poolPath, ingName, namespace, infraSettingName, key string, vsNode AviVsEvhSniModel, isSNI, isDedicated bool) {
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
	httpruleNameObjMap := make(map[string]akov1beta1.HTTPRulePaths)
	for _, httprule := range getHTTPRules {
		pathNSName := strings.Split(httprule, "/")
		httpRuleObj, err := lib.AKOControlConfig().CRDInformers().HTTPRuleInformer.Lister().HTTPRules(pathNSName[0]).Get(pathNSName[1])
		if err != nil {
			utils.AviLog.Debugf("key: %s, msg: httprule not found err: %+v", key, err)
			continue
		} else if httpRuleObj.Status.Status == lib.StatusRejected {
			continue
		} else {
			if lib.GetTenantInNamespace(httpRuleObj.Namespace) != vsNode.GetTenant() {
				utils.AviLog.Warnf("key: %s, msg: Tenant annotation in httpRule namespace %s does not matches with the tenant of host %s ", key, httpRuleObj.Namespace, host)
				continue
			}
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
			pathHMs := pool.HealthMonitorRefs
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
			// sni poolname match regex
			secureRgx := regexp.MustCompile(fmt.Sprintf(`^%s%s-%s.*-%s`, lib.GetNamePrefix(), rrNamespace, host+pathPrefix, ingName))
			// sharedvs poolname match regex
			insecureRgx := regexp.MustCompile(fmt.Sprintf(`^%s%s.*-%s-%s`, lib.GetNamePrefix(), host+pathPrefix, rrNamespace, ingName))
			if infraSettingName != "" {
				secureRgx = regexp.MustCompile(fmt.Sprintf(`^%s%s-%s-%s.*-%s`, lib.GetNamePrefix(), infraSettingName, rrNamespace, host+pathPrefix, ingName))
				insecureRgx = regexp.MustCompile(fmt.Sprintf(`^%s%s-%s.*-%s-%s`, lib.GetNamePrefix(), infraSettingName, host+pathPrefix, rrNamespace, ingName))
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
						pathSslProfile = proto.String(fmt.Sprintf("/api/sslprofile?name=%s", httpRulePath.TLS.SSLProfile))
					} else {
						pathSslProfile = proto.String(fmt.Sprintf("/api/sslprofile?name=%s", lib.DefaultPoolSSLProfile))
					}

					if httpRulePath.TLS.DestinationCA != "" {
						destinationCertNode = &AviPkiProfileNode{
							Name:   lib.GetPoolPKIProfileName(poolName),
							Tenant: vsNode.GetTenant(),
							CACert: httpRulePath.TLS.DestinationCA,
						}
						destinationCertNode.AviMarkers = lib.PopulatePoolNodeMarkers(namespace, host, "", pool.AviMarkers.ServiceName, []string{ingName}, []string{path})
					} else {
						destinationCertNode = nil
					}

					if httpRulePath.TLS.PKIProfile != "" {
						pathPkiProfile = proto.String(fmt.Sprintf("/api/pkiprofile?name=%s", httpRulePath.TLS.PKIProfile))
					}
				}

				var persistenceProfile *string
				if httpRulePath.ApplicationPersistence != "" {
					persistenceProfile = proto.String(fmt.Sprintf("/api/applicationpersistenceprofile?name=%s", httpRulePath.ApplicationPersistence))
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
				pool.HealthMonitorRefs = pathHMs
				pool.ApplicationPersistenceProfileRef = persistenceProfile

				if httpRulePath.EnableHttp2 != nil {
					pool.EnableHttp2 = httpRulePath.EnableHttp2
				}

				// from this path, generate refs to this pool node
				if httpRulePath.LoadBalancerPolicy.Algorithm != "" {
					pool.LbAlgorithm = proto.String(httpRulePath.LoadBalancerPolicy.Algorithm)
				}
				if pool.LbAlgorithm != nil &&
					*pool.LbAlgorithm == lib.LB_ALGORITHM_CONSISTENT_HASH {
					pool.LbAlgorithmHash = proto.String(httpRulePath.LoadBalancerPolicy.Hash)
					if pool.LbAlgorithmHash != nil &&
						*pool.LbAlgorithmHash == lib.LB_ALGORITHM_CONSISTENT_HASH_CUSTOM_HEADER {
						if httpRulePath.LoadBalancerPolicy.HostHeader != "" {
							pool.LbAlgorithmConsistentHashHdr = proto.String(httpRulePath.LoadBalancerPolicy.HostHeader)
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

func BuildL7SSORule(host, key string, vsNode AviVsEvhSniModel) {
	// use host to find out SSORule CRD if it exists
	// The host that comes here will have a proper FQDN, either from the Ingress/Route (foo.com)

	found, srNamespaceName := objects.SharedCRDLister().GetFQDNToSSORuleMapping(host)
	deleteCase := false
	if !found {
		utils.AviLog.Debugf("key: %s, msg: No SSORule found for virtualhost: %s in Cache", key, host)
		deleteCase = true
	}

	var err error
	var srNSName []string
	var ssoRule *akov1alpha2.SSORule
	if !deleteCase {
		srNSName = strings.Split(srNamespaceName, "/")
		ssoRule, err = lib.AKOControlConfig().CRDInformers().SSORuleInformer.Lister().SSORules(srNSName[0]).Get(srNSName[1])
		if err != nil {
			utils.AviLog.Debugf("key: %s, msg: No SSORule found for virtualhost: %s msg: %v", key, host, err)
			deleteCase = true
		} else if ssoRule.Status.Status == lib.StatusRejected {
			// do not apply a rejected SSORule, this way the VS would retain
			return
		} else {
			if lib.GetTenantInNamespace(ssoRule.Namespace) != vsNode.GetTenant() {
				utils.AviLog.Warnf("key: %s, msg: Tenant annotation in SSORule namespace %s does not matches with the tenant of host %s ", key, ssoRule.Namespace, host)
				return
			}
		}
	}
	var crdStatus lib.CRDMetadata

	if !deleteCase {
		copier.CopyWithOption(vsNode, &ssoRule.Spec, copier.Option{DeepCopy: true})
		//setting the fqdn to nil so that fqdn for child vs is not populated
		generatedFields := vsNode.GetGeneratedFields()
		generatedFields.Fqdn = nil
		generatedFields.ConvertToRef()
		if ssoRule.Spec.OauthVsConfig != nil {
			if len(ssoRule.Spec.OauthVsConfig.OauthSettings) != 0 {
				for i, oauthSetting := range ssoRule.Spec.OauthVsConfig.OauthSettings {
					if oauthSetting.AppSettings != nil {
						// getting clientSecret from k8s secret
						clientSecretObj, err := utils.GetInformers().SecretInformer.Lister().Secrets(ssoRule.Namespace).Get(*oauthSetting.AppSettings.ClientSecret)
						if err != nil || clientSecretObj == nil {
							utils.AviLog.Errorf("key: %s, msg: Client secret not found for ssoRule obj: %s msg: %v", key, *oauthSetting.AppSettings.ClientSecret, err)
							return
						}
						clientSecretString := string(clientSecretObj.Data["clientSecret"])
						generatedFields.OauthVsConfig.OauthSettings[i].AppSettings.ClientSecret = &clientSecretString
					}
					if oauthSetting.ResourceServer != nil {
						if oauthSetting.ResourceServer.OpaqueTokenParams != nil {
							// getting serverSecret from k8s secret
							serverSecretObj, err := utils.GetInformers().SecretInformer.Lister().Secrets(ssoRule.Namespace).Get(*oauthSetting.ResourceServer.OpaqueTokenParams.ServerSecret)
							if err != nil || serverSecretObj == nil {
								utils.AviLog.Errorf("key: %s, msg: Server secret not found for ssoRule obj: %s msg: %v", key, *oauthSetting.ResourceServer.OpaqueTokenParams.ServerSecret, err)
								return
							}
							serverSecretString := string(serverSecretObj.Data["serverSecret"])
							generatedFields.OauthVsConfig.OauthSettings[i].ResourceServer.OpaqueTokenParams.ServerSecret = &serverSecretString
						} else {
							// setting IntrospectionDataTimeout to nil if jwt params are set
							generatedFields.OauthVsConfig.OauthSettings[i].ResourceServer.IntrospectionDataTimeout = nil
						}
					}
				}
			}
		}

		if ssoRule.Spec.SamlSpConfig != nil {
			if *ssoRule.Spec.SamlSpConfig.AuthnReqAcsType != lib.SAML_AUTHN_REQ_ACS_TYPE_INDEX {
				generatedFields.SamlSpConfig.AcsIndex = nil
			}
		}
		crdStatus = lib.CRDMetadata{
			Type:   "SSORule",
			Value:  ssoRule.Namespace + "/" + ssoRule.Name,
			Status: lib.CRDActive,
		}

		utils.AviLog.Infof("key: %s, Successfully attached SSORule %s on vsNode %s", key, srNamespaceName, vsNode.GetName())
	} else {
		generatedFields := vsNode.GetGeneratedFields()
		generatedFields.OauthVsConfig = nil
		generatedFields.SamlSpConfig = nil
		generatedFields.SsoPolicyRef = nil
		generatedFields.Fqdn = nil
		if vsNode.GetServiceMetadata().CRDStatus.Value != "" {
			crdStatus = vsNode.GetServiceMetadata().CRDStatus
			crdStatus.Status = lib.CRDInactive
		}
		if srNamespaceName != "" {
			utils.AviLog.Infof("key: %s, Successfully detached SSORule %s from vsNode %s", key, srNamespaceName, vsNode.GetName())
		}
	}

	serviceMetadataObj := vsNode.GetServiceMetadata()
	serviceMetadataObj.CRDStatus = crdStatus
	vsNode.SetServiceMetadata(serviceMetadataObj)
}

func BuildL7Rule(host, key, l7RuleName, namespace string, vsNode AviVsEvhSniModel) {
	deleteL7RuleCase := false
	l7Rule, err := lib.AKOControlConfig().CRDInformers().L7RuleInformer.Lister().L7Rules(namespace).Get(l7RuleName)
	if err != nil {
		utils.AviLog.Debugf("key: %s, msg: No L7Rule found for virtualhost: %s msg: %v", key, host, err)
		deleteL7RuleCase = true
	} else if l7Rule.Status.Status == lib.StatusRejected {
		// do not apply a rejected L7Rule, this way the VS would retain
		return
	} else {
		if lib.GetTenantInNamespace(l7Rule.Namespace) != vsNode.GetTenant() {
			utils.AviLog.Warnf("key: %s, msg: Tenant annotation in l7Rule namespace %s does not matches with the tenant of host %s ", key, l7Rule.Namespace, host)
			return
		}
	}
	generatedFields := vsNode.GetGeneratedFields()
	if !deleteL7RuleCase {
		utils.AviLog.Debugf("key: %s, msg: applying l7 Rule %s", key, l7Rule.Name)
		copier.CopyWithOption(vsNode, &l7Rule.Spec, copier.Option{DeepCopy: true})
		if !vsNode.IsDedicatedVS() {
			if !vsNode.IsSharedVS() {
				generatedFields.ConvertL7RuleParentOnlyFieldsToNil()
			}
		}
		utils.AviLog.Infof("key: %s, Successfully attached L7Rule %s on vsNode %s", key, l7RuleName, vsNode.GetName())
		generatedFields.ConvertToRef()
	} else {
		generatedFields.ConvertL7RuleFieldsToNil()
	}
}
