/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/onsi/gomega"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	akov1alpha1 "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-operator/api/v1alpha1"
)

func TestMain(m *testing.M) {
	ret := m.Run()
	os.Exit(ret)
}

func getTestDefaultAKOConfig() akov1alpha1.AKOConfig {
	akoConfigSpec := akov1alpha1.AKOConfigSpec{
		ReplicaCount:    1,
		ImageRepository: "test-repo",
		ImagePullPolicy: "Always",
		AKOSettings: akov1alpha1.AKOSettings{
			LogLevel:               akov1alpha1.LogLevelType("INFO"),
			FullSyncFrequency:      "1800",
			APIServerPort:          8080,
			DeleteConfig:           false,
			DisableStaticRouteSync: false,
			ClusterName:            "test-cluster",
			CNIPlugin:              "test-cni",
			EnableEVH:              false,
			Layer7Only:             false,
			ServicesAPI:            false,
			EnableEvents:           true,
			IPFamily:               "V4",
			IstioEnabled:           false,
			VipPerNamespace:        false,
			BlockedNamespaceList:   []string{},
			UseDefaultSecretsOnly:  false,
		},

		NetworkSettings: akov1alpha1.NetworkSettings{
			NodeNetworkList: []akov1alpha1.NodeNetwork{
				{
					NetworkName: "test-nw",
					Cidrs:       []string{"10.10.10.0/24"},
				},
			},
			EnableRHI: false,
			VipNetworkList: []akov1alpha1.VipNetwork{
				{
					NetworkName: "test-nw",
					Cidr:        "10.10.10.0/24",
					V6Cidr:      "",
				},
			},
			BGPPeerLabels: []string{},
			NsxtT1LR:      "",
		},

		L7Settings: akov1alpha1.L7Settings{
			DefaultIngController: true,
			ShardVSSize:          akov1alpha1.VSSize("LARGE"),
			ServiceType:          akov1alpha1.ServiceTypeStr("ClusterIP"),
			PassthroughShardSize: akov1alpha1.PassthroughVSSize("SMALL"),
			NoPGForSNI:           false,
		},

		L4Settings: akov1alpha1.L4Settings{
			DefaultDomain: "test.com",
			AutoFQDN:      "default",
		},

		ControllerSettings: akov1alpha1.ControllerSettings{
			ServiceEngineGroupName: "test-group",
			ControllerVersion:      "1.1",
			CloudName:              "test-cloud",
			ControllerIP:           "10.10.10.11",
			TenantName:             "admin",
		},

		NodePortSelector: akov1alpha1.NodePortSelector{
			Key:   "key",
			Value: "value",
		},

		Resources: akov1alpha1.Resources{
			Limits: akov1alpha1.ResourceLimits{
				CPU:    "350m",
				Memory: "400Mi",
			},
			Requests: akov1alpha1.ResourceRequests{
				CPU:    "200m",
				Memory: "300Mi",
			},
		},

		Rbac: akov1alpha1.Rbac{
			PSPEnable: true,
		},
		MountPath: "/log",
		LogFile:   "ako.log",
		ImagePullSecrets: []akov1alpha1.ImagePullSecret{
			{
				Name: "regcred",
			},
		},
		AKOGatewayLogFile: "avi-gw.log",
		FeatureGates:      akov1alpha1.FeatureGates{GatewayAPI: true},
		GatewayAPI: akov1alpha1.GatewayAPI{
			Image: akov1alpha1.Image{
				PullPolicy: "Always",
				Repository: "test-repo-gateway-api",
			},
		},
	}
	akoConfig := akov1alpha1.AKOConfig{
		TypeMeta: v1.TypeMeta{
			Kind:       "AKOConfig",
			APIVersion: "v1alpha1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      "test-akoconfig",
			Namespace: "avi-system",
		},
		Spec: akoConfigSpec,
	}

	return akoConfig
}

