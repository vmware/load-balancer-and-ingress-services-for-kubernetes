/*
 * Copyright 2022-2023 VMware, Inc.
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
	"fmt"
	"net"
	"regexp"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/status"
	akov1alpha1 "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/apis/ako/v1alpha1"
	akov1alpha2 "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/apis/ako/v1alpha2"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
)

type Validator interface {
	ValidateHTTPRuleObj(key string, httprule *akov1alpha1.HTTPRule) error
	ValidateHostRuleObj(key string, hostrule *akov1alpha1.HostRule) error
	ValidateAviInfraSetting(key string, infraSetting *akov1alpha1.AviInfraSetting) error
	ValidateMultiClusterIngressObj(key string, multiClusterIngress *akov1alpha1.MultiClusterIngress) error
	ValidateServiceImportObj(key string, serviceImport *akov1alpha1.ServiceImport) error
	ValidateL4RuleObj(key string, l4Rule *akov1alpha2.L4Rule) error
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
func (l *leader) ValidateHostRuleObj(key string, hostrule *akov1alpha1.HostRule) error {

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
		if hostrule.Spec.VirtualHost.FqdnType != akov1alpha1.Exact {
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
	if hostrule.Spec.VirtualHost.TLS.SSLKeyCertificate.Type == akov1alpha1.HostRuleSecretTypeAviReference {
		refData[hostrule.Spec.VirtualHost.TLS.SSLKeyCertificate.Name] = "SslKeyCert"
	}

	if hostrule.Spec.VirtualHost.TLS.SSLKeyCertificate.Type == akov1alpha1.HostRuleSecretTypeSecretReference {
		secretName := hostrule.Spec.VirtualHost.TLS.SSLKeyCertificate.Name
		err := validateSecretReferenceInHostrule(hostrule.Namespace, secretName)
		if err != nil {
			status.UpdateHostRuleStatus(key, hostrule, status.UpdateCRDStatusOptions{Status: lib.StatusRejected, Error: err.Error()})
			return err
		}
	}
	if hostrule.Spec.VirtualHost.TLS.SSLKeyCertificate.AlternateCertificate.Type == akov1alpha1.HostRuleSecretTypeAviReference {
		refData[hostrule.Spec.VirtualHost.TLS.SSLKeyCertificate.AlternateCertificate.Name] = "SslKeyCert"
	}

	if hostrule.Spec.VirtualHost.TLS.SSLKeyCertificate.AlternateCertificate.Type == akov1alpha1.HostRuleSecretTypeSecretReference {
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

	if err := checkRefsOnController(key, refData); err != nil {
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

// validateHTTPRuleObj would do validation checks
// update internal CRD caches, and push relevant ingresses to ingestion
func (l *leader) ValidateHTTPRuleObj(key string, httprule *akov1alpha1.HTTPRule) error {

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

	if err := checkRefsOnController(key, refData); err != nil {
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
func (l *leader) ValidateAviInfraSetting(key string, infraSetting *akov1alpha1.AviInfraSetting) error {

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
		refData[vipNetwork.NetworkName] = "Network"
	}

	if infraSetting.Spec.SeGroup.Name != "" {
		refData[infraSetting.Spec.SeGroup.Name] = "ServiceEngineGroup"
	}

	if err := checkRefsOnController(key, refData); err != nil {
		status.UpdateAviInfraSettingStatus(key, infraSetting, status.UpdateCRDStatusOptions{
			Status: lib.StatusRejected,
			Error:  err.Error(),
		})
		return err
	}

	// This would add SEG labels only if they are not configured yet. In case there is a label mismatch
	// to any pre-existing SEG labels, the AviInfraSettig CR will get Rejected from the checkRefsOnController
	// step before this.
	if infraSetting.Spec.SeGroup.Name != "" {
		addSeGroupLabel(key, infraSetting.Spec.SeGroup.Name)
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

	if l4RuleSpec.ApplicationProfileRef != nil {
		refData[*l4RuleSpec.ApplicationProfileRef] = "AppProfile"
	}

	if l4RuleSpec.NetworkProfileRef != nil {
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

	if err := checkRefsOnController(key, refData); err != nil {
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

func (f *follower) ValidateHTTPRuleObj(key string, httprule *akov1alpha1.HTTPRule) error {
	utils.AviLog.Debugf("key: %s, AKO is not a leader, not validating HTTPRule object", key)
	return nil
}

func (f *follower) ValidateHostRuleObj(key string, hostrule *akov1alpha1.HostRule) error {
	utils.AviLog.Debugf("key: %s, AKO is not a leader, not validating HostRule object", key)
	return nil
}

func (f *follower) ValidateAviInfraSetting(key string, infraSetting *akov1alpha1.AviInfraSetting) error {
	utils.AviLog.Debugf("key: %s, AKO is not a leader, not validating AviInfraSetting object", key)
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

func (l *follower) ValidateL4RuleObj(key string, l4Rule *akov1alpha2.L4Rule) error {
	utils.AviLog.Debugf("key: %s, AKO is not a leader, not validating L4Rule object", key)
	return nil
}
