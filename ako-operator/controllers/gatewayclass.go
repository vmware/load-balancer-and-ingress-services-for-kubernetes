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

	logr "github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	gatewayv1beta1 "sigs.k8s.io/gateway-api/apis/v1beta1"

	akov1alpha1 "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-operator/api/v1alpha1"
)

func createOrUpdateGatewayClass(ctx context.Context, ako akov1alpha1.AKOConfig, log logr.Logger, r *AKOConfigReconciler) error {
	var oldGWClass gatewayv1beta1.GatewayClass

	if err := r.Get(ctx, getGWClassName(), &oldGWClass); err != nil {
		log.V(0).Info("no existing gatewayclass with name", "name", GWClassName)
	} else {
		if oldGWClass.GetName() != "" {
			log.V(0).Info("a gatewayclass with name already exists", "name",
				oldGWClass.GetName())
			if !ako.Spec.FeatureGates.GatewayAPI {
				log.V(0).Info("GatewayAPI feature gate is disabled, will delete gatewayclass with", "name", GWClassName)
				err := r.Delete(ctx, &oldGWClass)
				if err != nil {
					log.Error(err, "unable to delete gatewayclass", "name", oldGWClass.GetName())
				}
				return nil
			}

			// add this object in the global list
			objList := getObjectList()
			objList[types.NamespacedName{
				Name: oldGWClass.GetName(),
			}] = &oldGWClass
			return nil
		}
	}
	if !ako.Spec.FeatureGates.GatewayAPI {
		log.V(0).Info("GatewayAPI feature gate is disabled, will not create gatewayclass with name", "name", GWClassName)
		return nil
	}

	gwClass := BuildGatewayClass(ako, r, log)
	err := r.Create(ctx, &gwClass)
	if err != nil {
		log.Error(err, "unable to create gatewayclass", "name", gwClass.GetName())
		return err
	}

	var newGWClass gatewayv1beta1.GatewayClass
	err = r.Get(ctx, getGWClassName(), &newGWClass)
	if err != nil {
		log.V(0).Info("error getting a gatewayclass with name", "name", getGWClassName().Name, "err", err)
	}
	// update this object in the global list
	objList := getObjectList()
	objList[types.NamespacedName{
		Name: newGWClass.GetName(),
	}] = &newGWClass

	return nil
}

// BuildGatewayClass builds a gatewayclass object if GatewayAPI feature gate is enabled
func BuildGatewayClass(ako akov1alpha1.AKOConfig, r *AKOConfigReconciler, log logr.Logger) gatewayv1beta1.GatewayClass {
	gwClass := gatewayv1beta1.GatewayClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: GWClassName,
		},
		Spec: gatewayv1beta1.GatewayClassSpec{
			ControllerName: GWClassController,
		},
	}
	return gwClass
}
