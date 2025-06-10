/*
Copyright 2025.

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
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	akov1alpha1 "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-crd-operator/api/v1alpha1"
	ctrlutils "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-crd-operator/internal/utils"
)

var _ = Describe("HealthMonitor Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "test-resource"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default", // TODO(user):Modify as needed
		}
		healthmonitor := &akov1alpha1.HealthMonitor{}

		BeforeEach(func() {
			By("creating the custom resource for the Kind HealthMonitor")
			err := k8sClient.Get(ctx, typeNamespacedName, healthmonitor)
			if err != nil && errors.IsNotFound(err) {
				resource := &akov1alpha1.HealthMonitor{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
					},
					// TODO(user): Specify other spec details if needed.
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
		})

		AfterEach(func() {
			// TODO(user): Cleanup logic after each test, like removing the resource instance.
			resource := &akov1alpha1.HealthMonitor{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance HealthMonitor")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
		})
		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")
			controllerReconciler := &HealthMonitorReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
			// TODO(user): Add more specific assertions depending on your controller's reconciliation logic.
			// Example: If you expect a certain status condition after reconciliation, verify it here.
		})
	})

	Context("When creating markers", func() {
		It("should create markers with cluster name and namespace", func() {
			// Set environment variable for cluster name
			originalClusterName := os.Getenv("CLUSTER_NAME")
			defer os.Setenv("CLUSTER_NAME", originalClusterName)
			os.Setenv("CLUSTER_NAME", "test-cluster")

			markers := ctrlutils.CreateMarkers(originalClusterName, "test-namespace")

			Expect(markers).To(HaveLen(2))

			// Check cluster name marker
			found := false
			for _, marker := range markers {
				if *marker.Key == "clustername" {
					Expect(marker.Values).To(Equal([]string{"test-cluster"}))
					found = true
					break
				}
			}
			Expect(found).To(BeTrue(), "clustername marker should be present")

			// Check namespace marker
			found = false
			for _, marker := range markers {
				if *marker.Key == "namespace" {
					Expect(marker.Values).To(Equal([]string{"test-namespace"}))
					found = true
					break
				}
			}
			Expect(found).To(BeTrue(), "namespace marker should be present")
		})

		It("should handle missing cluster name environment variable", func() {
			// Unset environment variable for cluster name
			originalClusterName := os.Getenv("CLUSTER_NAME")
			defer os.Setenv("CLUSTER_NAME", originalClusterName)
			os.Unsetenv("CLUSTER_NAME")

			markers := ctrlutils.CreateMarkers("", "test-namespace")

			// Should only have namespace marker since cluster name is not set
			Expect(markers).To(HaveLen(1))
			Expect(*markers[0].Key).To(Equal("namespace"))
			Expect(markers[0].Values).To(Equal([]string{"test-namespace"}))
		})

		It("should handle empty namespace", func() {
			// Set environment variable for cluster name
			originalClusterName := os.Getenv("CLUSTER_NAME")
			defer os.Setenv("CLUSTER_NAME", originalClusterName)
			os.Setenv("CLUSTER_NAME", "test-cluster")

			markers := ctrlutils.CreateMarkers("", "")

			// Should only have cluster name marker since namespace is empty
			Expect(markers).To(HaveLen(1))
			Expect(*markers[0].Key).To(Equal("clustername"))
			Expect(markers[0].Values).To(Equal([]string{"test-cluster"}))
		})
	})
})
