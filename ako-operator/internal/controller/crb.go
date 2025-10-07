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
	"reflect"

	"github.com/go-logr/logr"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	akov1beta1 "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-operator/api/v1beta1"
)

func createOrUpdateClusterroleBinding(ctx context.Context, ako akov1beta1.AKOConfig, log logr.Logger, r *AKOConfigReconciler) error {
	desiredCRB := BuildClusterroleBinding(ako, r, log)
	var existingCRB rbacv1.ClusterRoleBinding

	err := r.Get(ctx, getCRBName(), &existingCRB)
	if err != nil {
		if !errors.IsNotFound(err) {
			log.Error(err, "failed to get ClusterRoleBinding")
			return err
		}
		// ClusterRoleBinding does not exist, create it.
		log.V(0).Info("no pre-existing clusterrolebinding, creating one")
		if err := r.Create(ctx, &desiredCRB); err != nil {
			log.Error(err, "unable to create clusterrolebinding", "name", desiredCRB.GetName())
			return err
		}
	} else {
		// ClusterRoleBinding exists, update it if necessary.
		if reflect.DeepEqual(existingCRB.Subjects, desiredCRB.Subjects) && reflect.DeepEqual(existingCRB.RoleRef, desiredCRB.RoleRef) {
			log.V(0).Info("no updates required for clusterrolebinding")
		} else {
			log.V(0).Info("a clusterrolebinding with name already exists, it will be updated", "name", existingCRB.GetName())
			// To update an existing object, we must modify the object that was fetched from the
			// cluster, as it contains the required resourceVersion.
			existingCRB.Subjects = desiredCRB.Subjects
			existingCRB.RoleRef = desiredCRB.RoleRef
			if err := r.Update(ctx, &existingCRB); err != nil {
				log.Error(err, "unable to update clusterrolebinding", "name", existingCRB.GetName())
				return err
			}
		}
	}

	// After creating or updating, fetch the latest state to update the global list.
	var currentCRB rbacv1.ClusterRoleBinding
	err = r.Get(ctx, getCRBName(), &currentCRB)
	if err != nil {
		log.Error(err, "error getting a clusterrolebinding with name", "name", getCRBName().Name)
		return err
	}
	// update this object in the global list
	objList := getObjectList()
	objList[types.NamespacedName{
		Name: currentCRB.GetName(),
	}] = &currentCRB

	return nil
}

func BuildClusterroleBinding(ako akov1beta1.AKOConfig, r *AKOConfigReconciler, log logr.Logger) rbacv1.ClusterRoleBinding {
	crb := rbacv1.ClusterRoleBinding{
		ObjectMeta: v1.ObjectMeta{
			Name: CRBName,
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

	return crb
}
