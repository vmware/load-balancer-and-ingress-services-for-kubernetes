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
	"context"
	"os"
	"sort"
	"strings"
	"sync"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	istiocrd "istio.io/client-go/pkg/clientset/versioned"
	istioInformer "istio.io/client-go/pkg/informers/externalversions/networking/v1alpha3"
	svcapi "sigs.k8s.io/service-apis/pkg/client/clientset/versioned"
	svcInformer "sigs.k8s.io/service-apis/pkg/client/informers/externalversions/apis/v1alpha1"

	"github.com/vmware/alb-sdk/go/models"

	akocrd "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/client/v1alpha1/clientset/versioned"
	akoinformer "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/client/v1alpha1/informers/externalversions/ako/v1alpha1"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/third_party/github.com/vmware/alb-sdk/go/clients"

	advl4crd "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/third_party/service-apis/client/clientset/versioned"
	advl4informer "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/third_party/service-apis/client/informers/externalversions/apis/v1alpha1pre1"
)

func init() {

}

type AdvL4Informers struct {
	GatewayInformer      advl4informer.GatewayInformer
	GatewayClassInformer advl4informer.GatewayClassInformer
}

type ServicesAPIInformers struct {
	GatewayInformer      svcInformer.GatewayInformer
	GatewayClassInformer svcInformer.GatewayClassInformer
}

type AKOCrdInformers struct {
	HostRuleInformer        akoinformer.HostRuleInformer
	HTTPRuleInformer        akoinformer.HTTPRuleInformer
	AviInfraSettingInformer akoinformer.AviInfraSettingInformer
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
}

var akoControlConfigInstance *akoControlConfig

func AKOControlConfig() *akoControlConfig {
	if akoControlConfigInstance == nil {
		akoControlConfigInstance = &akoControlConfig{
			controllerVersion: os.Getenv("CTRL_VERSION"),
		}
	}
	return akoControlConfigInstance
}

func (c *akoControlConfig) SetIsLeaderFlag(flag bool) {
	c.isLeaderLock.Lock()
	defer c.isLeaderLock.Unlock()
	c.isLeader = flag
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

func (c *akoControlConfig) CRDClientset() akocrd.Interface {
	return c.crdClientset
}

func (c *akoControlConfig) SetCRDEnabledParams(cs akocrd.Interface) {
	timeout := int64(120)
	_, aviInfraError := cs.AkoV1alpha1().AviInfraSettings().List(context.TODO(), metav1.ListOptions{TimeoutSeconds: &timeout})
	if aviInfraError != nil {
		utils.AviLog.Infof("ako.vmware.com/v1alpha1/AviInfraSetting not found/enabled on cluster: %v", aviInfraError)
		c.aviInfraSettingEnabled = false
	} else {
		utils.AviLog.Infof("ako.vmware.com/v1alpha1/AviInfraSetting enabled on cluster")
		c.aviInfraSettingEnabled = true
	}

	_, hostRulesError := cs.AkoV1alpha1().HostRules(metav1.NamespaceAll).List(context.TODO(), metav1.ListOptions{TimeoutSeconds: &timeout})
	if hostRulesError != nil {
		utils.AviLog.Infof("ako.vmware.com/v1alpha1/HostRule not found/enabled on cluster: %v", hostRulesError)
		c.hostRuleEnabled = false
	} else {
		utils.AviLog.Infof("ako.vmware.com/v1alpha1/HostRule enabled on cluster")
		c.hostRuleEnabled = true
	}

	_, httpRulesError := cs.AkoV1alpha1().HTTPRules(metav1.NamespaceAll).List(context.TODO(), metav1.ListOptions{TimeoutSeconds: &timeout})
	if httpRulesError != nil {
		utils.AviLog.Infof("ako.vmware.com/v1alpha1/HTTPRule not found/enabled on cluster: %v", httpRulesError)
		c.httpRuleEnabled = false
	} else {
		utils.AviLog.Infof("ako.vmware.com/v1alpha1/HTTPRule enabled on cluster")
		c.httpRuleEnabled = true
	}
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

func (c *akoControlConfig) ControllerVersion() string {
	return c.controllerVersion
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
