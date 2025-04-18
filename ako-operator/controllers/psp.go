/*
Copyright 2020 VMware, Inc.
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

//nolint:unparam
package controllers

import (
	"context"
	"reflect"

	logr "github.com/go-logr/logr"
	akov1alpha1 "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-operator/api/v1alpha1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func createOrUpdatePodSecurityPolicy(ctx context.Context, ako akov1alpha1.AKOConfig, log logr.Logger, r *AKOConfigReconciler) error {

	var oldPSP policyv1beta1.PodSecurityPolicy //nolint:errcheck

	if err := r.Get(ctx, getPSPName(), &oldPSP); err != nil {
		log.V(0).Info("no pre-existing podsecuritypolicy with name", "name", PSPName)
	} else {
		if oldPSP.GetName() != "" {
			log.V(0).Info("pre-existing podsecuritypolicy, will be updated", "name", oldPSP.GetName())
		}
	}

	if !ako.Spec.Rbac.PSPEnable {
		// PSP not required anymore, delete any existing psp
		objList := getObjectList()
		pspObj, ok := objList[getPSPName()]
		if !ok {
			return nil
		}
		r.deleteIfExists(ctx, getPSPName(), pspObj)
		return nil
	}

	psp := BuildPodSecurityPolicy(ako, r, log)
	if oldPSP.GetName() != "" {
		if reflect.DeepEqual(oldPSP.Spec, psp.Spec) {
			log.V(0).Info("no updates required for podsecuritypolicy")
			// add this object in the global list
			objList := getObjectList()
			objList[types.NamespacedName{
				Name: oldPSP.GetName(),
			}] = &oldPSP
			return nil
		}
		err := r.Update(ctx, &psp)
		if err != nil {
			log.Error(err, "unable to update podsecuritypolicy", "namespace", psp.GetNamespace(),
				"name", psp.GetName())
			return err
		}
	} else {
		err := r.Create(ctx, &psp)
		if err != nil {
			log.Error(err, "unable to create podsecuritypolicy", "namespace", psp.GetNamespace(),
				"name", psp.GetName())
			return err
		}
	}

	var newPSP policyv1beta1.PodSecurityPolicy
	err := r.Get(ctx, getPSPName(), &newPSP)
	if err != nil {
		log.V(0).Info("error getting a clusterrole with name", "name", getCRName().Name, "err", err)
	}
	// update this object in the global list
	objList := getObjectList()
	objList[types.NamespacedName{
		Name: newPSP.GetName(),
	}] = &newPSP

	return nil
}

func BuildPodSecurityPolicy(ako akov1alpha1.AKOConfig, r *AKOConfigReconciler, log logr.Logger) policyv1beta1.PodSecurityPolicy {
	priviledgedEscalation := false
	// conditionally add the api version
	psp := policyv1beta1.PodSecurityPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name: PSPName,
		},
		Spec: policyv1beta1.PodSecurityPolicySpec{
			Privileged:               false,
			AllowPrivilegeEscalation: &priviledgedEscalation,
			FSGroup: policyv1beta1.FSGroupStrategyOptions{
				Rule: "MustRunAs",
				Ranges: []policyv1beta1.IDRange{
					{
						Min: 1,
						Max: 65535,
					},
				},
			},
			HostNetwork: false,
			HostIPC:     false,
			HostPID:     false,
			RunAsUser: policyv1beta1.RunAsUserStrategyOptions{
				Rule: "RunAsAny",
			},
			SELinux: policyv1beta1.SELinuxStrategyOptions{Rule: "RunAsAny"},
			SupplementalGroups: policyv1beta1.SupplementalGroupsStrategyOptions{
				Rule: "MustRunAs",
				Ranges: []policyv1beta1.IDRange{
					{
						Min: 1,
						Max: 65535,
					},
				},
			},
			ReadOnlyRootFilesystem: false,
			Volumes:                []policyv1beta1.FSType{"configMap", "emptyDir", "projected", "secret", "downwardAPI"},
		},
	}

	return psp
}
