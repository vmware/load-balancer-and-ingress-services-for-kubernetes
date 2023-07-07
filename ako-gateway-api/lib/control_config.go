/*
 * Copyright 2023-2024 VMware, Inc.
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

	corev1 "k8s.io/api/core/v1"

	"github.com/vmware/alb-sdk/go/models"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/third_party/github.com/vmware/alb-sdk/go/clients"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	gwApi "sigs.k8s.io/gateway-api/pkg/client/clientset/versioned"
	gwApiInformer "sigs.k8s.io/gateway-api/pkg/client/informers/externalversions/apis/v1beta1"
)

type gwApiInformers struct {
	GatewayInformer      gwApiInformer.GatewayInformer
	GatewayClassInformer gwApiInformer.GatewayClassInformer
}

// akoControlConfig struct is intended to store all AKO related global
// variables, that are set as part of AKO bootup. This is a store of client-sets,
// informers, config parameters, and internally computed static configurations.

type akoControlConfig struct {

	// client-set and informer for Gateway API.
	gwApiCS        gwApi.Interface
	gwApiInformers *gwApiInformers

	// akoEventRecorder is used to store record.akoEventRecorder
	// that allows AKO to broadcast kubernetes Events.
	akoEventRecorder *utils.EventRecorder

	// akoPodObjectMeta holds AKO Pod ObjectMeta information
	akoPodObjectMeta *metav1.ObjectMeta

	// licenseType holds the default license tier which would be used by new Clouds. Enum options - ENTERPRISE_16, ENTERPRISE, ENTERPRISE_18, BASIC, ESSENTIALS.
	licenseType string

	// controllerVersion stores the version of the controller to
	// which AKO is communicating with
	controllerVersion string
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

func (c *akoControlConfig) SetGatewayAPIClientset(cs gwApi.Interface) {
	c.gwApiCS = cs
}

func (c *akoControlConfig) GatewayAPIClientset() gwApi.Interface {
	return c.gwApiCS
}

func (c *akoControlConfig) SetGatewayApiInformers(i *gwApiInformers) {
	c.gwApiInformers = i
}

func (c *akoControlConfig) GatewayApiInformers() *gwApiInformers {
	return c.gwApiInformers
}

func (c *akoControlConfig) ControllerVersion() string {
	return c.controllerVersion
}

func (c *akoControlConfig) SetControllerVersion(v string) {
	c.controllerVersion = v
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

	if lib.CompareVersions(version, "<", lib.GetAviMinSupportedVersion()) {
		AKOControlConfig().PodEventf(
			corev1.EventTypeWarning,
			lib.AKOShutdown, "AKO is running with unsupported Avi version %s",
			version,
		)
		utils.AviLog.Fatalf("AKO is not supported for the Avi version %s, Avi must be %s or more", version, lib.GetAviMinSupportedVersion())
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

func GetResponseFromURI(client *clients.AviClient, uri string) (models.SystemConfiguration, error) {
	response := models.SystemConfiguration{}
	err := lib.AviGet(client, uri, &response)

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
