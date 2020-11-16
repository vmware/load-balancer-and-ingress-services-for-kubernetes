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
	akov1alpha1 "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-operator/api/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestMain(m *testing.M) {
	ret := m.Run()
	os.Exit(ret)
}

func getTestDefaultAKOConfig() akov1alpha1.AKOConfig {
	akoConfigSpec := akov1alpha1.AKOConfigSpec{
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
		},

		NetworkSettings: akov1alpha1.NetworkSettings{
			NodeNetworkList: []akov1alpha1.NodeNetwork{
				{
					NetworkName: "test-nw",
					Cidrs:       []string{"10.10.10.0/24"},
				},
			},
			SubnetIP:     "10.10.10.1",
			SubnetPrefix: "24",
			EnableRHI:    false,
			NetworkName:  "test-nw",
		},

		L7Settings: akov1alpha1.L7Settings{
			DefaultIngController: true,
			ShardVSSize:          akov1alpha1.VSSize("LARGE"),
			ServiceType:          akov1alpha1.ServiceTypeStr("ClusterIP"),
			PassthroughShardSize: akov1alpha1.VSSize("SMALL"),
			ShardingScheme:       "hostname",
		},

		L4Settings: akov1alpha1.L4Settings{
			AdvancedL4:    false,
			DefaultDomain: "test.com",
		},

		ControllerSettings: akov1alpha1.ControllerSettings{
			ServiceEngineGroupName: "test-group",
			ControllerVersion:      "1.1",
			CloudName:              "test-cloud",
			ControllerIP:           "10.10.10.11",
		},

		NodePortSelector: akov1alpha1.NodePortSelector{
			Key:   "key",
			Value: "value",
		},

		Resources: akov1alpha1.Resources{
			Limits: akov1alpha1.ResourceLimits{
				CPU:    "250m",
				Memory: "300Mi",
			},
			Requests: akov1alpha1.ResourceRequests{
				CPU:    "100m",
				Memory: "200Mi",
			},
		},

		Rbac: akov1alpha1.Rbac{
			PSPEnable: true,
		},
		MountPath: "/log",
		LogFile:   "ako.log",
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

	t.Log("updating nodeNetworkList and verifying")
	akoConfig.Spec.NetworkSettings.NodeNetworkList = []akov1alpha1.NodeNetwork{}
	buildConfigMapAndVerify(cmDisable, akoConfig, true, false, t)

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
	buildStatefulSetAndVerify(sfRes, akoConfig, true, false, t)
}