func TestConfigmap(t *testing.T) {
	// Test for:
	// 1. Whether a configmap is generated from akoConfig
	// 2. Whether checksums are different: this would mean, updating the configmap resource
	// 3. Whether AKO reboot is required: for properties like logLevel and deleteConfig changes,
	//    reboot is not required, for others, reboot is required
	g := gomega.NewGomegaWithT(t)
	akoConfig := getTestDefaultAKOConfig()

	defaultCm, err := getTestDefaultConfigMap()
	g.Expect(err).To(gomega.BeNil())
	j, _ := json.MarshalIndent(defaultCm, "", "    ")
	t.Log(string(j))
	buildConfigMapAndVerify(defaultCm, akoConfig, false, true, t)

	t.Log("updating CNI plugin and verifying")
	akoConfig.Spec.AKOSettings.CNIPlugin = "test-cni2"
	cmCni := buildConfigMapAndVerify(defaultCm, akoConfig, true, false, t)

	t.Log("updating log level and verifying")
	akoConfig.Spec.LogLevel = "DEBUG"
	cmLog := buildConfigMapAndVerify(cmCni, akoConfig, false, false, t)

	t.Log("updating cloud name and verifying")
	akoConfig.Spec.CloudName = "test-cloud2"
	cmCloud := buildConfigMapAndVerify(cmLog, akoConfig, true, false, t)

	t.Log("updating deleteConfig and verifying")
	akoConfig.Spec.AKOSettings.DeleteConfig = true
	cmDelete := buildConfigMapAndVerify(cmCloud, akoConfig, false, false, t)

	t.Log("updating disableRouteSync and verifying")
	akoConfig.Spec.AKOSettings.DisableStaticRouteSync = true
	cmDisable := buildConfigMapAndVerify(cmDelete, akoConfig, true, false, t)

	t.Log("updating networkUUID in place of networkName and verifying")
	akoConfig.Spec.NetworkSettings.NodeNetworkList = []akov1alpha1.NodeNetwork{
		{
			NetworkUUID: "test-nw-uuid",
			Cidrs:       []string{"10.10.10.0/24"},
		},
	}
	akoConfig.Spec.NetworkSettings.VipNetworkList = []akov1alpha1.VipNetwork{
		{
			NetworkUUID: "test-nw-uuid",
			Cidr:        "10.10.10.0/24",
		},
	}
	cmNetworkUUID := buildConfigMapAndVerify(cmDisable, akoConfig, true, false, t)

	t.Log("updating nodeNetworkList and verifying")
	akoConfig.Spec.NetworkSettings.NodeNetworkList = []akov1alpha1.NodeNetwork{}
	cmNodeNetwork := buildConfigMapAndVerify(cmNetworkUUID, akoConfig, true, false, t)

	t.Log("updating ipFamily and verifying")
	akoConfig.Spec.AKOSettings.IPFamily = "V6"
	cmIPFamily := buildConfigMapAndVerify(cmNodeNetwork, akoConfig, true, false, t)

	t.Log("updating istioEnabled and verifying")
	akoConfig.Spec.AKOSettings.IstioEnabled = true
	cmIstioEnabled := buildConfigMapAndVerify(cmIPFamily, akoConfig, true, false, t)

	t.Log("updating blockedNamespaceList and verifying")
	akoConfig.Spec.AKOSettings.BlockedNamespaceList = []string{"blocked-ns"}
	cmBlockedNamespaceList := buildConfigMapAndVerify(cmIstioEnabled, akoConfig, true, false, t)

	t.Log("updating useDefaultSecretsOnly and verifying")
	akoConfig.Spec.AKOSettings.UseDefaultSecretsOnly = true
	cmUseDefaultSecretsOnly := buildConfigMapAndVerify(cmBlockedNamespaceList, akoConfig, true, false, t)

	t.Log("updating defaultLBController and verifying")
	akoConfig.Spec.L4Settings.DefaultLBController = false
	cmDefaultLBController := buildConfigMapAndVerify(cmUseDefaultSecretsOnly, akoConfig, true, false, t)

	t.Log("updating vrfName and verifying")
	akoConfig.Spec.ControllerSettings.VRFName = "test-vrf"
	cmVRFName := buildConfigMapAndVerify(cmDefaultLBController, akoConfig, true, false, t)

	t.Log("updating EnablePrometheus and verifying")
	akoConfig.Spec.FeatureGates.EnablePrometheus = true
	buildConfigMapAndVerify(cmVRFName, akoConfig, true, false, t)
}

func TestStatefulset(t *testing.T) {
	// Test for:
	// 1. Whether a statefulset is generated from akoConfig
	// 2. Whether an update is required for the statefulsets
	g := gomega.NewGomegaWithT(t)
	akoConfig := getTestDefaultAKOConfig()
	t.Log("verifying default statefulset parsing")
	defaultSf, err := getTestDefaultStatefulSet()
	g.Expect(err).To(gomega.BeNil())

	t.Log("verifying the generated statefulset and the default statefulset to be equal")
	sfFromConfig := buildStatefulSetAndVerify(defaultSf, akoConfig, false, false, t)

	t.Log("updating PVC claim in the akoConfig and verifying statefulset update")
	akoConfig.Spec.PersistentVolumeClaim = "pvc-name"
	sfPVC := buildStatefulSetAndVerify(sfFromConfig, akoConfig, true, false, t)

	t.Log("verifying the volume mounts after setting pvc")
	container := sfPVC.Spec.Template.Spec.Containers[0]
	g.Expect(container.VolumeMounts[0].MountPath).To(gomega.Equal(akoConfig.Spec.MountPath))

	t.Log("updating resources in the akoConfig and verifying statefulset update")
	akoConfig.Spec.Resources.Limits.CPU = "10m"
	sfRes := buildStatefulSetAndVerify(sfPVC, akoConfig, true, false, t)

	t.Log("updating an invalid imagePullPolicy and verifying the error")
	akoConfig.Spec.ImagePullPolicy = "invalid pull"
	buildStatefulSetAndVerify(sfRes, akoConfig, true, true, t)

	t.Log("updating ako api server port and verifying statefulset update")
	// fix the image pull policy first
	akoConfig.Spec.ImagePullPolicy = "Always"
	akoConfig.Spec.AKOSettings.APIServerPort = 9090
	sfPort := buildStatefulSetAndVerify(sfRes, akoConfig, true, false, t)

	t.Log("updating istioEnabled to true and verifying statefulset update")
	akoConfig.Spec.IstioEnabled = true
	sfIstio := buildStatefulSetAndVerify(sfPort, akoConfig, true, false, t)

	t.Log("verifying the volume mounts after setting istioEnabled to true")
	container = sfIstio.Spec.Template.Spec.Containers[0]
	g.Expect(container.VolumeMounts[1].MountPath).To(gomega.Equal("/etc/istio-output-certs/"))

	t.Log("updating gateway api feature gate to false and verifying statefulset update")
	akoConfig.Spec.FeatureGates.GatewayAPI = false
	sfFeatureGate := buildStatefulSetAndVerify(sfIstio, akoConfig, true, false, t)

	t.Log("verifying the removal of ako-gateway-api container from ako sts")
	lenContainers := len(sfFeatureGate.Spec.Template.Spec.Containers)
	g.Expect(lenContainers).To(gomega.Equal(1))
}
