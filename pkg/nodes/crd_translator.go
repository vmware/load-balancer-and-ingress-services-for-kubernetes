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

	akov1alpha1 "ako/pkg/apis/ako/v1alpha1"
	"ako/pkg/cache"
	"ako/pkg/lib"
	"ako/pkg/objects"

	"github.com/avinetworks/container-lib/utils"
)

func BuildL7HostRule(host, namespace, ingName, key string, vsNode *AviVsNode) {
	// use host to find out HostRule CRD if it exists
	found, hrNamespaceName := objects.SharedCRDLister().GetFQDNToHostruleMapping(host)
	deleteCase := false
	if !found {
		utils.AviLog.Warnf("key: %s, msg: No HostRule found for virtualhost: %s in Cache", key, host)
		deleteCase = true
	}

	var err error
	var hrNSName []string
	var hostrule *akov1alpha1.HostRule
	if !deleteCase {
		hrNSName = strings.Split(hrNamespaceName, "/")
		// the hostrule can be present in any namespace therefore putting it blank here
		hostrule, err = lib.GetCRDInformers().HostRuleInformer.Lister().HostRules(hrNSName[0]).Get(hrNSName[1])
		if err != nil {
			utils.AviLog.Warnf("key: %s, msg: No HostRule found for virtualhost: %s msg: %v", key, host, err)
			deleteCase = true
		}
	}

	if deleteCase {
		vsNode.SSLKeyCertAviRef = ""
		vsNode.WafPolicyRef = ""
		vsNode.NsPolicyRef = ""
		vsNode.HttpPolicySetRefs = []string{}
		vsNode.AppProfileRef = ""
		if vsNode.ServiceMetadata.CRDStatus.Value != "" {
			vsNode.ServiceMetadata.CRDStatus.Status = "INACTIVE"
		}
		return
	}

	// host specific
	var vsWafPolicy, vsNSPolicy, vsAppProfile, vsSslKeyCertificate string
	var vsHTTPPolicySets []string
	sslKeyCertRef := hostrule.Spec.VirtualHost.TLS.SSLKeyCertificate.Name
	if sslKeyCertRef != "" {
		vsSslKeyCertificate = fmt.Sprintf("/api/sslkeyandcertificate?name=%s", sslKeyCertRef)
		vsNode.SSLKeyCertRefs = []*AviTLSKeyCertNode{}
	}

	wafPolicyRef := hostrule.Spec.VirtualHost.WAFPolicy
	if wafPolicyRef != "" {
		vsWafPolicy = fmt.Sprintf("/api/wafpolicy?name=%s", wafPolicyRef)
	}

	nsPolicyRef := hostrule.Spec.VirtualHost.NetworkSecurityPolicy
	if nsPolicyRef != "" {
		vsNSPolicy = fmt.Sprintf("/api/networksecuritypolicy?name=%s", nsPolicyRef)
	}

	httpPolicySetRef := hostrule.Spec.VirtualHost.HTTPPolicy.PolicySets
	for _, policy := range httpPolicySetRef {
		if !utils.HasElem(vsHTTPPolicySets, fmt.Sprintf("/api/httppolicyset?name=%s", policy)) {
			vsHTTPPolicySets = append(vsHTTPPolicySets, fmt.Sprintf("/api/httppolicyset?name=%s", policy))
		}
	}

	// delete all auto-created HttpPolicySets by AKO if override is set
	if hostrule.Spec.VirtualHost.HTTPPolicy.Overwrite {
		vsNode.HttpPolicyRefs = []*AviHttpPolicySetNode{}
	}

	appProfileRef := hostrule.Spec.VirtualHost.ApplicationProfile
	if appProfileRef != "" {
		vsAppProfile = fmt.Sprintf("/api/applicationprofile?name=%s", appProfileRef)
	}

	vsNode.SSLKeyCertAviRef = vsSslKeyCertificate
	vsNode.WafPolicyRef = vsWafPolicy
	vsNode.NsPolicyRef = vsNSPolicy
	vsNode.HttpPolicySetRefs = vsHTTPPolicySets
	vsNode.AppProfileRef = vsAppProfile
	vsNode.ServiceMetadata.CRDStatus = cache.CRDMetadata{
		Type:   "HostRule",
		Value:  hostrule.Namespace + "/" + hostrule.Name,
		Status: "ACTIVE",
	}

	utils.AviLog.Infof("key: %s, Attached hostrule %s on vsNode %s", key, host, vsNode.Name)
}

