/*
 * Copyright 2020-2021 VMware, Inc.
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

package lib

import (
	"os"
	"sort"
	"strings"
	"sync"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	istiocrd "istio.io/client-go/pkg/clientset/versioned"
	istioInformer "istio.io/client-go/pkg/informers/externalversions/networking/v1alpha3"
	svcapi "sigs.k8s.io/service-apis/pkg/client/clientset/versioned"
	svcInformer "sigs.k8s.io/service-apis/pkg/client/informers/externalversions/apis/v1alpha1"

	"github.com/vmware/alb-sdk/go/models"

	akocrd "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/client/v1alpha1/clientset/versioned"

	v1alpha1akoinformer "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/client/v1alpha1/informers/externalversions/ako/v1alpha1"
	v1alpha2akocrd "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/client/v1alpha2/clientset/versioned"
	v1alpha2akoinformer "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/client/v1alpha2/informers/externalversions/ako/v1alpha2"
	v1beta1akocrd "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/client/v1beta1/clientset/versioned"
	v1beta1akoinformer "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/client/v1beta1/informers/externalversions/ako/v1beta1"

	"github.com/vmware/alb-sdk/go/clients"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	advl4crd "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/third_party/service-apis/client/clientset/versioned"
	advl4informer "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/third_party/service-apis/client/informers/externalversions/apis/v1alpha1pre1"
)

type AdvL4Informers struct {
	GatewayInformer      advl4informer.GatewayInformer
	GatewayClassInformer advl4informer.GatewayClassInformer
}

type ServicesAPIInformers struct {
	GatewayInformer      svcInformer.GatewayInformer
	GatewayClassInformer svcInformer.GatewayClassInformer
}

type AKOCrdInformers struct {
	HostRuleInformer        v1beta1akoinformer.HostRuleInformer
	HTTPRuleInformer        v1beta1akoinformer.HTTPRuleInformer
	AviInfraSettingInformer v1beta1akoinformer.AviInfraSettingInformer
	SSORuleInformer         v1alpha2akoinformer.SSORuleInformer
	L4RuleInformer          v1alpha2akoinformer.L4RuleInformer
	L7RuleInformer          v1alpha2akoinformer.L7RuleInformer
	HMInformer              v1alpha1akoinformer.HealthMonitorInformer
}

type IstioCRDInformers struct {
	VirtualServiceInformer  istioInformer.VirtualServiceInformer
	DestinationRuleInformer istioInformer.DestinationRuleInformer
	GatewayInformer         istioInformer.GatewayInformer
}

type BlockedNamespaces struct {
	BlockedNSMap map[string]struct{}
	nsChecksum   uint32
}

// akoControlConfig struct is intended to store all AKO related global
// variables, that are set as part of AKO bootup. This is a store of client-sets,
// informers, config parameters, and internally computed static configurations.
// TODO (shchauhan): Add other global parameters, which are currently present as independent
// global variables as part of lib.go
type akoControlConfig struct {
	// client-set and informer for v1alpha1pre1 services API.
	advL4Clientset    advl4crd.Interface
	akoAdvL4Informers *AdvL4Informers

	// client-set and informer for v1alpha1 services API.
	svcAPICS        svcapi.Interface
	svcAPIInformers *ServicesAPIInformers

	// client-set and informer for v1alpha1 AKO CRDs.
	// can't remove it as MCI and serviceimport uses it.
	crdClientset akocrd.Interface
	crdInformers *AKOCrdInformers

	// client-set and informer for v1alpha3 istio CRDs.
	istioClientset istiocrd.Interface
	istioInformers *IstioCRDInformers

	// akoEventRecorder is used to store record.akoEventRecorder
	// that allows AKO to broadcast kubernetes Events.
	akoEventRecorder *utils.EventRecorder

	// akoPodObjectMeta holds AKO Pod ObjectMeta information
	akoPodObjectMeta *metav1.ObjectMeta

	// aviInfraSettingEnabled is set to true if the cluster has
	// AviInfraSetting CRD installed.
	aviInfraSettingEnabled bool
	// hostRuleEnabled is set to true if the cluster has
	// HostRule CRD installed.
	hostRuleEnabled bool
	// httpRuleEnabled is set to true if the cluster has
	// HTTPRule CRD installed.
	httpRuleEnabled bool

	// client-set and informer for v1alpha2 of AKO CRD.
	v1alpha2crdClientset v1alpha2akocrd.Interface

	//client set and informer for v1beta1
	v1beta1crdClientset v1beta1akocrd.Interface

	// ssoRuleEnabled is set to true if the cluster has
	// SSORule CRD installed.
	ssoRuleEnabled bool

	// l4RuleEnabled is set to true if the cluster has
	// L4Rule CRD installed.
	l4RuleEnabled bool

	// l7RuleEnabled is set to true if the cluster has
	// L7Rule CRD installed.
	l7RuleEnabled bool

	// hmEnabled is set to true if the cluster has
	// HealthMonitor CRD installed
	hmRuleEnabled bool

	// licenseType holds the default license tier which would be used by new Clouds. Enum options - ENTERPRISE_16, ENTERPRISE, ENTERPRISE_18, BASIC, ESSENTIALS.
	licenseType string

	// primaryaAKO is set to true/false if as per primaryaAKO value
	// in values.yaml
	primaryaAKO bool

	//blockedNS contains map of blocked namespaces and checksum of it
	blockedNS BlockedNamespaces

	// leadership status of AKO
	isLeader     bool
	isLeaderLock sync.RWMutex

	// controllerVersion stores the version of the controller to
	// which AKO is communicating with
	controllerVersion string

	// defaultLBController is set to true/false as per defaultLBController value in values.yaml
	defaultLBController bool

	//Controller VRF Context is stored
	controllerVRFContext string

	//Prometheus enabled or not
	isPrometheusEnabled bool

	//endpointSlices Enabled
	isEndpointSlicesEnabled bool

	//fqdnReusePolicy is set to Strict/InterNamespaceAllowed according to whether AKO allows FQDN sharing across namespaces
	fqdnReusePolicy string
}

var akoControlConfigInstance *akoControlConfig

func AKOControlConfig() *akoControlConfig {
	if akoControlConfigInstance == nil {
		akoControlConfigInstance = &akoControlConfig{
			controllerVersion: initControllerVersion(),
		}
	}
	return akoControlConfigInstance
}

func (c *akoControlConfig) SetIsLeaderFlag(flag bool) {
	c.isLeaderLock.Lock()
	defer c.isLeaderLock.Unlock()
	c.isLeader = flag
}

func (c *akoControlConfig) SetDefaultLBController(flag bool) {
	c.defaultLBController = flag
}

func (c *akoControlConfig) IsLeader() bool {
	c.isLeaderLock.RLock()
	defer c.isLeaderLock.RUnlock()
	return c.isLeader
}

func (c *akoControlConfig) SetAKOInstanceFlag(flag bool) {
	c.primaryaAKO = flag
}

func (c *akoControlConfig) GetAKOInstanceFlag() bool {
	return c.primaryaAKO
}

func (c *akoControlConfig) SetAKOPrometheusFlag(flag bool) {
	c.isPrometheusEnabled = flag
}

func (c *akoControlConfig) GetAKOAKOPrometheusFlag() bool {
	return c.isPrometheusEnabled
}

func (c *akoControlConfig) SetEndpointSlicesEnabled(flag bool) {
	c.isEndpointSlicesEnabled = flag
}
func (c *akoControlConfig) GetEndpointSlicesEnabled() bool {
	return c.isEndpointSlicesEnabled
}

func (c *akoControlConfig) SetAKOBlockedNSList(nsList []string) {
	sort.Strings(nsList)
	val := strings.Join(nsList, ":")
	cksum := utils.Hash(val)
	if c.blockedNS.nsChecksum != cksum {
		nsMap := make(map[string]struct{})
		for _, ns := range nsList {
			nsMap[ns] = struct{}{}
		}
		c.blockedNS.nsChecksum = cksum
		c.blockedNS.BlockedNSMap = nsMap
	}
}
func (c *akoControlConfig) GetAKOBlockedNSList() map[string]struct{} {
	return c.blockedNS.BlockedNSMap
}
func (c *akoControlConfig) SetAdvL4Clientset(cs advl4crd.Interface) {
	c.advL4Clientset = cs
}

func (c *akoControlConfig) AdvL4Clientset() advl4crd.Interface {
	return c.advL4Clientset
}

func (c *akoControlConfig) SetAdvL4Informers(i *AdvL4Informers) {
	c.akoAdvL4Informers = i
}

func (c *akoControlConfig) AdvL4Informers() *AdvL4Informers {
	return c.akoAdvL4Informers
}

func (c *akoControlConfig) SetServicesAPIClientset(cs svcapi.Interface) {
	c.svcAPICS = cs
}

func (c *akoControlConfig) ServicesAPIClientset() svcapi.Interface {
	return c.svcAPICS
}

func (c *akoControlConfig) SetSvcAPIsInformers(i *ServicesAPIInformers) {
	c.svcAPIInformers = i
}

func (c *akoControlConfig) SvcAPIInformers() *ServicesAPIInformers {
	return c.svcAPIInformers
}

func (c *akoControlConfig) SetCRDClientset(cs akocrd.Interface) {
	c.crdClientset = cs
	c.SetCRDEnabledParams(cs)
}

func (c *akoControlConfig) SetCRDClientsetAndEnableInfraSettingParam(cs v1beta1akocrd.Interface) {
	c.v1beta1crdClientset = cs
	c.aviInfraSettingEnabled = true
}

func (c *akoControlConfig) CRDClientset() akocrd.Interface {
	return c.crdClientset
}

func (c *akoControlConfig) Setv1alpha2CRDClientset(cs v1alpha2akocrd.Interface) {
	c.v1alpha2crdClientset = cs
	c.Setv1alpha2CRDEnabledParams(cs)
}

func (c *akoControlConfig) V1alpha2CRDClientset() v1alpha2akocrd.Interface {
	return c.v1alpha2crdClientset
}

func (c *akoControlConfig) Setv1beta1CRDClientset(cs v1beta1akocrd.Interface) {
	c.v1beta1crdClientset = cs
	c.Setv1beta1CRDEnabledParams(cs)
}

func (c *akoControlConfig) V1beta1CRDClientset() v1beta1akocrd.Interface {
	return c.v1beta1crdClientset
}

func (c *akoControlConfig) HealthMonitorEnabled() bool {
	return c.hmRuleEnabled
}
func (c *akoControlConfig) Setv1beta1CRDEnabledParams(cs v1beta1akocrd.Interface) {
	c.aviInfraSettingEnabled = true
	c.hostRuleEnabled = true
	c.httpRuleEnabled = true
}

// CRDs are by default installed on all AKO deployments. So always enable CRD parameters.
func (c *akoControlConfig) SetCRDEnabledParams(cs akocrd.Interface) {
	c.aviInfraSettingEnabled = true
	c.hostRuleEnabled = true
	c.httpRuleEnabled = true
	c.hmRuleEnabled = true
}

func (c *akoControlConfig) Setv1alpha2CRDEnabledParams(cs v1alpha2akocrd.Interface) {
	c.ssoRuleEnabled = true
	c.l4RuleEnabled = true
	c.l7RuleEnabled = true
}

func (c *akoControlConfig) AviInfraSettingEnabled() bool {
	return c.aviInfraSettingEnabled
}

func (c *akoControlConfig) HostRuleEnabled() bool {
	return c.hostRuleEnabled
}

func (c *akoControlConfig) HttpRuleEnabled() bool {
	return c.httpRuleEnabled
}

func (c *akoControlConfig) SsoRuleEnabled() bool {
	return c.ssoRuleEnabled
}

func (c *akoControlConfig) L4RuleEnabled() bool {
	return c.l4RuleEnabled
}
func (c *akoControlConfig) L7RuleEnabled() bool {
	return c.l7RuleEnabled
}

func (c *akoControlConfig) ControllerVersion() string {
	return c.controllerVersion
}

func (c *akoControlConfig) SetControllerVersion(v string) {
	c.controllerVersion = v
}

func (c *akoControlConfig) IsAviDefaultLBController() bool {
	return c.defaultLBController
}

func (c *akoControlConfig) ControllerVRFContext() string {
	return c.controllerVRFContext
}

func (c *akoControlConfig) SetControllerVRFContext(v string) {
	c.controllerVRFContext = v
}

func (c *akoControlConfig) SetAKOFQDNReusePolicy(FQDNPolicy string) {
	// Empty or SNI deployment--> Allow across namespace
	if FQDNPolicy == "" || !IsEvhEnabled() {
		FQDNPolicy = FQDNReusePolicyOpen
	}

	if FQDNPolicy != FQDNReusePolicyOpen && FQDNPolicy != FQDNReusePolicyStrict {
		// if not one of it, set it to open
		FQDNPolicy = FQDNReusePolicyOpen
	}
	c.fqdnReusePolicy = FQDNPolicy
	utils.AviLog.Infof("AKO FQDN reuse policy is: %s", c.fqdnReusePolicy)
}

// This utility returns FQDN Reuse policy of AKO.
// Strict --> FQDN restrict to one namespace
// InternamespaceAllowed --> FQDN can be spanned across multiple namespaces
func (c *akoControlConfig) GetAKOFQDNReusePolicy() string {
	return c.fqdnReusePolicy
}

func initControllerVersion() string {
	version := os.Getenv("CTRL_VERSION")
	if version == "" {
		return version
	}

	// Ensure that the controllerVersion is less than the supported Avi maxVersion and more than minVersion.
	if CompareVersions(version, ">", GetAviMaxSupportedVersion()) {
		utils.AviLog.Infof("Setting the client version to AVI Max supported version %s", GetAviMaxSupportedVersion())
		version = GetAviMaxSupportedVersion()
		return version
	}

	if CompareVersions(version, "<", GetAviMinSupportedVersion()) {
		AKOControlConfig().PodEventf(
			corev1.EventTypeWarning,
			AKOShutdown, "AKO is running with unsupported Avi version %s",
			version,
		)
		utils.AviLog.Fatalf("AKO is not supported for the Avi version %s, Avi must be %s or more", version, GetAviMinSupportedVersion())
	}
	return ""
}

func (c *akoControlConfig) SetIstioClientset(cs istiocrd.Interface) {
	c.istioClientset = cs
}

func (c *akoControlConfig) IstioClientset() istiocrd.Interface {
	return c.istioClientset
}

func (c *akoControlConfig) SetCRDInformers(i *AKOCrdInformers) {
	c.crdInformers = i
}

func (c *akoControlConfig) CRDInformers() *AKOCrdInformers {
	return c.crdInformers
}

func (c *akoControlConfig) SetIstioCRDInformers(i *IstioCRDInformers) {
	c.istioInformers = i
}

func (c *akoControlConfig) IstioCRDInformers() *IstioCRDInformers {
	return c.istioInformers
}

func (c *akoControlConfig) SetEventRecorder(id string, client kubernetes.Interface, fake bool) {
	c.akoEventRecorder = utils.NewEventRecorder(id, client, fake)
}

func (c *akoControlConfig) EventsSetEnabled(enable string) {
	if enable == "true" {
		utils.AviLog.Infof("Enabling event broadcasting via AKO.")
		c.akoEventRecorder.Enabled = true
	} else {
		utils.AviLog.Infof("Disabling event broadcasting via AKO.")
		c.akoEventRecorder.Enabled = false
	}
}

func (c *akoControlConfig) EventRecorder() *utils.EventRecorder {
	return c.akoEventRecorder
}

func (c *akoControlConfig) SaveAKOPodObjectMeta(pod *v1.Pod) {
	c.akoPodObjectMeta = &pod.ObjectMeta
}

func (c *akoControlConfig) PodEventf(eventType, reason, message string, formatArgs ...string) {
	if c.akoPodObjectMeta != nil {
		if len(formatArgs) > 0 {
			c.EventRecorder().Eventf(&v1.Pod{ObjectMeta: *c.akoPodObjectMeta}, eventType, reason, message, formatArgs)
		} else {
			c.EventRecorder().Event(&v1.Pod{ObjectMeta: *c.akoPodObjectMeta}, eventType, reason, message)
		}
	}
}
func (c *akoControlConfig) IngressEventf(ingMeta metav1.ObjectMeta, eventType, reason, message string, formatArgs ...string) {

	if len(formatArgs) > 0 {
		c.EventRecorder().Eventf(&networkingv1.Ingress{ObjectMeta: ingMeta}, eventType, reason, message, formatArgs)
	} else {
		c.EventRecorder().Event(&networkingv1.Ingress{ObjectMeta: ingMeta}, eventType, reason, message)
	}

}
func GetResponseFromURI(client *clients.AviClient, uri string) (models.SystemConfiguration, error) {
	response := models.SystemConfiguration{}
	err := AviGet(client, uri, &response)

	if err != nil {
		utils.AviLog.Warnf("Unable to fetch system configuration, error %s", err.Error())
	}

	return response, err
}

func (c *akoControlConfig) GetLicenseType() string {
	return c.licenseType
}

func (c *akoControlConfig) SetLicenseType(client *clients.AviClient) {
	uri := "/api/systemconfiguration"
	response, err := GetResponseFromURI(client, uri)
	if err != nil {
		utils.AviLog.Warnf("Unable to fetch system configuration, error %s", err.Error())
		return
	}

	c.licenseType = *response.DefaultLicenseTier
}
