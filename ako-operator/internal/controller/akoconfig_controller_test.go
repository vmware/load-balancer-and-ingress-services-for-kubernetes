/*
Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.

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

package controller

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	akov1beta1 "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-operator/api/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// These tests use Ginkgo (BDD-style Go testing framework) and can be run using "make test" command.
var _ = Describe("AKOConfig Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "test-resource"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: AviSystemNS,
		}
		akoconfig := &akov1beta1.AKOConfig{}
		aviSecret := &corev1.Secret{}

		BeforeEach(func() {
			By("Create the avi-system namespace if it doesn't exist")
			ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: AviSystemNS}}
			if err := k8sClient.Get(ctx, types.NamespacedName{Name: AviSystemNS}, ns); err != nil {
				if errors.IsNotFound(err) {
					Expect(k8sClient.Create(ctx, ns)).To(Succeed())
				} else {
					Expect(err).NotTo(HaveOccurred())
				}
			}
			By("creating the custom resource for the Kind AKOConfig")
			err := k8sClient.Get(ctx, typeNamespacedName, akoconfig)
			if err != nil && errors.IsNotFound(err) {
				resource := &akov1beta1.AKOConfig{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: AviSystemNS,
						UID:       "test-uid", // Set a UID for controller reference
						//Finalizers: []string{
						//	"ako.vmware.com/cleanup",
						//},
					},
					Spec: akov1beta1.AKOConfigSpec{
						ImageRepository: "projects.packages.broadcom.com/ako/ako",
						ImagePullPolicy: "IfNotPresent",
						ReplicaCount:    1,
						AKOSettings: akov1beta1.AKOSettings{
							LogLevel:               "INFO",
							EnableEvents:           true,
							FullSyncFrequency:      "1800",
							APIServerPort:          8080,
							DeleteConfig:           false,
							DisableStaticRouteSync: false,
							ClusterName:            "test-cluster",
							CNIPlugin:              "test-cni",
							NSSelector: akov1beta1.NamespaceSelector{
								LabelKey:   "env",
								LabelValue: "prod",
							},
							EnableEVH:             false,
							Layer7Only:            false,
							ServicesAPI:           false,
							VipPerNamespace:       false,
							IstioEnabled:          false,
							BlockedNamespaceList:  []string{"kube-system", "kube-public"},
							IPFamily:              "V4",
							UseDefaultSecretsOnly: false,
						},
						NetworkSettings: akov1beta1.NetworkSettings{
							NodeNetworkList: []akov1beta1.NodeNetwork{
								{
									NetworkName: "node-net",
									Cidrs:       []string{"192.168.1.0/24"},
								},
							},
							EnableRHI: false,
							VipNetworkList: []akov1beta1.VipNetwork{
								{
									NetworkName: "test-network",
									Cidr:        "10.0.0.0/24",
								},
							},
							BGPPeerLabels: []string{"bgp-label1", "bgp-label2"},
							NsxtT1LR:      "/infra/tier-1s/avi-t1-test",
							DefaultDomain: "testdomain.com",
						},
						L7Settings: akov1beta1.L7Settings{
							DefaultIngController: true,
							ShardVSSize:          akov1beta1.VSSize("LARGE"),
							ServiceType:          akov1beta1.ServiceTypeStr("ClusterIP"),
							PassthroughShardSize: akov1beta1.PassthroughVSSize("SMALL"),
							NoPGForSNI:           false,
							FQDNReusePolicy:      akov1beta1.FQDNReusePolicy("Strict"),
						},
						L4Settings: akov1beta1.L4Settings{
							// DefaultDomain from L4Settings is not used if NetworkSettings.DefaultDomain is set
							DefaultDomain:       "testdomain-l4.com",
							AutoFQDN:            "default",
							DefaultLBController: true,
						},
						ControllerSettings: akov1beta1.ControllerSettings{
							ServiceEngineGroupName: "test-seg",
							ControllerVersion:      "22.1.3",
							CloudName:              "test-cloud",
							ControllerIP:           "10.10.10.10",
							TenantName:             "test-tenant",
							VRFName:                "test-vrf",
						},
						NodePortSelector: akov1beta1.NodePortSelector{
							Key:   "key",
							Value: "value",
						},
						Resources: akov1beta1.Resources{
							Limits:   akov1beta1.ResourceLimits{CPU: "100m", Memory: "100Mi"},
							Requests: akov1beta1.ResourceRequests{CPU: "50m", Memory: "50Mi"},
						},
						MountPath: "/log",
						LogFile:   "ako.log",
						ImagePullSecrets: []akov1beta1.ImagePullSecret{
							{
								Name: "regcred",
							},
						},
						AKOGatewayLogFile: "avi-gw.log",
						FeatureGates: akov1beta1.FeatureGates{
							GatewayAPI:          true,
							EnableEndpointSlice: true,
							EnablePrometheus:    true,
						},
						GatewayAPI: akov1beta1.GatewayAPI{
							Image: akov1beta1.Image{
								PullPolicy: "Always",
								Repository: "test-repo-gateway-api",
							},
						},
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())

				// Create the avi-secret, as it's needed by BuildStatefulSet
				aviSecret = &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      AviSecretName,
						Namespace: AviSystemNS, // Ensure this matches where AKOConfigReconciler expects it
					},
					Data: map[string][]byte{
						"username": []byte("admin"),
						"password": []byte("password"),
					},
				}
				err = k8sClient.Get(ctx, types.NamespacedName{Name: AviSecretName, Namespace: AviSystemNS}, &corev1.Secret{})
				if err != nil && errors.IsNotFound(err) {
					Expect(k8sClient.Create(ctx, aviSecret)).To(Succeed())
				} else {
					Expect(err).NotTo(HaveOccurred()) // Handle other errors
				}

			}
		})

		AfterEach(func() {
			// TODO(user): Cleanup logic after each test, like removing the resource instance.
			By("Deleting the AKOConfig resource")
			akoConfigToDelete := &akov1beta1.AKOConfig{}
			err := k8sClient.Get(ctx, typeNamespacedName, akoConfigToDelete)
			if err == nil {
				Expect(k8sClient.Delete(ctx, akoConfigToDelete)).To(Succeed())
			} else if !errors.IsNotFound(err) {
				Expect(err).NotTo(HaveOccurred())
			}

			// Cleanup the secret
			By("Deleting the avi-secret")
			err = k8sClient.Get(ctx, types.NamespacedName{Name: AviSecretName, Namespace: AviSystemNS}, aviSecret)
			if err == nil {
				Expect(k8sClient.Delete(ctx, aviSecret)).To(Succeed())
			}
		})
		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")
			controllerReconciler := &AKOConfigReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
				Config: cfg, // Ensure reconciler has the config, as it's used by createCRDs
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
			// TODO(user): Add more specific assertions depending on your controller's reconciliation logic.
			// Example: If you expect a certain status condition after reconciliation, verify it here.
			fetchedAKOConfig := &akov1beta1.AKOConfig{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, typeNamespacedName, fetchedAKOConfig)
				return err == nil && fetchedAKOConfig.Status.State == string(AKOConfigStatusReady)
			}, "30s", "1s").Should(BeTrue(), "AKOConfig status should be Ready")
		})

		It("should create a ConfigMap when AKOConfig is reconciled", func() {
			controllerReconciler := &AKOConfigReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
				Config: cfg,
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			createdCm := &corev1.ConfigMap{}
			cmNamespacedName := types.NamespacedName{Name: ConfigMapName, Namespace: AviSystemNS}

			Eventually(func() error {
				return k8sClient.Get(ctx, cmNamespacedName, createdCm)
			}, "10s", "1s").Should(Succeed())

			Expect(createdCm.Data[ControllerIP]).To(Equal("10.10.10.10"))
			Expect(createdCm.Data[ControllerVersion]).To(Equal("22.1.3"))
			Expect(createdCm.Data[CniPlugin]).To(Equal("test-cni"))
			Expect(createdCm.Data[EnableEVH]).To(Equal("false"))
			Expect(createdCm.Data[Layer7Only]).To(Equal("false"))
			Expect(createdCm.Data[ServicesAPI]).To(Equal("false"))
			Expect(createdCm.Data[VipPerNamespace]).To(Equal("false"))
			Expect(createdCm.Data[ShardVSSize]).To(Equal("LARGE"))
			Expect(createdCm.Data[PassthroughShardSize]).To(Equal("SMALL")) // Constant in utils.go is "passhthroughShardSize"
			Expect(createdCm.Data[FullSyncFrequency]).To(Equal("1800"))
			Expect(createdCm.Data[CloudName]).To(Equal("test-cloud"))
			Expect(createdCm.Data[ClusterName]).To(Equal("test-cluster"))
			Expect(createdCm.Data[EnableRHI]).To(Equal("false"))
			Expect(createdCm.Data[DefaultDomain]).To(Equal("testdomain.com"))
			Expect(createdCm.Data[DisableStaticRouteSync]).To(Equal("false"))
			Expect(createdCm.Data[DefaultIngController]).To(Equal("true"))
			Expect(createdCm.Data[VipNetworkList]).To(MatchJSON(`[{"networkName":"test-network","cidr":"10.0.0.0/24"}]`))
			Expect(createdCm.Data[BgpPeerLabels]).To(MatchJSON(`["bgp-label1","bgp-label2"]`))
			Expect(createdCm.Data[EnableEvents]).To(Equal("true"))
			Expect(createdCm.Data[LogLevel]).To(Equal("INFO"))
			Expect(createdCm.Data[DeleteConfig]).To(Equal("false"))
			Expect(createdCm.Data[AutoFQDN]).To(Equal("default"))
			Expect(createdCm.Data[ServiceType]).To(Equal("ClusterIP"))
			Expect(createdCm.Data[ServiceEngineGroupName]).To(Equal("test-seg"))
			Expect(createdCm.Data[NodeNetworkList]).To(MatchJSON(`[{"networkName":"node-net","cidrs":["192.168.1.0/24"]}]`))
			Expect(createdCm.Data[APIServerPort]).To(Equal("8080"))
			Expect(createdCm.Data[NSSyncLabelKey]).To(Equal("env"))
			Expect(createdCm.Data[NSSyncLabelValue]).To(Equal("prod"))
			Expect(createdCm.Data[TenantName]).To(Equal("test-tenant"))
			Expect(createdCm.Data[NoPGForSni]).To(Equal("false"))
			Expect(createdCm.Data[NsxtT1LR]).To(Equal("/infra/tier-1s/avi-t1-test"))
			Expect(createdCm.Data[PrimaryInstance]).To(Equal("true"))
			Expect(createdCm.Data[IstioEnabled]).To(Equal("false"))
			Expect(createdCm.Data[BlockedNamespaceList]).To(MatchJSON(`["kube-system","kube-public"]`))
			Expect(createdCm.Data[IPFamily]).To(Equal("V4"))
			Expect(createdCm.Data[EnableMCI]).To(Equal("false"))
			Expect(createdCm.Data[UseDefaultSecretsOnly]).To(Equal("false"))
			Expect(createdCm.Data[VRFName]).To(Equal("test-vrf"))
			Expect(createdCm.Data[DefaultLBController]).To(Equal("true"))
			Expect(createdCm.Data[EnablePrometheus]).To(Equal("true"))
			Expect(createdCm.Data[FQDNReusePolicy]).To(Equal("Strict"))
			Expect(createdCm.Data[EnableEndpointSlice]).To(Equal("true"))
			Expect(createdCm.Data).NotTo(HaveKey(NodeKey))   // Since ServiceType is ClusterIP
			Expect(createdCm.Data).NotTo(HaveKey(NodeValue)) // Since ServiceType is ClusterIP
		})

		It("should create a StatefulSet when AKOConfig is reconciled", func() {
			controllerReconciler := &AKOConfigReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
				Config: cfg,
			}
			currentAkoConfig := &akov1beta1.AKOConfig{}
			Expect(k8sClient.Get(ctx, typeNamespacedName, currentAkoConfig)).To(Succeed())

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			createdSf := &appsv1.StatefulSet{}
			sfNamespacedName := types.NamespacedName{Name: StatefulSetName, Namespace: AviSystemNS}

			Eventually(func() error {
				return k8sClient.Get(ctx, sfNamespacedName, createdSf)
			}, "10s", "1s").Should(Succeed())

			Expect(*createdSf.Spec.Replicas).To(Equal(int32(1)))
			Expect(createdSf.Spec.Template.Spec.Containers[0].Image).To(Equal("projects.packages.broadcom.com/ako/ako"))
			Expect(createdSf.Spec.Template.Spec.Containers[0].ImagePullPolicy).To(Equal(corev1.PullIfNotPresent))
			Expect(createdSf.Spec.Template.Spec.ServiceAccountName).To(Equal(ServiceAccountName))

			// Assert AKO container specific settings
			akoContainer := createdSf.Spec.Template.Spec.Containers[0]
			Expect(akoContainer.Name).To(Equal("ako"))
			Expect(akoContainer.LivenessProbe.HTTPGet.Port.IntValue()).To(Equal(8080))
			Expect(akoContainer.Resources.Limits.Cpu().String()).To(Equal("100m"))
			Expect(akoContainer.Resources.Limits.Memory().String()).To(Equal("100Mi"))
			Expect(akoContainer.Resources.Requests.Cpu().String()).To(Equal("50m"))
			Expect(akoContainer.Resources.Requests.Memory().String()).To(Equal("50Mi"))

			// Assert AKO container environment variables (sample)
			// You can be more exhaustive here by checking all env vars set by getEnvVars
			expectedEnvVarFound := false
			for _, envVar := range akoContainer.Env {
				if envVar.Name == "LOG_FILE_NAME" {
					Expect(envVar.Value).To(Equal("ako.log"))
					expectedEnvVarFound = true
					break
				}
			}
			Expect(expectedEnvVarFound).To(BeTrue(), "Expected LOG_FILE_NAME env var to be present")

			// Assert Gateway API container specific settings (if enabled)
			if akoconfig.Spec.FeatureGates.GatewayAPI {
				Expect(len(createdSf.Spec.Template.Spec.Containers)).To(Equal(2))
				gatewayContainer := createdSf.Spec.Template.Spec.Containers[1]
				Expect(gatewayContainer.Name).To(Equal("ako-gateway-api"))
				Expect(gatewayContainer.Image).To(Equal("test-repo-gateway-api"))
				Expect(gatewayContainer.ImagePullPolicy).To(Equal(corev1.PullAlways))
				Expect(gatewayContainer.Resources).To(Equal(akoContainer.Resources)) // Assuming they share resources
			}

			Expect(len(createdSf.Spec.Template.Spec.ImagePullSecrets)).To(Equal(1))
			Expect(createdSf.Spec.Template.Spec.ImagePullSecrets[0].Name).To(Equal("regcred"))
		})
	})
})