// BuildPoolHTTPRule notes
// when we get an ingress update and we are building the corresponding pools of that ingress
// we need to get all httprules which match ingress's host/path
func BuildPoolHTTPRule(host, path, ingName, namespace, key string, vsNode *AviVsNode, isSNI bool) {
	found, hrNamespaceName := objects.SharedCRDLister().GetFQDNToHostruleMapping(host)
	deleteCase := false
	if !found {
		utils.AviLog.Warnf("key: %s, msg: No HostRule found for virtualhost: %s in Cache", key, host)
		deleteCase = true
	}

	var pathRules map[string]string
	if !deleteCase {
		found, pathRules = objects.SharedCRDLister().GetHostHTTPRulesMapping(hrNamespaceName)
		if !found {
			utils.AviLog.Warnf("key: %s, msg: HTTPRules for hostrule %s not found", key, hrNamespaceName)
			deleteCase = true
		}
	}

	if deleteCase {
		for _, pool := range vsNode.PoolRefs {
			pool.SniEnabled = false
			pool.SslProfileRef = ""
			pool.ClientCertRef = ""
			pool.PkiProfileRef = ""
			pool.LbAlgorithm = ""
			pool.LbAlgorithmHash = ""
			if pool.ServiceMetadata.CRDStatus.Value != "" {
				pool.ServiceMetadata.CRDStatus.Status = "INACTIVE"
			}
		}
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
			utils.AviLog.Warnf("key: %s, msg: httprule not found err: %+v", key, err)
			continue
		}
		for _, path := range httpRuleObj.Spec.Paths {
			httpruleNameObjMap[httprule+path.Target] = path
		}
	}

	// iterate through httpRule which we get from GetHostHTTPRulesMapping
	// must contain hr1: {path1: rr1, path2: rr1, path3: rr2}
	for path, rule := range pathRules {
		rrNamespace := strings.Split(rule, "/")[0]
		httpRulePath, ok := httpruleNameObjMap[rule+path]
		if !ok {
			continue
		}
		for _, pool := range vsNode.PoolRefs {
			isPathSniEnabled := false
			var pathSslProfile, pathClientCert, pathPkiProfile string

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
				sslProfileRef := httpRulePath.TLS.SSLProfile
				if sslProfileRef != "" {
					isPathSniEnabled = true
					pathSslProfile = fmt.Sprintf("/api/sslprofile?name=%s", sslProfileRef)
				}

				clientCertRef := httpRulePath.TLS.ClientCertificate
				if clientCertRef != "" && isPathSniEnabled {
					pathClientCert = fmt.Sprintf("/api/sslkeyandcertificate?name=%s", clientCertRef)
				}

				pkiProfileRef := httpRulePath.TLS.PkiProfile
				if pkiProfileRef != "" && isPathSniEnabled {
					pathPkiProfile = fmt.Sprintf("/api/pkiprofile?name=%s", pkiProfileRef)
				}

				pool.SniEnabled = isPathSniEnabled
				if isPathSniEnabled {
					pool.SniEnabled = isPathSniEnabled
					pool.SslProfileRef = pathSslProfile
					pool.ClientCertRef = pathClientCert
					pool.PkiProfileRef = pathPkiProfile
				}

				// from this path, generate refs to this pool node
				pool.LbAlgorithm = httpRulePath.LoadBalancerPolicy.Algorithm
				pool.LbAlgorithmHash = httpRulePath.LoadBalancerPolicy.Hash
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
