/*
 * Copyright 2024 VMware, Inc.
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

package infratests

import (
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-infra/ingestion"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// Common test constants
const (
	CommonTestNamespace = "test-namespace"
	CommonTestCluster   = "test-cluster"
	CommonSEGValue      = "test-seg"
)

// CreateTestCluster creates a test cluster object
func CreateTestCluster(name, namespace, phase string, labels map[string]string) *unstructured.Unstructured {
	cluster := &unstructured.Unstructured{}
	cluster.SetAPIVersion("cluster.x-k8s.io/v1beta1")
	cluster.SetKind("Cluster")
	cluster.SetName(name)
	cluster.SetNamespace(namespace)
	cluster.SetResourceVersion("1")

	if labels != nil {
		cluster.SetLabels(labels)
	}

	// Set status with phase
	if phase != "" {
		status := map[string]interface{}{
			"phase": phase,
		}
		cluster.Object["status"] = status
	}

	return cluster
}

// CreateTestNamespace creates a test namespace with optional SEG annotation
func CreateTestNamespace(name string, hasSEG bool) *corev1.Namespace {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}

	if hasSEG {
		ns.Annotations = map[string]string{
			ingestion.ServiceEngineGroupAnnotation: CommonSEGValue,
		}
	}

	return ns
}
