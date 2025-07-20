/*
Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
All Rights Reserved.

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
	"context"
	"reflect"

	"github.com/go-logr/logr"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	akov1alpha1 "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-operator/api/v1alpha1"
)

func createOrUpdateClusterRole(ctx context.Context, ako akov1alpha1.AKOConfig, log logr.Logger, r *AKOConfigReconciler) error {
	var oldCR rbacv1.ClusterRole

	if err := r.Get(ctx, getCRName(), &oldCR); err != nil {
		log.V(0).Info("no existing clusterrole with name", "name", getCRName())
	} else {
		if oldCR.GetName() != "" {
			log.V(0).Info("a clusterrole with name already exists, it will be updated", "name",
				oldCR.GetName())
		}
	}

	cr := BuildClusterrole(ako, r, log)
	if oldCR.GetName() != "" {
		if reflect.DeepEqual(oldCR.Rules, cr.Rules) {
			log.V(0).Info("no updates required for clusterrole")
			// add this object in the global list
			objList := getObjectList()
			objList[types.NamespacedName{
				Name: oldCR.GetName(),
			}] = &oldCR
			return nil
		}
		err := r.Update(ctx, &cr)
		if err != nil {
			log.Error(err, "unable to update clusterrole", "namespace", cr.GetNamespace(), "name",
				cr.GetName())
			return err
		}
	} else {
		err := r.Create(ctx, &cr)
		if err != nil {
			log.Error(err, "unable to create clusterrole", "namespace", cr.GetNamespace(),
				"name", cr.GetName())
			return err
		}
	}

	var newCR rbacv1.ClusterRole
	err := r.Get(ctx, getCRName(), &newCR)
	if err != nil {
		log.V(0).Info("error getting a clusterrole with name", "name", getCRName().Name, "err", err)
		return err
	}
	// update this object in the global list
	objList := getObjectList()
	objList[types.NamespacedName{
		Name: newCR.GetName(),
	}] = &newCR

	return nil
}

func BuildClusterrole(ako akov1alpha1.AKOConfig, r *AKOConfigReconciler, log logr.Logger) rbacv1.ClusterRole {
	cr := rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: AKOCR,
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"*"},
				Verbs:     []string{"get", "watch", "list"},
			},
			{
				APIGroups: []string{"apps"},
				Resources: []string{"statefulsets", "statefulsets/status"},
				Verbs:     []string{"get", "watch", "list", "patch", "update"},
			},
			{
				APIGroups: []string{"extensions", "networking.k8s.io"},
				Resources: []string{"ingresses", "ingresses/status"},
				Verbs:     []string{"get", "watch", "list", "patch", "update"},
			},
			{
				APIGroups: []string{"networking.k8s.io"},
				Resources: []string{"ingressclasses"},
				Verbs:     []string{"get", "watch", "list"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"services", "services/status"},
				Verbs:     []string{"get", "watch", "list", "patch", "update"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"secrets"},
				Verbs:     []string{"get", "watch", "list", "patch", "update", "create"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"events"},
				Verbs:     []string{"create", "patch", "update"},
			},
			{
				APIGroups: []string{"crd.projectcalico.org"},
				Resources: []string{"blockaffinities"},
				Verbs:     []string{"get", "watch", "list"},
			},
			{
				APIGroups: []string{"network.openshift.io"},
				Resources: []string{"hostsubnets"},
				Verbs:     []string{"get", "watch", "list"},
			},
			{
				APIGroups: []string{"route.openshift.io"},
				Resources: []string{"routes", "routes/status"},
				Verbs:     []string{"get", "watch", "list", "patch", "update"},
			},
			{
				APIGroups: []string{"ako.vmware.com"},
				Resources: []string{"hostrules", "hostrules/status", "httprules", "httprules/status", "aviinfrasettings", "aviinfrasettings/status", "l4rules", "l4rules/status", "ssorules", "ssorules/status", "l7rules", "l7rules/status"},
				Verbs:     []string{"get", "watch", "list", "patch", "update"},
			},
			{
				APIGroups: []string{"networking.x-k8s.io"},
				Resources: []string{"gateways", "gateways/status", "gatewayclasses", "gatewayclasses/status"},
				Verbs:     []string{"get", "watch", "list", "patch", "update"},
			},
			{
				APIGroups: []string{"coordination.k8s.io"},
				Resources: []string{"leases"},
				Verbs:     []string{"create", "get", "update"},
			},
			{
				APIGroups: []string{"cilium.io"},
				Resources: []string{"ciliumnodes"},
				Verbs:     []string{"get", "watch", "list"},
			},
			{
				APIGroups: []string{"gateway.networking.k8s.io"},
				Resources: []string{"gatewayclasses", "gatewayclasses/status", "gateways", "gateways/status", "httproutes", "httproutes/status"},
				Verbs:     []string{"get", "watch", "list", "patch", "update"},
			},
		},
	}

	if ako.Spec.Rbac.PSPEnable {
		cr.Rules = append(cr.Rules, rbacv1.PolicyRule{
			APIGroups:     []string{"policy", "extensions"},
			Resources:     []string{"podsecuritypolicies"},
			Verbs:         []string{"use"},
			ResourceNames: []string{"ako"},
		})
	}

	return cr
}
