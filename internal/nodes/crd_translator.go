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
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	akov1alpha1 "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/apis/ako/v1alpha1"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/cache"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/status"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	"github.com/avinetworks/sdk/go/models"
)

func BuildL7HostRule(host, namespace, ingName, key string, vsNode *AviVsNode) {
	// use host to find out HostRule CRD if it exists
	found, hrNamespaceName := objects.SharedCRDLister().GetFQDNToHostruleMapping(host)
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
		hostrule, err = lib.GetCRDInformers().HostRuleInformer.Lister().HostRules(hrNSName[0]).Get(hrNSName[1])
		if err != nil {
			utils.AviLog.Debugf("key: %s, msg: No HostRule found for virtualhost: %s msg: %v", key, host, err)
			deleteCase = true
		} else if hostrule.Status.Status == lib.StatusRejected {
			// do not apply a rejected hostrule, this way the VS would retain
			return
		}
	}

	// host specific
	var vsWafPolicy, vsAppProfile, vsSslKeyCertificate, vsErrorPageProfile, vsAnalyticsProfile, vsSslProfile string
	var vsEnabled *bool
	var crdStatus cache.CRDMetadata

	// Initializing the values of vsHTTPPolicySets and vsDatascripts, using a nil value would impact the value of VS checksum
	vsHTTPPolicySets := []string{}
	vsDatascripts := []string{}

	if !deleteCase {
		if hostrule.Spec.VirtualHost.TLS.SSLKeyCertificate.Name != "" {
			vsSslKeyCertificate = fmt.Sprintf("/api/sslkeyandcertificate?name=%s", hostrule.Spec.VirtualHost.TLS.SSLKeyCertificate.Name)
			vsNode.SSLKeyCertRefs = []*AviTLSKeyCertNode{}
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
			vsNode.HttpPolicyRefs = []*AviHttpPolicySetNode{}
		}

		for _, script := range hostrule.Spec.VirtualHost.Datascripts {
			if !utils.HasElem(vsDatascripts, fmt.Sprintf("/api/vsdatascriptset?name=%s", script)) {
				vsDatascripts = append(vsDatascripts, fmt.Sprintf("/api/vsdatascriptset?name=%s", script))
			}
		}

		vsEnabled = hostrule.Spec.VirtualHost.EnableVirtualHost
		crdStatus = cache.CRDMetadata{
			Type:   "HostRule",
			Value:  hostrule.Namespace + "/" + hostrule.Name,
			Status: "ACTIVE",
		}
	} else {
		if vsNode.ServiceMetadata.CRDStatus.Value != "" {
			crdStatus = vsNode.ServiceMetadata.CRDStatus
			crdStatus.Status = "INACTIVE"
		}
	}

	vsNode.SSLKeyCertAviRef = vsSslKeyCertificate
	vsNode.WafPolicyRef = vsWafPolicy
	vsNode.HttpPolicySetRefs = vsHTTPPolicySets
	vsNode.AppProfileRef = vsAppProfile
	vsNode.AnalyticsProfileRef = vsAnalyticsProfile
	vsNode.ErrorPageProfileRef = vsErrorPageProfile
	vsNode.SSLProfileRef = vsSslProfile
	vsNode.VsDatascriptRefs = vsDatascripts
	vsNode.Enabled = vsEnabled
	vsNode.ServiceMetadata.CRDStatus = crdStatus

	utils.AviLog.Infof("key: %s, Attached hostrule %s on vsNode %s", key, hrNamespaceName, vsNode.Name)
}

