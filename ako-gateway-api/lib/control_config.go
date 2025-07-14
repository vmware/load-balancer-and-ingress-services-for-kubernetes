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

package lib

import (
	"os"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	gatewayclientset "sigs.k8s.io/gateway-api/pkg/client/clientset/versioned"
	gatewayinformerv1 "sigs.k8s.io/gateway-api/pkg/client/informers/externalversions/apis/v1"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	v1beta1akocrd "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/client/v1beta1/clientset/versioned"
	v1beta1akoinformer "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/client/v1beta1/informers/externalversions/ako/v1beta1"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
)

type GatewayAPIInformers struct {
	GatewayInformer      gatewayinformerv1.GatewayInformer
	GatewayClassInformer gatewayinformerv1.GatewayClassInformer
	HTTPRouteInformer    gatewayinformerv1.HTTPRouteInformer
}

// akoControlConfig struct is intended to store all AKO related global
// variables, that are set as part of AKO bootup. This is a store of client-sets,
// informers, config parameters, and internally computed static configurations.

type akoControlConfig struct {

	// client-set and informer for Gateway API.
	gwApiCS        gatewayclientset.Interface
	gwApiInformers *GatewayAPIInformers

	// akoEventRecorder is used to store record.akoEventRecorder
	// that allows AKO to broadcast kubernetes Events.
	akoEventRecorder *utils.EventRecorder

	// akoPodObjectMeta holds AKO Pod ObjectMeta information
	akoPodObjectMeta *metav1.ObjectMeta

	// controllerVersion stores the version of the controller to
	// which AKO is communicating with
	controllerVersion string

	// v1beta1 client set for AKO CRDs
	v1beta1crdClientset v1beta1akocrd.Interface

	// flag to enable AviInfraSetting informer
	aviInfraSettingEnabled bool

	// AviInfraSetting Informer
	aviInfraSettingInformer v1beta1akoinformer.AviInfraSettingInformer
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

func (c *akoControlConfig) SetGatewayAPIClientset(cs gatewayclientset.Interface) {
	c.gwApiCS = cs
}

func (c *akoControlConfig) GatewayAPIClientset() gatewayclientset.Interface {
	return c.gwApiCS
}

func (c *akoControlConfig) SetGatewayApiInformers(i *GatewayAPIInformers) {
	c.gwApiInformers = i
}

func (c *akoControlConfig) GatewayApiInformers() *GatewayAPIInformers {
	return c.gwApiInformers
}

func (c *akoControlConfig) ControllerVersion() string {
	return c.controllerVersion
}

func (c *akoControlConfig) SetControllerVersion(v string) {
	c.controllerVersion = v
}

func (c *akoControlConfig) SetV1Beta1CRDClientSetAndEnableAviInfraSettingParam(cs v1beta1akocrd.Interface) {
	c.v1beta1crdClientset = cs
	c.aviInfraSettingEnabled = true
}

func (c *akoControlConfig) V1Beta1CRDClientSet() v1beta1akocrd.Interface {
	return c.v1beta1crdClientset
}

func (c *akoControlConfig) SetAviInfraSettingInformer(aviInfraSettingInformer v1beta1akoinformer.AviInfraSettingInformer) {
	c.aviInfraSettingInformer = aviInfraSettingInformer
}

func (c *akoControlConfig) AviInfraSettingInformer() v1beta1akoinformer.AviInfraSettingInformer {
	return c.aviInfraSettingInformer
}

func (c *akoControlConfig) AviInfraSettingEnabled() bool {
	return c.aviInfraSettingEnabled
}

func initControllerVersion() string {
	version := os.Getenv("CTRL_VERSION")
	if version == "" {
		return version
	}

	// Ensure that the controllerVersion is less than the supported Avi maxVersion and more than minVersion.
	if lib.CompareVersions(version, ">", lib.GetAviMaxSupportedVersion()) {
		utils.AviLog.Infof("Setting the client version to AVI Max supported version %s", lib.GetAviMaxSupportedVersion())
		version = lib.GetAviMaxSupportedVersion()
		return version
	}
	return ""
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
