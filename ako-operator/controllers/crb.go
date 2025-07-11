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
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"

	akov1alpha1 "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-operator/api/v1alpha1"
)

func createOrUpdateClusterroleBinding(ctx context.Context, ako akov1alpha1.AKOConfig, log logr.Logger, r *AKOConfigReconciler) error {
	var oldCRB rbacv1.ClusterRoleBinding

	if err := r.Get(ctx, getCRBName(), &oldCRB); err != nil {
		log.V(0).Info("no existing clusterrolebinding with name", "name", CRBName)
	} else {
		if oldCRB.GetName() != "" {
			log.V(0).Info("a clusterrolebinding with name already exists, will update", "name",
				oldCRB.GetName())
		}
	}

	crb := BuildClusterroleBinding(ako, r, log)
	if oldCRB.ObjectMeta.GetName() != "" {
		if reflect.DeepEqual(oldCRB.Subjects, crb.Subjects) {
			log.V(0).Info("no updates required for clusterrolebinding")
			// add this object in the global list
			objList := getObjectList()
			objList[types.NamespacedName{
				Name: oldCRB.GetName(),
			}] = &oldCRB
			return nil
		}
		err := r.Update(ctx, &crb)
		if err != nil {
			log.Error(err, "unable to update clusterrolebinding", "namespace", crb.GetNamespace(),
				"name", crb.GetName())
			return err
		}
	} else {
		err := r.Create(ctx, &crb)
		if err != nil {
			log.Error(err, "unable to create clusterrolebinding", "namespace", crb.GetNamespace(),
				"name", crb.GetName())
			return err
		}
	}
	var newCRB rbacv1.ClusterRoleBinding
	err := r.Get(ctx, getCRBName(), &newCRB)
	if err != nil {
		log.V(0).Info("error getting a clusterrole with name", "name", getCRName().Name, "err", err)
	}
	// update this object in the global list
	objList := getObjectList()
	objList[types.NamespacedName{
		Name: newCRB.GetName(),
	}] = &newCRB

	return nil
}

func BuildClusterroleBinding(ako akov1alpha1.AKOConfig, r *AKOConfigReconciler, log logr.Logger) rbacv1.ClusterRoleBinding {
	crb := rbacv1.ClusterRoleBinding{
		ObjectMeta: v1.ObjectMeta{
			Name:      CRBName,
			Namespace: AviSystemNS,
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     AKOCR,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      AKOServiceAccount,
				Namespace: AviSystemNS,
			},
		},
	}

	err := ctrl.SetControllerReference(&ako, &crb, r.Scheme)
	if err != nil {
		log.Error(err, "error in setting controller reference, clusterrolebinding changes would be ignored")
	}
	return crb
}