// BuildPoolHTTPRule notes
// when we get an ingress update and we are building the corresponding pools of that ingress
// we need to get all httprules which match ingress's host/path
func BuildPoolHTTPRule(host, path, ingName, namespace, key string, vsNode *AviVsNode, isSNI bool) {
	found, pathRules := objects.SharedCRDLister().GetFqdnHTTPRulesMapping(host)
	if !found {
		utils.AviLog.Warnf("key: %s, msg: HTTPRules for fqdn %s not found", key, host)
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
		httpRuleObj, err := lib.GetCRDInformers().HTTPRuleInformer.Lister().HTTPRules(pathNSName[0]).Get(pathNSName[1])
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

		for _, pool := range vsNode.PoolRefs {
			isPathSniEnabled := pool.SniEnabled
			pathSslProfile := pool.SslProfileRef
			destinationCertNode := pool.PkiProfile
			pathHMs := pool.HealthMonitors

			// pathprefix match
			// lets say path: / and available pools are cluster--namespace-host_foo-ingName, cluster--namespace-host_bar-ingName
			// then cluster--namespace-host_-ingName should qualify for both pools
			// basic path prefix regex: ^<path_entered>.*
			pathPrefix := strings.ReplaceAll(path, "/", "_")
			// sni poolname match regex
			secureRgx := regexp.MustCompile(fmt.Sprintf(`^%s%s-%s%s.*-%s`, lib.GetNamePrefix(), rrNamespace, host, pathPrefix, ingName))
			// sharedvs poolname match regex
			insecureRgx := regexp.MustCompile(fmt.Sprintf(`^%s%s.*-%s-%s`, lib.GetNamePrefix(), host+pathPrefix, rrNamespace, ingName))

			if (secureRgx.MatchString(pool.Name) && isSNI) || (insecureRgx.MatchString(pool.Name) && !isSNI) {
				utils.AviLog.Debugf("key: %s, msg: computing poolNode %s for httprule.paths.target %s", key, pool.Name, path)
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
							Name:   lib.GetPoolPKIProfileName(pool.Name),
							Tenant: lib.GetTenant(),
							CACert: httpRulePath.TLS.DestinationCA,
						}
					} else {
						destinationCertNode = nil
					}
				}

				for _, hm := range httpRulePath.HealthMonitors {
					if !utils.HasElem(pathHMs, fmt.Sprintf("/api/healthmonitor?name=%s", hm)) {
						pathHMs = append(pathHMs, fmt.Sprintf("/api/healthmonitor?name=%s", hm))
					}
				}

				pool.SniEnabled = isPathSniEnabled
				pool.SslProfileRef = pathSslProfile
				pool.PkiProfile = destinationCertNode
				pool.HealthMonitors = pathHMs

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
				pool.ServiceMetadata.CRDStatus = cache.CRDMetadata{
					Type:   "HTTPRule",
					Value:  rule,
					Status: "ACTIVE",
				}
				utils.AviLog.Infof("key: %s, Attached httprule %s on pool %s", key, rule, pool.Name)
			}
		}
	}

	return
}

// validateHostRuleObj would do validation checks
// update internal CRD caches, and push relevant ingresses to ingestion
func validateHostRuleObj(key string, hostrule *akov1alpha1.HostRule) error {
	var err error
	fqdn := hostrule.Spec.VirtualHost.Fqdn
	foundHost, foundHR := objects.SharedCRDLister().GetFQDNToHostruleMapping(fqdn)
	if foundHost && foundHR != hostrule.Namespace+"/"+hostrule.Name {
		err = fmt.Errorf("duplicate fqdn %s found in %s", fqdn, foundHR)
		status.UpdateHostRuleStatus(key, hostrule, status.UpdateCRDStatusOptions{
			Status: lib.StatusRejected,
			Error:  err.Error(),
		})
		utils.AviLog.Warnf("key: %s, msg: %v", key, err)
		return err
	}

	refData := map[string]string{
		hostrule.Spec.VirtualHost.WAFPolicy:                  "WafPolicy",
		hostrule.Spec.VirtualHost.ApplicationProfile:         "AppProfile",
		hostrule.Spec.VirtualHost.TLS.SSLKeyCertificate.Name: "SslKeyCert",
		hostrule.Spec.VirtualHost.TLS.SSLProfile:             "SslProfile",
		hostrule.Spec.VirtualHost.AnalyticsProfile:           "AnalyticsProfile",
		hostrule.Spec.VirtualHost.ErrorPageProfile:           "ErrorPageProfile",
	}

	for _, policy := range hostrule.Spec.VirtualHost.HTTPPolicy.PolicySets {
		refData[policy] = "HttpPolicySet"
	}

	for _, script := range hostrule.Spec.VirtualHost.Datascripts {
		refData[script] = "VsDatascript"
	}

	for k, value := range refData {
		if k == "" {
			continue
		}

		if errStatus := checkRefOnController(key, value, k); errStatus != nil {
			status.UpdateHostRuleStatus(key, hostrule, status.UpdateCRDStatusOptions{
				Status: lib.StatusRejected,
				Error:  errStatus.Error(),
			})
			return errStatus
		}
	}
	status.UpdateHostRuleStatus(key, hostrule, status.UpdateCRDStatusOptions{
		Status: lib.StatusAccepted,
		Error:  "",
	})
	return nil
}

