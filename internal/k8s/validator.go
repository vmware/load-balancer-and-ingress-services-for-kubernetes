/*
 * Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
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

package k8s

import (
	"context"
	"fmt"
	"net"
	"regexp"
	"strings"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/status"
	akov1alpha1 "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/apis/ako/v1alpha1"
	akov1alpha2 "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/apis/ako/v1alpha2"
	akov1beta1 "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/apis/ako/v1beta1"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Validator interface {
	ValidateHTTPRuleObj(key string, httprule *akov1beta1.HTTPRule) error
	ValidateHostRuleObj(key string, hostrule *akov1beta1.HostRule) error
	ValidateAviInfraSetting(key string, infraSetting *akov1beta1.AviInfraSetting) error
	ValidateMultiClusterIngressObj(key string, multiClusterIngress *akov1alpha1.MultiClusterIngress) error
	ValidateServiceImportObj(key string, serviceImport *akov1alpha1.ServiceImport) error
	ValidateSSORuleObj(key string, ssoRule *akov1alpha2.SSORule) error
	ValidateL4RuleObj(key string, l4Rule *akov1alpha2.L4Rule) error
	ValidateL7RuleObj(key string, l7Rule *akov1alpha2.L7Rule) error
}

type (
	follower struct{}
	leader   struct{}
)

func NewValidator() Validator {
	if lib.AKOControlConfig().IsLeader() {
		return &leader{}
	}
	return &follower{}
}

// validateHostRuleObj would do validation checks
// update internal CRD caches, and push relevant ingresses to ingestion
func (l *leader) ValidateHostRuleObj(key string, hostrule *akov1beta1.HostRule) error {

	var err error
	fqdn := hostrule.Spec.VirtualHost.Fqdn
	foundHost, foundHR := objects.SharedCRDLister().GetFQDNToHostruleMapping(fqdn)
	if foundHost && foundHR != hostrule.Namespace+"/"+hostrule.Name {
		err = fmt.Errorf("duplicate fqdn %s found in %s", fqdn, foundHR)
		status.UpdateHostRuleStatus(key, hostrule, status.UpdateCRDStatusOptions{Status: lib.StatusRejected, Error: err.Error()})
		return err
	}

	// If it is not a Shared VS but TCP Settings are provided, then we reject it since these
	// TCP settings are not valid for the child VS.
	// TODO: move to translator?
	// if !strings.Contains(fqdn, lib.ShardVSSubstring) && hostrule.Spec.VirtualHost.TCPSettings != nil {
	// 	err = fmt.Errorf("Hostrule tcpSettings with fqdn %s cannot be applied to child Virtualservices", fqdn)
	// 	status.UpdateHostRuleStatus(key, hostrule, status.UpdateCRDStatusOptions{Status: lib.StatusRejected, Error: err.Error()})
	// 	return err
	// }

	if hostrule.Spec.VirtualHost.TCPSettings != nil && hostrule.Spec.VirtualHost.TCPSettings.LoadBalancerIP != "" {
		re := regexp.MustCompile(lib.IPRegex)
		if !re.MatchString(hostrule.Spec.VirtualHost.TCPSettings.LoadBalancerIP) {
			err = fmt.Errorf("loadBalancerIP %s is not a valid IP", hostrule.Spec.VirtualHost.TCPSettings.LoadBalancerIP)
			status.UpdateHostRuleStatus(key, hostrule, status.UpdateCRDStatusOptions{Status: lib.StatusRejected, Error: err.Error()})
			return err
		}
	}

	if hostrule.Spec.VirtualHost.Gslb.Fqdn != "" {
		if fqdn == hostrule.Spec.VirtualHost.Gslb.Fqdn {
			err = fmt.Errorf("GSLB FQDN and local FQDN are same")
			status.UpdateHostRuleStatus(key, hostrule, status.UpdateCRDStatusOptions{Status: lib.StatusRejected, Error: err.Error()})
			return err
		}
	}

	if hostrule.Spec.VirtualHost.TCPSettings != nil && len(hostrule.Spec.VirtualHost.TCPSettings.Listeners) > 0 {
		sslEnabled := false
		for _, listener := range hostrule.Spec.VirtualHost.TCPSettings.Listeners {
			if listener.EnableSSL {
				sslEnabled = true
				break
			}
		}
		if !sslEnabled {
			err = fmt.Errorf("Hosting parent virtualservice must have SSL enabled")
			status.UpdateHostRuleStatus(key, hostrule, status.UpdateCRDStatusOptions{Status: lib.StatusRejected, Error: err.Error()})
			return err
		}
	}

	if hostrule.Spec.VirtualHost.Aliases != nil {
		if hostrule.Spec.VirtualHost.FqdnType != akov1beta1.Exact {
			err = fmt.Errorf("Aliases is supported only when FQDN type is set as Exact")
			status.UpdateHostRuleStatus(key, hostrule, status.UpdateCRDStatusOptions{Status: lib.StatusRejected, Error: err.Error()})
			return err
		}

		if utils.HasElem(hostrule.Spec.VirtualHost.Aliases, fqdn) {
			err = fmt.Errorf("Duplicate entry found. Aliases field has same entry as the FQDN field")
			status.UpdateHostRuleStatus(key, hostrule, status.UpdateCRDStatusOptions{Status: lib.StatusRejected, Error: err.Error()})
			return err
		}

		if utils.ContainsDuplicate(hostrule.Spec.VirtualHost.Aliases) {
			err = fmt.Errorf("Aliases must be unique")
			status.UpdateHostRuleStatus(key, hostrule, status.UpdateCRDStatusOptions{Status: lib.StatusRejected, Error: err.Error()})
			return err
		}

		if hostrule.Spec.VirtualHost.Gslb.Fqdn != "" &&
			utils.HasElem(hostrule.Spec.VirtualHost.Aliases, hostrule.Spec.VirtualHost.Gslb.Fqdn) {
			err = fmt.Errorf("Aliases must not contain GSLB FQDN")
			status.UpdateHostRuleStatus(key, hostrule, status.UpdateCRDStatusOptions{Status: lib.StatusRejected, Error: err.Error()})
			return err
		}

		for cachedFQDN, cachedAliases := range objects.SharedCRDLister().GetAllFQDNToAliasesMapping() {
			if cachedFQDN == fqdn {
				continue
			}
			aliases := cachedAliases.([]string)
			for _, alias := range hostrule.Spec.VirtualHost.Aliases {
				if utils.HasElem(aliases, alias) {
					err = fmt.Errorf("%s is already in use by hostrule %s", alias, cachedFQDN)
					status.UpdateHostRuleStatus(key, hostrule, status.UpdateCRDStatusOptions{Status: lib.StatusRejected, Error: err.Error()})
					return err
				}
			}
		}
	}

	refData := map[string]string{
		hostrule.Spec.VirtualHost.WAFPolicy:          "WafPolicy",
		hostrule.Spec.VirtualHost.ApplicationProfile: "AppProfile",
		hostrule.Spec.VirtualHost.TLS.SSLProfile:     "SslProfile",
		hostrule.Spec.VirtualHost.AnalyticsProfile:   "AnalyticsProfile",
		hostrule.Spec.VirtualHost.ErrorPageProfile:   "ErrorPageProfile",
	}
	if hostrule.Spec.VirtualHost.TLS.SSLKeyCertificate.Type == akov1beta1.HostRuleSecretTypeAviReference {
		refData[hostrule.Spec.VirtualHost.TLS.SSLKeyCertificate.Name] = "SslKeyCert"
	}

	if hostrule.Spec.VirtualHost.TLS.SSLKeyCertificate.Type == akov1beta1.HostRuleSecretTypeSecretReference {
		secretName := hostrule.Spec.VirtualHost.TLS.SSLKeyCertificate.Name
		err := validateSecretReferenceInHostrule(hostrule.Namespace, secretName)
		if err != nil {
			status.UpdateHostRuleStatus(key, hostrule, status.UpdateCRDStatusOptions{Status: lib.StatusRejected, Error: err.Error()})
			return err
		}
	}
	if hostrule.Spec.VirtualHost.TLS.SSLKeyCertificate.AlternateCertificate.Type == akov1beta1.HostRuleSecretTypeAviReference {
		refData[hostrule.Spec.VirtualHost.TLS.SSLKeyCertificate.AlternateCertificate.Name] = "SslKeyCert"
	}

	if hostrule.Spec.VirtualHost.TLS.SSLKeyCertificate.AlternateCertificate.Type == akov1beta1.HostRuleSecretTypeSecretReference {
		secretName := hostrule.Spec.VirtualHost.TLS.SSLKeyCertificate.AlternateCertificate.Name
		err := validateSecretReferenceInHostrule(hostrule.Namespace, secretName)
		if err != nil {
			status.UpdateHostRuleStatus(key, hostrule, status.UpdateCRDStatusOptions{Status: lib.StatusRejected, Error: err.Error()})
			return err
		}
	}
	if len(hostrule.Spec.VirtualHost.ICAPProfile) > 1 {
		status.UpdateHostRuleStatus(key, hostrule, status.UpdateCRDStatusOptions{Status: lib.StatusRejected, Error: "Can only have 1 ICAP profile associated with VS"})
		return fmt.Errorf("Can only have 1 ICAP profile associated with VS")
	} else {
		for _, icapprofile := range hostrule.Spec.VirtualHost.ICAPProfile {
			refData[icapprofile] = "ICAPProfile"
		}
	}

	for _, policy := range hostrule.Spec.VirtualHost.HTTPPolicy.PolicySets {
		refData[policy] = "HttpPolicySet"
	}

	for _, script := range hostrule.Spec.VirtualHost.Datascripts {
		refData[script] = "VsDatascript"
	}

	// Validation for Network Security Policy
	// Check networkSecurityPolicy is of type ref.
	if hostrule.Spec.VirtualHost.NetworkSecurityPolicy != "" {
		refData[hostrule.Spec.VirtualHost.NetworkSecurityPolicy] = "NetworkSecurityPolicy"
	}
	tenant := lib.GetTenantInNamespace(hostrule.Namespace)

	if err := checkRefsOnController(key, refData, tenant); err != nil {
		status.UpdateHostRuleStatus(key, hostrule, status.UpdateCRDStatusOptions{Status: lib.StatusRejected, Error: err.Error()})
		return err
	}

	if hostrule.Spec.VirtualHost.L7Rule != "" {
		objects.SharedCRDLister().UpdateL7RuleToHostRuleMapping(hostrule.Namespace+"/"+hostrule.Spec.VirtualHost.L7Rule, hostrule.Name)
		_, err := lib.AKOControlConfig().CRDInformers().L7RuleInformer.Lister().L7Rules(hostrule.Namespace).Get(hostrule.Spec.VirtualHost.L7Rule)
		if err != nil {
			status.UpdateHostRuleStatus(key, hostrule, status.UpdateCRDStatusOptions{Status: lib.StatusRejected, Error: err.Error()})
			return err
		}
	}

	if strings.Contains(fqdn, lib.ShardVSSubstring) && hostrule.Spec.VirtualHost.UseRegex {
		err = fmt.Errorf("hostrule useRegex with fqdn %s cannot be applied to shared virtualservices", fqdn)
		status.UpdateHostRuleStatus(key, hostrule, status.UpdateCRDStatusOptions{Status: lib.StatusRejected, Error: err.Error()})
		return err
	}

	if strings.Contains(fqdn, lib.ShardVSSubstring) && hostrule.Spec.VirtualHost.ApplicationRootPath != "" {
		err = fmt.Errorf("hostrule applicationRootPath with fqdn %s cannot be applied to shared virtualservices", fqdn)
		status.UpdateHostRuleStatus(key, hostrule, status.UpdateCRDStatusOptions{Status: lib.StatusRejected, Error: err.Error()})
		return err
	}

	// No need to update status of hostrule object as accepted since it was accepted before.
	if hostrule.Status.Status == lib.StatusAccepted {
		return nil
	}

	status.UpdateHostRuleStatus(key, hostrule, status.UpdateCRDStatusOptions{Status: lib.StatusAccepted, Error: ""})
	return nil
}

func validateSecretReferenceInHostrule(namespace, secretName string) error {

	// reject the hostrule if the secret handling is restricted to the namespace where
	// AKO is installed.
	if utils.GetInformers().RouteInformer != nil &&
		namespace != utils.GetAKONamespace() &&
		utils.IsSecretsHandlingRestrictedToAKONS() {
		err := fmt.Errorf("secret handling is restricted to %s namespace only", utils.GetAKONamespace())
		return err
	}

	_, err := utils.GetInformers().SecretInformer.Lister().Secrets(namespace).Get(secretName)
	return err
}

func validateSecretReferenceInSSORule(namespace, secretName string) (*v1.Secret, error) {

	// reject the SSORule if the secret handling is restricted to the namespace where
	// AKO is installed.
	if utils.GetInformers().RouteInformer != nil &&
		namespace != utils.GetAKONamespace() &&
		utils.IsSecretsHandlingRestrictedToAKONS() {
		err := fmt.Errorf("secret handling is restricted to %s namespace only", utils.GetAKONamespace())
		return nil, err
	}

	secretObj, err := utils.GetInformers().SecretInformer.Lister().Secrets(namespace).Get(secretName)
	return secretObj, err
}

// validateHTTPRuleObj would do validation checks
// update internal CRD caches, and push relevant ingresses to ingestion
func (l *leader) ValidateHTTPRuleObj(key string, httprule *akov1beta1.HTTPRule) error {

	refData := make(map[string]string)
	for _, path := range httprule.Spec.Paths {
		if path.TLS.PKIProfile != "" && path.TLS.DestinationCA != "" {
			//if both pkiProfile and destCA set, reject httprule
			status.UpdateHTTPRuleStatus(key, httprule, status.UpdateCRDStatusOptions{
				Status: lib.StatusRejected,
				Error:  lib.HttpRulePkiAndDestCASetErr,
			})
			return fmt.Errorf("key: %s, msg: %s", key, lib.HttpRulePkiAndDestCASetErr)
		}
		refData[path.TLS.SSLProfile] = "SslProfile"
		refData[path.ApplicationPersistence] = "ApplicationPersistence"
		if path.TLS.PKIProfile != "" {
			refData[path.TLS.PKIProfile] = "PKIProfile"
		}

		for _, hm := range path.HealthMonitors {
			refData[hm] = "HealthMonitor"
		}
	}
	tenant := lib.GetTenantInNamespace(httprule.Namespace)

	if err := checkRefsOnController(key, refData, tenant); err != nil {
		status.UpdateHTTPRuleStatus(key, httprule, status.UpdateCRDStatusOptions{
			Status: lib.StatusRejected,
			Error:  err.Error(),
		})
		return err
	}

	// No need to update status of httprule object as accepted since it was accepted before.
	if httprule.Status.Status == lib.StatusAccepted {
		return nil
	}

	status.UpdateHTTPRuleStatus(key, httprule, status.UpdateCRDStatusOptions{
		Status: lib.StatusAccepted,
		Error:  "",
	})
	return nil
}

// validateAviInfraSetting would do validaion checks on the
// ingested AviInfraSetting objects
func (l *leader) ValidateAviInfraSetting(key string, infraSetting *akov1beta1.AviInfraSetting) error {

	if ((infraSetting.Spec.Network.EnableRhi != nil && !*infraSetting.Spec.Network.EnableRhi) || infraSetting.Spec.Network.EnableRhi == nil) &&
		len(infraSetting.Spec.Network.BgpPeerLabels) > 0 {
		err := fmt.Errorf("BGPPeerLabels cannot be set if EnableRhi is false.")
		status.UpdateAviInfraSettingStatus(key, infraSetting, status.UpdateCRDStatusOptions{
			Status: lib.StatusRejected,
			Error:  err.Error(),
		})
		return err
	}

	refData := make(map[string]string)
	for _, vipNetwork := range infraSetting.Spec.Network.VipNetworks {
		if vipNetwork.Cidr != "" {
			re := regexp.MustCompile(lib.IPCIDRRegex)
			if !re.MatchString(vipNetwork.Cidr) {
				err := fmt.Errorf("invalid CIDR configuration %s detected for networkName %s in vipNetworkList", vipNetwork.Cidr, vipNetwork.NetworkName)
				status.UpdateAviInfraSettingStatus(key, infraSetting, status.UpdateCRDStatusOptions{
					Status: lib.StatusRejected,
					Error:  err.Error(),
				})
				return err
			}
		}
		if vipNetwork.V6Cidr != "" {
			re := regexp.MustCompile(lib.IPV6CIDRRegex)
			if !re.MatchString(vipNetwork.V6Cidr) {
				err := fmt.Errorf("invalid IPv6 CIDR configuration %s detected for networkName %s in vipNetworkList", vipNetwork.V6Cidr, vipNetwork.NetworkName)
				status.UpdateAviInfraSettingStatus(key, infraSetting, status.UpdateCRDStatusOptions{
					Status: lib.StatusRejected,
					Error:  err.Error(),
				})
				return err
			}
		}
		// Give preference to network uuid
		if vipNetwork.NetworkUUID != "" {
			refData[vipNetwork.NetworkUUID] = "NetworkUUID"
		} else if vipNetwork.NetworkName != "" {
			refData[vipNetwork.NetworkName] = "Network"
		}
	}

	// Node network validation
	for _, nodeNetwork := range infraSetting.Spec.Network.NodeNetworks {
		if nodeNetwork.NetworkUUID != "" {
			refData[nodeNetwork.NetworkUUID] = "NetworkUUID"
		} else if nodeNetwork.NetworkName != "" {
			refData[nodeNetwork.NetworkName] = "Network"
		}
	}
	if infraSetting.Spec.SeGroup.Name != "" {
		refData[infraSetting.Spec.SeGroup.Name] = "ServiceEngineGroup"
	}
	if len(infraSetting.Spec.Network.Listeners) > 0 {
		sslEnabled := false
		for _, listener := range infraSetting.Spec.Network.Listeners {
			if listener.EnableSSL != nil && *listener.EnableSSL {
				sslEnabled = true
				break
			}
		}
		if !sslEnabled {
			err := fmt.Errorf("One of the port in aviInfraSetting must have SSL enabled")
			status.UpdateAviInfraSettingStatus(key, infraSetting, status.UpdateCRDStatusOptions{
				Status: lib.StatusRejected,
				Error:  err.Error(),
			})
			return err
		}
	}
	if err := checkRefsOnController(key, refData, lib.GetTenant()); err != nil {
		status.UpdateAviInfraSettingStatus(key, infraSetting, status.UpdateCRDStatusOptions{
			Status: lib.StatusRejected,
			Error:  err.Error(),
		})
		return err
	}

	// This would add SEG labels only if they are not configured yet. In case there is a label mismatch
	// to any pre-existing SEG labels, the AviInfraSettig CR will get Rejected from the checkRefsOnController
	// step before this.
	segMgmtNetworK := ""
	if infraSetting.Spec.SeGroup.Name != "" {
		addSeGroupLabel(key, infraSetting.Spec.SeGroup.Name)
		// Not required for NO access cloud
		if lib.GetCloudType() == lib.CLOUD_VCENTER {
			segMgmtNetworK = GetSEGManagementNetwork(infraSetting.Spec.SeGroup.Name)
		}
	}

	if len(infraSetting.Spec.Network.VipNetworks) > 0 {
		SetAviInfrasettingVIPNetworks(infraSetting.Name, segMgmtNetworK, infraSetting.Spec.SeGroup.Name, infraSetting.Spec.Network.VipNetworks)
	}

	if len(infraSetting.Spec.Network.NodeNetworks) > 0 {
		SetAviInfrasettingNodeNetworks(infraSetting.Name, segMgmtNetworK, infraSetting.Spec.SeGroup.Name, infraSetting.Spec.Network.NodeNetworks)
	}

	namespaces, err := utils.GetInformers().NSInformer.Informer().GetIndexer().ByIndex(lib.AviSettingNamespaceIndex, infraSetting.GetName())
	if err == nil && len(namespaces) > 0 {
		objects.InfraSettingL7Lister().UpdateInfraSettingToNamespaceMapping(infraSetting.GetName(), namespaces)
	} else {
		// This handles the case where an NS scoped infrasetting was deleted and later recreated without NS scope.
		objects.InfraSettingL7Lister().DeleteInfraSettingToNamespaceMapping(infraSetting.GetName())
	}

	// No need to update status of infra setting object as accepted since it was accepted before.
	if infraSetting.Status.Status == lib.StatusAccepted {
		return nil
	}

	status.UpdateAviInfraSettingStatus(key, infraSetting, status.UpdateCRDStatusOptions{
		Status: lib.StatusAccepted,
		Error:  "",
	})
	return nil
}

// validateMultiClusterIngressObj validates the MCI CRD changes before pushing it to ingestion
func (l *leader) ValidateMultiClusterIngressObj(key string, multiClusterIngress *akov1alpha1.MultiClusterIngress) error {

	var err error
	statusToUpdate := &akov1alpha1.MultiClusterIngressStatus{}
	defer func() {
		if err == nil {
			statusToUpdate.Status.Accepted = true
			status.UpdateMultiClusterIngressStatus(key, multiClusterIngress, statusToUpdate)
			return
		}
		statusToUpdate.Status.Accepted = false
		statusToUpdate.Status.Reason = err.Error()
		status.UpdateMultiClusterIngressStatus(key, multiClusterIngress, statusToUpdate)
	}()

	// Currently, we support only NodePort ServiceType.
	if !lib.IsNodePortMode() {
		err = fmt.Errorf("ServiceType must be of type NodePort")
		return err
	}

	// Currently, we support EVH mode only.
	if !lib.IsEvhEnabled() {
		err = fmt.Errorf("AKO must be in EVH mode")
		return err
	}

	if len(multiClusterIngress.Spec.Config) == 0 {
		err = fmt.Errorf("config must not be empty")
		return err
	}

	return nil
}

// validateServiceImportObj validates the SI CRD changes before pushing it to ingestion
func (l *leader) ValidateServiceImportObj(key string, serviceImport *akov1alpha1.ServiceImport) error {

	// CHECK ME: AMKO creates this and validation required?
	// TODO: validations needs a status field

	return nil
}

// ValidateSSORuleObj would do validation checks
// update internal CRD caches, and push relevant ingresses to ingestion
func (l *leader) ValidateSSORuleObj(key string, ssoRule *akov1alpha2.SSORule) error {
	var err error
	fqdn := *ssoRule.Spec.Fqdn
	foundHost, foundSR := objects.SharedCRDLister().GetFQDNToSSORuleMapping(fqdn)
	if foundHost && foundSR != ssoRule.Namespace+"/"+ssoRule.Name {
		err = fmt.Errorf("duplicate fqdn %s found in %s", fqdn, foundSR)
		status.UpdateSSORuleStatus(key, ssoRule, status.UpdateCRDStatusOptions{Status: lib.StatusRejected, Error: err.Error()})
		return err
	}

	refData := make(map[string]string)

	if ssoRule.Spec.SsoPolicyRef == nil {
		err = fmt.Errorf("SsoPolicyRef is not specified")
		status.UpdateSSORuleStatus(key, ssoRule, status.UpdateCRDStatusOptions{Status: lib.StatusRejected, Error: err.Error()})
		return err
	}
	refData[*ssoRule.Spec.SsoPolicyRef] = "SSOPolicy"

	if ssoRule.Spec.OauthVsConfig != nil {
		oauthConfigObj := ssoRule.Spec.OauthVsConfig

		if len(oauthConfigObj.OauthSettings) != 0 {
			for _, profile := range oauthConfigObj.OauthSettings {
				refData[*profile.AuthProfileRef] = "AuthProfile"

				if profile.AppSettings != nil {
					clientSecret := *profile.AppSettings.ClientSecret
					clientSecretObj, err := validateSecretReferenceInSSORule(ssoRule.Namespace, clientSecret)
					if err != nil {
						err = fmt.Errorf("Got error while fetching %s secret : %s", clientSecret, err.Error())
						status.UpdateSSORuleStatus(key, ssoRule, status.UpdateCRDStatusOptions{Status: lib.StatusRejected, Error: err.Error()})
						return err
					}
					if clientSecretObj == nil {
						err = fmt.Errorf("specified client secret is empty : %s", clientSecret)
						status.UpdateSSORuleStatus(key, ssoRule, status.UpdateCRDStatusOptions{Status: lib.StatusRejected, Error: err.Error()})
						return err
					}
					clientSecretString := string(clientSecretObj.Data["clientSecret"])
					if clientSecretString == "" {
						err = fmt.Errorf("clientSecret field not found in %s secret", clientSecret)
						status.UpdateSSORuleStatus(key, ssoRule, status.UpdateCRDStatusOptions{Status: lib.StatusRejected, Error: err.Error()})
						return err
					}
				}

				if profile.ResourceServer != nil {
					if *profile.ResourceServer.AccessType == lib.ACCESS_TOKEN_TYPE_JWT && profile.ResourceServer.JwtParams == nil {
						err = fmt.Errorf("Access Type is %s, but Jwt Params have not been specified", *profile.ResourceServer.AccessType)
						status.UpdateSSORuleStatus(key, ssoRule, status.UpdateCRDStatusOptions{Status: lib.StatusRejected, Error: err.Error()})
						return err
					}
					if *profile.ResourceServer.AccessType == lib.ACCESS_TOKEN_TYPE_OPAQUE && profile.ResourceServer.OpaqueTokenParams == nil {
						err = fmt.Errorf("Access Type is %s, but Opaque Token Params have not been specified", *profile.ResourceServer.AccessType)
						status.UpdateSSORuleStatus(key, ssoRule, status.UpdateCRDStatusOptions{Status: lib.StatusRejected, Error: err.Error()})
						return err
					}

					if profile.ResourceServer.OpaqueTokenParams != nil {
						serverSecret := *profile.ResourceServer.OpaqueTokenParams.ServerSecret
						serverSecretObj, err := utils.GetInformers().ClientSet.CoreV1().Secrets(ssoRule.Namespace).Get(context.TODO(), serverSecret, metav1.GetOptions{})
						if err != nil {
							err = fmt.Errorf("Got error while fetching %s secret : %s", serverSecret, err.Error())
							status.UpdateSSORuleStatus(key, ssoRule, status.UpdateCRDStatusOptions{Status: lib.StatusRejected, Error: err.Error()})
							return err
						}
						if serverSecretObj == nil {
							err = fmt.Errorf("specified server secret is empty : %s", serverSecret)
							status.UpdateSSORuleStatus(key, ssoRule, status.UpdateCRDStatusOptions{Status: lib.StatusRejected, Error: err.Error()})
							return err
						}
						serverSecretString := string(serverSecretObj.Data["serverSecret"])
						if serverSecretString == "" {
							err = fmt.Errorf("serverSecret field not found in %s secret", serverSecret)
							status.UpdateSSORuleStatus(key, ssoRule, status.UpdateCRDStatusOptions{Status: lib.StatusRejected, Error: err.Error()})
							return err
						}
					}
				}
			}
		}
	}
	if ssoRule.Spec.SamlSpConfig != nil {
		samlConfigObj := ssoRule.Spec.SamlSpConfig

		if samlConfigObj.SigningSslKeyAndCertificateRef != nil {
			refData[*samlConfigObj.SigningSslKeyAndCertificateRef] = "SslKeyCert"
		}
	}
	tenant := lib.GetTenantInNamespace(ssoRule.Namespace)

	if err := checkRefsOnController(key, refData, tenant); err != nil {
		status.UpdateSSORuleStatus(key, ssoRule, status.UpdateCRDStatusOptions{Status: lib.StatusRejected, Error: err.Error()})
		return err
	}

	// No need to update status of ssoRule object as accepted since it was accepted before.
	if ssoRule.Status.Status == lib.StatusAccepted {
		return nil
	}

	status.UpdateSSORuleStatus(key, ssoRule, status.UpdateCRDStatusOptions{Status: lib.StatusAccepted, Error: ""})
	return nil
}

// ValidateL4RuleObj would do validation checks and updates the status before
// pushing to ingestion
func (l *leader) ValidateL4RuleObj(key string, l4Rule *akov1alpha2.L4Rule) error {

	l4RuleSpec := l4Rule.Spec

	if l4RuleSpec.LoadBalancerIP != nil &&
		net.ParseIP(*l4RuleSpec.LoadBalancerIP) == nil {
		err := fmt.Errorf("loadBalancerIP %s is not valid", *l4RuleSpec.LoadBalancerIP)
		status.UpdateL4RuleStatus(key, l4Rule, status.UpdateCRDStatusOptions{
			Status: lib.StatusRejected,
			Error:  err.Error(),
		})
		return err
	}

	refData := make(map[string]string)

	if l4RuleSpec.AnalyticsProfileRef != nil {
		refData[*l4RuleSpec.AnalyticsProfileRef] = "AnalyticsProfile"
	}

	var isNetworkProfileTypeTCP bool
	if l4RuleSpec.ApplicationProfileRef != nil {
		isSSLEnabled := false
		for _, svc := range l4RuleSpec.Services {
			if *svc.EnableSsl {
				isSSLEnabled = true
			}
		}
		isL4SSL, err := checkForL4SSLAppProfile(key, *l4RuleSpec.ApplicationProfileRef)
		if err != nil {
			status.UpdateL4RuleStatus(key, l4Rule, status.UpdateCRDStatusOptions{
				Status: lib.StatusRejected,
				Error:  err.Error(),
			})
			return err
		}
		if isL4SSL {
			if !isSSLEnabled {
				sslErr := fmt.Errorf("SSL is not enabled in l4rule listener Spec but App Profile %s is of type SSL", *l4RuleSpec.ApplicationProfileRef)
				status.UpdateL4RuleStatus(key, l4Rule, status.UpdateCRDStatusOptions{
					Status: lib.StatusRejected,
					Error:  sslErr.Error(),
				})
				return sslErr
			}
			if l4RuleSpec.SslProfileRef != nil {
				refData[*l4RuleSpec.SslProfileRef] = "SslProfile"
			}
			for _, ref := range l4RuleSpec.SslKeyAndCertificateRefs {
				refData[ref] = "SslKeyCert"
			}
			if l4RuleSpec.NetworkProfileRef != nil {
				isNetworkProfileTypeTCP, err = checkForNetworkProfileTypeTCP(key, *l4RuleSpec.NetworkProfileRef)
				if err != nil {
					status.UpdateL4RuleStatus(key, l4Rule, status.UpdateCRDStatusOptions{
						Status: lib.StatusRejected,
						Error:  err.Error(),
					})
					return err
				}
			}
		} else {
			if *l4RuleSpec.ApplicationProfileRef != utils.DEFAULT_L4_APP_PROFILE {
				if isSSLEnabled {
					sslErr := fmt.Errorf("SSL is enabled in l4rule listener Spec but App Profile %s is not of type SSL", *l4RuleSpec.ApplicationProfileRef)
					status.UpdateL4RuleStatus(key, l4Rule, status.UpdateCRDStatusOptions{
						Status: lib.StatusRejected,
						Error:  sslErr.Error(),
					})
					return sslErr
				}
			}
			if l4RuleSpec.SslProfileRef != nil {
				sslProfileErr := fmt.Errorf("App Profile %s is not of type SSL but SslProfileRef is set", *l4RuleSpec.ApplicationProfileRef)
				status.UpdateL4RuleStatus(key, l4Rule, status.UpdateCRDStatusOptions{
					Status: lib.StatusRejected,
					Error:  sslProfileErr.Error(),
				})
				return sslProfileErr
			}
			if len(l4RuleSpec.SslKeyAndCertificateRefs) != 0 {
				sslKeyCertErr := fmt.Errorf("App Profile %s is not of type SSL but SslKeyAndCertificateRefs are set", *l4RuleSpec.ApplicationProfileRef)
				status.UpdateL4RuleStatus(key, l4Rule, status.UpdateCRDStatusOptions{
					Status: lib.StatusRejected,
					Error:  sslKeyCertErr.Error(),
				})
				return sslKeyCertErr
			}
		}
	}

	if l4RuleSpec.NetworkProfileRef != nil && !isNetworkProfileTypeTCP {
		refData[*l4RuleSpec.NetworkProfileRef] = "NetworkProfile"
	}

	if l4RuleSpec.NetworkSecurityPolicyRef != nil {
		refData[*l4RuleSpec.NetworkSecurityPolicyRef] = "NetworkSecurityPolicy"
	}

	if l4RuleSpec.SecurityPolicyRef != nil {
		refData[*l4RuleSpec.SecurityPolicyRef] = "SecurityPolicy"
	}

	for _, ref := range l4RuleSpec.VsDatascriptRefs {
		refData[ref] = "VsDatascript"
	}

	for _, backendProperties := range l4RuleSpec.BackendProperties {

		if backendProperties.ApplicationPersistenceProfileRef != nil {
			refData[*backendProperties.ApplicationPersistenceProfileRef] = "ApplicationPersistence"
		}

		for _, hm := range backendProperties.HealthMonitorRefs {
			refData[hm] = "HealthMonitor"
		}

		if backendProperties.PkiProfileRef != nil {
			refData[*backendProperties.PkiProfileRef] = "PKIProfile"
		}

		if backendProperties.SslKeyAndCertificateRef != nil {
			refData[*backendProperties.SslKeyAndCertificateRef] = "SslKeyCert"
		}

		if backendProperties.SslProfileRef != nil {
			refData[*backendProperties.SslProfileRef] = "SslProfile"
		}

		if err := validateLBAlgorithm(backendProperties); err != nil {
			status.UpdateL4RuleStatus(key, l4Rule, status.UpdateCRDStatusOptions{
				Status: lib.StatusRejected,
				Error:  err.Error(),
			})
			return err
		}
	}
	tenant := lib.GetTenantInNamespace(l4Rule.Namespace)
	if err := checkRefsOnController(key, refData, tenant); err != nil {
		status.UpdateL4RuleStatus(key, l4Rule, status.UpdateCRDStatusOptions{
			Status: lib.StatusRejected,
			Error:  err.Error(),
		})
		return err
	}

	// No need to update status of l4rule object as accepted since it was accepted before.
	if l4Rule.Status.Status == lib.StatusAccepted {
		return nil
	}

	status.UpdateL4RuleStatus(key, l4Rule, status.UpdateCRDStatusOptions{
		Status: lib.StatusAccepted,
		Error:  "",
	})

	return nil
}

// ValidateL7RuleObj would do validation checks and updates the status before
// pushing to ingestion
func (l *leader) ValidateL7RuleObj(key string, l7Rule *akov1alpha2.L7Rule) error {
	l7RuleSpec := l7Rule.Spec
	refData := make(map[string]string)
	if l7RuleSpec.BotPolicyRef != nil {
		refData[*l7RuleSpec.BotPolicyRef] = "BotPolicy"
	}

	if l7RuleSpec.SecurityPolicyRef != nil {
		refData[*l7RuleSpec.SecurityPolicyRef] = "SecurityPolicy"
	}

	if l7RuleSpec.TrafficCloneProfileRef != nil {
		refData[*l7RuleSpec.TrafficCloneProfileRef] = "TrafficCloneProfile"
	}
	tenant := lib.GetTenantInNamespace(l7Rule.Namespace)

	if err := checkRefsOnController(key, refData, tenant); err != nil {
		status.UpdateL7RuleStatus(key, l7Rule, status.UpdateCRDStatusOptions{
			Status: lib.StatusRejected,
			Error:  err.Error(),
		})
		return err
	}
	// No need to update status of l7rule object as accepted since it was accepted before.
	if l7Rule.Status.Status == lib.StatusAccepted {
		return nil
	}
	status.UpdateL7RuleStatus(key, l7Rule, status.UpdateCRDStatusOptions{Status: lib.StatusAccepted, Error: ""})
	return nil
}

func validateLBAlgorithm(backendProperties *akov1alpha2.BackendProperties) error {
	if backendProperties.LbAlgorithm == nil {
		return nil
	}
	switch *backendProperties.LbAlgorithm {
	case lib.LB_ALGORITHM_CONSISTENT_HASH:
		if backendProperties.LbAlgorithmHash == nil {
			return fmt.Errorf("lbAlgorithmHash must be specified when lbAlgorithm is \"%s\"", lib.LB_ALGORITHM_CONSISTENT_HASH)
		} else {
			if *backendProperties.LbAlgorithmHash == lib.LB_ALGORITHM_CONSISTENT_HASH_CUSTOM_HEADER &&
				backendProperties.LbAlgorithmConsistentHashHdr == nil {
				return fmt.Errorf("lbAlgorithmConsistentHashHdr must be specified when lbAlgorithmHash is \"%s\"", lib.LB_ALGORITHM_CONSISTENT_HASH_CUSTOM_HEADER)
			}
		}
	default:
		if backendProperties.LbAlgorithmHash != nil {
			return fmt.Errorf("lbAlgorithmHash must not be specified when lbAlgorithm is \"%s\"", *backendProperties.LbAlgorithm)
		}
		if backendProperties.LbAlgorithmConsistentHashHdr != nil {
			return fmt.Errorf("lbAlgorithmConsistentHashHdr must not be specified when lbAlgorithm is \"%s\"", *backendProperties.LbAlgorithm)
		}
	}
	return nil
}

func (f *follower) ValidateHTTPRuleObj(key string, httprule *akov1beta1.HTTPRule) error {
	utils.AviLog.Debugf("key: %s, AKO is not a leader, not validating HTTPRule object", key)
	return nil
}

func (f *follower) ValidateHostRuleObj(key string, hostrule *akov1beta1.HostRule) error {
	utils.AviLog.Debugf("key: %s, AKO is not a leader, not validating HostRule object", key)
	return nil
}

func (f *follower) ValidateAviInfraSetting(key string, infraSetting *akov1beta1.AviInfraSetting) error {

	utils.AviLog.Debugf("key: %s, AKO is not a leader, not validating AviInfraSetting object", key)
	// During AKO bootup as leader is not set, crd validation is not done.
	// This creates problem in vip network and pool network population.
	if infraSetting.Status.Status == lib.StatusAccepted {
		segMgmtNetworK := ""
		if infraSetting.Spec.SeGroup.Name != "" {
			addSeGroupLabel(key, infraSetting.Spec.SeGroup.Name)
			// Not required for no access cloud
			if lib.GetCloudType() == lib.CLOUD_VCENTER {
				segMgmtNetworK = GetSEGManagementNetwork(infraSetting.Spec.SeGroup.Name)
			}
		}

		if len(infraSetting.Spec.Network.VipNetworks) > 0 {
			SetAviInfrasettingVIPNetworks(infraSetting.Name, segMgmtNetworK, infraSetting.Spec.SeGroup.Name, infraSetting.Spec.Network.VipNetworks)
		}

		if len(infraSetting.Spec.Network.NodeNetworks) > 0 {
			SetAviInfrasettingNodeNetworks(infraSetting.Name, segMgmtNetworK, infraSetting.Spec.SeGroup.Name, infraSetting.Spec.Network.NodeNetworks)
		}
	}
	return nil
}

func (f *follower) ValidateMultiClusterIngressObj(key string, multiClusterIngress *akov1alpha1.MultiClusterIngress) error {
	utils.AviLog.Debugf("key: %s, AKO is not a leader, not validating MultiClusterIngress object", key)
	return nil
}

func (f *follower) ValidateServiceImportObj(key string, serviceImport *akov1alpha1.ServiceImport) error {

	// CHECK ME: AMKO creates this and validation required?
	// TODO: validations needs a status field
	utils.AviLog.Debugf("key: %s, AKO is not a leader, not validating ServiceImport object", key)
	return nil
}

func (f *follower) ValidateSSORuleObj(key string, ssoRule *akov1alpha2.SSORule) error {
	utils.AviLog.Debugf("key: %s, AKO is not a leader, not validating SSORule object", key)
	return nil
}

func (l *follower) ValidateL4RuleObj(key string, l4Rule *akov1alpha2.L4Rule) error {
	utils.AviLog.Debugf("key: %s, AKO is not a leader, not validating L4Rule object", key)
	return nil
}

func (f *follower) ValidateL7RuleObj(key string, l7Rule *akov1alpha2.L7Rule) error {
	utils.AviLog.Debugf("key: %s, AKO is not a leader, not validating L7Rule object", key)
	return nil
}
