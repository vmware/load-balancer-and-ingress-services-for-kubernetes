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

package controllers

import (
	"context"
	"errors"

	"github.com/go-logr/logr"
	akov1alpha1 "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
)

func createOrUpdateStatefulSet(ctx context.Context, ako akov1alpha1.AKOConfig, log logr.Logger, r *AKOConfigReconciler) error {
	var oldSf appsv1.StatefulSet

	if err := r.Get(ctx, getSFNamespacedName(), &oldSf); err != nil {
		log.V(0).Info("no pre-existing statefulset with name", "name", StatefulSetName)
	} else {
		if oldSf.GetName() != "" {
			log.V(0).Info("statefulset present", "name", oldSf.GetName())
		}
	}

	if oldSf.GetName() != "" && rebootRequired {
		log.V(0).Info("rebooting AKO as configmap has been changed")
		err := r.Client.Delete(ctx, &oldSf)
		if err != nil {
			// won't set rebootrequired to true here, as we will keep on rebooting AKO till the error
			// is resolved
			log.Error(err, "error while rebooting ako statefulset", "name", oldSf.GetName(),
				"namespace", oldSf.GetNamespace())
			return err
		}
		rebootRequired = false
		oldSf = appsv1.StatefulSet{}
	}

	sf, err := BuildStatefulSet(ako)
	if err != nil {
		log.Error(err, "error in building statefulset", "name", StatefulSetName)
		return nil
	}
	var cm corev1.ConfigMap
	if err := r.Get(context.TODO(), types.NamespacedName{Name: ConfigMapName, Namespace: AviSystemNS}, &cm); err != nil {
		log.V(0).Info("error getting a configmap", "err", err)
	}

	err = ctrl.SetControllerReference(&ako, &sf, r.Scheme)
	if err != nil {
		log.Error(err, "error in setting controller reference to statefulset, statefulset changes would be ignored")
	}

	if oldSf.GetName() != "" && !rebootRequired {
		if !isSfUpdateRequired(oldSf, sf) {
			log.V(0).Info("no updates required to the statefulset")
			return nil
		}
		err := r.Client.Update(ctx, &sf)
		if err != nil {
			log.Error(err, "unable to update statefulset", "namespace", sf.GetNamespace(),
				"name", sf.GetName())
			return err
		}
	} else {
		err := r.Create(ctx, &sf)
		if err != nil {
			log.Error(err, "unable to create statefulset", "namespace", sf.GetNamespace(),
				"name", sf.GetName())
			return err
		}
	}

	var newSf appsv1.StatefulSet
	err = r.Get(ctx, getSFNamespacedName(), &newSf)
	if err != nil {
		log.V(0).Info("error getting a statefulset with name", "name", StatefulSetName, "err", err)
		return err
	}
	// update this object in the global list
	objList := getObjectList()
	objList[getSFNamespacedName()] = &newSf
	log.V(0).Info("statefulset created/updated", "resource version", newSf.GetResourceVersion())
	return nil
}

func getPullPolicy(pullPolicy string) (v1.PullPolicy, error) {
	typedPullPolicy := v1.PullPolicy(pullPolicy)
	switch typedPullPolicy {
	case v1.PullAlways:
		return v1.PullAlways, nil
	case v1.PullIfNotPresent:
		return v1.PullIfNotPresent, nil
	case v1.PullNever:
		return v1.PullNever, nil
	default:
		return v1.PullPolicy(""), errors.New("invalid pull policy")
	}
}

func buildResources(ako akov1alpha1.AKOConfig) (v1.ResourceRequirements, error) {
	var rr v1.ResourceRequirements

	limitCPU, err := resource.ParseQuantity(ako.Spec.Resources.Limits.CPU)
	if err != nil {
		return rr, err
	}
	limitMemory, err := resource.ParseQuantity(ako.Spec.Resources.Limits.Memory)
	if err != nil {
		return rr, err
	}

	requestedCPU, err := resource.ParseQuantity(ako.Spec.Resources.Requests.CPU)
	if err != nil {
		return rr, err
	}

	requestedMemory, err := resource.ParseQuantity(ako.Spec.Resources.Requests.Memory)
	if err != nil {
		return rr, err
	}

	return v1.ResourceRequirements{
		Limits: v1.ResourceList{
			"cpu":    limitCPU,
			"memory": limitMemory,
		},
		Requests: v1.ResourceList{
			"cpu":    requestedCPU,
			"memory": requestedMemory,
		},
	}, nil
}

func BuildStatefulSet(ako akov1alpha1.AKOConfig) (appsv1.StatefulSet, error) {
	sf := appsv1.StatefulSet{}

	sf.ObjectMeta = metav1.ObjectMeta{
		Name:      StatefulSetName,
		Namespace: AviSystemNS,
	}

	image := ako.Spec.ImageRepository
	imagePullPolicy, err := getPullPolicy(ako.Spec.ImagePullPolicy)
	if err != nil {
		return sf, err
	}
	var replicas int32 = 1
	sf.Spec.Replicas = &replicas
	sf.Spec.ServiceName = ServiceName
	akoLabels := map[string]string{
		"app": "ako",
	}
	sf.Spec.Selector = &metav1.LabelSelector{
		MatchLabels: akoLabels,
	}

	// build the env vars
	envVars := getEnvVars(ako)

	volumeMounts := []v1.VolumeMount{}
	volumes := []v1.Volume{}
	if ako.Spec.PersistentVolumeClaim != "" {
		volumeMounts = append(volumeMounts, v1.VolumeMount{
			Name:      "ako-pv-storage",
			MountPath: ako.Spec.MountPath,
		})
		volumes = append(volumes, v1.Volume{
			Name: "ako-pv-storage",
			VolumeSource: v1.VolumeSource{
				PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
					ClaimName: ako.Spec.PersistentVolumeClaim,
				},
			},
		})
	}

	apiServerPort := ako.Spec.APIServerPort
	if apiServerPort == 0 {
		apiServerPort = 8080
	}

	resources, err := buildResources(ako)
	if err != nil {
		return sf, err
	}
	template := v1.PodTemplateSpec{}
	template.SetLabels(akoLabels)
	template.Spec = v1.PodSpec{
		ServiceAccountName: ServiceAccountName,
		Volumes:            volumes,
		Containers: []v1.Container{
			{
				Name:            "ako",
				VolumeMounts:    volumeMounts,
				Image:           image,
				ImagePullPolicy: imagePullPolicy,
				Lifecycle: &v1.Lifecycle{
					PreStop: &v1.Handler{
						Exec: &v1.ExecAction{
							Command: []string{"/bin/sh", "/var/pre_stop_hook.sh"},
						},
					},
				},
				Ports: []v1.ContainerPort{
					{
						Name:          "http",
						ContainerPort: 80,
						Protocol:      "TCP",
					},
				},
				Resources: resources,
				LivenessProbe: &v1.Probe{
					InitialDelaySeconds: 5,
					PeriodSeconds:       10,
					Handler: v1.Handler{
						HTTPGet: &v1.HTTPGetAction{
							Path: "/api/status",
							Port: intstr.FromInt(apiServerPort),
						},
					},
				},
				Env: envVars,
			},
		},
	}

	sf.Spec.Template = template

	return sf, nil
}