// validateHTTPRuleObj would do validation checks
// update internal CRD caches, and push relevant ingresses to ingestion
func validateHTTPRuleObj(key string, httprule *akov1alpha1.HTTPRule) error {
	refData := make(map[string]string)
	for _, path := range httprule.Spec.Paths {
		refData[path.TLS.SSLProfile] = "SslProfile"

		for _, hm := range path.HealthMonitors {
			refData[hm] = "HealthMonitor"
		}
	}

	for k, value := range refData {
		if k == "" {
			continue
		}

		if errStatus := checkRefOnController(key, value, k); errStatus != nil {
			status.UpdateHTTPRuleStatus(key, httprule, status.UpdateCRDStatusOptions{
				Status: lib.StatusRejected,
				Error:  errStatus.Error(),
			})
			return errStatus
		}
	}

	status.UpdateHTTPRuleStatus(key, httprule, status.UpdateCRDStatusOptions{
		Status: lib.StatusAccepted,
		Error:  "",
	})
	return nil
}

var refModelMap = map[string]string{
	"SslKeyCert":       "sslkeyandcertificate",
	"WafPolicy":        "wafpolicy",
	"HttpPolicySet":    "httppolicyset",
	"SslProfile":       "sslprofile",
	"AppProfile":       "applicationprofile",
	"AnalyticsProfile": "analyticsprofile",
	"ErrorPageProfile": "errorpageprofile",
	"VsDatascript":     "vsdatascriptset",
	"HealthMonitor":    "healthmonitor",
}

// checkRefOnController checks whether a provided ref on the controller
func checkRefOnController(key, refKey, refValue string) error {
	uri := fmt.Sprintf("/api/%s?name=%s&fields=name,type", refModelMap[refKey], refValue)
	clients := cache.SharedAVIClients()

	// assign the last avi client for ref checks
	aviClientLen := lib.GetshardSize()
	result, err := lib.AviGetCollectionRaw(clients.AviClient[aviClientLen], uri)
	if err != nil {
		utils.AviLog.Warnf("key: %s, msg: Get uri %v returned err %v", key, uri, err)
		return fmt.Errorf("%s \"%s\" not found on controller", refModelMap[refKey], refValue)
	}

	if result.Count == 0 {
		utils.AviLog.Warnf("key: %s, msg: No Objects found for refName: %s/%s", key, refModelMap[refKey], refValue)
		return fmt.Errorf("%s \"%s\" not found on controller", refModelMap[refKey], refValue)
	}

	if refKey == "AppProfile" {
		items := make([]json.RawMessage, result.Count)
		err = json.Unmarshal(result.Results, &items)
		if err != nil {
			utils.AviLog.Warnf("key: %s, msg: Failed to unmarshal data, err: %v", key, err)
			return fmt.Errorf("%s \"%s\" not found on controller", refModelMap[refKey], refValue)
		}

		appProf := models.ApplicationProfile{}
		err := json.Unmarshal(items[0], &appProf)
		if err != nil {
			utils.AviLog.Warnf("key: %s, msg: Failed to unmarshal data, err: %v", key, err)
			return fmt.Errorf("%s \"%s\" found on controller is invalid", refModelMap[refKey], refValue)
		}

		if *appProf.Type != lib.AllowedApplicationProfile {
			utils.AviLog.Warnf("key: %s, msg: applicationProfile: %s must be of type %s", key, refValue, lib.AllowedApplicationProfile)
			return fmt.Errorf("%s \"%s\" found on controller is invalid, must be of type: %s",
				refModelMap[refKey], refValue, lib.AllowedApplicationProfile)
		}
	}

	utils.AviLog.Infof("key: %s, msg: Ref found for %s/%s", key, refModelMap[refKey], refValue)
	return nil
}
